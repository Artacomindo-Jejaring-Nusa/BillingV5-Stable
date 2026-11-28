import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertCircle,
  CheckCircle2,
  FileText,
  History,
  Info,
  List,
  Loader2,
  Pencil,
  Plus,
  RefreshCw,
  Save,
  Tag,
  Trash2,
  X,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'

interface SKItem {
  id: number
  judul: string
  konten: string
  tipe: string
  versi: string | null
  created_at: string
}

interface FormState {
  id?: number
  judul: string
  konten: string
  tipe: string
  versi: string
}

function formatDate(dateString: string) {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

function formatTime(dateString: string) {
  if (!dateString) return ''
  return new Date(dateString).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function SKManagementView() {
  const [skItems, setSkItems] = useState<SKItem[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState<FormState>({
    judul: '',
    konten: '',
    tipe: 'Ketentuan',
    versi: '',
  })
  const [formErrors, setFormErrors] = useState<Partial<Record<keyof FormState, string>>>({})
  const [itemToDelete, setItemToDelete] = useState<SKItem | null>(null)

  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 10

  const showFeedback = useCallback((text: string, type: 'success' | 'error' = 'success') => {
    setFeedback({ type, text })
  }, [])

  useEffect(() => {
    if (!feedback) return
    const timer = setTimeout(() => setFeedback(null), 3500)
    return () => clearTimeout(timer)
  }, [feedback])

  const fetchSK = useCallback(async () => {
    setLoading(true)
    try {
      const response = await apiClient.get('/sk')
      const items = Array.isArray(response.data) ? response.data : (response.data?.data || [])
      setSkItems(items)
    } catch (error) {
      console.error('Gagal mengambil data S&K:', error)
      showFeedback('Gagal memuat data S&K', 'error')
    } finally {
      setLoading(false)
    }
  }, [showFeedback])

  useEffect(() => {
    void fetchSK()
  }, [fetchSK])

  const totalPages = Math.max(1, Math.ceil(skItems.length / itemsPerPage))
  const paginatedItems = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return skItems.slice(start, start + itemsPerPage)
  }, [skItems, currentPage])

  useEffect(() => {
    if (currentPage > totalPages) setCurrentPage(totalPages)
  }, [currentPage, totalPages])

  const openDialog = (item?: SKItem) => {
    setFormErrors({})
    if (item) {
      setIsEditing(true)
      setFormData({
        id: item.id,
        judul: item.judul || '',
        konten: item.konten || '',
        tipe: item.tipe || 'Ketentuan',
        versi: item.versi || '',
      })
    } else {
      setIsEditing(false)
      setFormData({
        judul: '',
        konten: '',
        tipe: 'Ketentuan',
        versi: '',
      })
    }
    setDialogOpen(true)
  }

  const closeDialog = () => {
    setDialogOpen(false)
    setFormErrors({})
  }

  const validateForm = (): boolean => {
    const errors: Partial<Record<keyof FormState, string>> = {}
    if (!formData.judul.trim()) errors.judul = 'Judul harus diisi'
    if (!formData.konten.trim()) errors.konten = 'Konten harus diisi'
    if (!formData.tipe) errors.tipe = 'Tipe harus dipilih'
    if (formData.tipe === 'Pembaruan' && !formData.versi.trim()) {
      errors.versi = 'Versi harus diisi untuk tipe Pembaruan'
    }
    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const saveItem = async () => {
    if (!validateForm()) {
      showFeedback('Form tidak valid. Harap periksa inputan Anda.', 'error')
      return
    }
    setSaving(true)
    try {
      const payload = {
        judul: formData.judul.trim(),
        konten: formData.konten,
        tipe: formData.tipe,
        versi: formData.tipe === 'Pembaruan' ? formData.versi.trim() : null,
      }
      if (isEditing && formData.id) {
        await apiClient.patch(`/sk/${formData.id}`, payload)
        showFeedback('Item S&K berhasil diperbarui', 'success')
      } else {
        await apiClient.post('/sk', payload)
        showFeedback('Item S&K baru berhasil ditambahkan', 'success')
      }
      await fetchSK()
      closeDialog()
    } catch (error: any) {
      console.error('Gagal menyimpan item S&K:', error)
      const apiErrorMsg =
        error?.response?.data?.detail || error?.response?.data?.message || 'Gagal menyimpan item S&K'
      showFeedback(apiErrorMsg, 'error')
    } finally {
      setSaving(false)
    }
  }

  const openDeleteDialog = (item: SKItem) => {
    setItemToDelete(item)
    setDeleteDialogOpen(true)
  }

  const closeDeleteDialog = () => {
    setDeleteDialogOpen(false)
    setItemToDelete(null)
  }

  const deleteItemConfirm = async () => {
    if (!itemToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/sk/${itemToDelete.id}`)
      showFeedback('Item S&K berhasil dihapus', 'success')
      await fetchSK()
      closeDeleteDialog()
    } catch (error: any) {
      console.error('Gagal menghapus item S&K:', error)
      const apiErrorMsg =
        error?.response?.data?.detail || error?.response?.data?.message || 'Gagal menghapus item S&K'
      showFeedback(apiErrorMsg, 'error')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="h-11 w-11 rounded-xl bg-primary text-primary-foreground flex items-center justify-center shrink-0 shadow-xs">
            <FileText className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Manajemen S&K dan Catatan Rilis</h1>
            <p className="text-xs text-muted-foreground mt-0.5">
              Kelola syarat &amp; ketentuan dan rilis update sistem
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchSK} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button size="sm" onClick={() => openDialog()}>
            <Plus className="h-4 w-4 mr-1.5" />
            Tambah Konten
          </Button>
        </div>
      </div>

      {/* Feedback Banner */}
      {feedback && (
        <div
          className={`flex items-center justify-between gap-3 px-4 py-2.5 rounded-xl border text-sm ${
            feedback.type === 'success'
              ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-300'
              : 'bg-red-50 dark:bg-red-950/40 border-red-200 dark:border-red-800 text-red-800 dark:text-red-300'
          }`}
        >
          <div className="flex items-center gap-2">
            {feedback.type === 'success' ? (
              <CheckCircle2 className="h-4 w-4 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 shrink-0" />
            )}
            <span className="font-medium">{feedback.text}</span>
          </div>
          <button type="button" onClick={() => setFeedback(null)} aria-label="Tutup notifikasi">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Main Content Card */}
      <Card className="shadow-xs py-0 overflow-hidden">
        <CardHeader className="px-5 py-4 flex flex-row items-center justify-between space-y-0">
          <div className="flex items-center gap-2.5">
            <List className="h-4 w-4 text-purple-600 dark:text-purple-400" />
            <CardTitle className="text-base font-bold">Daftar Konten</CardTitle>
          </div>
          <Badge
            variant="secondary"
            className="bg-purple-50 dark:bg-purple-950/50 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800 text-xs"
          >
            {skItems.length} item
          </Badge>
        </CardHeader>

        {loading && skItems.length > 0 && (
          <div className="h-0.5 w-full bg-primary/20 overflow-hidden">
            <div className="h-full w-1/3 bg-primary animate-pulse" />
          </div>
        )}

        <Separator />

        <CardContent className="p-0">
          {loading && skItems.length === 0 ? (
            <div className="p-5 space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </div>
          ) : skItems.length === 0 ? (
            <div className="py-16 text-center">
              <FileText className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
              <div className="text-sm font-medium text-foreground">Belum ada konten tersedia</div>
              <p className="text-xs text-muted-foreground mt-1 mb-4">
                Tambahkan syarat &amp; ketentuan atau catatan rilis pertama Anda.
              </p>
              <Button size="sm" onClick={() => openDialog()}>
                <Plus className="h-4 w-4 mr-1.5" />
                Tambah Konten
              </Button>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead className="px-5 py-3 text-xs font-semibold w-[35%]">Judul</TableHead>
                      <TableHead className="px-4 py-3 text-xs font-semibold w-[15%]">Tipe</TableHead>
                      <TableHead className="px-4 py-3 text-xs font-semibold w-[12%]">Versi</TableHead>
                      <TableHead className="px-4 py-3 text-xs font-semibold w-[23%]">Dibuat Pada</TableHead>
                      <TableHead className="px-5 py-3 text-xs font-semibold text-center w-[15%]">
                        Actions
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedItems.map((item) => {
                      const isPembaruan = item.tipe === 'Pembaruan'
                      return (
                        <TableRow key={item.id} className="hover:bg-muted/40">
                          <TableCell className="px-5 py-3.5">
                            <div className="font-medium text-sm text-foreground">{item.judul}</div>
                          </TableCell>
                          <TableCell className="px-4 py-3.5">
                            <span
                              className={`inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-md border ${
                                isPembaruan
                                  ? 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
                                  : 'bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800'
                              }`}
                            >
                              {isPembaruan ? (
                                <History className="h-3.5 w-3.5" />
                              ) : (
                                <FileText className="h-3.5 w-3.5" />
                              )}
                              {item.tipe}
                            </span>
                          </TableCell>
                          <TableCell className="px-4 py-3.5">
                            {item.versi ? (
                              <Badge
                                variant="outline"
                                className="font-mono text-xs text-sky-700 dark:text-sky-300 border-sky-300 dark:border-sky-800"
                              >
                                {item.versi}
                              </Badge>
                            ) : (
                              <span className="text-xs text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell className="px-4 py-3.5">
                            <div className="text-sm font-medium text-foreground">
                              {formatDate(item.created_at)}
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {formatTime(item.created_at)}
                            </div>
                          </TableCell>
                          <TableCell className="px-5 py-3.5">
                            <div className="flex items-center justify-center gap-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-950/50"
                                onClick={() => openDialog(item)}
                                title="Edit"
                              >
                                <Pencil className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50"
                                onClick={() => openDeleteDialog(item)}
                                title="Hapus"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-between px-5 py-3 border-t text-xs text-muted-foreground">
                  <span>
                    Halaman {currentPage} dari {totalPages} ({skItems.length} item)
                  </span>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 text-xs"
                      disabled={currentPage <= 1}
                      onClick={() => setCurrentPage((p) => p - 1)}
                    >
                      Sebelumnya
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 text-xs"
                      disabled={currentPage >= totalPages}
                      onClick={() => setCurrentPage((p) => p + 1)}
                    >
                      Berikutnya
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Add/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={(open) => !open && closeDialog()}>
        <DialogContent className="sm:max-w-[680px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {isEditing ? (
                <Pencil className="h-4 w-4 text-primary" />
              ) : (
                <Plus className="h-4 w-4 text-primary" />
              )}
              {isEditing ? 'Edit Konten' : 'Tambah Konten Baru'}
            </DialogTitle>
            <DialogDescription>
              Lengkapi informasi syarat &amp; ketentuan atau catatan rilis sistem.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label htmlFor="sk-judul">Judul Konten</Label>
              <Input
                id="sk-judul"
                value={formData.judul}
                onChange={(e) => {
                  setFormData((p) => ({ ...p, judul: e.target.value }))
                  if (formErrors.judul) setFormErrors((p) => ({ ...p, judul: undefined }))
                }}
                placeholder="Masukkan judul konten..."
              />
              {formErrors.judul && (
                <p className="text-xs text-destructive">{formErrors.judul}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="sk-konten">Konten</Label>
              <Textarea
                id="sk-konten"
                rows={6}
                value={formData.konten}
                onChange={(e) => {
                  setFormData((p) => ({ ...p, konten: e.target.value }))
                  if (formErrors.konten) setFormErrors((p) => ({ ...p, konten: undefined }))
                }}
                placeholder="Masukkan konten (mendukung HTML dasar)..."
              />
              {formErrors.konten && (
                <p className="text-xs text-destructive">{formErrors.konten}</p>
              )}
              <div className="flex items-center gap-2 rounded-lg border border-sky-200 dark:border-sky-800/60 bg-sky-50/70 dark:bg-sky-950/30 px-3 py-2 text-xs text-sky-800 dark:text-sky-300">
                <Info className="h-4 w-4 shrink-0 text-sky-600 dark:text-sky-400" />
                <span>
                  Anda dapat menggunakan tag HTML dasar seperti &lt;strong&gt;, &lt;em&gt;, &lt;p&gt;, &lt;br&gt;
                </span>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>Tipe Konten</Label>
              <Select
                value={formData.tipe}
                onValueChange={(val) => {
                  setFormData((p) => ({ ...p, tipe: val }))
                  if (formErrors.tipe) setFormErrors((p) => ({ ...p, tipe: undefined }))
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Pilih tipe konten" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Ketentuan">Ketentuan</SelectItem>
                  <SelectItem value="Pembaruan">Pembaruan</SelectItem>
                </SelectContent>
              </Select>
              {formErrors.tipe && (
                <p className="text-xs text-destructive">{formErrors.tipe}</p>
              )}
            </div>

            {formData.tipe === 'Pembaruan' && (
              <div className="space-y-1.5">
                <Label htmlFor="sk-versi" className="flex items-center gap-1.5">
                  <Tag className="h-3.5 w-3.5 text-sky-600 dark:text-sky-400" />
                  Versi
                </Label>
                <Input
                  id="sk-versi"
                  value={formData.versi}
                  onChange={(e) => {
                    setFormData((p) => ({ ...p, versi: e.target.value }))
                    if (formErrors.versi) setFormErrors((p) => ({ ...p, versi: undefined }))
                  }}
                  placeholder="contoh: v2.1.0"
                />
                {formErrors.versi && (
                  <p className="text-xs text-destructive">{formErrors.versi}</p>
                )}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={closeDialog} disabled={saving}>
              Batal
            </Button>
            <Button onClick={saveItem} disabled={saving}>
              {saving ? (
                <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
              ) : (
                <Save className="h-4 w-4 mr-1.5" />
              )}
              Simpan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={(open) => !open && closeDeleteDialog()}>
        <DialogContent className="sm:max-w-[460px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              Konfirmasi Hapus
            </DialogTitle>
            <DialogDescription>Tindakan ini tidak dapat dibatalkan.</DialogDescription>
          </DialogHeader>

          <div className="space-y-3 py-2 text-sm">
            <p className="text-foreground">Anda yakin ingin menghapus konten ini?</p>
            {itemToDelete && (
              <div className="rounded-lg border border-red-200 dark:border-red-900/60 bg-red-50 dark:bg-red-950/40 p-3 space-y-1">
                <div className="font-semibold text-foreground">{itemToDelete.judul}</div>
                <div className="text-xs text-muted-foreground">{itemToDelete.tipe}</div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={closeDeleteDialog} disabled={deleting}>
              Batal
            </Button>
            <Button variant="destructive" onClick={deleteItemConfirm} disabled={deleting}>
              {deleting ? (
                <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
              ) : (
                <Trash2 className="h-4 w-4 mr-1.5" />
              )}
              Hapus
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
