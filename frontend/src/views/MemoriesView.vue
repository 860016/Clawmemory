<template>
<<<<<<< HEAD
  <div class="memories-view">
    <div class="page-header">
      <h2>记忆管理</h2>
      <p class="page-desc">查看、搜索和管理 AI 的记忆数据</p>
    </div>
    <div class="page-toolbar">
      <el-radio-group v-model="layerFilter" @change="handleFilter">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="preference">偏好</el-radio-button>
        <el-radio-button label="knowledge">知识</el-radio-button>
        <el-radio-button label="short_term">短期</el-radio-button>
        <el-radio-button label="private">私密</el-radio-button>
      </el-radio-group>
      <el-input v-model="searchQuery" placeholder="搜索记忆..." style="width: 300px" @keydown.enter="handleSearchKeyword">
        <template #append>
          <el-dropdown @command="handleSearchType">
            <el-button>搜索</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="keyword">关键词</el-dropdown-item>
                <el-dropdown-item command="semantic">语义</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-input>
      <el-button type="primary" @click="showAddDialog = true">添加记忆</el-button>
      <el-button type="success" @click="handleOpenClawScan">从 OpenClaw 导入</el-button>
=======
  <div class="memories-page">
    <div class="page-header">
      <h1>{{ $t('memories.title') }}</h1>
      <el-button type="primary" @click="openAddDialog">
        <el-icon><Plus /></el-icon> {{ $t('memories.addMemory') }}
      </el-button>
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
    </div>

    <div class="toolbar">
      <el-input v-model="searchQuery" :placeholder="$t('memories.searchPlaceholder')" clearable @keyup.enter="handleSearch" class="search-input">
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
<<<<<<< HEAD
        <el-form-item label="键"><el-input v-model="editMemory.key" /></el-form-item>
        <el-form-item label="值"><el-input v-model="editMemory.value" type="textarea" :rows="8" :autosize="{ minRows: 4, maxRows: 20 }" /></el-form-item>
        <el-form-item label="重要性"><el-input-number v-model="editMemory.importance" :min="0" :max="1" :step="0.1" /></el-form-item>
=======
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveMemory" :loading="saving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- OpenClaw 导入对话框 -->
    <el-dialog v-model="showImportDialog" title="从 OpenClaw 导入记忆" width="650px">
      <div v-if="!scanResult">
        <el-alert v-if="scanError" :title="scanError" type="error" show-icon :closable="false" style="margin-bottom: 16px" />
        <p style="color: #666; margin-bottom: 16px">
          扫描本地 OpenClaw 记忆文件（支持 v1/v2/v3 所有版本），将已有记忆导入到当前系统。
        </p>
        <el-button type="primary" :loading="scanning" @click="handleScan">
          {{ scanning ? '扫描中...' : '开始扫描' }}
        </el-button>
      </div>
      <div v-else>
        <el-alert v-if="!scanResult.found" title="未检测到 OpenClaw 目录" type="warning" show-icon :closable="false" style="margin-bottom: 16px">
          <template #default>
            <p>请确认本机已安装 OpenClaw，且存在 <code>~/.openclaw/</code> 目录</p>
          </template>
        </el-alert>
        <template v-else>
          <p style="margin-bottom: 12px; color: #666">
            检测到 OpenClaw 目录：<code>{{ scanResult.openclaw_dir }}</code>
          </p>
          <el-table :data="scanResult.agents" stripe style="width: 100%">
            <el-table-column prop="agent_name" label="Agent" width="150" />
            <el-table-column prop="layout" label="版本" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.layout }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="files" label="记忆数" width="80" />
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="handlePreview(row.agent_name)">预览</el-button>
                <el-button size="small" type="primary" @click="handleImport(row.agent_name)">导入</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>

      <!-- Preview panel -->
      <div v-if="previewData" style="margin-top: 16px">
        <el-divider content-position="left">{{ previewData.agent_name }} — 共 {{ previewData.total }} 条</el-divider>
        <div style="max-height: 300px; overflow-y: auto">
          <div v-for="(mem, idx) in previewData.preview" :key="idx"
            style="padding: 8px; border-bottom: 1px solid #f0f0f0">
            <div style="display: flex; gap: 8px; align-items: center">
              <el-tag size="small">{{ mem.layer }}</el-tag>
              <strong>{{ mem.key }}</strong>
            </div>
            <p style="margin: 4px 0 0; color: #666; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">
              {{ mem.value?.substring(0, 200) }}
            </p>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
<<<<<<< HEAD
import { useMemoryStore } from '../stores/memory'
import { memoryApi } from '../api/memories'
=======
import { useI18n } from 'vue-i18n'
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import axios from '../api/client'

const { t } = useI18n()
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

<<<<<<< HEAD
// OpenClaw import state
const showImportDialog = ref(false)
const scanning = ref(false)
const scanResult = ref<any>(null)
const scanError = ref('')
const previewData = ref<any>(null)

onMounted(() => memoryStore.fetchMemories())
=======
const form = ref({ layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '' })
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

