import { useState, useEffect, useCallback, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import apiClient from '@/services/api'
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
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  PieChart,
  Pie,
  Cell,
  Legend,
} from 'recharts'
import {
  BarChart3,
  RefreshCw,
  FileDown,
  CalendarRange,
  FilterX,
  TicketCheck,
  CheckCircle2,
  Clock,
  AlertCircle,
  TrendingUp,
  PieChart as PieChartIcon,
  Tag,
  Users,
  Timer,
  AlertTriangle,
  Search,
  Phone,
  MapPin,
  FileSpreadsheet,
  FileText,
  FileCode2,
  Printer,
  ChevronRight,
  Loader2,
  UserCheck,
  ChevronLeft,
} from 'lucide-react'

interface Statistics {
  total_tickets: number
  open_tickets: number
  in_progress_tickets: number
  resolved_tickets: number
  closed_tickets: number
  high_priority_tickets?: number
  critical_priority_tickets?: number
  avg_resolution_time_hours: number | null
  tickets_this_month?: number
  unresolved_over_24h?: number
}

interface CategoryPerformance {
  category: string
  category_display: string
  total_tickets: number
  resolved_tickets: number
  resolution_rate_percent: number
  avg_resolution_hours: number
  avg_downtime_minutes: number
  total_downtime_minutes: number
}

interface UserPerformance {
  user_id: number
  user_name: string
  total_assigned: number
  resolved_tickets: number
  resolution_rate_percent: number
  avg_resolution_hours: number
}

interface DowntimeAnalysis {
  overall_statistics: {
    total_downtime_minutes: number
    total_downtime_hours: number
    total_downtime_days: number
    avg_downtime_minutes: number
    avg_downtime_hours: number
    max_downtime_minutes: number
    max_downtime_hours: number
    tickets_with_downtime: number
  }
  top_customers: Array<{
    customer_id: number
    customer_name: string
    total_downtime_minutes: number
    ticket_count: number
    avg_downtime_minutes: number
    total_downtime_hours: number
    avg_downtime_hours: number
  }>
}

interface MonthlyTrend {
  month: string
  month_name?: string
  statistics: {
    total: number
    resolved: number
    avg_resolution_hours: number
  }
}

interface TicketDetailItem {
  id: number
  ticket_number: string
  title: string
  description: string
  status: string
  priority: string
  category: string
  customer_name: string
  customer_address: string
  customer_phone: string
  customer_brand: string
  id_pelanggan: string
  ip_pelanggan: string
  onu_power: string
  created_at: string
  resolved_at?: string | null
  total_downtime_minutes?: number
  downtime_hours: string
  assigned_to: string
  latest_actions: Array<{
    action_description: string
    summary_problem: string
    summary_action: string
    action_date: string
    taken_by: string
  }>
}

const PIE_COLORS = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
  '#EC4899', '#14B8A6', '#F97316', '#06B6D4', '#84CC16',
]

