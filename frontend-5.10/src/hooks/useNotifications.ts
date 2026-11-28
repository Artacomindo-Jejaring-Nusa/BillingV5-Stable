import { useCallback, useEffect, useRef, useState } from 'react'
import apiClient from '@/services/api'
import type { User } from '@/stores/auth'

export interface AppNotification {
  id: string | number
  type: 'new_payment' | 'new_customer_for_noc' | 'new_customer' | 'new_technical_data' | 'new_chat' | string
  message?: string
  action?: string
  created_at?: string
  timestamp?: string
  read?: boolean
  data?: {
    invoice_number?: string
    pelanggan_nama?: string
    sender_name?: string
    message?: string
    room_id?: number
    [key: string]: unknown
  }
}

export interface NotificationToast {
  id: string | number
  type: string
  title: string
  message: string
}

const VALID_WS_TYPES = new Set([
  'new_payment',
  'new_technical_data',
  'new_customer_for_noc',
  'new_customer',
])

export function getNotificationTitle(type: string): string {
  switch (type) {
    case 'new_payment':
      return 'Pembayaran Diterima'
    case 'new_customer_for_noc':
    case 'new_customer':
      return 'Pelanggan Baru'
    case 'new_technical_data':
      return 'Data Teknis Baru'
    case 'new_chat':
      return 'Pesan Live Chat Baru'
    default:
      return 'Notifikasi'
  }
}

export function getNotificationMessage(notif: AppNotification): string {
  switch (notif.type) {
    case 'new_payment':
      return `${notif.data?.invoice_number || 'N/A'} dari ${notif.data?.pelanggan_nama || 'N/A'} telah lunas`
    case 'new_customer_for_noc':
    case 'new_customer':
      return `${notif.data?.pelanggan_nama || 'N/A'} perlu dibuatkan Data Teknis`
    case 'new_technical_data':
      return `Data teknis untuk ${notif.data?.pelanggan_nama || 'N/A'} telah ditambahkan`
    case 'new_chat':
      return `${notif.data?.sender_name || 'Pelanggan'}: ${notif.data?.message || notif.message || 'Pesan baru'}`
    default:
      return notif.message || 'Anda memiliki notifikasi baru'
  }
}

export function getNotificationRoute(type: string): string {
  switch (type) {
    case 'new_technical_data':
      return '/langganan'
    case 'new_customer_for_noc':
    case 'new_customer':
      return '/data-teknis'
    case 'new_payment':
      return '/invoices'
    case 'new_chat':
      return '/customer-chat'
    default:
      return '/'
  }
}

export function formatNotificationTime(dateString?: string): string {
  if (!dateString) return 'Baru saja'
  try {
    const date = new Date(dateString)
    if (Number.isNaN(date.getTime())) return 'Baru saja'
    const diffMs = Date.now() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Baru saja'
    if (diffMins < 60) return `${diffMins} menit lalu`
    if (diffHours < 24) return `${diffHours} jam lalu`
    if (diffDays < 7) return `${diffDays} hari lalu`
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return 'Baru saja'
  }
}

function isValidNotificationItem(notif: AppNotification): boolean {
  if (notif.action && notif.action.includes('/auth/')) return false
  if (!VALID_WS_TYPES.has(notif.type) && notif.type !== 'new_chat') return false
  if (notif.type === 'new_payment' && !notif.data?.invoice_number) return false
  if ((notif.type === 'new_customer_for_noc' || notif.type === 'new_customer') && !notif.data?.pelanggan_nama) {
    return false
  }
  if (notif.type === 'new_technical_data' && !notif.data?.pelanggan_nama) return false
  return true
}

function fallbackBeep() {
  try {
    const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    if (!AudioCtx) return
    const ctx = new AudioCtx()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(ctx.destination)
    osc.frequency.value = 800
    osc.type = 'sine'
    gain.gain.setValueAtTime(0.25, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.45)
    osc.start(ctx.currentTime)
    osc.stop(ctx.currentTime + 0.45)
  } catch {
    // ignore audio fallback error
  }
}

