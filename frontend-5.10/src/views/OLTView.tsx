import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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
  Router,
  Plus,
  RefreshCw,
  Pencil,
  Trash2,
  PlugZap,
  CheckCircle2,
  AlertCircle,
  Loader2,
  X,
  Network,
  Search,
  Copy,
  Activity,
  Cpu,
  ChevronLeft,
  ChevronRight,
  Hash,
  Server,
} from 'lucide-react'

interface OLT {
  id: number
  nama_olt: string
  ip_address: string
  tipe_olt: string
  username?: string
  mikrotik_server_id?: number | null
}

interface MikrotikSelectItem {
  id: number
  name: string
}

interface ZTEONUInfo {
  board: number
  pon: number
  onu_id: number
  name: string
  customer_name?: string
  onu_type: string
  serial_number: string
  rx_power: string
  status: string
}

interface ZTEONUDetail {
  board: number
  pon: number
  onu_id: number
  name: string
  customer_name?: string
  description: string
  onu_type: string
  serial_number: string
  rx_power: string
  tx_power: string
  status: string
  ip_address: string
  last_online: string
  last_offline: string
  uptime: string
  last_down_time_duration: string
  offline_reason: string
  gpon_optical_distance: string
}

interface ZTECardInfo {
  slot: number
  type: string
  role: string
}

interface ZTEPortInfo {
  name: string
  shelf: number
  slot: number
  port: number
  kind: string
  admin_status: string
  oper_status: string
  speed_mbps: number
}

interface ZTEUplinksData {
  cards: ZTECardInfo[]
  ports: ZTEPortInfo[]
}

const oltTypes = ['HSGQ', 'ZTE', 'Huawei', 'Fiberhome', 'Lainnya']

function formatDbm(val?: string | null) {
  if (!val || val === '-' || val === 'N/A') return '-'
  const clean = val.replace(/dBm/gi, '').trim()
  if (!clean || clean === '-') return '-'
  return `${clean} dBm`
}

function parsePower(val?: string | null): number | null {
  if (!val || val === '-' || val === 'N/A') return null
  const clean = val.replace(/dBm/gi, '').trim()
  const num = parseFloat(clean)
  return Number.isNaN(num) ? null : num
}

function getRxBadgeClass(powerStr?: string | null) {
  const p = parsePower(powerStr)
  if (p === null) return 'bg-muted text-muted-foreground'
  if (p >= -22.5 && p <= -10) return 'bg-emerald-100 text-emerald-800 border-emerald-200'
  if (p < -22.5 && p >= -26.5) return 'bg-amber-100 text-amber-800 border-amber-200'
  if (p < -26.5 || p > -8) return 'bg-red-100 text-red-800 border-red-200'
  return 'bg-blue-100 text-blue-800 border-blue-200'
}

function getRxQualityLabel(powerStr?: string | null) {
  const p = parsePower(powerStr)
  if (p === null) return 'Tidak Terdeteksi'
  if (p >= -22.5 && p <= -10) return 'Optimal'
  if (p < -22.5 && p >= -25.0) return 'Standar Baik'
  if (p < -25.0 && p >= -27.0) return 'Peringatan Redaman'
  if (p < -27.0) return 'Kritis (< -27 dBm)'
  if (p > -8) return 'Sinyal Terlalu Kuat'
  return 'Normal'
}

