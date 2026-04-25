<template>
  <div class="knowledge-page">
    <div class="page-hero">
      <div class="hero-content">
        <h1>🕸️ {{ $t('knowledge.title') }}</h1>
        <p>{{ entities.length }} {{ $t('knowledge.entities') }} · {{ relations.length }} {{ $t('knowledge.relations') }}</p>
      </div>
      <div class="hero-actions">
        <div class="search-box">
          <el-icon><Search /></el-icon>
          <input v-model="searchQuery" :placeholder="$t('knowledge.searchPlaceholder')" />
        </div>
        <button class="btn-primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon> {{ $t('knowledge.addEntity') }}
        </button>
      </div>
    </div>

    <div class="filter-bar">
      <button class="filter-chip" :class="{ active: !selectedType }" @click="selectedType = ''">
        {{ $t('knowledge.allCategories') }}
      </button>
      <button v-for="cat in categories" :key="cat.type" class="filter-chip" :class="{ active: selectedType === cat.type }" @click="selectedType = cat.type">
        {{ cat.icon }} {{ cat.label }} <span class="chip-count">{{ cat.count }}</span>
      </button>
    </div>

    <div class="content-area">
      <div v-if="loading" class="state-block">
        <el-icon class="spin"><Loading /></el-icon>
        <p>{{ $t('common.loading') }}</p>
      </div>

      <div v-else-if="error" class="state-block error">
        <div class="error-icon">⚠️</div>
        <p>{{ error }}</p>
        <button class="btn-primary" @click="loadData">{{ $t('common.retry') }}</button>
      </div>

      <div v-else-if="filteredEntities.length" class="cards-grid">
        <div v-for="entity in filteredEntities" :key="entity.id" class="entity-card" @click="openDetail(entity)">
          <div class="card-top">
            <div class="card-avatar" :style="{ background: getColor(entity.name) }">
              {{ entity.name?.charAt(0)?.toUpperCase() || '?' }}
            </div>
            <div class="card-title-area">
              <h3>{{ entity.name }}</h3>
              <span class="type-badge" :class="entity.entity_type">{{ getTypeLabel(entity.entity_type) }}</span>
            </div>
            <div class="card-menu" @click.stop>
              <el-dropdown trigger="click">
                <button class="menu-btn">⋯</button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="openEditDialog(entity)">
                      <el-icon><Edit /></el-icon> {{ $t('common.edit') }}
                    </el-dropdown-item>
                    <el-dropdown-item @click="confirmDelete(entity.id)" class="danger-item">
                      <el-icon><Delete /></el-icon> {{ $t('common.delete') }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
          <p class="card-desc">{{ entity.description || $t('knowledge.noDescription') }}</p>
          <div class="card-bottom">
            <span v-if="getEntityRelations(entity.id).length" class="rel-tag">
              <el-icon><Connection /></el-icon> {{ getEntityRelations(entity.id).length }}
            </span>
            <span class="card-time">{{ formatTime(entity.updated_at) }}</span>
          </div>
        </div>
      </div>

      <div v-else class="state-block empty">
        <div class="empty-visual">
          <div class="empty-circle">🕸️</div>
        </div>
        <h3>{{ $t('knowledge.emptyGraph') }}</h3>
        <p>{{ $t('knowledge.emptyHint') }}</p>
        <button class="btn-primary" @click="openCreateDialog">{{ $t('knowledge.addEntity') }}</button>
      </div>
    </div>

    <el-drawer v-model="detailVisible" :size="440" direction="rtl" :show-close="true">
      <template #header>
        <div class="drawer-header">
          <div class="drawer-avatar" :style="{ background: getColor(detailEntity?.name) }">
            {{ detailEntity?.name?.charAt(0)?.toUpperCase() || '?' }}
          </div>
          <div>
            <h2>{{ detailEntity?.name }}</h2>
            <span class="type-badge" :class="detailEntity?.entity_type">{{ getTypeLabel(detailEntity?.entity_type) }}</span>
          </div>
        </div>
      </template>
      <div v-if="detailEntity" class="drawer-body">
        <div class="drawer-section">
          <h4>{{ $t('knowledge.description') }}</h4>
          <p class="desc-text">{{ detailEntity.description || $t('knowledge.noDescription') }}</p>
        </div>
        <div class="drawer-section" v-if="getEntityRelations(detailEntity.id).length">
          <h4>{{ $t('knowledge.relations') }} ({{ getEntityRelations(detailEntity.id).length }})</h4>
          <div class="rel-list">
            <div v-for="r in getEntityRelations(detailEntity.id)" :key="r.id" class="rel-item">
              <span class="rel-node">{{ getEntityName(r.source_id) }}</span>
              <span class="rel-type">{{ r.relation_type }}</span>
              <span class="rel-node">{{ getEntityName(r.target_id) }}</span>
            </div>
          </div>
        </div>
        <div class="drawer-section">
          <h4>{{ $t('knowledge.metaInfo') }}</h4>
          <div class="meta-grid">
            <div class="meta-item">
              <span class="meta-label">ID</span>
              <span class="meta-value">{{ detailEntity.id }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ $t('knowledge.extractMethod') }}</span>
              <span class="meta-value">{{ detailEntity.extract_method || 'manual' }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ $t('knowledge.confidence') }}</span>
              <span class="meta-value">{{ ((detailEntity.confidence || 1.0) * 100).toFixed(0) }}%</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ $t('knowledge.updatedAt') }}</span>
              <span class="meta-value">{{ formatTime(detailEntity.updated_at) }}</span>
            </div>
          </div>
        </div>
        <div class="drawer-actions">
          <button class="btn-secondary" @click="openEditDialog(detailEntity); detailVisible = false">
            <el-icon><Edit /></el-icon> {{ $t('common.edit') }}
          </button>
          <button class="btn-danger" @click="confirmDelete(detailEntity.id); detailVisible = false">
            <el-icon><Delete /></el-icon> {{ $t('common.delete') }}
          </button>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="showCreateDialog" :title="$t('knowledge.addEntity')" width="480px" :close-on-click-modal="false">
      <div class="form-field">
        <label>{{ $t('knowledge.entityName') }} *</label>
        <input v-model="formData.name" class="field-input" :placeholder="$t('knowledge.entityNamePlaceholder')" />
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.entityType') }}</label>
        <select v-model="formData.entity_type" class="field-input">
          <option value="concept">💡 {{ $t('knowledge.types.concept') }}</option>
          <option value="person">👤 {{ $t('knowledge.types.person') }}</option>
          <option value="organization">🏢 {{ $t('knowledge.types.organization') }}</option>
          <option value="technology">🔧 {{ $t('knowledge.types.technology') }}</option>
          <option value="location">📍 {{ $t('knowledge.types.location') }}</option>
          <option value="event">📅 {{ $t('knowledge.types.event') }}</option>
        </select>
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.description') }}</label>
        <textarea v-model="formData.description" class="field-input" rows="3" :placeholder="$t('knowledge.descriptionPlaceholder')"></textarea>
      </div>
      <template #footer>
        <button class="btn-secondary" @click="showCreateDialog = false">{{ $t('common.cancel') }}</button>
        <button class="btn-primary" @click="createEntity">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEditDialog" :title="$t('knowledge.editEntity')" width="480px" :close-on-click-modal="false">
      <div class="form-field">
        <label>{{ $t('knowledge.entityName') }} *</label>
        <input v-model="formData.name" class="field-input" />
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.entityType') }}</label>
        <select v-model="formData.entity_type" class="field-input">
          <option value="concept">💡 {{ $t('knowledge.types.concept') }}</option>
          <option value="person">👤 {{ $t('knowledge.types.person') }}</option>
          <option value="organization">🏢 {{ $t('knowledge.types.organization') }}</option>
          <option value="technology">🔧 {{ $t('knowledge.types.technology') }}</option>
          <option value="location">📍 {{ $t('knowledge.types.location') }}</option>
          <option value="event">📅 {{ $t('knowledge.types.event') }}</option>
        </select>
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.description') }}</label>
        <textarea v-model="formData.description" class="field-input" rows="3"></textarea>
      </div>
      <template #footer>
        <button class="btn-secondary" @click="showEditDialog = false">{{ $t('common.cancel') }}</button>
        <button class="btn-primary" @click="updateEntity">{{ $t('common.save') }}</button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRelationDialog" :title="$t('knowledge.addRelation')" width="480px" :close-on-click-modal="false">
      <div class="form-field">
        <label>{{ $t('knowledge.sourceEntity') }}</label>
        <select v-model="relationForm.source_id" class="field-input">
          <option v-for="e in entities" :key="e.id" :value="e.id">{{ e.name }}</option>
        </select>
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.relationType') }}</label>
        <input v-model="relationForm.relation_type" class="field-input" :placeholder="$t('knowledge.relationTypePlaceholder')" />
      </div>
      <div class="form-field">
        <label>{{ $t('knowledge.targetEntity') }}</label>
        <select v-model="relationForm.target_id" class="field-input">
          <option v-for="e in entities" :key="e.id" :value="e.id">{{ e.name }}</option>
        </select>
      </div>
      <template #footer>
        <button class="btn-secondary" @click="showRelationDialog = false">{{ $t('common.cancel') }}</button>
        <button class="btn-primary" @click="createRelation">{{ $t('common.create') }}</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Connection, Edit, Delete, Loading } from '@element-plus/icons-vue'
