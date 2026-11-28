import { useCallback, useEffect, useMemo, useState } from 'react'
import { Edit, Loader2, Package, Plus, RefreshCw, Search, Tag, Trash2, Wifi } from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
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

export interface HargaLayanan {
  id_brand: string
  brand: string
  pajak: number
  xendit_key_name: string
}

export interface PaketLayanan {
  id?: number
  id_brand: string
  nama_paket: string
  kecepatan: number
  harga: number
}

function formatCurrency(value: number) {
  if (Number.isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value)
}

export function HargaLayananView() {
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_brand_&_paket')
  const canEdit = hasPermission('edit_brand_&_paket')
  const canDelete = hasPermission('delete_brand_&_paket')
  const [brands, setBrands] = useState<HargaLayanan[]>([])
  const [packages, setPackages] = useState<PaketLayanan[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [errorMsg, setErrorMsg] = useState('')

  // Brand Dialog
  const [brandDialogOpen, setBrandDialogOpen] = useState(false)
  const [editingBrand, setEditingBrand] = useState<HargaLayanan | null>(null)
  const [brandForm, setBrandForm] = useState<HargaLayanan>({
    id_brand: '',
    brand: '',
    pajak: 11,
    xendit_key_name: '',
  })

  // Package Dialog
  const [pkgDialogOpen, setPkgDialogOpen] = useState(false)
  const [editingPkg, setEditingPkg] = useState<PaketLayanan | null>(null)
  const [pkgForm, setPkgForm] = useState<PaketLayanan>({
    id_brand: '',
    nama_paket: '',
    kecepatan: 10,
    harga: 0,
  })

  // Delete Confirm
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'brand'; item: HargaLayanan } | { type: 'pkg'; item: PaketLayanan } | null>(null)

  const fetchAll = useCallback(async () => {
    setLoading(true)
    setErrorMsg('')
    try {
      const [bRes, pRes] = await Promise.all([
        apiClient.get<HargaLayanan[]>('/harga_layanan'),
        apiClient.get<PaketLayanan[]>('/paket_layanan'),
      ])
      setBrands(
        (bRes.data || []).map((b) => ({
          ...b,
          id_brand: b.id_brand.trim(),
        }))
      )
      setPackages(pRes.data || [])
    } catch (err) {
      console.error('Failed to load brand/paket:', err)
      setErrorMsg('Gagal memuat data brand dan paket layanan.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchAll()
  }, [fetchAll])

  const brandPackagesMap = useMemo(() => {
    const q = searchQuery.toLowerCase().trim()
    return brands.map((brand) => {
      const brandPkgs = packages.filter((p) => p.id_brand === brand.id_brand)
      const filtered = q
        ? brandPkgs.filter(
            (p) =>
              p.nama_paket.toLowerCase().includes(q) ||
              String(p.kecepatan).includes(q) ||
              brand.brand.toLowerCase().includes(q) ||
              brand.id_brand.toLowerCase().includes(q)
          )
        : brandPkgs
      return { brand, packages: filtered, totalCount: brandPkgs.length }
    })
  }, [brands, packages, searchQuery])

  const openBrandDialog = (item?: HargaLayanan) => {
    if (item) {
      setEditingBrand(item)
      setBrandForm({ ...item })
    } else {
      setEditingBrand(null)
      setBrandForm({ id_brand: '', brand: '', pajak: 11, xendit_key_name: '' })
    }
    setBrandDialogOpen(true)
  }

  const saveBrand = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      if (editingBrand) {
        await apiClient.patch(`/harga_layanan/${editingBrand.id_brand}`, brandForm)
      } else {
        await apiClient.post('/harga_layanan', brandForm)
      }
      setBrandDialogOpen(false)
      await fetchAll()
    } catch (err) {
      console.error('Error saving brand:', err)
      setErrorMsg('Gagal menyimpan brand.')
    } finally {
      setSaving(false)
    }
  }

  const openPkgDialog = (item?: PaketLayanan, brand?: HargaLayanan) => {
    if (item) {
      setEditingPkg(item)
      setPkgForm({ ...item })
    } else {
      setEditingPkg(null)
      setPkgForm({
        id_brand: brand?.id_brand || brands[0]?.id_brand || '',
        nama_paket: '',
        kecepatan: 10,
        harga: 0,
      })
    }
    setPkgDialogOpen(true)
  }

  const savePkg = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      if (editingPkg?.id) {
        await apiClient.patch(`/paket_layanan/${editingPkg.id}`, pkgForm)
      } else {
        await apiClient.post('/paket_layanan', pkgForm)
      }
      setPkgDialogOpen(false)
      await fetchAll()
    } catch (err) {
      console.error('Error saving package:', err)
      setErrorMsg('Gagal menyimpan paket layanan.')
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    try {
      if (deleteTarget.type === 'brand') {
        await apiClient.delete(`/harga_layanan/${deleteTarget.item.id_brand}`)
      } else if (deleteTarget.item.id) {
        await apiClient.delete(`/paket_layanan/${deleteTarget.item.id}`)
      }
      setDeleteTarget(null)
      await fetchAll()
    } catch (err) {
      console.error('Delete failed:', err)
      setErrorMsg('Gagal menghapus data. Pastikan tidak ada pelanggan yang masih terhubung.')
      setDeleteTarget(null)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Brand & Paket Layanan</h2>
          <p className="text-sm text-muted-foreground">
            Kelola brand provider internet dan paket kecepatan beserta tarifnya
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Badge variant="secondary" className="gap-1.5">
            <Tag className="h-3.5 w-3.5" />
            {brands.length} Brand
          </Badge>
          <Badge variant="secondary" className="gap-1.5">
            <Package className="h-3.5 w-3.5" />
            {packages.length} Paket
          </Badge>
          <Button variant="outline" size="sm" onClick={() => void fetchAll()} disabled={loading}>
            <RefreshCw className={`mr-1.5 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Segarkan
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openBrandDialog()}>
              <Plus className="mr-1.5 h-4 w-4" />
              Tambah Brand
            </Button>
          )}
        </div>
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search className="h-4 w-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
        <Input
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Cari nama paket, kecepatan (20 Mbps), atau brand..."
          className="pl-9"
        />
      </div>

      {errorMsg && (
        <div className="rounded-md border border-destructive/40 bg-destructive/10 px-4 py-2.5 text-sm text-destructive">
          {errorMsg}
        </div>
      )}

      {/* List */}
      {loading ? (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-44 w-full rounded-xl" />
          ))}
        </div>
      ) : brands.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center space-y-3">
            <Package className="h-10 w-10 text-muted-foreground mx-auto" />
            <div className="font-semibold">Belum Ada Brand Provider</div>
            <p className="text-sm text-muted-foreground">
              Tambahkan brand provider pertama untuk mulai mengatur paket layanan.
            </p>
            {canCreate && (
              <Button onClick={() => openBrandDialog()}>
                <Plus className="mr-1.5 h-4 w-4" />
                Tambah Brand Baru
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-5">
          {brandPackagesMap.map(({ brand, packages: pkgs, totalCount }) => (
            <Card key={brand.id_brand} className="overflow-hidden">
              <CardHeader className="bg-muted/35 border-b py-4 px-5 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="space-y-1">
                  <div className="flex items-center gap-2.5 flex-wrap">
                    <CardTitle className="text-lg font-bold tracking-tight">{brand.brand}</CardTitle>
                    <Badge variant="outline" className="font-mono text-[11px]">
                      {brand.id_brand}
                    </Badge>
                    <Badge variant="secondary" className="text-[11px]">
                      PPN {brand.pajak}%
                    </Badge>
                    {brand.xendit_key_name && (
                      <span className="text-xs text-muted-foreground">
                        • Xendit: <strong className="text-foreground font-medium">{brand.xendit_key_name}</strong>
                      </span>
                    )}
                  </div>
                  <CardDescription className="text-xs">
                    {totalCount} paket layanan terdaftar
                  </CardDescription>
                </div>

                <div className="flex items-center gap-1.5">
                  {canCreate && (
                    <Button size="sm" variant="outline" onClick={() => openPkgDialog(undefined, brand)}>
                      <Plus className="mr-1.5 h-3.5 w-3.5" />
                      Tambah Paket
                    </Button>
                  )}
                  {canEdit && (
                    <Button size="icon" variant="ghost" onClick={() => openBrandDialog(brand)} title="Edit Brand">
                      <Edit className="h-4 w-4" />
                    </Button>
                  )}
                  {canDelete && (
                    <Button
                      size="icon"
                      variant="ghost"
                      className="text-destructive hover:text-destructive"
                      onClick={() => setDeleteTarget({ type: 'brand', item: brand })}
                      title="Hapus Brand"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </CardHeader>

              <CardContent className="p-5">
                {pkgs.length === 0 ? (
                  <div className="py-8 text-center text-sm text-muted-foreground border border-dashed rounded-lg">
                    {searchQuery ? 'Tidak ada paket yang cocok dengan pencarian.' : 'Belum ada paket untuk brand ini.'}
                  </div>
                ) : (
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3.5">
                    {pkgs.map((pkg) => {
                      const hargaSetelahPajak = Math.round(pkg.harga * (1 + (brand.pajak || 0) / 100))
                      return (
                        <div
                          key={pkg.id}
                          className="group rounded-lg border bg-card p-4 flex flex-col justify-between gap-4 hover:border-primary/50 transition-colors"
                        >
                          <div className="space-y-2.5">
                            <div className="flex items-start justify-between gap-2">
                              <div className="inline-flex items-center gap-1.5 rounded-md bg-primary/10 text-primary px-2.5 py-1 text-xs font-semibold">
                                <Wifi className="h-3.5 w-3.5" />
                                <span>{pkg.kecepatan} Mbps</span>
                              </div>
                              <div className="flex items-center gap-0.5 opacity-80 group-hover:opacity-100">
                                {canEdit && (
                                  <Button
                                    size="icon"
                                    variant="ghost"
                                    className="h-7 w-7"
                                    onClick={() => openPkgDialog(pkg, brand)}
                                    title="Edit Paket"
                                  >
                                    <Edit className="h-3.5 w-3.5" />
                                  </Button>
                                )}
                                {canDelete && (
                                  <Button
                                    size="icon"
                                    variant="ghost"
                                    className="h-7 w-7 text-destructive hover:text-destructive"
                                    onClick={() => setDeleteTarget({ type: 'pkg', item: pkg })}
                                    title="Hapus Paket"
                                  >
                                    <Trash2 className="h-3.5 w-3.5" />
                                  </Button>
                                )}
                              </div>
                            </div>

                            <div className="font-semibold text-sm leading-snug">{pkg.nama_paket}</div>
                          </div>

                          <div className="pt-3 border-t space-y-0.5">
                            <div className="flex items-baseline gap-1">
                              <span className="text-xl font-bold tracking-tight">{formatCurrency(pkg.harga)}</span>
                              <span className="text-xs text-muted-foreground">/ bln</span>
                            </div>
                            <div className="text-[11px] text-muted-foreground">
                              Termasuk PPN {brand.pajak}%:{' '}
                              <span className="font-medium text-foreground">{formatCurrency(hargaSetelahPajak)}</span>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Brand Dialog */}
      <Dialog open={brandDialogOpen} onOpenChange={setBrandDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingBrand ? 'Edit Brand' : 'Tambah Brand Baru'}</DialogTitle>
            <DialogDescription>Atur identitas brand, persentase pajak, dan key Xendit.</DialogDescription>
          </DialogHeader>
          <form onSubmit={saveBrand} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="id_brand">ID Brand</Label>
              <Input
                id="id_brand"
                value={brandForm.id_brand}
                onChange={(e) => setBrandForm({ ...brandForm, id_brand: e.target.value })}
                disabled={!!editingBrand}
                placeholder="contoh: ajnusa"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="brand_name">Nama Brand</Label>
              <Input
                id="brand_name"
                value={brandForm.brand}
                onChange={(e) => setBrandForm({ ...brandForm, brand: e.target.value })}
                placeholder="contoh: JELANTIK"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="pajak">Pajak (%)</Label>
                <Input
                  id="pajak"
                  type="number"
                  step="0.1"
                  value={brandForm.pajak}
                  onChange={(e) => setBrandForm({ ...brandForm, pajak: Number(e.target.value) })}
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="xendit_key">Xendit Key Name</Label>
                <Input
                  id="xendit_key"
                  value={brandForm.xendit_key_name}
                  onChange={(e) => setBrandForm({ ...brandForm, xendit_key_name: e.target.value })}
                  placeholder="contoh: JELANTIK"
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setBrandDialogOpen(false)}>
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

      {/* Package Dialog */}
      <Dialog open={pkgDialogOpen} onOpenChange={setPkgDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingPkg ? 'Edit Paket Layanan' : 'Tambah Paket Layanan'}</DialogTitle>
            <DialogDescription>Tentukan nama paket, kecepatan (Mbps), dan tarif dasar.</DialogDescription>
          </DialogHeader>
          <form onSubmit={savePkg} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="nama_paket">Nama Paket</Label>
              <Input
                id="nama_paket"
                value={pkgForm.nama_paket}
                onChange={(e) => setPkgForm({ ...pkgForm, nama_paket: e.target.value })}
                placeholder="contoh: Internet 20 Mbps"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="kecepatan">Kecepatan (Mbps)</Label>
                <Input
                  id="kecepatan"
                  type="number"
                  value={pkgForm.kecepatan}
                  onChange={(e) => setPkgForm({ ...pkgForm, kecepatan: Number(e.target.value) })}
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="harga">Harga Dasar (Rp)</Label>
                <Input
                  id="harga"
                  type="number"
                  value={pkgForm.harga}
                  onChange={(e) => setPkgForm({ ...pkgForm, harga: Number(e.target.value) })}
                  required
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setPkgDialogOpen(false)}>
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

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Konfirmasi Hapus</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteTarget?.type === 'brand'
                ? `Yakin ingin menghapus brand "${deleteTarget.item.brand}"? Semua paket terkait juga akan terhapus.`
                : `Yakin ingin menghapus paket "${deleteTarget?.item.nama_paket}"?`}
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
