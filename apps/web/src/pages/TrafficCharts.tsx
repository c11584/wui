import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Spin, Select, Row, Col } from 'antd'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, AreaChart, Area } from 'recharts'

interface TrafficData {
  time: string
  upload: number
  download: number
  connections: number
}

interface TunnelStats {
  id: number
  name: string
  upload: number
  download: number
  connections: number
  status: string
}

export default function TrafficCharts() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('24h')
  const [trafficData, setTrafficData] = useState<TrafficData[]>([])
  const [tunnelStats, setTunnelStats] = useState<TunnelStats[]>([])
  const [totalUpload, setTotalUpload] = useState(0)
  const [totalDownload, setTotalDownload] = useState(0)

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [timeRange])

  const getToken = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (authStorage) {
      try {
        const { state } = JSON.parse(authStorage)
        return state?.token || ''
      } catch (e) {
        return ''
      }
    }
    return ''
  }

  const fetchData = async () => {
    try {
      const headers = {
        'Authorization': `Bearer ${getToken()}`
      }

      const [trafficRes, statsRes] = await Promise.all([
        fetch(`/api/system/traffic?range=${timeRange}`, { headers }),
        fetch('/api/tunnels', { headers })
      ])

      if (trafficRes.ok) {
        const data = await trafficRes.json()
        setTrafficData(data.data || [])
        setTotalUpload(data.totalUpload || 0)
        setTotalDownload(data.totalDownload || 0)
      }

      if (statsRes.ok) {
        const tunnels = await statsRes.json()
        setTunnelStats(tunnels.map((t: any) => ({
          id: t.id,
          name: t.name,
          upload: t.upload || 0,
          download: t.download || 0,
          connections: t.connections || 0,
          status: t.status
        })))
      }
    } catch (error) {
      console.error('Failed to fetch traffic data:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatYAxis = (value: number) => {
    return formatBytes(value)
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">{t('traffic.title')}</h1>
        <Select
          value={timeRange}
          onChange={setTimeRange}
          style={{ width: 150 }}
          options={[
            { value: '1h', label: t('traffic.last1h') },
            { value: '6h', label: t('traffic.last6h') },
            { value: '24h', label: t('traffic.last24h') },
            { value: '7d', label: t('traffic.last7d') },
            { value: '30d', label: t('traffic.last30d') },
          ]}
        />
      </div>

      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card className="bg-gray-800 border-gray-700">
            <div className="text-center">
              <div className="text-gray-400 mb-2">{t('traffic.totalUpload')}</div>
              <div className="text-2xl font-bold text-green-400">{formatBytes(totalUpload)}</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card className="bg-gray-800 border-gray-700">
            <div className="text-center">
              <div className="text-gray-400 mb-2">{t('traffic.totalDownload')}</div>
              <div className="text-2xl font-bold text-blue-400">{formatBytes(totalDownload)}</div>
            </div>
          </Card>
        </Col>
        <Col span={8}>
          <Card className="bg-gray-800 border-gray-700">
            <div className="text-center">
              <div className="text-gray-400 mb-2">{t('traffic.totalTraffic')}</div>
              <div className="text-2xl font-bold text-purple-400">{formatBytes(totalUpload + totalDownload)}</div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card className="bg-gray-800 border-gray-700" title={t('traffic.overview')}>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={trafficData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="time" stroke="#9CA3AF" />
            <YAxis tickFormatter={formatYAxis} stroke="#9CA3AF" />
            <Tooltip 
              contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
              formatter={(value) => value !== undefined ? formatBytes(value as number) : ''}
            />
            <Legend />
            <Area 
              type="monotone" 
              dataKey="upload" 
              name={t('traffic.upload')}
              stroke="#10B981" 
              fill="#10B981" 
              fillOpacity={0.3} 
            />
            <Area 
              type="monotone" 
              dataKey="download" 
              name={t('traffic.download')}
              stroke="#3B82F6" 
              fill="#3B82F6" 
              fillOpacity={0.3} 
            />
          </AreaChart>
        </ResponsiveContainer>
      </Card>

      <Card className="bg-gray-800 border-gray-700" title={t('traffic.connections')}>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={trafficData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="time" stroke="#9CA3AF" />
            <YAxis stroke="#9CA3AF" />
            <Tooltip 
              contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
            />
            <Legend />
            <Line 
              type="monotone" 
              dataKey="connections" 
              name={t('traffic.connections')}
              stroke="#F59E0B" 
              strokeWidth={2}
              dot={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Card className="bg-gray-800 border-gray-700" title={t('traffic.byTunnel')}>
        <div className="space-y-4">
          {tunnelStats.map(tunnel => (
            <div key={tunnel.id} className="flex items-center justify-between p-4 bg-gray-700 rounded">
              <div className="flex items-center space-x-4">
                <div className={`w-3 h-3 rounded-full ${tunnel.status === 'running' ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="text-white font-medium">{tunnel.name}</span>
              </div>
              <div className="flex space-x-8 text-sm">
                <div>
                  <span className="text-gray-400">{t('traffic.upload')}: </span>
                  <span className="text-green-400">{formatBytes(tunnel.upload)}</span>
                </div>
                <div>
                  <span className="text-gray-400">{t('traffic.download')}: </span>
                  <span className="text-blue-400">{formatBytes(tunnel.download)}</span>
                </div>
                <div>
                  <span className="text-gray-400">{t('traffic.connections')}: </span>
                  <span className="text-yellow-400">{tunnel.connections}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
