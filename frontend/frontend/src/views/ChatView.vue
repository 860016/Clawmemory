<template>
  <div class="chat-view">
    <!-- 会话侧栏 -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <el-button type="primary" size="small" style="width: 100%" @click="handleNewSession">新建会话</el-button>
      </div>
      <div class="session-list">
        <div
          v-for="s in chatStore.sessions"
          :key="s.id"
          :class="['session-item', { active: s.id === chatStore.currentSessionId }]"
          @click="handleSelectSession(s.id)"
        >
          <div class="session-title">{{ s.title || `会话 ${s.id}` }}</div>
          <div class="session-meta">{{ formatTime(s.created_at) }}</div>
          <el-button class="session-delete" size="small" type="danger" text @click.stop="handleDeleteSession(s.id)">×</el-button>
        </div>
        <el-empty v-if="chatStore.sessions.length === 0" description="暂无会话" :image-size="60" />
      </div>
    </div>

    <!-- 聊天主区域 -->
    <div class="chat-main">
      <div class="chat-header">
        <el-select v-model="currentAgentId" placeholder="选择Agent" style="width: 200px">
          <el-option v-for="a in agents" :key="a.id" :label="a.name" :value="a.id" />
        </el-select>
        <el-tag :type="wsConnected ? 'success' : 'danger'" style="margin-left: 12px">
          {{ wsConnected ? '已连接' : '未连接' }}
        </el-tag>
      </div>
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="chatMessages.length === 0" class="chat-empty">
          <el-empty description="发送一条消息开始对话" :image-size="100" />
        </div>
        <div v-for="(msg, i) in chatMessages" :key="i" :class="['message', msg.role]">
          <div class="message-content">{{ msg.content }}<span v-if="msg.streaming" class="cursor">|</span></div>
        </div>
      </div>
      <div class="chat-input">
        <el-input v-model="inputText" placeholder="输入消息..." @keydown.enter="sendMessage" :disabled="!wsConnected">
          <template #append><el-button @click="sendMessage" :disabled="!wsConnected">发送</el-button></template>
        </el-input>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { useWebSocket } from '../composables/useWebSocket'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import { ElMessageBox, ElMessage } from 'element-plus'

const auth = useAuthStore()
const chatStore = useChatStore()
const { connected: wsConnected, messages: wsMessages, connect, send } = useWebSocket()
const inputText = ref('')
const currentAgentId = ref(1)
const agents = ref([{ id: 1, name: 'main' }])
const chatMessages = ref<any[]>([])
const messagesContainer = ref<HTMLElement>()

onMounted(async () => {
  if (auth.token) connect(auth.token)
  await chatStore.fetchSessions()
})

watch(wsMessages, () => {
  const last = wsMessages.value[wsMessages.value.length - 1]
  if (!last) return
  if (last.type === 'chunk') {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg && lastMsg.role === 'assistant' && lastMsg.streaming) {
      lastMsg.content += last.content
    }
  } else if (last.type === 'done') {
    const lastMsg = chatMessages.value[chatMessages.value.length - 1]
    if (lastMsg) lastMsg.streaming = false
  } else if (last.type === 'error') {
    chatMessages.value.push({ role: 'system', content: `Error: ${last.message}` })
  }
  nextTick(() => {
    messagesContainer.value?.scrollTo(0, messagesContainer.value.scrollHeight)
  })
}, { deep: true })

watch(() => chatStore.currentSessionId, async (sessionId) => {
  if (sessionId) {
    await chatStore.fetchMessages(sessionId)
    chatMessages.value = chatStore.messages.map((m: any) => ({
      role: m.role,
      content: m.content,
      streaming: false,
    }))
    nextTick(() => {
      messagesContainer.value?.scrollTo(0, messagesContainer.value.scrollHeight)
    })
  }
})

async function handleNewSession() {
  await chatStore.createSession({ agent_id: currentAgentId.value })
  chatMessages.value = []
}

async function handleSelectSession(sessionId: number) {
  chatStore.currentSessionId.value = sessionId
}

async function handleDeleteSession(sessionId: number) {
  await ElMessageBox.confirm('确定删除此会话？', '确认')
  await chatStore.deleteSession(sessionId)
  if (chatStore.currentSessionId.value === sessionId) {
    chatMessages.value = []
  }
  ElMessage.success('已删除')
}

function sendMessage() {
  if (!inputText.value.trim()) return
  if (!chatStore.currentSessionId.value) {
    // Auto create session
    chatStore.createSession({ agent_id: currentAgentId.value }).then(() => {
      pushAndSend()
    })
  } else {
    pushAndSend()
  }
}

function pushAndSend() {
  chatMessages.value.push({ role: 'user', content: inputText.value })
  chatMessages.value.push({ role: 'assistant', content: '', streaming: true })
  send({ type: 'chat', content: inputText.value, agent_id: currentAgentId.value, session_id: chatStore.currentSessionId.value })
  inputText.value = ''
}

function formatTime(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}
</script>

<style scoped>
.chat-view { display: flex; height: 100%; }
.chat-sidebar { width: 240px; border-right: 1px solid #eee; display: flex; flex-direction: column; background: #fafafa; }
.sidebar-header { padding: 12px; border-bottom: 1px solid #eee; }
.session-list { flex: 1; overflow-y: auto; padding: 8px; }
.session-item { padding: 10px 12px; border-radius: 6px; cursor: pointer; margin-bottom: 4px; position: relative; transition: background 0.2s; }
.session-item:hover { background: #e8e8e8; }
.session-item.active { background: #d9ecff; }
.session-title { font-size: 14px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding-right: 20px; }
.session-meta { font-size: 12px; color: #999; margin-top: 4px; }
.session-delete { position: absolute; right: 4px; top: 50%; transform: translateY(-50%); font-size: 16px; }
.chat-main { flex: 1; display: flex; flex-direction: column; }
.chat-header { padding: 12px 16px; border-bottom: 1px solid #eee; display: flex; align-items: center; }
.chat-messages { flex: 1; overflow-y: auto; padding: 16px; }
.chat-empty { display: flex; align-items: center; justify-content: center; height: 100%; }
.message { margin-bottom: 12px; padding: 8px 12px; border-radius: 8px; max-width: 70%; word-break: break-word; }
.message.user { background: #409eff; color: white; margin-left: auto; }
.message.assistant { background: #f4f4f5; }
.message.system { background: #fdf6ec; color: #e6a23c; }
.cursor { animation: blink 1s infinite; }
@keyframes blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }
.chat-input { padding: 16px; border-top: 1px solid #eee; }
</style>
