<template>
  <div class="projects-page">
    <div class="page-header">
      <div class="header-left">
        <h1>📋 {{ $t('project.title') }}</h1>
        <div class="header-stats">
          <span class="stat-badge active">{{ stats.active }} {{ $t('project.active') }}</span>
          <span class="stat-badge paused">{{ stats.paused }} {{ $t('project.paused') }}</span>
          <span class="stat-badge completed">{{ stats.completed }} {{ $t('project.completed') }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon> {{ $t('project.create') }}
        </el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-input v-model="searchQuery" :placeholder="$t('project.searchPlaceholder')" clearable @keyup.enter="handleSearch" @clear="loadProjects" class="search-input">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-radio-group v-model="filterStatus" @change="loadProjects" size="default">
        <el-radio-button label="">{{ $t('project.all') }}</el-radio-button>
        <el-radio-button label="active">{{ $t('project.active') }}</el-radio-button>
        <el-radio-button label="paused">{{ $t('project.paused') }}</el-radio-button>
        <el-radio-button label="completed">{{ $t('project.completed') }}</el-radio-button>
      </el-radio-group>
    </div>

    <div v-if="loading" class="loading-state">
      <el-icon class="spin-icon" :size="32"><Loading /></el-icon>
      <span>{{ $t('common.loading') }}</span>
    </div>

    <div v-else-if="projects.length === 0" class="empty-state">
      <div class="empty-icon">📋</div>
      <p>{{ $t('project.empty') }}</p>
      <el-button type="primary" @click="openCreateDialog">{{ $t('project.createFirst') }}</el-button>
    </div>

    <div v-else class="projects-grid">
      <div
        v-for="p in projects"
        :key="p.id"
        class="project-card"
        :class="[p.status, { pinned: p.is_pinned }]"
        @click="openProject(p)"
      >
        <div class="card-header">
          <div class="card-title-row">
            <span class="pin-icon" v-if="p.is_pinned">📌</span>
            <h3 class="card-title">{{ p.name }}</h3>
            <span class="status-tag" :class="p.status">{{ statusLabel(p.status) }}</span>
          </div>
          <p class="card-desc" v-if="p.description">{{ truncate(p.description, 120) }}</p>
        </div>

        <div class="card-meta">
          <span v-if="p.category" class="meta-item">
            <el-icon><FolderOpened /></el-icon> {{ p.category }}
          </span>
          <span v-if="p.source_agent" class="meta-item">
            <el-icon><Monitor /></el-icon> {{ p.source_agent }}
          </span>
          <span class="meta-item">
            <el-icon><Clock /></el-icon> {{ formatDate(p.updated_at) }}
          </span>
        </div>

        <div class="card-progress" v-if="p.progress > 0">
          <el-progress :percentage="p.progress" :stroke-width="6" :color="progressColor(p.progress)" />
        </div>

        <div class="card-tags" v-if="parseTags(p.tags).length">
          <el-tag v-for="tag in parseTags(p.tags).slice(0, 3)" :key="tag" size="small" type="info">{{ tag }}</el-tag>
        </div>

        <div class="card-actions" @click.stop>
          <el-button size="small" text @click="openProject(p)">
            <el-icon><View /></el-icon>
          </el-button>
          <el-button size="small" text @click="editProject(p)">
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button size="small" text @click="confirmDelete(p)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <el-dialog v-model="showDetail" :title="currentProject?.name || ''" width="720px" :fullscreen="isMobile" top="5vh" class="project-detail-dialog" destroy-on-close>
      <div v-if="currentProject" class="project-detail">
        <div class="detail-header">
          <div class="detail-status">
            <el-select v-model="currentProject.status" size="small" @change="updateStatus">
              <el-option label="🟢 Active" value="active" />
              <el-option label="🟡 Paused" value="paused" />
              <el-option label="✅ Completed" value="completed" />
            </el-select>
            <el-progress :percentage="currentProject.progress" :stroke-width="8" :color="progressColor(currentProject.progress)" style="width: 120px" />
          </div>
          <div class="detail-meta">
            <span v-if="currentProject.category"><el-icon><FolderOpened /></el-icon> {{ currentProject.category }}</span>
            <span v-if="currentProject.source_agent"><el-icon><Monitor /></el-icon> {{ currentProject.source_agent }}</span>
            <span><el-icon><Clock /></el-icon> {{ formatDate(currentProject.updated_at) }}</span>
          </div>
        </div>

        <div class="detail-description" v-if="currentProject.description">
          <h4>{{ $t('project.description') }}</h4>
          <p>{{ currentProject.description }}</p>
        </div>

        <div class="detail-section" v-if="parseJSON(currentProject.key_decisions).length">
          <h4>🎯 {{ $t('project.keyDecisions') }}</h4>
          <ul class="decision-list">
            <li v-for="(d, i) in parseJSON(currentProject.key_decisions)" :key="i">{{ d }}</li>
          </ul>
        </div>

        <div class="detail-section" v-if="parseJSON(currentProject.action_items).length">
          <h4>📝 {{ $t('project.actionItems') }}</h4>
          <ul class="action-list">
            <li v-for="(a, i) in parseJSON(currentProject.action_items)" :key="i">
              <el-checkbox :model-value="false" @change="removeActionItem(i)" /> {{ a }}
            </li>
          </ul>
        </div>

        <div class="detail-section">
          <div class="section-header">
            <h4>📒 {{ $t('project.notes') }} ({{ notes.length }})</h4>
            <div class="section-actions">
              <el-button size="small" @click="extractFromMemories" :loading="extracting">
                <el-icon><MagicStick /></el-icon> {{ $t('project.extractFromMemories') }}
              </el-button>
              <el-button size="small" type="primary" @click="showAddNote = true">
                <el-icon><Plus /></el-icon> {{ $t('project.addNote') }}
              </el-button>
            </div>
          </div>

          <div v-if="notesLoading" class="notes-loading">
            <el-icon class="spin-icon"><Loading /></el-icon>
          </div>
          <div v-else-if="notes.length === 0" class="notes-empty">
            {{ $t('project.noNotes') }}
          </div>
          <div v-else class="notes-list">
            <div v-for="note in notes" :key="note.id" class="note-item" :class="{ 'key-point': note.is_key_point }">
              <div class="note-header">
                <span class="note-type-badge" :class="note.note_type">{{ noteTypeLabel(note.note_type) }}</span>
                <span v-if="note.is_key_point" class="key-badge">★</span>
                <span class="note-source" v-if="note.source">{{ note.source }}</span>
                <span class="note-time">{{ formatDate(note.created_at) }}</span>
                <el-button size="small" text @click="deleteNote(note.id)"><el-icon><Delete /></el-icon></el-button>
              </div>
              <div class="note-content">{{ note.content }}</div>
            </div>
          </div>
        </div>

        <div class="detail-section">
          <div class="section-header">
            <h4>🔗 {{ $t('project.openclawContext') }}</h4>
          </div>
          <p class="context-hint">{{ $t('project.contextHint') }}</p>
          <el-input
            type="textarea"
            :model-value="openclawPrompt"
            readonly
            :rows="3"
            class="context-input"
          />
          <el-button size="small" @click="copyContext" style="margin-top: 8px">
            <el-icon><CopyDocument /></el-icon> {{ $t('project.copyContext') }}
          </el-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="showCreate" :title="editingProject ? $t('project.edit') : $t('project.create')" width="560px" :fullscreen="isMobile" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item :label="$t('project.name')">
          <el-input v-model="form.name" :placeholder="$t('project.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.description')">
          <el-input type="textarea" v-model="form.description" :rows="3" :placeholder="$t('project.descPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.category')">
          <el-input v-model="form.category" :placeholder="$t('project.categoryPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.status')">
          <el-select v-model="form.status">
            <el-option label="Active" value="active" />
            <el-option label="Paused" value="paused" />
            <el-option label="Completed" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('project.progress')">
          <el-slider v-model="form.progress" :min="0" :max="100" />
        </el-form-item>
        <el-form-item :label="$t('project.tags')">
          <el-input v-model="form.tags" :placeholder="$t('project.tagsPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.keyDecisions')">
          <el-input type="textarea" v-model="form.key_decisions_text" :rows="3" :placeholder="$t('project.decisionsPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.actionItems')">
          <el-input type="textarea" v-model="form.action_items_text" :rows="3" :placeholder="$t('project.actionsPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAddNote" :title="$t('project.addNote')" width="500px" :fullscreen="isMobile" destroy-on-close>
      <el-form :model="noteForm" label-width="80px">
        <el-form-item :label="$t('project.noteContent')">
          <el-input type="textarea" v-model="noteForm.content" :rows="4" :placeholder="$t('project.notePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('project.noteType')">
          <el-select v-model="noteForm.note_type">
            <el-option label="Note" value="note" />
            <el-option label="Decision" value="decision" />
            <el-option label="Question" value="question" />
            <el-option label="Idea" value="idea" />
            <el-option label="Issue" value="issue" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('project.keyPoint')">
          <el-switch v-model="noteForm.is_key_point" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddNote = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitNote" :loading="noteSubmitting">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useIsMobile } from '../composables/useIsMobile'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Search, View, Edit, Delete, Clock, FolderOpened, Monitor,
  MagicStick, Loading, CopyDocument
} from '@element-plus/icons-vue'
import projectApi from '../api/project'

