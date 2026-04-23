package api

import (
	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cfg := config.Load()

	// 初始化服务
	authService := services.NewAuthService(db, cfg.JWTSecret)
	licenseManager := services.NewLicenseManager(db, cfg)

	// 公开路由
	public := r.Group("/api/v1")
	{
		public.POST("/auth/register", handleRegister(authService))
		public.POST("/auth/login", handleLogin(authService))
		public.POST("/auth/reset-password", handleResetPassword(authService))
		public.GET("/license/info", handleLicenseInfo(licenseManager))
		public.POST("/license/activate", handleLicenseActivate(licenseManager))
		public.POST("/license/deactivate", handleLicenseDeactivate(licenseManager))
	}

	// 需要认证的路由
	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg))
	{
		// 记忆管理
		authorized.GET("/memories", handleListMemories(db))
		authorized.POST("/memories", handleCreateMemory(db))
		authorized.GET("/memories/:id", handleGetMemory(db))
		authorized.PUT("/memories/:id", handleUpdateMemory(db))
		authorized.DELETE("/memories/:id", handleDeleteMemory(db))
		authorized.POST("/memories/:id/restore", handleRestoreMemory(db))
		authorized.GET("/memories/search/keyword", handleSearchKeyword(db))
		authorized.GET("/memories/search/semantic", handleSearchSemantic(db))

		// 知识图谱
		authorized.GET("/knowledge/entities", handleListEntities(db))
		authorized.POST("/knowledge/entities", handleCreateEntity(db))
		authorized.GET("/knowledge/relations", handleListRelations(db))
		authorized.POST("/knowledge/relations", handleCreateRelation(db))
		authorized.GET("/knowledge/graph", handleGetGraph(db))

		// Wiki
		authorized.GET("/wiki", handleListWiki(db))
		authorized.POST("/wiki", handleCreateWiki(db))
		authorized.GET("/wiki/:id", handleGetWiki(db))
		authorized.PUT("/wiki/:id", handleUpdateWiki(db))
		authorized.DELETE("/wiki/:id", handleDeleteWiki(db))

		// 日报
		authorized.GET("/reports", handleListReports(db))
		authorized.POST("/reports", handleCreateReport(db))
		authorized.GET("/reports/:date", handleGetReportByDate(db))

		// 统计
		authorized.GET("/stats", handleGetStats(db))
		authorized.GET("/stats/decay", handleGetDecayStats(db, licenseManager))

		// 设置
		authorized.GET("/settings", handleGetSettings())
		authorized.PUT("/settings", handleUpdateSettings())
	}
}
