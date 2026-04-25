<template>
  <div class="knowledge-page">
    <header class="page-header">
      <div class="header-left">
        <h1>🕸️ {{ $t('knowledge.title') }}</h1>
        <p class="subtitle">{{ entities.length }} {{ $t('knowledge.entities') }} · {{ relations.length }} {{ $t('knowledge.relations') }}</p>
      </div>
      <div class="header-right">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input v-model="searchQuery" :placeholder="$t('knowledge.searchPlaceholder')" class="search-input" />
        </div>
        <button class="cm-btn cm-btn-primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          {{ $t('knowledge.addEntity') }}
        </button>
        <button class="cm-btn cm-btn-secondary" @click="showRelationDialog = true">
          <el-icon><Connection /></el-icon>
          {{ $t('knowledge.addRelation') }}
        </button>
      </div>
    </header>

    <div class="category-bar">
      <button class="cat-btn" :class="{ active: !selectedCategory }" @click="selectedCategory = ''">
        {{ $t('knowledge.allCategories') }}
      </button>
      <button v-for="cat in categoryStats" :key="cat.type" class="cat-btn" :class="{ active: selectedCategory === cat.type }" @click="selectedCategory = cat.type">
        <span class="cat-icon">{{ cat.icon }}</span>
        <span>{{ cat.label }}</span>
        <span class="cat-count">{{ cat.count }}</span>
      </button>
    </div>

    <main class="page-content">
      <div v-if="loading" class="loading-state">
        <el-icon class="spin"><Loading /></el-icon>
        <p>{{ $t('common.loading') }}</p>
      </div>

      <div v-else-if="error" class="error-state">
        <el-icon><Warning /></el-icon>
        <p>{{ error }}</p>
        <button class="cm-btn cm-btn-primary" @click="loadData">{{ $t('common.retry') }}</button>
      </div>

      <div v-else class="entities-grid">
        <div v-for="entity in filteredEntities" :key="entity.id" class="entity-card" @click="openDetail(entity)">
          <div class="card-header">
            <div class="entity-avatar" :style="{ background: getAvatarColor(entity.name) }">
              {{ entity.name?.charAt(0)?.toUpperCase() || '?' }}
            </div>
            <div class="entity-info">
              <h3 class="entity-name">{{ entity.name }}</h3>
              <span class="entity-type" :class="entity.entity_type">{{ getTypeLabel(entity.entity_type) }}</span>
            </div>
          </div>
          <p class="entity-desc">{{ entity.description || $t('knowledge.noDescription') }}</p>
          <div class="card-footer">
            <span class="relation-count" v-if="getEntityRelations(entity.id).length">
              <el-icon><Connection /></el-icon>
              {{ getEntityRelations(entity.id).length }} {{ $t('knowledge.relations') }}
            </span>
            <div class="card-actions" @click.stop>
              <button class="action-btn" @click="openEditDialog(entity)" :title="$t('common.edit')">
                <el-icon><Edit /></el-icon>
              </button>
              <button class="action-btn danger" @click="confirmDelete(entity.id)" :title="$t('common.delete')">
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </div>
        </div>

        <div v-if="!filteredEntities.length && !loading" class="empty-state">
          <div class="empty-icon">◇</div>
          <p>{{ $t('knowledge.emptyGraph') }}</p>
          <button class="cm-btn cm-btn-primary" @click="openCreateDialog">{{ $t('knowledge.addEntity') }}</button>
        </div>
      </div>
    </main>

    <el-drawer v-model="detailVisible" :title="detailEntity?.name" size="400px" direction="rtl">
      <div v-if="detailEntity" class="detail-content">
        <div class="detail-header">
          <div class="detail-avatar" :style="{ background: getAvatarColor(detailEntity.name) }">
            {{ detailEntity.name?.charAt(0)?.toUpperCase() || '?' }}
          </div>
          <div class="detail-meta">
            <h2>{{ detailEntity.name }}</h2>
            <span class="detail-type" :class="detailEntity.entity_type">{{ getTypeLabel(detailEntity.entity_type) }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>{{ $t('knowledge.description') }}</h3>
          <p>{{ detailEntity.description || $t('knowledge.noDescription') }}</p>
        </div>
        <div class="detail-section" v-if="getEntityRelations(detailEntity.id).length">
          <h3>{{ $t('knowledge.relations') }}</h3>
          <div class="relation-list">
            <div v-for="r in getEntityRelations(detailEntity.id)" :key="r.id" class="relation-item">
              <span class="relation-node" :class="{ source: r.source_id === detailEntity.id }">
                {{ getEntityName(r.source_id) }}
              </span>
              <span class="relation-arrow">→ {{ r.relation_type }} →</span>
              <span class="relation-node" :class="{ target: r.target_id === detailEntity.id }">
                {{ getEntityName(r.target_id) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="showCreateDialog" :title="$t('knowledge.addEntity')" width="480px">
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
      <div class="form-group">
        <label>{{ $t('knowledge.description') }}</label>
        <textarea v-model="formData.description" class="cm-input" rows="3" :placeholder="$t('knowledge.descriptionPlaceholder')"></textarea>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showCreateDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="createEntity">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditDialog" :title="$t('knowledge.editEntity')" width="480px">
      <div class="form-group">
        <label>{{ $t('knowledge.entityName') }}</label>
        <input v-model="formData.name" class="cm-input" />
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
      <div class="form-group">
        <label>{{ $t('knowledge.description') }}</label>
        <textarea v-model="formData.description" class="cm-input" rows="3"></textarea>
      </div>
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showEditDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="updateEntity">{{ $t('common.save') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRelationDialog" :title="$t('knowledge.addRelation')" width="480px">
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
      <template #footer>
        <button class="cm-btn cm-btn-secondary" @click="showRelationDialog = false">{{ $t('common.cancel') }}</button>
        <button class="cm-btn cm-btn-primary" @click="createRelation">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Connection, Edit, Delete, Loading, Warning } from '@element-plus/icons-vue'
import axios from '../api/client'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const entities = ref<any[]>([])
const relations = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const searchQuery = ref('')
const selectedCategory = ref('')
const detailVisible = ref(false)
const detailEntity = ref<any>(null)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showRelationDialog = ref(false)
const editingEntity = ref<any>(null)

const formData = ref({ name: '', entity_type: 'concept', description: '' })
const relationForm = ref({ source_id: 0, target_id: 0, relation_type: '' })

const typeConfig: Record<string, { icon: string; label: string }> = {
  person: { icon: '👤', label: t('knowledge.types.person') },
  organization: { icon: '🏢', label: t('knowledge.types.organization') },
  location: { icon: '📍', label: t('knowledge.types.location') },
  concept: { icon: '💡', label: t('knowledge.types.concept') },
  technology: { icon: '🔧', label: t('knowledge.types.technology') },
  event: { icon: '📅', label: t('knowledge.types.event') },
}

const categoryStats = computed(() => {
  const stats: Record<string, { type: string; icon: string; label: string; count: number }> = {}
  entities.value.forEach(e => {
    if (!stats[e.entity_type]) {
      stats[e.entity_type] = { type: e.entity_type, ...typeConfig[e.entity_type] || { icon: '◇', label: e.entity_type }, count: 0 }
    }
    stats[e.entity_type].count++
  })
  return Object.values(stats)
})

const filteredEntities = computed(() => {
  let result = entities.value
  if (selectedCategory.value) {
    result = result.filter(e => e.entity_type === selectedCategory.value)
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(e => e.name?.toLowerCase().includes(q) || e.description?.toLowerCase().includes(q))
  }
  return result
})

function getTypeLabel(type: string) {
  return typeConfig[type]?.label || type
}

function getAvatarColor(name: string) {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f43f5e', '#f97316', '#eab308', '#22c55e', '#14b8a6', '#0ea5e9', '#3b82f6']
  const idx = name?.charCodeAt(0) % colors.length || 0
  return colors[idx]
}

function getEntityRelations(entityId: number) {
  return relations.value.filter(r => r.source_id === entityId || r.target_id === entityId)
}

function getEntityName(id: number) {
  return entities.value.find(e => e.id === id)?.name || 'Unknown'
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [entRes, relRes] = await Promise.all([
      axios.get('/knowledge/entities', { params: { page: 1, size: 200 }, _silent: true } as any),
      axios.get('/knowledge/relations', { params: { page: 1, size: 200 }, _silent: true } as any),
    ])
    entities.value = entRes.data.items || entRes.data || []
    relations.value = relRes.data.items || relRes.data || []
  } catch (e: any) {
    const msg = e.response?.data?.error || e.response?.data?.detail || t('common.loadFailed')
    error.value = msg
  } finally {
    loading.value = false
  }
}

function openDetail(entity: any) {
  detailEntity.value = entity
  detailVisible.value = true
}

function openCreateDialog() {
  formData.value = { name: '', entity_type: 'concept', description: '' }
  showCreateDialog.value = true
}

function openEditDialog(entity: any) {
  editingEntity.value = entity
  formData.value = { name: entity.name, entity_type: entity.entity_type, description: entity.description || '' }
  showEditDialog.value = true
}

async function createEntity() {
  if (!formData.value.name) {
    ElMessage.warning(t('knowledge.nameRequired'))
    return
  }
  try {
    const { data } = await axios.post('/knowledge/entities', formData.value)
    entities.value.unshift(data)
    showCreateDialog.value = false
    ElMessage.success(t('common.created'))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.createFailed'))
  }
}

async function updateEntity() {
  if (!editingEntity.value) return
  try {
    const { data } = await axios.put(`/knowledge/entities/${editingEntity.value.id}`, formData.value)
    const idx = entities.value.findIndex(e => e.id === editingEntity.value.id)
    if (idx >= 0) entities.value[idx] = data
    showEditDialog.value = false
    ElMessage.success(t('common.saved'))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.saveFailed'))
  }
}

async function confirmDelete(id: number) {
  try {
    await ElMessageBox.confirm(t('knowledge.confirmDelete'), t('common.confirm'), { type: 'warning' })
    await axios.delete(`/knowledge/entities/${id}`)
    entities.value = entities.value.filter(e => e.id !== id)
    relations.value = relations.value.filter(r => r.source_id !== id && r.target_id !== id)
    ElMessage.success(t('common.deleted'))
  } catch {}
}

async function createRelation() {
  if (!relationForm.value.source_id || !relationForm.value.target_id || !relationForm.value.relation_type) {
    ElMessage.warning(t('knowledge.fillRequired'))
    return
  }
  try {
    const { data } = await axios.post('/knowledge/relations', relationForm.value)
    relations.value.unshift(data)
    showRelationDialog.value = false
    ElMessage.success(t('common.created'))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.createFailed'))
  }
}

onMounted(loadData)
</script>

<style scoped>
.knowledge-page { height: 100%; display: flex; flex-direction: column; background: var(--cm-bg-secondary); }

.page-header { background: var(--cm-bg-primary); padding: 20px 24px; border-bottom: 1px solid var(--cm-border); display: flex; justify-content: space-between; align-items: center; }
.header-left h1 { font-size: 22px; font-weight: 700; color: var(--cm-text); margin: 0; }
.subtitle { font-size: 13px; color: var(--cm-text-muted); margin: 4px 0 0; }
.header-right { display: flex; align-items: center; gap: 12px; }
.search-box { position: relative; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); color: var(--cm-text-muted); }
.search-input { width: 200px; padding: 8px 12px 8px 32px; border: 1px solid var(--cm-border); border-radius: 8px; background: var(--cm-bg); color: var(--cm-text); font-size: 14px; }
.search-input:focus { outline: none; border-color: var(--cm-primary); }

.category-bar { padding: 12px 24px; display: flex; gap: 8px; flex-wrap: wrap; background: var(--cm-bg-primary); border-bottom: 1px solid var(--cm-border); }
.cat-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border: 1px solid var(--cm-border); border-radius: 20px; background: var(--cm-bg); color: var(--cm-text-secondary); font-size: 13px; cursor: pointer; }
.cat-btn:hover { border-color: var(--cm-primary); }
.cat-btn.active { background: var(--cm-primary); border-color: var(--cm-primary); color: white; }
.cat-icon { font-size: 14px; }
.cat-count { padding: 1px 6px; background: var(--cm-bg-tertiary); border-radius: 10px; font-size: 11px; }
.cat-btn.active .cat-count { background: rgba(255,255,255,0.2); }

.page-content { flex: 1; padding: 24px; overflow: auto; }

.loading-state, .error-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px; color: var(--cm-text-muted); }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-state { color: #f43f5e; }

.entities-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }

