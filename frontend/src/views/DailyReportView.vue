<template>
  <div class="daily-report-page">
    <div class="page-header">
      <h1>📊 {{ $t('dailyReport.title') }}</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleGenerateToday" :loading="generating">
          <el-icon><Refresh /></el-icon> {{ $t('dailyReport.generateToday') }}
        </el-button>
      </div>
    </div>

    <!-- 统计摘要 -->
    <div class="stats-summary">
      <div class="stat-card" v-for="stat in periodStats" :key="stat.label">
        <div class="stat-value">{{ stat.value }}</div>
        <div class="stat-label">{{ stat.label }}</div>
      </div>
    </div>

    <!-- 日报列表 -->
    <div class="report-list">
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>{{ $t('common.loading') }}</span>
      </div>

      <div v-else-if="reports.length === 0" class="empty-state">
        <div class="empty-icon">📄</div>
        <p>{{ $t('dailyReport.empty') }}</p>
        <p class="empty-hint">{{ $t('dailyReport.emptyHint') }}</p>
      </div>

      <div v-else class="report-cards">
        <div class="report-card" v-for="r in reports" :key="r.id" @click="openReport(r)">
          <div class="report-date">{{ formatDate(r.report_date) }}</div>
          <div class="report-summary">{{ r.summary }}</div>
          <div class="report-stats">
            <span class="stat-badge" v-if="r.stats?.new_memories">
              📝 {{ r.stats.new_memories }} {{ $t('dailyReport.memories') }}
            </span>
            <span class="stat-badge" v-if="r.stats?.new_entities">
              🔗 {{ r.stats.new_entities }} {{ $t('dailyReport.entities') }}
            </span>
            <span class="stat-badge" v-if="r.stats?.updated_wiki">
              📖 {{ r.stats.updated_wiki }} {{ $t('dailyReport.wiki') }}
            </span>
          </div>
          <div class="report-footer">
            <span class="report-time">{{ formatTime(r.generated_at) }}</span>
            <span class="report-pushed" v-if="r.is_pushed">✅ {{ $t('dailyReport.pushed') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 日报详情弹窗 -->
    <el-dialog v-model="showDetail" :title="$t('dailyReport.detail')" width="700px" class="report-detail-dialog">
      <div v-if="currentReport" class="detail-content">
        <div class="detail-header">
          <h2>{{ formatDate(currentReport.report_date) }}</h2>
          <span class="detail-model">{{ currentReport.ai_model }}</span>
        </div>

        <div class="detail-section">
          <h3>📋 {{ $t('dailyReport.summary') }}</h3>
          <p>{{ currentReport.summary }}</p>
        </div>

        <div class="detail-section" v-if="currentReport.highlights?.length">
          <h3>✨ {{ $t('dailyReport.highlights') }}</h3>
          <ul>
            <li v-for="(h, i) in currentReport.highlights" :key="i">{{ h }}</li>
          </ul>
        </div>

        <div class="detail-section" v-if="currentReport.knowledge_gained?.length">
          <h3>💡 {{ $t('dailyReport.knowledgeGained') }}</h3>
          <ul>
            <li v-for="(k, i) in currentReport.knowledge_gained" :key="i">{{ k }}</li>
          </ul>
        </div>

        <div class="detail-section" v-if="currentReport.pending_tasks?.length">
          <h3>⏳ {{ $t('dailyReport.pendingTasks') }}</h3>
          <ul>
            <li v-for="(t, i) in currentReport.pending_tasks" :key="i">{{ t }}</li>
          </ul>
        </div>

        <div class="detail-section" v-if="currentReport.tomorrow_suggestions?.length">
          <h3>🎯 {{ $t('dailyReport.tomorrowSuggestions') }}</h3>
          <ul>
            <li v-for="(s, i) in currentReport.tomorrow_suggestions" :key="i">{{ s }}</li>
          </ul>
        </div>

        <div class="detail-stats">
          <div class="detail-stat" v-for="(v, k) in currentReport.stats" :key="k">
            <span class="detail-stat-value">{{ v }}</span>
            <span class="detail-stat-label">{{ statLabel(String(k)) }}</span>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh, Loading } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const reports = ref<any[]>([])
const loading = ref(false)
const generating = ref(false)
const showDetail = ref(false)
const currentReport = ref<any>(null)
const periodStats = ref<any[]>([])

onMounted(async () => {
  await loadReports()
  await loadStats()
})

async function loadReports() {
  loading.value = true
  try {
    const { data } = await axios.get('/reports/daily', { params: { limit: 30 } })
    reports.value = data || []
  } catch {
    reports.value = []
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    const { data } = await axios.get('/reports/daily/stats', { params: { days: 7 } })
    const s = data || {}
    periodStats.value = [
      { label: t('dailyReport.periodDays'), value: s.period_days || 7 },
      { label: t('dailyReport.reportCount'), value: s.report_count || 0 },
      { label: t('dailyReport.totalMemories'), value: s.total_memories || 0 },
      { label: t('dailyReport.totalEntities'), value: s.total_entities || 0 },
      { label: t('dailyReport.totalWikiUpdated'), value: s.total_wiki_updated || 0 },
    ]
  } catch {}
}

async function handleGenerateToday() {
  generating.value = true
  try {
    const { data } = await axios.post('/reports/daily/generate')
    if (data.success) {
      ElMessage.success(data.message || t('dailyReport.generated'))
      await loadReports()
      await loadStats()
    } else {
      const reasons = data.reasons || []
      const reasonTexts = reasons.map((r: string) => t(`dailyReport.reason.${r}`))
      ElMessage.warning(`${t('dailyReport.generateSkipped')}: ${reasonTexts.join('、')}`)
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.detail || t('common.failed'))
  } finally {
    generating.value = false
  }
}

function openReport(r: any) {
  currentReport.value = r
  showDetail.value = true
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const weekdays = [t('dailyReport.sun'), t('dailyReport.mon'), t('dailyReport.tue'), t('dailyReport.wed'), t('dailyReport.thu'), t('dailyReport.fri'), t('dailyReport.sat')]
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${weekdays[d.getDay()]}`
}

function formatTime(dateStr: string | null) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function statLabel(key: string) {
  const map: Record<string, string> = {
    new_memories: t('dailyReport.memories'),
    new_entities: t('dailyReport.entities'),
    new_relations: t('dailyReport.relations'),
    updated_wiki: t('dailyReport.wiki'),
  }
  return map[key] || key
}
</script>

<style scoped>
.daily-report-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-wrap: wrap; gap: 12px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.header-actions { display: flex; gap: 8px; }

/* Stats Summary */
.stats-summary { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 12px; margin-bottom: 24px; }
.stat-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 10px; padding: 16px; text-align: center; }
.stat-value { font-size: 28px; font-weight: 700; color: #10B981; }
.stat-label { font-size: 12px; color: var(--cm-text-muted); margin-top: 4px; }

/* Report List */
.report-list { }
.loading-state, .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 300px; color: var(--cm-text-placeholder); }
.empty-icon { font-size: 48px; margin-bottom: 12px; }
.empty-hint { font-size: 13px; color: var(--cm-text-muted); }

/* Report Cards */
.report-cards { display: flex; flex-direction: column; gap: 12px; }
.report-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; cursor: pointer; transition: border-color 0.2s, transform 0.15s; }
.report-card:hover { border-color: rgba(16,185,129,0.3); transform: translateY(-1px); }
.report-date { font-size: 16px; font-weight: 700; color: var(--cm-text); margin-bottom: 8px; }
.report-summary { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.6; margin-bottom: 12px; }
.report-stats { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 8px; }
.stat-badge { padding: 2px 10px; background: var(--cm-border); border-radius: 12px; font-size: 12px; color: var(--cm-text-secondary); }
.report-footer { display: flex; justify-content: space-between; align-items: center; }
.report-time { font-size: 12px; color: var(--cm-text-muted); }
.report-pushed { font-size: 12px; color: #10B981; }

/* Detail Dialog */
.detail-content { }
.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; padding-bottom: 12px; border-bottom: 1px solid var(--cm-border); }
.detail-header h2 { font-size: 20px; color: var(--cm-text); margin: 0; }
.detail-model { font-size: 12px; color: var(--cm-text-muted); background: var(--cm-border); padding: 2px 8px; border-radius: 4px; }
.detail-section { margin-bottom: 20px; }
.detail-section h3 { font-size: 15px; color: var(--cm-text); margin: 0 0 10px; }
.detail-section p { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.6; }
.detail-section ul { margin: 0; padding-left: 20px; }
.detail-section li { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.8; }
.detail-stats { display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: 12px; padding-top: 16px; border-top: 1px solid var(--cm-border); }
.detail-stat { text-align: center; }
.detail-stat-value { font-size: 20px; font-weight: 700; color: #10B981; display: block; }
.detail-stat-label { font-size: 11px; color: var(--cm-text-muted); }

@media (max-width: 768px) {
  .stats-summary { grid-template-columns: repeat(2, 1fr); }
  .detail-stats { grid-template-columns: repeat(2, 1fr); }
}
</style>
