<template>
  <div class="knowledge-page">
    <div class="page-header">
      <h1>{{ $t('knowledge.title') }}</h1>
      <div class="header-actions">
        <el-button @click="activeTab = 'graph'">
          <el-icon><Connection /></el-icon> {{ $t('knowledge.graphView') }}
        </el-button>
        <el-button @click="activeTab = 'entities'">
          <el-icon><Grid /></el-icon> {{ $t('knowledge.entities') }}
        </el-button>
        <el-button @click="activeTab = 'relations'">
          <el-icon><Share /></el-icon> {{ $t('knowledge.relations') }}
        </el-button>
        <el-button type="primary" @click="showEntityDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('knowledge.addEntity') }}
        </el-button>
        <el-button v-if="licenseFeatures.ai_extract || licenseFeatures.auto_graph" type="success" @click="$router.push('/pro')">
          🤖 {{ $t('nav.pro') }}
        </el-button>
      </div>
    </div>

    <!-- 实体卡片分组视图（默认） -->
    <div v-if="activeTab === 'entities'" class="entities-view">
      <div class="list-header">
        <el-input v-model="entitySearch" :placeholder="$t('knowledge.searchEntity')" clearable class="search-input" />
        <div class="entity-limit" v-if="!licenseFeatures.unlimited_graph">
          {{ entities.length }}/50 ({{ $t('knowledge.freeLimit') }})
        </div>
      </div>

      <div v-if="filteredEntities.length === 0" class="empty-state">
        <div class="empty-icon">◇</div>
        <p>{{ $t('knowledge.emptyGraph') }}</p>
        <p class="empty-hint">{{ $t('knowledge.emptyHint') }}</p>
      </div>

      <!-- 按类型分组展示 -->
      <div v-else class="category-groups">
        <div v-for="group in entityGroups" :key="group.type" class="category-group">
          <div class="group-header">
            <span class="group-emoji">{{ group.emoji }}</span>
            <span class="group-title">{{ group.label }}</span>
            <span class="group-count">{{ group.items.length }}</span>
          </div>
          <div class="entity-grid">
            <div class="entity-card" v-for="e in group.items" :key="e.id" @click="editEntity(e)">
              <div class="entity-name">{{ e.name }}</div>
              <div class="entity-desc" v-if="e.description">{{ truncate(e.description, 80) }}</div>
              <div class="entity-card-footer">
                <span class="entity-type-badge" :class="e.entity_type">{{ e.entity_type }}</span>
                <div class="entity-actions">
                  <el-button text size="small" @click.stop="deleteEntity(e.id)">{{ $t('common.delete') }}</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 关系列表视图 -->
    <div v-if="activeTab === 'relations'" class="relations-view">
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
    </div>

    <!-- 图谱视图 -->
    <div v-if="activeTab === 'graph'" class="graph-view">
      <div class="graph-container" ref="graphContainer">
        <div v-if="graphData.nodes.length === 0" class="empty-graph">
          <div class="empty-icon">◇</div>
          <p>{{ $t('knowledge.emptyGraph') }}</p>
          <p class="empty-hint">{{ $t('knowledge.emptyHint') }}</p>
        </div>
        <div v-else ref="cyContainer" class="cy-container"></div>
      </div>
    </div>

    <el-dialog v-model="showEntityDialog" :title="editingEntity ? $t('common.edit') : $t('knowledge.addEntity')" width="440px">
      <el-form label-position="top">
        <el-form-item :label="$t('knowledge.entityName')">
          <el-input v-model="entityForm.name" :placeholder="$t('knowledge.entityNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('knowledge.entityType')">
          <el-select v-model="entityForm.entity_type" :placeholder="$t('knowledge.entityTypePlaceholder')" style="width: 100%">
            <el-option v-for="opt in typeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
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
import { Plus, Connection, Grid, Share } from '@element-plus/icons-vue'
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

// Entity type config: emoji + label
const typeConfig: Record<string, { emoji: string; label: string }> = {
  person: { emoji: '👤', label: t('knowledge.typePerson') },
  organization: { emoji: '🏢', label: t('knowledge.typeOrg') },
  location: { emoji: '📍', label: t('knowledge.typeLocation') },
  concept: { emoji: '💡', label: t('knowledge.typeConcept') },
  event: { emoji: '📅', label: t('knowledge.typeEvent') },
  tool: { emoji: '🛠', label: t('knowledge.typeTool') },
  project: { emoji: '📁', label: t('knowledge.typeProject') },
  technology: { emoji: '💻', label: t('knowledge.typeTech') },
  other: { emoji: '📦', label: t('knowledge.typeOther') },
}

const typeOptions = Object.entries(typeConfig).map(([value, { label }]) => ({ value, label }))

const filteredEntities = computed(() => {
  if (!entitySearch.value) return entities.value
  const q = entitySearch.value.toLowerCase()
  return entities.value.filter(e => e.name.toLowerCase().includes(q) || e.entity_type?.toLowerCase().includes(q))
})

const entityGroups = computed(() => {
  const groups: Record<string, any[]> = {}
  for (const e of filteredEntities.value) {
    const type = e.entity_type || 'other'
    if (!groups[type]) groups[type] = []
    groups[type].push(e)
  }
  // Sort: larger groups first, then by type name
  return Object.entries(groups)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([type, items]) => {
      const cfg = typeConfig[type] || typeConfig.other
      return { type, emoji: cfg.emoji, label: cfg.label, items }
    })
    .sort((a, b) => b.items.length - a.items.length)
})

