import api from './go-client'

export default {
  list(params?: any) {
    return api.get('/projects', { params })
  },
  get(id: number) {
    return api.get(`/projects/${id}`)
  },
  create(data: any) {
    return api.post('/projects', data)
  },
  update(id: number, data: any) {
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
  addNote(projectId: number, data: any) {
    return api.post(`/projects/${projectId}/notes`, data)
  },
  updateNote(noteId: number, data: any) {
    return api.put(`/projects/notes/${noteId}`, data)
  },
  deleteNote(noteId: number) {
    return api.delete(`/projects/notes/${noteId}`)
  },
  extractMemories(projectId: number) {
    return api.post(`/projects/${projectId}/extract-memories`)
  },
}
