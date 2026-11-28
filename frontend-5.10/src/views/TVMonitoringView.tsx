import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  Activity,
  BarChart3,
  Filter,
  MapPin,
  Maximize2,
  Minimize2,
  RefreshCw,
  RotateCw,
  X,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

interface TVLocationStats {
  lokasi: string
  total_users: number
  active_users: number
  suspended_users: number
}

interface TVMonitoringData {
  total_users: number
  active_users: number
  suspended_users: number
  stopped_users: number
  locations: TVLocationStats[]
  active_location: string
  updated_at: string
}

const REFRESH_INTERVAL = 30
const CYCLE_INTERVAL = 15

export function TVMonitoringView() {
  const navigate = useNavigate()
  const [data, setData] = useState<TVMonitoringData | null>(null)
  const [loading, setLoading] = useState(false)
  const [isReconnecting, setIsReconnecting] = useState(false)
  const [selectedLocation, setSelectedLocation] = useState('Semua Lokasi')
  const [autoCycleEnabled, setAutoCycleEnabled] = useState(false)
  const [cycleCountdown, setCycleCountdown] = useState(CYCLE_INTERVAL)
  const [refreshCountdown, setRefreshCountdown] = useState(REFRESH_INTERVAL)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [now, setNow] = useState(() => new Date())

  const fetchData = useCallback(async (showLoading = false, locOverride?: string) => {
    const loc = locOverride ?? selectedLocation
    if (showLoading) setLoading(true)
    try {
      const locParam = loc === 'Semua Lokasi' ? '' : loc
      const url = `/dashboard/tv-monitoring${locParam ? `?lokasi=${encodeURIComponent(locParam)}` : ''}`
      const res = await apiClient.get<TVMonitoringData>(url, {
        headers: { 'X-Skip-Network-Interceptor': 'true' },
        timeout: 15000,
      })
      if (res?.data) {
        setData(res.data)
        setIsReconnecting(false)
      }
    } catch {
      setIsReconnecting(true)
    } finally {
      setLoading(false)
      setRefreshCountdown(REFRESH_INTERVAL)
    }
  }, [selectedLocation])

  useEffect(() => {
    void fetchData(true)
  }, [fetchData])

  // Live clock
  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  // Auto refresh countdown
  useEffect(() => {
    const timer = setInterval(() => {
      setRefreshCountdown((prev) => {
        if (prev <= 1) {
          void fetchData(false)
          return REFRESH_INTERVAL
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [fetchData])

  const availableLocations = useMemo(() => {
    const locs = (data?.locations || []).map((l) => l.lokasi).filter(Boolean)
    return ['Semua Lokasi', ...Array.from(new Set(locs))]
  }, [data?.locations])

  // Auto cycle locations
  useEffect(() => {
    if (!autoCycleEnabled || availableLocations.length <= 1) return
    const timer = setInterval(() => {
      setCycleCountdown((prev) => {
        if (prev <= 1) {
          const idx = availableLocations.indexOf(selectedLocation)
          const nextLoc = availableLocations[(idx + 1) % availableLocations.length]
          setSelectedLocation(nextLoc)
          void fetchData(false, nextLoc)
          return CYCLE_INTERVAL
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [autoCycleEnabled, availableLocations, selectedLocation, fetchData])

  const totalUsers = data?.total_users ?? 0
  const activeUsers = data?.active_users ?? 0
  const suspendedUsers = data?.suspended_users ?? 0

  const activePct = totalUsers > 0 ? ((activeUsers / totalUsers) * 100).toFixed(1) : '0.0'
  const suspendedPct = totalUsers > 0 ? ((suspendedUsers / totalUsers) * 100).toFixed(1) : '0.0'

  const chartLocations = useMemo(() => {
    const list = [...(data?.locations || [])]
    return list.sort((a, b) => a.total_users - b.total_users)
  }, [data?.locations])

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      void document.documentElement.requestFullscreen().then(() => setIsFullscreen(true))
    } else {
      void document.exitFullscreen().then(() => setIsFullscreen(false))
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 text-foreground flex flex-col">
      {/* Top Bar */}
      <header className="border-b bg-card px-4 py-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="h-9 w-9 rounded-lg bg-sky-500/10 text-sky-600 flex items-center justify-center">
            <Activity className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-sm sm:text-base font-bold tracking-tight">NOC & BILLING WALLBOARD</h1>
            <p className="text-[11px] text-muted-foreground">PT Artacomindo Jejaring Nusa</p>
          </div>
        </div>

        {/* Center Controls */}
        <div className="flex items-center gap-2 flex-wrap">
          <div className="inline-flex items-center gap-1.5 rounded-md border bg-background px-2.5 h-9 text-xs">
            <MapPin className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-muted-foreground">Lokasi:</span>
            <select
              value={selectedLocation}
              onChange={(e) => {
                const next = e.target.value
                setSelectedLocation(next)
                void fetchData(true, next)
              }}
              className="bg-transparent font-medium focus:outline-none"
            >
              {availableLocations.map((loc) => (
                <option key={loc} value={loc} className="bg-card text-foreground">
                  {loc}
                </option>
              ))}
            </select>
          </div>

          <Button
            size="sm"
            variant={autoCycleEnabled ? 'default' : 'outline'}
            onClick={() => {
              setAutoCycleEnabled((v) => !v)
              setCycleCountdown(CYCLE_INTERVAL)
            }}
            className="h-9 text-xs gap-1.5"
          >
            <RotateCw className={`h-3.5 w-3.5 ${autoCycleEnabled ? 'animate-spin' : ''}`} />
            Auto Rotate
            {autoCycleEnabled && <span className="opacity-80">{cycleCountdown}s</span>}
          </Button>
        </div>

        {/* Right Clock & Actions */}
        <div className="flex items-center gap-3">
          <div className="hidden sm:block text-right leading-tight">
            <div className="text-[11px] text-muted-foreground">
              {now.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })}
            </div>
            <div className="text-sm font-mono font-bold">
              {now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            </div>
          </div>

          <div className="flex items-center gap-1">
            <Button variant="outline" size="sm" onClick={() => void fetchData(true)} className="h-9 gap-1.5 text-xs">
              <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
              <span>{refreshCountdown}s</span>
            </Button>
            <Button variant="outline" size="icon" onClick={toggleFullscreen} className="h-9 w-9" aria-label="Fullscreen">
              {isFullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
            </Button>
            <Button variant="ghost" size="icon" onClick={() => navigate('/dashboard')} className="h-9 w-9" aria-label="Tutup">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 p-4 md:p-6 space-y-5">
        {selectedLocation !== 'Semua Lokasi' && (
          <div className="flex items-center justify-between rounded-lg border border-sky-500/30 bg-sky-500/10 px-4 py-2 text-xs">
            <div className="flex items-center gap-2">
              <Filter className="h-3.5 w-3.5 text-sky-600" />
              <span>
                Menampilkan data khusus wilayah: <strong>{selectedLocation}</strong>
              </span>
            </div>
            <Button
              size="sm"
              variant="ghost"
              className="h-7 text-xs"
              onClick={() => {
                setSelectedLocation('Semua Lokasi')
                void fetchData(true, 'Semua Lokasi')
              }}
            >
              Kembali ke Semua Lokasi
            </Button>
          </div>
        )}

        {/* Row 1: 3 Metric Cards */}
        <section className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card>
            <CardContent className="p-5 flex items-center justify-between gap-4">
              <div>
                <div className="text-xs font-semibold tracking-wider text-muted-foreground uppercase">
                  TOTAL KESELURUHAN
                </div>
                <div className="text-4xl font-bold text-primary mt-1">
                  {totalUsers.toLocaleString('id-ID')}
                </div>
                <div className="text-xs text-muted-foreground mt-2">
                  Pelanggan FTTH Berjalan (Aktif & Terisolir)
                </div>
              </div>
              <img src="/total-user.gif" alt="Total pelanggan" className="h-16 w-16 object-contain shrink-0" />
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5 flex items-center justify-between gap-4">
              <div>
                <div className="text-xs font-semibold tracking-wider text-emerald-600 uppercase">AKTIF</div>
                <div className="text-4xl font-bold text-emerald-600 mt-1">
                  {activeUsers.toLocaleString('id-ID')}
                </div>
                <div className="text-xs text-muted-foreground mt-2">
                  <strong>{activePct}%</strong> Pelanggan Online & Aktif
                </div>
              </div>
              <img src="/user-aktif.gif" alt="Pelanggan aktif" className="h-16 w-16 object-contain shrink-0" />
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-5 flex items-center justify-between gap-4">
              <div>
                <div className="text-xs font-semibold tracking-wider text-destructive uppercase">SUSPENDED</div>
                <div className="text-4xl font-bold text-destructive mt-1">
                  {suspendedUsers.toLocaleString('id-ID')}
                </div>
                <div className="text-xs text-muted-foreground mt-2">
                  <strong>{suspendedPct}%</strong> Modem Terisolir (Tunggakan)
                </div>
              </div>
              <img src="/prabowo-marah.gif" alt="Pelanggan suspended" className="h-16 w-16 object-contain shrink-0" />
            </CardContent>
          </Card>
        </section>

        {/* Row 2: Full-width Bar Chart Statistik Berdasarkan Lokasi */}
        <Card>
          <CardContent className="p-5 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2.5">
                <BarChart3 className="h-5 w-5 text-sky-600" />
                <div>
                  <h2 className="text-base font-semibold">Statistik Berdasarkan Lokasi</h2>
                  <p className="text-xs text-muted-foreground">
                    Urutan dari terkecil (kiri) ke terbesar (kanan) • Klik batang untuk filter wilayah
                  </p>
                </div>
              </div>
            </div>

            <div className="h-80 w-full">
              {chartLocations.length === 0 ? (
                <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
                  {loading ? 'Memuat data wallboard...' : 'Belum ada data statistik lokasi.'}
                </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={chartLocations}
                    onClick={(state) => {
                      const loc = state?.activeLabel
                      if (typeof loc === 'string' && loc) {
                        const next = selectedLocation === loc ? 'Semua Lokasi' : loc
                        setSelectedLocation(next)
                        void fetchData(true, next)
                      }
                    }}
                  >
                    <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                    <XAxis dataKey="lokasi" tick={{ fontSize: 11 }} />
                    <YAxis tick={{ fontSize: 11 }} />
                    <Tooltip />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <Bar dataKey="active_users" name="Aktif" stackId="a" fill="#0284c7" />
                    <Bar dataKey="suspended_users" name="Suspended" stackId="a" fill="#dc2626" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
      </main>

      {/* Footer */}
      <footer className="border-t bg-card px-4 py-2 text-xs text-muted-foreground flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className={`h-2 w-2 rounded-full ${isReconnecting ? 'bg-amber-500' : 'bg-emerald-500'}`} />
          <span>{isReconnecting ? 'Menghubungkan Ulang ke Server...' : 'Sistem Monitoring Aktif'}</span>
          {data?.updated_at && (
            <>
              <span>•</span>
              <span>
                Pembaruan Data: <strong>{new Date(data.updated_at).toLocaleTimeString('id-ID')}</strong>
              </span>
            </>
          )}
        </div>
        <div>
          Tekan <strong>F11</strong> untuk Fullscreen Layar TV
        </div>
      </footer>
    </div>
  )
}
