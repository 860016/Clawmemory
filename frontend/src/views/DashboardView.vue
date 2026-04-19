<template>
  <div class="dashboard">
    <!-- Skills Tab -->
    <div v-if="activeTab === 'skills'" class="skills-embed">
      <div class="page-header">
        <h1>✨ {{ $t('skills.title') }}</h1>
        <el-button type="primary" @click="scanSkills" :loading="scanning">
          {{ scanning ? $t('skills.scanning') : $t('skills.scanSkills') }}
        </el-button>
      </div>
      <div v-if="!scanned && !scanning" class="empty-hint">{{ $t('skills.noSkills') }}</div>
      <div v-if="scanning" style="text-align:center;padding:40px;color:var(--cm-text-muted)">{{ $t('skills.scanning') }}</div>
      <div v-if="scanned && !scanning">
        <div v-if="globalSkills.length" class="skill-section">
          <h3 class="section-title">{{ $t('skills.globalSkills') }} ({{ globalSkills.length }})</h3>
          <div class="skill-grid">
            <div v-for="skill in globalSkills" :key="skill.skill_dir" class="skill-card" @click="showSkillDetail(skill)">
              <div class="skill-icon">🌐</div>
              <div class="skill-info">
                <div class="skill-name">{{ skill.name }}</div>
                <div class="skill-desc">{{ skill.description || '—' }}</div>
              </div>
              <div class="skill-meta">
                <span class="badge">{{ skill.version }}</span>
                <span class="scope global">{{ skill.scope }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="workspaceSkills.length" class="skill-section">
          <h3 class="section-title">{{ $t('skills.workspaceSkills') }} ({{ workspaceSkills.length }})</h3>
          <div class="skill-grid">
            <div v-for="skill in workspaceSkills" :key="skill.skill_dir" class="skill-card" @click="showSkillDetail(skill)">
              <div class="skill-icon">📁</div>
              <div class="skill-info">
                <div class="skill-name">{{ skill.name }}</div>
                <div class="skill-desc">{{ skill.description || '—' }}</div>
              </div>
              <div class="skill-meta">
                <span class="badge">{{ skill.version }}</span>
                <span class="scope workspace">{{ skill.scope }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!globalSkills.length && !workspaceSkills.length" class="empty-hint">{{ $t('skills.noSkills') }}</div>
      </div>
      <el-dialog v-model="skillDetailVisible" :title="skillDetail?.name" width="650px">
        <div v-if="skillDetail" style="font-size:14px;line-height:1.8">
          <div class="detail-row"><strong>{{ $t('skills.name') }}:</strong> {{ skillDetail.name }}</div>
          <div class="detail-row"><strong>{{ $t('skills.description') }}:</strong> {{ skillDetail.description || '—' }}</div>
          <div class="detail-row"><strong>{{ $t('skills.version') }}:</strong> {{ skillDetail.version }}</div>
          <div class="detail-row"><strong>{{ $t('skills.author') }}:</strong> {{ skillDetail.author }}</div>
          <div class="detail-row"><strong>{{ $t('skills.scope') }}:</strong> {{ skillDetail.scope }}</div>
          <div class="detail-row" v-if="skillDetail.tags?.length">
            <strong>Tags:</strong>
            <el-tag v-for="tag in skillDetail.tags" :key="tag" size="small" style="margin:2px">{{ tag }}</el-tag>
          </div>
          <div class="detail-row" v-if="skillDetail.files?.length">
            <strong>{{ $t('skills.files') }}:</strong>
            <div class="file-list">{{ skillDetail.files.join(', ') }}</div>
          </div>
          <div class="detail-body" v-if="skillDetail.body_full">
            <strong>Content:</strong>
            <pre>{{ skillDetail.body_full }}</pre>
          </div>
        </div>
      </el-dialog>
    </div>

    <!-- Stats Tab -->
    <div v-else-if="activeTab === 'stats'" class="stats-page">
      <div class="page-header">
        <h1>📈 {{ $t('dashboard.statsTitle') }}</h1>
        <div class="period-selector">
          <el-radio-group v-model="statsPeriod" size="small" @change="loadUsageStats">
            <el-radio-button :value="7">{{ $t('dashboard.last7d') }}</el-radio-button>
            <el-radio-button :value="30">{{ $t('dashboard.last30d') }}</el-radio-button>
            <el-radio-button :value="90">{{ $t('dashboard.last90d') }}</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <p class="page-desc">{{ $t('dashboard.statsSubtitle') }}</p>

      <!-- Summary Cards -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon-wrap mem-icon">🧠</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.memoryCount }}</div>
            <div class="stat-label">{{ $t('dashboard.memoryCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap entity-icon">🕸️</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.entityCount }}</div>
            <div class="stat-label">{{ $t('dashboard.entityCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap rel-icon">🔗</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.relationCount }}</div>
            <div class="stat-label">{{ $t('dashboard.relationCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap token-icon">🔤</div>
          <div class="stat-info">
            <div class="stat-value">{{ formatNumber(usageStats.totalEstimatedTokens) }}</div>
            <div class="stat-label">{{ $t('dashboard.totalEstimatedTokens') }}</div>
          </div>
        </div>
      </div>

      <div class="content-grid">
        <!-- Daily Trend Chart -->
        <div class="card chart-card">
          <div class="card-header">
            <h3>📊 {{ $t('dashboard.dailyTrend') }}</h3>
          </div>
          <div class="chart-container" ref="dailyTrendRef">
            <div v-if="!usageStats.dailyTrend?.length" class="empty-hint">{{ $t('common.noData') }}</div>
            <svg v-else :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="trend-chart" preserveAspectRatio="xMidYMid meet">
              <!-- Y axis -->
              <line x1="40" y1="10" x2="40" :y2="chartHeight - 25" stroke="var(--cm-border)" stroke-width="1" />
              <!-- X axis -->
              <line x1="40" :y1="chartHeight - 25" :x2="chartWidth - 10" :y2="chartHeight - 25" stroke="var(--cm-border)" stroke-width="1" />
              <!-- Grid lines -->
              <line v-for="i in 4" :key="'grid'+i" x1="40" :y1="10 + (i-1) * (chartHeight - 35) / 4" :x2="chartWidth - 10" :y2="10 + (i-1) * (chartHeight - 35) / 4" stroke="var(--cm-border)" stroke-width="0.5" stroke-dasharray="4,4" />
              <!-- Y axis labels -->
              <text v-for="i in 5" :key="'yl'+i" x="36" :y="10 + (i-1) * (chartHeight - 35) / 4 + 4" text-anchor="end" fill="var(--cm-text-muted)" font-size="10">{{ Math.round(maxDailyCount * (5-i) / 4) }}</text>
              <!-- Area fill -->
              <path :d="dailyTrendAreaPath" fill="rgba(16,185,129,0.1)" />
              <!-- Line -->
              <path :d="dailyTrendLinePath" fill="none" stroke="#10B981" stroke-width="2" stroke-linejoin="round" />
              <!-- Dots -->
              <circle v-for="(pt, idx) in dailyTrendPoints" :key="'dot'+idx" :cx="pt.x" :cy="pt.y" r="3" fill="#10B981" />
              <!-- X axis labels (show every few days) -->
              <text v-for="(pt, idx) in dailyTrendPoints" :key="'xl'+idx" v-show="idx % xLabelInterval === 0" :x="pt.x" :y="chartHeight - 8" text-anchor="middle" fill="var(--cm-text-muted)" font-size="9">{{ pt.label }}</text>
            </svg>
          </div>
        </div>

        <!-- Token Trend Chart -->
        <div class="card chart-card">
          <div class="card-header">
            <h3>🔤 {{ $t('dashboard.dailyTokenTrend') }}</h3>
          </div>
          <div class="chart-container" ref="tokenTrendRef">
            <div v-if="!usageStats.dailyTokenTrend?.length" class="empty-hint">{{ $t('common.noData') }}</div>
            <svg v-else :viewBox="`0 0 ${chartWidth} ${chartHeight}`" class="trend-chart" preserveAspectRatio="xMidYMid meet">
              <line x1="50" y1="10" x2="50" :y2="chartHeight - 25" stroke="var(--cm-border)" stroke-width="1" />
              <line x1="50" :y1="chartHeight - 25" :x2="chartWidth - 10" :y2="chartHeight - 25" stroke="var(--cm-border)" stroke-width="1" />
              <line v-for="i in 4" :key="'tgrid'+i" x1="50" :y1="10 + (i-1) * (chartHeight - 35) / 4" :x2="chartWidth - 10" :y2="10 + (i-1) * (chartHeight - 35) / 4" stroke="var(--cm-border)" stroke-width="0.5" stroke-dasharray="4,4" />
              <text v-for="i in 5" :key="'tyl'+i" x="46" :y="10 + (i-1) * (chartHeight - 35) / 4 + 4" text-anchor="end" fill="var(--cm-text-muted)" font-size="10">{{ formatNumber(Math.round(maxTokenCount * (5-i) / 4)) }}</text>
              <!-- Daily bars -->
              <rect v-for="(pt, idx) in tokenBarPoints" :key="'tbar'+idx" :x="pt.x" :y="pt.y" :width="pt.w" :height="pt.h" :fill="pt.color" rx="2" opacity="0.85" />
              <!-- Cumulative line -->
              <path :d="cumulativeLinePath" fill="none" stroke="#f59e0b" stroke-width="2" stroke-linejoin="round" />
              <text v-for="(pt, idx) in tokenBarPoints" :key="'txl'+idx" v-show="idx % xLabelInterval === 0" :x="pt.x + pt.w/2" :y="chartHeight - 8" text-anchor="middle" fill="var(--cm-text-muted)" font-size="9">{{ pt.label }}</text>
            </svg>
            <div class="chart-legend">
              <span class="legend-item"><span class="legend-color" style="background:#3b82f6"></span>{{ $t('dashboard.dailyTokens') }}</span>
              <span class="legend-item"><span class="legend-color" style="background:#f59e0b"></span>{{ $t('dashboard.cumulativeTokens') }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="content-grid">
        <!-- Source Distribution -->
        <div class="card">
          <div class="card-header">
            <h3>📂 {{ $t('dashboard.sourceDistribution') }}</h3>
          </div>
          <div class="distribution-bars" v-if="Object.keys(usageStats.sourceDistribution || {}).length">
            <div class="dist-bar" v-for="(count, source) in usageStats.sourceDistribution" :key="source">
              <div class="dist-label">{{ sourceLabel(source as string) }}</div>
              <div class="dist-track">
                <div class="dist-fill" :class="String(source)" :style="{ width: distWidth(count as number, usageStats.totalMemories) }"></div>
              </div>
              <div class="dist-count">{{ count }}</div>
            </div>
          </div>
          <div v-else class="empty-hint">{{ $t('common.noData') }}</div>
        </div>

        <!-- Importance Distribution -->
        <div class="card">
          <div class="card-header">
            <h3>⭐ {{ $t('dashboard.importanceDistribution') }}</h3>
          </div>
          <div class="distribution-bars" v-if="usageStats.importanceDistribution">
            <div class="dist-bar" v-for="(count, level) in usageStats.importanceDistribution" :key="level">
              <div class="dist-label">{{ importanceLabel(level as string) }}</div>
              <div class="dist-track">
                <div class="dist-fill importance" :class="String(level)" :style="{ width: distWidth(count as number, stats.memoryCount) }"></div>
              </div>
              <div class="dist-count">{{ count }}</div>
            </div>
          </div>
          <div v-else class="empty-hint">{{ $t('common.noData') }}</div>
        </div>
      </div>

      <div class="content-grid">
        <!-- Token by Layer -->
        <div class="card">
          <div class="card-header">
            <h3>🔤 {{ $t('dashboard.tokenByLayer') }}</h3>
          </div>
          <div class="distribution-bars" v-if="Object.keys(usageStats.tokenByLayer || {}).length">
            <div class="dist-bar" v-for="(tokens, layer) in usageStats.tokenByLayer" :key="layer">
              <div class="dist-label">{{ layerLabels[layer as string] || layer }}</div>
              <div class="dist-track">
                <div class="dist-fill layer" :class="String(layer)" :style="{ width: distWidth(tokens as number, usageStats.totalEstimatedTokens) }"></div>
              </div>
              <div class="dist-count">{{ formatNumber(tokens as number) }}</div>
            </div>
          </div>
          <div v-else class="empty-hint">{{ $t('common.noData') }}</div>
        </div>

        <!-- Entity Type Distribution -->
        <div class="card">
          <div class="card-header">
            <h3>🏷️ {{ $t('dashboard.entityTypeDistribution') }}</h3>
          </div>
          <div class="distribution-bars" v-if="Object.keys(usageStats.entityTypeDistribution || {}).length">
            <div class="dist-bar" v-for="(count, etype) in usageStats.entityTypeDistribution" :key="etype">
              <div class="dist-label">{{ etype }}</div>
              <div class="dist-track">
                <div class="dist-fill entity" :style="{ width: distWidth(count as number, stats.entityCount), background: entityTypeColor(etype as string) }"></div>
              </div>
              <div class="dist-count">{{ count }}</div>
            </div>
          </div>
          <div v-else class="empty-hint">{{ $t('common.noData') }}</div>
        </div>
      </div>

      <!-- Top Accessed -->
      <div class="card" v-if="usageStats.topAccessed?.length">
        <div class="card-header">
          <h3>🔥 {{ $t('dashboard.topAccessed') }}</h3>
        </div>
        <div class="top-accessed-list">
          <div class="accessed-item" v-for="(m, idx) in usageStats.topAccessed" :key="m.id">
            <div class="accessed-rank">{{ idx + 1 }}</div>
            <div class="accessed-info">
              <div class="accessed-key">{{ m.key }}</div>
              <div class="accessed-layer-badge" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</div>
            </div>
            <div class="accessed-count">{{ m.access_count }} {{ $t('dashboard.accessCount') }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Default: Overview -->
    <div v-else>
      <div class="page-header">
        <h1>📊 {{ $t('dashboard.title') }}</h1>
        <p class="page-desc">{{ $t('dashboard.subtitle') }}</p>
      </div>

      <el-alert
        v-if="!stats.passwordSet"
        :title="$t('dashboard.securityAlert')"
        type="warning"
        show-icon
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template #default>
          <span>{{ $t('dashboard.securityAlertDesc') }}</span>
          <el-button type="primary" size="small" text @click="$router.push('/settings')">{{ $t('dashboard.goSettings') }}</el-button>
        </template>
      </el-alert>

      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon-wrap mem-icon">🧠</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.memoryCount }}</div>
            <div class="stat-label">{{ $t('dashboard.memoryCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap entity-icon">🕸️</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.entityCount }}</div>
            <div class="stat-label">{{ $t('dashboard.entityCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap wiki-icon">📖</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.wikiCount }}</div>
            <div class="stat-label">{{ $t('dashboard.wikiCount') }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon-wrap pref-icon">⭐</div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.layerStats.preference || 0 }}</div>
            <div class="stat-label">{{ $t('dashboard.preferenceCount') }}</div>
          </div>
        </div>
      </div>

      <div class="content-grid">
        <div class="card">
          <div class="card-header">
            <h3>📊 {{ $t('dashboard.layerDistribution') }}</h3>
          </div>
          <div class="layer-bars">
            <div class="layer-bar" v-for="(count, layer) in stats.layerStats" :key="layer">
              <div class="layer-label">{{ layerLabels[layer] || layer }}</div>
              <div class="layer-track">
                <div class="layer-fill" :class="String(layer)" :style="{ width: barWidth(count) }"></div>
              </div>
              <div class="layer-count">{{ count }}</div>
            </div>
            <div v-if="Object.keys(stats.layerStats).length === 0" class="empty-hint">{{ $t('common.noData') }}</div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h3>🔑 {{ $t('dashboard.licenseStatus') }}</h3>
          </div>
          <div class="license-info">
            <div class="license-tier" :class="stats.license.tier">
              <span class="tier-icon">{{ stats.license.tier === 'oss' ? '◇' : '◆' }}</span>
              <span>{{ stats.license.tier === 'oss' ? $t('dashboard.freeVersion') : $t('dashboard.proVersion') }}</span>
            </div>
            <div v-if="stats.license.active" class="license-detail">
              <div v-if="stats.license.type">{{ $t('settings.version') }}：{{ stats.license.type === 'pro_annual' ? $t('dashboard.annual') : $t('dashboard.lifetime') }}</div>
              <div v-if="stats.license.expires_at">{{ $t('dashboard.expires') }}：{{ stats.license.expires_at }}</div>
              <div v-if="stats.license.device_slot">{{ $t('dashboard.device') }}：{{ stats.license.device_slot }}</div>
              <el-button type="primary" size="small" style="margin-top: 8px" @click="$router.push('/pro')">{{ $t('nav.pro') }} →</el-button>
            </div>
            <div v-else class="upgrade-hint">
              <p>{{ $t('dashboard.upgradeHint') }}</p>
              <el-button type="primary" size="small" @click="$router.push('/settings')">{{ $t('dashboard.viewUpgrade') }}</el-button>
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>🕐 {{ $t('dashboard.recentMemories') }}</h3>
          <el-button text type="primary" @click="$router.push('/memories')">{{ $t('dashboard.viewAll') }}</el-button>
        </div>
        <div class="recent-memories" v-if="stats.recentMemories && stats.recentMemories.length">
          <div class="memory-item" v-for="m in stats.recentMemories" :key="m.id">
            <div class="memory-layer-badge" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</div>
            <div class="memory-content">
              <div class="memory-key">{{ m.key }}</div>
              <div class="memory-value">{{ truncate(m.value, 80) }}</div>
            </div>
            <div class="memory-time">{{ formatTime(m.updated_at) }}</div>
          </div>
        </div>
        <div v-else class="empty-hint">{{ $t('common.noData') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import axios from '../api/client'

const { t } = useI18n()
const route = useRoute()
const activeTab = computed(() => (route.query.tab as string) || '')

const stats = ref<any>({
  memoryCount: 0, entityCount: 0, relationCount: 0, wikiCount: 0,
  layerStats: {}, recentMemories: [], license: { tier: 'oss', active: false },
  passwordSet: true,
})

// Skills state
const scanning = ref(false)
const scanned = ref(false)
const globalSkills = ref<any[]>([])
const workspaceSkills = ref<any[]>([])
const skillDetailVisible = ref(false)
const skillDetail = ref<any>(null)

// Stats state
const statsPeriod = ref(30)
const usageStats = ref<any>({
  dailyTrend: [],
  dailyTokenTrend: [],
  sourceDistribution: {},
  importanceDistribution: {},
  tokenByLayer: {},
  totalEstimatedTokens: 0,
  topAccessed: [],
  operationCounts: {},
  entityTypeDistribution: {},
  totalMemories: 0,
  days: 30,
})

// Chart dimensions
const chartWidth = 600
const chartHeight = 220
const chartPadding = { left: 50, right: 10, top: 10, bottom: 25 }

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

onMounted(async () => {
  try {
    const { data } = await axios.get('/stats')
    stats.value = data
  } catch {}
})

// Auto-scan skills when switching to skills tab
watch(activeTab, (tab) => {
  if (tab === 'skills' && !scanned.value) scanSkills()
  if (tab === 'stats') loadUsageStats()
}, { immediate: true })

async function scanSkills() {
  scanning.value = true
  try {
    const { data } = await axios.get('/openclaw-skills/scan')
    globalSkills.value = data.global_skills || []
    workspaceSkills.value = data.workspace_skills || []
    scanned.value = true
  } catch {
    globalSkills.value = []
    workspaceSkills.value = []
    scanned.value = true
  } finally {
    scanning.value = false
  }
}

async function showSkillDetail(skill: any) {
  try {
    const { data } = await axios.get('/openclaw-skills/detail', {
      params: { skill_dir: skill.skill_dir, scope: skill.scope },
    })
    skillDetail.value = data
  } catch {
    skillDetail.value = skill
  }
  skillDetailVisible.value = true
}

async function loadUsageStats() {
  try {
    const { data } = await axios.get('/stats/usage', { params: { days: statsPeriod.value } })
    usageStats.value = data
  } catch {}
}

// Chart computed values
const maxDailyCount = computed(() => {
  const trend = usageStats.value.dailyTrend || []
  if (!trend.length) return 1
  return Math.max(...trend.map((d: any) => d.count), 1)
})

const maxTokenCount = computed(() => {
  const trend = usageStats.value.dailyTokenTrend || []
  if (!trend.length) return 1
  return Math.max(...trend.map((d: any) => d.tokens), 1)
})

const xLabelInterval = computed(() => {
  const len = usageStats.value.dailyTrend?.length || 0
  if (len <= 7) return 1
  if (len <= 14) return 2
  if (len <= 30) return 5
  return 10
})

const dailyTrendPoints = computed(() => {
  const trend = usageStats.value.dailyTrend || []
  if (!trend.length) return []
  const plotW = chartWidth - chartPadding.left - chartPadding.right
  const plotH = chartHeight - chartPadding.top - chartPadding.bottom
  const max = maxDailyCount.value
  return trend.map((d: any, i: number) => ({
    x: chartPadding.left + (i / Math.max(trend.length - 1, 1)) * plotW,
    y: chartPadding.top + plotH - (d.count / max) * plotH,
    label: d.date.slice(5), // MM-DD
    count: d.count,
  }))
})

const dailyTrendLinePath = computed(() => {
  const pts = dailyTrendPoints.value
  if (!pts.length) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ')
})

const dailyTrendAreaPath = computed(() => {
  const pts = dailyTrendPoints.value
  if (!pts.length) return ''
  const plotH = chartHeight - chartPadding.top - chartPadding.bottom
  const baseY = chartPadding.top + plotH
  const linePath = pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ')
  return `${linePath} L${pts[pts.length - 1].x},${baseY} L${pts[0].x},${baseY} Z`
})

const tokenBarPoints = computed(() => {
  const trend = usageStats.value.dailyTokenTrend || []
  if (!trend.length) return []
  const plotW = chartWidth - chartPadding.left - chartPadding.right
  const plotH = chartHeight - chartPadding.top - chartPadding.bottom
  const max = maxTokenCount.value
  const barW = Math.max(plotW / trend.length - 2, 2)
  const colors = ['#3b82f6', '#6366f1', '#8b5cf6', '#a855f7', '#3b82f6']
  return trend.map((d: any, i: number) => ({
    x: chartPadding.left + (i / trend.length) * plotW + 1,
    y: chartPadding.top + plotH - (d.tokens / max) * plotH,
    w: barW,
    h: (d.tokens / max) * plotH,
    color: colors[i % colors.length],
    label: d.date.slice(5),
    cumulative: d.cumulative,
  }))
})

const cumulativeLinePath = computed(() => {
  const trend = usageStats.value.dailyTokenTrend || []
  if (!trend.length) return ''
  const plotW = chartWidth - chartPadding.left - chartPadding.right
  const plotH = chartHeight - chartPadding.top - chartPadding.bottom
  const maxCum = Math.max(...trend.map((d: any) => d.cumulative), 1)
  return trend.map((d: any, i: number) => {
    const x = chartPadding.left + (i / Math.max(trend.length - 1, 1)) * plotW
    const y = chartPadding.top + plotH - (d.cumulative / maxCum) * plotH
    return `${i === 0 ? 'M' : 'L'}${x},${y}`
  }).join(' ')
})

// Helper functions
function sourceLabel(source: string) {
  const map: Record<string, string> = {
    manual: t('dashboard.manual'),
    import: t('dashboard.import'),
    auto: t('dashboard.auto'),
  }
  return map[source] || source
}

function importanceLabel(level: string) {
  const map: Record<string, string> = {
    high: t('dashboard.high'),
    medium: t('dashboard.medium'),
    low: t('dashboard.low'),
  }
  return map[level] || level
}

function entityTypeColor(etype: string) {
  const colors: Record<string, string> = {
    person: '#3b82f6', project: '#10b981', tool: '#8b5cf6', concept: '#f59e0b',
    event: '#ef4444', tech: '#06b6d4', organization: '#ec4899', other: '#6b7280',
  }
  return colors[etype] || '#6b7280'
}

function barWidth(count: number) {
  const max = Math.max(...Object.values(stats.value.layerStats || {}).map(Number), 1)
  return `${(count / max) * 100}%`
}

function distWidth(count: number, total: number) {
  if (!total) return '0%'
  return `${(count / total) * 100}%`
}

function truncate(str: string, len: number) {
  return str && str.length > len ? str.slice(0, len) + '...' : str
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatNumber(n: number) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}
</script>

<style scoped>
.dashboard {
  padding: 28px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header { margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; }

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--cm-text);
  margin: 0;
}

.page-desc {
  color: var(--cm-text-muted);
  font-size: 14px;
  margin-top: 4px;
  width: 100%;
}

.period-selector { flex-shrink: 0; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--cm-primary), transparent);
  opacity: 0;
  transition: opacity 0.3s;
}

.stat-card:hover { border-color: rgba(16, 185, 129, 0.3); box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1); transform: translateY(-1px); }
.stat-card:hover::before { opacity: 1; }

.stat-icon-wrap { width: 44px; height: 44px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 22px; }
.mem-icon { background: rgba(16, 185, 129, 0.12); }
.entity-icon { background: rgba(6, 182, 212, 0.12); }
.rel-icon { background: rgba(59, 130, 246, 0.12); }
.token-icon { background: rgba(139, 92, 246, 0.12); }
.wiki-icon { background: rgba(255, 193, 7, 0.12); }
.pref-icon { background: rgba(233, 30, 99, 0.12); }
.stat-value { font-size: 28px; font-weight: 700; color: var(--cm-text); }
.stat-label { font-size: 13px; color: var(--cm-text-muted); margin-top: 2px; }

.content-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.card-header h3 { margin: 0; font-size: 15px; color: var(--cm-text); font-weight: 600; }

.chart-card { min-height: 300px; }
.chart-container { width: 100%; min-height: 220px; }
.trend-chart { width: 100%; height: auto; display: block; }
.chart-legend { display: flex; gap: 16px; justify-content: center; margin-top: 12px; }
.legend-item { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--cm-text-muted); }
.legend-color { width: 12px; height: 4px; border-radius: 2px; }

/* Layer bars */
.layer-bars { display: flex; flex-direction: column; gap: 12px; }
.layer-bar { display: flex; align-items: center; gap: 12px; }
.layer-label { width: 56px; font-size: 13px; color: var(--cm-text-muted); text-align: right; }
.layer-track { flex: 1; height: 6px; background: var(--cm-border); border-radius: 3px; overflow: hidden; }
.layer-fill { height: 100%; border-radius: 3px; transition: width 0.6s ease; }
.layer-fill.preference { background: #10B981; }
.layer-fill.knowledge { background: #06b6d4; }
.layer-fill.short_term { background: #ffc107; }
.layer-fill.private { background: #e91e63; }
.layer-count { width: 32px; font-size: 13px; color: var(--cm-text-muted); }

/* Distribution bars */
.distribution-bars { display: flex; flex-direction: column; gap: 10px; }
.dist-bar { display: flex; align-items: center; gap: 10px; }
.dist-label { width: 64px; font-size: 13px; color: var(--cm-text-muted); text-align: right; flex-shrink: 0; }
.dist-track { flex: 1; height: 8px; background: var(--cm-border); border-radius: 4px; overflow: hidden; }
.dist-fill { height: 100%; border-radius: 4px; transition: width 0.6s ease; min-width: 2px; }
.dist-fill.manual { background: #3b82f6; }
.dist-fill.import { background: #f59e0b; }
.dist-fill.auto { background: #10B981; }
.dist-fill.importance.high { background: #10B981; }
.dist-fill.importance.medium { background: #f59e0b; }
.dist-fill.importance.low { background: #ef4444; }
.dist-fill.layer.preference { background: #10B981; }
.dist-fill.layer.knowledge { background: #06b6d4; }
.dist-fill.layer.short_term { background: #ffc107; }
.dist-fill.layer.private { background: #e91e63; }
.dist-fill.entity { border-radius: 4px; }
.dist-count { width: 56px; font-size: 13px; color: var(--cm-text-muted); text-align: right; flex-shrink: 0; }

/* Top accessed */
.top-accessed-list { display: flex; flex-direction: column; gap: 8px; }
.accessed-item { display: flex; align-items: center; gap: 12px; padding: 10px 12px; border-radius: 8px; background: var(--cm-bg); border: 1px solid var(--cm-border); }
.accessed-rank { width: 28px; height: 28px; border-radius: 50%; background: rgba(16,185,129,0.12); color: #10B981; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 700; flex-shrink: 0; }
.accessed-info { flex: 1; min-width: 0; display: flex; align-items: center; gap: 8px; }
.accessed-key { font-size: 14px; font-weight: 600; color: var(--cm-text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.accessed-layer-badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; white-space: nowrap; flex-shrink: 0; }
.accessed-layer-badge.preference { background: rgba(16,185,129,0.15); color: #10B981; }
.accessed-layer-badge.knowledge { background: rgba(6,182,212,0.15); color: #06b6d4; }
.accessed-layer-badge.short_term { background: rgba(255,193,7,0.15); color: #ffc107; }
.accessed-layer-badge.private { background: rgba(233,30,99,0.15); color: #e91e63; }
.accessed-count { font-size: 13px; color: var(--cm-text-muted); white-space: nowrap; }

.license-tier { display: flex; align-items: center; gap: 8px; font-size: 18px; font-weight: 600; margin-bottom: 12px; }
.license-tier.oss { color: var(--cm-text-muted); }
.license-tier.pro { color: #10B981; }
.license-detail { font-size: 13px; color: var(--cm-text-muted); line-height: 1.8; }
.upgrade-hint { margin-top: 8px; font-size: 13px; color: var(--cm-text-muted); }
.upgrade-hint p { margin: 0 0 8px; }

.recent-memories { display: flex; flex-direction: column; gap: 8px; }
.memory-item { display: flex; align-items: center; gap: 12px; padding: 10px 12px; border-radius: 8px; background: var(--cm-bg); border: 1px solid var(--cm-border); }
.memory-layer-badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; white-space: nowrap; }
.memory-layer-badge.preference { background: rgba(16,185,129,0.15); color: #10B981; }
.memory-layer-badge.knowledge { background: rgba(6,182,212,0.15); color: #06b6d4; }
.memory-layer-badge.short_term { background: rgba(255,193,7,0.15); color: #ffc107; }
.memory-layer-badge.private { background: rgba(233,30,99,0.15); color: #e91e63; }
.memory-content { flex: 1; min-width: 0; }
.memory-key { font-size: 14px; font-weight: 600; color: var(--cm-text); }
.memory-value { font-size: 12px; color: var(--cm-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.memory-time { font-size: 12px; color: var(--cm-text-placeholder); white-space: nowrap; }

.empty-hint { text-align: center; padding: 32px; color: var(--cm-text-placeholder); }

/* Skills embed styles */
.skill-section { margin-bottom: 24px; }
.section-title { font-size: 15px; font-weight: 600; color: var(--cm-text); margin: 0 0 12px; padding-bottom: 8px; border-bottom: 1px solid var(--cm-border); }
.skill-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.skill-card { display: flex; align-items: center; gap: 12px; padding: 14px; border-radius: 10px; border: 1px solid var(--cm-border); background: var(--cm-bg-secondary); cursor: pointer; transition: all 0.2s; }
.skill-card:hover { border-color: var(--cm-primary); box-shadow: 0 2px 8px rgba(var(--cm-primary-rgb), 0.1); }
.skill-icon { font-size: 24px; flex-shrink: 0; }
.skill-info { flex: 1; min-width: 0; }
.skill-name { font-weight: 600; font-size: 14px; color: var(--cm-text); }
.skill-desc { font-size: 12px; color: var(--cm-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.skill-meta { display: flex; flex-direction: column; align-items: flex-end; gap: 4px; flex-shrink: 0; }
.badge { font-size: 11px; padding: 2px 8px; border-radius: 10px; background: rgba(var(--cm-primary-rgb), 0.1); color: var(--cm-primary); }
.scope { font-size: 10px; padding: 1px 6px; border-radius: 4px; text-transform: uppercase; }
.scope.global { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.scope.workspace { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.scope.agents { background: rgba(168, 85, 247, 0.1); color: #a855f7; }
.scope.workspace-agents { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
.scope.workspace-legacy { background: rgba(107, 114, 128, 0.1); color: #6b7280; }

/* Skill detail dialog */
.detail-row { margin-bottom: 10px; }
.detail-body { margin-top: 16px; }
.detail-body pre { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 8px; padding: 12px; overflow-x: auto; font-size: 13px; max-height: 400px; overflow-y: auto; white-space: pre-wrap; }
.file-list { font-size: 13px; color: var(--cm-text-muted); margin-top: 4px; }

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .content-grid { grid-template-columns: 1fr; }
  .skill-grid { grid-template-columns: 1fr; }
}
</style>
