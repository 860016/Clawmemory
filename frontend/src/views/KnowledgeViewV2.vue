<template>
  <div class="knowledge-v2">
    <!-- Header -->
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
            <input 
              v-model="searchQuery" 
              :placeholder="$t('knowledge.searchPlaceholder')"
              class="search-input"
            />
            <kbd class="shortcut">⌘K</kbd>
          </div>
          
          <button class="cm-btn cm-btn-primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            <span>{{ $t('knowledge.addEntity') }}</span>
          </button>
          
          <button class="cm-btn cm-btn-secondary" @click="showRelationDialog = true">
            <el-icon><Link /></el-icon>
            <span>{{ $t('knowledge.addRelation') }}</span>
          </button>
        </div>
      </div>
      
      <!-- View Toggle -->
      <div class="view-tabs">
        <button 
          v-for="tab in viewTabs" 
          :key="tab.key"
          class="tab-btn"
          :class="{ active: currentView === tab.key }"
          @click="currentView = tab.key"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </button>
      </div>
    </header>

    <!-- Content -->
    <main class="knowledge-content">
      <!-- Grid View -->
      <div v-if="currentView === 'grid'" class="grid-view">
        <div class="category-filters">
          <button 
            v-for="cat in categories" 
            :key="cat.type"
            class="filter-chip"
            :class="{ active: selectedCategory === cat.type }"
            @click="selectedCategory = selectedCategory === cat.type ? '' : cat.type"
          >
            <span class="chip-icon">{{ cat.icon }}</span>
            <span>{{ cat.label }}</span>
            <span class="chip-count">{{ cat.count }}</span>
          </button>
        </div>
        
        <div class="entities-grid">
          <div 
            v-for="entity in filteredEntities" 
            :key="entity.id"
            class="entity-card-v2"
            :class="{ 'has-relations': getEntityRelations(entity.id).length > 0 }"
            @click="openEntityDetail(entity)"
          >
            <div class="card-header">
              <div class="entity-avatar" :style="{ background: getAvatarColor(entity.name) }">
                {{ entity.name.charAt(0).toUpperCase() }}
              </div>
              <div class="entity-meta">
                <h3 class="entity-name">{{ entity.name }}</h3>
                <span class="entity-type" :class="entity.entity_type">{{ entity.entity_type }}</span>
              </div>
              <div class="relation-count" v-if="getEntityRelations(entity.id).length">
                <el-icon><Connection /></el-icon>
                <span>{{ getEntityRelations(entity.id).length }}</span>
              </div>
            </div>
            
            <p class="entity-description">{{ entity.description || $t('knowledge.noDescription') }}</p>
            
            <div class="card-footer">
              <div class="confidence-bar" v-if="entity.confidence">
                <div class="confidence-fill" :style="{ width: entity.confidence * 100 + '%' }"></div>
              </div>
              <span class="confidence-label" v-if="entity.confidence">
                {{ Math.round(entity.confidence * 100) }}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Graph View -->
      <div v-if="currentView === 'graph'" class="graph-view-v2">
        <div class="graph-toolbar">
          <div class="toolbar-group">
            <button class="tool-btn" @click="resetGraph" title="Reset">
              <el-icon><RefreshRight /></el-icon>
            </button>
            <button class="tool-btn" @click="zoomIn" title="Zoom In">
              <el-icon><ZoomIn /></el-icon>
            </button>
            <button class="tool-btn" @click="zoomOut" title="Zoom Out">
              <el-icon><ZoomOut /></el-icon>
            </button>
          </div>
          <div class="toolbar-group">
            <span class="graph-stats">{{ graphData.nodes.length }} nodes · {{ graphData.edges.length }} edges</span>
          </div>
        </div>
        <div ref="graphContainer" class="graph-canvas"></div>
      </div>

      <!-- List View -->
      <div v-if="currentView === 'list'" class="list-view">
        <div class="list-header">
          <span class="list-col name">{{ $t('knowledge.name') }}</span>
          <span class="list-col type">{{ $t('knowledge.type') }}</span>
          <span class="list-col relations">{{ $t('knowledge.relations') }}</span>
          <span class="list-col confidence">{{ $t('knowledge.confidence') }}</span>
          <span class="list-col actions"></span>
        </div>
        <div class="list-body">
          <div 
            v-for="entity in filteredEntities" 
            :key="entity.id"
            class="list-row"
            @click="openEntityDetail(entity)"
          >
            <span class="list-col name">
              <div class="entity-avatar-small" :style="{ background: getAvatarColor(entity.name) }">
                {{ entity.name.charAt(0).toUpperCase() }}
              </div>
              {{ entity.name }}
            </span>
            <span class="list-col type">
              <span class="type-badge" :class="entity.entity_type">{{ entity.entity_type }}</span>
            </span>
            <span class="list-col relations">{{ getEntityRelations(entity.id).length }}</span>
            <span class="list-col confidence">
              <div class="confidence-bar-mini">
                <div class="confidence-fill" :style="{ width: (entity.confidence || 0) * 100 + '%' }"></div>
              </div>
            </span>
            <span class="list-col actions">
              <button class="action-btn" @click.stop="editEntity(entity)">
                <el-icon><Edit /></el-icon>
              </button>
              <button class="action-btn danger" @click.stop="deleteEntity(entity.id)">
                <el-icon><Delete /></el-icon>
              </button>
            </span>
          </div>
        </div>
      </div>
    </main>

    <!-- Entity Detail Drawer -->
    <el-drawer
      v-model="showDetailDrawer"
      :title="selectedEntity?.name"
      size="500px"
      class="entity-drawer"
    >
      <div v-if="selectedEntity" class="drawer-content">
        <div class="drawer-header">
          <div class="entity-avatar-large" :style="{ background: getAvatarColor(selectedEntity.name) }">
            {{ selectedEntity.name.charAt(0).toUpperCase() }}
          </div>
          <div>
            <h2>{{ selectedEntity.name }}</h2>
            <span class="type-badge-large" :class="selectedEntity.entity_type">{{ selectedEntity.entity_type }}</span>
          </div>
        </div>
        
        <div class="drawer-section">
          <h3>{{ $t('knowledge.description') }}</h3>
          <p>{{ selectedEntity.description || $t('knowledge.noDescription') }}</p>
        </div>
        
        <div class="drawer-section">
          <h3>{{ $t('knowledge.relations') }}</h3>
          <div class="relation-list-v2">
            <div 
              v-for="rel in getEntityRelations(selectedEntity.id)" 
              :key="rel.id"
              class="relation-item-v2"
            >
              <div class="relation-node" :class="{ 'is-source': rel.source_id === selectedEntity.id }">
                {{ getEntityName(rel.source_id) }}
              </div>
              <div class="relation-arrow">
                <span class="relation-type">{{ rel.relation_type }}</span>
                <el-icon><ArrowRight /></el-icon>
              </div>
              <div class="relation-node" :class="{ 'is-target': rel.target_id === selectedEntity.id }">
                {{ getEntityName(rel.target_id) }}
              </div>
            </div>
          </div>
        </div>
        
        <div class="drawer-section" v-if="selectedEntity.properties">
          <h3>{{ $t('knowledge.properties') }}</h3>
          <div class="properties-grid">
            <div v-for="(value, key) in JSON.parse(selectedEntity.properties)" :key="key" class="property-item">
              <span class="property-key">{{ key }}</span>
              <span class="property-value">{{ value }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-drawer>

    <!-- Create Entity Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="$t('knowledge.createEntity')"
      width="600px"
      class="modern-dialog"
    >
      <div class="form-grid">
        <div class="form-group">
          <label>{{ $t('knowledge.name') }}</label>
          <input v-model="newEntity.name" class="cm-input" :placeholder="$t('knowledge.namePlaceholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('knowledge.type') }}</label>
          <select v-model="newEntity.entity_type" class="cm-input">
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
          <textarea v-model="newEntity.description" class="cm-input" rows="4" :placeholder="$t('knowledge.descriptionPlaceholder')"></textarea>
        </div>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showCreateDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="createEntity">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Search, Plus, Link, Connection, Edit, Delete,
  RefreshRight, ZoomIn, ZoomOut, ArrowRight, Grid, List
} from '@element-plus/icons-vue'
import { useKnowledgeStore } from '@/stores/knowledge'

