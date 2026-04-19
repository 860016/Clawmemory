<template>
  <div class="pro-page">
    <div class="page-header">
      <h1>🚀 {{ $t('pro.title') }}</h1>
      <span class="pro-badge" v-if="isPro">PRO</span>
      <el-button v-else type="primary" size="small" @click="$router.push('/settings')">{{ $t('pro.upgrade') }}</el-button>
    </div>

    <div class="pro-grid" v-if="isPro">
      <!-- Memory Decay -->
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">📉</span>
          <span class="card-title">{{ $t('pro.decay') }}</span>
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
            <el-button size="small" @click="loadDecayStats" :loading="loading.decay">{{ $t('pro.refreshStats') }}</el-button>
            <el-button size="small" type="primary" @click="applyDecay" :loading="loading.applyDecay">{{ $t('pro.applyDecay') }}</el-button>
          </div>
          <!-- Prune suggestions -->
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
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">⚠️</span>
          <span class="card-title">{{ $t('pro.conflicts') }}</span>
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
            <el-button size="small" @click="scanConflicts" :loading="loading.conflicts">{{ $t('pro.scanConflicts') }}</el-button>
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
                <el-button v-if="c.severity === 'low'" size="small" type="primary" text @click="resolveConflict(i, 'merge')">
                  {{ $t('pro.merge') }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Token Stats & Router -->
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">🧠</span>
          <span class="card-title">{{ $t('pro.tokenRouter') }}</span>
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
            <el-button size="small" @click="loadTokenStats" :loading="loading.tokenStats">{{ $t('pro.refreshTokens') }}</el-button>
          </div>
          <div class="router-test">
            <el-input v-model="testMessage" :placeholder="$t('pro.testMessage')" size="small" />
            <el-button size="small" type="primary" @click="testRoute" :loading="loading.route">{{ $t('pro.testRoute') }}</el-button>
          </div>
          <div v-if="routeResult" class="route-result">
            <div class="route-model">{{ $t('pro.selectedModel') }}: <strong>{{ routeResult.selected_model }}</strong></div>
            <div class="route-complexity">{{ $t('pro.complexity') }}: {{ routeResult.complexity }}</div>
          </div>
        </div>
      </div>

      <!-- AI Extract -->
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">🔍</span>
          <span class="card-title">{{ $t('pro.aiExtract') }}</span>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.aiExtractDesc') }}</p>
          <div class="card-actions">
            <el-button size="small" type="primary" @click="runAiExtract" :loading="loading.extract">{{ $t('pro.runExtract') }}</el-button>
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
      <div class="pro-card">
        <div class="card-header">
          <span class="card-icon">🕸️</span>
          <span class="card-title">{{ $t('pro.autoGraph') }}</span>
        </div>
        <div class="card-body">
          <p class="card-desc">{{ $t('pro.autoGraphDesc') }}</p>
          <div class="card-actions">
            <el-button size="small" @click="runAutoGraph(false)" :loading="loading.graph">{{ $t('pro.generateGraph') }}</el-button>
            <el-button size="small" type="danger" @click="runAutoGraph(true)" :loading="loading.graph">{{ $t('pro.regenerateGraph') }}</el-button>
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
    </div>

    <!-- Not Pro -->
    <div v-else class="pro-upsell">
      <div class="upsell-icon">🚀</div>
      <h2>{{ $t('pro.unlockPro') }}</h2>
      <p>{{ $t('pro.upsellDesc') }}</p>
      <el-button type="primary" @click="$router.push('/settings')">{{ $t('pro.viewPricing') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import axios from '../api/client'
import proApi from '../api/pro'

const { t } = useI18n()

const isPro = ref(false)
const loading = ref<Record<string, boolean>>({})

// Decay
const decayStats = ref<any>(null)
const pruneSuggestions = ref<any[]>([])

// Conflicts
const conflicts = ref<any[]>([])
const conflictSummary = ref<any>(null)

// Token
const tokenStatData = ref<any>(null)
const testMessage = ref('')
const routeResult = ref<any>(null)

// Extract
const extractResult = ref<any>(null)

// Graph
const graphResult = ref<any>(null)

// Backup
const backupSchedule = ref({ enabled: false, interval_hours: 24 })

onMounted(async () => {
  try {
    const { data } = await axios.get('/license/info')
    isPro.value = data.tier !== 'oss' && data.active
  } catch {}
  if (isPro.value) {
    loadDecayStats()
    loadTokenStats()
    loadBackupSchedule()
  }
})

async function loadDecayStats() {
  loading.value.decay = true
  try {
    const { data } = await proApi.getDecayStats()
    decayStats.value = data
    pruneSuggestions.value = (data.memories || []).filter((m: any) => m.should_prune)
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.decay = false }
}

async function applyDecay() {
  loading.value.applyDecay = true
  try {
    const { data } = await proApi.applyDecay()
    ElMessage.success(t('pro.decayApplied', { updated: data.updated, deleted: data.auto_deleted }))
    loadDecayStats()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.applyDecay = false }
}

async function scanConflicts() {
  loading.value.conflicts = true
  try {
    const { data } = await proApi.scanConflicts()
    conflicts.value = data.conflicts
    conflictSummary.value = data.summary
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.conflicts = false }
}

async function resolveConflict(index: number, strategy: string) {
  try {
    await proApi.resolveConflict(index, strategy)
    ElMessage.success(t('pro.conflictResolved'))
    scanConflicts()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  }
}

async function loadTokenStats() {
  loading.value.tokenStats = true
  try {
    const { data } = await proApi.getTokenStats()
    tokenStatData.value = data
  } catch (e: any) {
    if (e.response?.status !== 403) ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.tokenStats = false }
}

async function testRoute() {
  if (!testMessage.value) return
  loading.value.route = true
  try {
    const { data } = await proApi.routeToken(testMessage.value)
    routeResult.value = data
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.route = false }
}

async function runAiExtract() {
  loading.value.extract = true
  try {
    const { data } = await proApi.aiExtract()
    extractResult.value = data
    ElMessage.success(t('pro.extractDone', { entities: data.entities_extracted, relations: data.relations_extracted }))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally { loading.value.extract = false }
}

async function runAutoGraph(overwrite: boolean) {
  loading.value.graph = true
  try {
    const { data } = await proApi.autoGraph(overwrite)
    graphResult.value = data
    ElMessage.success(t('pro.graphDone', { entities: data.entities_created, relations: data.relations_created }))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
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
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  }
}
</script>

<style scoped>
.pro-page { padding: 28px; }
.page-header { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.pro-badge { background: rgba(16,185,129,0.15); color: #10B981; padding: 2px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.pro-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(400px, 1fr)); gap: 16px; }
.pro-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; overflow: hidden; transition: border-color 0.2s ease, box-shadow 0.2s ease; }
.pro-card:hover { border-color: rgba(16,185,129,0.25); box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.card-header { display: flex; align-items: center; gap: 8px; padding: 14px 18px; border-bottom: 1px solid var(--cm-border); }
.card-icon { font-size: 18px; }
.card-title { font-size: 15px; font-weight: 600; color: var(--cm-text); }
.card-body { padding: 16px 18px; }
.card-desc { color: var(--cm-text-muted); font-size: 13px; margin: 0 0 12px; }
.card-actions { display: flex; gap: 8px; margin-top: 12px; }
.stats-row { display: flex; gap: 20px; }
.stat-item { display: flex; flex-direction: column; align-items: center; }
.stat-value { font-size: 22px; font-weight: 700; color: var(--cm-text); }
.stat-value.warn { color: #d29922; }
.stat-value.danger { color: #f85149; }
.stat-label { font-size: 11px; color: var(--cm-text-muted); margin-top: 2px; }
.prune-section { margin-top: 14px; }
.sub-title { font-size: 13px; color: var(--cm-text-muted); margin-bottom: 8px; }
.prune-list { max-height: 150px; overflow-y: auto; }
.prune-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-size: 12px; color: var(--cm-text-secondary); }
.prune-key { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prune-imp { color: #d29922; font-weight: 600; }
.conflict-list { margin-top: 12px; }
.conflict-item { padding: 8px 0; border-bottom: 1px solid var(--cm-border); }
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
.backup-schedule .setting-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; color: var(--cm-text); font-size: 14px; }
.pro-upsell { text-align: center; padding: 80px 20px; }
.upsell-icon { font-size: 48px; margin-bottom: 16px; }
.pro-upsell h2 { color: var(--cm-text); margin: 0 0 12px; }
.pro-upsell p { color: var(--cm-text-muted); margin-bottom: 20px; }
@media (max-width: 768px) { .pro-grid { grid-template-columns: 1fr; } }
</style>
