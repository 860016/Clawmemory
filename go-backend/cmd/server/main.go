package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"clawmemory/internal/api"
	"clawmemory/internal/config"
	"clawmemory/internal/database"
	"clawmemory/internal/middleware"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	resetPassword := flag.String("reset-password", "", "Reset user password. Usage: --reset-password NEW_PASSWORD or --reset-password USERNAME:NEW_PASSWORD")
	resetUser := flag.String("reset-user", "", "Specify username for password reset (optional, defaults to first user)")
	showVersion := flag.Bool("version", false, "Show version info")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ClawMemory v%s\n", config.AppVersion)
		fmt.Printf("GitHub: %s\n", config.GitHubRepoURL)
		os.Exit(0)
	}

	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)

	envPaths := []string{
		filepath.Join(exeDir, ".env"),
		".env",
	}
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			log.Printf("Loaded .env from %s", p)
			break
		}
	}

	cfg := config.Load()

	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if *resetPassword != "" {
		username := *resetUser
		password := *resetPassword
		if idx := indexByte(password, ':'); idx > 0 {
			username = password[:idx]
			password = password[idx+1:]
		}
		resetAdminPassword(db, username, password)
		os.Exit(0)
	}

	services.Init(db)

	logger := services.InitLogger(filepath.Join(exeDir, "logs"), "info", "clawmemory")
	logger.SetDB(db)
	metrics := services.InitMetrics(db)
	_ = metrics

	services.InitSecurity(db, cfg.JWTSecret)
	services.InitRiskSwitchService(db)

	gracefulShutdown := services.NewGracefulShutdown(db)
	backupSvc := services.NewBackupService(db, filepath.Join(exeDir, "backups"))
	_ = backupSvc

	proProvider := services.InitProProvider(db, cfg)

	autoCreateAPIKey(db)

	log.Printf("✅ Pro features active (tier: %s)", proProvider.GetTier())

	syncService := services.GetOpenClawSyncService(db)
	go syncService.Start()
	log.Printf("Agent auto-sync service started")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			svc := services.NewSkillLearningService(db)
			if deleted, err := svc.CleanupOldTraces(30); err == nil && deleted > 0 {
				log.Printf("Skill cleanup: removed %d old action traces", deleted)
			}
			if deleted, err := svc.CleanupDismissedSuggestions(); err == nil && deleted > 0 {
				log.Printf("Skill cleanup: removed %d dismissed suggestions", deleted)
			}
		}
	}()
	log.Printf("Skill auto-cleanup service started (daily)")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	api.RegisterRoutes(r, db)

	frontendDist := filepath.Join(exeDir, "frontend_dist")
	r.Static("/assets", filepath.Join(frontendDist, "assets"))
	r.StaticFile("/favicon.ico", filepath.Join(frontendDist, "favicon.ico"))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 5 && path[:5] == "/api/" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File(filepath.Join(frontendDist, "index.html"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8765"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	addr := host + ":" + port
	log.Printf("ClawMemory v%s starting on %s", config.AppVersion, addr)

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gracefulShutdown.Shutdown(10 * time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

func autoCreateAPIKey(db *gorm.DB) {
	var count int64
	db.Model(&models.APIKey{}).Count(&count)
	if count > 0 {
		return
	}

	var user models.User
	if err := db.First(&user).Error; err != nil {
		return
	}

	svc := services.NewAPIKeyService(db)
	apiKey, rawKey, err := svc.Create(user.ID, "ClawMemory Auto")
	if err != nil {
		log.Printf("Warning: failed to auto-create API key: %v", err)
		return
	}

	log.Printf("========================================")
	log.Printf("  Auto-generated API Key for ClawMemory")
	log.Printf("  Name: %s", apiKey.Name)
	log.Printf("  Key:  %s", rawKey)
	log.Printf("  Please save this key, it won't be shown again!")
	log.Printf("========================================")
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func resetAdminPassword(db *gorm.DB, username, newPassword string) {
	if len(newPassword) < 6 {
		log.Fatal("Error: password must be at least 6 characters")
	}

	if username == "" {
		var userCount int64
		db.Model(&models.User{}).Count(&userCount)
		if userCount == 0 {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Fatal("Error: failed to hash password:", err)
			}
			user := models.User{
				Username:  "admin",
				Password:  string(hashedPassword),
				Role:      "admin",
				IsFounder: true,
			}
			if err := db.Create(&user).Error; err != nil {
				log.Fatal("Error: failed to create admin user:", err)
			}
			fmt.Println("✅ Founder account 'admin' created with password set successfully.")
			fmt.Println("Please restart the server without --reset-password flag.")
			return
		}
		var firstUser models.User
		if err := db.Order("id ASC").First(&firstUser).Error; err != nil {
			log.Fatal("Error: failed to query user:", err)
		}
		username = firstUser.Username
		fmt.Printf("No username specified, resetting password for first user '%s'\n", username)
	}

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Fatalf("Error: user '%s' not found", username)
		}
		log.Fatal("Error: failed to query user:", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error: failed to hash password:", err)
	}

	if err := db.Model(&user).Updates(map[string]interface{}{
		"password":        string(hashedPassword),
		"token_version":   gorm.Expr("token_version + 1"),
		"failed_attempts": 0,
		"locked_until":    nil,
		"refresh_token":   "",
	}).Error; err != nil {
		log.Fatal("Error: failed to update password:", err)
	}

	fmt.Printf("✅ Password for user '%s' has been reset successfully.\n", user.Username)
	fmt.Println("Please restart the server without --reset-password flag.")
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println("  ./clawmemory --reset-password NEW_PASSWORD              (reset first user)")
	fmt.Println("  ./clawmemory --reset-password admin:NEW_PASSWORD       (reset specific user)")
	fmt.Println("  ./clawmemory --reset-password NEW_PASSWORD --reset-user admin")
}
