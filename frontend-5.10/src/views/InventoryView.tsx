import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
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
  Archive,
  Package,
  Shapes,
  Tag,
  History,
  Plus,
  RefreshCw,
  Search,
  FilterX,
  FileSpreadsheet,
  Download,
  Upload,
  Pencil,
  Trash2,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
  ChevronLeft,
  ChevronRight,
  MapPin,
  Calendar,
  User,
} from 'lucide-react'

interface ItemType {
  id: number
  name: string
}

interface InventoryStatus {
  id: number
  name: string
}

interface InventoryItem {
  id: number
  serial_number: string
  mac_address: string | null
  location: string | null
  notes: string | null
  purchase_date: string | null
  item_type_id: number
  status_id: number
  item_type: ItemType
  status: InventoryStatus
}

interface HistoryLog {
  id?: number
  timestamp: string
  serial_number?: string
  action: string
  user?: {
    name?: string
    role?: string
  }
}

interface BulkImportResult {
  success: boolean
  message?: string
  success_count: number
  error_count?: number
  errors?: string[]
}

function getStatusBadgeClass(statusName = '') {
  const n = statusName.toLowerCase()
  if (n === 'terpasang') return 'bg-emerald-100 text-emerald-800 border-emerald-200'
  if (n === 'rusak' || n === 'hilang') return 'bg-red-100 text-red-800 border-red-200'
  if (n === 'dalam perbaikan' || n === 'perbaikan') return 'bg-amber-100 text-amber-800 border-amber-200'
  if (n === 'gudang') return 'bg-blue-100 text-blue-800 border-blue-200'
  if (n === 'dismantle') return 'bg-muted text-foreground/90 border-border'
  return 'bg-indigo-100 text-indigo-800 border-indigo-200'
}

function getActionMeta(action = '') {
  const lower = action.toLowerCase()
  const match = action.match(/^(Created|Updated|Deleted) item - (.+)$/i)
  const detail = match ? match[2] : action

  if (lower.includes('created')) {
    return { label: 'Dibuat', cls: 'bg-emerald-100 text-emerald-800 border-emerald-200', detail }
  }
  if (lower.includes('updated')) {
    return { label: 'Diubah', cls: 'bg-amber-100 text-amber-800 border-amber-200', detail }
  }
  if (lower.includes('deleted')) {
    return { label: 'Dihapus', cls: 'bg-red-100 text-red-800 border-red-200', detail }
  }
  return { label: 'Lainnya', cls: 'bg-muted text-foreground/90 border-border', detail }
}

