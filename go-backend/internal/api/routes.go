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

	authService := services.NewAuthService(db, cfg.JWTSecret)
	licenseManager := services.NewLicenseManager(db, cfg)
	proProxy := services.NewProProxy(db, cfg)

	public := r.Group("/api/v1")
	{
		public.GET("/auth/init-status", handleInitStatus(authService))
		public.POST("/auth/set-password", handleSetPassword(authService))
		public.POST("/auth/login", handleLogin(authService))
		public.POST("/auth/register", handleRegister(authService))
		public.POST("/auth/reset-password", handleResetPassword(authService))
		public.GET("/license/info", handleLicenseInfo(licenseManager))
		public.POST("/license/activate", handleLicenseActivate(licenseManager))
		public.POST("/license/deactivate", handleLicenseDeactivate(licenseManager))
	}

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg))
	{
		authorized.GET("/auth/me", handleGetMe(authService))

		authorized.GET("/memories", handleListMemories(db))
		authorized.POST("/memories", handleCreateMemory(db))
		authorized.GET("/memories/:id", handleGetMemory(db))
		authorized.PUT("/memories/:id", handleUpdateMemory(db))
		authorized.DELETE("/memories/:id", handleDeleteMemory(db))
		authorized.POST("/memories/:id/restore", handleRestoreMemory(db))
		authorized.GET("/memories/search/keyword", handleSearchKeyword(db))
		authorized.GET("/memories/search/semantic", handleSearchSemantic(db))

		authorized.GET("/knowledge/entities", handleListEntities(db))
		authorized.POST("/knowledge/entities", handleCreateEntity(db))
		authorized.GET("/knowledge/relations", handleListRelations(db))
		authorized.POST("/knowledge/relations", handleCreateRelation(db))
		authorized.GET("/knowledge/graph", handleGetGraph(db))

		authorized.GET("/wiki", handleListWiki(db))
		authorized.POST("/wiki", handleCreateWiki(db))
		authorized.GET("/wiki/:id", handleGetWiki(db))
		authorized.PUT("/wiki/:id", handleUpdateWiki(db))
		authorized.DELETE("/wiki/:id", handleDeleteWiki(db))

		authorized.GET("/reports", handleListReports(db))
		authorized.POST("/reports", handleCreateReport(db))
		authorized.GET("/reports/:date", handleGetReportByDate(db))

		authorized.GET("/stats", handleGetStats(db))
		authorized.GET("/stats/usage", handleGetUsageStats(db))

		authorized.GET("/settings", handleGetSettings(db))
		authorized.PUT("/settings", handleUpdateSettings(db))

		authorized.GET("/memories/decay/stats", handleDecayStats(db))
		authorized.POST("/memories/decay/apply", handleDecayApply(db))
		authorized.PUT("/memories/decay/settings", handleDecaySettingsUpdate(db))
		authorized.GET("/memories/decay/settings", handleDecaySettingsGet(db))
		authorized.DELETE("/memories/trash", handleEmptyTrash(db))
		authorized.GET("/memories/trash", handleListTrash(db))

		authorized.GET("/memories/export", handleExportData(db))
		authorized.POST("/memories/import", handleImportData(db))
		authorized.GET("/data/export", handleExportData(db))
		authorized.POST("/data/import", handleImportData(db))

		authorized.GET("/memories/dedup/scan", handleDedupScan(db))
		authorized.POST("/memories/dedup/merge", handleDedupMerge(db))

		authorized.GET("/memories/health", handleMemoryHealth(db))

		authorized.GET("/memories/recommend", handleMemoryRecommend(db))

		authorized.GET("/openclaw-skills/scan", handleScanSkills)
		authorized.GET("/openclaw-skills/detail", handleSkillDetail)

		authorized.GET("/openclaw-memories/scan", handleScanOpenClawMemories)
		authorized.GET("/openclaw-memories/scan/:agentName", handleScanOpenClawAgent)
		authorized.POST("/openclaw-memories/import", handleImportOpenClawMemories(db))

		authorized.GET("/chromadb/status", handleChromaDBStatus)
		authorized.POST("/chromadb/install", handleChromaDBInstall)

		authorized.GET("/backups", handleListBackups)
		authorized.POST("/backups", handleCreateBackup)

		pro := authorized.Group("/pro")
		{
			pro.GET("/decay/stats", handleProDecayStats(proProxy, db))
			pro.POST("/decay/apply", handleProDecayApply(proProxy, db))
			pro.POST("/reinforce/:id", handleProReinforce(proProxy, db))
			pro.GET("/prune-suggest", handleProPruneSuggest(proProxy, db))
			pro.GET("/conflicts/scan", handleProConflictScan(proProxy, db))
			pro.POST("/conflicts/resolve/:index", handleProConflictResolve(proProxy))
			pro.POST("/token/route", handleProTokenRoute(proProxy))
			pro.GET("/token/stats", handleProTokenStats(proProxy))
			pro.POST("/ai/extract", handleProAIExtract(proProxy, db))
			pro.POST("/auto-graph", handleProAutoGraph(proProxy))
			pro.GET("/backup/schedule", handleProBackupSchedule(proProxy))
			pro.POST("/backup/schedule", handleProSetBackupSchedule(proProxy))

			pro.POST("/compress/preview", handleProCompressPreview(proProxy, db))
			pro.POST("/compress/apply", handleProCompressApply(proProxy, db))
			pro.GET("/compress/config", handleProCompressConfig(proProxy))
			pro.PUT("/compress/config", handleProSetCompressConfig(proProxy))

			pro.GET("/evolution/insights", handleProEvolutionInsights(proProxy))
			pro.POST("/evolution/discover", handleProEvolutionDiscover(proProxy, db))
			pro.POST("/evolution/infer", handleProEvolutionInfer(proProxy))
			pro.POST("/evolution/importance", handleProEvolutionImportance(proProxy, db))
			pro.POST("/evolution/prefetch", handleProEvolutionPrefetch(proProxy, db))
		}
	}
}
