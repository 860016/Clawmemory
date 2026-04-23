<template>
  <div class="daily-report-v2">
    <!-- Header -->
    <header class="report-header">
      <div class="header-content">
        <div class="header-title">
          <div class="title-icon gradient-bg">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
              <line x1="16" y1="2" x2="16" y2="6"/>
              <line x1="8" y1="2" x2="8" y2="6"/>
              <line x1="3" y1="10" x2="21" y2="10"/>
            </svg>
          </div>
          <div>
            <h1>{{ $t('dailyReport.title') }}</h1>
            <p class="subtitle">{{ $t('dailyReport.subtitle') }}</p>
          </div>
        </div>
        
        <div class="header-actions">
          <div class="date-navigator">
            <button class="nav-btn" @click="prevDate">
              <el-icon><ArrowLeft /></el-icon>
            </button>
            <span class="current-date">{{ formatCurrentDate }}</span>
            <button class="nav-btn" @click="nextDate">
              <el-icon><ArrowRight /></el-icon>
            </button>
          </div>
          
          <button 
            class="cm-btn cm-btn-primary" 
            @click="generateReport"
            :disabled="generating"
          >
            <el-icon v-if="generating" class="is-loading"><Loading /></el-icon>
            <el-icon v-else><MagicStick /></el-icon>
            <span>{{ generating ? $t('dailyReport.generating') : $t('dailyReport.generate') }}</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Stats Cards -->
    <div class="stats-section">
      <div class="stats-grid">
        <div class="stat-card-v2" v-for="stat in statsData" :key="stat.key">
          <div class="stat-icon" :class="stat.color">
            <el-icon><component :is="stat.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ stat.value }}</span>
            <span class="stat-label">{{ stat.label }}</span>
          </div>
          <div class="stat-trend" :class="stat.trend > 0 ? 'up' : 'down'" v-if="stat.trend !== undefined">
            <el-icon><component :is="stat.trend > 0 ? 'ArrowUp' : 'ArrowDown'" /></el-icon>
            <span>{{ Math.abs(stat.trend) }}%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="report-content">
      <!-- Left: Report List -->
      <div class="report-sidebar">
        <div class="sidebar-header">
          <h3>{{ $t('dailyReport.history') }}</h3>
          <span class="count">{{ reports.length }} {{ $t('dailyReport.reports') }}</span>
        </div>
        
        <div class="report-timeline">
          <div 
            v-for="report in reports" 
            :key="report.id"
            class="timeline-item"
            :class="{ active: selectedReport?.id === report.id }"
            @click="selectReport(report)"
          >
            <div class="timeline-dot" :class="{ 'has-content': report.summary }"></div>
            <div class="timeline-content">
              <div class="timeline-date">{{ formatDate(report.report_date) }}</div>
              <div class="timeline-summary">{{ truncate(report.summary, 60) }}</div>
              <div class="timeline-tags" v-if="report.stats">
                <span class="tag" v-if="report.stats.new_memories">
                  <el-icon><Document /></el-icon>
                  {{ report.stats.new_memories }}
                </span>
                <span class="tag" v-if="report.stats.new_entities">
                  <el-icon><Connection /></el-icon>
                  {{ report.stats.new_entities }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Report Detail -->
      <div class="report-detail" v-if="selectedReport">
        <div class="detail-header">
          <div class="detail-date">
            <span class="day">{{ getDay(selectedReport.report_date) }}</span>
            <span class="month">{{ getMonth(selectedReport.report_date) }}</span>
          </div>
          <div class="detail-actions">
            <button class="action-btn" @click="exportReport" title="Export">
              <el-icon><Download /></el-icon>
            </button>
            <button class="action-btn" @click="shareReport" title="Share">
              <el-icon><Share /></el-icon>
            </button>
          </div>
        </div>

        <div class="detail-content">
          <!-- Summary Section -->
          <section class="content-section">
            <div class="section-header">
              <div class="section-icon gradient-bg">
                <el-icon><DocumentChecked /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.summary') }}</h2>
            </div>
            <p class="summary-text">{{ selectedReport.summary }}</p>
          </section>

          <!-- Highlights Section -->
          <section class="content-section" v-if="selectedReport.highlights?.length">
            <div class="section-header">
              <div class="section-icon gradient-bg purple">
                <el-icon><Star /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.highlights') }}</h2>
            </div>
            <div class="highlight-cards">
              <div 
                v-for="(highlight, i) in selectedReport.highlights" 
                :key="i"
                class="highlight-card"
              >
                <div class="highlight-number">{{ i + 1 }}</div>
                <p>{{ highlight }}</p>
              </div>
            </div>
          </section>

          <!-- Knowledge Section -->
          <section class="content-section" v-if="selectedReport.knowledge_gained?.length">
            <div class="section-header">
              <div class="section-icon gradient-bg green">
                <el-icon><Lightbulb /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.knowledgeGained') }}</h2>
            </div>
            <ul class="knowledge-list">
              <li v-for="(item, i) in selectedReport.knowledge_gained" :key="i">
                <el-icon><Check /></el-icon>
                <span>{{ item }}</span>
              </li>
            </ul>
          </section>

          <!-- Tasks Section -->
          <section class="content-section" v-if="selectedReport.pending_tasks?.length">
            <div class="section-header">
              <div class="section-icon gradient-bg orange">
                <el-icon><Timer /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.pendingTasks') }}</h2>
            </div>
            <div class="task-list">
              <div 
                v-for="(task, i) in selectedReport.pending_tasks" 
                :key="i"
                class="task-item"
              >
                <el-checkbox v-model="task.completed">
                  <span :class="{ 'task-completed': task.completed }">{{ task.text || task }}</span>
                </el-checkbox>
              </div>
            </div>
          </section>

          <!-- Suggestions Section -->
          <section class="content-section" v-if="selectedReport.tomorrow_suggestions?.length">
            <div class="section-header">
              <div class="section-icon gradient-bg blue">
                <el-icon><Compass /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.tomorrowSuggestions') }}</h2>
            </div>
            <div class="suggestion-cards">
              <div 
                v-for="(suggestion, i) in selectedReport.tomorrow_suggestions" 
                :key="i"
                class="suggestion-card"
              >
                <div class="suggestion-icon">💡</div>
                <p>{{ suggestion }}</p>
              </div>
            </div>
          </section>

          <!-- Stats Section -->
          <section class="content-section" v-if="selectedReport.stats">
            <div class="section-header">
              <div class="section-icon gradient-bg">
                <el-icon><DataAnalysis /></el-icon>
              </div>
              <h2>{{ $t('dailyReport.statistics') }}</h2>
            </div>
            <div class="stats-detail">
              <div class="stat-item" v-for="(value, key) in selectedReport.stats" :key="key">
                <div class="stat-item-value">{{ value }}</div>
                <div class="stat-item-label">{{ statLabel(key) }}</div>
              </div>
            </div>
          </section>
        </div>
      </div>

      <!-- Empty State -->
      <div class="report-detail empty" v-else>
        <div class="empty-illustration">
          <svg viewBox="0 0 200 200" width="120" height="120">
            <rect x="40" y="60" width="120" height="100" rx="8" fill="none" stroke="var(--cm-primary)" stroke-width="2" opacity="0.3"/>
            <rect x="55" y="80" width="90" height="6" rx="3" fill="var(--cm-primary)" opacity="0.2"/>
            <rect x="55" y="95" width="70" height="6" rx="3" fill="var(--cm-primary)" opacity="0.2"/>
            <rect x="55" y="110" width="80" height="6" rx="3" fill="var(--cm-primary)" opacity="0.2"/>
            <circle cx="100" cy="40" r="20" fill="none" stroke="var(--cm-primary)" stroke-width="2" opacity="0.3"/>
            <path d="M90 35 L100 45 L115 30" fill="none" stroke="var(--cm-primary)" stroke-width="2" opacity="0.3"/>
          </svg>
        </div>
        <h3>{{ $t('dailyReport.selectReport') }}</h3>
        <p>{{ $t('dailyReport.selectHint') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArrowLeft, ArrowRight, MagicStick, Loading,
  Document, Connection, Download, Share,
  DocumentChecked, Star, Sunny, Timer, Compass,
  DataAnalysis, Check, ArrowUp, ArrowDown
} from '@element-plus/icons-vue'

const { t } = useI18n()

// State
const currentDate = ref(new Date())
const selectedReport = ref<any>(null)
const generating = ref(false)

// Mock data - replace with actual store data
const reports = ref([
  {
    id: 1,
    report_date: '2026-04-23',
    summary: '今天记录了 15 条新记忆，创建了 3 个知识实体，更新了 Wiki 文档。整体知识管理效率较高。',
    highlights: [
      '完成了项目架构设计文档',
      '整理了 Go 后端迁移方案',
      '优化了记忆衰减算法',
    ],
    knowledge_gained: [
      '掌握了 Go 的交叉编译技巧',
      '理解了 Cloudflare Workers 的免费额度限制',
      '学习了现代化 UI 设计原则',
    ],
    pending_tasks: [
      { text: '完成前端 UI 重构', completed: false },
      { text: '测试全平台编译', completed: false },
      { text: '更新文档', completed: true },
    ],
    tomorrow_suggestions: [
      '继续完善知识库功能',
      '优化日报生成算法',
      '添加更多数据可视化',
    ],
    stats: {
      new_memories: 15,
      new_entities: 3,
      updated_wiki: 2,
      active_hours: 6,
    },
  },
  {
    id: 2,
    report_date: '2026-04-22',
    summary: '昨天主要进行了代码重构工作，优化了数据库查询性能。',
    highlights: ['重构了认证模块', '优化了查询性能'],
    stats: {
      new_memories: 8,
      new_entities: 1,
      updated_wiki: 0,
      active_hours: 4,
    },
  },
])

const statsData = computed(() => [
  {
    key: 'memories',
    icon: 'Document',
    value: reports.value[0]?.stats?.new_memories || 0,
    label: t('dailyReport.newMemories'),
    color: 'blue',
    trend: 12,
  },
  {
    key: 'entities',
    icon: 'Connection',
    value: reports.value[0]?.stats?.new_entities || 0,
    label: t('dailyReport.newEntities'),
    color: 'purple',
    trend: 5,
  },
  {
    key: 'wiki',
    icon: 'DocumentChecked',
    value: reports.value[0]?.stats?.updated_wiki || 0,
    label: t('dailyReport.updatedWiki'),
    color: 'green',
    trend: -3,
  },
  {
    key: 'hours',
    icon: 'Timer',
    value: reports.value[0]?.stats?.active_hours || 0,
    label: t('dailyReport.activeHours'),
    color: 'orange',
    trend: 8,
  },
])

const formatCurrentDate = computed(() => {
  return currentDate.value.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  })
})

