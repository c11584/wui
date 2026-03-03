import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Table, Button, Modal, Form, Input, InputNumber, Switch, Space, message, Popconfirm, Card, Tag, DatePicker } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

interface Coupon {
  id: number
  createdAt: string
  code: string
  discount: number
  isPercent: boolean
  minAmount: number
  maxDiscount: number
  maxUses: number
  usedCount: number
  startTime: string
  endTime: string
  enabled: boolean
}

export default function Coupons() {
  const { t } = useTranslation()
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingCoupon, setEditingCoupon] = useState<Coupon | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchCoupons()
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

  const fetchCoupons = async () => {
    try {
      const response = await fetch('/api/commerce/coupons', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setCoupons(result.data || [])
    } catch (error) {
      message.error(t('coupons.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setEditingCoupon(null)
    form.resetFields()
    form.setFieldsValue({
      discount: 10,
      isPercent: true,
      minAmount: 0,
      maxDiscount: 0,
      maxUses: -1,
      startTime: dayjs().startOf('day'),
      endTime: dayjs().add(3, 'month').endOf('day'),
      enabled: true
    })
    setModalVisible(true)
  }

  const handleEdit = (coupon: Coupon) => {
    setEditingCoupon(coupon)
    form.setFieldsValue({
      code: coupon.code,
      discount: coupon.discount,
      isPercent: coupon.isPercent,
      minAmount: coupon.minAmount,
      maxDiscount: coupon.maxDiscount,
      maxUses: coupon.maxUses,
      startTime: dayjs(coupon.startTime),
      endTime: dayjs(coupon.endTime),
      enabled: coupon.enabled
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/commerce/coupons/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      if (response.ok) {
        message.success(t('coupons.deleteSuccess'))
        fetchCoupons()
      } else {
        message.error(t('coupons.deleteError'))
      }
    } catch (error) {
      message.error(t('coupons.deleteError'))
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      const url = editingCoupon 
        ? `/api/commerce/coupons/${editingCoupon.id}` 
        : '/api/commerce/coupons'
      const method = editingCoupon ? 'PUT' : 'POST'

      const body = {
        code: values.code,
        discount: values.discount,
        isPercent: values.isPercent,
        minAmount: values.minAmount || 0,
        maxDiscount: values.maxDiscount || 0,
        maxUses: values.maxUses || -1,
        startTime: values.startTime.toISOString(),
        endTime: values.endTime.toISOString(),
        enabled: values.enabled
      }

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify(body)
      })

      if (response.ok) {
        message.success(editingCoupon ? t('coupons.updateSuccess') : t('coupons.createSuccess'))
        setModalVisible(false)
        fetchCoupons()
      } else {
        const data = await response.json()
        message.error(data.error || t('coupons.saveError'))
      }
    } catch (error) {
      message.error(t('coupons.saveError'))
    }
  }

  const generateCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    let code = ''
    for (let i = 0; i < 10; i++) {
      code += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    form.setFieldsValue({ code })
  }

  const columns = [
    {
      title: t('coupons.code'),
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => <code style={{ backgroundColor: 'var(--bg-secondary)', padding: '2px 8px', borderRadius: '4px', color: 'var(--text-primary)' }}>{code}</code>
    },
    {
      title: t('coupons.discount'),
      dataIndex: 'discount',
      key: 'discount',
      render: (discount: number, record: Coupon) => (
        <span style={{ color: '#22c55e', fontWeight: 600 }}>
          {record.isPercent ? `${discount}%` : `¥${discount}`}
        </span>
      )
    },
    {
      title: t('coupons.minAmount'),
      dataIndex: 'minAmount',
      key: 'minAmount',
      render: (v: number) => v > 0 ? `¥${v}` : '-'
    },
    {
      title: t('coupons.usage'),
      key: 'usage',
      render: (_: any, record: Coupon) => (
        <span>
          {record.usedCount} / {record.maxUses === -1 ? '∞' : record.maxUses}
        </span>
      )
    },
    {
      title: t('coupons.validPeriod'),
      key: 'validPeriod',
      render: (_: any, record: Coupon) => (
        <span style={{ fontSize: '14px', color: 'var(--text-secondary)' }}>
          {dayjs(record.startTime).format('YYYY-MM-DD')} ~ {dayjs(record.endTime).format('YYYY-MM-DD')}
        </span>
      )
    },
    {
      title: t('coupons.status'),
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? t('common.enabled') : t('common.disabled')}
        </Tag>
      )
    },
    {
      title: t('coupons.actions'),
      key: 'actions',
      render: (_: any, record: Coupon) => (
        <Space>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)} size="small">
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('coupons.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Button danger icon={<DeleteOutlined />} size="small">
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600, color: 'var(--text-primary)' }}>{t('coupons.title')}</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          {t('coupons.createCoupon')}
        </Button>
      </div>

      <Card style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
        <Table
          dataSource={coupons}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={editingCoupon ? t('coupons.editCoupon') : t('coupons.createCoupon')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item label={t('coupons.code')} required>
            <div style={{ display: 'flex', gap: '8px' }}>
              <Form.Item name="code" noStyle rules={[{ required: true }]}>
                <Input style={{ flex: 1 }} />
              </Form.Item>
              <Button onClick={generateCode}>{t('coupons.generate')}</Button>
            </div>
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name="discount" label={t('coupons.discountValue')} rules={[{ required: true }]}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="isPercent" label={t('coupons.discountType')} valuePropName="checked">
              <Switch checkedChildren="%" unCheckedChildren="¥" />
            </Form.Item>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name="minAmount" label={t('coupons.minOrderAmount')}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="maxDiscount" label={t('coupons.maxDiscount')}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </div>

          <Form.Item name="maxUses" label={t('coupons.maxUses')} extra="-1 for unlimited">
            <InputNumber min={-1} style={{ width: '100%' }} />
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name="startTime" label={t('coupons.startTime')} rules={[{ required: true }]}>
              <DatePicker style={{ width: '100%' }} showTime />
            </Form.Item>
            <Form.Item name="endTime" label={t('coupons.endTime')} rules={[{ required: true }]}>
              <DatePicker style={{ width: '100%' }} showTime />
            </Form.Item>
          </div>

          <Form.Item name="enabled" label={t('coupons.enabled')} valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingCoupon ? t('common.update') : t('common.create')}
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
