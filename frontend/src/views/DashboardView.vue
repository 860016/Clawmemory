<template>
  <div class="dashboard">
    <div class="page-header">
      <h1>{{ $t('dashboard.title') }}</h1>
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
          <div v-if="Object.keys(stats.layerStats).length === 0" class="empty-hint">{{ $t('common.noData') }}</div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>{{ $t('dashboard.licenseStatus') }}</h3>
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
        <h3>{{ $t('dashboard.recentMemories') }}</h3>
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
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Collection, Connection, Document, Star } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const stats = ref<any>({
  memoryCount: 0, entityCount: 0, wikiCount: 0,
  layerStats: {}, recentMemories: [], license: { tier: 'oss', active: false },
  passwordSet: true,
})

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
}
</script>

<style scoped>
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
</style>
