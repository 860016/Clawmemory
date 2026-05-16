package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := isOriginAllowed(origin)
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Platform")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}

	if customOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); customOrigins != "" {
		for _, allowed := range strings.Split(customOrigins, ",") {
			allowed = strings.TrimSpace(allowed)
			if allowed != "" && origin == allowed {
				return true
			}
		}
		return false
	}

	allowedLocalhosts := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://0.0.0.0",
		"https://localhost",
		"https://127.0.0.1",
	}
	for _, prefix := range allowedLocalhosts {
		if strings.HasPrefix(origin, prefix) {
			return true
		}
	}
	if strings.HasPrefix(origin, "http://192.168.") || strings.HasPrefix(origin, "https://192.168.") {
		return true
	}
	if strings.HasPrefix(origin, "http://10.") || strings.HasPrefix(origin, "https://10.") {
		return true
	}
	if strings.HasPrefix(origin, "http://172.") || strings.HasPrefix(origin, "https://172.") {
		trimmed := origin
		if strings.HasPrefix(trimmed, "https://") {
			trimmed = "http://" + strings.TrimPrefix(trimmed, "https://")
		}
		parts := strings.SplitN(strings.TrimPrefix(trimmed, "http://172."), ".", 2)
		if len(parts) > 0 {
			var second int
			if _, err := fmt.Sscanf(parts[0], "%d", &second); err == nil && second >= 16 && second <= 31 {
				return true
			}
		}
	}
	return false
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return param.TimeStamp.Format(time.RFC3339) + " " +
			param.Method + " " + param.Path + " " +
			param.ClientIP + " " + param.ErrorMessage + "\n"
	})
}

func Auth(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API keys are not allowed on management endpoints, use /api/v1/external/ instead",
			})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			uid, ok := claims["user_id"].(float64)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: missing user_id"})
				return
			}
			c.Set("user_id", uint(uid))

			if tokenVersion, ok := claims["token_version"].(float64); ok {
				var user struct {
					TokenVersion int
				}
				if err := db.Table("users").Select("token_version").Where("id = ?", uint(uid)).First(&user).Error; err == nil {
					if user.TokenVersion != int(tokenVersion) {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "token has been revoked, please login again",
						})
						return
					}
				}
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		c.Set("auth_method", "jwt")
		c.Next()
	}
}

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyHeader := c.GetHeader("X-API-Key")
		if apiKeyHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}

		svc := services.NewAPIKeyService(db)
		apiKey, err := svc.Validate(apiKeyHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("user_id", apiKey.UserID)
		c.Set("auth_method", "apikey")
		c.Set("api_key_id", apiKey.ID)
		c.Set("api_key_name", apiKey.Name)
		c.Set("api_key_permissions", apiKey.Permissions)
		c.Set("agent_name", apiKey.AgentName)

		platform := c.GetHeader("X-Platform")
		if platform == "" {
			if apiKey.AgentName != "" {
				platform = apiKey.AgentName
			} else {
				platform = "clawmemory"
			}
		}
		c.Set("platform", platform)

		c.Next()
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	maxKeys  int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		maxKeys:  10000,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	requests := rl.requests[key]
	valid := requests[:0]
	for _, t := range requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	if len(valid) == 0 && len(rl.requests) >= rl.maxKeys {
		rl.evictOldest()
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

func (rl *rateLimiter) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range rl.requests {
		if len(v) > 0 {
			if first || v[0].Before(oldestTime) {
				oldestKey = k
				oldestTime = v[0]
				first = false
			}
		} else {
			delete(rl.requests, k)
			return
		}
	}
	if oldestKey != "" {
		delete(rl.requests, oldestKey)
	}
}

func RateLimit(limit int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			limiter.mu.Lock()
			cutoff := time.Now().Add(-window)
			for k, v := range limiter.requests {
				valid := v[:0]
				for _, t := range v {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(limiter.requests, k)
				} else {
					limiter.requests[k] = valid
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			return
		}
		c.Next()
	}
}

func APIKeyRateLimit() gin.HandlerFunc {
	return RateLimit(60, time.Minute, func(c *gin.Context) string {
		keyID, _ := c.Get("api_key_id")
		return fmt.Sprintf("apikey:%v", keyID)
	})
}

func JWTRateLimit() gin.HandlerFunc {
	return RateLimit(120, time.Minute, func(c *gin.Context) string {
		userID, _ := c.Get("user_id")
		return fmt.Sprintf("jwt:%v", userID)
	})
}

func LoginRateLimit() gin.HandlerFunc {
	return RateLimit(10, time.Minute, func(c *gin.Context) string {
		return "login:" + c.ClientIP()
	})
}

func GetUserID(c *gin.Context) uint {
	userID, _ := c.Get("user_id")
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

func GetPlatform(c *gin.Context) string {
	platform, _ := c.Get("platform")
	if p, ok := platform.(string); ok {
		return p
	}
	return "clawmemory"
}

func GetAgentName(c *gin.Context) string {
	agentName, _ := c.Get("agent_name")
	if a, ok := agentName.(string); ok {
		return a
	}
	return ""
}

func GetAPIKeyPermissions(c *gin.Context) string {
	perms, _ := c.Get("api_key_permissions")
	if p, ok := perms.(string); ok {
		return p
	}
	return ""
}

func HasAPIKeyPermission(c *gin.Context, permission string) bool {
	authMethod, _ := c.Get("auth_method")
	if authMethod != "apikey" {
		return true
	}
	perms := GetAPIKeyPermissions(c)
	if perms == "" {
		return false
	}
	for _, p := range splitPermissions(perms) {
		if p == "admin" || p == "*" || p == permission {
			return true
		}
		if isParentPermission(p, permission) {
			return true
		}
	}
	return false
}

func isParentPermission(parent, child string) bool {
	if parent == "read" {
		return strings.HasSuffix(child, ":read")
	}
	if parent == "write" {
		return strings.HasSuffix(child, ":read") || strings.HasSuffix(child, ":write")
	}
	parts := strings.SplitN(parent, ":", 2)
	if len(parts) == 2 {
		resource := parts[0]
		scope := parts[1]
		childParts := strings.SplitN(child, ":", 2)
		if len(childParts) == 2 && childParts[0] == resource {
			if scope == "write" {
				return childParts[1] == "read" || childParts[1] == "write"
			}
			if scope == "read" {
				return childParts[1] == "read"
			}
		}
	}
	return false
}

func splitPermissions(perms string) []string {
	var result []string
	for _, p := range strings.Split(perms, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
