<template>
  <div class="session-memory-page">
    <div class="page-header">
      <h1>📋 {{ $t('sessionMemory.title') }}</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> {{ $t('sessionMemory.add') }}
        </el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-input v-model="filterSessionId" :placeholder="$t('sessionMemory.filterSessionId')" clearable @keyup.enter="loadSessions" @clear="loadSessions" style="width: 250px">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="filterStatus" @change="loadSessions" clearable :placeholder="$t('sessionMemory.filterStatus')" style="width: 140px">
        <el-option :label="$t('sessionMemory.statusActive')" value="active" />
        <el-option :label="$t('sessionMemory.statusArchived')" value="archived" />
      </el-select>
    </div>

    <div v-if="loading" class="skeleton-grid">
      <div class="skeleton-card" v-for="i in 3" :key="i">
        <div class="cm-skeleton cm-skeleton-text" style="width:30%"></div>
        <div class="cm-skeleton cm-skeleton-title" style="width:60%"></div>
        <div class="cm-skeleton cm-skeleton-text" style="width:90%"></div>
        <div class="cm-skeleton cm-skeleton-text" style="width:40%"></div>
      </div>
    </div>

    <div v-else class="session-list">
      <div class="session-card" v-for="s in sessions" :key="s.id" @click="openDetail(s)">
        <div class="card-top">
          <span class="session-id-tag" v-if="s.session_id">{{ s.session_id }}</span>
          <el-tag size="small" :type="s.status === 'active' ? 'success' : 'info'">{{ s.status }}</el-tag>
          <span class="token-count" v-if="s.token_count">~{{ s.token_count }} tokens</span>
        </div>
        <div class="card-title">{{ s.title || $t('sessionMemory.untitled') }}</div>
        <div class="card-preview">{{ getPreview(s) }}</div>
        <div class="card-footer">
          <span class="card-meta">{{ formatTime(s.updated_at) }}</span>
          <div class="card-actions">
            <el-button text size="small" @click.stop="editSession(s)">{{ $t('common.edit') }}</el-button>
            <el-button text size="small" type="danger" @click.stop="deleteSession(s.id)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!sessions.length" class="cm-empty-state">
      <div class="cm-empty-icon">📋</div>
      <div class="cm-empty-title">{{ $t('sessionMemory.empty') }}</div>
      <div class="cm-empty-desc">创建一个会话记忆来记录工作状态和上下文信息</div>
      <div class="cm-empty-action">
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> {{ $t('sessionMemory.add') }}
        </el-button>
      </div>
    </div>

    <el-dialog v-model="showDetailDialog" :title="detailSession?.title || $t('sessionMemory.untitled')" width="800px" :fullscreen="isMobile" class="custom-dialog">
      <div v-if="detailSession" class="session-detail">
        <div class="detail-meta">
          <el-tag size="small" :type="detailSession.status === 'active' ? 'success' : 'info'">{{ detailSession.status }}</el-tag>
          <span v-if="detailSession.session_id">{{ $t('sessionMemory.sessionIdLabel') }}: {{ detailSession.session_id }}</span>
          <span v-if="detailSession.token_count">~{{ detailSession.token_count }} {{ $t('sessionMemory.tokenCountLabel') }}</span>
        </div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="$t('sessionMemory.currentState')" v-if="detailSession.current_state">{{ detailSession.current_state }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.taskSpec')" v-if="detailSession.task_spec">{{ detailSession.task_spec }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.filesAndFuncs')" v-if="detailSession.files_and_funcs">{{ detailSession.files_and_funcs }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.workflow')" v-if="detailSession.workflow">{{ detailSession.workflow }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.errors')" v-if="detailSession.errors">{{ detailSession.errors }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.docs')" v-if="detailSession.docs">{{ detailSession.docs }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.learnings')" v-if="detailSession.learnings">{{ detailSession.learnings }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.keyResults')" v-if="detailSession.key_results">{{ detailSession.key_results }}</el-descriptions-item>
          <el-descriptions-item :label="$t('sessionMemory.worklog')" v-if="detailSession.worklog">{{ detailSession.worklog }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="showDetailDialog = false">{{ $t('common.close') }}</el-button>
        <el-button type="primary" @click="editSession(detailSession); showDetailDialog = false">{{ $t('common.edit') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAddDialog" :title="editingSession ? $t('sessionMemory.editTitle') : $t('sessionMemory.addTitle')" width="700px" :fullscreen="isMobile" class="custom-dialog">
      <el-form label-position="top">
        <el-form-item :label="$t('sessionMemory.sessionId')">
          <el-input v-model="sessionForm.session_id" :placeholder="$t('sessionMemory.sessionIdPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.titleField')">
          <el-input v-model="sessionForm.title" :placeholder="$t('sessionMemory.titlePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.currentState')">
          <el-input v-model="sessionForm.current_state" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.taskSpec')">
          <el-input v-model="sessionForm.task_spec" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.filesAndFuncs')">
          <el-input v-model="sessionForm.files_and_funcs" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.workflow')">
          <el-input v-model="sessionForm.workflow" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.errors')">
          <el-input v-model="sessionForm.errors" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.docs')">
          <el-input v-model="sessionForm.docs" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.learnings')">
          <el-input v-model="sessionForm.learnings" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.keyResults')">
          <el-input v-model="sessionForm.key_results" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="$t('sessionMemory.worklog')">
          <el-input v-model="sessionForm.worklog" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveSession" :loading="saving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useIsMobile } from '../composables/useIsMobile'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { sessionMemoryApi } from '../api/go-memories'
import { translateError } from '../i18n'

const { t } = useI18n()
const sessions = ref<any[]>([])
const { isMobile } = useIsMobile()
const loading = ref(true)
const filterSessionId = ref('')
const filterStatus = ref('')
const showDetailDialog = ref(false)
const detailSession = ref<any>(null)
const showAddDialog = ref(false)
const editingSession = ref<any>(null)
const saving = ref(false)

const emptyForm = () => ({
  session_id: '', title: '', current_state: '', task_spec: '',
  files_and_funcs: '', workflow: '', errors: '', docs: '',
  learnings: '', key_results: '', worklog: '',
})
const sessionForm = ref(emptyForm())

onMounted(() => { loadSessions() })

async function loadSessions() {
  try {
    const params: any = {}
    if (filterSessionId.value) params.session_id = filterSessionId.value
    if (filterStatus.value) params.status = filterStatus.value
    const { data } = await sessionMemoryApi.list(params)
    sessions.value = data.items || []
  } catch { sessions.value = [] }
  finally { loading.value = false }
}

function openDetail(s: any) {
  detailSession.value = s
  showDetailDialog.value = true
}

function openAddDialog() {
  editingSession.value = null
  sessionForm.value = emptyForm()
  showAddDialog.value = true
}

function editSession(s: any) {
  editingSession.value = s
  sessionForm.value = {
    session_id: s.session_id || '',
    title: s.title || '',
    current_state: s.current_state || '',
    task_spec: s.task_spec || '',
    files_and_funcs: s.files_and_funcs || '',
    workflow: s.workflow || '',
    errors: s.errors || '',
    docs: s.docs || '',
    learnings: s.learnings || '',
    key_results: s.key_results || '',
    worklog: s.worklog || '',
  }
  showAddDialog.value = true
}

async function saveSession() {
  saving.value = true
  try {
    if (editingSession.value) {
      await sessionMemoryApi.update(editingSession.value.id, sessionForm.value)
    } else {
      await sessionMemoryApi.create(sessionForm.value)
    }
    ElMessage.success(t('common.success'))
    showAddDialog.value = false
    await loadSessions()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    saving.value = false
  }
}

async function deleteSession(id: number) {
  try {
    await ElMessageBox.confirm(t('sessionMemory.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await sessionMemoryApi.delete(id)
    ElMessage.success(t('common.success'))
    await loadSessions()
  } catch { /* user cancelled or delete failed */ }
}

function getPreview(s: any) {
  const parts = [s.current_state, s.task_spec, s.key_results].filter(Boolean)
  const text = parts.join(' · ')
  return text.length > 120 ? text.slice(0, 120) + '...' : text
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.session-memory-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-wrap: wrap; gap: 12px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.header-actions { display: flex; gap: 8px; }
.toolbar { display: flex; gap: 16px; margin-bottom: 20px; align-items: center; flex-wrap: wrap; }
.session-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 12px; }
.session-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; cursor: pointer; transition: all 0.2s ease; }
.session-card:hover { border-color: rgba(16,185,129,0.3); box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.card-top { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.session-id-tag { padding: 2px 8px; background: rgba(16,185,129,0.15); color: #10b981; border-radius: 4px; font-size: 11px; font-weight: 600; }
.token-count { font-size: 11px; color: var(--cm-text-muted); }
.card-title { font-size: 15px; font-weight: 600; color: var(--cm-text); margin-bottom: 6px; }
.card-preview { font-size: 13px; color: var(--cm-text-muted); line-height: 1.5; }
.card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; padding-top: 8px; border-top: 1px solid var(--cm-border); }
.card-meta { font-size: 11px; color: var(--cm-text-placeholder); }
.card-actions { display: flex; gap: 4px; }
.detail-meta { display: flex; gap: 12px; align-items: center; margin-bottom: 16px; font-size: 13px; color: var(--cm-text-muted); }

.skeleton-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 12px; }
.skeleton-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; display: flex; flex-direction: column; gap: 8px; }

@media (max-width: 768px) {
  .session-memory-page { padding: 16px; }
  .page-header h1 { font-size: 20px; }
  .session-list { grid-template-columns: 1fr; }
  .session-card { padding: 14px; }
  .card-title { font-size: 14px; }
  .card-preview { font-size: 12px; }
  .toolbar { gap: 10px; flex-direction: column; align-items: stretch; }
  .toolbar .el-input { width: 100% !important; }
  .toolbar .el-select { width: 100% !important; }
  .header-actions { width: 100%; }
  .header-actions .el-button { flex: 1; }
}

@media (max-width: 480px) {
  .session-memory-page { padding: 12px; }
  .page-header h1 { font-size: 18px; }
  .session-card { padding: 10px; }
  .card-top { flex-wrap: wrap; gap: 4px; }
  .card-title { font-size: 13px; }
  .card-preview { font-size: 11px; }
  .card-footer { flex-direction: column; gap: 6px; align-items: flex-start; }
  .card-actions { width: 100%; justify-content: flex-end; }
  .detail-meta { flex-wrap: wrap; gap: 6px; font-size: 12px; }
}
</style>
