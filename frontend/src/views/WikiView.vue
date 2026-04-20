<template>
  <div class="wiki-page">
    <div class="page-header">
      <div class="header-left">
        <h1>📖 {{ $t('wiki.title') }}</h1>
        <div class="header-stats">
          <span class="stat-badge completed">{{ stats.completed }} {{ $t('wiki.completed') }}</span>
          <span class="stat-badge in-progress">{{ stats.in_progress }} {{ $t('wiki.inProgress') }}</span>
          <span class="stat-badge draft">{{ stats.draft }} {{ $t('wiki.draft') }}</span>
          <span class="stat-badge ai" v-if="stats.ai_generated">🤖 {{ stats.ai_generated }} AI</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openExtractDialog" v-if="llmAvailable">
          <el-icon><MagicStick /></el-icon> AI {{ $t('wiki.extractFromConversation') }}
        </el-button>
        <el-button type="primary" @click="openNewPage">
          <el-icon><Plus /></el-icon> {{ $t('wiki.addPage') }}
        </el-button>
      </div>
    </div>

    <div class="wiki-layout">
      <!-- 侧边栏：分类/状态/导航 -->
      <div class="wiki-sidebar">
        <el-input v-model="searchQuery" :placeholder="$t('wiki.searchPlaceholder')" clearable @keyup.enter="handleSearch" class="search-input" size="small">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <div class="sidebar-section">
          <div class="sidebar-title">{{ $t('wiki.status') }}</div>
          <div class="status-list">
            <div class="status-item" :class="{ active: selectedStatus === '' }" @click="selectedStatus = ''; loadPages()">
              <span class="status-dot all"></span> {{ $t('wiki.allPages') }}
            </div>
            <div class="status-item" :class="{ active: selectedStatus === 'completed' }" @click="selectedStatus = 'completed'; loadPages()">
              <span class="status-dot completed"></span> {{ $t('wiki.completed') }}
            </div>
            <div class="status-item" :class="{ active: selectedStatus === 'in_progress' }" @click="selectedStatus = 'in_progress'; loadPages()">
              <span class="status-dot in-progress"></span> {{ $t('wiki.inProgress') }}
            </div>
            <div class="status-item" :class="{ active: selectedStatus === 'draft' }" @click="selectedStatus = 'draft'; loadPages()">
              <span class="status-dot draft"></span> {{ $t('wiki.draft') }}
            </div>
          </div>
        </div>

        <div class="sidebar-section">
          <div class="sidebar-title">{{ $t('wiki.categories') }}</div>
          <div class="category-list">
            <div class="category-item" :class="{ active: selectedCategory === '' }" @click="selectedCategory = ''; loadPages()">
              {{ $t('wiki.allCategories') }}
            </div>
            <div class="category-item" :class="{ active: selectedCategory === cat }" v-for="cat in categories" :key="cat" @click="selectedCategory = cat; loadPages()">
              {{ cat }}
            </div>
          </div>
        </div>

        <div class="sidebar-section" v-if="searchResults.length">
          <div class="sidebar-title">{{ $t('wiki.searchResults') }} ({{ searchResults.length }})</div>
          <div class="category-list">
            <div class="category-item" v-for="p in searchResults" :key="'s'+p.id" @click="viewPage(p.id)">
              {{ p.title }}
            </div>
          </div>
        </div>

        <div class="sidebar-section">
          <div class="sidebar-title">{{ $t('wiki.pinned') }}</div>
          <div class="category-list">
            <div class="category-item pinned" v-for="p in pinnedPages" :key="'pin'+p.id" @click="viewPage(p.id)">
              📌 {{ p.title }}
            </div>
            <div v-if="!pinnedPages.length" class="empty-hint-small">{{ $t('common.noData') }}</div>
          </div>
        </div>
      </div>

      <!-- 主内容区 -->
      <div class="wiki-main">
        <!-- 页面列表模式（未选中页面时） -->
        <div v-if="!viewingPage" class="page-grid">
          <div class="page-card" v-for="p in pages" :key="p.id" @click="viewPage(p.id)">
            <div class="card-top">
              <span class="category-tag" v-if="p.category">{{ p.category }}</span>
              <span class="status-badge" :class="p.status">{{ getStatusLabel(p.status) }}</span>
              <span class="pin-badge" v-if="p.is_pinned">◆ {{ $t('wiki.pinned') }}</span>
              <span class="ai-badge" v-if="p.ai_generated">🤖 {{ Math.round(p.ai_confidence * 100) }}%</span>
            </div>
            <div class="card-title">{{ p.title }}</div>
            <div class="card-summary" v-if="p.summary">{{ p.summary }}</div>
            <div class="card-preview" v-else-if="p.content">{{ getPreview(p.content) }}</div>
            <div class="card-tags" v-if="p.tags && p.tags.length">
              <span class="tag" v-for="tag in p.tags.slice(0, 3)" :key="tag">{{ tag }}</span>
              <span class="tag" v-if="p.tags.length > 3">+{{ p.tags.length - 3 }}</span>
            </div>
            <div class="card-footer">
              <div class="card-meta">{{ formatTime(p.updated_at) }}</div>
              <div class="card-actions" @click.stop>
                <el-button size="small" text type="primary" @click="markComplete(p)" v-if="p.status !== 'completed'">{{ $t('wiki.markComplete') }}</el-button>
                <el-button size="small" text type="warning" @click="refinePage(p)" v-if="llmAvailable">{{ $t('wiki.refine') }}</el-button>
              </div>
            </div>
          </div>
          <div v-if="!pages.length" class="empty-state">
            <div class="empty-icon">◇</div>
            <p>{{ $t('common.noData') }}</p>
            <el-button type="primary" @click="openNewPage">{{ $t('wiki.addPage') }}</el-button>
          </div>
        </div>

        <!-- 页面查看模式 -->
        <div v-else class="page-view">
          <div class="view-header">
            <el-button text @click="viewingPage = null" class="back-btn">
              <el-icon><ArrowLeft /></el-icon> {{ $t('common.back') }}
            </el-button>
            <div class="view-actions">
              <el-button text type="primary" @click="editCurrentPage">{{ $t('common.edit') }}</el-button>
              <el-button text type="warning" @click="refinePage(viewingPage)" v-if="llmAvailable">{{ $t('wiki.refine') }}</el-button>
              <el-button text type="success" @click="markComplete(viewingPage)" v-if="viewingPage.status !== 'completed'">{{ $t('wiki.markComplete') }}</el-button>
              <el-button text type="danger" @click="deleteCurrentPage">{{ $t('common.delete') }}</el-button>
            </div>
          </div>
          <div class="view-meta">
            <span class="category-tag" v-if="viewingPage.category">{{ viewingPage.category }}</span>
            <span class="status-badge" :class="viewingPage.status">{{ getStatusLabel(viewingPage.status) }}</span>
            <span class="ai-badge" v-if="viewingPage.ai_generated">🤖 {{ Math.round(viewingPage.ai_confidence * 100) }}%</span>
            <span class="view-date">{{ formatTime(viewingPage.updated_at) }}</span>
          </div>
          <h1 class="view-title">{{ viewingPage.title }}</h1>
          <div class="view-tags" v-if="viewingPage.tags && viewingPage.tags.length">
            <span class="tag" v-for="tag in viewingPage.tags" :key="tag">{{ tag }}</span>
          </div>

          <!-- AI 生成的摘要 -->
          <div class="ai-summary" v-if="viewingPage.summary">
            <div class="ai-summary-header">🤖 {{ $t('wiki.aiSummary') }}</div>
            <div class="ai-summary-content">{{ viewingPage.summary }}</div>
          </div>

          <!-- 关键决策 -->
          <div class="key-section" v-if="viewingPage.key_decisions && viewingPage.key_decisions.length">
            <div class="key-section-header">🎯 {{ $t('wiki.keyDecisions') }}</div>
            <ul class="key-list">
              <li v-for="(decision, idx) in viewingPage.key_decisions" :key="idx">{{ decision }}</li>
            </ul>
          </div>

          <!-- 待办事项 -->
          <div class="key-section" v-if="viewingPage.action_items && viewingPage.action_items.length">
            <div class="key-section-header">📋 {{ $t('wiki.actionItems') }}</div>
            <ul class="action-list">
              <li v-for="(item, idx) in viewingPage.action_items" :key="idx" class="action-item">
                <el-checkbox>{{ item }}</el-checkbox>
              </li>
            </ul>
          </div>

          <div class="markdown-body" v-html="renderedContent"></div>
        </div>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <el-dialog v-model="showEditor" :title="isEditing ? $t('wiki.editPage') : $t('wiki.newPage')" width="800px" top="5vh" :close-on-click-modal="false">
      <el-form label-position="top">
        <div class="editor-row">
          <el-form-item :label="$t('wiki.titleField')" class="editor-title">
            <el-input v-model="pageForm.title" :placeholder="$t('wiki.titlePlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('wiki.categoryField')" class="editor-category">
            <el-input v-model="pageForm.category" :placeholder="$t('wiki.categoryPlaceholder')" />
          </el-form-item>
        </div>
        <el-form-item :label="$t('wiki.content')">
          <el-input v-model="pageForm.content" type="textarea" :rows="16" :placeholder="$t('wiki.contentPlaceholder')" class="wiki-editor" />
        </el-form-item>
        <div class="editor-row">
          <el-form-item :label="$t('wiki.status')" class="editor-status">
            <el-select v-model="pageForm.status" style="width: 100%">
              <el-option :label="$t('wiki.draft')" value="draft" />
              <el-option :label="$t('wiki.inProgress')" value="in_progress" />
              <el-option :label="$t('wiki.completed')" value="completed" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('wiki.tags')" class="editor-tags">
            <el-input v-model="pageForm.tagsStr" :placeholder="$t('wiki.tagsPlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('wiki.parentPage')" class="editor-parent">
            <el-select v-model="pageForm.parent_id" :placeholder="$t('wiki.noParent')" clearable style="width: 100%">
              <el-option :label="$t('wiki.noParent')" :value="null" />
              <el-option v-for="p in allPages" :key="p.id" :label="p.title" :value="p.id" v-show="p.id !== currentPageId" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showEditor = false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" @click="savePage" :loading="saving">{{ $t('common.save') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- AI 提取对话框 -->
    <el-dialog v-model="showExtractDialog" :title="$t('wiki.extractFromConversation')" width="700px" top="10vh" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item :label="$t('wiki.conversationContent')">
          <el-input v-model="extractForm.conversation" type="textarea" :rows="12" :placeholder="$t('wiki.conversationPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('wiki.isComplete')">
          <el-switch v-model="extractForm.is_complete" :active-text="$t('wiki.complete')" :inactive-text="$t('wiki.inProgress')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showExtractDialog = false">{{ $t('common.cancel') }}</el-button>
          <el-button type="primary" @click="extractKnowledge" :loading="extracting">
            <el-icon><MagicStick /></el-icon> {{ $t('wiki.extract') }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, ArrowLeft, MagicStick } from '@element-plus/icons-vue'
