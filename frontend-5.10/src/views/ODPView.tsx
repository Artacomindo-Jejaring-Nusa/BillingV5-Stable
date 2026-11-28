import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
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
  MapPin,
  Plus,
  RefreshCw,
  Search,
  Table as TableIcon,
  Pencil,
  Trash2,
  GitFork,
  Loader2,
  CheckCircle2,
  AlertCircle,
  X,
  Layers,
  Move,
  Map as MapIcon,
  Satellite,
  Crosshair,
  Server,
  Router,
  User,
  Cable,
} from 'lucide-react'

interface OltNode {
  id: number
  nama_olt: string
  ip_address?: string
  latitude?: number | null
  longitude?: number | null
}

interface ServerNode {
  id: number
  name: string
  host_ip?: string
  latitude?: number | null
  longitude?: number | null
}

interface OdcNode {
  id: number
  kode_odc: string
  nama_odc: string
  alamat?: string | null
  kapasitas_port: number
  latitude?: number | null
  longitude?: number | null
  olt_id?: number | null
}

interface OdpNode {
  id: number
  kode_odp: string
  alamat?: string | null
  kapasitas_port: number
  port_terpakai?: number
  latitude?: number | null
  longitude?: number | null
  olt_id: number
  odc_id?: number | null
  parent_odp_id?: number | null
  olt?: { id: number; nama_olt: string }
  odc?: { id: number; kode_odc: string; nama_odc: string }
}

interface PelangganNode {
  id: number
  nama: string
  alamat: string
  no_telp: string
  latitude?: number | null
  longitude?: number | null
  odp_id?: number | null
  kode_odp?: string
  port_odp?: number | null
  username_pppoe?: string
  status: string
}

function getCapacityStatus(used = 0, total = 8) {
  const pct = total > 0 ? Math.round((used / total) * 100) : 0
  if (pct >= 90) return { pct, color: '#ef4444', bar: 'bg-red-500', badge: 'bg-red-100 text-red-800 border-red-200', label: 'Penuh' }
  if (pct >= 70) return { pct, color: '#f59e0b', bar: 'bg-amber-500', badge: 'bg-amber-100 text-amber-800 border-amber-200', label: 'Tinggi' }
  return { pct, color: '#10b981', bar: 'bg-emerald-500', badge: 'bg-emerald-100 text-emerald-800 border-emerald-200', label: 'Tersedia' }
}