// Methods
function prevDate() {
  currentDate.value = new Date(currentDate.value.setDate(currentDate.value.getDate() - 1))
}

function nextDate() {
  currentDate.value = new Date(currentDate.value.setDate(currentDate.value.getDate() + 1))
}

function selectReport(report: any) {
  selectedReport.value = report
}

function generateReport() {
  generating.value = true
  // TODO: Call API to generate report
  setTimeout(() => {
    generating.value = false
  }, 2000)
}

function exportReport() {
  // TODO: Export report
}

function shareReport() {
  // TODO: Share report
}

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
  })
}

function getDay(date: string) {
  return new Date(date).getDate()
}

function getMonth(date: string) {
  return new Date(date).toLocaleDateString('zh-CN', { month: 'short' })
}

function truncate(text: string, length: number) {
  if (!text) return ''
  return text.length > length ? text.slice(0, length) + '...' : text
}

function statLabel(key: string | number) {
  const labels: Record<string, string> = {
    new_memories: t('dailyReport.memories'),
    new_entities: t('dailyReport.entities'),
    updated_wiki: t('dailyReport.wiki'),
    active_hours: t('dailyReport.hours'),
  }
  return labels[String(key)] || String(key)
}

onMounted(() => {
  if (reports.value.length > 0) {
    selectedReport.value = reports.value[0]
  }
})
</script>

