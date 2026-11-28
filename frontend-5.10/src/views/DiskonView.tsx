import { useCallback, useEffect, useMemo, useState } from 'react'
import { Edit, Loader2, Pause, Percent, Play, Plus, RefreshCw, Search, Trash2, X } from 'lucide-react'
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

export interface DiskonItem {
  id: number
  nama_diskon: string
  persentase_diskon: number
  cluster: string
  is_active: boolean
  tgl_mulai?: string | null
  tgl_selesai?: string | null
}

const emptyForm = (): Partial<DiskonItem> => ({
  nama_diskon: '',
  persentase_diskon: 10,
  cluster: '',
  is_active: true,
  tgl_mulai: '',
  tgl_selesai: '',
})

export function DiskonView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_diskon')
  const canEdit = hasPermission('edit_diskon')
  const canDelete = hasPermission('delete_diskon')
  const [list, setList] = useState<DiskonItem[]>([])
  const [clusters, setClusters] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all')

  const [formOpen, setFormOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form, setForm] = useState<Partial<DiskonItem>>(emptyForm())
  const [formError, setFormError] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<DiskonItem | null>(null)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const fetchDiskon = useCallback(async () => {
    setLoading(true)
    try {
      const [dRes, cRes] = await Promise.allSettled([
        apiClient.get('/diskon'),
        apiClient.get<string[]>('/diskon/clusters/list'),
      ])
      if (dRes.status === 'fulfilled') {
        const raw = dRes.value.data?.data ?? dRes.value.data
        setList(Array.isArray(raw) ? raw : [])
      }
      if (cRes.status === 'fulfilled' && Array.isArray(cRes.value.data) && cRes.value.data.length > 0) {
        setClusters(cRes.value.data)
      } else {
        try {
          const locRes = await apiClient.get<string[]>('/pelanggan/lokasi/unik')
          if (Array.isArray(locRes.data)) setClusters(locRes.data)
        } catch {
          // ignore
        }
      }
    } catch {
      setFeedback({ type: 'error', text: 'Gagal memuat daftar diskon.' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchDiskon()
  }, [fetchDiskon])

  const filteredList = useMemo(() => {
    const q = searchQuery.toLowerCase().trim()
    return list.filter((item) => {
      if (q && !item.nama_diskon.toLowerCase().includes(q) && !item.cluster.toLowerCase().includes(q)) {
        return false
      }
      if (statusFilter === 'active' && !item.is_active) return false
      if (statusFilter === 'inactive' && item.is_active) return false
      return true
    })
  }, [list, searchQuery, statusFilter])

  const openCreate = () => {
    setEditingId(null)
    setForm(emptyForm())
    setFormError('')
    setFormOpen(true)
  }

  const openEdit = (item: DiskonItem) => {
    setEditingId(item.id)
    setForm({
      ...item,
      tgl_mulai: item.tgl_mulai ? new Date(item.tgl_mulai).toISOString().split('T')[0] : '',
      tgl_selesai: item.tgl_selesai ? new Date(item.tgl_selesai).toISOString().split('T')[0] : '',
    })
    setFormError('')
    setFormOpen(true)
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.nama_diskon?.trim() || !form.cluster?.trim()) {
      setFormError('Nama diskon dan cluster wajib diisi.')
      return
    }
    const pct = Number(form.persentase_diskon)
    if (Number.isNaN(pct) || pct <= 0 || pct > 100) {
      setFormError('Persentase diskon harus antara 1% hingga 100%.')
      return
    }
    setSaving(true)
    setFormError('')
    try {
      if (editingId) {
        await apiClient.put(`/diskon/${editingId}`, form)
        setFeedback({ type: 'success', text: 'Diskon berhasil diperbarui.' })
      } else {
        await apiClient.post('/diskon', form)
        setFeedback({ type: 'success', text: 'Diskon berhasil ditambahkan.' })
      }
      await fetchDiskon()
      setFormOpen(false)
    } catch {
      setFormError('Gagal menyimpan diskon.')
    } finally {
      setSaving(false)
    }
  }

  const toggleActive = async (item: DiskonItem) => {
    try {
      if (item.is_active) {
        await apiClient.put(`/diskon/${item.id}`, { is_active: false })
      } else {
        await apiClient.post(`/diskon/${item.id}/activate`)
      }
      await fetchDiskon()
    } catch {
      setFeedback({ type: 'error', text: 'Gagal mengubah status diskon.' })
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      await apiClient.delete(`/diskon/${deleteTarget.id}`)
      await fetchDiskon()
      setDeleteTarget(null)
      setFeedback({ type: 'success', text: 'Diskon berhasil dihapus.' })
    } catch {
      setFeedback({ type: 'error', text: 'Gagal menghapus diskon.' })
      setDeleteTarget(null)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Manajemen Diskon</h2>
          <p className="text-sm text-muted-foreground">
            Kelola diskon berdasarkan cluster/alamat pelanggan
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="gap-1">
            <Percent className="h-3.5 w-3.5" />
            {list.length} Diskon
          </Badge>
          {canCreate && (
            <Button size="sm" onClick={openCreate}>
              <Plus className="mr-1.5 h-4 w-4" />
              Tambah Diskon
            </Button>
          )}
        </div>
      </div>

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

      <Card>
        <CardContent className="pt-5">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <Search className="h-4 w-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Cari nama diskon atau cluster..."
                className="pl-9"
              />
            </div>
            <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as typeof statusFilter)}>
              <SelectTrigger className="w-full sm:w-44">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua Status</SelectItem>
                <SelectItem value="active">Aktif</SelectItem>
                <SelectItem value="inactive">Nonaktif</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" onClick={() => void fetchDiskon()} disabled={loading}>
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        <CardContent className="p-0">
          {loading ? (
            <div className="p-4 space-y-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : filteredList.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">
              Belum ada diskon yang sesuai.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Nama Diskon</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Persentase</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Cluster</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Status</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide">Periode</TableHead>
                    <TableHead className="px-4 py-3 text-xs uppercase tracking-wide text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredList.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="px-4 py-3.5 font-medium">{item.nama_diskon}</TableCell>
                      <TableCell className="px-4 py-3.5">
                        <Badge variant="secondary">{item.persentase_diskon}%</Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-sm">{item.cluster}</TableCell>
                      <TableCell className="px-4 py-3.5">
                        <button type="button" onClick={() => void toggleActive(item)} title="Klik untuk ubah status">
                          <Badge variant={item.is_active ? 'secondary' : 'outline'} className={item.is_active ? 'text-emerald-600' : ''}>
                            {item.is_active ? 'Aktif' : 'Nonaktif'}
                          </Badge>
                        </button>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs text-muted-foreground">
                        {item.tgl_mulai ? new Date(item.tgl_mulai).toLocaleDateString('id-ID') : '-'} s/d{' '}
                        {item.tgl_selesai ? new Date(item.tgl_selesai).toLocaleDateString('id-ID') : '-'}
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-right whitespace-nowrap">
                        {canEdit && (
                          <Button
                            size="icon"
                            variant="ghost"
                            onClick={() => void toggleActive(item)}
                            title={item.is_active ? 'Nonaktifkan Diskon' : 'Aktifkan Diskon'}
                            className={item.is_active ? 'text-amber-600' : 'text-emerald-600'}
                          >
                            {item.is_active ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                          </Button>
                        )}
                        {canEdit && (
                          <Button size="icon" variant="ghost" onClick={() => openEdit(item)} title="Edit">
                            <Edit className="h-4 w-4" />
                          </Button>
                        )}
                        {canDelete && (
                          <Button
                            size="icon"
                            variant="ghost"
                            className="text-destructive hover:text-destructive"
                            onClick={() => setDeleteTarget(item)}
                            title="Hapus"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Add / Edit Dialog */}
      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingId ? 'Edit Diskon' : 'Tambah Diskon Baru'}</DialogTitle>
            <DialogDescription>Atur nama, persentase diskon, dan cluster tujuan.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSave} className="space-y-4">
            {formError && (
              <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {formError}
              </div>
            )}
            <div className="space-y-1.5">
              <Label htmlFor="nama_diskon">Nama Diskon *</Label>
              <Input
                id="nama_diskon"
                value={form.nama_diskon || ''}
                onChange={(e) => setForm({ ...form, nama_diskon: e.target.value })}
                placeholder="contoh: Promo Cluster Parama"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="persentase_diskon">Persentase (%) *</Label>
                <Input
                  id="persentase_diskon"
                  type="number"
                  min={1}
                  max={100}
                  value={form.persentase_diskon ?? 10}
                  onChange={(e) => setForm({ ...form, persentase_diskon: Number(e.target.value) })}
                />
              </div>
              <div className="space-y-1.5">
                <Label>
                  Cluster / Alamat <span className="text-destructive">*</span>
                </Label>
                <Select
                  value={form.cluster || ''}
                  onValueChange={(v) => setForm({ ...form, cluster: v })}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Pilih Cluster / Alamat" />
                  </SelectTrigger>
                  <SelectContent>
                    {clusters.map((c) => (
                      <SelectItem key={c} value={c}>
                        {c}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="tgl_mulai_d">Tanggal Mulai</Label>
                <Input
                  id="tgl_mulai_d"
                  type="date"
                  value={form.tgl_mulai || ''}
                  onChange={(e) => setForm({ ...form, tgl_mulai: e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="tgl_selesai_d">Tanggal Berakhir</Label>
                <Input
                  id="tgl_selesai_d"
                  type="date"
                  min={form.tgl_mulai || undefined}
                  value={form.tgl_selesai || ''}
                  onChange={(e) => setForm({ ...form, tgl_selesai: e.target.value })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setFormOpen(false)}>
                Batal
              </Button>
              <Button type="submit" disabled={saving}>
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {saving ? 'Menyimpan...' : 'Simpan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Diskon</AlertDialogTitle>
            <AlertDialogDescription>
              Yakin ingin menghapus diskon "{deleteTarget?.nama_diskon}"?
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
