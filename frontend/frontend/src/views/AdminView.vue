<template>
  <div class="admin-view" style="padding: 16px">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 用户管理 -->
      <el-tab-pane label="用户管理" name="users">
        <el-table :data="adminStore.users" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="role" label="角色" width="100" />
          <el-table-column prop="is_active" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'danger'">{{ row.is_active ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" @click="handleToggleActive(row)">{{ row.is_active ? '禁用' : '启用' }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 模型管理 -->
      <el-tab-pane label="模型管理" name="models">
        <el-button type="primary" style="margin-bottom: 16px" @click="showModelDialog = true">添加模型</el-button>
        <el-table :data="adminStore.models" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" width="150" />
          <el-table-column prop="provider" label="提供商" width="120" />
          <el-table-column prop="model_id" label="模型ID" width="200" />
          <el-table-column prop="base_url" label="Base URL" show-overflow-tooltip />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteModel(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 技能管理 -->
      <el-tab-pane label="技能管理" name="skills">
        <el-button type="primary" style="margin-bottom: 16px" @click="showSkillDialog = true">添加技能</el-button>
        <el-table :data="adminStore.skills" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" width="180" />
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column label="配置" show-overflow-tooltip>
            <template #default="{ row }">{{ row.config ? JSON.stringify(row.config) : '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteSkill(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 节点管理 -->
      <el-tab-pane label="节点管理" name="nodes">
        <el-table :data="adminStore.nodes" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="节点名" width="180" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="agent_id" label="Agent ID" width="100" />
          <el-table-column prop="last_ping" label="最后心跳" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="handleDeleteNode(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 备份管理 -->
      <el-tab-pane label="备份管理" name="backups">
        <div style="display: flex; gap: 12px; margin-bottom: 16px">
          <el-button type="primary" @click="handleCreateBackup">创建备份</el-button>
          <el-upload :show-file-list="false" :before-upload="handleUploadBackup" accept=".zip">
            <el-button type="success">导入备份</el-button>
          </el-upload>
        </div>
        <el-table :data="adminStore.backups" stripe style="width: 100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="filename" label="文件名" width="250" />
          <el-table-column prop="size" label="大小" width="100">
            <template #default="{ row }">{{ formatSize(row.size) }}</template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="handleDownloadBackup(row.id)">下载</el-button>
              <el-button size="small" type="warning" @click="handleRestoreBackup(row.id)">恢复</el-button>
              <el-button size="small" type="danger" @click="handleDeleteBackup(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 授权信息 -->
      <el-tab-pane label="授权信息" name="license">
        <el-descriptions :column="2" border style="margin-top: 16px">
          <el-descriptions-item label="授权状态">
            <el-tag :type="adminStore.licenseInfo?.active ? 'success' : 'danger'">
              {{ adminStore.licenseInfo?.active ? '已激活' : '未激活' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="授权类型">
            <el-tag>{{ adminStore.licenseInfo?.tier || 'OSS' }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="到期时间">{{ adminStore.licenseInfo?.expires_at || '永久' }}</el-descriptions-item>
          <el-descriptions-item label="设备数">{{ adminStore.licenseInfo?.device_count ?? 0 }} / {{ adminStore.licenseInfo?.max_devices ?? 1 }}</el-descriptions-item>
        </el-descriptions>
        <el-divider />
        <el-form :inline="true" style="margin-top: 16px">
          <el-form-item label="激活授权码">
            <el-input v-model="licenseKey" placeholder="输入授权码" style="width: 300px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleActivateLicense">激活</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <!-- 添加模型对话框 -->
    <el-dialog v-model="showModelDialog" title="添加模型" width="500px">
      <el-form :model="modelForm" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="modelForm.name" /></el-form-item>
        <el-form-item label="提供商" required>
          <el-select v-model="modelForm.provider" style="width: 100%">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Anthropic" value="anthropic" />
            <el-option label="Ollama" value="ollama" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型ID" required><el-input v-model="modelForm.model_id" placeholder="如 gpt-4o-mini" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="modelForm.api_key" type="password" show-password /></el-form-item>
        <el-form-item label="Base URL"><el-input v-model="modelForm.base_url" placeholder="可选，如 https://api.openai.com/v1" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showModelDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveModel">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加技能对话框 -->
    <el-dialog v-model="showSkillDialog" title="添加技能" width="500px">
      <el-form :model="skillForm" label-width="80px">
        <el-form-item label="名称" required><el-input v-model="skillForm.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="skillForm.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="配置">
          <el-input v-model="skillForm.configStr" type="textarea" :rows="3" placeholder='JSON 格式配置' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSkillDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveSkill">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAdminStore } from '../stores/admin'
import { backupsApi } from '../api/backups'
import { ElMessage, ElMessageBox } from 'element-plus'

const adminStore = useAdminStore()
const activeTab = ref('users')
const licenseKey = ref('')

// Model form
const showModelDialog = ref(false)
const modelForm = ref({ name: '', provider: 'openai', model_id: '', api_key: '', base_url: '' })

// Skill form
const showSkillDialog = ref(false)
const skillForm = ref({ name: '', description: '', configStr: '' })

onMounted(() => {
  adminStore.fetchUsers()
})

function handleTabChange(tab: string) {
  if (tab === 'users') adminStore.fetchUsers()
  else if (tab === 'models') adminStore.fetchModels()
  else if (tab === 'skills') adminStore.fetchSkills()
  else if (tab === 'nodes') adminStore.fetchNodes()
  else if (tab === 'backups') adminStore.fetchBackups()
  else if (tab === 'license') adminStore.fetchLicenseInfo()
}

async function handleToggleActive(user: any) {
  await adminStore.toggleUserActive(user.id, !user.is_active)
  user.is_active = !user.is_active
  ElMessage.success('操作成功')
}

async function handleSaveModel() {
  if (!modelForm.value.name || !modelForm.value.model_id) return ElMessage.warning('请填写必填项')
  await adminStore.createModel(modelForm.value)
  ElMessage.success('添加成功')
  showModelDialog.value = false
  modelForm.value = { name: '', provider: 'openai', model_id: '', api_key: '', base_url: '' }
}

async function handleDeleteModel(id: number) {
  await ElMessageBox.confirm('确定删除此模型配置？', '确认')
  await adminStore.deleteModel(id)
  ElMessage.success('已删除')
}

async function handleSaveSkill() {
  if (!skillForm.value.name) return ElMessage.warning('请填写名称')
  let config: Record<string, any> | undefined
  if (skillForm.value.configStr) {
    try { config = JSON.parse(skillForm.value.configStr) } catch { return ElMessage.error('配置 JSON 格式错误') }
  }
  await adminStore.createSkill({ name: skillForm.value.name, description: skillForm.value.description, config })
  ElMessage.success('添加成功')
  showSkillDialog.value = false
  skillForm.value = { name: '', description: '', configStr: '' }
}

async function handleDeleteSkill(id: number) {
  await ElMessageBox.confirm('确定删除此技能？', '确认')
  await adminStore.deleteSkill(id)
  ElMessage.success('已删除')
}

async function handleDeleteNode(id: number) {
  await ElMessageBox.confirm('确定删除此节点？', '确认')
  await adminStore.deleteNode(id)
  ElMessage.success('已删除')
}

async function handleCreateBackup() {
  await adminStore.createBackup()
  ElMessage.success('备份创建成功')
}

async function handleDownloadBackup(id: number) {
  const resp = await backupsApi.download(id)
  const url = window.URL.createObjectURL(new Blob([resp.data]))
  const a = document.createElement('a')
  a.href = url
  a.download = `backup_${id}.zip`
  a.click()
  window.URL.revokeObjectURL(url)
}

async function handleRestoreBackup(id: number) {
  await ElMessageBox.confirm('恢复备份将覆盖当前数据，确定继续？', '警告', { type: 'warning' })
  await backupsApi.restore(id)
  ElMessage.success('恢复成功')
}

async function handleDeleteBackup(id: number) {
  await ElMessageBox.confirm('确定删除此备份？', '确认')
  await adminStore.deleteBackup(id)
  ElMessage.success('已删除')
}

async function handleUploadBackup(file: File) {
  await backupsApi.upload(file)
  ElMessage.success('导入成功')
  adminStore.fetchBackups()
  return false
}

async function handleActivateLicense() {
  if (!licenseKey.value) return ElMessage.warning('请输入授权码')
  await adminStore.activateLicense(licenseKey.value)
  ElMessage.success('激活成功')
  licenseKey.value = ''
}

function formatSize(bytes: number) {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
