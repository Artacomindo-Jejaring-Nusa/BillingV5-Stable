import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertCircle,
  AlertTriangle,
  Check,
  CheckCircle2,
  Copy,
  KeyRound,
  Loader2,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
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
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface APIKeyItem {
  id: number
  name: string
  prefix: string
  role_id: number
  role_name: string
  is_active: boolean
  created_at: string
}

interface RoleOption {
  id: number
  name: string
}

function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return dateStr
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function ApiKeysView() {
  const [apiKeys, setApiKeys] = useState<APIKeyItem[]>([])
  const [roles, setRoles] = useState<RoleOption[]>([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [rolesLoading, setRolesLoading] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [statusLoadingId, setStatusLoadingId] = useState<number | null>(null)

  const [generateDialogOpen, setGenerateDialogOpen] = useState(false)
  const [resultDialogOpen, setResultDialogOpen] = useState(false)
  const [keyToDelete, setKeyToDelete] = useState<APIKeyItem | null>(null)

  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyRoleId, setNewKeyRoleId] = useState('')
  const [formErrors, setFormErrors] = useState<{ name?: string; role_id?: string }>({})
  const [generatedRawKey, setGeneratedRawKey] = useState('')
  const [copied, setCopied] = useState(false)

  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const showFeedback = useCallback((text: string, type: 'success' | 'error' = 'success') => {
    setFeedback({ type, text })
  }, [])

  useEffect(() => {
    if (!feedback) return
    const timer = setTimeout(() => setFeedback(null), 4000)
    return () => clearTimeout(timer)
  }, [feedback])

  const fetchApiKeys = useCallback(async () => {
    setLoading(true)
    try {
      const response = await apiClient.get('/api-keys')
      const list = response.data?.data ?? []
      setApiKeys(Array.isArray(list) ? list : [])
    } catch (err: any) {
      showFeedback(err?.response?.data?.error || 'Gagal memuat daftar API Key.', 'error')
    } finally {
      setLoading(false)
    }
  }, [showFeedback])

  const fetchRoles = useCallback(async () => {
    setRolesLoading(true)
    try {
      const response = await apiClient.get('/roles')
      const list = response.data?.data ?? []
      setRoles(Array.isArray(list) ? list : [])
    } catch (err: any) {
      showFeedback(err?.response?.data?.error || 'Gagal memuat daftar role.', 'error')
    } finally {
      setRolesLoading(false)
    }
  }, [showFeedback])

  useEffect(() => {
    void fetchApiKeys()
  }, [fetchApiKeys])

  const filteredKeys = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return apiKeys
    return apiKeys.filter(
      (k) =>
        k.name.toLowerCase().includes(q) ||
        k.prefix.toLowerCase().includes(q) ||
        (k.role_name || '').toLowerCase().includes(q)
    )
  }, [apiKeys, search])

  const openGenerateDialog = () => {
    setNewKeyName('')
    setNewKeyRoleId('')
    setFormErrors({})
    setGenerateDialogOpen(true)
    void fetchRoles()
  }

  const submitGenerate = async () => {
    const errors: { name?: string; role_id?: string } = {}
    if (!newKeyName.trim()) errors.name = 'Nama integrasi wajib diisi'
    if (!newKeyRoleId) errors.role_id = 'Hak akses (Role) wajib dipilih'
    setFormErrors(errors)
    if (Object.keys(errors).length > 0) return

    setGenerating(true)
    try {
      const response = await apiClient.post('/api-keys', {
        name: newKeyName.trim(),
        role_id: Number(newKeyRoleId),
      })
      const rawKey = response.data?.data?.raw_key || ''
      setGeneratedRawKey(rawKey)
      setCopied(false)
      setGenerateDialogOpen(false)
      setResultDialogOpen(true)
      await fetchApiKeys()
    } catch (err: any) {
      showFeedback(err?.response?.data?.error || 'Gagal membuat API Key.', 'error')
    } finally {
      setGenerating(false)
    }
  }

  const closeResultDialog = () => {
    setResultDialogOpen(false)
    setGeneratedRawKey('')
    setCopied(false)
    showFeedback('API Key berhasil disimpan!', 'success')
  }

  const copyToClipboard = async () => {
    if (!generatedRawKey) return
    try {
      await navigator.clipboard.writeText(generatedRawKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      showFeedback('Gagal menyalin ke clipboard.', 'error')
    }
  }

  const toggleKeyStatus = async (item: APIKeyItem) => {
    setStatusLoadingId(item.id)
    try {
      await apiClient.patch(`/api-keys/${item.id}/toggle`)
      setApiKeys((prev) =>
        prev.map((k) => (k.id === item.id ? { ...k, is_active: !k.is_active } : k))
      )
      showFeedback(`Status API Key "${item.name}" berhasil diubah.`, 'success')
    } catch (err: any) {
      showFeedback(err?.response?.data?.error || 'Gagal mengubah status API Key.', 'error')
    } finally {
      setStatusLoadingId(null)
    }
  }

  const submitDelete = async () => {
    if (!keyToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/api-keys/${keyToDelete.id}`)
      const deletedName = keyToDelete.name
      setKeyToDelete(null)
      showFeedback(`API Key "${deletedName}" berhasil dihapus/revoked.`, 'success')
      await fetchApiKeys()
    } catch (err: any) {
      showFeedback(err?.response?.data?.error || 'Gagal menghapus API Key.', 'error')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="h-11 w-11 rounded-xl bg-primary text-primary-foreground flex items-center justify-center shrink-0">
            <KeyRound className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Integrasi API Key</h1>
            <p className="text-xs text-muted-foreground mt-0.5">
              Kelola API Key untuk mengintegrasikan Billing dengan portal pelanggan atau aplikasi pihak ketiga
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchApiKeys} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button size="sm" onClick={openGenerateDialog}>
            <Plus className="h-4 w-4 mr-1.5" />
            Generate API Key
          </Button>
        </div>
      </div>

      {/* Security Info Banner */}
      <div className="flex items-start gap-3 rounded-xl border border-sky-200 dark:border-sky-900/60 bg-sky-50/70 dark:bg-sky-950/30 px-4 py-3 text-xs sm:text-sm text-sky-900 dark:text-sky-200">
        <ShieldCheck className="h-5 w-5 text-sky-600 dark:text-sky-400 shrink-0 mt-0.5" />
        <div>
          <strong className="font-semibold">Informasi Keamanan:</strong> Raw token hanya ditampilkan{' '}
          <strong className="font-semibold">satu kali</strong> saat dibuat. Sistem hanya menyimpan
          hash SHA-256 demi keamanan. Simpan API Key di tempat yang aman setelah dibuat.
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
      <Card className="py-0 overflow-hidden">
        <CardHeader className="px-5 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-2.5">
            <CardTitle className="text-base font-bold">Daftar API Key</CardTitle>
            <Badge variant="secondary" className="text-xs">
              {filteredKeys.length} kunci
            </Badge>
          </div>
          <div className="relative w-full sm:w-72">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari nama, prefix, atau role..."
              className="pl-9 h-9 text-xs"
            />
          </div>
        </CardHeader>

        {loading && apiKeys.length > 0 && (
          <div className="h-0.5 w-full bg-primary/20 overflow-hidden">
            <div className="h-full w-1/3 bg-primary animate-pulse" />
          </div>
        )}

        <Separator />

        <CardContent className="p-0">
          {loading && apiKeys.length === 0 ? (
            <div className="p-5 space-y-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full rounded-lg" />
              ))}
            </div>
          ) : filteredKeys.length === 0 ? (
            <div className="py-16 text-center">
              <KeyRound className="h-11 w-11 text-muted-foreground/40 mx-auto mb-3" />
              <div className="text-sm font-medium text-foreground">
                {search ? 'Tidak ada API Key yang cocok dengan pencarian' : 'Belum ada API Key terdaftar'}
              </div>
              <p className="text-xs text-muted-foreground mt-1 mb-4">
                {search
                  ? 'Coba ubah kata kunci pencarian Anda.'
                  : 'Generate API Key baru untuk menghubungkan layanan eksternal.'}
              </p>
              {!search && (
                <Button size="sm" onClick={openGenerateDialog}>
                  <Plus className="h-4 w-4 mr-1.5" />
                  Generate API Key
                </Button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="px-5 py-3 text-xs font-semibold">Nama Integrasi</TableHead>
                    <TableHead className="px-4 py-3 text-xs font-semibold">Prefix Key</TableHead>
                    <TableHead className="px-4 py-3 text-xs font-semibold">Hak Akses (Role)</TableHead>
                    <TableHead className="px-4 py-3 text-xs font-semibold text-center">Status</TableHead>
                    <TableHead className="px-4 py-3 text-xs font-semibold">Tanggal Dibuat</TableHead>
                    <TableHead className="px-5 py-3 text-xs font-semibold text-center">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredKeys.map((item) => (
                    <TableRow key={item.id} className="hover:bg-muted/40">
                      <TableCell className="px-5 py-3.5">
                        <div className="font-medium text-sm text-foreground">{item.name}</div>
                      </TableCell>
                      <TableCell className="px-4 py-3.5">
                        <Badge
                          variant="outline"
                          className="font-mono text-xs bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800"
                        >
                          {item.prefix}...
                        </Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5">
                        <Badge variant="secondary" className="text-xs font-medium">
                          {item.role_name}
                        </Badge>
                      </TableCell>
                      <TableCell className="px-4 py-3.5">
                        <div className="flex items-center justify-center gap-2">
                          {statusLoadingId === item.id && (
                            <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
                          )}
                          <Switch
                            checked={item.is_active}
                            disabled={statusLoadingId === item.id}
                            onCheckedChange={() => void toggleKeyStatus(item)}
                            aria-label={`Toggle status ${item.name}`}
                          />
                        </div>
                      </TableCell>
                      <TableCell className="px-4 py-3.5 text-xs text-muted-foreground">
                        {formatDate(item.created_at)}
                      </TableCell>
                      <TableCell className="px-5 py-3.5 text-center">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50"
                          onClick={() => setKeyToDelete(item)}
                          title="Hapus/Revoke API Key"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Dialog: Generate API Key */}
      <Dialog open={generateDialogOpen} onOpenChange={(open) => !generating && setGenerateDialogOpen(open)}>
        <DialogContent className="sm:max-w-[480px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-primary" />
              Generate API Key Baru
            </DialogTitle>
            <DialogDescription>
              Tentukan nama integrasi dan role hak akses untuk API Key ini.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label htmlFor="api-key-name">Nama/Deskripsi Integrasi</Label>
              <Input
                id="api-key-name"
                value={newKeyName}
                onChange={(e) => {
                  setNewKeyName(e.target.value)
                  if (formErrors.name) setFormErrors((p) => ({ ...p, name: undefined }))
                }}
                placeholder="Contoh: Flutter Customer Portal"
              />
              {formErrors.name && (
                <p className="text-xs text-destructive">{formErrors.name}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label>Hak Akses (Role)</Label>
              <Select
                value={newKeyRoleId}
                disabled={rolesLoading}
                onValueChange={(val) => {
                  setNewKeyRoleId(val)
                  if (formErrors.role_id) setFormErrors((p) => ({ ...p, role_id: undefined }))
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue
                    placeholder={rolesLoading ? 'Memuat daftar role...' : 'Pilih Role untuk API Key ini'}
                  />
                </SelectTrigger>
                <SelectContent>
                  {roles.map((role) => (
                    <SelectItem key={role.id} value={String(role.id)}>
                      {role.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {formErrors.role_id && (
                <p className="text-xs text-destructive">{formErrors.role_id}</p>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setGenerateDialogOpen(false)}
              disabled={generating}
            >
              Batal
            </Button>
            <Button onClick={submitGenerate} disabled={generating}>
              {generating && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Generate Key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Show Raw Key Once */}
      <Dialog open={resultDialogOpen} onOpenChange={() => {}}>
        <DialogContent
          className="sm:max-w-[540px]"
          onEscapeKeyDown={(e) => e.preventDefault()}
          onPointerDownOutside={(e) => e.preventDefault()}
        >
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-emerald-700 dark:text-emerald-400">
              <CheckCircle2 className="h-5 w-5" />
              API Key Berhasil Dibuat
            </DialogTitle>
            <DialogDescription>
              Salin token di bawah ini dan simpan di tempat yang aman. Kunci ini tidak akan ditampilkan lagi.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="rounded-xl border border-dashed border-emerald-300 dark:border-emerald-800 bg-emerald-50/70 dark:bg-emerald-950/30 p-4 flex items-center justify-between gap-3">
              <div className="min-w-0 flex-1 space-y-1">
                <div className="text-[11px] font-bold uppercase tracking-wider text-emerald-700 dark:text-emerald-400">
                  Raw API Key Token
                </div>
                <div className="font-mono text-sm font-bold text-foreground break-all select-all">
                  {generatedRawKey}
                </div>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="shrink-0"
                onClick={() => void copyToClipboard()}
              >
                {copied ? (
                  <>
                    <Check className="h-4 w-4 mr-1.5 text-emerald-600" />
                    Tersalin
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 mr-1.5" />
                    Salin
                  </>
                )}
              </Button>
            </div>

            <div className="rounded-lg border border-amber-200 dark:border-amber-800/60 bg-amber-50/70 dark:bg-amber-950/30 p-3 flex items-start gap-2.5 text-xs text-amber-900 dark:text-amber-200">
              <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
              <span>
                Setelah Anda menutup dialog ini, token mentah tidak dapat dilihat kembali.
              </span>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={closeResultDialog} className="w-full sm:w-auto">
              Saya Sudah Menyalinnya
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Delete / Revoke Confirmation */}
      <Dialog open={Boolean(keyToDelete)} onOpenChange={(open) => !open && setKeyToDelete(null)}>
        <DialogContent className="sm:max-w-[440px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              Hapus/Revoke API Key?
            </DialogTitle>
            <DialogDescription>
              Semua integrasi yang menggunakan kunci ini akan terputus seketika.
            </DialogDescription>
          </DialogHeader>

          {keyToDelete && (
            <div className="rounded-lg border border-red-200 dark:border-red-900/60 bg-red-50 dark:bg-red-950/40 p-3 text-sm space-y-1">
              <div className="font-semibold text-foreground">{keyToDelete.name}</div>
              <div className="text-xs font-mono text-muted-foreground">
                Prefix: {keyToDelete.prefix}... ({keyToDelete.role_name})
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setKeyToDelete(null)} disabled={deleting}>
              Batal
            </Button>
            <Button variant="destructive" onClick={submitDelete} disabled={deleting}>
              {deleting ? (
                <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
              ) : (
                <Trash2 className="h-4 w-4 mr-1.5" />
              )}
              Hapus Kunci
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
