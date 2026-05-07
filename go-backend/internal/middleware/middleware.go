package middleware

import (
	"fmt"
	"net/http"
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
	allowedLocalhosts := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://0.0.0.0",
	}
	for _, prefix := range allowedLocalhosts {
		if strings.HasPrefix(origin, prefix) {
			return true
		}
	}
	if strings.HasPrefix(origin, "http://192.168.") || strings.HasPrefix(origin, "http://10.") {
		return true
	}
	if strings.HasPrefix(origin, "http://172.") {
		parts := strings.SplitN(strings.TrimPrefix(origin, "http://172."), ".", 2)
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
			if uid, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", uint(uid))
			}
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

		platform := c.GetHeader("X-Platform")
		if platform == "" {
			platform = "openclaw"
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
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
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

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

func RateLimit(limit int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	go func() {
		for {
			time.Sleep(5 * time.Minute)
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
	return 1
}

func GetPlatform(c *gin.Context) string {
	platform, _ := c.Get("platform")
	if p, ok := platform.(string); ok {
		return p
	}
	return "openclaw"
}
