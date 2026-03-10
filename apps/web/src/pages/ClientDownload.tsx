import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Button, Table, Tag, Space, message, Modal, Typography } from 'antd'
import { DownloadOutlined, AppleOutlined, WindowsOutlined, AndroidOutlined, GlobalOutlined, CopyOutlined } from '@ant-design/icons'

const { Paragraph } = Typography

interface ClientDownload {
  name: string
  platform: string
  version: string
  size: string
  downloadUrl: string
  sha256: string
  icon: React.ReactNode
}

interface SubscriptionToken {
  token: string
  clashUrl: string
  v2rayUrl: string
}

export default function ClientDownload() {
  const { t } = useTranslation()
  const [selectedClient, setSelectedClient] = useState<ClientDownload | null>(null)
  const [modalVisible, setModalVisible] = useState(false)
  const [subscriptionToken, setSubscriptionToken] = useState<SubscriptionToken | null>(null)

  useEffect(() => {
    fetchSubscriptionToken()
  }, [])

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

  const fetchSubscriptionToken = async () => {
    try {
      const response = await fetch('/api/subscription/token', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.success) {
        setSubscriptionToken(result.data)
      }
    } catch (error) {
      console.error('Failed to fetch subscription token:', error)
    }
  }

  const clients: ClientDownload[] = [
    {
      name: 'v2rayN',
      platform: 'Windows',
      version: '6.42',
      size: '28.5 MB',
      downloadUrl: 'https://github.com/2dust/v2rayN/releases/download/6.42/v2rayN-With-Core.zip',
      sha256: 'abc123...',
      icon: <WindowsOutlined className="text-3xl text-blue-400" />,
    },
    {
      name: 'v2rayNG',
      platform: 'Android',
      version: '1.8.12',
      size: '18.2 MB',
      downloadUrl: 'https://github.com/2dust/v2rayNG/releases/download/1.8.12/v2rayNG_1.8.12_arm64-v8a.apk',
      sha256: 'def456...',
      icon: <AndroidOutlined className="text-3xl text-green-400" />,
    },
    {
      name: 'Clash Verge',
      platform: 'Windows / macOS / Linux',
      version: '1.3.8',
      size: '42.1 MB',
      downloadUrl: 'https://github.com/zzzgydi/clash-verge/releases/download/v1.3.8/clash-verge_1.3.8_x64-setup.exe',
      sha256: 'ghi789...',
      icon: <GlobalOutlined className="text-3xl text-purple-400" />,
    },
    {
      name: 'ClashX Pro',
      platform: 'macOS',
      version: '1.118.0',
      size: '15.8 MB',
      downloadUrl: 'https://github.com/yichengchen/clashX/releases/download/1.118.0/ClashX.Pro.dmg',
      sha256: 'jkl012...',
      icon: <AppleOutlined className="text-3xl text-gray-400" />,
    },
    {
      name: 'Qv2ray',
      platform: 'Windows / macOS / Linux',
      version: '2.7.0',
      size: '35.6 MB',
      downloadUrl: 'https://github.com/Qv2ray/Qv2ray/releases/download/v2.7.0/Qv2ray.v2.7.0.Windows-x64.exe',
      sha256: 'mno345...',
      icon: <GlobalOutlined className="text-3xl text-cyan-400" />,
    },
    {
      name: 'Shadowrocket',
      platform: 'iOS',
      version: '2.2.8',
      size: '12.3 MB',
      downloadUrl: 'https://apps.apple.com/app/shadowrocket/id932747118',
      sha256: '-',
      icon: <AppleOutlined className="text-3xl text-orange-400" />,
    },
  ]

  const handleDownload = (client: ClientDownload) => {
    setSelectedClient(client)
    setModalVisible(true)
  }

  const copySubscriptionUrl = async (format: string = 'clash') => {
    if (!subscriptionToken) {
      message.error(t('clientDownload.copyFailed'))
      return
    }

    try {
      let urlPath: string
      switch (format) {
        case 'clash':
          urlPath = subscriptionToken.clashUrl
          break
        case 'v2ray':
        case 'vmess':
          urlPath = subscriptionToken.v2rayUrl
          break
        default:
          urlPath = subscriptionToken.clashUrl
      }
      
      const url = `${window.location.origin}${urlPath}`

      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(url)
        message.success(t('clientDownload.copied'))
        return
      }

      const textArea = document.createElement('textarea')
      textArea.value = url
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()

      const successful = document.execCommand('copy')
      document.body.removeChild(textArea)

      if (successful) {
        message.success(t('clientDownload.copied'))
      } else {
        message.error(t('clientDownload.copyFailed'))
      }
    } catch (e) {
      console.error('Copy failed:', e)
      message.error(t('clientDownload.copyFailed'))
    }
  }

  const columns = [
    {
      title: t('clientDownload.client'),
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: ClientDownload) => (
        <div className="flex items-center space-x-3">
          {record.icon}
          <div>
            <div className="text-white font-medium">{name}</div>
            <div className="text-gray-400 text-sm">{record.platform}</div>
          </div>
        </div>
      ),
    },
    {
      title: t('clientDownload.version'),
      dataIndex: 'version',
      key: 'version',
      render: (version: string) => <Tag color="blue">{version}</Tag>,
    },
    {
      title: t('clientDownload.size'),
      dataIndex: 'size',
      key: 'size',
      render: (size: string) => <span className="text-gray-300">{size}</span>,
    },
    {
      title: t('clientDownload.actions'),
      key: 'actions',
      render: (_: any, record: ClientDownload) => (
        <Button
          type="primary"
          icon={<DownloadOutlined />}
          onClick={() => handleDownload(record)}
        >
          {t('clientDownload.download')}
        </Button>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">{t('clientDownload.title')}</h1>
      </div>

      <Card className="bg-gray-800 border-gray-700">
        <div className="mb-4">
          <h3 className="text-lg font-medium text-white mb-2">{t('clientDownload.quickSetup')}</h3>
          <p className="text-gray-400 mb-4">{t('clientDownload.quickSetupDesc')}</p>
          <Space>
            <Button type="primary" icon={<CopyOutlined />} onClick={() => copySubscriptionUrl()}>
              {t('clientDownload.copySubscriptionUrl')}
            </Button>
          </Space>
        </div>
      </Card>

      <Card className="bg-gray-800 border-gray-700">
        <Table
          dataSource={clients}
          columns={columns}
          rowKey="name"
          pagination={false}
        />
      </Card>

      <Card className="bg-gray-800 border-gray-700" title={t('clientDownload.tutorial')}>
        <div className="space-y-4">
          <div className="p-4 bg-gray-700 rounded">
            <h4 className="text-white font-medium mb-2">1. {t('clientDownload.step1')}</h4>
            <p className="text-gray-400 text-sm">{t('clientDownload.step1Desc')}</p>
          </div>
          <div className="p-4 bg-gray-700 rounded">
            <h4 className="text-white font-medium mb-2">2. {t('clientDownload.step2')}</h4>
            <p className="text-gray-400 text-sm">{t('clientDownload.step2Desc')}</p>
          </div>
          <div className="p-4 bg-gray-700 rounded">
            <h4 className="text-white font-medium mb-2">3. {t('clientDownload.step3')}</h4>
            <p className="text-gray-400 text-sm">{t('clientDownload.step3Desc')}</p>
          </div>
        </div>
      </Card>

      <Modal
        title={t('clientDownload.downloadInfo')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setModalVisible(false)}>
            {t('common.close')}
          </Button>,
          <Button
            key="download"
            type="primary"
            icon={<DownloadOutlined />}
            onClick={() => {
              if (selectedClient) {
                window.open(selectedClient.downloadUrl, '_blank')
              }
            }}
          >
            {t('clientDownload.goToDownload')}
          </Button>,
        ]}
      >
        {selectedClient && (
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              {selectedClient.icon}
              <div>
                <h3 className="text-lg font-medium text-white">{selectedClient.name}</h3>
                <p className="text-gray-400">{selectedClient.platform}</p>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-gray-400">{t('clientDownload.version')}: </span>
                <span className="text-white">{selectedClient.version}</span>
              </div>
              <div>
                <span className="text-gray-400">{t('clientDownload.size')}: </span>
                <span className="text-white">{selectedClient.size}</span>
              </div>
            </div>
            <div>
              <span className="text-gray-400">SHA256: </span>
              <Paragraph copyable className="text-white text-sm mb-0">
                {selectedClient.sha256}
              </Paragraph>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
