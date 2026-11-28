import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Ban,
  CheckCircle2,
  Copy,
  Download,
  ExternalLink,
  Eye,
  Filter,
  Loader2,
  MessageCircle,
  Plus,
  RefreshCw,
  RotateCcw,
  Search,
  Trash2,
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

export interface InvoiceItem {
  id: number
  invoice_number: string
  pelanggan_id: number
  langganan_id?: number
  pelanggan_nama?: string
  pelanggan_email?: string
  pelanggan_no_telp?: string
  id_pelanggan?: string
  pelanggan?: {
    id: number
    nama: string
    customer_id?: string
    email?: string
    no_telp?: string
    alamat?: string
  }
  total_harga: number
  status_invoice: string
  tgl_invoice?: string
  created_at?: string
  tgl_jatuh_tempo?: string
  paid_at?: string | null
  metode_pembayaran?: string | null
  payment_link?: string | null
  invoice_type?: string | null
  brand?: string | null
}

interface LanggananOption {
  id: number
  pelanggan_id: number
  nama_pelanggan?: string
  nama_paket?: string
  pelanggan?: {
    nama: string
    alamat?: string
  }
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
  return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
}

function getDueHint(tgl?: string | null, status?: string) {
  if (!tgl || status === 'Lunas') return null
  const diff = Math.ceil((new Date(tgl).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return `Terlambat ${Math.abs(diff)} hari`
  if (diff === 0) return 'Jatuh tempo hari ini'
  return `${diff} hari lagi`
}

export function InvoicesView() {
  const [list, setList] = useState<InvoiceItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [initialLoading, setInitialLoading] = useState(true)
  const [fetching, setFetching] = useState(false)

  // Filters & Pagination
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreateInvoice = hasPermission('create_invoices')
  const canEditInvoice = hasPermission('edit_invoices')
  const canDeleteInvoice = hasPermission('delete_invoices')
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  const [statusFilter, setStatusFilter] = useState('all')
  const [xenditFilter, setXenditFilter] = useState('all')
  const [showPaidInvoices, setShowPaidInvoices] = useState(false)
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)

  // Modals
  const [detailTarget, setDetailTarget] = useState<InvoiceItem | null>(null)
  const [markPaidTarget, setMarkPaidTarget] = useState<InvoiceItem | null>(null)
  const [paymentMethod, setPaymentMethod] = useState<'Cash' | 'Transfer Bank'>('Cash')
  const [markingPaid, setMarkingPaid] = useState(false)

  const [expireTarget, setExpireTarget] = useState<InvoiceItem | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<InvoiceItem | null>(null)
  const [reinvoiceTarget, setReinvoiceTarget] = useState<InvoiceItem | null>(null)

  // Buat Invoice Manual Modal
  const [generateOpen, setGenerateOpen] = useState(false)
  const [langgananList, setLanggananList] = useState<LanggananOption[]>([])
  const [langgananSearch, setLanggananSearch] = useState('')
  const [selectedLanggananId, setSelectedLanggananId] = useState<number | null>(null)
  const [generating, setGenerating] = useState(false)

  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    const t = setTimeout(() => {
      setDebouncedSearch(searchQuery)
      setPage(1)
    }, 500)
    return () => clearTimeout(t)
  }, [searchQuery])

  const fetchInvoices = useCallback(async () => {
    setFetching(true)
    try {
      const params = new URLSearchParams()
      if (debouncedSearch) params.append('search', debouncedSearch)
      if (statusFilter !== 'all') {
        params.append('status_invoice', statusFilter)
      } else if (!showPaidInvoices) {
        params.append('status_invoice', 'Belum Bayar')
      }
      if (xenditFilter !== 'all') params.append('xendit_status', xenditFilter)
      if (startDate) params.append('start_date', startDate)
      if (endDate) params.append('end_date', endDate)
      params.append('skip', String((page - 1) * limit))
      params.append('limit', String(limit))

      const [invRes, countRes] = await Promise.allSettled([
        apiClient.get(`/invoices?${params.toString()}`),
        apiClient.get(`/invoices/count?${params.toString()}`),
      ])

      if (invRes.status === 'fulfilled') {
        const raw = invRes.value.data?.data ?? invRes.value.data
        setList(Array.isArray(raw) ? raw : [])
        if (invRes.value.data?.total_count !== undefined) {
          setTotalCount(invRes.value.data.total_count)
        }
      }
      if (countRes.status === 'fulfilled' && countRes.value.data?.count !== undefined) {
        setTotalCount(countRes.value.data.count)
      }
    } catch {
      setFeedback({ type: 'error', text: 'Gagal memuat data invoice.' })
    } finally {
      setInitialLoading(false)
      setFetching(false)
    }
  }, [debouncedSearch, statusFilter, xenditFilter, showPaidInvoices, startDate, endDate, page, limit])

  useEffect(() => {
    void fetchInvoices()
  }, [fetchInvoices])

  const stats = useMemo(() => {
    const lunas = list.filter((i) => i.status_invoice === 'Lunas').length
    const belumBayar = list.filter((i) => i.status_invoice === 'Belum Bayar').length
    const expired = list.filter((i) => i.status_invoice === 'Expired').length
    return { lunas, belumBayar, expired }
  }, [list])

  const openGenerateModal = async () => {
    setLanggananSearch('')
    setSelectedLanggananId(null)
    setGenerateOpen(true)
    if (langgananList.length === 0) {
      try {
        const res = await apiClient.get('/langganan?for_invoice_selection=true&limit=5000')
        const raw = res.data?.data ?? res.data
        setLanggananList(Array.isArray(raw) ? raw : [])
      } catch {
        setLanggananList([])
      }
    }
  }

  const filteredLangganan = useMemo(() => {
    const q = langgananSearch.trim().toLowerCase()
    if (q.length < 3) return []
    return langgananList
      .filter((l) => {
        const name = (l.nama_pelanggan || l.pelanggan?.nama || '').toLowerCase()
        return name.includes(q)
      })
      .slice(0, 20)
  }, [langgananList, langgananSearch])

  const handleGenerateManual = async () => {
    if (!selectedLanggananId || generating) return
    setGenerating(true)
    try {
      await apiClient.post('/invoices/generate', { langganan_id: selectedLanggananId })
      await fetchInvoices()
      setGenerateOpen(false)
      setFeedback({ type: 'success', text: 'Invoice manual berhasil dibuat.' })
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { detail?: string; error?: string } } })?.response?.data?.detail ||
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Gagal membuat invoice manual.'
      setFeedback({ type: 'error', text: msg })
    } finally {
      setGenerating(false)
    }
  }

  const copyPaymentLink = async (inv: InvoiceItem) => {
    if (!inv.payment_link) {
      setFeedback({ type: 'error', text: 'Link pembayaran belum tersedia untuk invoice ini.' })
      return
    }
    await navigator.clipboard.writeText(inv.payment_link)
    setFeedback({ type: 'success', text: `Link pembayaran ${inv.invoice_number} disalin.` })
  }

  const handleResendWa = async (inv: InvoiceItem) => {
    try {
      setFeedback({ type: 'success', text: 'Sedang mengirim notifikasi WhatsApp...' })
      const res = await apiClient.post(`/invoices/${inv.id}/resend-wa`)
      setFeedback({
        type: 'success',
        text: res.data?.message || `Notifikasi WhatsApp untuk ${inv.invoice_number} berhasil dikirim ulang!`,
      })
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string; detail?: string } } })?.response?.data?.error ||
        (err as { response?: { data?: { detail?: string } } })?.response?.data?.detail ||
        'Gagal mengirim ulang notifikasi WhatsApp.'
      setFeedback({ type: 'error', text: msg })
    }
  }

  const confirmMarkAsPaid = async () => {
    if (!markPaidTarget) return
    setMarkingPaid(true)
    try {
      await apiClient.post(`/invoices/${markPaidTarget.id}/mark-as-paid`, {
        metode_pembayaran: paymentMethod,
      })
      await fetchInvoices()
      setMarkPaidTarget(null)
      setFeedback({ type: 'success', text: `Invoice ditandai Lunas (${paymentMethod}).` })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menandai invoice lunas.' })
    } finally {
      setMarkingPaid(false)
    }
  }

  const confirmExpire = async () => {
    if (!expireTarget) return
    try {
      await apiClient.post(`/invoices/${expireTarget.id}/expire`)
      await fetchInvoices()
      setExpireTarget(null)
      setFeedback({ type: 'success', text: 'Invoice berhasil diubah menjadi Expired.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengubah status invoice ke Expired.' })
      setExpireTarget(null)
    }
  }

  const confirmReinvoice = async () => {
    if (!reinvoiceTarget) return
    try {
      if (reinvoiceTarget.payment_link) {
        await apiClient.post(`/invoices/create_reinvoice/${reinvoiceTarget.id}`)
      } else {
        await apiClient.post(`/invoices/${reinvoiceTarget.id}/retry-xendit`)
      }
      await fetchInvoices()
      setReinvoiceTarget(null)
      setFeedback({ type: 'success', text: 'Invoice berhasil diterbitkan ulang.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menerbitkan ulang invoice.' })
      setReinvoiceTarget(null)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await apiClient.delete(`/invoices/${deleteTarget.id}`)
      await fetchInvoices()
      setDeleteTarget(null)
      setFeedback({ type: 'success', text: 'Invoice berhasil dihapus.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menghapus invoice.' })
      setDeleteTarget(null)
    }
  }

  const handleExportComprehensive = async () => {
    try {
      const params: Record<string, string> = {}
      if (debouncedSearch) params.search = debouncedSearch
      if (statusFilter !== 'all') params.status_invoice = statusFilter
      if (startDate) params.start_date = startDate
      if (endDate) params.end_date = endDate

      const res = await apiClient.get('/invoices/export-payment-links-excel', {
        params,
        responseType: 'blob',
      })
      const url = URL.createObjectURL(
        new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
      )
      const a = document.createElement('a')
      a.href = url
      a.download = `payment-links-${new Date().toISOString().split('T')[0]}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengekspor laporan invoice.' })
    }
  }

  const statusBadge = (status: string) => {
    if (status === 'Lunas') return <Badge variant="secondary" className="text-emerald-700 dark:text-emerald-300">Lunas</Badge>
    if (status === 'Expired') return <Badge variant="destructive">Expired</Badge>
    return <Badge className="bg-amber-500/15 text-amber-700 dark:text-amber-300 hover:bg-amber-500/20">{status}</Badge>
  }

  const totalPages = Math.max(1, Math.ceil(totalCount / limit))

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Invoices</h2>
          <p className="text-sm text-muted-foreground">Kelola tagihan dan pembayaran pelanggan</p>
        </div>
        {canCreateInvoice && (
          <Button size="sm" onClick={() => void openGenerateModal()}>
            <Plus className="mr-1.5 h-4 w-4" />
            Buat Invoice Manual
          </Button>
        )}
      </div>

      {/* Compact 4-Column Summary Strip */}
      <Card>
        <CardContent className="p-0">
          <div className="grid grid-cols-2 lg:grid-cols-4 divide-y lg:divide-y-0 sm:divide-x">
            <div className="px-4 py-3">
              <div className="text-lg font-bold text-emerald-600">{stats.lunas}</div>
              <div className="text-xs text-muted-foreground">Invoice Lunas (halaman ini)</div>
            </div>
            <div className="px-4 py-3">
              <div className="text-lg font-bold text-amber-600">{stats.belumBayar}</div>
              <div className="text-xs text-muted-foreground">Belum Bayar (halaman ini)</div>
            </div>
            <div className="px-4 py-3">
              <div className="text-lg font-bold text-destructive">{stats.expired}</div>
              <div className="text-xs text-muted-foreground">Expired (halaman ini)</div>
            </div>
            <div className="px-4 py-3">
              <div className="text-lg font-bold">
                {list.length} / {totalCount.toLocaleString('id-ID')}
              </div>
              <div className="text-xs text-muted-foreground">Total Invoice</div>
            </div>
          </div>
        </CardContent>
      </Card>

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

      {/* Search & Filter */}
      <Card>
        <CardContent className="pt-5 space-y-3">
          <div className="flex flex-col sm:flex-row gap-2">
            <div className="relative flex-1">
              <Search className="h-4 w-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari (No. Invoice, Nama, ID)..."
                className="pl-9"
              />
            </div>
            <Button variant="outline" onClick={() => setShowFilters((v) => !v)}>
              <Filter className="mr-1.5 h-4 w-4" />
              Filter
            </Button>
            <Button variant="outline" onClick={() => void fetchInvoices()} disabled={fetching}>
              <RefreshCw className={`h-4 w-4 ${fetching ? 'animate-spin' : ''}`} />
            </Button>
          </div>

          {showFilters && (
            <div className="space-y-4 pt-3 border-t">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Status Invoice
                  </Label>
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
                      <SelectItem value="Belum Bayar">Belum Bayar</SelectItem>
                      <SelectItem value="Lunas">Lunas</SelectItem>
                      <SelectItem value="Expired">Expired</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-1.5">
                  <Label className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Status Xendit
                  </Label>
                  <Select
                    value={xenditFilter}
                    onValueChange={(v) => {
                      setXenditFilter(v)
                      setPage(1)
                    }}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Semua Status Xendit" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">Semua</SelectItem>
                      <SelectItem value="pending">Pending</SelectItem>
                      <SelectItem value="paid">Paid</SelectItem>
                      <SelectItem value="expired">Expired</SelectItem>
                      <SelectItem value="failed">Failed</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-1.5">
                  <Label className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Tampilkan
                  </Label>
                  <Select
                    value={String(limit)}
                    onValueChange={(v) => {
                      setLimit(Number(v))
                      setPage(1)
                    }}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {[10, 15, 25, 50, 100].map((n) => (
                        <SelectItem key={n} value={String(n)}>
                          {n} item
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-1.5">
                <Label className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                  Rentang Tanggal
                </Label>
                <div className="grid grid-cols-1 sm:grid-cols-[1fr_auto_1fr] items-center gap-2">
                  <Input
                    type="date"
                    aria-label="Dari tanggal"
                    value={startDate}
                    onChange={(e) => {
                      setStartDate(e.target.value)
                      setPage(1)
                    }}
                  />
                  <span className="text-muted-foreground text-center hidden sm:inline">→</span>
                  <Input
                    type="date"
                    aria-label="Sampai tanggal"
                    value={endDate}
                    onChange={(e) => {
                      setEndDate(e.target.value)
                      setPage(1)
                    }}
                  />
                </div>
              </div>

              <div className="flex flex-wrap items-center justify-between gap-3 pt-1">
                <label className="inline-flex items-center gap-2.5 cursor-pointer select-none text-sm">
                  <input
                    type="checkbox"
                    checked={showPaidInvoices}
                    onChange={(e) => {
                      setShowPaidInvoices(e.target.checked)
                      setPage(1)
                    }}
                    className="h-4 w-4 rounded border-input"
                  />
                  <span>Tampilkan Lunas & Expired</span>
                </label>

                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setSearchQuery('')
                    setStatusFilter('all')
                    setXenditFilter('all')
                    setShowPaidInvoices(false)
                    setStartDate('')
                    setEndDate('')
                    setLimit(10)
                    setPage(1)
                  }}
                >
                  <RefreshCw className="mr-1.5 h-3.5 w-3.5" />
                  Reset Semua Filter
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Daftar Tagihan */}
      <Card className="overflow-hidden">
        {fetching && !initialLoading && (
          <div className="h-1 w-full bg-primary/20 overflow-hidden">
            <div className="h-full w-1/3 bg-primary animate-pulse" />
          </div>
        )}
        <CardContent className="p-0">
          <div className="px-4 py-3 border-b flex flex-wrap items-center justify-between gap-2">
            <div>
              <div className="font-semibold text-sm">Daftar Tagihan</div>
              <div className="text-xs text-muted-foreground">Kelola dan pantau status pembayaran</div>
            </div>
            <Button size="sm" variant="outline" onClick={() => void handleExportComprehensive()}>
              <Download className="mr-1.5 h-4 w-4" />
              Export Laporan Komprehensif
            </Button>
          </div>

          {initialLoading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : list.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">
              Belum ada data invoice yang ditemukan.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Nomor Invoice</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Pelanggan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Total Tagihan</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Status</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Jatuh Tempo</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((inv) => {
                    const nama = inv.pelanggan?.nama || inv.pelanggan_nama || `Pelanggan #${inv.pelanggan_id}`
                    const idPel = inv.id_pelanggan || inv.pelanggan?.customer_id || '-'
                    const dueHint = getDueHint(inv.tgl_jatuh_tempo, inv.status_invoice)
                    return (
                      <TableRow key={inv.id}>
                        <TableCell className="px-4 py-3.5">
                          <div className="font-mono text-xs font-semibold max-w-[240px] truncate" title={inv.invoice_number}>
                            {inv.invoice_number}
                          </div>
                          <div className="text-[11px] text-muted-foreground">
                            {formatDate(inv.tgl_invoice || inv.created_at)}
                          </div>
                        </TableCell>
                        <TableCell className="px-4 py-3.5">
                          <div className="font-medium text-sm">{nama}</div>
                          <div className="text-xs text-muted-foreground font-mono">ID: {idPel}</div>
                        </TableCell>
                        <TableCell className="px-4 py-3.5 font-semibold text-emerald-600 whitespace-nowrap">
                          {formatCurrency(inv.total_harga)}
                        </TableCell>
                        <TableCell className="px-4 py-3.5 whitespace-nowrap">
                          {statusBadge(inv.status_invoice)}
                        </TableCell>
                        <TableCell className="px-4 py-3.5 whitespace-nowrap">
                          <div className="text-xs font-medium">{formatDate(inv.tgl_jatuh_tempo)}</div>
                          {dueHint && <div className="text-[11px] text-amber-600">{dueHint}</div>}
                        </TableCell>
                        <TableCell className="px-4 py-3.5 text-right whitespace-nowrap">
                          <div className="inline-flex items-center gap-0.5">
                            <Button
                              size="icon"
                              variant="ghost"
                              className="h-8 w-8"
                              onClick={() => void copyPaymentLink(inv)}
                              title="Salin Link Pembayaran"
                            >
                              <Copy className="h-4 w-4" />
                            </Button>
                            <Button
                              size="icon"
                              variant="ghost"
                              className="h-8 w-8"
                              onClick={() => setDetailTarget(inv)}
                              title="Lihat Detail"
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                            {canDeleteInvoice && (
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-destructive hover:text-destructive"
                                onClick={() => setDeleteTarget(inv)}
                                title="Hapus"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            )}
                            {canEditInvoice && inv.status_invoice !== 'Lunas' && (
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-emerald-600 hover:text-emerald-700"
                                onClick={() => {
                                  setPaymentMethod('Cash')
                                  setMarkPaidTarget(inv)
                                }}
                                title="Tandai Lunas"
                              >
                                <CheckCircle2 className="h-4 w-4" />
                              </Button>
                            )}
                            {canEditInvoice && inv.status_invoice === 'Belum Bayar' && (
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-amber-600 hover:text-amber-700"
                                onClick={() => setExpireTarget(inv)}
                                title="Jadikan Expired"
                              >
                                <Ban className="h-4 w-4" />
                              </Button>
                            )}
                            {canEditInvoice && inv.status_invoice !== 'Lunas' && (
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-emerald-600 hover:text-emerald-700"
                                onClick={() => void handleResendWa(inv)}
                                title="Kirim Ulang WA"
                              >
                                <MessageCircle className="h-4 w-4" />
                              </Button>
                            )}
                            {canEditInvoice && inv.status_invoice !== 'Lunas' && (
                              <Button
                                size="icon"
                                variant="ghost"
                                className="h-8 w-8 text-sky-600 hover:text-sky-700"
                                onClick={() => setReinvoiceTarget(inv)}
                                title="Terbitkan Ulang / Retry"
                              >
                                <RotateCcw className="h-4 w-4" />
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
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
                  Halaman {page} dari {totalPages} ({totalCount.toLocaleString('id-ID')} invoice)
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

      {/* Buat Invoice Manual Modal */}
      <Dialog open={generateOpen} onOpenChange={setGenerateOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Buat Invoice Manual</DialogTitle>
            <DialogDescription>Pilih langganan pelanggan untuk membuat invoice baru</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="rounded-md bg-sky-500/10 border border-sky-500/25 px-3 py-2 text-xs text-sky-700 dark:text-sky-300">
              Pilih langganan pelanggan yang ingin dibuatkan invoice-nya sekarang.
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="search-langganan">Pilih Langganan Pelanggan</Label>
              <Input
                id="search-langganan"
                value={langgananSearch}
                onChange={(e) => {
                  setLanggananSearch(e.target.value)
                  setSelectedLanggananId(null)
                }}
                placeholder="Ketik minimal 3 huruf nama pelanggan..."
              />
            </div>

            {langgananSearch.trim().length < 3 ? (
              <div className="rounded-md border border-dashed p-3 text-xs text-muted-foreground text-center">
                Ketik minimal 3 karakter...
              </div>
            ) : filteredLangganan.length === 0 ? (
              <div className="rounded-md border border-dashed p-3 text-xs text-muted-foreground text-center">
                Langganan tidak ditemukan.
              </div>
            ) : (
              <div className="max-h-52 overflow-y-auto rounded-md border divide-y">
                {filteredLangganan.map((l) => {
                  const nama = l.nama_pelanggan || l.pelanggan?.nama || `Langganan #${l.id}`
                  const isSelected = selectedLanggananId === l.id
                  return (
                    <button
                      key={l.id}
                      type="button"
                      onClick={() => setSelectedLanggananId(l.id)}
                      className={`w-full text-left px-3 py-2.5 text-xs flex items-center justify-between hover:bg-accent ${
                        isSelected ? 'bg-primary/10 font-semibold' : ''
                      }`}
                    >
                      <span>{nama}</span>
                      <span className="text-muted-foreground">{l.nama_paket || ''}</span>
                    </button>
                  )
                })}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGenerateOpen(false)}>
              Batal
            </Button>
            <Button onClick={() => void handleGenerateManual()} disabled={!selectedLanggananId || generating}>
              {generating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              + Buat Invoice
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Tandai Lunas Modal (Cash / Transfer Bank) */}
      <Dialog open={!!markPaidTarget} onOpenChange={(o) => !o && setMarkPaidTarget(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Tandai Invoice Lunas</DialogTitle>
            <DialogDescription>
              Pilih metode pembayaran untuk invoice{' '}
              <span className="font-mono font-semibold">{markPaidTarget?.invoice_number}</span>
            </DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-2 gap-3 py-2">
            {(['Cash', 'Transfer Bank'] as const).map((method) => (
              <button
                key={method}
                type="button"
                onClick={() => setPaymentMethod(method)}
                className={`rounded-lg border p-3.5 text-sm font-medium text-center transition-colors ${
                  paymentMethod === method
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'hover:bg-accent'
                }`}
              >
                {method}
              </button>
            ))}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setMarkPaidTarget(null)}>
              Batal
            </Button>
            <Button onClick={() => void confirmMarkAsPaid()} disabled={markingPaid}>
              {markingPaid && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Konfirmasi Lunas ({paymentMethod})
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Detail Invoice Modal */}
      <Dialog open={!!detailTarget} onOpenChange={(o) => !o && setDetailTarget(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Detail Invoice</DialogTitle>
            <DialogDescription className="font-mono text-xs">{detailTarget?.invoice_number}</DialogDescription>
          </DialogHeader>
          {detailTarget && (
            <div className="space-y-2.5 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Pelanggan</span>
                <span className="font-medium">
                  {detailTarget.pelanggan?.nama || detailTarget.pelanggan_nama || '-'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Total Tagihan</span>
                <span className="font-bold text-emerald-600">{formatCurrency(detailTarget.total_harga)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Status</span>
                {statusBadge(detailTarget.status_invoice)}
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Jatuh Tempo</span>
                <span>{formatDate(detailTarget.tgl_jatuh_tempo)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Metode Pembayaran</span>
                <span>{detailTarget.metode_pembayaran || '-'}</span>
              </div>
              {detailTarget.payment_link && (
                <div className="pt-2 border-t flex items-center justify-between gap-2">
                  <span className="text-xs text-muted-foreground truncate">{detailTarget.payment_link}</span>
                  <Button size="sm" variant="outline" asChild>
                    <a href={detailTarget.payment_link} target="_blank" rel="noreferrer">
                      <ExternalLink className="mr-1.5 h-3.5 w-3.5" />
                      Buka
                    </a>
                  </Button>
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            <Button onClick={() => setDetailTarget(null)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Expire Confirm */}
      <AlertDialog open={!!expireTarget} onOpenChange={(o) => !o && setExpireTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Jadikan Invoice Expired?</AlertDialogTitle>
            <AlertDialogDescription>
              Invoice {expireTarget?.invoice_number} akan diubah statusnya menjadi Expired.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction onClick={confirmExpire}>Ya, Jadikan Expired</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Reinvoice / Retry Confirm */}
      <AlertDialog open={!!reinvoiceTarget} onOpenChange={(o) => !o && setReinvoiceTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Terbitkan Ulang Invoice?</AlertDialogTitle>
            <AlertDialogDescription>
              Sistem akan menerbitkan ulang / me-retry link pembayaran untuk {reinvoiceTarget?.invoice_number}.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Batal</AlertDialogCancel>
            <AlertDialogAction onClick={confirmReinvoice}>Ya, Terbitkan Ulang</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Confirm */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Invoice</AlertDialogTitle>
            <AlertDialogDescription>
              Yakin ingin menghapus invoice {deleteTarget?.invoice_number}?
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
