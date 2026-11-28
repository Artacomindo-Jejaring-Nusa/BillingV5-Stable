import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import {
  AlertTriangle,
  Archive,
  Bell,
  Calculator,
  CheckCheck,
  ChevronDown,
  ChevronRight,
  CircleCheck,
  Crown,
  FileEdit,
  FileText,
  History,
  KeyRound,
  LayoutDashboard,
  LifeBuoy,
  Loader2,
  LogOut,
  Menu,
  MessageSquare,
  Moon,
  Network,
  Package,
  Percent,
  Pin,
  PinOff,
  Receipt,
  RefreshCw,
  Router,
  Search,
  Server,
  Settings,
  Shield,
  ShieldCheck,
  Sun,
  TrendingUp,
  Tv,
  UserPlus,
  Users,
  Wifi,
  Wrench,
  BarChart3,
  GitFork,
  X,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { getDefaultHomeRoute } from '@/router/guards'
import { useSettingsStore } from '@/stores/settings'
import { useMutationLoading } from '@/hooks/useMutationLoading'
import {
  formatNotificationTime,
  getNotificationMessage,
  getNotificationRoute,
  getNotificationTitle,
  useNotifications,
  type AppNotification,
} from '@/hooks/useNotifications'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import logoLight from '@/assets/images/Jelantik-Light.webp'
import logoDark from '@/assets/images/Jelantik-Dark.webp'

interface MenuItem {
  title: string
  to: string
  icon: React.ComponentType<{ className?: string }>
  permission: string | null
  value: string
}

interface MenuGroup {
  section: string
  title: string
  value: string
  icon: React.ComponentType<{ className?: string }>
  permission: string | null
  items: MenuItem[]
}

const MENU_GROUPS: MenuGroup[] = [
  {
    section: 'DASHBOARD',
    title: 'Dashboard',
    value: 'dashboard-group',
    icon: LayoutDashboard,
    permission: 'view_dashboard',
    items: [
      { title: 'Dashboard Admin', to: '/dashboard', icon: Crown, permission: 'view_dashboard', value: 'dashboard-admin' },
      { title: 'Dashboard Jakinet', to: '/dashboard-pelanggan', icon: Users, permission: 'view_dashboard_pelanggan', value: 'dashboard-jakinet' },
      { title: 'Monitoring TV / NOC', to: '/tv-monitoring', icon: Tv, permission: 'view_dashboard', value: 'monitoring-tv' },
    ],
  },
  {
    section: 'FTTH',
    title: 'Management Pelanggan',
    value: 'management-pelanggan',
    icon: Network,
    permission: null,
    items: [
      { title: 'Pelanggan', to: '/pelanggan', icon: Users, permission: 'view_pelanggan', value: 'pelanggan' },
      { title: 'Data Teknis', to: '/data-teknis', icon: GitFork, permission: 'view_data_teknis', value: 'teknis' },
      { title: 'Langganan', to: '/langganan', icon: Wifi, permission: 'view_langganan', value: 'langganan' },
      { title: 'Brand & Paket', to: '/harga-layanan', icon: Package, permission: 'view_brand_&_paket', value: 'harga' },
    ],
  },
  {
    section: 'BILLING',
    title: 'Billing & Reports',
    value: 'billing-group',
    icon: Receipt,
    permission: null,
    items: [
      { title: 'Invoices', to: '/invoices', icon: Receipt, permission: 'view_invoices', value: 'invoices' },
      { title: 'Diskon', to: '/diskon', icon: Percent, permission: 'view_diskon', value: 'diskon' },
      { title: 'Laporan Pendapatan', to: '/reports/revenue', icon: TrendingUp, permission: 'view_reports_revenue', value: 'revenue-report' },
    ],
  },
  {
    section: 'LAINNYA',
    title: 'Tools & Resources',
    value: 'tools-group',
    icon: Wrench,
    permission: null,
    items: [
      { title: 'Simulasi Harga', to: '/kalkulator', icon: Calculator, permission: 'view_simulasi_harga', value: 'kalkulator' },
      { title: 'Kalkulator Diskon', to: '/kalkulator-diskon', icon: Percent, permission: 'view_simulasi_harga', value: 'kalkulator-diskon' },
      { title: 'S&K', to: '/syarat-ketentuan', icon: FileText, permission: null, value: 'sk' },
    ],
  },
  {
    section: 'SUPPORT',
    title: 'Customer Support',
    value: 'support-group',
    icon: LifeBuoy,
    permission: null,
    items: [
      { title: 'Live Chat Pelanggan', to: '/customer-chat', icon: MessageSquare, permission: 'view_customer_chat', value: 'customer-chat' },
      { title: 'Trouble Tickets', to: '/trouble-tickets', icon: LifeBuoy, permission: 'view_trouble_tickets', value: 'trouble-tickets' },
      { title: 'Ticket Reports', to: '/trouble-tickets/reports', icon: BarChart3, permission: 'view_trouble_tickets', value: 'trouble-ticket-reports' },
    ],
  },
  {
    section: 'NETWORK MANAGEMENT',
    title: 'Network Management',
    value: 'network-management-group',
    icon: Server,
    permission: null,
    items: [
      { title: 'Mikrotik Servers', to: '/mikrotik', icon: Server, permission: 'view_mikrotik_servers', value: 'mikrotik' },
      { title: 'OLT Management', to: '/network-management/olt', icon: Router, permission: 'view_olt', value: 'olt' },
      { title: 'ODP Management', to: '/odp-management', icon: Network, permission: 'view_odp_management', value: 'odp' },
      { title: 'Manajemen Inventaris', to: '/inventory', icon: Archive, permission: 'view_inventory', value: 'inventory' },
    ],
  },
  {
    section: 'MANAGEMENT',
    title: 'System Management',
    value: 'system-management-group',
    icon: Settings,
    permission: null,
    items: [
      { title: 'Users', to: '/users', icon: Users, permission: 'view_users', value: 'users' },
      { title: 'Roles', to: '/roles', icon: Shield, permission: 'view_roles', value: 'roles' },
      { title: 'Permissions', to: '/permissions', icon: ShieldCheck, permission: 'view_permissions', value: 'permissions' },
      { title: 'Activity Log', to: '/activity-logs', icon: History, permission: 'view_activity_log', value: 'activity-logs' },
      { title: 'Kelola S&K', to: '/management/sk', icon: FileEdit, permission: 'manage_sk', value: 'sk-management' },
      { title: 'Pengaturan', to: '/management/settings', icon: Settings, permission: 'manage_settings', value: 'settings' },
      { title: 'API Keys', to: '/management/api-keys', icon: KeyRound, permission: 'manage_settings', value: 'api-keys' },
    ],
  },
]

export function DefaultLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const token = useAuthStore((s) => s.token)
  const logout = useAuthStore((s) => s.logout)
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const isMutatingLoading = useMutationLoading()

  const maintenanceMode = useSettingsStore((s) => s.maintenanceMode)
  const theme = useSettingsStore((s) => s.theme)
  const toggleTheme = useSettingsStore((s) => s.toggleTheme)

  const [isPinned, setIsPinned] = useState(true)
  const [isHovered, setIsHovered] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({
    'dashboard-group': true,
    'management-pelanggan': true,
    'billing-group': true,
    'support-group': true,
  })
  const [badges, setBadges] = useState({
    suspended_count: 0,
    stopped_count: 0,
    unpaid_invoice_count: 0,
    total_invoice_count: 0,
    open_tickets_count: 0,
  })

  const fetchBadges = useCallback(async () => {
    try {
      const res = await apiClient.get('/dashboard/sidebar-badges', { headers: { 'X-Silent': 'true' } })
      if (res.data) {
        setBadges({
          suspended_count: Number(res.data.suspended_count || 0),
          stopped_count: Number(res.data.stopped_count || 0),
          unpaid_invoice_count: Number(res.data.unpaid_invoice_count || 0),
          total_invoice_count: Number(res.data.total_invoice_count || 0),
          open_tickets_count: Number(res.data.open_tickets_count || 0),
        })
      }
    } catch {
      // ignore silent badge fetch errors
    }
  }, [])

  useEffect(() => {
    void fetchBadges()
    const interval = setInterval(() => void fetchBadges(), 60000)
    return () => clearInterval(interval)
  }, [fetchBadges])

  const { notifications, toast, dismissToast, markAsRead, markAllAsRead } = useNotifications({
    token,
    user,
    canViewChat: hasPermission('view_customer_chat'),
    currentPathname: location.pathname,
    onRefreshBadges: fetchBadges,
  })

  const handleNotificationClick = async (notif: AppNotification) => {
    await markAsRead(notif)
    navigate(getNotificationRoute(notif.type))
  }

  const renderNotifIcon = (type: string) => {
    switch (type) {
      case 'new_payment':
        return <CircleCheck className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
      case 'new_customer_for_noc':
      case 'new_customer':
        return <UserPlus className="h-4 w-4 text-sky-600 dark:text-sky-400" />
      case 'new_technical_data':
        return <Network className="h-4 w-4 text-blue-600 dark:text-blue-400" />
      case 'new_chat':
        return <MessageSquare className="h-4 w-4 text-amber-600 dark:text-amber-400" />
      default:
        return <Bell className="h-4 w-4 text-muted-foreground" />
    }
  }

  const isSidebarExpanded = isPinned || isHovered
  const roleName = user?.role?.name?.toLowerCase() || ''
  const isAdmin = ['admin', 'superadmin', 'super admin', 'direktur'].includes(roleName)

  const visibleGroups = useMemo(() => {
    return MENU_GROUPS.map((group) => {
      const items = group.items.filter((item) => {
        if (!item.permission) return true
        // When user permissions haven't loaded yet, keep items visible until /users/me resolves
        if (!user) return true
        return hasPermission(item.permission)
      })
      return { ...group, items }
    }).filter((group) => group.items.length > 0)
  }, [user, hasPermission])

  const currentPageTitle = useMemo(() => {
    for (const group of MENU_GROUPS) {
      const found = group.items.find((item) => location.pathname === item.to || location.pathname.startsWith(`${item.to}/`))
      if (found) return found.title
    }
    return 'Artacom Billing System'
  }, [location.pathname])

  const toggleGroup = (value: string) => {
    setOpenGroups((prev) => ({ ...prev, [value]: !prev[value] }))
  }

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = searchQuery.trim()
    if (!trimmed) return
    navigate(`/global-search-results?q=${encodeURIComponent(trimmed)}`)
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const userInitials = user?.name
    ? user.name
        .split(' ')
        .map((n) => n[0])
        .slice(0, 2)
        .join('')
        .toUpperCase()
    : 'AD'

  const renderSidebarNav = (expanded: boolean, onItemClick?: () => void) => (
    <div className="flex flex-1 flex-col overflow-y-auto px-3 py-3 space-y-4">
      {visibleGroups.map((group) => {
        const isGroupOpen = openGroups[group.value] ?? false
        const GroupIcon = group.icon
        return (
          <div key={group.value} className="space-y-1">
            {expanded && (
              <div className="px-2 text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
                {group.section}
              </div>
            )}
            <button
              type="button"
              onClick={() => toggleGroup(group.value)}
              className={cn(
                'flex w-full items-center gap-3 rounded-md px-2.5 py-2 text-sm font-medium text-foreground/85 hover:bg-accent hover:text-accent-foreground transition-colors',
                !expanded && 'justify-center px-2'
              )}
              title={!expanded ? group.title : undefined}
            >
              <GroupIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
              {expanded && (
                <>
                  <span className="flex-1 text-left truncate">{group.title}</span>
                  {isGroupOpen ? (
                    <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
                  ) : (
                    <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
                  )}
                </>
              )}
            </button>

            {(isGroupOpen || !expanded) && (
              <div className={cn('space-y-0.5', expanded && 'pl-3 border-l border-border ml-4')}>
                {group.items.map((item) => {
                  const ItemIcon = item.icon
                  const isActive = location.pathname === item.to
                  return (
                    <Link
                      key={item.value}
                      to={item.to}
                      onClick={onItemClick}
                      title={!expanded ? item.title : undefined}
                      className={cn(
                        'flex items-center gap-3 rounded-md px-2.5 py-2 text-sm transition-colors min-h-[40px]',
                        isActive
                          ? 'bg-primary text-primary-foreground font-medium'
                          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
                        !expanded && 'justify-center px-2'
                      )}
                    >
                      <ItemIcon className="h-4 w-4 shrink-0" />
                      {expanded && <span className="truncate flex-1">{item.title}</span>}
                      {expanded && item.value === 'langganan' && (
                        <div className="flex items-center gap-1 shrink-0">
                          {badges.suspended_count > 0 && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="inline-flex items-center rounded-full border border-red-300 bg-red-50 px-1.5 py-0.5 text-[10px] font-semibold text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
                                  {badges.suspended_count}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent side="right">
                                {badges.suspended_count} langganan berstatus "Suspended"
                              </TooltipContent>
                            </Tooltip>
                          )}
                          {badges.stopped_count > 0 && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="inline-flex items-center rounded-full border border-slate-300 bg-slate-100 px-1.5 py-0.5 text-[10px] font-semibold text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300">
                                  {badges.stopped_count}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent side="right">
                                {badges.stopped_count} langganan berstatus "Berhenti"
                              </TooltipContent>
                            </Tooltip>
                          )}
                        </div>
                      )}
                      {expanded && item.value === 'invoices' && (
                        <div className="flex items-center gap-1 shrink-0">
                          {badges.unpaid_invoice_count > 0 && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="inline-flex items-center rounded-full border border-amber-300 bg-amber-50 px-1.5 py-0.5 text-[10px] font-semibold text-amber-700 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-300">
                                  {badges.unpaid_invoice_count}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent side="right">
                                {badges.unpaid_invoice_count} invoice belum dibayar
                              </TooltipContent>
                            </Tooltip>
                          )}
                          {badges.total_invoice_count > 0 && (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <span className="inline-flex items-center rounded-full border border-sky-300 bg-sky-50 px-1.5 py-0.5 text-[10px] font-semibold text-sky-700 dark:border-sky-800 dark:bg-sky-950 dark:text-sky-300">
                                  {badges.total_invoice_count}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent side="right">
                                {badges.total_invoice_count} total keseluruhan invoice
                              </TooltipContent>
                            </Tooltip>
                          )}
                        </div>
                      )}
                      {expanded && item.value === 'trouble-tickets' && badges.open_tickets_count > 0 && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="inline-flex items-center rounded-full border border-rose-300 bg-rose-50 px-1.5 py-0.5 text-[10px] font-semibold text-rose-700 dark:border-rose-800 dark:bg-rose-950 dark:text-rose-300">
                              {badges.open_tickets_count}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent side="right">
                            {badges.open_tickets_count} tiket gangguan terbuka
                          </TooltipContent>
                        </Tooltip>
                      )}
                    </Link>
                  )
                })}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col">
      {/* Global Mutation Loading Overlay (above Dialogs) */}
      {isMutatingLoading && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-slate-950/40 backdrop-blur-[1px]"
          role="status"
          aria-live="polite"
        >
          <div className="rounded-xl border bg-card px-7 py-6 shadow-xl text-center space-y-2.5 min-w-[220px]">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto" />
            <div className="text-sm font-semibold">Sedang Memproses...</div>
            <div className="text-xs text-muted-foreground">Mohon tunggu sebentar</div>
          </div>
        </div>
      )}

      {/* Maintenance Overlay for Non-Admin */}
      {maintenanceMode.isActive && Boolean(user) && !isAdmin && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/95 p-4">
          <div className="max-w-md w-full rounded-xl border bg-card p-6 text-center shadow-lg space-y-4">
            <AlertTriangle className="h-12 w-12 text-amber-500 mx-auto" />
            <h1 className="text-xl font-bold">Sistem Sedang Dalam Maintenance</h1>
            <p className="text-sm text-muted-foreground">{maintenanceMode.message}</p>
            <Button onClick={() => window.location.reload()} className="w-full">
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh Halaman
            </Button>
          </div>
        </div>
      )}

      {/* Admin Maintenance Banner */}
      {maintenanceMode.isActive && isAdmin && (
        <div className="bg-amber-500 text-amber-950 px-4 py-1.5 text-xs font-medium flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 shrink-0" />
            <span>
              <strong>Mode Maintenance Aktif:</strong> {maintenanceMode.message}
            </span>
          </div>
          <Badge variant="destructive" className="text-[11px]">
            Admin Access
          </Badge>
        </div>
      )}

      <div className="flex flex-1 min-h-0">
        {/* Desktop Sidebar */}
        <aside
          onMouseEnter={() => !isPinned && setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
          className={cn(
            'hidden lg:flex flex-col border-r bg-card transition-all duration-200 shrink-0 sticky top-0 h-screen',
            isSidebarExpanded ? 'w-68' : 'w-[70px]'
          )}
        >
          {/* Sidebar Header */}
          <div className="h-16 flex items-center justify-between px-3 border-b shrink-0">
            <Link to={getDefaultHomeRoute(hasPermission) || '/data-teknis'} className="flex items-center gap-2.5 min-w-0">
              {isSidebarExpanded ? (
                <>
                  <img
                    src={theme === 'dark' ? logoDark : logoLight}
                    alt="Jelantik Logo"
                    className="h-9 w-9 shrink-0 object-contain"
                  />
                  <div className="min-w-0 leading-tight">
                    <div className="text-sm font-bold tracking-tight truncate">ARTACOM FTTH</div>
                    <div className="text-[11px] font-semibold tracking-wide text-primary uppercase truncate">
                      PORTAL CUSTOMER V5
                    </div>
                  </div>
                </>
              ) : (
                <div className="h-9 w-9 rounded-lg bg-primary text-primary-foreground flex items-center justify-center font-bold text-base shrink-0">
                  J
                </div>
              )}
            </Link>
            {isSidebarExpanded && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0"
                    onClick={() => setIsPinned((p) => !p)}
                    aria-label={isPinned ? 'Lepas pin sidebar' : 'Kunci sidebar'}
                  >
                    {isPinned ? <Pin className="h-4 w-4" /> : <PinOff className="h-4 w-4 text-muted-foreground" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">
                  {isPinned ? 'Lepas pin (Auto-hide)' : 'Kunci sidebar terbuka'}
                </TooltipContent>
              </Tooltip>
            )}
          </div>

          {renderSidebarNav(isSidebarExpanded)}

          {/* Sidebar Footer */}
          <div className="border-t p-3 shrink-0">
            <Button
              variant="outline"
              onClick={handleLogout}
              className={cn('w-full', !isSidebarExpanded && 'px-0')}
              title={!isSidebarExpanded ? 'Keluar' : undefined}
            >
              <LogOut className="h-4 w-4 shrink-0" />
              {isSidebarExpanded && <span className="ml-2">Keluar</span>}
            </Button>
          </div>
        </aside>

        {/* Mobile Drawer */}
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetContent side="left" className="w-72 p-0 flex flex-col">
            <SheetHeader className="h-16 px-4 border-b flex flex-row items-center gap-2.5 text-left">
              <img
                src={theme === 'dark' ? logoDark : logoLight}
                alt="Jelantik Logo"
                className="h-9 w-9 shrink-0 object-contain"
              />
              <div className="min-w-0 leading-tight">
                <SheetTitle className="text-sm font-bold tracking-tight">ARTACOM FTTH</SheetTitle>
                <p className="text-[11px] font-semibold tracking-wide text-primary uppercase">
                  PORTAL CUSTOMER V5
                </p>
              </div>
            </SheetHeader>
            {renderSidebarNav(true, () => setMobileOpen(false))}
            <div className="border-t p-3">
              <Button variant="outline" onClick={handleLogout} className="w-full min-h-[44px]">
                <LogOut className="mr-2 h-4 w-4" />
                Keluar
              </Button>
            </div>
          </SheetContent>
        </Sheet>

        {/* Main Column */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Topbar */}
          <header className="h-16 border-b bg-card px-4 flex items-center justify-between gap-3 sticky top-0 z-30">
            <div className="flex items-center gap-3 min-w-0">
              <Button
                variant="ghost"
                size="icon"
                className="lg:hidden min-h-[44px] min-w-[44px]"
                onClick={() => setMobileOpen(true)}
                aria-label="Buka menu navigasi"
              >
                <Menu className="h-5 w-5" />
              </Button>
              <h1 className="text-base sm:text-lg font-semibold truncate">{currentPageTitle}</h1>
            </div>

            <div className="flex items-center gap-2 sm:gap-3">
              {/* Global Search */}
              <form onSubmit={handleSearchSubmit} className="hidden md:flex items-center relative w-64">
                <Search className="h-4 w-4 text-muted-foreground absolute left-2.5 pointer-events-none" />
                <Input
                  type="search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Cari pelanggan, invoice..."
                  className="pl-8 h-9 text-sm"
                />
              </form>

              {/* Theme Toggle */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={toggleTheme}
                    aria-label={theme === 'dark' ? 'Aktifkan mode terang' : 'Aktifkan mode gelap'}
                    className="min-h-[40px] min-w-[40px]"
                  >
                    {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{theme === 'dark' ? 'Mode Terang' : 'Mode Gelap'}</TooltipContent>
              </Tooltip>

              {/* Notifications Dropdown */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Notifikasi"
                    className="relative min-h-[40px] min-w-[40px]"
                  >
                    <Bell className="h-4 w-4" />
                    {notifications.length > 0 && (
                      <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-destructive text-destructive-foreground text-[10px] font-bold flex items-center justify-center">
                        {notifications.length > 99 ? '99+' : notifications.length}
                      </span>
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-80 sm:w-96 p-0">
                  <DropdownMenuLabel className="flex items-center justify-between px-3.5 py-2.5">
                    <div>
                      <div className="text-sm font-semibold">Notifikasi</div>
                      <div className="text-[11px] font-normal text-muted-foreground">
                        {notifications.length === 0
                          ? 'Tidak ada notifikasi baru'
                          : `${notifications.length} notifikasi baru`}
                      </div>
                    </div>
                    {notifications.length > 0 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 px-2 text-[11px] text-primary"
                        onClick={() => void markAllAsRead()}
                      >
                        <CheckCheck className="h-3.5 w-3.5 mr-1" />
                        Tandai Dibaca
                      </Button>
                    )}
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator className="m-0" />
                  {notifications.length === 0 ? (
                    <div className="py-8 text-center text-sm text-muted-foreground">
                      Belum ada notifikasi baru saat ini.
                    </div>
                  ) : (
                    <div className="max-h-80 overflow-y-auto divide-y divide-border">
                      {notifications.map((notif) => (
                        <button
                          key={notif.id}
                          type="button"
                          onClick={() => void handleNotificationClick(notif)}
                          className="w-full px-3.5 py-2.5 flex items-start gap-3 text-left hover:bg-muted/50 transition-colors"
                        >
                          <div className="h-8 w-8 rounded-lg bg-muted flex items-center justify-center shrink-0 mt-0.5">
                            {renderNotifIcon(notif.type)}
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center justify-between gap-2">
                              <span className="text-xs font-semibold text-foreground truncate">
                                {getNotificationTitle(notif.type)}
                              </span>
                              <span className="text-[10px] text-muted-foreground shrink-0">
                                {formatNotificationTime(notif.created_at || notif.timestamp)}
                              </span>
                            </div>
                            <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                              {getNotificationMessage(notif)}
                            </p>
                          </div>
                        </button>
                      ))}
                    </div>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>

              <Separator orientation="vertical" className="h-6 hidden sm:block" />

              {/* User Menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" className="flex items-center gap-2 px-2 h-10">
                    <Avatar className="h-7 w-7">
                      <AvatarFallback className="text-xs bg-primary text-primary-foreground">
                        {userInitials}
                      </AvatarFallback>
                    </Avatar>
                    <div className="hidden sm:block text-left leading-tight">
                      <div className="text-xs font-medium truncate max-w-[120px]">
                        {user?.name || 'Memuat...'}
                      </div>
                      <div className="text-[10px] text-muted-foreground truncate max-w-[120px]">
                        {user?.role?.name || 'User'}
                      </div>
                    </div>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-52">
                  <DropdownMenuLabel>
                    <div className="text-sm font-medium">{user?.name || 'Akun'}</div>
                    <div className="text-xs font-normal text-muted-foreground truncate">{user?.email}</div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => navigate('/management/settings')}>
                    <Settings className="mr-2 h-4 w-4" />
                    Pengaturan
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={handleLogout} className="text-destructive focus:text-destructive">
                    <LogOut className="mr-2 h-4 w-4" />
                    Keluar
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </header>

          {/* Page Content */}
          <main className="flex-1 bg-muted/30">
            <ErrorBoundary>
              <Outlet />
            </ErrorBoundary>
          </main>

          {/* Footer */}
          <footer className="border-t bg-card px-4 py-3 text-center text-xs text-muted-foreground print:hidden">
            <span>
              &copy; {new Date().getFullYear()}{' '}
              <strong className="font-semibold text-foreground">Artacom Billing System</strong>. All
              Rights Reserved. Designed by{' '}
              <a
                href="https://www.instagram.com/amad.dyk/"
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary hover:underline"
              >
                amad.dyk
              </a>
            </span>
          </footer>
        </div>
      </div>

      {/* Real-time WebSocket Notification Toast */}
      {toast && (
        <div
          role="button"
          tabIndex={0}
          onClick={() => {
            dismissToast()
            navigate(getNotificationRoute(toast.type))
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              dismissToast()
              navigate(getNotificationRoute(toast.type))
            }
          }}
          className="fixed bottom-5 right-5 z-50 max-w-sm w-[calc(100%-2.5rem)] rounded-xl border bg-card text-foreground shadow-lg p-3.5 flex items-start gap-3 cursor-pointer hover:bg-muted/40 transition-colors"
        >
          <div className="h-9 w-9 rounded-lg bg-muted flex items-center justify-center shrink-0">
            {renderNotifIcon(toast.type)}
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-xs font-bold text-foreground">{toast.title}</div>
            <div className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{toast.message}</div>
          </div>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              dismissToast()
            }}
            aria-label="Tutup notifikasi"
            className="text-muted-foreground hover:text-foreground shrink-0"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  )
}