import { marked } from 'marked'
import wikiApi from '../api/wiki'

const { t } = useI18n()
const route = useRoute()
const pages = ref<any[]>([])
const allPages = ref<any[]>([])
const categories = ref<string[]>([])
const searchResults = ref<any[]>([])
const searchQuery = ref('')
const selectedCategory = ref('')
const selectedStatus = ref('')
const showEditor = ref(false)
const showExtractDialog = ref(false)
const isEditing = ref(false)
const currentPageId = ref<number | null>(null)
const viewingPage = ref<any>(null)
const saving = ref(false)
const extracting = ref(false)
const llmAvailable = ref(false)

const stats = ref({
  total: 0,
  completed: 0,
  in_progress: 0,
  draft: 0,
  ai_generated: 0,
})

const pageForm = ref({
  title: '', content: '', category: '', tagsStr: '', parent_id: null as number | null, is_pinned: false, status: 'draft',
})

const extractForm = ref({
  conversation: '',
  is_complete: true,
})

const pinnedPages = computed(() => pages.value.filter(p => p.is_pinned))

const renderedContent = computed(() => {
  if (!viewingPage.value?.content) return ''
  return marked(viewingPage.value.content, { breaks: true, gfm: true })
})

onMounted(() => { loadPages(); loadCategories(); loadStats(); checkLLMAvailability() })

