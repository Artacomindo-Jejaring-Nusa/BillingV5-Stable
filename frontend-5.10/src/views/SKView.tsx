import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Archive,
  BookOpen,
  Calendar,
  ChevronDown,
  ChevronUp,
  Clock,
  FileText,
  Gavel,
  RefreshCw,
  Search,
  ShieldCheck,
  Sparkles,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'

interface SKItem {
  id: number
  judul: string
  konten: string
  tipe: string
  versi: string | null
  created_at: string
}

function formatDateLong(dateString: string) {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatDateShort(dateString: string) {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function formatContent(content: string): string {
  if (!content) return ''
  const isFullHTML = content.includes('<!DOCTYPE') || content.includes('<html')
  if (isFullHTML) {
    const styleRegex = /<style[^>]*>([\s\S]*?)<\/style>/gi
    const styles = content.match(styleRegex) || []
    const bodyMatch = content.match(/<body[^>]*>([\s\S]*?)<\/body>/i)
    const bodyContent = bodyMatch ? bodyMatch[1] : content
    return `${styles.join('\n')}\n${bodyContent}`
  }
  return content.replace(/\\n/g, '<br>').replace(/\n/g, '<br>')
}

function tipeBadgeClass(tipe: string) {
  switch (tipe) {
    case 'Pembaruan':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300'
    case 'Syarat & Ketentuan':
      return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300'
    case 'Kebijakan':
      return 'bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950/40 dark:text-sky-300'
    case 'Panduan':
      return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

function TipeIcon({ tipe }: { tipe: string }) {
  switch (tipe) {
    case 'Pembaruan':
      return <RefreshCw className="h-3.5 w-3.5" />
    case 'Syarat & Ketentuan':
      return <Gavel className="h-3.5 w-3.5" />
    case 'Kebijakan':
      return <ShieldCheck className="h-3.5 w-3.5" />
    case 'Panduan':
      return <BookOpen className="h-3.5 w-3.5" />
    default:
      return <FileText className="h-3.5 w-3.5" />
  }
}

function getStatusMeta(createdAt: string) {
  const days = Math.floor((Date.now() - new Date(createdAt).getTime()) / (1000 * 60 * 60 * 24))
  if (days <= 7) {
    return {
      text: 'Baru',
      cls: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300',
      Icon: Sparkles,
    }
  }
  if (days <= 30) {
    return {
      text: 'Terkini',
      cls: 'bg-sky-50 text-sky-700 border-sky-200 dark:bg-sky-950/40 dark:text-sky-300',
      Icon: Clock,
    }
  }
  return {
    text: 'Arsip',
    cls: 'bg-muted text-muted-foreground',
    Icon: Archive,
  }
}

export function SKView() {
  const [skItems, setSkItems] = useState<SKItem[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedType, setSelectedType] = useState('all')
  const [expandedIds, setExpandedIds] = useState<number[]>([])

  const fetchSK = useCallback(async () => {
    setLoading(true)
    try {
      const res = await apiClient.get('/sk')
      const list = Array.isArray(res.data) ? res.data : res.data?.data || []
      setSkItems(list)
    } catch (err) {
      console.error('Gagal mengambil data S&K:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchSK()
  }, [fetchSK])

  const typeOptions = useMemo(
    () => Array.from(new Set(skItems.map((i) => i.tipe).filter(Boolean))),
    [skItems],
  )

  const filteredItems = useMemo(() => {
    let list = [...skItems]
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      list = list.filter(
        (item) =>
          item.judul.toLowerCase().includes(q) ||
          item.konten.toLowerCase().includes(q) ||
          item.tipe.toLowerCase().includes(q),
      )
    }
    if (selectedType !== 'all') {
      list = list.filter((item) => item.tipe === selectedType)
    }
    return list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }, [skItems, searchQuery, selectedType])

  const updateCount = useMemo(() => skItems.filter((i) => i.tipe === 'Pembaruan').length, [skItems])
  const skCount = useMemo(() => skItems.filter((i) => i.tipe !== 'Pembaruan').length, [skItems])

  const lastUpdateLabel = useMemo(() => {
    if (!skItems.length) return '-'
    const latest = skItems.reduce((a, b) =>
      new Date(b.created_at) > new Date(a.created_at) ? b : a,
    )
    const days = Math.floor((Date.now() - new Date(latest.created_at).getTime()) / (1000 * 60 * 60 * 24))
    return days === 0 ? 'Hari ini' : `${days} hari`
  }, [skItems])

  const toggleExpand = (id: number) => {
    setExpandedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div className="flex items-center gap-3.5">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground">
          <FileText className="h-5 w-5" />
        </div>
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Syarat & Ketentuan</h2>
          <p className="text-sm text-muted-foreground">
            Informasi terkini tentang syarat ketentuan dan catatan rilis sistem
          </p>
        </div>
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-5 flex items-center gap-3.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-600 text-white">
              <FileText className="h-5 w-5" />
            </div>
            <div>
              <p className="text-2xl font-bold">{skItems.length}</p>
              <p className="text-xs text-muted-foreground">Total Dokumen</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 flex items-center gap-3.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-emerald-600 text-white">
              <RefreshCw className="h-5 w-5" />
            </div>
            <div>
              <p className="text-2xl font-bold">{updateCount}</p>
              <p className="text-xs text-muted-foreground">Pembaruan</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 flex items-center gap-3.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-sky-600 text-white">
              <Gavel className="h-5 w-5" />
            </div>
            <div>
              <p className="text-2xl font-bold">{skCount}</p>
              <p className="text-xs text-muted-foreground">S&K</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 flex items-center gap-3.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-amber-500 text-white">
              <Clock className="h-5 w-5" />
            </div>
            <div>
              <p className="text-2xl font-bold">{lastUpdateLabel}</p>
              <p className="text-xs text-muted-foreground">Update Terakhir</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search & Filter */}
      <Card>
        <CardContent className="pt-5">
          <div className="grid grid-cols-1 md:grid-cols-12 gap-3">
            <div className="relative md:col-span-8">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari dokumen..."
                className="pl-9"
              />
            </div>
            <div className="md:col-span-4">
              <Select value={selectedType} onValueChange={setSelectedType}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Filter berdasarkan tipe" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Tipe</SelectItem>
                  {typeOptions.map((t) => (
                    <SelectItem key={t} value={t}>
                      {t}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Documents List */}
      <Card>
        <CardContent className="p-0 divide-y">
          {loading ? (
            <div className="p-6 space-y-3">
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </div>
          ) : filteredItems.length === 0 ? (
            <div className="py-12 text-center space-y-2">
              <FileText className="h-12 w-12 mx-auto text-muted-foreground" />
              <p className="font-semibold">Tidak ada dokumen ditemukan</p>
              <p className="text-sm text-muted-foreground">
                {searchQuery || selectedType !== 'all'
                  ? 'Coba ubah kriteria pencarian Anda'
                  : 'Belum ada dokumen S&K yang tersedia'}
              </p>
            </div>
          ) : (
            filteredItems.map((item) => {
              const expanded = expandedIds.includes(item.id)
              const statusMeta = getStatusMeta(item.created_at)
              const StatusIcon = statusMeta.Icon
              return (
                <div key={item.id} className="transition-colors">
                  <button
                    type="button"
                    onClick={() => toggleExpand(item.id)}
                    className="w-full flex flex-wrap items-center justify-between gap-3 px-5 py-4 text-left hover:bg-muted/40 transition-colors"
                  >
                    <div className="flex flex-wrap items-center gap-2.5 min-w-0">
                      <Badge variant="outline" className={`gap-1 ${tipeBadgeClass(item.tipe)}`}>
                        <TipeIcon tipe={item.tipe} />
                        {item.tipe}
                      </Badge>
                      <span className="font-semibold text-sm">{item.judul}</span>
                    </div>
                    <div className="flex items-center gap-2.5 shrink-0">
                      {item.versi && (
                        <Badge variant="outline" className="text-[11px]">
                          v{item.versi}
                        </Badge>
                      )}
                      <span className="text-xs text-muted-foreground">{formatDateShort(item.created_at)}</span>
                      <Button variant="ghost" size="icon" className="h-7 w-7" tabIndex={-1}>
                        {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                      </Button>
                    </div>
                  </button>

                  {expanded && (
                    <div className="px-5 pb-5 pt-2 border-t bg-muted/20 space-y-4">
                      <div
                        className="prose prose-sm max-w-none dark:prose-invert text-sm leading-relaxed"
                        dangerouslySetInnerHTML={{ __html: formatContent(item.konten) }}
                      />
                      <div className="flex flex-wrap items-center justify-between gap-2 pt-3 border-t text-xs text-muted-foreground">
                        <span className="flex items-center gap-1.5">
                          <Calendar className="h-3.5 w-3.5" />
                          Dipublikasikan pada {formatDateLong(item.created_at)}
                        </span>
                        <Badge variant="outline" className={`gap-1 ${statusMeta.cls}`}>
                          <StatusIcon className="h-3 w-3" />
                          {statusMeta.text}
                        </Badge>
                      </div>
                    </div>
                  )}
                </div>
              )
            })
          )}
        </CardContent>
      </Card>
    </div>
  )
}