<style scoped>
.daily-report-v2 {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Header */
.report-header {
  background: var(--cm-bg-primary);
  border-bottom: 1px solid var(--cm-border);
  padding: var(--cm-space-6);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
}

.title-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--cm-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.gradient-bg {
  background: var(--cm-primary-gradient);
}

.gradient-bg.purple {
  background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%);
}

.gradient-bg.green {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
}

.gradient-bg.orange {
  background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
}

.gradient-bg.blue {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
}

.header-title h1 {
  font-size: 24px;
  font-weight: 700;
}

.subtitle {
  font-size: 14px;
  color: var(--cm-text-secondary);
  margin-top: 2px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--cm-space-4);
}

.date-navigator {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
  padding: var(--cm-space-1);
}

.nav-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--cm-text-secondary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.nav-btn:hover {
  background: var(--cm-bg-primary);
  color: var(--cm-text-primary);
}

.current-date {
  font-size: 14px;
  font-weight: 500;
  color: var(--cm-text-primary);
  padding: 0 var(--cm-space-3);
  min-width: 140px;
  text-align: center;
}

/* Stats Section */
.stats-section {
  padding: var(--cm-space-6);
  background: var(--cm-bg-secondary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--cm-space-4);
}

.stat-card-v2 {
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  padding: var(--cm-space-5);
  display: flex;
  align-items: center;
  gap: var(--cm-space-4);
  transition: all var(--cm-transition-normal);
}