const ticketStatusOptions = [
  { label: 'Open', value: 'open' },
  { label: 'In Progress', value: 'in_progress' },
  { label: 'Pending Customer', value: 'pending_customer' },
  { label: 'Pending Vendor', value: 'pending_vendor' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' },
  { label: 'Cancelled', value: 'cancelled' },
]

const ticketCategoryOptions = [
  { label: 'No Connection', value: 'no_connection' },
  { label: 'Slow Connection', value: 'slow_connection' },
  { label: 'Intermittent', value: 'intermittent' },
  { label: 'Hardware Issue', value: 'hardware_issue' },
  { label: 'Cable Issue', value: 'cable_issue' },
  { label: 'ONU Issue', value: 'onu_issue' },
  { label: 'OLT Issue', value: 'olt_issue' },
  { label: 'Mikrotik Issue', value: 'mikrotik_issue' },
  { label: 'Other', value: 'other' },
]

const brandOptions = [
  { label: 'JAKINET', value: 'JAKINET' },
  { label: 'JELANTIK', value: 'JELANTIK' },
  { label: 'JELANTIK NAGRAK', value: 'JELANTIK NAGRAK' },
]

function formatHours(hours: number | null | undefined) {
  if (!hours) return '0h'
  if (hours < 1) return `${Math.round(hours * 60)}m`
  if (hours < 24) return `${Math.round(hours)}h`
  return `${Math.floor(hours / 24)}d ${Math.round(hours % 24)}h`
}

function formatStatus(status: string) {
  if (status === 'pending_customer' || status === 'pending_vendor') return 'Pending'
  return status.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function formatCategory(cat: string) {
  return cat.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())
}

function getStatusBadgeClass(status: string) {
  const map: Record<string, string> = {
    open: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800',
    in_progress: 'bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800',
    pending_customer: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-950/50 dark:text-orange-300 dark:border-orange-800',
    pending_vendor: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-950/50 dark:text-orange-300 dark:border-orange-800',
    resolved: 'bg-emerald-100 text-emerald-800 border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800',
    closed: 'bg-muted text-foreground/90 border-border',
    cancelled: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-950/50 dark:text-red-300 dark:border-red-800',
  }
  return map[status] || 'bg-muted text-foreground/90 border-border'
}

function getDowntimeBadgeClass(minutes: number | null | undefined) {
  if (!minutes || minutes <= 0) return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-300'
  if (minutes <= 60) return 'bg-amber-100 text-amber-800 dark:bg-amber-950/50 dark:text-amber-300'
  if (minutes <= 360) return 'bg-red-100 text-red-800 dark:bg-red-950/50 dark:text-red-300'
  return 'bg-purple-100 text-purple-800 dark:bg-purple-950/50 dark:text-purple-300'
}

function getProgressBarColor(rate: number) {
  if (rate >= 90) return 'bg-emerald-500'
  if (rate >= 70) return 'bg-amber-500'
  return 'bg-red-500'
}

function escapeCsvCell(val: unknown) {
  const str = String(val ?? '')
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`
  }
  return str
}

export function TroubleTicketReportView() {
  const navigate = useNavigate()

  const [loading, setLoading] = useState(false)
  const [loadingMonthly, setLoadingMonthly] = useState(false)
  const [loadingCategory, setLoadingCategory] = useState(false)
  const [loadingUser, setLoadingUser] = useState(false)
  const [loadingDowntime, setLoadingDowntime] = useState(false)
  const [loadingDetails, setLoadingDetails] = useState(false)

  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [trendPeriod, setTrendPeriod] = useState<'6' | '12' | '24'>('12')

  const [statistics, setStatistics] = useState<Statistics | null>(null)
  const [monthlyTrends, setMonthlyTrends] = useState<MonthlyTrend[]>([])
  const [categoryPerformance, setCategoryPerformance] = useState<CategoryPerformance[]>([])
  const [userPerformance, setUserPerformance] = useState<UserPerformance[]>([])
  const [downtimeAnalysis, setDowntimeAnalysis] = useState<DowntimeAnalysis | null>(null)
  const [ticketDetails, setTicketDetails] = useState<TicketDetailItem[]>([])

  // Ticket details filters & pagination
  const [ticketSearch, setTicketSearch] = useState('')
  const [ticketStatusFilter, setTicketStatusFilter] = useState('')
  const [ticketCategoryFilter, setTicketCategoryFilter] = useState('')
  const [detailsPage, setDetailsPage] = useState(1)
  const detailsPerPage = 10

  // Export modal
  const [showExportDialog, setShowExportDialog] = useState(false)
  const [exportBrandFilter, setExportBrandFilter] = useState('')
  const [exporting, setExporting] = useState(false)

  const loadStatistics = useCallback(async () => {
    try {
      const r = await apiClient.get('/trouble-tickets/statistics/dashboard')
      setStatistics(r.data)
    } catch (err) {
      console.error('Failed to load statistics:', err)
    }
  }, [])

  const loadMonthlyTrends = useCallback(async () => {
    setLoadingMonthly(true)
    try {
      const params: Record<string, unknown> = { months: parseInt(trendPeriod, 10) }
      if (dateFrom) params.date_from = dateFrom
      if (dateTo) params.date_to = dateTo
      const r = await apiClient.get('/trouble-tickets/reports/monthly-trends', { params })
      setMonthlyTrends(r.data.trends || [])
    } catch {
      setMonthlyTrends([])
    } finally {
      setLoadingMonthly(false)
    }
  }, [trendPeriod, dateFrom, dateTo])

  const loadCategoryPerformance = useCallback(async () => {
    setLoadingCategory(true)
    try {
      const params: Record<string, unknown> = {}
      if (dateFrom) params.date_from = dateFrom
      if (dateTo) params.date_to = dateTo
      const r = await apiClient.get('/trouble-tickets/reports/category-performance', { params })
      setCategoryPerformance(r.data.category_performance || [])
    } catch {
      setCategoryPerformance([])
    } finally {
      setLoadingCategory(false)
    }
  }, [dateFrom, dateTo])

  const loadUserPerformance = useCallback(async () => {
    setLoadingUser(true)
    try {
      const params: Record<string, unknown> = {}
      if (dateFrom) params.date_from = dateFrom
      if (dateTo) params.date_to = dateTo
      const r = await apiClient.get('/trouble-tickets/reports/user-performance', { params })
      setUserPerformance(r.data.user_performance || [])
    } catch {
      setUserPerformance([])
    } finally {
      setLoadingUser(false)
    }
  }, [dateFrom, dateTo])

  const loadDowntimeAnalysis = useCallback(async () => {
    setLoadingDowntime(true)
    try {
      const params: Record<string, unknown> = {}
      if (dateFrom) params.date_from = dateFrom
      if (dateTo) params.date_to = dateTo
      const r = await apiClient.get('/trouble-tickets/reports/downtime-analysis', { params })
      if (r.data && r.data.overall_statistics) {
        setDowntimeAnalysis(r.data)
      } else {
        setDowntimeAnalysis(null)
      }
    } catch {
      setDowntimeAnalysis(null)
    } finally {
      setLoadingDowntime(false)
    }
  }, [dateFrom, dateTo])

  const loadTicketDetails = useCallback(async () => {
    setLoadingDetails(true)
    try {
      const params: Record<string, unknown> = {
        skip: 0,
        limit: 100,
        include_relations: true,
        ...(dateFrom && { date_from: dateFrom }),
        ...(dateTo && { date_to: dateTo }),
      }
      const r = await apiClient.get('/trouble-tickets', { params })
      const rawTickets = r.data.data || []

      const detailed = await Promise.all(
        rawTickets.map(async (t: Record<string, any>) => {
          let latestActions: TicketDetailItem['latest_actions'] = []
          try {
            const actRes = await apiClient.get(`/trouble-tickets/${t.id}/actions`)
            const actions = actRes.data.data || actRes.data || []
            latestActions = actions.slice(0, 3).map((a: Record<string, any>) => ({
              action_description: a.action_description || '',
              summary_problem: a.summary_problem || '',
              summary_action: a.summary_action || '',
              action_date: a.created_at,
              taken_by: a.taken_user?.name || 'System',
            }))
          } catch {
            latestActions = []
          }
          return {
            id: t.id,
            ticket_number: t.ticket_number,
            title: t.title || '-',
            description: t.description || '',
            status: t.status || '',
            priority: t.priority || '',
            category: t.category || '',
            customer_name: t.pelanggan?.nama || 'N/A',
            customer_address: t.pelanggan?.alamat || '',
            customer_phone: t.pelanggan?.no_telp || '-',
            customer_brand: t.pelanggan?.harga_layanan?.brand || '-',
            id_pelanggan: t.data_teknis?.id_pelanggan || '-',
            ip_pelanggan: t.data_teknis?.ip_pelanggan || '-',
            onu_power: t.data_teknis?.onu_power ? String(t.data_teknis.onu_power) : '-',
            created_at: t.created_at,
            resolved_at: t.resolved_at,
            total_downtime_minutes: t.total_downtime_minutes || 0,
            downtime_hours: t.total_downtime_minutes ? (t.total_downtime_minutes / 60).toFixed(2) : '0',
            assigned_to: t.assigned_user?.name || 'Belum ditugaskan',
            latest_actions: latestActions,
          } as TicketDetailItem
        })
      )
      setTicketDetails(detailed)
      setDetailsPage(1)
    } catch (err) {
      console.error('Failed to load ticket details:', err)
      setTicketDetails([])
    } finally {
      setLoadingDetails(false)
    }
  }, [dateFrom, dateTo])

  const refreshAllData = useCallback(async () => {
    setLoading(true)
    try {
      await Promise.all([
        loadStatistics(),
        loadMonthlyTrends(),
        loadCategoryPerformance(),
        loadUserPerformance(),
        loadDowntimeAnalysis(),
        loadTicketDetails(),
      ])
    } finally {
      setLoading(false)
    }
  }, [loadStatistics, loadMonthlyTrends, loadCategoryPerformance, loadUserPerformance, loadDowntimeAnalysis, loadTicketDetails])

  useEffect(() => {
    refreshAllData()
  }, [refreshAllData])

  const resolutionRate = useMemo(() => {
    if (!statistics || !statistics.total_tickets) return 0
    const resolved = (statistics.resolved_tickets || 0) + (statistics.closed_tickets || 0)
    return Math.round((resolved / statistics.total_tickets) * 100)
  }, [statistics])

  const resolutionTimeStatus = useMemo(() => {
    const h = statistics?.avg_resolution_time_hours || 0
    if (h < 24) return 'Excellent'
    if (h < 72) return 'Good'
    return 'Needs Improvement'
  }, [statistics])

  const trendsChartData = useMemo(() => {
    const sorted = [...monthlyTrends].sort((a, b) => new Date(a.month).getTime() - new Date(b.month).getTime())
    return sorted.map((m) => ({
      name: new Date(m.month).toLocaleDateString('id-ID', { month: 'short', year: 'numeric' }),
      created: m.statistics.total,
      resolved: m.statistics.resolved,
    }))
  }, [monthlyTrends])

  const categoryChartData = useMemo(() => {
    return categoryPerformance.map((c) => ({
      name: c.category_display || formatCategory(c.category),
      value: c.total_tickets,
    }))
  }, [categoryPerformance])

  const filteredTicketDetails = useMemo(() => {
    const q = ticketSearch.toLowerCase().trim()
    return ticketDetails.filter((t) => {
      const matchSearch =
        !q ||
        t.ticket_number.toLowerCase().includes(q) ||
        t.customer_name.toLowerCase().includes(q) ||
        t.title.toLowerCase().includes(q) ||
        t.assigned_to.toLowerCase().includes(q) ||
        t.description.toLowerCase().includes(q)
      const matchStatus = !ticketStatusFilter || t.status === ticketStatusFilter
      const matchCategory = !ticketCategoryFilter || t.category === ticketCategoryFilter
      return matchSearch && matchStatus && matchCategory
    })
  }, [ticketDetails, ticketSearch, ticketStatusFilter, ticketCategoryFilter])

  const totalDetailPages = Math.max(1, Math.ceil(filteredTicketDetails.length / detailsPerPage))
  const paginatedTicketDetails = useMemo(() => {
    const start = (detailsPage - 1) * detailsPerPage
    return filteredTicketDetails.slice(start, start + detailsPerPage)
  }, [filteredTicketDetails, detailsPage])

  const clearDateFilter = () => {
    setDateFrom('')
    setDateTo('')
  }

  const clearTicketFilters = () => {
    setTicketSearch('')
    setTicketStatusFilter('')
    setTicketCategoryFilter('')
    setDetailsPage(1)
  }

  // Export Logic
  const handleExport = async (format: 'excel' | 'csv' | 'pdf' | 'json') => {
    setExporting(true)
    try {
      const params: Record<string, unknown> = {
        skip: 0,
        limit: 200,
        include_relations: true,
        ...(dateFrom && { date_from: dateFrom }),
        ...(dateTo && { date_to: dateTo }),
        ...(exportBrandFilter && { brand: exportBrandFilter }),
      }
      const r = await apiClient.get('/trouble-tickets', { params })
      const raw = r.data.data || []

      const rows = raw.map((t: Record<string, any>) => ({
        ticket_number: t.ticket_number || '-',
        title: t.title || '-',
        customer_name: t.pelanggan?.nama || 'N/A',
        brand: t.pelanggan?.harga_layanan?.brand || '-',
        phone: t.pelanggan?.no_telp || '-',
        address: t.pelanggan?.alamat || '-',
        ip_pelanggan: t.data_teknis?.ip_pelanggan || '-',
        status: formatStatus(t.status || ''),
        category: formatCategory(t.category || ''),
        assigned_to: t.assigned_user?.name || 'Unassigned',
        downtime_minutes: t.total_downtime_minutes || 0,
        created_at: t.created_at ? new Date(t.created_at).toLocaleString('id-ID') : '-',
        resolved_at: t.resolved_at ? new Date(t.resolved_at).toLocaleString('id-ID') : '-',
      }))

      const dateTag = new Date().toISOString().slice(0, 10)

      if (format === 'json') {
        const payload = {
          metadata: {
            title: exportBrandFilter ? `Trouble Ticket Report - ${exportBrandFilter}` : 'Trouble Ticket Report',
            generated_at: new Date().toISOString(),
            period: { from: dateFrom || 'All time', to: dateTo || 'Present', brand: exportBrandFilter || 'All Brands' },
          },
          statistics,
          monthly_trends: monthlyTrends,
          category_performance: categoryPerformance,
          user_performance: userPerformance,
          downtime_analysis: downtimeAnalysis,
          tickets: rows,
        }
        downloadBlob(JSON.stringify(payload, null, 2), `trouble-ticket-report-${dateTag}.json`, 'application/json')
      } else if (format === 'csv') {
        const headers = [
          'Ticket Number', 'Title', 'Customer', 'Brand', 'Phone', 'Address',
          'IP Pelanggan', 'Status', 'Category', 'Assigned To', 'Downtime (Menit)', 'Created At', 'Resolved At',
        ]
        const csvLines = [
          headers.join(','),
          ...rows.map((row: Record<string, unknown>) =>
            [
              row.ticket_number, row.title, row.customer_name, row.brand, row.phone, row.address,
              row.ip_pelanggan, row.status, row.category, row.assigned_to, row.downtime_minutes, row.created_at, row.resolved_at,
            ].map(escapeCsvCell).join(',')
          ),
        ]
        downloadBlob('\uFEFF' + csvLines.join('\n'), `trouble-ticket-report-${dateTag}.csv`, 'text/csv;charset=utf-8;')
      } else if (format === 'excel') {
        const tableRows = rows
          .map(
            (row: Record<string, unknown>) =>
              `<tr><td>${row.ticket_number}</td><td>${row.title}</td><td>${row.customer_name}</td><td>${row.brand}</td><td>${row.phone}</td><td>${row.address}</td><td>${row.ip_pelanggan}</td><td>${row.status}</td><td>${row.category}</td><td>${row.assigned_to}</td><td>${row.downtime_minutes}</td><td>${row.created_at}</td><td>${row.resolved_at}</td></tr>`
          )
          .join('')
        const html = `<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel"><head><meta charset="UTF-8"/></head><body>
          <h3>Trouble Ticket Report (${exportBrandFilter || 'All Brands'})</h3>
          <p>Periode: ${dateFrom || 'All time'} s/d ${dateTo || 'Present'}</p>
          <table border="1">
            <thead><tr><th>Ticket #</th><th>Title</th><th>Customer</th><th>Brand</th><th>Phone</th><th>Address</th><th>IP</th><th>Status</th><th>Category</th><th>Assigned To</th><th>Downtime (m)</th><th>Created</th><th>Resolved</th></tr></thead>
            <tbody>${tableRows}</tbody>
          </table></body></html>`
        downloadBlob(html, `trouble-ticket-report-${dateTag}.xls`, 'application/vnd.ms-excel')
      } else if (format === 'pdf') {
        const win = window.open('', '_blank')
        if (win) {
          const tableRows = rows.slice(0, 50)
            .map(
              (row: Record<string, unknown>) =>
                `<tr><td>${row.ticket_number}</td><td>${row.customer_name}</td><td>${row.title}</td><td>${row.status}</td><td>${row.category}</td><td>${row.assigned_to}</td><td>${row.downtime_minutes}m</td></tr>`
            )
            .join('')
          win.document.write(`
            <html><head><title>Trouble Ticket Report</title>
            <style>body{font-family:Arial,sans-serif;padding:24px;font-size:12px}h2{margin-bottom:4px}table{width:100%;border-collapse:collapse;margin-top:16px}th,td{border:1px solid #ddd;padding:6px 8px;text-align:left}th{background:#f1f5f9}</style>
            </head><body>
            <h2>Trouble Ticket Analysis Report (${exportBrandFilter || 'All Brands'})</h2>
            <p>Periode: ${dateFrom || 'All time'} s/d ${dateTo || 'Present'} | Total Tickets: ${statistics?.total_tickets || 0} | Resolution Rate: ${resolutionRate}%</p>
            <table><thead><tr><th>Ticket #</th><th>Customer</th><th>Title</th><th>Status</th><th>Category</th><th>Assigned</th><th>Downtime</th></tr></thead><tbody>${tableRows}</tbody></table>
            </body></html>
          `)
          win.document.close()
          win.focus()
          win.print()
        }
      }

      setShowExportDialog(false)
    } catch (err) {
      console.error('Export failed:', err)
      alert('Gagal mengekspor laporan.')
    } finally {
      setExporting(false)
    }
  }

  function downloadBlob(content: string, filename: string, mime: string) {
    const blob = new Blob([content], { type: mime })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="p-4 md:p-6 space-y-5">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <BarChart3 className="h-6 w-6 text-blue-600" />
            <h1 className="text-xl font-bold text-foreground">Trouble Ticket Reports</h1>
            <Badge variant="secondary" className="bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 text-[11px]">Analytics</Badge>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            Analisis komprehensif, tracking downtime, dan performa penanganan gangguan
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={refreshAllData} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Segarkan Data
          </Button>
          <Button size="sm" className="bg-emerald-600 hover:bg-emerald-700 text-white" onClick={() => setShowExportDialog(true)} disabled={loading || exporting}>
            <FileDown className="h-4 w-4 mr-1.5" />
            Export Laporan
          </Button>
        </div>
      </div>

      {/* Periode Laporan Filter */}
      <Card className="shadow-xs py-0">
        <CardContent className="p-3.5">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-950/50 text-blue-600">
                <CalendarRange className="h-4 w-4" />
              </div>
              <div>
                <div className="text-sm font-semibold text-foreground">Periode Laporan</div>
                <div className="text-xs text-muted-foreground">Filter analitik berdasarkan rentang tanggal</div>
              </div>
              {(dateFrom || dateTo) && (
                <Badge className="bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 hover:bg-blue-50 dark:hover:bg-blue-950/50 text-[11px]">Filter Aktif</Badge>
              )}
            </div>
            <div className="flex flex-wrap items-center gap-2.5">
              <div className="flex items-center gap-1.5">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">Dari</Label>
                <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} className="h-9 w-[150px] text-xs" />
              </div>
              <div className="flex items-center gap-1.5">
                <Label className="text-xs text-muted-foreground whitespace-nowrap">Sampai</Label>
                <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} className="h-9 w-[150px] text-xs" />
              </div>
              {(dateFrom || dateTo) && (
                <Button variant="ghost" size="sm" onClick={clearDateFilter} className="h-9 text-xs text-red-600 hover:text-red-700">
                  <FilterX className="h-3.5 w-3.5 mr-1" />
                  Reset
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Key Performance Metrics */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs border-l-4 border-l-blue-500">
          <div className="p-2 rounded-lg bg-blue-50 dark:bg-blue-950/50 text-blue-600 shrink-0">
            <TicketCheck className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <div className="text-[11px] font-medium text-muted-foreground truncate">Total Tickets</div>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-foreground">{statistics?.total_tickets || 0}</span>
              <span className="text-[10px] text-muted-foreground truncate">Total dibuat</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs border-l-4 border-l-emerald-500">
          <div className="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 shrink-0">
            <CheckCircle2 className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <div className="text-[11px] font-medium text-muted-foreground truncate">Resolution Rate</div>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-emerald-600">{resolutionRate}%</span>
              <span className="text-[10px] text-muted-foreground truncate">Terselesaikan</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs border-l-4 border-l-amber-500">
          <div className="p-2 rounded-lg bg-amber-50 dark:bg-amber-950/50 text-amber-600 shrink-0">
            <Clock className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <div className="text-[11px] font-medium text-muted-foreground truncate">Avg Resolution</div>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-amber-600">{formatHours(statistics?.avg_resolution_time_hours)}</span>
              <span className="text-[10px] text-muted-foreground truncate">{resolutionTimeStatus}</span>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-xl border bg-card px-3.5 py-2.5 shadow-xs border-l-4 border-l-rose-500">
          <div className="p-2 rounded-lg bg-rose-50 dark:bg-rose-950/50 text-rose-600 shrink-0">
            <AlertCircle className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <div className="text-[11px] font-medium text-muted-foreground truncate">Open Tickets</div>
            <div className="flex items-baseline gap-2">
              <span className="text-lg font-bold text-rose-600">{statistics?.open_tickets || 0}</span>
              <span className="text-[10px] text-muted-foreground truncate">Perlu atensi</span>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        {/* Monthly Trends */}
        <Card className="lg:col-span-8 shadow-xs">
          <CardHeader className="pb-2">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-4 w-4 text-blue-600" />
                <div>
                  <CardTitle className="text-sm font-semibold">Monthly Trends</CardTitle>
                  <p className="text-xs text-muted-foreground">Tren volume dan resolusi tiket bulanan</p>
                </div>
              </div>
              <div className="inline-flex rounded-lg border p-0.5 bg-muted/50">
                {(['6', '12', '24'] as const).map((p) => (
                  <button
                    key={p}
                    type="button"
                    onClick={() => setTrendPeriod(p)}
                    className={`px-2.5 py-1 text-xs font-medium rounded-md transition-colors ${
                      trendPeriod === p ? 'bg-card text-blue-600 shadow-xs' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    {p === '6' ? '6 Bulan' : p === '12' ? '1 Tahun' : '2 Tahun'}
                  </button>
                ))}
              </div>
            </div>
          </CardHeader>
          <Separator />
          <CardContent className="p-4">
            {loadingMonthly ? (
              <div className="h-64 flex items-center justify-center text-xs text-muted-foreground">
                <Loader2 className="h-5 w-5 animate-spin mr-2 text-blue-500" />
                Memuat tren bulanan...
              </div>
            ) : trendsChartData.length === 0 ? (
              <div className="h-64 flex flex-col items-center justify-center text-muted-foreground text-xs">
                <TrendingUp className="h-8 w-8 mb-2 opacity-40" />
                Tidak ada data tren pada rentang waktu ini
              </div>
            ) : (
              <div className="h-64 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={trendsChartData} margin={{ top: 10, right: 10, left: -15, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorCreated" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.2} />
                        <stop offset="95%" stopColor="#3B82F6" stopOpacity={0} />
                      </linearGradient>
                      <linearGradient id="colorResolved" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10B981" stopOpacity={0.2} />
                        <stop offset="95%" stopColor="#10B981" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                    <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                    <YAxis allowDecimals={false} tick={{ fontSize: 11 }} />
                    <RechartsTooltip />
                    <Legend wrapperStyle={{ fontSize: 12 }} />
                    <Area type="monotone" dataKey="created" name="Tickets Created" stroke="#3B82F6" fillOpacity={1} fill="url(#colorCreated)" strokeWidth={2} />
                    <Area type="monotone" dataKey="resolved" name="Tickets Resolved" stroke="#10B981" fillOpacity={1} fill="url(#colorResolved)" strokeWidth={2} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Category Distribution */}
        <Card className="lg:col-span-4 shadow-xs">
          <CardHeader className="pb-2">
            <div className="flex items-center gap-2">
              <PieChartIcon className="h-4 w-4 text-blue-600" />
              <div>
                <CardTitle className="text-sm font-semibold">Category Distribution</CardTitle>
                <p className="text-xs text-muted-foreground">Persentase tiket per kategori</p>
              </div>
            </div>
          </CardHeader>
          <Separator />
          <CardContent className="p-4">
            {loadingCategory ? (
              <div className="h-64 flex items-center justify-center text-xs text-muted-foreground">
                <Loader2 className="h-5 w-5 animate-spin mr-2 text-blue-500" />
                Memuat kategori...
              </div>
            ) : categoryChartData.length === 0 ? (
              <div className="h-64 flex flex-col items-center justify-center text-muted-foreground text-xs">
                <PieChartIcon className="h-8 w-8 mb-2 opacity-40" />
                Tidak ada data kategori
              </div>
            ) : (
              <div className="h-64 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={categoryChartData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="45%"
                      innerRadius={45}
                      outerRadius={72}
                      paddingAngle={2}
                    >
                      {categoryChartData.map((_, i) => (
                        <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                      ))}
                    </Pie>
                    <RechartsTooltip />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Category & Technician Performance Tables */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Category Performance */}
        <Card className="shadow-xs">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Tag className="h-4 w-4 text-blue-600" />
                <CardTitle className="text-sm font-semibold">Category Performance</CardTitle>
              </div>
              <Badge variant="secondary" className="text-xs">{categoryPerformance.length} Kategori</Badge>
            </div>
          </CardHeader>
          <Separator />
          <div className="overflow-x-auto">
            {loadingCategory ? (
              <div className="py-12 text-center text-xs text-muted-foreground">Memuat data...</div>
            ) : categoryPerformance.length === 0 ? (
              <div className="py-12 text-center text-xs text-muted-foreground">Belum ada data performa kategori</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50/70">
                    <TableHead className="text-xs">Kategori</TableHead>
                    <TableHead className="text-xs text-right">Total</TableHead>
                    <TableHead className="text-xs text-right">Resolved</TableHead>
                    <TableHead className="text-xs">Resolution Rate</TableHead>
                    <TableHead className="text-xs text-right">Avg Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {categoryPerformance.map((c) => (
                    <TableRow key={c.category}>
                      <TableCell>
                        <Badge variant="outline" className="text-xs font-medium">
                          {c.category_display || formatCategory(c.category)}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right text-sm font-medium">{c.total_tickets}</TableCell>
                      <TableCell className="text-right text-sm">{c.resolved_tickets}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 min-w-[110px]">
                          <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${getProgressBarColor(c.resolution_rate_percent)}`}
                              style={{ width: `${Math.min(100, c.resolution_rate_percent)}%` }}
                            />
                          </div>
                          <span className="text-xs font-semibold w-9 text-right">{Math.ceil(c.resolution_rate_percent)}%</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right text-xs font-medium text-muted-foreground">
                        {formatHours(c.avg_resolution_hours)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </Card>

        {/* Technician Performance */}
        <Card className="shadow-xs">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-emerald-600" />
                <CardTitle className="text-sm font-semibold">Technician Performance</CardTitle>
              </div>
              <Badge variant="secondary" className="bg-emerald-50 dark:bg-emerald-950/50 text-emerald-700 dark:text-emerald-300 text-xs">{userPerformance.length} Teknisi</Badge>
            </div>
          </CardHeader>
          <Separator />
          <div className="overflow-x-auto">
            {loadingUser ? (
              <div className="py-12 text-center text-xs text-muted-foreground">Memuat data...</div>
            ) : userPerformance.length === 0 ? (
              <div className="py-12 text-center text-xs text-muted-foreground">Belum ada data performa teknisi</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50/70">
                    <TableHead className="text-xs">Technician</TableHead>
                    <TableHead className="text-xs text-right">Assigned</TableHead>
                    <TableHead className="text-xs text-right">Resolved</TableHead>
                    <TableHead className="text-xs">Resolution Rate</TableHead>
                    <TableHead className="text-xs text-right">Avg Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {userPerformance.map((u) => (
                    <TableRow key={u.user_id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <div className="h-7 w-7 rounded-full bg-muted border flex items-center justify-center text-xs font-bold text-foreground/90">
                            {u.user_name.charAt(0).toUpperCase()}
                          </div>
                          <span className="text-sm font-medium text-foreground">{u.user_name}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right text-sm font-medium">{u.total_assigned}</TableCell>
                      <TableCell className="text-right text-sm">{u.resolved_tickets}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2 min-w-[110px]">
                          <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${getProgressBarColor(u.resolution_rate_percent)}`}
                              style={{ width: `${Math.min(100, u.resolution_rate_percent)}%` }}
                            />
                          </div>
                          <span className="text-xs font-semibold w-9 text-right">{Math.ceil(u.resolution_rate_percent)}%</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right text-xs font-medium text-muted-foreground">
                        {formatHours(u.avg_resolution_hours)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </Card>
      </div>

      {/* Downtime Analysis */}
      <Card className="shadow-xs">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Timer className="h-4 w-4 text-red-600" />
              <div>
                <CardTitle className="text-sm font-semibold">Downtime Analysis</CardTitle>
                <p className="text-xs text-muted-foreground">Pelacakan akumulasi durasi downtime dan pelanggan terdampak</p>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={loadDowntimeAnalysis} disabled={loadingDowntime} className="h-8 text-xs">
              <RefreshCw className={`h-3.5 w-3.5 mr-1 ${loadingDowntime ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </CardHeader>
        <Separator />
        <CardContent className="p-4 space-y-4">
          {loadingDowntime ? (
            <div className="py-10 text-center text-xs text-muted-foreground">Menganalisis data downtime...</div>
          ) : downtimeAnalysis?.overall_statistics ? (
            <>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
                <div className="border rounded-xl p-3.5 bg-blue-50/40 dark:bg-blue-950/30">
                  <div className="text-[11px] font-bold uppercase text-muted-foreground">Total Downtime</div>
                  <div className="text-xl font-bold text-foreground mt-1">
                    {formatHours(downtimeAnalysis.overall_statistics.total_downtime_hours)}
                  </div>
                  <div className="text-[11px] text-muted-foreground">Akumulasi semua tiket</div>
                </div>
                <div className="border rounded-xl p-3.5 bg-emerald-50/40 dark:bg-emerald-950/30">
                  <div className="text-[11px] font-bold uppercase text-muted-foreground">Avg Downtime</div>
                  <div className="text-xl font-bold text-emerald-700 dark:text-emerald-400 mt-1">
                    {formatHours(downtimeAnalysis.overall_statistics.avg_downtime_hours)}
                  </div>
                  <div className="text-[11px] text-muted-foreground">Rata-rata per tiket</div>
                </div>
                <div className="border rounded-xl p-3.5 bg-amber-50/40 dark:bg-amber-950/30">
                  <div className="text-[11px] font-bold uppercase text-muted-foreground">Max Downtime</div>
                  <div className="text-xl font-bold text-amber-700 dark:text-amber-400 mt-1">
                    {formatHours(downtimeAnalysis.overall_statistics.max_downtime_hours)}
                  </div>
                  <div className="text-[11px] text-muted-foreground">Durasi terlama</div>
                </div>
                <div className="border rounded-xl p-3.5 bg-rose-50/40 dark:bg-rose-950/30">
                  <div className="text-[11px] font-bold uppercase text-muted-foreground">Tiket Terdampak</div>
                  <div className="text-xl font-bold text-rose-700 dark:text-rose-400 mt-1">
                    {downtimeAnalysis.overall_statistics.tickets_with_downtime}
                  </div>
                  <div className="text-[11px] text-muted-foreground">Dengan catatan downtime</div>
                </div>
              </div>

              {downtimeAnalysis.top_customers && downtimeAnalysis.top_customers.length > 0 && (
                <div>
                  <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-2 flex items-center gap-1.5">
                    <AlertTriangle className="h-3.5 w-3.5 text-amber-500" />
                    Pelanggan dengan Downtime Tertinggi (Top 5)
                  </h4>
                  <div className="border rounded-lg overflow-hidden">
                    <Table>
                      <TableHeader>
                        <TableRow className="bg-muted/50">
                          <TableHead className="text-xs">Customer</TableHead>
                          <TableHead className="text-xs text-right">Tickets</TableHead>
                          <TableHead className="text-xs text-right">Total Downtime</TableHead>
                          <TableHead className="text-xs text-right">Avg Downtime</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {downtimeAnalysis.top_customers.slice(0, 5).map((cust) => (
                          <TableRow key={cust.customer_id}>
                            <TableCell className="text-sm font-medium text-foreground">{cust.customer_name}</TableCell>
                            <TableCell className="text-right text-sm">{cust.ticket_count}</TableCell>
                            <TableCell className="text-right">
                              <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold ${getDowntimeBadgeClass(cust.total_downtime_minutes)}`}>
                                {formatHours(cust.total_downtime_hours)}
                              </span>
                            </TableCell>
                            <TableCell className="text-right text-xs text-muted-foreground">
                              {formatHours(cust.avg_downtime_hours)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="py-10 text-center text-xs text-muted-foreground">
              Tidak ada data downtime pada periode ini.
            </div>
          )}
        </CardContent>
      </Card>

      {/* Detail Tiket & Informasi Pelanggan */}
      <Card className="shadow-xs">
        <CardHeader className="pb-2">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <CardTitle className="text-sm font-semibold">Detail Tiket & Informasi Pelanggan</CardTitle>
              <p className="text-xs text-muted-foreground">Data lengkap riwayat penanganan gangguan dan data teknis pelanggan</p>
            </div>
            <Badge variant="secondary" className="text-xs self-start sm:self-auto">
              {filteredTicketDetails.length} Tiket
            </Badge>
          </div>
        </CardHeader>
        <Separator />
        <CardContent className="p-4 space-y-3">
          {/* Filter Bar */}
          <div className="flex flex-col md:flex-row gap-2.5">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                value={ticketSearch}
                onChange={(e) => { setTicketSearch(e.target.value); setDetailsPage(1) }}
                placeholder="Cari no tiket, pelanggan, atau teknisi..."
                className="pl-9 h-9"
              />
            </div>
            <div className="grid grid-cols-2 md:flex gap-2.5">
              <Select value={ticketStatusFilter || '__all__'} onValueChange={(v) => { setTicketStatusFilter(v === '__all__' ? '' : v); setDetailsPage(1) }}>
                <SelectTrigger className="h-9 md:w-[160px]"><SelectValue placeholder="Filter Status" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Status</SelectItem>
                  {ticketStatusOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
              <Select value={ticketCategoryFilter || '__all__'} onValueChange={(v) => { setTicketCategoryFilter(v === '__all__' ? '' : v); setDetailsPage(1) }}>
                <SelectTrigger className="h-9 md:w-[170px]"><SelectValue placeholder="Filter Kategori" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Kategori</SelectItem>
                  {ticketCategoryOptions.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
              </Select>
              {(ticketSearch || ticketStatusFilter || ticketCategoryFilter) && (
                <Button variant="ghost" size="sm" onClick={clearTicketFilters} className="h-9 text-xs text-red-600">
                  <FilterX className="h-3.5 w-3.5 mr-1" />Reset
                </Button>
              )}
            </div>
          </div>

          {loadingDetails ? (
            <div className="py-12 text-center text-xs text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin inline mr-2 text-blue-500" />
              Memuat detail tiket...
            </div>
          ) : paginatedTicketDetails.length === 0 ? (
            <div className="py-12 text-center text-xs text-muted-foreground">Tidak ada tiket yang cocok dengan filter</div>
          ) : (
            <>
              <div className="border rounded-lg overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead className="text-xs">Ticket #</TableHead>
                      <TableHead className="text-xs">Pelanggan</TableHead>
                      <TableHead className="text-xs">Data Teknis</TableHead>
                      <TableHead className="text-xs">Masalah</TableHead>
                      <TableHead className="text-xs">Status & Kategori</TableHead>
                      <TableHead className="text-xs">Downtime</TableHead>
                      <TableHead className="text-xs">Assignment</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedTicketDetails.map((item) => (
                      <TableRow
                        key={item.id}
                        className="hover:bg-muted/40 cursor-pointer"
                        onClick={() => navigate(`/trouble-tickets/${item.id}/action`)}
                      >
                        <TableCell>
                          <span className="inline-flex items-center gap-1 text-xs font-mono font-semibold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-950/50 px-2 py-0.5 rounded">
                            {item.ticket_number}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm font-medium text-foreground">{item.customer_name}</div>
                          <div className="text-xs text-muted-foreground flex items-center gap-1">
                            <Phone className="h-3 w-3" /> {item.customer_phone}
                          </div>
                          {item.customer_address && (
                            <div className="text-[11px] text-muted-foreground truncate max-w-[200px] flex items-center gap-1">
                              <MapPin className="h-3 w-3 shrink-0" /> {item.customer_address}
                            </div>
                          )}
                        </TableCell>
                        <TableCell className="text-xs space-y-0.5">
                          <div><span className="text-muted-foreground">IP:</span> <span className="font-mono">{item.ip_pelanggan}</span></div>
                          <div><span className="text-muted-foreground">ID:</span> {item.id_pelanggan}</div>
                          <div><span className="text-muted-foreground">ONU:</span> {item.onu_power !== '-' ? `${item.onu_power} dBm` : '-'}</div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm font-medium text-foreground max-w-[220px] truncate">{item.title}</div>
                          <div className="text-xs text-muted-foreground max-w-[220px] truncate">{item.description || '-'}</div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col gap-1 items-start">
                            <span className={`inline-flex items-center text-[11px] font-semibold px-2 py-0.5 rounded-full border ${getStatusBadgeClass(item.status)}`}>
                              {formatStatus(item.status)}
                            </span>
                            <Badge variant="outline" className="text-[10px]">{formatCategory(item.category)}</Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold ${getDowntimeBadgeClass(item.total_downtime_minutes)}`}>
                            {item.downtime_hours} jam
                          </span>
                          <div className="text-[11px] text-muted-foreground mt-0.5">{item.total_downtime_minutes || 0} menit</div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 text-xs font-medium text-foreground">
                            <UserCheck className="h-3.5 w-3.5 text-emerald-600 shrink-0" />
                            {item.assigned_to}
                          </div>
                          <div className="text-[11px] text-muted-foreground">{item.latest_actions.length} riwayat tindakan</div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination */}
              <div className="flex items-center justify-between pt-2">
                <div className="text-xs text-muted-foreground">
                  Halaman {detailsPage} dari {totalDetailPages}
                </div>
                <div className="flex items-center gap-1">
                  <Button variant="outline" size="icon" className="h-8 w-8" disabled={detailsPage === 1} onClick={() => setDetailsPage((p) => p - 1)}>
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button variant="outline" size="icon" className="h-8 w-8" disabled={detailsPage >= totalDetailPages} onClick={() => setDetailsPage((p) => p + 1)}>
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Export Dialog */}
      <Dialog open={showExportDialog} onOpenChange={setShowExportDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FileDown className="h-5 w-5 text-emerald-600" />
              Export Laporan Trouble Ticket
            </DialogTitle>
            <DialogDescription>Unduh data analitik dan daftar tiket sesuai periode aktif</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div>
              <Label className="text-xs font-semibold text-muted-foreground mb-1.5 block">Filter Brand (Opsional)</Label>
              <Select value={exportBrandFilter || '__all__'} onValueChange={(v) => setExportBrandFilter(v === '__all__' ? '' : v)}>
                <SelectTrigger className="w-full"><SelectValue placeholder="Semua Brand" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">Semua Brand</SelectItem>
                  {brandOptions.map((b) => <SelectItem key={b.value} value={b.value}>{b.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label className="text-xs font-semibold text-muted-foreground block">Pilih Format File</Label>
              {[
                { key: 'excel' as const, title: 'Excel (XLS)', desc: 'Tabel terformat untuk Microsoft Excel', icon: <FileSpreadsheet className="h-5 w-5 text-emerald-600" /> },
                { key: 'csv' as const, title: 'CSV', desc: 'Data tabel lengkap dipisahkan koma', icon: <FileText className="h-5 w-5 text-blue-600" /> },
                { key: 'pdf' as const, title: 'PDF / Print Report', desc: 'Tampilan cetak ringkas siap simpan PDF', icon: <Printer className="h-5 w-5 text-red-600" /> },
                { key: 'json' as const, title: 'JSON Data', desc: 'Data mentah lengkap termasuk analitik', icon: <FileCode2 className="h-5 w-5 text-amber-600" /> },
              ].map((fmt) => (
                <button
                  key={fmt.key}
                  type="button"
                  disabled={exporting}
                  onClick={() => handleExport(fmt.key)}
                  className="w-full flex items-center justify-between p-3 rounded-xl border hover:bg-muted/50 hover:border-blue-300 transition-colors text-left disabled:opacity-50"
                >
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-muted/50 border">{fmt.icon}</div>
                    <div>
                      <div className="text-sm font-semibold text-foreground">{fmt.title}</div>
                      <div className="text-xs text-muted-foreground">{fmt.desc}</div>
                    </div>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
              ))}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowExportDialog(false)} disabled={exporting}>
              Batal
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
