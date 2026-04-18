<template>
  <div class="knowledge-page">
    <div class="page-header">
      <h1>{{ $t('knowledge.title') }}</h1>
      <div class="header-actions">
        <el-button @click="activeTab = 'graph'">
          <el-icon><Connection /></el-icon> {{ $t('knowledge.graphView') }}
        </el-button>
        <el-button type="primary" @click="showEntityDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('knowledge.addEntity') }}
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="main-tabs">
      <el-tab-pane :label="$t('knowledge.graphView')" name="graph">
        <div class="graph-container" ref="graphContainer">
          <div v-if="graphData.nodes.length === 0" class="empty-graph">
            <div class="empty-icon">◇</div>
            <p>{{ $t('knowledge.emptyGraph') }}</p>
            <p class="empty-hint">{{ $t('knowledge.emptyHint') }}</p>
          </div>
          <div v-else ref="cyContainer" class="cy-container"></div>
        </div>
      </el-tab-pane>

      <el-tab-pane name="entities">
        <template #label>
          {{ $t('knowledge.entities') }} <el-badge :value="entities.length" type="info" :max="99" class="tab-badge" />
        </template>
        <div class="list-header">
          <el-input v-model="entitySearch" :placeholder="$t('knowledge.searchEntity')" clearable class="search-input" />
          <div class="entity-limit" v-if="!licenseFeatures.unlimited_graph">
            {{ entities.length }}/50 ({{ $t('knowledge.freeLimit') }})
          </div>
        </div>
        <div class="entity-grid">
          <div class="entity-card" v-for="e in filteredEntities" :key="e.id">
            <div class="entity-type-tag" :class="e.entity_type">{{ e.entity_type }}</div>
            <div class="entity-name">{{ e.name }}</div>
            <div class="entity-desc" v-if="e.description">{{ truncate(e.description, 60) }}</div>
            <div class="entity-actions">
              <el-button text size="small" @click="editEntity(e)">{{ $t('common.edit') }}</el-button>
              <el-button text size="small" type="danger" @click="deleteEntity(e.id)">{{ $t('common.delete') }}</el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane name="relations">
        <template #label>
          {{ $t('knowledge.relations') }} <el-badge :value="relations.length" type="info" :max="99" class="tab-badge" />
        </template>
        <div class="list-header">
          <el-button type="primary" size="small" @click="showRelationDialog = true">{{ $t('knowledge.addRelation') }}</el-button>
          <div class="entity-limit" v-if="!licenseFeatures.unlimited_graph">
            {{ relations.length }}/100 ({{ $t('knowledge.freeLimit') }})
          </div>
        </div>
        <div class="relation-list">
          <div class="relation-item" v-for="r in relations" :key="r.id">
            <span class="rel-node">{{ getEntityName(r.source_id) }}</span>
            <span class="rel-arrow">—{{ r.relation_type }}→</span>
            <span class="rel-node">{{ getEntityName(r.target_id) }}</span>
            <el-button text size="small" type="danger" @click="deleteRelation(r.id)" class="rel-delete">{{ $t('common.delete') }}</el-button>
          </div>
          <div v-if="!relations.length" class="empty-hint">{{ $t('knowledge.noRelations') }}</div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showEntityDialog" :title="editingEntity ? $t('common.edit') : $t('knowledge.addEntity')" width="440px">
      <el-form label-position="top">
        <el-form-item :label="$t('knowledge.entityName')">
          <el-input v-model="entityForm.name" :placeholder="$t('knowledge.entityNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('knowledge.entityType')">
          <el-input v-model="entityForm.entity_type" :placeholder="$t('knowledge.entityTypePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('knowledge.description')">
          <el-input v-model="entityForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEntityDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveEntity" :loading="saving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRelationDialog" :title="$t('knowledge.addRelation')" width="440px">
      <el-form label-position="top">
        <el-form-item :label="$t('knowledge.sourceEntity')">
          <el-select v-model="relationForm.source_id" :placeholder="$t('knowledge.sourceEntity')" style="width: 100%">
            <el-option v-for="e in entities" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('knowledge.relationType')">
          <el-input v-model="relationForm.relation_type" :placeholder="$t('knowledge.relationTypePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('knowledge.targetEntity')">
          <el-select v-model="relationForm.target_id" :placeholder="$t('knowledge.targetEntity')" style="width: 100%">
            <el-option v-for="e in entities" :key="e.id" :label="e.name" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('knowledge.description')">
          <el-input v-model="relationForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRelationDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveRelation" :loading="saving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection } from '@element-plus/icons-vue'
import axios from '../api/client'
import cytoscape from 'cytoscape'

const { t } = useI18n()
const activeTab = ref('entities')
const entities = ref<any[]>([])
const relations = ref<any[]>([])
const graphData = ref<{ nodes: any[]; edges: any[] }>({ nodes: [], edges: [] })
const entitySearch = ref('')
const showEntityDialog = ref(false)
const showRelationDialog = ref(false)
const editingEntity = ref<any>(null)
const saving = ref(false)
const cyContainer = ref<HTMLElement>()
const licenseFeatures = ref<any>({})

const entityForm = ref({ name: '', entity_type: '', description: '' })
const relationForm = ref({ source_id: null as number | null, relation_type: '', target_id: null as number | null, description: '' })

const filteredEntities = computed(() => {
  if (!entitySearch.value) return entities.value
  const q = entitySearch.value.toLowerCase()
  return entities.value.filter(e => e.name.toLowerCase().includes(q) || e.entity_type.toLowerCase().includes(q))
})