.stat-card-v2:hover {
  box-shadow: var(--cm-shadow-md);
  transform: translateY(-2px);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--cm-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.blue { background: linear-gradient(135deg, #3b82f6, #60a5fa); }
.stat-icon.purple { background: linear-gradient(135deg, #8b5cf6, #a78bfa); }
.stat-icon.green { background: linear-gradient(135deg, #10b981, #34d399); }
.stat-icon.orange { background: linear-gradient(135deg, #f59e0b, #fbbf24); }

.stat-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--cm-text-primary);
}

.stat-label {
  font-size: 13px;
  color: var(--cm-text-secondary);
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: var(--cm-radius-full);
}

.stat-trend.up {
  color: var(--cm-success);
  background: rgba(16, 185, 129, 0.1);
}

.stat-trend.down {
  color: var(--cm-error);
  background: rgba(239, 68, 68, 0.1);
}

/* Report Content */
.report-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Sidebar */
.report-sidebar {
  width: 320px;
  background: var(--cm-bg-primary);
  border-right: 1px solid var(--cm-border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--cm-space-4) var(--cm-space-5);
  border-bottom: 1px solid var(--cm-border);
}

.sidebar-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.count {
  font-size: 13px;
  color: var(--cm-text-tertiary);
}

.report-timeline {
  flex: 1;
  overflow: auto;
  padding: var(--cm-space-4);
}

.timeline-item {
  display: flex;
  gap: var(--cm-space-3);
  padding: var(--cm-space-3);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  transition: all var(--cm-transition-fast);
  margin-bottom: var(--cm-space-2);
}

.timeline-item:hover {
  background: var(--cm-bg-secondary);
}

.timeline-item.active {
  background: var(--cm-bg-secondary);
  border-left: 3px solid var(--cm-primary);
}

.timeline-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--cm-border);
  margin-top: 6px;
  flex-shrink: 0;
}

.timeline-dot.has-content {
  background: var(--cm-primary);
}

.timeline-content {
  flex: 1;
  min-width: 0;
}

.timeline-date {
  font-size: 13px;
  font-weight: 600;
  color: var(--cm-text-primary);
  margin-bottom: 2px;
}

.timeline-summary {
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.timeline-tags {
  display: flex;
  gap: var(--cm-space-2);
  margin-top: var(--cm-space-2);
}

.tag {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  padding: 2px 8px;
  background: var(--cm-bg-tertiary);
  border-radius: var(--cm-radius-full);
  color: var(--cm-text-secondary);
}

/* Report Detail */
.report-detail {
  flex: 1;
  overflow: auto;
  padding: var(--cm-space-6);
}

.report-detail.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--cm-text-tertiary);
}

.empty-illustration {
  margin-bottom: var(--cm-space-6);
}

.empty h3 {
  font-size: 18px;
  color: var(--cm-text-secondary);
  margin-bottom: var(--cm-space-2);
}

.empty p {
  font-size: 14px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--cm-space-6);
}

.detail-date {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: var(--cm-primary-gradient);
  color: white;
  padding: var(--cm-space-4) var(--cm-space-6);
  border-radius: var(--cm-radius-lg);
}

.detail-date .day {
  font-size: 36px;
  font-weight: 700;
  line-height: 1;
}

.detail-date .month {
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.detail-actions {
  display: flex;
  gap: var(--cm-space-2);
}

.action-btn {
  width: 40px;
  height: 40px;
  border: 1px solid var(--cm-border);
  background: var(--cm-bg-primary);
  color: var(--cm-text-secondary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.action-btn:hover {
  border-color: var(--cm-primary);
  color: var(--cm-primary);
}

/* Content Sections */
.content-section {
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  padding: var(--cm-space-6);
  margin-bottom: var(--cm-space-4);
}

.section-header {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  margin-bottom: var(--cm-space-4);
}

.section-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--cm-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.section-header h2 {
  font-size: 18px;
  font-weight: 600;
}

.summary-text {
  font-size: 16px;
  line-height: 1.7;
  color: var(--cm-text-primary);
}

/* Highlight Cards */
.highlight-cards {
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-3);
}

.highlight-card {
  display: flex;
  align-items: flex-start;
  gap: var(--cm-space-3);
  padding: var(--cm-space-4);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
  border-left: 3px solid var(--cm-primary);
}

.highlight-number {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--cm-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  flex-shrink: 0;
}

.highlight-card p {
  font-size: 15px;
  line-height: 1.5;
  color: var(--cm-text-primary);
}

/* Knowledge List */
.knowledge-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-3);
}

.knowledge-list li {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  padding: var(--cm-space-3);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
  font-size: 15px;
}

.knowledge-list li .el-icon {
  color: var(--cm-success);
}

/* Task List */
.task-list {
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-2);
}

.task-item {
  padding: var(--cm-space-3);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
}

.task-completed {
  text-decoration: line-through;
  color: var(--cm-text-tertiary);
}

/* Suggestion Cards */
.suggestion-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--cm-space-3);
}

.suggestion-card {
  display: flex;
  align-items: flex-start;
  gap: var(--cm-space-3);
  padding: var(--cm-space-4);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
}

.suggestion-icon {
  font-size: 24px;
}

.suggestion-card p {
  font-size: 14px;
  line-height: 1.5;
  color: var(--cm-text-primary);
}

/* Stats Detail */
.stats-detail {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--cm-space-4);
}

.stat-item {
  text-align: center;
  padding: var(--cm-space-4);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
}

.stat-item-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--cm-primary);
  display: block;
}

.stat-item-label {
  font-size: 13px;
  color: var(--cm-text-secondary);
  margin-top: 4px;
}

/* Responsive */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .report-content {
    flex-direction: column;
  }
  
  .report-sidebar {
    width: 100%;
    max-height: 200px;
  }
  
  .stats-detail {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: var(--cm-space-4);
    align-items: stretch;
  }
  
  .header-actions {
    flex-wrap: wrap;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .suggestion-cards {
    grid-template-columns: 1fr;
  }
}
</style>
