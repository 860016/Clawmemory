<template>
<<<<<<< HEAD
  <div class="dashboard-view">
    <h2 class="page-title">控制面板</h2>

    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: linear-gradient(135deg, #4f6ef7, #7b93fa)">
            <el-icon :size="22"><Collection /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.memoryCount }}</div>
            <div class="stat-label">记忆总数</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: linear-gradient(135deg, #67c23a, #85ce61)">
            <el-icon :size="22"><Connection /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.entityCount }}</div>
            <div class="stat-label">实体数</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: linear-gradient(135deg, #e6a23c, #f0c78a)">
            <el-icon :size="22"><ChatDotRound /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.sessionCount }}</div>
            <div class="stat-label">会话数</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-icon" style="background: linear-gradient(135deg, #f56c6c, #f89898)">
            <el-icon :size="22"><MagicStick /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.skillCount }}</div>
            <div class="stat-label">技能数</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="14">
        <el-card shadow="hover" class="section-card">
          <template #header>
            <div class="card-header">
              <span>最近记忆</span>
              <el-button text type="primary" @click="$router.push('/memories')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentMemories" size="small" stripe :show-header="true">
            <el-table-column prop="key" label="键" show-overflow-tooltip />
            <el-table-column prop="layer" label="层级" width="90">
              <template #default="{ row }">
                <el-tag size="small" :type="layerTagType(row.layer)">{{ row.layer }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" width="160" />
          </el-table>
          <el-empty v-if="!recentMemories.length" description="暂无记忆" :image-size="60" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="hover" class="section-card">
          <template #header>
            <div class="card-header">
              <span>最近会话</span>
              <el-button text type="primary" @click="$router.push('/chat')">查看全部</el-button>
            </div>
          </template>
          <div v-for="s in recentSessions" :key="s.id" class="session-item" @click="$router.push('/chat')">
            <div class="session-title">{{ s.title || `会话 ${s.id}` }}</div>
            <div class="session-time">{{ formatTime(s.created_at) }}</div>
          </div>
          <el-empty v-if="!recentSessions.length" description="暂无会话" :image-size="60" />
        </el-card>

        <el-card shadow="hover" class="section-card" style="margin-top: 16px">
          <template #header><span>授权状态</span></template>
          <div class="license-info">
            <div class="license-row">
              <span class="license-label">版本</span>
              <el-tag :type="licenseInfo?.tier === 'oss' ? 'info' : 'success'" effect="dark" round>
                {{ licenseInfo?.tier?.toUpperCase() || 'OSS' }}
              </el-tag>
            </div>
            <div class="license-row">
              <span class="license-label">状态</span>
              <el-tag :type="licenseInfo?.active ? 'success' : 'warning'" effect="plain" round>
                {{ licenseInfo?.active ? '已激活' : '未激活' }}
              </el-tag>
            </div>
            <div class="license-row">
              <span class="license-label">到期</span>
              <span>{{ licenseInfo?.expires_at || '永久' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
=======
  <div class="dashboard">
    <div class="page-header">
      <h1>{{ $t('dashboard.title') }}</h1>
      <p class="page-desc">{{ $t('dashboard.subtitle') }}</p>
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon-wrap mem-icon">
          <el-icon :size="22"><Collection /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.memoryCount }}</div>
          <div class="stat-label">{{ $t('dashboard.memoryCount') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap entity-icon">
          <el-icon :size="22"><Connection /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.entityCount }}</div>
          <div class="stat-label">{{ $t('dashboard.entityCount') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap wiki-icon">
          <el-icon :size="22"><Document /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.wikiCount }}</div>
          <div class="stat-label">{{ $t('dashboard.wikiCount') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap pref-icon">
          <el-icon :size="22"><Star /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.layerStats.preference || 0 }}</div>
          <div class="stat-label">{{ $t('dashboard.preferenceCount') }}</div>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="card">
        <div class="card-header">
          <h3>{{ $t('dashboard.layerDistribution') }}</h3>
        </div>
        <div class="layer-bars">
          <div class="layer-bar" v-for="(count, layer) in stats.layerStats" :key="layer">
            <div class="layer-label">{{ layerLabels[layer] || layer }}</div>
            <div class="layer-track">
              <div class="layer-fill" :class="layer" :style="{ width: barWidth(count) }"></div>
            </div>
            <div class="layer-count">{{ count }}</div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>{{ $t('dashboard.licenseStatus') }}</h3>
        </div>
        <div class="license-info">
          <div class="license-tier" :class="license.tier">
            <span class="tier-icon">{{ license.tier === 'oss' ? '◇' : '◆' }}</span>
            <span>{{ license.tier === 'oss' ? $t('dashboard.freeVersion') : $t('dashboard.proVersion') }}</span>
          </div>
          <div v-if="license.tier !== 'oss'" class="license-detail">
            <div v-if="license.type">{{ $t('settings.version') }}：{{ license.type === 'pro_annual' ? $t('dashboard.annual') : $t('dashboard.lifetime') }}</div>
            <div v-if="license.expires_at">{{ $t('dashboard.expires') }}：{{ license.expires_at }}</div>
            <div v-if="license.device_slot">{{ $t('dashboard.device') }}：{{ license.device_slot }}</div>
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
        <h3>{{ $t('dashboard.recentMemories') }}</h3>
        <el-button text type="primary" @click="$router.push('/memories')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
      <div class="recent-memories" v-if="recentMemories.length">
        <div class="memory-item" v-for="m in recentMemories" :key="m.id">
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
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Collection, Connection, Document, Star } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const stats = ref<any>({ memoryCount: 0, entityCount: 0, relationCount: 0, wikiCount: 0, layerStats: {} })
const recentMemories = ref<any[]>([])
const license = ref<any>({ tier: 'oss', features: [] })

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