const { t } = useI18n()

const projects = ref<any[]>([])
const { isMobile } = useIsMobile()
const loading = ref(false)
const searchQuery = ref('')
const filterStatus = ref('')
const showDetail = ref(false)
const showCreate = ref(false)
const showAddNote = ref(false)
const currentProject = ref<any>(null)
const editingProject = ref<any>(null)
const notes = ref<any[]>([])
const notesLoading = ref(false)
const extracting = ref(false)
const submitting = ref(false)
const noteSubmitting = ref(false)

const stats = computed(() => {
  const s = { active: 0, paused: 0, completed: 0 }
  projects.value.forEach(p => { if (s[p.status as keyof typeof s] !== undefined) s[p.status as keyof typeof s]++ })
  return s
})

const form = ref<any>({
  name: '', description: '', category: '', status: 'active', progress: 0,
  tags: '', key_decisions_text: '', action_items_text: ''
})

const noteForm = ref({ content: '', note_type: 'note', is_key_point: false })

const openclawPrompt = computed(() => {
  if (!currentProject.value) return ''
  return `@project:${currentProject.value.name}`
})

function statusLabel(s: string) {
  const map: Record<string, string> = { active: '🟢 Active', paused: '🟡 Paused', completed: '✅ Completed' }
  return map[s] || s
}

