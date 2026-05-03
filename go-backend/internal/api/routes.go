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
		public.POST("/auth/login", middleware.LoginRateLimit(), handleLogin(authService))
		public.POST("/auth/register", middleware.LoginRateLimit(), handleRegister(authService))
		public.POST("/auth/forgot-password", handleForgotPassword(authService))
		public.GET("/install-status", handleInstallStatus(db))
		public.GET("/check-update", handleCheckUpdate)
		public.POST("/license/activate", handleLicenseActivate(licenseManager))
	}

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg, db), middleware.JWTRateLimit())
	{
		authorized.GET("/license/info", handleLicenseInfo(proProxy))
		authorized.POST("/license/deactivate", handleLicenseDeactivate(licenseManager))
		authorized.GET("/auth/me", handleGetMe(authService))
		authorized.POST("/auth/change-password", handleChangePassword(authService))
		authorized.POST("/auth/reset-password", handleResetPassword(authService))

		authorized.GET("/memories", handleListMemories(db))
		authorized.POST("/memories", handleCreateMemory(db))
		authorized.GET("/memories/:id", handleGetMemory(db))
		authorized.PUT("/memories/:id", handleUpdateMemory(db))
		authorized.DELETE("/memories/:id", handleDeleteMemory(db))
		authorized.POST("/memories/:id/restore", handleRestoreMemory(db))
		authorized.POST("/memories/:id/decrypt", handleDecryptMemory(db))
		authorized.GET("/memories/search/keyword", handleSearchKeyword(db))
		authorized.GET("/memories/search/semantic", handleSearchSemantic(db))

		authorized.GET("/knowledge/entities", handleListEntities(db))
		authorized.POST("/knowledge/entities", handleCreateEntity(db))
		authorized.GET("/knowledge/entities/:id", handleGetEntity(db))
		authorized.PUT("/knowledge/entities/:id", handleUpdateEntity(db))
		authorized.DELETE("/knowledge/entities/:id", handleDeleteEntity(db))
		authorized.GET("/knowledge/relations", handleListRelations(db))
		authorized.POST("/knowledge/relations", handleCreateRelation(db))
		authorized.DELETE("/knowledge/relations/:id", handleDeleteRelation(db))
		authorized.GET("/knowledge/graph", handleGetGraph(db))

		authorized.GET("/wiki", handleListWiki(db))
		authorized.POST("/wiki", handleCreateWiki(db))
		authorized.GET("/wiki/tree", handleWikiTree(db))
		authorized.GET("/wiki/search", handleWikiSearch(db))
		authorized.GET("/wiki/categories", handleWikiCategories(db))
		authorized.GET("/wiki/stats", handleWikiStats(db))
		authorized.GET("/wiki/config", handleWikiConfig(db))
		authorized.GET("/wiki/:id", handleGetWiki(db))
		authorized.PUT("/wiki/:id", handleUpdateWiki(db))
		authorized.DELETE("/wiki/:id", handleDeleteWiki(db))
		authorized.POST("/wiki/:id/mark-complete", handleWikiMarkComplete(db))
		authorized.POST("/wiki/:id/mark-in-progress", handleWikiMarkInProgress(db))
		authorized.POST("/wiki/ai/extract", handleWikiAIExtract(db))
		authorized.POST("/wiki/:id/refine", handleWikiRefine(db))

		authorized.GET("/projects", handleListProjects(db))
		authorized.POST("/projects", handleCreateProject(db))
		authorized.GET("/projects/search", handleProjectSearch(db))
		authorized.GET("/projects/categories", handleProjectCategories(db))
		authorized.GET("/projects/context", handleProjectContext(db))
		authorized.GET("/projects/:id", handleGetProject(db))
		authorized.PUT("/projects/:id", handleUpdateProject(db))
		authorized.DELETE("/projects/:id", handleDeleteProject(db))
		authorized.GET("/projects/:id/notes", handleProjectNotes(db))
		authorized.POST("/projects/:id/notes", handleAddProjectNote(db))
		authorized.PUT("/projects/notes/:noteId", handleUpdateProjectNote(db))
		authorized.DELETE("/projects/notes/:noteId", handleDeleteProjectNote(db))
		authorized.POST("/projects/:id/extract-memories", handleProjectExtractMemories(db))

		authorized.GET("/reports", handleListReports(db))
		authorized.POST("/reports", handleCreateReport(db))
		authorized.POST("/reports/generate", handleGenerateReport(db))
		authorized.GET("/reports/:date", handleGetReportByDate(db))

		authorized.GET("/stats", handleGetStats(db))
		authorized.GET("/stats/usage", handleGetUsageStats(db))

		authorized.GET("/settings", handleGetSettings(db))
		authorized.PUT("/settings", handleUpdateSettings(db))

		authorized.GET("/api-keys", handleListAPIKeys(db))
		authorized.POST("/api-keys", handleCreateAPIKey(db))
		authorized.DELETE("/api-keys/:id", handleDeleteAPIKey(db))

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

		authorized.GET("/memories/smart-load", handleSmartLoad(db))
		authorized.POST("/memories/:id/reinforce", handleReinforceMemory(db))
		authorized.POST("/memories/generate-summaries", handleGenerateSummaries(db))
		authorized.POST("/memories/extract", handleExtractMemories(db))
		authorized.POST("/memories/extract-and-save", handleExtractAndSave(db))
		authorized.POST("/memories/:id/verify", handleVerifyMemory(db))
		authorized.POST("/memories/scan-secrets", handleScanSecrets(db))

		authorized.GET("/session-memories", handleListSessionMemories(db))
		authorized.POST("/session-memories", handleCreateSessionMemory(db))
		authorized.GET("/session-memories/:id", handleGetSessionMemory(db))
		authorized.PUT("/session-memories/:id", handleUpdateSessionMemory(db))
		authorized.DELETE("/session-memories/:id", handleDeleteSessionMemory(db))

		authorized.GET("/openclaw-skills/scan", handleScanSkills)
		authorized.GET("/openclaw-skills/detail", handleSkillDetail)
		authorized.POST("/openclaw-skills/install", handleInstallSkill)

		authorized.GET("/openclaw-memories/scan", handleScanOpenClawMemories)
		authorized.GET("/openclaw-memories/scan/:agentName", handleScanOpenClawAgent)
		authorized.POST("/openclaw-memories/import", handleImportOpenClawMemories(db))
		authorized.GET("/openclaw-sync/status", handleOpenClawSyncStatus(db))
		authorized.POST("/openclaw-sync/force", handleOpenClawSyncForce(db))
		authorized.POST("/openclaw-sync/toggle", handleOpenClawSyncToggle(db))
		authorized.POST("/memories/auto-import", handleAutoImportMemories(db))

		authorized.GET("/chromadb/status", handleChromaDBStatus(db))
		authorized.POST("/chromadb/install", handleChromaDBInstall(db))
		authorized.POST("/chromadb/sync", handleChromaDBSync(db))

		authorized.GET("/backups", handleListBackups)
		authorized.POST("/backups", handleCreateBackup)
		authorized.GET("/backups/:filename", handleDownloadBackup(db))
		authorized.POST("/backups/:filename/restore", handleRestoreBackup(db))
		authorized.DELETE("/backups/:filename", handleDeleteBackup(db))

		pro := authorized.Group("/pro")
		{
			pro.GET("/decay/stats", handleProDecayStats(proProxy, db))
			pro.POST("/decay/apply", handleProDecayApply(proProxy, db))
			pro.POST("/reinforce/:id", handleProReinforce(proProxy, db))
			pro.GET("/prune-suggest", handleProPruneSuggest(proProxy, db))
			pro.GET("/conflicts/scan", handleProConflictScan(proProxy, db))
			pro.POST("/conflicts/resolve/:index", handleProConflictResolve(proProxy, db))
			pro.GET("/token/route", handleProTokenRoute(proProxy, db))
			pro.GET("/token/stats", handleProTokenStats(proProxy, db))
			pro.POST("/ai/extract", handleProAIExtract(proProxy, db))
			pro.POST("/auto-graph", handleProAutoGraph(proProxy, db))
			pro.GET("/backup/schedule", handleProBackupSchedule(proProxy, db))
			pro.POST("/backup/schedule", handleProSetBackupSchedule(proProxy, db))

			pro.POST("/compress/preview", handleProCompressPreview(proProxy, db))
			pro.POST("/compress/apply", handleProCompressApply(proProxy, db))
			pro.GET("/compress/config", handleProCompressConfig(proProxy, db))
			pro.PUT("/compress/config", handleProSetCompressConfig(proProxy, db))

			pro.GET("/evolution/insights", handleProEvolutionInsights(proProxy, db))
			pro.POST("/evolution/discover", handleProEvolutionDiscover(proProxy, db))
			pro.POST("/evolution/infer", handleProEvolutionInfer(proProxy, db))
			pro.POST("/evolution/importance", handleProEvolutionImportance(proProxy, db))
			pro.POST("/evolution/prefetch", handleProEvolutionPrefetch(proProxy, db))
		}
	}

	external := r.Group("/api/v1/external")
	external.Use(middleware.APIKeyAuth(db), middleware.APIKeyRateLimit())
	{
		external.POST("/memories", handleExternalCreateMemory(db))
		external.POST("/memories/batch", handleExternalBatchCreateMemories(db))
		external.GET("/memories/search", handleExternalSearchMemories(db))
		external.POST("/conversations", handleExternalPushConversation(db))
	}
}
