<template>
  <div class="advanced-page">
    <div class="page-hero">
      <div class="hero-content">
        <h1>🚀 {{ $t('advanced.title') }}</h1>
        <span class="advanced-badge">{{ $t('nav.advancedVersion') }}</span>
      </div>
    </div>

    <div class="advanced-grid">
      <div class="mode-banner">
        <span>⚡ {{ $t('advanced.localAdvancedMode') }}</span>
      </div>

      <!-- Memory Decay -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'decay' }" id="advanced-decay">
        <div class="card-header">
          <span class="card-icon">📉</span>
          <span class="card-title">{{ $t('advanced.decay') }}</span>
          <div class="card-help" @mouseenter="showHelp.decay = true" @mouseleave="showHelp.decay = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.decay">
              <div class="help-title">{{ $t('advanced.decay') }}</div>
              <div class="help-text">{{ $t('advanced.decayHelp') }}</div>
              <div class="help-stages">
                <div class="help-stage"><span class="stage-dot green"></span>{{ $t('advanced.decayStage1') }}</div>
                <div class="help-stage"><span class="stage-dot yellow"></span>{{ $t('advanced.decayStage2') }}</div>
                <div class="help-stage"><span class="stage-dot orange"></span>{{ $t('advanced.decayStage3') }}</div>
                <div class="help-stage"><span class="stage-dot red"></span>{{ $t('advanced.decayStage4') }}</div>
              </div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="decay-logic-box">
            <div class="logic-title">📐 {{ $t('advanced.decayLogicTitle') }}</div>
            <div class="logic-rules">
              <div class="logic-rule"><span class="rule-action archive">Archive</span> {{ $t('advanced.decayRuleArchive') }}</div>
              <div class="logic-rule"><span class="rule-action trash">Trash</span> {{ $t('advanced.decayRuleTrash') }}</div>
              <div class="logic-rule"><span class="rule-action keep">Keep</span> {{ $t('advanced.decayRuleKeep') }}</div>
            </div>
          </div>
          <div class="stats-row" v-if="decayStats">
            <div class="stat-item">
              <span class="stat-value">{{ decayStats.total }}</span>
              <span class="stat-label">{{ $t('advanced.totalMemories') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value warn">{{ decayStats.archived }}</span>
              <span class="stat-label">{{ $t('advanced.archivedMemories') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value danger">{{ decayStats.trashed }}</span>
              <span class="stat-label">{{ $t('advanced.trashedMemories') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="loadDecayStats" :disabled="loading.decay" :loading="loading.decay">
              {{ $t('advanced.refreshStats') }}
            </el-button>
            <el-button @click="viewTrash">{{ $t('advanced.viewTrash') }}</el-button>
          </div>
          <div v-if="decayStats?.prune_candidates?.length" class="prune-section">
            <div class="sub-title">{{ $t('advanced.pruneSuggestions') }} ({{ decayStats.prune_candidates.length }})</div>
            <div class="prune-list">
              <div v-for="s in decayStats.prune_candidates.slice(0, 10)" :key="s.id" class="prune-item">
                <span class="prune-key">{{ s.key }}</span>
                <span class="prune-imp">{{ s.importance?.toFixed(2) }}</span>
              </div>
            </div>
          </div>

          <div v-if="communitiesResult" class="evolution-insights">
            <div class="sub-title">🏘️ Communities ({{ communitiesResult.total || 0 }})</div>
            <div v-for="(c, i) in (communitiesResult.communities || []).slice(0, 8)" :key="i" class="insight-section" style="margin-bottom:8px">
              <div class="insight-label" style="font-weight:600">{{ c.top_entity }} Domain ({{ c.size }} entities)</div>
              <div class="insight-list">
                <div v-for="(e, j) in (c.entities || []).slice(0, 6)" :key="j" class="insight-item">
                  <span class="insight-label">{{ e.name }}</span>
                  <el-tag size="small" type="info">{{ e.entity_type }}</el-tag>
                  <span class="prune-imp">{{ Math.round((e.confidence || 0) * 100) }}%</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="communitiesToWikiResult" class="evolution-insights">
            <div class="sub-title">📚 Wiki Generation Result</div>
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ communitiesToWikiResult.communities_found }}</span>
                <span class="stat-label">Communities</span>
              </div>
              <div class="stat-item">
                <span class="stat-value success">{{ communitiesToWikiResult.wiki_pages_created }}</span>
                <span class="stat-label">Pages Created</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ (communitiesToWikiResult.categories_created || []).length }}</span>
                <span class="stat-label">New Categories</span>
              </div>
            </div>
            <div v-if="(communitiesToWikiResult.categories_created || []).length" class="insight-section">
              <div class="insight-list">
                <el-tag v-for="cat in communitiesToWikiResult.categories_created" :key="cat" size="small" type="success" style="margin:2px">{{ cat }}</el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Conflict Scan -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'conflicts' }" id="advanced-conflicts">
        <div class="card-header">
          <span class="card-icon">⚠️</span>
          <span class="card-title">{{ $t('advanced.conflicts') }}</span>
          <span v-if="conflictSummary" class="mode-tag" :class="conflictMode === 'ai' ? 'mode-ai' : 'mode-local'">
            {{ conflictMode === 'ai' ? 'AI' : 'Local' }}
          </span>
          <div class="card-help" @mouseenter="showHelp.conflicts = true" @mouseleave="showHelp.conflicts = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.conflicts">
              <div class="help-title">{{ $t('advanced.conflicts') }}</div>
              <div class="help-text">{{ $t('advanced.conflictsHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="stats-row" v-if="conflictSummary">
            <div class="stat-item">
              <span class="stat-value warn">{{ conflictSummary.total }}</span>
              <span class="stat-label">{{ $t('advanced.conflictCount') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ conflictSummary.auto_resolvable }}</span>
              <span class="stat-label">{{ $t('advanced.autoResolvable') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value danger">{{ conflictSummary.needs_review }}</span>
              <span class="stat-label">{{ $t('advanced.needsReview') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="scanConflicts" :disabled="loading.conflicts" :loading="loading.conflicts">
              {{ $t('advanced.scanConflicts') }}
            </el-button>
          </div>
          <div v-if="conflicts.length" class="conflict-list">
            <div v-for="(c, i) in conflicts.slice(0, 10)" :key="i" class="conflict-item">
              <div class="conflict-key">{{ c.key || c.description }}</div>
              <div class="conflict-desc" v-if="c.description && c.key">{{ c.description }}</div>
              <div class="conflict-values" v-if="c.memories && c.memories.length >= 2">
                <span class="val-a">{{ c.memories[0].value?.substring(0, 80) }}</span>
                <span class="vs">vs</span>
                <span class="val-b">{{ c.memories[1].value?.substring(0, 80) }}</span>
              </div>
              <div class="conflict-suggestion" v-if="c.suggestion">
                <span class="suggestion-label">💡 {{ $t('advanced.suggestion') }}:</span>
                <span class="suggestion-text">{{ c.suggestion }}</span>
              </div>
              <div class="conflict-meta">
                <el-tag size="small" :type="c.severity === 'exact_duplicate' || c.severity === 'critical' ? 'danger' : c.severity === 'similar_content' || c.severity === 'warning' ? 'warning' : 'info'">
                  {{ c.severity }}
                </el-tag>
                <el-tag size="small" v-if="c.type" type="info">{{ c.type }}</el-tag>
                <el-button v-if="c.severity !== 'exact_duplicate' && c.severity !== 'critical' && c.memories?.length >= 2" size="small" type="primary" @click="resolveConflict(i, 'merge')">
                  {{ $t('advanced.merge') }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Smart Router -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'router' }" id="advanced-router">
        <div class="card-header">
          <span class="card-icon">🧠</span>
          <span class="card-title">{{ $t('advanced.smartRouter') }}</span>
          <div class="card-help" @mouseenter="showHelp.router = true" @mouseleave="showHelp.router = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.router">
              <div class="help-title">{{ $t('advanced.smartRouter') }}</div>
              <div class="help-text">{{ $t('advanced.smartRouterHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.smartRouterDesc') }}</p>
          <div class="router-test">
            <el-input v-model="testMessage" :placeholder="$t('advanced.testMessage')" size="small" />
            <el-button type="primary" @click="testRoute" :disabled="loading.route" :loading="loading.route">
              {{ $t('advanced.testRoute') }}
            </el-button>
          </div>
          <div v-if="routeResult" class="route-result">
            <div class="route-model">{{ $t('advanced.selectedModel') }}: <strong>{{ routeResult.selected_model }}</strong></div>
            <div class="route-complexity">{{ $t('advanced.complexity') }}: {{ routeResult.complexity }}</div>
          </div>
        </div>
      </div>

      <!-- Token Stats -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'tokenStats' }" id="advanced-tokenStats">
        <div class="card-header">
          <span class="card-icon">📊</span>
          <span class="card-title">{{ $t('advanced.tokenStats') }}</span>
          <div class="card-help" @mouseenter="showHelp.tokenStats = true" @mouseleave="showHelp.tokenStats = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.tokenStats">
              <div class="help-title">{{ $t('advanced.tokenStats') }}</div>
              <div class="help-text">{{ $t('advanced.tokenStatsHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="stats-row" v-if="tokenStatData">
            <div class="stat-item">
              <span class="stat-value">{{ tokenStatData.total_estimated_tokens }}</span>
              <span class="stat-label">{{ $t('advanced.totalTokens') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ tokenStatData.avg_tokens_per_memory }}</span>
              <span class="stat-label">{{ $t('advanced.avgTokens') }}</span>
            </div>
          </div>
          <div v-if="tokenStatData && tokenStatData.details" class="token-details">
            <div class="token-detail-item">
              <span class="token-label">📝 {{ $t('advanced.memoryTokens') }}</span>
              <span class="token-value">{{ tokenStatData.details.memory_tokens }} <small>({{ tokenStatData.details.memory_count }} memories × ~200)</small></span>
            </div>
            <div class="token-detail-item">
              <span class="token-label">🏷️ {{ $t('advanced.entityTokens') }}</span>
              <span class="token-value">{{ tokenStatData.details.entity_tokens }} <small>({{ tokenStatData.details.entity_count }} entities × ~50)</small></span>
            </div>
            <div class="token-detail-item">
              <span class="token-label">🔗 {{ $t('advanced.relationTokens') }}</span>
              <span class="token-value">{{ tokenStatData.details.relation_tokens }} <small>({{ tokenStatData.details.relation_count }} relations × ~30)</small></span>
            </div>
            <div class="token-note">* {{ $t('advanced.tokenEstimateNote') }}</div>
          </div>
          <div class="card-actions">
            <el-button @click="loadTokenStats" :disabled="loading.tokenStats" :loading="loading.tokenStats">
              {{ $t('advanced.refreshTokens') }}
            </el-button>
          </div>
        </div>
      </div>

      <!-- AI Extract -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'extract' }" id="advanced-extract">
        <div class="card-header">
          <span class="card-icon">🔍</span>
          <span class="card-title">{{ $t('advanced.aiExtract') }}</span>
          <span v-if="extractResult" class="mode-tag" :class="extractResult.mode === 'ai' ? 'mode-ai' : 'mode-local'">
            {{ extractResult.mode === 'ai' ? 'AI' : 'Local' }}
          </span>
          <div class="card-help" @mouseenter="showHelp.extract = true" @mouseleave="showHelp.extract = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.extract">
              <div class="help-title">{{ $t('advanced.aiExtract') }}</div>
              <div class="help-text">{{ $t('advanced.aiExtractHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.aiExtractDesc') }}</p>
          <div class="card-actions">
            <el-button type="primary" @click="runAiExtract" :disabled="loading.extract" :loading="loading.extract">
              {{ $t('advanced.runExtract') }}
            </el-button>
          </div>
          <div v-if="extractResult" class="extract-result">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value success">{{ extractResult.entities?.length || extractResult.entities_extracted || 0 }}</span>
                <span class="stat-label">{{ $t('advanced.entitiesExtracted') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ extractResult.relations?.length || extractResult.relations_extracted || 0 }}</span>
                <span class="stat-label">{{ $t('advanced.relationsExtracted') }}</span>
              </div>
            </div>
            <div v-if="extractResult.entities && extractResult.entities.length > 0" class="extract-section">
              <div class="sub-title">{{ $t('advanced.entitiesExtracted') }} ({{ extractResult.entities.length }})</div>
              <div class="extract-list">
                <div v-for="(e, i) in extractResult.entities.slice(0, 20)" :key="i" class="extract-item">
                  <span class="extract-name">{{ e.name }}</span>
                  <el-tag size="small" type="info">{{ e.entity_type || e.type }}</el-tag>
                  <span class="extract-conf">{{ Math.round((e.confidence || 0) * 100) }}%</span>
                </div>
              </div>
            </div>
            <div v-if="extractResult.relations && extractResult.relations.length > 0" class="extract-section">
              <div class="sub-title">{{ $t('advanced.relationsExtracted') }} ({{ extractResult.relations.length }})</div>
              <div class="extract-list">
                <div v-for="(r, i) in extractResult.relations.slice(0, 15)" :key="i" class="extract-item relation-item">
                  <span class="rel-source">{{ r.source }}</span>
                  <span class="rel-arrow">→</span>
                  <el-tag size="small" type="warning">{{ r.type }}</el-tag>
                  <span class="rel-arrow">→</span>
                  <span class="rel-target">{{ r.target }}</span>
                  <span class="extract-conf">{{ Math.round((r.confidence || 0) * 100) }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quality Assess & Auto Fix -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'quality' }" id="advanced-quality">
        <div class="card-header">
          <span class="card-icon">🩺</span>
          <span class="card-title">{{ $t('advanced.qualityAssess') }}</span>
          <div class="card-help" @mouseenter="showHelp.quality = true" @mouseleave="showHelp.quality = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.quality">
              <div class="help-title">{{ $t('advanced.qualityAssess') }}</div>
              <div class="help-text">{{ $t('advanced.qualityAssessHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.qualityAssessDesc') }}</p>
          <div class="card-actions">
            <el-button @click="assessQuality" :disabled="loading.quality" :loading="loading.quality">
              {{ $t('advanced.assessQuality') }}
            </el-button>
            <el-button type="primary" @click="autoFix" :disabled="loading.autoFix || !qualityResult" :loading="loading.autoFix">
              {{ $t('advanced.autoFix') }}
            </el-button>
          </div>
          <div v-if="qualityResult" class="quality-result">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value warn">{{ qualityResult.total }}</span>
                <span class="stat-label">{{ $t('advanced.qualityIssues') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value success">{{ qualityResult.auto_fixable }}</span>
                <span class="stat-label">{{ $t('advanced.autoFixable') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ qualityResult.memory_count }}</span>
                <span class="stat-label">{{ $t('advanced.totalMemories') }}</span>
              </div>
            </div>
            <div v-if="qualityResult.issues && qualityResult.issues.length > 0" class="extract-section">
              <div class="sub-title">{{ $t('advanced.qualityIssues') }} ({{ qualityResult.issues.length }})</div>
              <div class="extract-list">
                <div v-for="(issue, i) in qualityResult.issues.slice(0, 15)" :key="i" class="extract-item">
                  <span class="extract-name">{{ issue.key }}</span>
                  <el-tag size="small" :type="issue.severity === 'high' ? 'danger' : issue.severity === 'medium' ? 'warning' : 'info'">{{ issue.issue_type }}</el-tag>
                  <el-tag size="small" v-if="issue.auto_fix" type="success">{{ $t('advanced.fixable') }}</el-tag>
                  <span class="extract-conf" style="color: var(--cm-text-muted); font-size: 11px">{{ issue.detail }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-if="autoFixResult" class="quality-result" style="margin-top: 8px">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value success">{{ autoFixResult.fixed }}</span>
                <span class="stat-label">{{ $t('advanced.fixedCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ autoFixResult.skipped }}</span>
                <span class="stat-label">{{ $t('advanced.skippedFix') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value danger">{{ autoFixResult.failed }}</span>
                <span class="stat-label">{{ $t('advanced.failedFix') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Auto Graph -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'graph' }" id="advanced-graph">
        <div class="card-header">
          <span class="card-icon">🕸️</span>
          <span class="card-title">{{ $t('advanced.autoGraph') }}</span>
          <div class="card-help" @mouseenter="showHelp.graph = true" @mouseleave="showHelp.graph = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.graph">
              <div class="help-title">{{ $t('advanced.autoGraph') }}</div>
              <div class="help-text">{{ $t('advanced.autoGraphHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.autoGraphDesc') }}</p>
          <div class="card-actions">
            <el-button @click="runAutoGraph(false)" :disabled="loading.graph" :loading="loading.graph">
              {{ $t('advanced.generateGraph') }}
            </el-button>
            <el-button type="danger" @click="runAutoGraph(true)" :disabled="loading.graph" :loading="loading.graph">
              {{ $t('advanced.regenerateGraph') }}
            </el-button>
          </div>
          <div v-if="graphResult" class="graph-result">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value success">{{ graphResult.created || 0 }}</span>
                <span class="stat-label">{{ $t('advanced.newRelations') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ graphResult.total_pairs || 0 }}</span>
                <span class="stat-label">{{ $t('advanced.existingRelations') }}</span>
              </div>
              <div class="stat-item" v-if="graphResult.skipped">
                <span class="stat-value warn">{{ graphResult.skipped }}</span>
                <span class="stat-label">{{ $t('advanced.skippedRelations') }}</span>
              </div>
            </div>
            <div v-if="graphResult.relations && graphResult.relations.length > 0" class="graph-section">
              <div class="sub-title">{{ $t('advanced.newRelations') }} ({{ graphResult.relations.length }})</div>
              <div class="graph-list">
                <div v-for="(r, i) in graphResult.relations.slice(0, 15)" :key="i" class="graph-item">
                  <span class="graph-source">{{ r.source_name || r.source }}</span>
                  <span class="graph-arrow">→</span>
                  <el-tag size="small" type="success">{{ r.relation_type || r.type }}</el-tag>
                  <span class="graph-arrow">→</span>
                  <span class="graph-target">{{ r.target_name || r.target }}</span>
                  <span class="graph-weight" v-if="r.weight">{{ (r.weight * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Memory Compression -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'compress' }" id="advanced-compress">
        <div class="card-header">
          <span class="card-icon">🗜️</span>
          <span class="card-title">{{ $t('advanced.compress') }}</span>
          <div class="card-help" @mouseenter="showHelp.compress = true" @mouseleave="showHelp.compress = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.compress">
              <div class="help-title">{{ $t('advanced.compress') }}</div>
              <div class="help-text">{{ $t('advanced.compressHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.compressDesc') }}</p>
          <div class="compress-levels">
            <div class="level-option" :class="{ active: compressLevel === 'light' }" @click="compressLevel = 'light'">
              <div class="level-name">{{ $t('advanced.compressLight') }}</div>
              <div class="level-rate">{{ $t('advanced.compressLightRate') }}</div>
              <div class="level-desc">{{ $t('advanced.compressLightDesc') }}</div>
            </div>
            <div class="level-option" :class="{ active: compressLevel === 'medium' }" @click="compressLevel = 'medium'">
              <div class="level-name">{{ $t('advanced.compressMedium') }}</div>
              <div class="level-rate">{{ $t('advanced.compressMediumRate') }}</div>
              <div class="level-desc">{{ $t('advanced.compressMediumDesc') }}</div>
            </div>
            <div class="level-option" :class="{ active: compressLevel === 'deep' }" @click="compressLevel = 'deep'">
              <div class="level-name">{{ $t('advanced.compressDeep') }}</div>
              <div class="level-rate">{{ $t('advanced.compressDeepRate') }}</div>
              <div class="level-desc">{{ $t('advanced.compressDeepDesc') }}</div>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="previewCompress" :disabled="loading.compressPreview" :loading="loading.compressPreview">
              {{ $t('advanced.previewCompress') }}
            </el-button>
            <el-button type="primary" @click="applyCompress" :disabled="loading.compressApply" :loading="loading.compressApply">
              {{ $t('advanced.applyCompress') }}
            </el-button>
          </div>
          <div v-if="compressPreviewData" class="compress-result">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ compressPreviewData.total }}</span>
                <span class="stat-label">{{ $t('advanced.compressedResult') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ compressPreviewData.threshold }}</span>
                <span class="stat-label">{{ $t('advanced.compressRatio') }}</span>
              </div>
            </div>
            <div v-if="compressPreviewData.preview" class="compress-details">
              <div v-for="(d, i) in compressPreviewData.preview.slice(0, 5)" :key="i" class="compress-detail-item">
                <span class="detail-action">{{ d.action }}</span>
                <span class="detail-target">{{ d.key }} (importance: {{ d.importance?.toFixed(2) }})</span>
              </div>
            </div>
          </div>
          <div class="auto-compress-setting">
            <div class="setting-item">
              <span>{{ $t('advanced.autoCompress') }}</span>
              <el-switch v-model="compressConfig.auto_enabled" @change="saveCompressConfig" />
            </div>
            <div class="setting-item" v-if="compressConfig.auto_enabled">
              <span>{{ $t('advanced.compressThreshold') }}</span>
              <el-select v-model="compressConfig.threshold" size="small" @change="saveCompressConfig" style="width: 140px">
                <el-option :label="$t('advanced.threshold500')" :value="500" />
                <el-option :label="$t('advanced.threshold1000')" :value="1000" />
                <el-option :label="$t('advanced.threshold2000')" :value="2000" />
              </el-select>
            </div>
          </div>
        </div>
      </div>

      <!-- Evolution Engine -->
      <div class="advanced-card" :class="{ 'section-highlight': activeSection === 'evolution' }" id="advanced-evolution">
        <div class="card-header">
          <span class="card-icon">🧬</span>
          <span class="card-title">{{ $t('advanced.evolution') }}</span>
          <div class="card-help" @mouseenter="showHelp.evolution = true" @mouseleave="showHelp.evolution = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.evolution">
              <div class="help-title">{{ $t('advanced.evolution') }}</div>
              <div class="help-text">{{ $t('advanced.evolutionHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('advanced.evolutionDesc') }}</p>
          <div class="evolution-actions">
            <el-button @click="loadEvolutionInsights" :disabled="loading.insights" :loading="loading.insights">
              {{ $t('advanced.evolutionInsights') }}
            </el-button>
            <el-button @click="runDiscoverRelations" :disabled="loading.discover" :loading="loading.discover">
              {{ $t('advanced.discoverRelations') }}
            </el-button>
            <el-button @click="runInferChains" :disabled="loading.infer" :loading="loading.infer">
              {{ $t('advanced.inferChains') }}
            </el-button>
            <el-button @click="runImportanceAdjust" :disabled="loading.importance" :loading="loading.importance">
              {{ $t('advanced.importanceAdjust') }}
            </el-button>
            <el-button type="success" @click="runGraphReasoning" :disabled="loading.graphReasoning" :loading="loading.graphReasoning" plain>
              🕸️ Graph Reasoning
            </el-button>
            <el-button type="warning" @click="runCentrality" :disabled="loading.centrality" :loading="loading.centrality" plain>
              🎯 Centrality
            </el-button>
            <el-button type="info" @click="runCommunityDiscovery" :disabled="loading.communities" :loading="loading.communities" plain>
              🏘️ Communities
            </el-button>
            <el-button type="success" @click="runCommunitiesToWiki" :disabled="loading.communitiesToWiki" :loading="loading.communitiesToWiki" plain>
              📚 To Wiki
            </el-button>
          </div>
          <div v-if="evolutionInsights" class="evolution-insights">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ evolutionInsights.total_memories }}</span>
                <span class="stat-label">{{ $t('advanced.totalMemoriesLabel') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value success">{{ evolutionInsights.relations_count }}</span>
                <span class="stat-label">{{ $t('advanced.relationsCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ evolutionInsights.discovered_relations }}</span>
                <span class="stat-label">{{ $t('advanced.newDiscoveries') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ evolutionInsights.inferred_chains }}</span>
                <span class="stat-label">{{ $t('advanced.inferChainsLabel') }}</span>
              </div>
            </div>
            <div v-if="evolutionInsights.layer_stats && Object.keys(evolutionInsights.layer_stats).length > 0" class="insight-section">
              <div class="sub-title">{{ $t('advanced.layerDistribution') }}</div>
              <div class="insight-list">
                <div v-for="(count, layer) in evolutionInsights.layer_stats" :key="layer" class="insight-item">
                  <span class="insight-label">{{ layer }}</span>
                  <span class="insight-value">{{ count }}</span>
                </div>
              </div>
            </div>
            <div v-if="evolutionInsights.source_stats && Object.keys(evolutionInsights.source_stats).length > 0" class="insight-section">
              <div class="sub-title">{{ $t('advanced.sourceDistribution') }}</div>
              <div class="insight-list">
                <div v-for="(count, source) in evolutionInsights.source_stats" :key="source" class="insight-item">
                  <span class="insight-label">{{ source }}</span>
                  <span class="insight-value">{{ count }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-if="discoverResult" class="discover-result">
            <div class="sub-title">{{ $t('advanced.discoveredRelations') }} ({{ discoverResult.discoveries?.length || 0 }})</div>
            <div class="discover-list">
              <div v-for="(r, i) in (discoverResult.discoveries || []).slice(0, 8)" :key="i" class="discover-item">
                <span class="discover-type">{{ r.relation_type }}</span>
                <span class="discover-arrow">×</span>
                <span class="discover-count">{{ r.count }}</span>
                <el-tag size="small" type="info">{{ Math.round(r.confidence * 100) }}%</el-tag>
              </div>
            </div>
          </div>
          <div v-if="inferResult" class="infer-result">
            <div class="sub-title">{{ $t('advanced.inferChainsTitle') }} ({{ inferResult.inferences?.length || 0 }})</div>
            <div class="chain-list">
              <div v-for="(c, i) in (inferResult.inferences || []).slice(0, 5)" :key="i" class="chain-item">
                <div class="chain-nodes">
                  <span class="chain-node">{{ c.entity_name }}</span>
                </div>
                <div class="chain-conclusion">
                  {{ c.reason }} ({{ Math.round(c.confidence * 100) }}%)
                </div>
              </div>
            </div>
          </div>
          <div v-if="graphReasoningResult" class="discover-result">
            <div class="sub-title">🕸️ Inferred Relations via Graph ({{ graphReasoningResult.total_inferred || 0 }})</div>
            <div class="discover-list">
              <div v-for="(r, i) in (graphReasoningResult.inferred_relations || []).slice(0, 10)" :key="i" class="discover-item">
                <span class="discover-type">{{ r.source }}</span>
                <span class="discover-arrow">→</span>
                <span class="discover-type">{{ r.target }}</span>
                <el-tag size="small" type="warning">{{ r.relation_type }}</el-tag>
                <span class="prune-imp" style="margin-left:4px">via {{ r.via }}</span>
                <el-tag size="small" type="info">{{ Math.round((r.confidence || 0) * 100) }}%</el-tag>
              </div>
            </div>
          </div>
          <div v-if="centralityResult" class="evolution-insights">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ centralityResult.total_entities }}</span>
                <span class="stat-label">Entities</span>
              </div>
              <div class="stat-item">
                <span class="stat-value success">{{ centralityResult.hub_count }}</span>
                <span class="stat-label">Hubs</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ centralityResult.isolated_count }}</span>
                <span class="stat-label">Isolated</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ (centralityResult.avg_degree || 0).toFixed(1) }}</span>
                <span class="stat-label">Avg Degree</span>
              </div>
            </div>
            <div v-if="centralityResult.top_entities && centralityResult.top_entities.length > 0" class="insight-section">
              <div class="sub-title">🎯 Top Central Entities</div>
              <div class="insight-list">
                <div v-for="(e, i) in centralityResult.top_entities.slice(0, 10)" :key="i" class="insight-item">
                  <span class="insight-label">{{ e.name }}</span>
                  <el-tag size="small" type="info">{{ e.entity_type }}</el-tag>
                  <span class="prune-imp">deg:{{ e.degree }}</span>
                  <span class="prune-imp">{{ (e.score || 0).toFixed(2) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ElMessageBox } from 'element-plus'
import { toolboxApi } from '../api/go-toolbox'
import { memoryApi } from '../api/go-memories'
import { translateError, extractApiError } from '../i18n'

const { t } = useI18n()
const route = useRoute()

const loading = ref<Record<string, boolean>>({})
const activeSection = ref((route.query.section as string) || '')
const showHelp = reactive<Record<string, boolean>>({
  decay: false, conflicts: false, router: false, tokenStats: false,
  extract: false, graph: false, compress: false, evolution: false, quality: false,
})

const decayStats = ref<any>(null)

const conflicts = ref<any[]>([])
const conflictSummary = ref<any>(null)
const conflictMode = ref('local')

const tokenStatData = ref<any>(null)
const testMessage = ref('')
const routeResult = ref<any>(null)

const extractResult = ref<any>(null)

const qualityResult = ref<any>(null)
const autoFixResult = ref<any>(null)

const graphResult = ref<any>(null)

const compressLevel = ref<'light' | 'medium' | 'deep'>('light')
const compressPreviewData = ref<any>(null)
const compressConfig = ref<any>({ auto_enabled: false, threshold: 1000, level: 'light' })

const evolutionInsights = ref<any>(null)
const discoverResult = ref<any>(null)
const inferResult = ref<any>(null)
const graphReasoningResult = ref<any>(null)
const centralityResult = ref<any>(null)
const communitiesResult = ref<any>(null)
const communitiesToWikiResult = ref<any>(null)
onMounted(async () => {
  loadDecayStats()
  loadTokenStats()
  loadCompressConfig()
  if (activeSection.value) {
    nextTick(() => scrollToSection(activeSection.value))
  }
})

watch(() => route.query.section, (section) => {
  if (section && typeof section === 'string') {
    activeSection.value = section
    nextTick(() => scrollToSection(section))
  }
})

function scrollToSection(section: string) {
  const el = document.getElementById(`advanced-${section}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('section-highlight')
    setTimeout(() => el.classList.remove('section-highlight'), 2000)
  }
}

async function loadDecayStats() {
  loading.value.decay = true
  try {
    const { data } = await memoryApi.getDecayStats()
    decayStats.value = data
  } catch (e: unknown) {
    if ((e as { response?: { status?: number } })?.response?.status !== 403) ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.decay = false }
}

function viewTrash() {
  window.location.href = '/trash'
}

async function scanConflicts() {
  loading.value.conflicts = true
  try {
    const { data } = await toolboxApi.scanConflicts()
    conflicts.value = data.conflicts
    conflictMode.value = data.mode || 'local'
    conflictSummary.value = { total: data.total, auto_resolvable: data.auto_resolvable || 0, needs_review: data.needs_review || data.total }
  } catch (e: unknown) {
    if ((e as { response?: { status?: number } })?.response?.status !== 403) ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.conflicts = false }
}

async function resolveConflict(index: number, strategy: string) {
  try {
    await toolboxApi.resolveConflict(index, strategy)
    ElMessage.success(t('advanced.conflictResolved'))
    scanConflicts()
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  }
}

async function loadTokenStats() {
  loading.value.tokenStats = true
  try {
    const { data } = await toolboxApi.getTokenStats()
    tokenStatData.value = {
      total_estimated_tokens: data.total_tokens,
      avg_tokens_per_memory: data.memory_count > 0 ? Math.round(data.total_tokens / data.memory_count) : 0,
      details: {
        memory_tokens: data.memory_tokens || 0,
        entity_tokens: data.entity_tokens || 0,
        relation_tokens: data.relation_tokens || 0,
        memory_count: data.memory_count || 0,
        entity_count: data.entity_count || 0,
        relation_count: data.relation_count || 0,
      }
    }
  } catch (e: unknown) {
    if ((e as { response?: { status?: number } })?.response?.status !== 403) ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.tokenStats = false }
}

async function testRoute() {
  if (!testMessage.value) return
  loading.value.route = true
  try {
    const { data } = await toolboxApi.routeModel(testMessage.value)
    routeResult.value = { selected_model: data.recommended_layer || data.strategy, complexity: data.complexity || 'simple' }
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.route = false }
}

async function runAiExtract() {
  loading.value.extract = true
  try {
    const { data } = await toolboxApi.aiExtract()
    extractResult.value = data
    const entityCount = data.entities?.length || data.extracted || 0
    const relationCount = data.relations?.length || 0
    ElMessage.success(t('advanced.extractDone', { entities: entityCount, relations: relationCount }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.extract = false }
}

async function assessQuality() {
  loading.value.quality = true
  try {
    const { data } = await memoryApi.assessQuality()
    qualityResult.value = data
    const issueCount = data.total || 0
    if (issueCount > 0) {
      ElMessage.warning(t('advanced.qualityIssuesFound', { count: issueCount }))
    } else {
      ElMessage.success(t('advanced.qualityNoIssues'))
    }
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.quality = false }
}

async function autoFix() {
  loading.value.autoFix = true
  try {
    const { data } = await memoryApi.autoFix()
    autoFixResult.value = data
    ElMessage.success(t('advanced.autoFixDone', { fixed: data.fixed || 0 }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.autoFix = false }
}

async function runAutoGraph(overwrite: boolean) {
  loading.value.graph = true
  try {
    const { data } = await toolboxApi.autoGraph(overwrite)
    graphResult.value = data
    ElMessage.success(t('advanced.graphDone', { entities: data.total_pairs || data.total || 0, relations: data.created || 0 }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.graph = false }
}

async function previewCompress() {
  loading.value.compressPreview = true
  try {
    const { data } = await toolboxApi.compressPreview(compressLevel.value)
    compressPreviewData.value = data
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.compressPreview = false }
}

async function applyCompress() {
  loading.value.compressApply = true
  try {
    const { data } = await toolboxApi.compressApply(compressLevel.value)
    ElMessage.success(t('advanced.compressDone', { count: data.archived, ratio: data.total > 0 ? Math.round(data.archived / data.total * 100) : 0 }))
    compressPreviewData.value = null
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.compressApply = false }
}

async function saveCompressConfig() {
  try {
    await toolboxApi.setCompressConfig(compressConfig.value)
    ElMessage.success(t('common.success'))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  }
}

async function loadCompressConfig() {
  try {
    const { data } = await toolboxApi.getCompressConfig()
    const cfg = data.config || data
    compressConfig.value = { auto_enabled: cfg.auto_compress || false, threshold: cfg.threshold || 1000, level: cfg.level || 'light' }
  } catch { compressConfig.value = { auto_enabled: false, threshold: 1000, level: 'light' } }
}

async function loadEvolutionInsights() {
  loading.value.insights = true
  try {
    const { data } = await memoryApi.getEvolutionInsights()
    evolutionInsights.value = {
      total_memories: data.total,
      relations_count: data.relations_count || 0,
      discovered_relations: data.discovered_relations || 0,
      inferred_chains: data.inferred_chains || 0,
      layer_stats: data.layer_stats || {},
      source_stats: data.source_stats || {},
    }
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.insights = false }
}

async function runDiscoverRelations() {
  loading.value.discover = true
  try {
    const { data } = await memoryApi.runEvolution('discover')
    discoverResult.value = data
    ElMessage.success(t('advanced.discoveredCount', { count: data.discoveries?.length || 0 }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.discover = false }
}

async function runInferChains() {
  loading.value.infer = true
  try {
    const { data } = await memoryApi.runEvolution('infer')
    inferResult.value = data
    ElMessage.success(t('advanced.inferredCount', { count: data.inferences?.length || 0 }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.infer = false }
}

async function runImportanceAdjust() {
  loading.value.importance = true
  try {
    const { data } = await memoryApi.runEvolution('importance')
    ElMessage.success(t('advanced.adjustedCount', { count: data.total || 0 }))
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.importance = false }
}

async function runGraphReasoning() {
  loading.value.graphReasoning = true
  try {
    const { data } = await memoryApi.getGraphReasoning()
    graphReasoningResult.value = data
    ElMessage.success(`Inferred ${data.total_inferred || 0} relations via graph reasoning`)
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.graphReasoning = false }
}

async function runCentrality() {
  loading.value.centrality = true
  try {
    const { data } = await memoryApi.getCentrality()
    centralityResult.value = data
    ElMessage.success(`Analyzed ${data.total_entities || 0} entities, ${data.hub_count || 0} hubs`)
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.centrality = false }
}

async function runCommunityDiscovery() {
  loading.value.communities = true
  try {
    const { data } = await memoryApi.getCommunities()
    communitiesResult.value = data
    ElMessage.success(`Found ${data.total || 0} communities`)
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.communities = false }
}

async function runCommunitiesToWiki() {
  loading.value.communitiesToWiki = true
  try {
    const { data } = await memoryApi.communitiesToWiki()
    communitiesToWikiResult.value = data
    ElMessage.success(`Created ${data.wiki_pages_created || 0} wiki pages from ${data.communities_found || 0} communities`)
  } catch (e: unknown) {
    ElMessage.error(extractApiError(e, t('common.failed')))
  } finally { loading.value.communitiesToWiki = false }
}

</script>

<style scoped>
.advanced-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--cm-bg-secondary, #f5f5f5);
}

.page-hero {
  background: var(--cm-bg-primary, #fff);
  padding: 24px 28px;
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.hero-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hero-content h1 {
  font-size: 24px;
  font-weight: 800;
  color: var(--cm-text, #1a1a1a);
  margin: 0;
}

.advanced-badge {
  background: rgba(16,185,129,0.15);
  color: #10B981;
  padding: 2px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}

.advanced-grid {
  padding: 24px 28px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
  align-content: start;
}

.mode-banner {
  grid-column: 1 / -1;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 10px;
  padding: 12px 16px;
}

.mode-banner span:first-child {
  font-weight: 600;
  color: #3B82F6;
  font-size: 14px;
}

.advanced-card {
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 14px;
  transition: all 0.25s ease;
}

.advanced-card:hover {
  border-color: var(--cm-primary, #10b981);
  box-shadow: 0 8px 24px rgba(16,185,129,0.12);
  transform: translateY(-2px);
}

.advanced-card.section-highlight {
  border-color: var(--cm-primary, #10b981);
  box-shadow: 0 0 0 2px rgba(16,185,129,0.2);
  transition: all 0.3s ease;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
  position: relative;
}

.card-icon {
  font-size: 18px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text, #1a1a1a);
  flex: 1;
}

.card-help {
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
}

.help-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--cm-bg-tertiary, #f0f0f0);
  color: var(--cm-text-secondary, #666);
  font-size: 13px;
  font-weight: 600;
  transition: all 0.2s;
}

.card-help:hover .help-icon {
  background: var(--cm-primary, #10b981);
  color: #fff;
}

.help-popup {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  width: 300px;
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 12px 32px rgba(0,0,0,0.15);
  z-index: 100;
  font-size: 13px;
}

.help-title {
  font-weight: 600;
  color: var(--cm-text, #1a1a1a);
  margin-bottom: 8px;
  font-size: 14px;
}

.help-text {
  color: var(--cm-text-secondary, #666);
  line-height: 1.6;
  margin-bottom: 8px;
}

.help-stages {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.help-stage {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--cm-text-secondary, #666);
}

.stage-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stage-dot.green { background: #22c55e; }
.stage-dot.yellow { background: #eab308; }
.stage-dot.orange { background: #f97316; }
.stage-dot.red { background: #ef4444; }

.card-body {
  padding: 16px 18px;
}

.card-desc {
  color: var(--cm-text-secondary, #666);
  font-size: 13px;
  margin: 0 0 12px;
  line-height: 1.6;
}

.card-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.stats-row {
  display: flex;
  gap: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--cm-text, #1a1a1a);
}

.stat-value.warn { color: #d29922; }
.stat-value.danger { color: #f85149; }
.stat-value.success { color: #10B981; }
.stat-label {
  font-size: 11px;
  color: var(--cm-text-muted, #999);
  margin-top: 2px;
}

.prune-section { margin-top: 14px; }
.sub-title { font-size: 13px; color: var(--cm-text-muted); margin-bottom: 8px; }
.prune-list { max-height: 150px; overflow-y: auto; }
.prune-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-size: 12px; color: var(--cm-text-secondary); }
.prune-key { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prune-imp { color: #d29922; font-weight: 600; }

.conflict-list { margin-top: 12px; }
.conflict-item { padding: 8px 0; border-bottom: 1px solid var(--cm-border, #e5e5e5); }
.conflict-key { font-size: 13px; font-weight: 600; color: var(--cm-text); }
.conflict-values { display: flex; align-items: center; gap: 6px; margin: 4px 0; font-size: 12px; color: var(--cm-text-muted); }
.vs { color: #d29922; font-weight: 600; }
.val-a, .val-b { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.conflict-meta { display: flex; align-items: center; gap: 8px; }

.router-test { display: flex; gap: 8px; margin-top: 12px; }
.router-test .el-input { flex: 1; }
.route-result { margin-top: 10px; font-size: 13px; color: var(--cm-text-secondary); }
.route-model { margin-bottom: 4px; }
.route-model strong { color: #10B981; }

.extract-result, .graph-result { display: flex; flex-direction: column; gap: 12px; margin-top: 12px; }

.mode-tag {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
}
.mode-ai { background: rgba(16,185,129,0.15); color: #10B981; }
.mode-local { background: rgba(144,164,174,0.15); color: #90A4AE; }

.extract-section, .graph-section, .insight-section { margin-top: 8px; }
.extract-list, .graph-list, .insight-list { display: flex; flex-direction: column; gap: 6px; margin-top: 6px; }
.extract-item, .graph-item, .insight-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: var(--cm-bg, #fff);
  border-radius: 6px;
  font-size: 13px;
}
.extract-name { font-weight: 600; color: var(--cm-text); min-width: 80px; }
.extract-conf { color: var(--cm-text-muted); font-size: 11px; margin-left: auto; }
.relation-item { flex-wrap: wrap; }
.rel-source, .rel-target { color: var(--cm-text); font-weight: 500; }
.rel-arrow { color: var(--cm-text-muted); font-size: 12px; }
.graph-source, .graph-target { color: var(--cm-text); font-weight: 500; }
.graph-arrow { color: var(--cm-text-muted); font-size: 12px; }
.graph-weight { color: var(--cm-text-muted); font-size: 11px; margin-left: auto; }
.graph-info { margin-top: 4px; }
.info-text { font-size: 12px; color: var(--cm-text-muted); }
.insight-label { color: var(--cm-text); font-weight: 500; }
.insight-value { color: var(--cm-primary, #10B981); font-weight: 600; margin-left: auto; }

.decay-logic-box {
  background: var(--cm-bg, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 8px;
  padding: 10px 12px;
  margin-bottom: 12px;
}
.logic-title { font-size: 13px; font-weight: 600; color: var(--cm-text); margin-bottom: 6px; }
.logic-rules { display: flex; flex-direction: column; gap: 4px; }
.logic-rule { font-size: 12px; color: var(--cm-text-secondary); display: flex; align-items: center; gap: 8px; }
.rule-action {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  min-width: 56px;
  text-align: center;
}
.rule-action.archive { background: rgba(250,173,20,0.15); color: #FAAD14; }
.rule-action.trash { background: rgba(255,77,79,0.15); color: #FF4D4F; }
.rule-action.keep { background: rgba(16,185,129,0.15); color: #10B981; }

.conflict-desc { font-size: 12px; color: var(--cm-text-secondary); margin-top: 4px; }
.conflict-suggestion {
  margin-top: 6px;
  padding: 6px 10px;
  background: rgba(16,185,129,0.06);
  border-radius: 6px;
  font-size: 12px;
}
.suggestion-label { color: var(--cm-text); font-weight: 500; }
.suggestion-text { color: var(--cm-text-secondary); }

.token-details { margin-top: 8px; display: flex; flex-direction: column; gap: 6px; }
.token-detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  font-size: 13px;
}
.token-label { color: var(--cm-text-secondary); }
.token-value { color: var(--cm-text); font-weight: 500; }
.token-value small { color: var(--cm-text-muted); font-weight: 400; }
.token-note { font-size: 11px; color: var(--cm-text-muted); margin-top: 4px; font-style: italic; }

.backup-schedule .setting-item,
.auto-compress-setting .setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  color: var(--cm-text);
  font-size: 14px;
}

.compress-levels { display: flex; gap: 10px; margin: 12px 0; }
.level-option {
  flex: 1;
  padding: 10px 8px;
  border: 2px solid var(--cm-border, #e5e5e5);
  border-radius: 10px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
}
.level-option:hover { border-color: rgba(16,185,129,0.4); }
.level-option.active { border-color: var(--cm-primary, #10b981); background: rgba(16,185,129,0.08); }
.level-name { font-size: 14px; font-weight: 600; color: var(--cm-text); }
.level-rate { font-size: 18px; font-weight: 700; color: var(--cm-primary, #10b981); margin: 4px 0; }
.level-desc { font-size: 11px; color: var(--cm-text-muted); }

.compress-result { margin-top: 12px; }
.compress-details { margin-top: 10px; }
.compress-detail-item { display: flex; align-items: center; gap: 8px; padding: 3px 0; font-size: 12px; color: var(--cm-text-secondary); }
.detail-action { background: rgba(16,185,129,0.1); color: var(--cm-primary, #10b981); padding: 1px 6px; border-radius: 4px; font-size: 11px; }
.detail-target { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.auto-compress-setting { margin-top: 14px; border-top: 1px solid var(--cm-border, #e5e5e5); padding-top: 10px; }

.evolution-actions { display: flex; flex-wrap: wrap; gap: 8px; margin: 12px 0; }
.evolution-insights { margin-top: 12px; }

.discover-result { margin-top: 12px; }
.discover-list { max-height: 200px; overflow-y: auto; }
.discover-item { display: flex; align-items: center; gap: 6px; padding: 4px 0; font-size: 12px; color: var(--cm-text-secondary); }
.discover-source, .discover-target { max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 500; }
.discover-type { color: var(--cm-primary, #10b981); font-weight: 600; font-size: 11px; }
.discover-arrow { color: var(--cm-text-muted); }

.infer-result { margin-top: 12px; }
.chain-list { max-height: 200px; overflow-y: auto; }
.chain-item { padding: 8px 0; border-bottom: 1px solid var(--cm-border, #e5e5e5); }
.chain-nodes { font-size: 12px; color: var(--cm-text-secondary); }
.chain-node { font-weight: 500; }
.chain-arrow { color: var(--cm-primary, #10b981); font-weight: 600; }
.chain-conclusion { font-size: 13px; color: var(--cm-primary, #10b981); font-weight: 600; margin-top: 4px; }



@media (max-width: 768px) {
  .advanced-grid {
    grid-template-columns: 1fr;
    padding: 16px;
  }
  .page-hero { padding: 16px; }
  .compress-levels { flex-direction: column; }
  .router-test { flex-direction: column; }
  .conflict-values { flex-direction: column; }
}

@media (max-width: 480px) {
  .advanced-grid { padding: 12px; }
  .page-hero { padding: 12px; }
  .hero-content h1 { font-size: 20px; }
  .help-popup { width: 240px; right: -10px; }
}
</style>
