<template>
  <div class="trash-page">
    <div class="page-header">
      <h1>🗑️ {{ $t('trash.title') }}</h1>
      <div class="header-actions">
        <el-button @click="loadTrash" :loading="loading">{{ $t('common.refresh') }}</el-button>
        <el-button type="danger" @click="emptyTrash" :disabled="!items.length" :loading="emptying">{{ $t('trash.emptyAll') }}</el-button>
      </div>
    </div>

    <div v-if="!items.length && !loading" class="empty-state">
      <div class="empty-icon">♻️</div>
      <div class="empty-text">{{ $t('trash.empty') }}</div>
    </div>

    <div v-else class="trash-list">
      <div v-for="item in items" :key="item.id" class="trash-item">
        <div class="item-main">
          <div class="item-key">{{ item.key }}</div>
          <div class="item-value" v-if="item.value">{{ truncate(item.value, 200) }}</div>
          <div class="item-meta">
            <el-tag size="small" type="info">{{ item.layer || 'knowledge' }}</el-tag>
            <span class="item-importance" v-if="item.importance !== undefined">⭐ {{ item.importance.toFixed(2) }}</span>
            <span class="item-date" v-if="item.trashed_at">{{ item.trashed_at }}</span>
          </div>
        </div>
        <div class="item-actions">
          <el-button size="small" type="primary" @click="restoreItem(item.id)" :loading="restoring[item.id]">
            {{ $t('trash.restore') }}
          </el-button>
          <el-button size="small" type="danger" @click="deleteItem(item.id)">
            {{ $t('common.delete') }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { memoryApi } from '../api/go-memories'

const { t } = useI18n()
const items = ref<any[]>([])
const loading = ref(false)
const emptying = ref(false)
const restoring = reactive<Record<number, boolean>>({})

function truncate(s: string, max: number) {
  if (!s) return ''
  return s.length > max ? s.slice(0, max) + '...' : s
}

async function loadTrash() {
  loading.value = true
  try {
    const { data } = await memoryApi.listTrash()
    items.value = data.items || []
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

async function restoreItem(id: number) {
  restoring[id] = true
  try {
    await memoryApi.restore(id)
    ElMessage.success(t('trash.restored'))
    items.value = items.value.filter(i => i.id !== id)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.failed'))
  } finally {
    restoring[id] = false
  }
}

async function deleteItem(id: number) {
  try {
    await ElMessageBox.confirm(t('trash.deleteConfirm'), t('common.confirm'), { type: 'warning' })
    await memoryApi.delete(id)
    ElMessage.success(t('common.success'))
    items.value = items.value.filter(i => i.id !== id)
  } catch {}
}

async function emptyTrash() {
  try {
    await ElMessageBox.confirm(t('trash.emptyConfirm'), t('common.confirm'), { type: 'warning' })
    emptying.value = true
    await memoryApi.emptyTrash()
    ElMessage.success(t('trash.emptied'))
    items.value = []
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(e.response?.data?.error || t('common.failed'))
    }
  } finally {
    emptying.value = false
  }
}

onMounted(loadTrash)
</script>

<style scoped>
.trash-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--cm-bg-secondary, #f5f5f5);
}
.page-header {
  background: var(--cm-bg-primary, #fff);
  padding: 24px 28px;
  border-bottom: 1px solid var(--cm-border, #e5e5e5);
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.page-header h1 {
  font-size: 20px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--cm-text-muted);
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}
.empty-text {
  font-size: 15px;
}
.trash-list {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.trash-item {
  background: var(--cm-bg-primary, #fff);
  border: 1px solid var(--cm-border, #e5e5e5);
  border-radius: 8px;
  padding: 14px 18px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}
.item-main {
  flex: 1;
  min-width: 0;
}
.item-key {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  word-break: break-all;
}
.item-value {
  font-size: 13px;
  color: var(--cm-text-secondary);
  margin-top: 4px;
  word-break: break-all;
}
.item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--cm-text-muted);
}
.item-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
@media (max-width: 768px) {
  .trash-item {
    flex-direction: column;
  }
  .item-actions {
    align-self: flex-end;
  }
}
</style>
