import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  Server,
  Plus,
  RefreshCw,
  Search,
  FilterX,
  Wifi,
  WifiOff,
  HelpCircle,
  Pencil,
  Trash2,
  Eye,
  EyeOff,
  Loader2,
  CheckCircle2,
  AlertCircle,
  Network,
  User,
  Cpu,
  PlugZap,
  X,
} from 'lucide-react'

interface MikrotikServer {
  id: number | null
  name: string
  host_ip: string
  port: number
  username: string
  password?: string
  is_active: boolean
  ros_version?: string | null
  last_connection_status?: string | null
}

const defaultServer: MikrotikServer = {
  id: null,
  name: '',
  host_ip: '',
  port: 8728,
  username: '',
  password: '',
  is_active: true,
}

function getConnectionBadge(status?: string | null) {
  const s = (status || '').toLowerCase()
  if (s === 'success') {
    return {
      label: 'Success',
      cls: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      icon: <Wifi className="h-3 w-3" />,
    }
  }
  if (s === 'failure' || s === 'failed') {
    return {
      label: 'Failed',
      cls: 'bg-red-50 text-red-700 border-red-200',
      icon: <WifiOff className="h-3 w-3" />,
    }
  }
  return {
    label: status || 'Unknown',
    cls: 'bg-muted text-muted-foreground border-border',
    icon: <HelpCircle className="h-3 w-3" />,
  }
}

