package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"clawmemory/internal/api"
	"clawmemory/internal/config"
	"clawmemory/internal/database"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
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

	services.Init(db)

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
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
