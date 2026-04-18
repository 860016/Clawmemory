<template>
  <div class="wiki-page">
    <div class="page-header">
      <h1>{{ $t('wiki.title') }}</h1>
      <el-button type="primary" @click="openNewPage">
        <el-icon><Plus /></el-icon> {{ $t('wiki.addPage') }}
      </el-button>
    </div>

    <div class="toolbar">
      <el-input v-model="searchQuery" :placeholder="$t('wiki.searchPlaceholder')" clearable @keyup.enter="handleSearch" class="search-input">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="selectedCategory" :placeholder="$t('wiki.categories')" clearable @change="loadPages" style="width: 180px">
        <el-option :label="$t('wiki.allPages')" value="" />
        <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
      </el-select>
    </div>

    <!-- 搜索结果 -->
    <div v-if="searchResults.length" class="search-results">
      <div class="section-title">{{ $t('wiki.searchResults') }} ({{ searchResults.length }})</div>
      <div class="page-card" v-for="p in searchResults" :key="'s'+p.id" @click="viewPage(p.id)">
        <div class="card-top">
          <span class="category-tag" v-if="p.category">{{ p.category }}</span>
          <span class="pin-badge" v-if="p.is_pinned">◆</span>
        </div>
        <div class="card-title">{{ p.title }}</div>
        <div class="card-meta">{{ $t('wiki.lastUpdated') }}: {{ formatTime(p.updated_at) }}</div>
      </div>
    </div>

    <!-- 页面列表 -->
    <div v-else class="page-list">
      <div class="page-card" v-for="p in pages" :key="p.id" @click="viewPage(p.id)">
        <div class="card-top">
          <span class="category-tag" v-if="p.category">{{ p.category }}</span>
          <span class="pin-badge" v-if="p.is_pinned">◆ {{ $t('wiki.pinned') }}</span>
        </div>
        <div class="card-title">{{ p.title }}</div>
        <div class="card-tags" v-if="p.tags && p.tags.length">
          <span class="tag" v-for="tag in p.tags" :key="tag">{{ tag }}</span>
        </div>
        <div class="card-meta">{{ $t('wiki.lastUpdated') }}: {{ formatTime(p.updated_at) }}</div>
      </div>
      <div v-if="!pages.length" class="empty-hint">{{ $t('common.noData') }}</div>
    </div>

    <!-- 编辑/查看页面 -->
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
          <el-button v-if="isEditing" type="danger" text @click="deleteCurrentPage">{{ $t('common.delete') }}</el-button>
          <div class="footer-right">
            <el-button @click="showEditor = false">{{ $t('common.cancel') }}</el-button>
            <el-button type="primary" @click="savePage" :loading="saving">{{ $t('common.save') }}</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import wikiApi from '../api/wiki'

const { t } = useI18n()
const pages = ref<any[]>([])
const allPages = ref<any[]>([])
const categories = ref<string[]>([])
const searchResults = ref<any[]>([])
const searchQuery = ref('')
const selectedCategory = ref('')
const showEditor = ref(false)
const isEditing = ref(false)
const currentPageId = ref<number | null>(null)
const saving = ref(false)

const pageForm = ref({
  title: '', content: '', category: '', tagsStr: '', parent_id: null as number | null, is_pinned: false,
})

onMounted(() => { loadPages(); loadCategories() })

async function loadPages() {
  try {
    const params: any = {}
    if (selectedCategory.value) params.category = selectedCategory.value
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
    isEditing.value = true
    currentPageId.value = id
    pageForm.value = {
      title: data.title,
      content: data.content || '',
      category: data.category || '',
      tagsStr: (data.tags || []).join(', '),
      parent_id: data.parent_id,
      is_pinned: data.is_pinned,
    }
    await loadAllPages()
    showEditor.value = true
  } catch (e: any) {
    ElMessage.error(t('common.failed'))
  }
}

async function savePage() {
  if (!pageForm.value.title) { ElMessage.warning(t('wiki.fillTitle')); return }
  saving.value = true
  try {
    const payload: any = {
      title: pageForm.value.title,
      content: pageForm.value.content,
      category: pageForm.value.category || null,
      tags: pageForm.value.tagsStr ? pageForm.value.tagsStr.split(',').map((s: string) => s.trim()).filter(Boolean) : [],
      parent_id: pageForm.value.parent_id,
      is_pinned: pageForm.value.is_pinned,
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
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteCurrentPage() {
  if (!currentPageId.value) return
  try {
    await ElMessageBox.confirm(t('wiki.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await wikiApi.deletePage(currentPageId.value)
    ElMessage.success(t('wiki.deleted'))
    showEditor.value = false
    searchResults.value = []
    await Promise.all([loadPages(), loadCategories()])
  } catch {}
}

function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style scoped>
.wiki-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: #e6edf3; margin: 0; }
.toolbar { display: flex; gap: 16px; margin-bottom: 20px; }
.search-input { width: 300px; }
.section-title { font-size: 14px; font-weight: 600; color: #7d8590; margin-bottom: 12px; }
.page-list, .search-results { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.page-card {
  background: #161b22; border: 1px solid #21262d; border-radius: 12px; padding: 16px;
  cursor: pointer; transition: all 0.2s;
}
.page-card:hover { border-color: rgba(0,212,170,0.4); transform: translateY(-2px); }
.card-top { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.category-tag { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; background: rgba(0,188,212,0.15); color: #00bcd4; }
.pin-badge { font-size: 11px; color: #00d4aa; font-weight: 600; }
.card-title { font-size: 16px; font-weight: 600; color: #e6edf3; margin-bottom: 6px; }
.card-tags { display: flex; gap: 4px; flex-wrap: wrap; margin-bottom: 6px; }
.tag { padding: 1px 8px; background: #21262d; border-radius: 4px; font-size: 11px; color: #7d8590; }
.card-meta { font-size: 11px; color: #484f58; }
.empty-hint { text-align: center; padding: 48px; color: #484f58; grid-column: 1/-1; }

.editor-row { display: flex; gap: 16px; }
.editor-title { flex: 2; }
.editor-category { flex: 1; }
.editor-tags { flex: 2; }
.editor-parent { flex: 1; }

.dialog-footer { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.footer-right { display: flex; gap: 8px; }

.wiki-editor :deep(textarea) {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
}
</style>
