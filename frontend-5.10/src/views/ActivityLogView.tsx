import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import apiClient from '@/services/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
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
  History,
  Server,
  UserCheck,
  Search,
  FilterX,
  RefreshCw,
  Eye,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Info,
  AlertTriangle,
  AlertCircle,
  Copy,
  Check,
} from 'lucide-react'

interface UserOption {
  id: number
  name: string
  email: string
}

interface ActivityLogItem {
  id: number
  user_id?: number
  user?: {
    id: number
    name: string
    email: string
  } | null
  action: string
  details?: string | null
  timestamp?: string
  created_at?: string
}

interface SystemLogItem {
  id: number | string
  level: string
  module: string
  action: string
  message: string
  rawMessage: string
  timestamp: string
}

function parseSystemLogItem(raw: Record<string, unknown>): SystemLogItem {
  const ts = String(raw.timestamp || raw.created_at || new Date().toISOString())
  const rawMsg = String(raw.message || '').trim()
  let mod = String(raw.module || '').trim()
  let act = String(raw.action || '').trim()
  let displayMsg = rawMsg

  if (!mod && rawMsg.includes(':')) {
    const colonIdx = rawMsg.indexOf(':')
    const candidate = rawMsg.substring(0, colonIdx).trim()
    if (candidate.length > 0 && candidate.length < 35 && !candidate.includes('\n')) {
      mod = candidate
      displayMsg = rawMsg.substring(colonIdx + 1).trim()
    }
  }

  if (!act && displayMsg.includes(' - ')) {
    const dashIdx = displayMsg.indexOf(' - ')
    const candidate = displayMsg.substring(0, dashIdx).trim()
    if (candidate.length > 0 && candidate.length < 40) {
      act = candidate
    }
  }

  return {
    id: (raw.id as number | string) || `${ts}-${Math.random().toString(36).slice(2, 7)}`,
    level: String(raw.level || 'INFO').toUpperCase(),
    module: mod || 'System',
    action: act || '-',
    message: displayMsg || rawMsg || '-',
    rawMessage: rawMsg,
    timestamp: ts,
  }
}

function getActionBadgeClass(action: string): string {
  const a = (action || '').toUpperCase()
  if (a.includes('DELETE') || a.includes('HAPUS') || a.includes('REMOVE')) {
    return 'bg-rose-50 text-rose-700 border-rose-200'
  }
  if (a.includes('PATCH') || a.includes('PUT') || a.includes('UPDATE') || a.includes('EDIT')) {
    return 'bg-amber-50 text-amber-700 border-amber-200'
  }
  if (a.includes('POST') || a.includes('CREATE') || a.includes('TAMBAH') || a.includes('ADD')) {
    return 'bg-emerald-50 text-emerald-700 border-emerald-200'
  }
  return 'bg-sky-50 text-sky-700 border-sky-200'
}

function getSystemLevelMeta(level: string) {
  const l = (level || 'INFO').toUpperCase()
  if (l === 'ERROR' || l === 'FATAL') {
    return { cls: 'bg-rose-50 text-rose-700 border-rose-200', icon: <AlertCircle className="h-3 w-3" /> }
  }
  if (l === 'WARN' || l === 'WARNING') {
    return { cls: 'bg-amber-50 text-amber-700 border-amber-200', icon: <AlertTriangle className="h-3 w-3" /> }
  }
  return { cls: 'bg-sky-50 text-sky-700 border-sky-200', icon: <Info className="h-3 w-3" /> }
}

function formatDateTime(ts?: string) {
  if (!ts) return { date: '-', time: '' }
  const d = new Date(ts)
  if (Number.isNaN(d.getTime())) return { date: ts, time: '' }
  return {
    date: d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }),
    time: d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
  }
}

function formatDetailsText(details?: string | null): string {
  if (!details) return '-'
  try {
    const parsed = JSON.parse(details)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return details
  }
}