watch(() => route.query.tab, (tab) => {
  if (tab === 'categories') {
    selectedCategory.value = ''
  }
})

async function loadPages() {
  try {
    const params: any = {}
    if (selectedCategory.value) params.category = selectedCategory.value
    if (selectedStatus.value) params.status = selectedStatus.value
    const { data } = await wikiApi.listPages(params)
    pages.value = data || []
  } catch { pages.value = [] }
}

async function loadCategories() {
  try {
    const { data } = await wikiApi.getCategories()
    categories.value = data || []
  } catch { categories.value = [] }
}

async function loadAllPages() {
  try {
    const { data } = await wikiApi.listPages()
    allPages.value = data || []
  } catch {}
}

async function loadStats() {
  try {
    const { data } = await wikiApi.getStats()
    if (data) stats.value = data
  } catch {}
}

async function checkLLMAvailability() {
  try {
    const { data } = await wikiApi.getConfig()
    llmAvailable.value = data?.llm_available || false
  } catch {}
}

async function handleSearch() {
  if (!searchQuery.value) { searchResults.value = []; return }
  try {
    const { data } = await wikiApi.search(searchQuery.value)
    searchResults.value = data || []
  } catch { searchResults.value = [] }
}

function openNewPage() {
  isEditing.value = false
  currentPageId.value = null
  pageForm.value = { title: '', content: '', category: '', tagsStr: '', parent_id: null, is_pinned: false, status: 'draft' }
  loadAllPages()
  showEditor.value = true
}

