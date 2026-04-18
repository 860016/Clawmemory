import axios from './client'

export default {
  listPages(category?: string) {
    const params: any = {}
    if (category) params.category = category
    return axios.get('/api/v1/wiki/pages', { params })
  },
  getPage(id: number) {
    return axios.get(`/api/v1/wiki/pages/${id}`)
  },
  getTree() {
    return axios.get('/api/v1/wiki/pages/tree')
  },
  createPage(data: any) {
    return axios.post('/api/v1/wiki/pages', data)
  },
  updatePage(id: number, data: any) {
    return axios.put(`/api/v1/wiki/pages/${id}`, data)
  },
  deletePage(id: number) {
    return axios.delete(`/api/v1/wiki/pages/${id}`)
  },
  search(query: string, limit = 20) {
    return axios.get('/api/v1/wiki/search', { params: { q: query, limit } })
  },
  getCategories() {
    return axios.get('/api/v1/wiki/categories')
  },
}