function noteTypeLabel(t: string) {
  const map: Record<string, string> = { note: '📝', decision: '🎯', question: '❓', idea: '💡', issue: '⚠️', memory_extract: '🧠' }
  return map[t] || '📝'
}

function progressColor(p: number) {
  if (p >= 80) return '#67c23a'
  if (p >= 50) return '#409eff'
  if (p >= 20) return '#e6a23c'
  return '#909399'
}

function truncate(s: string, n: number) {
  return s.length > n ? s.slice(0, n) + '...' : s
}

function formatDate(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString()
}

function parseTags(tags: string) {
  if (!tags) return []
  try {
    const arr = JSON.parse(tags)
    return Array.isArray(arr) ? arr : tags.split(',').filter(Boolean)
  } catch {
    return tags.split(',').filter(Boolean)
  }
}

function parseJSON(val: string) {
  if (!val) return []
  try {
    const arr = JSON.parse(val)
    return Array.isArray(arr) ? arr : []
  } catch {
    return val.split('\n').filter(Boolean)
  }
}

async function loadProjects() {
  loading.value = true
  try {
    const params: any = { page: 1, size: 50 }
    if (filterStatus.value) params.status = filterStatus.value
    const { data } = await projectApi.list(params)
    projects.value = data.items || data || []
  } catch {
    projects.value = []
  } finally {
    loading.value = false
  }
}

async function handleSearch() {
  if (!searchQuery.value.trim()) { loadProjects(); return }
  try {
    const { data } = await projectApi.search(searchQuery.value)
    projects.value = data.items || data || []
  } catch {
    ElMessage.error(t('project.searchFailed'))
  }
}

async function openProject(p: any) {
  currentProject.value = p
  showDetail.value = true
  await loadNotes(p.id)
}