function formatDateId(date?: string | null) {
  if (!date) return '-'
  const d = new Date(date)
  if (Number.isNaN(d.getTime())) return '-'
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function formatDateTimeId(ts?: string) {
  if (!ts) return '-'
  const d = new Date(ts)
  if (Number.isNaN(d.getTime())) return '-'
  return d.toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function InventoryView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_inventory')
  const canEdit = hasPermission('edit_inventory')
  const canDelete = hasPermission('delete_inventory')
  const [activeTab, setActiveTab] = useState<'items' | 'types' | 'statuses' | 'history'>('items')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  // Data
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([])
  const [itemTypes, setItemTypes] = useState<ItemType[]>([])
  const [statuses, setStatuses] = useState<InventoryStatus[]>([])
  const [globalHistoryLogs, setGlobalHistoryLogs] = useState<HistoryLog[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)

  // Filters - Items
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedType, setSelectedType] = useState<string>('')
  const [selectedStatus, setSelectedStatus] = useState<string>('')
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 15

  // Filters - Global History
  const [historySearchQuery, setHistorySearchQuery] = useState('')
  const [selectedHistoryType, setSelectedHistoryType] = useState<string>('')
  const [historyPage, setHistoryPage] = useState(1)

  // Item Dialog
  const [showItemDialog, setShowItemDialog] = useState(false)
  const [editedItem, setEditedItem] = useState<Partial<InventoryItem>>({})

  // Type Dialog
  const [showTypeDialog, setShowTypeDialog] = useState(false)
  const [editedType, setEditedType] = useState<Partial<ItemType>>({})

  // Status Dialog
  const [showStatusDialog, setShowStatusDialog] = useState(false)
  const [editedStatus, setEditedStatus] = useState<Partial<InventoryStatus>>({})

  // Item History Dialog
  const [showHistoryDialog, setShowHistoryDialog] = useState(false)
  const [selectedItemForHistory, setSelectedItemForHistory] = useState<InventoryItem | null>(null)
  const [itemHistoryLogs, setItemHistoryLogs] = useState<HistoryLog[]>([])
  const [itemHistoryLoading, setItemHistoryLoading] = useState(false)

  // Bulk Import Dialog
  const [showBulkDialog, setShowBulkDialog] = useState(false)
  const [bulkFile, setBulkFile] = useState<File | null>(null)
  const [bulkLoading, setBulkLoading] = useState(false)
  const [bulkResult, setBulkResult] = useState<BulkImportResult | null>(null)
  const [bulkError, setBulkError] = useState<string | null>(null)

  // Feedback
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4000)
  }

  const fetchGlobalHistory = useCallback(async () => {
    setHistoryLoading(true)
    try {
      const res = await apiClient.get('/inventory/history/all')
      setGlobalHistoryLogs(Array.isArray(res.data) ? res.data : res.data?.data || [])
    } catch {
      setGlobalHistoryLogs([])
    } finally {
      setHistoryLoading(false)
    }
  }, [])

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [itemsRes, typesRes, statusesRes] = await Promise.all([
        apiClient.get('/inventory?page_size=-1'),
        apiClient.get('/inventory-types'),
        apiClient.get('/inventory-statuses'),
      ])
      setInventoryItems(Array.isArray(itemsRes.data) ? itemsRes.data : itemsRes.data?.data || [])
      setItemTypes(Array.isArray(typesRes.data) ? typesRes.data : typesRes.data?.data || [])
      setStatuses(Array.isArray(statusesRes.data) ? statusesRes.data : statusesRes.data?.data || [])
      await fetchGlobalHistory()
    } catch {
      notify('Gagal memuat data inventaris', 'error')
    } finally {
      setLoading(false)
    }
  }, [fetchGlobalHistory])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  // Summary Stats
  const stats = useMemo(() => {
    const total = inventoryItems.length
    const terpasang = inventoryItems.filter((i) => i.status?.name?.toLowerCase() === 'terpasang').length
    const gudang = inventoryItems.filter((i) => i.status?.name?.toLowerCase() === 'gudang').length
    const bermasalah = inventoryItems.filter((i) => {
      const n = i.status?.name?.toLowerCase() || ''
      return n === 'rusak' || n === 'hilang' || n.includes('perbaikan')
    }).length
    return { total, terpasang, gudang, bermasalah }
  }, [inventoryItems])

  // Filtered Items
  const filteredItems = useMemo(() => {
    let items = inventoryItems
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase().trim()
      items = items.filter(
        (i) =>
          i.serial_number.toLowerCase().includes(q) ||
          (i.mac_address && i.mac_address.toLowerCase().includes(q)) ||
          (i.location && i.location.toLowerCase().includes(q))
      )
    }
    if (selectedType) {
      items = items.filter((i) => String(i.item_type_id) === selectedType)
    }
    if (selectedStatus) {
      items = items.filter((i) => String(i.status_id) === selectedStatus)
    }
    return items
  }, [inventoryItems, searchQuery, selectedType, selectedStatus])

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / itemsPerPage))
  const paginatedItems = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return filteredItems.slice(start, start + itemsPerPage)
  }, [filteredItems, currentPage])

  // Filtered Global History
  const filteredHistory = useMemo(() => {
    let logs = globalHistoryLogs
    if (historySearchQuery.trim()) {
      const q = historySearchQuery.toLowerCase().trim()
      logs = logs.filter(
        (l) =>
          (l.serial_number || '').toLowerCase().includes(q) ||
          (l.action || '').toLowerCase().includes(q)
      )
    }
    if (selectedHistoryType) {
      logs = logs.filter((l) => (l.action || '').toLowerCase().includes(selectedHistoryType))
    }
    return logs
  }, [globalHistoryLogs, historySearchQuery, selectedHistoryType])

  const totalHistoryPages = Math.max(1, Math.ceil(filteredHistory.length / itemsPerPage))
  const paginatedHistory = useMemo(() => {
    const start = (historyPage - 1) * itemsPerPage
    return filteredHistory.slice(start, start + itemsPerPage)
  }, [filteredHistory, historyPage])

  // Item CRUD
  const openItemDialog = (item?: InventoryItem) => {
    if (item) {
      setEditedItem({
        ...item,
        purchase_date: item.purchase_date ? item.purchase_date.slice(0, 10) : '',
      })
    } else {
      setEditedItem({
        serial_number: '',
        mac_address: '',
        location: '',
        notes: '',
        purchase_date: '',
        item_type_id: itemTypes[0]?.id,
        status_id: statuses[0]?.id,
      })
    }
    setShowItemDialog(true)
  }

  const saveItem = async () => {
    if (!editedItem.serial_number?.trim() || !editedItem.item_type_id || !editedItem.status_id) {
      notify('Serial Number, Tipe Item, dan Status wajib diisi', 'error')
      return
    }
    setSaving(true)
    try {
      const payload: Record<string, unknown> = {
        serial_number: editedItem.serial_number.trim(),
        mac_address: editedItem.mac_address?.trim() || null,
        location: editedItem.location?.trim() || null,
        notes: editedItem.notes?.trim() || null,
        purchase_date: editedItem.purchase_date || null,
        item_type_id: Number(editedItem.item_type_id),
        status_id: Number(editedItem.status_id),
      }
      if (editedItem.id) {
        await apiClient.patch(`/inventory/${editedItem.id}`, payload)
        notify('Item inventaris berhasil diperbarui', 'success')
      } else {
        await apiClient.post('/inventory', payload)
        notify('Item inventaris berhasil ditambahkan', 'success')
      }
      setShowItemDialog(false)
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menyimpan perangkat', 'error')
    } finally {
      setSaving(false)
    }
  }

  const deleteItem = async (item: InventoryItem) => {
    if (!window.confirm(`Hapus item SN: ${item.serial_number}?`)) return
    try {
      await apiClient.delete(`/inventory/${item.id}`)
      notify('Item inventaris berhasil dihapus', 'success')
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menghapus item', 'error')
    }
  }

  // Type CRUD
  const saveType = async () => {
    if (!editedType.name?.trim()) {
      notify('Nama tipe item wajib diisi', 'error')
      return
    }
    setSaving(true)
    try {
      const payload = { name: editedType.name.trim() }
      if (editedType.id) {
        await apiClient.patch(`/inventory-types/${editedType.id}`, payload)
        notify('Tipe item berhasil diperbarui', 'success')
      } else {
        await apiClient.post('/inventory-types', payload)
        notify('Tipe item berhasil ditambahkan', 'success')
      }
      setShowTypeDialog(false)
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menyimpan tipe item', 'error')
    } finally {
      setSaving(false)
    }
  }

  const deleteType = async (item: ItemType) => {
    if (!window.confirm(`Hapus tipe: ${item.name}?`)) return
    try {
      await apiClient.delete(`/inventory-types/${item.id}`)
      notify('Tipe item berhasil dihapus', 'success')
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menghapus tipe item', 'error')
    }
  }

  // Status CRUD
  const saveStatus = async () => {
    if (!editedStatus.name?.trim()) {
      notify('Nama status wajib diisi', 'error')
      return
    }
    setSaving(true)
    try {
      const payload = { name: editedStatus.name.trim() }
      if (editedStatus.id) {
        await apiClient.patch(`/inventory-statuses/${editedStatus.id}`, payload)
        notify('Status berhasil diperbarui', 'success')
      } else {
        await apiClient.post('/inventory-statuses', payload)
        notify('Status berhasil ditambahkan', 'success')
      }
      setShowStatusDialog(false)
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menyimpan status', 'error')
    } finally {
      setSaving(false)
    }
  }

  const deleteStatus = async (item: InventoryStatus) => {
    if (!window.confirm(`Hapus status: ${item.name}?`)) return
    try {
      await apiClient.delete(`/inventory-statuses/${item.id}`)
      notify('Status berhasil dihapus', 'success')
      await fetchData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { detail?: string; error?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menghapus status', 'error')
    }
  }

  // Item History
  const openItemHistory = async (item: InventoryItem) => {
    setSelectedItemForHistory(item)
    setShowHistoryDialog(true)
    setItemHistoryLoading(true)
    try {
      const res = await apiClient.get(`/inventory/${item.id}/history`)
      setItemHistoryLogs(Array.isArray(res.data) ? res.data : res.data?.data || [])
    } catch {
      setItemHistoryLogs([])
    } finally {
      setItemHistoryLoading(false)
    }
  }

  // Export & Template
  const exportToExcel = () => {
    const rows = filteredItems
      .map(
        (item) =>
          `<tr><td>${item.serial_number}</td><td>${item.mac_address || '-'}</td><td>${item.item_type?.name || '-'}</td><td>${item.status?.name || '-'}</td><td>${item.location || '-'}</td><td>${formatDateId(item.purchase_date)}</td><td>${item.notes || '-'}</td></tr>`
      )
      .join('')
    const html = `<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel"><head><meta charset="UTF-8"/></head><body>
      <table border="1"><thead><tr><th>Serial Number</th><th>MAC Address</th><th>Tipe</th><th>Status</th><th>Lokasi</th><th>Tanggal Pembelian</th><th>Catatan</th></tr></thead><tbody>${rows}</tbody></table>
    </body></html>`
    const blob = new Blob([html], { type: 'application/vnd.ms-excel' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `inventaris_perangkat_${new Date().toISOString().split('T')[0]}.xls`
    a.click()
    URL.revokeObjectURL(url)
  }

  const downloadTemplate = () => {
    const csv =
      'Serial Number,MAC Address,Tipe Barang,Status,Lokasi,Tanggal Pembelian,Catatan\n' +
      'SN001234567890,AA:BB:CC:DD:EE:FF,ONT,Gudang,Gudang Utama,2024-01-15,Contoh catatan'
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `template_inventaris_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  // Bulk Import
  const handleBulkImport = async () => {
    if (!bulkFile) {
      setBulkError('Silakan pilih file terlebih dahulu')
      return
    }
    setBulkLoading(true)
    setBulkError(null)
    setBulkResult(null)
    try {
      const fd = new FormData()
      fd.append('file', bulkFile)
      const res = await apiClient.post('/inventory/bulk-import', fd, {
        headers: {
          'Content-Type': 'multipart/form-data',
          'X-Skip-Network-Interceptor': 'true',
        },
        timeout: 120000,
      })
      setBulkResult(res.data)
      if (res.data.success_count > 0) {
        await fetchData()
      }
      if (res.data.success && (!res.data.error_count || res.data.error_count === 0)) {
        setTimeout(() => setShowBulkDialog(false), 1800)
      }
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string; message?: string; detail?: string } }; message?: string }
      setBulkError(e.response?.data?.error || e.response?.data?.message || e.response?.data?.detail || e.message || 'Gagal mengimport file.')
    } finally {
      setBulkLoading(false)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-primary text-primary-foreground">
            <Archive className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Manajemen Inventaris</h1>
            <p className="text-xs text-muted-foreground">Kelola perangkat, tipe item, status, dan riwayat pergerakan</p>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={fetchData} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Compact Stats Strip */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-muted text-foreground/90"><Package className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Total Perangkat</div>
            <div className="text-lg font-bold text-foreground">{stats.total}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-emerald-50 text-emerald-600"><CheckCircle2 className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Terpasang</div>
            <div className="text-lg font-bold text-emerald-600">{stats.terpasang}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-blue-50 text-blue-600"><Archive className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Di Gudang</div>
            <div className="text-lg font-bold text-blue-600">{stats.gudang}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-red-50 text-red-600"><AlertCircle className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Rusak / Perbaikan</div>
            <div className="text-lg font-bold text-red-600">{stats.bermasalah}</div>
          </div>
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

      {/* Tabs Container */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-4">
        <TabsList className="grid grid-cols-2 sm:grid-cols-4 w-full sm:w-auto">
          <TabsTrigger value="items" className="gap-1.5 text-xs sm:text-sm">
            <Package className="h-4 w-4" /> Daftar Perangkat
          </TabsTrigger>
          <TabsTrigger value="types" className="gap-1.5 text-xs sm:text-sm">
            <Shapes className="h-4 w-4" /> Tipe Item
          </TabsTrigger>
          <TabsTrigger value="statuses" className="gap-1.5 text-xs sm:text-sm">
            <Tag className="h-4 w-4" /> Status
          </TabsTrigger>
          <TabsTrigger value="history" className="gap-1.5 text-xs sm:text-sm">
            <History className="h-4 w-4" /> History Pergerakan
          </TabsTrigger>
        </TabsList>

        {/* TAB 1: ITEMS */}
        <TabsContent value="items" className="space-y-4 mt-0">
          <Card className="shadow-xs">
            <CardHeader className="py-3.5 px-4">
              <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
                <div>
                  <CardTitle className="text-base font-semibold">Daftar Perangkat</CardTitle>
                  <p className="text-xs text-muted-foreground">Kelola semua perangkat inventaris ({filteredItems.length} item)</p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <Button variant="outline" size="sm" onClick={exportToExcel}>
                    <FileSpreadsheet className="h-4 w-4 mr-1.5 text-emerald-600" />
                    Export Excel
                  </Button>
                  <Button variant="outline" size="sm" onClick={downloadTemplate}>
                    <Download className="h-4 w-4 mr-1.5 text-blue-600" />
                    Template CSV
                  </Button>
                  {canCreate && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setBulkFile(null)
                        setBulkResult(null)
                        setBulkError(null)
                        setShowBulkDialog(true)
                      }}
                    >
                      <Upload className="h-4 w-4 mr-1.5 text-amber-600" />
                      Bulk Import
                    </Button>
                  )}
                  {canCreate && (
                    <Button size="sm" onClick={() => openItemDialog()}>
                      <Plus className="h-4 w-4 mr-1.5" />
                      Tambah Item
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="p-4 space-y-4">
              {/* Filter Bar */}
              <div className="flex flex-col md:flex-row gap-2.5">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    value={searchQuery}
                    onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1) }}
                    placeholder="Cari (SN, MAC, Lokasi)..."
                    className="pl-9 h-9"
                  />
                </div>
                <div className="grid grid-cols-2 md:flex gap-2.5">
                  <Select value={selectedType || '__all__'} onValueChange={(v) => { setSelectedType(v === '__all__' ? '' : v); setCurrentPage(1) }}>
                    <SelectTrigger className="h-9 md:w-[170px] w-full"><SelectValue placeholder="Semua Tipe" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__all__">Semua Tipe</SelectItem>
                      {itemTypes.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}
                    </SelectContent>
                  </Select>
                  <Select value={selectedStatus || '__all__'} onValueChange={(v) => { setSelectedStatus(v === '__all__' ? '' : v); setCurrentPage(1) }}>
                    <SelectTrigger className="h-9 md:w-[170px] w-full"><SelectValue placeholder="Semua Status" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__all__">Semua Status</SelectItem>
                      {statuses.map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>)}
                    </SelectContent>
                  </Select>
                  {(searchQuery || selectedType || selectedStatus) && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => { setSearchQuery(''); setSelectedType(''); setSelectedStatus(''); setCurrentPage(1) }}
                      className="h-9 text-xs text-red-600"
                    >
                      <FilterX className="h-3.5 w-3.5 mr-1" /> Reset
                    </Button>
                  )}
                </div>
              </div>

              {/* Table */}
              {loading && inventoryItems.length === 0 ? (
                <div className="py-16 text-center text-xs text-muted-foreground">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
                  Memuat data inventaris...
                </div>
              ) : (
                <>
                  <div className="border rounded-lg overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow className="bg-muted/50">
                          <TableHead className="text-xs px-4 py-3">Serial Number</TableHead>
                          <TableHead className="text-xs px-4 py-3">MAC Address</TableHead>
                          <TableHead className="text-xs px-4 py-3">Tipe</TableHead>
                          <TableHead className="text-xs px-4 py-3">Status</TableHead>
                          <TableHead className="text-xs px-4 py-3">Lokasi</TableHead>
                          <TableHead className="text-xs px-4 py-3">Tgl Pembelian</TableHead>
                          <TableHead className="text-xs px-4 py-3 text-right">Aksi</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {paginatedItems.length === 0 ? (
                          <TableRow>
                            <TableCell colSpan={7} className="text-center py-12 text-xs text-muted-foreground">
                              Belum ada perangkat inventaris
                            </TableCell>
                          </TableRow>
                        ) : (
                          paginatedItems.map((item) => (
                            <TableRow key={item.id} className="hover:bg-muted/40">
                              <TableCell className="px-4 py-2.5 font-mono text-xs font-semibold text-foreground">
                                {item.serial_number}
                              </TableCell>
                              <TableCell className="px-4 py-2.5 font-mono text-xs text-muted-foreground">
                                {item.mac_address || '-'}
                              </TableCell>
                              <TableCell className="px-4 py-2.5 text-xs">
                                <Badge variant="outline">{item.item_type?.name || '-'}</Badge>
                              </TableCell>
                              <TableCell className="px-4 py-2.5">
                                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClass(item.status?.name)}`}>
                                  {item.status?.name || '-'}
                                </span>
                              </TableCell>
                              <TableCell className="px-4 py-2.5 text-xs text-foreground/90">
                                {item.location ? (
                                  <span className="inline-flex items-center gap-1"><MapPin className="h-3 w-3 text-muted-foreground" /> {item.location}</span>
                                ) : '-'}
                              </TableCell>
                              <TableCell className="px-4 py-2.5 text-xs text-muted-foreground">
                                {item.purchase_date ? (
                                  <span className="inline-flex items-center gap-1"><Calendar className="h-3 w-3 text-muted-foreground" /> {formatDateId(item.purchase_date)}</span>
                                ) : '-'}
                              </TableCell>
                              <TableCell className="px-4 py-2.5 text-right">
                                <div className="flex items-center justify-end gap-1">
                                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openItemHistory(item)} title="History">
                                    <History className="h-3.5 w-3.5 text-cyan-600" />
                                  </Button>
                                  {canEdit && (
                                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openItemDialog(item)} title="Edit">
                                      <Pencil className="h-3.5 w-3.5 text-blue-600" />
                                    </Button>
                                  )}
                                  {canDelete && (
                                    <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => deleteItem(item)} title="Hapus">
                                      <Trash2 className="h-3.5 w-3.5 text-red-500" />
                                    </Button>
                                  )}
                                </div>
                              </TableCell>
                            </TableRow>
                          ))
                        )}
                      </TableBody>
                    </Table>
                  </div>

                  {/* Pagination */}
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>Halaman {currentPage} dari {totalPages}</span>
                    <div className="flex items-center gap-1">
                      <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage <= 1} onClick={() => setCurrentPage((p) => p - 1)}>
                        <ChevronLeft className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="outline" size="icon" className="h-7 w-7" disabled={currentPage >= totalPages} onClick={() => setCurrentPage((p) => p + 1)}>
                        <ChevronRight className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* TAB 2: ITEM TYPES */}
        <TabsContent value="types" className="mt-0">
          <Card className="shadow-xs">
            <CardHeader className="py-3.5 px-4 flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base font-semibold">Tipe Item yang Tersedia</CardTitle>
                <p className="text-xs text-muted-foreground">Kelola kategori perangkat inventaris</p>
              </div>
              {canCreate && (
                <Button size="sm" onClick={() => { setEditedType({ name: '' }); setShowTypeDialog(true) }}>
                  <Plus className="h-4 w-4 mr-1.5" /> Tambah Tipe Baru
                </Button>
              )}
            </CardHeader>
            <Separator />
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="text-xs px-4 py-3 w-[80px]">ID</TableHead>
                    <TableHead className="text-xs px-4 py-3">Nama Tipe</TableHead>
                    <TableHead className="text-xs px-4 py-3 text-right w-[140px]">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {itemTypes.length === 0 ? (
                    <TableRow><TableCell colSpan={3} className="text-center py-10 text-xs text-muted-foreground">Belum ada tipe item</TableCell></TableRow>
                  ) : (
                    itemTypes.map((t) => (
                      <TableRow key={t.id}>
                        <TableCell className="px-4 py-2.5 font-mono text-xs text-muted-foreground">#{t.id}</TableCell>
                        <TableCell className="px-4 py-2.5 text-sm font-medium text-foreground">{t.name}</TableCell>
                        <TableCell className="px-4 py-2.5 text-right">
                          {canEdit && (
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => { setEditedType({ ...t }); setShowTypeDialog(true) }}>
                              <Pencil className="h-3.5 w-3.5 text-blue-600" />
                            </Button>
                          )}
                          {canDelete && (
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => deleteType(t)}>
                              <Trash2 className="h-3.5 w-3.5 text-red-500" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </Card>
        </TabsContent>

        {/* TAB 3: STATUSES */}
        <TabsContent value="statuses" className="mt-0">
          <Card className="shadow-xs">
            <CardHeader className="py-3.5 px-4 flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base font-semibold">Status Inventaris</CardTitle>
                <p className="text-xs text-muted-foreground">Kelola status kondisi perangkat</p>
              </div>
              {canCreate && (
                <Button size="sm" onClick={() => { setEditedStatus({ name: '' }); setShowStatusDialog(true) }}>
                  <Plus className="h-4 w-4 mr-1.5" /> Tambah Status Baru
                </Button>
              )}
            </CardHeader>
            <Separator />
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="text-xs px-4 py-3 w-[80px]">ID</TableHead>
                    <TableHead className="text-xs px-4 py-3">Nama Status</TableHead>
                    <TableHead className="text-xs px-4 py-3 text-right w-[140px]">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {statuses.length === 0 ? (
                    <TableRow><TableCell colSpan={3} className="text-center py-10 text-xs text-muted-foreground">Belum ada status inventaris</TableCell></TableRow>
                  ) : (
                    statuses.map((s) => (
                      <TableRow key={s.id}>
                        <TableCell className="px-4 py-2.5 font-mono text-xs text-muted-foreground">#{s.id}</TableCell>
                        <TableCell className="px-4 py-2.5">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getStatusBadgeClass(s.name)}`}>
                            {s.name}
                          </span>
                        </TableCell>
                        <TableCell className="px-4 py-2.5 text-right">
                          {canEdit && (
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => { setEditedStatus({ ...s }); setShowStatusDialog(true) }}>
                              <Pencil className="h-3.5 w-3.5 text-blue-600" />
                            </Button>
                          )}
                          {canDelete && (
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => deleteStatus(s)}>
                              <Trash2 className="h-3.5 w-3.5 text-red-500" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </Card>
        </TabsContent>

        {/* TAB 4: GLOBAL HISTORY */}
        <TabsContent value="history" className="mt-0">
          <Card className="shadow-xs">
            <CardHeader className="py-3.5 px-4 flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base font-semibold">History Pergerakan Perangkat</CardTitle>
                <p className="text-xs text-muted-foreground">Semua riwayat pembuatan, perubahan status, dan penghapusan perangkat</p>
              </div>
              <Button variant="outline" size="sm" onClick={fetchGlobalHistory} disabled={historyLoading}>
                <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${historyLoading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
            </CardHeader>
            <Separator />
            <CardContent className="p-4 space-y-4">
              <div className="flex flex-col sm:flex-row gap-2.5">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    value={historySearchQuery}
                    onChange={(e) => { setHistorySearchQuery(e.target.value); setHistoryPage(1) }}
                    placeholder="Cari SN atau detail aksi..."
                    className="pl-9 h-9"
                  />
                </div>
                <Select value={selectedHistoryType || '__all__'} onValueChange={(v) => { setSelectedHistoryType(v === '__all__' ? '' : v); setHistoryPage(1) }}>
                  <SelectTrigger className="h-9 sm:w-[170px] w-full"><SelectValue placeholder="Semua Aksi" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__all__">Semua Aksi</SelectItem>
                    <SelectItem value="created">Dibuat</SelectItem>
                    <SelectItem value="updated">Diubah</SelectItem>
                    <SelectItem value="deleted">Dihapus</SelectItem>
                  </SelectContent>
                </Select>
                {(historySearchQuery || selectedHistoryType) && (
                  <Button variant="ghost" size="sm" onClick={() => { setHistorySearchQuery(''); setSelectedHistoryType(''); setHistoryPage(1) }} className="h-9 text-xs text-red-600">
                    <FilterX className="h-3.5 w-3.5 mr-1" /> Reset
                  </Button>
                )}
              </div>

              <div className="border rounded-lg overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead className="text-xs px-4 py-3">Waktu</TableHead>
                      <TableHead className="text-xs px-4 py-3">Serial Number</TableHead>
                      <TableHead className="text-xs px-4 py-3">Aksi & Detail</TableHead>
                      <TableHead className="text-xs px-4 py-3">Oleh</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedHistory.length === 0 ? (
                      <TableRow><TableCell colSpan={4} className="text-center py-10 text-xs text-muted-foreground">Belum ada riwayat pergerakan</TableCell></TableRow>
                    ) : (
                      paginatedHistory.map((log, idx) => {
                        const meta = getActionMeta(log.action)
                        return (
                          <TableRow key={log.id || idx}>
                            <TableCell className="px-4 py-2.5 text-xs text-muted-foreground whitespace-nowrap">
                              {formatDateTimeId(log.timestamp)}
                            </TableCell>
                            <TableCell className="px-4 py-2.5 font-mono text-xs font-semibold text-foreground">
                              {log.serial_number || '-'}
                            </TableCell>
                            <TableCell className="px-4 py-2.5">
                              <div className="flex flex-col items-start gap-1">
                                <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-semibold border ${meta.cls}`}>
                                  {meta.label}
                                </span>
                                <span className="text-xs text-muted-foreground">{meta.detail}</span>
                              </div>
                            </TableCell>
                            <TableCell className="px-4 py-2.5 text-xs">
                              <div className="flex items-center gap-1.5">
                                <User className="h-3.5 w-3.5 text-muted-foreground" />
                                <span className="font-medium text-foreground">{log.user?.name || 'System'}</span>
                              </div>
                            </TableCell>
                          </TableRow>
                        )
                      })
                    )}
                  </TableBody>
                </Table>
              </div>

              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Halaman {historyPage} dari {totalHistoryPages}</span>
                <div className="flex items-center gap-1">
                  <Button variant="outline" size="icon" className="h-7 w-7" disabled={historyPage <= 1} onClick={() => setHistoryPage((p) => p - 1)}>
                    <ChevronLeft className="h-3.5 w-3.5" />
                  </Button>
                  <Button variant="outline" size="icon" className="h-7 w-7" disabled={historyPage >= totalHistoryPages} onClick={() => setHistoryPage((p) => p + 1)}>
                    <ChevronRight className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Add / Edit Item Dialog */}
      <Dialog open={showItemDialog} onOpenChange={setShowItemDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editedItem.id ? 'Edit Perangkat' : 'Tambah Item Inventaris'}</DialogTitle>
            <DialogDescription>Lengkapi informasi serial number, tipe, dan status perangkat</DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Serial Number *</Label>
                <Input
                  value={editedItem.serial_number || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, serial_number: e.target.value }))}
                  placeholder="SN00123456"
                  className="font-mono"
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">MAC Address</Label>
                <Input
                  value={editedItem.mac_address || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, mac_address: e.target.value }))}
                  placeholder="AA:BB:CC:DD:EE:FF"
                  className="font-mono"
                />
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Tipe Item *</Label>
                <Select
                  value={editedItem.item_type_id ? String(editedItem.item_type_id) : ''}
                  onValueChange={(v) => setEditedItem((p) => ({ ...p, item_type_id: Number(v) }))}
                >
                  <SelectTrigger className="w-full"><SelectValue placeholder="Pilih tipe..." /></SelectTrigger>
                  <SelectContent>
                    {itemTypes.map((t) => <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label className="text-sm mb-1 block">Status *</Label>
                <Select
                  value={editedItem.status_id ? String(editedItem.status_id) : ''}
                  onValueChange={(v) => setEditedItem((p) => ({ ...p, status_id: Number(v) }))}
                >
                  <SelectTrigger className="w-full"><SelectValue placeholder="Pilih status..." /></SelectTrigger>
                  <SelectContent>
                    {statuses.map((s) => <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Lokasi</Label>
                <Input
                  value={editedItem.location || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, location: e.target.value }))}
                  placeholder="Gudang Utama"
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">Tanggal Pembelian</Label>
                <Input
                  type="date"
                  value={editedItem.purchase_date || ''}
                  onChange={(e) => setEditedItem((p) => ({ ...p, purchase_date: e.target.value }))}
                />
              </div>
            </div>
            <div>
              <Label className="text-sm mb-1 block">Catatan</Label>
              <Textarea
                value={editedItem.notes || ''}
                onChange={(e) => setEditedItem((p) => ({ ...p, notes: e.target.value }))}
                placeholder="Catatan tambahan..."
                rows={2}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowItemDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveItem} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Simpan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add / Edit Type Dialog */}
      <Dialog open={showTypeDialog} onOpenChange={setShowTypeDialog}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{editedType.id ? 'Edit Tipe Item' : 'Tambah Tipe Baru'}</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <Label className="text-sm mb-1 block">Nama Tipe *</Label>
            <Input
              value={editedType.name || ''}
              onChange={(e) => setEditedType((p) => ({ ...p, name: e.target.value }))}
              placeholder="Contoh: ONT, Router, Switch"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTypeDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveType} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Simpan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add / Edit Status Dialog */}
      <Dialog open={showStatusDialog} onOpenChange={setShowStatusDialog}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{editedStatus.id ? 'Edit Status' : 'Tambah Status Baru'}</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <Label className="text-sm mb-1 block">Nama Status *</Label>
            <Input
              value={editedStatus.name || ''}
              onChange={(e) => setEditedStatus((p) => ({ ...p, name: e.target.value }))}
              placeholder="Contoh: Gudang, Terpasang, Rusak"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowStatusDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveStatus} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Simpan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Item History Dialog */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="sm:max-w-2xl max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <History className="h-5 w-5 text-blue-600" />
              Riwayat Perangkat — <span className="font-mono">{selectedItemForHistory?.serial_number}</span>
            </DialogTitle>
            <DialogDescription>
              {selectedItemForHistory?.item_type?.name || '-'} • Lokasi: {selectedItemForHistory?.location || '-'}
            </DialogDescription>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto py-2">
            {itemHistoryLoading ? (
              <div className="py-10 text-center text-xs text-muted-foreground">
                <Loader2 className="h-5 w-5 animate-spin mx-auto mb-2 text-blue-600" />
                Memuat riwayat...
              </div>
            ) : itemHistoryLogs.length === 0 ? (
              <div className="py-10 text-center text-xs text-muted-foreground">Belum ada riwayat pergerakan untuk perangkat ini</div>
            ) : (
              <div className="space-y-2.5">
                {itemHistoryLogs.map((log, i) => {
                  const meta = getActionMeta(log.action)
                  return (
                    <div key={log.id || i} className="border rounded-lg p-3 text-xs space-y-1">
                      <div className="flex items-center justify-between">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full font-semibold border ${meta.cls}`}>
                          {meta.label}
                        </span>
                        <span className="text-muted-foreground">{formatDateTimeId(log.timestamp)}</span>
                      </div>
                      <div className="text-foreground/90">{meta.detail}</div>
                      <div className="text-muted-foreground">Oleh: {log.user?.name || 'System'}</div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowHistoryDialog(false)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Bulk Import Dialog */}
      <Dialog open={showBulkDialog} onOpenChange={setShowBulkDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5 text-amber-600" />
              Bulk Import Inventaris
            </DialogTitle>
            <DialogDescription>Import data inventaris dari file Excel (.xlsx/.xls) atau CSV</DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <Input
              type="file"
              accept=".xlsx,.xls,.csv"
              disabled={bulkLoading}
              onChange={(e) => setBulkFile(e.target.files?.[0] || null)}
            />
            {bulkError && (
              <div className="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg p-3">{bulkError}</div>
            )}
            {bulkResult && (
              <div className={`border rounded-lg p-3 text-xs space-y-1.5 ${bulkResult.error_count ? 'bg-amber-50 border-amber-200 text-amber-800' : 'bg-emerald-50 border-emerald-200 text-emerald-800'}`}>
                <div className="font-semibold">
                  Berhasil: {bulkResult.success_count || 0} item • Gagal: {bulkResult.error_count || 0} item
                </div>
                {bulkResult.errors && bulkResult.errors.length > 0 && (
                  <div className="max-h-32 overflow-y-auto space-y-1 pt-1 border-t border-amber-200">
                    {bulkResult.errors.map((err, i) => (
                      <div key={i}>• {err}</div>
                    ))}
                  </div>
                )}
              </div>
            )}
            <p className="text-[11px] text-muted-foreground">
              Pastikan kolom file sesuai template: Serial Number, MAC Address, Tipe Barang, Status, Lokasi, Tanggal Pembelian, Catatan.
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)} disabled={bulkLoading}>Batal</Button>
            <Button onClick={handleBulkImport} disabled={bulkLoading || !bulkFile}>
              {bulkLoading && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Import Data
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
