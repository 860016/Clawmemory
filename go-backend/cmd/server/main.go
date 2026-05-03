package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	resetPassword := flag.String("reset-password", "", "Reset admin password")
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
		resetAdminPassword(db, *resetPassword)
		os.Exit(0)
	}

	services.Init(db)

	autoCreateAPIKey(db)

	syncService := services.GetOpenClawSyncService(db)
	go syncService.Start()
	log.Printf("OpenClaw auto-sync service started")

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
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
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
	apiKey, rawKey, err := svc.Create(user.ID, "OpenClaw Auto")
	if err != nil {
		log.Printf("Warning: failed to auto-create API key: %v", err)
		return
	}

	log.Printf("========================================")
	log.Printf("  Auto-generated API Key for OpenClaw")
	log.Printf("  Name: %s", apiKey.Name)
	log.Printf("  Key:  %s", rawKey)
	log.Printf("  Please save this key, it won't be shown again!")
	log.Printf("========================================")
}

func resetAdminPassword(db *gorm.DB, newPassword string) {
	if len(newPassword) < 4 {
		log.Fatal("Error: password must be at least 4 characters")
	}

	var user models.User
	if err := db.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Fatal("Error: no user found in database. Please start the server first to create an account.")
		}
		log.Fatal("Error: failed to query user:", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error: failed to hash password:", err)
	}

	if err := db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		log.Fatal("Error: failed to update password:", err)
	}

	fmt.Printf("✅ Password for user '%s' has been reset successfully.\n", user.Username)
	fmt.Println("Please restart the server without --reset-password flag.")
}
