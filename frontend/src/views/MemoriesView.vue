<template>
  <div class="memories-page">
    <div class="page-header">
      <h1>🧠 {{ $t('memories.title') }}</h1>
      <div class="header-actions">
        <template v-if="!mobileActionsCollapsed">
          <el-button @click="showExtractDialog = true">
            <el-icon><MagicStick /></el-icon> {{ $t('memories.extractMemory') }}
          </el-button>
          <el-button type="warning" @click="aiConflictScan" :loading="aiScanning" plain>
            <el-icon><Warning /></el-icon> {{ $t('memories.aiConflictScan') }}
          </el-button>
          <el-button @click="handleOpenClawScan">
            <el-icon><Upload /></el-icon> {{ $t('memories.importOpenClaw') }}
          </el-button>
        </template>
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> {{ $t('memories.addMemory') }}
        </el-button>
        <el-button class="mobile-toggle-btn" @click="mobileActionsCollapsed = !mobileActionsCollapsed">
          <el-icon><component :is="mobileActionsCollapsed ? 'ArrowDown' : 'ArrowUp'" /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-input v-model="searchQuery" :placeholder="$t('memories.searchPlaceholder')" clearable @keyup.enter="handleSearch" @clear="searchResults = []; smartLoadResult = null" class="search-input">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select v-model="searchMode" style="width: 130px" @change="searchResults = []; smartLoadResult = null">
        <el-option :label="$t('memories.searchKeyword')" value="keyword" />
        <el-option :label="$t('memories.searchSemantic')" value="semantic" />
        <el-option :label="$t('memories.searchSmart')" value="smart" />
      </el-select>
      <el-radio-group v-model="currentLayer" @change="loadMemories" size="default">
        <el-radio-button label="">{{ $t('memories.all') }}</el-radio-button>
        <el-radio-button label="preference">{{ $t('memories.preference') }}</el-radio-button>
        <el-radio-button label="knowledge">{{ $t('memories.knowledge') }}</el-radio-button>
        <el-radio-button label="short_term">{{ $t('memories.shortTerm') }}</el-radio-button>
        <el-radio-button label="private">{{ $t('memories.private') }}</el-radio-button>
      </el-radio-group>
      <el-select v-model="currentMemoryType" @change="loadMemories" :placeholder="$t('memories.memoryType')" clearable style="width: 140px">
        <el-option :label="$t('memories.all')" value="" />
        <el-option label="Knowledge" value="knowledge" />
        <el-option label="Feedback" value="feedback" />
        <el-option label="Project" value="project" />
        <el-option label="Reference" value="reference" />
        <el-option label="User" value="user" />
      </el-select>
      <el-select v-model="currentSourceAgent" @change="loadMemories" :placeholder="$t('memories.sourceAgent')" clearable style="width: 140px">
        <el-option :label="$t('memories.all')" value="" />
        <el-option v-for="a in connectedAgents" :key="a.name" :label="a.display_name || a.name" :value="a.name" />
      </el-select>
      <el-select v-model="currentVisibility" @change="loadMemories" :placeholder="$t('memories.visibility')" clearable style="width: 130px">
        <el-option :label="$t('memories.all')" value="" />
        <el-option :label="$t('memories.visibilityPrivate')" value="private" />
        <el-option :label="$t('memories.visibilityShared')" value="shared" />
        <el-option :label="$t('memories.visibilityPublic')" value="public" />
      </el-select>
    </div>

    <div v-if="smartLoadResult" class="smart-load-info">
      <div class="smart-load-header">
        <span>{{ $t('memories.smartLoadInfo', { count: smartLoadResult.memories.length, tokens: smartLoadResult.total_tokens, budget: smartLoadResult.token_budget }) }}</span>
        <el-tag size="small" :type="smartLoadResult.engine === 'smart_v1' ? 'success' : 'info'">{{ smartLoadResult.engine }}</el-tag>
      </div>
      <div v-if="smartLoadResult.suggestions?.length" class="smart-suggestions">
        <span v-for="(s, i) in smartLoadResult.suggestions" :key="i" class="suggestion-item">{{ s }}</span>
      </div>
    </div>

    <div v-if="searchResults.length" class="search-results">
      <div class="section-title">{{ $t('memories.searchResults') }} ({{ searchResults.length }})</div>
      <div class="memory-card" v-for="m in searchResults" :key="'s'+m.id" :class="{ 'smart-summary': m.load_level === 'summary', 'smart-full': m.load_level === 'full' }">
        <div class="card-top">
          <span class="layer-tag" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</span>
          <span class="importance" :class="importanceClass(m.importance)">{{ (m.importance * 100).toFixed(0) }}%</span>
          <el-tag v-if="m.load_level" size="small" :type="m.load_level === 'full' ? 'success' : m.load_level === 'summary' ? 'warning' : 'info'" class="load-level-tag">
            {{ m.load_level }}
          </el-tag>
          <span v-if="m.score" class="score-badge">{{ m.score.toFixed(2) }}</span>
        </div>
        <div class="card-key">{{ m.key }}</div>
        <div class="card-value" v-if="m.is_encrypted">
          <el-tag type="warning" size="small" style="margin-right: 6px">🔒 {{ $t('settings.encrypted') }}</el-tag>
          <el-button text size="small" type="primary" @click="decryptMemory(m)">{{ $t('settings.decrypt') }}</el-button>
        </div>
        <div class="card-value" v-else-if="m.value">{{ m.value }}</div>
        <div class="card-summary" v-if="!m.is_encrypted && !m.value && m.summary">{{ m.summary }} <el-tag size="small" type="info">{{$t('memories.summaryTag')}}</el-tag></div>
        <div class="card-footer">
          <span class="card-meta">{{ m.source }} · {{ formatTime(m.updated_at || m.created_at) }}</span>
          <div class="card-actions">
            <el-button text size="small" type="success" @click="reinforceMemory(m.id)" :title="$t('memories.reinforceTip')">📌</el-button>
            <el-button text size="small" @click="editMemory(m)">{{ $t('common.edit') }}</el-button>
            <el-button text size="small" type="danger" @click="deleteMemory(m.id)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="memory-list">
      <div v-if="loadingList" class="skeleton-list">
        <div class="skeleton-card" v-for="i in 5" :key="i">
          <div class="skeleton-line" style="width: 40%; height: 16px;"></div>
          <div class="skeleton-line" style="width: 80%; height: 14px; margin-top: 10px;"></div>
          <div class="skeleton-line" style="width: 60%; height: 14px; margin-top: 6px;"></div>
        </div>
      </div>
      <template v-else>
        <el-empty v-if="memories.length === 0" :description="$t('common.noData')" />
        <div v-for="m in memories" :key="m.id" class="memory-card">
        <div class="card-top">
          <span class="layer-tag" :class="m.layer">{{ layerLabels[m.layer] || m.layer }}</span>
          <span class="importance" :class="importanceClass(m.importance)">{{ (m.importance * 100).toFixed(0) }}%</span>
          <el-tag v-if="m.is_encrypted" type="warning" size="small">🔒</el-tag>
          <el-tag v-if="m.memory_type && m.memory_type !== 'knowledge'" size="small" class="type-tag">{{ m.memory_type }}</el-tag>
          <el-tag v-if="m.source_agent" size="small" type="info" class="source-agent-tag">{{ m.source_agent }}</el-tag>
          <el-tag v-if="m.visibility && m.visibility !== 'private'" size="small" :type="m.visibility === 'public' ? 'success' : 'warning'" class="visibility-tag">{{ visibilityLabels[m.visibility] || m.visibility }}</el-tag>
          <el-tag v-if="m.reinforce_count > 0" size="small" type="success" class="reinforce-badge">📌 {{ m.reinforce_count }}</el-tag>
          <span v-if="m.verified_at" class="verified-badge" :title="$t('memories.verifiedAt', { date: m.verified_at })">✅</span>
        </div>
        <div class="card-key">{{ m.key }}</div>
        <div class="card-value" v-if="m.is_encrypted">
          <el-tag type="warning" size="small" style="margin-right: 6px">🔒 {{ $t('settings.encrypted') }}</el-tag>
          <span style="color: var(--cm-text-muted); font-size: 12px">{{ $t('settings.encryptedContent') }}</span>
          <el-button text size="small" type="primary" @click="decryptMemory(m)" style="margin-left: 8px">{{ $t('settings.decrypt') }}</el-button>
        </div>
        <div class="card-value" v-else>{{ truncate(m.value, 200) }}</div>
        <div class="card-summary-line" v-if="!m.is_encrypted && m.summary && m.summary !== m.value">💡 {{ m.summary }}</div>
        <div class="card-tags" v-if="m.tags && m.tags.length">
          <span class="tag" v-for="t in m.tags" :key="t">{{ t }}</span>
        </div>
        <div class="card-freshness" v-if="getFreshness(m)">
          <span class="freshness-tag" :class="getFreshness(m!)!.level">{{ getFreshness(m!)!.label }}</span>
        </div>
        <div class="card-footer">
          <span class="card-meta">{{ m.source }} · {{ formatTime(m.updated_at) }}</span>
          <div class="card-actions">
            <el-button text size="small" type="success" @click="reinforceMemory(m.id)" :title="$t('memories.reinforceTip')">📌</el-button>
            <el-button text size="small" type="primary" @click="verifyMemory(m.id)" :title="$t('memories.verifyTip')">✅</el-button>
            <el-button text size="small" @click="editMemory(m)">{{ $t('common.edit') }}</el-button>
            <el-button text size="small" type="danger" @click="deleteMemory(m.id)">{{ $t('common.delete') }}</el-button>
          </div>
        </div>
      </div>
      </template>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination v-model:current-page="currentPage" :page-size="pageSize" :total="total" layout="prev, pager, next" @current-change="loadMemories" />
    </div>

    <el-dialog v-model="showAddDialog" :title="editingMemory ? $t('memories.editTitle') : $t('memories.addTitle')" width="600px" class="custom-dialog">
      <el-form label-position="top">
        <el-form-item :label="$t('memories.layer')">
          <el-select v-model="form.layer" style="width: 100%">
            <el-option :label="$t('memories.preference') + ' (preference)'" value="preference" />
            <el-option :label="$t('memories.knowledge') + ' (knowledge)'" value="knowledge" />
            <el-option :label="$t('memories.shortTerm') + ' (short_term)'" value="short_term" />
            <el-option :label="$t('memories.private') + ' (private)'" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('memories.memoryType')">
          <el-select v-model="form.memory_type" style="width: 100%">
            <el-option label="Knowledge" value="knowledge" />
            <el-option label="User" value="user" />
            <el-option label="Feedback" value="feedback" />
            <el-option label="Project" value="project" />
            <el-option label="Reference" value="reference" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('memories.titleField')">
          <el-input v-model="form.key" :placeholder="$t('memories.titlePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('memories.content')">
          <el-input v-model="form.value" type="textarea" :rows="8" :placeholder="$t('memories.contentPlaceholder')" resize="vertical" style="min-height: 120px" />
        </el-form-item>
        <el-form-item :label="$t('memories.importanceField')">
          <el-slider v-model="form.importance" :min="0" :max="100" :format-tooltip="(v: number) => v + '%'" />
        </el-form-item>
        <el-form-item :label="$t('memories.tags')">
          <el-input v-model="form.tagsStr" :placeholder="$t('memories.tagsPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('memories.visibility')">
          <el-select v-model="form.visibility" style="width: 100%">
            <el-option :label="$t('memories.visibilityPrivate') + ' (private)'" value="private" />
            <el-option :label="$t('memories.visibilityShared') + ' (shared)'" value="shared" />
            <el-option :label="$t('memories.visibilityPublic') + ' (public)'" value="public" />
          </el-select>
          <div v-if="form.visibility !== 'private'" class="visibility-warning">
            <el-alert type="warning" :closable="false" style="margin-top: 6px">
              <template #title>{{ form.visibility === 'public' ? $t('memories.visibilityPublicWarning') : $t('memories.visibilitySharedWarning') }}</template>
            </el-alert>
          </div>
        </el-form-item>
        <el-form-item :label="$t('memories.sourceAgent')">
          <el-input v-model="form.source_agent" :placeholder="$t('memories.sourceAgentPlaceholder')" />
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
          <div v-if="scanResult.clients?.length" style="margin-bottom: 12px">
            <p style="color: var(--cm-text-muted); margin-bottom: 8px">{{ $t('memories.detectedClients') }}：</p>
            <div style="display: flex; flex-wrap: wrap; gap: 6px">
              <el-tag v-for="cl in scanResult.clients" :key="cl.name" size="small" type="success">{{ cl.display_name }}</el-tag>
            </div>
          </div>
          <p style="margin-bottom: 12px; color: var(--cm-text-muted)">
            {{ $t('memories.detectedDir') }}：<code>{{ scanResult.scanned_dirs || scanResult.openclaw_dir }}</code>
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

    <!-- Extract Memory Dialog -->
    <el-dialog v-model="showExtractDialog" :title="$t('memories.extractTitle')" width="700px" class="custom-dialog">
      <el-input v-model="extractContent" type="textarea" :rows="8" :placeholder="$t('memories.extractPlaceholder')" />
      <div v-if="extractResult" style="margin-top: 16px">
        <el-divider content-position="left">{{ $t('memories.extractedCount', { count: extractResult.count }) }}</el-divider>
        <div v-if="extractResult.warnings?.length" style="margin-bottom: 12px">
          <el-alert v-for="(w, i) in extractResult.warnings" :key="i" :title="w" type="warning" show-icon :closable="false" style="margin-bottom: 4px" />
        </div>
        <div v-for="(em, idx) in extractResult.memories" :key="idx" class="extract-item">
          <div class="extract-item-header">
            <el-tag size="small">{{ em.layer }}</el-tag>
            <el-tag size="small" type="info">{{ em.memory_type }}</el-tag>
            <span class="extract-reason">{{ em.reason }}</span>
          </div>
          <div class="extract-item-key">{{ em.key }}</div>
          <div class="extract-item-value">{{ truncate(em.value, 150) }}</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showExtractDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="info" @click="handleExtract" :loading="extracting">{{ $t('memories.extractBtn') }}</el-button>
        <el-button type="primary" @click="handleExtractAndSave" :loading="extracting" v-if="extractResult?.count > 0">{{ $t('memories.extractAndSave') }}</el-button>
      </template>
    </el-dialog>

    <!-- Secret Warning Dialog -->
    <el-dialog v-model="showSecretWarning" :title="$t('memories.secretWarningTitle')" width="500px">
      <el-alert type="error" show-icon :closable="false" style="margin-bottom: 12px">
        <template #title>{{ $t('memories.secretWarningMsg') }}</template>
      </el-alert>
      <div v-for="(match, idx) in secretMatches" :key="idx" style="margin-bottom: 8px">
        <el-tag :type="match.severity === 'high' ? 'danger' : 'warning'" size="small">{{ match.description }}</el-tag>
      </div>
      <template #footer>
        <el-button @click="showSecretWarning = false; pendingSaveData = null">{{ $t('common.cancel') }}</el-button>
        <el-button type="warning" @click="forceSaveWithSecret">{{ $t('memories.saveAnyway') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Upload, Loading, MagicStick, Warning } from '@element-plus/icons-vue'
import axios from '../api/go-client'
import { translateError } from '../i18n'
import { memoryApi } from '../api/go-memories'
import { agentApi } from '../api/go-agents'
import { aiApi } from '../api/go-ai'

const { t } = useI18n()
const route = useRoute()
const memories = ref<any[]>([])
const searchResults = ref<any[]>([])
const searchQuery = ref('')
const searchMode = ref('keyword')
const smartLoadResult = ref<any>(null)
const currentLayer = ref('')
const currentMemoryType = ref('')
const currentSourceAgent = ref('')
const currentVisibility = ref('')
const currentPage = ref(1)
const pageSize = 20
const total = ref(0)
const showAddDialog = ref(false)
const editingMemory = ref<any>(null)
const saving = ref(false)
const loadingList = ref(true)
const mobileActionsCollapsed = ref(window.innerWidth <= 768)

// OpenClaw import state
const showImportDialog = ref(false)
const scanning = ref(false)
const scanResult = ref<any>(null)
const scanError = ref('')
const previewData = ref<any>(null)
const importing = ref(false)

const form = ref({ layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '', memory_type: 'knowledge', visibility: 'private', source_agent: '' })

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

const visibilityLabels: Record<string, string> = {
  private: t('memories.visibilityPrivate'),
  shared: t('memories.visibilityShared'),
  public: t('memories.visibilityPublic'),
}

const connectedAgents = ref<any[]>([])

async function loadConnectedAgents() {
  try {
    const { data } = await agentApi.getConnected()
    connectedAgents.value = data.agents || []
  } catch { connectedAgents.value = [] }
}

const showExtractDialog = ref(false)
const extractContent = ref('')
const extractResult = ref<any>(null)
const extracting = ref(false)
const showSecretWarning = ref(false)
const secretMatches = ref<any[]>([])
const pendingSaveData = ref<any>(null)

onMounted(() => {
  loadMemories()
  loadConnectedAgents()
  if (route.query.import === 'openclaw' || route.query.import === 'agent') {
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
  form.value = { layer: 'knowledge', key: '', value: '', importance: 50, tagsStr: '', memory_type: 'knowledge', visibility: 'private', source_agent: '' }
  showAddDialog.value = true
}

async function loadMemories() {
  loadingList.value = true
  try {
    const params: any = { page: currentPage.value, size: pageSize }
    if (currentLayer.value) params.layer = currentLayer.value
    if (currentMemoryType.value) params.memory_type = currentMemoryType.value
    if (currentSourceAgent.value) params.source_agent = currentSourceAgent.value
    if (currentVisibility.value) params.visibility = currentVisibility.value
    const { data } = await axios.get('/memories', { params })
    memories.value = data.items || []
    total.value = data.total || 0
  } catch { memories.value = []; total.value = 0 }
  finally { loadingList.value = false }
}

async function handleSearch() {
  if (!searchQuery.value) { searchResults.value = []; smartLoadResult.value = null; return }

  if (searchMode.value === 'smart') {
    try {
      const { data } = await axios.get('/memories/smart-load', {
        params: { q: searchQuery.value, token_budget: 2000, load_level: 'auto' }
      })
      smartLoadResult.value = data
      searchResults.value = (data.memories || []).map((m: any) => ({
        ...m,
        importance: m.importance || 0.5,
        tags: m.tags || [],
        updated_at: m.updated_at || m.created_at || new Date().toISOString(),
      }))
    } catch {
      smartLoadResult.value = null
      const q = searchQuery.value.toLowerCase()
      searchResults.value = memories.value.filter(m =>
        m.key?.toLowerCase().includes(q) || m.value?.toLowerCase().includes(q)
      )
    }
    return
  }

  const endpoint = searchMode.value === 'semantic' ? '/memories/search/semantic' : '/memories/search/keyword'
  try {
    const { data } = await axios.get(endpoint, { params: { q: searchQuery.value, limit: 20 } })
    const results = data.items || data || []
    searchResults.value = results.map((m: any) => ({
      ...m,
      importance: m.importance || 0.5,
      tags: m.tags || [],
      updated_at: m.updated_at || new Date().toISOString(),
    }))
    smartLoadResult.value = null
  } catch {
    const q = searchQuery.value.toLowerCase()
    searchResults.value = memories.value.filter(m =>
      m.key?.toLowerCase().includes(q) || m.value?.toLowerCase().includes(q)
    )
  }
}

async function reinforceMemory(id: number) {
  try {
    await axios.post(`/memories/${id}/reinforce`)
    ElMessage.success(t('memories.reinforced'))
    await loadMemories()
    if (searchResults.value.length) handleSearch()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

function editMemory(m: any) {
  editingMemory.value = m
  form.value = { layer: m.layer, key: m.key, value: m.value, importance: Math.round(m.importance * 100), tagsStr: (m.tags || []).join(', '), memory_type: m.memory_type || 'knowledge', visibility: m.visibility || 'private', source_agent: m.source_agent || '' }
  showAddDialog.value = true
}

async function saveMemory() {
  if (!form.value.key || !form.value.value) { ElMessage.warning(t('memories.fillRequired')); return }
  saving.value = true
  try {
    const payload: any = { layer: form.value.layer, key: form.value.key, value: form.value.value, importance: form.value.importance / 100, tags: form.value.tagsStr ? form.value.tagsStr.split(',').map((s: string) => s.trim()).filter(Boolean) : [], memory_type: form.value.memory_type, visibility: form.value.visibility || 'private' }
    if (form.value.source_agent) payload.source_agent = form.value.source_agent
    try {
      const { data: scanResult } = await memoryApi.scanSecrets(form.value.key + ' ' + form.value.value)
      if (scanResult?.found) {
        secretMatches.value = scanResult.matches || []
        pendingSaveData.value = payload
        showSecretWarning.value = true
        saving.value = false
        return
      }
    } catch {}
    await doSaveMemory(payload)
  } catch (e: any) { ElMessage.error(translateError(e.response?.data?.error || e.response?.data?.detail, t('common.failed'))) }
  finally { saving.value = false }
}

const aiScanning = ref(false)

async function aiConflictScan() {
  aiScanning.value = true
  try {
    const { data } = await aiApi.conflictScan()
    const conflicts = data.conflicts?.length || 0
    if (conflicts === 0) {
      ElMessage.success(t('memories.aiNoConflicts'))
    } else {
      ElMessage.warning(t('memories.aiConflictsFound', { count: conflicts }))
    }
  } catch (e: any) {
    const errMsg = e.response?.data?.error || ''
    ElMessage.error(translateError(errMsg, t('common.failed')))
  } finally {
    aiScanning.value = false
  }
}

async function deleteMemory(id: number) {
  try {
    await ElMessageBox.confirm(t('memories.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/memories/${id}`)
    ElMessage.success(t('memories.deleted'))
    searchResults.value = []
    await loadMemories()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
    }
  }
}

async function decryptMemory(m: any) {
  try {
    const { data } = await axios.post(`/memories/${m.id}/decrypt`)
    if (data.encrypted) {
      m.value = data.value
      m.is_encrypted = false
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.failed'))
  }
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
    const { data } = await agentApi.scanMemories()
    scanResult.value = data
  } catch (e: any) {
    scanError.value = translateError(e.response?.data?.error || e.response?.data?.detail, t('memories.scanFailed'))
  } finally {
    scanning.value = false
  }
}

async function handlePreview(agentName: string) {
  previewData.value = null
  try {
    const { data } = await agentApi.scanAgentMemories(agentName)
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
    const { data } = await agentApi.importMemories({
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

async function verifyMemory(id: number) {
  try {
    await memoryApi.verify(id)
    ElMessage.success(t('memories.verified'))
    await loadMemories()
  } catch {
    ElMessage.error(t('common.failed'))
  }
}

async function handleExtract() {
  if (!extractContent.value) return
  extracting.value = true
  try {
    const { data } = await memoryApi.extract(extractContent.value)
    extractResult.value = data
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    extracting.value = false
  }
}

async function handleExtractAndSave() {
  if (!extractContent.value) return
  extracting.value = true
  try {
    const { data } = await memoryApi.extractAndSave(extractContent.value, true)
    ElMessage.success(t('memories.extractSaved', { count: data.saved }))
    showExtractDialog.value = false
    extractContent.value = ''
    extractResult.value = null
    await loadMemories()
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    extracting.value = false
  }
}

function getFreshness(m: any): { level: string; label: string } | null {
  if (!m.updated_at) return null
  const days = (Date.now() - new Date(m.updated_at).getTime()) / (1000 * 60 * 60 * 24)
  if (days <= 1) return { level: 'fresh', label: t('memories.freshnessFresh') }
  if (days <= 7) return { level: 'recent', label: t('memories.freshnessRecent') }
  if (days <= 30) return { level: 'aging', label: t('memories.freshnessAging') }
  return { level: 'stale', label: t('memories.freshnessStale') }
}

function forceSaveWithSecret() {
  showSecretWarning.value = false
  if (pendingSaveData.value) {
    doSaveMemory(pendingSaveData.value)
    pendingSaveData.value = null
  }
}

async function doSaveMemory(payload: any) {
  saving.value = true
  try {
    if (editingMemory.value) await axios.put(`/memories/${editingMemory.value.id}`, payload)
    else await axios.post('/memories', payload)
    ElMessage.success(t('common.success'))
    showAddDialog.value = false
    editingMemory.value = null
    searchResults.value = []
    await loadMemories()
  } catch (e: any) { ElMessage.error(translateError(e.response?.data?.error || e.response?.data?.detail, t('common.failed'))) }
  finally { saving.value = false }
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

.skeleton-list { display: flex; flex-direction: column; gap: 12px; }
.skeleton-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; }
.skeleton-line { background: linear-gradient(90deg, var(--cm-border) 25%, rgba(16,185,129,0.08) 50%, var(--cm-border) 75%); background-size: 200% 100%; animation: skeleton-pulse 1.5s ease-in-out infinite; border-radius: 4px; }
@keyframes skeleton-pulse { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }
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
.type-tag { background: rgba(139,92,246,0.15); color: #8b5cf6; border: none; }
.source-agent-tag { background: rgba(6,182,212,0.12); color: #06b6d4; border: none; }
.visibility-tag { border: none; }
.visibility-warning { margin-top: 0; }
.verified-badge { font-size: 12px; }
.card-freshness { margin-top: 6px; }
.freshness-tag { padding: 1px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; }
.freshness-tag.fresh { background: rgba(16,185,129,0.15); color: #10B981; }
.freshness-tag.recent { background: rgba(6,182,212,0.15); color: #06b6d4; }
.freshness-tag.aging { background: rgba(255,193,7,0.15); color: #ffc107; }
.freshness-tag.stale { background: rgba(239,68,68,0.15); color: #ef4444; }
.extract-item { padding: 10px; border: 1px solid var(--cm-border); border-radius: 8px; margin-bottom: 8px; }
.extract-item-header { display: flex; gap: 6px; align-items: center; margin-bottom: 4px; }
.extract-reason { font-size: 11px; color: var(--cm-text-muted); }
.extract-item-key { font-weight: 600; font-size: 13px; margin-bottom: 2px; }
.extract-item-value { font-size: 12px; color: var(--cm-text-muted); }
.card-meta { font-size: 11px; color: var(--cm-text-placeholder); }
.card-actions { display: flex; gap: 4px; }
.pagination { display: flex; justify-content: center; margin-top: 20px; }
.smart-load-info { background: var(--cm-bg-secondary, #f5f7fa); border-radius: 8px; padding: 12px 16px; margin-bottom: 16px; }
.smart-load-header { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.smart-suggestions { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.suggestion-item { background: var(--cm-bg-tertiary, #e4e7ed); border-radius: 4px; padding: 2px 8px; font-size: 12px; color: var(--cm-text-secondary); }
.load-level-tag { margin-left: 4px; }
.score-badge { font-size: 11px; color: var(--cm-text-secondary); margin-left: 4px; background: var(--cm-bg-tertiary, #e4e7ed); padding: 1px 6px; border-radius: 4px; }
.reinforce-badge { margin-left: 4px; }
.card-summary-line { font-size: 12px; color: var(--cm-text-secondary); margin-top: 4px; }
.smart-summary { border-left: 3px solid #e6a23c; }
.smart-full { border-left: 3px solid #67c23a; }
.card-summary { font-size: 13px; color: var(--cm-text-secondary); }
.import-loading { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px 0; }
.loading-spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.mobile-toggle-btn { display: none; }
@media (max-width: 768px) {
  .mobile-toggle-btn { display: inline-flex; }
  .header-actions {
    width: 100%;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 6px;
  }
  .memories-page {
    padding: 16px;
  }
  .search-input {
    width: 100%;
  }
  .memory-list,
  .search-results {
    grid-template-columns: 1fr;
  }
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
  .page-header h1 {
    font-size: 20px;
  }
  .toolbar {
    gap: 10px;
    margin-bottom: 14px;
  }
  .memory-card {
    padding: 14px;
  }
  .card-key {
    font-size: 14px;
  }
  .card-value {
    font-size: 12px;
  }
  .card-footer {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
  .card-actions {
    width: 100%;
    justify-content: flex-end;
  }
  .el-radio-group {
    flex-wrap: wrap;
  }
}

@media (max-width: 480px) {
  .memories-page {
    padding: 12px;
  }
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .header-actions {
    width: 100%;
    flex-direction: column;
  }
  .header-actions .el-button {
    width: 100%;
  }
  .page-header h1 {
    font-size: 18px;
  }
  .memory-card {
    padding: 12px;
    border-radius: 10px;
  }
  .card-key {
    font-size: 13px;
  }
  .card-value {
    font-size: 11px;
    line-height: 1.5;
  }
  .card-top {
    flex-wrap: wrap;
    gap: 4px;
  }
  .layer-tag {
    font-size: 10px;
    padding: 1px 8px;
  }
  .importance {
    font-size: 10px;
  }
  .card-meta {
    font-size: 10px;
  }
  .el-radio-group .el-radio-button {
    margin-bottom: 4px;
  }
  .el-radio-button__inner {
    padding: 6px 10px;
    font-size: 12px;
  }
}
</style>
