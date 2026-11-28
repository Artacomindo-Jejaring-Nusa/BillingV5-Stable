import axios, { type AxiosRequestConfig } from 'axios'
import { getEncryptedToken, removeEncryptedToken, setEncryptedToken } from '@/utils/crypto'
import { errorStorage, networkErrorRateLimit } from './errorStorage'

type NavigateFn = (to: string) => void
let navigateHandler: NavigateFn | null = null

export function setApiNavigateHandler(fn: NavigateFn) {
  navigateHandler = fn
}

function redirectTo(path: string) {
  if (navigateHandler) {
    navigateHandler(path)
  } else if (typeof window !== 'undefined') {
    window.location.href = path
  }
}

let activeMutations = 0
const mutationListeners = new Set<(loading: boolean) => void>()

export function subscribeMutationLoading(listener: (loading: boolean) => void): () => void {
  mutationListeners.add(listener)
  return () => mutationListeners.delete(listener)
}

function notifyMutationListeners() {
  const isLoading = activeMutations > 0
  mutationListeners.forEach((fn) => fn(isLoading))
}

interface TrackedConfig extends AxiosRequestConfig {
  _trackedMutation?: boolean
  _retry?: boolean
}

function shouldTrackMutation(config?: TrackedConfig): boolean {
  const method = (config?.method || '').toLowerCase()
  if (!['post', 'put', 'patch', 'delete'].includes(method)) return false
  const headers = config?.headers as Record<string, unknown> | undefined
  if (headers?.['X-Silent'] || headers?.['X-Skip-Loading']) return false
  const url = String(config?.url || '')
  const skipPrefixes = ['/auth/refresh', '/auth/login', '/calculator', '/langganan/calculate', '/customer-chat', '/notifications']
  return !skipPrefixes.some((p) => url.includes(p))
}

function finishMutation(config?: TrackedConfig) {
  if (config?._trackedMutation) {
    config._trackedMutation = false
    activeMutations = Math.max(0, activeMutations - 1)
    notifyMutationListeners()
  }
}

export const getBaseURL = (): string => {
  const envUrl = import.meta.env.VITE_API_BASE_URL

  if (import.meta.env.DEV) {
    if (envUrl) return envUrl
    const protocol = window.location.protocol
    const hostname = window.location.hostname
    return `${protocol}//${hostname}:8000/api/v1`
  }

  if (envUrl && envUrl.startsWith('http')) {
    try {
      const urlObj = new URL(envUrl)
      if (window.location.hostname !== urlObj.hostname) {
        return '/api/v1'
      }
    } catch {
      // ignore URL parse error
    }
    return envUrl
  }

  return envUrl || '/api/v1'
}

const apiClient = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000,
})

