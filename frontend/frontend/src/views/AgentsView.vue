<template>
  <div class="agents-view" style="padding: 16px">
    <div style="display: flex; align-items: center; margin-bottom: 16px; gap: 12px">
      <h3 style="margin: 0">Agent 管理</h3>
      <el-button type="primary" @click="showCreateDialog = true">创建 Agent</el-button>
    </div>

    <el-table :data="agents" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" width="180" />
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
      <el-table-column prop="model_id" label="模型" width="120" />
      <el-table-column prop="is_default" label="默认" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_default ? 'success' : 'info'" size="small">{{ row.is_default ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="170" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建 Agent 对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建 Agent" width="500px">
      <el-form :model="agentForm" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="agentForm.name" placeholder="Agent 名称" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="agentForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑 Agent 对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑 Agent" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="名称"><el-input v-model="editForm.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="editForm.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="模型ID"><el-input v-model="editForm.model_id" placeholder="可选" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { agentsApi } from '../api/agents'
import { ElMessage, ElMessageBox } from 'element-plus'

const agents = ref<any[]>([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const agentForm = ref({ name: '', description: '' })
const editForm = ref({ id: 0, name: '', description: '', model_id: '' })

onMounted(() => fetchAgents())

async function fetchAgents() {
  try {
    const resp = await agentsApi.list()
    agents.value = Array.isArray(resp.data) ? resp.data : resp.data.items ?? []
  } catch { agents.value = [] }
}

async function handleCreate() {
  if (!agentForm.value.name) return ElMessage.warning('请输入名称')
  await agentsApi.create(agentForm.value)
  ElMessage.success('创建成功')
  showCreateDialog.value = false
  agentForm.value = { name: '', description: '' }
  fetchAgents()
}

function handleEdit(row: any) {
  editForm.value = { id: row.id, name: row.name, description: row.description || '', model_id: row.model_id || '' }
  showEditDialog.value = true
}

async function handleSaveEdit() {
  await agentsApi.update(editForm.value.id, { name: editForm.value.name, description: editForm.value.description, model_id: editForm.value.model_id })
  ElMessage.success('保存成功')
  showEditDialog.value = false
  fetchAgents()
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定删除此 Agent？相关数据也会被删除。', '确认')
  await agentsApi.delete(id)
  ElMessage.success('已删除')
  fetchAgents()
}
</script>
