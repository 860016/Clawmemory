<template>
  <div class="knowledge-view" style="padding: 16px">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 实体列表 -->
      <el-tab-pane label="实体列表" name="entities">
        <div style="display: flex; align-items: center; margin-bottom: 16px; gap: 12px">
          <el-input v-model="entitySearch" placeholder="搜索实体..." style="width: 250px" clearable />
          <el-button type="primary" @click="showEntityDialog = true">添加实体</el-button>
        </div>
        <el-table :data="filteredEntities" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" width="180" />
          <el-table-column prop="entity_type" label="类型" width="120" />
          <el-table-column label="属性" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.properties ? JSON.stringify(row.properties) : '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170" />
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" @click="handleEditEntity(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteEntity(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 关系列表 -->
      <el-tab-pane label="关系列表" name="relations">
        <div style="display: flex; align-items: center; margin-bottom: 16px; gap: 12px">
          <el-button type="primary" @click="showRelationDialog = true">添加关系</el-button>
        </div>
        <el-table :data="knowledgeStore.relations" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="source_id" label="源实体ID" width="100" />
          <el-table-column prop="relation_type" label="关系类型" width="150" />
          <el-table-column prop="target_id" label="目标实体ID" width="100" />
          <el-table-column label="属性" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.properties ? JSON.stringify(row.properties) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteRelation(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 知识图谱 -->
      <el-tab-pane label="知识图谱" name="graph">
        <el-alert v-if="!isPro" type="info" :closable="false" style="margin-bottom: 16px">
          知识图谱可视化功能在 Pro 版本中提供
        </el-alert>
        <template v-else>
          <div style="display: flex; gap: 12px; margin-bottom: 12px">
            <el-button type="primary" @click="handleLoadGraph">刷新图谱</el-button>
            <el-button @click="handleResetLayout">重置布局</el-button>
            <el-tag>实体: {{ knowledgeStore.entities.length }}</el-tag>
            <el-tag type="success">关系: {{ knowledgeStore.relations.length }}</el-tag>
          </div>
          <div ref="graphContainer" class="graph-container"></div>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 添加/编辑实体对话框 -->
    <el-dialog v-model="showEntityDialog" :title="editingEntity ? '编辑实体' : '添加实体'" width="500px">
      <el-form :model="entityForm" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="entityForm.name" placeholder="实体名称" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="entityForm.entity_type" placeholder="选择类型" style="width: 100%">
            <el-option label="人物" value="person" />
            <el-option label="组织" value="organization" />
            <el-option label="地点" value="location" />
            <el-option label="概念" value="concept" />
            <el-option label="事件" value="event" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="属性">
          <el-input v-model="entityForm.propertiesStr" type="textarea" :rows="3" placeholder='JSON 格式，如 {"key": "value"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEntityDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEntity">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加关系对话框 -->
    <el-dialog v-model="showRelationDialog" title="添加关系" width="500px">
      <el-form :model="relationForm" label-width="80px">
        <el-form-item label="源实体" required>
          <el-select v-model="relationForm.source_id" placeholder="选择源实体" style="width: 100%">
            <el-option v-for="e in knowledgeStore.entities" :key="e.id" :label="`${e.name} (${e.id})`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关系类型" required>
          <el-input v-model="relationForm.relation_type" placeholder="如：knows, works_at, located_in" />
        </el-form-item>
        <el-form-item label="目标实体" required>
          <el-select v-model="relationForm.target_id" placeholder="选择目标实体" style="width: 100%">
            <el-option v-for="e in knowledgeStore.entities" :key="e.id" :label="`${e.name} (${e.id})`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="属性">
          <el-input v-model="relationForm.propertiesStr" type="textarea" :rows="3" placeholder='JSON 格式，如 {"since": "2024"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRelationDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRelation">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useKnowledgeStore } from '../stores/knowledge'
import { useAuthStore } from '../stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import cytoscape, { Core } from 'cytoscape'

const knowledgeStore = useKnowledgeStore()
const auth = useAuthStore()
const isPro = computed(() => auth.role === 'admin')

const activeTab = ref('entities')
const entitySearch = ref('')
const graphContainer = ref<HTMLElement>()
let cy: Core | null = null

// Entity form
const showEntityDialog = ref(false)
const editingEntity = ref<any>(null)
const entityForm = ref({ name: '', entity_type: '', propertiesStr: '' })

// Relation form
const showRelationDialog = ref(false)
const relationForm = ref({ source_id: 0, target_id: 0, relation_type: '', propertiesStr: '' })

// Color map for entity types
const typeColors: Record<string, string> = {
  person: '#409eff',
  organization: '#67c23a',
  location: '#e6a23c',
  concept: '#f56c6c',
  event: '#909399',
  other: '#b37feb',
}

const filteredEntities = computed(() => {
  if (!entitySearch.value) return knowledgeStore.entities
  const q = entitySearch.value.toLowerCase()
  return knowledgeStore.entities.filter(e => e.name?.toLowerCase().includes(q) || e.entity_type?.toLowerCase().includes(q))
})

onMounted(() => {
  knowledgeStore.fetchEntities()
  knowledgeStore.fetchRelations()
})

onBeforeUnmount(() => {
  if (cy) { cy.destroy(); cy = null }
})

function handleTabChange(tab: string) {
  if (tab === 'entities') knowledgeStore.fetchEntities()
  else if (tab === 'relations') knowledgeStore.fetchRelations()
  else if (tab === 'graph' && isPro.value) handleLoadGraph()
}

function handleEditEntity(row: any) {
  editingEntity.value = row
  entityForm.value = {
    name: row.name,
    entity_type: row.entity_type,
    propertiesStr: row.properties ? JSON.stringify(row.properties, null, 2) : '',
  }
  showEntityDialog.value = true
}

async function handleSaveEntity() {
  if (!entityForm.value.name || !entityForm.value.entity_type) {
    return ElMessage.warning('请填写必填项')
  }
  let properties: Record<string, any> | undefined
  if (entityForm.value.propertiesStr) {
    try { properties = JSON.parse(entityForm.value.propertiesStr) } catch { return ElMessage.error('属性 JSON 格式错误') }
  }
  try {
    if (editingEntity.value) {
      await knowledgeStore.updateEntity(editingEntity.value.id, { name: entityForm.value.name, entity_type: entityForm.value.entity_type, properties })
      ElMessage.success('更新成功')
    } else {
      await knowledgeStore.createEntity({ name: entityForm.value.name, entity_type: entityForm.value.entity_type, properties })
      ElMessage.success('添加成功')
    }
    showEntityDialog.value = false
    editingEntity.value = null
    entityForm.value = { name: '', entity_type: '', propertiesStr: '' }
  } catch { /* handled by interceptor */ }
}

async function handleDeleteEntity(id: number) {
  await ElMessageBox.confirm('确定删除此实体？关联的关系也会被删除。', '确认')
  await knowledgeStore.deleteEntity(id)
  ElMessage.success('已删除')
}

async function handleSaveRelation() {
  if (!relationForm.value.source_id || !relationForm.value.target_id || !relationForm.value.relation_type) {
    return ElMessage.warning('请填写必填项')
  }
  let properties: Record<string, any> | undefined
  if (relationForm.value.propertiesStr) {
    try { properties = JSON.parse(relationForm.value.propertiesStr) } catch { return ElMessage.error('属性 JSON 格式错误') }
  }
  try {
    await knowledgeStore.createRelation({
      source_id: relationForm.value.source_id,
      target_id: relationForm.value.target_id,
      relation_type: relationForm.value.relation_type,
      properties,
    })
    ElMessage.success('添加成功')
    showRelationDialog.value = false
    relationForm.value = { source_id: 0, target_id: 0, relation_type: '', propertiesStr: '' }
    knowledgeStore.fetchRelations()
  } catch { /* handled by interceptor */ }
}

async function handleDeleteRelation(id: number) {
  await ElMessageBox.confirm('确定删除此关系？', '确认')
  await knowledgeStore.deleteRelation(id)
  ElMessage.success('已删除')
}

async function handleLoadGraph() {
  await knowledgeStore.fetchGraphData()
  await nextTick()
  renderGraph()
}

function renderGraph() {
  if (!graphContainer.value) return

  const data = knowledgeStore.graphData
  const nodes = (data.nodes ?? []).map((n: any) => ({
    data: {
      id: String(n.id),
      label: n.name || `Entity ${n.id}`,
      type: n.entity_type || 'other',
      color: typeColors[n.entity_type] || typeColors.other,
    },
  }))
  const edges = (data.edges ?? []).map((e: any, i: number) => ({
    data: {
      id: `e${i}`,
      source: String(e.source_id),
      target: String(e.target_id),
      label: e.relation_type || '',
    },
  }))

  if (cy) cy.destroy()

  cy = cytoscape({
    container: graphContainer.value,
    elements: [...nodes, ...edges],
    style: [
      {
        selector: 'node',
        style: {
          'background-color': 'data(color)',
          label: 'data(label)',
          'text-outline-color': '#fff',
          'text-outline-width': 2,
          'font-size': 12,
          color: '#333',
          width: 40,
          height: 40,
        },
      },
      {
        selector: 'edge',
        style: {
          width: 2,
          'line-color': '#ccc',
          'target-arrow-color': '#999',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier',
          label: 'data(label)',
          'font-size': 10,
          color: '#999',
          'text-rotation': 'autorotate',
          'text-outline-color': '#fff',
          'text-outline-width': 1,
        },
      },
      {
        selector: 'node:active',
        style: { 'overlay-opacity': 0.1 },
      },
    ],
    layout: {
      name: 'cose',
      padding: 30,
      animate: true,
      animationDuration: 500,
    },
    userZoomingEnabled: true,
    userPanningEnabled: true,
    boxSelectionEnabled: false,
  })

  // Click node to show details
  cy.on('tap', 'node', (evt) => {
    const node = evt.target
    ElMessage.info(`${node.data('label')} (${node.data('type')})`)
  })
}

function handleResetLayout() {
  if (!cy) return
  cy.layout({ name: 'cose', padding: 30, animate: true, animationDuration: 500 }).run()
}
</script>

<style scoped>
.graph-container {
  min-height: 500px;
  border: 1px solid #e6e6e6;
  border-radius: 4px;
  background: #fafafa;
}
</style>
