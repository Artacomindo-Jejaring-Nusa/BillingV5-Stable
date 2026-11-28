import { useState, useEffect, useCallback, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import apiClient, { getBaseURL } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
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
  ArrowLeft,
  RefreshCw,
  TicketCheck,
  User,
  MapPin,
  Clock,
  PauseCircle,
  CheckCircle2,
  Archive,
  XCircle,
  AlertCircle,
  Timer,
  UserPlus,
  Settings2,
  History,
  FileText,
  Paperclip,
  Image as ImageIcon,
  Video,
  File,
  Download,
  Upload,
  X,
  Loader2,
  Phone,
  Server,
  AlertTriangle,
  ArrowRight,
  Check,
} from 'lucide-react'

interface TroubleTicket {
  id: number
  ticket_number: string
  title: string
  description?: string
  status: string
  priority: string
  category: string
  created_at: string
  updated_at: string
  resolved_at?: string | null
  downtime_start?: string | null
  downtime_end?: string | null
  total_downtime_minutes?: number
  pending_start?: string | null
  total_pending_minutes: number
  pelanggan?: {
    id: number
    nama: string
    alamat: string
    no_telp?: string
    harga_layanan?: { brand: string }
  }
  data_teknis?: {
    id_pelanggan: string
    ip_pelanggan?: string
    olt?: string
    pon?: string
  }
  assigned_user?: {
    id: number
    name: string
    role?: { name: string }
  }
}

interface ActionTaken {
  id: number
  ticket_id: number
  action_description: string
  summary_problem: string
  summary_action: string
  evidence?: string
  notes?: string
  taken_by?: number
  created_at: string
  taken_user?: {
    id: number
    name: string
  }
}

interface TicketHistory {
  id: number
  ticket_id: number
  old_status?: string
  new_status: string
  changed_by?: number
  notes?: string
  created_at: string
  changed_user?: {
    id: number
    name: string
  }
}

interface TechnicianUser {
  id: number
  name: string
  role?: { name: string }
}

const actionStatusOptions = [
  { label: 'Open', value: 'open' },
  { label: 'Pending', value: 'pending_customer' },
  { label: 'Stop Pending (Resume)', value: 'stop_pending' },
  { label: 'Closed', value: 'closed' },
]