export function ActivityLogView() {
  const [activeTab, setActiveTab] = useState<'user' | 'system'>('user')

  // Tab 1: User Activity Logs
  const [users, setUsers] = useState<UserOption[]>([])
  const [logs, setLogs] = useState<ActivityLogItem[]>([])
  const [totalLogs, setTotalLogs] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(15)

  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [userIdFilter, setUserIdFilter] = useState('')
  const [actionFilter, setActionFilter] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')

  // Tab 2: System Logs
  const [systemLogs, setSystemLogs] = useState<SystemLogItem[]>([])
  const [totalSystemLogs, setTotalSystemLogs] = useState(0)
  const [systemLoading, setSystemLoading] = useState(false)
  const [systemPage, setSystemPage] = useState(1)
  const [systemLimit, setSystemLimit] = useState(15)
  const [autoRefresh, setAutoRefresh] = useState(false)

  const [sysSearchInput, setSysSearchInput] = useState('')
  const [sysDebouncedSearch, setSysDebouncedSearch] = useState('')
  const [sysLevel, setSysLevel] = useState('')
  const [sysDateFrom, setSysDateFrom] = useState('')
  const [sysDateTo, setSysDateTo] = useState('')

  // Detail Modal
  const [detailModal, setDetailModal] = useState<{
    title: string
    subtitle: string
    content: string
    badge?: string
  } | null>(null)
  const [copiedDetail, setCopiedDetail] = useState(false)

  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const sysSearchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      setDebouncedSearch(searchInput)
      setPage(1)
    }, 500)
    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    }
  }, [searchInput])

  useEffect(() => {
    if (sysSearchTimerRef.current) clearTimeout(sysSearchTimerRef.current)
    sysSearchTimerRef.current = setTimeout(() => {
      setSysDebouncedSearch(sysSearchInput)
      setSystemPage(1)
    }, 500)
    return () => {
      if (sysSearchTimerRef.current) clearTimeout(sysSearchTimerRef.current)
    }
  }, [sysSearchInput])

  useEffect(() => {
    apiClient
      .get('/users', { params: { skip: 0, limit: 1000 } })
      .then((res) => {
        const list = Array.isArray(res.data) ? res.data : res.data?.data || res.data?.items || []
        setUsers(list)
      })
      .catch(() => {})
  }, [])

  const fetchLogs = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = {
        skip: (page - 1) * limit,
        limit,
      }
      if (debouncedSearch.trim()) params.search = debouncedSearch.trim()
      if (userIdFilter) params.user_id = userIdFilter
      if (actionFilter) params.action = actionFilter
      if (dateFrom) params.date_from = dateFrom
      if (dateTo) params.date_to = dateTo

      const res = await apiClient.get('/activity-logs', { params })
      const items = res.data?.items || res.data?.data || []
      setLogs(Array.isArray(items) ? items : [])
      setTotalLogs(res.data?.total ?? res.data?.meta?.total ?? 0)
    } catch {
      setLogs([])
    } finally {
      setLoading(false)
    }
  }, [page, limit, debouncedSearch, userIdFilter, actionFilter, dateFrom, dateTo])

  useEffect(() => {
    fetchLogs()
  }, [fetchLogs])

  const fetchSystemLogs = useCallback(
    async (silent = false) => {
      if (!silent) setSystemLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: systemPage,
          page_size: systemLimit,
        }
        if (sysDebouncedSearch.trim()) params.search = sysDebouncedSearch.trim()
        if (sysLevel) params.level = sysLevel
        if (sysDateFrom) params.date_from = sysDateFrom
        if (sysDateTo) params.date_to = sysDateTo

        const res = await apiClient.get('/system-logs', { params })
        const rawList = res.data?.data || res.data?.items || (Array.isArray(res.data) ? res.data : [])
        const mapped = Array.isArray(rawList) ? rawList.map(parseSystemLogItem) : []
        setSystemLogs(mapped)
        setTotalSystemLogs(res.data?.total ?? res.data?.meta?.total ?? mapped.length)
      } catch {
        if (!silent) setSystemLogs([])
      } finally {
        if (!silent) setSystemLoading(false)
      }
    },
    [systemPage, systemLimit, sysDebouncedSearch, sysLevel, sysDateFrom, sysDateTo]
  )

  useEffect(() => {
    if (activeTab === 'system') {
      fetchSystemLogs(false)
    }
  }, [activeTab, fetchSystemLogs])

  useEffect(() => {
    if (!autoRefresh || activeTab !== 'system') return
    const timer = setInterval(() => {
      fetchSystemLogs(true)
    }, 15000)
    return () => clearInterval(timer)
  }, [autoRefresh, activeTab, fetchSystemLogs])

  const hasUserFilters = Boolean(searchInput || userIdFilter || actionFilter || dateFrom || dateTo)
  const hasSystemFilters = Boolean(sysSearchInput || sysLevel || sysDateFrom || sysDateTo)

  const clearUserFilters = () => {
    setSearchInput('')
    setDebouncedSearch('')
    setUserIdFilter('')
    setActionFilter('')
    setDateFrom('')
    setDateTo('')
    setPage(1)
  }

  const clearSystemFilters = () => {
    setSysSearchInput('')
    setSysDebouncedSearch('')
    setSysLevel('')
    setSysDateFrom('')
    setSysDateTo('')
    setSystemPage(1)
  }

  const setDateRangeQuick = (preset: 'today' | '7days' | 'month') => {
    const now = new Date()
    const fmt = (d: Date) =>
      `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const todayStr = fmt(now)
    if (preset === 'today') {
      setSysDateFrom(todayStr)
      setSysDateTo(todayStr)
    } else if (preset === '7days') {
      const start = new Date(now)
      start.setDate(now.getDate() - 6)
      setSysDateFrom(fmt(start))
      setSysDateTo(todayStr)
    } else {
      const start = new Date(now.getFullYear(), now.getMonth(), 1)
      setSysDateFrom(fmt(start))
      setSysDateTo(todayStr)
    }
    setSystemPage(1)
  }

  const totalUserPages = Math.max(1, Math.ceil(totalLogs / limit))
  const totalSystemPages = Math.max(1, Math.ceil(totalSystemLogs / systemLimit))

  const copyDetailContent = async () => {
    if (!detailModal) return
    try {
      await navigator.clipboard.writeText(detailModal.content)
      setCopiedDetail(true)
      setTimeout(() => setCopiedDetail(false), 1800)
    } catch {
      // ignore
    }
  }

  const userNameMap = useMemo(() => {
    const m = new Map<number, UserOption>()
    for (const u of users) m.set(u.id, u)
    return m
  }, [users])

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-primary text-primary-foreground">
            <History className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Log Aktivitas</h1>
            <p className="text-xs text-muted-foreground">Melacak perubahan data pengguna, cron auto-billing, & callback Xendit</p>
          </div>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'user' | 'system')} className="space-y-4">
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="user" className="gap-2 text-xs font-semibold">
            <UserCheck className="h-4 w-4" />
            Aktivitas Pengguna
          </TabsTrigger>
          <TabsTrigger value="system" className="gap-2 text-xs font-semibold">
            <Server className="h-4 w-4" />
            Sistem & Otomatisasi
          </TabsTrigger>
        </TabsList>

        {/* TAB 1: User Activity Logs */}
        <TabsContent value="user" className="space-y-4">
          <Card className="shadow-xs">
            <CardHeader className="py-2.5 px-4 flex flex-row items-center justify-between">
              <CardTitle className="text-sm font-semibold">Filter Pencarian Aktivitas User</CardTitle>
              <div className="flex items-center gap-2">
                {hasUserFilters && (
                  <Button variant="ghost" size="sm" className="h-8 text-xs text-rose-600" onClick={clearUserFilters}>
                    <FilterX className="h-3.5 w-3.5 mr-1" />
                    Reset Filter
                  </Button>
                )}
                <Button variant="outline" size="sm" className="h-8 text-xs" onClick={fetchLogs} disabled={loading}>
                  <RefreshCw className={`h-3.5 w-3.5 mr-1 ${loading ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="p-3.5">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-12 gap-2.5 items-end">
                <div className="lg:col-span-4 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Kata Kunci</label>
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      placeholder="Cari aktivitas, user, atau detail..."
                      className="pl-8 h-9 text-xs w-full"
                    />
                  </div>
                </div>
                <div className="lg:col-span-2 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Pengguna</label>
                  <Select value={userIdFilter || '__all__'} onValueChange={(v) => { setUserIdFilter(v === '__all__' ? '' : v); setPage(1) }}>
                    <SelectTrigger className="h-9 text-xs w-full"><SelectValue placeholder="Semua Pengguna" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__all__">Semua Pengguna</SelectItem>
                      {users.map((u) => (
                        <SelectItem key={u.id} value={String(u.id)}>
                          {u.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="lg:col-span-2 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Aksi</label>
                  <Select value={actionFilter || '__all__'} onValueChange={(v) => { setActionFilter(v === '__all__' ? '' : v); setPage(1) }}>
                    <SelectTrigger className="h-9 text-xs w-full"><SelectValue placeholder="Semua Aksi" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__all__">Semua Aksi</SelectItem>
                      <SelectItem value="POST">POST (Create)</SelectItem>
                      <SelectItem value="PATCH">PATCH (Update)</SelectItem>
                      <SelectItem value="DELETE">DELETE (Remove)</SelectItem>
                      <SelectItem value="GET">GET (View)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="lg:col-span-2 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Dari Tanggal</label>
                  <Input
                    type="date"
                    value={dateFrom}
                    onChange={(e) => { setDateFrom(e.target.value); setPage(1) }}
                    className="h-9 text-xs w-full"
                  />
                </div>
                <div className="lg:col-span-2 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Sampai Tanggal</label>
                  <Input
                    type="date"
                    value={dateTo}
                    onChange={(e) => { setDateTo(e.target.value); setPage(1) }}
                    className="h-9 text-xs w-full"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-xs overflow-hidden">
            <div className="px-4 py-3 border-b flex items-center justify-between">
              <div className="text-sm font-semibold text-foreground">Riwayat Aktivitas Pengguna</div>
              <Badge variant="secondary" className="text-xs">{totalLogs.toLocaleString('id-ID')} aktivitas</Badge>
            </div>
            {loading && logs.length > 0 && (
              <div className="h-0.5 w-full bg-muted overflow-hidden">
                <div className="h-full w-1/3 bg-primary animate-pulse" />
              </div>
            )}
            {loading && logs.length === 0 ? (
              <div className="py-16 text-center text-xs text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-foreground/90" />
                Memuat data log aktivitas...
              </div>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-muted/50">
                        <TableHead className="text-xs px-5 py-3 min-w-[260px]">Pengguna</TableHead>
                        <TableHead className="text-xs px-5 py-3 w-[240px]">Aksi</TableHead>
                        <TableHead className="text-xs px-5 py-3 w-[180px]">Waktu</TableHead>
                        <TableHead className="text-xs px-5 py-3 text-right w-[90px]">Detail</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {logs.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={4} className="text-center py-12 text-xs text-muted-foreground">
                            Belum ada aktivitas yang tercatat
                          </TableCell>
                        </TableRow>
                      ) : (
                        logs.map((item) => {
                          const resolvedUser = item.user || (item.user_id ? userNameMap.get(item.user_id) : null)
                          const userName = resolvedUser?.name || 'System'
                          const userEmail = resolvedUser?.email || '-'
                          const ts = formatDateTime(item.timestamp || item.created_at)
                          return (
                            <TableRow key={item.id} className="hover:bg-muted/40">
                              <TableCell className="px-5 py-3">
                                <div className="flex items-center gap-3">
                                  <div className="h-8 w-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-bold shrink-0">
                                    {userName.charAt(0).toUpperCase()}
                                  </div>
                                  <div className="min-w-0">
                                    <div className="text-xs font-semibold text-foreground truncate">{userName}</div>
                                    <div className="text-[11px] text-muted-foreground truncate">{userEmail}</div>
                                  </div>
                                </div>
                              </TableCell>
                              <TableCell className="px-5 py-3">
                                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-semibold border ${getActionBadgeClass(item.action)}`}>
                                  {item.action}
                                </span>
                              </TableCell>
                              <TableCell className="px-5 py-3 whitespace-nowrap">
                                <div className="text-xs font-medium text-foreground">{ts.date}</div>
                                <div className="text-[11px] text-muted-foreground font-mono">{ts.time}</div>
                              </TableCell>
                              <TableCell className="px-5 py-3 text-right">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    setCopiedDetail(false)
                                    setDetailModal({
                                      title: 'Detail Aktivitas Pengguna',
                                      subtitle: `${userName} • ${ts.date} ${ts.time}`,
                                      badge: item.action,
                                      content: formatDetailsText(item.details),
                                    })
                                  }}
                                  title="Lihat Detail"
                                >
                                  <Eye className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          )
                        })
                      )}
                    </TableBody>
                  </Table>
                </div>

                <Separator />
                <div className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-xs text-muted-foreground">
                  <div className="flex items-center gap-2">
                    <span>Halaman {page} dari {totalUserPages}</span>
                    <Select value={String(limit)} onValueChange={(v) => { setLimit(Number(v)); setPage(1) }}>
                      <SelectTrigger className="h-7 w-[75px] text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {[10, 15, 25, 50, 100].map((n) => (
                          <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                      <ChevronLeft className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="outline" size="icon" className="h-7 w-7" disabled={page >= totalUserPages} onClick={() => setPage((p) => p + 1)}>
                      <ChevronRight className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              </>
            )}
          </Card>
        </TabsContent>

        {/* TAB 2: System & Automation Logs */}
        <TabsContent value="system" className="space-y-4">
          <Card className="shadow-xs">
            <CardHeader className="py-2.5 px-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <CardTitle className="text-sm font-semibold">Filter Log Sistem & Otomatisasi</CardTitle>
              <div className="flex flex-wrap items-center gap-2.5">
                <div className="flex items-center gap-2 rounded-full border bg-muted/50 px-3 py-1">
                  <span className={`h-2 w-2 rounded-full ${autoRefresh ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'}`} />
                  <span className="text-xs font-medium text-foreground/90">Live (15s)</span>
                  <Switch checked={autoRefresh} onCheckedChange={setAutoRefresh} />
                </div>
                {hasSystemFilters && (
                  <Button variant="ghost" size="sm" className="h-8 text-xs text-rose-600" onClick={clearSystemFilters}>
                    <FilterX className="h-3.5 w-3.5 mr-1" />
                    Reset
                  </Button>
                )}
                <Button variant="outline" size="sm" className="h-8 text-xs" onClick={() => fetchSystemLogs(false)} disabled={systemLoading}>
                  <RefreshCw className={`h-3.5 w-3.5 mr-1 ${systemLoading ? 'animate-spin' : ''}`} />
                  Refresh
                </Button>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="p-3.5 space-y-2.5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="inline-flex rounded-lg border bg-muted/50 p-0.5">
                  {(['', 'INFO', 'WARN', 'ERROR'] as const).map((lvl) => (
                    <button
                      key={lvl || 'all'}
                      type="button"
                      onClick={() => { setSysLevel(lvl); setSystemPage(1) }}
                      className={`rounded-md px-3 py-1 text-xs font-medium transition-colors ${
                        sysLevel === lvl ? 'bg-primary text-primary-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'
                      }`}
                    >
                      {lvl || 'Semua'}
                    </button>
                  ))}
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-[11px] text-muted-foreground mr-1">Rentang Cepat:</span>
                  <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => setDateRangeQuick('today')}>Hari Ini</Button>
                  <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => setDateRangeQuick('7days')}>7 Hari</Button>
                  <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => setDateRangeQuick('month')}>Bulan Ini</Button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-12 gap-2.5 items-end">
                <div className="md:col-span-6 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Kata Kunci</label>
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                      value={sysSearchInput}
                      onChange={(e) => setSysSearchInput(e.target.value)}
                      placeholder="Contoh: SchedulerManager, XENDIT, SUSPEND, CRON..."
                      className="pl-8 h-9 text-xs w-full"
                    />
                  </div>
                </div>
                <div className="md:col-span-3 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Dari Tanggal</label>
                  <Input
                    type="date"
                    value={sysDateFrom}
                    onChange={(e) => { setSysDateFrom(e.target.value); setSystemPage(1) }}
                    className="h-9 text-xs w-full"
                  />
                </div>
                <div className="md:col-span-3 space-y-1">
                  <label className="text-[11px] font-medium text-muted-foreground">Sampai Tanggal</label>
                  <Input
                    type="date"
                    value={sysDateTo}
                    onChange={(e) => { setSysDateTo(e.target.value); setSystemPage(1) }}
                    className="h-9 text-xs w-full"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="shadow-xs overflow-hidden">
            <div className="px-4 py-3 border-b flex items-center justify-between">
              <div className="text-sm font-semibold text-foreground">Log Eksekusi Sistem</div>
              <Badge variant="secondary" className="text-xs">{totalSystemLogs.toLocaleString('id-ID')} log</Badge>
            </div>
            {systemLoading && systemLogs.length > 0 && (
              <div className="h-0.5 w-full bg-muted overflow-hidden">
                <div className="h-full w-1/3 bg-primary animate-pulse" />
              </div>
            )}
            {systemLoading && systemLogs.length === 0 ? (
              <div className="py-16 text-center text-xs text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-foreground/90" />
                Memuat system logs...
              </div>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-muted/50">
                        <TableHead className="text-xs px-5 py-3 w-[110px]">Level</TableHead>
                        <TableHead className="text-xs px-5 py-3 w-[200px]">Modul</TableHead>
                        <TableHead className="text-xs px-5 py-3 min-w-[300px]">Pesan</TableHead>
                        <TableHead className="text-xs px-5 py-3 w-[160px]">Waktu</TableHead>
                        <TableHead className="text-xs px-5 py-3 text-right w-[80px]">Detail</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {systemLogs.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={5} className="text-center py-12 text-xs text-muted-foreground">
                            Belum ada log sistem ditemukan
                          </TableCell>
                        </TableRow>
                      ) : (
                        systemLogs.map((item) => {
                          const lvl = getSystemLevelMeta(item.level)
                          const ts = formatDateTime(item.timestamp)
                          return (
                            <TableRow key={item.id} className="hover:bg-muted/40">
                              <TableCell className="px-5 py-3">
                                <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-semibold border ${lvl.cls}`}>
                                  {lvl.icon}
                                  {item.level}
                                </span>
                              </TableCell>
                              <TableCell className="px-5 py-3">
                                <div className="text-xs font-semibold text-foreground">{item.module}</div>
                                {item.action !== '-' && (
                                  <div className="text-[11px] text-muted-foreground truncate max-w-[180px]">{item.action}</div>
                                )}
                              </TableCell>
                              <TableCell className="px-5 py-3 text-xs text-foreground/90 max-w-md truncate">
                                {item.message}
                              </TableCell>
                              <TableCell className="px-5 py-3 whitespace-nowrap">
                                <div className="text-xs font-medium text-foreground">{ts.date}</div>
                                <div className="text-[11px] text-muted-foreground font-mono">{ts.time}</div>
                              </TableCell>
                              <TableCell className="px-5 py-3 text-right">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    setCopiedDetail(false)
                                    setDetailModal({
                                      title: `Log Sistem: ${item.module}`,
                                      subtitle: `${ts.date} ${ts.time}`,
                                      badge: item.level,
                                      content: item.rawMessage || item.message,
                                    })
                                  }}
                                >
                                  <Eye className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          )
                        })
                      )}
                    </TableBody>
                  </Table>
                </div>

                <Separator />
                <div className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-xs text-muted-foreground">
                  <div className="flex items-center gap-2">
                    <span>Halaman {systemPage} dari {totalSystemPages}</span>
                    <Select value={String(systemLimit)} onValueChange={(v) => { setSystemLimit(Number(v)); setSystemPage(1) }}>
                      <SelectTrigger className="h-7 w-[75px] text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {[15, 25, 50, 100].map((n) => (
                          <SelectItem key={n} value={String(n)}>{n}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button variant="outline" size="icon" className="h-7 w-7" disabled={systemPage <= 1} onClick={() => setSystemPage((p) => p - 1)}>
                      <ChevronLeft className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="outline" size="icon" className="h-7 w-7" disabled={systemPage >= totalSystemPages} onClick={() => setSystemPage((p) => p + 1)}>
                      <ChevronRight className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              </>
            )}
          </Card>
        </TabsContent>
      </Tabs>

      {/* Detail Dialog */}
      <Dialog open={Boolean(detailModal)} onOpenChange={(o) => !o && setDetailModal(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {detailModal?.title}
              {detailModal?.badge && <Badge variant="secondary" className="text-xs">{detailModal.badge}</Badge>}
            </DialogTitle>
            <DialogDescription>{detailModal?.subtitle}</DialogDescription>
          </DialogHeader>
          <div className="relative">
            <pre className="rounded-lg border bg-muted/50 p-3 text-xs font-mono text-foreground whitespace-pre-wrap break-all max-h-[360px] overflow-y-auto">
              {detailModal?.content}
            </pre>
            <Button
              variant="outline"
              size="sm"
              className="mt-2 h-7 text-xs"
              onClick={copyDetailContent}
            >
              {copiedDetail ? <Check className="h-3.5 w-3.5 mr-1 text-emerald-600" /> : <Copy className="h-3.5 w-3.5 mr-1" />}
              {copiedDetail ? 'Tersalin' : 'Salin Konten'}
            </Button>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailModal(null)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
