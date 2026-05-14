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
	proProvider := services.GetProProvider()
	aiRouter := ai.NewAIRouter(db)
	aiSvc := ai.NewAIService(aiRouter, db)
	embedAdapter := &aiEmbeddingAdapter{router: aiRouter, provider: proProvider}
	services.InitEmbeddingService(db, embedAdapter)
	services.SetEmbeddingAIRouter(embedAdapter)

	if lpp, ok := proProvider.(*services.LocalProProvider); ok {
		lpp.SetBackupPaths(cfg.BackupsDir, cfg.DatabasePath)
		lpp.SetAIService(aiSvc)
	}

	public := r.Group("/api/v1")
	{
		public.GET("/auth/init-status", handleInitStatus(authService))
		public.POST("/auth/set-password", middleware.LoginRateLimit(), handleSetPassword(authService))
		public.POST("/auth/login", middleware.LoginRateLimit(), handleLogin(authService))
		public.POST("/auth/register", middleware.LoginRateLimit(), handleRegister(authService))
		public.POST("/auth/register-with-invitation", middleware.LoginRateLimit(), handleRegisterWithInvitation(authService))
		public.POST("/auth/forgot-password", middleware.LoginRateLimit(), handleForgotPassword(authService))
		public.GET("/install-status", handleInstallStatus(db))
		public.GET("/check-update", handleCheckUpdate)
		public.GET("/health", handleHealthCheck(db))
		public.POST("/license/activate", middleware.LoginRateLimit(), handleLicenseActivate(proProvider))
	}

	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg, db), middleware.JWTRateLimit())
	{
		authorized.GET("/license/info", handleLicenseInfo(proProvider))
		authorized.POST("/license/deactivate", handleLicenseDeactivate(proProvider))
		authorized.GET("/auth/me", handleGetMe(authService))
		authorized.POST("/auth/change-password", handleChangePassword(authService))
		authorized.POST("/auth/reset-password", handleResetPassword(authService))
		authorized.POST("/auth/revoke-tokens", handleRevokeAllTokens(db, authService))

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
		authorized.GET("/memories/search/keyword", handleSearchKeyword(db))
		authorized.GET("/memories/search/semantic", handleSearchSemantic(db))
		authorized.GET("/memories/search/graph-rag", handleSearchGraphRAG(db))
		authorized.GET("/memories/:id/history", handleMemoryHistory(db))
		authorized.GET("/memories/:id/evolution", handleMemoryEvolution(db))
		authorized.GET("/sessions/memories", handleSessionMemories(db))
		authorized.PUT("/sessions/memories", handleSessionMemoryUpsert(db))

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
		authorized.POST("/memories/decay/apply", handleDecayApply(db))
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

		pro := authorized.Group("/pro")
		{
			pro.GET("/decay/stats", handleProDecayStats(proProvider, db))
			pro.POST("/decay/apply", handleProDecayApply(aiSvc, proProvider, db))
			pro.POST("/reinforce/:id", handleProReinforce(proProvider, db))
			pro.GET("/prune-suggest", handleProPruneSuggest(proProvider, db))
			pro.GET("/conflicts/scan", handleProConflictScan(aiSvc, proProvider, db))
			pro.POST("/conflicts/resolve/:index", handleProConflictResolve(proProvider, db))
			pro.GET("/token/route", handleProTokenRoute(aiSvc, proProvider, db))
			pro.GET("/token/stats", handleProTokenStats(proProvider, db))
			pro.POST("/ai/extract", handleProAIExtract(aiSvc, proProvider, db))
			pro.POST("/auto-graph", handleProAutoGraph(proProvider, db))
			pro.GET("/backup/schedule", handleProBackupSchedule(proProvider, db))
			pro.POST("/backup/schedule", handleProSetBackupSchedule(proProvider, db))

			pro.POST("/compress/preview", handleProCompressPreview(proProvider, db))
			pro.POST("/compress/apply", handleProCompressApply(aiSvc, proProvider, db))
			pro.GET("/compress/config", handleProCompressConfig(proProvider, db))
			pro.PUT("/compress/config", handleProSetCompressConfig(proProvider, db))

			pro.GET("/evolution/insights", handleProEvolutionInsights(proProvider, db))
			pro.POST("/evolution/discover", handleProEvolutionDiscover(aiSvc, proProvider, db))
			pro.POST("/evolution/infer", handleProEvolutionInfer(proProvider, db))
			pro.POST("/evolution/importance", handleProEvolutionImportance(proProvider, db))
			pro.POST("/evolution/prefetch", handleProEvolutionPrefetch(proProvider, db))
		}

		aiGroup := authorized.Group("/ai")
		{
			aiGroup.GET("/config", handleAIConfig(aiRouter, proProvider))
			aiGroup.PUT("/config", handleAIConfigUpdate(aiRouter, proProvider))
			aiGroup.POST("/test", handleAITestConnection(aiRouter, proProvider))
			aiGroup.GET("/usage", handleAIUsage(aiRouter))
			aiGroup.GET("/providers", handleAIProviders(proProvider))

			aiGroup.POST("/extract", handleAIExtract(aiSvc, proProvider))
			aiGroup.POST("/conflict-scan", handleAIConflictScan(aiSvc, proProvider))
			aiGroup.POST("/decay-evaluate", handleAIDecayEvaluate(aiSvc, proProvider))
			aiGroup.GET("/daily-report", handleAIDailyReport(aiSvc, proProvider, db))
			aiGroup.POST("/wiki-generate", handleAIWikiGenerate(aiSvc, proProvider))
			aiGroup.POST("/compress", handleAICompress(aiSvc, proProvider))
			aiGroup.POST("/discover-relations", handleAIDiscoverRelations(aiSvc, proProvider))
			aiGroup.POST("/smart-route", handleAISmartRoute(aiSvc, proProvider))
			aiGroup.POST("/extract-facts", handleAIExtractFacts(aiSvc, proProvider))
			aiGroup.POST("/consolidate", handleAIConsolidate(aiSvc, proProvider))
			aiGroup.POST("/process-conversation", handleAIProcessConversation(aiSvc, proProvider))
			aiGroup.POST("/context-assemble", handleAIAssembleContext(aiSvc, proProvider))
			aiGroup.POST("/nudge-reflect", handleAINudgeReflect(aiSvc, proProvider))
			aiGroup.POST("/self-refine", handleAISelfRefine(aiSvc, proProvider))
			aiGroup.POST("/user-profile", handleAIUserProfile(aiSvc, proProvider))
		}

		skillGroup := authorized.Group("/skills")
		{
			skillGroup.POST("/actions", handleSkillRecordAction(db))
			skillGroup.POST("/actions/batch", handleSkillRecordAction(db))
			skillGroup.GET("/detect", handleSkillDetectPatterns(db))
			skillGroup.POST("/create", handleSkillAutoCreate(db, aiSvc, proProvider))
			skillGroup.GET("/list", handleSkillList(db))
			skillGroup.GET("/match", handleSkillMatch(db))
			skillGroup.PATCH("/:id", handleSkillPatch(db))
			skillGroup.POST("/:id/improve", handleSkillImprove(db, aiSvc, proProvider))
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
	}
}

type aiEmbeddingAdapter struct {
	router   *ai.AIRouter
	provider services.ProProvider
}

func (a *aiEmbeddingAdapter) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	isPro := a.provider.IsPro()
	resp, err := a.router.Embed(ctx, 1, isPro, texts)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return resp.Vectors, nil
}

func (a *aiEmbeddingAdapter) AIEmbed(ctx context.Context, userID uint, isPro bool, texts []string) ([][]float64, error) {
	resp, err := a.router.Embed(ctx, userID, isPro, texts)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return resp.Vectors, nil
}