export function useNotifications(params: {
  token: string | null
  user: User | null
  canViewChat: boolean
  currentPathname: string
  onRefreshBadges: () => void
}) {
  const { token, user, canViewChat, currentPathname, onRefreshBadges } = params
  const [notifications, setNotifications] = useState<AppNotification[]>([])
  const [toast, setToast] = useState<NotificationToast | null>(null)

  const audioCtxRef = useRef<AudioContext | null>(null)
  const pathnameRef = useRef(currentPathname)
  const refreshBadgesRef = useRef(onRefreshBadges)

  useEffect(() => {
    pathnameRef.current = currentPathname
  }, [currentPathname])

  useEffect(() => {
    refreshBadgesRef.current = onRefreshBadges
  }, [onRefreshBadges])

  // Initialize AudioContext on first user interaction
  useEffect(() => {
    const initAudio = () => {
      if (audioCtxRef.current) return
      try {
        const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
        if (AudioCtx) {
          audioCtxRef.current = new AudioCtx()
        }
      } catch {
        // ignore
      }
    }
    window.addEventListener('click', initAudio, { once: true })
    window.addEventListener('keydown', initAudio, { once: true })
    return () => {
      window.removeEventListener('click', initAudio)
      window.removeEventListener('keydown', initAudio)
    }
  }, [])

  const playSound = useCallback((type: string) => {
    try {
      if (audioCtxRef.current && audioCtxRef.current.state === 'suspended') {
        void audioCtxRef.current.resume()
      }
      let audioFile = '/pembayaran.mp3'
      switch (type) {
        case 'new_payment':
          audioFile = '/pembayaran.mp3'
          break
        case 'new_customer_for_noc':
        case 'new_customer':
          audioFile = '/payment.mp3'
          break
        case 'new_technical_data':
          audioFile = '/noc_finance.mp3'
          break
        case 'new_chat':
          audioFile = '/chat-notification.mp3'
          break
        default:
          audioFile = '/pembayaran.mp3'
      }
      const audio = new Audio(audioFile)
      audio.addEventListener('error', () => fallbackBeep())
      const p = audio.play()
      if (p) {
        p.catch(() => fallbackBeep())
      }
    } catch {
      fallbackBeep()
    }
  }, [])

  const pushIncomingNotification = useCallback(
    (item: AppNotification) => {
      setNotifications((prev) => [item, ...prev].slice(0, 20))
      setToast({
        id: item.id,
        type: item.type,
        title: getNotificationTitle(item.type),
        message: getNotificationMessage(item),
      })
      playSound(item.type)
      refreshBadgesRef.current()
    },
    [playSound]
  )

  useEffect(() => {
    if (!toast) return
    const timer = setTimeout(() => setToast(null), 6000)
    return () => clearTimeout(timer)
  }, [toast])

  // Fetch initial unread notifications
  const fetchUnread = useCallback(async () => {
    if (!token) return
    try {
      const res = await apiClient.get('/notifications/unread', { headers: { 'X-Silent': 'true' } })
      const raw: AppNotification[] = res.data?.notifications || []
      const filtered = raw
        .map((n) => (n.type === 'new_customer' ? { ...n, type: 'new_customer_for_noc' } : n))
        .filter(isValidNotificationItem)
        .slice(0, 20)
      setNotifications(filtered)
    } catch {
      // ignore silent unread fetch failure
    }
  }, [token])

  useEffect(() => {
    void fetchUnread()
  }, [fetchUnread])

  // Main WebSocket (/ws/notifications)
  useEffect(() => {
    if (!token || token.length < 10) return

    let socket: WebSocket | null = null
    let pingTimer: ReturnType<typeof setInterval> | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let retryCount = 0
    let disposed = false

    const connect = () => {
      if (disposed || retryCount >= 100) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const encodedToken = encodeURIComponent(token)
      const wsUrl = import.meta.env.DEV
        ? `${protocol}//${window.location.hostname}:8000/ws/notifications?token=${encodedToken}`
        : `${protocol}//${window.location.host}/ws/notifications?token=${encodedToken}`

      socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        retryCount = 0
        if (pingTimer) clearInterval(pingTimer)
        pingTimer = setInterval(() => {
          if (socket?.readyState === WebSocket.OPEN) {
            socket.send('ping')
          }
        }, 30000)
      }

      socket.onmessage = (event) => {
        if (!event.data || event.data === 'pong' || event.data === 'ping') return
        try {
          const raw = typeof event.data === 'string' ? JSON.parse(event.data) : event.data
          if (!raw || typeof raw !== 'object') return
          if (raw.type === 'ping' || raw.type === 'pong') return

          const notif: AppNotification = {
            ...raw,
            id: raw.id || `${Date.now()}-${Math.floor(Math.random() * 10000)}`,
            type: raw.type === 'new_customer' ? 'new_customer_for_noc' : raw.type,
            created_at: raw.created_at || raw.timestamp || new Date().toISOString(),
            data: raw.data || {},
          }
          if (!isValidNotificationItem(notif)) return

          pushIncomingNotification(notif)
          window.dispatchEvent(new CustomEvent('new-notification', { detail: notif }))
        } catch {
          // ignore malformed ws message
        }
      }

      socket.onerror = () => {
        socket?.close()
      }

      socket.onclose = (ev) => {
        if (pingTimer) {
          clearInterval(pingTimer)
          pingTimer = null
        }
        if (disposed || ev.code === 1008 || ev.code === 4001) return
        retryCount += 1
        const delay = Math.min(30000, 1000 * Math.pow(2, Math.min(retryCount, 5)))
        reconnectTimer = setTimeout(connect, delay)
      }
    }

    connect()

    return () => {
      disposed = true
      if (pingTimer) clearInterval(pingTimer)
      if (reconnectTimer) clearTimeout(reconnectTimer)
      if (socket) {
        socket.onclose = null
        socket.close(1000, 'Unmount')
      }
    }
  }, [token, pushIncomingNotification])

  // Live Chat WebSocket listener for incoming customer messages across the app
  useEffect(() => {
    if (!token || !user?.id || !canViewChat) return

    let chatWs: WebSocket | null = null
    let pingTimer: ReturnType<typeof setInterval> | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let disposed = false

    const connectChatWs = () => {
      if (disposed) return
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const userName = encodeURIComponent(user.name || 'Admin')
      const wsUrl = import.meta.env.DEV
        ? `${protocol}//${window.location.hostname}:8000/api/v1/chat/ws?role=admin&user_id=${user.id}&name=${userName}`
        : `${protocol}//${window.location.host}/api/v1/chat/ws?role=admin&user_id=${user.id}&name=${userName}`

      chatWs = new WebSocket(wsUrl)

      chatWs.onopen = () => {
        if (pingTimer) clearInterval(pingTimer)
        pingTimer = setInterval(() => {
          if (chatWs?.readyState === WebSocket.OPEN) {
            chatWs.send(JSON.stringify({ event: 'ping', data: {} }))
          }
        }, 30000)
      }

      chatWs.onmessage = (evt) => {
        try {
          const payload = JSON.parse(evt.data)
          if (payload?.event === 'new_message' && payload?.data?.sender_type === 'customer') {
            const msgData = payload.data
            const isOnChatPage = pathnameRef.current.startsWith('/customer-chat')
            if (isOnChatPage) {
              // Play sound directly on chat page without cluttering bell dropdown
              playSound('new_chat')
              return
            }
            const chatNotif: AppNotification = {
              id: `chat-${msgData.id || Date.now()}`,
              type: 'new_chat',
              created_at: msgData.created_at || new Date().toISOString(),
              data: {
                sender_name: msgData.sender_name || 'Pelanggan',
                message: msgData.message || (msgData.attachment_url ? 'Mengirim lampiran' : 'Pesan baru'),
                room_id: msgData.room_id,
              },
            }
            pushIncomingNotification(chatNotif)
          }
        } catch {
          // ignore parse error
        }
      }

      chatWs.onclose = () => {
        if (pingTimer) {
          clearInterval(pingTimer)
          pingTimer = null
        }
        if (!disposed) {
          reconnectTimer = setTimeout(connectChatWs, 5000)
        }
      }
    }

    connectChatWs()

    return () => {
      disposed = true
      if (pingTimer) clearInterval(pingTimer)
      if (reconnectTimer) clearTimeout(reconnectTimer)
      if (chatWs) {
        chatWs.onclose = null
        chatWs.close(1000, 'Unmount')
      }
    }
  }, [token, user?.id, user?.name, canViewChat, playSound, pushIncomingNotification])

  const markAsRead = useCallback(async (notif: AppNotification) => {
    setNotifications((prev) => prev.filter((n) => n.id !== notif.id))
    if (String(notif.id).startsWith('chat-')) return
    try {
      await apiClient.post(`/notifications/${notif.id}/mark-as-read`, undefined, {
        headers: { 'X-Silent': 'true' },
      })
    } catch {
      // ignore error
    }
  }, [])

  const markAllAsRead = useCallback(async () => {
    setNotifications([])
    try {
      await apiClient.post('/notifications/mark-all-as-read', undefined, {
        headers: { 'X-Silent': 'true' },
      })
    } catch {
      // ignore error
    }
  }, [])

  return {
    notifications,
    toast,
    dismissToast: () => setToast(null),
    markAsRead,
    markAllAsRead,
  }
}
