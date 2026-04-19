import axios from './client'

export default {
  listPages(category?: string) {
    const params: any = {}
    if (category) params.category = category
    return axios.get('/wiki/pages', { params })
  },
  getPage(id: number) {
    return axios.get(`/wiki/pages/${id}`)
  },
  getTree() {
    return axios.get('/wiki/pages/tree')
  },
  createPage(data: any) {
    return axios.post('/wiki/pages', data)
  },
  updatePage(id: number, data: any) {
    return axios.put(`/wiki/pages/${id}`, data)
  },
  deletePage(id: number) {
    return axios.delete(`/wiki/pages/${id}`)
  },
  search(query: string, limit = 20) {
    return axios.get('/wiki/search', { params: { q: query, limit } })
  },
  getCategories() {
    return axios.get('/wiki/categories')
  },
}
