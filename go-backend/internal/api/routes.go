package api

import (
	"clawmemory/internal/ai"
	"clawmemory/internal/config"
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	cfg := config.Load()

	authService := services.NewAuthService(db, cfg.JWTSecret)
	aiRouter := ai.NewAIRouter(db)
	aiSvc := ai.NewAIService(aiRouter, db)
	embedAdapter := &aiEmbeddingAdapter{router: aiRouter}
	services.InitEmbeddingService(db, embedAdapter)
	services.SetEmbeddingAIRouter(embedAdapter)

	public := r.Group("/api/v1")
	{
		public.GET("/auth/init-status", handleInitStatus(authService))
		public.POST("/auth/set-password", middleware.LoginRateLimit(), handleSetPassword(authService))
		public.POST("/auth/login", middleware.LoginRateLimit(), handleLogin(authService))
		public.POST("/auth/refresh", handleRefreshToken(authService))
		public.POST("/auth/login-status", handleLoginStatus(authService))
		public.POST("/auth/register", middleware.LoginRateLimit(), handleRegister(authService))
		public.POST("/auth/register-with-invitation", middleware.LoginRateLimit(), handleRegisterWithInvitation(authService))
		public.POST("/auth/forgot-password", middleware.LoginRateLimit(), handleForgotPassword(authService))
		public.GET("/install-status", handleInstallStatus(db))
		public.GET("/health", handleHealthCheck(db))
	}

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg, db), middleware.JWTRateLimit())
	{
		authorized.GET("/auth/me", handleGetMe(authService))
		authorized.POST("/auth/change-password", handleChangePassword(authService))
		authorized.POST("/auth/reset-password", handleResetPassword(authService))
		authorized.POST("/auth/revoke-tokens", handleRevokeAllTokens(db, authService))

		authorized.GET("/users", handleListUsers(db))
		authorized.POST("/users/reset-password", handleAdminResetUserPassword(db, authService))

		authorized.GET("/invitations", handleListInvitations(db))
		authorized.POST("/invitations", handleCreateInvitation(db))
		authorized.DELETE("/invitations/:id", handleDeleteInvitation(db))

		authorized.GET("/memories", handleListMemories(db))
		authorized.POST("/memories", handleCreateMemory(db))
		authorized.GET("/memories/:id", handleGetMemory(db))
		authorized.PUT("/memories/:id", handleUpdateMemory(db))
		authorized.DELETE("/memories/:id", handleDeleteMemory(db))
		authorized.POST("/memories/:id/restore", handleRestoreMemory(db))
		authorized.POST("/memories/:id/decrypt", handleDecryptMemory(db))
		authorized.GET("/memories/search", handleSearchMemories(db))
		authorized.GET("/memories/:id/history", handleMemoryHistory(db))
		authorized.GET("/memories/:id/evolution", handleMemoryEvolution(db))
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
		authorized.POST("/wiki/ai/extract", handleWikiAIExtract(aiSvc, db))
		authorized.POST("/wiki/:id/refine", handleWikiRefine(aiSvc, db))

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
		authorized.GET("/stats/metrics", handleMetrics(db))

		authorized.GET("/audit-log", handleAuditLog(db))

		authorized.GET("/risk-switches", handleGetRiskSwitches(db))
		authorized.PUT("/risk-switches", handleSetRiskSwitches(db))

		authorized.GET("/writeback/targets", handleGetWritebackTargets(db))
		authorized.POST("/writeback/preview", handlePreviewWriteback(db))
		authorized.POST("/writeback/execute", handleExecuteWriteback(db))

		authorized.GET("/settings", handleGetSettings(db))
		authorized.PUT("/settings", handleUpdateSettings(db))

		authorized.GET("/api-keys", handleListAPIKeys(db))
		authorized.POST("/api-keys", handleCreateAPIKey(db))
		authorized.DELETE("/api-keys/:id", handleDeleteAPIKey(db))

		authorized.GET("/memories/decay/stats", handleDecayStats(db))
		authorized.POST("/memories/decay/apply", handleDecayApply(aiSvc, db))
		authorized.PUT("/memories/decay/settings", handleDecaySettingsUpdate(db))
		authorized.GET("/memories/decay/settings", handleDecaySettingsGet(db))
		authorized.DELETE("/memories/trash", handleEmptyTrash(db))
		authorized.GET("/memories/trash", handleListTrash(db))

		authorized.POST("/data/export", handleExportData(db))
		authorized.POST("/data/import", handleImportData(db))

		authorized.GET("/memories/dedup/scan", handleDedupScan(db))
		authorized.POST("/memories/dedup/merge", handleDedupMerge(db))

		authorized.GET("/memories/health", handleMemoryHealth(db))
		authorized.GET("/memories/quality", handleMemoryQuality(db))
		authorized.POST("/memories/auto-fix", handleMemoryAutoFix(db))

		authorized.GET("/mcp/config", handleMCPConfig(db))

		authorized.GET("/memories/governance/status", handleGovernanceStatus(db))
		authorized.POST("/memories/governance/run", handleGovernanceRun(db))
		authorized.PUT("/memories/governance/config", handleGovernanceConfig(db))

		authorized.GET("/memories/smart-load", handleSmartLoad(db))
		authorized.POST("/memories/:id/reinforce", handleReinforceMemory(db))
		authorized.POST("/memories/generate-summaries", handleGenerateSummaries(db))
		authorized.POST("/memories/extract", handleExtractMemories(db))
		authorized.POST("/memories/extract-and-save", handleExtractAndSave(db))
		authorized.POST("/memories/:id/verify", handleVerifyMemory(db))
		authorized.POST("/memories/scan-secrets", handleScanSecrets(db))
		authorized.POST("/memories/validate", handleBatchValidate(db))

		authorized.GET("/memories/templates", handleListTemplates(db))
		authorized.POST("/memories/templates", handleCreateTemplate(db))
		authorized.DELETE("/memories/templates/:name", handleDeleteTemplate(db))
		authorized.POST("/memories/templates/:name/apply", handleApplyTemplate(db))

		authorized.GET("/session-memories", handleListSessionMemories(db))
		authorized.POST("/session-memories", handleCreateSessionMemory(db))
		authorized.GET("/session-memories/:id", handleGetSessionMemory(db))
		authorized.PUT("/session-memories/:id", handleUpdateSessionMemory(db))
		authorized.DELETE("/session-memories/:id", handleDeleteSessionMemory(db))

		authorized.GET("/agents/connected", handleGetConnectedAgents)
		authorized.GET("/agent-memories/scan", handleScanOpenClawMemories)
		authorized.GET("/agent-memories/scan/:agentName", handleScanOpenClawAgent)
		authorized.POST("/agent-memories/import", handleImportOpenClawMemories(db))
		authorized.GET("/agent-sync/status", handleOpenClawSyncStatus(db))
		authorized.POST("/agent-sync/force", handleOpenClawSyncForce(db))
		authorized.POST("/agent-sync/toggle", handleOpenClawSyncToggle(db))
		authorized.GET("/agent/agents-md", handleGetAgentsMD(db))
		authorized.POST("/memories/auto-import", handleAutoImportMemories(db))

		authorized.GET("/chromadb/status", handleChromaDBStatus(db))
		authorized.POST("/chromadb/install", handleChromaDBInstall(db))
		authorized.POST("/chromadb/sync", handleChromaDBSync(db))

		authorized.GET("/openclaw-skills/scan", handleScanSkills)
		authorized.GET("/openclaw-skills/detail", handleSkillDetail)

		authorized.POST("/shares", handleShareMemory(db))
		authorized.GET("/shares/pending", handleListPendingShares(db))
		authorized.GET("/shares/outbound", handleListOutboundShares(db))
		authorized.POST("/shares/:id/approve", handleApproveShare(db))
		authorized.POST("/shares/:id/reject", handleRejectShare(db))
		authorized.POST("/shares/:id/revoke", handleRevokeShare(db))
		authorized.GET("/shares/agent/:agent", handleGetAgentMemories(db))

		authorized.GET("/share-rules", handleListShareRules(db))
		authorized.POST("/share-rules", handleCreateShareRule(db))
		authorized.PUT("/share-rules/:id", handleUpdateShareRule(db))
		authorized.DELETE("/share-rules/:id", handleDeleteShareRule(db))

		authorized.GET("/reasoning/config", handleGetReasoningConfig(db))
		authorized.PUT("/reasoning/config", handleSetReasoningConfig(db))
		authorized.POST("/reasoning/test", handleTestReasoningConnection(db))
		authorized.POST("/reasoning/execute", handleReason(db))

		authorized.GET("/backups", handleListBackups)
		authorized.POST("/backups", handleCreateBackup)
		authorized.GET("/backups/:filename", handleDownloadBackup(db))
		authorized.POST("/backups/:filename/restore", handleRestoreBackup(db))
		authorized.DELETE("/backups/:filename", handleDeleteBackup(db))

		toolbox := services.NewToolboxService(db)

		authorized.GET("/toolbox/conflicts/scan", handleToolboxConflictScan(aiSvc, toolbox, db))
		authorized.POST("/toolbox/conflicts/resolve/:index", handleToolboxConflictResolve(toolbox, db))
		authorized.GET("/toolbox/token/route", handleToolboxTokenRoute(aiSvc, toolbox, db))
		authorized.GET("/toolbox/token/stats", handleToolboxTokenStats(toolbox, db))
		authorized.POST("/toolbox/ai/extract", handleToolboxAIExtract(aiSvc, toolbox, db))
		authorized.POST("/toolbox/auto-graph", handleToolboxAutoGraph(db))

		authorized.POST("/toolbox/compress/preview", handleToolboxCompressPreview(db))
		authorized.POST("/toolbox/compress/apply", handleToolboxCompressApply(aiSvc, db))
		authorized.GET("/toolbox/compress/config", handleToolboxCompressConfig(db))
		authorized.PUT("/toolbox/compress/config", handleToolboxSetCompressConfig(db))

		authorized.GET("/memories/evolution/insights", handleEvolutionInsights(db))
		authorized.POST("/memories/evolution/run", handleEvolutionRun(aiSvc, db))

		aiGroup := authorized.Group("/ai")
		{
			aiGroup.GET("/config", handleAIConfig(aiRouter))
			aiGroup.PUT("/config", handleAIConfigUpdate(aiRouter))
			aiGroup.POST("/test", handleAITestConnection(aiRouter))
			aiGroup.GET("/usage", handleAIUsage(aiRouter))
			aiGroup.GET("/providers", handleAIProviders())

			aiGroup.POST("/extract", handleAIExtract(aiSvc))
			aiGroup.POST("/conflict-scan", handleAIConflictScan(aiSvc))
			aiGroup.POST("/decay-evaluate", handleAIDecayEvaluate(aiSvc))
			aiGroup.GET("/daily-report", handleAIDailyReport(aiSvc, db))
			aiGroup.POST("/wiki-generate", handleAIWikiGenerate(aiSvc))
			aiGroup.POST("/compress", handleAICompress(aiSvc))
			aiGroup.POST("/discover-relations", handleAIDiscoverRelations(aiSvc))
			aiGroup.POST("/smart-route", handleAISmartRoute(aiSvc))
			aiGroup.POST("/extract-facts", handleAIExtractFacts(aiSvc))
			aiGroup.POST("/consolidate", handleAIConsolidate(aiSvc))
			aiGroup.POST("/process-conversation", handleAIProcessConversation(aiSvc))
			aiGroup.POST("/context-assemble", handleAIAssembleContext(aiSvc))
			aiGroup.POST("/nudge-reflect", handleAINudgeReflect(aiSvc))
			aiGroup.POST("/self-refine", handleAISelfRefine(aiSvc))
			aiGroup.POST("/user-profile", handleAIUserProfile(aiSvc))
		}

		skillGroup := authorized.Group("/skills")
		{
			skillGroup.POST("/actions", handleSkillRecordAction(db))
			skillGroup.POST("/actions/batch", handleSkillRecordAction(db))
			skillGroup.GET("/detect", handleSkillDetectPatterns(db))
			skillGroup.POST("/create", handleSkillAutoCreate(db, aiSvc))
			skillGroup.GET("/list", handleSkillList(db))
			skillGroup.GET("/match", handleSkillMatch(db))
			skillGroup.PATCH("/:id", handleSkillPatch(db))
			skillGroup.POST("/:id/improve", handleSkillImprove(db, aiSvc))
			skillGroup.POST("/:id/usage", handleSkillRecordUsage(db))
			skillGroup.GET("/suggestions", handleSkillSuggestions(db))
			skillGroup.POST("/suggestions/generate", handleSkillSuggestions(db))
			skillGroup.POST("/suggestions/:id/dismiss", handleSkillSuggestions(db))
		}
	}

	external := r.Group("/api/v1/external")
	external.Use(middleware.APIKeyAuth(db), middleware.APIKeyRateLimit())
	{
		external.POST("/memories", handleExternalCreateMemory(db))
		external.POST("/memories/batch", handleExternalBatchCreateMemories(db))
		external.GET("/memories/search", handleExternalSearchMemories(db))
		external.GET("/memories/context", handleExternalMemoryContext(db))
		external.POST("/conversations", handleExternalPushConversation(db))
		external.POST("/conversations/batch", handleExternalBatchPushConversations(db))
		external.POST("/sessions/track", handleExternalSessionTrack(db))
		external.POST("/session-memories", handleExternalCreateSessionMemory(db))
		external.POST("/reason", handleExternalReason(db))
		external.POST("/ai/nudge-reflect", handleExternalAINudgeReflect(aiSvc))
		external.POST("/ai/process-conversation", handleExternalAIProcessConversation(aiSvc))
		external.POST("/skills/actions", handleExternalSkillRecordAction(db))
	}
}

type aiEmbeddingAdapter struct {
	router *ai.AIRouter
}

func (a *aiEmbeddingAdapter) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	resp, err := a.router.Embed(ctx, 1, texts)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return resp.Vectors, nil
}

func (a *aiEmbeddingAdapter) AIEmbed(ctx context.Context, userID uint, texts []string) ([][]float64, error) {
	resp, err := a.router.Embed(ctx, userID, texts)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return resp.Vectors, nil
}
