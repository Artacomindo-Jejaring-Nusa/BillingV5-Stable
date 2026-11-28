import { useCallback, useEffect, useMemo, useState } from 'react'
import { Download, Edit, Filter, Loader2, Plus, RefreshCw, Search, Trash2, Upload, X } from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_LOCATIONS, fetchActiveLocations } from '@/utils/locations'
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
import type { HargaLayanan } from './HargaLayananView'

export interface PelangganItem {
  id: number
  customer_id?: string
  nama: string
  no_ktp?: string
  email?: string
  no_telp?: string
  layanan?: string
  alamat?: string
  blok?: string
  unit?: string
  tgl_instalasi?: string
  alamat_custom?: string | null
  id_brand?: string | null
  harga_layanan?: HargaLayanan | null
  connection_status?: string
}

const LAYANAN_OPTIONS = [
  'Internet 10 Mbps',
  'Internet 20 Mbps',
  'Internet 30 Mbps',
  'Internet 50 Mbps',
]

const emptyForm = (): Partial<PelangganItem> => ({
  nama: '',
  no_ktp: '',
  email: '',
  no_telp: '',
  layanan: '',
  alamat: '',
  blok: '',
  unit: '',
  tgl_instalasi: new Date().toISOString().split('T')[0],
  alamat_custom: '',
  id_brand: '',
})