const { t } = useI18n()
const store = useKnowledgeStore()

// State
const currentView = ref('grid')
const searchQuery = ref('')
const selectedCategory = ref('')
const showDetailDrawer = ref(false)
const showCreateDialog = ref(false)
const showRelationDialog = ref(false)
const selectedEntity = ref<any>(null)
const newEntity = ref({ name: '', entity_type: 'person', description: '' })

// View tabs
const viewTabs = [
  { key: 'grid', label: t('knowledge.gridView'), icon: Grid },
  { key: 'graph', label: t('knowledge.graphView'), icon: Connection },
  { key: 'list', label: t('knowledge.listView'), icon: List },
]

// Data
const entities = computed(() => store.entities)
const relations = computed(() => store.relations)

const categories = computed(() => {
  const types = [...new Set(entities.value.map(e => e.entity_type))]
  const icons: Record<string, string> = {
    person: '👤',
    organization: '🏢',
    location: '📍',
    concept: '💡',
    technology: '⚡',
    event: '📅',
  }
  return types.map(type => ({
    type,
    label: t(`knowledge.types.${type}`),
    icon: icons[type] || '📄',
    count: entities.value.filter(e => e.entity_type === type).length,
  }))
})

const filteredEntities = computed(() => {
  let result = entities.value
  
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(e => 
      e.name.toLowerCase().includes(q) ||
      e.description?.toLowerCase().includes(q)
    )
  }
  
  if (selectedCategory.value) {
    result = result.filter(e => e.entity_type === selectedCategory.value)
  }
  
  return result
})

