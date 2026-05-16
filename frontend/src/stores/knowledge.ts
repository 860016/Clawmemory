import { defineStore } from 'pinia'
import { ref } from 'vue'
import { knowledgeApi } from '../api/go-knowledge'

export const useKnowledgeStore = defineStore('knowledge', () => {
  const entities = ref<any[]>([])
  const relations = ref<any[]>([])
  const graphData = ref<{ nodes: any[]; edges: any[] }>({ nodes: [], edges: [] })
  const entityTotal = ref(0)
  const relationTotal = ref(0)

  async function fetchEntities(page = 1, size = 20) {
    const resp = await knowledgeApi.listEntities({ page, size })
    entities.value = resp.data.items ?? resp.data
    entityTotal.value = resp.data.total ?? entities.value.length
  }

  async function createEntity(data: { name: string; entity_type: string; properties?: Record<string, any> }) {
    const resp = await knowledgeApi.createEntity(data)
    entities.value.unshift(resp.data)
    return resp.data
  }

  async function updateEntity(id: number, data: { name?: string; entity_type?: string; properties?: Record<string, any> }) {
    const resp = await knowledgeApi.updateEntity(id, data)
    const idx = entities.value.findIndex(e => e.id === id)
    if (idx >= 0) entities.value[idx] = resp.data
    return resp.data
  }

  async function deleteEntity(id: number) {
    await knowledgeApi.deleteEntity(id)
    entities.value = entities.value.filter(e => e.id !== id)
  }

  async function fetchRelations(page = 1, size = 20) {
    const resp = await knowledgeApi.listRelations()
    const items = resp.data.items ?? resp.data
    const start = (page - 1) * size
    relations.value = items.slice(start, start + size)
    relationTotal.value = items.length
  }

  async function createRelation(data: { source_id: number; target_id: number; relation_type: string; properties?: Record<string, any> }) {
    const resp = await knowledgeApi.createRelation(data)
    relations.value.unshift(resp.data)
    return resp.data
  }

  async function deleteRelation(id: number) {
    await knowledgeApi.deleteRelation(id)
    relations.value = relations.value.filter(r => r.id !== id)
  }

  async function fetchGraphData(depth?: number) {
    const resp = await knowledgeApi.getGraph()
    graphData.value = resp.data
  }

  return {
    entities, relations, graphData, entityTotal, relationTotal,
    fetchEntities, createEntity, updateEntity, deleteEntity,
    fetchRelations, createRelation, deleteRelation, fetchGraphData,
  }
})
