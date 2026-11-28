import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  ArrowLeft,
  Clock,
  FileText,
  History,
  Loader2,
  Receipt,
  Search,
  SearchX,
  TicketCheck,
  Users,
  Wifi,
} from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

interface SearchResult {
  id: string | number
  type: string
  title: string
  subtitle?: string
  url?: string
  data?: Record<string, any>
}

interface SearchResultsPayload {
  query: string
  categories: string[]
  results: Record<string, SearchResult[]>
  total_count: number
}

const ITEMS_PER_PAGE = 20

const CATEGORY_META: Record<string, { title: string; icon: React.ReactNode }> = {
  pelanggan: { title: 'Pelanggan', icon: <Users className="h-4 w-4 text-blue-600 dark:text-blue-400" /> },
  langganan: { title: 'Langganan', icon: <Wifi className="h-4 w-4 text-emerald-600 dark:text-emerald-400" /> },
  invoices: { title: 'Invoice', icon: <Receipt className="h-4 w-4 text-amber-600 dark:text-amber-400" /> },
  tickets: { title: 'Trouble Tickets', icon: <TicketCheck className="h-4 w-4 text-rose-600 dark:text-rose-400" /> },
  activity_logs: { title: 'Activity Log', icon: <History className="h-4 w-4 text-sky-600 dark:text-sky-400" /> },
}

function getItemBadge(item: SearchResult, category: string): { label: string; cls: string } | null {
  const status = item.data?.status
  if (category === 'langganan' && status) {
    const label = status === 'suspended' ? 'Suspended' : status
    const cls =
      status === 'active' || status === 'Aktif'
        ? 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
        : status === 'suspended' || status === 'Suspended'
          ? 'bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border-amber-200 dark:border-amber-800'
          : 'bg-red-50 dark:bg-red-950/50 text-red-700 dark:text-red-300 border-red-200 dark:border-red-800'
    return { label, cls }
  }
  if (category === 'invoices' && status) {
    const label = status === 'unpaid' ? 'Belum Bayar' : status
    const cls =
      status === 'unpaid' || status === 'Belum Bayar'
        ? 'bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border-amber-200 dark:border-amber-800'
        : 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
    return { label, cls }
  }
  if (category === 'tickets' && status) {
    return {
      label: status,
      cls: 'bg-muted text-foreground border-border',
    }
  }
  return null
}

