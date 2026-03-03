import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { tunnelApi } from '../api/tunnels'
import type { Tunnel } from '../types'
import { Activity, Upload, Download, TrendingUp } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export default function Dashboard() {
  const { t } = useTranslation()
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchTunnels()
    const interval = setInterval(fetchTunnels, 5000)
    return () => clearInterval(interval)
  }, [])

  const fetchTunnels = async () => {
    try {
      const data = await tunnelApi.list()
      setTunnels(data)
    } catch (error) {
      console.error('Failed to fetch tunnels:', error)
    } finally {
      setLoading(false)
    }
  }

  const totalUpload = tunnels.reduce((sum, t) => sum + t.uploadBytes, 0)
  const totalDownload = tunnels.reduce((sum, t) => sum + t.downloadBytes, 0)
  const totalConnections = tunnels.reduce((sum, t) => sum + t.connections, 0)
  const activeTunnels = tunnels.filter((t) => t.enabled).length

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">{t('dashboard.title')}</h1>
        <p className="text-gray-400 mt-1">{t('dashboard.subtitle')}</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">{t('dashboard.activeTunnels')}</p>
              <p className="text-3xl font-bold text-white">{activeTunnels}</p>
            </div>
            <div className="p-3 bg-blue-500 bg-opacity-20 rounded-lg">
              <Activity className="text-blue-500" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">{t('dashboard.totalUpload')}</p>
              <p className="text-3xl font-bold text-white">{formatBytes(totalUpload)}</p>
            </div>
            <div className="p-3 bg-green-500 bg-opacity-20 rounded-lg">
              <Upload className="text-green-500" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">{t('dashboard.totalDownload')}</p>
              <p className="text-3xl font-bold text-white">{formatBytes(totalDownload)}</p>
            </div>
            <div className="p-3 bg-purple-500 bg-opacity-20 rounded-lg">
              <Download className="text-purple-500" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm mb-1">{t('dashboard.connections')}</p>
              <p className="text-3xl font-bold text-white">{totalConnections}</p>
            </div>
            <div className="p-3 bg-yellow-500 bg-opacity-20 rounded-lg">
              <TrendingUp className="text-yellow-500" size={24} />
            </div>
          </div>
        </div>
      </div>

      <div className="mb-8">
        <Link
          to="/tunnels"
          className="inline-flex items-center px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
        >
          {t('dashboard.manageTunnels')}
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg border border-gray-700">
        <div className="px-6 py-4 border-b border-gray-700">
          <h2 className="text-lg font-semibold text-white">{t('dashboard.recentTunnels')}</h2>
        </div>
        {loading ? (
          <div className="p-12 text-center text-gray-400">{t('common.loading')}</div>
        ) : tunnels.length === 0 ? (
          <div className="p-12 text-center">
            <p className="text-gray-400 mb-4">{t('dashboard.noTunnels')}</p>
            <Link
              to="/tunnels"
              className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              {t('dashboard.createFirst')}
            </Link>
          </div>
        ) : (
          <div className="divide-y divide-gray-700">
            {tunnels.slice(0, 5).map((tunnel) => (
              <div key={tunnel.id} className="px-6 py-4 hover:bg-gray-700 transition-colors">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-white font-medium">{tunnel.name}</h3>
                      <span
                        className={`px-2 py-1 text-xs rounded ${
                          tunnel.enabled
                            ? 'bg-green-900 text-green-300'
                            : 'bg-gray-700 text-gray-400'
                        }`}
                      >
                        {tunnel.enabled ? t('tunnel.status.active') : t('tunnel.status.inactive')}
                      </span>
                    </div>
                    <div className="flex gap-4 text-sm text-gray-400">
                      <span>
                        {tunnel.inboundProtocol.toUpperCase()} :{tunnel.inboundPort}
                      </span>
                      <span>•</span>
                      <span>{tunnel.outbounds.length} outbound(s)</span>
                      <span>•</span>
                      <span>UDP: {tunnel.udpEnabled ? '✓' : '✗'}</span>
                      {tunnel.remark && (
                        <>
                          <span>•</span>
                          <span>{tunnel.remark}</span>
                        </>
                      )}
                    </div>
                  </div>
                  <div className="text-right text-sm text-gray-400">
                    <div>↑ {formatBytes(tunnel.uploadBytes)}</div>
                    <div>↓ {formatBytes(tunnel.downloadBytes)}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
