import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import apiClient, { troubleTicketAPI } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  TicketCheck,
  Plus,
  RefreshCw,
  Search,
  Eye,
  Settings2,
  Trash2,
  Clock,
  AlertCircle,
  CheckCircle2,
  PauseCircle,
  XCircle,
  Archive,
  FilterX,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Timer,
  Users,
  Upload,
  Paperclip,
  X,
  FileText,
  UserCheck,
  Server,
} from 'lucide-react'

interface TroubleTicket {
  id: number
  ticket_number: string
  title: string
  description: string
  status: string
  priority: string
  category: string
  pelanggan?: {
    id: number
    nama: string
    alamat: string
    no_telp?: string
    harga_layanan?: { brand: string }
  }
  data_teknis?: { id_pelanggan: string; ip_pelanggan?: string }
  assigned_user?: { id: number; name: string }
  downtime_start?: string | null
  downtime_end?: string | null
  total_downtime_minutes?: number
  pending_start?: string | null
  total_pending_minutes: number
  resolved_at?: string | null
  created_at: string
  updated_at: string
}

interface Statistics {
  total_tickets: number
  open_tickets: number
  in_progress_tickets: number
  resolved_tickets: number
  closed_tickets: number
}

interface Customer {
  id: number
  nama: string
  no_telp?: string
  alamat?: string
  harga_layanan?: { brand: string }
}

interface DataTeknis {
  id: number
  display_name: string
  pelanggan_id?: number
  ip_pelanggan?: string
  id_pelanggan?: string
  olt?: string
  olt_custom?: string
  pon?: string
  port_odp?: string
  sn?: string
}

interface TechnicianUser {
  id: number
  name: string
  nama?: string
  username?: string
  email?: string
  role?: { name: string }
}

const statusOptions = [
  { label: 'Open', value: 'open' },
  { label: 'Pending', value: 'pending_customer' },
  { label: 'Closed', value: 'closed' },
]

const categoryOptions = [
  { label: 'No Connection', value: 'no_connection' },
  { label: 'Slow Connection', value: 'slow_connection' },
  { label: 'Intermittent', value: 'intermittent' },
  { label: 'Hardware Issue', value: 'hardware_issue' },
  { label: 'Cable Issue', value: 'cable_issue' },
  { label: 'ONU Issue', value: 'onu_issue' },
  { label: 'OLT Issue', value: 'olt_issue' },
  { label: 'Mikrotik Issue', value: 'mikrotik_issue' },
  { label: 'Other', value: 'other' },
]

const brandOptions = [
  { label: 'JAKINET', value: 'JAKINET' },
  { label: 'JELANTIK', value: 'JELANTIK' },
  { label: 'JELANTIK NAGRAK', value: 'JELANTIK NAGRAK' },
]

const titleOptions = [
  'Lemot',
  'LOS',
  'Putus-Putus',
  'Isolir',
  'No Internet',
  'Fisik/Perangkat',
  'Lain-lain',
]

const categoryFormOptions = [
  { label: 'No Connection', value: 'no_connection' },
  { label: 'Slow Connection', value: 'slow_connection' },
  { label: 'Intermittent Connection', value: 'intermittent' },
  { label: 'Hardware Issue', value: 'hardware_issue' },
  { label: 'Cable Issue', value: 'cable_issue' },
  { label: 'ONU Issue', value: 'onu_issue' },
  { label: 'OLT Issue', value: 'olt_issue' },
  { label: 'Mikrotik Issue', value: 'mikrotik_issue' },
  { label: 'Other', value: 'other' },
]

