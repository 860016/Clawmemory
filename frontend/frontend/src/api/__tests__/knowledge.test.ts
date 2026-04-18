import { describe, it, expect, vi, beforeEach } from 'vitest'
import api from '../../api/client'
import { knowledgeApi } from '../../api/knowledge'

vi.mock('../../api/client')
const mockedApi = vi.mocked(api)

describe('knowledge API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list entities', async () => {
    const mockResponse = { data: [{ id: 1, name: 'Test', entity_type: 'concept' }] }
    mockedApi.get.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.listEntities({ page: 1, size: 20 })

    expect(mockedApi.get).toHaveBeenCalledWith('/knowledge/entities', {
      params: { page: 1, size: 20 },
    })
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should create entity', async () => {
    const payload = { name: 'NewEntity', entity_type: 'concept' }
    const mockResponse = { data: { id: 1, ...payload } }
    mockedApi.post.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.createEntity(payload)

    expect(mockedApi.post).toHaveBeenCalledWith('/knowledge/entities', payload)
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should update entity', async () => {
    const payload = { name: 'Updated' }
    const mockResponse = { data: { id: 1, ...payload } }
    mockedApi.put.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.updateEntity(1, payload)

    expect(mockedApi.put).toHaveBeenCalledWith('/knowledge/entities/1', payload)
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should delete entity', async () => {
    mockedApi.delete.mockResolvedValue({ data: null })

    await knowledgeApi.deleteEntity(1)

    expect(mockedApi.delete).toHaveBeenCalledWith('/knowledge/entities/1')
  })

  it('should list relations', async () => {
    const mockResponse = { data: [{ id: 1, source_id: 1, target_id: 2, relation_type: 'related' }] }
    mockedApi.get.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.listRelations({ page: 1, size: 20 })

    expect(mockedApi.get).toHaveBeenCalledWith('/knowledge/relations', {
      params: { page: 1, size: 20 },
    })
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should create relation', async () => {
    const payload = { source_id: 1, target_id: 2, relation_type: 'related' }
    const mockResponse = { data: { id: 1, ...payload } }
    mockedApi.post.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.createRelation(payload)

    expect(mockedApi.post).toHaveBeenCalledWith('/knowledge/relations', payload)
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should get graph data', async () => {
    const mockResponse = { data: { nodes: [], edges: [] } }
    mockedApi.get.mockResolvedValue(mockResponse)

    const result = await knowledgeApi.getGraphData({ depth: 2 })

    expect(mockedApi.get).toHaveBeenCalledWith('/knowledge/graph', {
      params: { depth: 2 },
    })
    expect(result.data).toEqual(mockResponse.data)
  })
})