onMounted(() => loadMemories())

function openAddDialog() {
  editingMemory.value = null
  form.value = { layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '' }
  showAddDialog.value = true
}

async function loadMemories() {
  try {
    const params: any = { page: currentPage.value, size: pageSize }
    if (currentLayer.value) params.layer = currentLayer.value
    const { data } = await axios.get('/api/v1/memories', { params })
    memories.value = data.items || []
    total.value = data.total || 0
  } catch {}
}

async function handleSearch() {
  if (!searchQuery.value) { searchResults.value = []; return }
  try {
    const { data } = await axios.get('/api/v1/memories/search/keyword', { params: { q: searchQuery.value, limit: 20 } })
    searchResults.value = data || []
  } catch { searchResults.value = [] }
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
    if (editingMemory.value) await axios.put(`/api/v1/memories/${editingMemory.value.id}`, payload)
    else await axios.post('/api/v1/memories', payload)
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
    await axios.delete(`/api/v1/memories/${id}`)
    ElMessage.success(t('memories.deleted'))
    searchResults.value = []
    await loadMemories()
  } catch {}
}

function importanceClass(v: number) { return v >= 0.7 ? 'high' : v >= 0.3 ? 'medium' : 'low' }
function truncate(str: string, len: number) { return str && str.length > len ? str.slice(0, len) + '...' : str }
function formatTime(t: string) {
  if (!t) return ''
  const d = new Date(t)
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
    const resp = await memoryApi.scanOpenClaw()
    scanResult.value = resp.data
  } catch (e: any) {
    scanError.value = e.response?.data?.detail || '扫描失败，请确认 OpenClaw 已安装'
  } finally {
    scanning.value = false
  }
}

async function handlePreview(agentName: string) {
  previewData.value = null
  try {
    const resp = await memoryApi.scanOpenClawAgent(agentName)
    previewData.value = resp.data
  } catch {
    ElMessage.error('预览失败')
  }
}

async function handleImport(agentName: string) {
  const targetAgentId = memoryStore.memories.length > 0 ? undefined : undefined
  try {
    await ElMessageBox.confirm(
      `确定将 Agent "${agentName}" 的记忆导入到当前系统？`,
      '确认导入',
      { type: 'info' }
    )
  } catch { return }

  try {
    const resp = await memoryApi.importOpenClaw({
      agent_name: agentName,
      skip_existing: true,
      layer: 'knowledge',
    })
    const { imported, skipped, errors } = resp.data
    ElMessage.success(`导入完成：${imported} 条导入，${skipped} 条跳过，${errors} 条错误`)
    showImportDialog.value = false
    memoryStore.fetchMemories()
  } catch {
    ElMessage.error('导入失败')
  }
}
</script>

<style scoped>
<<<<<<< HEAD
.page-header {
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.page-desc {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.page-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.memories-view :deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}
.memories-view :deep(.el-table th.el-table__cell) {
  background: var(--el-fill-color-lighter);
  font-weight: 600;
}
=======
.memories-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: #e6edf3; margin: 0; }
.toolbar { display: flex; gap: 16px; margin-bottom: 20px; align-items: center; }
.search-input { width: 300px; }
.section-title { font-size: 14px; font-weight: 600; color: #7d8590; margin-bottom: 12px; }
.memory-list, .search-results { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 12px; }
.memory-card { background: #161b22; border: 1px solid #21262d; border-radius: 12px; padding: 16px; transition: border-color 0.2s; }
.memory-card:hover { border-color: rgba(0,212,170,0.3); }
.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.layer-tag { padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.layer-tag.preference { background: rgba(0,212,170,0.15); color: #00d4aa; }
.layer-tag.knowledge { background: rgba(0,188,212,0.15); color: #00bcd4; }
.layer-tag.short_term { background: rgba(255,193,7,0.15); color: #ffc107; }
.layer-tag.private { background: rgba(233,30,99,0.15); color: #e91e63; }
.importance { font-size: 11px; font-weight: 600; }
.importance.high { color: #e91e63; }
.importance.medium { color: #ffc107; }
.importance.low { color: #7d8590; }
.card-key { font-size: 15px; font-weight: 600; color: #e6edf3; margin-bottom: 6px; }
.card-value { font-size: 13px; color: #7d8590; line-height: 1.6; white-space: pre-wrap; }
.card-tags { display: flex; gap: 4px; flex-wrap: wrap; margin-top: 8px; }
.tag { padding: 1px 8px; background: #21262d; border-radius: 4px; font-size: 11px; color: #7d8590; }
.card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 12px; padding-top: 8px; border-top: 1px solid #21262d; }
.card-meta { font-size: 11px; color: #484f58; }
.card-actions { display: flex; gap: 4px; }
.pagination { display: flex; justify-content: center; margin-top: 20px; }
>>>>>>> fb055c7 (feat: v3.0 - Wiki知识库 + 科技感UI + i18n + Rust PyO3核心 + Pro功能)
</style>