import axios from '../api/client'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const entities = ref<any[]>([])
const relations = ref<any[]>([])
const loading = ref(true)
const error = ref('')
const searchQuery = ref('')
const selectedType = ref('')
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

const categories = computed(() => {
  const stats: Record<string, { type: string; icon: string; label: string; count: number }> = {}
  entities.value.forEach(e => {
    if (!stats[e.entity_type]) {
      const cfg = typeConfig[e.entity_type] || { icon: '◇', label: e.entity_type }
      stats[e.entity_type] = { type: e.entity_type, ...cfg, count: 0 }
    }
    stats[e.entity_type].count++
  })
  return Object.values(stats)
})

const filteredEntities = computed(() => {
  let result = entities.value
  if (selectedType.value) {
    result = result.filter(e => e.entity_type === selectedType.value)
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(e =>
      e.name?.toLowerCase().includes(q) ||
      e.description?.toLowerCase().includes(q)
    )
  }
  return result
})

function getTypeLabel(type: string) {
  return typeConfig[type]?.label || type
}

function getColor(name: string) {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f43f5e', '#f97316', '#eab308', '#22c55e', '#14b8a6', '#0ea5e9', '#3b82f6']
  const idx = name?.charCodeAt(0) % colors.length || 0
  return colors[idx]
}

