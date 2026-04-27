import axios from './client'

export default {
  list(params?: any) {
    return axios.get('/projects', { params })
  },
  get(id: number) {
    return axios.get(`/projects/${id}`)
  },
  create(data: any) {
    return axios.post('/projects', data)
  },
  update(id: number, data: any) {
    return axios.put(`/projects/${id}`, data)
  },
  delete(id: number) {
    return axios.delete(`/projects/${id}`)
  },
  search(query: string, limit = 20) {
    return axios.get('/projects/search', { params: { q: query, limit } })
  },
  categories() {
    return axios.get('/projects/categories')
  },
  context(name: string) {
    return axios.get('/projects/context', { params: { name } })
  },
  getNotes(projectId: number) {
    return axios.get(`/projects/${projectId}/notes`)
  },
  addNote(projectId: number, data: any) {
    return axios.post(`/projects/${projectId}/notes`, data)
  },
  updateNote(noteId: number, data: any) {
    return axios.put(`/projects/notes/${noteId}`, data)
  },
  deleteNote(noteId: number) {
    return axios.delete(`/projects/notes/${noteId}`)
  },
  extractMemories(projectId: number) {
    return axios.post(`/projects/${projectId}/extract-memories`)
  },
}
