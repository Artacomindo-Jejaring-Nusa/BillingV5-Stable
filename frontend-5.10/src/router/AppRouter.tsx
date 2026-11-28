import { useEffect } from 'react'
import { BrowserRouter, HashRouter, Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import { setApiNavigateHandler } from '@/services/api'
import { GuestRoute, HomeRedirect, PermissionRoute, ProtectedRoute } from './guards'
import { PlaceholderPage } from '@/components/PlaceholderPage'
import { DefaultLayout } from '@/layouts/DefaultLayout'
import { LoginView } from '@/views/LoginView'
import { NotFoundView } from '@/views/NotFoundView'
import { NetworkErrorView } from '@/views/NetworkErrorView'
import { PrivacyPolicyView } from '@/views/PrivacyPolicyView'
import { DashboardView } from '@/views/DashboardView'
import { DashboardPelangganView } from '@/views/DashboardPelangganView'
import { RevenueReportView } from '@/views/RevenueReportView'
import { TVMonitoringView } from '@/views/TVMonitoringView'
import { PelangganView } from '@/views/PelangganView'
import { DataTeknisView } from '@/views/DataTeknisView'
import { LanggananView } from '@/views/LanggananView'
import { HargaLayananView } from '@/views/HargaLayananView'
import { InvoicesView } from '@/views/InvoicesView'
import { DiskonView } from '@/views/DiskonView'
import { CalculatorView } from '@/views/CalculatorView'
import { KalkulatorDiskonView } from '@/views/KalkulatorDiskonView'
import { SKView } from '@/views/SKView'
import { SKManagementView } from '@/views/SKManagementView'
import { SystemSettingsView } from '@/views/SystemSettingsView'
import { ApiKeysView } from '@/views/ApiKeysView'
import { GlobalSearchResultsView } from '@/views/GlobalSearchResultsView'
import { CustomerChatView } from '@/views/CustomerChatView'
import { TroubleTicketView } from '@/views/TroubleTicketView'
import { TroubleTicketActionView } from '@/views/TroubleTicketActionView'
import { TroubleTicketReportView } from '@/views/TroubleTicketReportView'
import { MikrotikView } from '@/views/MikrotikView'
import { OLTView } from '@/views/OLTView'
import { ODPView } from '@/views/ODPView'
import { InventoryView } from '@/views/InventoryView'
import { UsersView } from '@/views/UsersView'
import { RolesView } from '@/views/RolesView'
import { PermissionsView } from '@/views/PermissionsView'
import { ActivityLogView } from '@/views/ActivityLogView'

function ApiNavigationBinder() {
  const navigate = useNavigate()
  useEffect(() => {
    setApiNavigateHandler((to) => navigate(to))
  }, [navigate])
  return null
}

function AppRoutes() {
  return (
    <>
      <ApiNavigationBinder />
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route element={<DefaultLayout />}>
          <Route path="/" element={<HomeRedirect />} />
          <Route
            path="/dashboard"
            element={
              <PermissionRoute permission="view_dashboard">
                <DashboardView />
              </PermissionRoute>
            }
          />
          <Route
            path="/dashboard-pelanggan"
            element={
              <PermissionRoute permission="view_dashboard_pelanggan">
                <DashboardPelangganView />
              </PermissionRoute>
            }
          />
          <Route
            path="/reports/revenue"
            element={
              <PermissionRoute permission="view_reports_revenue">
                <RevenueReportView />
              </PermissionRoute>
            }
          />

          <Route path="/pelanggan" element={<PermissionRoute permission="view_pelanggan"><PelangganView /></PermissionRoute>} />
          <Route path="/langganan" element={<PermissionRoute permission="view_langganan"><LanggananView /></PermissionRoute>} />
          <Route path="/langganan/:id/edit" element={<PermissionRoute permission="view_langganan"><LanggananView /></PermissionRoute>} />
          <Route path="/invoices" element={<PermissionRoute permission="view_invoices"><InvoicesView /></PermissionRoute>} />
          <Route path="/harga-layanan" element={<PermissionRoute permission="view_brand_&_paket"><HargaLayananView /></PermissionRoute>} />
          <Route
            path="/diskon"
            element={
              <PermissionRoute permissions={['view_diskon', 'manage_diskon']}>
                <DiskonView />
              </PermissionRoute>
            }
          />
          <Route
            path="/kalkulator"
            element={
              <PermissionRoute permissions={['view_simulasi_harga', 'use_calculator']}>
                <CalculatorView />
              </PermissionRoute>
            }
          />
          <Route
            path="/kalkulator-diskon"
            element={
              <PermissionRoute permissions={['view_simulasi_harga', 'use_calculator']}>
                <KalkulatorDiskonView />
              </PermissionRoute>
            }
          />

          <Route
            path="/trouble-tickets"
            element={
              <PermissionRoute permissions={['view_trouble_tickets']}>
                <TroubleTicketView />
              </PermissionRoute>
            }
          />
          <Route
            path="/trouble-tickets/:id"
            element={
              <PermissionRoute permissions={['view_trouble_tickets']}>
                <TroubleTicketActionView />
              </PermissionRoute>
            }
          />
          <Route
            path="/trouble-tickets/:id/action"
            element={
              <PermissionRoute permissions={['view_trouble_tickets']}>
                <TroubleTicketActionView />
              </PermissionRoute>
            }
          />
          <Route
            path="/trouble-tickets/reports"
            element={
              <PermissionRoute permissions={['view_trouble_tickets']}>
                <TroubleTicketReportView />
              </PermissionRoute>
            }
          />
          <Route
            path="/customer-chat"
            element={
              <PermissionRoute permission="view_customer_chat">
                <CustomerChatView />
              </PermissionRoute>
            }
          />

          <Route path="/data-teknis" element={<PermissionRoute permission="view_data_teknis"><DataTeknisView /></PermissionRoute>} />
          <Route path="/mikrotik" element={<PermissionRoute permission="view_mikrotik_servers"><MikrotikView /></PermissionRoute>} />
          <Route
            path="/network-management/olt"
            element={
              <PermissionRoute permission="view_olt">
                <OLTView />
              </PermissionRoute>
            }
          />
          <Route
            path="/odp-management"
            element={
              <PermissionRoute permission="view_odp_management">
                <ODPView />
              </PermissionRoute>
            }
          />
          <Route path="/topology/olt/:olt_id" element={<PlaceholderPage title="Topologi OLT" taskNumber={7} />} />
          <Route path="/traffic-monitoring" element={<PlaceholderPage title="Traffic Monitoring" taskNumber={7} />} />
          <Route path="/monitoring-tv" element={<TVMonitoringView />} />
          <Route path="/inventory" element={<PermissionRoute permission="view_inventory"><InventoryView /></PermissionRoute>} />

          <Route path="/users" element={<PermissionRoute permission="view_users"><UsersView /></PermissionRoute>} />
          <Route path="/roles" element={<PermissionRoute permission="view_roles"><RolesView /></PermissionRoute>} />
          <Route path="/permissions" element={<PermissionRoute permission="view_permissions"><PermissionsView /></PermissionRoute>} />
          <Route path="/syarat-ketentuan" element={<SKView />} />
          <Route
            path="/management/sk"
            element={
              <PermissionRoute permission="manage_sk">
                <SKManagementView />
              </PermissionRoute>
            }
          />
          <Route
            path="/activity-logs"
            element={
              <PermissionRoute permissions={['view_activity_log']}>
                <ActivityLogView />
              </PermissionRoute>
            }
          />
          <Route
            path="/global-search-results"
            element={
              <PermissionRoute
                permissions={['view_pelanggan', 'view_langganan', 'view_activity_log', 'view_trouble_tickets', 'view_invoices']}
              >
                <GlobalSearchResultsView />
              </PermissionRoute>
            }
          />
          <Route path="/management/settings" element={<SystemSettingsView />} />
          <Route path="/management/api-keys" element={<ApiKeysView />} />
          </Route>
        </Route>

        <Route element={<GuestRoute />}>
          <Route path="/login" element={<LoginView />} />
        </Route>

        <Route path="/tv-monitoring" element={<TVMonitoringView />} />
        <Route path="/tv" element={<Navigate to="/tv-monitoring" replace />} />
        <Route path="/privacy-policy" element={<PrivacyPolicyView />} />
        <Route path="/kebijakan-privasi" element={<PrivacyPolicyView />} />
        <Route path="/network-error" element={<NetworkErrorView />} />
        <Route path="/404" element={<NotFoundView />} />
        <Route path="*" element={<NotFoundView />} />
      </Routes>
    </>
  )
}

export function AppRouter() {
  const isElectron = typeof window !== 'undefined' && 'electronAPI' in window
  if (isElectron) {
    return (
      <HashRouter>
        <AppRoutes />
      </HashRouter>
    )
  }
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