const graphData = computed(() => ({
  nodes: entities.value.map(e => ({
    id: String(e.id),
    label: e.name,
    group: e.entity_type,
    value: getEntityRelations(e.id).length + 1,
  })),
  edges: relations.value.map(r => ({
    from: String(r.source_id),
    to: String(r.target_id),
    label: r.relation_type,
  })),
}))

// Methods
function getEntityRelations(entityId: number) {
  return relations.value.filter(r => r.source_id === entityId || r.target_id === entityId)
}

function getEntityName(id: number) {
  return entities.value.find(e => e.id === id)?.name || 'Unknown'
}

function getAvatarColor(name: string) {
  const colors = [
    '#6366f1', '#8b5cf6', '#ec4899', '#f43f5e',
    '#f97316', '#eab308', '#22c55e', '#06b6d4',
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

function openEntityDetail(entity: any) {
  selectedEntity.value = entity
  showDetailDrawer.value = true
}

function createEntity() {
  store.createEntity(newEntity.value)
  showCreateDialog.value = false
  newEntity.value = { name: '', entity_type: 'person', description: '' }
}

function editEntity(entity: any) {
  selectedEntity.value = entity
  // TODO: Open edit dialog
}

function deleteEntity(id: number) {
  store.deleteEntity(id)
}

function resetGraph() {
  // TODO: Reset graph view
}

function zoomIn() {
  // TODO: Zoom in
}

function zoomOut() {
  // TODO: Zoom out
}

// Keyboard shortcut
onMounted(() => {
  document.addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      document.querySelector('.search-input')?.focus()
    }
  })
})
</script>

<style scoped>
.knowledge-v2 {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Header */
.knowledge-header {
  background: var(--cm-bg-primary);
  border-bottom: 1px solid var(--cm-border);
  padding: var(--cm-space-6);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--cm-space-4);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
}

.title-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--cm-radius-lg);
  background: var(--cm-primary-gradient);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-title h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--cm-text-primary);
}

.subtitle {
  font-size: 14px;
  color: var(--cm-text-secondary);
  margin-top: 2px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
}

.search-box {
  position: relative;
  width: 300px;
}

.search-icon {
  position: absolute;
  left: var(--cm-space-3);
  top: 50%;
  transform: translateY(-50%);
  color: var(--cm-text-tertiary);
}

.search-input {
  width: 100%;
  padding: var(--cm-space-2) var(--cm-space-3) var(--cm-space-2) 36px;
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-md);
  background: var(--cm-bg-secondary);
  color: var(--cm-text-primary);
  font-size: 14px;
  transition: all var(--cm-transition-fast);
}

