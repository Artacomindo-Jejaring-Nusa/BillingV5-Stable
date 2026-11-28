import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
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
  Shield,
  ShieldCheck,
  KeyRound,
  Plus,
  RefreshCw,
  Pencil,
  Trash2,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
} from 'lucide-react'

interface PermissionItem {
  id: number
  name: string
}

interface RoleItem {
  id: number
  name: string
  permissions: PermissionItem[]
}

function formatPermissionLabel(name: string) {
  const n = name.toLowerCase()
  if (n.startsWith('view_widget_')) {
    return 'Widget: ' + n.replace('view_widget_', '').replace(/_/g, ' ')
  }
  return n
    .replace(/^view_/, 'Lihat ')
    .replace(/^create_/, 'Tambah ')
    .replace(/^edit_/, 'Edit ')
    .replace(/^delete_/, 'Hapus ')
    .replace(/^manage_/, 'Kelola ')
    .replace(/_/g, ' ')
}

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

export function RolesView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_roles')
  const canEdit = hasPermission('edit_roles')
  const canDelete = hasPermission('delete_roles')
  const verifyToken = useAuthStore((s) => s.verifyToken)
  const currentUserRoleName = useAuthStore((s) => s.user?.role?.name)

  const [roles, setRoles] = useState<RoleItem[]>([])
  const [allPermissions, setAllPermissions] = useState<PermissionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  // Dialogs
  const [showFormDialog, setShowFormDialog] = useState(false)
  const [editedId, setEditedId] = useState<number | null>(null)
  const [roleName, setRoleName] = useState('')
  const [selectedPermissions, setSelectedPermissions] = useState<number[]>([])
  const [permSearch, setPermSearch] = useState('')

  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<RoleItem | null>(null)

  // Feedback
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4000)
  }

  const fetchAll = useCallback(async () => {
    setLoading(true)
    try {
      const [rolesRes, permsRes] = await Promise.all([
        apiClient.get('/roles'),
        apiClient.get('/permissions'),
      ])
      setRoles(Array.isArray(rolesRes.data) ? rolesRes.data : rolesRes.data?.data || [])
      setAllPermissions(Array.isArray(permsRes.data) ? permsRes.data : permsRes.data?.data || [])
    } catch {
      notify('Gagal memuat data roles & permissions', 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  const groupedPermissions = useMemo(() => {
    const q = permSearch.toLowerCase().trim()
    const filtered = q
      ? allPermissions.filter(
          (p) =>
            p.name.toLowerCase().includes(q) ||
            formatPermissionLabel(p.name).toLowerCase().includes(q) ||
            getPermissionCategory(p.name).toLowerCase().includes(q)
        )
      : allPermissions

    const order = [
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

    const rawGroups = filtered.reduce((groups, perm) => {
      const cat = getPermissionCategory(perm.name)
      if (!groups[cat]) groups[cat] = []
      groups[cat].push(perm)
      return groups
    }, {} as Record<string, PermissionItem[]>)

    const sorted: Record<string, PermissionItem[]> = {}
    order.forEach((cat) => {
      if (rawGroups[cat]?.length) {
        sorted[cat] = rawGroups[cat].sort((a, b) => a.name.localeCompare(b.name))
      }
    })
    return sorted
  }, [allPermissions, permSearch])

  const openDialog = (role?: RoleItem) => {
    setPermSearch('')
    if (role) {
      setEditedId(role.id)
      setRoleName(role.name)
      setSelectedPermissions((role.permissions || []).map((p) => p.id))
    } else {
      setEditedId(null)
      setRoleName('')
      setSelectedPermissions([])
    }
    setShowFormDialog(true)
  }

  const togglePermission = (id: number) => {
    setSelectedPermissions((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )
  }

  const toggleGroupPermissions = (perms: PermissionItem[]) => {
    const ids = perms.map((p) => p.id)
    const allSelected = ids.every((id) => selectedPermissions.includes(id))
    if (allSelected) {
      setSelectedPermissions((prev) => prev.filter((id) => !ids.includes(id)))
    } else {
      setSelectedPermissions((prev) => Array.from(new Set([...prev, ...ids])))
    }
  }

  const selectAllPermissions = () => {
    setSelectedPermissions(allPermissions.map((p) => p.id))
  }

  const clearAllPermissions = () => {
    setSelectedPermissions([])
  }

  const saveRole = async () => {
    if (!roleName.trim()) {
      notify('Nama role tidak boleh kosong', 'error')
      return
    }
    setSaving(true)
    const payload = {
      name: roleName.trim(),
      permission_ids: selectedPermissions,
    }
    try {
      if (editedId !== null) {
        await apiClient.patch(`/roles/${editedId}`, payload)
        notify('Role berhasil diperbarui', 'success')
      } else {
        await apiClient.post('/roles', payload)
        notify('Role baru berhasil dibuat', 'success')
      }
      setShowFormDialog(false)
      await fetchAll()

      // Refresh current user's RBAC permissions if their role was updated
      if (currentUserRoleName && currentUserRoleName.toLowerCase() === roleName.trim().toLowerCase()) {
        await verifyToken()
      }
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string; detail?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menyimpan role', 'error')
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!itemToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/roles/${itemToDelete.id}`)
      setShowDeleteDialog(false)
      setItemToDelete(null)
      notify('Role berhasil dihapus', 'success')
      fetchAll()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string; detail?: string } } }
      notify(e.response?.data?.error || e.response?.data?.detail || 'Gagal menghapus role', 'error')
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-primary text-primary-foreground">
            <Shield className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Role Management</h1>
            <p className="text-xs text-muted-foreground">Kelola roles dan hak akses (RBAC Permissions) sistem</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchAll} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah Role
            </Button>
          )}
        </div>
      </div>

      {/* Compact Stats Strip */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-blue-50 text-blue-600"><ShieldCheck className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Total Roles</div>
            <div className="text-lg font-bold text-foreground">{roles.length}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-emerald-50 text-emerald-600"><KeyRound className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Total Permissions</div>
            <div className="text-lg font-bold text-emerald-600">{allPermissions.length}</div>
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

      {/* Roles Table */}
      <Card className="shadow-xs">
        <CardHeader className="py-3 px-4 flex flex-row items-center justify-between">
          <CardTitle className="text-sm font-semibold">Daftar Roles</CardTitle>
          <Badge variant="secondary" className="text-xs">{roles.length} roles</Badge>
        </CardHeader>
        <Separator />
        {loading && roles.length === 0 ? (
          <div className="py-16 text-center text-xs text-muted-foreground">
            <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
            Memuat data roles...
          </div>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/50">
                  <TableHead className="text-xs px-4 py-3 w-[60px] text-center">No.</TableHead>
                  <TableHead className="text-xs px-4 py-3">Role Information</TableHead>
                  <TableHead className="text-xs px-4 py-3 text-center">Permissions</TableHead>
                  <TableHead className="text-xs px-4 py-3 text-right w-[180px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {roles.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center py-10 text-xs text-muted-foreground">
                      Tidak ada data roles
                    </TableCell>
                  </TableRow>
                ) : (
                  roles.map((role, idx) => {
                    const count = role.permissions?.length || 0
                    const permCls =
                      count > 5
                        ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                        : count > 2
                          ? 'bg-amber-50 text-amber-700 border-amber-200'
                          : 'bg-red-50 text-red-700 border-red-200'
                    return (
                      <TableRow key={role.id} className="hover:bg-muted/40">
                        <TableCell className="px-4 py-2.5 text-center font-mono text-xs text-muted-foreground">
                          {idx + 1}
                        </TableCell>
                        <TableCell className="px-4 py-2.5">
                          <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-lg bg-primary text-primary-foreground flex items-center justify-center shrink-0">
                              <Shield className="h-4 w-4" />
                            </div>
                            <div>
                              <div className="text-sm font-bold text-foreground">{role.name}</div>
                              <div className="text-[11px] text-muted-foreground font-mono">Role ID: {role.id}</div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="px-4 py-2.5 text-center">
                          <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${permCls}`}>
                            <KeyRound className="h-3 w-3" />
                            {count} Permissions
                          </span>
                        </TableCell>
                        <TableCell className="px-4 py-2.5 text-right">
                          <div className="flex items-center justify-end gap-1.5">
                            {canEdit && (
                              <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => openDialog(role)}>
                                <Pencil className="h-3 w-3 mr-1 text-blue-600" /> Edit
                              </Button>
                            )}
                            {canDelete && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-7 text-xs text-red-600 hover:bg-red-50"
                                onClick={() => {
                                  setItemToDelete(role)
                                  setShowDeleteDialog(true)
                                }}
                              >
                                <Trash2 className="h-3 w-3 mr-1" /> Hapus
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </Card>

      {/* Add / Edit Role Dialog */}
      <Dialog open={showFormDialog} onOpenChange={setShowFormDialog}>
        <DialogContent className="w-[95vw] sm:max-w-4xl max-h-[90vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <ShieldCheck className="h-5 w-5 text-blue-600" />
              {editedId !== null ? 'Edit Role & Permissions' : 'Tambah Role Baru'}
            </DialogTitle>
            <DialogDescription>
              {editedId !== null ? 'Ubah nama role dan atur hak akses RBAC' : 'Buat role baru dan pilih hak aksesnya'}
            </DialogDescription>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto space-y-4 py-2 pr-1">
            <div>
              <Label className="text-sm mb-1.5 block">Nama Role *</Label>
              <Input
                value={roleName}
                onChange={(e) => setRoleName(e.target.value)}
                placeholder="Contoh: NOC, Teknisi, CS, Finance"
              />
            </div>

            <Separator />

            <div className="space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <KeyRound className="h-4 w-4 text-blue-600" />
                  <span className="text-sm font-bold text-foreground">Hak Akses (Permissions)</span>
                  <Badge variant="secondary" className="text-xs">{selectedPermissions.length} / {allPermissions.length} dipilih</Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Input
                    value={permSearch}
                    onChange={(e) => setPermSearch(e.target.value)}
                    placeholder="Cari permission..."
                    className="h-8 w-44 text-xs"
                  />
                  <Button type="button" variant="outline" size="sm" className="h-8 text-xs" onClick={selectAllPermissions}>
                    Pilih Semua
                  </Button>
                  <Button type="button" variant="ghost" size="sm" className="h-8 text-xs text-red-600" onClick={clearAllPermissions}>
                    Clear All
                  </Button>
                </div>
              </div>

              {/* Grouped Permissions Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {Object.entries(groupedPermissions).map(([groupName, perms]) => {
                  const selectedInGroup = perms.filter((p) => selectedPermissions.includes(p.id)).length
                  const allGroupSelected = selectedInGroup === perms.length && perms.length > 0
                  return (
                    <Card
                      key={groupName}
                      className={`py-0 shadow-none border ${
                        allGroupSelected
                          ? 'border-emerald-300 bg-emerald-50/30'
                          : selectedInGroup > 0
                            ? 'border-amber-300 bg-amber-50/20'
                            : 'border-border'
                      }`}
                    >
                      <CardContent className="p-3 space-y-2">
                        <div className="flex items-center justify-between border-b pb-2">
                          <label className="flex items-center gap-2 cursor-pointer font-semibold text-xs capitalize text-foreground">
                            <input
                              type="checkbox"
                              checked={allGroupSelected}
                              onChange={() => toggleGroupPermissions(perms)}
                              className="rounded border-border text-blue-600"
                            />
                            {groupName.replace(/_/g, ' ')}
                          </label>
                          <Badge variant="outline" className="text-[10px] font-mono">
                            {selectedInGroup}/{perms.length}
                          </Badge>
                        </div>
                        <div className="max-h-48 overflow-y-auto space-y-1.5 pr-1">
                          {perms.map((perm) => (
                            <label key={perm.id} className="flex items-start gap-2 text-xs text-foreground/90 cursor-pointer hover:text-foreground">
                              <input
                                type="checkbox"
                                checked={selectedPermissions.includes(perm.id)}
                                onChange={() => togglePermission(perm.id)}
                                className="rounded border-border text-blue-600 mt-0.5 shrink-0"
                              />
                              <span className="leading-snug capitalize" title={perm.name}>{formatPermissionLabel(perm.name)}</span>
                            </label>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                  )
                })}
              </div>
            </div>
          </div>

          <DialogFooter className="border-t pt-3">
            <Button variant="outline" onClick={() => setShowFormDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveRole} disabled={saving || !roleName.trim()}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              {editedId !== null ? 'Simpan Perubahan' : 'Buat Role'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">Konfirmasi Hapus Role</DialogTitle>
            <DialogDescription>
              Anda akan menghapus role <strong>{itemToDelete?.name}</strong>. Tindakan ini tidak dapat dibatalkan dan akan menghapus semua permission terkait.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>Batal</Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Ya, Hapus Role
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
