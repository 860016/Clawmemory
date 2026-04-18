import { defineStore } from 'pinia'
import { ref } from 'vue'
import { memoryApi } from '../api/memories'

export const useMemoryStore = defineStore('memory', () => {
  const memories = ref<any[]>([])
  const total = ref(0)
  const currentLayer = ref('')
  const searchQuery = ref('')

  async function fetchMemories(layer?: string, page = 1) {
    const resp = await memoryApi.list({ layer: layer || undefined, page, size: 20 })
    memories.value = resp.data.items
    total.value = resp.data.total
  }

  async function createMemory(data: any) {
    await memoryApi.create(data)
    await fetchMemories(currentLayer.value || undefined)
  }

  async function updateMemory(id: number, data: any) {
    await memoryApi.update(id, data)
    await fetchMemories(currentLayer.value || undefined)
  }

  async function deleteMemory(id: number) {
    await memoryApi.delete(id)
    await fetchMemories(currentLayer.value || undefined)
  }

  async function searchKeyword(q: string) {
    const resp = await memoryApi.searchKeyword(q)
    memories.value = resp.data
    total.value = resp.data.length
  }

  async function searchSemantic(q: string) {
    const resp = await memoryApi.searchSemantic(q)
    memories.value = resp.data
    total.value = resp.data.length
  }

  return { memories, total, currentLayer, searchQuery, fetchMemories, createMemory, deleteMemory, searchKeyword, searchSemantic }
})
