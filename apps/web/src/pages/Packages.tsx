import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Table, Button, Modal, Form, Input, InputNumber, Switch, Space, message, Popconfirm, Card, Select } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'

interface Package {
  id: number
  createdAt: string
  name: string
  description: string
  price: number
  currency: string
  duration: number
  maxTunnels: number
  maxTraffic: number
  maxSpeed: number
  features: string
  isPopular: boolean
  enabled: boolean
  sortOrder: number
}

export default function Packages() {
  const { t } = useTranslation()
  const [packages, setPackages] = useState<Package[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingPackage, setEditingPackage] = useState<Package | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchPackages()
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

  const fetchPackages = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/commerce/packages', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setPackages(result.data || [])
    } catch (error) {
      message.error(t('packages.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const GB_TO_BYTES = 1024 * 1024 * 1024

  const handleCreate = async (values: any) => {
    try {
      const response = await fetch('/api/commerce/packages', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          ...values,
          maxTraffic: values.maxTraffic ? values.maxTraffic * GB_TO_BYTES : 0
        }),
      })

      if (response.ok) {
        message.success(t('packages.createSuccess'))
        setModalVisible(false)
        form.resetFields()
        fetchPackages()
      } else {
        const data = await response.json()
        message.error(data.error || t('packages.createFailed'))
      }
    } catch (error) {
      message.error(t('packages.createFailed'))
    }
  }

  const handleUpdate = async (values: any) => {
    if (!editingPackage) return
    try {
      const response = await fetch(`/api/commerce/packages/${editingPackage.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          ...values,
          maxTraffic: values.maxTraffic ? values.maxTraffic * GB_TO_BYTES : 0
        }),
      })

      if (response.ok) {
        message.success(t('packages.updateSuccess'))
        setModalVisible(false)
        form.resetFields()
        fetchPackages()
      } else {
        const data = await response.json()
        message.error(data.error || t('packages.updateFailed'))
      }
    } catch (error) {
      message.error(t('packages.updateFailed'))
    }
  }

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/commerce/packages/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })

      if (response.ok) {
        message.success(t('packages.deleteSuccess'))
        fetchPackages()
      } else {
        message.error(t('packages.deleteError'))
      }
    } catch (error) {
      message.error(t('packages.deleteError'))
    }
  }

  const openCreateModal = () => {
    form.setFieldsValue({
      name: '',
      description: '',
      price: 0,
      currency: 'CNY',
      duration: 30,
      maxTunnels: 5,
      maxTraffic: 100,
      maxSpeed: 0,
      features: '',
      isPopular: false,
      enabled: true,
      sortOrder: 0
    })
    setEditingPackage(null)
    setModalVisible(true)
  }

  const openEditModal = (pkg: Package) => {
    form.setFieldsValue({
      ...pkg,
      features: pkg.features || '',
      maxTraffic: pkg.maxTraffic ? Math.round(pkg.maxTraffic / GB_TO_BYTES) : 0
    })
    setEditingPackage(pkg)
    setModalVisible(true)
  }

  const columns = [
    {
      title: t('packages.name'),
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: t('packages.description'),
      dataIndex: 'description',
      key: 'description',
      width: 200,
    },
    {
      title: t('packages.price'),
      dataIndex: 'price',
      key: 'price',
      width: 120,
      render: (v: number, record: Package) => (
        <span style={{ color: '#22c55e', fontWeight: 600 }}>
          {record.currency} {v.toFixed(2)}
        </span>
      ),
    },
    {
      title: t('packages.durationDays'),
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (v: number) => `${v} ${t('store.days')}`,
    },
    {
      title: t('packages.maxTunnels'),
      dataIndex: 'maxTunnels',
      key: 'maxTunnels',
      width: 100,
    },
    {
      title: t('packages.maxTrafficGB'),
      dataIndex: 'maxTraffic',
      key: 'maxTraffic',
      width: 120,
      render: (v: number) => `${(v / 1073741824).toFixed(0)} GB`,
    },
    {
      title: t('packages.maxSpeedMbps'),
      dataIndex: 'maxSpeed',
      key: 'maxSpeed',
      width: 100,
      render: (v: number) => v > 0 ? `${v} Mbps` : t('common.unlimited'),
    },
    {
      title: t('packages.isPopular'),
      dataIndex: 'isPopular',
      key: 'isPopular',
      width: 80,
      render: (v: boolean) => (
        <Switch checked={v} disabled size="small" />
      ),
    },
    {
      title: t('packages.enabled'),
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (v: boolean) => (
        <Switch checked={v} disabled size="small" />
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 150,
      render: (_: any, record: Package) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('packages.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Button
              type="link"
              danger
              size="small"
              icon={<DeleteOutlined />}
            >
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card style={{ marginBottom: '24px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 500, color: 'var(--text-primary)' }}>{t('packages.title')}</h2>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={openCreateModal}
          >
            {t('packages.createPackage')}
          </Button>
        </div>

        <Table
          dataSource={packages}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title={editingPackage ? t('packages.editPackage') : t('packages.createPackage')}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          form.resetFields()
          setEditingPackage(null)
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={editingPackage ? handleUpdate : handleCreate}
        >
          <Form.Item name="name" label={t('packages.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label={t('packages.description')}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space style={{ width: '100%' }} size="large">
            <Form.Item name="price" label={t('packages.price')} rules={[{ required: true }]}>
              <InputNumber min={0} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item name="currency" label={t('packages.currency')}>
              <Select style={{ width: 100 }}>
                <Select.Option value="CNY">CNY</Select.Option>
                <Select.Option value="USD">USD</Select.Option>
              </Select>
            </Form.Item>
          </Space>
          <Form.Item name="duration" label={t('packages.durationDays')} rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Space style={{ width: '100%' }} size="large">
            <Form.Item name="maxTunnels" label={t('packages.maxTunnels')}>
              <InputNumber min={0} style={{ width: 100 }} />
            </Form.Item>
            <Form.Item name="maxTraffic" label={t('packages.maxTrafficGB')}>
              <InputNumber min={0} style={{ width: 120 }} />
            </Form.Item>
            <Form.Item name="maxSpeed" label={t('packages.maxSpeedMbps')}>
              <InputNumber min={0} style={{ width: 100 }} />
            </Form.Item>
          </Space>
          <Form.Item name="features" label={t('packages.features')}>
            <Input.TextArea rows={3} placeholder={t('packages.featuresPlaceholder')} />
          </Form.Item>
          <Space style={{ width: '100%' }} size="large">
            <Form.Item name="isPopular" valuePropName="checked" label={t('packages.isPopular')}>
              <Switch />
            </Form.Item>
            <Form.Item name="enabled" valuePropName="checked" label={t('packages.enabled')}>
              <Switch />
            </Form.Item>
          </Space>
          <Form.Item name="sortOrder" label={t('packages.sortOrder')}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingPackage ? t('common.update') : t('common.create')}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
