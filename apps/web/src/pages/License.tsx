import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Button, Input, Form, message, Descriptions, Tag, Alert, Spin } from 'antd'
import { KeyOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'

interface LicenseInfo {
  isValid: boolean
  message: string
  type: string
  plan: string
  maxTunnels: number
  maxUsers: number
  maxTraffic: number
  features: string
  expiresAt: string
}

export default function License() {
  const { t } = useTranslation()
  const [licenseInfo, setLicenseInfo] = useState<LicenseInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [activating, setActivating] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchLicenseInfo()
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

  const fetchLicenseInfo = async () => {
    try {
      const response = await fetch('/api/license/info', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      
      if (result.data) {
        setLicenseInfo(result.data)
      }
    } catch (error) {
      console.error('Failed to fetch license info:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleActivate = async (values: { licenseKey: string }) => {
    setActivating(true)
    try {
      const response = await fetch('/api/license/activate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          licenseKey: values.licenseKey,
          instanceId: getInstanceId()
        })
      })

      const result = await response.json()

      if (result.valid || result.success) {
        message.success(t('license.activateSuccess'))
        setLicenseInfo(result.data)
        form.resetFields()
      } else {
        message.error(result.message || t('license.activateError'))
      }
    } catch (error) {
      message.error(t('license.activateError'))
    } finally {
      setActivating(false)
    }
  }

  const getInstanceId = () => {
    let instanceId = localStorage.getItem('wui-instance-id')
    if (!instanceId) {
      instanceId = 'wui-' + Math.random().toString(36).substr(2, 9)
      localStorage.setItem('wui-instance-id', instanceId)
    }
    return instanceId
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return 'Unlimited'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin />
      </div>
    )
  }

  return (
    <div className="space-y-6" style={{ padding: '24px' }}>
      <h1 className="text-2xl font-bold" style={{ color: 'var(--text-primary)' }}>{t('license.title')}</h1>

      {licenseInfo && licenseInfo.isValid ? (
        <Card style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
          <div className="flex items-center mb-4">
            <CheckCircleOutlined className="text-2xl mr-2" style={{ color: '#22c55e' }} />
            <span className="text-lg font-medium" style={{ color: '#22c55e' }}>{t('license.active')}</span>
          </div>

          <Descriptions column={2} styles={{ label: { color: 'var(--text-tertiary)' }, content: { color: 'var(--text-primary)' } }}>
            <Descriptions.Item label={t('license.type')}>
              <Tag color="blue">{licenseInfo.type}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('license.plan')}>
              <Tag color="gold">{licenseInfo.plan}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('license.maxTunnels')}>
              {licenseInfo.maxTunnels || 'Unlimited'}
            </Descriptions.Item>
            <Descriptions.Item label={t('license.maxUsers')}>
              {licenseInfo.maxUsers || 'Unlimited'}
            </Descriptions.Item>
            <Descriptions.Item label={t('license.maxTraffic')}>
              {formatBytes(licenseInfo.maxTraffic)}
            </Descriptions.Item>
            <Descriptions.Item label={t('license.expiresAt')}>
              {licenseInfo.expiresAt ? new Date(licenseInfo.expiresAt).toLocaleDateString() : t('license.lifetime')}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      ) : (
        <>
          <Alert
            message={t('license.noLicense')}
            description={t('license.noLicenseDesc')}
            type="warning"
            showIcon
            icon={<CloseCircleOutlined />}
          />

          <Card title={t('license.activateLicense')} style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
            <Form form={form} onFinish={handleActivate} layout="vertical">
              <Form.Item
                name="licenseKey"
                label={t('license.licenseKey')}
                rules={[{ required: true }]}
                extra={t('license.licenseKeyFormat')}
              >
                <Input placeholder="WUI-XXXX-XXXX-XXXX" prefix={<KeyOutlined />} />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={activating}>
                  {t('license.activate')}
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </>
      )}

      <Card title={t('license.howToGet')} style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
        <div className="space-y-4" style={{ color: 'var(--text-secondary)' }}>
          <p>{t('license.howToGetDesc')}</p>
          <ul className="list-disc list-inside space-y-2">
            <li>{t('license.visitWebsite')}</li>
            <li>{t('license.contactSales')}</li>
            <li>{t('license.freeTrial')}</li>
          </ul>
        </div>
      </Card>
    </div>
  )
}
