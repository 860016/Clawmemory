<template>
  <div class="wiki-page">
    <div class="page-header">
      <h1>📖 {{ $t('wiki.title') }}</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openNewPage">
          <el-icon><Plus /></el-icon> {{ $t('wiki.addPage') }}
        </el-button>
      </div>
    </div>

    <div class="wiki-layout">
      <!-- 侧边栏：分类/导航 -->
      <div class="wiki-sidebar">
        <el-input v-model="searchQuery" :placeholder="$t('wiki.searchPlaceholder')" clearable @keyup.enter="handleSearch" class="search-input" size="small">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <div class="sidebar-section">
          <div class="sidebar-title">{{ $t('wiki.categories') }}</div>
          <div class="category-list">
            <div class="category-item" :class="{ active: !selectedCategory }" @click="selectedCategory = ''; loadPages()">
              {{ $t('wiki.allPages') }}
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
              <span class="pin-badge" v-if="p.is_pinned">◆ {{ $t('wiki.pinned') }}</span>
            </div>
            <div class="card-title">{{ p.title }}</div>
            <div class="card-preview" v-if="p.content">{{ getPreview(p.content) }}</div>
            <div class="card-tags" v-if="p.tags && p.tags.length">
              <span class="tag" v-for="tag in p.tags.slice(0, 3)" :key="tag">{{ tag }}</span>
              <span class="tag" v-if="p.tags.length > 3">+{{ p.tags.length - 3 }}</span>
            </div>
            <div class="card-meta">{{ formatTime(p.updated_at) }}</div>
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
              <el-button text type="danger" @click="deleteCurrentPage">{{ $t('common.delete') }}</el-button>
            </div>
          </div>
          <div class="view-meta">
            <span class="category-tag" v-if="viewingPage.category">{{ viewingPage.category }}</span>
            <span class="view-date">{{ formatTime(viewingPage.updated_at) }}</span>
          </div>
          <h1 class="view-title">{{ viewingPage.title }}</h1>
          <div class="view-tags" v-if="viewingPage.tags && viewingPage.tags.length">
            <span class="tag" v-for="tag in viewingPage.tags" :key="tag">{{ tag }}</span>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, ArrowLeft } from '@element-plus/icons-vue'
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
const showEditor = ref(false)
const isEditing = ref(false)
const currentPageId = ref<number | null>(null)
const viewingPage = ref<any>(null)
const saving = ref(false)

const pageForm = ref({
  title: '', content: '', category: '', tagsStr: '', parent_id: null as number | null, is_pinned: false,
})

const pinnedPages = computed(() => pages.value.filter(p => p.is_pinned))

const renderedContent = computed(() => {
  if (!viewingPage.value?.content) return ''
  return marked(viewingPage.value.content, { breaks: true, gfm: true })
})

onMounted(() => { loadPages(); loadCategories() })

watch(() => route.query.tab, (tab) => {
  if (tab === 'categories') {
    // Just highlight the categories section in sidebar - it's already visible
    selectedCategory.value = ''
  }
})

async function loadPages() {
  try {
    const { data } = await wikiApi.listPages(selectedCategory.value || undefined)
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
  pageForm.value = { title: '', content: '', category: '', tagsStr: '', parent_id: null, is_pinned: false }
  loadAllPages()
  showEditor.value = true
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
    }
    if (isEditing.value && currentPageId.value) {
      await wikiApi.updatePage(currentPageId.value, payload)
    } else {
      await wikiApi.createPage(payload)
    }
    ElMessage.success(t('common.success'))
    showEditor.value = false
    searchResults.value = []
    await Promise.all([loadPages(), loadCategories()])
    // If editing the current viewed page, refresh view
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
    await Promise.all([loadPages(), loadCategories()])
  } catch {}
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
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }

/* Layout */
.wiki-layout { display: flex; gap: 20px; }
.wiki-sidebar { width: 220px; flex-shrink: 0; }
.wiki-main { flex: 1; min-width: 0; }

/* Sidebar */
.search-input { margin-bottom: 16px; }
.sidebar-section { margin-bottom: 16px; }
.sidebar-title { font-size: 11px; font-weight: 600; color: var(--cm-text-placeholder); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 8px; }
.category-list { display: flex; flex-direction: column; gap: 2px; }
.category-item { padding: 6px 10px; border-radius: 6px; font-size: 13px; color: var(--cm-text-muted); cursor: pointer; transition: all 0.15s; }
.category-item:hover { background: rgba(16,185,129,0.08); color: var(--cm-text); }
.category-item.active { background: rgba(16,185,129,0.15); color: #10B981; font-weight: 600; }
.category-item.pinned { color: #10B981; }
.empty-hint-small { font-size: 12px; color: var(--cm-text-placeholder); padding: 4px 10px; }

/* Page Grid */
.page-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.page-card {
  background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px;
  cursor: pointer; transition: all 0.2s;
}
.page-card:hover { border-color: rgba(16,185,129,0.4); transform: translateY(-2px); box-shadow: var(--cm-shadow); }
.card-top { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.category-tag { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; background: rgba(6,182,212,0.15); color: #06b6d4; }
.pin-badge { font-size: 11px; color: #10B981; font-weight: 600; }
.card-title { font-size: 16px; font-weight: 600; color: var(--cm-text); margin-bottom: 6px; }
.card-preview { font-size: 12px; color: var(--cm-text-muted); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 8px; }
.card-tags { display: flex; gap: 4px; flex-wrap: wrap; margin-bottom: 6px; }
.tag { padding: 1px 8px; background: var(--cm-border); border-radius: 4px; font-size: 11px; color: var(--cm-text-muted); }
.card-meta { font-size: 11px; color: var(--cm-text-placeholder); }

/* Page View */
.page-view { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 24px 32px; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.back-btn { color: var(--cm-text-muted); }
.view-actions { display: flex; gap: 4px; }
.view-meta { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.view-date { font-size: 12px; color: var(--cm-text-placeholder); }
.view-title { font-size: 28px; font-weight: 700; color: var(--cm-text); margin: 0 0 12px; }
.view-tags { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 20px; }

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
.editor-row { display: flex; gap: 16px; }
.editor-title { flex: 2; }
.editor-category { flex: 1; }
.editor-tags { flex: 2; }
.editor-parent { flex: 1; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 8px; width: 100%; }

.wiki-editor :deep(textarea) {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .wiki-layout { flex-direction: column; }
  .wiki-sidebar { width: 100%; }
}
</style>