onMounted(async () => {
  await loadAll()
  try {
    const { data } = await axios.get('/api/v1/license/info')
    licenseFeatures.value = (data.features || []).reduce((acc: any, f: string) => { acc[f] = true; return acc }, {})
  } catch {}
})

async function loadAll() { await Promise.all([loadEntities(), loadRelations(), loadGraph()]) }

async function loadEntities() {
  try { const { data } = await axios.get('/api/v1/knowledge/entities'); entities.value = data || [] }
  catch { entities.value = [] }
}

async function loadRelations() {
  try { const { data } = await axios.get('/api/v1/knowledge/relations'); relations.value = data || [] }
  catch { relations.value = [] }
}

async function loadGraph() {
  try {
    const { data } = await axios.get('/api/v1/knowledge/graph')
    graphData.value = data
    if (activeTab.value === 'graph' && data.nodes.length) { await nextTick(); renderGraph() }
  } catch { graphData.value = { nodes: [], edges: [] } }
}

watch(activeTab, async (val) => {
  if (val === 'graph' && graphData.value.nodes.length) { await nextTick(); renderGraph() }
})

function renderGraph() {
  if (!cyContainer.value) return
  const nodes = graphData.value.nodes.map((n: any) => ({ data: { id: String(n.id), label: n.name, type: n.type } }))
  const edges = graphData.value.edges.map((e: any, i: number) => ({ data: { id: `e${i}`, source: String(e.source_id), target: String(e.target_id), label: e.relation_type } }))
  cytoscape({
    container: cyContainer.value, elements: [...nodes, ...edges],
    style: [
      { selector: 'node', style: { 'label': 'data(label)', 'text-wrap': 'wrap', 'background-color': '#00d4aa', 'color': '#e6edf3', 'font-size': 12, 'text-valign': 'center', 'text-halign': 'center', 'width': 60, 'height': 60 } },
      { selector: 'edge', style: { 'label': 'data(label)', 'width': 2, 'line-color': '#30363d', 'target-arrow-color': '#30363d', 'target-arrow-shape': 'triangle', 'font-size': 10, 'color': '#7d8590', 'text-rotation': 'autorotate' } },
    ],
    layout: { name: 'cose', padding: 30 },
  })
}

function editEntity(e: any) {
  editingEntity.value = e
  entityForm.value = { name: e.name, entity_type: e.entity_type, description: e.description || '' }
  showEntityDialog.value = true
}

async function saveEntity() {
  if (!entityForm.value.name || !entityForm.value.entity_type) { ElMessage.warning(t('knowledge.fillRequired')); return }
  saving.value = true
  try {
    if (editingEntity.value) await axios.put(`/api/v1/knowledge/entities/${editingEntity.value.id}`, entityForm.value)
    else await axios.post('/api/v1/knowledge/entities', entityForm.value)
    ElMessage.success(t('common.success'))
    showEntityDialog.value = false; editingEntity.value = null; await loadAll()
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteEntity(id: number) {
  try {
    await ElMessageBox.confirm(t('knowledge.deleteEntityConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/api/v1/knowledge/entities/${id}`)
    ElMessage.success(t('common.success')); await loadAll()
  } catch {}
}

async function saveRelation() {
  if (!relationForm.value.source_id || !relationForm.value.target_id || !relationForm.value.relation_type) { ElMessage.warning(t('knowledge.fillRequired')); return }
  saving.value = true
  try {
    await axios.post('/api/v1/knowledge/relations', relationForm.value)
    ElMessage.success(t('common.success'))
    showRelationDialog.value = false; relationForm.value = { source_id: null, relation_type: '', target_id: null, description: '' }; await loadAll()
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteRelation(id: number) {
  try { await axios.delete(`/api/v1/knowledge/relations/${id}`); ElMessage.success(t('common.success')); await loadAll() } catch {}
}

function getEntityName(id: number) { const e = entities.value.find(e => e.id === id); return e ? e.name : `#${id}` }
function truncate(str: string, len: number) { return str && str.length > len ? str.slice(0, len) + '...' : str }
</script>

<style scoped>
.knowledge-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: #e6edf3; margin: 0; }
.header-actions { display: flex; gap: 8px; }
.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.search-input { width: 250px; }
.entity-limit { font-size: 12px; color: #ffc107; font-weight: 600; }
.entity-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; }
.entity-card { background: #161b22; border: 1px solid #21262d; border-radius: 10px; padding: 14px; }
.entity-type-tag { display: inline-block; padding: 1px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; background: rgba(0,212,170,0.15); color: #00d4aa; margin-bottom: 6px; }
.entity-name { font-size: 14px; font-weight: 600; color: #e6edf3; }
.entity-desc { font-size: 12px; color: #7d8590; margin-top: 4px; }
.entity-actions { margin-top: 8px; display: flex; gap: 4px; }
.relation-list { display: flex; flex-direction: column; gap: 8px; }
.relation-item { display: flex; align-items: center; gap: 8px; background: #161b22; border: 1px solid #21262d; padding: 10px 14px; border-radius: 8px; }
.rel-node { font-weight: 600; color: #e6edf3; }
.rel-arrow { color: #00d4aa; font-size: 13px; }
.rel-delete { margin-left: auto; }
.graph-container { background: #161b22; border: 1px solid #21262d; border-radius: 12px; min-height: 500px; }
.cy-container { width: 100%; height: 500px; }
.empty-graph { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 500px; color: #484f58; }
.empty-icon { font-size: 48px; margin-bottom: 12px; color: #00d4aa; }
.empty-hint { font-size: 13px; color: #7d8590; }
.tab-badge { margin-left: 4px; }
</style>
