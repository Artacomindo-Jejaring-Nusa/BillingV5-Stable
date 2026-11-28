import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { getEncryptedToken } from '@/utils/crypto'

const HOME_ROUTE_PRIORITY: Array<{ permission: string; path: string }> = [
  { permission: 'view_dashboard', path: '/dashboard' },
  { permission: 'view_dashboard_pelanggan', path: '/dashboard-pelanggan' },
  { permission: 'view_data_teknis', path: '/data-teknis' },
  { permission: 'view_trouble_tickets', path: '/trouble-tickets' },
  { permission: 'view_pelanggan', path: '/pelanggan' },
  { permission: 'view_langganan', path: '/langganan' },
  { permission: 'view_invoices', path: '/invoices' },
  { permission: 'view_customer_chat', path: '/customer-chat' },
  { permission: 'view_mikrotik_servers', path: '/mikrotik' },
  { permission: 'view_olt', path: '/olt' },
  { permission: 'view_odp_management', path: '/odp-management' },
  { permission: 'view_inventory', path: '/inventory' },
  { permission: 'view_users', path: '/users' },
  { permission: 'view_roles', path: '/roles' },
  { permission: 'view_permissions', path: '/permissions' },
]

export function getDefaultHomeRoute(hasPermission: (p: string) => boolean): string | null {
  for (const item of HOME_ROUTE_PRIORITY) {
    if (hasPermission(item.permission)) {
      return item.path
    }
  }
  return null
}

export function HomeRedirect() {
  const user = useAuthStore((s) => s.user)
  const hasPermission = useAuthStore((s) => s.hasPermission)
  if (!user) return null
  const target = getDefaultHomeRoute(hasPermission)
  if (!target) {
    return (
      <div className="p-8 text-center text-sm text-muted-foreground">
        Akun Anda belum memiliki hak akses halaman. Hubungi Administrator.
      </div>
    )
  }
  return <Navigate to={target} replace />
}

export function ProtectedRoute() {
  const token = useAuthStore((s) => s.token) || getEncryptedToken('access_token')
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}

export function GuestRoute() {
  const token = useAuthStore((s) => s.token) || getEncryptedToken('access_token')
  const hasPermission = useAuthStore((s) => s.hasPermission)
  if (token) {
    return <Navigate to={getDefaultHomeRoute(hasPermission) || '/dashboard'} replace />
  }
  return <Outlet />
}

interface PermissionRouteProps {
  permission?: string
  permissions?: string[]
  roles?: string[]
  children: React.ReactNode
}

export function PermissionRoute({ permission, permissions, roles, children }: PermissionRouteProps) {
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const hasAnyPermission = useAuthStore((s) => s.hasAnyPermission)
  const hasRole = useAuthStore((s) => s.hasRole)

  // Wait until user profile is loaded before enforcing role/permission check
  if (!user) return <>{children}</>

  const denied =
    (permission && !hasPermission(permission)) ||
    (permissions && permissions.length > 0 && !hasAnyPermission(permissions)) ||
    (roles && roles.length > 0 && !hasRole(roles))

  if (denied) {
    const fallback = getDefaultHomeRoute(hasPermission)
    if (fallback && fallback !== location.pathname) {
      return <Navigate to={fallback} replace />
    }
    return (
      <div className="p-8 text-center">
        <h2 className="text-lg font-semibold text-foreground">Akses Ditolak</h2>
        <p className="text-sm text-muted-foreground mt-1">Anda tidak memiliki hak akses untuk membuka halaman ini.</p>
      </div>
    )
  }

  return <>{children}</>
}
