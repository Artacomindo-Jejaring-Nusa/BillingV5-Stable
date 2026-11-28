import { create } from 'zustand'
import apiClient from '@/services/api'

interface MaintenanceState {
  isActive: boolean
  message: string
}

type ThemeMode = 'light' | 'dark'

interface SettingsState {
  maintenanceMode: MaintenanceState
  isLoading: boolean
  theme: ThemeMode
  setTheme: (theme: ThemeMode) => void
  toggleTheme: () => void
  fetchMaintenanceStatus: () => Promise<void>
}

function getInitialTheme(): ThemeMode {
  if (typeof window === 'undefined') return 'light'
  const saved = localStorage.getItem('billing_theme') as ThemeMode | null
  if (saved === 'dark' || saved === 'light') {
    document.documentElement.classList.toggle('dark', saved === 'dark')
    return saved
  }
  return 'light'
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  maintenanceMode: {
    isActive: false,
    message: 'Sistem sedang dalam tahap maintenance.',
  },
  isLoading: false,
  theme: getInitialTheme(),

  setTheme: (theme: ThemeMode) => {
    localStorage.setItem('billing_theme', theme)
    document.documentElement.classList.toggle('dark', theme === 'dark')
    set({ theme })
  },

  toggleTheme: () => {
    const next = get().theme === 'dark' ? 'light' : 'dark'
    get().setTheme(next)
  },

  fetchMaintenanceStatus: async () => {
    set({ isLoading: true })
    try {
      const response = await apiClient.get('/system/settings/maintenance_mode', {
        headers: { 'X-Silent': 'true' },
      })
      const setting = response.data?.data?.value
      if (setting && typeof setting === 'string') {
        const [status, ...msgParts] = setting.split('|')
        set({
          maintenanceMode: {
            isActive: status === 'true',
            message: msgParts.join('|') || 'Sistem sedang dalam tahap maintenance.',
          },
        })
      } else {
        set((s) => ({ maintenanceMode: { ...s.maintenanceMode, isActive: false } }))
      }
    } catch {
      set((s) => ({ maintenanceMode: { ...s.maintenanceMode, isActive: false } }))
    } finally {
      set({ isLoading: false })
    }
  },
}))