async function loadNotes(projectId: number) {
  notesLoading.value = true
  try {
    const { data } = await projectApi.getNotes(projectId)
    notes.value = data.items || data || []
  } catch {
    notes.value = []
  } finally {
    notesLoading.value = false
  }
}

function openCreateDialog() {
  editingProject.value = null
  form.value = { name: '', description: '', category: '', status: 'active', progress: 0, tags: '', key_decisions_text: '', action_items_text: '' }
  showCreate.value = true
}

function editProject(p: any) {
  editingProject.value = p
  form.value = {
    name: p.name, description: p.description || '', category: p.category || '',
    status: p.status, progress: p.progress || 0, tags: parseTags(p.tags).join(', '),
    key_decisions_text: parseJSON(p.key_decisions).join('\n'),
    action_items_text: parseJSON(p.action_items).join('\n')
  }
  showCreate.value = true
}

async function submitForm() {
  if (!form.value.name.trim()) { ElMessage.warning(t('project.nameRequired')); return }
  submitting.value = true
  try {
    const payload: any = {
      name: form.value.name,
      description: form.value.description,
      category: form.value.category,
      status: form.value.status,
      progress: form.value.progress,
      tags: form.value.tags ? form.value.tags.split(',').map((s: string) => s.trim()).filter(Boolean) : [],
      key_decisions: form.value.key_decisions_text ? form.value.key_decisions_text.split('\n').filter(Boolean) : [],
      action_items: form.value.action_items_text ? form.value.action_items_text.split('\n').filter(Boolean) : [],
    }
    if (editingProject.value) {
      await projectApi.update(editingProject.value.id, payload)
    } else {
      await projectApi.create(payload)
    }
    ElMessage.success(t('common.saved'))
    showCreate.value = false
    await loadProjects()
  } catch {
    ElMessage.error(t('common.saveFailed'))
  } finally {
    submitting.value = false
  }
}

async function updateStatus() {
  if (!currentProject.value) return
  try {
    await projectApi.update(currentProject.value.id, { status: currentProject.value.status })
    ElMessage.success(t('common.saved'))
    await loadProjects()
  } catch {
    ElMessage.error(t('common.saveFailed'))
  }
}

async function confirmDelete(p: any) {
  try {
    await ElMessageBox.confirm(t('project.deleteConfirm', { name: p.name }), t('common.confirm'), { type: 'warning' })
  } catch { return }
  try {
    await projectApi.delete(p.id)
    ElMessage.success(t('common.deleted'))
    await loadProjects()
  } catch {
    ElMessage.error(t('common.deleteFailed'))
  }
}

async function submitNote() {
  if (!noteForm.value.content.trim()) { ElMessage.warning(t('project.noteRequired')); return }
  noteSubmitting.value = true
  try {
    await projectApi.addNote(currentProject.value.id, noteForm.value)
    ElMessage.success(t('common.saved'))
    showAddNote.value = false
    noteForm.value = { content: '', note_type: 'note', is_key_point: false }
    await loadNotes(currentProject.value.id)
  } catch {
    ElMessage.error(t('common.saveFailed'))
  } finally {
    noteSubmitting.value = false
  }
}

async function deleteNote(noteId: number) {
  try {
    await projectApi.deleteNote(noteId)
    await loadNotes(currentProject.value.id)
  } catch {
    ElMessage.error(t('common.deleteFailed'))
  }
}

async function extractFromMemories() {
  if (!currentProject.value) return
  extracting.value = true
  try {
    const { data } = await projectApi.extractMemories(currentProject.value.id)
    ElMessage.success(t('project.extracted', { count: data.extracted || 0 }))
    await loadNotes(currentProject.value.id)
  } catch {
    ElMessage.error(t('project.extractFailed'))
  } finally {
    extracting.value = false
  }
}

function removeActionItem(index: number) {
  if (!currentProject.value) return
  const items = parseJSON(currentProject.value.action_items)
  items.splice(index, 1)
  projectApi.update(currentProject.value.id, { action_items: items })
  currentProject.value.action_items = JSON.stringify(items)
}

