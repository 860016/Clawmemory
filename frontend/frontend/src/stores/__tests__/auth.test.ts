import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from '../../stores/auth'
import { authApi } from '../../api/auth'

vi.mock('../../api/auth')
const mockedAuthApi = vi.mocked(authApi)

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('should initialize with default state', () => {
    const store = useAuthStore()
    expect(store.user).toBeNull()
    expect(store.token).toBeNull()
    expect(store.isLoggedIn).toBe(false)
  })

  it('should login successfully', async () => {
    const store = useAuthStore()
    mockedAuthApi.login.mockResolvedValue({
      data: { access_token: 'test-token', token_type: 'bearer', role: 'admin', username: 'admin' },
    } as any)

    await store.login('admin', 'password')

    expect(store.token).toBe('test-token')
    expect(localStorage.getItem('token')).toBe('test-token')
  })

  it('should logout and clear state', () => {
    const store = useAuthStore()
    localStorage.setItem('token', 'test-token')
    store.token = 'test-token'

    store.logout()

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('should checkAuth with stored token', async () => {
    localStorage.setItem('token', 'stored-token')
    mockedAuthApi.getMe.mockResolvedValue({
      data: { id: 1, username: 'admin', role: 'admin', is_active: true },
    } as any)

    const store = useAuthStore()
    await store.checkAuth()

    expect(store.token).toBe('stored-token')
    expect(store.user).toBeTruthy()
  })

  it('should clear token if checkAuth fails', async () => {
    localStorage.setItem('token', 'bad-token')
    mockedAuthApi.getMe.mockRejectedValue(new Error('Unauthorized'))

    const store = useAuthStore()
    await store.checkAuth()

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('should change password successfully', async () => {
    const store = useAuthStore()
    mockedAuthApi.changePassword.mockResolvedValue({
      data: { message: 'Password changed successfully' },
    } as any)

    const result = await store.changePassword('old', 'new123')

    expect(mockedAuthApi.changePassword).toHaveBeenCalledWith({
      old_password: 'old',
      new_password: 'new123',
    })
    expect(result).toBe(true)
  })

  it('should handle change password error', async () => {
    const store = useAuthStore()
    mockedAuthApi.changePassword.mockRejectedValue(new Error('Wrong password'))

    const result = await store.changePassword('wrong', 'new123')

    expect(result).toBe(false)
  })
})
