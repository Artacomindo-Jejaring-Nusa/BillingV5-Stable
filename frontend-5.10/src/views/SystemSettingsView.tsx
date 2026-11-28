import { useCallback, useEffect, useRef, useState } from 'react'
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  Clock,
  History,
  Loader2,
  MapPin,
  MessageSquareText,
  PauseCircle,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Save,
  Settings,
  Trash2,
  Wrench,
  X,
  Zap,
} from 'lucide-react'
import apiClient from '@/services/api'
import { useSettingsStore } from '@/stores/settings'
import {
  fetchAllLocations,
  LOCATIONS_SETTING_KEY,
  type LocationItem,
} from '@/utils/locations'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

interface SchedulerJob {
  key: string
  name: string
  description: string
  schedule: string
  enabled: boolean
  last_run: string
  next_run: string
  is_running: boolean
}

interface ChatTemplate {
  id: number
  shortcut: string
  title: string
  content: string
  icon?: string
  sort_order?: number
}

function formatDateTime(dateStr?: string) {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return 'Belum pernah'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return d.toLocaleString('id-ID', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export function SystemSettingsView() {
  const maintenanceMode = useSettingsStore((s) => s.maintenanceMode)
  const fetchMaintenanceStatus = useSettingsStore((s) => s.fetchMaintenanceStatus)

  const [activeTab, setActiveTab] = useState<'maintenance' | 'scheduler' | 'templates' | 'locations'>('maintenance')
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // Maintenance state
  const [maintenanceActive, setMaintenanceActive] = useState(false)
  const [maintenanceMessage, setMaintenanceMessage] = useState('')
  const [savingMaintenance, setSavingMaintenance] = useState(false)

  // Scheduler state
  const [schedulerEnabled, setSchedulerEnabled] = useState(true)
  const [togglingGlobal, setTogglingGlobal] = useState(false)
  const [loadingJobs, setLoadingJobs] = useState(true)
  const [savingScheduler, setSavingScheduler] = useState(false)
  const [triggeringJob, setTriggeringJob] = useState('')
  const [jobs, setJobs] = useState<SchedulerJob[]>([])
  const [successJobs, setSuccessJobs] = useState<Record<string, boolean>>({})
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Live Chat Templates state
  const [chatTemplates, setChatTemplates] = useState<ChatTemplate[]>([])
  const [loadingTemplates, setLoadingTemplates] = useState(true)
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false)
  const [editingTemplateId, setEditingTemplateId] = useState<number | null>(null)
  const [savingTemplate, setSavingTemplate] = useState(false)
  const [templateToDelete, setTemplateToDelete] = useState<ChatTemplate | null>(null)
  const [deletingTemplate, setDeletingTemplate] = useState(false)
  const [templateForm, setTemplateForm] = useState({
    shortcut: '',
    title: '',
    content: '',
    icon: 'mdi-message-text-outline',
    sort_order: 0,
  })
  const [templateErrors, setTemplateErrors] = useState<{ shortcut?: string; content?: string }>({})

  // Master Locations state
  const [locations, setLocations] = useState<LocationItem[]>([])
  const [loadingLocations, setLoadingLocations] = useState(true)
  const [savingLocations, setSavingLocations] = useState(false)
  const [newLocationName, setNewLocationName] = useState('')

  const showFeedback = useCallback((text: string, type: 'success' | 'error' = 'success') => {
    setFeedback({ type, text })
  }, [])

  useEffect(() => {
    if (!feedback) return
    const timer = setTimeout(() => setFeedback(null), 4000)
    return () => clearTimeout(timer)
  }, [feedback])

  useEffect(() => {
    setMaintenanceActive(maintenanceMode.isActive)
    setMaintenanceMessage(maintenanceMode.message)
  }, [maintenanceMode.isActive, maintenanceMode.message])

  useEffect(() => {
    return () => {
      if (pollTimerRef.current) clearInterval(pollTimerRef.current)
    }
  }, [])

  const fetchSchedulerStatus = useCallback(async () => {
    try {
      const res = await apiClient.get('/system/scheduler/status')
      setSchedulerEnabled(res.data?.data?.enabled ?? true)
    } catch (err) {
      console.error('Gagal memuat status global scheduler:', err)
    }
  }, [])

  const fetchSchedulerJobs = useCallback(
    async (silent = false) => {
      if (!silent) setLoadingJobs(true)
      try {
        const res = await apiClient.get('/system/scheduler/jobs')
        const list = res.data?.data ?? res.data ?? []
        setJobs(Array.isArray(list) ? list : [])
      } catch (err) {
        console.error('Gagal mengambil data scheduler:', err)
        if (!silent) showFeedback('Gagal memuat daftar scheduler.', 'error')
      } finally {
        if (!silent) setLoadingJobs(false)
      }
    },
    [showFeedback]
  )

  const fetchChatTemplates = useCallback(async () => {
    setLoadingTemplates(true)
    try {
      const res = await apiClient.get('/chat/templates')
      const list = res.data?.data ?? []
      setChatTemplates(Array.isArray(list) ? list : [])
    } catch (err) {
      console.error('Gagal memuat template chat:', err)
      showFeedback('Gagal memuat template Live Chat.', 'error')
    } finally {
      setLoadingTemplates(false)
    }
  }, [showFeedback])

  const fetchLocations = useCallback(async () => {
    setLoadingLocations(true)
    try {
      const list = await fetchAllLocations()
      setLocations(list)
    } finally {
      setLoadingLocations(false)
    }
  }, [])

  const refreshAll = useCallback(async () => {
    await Promise.all([
      fetchMaintenanceStatus(),
      fetchSchedulerStatus(),
      fetchSchedulerJobs(),
      fetchChatTemplates(),
      fetchLocations(),
    ])
  }, [fetchMaintenanceStatus, fetchSchedulerStatus, fetchSchedulerJobs, fetchChatTemplates, fetchLocations])

  useEffect(() => {
    void refreshAll()
  }, [refreshAll])

  // Maintenance Handlers
  const saveMaintenanceSettings = async () => {
    setSavingMaintenance(true)
    try {
      const valueToSave = `${maintenanceActive}|${maintenanceMessage}`
      await apiClient.post('/system/settings', {
        key: 'maintenance_mode',
        value: valueToSave,
      })
      await fetchMaintenanceStatus()
      showFeedback('Pengaturan mode maintenance berhasil disimpan.', 'success')
    } catch (err) {
      console.error('Gagal menyimpan pengaturan maintenance:', err)
      showFeedback('Gagal menyimpan pengaturan maintenance.', 'error')
    } finally {
      setSavingMaintenance(false)
    }
  }

  // Scheduler Handlers
  const handleToggleGlobalScheduler = async (nextEnabled: boolean) => {
    const prev = schedulerEnabled
    setSchedulerEnabled(nextEnabled)
    setTogglingGlobal(true)
    try {
      await apiClient.post('/system/scheduler/toggle', { enabled: nextEnabled })
      showFeedback(
        `Scheduler global berhasil ${nextEnabled ? 'diaktifkan' : 'dimatikan'}.`,
        'success'
      )
      await fetchSchedulerJobs(true)
    } catch (err) {
      console.error('Gagal mengubah status global scheduler:', err)
      setSchedulerEnabled(prev)
      showFeedback('Gagal mengubah status global scheduler.', 'error')
    } finally {
      setTogglingGlobal(false)
    }
  }

  const updateJobField = (key: string, patch: Partial<SchedulerJob>) => {
    setJobs((prev) => prev.map((j) => (j.key === key ? { ...j, ...patch } : j)))
  }

  const saveSchedulerSettings = async () => {
    setSavingScheduler(true)
    try {
      for (const job of jobs) {
        await apiClient.post(`/system/scheduler/jobs/${job.key}`, {
          schedule: job.schedule,
          enabled: job.enabled,
        })
      }
      showFeedback('Pengaturan jadwal scheduler berhasil disimpan.', 'success')
      await fetchSchedulerJobs(true)
    } catch (err) {
      console.error('Gagal menyimpan pengaturan scheduler:', err)
      showFeedback('Gagal menyimpan pengaturan scheduler.', 'error')
    } finally {
      setSavingScheduler(false)
    }
  }

  const runJobNow = async (jobKey: string) => {
    setTriggeringJob(jobKey)
    try {
      const res = await apiClient.post(`/system/scheduler/jobs/${jobKey}/run`)
      const msg = res.data?.message || 'Tugas berhasil dipicu di background.'
      showFeedback(msg, 'success')

      setSuccessJobs((prev) => ({ ...prev, [jobKey]: true }))
      setTimeout(() => {
        setSuccessJobs((prev) => ({ ...prev, [jobKey]: false }))
      }, 3000)

      updateJobField(jobKey, { is_running: true })
      await fetchSchedulerJobs(true)

      if (pollTimerRef.current) clearInterval(pollTimerRef.current)
      let pollCount = 0
      pollTimerRef.current = setInterval(() => {
        pollCount += 1
        void fetchSchedulerJobs(true)
        if (pollCount >= 5 && pollTimerRef.current) {
          clearInterval(pollTimerRef.current)
          pollTimerRef.current = null
        }
      }, 2000)
    } catch (err: any) {
      console.error('Gagal memicu tugas scheduler:', err)
      const errMsg = err?.response?.data?.error || 'Gagal memicu tugas scheduler.'
      showFeedback(errMsg, 'error')
    } finally {
      setTriggeringJob('')
    }
  }

  // Chat Template Handlers
  const openAddTemplate = () => {
    setEditingTemplateId(null)
    setTemplateErrors({})
    setTemplateForm({
      shortcut: '',
      title: '',
      content: '',
      icon: 'mdi-message-text-outline',
      sort_order: chatTemplates.length + 1,
    })
    setTemplateDialogOpen(true)
  }

  const openEditTemplate = (tpl: ChatTemplate) => {
    setEditingTemplateId(tpl.id)
    setTemplateErrors({})
    setTemplateForm({
      shortcut: tpl.shortcut,
      title: tpl.title,
      content: tpl.content,
      icon: tpl.icon || 'mdi-message-text-outline',
      sort_order: tpl.sort_order || 0,
    })
    setTemplateDialogOpen(true)
  }

  const saveChatTemplate = async () => {
    const errors: { shortcut?: string; content?: string } = {}
    if (!templateForm.shortcut.trim()) errors.shortcut = 'Shortcut wajib diisi'
    if (!templateForm.content.trim()) errors.content = 'Isi pesan tidak boleh kosong'
    setTemplateErrors(errors)
    if (Object.keys(errors).length > 0) {
      showFeedback('Shortcut dan Isi Pesan wajib diisi.', 'error')
      return
    }

    setSavingTemplate(true)
    try {
      const cleanShortcut = templateForm.shortcut.replace(/^\//, '').trim().toLowerCase()
      const payload = {
        shortcut: cleanShortcut,
        title: templateForm.title.trim() || cleanShortcut,
        content: templateForm.content.trim(),
        icon: templateForm.icon.trim() || 'mdi-message-text-outline',
        sort_order: templateForm.sort_order || 0,
      }
      if (editingTemplateId) {
        await apiClient.put(`/chat/templates/${editingTemplateId}`, payload)
        showFeedback('Template chat berhasil diperbarui.', 'success')
      } else {
        await apiClient.post('/chat/templates', payload)
        showFeedback('Template chat berhasil ditambahkan.', 'success')
      }
      setTemplateDialogOpen(false)
      await fetchChatTemplates()
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Terjadi kesalahan'
      showFeedback(`Gagal menyimpan template: ${msg}`, 'error')
    } finally {
      setSavingTemplate(false)
    }
  }

  const confirmDeleteTemplate = async () => {
    if (!templateToDelete) return
    setDeletingTemplate(true)
    try {
      await apiClient.delete(`/chat/templates/${templateToDelete.id}`)
      showFeedback('Template berhasil dihapus.', 'success')
      setTemplateToDelete(null)
      await fetchChatTemplates()
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Terjadi kesalahan'
      showFeedback(`Gagal menghapus template: ${msg}`, 'error')
    } finally {
      setDeletingTemplate(false)
    }
  }

  // Master Locations Handlers
  const handleAddLocation = () => {
    const trimmed = newLocationName.trim()
    if (!trimmed) {
      showFeedback('Nama lokasi tidak boleh kosong.', 'error')
      return
    }
    if (locations.some((l) => l.name.toLowerCase() === trimmed.toLowerCase())) {
      showFeedback(`Lokasi "${trimmed}" sudah ada di daftar.`, 'error')
      return
    }
    setLocations((prev) => [...prev, { name: trimmed, is_active: true }])
    setNewLocationName('')
  }

  const toggleLocationActive = (name: string) => {
    setLocations((prev) =>
      prev.map((l) => (l.name === name ? { ...l, is_active: !l.is_active } : l))
    )
  }

  const removeLocation = (name: string) => {
    setLocations((prev) => prev.filter((l) => l.name !== name))
  }

  const saveLocations = async () => {
    if (locations.length === 0) {
      showFeedback('Minimal harus ada 1 lokasi di daftar.', 'error')
      return
    }
    setSavingLocations(true)
    try {
      await apiClient.post('/system/settings', {
        key: LOCATIONS_SETTING_KEY,
        value: JSON.stringify(locations),
      })
      showFeedback('Daftar lokasi pelanggan berhasil disimpan.', 'success')
    } catch (err) {
      console.error('Gagal menyimpan daftar lokasi:', err)
      showFeedback('Gagal menyimpan daftar lokasi.', 'error')
    } finally {
      setSavingLocations(false)
    }
  }

  const activeJobsCount = jobs.filter((j) => j.enabled).length

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="h-11 w-11 rounded-xl bg-primary text-primary-foreground flex items-center justify-center shrink-0">
            <Settings className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground">Pengaturan Sistem</h1>
            <p className="text-xs text-muted-foreground mt-0.5">
              Kelola mode maintenance, penjadwalan cronjob, dan template balasan Live Chat
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => void refreshAll()}
          disabled={loadingJobs || loadingTemplates}
        >
          <RefreshCw
            className={`h-4 w-4 mr-1.5 ${loadingJobs || loadingTemplates ? 'animate-spin' : ''}`}
          />
          Segarkan
        </Button>
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

      {/* Main Operational Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={(v) => setActiveTab(v as 'maintenance' | 'scheduler' | 'templates' | 'locations')}
        className="space-y-4"
      >
        <TabsList className="grid w-full sm:w-auto sm:inline-grid grid-cols-2 sm:grid-cols-4 h-auto p-1">
          <TabsTrigger value="maintenance" className="gap-2 py-2 px-3.5 text-xs sm:text-sm">
            <Wrench className="h-4 w-4 shrink-0" />
            <span className="truncate">Maintenance</span>
            {maintenanceActive && (
              <span className="h-2 w-2 rounded-full bg-amber-500 shrink-0" title="Aktif" />
            )}
          </TabsTrigger>
          <TabsTrigger value="scheduler" className="gap-2 py-2 px-3.5 text-xs sm:text-sm">
            <Clock className="h-4 w-4 shrink-0" />
            <span className="truncate">Scheduler &amp; Cron</span>
            <Badge variant="secondary" className="hidden md:inline-flex text-[10px] px-1.5 py-0">
              {activeJobsCount}/{jobs.length}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="templates" className="gap-2 py-2 px-3.5 text-xs sm:text-sm">
            <Zap className="h-4 w-4 shrink-0" />
            <span className="truncate">Template Chat</span>
            <Badge variant="secondary" className="hidden md:inline-flex text-[10px] px-1.5 py-0">
              {chatTemplates.length}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="locations" className="gap-2 py-2 px-3.5 text-xs sm:text-sm">
            <MapPin className="h-4 w-4 shrink-0" />
            <span className="truncate">Lokasi / Area</span>
            <Badge variant="secondary" className="hidden md:inline-flex text-[10px] px-1.5 py-0">
              {locations.filter((l) => l.is_active).length}
            </Badge>
          </TabsTrigger>
        </TabsList>

        {/* ================= TAB 1: MAINTENANCE MODE ================= */}
        <TabsContent value="maintenance" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-start">
            <Card className="lg:col-span-7 py-0 overflow-hidden">
              <CardHeader className="px-5 py-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <CardTitle className="text-base font-bold">Mode Maintenance</CardTitle>
                    <CardDescription className="text-xs mt-0.5">
                      Kontrol status banner pemberitahuan sistem secara global
                    </CardDescription>
                  </div>
                  <Badge
                    variant="outline"
                    className={
                      maintenanceActive
                        ? 'bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border-amber-300 dark:border-amber-800'
                        : 'text-muted-foreground'
                    }
                  >
                    {maintenanceActive ? 'Aktif' : 'Nonaktif'}
                  </Badge>
                </div>
              </CardHeader>
              <Separator />
              <CardContent className="p-5 space-y-5">
                <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/40 p-4">
                  <div className="space-y-0.5">
                    <Label htmlFor="maintenance-switch" className="text-sm font-semibold cursor-pointer">
                      Aktifkan Mode Maintenance
                    </Label>
                    <p className="text-xs text-muted-foreground">
                      Menampilkan banner pemberitahuan di seluruh portal saat update atau perbaikan.
                    </p>
                  </div>
                  <Switch
                    id="maintenance-switch"
                    checked={maintenanceActive}
                    onCheckedChange={setMaintenanceActive}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="maintenance-message" className="text-sm font-medium">
                    Pesan Maintenance
                  </Label>
                  <Textarea
                    id="maintenance-message"
                    rows={3}
                    value={maintenanceMessage}
                    onChange={(e) => setMaintenanceMessage(e.target.value)}
                    disabled={!maintenanceActive}
                    placeholder="Contoh: Sistem akan di-update pukul 23:00 WIB."
                  />
                  <p className="text-xs text-muted-foreground">
                    Pesan ini akan tampil langsung kepada pengguna selama mode maintenance aktif.
                  </p>
                </div>

                <div className="flex justify-end pt-2 border-t">
                  <Button onClick={saveMaintenanceSettings} disabled={savingMaintenance}>
                    {savingMaintenance ? (
                      <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4 mr-1.5" />
                    )}
                    Simpan Pengaturan
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Live Preview Card */}
            <Card className="lg:col-span-5 py-0 overflow-hidden">
              <CardHeader className="px-5 py-4">
                <CardTitle className="text-sm font-semibold">Pratinjau Banner</CardTitle>
                <CardDescription className="text-xs">
                  Tampilan aktual yang dilihat pengguna
                </CardDescription>
              </CardHeader>
              <Separator />
              <CardContent className="p-5">
                {maintenanceActive ? (
                  <div className="rounded-xl border border-amber-300 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/40 p-4 flex items-start gap-3">
                    <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
                    <div className="space-y-1 min-w-0">
                      <div className="text-xs font-bold uppercase tracking-wide text-amber-800 dark:text-amber-300">
                        Pemberitahuan Maintenance
                      </div>
                      <p className="text-sm text-amber-900 dark:text-amber-100 break-words">
                        {maintenanceMessage.trim() || 'Sistem sedang dalam tahap maintenance.'}
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="rounded-xl border border-dashed p-6 text-center space-y-1.5">
                    <CheckCircle2 className="h-6 w-6 text-emerald-600 dark:text-emerald-400 mx-auto" />
                    <div className="text-sm font-medium text-foreground">Sistem Berjalan Normal</div>
                    <p className="text-xs text-muted-foreground">
                      Banner maintenance tidak ditampilkan ke pengguna.
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* ================= TAB 2: SCHEDULER & CRONJOB ================= */}
        <TabsContent value="scheduler" className="space-y-4">
          {/* Global Kill-Switch */}
          <Card
            className={`py-0 overflow-hidden border ${
              schedulerEnabled
                ? 'border-emerald-200 dark:border-emerald-900/60'
                : 'border-red-200 dark:border-red-900/60'
            }`}
          >
            <CardContent
              className={`p-4 md:p-5 flex flex-col md:flex-row md:items-center justify-between gap-4 ${
                schedulerEnabled
                  ? 'bg-emerald-50/50 dark:bg-emerald-950/20'
                  : 'bg-red-50/50 dark:bg-red-950/20'
              }`}
            >
              <div className="flex items-start gap-3">
                {schedulerEnabled ? (
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
                ) : (
                  <PauseCircle className="h-5 w-5 text-red-600 dark:text-red-400 shrink-0 mt-0.5" />
                )}
                <div>
                  <div className="text-sm font-bold text-foreground">
                    Status Scheduler Global:{' '}
                    <span
                      className={
                        schedulerEnabled
                          ? 'text-emerald-700 dark:text-emerald-400'
                          : 'text-red-700 dark:text-red-400'
                      }
                    >
                      {schedulerEnabled ? 'AKTIF' : 'BERHENTI (PAUSED)'}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Sakelar ini mematikan seluruh tugas otomatis secara instan. Disarankan dimatikan
                    saat melakukan import data massal.
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3 shrink-0">
                {togglingGlobal && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                <Switch
                  checked={schedulerEnabled}
                  disabled={togglingGlobal}
                  onCheckedChange={(checked) => void handleToggleGlobalScheduler(checked)}
                  aria-label="Toggle Scheduler Global"
                />
                <span className="text-xs font-medium text-foreground">
                  {schedulerEnabled ? 'Scheduler Aktif' : 'Scheduler Berhenti'}
                </span>
              </div>
            </CardContent>
          </Card>

          {/* Jobs List Card */}
          <Card className="py-0 overflow-hidden">
            <CardHeader className="px-5 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <CardTitle className="text-base font-bold">Daftar Tugas Terjadwal (Cronjob)</CardTitle>
                <CardDescription className="text-xs mt-0.5">
                  Atur ekspresi cron per tugas atau picu eksekusi manual di background
                </CardDescription>
              </div>
              <Button
                size="sm"
                onClick={saveSchedulerSettings}
                disabled={savingScheduler || loadingJobs || jobs.length === 0}
              >
                {savingScheduler ? (
                  <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-1.5" />
                )}
                Simpan Jadwal
              </Button>
            </CardHeader>
            <Separator />

            <CardContent className="p-0">
              {loadingJobs ? (
                <div className="p-5 space-y-3">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Skeleton key={i} className="h-20 w-full rounded-lg" />
                  ))}
                </div>
              ) : jobs.length === 0 ? (
                <div className="py-12 text-center text-sm text-muted-foreground">
                  Tidak ada tugas scheduler yang terdaftar.
                </div>
              ) : (
                <div className="divide-y divide-border">
                  {jobs.map((job) => {
                    const isTriggering = triggeringJob === job.key
                    const isSuccess = Boolean(successJobs[job.key])
                    return (
                      <div
                        key={job.key}
                        className="p-4 md:px-5 md:py-4 flex flex-col lg:flex-row lg:items-center justify-between gap-4 hover:bg-muted/30 transition-colors"
                      >
                        {/* Job Info */}
                        <div className="space-y-1.5 min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="text-sm font-semibold text-foreground">{job.name}</span>
                            <Badge
                              variant="outline"
                              className={
                                job.is_running
                                  ? 'bg-amber-50 dark:bg-amber-950/50 text-amber-700 dark:text-amber-300 border-amber-300 dark:border-amber-800 text-[11px]'
                                  : job.enabled
                                    ? 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border-emerald-300 dark:border-emerald-800 text-[11px]'
                                    : 'text-muted-foreground text-[11px]'
                              }
                            >
                              {job.is_running ? 'Running' : job.enabled ? 'Aktif' : 'Nonaktif'}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground">{job.description}</p>
                          <div className="flex flex-wrap items-center gap-x-5 gap-y-1 text-[11px] text-muted-foreground pt-0.5">
                            <span className="inline-flex items-center gap-1">
                              <History className="h-3.5 w-3.5" />
                              Terakhir: <strong className="font-medium text-foreground/90">{formatDateTime(job.last_run)}</strong>
                            </span>
                            <span className="inline-flex items-center gap-1">
                              <Clock className="h-3.5 w-3.5" />
                              Berikutnya: <strong className="font-medium text-foreground/90">{formatDateTime(job.next_run)}</strong>
                            </span>
                          </div>
                        </div>

                        {/* Controls */}
                        <div className="flex flex-wrap sm:flex-nowrap items-center gap-3 lg:w-[420px] shrink-0">
                          <Switch
                            checked={job.enabled}
                            onCheckedChange={(checked) => updateJobField(job.key, { enabled: checked })}
                            aria-label={`Aktifkan ${job.name}`}
                          />
                          <div className="flex-1 min-w-[160px]">
                            <Input
                              value={job.schedule}
                              disabled={!job.enabled}
                              onChange={(e) => updateJobField(job.key, { schedule: e.target.value })}
                              placeholder="0 12 * * *"
                              className="h-9 font-mono text-xs"
                              aria-label={`Cron expression untuk ${job.name}`}
                            />
                          </div>
                          <Button
                            variant={job.is_running || isSuccess ? 'default' : 'outline'}
                            size="sm"
                            className="h-9 min-w-[108px]"
                            disabled={job.is_running || isSuccess || isTriggering}
                            onClick={() => void runJobNow(job.key)}
                          >
                            {isTriggering || job.is_running ? (
                              <>
                                <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                                Running...
                              </>
                            ) : isSuccess ? (
                              <>
                                <CheckCircle2 className="h-3.5 w-3.5 mr-1.5" />
                                Dipicu
                              </>
                            ) : (
                              <>
                                <Play className="h-3.5 w-3.5 mr-1.5" />
                                Jalankan
                              </>
                            )}
                          </Button>
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* ================= TAB 3: LIVE CHAT TEMPLATES ================= */}
        <TabsContent value="templates" className="space-y-4">
          <Card className="py-0 overflow-hidden">
            <CardHeader className="px-5 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <CardTitle className="text-base font-bold">Template Pesan Cepat Live Chat</CardTitle>
                <CardDescription className="text-xs mt-0.5">
                  CS dapat mengetik <code className="px-1 py-0.5 rounded bg-muted font-mono">/pintasan</code> di kolom chat untuk mengirim balasan instan.
                </CardDescription>
              </div>
              <Button size="sm" onClick={openAddTemplate}>
                <Plus className="h-4 w-4 mr-1.5" />
                Tambah Template
              </Button>
            </CardHeader>
            <Separator />

            <CardContent className="p-0">
              {loadingTemplates ? (
                <div className="p-5 space-y-3">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Skeleton key={i} className="h-12 w-full rounded-lg" />
                  ))}
                </div>
              ) : chatTemplates.length === 0 ? (
                <div className="py-14 text-center">
                  <MessageSquareText className="h-10 w-10 text-muted-foreground/40 mx-auto mb-2.5" />
                  <div className="text-sm font-medium text-foreground">
                    Belum ada template yang disimpan
                  </div>
                  <p className="text-xs text-muted-foreground mt-1 mb-4">
                    Buat template balasan cepat pertama untuk membantu CS merespons lebih cepat.
                  </p>
                  <Button size="sm" onClick={openAddTemplate}>
                    <Plus className="h-4 w-4 mr-1.5" />
                    Tambah Template
                  </Button>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-muted/50">
                        <TableHead className="px-5 py-3 text-xs font-semibold w-[150px]">
                          Pintasan (/)
                        </TableHead>
                        <TableHead className="px-4 py-3 text-xs font-semibold w-[200px]">
                          Judul
                        </TableHead>
                        <TableHead className="px-4 py-3 text-xs font-semibold">Isi Pesan</TableHead>
                        <TableHead className="px-5 py-3 text-xs font-semibold text-center w-[110px]">
                          Aksi
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {chatTemplates.map((tpl) => (
                        <TableRow key={tpl.id} className="hover:bg-muted/40">
                          <TableCell className="px-5 py-3.5 align-top">
                            <Badge
                              variant="outline"
                              className="font-mono text-xs bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800"
                            >
                              /{tpl.shortcut}
                            </Badge>
                          </TableCell>
                          <TableCell className="px-4 py-3.5 align-top">
                            <div className="text-sm font-medium text-foreground">{tpl.title}</div>
                          </TableCell>
                          <TableCell className="px-4 py-3.5 align-top">
                            <p className="text-xs text-muted-foreground whitespace-pre-wrap line-clamp-2 max-w-xl">
                              {tpl.content}
                            </p>
                          </TableCell>
                          <TableCell className="px-5 py-3.5 align-top">
                            <div className="flex items-center justify-center gap-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-950/50"
                                onClick={() => openEditTemplate(tpl)}
                                title="Edit Template"
                              >
                                <Pencil className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50"
                                onClick={() => setTemplateToDelete(tpl)}
                                title="Hapus Template"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* ================= TAB 4: MASTER LOKASI / AREA ================= */}
        <TabsContent value="locations" className="space-y-4">
          <Card className="py-0 overflow-hidden">
            <CardHeader className="px-5 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <CardTitle className="text-base font-bold">Daftar Lokasi / Area Pelanggan</CardTitle>
                <CardDescription className="text-xs mt-0.5">
                  Lokasi yang aktif akan langsung muncul di pilihan alamat pada form &amp; filter Pelanggan.
                </CardDescription>
              </div>
              <Button
                size="sm"
                onClick={saveLocations}
                disabled={savingLocations || loadingLocations}
              >
                {savingLocations ? (
                  <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-1.5" />
                )}
                Simpan Lokasi
              </Button>
            </CardHeader>
            <Separator />

            <CardContent className="p-5 space-y-4">
              {/* Add New Location */}
              <div className="flex flex-col sm:flex-row gap-2.5">
                <Input
                  value={newLocationName}
                  onChange={(e) => setNewLocationName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault()
                      handleAddLocation()
                    }
                  }}
                  placeholder="Tambah nama lokasi/cluster baru (contoh: Rusun Rorotan)"
                  className="flex-1"
                />
                <Button type="button" variant="outline" onClick={handleAddLocation}>
                  <Plus className="h-4 w-4 mr-1.5" />
                  Tambah ke Daftar
                </Button>
              </div>

              <div className="rounded-lg border border-amber-200 dark:border-amber-900/50 bg-amber-50/60 dark:bg-amber-950/30 px-3.5 py-2.5 text-xs text-amber-900 dark:text-amber-200 flex items-start gap-2">
                <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
                <span>
                  Disarankan menonaktifkan (toggle Off) daripada menghapus lokasi lama agar data
                  riwayat pelanggan dan notifikasi wilayah (seperti Waringin &amp; Parama) tetap
                  konsisten. Jangan lupa klik <strong>Simpan Lokasi</strong> setelah mengubah daftar.
                </span>
              </div>

              {loadingLocations ? (
                <div className="space-y-2">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Skeleton key={i} className="h-11 w-full rounded-lg" />
                  ))}
                </div>
              ) : (
                <div className="border rounded-xl overflow-hidden divide-y divide-border">
                  {locations.map((loc) => (
                    <div
                      key={loc.name}
                      className="px-4 py-2.5 flex items-center justify-between gap-3 hover:bg-muted/30"
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <MapPin className="h-4 w-4 text-muted-foreground shrink-0" />
                        <span
                          className={`text-sm font-medium truncate ${
                            loc.is_active ? 'text-foreground' : 'text-muted-foreground line-through'
                          }`}
                        >
                          {loc.name}
                        </span>
                        <Badge
                          variant="outline"
                          className={
                            loc.is_active
                              ? 'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800 text-[10px]'
                              : 'text-muted-foreground text-[10px]'
                          }
                        >
                          {loc.is_active ? 'Aktif' : 'Nonaktif'}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <Switch
                          checked={loc.is_active}
                          onCheckedChange={() => toggleLocationActive(loc.name)}
                          aria-label={`Aktifkan lokasi ${loc.name}`}
                        />
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50"
                          onClick={() => removeLocation(loc.name)}
                          title="Hapus dari daftar"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Dialog: Add / Edit Chat Template */}
      <Dialog open={templateDialogOpen} onOpenChange={setTemplateDialogOpen}>
        <DialogContent className="sm:max-w-[520px]">
          <DialogHeader>
            <DialogTitle>
              {editingTemplateId ? 'Edit Template Chat' : 'Tambah Template Chat Baru'}
            </DialogTitle>
            <DialogDescription>
              Atur kata kunci pintasan slash (/) dan isi pesan balasan untuk CS.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label htmlFor="tpl-shortcut">Pintasan (Shortcut)</Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-mono text-muted-foreground">
                  /
                </span>
                <Input
                  id="tpl-shortcut"
                  value={templateForm.shortcut}
                  onChange={(e) => {
                    setTemplateForm((p) => ({ ...p, shortcut: e.target.value }))
                    if (templateErrors.shortcut) {
                      setTemplateErrors((p) => ({ ...p, shortcut: undefined }))
                    }
                  }}
                  placeholder="contoh: salam, cekteknis, promo"
                  className="pl-7 font-mono"
                />
              </div>
              {templateErrors.shortcut ? (
                <p className="text-xs text-destructive">{templateErrors.shortcut}</p>
              ) : (
                <p className="text-[11px] text-muted-foreground">
                  Ketikkan kata ini setelah tanda slash (/) di kolom chat.
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="tpl-title">Judul Template</Label>
              <Input
                id="tpl-title"
                value={templateForm.title}
                onChange={(e) => setTemplateForm((p) => ({ ...p, title: e.target.value }))}
                placeholder="contoh: Salam Pembuka"
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="tpl-content">Isi Pesan Balasan</Label>
              <Textarea
                id="tpl-content"
                rows={4}
                value={templateForm.content}
                onChange={(e) => {
                  setTemplateForm((p) => ({ ...p, content: e.target.value }))
                  if (templateErrors.content) {
                    setTemplateErrors((p) => ({ ...p, content: undefined }))
                  }
                }}
                placeholder="Tuliskan template balasan CS di sini..."
              />
              {templateErrors.content && (
                <p className="text-xs text-destructive">{templateErrors.content}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="tpl-icon">Ikon MDI (Opsional)</Label>
              <Input
                id="tpl-icon"
                value={templateForm.icon}
                onChange={(e) => setTemplateForm((p) => ({ ...p, icon: e.target.value }))}
                placeholder="mdi-message-text-outline"
              />
              <p className="text-[11px] text-muted-foreground">
                Nama ikon Material Design Icons, misal: mdi-hand-wave-outline
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setTemplateDialogOpen(false)}
              disabled={savingTemplate}
            >
              Batal
            </Button>
            <Button onClick={saveChatTemplate} disabled={savingTemplate}>
              {savingTemplate ? (
                <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
              ) : (
                <Save className="h-4 w-4 mr-1.5" />
              )}
              Simpan Template
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog: Delete Chat Template Confirmation */}
      <Dialog
        open={Boolean(templateToDelete)}
        onOpenChange={(open) => !open && setTemplateToDelete(null)}
      >
        <DialogContent className="sm:max-w-[440px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              Hapus Template Chat
            </DialogTitle>
            <DialogDescription>Tindakan ini tidak dapat dibatalkan.</DialogDescription>
          </DialogHeader>
          {templateToDelete && (
            <div className="rounded-lg border border-red-200 dark:border-red-900/60 bg-red-50 dark:bg-red-950/40 p-3 text-sm space-y-1">
              <div className="font-mono font-semibold text-foreground">
                /{templateToDelete.shortcut}
              </div>
              <div className="text-xs text-muted-foreground">{templateToDelete.title}</div>
            </div>
          )}
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setTemplateToDelete(null)}
              disabled={deletingTemplate}
            >
              Batal
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDeleteTemplate}
              disabled={deletingTemplate}
            >
              {deletingTemplate ? (
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
