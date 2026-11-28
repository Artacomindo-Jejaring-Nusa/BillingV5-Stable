import { Fragment, useCallback, useEffect, useState } from 'react'
import {
  Activity,
  ChevronDown,
  ChevronUp,
  Download,
  Edit,
  Eye,
  EyeOff,
  Filter,
  Loader2,
  Network,
  Plus,
  RefreshCw,
  Search,
  Server,
  Trash2,
  Upload,
  X,
  Zap,
} from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

export interface DataTeknisItem {
  id: number
  pelanggan_id: number
  id_vlan: string
  id_pelanggan: string
  password_pppoe: string
  ip_pelanggan: string
  profile_pppoe: string
  olt: string
  olt_custom?: string | null
  pon: number
  otb: number
  odc: number
  odp_id?: number | null
  odp: number
  port_odp?: number | null
  speedtest_proof?: string | null
  onu_power: number
  mikrotik_server_id: number
  sn?: string | null
  tipe_ont?: string | null
  onu_id?: number | null
  pelanggan?: {
    id: number
    nama: string
    alamat?: string
  }
}

interface SelectOption {
  id: number
  nama: string
}

const MIKROTIK_VLAN_INFO = [
  { name: 'Tambun', vlan: '100' },
  { name: 'Nagrak', vlan: '105' },
  { name: 'Parama', vlan: '10' },
  { name: 'Waringin', vlan: '10' },
  { name: 'Casanova', vlan: '10' },
  { name: 'Pinus', vlan: '101/100' },
  { name: 'Pulogebang', vlan: '103' },
  { name: 'Tipar Cakung', vlan: '102' },
]

const emptyForm = (): Partial<DataTeknisItem> => ({
  pelanggan_id: undefined,
  id_vlan: '',
  id_pelanggan: '',
  password_pppoe: '',
  ip_pelanggan: '',
  profile_pppoe: '',
  olt: '',
  olt_custom: '',
  pon: 0,
  otb: 0,
  odc: 0,
  odp: 0,
  odp_id: null,
  port_odp: null,
  onu_power: 0,
  mikrotik_server_id: undefined,
  sn: '',
  tipe_ont: '',
  onu_id: null,
})