export function ODPView() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const canCreate = hasPermission('create_odp_management')
  const canEdit = hasPermission('edit_odp_management')
  const canDelete = hasPermission('delete_odp_management')

  const [odps, setOdps] = useState<OdpNode[]>([])
  const [olts, setOlts] = useState<OltNode[]>([])
  const [servers, setServers] = useState<ServerNode[]>([])
  const [odcs, setOdcs] = useState<OdcNode[]>([])
  const [pelanggans, setPelanggans] = useState<PelangganNode[]>([])

  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const [showTable, setShowTable] = useState(true)
  const [tableSearch, setTableSearch] = useState('')
  const [mapSearch, setMapSearch] = useState('')
  const [selectedOdp, setSelectedOdp] = useState<OdpNode | null>(null)

  // Map controls
  const [basemap, setBasemap] = useState<'osm' | 'satellite'>('osm')
  const [isEditMode, setIsEditMode] = useState(false)
  const [layers, setLayers] = useState({
    server: true,
    olt: true,
    odc: true,
    odp: true,
    pelanggan: true,
    lines: true,
  })

  // Dialogs
  const [showFormDialog, setShowFormDialog] = useState(false)
  const [editedItem, setEditedItem] = useState<Partial<OdpNode>>({})
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [itemToDelete, setItemToDelete] = useState<OdpNode | null>(null)

  // Feedback
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const feedbackTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Leaflet refs
  const mapContainerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<L.Map | null>(null)
  const osmLayerRef = useRef<L.TileLayer | null>(null)
  const satLayerRef = useRef<L.TileLayer | null>(null)
  const layerGroupRef = useRef<L.LayerGroup | null>(null)

  const notify = (text: string, type: 'success' | 'error') => {
    setFeedback({ type, text })
    if (feedbackTimerRef.current) clearTimeout(feedbackTimerRef.current)
    feedbackTimerRef.current = setTimeout(() => setFeedback(null), 4000)
  }

  const fetchAllData = useCallback(async () => {
    setLoading(true)
    try {
      try {
        const mapRes = await apiClient.get('/network/map-overview')
        const data = mapRes.data?.data
        if (data) {
          setServers(data.servers || [])
          setOlts(data.olts || [])
          setOdcs(data.odcs || [])
          setOdps(data.odps || [])
          setPelanggans(data.pelanggans || [])
          return
        }
      } catch {
        // Fallback to individual endpoints
      }

      const [odpRes, oltRes] = await Promise.all([apiClient.get('/odp'), apiClient.get('/olt')])
      setOdps(Array.isArray(odpRes.data) ? odpRes.data : odpRes.data.data || [])
      setOlts(Array.isArray(oltRes.data) ? oltRes.data : oltRes.data.data || [])

      try {
        const odcRes = await apiClient.get('/odc')
        setOdcs(Array.isArray(odcRes.data) ? odcRes.data : odcRes.data.data || [])
      } catch {
        setOdcs([])
      }
    } catch {
      notify('Gagal memuat data jaringan ODP', 'error')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAllData()
  }, [fetchAllData])

  // Initialize Leaflet Map once
  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return

    const map = L.map(mapContainerRef.current, {
      center: [-6.2383, 106.9756],
      zoom: 13,
      zoomControl: true,
    })

    const osm = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    })
    const sat = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
      attribution: '&copy; Esri',
      maxZoom: 19,
    })

    osm.addTo(map)
    const group = L.layerGroup().addTo(map)

    mapRef.current = map
    osmLayerRef.current = osm
    satLayerRef.current = sat
    layerGroupRef.current = group

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [])

  // Handle basemap switch
  useEffect(() => {
    const map = mapRef.current
    const osm = osmLayerRef.current
    const sat = satLayerRef.current
    if (!map || !osm || !sat) return

    if (basemap === 'osm') {
      if (map.hasLayer(sat)) map.removeLayer(sat)
      if (!map.hasLayer(osm)) osm.addTo(map)
    } else {
      if (map.hasLayer(osm)) map.removeLayer(osm)
      if (!map.hasLayer(sat)) sat.addTo(map)
    }
  }, [basemap])

  // Render markers & lines on map
  useEffect(() => {
    const map = mapRef.current
    const group = layerGroupRef.current
    if (!map || !group) return

    group.clearLayers()
    const bounds: L.LatLngExpression[] = []

    // Lines
    if (layers.lines) {
      odcs.forEach((odc) => {
        if (!odc.latitude || !odc.longitude || !odc.olt_id) return
        const olt = olts.find((o) => o.id === odc.olt_id)
        if (olt?.latitude && olt?.longitude) {
          L.polyline(
            [
              [olt.latitude, olt.longitude],
              [odc.latitude, odc.longitude],
            ],
            { color: '#6366f1', weight: 2.5, dashArray: '6 4', opacity: 0.7 }
          ).addTo(group)
        }
      })

      odps.forEach((odp) => {
        if (!odp.latitude || !odp.longitude) return
        const odc = odp.odc_id ? odcs.find((c) => c.id === odp.odc_id) : null
        const olt = olts.find((o) => o.id === odp.olt_id)
        const parentLat = odc?.latitude ?? olt?.latitude
        const parentLng = odc?.longitude ?? olt?.longitude
        if (parentLat && parentLng) {
          L.polyline(
            [
              [parentLat, parentLng],
              [odp.latitude, odp.longitude],
            ],
            { color: '#0ea5e9', weight: 2, opacity: 0.7 }
          ).addTo(group)
        }
      })

      if (layers.pelanggan) {
        pelanggans.forEach((p) => {
          if (!p.latitude || !p.longitude || !p.odp_id) return
          const odp = odps.find((o) => o.id === p.odp_id)
          if (odp?.latitude && odp?.longitude) {
            L.polyline(
              [
                [odp.latitude, odp.longitude],
                [p.latitude, p.longitude],
              ],
              { color: '#94a3b8', weight: 1.2, dashArray: '3 3', opacity: 0.6 }
            ).addTo(group)
          }
        })
      }
    }

    // Servers
    if (layers.server) {
      servers.forEach((s) => {
        if (!s.latitude || !s.longitude) return
        bounds.push([s.latitude, s.longitude])
        L.circleMarker([s.latitude, s.longitude], {
          radius: 8,
          fillColor: '#0f172a',
          color: '#ffffff',
          weight: 2,
          fillOpacity: 0.95,
        })
          .bindTooltip(`Server: ${s.name}`, { direction: 'top' })
          .addTo(group)
      })
    }

    // OLTs
    if (layers.olt) {
      olts.forEach((o) => {
        if (!o.latitude || !o.longitude) return
        bounds.push([o.latitude, o.longitude])
        L.circleMarker([o.latitude, o.longitude], {
          radius: 9,
          fillColor: '#7c3aed',
          color: '#ffffff',
          weight: 2,
          fillOpacity: 0.95,
        })
          .bindTooltip(`OLT: ${o.nama_olt}`, { direction: 'top' })
          .addTo(group)
      })
    }

    // ODCs
    if (layers.odc) {
      odcs.forEach((c) => {
        if (!c.latitude || !c.longitude) return
        bounds.push([c.latitude, c.longitude])
        L.circleMarker([c.latitude, c.longitude], {
          radius: 8,
          fillColor: '#f97316',
          color: '#ffffff',
          weight: 2,
          fillOpacity: 0.95,
        })
          .bindTooltip(`ODC: ${c.nama_odc}`, { direction: 'top' })
          .addTo(group)
      })
    }

    // ODPs
    if (layers.odp) {
      odps.forEach((odp) => {
        if (!odp.latitude || !odp.longitude) return
        bounds.push([odp.latitude, odp.longitude])
        const cap = getCapacityStatus(odp.port_terpakai || 0, odp.kapasitas_port || 8)

        if (isEditMode) {
          const icon = L.divIcon({
            className: 'custom-odp-drag-icon',
            html: `<div style="width:18px;height:18px;border-radius:9999px;background:${cap.color};border:2px solid #fff;box-shadow:0 2px 6px rgba(0,0,0,.35);cursor:move;"></div>`,
            iconSize: [18, 18],
            iconAnchor: [9, 9],
          })
          const marker = L.marker([odp.latitude, odp.longitude], { icon, draggable: true })
          marker.on('dragend', async () => {
            const pos = marker.getLatLng()
            try {
              await apiClient.patch(`/odp/${odp.id}`, {
                latitude: parseFloat(pos.lat.toFixed(6)),
                longitude: parseFloat(pos.lng.toFixed(6)),
              })
              setOdps((prev) =>
                prev.map((o) => (o.id === odp.id ? { ...o, latitude: pos.lat, longitude: pos.lng } : o))
              )
              notify(`Koordinat ${odp.kode_odp} diperbarui (${pos.lat.toFixed(5)}, ${pos.lng.toFixed(5)})`, 'success')
            } catch {
              notify('Gagal menyimpan koordinat baru', 'error')
            }
          })
          marker.bindTooltip(`${odp.kode_odp} (Drag untuk pindah)`, { direction: 'top' }).addTo(group)
        } else {
          const marker = L.circleMarker([odp.latitude, odp.longitude], {
            radius: selectedOdp?.id === odp.id ? 10 : 7,
            fillColor: cap.color,
            color: selectedOdp?.id === odp.id ? '#1d4ed8' : '#ffffff',
            weight: selectedOdp?.id === odp.id ? 3 : 2,
            fillOpacity: 0.95,
          })
          marker.on('click', () => setSelectedOdp(odp))
          marker
            .bindTooltip(`${odp.kode_odp} (${odp.port_terpakai || 0}/${odp.kapasitas_port} port)`, { direction: 'top' })
            .addTo(group)
        }
      })
    }

    // Pelanggans
    if (layers.pelanggan) {
      pelanggans.forEach((p) => {
        if (!p.latitude || !p.longitude) return
        L.circleMarker([p.latitude, p.longitude], {
          radius: 4.5,
          fillColor: p.status === 'aktif' || p.status === 'active' ? '#3b82f6' : '#ef4444',
          color: '#ffffff',
          weight: 1.5,
          fillOpacity: 0.9,
        })
          .bindTooltip(`${p.nama}${p.port_odp ? ` (Port #${p.port_odp})` : ''}`, { direction: 'top' })
          .addTo(group)
      })
    }

    if (bounds.length > 0 && !selectedOdp) {
      map.fitBounds(L.latLngBounds(bounds), { padding: [30, 30], maxZoom: 16 })
    }
  }, [odps, olts, servers, odcs, pelanggans, layers, isEditMode, selectedOdp])

  // Map click when in edit mode to add ODP
  useEffect(() => {
    const map = mapRef.current
    if (!map) return
    const handleClick = (e: L.LeafletMouseEvent) => {
      if (!isEditMode) return
      setEditedItem({
        kapasitas_port: 8,
        latitude: parseFloat(e.latlng.lat.toFixed(6)),
        longitude: parseFloat(e.latlng.lng.toFixed(6)),
        olt_id: olts.length > 0 ? olts[0].id : undefined,
      })
      setShowFormDialog(true)
    }
    map.on('click', handleClick)
    return () => {
      map.off('click', handleClick)
    }
  }, [isEditMode, olts])

  const focusOdpOnMap = (odp: OdpNode) => {
    setSelectedOdp(odp)
    if (mapRef.current && odp.latitude && odp.longitude) {
      mapRef.current.setView([odp.latitude, odp.longitude], 17, { animate: true })
    }
  }

  const handleMapSearch = () => {
    const q = mapSearch.trim().toLowerCase()
    if (!q) return
    const foundOdp = odps.find(
      (o) => o.kode_odp.toLowerCase().includes(q) || (o.alamat || '').toLowerCase().includes(q)
    )
    if (foundOdp) {
      focusOdpOnMap(foundOdp)
      return
    }
    const foundCust = pelanggans.find((p) => p.nama.toLowerCase().includes(q))
    if (foundCust?.latitude && foundCust?.longitude && mapRef.current) {
      mapRef.current.setView([foundCust.latitude, foundCust.longitude], 18, { animate: true })
      return
    }
    notify(`Tidak ditemukan titik untuk "${mapSearch}"`, 'error')
  }

  // Stats
  const summaryStats = useMemo(() => {
    const totalOdp = odps.length
    const totalPorts = odps.reduce((sum, o) => sum + (o.kapasitas_port || 0), 0)
    const usedPorts = odps.reduce((sum, o) => sum + (o.port_terpakai || 0), 0)
    const avgUtil = totalPorts > 0 ? Math.round((usedPorts / totalPorts) * 100) : 0
    return { totalOdp, totalPorts, usedPorts, avgUtil }
  }, [odps])

  // Customers connected to selected ODP
  const selectedOdpCustomers = useMemo(() => {
    if (!selectedOdp) return []
    return pelanggans.filter(
      (p) => p.odp_id === selectedOdp.id || (p.kode_odp && p.kode_odp === selectedOdp.kode_odp)
    )
  }, [selectedOdp, pelanggans])

  // Filtered table ODPs
  const filteredOdps = useMemo(() => {
    if (!tableSearch.trim()) return odps
    const q = tableSearch.trim().toLowerCase()
    return odps.filter(
      (o) =>
        o.kode_odp.toLowerCase().includes(q) ||
        (o.alamat || '').toLowerCase().includes(q) ||
        (o.olt?.nama_olt || '').toLowerCase().includes(q) ||
        (o.odc?.nama_odc || '').toLowerCase().includes(q)
    )
  }, [odps, tableSearch])

  // CRUD
  const openDialog = (item?: OdpNode) => {
    setEditedItem(
      item
        ? { ...item }
        : {
            kapasitas_port: 8,
            olt_id: olts.length > 0 ? olts[0].id : undefined,
          }
    )
    setShowFormDialog(true)
  }

  const saveOdp = async () => {
    if (!editedItem.kode_odp?.trim() || !editedItem.olt_id) {
      notify('Kode ODP dan Target OLT wajib diisi!', 'error')
      return
    }
    setSaving(true)
    try {
      if (editedItem.id) {
        await apiClient.patch(`/odp/${editedItem.id}`, editedItem)
        notify(`ODP ${editedItem.kode_odp} berhasil diperbarui`, 'success')
      } else {
        await apiClient.post('/odp', editedItem)
        notify(`ODP ${editedItem.kode_odp} berhasil ditambahkan`, 'success')
      }
      setShowFormDialog(false)
      setEditedItem({})
      fetchAllData()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } } }
      notify(e.response?.data?.error || 'Gagal menyimpan ODP', 'error')
    } finally {
      setSaving(false)
    }
  }

  const confirmDelete = async () => {
    if (!itemToDelete) return
    setDeleting(true)
    try {
      await apiClient.delete(`/odp/${itemToDelete.id}`)
      notify(`ODP ${itemToDelete.kode_odp} berhasil dihapus`, 'success')
      if (selectedOdp?.id === itemToDelete.id) setSelectedOdp(null)
      setShowDeleteDialog(false)
      setItemToDelete(null)
      fetchAllData()
    } catch {
      notify('Gagal menghapus ODP', 'error')
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
            <MapPin className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">ODP & GIS Network Mapping</h1>
            <p className="text-xs text-muted-foreground">
              Pemetaan infrastruktur FTTH (Server &gt; OLT &gt; ODC &gt; ODP &gt; Pelanggan) & manajemen splitter port
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setShowTable((v) => !v)}>
            <TableIcon className="h-4 w-4 mr-1.5" />
            {showTable ? 'Sembunyikan Tabel' : 'Tampilkan Tabel'}
          </Button>
          <Button variant="outline" size="sm" onClick={fetchAllData} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          {canCreate && (
            <Button size="sm" onClick={() => openDialog()}>
              <Plus className="h-4 w-4 mr-1.5" />
              Tambah ODP
            </Button>
          )}
        </div>
      </div>

      {/* Compact KPI Strip */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-blue-50 text-blue-600"><MapPin className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Total Titik ODP</div>
            <div className="text-lg font-bold text-foreground">{summaryStats.totalOdp}</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-purple-50 text-purple-600"><Cable className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Kapasitas Port</div>
            <div className="text-lg font-bold text-foreground">{summaryStats.totalPorts} Port</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-emerald-50 text-emerald-600"><User className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Port Terpakai</div>
            <div className="text-lg font-bold text-emerald-600">{summaryStats.usedPorts} Port</div>
          </div>
        </div>
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs">
          <div className="p-2 rounded-lg bg-amber-50 text-amber-600"><GitFork className="h-4 w-4" /></div>
          <div>
            <div className="text-[11px] text-muted-foreground">Utilisasi Jaringan</div>
            <div className="text-lg font-bold text-amber-600">{summaryStats.avgUtil}%</div>
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

      {/* Map-Centric Split View + ODP Splitter Inspector */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <Card className={`${selectedOdp ? 'lg:col-span-8' : 'lg:col-span-12'} shadow-xs overflow-hidden py-0`}>
          {/* Map Toolbar */}
          <div className="p-3 border-b bg-card flex flex-wrap items-center gap-2">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={mapSearch}
                onChange={(e) => setMapSearch(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleMapSearch()}
                placeholder="Cari ODP atau pelanggan di peta..."
                className="pl-9 h-8 text-xs"
              />
            </div>
            <Button variant="secondary" size="sm" className="h-8 text-xs" onClick={handleMapSearch}>
              Cari
            </Button>

            {canEdit && (
              <Button
                variant={isEditMode ? 'destructive' : 'outline'}
                size="sm"
                className="h-8 text-xs"
                onClick={() => setIsEditMode((v) => !v)}
              >
                <Move className="h-3.5 w-3.5 mr-1" />
                {isEditMode ? 'Mode Geser: AKTIF' : 'Atur Posisi'}
              </Button>
            )}

            <div className="inline-flex rounded-lg border p-0.5 bg-muted/50">
              <button
                type="button"
                onClick={() => setBasemap('osm')}
                className={`px-2 py-1 text-xs rounded-md flex items-center gap-1 ${basemap === 'osm' ? 'bg-card shadow-xs font-medium' : 'text-muted-foreground'}`}
              >
                <MapIcon className="h-3.5 w-3.5" /> Peta
              </button>
              <button
                type="button"
                onClick={() => setBasemap('satellite')}
                className={`px-2 py-1 text-xs rounded-md flex items-center gap-1 ${basemap === 'satellite' ? 'bg-card shadow-xs font-medium' : 'text-muted-foreground'}`}
              >
                <Satellite className="h-3.5 w-3.5" /> Satelit
              </button>
            </div>
          </div>

          {/* Layer Toggles Strip */}
          <div className="px-3 py-1.5 border-b bg-muted/50 flex flex-wrap items-center gap-3 text-[11px] text-muted-foreground">
            <span className="font-semibold flex items-center gap-1 text-muted-foreground">
              <Layers className="h-3.5 w-3.5" /> Layer:
            </span>
            {(
              [
                ['server', 'Server'],
                ['olt', 'OLT'],
                ['odc', 'ODC'],
                ['odp', 'ODP'],
                ['pelanggan', 'Pelanggan'],
                ['lines', 'Jalur Kabel'],
              ] as const
            ).map(([key, label]) => (
              <label key={key} className="flex items-center gap-1 cursor-pointer">
                <input
                  type="checkbox"
                  checked={layers[key]}
                  onChange={(e) => setLayers((prev) => ({ ...prev, [key]: e.target.checked }))}
                  className="rounded border-border text-blue-600"
                />
                <span>{label}</span>
              </label>
            ))}
            <div className="ml-auto hidden sm:flex items-center gap-2 text-[10px]">
              <span className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-emerald-500" /> &lt;70%</span>
              <span className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-amber-500" /> 70-89%</span>
              <span className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-red-500" /> &ge;90%</span>
            </div>
          </div>

          {/* Map Container */}
          <div className="relative">
            <div ref={mapContainerRef} className="w-full h-[520px] z-0" />
            {isEditMode && (
              <div className="absolute bottom-3 left-3 z-[400] bg-slate-900/90 text-white text-xs px-3 py-1.5 rounded-lg shadow">
                Mode Edit Aktif: Drag marker ODP untuk pindah posisi, atau klik area peta untuk tambah ODP baru.
              </div>
            )}
          </div>
        </Card>

        {/* Contextual ODP Splitter Inspector */}
        {selectedOdp && (
          <Card className="lg:col-span-4 shadow-xs flex flex-col py-0 overflow-hidden">
            <CardHeader className="py-3 px-4 bg-primary text-primary-foreground flex flex-row items-center justify-between">
              <div>
                <div className="text-xs text-primary-foreground/70">ODP Inspector</div>
                <CardTitle className="text-base font-bold">{selectedOdp.kode_odp}</CardTitle>
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7 text-primary-foreground/70 hover:text-primary-foreground" onClick={() => setSelectedOdp(null)}>
                <X className="h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent className="p-4 space-y-4 flex-1 overflow-y-auto max-h-[540px]">
              <div className="space-y-1.5 text-xs">
                <div className="text-muted-foreground">{selectedOdp.alamat || 'Alamat belum diisi'}</div>
                <div className="font-mono text-muted-foreground flex items-center gap-1">
                  <Crosshair className="h-3.5 w-3.5 text-blue-600" />
                  {selectedOdp.latitude && selectedOdp.longitude
                    ? `${selectedOdp.latitude.toFixed(5)}, ${selectedOdp.longitude.toFixed(5)}`
                    : 'Belum ada koordinat'}
                </div>
                <div className="flex flex-wrap gap-1.5 pt-1">
                  <Badge variant="secondary" className="bg-purple-50 text-purple-700 text-[11px]">
                    <Router className="h-3 w-3 mr-1" /> OLT: {selectedOdp.olt?.nama_olt || '-'}
                  </Badge>
                  {selectedOdp.odc && (
                    <Badge variant="secondary" className="bg-orange-50 text-orange-700 text-[11px]">
                      <Server className="h-3 w-3 mr-1" /> ODC: {selectedOdp.odc.nama_odc}
                    </Badge>
                  )}
                </div>
              </div>

              {/* Capacity Bar */}
              {(() => {
                const cap = getCapacityStatus(selectedOdp.port_terpakai || 0, selectedOdp.kapasitas_port || 8)
                return (
                  <div className="bg-muted/50 border rounded-lg p-3 space-y-1.5">
                    <div className="flex items-center justify-between text-xs">
                      <span className="font-semibold text-foreground/90">Kapasitas Splitter</span>
                      <span className="font-mono font-bold">
                        {selectedOdp.port_terpakai || 0} / {selectedOdp.kapasitas_port || 8} ({cap.pct}%)
                      </span>
                    </div>
                    <div className="h-2 bg-muted rounded-full overflow-hidden">
                      <div className={`h-full rounded-full ${cap.bar}`} style={{ width: `${Math.min(100, cap.pct)}%` }} />
                    </div>
                  </div>
                )
              })()}

              {/* Splitter Port Matrix */}
              <div>
                <div className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-2">
                  Splitter Port Matrix (1:{selectedOdp.kapasitas_port || 8})
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {Array.from({ length: selectedOdp.kapasitas_port || 8 }, (_, i) => i + 1).map((portNum) => {
                    const occupant = selectedOdpCustomers.find((c) => c.port_odp === portNum)
                    const isUsed = Boolean(occupant) || portNum <= (selectedOdp.port_terpakai || 0)
                    return (
                      <div
                        key={portNum}
                        className={`rounded-lg border p-2 text-xs ${
                          occupant
                            ? 'bg-blue-50/70 border-blue-200'
                            : isUsed
                              ? 'bg-emerald-50/70 border-emerald-200'
                              : 'bg-muted/50 border-border'
                        }`}
                      >
                        <div className="flex items-center justify-between font-mono font-bold text-[11px]">
                          <span>Port #{portNum}</span>
                          <span className={`h-2 w-2 rounded-full ${occupant || isUsed ? 'bg-emerald-500' : 'bg-muted-foreground/30'}`} />
                        </div>
                        <div className="mt-1 truncate text-[11px] font-medium text-foreground/90">
                          {occupant ? occupant.nama : isUsed ? 'Terpakai' : 'Kosong / Spare'}
                        </div>
                        {occupant?.username_pppoe && (
                          <div className="text-[10px] font-mono text-muted-foreground truncate">{occupant.username_pppoe}</div>
                        )}
                      </div>
                    )
                  })}
                </div>
              </div>

              <div className="flex items-center gap-2 pt-2 border-t">
                <Button variant="outline" size="sm" className="flex-1" onClick={() => navigate(`/topology/olt/${selectedOdp.olt_id}`)}>
                  <GitFork className="h-3.5 w-3.5 mr-1" /> Topologi
                </Button>
                {canEdit && (
                  <Button variant="outline" size="sm" className="flex-1" onClick={() => openDialog(selectedOdp)}>
                    <Pencil className="h-3.5 w-3.5 mr-1" /> Edit
                  </Button>
                )}
                {canDelete && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-red-600 hover:bg-red-50"
                    onClick={() => {
                      setItemToDelete(selectedOdp)
                      setShowDeleteDialog(true)
                    }}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Collapsible ODP Table */}
      {showTable && (
        <Card className="shadow-xs">
          <CardHeader className="py-3 px-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <CardTitle className="text-sm font-semibold">Daftar Titik ODP Terdaftar</CardTitle>
                <Badge variant="secondary" className="text-xs">{filteredOdps.length} ODP</Badge>
              </div>
              <div className="relative w-full sm:w-72">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  value={tableSearch}
                  onChange={(e) => setTableSearch(e.target.value)}
                  placeholder="Cari kode ODP, alamat, OLT..."
                  className="pl-9 h-8 text-xs"
                />
              </div>
            </div>
          </CardHeader>
          <Separator />
          <div className="overflow-x-auto px-2">
            <Table className="w-full">
              <TableHeader>
                <TableRow className="bg-muted/50">
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[140px]">Kode ODP</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 min-w-[200px]">Alamat / Lokasi</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[170px]">Kapasitas Port</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[130px]">OLT</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[130px]">ODC</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[170px]">Koordinat</TableHead>
                  <TableHead className="text-xs font-semibold px-4 py-3 w-[170px] text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredOdps.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-10 text-xs text-muted-foreground">
                      {loading ? 'Memuat data ODP...' : 'Belum ada data ODP'}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredOdps.map((item) => {
                    const cap = getCapacityStatus(item.port_terpakai || 0, item.kapasitas_port || 8)
                    return (
                      <TableRow key={item.id} className="hover:bg-muted/40 cursor-pointer" onClick={() => focusOdpOnMap(item)}>
                        <TableCell className="px-4 py-3 font-semibold text-blue-700 text-sm">{item.kode_odp}</TableCell>
                        <TableCell className="px-4 py-3 text-xs text-muted-foreground max-w-[240px] truncate">{item.alamat || '-'}</TableCell>
                        <TableCell className="px-4 py-3">
                          <div className="flex items-center gap-2 min-w-[130px]">
                            <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                              <div className={`h-full rounded-full ${cap.bar}`} style={{ width: `${Math.min(100, cap.pct)}%` }} />
                            </div>
                            <span className="text-xs font-mono font-semibold">
                              {item.port_terpakai || 0}/{item.kapasitas_port}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="px-4 py-3">
                          <Badge variant="secondary" className="bg-purple-50 text-purple-700 text-xs">
                            {item.olt?.nama_olt || '-'}
                          </Badge>
                        </TableCell>
                        <TableCell className="px-4 py-3">
                          {item.odc ? (
                            <Badge variant="secondary" className="bg-orange-50 text-orange-700 text-xs">{item.odc.nama_odc}</Badge>
                          ) : (
                            <span className="text-xs text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell className="px-4 py-3 font-mono text-xs text-muted-foreground">
                          {item.latitude && item.longitude
                            ? `${item.latitude.toFixed(5)}, ${item.longitude.toFixed(5)}`
                            : <span className="text-red-500">Belum ada</span>}
                        </TableCell>
                        <TableCell className="px-4 py-3 text-right" onClick={(e) => e.stopPropagation()}>
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="sm" className="h-7 text-xs text-teal-700" onClick={() => navigate(`/topology/olt/${item.olt_id}`)}>
                              <GitFork className="h-3.5 w-3.5 mr-1" /> Topologi
                            </Button>
                            {canEdit && (
                              <Button variant="ghost" size="sm" className="h-7 text-xs text-blue-600" onClick={() => openDialog(item)}>
                                <Pencil className="h-3.5 w-3.5 mr-1" /> Edit
                              </Button>
                            )}
                            {canDelete && (
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 text-xs text-red-600"
                                onClick={() => {
                                  setItemToDelete(item)
                                  setShowDeleteDialog(true)
                                }}
                              >
                                <Trash2 className="h-3.5 w-3.5" />
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
        </Card>
      )}

      {/* Add / Edit ODP Dialog */}
      <Dialog open={showFormDialog} onOpenChange={setShowFormDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editedItem.id ? 'Edit Titik ODP' : 'Tambah Titik ODP Baru'}</DialogTitle>
            <DialogDescription>Atur kode ODP, kapasitas splitter port, relasi OLT/ODC, dan koordinat GIS</DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <div>
              <Label className="text-sm mb-1 block">Kode ODP *</Label>
              <Input
                value={editedItem.kode_odp || ''}
                onChange={(e) => setEditedItem((p) => ({ ...p, kode_odp: e.target.value }))}
                placeholder="Contoh: ODP-TMB-001"
              />
            </div>
            <div>
              <Label className="text-sm mb-1 block">Alamat / Titik Tiang ODP</Label>
              <Input
                value={editedItem.alamat || ''}
                onChange={(e) => setEditedItem((p) => ({ ...p, alamat: e.target.value }))}
                placeholder="Lokasi tiang ODP..."
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label className="text-sm mb-1 block">Kapasitas Port</Label>
                <Input
                  type="number"
                  value={editedItem.kapasitas_port ?? 8}
                  onChange={(e) => setEditedItem((p) => ({ ...p, kapasitas_port: Number(e.target.value) || 8 }))}
                />
              </div>
              <div>
                <Label className="text-sm mb-1 block">Terhubung ke OLT *</Label>
                <Select
                  value={editedItem.olt_id ? String(editedItem.olt_id) : ''}
                  onValueChange={(v) => setEditedItem((p) => ({ ...p, olt_id: Number(v) }))}
                >
                  <SelectTrigger className="w-full"><SelectValue placeholder="Pilih OLT..." /></SelectTrigger>
                  <SelectContent>
                    {olts.map((o) => <SelectItem key={o.id} value={String(o.id)}>{o.nama_olt}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label className="text-sm mb-1 block">Terhubung ke ODC (Opsional)</Label>
              <Select
                value={editedItem.odc_id ? String(editedItem.odc_id) : '__none__'}
                onValueChange={(v) => setEditedItem((p) => ({ ...p, odc_id: v === '__none__' ? null : Number(v) }))}
              >
                <SelectTrigger className="w-full"><SelectValue placeholder="Pilih ODC..." /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__none__">-- Tidak ada --</SelectItem>
                  {odcs.map((c) => <SelectItem key={c.id} value={String(c.id)}>{c.nama_odc}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="bg-muted/50 border rounded-lg p-3 space-y-2">
              <div className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
                <Crosshair className="h-3.5 w-3.5" /> Koordinat GIS
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <Label className="text-xs mb-1 block">Latitude</Label>
                  <Input
                    type="number"
                    step="any"
                    value={editedItem.latitude ?? ''}
                    onChange={(e) => setEditedItem((p) => ({ ...p, latitude: e.target.value ? Number(e.target.value) : null }))}
                    placeholder="-6.238300"
                    className="font-mono text-xs"
                  />
                </div>
                <div>
                  <Label className="text-xs mb-1 block">Longitude</Label>
                  <Input
                    type="number"
                    step="any"
                    value={editedItem.longitude ?? ''}
                    onChange={(e) => setEditedItem((p) => ({ ...p, longitude: e.target.value ? Number(e.target.value) : null }))}
                    placeholder="106.975600"
                    className="font-mono text-xs"
                  />
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowFormDialog(false)} disabled={saving}>Batal</Button>
            <Button onClick={saveOdp} disabled={saving}>
              {saving && <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />}
              Simpan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">Konfirmasi Hapus ODP</DialogTitle>
            <DialogDescription>
              Yakin ingin menghapus ODP <strong>{itemToDelete?.kode_odp}</strong>? Tindakan ini tidak dapat dibatalkan.
            </DialogDescription>
          </DialogHeader>
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
