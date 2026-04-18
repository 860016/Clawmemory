import { describe, it, expect, vi, beforeEach } from 'vitest'
import api from '../../api/client'
import { authApi } from '../../api/auth'

vi.mock('../../api/client')
const mockedApi = vi.mocked(api)

describe('auth API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should call getInitStatus endpoint', async () => {
    const mockResponse = { data: { initialized: false } }
    mockedApi.get.mockResolvedValue(mockResponse)

    const result = await authApi.getInitStatus()

    expect(mockedApi.get).toHaveBeenCalledWith('/auth/init-status')
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should call init endpoint', async () => {
    const mockResponse = { data: { username: 'admin', access_token: 'token123' } }
    mockedApi.post.mockResolvedValue(mockResponse)

    const result = await authApi.init({ username: 'admin', password: 'admin123' })

    expect(mockedApi.post).toHaveBeenCalledWith('/auth/init', {
      username: 'admin',
      password: 'admin123',
    })
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should call login endpoint', async () => {
    const mockResponse = { data: { access_token: 'token456', token_type: 'bearer' } }
    mockedApi.post.mockResolvedValue(mockResponse)

    const result = await authApi.login({ username: 'admin', password: 'admin123' })

    expect(mockedApi.post).toHaveBeenCalledWith('/auth/login', {
      username: 'admin',
      password: 'admin123',
    })
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should call getMe endpoint', async () => {
    const mockResponse = { data: { id: 1, username: 'admin', role: 'admin' } }
    mockedApi.get.mockResolvedValue(mockResponse)

    const result = await authApi.getMe()

    expect(mockedApi.get).toHaveBeenCalledWith('/auth/me')
    expect(result.data).toEqual(mockResponse.data)
  })

  it('should call changePassword endpoint', async () => {
    const mockResponse = { data: { message: 'Password changed successfully' } }
    mockedApi.put.mockResolvedValue(mockResponse)

    const result = await authApi.changePassword({
      old_password: 'old',
      new_password: 'new123',
    })

    expect(mockedApi.put).toHaveBeenCalledWith('/auth/change-password', {
      old_password: 'old',
      new_password: 'new123',
    })
    expect(result.data).toEqual(mockResponse.data)
  })
})