export function GlobalSearchResultsView() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const queryParam = searchParams.get('q') || ''

  const [loading, setLoading] = useState(false)
  const [searchResults, setSearchResults] = useState<SearchResultsPayload | null>(null)
  const [activeCategories, setActiveCategories] = useState<string[]>([])
  const [categoryPages, setCategoryPages] = useState<Record<string, number>>({})
  const [searchHistory, setSearchHistory] = useState<string[]>(() => {
    try {
      const saved = localStorage.getItem('search_history')
      return saved ? JSON.parse(saved) : []
    } catch {
      return []
    }
  })

  const addToSearchHistory = useCallback((query: string) => {
    const trimmed = query.trim()
    if (trimmed.length < 2) return
    setSearchHistory((prev) => {
      const next = [trimmed, ...prev.filter((q) => q !== trimmed)].slice(0, 10)
      try {
        localStorage.setItem('search_history', JSON.stringify(next))
      } catch {
        // ignore storage errors
      }
      return next
    })
  }, [])

  const performSearch = useCallback(
    async (query: string) => {
      const q = query.trim()
      if (q.length < 2) {
        setSearchResults(null)
        setLoading(false)
        return
      }
      setLoading(true)
      try {
        const tasks: Array<Promise<{ cat: string; items: SearchResult[] }>> = []

        if (hasPermission('view_pelanggan')) {
          tasks.push(
            apiClient
              .get('/pelanggan', { params: { search: q, limit: 20 }, headers: { 'X-Silent': 'true' } })
              .then((res) => {
                const rows = Array.isArray(res.data?.data) ? res.data.data : Array.isArray(res.data) ? res.data : []
                return {
                  cat: 'pelanggan',
                  items: rows.map((p: Record<string, any>) => ({
                    id: p.id,
                    type: 'pelanggan',
                    title: p.nama || '-',
                    subtitle: [p.alamat, p.no_telp, p.email].filter(Boolean).join(' • '),
                    url: `/pelanggan`,
                    data: p,
                  })),
                }
              })
              .catch(() => ({ cat: 'pelanggan', items: [] }))
          )
        }

        if (hasPermission('view_langganan')) {
          tasks.push(
            apiClient
              .get('/langganan', { params: { search: q, limit: 20 }, headers: { 'X-Silent': 'true' } })
              .then((res) => {
                const rows = Array.isArray(res.data?.data) ? res.data.data : []
                return {
                  cat: 'langganan',
                  items: rows.map((l: Record<string, any>) => ({
                    id: l.id,
                    type: 'langganan',
                    title: l.pelanggan?.nama || `Langganan #${l.id}`,
                    subtitle: [l.paket_layanan?.nama_paket, l.metode_pembayaran].filter(Boolean).join(' • '),
                    url: `/langganan`,
                    data: { status: l.status },
                  })),
                }
              })
              .catch(() => ({ cat: 'langganan', items: [] }))
          )
        }

        if (hasPermission('view_invoices')) {
          tasks.push(
            apiClient
              .get('/invoices', { params: { search: q, limit: 20 }, headers: { 'X-Silent': 'true' } })
              .then((res) => {
                const rows = Array.isArray(res.data?.data) ? res.data.data : []
                return {
                  cat: 'invoices',
                  items: rows.map((inv: Record<string, any>) => ({
                    id: inv.id,
                    type: 'invoices',
                    title: inv.invoice_number || `Invoice #${inv.id}`,
                    subtitle: [inv.pelanggan?.nama, inv.id_pelanggan].filter(Boolean).join(' • '),
                    url: `/invoices`,
                    data: { status: inv.status_invoice },
                  })),
                }
              })
              .catch(() => ({ cat: 'invoices', items: [] }))
          )
        }

        if (hasPermission('view_trouble_tickets')) {
          tasks.push(
            apiClient
              .get('/trouble-tickets', { params: { search: q, limit: 20 }, headers: { 'X-Silent': 'true' } })
              .then((res) => {
                const rows = Array.isArray(res.data?.data) ? res.data.data : []
                return {
                  cat: 'tickets',
                  items: rows.map((t: Record<string, any>) => ({
                    id: t.id,
                    type: 'tickets',
                    title: `${t.ticket_number || `#${t.id}`} — ${t.title || ''}`,
                    subtitle: [t.pelanggan?.nama, t.category].filter(Boolean).join(' • '),
                    url: `/trouble-tickets/${t.id}/action`,
                    data: { status: t.status, priority: t.priority },
                  })),
                }
              })
              .catch(() => ({ cat: 'tickets', items: [] }))
          )
        }

        if (hasPermission('view_activity_log')) {
          tasks.push(
            apiClient
              .get('/activity-logs', { params: { search: q, limit: 20 }, headers: { 'X-Silent': 'true' } })
              .then((res) => {
                const rows = Array.isArray(res.data?.items)
                  ? res.data.items
                  : Array.isArray(res.data?.data)
                    ? res.data.data
                    : []
                return {
                  cat: 'activity_logs',
                  items: rows.map((a: Record<string, any>) => ({
                    id: a.id,
                    type: 'activity_logs',
                    title: a.action || 'Activity',
                    subtitle: [a.user?.name, a.details].filter(Boolean).join(' • '),
                    url: `/activity-logs`,
                    data: a,
                  })),
                }
              })
              .catch(() => ({ cat: 'activity_logs', items: [] }))
          )
        }

        const settled = await Promise.all(tasks)
        const resultsMap: Record<string, SearchResult[]> = {}
        const categories: string[] = []
        let totalCount = 0

        for (const entry of settled) {
          if (entry.items.length > 0) {
            categories.push(entry.cat)
            resultsMap[entry.cat] = entry.items
            totalCount += entry.items.length
          }
        }

        setSearchResults({
          query: q,
          categories,
          results: resultsMap,
          total_count: totalCount,
        })

        const initialPages: Record<string, number> = {}
        categories.forEach((cat) => {
          initialPages[cat] = 1
        })
        setCategoryPages(initialPages)
      } finally {
        setLoading(false)
      }
    },
    [hasPermission]
  )

  useEffect(() => {
    setActiveCategories([])
    void performSearch(queryParam)
  }, [queryParam, performSearch])

  const availableCategories = useMemo(() => searchResults?.categories || [], [searchResults])
  const totalResults = searchResults?.total_count || 0

  const categorizedResults = useMemo(() => {
    if (!searchResults) return {}
    const selected = activeCategories.length > 0 ? activeCategories : availableCategories
    const map: Record<string, SearchResult[]> = {}
    selected.forEach((cat) => {
      const list = searchResults.results?.[cat] || []
      if (list.length > 0) map[cat] = list
    })
    return map
  }, [searchResults, activeCategories, availableCategories])

  const toggleCategory = (category: string) => {
    setActiveCategories((prev) =>
      prev.includes(category) ? prev.filter((c) => c !== category) : [...prev, category]
    )
    setCategoryPages((prev) => ({ ...prev, [category]: 1 }))
  }

  const handleItemClick = (item: SearchResult) => {
    addToSearchHistory(queryParam)
    if (item.url) {
      navigate(item.url)
    }
  }

  return (
    <div className="p-4 md:p-6 max-w-6xl mx-auto space-y-5">
      {/* Header Card */}
      <Card className="py-0 overflow-hidden">
        <CardContent className="p-4 md:p-5 space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-primary text-primary-foreground flex items-center justify-center shrink-0">
                <Search className="h-5 w-5" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-foreground">Hasil Pencarian Global</h1>
                <p className="text-xs text-muted-foreground">
                  Menampilkan <strong className="text-foreground">{totalResults}</strong> hasil
                  untuk &quot;{queryParam}&quot;
                </p>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
              <ArrowLeft className="h-4 w-4 mr-1.5" />
              Kembali
            </Button>
          </div>

          {availableCategories.length > 1 && (
            <div className="flex flex-wrap gap-2 pt-1">
              {availableCategories.map((cat) => {
                const active = activeCategories.includes(cat)
                const meta = CATEGORY_META[cat] || { title: cat, icon: <FileText className="h-4 w-4" /> }
                return (
                  <button
                    key={cat}
                    type="button"
                    onClick={() => toggleCategory(cat)}
                    className={`inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'bg-muted/50 text-foreground hover:bg-muted border-border'
                    }`}
                  >
                    {meta.icon}
                    <span>{meta.title}</span>
                  </button>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Content */}
      {loading ? (
        <div className="py-16 text-center space-y-2">
          <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto" />
          <div className="text-sm font-medium text-foreground">Mencari...</div>
          <div className="text-xs text-muted-foreground">Sedang mencari di seluruh sistem</div>
        </div>
      ) : Object.keys(categorizedResults).length > 0 ? (
        <div className="space-y-4">
          {Object.entries(categorizedResults).map(([category, items]) => {
            const meta = CATEGORY_META[category] || {
              title: category,
              icon: <FileText className="h-4 w-4 text-muted-foreground" />,
            }
            const page = categoryPages[category] || 1
            const pageCount = Math.max(1, Math.ceil(items.length / ITEMS_PER_PAGE))
            const visibleItems = items.slice((page - 1) * ITEMS_PER_PAGE, page * ITEMS_PER_PAGE)

            return (
              <Card key={category} className="py-0 overflow-hidden">
                <CardHeader className="px-5 py-3.5 flex flex-row items-center justify-between space-y-0">
                  <div className="flex items-center gap-2">
                    {meta.icon}
                    <CardTitle className="text-sm font-bold">{meta.title}</CardTitle>
                  </div>
                  <Badge variant="secondary" className="text-xs">
                    {items.length} hasil
                  </Badge>
                </CardHeader>
                <Separator />
                <CardContent className="p-0 divide-y divide-border">
                  {visibleItems.map((item) => {
                    const badge = getItemBadge(item, category)
                    return (
                      <button
                        key={`${category}-${item.id}`}
                        type="button"
                        onClick={() => handleItemClick(item)}
                        className="w-full px-5 py-3 flex items-center justify-between gap-3 text-left hover:bg-muted/40 transition-colors"
                      >
                        <div className="min-w-0">
                          <div className="text-sm font-medium text-foreground truncate">
                            {item.title}
                          </div>
                          {item.subtitle && (
                            <div className="text-xs text-muted-foreground truncate mt-0.5">
                              {item.subtitle}
                            </div>
                          )}
                        </div>
                        {badge && (
                          <span
                            className={`inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-md border shrink-0 ${badge.cls}`}
                          >
                            {badge.label}
                          </span>
                        )}
                      </button>
                    )
                  })}

                  {pageCount > 1 && (
                    <div className="px-5 py-3 flex items-center justify-between text-xs text-muted-foreground">
                      <span>
                        Halaman {page} dari {pageCount}
                      </span>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 text-xs"
                          disabled={page <= 1}
                          onClick={() =>
                            setCategoryPages((p) => ({ ...p, [category]: Math.max(1, page - 1) }))
                          }
                        >
                          Sebelumnya
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 text-xs"
                          disabled={page >= pageCount}
                          onClick={() =>
                            setCategoryPages((p) => ({
                              ...p,
                              [category]: Math.min(pageCount, page + 1),
                            }))
                          }
                        >
                          Berikutnya
                        </Button>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      ) : (
        <Card className="py-0">
          <CardContent className="py-14 text-center space-y-3">
            <SearchX className="h-12 w-12 text-muted-foreground/40 mx-auto" />
            <div className="text-base font-semibold text-foreground">Tidak Ada Hasil</div>
            <p className="text-xs text-muted-foreground max-w-md mx-auto">
              Tidak ada hasil yang ditemukan untuk &quot;{queryParam}&quot;. Periksa ejaan kata
              kunci, gunakan kata kunci lebih umum, atau pastikan Anda memiliki hak akses.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Search History */}
      {searchHistory.length > 0 && (
        <Card className="py-0 overflow-hidden">
          <CardHeader className="px-5 py-3 flex flex-row items-center gap-2 space-y-0">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <CardTitle className="text-sm font-semibold">Pencarian Terbaru</CardTitle>
          </CardHeader>
          <Separator />
          <CardContent className="p-3.5 flex flex-wrap gap-2">
            {searchHistory.map((hist) => (
              <button
                key={hist}
                type="button"
                onClick={() => setSearchParams({ q: hist })}
                className="inline-flex items-center gap-1.5 rounded-md border bg-muted/40 hover:bg-muted px-2.5 py-1 text-xs text-foreground transition-colors"
              >
                <Search className="h-3 w-3 text-muted-foreground" />
                {hist}
              </button>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
