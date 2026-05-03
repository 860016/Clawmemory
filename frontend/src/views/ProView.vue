<template>
  <div class="pro-page">
    <div class="page-hero">
      <div class="hero-content">
        <h1>🚀 {{ $t('pro.title') }}</h1>
        <span class="pro-badge" v-if="isPro">PRO</span>
      </div>
    </div>

    <div class="pro-grid" v-if="isPro">
      <div class="mode-banner">
        <span>⚡ {{ $t('pro.localProMode') }}</span>
      </div>

      <!-- Memory Decay -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'decay' }" id="pro-decay">
        <div class="card-header">
          <span class="card-icon">📉</span>
          <span class="card-title">{{ $t('pro.decay') }}</span>
          <div class="card-help" @mouseenter="showHelp.decay = true" @mouseleave="showHelp.decay = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.decay">
              <div class="help-title">{{ $t('pro.decay') }}</div>
              <div class="help-text">{{ $t('pro.decayHelp') }}</div>
              <div class="help-stages">
                <div class="help-stage"><span class="stage-dot green"></span>{{ $t('pro.decayStage1') }}</div>
                <div class="help-stage"><span class="stage-dot yellow"></span>{{ $t('pro.decayStage2') }}</div>
                <div class="help-stage"><span class="stage-dot orange"></span>{{ $t('pro.decayStage3') }}</div>
                <div class="help-stage"><span class="stage-dot red"></span>{{ $t('pro.decayStage4') }}</div>
              </div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="stats-row" v-if="decayStats">
            <div class="stat-item">
              <span class="stat-value">{{ decayStats.total }}</span>
              <span class="stat-label">{{ $t('pro.totalMemories') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value warn">{{ decayStats.prune_candidates }}</span>
              <span class="stat-label">{{ $t('pro.pruneCandidates') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ decayStats.avg_importance }}</span>
              <span class="stat-label">{{ $t('pro.avgImportance') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="loadDecayStats" :disabled="loading.decay" :loading="loading.decay">
              {{ $t('pro.refreshStats') }}
            </el-button>
            <el-button type="primary" @click="applyDecay" :disabled="loading.applyDecay" :loading="loading.applyDecay">
              {{ $t('pro.applyDecay') }}
            </el-button>
          </div>
          <div v-if="pruneSuggestions.length" class="prune-section">
            <div class="sub-title">{{ $t('pro.pruneSuggestions') }} ({{ pruneSuggestions.length }})</div>
            <div class="prune-list">
              <div v-for="s in pruneSuggestions.slice(0, 10)" :key="s.id" class="prune-item">
                <span class="prune-key">{{ s.key }}</span>
                <el-tag size="small" type="info">{{ s.layer }}</el-tag>
                <span class="prune-imp">{{ s.decayed_importance }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Conflict Scan -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'conflicts' }" id="pro-conflicts">
        <div class="card-header">
          <span class="card-icon">⚠️</span>
          <span class="card-title">{{ $t('pro.conflicts') }}</span>
          <div class="card-help" @mouseenter="showHelp.conflicts = true" @mouseleave="showHelp.conflicts = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.conflicts">
              <div class="help-title">{{ $t('pro.conflicts') }}</div>
              <div class="help-text">{{ $t('pro.conflictsHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="stats-row" v-if="conflictSummary">
            <div class="stat-item">
              <span class="stat-value warn">{{ conflictSummary.total }}</span>
              <span class="stat-label">{{ $t('pro.conflictCount') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ conflictSummary.auto_resolvable }}</span>
              <span class="stat-label">{{ $t('pro.autoResolvable') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value danger">{{ conflictSummary.needs_review }}</span>
              <span class="stat-label">{{ $t('pro.needsReview') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="scanConflicts" :disabled="loading.conflicts" :loading="loading.conflicts">
              {{ $t('pro.scanConflicts') }}
            </el-button>
          </div>
          <div v-if="conflicts.length" class="conflict-list">
            <div v-for="(c, i) in conflicts.slice(0, 10)" :key="i" class="conflict-item">
              <div class="conflict-key">{{ c.key }}</div>
              <div class="conflict-values">
                <span class="val-a">{{ c.value_a?.substring(0, 60) }}</span>
                <span class="vs">vs</span>
                <span class="val-b">{{ c.value_b?.substring(0, 60) }}</span>
              </div>
              <div class="conflict-meta">
                <el-tag size="small" :type="c.severity === 'high' ? 'danger' : c.severity === 'medium' ? 'warning' : 'info'">
                  {{ c.severity }}
                </el-tag>
                <el-button v-if="c.severity === 'low'" size="small" type="primary" @click="resolveConflict(i, 'merge')">
                  {{ $t('pro.merge') }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Smart Router -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'router' }" id="pro-router">
        <div class="card-header">
          <span class="card-icon">🧠</span>
          <span class="card-title">{{ $t('pro.smartRouter') }}</span>
          <div class="card-help" @mouseenter="showHelp.router = true" @mouseleave="showHelp.router = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.router">
              <div class="help-title">{{ $t('pro.smartRouter') }}</div>
              <div class="help-text">{{ $t('pro.smartRouterHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.smartRouterDesc') }}</p>
          <div class="router-test">
            <el-input v-model="testMessage" :placeholder="$t('pro.testMessage')" size="small" />
            <el-button type="primary" @click="testRoute" :disabled="loading.route" :loading="loading.route">
              {{ $t('pro.testRoute') }}
            </el-button>
          </div>
          <div v-if="routeResult" class="route-result">
            <div class="route-model">{{ $t('pro.selectedModel') }}: <strong>{{ routeResult.selected_model }}</strong></div>
            <div class="route-complexity">{{ $t('pro.complexity') }}: {{ routeResult.complexity }}</div>
          </div>
        </div>
      </div>

      <!-- Token Stats -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'tokenStats' }" id="pro-tokenStats">
        <div class="card-header">
          <span class="card-icon">📊</span>
          <span class="card-title">{{ $t('pro.tokenStats') }}</span>
          <div class="card-help" @mouseenter="showHelp.tokenStats = true" @mouseleave="showHelp.tokenStats = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.tokenStats">
              <div class="help-title">{{ $t('pro.tokenStats') }}</div>
              <div class="help-text">{{ $t('pro.tokenStatsHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="stats-row" v-if="tokenStatData">
            <div class="stat-item">
              <span class="stat-value">{{ tokenStatData.total_estimated_tokens }}</span>
              <span class="stat-label">{{ $t('pro.totalTokens') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ tokenStatData.avg_tokens_per_memory }}</span>
              <span class="stat-label">{{ $t('pro.avgTokens') }}</span>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="loadTokenStats" :disabled="loading.tokenStats" :loading="loading.tokenStats">
              {{ $t('pro.refreshTokens') }}
            </el-button>
          </div>
        </div>
      </div>

      <!-- AI Extract -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'extract' }" id="pro-extract">
        <div class="card-header">
          <span class="card-icon">🔍</span>
          <span class="card-title">{{ $t('pro.aiExtract') }}</span>
          <div class="card-help" @mouseenter="showHelp.extract = true" @mouseleave="showHelp.extract = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.extract">
              <div class="help-title">{{ $t('pro.aiExtract') }}</div>
              <div class="help-text">{{ $t('pro.aiExtractHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.aiExtractDesc') }}</p>
          <div class="card-actions">
            <el-button type="primary" @click="runAiExtract" :disabled="loading.extract" :loading="loading.extract">
              {{ $t('pro.runExtract') }}
            </el-button>
          </div>
          <div v-if="extractResult" class="extract-result">
            <div class="stat-item">
              <span class="stat-value">{{ extractResult.entities_extracted }}</span>
              <span class="stat-label">{{ $t('pro.entitiesExtracted') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ extractResult.relations_extracted }}</span>
              <span class="stat-label">{{ $t('pro.relationsExtracted') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Auto Graph -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'graph' }" id="pro-graph">
        <div class="card-header">
          <span class="card-icon">🕸️</span>
          <span class="card-title">{{ $t('pro.autoGraph') }}</span>
          <div class="card-help" @mouseenter="showHelp.graph = true" @mouseleave="showHelp.graph = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.graph">
              <div class="help-title">{{ $t('pro.autoGraph') }}</div>
              <div class="help-text">{{ $t('pro.autoGraphHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.autoGraphDesc') }}</p>
          <div class="card-actions">
            <el-button @click="runAutoGraph(false)" :disabled="loading.graph" :loading="loading.graph">
              {{ $t('pro.generateGraph') }}
            </el-button>
            <el-button type="danger" @click="runAutoGraph(true)" :disabled="loading.graph" :loading="loading.graph">
              {{ $t('pro.regenerateGraph') }}
            </el-button>
          </div>
          <div v-if="graphResult" class="graph-result">
            <div class="stat-item">
              <span class="stat-value">{{ graphResult.entities_created }}</span>
              <span class="stat-label">{{ $t('pro.newEntities') }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ graphResult.relations_created }}</span>
              <span class="stat-label">{{ $t('pro.newRelations') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Auto Backup -->
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">💾</span>
          <span class="card-title">{{ $t('pro.autoBackup') }}</span>
          <div class="card-help" @mouseenter="showHelp.backup = true" @mouseleave="showHelp.backup = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.backup">
              <div class="help-title">{{ $t('pro.autoBackup') }}</div>
              <div class="help-text">{{ $t('pro.autoBackupHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <div class="backup-schedule">
            <div class="setting-item">
              <span>{{ $t('pro.autoBackupEnabled') }}</span>
              <el-switch v-model="backupSchedule.enabled" @change="saveBackupSchedule" />
            </div>
            <div class="setting-item" v-if="backupSchedule.enabled">
              <span>{{ $t('pro.backupInterval') }}</span>
              <el-select v-model="backupSchedule.interval_hours" size="small" @change="saveBackupSchedule" style="width: 140px">
                <el-option :label="$t('pro.every6h')" :value="6" />
                <el-option :label="$t('pro.every12h')" :value="12" />
                <el-option :label="$t('pro.every24h')" :value="24" />
                <el-option :label="$t('pro.every7d')" :value="168" />
              </el-select>
            </div>
          </div>
        </div>
      </div>

      <!-- Memory Compression -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'compress' }" id="pro-compress">
        <div class="card-header">
          <span class="card-icon">🗜️</span>
          <span class="card-title">{{ $t('pro.compress') }}</span>
          <div class="card-help" @mouseenter="showHelp.compress = true" @mouseleave="showHelp.compress = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.compress">
              <div class="help-title">{{ $t('pro.compress') }}</div>
              <div class="help-text">{{ $t('pro.compressHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.compressDesc') }}</p>
          <div class="compress-levels">
            <div class="level-option" :class="{ active: compressLevel === 'light' }" @click="compressLevel = 'light'">
              <div class="level-name">{{ $t('pro.compressLight') }}</div>
              <div class="level-rate">{{ $t('pro.compressLightRate') }}</div>
              <div class="level-desc">{{ $t('pro.compressLightDesc') }}</div>
            </div>
            <div class="level-option" :class="{ active: compressLevel === 'medium' }" @click="compressLevel = 'medium'">
              <div class="level-name">{{ $t('pro.compressMedium') }}</div>
              <div class="level-rate">{{ $t('pro.compressMediumRate') }}</div>
              <div class="level-desc">{{ $t('pro.compressMediumDesc') }}</div>
            </div>
            <div class="level-option" :class="{ active: compressLevel === 'deep' }" @click="compressLevel = 'deep'">
              <div class="level-name">{{ $t('pro.compressDeep') }}</div>
              <div class="level-rate">{{ $t('pro.compressDeepRate') }}</div>
              <div class="level-desc">{{ $t('pro.compressDeepDesc') }}</div>
            </div>
          </div>
          <div class="card-actions">
            <el-button @click="previewCompress" :disabled="loading.compressPreview" :loading="loading.compressPreview">
              {{ $t('pro.previewCompress') }}
            </el-button>
            <el-button type="primary" @click="applyCompress" :disabled="loading.compressApply" :loading="loading.compressApply">
              {{ $t('pro.applyCompress') }}
            </el-button>
          </div>
          <div v-if="compressPreviewData" class="compress-result">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ compressPreviewData.original_count }}</span>
                <span class="stat-label">{{ $t('pro.originalMemories') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ compressPreviewData.compressed_count }}</span>
                <span class="stat-label">{{ $t('pro.compressedResult') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value" :class="{ success: compressPreviewData.ratio > 0.5 }">{{ Math.round(compressPreviewData.ratio * 100) }}%</span>
                <span class="stat-label">{{ $t('pro.compressRatio') }}</span>
              </div>
            </div>
            <div v-if="compressPreviewData.details" class="compress-details">
              <div v-for="(d, i) in compressPreviewData.details.slice(0, 5)" :key="i" class="compress-detail-item">
                <span class="detail-action">{{ d.action }}</span>
                <span class="detail-target">{{ d.target }}</span>
              </div>
            </div>
          </div>
          <div class="auto-compress-setting">
            <div class="setting-item">
              <span>{{ $t('pro.autoCompress') }}</span>
              <el-switch v-model="compressConfig.auto_enabled" @change="saveCompressConfig" />
            </div>
            <div class="setting-item" v-if="compressConfig.auto_enabled">
              <span>{{ $t('pro.compressThreshold') }}</span>
              <el-select v-model="compressConfig.threshold" size="small" @change="saveCompressConfig" style="width: 140px">
                <el-option :label="$t('pro.threshold500')" :value="500" />
                <el-option :label="$t('pro.threshold1000')" :value="1000" />
                <el-option :label="$t('pro.threshold2000')" :value="2000" />
              </el-select>
            </div>
          </div>
        </div>
      </div>

      <!-- Evolution Engine -->
      <div class="pro-card" :class="{ 'section-highlight': activeSection === 'evolution' }" id="pro-evolution">
        <div class="card-header">
          <span class="card-icon">🧬</span>
          <span class="card-title">{{ $t('pro.evolution') }}</span>
          <div class="card-help" @mouseenter="showHelp.evolution = true" @mouseleave="showHelp.evolution = false">
            <span class="help-icon">ⓘ</span>
            <div class="help-popup" v-if="showHelp.evolution">
              <div class="help-title">{{ $t('pro.evolution') }}</div>
              <div class="help-text">{{ $t('pro.evolutionHelp') }}</div>
            </div>
          </div>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.evolutionDesc') }}</p>
          <div class="evolution-actions">
            <el-button @click="loadEvolutionInsights" :disabled="loading.insights" :loading="loading.insights">
              {{ $t('pro.evolutionInsights') }}
            </el-button>
            <el-button @click="runDiscoverRelations" :disabled="loading.discover" :loading="loading.discover">
              {{ $t('pro.discoverRelations') }}
            </el-button>
            <el-button @click="runInferChains" :disabled="loading.infer" :loading="loading.infer">
              {{ $t('pro.inferChains') }}
            </el-button>
            <el-button @click="runImportanceAdjust" :disabled="loading.importance" :loading="loading.importance">
              {{ $t('pro.importanceAdjust') }}
            </el-button>
          </div>
          <div v-if="evolutionInsights" class="evolution-insights">
            <div class="stats-row">
              <div class="stat-item">
                <span class="stat-value">{{ evolutionInsights.total_memories }}</span>
                <span class="stat-label">{{ $t('pro.totalMemoriesLabel') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value success">{{ evolutionInsights.relations_count }}</span>
                <span class="stat-label">{{ $t('pro.relationsCount') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value warn">{{ evolutionInsights.discovered_relations }}</span>
                <span class="stat-label">{{ $t('pro.newDiscoveries') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ evolutionInsights.inferred_chains }}</span>
                <span class="stat-label">{{ $t('pro.inferChainsLabel') }}</span>
              </div>
            </div>
          </div>
          <div v-if="discoverResult" class="discover-result">
            <div class="sub-title">{{ $t('pro.discoveredRelations') }} ({{ discoverResult.relations?.length || 0 }})</div>
            <div class="discover-list">
              <div v-for="(r, i) in (discoverResult.relations || []).slice(0, 8)" :key="i" class="discover-item">
                <span class="discover-source">{{ r.source }}</span>
                <span class="discover-arrow">→</span>
                <span class="discover-type">{{ r.type }}</span>
                <span class="discover-arrow">→</span>
                <span class="discover-target">{{ r.target }}</span>
                <el-tag size="small" type="info">{{ Math.round(r.confidence * 100) }}%</el-tag>
              </div>
            </div>
          </div>
          <div v-if="inferResult" class="infer-result">
            <div class="sub-title">{{ $t('pro.inferChainsTitle') }} ({{ inferResult.chains?.length || 0 }})</div>
            <div class="chain-list">
              <div v-for="(c, i) in (inferResult.chains || []).slice(0, 5)" :key="i" class="chain-item">
                <div class="chain-nodes">
                  <span v-for="(n, j) in c.nodes" :key="j" class="chain-node">
                    {{ n }}<span v-if="j < c.nodes.length - 1" class="chain-arrow"> → </span>
                  </span>
                </div>
                <div class="chain-conclusion">
                  ∴ {{ c.conclusion }}
                </div>
              </div>
            </div>
          </div>
          <div class="prefetch-section">
            <div class="setting-item">
              <span>{{ $t('pro.memoryPrefetch') }}</span>
            </div>
            <div class="router-test">
              <el-input v-model="prefetchContext" :placeholder="$t('pro.prefetchPlaceholder')" size="small" />
              <el-button type="primary" @click="runPrefetch" :disabled="loading.prefetch" :loading="loading.prefetch">
                {{ $t('pro.prefetch') }}
              </el-button>
            </div>
            <div v-if="prefetchResult" class="prefetch-result">
              <span class="prefetch-count">{{ $t('pro.prefetchMatched', { count: prefetchResult.matched_count }) }}</span>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Not Pro - blurred overlay -->
    <div v-else class="pro-locked" @click="$router.push('/settings')">
      <div class="locked-content">
        <div class="locked-cards">
          <div class="locked-card" v-for="i in 6" :key="i">
            <div class="locked-card-header"></div>
            <div class="locked-card-body">
              <div class="locked-line w80"></div>
              <div class="locked-line w60"></div>
              <div class="locked-line w40"></div>
            </div>
          </div>
        </div>
        <div class="locked-overlay">
          <div class="locked-icon">🔒</div>
          <h2>{{ $t('pro.unlockPro') }}</h2>
          <p>{{ $t('pro.upsellDesc') }}</p>
          <el-button type="primary" size="large" @click.stop="$router.push('/settings')">{{ $t('pro.viewPricing') }}</el-button>
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
import axios from '../api/go-client'
import proApi from '../api/pro'
import { translateError } from '../i18n'

const { t } = useI18n()
const route = useRoute()

const isPro = ref(false)
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
  try {
    const { data } = await axios.get('/license/info')
    isPro.value = data.tier === 'pro' || data.active === true
  } catch {
    isPro.value = false
  }
  if (isPro.value) {
    loadDecayStats()
    loadTokenStats()
    loadBackupSchedule()
  }
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
  const el = document.getElementById(`pro-${section}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('section-highlight')
    setTimeout(() => el.classList.remove('section-highlight'), 2000)
  }
}

async function loadDecayStats() {
  loading.value.decay = true
  try {
    const { data } = await proApi.getDecayStats()
    decayStats.value = data
    pruneSuggestions.value = data.suggestions || []
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.decay = false }
}

async function applyDecay() {
  loading.value.applyDecay = true
  try {
    const { data } = await proApi.applyDecay()
    ElMessage.success(t('pro.decayApplied', { updated: data.updated, deleted: data.auto_deleted }))
    loadDecayStats()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.applyDecay = false }
}

async function scanConflicts() {
  loading.value.conflicts = true
  try {
    const { data } = await proApi.scanConflicts()
    conflicts.value = data.conflicts
    conflictSummary.value = { total: data.total, auto_resolvable: 0, needs_review: data.total }
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.conflicts = false }
}

async function resolveConflict(index: number, strategy: string) {
  try {
    await proApi.resolveConflict(index, strategy)
    ElMessage.success(t('pro.conflictResolved'))
    scanConflicts()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function loadTokenStats() {
  loading.value.tokenStats = true
  try {
    const { data } = await proApi.getTokenStats()
    tokenStatData.value = { total_estimated_tokens: data.total_tokens_used, avg_tokens_per_memory: data.total_memories > 0 ? Math.round(data.total_tokens_used / data.total_memories) : 0 }
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.tokenStats = false }
}

async function testRoute() {
  if (!testMessage.value) return
  loading.value.route = true
  try {
    const { data } = await proApi.routeToken(testMessage.value)
    routeResult.value = { selected_model: data.model || data.provider, complexity: data.estimated_tokens > 500 ? 'high' : data.estimated_tokens > 200 ? 'medium' : 'low' }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.route = false }
}

async function runAiExtract() {
  loading.value.extract = true
  try {
    const { data } = await proApi.aiExtract()
    extractResult.value = { entities_extracted: data.total, relations_extracted: 0 }
    ElMessage.success(t('pro.extractDone', { entities: data.total, relations: 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.extract = false }
}

async function runAutoGraph(overwrite: boolean) {
  loading.value.graph = true
  try {
    const { data } = await proApi.autoGraph(overwrite)
    graphResult.value = data
    ElMessage.success(t('pro.graphDone', { entities: data.entities_created, relations: data.relations_created }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.graph = false }
}

async function loadBackupSchedule() {
  try {
    const { data } = await proApi.getBackupSchedule()
    backupSchedule.value = data
  } catch {}
}

async function saveBackupSchedule() {
  try {
    await proApi.setBackupSchedule(backupSchedule.value)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function previewCompress() {
  loading.value.compressPreview = true
  try {
    const { data } = await proApi.compressPreview(compressLevel.value)
    compressPreviewData.value = data
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.compressPreview = false }
}

async function applyCompress() {
  loading.value.compressApply = true
  try {
    const { data } = await proApi.compressApply(compressLevel.value)
    ElMessage.success(t('pro.compressDone', { count: data.compressed, ratio: Math.round(data.ratio * 100) }))
    compressPreviewData.value = null
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.compressApply = false }
}

async function saveCompressConfig() {
  try {
    await proApi.setCompressConfig(compressConfig.value)
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function loadEvolutionInsights() {
  loading.value.insights = true
  try {
    const { data } = await proApi.getEvolutionInsights()
    evolutionInsights.value = { total_memories: data.total_memories, relations_count: data.total_relations, discovered_relations: 0, inferred_chains: 0 }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.insights = false }
}

async function runDiscoverRelations() {
  loading.value.discover = true
  try {
    const { data } = await proApi.discoverRelations()
    discoverResult.value = data
    ElMessage.success(t('pro.discoveredCount', { count: data.relations?.length || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.discover = false }
}

async function runInferChains() {
  loading.value.infer = true
  try {
    const { data } = await proApi.inferChains()
    inferResult.value = data
    ElMessage.success(t('pro.inferredCount', { count: data.chains?.length || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.infer = false }
}

async function runImportanceAdjust() {
  loading.value.importance = true
  try {
    const { data } = await proApi.getImportanceAdjustments()
    ElMessage.success(t('pro.adjustedCount', { count: data.total || 0 }))
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.importance = false }
}

async function runPrefetch() {
  if (!prefetchContext.value) return
  loading.value.prefetch = true
  try {
    const { data } = await proApi.prefetchMemories(prefetchContext.value)
    prefetchResult.value = { matched_count: data.total || 0 }
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally { loading.value.prefetch = false }
}
</script>

<style scoped>
.pro-page {
  height: 100%;
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

.pro-badge {
  background: rgba(16,185,129,0.15);
  color: #10B981;
  padding: 2px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}

.pro-grid {
  flex: 1;
  padding: 24px 28px;
  overflow-y: auto;
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

.pro-card {
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 14px;
  overflow: hidden;
  transition: all 0.25s ease;
}

.pro-card:hover {
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 8px 24px rgba(99,102,241,0.12);
  transform: translateY(-2px);
}

.pro-card.section-highlight {
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

/* ===== Pro Locked (Not Activated) ===== */
.pro-locked {
  flex: 1;
  cursor: pointer;
  overflow: hidden;
}

.locked-content {
  position: relative;
  height: 100%;
  padding: 28px;
}

.locked-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  filter: blur(6px);
  opacity: 0.4;
  pointer-events: none;
  user-select: none;
}

.locked-card {
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 14px;
  overflow: hidden;
}

.locked-card-header {
  height: 48px;
  background: var(--cm-bg-secondary, #f5f5f5);
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
}

.locked-card-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.locked-line {
  height: 12px;
  background: var(--cm-bg-tertiary, #f0f0f0);
  border-radius: 6px;
}

.locked-line.w80 { width: 80%; }
.locked-line.w60 { width: 60%; }
.locked-line.w40 { width: 40%; }

.locked-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  z-index: 10;
}

.locked-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.locked-overlay h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--cm-text, #1a1a1a);
  margin: 0 0 12px;
}

.locked-overlay p {
  color: var(--cm-text-secondary, #666);
  margin-bottom: 24px;
  max-width: 400px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .pro-grid {
    grid-template-columns: 1fr;
    padding: 16px;
  }
  .page-hero { padding: 16px; }
  .locked-cards { grid-template-columns: 1fr; }
  .compress-levels { flex-direction: column; }
  .router-test { flex-direction: column; }
  .conflict-values { flex-direction: column; }
  .extract-result, .graph-result { flex-direction: column; gap: 12px; }
}

@media (max-width: 480px) {
  .pro-grid { padding: 12px; }
  .page-hero { padding: 12px; }
  .hero-content h1 { font-size: 20px; }
  .help-popup { width: 240px; right: -10px; }
}
</style>
