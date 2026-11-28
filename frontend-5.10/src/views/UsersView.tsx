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
  Users,
  UserPlus,
  RefreshCw,
  Search,
  FilterX,
  ShieldCheck,
  ShieldAlert,
  Pencil,
  Eye,
  Trash2,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
  ChevronLeft,
  ChevronRight,
  Phone,
  KeyRound,
  UserCheck,
  Clock,
} from 'lucide-react'

interface PermissionItem {
  id: number
  name: string
}

interface RoleItem {
  id: number
  name: string
  permissions?: PermissionItem[]
}

interface UserItem {
  id: number
  name: string
  email: string
  phone_no?: string | null
  role_id?: number | null
}

function getInitials(name = '') {
  return name
    .split(' ')
    .map((w) => w.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

function getAvatarBg(name = '') {
  const colors = [
    'bg-blue-600',
    'bg-emerald-600',
    'bg-purple-600',
    'bg-amber-600',
    'bg-rose-600',
    'bg-cyan-600',
    'bg-indigo-600',
  ]
  return colors[name.length % colors.length]
}

function validatePassword(pw: string, isNew: boolean): string | null {
  if (isNew && !pw) return 'Password wajib diisi untuk user baru'
  if (!pw) return null
  if (pw.length < 8) return 'Password minimal 8 karakter'
  if (!/[A-Z]/.test(pw)) return 'Password harus mengandung minimal 1 huruf kapital'
  if (!/[a-z]/.test(pw)) return 'Password harus mengandung minimal 1 huruf kecil'
  if (!/[0-9]/.test(pw)) return 'Password harus mengandung minimal 1 angka'
  if (!/[!@#$%^&*(),.?":{}|<>]/.test(pw)) return 'Password harus mengandung minimal 1 karakter khusus'
  if (/\s/.test(pw)) return 'Password tidak boleh mengandung spasi'
  return null
}

export function UsersView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_users')
  const canEdit = hasPermission('edit_users')
  const canDelete = hasPermission('delete_users')
  const [users, setUsers] = useState<UserItem[]>([])
  const [roles, setRoles] = useState<RoleItem[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  // Filters & Pagination
  const [searchQuery, setSearchQuery] = useState('')
  const [filterRole, setFilterRole] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 10

  // Dialogs
  const [showFormDialog, setShowFormDialog] = useState(false)
  const [isEdit, setIsEdit] = useState(false)
  const [formData, setFormData] = useState({
    id: 0,
    name: '',
    email: '',
    password: '',
    role_id: '' as string,
    phone_no: '',
  })
  const [formError, setFormError] = useState('')

  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<UserItem | null>(null)

  const [showDetailDialog, setShowDetailDialog] = useState(false)
  const [selectedUser, setSelectedUser] = useState<UserItem | null>(null)

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
      const [usersRes, rolesRes] = await Promise.all([
        apiClient.get('/users', { params: { limit: 1000 } }),
        apiClient.get('/roles'),
      ])
      setUsers(Array.isArray(usersRes.data) ? usersRes.data : usersRes.data?.data || [])
      setRoles(Array.isArray(rolesRes.data) ? rolesRes.data : rolesRes.data?.data || [])
    } catch {
      notify('Gagal memuat data users', 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  const getRoleById = useCallback(
    (roleId?: number | string | null) => {
      if (!roleId) return null
      return roles.find((r) => String(r.id) === String(roleId)) || null
    },
    [roles]
  )

  const filteredUsers = useMemo(() => {
    let list = users
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase().trim()
      list = list.filter(
        (u) => u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q)
      )
    }
    if (filterRole) {
      if (filterRole === 'no-role') {
        list = list.filter((u) => !u.role_id)
      } else {
        list = list.filter((u) => String(u.role_id) === filterRole)
      }
    }
    return list
  }, [users, searchQuery, filterRole])

  const totalPages = Math.max(1, Math.ceil(filteredUsers.length / itemsPerPage))
  const paginatedUsers = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage
    return filteredUsers.slice(start, start + itemsPerPage)
  }, [filteredUsers, currentPage])

  const usersWithoutRole = useMemo(() => users.filter((u) => !u.role_id).length, [users])

  const openDialog = (user?: UserItem) => {
    setFormError('')
    if (user) {
      setIsEdit(true)
      setFormData({
        id: user.id,
        name: user.name || '',
        email: user.email || '',
        password: '',
        role_id: user.role_id ? String(user.role_id) : '',
        phone_no: user.phone_no || '',
      })
    } else {
      setIsEdit(false)
      setFormData({
        id: 0,
        name: '',
        email: '',
        password: '',
        role_id: '',
        phone_no: '',
      })
    }
    setShowFormDialog(true)
  }

  const saveUser = async () => {
    setFormError('')
    if (!formData.name.trim()) {
      setFormError('Nama lengkap wajib diisi')
      return
    }
    if (!formData.email.trim() || !/.+@.+\..+/.test(formData.email)) {
      setFormError('Email tidak valid')
      return
    }
    const pwErr = validatePassword(formData.password, !isEdit)
    if (pwErr) {
      setFormError(pwErr)
      return
    }

    setSaving(true)
    try {
      const roleIdVal = formData.role_id ? parseInt(formData.role_id, 10) : null
      if (isEdit) {
        const payload: Record<string, unknown> = {
          id: formData.id,
          name: formData.name.trim(),
          email: formData.email.trim(),
          role_id: roleIdVal,
          phone_no: formData.phone_no.trim() || null,
        }
        if (formData.password) {
          payload.password = formData.password
        }
        await apiClient.patch(`/users/${formData.id}`, payload)
        notify('User berhasil diperbarui', 'success')
      } else {
        const payload: Record<string, unknown> = {
          name: formData.name.trim(),
          email: formData.email.trim(),
          password: formData.password,
          role_id: roleIdVal,
        }
        if (formData.phone_no.trim()) {
          payload.phone_no = formData.phone_no.trim()
        }
        await apiClient.post('/users', payload)
        notify('User baru berhasil ditambahkan', 'success')
      }
      setShowFormDialog(false)
      fetchAll()
    } catch (err: unknown) {
      const e = err as { response?: { status?: number; data?: { error?: string; detail?: string } } }
      let msg = 'Terjadi kesalahan saat menyimpan user.'
      if (e.response?.status === 422) {
        const errData = e.response.data
        if (errData?.error?.includes('Password')) {
          msg = 'Password harus minimal 8 karakter, huruf kapital, huruf kecil, angka, karakter khusus, dan tanpa spasi.'
        } else if (errData?.error?.includes('email')) {
          msg = 'Email sudah digunakan oleh user lain.'
        } else if (errData?.detail) {
          msg = typeof errData.detail === 'string' ? errData.detail : JSON.stringify(errData.detail)
        }
      } else if (e.response?.status === 409) {
        msg = 'User dengan email tersebut sudah ada.'
      }
      setFormError(msg)
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!itemToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/users/${itemToDelete.id}`)
      setShowDeleteDialog(false)
      setItemToDelete(null)
      notify('User berhasil dihapus', 'success')
      fetchAll()
    } catch {
      notify('Gagal menghapus user', 'error')
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
            <Users className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">User Management</h1>
            <p className="text-xs text-muted-foreground">Kelola pengguna sistem dan assign roles</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchAll} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openDialog()}>
              <UserPlus className="h-4 w-4 mr-1.5" />
              Tambah User
            </Button>
          )}
        </div>
      </div>

      {/* Compact Stats Strip */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-emerald-50 text-emerald-600"><UserCheck className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Total Users</div>
            <div className="text-lg font-bold text-foreground">{users.length}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-blue-50 text-blue-600"><ShieldCheck className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Active Roles</div>
            <div className="text-lg font-bold text-blue-600">{roles.length}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-amber-50 text-amber-600"><ShieldAlert className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">No Role Assigned</div>
            <div className="text-lg font-bold text-amber-600">{usersWithoutRole}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-cyan-50 text-cyan-600"><Clock className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Recent Users</div>
            <div className="text-lg font-bold text-cyan-600">{Math.min(users.length, 5)}</div>
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

      {/* Filter Bar */}
      <Card className="shadow-xs py-0">
        <CardContent className="p-3.5">
          <div className="flex flex-col sm:flex-row gap-2.5">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1) }}
                placeholder="Cari user berdasarkan nama atau email..."
                className="pl-9 h-9"
              />
            </div>
            <Select value={filterRole || '__all__'} onValueChange={(v) => { setFilterRole(v === '__all__' ? '' : v); setCurrentPage(1) }}>
              <SelectTrigger className="h-9 sm:w-[210px] w-full">
                <SelectValue placeholder="Filter by Role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">Semua Role</SelectItem>
                <SelectItem value="no-role">No Role Assigned</SelectItem>
                {roles.map((r) => (
                  <SelectItem key={r.id} value={String(r.id)}>{r.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            {(searchQuery || filterRole) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => { setSearchQuery(''); setFilterRole(''); setCurrentPage(1) }}
                className="h-9 text-xs text-red-600"
              >
                <FilterX className="h-3.5 w-3.5 mr-1" /> Clear
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Users Table */}
      <Card className="shadow-xs">
        <CardHeader className="py-3 px-4 flex flex-row items-center justify-between">
          <CardTitle className="text-sm font-semibold">Daftar Users</CardTitle>
          <Badge variant="secondary" className="text-xs">{filteredUsers.length} users</Badge>
        </CardHeader>
        <Separator />
        {loading && users.length === 0 ? (
          <div className="py-16 text-center text-xs text-muted-foreground">
            <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2 text-blue-600" />
            Memuat data users...
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50">
                    <TableHead className="text-xs px-4 py-3 w-[60px] text-center">No.</TableHead>
                    <TableHead className="text-xs px-4 py-3">User Information</TableHead>
                    <TableHead className="text-xs px-4 py-3">WhatsApp</TableHead>
                    <TableHead className="text-xs px-4 py-3 text-center">Role</TableHead>
                    <TableHead className="text-xs px-4 py-3 text-center">Status</TableHead>
                    <TableHead className="text-xs px-4 py-3 text-right w-[150px]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedUsers.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-10 text-xs text-muted-foreground">
                        Tidak ada data users
                      </TableCell>
                    </TableRow>
                  ) : (
                    paginatedUsers.map((u, idx) => {
                      const role = getRoleById(u.role_id)
                      return (
                        <TableRow key={u.id} className="hover:bg-muted/40">
                          <TableCell className="px-4 py-2.5 text-center text-xs text-muted-foreground font-mono">
                            {(currentPage - 1) * itemsPerPage + idx + 1}
                          </TableCell>
                          <TableCell className="px-4 py-2.5">
                            <div className="flex items-center gap-3">
                              <div className={`h-9 w-9 rounded-full ${getAvatarBg(u.name)} text-white flex items-center justify-center text-xs font-bold shrink-0`}>
                                {getInitials(u.name)}
                              </div>
                              <div>
                                <div className="text-sm font-semibold text-foreground">{u.name}</div>
                                <div className="text-xs text-muted-foreground">{u.email}</div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell className="px-4 py-2.5 text-xs font-mono text-muted-foreground">
                            {u.phone_no || '-'}
                          </TableCell>
                          <TableCell className="px-4 py-2.5 text-center">
                            <span
                              className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${
                                role
                                  ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                                  : 'bg-amber-50 text-amber-700 border-amber-200'
                              }`}
                            >
                              {role ? <ShieldCheck className="h-3 w-3" /> : <ShieldAlert className="h-3 w-3" />}
                              {role?.name || 'No Role'}
                            </span>
                          </TableCell>
                          <TableCell className="px-4 py-2.5 text-center">
                            <Badge
                              className={
                                role
                                  ? 'bg-emerald-100 text-emerald-800 hover:bg-emerald-100'
                                  : 'bg-amber-100 text-amber-800 hover:bg-amber-100'
                              }
                            >
                              {role ? 'Active' : 'Incomplete'}
                            </Badge>
                          </TableCell>
                          <TableCell className="px-4 py-2.5 text-right">
                            <div className="flex items-center justify-end gap-1">
                              {canEdit && (
                                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openDialog(u)} title="Edit">
                                  <Pencil className="h-3.5 w-3.5 text-blue-600" />
                                </Button>
                              )}
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-7 w-7"
                                onClick={() => {
                                  setSelectedUser(u)
                                  setShowDetailDialog(true)
                                }}
                                title="Detail"
                              >
                                <Eye className="h-3.5 w-3.5 text-cyan-600" />
                              </Button>
                              {canDelete && (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    setItemToDelete(u)
                                    setShowDeleteDialog(true)
                                  }}
                                  title="Hapus"
                                >
                                  <Trash2 className="h-3.5 w-3.5 text-red-500" />
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

            <div className="flex items-center justify-between px-4 py-3 border-t text-xs text-muted-foreground">
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
      </Card>

      {/* Add / Edit User Dialog */}
      <Dialog open={showFormDialog} onOpenChange={setShowFormDialog}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <UserPlus className="h-5 w-5 text-blue-600" />
              {isEdit ? 'Edit User' : 'Tambah User Baru'}
            </DialogTitle>
            <DialogDescription>
              {isEdit ? 'Ubah informasi user dan role' : 'Tambahkan user baru ke sistem'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            {formError && (
              <div className="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg p-3">
                {formError}
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Nama Lengkap *</Label>
                <Input
                  value={formData.name}
                  onChange={(e) => setFormData((p) => ({ ...p, name: e.target.value }))}
                  placeholder="Masukkan nama lengkap"
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">Email Address *</Label>
                <Input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData((p) => ({ ...p, email: e.target.value }))}
                  placeholder="user@example.com"
                />
              </div>
            </div>

            <div>
              <Label className="text-sm mb-1 block">No. Telepon (WhatsApp)</Label>
              <Input
                value={formData.phone_no}
                onChange={(e) => setFormData((p) => ({ ...p, phone_no: e.target.value }))}
                placeholder="Contoh: 6281312345678"
              />
              <p className="text-[11px] text-muted-foreground mt-1">Format: 62xxxx (tanpa + atau 0 di depan)</p>
            </div>

            <Separator />

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Password {!isEdit && '*'}</Label>
                <Input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData((p) => ({ ...p, password: e.target.value }))}
                  placeholder={isEdit ? 'Kosongkan jika tidak diubah' : 'Contoh: MyPassword123!'}
                />
                <p className="text-[11px] text-muted-foreground mt-1">
                  Min. 8 karakter, kapital, kecil, angka, simbol, tanpa spasi
                </p>
              </div>
              <div>
                <Label className="text-sm mb-1 block">Assign Role</Label>
                <Select
                  value={formData.role_id || '__none__'}
                  onValueChange={(v) => setFormData((p) => ({ ...p, role_id: v === '__none__' ? '' : v }))}
                >
                  <SelectTrigger className="w-full"><SelectValue placeholder="Pilih role..." /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">-- No Role --</SelectItem>
                    {roles.map((r) => (
                      <SelectItem key={r.id} value={String(r.id)}>{r.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowFormDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveUser} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              {isEdit ? 'Simpan Perubahan' : 'Buat User'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* User Detail Dialog */}
      <Dialog open={showDetailDialog} onOpenChange={setShowDetailDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Detail User</DialogTitle>
            <DialogDescription>Informasi lengkap user dan hak akses role</DialogDescription>
          </DialogHeader>
          {selectedUser && (
            <div className="space-y-4 py-2">
              <div className="flex items-center gap-4">
                <div className={`h-14 w-14 rounded-full ${getAvatarBg(selectedUser.name)} text-white flex items-center justify-center text-lg font-bold shrink-0`}>
                  {getInitials(selectedUser.name)}
                </div>
                <div>
                  <h3 className="text-base font-bold text-foreground">{selectedUser.name}</h3>
                  <p className="text-xs text-muted-foreground">{selectedUser.email}</p>
                  {selectedUser.phone_no && (
                    <Badge variant="secondary" className="mt-1 text-xs bg-emerald-50 text-emerald-700">
                      <Phone className="h-3 w-3 mr-1" /> {selectedUser.phone_no}
                    </Badge>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 text-xs">
                <div className="border rounded-lg p-3">
                  <div className="text-muted-foreground">User ID</div>
                  <div className="font-mono font-bold text-sm mt-0.5">#{selectedUser.id}</div>
                </div>
                <div className="border rounded-lg p-3">
                  <div className="text-muted-foreground">Role</div>
                  <div className="font-bold text-sm mt-0.5">{getRoleById(selectedUser.role_id)?.name || 'No Role'}</div>
                </div>
              </div>

              {(() => {
                const perms = getRoleById(selectedUser.role_id)?.permissions || []
                if (perms.length === 0) return null
                return (
                  <div>
                    <div className="text-xs font-semibold text-muted-foreground mb-2 flex items-center gap-1">
                      <KeyRound className="h-3.5 w-3.5" /> Permissions ({perms.length})
                    </div>
                    <div className="flex flex-wrap gap-1.5 max-h-40 overflow-y-auto">
                      {perms.map((p) => (
                        <Badge key={p.id} variant="outline" className="text-[11px]">
                          {p.name}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )
              })()}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDetailDialog(false)}>Tutup</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">Konfirmasi Hapus User</DialogTitle>
            <DialogDescription>
              Anda akan menghapus user <strong>{itemToDelete?.name}</strong> ({itemToDelete?.email}). Tindakan ini tidak dapat dibatalkan.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>Batal</Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Ya, Hapus User
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
