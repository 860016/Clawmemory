<template>
  <div class="memories-page">
    <div class="page-header">
      <h1>🧠 {{ $t('memories.title') }}</h1>
      <div class="header-actions">
        <el-button @click="handleOpenClawScan">
          <el-icon><Upload /></el-icon> {{ $t('memories.importOpenClaw') }}
        </el-button>
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> {{ $t('memories.addMemory') }}
        </el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-input v-model="searchQuery" :placeholder="$t('memories.searchPlaceholder')" clearable @keyup.enter="handleSearch" @clear="searchResults = []" class="search-input">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-radio-group v-model="currentLayer" @change="loadMemories" size="default">
        <el-radio-button label="">{{ $t('memories.all') }}</el-radio-button>
        <el-radio-button label="preference">{{ $t('memories.preference') }}</el-radio-button>
        <el-radio-button label="knowledge">{{ $t('memories.knowledge') }}</el-radio-button>
        <el-radio-button label="short_term">{{ $t('memories.shortTerm') }}</el-radio-button>
        <el-radio-button label="private">{{ $t('memories.private') }}</el-radio-button>
      </el-radio-group>
    </div>

    <div v-if="searchResults.length" class="search-results">
      <div class="section-title">{{ $t('memories.searchResults') }} ({{ searchResults.length }})</div>
      <div class="memory-card" v-for="m in searchResults" :key="'s'+m.id">
        <div class="card-top">
          <span class="layer-tag" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</span>
          <span class="importance" :class="importanceClass(m.importance)">{{ (m.importance * 100).toFixed(0) }}%</span>
        </div>
        <div class="card-key">{{ m.key }}</div>
        <div class="card-value">{{ m.value }}</div>
        <div class="card-footer">
          <span class="card-meta">{{ m.source }} · {{ formatTime(m.updated_at) }}</span>
          <div class="card-actions">
            <el-button text size="small" @click="editMemory(m)">{{ $t('common.edit') }}</el-button>
            <el-button text size="small" type="danger" @click="deleteMemory(m.id)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="memory-list">
      <div class="memory-card" v-for="m in memories" :key="m.id">
        <div class="card-top">
          <span class="layer-tag" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</span>
          <span class="importance" :class="importanceClass(m.importance)">{{ (m.importance * 100).toFixed(0) }}%</span>
        </div>
        <div class="card-key">{{ m.key }}</div>
        <div class="card-value">{{ truncate(m.value, 200) }}</div>
        <div class="card-tags" v-if="m.tags && m.tags.length">
          <span class="tag" v-for="t in m.tags" :key="t">{{ t }}</span>
        </div>
        <div class="card-footer">
          <span class="card-meta">{{ m.source }} · {{ formatTime(m.updated_at) }}</span>
          <div class="card-actions">
            <el-button text size="small" @click="editMemory(m)">{{ $t('common.edit') }}</el-button>
            <el-button text size="small" type="danger" @click="deleteMemory(m.id)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination v-model:current-page="currentPage" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="loadMemories" />
    </div>

    <el-dialog v-model="showAddDialog" :title="editingMemory ? $t('memories.editTitle') : $t('memories.addTitle')" width="500px" class="custom-dialog">
      <el-form label-position="top">
        <el-form-item :label="$t('memories.layer')">
          <el-select v-model="form.layer" style="width: 100%">
            <el-option :label="$t('memories.preference') + ' (preference)'" value="preference" />
            <el-option :label="$t('memories.knowledge') + ' (knowledge)'" value="knowledge" />
            <el-option :label="$t('memories.shortTerm') + ' (short_term)'" value="short_term" />
            <el-option :label="$t('memories.private') + ' (private)'" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('memories.titleField')">
          <el-input v-model="form.key" :placeholder="$t('memories.titlePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('memories.content')">
          <el-input v-model="form.value" type="textarea" :rows="4" :placeholder="$t('memories.contentPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('memories.importanceField')">
          <el-slider v-model="form.importance" :min="0" :max="100" :format-tooltip="(v: number) => v + '%'" />
        </el-form-item>
        <el-form-item :label="$t('memories.tags')">
          <el-input v-model="form.tagsStr" :placeholder="$t('memories.tagsPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveMemory" :loading="saving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- OpenClaw 导入对话框 -->
    <el-dialog v-model="showImportDialog" :title="$t('memories.importTitle')" width="650px">
      <div v-if="!scanResult">
        <el-alert v-if="scanError" :title="scanError" type="error" show-icon :closable="false" style="margin-bottom: 16px" />
        <p style="color: var(--cm-text-muted); margin-bottom: 16px">
          {{ $t('memories.importDesc') }}
        </p>
        <el-button type="primary" :loading="scanning" @click="handleScan">
          {{ scanning ? $t('memories.scanning') : $t('memories.startScan') }}
        </el-button>
      </div>
      <div v-else>
        <el-alert v-if="!scanResult.found" :title="$t('memories.noOpenClawDir')" type="warning" show-icon :closable="false" style="margin-bottom: 16px">
          <template #default>
            <p>{{ $t('memories.openClawDirHint') }}</p>
          </template>
        </el-alert>
        <template v-else>
          <p style="margin-bottom: 12px; color: var(--cm-text-muted)">
            {{ $t('memories.detectedDir') }}：<code>{{ scanResult.openclaw_dir }}</code>
          </p>
          <el-table :data="scanResult.agents" stripe style="width: 100%">
            <el-table-column prop="agent_name" :label="$t('memories.agentCol')" width="150" />
            <el-table-column prop="layout" :label="$t('memories.versionCol')" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.layout }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="files" :label="$t('memories.memoryCountCol')" width="80" />
            <el-table-column :label="$t('common.actions')" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="handlePreview(row.agent_name)">{{ $t('memories.preview') }}</el-button>
                <el-button size="small" type="primary" @click="handleImport(row.agent_name)">{{ $t('memories.importBtn') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>

      <!-- Preview panel -->
      <!-- Import loading overlay -->
      <div v-if="importing" class="import-loading">
        <el-icon class="loading-spin" :size="32" color="#10B981"><Loading /></el-icon>
        <p style="margin-top: 12px; color: var(--cm-text-muted); font-size: 14px">{{ $t('memories.importing') }}</p>
      </div>
      <div v-else-if="previewData" style="margin-top: 16px">
        <el-divider content-position="left">{{ previewData.agent_name }} — {{ $t('memories.totalCount', { count: previewData.total }) }}</el-divider>
        <div style="max-height: 300px; overflow-y: auto">
          <div v-for="(mem, idx) in previewData.preview" :key="idx"
            style="padding: 8px; border-bottom: 1px solid var(--cm-border)">
            <div style="display: flex; gap: 8px; align-items: center">
              <el-tag size="small">{{ mem.layer }}</el-tag>
              <strong>{{ mem.key }}</strong>
            </div>
            <p style="margin: 4px 0 0; color: var(--cm-text-muted); font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">
              {{ mem.value?.substring(0, 200) }}
            </p>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Upload, Loading } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
const route = useRoute()
const memories = ref<any[]>([])
const searchResults = ref<any[]>([])
const searchQuery = ref('')
const currentLayer = ref('')
const currentPage = ref(1)
const pageSize = 20
const total = ref(0)
const showAddDialog = ref(false)
const editingMemory = ref<any>(null)
const saving = ref(false)

// OpenClaw import state
const showImportDialog = ref(false)
const scanning = ref(false)
const scanResult = ref<any>(null)
const scanError = ref('')
const previewData = ref<any>(null)
const importing = ref(false)

const form = ref({ layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '' })

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

onMounted(() => {
  loadMemories()
  // Handle ?import=openclaw query param
  if (route.query.import === 'openclaw') {
    handleOpenClawScan()
  }
})

watch(() => route.query.import, (val) => {
  if (val === 'openclaw') {
    handleOpenClawScan()
  }
})

function openAddDialog() {
  editingMemory.value = null
  form.value = { layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '' }
  showAddDialog.value = true
}

async function loadMemories() {
  try {
    const params: any = { page: currentPage.value, size: pageSize }
    if (currentLayer.value) params.layer = currentLayer.value
    const { data } = await axios.get('/memories', { params })
    memories.value = data.items || []
    total.value = data.total || 0
  } catch {}
}

async function handleSearch() {
  if (!searchQuery.value) { searchResults.value = []; return }
  try {
    const { data } = await axios.get('/memories/search/keyword', { params: { q: searchQuery.value, limit: 20 } })
    // FTS5 returns {id, key, value, layer, source, rank} — normalize to match memory card format
    searchResults.value = (data || []).map((m: any) => ({
      ...m,
      importance: m.importance || 0.5,
      tags: m.tags || [],
      updated_at: m.updated_at || new Date().toISOString(),
    }))
  } catch {
    // Fallback: client-side filter
    const q = searchQuery.value.toLowerCase()
    searchResults.value = memories.value.filter(m =>
      m.key?.toLowerCase().includes(q) || m.value?.toLowerCase().includes(q)
    )
  }
}

function editMemory(m: any) {
  editingMemory.value = m
  form.value = { layer: m.layer, key: m.key, value: m.value, importance: Math.round(m.importance * 100), tagsStr: (m.tags || []).join(', ') }
  showAddDialog.value = true
}

async function saveMemory() {
  if (!form.value.key || !form.value.value) { ElMessage.warning(t('memories.fillRequired')); return }
  saving.value = true
  try {
    const payload: any = { layer: form.value.layer, key: form.value.key, value: form.value.value, importance: form.value.importance / 100, tags: form.value.tagsStr ? form.value.tagsStr.split(',').map((s: string) => s.trim()).filter(Boolean) : [] }
    if (editingMemory.value) await axios.put(`/memories/${editingMemory.value.id}`, payload)
    else await axios.post('/memories', payload)
    ElMessage.success(t('common.success'))
    showAddDialog.value = false
    editingMemory.value = null
    searchResults.value = []
    await loadMemories()
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteMemory(id: number) {
  try {
    await ElMessageBox.confirm(t('memories.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/memories/${id}`)
    ElMessage.success(t('memories.deleted'))
    searchResults.value = []
    await loadMemories()
  } catch {}
}

function importanceClass(v: number) { return v >= 0.7 ? 'high' : v >= 0.3 ? 'medium' : 'low' }
function truncate(str: string, len: number) { return str && str.length > len ? str.slice(0, len) + '...' : str }
function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

// OpenClaw import
function handleOpenClawScan() {
  scanResult.value = null
  scanError.value = ''
  previewData.value = null
  showImportDialog.value = true
}

async function handleScan() {
  scanning.value = true
  scanError.value = ''
  try {
    const { data } = await axios.get('/openclaw-memories/scan')
    scanResult.value = data
  } catch (e: any) {
    scanError.value = e.response?.data?.detail || t('memories.scanFailed')
  } finally {
    scanning.value = false
  }
}

async function handlePreview(agentName: string) {
  previewData.value = null
  try {
    const { data } = await axios.get(`/openclaw-memories/scan/${encodeURIComponent(agentName)}`)
    previewData.value = data
  } catch {
    ElMessage.error(t('memories.previewFailed'))
  }
}

async function handleImport(agentName: string) {
  try {
    await ElMessageBox.confirm(
      t('memories.importConfirm', { name: agentName }),
      t('memories.confirmImport'),
      { type: 'info' }
    )
  } catch { return }

  importing.value = true
  try {
    const { data } = await axios.post('/openclaw-memories/import', {
      agent_name: agentName,
      skip_existing: true,
      layer: 'knowledge',
    })
    const { imported, skipped, errors } = data
    ElMessage.success(t('memories.importDone', { imported, skipped, errors }))
    showImportDialog.value = false
    await loadMemories()
  } catch {
    ElMessage.error(t('memories.importFailed'))
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.memories-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-wrap: wrap; gap: 12px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.header-actions { display: flex; gap: 8px; }
.toolbar { display: flex; gap: 16px; margin-bottom: 20px; align-items: center; flex-wrap: wrap; }
.search-input { width: 300px; }
.section-title { font-size: 14px; font-weight: 600; color: var(--cm-text-muted); margin-bottom: 12px; }
.memory-list, .search-results { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 12px; }
.memory-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; transition: all 0.2s ease; position: relative; overflow: hidden; }
.memory-card::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px; background: linear-gradient(90deg, transparent, var(--cm-primary), transparent); opacity: 0; transition: opacity 0.3s; }
.memory-card:hover { border-color: rgba(16,185,129,0.3); box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.memory-card:hover::before { opacity: 1; }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.layer-tag { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.layer-tag.preference { background: rgba(16,185,129,0.15); color: #10B981; }
.layer-tag.knowledge { background: rgba(6,182,212,0.15); color: #06b6d4; }
.layer-tag.short_term { background: rgba(255,193,7,0.15); color: #ffc107; }
.layer-tag.private { background: rgba(233,30,99,0.15); color: #e91e63; }
.importance { font-size: 11px; font-weight: 600; }
.importance.high { color: #e91e63; }
.importance.medium { color: #ffc107; }
.importance.low { color: var(--cm-text-muted); }
.card-key { font-size: 15px; font-weight: 600; color: var(--cm-text); margin-bottom: 6px; }
.card-value { font-size: 13px; color: var(--cm-text-muted); line-height: 1.6; white-space: pre-wrap; }
.card-tags { display: flex; gap: 4px; flex-wrap: wrap; margin-top: 8px; }
.tag { padding: 1px 8px; background: var(--cm-border); border-radius: 4px; font-size: 11px; color: var(--cm-text-muted); }
.card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; padding-top: 8px; border-top: 1px solid var(--cm-border); }
.card-meta { font-size: 11px; color: var(--cm-text-placeholder); }
.card-actions { display: flex; gap: 4px; }
.pagination { display: flex; justify-content: center; margin-top: 20px; }
.import-loading { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px 0; }
.loading-spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@media (max-width: 768px) { .search-input { width: 100%; } .memory-list, .search-results { grid-template-columns: 1fr; } .header-actions { width: 100%; justify-content: flex-end; } }
</style>