.search-input:focus {
  outline: none;
  border-color: var(--cm-primary);
  background: var(--cm-bg-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.shortcut {
  position: absolute;
  right: var(--cm-space-2);
  top: 50%;
  transform: translateY(-50%);
  padding: 2px 6px;
  background: var(--cm-bg-tertiary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-sm);
  font-size: 11px;
  color: var(--cm-text-tertiary);
}

/* View Tabs */
.view-tabs {
  display: flex;
  gap: var(--cm-space-1);
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  padding: var(--cm-space-2) var(--cm-space-4);
  border: none;
  background: transparent;
  color: var(--cm-text-secondary);
  font-size: 14px;
  font-weight: 500;
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  transition: all var(--cm-transition-fast);
}

.tab-btn:hover {
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-primary);
}

.tab-btn.active {
  background: var(--cm-primary);
  color: white;
}

/* Content */
.knowledge-content {
  flex: 1;
  overflow: auto;
  padding: var(--cm-space-6);
}

/* Grid View */
.category-filters {
  display: flex;
  gap: var(--cm-space-2);
  margin-bottom: var(--cm-space-6);
  flex-wrap: wrap;
}

.filter-chip {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
  padding: var(--cm-space-2) var(--cm-space-3);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-full);
  background: var(--cm-bg-primary);
  color: var(--cm-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--cm-transition-fast);
}

.filter-chip:hover {
  border-color: var(--cm-primary-light);
  color: var(--cm-text-primary);
}

.filter-chip.active {
  background: var(--cm-primary);
  border-color: var(--cm-primary);
  color: white;
}

.chip-icon {
  font-size: 14px;
}

.chip-count {
  padding: 1px 6px;
  background: var(--cm-bg-tertiary);
  border-radius: var(--cm-radius-full);
  font-size: 11px;
  font-weight: 600;
}

.filter-chip.active .chip-count {
  background: rgba(255, 255, 255, 0.2);
}

/* Entity Cards */
.entities-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--cm-space-4);
}

.entity-card-v2 {
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  padding: var(--cm-space-5);
  cursor: pointer;
  transition: all var(--cm-transition-normal);
}

.entity-card-v2:hover {
  border-color: var(--cm-primary-light);
  box-shadow: var(--cm-shadow-lg);
  transform: translateY(-2px);
}

.entity-card-v2.has-relations {
  border-left: 3px solid var(--cm-primary);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  margin-bottom: var(--cm-space-3);
}

.entity-avatar {
  width: 40px;
  height: 40px;
  border-radius: var(--cm-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 16px;
}

.entity-meta {
  flex: 1;
}

.entity-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--cm-text-primary);
  margin-bottom: 2px;
}

.entity-type {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: var(--cm-radius-full);
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-secondary);
}

.relation-count {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--cm-primary);
  font-weight: 600;
}

.entity-description {
  font-size: 14px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin-bottom: var(--cm-space-3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
}

.confidence-bar {
  flex: 1;
  height: 4px;
  background: var(--cm-bg-tertiary);
  border-radius: var(--cm-radius-full);
  overflow: hidden;
}

.confidence-fill {
  height: 100%;
  background: var(--cm-primary-gradient);
  border-radius: var(--cm-radius-full);
  transition: width 0.3s ease;
}

.confidence-label {
  font-size: 12px;
  color: var(--cm-text-tertiary);
  font-weight: 600;
}

/* Graph View */
.graph-view-v2 {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.graph-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--cm-space-3);
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  margin-bottom: var(--cm-space-4);
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: var(--cm-space-2);
}

.tool-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: var(--cm-text-secondary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.tool-btn:hover {
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-primary);
}

.graph-stats {
  font-size: 13px;
  color: var(--cm-text-tertiary);
}

.graph-canvas {
  flex: 1;
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
}

/* List View */
.list-view {
  background: var(--cm-bg-primary);
  border: 1px solid var(--cm-border);
  border-radius: var(--cm-radius-lg);
  overflow: hidden;
}

