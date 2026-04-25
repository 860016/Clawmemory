<template>
  <div class="knowledge-v2">
    <header class="knowledge-header">
      <div class="header-content">
        <div class="header-title">
          <div class="title-icon">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
            </svg>
          </div>
          <div>
            <h1>{{ $t('knowledge.title') }}</h1>
            <p class="subtitle">{{ entities.length }} {{ $t('knowledge.entities') }} · {{ relations.length }} {{ $t('knowledge.relations') }}</p>
          </div>
        </div>
        <div class="header-actions">
          <div class="search-box">
            <el-icon class="search-icon"><Search /></el-icon>
            <input v-model="searchQuery" :placeholder="$t('knowledge.searchPlaceholder')" class="search-input" />
          </div>
          <button class="cm-btn cm-btn-primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            <span>{{ $t('knowledge.addEntity') }}</span>
          </button>
          <button class="cm-btn cm-btn-secondary" @click="showRelationDialog = true">
            <el-icon><Link /></el-icon>
            <span>{{ $t('knowledge.addRelation') }}</span>
          </button>
        </div>
      </div>
      <div class="view-tabs">
        <button v-for="tab in viewTabs" :key="tab.key" class="tab-btn" :class="{ active: currentView === tab.key }" @click="currentView = tab.key">
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </button>
      </div>
    </header>

    <main class="knowledge-content">
      <div v-if="currentView === 'grid'" class="grid-view">
        <div class="category-filters">
          <button v-for="cat in categories" :key="cat.type" class="filter-chip" :class="{ active: selectedCategory === cat.type }" @click="selectedCategory = selectedCategory === cat.type ? '' : cat.type">
            <span class="chip-icon">{{ cat.icon }}</span>
            <span>{{ cat.label }}</span>
            <span class="chip-count">{{ cat.count }}</span>
          </button>
        </div>
        <div class="entities-grid">
          <div v-for="entity in filteredEntities" :key="entity.id" class="entity-card" @click="openDetail(entity)">
            <div class="card-top">
              <div class="entity-avatar" :style="{ background: getAvatarColor(entity.name) }">
                {{ entity.name.charAt(0).toUpperCase() }}
              </div>
              <div class="entity-meta">
                <h3 class="entity-name">{{ entity.name }}</h3>
                <span class="entity-type-badge" :class="entity.entity_type">{{ getTypeLabel(entity.entity_type) }}</span>
              </div>
            </div>
            <p class="entity-desc">{{ entity.description || $t('knowledge.noDescription') }}</p>
            <div class="card-bottom">
              <div class="relation-info" v-if="getEntityRelations(entity.id).length">
                <el-icon><Connection /></el-icon>
                <span>{{ getEntityRelations(entity.id).length }} {{ $t('knowledge.relations') }}</span>
              </div>
              <div class="confidence-info" v-if="entity.confidence">
                <div class="confidence-bar"><div class="confidence-fill" :style="{ width: entity.confidence * 100 + '%' }"></div></div>
                <span>{{ Math.round(entity.confidence * 100) }}%</span>
              </div>
              <div class="card-actions" @click.stop>
                <button class="card-action-btn" @click="openEditDialog(entity)" :title="$t('common.edit')">
                  <el-icon><Edit /></el-icon>
                </button>
                <button class="card-action-btn danger" @click="confirmDelete(entity.id)" :title="$t('common.delete')">
                  <el-icon><Delete /></el-icon>
                </button>
              </div>
            </div>
          </div>
          <div v-if="!filteredEntities.length" class="empty-state">
            <div class="empty-icon">◇</div>
            <p>{{ $t('knowledge.emptyGraph') }}</p>
            <button class="cm-btn cm-btn-primary" @click="openCreateDialog">{{ $t('knowledge.addEntity') }}</button>
          </div>
        </div>
      </div>

      <div v-if="currentView === 'graph'" class="graph-view-v2">
        <div class="graph-toolbar">
          <div class="toolbar-group">
            <button class="tool-btn" @click="resetGraph" :title="$t('knowledge.reset')">
              <el-icon><RefreshRight /></el-icon>
            </button>
            <button class="tool-btn" @click="zoomIn" :title="$t('knowledge.zoomIn')">
              <el-icon><ZoomIn /></el-icon>
            </button>
            <button class="tool-btn" @click="zoomOut" :title="$t('knowledge.zoomOut')">
              <el-icon><ZoomOut /></el-icon>
            </button>
          </div>
          <div class="toolbar-group">
            <span class="graph-stats">{{ graphData.nodes.length }} {{ $t('knowledge.entities') }} · {{ graphData.edges.length }} {{ $t('knowledge.relations') }}</span>
          </div>
        </div>
        <div ref="graphContainer" class="graph-canvas"></div>
      </div>

      <div v-if="currentView === 'list'" class="list-view">
        <div class="list-header">
          <span class="list-col name">{{ $t('knowledge.name') }}</span>
          <span class="list-col type">{{ $t('knowledge.type') }}</span>
          <span class="list-col relations">{{ $t('knowledge.relations') }}</span>
          <span class="list-col actions">{{ $t('knowledge.actions') }}</span>
        </div>
        <div class="list-body">
          <div v-for="entity in filteredEntities" :key="entity.id" class="list-row" @click="openDetail(entity)">
            <span class="list-col name">
              <div class="entity-avatar-small" :style="{ background: getAvatarColor(entity.name) }">{{ entity.name.charAt(0).toUpperCase() }}</div>
              {{ entity.name }}
            </span>
            <span class="list-col type">
              <span class="type-badge" :class="entity.entity_type">{{ getTypeLabel(entity.entity_type) }}</span>
            </span>
            <span class="list-col relations">{{ getEntityRelations(entity.id).length }}</span>
            <span class="list-col actions" @click.stop>
              <button class="action-btn" @click="openEditDialog(entity)"><el-icon><Edit /></el-icon></button>
              <button class="action-btn danger" @click="confirmDelete(entity.id)"><el-icon><Delete /></el-icon></button>
            </span>
          </div>
        </div>
      </div>
    </main>

    <el-drawer v-model="showDetailDrawer" :title="selectedEntity?.name" size="520px" class="entity-drawer">
      <div v-if="selectedEntity" class="drawer-content">
        <div class="drawer-header">
          <div class="entity-avatar-large" :style="{ background: getAvatarColor(selectedEntity.name) }">{{ selectedEntity.name.charAt(0).toUpperCase() }}</div>
          <div>
            <h2>{{ selectedEntity.name }}</h2>
            <span class="type-badge-large" :class="selectedEntity.entity_type">{{ getTypeLabel(selectedEntity.entity_type) }}</span>
          </div>
          <div class="drawer-actions">
            <button class="cm-btn cm-btn-secondary" @click="openEditDialog(selectedEntity); showDetailDrawer = false">{{ $t('common.edit') }}</button>
            <button class="cm-btn cm-btn-danger" @click="confirmDelete(selectedEntity.id); showDetailDrawer = false">{{ $t('common.delete') }}</button>
          </div>
        </div>
        <div class="drawer-section">
          <h3>{{ $t('knowledge.description') }}</h3>
          <p>{{ selectedEntity.description || $t('knowledge.noDescription') }}</p>
        </div>
        <div class="drawer-section">
          <h3>{{ $t('knowledge.relations') }}</h3>
          <div class="relation-list" v-if="getEntityRelations(selectedEntity.id).length">
            <div v-for="rel in getEntityRelations(selectedEntity.id)" :key="rel.id" class="relation-item">
              <div class="relation-node" :class="{ 'is-source': rel.source_id === selectedEntity.id }">{{ getEntityName(rel.source_id) }}</div>
              <div class="relation-arrow">
                <span class="relation-type">{{ rel.relation_type }}</span>
                <el-icon><ArrowRight /></el-icon>
              </div>
              <div class="relation-node" :class="{ 'is-target': rel.target_id === selectedEntity.id }">{{ getEntityName(rel.target_id) }}</div>
              <button class="action-btn danger" @click="deleteRelation(rel.id)" :title="$t('common.delete')"><el-icon><Delete /></el-icon></button>
            </div>
          </div>
          <p v-else class="empty-hint">{{ $t('knowledge.noRelations') }}</p>
        </div>
        <div class="drawer-section" v-if="selectedEntity.properties && selectedEntity.properties !== '{}'">
          <h3>{{ $t('knowledge.properties') }}</h3>
          <div class="properties-grid">
            <div v-for="(value, key) in parseProperties(selectedEntity.properties)" :key="key" class="property-item">
              <span class="property-key">{{ key }}</span>
              <span class="property-value">{{ value }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="showCreateDialog" :title="$t('knowledge.createEntity')" width="560px">
      <div class="form-grid">
        <div class="form-group">
          <label>{{ $t('knowledge.entityName') }}</label>
          <input v-model="formData.name" class="cm-input" :placeholder="$t('knowledge.entityNamePlaceholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.entityType') }}</label>
          <select v-model="formData.entity_type" class="cm-input">
            <option value="person">{{ $t('knowledge.types.person') }}</option>
            <option value="organization">{{ $t('knowledge.types.organization') }}</option>
            <option value="location">{{ $t('knowledge.types.location') }}</option>
            <option value="concept">{{ $t('knowledge.types.concept') }}</option>
            <option value="technology">{{ $t('knowledge.types.technology') }}</option>
            <option value="event">{{ $t('knowledge.types.event') }}</option>
          </select>
        </div>
        <div class="form-group full-width">
          <label>{{ $t('knowledge.description') }}</label>
          <textarea v-model="formData.description" class="cm-input" rows="3" :placeholder="$t('knowledge.descriptionPlaceholder')"></textarea>
        </div>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showCreateDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="createEntity">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditDialog" :title="$t('knowledge.editEntity')" width="560px">
      <div class="form-grid">
        <div class="form-group">
          <label>{{ $t('knowledge.entityName') }}</label>
          <input v-model="formData.name" class="cm-input" :placeholder="$t('knowledge.entityNamePlaceholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.entityType') }}</label>
          <select v-model="formData.entity_type" class="cm-input">
            <option value="person">{{ $t('knowledge.types.person') }}</option>
            <option value="organization">{{ $t('knowledge.types.organization') }}</option>
            <option value="location">{{ $t('knowledge.types.location') }}</option>
            <option value="concept">{{ $t('knowledge.types.concept') }}</option>
            <option value="technology">{{ $t('knowledge.types.technology') }}</option>
            <option value="event">{{ $t('knowledge.types.event') }}</option>
          </select>
        </div>
        <div class="form-group full-width">
          <label>{{ $t('knowledge.description') }}</label>
          <textarea v-model="formData.description" class="cm-input" rows="3" :placeholder="$t('knowledge.descriptionPlaceholder')"></textarea>
        </div>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showEditDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="updateEntity">{{ $t('common.save') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRelationDialog" :title="$t('knowledge.addRelation')" width="560px">
      <div class="form-grid">
        <div class="form-group">
          <label>{{ $t('knowledge.sourceEntity') }}</label>
          <select v-model="relationForm.source_id" class="cm-input">
            <option v-for="e in entities" :key="e.id" :value="e.id">{{ e.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.relationType') }}</label>
          <input v-model="relationForm.relation_type" class="cm-input" :placeholder="$t('knowledge.relationTypePlaceholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.targetEntity') }}</label>
          <select v-model="relationForm.target_id" class="cm-input">
            <option v-for="e in entities" :key="e.id" :value="e.id">{{ e.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.description') }}</label>
          <input v-model="relationForm.description" class="cm-input" />
        </div>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showRelationDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="createRelation">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Link, Connection, Edit, Delete, RefreshRight, ZoomIn, ZoomOut, ArrowRight, Grid, List } from '@element-plus/icons-vue'
import { useKnowledgeStore } from '@/stores/knowledge'

const { t } = useI18n()
const store = useKnowledgeStore()

const currentView = ref('grid')
const searchQuery = ref('')
const selectedCategory = ref('')
const showDetailDrawer = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showRelationDialog = ref(false)
const selectedEntity = ref<any>(null)
const editingEntityId = ref<number | null>(null)
const formData = ref({ name: '', entity_type: 'person', description: '' })
const relationForm = ref({ source_id: 0, target_id: 0, relation_type: '', description: '' })

const viewTabs = [
  { key: 'grid', label: t('knowledge.gridView'), icon: Grid },
  { key: 'graph', label: t('knowledge.graphView'), icon: Connection },
  { key: 'list', label: t('knowledge.listView'), icon: List },
]

const entities = computed(() => store.entities)
const relations = computed(() => store.relations)

const typeLabels: Record<string, string> = {
  person: t('knowledge.types.person'),
  organization: t('knowledge.types.organization'),
  location: t('knowledge.types.location'),
  concept: t('knowledge.types.concept'),
  technology: t('knowledge.types.technology'),
  event: t('knowledge.types.event'),
}

function getTypeLabel(type: string) {
  return typeLabels[type] || type
}

const categories = computed(() => {
  const types = [...new Set(entities.value.map((e: any) => e.entity_type))]
  const icons: Record<string, string> = { person: '👤', organization: '🏢', location: '📍', concept: '💡', technology: '⚡', event: '📅' }
  return types.map(type => ({
    type,
    label: getTypeLabel(type),
    icon: icons[type] || '📄',
    count: entities.value.filter((e: any) => e.entity_type === type).length,
  }))
})

const filteredEntities = computed(() => {
  let result = entities.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter((e: any) => e.name.toLowerCase().includes(q) || e.description?.toLowerCase().includes(q))
  }
  if (selectedCategory.value) {
    result = result.filter((e: any) => e.entity_type === selectedCategory.value)
  }
  return result
})

const graphData = computed(() => ({
  nodes: entities.value.map((e: any) => ({ id: String(e.id), label: e.name, group: e.entity_type, value: getEntityRelations(e.id).length + 1 })),
  edges: relations.value.map((r: any) => ({ from: String(r.source_id), to: String(r.target_id), label: r.relation_type })),
}))

function getEntityRelations(entityId: number) {
  return relations.value.filter((r: any) => r.source_id === entityId || r.target_id === entityId)
}

function getEntityName(id: number) {
  return entities.value.find((e: any) => e.id === id)?.name || 'Unknown'
}

function getAvatarColor(name: string) {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f43f5e', '#f97316', '#eab308', '#22c55e', '#06b6d4']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

function parseProperties(props: string) {
  try { return JSON.parse(props) } catch { return {} }
}

function openDetail(entity: any) {
  selectedEntity.value = entity
  showDetailDrawer.value = true
}

function openCreateDialog() {
  formData.value = { name: '', entity_type: 'person', description: '' }
  showCreateDialog.value = true
}

function openEditDialog(entity: any) {
  editingEntityId.value = entity.id
  formData.value = { name: entity.name, entity_type: entity.entity_type, description: entity.description || '' }
  showEditDialog.value = true
}

async function createEntity() {
  if (!formData.value.name.trim()) { ElMessage.warning(t('knowledge.fillRequired')); return }
  try {
    await store.createEntity(formData.value)
    showCreateDialog.value = false
    ElMessage.success(t('common.created'))
  } catch { ElMessage.error(t('common.failed')) }
}

async function updateEntity() {
  if (!editingEntityId.value) return
  try {
    await store.updateEntity(editingEntityId.value, formData.value)
    showEditDialog.value = false
    ElMessage.success(t('common.saved'))
    if (selectedEntity.value?.id === editingEntityId.value) {
      selectedEntity.value = { ...selectedEntity.value, ...formData.value }
    }
  } catch { ElMessage.error(t('common.failed')) }
}

async function confirmDelete(id: number) {
  try {
    await ElMessageBox.confirm(t('knowledge.deleteEntityConfirm'), t('common.confirm'), { type: 'warning' })
    await store.deleteEntity(id)
    if (selectedEntity.value?.id === id) showDetailDrawer.value = false
    ElMessage.success(t('common.deleted'))
  } catch {}
}

async function createRelation() {
  if (!relationForm.value.source_id || !relationForm.value.target_id || !relationForm.value.relation_type) {
    ElMessage.warning(t('knowledge.fillRequired'))
    return
  }
  try {
    await store.createRelation(relationForm.value)
    showRelationDialog.value = false
    relationForm.value = { source_id: 0, target_id: 0, relation_type: '', description: '' }
    ElMessage.success(t('common.created'))
  } catch { ElMessage.error(t('common.failed')) }
}

async function deleteRelation(id: number) {
  try {
    await store.deleteRelation(id)
    ElMessage.success(t('common.deleted'))
  } catch { ElMessage.error(t('common.failed')) }
}

function resetGraph() {}
function zoomIn() {}
function zoomOut() {}

onMounted(async () => {
  try {
    await Promise.all([store.fetchEntities(1, 100), store.fetchRelations(1, 100)])
  } catch (e) {
    console.error('Failed to load knowledge data:', e)
  }
})
</script>

<style scoped>
.knowledge-v2 { height: 100%; display: flex; flex-direction: column; }
.knowledge-header { background: var(--cm-bg-primary); border-bottom: 1px solid var(--cm-border); padding: var(--cm-space-6); }
.header-content { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--cm-space-4); }
.header-title { display: flex; align-items: center; gap: var(--cm-space-3); }
.title-icon { width: 48px; height: 48px; border-radius: var(--cm-radius-lg); background: var(--cm-primary-gradient); color: white; display: flex; align-items: center; justify-content: center; }
.header-title h1 { font-size: 24px; font-weight: 700; color: var(--cm-text-primary); }
.subtitle { font-size: 14px; color: var(--cm-text-secondary); margin-top: 2px; }
.header-actions { display: flex; align-items: center; gap: var(--cm-space-3); }
.search-box { position: relative; width: 280px; }
.search-icon { position: absolute; left: var(--cm-space-3); top: 50%; transform: translateY(-50%); color: var(--cm-text-tertiary); }
.search-input { width: 100%; padding: var(--cm-space-2) var(--cm-space-3) var(--cm-space-2) 36px; border: 1px solid var(--cm-border); border-radius: var(--cm-radius-md); background: var(--cm-bg-secondary); color: var(--cm-text-primary); font-size: 14px; }
.search-input:focus { outline: none; border-color: var(--cm-primary); box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1); }
.view-tabs { display: flex; gap: var(--cm-space-1); }
.tab-btn { display: flex; align-items: center; gap: var(--cm-space-2); padding: var(--cm-space-2) var(--cm-space-4); border: none; background: transparent; color: var(--cm-text-secondary); font-size: 14px; font-weight: 500; border-radius: var(--cm-radius-md); cursor: pointer; }
.tab-btn:hover { background: var(--cm-bg-tertiary); color: var(--cm-text-primary); }
.tab-btn.active { background: var(--cm-primary); color: white; }
.knowledge-content { flex: 1; overflow: auto; padding: var(--cm-space-6); }
.category-filters { display: flex; gap: var(--cm-space-2); margin-bottom: var(--cm-space-5); flex-wrap: wrap; }
.filter-chip { display: flex; align-items: center; gap: var(--cm-space-2); padding: var(--cm-space-2) var(--cm-space-3); border: 1px solid var(--cm-border); border-radius: 20px; background: var(--cm-bg-primary); color: var(--cm-text-secondary); font-size: 13px; cursor: pointer; }
.filter-chip:hover { border-color: var(--cm-primary); color: var(--cm-text-primary); }
.filter-chip.active { background: var(--cm-primary); border-color: var(--cm-primary); color: white; }
.chip-icon { font-size: 14px; }
.chip-count { padding: 1px 6px; background: var(--cm-bg-tertiary); border-radius: 20px; font-size: 11px; font-weight: 600; }
.filter-chip.active .chip-count { background: rgba(255,255,255,0.2); }
.entities-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: var(--cm-space-4); }
.entity-card { background: var(--cm-bg-primary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; cursor: pointer; transition: all 0.2s; }
.entity-card:hover { border-color: var(--cm-primary); box-shadow: 0 4px 12px rgba(0,0,0,0.08); transform: translateY(-2px); }
.card-top { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.entity-avatar { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 16px; flex-shrink: 0; }
.entity-meta { flex: 1; min-width: 0; }
.entity-name { font-size: 16px; font-weight: 600; color: var(--cm-text-primary); margin: 0 0 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.entity-type-badge { font-size: 12px; padding: 2px 8px; border-radius: 20px; background: var(--cm-bg-tertiary); color: var(--cm-text-secondary); }
.entity-desc { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.5; margin: 0 0 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.card-bottom { display: flex; align-items: center; gap: 8px; }
.relation-info { display: flex; align-items: center; gap: 4px; font-size: 13px; color: var(--cm-primary); font-weight: 500; }
.confidence-info { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--cm-text-tertiary); }
.confidence-bar { width: 60px; height: 4px; background: var(--cm-bg-tertiary); border-radius: 4px; overflow: hidden; }
.confidence-fill { height: 100%; background: var(--cm-primary-gradient); border-radius: 4px; }
.card-actions { margin-left: auto; display: flex; gap: 4px; }
.card-action-btn { width: 28px; height: 28px; border: none; background: transparent; color: var(--cm-text-tertiary); border-radius: 6px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.card-action-btn:hover { background: var(--cm-bg-tertiary); color: var(--cm-text-primary); }
.card-action-btn.danger:hover { background: rgba(239,68,68,0.1); color: #ef4444; }
.empty-state { text-align: center; padding: 60px 20px; grid-column: 1 / -1; }
.empty-icon { font-size: 48px; color: var(--cm-text-tertiary); margin-bottom: 12px; }
.empty-state p { color: var(--cm-text-secondary); margin-bottom: 16px; }
.graph-view-v2 { height: 100%; display: flex; flex-direction: column; }
.graph-toolbar { display: flex; justify-content: space-between; align-items: center; padding: var(--cm-space-3); background: var(--cm-bg-primary); border: 1px solid var(--cm-border); border-radius: 12px; margin-bottom: var(--cm-space-4); }
.toolbar-group { display: flex; align-items: center; gap: var(--cm-space-2); }
.tool-btn { width: 36px; height: 36px; border: none; background: transparent; color: var(--cm-text-secondary); border-radius: 8px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.tool-btn:hover { background: var(--cm-bg-tertiary); color: var(--cm-text-primary); }
.graph-stats { font-size: 13px; color: var(--cm-text-tertiary); }
.graph-canvas { flex: 1; background: var(--cm-bg-primary); border: 1px solid var(--cm-border); border-radius: 12px; min-height: 400px; }
.list-view { background: var(--cm-bg-primary); border: 1px solid var(--cm-border); border-radius: 12px; overflow: hidden; }
.list-header { display: grid; grid-template-columns: 2fr 1fr 100px 100px; gap: var(--cm-space-4); padding: 12px 20px; background: var(--cm-bg-secondary); border-bottom: 1px solid var(--cm-border); font-size: 12px; font-weight: 600; color: var(--cm-text-tertiary); }
.list-row { display: grid; grid-template-columns: 2fr 1fr 100px 100px; gap: var(--cm-space-4); padding: 12px 20px; align-items: center; border-bottom: 1px solid var(--cm-border); cursor: pointer; }
.list-row:hover { background: var(--cm-bg-secondary); }
.list-col.name { display: flex; align-items: center; gap: 10px; font-weight: 500; }
.entity-avatar-small { width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600; font-size: 14px; }
.type-badge { font-size: 12px; padding: 2px 10px; border-radius: 20px; background: var(--cm-bg-tertiary); color: var(--cm-text-secondary); }
.list-col.actions { display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border: none; background: transparent; color: var(--cm-text-tertiary); border-radius: 6px; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; }
.action-btn:hover { background: var(--cm-bg-tertiary); color: var(--cm-text-primary); }
.action-btn.danger:hover { background: rgba(239,68,68,0.1); color: #ef4444; }
.drawer-content { padding: var(--cm-space-5); }
.drawer-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.entity-avatar-large { width: 56px; height: 56px; border-radius: 14px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 24px; flex-shrink: 0; }
.drawer-header h2 { font-size: 20px; font-weight: 700; margin: 0; }
.type-badge-large { font-size: 13px; padding: 4px 12px; border-radius: 20px; background: var(--cm-bg-tertiary); color: var(--cm-text-secondary); margin-top: 4px; display: inline-block; }
.drawer-actions { margin-left: auto; display: flex; gap: 8px; }
.drawer-section { margin-bottom: 24px; }
.drawer-section h3 { font-size: 13px; font-weight: 600; color: var(--cm-text-secondary); margin-bottom: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
.drawer-section p { font-size: 15px; line-height: 1.6; color: var(--cm-text-primary); }
.empty-hint { font-size: 14px; color: var(--cm-text-tertiary); }
.relation-list { display: flex; flex-direction: column; gap: 8px; }
.relation-item { display: flex; align-items: center; gap: 10px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 8px; }
.relation-node { padding: 4px 10px; background: var(--cm-bg-primary); border-radius: 6px; font-size: 14px; font-weight: 500; }
.relation-node.is-source { border-left: 3px solid var(--cm-primary); }
.relation-node.is-target { border-left: 3px solid #10b981; }
.relation-arrow { display: flex; align-items: center; gap: 4px; color: var(--cm-text-tertiary); font-size: 12px; }
.relation-type { font-weight: 600; }
.properties-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.property-item { display: flex; flex-direction: column; gap: 2px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 8px; }
.property-key { font-size: 11px; color: var(--cm-text-tertiary); text-transform: uppercase; letter-spacing: 0.5px; }
.property-value { font-size: 14px; color: var(--cm-text-primary); font-weight: 500; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-width { grid-column: 1 / -1; }
.form-group label { font-size: 14px; font-weight: 500; color: var(--cm-text-secondary); }
.cm-input { width: 100%; padding: 8px 12px; border: 1px solid var(--cm-border); border-radius: 8px; background: var(--cm-bg-secondary); color: var(--cm-text-primary); font-size: 14px; }
.cm-input:focus { outline: none; border-color: var(--cm-primary); }
.cm-btn { display: inline-flex; align-items: center; gap: 6px; padding: 8px 16px; border: none; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; }
.cm-btn-primary { background: var(--cm-primary); color: white; }
.cm-btn-primary:hover { opacity: 0.9; }
.cm-btn-secondary { background: var(--cm-bg-tertiary); color: var(--cm-text-primary); }
.cm-btn-secondary:hover { opacity: 0.9; }
.cm-btn-danger { background: #ef4444; color: white; }
.cm-btn-danger:hover { opacity: 0.9; }
@media (max-width: 768px) {
  .header-content { flex-direction: column; gap: var(--cm-space-4); align-items: stretch; }
  .header-actions { flex-wrap: wrap; }
  .search-box { width: 100%; }
  .entities-grid { grid-template-columns: 1fr; }
}
</style>
