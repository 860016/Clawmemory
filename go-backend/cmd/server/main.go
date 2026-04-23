package main

import (
	"log"
	"os"

	"clawmemory/internal/api"
	"clawmemory/internal/config"
	"clawmemory/internal/database"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}

	// 自动迁移
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化服务
	services.Init(db)

	// 设置路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// 注册 API 路由
	api.RegisterRoutes(r, db)

	// 静态文件（前端）
	r.Static("/assets", "./frontend_dist/assets")
	r.StaticFile("/", "./frontend_dist/index.html")

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