apiClient.interceptors.request.use(
  (config) => {
    const token = getEncryptedToken('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    const tracked = config as TrackedConfig
    if (shouldTrackMutation(tracked) && !tracked._trackedMutation) {
      tracked._trackedMutation = true
      activeMutations++
      notifyMutationListeners()
    }
    return config
  },
  (error) => Promise.reject(error)
)

apiClient.interceptors.response.use(
  (response) => {
    finishMutation(response.config as TrackedConfig)
    return response
  },
  async (error) => {
    const originalRequest = error.config as TrackedConfig | undefined
    finishMutation(originalRequest)

    if (
      error.response?.status === 401 &&
      !originalRequest?._retry &&
      !originalRequest?.url?.includes('/auth/refresh')
    ) {
      if (originalRequest) {
        originalRequest._retry = true
      }

      const refreshToken = getEncryptedToken('refresh_token')
      if (!refreshToken) {
        removeEncryptedToken('access_token')
        removeEncryptedToken('refresh_token')
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          redirectTo('/login')
        }
        return Promise.reject(error)
      }

      try {
        const response = await apiClient.post('/auth/refresh', {
          refresh_token: refreshToken,
        })
        const { access_token, refresh_token } = response.data
        setEncryptedToken('access_token', access_token)
        if (refresh_token) {
          setEncryptedToken('refresh_token', refresh_token)
        }
        if (originalRequest?.headers) {
          (originalRequest.headers as Record<string, string>).Authorization = `Bearer ${access_token}`
        }
        return apiClient(originalRequest!)
      } catch (refreshError) {
        removeEncryptedToken('access_token')
        removeEncryptedToken('refresh_token')
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
          redirectTo('/login')
        }
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error?.config as TrackedConfig | undefined
    if (!originalRequest) return Promise.reject(error)

    const headers = originalRequest.headers as Record<string, unknown> | undefined
    const pathname = typeof window !== 'undefined' ? window.location.pathname : ''
    const href = typeof window !== 'undefined' ? window.location.href : ''

    const isBypassed =
      Boolean(headers?.['X-Skip-Network-Interceptor']) ||
      Boolean(headers?.['X-Silent']) ||
      pathname.includes('/tv') ||
      pathname.includes('/customer-chat') ||
      pathname.includes('/chat') ||
      href.includes('/tv') ||
      href.includes('/customer-chat') ||
      href.includes('/chat')

    if (isBypassed) return Promise.reject(error)

    if (!networkErrorRateLimit.canRedirect()) {
      return Promise.reject(error)
    }

    if (!error.response && error.code === 'NETWORK_ERROR') {
      const errorId = errorStorage.storeError('network', 0, originalRequest.url, originalRequest.method?.toUpperCase(), error.message)
      redirectTo(`/network-error?errorId=${errorId}`)
      return Promise.reject(error)
    }

    if (error.code === 'ECONNABORTED' && error.message?.includes('timeout')) {
      const errorId = errorStorage.storeError('timeout', 0, originalRequest.url, originalRequest.method?.toUpperCase(), 'Request timeout - server took too long to respond')
      redirectTo(`/network-error?errorId=${errorId}`)
      return Promise.reject(error)
    }

    if (error.response?.status === 503 || error.response?.status === 504) {
      const errorId = errorStorage.storeError('server', error.response.status, originalRequest.url, originalRequest.method?.toUpperCase(), `Server error: ${error.response.status} ${error.response.statusText}`)
      redirectTo(`/network-error?errorId=${errorId}`)
      return Promise.reject(error)
    }

    if (
      error.code === 'ERR_NETWORK' ||
      error.code === 'ERR_CONNECTION_REFUSED' ||
      error.message?.includes('Network Error') ||
      error.message?.includes('fetch failed')
    ) {
      const errorId = errorStorage.storeError('network', 0, originalRequest.url, originalRequest.method?.toUpperCase(), 'Cannot connect to server - server may be down')
      redirectTo(`/network-error?errorId=${errorId}`)
      return Promise.reject(error)
    }

    return Promise.reject(error)
  }
)

export const checkServerHealth = async () => {
  try {
    const response = await apiClient.get('/health', {
      headers: { 'X-Skip-Network-Interceptor': 'true' },
      timeout: 5000,
    })
    return { success: true, status: response.status, data: response.data }
  } catch (error) {
    const err = error as { message?: string; code?: string }
    return {
      success: false,
      error: err.message || 'Health check failed',
      code: err.code || 'UNKNOWN',
    }
  }
}

export const retryRequest = async (originalRequest: AxiosRequestConfig, maxRetries = 3) => {
  let retryCount = 0
  while (retryCount < maxRetries) {
    try {
      return await apiClient({
        ...originalRequest,
        headers: {
          ...originalRequest.headers,
          'X-Skip-Network-Interceptor': 'true',
        },
      })
    } catch (error) {
      retryCount++
      if (retryCount >= maxRetries) throw error
      await new Promise((resolve) => setTimeout(resolve, Math.pow(2, retryCount) * 1000))
    }
  }
}

export const troubleTicketAPI = {
  getTickets: (params?: Record<string, unknown>) => apiClient.get('/trouble-tickets', { params }),
  getTicket: (id: number) => apiClient.get(`/trouble-tickets/${id}`),
  createTicket: (data: Record<string, unknown>) => apiClient.post('/trouble-tickets', data),
  updateTicket: (id: number, data: Record<string, unknown>) => apiClient.patch(`/trouble-tickets/${id}`, data),
  updateStatus: (id: number, data: Record<string, unknown>) => apiClient.post(`/trouble-tickets/${id}/status`, data),
  updateDowntime: (id: number, data: Record<string, unknown>) => apiClient.post(`/trouble-tickets/${id}/downtime`, data),
  assignTicket: (id: number, data: Record<string, unknown>) => apiClient.post(`/trouble-tickets/${id}/assign`, data),
  getHistory: (id: number) => apiClient.get(`/trouble-tickets/${id}/history`),
  getStatistics: () => apiClient.get('/trouble-tickets/statistics/dashboard'),
}

export default apiClient
