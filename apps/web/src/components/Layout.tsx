import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'
import { useSystemStore } from '../stores/system'
import { useThemeStore } from '../stores/theme'
import { systemApi } from '../api/system'
import {
  LayoutDashboard, LogOut, Menu, X, Network, Users, Settings,
  Key, ShoppingCart, FileText, Code, Package, Ticket, Activity,
  Download, Sun, Moon, Globe, ChevronDown, KeyRound, CreditCard
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const { t, i18n } = useTranslation()
  const { user, logout } = useAuthStore()
  const { mode, setMode } = useSystemStore()
  const { theme, toggleTheme } = useThemeStore()
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [langMenuOpen, setLangMenuOpen] = useState(false)

  useEffect(() => {
    systemApi.getInfo()
      .then((info) => {
        setMode(info.mode)
      })
      .catch((err) => {
        console.error('Failed to fetch system info:', err)
      })
  }, [setMode])

  const navItems = [
    { path: '/', label: t('dashboard.title'), icon: LayoutDashboard },
    { path: '/tunnels', label: t('tunnel.title'), icon: Network },
    { path: '/traffic', label: t('traffic.title'), icon: Activity },
    { path: '/client-download', label: t('clientDownload.title'), icon: Download },
    { path: '/store', label: t('store.title'), icon: ShoppingCart },
    { path: '/orders', label: t('orders.title'), icon: FileText },
    { path: '/api-tokens', label: t('apiTokens.title'), icon: Code },
    { path: '/settings', label: t('settings.title'), icon: Settings },
    { path: '/license', label: t('license.title'), icon: Key },
  ]

  const adminNavItems = user?.role === 'admin' && mode === 'admin' ? [
    { path: '/users', label: t('users.title'), icon: Users },
    { path: '/packages', label: t('packages.title'), icon: Package },
    { path: '/coupons', label: t('coupons.title'), icon: Ticket },
    { path: '/payment-settings', label: t('paymentSettings.title'), icon: CreditCard },
    { path: '/license-manage', label: t('licenseManage.title'), icon: KeyRound },
  ] : []

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng)
    localStorage.setItem('language', lng)
    setLangMenuOpen(false)
  }

  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--bg-primary)' }}>
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-20 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <aside
        className={`fixed top-0 left-0 z-30 h-full w-56 transform transition-transform duration-200 ease-in-out lg:translate-x-0 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
        style={{ backgroundColor: 'var(--sidebar-bg)', borderRight: '1px solid var(--border-color)' }}
      >
        <div className="flex items-center justify-between h-12 px-4" style={{ borderBottom: '1px solid var(--border-color)' }}>
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg flex items-center justify-center" style={{ backgroundColor: 'var(--accent)' }}>
              <span className="text-white font-bold text-sm">W</span>
            </div>
            <span className="font-semibold text-base" style={{ color: 'var(--text-primary)' }}>WUI</span>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden p-1.5 rounded-md transition-colors"
            style={{ color: 'var(--text-secondary)' }}
          >
            <X size={18} />
          </button>
        </div>

        <nav className="p-2 space-y-0.5 overflow-y-auto" style={{ height: 'calc(100% - 48px)' }}>
          {[...navItems, ...adminNavItems].map((item, index) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            const isAdminItem = index >= navItems.length
            return (
              <div key={item.path}>
                {isAdminItem && index === navItems.length && (
                  <div className="pt-3 pb-1.5 px-3">
                    <span className="text-[10px] uppercase tracking-wider font-medium" style={{ color: 'var(--text-tertiary)' }}>
                      {t('common.admin')}
                    </span>
                  </div>
                )}
                <Link
                  to={item.path}
                  onClick={() => setSidebarOpen(false)}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-all ${
                    isActive ? 'font-medium' : ''
                  }`}
                  style={{
                    backgroundColor: isActive ? 'var(--accent)' : 'transparent',
                    color: isActive ? '#ffffff' : 'var(--text-secondary)',
                  }}
                >
                  <Icon size={16} />
                  <span>{item.label}</span>
                </Link>
              </div>
            )
          })}
        </nav>
      </aside>

      <div className="lg:ml-56">
        <header
          className="h-12 flex items-center px-4 justify-between sticky top-0 z-10"
          style={{ backgroundColor: 'var(--header-bg)', borderBottom: '1px solid var(--border-color)' }}
        >
          <button
            onClick={() => setSidebarOpen(true)}
            className="lg:hidden p-1.5 rounded-md transition-colors"
            style={{ color: 'var(--text-secondary)' }}
          >
            <Menu size={18} />
          </button>

          <div className="flex items-center gap-1 ml-auto">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-md transition-colors hover:bg-opacity-80"
              style={{ color: 'var(--text-secondary)' }}
              title={theme === 'dark' ? t('common.lightMode') : t('common.darkMode')}
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
            </button>

            <div className="relative">
              <button
                onClick={() => setLangMenuOpen(!langMenuOpen)}
                className="flex items-center gap-1 p-2 rounded-md transition-colors hover:bg-opacity-80"
                style={{ color: 'var(--text-secondary)' }}
              >
                <Globe size={16} />
                <ChevronDown size={12} />
              </button>

              {langMenuOpen && (
                <div
                  className="absolute right-0 mt-1 py-1 rounded-md shadow-lg min-w-[80px]"
                  style={{ backgroundColor: 'var(--card-bg)', border: '1px solid var(--border-color)' }}
                >
                  <button
                    onClick={() => changeLanguage('zh')}
                    className={`w-full px-3 py-1.5 text-left text-sm transition-colors ${
                      i18n.language === 'zh' ? 'font-medium' : ''
                    }`}
                    style={{ color: i18n.language === 'zh' ? 'var(--accent)' : 'var(--text-secondary)' }}
                  >
                    中文
                  </button>
                  <button
                    onClick={() => changeLanguage('en')}
                    className={`w-full px-3 py-1.5 text-left text-sm transition-colors ${
                      i18n.language === 'en' ? 'font-medium' : ''
                    }`}
                    style={{ color: i18n.language === 'en' ? 'var(--accent)' : 'var(--text-secondary)' }}
                  >
                    EN
                  </button>
                </div>
              )}
            </div>

            <div className="h-4 w-px mx-1" style={{ backgroundColor: 'var(--border-color)' }} />

            <span className="text-sm px-2" style={{ color: 'var(--text-secondary)' }}>
              {user?.username}
            </span>

            <button
              onClick={logout}
              className="p-2 rounded-md transition-colors hover:bg-opacity-80"
              style={{ color: 'var(--text-secondary)' }}
              title={t('common.logout')}
            >
              <LogOut size={16} />
            </button>
          </div>
        </header>

        <main className="p-4">{children}</main>
      </div>
    </div>
  )
}