async function copyContext() {
  try {
    const { data } = await projectApi.context(currentProject.value.name)
    const text = data.context || ''
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('project.copied'))
  } catch {
    ElMessage.error(t('project.copyFailed'))
  }
}

onMounted(loadProjects)
</script>

<style scoped>
.projects-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
}

.header-stats {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.stat-badge {
  font-size: 13px;
  padding: 2px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.stat-badge.active { background: rgba(103, 194, 58, 0.15); color: #67c23a; }
.stat-badge.paused { background: rgba(230, 162, 60, 0.15); color: #e6a23c; }
.stat-badge.completed { background: rgba(64, 158, 255, 0.15); color: #409eff; }

.header-actions {
  display: flex;
  gap: 8px;
}

.toolbar {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  align-items: center;
}

.search-input {
  max-width: 320px;
}

.loading-state, .empty-state {
  text-align: center;
  padding: 80px 20px;
  color: var(--el-text-color-secondary);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}

.project-card {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.project-card:hover {
  border-color: var(--el-color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.project-card.pinned {
  border-left: 3px solid var(--el-color-warning);
}

.card-header {
  margin-bottom: 12px;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
}

.status-tag.active { background: rgba(103, 194, 58, 0.15); color: #67c23a; }
.status-tag.paused { background: rgba(230, 162, 60, 0.15); color: #e6a23c; }
.status-tag.completed { background: rgba(64, 158, 255, 0.15); color: #409eff; }

.card-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin: 0;
  line-height: 1.5;
}

.card-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--el-text-color-regular);
  margin-bottom: 10px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.card-progress {
  margin-bottom: 10px;
}

.card-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.card-actions {
  display: flex;
  gap: 4px;
  justify-content: flex-end;
  opacity: 0;
  transition: opacity 0.2s;
}

.project-card:hover .card-actions {
  opacity: 1;
}

.project-detail {
  max-height: 70vh;
  overflow-y: auto;
}

.detail-header {
  margin-bottom: 20px;
}

.detail-status {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
}

.detail-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.detail-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.detail-description {
  margin-bottom: 20px;
}

.detail-description h4 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}

.detail-description p {
  font-size: 14px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 10px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.section-header h4 {
  margin-bottom: 0;
}

.section-actions {
  display: flex;
  gap: 8px;
}

.decision-list, .action-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.decision-list li, .action-list li {
  padding: 6px 0;
  font-size: 14px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  display: flex;
  align-items: center;
  gap: 8px;
}

.notes-loading, .notes-empty {
  text-align: center;
  padding: 20px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.notes-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.note-item {
  padding: 12px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.note-item.key-point {
  border-left: 3px solid var(--el-color-warning);
  background: rgba(230, 162, 60, 0.05);
}

.note-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 12px;
}

.note-type-badge {
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.note-type-badge.note { background: rgba(64, 158, 255, 0.1); color: #409eff; }
.note-type-badge.decision { background: rgba(103, 194, 58, 0.1); color: #67c23a; }
.note-type-badge.question { background: rgba(230, 162, 60, 0.1); color: #e6a23c; }
.note-type-badge.idea { background: rgba(144, 147, 153, 0.1); color: #909399; }
.note-type-badge.issue { background: rgba(245, 108, 108, 0.1); color: #f56c6c; }
.note-type-badge.memory_extract { background: rgba(155, 89, 182, 0.1); color: #9b59b6; }

.key-badge {
  color: var(--el-color-warning);
  font-size: 14px;
}

.note-source {
  color: var(--el-text-color-secondary);
}

.note-time {
  color: var(--el-text-color-secondary);
  margin-left: auto;
}

.note-content {
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.context-hint {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.context-input {
  font-family: monospace;
}

@media (max-width: 768px) {
  .project-page { padding: 16px; }
  .page-header h1 { font-size: 20px; }
  .header-actions { width: 100%; flex-wrap: wrap; }
  .header-actions .el-button { flex: 1; min-width: 0; }
  .project-grid { grid-template-columns: 1fr; }
  .project-card { padding: 14px; }
  .detail-header { flex-direction: column; align-items: flex-start; gap: 10px; }
  .detail-stats { flex-wrap: wrap; }
  .notes-list { gap: 8px; }
  .note-card { padding: 10px; }
}
</style>