function msToHHMMSS(ms: number) {
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

function formatDateTime(dateStr: string) {
  return new Date(dateStr).toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatStatus(status: string) {
  if (status === 'pending_customer' || status === 'pending_vendor') return 'Pending'
  return status.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function formatCategory(cat: string) {
  return cat.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function formatPriority(p: string) {
  return p ? p.charAt(0).toUpperCase() + p.slice(1) : '-'
}

function formatDowntimeMinutes(minutes: number) {
  if (minutes < 60) return `${minutes}m`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  if (m === 0) return `${h}h`
  return `${h}h ${m}m`
}

function getStatusBadgeClass(status: string) {
  const map: Record<string, string> = {
    open: 'bg-amber-100 text-amber-800 border-amber-200',
    in_progress: 'bg-blue-100 text-blue-800 border-blue-200',
    pending_customer: 'bg-orange-100 text-orange-800 border-orange-200',
    pending_vendor: 'bg-orange-100 text-orange-800 border-orange-200',
    resolved: 'bg-emerald-100 text-emerald-800 border-emerald-200',
    closed: 'bg-muted text-foreground/90 border-border',
    cancelled: 'bg-red-100 text-red-800 border-red-200',
  }
  return map[status] || 'bg-muted text-foreground/90 border-border'
}

function getPriorityBadgeClass(priority: string) {
  const map: Record<string, string> = {
    low: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    medium: 'bg-amber-50 text-amber-700 border-amber-200',
    high: 'bg-red-50 text-red-700 border-red-200',
    critical: 'bg-red-100 text-red-900 border-red-300',
  }
  return map[priority] || 'bg-muted/50 text-foreground/90 border-border'
}

function getStatusIcon(status: string) {
  const map: Record<string, React.ReactNode> = {
    open: <Clock className="h-3.5 w-3.5" />,
    in_progress: <Timer className="h-3.5 w-3.5" />,
    pending_customer: <PauseCircle className="h-3.5 w-3.5" />,
    pending_vendor: <PauseCircle className="h-3.5 w-3.5" />,
    resolved: <CheckCircle2 className="h-3.5 w-3.5" />,
    closed: <Archive className="h-3.5 w-3.5" />,
    cancelled: <XCircle className="h-3.5 w-3.5" />,
  }
  return map[status] || <AlertCircle className="h-3.5 w-3.5" />
}

function getInitials(name: string) {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

function isImageUrl(url: string) {
  if (url.startsWith('blob:')) return false
  const ext = url.split('.').pop()?.toLowerCase().split('?')[0]
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'].includes(ext || '')
}

function isVideoUrl(url: string) {
  if (url.startsWith('blob:')) return false
  const ext = url.split('.').pop()?.toLowerCase().split('?')[0]
  return ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'webm'].includes(ext || '')
}

function getEvidenceName(url: string) {
  if (!url) return 'Unknown file'
  if (url.startsWith('blob:')) return 'Evidence_Image.png'
  if (url.includes('/')) {
    return (url.split('/').pop() || url).split('?')[0]
  }
  return url.split('?')[0]
}

function getFullEvidenceUrl(url: string): string | undefined {
  if (!url) return undefined
  const base = getBaseURL().replace(/\/api\/v1\/?$/, '')
  if (url.startsWith('blob:')) return undefined
  if (url.startsWith('http')) return url
  if (url.startsWith('/static/')) return `${base}${url}`
  return `${base}/static/uploads/evidence/${url}`
}

export function TroubleTicketActionView() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const canEditTicket = useAuthStore((s) => s.hasPermission('edit_trouble_tickets'))
  const ticketId = Number(id)

  const [loading, setLoading] = useState(false)
  const [ticket, setTicket] = useState<TroubleTicket | null>(null)
  const [actionHistory, setActionHistory] = useState<ActionTaken[]>([])
  const [statusHistory, setStatusHistory] = useState<TicketHistory[]>([])
  const [now, setNow] = useState(new Date())

  // Assign dialog
  const [showAssignDialog, setShowAssignDialog] = useState(false)
  const [technicians, setTechnicians] = useState<TechnicianUser[]>([])
  const [loadingTechs, setLoadingTechs] = useState(false)
  const [selectedTechId, setSelectedTechId] = useState<string>('')
  const [assignNotes, setAssignNotes] = useState('')
  const [assigning, setAssigning] = useState(false)

  // Action dialog
  const [showActionDialog, setShowActionDialog] = useState(false)
  const [actionStatus, setActionStatus] = useState<string>('')
  const [actionDesc, setActionDesc] = useState('')
  const [summaryProblem, setSummaryProblem] = useState('')
  const [summaryAction, setSummaryAction] = useState('')
  const [actionNotes, setActionNotes] = useState('')
  const [evidenceFiles, setEvidenceFiles] = useState<File[]>([])
  const [submittingAction, setSubmittingAction] = useState(false)
  const [actionError, setActionError] = useState('')

  // Evidence viewer
  const [showEvidenceDialog, setShowEvidenceDialog] = useState(false)
  const [previewEvidenceUrl, setPreviewEvidenceUrl] = useState('')
  const [previewEvidenceType, setPreviewEvidenceType] = useState<'image' | 'video'>('image')
  const [previewEvidenceName, setPreviewEvidenceName] = useState('')

  const fetchAllData = useCallback(async () => {
    if (!ticketId) return
    setLoading(true)
    try {
      const [ticketRes, actionsRes, historyRes] = await Promise.allSettled([
        apiClient.get(`/trouble-tickets/${ticketId}`),
        apiClient.get(`/trouble-tickets/${ticketId}/actions`),
        apiClient.get(`/trouble-tickets/${ticketId}/history`),
      ])
      if (ticketRes.status === 'fulfilled') {
        setTicket(ticketRes.value.data.data || ticketRes.value.data)
      }
      if (actionsRes.status === 'fulfilled') {
        setActionHistory(actionsRes.value.data.data || actionsRes.value.data || [])
      }
      if (historyRes.status === 'fulfilled') {
        setStatusHistory(historyRes.value.data.data || historyRes.value.data || [])
      }
    } catch (err) {
      console.error('Failed to load ticket data:', err)
    } finally {
      setLoading(false)
    }
  }, [ticketId])

  useEffect(() => {
    fetchAllData()
  }, [fetchAllData])

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  const isResolvedOrClosed = useMemo(() => {
    if (!ticket) return false
    const s = (ticket.status || '').toLowerCase().trim()
    return s === 'resolved' || s === 'closed' || s === 'cancelled' || !!ticket.downtime_end
  }, [ticket])

  const isPending = useMemo(() => {
    if (!ticket) return false
    const s = (ticket.status || '').toLowerCase().trim()
    return s === 'pending' || s === 'pending_customer' || s === 'pending_vendor'
  }, [ticket])

  const completedDowntimeMinutes = useMemo(() => {
    if (!ticket) return 0
    if (typeof ticket.total_downtime_minutes === 'number' && ticket.total_downtime_minutes > 0) {
      return ticket.total_downtime_minutes
    }
    if (ticket.downtime_start && (ticket.downtime_end || ticket.resolved_at || ticket.updated_at)) {
      const start = new Date(ticket.downtime_start)
      const end = new Date(ticket.downtime_end || ticket.resolved_at || ticket.updated_at || start)
      const pendingMs = (ticket.total_pending_minutes || 0) * 60000
      return Math.floor(Math.max(0, end.getTime() - start.getTime() - pendingMs) / 60000)
    }
    return 0
  }, [ticket])

  const { liveDowntime, livePending, downtimeMinutesLive } = useMemo(() => {
    if (!ticket?.created_at) {
      return { liveDowntime: '00:00:00', livePending: '00:00:00', downtimeMinutesLive: 0 }
    }
    const start = new Date(ticket.downtime_start || ticket.created_at)
    const totalPendingMs = (ticket.total_pending_minutes || 0) * 60000

    if (isResolvedOrClosed) {
      const end = new Date(ticket.downtime_end || ticket.resolved_at || ticket.updated_at || start)
      const diff = Math.max(0, end.getTime() - start.getTime() - totalPendingMs)
      return {
        liveDowntime: msToHHMMSS(diff),
        livePending: msToHHMMSS(totalPendingMs),
        downtimeMinutesLive: Math.floor(diff / 60000),
      }
    }

    if (isPending) {
      const pStart = ticket.pending_start ? new Date(ticket.pending_start) : now
      const frozenMs = Math.max(0, pStart.getTime() - start.getTime() - totalPendingMs)
      const currentPendingMs = Math.max(0, now.getTime() - pStart.getTime())
      return {
        liveDowntime: msToHHMMSS(frozenMs),
        livePending: msToHHMMSS(totalPendingMs + currentPendingMs),
        downtimeMinutesLive: Math.floor(frozenMs / 60000),
      }
    }

    const diff = Math.max(0, now.getTime() - start.getTime() - totalPendingMs)
    return {
      liveDowntime: msToHHMMSS(diff),
      livePending: msToHHMMSS(totalPendingMs),
      downtimeMinutesLive: Math.floor(diff / 60000),
    }
  }, [ticket, now, isResolvedOrClosed, isPending])

  const downtimeSeverity = useMemo(() => {
    if (isPending) {
      return {
        bg: 'bg-orange-50 border-orange-200',
        text: 'text-orange-700',
        badge: 'bg-orange-100 text-orange-800',
        label: 'Clock Paused (Pending)',
      }
    }
    const mins = isResolvedOrClosed ? completedDowntimeMinutes : downtimeMinutesLive
    if (mins < 60) {
      return {
        bg: 'bg-emerald-50 border-emerald-200',
        text: 'text-emerald-700',
        badge: 'bg-emerald-100 text-emerald-800',
        label: isResolvedOrClosed ? 'Total Downtime Recorded' : 'Running Normally',
      }
    }
    if (mins < 240) {
      return {
        bg: 'bg-amber-50 border-amber-200',
        text: 'text-amber-700',
        badge: 'bg-amber-100 text-amber-800',
        label: isResolvedOrClosed ? 'Total Downtime Recorded' : 'Attention Needed',
      }
    }
    return {
      bg: 'bg-red-50 border-red-200',
      text: 'text-red-700',
      badge: 'bg-red-100 text-red-800',
      label: isResolvedOrClosed ? 'Total Downtime Recorded' : mins < 1440 ? 'Critical Downtime' : 'Severe Downtime',
    }
  }, [isPending, isResolvedOrClosed, completedDowntimeMinutes, downtimeMinutesLive])

  const parsedActions = useMemo(() => {
    return actionHistory.map((action) => {
      let evidenceArray: string[] = []
      if (action.evidence) {
        try {
          const parsed = JSON.parse(action.evidence)
          if (Array.isArray(parsed)) evidenceArray = parsed
        } catch {
          evidenceArray = []
        }
      }
      return { ...action, evidenceArray }
    })
  }, [actionHistory])

  // Assign handlers
  const openAssignDialog = async () => {
    setSelectedTechId(ticket?.assigned_user?.id ? String(ticket.assigned_user.id) : '')
    setAssignNotes('')
    setShowAssignDialog(true)
    setLoadingTechs(true)
    try {
      let res
      try {
        res = await apiClient.get('/trouble-tickets/technicians')
      } catch {
        res = await apiClient.get('/users?limit=1000')
      }
      const raw: TechnicianUser[] = res.data.data || res.data || []
      const seenIds = new Set<number>()
      const seenNames = new Set<string>()
      const unique = raw.filter((u) => {
        if (seenIds.has(u.id)) return false
        seenIds.add(u.id)
        const n = u.name?.trim().toLowerCase()
        if (n && seenNames.has(n)) return false
        if (n) seenNames.add(n)
        return true
      })
      unique.sort((a, b) => (a.name || '').localeCompare(b.name || ''))
      setTechnicians(unique)
    } catch {
      setTechnicians([])
    } finally {
      setLoadingTechs(false)
    }
  }

  const handleAssignSubmit = async () => {
    if (!selectedTechId) return
    setAssigning(true)
    try {
      await apiClient.post(`/trouble-tickets/${ticketId}/assign`, {
        assigned_to: Number(selectedTechId),
        notes: assignNotes,
      })
      setShowAssignDialog(false)
      fetchAllData()
    } catch (err) {
      console.error('Failed to assign ticket:', err)
      alert('Gagal meng-assign ticket.')
    } finally {
      setAssigning(false)
    }
  }

  // Action dialog handlers
  const openActionDialog = () => {
    setActionStatus('')
    setActionDesc('')
    setSummaryProblem('')
    setSummaryAction('')
    setActionNotes('')
    setEvidenceFiles([])
    setActionError('')
    setShowActionDialog(true)
  }

  const effectiveStatus = actionStatus === 'stop_pending' ? 'open' : actionStatus
  const requiresSummary = effectiveStatus === 'closed' || effectiveStatus === 'resolved'
  const showActionDetails = Boolean(actionStatus) || Boolean(actionDesc)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    setEvidenceFiles((prev) => [...prev, ...Array.from(e.target.files!)])
    e.target.value = ''
  }

  const removeEvidenceFile = (idx: number) => {
    setEvidenceFiles((prev) => prev.filter((_, i) => i !== idx))
  }

  const handleActionSubmit = async () => {
    setActionError('')
    if (requiresSummary && (!actionDesc.trim() || !summaryProblem.trim() || !summaryAction.trim())) {
      setActionError('Action Description, Summary Problem, dan Summary Action wajib diisi saat menutup ticket.')
      return
    }
    if (!actionStatus && !actionDesc.trim() && !actionNotes.trim() && evidenceFiles.length === 0) {
      setActionError('Pilih status baru atau isi Action Description / Notes.')
      return
    }

    setSubmittingAction(true)
    try {
      let evidenceUrls: string[] = []
      if (evidenceFiles.length > 0) {
        evidenceUrls = await Promise.all(
          evidenceFiles.map(async (fileItem) => {
            const fd = new FormData()
            fd.append('file', fileItem)
            const r = await apiClient.post('/uploads/evidence', fd, {
              headers: { 'Content-Type': 'multipart/form-data' },
            })
            return r.data.file_url
          })
        )
      }

      const payload = {
        status: effectiveStatus || '',
        notes: actionNotes || '',
        action_description: actionDesc || '',
        summary_problem: summaryProblem || '',
        summary_action: summaryAction || '',
        evidence: evidenceUrls.length > 0 ? JSON.stringify(evidenceUrls) : '',
      }

      if (effectiveStatus) {
        await apiClient.post(`/trouble-tickets/${ticketId}/status`, payload)
      } else {
        const { status: _s, ...actionOnly } = payload
        await apiClient.post(`/trouble-tickets/${ticketId}/action`, actionOnly)
      }

      setShowActionDialog(false)
      fetchAllData()
    } catch (err: unknown) {
      const error = err as { response?: { status?: number; data?: { detail?: string } } }
      let msg = 'Gagal menyimpan action ticket.'
      if (error.response?.status === 413) msg = 'Ukuran file terlalu besar (maks. 10MB).'
      else if (error.response?.status === 415) msg = 'Format file tidak didukung.'
      else if (error.response?.status === 422) msg = 'Validasi gagal. Periksa field yang wajib diisi.'
      else if (error.response?.data?.detail) msg = error.response.data.detail
      setActionError(msg)
    } finally {
      setSubmittingAction(false)
    }
  }

  const openEvidence = (rawUrl: string) => {
    const fullUrl = getFullEvidenceUrl(rawUrl)
    if (!fullUrl) {
      alert('File evidence tidak tersedia.')
      return
    }
    if (isImageUrl(rawUrl)) {
      setPreviewEvidenceUrl(fullUrl)
      setPreviewEvidenceType('image')
      setPreviewEvidenceName(getEvidenceName(rawUrl))
      setShowEvidenceDialog(true)
    } else if (isVideoUrl(rawUrl)) {
      setPreviewEvidenceUrl(fullUrl)
      setPreviewEvidenceType('video')
      setPreviewEvidenceName(getEvidenceName(rawUrl))
      setShowEvidenceDialog(true)
    } else {
      window.open(fullUrl, '_blank')
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-card border rounded-xl p-4 md:p-5 shadow-sm">
        <div className="flex items-start sm:items-center gap-3">
          <Button variant="outline" size="icon" onClick={() => navigate('/trouble-tickets')} className="shrink-0">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center flex-wrap gap-2">
              <h1 className="text-xl font-bold text-foreground">Ticket Action Management</h1>
              {ticket && (
                <>
                  <Badge variant="secondary" className="font-mono text-xs bg-blue-50 text-blue-700">
                    <TicketCheck className="h-3 w-3 mr-1" />
                    {ticket.ticket_number}
                  </Badge>
                  <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full border ${getStatusBadgeClass(ticket.status)}`}>
                    {getStatusIcon(ticket.status)}
                    {formatStatus(ticket.status)}
                  </span>
                </>
              )}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Kelola status, tracking downtime, assignment teknisi, dan riwayat penanganan gangguan
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <Button variant="outline" size="sm" onClick={fetchAllData} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canEditTicket && (
            <Button variant="outline" size="sm" onClick={openAssignDialog} disabled={!ticket}>
              <UserPlus className="h-4 w-4 mr-1.5" />
              Assign Teknisi
            </Button>
          )}
          {canEditTicket && (
            <Button size="sm" onClick={openActionDialog} disabled={!ticket}>
              <Settings2 className="h-4 w-4 mr-1.5" />
              Update / Add Action
            </Button>
          )}
        </div>
      </div>

      {loading && !ticket ? (
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-7 w-7 animate-spin text-blue-600" />
          <span className="ml-2 text-sm text-muted-foreground">Memuat detail ticket...</span>
        </div>
      ) : ticket ? (
        <>
          {/* Ticket Info + Downtime Clock */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Ticket & Customer Details */}
            <Card className="lg:col-span-7 shadow-sm">
              <CardHeader className="pb-3">
                <CardTitle className="text-base font-semibold flex items-center gap-2">
                  <FileText className="h-4 w-4 text-blue-600" />
                  Detail Ticket & Pelanggan
                </CardTitle>
              </CardHeader>
              <Separator />
              <CardContent className="p-5 space-y-4">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant="outline" className="text-xs">
                    Kategori: {formatCategory(ticket.category)}
                  </Badge>
                  <span className={`inline-flex items-center text-xs font-semibold px-2.5 py-0.5 rounded-full border ${getPriorityBadgeClass(ticket.priority)}`}>
                    Prioritas: {formatPriority(ticket.priority)}
                  </span>
                  {ticket.pelanggan?.harga_layanan?.brand && (
                    <Badge className="bg-emerald-50 text-emerald-700 hover:bg-emerald-50 text-xs">
                      {ticket.pelanggan.harga_layanan.brand}
                    </Badge>
                  )}
                </div>

                <div>
                  <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Indikasi Awal / Judul</div>
                  <div className="text-base font-semibold text-foreground mt-0.5">{ticket.title}</div>
                  {ticket.description && (
                    <p className="text-sm text-muted-foreground mt-1.5 bg-muted/50 p-3 rounded-lg border">
                      {ticket.description}
                    </p>
                  )}
                </div>

                <Separator />

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
                  <div className="space-y-2">
                    <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Pelanggan</div>
                    <div className="flex items-start gap-2">
                      <User className="h-4 w-4 text-blue-600 mt-0.5 shrink-0" />
                      <div>
                        <div className="font-medium text-foreground">{ticket.pelanggan?.nama || 'N/A'}</div>
                        {ticket.pelanggan?.no_telp && (
                          <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                            <Phone className="h-3 w-3" /> {ticket.pelanggan.no_telp}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="flex items-start gap-2">
                      <MapPin className="h-4 w-4 text-emerald-600 mt-0.5 shrink-0" />
                      <span className="text-xs text-muted-foreground">{ticket.pelanggan?.alamat || 'N/A'}</span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Penanganan & Teknis</div>
                    <div className="flex items-center gap-2">
                      <UserPlus className="h-4 w-4 text-purple-600 shrink-0" />
                      <span className="text-sm">
                        Assigned: <strong className="text-foreground">{ticket.assigned_user?.name || 'Unassigned'}</strong>
                      </span>
                    </div>
                    {ticket.data_teknis && (
                      <div className="flex items-start gap-2">
                        <Server className="h-4 w-4 text-cyan-600 mt-0.5 shrink-0" />
                        <div className="text-xs text-muted-foreground space-y-0.5">
                          <div>ID: <span className="font-mono">{ticket.data_teknis.id_pelanggan}</span></div>
                          {ticket.data_teknis.ip_pelanggan && (
                            <div>IP: <span className="font-mono">{ticket.data_teknis.ip_pelanggan}</span></div>
                          )}
                          {(ticket.data_teknis.olt || ticket.data_teknis.pon) && (
                            <div>OLT/PON: {ticket.data_teknis.olt || '-'} / {ticket.data_teknis.pon || '-'}</div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Downtime Clock Card */}
            <Card className="lg:col-span-5 shadow-sm">
              <CardHeader className="pb-3">
                <CardTitle className="text-base font-semibold flex items-center gap-2">
                  <Timer className="h-4 w-4 text-red-600" />
                  Downtime Clock Tracking
                </CardTitle>
              </CardHeader>
              <Separator />
              <CardContent className="p-5 space-y-4">
                <div className={`rounded-xl border p-5 text-center ${downtimeSeverity.bg}`}>
                  <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-1">
                    Total Downtime
                  </div>
                  <div className={`text-4xl font-mono font-bold tracking-tight ${downtimeSeverity.text}`}>
                    {liveDowntime}
                  </div>
                  <div className="mt-2.5 flex items-center justify-center gap-2">
                    <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full ${downtimeSeverity.badge}`}>
                      {isPending ? <PauseCircle className="h-3 w-3" /> : isResolvedOrClosed ? <CheckCircle2 className="h-3 w-3" /> : <Timer className="h-3 w-3" />}
                      {downtimeSeverity.label}
                    </span>
                    {isResolvedOrClosed && completedDowntimeMinutes > 0 && (
                      <Badge variant="outline" className="text-xs bg-card">
                        {formatDowntimeMinutes(completedDowntimeMinutes)}
                      </Badge>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div className="bg-muted/50 border rounded-lg p-3">
                    <div className="text-muted-foreground font-semibold uppercase">Start Time</div>
                    <div className="font-medium text-foreground mt-1">
                      {ticket.created_at ? formatDateTime(ticket.created_at) : '-'}
                    </div>
                  </div>
                  <div className="bg-muted/50 border rounded-lg p-3">
                    <div className="text-muted-foreground font-semibold uppercase">End Time</div>
                    <div className="font-medium text-foreground mt-1">
                      {ticket.downtime_end
                        ? formatDateTime(ticket.downtime_end)
                        : isResolvedOrClosed
                          ? formatDateTime(ticket.resolved_at || ticket.updated_at)
                          : 'Masih berjalan'}
                    </div>
                  </div>
                </div>

                {(ticket.total_pending_minutes > 0 || ticket.pending_start) && (
                  <div className="flex items-center justify-between bg-orange-50 border border-orange-200 rounded-lg px-3.5 py-2.5 text-xs">
                    <div className="flex items-center gap-2 text-orange-800 font-medium">
                      <PauseCircle className="h-4 w-4 text-orange-600" />
                      Pending Duration
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-bold text-orange-800">{livePending}</span>
                      {ticket.pending_start && (
                        <Badge className="bg-orange-500 text-white text-[10px] h-4">Active</Badge>
                      )}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Action Taken History & Status History */}
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            {/* Action Taken History */}
            <Card className="shadow-sm">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold flex items-center gap-2">
                    <Settings2 className="h-4 w-4 text-blue-600" />
                    Action Taken History
                  </CardTitle>
                  <Badge variant="secondary">{parsedActions.length}</Badge>
                </div>
              </CardHeader>
              <Separator />
              <CardContent className="p-5">
                {parsedActions.length === 0 ? (
                  <div className="text-center py-12 text-muted-foreground">
                    <History className="h-10 w-10 mx-auto mb-2 opacity-50" />
                    <p className="text-sm font-medium">Belum ada Action Taken</p>
                    <p className="text-xs mt-1">Klik "Update / Add Action" untuk mencatat penanganan</p>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {parsedActions.map((item) => (
                      <div key={item.id} className="border rounded-xl p-4 bg-card space-y-3">
                        <div className="flex items-center justify-between gap-2">
                          <div className="flex items-center gap-2.5">
                            {item.taken_user ? (
                              <>
                                <div className="h-8 w-8 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center text-xs font-bold">
                                  {getInitials(item.taken_user.name)}
                                </div>
                                <div>
                                  <div className="text-sm font-semibold text-foreground">{item.taken_user.name}</div>
                                  <div className="text-[11px] text-muted-foreground">Technician</div>
                                </div>
                              </>
                            ) : (
                              <Badge variant="outline" className="text-xs">Action Taken</Badge>
                            )}
                          </div>
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {formatDateTime(item.created_at)}
                          </span>
                        </div>

                        {item.action_description && (
                          <div className="bg-blue-50/70 border border-blue-100 rounded-lg p-3">
                            <div className="text-[11px] font-bold uppercase text-blue-700 mb-1">Action Description</div>
                            <p className="text-sm text-foreground whitespace-pre-wrap">{item.action_description}</p>
                          </div>
                        )}

                        {item.summary_problem && (
                          <div className="bg-amber-50/70 border border-amber-100 rounded-lg p-3">
                            <div className="text-[11px] font-bold uppercase text-amber-700 mb-1">Summary Problem</div>
                            <p className="text-sm text-foreground whitespace-pre-wrap">{item.summary_problem}</p>
                          </div>
                        )}

                        {item.summary_action && (
                          <div className="bg-emerald-50/70 border border-emerald-100 rounded-lg p-3">
                            <div className="text-[11px] font-bold uppercase text-emerald-700 mb-1">Summary Action</div>
                            <p className="text-sm text-foreground whitespace-pre-wrap">{item.summary_action}</p>
                          </div>
                        )}

                        {item.evidenceArray.length > 0 && (
                          <div>
                            <div className="text-[11px] font-bold uppercase text-muted-foreground mb-2 flex items-center gap-1">
                              <Paperclip className="h-3 w-3" />
                              Evidence ({item.evidenceArray.length})
                            </div>
                            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                              {item.evidenceArray.map((ev, idx) => {
                                const fullUrl = getFullEvidenceUrl(ev)
                                return (
                                  <div
                                    key={`${item.id}-${idx}`}
                                    onClick={() => openEvidence(ev)}
                                    className="border rounded-lg overflow-hidden cursor-pointer hover:border-blue-400 transition-colors bg-muted/50"
                                  >
                                    {isImageUrl(ev) && fullUrl ? (
                                      <img src={fullUrl} alt={getEvidenceName(ev)} className="w-full h-24 object-cover" />
                                    ) : isVideoUrl(ev) ? (
                                      <div className="h-24 flex flex-col items-center justify-center text-red-500">
                                        <Video className="h-6 w-6 mb-1" />
                                        <span className="text-[10px]">Video</span>
                                      </div>
                                    ) : (
                                      <div className="h-24 flex flex-col items-center justify-center text-muted-foreground">
                                        <File className="h-6 w-6 mb-1" />
                                        <span className="text-[10px]">File</span>
                                      </div>
                                    )}
                                    <div className="px-2 py-1 text-[11px] text-muted-foreground truncate border-t bg-card">
                                      {getEvidenceName(ev)}
                                    </div>
                                  </div>
                                )
                              })}
                            </div>
                          </div>
                        )}

                        {item.notes && (
                          <div className="bg-muted/50 border rounded-lg p-3">
                            <div className="text-[11px] font-bold uppercase text-muted-foreground mb-1">Notes</div>
                            <p className="text-xs text-foreground/90 whitespace-pre-wrap">{item.notes}</p>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Status History */}
            <Card className="shadow-sm">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold flex items-center gap-2">
                    <History className="h-4 w-4 text-cyan-600" />
                    Status History
                  </CardTitle>
                  <Badge variant="secondary">{statusHistory.length}</Badge>
                </div>
              </CardHeader>
              <Separator />
              <CardContent className="p-5">
                {statusHistory.length === 0 ? (
                  <div className="text-center py-12 text-muted-foreground">
                    <CheckCircle2 className="h-10 w-10 mx-auto mb-2 opacity-50" />
                    <p className="text-sm font-medium">Belum ada perubahan status</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {statusHistory.map((item) => (
                      <div key={item.id} className="border rounded-xl p-4 bg-card space-y-2">
                        <div className="flex items-center justify-between flex-wrap gap-2">
                          <div className="flex items-center gap-2">
                            {item.old_status && (
                              <>
                                <span className={`inline-flex items-center text-xs px-2 py-0.5 rounded-full border ${getStatusBadgeClass(item.old_status)}`}>
                                  {formatStatus(item.old_status)}
                                </span>
                                <ArrowRight className="h-3.5 w-3.5 text-muted-foreground" />
                              </>
                            )}
                            <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full border ${getStatusBadgeClass(item.new_status)}`}>
                              {getStatusIcon(item.new_status)}
                              {formatStatus(item.new_status)}
                            </span>
                          </div>
                          <span className="text-xs text-muted-foreground">{formatDateTime(item.created_at)}</span>
                        </div>
                        {item.changed_user && (
                          <div className="text-xs text-muted-foreground">
                            Oleh: <span className="font-medium text-foreground">{item.changed_user.name}</span>
                          </div>
                        )}
                        {item.notes && (
                          <div className="text-xs text-foreground/90 bg-muted/50 border rounded-lg p-2.5">
                            {item.notes}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      ) : (
        <div className="text-center py-20 text-muted-foreground">Ticket tidak ditemukan.</div>
      )}

      {/* Assign Ticket Dialog */}
      <Dialog open={showAssignDialog} onOpenChange={setShowAssignDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <UserPlus className="h-5 w-5 text-blue-600" />
              Assign Ticket
            </DialogTitle>
            <DialogDescription>
              Pilih teknisi/tim untuk menangani ticket {ticket?.ticket_number}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div>
              <Label className="text-sm mb-1.5 block">Assign To *</Label>
              {loadingTechs ? (
                <div className="text-xs text-muted-foreground py-2">Memuat daftar teknisi...</div>
              ) : (
                <Select value={selectedTechId} onValueChange={setSelectedTechId}>
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih anggota tim..." />
                  </SelectTrigger>
                  <SelectContent>
                    {technicians.map((u) => (
                      <SelectItem key={u.id} value={String(u.id)}>
                        {u.name} {u.role?.name ? `(${u.role.name})` : ''}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
            <div>
              <Label className="text-sm mb-1.5 block">Catatan Assignment (Opsional)</Label>
              <Textarea
                value={assignNotes}
                onChange={(e) => setAssignNotes(e.target.value)}
                placeholder="Tambahkan catatan penugasan..."
                rows={3}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAssignDialog(false)} disabled={assigning}>
              Batal
            </Button>
            <Button onClick={handleAssignSubmit} disabled={assigning || !selectedTechId}>
              {assigning && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Assign
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Update Status & Add Action Dialog */}
      <Dialog open={showActionDialog} onOpenChange={setShowActionDialog}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Settings2 className="h-5 w-5 text-blue-600" />
              Update Ticket Action — {ticket?.ticket_number}
            </DialogTitle>
            <DialogDescription>
              Ubah status, catat action taken, summary problem, dan upload evidence
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            {actionError && (
              <div className="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg p-3 flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 shrink-0" />
                {actionError}
              </div>
            )}

            <div>
              <Label className="text-sm font-medium mb-1.5 block">Status Baru (Opsional)</Label>
              <Select
                value={actionStatus || '__none__'}
                onValueChange={(v) => setActionStatus(v === '__none__' ? '' : v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Biarkan kosong jika tidak mengubah status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">-- Tidak mengubah status --</SelectItem>
                  {actionStatusOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-[11px] text-muted-foreground mt-1">
                Pilih <strong>Pending</strong> untuk menjeda downtime clock, <strong>Stop Pending</strong> untuk melanjutkan, atau <strong>Closed</strong> untuk menyelesaikan ticket.
              </p>
            </div>

            {(showActionDetails || true) && (
              <div>
                <Label className="text-sm font-medium mb-1.5 block">
                  Action Description {requiresSummary && <span className="text-red-500">*</span>}
                </Label>
                <Textarea
                  value={actionDesc}
                  onChange={(e) => setActionDesc(e.target.value)}
                  placeholder="Jelaskan tindakan teknis yang dilakukan..."
                  rows={3}
                />
              </div>
            )}

            {requiresSummary && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 bg-amber-50/70 border border-amber-200 rounded-xl p-3.5">
                <div className="sm:col-span-2 text-xs font-semibold text-amber-800 flex items-center gap-1.5">
                  <AlertCircle className="h-3.5 w-3.5" />
                  Summary Wajib Diisi untuk Menutup Ticket
                </div>
                <div>
                  <Label className="text-xs font-medium mb-1 block">Summary Problem *</Label>
                  <Textarea
                    value={summaryProblem}
                    onChange={(e) => setSummaryProblem(e.target.value)}
                    placeholder="Ringkasan akar masalah..."
                    rows={3}
                    className="bg-card"
                  />
                </div>
                <div>
                  <Label className="text-xs font-medium mb-1 block">Summary Action *</Label>
                  <Textarea
                    value={summaryAction}
                    onChange={(e) => setSummaryAction(e.target.value)}
                    placeholder="Ringkasan tindakan penyelesaian..."
                    rows={3}
                    className="bg-card"
                  />
                </div>
              </div>
            )}

            {!requiresSummary && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <Label className="text-xs font-medium text-muted-foreground mb-1 block">Summary Problem (Opsional)</Label>
                  <Textarea
                    value={summaryProblem}
                    onChange={(e) => setSummaryProblem(e.target.value)}
                    placeholder="Ringkasan masalah sementara..."
                    rows={2}
                  />
                </div>
                <div>
                  <Label className="text-xs font-medium text-muted-foreground mb-1 block">Summary Action (Opsional)</Label>
                  <Textarea
                    value={summaryAction}
                    onChange={(e) => setSummaryAction(e.target.value)}
                    placeholder="Ringkasan tindakan sementara..."
                    rows={2}
                  />
                </div>
              </div>
            )}

            <div>
              <Label className="text-sm font-medium mb-1.5 block">Catatan Tambahan (Notes)</Label>
              <Textarea
                value={actionNotes}
                onChange={(e) => setActionNotes(e.target.value)}
                placeholder="Catatan tambahan..."
                rows={2}
              />
            </div>

            <div>
              <Label className="text-sm font-medium mb-1.5 block">Upload Evidence (Opsional)</Label>
              <label className="flex items-center justify-center gap-2 border-2 border-dashed rounded-xl p-4 cursor-pointer hover:bg-muted/50 transition-colors text-sm text-muted-foreground">
                <Upload className="h-4 w-4 text-blue-600" />
                <span>Pilih file gambar, video, atau dokumen</span>
                <input
                  type="file"
                  multiple
                  accept="image/*,video/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.zip"
                  className="hidden"
                  onChange={handleFileSelect}
                />
              </label>
              {evidenceFiles.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-2.5">
                  {evidenceFiles.map((f, idx) => (
                    <Badge key={idx} variant="secondary" className="gap-1.5 py-1 px-2.5">
                      <ImageIcon className="h-3 w-3" />
                      <span className="max-w-[160px] truncate">{f.name}</span>
                      <button type="button" onClick={() => removeEvidenceFile(idx)} className="hover:text-red-600">
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowActionDialog(false)} disabled={submittingAction}>
              Batal
            </Button>
            <Button onClick={handleActionSubmit} disabled={submittingAction}>
              {submittingAction ? <Loader2 className="h-4 w-4 mr-1.5 animate-spin" /> : <Check className="h-4 w-4 mr-1.5" />}
              Simpan Action
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Evidence Preview Dialog */}
      <Dialog open={showEvidenceDialog} onOpenChange={setShowEvidenceDialog}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle className="truncate">{previewEvidenceName}</DialogTitle>
          </DialogHeader>
          <div className="flex items-center justify-center bg-muted rounded-lg p-3 max-h-[70vh] overflow-hidden">
            {previewEvidenceType === 'image' ? (
              <img src={previewEvidenceUrl} alt={previewEvidenceName} className="max-h-[65vh] object-contain rounded" />
            ) : (
              <video src={previewEvidenceUrl} controls className="max-h-[65vh] w-full rounded" />
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEvidenceDialog(false)}>Tutup</Button>
            <Button asChild>
              <a href={previewEvidenceUrl} download={previewEvidenceName} target="_blank" rel="noreferrer">
                <Download className="h-4 w-4 mr-1.5" />
                Download
              </a>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