export function MikrotikView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_mikrotik_servers')
  const canEdit = hasPermission('edit_mikrotik_servers')
  const canDelete = hasPermission('delete_mikrotik_servers')
  const [servers, setServers] = useState<MikrotikServer[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [testingId, setTestingId] = useState<number | null>(null)

  // Filters
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedIsActive, setSelectedIsActive] = useState<string>('')
  const [selectedConnStatus, setSelectedConnStatus] = useState<string>('')

  // Form Dialog
  const [showFormDialog, setShowFormDialog] = useState(false)
  const [isEdit, setIsEdit] = useState(false)
  const [formData, setFormData] = useState<MikrotikServer>({ ...defaultServer })
  const [showPassword, setShowPassword] = useState(false)

  // Delete Dialog
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<MikrotikServer | null>(null)

  // Feedback banner
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4500)
  }

  const fetchServers = useCallback(
    async (searchOverride?: string) => {
      setLoading(true)
      try {
        const params = new URLSearchParams()
        const q = searchOverride !== undefined ? searchOverride : searchQuery
        if (q.trim()) params.append('search', q.trim())
        if (selectedIsActive !== '') params.append('is_active', selectedIsActive)
        if (selectedConnStatus) params.append('last_connection_status', selectedConnStatus)

        const res = await apiClient.get(`/mikrotik_servers?${params.toString()}`)
        const list = Array.isArray(res.data) ? res.data : res.data.data || []
        setServers(list)
      } catch (err) {
        console.error('Failed to fetch mikrotik servers:', err)
      } finally {
        setLoading(false)
      }
    },
    [searchQuery, selectedIsActive, selectedConnStatus]
  )

  useEffect(() => {
    fetchServers()
  }, [selectedIsActive, selectedConnStatus]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSearchChange = (val: string) => {
    setSearchQuery(val)
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      fetchServers(val)
    }, 500)
  }

  const resetFilters = () => {
    setSearchQuery('')
    setSelectedIsActive('')
    setSelectedConnStatus('')
    fetchServers('')
  }

  const hasFilters = Boolean(searchQuery || selectedIsActive !== '' || selectedConnStatus)

  const stats = useMemo(() => {
    const total = servers.length
    const active = servers.filter((s) => s.is_active).length
    const connected = servers.filter((s) => (s.last_connection_status || '').toLowerCase() === 'success').length
    const failed = servers.filter((s) => {
      const st = (s.last_connection_status || '').toLowerCase()
      return st === 'failure' || st === 'failed'
    }).length
    return { total, active, connected, failed }
  }, [servers])

  const openDialog = (server?: MikrotikServer) => {
    if (server) {
      setIsEdit(true)
      setFormData({ ...server, password: '' })
    } else {
      setIsEdit(false)
      setFormData({ ...defaultServer })
    }
    setShowPassword(false)
    setShowFormDialog(true)
  }

  const closeDialog = () => {
    setShowFormDialog(false)
    setFormData({ ...defaultServer })
    setIsEdit(false)
    setShowPassword(false)
  }

  const saveServer = async () => {
    if (!formData.name.trim() || !formData.host_ip.trim() || !formData.username.trim()) return
    setSaving(true)
    const payload: Partial<MikrotikServer> = { ...formData }
    if (isEdit && !payload.password) {
      delete payload.password
    }

    try {
      if (isEdit && payload.id) {
        await apiClient.patch(`/mikrotik_servers/${payload.id}`, payload)
        notify('Server berhasil diperbarui!', 'success')
      } else {
        await apiClient.post('/mikrotik_servers', payload)
        notify('Server baru berhasil ditambahkan!', 'success')
      }
      closeDialog()
      fetchServers()
    } catch (err) {
      console.error(err)
      notify('Gagal menyimpan server. Silakan coba lagi.', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTestConnection = async (server: MikrotikServer) => {
    if (!server.id) return
    setTestingId(server.id)
    try {
      const res = await apiClient.post(`/mikrotik_servers/${server.id}/test_connection`)
      const testResult = res.data.test_result
      const updatedServer = res.data.updated_server
      const rosVersion = testResult?.data?.ros_version || 'N/A'
      notify(`${testResult?.message || 'Koneksi berhasil'} (ROS: ${rosVersion})`, 'success')
      if (updatedServer) {
        setServers((prev) => prev.map((s) => (s.id === server.id ? updatedServer : s)))
      }
    } catch (err: unknown) {
      const error = err as { response?: { data?: { test_result?: { message?: string }; updated_server?: MikrotikServer } } }
      if (error.response?.data) {
        const { test_result, updated_server } = error.response.data
        notify(test_result?.message || 'Koneksi gagal.', 'error')
        if (updated_server) {
          setServers((prev) => prev.map((s) => (s.id === server.id ? updated_server : s)))
        }
      } else {
        notify('Terjadi kesalahan jaringan saat menguji koneksi.', 'error')
      }
    } finally {
      setTestingId(null)
    }
  }

  const openDeleteDialog = (server: MikrotikServer) => {
    setItemToDelete(server)
    setShowDeleteDialog(true)
  }

  const confirmDelete = async () => {
    if (!itemToDelete?.id) return
    setDeleting(true)
    try {
      await apiClient.delete(`/mikrotik_servers/${itemToDelete.id}`)
      setShowDeleteDialog(false)
      setItemToDelete(null)
      notify('Server berhasil dihapus!', 'success')
      fetchServers()
    } catch (err) {
      console.error(err)
      notify('Gagal menghapus server. Silakan coba lagi.', 'error')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-primary text-primary-foreground">
              <Server className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-foreground">Mikrotik Servers</h1>
              <p className="text-xs text-muted-foreground">Kelola koneksi API dan status router Mikrotik Anda</p>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="hidden md:flex items-center gap-1.5 mr-2 text-xs">
            <Badge variant="outline" className="bg-card font-medium">{stats.total} Server</Badge>
            <Badge className="bg-emerald-50 text-emerald-700 hover:bg-emerald-50 border border-emerald-200">{stats.connected} Connected</Badge>
            {stats.failed > 0 && (
              <Badge className="bg-red-50 text-red-700 hover:bg-red-50 border border-red-200">{stats.failed} Failed</Badge>
            )}
          </div>
          <Button variant="outline" size="sm" onClick={() => fetchServers()} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah Server
            </Button>
          )}
        </div>
      </div>

      {/* Feedback Banner */}
      {feedback && (
        <div
          className={`flex items-center justify-between gap-3 px-4 py-3 rounded-xl border text-sm ${
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
          <button type="button" onClick={() => setFeedback(null)} className="opacity-70 hover:opacity-100">
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Filter Bar */}
      <Card className="shadow-xs py-0">
        <CardContent className="p-3.5">
          <div className="flex flex-col md:flex-row md:items-center gap-2.5">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder="Cari server berdasarkan nama atau IP..."
                className="pl-9 h-9"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:flex gap-2.5">
              <Select
                value={selectedIsActive === '' ? '__all__' : selectedIsActive}
                onValueChange={(v) => setSelectedIsActive(v === '__all__' ? '' : v)}
              >
                <SelectTrigger className="h-9 md:w-[170px] w-full">
                  <SelectValue placeholder="Status Server" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Status Server</SelectItem>
                  <SelectItem value="true">Aktif</SelectItem>
                  <SelectItem value="false">Nonaktif</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={selectedConnStatus || '__all__'}
                onValueChange={(v) => setSelectedConnStatus(v === '__all__' ? '' : v)}
              >
                <SelectTrigger className="h-9 md:w-[170px] w-full">
                  <SelectValue placeholder="Status Koneksi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Koneksi</SelectItem>
                  <SelectItem value="Success">Success</SelectItem>
                  <SelectItem value="Failed">Failed</SelectItem>
                </SelectContent>
              </Select>

              {hasFilters && (
                <Button variant="ghost" size="sm" onClick={resetFilters} className="h-9 text-xs text-red-600 hover:text-red-700">
                  <FilterX className="h-3.5 w-3.5 mr-1" />
                  Reset
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Server Grid */}
      {loading && servers.length === 0 ? (
        <div className="flex items-center justify-center py-20 text-sm text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin text-blue-600 mr-2" />
          Memuat daftar server Mikrotik...
        </div>
      ) : servers.length === 0 ? (
        <Card className="shadow-xs">
          <CardContent className="py-16 text-center">
            <Server className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
            <h3 className="text-base font-semibold text-foreground/90">Belum ada server ditemukan</h3>
            <p className="text-xs text-muted-foreground mt-1 mb-4">
              {hasFilters ? 'Coba ubah filter pencarian Anda' : 'Tambahkan server Mikrotik pertama untuk mulai mengelola koneksi'}
            </p>
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah Server Baru
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {servers.map((server) => {
            const conn = getConnectionBadge(server.last_connection_status)
            const isTesting = testingId === server.id
            return (
              <Card
                key={server.id}
                className="relative overflow-hidden shadow-xs hover:shadow-md transition-shadow border py-0 flex flex-col"
              >
                <div className={`h-1 w-full ${server.is_active ? 'bg-emerald-500' : 'bg-muted-foreground/30'}`} />
                <CardContent className="p-4 flex-1 flex flex-col gap-3.5">
                  {/* Top Row */}
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3 min-w-0">
                      <div
                        className={`h-10 w-10 rounded-lg flex items-center justify-center shrink-0 border ${
                          server.is_active
                            ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
                            : 'bg-muted/50 border-border text-muted-foreground'
                        }`}
                      >
                        <Network className="h-5 w-5" />
                      </div>
                      <div className="min-w-0">
                        <h3 className="text-sm font-bold text-foreground truncate">{server.name}</h3>
                        <div className="flex items-center gap-1.5 mt-0.5">
                          <span
                            className={`h-2 w-2 rounded-full ${server.is_active ? 'bg-emerald-500' : 'bg-muted-foreground/50'}`}
                          />
                          <span className={`text-xs font-medium ${server.is_active ? 'text-emerald-700' : 'text-muted-foreground'}`}>
                            {server.is_active ? 'Active' : 'Nonaktif'}
                          </span>
                        </div>
                      </div>
                    </div>
                    <span className={`inline-flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full border shrink-0 ${conn.cls}`}>
                      {conn.icon}
                      {conn.label}
                    </span>
                  </div>

                  <Separator />

                  {/* Server Specs Grid */}
                  <div className="grid grid-cols-2 gap-3 text-xs">
                    <div className="bg-muted/50 border rounded-lg px-2.5 py-2">
                      <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">IP Address</div>
                      <div className="font-mono font-semibold text-foreground mt-0.5 truncate">{server.host_ip}</div>
                    </div>
                    <div className="bg-muted/50 border rounded-lg px-2.5 py-2">
                      <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Port API</div>
                      <div className="font-mono font-semibold text-foreground mt-0.5">{server.port}</div>
                    </div>
                    <div className="bg-muted/50 border rounded-lg px-2.5 py-2">
                      <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <User className="h-3 w-3" /> Username
                      </div>
                      <div className="font-medium text-foreground mt-0.5 truncate">{server.username}</div>
                    </div>
                    <div className="bg-muted/50 border rounded-lg px-2.5 py-2">
                      <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Cpu className="h-3 w-3" /> ROS Version
                      </div>
                      <div className="font-medium text-foreground mt-0.5 truncate">{server.ros_version || '-'}</div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 pt-1 mt-auto">
                    <Button
                      size="sm"
                      className="flex-1 h-9 bg-primary hover:bg-primary/90 text-primary-foreground"
                      disabled={isTesting}
                      onClick={() => handleTestConnection(server)}
                    >
                      {isTesting ? (
                        <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                      ) : (
                        <PlugZap className="h-3.5 w-3.5 mr-1.5" />
                      )}
                      {isTesting ? 'Testing...' : 'Test Connection'}
                    </Button>
                    {(canEdit || canDelete) && (
                      <TooltipProvider delayDuration={200}>
                        {canEdit && (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button variant="outline" size="icon" className="h-9 w-9" onClick={() => openDialog(server)}>
                                <Pencil className="h-4 w-4 text-muted-foreground" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Edit Server</TooltipContent>
                          </Tooltip>
                        )}
                        {canDelete && (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button variant="outline" size="icon" className="h-9 w-9 hover:bg-red-50 hover:border-red-200" onClick={() => openDeleteDialog(server)}>
                                <Trash2 className="h-4 w-4 text-red-500" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Hapus Server</TooltipContent>
                          </Tooltip>
                        )}
                      </TooltipProvider>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Add / Edit Dialog */}
      <Dialog open={showFormDialog} onOpenChange={setShowFormDialog}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Server className="h-5 w-5 text-blue-600" />
              {isEdit ? 'Edit Server Mikrotik' : 'Tambah Server Baru'}
            </DialogTitle>
            <DialogDescription>
              {isEdit ? 'Perbarui informasi koneksi server Mikrotik' : 'Tambahkan konfigurasi server Mikrotik baru ke sistem'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div>
              <Label className="text-xs font-semibold uppercase text-muted-foreground mb-2 block">Informasi Server</Label>
              <div className="space-y-3">
                <div>
                  <Label className="text-sm mb-1 block">Nama Server *</Label>
                  <Input
                    value={formData.name}
                    onChange={(e) => setFormData((p) => ({ ...p, name: e.target.value }))}
                    placeholder="Contoh: Mikrotik Parama"
                  />
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div className="sm:col-span-2">
                    <Label className="text-sm mb-1 block">Host / IP Address *</Label>
                    <Input
                      value={formData.host_ip}
                      onChange={(e) => setFormData((p) => ({ ...p, host_ip: e.target.value }))}
                      placeholder="Contoh: 13.1.1.6"
                      className="font-mono"
                    />
                  </div>
                  <div>
                    <Label className="text-sm mb-1 block">Port API</Label>
                    <Input
                      type="number"
                      value={formData.port}
                      onChange={(e) => setFormData((p) => ({ ...p, port: Number(e.target.value) || 8728 }))}
                      placeholder="8728"
                      className="font-mono"
                    />
                  </div>
                </div>
              </div>
            </div>

            <Separator />

            <div>
              <Label className="text-xs font-semibold uppercase text-muted-foreground mb-2 block">Kredensial Login</Label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <Label className="text-sm mb-1 block">Username *</Label>
                  <Input
                    value={formData.username}
                    onChange={(e) => setFormData((p) => ({ ...p, username: e.target.value }))}
                    placeholder="api"
                    autoComplete="username"
                  />
                </div>
                <div>
                  <Label className="text-sm mb-1 block">Password</Label>
                  <div className="relative">
                    <Input
                      type={showPassword ? 'text' : 'password'}
                      value={formData.password || ''}
                      onChange={(e) => setFormData((p) => ({ ...p, password: e.target.value }))}
                      placeholder={isEdit ? 'Kosongkan jika tidak berubah' : 'Masukkan password'}
                      autoComplete="current-password"
                      className="pr-9"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((v) => !v)}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground/90"
                    >
                      {showPassword ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between rounded-xl border bg-muted/50 px-4 py-3">
              <div>
                <div className="text-sm font-semibold text-foreground">Aktivasi Server</div>
                <div className="text-xs text-muted-foreground">
                  {formData.is_active ? 'Server aktif dan dapat digunakan' : 'Server dinonaktifkan sementara'}
                </div>
              </div>
              <Switch
                checked={formData.is_active}
                onCheckedChange={(checked) => setFormData((p) => ({ ...p, is_active: checked }))}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={closeDialog} disabled={saving}>Batal</Button>
            <Button
              onClick={saveServer}
              disabled={saving || !formData.name.trim() || !formData.host_ip.trim() || !formData.username.trim()}
            >
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              {isEdit ? 'Update Server' : 'Tambah Server'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-600">
              <Trash2 className="h-5 w-5" />
              Konfirmasi Hapus Server
            </DialogTitle>
            <DialogDescription>
              Anda yakin ingin menghapus server <strong>{itemToDelete?.name}</strong>?
            </DialogDescription>
          </DialogHeader>
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 text-xs text-amber-800">
            Tindakan ini tidak dapat dibatalkan. Semua konfigurasi yang terkait dengan server ini akan hilang.
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>Batal</Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Ya, Hapus
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