.entity-card { background: var(--cm-bg-primary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 16px; cursor: pointer; transition: all 0.2s; }
.entity-card:hover { border-color: var(--cm-primary); box-shadow: 0 4px 12px rgba(0,0,0,0.08); transform: translateY(-2px); }

.card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.entity-avatar { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 16px; }
.entity-info { flex: 1; min-width: 0; }
.entity-name { font-size: 15px; font-weight: 600; color: var(--cm-text); margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.entity-type { font-size: 12px; padding: 2px 8px; border-radius: 12px; background: var(--cm-bg-tertiary); color: var(--cm-text-secondary); }

.entity-desc { font-size: 13px; color: var(--cm-text-secondary); line-height: 1.5; margin: 0 0 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.card-footer { display: flex; align-items: center; gap: 8px; }
.relation-count { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--cm-primary); }
.card-actions { margin-left: auto; display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border: none; background: transparent; color: var(--cm-text-muted); border-radius: 6px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.action-btn:hover { background: var(--cm-bg-tertiary); color: var(--cm-text); }
.action-btn.danger:hover { background: rgba(239,68,68,0.1); color: #ef4444; }

.empty-state { text-align: center; padding: 60px; grid-column: 1 / -1; }
.empty-icon { font-size: 48px; color: var(--cm-text-muted); margin-bottom: 12px; }
.empty-state p { color: var(--cm-text-secondary); margin-bottom: 16px; }

.detail-content { padding: 20px; }
.detail-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.detail-avatar { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 20px; }
.detail-meta h2 { font-size: 18px; font-weight: 700; margin: 0; }
.detail-type { font-size: 12px; padding: 3px 10px; border-radius: 12px; background: var(--cm-bg-tertiary); color: var(--cm-text-secondary); margin-top: 4px; display: inline-block; }
.detail-section { margin-bottom: 20px; }
.detail-section h3 { font-size: 13px; font-weight: 600; color: var(--cm-text-muted); margin-bottom: 8px; }
.detail-section p { font-size: 14px; color: var(--cm-text-secondary); line-height: 1.6; }
.relation-list { display: flex; flex-direction: column; gap: 8px; }
.relation-item { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: var(--cm-bg); border-radius: 8px; font-size: 13px; }
.relation-node { padding: 3px 8px; background: var(--cm-bg-primary); border-radius: 6px; font-weight: 500; }
.relation-node.source { border-left: 3px solid var(--cm-primary); }
.relation-node.target { border-left: 3px solid #10b981; }
.relation-arrow { color: var(--cm-text-muted); }

.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; font-weight: 600; color: var(--cm-text-secondary); margin-bottom: 6px; }
.cm-input { width: 100%; padding: 10px 12px; border: 1px solid var(--cm-border); border-radius: 8px; background: var(--cm-bg); color: var(--cm-text); font-size: 14px; }
.cm-input:focus { outline: none; border-color: var(--cm-primary); }
.cm-btn { display: inline-flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1px solid var(--cm-border); border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; }
.cm-btn-primary { background: var(--cm-primary); border-color: var(--cm-primary); color: white; }
.cm-btn-secondary { background: var(--cm-bg); color: var(--cm-text-secondary); }

@media (max-width: 768px) {
  .page-header { flex-direction: column; gap: 12px; align-items: stretch; }
  .header-right { flex-wrap: wrap; }
  .search-input { width: 100%; }
  .entities-grid { grid-template-columns: 1fr; }
}
</style>