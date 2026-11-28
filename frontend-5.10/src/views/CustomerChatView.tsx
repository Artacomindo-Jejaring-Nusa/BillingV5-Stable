import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Bot,
  Check,
  CheckCheck,
  CheckCircle2,
  Clock,
  Copy,
  ExternalLink,
  Image as ImageIcon,
  LockOpen,
  MapPin,
  MessageSquare,
  MessageSquarePlus,
  PanelRightClose,
  PanelRightOpen,
  Pencil,
  Phone,
  Plus,
  RefreshCw,
  Search,
  Send,
  Settings2,
  Smile,
  Sparkles,
  StickyNote,
  Ticket,
  Trash2,
  UserCheck,
  UserMinus,
  Users,
  Volume2,
  VolumeX,
  Wifi,
  X,
  Zap,
} from 'lucide-react'
import apiClient, { troubleTicketAPI } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

interface ChatAdmin {
  id: number
  nama?: string
  name?: string
  email?: string
}

interface ChatCustomer {
  id: number
  nama: string
  customer_id?: string
  no_telp?: string
  alamat?: string
  blok?: string
  unit?: string
  layanan?: string
  harga_layanan?: { nama_layanan?: string }
  langganan?: Array<{ status?: string }>
  data_teknis?: { id?: number; id_pelanggan?: string }
}

interface ChatRoom {
  id: number
  pelanggan_id: number
  brand?: string
  status: string
  last_message_text?: string
  last_message_at?: string
  unread_count_admin: number
  assigned_admin_id?: number | null
  assigned_admin?: ChatAdmin | null
  pelanggan?: ChatCustomer
}

interface ChatMessage {
  id: number | null
  room_id: number
  sender_type: 'admin' | 'customer' | 'system' | string
  sender_name?: string
  message: string
  message_type?: 'text' | 'image' | string
  attachment_url?: string
  status?: 'pending' | 'sent' | 'delivered' | 'read' | 'error' | string
  created_at: string
  temp_id?: string
}

interface QuickTemplate {
  id?: number
  shortcut: string
  title?: string
  label?: string
  content?: string
  text?: string
}

interface CustomerTicket {
  id: number
  ticket_number: string
  title: string
  category: string
  status: string
  created_at: string
}

const DEFAULT_TEMPLATES: QuickTemplate[] = [
  {
    shortcut: 'salam',
    title: 'Salam',
    content: 'Halo, selamat datang di layanan Customer Care Artacom. Ada yang bisa kami bantu?',
  },
  {
    shortcut: 'cekteknis',
    title: 'Cek Teknis',
    content: 'Baik pak/bu, mohon ditunggu sebentar ya. Sedang kami lakukan pengecekan ke tim teknis lapangan.',
  },
  {
    shortcut: 'lunas',
    title: 'Lunas & Aktif',
    content: 'Terima kasih atas konfirmasinya. Tagihan Anda telah terverifikasi dan layanan internet sudah aktif normal kembali.',
  },
  {
    shortcut: 'fotomodem',
    title: 'Foto Modem',
    content: 'Bisa tolong difotokan lampu indikator (PON / LOS / Internet) yang menyala pada perangkat modem router Anda?',
  },
  {
    shortcut: 'restart',
    title: 'Restart Modem',
    content: 'Mohon bantuannya untuk mematikan modem selama 1-2 menit, lalu nyalakan kembali dan kabari kami hasilnya ya.',
  },
]

const EMOJI_LIST = [
  '😀', '😊', '🙏', '👍', '👌', '👋', '🤝', '❤️',
  '🎉', '✅', '⚡', '📶', '🌐', '🔧', '🛠️', '📡',
  '📸', '📄', '💡', '⏳', '🕐', '📞', '💬', '✨',
]

const TICKET_CATEGORIES = [
  { value: 'slow_connection', label: 'Koneksi Lambat' },
  { value: 'no_internet', label: 'Tidak Ada Koneksi / LOS' },
  { value: 'cable_issue', label: 'Masalah Kabel / Putus' },
  { value: 'modem_issue', label: 'Kendala Modem / ONT' },
  { value: 'billing_issue', label: 'Kendala Tagihan / Pembayaran' },
  { value: 'other', label: 'Lainnya' },
]

const TICKET_PRIORITIES = [
  { value: 'low', label: 'Rendah (Low)' },
  { value: 'medium', label: 'Sedang (Medium)' },
  { value: 'high', label: 'Tinggi (High)' },
  { value: 'urgent', label: 'Darurat (Urgent)' },
]

function getInitials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((n) => n[0]?.toUpperCase() || '')
    .join('')
}

function getBrandShortLabel(brand?: string) {
  if (!brand || brand === 'ALL') return 'Semua'
  const b = brand.toUpperCase()
  if (b.includes('NAGRAK')) return 'Nagrak'
  if (b.includes('JELANTIK')) return 'Jelantik'
  if (b.includes('JAKINET')) return 'Jakinet'
  return brand
}

function getBrandAvatarClass(brand?: string) {
  const b = (brand || '').toUpperCase()
  if (b.includes('JAKINET')) return 'bg-rose-600 text-white'
  if (b.includes('NAGRAK')) return 'bg-emerald-600 text-white'
  return 'bg-blue-600 text-white'
}

function getFullMediaUrl(url?: string): string {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (import.meta.env.DEV) {
    return `${window.location.protocol}//${window.location.hostname}:8000${url.startsWith('/') ? '' : '/'}${url}`
  }
  return url
}

function formatTimestamp(ts?: string) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })
}

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

function formatDateHeader(ts: string) {
  const d = new Date(ts)
  const now = new Date()
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (d.toDateString() === now.toDateString()) return 'Hari Ini'
  if (d.toDateString() === yesterday.toDateString()) return 'Kemarin'
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
}

function ticketStatusClass(status: string) {
  switch (status) {
    case 'resolved':
    case 'closed':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200'
    case 'in_progress':
      return 'bg-blue-50 text-blue-700 border-blue-200'
    case 'pending':
      return 'bg-amber-50 text-amber-700 border-amber-200'
    default:
      return 'bg-rose-50 text-rose-700 border-rose-200'
  }
}

