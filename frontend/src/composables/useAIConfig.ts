import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { aiApi } from '../api/go-ai'

export interface AIConfigState {
  configured: boolean
  providerId: string
  providerName: string
  model: string
  loading: boolean
}

let _cachedConfig: AIConfigState | null = null

export function useAIConfig(autoCheck = true) {
  const router = useRouter()
  const configured = ref(false)
  const providerId = ref('')
  const providerName = ref('')
  const model = ref('')
  const loading = ref(false)

  async function checkAIConfig(): Promise<boolean> {
    loading.value = true
    try {
      const { data } = await aiApi.getConfig()
      const hasProvider = !!(data && data.provider_id)
      configured.value = hasProvider
      providerId.value = data?.provider_id || ''
      providerName.value = data?.provider_name || ''
      model.value = data?.model || ''
      _cachedConfig = { configured: hasProvider, providerId: providerId.value, providerName: providerName.value, model: model.value, loading: false }
      return hasProvider
    } catch {
      configured.value = false
      _cachedConfig = { configured: false, providerId: '', providerName: '', model: '', loading: false }
      return false
    } finally {
      loading.value = false
    }
  }

  async function requireAIConfig(featureName?: string): Promise<boolean> {
    const isConfigured = await checkAIConfig()
    if (!isConfigured) {
      const label = featureName ? `"${featureName}"` : ''
      try {
        await ElMessageBox.confirm(
          `该功能需要先配置 AI 提供商才能使用。${label}是否前往设置页配置？`,
          'AI 未配置',
          { confirmButtonText: '去配置', cancelButtonText: '取消', type: 'warning' }
        )
        router.push({ path: '/settings', query: { section: 'ai' } })
      } catch {
        // user cancelled
      }
      return false
    }
    return true
  }

  function invalidateCache() {
    _cachedConfig = null
  }

  if (autoCheck) {
    onMounted(() => checkAIConfig())
  }

  return {
    configured,
    providerId,
    providerName,
    model,
    loading,
    checkAIConfig,
    requireAIConfig,
    invalidateCache,
  }
}

export function getCachedAIConfig(): AIConfigState | null {
  return _cachedConfig
}
