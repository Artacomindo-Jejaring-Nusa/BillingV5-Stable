import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  AlertCircle,
  CalendarClock,
  FileCheck,
  RefreshCw,
  Server,
  TrendingUp,
  Tv,
  Users,
  Wallet,
} from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface RevenueBreakdown {
  brand: string
  revenue: number
}

interface RevenueSummary {
  total: number
  periode: string
  breakdown: RevenueBreakdown[]
}

interface StatCard {
  title: string
  value: number | string
}

interface LabelDataChart {
  labels: string[]
  data: number[]
}

interface InvoiceSummaryChart {
  labels: string[]
  total: number[]
  lunas: number[]
  menunggu: number[]
  expired: number[]
  otomatis: number[]
  manual: number[]
  reinvoice: number[]
}

interface DashboardPayload {
  revenue_summary?: RevenueSummary
  stat_cards?: StatCard[]
  lokasi_chart?: LabelDataChart
  paket_chart?: LabelDataChart
  growth_chart?: LabelDataChart
  invoice_summary_chart?: InvoiceSummaryChart
  status_langganan_chart?: LabelDataChart
  status_chart?: LabelDataChart
  alamat_chart?: LabelDataChart
  loyalitas_pembayaran_chart?: LabelDataChart
  loyalitas_chart?: LabelDataChart
}

interface InvoiceMonitorData {
  target_date: string
  status: string
  status_color: string
  total_should_have: number
  total_generated: number
  total_skipped: number
  success_rate: number
  message: string
  detail_url?: string
}

interface FutureProjectionData {
  target_date: string
  generation_date: string
  estimated_customers?: number
  system_status?: string
  generation_days_until?: number
  days_until?: number
  percentage_of_active?: number
}

interface PaketDetail {
  total_pelanggan: number
  breakdown_lokasi?: Array<{ nama: string; jumlah: number }>
}

const CHART_COLORS = ['#2563eb', '#16a34a', '#d97706', '#dc2626', '#7c3aed', '#0891b2', '#db2777']

function formatCurrency(value?: number): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value)
}