export function PelangganView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_pelanggan')
  const canEdit = hasPermission('edit_pelanggan')
  const canDelete = hasPermission('delete_pelanggan')
  const [list, setList] = useState<PelangganItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [brands, setBrands] = useState<HargaLayanan[]>([])
  const [initialLoading, setInitialLoading] = useState(true)
  const [fetching, setFetching] = useState(false)
  const [saving, setSaving] = useState(false)
  const [alamatOptions, setAlamatOptions] = useState<string[]>(DEFAULT_LOCATIONS)

  // Filters & Pagination
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [alamatFilter, setAlamatFilter] = useState('all')
  const [brandFilter, setBrandFilter] = useState('all')
  const [layananFilter, setLayananFilter] = useState('all')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(15)

  // Dialogs
  const [formOpen, setFormOpen] = useState(false)
  const [formStep, setFormStep] = useState<1 | 2>(1)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<Partial<PelangganItem>>(emptyForm())
  const [formError, setFormError] = useState('')

  const [deleteTarget, setDeleteTarget] = useState<PelangganItem | null>(null)
  const [importOpen, setImportOpen] = useState(false)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importing, setImporting] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // Debounce search 500ms
  useEffect(() => {
    const t = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(1)
    }, 500)
    return () => clearTimeout(t)
  }, [searchQuery])

  const fetchBrands = useCallback(async () => {
    try {
      const res = await apiClient.get<HargaLayanan[]>('/harga_layanan')
      setBrands(res.data || [])
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    void fetchActiveLocations().then(setAlamatOptions)
  }, [])

  const formAlamatOptions = useMemo(() => {
    if (form.alamat && !alamatOptions.includes(form.alamat)) {
      return [form.alamat, ...alamatOptions]
    }
    return alamatOptions
  }, [alamatOptions, form.alamat])

  const fetchPelanggan = useCallback(async () => {
    setFetching(true)
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (alamatFilter !== 'all') params.append('alamat', alamatFilter)
      if (brandFilter !== 'all') params.append('id_brand', brandFilter)
      if (layananFilter !== 'all') params.append('layanan', layananFilter)
      params.append('skip', String((page - 1) * limit))
      params.append('limit', String(limit))
      params.append('use_minimal_loading', 'true')

      const res = await apiClient.get(`/pelanggan?${params.toString()}`)
      const payload = res.data
      if (payload?.data && payload.total_count !== undefined) {
        setList(payload.data)
        setTotalCount(payload.total_count)
      } else if (Array.isArray(payload)) {
        setList(payload)
        setTotalCount(payload.length)
      }
    } catch (err) {
      console.error('Failed to fetch pelanggan:', err)
      setFeedback({ type: 'error', text: 'Gagal memuat data pelanggan.' })
    } finally {
      setInitialLoading(false)
      setFetching(false)
    }
  }, [debouncedSearch, alamatFilter, brandFilter, layananFilter, page, limit])

  useEffect(() => {
    void fetchBrands()
  }, [fetchBrands])

  useEffect(() => {
    void fetchPelanggan()
  }, [fetchPelanggan])

  const openCreate = () => {
    setEditingId(null)
    setForm(emptyForm())
    setFormError('')
    setFormStep(1)
    setFormOpen(true)
  }

  const openEdit = (item: PelangganItem) => {
    setEditingId(item.id)
    setForm({
      ...item,
      tgl_instalasi: item.tgl_instalasi ? new Date(item.tgl_instalasi).toISOString().split('T')[0] : '',
      id_brand: item.id_brand || '',
    })
    setFormError('')
    setFormStep(1)
    setFormOpen(true)
  }

  const validateStep1 = (): string | null => {
    if (!form.nama?.trim()) return 'Nama lengkap wajib diisi.'
    if (!form.no_ktp?.trim() || !/^[0-9]{16}$/.test(form.no_ktp.trim())) {
      return 'Nomor KTP wajib diisi dan harus 16 digit angka.'
    }
    if (!form.email?.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
      return 'Email wajib diisi dengan format yang valid.'
    }
    if (!form.no_telp?.trim() || !/^[+]?[0-9\s()]{10,}$/.test(form.no_telp.trim())) {
      return 'Nomor telepon wajib diisi (minimal 10 digit angka, tanpa tanda "-").'
    }
    return null
  }

  const validateStep2 = (): string | null => {
    if (!form.id_brand) return 'Brand wajib dipilih.'
    if (!form.layanan) return 'Layanan wajib dipilih.'
    if (!form.alamat) return 'Alamat / Lokasi wajib dipilih.'
    if (!form.blok?.trim()) return 'Blok / Tower wajib diisi.'
    if (!form.unit?.trim()) return 'Unit / No. Rumah wajib diisi.'
    return null
  }

  const handleNextStep = () => {
    const err = validateStep1()
    if (err) {
      setFormError(err)
      return
    }
    setFormError('')
    setFormStep(2)
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (formStep !== 2) {
      handleNextStep()
      return
    }
    const step2Err = validateStep2()
    if (step2Err) {
      setFormError(step2Err)
      return
    }
    setSaving(true)
    setFormError('')
    try {
      const payload = { ...form, id_brand: form.id_brand || null }
      if (editingId) {
        await apiClient.put(`/pelanggan/${editingId}`, payload)
        await fetchPelanggan()
        setFeedback({ type: 'success', text: 'Data pelanggan berhasil diperbarui.' })
      } else {
        await apiClient.post('/pelanggan', payload)
        await fetchPelanggan()
        setFeedback({ type: 'success', text: 'Data pelanggan berhasil ditambahkan.' })
      }
      setFormOpen(false)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string; detail?: string } } })?.response?.data?.error ||
        (err as { response?: { data?: { detail?: string } } })?.response?.data?.detail ||
        'Gagal menyimpan data pelanggan.'
      setFormError(msg)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await apiClient.delete(`/pelanggan/${deleteTarget.id}`)
      setFeedback({ type: 'success', text: `Pelanggan "${deleteTarget.nama}" berhasil dihapus.` })
      setDeleteTarget(null)
      await fetchPelanggan()
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menghapus data pelanggan.' })
      setDeleteTarget(null)
    }
  }

  const handleExport = async (format: 'csv' | 'excel') => {
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (alamatFilter !== 'all') params.append('alamat', alamatFilter)
      if (brandFilter !== 'all') params.append('id_brand', brandFilter)
      params.append('format', format)
      const res = await apiClient.get(`/pelanggan/export?${params.toString()}`, { responseType: 'blob' })
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = `pelanggan.${format === 'excel' ? 'xlsx' : 'csv'}`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengekspor data pelanggan.' })
    }
  }

  const handleDownloadTemplate = async () => {
    try {
      const res = await apiClient.get('/pelanggan/template/csv', { responseType: 'blob' })
      const url = URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = 'template_import_pelanggan.csv'
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengunduh template CSV pelanggan.' })
    }
  }

  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!importFile) return
    setImporting(true)
    try {
      const fd = new FormData()
      fd.append('file', importFile)
      await apiClient.post('/pelanggan/import', fd)
      setImportOpen(false)
      setImportFile(null)
      setFeedback({ type: 'success', text: 'Import data pelanggan berhasil.' })
      await fetchPelanggan()
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengimpor file CSV pelanggan.' })
    } finally {
      setImporting(false)
    }
  }

  const getBrandName = (id?: string | null) => {
    if (!id) return '-'
    return brands.find((b) => b.id_brand === id)?.brand || id
  }

  const totalPages = Math.max(1, Math.ceil(totalCount / limit))

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Data Pelanggan</h2>
          <p className="text-sm text-muted-foreground">
            Total {totalCount.toLocaleString('id-ID')} pelanggan terdaftar
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
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
            <Button size="sm" onClick={openCreate}>
              <Plus className="mr-1.5 h-4 w-4" />
              Tambah Pelanggan
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
          <button type="button" onClick={() => setFeedback(null)} aria-label="Tutup notifikasi">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Search & Filters */}
      <Card>
        <CardContent className="pt-5 space-y-3">
          <div className="flex flex-col sm:flex-row gap-2">
            <div className="relative flex-1">
              <Search className="h-4 w-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari nama, email, atau no. telepon..."
                className="pl-9"
              />
            </div>
            <Button variant="outline" onClick={() => setShowFilters((v) => !v)}>
              <Filter className="mr-1.5 h-4 w-4" />
              Filter
            </Button>
            <Button variant="outline" onClick={() => void fetchPelanggan()} disabled={fetching}>
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
                <SelectTrigger>
                  <SelectValue placeholder="Semua Alamat" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Alamat</SelectItem>
                  {alamatOptions.map((a) => (
                    <SelectItem key={a} value={a}>
                      {a}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={brandFilter}
                onValueChange={(v) => {
                  setBrandFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger>
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

              <Select
                value={layananFilter}
                onValueChange={(v) => {
                  setLayananFilter(v)
                  setPage(1)
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Semua Layanan" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Layanan</SelectItem>
                  {LAYANAN_OPTIONS.map((l) => (
                    <SelectItem key={l} value={l}>
                      {l}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button
                variant="ghost"
                onClick={() => {
                  setSearchQuery('')
                  setAlamatFilter('all')
                  setBrandFilter('all')
                  setLayananFilter('all')
                  setPage(1)
                }}
              >
                Reset Filter
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Table */}
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
              Tidak ada data pelanggan yang ditemukan.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="w-14 px-4 py-3 text-xs uppercase tracking-wide">No</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Customer ID</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Pelanggan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">No. KTP</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Email</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Alamat</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Alamat Tambahan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Blok</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Unit</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">No. Telepon</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Layanan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Brand</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide whitespace-nowrap">Tgl Instalasi</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide text-right whitespace-nowrap">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((p, index) => (
                    <TableRow key={p.id}>
                      <TableCell className="px-4 py-3.5 text-xs text-muted-foreground">
                        {(page - 1) * limit + index + 1}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 font-mono font-semibold text-xs whitespace-nowrap">
                        {p.customer_id || '-'}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 font-medium whitespace-nowrap">{p.nama}</TableCell>
                      <TableCell className="px-4 py-3.5 font-mono text-xs whitespace-nowrap">{p.no_ktp || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.email || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.alamat || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs max-w-[220px] truncate" title={p.alamat_custom || ''}>
                        {p.alamat_custom || '-'}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.blok || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.unit || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.no_telp || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">{p.layanan || '-'}</TableCell>
                      <TableCell className="px-4 py-3.5 whitespace-nowrap">
                        <Badge variant="outline" className="text-xs">
                          {p.harga_layanan?.brand || getBrandName(p.id_brand)}
                        </Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs whitespace-nowrap">
                        {p.tgl_instalasi
                          ? new Date(p.tgl_instalasi).toLocaleDateString('id-ID', {
                              day: '2-digit',
                              month: 'short',
                              year: 'numeric',
                            })
                          : '-'}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-right whitespace-nowrap">
                        {canEdit && (
                          <Button size="icon" variant="ghost" onClick={() => openEdit(p)} title="Edit">
                            <Edit className="h-4 w-4" />
                          </Button>
                        )}
                        {canDelete && (
                          <Button
                            size="icon"
                            variant="ghost"
                            className="text-destructive hover:text-destructive"
                            onClick={() => setDeleteTarget(p)}
                            title="Hapus"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}

          {/* Pagination */}
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

      {/* Add/Edit Dialog (Wide 2-Step Form matching existing) */}
      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="sm:max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingId ? 'Edit Data Pelanggan' : 'Tambah Pelanggan Baru'}</DialogTitle>
            <DialogDescription>
              Lengkapi data pribadi serta informasi alamat dan layanan pelanggan.
            </DialogDescription>
          </DialogHeader>

          {/* Stepper Indicator */}
          <div className="flex items-center gap-3 rounded-lg border bg-muted/30 px-4 py-3">
            <button
              type="button"
              onClick={() => setFormStep(1)}
              className="flex items-center gap-2 text-sm font-medium"
            >
              <span
                className={`h-6 w-6 rounded-full text-xs font-bold flex items-center justify-center ${
                  formStep === 1 ? 'bg-primary text-primary-foreground' : 'bg-emerald-600 text-white'
                }`}
              >
                1
              </span>
              <span className={formStep === 1 ? 'text-foreground' : 'text-muted-foreground'}>Info Pribadi</span>
            </button>
            <div className="h-px flex-1 bg-border" />
            <button
              type="button"
              onClick={() => {
                if (!validateStep1()) {
                  setFormError('')
                  setFormStep(2)
                } else {
                  handleNextStep()
                }
              }}
              className="flex items-center gap-2 text-sm font-medium"
            >
              <span
                className={`h-6 w-6 rounded-full text-xs font-bold flex items-center justify-center ${
                  formStep === 2 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'
                }`}
              >
                2
              </span>
              <span className={formStep === 2 ? 'text-foreground' : 'text-muted-foreground'}>Alamat & Layanan</span>
            </button>
          </div>

          <form onSubmit={handleSave} className="space-y-5">
            {formError && (
              <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {formError}
              </div>
            )}

            {formStep === 1 ? (
              <div className="rounded-lg border p-5 space-y-4">
                <div className="text-center space-y-0.5">
                  <h3 className="text-base font-semibold">Informasi Pribadi</h3>
                  <p className="text-xs text-muted-foreground">Masukkan data pribadi pelanggan dengan lengkap</p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="nama">
                      Nama Lengkap <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="nama"
                      value={form.nama || ''}
                      onChange={(e) => setForm({ ...form, nama: e.target.value })}
                      placeholder="Masukkan nama lengkap"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="no_ktp">
                      Nomor KTP <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="no_ktp"
                      maxLength={16}
                      value={form.no_ktp || ''}
                      onChange={(e) => setForm({ ...form, no_ktp: e.target.value })}
                      placeholder="16 digit NIK"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="email">
                      Email <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="email"
                      type="email"
                      value={form.email || ''}
                      onChange={(e) => setForm({ ...form, email: e.target.value })}
                      placeholder="email@contoh.com"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="no_telp">
                      Nomor Telepon / WA <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="no_telp"
                      value={form.no_telp || ''}
                      onChange={(e) => setForm({ ...form, no_telp: e.target.value })}
                      placeholder="08xxxxxxxxxx"
                    />
                  </div>
                </div>
              </div>
            ) : (
              <div className="rounded-lg border p-5 space-y-4">
                <div className="text-center space-y-0.5">
                  <h3 className="text-base font-semibold">Alamat & Layanan</h3>
                  <p className="text-xs text-muted-foreground">Pilih lokasi pemasangan, paket layanan, dan brand</p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-1.5">
                    <Label>
                      Brand <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.id_brand || ''}
                      onValueChange={(v) => setForm({ ...form, id_brand: v })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Pilih Brand" />
                      </SelectTrigger>
                      <SelectContent>
                        {brands.map((b) => (
                          <SelectItem key={b.id_brand} value={b.id_brand}>
                            {b.brand}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label>
                      Layanan <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.layanan || ''}
                      onValueChange={(v) => setForm({ ...form, layanan: v })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Pilih Layanan" />
                      </SelectTrigger>
                      <SelectContent>
                        {LAYANAN_OPTIONS.map((l) => (
                          <SelectItem key={l} value={l}>
                            {l}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label>
                      Alamat / Lokasi <span className="text-destructive">*</span>
                    </Label>
                    <Select
                      value={form.alamat || ''}
                      onValueChange={(v) => setForm({ ...form, alamat: v })}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Pilih Lokasi" />
                      </SelectTrigger>
                      <SelectContent>
                        {formAlamatOptions.map((a) => (
                          <SelectItem key={a} value={a}>
                            {a}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="tgl_instalasi">Tanggal Instalasi</Label>
                    <Input
                      id="tgl_instalasi"
                      type="date"
                      value={form.tgl_instalasi || ''}
                      onChange={(e) => setForm({ ...form, tgl_instalasi: e.target.value })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="blok">
                      Blok / Tower <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="blok"
                      value={form.blok || ''}
                      onChange={(e) => setForm({ ...form, blok: e.target.value })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="unit">
                      Unit / No. Rumah <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="unit"
                      value={form.unit || ''}
                      onChange={(e) => setForm({ ...form, unit: e.target.value })}
                    />
                  </div>
                  <div className="space-y-1.5 sm:col-span-2">
                    <Label htmlFor="alamat_custom">Alamat Detail / Catatan</Label>
                    <Input
                      id="alamat_custom"
                      value={form.alamat_custom || ''}
                      onChange={(e) => setForm({ ...form, alamat_custom: e.target.value })}
                    />
                  </div>
                </div>
              </div>
            )}

            <DialogFooter>
              {formStep === 1 ? (
                <>
                  <Button key="btn-cancel" type="button" variant="outline" onClick={() => setFormOpen(false)}>
                    Batal
                  </Button>
                  <Button key="btn-next" type="button" onClick={handleNextStep}>
                    Lanjut
                  </Button>
                </>
              ) : (
                <>
                  <Button key="btn-back" type="button" variant="outline" onClick={() => setFormStep(1)}>
                    Kembali
                  </Button>
                  <Button key="btn-save" type="submit" disabled={saving}>
                    {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {saving ? 'Menyimpan & Memuat Data...' : 'Simpan'}
                  </Button>
                </>
              )}
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Import CSV Dialog */}
      <Dialog open={importOpen} onOpenChange={setImportOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import Data Pelanggan (CSV)</DialogTitle>
            <DialogDescription>Pilih file CSV untuk mengimpor data pelanggan secara massal.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleImport} className="space-y-4">
            <div className="flex items-center justify-between rounded-md border bg-muted/40 px-3 py-2 text-xs">
              <span className="text-muted-foreground">Belum punya format CSV?</span>
              <Button type="button" variant="outline" size="sm" onClick={() => void handleDownloadTemplate()}>
                <Download className="mr-1.5 h-3.5 w-3.5" />
                Download Template
              </Button>
            </div>
            <Input
              type="file"
              accept=".csv"
              onChange={(e) => setImportFile(e.target.files?.[0] || null)}
              required
            />
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
            <AlertDialogTitle>Hapus Pelanggan</AlertDialogTitle>
            <AlertDialogDescription>
              Yakin ingin menghapus data pelanggan "{deleteTarget?.nama}"?
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
