<template>
  <div class="knowledge-page">
    <div class="page-header">
      <h1>🕸️ {{ $t('knowledge.title') }}</h1>
      <div class="header-actions">
        <el-button @click="activeTab = 'entities'">
          <el-icon><Grid /></el-icon> {{ $t('knowledge.entities') }}
        </el-button>
        <el-button @click="activeTab = 'relations'">
          <el-icon><Share /></el-icon> {{ $t('knowledge.relations') }}
        </el-button>
        <el-button @click="activeTab = 'graph'">
          <el-icon><Connection /></el-icon> {{ $t('knowledge.graphView') }}
        </el-button>
        <el-button @click="activeTab = 'analysis'" v-if="licenseFeatures.ai_extract">
          <el-icon><DataAnalysis /></el-icon> {{ $t('knowledge.analysis') }}
        </el-button>
        <el-button type="primary" @click="showEntityDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('knowledge.addEntity') }}
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

    <!-- 图谱分析视图 -->
    <div v-if="activeTab === 'analysis'" class="analysis-view">
      <div class="analysis-header">
        <el-input v-model="semanticSearchQuery" :placeholder="$t('knowledge.semanticSearch')" clearable class="semantic-search" @keyup.enter="handleSemanticSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="handleSemanticSearch" :loading="semanticSearching">{{ $t('knowledge.search') }}</el-button>
      </div>

      <!-- 语义搜索结果 -->
      <div v-if="semanticSearchResults.length > 0" class="semantic-results">
        <h3>{{ $t('knowledge.semanticResults') }} ({{ semanticSearchResults.length }})</h3>
        <div class="result-list">
          <div class="result-item" v-for="r in semanticSearchResults" :key="r.entity_id">
            <span class="result-name">{{ r.name }}</span>
            <span class="result-type-badge" :class="r.type">{{ r.type }}</span>
            <span class="result-score">{{ Math.round(r.score * 100) }}%</span>
          </div>
        </div>
      </div>

      <!-- 图谱统计 -->
      <div class="stats-grid">
        <div class="stat-card" v-for="stat in graphStats" :key="stat.label">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>

      <!-- 中心度分析 -->
      <div class="analysis-section" v-if="centralityData.length">
        <h3>🎯 {{ $t('knowledge.centralityAnalysis') }}</h3>
        <div class="centrality-list">
          <div class="centrality-item" v-for="c in centralityData.slice(0, 10)" :key="c.id">
            <span class="centrality-name">{{ c.name }}</span>
            <span class="centrality-type" :class="c.type">{{ c.type }}</span>
            <div class="centrality-bars">
              <div class="bar-row">
                <span class="bar-label">Degree</span>
                <div class="bar"><div class="bar-fill" :style="{ width: (c.degree_centrality * 100) + '%' }"></div></div>
                <span class="bar-value">{{ c.degree_centrality }}</span>
              </div>
              <div class="bar-row">
                <span class="bar-label">Betweenness</span>
                <div class="bar"><div class="bar-fill" :style="{ width: (c.betweenness_centrality * 100) + '%' }"></div></div>
                <span class="bar-value">{{ c.betweenness_centrality }}</span>
              </div>
              <div class="bar-row">
                <span class="bar-label">PageRank</span>
                <div class="bar"><div class="bar-fill" :style="{ width: (c.pagerank * 100) + '%' }"></div></div>
                <span class="bar-value">{{ c.pagerank }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 社区发现 -->
      <div class="analysis-section" v-if="communitiesData.length">
        <h3>🔗 {{ $t('knowledge.communityDetection') }}</h3>
        <div class="community-grid">
          <div class="community-card" v-for="comm in communitiesData" :key="comm.community_id">
            <div class="community-header">
              <span class="community-id">#{{ comm.community_id }}</span>
              <span class="community-size">{{ comm.size }} {{ $t('knowledge.entities') }}</span>
            </div>
            <div class="community-entities">
              <span class="community-entity-tag" v-for="e in comm.entities.slice(0, 8)" :key="e.id">
                {{ e.name }}
              </span>
              <span class="community-entity-tag more" v-if="comm.entities.length > 8">+{{ comm.entities.length - 8 }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="!centralityData.length && !communitiesData.length" class="empty-analysis">
        <div class="empty-icon">◇</div>
        <p>{{ $t('knowledge.emptyAnalysis') }}</p>
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
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection, Grid, Share, DataAnalysis, Search } from '@element-plus/icons-vue'
import axios from '../api/client'
import cytoscape from 'cytoscape'

const { t } = useI18n()
const route = useRoute()
const activeTab = ref((route.query.tab as string) || 'entities')
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

const semanticSearchQuery = ref('')
const semanticSearchResults = ref<any[]>([])
const semanticSearching = ref(false)
const centralityData = ref<any[]>([])
const communitiesData = ref<any[]>([])
const graphStats = ref<any[]>([])

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

async function loadAnalysis() {
  if (!licenseFeatures.value.ai_extract) return
  try {
    const [centrality, communities, stats] = await Promise.all([
      axios.get('/knowledge/analysis/centrality'),
      axios.get('/knowledge/analysis/communities'),
      axios.get('/knowledge/analysis/stats'),
    ])
    centralityData.value = centrality.data || []
    communitiesData.value = communities.data || []
    const s = stats.data || {}
    graphStats.value = [
      { label: t('knowledge.totalEntities'), value: s.total_entities || 0 },
      { label: t('knowledge.totalRelations'), value: s.total_relations || 0 },
      { label: t('knowledge.density'), value: s.density || 0 },
      { label: t('knowledge.avgDegree'), value: s.avg_degree || 0 },
      { label: t('knowledge.connectedComponents'), value: s.connected_components || 0 },
    ]
  } catch {}
}

async function handleSemanticSearch() {
  if (!semanticSearchQuery.value.trim()) return
  semanticSearching.value = true
  try {
    const { data } = await axios.get('/knowledge/entities/search', {
      params: { q: semanticSearchQuery.value, limit: 20 }
    })
    semanticSearchResults.value = data || []
  } catch { semanticSearchResults.value = [] }
  finally { semanticSearching.value = false }
}

watch(activeTab, async (val) => {
  if (val === 'graph' && graphData.value.nodes.length) { await nextTick(); renderGraph() }
  if (val === 'analysis') { await loadAnalysis() }
})

watch(() => route.query.tab, (tab) => {
  if (tab && ['entities', 'relations', 'graph'].includes(tab as string)) {
    activeTab.value = tab as string
  }
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
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-wrap: wrap; gap: 12px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.header-actions { display: flex; gap: 8px; }
.list-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; flex-wrap: wrap; gap: 8px; }
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

@media (max-width: 768px) {
  .knowledge-page {
    padding: 16px;
  }
  .search-input {
    width: 100%;
  }
  .entity-grid {
    grid-template-columns: 1fr;
  }
  .graph-container {
    min-height: 350px;
  }
  .cy-container {
    height: 350px;
  }
  .page-header h1 {
    font-size: 20px;
  }
  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }
  .header-actions .el-button {
    flex: 1;
    min-width: 0;
    justify-content: center;
    padding: 0 8px;
    font-size: 12px;
  }
  .list-header {
    flex-direction: column;
    align-items: stretch;
  }
  .entity-limit {
    text-align: center;
  }
  .relation-item {
    flex-wrap: wrap;
    gap: 4px;
    padding: 8px 10px;
  }
  .rel-node {
    font-size: 13px;
  }
  .rel-arrow {
    font-size: 11px;
  }
  .semantic-search {
    max-width: 100%;
    flex: 1;
  }
  .analysis-header {
    flex-direction: column;
    align-items: stretch;
  }
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .community-grid {
    grid-template-columns: 1fr;
  }
  .centrality-name {
    font-size: 13px;
  }
  .bar-label {
    width: 60px;
    font-size: 10px;
  }
  .bar-value {
    width: 32px;
    font-size: 10px;
  }
}

@media (max-width: 480px) {
  .knowledge-page {
    padding: 12px;
  }
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .page-header h1 {
    font-size: 18px;
  }
  .header-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }
  .header-actions .el-button {
    font-size: 11px;
    padding: 0 6px;
  }
  .entity-card {
    padding: 12px;
  }
  .entity-name {
    font-size: 14px;
  }
  .entity-desc {
    font-size: 12px;
  }
  .group-header {
    margin-bottom: 8px;
  }
  .group-emoji {
    font-size: 18px;
  }
  .group-title {
    font-size: 14px;
  }
  .graph-container {
    min-height: 280px;
  }
  .cy-container {
    height: 280px;
  }
  .stat-card {
    padding: 12px;
  }
  .stat-value {
    font-size: 22px;
  }
  .stat-label {
    font-size: 11px;
  }
}

