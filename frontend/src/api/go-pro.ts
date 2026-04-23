import api from './go-client'

export const proApi = {
  // 授权
  getLicenseInfo: () => api.get('/license/info'),
  activate: (licenseKey: string) =>
    api.post('/license/activate', { license_key: licenseKey }),
  deactivate: () => api.post('/license/deactivate'),

  // 安装 Pro 模块（Go 版不需要安装）
  installProModule: () => Promise.resolve({ data: { installed: true } }),
  getInstallStatus: () => Promise.resolve({ data: { installed: true } }),

  // 衰减（Pro）
  getDecayStats: () => api.get('/stats/decay'),
  decayBatch: () => Promise.reject(new Error('Not implemented')),
  getStageInfo: (_stage: number) => Promise.resolve({ data: { label: 'unknown' } }),

  // 冲突检测（Pro）
  scanConflicts: () => Promise.reject(new Error('Not implemented')),
  resolveConflict: (_id1: number, _id2: number) => Promise.reject(new Error('Not implemented')),

  // 智能路由（Pro）
  routeModel: (_text: string) => Promise.reject(new Error('Not implemented')),
  getTokenStats: () => Promise.reject(new Error('Not implemented')),
}
