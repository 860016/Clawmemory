<template>
  <div class="advanced-page">
    <div class="page-hero">
      <div class="hero-content">
        <h1>🚀 {{ $t('advanced.title') }}</h1>
        <span class="advanced-badge">Advanced</span>
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
          <div class="stats-row" v-if="decayStats">
            <div class="stat-item">
              <span class="stat-value">{{ decayStats.total }}</span>
              <span class="stat-label">{{ $t('advanced.totalMemories') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value warn">{{ decayStats.archived }}</span>
              <span class="stat-label">{{ $t('advanced.pruneCandidates') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ decayStats.avg_importance }}</span>
              <span class="stat-label">{{ $t('advanced.avgImportance') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="loadDecayStats" :disabled="loading.decay" :loading="loading.decay">
              {{ $t('advanced.refreshStats') }}
            </el-button>
            <el-button type="primary" @click="applyDecay" :disabled="loading.applyDecay" :loading="loading.applyDecay">
              {{ $t('advanced.applyDecay') }}
            </el-button>
          </div>
          <div v-if="pruneSuggestions.length" class="prune-section">
            <div class="sub-title">{{ $t('advanced.pruneSuggestions') }} ({{ pruneSuggestions.length }})</div>
            <div class="prune-list">
              <div v-for="s in pruneSuggestions.slice(0, 10)" :key="s.memory_id" class="prune-item">
                <span class="prune-key">{{ s.key }}</span>
                <el-tag size="small" type="info">{{ s.suggestion }}</el-tag>
                <span class="prune-imp">{{ s.importance?.toFixed(2) }}</span>
                <span class="prune-days">{{ s.days_old }}d</span>
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
              <div class="conflict-key">{{ c.key }}</div>
              <div class="conflict-values" v-if="c.memories && c.memories.length >= 2">
                <span class="val-a">{{ c.memories[0].value?.substring(0, 60) }}</span>
                <span class="vs">vs</span>
                <span class="val-b">{{ c.memories[1].value?.substring(0, 60) }}</span>
              </div>
              <div class="conflict-meta">
                <el-tag size="small" :type="c.severity === 'exact_duplicate' ? 'danger' : c.severity === 'similar_content' ? 'warning' : 'info'">
                  {{ c.severity }}
                </el-tag>
                <el-button v-if="c.severity !== 'exact_duplicate'" size="small" type="primary" @click="resolveConflict(i, 'merge')">
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
            <div class="stat-item">
              <span class="stat-value">{{ extractResult.entities_extracted }}</span>
              <span class="stat-label">{{ $t('advanced.entitiesExtracted') }}</span>
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
            <div class="stat-item">
              <span class="stat-value">{{ graphResult.total_pairs }}</span>
              <span class="stat-label">{{ $t('advanced.newEntities') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ graphResult.created }}</span>
              <span class="stat-label">{{ $t('advanced.newRelations') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Auto Backup -->
      <div class="advanced-card">
        <div class="card-header">
          <span class="card-icon">💾</span>
          <span class="card-title">{{ $t('advanced.autoBackup') }}</span>
          <div class="card-help" @mouseenter="showHelp.backup = true" @mouseleave="showHelp.backup = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.backup">
              <div class="help-title">{{ $t('advanced.autoBackup') }}</div>
              <div class="help-text">{{ $t('advanced.autoBackupHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="backup-schedule">
            <div class="setting-item">
              <span>{{ $t('advanced.autoBackupEnabled') }}</span>
              <el-switch v-model="backupSchedule.enabled" @change="saveBackupSchedule" />
            </div>
            <div class="setting-item" v-if="backupSchedule.enabled">
              <span>{{ $t('advanced.backupInterval') }}</span>
              <el-select v-model="backupSchedule.interval_hours" size="small" @change="saveBackupSchedule" style="width: 140px">
                <el-option :label="$t('advanced.every6h')" :value="6" />
                <el-option :label="$t('advanced.every12h')" :value="12" />
                <el-option :label="$t('advanced.every24h')" :value="24" />
                <el-option :label="$t('advanced.every7d')" :value="168" />
              </el-select>
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
          <div class="prefetch-section">
            <div class="setting-item">
              <span>{{ $t('advanced.memoryPrefetch') }}</span>
            </div>
            <div class="router-test">
              <el-input v-model="prefetchContext" :placeholder="$t('advanced.prefetchPlaceholder')" size="small" />
              <el-button type="primary" @click="runPrefetch" :disabled="loading.prefetch" :loading="loading.prefetch">
                {{ $t('advanced.prefetch') }}
              </el-button>
            </div>
            <div v-if="prefetchResult" class="prefetch-result">
              <span class="prefetch-count">{{ $t('advanced.prefetchMatched', { count: prefetchResult.matched_count }) }}</span>
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
import { toolboxApi } from '../api/go-toolbox'
import { translateError } from '../i18n'

const { t } = useI18n()
const route = useRoute()

const loading = ref<Record<string, boolean>>({})
const activeSection = ref((route.query.section as string) || '')
const showHelp = reactive<Record<string, boolean>>({
  decay: false, conflicts: false, router: false, tokenStats: false,
  extract: false, graph: false, backup: false, compress: false, evolution: false,
})

const decayStats = ref<any>(null)
const pruneSuggestions = ref<any[]>([])

const conflicts = ref<any[]>([])
const conflictSummary = ref<any>(null)

const tokenStatData = ref<any>(null)
const testMessage = ref('')
const routeResult = ref<any>(null)

const extractResult = ref<any>(null)

const graphResult = ref<any>(null)

const backupSchedule = ref({ enabled: false, interval_hours: 24 })

const compressLevel = ref<'light' | 'medium' | 'deep'>('light')
const compressPreviewData = ref<any>(null)
const compressConfig = ref<any>({ auto_enabled: false, threshold: 1000, level: 'light' })

const evolutionInsights = ref<any>(null)
const discoverResult = ref<any>(null)
const inferResult = ref<any>(null)
const prefetchContext = ref('')
const prefetchResult = ref<any>(null)

onMounted(async () => {
  loadDecayStats()
  loadTokenStats()
  loadBackupSchedule()
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
    const { data } = await toolboxApi.getDecayStats()
    decayStats.value = data
    const { data: pruneData } = await toolboxApi.getPruneSuggestions()
    pruneSuggestions.value = pruneData.suggestions || []
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.decay = false }
}

async function applyDecay() {
  loading.value.applyDecay = true
  try {
    const { data } = await toolboxApi.applyDecay()
    ElMessage.success(t('advanced.decayApplied', { updated: data.adjusted || data.processed || 0, deleted: data.trashed || 0 }))
    loadDecayStats()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.applyDecay = false }
}

async function scanConflicts() {
  loading.value.conflicts = true
  try {
    const { data } = await toolboxApi.scanConflicts()
    conflicts.value = data.conflicts
    conflictSummary.value = { total: data.total, auto_resolvable: data.auto_resolvable || 0, needs_review: data.needs_review || data.total }
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.conflicts = false }
}

async function resolveConflict(index: number, strategy: string) {
  try {
    await toolboxApi.resolveConflict(index, strategy)
    ElMessage.success(t('advanced.conflictResolved'))
    scanConflicts()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function loadTokenStats() {
  loading.value.tokenStats = true
  try {
    const { data } = await toolboxApi.getTokenStats()
    tokenStatData.value = { total_estimated_tokens: data.total_tokens, avg_tokens_per_memory: data.memory_count > 0 ? Math.round(data.total_tokens / data.memory_count) : 0 }
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.tokenStats = false }
}

async function testRoute() {
  if (!testMessage.value) return
  loading.value.route = true
  try {
    const { data } = await toolboxApi.routeModel(testMessage.value)
    routeResult.value = { selected_model: data.recommended_layer || data.strategy, complexity: data.complexity || 'simple' }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.route = false }
}

async function runAiExtract() {
  loading.value.extract = true
  try {
    const { data } = await toolboxApi.aiExtract()
    extractResult.value = { entities_extracted: data.extracted, relations_extracted: 0 }
    ElMessage.success(t('advanced.extractDone', { entities: data.extracted, relations: 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.extract = false }
}

async function runAutoGraph(overwrite: boolean) {
  loading.value.graph = true
  try {
    const { data } = await toolboxApi.autoGraph(overwrite)
    graphResult.value = data
    ElMessage.success(t('advanced.graphDone', { entities: 0, relations: data.created }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.graph = false }
}

async function loadBackupSchedule() {
  try {
    const { data } = await toolboxApi.getBackupSchedule()
    backupSchedule.value = data
  } catch { backupSchedule.value = { enabled: false, interval_hours: 24 } }
}

async function saveBackupSchedule() {
  try {
    await toolboxApi.setBackupSchedule(backupSchedule.value)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function previewCompress() {
  loading.value.compressPreview = true
  try {
    const { data } = await toolboxApi.compressPreview(compressLevel.value)
    compressPreviewData.value = data
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.compressPreview = false }
}

async function applyCompress() {
  loading.value.compressApply = true
  try {
    const { data } = await toolboxApi.compressApply(compressLevel.value)
    ElMessage.success(t('advanced.compressDone', { count: data.archived, ratio: data.total > 0 ? Math.round(data.archived / data.total * 100) : 0 }))
    compressPreviewData.value = null
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.compressApply = false }
}

async function saveCompressConfig() {
  try {
    await toolboxApi.setCompressConfig(compressConfig.value)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
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
    const { data } = await toolboxApi.getEvolutionInsights()
    evolutionInsights.value = { total_memories: data.total, relations_count: 0, discovered_relations: 0, inferred_chains: 0 }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.insights = false }
}

async function runDiscoverRelations() {
  loading.value.discover = true
  try {
    const { data } = await toolboxApi.discoverRelations()
    discoverResult.value = data
    ElMessage.success(t('advanced.discoveredCount', { count: data.discoveries?.length || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.discover = false }
}

async function runInferChains() {
  loading.value.infer = true
  try {
    const { data } = await toolboxApi.inferChains()
    inferResult.value = data
    ElMessage.success(t('advanced.inferredCount', { count: data.inferences?.length || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.infer = false }
}

async function runImportanceAdjust() {
  loading.value.importance = true
  try {
    const { data } = await toolboxApi.getImportanceAdjustments()
    ElMessage.success(t('advanced.adjustedCount', { count: data.total || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.importance = false }
}

async function runPrefetch() {
  if (!prefetchContext.value) return
  loading.value.prefetch = true
  try {
    const { data } = await toolboxApi.prefetchMemories(prefetchContext.value)
    prefetchResult.value = { matched_count: data.total || 0 }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.prefetch = false }
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
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 8px 24px rgba(99,102,241,0.12);
  transform: translateY(-2px);
}

.advanced-card.section-highlight {
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 0 0 2px rgba(99,102,241,0.2);
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
  background: var(--cm-primary, #6366f1);
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

.extract-result, .graph-result { display: flex; gap: 20px; margin-top: 12px; }

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
.level-option:hover { border-color: rgba(99,102,241,0.4); }
.level-option.active { border-color: var(--cm-primary, #6366f1); background: rgba(99,102,241,0.08); }
.level-name { font-size: 14px; font-weight: 600; color: var(--cm-text); }
.level-rate { font-size: 18px; font-weight: 700; color: var(--cm-primary, #6366f1); margin: 4px 0; }
.level-desc { font-size: 11px; color: var(--cm-text-muted); }

.compress-result { margin-top: 12px; }
.compress-details { margin-top: 10px; }
.compress-detail-item { display: flex; align-items: center; gap: 8px; padding: 3px 0; font-size: 12px; color: var(--cm-text-secondary); }
.detail-action { background: rgba(99,102,241,0.1); color: var(--cm-primary, #6366f1); padding: 1px 6px; border-radius: 4px; font-size: 11px; }
.detail-target { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.auto-compress-setting { margin-top: 14px; border-top: 1px solid var(--cm-border, #e5e5e5); padding-top: 10px; }

.evolution-actions { display: flex; flex-wrap: wrap; gap: 8px; margin: 12px 0; }
.evolution-insights { margin-top: 12px; }

.discover-result { margin-top: 12px; }
.discover-list { max-height: 200px; overflow-y: auto; }
.discover-item { display: flex; align-items: center; gap: 6px; padding: 4px 0; font-size: 12px; color: var(--cm-text-secondary); }
.discover-source, .discover-target { max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 500; }
.discover-type { color: var(--cm-primary, #6366f1); font-weight: 600; font-size: 11px; }
.discover-arrow { color: var(--cm-text-muted); }

.infer-result { margin-top: 12px; }
.chain-list { max-height: 200px; overflow-y: auto; }
.chain-item { padding: 8px 0; border-bottom: 1px solid var(--cm-border, #e5e5e5); }
.chain-nodes { font-size: 12px; color: var(--cm-text-secondary); }
.chain-node { font-weight: 500; }
.chain-arrow { color: var(--cm-primary, #6366f1); font-weight: 600; }
.chain-conclusion { font-size: 13px; color: var(--cm-primary, #6366f1); font-weight: 600; margin-top: 4px; }

.prefetch-section { margin-top: 14px; border-top: 1px solid var(--cm-border, #e5e5e5); padding-top: 10px; }
.prefetch-result { margin-top: 8px; }
.prefetch-count { font-size: 13px; color: var(--cm-primary, #6366f1); font-weight: 500; }

@media (max-width: 768px) {
  .advanced-grid {
    grid-template-columns: 1fr;
    padding: 16px;
  }
  .page-hero { padding: 16px; }
  .compress-levels { flex-direction: column; }
  .router-test { flex-direction: column; }
  .conflict-values { flex-direction: column; }
  .extract-result, .graph-result { flex-direction: column; gap: 12px; }
}

@media (max-width: 480px) {
  .advanced-grid { padding: 12px; }
  .page-hero { padding: 12px; }
  .hero-content h1 { font-size: 20px; }
  .help-popup { width: 240px; right: -10px; }
}
</style>