.list-header {
  display: grid;
  grid-template-columns: 2fr 1fr 100px 120px 100px;
  gap: var(--cm-space-4);
  padding: var(--cm-space-3) var(--cm-space-5);
  background: var(--cm-bg-secondary);
  border-bottom: 1px solid var(--cm-border);
  font-size: 12px;
  font-weight: 600;
  color: var(--cm-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.list-row {
  display: grid;
  grid-template-columns: 2fr 1fr 100px 120px 100px;
  gap: var(--cm-space-4);
  padding: var(--cm-space-3) var(--cm-space-5);
  align-items: center;
  border-bottom: 1px solid var(--cm-border-light);
  cursor: pointer;
  transition: background var(--cm-transition-fast);
}

.list-row:hover {
  background: var(--cm-bg-secondary);
}

.list-col.name {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  font-weight: 500;
}

.entity-avatar-small {
  width: 32px;
  height: 32px;
  border-radius: var(--cm-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.type-badge {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: var(--cm-radius-full);
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-secondary);
}

.confidence-bar-mini {
  width: 80px;
  height: 4px;
  background: var(--cm-bg-tertiary);
  border-radius: var(--cm-radius-full);
  overflow: hidden;
}

.action-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--cm-text-tertiary);
  border-radius: var(--cm-radius-md);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all var(--cm-transition-fast);
}

.action-btn:hover {
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-primary);
}

.action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.1);
  color: var(--cm-error);
}

/* Drawer */
.entity-drawer :deep(.el-drawer__header) {
  margin-bottom: 0;
  padding: var(--cm-space-5);
  border-bottom: 1px solid var(--cm-border);
}

.drawer-content {
  padding: var(--cm-space-5);
}

.drawer-header {
  display: flex;
  align-items: center;
  gap: var(--cm-space-4);
  margin-bottom: var(--cm-space-6);
}

.entity-avatar-large {
  width: 64px;
  height: 64px;
  border-radius: var(--cm-radius-xl);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 28px;
}

.type-badge-large {
  font-size: 13px;
  padding: 4px 12px;
  border-radius: var(--cm-radius-full);
  background: var(--cm-bg-tertiary);
  color: var(--cm-text-secondary);
  margin-top: var(--cm-space-1);
  display: inline-block;
}

.drawer-section {
  margin-bottom: var(--cm-space-6);
}

.drawer-section h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text-secondary);
  margin-bottom: var(--cm-space-3);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.drawer-section p {
  font-size: 15px;
  line-height: 1.6;
  color: var(--cm-text-primary);
}

/* Relation List */
.relation-list-v2 {
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-3);
}

.relation-item-v2 {
  display: flex;
  align-items: center;
  gap: var(--cm-space-3);
  padding: var(--cm-space-3);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
}

.relation-node {
  padding: var(--cm-space-2) var(--cm-space-3);
  background: var(--cm-bg-primary);
  border-radius: var(--cm-radius-md);
  font-size: 14px;
  font-weight: 500;
}

.relation-node.is-source {
  border-left: 3px solid var(--cm-primary);
}

.relation-node.is-target {
  border-left: 3px solid var(--cm-secondary);
}

.relation-arrow {
  display: flex;
  align-items: center;
  gap: var(--cm-space-1);
  color: var(--cm-text-tertiary);
}

.relation-type {
  font-size: 12px;
  font-weight: 600;
}

/* Properties Grid */
.properties-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--cm-space-3);
}

.property-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--cm-space-3);
  background: var(--cm-bg-secondary);
  border-radius: var(--cm-radius-md);
}

.property-key {
  font-size: 12px;
  color: var(--cm-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.property-value {
  font-size: 14px;
  color: var(--cm-text-primary);
  font-weight: 500;
}

/* Form */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--cm-space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--cm-space-2);
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  font-size: 14px;
  font-weight: 500;
  color: var(--cm-text-secondary);
}

.modern-dialog :deep(.el-dialog__header) {
  padding: var(--cm-space-5);
  border-bottom: 1px solid var(--cm-border);
}

.modern-dialog :deep(.el-dialog__body) {
  padding: var(--cm-space-5);
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: var(--cm-space-4) var(--cm-space-5);
  border-top: 1px solid var(--cm-border);
}

/* Responsive */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: var(--cm-space-4);
    align-items: stretch;
  }
  
  .header-actions {
    flex-wrap: wrap;
  }
  
  .search-box {
    width: 100%;
  }
  
  .entities-grid {
    grid-template-columns: 1fr;
  }
  
  .list-header,
  .list-row {
    grid-template-columns: 1fr auto;
  }
  
  .list-col.type,
  .list-col.relations,
  .list-col.confidence {
    display: none;
  }
}
</style>
