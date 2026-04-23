// Go 后端 API 统一导出
// 使用方式：将原来的 import { xxxApi } from '@/api/xxx' 改为 import { xxxApi } from '@/api/go-index'

export { authApi } from './go-auth'
export { memoryApi } from './go-memories'
export { knowledgeApi } from './go-knowledge'
export { wikiApi } from './go-wiki'
export { backupApi } from './go-backups'
export { proApi } from './go-pro'
export { reportApi } from './go-reports'
