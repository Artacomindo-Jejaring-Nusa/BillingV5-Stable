import { create } from 'zustand'
import apiClient from '@/services/api'
import { getEncryptedToken, removeEncryptedToken, setEncryptedToken } from '@/utils/crypto'

export interface Permission {
  name: string
}

export interface Role {
  name: string
  permissions: Permission[]
}

export interface User {
  id: number
  email: string
  name: string
  role: Role
}

interface AuthState {
  token: string | null
  user: User | null
  isLoggingOut: boolean
  isInitialized: boolean
  setToken: (newToken: string) => void
  hasPermission: (permissionName: string) => boolean
  hasAnyPermission: (permissions: string[]) => boolean
  hasRole: (roles: string[]) => boolean
  login: (email: string, password: string) => Promise<boolean>
  logout: () => Promise<void>
  verifyToken: () => Promise<boolean>
  refreshToken: () => Promise<boolean>
  initializeAuth: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: getEncryptedToken('access_token'),
  user: null,
  isLoggingOut: false,
  isInitialized: false,

  setToken: (newToken: string) => {
    setEncryptedToken('access_token', newToken)
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
    set({ token: newToken })
  },

  hasPermission: (permissionName: string) => {
    const user = get().user
    if (!user?.role?.permissions) return false
    return user.role.permissions.some((p) => p.name === permissionName)
  },

  hasAnyPermission: (permissions: string[]) => {
    const user = get().user
    if (!user?.role?.permissions) return false
    return permissions.some((perm) => user.role.permissions.some((p) => p.name === perm))
  },

  hasRole: (roles: string[]) => {
    const user = get().user
    if (!user?.role?.name) return false
    const roleLower = user.role.name.toLowerCase()
    if (roleLower === 'superadmin' || roleLower === 'super admin') return true
    return roles.some((r) => r.toLowerCase() === roleLower)
  },

  verifyToken: async () => {
    const token = get().token
    if (!token) return false
    try {
      apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`
      const response = await apiClient.get('/users/me')
      const userData = response.data?.data ?? response.data
      set({ user: userData })
      return true
    } catch (error) {
      console.error('Token verification failed:', error)
      await get().logout()
      return false
    }
  },

  login: async (email: string, password: string) => {
    try {
      const formData = new URLSearchParams()
      formData.append('username', email)
      formData.append('password', password)

      const response = await apiClient.post('/auth/token', formData, {
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      })

      const { access_token, refresh_token } = response.data
      get().setToken(access_token)
      if (refresh_token) {
        setEncryptedToken('refresh_token', refresh_token)
      }
      return await get().verifyToken()
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  },

  logout: async () => {
    set({ isLoggingOut: true })
    const refreshToken = getEncryptedToken('refresh_token')
    try {
      if (refreshToken) {
        await apiClient.post('/auth/logout', { refresh_token: refreshToken })
      }
    } catch (error) {
      console.error('Logout failed on backend, proceeding with frontend logout:', error)
    } finally {
      removeEncryptedToken('access_token')
      removeEncryptedToken('refresh_token')
      delete apiClient.defaults.headers.common['Authorization']
      set({ token: null, user: null })
      setTimeout(() => set({ isLoggingOut: false }), 1000)
    }
  },

  refreshToken: async () => {
    try {
      const refreshToken = getEncryptedToken('refresh_token')
      if (!refreshToken) return false
      const response = await apiClient.post('/auth/refresh', { refresh_token: refreshToken })
      const { access_token, refresh_token } = response.data
      get().setToken(access_token)
      if (refresh_token) {
        setEncryptedToken('refresh_token', refresh_token)
      }
      return true
    } catch (error) {
      console.error('[Auth] Token refresh failed:', error)
      return false
    }
  },

  initializeAuth: async () => {
    const token = get().token
    if (token) {
      await get().verifyToken()
    }
    set({ isInitialized: true })
  },
}))
