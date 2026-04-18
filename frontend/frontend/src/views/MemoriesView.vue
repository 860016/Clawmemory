<template>
  <div class="memories-view" style="padding: 16px">
    <div style="display: flex; align-items: center; margin-bottom: 16px; gap: 12px">
      <el-radio-group v-model="layerFilter" @change="handleFilter">
        <el-radio-button label="">全部</el-radio-button>
        <el-radio-button label="preference">偏好</el-radio-button>
        <el-radio-button label="knowledge">知识</el-radio-button>
        <el-radio-button label="short_term">短期</el-radio-button>
        <el-radio-button label="private">私密</el-radio-button>
      </el-radio-group>
      <el-input v-model="searchQuery" placeholder="搜索记忆..." style="width: 300px" @keydown.enter="handleSearchKeyword">
        <template #append>
          <el-dropdown @command="handleSearchType">
            <el-button>搜索</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="keyword">关键词</el-dropdown-item>
                <el-dropdown-item command="semantic">语义</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-input>
      <el-button type="primary" @click="showAddDialog = true">添加记忆</el-button>
    </div>

    <el-table :data="memoryStore.memories" stripe style="width: 100%">
      <el-table-column prop="key" label="键" width="200" />
      <el-table-column prop="value" label="值" show-overflow-tooltip />
      <el-table-column prop="layer" label="层级" width="100" />
      <el-table-column prop="importance" label="重要性" width="80" />
      <el-table-column prop="source" label="来源" width="80" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加记忆对话框 -->
    <el-dialog v-model="showAddDialog" title="添加记忆">
      <el-form :model="newMemory" label-width="80px">
        <el-form-item label="层级">
          <el-select v-model="newMemory.layer">
            <el-option label="偏好" value="preference" />
            <el-option label="知识" value="knowledge" />
            <el-option label="短期" value="short_term" />
            <el-option label="私密" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item label="键"><el-input v-model="newMemory.key" /></el-form-item>
        <el-form-item label="值"><el-input v-model="newMemory.value" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd">添加</el-button>
      </template>
    </el-dialog>

    <!-- 编辑记忆对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑记忆">
      <el-form :model="editMemory" label-width="80px">
        <el-form-item label="层级">
          <el-select v-model="editMemory.layer">
            <el-option label="偏好" value="preference" />
            <el-option label="知识" value="knowledge" />
            <el-option label="短期" value="short_term" />
            <el-option label="私密" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item label="键"><el-input v-model="editMemory.key" /></el-form-item>
        <el-form-item label="值"><el-input v-model="editMemory.value" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="重要性"><el-input-number v-model="editMemory.importance" :min="0" :max="1" :step="0.1" /></el-form-item>
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
import { useMemoryStore } from '../stores/memory'
import { ElMessage, ElMessageBox } from 'element-plus'

const memoryStore = useMemoryStore()
const layerFilter = ref('')
const searchQuery = ref('')
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const newMemory = ref({ layer: 'knowledge', key: '', value: '' })
const editMemory = ref({ id: 0, layer: '', key: '', value: '', importance: 0.5 })

onMounted(() => memoryStore.fetchMemories())

function handleFilter() { memoryStore.fetchMemories(layerFilter.value || undefined) }

async function handleSearchKeyword() { if (searchQuery.value) await memoryStore.searchKeyword(searchQuery.value) }

function handleSearchType(type: string) {
  if (!searchQuery.value) return
  if (type === 'semantic') memoryStore.searchSemantic(searchQuery.value)
  else memoryStore.searchKeyword(searchQuery.value)
}

async function handleAdd() {
  await memoryStore.createMemory(newMemory.value)
  showAddDialog.value = false
  newMemory.value = { layer: 'knowledge', key: '', value: '' }
  ElMessage.success('添加成功')
}

function handleEdit(row: any) {
  editMemory.value = { id: row.id, layer: row.layer, key: row.key, value: row.value, importance: row.importance ?? 0.5 }
  showEditDialog.value = true
}

async function handleSaveEdit() {
  await memoryStore.updateMemory(editMemory.value.id, {
    key: editMemory.value.key,
    value: editMemory.value.value,
    layer: editMemory.value.layer,
    importance: editMemory.value.importance,
  })
  showEditDialog.value = false
  ElMessage.success('更新成功')
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm('确定删除此记忆？', '确认')
  await memoryStore.deleteMemory(id)
  ElMessage.success('已删除')
}
</script>
