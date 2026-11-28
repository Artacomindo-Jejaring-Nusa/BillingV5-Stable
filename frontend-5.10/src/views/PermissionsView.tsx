import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import apiClient from '@/services/api'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
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
  ShieldCheck,
  KeyRound,
  PlusCircle,
  Eye,
  Pencil,
  Trash2,
  RefreshCw,
  Search,
  FilterX,
  Wand2,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
  ChevronLeft,
  ChevronRight,
  Copy,
  Check,
  LayoutGrid,
  List,
} from 'lucide-react'

interface PermissionItem {
  id: number
  name: string
}

type ActionKind = 'view' | 'create' | 'edit' | 'delete' | 'manage' | 'other'

function getActionKind(name: string): ActionKind {
  const n = name.toLowerCase()
  if (n.startsWith('view_')) return 'view'
  if (n.startsWith('create_')) return 'create'
  if (n.startsWith('edit_')) return 'edit'
  if (n.startsWith('delete_')) return 'delete'
  if (n.startsWith('manage_')) return 'manage'
  return 'other'
}

const ACTION_META: Record<ActionKind, { label: string; cls: string; dot: string; icon: React.ReactNode }> = {
  view: {
    label: 'View',
    cls: 'bg-sky-50 text-sky-700 border-sky-200 hover:bg-sky-100',
    dot: 'bg-sky-500',
    icon: <Eye className="h-3 w-3" />,
  },
  create: {
    label: 'Create',
    cls: 'bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-100',
    dot: 'bg-emerald-500',
    icon: <PlusCircle className="h-3 w-3" />,
  },
  edit: {
    label: 'Edit',
    cls: 'bg-amber-50 text-amber-700 border-amber-200 hover:bg-amber-100',
    dot: 'bg-amber-500',
    icon: <Pencil className="h-3 w-3" />,
  },
  delete: {
    label: 'Delete',
    cls: 'bg-rose-50 text-rose-700 border-rose-200 hover:bg-rose-100',
    dot: 'bg-rose-500',
    icon: <Trash2 className="h-3 w-3" />,
  },
  manage: {
    label: 'Manage',
    cls: 'bg-violet-50 text-violet-700 border-violet-200 hover:bg-violet-100',
    dot: 'bg-violet-500',
    icon: <ShieldCheck className="h-3 w-3" />,
  },
  other: {
    label: 'Other',
    cls: 'bg-muted/50 text-foreground/90 border-border hover:bg-muted',
    dot: 'bg-slate-400',
    icon: <KeyRound className="h-3 w-3" />,
  },
}

const CATEGORY_ORDER = [
  'Dashboard Admin & Widget',
  'Dashboard Jakinet & Widget',
  'FTTH & Pelanggan',
  'Billing & Laporan',
  'Customer Support',
  'Network Management',
  'Tools & S&K',
  'System & RBAC',
  'Lainnya',
]

function getPermissionCategory(name: string): string {
  const n = name.toLowerCase()
  if (n.includes('dashboard_pelanggan') || n.includes('widget_pelanggan_') || n === 'view_widget_statistik_pelanggan') {
    return 'Dashboard Jakinet & Widget'
  }
  if (n === 'view_dashboard' || n.endsWith('_dashboard') || n.startsWith('view_widget_') || n.includes('ai_analytics')) {
    return 'Dashboard Admin & Widget'
  }
  if (n.includes('pelanggan') || n.includes('data_teknis') || n.includes('langganan') || n.includes('brand')) {
    return 'FTTH & Pelanggan'
  }
  if (n.includes('invoice') || n.includes('diskon') || n.includes('revenue')) {
    return 'Billing & Laporan'
  }
  if (n.includes('simulasi_harga') || n.endsWith('_sk') || n === 'manage_sk') {
    return 'Tools & S&K'
  }
  if (n.includes('trouble_tickets') || n.includes('uploads') || n.includes('customer_chat') || n.includes('live_chat')) {
    return 'Customer Support'
  }
  if (n.includes('mikrotik') || n.includes('olt') || n.includes('odp') || n.includes('traffic') || n.includes('inventory')) {
    return 'Network Management'
  }
  if (n.includes('users') || n.includes('roles') || n.includes('permissions') || n.includes('activity_log') || n.includes('settings')) {
    return 'System & RBAC'
  }
  return 'Lainnya'
}

function getResourceKey(name: string): string {
  const n = name.toLowerCase()
  if (n.startsWith('view_widget_')) return 'Widget: ' + formatLabel(n.replace(/^view_widget_/, ''))
  return formatLabel(n.replace(/^(view|create|edit|delete|manage)_/, ''))
}