export function CustomerChatView() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  // Rooms & filters
  const [rooms, setRooms] = useState<ChatRoom[]>([])
  const [isLoadingRooms, setIsLoadingRooms] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedBrand, setSelectedBrand] = useState('ALL')
  const [selectedStatusTab, setSelectedStatusTab] = useState<'open' | 'closed' | 'all'>('open')
  const [selectedAssignment, setSelectedAssignment] = useState<'all' | 'mine' | 'unassigned'>('all')
  const [roomCounts, setRoomCounts] = useState({ open: 0, closed: 0, all: 0, mine: 0, unassigned: 0 })

  // Active room & messages
  const [activeRoom, setActiveRoom] = useState<ChatRoom | null>(null)
  const [activeMessages, setActiveMessages] = useState<ChatMessage[]>([])
  const [isLoadingMessages, setIsLoadingMessages] = useState(false)
  const [isWsConnected, setIsWsConnected] = useState(false)
  const [isCustomerTyping, setIsCustomerTyping] = useState(false)
  const [isAssigning, setIsAssigning] = useState(false)
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false)
  const [isSoundEnabled, setIsSoundEnabled] = useState(true)

  // Input, Slash Menu, Emoji, Mode
  const [inputMessage, setInputMessage] = useState('')
  const [inputMode, setInputMode] = useState<'reply' | 'note'>('reply')
  const [showEmojiPicker, setShowEmojiPicker] = useState(false)
  const [quickTemplates, setQuickTemplates] = useState<QuickTemplate[]>(DEFAULT_TEMPLATES)
  const [showSlashMenu, setShowSlashMenu] = useState(false)
  const [slashQuery, setSlashQuery] = useState('')
  const [selectedSlashIndex, setSelectedSlashIndex] = useState(0)

  // Manage Templates dialog
  const [manageTemplatesOpen, setManageTemplatesOpen] = useState(false)
  const [newTplShortcut, setNewTplShortcut] = useState('')
  const [newTplTitle, setNewTplTitle] = useState('')
  const [newTplContent, setNewTplContent] = useState('')
  const [savingTemplate, setSavingTemplate] = useState(false)

  // Image upload & lightbox
  const [isUploadingImage, setIsUploadingImage] = useState(false)
  const [imageUploadOpen, setImageUploadOpen] = useState(false)
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null)
  const [selectedImagePreviewUrl, setSelectedImagePreviewUrl] = useState<string | null>(null)
  const [imageCaption, setImageCaption] = useState('')
  const [lightboxUrl, setLightboxUrl] = useState<string | null>(null)

  // Close room dialog
  const [closeRoomOpen, setCloseRoomOpen] = useState(false)
  const [sendClosingMessage, setSendClosingMessage] = useState(true)
  const [closingMessageText, setClosingMessageText] = useState('')

  // Panel 3: Contact 360 + Trouble Ticket
  const [showInfoPanel, setShowInfoPanel] = useState(true)
  const [rightPanelTab, setRightPanelTab] = useState<'profile' | 'ticket'>('profile')
  const [customerTickets, setCustomerTickets] = useState<CustomerTicket[]>([])
  const [isLoadingTickets, setIsLoadingTickets] = useState(false)
  const [isAutoFilling, setIsAutoFilling] = useState(false)
  const [isSubmittingTicket, setIsSubmittingTicket] = useState(false)
  const [createdTicketResult, setCreatedTicketResult] = useState<CustomerTicket | null>(null)
  const [ticketForm, setTicketForm] = useState({
    category: 'slow_connection',
    priority: 'medium',
    title: '',
    description: '',
  })

  // Toast feedback
  const [toast, setToast] = useState<{ text: string; type: 'success' | 'error' | 'info' } | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const activeRoomRef = useRef<ChatRoom | null>(null)
  const messagesEndRef = useRef<HTMLDivElement | null>(null)
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const chatInputRef = useRef<HTMLTextAreaElement | null>(null)

  useEffect(() => {
    activeRoomRef.current = activeRoom
  }, [activeRoom])

  const showToast = useCallback((text: string, type: 'success' | 'error' | 'info' = 'success') => {
    setToast({ text, type })
    window.setTimeout(() => setToast(null), 3200)
  }, [])

  const scrollToBottom = useCallback(() => {
    window.setTimeout(() => {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, 50)
  }, [])

  const sendWsEvent = useCallback((event: string, data: Record<string, unknown>) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ event, data }))
    }
  }, [])

  const fetchRoomCounts = useCallback(async () => {
    try {
      const res = await apiClient.get('/chat/rooms/counts', {
        params: {
          admin_id: user?.id || 0,
          brand: selectedBrand !== 'ALL' ? selectedBrand : undefined,
        },
        headers: { 'X-Skip-Network-Interceptor': 'true' },
      })
      if (res.data?.data) {
        setRoomCounts((prev) => ({ ...prev, ...res.data.data }))
      }
    } catch {
      // ignore
    }
  }, [user?.id, selectedBrand])

  const fetchRooms = useCallback(
    async (silent = false) => {
      if (!silent) setIsLoadingRooms(true)
      try {
        const res = await apiClient.get('/chat/rooms', {
          params: { page: 1, page_size: 100 },
          headers: { 'X-Skip-Network-Interceptor': 'true' },
        })
        if (Array.isArray(res.data?.data)) {
          setRooms(res.data.data)
        }
      } catch (err) {
        console.error('Fetch chat rooms error:', err)
      } finally {
        if (!silent) setIsLoadingRooms(false)
      }
      void fetchRoomCounts()
    },
    [fetchRoomCounts],
  )

  const fetchTemplates = useCallback(async () => {
    try {
      const res = await apiClient.get('/chat/templates')
      if (Array.isArray(res.data?.data) && res.data.data.length > 0) {
        setQuickTemplates(res.data.data)
      }
    } catch {
      // fallback to default templates
    }
  }, [])

  const loadCustomerTickets = useCallback(async (pelangganId: number) => {
    setIsLoadingTickets(true)
    try {
      const res = await troubleTicketAPI.getTickets({ pelanggan_id: pelangganId, pageSize: 5 })
      setCustomerTickets(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setCustomerTickets([])
    } finally {
      setIsLoadingTickets(false)
    }
  }, [])

  // WebSocket setup
  useEffect(() => {
    let reconnectTimer: number | null = null
    let heartbeatTimer: number | null = null
    let unmounted = false

    const connect = () => {
      if (unmounted) return
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const userId = user?.id || 1
      const userName = encodeURIComponent(user?.name || 'Admin CS')
      const wsUrl = import.meta.env.DEV
        ? `${protocol}//${window.location.hostname}:8000/api/v1/chat/ws?role=admin&user_id=${userId}&name=${userName}`
        : `${protocol}//${window.location.host}/api/v1/chat/ws?role=admin&user_id=${userId}&name=${userName}`

      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          setIsWsConnected(true)
          heartbeatTimer = window.setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ event: 'ping', data: {} }))
            }
          }, 25000)
          void fetchRooms(true)
        }

        ws.onmessage = (evt) => {
          try {
            const payload = JSON.parse(evt.data)
            const { event, data } = payload
            if (!event || !data) return

            if (event === 'ack_sent' || event === 'ack_message') {
              setActiveMessages((prev) =>
                prev.map((m) =>
                  (data.temp_id && m.temp_id === data.temp_id) || (data.id && m.id === data.id)
                    ? { ...m, id: data.id, status: data.status || 'sent' }
                    : m,
                ),
              )
            } else if (event === 'message_error') {
              showToast(`Gagal mengirim pesan: ${data.error || 'Terjadi kesalahan'}`, 'error')
              setActiveMessages((prev) =>
                prev.map((m) => (data.temp_id && m.temp_id === data.temp_id ? { ...m, status: 'error' } : m)),
              )
            } else if (event === 'new_message') {
              setRooms((prev) => {
                const idx = prev.findIndex((r) => r.id === data.room_id)
                if (idx === -1) {
                  void fetchRooms(true)
                  return prev
                }
                const isCurrent = activeRoomRef.current?.id === data.room_id
                const updated: ChatRoom = {
                  ...prev[idx],
                  last_message_text: data.message || (data.message_type === 'image' ? '📷 Gambar' : ''),
                  last_message_at: data.created_at,
                  status: 'open',
                  unread_count_admin:
                    !isCurrent && data.sender_type === 'customer'
                      ? (prev[idx].unread_count_admin || 0) + 1
                      : prev[idx].unread_count_admin,
                }
                const next = [...prev]
                next.splice(idx, 1)
                return [updated, ...next]
              })

              if (activeRoomRef.current?.id === data.room_id) {
                setActiveRoom((prev) => (prev ? { ...prev, status: 'open' } : prev))
                setActiveMessages((prev) => {
                  const exists = prev.some(
                    (m) => (data.id && m.id === data.id) || (data.temp_id && m.temp_id === data.temp_id),
                  )
                  return exists ? prev : [...prev, data]
                })
                scrollToBottom()
                if (data.sender_type === 'customer') {
                  apiClient.post(`/chat/messages/${data.room_id}/read?reader_type=admin`).catch(() => {})
                  sendWsEvent('read_room', { room_id: data.room_id })
                }
              }
            } else if (event === 'typing') {
              if (activeRoomRef.current?.id === data.room_id && data.sender_type === 'customer') {
                setIsCustomerTyping(Boolean(data.is_typing))
              }
            } else if (event === 'room_status_update') {
              setRooms((prev) => prev.map((r) => (r.id === data.room_id ? { ...r, status: data.status } : r)))
              if (activeRoomRef.current?.id === data.room_id) {
                setActiveRoom((prev) => (prev ? { ...prev, status: data.status } : prev))
              }
              void fetchRoomCounts()
            } else if (event === 'room_assigned') {
              setRooms((prev) =>
                prev.map((r) =>
                  r.id === data.room_id
                    ? { ...r, assigned_admin_id: data.admin_id, assigned_admin: data.room?.assigned_admin || r.assigned_admin }
                    : r,
                ),
              )
              if (activeRoomRef.current?.id === data.room_id) {
                setActiveRoom((prev) =>
                  prev
                    ? { ...prev, assigned_admin_id: data.admin_id, assigned_admin: data.room?.assigned_admin || prev.assigned_admin }
                    : prev,
                )
              }
              void fetchRoomCounts()
            } else if (event === 'room_unassigned') {
              setRooms((prev) =>
                prev.map((r) => (r.id === data.room_id ? { ...r, assigned_admin_id: null, assigned_admin: null } : r)),
              )
              if (activeRoomRef.current?.id === data.room_id) {
                setActiveRoom((prev) => (prev ? { ...prev, assigned_admin_id: null, assigned_admin: null } : prev))
              }
              void fetchRoomCounts()
            }
          } catch {
            // ignore
          }
        }

        ws.onclose = () => {
          setIsWsConnected(false)
          if (heartbeatTimer) window.clearInterval(heartbeatTimer)
          if (!unmounted) {
            reconnectTimer = window.setTimeout(connect, 3000)
          }
        }

        ws.onerror = () => setIsWsConnected(false)
      } catch {
        setIsWsConnected(false)
      }
    }

    connect()
    return () => {
      unmounted = true
      if (heartbeatTimer) window.clearInterval(heartbeatTimer)
      if (reconnectTimer) window.clearTimeout(reconnectTimer)
      wsRef.current?.close()
    }
  }, [user?.id, user?.name, fetchRooms, fetchRoomCounts, scrollToBottom, sendWsEvent, showToast])

  useEffect(() => {
    void fetchRooms()
    void fetchTemplates()
  }, [fetchRooms, fetchTemplates])

  // Select room
  const selectRoom = async (room: ChatRoom) => {
    setActiveRoom(room)
    setIsLoadingMessages(true)
    setActiveMessages([])
    setIsCustomerTyping(false)
    setTicketForm({ category: 'slow_connection', priority: 'medium', title: '', description: '' })

    if (room.pelanggan_id) {
      void loadCustomerTickets(room.pelanggan_id)
    }

    try {
      const res = await apiClient.get(`/chat/messages/${room.id}?limit=200`)
      setActiveMessages(Array.isArray(res.data?.data) ? res.data.data : [])
      if (room.unread_count_admin > 0) {
        setRooms((prev) => prev.map((r) => (r.id === room.id ? { ...r, unread_count_admin: 0 } : r)))
        apiClient.post(`/chat/messages/${room.id}/read?reader_type=admin`).catch(() => {})
        sendWsEvent('read_room', { room_id: room.id })
      }
      scrollToBottom()
    } catch (err) {
      console.error('Fetch messages error:', err)
    } finally {
      setIsLoadingMessages(false)
    }
  }

  // Filtered rooms & counts
  const computedCounts = useMemo(() => {
    const brandFiltered =
      selectedBrand === 'ALL' ? rooms : rooms.filter((r) => (r.brand || '').toUpperCase().includes(selectedBrand))
    const open = roomCounts.open || brandFiltered.filter((r) => r.status !== 'closed').length
    const closed = roomCounts.closed || brandFiltered.filter((r) => r.status === 'closed').length
    const all = roomCounts.all || brandFiltered.length
    const mine = brandFiltered.filter((r) => r.assigned_admin_id === user?.id).length
    const unassigned = brandFiltered.filter((r) => !r.assigned_admin_id).length
    const assigned = brandFiltered.filter((r) => Boolean(r.assigned_admin_id)).length
    return { open, closed, all, mine, unassigned, assigned }
  }, [rooms, selectedBrand, roomCounts, user?.id])

  const totalUnreadCount = useMemo(
    () => rooms.reduce((sum, r) => sum + (r.unread_count_admin || 0), 0),
    [rooms],
  )

  const filteredRooms = useMemo(() => {
    return rooms.filter((r) => {
      if (selectedBrand !== 'ALL' && !(r.brand || '').toUpperCase().includes(selectedBrand)) return false
      if (selectedStatusTab === 'open' && r.status === 'closed') return false
      if (selectedStatusTab === 'closed' && r.status !== 'closed') return false
      if (selectedAssignment === 'mine' && r.assigned_admin_id !== user?.id) return false
      if (selectedAssignment === 'unassigned' && r.assigned_admin_id) return false
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase()
        const name = (r.pelanggan?.nama || '').toLowerCase()
        const phone = (r.pelanggan?.no_telp || '').toLowerCase()
        const msg = (r.last_message_text || '').toLowerCase()
        if (!name.includes(q) && !phone.includes(q) && !msg.includes(q)) return false
      }
      return true
    })
  }, [rooms, selectedBrand, selectedStatusTab, selectedAssignment, searchQuery, user?.id])

  // Slash command filtering
  const filteredSlashTemplates = useMemo(() => {
    if (!slashQuery) return quickTemplates
    return quickTemplates.filter(
      (t) =>
        t.shortcut.toLowerCase().includes(slashQuery) ||
        (t.title || t.label || '').toLowerCase().includes(slashQuery) ||
        (t.content || t.text || '').toLowerCase().includes(slashQuery),
    )
  }, [quickTemplates, slashQuery])

  const sendAdminTextMessage = (rawText: string) => {
    const text = rawText.trim()
    if (!text || !activeRoom) return

    const tempId = `admin_temp_${Date.now()}`
    const localMsg: ChatMessage = {
      id: null,
      room_id: activeRoom.id,
      sender_type: 'admin',
      sender_name: user?.name || 'Admin CS',
      message: text,
      message_type: 'text',
      status: 'pending',
      created_at: new Date().toISOString(),
      temp_id: tempId,
    }

    setActiveMessages((prev) => [...prev, localMsg])
    setActiveRoom((prev) => (prev ? { ...prev, status: 'open', last_message_text: text, last_message_at: localMsg.created_at } : prev))
    setRooms((prev) => {
      const idx = prev.findIndex((r) => r.id === activeRoom.id)
      if (idx === -1) return prev
      const updated = {
        ...prev[idx],
        status: 'open',
        last_message_text: text,
        last_message_at: localMsg.created_at,
        assigned_admin_id: prev[idx].assigned_admin_id || user?.id,
      }
      const next = [...prev]
      next.splice(idx, 1)
      return [updated, ...next]
    })

    scrollToBottom()
    sendWsEvent('send_message', {
      room_id: activeRoom.id,
      message: text,
      temp_id: tempId,
      message_type: 'text',
    })
  }

  const sendAdminMessage = () => {
    let text = inputMessage.trim()
    if (!text || !activeRoom) return
    if (inputMode === 'note') {
      text = `📝 [Catatan Internal] ${text}`
    }
    setInputMessage('')
    setShowSlashMenu(false)
    setShowEmojiPicker(false)
    sendAdminTextMessage(text)
  }

  const handleChatChange = (val: string) => {
    setInputMessage(val)
    if (activeRoom) {
      sendWsEvent('typing', { room_id: activeRoom.id, is_typing: val.trim().length > 0 })
    }
    const lastSlash = val.lastIndexOf('/')
    if (lastSlash !== -1 && (lastSlash === 0 || val[lastSlash - 1] === ' ' || val[lastSlash - 1] === '\n')) {
      const q = val.slice(lastSlash + 1)
      if (!/\s/.test(q)) {
        setSlashQuery(q.toLowerCase())
        setShowSlashMenu(true)
        setSelectedSlashIndex(0)
        return
      }
    }
    setShowSlashMenu(false)
  }

  const applySlashTemplate = (tpl: QuickTemplate) => {
    const content = tpl.content || tpl.text || ''
    const lastSlash = inputMessage.lastIndexOf('/')
    const next = lastSlash !== -1 ? inputMessage.slice(0, lastSlash) + content : content
    setInputMessage(next)
    setShowSlashMenu(false)
    window.setTimeout(() => chatInputRef.current?.focus(), 10)
  }

  const handleChatKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (showSlashMenu && filteredSlashTemplates.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedSlashIndex((i) => (i + 1) % filteredSlashTemplates.length)
        return
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedSlashIndex((i) => (i - 1 + filteredSlashTemplates.length) % filteredSlashTemplates.length)
        return
      }
      if ((e.key === 'Enter' || e.key === 'Tab') && !e.shiftKey) {
        e.preventDefault()
        applySlashTemplate(filteredSlashTemplates[selectedSlashIndex])
        return
      }
      if (e.key === 'Escape') {
        e.preventDefault()
        setShowSlashMenu(false)
        return
      }
    }
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendAdminMessage()
    }
  }

  // Image Upload
  const onImageSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    if (!file.type.startsWith('image/')) {
      showToast('Hanya file gambar yang didukung (JPG, PNG, WEBP, GIF)', 'error')
      return
    }
    if (file.size > 25 * 1024 * 1024) {
      showToast('Ukuran gambar maksimal 25MB', 'error')
      return
    }
    setSelectedImageFile(file)
    setSelectedImagePreviewUrl(URL.createObjectURL(file))
    setImageCaption('')
    setImageUploadOpen(true)
  }

  const cancelImageUpload = () => {
    setImageUploadOpen(false)
    if (selectedImagePreviewUrl) URL.revokeObjectURL(selectedImagePreviewUrl)
    setSelectedImagePreviewUrl(null)
    setSelectedImageFile(null)
    setImageCaption('')
  }

  const sendImageMessage = async () => {
    if (!selectedImageFile || !activeRoom) return
    setIsUploadingImage(true)
    try {
      const formData = new FormData()
      formData.append('file', selectedImageFile)
      const uploadRes = await apiClient.post('/uploads/chat', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      const fileUrl = uploadRes.data?.file_url
      if (!fileUrl) throw new Error('Gagal mendapatkan URL gambar')

      const caption = imageCaption.trim()
      const tempId = `admin_img_${Date.now()}`
      const localMsg: ChatMessage = {
        id: null,
        room_id: activeRoom.id,
        sender_type: 'admin',
        sender_name: user?.name || 'Admin CS',
        message: caption,
        message_type: 'image',
        attachment_url: fileUrl,
        status: 'pending',
        created_at: new Date().toISOString(),
        temp_id: tempId,
      }
      setActiveMessages((prev) => [...prev, localMsg])
      scrollToBottom()
      sendWsEvent('send_message', {
        room_id: activeRoom.id,
        message: caption,
        message_type: 'image',
        attachment_url: fileUrl,
        temp_id: tempId,
      })
      cancelImageUpload()
    } catch (err) {
      console.error('Upload image failed:', err)
      showToast('Gagal mengunggah gambar', 'error')
    } finally {
      setIsUploadingImage(false)
    }
  }

  // Room assignment & status
  const assignRoomToMe = async () => {
    if (!activeRoom || !user?.id) return
    setIsAssigning(true)
    try {
      await apiClient.post(`/chat/rooms/${activeRoom.id}/assign`, { admin_id: user.id })
      const adminObj: ChatAdmin = { id: user.id, nama: user.name || 'Admin CS', email: user.email }
      setActiveRoom((prev) => (prev ? { ...prev, assigned_admin_id: user.id, assigned_admin: adminObj } : prev))
      setRooms((prev) =>
        prev.map((r) => (r.id === activeRoom.id ? { ...r, assigned_admin_id: user.id, assigned_admin: adminObj } : r)),
      )
      showToast('Percakapan berhasil diambil alih')
      void fetchRoomCounts()
    } catch {
      showToast('Gagal mengambil alih percakapan', 'error')
    } finally {
      setIsAssigning(false)
    }
  }

  const unassignActiveRoom = async () => {
    if (!activeRoom) return
    setIsAssigning(true)
    try {
      await apiClient.post(`/chat/rooms/${activeRoom.id}/unassign`)
      setActiveRoom((prev) => (prev ? { ...prev, assigned_admin_id: null, assigned_admin: null } : prev))
      setRooms((prev) =>
        prev.map((r) => (r.id === activeRoom.id ? { ...r, assigned_admin_id: null, assigned_admin: null } : r)),
      )
      showToast('Penugasan percakapan dilepas', 'info')
      void fetchRoomCounts()
    } catch {
      showToast('Gagal melepas penugasan', 'error')
    } finally {
      setIsAssigning(false)
    }
  }

  const openCloseRoomDialog = () => {
    if (!activeRoom) return
    const brandName = getBrandShortLabel(activeRoom.brand)
    setClosingMessageText(
      `Terima kasih telah menghubungi layanan Customer Care ${brandName === 'Semua' ? 'Artacom' : brandName}. Sesi percakapan ini kami tandai selesai. Silakan hubungi kami kembali jika ada kendala lain.`,
    )
    setSendClosingMessage(true)
    setCloseRoomOpen(true)
  }

  const confirmCloseRoom = async () => {
    if (!activeRoom) return
    setIsUpdatingStatus(true)
    try {
      if (sendClosingMessage && closingMessageText.trim()) {
        sendAdminTextMessage(closingMessageText.trim())
      }
      await apiClient.post(`/chat/rooms/${activeRoom.id}/close`)
      setActiveRoom((prev) => (prev ? { ...prev, status: 'closed' } : prev))
      setRooms((prev) => prev.map((r) => (r.id === activeRoom.id ? { ...r, status: 'closed' } : r)))
      setCloseRoomOpen(false)
      showToast('Percakapan ditandai selesai')
      void fetchRoomCounts()
    } catch {
      showToast('Gagal menyelesaikan percakapan', 'error')
    } finally {
      setIsUpdatingStatus(false)
    }
  }

  const reopenActiveRoom = async () => {
    if (!activeRoom) return
    setIsUpdatingStatus(true)
    try {
      await apiClient.post(`/chat/rooms/${activeRoom.id}/reopen`)
      setActiveRoom((prev) => (prev ? { ...prev, status: 'open' } : prev))
      setRooms((prev) => prev.map((r) => (r.id === activeRoom.id ? { ...r, status: 'open' } : r)))
      showToast('Percakapan dibuka kembali')
      void fetchRoomCounts()
    } catch {
      showToast('Gagal membuka kembali percakapan', 'error')
    } finally {
      setIsUpdatingStatus(false)
    }
  }

  // Trouble Ticket
  const autoFillTicketFromChat = async () => {
    if (!activeRoom) return
    setIsAutoFilling(true)
    try {
      const res = await apiClient.post(`/chat/rooms/${activeRoom.id}/auto-fill-ticket`)
      if (res.data?.data) {
        const d = res.data.data
        setTicketForm({
          category: d.category || 'cable_issue',
          priority: d.priority || 'medium',
          title: d.title || 'Kendala Layanan Pelanggan',
          description: d.description || '',
        })
        showToast('Formulir diisi otomatis dari riwayat chat')
      }
    } catch {
      showToast('Gagal menganalisis riwayat chat', 'error')
    } finally {
      setIsAutoFilling(false)
    }
  }

  const submitTroubleTicket = async () => {
    if (!activeRoom) return
    if (!ticketForm.title.trim() || !ticketForm.description.trim()) {
      showToast('Judul dan deskripsi keluhan wajib diisi', 'error')
      return
    }
    setIsSubmittingTicket(true)
    try {
      const res = await troubleTicketAPI.createTicket({
        pelanggan_id: Number(activeRoom.pelanggan_id),
        title: ticketForm.title.trim(),
        description: ticketForm.description.trim(),
        category: ticketForm.category,
        priority: ticketForm.priority,
        data_teknis_id: activeRoom.pelanggan?.data_teknis?.id
          ? Number(activeRoom.pelanggan.data_teknis.id)
          : undefined,
      })
      const created = res.data?.data
      setCreatedTicketResult(created || null)
      setTicketForm({ category: 'slow_connection', priority: 'medium', title: '', description: '' })
      void loadCustomerTickets(activeRoom.pelanggan_id)
      showToast(`Trouble Ticket #${created?.ticket_number || ''} berhasil diterbitkan`)
    } catch {
      showToast('Gagal membuat Trouble Ticket', 'error')
    } finally {
      setIsSubmittingTicket(false)
    }
  }

  const sendTicketNoticeToChat = () => {
    if (!activeRoom || !createdTicketResult) return
    const custName = activeRoom.pelanggan?.nama || 'Bapak/Ibu'
    sendAdminTextMessage(
      `Halo ${custName}, keluhan Anda telah kami catat dengan nomor Trouble Ticket #${createdTicketResult.ticket_number}. Tim teknisi kami segera memproses dan melakukan pengecekan. Mohon ditunggu ya. Terima kasih!`,
    )
    setCreatedTicketResult(null)
  }

  // Templates CRUD
  const [editingTemplateId, setEditingTemplateId] = useState<number | null>(null)

  const handleSaveTemplate = async () => {
    if (!newTplShortcut.trim() || !newTplTitle.trim() || !newTplContent.trim()) return
    setSavingTemplate(true)
    try {
      const payload = {
        shortcut: newTplShortcut.trim().replace(/^\//, '').toLowerCase(),
        title: newTplTitle.trim(),
        content: newTplContent.trim(),
      }
      if (editingTemplateId) {
        await apiClient.put(`/chat/templates/${editingTemplateId}`, payload)
        showToast('Template berhasil diperbarui')
      } else {
        await apiClient.post('/chat/templates', payload)
        showToast('Template berhasil ditambahkan')
      }
      setEditingTemplateId(null)
      setNewTplShortcut('')
      setNewTplTitle('')
      setNewTplContent('')
      await fetchTemplates()
    } catch {
      showToast('Gagal menyimpan template', 'error')
    } finally {
      setSavingTemplate(false)
    }
  }

  const startEditTemplate = (tpl: QuickTemplate) => {
    if (!tpl.id) return
    setEditingTemplateId(tpl.id)
    setNewTplShortcut(tpl.shortcut || '')
    setNewTplTitle(tpl.title || tpl.label || '')
    setNewTplContent(tpl.content || tpl.text || '')
  }

  const cancelEditTemplate = () => {
    setEditingTemplateId(null)
    setNewTplShortcut('')
    setNewTplTitle('')
    setNewTplContent('')
  }

  const handleDeleteTemplate = async (id?: number) => {
    if (!id) return
    try {
      await apiClient.delete(`/chat/templates/${id}`)
      await fetchTemplates()
      showToast('Template dihapus', 'info')
    } catch {
      showToast('Gagal menghapus template', 'error')
    }
  }

  const copyPhone = async (phone?: string) => {
    if (!phone) return
    try {
      await navigator.clipboard.writeText(phone)
      showToast('Nomor telepon disalin')
    } catch {
      // ignore
    }
  }

  const openWhatsApp = (phone?: string) => {
    if (!phone) return
    let cleaned = phone.replace(/\D/g, '')
    if (cleaned.startsWith('0')) cleaned = `62${cleaned.slice(1)}`
    window.open(`https://wa.me/${cleaned}`, '_blank', 'noopener,noreferrer')
  }

  return (
    <div className="p-3 md:p-4 h-[calc(100vh-4rem)] flex flex-col">
      {/* Toast Notification */}
      {toast && (
        <div
          className={`fixed bottom-5 right-5 z-50 rounded-lg px-4 py-2.5 text-sm font-medium shadow-lg border ${
            toast.type === 'error'
              ? 'bg-rose-600 text-white border-rose-700'
              : toast.type === 'info'
                ? 'bg-slate-800 text-white border-slate-700'
                : 'bg-emerald-600 text-white border-emerald-700'
          }`}
        >
          {toast.text}
        </div>
      )}

      {/* 3-Pane Omnichannel Workspace */}
      <div className="flex-1 min-h-0 rounded-xl border bg-card shadow-sm flex overflow-hidden">
        {/* ================= PANE 1: CONVERSATION INBOX ================= */}
        <aside
          className={`w-full md:w-80 lg:w-[330px] shrink-0 border-r flex flex-col bg-card ${
            activeRoom ? 'hidden md:flex' : 'flex'
          }`}
        >
          {/* Header */}
          <div className="px-4 py-3 border-b flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <MessageSquare className="h-4 w-4" />
              </div>
              <div>
                <h2 className="text-sm font-bold leading-tight">Live Chat Pelanggan</h2>
                <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                  <span
                    className={`h-2 w-2 rounded-full ${isWsConnected ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'}`}
                  />
                  <span>{isWsConnected ? 'Online' : 'Menghubungkan...'}</span>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              {totalUnreadCount > 0 && (
                <Badge className="bg-rose-600 text-white text-[10px] px-1.5 h-5">{totalUnreadCount}</Badge>
              )}
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => void fetchRooms()} title="Muat ulang">
                <RefreshCw className={`h-4 w-4 ${isLoadingRooms ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </div>

          {/* Search & Brand Filter */}
          <div className="px-3 pt-3 pb-2 flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari chat..."
                className="pl-8 h-8 text-xs"
              />
            </div>
            <Select value={selectedBrand} onValueChange={setSelectedBrand}>
              <SelectTrigger className="w-28 h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">Semua</SelectItem>
                <SelectItem value="JAKINET">Jakinet</SelectItem>
                <SelectItem value="JELANTIK">Jelantik</SelectItem>
                <SelectItem value="JELANTIK NAGRAK">Nagrak</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Status Tabs */}
          <div className="px-3 pb-2">
            <div className="grid grid-cols-3 gap-1 rounded-lg bg-muted p-1">
              {[
                { key: 'open' as const, label: 'Terbuka', count: computedCounts.open },
                { key: 'closed' as const, label: 'Selesai', count: computedCounts.closed },
                { key: 'all' as const, label: 'Semua', count: computedCounts.all },
              ].map((t) => (
                <button
                  key={t.key}
                  type="button"
                  onClick={() => setSelectedStatusTab(t.key)}
                  className={`flex items-center justify-center gap-1 rounded-md py-1 text-[11px] font-semibold transition-colors ${
                    selectedStatusTab === t.key
                      ? 'bg-background text-foreground shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <span>{t.label}</span>
                  <span className="text-[10px] opacity-75">({t.count})</span>
                </button>
              ))}
            </div>
          </div>

          {/* Assignment Filter */}
          <div className="px-3 pb-2.5 border-b">
            <div className="grid grid-cols-3 gap-1">
              {[
                { key: 'all' as const, label: 'Semua', icon: Users },
                { key: 'mine' as const, label: `Saya (${computedCounts.mine})`, icon: UserCheck },
                { key: 'unassigned' as const, label: `Antri (${computedCounts.unassigned})`, icon: Clock },
              ].map((a) => {
                const Icon = a.icon
                return (
                  <button
                    key={a.key}
                    type="button"
                    onClick={() => setSelectedAssignment(a.key)}
                    className={`flex items-center justify-center gap-1 rounded-md border py-1 text-[11px] font-medium transition-colors ${
                      selectedAssignment === a.key
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-transparent bg-muted/40 text-muted-foreground hover:bg-muted'
                    }`}
                  >
                    <Icon className="h-3 w-3" />
                    <span>{a.label}</span>
                  </button>
                )
              })}
            </div>
          </div>

          {/* Room List */}
          <div className="flex-1 overflow-y-auto divide-y">
            {isLoadingRooms ? (
              <div className="py-12 text-center text-xs text-muted-foreground">Memuat percakapan...</div>
            ) : filteredRooms.length === 0 ? (
              <div className="py-12 px-4 text-center space-y-1">
                <MessageSquare className="h-8 w-8 mx-auto text-muted-foreground/50" />
                <p className="text-sm font-medium">Tidak ada percakapan</p>
                <p className="text-xs text-muted-foreground">
                  {selectedStatusTab === 'closed' ? 'Belum ada percakapan yang selesai' : 'Belum ada obrolan aktif'}
                </p>
              </div>
            ) : (
              filteredRooms.map((room) => {
                const active = activeRoom?.id === room.id
                const name = room.pelanggan?.nama || `Pelanggan #${room.pelanggan_id}`
                return (
                  <button
                    key={room.id}
                    type="button"
                    onClick={() => void selectRoom(room)}
                    className={`w-full text-left px-3.5 py-3 flex items-start gap-3 transition-colors hover:bg-muted/50 ${
                      active ? 'bg-primary/5 border-l-2 border-l-primary' : ''
                    }`}
                  >
                    <div className="relative shrink-0 mt-0.5">
                      <div
                        className={`flex h-9 w-9 items-center justify-center rounded-full text-xs font-bold ${getBrandAvatarClass(
                          room.brand,
                        )}`}
                      >
                        {getInitials(name)}
                      </div>
                      {room.status !== 'closed' && (
                        <span className="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border-2 border-background bg-emerald-500" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-baseline justify-between gap-2">
                        <span className="text-xs font-bold truncate">{name}</span>
                        <span className="text-[10px] text-muted-foreground shrink-0 tabular-nums">
                          {formatTimestamp(room.last_message_at)}
                        </span>
                      </div>
                      <div className="flex items-center justify-between gap-2 mt-0.5">
                        <span className="text-xs text-muted-foreground truncate">
                          {room.last_message_text || 'Mulai obrolan...'}
                        </span>
                        {room.unread_count_admin > 0 && (
                          <Badge className="bg-rose-600 text-white text-[10px] h-4 px-1.5 shrink-0">
                            {room.unread_count_admin > 99 ? '99+' : room.unread_count_admin}
                          </Badge>
                        )}
                      </div>
                      <div className="flex items-center gap-1.5 mt-1.5 flex-wrap">
                        <Badge variant="secondary" className="text-[10px] px-1.5 py-0 h-4">
                          {getBrandShortLabel(room.brand)}
                        </Badge>
                        {room.status === 'closed' && (
                          <span className="text-[10px] text-muted-foreground">• Resolved</span>
                        )}
                        {room.assigned_admin_id === user?.id ? (
                          <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 text-primary border-primary/40">
                            Saya
                          </Badge>
                        ) : !room.assigned_admin_id ? (
                          <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 text-amber-600 border-amber-300">
                            Unassigned
                          </Badge>
                        ) : (
                          <span className="text-[10px] text-muted-foreground truncate max-w-[90px]">
                            {room.assigned_admin?.nama || room.assigned_admin?.name || 'Admin'}
                          </span>
                        )}
                      </div>
                    </div>
                  </button>
                )
              })
            )}
          </div>

          {/* Bottom Summary Footer */}
          <div className="px-4 py-2 border-t flex items-center justify-between text-[11px] text-muted-foreground">
            <span>
              Assigned: <strong className="text-foreground">{computedCounts.assigned}</strong>
            </span>
            <span>
              Unassigned: <strong className="text-foreground">{computedCounts.unassigned}</strong>
            </span>
          </div>
        </aside>

        {/* ================= PANE 2: ACTIVE CHAT ROOM ================= */}
        <main className={`flex-1 min-w-0 flex flex-col bg-muted/20 ${!activeRoom ? 'hidden md:flex' : 'flex'}`}>
          {!activeRoom ? (
            <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary mb-4">
                <MessageSquarePlus className="h-8 w-8" />
              </div>
              <h3 className="text-lg font-bold">Pusat Layanan Chat Pelanggan</h3>
              <p className="text-sm text-muted-foreground max-w-sm mt-1">
                Pilih salah satu percakapan di sebelah kiri untuk membaca, membalas pesan, atau membuat Trouble Ticket.
              </p>
            </div>
          ) : (
            <>
              {/* Chat Header */}
              <header className="px-4 py-2.5 border-b bg-card flex items-center justify-between gap-2">
                <div className="flex items-center gap-3 min-w-0">
                  <Button variant="ghost" size="icon" className="md:hidden h-8 w-8" onClick={() => setActiveRoom(null)}>
                    <ArrowLeft className="h-4 w-4" />
                  </Button>
                  <div
                    className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-bold ${getBrandAvatarClass(
                      activeRoom.brand,
                    )}`}
                  >
                    {getInitials(activeRoom.pelanggan?.nama || 'Pelanggan')}
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="text-sm font-bold truncate">
                        {activeRoom.pelanggan?.nama || `Pelanggan #${activeRoom.pelanggan_id}`}
                      </h3>
                      <Badge variant="secondary" className="text-[10px]">
                        {getBrandShortLabel(activeRoom.brand)}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Phone className="h-3 w-3" />
                        {activeRoom.pelanggan?.no_telp || '-'}
                      </span>
                      {activeRoom.pelanggan?.no_telp && (
                        <button
                          type="button"
                          onClick={() => void copyPhone(activeRoom.pelanggan?.no_telp)}
                          className="hover:text-foreground"
                          title="Salin nomor"
                        >
                          <Copy className="h-3 w-3" />
                        </button>
                      )}
                      {isCustomerTyping && (
                        <span className="text-primary font-medium animate-pulse">sedang mengetik...</span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Header Actions */}
                <div className="flex items-center gap-1.5 shrink-0">
                  {activeRoom.status !== 'closed' && (
                    activeRoom.assigned_admin_id === user?.id ? (
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 text-xs"
                        disabled={isAssigning}
                        onClick={() => void unassignActiveRoom()}
                        title="Lepas Penugasan"
                      >
                        <UserMinus className="mr-1 h-3.5 w-3.5" />
                        Lepas
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        className="h-7 text-xs"
                        disabled={isAssigning}
                        onClick={() => void assignRoomToMe()}
                      >
                        <UserCheck className="mr-1 h-3.5 w-3.5" />
                        Ambil Alih
                      </Button>
                    )
                  )}

                  <Button
                    variant="outline"
                    size="sm"
                    className="h-7 text-xs"
                    onClick={() => {
                      setShowInfoPanel(true)
                      setRightPanelTab('ticket')
                    }}
                  >
                    <Ticket className="mr-1 h-3.5 w-3.5" />
                    <span className="hidden sm:inline">Buat Tiket</span>
                  </Button>

                  {activeRoom.status !== 'closed' ? (
                    <Button
                      variant="secondary"
                      size="sm"
                      className="h-7 text-xs"
                      disabled={isUpdatingStatus}
                      onClick={openCloseRoomDialog}
                    >
                      <CheckCircle2 className="mr-1 h-3.5 w-3.5 text-emerald-600" />
                      Resolve
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      disabled={isUpdatingStatus}
                      onClick={() => void reopenActiveRoom()}
                    >
                      <LockOpen className="mr-1 h-3.5 w-3.5" />
                      Buka Kembali
                    </Button>
                  )}

                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => setIsSoundEnabled((v) => !v)}
                    title={isSoundEnabled ? 'Suara Notifikasi Aktif' : 'Suara Notifikasi Mati'}
                  >
                    {isSoundEnabled ? <Volume2 className="h-4 w-4" /> : <VolumeX className="h-4 w-4 text-muted-foreground" />}
                  </Button>

                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => setShowInfoPanel((v) => !v)}
                    title="Panel Profil & Tiket"
                  >
                    {showInfoPanel ? <PanelRightClose className="h-4 w-4" /> : <PanelRightOpen className="h-4 w-4" />}
                  </Button>
                </div>
              </header>

              {/* Messages Stream */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {isLoadingMessages ? (
                  <div className="py-12 text-center text-xs text-muted-foreground">Memuat riwayat obrolan...</div>
                ) : activeMessages.length === 0 ? (
                  <div className="py-12 text-center space-y-1 text-muted-foreground">
                    <MessageSquare className="h-9 w-9 mx-auto opacity-40" />
                    <p className="text-sm font-medium">Belum ada pesan dalam obrolan ini</p>
                    <p className="text-xs">Ketik balasan di bawah untuk memulai percakapan</p>
                  </div>
                ) : (
                  <>
                    {activeRoom.status === 'closed' && (
                      <div className="text-center my-2">
                        <span className="inline-flex items-center gap-1.5 rounded-full border bg-card px-3.5 py-1 text-xs text-muted-foreground">
                          <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                          Percakapan ini selesai. Kirim pesan baru untuk membuka kembali.
                        </span>
                      </div>
                    )}

                    {activeMessages.map((msg, idx) => {
                      const prevDate = idx > 0 ? new Date(activeMessages[idx - 1].created_at).toDateString() : ''
                      const curDate = new Date(msg.created_at).toDateString()
                      const showDate = idx === 0 || prevDate !== curDate
                      const isAdmin = msg.sender_type === 'admin'
                      const isSystem = msg.sender_type === 'system'
                      const isNote = isAdmin && msg.message.startsWith('📝 [Catatan Internal]')

                      return (
                        <div key={msg.id || msg.temp_id || idx} className="space-y-2">
                          {showDate && (
                            <div className="flex items-center justify-center my-2">
                              <span className="rounded-full bg-muted px-3 py-0.5 text-[11px] font-medium text-muted-foreground">
                                {formatDateHeader(msg.created_at)}
                              </span>
                            </div>
                          )}

                          <div className={`flex ${isAdmin ? 'justify-end' : 'justify-start'}`}>
                            <div
                              className={`max-w-[78%] sm:max-w-[68%] rounded-2xl px-3.5 py-2.5 shadow-xs ${
                                isNote
                                  ? 'bg-amber-50 border border-amber-300 text-amber-950 dark:bg-amber-950/40 dark:border-amber-800 dark:text-amber-100'
                                  : isAdmin
                                    ? 'bg-primary text-primary-foreground rounded-br-sm'
                                    : 'bg-card border rounded-bl-sm text-foreground'
                              }`}
                            >
                              {isSystem && (
                                <div className="flex items-center gap-1.5 text-[11px] font-bold text-primary mb-1">
                                  <Bot className="h-3.5 w-3.5" />
                                  <span>Asisten Virtual AI</span>
                                </div>
                              )}

                              {(msg.attachment_url || msg.message_type === 'image') && msg.attachment_url && (
                                <button
                                  type="button"
                                  onClick={() => setLightboxUrl(getFullMediaUrl(msg.attachment_url))}
                                  className="block mb-2 overflow-hidden rounded-lg border bg-muted"
                                >
                                  <img
                                    src={getFullMediaUrl(msg.attachment_url)}
                                    alt="Lampiran chat"
                                    className="max-h-64 max-w-[260px] object-cover"
                                  />
                                </button>
                              )}

                              {msg.message && (
                                <p className="text-sm whitespace-pre-wrap break-words leading-relaxed">{msg.message}</p>
                              )}

                              <div
                                className={`flex items-center justify-end gap-1 mt-1 text-[10px] ${
                                  isNote
                                    ? 'text-amber-700 dark:text-amber-300'
                                    : isAdmin
                                      ? 'text-primary-foreground/75'
                                      : 'text-muted-foreground'
                                }`}
                              >
                                <span className="tabular-nums">{formatTime(msg.created_at)}</span>
                                {isAdmin && (
                                  <>
                                    {msg.status === 'pending' && <Clock className="h-3 w-3" />}
                                    {msg.status === 'sent' && <Check className="h-3 w-3" />}
                                    {msg.status === 'delivered' && <CheckCheck className="h-3.5 w-3.5" />}
                                    {msg.status === 'read' && <CheckCheck className="h-3.5 w-3.5 text-sky-300" />}
                                  </>
                                )}
                              </div>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                    <div ref={messagesEndRef} />
                  </>
                )}
              </div>

              {/* Quick Templates Bar */}
              {activeRoom.status !== 'closed' && (
                <div className="px-3 py-1.5 border-t bg-card flex items-center gap-2 overflow-x-auto">
                  <Zap className="h-3.5 w-3.5 text-amber-500 shrink-0" />
                  <span className="text-[11px] font-semibold text-muted-foreground shrink-0">Templat:</span>
                  <div className="flex items-center gap-1.5 flex-1 overflow-x-auto py-0.5">
                    {quickTemplates.map((tpl, i) => (
                      <button
                        key={tpl.id || tpl.shortcut || i}
                        type="button"
                        onClick={() => {
                          setInputMessage(tpl.content || tpl.text || '')
                          chatInputRef.current?.focus()
                        }}
                        className="shrink-0 rounded-full border bg-primary/5 hover:bg-primary/10 text-primary px-2.5 py-0.5 text-[11px] font-medium transition-colors"
                      >
                        /{tpl.shortcut} · {tpl.title || tpl.label}
                      </button>
                    ))}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-[11px] shrink-0"
                    onClick={() => setManageTemplatesOpen(true)}
                  >
                    <Settings2 className="mr-1 h-3 w-3" />
                    Kelola
                  </Button>
                </div>
              )}

              {/* Bottom Input Bar */}
              <footer className="relative px-4 py-3 border-t bg-card space-y-2">
                {/* Reply / Internal Note Switcher */}
                <div className="flex items-center gap-1.5">
                  <div className="inline-flex rounded-full bg-muted p-0.5">
                    <button
                      type="button"
                      onClick={() => setInputMode('reply')}
                      className={`rounded-full px-3 py-1 text-[11px] font-semibold transition-colors ${
                        inputMode === 'reply' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground'
                      }`}
                    >
                      Balas Pelanggan
                    </button>
                    <button
                      type="button"
                      onClick={() => setInputMode('note')}
                      className={`flex items-center gap-1 rounded-full px-3 py-1 text-[11px] font-semibold transition-colors ${
                        inputMode === 'note' ? 'bg-amber-500 text-white' : 'text-muted-foreground'
                      }`}
                    >
                      <StickyNote className="h-3 w-3" />
                      Catatan Internal
                    </button>
                  </div>
                </div>

                {/* Slash Command Autocomplete Popover */}
                {showSlashMenu && filteredSlashTemplates.length > 0 && (
                  <div className="absolute bottom-full left-4 right-4 mb-2 z-30 rounded-xl border bg-popover shadow-lg overflow-hidden">
                    <div className="flex items-center justify-between px-3 py-1.5 border-b bg-muted/50 text-[11px] text-muted-foreground">
                      <span className="font-semibold flex items-center gap-1 text-foreground">
                        <Zap className="h-3.5 w-3.5 text-primary" />
                        Template Pintasan ({filteredSlashTemplates.length})
                      </span>
                      <span>[Enter] pilih • [↑↓] navigasi • [Esc] tutup</span>
                    </div>
                    <div className="max-h-52 overflow-y-auto py-1">
                      {filteredSlashTemplates.map((tpl, idx) => (
                        <button
                          key={tpl.id || tpl.shortcut || idx}
                          type="button"
                          onMouseEnter={() => setSelectedSlashIndex(idx)}
                          onClick={() => applySlashTemplate(tpl)}
                          className={`w-full text-left px-3 py-2 flex items-start gap-2.5 text-xs transition-colors ${
                            selectedSlashIndex === idx ? 'bg-primary/10' : 'hover:bg-muted/50'
                          }`}
                        >
                          <Badge className="shrink-0 text-[10px]">/{tpl.shortcut}</Badge>
                          <div className="min-w-0">
                            <p className="font-semibold">{tpl.title || tpl.label}</p>
                            <p className="text-muted-foreground truncate">{tpl.content || tpl.text}</p>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Emoji Picker Popover */}
                {showEmojiPicker && (
                  <div className="absolute bottom-full left-4 mb-2 z-30 w-64 rounded-xl border bg-popover p-2.5 shadow-lg">
                    <div className="flex items-center justify-between mb-1.5 px-1">
                      <span className="text-[11px] font-semibold text-muted-foreground">Pilih Emoji</span>
                      <button type="button" onClick={() => setShowEmojiPicker(false)}>
                        <X className="h-3.5 w-3.5 text-muted-foreground" />
                      </button>
                    </div>
                    <div className="grid grid-cols-8 gap-1">
                      {EMOJI_LIST.map((em) => (
                        <button
                          key={em}
                          type="button"
                          onClick={() => {
                            setInputMessage((prev) => prev + em)
                            chatInputRef.current?.focus()
                          }}
                          className="flex h-7 w-7 items-center justify-center rounded hover:bg-muted text-base"
                        >
                          {em}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="flex items-end gap-2">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    onClick={() => setShowEmojiPicker((v) => !v)}
                    title="Pilih Emoji"
                  >
                    <Smile className="h-4 w-4" />
                  </Button>

                  <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={onImageSelected} />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    disabled={isUploadingImage}
                    onClick={() => fileInputRef.current?.click()}
                    title="Kirim Gambar"
                  >
                    <ImageIcon className="h-4 w-4" />
                  </Button>

                  <Textarea
                    ref={chatInputRef}
                    rows={1}
                    value={inputMessage}
                    onChange={(e) => handleChatChange(e.target.value)}
                    onKeyDown={handleChatKeyDown}
                    onFocus={() => setShowEmojiPicker(false)}
                    placeholder={
                      inputMode === 'note'
                        ? 'Tulis catatan internal tim CS...'
                        : 'Ketik balasan CS... (Ketik / untuk template, Enter untuk kirim)'
                    }
                    className="min-h-[38px] max-h-28 resize-none text-sm"
                  />

                  <Button
                    type="button"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    disabled={!inputMessage.trim()}
                    onClick={sendAdminMessage}
                    title="Kirim Pesan"
                  >
                    <Send className="h-4 w-4" />
                  </Button>
                </div>
              </footer>
            </>
          )}
        </main>

        {/* ================= PANE 3: CONTACT 360 & CREATE TICKET ================= */}
        {showInfoPanel && activeRoom && (
          <aside className="hidden lg:flex w-80 xl:w-[350px] shrink-0 border-l flex-col bg-card overflow-y-auto">
            {/* Panel Tab Header */}
            <div className="px-4 border-b flex items-center justify-between shrink-0">
              <div className="flex items-center gap-4">
                <button
                  type="button"
                  onClick={() => setRightPanelTab('profile')}
                  className={`py-3 text-xs font-bold border-b-2 transition-colors ${
                    rightPanelTab === 'profile'
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                >
                  Profil 360
                </button>
                <button
                  type="button"
                  onClick={() => setRightPanelTab('ticket')}
                  className={`py-3 text-xs font-bold border-b-2 flex items-center gap-1.5 transition-colors ${
                    rightPanelTab === 'ticket'
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <span>Trouble Ticket</span>
                  {customerTickets.length > 0 && (
                    <Badge variant="secondary" className="text-[10px] px-1.5 h-4">
                      {customerTickets.length}
                    </Badge>
                  )}
                </button>
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowInfoPanel(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>

            {rightPanelTab === 'profile' ? (
              <div className="p-4 space-y-4">
                <div className="text-center space-y-1.5 pb-2 border-b">
                  <div
                    className={`mx-auto flex h-12 w-12 items-center justify-center rounded-full text-sm font-bold ${getBrandAvatarClass(
                      activeRoom.brand,
                    )}`}
                  >
                    {getInitials(activeRoom.pelanggan?.nama || 'Pelanggan')}
                  </div>
                  <h4 className="text-sm font-bold">{activeRoom.pelanggan?.nama || '-'}</h4>
                  <p className="text-xs text-muted-foreground">
                    ID: <strong>{activeRoom.pelanggan?.customer_id || `#${activeRoom.pelanggan_id}`}</strong>
                  </p>
                  <Badge variant="secondary" className="text-[10px]">
                    {getBrandShortLabel(activeRoom.brand)}
                  </Badge>
                </div>

                {/* Kontak & Lokasi */}
                <div className="space-y-1.5">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Kontak & Lokasi</p>
                  <div className="rounded-lg border p-3 space-y-2.5 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="flex items-center gap-1.5 font-medium">
                        <Phone className="h-3.5 w-3.5 text-primary" />
                        {activeRoom.pelanggan?.no_telp || '-'}
                      </span>
                      {activeRoom.pelanggan?.no_telp && (
                        <button type="button" onClick={() => void copyPhone(activeRoom.pelanggan?.no_telp)}>
                          <Copy className="h-3.5 w-3.5 text-muted-foreground hover:text-foreground" />
                        </button>
                      )}
                    </div>
                    <div className="border-t pt-2 space-y-1">
                      <span className="flex items-center gap-1 text-muted-foreground">
                        <MapPin className="h-3.5 w-3.5" />
                        Alamat Pemasangan:
                      </span>
                      <p className="font-medium leading-snug">
                        {activeRoom.pelanggan?.alamat || '-'}
                        {(activeRoom.pelanggan?.blok || activeRoom.pelanggan?.unit) && (
                          <span className="block text-primary font-semibold mt-0.5">
                            Blok {activeRoom.pelanggan?.blok || '-'} / Unit {activeRoom.pelanggan?.unit || '-'}
                          </span>
                        )}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Layanan FTTH */}
                <div className="space-y-1.5">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Layanan FTTH</p>
                  <div className="rounded-lg border p-3 space-y-2 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground flex items-center gap-1.5">
                        <Wifi className="h-3.5 w-3.5 text-primary" />
                        Paket:
                      </span>
                      <span className="font-bold text-primary">
                        {activeRoom.pelanggan?.harga_layanan?.nama_layanan ||
                          activeRoom.pelanggan?.layanan ||
                          'Internet FTTH'}
                      </span>
                    </div>
                    <div className="flex items-center justify-between border-t pt-2">
                      <span className="text-muted-foreground">Status:</span>
                      <Badge
                        variant="outline"
                        className={
                          (activeRoom.pelanggan?.langganan?.[0]?.status || 'Aktif') === 'Aktif'
                            ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                            : 'bg-rose-50 text-rose-700 border-rose-200'
                        }
                      >
                        {activeRoom.pelanggan?.langganan?.[0]?.status || 'Aktif'}
                      </Badge>
                    </div>
                    {activeRoom.pelanggan?.data_teknis?.id_pelanggan && (
                      <div className="flex items-center justify-between border-t pt-2">
                        <span className="text-muted-foreground">PPPoE:</span>
                        <code className="rounded border bg-muted px-1.5 py-0.5 font-mono text-[11px]">
                          {activeRoom.pelanggan.data_teknis.id_pelanggan}
                        </code>
                      </div>
                    )}
                  </div>
                </div>

                {/* Quick Actions */}
                <div className="space-y-2 pt-1">
                  <Button className="w-full" variant="outline" size="sm" onClick={() => setRightPanelTab('ticket')}>
                    <Ticket className="mr-2 h-4 w-4 text-primary" />
                    Buat Trouble Ticket
                  </Button>
                  {activeRoom.pelanggan?.no_telp && (
                    <Button
                      className="w-full bg-emerald-600 hover:bg-emerald-700 text-white"
                      size="sm"
                      onClick={() => openWhatsApp(activeRoom.pelanggan?.no_telp)}
                    >
                      <ExternalLink className="mr-2 h-4 w-4" />
                      Chat via WhatsApp Web
                    </Button>
                  )}
                  <Button className="w-full" variant="secondary" size="sm" onClick={() => navigate('/pelanggan')}>
                    Buka Data Pelanggan
                  </Button>
                </div>
              </div>
            ) : (
              <div className="p-4 space-y-4">
                {/* Auto-Fill Banner */}
                <div className="rounded-lg border bg-primary/5 p-3 space-y-2">
                  <div className="flex items-center gap-2">
                    <Sparkles className="h-4 w-4 text-primary shrink-0" />
                    <div>
                      <p className="text-xs font-bold">Asisten Tiket Cerdas</p>
                      <p className="text-[11px] text-muted-foreground">Deteksi otomatis keluhan dari percakapan</p>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    className="w-full h-8 text-xs"
                    disabled={isAutoFilling}
                    onClick={() => void autoFillTicketFromChat()}
                  >
                    <Zap className="mr-1.5 h-3.5 w-3.5" />
                    {isAutoFilling ? 'Menganalisis...' : 'Auto-Fill dari Chat'}
                  </Button>
                </div>

                {/* Form Trouble Ticket */}
                <div className="space-y-3">
                  <p className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                    Formulir Trouble Ticket
                  </p>
                  <div className="space-y-1">
                    <Label className="text-xs">Kategori Gangguan *</Label>
                    <Select
                      value={ticketForm.category}
                      onValueChange={(v) => setTicketForm((f) => ({ ...f, category: v }))}
                    >
                      <SelectTrigger className="w-full h-8 text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {TICKET_CATEGORIES.map((c) => (
                          <SelectItem key={c.value} value={c.value}>
                            {c.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Judul Tiket *</Label>
                    <Input
                      value={ticketForm.title}
                      onChange={(e) => setTicketForm((f) => ({ ...f, title: e.target.value }))}
                      placeholder="Contoh: Kabel Fiber Putus / WiFi Lemot"
                      className="h-8 text-xs"
                    />
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Tingkat Prioritas *</Label>
                    <Select
                      value={ticketForm.priority}
                      onValueChange={(v) => setTicketForm((f) => ({ ...f, priority: v }))}
                    >
                      <SelectTrigger className="w-full h-8 text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {TICKET_PRIORITIES.map((p) => (
                          <SelectItem key={p.value} value={p.value}>
                            {p.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Deskripsi Keluhan *</Label>
                    <Textarea
                      rows={3}
                      value={ticketForm.description}
                      onChange={(e) => setTicketForm((f) => ({ ...f, description: e.target.value }))}
                      placeholder="Tuliskan keluhan atau laporan gangguan..."
                      className="text-xs"
                    />
                  </div>

                  <Button
                    className="w-full"
                    size="sm"
                    disabled={isSubmittingTicket}
                    onClick={() => void submitTroubleTicket()}
                  >
                    <Ticket className="mr-1.5 h-4 w-4" />
                    {isSubmittingTicket ? 'Menerbitkan...' : 'Terbitkan Trouble Ticket'}
                  </Button>
                </div>

                {/* Riwayat Tiket */}
                <div className="border-t pt-3 space-y-2">
                  <div className="flex items-center justify-between text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                    <span>Riwayat Tiket Pelanggan</span>
                    <span>({customerTickets.length})</span>
                  </div>
                  {isLoadingTickets ? (
                    <p className="text-xs text-muted-foreground text-center py-3">Memuat riwayat...</p>
                  ) : customerTickets.length === 0 ? (
                    <p className="text-xs text-muted-foreground text-center py-3 rounded-lg border">
                      Belum ada riwayat tiket gangguan.
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {customerTickets.map((t) => (
                        <div key={t.id} className="rounded-lg border p-2.5 text-xs space-y-1">
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-bold text-primary">{t.ticket_number}</span>
                            <Badge variant="outline" className={`text-[10px] ${ticketStatusClass(t.status)}`}>
                              {t.status}
                            </Badge>
                          </div>
                          <p className="font-medium truncate">{t.title}</p>
                          <div className="flex items-center justify-between text-[10px] text-muted-foreground">
                            <span>{t.category}</span>
                            <span>{formatTimestamp(t.created_at)}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            )}
          </aside>
        )}
      </div>

      {/* Dialog: Image Upload Preview */}
      <Dialog open={imageUploadOpen} onOpenChange={(o) => !o && cancelImageUpload()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Kirim Gambar</DialogTitle>
            <DialogDescription>Pratinjau gambar sebelum dikirim ke pelanggan.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {selectedImagePreviewUrl && (
              <div className="flex items-center justify-center rounded-lg border bg-muted p-2 max-h-64 overflow-hidden">
                <img src={selectedImagePreviewUrl} alt="Preview" className="max-h-60 object-contain rounded" />
              </div>
            )}
            <Input
              value={imageCaption}
              onChange={(e) => setImageCaption(e.target.value)}
              placeholder="Tambah keterangan gambar... (opsional)"
              onKeyDown={(e) => e.key === 'Enter' && void sendImageMessage()}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={cancelImageUpload} disabled={isUploadingImage}>
              Batal
            </Button>
            <Button onClick={() => void sendImageMessage()} disabled={isUploadingImage}>
              {isUploadingImage ? 'Mengunggah...' : 'Kirim'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Fullscreen Lightbox */}
      <Dialog open={Boolean(lightboxUrl)} onOpenChange={(o) => !o && setLightboxUrl(null)}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>Pratinjau Gambar</DialogTitle>
          </DialogHeader>
          {lightboxUrl && (
            <div className="flex items-center justify-center bg-black/90 rounded-lg p-2 max-h-[75vh]">
              <img src={lightboxUrl} alt="Lampiran" className="max-h-[70vh] object-contain" />
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Dialog: Resolve Conversation */}
      <Dialog open={closeRoomOpen} onOpenChange={setCloseRoomOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Selesaikan Percakapan</DialogTitle>
            <DialogDescription>
              Tandai sesi percakapan dengan <strong>{activeRoom?.pelanggan?.nama || 'Pelanggan'}</strong> sebagai selesai.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={sendClosingMessage}
                onChange={(e) => setSendClosingMessage(e.target.checked)}
                className="rounded"
              />
              <span>Kirim pesan penutup otomatis ke pelanggan</span>
            </label>
            {sendClosingMessage && (
              <Textarea
                rows={3}
                value={closingMessageText}
                onChange={(e) => setClosingMessageText(e.target.value)}
                placeholder="Tulis pesan penutup..."
              />
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCloseRoomOpen(false)} disabled={isUpdatingStatus}>
              Batal
            </Button>
            <Button onClick={() => void confirmCloseRoom()} disabled={isUpdatingStatus}>
              {isUpdatingStatus ? 'Menyimpan...' : 'Selesaikan & Tutup'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Ticket Created Success */}
      <Dialog open={Boolean(createdTicketResult)} onOpenChange={(o) => !o && setCreatedTicketResult(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Trouble Ticket Diterbitkan</DialogTitle>
            <DialogDescription>
              Tiket <strong>#{createdTicketResult?.ticket_number}</strong> berhasil dibuat untuk pelanggan ini.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreatedTicketResult(null)}>
              Tutup
            </Button>
            <Button onClick={sendTicketNoticeToChat}>
              <Send className="mr-2 h-4 w-4" />
              Kirim Info Tiket ke Chat
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Manage Quick Templates */}
      <Dialog open={manageTemplatesOpen} onOpenChange={setManageTemplatesOpen}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>Kelola Template Pintasan (/)</DialogTitle>
            <DialogDescription>Tambah atau hapus template balasan cepat untuk mengetik /shortcut di chat.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <Input
                value={newTplShortcut}
                onChange={(e) => setNewTplShortcut(e.target.value)}
                placeholder="Shortcut (contoh: salam)"
              />
              <Input
                value={newTplTitle}
                onChange={(e) => setNewTplTitle(e.target.value)}
                placeholder="Judul (contoh: Salam Pembuka)"
              />
            </div>
            <div className="flex gap-2">
              <Textarea
                rows={2}
                value={newTplContent}
                onChange={(e) => setNewTplContent(e.target.value)}
                placeholder="Isi pesan template..."
                className="flex-1"
              />
              <div className="flex flex-col gap-1.5">
                <Button onClick={() => void handleSaveTemplate()} disabled={savingTemplate}>
                  <Plus className="mr-1 h-4 w-4" />
                  {editingTemplateId ? 'Update' : 'Tambah'}
                </Button>
                {editingTemplateId && (
                  <Button variant="outline" size="sm" onClick={cancelEditTemplate}>
                    Batal
                  </Button>
                )}
              </div>
            </div>
            <div className="max-h-56 overflow-y-auto divide-y rounded-lg border">
              {quickTemplates.map((tpl, idx) => (
                <div key={tpl.id || tpl.shortcut || idx} className="flex items-start justify-between gap-2 p-2.5 text-xs">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary">/{tpl.shortcut}</Badge>
                      <span className="font-bold">{tpl.title || tpl.label}</span>
                    </div>
                    <p className="text-muted-foreground mt-1">{tpl.content || tpl.text}</p>
                  </div>
                  {tpl.id && (
                    <div className="flex items-center gap-1 shrink-0">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-sky-600"
                        onClick={() => startEditTemplate(tpl)}
                        title="Edit template"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-rose-600"
                        onClick={() => void handleDeleteTemplate(tpl.id)}
                        title="Hapus template"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
