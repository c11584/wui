import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import { useSystemStore } from './stores/system'
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import Dashboard from './pages/Dashboard'
import Tunnels from './pages/Tunnels'
import Settings from './pages/Settings'
import Users from './pages/Users'
import License from './pages/License'
import Store from './pages/Store'
import Orders from './pages/Orders'
import APITokens from './pages/APITokens'
import Packages from './pages/Packages'
import Coupons from './pages/Coupons'
import TrafficCharts from './pages/TrafficCharts'
import ClientDownload from './pages/ClientDownload'
import LicenseManage from './pages/LicenseManage'
import PaymentSettings from './pages/PaymentSettings'
import Layout from './components/Layout'

function App() {
  const { token, user } = useAuthStore()
  const { mode } = useSystemStore()

  if (!token) {
    return (
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/tunnels" element={<Tunnels />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/license" element={<License />} />
        <Route path="/store" element={<Store />} />
        <Route path="/orders" element={<Orders />} />
        <Route path="/api-tokens" element={<APITokens />} />
        <Route path="/traffic" element={<TrafficCharts />} />
        <Route path="/client-download" element={<ClientDownload />} />
        {user?.role === 'admin' && mode === 'admin' && (
          <>
            <Route path="/users" element={<Users />} />
            <Route path="/packages" element={<Packages />} />
            <Route path="/coupons" element={<Coupons />} />
            <Route path="/payment-settings" element={<PaymentSettings />} />
            <Route path="/license-manage" element={<LicenseManage />} />
          </>
        )}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  )
}

export default App
