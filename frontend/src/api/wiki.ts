import axios from './client'

export default {
  listPages(params?: any) {
    return axios.get('/wiki', { params })
  },
  getPage(id: number) {
    return axios.get(`/wiki/${id}`)
  },
  getTree() {
    return axios.get('/wiki/tree')
  },
  createPage(data: any) {
    return axios.post('/wiki', data)
  },
  updatePage(id: number, data: any) {
    return axios.put(`/wiki/${id}`, data)
  },
  deletePage(id: number) {
    return axios.delete(`/wiki/${id}`)
  },
  search(query: string, limit = 20) {
    return axios.get('/wiki/search', { params: { q: query, limit } })
  },
  getCategories() {
    return axios.get('/wiki/categories')
  },
  getStats() {
    return axios.get('/wiki/stats')
  },
  getConfig() {
    return axios.get('/wiki/config')
  },
  extractFromConversation(conversation: string, is_complete: boolean = true) {
    return axios.post('/wiki/ai/extract', { conversation, is_complete })
  },
  refinePage(id: number, additional_context: string = '') {
    return axios.post(`/wiki/${id}/refine`, { additional_context })
  },
  markComplete(id: number) {
    return axios.post(`/wiki/${id}/mark-complete`)
  },
  markInProgress(id: number) {
    return axios.post(`/wiki/${id}/mark-in-progress`)
  },
}