function formatDate(dateString?: string): string {
  if (!dateString) return '-'
  const d = new Date(dateString)
  if (Number.isNaN(d.getTime())) return dateString
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

function toChartRows(chart?: LabelDataChart) {
  if (!chart?.labels?.length) return []
  return chart.labels.map((name, idx) => ({
    name,
    value: chart.data?.[idx] ?? 0,
  }))
}

export function DashboardView() {
  const user = useAuthStore((s) => s.user)
  const roleName = (user?.role?.name || '').toLowerCase()
  const isAdminOrManager = ['superadmin', 'admin', 'manager', 'direktur'].includes(roleName)

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [data, setData] = useState<DashboardPayload | null>(null)
  const [invoiceStats, setInvoiceStats] = useState<StatCard[]>([])
  const [paketDetails, setPaketDetails] = useState<Record<string, PaketDetail>>({})
  const [selectedPaket, setSelectedPaket] = useState<{ title: string; detail: PaketDetail } | null>(null)

  const [invoiceMonitor, setInvoiceMonitor] = useState<InvoiceMonitorData | null>(null)
  const [futureProjection, setFutureProjection] = useState<FutureProjectionData | null>(null)
  const [loadingMonitors, setLoadingMonitors] = useState(false)

  const fetchDashboard = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [dashRes, invRes, paketRes] = await Promise.allSettled([
        apiClient.get<DashboardPayload>('/dashboard'),
        apiClient.get('/invoices/summary'),
        apiClient.get<Record<string, PaketDetail>>('/dashboard/paket-details'),
      ])

      if (dashRes.status === 'fulfilled') {
        setData(dashRes.value.data)
      } else {
        throw dashRes.reason
      }

      if (invRes.status === 'fulfilled' && invRes.value.data) {
        const s = invRes.value.data
        setInvoiceStats([
          { title: 'Invoice Otomatis', value: s.invoice_types?.automatic ?? 0 },
          { title: 'Invoice Manual', value: s.invoice_types?.manual ?? 0 },
          { title: 'Reinvoice', value: s.total_reinvoice ?? 0 },
        ])
      }

      if (paketRes.status === 'fulfilled' && paketRes.value.data) {
        setPaketDetails(paketRes.value.data)
      }
    } catch (err) {
      console.error('Failed to load dashboard:', err)
      setError('Gagal memuat data dashboard. Silakan coba muat ulang.')
    } finally {
      setLoading(false)
    }
  }, [])

  const fetchMonitors = useCallback(async () => {
    if (!isAdminOrManager) return
    setLoadingMonitors(true)
    try {
      const today = new Date()
      let targetYear = today.getFullYear()
      let targetMonth = today.getMonth() + 1
      if (targetMonth > 11) {
        targetMonth = 0
        targetYear++
      }
      const t1 = new Date(targetYear, targetMonth, 1)
      const genT1 = new Date(t1.getTime() - 5 * 24 * 60 * 60 * 1000)
      let finalTarget = t1
      if (today >= genT1) {
        let m2 = targetMonth + 1
        let y2 = targetYear
        if (m2 > 11) {
          m2 = 0
          y2++
        }
        finalTarget = new Date(y2, m2, 1)
      }
      const targetDateStr = `${finalTarget.getFullYear()}-${String(finalTarget.getMonth() + 1).padStart(2, '0')}-${String(finalTarget.getDate()).padStart(2, '0')}`

      const [monRes, projRes] = await Promise.allSettled([
        apiClient.get<InvoiceMonitorData>('/dashboard/invoice-generation-monitor'),
        apiClient.get<FutureProjectionData>(`/dashboard/future-invoice-projection?target_date=${targetDateStr}`),
      ])

      if (monRes.status === 'fulfilled') setInvoiceMonitor(monRes.value.data)
      if (projRes.status === 'fulfilled') setFutureProjection(projRes.value.data)
    } finally {
      setLoadingMonitors(false)
    }
  }, [isAdminOrManager])

  useEffect(() => {
    void fetchDashboard()
  }, [fetchDashboard])

  useEffect(() => {
    void fetchMonitors()
  }, [fetchMonitors])

  const allStats = useMemo(() => [...(data?.stat_cards || []), ...invoiceStats], [data?.stat_cards, invoiceStats])
  const customerStats = useMemo(() => allStats.filter((s) => s.title.toLowerCase().includes('pelanggan')), [allStats])
  const otherStats = useMemo(() => allStats.filter((s) => !s.title.toLowerCase().includes('pelanggan')), [allStats])

  const lokasiRows = useMemo(() => toChartRows(data?.lokasi_chart), [data?.lokasi_chart])
  const paketRows = useMemo(() => toChartRows(data?.paket_chart), [data?.paket_chart])
  const growthRows = useMemo(() => toChartRows(data?.growth_chart), [data?.growth_chart])
  const statusRows = useMemo(
    () => toChartRows(data?.status_langganan_chart || data?.status_chart),
    [data?.status_langganan_chart, data?.status_chart]
  )
  const alamatRows = useMemo(() => toChartRows(data?.alamat_chart), [data?.alamat_chart])
  const loyalitasRows = useMemo(
    () => toChartRows(data?.loyalitas_pembayaran_chart || data?.loyalitas_chart),
    [data?.loyalitas_pembayaran_chart, data?.loyalitas_chart]
  )

  const invoiceRows = useMemo(() => {
    const c = data?.invoice_summary_chart
    if (!c?.labels?.length) return []
    return c.labels.map((name, i) => ({
      name,
      total: c.total?.[i] ?? 0,
      lunas: c.lunas?.[i] ?? 0,
      menunggu: c.menunggu?.[i] ?? 0,
      expired: c.expired?.[i] ?? 0,
    }))
  }, [data?.invoice_summary_chart])

  if (error) {
    return (
      <div className="p-6">
        <Card className="max-w-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              Gagal Memuat Dashboard
            </CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={fetchDashboard}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Coba Lagi
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Dashboard</h2>
          <p className="text-sm text-muted-foreground">
            Monitoring Customer Fiber To The Home Artacom Portal Systems
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Button variant="outline" size="sm" asChild>
            <Link to="/tv-monitoring" target="_blank">
              <Tv className="mr-2 h-4 w-4" />
              Layar TV Monitoring
            </Link>
          </Button>
          <Button variant="outline" size="sm" onClick={() => { void fetchDashboard(); void fetchMonitors() }} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Segarkan
          </Button>
        </div>
      </div>

      {/* Top Grid: Piutang + Compact Customer & System KPI Strips */}
      {loading ? (
        <div className="space-y-3">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-3">
            <Skeleton className="h-28 lg:col-span-4 rounded-xl" />
            <Skeleton className="h-28 lg:col-span-8 rounded-xl" />
          </div>
          <Skeleton className="h-16 rounded-xl" />
        </div>
      ) : (
        <div className="space-y-3">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-3">
            {/* Piutang / Revenue Summary */}
            {data?.revenue_summary && (
              <Card className="lg:col-span-4 bg-slate-900 text-white dark:bg-slate-900/90 dark:border-slate-800">
                <CardContent className="p-4 flex flex-col justify-between h-full gap-2">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <div className="text-xs text-white/75 font-medium">
                        Piutang • Periode {data.revenue_summary.periode}
                      </div>
                      <div className="text-2xl font-bold tracking-tight mt-0.5">
                        {formatCurrency(data.revenue_summary.total)}
                      </div>
                    </div>
                    <Wallet className="h-5 w-5 opacity-80 shrink-0" />
                  </div>
                  {data.revenue_summary.breakdown?.length > 0 && (
                    <div className="flex flex-wrap gap-x-4 gap-y-1 pt-2 border-t border-white/15 text-xs">
                      {data.revenue_summary.breakdown.map((item) => (
                        <div key={item.brand} className="flex items-center gap-1.5">
                          <span className="opacity-75">{item.brand}:</span>
                          <span className="font-semibold">{formatCurrency(item.revenue)}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Customer Stats grouped in 1 compact card */}
            {customerStats.length > 0 && (
              <Card className={data?.revenue_summary ? 'lg:col-span-8' : 'lg:col-span-12'}>
                <CardContent className="p-0 h-full">
                  <div className="grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x h-full">
                    {customerStats.map((stat) => (
                      <div key={stat.title} className="p-4 flex flex-col justify-between">
                        <div className="flex items-center justify-between gap-2 text-xs font-medium text-muted-foreground">
                          <span className="truncate">{stat.title}</span>
                          <Users className="h-3.5 w-3.5 shrink-0" />
                        </div>
                        <div className="text-2xl font-bold mt-2">{stat.value}</div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>

          {/* Server & Invoice Type Stats in 1 compact horizontal strip */}
          {otherStats.length > 0 && (
            <Card>
              <CardContent className="p-0">
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 divide-y sm:divide-y-0 sm:divide-x">
                  {otherStats.map((stat) => (
                    <div key={stat.title} className="px-4 py-3 flex items-center justify-between gap-2">
                      <div className="min-w-0">
                        <div className="text-[11px] font-medium text-muted-foreground truncate">
                          {stat.title}
                        </div>
                        <div className="text-lg font-bold leading-tight mt-0.5">{stat.value}</div>
                      </div>
                      <Server className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Charts Row 1: Lokasi & Paket */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Pelanggan per Alamat</CardTitle>
            <CardDescription>Distribusi pelanggan aktif di setiap lokasi</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : lokasiRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data distribusi lokasi.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={lokasiRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Bar dataKey="value" name="Jumlah Pelanggan" fill="#2563eb" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Pelanggan per Paket</CardTitle>
            <CardDescription>Klik bar untuk melihat rincian lokasi per paket</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : paketRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data paket layanan.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={paketRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Bar
                    dataKey="value"
                    name="Jumlah Pelanggan"
                    fill="#16a34a"
                    radius={[4, 4, 0, 0]}
                    className="cursor-pointer"
                    onClick={(barData) => {
                      const label = barData?.name
                      if (label && paketDetails[label]) {
                        setSelectedPaket({ title: `Rincian Paket: ${label}`, detail: paketDetails[label] })
                      }
                    }}
                  />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts Row 2: Growth & Invoice Summary */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-primary" />
              Pertumbuhan Pelanggan Baru
            </CardTitle>
            <CardDescription>Jumlah pelanggan baru per bulan</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : growthRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data pertumbuhan pelanggan.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={growthRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Line type="monotone" dataKey="value" name="Pelanggan Baru" stroke="#2563eb" strokeWidth={2.5} dot={{ r: 4 }} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Ringkasan Invoice Bulanan</CardTitle>
            <CardDescription>Perbandingan status Lunas, Menunggu, dan Expired</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : invoiceRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data ringkasan invoice.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart data={invoiceRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Legend wrapperStyle={{ fontSize: 11 }} />
                  <Bar dataKey="lunas" name="Lunas" stackId="a" fill="#16a34a" />
                  <Bar dataKey="menunggu" name="Menunggu" stackId="a" fill="#f59e0b" />
                  <Bar dataKey="expired" name="Expired" stackId="a" fill="#dc2626" />
                  <Line type="monotone" dataKey="total" name="Total Invoice" stroke="#7c3aed" strokeWidth={2} />
                </ComposedChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts Row 3: Status, Loyalitas, Top 7 Alamat */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { title: 'Status Langganan', desc: 'Aktif, Suspended, Berhenti', rows: statusRows },
          { title: 'Loyalitas Pembayaran', desc: 'Segmen ketepatan bayar pelanggan', rows: loyalitasRows },
          { title: 'Top Lokasi Pelanggan', desc: 'Lokasi dengan pelanggan aktif terbanyak', rows: alamatRows },
        ].map((chart) => (
          <Card key={chart.title}>
            <CardHeader>
              <CardTitle className="text-base">{chart.title}</CardTitle>
              <CardDescription>{chart.desc}</CardDescription>
            </CardHeader>
            <CardContent className="h-64">
              {loading ? (
                <Skeleton className="h-full w-full" />
              ) : chart.rows.length === 0 ? (
                <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                  Belum ada data.
                </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={chart.rows} dataKey="value" nameKey="name" innerRadius={45} outerRadius={75} paddingAngle={2}>
                      {chart.rows.map((entry, idx) => (
                        <Cell key={entry.name} fill={CHART_COLORS[idx % CHART_COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Admin / Manager Invoice Monitors */}
      {isAdminOrManager && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {invoiceMonitor && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0">
                <div>
                  <CardTitle className="text-base flex items-center gap-2">
                    <FileCheck className="h-4 w-4 text-primary" />
                    Invoice Generation Monitor
                  </CardTitle>
                  <CardDescription>Target: {formatDate(invoiceMonitor.target_date)}</CardDescription>
                </div>
                <Badge variant={invoiceMonitor.total_skipped > 0 ? 'destructive' : 'secondary'}>
                  {invoiceMonitor.status}
                </Badge>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid grid-cols-4 gap-2 text-center">
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Seharusnya</div>
                    <div className="text-lg font-bold">{invoiceMonitor.total_should_have}</div>
                  </div>
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Berhasil</div>
                    <div className="text-lg font-bold text-emerald-600">{invoiceMonitor.total_generated}</div>
                  </div>
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Terlewat</div>
                    <div className="text-lg font-bold text-destructive">{invoiceMonitor.total_skipped}</div>
                  </div>
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Success</div>
                    <div className="text-lg font-bold">{invoiceMonitor.success_rate}%</div>
                  </div>
                </div>
                <p className="text-xs text-muted-foreground">{invoiceMonitor.message}</p>
              </CardContent>
            </Card>
          )}

          {futureProjection && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0">
                <div>
                  <CardTitle className="text-base flex items-center gap-2">
                    <CalendarClock className="h-4 w-4 text-primary" />
                    Proyeksi Generate Mendatang
                  </CardTitle>
                  <CardDescription>Target: {formatDate(futureProjection.target_date)}</CardDescription>
                </div>
                <Button variant="ghost" size="icon" onClick={fetchMonitors} disabled={loadingMonitors} aria-label="Segarkan proyeksi">
                  <RefreshCw className={`h-4 w-4 ${loadingMonitors ? 'animate-spin' : ''}`} />
                </Button>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid grid-cols-3 gap-2 text-center">
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Estimasi Pelanggan</div>
                    <div className="text-lg font-bold">{futureProjection.estimated_customers ?? 0}</div>
                  </div>
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Status Sistem</div>
                    <div className="text-lg font-bold">{futureProjection.system_status || 'Siap'}</div>
                  </div>
                  <div className="rounded-lg border p-2">
                    <div className="text-[11px] text-muted-foreground">Jadwal (H-5)</div>
                    <div className="text-xs font-semibold mt-1">{formatDate(futureProjection.generation_date)}</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Paket Detail Dialog */}
      <Dialog open={!!selectedPaket} onOpenChange={(open) => !open && setSelectedPaket(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{selectedPaket?.title}</DialogTitle>
            <DialogDescription>
              Total pelanggan: {selectedPaket?.detail.total_pelanggan ?? 0}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 max-h-72 overflow-y-auto">
            {(selectedPaket?.detail.breakdown_lokasi || []).map((loc) => (
              <div key={loc.nama} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                <span>{loc.nama}</span>
                <Badge variant="secondary">{loc.jumlah}</Badge>
              </div>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
