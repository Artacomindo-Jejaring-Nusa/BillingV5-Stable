import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Brain,
  ChartLine,
  Download,
  FileText,
  Lightbulb,
  ListFilter,
  Play,
  RefreshCw,
  Search,
  TrendingUp,
  Users,
  Wallet,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface BillStat {
  count: number
  nominal: number
  diskon: number
  biaya_pasang: number
  total: number
}

interface TaxStat {
  ppn: number
  bhp: number
  uso: number
  total_pajak: number
}

interface PaymentMethodStat {
  method: string
  count: number
  total_amount: number
  pajak: number
  diskon: number
}

interface InvoiceReportItem {
  id?: number
  invoice_number: string
  pelanggan_nama: string
  brand?: string
  alamat?: string
  status_invoice?: string
  tgl_invoice?: string
  tgl_lunas?: string | null
  metode?: string | null
  total_harga: number
}

interface RevenueReportResponse {
  total_pendapatan: number
  total_invoices: number
  financial_summary: {
    total_pemasukan: number
    total_pengeluaran: number
    saldo_akhir: number
  }
  billing_summary: {
    total_tagihan: BillStat
    lunas: BillStat
    pending: BillStat
    expired: BillStat
  }
  tax_summary: {
    total: TaxStat
    lunas: TaxStat
    pending: TaxStat
    expired: TaxStat
  }
  payment_methods: PaymentMethodStat[]
  rincian_invoice?: InvoiceReportItem[]
}

interface MLClusterSummary {
  key: string
  name: string
  count: number
  percentage: number
  total_nominal: number
  color: string
  risk_level: string
}

interface MLCashFlowPhase {
  phase: string
  description: string
  estimated_amount: number
  percentage: number
  target_days: string
  status_class: string
}

interface MLClassifiedCustomer {
  customer_id: number
  customer_name: string
  no_telp: string
  brand: string
  monthly_bill: number
  cluster_key: string
  cluster_name: string
  cluster_color: string
  risk_score: number
  risk_level: string
  avg_payment_day: number
  days_to_pay_avg: number
  paid_ratio: number
  recommendation: string
}

interface MLRevenueInsightResponse {
  status: string
  cycle_info: {
    invoice_issuance_date: string
    due_date: string
    grace_period_cutoff: string
    total_active_subscribers: number
  }
  summary: {
    total_customers: number
    projected_billing: number
    projected_early_collection: number
    projected_grace_collection: number
    projected_at_risk: number
    projected_recovery_rate: number
  }
  clusters_summary: MLClusterSummary[]
  cash_flow_forecast: MLCashFlowPhase[]
  classified_customers: MLClassifiedCustomer[]
  actionable_insights: string[]
}