onMounted(async () => {
  await loadAll()
  try {
    const { data } = await axios.get('/license/info')
    licenseFeatures.value = (data.features || []).reduce((acc: any, f: string) => { acc[f] = true; return acc }, {})
  } catch {}
})

async function loadAll() { await Promise.all([loadEntities(), loadRelations(), loadGraph()]) }

async function loadEntities() {
  try { const { data } = await axios.get('/knowledge/entities'); entities.value = data || [] }
  catch { entities.value = [] }
}

async function loadRelations() {
  try { const { data } = await axios.get('/knowledge/relations'); relations.value = data || [] }
  catch { relations.value = [] }
}

async function loadGraph() {
  try {
    const { data } = await axios.get('/knowledge/graph')
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
      { selector: 'node', style: { 'label': 'data(label)', 'text-wrap': 'wrap', 'background-color': '#10B981', 'color': '#e6edf3', 'font-size': 12, 'text-valign': 'center', 'text-halign': 'center', 'width': 60, 'height': 60 } },
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
    if (editingEntity.value) await axios.put(`/knowledge/entities/${editingEntity.value.id}`, entityForm.value)
    else await axios.post('/knowledge/entities', entityForm.value)
    ElMessage.success(t('common.success'))
    showEntityDialog.value = false; editingEntity.value = null; await loadAll()
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteEntity(id: number) {
  try {
    await ElMessageBox.confirm(t('knowledge.deleteEntityConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/knowledge/entities/${id}`)
    ElMessage.success(t('common.success')); await loadAll()
  } catch {}
}

async function saveRelation() {
  if (!relationForm.value.source_id || !relationForm.value.target_id || !relationForm.value.relation_type) { ElMessage.warning(t('knowledge.fillRequired')); return }
  saving.value = true
  try {
    await axios.post('/knowledge/relations', relationForm.value)
    ElMessage.success(t('common.success'))
    showRelationDialog.value = false; relationForm.value = { source_id: null, relation_type: '', target_id: null, description: '' }; await loadAll()
  } catch (e: any) { ElMessage.error(e.response?.data?.detail || t('common.failed')) }
  finally { saving.value = false }
}

async function deleteRelation(id: number) {
  try { await axios.delete(`/knowledge/relations/${id}`); ElMessage.success(t('common.success')); await loadAll() } catch {}
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

/* Category Groups */
.category-groups { display: flex; flex-direction: column; gap: 24px; }
.category-group { }
.group-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--cm-border); }
.group-emoji { font-size: 22px; }
.group-title { font-size: 16px; font-weight: 600; color: var(--cm-text); }
.group-count { font-size: 12px; color: var(--cm-text-muted); background: var(--cm-border); padding: 1px 8px; border-radius: 8px; }

/* Entity Grid - like 5nav */
.entity-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 12px; }
.entity-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 10px; padding: 16px; cursor: pointer; transition: border-color 0.2s, transform 0.15s; }
.entity-card:hover { border-color: rgba(16,185,129,0.3); transform: translateY(-1px); }
.entity-name { font-size: 14px; font-weight: 600; color: var(--cm-text); margin-bottom: 4px; }
.entity-desc { font-size: 12px; color: var(--cm-text-muted); margin-top: 4px; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.entity-card-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 10px; }
.entity-type-badge { display: inline-block; padding: 1px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; }
.entity-type-badge.person { background: rgba(64,158,255,0.15); color: #409eff; }
.entity-type-badge.organization { background: rgba(103,194,58,0.15); color: #67c23a; }
.entity-type-badge.location { background: rgba(230,162,60,0.15); color: #e6a23c; }
.entity-type-badge.concept { background: rgba(245,108,108,0.15); color: #f56c6c; }
.entity-type-badge.event { background: rgba(144,147,153,0.15); color: #909399; }
.entity-type-badge.tool { background: rgba(16,185,129,0.15); color: #10B981; }
.entity-type-badge.project { background: rgba(179,127,235,0.15); color: #b37feb; }
.entity-type-badge.technology { background: rgba(6,182,212,0.15); color: #06b6d4; }
.entity-type-badge.other { background: rgba(125,133,144,0.15); color: var(--cm-text-muted); }

/* Relations */
.relation-list { display: flex; flex-direction: column; gap: 8px; }
.relation-item { display: flex; align-items: center; gap: 8px; background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); padding: 10px 14px; border-radius: 8px; }
.rel-node { font-weight: 600; color: var(--cm-text); }
.rel-arrow { color: #10B981; font-size: 13px; }
.rel-delete { margin-left: auto; }

/* Graph */
.graph-container { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; min-height: 500px; }
.cy-container { width: 100%; height: 500px; }
.empty-graph, .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 300px; color: var(--cm-text-placeholder); }
.empty-icon { font-size: 48px; margin-bottom: 12px; color: #10B981; }
.empty-hint { font-size: 13px; color: var(--cm-text-muted); }

/* Other */
.entity-actions { display: flex; gap: 4px; }
</style>