function formatLabel(raw: string): string {
  return raw
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (l) => l.toUpperCase())
    .replace(/\bOlt\b/g, 'OLT')
    .replace(/\bOdp\b/g, 'ODP')
    .replace(/\bSk\b/g, 'S&K')
    .replace(/\bAi\b/g, 'AI')
}

export function PermissionsView() {
  const [permissions, setPermissions] = useState<PermissionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [generating, setGenerating] = useState(false)

  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [actionFilter, setActionFilter] = useState<ActionKind | ''>('')
  const [viewMode, setViewMode] = useState<'matrix' | 'table'>('matrix')
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 20

  const [detailItem, setDetailItem] = useState<PermissionItem | null>(null)
  const [copiedKey, setCopiedKey] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4000)
  }

  const fetchPermissions = useCallback(async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/permissions')
      setPermissions(Array.isArray(res.data) ? res.data : res.data?.data || [])
    } catch {
      notify('Gagal memuat daftar permissions', 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchPermissions()
  }, [fetchPermissions])

  const generatePermissions = async () => {
    setGenerating(true)
    try {
      const res = await apiClient.post('/permissions/generate')
      const count = res.data?.created ?? 0
      notify(`Permissions berhasil di-generate (${count} permission baru ditambahkan)!`, 'success')
      fetchPermissions()
    } catch (err: unknown) {
      const e = err as { response?: { status?: number; data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.detail || e.response?.data?.error || 'Gagal men-generate permissions', 'error')
    } finally {
      setGenerating(false)
    }
  }

  const stats = useMemo(() => {
    const counts: Record<ActionKind, number> = { view: 0, create: 0, edit: 0, delete: 0, manage: 0, other: 0 }
    for (const p of permissions) counts[getActionKind(p.name)]++
    return { total: permissions.length, ...counts }
  }, [permissions])

  const categories = useMemo(() => {
    const counts = new Map<string, number>()
    for (const p of permissions) {
      const cat = getPermissionCategory(p.name)
      counts.set(cat, (counts.get(cat) || 0) + 1)
    }
    return CATEGORY_ORDER.filter((c) => counts.has(c)).map((c) => ({ name: c, count: counts.get(c) || 0 }))
  }, [permissions])

  const filteredPermissions = useMemo(() => {
    const q = search.toLowerCase().trim()
    return permissions.filter((p) => {
      const cat = getPermissionCategory(p.name)
      const kind = getActionKind(p.name)
      const matchSearch =
        !q ||
        p.name.toLowerCase().includes(q) ||
        formatLabel(p.name).toLowerCase().includes(q) ||
        cat.toLowerCase().includes(q)
      const matchCat = !categoryFilter || cat === categoryFilter
      const matchAction = !actionFilter || kind === actionFilter
      return matchSearch && matchCat && matchAction
    })
  }, [permissions, search, categoryFilter, actionFilter])

  // Group by Category -> Resource for Matrix View
  const groupedMatrix = useMemo(() => {
    const catMap = new Map<string, Map<string, PermissionItem[]>>()
    for (const p of filteredPermissions) {
      const cat = getPermissionCategory(p.name)
      const res = getResourceKey(p.name)
      if (!catMap.has(cat)) catMap.set(cat, new Map())
      const resMap = catMap.get(cat)!
      if (!resMap.has(res)) resMap.set(res, [])
      resMap.get(res)!.push(p)
    }

    const actionOrder: Record<ActionKind, number> = { view: 1, create: 2, edit: 3, delete: 4, manage: 5, other: 6 }
    return CATEGORY_ORDER.filter((c) => catMap.has(c)).map((catName) => {
      const resMap = catMap.get(catName)!
      const resources = Array.from(resMap.entries())
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([resourceName, items]) => ({
          resourceName,
          items: items.sort((a, b) => actionOrder[getActionKind(a.name)] - actionOrder[getActionKind(b.name)]),
        }))
      const total = resources.reduce((sum, r) => sum + r.items.length, 0)
      return { catName, total, resources }
    })
  }, [filteredPermissions])

  const totalPages = Math.max(1, Math.ceil(filteredPermissions.length / itemsPerPage))
  const paginatedPermissions = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return filteredPermissions.slice(start, start + itemsPerPage)
  }, [filteredPermissions, currentPage])

  const copyKey = async (key: string) => {
    try {
      await navigator.clipboard.writeText(key)
      setCopiedKey(true)
      setTimeout(() => setCopiedKey(false), 1800)
    } catch {
      // ignore clipboard errors
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-primary text-primary-foreground">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Permissions Matrix</h1>
            <p className="text-xs text-muted-foreground">Pemetaan seluruh hak akses RBAC per modul & resource sistem</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchPermissions} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button size="sm" onClick={generatePermissions} disabled={generating}>
            {generating ? <Loader2 className="h-4 w-4 mr-1.5 animate-spin" /> : <Wand2 className="h-4 w-4 mr-1.5" />}
            Auto Generate
          </Button>
        </div>
      </div>

      {/* Unified Action Distribution Bar (clickable filters) */}
      <div className="rounded-xl border bg-card px-4 py-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <KeyRound className="h-4 w-4 text-muted-foreground" />
          <span className="text-xs font-medium text-muted-foreground">Total Terdaftar:</span>
          <span className="text-sm font-bold text-foreground">{stats.total} Permissions</span>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          {(['view', 'create', 'edit', 'delete', 'manage'] as ActionKind[]).map((kind) => {
            const meta = ACTION_META[kind]
            const active = actionFilter === kind
            return (
              <button
                key={kind}
                type="button"
                onClick={() => {
                  setActionFilter(active ? '' : kind)
                  setCurrentPage(1)
                }}
                className={`inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs font-medium transition-colors ${
                  active
                    ? 'bg-primary text-primary-foreground border-primary'
                    : 'bg-muted/40 text-foreground/90 border-border hover:bg-muted'
                }`}
              >
                <span className={`h-1.5 w-1.5 rounded-full ${active ? 'bg-card' : meta.dot}`} />
                <span>{meta.label}</span>
                <span className={`font-mono text-[11px] ${active ? 'text-primary-foreground/70' : 'text-muted-foreground'}`}>
                  {stats[kind]}
                </span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Feedback Banner */}
      {feedback && (
        <div
          className={`flex items-center justify-between gap-3 px-4 py-2.5 rounded-xl border text-sm ${
            feedback.type === 'success'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-800'
              : 'bg-red-50 border-red-200 text-red-800'
          }`}
        >
          <div className="flex items-center gap-2">
            {feedback.type === 'success' ? (
              <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 text-red-600 shrink-0" />
            )}
            <span className="font-medium">{feedback.text}</span>
          </div>
          <button type="button" onClick={() => setFeedback(null)}><X className="h-4 w-4" /></button>
        </div>
      )}

      {/* Filter Toolbar & Category Chips */}
      <Card className="p-3.5 space-y-3 shadow-xs">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => { setSearch(e.target.value); setCurrentPage(1) }}
              placeholder="Cari modul, nama permission, atau kode key..."
              className="pl-9 h-8.5 text-xs"
            />
          </div>

          <div className="flex items-center gap-2">
            {(search || categoryFilter || actionFilter) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSearch('')
                  setCategoryFilter('')
                  setActionFilter('')
                  setCurrentPage(1)
                }}
                className="h-8 text-xs text-rose-600"
              >
                <FilterX className="h-3.5 w-3.5 mr-1" /> Reset
              </Button>
            )}

            <div className="inline-flex rounded-lg border bg-muted/50 p-0.5">
              <button
                type="button"
                onClick={() => setViewMode('matrix')}
                className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                  viewMode === 'matrix' ? 'bg-card text-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <LayoutGrid className="h-3.5 w-3.5" />
                Matrix Modul
              </button>
              <button
                type="button"
                onClick={() => setViewMode('table')}
                className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                  viewMode === 'table' ? 'bg-card text-foreground shadow-xs' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <List className="h-3.5 w-3.5" />
                Daftar Flat
              </button>
            </div>
          </div>
        </div>

        {/* Category Filter Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-0.5">
          <button
            type="button"
            onClick={() => { setCategoryFilter(''); setCurrentPage(1) }}
            className={`shrink-0 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
              !categoryFilter ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted'
            }`}
          >
            Semua ({permissions.length})
          </button>
          {categories.map((c) => (
            <button
              key={c.name}
              type="button"
              onClick={() => { setCategoryFilter(categoryFilter === c.name ? '' : c.name); setCurrentPage(1) }}
              className={`shrink-0 rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                categoryFilter === c.name
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted'
              }`}
            >
              {c.name} ({c.count})
            </button>
          ))}
        </div>
      </Card>

      {/* Main Content */}
      {loading && permissions.length === 0 ? (
        <Card className="py-16 text-center text-xs text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-foreground/90" />
          Memuat permissions...
        </Card>
      ) : filteredPermissions.length === 0 ? (
        <Card className="py-14 text-center text-xs text-muted-foreground">
          Tidak ada permissions yang cocok dengan filter.
        </Card>
      ) : viewMode === 'matrix' ? (
        <div className="space-y-4">
          {groupedMatrix.map((group) => (
            <Card key={group.catName} className="overflow-hidden shadow-xs">
              <div className="bg-muted/60 border-b px-4 py-2.5 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold uppercase tracking-wider text-foreground">{group.catName}</span>
                  <Badge variant="secondary" className="text-[11px] font-semibold">
                    {group.resources.length} fitur · {group.total} keys
                  </Badge>
                </div>
              </div>
              <div className="divide-y">
                {group.resources.map((res) => (
                  <div
                    key={res.resourceName}
                    className="px-4 py-2.5 flex flex-col md:flex-row md:items-center justify-between gap-2.5 hover:bg-muted/30 transition-colors"
                  >
                    <div className="min-w-[210px]">
                      <div className="text-sm font-medium text-foreground">{res.resourceName}</div>
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5">
                      {res.items.map((perm) => {
                        const kind = getActionKind(perm.name)
                        const meta = ACTION_META[kind]
                        return (
                          <button
                            key={perm.id}
                            type="button"
                            onClick={() => { setCopiedKey(false); setDetailItem(perm) }}
                            title={`#${perm.id} • ${perm.name}`}
                            className={`inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-medium transition-colors ${meta.cls}`}
                          >
                            {meta.icon}
                            <span>{kind === 'other' ? formatLabel(perm.name) : meta.label}</span>
                            <span className="font-mono text-[10px] opacity-60">#{perm.id}</span>
                          </button>
                        )
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="shadow-xs overflow-hidden">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/50">
                  <TableHead className="text-xs px-4 py-3 w-[80px]">ID</TableHead>
                  <TableHead className="text-xs px-4 py-3">Permission</TableHead>
                  <TableHead className="text-xs px-4 py-3">Kode Key</TableHead>
                  <TableHead className="text-xs px-4 py-3">Kategori Modul</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {paginatedPermissions.map((perm) => {
                  const meta = ACTION_META[getActionKind(perm.name)]
                  return (
                    <TableRow
                      key={perm.id}
                      className="hover:bg-muted/40 cursor-pointer"
                      onClick={() => { setCopiedKey(false); setDetailItem(perm) }}
                    >
                      <TableCell className="px-4 py-2.5 font-mono text-xs text-muted-foreground">#{perm.id}</TableCell>
                      <TableCell className="px-4 py-2.5">
                        <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md text-xs font-medium border ${meta.cls}`}>
                          {meta.icon}
                          {formatLabel(perm.name)}
                        </span>
                      </TableCell>
                      <TableCell className="px-4 py-2.5 font-mono text-xs text-muted-foreground">{perm.name}</TableCell>
                      <TableCell className="px-4 py-2.5">
                        <Badge variant="outline" className="text-xs">{getPermissionCategory(perm.name)}</Badge>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
          <Separator />
          <div className="flex items-center justify-between px-4 py-3 text-xs text-muted-foreground">
            <span>
              Halaman {currentPage} dari {totalPages} ({filteredPermissions.length} permissions)
            </span>
            <div className="flex items-center gap-1">
              <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => p - 1)}>
                <ChevronLeft className="h-3.5 w-3.5" />
              </Button>
              <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage >= totalPages} onClick={() => setCurrentPage((p) => p + 1)}>
                <ChevronRight className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
        </Card>
      )}

      {/* Permission Detail Dialog */}
      <Dialog open={Boolean(detailItem)} onOpenChange={(o) => !o && setDetailItem(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Detail Permission</DialogTitle>
            <DialogDescription>Informasi kode hak akses RBAC</DialogDescription>
          </DialogHeader>
          {detailItem && (
            <div className="space-y-2.5 text-xs py-2">
              <div className="flex justify-between items-center border-b pb-2">
                <span className="text-muted-foreground">ID</span>
                <span className="font-mono font-bold">#{detailItem.id}</span>
              </div>
              <div className="flex justify-between items-center border-b pb-2 gap-2">
                <span className="text-muted-foreground shrink-0">Kode Key</span>
                <div className="flex items-center gap-1.5 min-w-0">
                  <span className="font-mono font-semibold text-foreground truncate">{detailItem.name}</span>
                  <button
                    type="button"
                    onClick={() => copyKey(detailItem.name)}
                    className="p-1 rounded hover:bg-muted text-muted-foreground"
                    title="Salin kode key"
                  >
                    {copiedKey ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
                  </button>
                </div>
              </div>
              <div className="flex justify-between items-center border-b pb-2">
                <span className="text-muted-foreground">Resource</span>
                <span className="font-medium">{getResourceKey(detailItem.name)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Kategori</span>
                <span className="font-medium">{getPermissionCategory(detailItem.name)}</span>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailItem(null)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
