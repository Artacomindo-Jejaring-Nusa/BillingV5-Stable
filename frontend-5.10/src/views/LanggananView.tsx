import { useCallback, useEffect, useState } from 'react'
import {
  ArrowUpDown,
  ChevronDown,
  Copy,
  Download,
  Edit,
  ExternalLink,
  Eye,
  FileSpreadsheet,
  Filter,
  Loader2,
  Plus,
  Receipt,
  RefreshCw,
  Search,
  Trash2,
  Upload,
  X,
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
import type { HargaLayanan, PaketLayanan } from './HargaLayananView'

export interface LanggananItem {
  id: number
  pelanggan_id: number
  paket_layanan_id: number
  nama_paket?: string
  nama_pelanggan?: string
  no_telp?: string
  alamat?: string
  brand?: string
  id_brand?: string
  status: string
  tgl_jatuh_tempo: string | null
  tgl_jatuh_tempo_pembayaran?: string | null
  tgl_invoice_terakhir?: string | null
  metode_pembayaran: string
  harga_awal: number | null
  harga_final: number
  tgl_mulai_langganan?: string | null
  tgl_berhenti?: string | null
  alasan_berhenti?: string | null
  status_modem?: string | null
}

interface PelangganSelectItem {
  id: number
  nama: string
  id_brand?: string
}

interface InvoiceHistoryItem {
  id: number
  invoice_number: string
  created_at?: string
  tgl_invoice?: string
  tgl_jatuh_tempo?: string
  paid_at?: string | null
  total_harga: number
  status_invoice: string
  payment_link?: string | null
}

interface DataTeknisSummary {
  id_pelanggan?: string
  ip_pelanggan?: string
  profile_pppoe?: string
  id_vlan?: string
  olt?: string
  onu_power?: number
}

function formatCurrency(val?: number | null) {
  if (typeof val !== 'number' || Number.isNaN(val)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(val)
}

function formatDate(d?: string | null) {
  if (!d) return '-'
  const date = new Date(d)
  if (Number.isNaN(date.getTime())) return d
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

export function LanggananView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_langganan')
  const canEdit = hasPermission('edit_langganan')
  const canDelete = hasPermission('delete_langganan')
  const [list, setList] = useState<LanggananItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [initialLoading, setInitialLoading] = useState(true)
  const [fetching, setFetching] = useState(false)
  const [saving, setSaving] = useState(false)
  const [exporting, setExporting] = useState(false)

  // Dropdowns
  const [locations, setLocations] = useState<string[]>([])
  const [brands, setBrands] = useState<HargaLayanan[]>([])
  const [paketList, setPaketList] = useState<PaketLayanan[]>([])
  const [pelangganOptions, setPelangganOptions] = useState<PelangganSelectItem[]>([])

  // Filters & Server-side Sorting & Pagination
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [alamatFilter, setAlamatFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('all')
  const [brandFilter, setBrandFilter] = useState('all')
  const [sortBy, setSortBy] = useState<{ key: string; order: 'asc' | 'desc' } | null>(null)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(15)

  // Form Dialog
  const [formOpen, setFormOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<Partial<LanggananItem>>({
    status: 'Aktif',
    metode_pembayaran: 'Otomatis',
    harga_awal: 0,
    status_modem: 'Terpasang',
  })
  const [proratePlusFull, setProratePlusFull] = useState(false)
  const [isPaketLocked, setIsPaketLocked] = useState(false)
  const [filteredPaketList, setFilteredPaketList] = useState<PaketLayanan[]>([])
  const [formError, setFormError] = useState('')

  // Tagihan Dialog
  const [tagihanTarget, setTagihanTarget] = useState<LanggananItem | null>(null)
  const [tagihanList, setTagihanList] = useState<InvoiceHistoryItem[]>([])
  const [loadingTagihan, setLoadingTagihan] = useState(false)

  // View Detail Dialog
  const [viewTarget, setViewTarget] = useState<LanggananItem | null>(null)
  const [viewTeknis, setViewTeknis] = useState<DataTeknisSummary | null>(null)
  const [viewStats, setViewStats] = useState({ total: 0, onTime: 0, late: 0 })
  const [loadingView, setLoadingView] = useState(false)

  const [deleteTarget, setDeleteTarget] = useState<LanggananItem | null>(null)
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
    const [locRes, brandRes, pkgRes] = await Promise.allSettled([
      apiClient.get<string[]>('/pelanggan/lokasi/unik'),
      apiClient.get<HargaLayanan[]>('/harga_layanan'),
      apiClient.get<PaketLayanan[]>('/paket_layanan'),
    ])
    if (locRes.status === 'fulfilled' && Array.isArray(locRes.value.data)) setLocations(locRes.value.data)
    if (brandRes.status === 'fulfilled' && Array.isArray(brandRes.value.data)) setBrands(brandRes.value.data)
    if (pkgRes.status === 'fulfilled' && Array.isArray(pkgRes.value.data)) setPaketList(pkgRes.value.data)
  }, [])

  const fetchLangganan = useCallback(async () => {
    setFetching(true)
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (alamatFilter !== 'all') params.append('alamat', alamatFilter)
      if (statusFilter !== 'all') params.append('status', statusFilter)
      if (brandFilter !== 'all') params.append('id_brand', brandFilter)
      if (sortBy) {
        params.append('sort_by', sortBy.key)
        params.append('sort_order', sortBy.order)
      }
      params.append('skip', String((page - 1) * limit))
      params.append('limit', String(limit))

      const res = await apiClient.get(`/langganan?${params.toString()}`)
      const payload = res.data
      if (payload?.data && payload.total_count !== undefined) {
        setList(payload.data)
        setTotalCount(payload.total_count)
      } else if (Array.isArray(payload)) {
        setList(payload)
        setTotalCount(payload.length)
      }
    } catch {
      setFeedback({ type: 'error', text: 'Gagal memuat data langganan.' })
    } finally {
      setInitialLoading(false)
      setFetching(false)
    }
  }, [debouncedSearch, alamatFilter, statusFilter, brandFilter, sortBy, page, limit])

  useEffect(() => {
    void fetchDropdowns()
  }, [fetchDropdowns])

  useEffect(() => {
    void fetchLangganan()
  }, [fetchLangganan])

  // Auto-calculate price
  useEffect(() => {
    if (!formOpen || !form.pelanggan_id || !form.paket_layanan_id) return
    const metode = form.metode_pembayaran || 'Otomatis'
    if (metode === 'Prorate' && !form.tgl_mulai_langganan && !editingId) return

    let cancelled = false
    const calc = async () => {
      try {
        const endpoint =
          metode === 'Prorate' && proratePlusFull
            ? '/langganan/calculate-prorate-plus-full'
            : '/langganan/calculate-price'
        const res = await apiClient.post(endpoint, {
          paket_layanan_id: form.paket_layanan_id,
          metode_pembayaran: metode,
          pelanggan_id: form.pelanggan_id,
          ...(form.tgl_mulai_langganan && { tgl_mulai: form.tgl_mulai_langganan }),
        })
        if (cancelled) return
        setForm((prev) => ({
          ...prev,
          harga_awal:
            metode === 'Prorate' && proratePlusFull ? res.data.harga_total_awal : res.data.harga_awal,
          tgl_jatuh_tempo: res.data.tgl_jatuh_tempo
            ? new Date(res.data.tgl_jatuh_tempo).toISOString().split('T')[0]
            : prev.tgl_jatuh_tempo,
          tgl_jatuh_tempo_pembayaran: res.data.tgl_jatuh_tempo_pembayaran
            ? new Date(res.data.tgl_jatuh_tempo_pembayaran).toISOString().split('T')[0]
            : prev.tgl_jatuh_tempo_pembayaran,
          tgl_mulai_langganan: res.data.tgl_mulai_langganan
            ? new Date(res.data.tgl_mulai_langganan).toISOString().split('T')[0]
            : prev.tgl_mulai_langganan,
        }))
      } catch {
        // ignore calculation failure while typing
      }
    }
    void calc()
    return () => {
      cancelled = true
    }
  }, [
    formOpen,
    editingId,
    form.pelanggan_id,
    form.paket_layanan_id,
    form.metode_pembayaran,
    form.tgl_mulai_langganan,
    proratePlusFull,
  ])

  const handlePelangganSelect = async (pelangganIdStr: string) => {
    const pelangganId = Number(pelangganIdStr)
    setFormError('')
    setIsPaketLocked(false)
    setForm((prev) => ({ ...prev, pelanggan_id: pelangganId, paket_layanan_id: undefined }))

    try {
      const res = await apiClient.get(`/pelanggan/${pelangganId}`)
      const detail = res.data?.data ?? res.data
      const tglMulai = detail?.tgl_instalasi
        ? new Date(detail.tgl_instalasi).toISOString().split('T')[0]
        : new Date().toISOString().split('T')[0]

      const brandFiltered = detail?.id_brand
        ? paketList.filter((p) => p.id_brand === detail.id_brand)
        : paketList
      setFilteredPaketList(brandFiltered)

      const matchedPaket = detail?.layanan
        ? brandFiltered.find((p) => p.nama_paket === detail.layanan)
        : undefined

      if (matchedPaket) {
        setIsPaketLocked(true)
      }

      setForm((prev) => ({
        ...prev,
        pelanggan_id: pelangganId,
        tgl_mulai_langganan: tglMulai,
        paket_layanan_id: matchedPaket?.id ?? prev.paket_layanan_id,
      }))
    } catch {
      setFilteredPaketList(paketList)
    }
  }

  const openCreate = async () => {
    setEditingId(null)
    setProratePlusFull(false)
    setIsPaketLocked(false)
    setFilteredPaketList(paketList)
    setForm({
      status: 'Aktif',
      metode_pembayaran: 'Otomatis',
      harga_awal: 0,
      status_modem: 'Terpasang',
      tgl_mulai_langganan: new Date().toISOString().split('T')[0],
    })
    setFormError('')
    setFormOpen(true)
    try {
      const [pelRes, activeRes] = await Promise.all([
        apiClient.get<{ data: PelangganSelectItem[] }>('/pelanggan?for_invoice_selection=true&limit=2000'),
        apiClient.get<number[]>('/langganan/active-pelanggan-ids'),
      ])
      const activeSet = new Set(activeRes.data || [])
      const allPel = pelRes.data?.data || []
      setPelangganOptions(allPel.filter((p) => !activeSet.has(p.id)))
    } catch {
      setPelangganOptions([])
    }
  }

  const openEdit = (item: LanggananItem) => {
    setEditingId(item.id)
    setProratePlusFull(false)
    setIsPaketLocked(false)
    setFilteredPaketList(paketList)
    setForm({
      ...item,
      tgl_jatuh_tempo: item.tgl_jatuh_tempo ? new Date(item.tgl_jatuh_tempo).toISOString().split('T')[0] : '',
      tgl_jatuh_tempo_pembayaran: item.tgl_jatuh_tempo_pembayaran
        ? new Date(item.tgl_jatuh_tempo_pembayaran).toISOString().split('T')[0]
        : '',
      tgl_mulai_langganan: item.tgl_mulai_langganan
        ? new Date(item.tgl_mulai_langganan).toISOString().split('T')[0]
        : '',
      tgl_berhenti: item.tgl_berhenti ? new Date(item.tgl_berhenti).toISOString().split('T')[0] : '',
      status_modem: item.status_modem || (item.status === 'Berhenti' ? 'Diambil' : 'Terpasang'),
    })
    setFormError('')
    setFormOpen(true)
  }

  const openTagihanModal = async (item: LanggananItem) => {
    setTagihanTarget(item)
    setTagihanList([])
    setLoadingTagihan(true)
    try {
      const res = await apiClient.get(`/invoices?pelanggan_id=${item.pelanggan_id}&limit=100`)
      const raw = res.data?.data ?? res.data
      setTagihanList(Array.isArray(raw) ? raw : [])
    } catch {
      setTagihanList([])
    } finally {
      setLoadingTagihan(false)
    }
  }

  const openViewModal = async (item: LanggananItem) => {
    setViewTarget(item)
    setViewTeknis(null)
    setViewStats({ total: 0, onTime: 0, late: 0 })
    setLoadingView(true)
    try {
      const [dtRes, invRes] = await Promise.allSettled([
        apiClient.get(`/data_teknis/by-pelanggan/${item.pelanggan_id}`),
        apiClient.get(`/invoices?pelanggan_id=${item.pelanggan_id}&limit=100`),
      ])
      if (dtRes.status === 'fulfilled') {
        const arr = Array.isArray(dtRes.value.data) ? dtRes.value.data : [dtRes.value.data]
        setViewTeknis(arr[0] || null)
      }
      if (invRes.status === 'fulfilled') {
        const raw = invRes.value.data?.data ?? invRes.value.data
        const invs: InvoiceHistoryItem[] = Array.isArray(raw) ? raw : []
        const late = invs.filter((inv) => {
          if (!inv.paid_at || !inv.tgl_jatuh_tempo) return false
          return new Date(inv.paid_at) > new Date(inv.tgl_jatuh_tempo)
        }).length
        setViewStats({
          total: invs.length,
          late,
          onTime: invs.length - late,
        })
      }
    } finally {
      setLoadingView(false)
    }
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (saving) return
    if (!form.pelanggan_id || !form.paket_layanan_id) {
      setFormError('Pelanggan dan Paket Layanan wajib dipilih.')
      return
    }
    setSaving(true)
    setFormError('')
    try {
      if (!editingId) {
        const dtRes = await apiClient.get(`/data_teknis/by-pelanggan/${form.pelanggan_id}`)
        if (!dtRes.data || dtRes.data.length === 0) {
          setFormError(
            'Pelanggan belum memiliki Data Teknis. Hubungi NOC untuk menambahkan Data Teknis terlebih dahulu.'
          )
          setSaving(false)
          return
        }
        await apiClient.post('/langganan', form)
        await fetchLangganan()
        setFeedback({ type: 'success', text: 'Langganan berhasil ditambahkan.' })
      } else {
        await apiClient.patch(`/langganan/${editingId}`, form)
        await fetchLangganan()
        setFeedback({ type: 'success', text: 'Langganan berhasil diperbarui.' })
      }
      setFormOpen(false)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { detail?: string; error?: string } } })?.response?.data?.detail ||
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Gagal menyimpan data langganan.'
      setFormError(msg)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await apiClient.delete(`/langganan/${deleteTarget.id}`)
      await fetchLangganan()
      setFeedback({ type: 'success', text: 'Data langganan berhasil dihapus.' })
      setDeleteTarget(null)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menghapus data langganan.' })
      setDeleteTarget(null)
    }
  }

  const handleSort = (key: string) => {
    setSortBy((prev) => {
      if (prev?.key === key) {
        return prev.order === 'asc' ? { key, order: 'desc' } : null
      }
      return { key, order: 'asc' }
    })
    setPage(1)
  }

  const buildFilterParams = () => {
    const params = new URLSearchParams()
    if (debouncedSearch) params.append('search', debouncedSearch)
    if (alamatFilter !== 'all') params.append('alamat', alamatFilter)
    if (statusFilter !== 'all') params.append('status', statusFilter)
    if (brandFilter !== 'all') params.append('id_brand', brandFilter)
    return params
  }

  const handleExport = async (mode: 'csv' | 'excel' | 'multi-sheet') => {
    setExporting(true)
    try {
      const params = buildFilterParams()
      let urlPath = '/langganan/export'
      let filename = 'langganan.csv'
      if (mode === 'excel') {
        params.append('format', 'excel')
        filename = 'langganan.xlsx'
      } else if (mode === 'csv') {
        params.append('format', 'csv')
        filename = 'langganan.csv'
      } else {
        params.append('limit', '5000')
        urlPath = '/langganan/export/excel/multi-sheet'
        filename = `Laporan_Langganan_Lengkap_${new Date().toISOString().slice(0, 10)}.xlsx`
      }

      const res = await apiClient.get(`${urlPath}?${params.toString()}`, { responseType: 'blob' })
      const url = URL.createObjectURL(new Blob([res.data]))
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengekspor data langganan.' })
    } finally {
      setExporting(false)
    }
  }

  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!importFile) return
    setImporting(true)
    try {
      const fd = new FormData()
      fd.append('file', importFile)
      await apiClient.post('/langganan/import/csv', fd)
      setImportOpen(false)
      setImportFile(null)
      await fetchLangganan()
      setFeedback({ type: 'success', text: 'Import data langganan berhasil.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengimpor file CSV langganan.' })
    } finally {
      setImporting(false)
    }
  }

  const copyText = async (text?: string | null) => {
    if (!text) return
    await navigator.clipboard.writeText(text)
    setFeedback({ type: 'success', text: 'Link pembayaran disalin ke clipboard.' })
  }

  const statusBadgeVariant = (status: string) => {
    const s = status.toLowerCase()
    if (s === 'aktif' || s === 'lunas') return 'secondary'
    if (s === 'suspended' || s === 'expired') return 'destructive'
    return 'outline'
  }

  const totalPages = Math.max(1, Math.ceil(totalCount / limit))

  return (
    <div className="p-4 md:p-6 space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Manajemen Langganan</h2>
          <p className="text-sm text-muted-foreground">
            Total {totalCount.toLocaleString('id-ID')} langganan pelanggan
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          {canCreate && (
            <Button variant="outline" size="sm" onClick={() => setImportOpen(true)}>
              <Upload className="mr-1.5 h-4 w-4" />
              Import
            </Button>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" disabled={exporting}>
                <Download className="mr-1.5 h-4 w-4" />
                {exporting ? 'Mengekspor...' : 'Export'}
                <ChevronDown className="ml-1.5 h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-72">
              <DropdownMenuItem onClick={() => void handleExport('csv')}>
                <Download className="mr-2 h-4 w-4" />
                Export sebagai CSV
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => void handleExport('excel')}>
                <FileSpreadsheet className="mr-2 h-4 w-4" />
                Export sebagai Excel
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => void handleExport('multi-sheet')} className="items-start">
                <FileSpreadsheet className="mr-2 h-4 w-4 mt-0.5 text-primary shrink-0" />
                <div>
                  <div className="font-medium text-primary">Export Multi-Sheet (Lengkap)</div>
                  <div className="text-[11px] text-muted-foreground">
                    Data langganan + laporan pasang & berhenti + riwayat invoice + statistik
                  </div>
                </div>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {canCreate && (
            <Button size="sm" onClick={() => void openCreate()}>
              <Plus className="mr-1.5 h-4 w-4" />
              Tambah Langganan
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
                placeholder="Cari berdasarkan nama pelanggan..."
                className="pl-9"
              />
            </div>
            <Button variant="outline" onClick={() => setShowFilters((v) => !v)}>
              <Filter className="mr-1.5 h-4 w-4" />
              Filter
            </Button>
            <Button variant="outline" onClick={() => void fetchLangganan()} disabled={fetching}>
              <RefreshCw className={`h-4 w-4 ${fetching ? 'animate-spin' : ''}`} />
            </Button>
          </div>

          {showFilters && (
            <div className="grid grid-cols-1 sm:grid-cols-4 gap-3 pt-2 border-t">
              <Select
                value={alamatFilter}
                onValueChange={(v) => {
                  setAlamatFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Alamat" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Alamat</SelectItem>
                  {locations.map((loc) => (
                    <SelectItem key={loc} value={loc}>
                      {loc}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={statusFilter}
                onValueChange={(v) => {
                  setStatusFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Status</SelectItem>
                  <SelectItem value="Aktif">Aktif</SelectItem>
                  <SelectItem value="Suspended">Suspended</SelectItem>
                  <SelectItem value="Berhenti">Berhenti</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={brandFilter}
                onValueChange={(v) => {
                  setBrandFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Brand" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Brand</SelectItem>
                  {brands.map((b) => (
                    <SelectItem key={b.id_brand} value={b.id_brand}>
                      {b.brand}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button
                variant="ghost"
                onClick={() => {
                  setSearchQuery('')
                  setAlamatFilter('all')
                  setStatusFilter('all')
                  setBrandFilter('all')
                  setSortBy(null)
                  setPage(1)
                }}
              >
                Reset Filter
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        {fetching && !initialLoading && (
          <div className="h-1 w-full bg-primary/20 overflow-hidden">
            <div className="h-full w-1/3 bg-primary animate-pulse" />
          </div>
        )}
        <CardContent className="p-0">
          {initialLoading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : list.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">
              Tidak ada data langganan yang ditemukan.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="w-12 px-4 py-3 text-xs uppercase tracking-wide">No</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">
                      <button
                        type="button"
                        onClick={() => handleSort('pelanggan.nama')}
                        className="inline-flex items-center gap-1 font-medium hover:text-foreground"
                      >
                        Nama Pelanggan
                        <ArrowUpDown className="h-3.5 w-3.5" />
                      </button>
                    </TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Alamat</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">No. Telepon</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Brand</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Paket Layanan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Metode Bayar</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">
                      <button
                        type="button"
                        onClick={() => handleSort('harga_final')}
                        className="inline-flex items-center gap-1 font-medium hover:text-foreground"
                      >
                        Harga
                        <ArrowUpDown className="h-3.5 w-3.5" />
                      </button>
                    </TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">
                      <button
                        type="button"
                        onClick={() => handleSort('status')}
                        className="inline-flex items-center gap-1 font-medium hover:text-foreground"
                      >
                        Status
                        <ArrowUpDown className="h-3.5 w-3.5" />
                      </button>
                    </TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Jatuh Tempo</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide text-right whitespace-nowrap">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((item, idx) => (
                    <TableRow key={item.id}>
                      <TableCell className="px-4 py-3.5 text-xs text-muted-foreground">
                        {(page - 1) * limit + idx + 1}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 font-medium whitespace-nowrap">
                        {item.nama_pelanggan || `Pelanggan #${item.pelanggan_id}`}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{item.alamat || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{item.no_telp || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">
                        <Badge variant="outline">{item.brand || '-'}</Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{item.nama_paket || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">
                        <Badge variant="secondary">{item.metode_pembayaran}</Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 font-semibold text-xs whitespace-nowrap">
                        {formatCurrency(item.harga_final || item.harga_awal)}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 whitespace-nowrap">
                        <Badge variant={statusBadgeVariant(item.status)}>{item.status}</Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">
                        {formatDate(item.tgl_jatuh_tempo)}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-right whitespace-nowrap">
                        <div className="inline-flex items-center gap-1">
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-7 px-2 text-xs text-purple-700 border-purple-200 hover:bg-purple-50 dark:text-purple-300 dark:border-purple-800 dark:hover:bg-purple-950"
                            onClick={() => void openTagihanModal(item)}
                          >
                            <Receipt className="mr-1 h-3.5 w-3.5" />
                            Tagihan
                          </Button>
                          {canEdit && (
                            <Button size="sm" variant="outline" className="h-7 px-2 text-xs" onClick={() => openEdit(item)}>
                              <Edit className="mr-1 h-3.5 w-3.5" />
                              Edit
                            </Button>
                          )}
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-7 px-2 text-xs text-sky-700 border-sky-200 hover:bg-sky-50 dark:text-sky-300 dark:border-sky-800 dark:hover:bg-sky-950"
                            onClick={() => void openViewModal(item)}
                          >
                            <Eye className="mr-1 h-3.5 w-3.5" />
                            View
                          </Button>
                          {canDelete && (
                            <Button
                              size="sm"
                              variant="outline"
                              className="h-7 px-2 text-xs text-destructive hover:text-destructive"
                              onClick={() => setDeleteTarget(item)}
                            >
                              <Trash2 className="mr-1 h-3.5 w-3.5" />
                              Hapus
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
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

      {/* Riwayat Tagihan Modal */}
      <Dialog open={!!tagihanTarget} onOpenChange={(o) => !o && setTagihanTarget(null)}>
        <DialogContent className="sm:max-w-4xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Riwayat Tagihan: {tagihanTarget?.nama_pelanggan}</DialogTitle>
            <DialogDescription>
              ID Pelanggan: {tagihanTarget?.pelanggan_id} • Paket: {tagihanTarget?.nama_paket || '-'}
            </DialogDescription>
          </DialogHeader>

          {loadingTagihan ? (
            <div className="space-y-2 py-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : tagihanList.length === 0 ? (
            <div className="py-10 text-center text-sm text-muted-foreground">
              Belum ada riwayat tagihan untuk pelanggan ini.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead>Nomor Invoice</TableHead>
                    <TableHead>Tanggal</TableHead>
                    <TableHead>Jatuh Tempo</TableHead>
                    <TableHead>Total</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Link Pembayaran</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tagihanList.map((inv) => (
                    <TableRow key={inv.id}>
                      <TableCell className="font-mono text-xs font-medium">{inv.invoice_number}</TableCell>
                      <TableCell className="text-xs">{formatDate(inv.tgl_invoice || inv.created_at)}</TableCell>
                      <TableCell className="text-xs">{formatDate(inv.tgl_jatuh_tempo)}</TableCell>
                      <TableCell className="text-xs font-semibold text-emerald-600">
                        {formatCurrency(inv.total_harga)}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusBadgeVariant(inv.status_invoice)}>{inv.status_invoice}</Badge>
                      </TableCell>
                      <TableCell className="text-right whitespace-nowrap">
                        {inv.payment_link ? (
                          <div className="inline-flex items-center gap-1">
                            <Button
                              size="icon"
                              variant="ghost"
                              className="h-7 w-7"
                              onClick={() => void copyText(inv.payment_link)}
                              title="Salin link pembayaran"
                            >
                              <Copy className="h-3.5 w-3.5" />
                            </Button>
                            <Button size="icon" variant="ghost" className="h-7 w-7" asChild title="Buka link">
                              <a href={inv.payment_link} target="_blank" rel="noreferrer">
                                <ExternalLink className="h-3.5 w-3.5" />
                              </a>
                            </Button>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">-</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}

          <DialogFooter>
            <Button onClick={() => setTagihanTarget(null)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* View Detail Modal */}
      <Dialog open={!!viewTarget} onOpenChange={(o) => !o && setViewTarget(null)}>
        <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Detail Langganan: {viewTarget?.nama_pelanggan}</DialogTitle>
            <DialogDescription>
              {viewTarget?.brand} • {viewTarget?.nama_paket}
            </DialogDescription>
          </DialogHeader>

          {loadingView ? (
            <div className="space-y-2 py-4">
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          ) : (
            viewTarget && (
              <div className="space-y-4 text-sm">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="rounded-lg border p-4 space-y-2">
                    <div className="font-semibold text-xs uppercase text-muted-foreground">Informasi Langganan</div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Status</span>
                      <Badge variant={statusBadgeVariant(viewTarget.status)}>{viewTarget.status}</Badge>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Harga Tagihan</span>
                      <span className="font-semibold">{formatCurrency(viewTarget.harga_final || viewTarget.harga_awal)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Metode Bayar</span>
                      <span>{viewTarget.metode_pembayaran}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Mulai Langganan</span>
                      <span>{formatDate(viewTarget.tgl_mulai_langganan)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Jatuh Tempo</span>
                      <span>{formatDate(viewTarget.tgl_jatuh_tempo)}</span>
                    </div>
                    {viewTarget.alasan_berhenti && (
                      <div className="flex justify-between text-destructive">
                        <span>Alasan Berhenti</span>
                        <span>{viewTarget.alasan_berhenti}</span>
                      </div>
                    )}
                  </div>

                  <div className="rounded-lg border p-4 space-y-2">
                    <div className="font-semibold text-xs uppercase text-muted-foreground">Informasi Teknis</div>
                    {viewTeknis ? (
                      <>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">PPPoE Secret</span>
                          <span className="font-mono text-xs">{viewTeknis.id_pelanggan || '-'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">IP Pelanggan</span>
                          <span className="font-mono text-xs">{viewTeknis.ip_pelanggan || '-'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Profile / VLAN</span>
                          <span>
                            {viewTeknis.profile_pppoe || '-'} / {viewTeknis.id_vlan || '-'}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">OLT</span>
                          <span>{viewTeknis.olt || '-'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">ONU Power</span>
                          <span>{viewTeknis.onu_power ?? 0} dBm</span>
                        </div>
                      </>
                    ) : (
                      <div className="text-xs text-muted-foreground">Data teknis belum tersedia.</div>
                    )}
                  </div>
                </div>

                <div className="rounded-lg border p-4">
                  <div className="font-semibold text-xs uppercase text-muted-foreground mb-3">
                    Ringkasan Pembayaran
                  </div>
                  <div className="grid grid-cols-3 gap-3 text-center">
                    <div className="rounded-md bg-muted/40 p-2.5">
                      <div className="text-xs text-muted-foreground">Total Invoice</div>
                      <div className="text-lg font-bold">{viewStats.total}</div>
                    </div>
                    <div className="rounded-md bg-emerald-500/10 p-2.5">
                      <div className="text-xs text-emerald-700 dark:text-emerald-300">Tepat Waktu</div>
                      <div className="text-lg font-bold text-emerald-600">{viewStats.onTime}</div>
                    </div>
                    <div className="rounded-md bg-destructive/10 p-2.5">
                      <div className="text-xs text-destructive">Terlambat</div>
                      <div className="text-lg font-bold text-destructive">{viewStats.late}</div>
                    </div>
                  </div>
                </div>
              </div>
            )
          )}

          <DialogFooter>
            <Button onClick={() => setViewTarget(null)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add / Edit Dialog (2-Column Layout matching existing) */}
      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="sm:max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingId ? 'Edit Langganan' : 'Tambah Langganan Baru'}</DialogTitle>
            <DialogDescription>
              {editingId ? 'Perbarui informasi langganan pelanggan' : 'Tambahkan langganan baru untuk pelanggan'}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSave} className="space-y-5">
            {formError && (
              <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {formError}
              </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
              {/* Left Column: Informasi Pelanggan + Paket & Pembayaran */}
              <div className="space-y-5">
                <div className="rounded-lg border p-4 space-y-3">
                  <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Informasi Pelanggan
                  </div>
                  {!editingId ? (
                    <div className="space-y-1.5">
                      <Label>
                        Pilih Pelanggan <span className="text-destructive">*</span>
                      </Label>
                      <Select
                        value={form.pelanggan_id ? String(form.pelanggan_id) : ''}
                        onValueChange={(v) => void handlePelangganSelect(v)}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Pilih pelanggan..." />
                        </SelectTrigger>
                        <SelectContent>
                          {pelangganOptions.map((p) => (
                            <SelectItem key={p.id} value={String(p.id)}>
                              {p.nama}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  ) : (
                    <div className="space-y-1.5">
                      <Label>Pelanggan</Label>
                      <Input value={form.nama_pelanggan || ''} disabled />
                    </div>
                  )}
                </div>

                <div className="rounded-lg border p-4 space-y-3">
                  <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Paket & Pembayaran
                  </div>

                  <div className="space-y-1.5">
                    <Label>
                      Pilih Paket Layanan <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.paket_layanan_id ? String(form.paket_layanan_id) : ''}
                      onValueChange={(v) => setForm({ ...form, paket_layanan_id: Number(v) })}
                      disabled={(!editingId && !form.pelanggan_id) || isPaketLocked}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue
                          placeholder={
                            !editingId && !form.pelanggan_id
                              ? 'Pilih pelanggan terlebih dahulu'
                              : 'Pilih Paket Layanan'
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {(filteredPaketList.length > 0 ? filteredPaketList : paketList).map((pk) => (
                          <SelectItem key={pk.id} value={String(pk.id)}>
                            {pk.nama_paket} ({formatCurrency(pk.harga)})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {isPaketLocked && (
                      <p className="text-[11px] text-muted-foreground">
                        Paket dipilih otomatis sesuai data registrasi pelanggan.
                      </p>
                    )}
                  </div>

                  <div className="space-y-1.5">
                    <Label>
                      Metode Pembayaran <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.metode_pembayaran || 'Otomatis'}
                      onValueChange={(v) => setForm({ ...form, metode_pembayaran: v })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="Otomatis">Otomatis (Bulanan)</SelectItem>
                        <SelectItem value="Prorate">Prorate</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  {form.metode_pembayaran === 'Prorate' && (
                    <div className="flex items-center gap-2 pt-1">
                      <input
                        id="prorate_plus"
                        type="checkbox"
                        checked={proratePlusFull}
                        onChange={(e) => setProratePlusFull(e.target.checked)}
                        className="h-4 w-4 rounded border-input"
                      />
                      <Label htmlFor="prorate_plus" className="text-xs font-normal">
                        Prorate + Full Bulan Depan
                      </Label>
                    </div>
                  )}

                  <div className="rounded-md bg-muted/50 px-3.5 py-2.5 flex items-center justify-between">
                    <span className="text-xs font-medium text-muted-foreground">Total Harga Bulanan</span>
                    <span className="text-base font-bold">{formatCurrency(form.harga_awal ?? 0)}</span>
                  </div>
                </div>
              </div>

              {/* Right Column: Status & Jadwal */}
              <div className="rounded-lg border p-4 space-y-3.5 h-fit">
                <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Status & Jadwal
                </div>

                <div className="space-y-1.5">
                  <Label>
                    Status Langganan <span className="text-destructive">*</span>
                  </Label>
                  <Select
                    value={form.status || 'Aktif'}
                    onValueChange={(v) => setForm({ ...form, status: v })}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Aktif">Aktif</SelectItem>
                      <SelectItem value="Suspended">Suspended</SelectItem>
                      <SelectItem value="Berhenti">Berhenti</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="tgl_mulai">
                      Tanggal Mulai Langganan <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="tgl_mulai"
                      type="date"
                      value={form.tgl_mulai_langganan || ''}
                      onChange={(e) => setForm({ ...form, tgl_mulai_langganan: e.target.value })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="tgl_jatuh_tempo">
                      Tanggal Berakhir Langganan <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="tgl_jatuh_tempo"
                      type="date"
                      value={form.tgl_jatuh_tempo || ''}
                      onChange={(e) => setForm({ ...form, tgl_jatuh_tempo: e.target.value })}
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="tgl_jatuh_tempo_pembayaran">
                    Tanggal Jatuh Tempo Pembayaran <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="tgl_jatuh_tempo_pembayaran"
                    type="date"
                    value={form.tgl_jatuh_tempo_pembayaran || ''}
                    onChange={(e) => setForm({ ...form, tgl_jatuh_tempo_pembayaran: e.target.value })}
                  />
                  <p className="text-[11px] text-muted-foreground">Tanggal batas akhir pembayaran invoice</p>
                </div>

                <div className="space-y-1.5">
                  <Label>
                    Status Modem <span className="text-xs text-muted-foreground font-normal">(Wajib diisi)</span>
                  </Label>
                  <Select
                    value={form.status_modem || 'Terpasang'}
                    onValueChange={(v) => setForm({ ...form, status_modem: v })}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Terpasang">Terpasang</SelectItem>
                      <SelectItem value="Replacement">Replacement</SelectItem>
                      <SelectItem value="Rusak">Rusak</SelectItem>
                      <SelectItem value="Diambil">Diambil</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {form.status === 'Berhenti' && (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                    <div className="space-y-1.5">
                      <Label htmlFor="tgl_berhenti">Tanggal Berhenti</Label>
                      <Input
                        id="tgl_berhenti"
                        type="date"
                        value={form.tgl_berhenti || ''}
                        onChange={(e) => setForm({ ...form, tgl_berhenti: e.target.value })}
                      />
                    </div>
                    <div className="space-y-1.5">
                      <Label htmlFor="alasan_berhenti">Alasan Berhenti</Label>
                      <Input
                        id="alasan_berhenti"
                        value={form.alasan_berhenti || ''}
                        onChange={(e) => setForm({ ...form, alasan_berhenti: e.target.value })}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setFormOpen(false)}>
                Batal
              </Button>
              <Button type="submit" disabled={saving}>
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {saving ? 'Menyimpan...' : editingId ? 'Simpan Perubahan' : '+ Tambah Langganan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Import CSV Dialog */}
      <Dialog open={importOpen} onOpenChange={setImportOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import Data Langganan (CSV)</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleImport} className="space-y-4">
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

      {/* Delete Confirm */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Langganan</AlertDialogTitle>
            <AlertDialogDescription>
              Yakin ingin menghapus langganan "{deleteTarget?.nama_pelanggan}"?
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