export function OLTView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_olt')
  const canEdit = hasPermission('edit_olt')
  const canDelete = hasPermission('delete_olt')
  const [olts, setOlts] = useState<OLT[]>([])
  const [mikrotikList, setMikrotikList] = useState<MikrotikSelectItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMikrotiks, setLoadingMikrotiks] = useState(false)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [testingId, setTestingId] = useState<number | null>(null)
  const [connectedIds, setConnectedIds] = useState<Record<number, boolean>>({})

  // Form Dialog
  const [showFormDialog, setShowFormDialog] = useState(false)
  const [editedItem, setEditedItem] = useState<Partial<OLT> & { password?: string }>({ tipe_olt: 'HSGQ' })

  // Delete Dialog
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<OLT | null>(null)

  // PON Explorer
  const [showExplorer, setShowExplorer] = useState(false)
  const [activeOlt, setActiveOlt] = useState<OLT | null>(null)
  const [explorerTab, setExplorerTab] = useState<'pon' | 'uplinks'>('pon')
  const [selectedBoard, setSelectedBoard] = useState(1)
  const [selectedPon, setSelectedPon] = useState(1)
  const [onusList, setOnusList] = useState<ZTEONUInfo[]>([])
  const [uplinksData, setUplinksData] = useState<ZTEUplinksData | null>(null)
  const [loadingPonData, setLoadingPonData] = useState(false)
  const [loadingUplinks, setLoadingUplinks] = useState(false)
  const [onuSearch, setOnuSearch] = useState('')
  const [itemsPerPage, setItemsPerPage] = useState(10)
  const [currentPage, setCurrentPage] = useState(1)

  // Empty ONU IDs
  const [emptyOnuIds, setEmptyOnuIds] = useState<number[]>([])
  const [showEmptyIds, setShowEmptyIds] = useState(false)
  const [loadingEmptyIds, setLoadingEmptyIds] = useState(false)

  // ONU Detail
  const [selectedOnuDetail, setSelectedOnuDetail] = useState<ZTEONUDetail | null>(null)
  const [showOnuDetail, setShowOnuDetail] = useState(false)
  const [loadingOnuDetail, setLoadingOnuDetail] = useState(false)

  // Feedback
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4000)
  }

  const fetchOLTs = useCallback(async () => {
    setLoading(true)
    try {
      const r = await apiClient.get('/olt')
      setOlts(Array.isArray(r.data) ? r.data : r.data.data || [])
    } catch {
      notify('Gagal memuat data OLT', 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  const fetchMikrotiks = useCallback(async () => {
    setLoadingMikrotiks(true)
    try {
      const r = await apiClient.get('/mikrotik_servers')
      setMikrotikList(Array.isArray(r.data) ? r.data : r.data.data || [])
    } catch {
      setMikrotikList([])
    } finally {
      setLoadingMikrotiks(false)
    }
  }, [])

  useEffect(() => {
    fetchOLTs()
    fetchMikrotiks()
  }, [fetchOLTs, fetchMikrotiks])

  const getMikrotikName = (id?: number | null) => {
    if (!id) return '-'
    return mikrotikList.find((m) => m.id === id)?.name || 'Unknown'
  }

  const isEditMode = Boolean(editedItem.id)

  const openDialog = (item?: OLT) => {
    setEditedItem(item ? { ...item, password: '' } : { tipe_olt: 'HSGQ' })
    setShowFormDialog(true)
  }

  const saveOLT = async () => {
    if (!editedItem.nama_olt?.trim() || !editedItem.ip_address?.trim() || !editedItem.tipe_olt || !editedItem.mikrotik_server_id) {
      notify('Lengkapi semua field wajib (*)', 'error')
      return
    }
    setSaving(true)
    const payload = { ...editedItem }
    if (isEditMode && !payload.password) {
      delete payload.password
    }
    try {
      if (isEditMode && payload.id) {
        await apiClient.patch(`/olt/${payload.id}`, payload)
      } else {
        await apiClient.post('/olt', payload)
      }
      setShowFormDialog(false)
      setEditedItem({})
      fetchOLTs()
      notify(`OLT berhasil ${isEditMode ? 'diperbarui' : 'ditambahkan'}`, 'success')
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: Array<{ msg?: string }> } } }
      notify(e.response?.data?.detail?.[0]?.msg || 'Gagal menyimpan data OLT', 'error')
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!itemToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/olt/${itemToDelete.id}`)
      fetchOLTs()
      notify('OLT berhasil dihapus', 'success')
    } catch {
      notify('Gagal menghapus OLT', 'error')
    } finally {
      setDeleting(false)
      setShowDeleteDialog(false)
      setItemToDelete(null)
    }
  }

  const testConnection = async (item: OLT) => {
    setTestingId(item.id)
    try {
      const res = await apiClient.post(`/olt/${item.id}/test-connection`)
      notify(res.data.message || 'Koneksi berhasil!', 'success')
      setConnectedIds((prev) => ({ ...prev, [item.id]: true }))
      setTimeout(() => {
        setConnectedIds((prev) => ({ ...prev, [item.id]: false }))
      }, 4000)
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } }
      notify(e.response?.data?.message || 'Koneksi gagal, terjadi error.', 'error')
    } finally {
      setTestingId(null)
    }
  }

  // PON Explorer Logic
  const isHSGQ = useMemo(() => {
    const type = (activeOlt?.tipe_olt || '').toLowerCase()
    const name = (activeOlt?.nama_olt || '').toLowerCase()
    return type.includes('hsgq') || name.includes('hsgq')
  }, [activeOlt])

  const availableBoards = useMemo(() => (isHSGQ ? [1] : [1, 2]), [isHSGQ])
  const availablePons = useMemo(
    () => (isHSGQ ? [1, 2, 3, 4] : Array.from({ length: 16 }, (_, i) => i + 1)),
    [isHSGQ]
  )

  const loadPonONUs = useCallback(async (oltId: number, board: number, pon: number) => {
    setLoadingPonData(true)
    try {
      const res = await apiClient.get(`/olt/${oltId}/board/${board}/pon/${pon}/onus`)
      setOnusList(res.data?.data || [])
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } }
      notify(e.response?.data?.message || 'Gagal memuat data ONU dari OLT', 'error')
      setOnusList([])
    } finally {
      setLoadingPonData(false)
    }
  }, [])

  const loadUplinks = useCallback(async (oltId: number) => {
    setLoadingUplinks(true)
    try {
      const res = await apiClient.get(`/olt/${oltId}/uplinks`)
      setUplinksData(res.data?.data || null)
    } catch {
      setUplinksData(null)
    } finally {
      setLoadingUplinks(false)
    }
  }, [])

  const openExplorer = (olt: OLT) => {
    setActiveOlt(olt)
    setExplorerTab('pon')
    setSelectedBoard(1)
    setSelectedPon(1)
    setOnuSearch('')
    setCurrentPage(1)
    setShowExplorer(true)
    loadPonONUs(olt.id, 1, 1)
    loadUplinks(olt.id)
  }

  const refreshPonData = async () => {
    if (!activeOlt) return
    setLoadingPonData(true)
    try {
      await apiClient.delete(`/olt/${activeOlt.id}/board/${selectedBoard}/pon/${selectedPon}/cache`)
      await loadPonONUs(activeOlt.id, selectedBoard, selectedPon)
      notify('Cache dibersihkan & data PON diperbarui dari OLT', 'success')
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } }
      notify(e.response?.data?.message || 'Gagal merefresh data PON', 'error')
    } finally {
      setLoadingPonData(false)
    }
  }

  const openEmptyIdsDialog = async () => {
    if (!activeOlt) return
    setLoadingEmptyIds(true)
    try {
      const res = await apiClient.get(`/olt/${activeOlt.id}/board/${selectedBoard}/pon/${selectedPon}/empty-onu-ids`)
      setEmptyOnuIds(res.data?.data || [])
      setShowEmptyIds(true)
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } }
      notify(e.response?.data?.message || 'Gagal mengambil daftar slot kosong', 'error')
    } finally {
      setLoadingEmptyIds(false)
    }
  }

  const openOnuDetail = async (onu: ZTEONUInfo) => {
    if (!activeOlt) return
    setSelectedOnuDetail(null)
    setShowOnuDetail(true)
    setLoadingOnuDetail(true)
    try {
      const res = await apiClient.get(`/olt/${activeOlt.id}/board/${selectedBoard}/pon/${selectedPon}/onu/${onu.onu_id}`)
      const detail: ZTEONUDetail | null = res.data?.data || null
      setSelectedOnuDetail(detail)
      if (detail?.rx_power) {
        setOnusList((prev) =>
          prev.map((o) => (o.onu_id === onu.onu_id ? { ...o, rx_power: detail.rx_power } : o))
        )
      }
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } } }
      notify(e.response?.data?.message || 'Gagal mengambil detail telemetri ONU', 'error')
      setShowOnuDetail(false)
    } finally {
      setLoadingOnuDetail(false)
    }
  }

  const filteredOnus = useMemo(() => {
    if (!onuSearch.trim()) return onusList
    const q = onuSearch.trim().toLowerCase()
    return onusList.filter(
      (o) =>
        o.onu_id.toString().includes(q) ||
        (o.name || '').toLowerCase().includes(q) ||
        (o.customer_name || '').toLowerCase().includes(q) ||
        (o.serial_number || '').toLowerCase().includes(q) ||
        (o.status || '').toLowerCase().includes(q)
    )
  }, [onusList, onuSearch])

  const totalPages = Math.max(1, Math.ceil(filteredOnus.length / itemsPerPage))
  const paginatedOnus = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return filteredOnus.slice(start, start + itemsPerPage)
  }, [filteredOnus, currentPage, itemsPerPage])

  const countOnline = useMemo(() => filteredOnus.filter((o) => o.status?.toLowerCase() === 'online').length, [filteredOnus])
  const countOffline = useMemo(() => filteredOnus.filter((o) => o.status?.toLowerCase() !== 'online').length, [filteredOnus])

  const parsedDiagnostics = useMemo(() => {
    if (!selectedOnuDetail?.description) return null
    const desc = selectedOnuDetail.description
    const tempMatch = desc.match(/Suhu:\s*([^\s|]+)/i)
    const voltMatch = desc.match(/Voltase:\s*([^\s|]+)/i)
    const biasMatch = desc.match(/Arus Bias:\s*([^\s|]+)/i)
    if (tempMatch || voltMatch || biasMatch) {
      return {
        temperature: tempMatch?.[1] || null,
        voltage: voltMatch?.[1] || null,
        biasCurrent: biasMatch?.[1] || null,
      }
    }
    return null
  }, [selectedOnuDetail])

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-primary text-primary-foreground">
            <Router className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">OLT Management</h1>
              <Badge variant="secondary" className="bg-blue-50 text-blue-700 text-[11px]">FTTH Core</Badge>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">
              Monitoring gateway OLT, manajemen port PON, dan diagnostik telemetri ONU real-time
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchOLTs} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah OLT
            </Button>
          )}
        </div>
      </div>

      {/* Feedback Banner */}
      {feedback && (
        <div
          className={`flex items-center justify-between gap-3 px-4 py-3 rounded-xl border text-sm ${
            feedback.type === 'success'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-800'
              : 'bg-red-50 border-red-200 text-red-800'
          }`}
        >
          <div className="flex items-center gap-2">
            {feedback.type === 'success' ? (
              <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 text-red-600 shrink-0" />
            )}
            <span className="font-medium">{feedback.text}</span>
          </div>
          <button type="button" onClick={() => setFeedback(null)} className="opacity-70 hover:opacity-100">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* OLT Grid */}
      {loading && olts.length === 0 ? (
        <div className="flex items-center justify-center py-20 text-sm text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin text-blue-600 mr-2" />
          Memuat gateway OLT...
        </div>
      ) : olts.length === 0 ? (
        <Card className="shadow-xs">
          <CardContent className="py-16 text-center">
            <Server className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
            <h3 className="text-base font-semibold text-foreground/90">Belum Ada Gateway OLT</h3>
            <p className="text-xs text-muted-foreground mt-1 mb-4">
              Tambahkan OLT baru untuk mengaktifkan pemantauan PON dan ONU.
            </p>
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah OLT Baru
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {olts.map((item) => {
            const isTesting = testingId === item.id
            const isConnected = connectedIds[item.id]
            const isHsgqType = (item.tipe_olt || '').toLowerCase() === 'hsgq'
            return (
              <Card key={item.id} className="shadow-xs hover:shadow-md transition-shadow border py-0 flex flex-col">
                <CardContent className="p-4 flex-1 flex flex-col gap-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2.5 min-w-0">
                      <div className="h-10 w-10 rounded-lg bg-muted border flex items-center justify-center shrink-0 text-foreground/90">
                        <Router className="h-5 w-5" />
                      </div>
                      <div className="min-w-0">
                        <h3 className="text-sm font-bold text-foreground truncate">{item.nama_olt}</h3>
                        <div className="text-xs font-mono text-muted-foreground truncate">{item.ip_address}</div>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      {canEdit && (
                        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openDialog(item)} title="Edit OLT">
                          <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      )}
                      {canDelete && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 hover:bg-red-50"
                          onClick={() => {
                            setItemToDelete(item)
                            setShowDeleteDialog(true)
                          }}
                          title="Hapus OLT"
                        >
                          <Trash2 className="h-3.5 w-3.5 text-red-500" />
                        </Button>
                      )}
                    </div>
                  </div>

                  <Separator />

                  <div className="space-y-2 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Tipe OLT</span>
                      <Badge
                        variant="secondary"
                        className={isHsgqType ? 'bg-teal-50 text-teal-700' : 'bg-indigo-50 text-indigo-700'}
                      >
                        {item.tipe_olt}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Username</span>
                      <span className="font-mono font-medium text-foreground">{item.username || '-'}</span>
                    </div>
                    <div className="flex items-center justify-between bg-muted/50 border rounded-lg px-2.5 py-1.5">
                      <span className="text-muted-foreground">Mikrotik</span>
                      <span className="font-semibold text-foreground truncate max-w-[140px]">
                        {getMikrotikName(item.mikrotik_server_id)}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 pt-1 mt-auto">
                    <Button
                      size="sm"
                      className="flex-1 h-9 bg-primary hover:bg-primary/90 text-primary-foreground"
                      onClick={() => openExplorer(item)}
                    >
                      <Network className="h-3.5 w-3.5 mr-1.5" />
                      PON Explorer
                    </Button>
                    <Button
                      size="sm"
                      variant={isConnected ? 'default' : 'outline'}
                      className={`h-9 ${isConnected ? 'bg-emerald-600 hover:bg-emerald-700 text-white' : ''}`}
                      disabled={isTesting}
                      onClick={() => testConnection(item)}
                    >
                      {isTesting ? (
                        <Loader2 className="h-3.5 w-3.5 mr-1 animate-spin" />
                      ) : isConnected ? (
                        <CheckCircle2 className="h-3.5 w-3.5 mr-1" />
                      ) : (
                        <PlugZap className="h-3.5 w-3.5 mr-1" />
                      )}
                      {isConnected ? 'Terhubung' : 'Test'}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Add / Edit OLT Dialog */}
      <Dialog open={showFormDialog} onOpenChange={setShowFormDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Router className="h-5 w-5 text-blue-600" />
              {isEditMode ? 'Edit Gateway OLT' : 'Tambah Gateway OLT'}
            </DialogTitle>
            <DialogDescription>Konfigurasi parameter OLT & integrasi Mikrotik</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div>
              <Label className="text-sm mb-1 block">Nama OLT *</Label>
              <Input
                value={editedItem.nama_olt || ''}
                onChange={(e) => setEditedItem((p) => ({ ...p, nama_olt: e.target.value }))}
                placeholder="Contoh: OLT-Parama"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">IP Address *</Label>
                <Input
                  value={editedItem.ip_address || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, ip_address: e.target.value }))}
                  placeholder="192.168.100.9"
                  className="font-mono"
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">Tipe Perangkat *</Label>
                <Select
                  value={editedItem.tipe_olt || 'HSGQ'}
                  onValueChange={(v) => setEditedItem((p) => ({ ...p, tipe_olt: v }))}
                >
                  <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {oltTypes.map((t) => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <Separator />

            <div>
              <Label className="text-sm mb-1 block">Mikrotik Server Terkait *</Label>
              <Select
                disabled={loadingMikrotiks}
                value={editedItem.mikrotik_server_id ? String(editedItem.mikrotik_server_id) : ''}
                onValueChange={(v) => setEditedItem((p) => ({ ...p, mikrotik_server_id: Number(v) }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={loadingMikrotiks ? 'Loading...' : 'Pilih server Mikrotik...'} />
                </SelectTrigger>
                <SelectContent>
                  {mikrotikList.map((m) => <SelectItem key={m.id} value={String(m.id)}>{m.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Username</Label>
                <Input
                  value={editedItem.username || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, username: e.target.value }))}
                  placeholder="root"
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">
                  Password {!isEditMode && '*'}
                </Label>
                <Input
                  type="password"
                  value={editedItem.password || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, password: e.target.value }))}
                  placeholder={isEditMode ? 'Kosongkan jika tidak diubah' : 'Password OLT'}
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowFormDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveOLT} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              {isEditMode ? 'Simpan Perubahan' : 'Tambah OLT'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">Hapus Gateway OLT</DialogTitle>
            <DialogDescription>
              Apakah Anda yakin ingin menghapus <strong>{itemToDelete?.nama_olt}</strong> ({itemToDelete?.ip_address})?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>Batal</Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Ya, Hapus
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* PON Explorer & Telemetry Dialog */}
      <Dialog open={showExplorer} onOpenChange={setShowExplorer}>
        <DialogContent className="w-[95vw] sm:max-w-6xl max-h-[90vh] flex flex-col p-0 overflow-hidden">
          <div className="px-5 py-4 border-b bg-card flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-muted text-foreground/90">
                <Network className="h-5 w-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-base font-bold text-foreground">{activeOlt?.nama_olt}</h2>
                  <Badge variant="secondary" className="text-xs">{activeOlt?.tipe_olt}</Badge>
                </div>
                <p className="text-xs text-muted-foreground font-mono">
                  IP: {activeOlt?.ip_address} • {isHSGQ ? 'HSGQ EPON Telnet Gateway' : 'ZTE C320 Gateway'}
                </p>
              </div>
            </div>
          </div>

          <Tabs value={explorerTab} onValueChange={(v) => setExplorerTab(v as 'pon' | 'uplinks')} className="flex-1 flex flex-col overflow-hidden">
            <div className="px-5 pt-2 bg-muted/50 border-b">
              <TabsList>
                <TabsTrigger value="pon">PON & Telemetri ONU</TabsTrigger>
                <TabsTrigger value="uplinks">Kartu & Uplink Port</TabsTrigger>
              </TabsList>
            </div>

            <div className="flex-1 overflow-y-auto p-4 md:p-5 bg-muted/30">
              <TabsContent value="pon" className="mt-0 space-y-4">
                {/* Filter Bar */}
                <div className="bg-card border rounded-xl p-3 flex flex-wrap items-center gap-3">
                  {!isHSGQ && (
                    <Select
                      value={String(selectedBoard)}
                      onValueChange={(v) => {
                        const b = Number(v)
                        setSelectedBoard(b)
                        setCurrentPage(1)
                        if (activeOlt) loadPonONUs(activeOlt.id, b, selectedPon)
                      }}
                    >
                      <SelectTrigger className="h-9 w-[130px]"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {availableBoards.map((b) => <SelectItem key={b} value={String(b)}>Board {b}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  )}
                  <Select
                    value={String(selectedPon)}
                    onValueChange={(v) => {
                      const p = Number(v)
                      setSelectedPon(p)
                      setCurrentPage(1)
                      if (activeOlt) loadPonONUs(activeOlt.id, selectedBoard, p)
                    }}
                  >
                    <SelectTrigger className="h-9 w-[130px]"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {availablePons.map((p) => <SelectItem key={p} value={String(p)}>PON {p}</SelectItem>)}
                    </SelectContent>
                  </Select>

                  <div className="relative flex-1 min-w-[200px]">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      value={onuSearch}
                      onChange={(e) => { setOnuSearch(e.target.value); setCurrentPage(1) }}
                      placeholder="Cari ONU ID, Nama, atau MAC..."
                      className="pl-9 h-9"
                    />
                  </div>

                  <div className="flex items-center gap-2 ml-auto">
                    <Button variant="outline" size="sm" onClick={openEmptyIdsDialog} disabled={loadingEmptyIds}>
                      <Hash className="h-3.5 w-3.5 mr-1" />
                      {loadingEmptyIds ? 'Loading...' : 'Cek Slot Kosong'}
                    </Button>
                    <Button variant="outline" size="sm" onClick={refreshPonData} disabled={loadingPonData}>
                      <RefreshCw className={`h-3.5 w-3.5 mr-1 ${loadingPonData ? 'animate-spin' : ''}`} />
                      Refresh Data
                    </Button>
                  </div>
                </div>

                {/* Counters */}
                <div className="flex flex-wrap items-center gap-2 text-xs">
                  <Badge variant="outline" className="bg-card">Total ONU: {filteredOnus.length}</Badge>
                  <Badge className="bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-50">
                    Online: {countOnline}
                  </Badge>
                  <Badge className="bg-rose-50 text-rose-700 border border-rose-200 hover:bg-rose-50">
                    Offline: {countOffline}
                  </Badge>
                </div>

                {/* ONUs Table */}
                <div className="bg-card border rounded-xl overflow-hidden">
                  {loadingPonData ? (
                    <div className="py-16 text-center text-xs text-muted-foreground">
                      <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
                      Mengambil data ONU dari OLT...
                    </div>
                  ) : (
                    <div className="overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow className="bg-muted/50">
                            <TableHead className="text-xs w-[80px]">ONU ID</TableHead>
                            <TableHead className="text-xs">Nama / Deskripsi</TableHead>
                            <TableHead className="text-xs">Tipe Modem</TableHead>
                            <TableHead className="text-xs">Serial Number / MAC</TableHead>
                            <TableHead className="text-xs text-center">Rx Power</TableHead>
                            <TableHead className="text-xs text-center">Status</TableHead>
                            <TableHead className="text-xs text-center">Detail</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {paginatedOnus.length === 0 ? (
                            <TableRow>
                              <TableCell colSpan={7} className="text-center py-12 text-xs text-muted-foreground">
                                Tidak ada ONU ditemukan pada Board {selectedBoard} PON {selectedPon}
                              </TableCell>
                            </TableRow>
                          ) : (
                            paginatedOnus.map((onu) => (
                              <TableRow key={onu.onu_id}>
                                <TableCell className="font-mono text-xs font-bold text-foreground/90">#{onu.onu_id}</TableCell>
                                <TableCell>
                                  <div className="text-sm font-medium text-foreground">{onu.name || '-'}</div>
                                  {onu.customer_name && <div className="text-xs text-muted-foreground">{onu.customer_name}</div>}
                                </TableCell>
                                <TableCell>
                                  {onu.onu_type ? <Badge variant="secondary" className="text-xs">{onu.onu_type}</Badge> : '-'}
                                </TableCell>
                                <TableCell>
                                  <div className="flex items-center gap-1.5 font-mono text-xs">
                                    <span>{onu.serial_number}</span>
                                    <button
                                      type="button"
                                      onClick={() => {
                                        navigator.clipboard.writeText(onu.serial_number)
                                        notify(`SN/MAC disalin: ${onu.serial_number}`, 'success')
                                      }}
                                      className="text-muted-foreground hover:text-foreground/90"
                                    >
                                      <Copy className="h-3.5 w-3.5" />
                                    </button>
                                  </div>
                                </TableCell>
                                <TableCell className="text-center">
                                  {onu.rx_power ? (
                                    <span className={`inline-flex items-center px-2 py-0.5 rounded-md font-mono text-xs font-semibold border ${getRxBadgeClass(onu.rx_power)}`}>
                                      {formatDbm(onu.rx_power)}
                                    </span>
                                  ) : (
                                    <Button variant="ghost" size="sm" className="h-7 text-xs text-blue-600" onClick={() => openOnuDetail(onu)}>
                                      <Activity className="h-3 w-3 mr-1" /> Cek Redaman
                                    </Button>
                                  )}
                                </TableCell>
                                <TableCell className="text-center">
                                  <Badge
                                    className={
                                      onu.status?.toLowerCase() === 'online'
                                        ? 'bg-emerald-100 text-emerald-800 hover:bg-emerald-100'
                                        : 'bg-red-100 text-red-800 hover:bg-red-100'
                                    }
                                  >
                                    {onu.status || 'Offline'}
                                  </Badge>
                                </TableCell>
                                <TableCell className="text-center">
                                  <Button variant="ghost" size="sm" className="h-7 text-xs text-blue-600" onClick={() => openOnuDetail(onu)}>
                                    Diagnosa
                                  </Button>
                                </TableCell>
                              </TableRow>
                            ))
                          )}
                        </TableBody>
                      </Table>
                    </div>
                  )}

                  {/* Pagination */}
                  <div className="flex items-center justify-between px-4 py-2.5 border-t bg-card text-xs text-muted-foreground">
                    <div className="flex items-center gap-2">
                      <span>Per halaman:</span>
                      <Select value={String(itemsPerPage)} onValueChange={(v) => { setItemsPerPage(Number(v)); setCurrentPage(1) }}>
                        <SelectTrigger className="h-7 w-[65px] text-xs"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {[10, 25, 50, 100].map((n) => <SelectItem key={n} value={String(n)}>{n}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="flex items-center gap-2">
                      <span>Halaman {currentPage} / {totalPages}</span>
                      <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => p - 1)}>
                        <ChevronLeft className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage >= totalPages} onClick={() => setCurrentPage((p) => p + 1)}>
                        <ChevronRight className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="uplinks" className="mt-0">
                {loadingUplinks ? (
                  <div className="py-16 text-center text-xs text-muted-foreground">
                    <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
                    Memuat kartu & uplink port...
                  </div>
                ) : uplinksData ? (
                  <div className="grid grid-cols-1 md:grid-cols-12 gap-4">
                    <Card className="md:col-span-5 shadow-xs">
                      <CardContent className="p-4">
                        <h3 className="text-sm font-bold text-foreground mb-3 flex items-center gap-2">
                          <Cpu className="h-4 w-4 text-blue-600" /> Kartu Terpasang (Line Cards)
                        </h3>
                        <div className="divide-y">
                          {uplinksData.cards.map((c, idx) => (
                            <div key={idx} className="py-2.5 flex items-center justify-between text-xs">
                              <div>
                                <div className="font-semibold text-foreground">{c.type}</div>
                                <div className="text-muted-foreground">Role: {c.role}</div>
                              </div>
                              <Badge variant="outline">Slot {c.slot}</Badge>
                            </div>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                    <Card className="md:col-span-7 shadow-xs">
                      <CardContent className="p-4">
                        <h3 className="text-sm font-bold text-foreground mb-3 flex items-center gap-2">
                          <Network className="h-4 w-4 text-blue-600" /> Port Uplink Ethernet
                        </h3>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead className="text-xs">Port</TableHead>
                              <TableHead className="text-xs">Kind</TableHead>
                              <TableHead className="text-xs text-center">Admin</TableHead>
                              <TableHead className="text-xs text-center">Link</TableHead>
                              <TableHead className="text-xs text-right">Speed</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {uplinksData.ports.map((port, idx) => (
                              <TableRow key={idx}>
                                <TableCell className="font-mono text-xs font-bold">{port.name}</TableCell>
                                <TableCell><Badge variant="secondary" className="text-[10px]">{port.kind}</Badge></TableCell>
                                <TableCell className="text-center">
                                  <Badge className={port.admin_status === 'up' ? 'bg-emerald-100 text-emerald-800' : 'bg-muted text-muted-foreground'}>
                                    {port.admin_status}
                                  </Badge>
                                </TableCell>
                                <TableCell className="text-center">
                                  <Badge className={port.oper_status === 'up' ? 'bg-emerald-100 text-emerald-800' : 'bg-red-100 text-red-800'}>
                                    {port.oper_status}
                                  </Badge>
                                </TableCell>
                                <TableCell className="text-right font-mono text-xs">{port.speed_mbps} Mbps</TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </CardContent>
                    </Card>
                  </div>
                ) : (
                  <div className="py-12 text-center text-xs text-muted-foreground">Data uplink tidak tersedia</div>
                )}
              </TabsContent>
            </div>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* ONU Telemetry Detail Dialog */}
      <Dialog open={showOnuDetail} onOpenChange={setShowOnuDetail}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{selectedOnuDetail?.name || 'Detail Telemetri ONU'}</DialogTitle>
            <DialogDescription className="font-mono text-xs">
              {selectedOnuDetail?.customer_name ? `${selectedOnuDetail.customer_name} • ` : ''}
              SN/MAC: {selectedOnuDetail?.serial_number || '-'}
            </DialogDescription>
          </DialogHeader>
          {loadingOnuDetail ? (
            <div className="py-12 text-center text-xs text-muted-foreground">
              <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
              Mengambil telemetri real-time dari OLT...
            </div>
          ) : selectedOnuDetail ? (
            <div className="space-y-3 py-1">
              <div className={`border rounded-xl p-4 text-center ${getRxBadgeClass(selectedOnuDetail.rx_power)}`}>
                <div className="text-[11px] font-bold uppercase tracking-wider mb-1">
                  Optical Rx Power • {getRxQualityLabel(selectedOnuDetail.rx_power)}
                </div>
                <div className="text-3xl font-mono font-bold">{formatDbm(selectedOnuDetail.rx_power)}</div>
                <div className="text-xs mt-1">Tx Power: <strong>{formatDbm(selectedOnuDetail.tx_power)}</strong></div>
              </div>

              {parsedDiagnostics && (
                <div className="grid grid-cols-3 gap-2 text-center text-xs">
                  <div className="border rounded-lg p-2.5 bg-muted/50">
                    <div className="text-muted-foreground">Suhu Modul</div>
                    <div className="font-mono font-bold mt-0.5">{parsedDiagnostics.temperature || '-'}</div>
                  </div>
                  <div className="border rounded-lg p-2.5 bg-muted/50">
                    <div className="text-muted-foreground">Tegangan</div>
                    <div className="font-mono font-bold mt-0.5">{parsedDiagnostics.voltage || '-'}</div>
                  </div>
                  <div className="border rounded-lg p-2.5 bg-muted/50">
                    <div className="text-muted-foreground">Arus Bias</div>
                    <div className="font-mono font-bold mt-0.5">{parsedDiagnostics.biasCurrent || '-'}</div>
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-2.5 text-xs">
                <div className="border rounded-lg p-2.5">
                  <div className="text-muted-foreground">Jarak FO</div>
                  <div className="font-mono font-semibold mt-0.5">
                    {selectedOnuDetail.gpon_optical_distance ? `${selectedOnuDetail.gpon_optical_distance} Meter` : '-'}
                  </div>
                </div>
                <div className="border rounded-lg p-2.5">
                  <div className="text-muted-foreground">Status Modem</div>
                  <div className={`font-semibold mt-0.5 ${selectedOnuDetail.status?.toLowerCase() === 'online' ? 'text-emerald-600' : 'text-red-600'}`}>
                    {selectedOnuDetail.status || '-'}
                  </div>
                </div>
                <div className="border rounded-lg p-2.5">
                  <div className="text-muted-foreground">Tipe Perangkat</div>
                  <div className="font-semibold mt-0.5">{selectedOnuDetail.onu_type || 'EPON'}</div>
                </div>
                <div className="border rounded-lg p-2.5">
                  <div className="text-muted-foreground">IP Management</div>
                  <div className="font-mono font-semibold mt-0.5">{selectedOnuDetail.ip_address || '-'}</div>
                </div>
              </div>

              <div className="border rounded-lg p-2.5 text-xs">
                <div className="text-muted-foreground">Uptime Terakhir</div>
                <div className="font-mono font-medium mt-0.5">{selectedOnuDetail.uptime || '-'}</div>
              </div>
              <div className="border rounded-lg p-2.5 text-xs">
                <div className="text-muted-foreground">Alasan Offline / Diagnosa</div>
                <div className="font-semibold mt-0.5">{selectedOnuDetail.offline_reason || 'Normal / Tidak Ada Error'}</div>
              </div>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>

      {/* Empty ONU IDs Dialog */}
      <Dialog open={showEmptyIds} onOpenChange={setShowEmptyIds}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Slot ONU ID Bebas (Kosong)</DialogTitle>
            <DialogDescription>Board {selectedBoard} • PON {selectedPon}</DialogDescription>
          </DialogHeader>
          {emptyOnuIds.length > 0 ? (
            <div className="flex flex-wrap gap-1.5 max-h-56 overflow-y-auto p-3 border rounded-lg bg-muted/50">
              {emptyOnuIds.map((id) => (
                <Badge key={id} variant="outline" className="font-mono bg-card">#{id}</Badge>
              ))}
            </div>
          ) : (
            <div className="text-center py-6 text-xs text-muted-foreground">Semua slot ONU ID pada port ini sudah penuh.</div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEmptyIds(false)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