/* Analysis View */
.analysis-view { display: flex; flex-direction: column; gap: 24px; }
.analysis-header { display: flex; gap: 12px; align-items: center; }
.semantic-search { flex: 1; max-width: 400px; }

/* Semantic Results */
.semantic-results { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; }
.semantic-results h3 { font-size: 16px; color: var(--cm-text); margin: 0 0 12px; }
.result-list { display: flex; flex-direction: column; gap: 8px; }
.result-item { display: flex; align-items: center; gap: 8px; background: var(--cm-bg); padding: 10px 14px; border-radius: 8px; }
.result-name { font-weight: 600; color: var(--cm-text); flex: 1; }
.result-type-badge { padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; }
.result-type-badge.person { background: rgba(64,158,255,0.15); color: #409eff; }
.result-type-badge.organization { background: rgba(103,194,58,0.15); color: #67c23a; }
.result-type-badge.location { background: rgba(230,162,60,0.15); color: #e6a23c; }
.result-type-badge.concept { background: rgba(245,108,108,0.15); color: #f56c6c; }
.result-type-badge.event { background: rgba(144,147,153,0.15); color: #909399; }
.result-type-badge.tool { background: rgba(16,185,129,0.15); color: #10B981; }
.result-type-badge.project { background: rgba(179,127,235,0.15); color: #b37feb; }
.result-type-badge.technology { background: rgba(6,182,212,0.15); color: #06b6d4; }
.result-type-badge.other { background: rgba(125,133,144,0.15); color: var(--cm-text-muted); }
.result-score { font-size: 12px; color: #10B981; font-weight: 600; }

/* Stats Grid */
.stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 12px; }
.stat-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 10px; padding: 16px; text-align: center; }
.stat-value { font-size: 28px; font-weight: 700; color: #10B981; }
.stat-label { font-size: 12px; color: var(--cm-text-muted); margin-top: 4px; }

/* Analysis Sections */
.analysis-section { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; }
.analysis-section h3 { font-size: 16px; color: var(--cm-text); margin: 0 0 16px; }

/* Centrality List */
.centrality-list { display: flex; flex-direction: column; gap: 12px; }
.centrality-item { background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 8px; padding: 12px; }
.centrality-name { font-weight: 600; color: var(--cm-text); margin-right: 8px; }
.centrality-type { padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; }
.centrality-type.person { background: rgba(64,158,255,0.15); color: #409eff; }
.centrality-type.organization { background: rgba(103,194,58,0.15); color: #67c23a; }
.centrality-type.location { background: rgba(230,162,60,0.15); color: #e6a23c; }
.centrality-type.concept { background: rgba(245,108,108,0.15); color: #f56c6c; }
.centrality-type.event { background: rgba(144,147,153,0.15); color: #909399; }
.centrality-type.tool { background: rgba(16,185,129,0.15); color: #10B981; }
.centrality-type.project { background: rgba(179,127,235,0.15); color: #b37feb; }
.centrality-type.technology { background: rgba(6,182,212,0.15); color: #06b6d4; }
.centrality-bars { margin-top: 8px; display: flex; flex-direction: column; gap: 4px; }
.bar-row { display: flex; align-items: center; gap: 8px; }
.bar-label { font-size: 11px; color: var(--cm-text-muted); width: 80px; }
.bar { flex: 1; height: 6px; background: var(--cm-border); border-radius: 3px; overflow: hidden; }
.bar-fill { height: 100%; background: #10B981; border-radius: 3px; transition: width 0.3s; }
.bar-value { font-size: 11px; color: var(--cm-text); width: 40px; text-align: right; }

/* Community Grid */
.community-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; }
.community-card { background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 8px; padding: 12px; }
.community-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.community-id { font-size: 12px; font-weight: 600; color: var(--cm-text); }
.community-size { font-size: 11px; color: var(--cm-text-muted); }
.community-entities { display: flex; gap: 4px; flex-wrap: wrap; }
.community-entity-tag { padding: 2px 8px; background: var(--cm-border); border-radius: 4px; font-size: 11px; color: var(--cm-text-secondary); }
.community-entity-tag.more { background: rgba(16,185,129,0.15); color: #10B981; }

.empty-analysis { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; color: var(--cm-text-placeholder); }
</style>