onMounted(async () => {
<<<<<<< HEAD
  const promises = [
    api.get('/memories', { params: { page: 1, size: 5 } }).catch(() => ({ data: { items: [], total: 0 } })),
    api.get('/knowledge/entities', { params: { page: 1, size: 1 } }).catch(() => ({ data: { items: [], total: 0 } })),
    api.get('/chat/sessions', { params: { limit: 5 } }).catch(() => ({ data: [] })),
    api.get('/skills').catch(() => ({ data: [] })),
    api.get('/license/info').catch(() => ({ data: null })),
  ]
  const [memResp, entityResp, sessionResp, skillResp, licResp] = await Promise.all(promises)

  stats.memoryCount = memResp.data.total ?? memResp.data.items?.length ?? 0
  recentMemories.value = memResp.data.items ?? memResp.data ?? []

  stats.entityCount = entityResp.data.total ?? entityResp.data.items?.length ?? 0

  const sessions = Array.isArray(sessionResp.data) ? sessionResp.data : sessionResp.data.items ?? []
  stats.sessionCount = sessions.length
  recentSessions.value = sessions.slice(0, 5)

  const skills = Array.isArray(skillResp.data) ? skillResp.data : skillResp.data.items ?? []
  stats.skillCount = skills.length

  licenseInfo.value = licResp.data
  if (licResp.data) localStorage.setItem('licenseInfo', JSON.stringify(licResp.data))
})

function layerTagType(layer: string) {
  const map: Record<string, string> = { preference: 'warning', knowledge: 'primary', short_term: 'info', private: 'danger' }
  return map[layer] || 'info'
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
=======
  await Promise.all([loadStats(), loadRecent(), loadLicense()])
})

async function loadStats() {
  try {
    const { data } = await axios.get('/api/v1/memories', { params: { size: 1 } })
    const total = data.total || 0
    const layerStats: Record<string, number> = {}
    for (const layer of ['preference', 'knowledge', 'short_term', 'private']) {
      const { data: ld } = await axios.get('/api/v1/memories', { params: { layer, size: 1 } })
      if (ld.total > 0) layerStats[layer] = ld.total
    }
    stats.value = { memoryCount: total, entityCount: 0, relationCount: 0, wikiCount: 0, layerStats }
  } catch {}
  try {
    const { data: entities } = await axios.get('/api/v1/knowledge/entities')
    stats.value.entityCount = entities.length || 0
  } catch {}
  try {
    const { data: pages } = await axios.get('/api/v1/wiki/pages')
    stats.value.wikiCount = pages.length || 0
  } catch {}
}

async function loadRecent() {
  try {
    const { data } = await axios.get('/api/v1/memories', { params: { size: 5 } })
    recentMemories.value = data.items || []
  } catch {}
}

async function loadLicense() {
  try {
    const { data } = await axios.get('/api/v1/license/info')
    license.value = data
  } catch {}
}

function barWidth(count: number) {
  const max = Math.max(...Object.values(stats.value.layerStats || {}).map(Number), 1)
  return `${(count / max) * 100}%`
}

function truncate(str: string, len: number) {
  return str && str.length > len ? str.slice(0, len) + '...' : str
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
}
</script>

