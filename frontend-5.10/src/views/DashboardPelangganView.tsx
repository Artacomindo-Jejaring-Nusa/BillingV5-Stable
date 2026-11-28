import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
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
  CheckCircle2,
  RefreshCw,
  TrendingUp,
  UserCheck,
  UserMinus,
  UserPlus,
  Users,
  Wallet,
  Wifi,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface JakinetStats {
  pelanggan_aktif: number
  pelanggan_baru_bulan_ini: number
  pelanggan_berhenti_bulan_ini: number
  pelanggan_jakinet_aktif: number
  pendapatan_jakinet_bulan_ini?: number
}

interface ChartSeries {
  labels: string[]
  data: number[]
}

const DEFAULT_STATS: JakinetStats = {
  pelanggan_aktif: 0,
  pelanggan_baru_bulan_ini: 0,
  pelanggan_berhenti_bulan_ini: 0,
  pelanggan_jakinet_aktif: 0,
  pendapatan_jakinet_bulan_ini: 0,
}

function formatCurrency(value?: number) {
  if (typeof value !== 'number' || Number.isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value)
}

export function DashboardPelangganView() {
  const [timespan, setTimespan] = useState<'6m' | '1y' | 'all'>('6m')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [stats, setStats] = useState<JakinetStats>(DEFAULT_STATS)
  const [growthChart, setGrowthChart] = useState<ChartSeries>({ labels: [], data: [] })
  const [revenueChart, setRevenueChart] = useState<ChartSeries>({ labels: [], data: [] })

  const fetchData = useCallback(async (span: string) => {
    setLoading(true)
    setError('')
    try {
      const response = await apiClient.get('/dashboard-pelanggan', { params: { timespan: span } })
      const d = response.data || {}
      const m = d.main_stats || {}
      setStats({
        pelanggan_aktif: Number(m.pelanggan_aktif ?? 0),
        pelanggan_baru_bulan_ini: Number(m.pelanggan_baru_bulan_ini ?? 0),
        pelanggan_berhenti_bulan_ini: Number(m.pelanggan_berhenti_bulan_ini ?? 0),
        pelanggan_jakinet_aktif: Number(m.pelanggan_jakinet_aktif ?? 0),
        pendapatan_jakinet_bulan_ini: Number(m.pendapatan_jakinet_bulan_ini ?? 0),
      })
      setGrowthChart({
        labels: Array.isArray(d.growth_chart?.labels) ? d.growth_chart.labels : [],
        data: Array.isArray(d.growth_chart?.data) ? d.growth_chart.data : [],
      })
      setRevenueChart({
        labels: Array.isArray(d.revenue_chart?.labels) ? d.revenue_chart.labels : [],
        data: Array.isArray(d.revenue_chart?.data) ? d.revenue_chart.data : [],
      })
    } catch (err) {
      console.error('Failed to load dashboard pelanggan:', err)
      setError('Gagal memuat data dashboard pelanggan JakiNet.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchData(timespan)
  }, [fetchData, timespan])

  const jakiNetPercentage = useMemo(() => {
    if (!stats.pelanggan_aktif) return '0.0'
    return ((stats.pelanggan_jakinet_aktif / stats.pelanggan_aktif) * 100).toFixed(1)
  }, [stats.pelanggan_aktif, stats.pelanggan_jakinet_aktif])

  const averageRevenuePerUser = useMemo(() => {
    const rev = stats.pendapatan_jakinet_bulan_ini || 0
    if (!stats.pelanggan_jakinet_aktif || !rev) return 0
    return Math.round(rev / stats.pelanggan_jakinet_aktif)
  }, [stats.pendapatan_jakinet_bulan_ini, stats.pelanggan_jakinet_aktif])

  const pieData = useMemo(() => {
    const activeExisting = Math.max(0, stats.pelanggan_jakinet_aktif - stats.pelanggan_baru_bulan_ini)
    return [
      { name: 'Pelanggan Aktif (Lama)', value: activeExisting, color: '#0f172a' },
      { name: 'Pelanggan Baru (Bulan Ini)', value: stats.pelanggan_baru_bulan_ini, color: '#059669' },
      { name: 'Pelanggan Berhenti (Bulan Ini)', value: stats.pelanggan_berhenti_bulan_ini, color: '#e11d48' },
    ]
  }, [stats])

  const mutationBarData = useMemo(
    () => [
      { name: 'Baru (Bulan Ini)', value: stats.pelanggan_baru_bulan_ini, fill: '#059669' },
      { name: 'Berhenti (Bulan Ini)', value: stats.pelanggan_berhenti_bulan_ini, fill: '#e11d48' },
    ],
    [stats.pelanggan_baru_bulan_ini, stats.pelanggan_berhenti_bulan_ini]
  )

  const growthRows = useMemo(
    () => (growthChart?.labels || []).map((name, i) => ({ name, value: growthChart?.data?.[i] ?? 0 })),
    [growthChart]
  )

  const revenueRows = useMemo(
    () => (revenueChart?.labels || []).map((name, i) => ({ name, value: revenueChart?.data?.[i] ?? 0 })),
    [revenueChart]
  )

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Dashboard Monitoring Pelanggan JakiNet</h2>
          <p className="text-sm text-muted-foreground">
            Pantau performa, pertumbuhan, dan aktivitas finansial pelanggan JakiNet secara real-time
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="secondary" className="gap-1.5 text-emerald-700 dark:text-emerald-300">
            <span className="h-2 w-2 rounded-full bg-emerald-500" />
            Sistem Real-time
          </Badge>
          <div className="inline-flex rounded-md border bg-card p-0.5">
            {(['6m', '1y', 'all'] as const).map((span) => (
              <Button
                key={span}
                size="sm"
                variant={timespan === span ? 'default' : 'ghost'}
                onClick={() => setTimespan(span)}
                className="h-8 text-xs"
              >
                {span === '6m' ? '6 Bulan' : span === '1y' ? '1 Tahun' : 'Semua'}
              </Button>
            ))}
          </div>
          <Button variant="outline" size="sm" onClick={() => void fetchData(timespan)} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Segarkan Data
          </Button>
        </div>
      </div>

      {error && (
        <Card className="border-destructive/50">
          <CardContent className="pt-6 flex items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm text-destructive">
              <AlertCircle className="h-4 w-4" />
              <span>{error}</span>
            </div>
            <Button size="sm" variant="outline" onClick={() => void fetchData(timespan)}>
              Coba Lagi
            </Button>
          </CardContent>
        </Card>
      )}

      {/* 4 Compact Stat Cards */}
      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-xl" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground uppercase">
                <span>Pelanggan Aktif</span>
                <UserCheck className="h-4 w-4 text-primary" />
              </div>
              <div className="text-2xl font-bold mt-1">{stats.pelanggan_aktif.toLocaleString('id-ID')}</div>
              <div className="text-xs text-muted-foreground mt-1">Total seluruh pelanggan berstatus aktif</div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground uppercase">
                <span>Pelanggan Baru</span>
                <UserPlus className="h-4 w-4 text-emerald-600" />
              </div>
              <div className="text-2xl font-bold text-emerald-600 mt-1">
                +{stats.pelanggan_baru_bulan_ini.toLocaleString('id-ID')}
              </div>
              <div className="text-xs text-muted-foreground mt-1">Bulan ini</div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground uppercase">
                <span>Berhenti</span>
                <UserMinus className="h-4 w-4 text-destructive" />
              </div>
              <div className="text-2xl font-bold text-destructive mt-1">
                {stats.pelanggan_berhenti_bulan_ini.toLocaleString('id-ID')}
              </div>
              <div className="text-xs text-muted-foreground mt-1">Bulan ini</div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground uppercase">
                <span>JakiNet Aktif</span>
                <Wifi className="h-4 w-4 text-sky-600" />
              </div>
              <div className="text-2xl font-bold text-sky-600 mt-1">
                {stats.pelanggan_jakinet_aktif.toLocaleString('id-ID')}
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                <strong>{jakiNetPercentage}%</strong> dari total pelanggan
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Revenue Hero Card */}
      {!loading && (
        <Card className="bg-card">
          <CardContent className="p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="space-y-1">
              <div className="text-[11px] font-semibold tracking-wider text-primary uppercase">
                Financial Performance
              </div>
              <div className="text-lg font-semibold">Pendapatan JakiNet</div>
              <p className="text-xs text-muted-foreground">
                Estimasi total tagihan dan penerimaan pelanggan JakiNet bulan berjalan
              </p>
              <div className="flex items-baseline gap-2 pt-1">
                <span className="text-3xl font-bold tracking-tight">
                  {formatCurrency(stats.pendapatan_jakinet_bulan_ini)}
                </span>
                <span className="text-xs text-muted-foreground">/ bulan ini</span>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              <div className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs">
                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                <span>
                  Penetrasi JakiNet: <strong>{jakiNetPercentage}%</strong>
                </span>
              </div>
              <div className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs">
                <Wallet className="h-3.5 w-3.5 text-primary" />
                <span>
                  ARPU: <strong>{formatCurrency(averageRevenuePerUser)}</strong> / user
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Users className="h-4 w-4 text-primary" />
              Distribusi Pelanggan JakiNet
            </CardTitle>
            <CardDescription>Komposisi aktif, baru, dan berhenti</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={pieData} dataKey="value" nameKey="name" innerRadius={52} outerRadius={84} paddingAngle={2}>
                    {pieData.map((item) => (
                      <Cell key={item.name} fill={item.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend wrapperStyle={{ fontSize: 11 }} />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-primary" />
              Tren Pertumbuhan Kumulatif
            </CardTitle>
            <CardDescription>Grafik akumulasi pelanggan JakiNet</CardDescription>
          </CardHeader>
          <CardContent className="h-72">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : growthRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data pertumbuhan pelanggan JakiNet.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={growthRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Line type="monotone" dataKey="value" name="Pelanggan" stroke="#059669" strokeWidth={2.5} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Pelanggan Baru vs Berhenti</CardTitle>
            <CardDescription>Perbandingan mutasi bulan berjalan</CardDescription>
          </CardHeader>
          <CardContent className="h-64">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={mutationBarData}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip />
                  <Bar dataKey="value" name="Jumlah" radius={[4, 4, 0, 0]}>
                    {mutationBarData.map((entry) => (
                      <Cell key={entry.name} fill={entry.fill} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Tren Pendapatan JakiNet</CardTitle>
            <CardDescription>Total pendapatan per periode</CardDescription>
          </CardHeader>
          <CardContent className="h-64">
            {loading ? (
              <Skeleton className="h-full w-full" />
            ) : revenueRows.length === 0 ? (
              <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                Belum ada data pendapatan.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={revenueRows}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 11 }} />
                  <Tooltip formatter={(val) => formatCurrency(Number(val))} />
                  <Bar dataKey="value" name="Pendapatan" fill="#2563eb" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
