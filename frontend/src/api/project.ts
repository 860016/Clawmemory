import api from './go-client'
import type { ProjectCreateParams } from './types'

export default {
  list(params?: { page?: number; size?: number; category?: string; status?: string }) {
    return api.get('/projects', { params })
  },
  get(id: number) {
    return api.get(`/projects/${id}`)
  },
  create(data: ProjectCreateParams) {
    return api.post('/projects', data)
  },
  update(id: number, data: Partial<ProjectCreateParams>) {
    return api.put(`/projects/${id}`, data)
  },
  delete(id: number) {
    return api.delete(`/projects/${id}`)
  },
  search(query: string, limit = 20) {
    return api.get('/projects/search', { params: { q: query, limit } })
  },
  categories() {
    return api.get('/projects/categories')
  },
  context(name: string) {
    return api.get('/projects/context', { params: { name } })
  },
  getNotes(projectId: number) {
    return api.get(`/projects/${projectId}/notes`)
  },
  addNote(projectId: number, data: { content: string }) {
    return api.post(`/projects/${projectId}/notes`, data)
  },
  updateNote(noteId: number, data: { content?: string }) {
    return api.put(`/projects/notes/${noteId}`, data)
  },
  deleteNote(noteId: number) {
    return api.delete(`/projects/notes/${noteId}`)
  },
  extractMemories(projectId: number) {
    return api.post(`/projects/${projectId}/extract-memories`)
  },
  discover() {
    return api.post('/projects/discover')
  },
  generateWiki(projectId: number) {
    return api.post(`/projects/${projectId}/generate-wiki`)
  },
}