export function DataTeknisView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_data_teknis')
  const canEdit = hasPermission('edit_data_teknis')
  const canDelete = hasPermission('delete_data_teknis')
  const [list, setList] = useState<DataTeknisItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [initialLoading, setInitialLoading] = useState(true)
  const [fetching, setFetching] = useState(false)
  const [saving, setSaving] = useState(false)

  // Dropdowns
  const [oltOptions, setOltOptions] = useState<string[]>([])
  const [profileOptions, setProfileOptions] = useState<string[]>([])
  const [vlanOptions, setVlanOptions] = useState<string[]>([])
  const [mikrotikServers, setMikrotikServers] = useState<SelectOption[]>([])
  const [odpList, setOdpList] = useState<Array<{ id: number; nama_odp?: string; kode_odp?: string }>>([])
  const [unconfiguredPelanggan, setUnconfiguredPelanggan] = useState<SelectOption[]>([])

  // Filters & Pagination
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [oltFilter, setOltFilter] = useState('all')
  const [profileFilter, setProfileFilter] = useState('all')
  const [vlanFilter, setVlanFilter] = useState('all')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(15)

  // Expanded rows & live telemetry cache
  const [expandedRows, setExpandedRows] = useState<Record<number, boolean>>({})
  const [liveOnuMap, setLiveOnuMap] = useState<Record<number, Record<string, unknown>>>({})
  const [liveOnuLoading, setLiveOnuLoading] = useState<Record<number, boolean>>({})
  const [syncingOnu, setSyncingOnu] = useState<Record<number, boolean>>({})

  // Bulk Auto-Sync OLT
  const [bulkSyncOpen, setBulkSyncOpen] = useState(false)
  const [bulkSyncOlt, setBulkSyncOlt] = useState('Pulogebang')
  const [bulkSyncLoading, setBulkSyncLoading] = useState(false)
  const [bulkSyncResult, setBulkSyncResult] = useState<{
    total_scanned?: number
    updated_count?: number
    not_found_count?: number
    message?: string
  } | null>(null)

  // Scan Detected ONUs
  const [detectedOnuOpen, setDetectedOnuOpen] = useState(false)
  const [pickOlt, setPickOlt] = useState('Tipar Cakung')
  const [pickBoard, setPickBoard] = useState(1)
  const [pickPon, setPickPon] = useState(1)
  const [detectedOnusList, setDetectedOnusList] = useState<Array<{
    serial_number?: string
    name?: string
    onu_type?: string
    rx_power?: string | number
    onu_id?: number
    status?: string
  }>>([])
  const [loadingDetectedOnus, setLoadingDetectedOnus] = useState(false)
  const [searchDetectedOnu, setSearchDetectedOnu] = useState('')

  // 3-Step Dialog
  const [formOpen, setFormOpen] = useState(false)
  const [formStep, setFormStep] = useState<1 | 2 | 3>(1)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<Partial<DataTeknisItem>>(emptyForm())
  const [showPassword, setShowPassword] = useState(false)
  const [speedtestFile, setSpeedtestFile] = useState<File | null>(null)
  const [formError, setFormError] = useState('')
  const [ipCheckMsg, setIpCheckMsg] = useState('')

  const [deleteTarget, setDeleteTarget] = useState<DataTeknisItem | null>(null)
  const [importOpen, setImportOpen] = useState(false)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importing, setImporting] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    const t = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(1)
    }, 500)
    return () => clearTimeout(t)
  }, [searchQuery])

  const fetchDropdowns = useCallback(async () => {
    const [oltRes, profRes, vlanRes, mkRes, odpRes] = await Promise.allSettled([
      apiClient.get('/data_teknis/available-olt'),
      apiClient.get('/data_teknis/available-profiles'),
      apiClient.get('/data_teknis/available-vlans'),
      apiClient.get('/mikrotik_servers'),
      apiClient.get('/odp'),
    ])
    if (oltRes.status === 'fulfilled' && Array.isArray(oltRes.value.data)) setOltOptions(oltRes.value.data)
    if (profRes.status === 'fulfilled' && Array.isArray(profRes.value.data)) setProfileOptions(profRes.value.data)
    if (vlanRes.status === 'fulfilled' && Array.isArray(vlanRes.value.data)) setVlanOptions(vlanRes.value.data)
    if (mkRes.status === 'fulfilled') {
      const items = Array.isArray(mkRes.value.data) ? mkRes.value.data : mkRes.value.data?.data || []
      setMikrotikServers(
        items.map((m: { id: number; name?: string; nama?: string }) => ({
          id: m.id,
          nama: m.name || m.nama || `Server #${m.id}`,
        }))
      )
    }
    if (odpRes.status === 'fulfilled') {
      const items = Array.isArray(odpRes.value.data) ? odpRes.value.data : odpRes.value.data?.data || []
      setOdpList(items)
    }
  }, [])

  const fetchDataTeknis = useCallback(async () => {
    setFetching(true)
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (oltFilter !== 'all') params.append('olt', oltFilter)
      if (profileFilter !== 'all') params.append('profile', profileFilter)
      if (vlanFilter !== 'all') params.append('vlan', vlanFilter)
      params.append('skip', String((page - 1) * limit))
      params.append('limit', String(limit))

      const res = await apiClient.get(`/data_teknis?${params.toString()}`)
      const payload = res.data
      if (payload?.data && payload.total_count !== undefined) {
        setList(payload.data)
        setTotalCount(payload.total_count)
      } else if (Array.isArray(payload)) {
        setList(payload)
        setTotalCount(payload.length)
      }
    } catch {
      setFeedback({ type: 'error', text: 'Gagal memuat data teknis.' })
    } finally {
      setInitialLoading(false)
      setFetching(false)
    }
  }, [debouncedSearch, oltFilter, profileFilter, vlanFilter, page, limit])

  useEffect(() => {
    void fetchDropdowns()
  }, [fetchDropdowns])

  useEffect(() => {
    void fetchDataTeknis()
  }, [fetchDataTeknis])

  const openCreate = async () => {
    setEditingId(null)
    setForm(emptyForm())
    setFormStep(1)
    setSpeedtestFile(null)
    setShowPassword(false)
    setFormError('')
    setIpCheckMsg('')
    setFormOpen(true)
    try {
      const res = await apiClient.get('/data_teknis/unconfigured-pelanggan')
      const items = Array.isArray(res.data) ? res.data : res.data?.data || []
      setUnconfiguredPelanggan(items)
    } catch {
      setUnconfiguredPelanggan([])
    }
  }

  const openEdit = (item: DataTeknisItem) => {
    setEditingId(item.id)
    setForm({ ...item })
    setFormStep(1)
    setSpeedtestFile(null)
    setShowPassword(false)
    setFormError('')
    setIpCheckMsg('')
    setFormOpen(true)
  }

  const handleServerChange = async (serverIdStr: string) => {
    const serverId = Number(serverIdStr)
    setForm((prev) => ({ ...prev, mikrotik_server_id: serverId }))
    if (!editingId && serverId) {
      try {
        const res = await apiClient.get(`/data_teknis/last-ip/${serverId}`)
        if (res.data?.next_ip) {
          setForm((prev) => ({ ...prev, ip_pelanggan: res.data.next_ip }))
        }
      } catch {
        // ignore
      }
    }
  }

  const handleCheckIp = async () => {
    if (!form.ip_pelanggan || !form.mikrotik_server_id) return
    try {
      const res = await apiClient.post('/data_teknis/check-ip', {
        ip_pelanggan: form.ip_pelanggan,
        mikrotik_server_id: form.mikrotik_server_id,
        exclude_id: editingId,
      })
      setIpCheckMsg(res.data?.available ? 'IP tersedia' : res.data?.message || 'IP sudah digunakan')
    } catch {
      setIpCheckMsg('Gagal memeriksa IP')
    }
  }

  const validateStep1 = (): string | null => {
    if (!editingId && !form.pelanggan_id) return 'Pilih pelanggan terlebih dahulu.'
    if (!form.mikrotik_server_id) return 'Mikrotik Server wajib dipilih.'
    if (!form.id_pelanggan?.trim() || /\s/.test(form.id_pelanggan)) {
      return 'ID Pelanggan (PPPoE) wajib diisi dan tidak boleh mengandung spasi.'
    }
    if (!form.password_pppoe?.trim() || /\s/.test(form.password_pppoe)) {
      return 'Password PPPoE wajib diisi dan tidak boleh mengandung spasi.'
    }
    if (!form.ip_pelanggan?.trim()) return 'IP Pelanggan wajib diisi.'
    if (!form.profile_pppoe?.trim()) return 'Profile PPPoE wajib diisi.'
    if (!form.id_vlan?.trim()) return 'ID VLAN wajib diisi.'
    return null
  }

  const handleNextStep = () => {
    if (formStep === 1) {
      const err = validateStep1()
      if (err) {
        setFormError(err)
        return
      }
      setFormError('')
      setFormStep(2)
    } else if (formStep === 2) {
      setFormError('')
      setFormStep(3)
    }
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (formStep !== 3) {
      handleNextStep()
      return
    }
    setSaving(true)
    setFormError('')
    try {
      let speedtestProofUrl = form.speedtest_proof || null
      if (speedtestFile) {
        const fd = new FormData()
        fd.append('file', speedtestFile)
        const upRes = await apiClient.post('/uploads/speedtest', fd)
        speedtestProofUrl = upRes.data?.url || upRes.data?.path || speedtestProofUrl
      }

      const payload = {
        ...form,
        speedtest_proof: speedtestProofUrl,
      }

      if (editingId) {
        await apiClient.patch(`/data_teknis/${editingId}`, payload)
        await fetchDataTeknis()
        setFeedback({ type: 'success', text: 'Data teknis berhasil diperbarui.' })
      } else {
        await apiClient.post('/data_teknis', payload)
        await fetchDataTeknis()
        setFeedback({ type: 'success', text: 'Data teknis berhasil ditambahkan.' })
      }
      setFormOpen(false)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { detail?: string; error?: string } } })?.response?.data?.detail ||
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Gagal menyimpan data teknis.'
      setFormError(msg)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await apiClient.delete(`/data_teknis/${deleteTarget.id}`)
      await fetchDataTeknis()
      setFeedback({ type: 'success', text: 'Data teknis berhasil dihapus.' })
      setDeleteTarget(null)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menghapus data teknis.' })
      setDeleteTarget(null)
    }
  }

  const handleCheckLiveOnu = async (item: DataTeknisItem) => {
    setExpandedRows((prev) => ({ ...prev, [item.id]: true }))
    setLiveOnuLoading((prev) => ({ ...prev, [item.id]: true }))
    try {
      const res = await apiClient.get(`/data_teknis/${item.id}/live-onu`)
      setLiveOnuMap((prev) => ({ ...prev, [item.id]: res.data }))
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengambil status Live ONU dari OLT.' })
    } finally {
      setLiveOnuLoading((prev) => ({ ...prev, [item.id]: false }))
    }
  }

  const handleSyncOnuPower = async (item: DataTeknisItem, powerVal?: unknown) => {
    setSyncingOnu((prev) => ({ ...prev, [item.id]: true }))
    try {
      const parsed = powerVal !== undefined && powerVal !== null && powerVal !== '' ? parseFloat(String(powerVal)) : undefined
      const payload = parsed !== undefined && !Number.isNaN(parsed) ? { onu_power: parsed } : {}
      const res = await apiClient.post(`/data_teknis/${item.id}/sync-onu-power`, payload)
      const updatedVal = res.data?.data?.onu_power ?? parsed ?? item.onu_power
      setList((prev) => prev.map((row) => (row.id === item.id ? { ...row, onu_power: updatedVal } : row)))
      setFeedback({ type: 'success', text: `Redaman berhasil disinkronkan ke Database: ${updatedVal} dBm` })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal sinkronisasi redaman ke Database.' })
    } finally {
      setSyncingOnu((prev) => ({ ...prev, [item.id]: false }))
    }
  }

  const openBulkSyncDialog = () => {
    setBulkSyncOlt(oltFilter !== 'all' ? oltFilter : 'Pulogebang')
    setBulkSyncResult(null)
    setBulkSyncOpen(true)
  }

  const runBulkSync = async () => {
    if (!bulkSyncOlt) return
    setBulkSyncLoading(true)
    setBulkSyncResult(null)
    try {
      const res = await apiClient.post(
        '/data_teknis/bulk-sync-olt',
        { olt: bulkSyncOlt },
        { timeout: 180000 },
      )
      setBulkSyncResult(res.data?.data || { message: res.data?.message || 'Selesai' })
      setFeedback({ type: 'success', text: res.data?.message || `Sinkronisasi massal OLT ${bulkSyncOlt} selesai!` })
      await fetchDataTeknis()
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menjalankan sinkronisasi massal OLT.' })
    } finally {
      setBulkSyncLoading(false)
    }
  }

  const fetchDetectedONUs = async (oltName = pickOlt, board = pickBoard, pon = pickPon) => {
    setLoadingDetectedOnus(true)
    setDetectedOnusList([])
    try {
      const res = await apiClient.get('/data_teknis/detected-onus', {
        params: { olt: oltName, board, pon },
      })
      setDetectedOnusList(Array.isArray(res.data?.data) ? res.data.data : [])
    } catch {
      setFeedback({ type: 'error', text: 'Gagal memindai ONU dari OLT.' })
    } finally {
      setLoadingDetectedOnus(false)
    }
  }

  const openDetectedOnuModal = () => {
    const initialOlt = form.olt || 'Tipar Cakung'
    const initialPon = form.pon && form.pon > 0 ? form.pon : 1
    setPickOlt(initialOlt)
    setPickBoard(1)
    setPickPon(initialPon)
    setSearchDetectedOnu('')
    setDetectedOnuOpen(true)
    void fetchDetectedONUs(initialOlt, 1, initialPon)
  }

  const selectDetectedONU = (onu: {
    serial_number?: string
    name?: string
    onu_type?: string
    rx_power?: string | number
    onu_id?: number
  }) => {
    setForm((prev) => ({
      ...prev,
      sn: onu.serial_number || prev.sn,
      tipe_ont: onu.onu_type || prev.tipe_ont,
      onu_power: onu.rx_power !== undefined && !Number.isNaN(parseFloat(String(onu.rx_power)))
        ? parseFloat(String(onu.rx_power))
        : prev.onu_power,
      onu_id: onu.onu_id || prev.onu_id,
      pon: pickPon || prev.pon,
      olt: pickOlt || prev.olt,
    }))
    setDetectedOnuOpen(false)
    setFeedback({ type: 'success', text: `ONU ${onu.serial_number || onu.name} (PON ${pickPon}) berhasil dipilih.` })
  }

  const handleDownloadTemplate = async () => {
    try {
      const res = await apiClient.get('/data_teknis/template/csv', { responseType: 'blob' })
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = 'template_import_teknis.csv'
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengunduh template CSV.' })
    }
  }

  const handleExport = async (format: 'csv' | 'excel') => {
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (oltFilter !== 'all') params.append('olt', oltFilter)
      params.append('format', format)
      const res = await apiClient.get(`/data_teknis/export?${params.toString()}`, { responseType: 'blob' })
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = `data_teknis.${format === 'excel' ? 'xlsx' : 'csv'}`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengekspor data teknis.' })
    }
  }

  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!importFile) return
    setImporting(true)
    try {
      const fd = new FormData()
      fd.append('file', importFile)
      await apiClient.post('/data_teknis/import/csv', fd)
      setImportOpen(false)
      setImportFile(null)
      await fetchDataTeknis()
      setFeedback({ type: 'success', text: 'Import data teknis berhasil.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengimpor file CSV data teknis.' })
    } finally {
      setImporting(false)
    }
  }

  const getOdpLabel = (item: DataTeknisItem) => {
    if (item.odp_id) {
      const found = odpList.find((o) => o.id === item.odp_id)
      if (found) return found.nama_odp || found.kode_odp || `ODP #${found.id}`
    }
    return item.odp ? `ODP ${item.odp}` : 'N/A'
  }

  const totalPages = Math.max(1, Math.ceil(totalCount / limit))

  return (
    <div className="p-4 md:p-6 space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Data Teknis Pelanggan</h2>
          <p className="text-sm text-muted-foreground">
            Total {totalCount.toLocaleString('id-ID')} konfigurasi teknis PPPoE & OLT
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Button variant="outline" size="sm" onClick={openBulkSyncDialog} disabled={bulkSyncLoading}>
            <Zap className="mr-1.5 h-4 w-4 text-amber-500" />
            Auto-Sync OLT
          </Button>
          {canCreate && (
            <Button variant="outline" size="sm" onClick={() => setImportOpen(true)}>
              <Upload className="mr-1.5 h-4 w-4" />
              Import
            </Button>
          )}
          <Button variant="outline" size="sm" onClick={() => void handleExport('csv')}>
            <Download className="mr-1.5 h-4 w-4" />
            Export CSV
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => void openCreate()}>
              <Plus className="mr-1.5 h-4 w-4" />
              Tambah Data
            </Button>
          )}
        </div>
      </div>

      {feedback && (
        <div
          className={`flex items-center justify-between rounded-md border px-4 py-2.5 text-sm ${
            feedback.type === 'success'
              ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
              : 'border-destructive/40 bg-destructive/10 text-destructive'
          }`}
        >
          <span>{feedback.text}</span>
          <button type="button" onClick={() => setFeedback(null)} aria-label="Tutup">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      <Card>
        <CardContent className="pt-5 space-y-3">
          <div className="flex flex-col sm:flex-row gap-2">
            <div className="relative flex-1">
              <Search className="h-4 w-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari nama, ID PPPoE, atau IP pelanggan..."
                className="pl-9"
              />
            </div>
            <Button variant="outline" onClick={() => setShowFilters((v) => !v)}>
              <Filter className="mr-1.5 h-4 w-4" />
              Filter
            </Button>
            <Button variant="outline" onClick={() => void fetchDataTeknis()} disabled={fetching}>
              <RefreshCw className={`h-4 w-4 ${fetching ? 'animate-spin' : ''}`} />
            </Button>
          </div>

          {showFilters && (
            <div className="grid grid-cols-1 sm:grid-cols-4 gap-3 pt-2 border-t">
              <Select
                value={oltFilter}
                onValueChange={(v) => {
                  setOltFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua OLT" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua OLT</SelectItem>
                  {oltOptions.map((o) => (
                    <SelectItem key={o} value={o}>
                      {o}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={profileFilter}
                onValueChange={(v) => {
                  setProfileFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Profile" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Profile</SelectItem>
                  {profileOptions.map((p) => (
                    <SelectItem key={p} value={p}>
                      {p}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={vlanFilter}
                onValueChange={(v) => {
                  setVlanFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua VLAN" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua VLAN</SelectItem>
                  {vlanOptions.map((v) => (
                    <SelectItem key={v} value={v}>
                      {v}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button
                variant="ghost"
                onClick={() => {
                  setSearchQuery('')
                  setOltFilter('all')
                  setProfileFilter('all')
                  setVlanFilter('all')
                  setPage(1)
                }}
              >
                Reset Filter
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Informasi VLAN Mikrotik */}
      <Card>
        <CardContent className="p-4 space-y-2.5">
          <div className="flex items-center gap-2 text-sm font-semibold">
            <Network className="h-4 w-4 text-primary" />
            <span>Informasi VLAN Mikrotik</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {MIKROTIK_VLAN_INFO.map((info) => (
              <Badge key={info.name} variant="secondary" className="gap-1.5 px-3 py-1 text-xs font-medium">
                <Server className="h-3.5 w-3.5 text-primary" />
                <strong>{info.name}:</strong> VLAN {info.vlan}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Daftar Infrastruktur Table */}
      <Card className="overflow-hidden">
        {fetching && !initialLoading && (
          <div className="h-1 w-full bg-primary/20 overflow-hidden">
            <div className="h-full w-1/3 bg-primary animate-pulse" />
          </div>
        )}
        <CardContent className="p-0">
          <div className="px-4 py-3 border-b font-semibold text-sm flex items-center gap-2">
            <Server className="h-4 w-4 text-primary" />
            <span>Daftar Infrastruktur</span>
          </div>

          {initialLoading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : list.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">
              Tidak ada data teknis yang ditemukan.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="w-14 px-4 py-3 text-xs uppercase tracking-wide">No</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Nama Pelanggan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">PPPoE Secret</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Password PPPoE</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">IP Pelanggan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">OLT</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">ONU Power</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((dt, index) => {
                    const isExpanded = !!expandedRows[dt.id]
                    const liveData = liveOnuMap[dt.id]
                    const isCheckingLive = !!liveOnuLoading[dt.id]
                    const namaPelanggan = dt.pelanggan?.nama || `Pelanggan #${dt.pelanggan_id}`
                    return (
                      <Fragment key={dt.id}>
                        <TableRow>
                          <TableCell className="px-4 py-3.5 text-xs text-muted-foreground">
                            {(page - 1) * limit + index + 1}
                          </TableCell>
                          <TableCell className="px-4 py-3.5 font-medium text-sm whitespace-nowrap">
                            {namaPelanggan}
                          </TableCell>
                          <TableCell className="px-4 py-3.5 font-mono text-xs whitespace-nowrap">
                            {dt.id_pelanggan || '-'}
                          </TableCell>
                          <TableCell className="px-4 py-3.5 font-mono text-xs whitespace-nowrap">
                            {dt.password_pppoe || '-'}
                          </TableCell>
                          <TableCell className="px-4 py-3.5">
                            <Badge variant="secondary" className="font-mono text-xs">
                              {dt.ip_pelanggan || '-'}
                            </Badge>
                          </TableCell>
                          <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">
                            {dt.olt || dt.olt_custom || '-'}
                          </TableCell>
                          <TableCell className="px-4 py-3.5">
                            <div className="flex items-center gap-1.5">
                              <Badge
                                variant={dt.onu_power < -27 ? 'destructive' : 'secondary'}
                                className="font-mono text-xs"
                              >
                                {dt.onu_power ?? 0} dBm
                              </Badge>
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-7 w-7"
                                onClick={() => void handleCheckLiveOnu(dt)}
                                title="Cek Sinyal Live OLT"
                              >
                                <Activity className="h-3.5 w-3.5 text-sky-600" />
                              </Button>
                            </div>
                          </TableCell>
                          <TableCell className="px-4 py-3.5 text-right whitespace-nowrap">
                            <div className="inline-flex items-center gap-1">
                              {canEdit && (
                                <Button size="sm" variant="outline" onClick={() => openEdit(dt)} className="h-8 text-xs">
                                  <Edit className="mr-1 h-3.5 w-3.5" />
                                  Edit
                                </Button>
                              )}
                              {canDelete && (
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="h-8 text-xs text-destructive hover:text-destructive"
                                  onClick={() => setDeleteTarget(dt)}
                                >
                                  <Trash2 className="mr-1 h-3.5 w-3.5" />
                                  Hapus
                                </Button>
                              )}
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8"
                                onClick={() =>
                                  setExpandedRows((prev) => ({ ...prev, [dt.id]: !prev[dt.id] }))
                                }
                                aria-label="Detail infrastruktur"
                              >
                                {isExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>

                        {isExpanded && (
                          <TableRow className="bg-muted/20 hover:bg-muted/20">
                            <TableCell colSpan={8} className="p-4">
                              <div className="rounded-lg border bg-card p-4 space-y-4">
                                <div className="flex flex-wrap items-center justify-between gap-2 pb-2 border-b">
                                  <div className="font-semibold text-sm flex items-center gap-2 text-sky-700 dark:text-sky-400">
                                    <Activity className="h-4 w-4" />
                                    Detail Infrastruktur & Telemetri Live
                                  </div>
                                  <div className="flex items-center gap-2">
                                    <Button
                                      size="sm"
                                      variant="secondary"
                                      className="h-7 text-xs"
                                      onClick={() => void handleCheckLiveOnu(dt)}
                                      disabled={isCheckingLive}
                                    >
                                      {isCheckingLive ? 'Mengecek...' : 'Cek Sinyal Live OLT'}
                                    </Button>
                                    <Badge variant="outline" className="font-mono text-xs">
                                      ID: {dt.id_pelanggan || '-'}
                                    </Badge>
                                  </div>
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-xs">
                                  <div className="space-y-1.5">
                                    <div className="font-semibold text-foreground">Info Jaringan</div>
                                    <div className="text-muted-foreground">Password: {dt.password_pppoe || '-'}</div>
                                    <div className="text-muted-foreground">Profile: {dt.profile_pppoe || '-'}</div>
                                    <div className="text-muted-foreground">VLAN: {dt.id_vlan || '-'}</div>
                                  </div>

                                  <div className="space-y-1.5">
                                    <div className="font-semibold text-foreground">Info Infrastruktur</div>
                                    <div className="text-muted-foreground">OLT: {dt.olt || dt.olt_custom || '-'}</div>
                                    <div className="text-muted-foreground">
                                      PON: {dt.pon ?? 0} | OTB: {dt.otb ?? 0} | ODC: {dt.odc ?? 0}
                                    </div>
                                    <div className="text-muted-foreground">
                                      ODP: {getOdpLabel(dt)} (Port: {dt.port_odp ?? '-'})
                                    </div>
                                    <div className="text-muted-foreground">SN: {dt.sn || 'N/A'}</div>
                                  </div>

                                  <div className="space-y-1.5">
                                    <div className="flex items-center justify-between">
                                      <div className="font-semibold text-foreground">Diagnosa OLT Real-Time</div>
                                      {liveData && (
                                        <Button
                                          size="sm"
                                          variant="outline"
                                          className="h-6 px-2 text-[11px]"
                                          disabled={Boolean(syncingOnu[dt.id])}
                                          onClick={() =>
                                            void handleSyncOnuPower(
                                              dt,
                                              (liveData as Record<string, unknown>).rx_power ??
                                                (liveData as Record<string, unknown>).onu_power,
                                            )
                                          }
                                        >
                                          {syncingOnu[dt.id] ? 'Menyimpan...' : 'Sync ke DB'}
                                        </Button>
                                      )}
                                    </div>
                                    {liveData ? (
                                      <pre className="rounded border bg-muted p-2 text-[11px] overflow-x-auto max-h-28">
                                        {JSON.stringify(liveData, null, 2)}
                                      </pre>
                                    ) : (
                                      <div className="rounded border border-dashed p-3 text-center space-y-2 text-muted-foreground">
                                        <div>Belum ada data telemetri live</div>
                                        <Button
                                          size="sm"
                                          variant="outline"
                                          className="h-7 text-xs"
                                          onClick={() => void handleCheckLiveOnu(dt)}
                                          disabled={isCheckingLive}
                                        >
                                          Cek Sekarang
                                        </Button>
                                      </div>
                                    )}
                                  </div>

                                  <div className="space-y-1.5">
                                    <div className="font-semibold text-foreground">Bukti Speedtest</div>
                                    {dt.speedtest_proof ? (
                                      <a
                                        href={`${apiClient.defaults.baseURL || ''}${dt.speedtest_proof}`}
                                        target="_blank"
                                        rel="noreferrer"
                                        className="text-primary underline"
                                      >
                                        Lihat Gambar Speedtest
                                      </a>
                                    ) : (
                                      <div className="text-muted-foreground">Tidak ada gambar.</div>
                                    )}
                                  </div>
                                </div>
                              </div>
                            </TableCell>
                          </TableRow>
                        )}
                      </Fragment>
                    )
                  })}
                </TableBody>
              </Table>
            </div>
          )}

          {!initialLoading && totalCount > 0 && (
            <div className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3 text-xs text-muted-foreground">
              <div className="flex items-center gap-2">
                <span>
                  Halaman {page} dari {totalPages} ({totalCount.toLocaleString('id-ID')} data)
                </span>
                <Select
                  value={String(limit)}
                  onValueChange={(v) => {
                    setLimit(Number(v))
                    setPage(1)
                  }}
                >
                  <SelectTrigger className="h-8 w-20 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[15, 25, 50, 100].map((n) => (
                      <SelectItem key={n} value={String(n)}>
                        {n}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                  Sebelumnya
                </Button>
                <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                  Berikutnya
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Wide 3-Step Form Dialog */}
      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="sm:max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingId ? 'Edit Data Teknis' : 'Tambah Data Teknis'}</DialogTitle>
            <DialogDescription>Lengkapi informasi teknis jaringan, infrastruktur, dan ONU.</DialogDescription>
          </DialogHeader>

          {/* 3-Step Indicator */}
          <div className="flex items-center gap-2 rounded-lg border bg-muted/30 px-4 py-3 text-xs sm:text-sm">
            {[
              { step: 1 as const, label: 'Jaringan' },
              { step: 2 as const, label: 'Infrastruktur' },
              { step: 3 as const, label: 'ONU' },
            ].map((s, idx) => (
              <Fragment key={s.step}>
                <button
                  type="button"
                  onClick={() => {
                    if (s.step < formStep) setFormStep(s.step)
                  }}
                  className="flex items-center gap-2 font-medium shrink-0"
                >
                  <span
                    className={`h-6 w-6 rounded-full text-xs font-bold flex items-center justify-center ${
                      formStep === s.step
                        ? 'bg-primary text-primary-foreground'
                        : formStep > s.step
                          ? 'bg-emerald-600 text-white'
                          : 'bg-muted text-muted-foreground'
                    }`}
                  >
                    {s.step}
                  </span>
                  <span className={formStep === s.step ? 'text-foreground' : 'text-muted-foreground'}>
                    {s.label}
                  </span>
                </button>
                {idx < 2 && <div className="h-px flex-1 bg-border" />}
              </Fragment>
            ))}
          </div>

          <form onSubmit={handleSave} className="space-y-5">
            {formError && (
              <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {formError}
              </div>
            )}

            {formStep === 1 && (
              <div className="rounded-lg border p-5 space-y-4">
                <h3 className="text-base font-semibold">Informasi Jaringan</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {!editingId && (
                    <div className="space-y-1.5 sm:col-span-2">
                      <Label>
                        Pilih Pelanggan <span className="text-destructive">*</span>
                      </Label>
                      <Select
                        value={form.pelanggan_id ? String(form.pelanggan_id) : ''}
                        onValueChange={(v) => {
                          setFormError('')
                          setForm({ ...form, pelanggan_id: Number(v) })
                        }}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Pilih pelanggan yang belum terkonfigurasi..." />
                        </SelectTrigger>
                        <SelectContent>
                          {unconfiguredPelanggan.map((p) => (
                            <SelectItem key={p.id} value={String(p.id)}>
                              {p.nama}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}

                  <div className="space-y-1.5">
                    <Label>
                      Mikrotik Server <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.mikrotik_server_id ? String(form.mikrotik_server_id) : ''}
                      onValueChange={(v) => {
                        setFormError('')
                        void handleServerChange(v)
                      }}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Pilih Mikrotik Server" />
                      </SelectTrigger>
                      <SelectContent>
                        {mikrotikServers.map((s) => (
                          <SelectItem key={s.id} value={String(s.id)}>
                            {s.nama}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="id_pelanggan">
                      ID Pelanggan (PPPoE) <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="id_pelanggan"
                      value={form.id_pelanggan || ''}
                      onChange={(e) => setForm({ ...form, id_pelanggan: e.target.value })}
                      placeholder="contoh: user123"
                    />
                    <p className="text-[11px] text-muted-foreground">Tidak boleh mengandung spasi.</p>
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="password_pppoe">
                      Password PPPoE <span className="text-destructive">*</span>
                    </Label>
                    <div className="relative">
                      <Input
                        id="password_pppoe"
                        type={showPassword ? 'text' : 'password'}
                        value={form.password_pppoe || ''}
                        onChange={(e) => setForm({ ...form, password_pppoe: e.target.value })}
                        placeholder="contoh: support123.!!"
                        className="pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword((v) => !v)}
                        className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        aria-label="Toggle password"
                      >
                        {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    <p className="text-[11px] text-muted-foreground">Tidak boleh mengandung spasi.</p>
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="ip_pelanggan">
                      IP Pelanggan <span className="text-destructive">*</span>
                    </Label>
                    <div className="flex gap-1.5">
                      <Input
                        id="ip_pelanggan"
                        value={form.ip_pelanggan || ''}
                        onChange={(e) => setForm({ ...form, ip_pelanggan: e.target.value })}
                        placeholder="192.168.x.x"
                      />
                      <Button type="button" variant="outline" size="sm" onClick={() => void handleCheckIp()}>
                        Cek
                      </Button>
                    </div>
                    {ipCheckMsg && <p className="text-[11px] text-muted-foreground">{ipCheckMsg}</p>}
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="profile_pppoe">
                      Profile PPPoE <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="profile_pppoe"
                      value={form.profile_pppoe || ''}
                      onChange={(e) => setForm({ ...form, profile_pppoe: e.target.value })}
                      placeholder="contoh: 10Mbps-j"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="id_vlan">
                      ID VLAN <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="id_vlan"
                      value={form.id_vlan || ''}
                      onChange={(e) => setForm({ ...form, id_vlan: e.target.value })}
                      placeholder="contoh: 10"
                    />
                  </div>
                </div>
              </div>
            )}

            {formStep === 2 && (
              <div className="rounded-lg border p-5 space-y-4">
                <h3 className="text-base font-semibold">Detail Infrastruktur</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <Label>Terhubung ke ODP</Label>
                    <Select
                      value={form.odp_id ? String(form.odp_id) : ''}
                      onValueChange={(v) => setForm({ ...form, odp_id: Number(v) })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Pilih ODP" />
                      </SelectTrigger>
                      <SelectContent>
                        {odpList.map((o) => (
                          <SelectItem key={o.id} value={String(o.id)}>
                            {o.nama_odp || o.kode_odp || `ODP #${o.id}`}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="olt">OLT / OLT Custom</Label>
                    <Input
                      id="olt"
                      value={form.olt || ''}
                      onChange={(e) => setForm({ ...form, olt: e.target.value })}
                      placeholder="contoh: Parama"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="pon">PON</Label>
                    <Input
                      id="pon"
                      type="number"
                      value={form.pon ?? 0}
                      onChange={(e) => setForm({ ...form, pon: Number(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="otb">OTB</Label>
                    <Input
                      id="otb"
                      type="number"
                      value={form.otb ?? 0}
                      onChange={(e) => setForm({ ...form, otb: Number(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="odc">ODC</Label>
                    <Input
                      id="odc"
                      type="number"
                      value={form.odc ?? 0}
                      onChange={(e) => setForm({ ...form, odc: Number(e.target.value) })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="port_odp">Port ODP</Label>
                    <Input
                      id="port_odp"
                      type="number"
                      value={form.port_odp ?? ''}
                      onChange={(e) =>
                        setForm({ ...form, port_odp: e.target.value ? Number(e.target.value) : null })
                      }
                    />
                  </div>
                </div>
              </div>
            )}

            {formStep === 3 && (
              <div className="rounded-lg border p-5 space-y-4">
                <h3 className="text-base font-semibold">Detail ONU</h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="onu_power">ONU Power (dBm)</Label>
                    <Input
                      id="onu_power"
                      type="number"
                      step="0.01"
                      value={form.onu_power ?? 0}
                      onChange={(e) => setForm({ ...form, onu_power: Number(e.target.value) })}
                    />
                    <p className="text-[11px] text-muted-foreground">
                      Range: -40 dBm (sinyal lemah) hingga 10 dBm (sinyal sangat kuat)
                    </p>
                  </div>
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <Label htmlFor="sn">Serial Number (SN) ONU</Label>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="h-6 px-2 text-[11px]"
                        onClick={openDetectedOnuModal}
                      >
                        <Search className="mr-1 h-3 w-3" />
                        Scan ONU dari OLT
                      </Button>
                    </div>
                    <Input
                      id="sn"
                      value={form.sn || ''}
                      onChange={(e) => setForm({ ...form, sn: e.target.value })}
                      placeholder="Contoh: RTEGC69E3F18"
                    />
                  </div>
                  <div className="space-y-1.5 sm:col-span-2">
                    <Label htmlFor="speedtest">Unggah Bukti Speedtest (Opsional)</Label>
                    <Input
                      id="speedtest"
                      type="file"
                      accept="image/*"
                      onChange={(e) => setSpeedtestFile(e.target.files?.[0] || null)}
                    />
                  </div>
                </div>
              </div>
            )}

            <DialogFooter className="flex justify-between sm:justify-between">
              <div>
                {formStep > 1 && (
                  <Button
                    key="btn-back"
                    type="button"
                    variant="outline"
                    onClick={() => setFormStep((s) => (s === 3 ? 2 : 1))}
                  >
                    Kembali
                  </Button>
                )}
              </div>
              <div className="flex gap-2">
                <Button key="btn-cancel" type="button" variant="outline" onClick={() => setFormOpen(false)}>
                  Batal
                </Button>
                {formStep < 3 ? (
                  <Button key="btn-next" type="button" onClick={handleNextStep}>
                    Lanjut
                  </Button>
                ) : (
                  <Button key="btn-save" type="submit" disabled={saving}>
                    {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {saving ? 'Menyimpan...' : 'Simpan'}
                  </Button>
                )}
              </div>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Import CSV Dialog */}
      <Dialog open={importOpen} onOpenChange={setImportOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import Data Teknis (CSV)</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleImport} className="space-y-4">
            <div className="flex items-center justify-between rounded-md border bg-muted/40 px-3 py-2 text-xs">
              <span className="text-muted-foreground">Belum punya format CSV?</span>
              <Button type="button" variant="outline" size="sm" onClick={() => void handleDownloadTemplate()}>
                <Download className="mr-1.5 h-3.5 w-3.5" />
                Download Template
              </Button>
            </div>
            <Input type="file" accept=".csv" onChange={(e) => setImportFile(e.target.files?.[0] || null)} required />
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setImportOpen(false)}>
                Batal
              </Button>
              <Button type="submit" disabled={importing || !importFile}>
                {importing ? 'Mengimpor...' : 'Upload & Import'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Bulk Auto-Sync OLT Dialog */}
      <Dialog open={bulkSyncOpen} onOpenChange={setBulkSyncOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Auto-Sync Redaman ONU (Massal)</DialogTitle>
            <DialogDescription>
              Sinkronisasi nilai redaman (ONU Power) seluruh pelanggan pada OLT terpilih ke database.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label>Target OLT</Label>
              <Select value={bulkSyncOlt} onValueChange={setBulkSyncOlt}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih OLT" />
                </SelectTrigger>
                <SelectContent>
                  {(oltOptions.length > 0 ? oltOptions : ['Pulogebang', 'Tipar Cakung', 'Tambun', 'Nagrak', 'Parama', 'Waringin', 'Casanova', 'Pinus']).map((o) => (
                    <SelectItem key={o} value={o}>
                      {o}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {bulkSyncResult && (
              <div className="rounded-md border bg-muted/40 p-3 text-xs space-y-1">
                <div>Total dipindai: <strong>{bulkSyncResult.total_scanned ?? '-'}</strong></div>
                <div>Berhasil diperbarui: <strong>{bulkSyncResult.updated_count ?? '-'}</strong></div>
                {bulkSyncResult.not_found_count !== undefined && (
                  <div>Tidak ditemukan: <strong>{bulkSyncResult.not_found_count}</strong></div>
                )}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setBulkSyncOpen(false)}>
              Tutup
            </Button>
            <Button type="button" onClick={() => void runBulkSync()} disabled={bulkSyncLoading}>
              {bulkSyncLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {bulkSyncLoading ? 'Menyinkronkan...' : 'Mulai Sync OLT'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Scan Detected ONU Dialog */}
      <Dialog open={detectedOnuOpen} onOpenChange={setDetectedOnuOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Scan ONU Terdeteksi dari OLT</DialogTitle>
            <DialogDescription>Pilih ONU yang terdeteksi untuk mengisi otomatis SN, Tipe ONT, dan Redaman.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div className="grid grid-cols-1 sm:grid-cols-4 gap-2">
              <Select value={pickOlt} onValueChange={setPickOlt}>
                <SelectTrigger>
                  <SelectValue placeholder="OLT" />
                </SelectTrigger>
                <SelectContent>
                  {(oltOptions.length > 0 ? oltOptions : ['Tipar Cakung', 'Pulogebang', 'Tambun', 'Nagrak', 'Parama', 'Waringin', 'Casanova', 'Pinus']).map((o) => (
                    <SelectItem key={o} value={o}>
                      {o}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                type="number"
                min={1}
                value={pickBoard}
                onChange={(e) => setPickBoard(Number(e.target.value) || 1)}
                placeholder="Board"
              />
              <Input
                type="number"
                min={1}
                value={pickPon}
                onChange={(e) => setPickPon(Number(e.target.value) || 1)}
                placeholder="PON"
              />
              <Button type="button" onClick={() => void fetchDetectedONUs()} disabled={loadingDetectedOnus}>
                {loadingDetectedOnus ? 'Memindai...' : 'Scan Ulang'}
              </Button>
            </div>
            <Input
              value={searchDetectedOnu}
              onChange={(e) => setSearchDetectedOnu(e.target.value)}
              placeholder="Filter SN atau nama ONU..."
            />
            <div className="max-h-64 overflow-y-auto divide-y rounded-md border">
              {loadingDetectedOnus ? (
                <div className="p-6 text-center text-xs text-muted-foreground">Memindai ONU dari OLT...</div>
              ) : detectedOnusList.filter((o) => {
                const q = searchDetectedOnu.toLowerCase().trim()
                if (!q) return true
                return (o.serial_number || '').toLowerCase().includes(q) || (o.name || '').toLowerCase().includes(q)
              }).length === 0 ? (
                <div className="p-6 text-center text-xs text-muted-foreground">Tidak ada ONU ditemukan.</div>
              ) : (
                detectedOnusList
                  .filter((o) => {
                    const q = searchDetectedOnu.toLowerCase().trim()
                    if (!q) return true
                    return (o.serial_number || '').toLowerCase().includes(q) || (o.name || '').toLowerCase().includes(q)
                  })
                  .map((onu, idx) => (
                    <div key={`${onu.serial_number || idx}`} className="flex items-center justify-between gap-2 p-2.5 text-xs hover:bg-muted/40">
                      <div className="min-w-0">
                        <div className="font-mono font-semibold">{onu.serial_number || onu.name || '-'}</div>
                        <div className="text-muted-foreground">
                          {onu.name || '-'} · Tipe: {onu.onu_type || '-'} · Rx: {onu.rx_power ?? '-'} dBm · ID: {onu.onu_id ?? '-'}
                        </div>
                      </div>
                      <Button type="button" size="sm" variant="outline" onClick={() => selectDetectedONU(onu)}>
                        Pilih
                      </Button>
                    </div>
                  ))
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Data Teknis</AlertDialogTitle>
            <AlertDialogDescription>
              Yakin ingin menghapus data teknis untuk "{deleteTarget?.pelanggan?.nama || deleteTarget?.id_pelanggan}"?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              Hapus
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