function formatFileSize(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${Math.round((bytes / Math.pow(k, i)) * 100) / 100} ${sizes[i]}`
}

function getStatusColor(status: string) {
  const map: Record<string, string> = {
    open: 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800',
    in_progress: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800',
    pending_customer: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-950/50 dark:text-orange-300 dark:border-orange-800',
    pending_vendor: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-950/50 dark:text-orange-300 dark:border-orange-800',
    resolved: 'bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800',
    closed: 'bg-muted text-muted-foreground border-border',
    cancelled: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-950/50 dark:text-red-300 dark:border-red-800',
  }
  return map[status] || 'bg-muted text-muted-foreground border-border'
}

function getStatusIcon(status: string) {
  const map: Record<string, React.ReactNode> = {
    open: <Clock className="h-3 w-3" />,
    in_progress: <Timer className="h-3 w-3" />,
    pending_customer: <PauseCircle className="h-3 w-3" />,
    pending_vendor: <PauseCircle className="h-3 w-3" />,
    resolved: <CheckCircle2 className="h-3 w-3" />,
    closed: <Archive className="h-3 w-3" />,
    cancelled: <XCircle className="h-3 w-3" />,
  }
  return map[status] || <AlertCircle className="h-3 w-3" />
}

function formatStatus(status: string) {
  if (status === 'pending_customer' || status === 'pending_vendor') return 'Pending'
  return status.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function formatCategory(cat: string) {
  return cat.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('id-ID')
}

function formatTime(d: string) {
  return new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

function isTicketPending(t: TroubleTicket) {
  const s = (t.status || '').toLowerCase().trim()
  return s === 'pending' || s === 'pending_customer' || s === 'pending_vendor'
}

function isTicketResolvedOrClosed(t: TroubleTicket) {
  const s = (t.status || '').toLowerCase().trim()
  return s === 'resolved' || s === 'closed' || s === 'cancelled' || !!t.downtime_end
}

function getCompletedTicketMinutes(t: TroubleTicket) {
  if (typeof t.total_downtime_minutes === 'number' && t.total_downtime_minutes > 0) {
    return t.total_downtime_minutes
  }
  if (t.downtime_start && t.downtime_end) {
    const start = new Date(t.downtime_start)
    const end = new Date(t.downtime_end)
    const totalPendingMs = (t.total_pending_minutes || 0) * 60000
    return Math.floor(Math.max(0, end.getTime() - start.getTime() - totalPendingMs) / 60000)
  }
  return 0
}

function formatDowntimeMinutes(minutes: number) {
  if (minutes < 60) return `${minutes}m`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  if (m === 0) return `${h}h`
  return `${h}h ${m}m`
}

function getDowntimeColorClass(minutes: number) {
  if (minutes < 60) return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-300'
  if (minutes < 240) return 'bg-amber-100 text-amber-800 dark:bg-amber-950/50 dark:text-amber-300'
  if (minutes < 1440) return 'bg-red-100 text-red-800 dark:bg-red-950/50 dark:text-red-300'
  return 'bg-red-200 text-red-900 dark:bg-red-900/60 dark:text-red-200'
}

function calculateLiveDowntime(t: TroubleTicket, now: Date) {
  if (!t.created_at) return '00:00:00'
  const start = new Date(t.downtime_start || t.created_at)
  const totalPendingMs = (t.total_pending_minutes || 0) * 60000

  if (isTicketPending(t)) {
    const pStart = t.pending_start ? new Date(t.pending_start) : now
    const frozenMs = Math.max(0, pStart.getTime() - start.getTime() - totalPendingMs)
    return msToHHMMSS(frozenMs)
  }

  const diff = Math.max(0, now.getTime() - start.getTime() - totalPendingMs)
  return msToHHMMSS(diff)
}

function getLiveDowntimeColor(t: TroubleTicket, now: Date) {
  if (isTicketPending(t)) return 'bg-orange-100 text-orange-800 dark:bg-orange-950/50 dark:text-orange-300'
  if (!t.created_at) return 'bg-muted text-muted-foreground'
  const start = new Date(t.downtime_start || t.created_at)
  const totalPendingMs = (t.total_pending_minutes || 0) * 60000
  const minutes = Math.floor(Math.max(0, now.getTime() - start.getTime() - totalPendingMs) / 60000)
  if (minutes < 60) return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-300'
  if (minutes < 240) return 'bg-amber-100 text-amber-800 dark:bg-amber-950/50 dark:text-amber-300'
  if (minutes < 1440) return 'bg-red-100 text-red-800 dark:bg-red-950/50 dark:text-red-300'
  return 'bg-red-200 text-red-900 dark:bg-red-900/60 dark:text-red-200'
}

function msToHHMMSS(ms: number) {
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

export function TroubleTicketView() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreateTicket = hasPermission('create_trouble_tickets')
  const canEditTicket = hasPermission('edit_trouble_tickets')
  const canDeleteTicket = hasPermission('delete_trouble_tickets')
  const [loading, setLoading] = useState(false)
  const [tickets, setTickets] = useState<TroubleTicket[]>([])
  const [statistics, setStatistics] = useState<Statistics | null>(null)
  const [totalItems, setTotalItems] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [itemsPerPage, setItemsPerPage] = useState(15)
  const [now, setNow] = useState(new Date())

  const [filterStatus, setFilterStatus] = useState('')
  const [filterCategory, setFilterCategory] = useState('')
  const [filterBrand, setFilterBrand] = useState('')
  const [filterSearch, setFilterSearch] = useState('')
  const [showClosed, setShowClosed] = useState(false)

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [ticketToDelete, setTicketToDelete] = useState<TroubleTicket | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [creating, setCreating] = useState(false)

  // Create form
  const [customers, setCustomers] = useState<Customer[]>([])
  const [technicalData, setTechnicalData] = useState<DataTeknis[]>([])
  const [teknisiUsers, setTeknisiUsers] = useState<TechnicianUser[]>([])
  const [loadingCustomers, setLoadingCustomers] = useState(false)
  const [loadingTechData, setLoadingTechData] = useState(false)
  const [loadingTeknisi, setLoadingTeknisi] = useState(false)
  const [customerSearch, setCustomerSearch] = useState('')
  const [filteredCustomers, setFilteredCustomers] = useState<Customer[]>([])
  const [evidenceFiles, setEvidenceFiles] = useState<File[]>([])
  const [formData, setFormData] = useState({
    pelanggan_id: 0,
    data_teknis_id: 0,
    title: '',
    description: '',
    category: '',
    priority: 'medium',
    assigned_to: null as number | null,
  })

  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchTickets = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = {
        page: currentPage,
        pageSize: itemsPerPage,
        skip: (currentPage - 1) * itemsPerPage,
        limit: itemsPerPage,
        show_closed: showClosed ? 'true' : 'false',
        include_relations: true,
      }
      if (filterStatus) params.status = filterStatus
      if (filterCategory) params.category = filterCategory
      if (filterBrand) {
        params.id_brand = filterBrand
        params.brand = filterBrand
      }
      if (filterSearch) params.search = filterSearch

      const response = await troubleTicketAPI.getTickets(params)
      let all = response.data.data || []
      if (!showClosed && !filterStatus) {
        all = all.filter((t: TroubleTicket) => {
          const s = (t.status || '').toLowerCase().trim()
          return s !== 'closed' && s !== 'resolved' && s !== 'cancelled'
        })
      }
      setTickets(all)
      setTotalItems(response.data.total !== undefined ? response.data.total : all.length)
    } catch (err) {
      console.error('Failed to fetch tickets:', err)
    } finally {
      setLoading(false)
    }
  }, [currentPage, itemsPerPage, filterStatus, filterCategory, filterBrand, filterSearch, showClosed])

  const fetchStatistics = useCallback(async () => {
    try {
      const r = await troubleTicketAPI.getStatistics()
      setStatistics(r.data)
    } catch (err) {
      console.error('Failed to fetch statistics:', err)
    }
  }, [])

  const refreshData = useCallback(() => {
    fetchTickets()
    fetchStatistics()
  }, [fetchTickets, fetchStatistics])

  useEffect(() => {
    refreshData()
  }, [refreshData])

  useEffect(() => {
    timerRef.current = setInterval(() => setNow(new Date()), 1000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [])

  const handleSearchChange = (val: string) => {
    setFilterSearch(val)
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      setCurrentPage(1)
      fetchTickets()
    }, 500)
  }

  const applyFilter = () => {
    setCurrentPage(1)
  }

  useEffect(() => {
    fetchTickets()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage, itemsPerPage, filterStatus, filterCategory, filterBrand, showClosed])

  const clearFilters = () => {
    setFilterStatus('')
    setFilterCategory('')
    setFilterBrand('')
    setFilterSearch('')
    setShowClosed(false)
    setCurrentPage(1)
  }

  const hasActiveFilters = filterStatus || filterCategory || filterBrand || filterSearch || showClosed
  const activeFilterCount = [filterStatus, filterCategory, filterBrand, filterSearch, showClosed].filter(Boolean).length
  const totalPages = Math.ceil(totalItems / itemsPerPage)

  // Create ticket
  const fetchCustomers = async () => {
    setLoadingCustomers(true)
    try {
      let r
      try {
        r = await apiClient.get('/trouble-tickets/customers')
      } catch {
        r = await apiClient.get('/pelanggan?limit=10000&include_relations=true')
      }
      const raw: Customer[] = r.data.data || r.data || []
      const unique = raw.reduce((acc: Customer[], cur) => {
        if (!acc.some((c) => c.id === cur.id)) acc.push(cur)
        return acc
      }, [])
      unique.sort((a, b) => a.nama.localeCompare(b.nama))
      setCustomers(unique)
      setFilteredCustomers(unique)
    } catch {
      setCustomers([])
      setFilteredCustomers([])
    } finally {
      setLoadingCustomers(false)
    }
  }

  const fetchTechnicalData = async (pelangganId: number) => {
    setLoadingTechData(true)
    try {
      const r = await apiClient.get(`/data_teknis/by-pelanggan/${pelangganId}`)
      const raw = r.data?.data || r.data || []
      const list = Array.isArray(raw) ? raw : []
      const unique = list.reduce((acc: DataTeknis[], cur: DataTeknis) => {
        if (!acc.some((d) => d.id === cur.id)) {
          acc.push({
            ...cur,
            display_name: `${cur.ip_pelanggan || 'No IP'} - ${cur.olt || cur.olt_custom || 'No OLT'}`,
          })
        }
        return acc
      }, [])
      setTechnicalData(unique)
    } catch {
      setTechnicalData([])
    } finally {
      setLoadingTechData(false)
    }
  }

  const fetchTeknisiUsers = async () => {
    setLoadingTeknisi(true)
    try {
      let r
      try {
        r = await apiClient.get('/trouble-tickets/technicians')
      } catch {
        r = await apiClient.get('/users?limit=1000')
      }
      const raw: TechnicianUser[] = r.data.data || r.data || []
      const seenIds = new Set<number>()
      const seenNames = new Set<string>()
      const filtered = raw.filter((u) => {
        if (seenIds.has(u.id)) return false
        seenIds.add(u.id)
        if (u.role?.name && u.role.name !== 'Teknisi') return false
        const n = (u.name || u.nama || '').trim().toLowerCase()
        if (n && seenNames.has(n)) return false
        if (n) seenNames.add(n)
        return true
      })
      filtered.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
      setTeknisiUsers(filtered)
    } catch {
      setTeknisiUsers([])
    } finally {
      setLoadingTeknisi(false)
    }
  }

  const openCreateDialog = () => {
    setFormData({
      pelanggan_id: 0,
      data_teknis_id: 0,
      title: '',
      description: '',
      category: '',
      priority: 'medium',
      assigned_to: null,
    })
    setTechnicalData([])
    setEvidenceFiles([])
    setCustomerSearch('')
    setShowCreateDialog(true)
    fetchCustomers()
    fetchTeknisiUsers()
  }

  useEffect(() => {
    if (!customerSearch) {
      setFilteredCustomers(customers)
      return
    }
    const q = customerSearch.toLowerCase()
    setFilteredCustomers(
      customers.filter(
        (c) =>
          c.nama.toLowerCase().includes(q) ||
          (c.no_telp || '').toLowerCase().includes(q) ||
          (c.alamat || '').toLowerCase().includes(q)
      )
    )
  }, [customerSearch, customers])

  const handleEvidenceFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    setEvidenceFiles((prev) => [...prev, ...Array.from(e.target.files!)])
    e.target.value = ''
  }

  const removeEvidenceFile = (idx: number) => {
    setEvidenceFiles((prev) => prev.filter((_, i) => i !== idx))
  }

  const handleCreateTicket = async () => {
    if (!formData.pelanggan_id || !formData.title || !formData.description || !formData.category) return
    setCreating(true)
    try {
      const evidenceUrls: string[] = []
      for (const file of evidenceFiles) {
        const fd = new FormData()
        fd.append('file', file)
        const uploadRes = await apiClient.post('/uploads/evidence', fd, {
          headers: { 'Content-Type': 'multipart/form-data' },
        })
        if (uploadRes.data?.file_url) {
          evidenceUrls.push(uploadRes.data.file_url)
        }
      }

      await troubleTicketAPI.createTicket({
        pelanggan_id: formData.pelanggan_id,
        data_teknis_id: formData.data_teknis_id || null,
        title: formData.title,
        description: formData.description,
        category: formData.category,
        priority: formData.priority,
        assigned_to: formData.assigned_to || null,
        evidence: evidenceUrls.length > 0 ? JSON.stringify(evidenceUrls) : null,
      })
      setShowCreateDialog(false)
      refreshData()
    } catch (err) {
      console.error('Failed to create ticket:', err)
      alert('Gagal membuat ticket. Silakan coba lagi.')
    } finally {
      setCreating(false)
    }
  }

  const confirmDelete = (t: TroubleTicket) => {
    setTicketToDelete(t)
    setShowDeleteDialog(true)
  }

  const handleDelete = async () => {
    if (!ticketToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/trouble-tickets/${ticketToDelete.id}`)
      setShowDeleteDialog(false)
      setTicketToDelete(null)
      refreshData()
    } catch (err) {
      console.error('Failed to delete ticket:', err)
      alert('Gagal menghapus ticket.')
    } finally {
      setDeleting(false)
    }
  }

  const canDelete = (t: TroubleTicket) => ['resolved', 'closed', 'cancelled'].includes(t.status)

  const statCards = [
    { label: 'Total Tickets', value: statistics?.total_tickets || 0, color: 'from-slate-500 to-slate-700', bg: 'bg-muted/50 border-border', icon: <TicketCheck className="h-4 w-4" /> },
    { label: 'Open', value: statistics?.open_tickets || 0, color: 'from-amber-500 to-amber-600', bg: 'bg-amber-50/70 dark:bg-amber-950/30 border-amber-200/80 dark:border-amber-800/50', icon: <Clock className="h-4 w-4" /> },
    { label: 'In Progress', value: statistics?.in_progress_tickets || 0, color: 'from-blue-500 to-blue-600', bg: 'bg-blue-50/70 dark:bg-blue-950/30 border-blue-200/80 dark:border-blue-800/50', icon: <Timer className="h-4 w-4" /> },
    { label: 'Resolved', value: statistics?.resolved_tickets || 0, color: 'from-emerald-500 to-emerald-600', bg: 'bg-emerald-50/70 dark:bg-emerald-950/30 border-emerald-200/80 dark:border-emerald-800/50', icon: <CheckCircle2 className="h-4 w-4" /> },
    { label: 'Closed', value: statistics?.closed_tickets || 0, color: 'from-gray-400 to-gray-500', bg: 'bg-muted/50 border-border', icon: <Archive className="h-4 w-4" /> },
  ]

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-xl font-bold text-foreground flex items-center gap-2">
            <TicketCheck className="h-6 w-6 text-blue-600" />
            Trouble Ticket System
          </h1>
          <p className="text-xs text-muted-foreground mt-0.5">Kelola laporan masalah dan tracking downtime pelanggan</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={refreshData} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreateTicket && (
            <Button size="sm" onClick={openCreateDialog}>
              <Plus className="h-4 w-4 mr-1" />
              Buat Ticket Baru
            </Button>
          )}
        </div>
      </div>

      {/* Compact Stats Strip */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2.5">
        {statCards.map((s) => (
          <div
            key={s.label}
            className={`flex items-center gap-3 rounded-xl border px-3.5 py-2.5 shadow-xs ${s.bg}`}
          >
            <div className={`p-2 rounded-lg bg-gradient-to-br ${s.color} text-white shrink-0`}>
              {s.icon}
            </div>
            <div className="min-w-0">
              <div className="text-lg font-bold leading-tight text-foreground">{s.value}</div>
              <div className="text-[11px] text-muted-foreground font-medium truncate">{s.label}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Compact Filters Bar */}
      <Card className="shadow-xs py-0">
        <CardContent className="p-3.5">
          <div className="flex flex-col lg:flex-row lg:items-center gap-2.5">
            <div className="relative flex-1 min-w-[220px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={filterSearch}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder="Cari ticket #, judul, nama pelanggan..."
                className="pl-9 h-9"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5 lg:w-[540px] shrink-0">
              <Select value={filterStatus} onValueChange={(v) => { setFilterStatus(v === '__all__' ? '' : v); applyFilter() }}>
                <SelectTrigger className="h-9 w-full"><SelectValue placeholder="Semua Status" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Status</SelectItem>
                  {statusOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
              <Select value={filterCategory} onValueChange={(v) => { setFilterCategory(v === '__all__' ? '' : v); applyFilter() }}>
                <SelectTrigger className="h-9 w-full"><SelectValue placeholder="Semua Kategori" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Kategori</SelectItem>
                  {categoryOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
              <Select value={filterBrand} onValueChange={(v) => { setFilterBrand(v === '__all__' ? '' : v); applyFilter() }}>
                <SelectTrigger className="h-9 w-full"><SelectValue placeholder="Semua Brand" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Brand</SelectItem>
                  {brandOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center justify-between lg:justify-end gap-3 shrink-0">
              <label className="flex items-center gap-2 text-xs font-medium cursor-pointer whitespace-nowrap px-2.5 py-1.5 rounded-lg border bg-muted/50 hover:bg-muted transition-colors">
                <input
                  type="checkbox"
                  checked={showClosed}
                  onChange={(e) => setShowClosed(e.target.checked)}
                  className="rounded border-border text-blue-600 focus:ring-blue-500"
                />
                <span className="text-foreground/90">Show Closed</span>
              </label>
              {hasActiveFilters && (
                <Button variant="ghost" size="sm" onClick={clearFilters} className="text-red-500 hover:text-red-700 h-8 px-2 text-xs">
                  <FilterX className="h-3.5 w-3.5 mr-1" />
                  Clear ({activeFilterCount})
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Ticket Table */}
      <Card className="shadow-sm">
        <CardHeader className="pb-3 pt-4 px-4">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-semibold">Ticket List</CardTitle>
            <Badge variant={totalItems > 0 ? 'default' : 'secondary'}>{totalItems} Total</Badge>
          </div>
        </CardHeader>
        <Separator />
        {loading && tickets.length === 0 ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-6 w-6 animate-spin text-blue-500" />
            <span className="ml-2 text-sm text-muted-foreground">Loading tickets...</span>
          </div>
        ) : tickets.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
            <TicketCheck className="h-12 w-12 mb-3" />
            <p className="font-medium">Tidak ada ticket ditemukan</p>
            <p className="text-sm mt-1">Coba ubah filter atau buat ticket baru</p>
          </div>
        ) : (
          <>
            {loading && <div className="h-0.5 bg-blue-500 animate-pulse" />}
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[140px]">Ticket #</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Judul</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Pelanggan</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[100px]">Status</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[120px]">Kategori</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[110px]">Assigned</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[110px]">Downtime</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[100px]">Dibuat</TableHead>
                    <TableHead className="text-xs font-semibold uppercase tracking-wider text-muted-foreground w-[110px] text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tickets.map((t) => (
                    <TableRow key={t.id} className="hover:bg-muted/50/50 cursor-pointer" onClick={() => navigate(`/trouble-tickets/${t.id}/action`)}>
                      <TableCell>
                        <span className="inline-flex items-center gap-1 text-xs font-semibold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-950/50 px-2 py-1 rounded-md">
                          <TicketCheck className="h-3 w-3" />
                          {t.ticket_number}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="text-sm font-medium text-foreground">{t.title}</div>
                        {t.description && (
                          <div className="text-xs text-muted-foreground mt-0.5 truncate max-w-[200px]">
                            {t.description.substring(0, 60)}{t.description.length > 60 ? '...' : ''}
                          </div>
                        )}
                      </TableCell>
                      <TableCell>
                        {t.pelanggan ? (
                          <div>
                            <div className="flex items-center gap-2">
                              <div className="h-7 w-7 rounded-full bg-blue-100 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 flex items-center justify-center text-xs font-bold flex-shrink-0">
                                {t.pelanggan.nama.charAt(0).toUpperCase()}
                              </div>
                              <div>
                                <div className="text-sm font-medium text-foreground">{t.pelanggan.nama}</div>
                                <div className="text-xs text-muted-foreground">{t.pelanggan.alamat}</div>
                              </div>
                            </div>
                            {t.pelanggan.harga_layanan?.brand && (
                              <Badge variant="secondary" className="text-[10px] mt-1 bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300">{t.pelanggan.harga_layanan.brand}</Badge>
                            )}
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded-full border ${getStatusColor(t.status)}`}>
                          {getStatusIcon(t.status)}
                          {formatStatus(t.status)}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-xs font-medium">{formatCategory(t.category)}</Badge>
                      </TableCell>
                      <TableCell>
                        {t.assigned_user ? (
                          <div className="flex items-center gap-1.5">
                            <Users className="h-3 w-3 text-muted-foreground" />
                            <span className="text-sm text-foreground/90">{t.assigned_user.name}</span>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">Unassigned</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {!isTicketResolvedOrClosed(t) ? (
                          <span className={`inline-flex items-center gap-1 text-xs font-mono font-bold px-2 py-1 rounded-md ${isTicketPending(t) ? 'bg-orange-100 dark:bg-orange-950/50 text-orange-800 dark:text-orange-300' : getLiveDowntimeColor(t, now)}`}>
                            {isTicketPending(t) && <PauseCircle className="h-3 w-3" />}
                            {calculateLiveDowntime(t, now)}
                          </span>
                        ) : t.total_downtime_minutes && t.total_downtime_minutes > 0 ? (
                          <span className={`inline-flex items-center text-xs font-semibold px-2 py-1 rounded-md ${getDowntimeColorClass(t.total_downtime_minutes)}`}>
                            {formatDowntimeMinutes(t.total_downtime_minutes)}
                          </span>
                        ) : t.downtime_start && t.downtime_end ? (
                          <span className={`inline-flex items-center text-xs font-semibold px-2 py-1 rounded-md ${getDowntimeColorClass(getCompletedTicketMinutes(t))}`}>
                            {formatDowntimeMinutes(getCompletedTicketMinutes(t))}
                          </span>
                        ) : (
                          <span className="text-xs text-emerald-600 font-medium">No downtime</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="text-xs text-foreground/90">{formatDate(t.created_at)}</div>
                        <div className="text-[11px] text-muted-foreground">{formatTime(t.created_at)}</div>
                      </TableCell>
                      <TableCell className="text-right" onClick={(e) => e.stopPropagation()}>
                        <TooltipProvider delayDuration={200}>
                          <div className="flex items-center justify-end gap-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => navigate(`/trouble-tickets/${t.id}`)}>
                                  <Eye className="h-3.5 w-3.5 text-blue-600" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>Lihat Detail</TooltipContent>
                            </Tooltip>
                            {canEditTicket && (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => navigate(`/trouble-tickets/${t.id}/action`)}>
                                    <Settings2 className="h-3.5 w-3.5 text-emerald-600" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>Kelola Aksi</TooltipContent>
                              </Tooltip>
                            )}
                            {canDeleteTicket && canDelete(t) && (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => confirmDelete(t)}>
                                    <Trash2 className="h-3.5 w-3.5 text-red-500" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent>Hapus Ticket</TooltipContent>
                              </Tooltip>
                            )}
                          </div>
                        </TooltipProvider>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between px-4 py-3 border-t">
              <div className="text-xs text-muted-foreground">
                Menampilkan {(currentPage - 1) * itemsPerPage + 1}-{Math.min(currentPage * itemsPerPage, totalItems)} dari {totalItems}
              </div>
              <div className="flex items-center gap-2">
                <Select value={String(itemsPerPage)} onValueChange={(v) => { setItemsPerPage(Number(v)); setCurrentPage(1) }}>
                  <SelectTrigger className="h-8 w-[70px] text-xs"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {[10, 15, 25, 50].map((n) => <SelectItem key={n} value={String(n)}>{n}</SelectItem>)}
                  </SelectContent>
                </Select>
                <div className="flex items-center gap-1">
                  <Button variant="outline" size="icon" className="h-8 w-8" disabled={currentPage === 1} onClick={() => setCurrentPage((p) => p - 1)}>
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <span className="text-xs text-muted-foreground px-2">{currentPage} / {totalPages || 1}</span>
                  <Button variant="outline" size="icon" className="h-8 w-8" disabled={currentPage >= totalPages} onClick={() => setCurrentPage((p) => p + 1)}>
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          </>
        )}
      </Card>

      {/* Create Ticket Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="w-[95vw] sm:max-w-5xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <TicketCheck className="h-5 w-5 text-blue-600" />
              Create New Trouble Ticket
            </DialogTitle>
            <DialogDescription>Report and track technical issues efficiently</DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 py-2">
            {/* Kolom Kiri: Customer Information & Technical Data */}
            <div className="lg:col-span-5 space-y-4">
              <div>
                <Label className="text-xs font-semibold text-muted-foreground mb-1.5 block uppercase">Customer Information</Label>
                <Separator className="mb-3" />
                <div className="space-y-3">
                  <div>
                    <Label className="text-sm mb-1.5 block">Customer *</Label>
                    <Input
                      value={customerSearch}
                      onChange={(e) => setCustomerSearch(e.target.value)}
                      placeholder="Cari nama, no. telp, alamat..."
                      className="mb-2"
                    />
                    {loadingCustomers ? (
                      <div className="text-xs text-muted-foreground py-2">Memuat pelanggan...</div>
                    ) : (
                      <div className="max-h-40 overflow-y-auto border rounded-md">
                        {filteredCustomers.slice(0, 30).map((c) => (
                          <div
                            key={c.id}
                            className={`px-3 py-2 cursor-pointer hover:bg-blue-50 dark:hover:bg-blue-950/40 text-sm flex items-center gap-2 ${formData.pelanggan_id === c.id ? 'bg-blue-50 dark:bg-blue-950/50 font-semibold' : ''}`}
                            onClick={() => {
                              setFormData((p) => ({ ...p, pelanggan_id: c.id, data_teknis_id: 0 }))
                              fetchTechnicalData(c.id)
                            }}
                          >
                            <div className="h-6 w-6 rounded-full bg-blue-100 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 flex items-center justify-center text-xs font-bold flex-shrink-0">
                              {c.nama.charAt(0).toUpperCase()}
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="font-medium truncate">{c.nama}</div>
                              <div className="text-xs text-muted-foreground truncate">{c.no_telp || '-'} | {c.alamat || '-'}</div>
                            </div>
                            {formData.pelanggan_id === c.id && <CheckCircle2 className="h-4 w-4 text-blue-600 shrink-0" />}
                          </div>
                        ))}
                        {filteredCustomers.length === 0 && (
                          <div className="text-xs text-muted-foreground text-center py-3">Pelanggan tidak ditemukan</div>
                        )}
                      </div>
                    )}
                  </div>

                  <div>
                    <Label className="text-sm mb-1.5 block">Technical Data (Optional)</Label>
                    <Select
                      disabled={!formData.pelanggan_id || loadingTechData}
                      value={formData.data_teknis_id ? String(formData.data_teknis_id) : '__none__'}
                      onValueChange={(v) => setFormData((p) => ({ ...p, data_teknis_id: v === '__none__' ? 0 : Number(v) }))}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue
                          placeholder={
                            !formData.pelanggan_id
                              ? 'Select customer first'
                              : loadingTechData
                                ? 'Loading...'
                                : technicalData.length === 0
                                  ? 'No technical data available'
                                  : 'Select technical data...'
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__none__">-- Tidak ada / Kosongkan --</SelectItem>
                        {technicalData.map((d) => (
                          <SelectItem key={d.id} value={String(d.id)}>
                            {d.display_name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-[11px] text-muted-foreground mt-1">
                      {formData.pelanggan_id ? 'Select technical data if available' : 'Select customer first'}
                    </p>
                  </div>

                  {/* Customer & Technical Preview */}
                  {(() => {
                    const selectedCust = customers.find((c) => c.id === formData.pelanggan_id)
                    const selectedTech = technicalData.find((d) => d.id === formData.data_teknis_id)
                    if (!selectedCust) return null
                    return (
                      <div className="bg-muted/50 border rounded-lg p-3 text-xs space-y-1.5">
                        <div className="flex items-center justify-between">
                          <span className="font-semibold text-foreground/90">{selectedCust.nama}</span>
                          {selectedCust.harga_layanan?.brand && (
                            <Badge variant="secondary" className="text-[10px] bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300">
                              {selectedCust.harga_layanan.brand}
                            </Badge>
                          )}
                        </div>
                        <div className="text-muted-foreground">{selectedCust.alamat || '-'}</div>
                        <div className="text-muted-foreground">Telp: {selectedCust.no_telp || '-'}</div>
                        {selectedTech && (
                          <div className="pt-1.5 border-t flex flex-wrap items-center gap-1.5 text-muted-foreground">
                            <Server className="h-3.5 w-3.5 text-cyan-600 shrink-0" />
                            <span className="font-mono">{selectedTech.ip_pelanggan || '-'}</span>
                            <span>•</span>
                            <span>OLT: {selectedTech.olt || selectedTech.olt_custom || '-'}</span>
                            <span>•</span>
                            <span>PON: {selectedTech.pon || '-'}</span>
                          </div>
                        )}
                      </div>
                    )
                  })()}
                </div>
              </div>
            </div>

            {/* Kolom Kanan: Indikasi Awal, Classification, Evidence */}
            <div className="lg:col-span-7 space-y-4">
              {/* Indikasi Awal */}
              <div>
                <Label className="text-xs font-semibold text-muted-foreground mb-1.5 block uppercase">Indikasi Awal</Label>
                <Separator className="mb-3" />
                <div className="space-y-3">
                  <div>
                    <Label className="text-sm mb-1.5 block">Title *</Label>
                    <Select value={formData.title} onValueChange={(v) => setFormData((p) => ({ ...p, title: v }))}>
                      <SelectTrigger className="w-full"><SelectValue placeholder="Pilih indikasi awal..." /></SelectTrigger>
                      <SelectContent>
                        {titleOptions.map((t) => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label className="text-sm mb-1.5 block">Description *</Label>
                    <Textarea
                      value={formData.description}
                      onChange={(e) => setFormData((p) => ({ ...p, description: e.target.value }))}
                      placeholder="Provide detailed information about the issue... (min. 10 karakter)"
                      rows={3}
                    />
                  </div>
                </div>
              </div>

              {/* Classification */}
              <div>
                <Label className="text-xs font-semibold text-muted-foreground mb-1.5 block uppercase">Classification</Label>
                <Separator className="mb-3" />
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <Label className="text-sm mb-1.5 block">Category *</Label>
                    <Select value={formData.category} onValueChange={(v) => setFormData((p) => ({ ...p, category: v }))}>
                      <SelectTrigger className="w-full"><SelectValue placeholder="Pilih kategori..." /></SelectTrigger>
                      <SelectContent>
                        {categoryFormOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label className="text-sm mb-1.5 block flex items-center gap-1.5">
                      <UserCheck className="h-3.5 w-3.5 text-blue-600" />
                      Assigned To (Teknisi Only)
                    </Label>
                    <Select
                      disabled={loadingTeknisi}
                      value={formData.assigned_to ? String(formData.assigned_to) : '__auto__'}
                      onValueChange={(v) => setFormData((p) => ({ ...p, assigned_to: v === '__auto__' ? null : Number(v) }))}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder={loadingTeknisi ? 'Loading teknisi...' : 'Auto-Assign berdasarkan lokasi'} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__auto__">-- Auto-Assign (Kosongkan) --</SelectItem>
                        {teknisiUsers.map((u) => (
                          <SelectItem key={u.id} value={String(u.id)}>
                            {u.name || u.nama || u.username || `Teknisi #${u.id}`} ({u.role?.name || 'Teknisi'})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-[11px] text-muted-foreground mt-1">
                      Kosongkan (Clear) untuk Auto-Assign berdasarkan lokasi pelanggan
                    </p>
                  </div>
                </div>
              </div>

              {/* Evidence Upload */}
              <div>
                <div className="flex items-center gap-2 mb-1.5">
                  <Label className="text-xs font-semibold text-muted-foreground uppercase">Evidence Upload</Label>
                  <Badge variant="secondary" className="text-[10px]">Optional</Badge>
                </div>
                <Separator className="mb-3" />
                <label className="flex items-center justify-center gap-2 border-2 border-dashed rounded-xl p-3.5 cursor-pointer hover:bg-muted/50 transition-colors text-sm text-muted-foreground">
                  <Upload className="h-4 w-4 text-blue-600" />
                  <span>Upload screenshots, photos, or documents</span>
                  <input
                    type="file"
                    multiple
                    accept="image/*,.pdf,.doc,.docx,.txt"
                    className="hidden"
                    onChange={handleEvidenceFileSelect}
                  />
                </label>
                {evidenceFiles.length > 0 && (
                  <div className="mt-2.5 space-y-2">
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <span className="font-medium flex items-center gap-1">
                        <Paperclip className="h-3.5 w-3.5" />
                        {evidenceFiles.length} file dipilih
                      </span>
                      <Button type="button" variant="ghost" size="sm" className="h-6 text-xs text-red-500" onClick={() => setEvidenceFiles([])}>
                        Clear All
                      </Button>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      {evidenceFiles.map((f, idx) => (
                        <div key={idx} className="flex items-center justify-between gap-2 border rounded-lg px-3 py-1.5 bg-muted/50 text-xs">
                          <div className="flex items-center gap-2 min-w-0">
                            <FileText className="h-4 w-4 text-blue-600 shrink-0" />
                            <div className="min-w-0">
                              <div className="font-medium truncate">{f.name}</div>
                              <div className="text-[10px] text-muted-foreground">{formatFileSize(f.size)}</div>
                            </div>
                          </div>
                          <button type="button" onClick={() => removeEvidenceFile(idx)} className="text-muted-foreground hover:text-red-600">
                            <X className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={creating}>Cancel</Button>
            <Button
              onClick={handleCreateTicket}
              disabled={creating || !formData.pelanggan_id || !formData.title || !formData.category || !formData.description || formData.description.length < 10}
            >
              {creating && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              Create Ticket
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <Trash2 className="h-5 w-5" />
              Konfirmasi Hapus Ticket
            </DialogTitle>
            <DialogDescription>Tindakan ini tidak dapat dibatalkan.</DialogDescription>
          </DialogHeader>
          {ticketToDelete && (
            <div className="bg-red-50 dark:bg-red-950/40 border border-red-100 dark:border-red-900/50 rounded-lg p-4 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Ticket:</span>
                <span className="font-semibold">{ticketToDelete.ticket_number}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Judul:</span>
                <span className="font-medium">{ticketToDelete.title}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Status:</span>
                <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full border ${getStatusColor(ticketToDelete.status)}`}>
                  {formatStatus(ticketToDelete.status)}
                </span>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>Batal</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
              Hapus Ticket
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