function getEntityRelations(entityId: number) {
  return relations.value.filter(r => r.source_id === entityId || r.target_id === entityId)
}

function getEntityName(id: number) {
  return entities.value.find(e => e.id === id)?.name || '?'
}

function formatTime(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [entRes, relRes] = await Promise.all([
      axios.get('/knowledge/entities', { params: { page: 1, size: 500 } }),
      axios.get('/knowledge/relations', { params: { page: 1, size: 500 } }),
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
.knowledge-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--cm-bg-secondary, #f5f5f5);
}

.page-hero {
  background: var(--cm-bg-primary, #fff);
  padding: 24px 28px;
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}
.hero-content h1 {
  font-size: 24px;
  font-weight: 800;
  color: var(--cm-text, #1a1a1a);
  margin: 0;
}
.hero-content p {
  font-size: 13px;
  color: var(--cm-text-muted, #999);
  margin: 4px 0 0;
}
.hero-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}
.search-box .el-icon {
  position: absolute;
  left: 10px;
  color: var(--cm-text-muted, #999);
  font-size: 14px;
}
.search-box input {
  width: 220px;
  padding: 8px 12px 8px 32px;
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 10px;
  background: var(--cm-bg, #fafafa);
  color: var(--cm-text, #1a1a1a);
  font-size: 14px;
  transition: all 0.2s;
}
.search-box input:focus {
  outline: none;
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.filter-bar {
  padding: 14px 28px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  background: var(--cm-bg-primary, #fff);
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
}
.filter-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 20px;
  background: var(--cm-bg, #fafafa);
  color: var(--cm-text-secondary, #666);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.filter-chip:hover {
  border-color: var(--cm-primary, #6366f1);
  color: var(--cm-primary, #6366f1);
}
.filter-chip.active {
  background: var(--cm-primary, #6366f1);
  border-color: var(--cm-primary, #6366f1);
  color: #fff;
}
.chip-count {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(0,0,0,0.06);
}
.filter-chip.active .chip-count {
  background: rgba(255,255,255,0.25);
}

.content-area {
  flex: 1;
  padding: 24px 28px;
  overflow: auto;
}

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--cm-text-muted, #999);
  gap: 12px;
}
.state-block.error { color: #ef4444; }
.state-block .spin { animation: spin 1s linear infinite; font-size: 32px; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-icon { font-size: 40px; }
.state-block p { font-size: 14px; }

.state-block.empty {
  padding: 100px 20px;
}
.empty-visual {
  margin-bottom: 16px;
}
.empty-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(99,102,241,0.1), rgba(139,92,246,0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  margin: 0 auto;
}
.state-block.empty h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--cm-text, #1a1a1a);
  margin: 0;
}
.state-block.empty p {
  color: var(--cm-text-secondary, #666);
  margin-bottom: 16px;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.entity-card {
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 14px;
  padding: 18px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}
.entity-card:hover {
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 8px 24px rgba(99,102,241,0.12);
  transform: translateY(-3px);
}

.card-top {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.card-avatar {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 17px;
  flex-shrink: 0;
}
.card-title-area {
  flex: 1;
  min-width: 0;
}
.card-title-area h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text, #1a1a1a);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.type-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--cm-bg-tertiary, #f0f0f0);
  color: var(--cm-text-secondary, #666);
  display: inline-block;
  margin-top: 3px;
}
.type-badge.concept { background: #eef2ff; color: #6366f1; }
.type-badge.person { background: #fef3c7; color: #d97706; }
.type-badge.organization { background: #dbeafe; color: #2563eb; }
.type-badge.technology { background: #d1fae5; color: #059669; }
.type-badge.location { background: #fce7f3; color: #db2777; }
.type-badge.event { background: #fee2e2; color: #dc2626; }

.card-menu {
  flex-shrink: 0;
}
.menu-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--cm-text-muted, #999);
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.menu-btn:hover {
  background: var(--cm-bg-tertiary, #f0f0f0);
}

.card-desc {
  font-size: 13px;
  color: var(--cm-text-secondary, #666);
  line-height: 1.6;
  margin: 0 0 14px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.rel-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--cm-primary, #6366f1);
  background: rgba(99,102,241,0.08);
  padding: 3px 8px;
  border-radius: 6px;
}
.card-time {
  font-size: 11px;
  color: var(--cm-text-muted, #999);
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: var(--cm-primary, #6366f1);
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}
.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: var(--cm-bg, #fafafa);
  color: var(--cm-text, #1a1a1a);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-secondary:hover {
  border-color: var(--cm-primary, #6366f1);
}
.btn-danger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: rgba(239,68,68,0.08);
  color: #ef4444;
  border: 1px solid rgba(239,68,68,0.2);
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
}
.btn-danger:hover {
  background: rgba(239,68,68,0.15);
}

.drawer-header {
  display: flex;
  align-items: center;
  gap: 14px;
}
.drawer-avatar {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 20px;
}
.drawer-header h2 {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
}

.drawer-body {
  padding: 8px 0;
}
.drawer-section {
  margin-bottom: 24px;
}
.drawer-section h4 {
  font-size: 12px;
  font-weight: 600;
  color: var(--cm-text-muted, #999);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0 0 10px;
}
.desc-text {
  font-size: 14px;
  color: var(--cm-text-secondary, #666);
  line-height: 1.7;
  margin: 0;
}

.rel-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.rel-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--cm-bg, #fafafa);
  border-radius: 10px;
  font-size: 13px;
}
.rel-node {
  padding: 3px 10px;
  background: var(--cm-bg-primary, #fff);
  border-radius: 8px;
  font-weight: 500;
  color: var(--cm-text, #1a1a1a);
  border: 1px solid var(--cm-border, #e5e5e5);
}
.rel-type {
  color: var(--cm-primary, #6366f1);
  font-weight: 500;
  font-size: 12px;
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.meta-item {
  padding: 10px 12px;
  background: var(--cm-bg, #fafafa);
  border-radius: 8px;
}
.meta-label {
  display: block;
  font-size: 11px;
  color: var(--cm-text-muted, #999);
  margin-bottom: 4px;
}
.meta-value {
  font-size: 13px;
  color: var(--cm-text, #1a1a1a);
  font-weight: 500;
}

.drawer-actions {
  display: flex;
  gap: 10px;
  padding-top: 16px;
  border-top: 1px solid var(--cm-border, #e5e5e5);
}

.form-field {
  margin-bottom: 18px;
}
.form-field label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--cm-text, #1a1a1a);
  margin-bottom: 6px;
}
.field-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 10px;
  background: var(--cm-bg, #fafafa);
  color: var(--cm-text, #1a1a1a);
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
}
.field-input:focus {
  outline: none;
  border-color: var(--cm-primary, #6366f1);
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}
</style>
