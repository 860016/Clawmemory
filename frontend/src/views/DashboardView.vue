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
        <div v-if="allSkills.length" class="skill-grid">
          <div v-for="skill in allSkills" :key="skill.skill_dir" class="skill-card" @click="showSkillDetail(skill)">
            <div class="skill-icon">{{ skill.scope === 'global' ? '🌐' : '📁' }}</div>
            <div class="skill-info">
              <div class="skill-name">{{ skill.name }}</div>
              <div class="skill-desc">{{ skill.description || '—' }}</div>
            </div>
            <div class="skill-meta">
              <span class="badge">{{ skill.version }}</span>
              <span class="scope" :class="skill.scope">{{ skill.scope }}</span>
            </div>
          </div>
        </div>
        <div v-else class="empty-hint">{{ $t('skills.noSkills') }}</div>
      </div>
      <el-dialog v-model="skillDetailVisible" :title="skillDetail?.name" width="500px">
        <div v-if="skillDetail" style="font-size:14px;line-height:1.8">
          <div><strong>{{ $t('skills.name') }}:</strong> {{ skillDetail.name }}</div>
          <div><strong>{{ $t('skills.description') }}:</strong> {{ skillDetail.description || '—' }}</div>
          <div><strong>{{ $t('skills.version') }}:</strong> {{ skillDetail.version }}</div>
          <div><strong>{{ $t('skills.author') }}:</strong> {{ skillDetail.author }}</div>
          <div><strong>{{ $t('skills.scope') }}:</strong> {{ skillDetail.scope }}</div>
        </div>
      </el-dialog>
    </div>

    <!-- Default: Overview + Stats -->
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
  memoryCount: 0, entityCount: 0, wikiCount: 0,
  layerStats: {}, recentMemories: [], license: { tier: 'oss', active: false },
  passwordSet: true,
})

// Skills state
const scanning = ref(false)
const scanned = ref(false)
const globalSkills = ref<any[]>([])
const workspaceSkills = ref<any[]>([])
const allSkills = computed(() => [...globalSkills.value, ...workspaceSkills.value])
const skillDetailVisible = ref(false)
const skillDetail = ref<any>(null)

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

function showSkillDetail(skill: any) {
  skillDetail.value = skill
  skillDetailVisible.value = true
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
}
</script>

<style scoped>
.dashboard {
  padding: 28px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header { margin-bottom: 28px; display: flex; justify-content: space-between; align-items: center; }

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
}

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
.wiki-icon { background: rgba(255, 193, 7, 0.12); }
.pref-icon { background: rgba(233, 30, 99, 0.12); }
.stat-value { font-size: 28px; font-weight: 700; color: var(--cm-text); }
.stat-label { font-size: 13px; color: var(--cm-text-muted); margin-top: 2px; }

.content-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.card-header h3 { margin: 0; font-size: 15px; color: var(--cm-text); font-weight: 600; }

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

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .content-grid { grid-template-columns: 1fr; }
  .skill-grid { grid-template-columns: 1fr; }
}
</style>
