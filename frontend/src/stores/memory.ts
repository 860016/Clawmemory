import { defineStore } from 'pinia'
import { ref } from 'vue'
import { memoryApi } from '../api/go-memories'

export const useMemoryStore = defineStore('memory', () => {
  const memories = ref<any[]>([])
  const total = ref(0)
  const currentLayer = ref('')
  const currentSourceAgent = ref('')
  const currentVisibility = ref('')
  const searchQuery = ref('')

  async function fetchMemories(layer?: string, page = 1, sourceAgent?: string, visibility?: string) {
    const resp = await memoryApi.list({
      layer: layer || undefined,
      page,
      size: 20,
      source_agent: sourceAgent || undefined,
      visibility: visibility || undefined,
    })
    memories.value = resp.data.items
    total.value = resp.data.total
  }

  async function createMemory(data: any) {
    await memoryApi.create(data)
    await fetchMemories(currentLayer.value || undefined, 1, currentSourceAgent.value || undefined, currentVisibility.value || undefined)
  }

  async function updateMemory(id: number, data: any) {
    await memoryApi.update(id, data)
    await fetchMemories(currentLayer.value || undefined, 1, currentSourceAgent.value || undefined, currentVisibility.value || undefined)
  }

  async function deleteMemory(id: number) {
    await memoryApi.delete(id)
    await fetchMemories(currentLayer.value || undefined, 1, currentSourceAgent.value || undefined, currentVisibility.value || undefined)
  }

  async function searchKeyword(q: string) {
    const resp = await memoryApi.searchKeyword(q)
    const items = Array.isArray(resp.data) ? resp.data : (resp.data.items || [])
    memories.value = items
    total.value = items.length
  }

  async function searchSemantic(q: string) {
    const resp = await memoryApi.searchSemantic(q)
    const items = Array.isArray(resp.data) ? resp.data : (resp.data.items || [])
    memories.value = items
    total.value = items.length
  }

  return {
    memories, total, currentLayer, currentSourceAgent, currentVisibility, searchQuery,
    fetchMemories, createMemory, updateMemory, deleteMemory, searchKeyword, searchSemantic,
  }
})