function formatCurrency(val?: number) {
  if (typeof val !== 'number' || Number.isNaN(val)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(val)
}

function formatShortDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(`${dateStr}T00:00:00`).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function defaultDateRange() {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const d = String(now.getDate()).padStart(2, '0')
  return { start: `${y}-${m}-01`, end: `${y}-${m}-${d}` }
}

function statusBadgeClass(status?: string) {
  switch (status) {
    case 'Lunas':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800'
    case 'Belum Dibayar':
    case 'Pending':
      return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800'
    case 'Expired':
      return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

export function RevenueReportView() {
  const initial = defaultDateRange()
  const [activeTab, setActiveTab] = useState('summary')
  const [startDate, setStartDate] = useState(initial.start)
  const [endDate, setEndDate] = useState(initial.end)
  const [selectedLocation, setSelectedLocation] = useState('all')
  const [selectedBrand, setSelectedBrand] = useState('all')
  const [detailStatusFilter, setDetailStatusFilter] = useState('all')

  const [locations, setLocations] = useState<string[]>([])
  const [brands, setBrands] = useState<Array<{ id: string; nama: string }>>([])

  const [reportSummary, setReportSummary] = useState<RevenueReportResponse | null>(null)
  const [invoiceDetails, setInvoiceDetails] = useState<InvoiceReportItem[]>([])
  const [mlInsights, setMlInsights] = useState<MLRevenueInsightResponse | null>(null)

  const [mlClusterFilter, setMlClusterFilter] = useState('all')
  const [mlSearch, setMlSearch] = useState('')

  const [loadingReport, setLoadingReport] = useState(false)
  const [loadingDetails, setLoadingDetails] = useState(false)
  const [loadingML, setLoadingML] = useState(false)
  const [page, setPage] = useState(1)
  const pageSize = 15

  const fetchFilterOptions = useCallback(async () => {
    const [locRes, brandRes] = await Promise.allSettled([
      apiClient.get('/pelanggan/lokasi/unik'),
      apiClient.get('/harga_layanan'),
    ])
    if (locRes.status === 'fulfilled' && Array.isArray(locRes.value.data)) {
      setLocations(locRes.value.data)
    }
    if (brandRes.status === 'fulfilled') {
      const list = Array.isArray(brandRes.value.data) ? brandRes.value.data : brandRes.value.data?.data || []
      setBrands(
        list.map((b: { id_brand?: string; id?: string; brand?: string; nama_brand?: string }) => ({
          id: String(b.id_brand ?? b.id ?? ''),
          nama: b.brand || b.nama_brand || String(b.id_brand ?? ''),
        })),
      )
    }
  }, [])

  const fetchDetailsPage = useCallback(
    async (targetPage = 1) => {
      setLoadingDetails(true)
      try {
        const baseParams = {
          start_date: startDate,
          end_date: endDate,
          ...(selectedLocation !== 'all' && { alamat: selectedLocation }),
          ...(selectedBrand !== 'all' && { id_brand: selectedBrand }),
        }
        const detRes = await apiClient.get<InvoiceReportItem[]>('/reports/revenue/details', {
          params: { ...baseParams, skip: (targetPage - 1) * pageSize, limit: pageSize },
        })
        setInvoiceDetails(Array.isArray(detRes.data) ? detRes.data : [])
        setPage(targetPage)
      } catch (err) {
        console.error('Fetch revenue details failed:', err)
      } finally {
        setLoadingDetails(false)
      }
    },
    [startDate, endDate, selectedLocation, selectedBrand],
  )

  const fetchReport = useCallback(
    async (targetPage = 1) => {
      setLoadingReport(true)
      try {
        const baseParams = {
          start_date: startDate,
          end_date: endDate,
          ...(selectedLocation !== 'all' && { alamat: selectedLocation }),
          ...(selectedBrand !== 'all' && { id_brand: selectedBrand }),
        }

        const [sumRes, detRes] = await Promise.all([
          apiClient.get<RevenueReportResponse>('/reports/revenue', { params: baseParams }),
          apiClient.get<InvoiceReportItem[]>('/reports/revenue/details', {
            params: { ...baseParams, skip: (targetPage - 1) * pageSize, limit: pageSize },
          }),
        ])
        setReportSummary(sumRes.data)
        setInvoiceDetails(Array.isArray(detRes.data) ? detRes.data : [])
        setPage(targetPage)
      } catch (err) {
        console.error('Fetch revenue report failed:', err)
      } finally {
        setLoadingReport(false)
      }
    },
    [startDate, endDate, selectedLocation, selectedBrand],
  )

  const fetchMLInsights = useCallback(async () => {
    setLoadingML(true)
    try {
      const params = {
        ...(selectedLocation !== 'all' && { location: selectedLocation }),
        ...(selectedBrand !== 'all' && { brand: selectedBrand }),
      }
      const res = await apiClient.get<MLRevenueInsightResponse>('/reports/revenue/ml-insights', { params })
      setMlInsights(res.data)
    } catch (err) {
      console.error('Fetch ML insights failed:', err)
    } finally {
      setLoadingML(false)
    }
  }, [selectedLocation, selectedBrand])

  useEffect(() => {
    void fetchFilterOptions()
    void fetchReport(1)
  }, [fetchFilterOptions, fetchReport])

  useEffect(() => {
    if (activeTab === 'ml' && !mlInsights) {
      void fetchMLInsights()
    }
  }, [activeTab, mlInsights, fetchMLInsights])

  const filteredInvoiceDetails = useMemo(() => {
    if (detailStatusFilter === 'all') return invoiceDetails
    return invoiceDetails.filter((item) => item.status_invoice === detailStatusFilter)
  }, [invoiceDetails, detailStatusFilter])

  const filteredClassifiedCustomers = useMemo(() => {
    if (!mlInsights?.classified_customers) return []
    let list = mlInsights.classified_customers
    if (mlClusterFilter !== 'all') {
      list = list.filter((c) => c.cluster_key === mlClusterFilter)
    }
    if (mlSearch.trim()) {
      const q = mlSearch.toLowerCase()
      list = list.filter(
        (c) =>
          c.customer_name.toLowerCase().includes(q) ||
          (c.no_telp && c.no_telp.toLowerCase().includes(q)),
      )
    }
    return list
  }, [mlInsights, mlClusterFilter, mlSearch])

  const exportCsv = async () => {
    try {
      const baseParams = {
        start_date: startDate,
        end_date: endDate,
        ...(selectedLocation !== 'all' && { alamat: selectedLocation }),
        ...(selectedBrand !== 'all' && { id_brand: selectedBrand }),
      }
      const res = await apiClient.get<InvoiceReportItem[]>('/reports/revenue/details', { params: baseParams })
      const allData = Array.isArray(res.data) ? res.data : invoiceDetails
      if (!allData.length) return

      const headers = ['Nomor Invoice', 'Nama Pelanggan', 'Brand', 'Alamat', 'Status', 'Tanggal Bayar', 'Metode', 'Jumlah']
      const rows = allData.map((item) => [
        item.invoice_number,
        item.pelanggan_nama,
        item.brand || '',
        item.alamat || '',
        item.status_invoice || '',
        item.tgl_lunas ? new Date(item.tgl_lunas).toLocaleString('id-ID') : '-',
        item.metode || '-',
        String(item.total_harga),
      ])
      const csv = [headers, ...rows].map((r) => r.map((v) => `"${String(v).replace(/"/g, '""')}"`).join(',')).join('\n')
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `Laporan_Keuangan_${startDate}_${endDate}.csv`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      console.error('Export CSV failed:', err)
    }
  }

  const bs = reportSummary?.billing_summary
  const belumTerbayarTotal = (bs?.pending?.total || 0) + (bs?.expired?.total || 0)

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Laporan Keuangan</h2>
        <p className="text-sm text-muted-foreground">Monitoring detail pendapatan, tagihan, dan pajak</p>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 items-end">
            {activeTab === 'summary' && (
              <>
                <div className="space-y-1.5">
                  <Label htmlFor="start-date">Tanggal Awal</Label>
                  <Input id="start-date" type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="end-date">Tanggal Akhir</Label>
                  <Input id="end-date" type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
                </div>
              </>
            )}
            <div className="space-y-1.5">
              <Label>Wilayah</Label>
              <Select value={selectedLocation} onValueChange={setSelectedLocation}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Wilayah" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Wilayah</SelectItem>
                  {locations.map((loc) => (
                    <SelectItem key={loc} value={loc}>
                      {loc}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Brand</Label>
              <Select value={selectedBrand} onValueChange={setSelectedBrand}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Semua Brand" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Brand</SelectItem>
                  {brands.map((b) => (
                    <SelectItem key={b.id} value={b.id}>
                      {b.nama}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex gap-2">
              <Button
                className="flex-1"
                onClick={() => (activeTab === 'ml' ? void fetchMLInsights() : void fetchReport(1))}
                disabled={loadingReport || loadingML}
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${loadingReport || loadingML ? 'animate-spin' : ''}`} />
                {activeTab === 'ml' ? 'Analisis Ulang ML' : 'Tampilkan'}
              </Button>
              {activeTab === 'summary' && (
                <Button
                  variant="outline"
                  onClick={() => void exportCsv()}
                  disabled={!reportSummary || reportSummary.total_invoices === 0}
                  title="Export CSV"
                >
                  <Download className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="summary" className="gap-2">
            <ChartLine className="h-4 w-4" />
            Ringkasan Finansial & Tagihan
          </TabsTrigger>
          <TabsTrigger value="ml" className="gap-2">
            <Brain className="h-4 w-4" />
            Analisis Pola Pembayaran
            <Badge variant="secondary" className="ml-1 text-[10px]">
              ML Model
            </Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="summary" className="space-y-6 mt-4">
          {/* 1. Ringkasan Keuangan (3 Cards) */}
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-sm font-semibold">
              <Wallet className="h-4 w-4 text-muted-foreground" />
              <span>Ringkasan Keuangan</span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card className="border-t-4 border-t-emerald-500">
                <CardContent className="pt-5">
                  <p className="text-xs text-muted-foreground mb-1">Total Pemasukan</p>
                  {loadingReport ? (
                    <Skeleton className="h-8 w-40" />
                  ) : (
                    <p className="text-2xl font-bold text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(reportSummary?.financial_summary?.total_pemasukan || 0)}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground mt-2">Uang yang diterima sistem</p>
                </CardContent>
              </Card>

              <Card className="border-t-4 border-t-blue-600">
                <CardContent className="pt-5">
                  <p className="text-xs text-muted-foreground mb-1">Total Tagihan Dicetak</p>
                  {loadingReport ? (
                    <Skeleton className="h-8 w-40" />
                  ) : (
                    <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                      {formatCurrency(bs?.total_tagihan?.total || 0)}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground mt-2">Nilai invoice yang diterbitkan</p>
                </CardContent>
              </Card>

              <Card className="border-t-4 border-t-amber-500">
                <CardContent className="pt-5">
                  <p className="text-xs text-muted-foreground mb-1">Belum Terbayar</p>
                  {loadingReport ? (
                    <Skeleton className="h-8 w-40" />
                  ) : (
                    <p className="text-2xl font-bold text-amber-600 dark:text-amber-400">
                      {formatCurrency(belumTerbayarTotal)}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground mt-2">Potensi pendapatan (Pending + Expired)</p>
                </CardContent>
              </Card>
            </div>
          </div>

          {/* 2. Ringkasan Tagihan (4 Cards: Total Tagihan, Lunas, Pending, Expired) */}
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2 text-sm font-semibold">
              <FileText className="h-4 w-4 text-muted-foreground" />
              <span>Ringkasan Tagihan</span>
              <Badge variant="secondary" className="text-xs font-normal">
                {formatShortDate(startDate)} - {formatShortDate(endDate)}
              </Badge>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {[
                {
                  key: 'total',
                  label: 'Total Tagihan',
                  stat: bs?.total_tagihan,
                  titleCls: 'text-foreground',
                  countCls: 'text-foreground',
                  totalCls: 'text-foreground',
                },
                {
                  key: 'lunas',
                  label: 'Lunas',
                  stat: bs?.lunas,
                  titleCls: 'text-emerald-600 dark:text-emerald-400',
                  countCls: 'text-emerald-600 dark:text-emerald-400',
                  totalCls: 'text-emerald-600 dark:text-emerald-400',
                },
                {
                  key: 'pending',
                  label: 'Pending',
                  stat: bs?.pending,
                  titleCls: 'text-amber-600 dark:text-amber-400',
                  countCls: 'text-amber-600 dark:text-amber-400',
                  totalCls: 'text-amber-600 dark:text-amber-400',
                },
                {
                  key: 'expired',
                  label: 'Expired',
                  stat: bs?.expired,
                  titleCls: 'text-rose-600 dark:text-rose-400',
                  countCls: 'text-rose-600 dark:text-rose-400',
                  totalCls: 'text-rose-600 dark:text-rose-400',
                },
              ].map((card) => (
                <Card key={card.key}>
                  <CardContent className="pt-5 space-y-3">
                    <div>
                      <p className={`text-sm font-semibold ${card.titleCls}`}>{card.label}</p>
                      {loadingReport ? (
                        <Skeleton className="h-9 w-16 mt-1" />
                      ) : (
                        <p className={`text-3xl font-extrabold mt-1 ${card.countCls}`}>
                          {(card.stat?.count || 0).toLocaleString('id-ID')}
                        </p>
                      )}
                    </div>
                    <div className="space-y-1 text-xs text-muted-foreground">
                      <div className="flex justify-between">
                        <span>Nominal:</span>
                        <span className="font-medium text-foreground">{formatCurrency(card.stat?.nominal || 0)}</span>
                      </div>
                      <div className="flex justify-between text-rose-600 dark:text-rose-400">
                        <span>Diskon:</span>
                        <span className="font-medium">{formatCurrency(card.stat?.diskon || 0)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Biaya Pasang:</span>
                        <span className="font-medium text-foreground">{formatCurrency(card.stat?.biaya_pasang || 0)}</span>
                      </div>
                    </div>
                    <div className={`flex justify-between pt-2 border-t text-sm font-bold ${card.totalCls}`}>
                      <span>Total:</span>
                      <span>{formatCurrency(card.stat?.total || 0)}</span>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          {/* 3. Data Tagihan (Detail) */}
          <Card>
            <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pb-3">
              <CardTitle className="text-base">Data Tagihan (Detail)</CardTitle>
              <div className="w-full sm:w-48">
                <Select value={detailStatusFilter} onValueChange={setDetailStatusFilter}>
                  <SelectTrigger className="w-full h-8 text-xs">
                    <SelectValue placeholder="Semua Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua Status</SelectItem>
                    <SelectItem value="Lunas">Lunas</SelectItem>
                    <SelectItem value="Belum Dibayar">Belum Dibayar</SelectItem>
                    <SelectItem value="Expired">Expired</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardHeader>
            <CardContent>
              {loadingReport || loadingDetails ? (
                <div className="space-y-2">
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                </div>
              ) : filteredInvoiceDetails.length === 0 ? (
                <div className="py-10 text-center text-sm text-muted-foreground">
                  Tidak ada transaksi pada rentang tanggal ini.
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>No. Invoice</TableHead>
                          <TableHead>Pelanggan</TableHead>
                          <TableHead>Brand</TableHead>
                          <TableHead>Alamat</TableHead>
                          <TableHead>Status</TableHead>
                          <TableHead>Tgl Bayar</TableHead>
                          <TableHead>Metode</TableHead>
                          <TableHead className="text-right">Total</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {filteredInvoiceDetails.map((item, idx) => (
                          <TableRow key={`${item.invoice_number}-${idx}`}>
                            <TableCell className="font-mono text-xs font-medium">{item.invoice_number}</TableCell>
                            <TableCell className="font-medium">{item.pelanggan_nama}</TableCell>
                            <TableCell>
                              {item.brand ? (
                                <Badge variant="secondary" className="text-[11px]">
                                  {item.brand}
                                </Badge>
                              ) : (
                                '-'
                              )}
                            </TableCell>
                            <TableCell className="text-muted-foreground text-xs">{item.alamat || '-'}</TableCell>
                            <TableCell>
                              <Badge variant="outline" className={`text-[11px] ${statusBadgeClass(item.status_invoice)}`}>
                                {item.status_invoice || '-'}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-xs">
                              {item.tgl_lunas ? new Date(item.tgl_lunas).toLocaleString('id-ID') : '-'}
                            </TableCell>
                            <TableCell className="text-xs">{item.metode || '-'}</TableCell>
                            <TableCell className="text-right font-semibold">{formatCurrency(item.total_harga)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>
                      Halaman {page} (Total {reportSummary?.total_invoices || 0} tagihan)
                    </span>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={page <= 1 || loadingDetails}
                        onClick={() => void fetchDetailsPage(page - 1)}
                      >
                        Sebelumnya
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={invoiceDetails.length < pageSize || loadingDetails}
                        onClick={() => void fetchDetailsPage(page + 1)}
                      >
                        Berikutnya
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="ml" className="space-y-6 mt-4">
          {loadingML && !mlInsights ? (
            <div className="py-12 text-center space-y-3">
              <RefreshCw className="h-8 w-8 animate-spin mx-auto text-indigo-600" />
              <p className="text-sm text-muted-foreground">Memproses model Machine Learning & Clustering K-Means...</p>
            </div>
          ) : mlInsights ? (
            <>
              {/* 1. Cycle Info Banner */}
              <Card className="border-indigo-200 bg-indigo-50/70 dark:bg-indigo-950/30 dark:border-indigo-900">
                <CardContent className="pt-5 flex flex-wrap items-center justify-between gap-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-full bg-indigo-600 text-white">
                      <Brain className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="font-bold text-sm text-indigo-950 dark:text-indigo-100">
                        Analisis Pola Pembayaran & Prediksi Cash Flow (Model K-Means)
                      </p>
                      <p className="text-xs text-indigo-800/80 dark:text-indigo-200/80">
                        Siklus Penagihan: Invoice Terbit <strong>Tgl 26/27</strong> | Jatuh Tempo <strong>Tgl 01</strong> | Batas Masa Tenggang <strong>Tgl 10</strong> | Auto-Isolir <strong>Tgl 11 00:00 WIB</strong>
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge className="bg-indigo-600 text-white">
                      {mlInsights.cycle_info?.total_active_subscribers || 0} Pelanggan Teranalisis
                    </Badge>
                    <Button variant="outline" size="sm" onClick={() => void fetchMLInsights()} disabled={loadingML}>
                      <RefreshCw className={`mr-1.5 h-3.5 w-3.5 ${loadingML ? 'animate-spin' : ''}`} />
                      Refresh Model
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* 2. 4 Projection Metric Cards */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="border-t-4 border-t-blue-600">
                  <CardContent className="pt-5">
                    <p className="text-xs text-muted-foreground mb-1">Total Proyeksi Billing</p>
                    <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                      {formatCurrency(mlInsights.summary?.projected_billing || 0)}
                    </p>
                    <p className="text-xs text-muted-foreground mt-2">Target penagihan siklus bulan ini</p>
                  </CardContent>
                </Card>
                <Card className="border-t-4 border-t-emerald-500">
                  <CardContent className="pt-5">
                    <p className="text-xs text-muted-foreground mb-1">Puncak Awal (s/d Tgl 5)</p>
                    <p className="text-2xl font-bold text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(mlInsights.summary?.projected_early_collection || 0)}
                    </p>
                    <p className="text-xs text-muted-foreground mt-2">Estimasi masuk pasca terbit & awal bulan</p>
                  </CardContent>
                </Card>
                <Card className="border-t-4 border-t-amber-500">
                  <CardContent className="pt-5">
                    <p className="text-xs text-muted-foreground mb-1">Masa Tenggang (Tgl 6 - 10)</p>
                    <p className="text-2xl font-bold text-amber-600 dark:text-amber-400">
                      {formatCurrency(mlInsights.summary?.projected_grace_collection || 0)}
                    </p>
                    <p className="text-xs text-muted-foreground mt-2">Sebelum isolir otomatis tgl 11</p>
                  </CardContent>
                </Card>
                <Card className="border-t-4 border-t-rose-500">
                  <CardContent className="pt-5">
                    <p className="text-xs text-muted-foreground mb-1">Pendapatan Berisiko (Tgl 11+)</p>
                    <p className="text-2xl font-bold text-rose-600 dark:text-rose-400">
                      {formatCurrency(mlInsights.summary?.projected_at_risk || 0)}
                    </p>
                    <p className="text-xs text-muted-foreground mt-2">
                      Tingkat Recovery: <strong>{mlInsights.summary?.projected_recovery_rate || 0}%</strong>
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* 3. Fase Proyeksi Cash Flow */}
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm font-semibold">
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  <span>Proyeksi Cash Flow Pembayaran per Fase Siklus 26/27</span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  {(mlInsights.cash_flow_forecast || []).map((phase, idx) => (
                    <Card key={idx}>
                      <CardContent className="pt-5 space-y-2">
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-xs font-bold">{phase.phase}</span>
                          <Badge variant="secondary" className="text-[10px]">
                            {phase.target_days}
                          </Badge>
                        </div>
                        <p className="text-lg font-extrabold">{formatCurrency(phase.estimated_amount)}</p>
                        <div className="flex items-center justify-between text-xs text-muted-foreground">
                          <span>{phase.description}</span>
                          <span className="font-bold text-foreground">{phase.percentage}%</span>
                        </div>
                        <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                          <div
                            className={`h-full rounded-full ${
                              idx === 0
                                ? 'bg-teal-500'
                                : idx === 1
                                  ? 'bg-blue-600'
                                  : idx === 2
                                    ? 'bg-amber-500'
                                    : 'bg-rose-500'
                            }`}
                            style={{ width: `${Math.min(100, Math.max(0, phase.percentage))}%` }}
                          />
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </div>

              {/* 4. Klaster Perilaku Pelanggan */}
              <div className="space-y-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2 text-sm font-semibold">
                    <Users className="h-4 w-4 text-muted-foreground" />
                    <span>Klaster Perilaku Pelanggan (Segmentasi ML K-Means)</span>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    Klik kartu klaster untuk memfilter daftar pelanggan di bawah
                  </span>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  {(mlInsights.clusters_summary || []).map((cluster) => {
                    const active = mlClusterFilter === cluster.key
                    return (
                      <Card
                        key={cluster.key}
                        onClick={() => setMlClusterFilter(active ? 'all' : cluster.key)}
                        className="cursor-pointer transition-shadow hover:shadow-md"
                        style={active ? { borderColor: cluster.color, borderWidth: '2px' } : undefined}
                      >
                        <CardContent className="pt-5 space-y-2">
                          <div className="flex items-center justify-between">
                            <Badge style={{ backgroundColor: cluster.color }} className="text-white text-[10px]">
                              {cluster.risk_level} Risk
                            </Badge>
                            <span className="text-xs font-bold" style={{ color: cluster.color }}>
                              {cluster.percentage}% User
                            </span>
                          </div>
                          <p className="text-sm font-bold">{cluster.name}</p>
                          <p className="text-2xl font-extrabold">
                            {cluster.count}{' '}
                            <span className="text-xs font-normal text-muted-foreground">pelanggan</span>
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Total Nominal: <strong className="text-foreground">{formatCurrency(cluster.total_nominal)}</strong>
                          </p>
                        </CardContent>
                      </Card>
                    )
                  })}
                </div>
              </div>

              {/* 5. Rekomendasi Tindakan */}
              {mlInsights.actionable_insights && mlInsights.actionable_insights.length > 0 && (
                <Card className="bg-muted/40">
                  <CardContent className="pt-5">
                    <div className="flex items-center gap-2 text-sm font-bold mb-2">
                      <Lightbulb className="h-4 w-4 text-indigo-600" />
                      <span>Rekomendasi Tindakan Otomatis Sistem</span>
                    </div>
                    <ul className="list-disc pl-5 space-y-1 text-xs text-muted-foreground">
                      {mlInsights.actionable_insights.map((insight, idx) => (
                        <li key={idx}>{insight}</li>
                      ))}
                    </ul>
                  </CardContent>
                </Card>
              )}

              {/* 6. Tabel Klasifikasi Pelanggan */}
              <Card>
                <CardHeader className="flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <ListFilter className="h-4 w-4 text-muted-foreground" />
                    <CardTitle className="text-base">Daftar Klasifikasi Pelanggan & Pola Bayar</CardTitle>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Select value={mlClusterFilter} onValueChange={setMlClusterFilter}>
                      <SelectTrigger className="w-56">
                        <SelectValue placeholder="Filter Klaster" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">Semua Klaster</SelectItem>
                        <SelectItem value="early_ontime">Early Birds (Tepat Waktu)</SelectItem>
                        <SelectItem value="grace_period">Masa Tenggang (2-10)</SelectItem>
                        <SelectItem value="chronic_late">Terlambat Kronis (Rawan Isolir)</SelectItem>
                        <SelectItem value="at_risk_churn">At-Risk / Potensi Churn</SelectItem>
                      </SelectContent>
                    </Select>
                    <div className="relative w-56">
                      <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        value={mlSearch}
                        onChange={(e) => setMlSearch(e.target.value)}
                        placeholder="Cari nama / no. telp..."
                        className="pl-8"
                      />
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Pelanggan</TableHead>
                          <TableHead>Brand</TableHead>
                          <TableHead className="text-right">Tagihan Bulanan</TableHead>
                          <TableHead>Klaster Pola Bayar</TableHead>
                          <TableHead className="text-center">Rata-rata Tgl Bayar</TableHead>
                          <TableHead className="text-center">Waktu Bayar</TableHead>
                          <TableHead className="text-center">Skor Risiko</TableHead>
                          <TableHead>Rekomendasi Tindakan</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody className="text-xs">
                        {filteredClassifiedCustomers.length === 0 ? (
                          <TableRow>
                            <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                              Tidak ada data pelanggan untuk filter ini.
                            </TableCell>
                          </TableRow>
                        ) : (
                          filteredClassifiedCustomers.slice(0, 50).map((item) => (
                            <TableRow key={item.customer_id}>
                              <TableCell>
                                <div className="font-bold">{item.customer_name}</div>
                                <div className="text-muted-foreground">{item.no_telp || '-'}</div>
                              </TableCell>
                              <TableCell>{item.brand}</TableCell>
                              <TableCell className="text-right font-medium">{formatCurrency(item.monthly_bill)}</TableCell>
                              <TableCell>
                                <Badge variant="secondary" style={{ color: item.cluster_color }}>
                                  {item.cluster_name}
                                </Badge>
                              </TableCell>
                              <TableCell className="text-center">
                                {item.avg_payment_day > 0 ? `Tgl ${Math.round(item.avg_payment_day)}` : '-'}
                              </TableCell>
                              <TableCell className="text-center">
                                {item.days_to_pay_avg > 0 ? `${item.days_to_pay_avg} hari` : '-'}
                              </TableCell>
                              <TableCell className="text-center">
                                <Badge
                                  variant="outline"
                                  className={
                                    item.risk_score >= 70
                                      ? 'bg-rose-50 text-rose-700 border-rose-200'
                                      : item.risk_score >= 40
                                        ? 'bg-amber-50 text-amber-700 border-amber-200'
                                        : 'bg-emerald-50 text-emerald-700 border-emerald-200'
                                  }
                                >
                                  {item.risk_score} / 100
                                </Badge>
                              </TableCell>
                              <TableCell className="text-muted-foreground">{item.recommendation}</TableCell>
                            </TableRow>
                          ))
                        )}
                      </TableBody>
                    </Table>
                  </div>
                </CardContent>
              </Card>
            </>
          ) : (
            <Card>
              <CardContent className="py-12 text-center space-y-3">
                <Brain className="h-12 w-12 mx-auto text-muted-foreground" />
                <p className="text-sm text-muted-foreground">
                  Belum ada data analisis Machine Learning. Klik tombol di bawah untuk memproses.
                </p>
                <Button onClick={() => void fetchMLInsights()}>
                  <Play className="mr-2 h-4 w-4" />
                  Mulai Analisis Machine Learning
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