function openExtractDialog() {
  extractForm.value = { conversation: '', is_complete: true }
  showExtractDialog.value = true
}

async function viewPage(id: number) {
  try {
    const { data } = await wikiApi.getPage(id)
    viewingPage.value = data
  } catch (e: any) {
    ElMessage.error(t('common.failed'))
  }
}

function editCurrentPage() {
  if (!viewingPage.value) return
  const p = viewingPage.value
  isEditing.value = true
  currentPageId.value = p.id
  pageForm.value = {
    title: p.title, content: p.content || '', category: p.category || '',
    tagsStr: (p.tags || []).join(', '), parent_id: p.parent_id, is_pinned: p.is_pinned,
    status: p.status || 'draft',
  }
  loadAllPages()
  showEditor.value = true
}

async function savePage() {
  if (!pageForm.value.title) { ElMessage.warning(t('wiki.fillTitle')); return }
  saving.value = true
  try {
    const payload: any = {
      title: pageForm.value.title, content: pageForm.value.content,
      category: pageForm.value.category || null,
      tags: pageForm.value.tagsStr ? pageForm.value.tagsStr.split(',').map((s: string) => s.trim()).filter(Boolean) : [],
      parent_id: pageForm.value.parent_id, is_pinned: pageForm.value.is_pinned,
      status: pageForm.value.status,
    }
    if (isEditing.value && currentPageId.value) {
      await wikiApi.updatePage(currentPageId.value, payload)
    } else {
      await wikiApi.createPage(payload)
    }
    ElMessage.success(t('common.success'))
    showEditor.value = false
    searchResults.value = []
    await Promise.all([loadPages(), loadCategories(), loadStats()])
    if (viewingPage.value && currentPageId.value === viewingPage.value.id) {
      await viewPage(viewingPage.value.id)
    }
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteCurrentPage() {
  const id = viewingPage.value?.id
  if (!id) return
  try {
    await ElMessageBox.confirm(t('wiki.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await wikiApi.deletePage(id)
    ElMessage.success(t('wiki.deleted'))
    viewingPage.value = null
    searchResults.value = []
    await Promise.all([loadPages(), loadCategories(), loadStats()])
  } catch {}
}

async function markComplete(page: any) {
  try {
    await wikiApi.markComplete(page.id)
    ElMessage.success(t('wiki.markedComplete'))
    await Promise.all([loadPages(), loadStats()])
    if (viewingPage.value?.id === page.id) {
      await viewPage(page.id)
    }
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
}

async function refinePage(page: any) {
  try {
    await wikiApi.refinePage(page.id, '')
    ElMessage.success(t('wiki.pageRefined'))
    await Promise.all([loadPages(), loadStats()])
    if (viewingPage.value?.id === page.id) {
      await viewPage(page.id)
    }
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
}

async function extractKnowledge() {
  if (!extractForm.value.conversation.trim()) {
    ElMessage.warning(t('wiki.fillConversation'))
    return
  }
  extracting.value = true
  try {
    await wikiApi.extractFromConversation(extractForm.value.conversation, extractForm.value.is_complete)
    ElMessage.success(t('wiki.extractSuccess'))
    showExtractDialog.value = false
    await Promise.all([loadPages(), loadStats()])
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { extracting.value = false }
}

function getStatusLabel(status: string) {
  const labels: Record<string, string> = {
    draft: t('wiki.draft'),
    in_progress: t('wiki.inProgress'),
    completed: t('wiki.completed'),
  }
  return labels[status] || status
}

function getPreview(content: string) {
  if (!content) return ''
  const text = content.replace(/[#*`\[\]()>_-]/g, '').slice(0, 120)
  return text.length < content.replace(/[#*`\[\]()>_-]/g, '').length ? text + '...' : text
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style scoped>
.wiki-page { padding: 28px; max-width: 1400px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.header-left { flex: 1; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0 0 8px; }
.header-stats { display: flex; gap: 8px; flex-wrap: wrap; }
.stat-badge { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.stat-badge.completed { background: rgba(16, 185, 129, 0.15); color: #10B981; }
.stat-badge.in-progress { background: rgba(245, 158, 11, 0.15); color: #F59E0B; }
.stat-badge.draft { background: rgba(107, 114, 128, 0.15); color: #6B7280; }
.stat-badge.ai { background: rgba(139, 92, 246, 0.15); color: #8B5CF6; }
.header-actions { display: flex; gap: 8px; }

/* Layout */
.wiki-layout { display: flex; gap: 20px; }
.wiki-sidebar { width: 220px; flex-shrink: 0; }
.wiki-main { flex: 1; min-width: 0; }

/* Sidebar */
.search-input { margin-bottom: 16px; }
.sidebar-section { margin-bottom: 16px; }
.sidebar-title { font-size: 11px; font-weight: 600; color: var(--cm-text-placeholder); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 8px; }
.category-list, .status-list { display: flex; flex-direction: column; gap: 2px; }
.category-item, .status-item { padding: 6px 10px; border-radius: 6px; font-size: 13px; color: var(--cm-text-muted); cursor: pointer; transition: all 0.15s; display: flex; align-items: center; gap: 8px; }
.category-item:hover, .status-item:hover { background: rgba(16,185,129,0.08); color: var(--cm-text); }
.category-item.active, .status-item.active { background: rgba(16,185,129,0.15); color: #10B981; font-weight: 600; }
.category-item.pinned { color: #10B981; }
.empty-hint-small { font-size: 12px; color: var(--cm-text-placeholder); padding: 4px 10px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.status-dot.all { background: var(--cm-text-muted); }
.status-dot.completed { background: #10B981; }
.status-dot.in-progress { background: #F59E0B; }
.status-dot.draft { background: #6B7280; }

/* Page Grid */
.page-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.page-card {
  background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px;
  cursor: pointer; transition: all 0.2s; display: flex; flex-direction: column;
}
.page-card:hover { border-color: rgba(16,185,129,0.4); transform: translateY(-2px); box-shadow: var(--cm-shadow); }
.card-top { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; flex-wrap: wrap; }
.category-tag { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; background: rgba(6,182,212,0.15); color: #06b6d4; }
.status-badge { padding: 2px 8px; border-radius: 10px; font-size: 10px; font-weight: 600; }
.status-badge.completed { background: rgba(16, 185, 129, 0.15); color: #10B981; }
.status-badge.in_progress { background: rgba(245, 158, 11, 0.15); color: #F59E0B; }
.status-badge.draft { background: rgba(107, 114, 128, 0.15); color: #6B7280; }
.pin-badge { font-size: 11px; color: #10B981; font-weight: 600; }
.ai-badge { padding: 2px 6px; border-radius: 8px; font-size: 10px; background: rgba(139, 92, 246, 0.15); color: #8B5CF6; }
.card-title { font-size: 16px; font-weight: 600; color: var(--cm-text); margin-bottom: 6px; }
.card-summary, .card-preview { font-size: 12px; color: var(--cm-text-muted); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8px; }
.card-summary { color: var(--cm-text-secondary); font-style: italic; }
.card-tags { display: flex; gap: 4px; flex-wrap: wrap; margin-bottom: 8px; }
.tag { padding: 1px 8px; background: var(--cm-border); border-radius: 4px; font-size: 11px; color: var(--cm-text-muted); }
.card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: auto; padding-top: 8px; border-top: 1px solid var(--cm-border); }
.card-meta { font-size: 11px; color: var(--cm-text-placeholder); }
.card-actions { display: flex; gap: 4px; }

/* Page View */
.page-view { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 24px 32px; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.back-btn { color: var(--cm-text-muted); }
.view-actions { display: flex; gap: 4px; }
.view-meta { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; flex-wrap: wrap; }
.view-date { font-size: 12px; color: var(--cm-text-placeholder); }
.view-title { font-size: 28px; font-weight: 700; color: var(--cm-text); margin: 0 0 12px; }
.view-tags { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 20px; }

/* AI Summary */
.ai-summary { background: rgba(139, 92, 246, 0.08); border: 1px solid rgba(139, 92, 246, 0.2); border-radius: 8px; padding: 16px; margin-bottom: 20px; }
.ai-summary-header { font-size: 13px; font-weight: 600; color: #8B5CF6; margin-bottom: 8px; }
.ai-summary-content { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.6; }

/* Key Sections */
.key-section { background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 8px; padding: 16px; margin-bottom: 20px; }
.key-section-header { font-size: 14px; font-weight: 600; color: var(--cm-text); margin-bottom: 12px; }
.key-list, .action-list { list-style: none; padding: 0; margin: 0; }
.key-list li { padding: 8px 0; border-bottom: 1px solid var(--cm-border); font-size: 14px; color: var(--cm-text-secondary); }
.key-list li:last-child { border-bottom: none; }
.key-list li::before { content: "▸ "; color: #10B981; }
.action-item { padding: 8px 0; border-bottom: 1px solid var(--cm-border); }
.action-item:last-child { border-bottom: none; }

/* Markdown Body */
.markdown-body { color: var(--cm-text-secondary); line-height: 1.7; font-size: 15px; }
.markdown-body :deep(h1) { font-size: 24px; color: var(--cm-text); border-bottom: 1px solid var(--cm-border); padding-bottom: 8px; margin-top: 24px; }
.markdown-body :deep(h2) { font-size: 20px; color: var(--cm-text); border-bottom: 1px solid var(--cm-border); padding-bottom: 6px; margin-top: 20px; }
.markdown-body :deep(h3) { font-size: 16px; color: var(--cm-text); margin-top: 16px; }
.markdown-body :deep(p) { margin: 8px 0; }
.markdown-body :deep(code) { background: var(--cm-border); padding: 2px 6px; border-radius: 4px; font-size: 13px; color: #10B981; }
.markdown-body :deep(pre) { background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 8px; padding: 16px; overflow-x: auto; margin: 12px 0; }
.markdown-body :deep(pre code) { background: none; padding: 0; color: var(--cm-text-secondary); }
.markdown-body :deep(ul), .markdown-body :deep(ol) { padding-left: 24px; }
.markdown-body :deep(li) { margin: 4px 0; }
.markdown-body :deep(blockquote) { border-left: 3px solid #10B981; padding-left: 12px; color: var(--cm-text-muted); margin: 12px 0; }
.markdown-body :deep(a) { color: #10B981; text-decoration: none; }
.markdown-body :deep(a:hover) { text-decoration: underline; }
.markdown-body :deep(table) { border-collapse: collapse; width: 100%; margin: 12px 0; }
.markdown-body :deep(th), .markdown-body :deep(td) { border: 1px solid var(--cm-border); padding: 8px 12px; text-align: left; }
.markdown-body :deep(th) { background: var(--cm-bg); color: var(--cm-text); font-weight: 600; }
.markdown-body :deep(hr) { border: none; border-top: 1px solid var(--cm-border); margin: 20px 0; }
.markdown-body :deep(img) { max-width: 100%; border-radius: 8px; }

/* Empty State */
.empty-state { grid-column: 1/-1; display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 300px; color: var(--cm-text-placeholder); }
.empty-icon { font-size: 48px; margin-bottom: 12px; color: #10B981; }

/* Editor */
.editor-row { display: flex; gap: 16px; flex-wrap: wrap; }
.editor-title { flex: 2; min-width: 200px; }
.editor-category { flex: 1; min-width: 150px; }
.editor-status { flex: 1; min-width: 150px; }
.editor-tags { flex: 2; min-width: 200px; }
.editor-parent { flex: 1; min-width: 150px; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 8px; width: 100%; }

.wiki-editor :deep(textarea) {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .wiki-layout { flex-direction: column; }
  .wiki-sidebar { width: 100%; }
  .page-grid { grid-template-columns: 1fr; }
}
</style>