<style scoped>
<<<<<<< HEAD
.dashboard-view { padding: 24px; }
.page-title { font-size: 20px; font-weight: 700; color: #303133; margin: 0 0 20px; }

.stats-row { margin-bottom: 4px; }
.stat-card {
  background: #fff; border-radius: var(--cm-radius); padding: 20px;
  display: flex; align-items: center; gap: 16px;
  box-shadow: var(--cm-card-shadow); transition: transform 0.2s;
}
.stat-card:hover { transform: translateY(-2px); }
.stat-icon {
  width: 48px; height: 48px; border-radius: 12px; display: flex;
  align-items: center; justify-content: center; color: #fff; flex-shrink: 0;
}
.stat-value { font-size: 28px; font-weight: 700; color: #303133; line-height: 1; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }

.section-card { border-radius: var(--cm-radius) !important; }
.card-header { display: flex; justify-content: space-between; align-items: center; font-weight: 600; }

.session-item {
  padding: 10px 0; border-bottom: 1px solid #f0f0f0; cursor: pointer; transition: background 0.2s;
}
.session-item:hover { background: #f5f7fa; margin: 0 -12px; padding: 10px 12px; border-radius: 6px; }
.session-item:last-child { border-bottom: none; }
.session-title { font-size: 14px; color: #303133; font-weight: 500; }
.session-time { font-size: 12px; color: #c0c4cc; margin-top: 4px; }

.license-info { display: flex; flex-direction: column; gap: 10px; }
.license-row { display: flex; justify-content: space-between; align-items: center; font-size: 14px; }
.license-label { color: #909399; }
=======
.dashboard {
  padding: 28px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header { margin-bottom: 28px; }

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: #e6edf3;
  margin: 0;
}

.page-desc {
  color: #7d8590;
  font-size: 14px;
  margin-top: 4px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: #161b22;
  border: 1px solid #21262d;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: border-color 0.2s;
}

.stat-card:hover {
  border-color: rgba(0, 212, 170, 0.3);
}

.stat-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mem-icon { background: rgba(0, 212, 170, 0.12); color: #00d4aa; }
.entity-icon { background: rgba(0, 188, 212, 0.12); color: #00bcd4; }
.wiki-icon { background: rgba(255, 193, 7, 0.12); color: #ffc107; }
.pref-icon { background: rgba(233, 30, 99, 0.12); color: #e91e63; }

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #e6edf3;
}

.stat-label {
  font-size: 13px;
  color: #7d8590;
  margin-top: 2px;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.card {
  background: #161b22;
  border: 1px solid #21262d;
  border-radius: 12px;
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-header h3 {
  margin: 0;
  font-size: 15px;
  color: #e6edf3;
  font-weight: 600;
}

.layer-bars {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.layer-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.layer-label {
  width: 56px;
  font-size: 13px;
  color: #7d8590;
  text-align: right;
}

.layer-track {
  flex: 1;
  height: 6px;
  background: #21262d;
  border-radius: 3px;
  overflow: hidden;
}

.layer-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.6s ease;
}

.layer-fill.preference { background: #00d4aa; }
.layer-fill.knowledge { background: #00bcd4; }
.layer-fill.short_term { background: #ffc107; }
.layer-fill.private { background: #e91e63; }

.layer-count {
  width: 32px;
  font-size: 13px;
  color: #7d8590;
}

.license-tier {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.license-tier.oss { color: #7d8590; }
.license-tier.pro { color: #00d4aa; }

.license-detail {
  font-size: 13px;
  color: #7d8590;
  line-height: 1.8;
}

.upgrade-hint {
  margin-top: 8px;
  font-size: 13px;
  color: #7d8590;
}

.upgrade-hint p { margin: 0 0 8px; }

.recent-memories {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.memory-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #0d1117;
  border: 1px solid #21262d;
}

.memory-layer-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.memory-layer-badge.preference { background: rgba(0,212,170,0.15); color: #00d4aa; }
.memory-layer-badge.knowledge { background: rgba(0,188,212,0.15); color: #00bcd4; }
.memory-layer-badge.short_term { background: rgba(255,193,7,0.15); color: #ffc107; }
.memory-layer-badge.private { background: rgba(233,30,99,0.15); color: #e91e63; }

.memory-content {
  flex: 1;
  min-width: 0;
}

.memory-key {
  font-size: 14px;
  font-weight: 600;
  color: #e6edf3;
}

.memory-value {
  font-size: 12px;
  color: #7d8590;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.memory-time {
  font-size: 12px;
  color: #484f58;
  white-space: nowrap;
}

.empty-hint {
  text-align: center;
  padding: 32px;
  color: #484f58;
}

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .content-grid { grid-template-columns: 1fr; }
}
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
</style>
