import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { 
  Card, 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  InputNumber, 
  Select, 
  DatePicker, 
  Space, 
  Tag, 
  message,
  Popconfirm
} from 'antd'
import { 
  PlusOutlined, 
  DeleteOutlined, 
  CopyOutlined, 
  ReloadOutlined,
  KeyOutlined,
  EditOutlined 
} from '@ant-design/icons'
import dayjs from 'dayjs'

interface LicenseKey {
  id: number
  licenseKey: string
  type: string
  plan: string
  maxTunnels: number
  maxUsers: number
  maxTraffic: number
  features: string
  expiresAt: string | null
  status: string
  usedBy: number | null
  usedAt: string | null
  createdAt: string
  user?: { username: string; email: string }
}

export default function LicenseManage() {
  const { t } = useTranslation()
  const [licenses, setLicenses] = useState<LicenseKey[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editingLicense, setEditingLicense] = useState<LicenseKey | null>(null)
  const [generatedLicense, setGeneratedLicense] = useState<string | null>(null)
  const [form] = Form.useForm()
  const [editForm] = Form.useForm()

  useEffect(() => {
    fetchLicenses()
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

  const fetchLicenses = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/licenses', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setLicenses(result.data || [])
    } catch (error) {
      message.error(t('licenseManage.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const handleGenerateLicense = async (values: any) => {
    try {
      const response = await fetch('/api/licenses', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          ...values,
          expiresAt: values.expiresAt ? values.expiresAt.toISOString() : null,
          maxTraffic: values.maxTraffic ? values.maxTraffic * 1024 * 1024 * 1024 : 0
        })
      })

      const result = await response.json()
      if (response.ok) {
        setGeneratedLicense(result.data.licenseKey)
        message.success(t('licenseManage.generateSuccess'))
        fetchLicenses()
      } else {
        message.error(result.error || t('licenseManage.generateError'))
      }
    } catch (error) {
      message.error(t('licenseManage.generateError'))
    }
  }

  const handleDeleteLicense = async (id: number) => {
    try {
      const response = await fetch(`/api/licenses/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })

      if (response.ok) {
        message.success(t('licenseManage.deleteSuccess'))
        fetchLicenses()
      } else {
        message.error(t('licenseManage.deleteError'))
      }
    } catch (error) {
      message.error(t('licenseManage.deleteError'))
    }
  }

  const handleRevokeLicense = async (id: number) => {
    try {
      const response = await fetch(`/api/licenses/${id}/revoke`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })

      if (response.ok) {
        message.success(t('licenseManage.revokeSuccess'))
        fetchLicenses()
      } else {
        message.error(t('licenseManage.revokeError'))
      }
    } catch (error) {
      message.error(t('licenseManage.revokeError'))
    }
  }

  const openEditModal = (license: LicenseKey) => {
    setEditingLicense(license)
    editForm.setFieldsValue({
      type: license.type,
      plan: license.plan,
      maxTunnels: license.maxTunnels,
      maxUsers: license.maxUsers,
      maxTraffic: license.maxTraffic ? Math.round(license.maxTraffic / (1024 * 1024 * 1024)) : 0,
      features: license.features || '',
      expiresAt: license.expiresAt ? dayjs(license.expiresAt) : null
    })
    setEditModalVisible(true)
  }

  const handleUpdateLicense = async (values: any) => {
    if (!editingLicense) return
    try {
      const response = await fetch(`/api/licenses/${editingLicense.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          type: values.type,
          plan: values.plan,
          maxTunnels: values.maxTunnels,
          maxUsers: values.maxUsers,
          maxTraffic: values.maxTraffic ? values.maxTraffic * 1024 * 1024 * 1024 : 0,
          features: values.features || '',
          expiresAt: values.expiresAt ? values.expiresAt.toISOString() : null
        })
      })

      if (response.ok) {
        message.success(t('licenseManage.updateSuccess'))
        setEditModalVisible(false)
        setEditingLicense(null)
        fetchLicenses()
      } else {
        const result = await response.json()
        message.error(result.error || t('licenseManage.updateError'))
      }
    } catch (error) {
      message.error(t('licenseManage.updateError'))
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    message.success(t('common.copied'))
  }

  const generateRandomKey = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    const segments = []
    for (let i = 0; i < 3; i++) {
      let segment = ''
      for (let j = 0; j < 4; j++) {
        segment += chars.charAt(Math.floor(Math.random() * chars.length))
      }
      segments.push(segment)
    }
    return `WUI-${segments.join('-')}`
  }

  const getStatusTag = (status: string) => {
    const colors: Record<string, string> = {
      unused: 'green',
      used: 'blue',
      expired: 'red',
      revoked: 'default'
    }
    const labels: Record<string, string> = {
      unused: t('licenseManage.statusUnused'),
      used: t('licenseManage.statusUsed'),
      expired: t('licenseManage.statusExpired'),
      revoked: t('licenseManage.statusRevoked')
    }
    return <Tag color={colors[status] || 'default'}>{labels[status] || status}</Tag>
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return t('licenseManage.unlimited')
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const columns = [
    {
      title: t('licenseManage.licenseKey'),
      dataIndex: 'licenseKey',
      key: 'licenseKey',
      render: (key: string) => (
        <Space>
          <code style={{ fontSize: '12px', color: 'var(--text-primary)' }}>{key}</code>
          <Button 
            type="link" 
            size="small" 
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(key)}
          />
        </Space>
      )
    },
    {
      title: t('licenseManage.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="blue">{type}</Tag>
    },
    {
      title: t('licenseManage.plan'),
      dataIndex: 'plan',
      key: 'plan',
      render: (plan: string) => <Tag color="gold">{plan}</Tag>
    },
    {
      title: t('licenseManage.maxTunnels'),
      dataIndex: 'maxTunnels',
      key: 'maxTunnels',
      render: (v: number) => v === 0 ? t('licenseManage.unlimited') : v
    },
    {
      title: t('licenseManage.maxUsers'),
      dataIndex: 'maxUsers',
      key: 'maxUsers',
      render: (v: number) => v === 0 ? t('licenseManage.unlimited') : v
    },
    {
      title: t('licenseManage.maxTraffic'),
      dataIndex: 'maxTraffic',
      key: 'maxTraffic',
      render: (v: number) => formatBytes(v)
    },
    {
      title: t('licenseManage.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status)
    },
    {
      title: t('licenseManage.usedBy'),
      key: 'usedBy',
      render: (_: any, record: LicenseKey) => {
        if (record.usedBy) {
          return <span style={{ color: 'var(--text-primary)' }}>{record.user?.username || record.usedBy}</span>
        }
        return <span style={{ color: 'var(--text-tertiary)' }}>-</span>
      }
    },
    {
      title: t('licenseManage.expiresAt'),
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD') : t('licenseManage.lifetime')
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_: any, record: LicenseKey) => (
        <Space>
          <Button 
            type="link" 
            size="small" 
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            {t('common.edit')}
          </Button>
          {record.status === 'used' && (
            <Popconfirm
              title={t('licenseManage.revokeConfirm')}
              onConfirm={() => handleRevokeLicense(record.id)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Button type="link" size="small" danger>
                {t('licenseManage.revoke')}
              </Button>
            </Popconfirm>
          )}
          {record.status === 'unused' && (
            <Popconfirm
              title={t('licenseManage.deleteConfirm')}
              onConfirm={() => handleDeleteLicense(record.id)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                {t('common.delete')}
              </Button>
            </Popconfirm>
          )}
        </Space>
      )
    }
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card 
        style={{ marginBottom: '24px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchLicenses}>
              {t('common.reload')}
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => {
              form.setFieldsValue({
                licenseKey: generateRandomKey(),
                type: 'standard',
                plan: 'basic',
                maxTunnels: 5,
                maxUsers: 1,
                maxTraffic: 100
              })
              setGeneratedLicense(null)
              setModalVisible(true)
            }}>
              {t('licenseManage.generateLicense')}
            </Button>
          </Space>
        }
      >
        <div style={{ marginBottom: '16px' }}>
          <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 500, color: 'var(--text-primary)' }}>
            {t('licenseManage.title')}
          </h2>
          <p style={{ margin: '8px 0 0 0', color: 'var(--text-tertiary)', fontSize: '14px' }}>
            {t('licenseManage.subtitle')}
          </p>
        </div>

        <Table
          dataSource={licenses}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 20 }}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title={t('licenseManage.generateLicense')}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          setGeneratedLicense(null)
        }}
        footer={null}
        width={600}
      >
        {generatedLicense ? (
          <div style={{ textAlign: 'center', padding: '24px' }}>
            <KeyOutlined style={{ fontSize: '48px', color: '#22c55e', marginBottom: '16px' }} />
            <h3 style={{ color: 'var(--text-primary)', marginBottom: '16px' }}>{t('licenseManage.generateSuccess')}</h3>
            <div style={{ 
              padding: '16px', 
              backgroundColor: 'var(--bg-tertiary)', 
              borderRadius: '8px',
              marginBottom: '16px'
            }}>
              <code style={{ fontSize: '18px', color: 'var(--accent)' }}>{generatedLicense}</code>
            </div>
            <Button 
              type="primary" 
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(generatedLicense)}
            >
              {t('common.copy')}
            </Button>
            <p style={{ marginTop: '16px', color: '#f59e0b', fontSize: '14px' }}>
              {t('licenseManage.copyWarning')}
            </p>
          </div>
        ) : (
          <Form form={form} onFinish={handleGenerateLicense} layout="vertical">
            <Form.Item name="licenseKey" label={t('licenseManage.licenseKey')} rules={[{ required: true }]}>
              <Input 
                addonAfter={
                  <Button size="small" onClick={() => form.setFieldsValue({ licenseKey: generateRandomKey() })}>
                    {t('licenseManage.random')}
                  </Button>
                }
              />
            </Form.Item>
            <Space style={{ width: '100%' }}>
              <Form.Item name="type" label={t('licenseManage.type')} style={{ width: '150px' }}>
                <Select>
                  <Select.Option value="trial">{t('licenseManage.typeTrial')}</Select.Option>
                  <Select.Option value="standard">{t('licenseManage.typeStandard')}</Select.Option>
                  <Select.Option value="professional">{t('licenseManage.typeProfessional')}</Select.Option>
                  <Select.Option value="enterprise">{t('licenseManage.typeEnterprise')}</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item name="plan" label={t('licenseManage.plan')} style={{ width: '150px' }}>
                <Select>
                  <Select.Option value="basic">Basic</Select.Option>
                  <Select.Option value="pro">Pro</Select.Option>
                  <Select.Option value="ultimate">Ultimate</Select.Option>
                </Select>
              </Form.Item>
            </Space>
            <Space style={{ width: '100%' }}>
              <Form.Item name="maxTunnels" label={t('licenseManage.maxTunnels')} style={{ width: '150px' }}>
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="maxUsers" label={t('licenseManage.maxUsers')} style={{ width: '150px' }}>
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item name="maxTraffic" label={t('licenseManage.maxTrafficGB')} style={{ width: '150px' }}>
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Space>
            <Form.Item name="expiresAt" label={t('licenseManage.expiresAt')}>
              <DatePicker style={{ width: '100%' }} placeholder={t('licenseManage.expiresPlaceholder')} />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit">
                  {t('licenseManage.generate')}
                </Button>
                <Button onClick={() => setModalVisible(false)}>
                  {t('common.cancel')}
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title={t('licenseManage.editLicense')}
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false)
          setEditingLicense(null)
        }}
        footer={null}
        width={600}
      >
        <Form form={editForm} onFinish={handleUpdateLicense} layout="vertical">
          <Space style={{ width: '100%' }}>
            <Form.Item name="type" label={t('licenseManage.type')} style={{ width: '150px' }}>
              <Select>
                <Select.Option value="trial">{t('licenseManage.typeTrial')}</Select.Option>
                <Select.Option value="standard">{t('licenseManage.typeStandard')}</Select.Option>
                <Select.Option value="professional">{t('licenseManage.typeProfessional')}</Select.Option>
                <Select.Option value="enterprise">{t('licenseManage.typeEnterprise')}</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="plan" label={t('licenseManage.plan')} style={{ width: '150px' }}>
              <Select>
                <Select.Option value="basic">Basic</Select.Option>
                <Select.Option value="pro">Pro</Select.Option>
                <Select.Option value="ultimate">Ultimate</Select.Option>
              </Select>
            </Form.Item>
          </Space>
          <Space style={{ width: '100%' }}>
            <Form.Item name="maxTunnels" label={t('licenseManage.maxTunnels')} style={{ width: '150px' }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="maxUsers" label={t('licenseManage.maxUsers')} style={{ width: '150px' }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="maxTraffic" label={t('licenseManage.maxTrafficGB')} style={{ width: '150px' }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Space>
          <Form.Item name="features" label={t('licenseManage.features')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="expiresAt" label={t('licenseManage.expiresAt')}>
            <DatePicker style={{ width: '100%' }} placeholder={t('licenseManage.expiresPlaceholder')} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {t('common.update')}
              </Button>
              <Button onClick={() => setEditModalVisible(false)}>
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
