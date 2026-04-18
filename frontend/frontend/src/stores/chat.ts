import { defineStore } from 'pinia'
import { ref } from 'vue'
import { chatApi } from '../api/chat'

export const useChatStore = defineStore('chat', () => {
  const sessions = ref<any[]>([])
  const currentSessionId = ref<number | null>(null)
  const messages = ref<any[]>([])

  async function fetchSessions() {
    const resp = await chatApi.listSessions()
    sessions.value = resp.data
  }

  async function createSession(data: { agent_id: number; title?: string }) {
    const resp = await chatApi.createSession(data)
    sessions.value.unshift(resp.data)
    currentSessionId.value = resp.data.id
    return resp.data
  }

  async function fetchMessages(sessionId: number, limit?: number) {
    const resp = await chatApi.getMessages(sessionId, limit)
    messages.value = resp.data
    currentSessionId.value = sessionId
  }

  async function deleteSession(sessionId: number) {
    await chatApi.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    if (currentSessionId.value === sessionId) {
      currentSessionId.value = sessions.value[0]?.id ?? null
    }
  }

  function clearMessages() {
    messages.value = []
    currentSessionId.value = null
  }

  return { sessions, currentSessionId, messages, fetchSessions, createSession, fetchMessages, deleteSession, clearMessages }
})
