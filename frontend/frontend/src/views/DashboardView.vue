<template>
  <div class="dashboard-view" style="padding: 20px">
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="记忆总数" :value="stats.memoryCount">
            <template #prefix><el-icon><Collection /></el-icon></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="实体数" :value="stats.entityCount">
            <template #prefix><el-icon><Connection /></el-icon></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="会话数" :value="stats.sessionCount">
            <template #prefix><el-icon><ChatDotRound /></el-icon></template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="技能数" :value="stats.skillCount">
            <template #prefix><el-icon><MagicStick /></el-icon></template>
          </el-statistic>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card header="最近记忆" shadow="hover">
          <el-table :data="recentMemories" size="small" stripe>
            <el-table-column prop="key" label="键" show-overflow-tooltip />
            <el-table-column prop="layer" label="层级" width="80" />
            <el-table-column prop="created_at" label="时间" width="160" />
          </el-table>
          <el-empty v-if="!recentMemories.length" description="暂无记忆" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="最近会话" shadow="hover">
          <el-table :data="recentSessions" size="small" stripe>
            <el-table-column prop="title" label="标题" show-overflow-tooltip />
            <el-table-column prop="agent_id" label="Agent" width="80" />
            <el-table-column prop="created_at" label="时间" width="160" />
          </el-table>
          <el-empty v-if="!recentSessions.length" description="暂无会话" />
        </el-card>

        <el-card header="授权状态" shadow="hover" style="margin-top: 20px">
          <el-descriptions :column="2" size="small">
            <el-descriptions-item label="版本">
              <el-tag :type="licenseInfo?.tier === 'oss' ? 'info' : 'success'">{{ licenseInfo?.tier?.toUpperCase() || 'OSS' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="licenseInfo?.active ? 'success' : 'warning'">{{ licenseInfo?.active ? '已激活' : '未激活' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="到期时间">{{ licenseInfo?.expires_at || '永久' }}</el-descriptions-item>
            <el-descriptions-item label="最大设备数">{{ licenseInfo?.max_devices ?? 1 }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import api from '../api/client'
import { Collection, Connection, ChatDotRound, MagicStick } from '@element-plus/icons-vue'

const stats = reactive({ memoryCount: 0, entityCount: 0, sessionCount: 0, skillCount: 0 })
const recentMemories = ref<any[]>([])
const recentSessions = ref<any[]>([])
const licenseInfo = ref<any>(null)

onMounted(async () => {
  const promises = [
    api.get('/memories', { params: { page: 1, size: 5 } }).catch(() => ({ data: { items: [], total: 0 } })),
    api.get('/knowledge/entities', { params: { page: 1, size: 1 } }).catch(() => ({ data: { items: [], total: 0 } })),
    api.get('/chat/sessions', { params: { limit: 5 } }).catch(() => ({ data: [] })),
    api.get('/skills').catch(() => ({ data: [] })),
    api.get('/license/info').catch(() => ({ data: null })),
  ]
  const [memResp, entityResp, sessionResp, skillResp, licResp] = await Promise.all(promises)

  stats.memoryCount = memResp.data.total ?? memResp.data.items?.length ?? 0
  recentMemories.value = memResp.data.items ?? memResp.data ?? []

  stats.entityCount = entityResp.data.total ?? entityResp.data.items?.length ?? 0

  const sessions = Array.isArray(sessionResp.data) ? sessionResp.data : sessionResp.data.items ?? []
  stats.sessionCount = sessions.length
  recentSessions.value = sessions.slice(0, 5)

  const skills = Array.isArray(skillResp.data) ? skillResp.data : skillResp.data.items ?? []
  stats.skillCount = skills.length

  licenseInfo.value = licResp.data
})
</script>
