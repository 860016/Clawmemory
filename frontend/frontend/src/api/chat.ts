import api from './client'

export const chatApi = {
  listSessions: () => api.get('/chat/sessions'),
  createSession: (data: { agent_id: number; title?: string }) => api.post('/chat/sessions', data),
  getMessages: (sessionId: number, limit?: number) => api.get(`/chat/sessions/${sessionId}/messages`, { params: { limit } }),
  deleteSession: (sessionId: number) => api.delete(`/chat/sessions/${sessionId}`),
}
