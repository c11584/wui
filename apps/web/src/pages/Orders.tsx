import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Table, Button, Modal, Form, Select, message, Card, Tag, InputNumber, QRCode, Space } from 'antd'
import { EyeOutlined, CreditCardOutlined, ReloadOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

interface Order {
  id: number
  createdAt: string
  orderNo: string
  userId: number
  packageId: number
  amount: number
  currency: string
  status: string
  payMethod: string
  payTime: string | null
  tradeNo: string
  couponId: number | null
  discount: number
  expireAt: string | null
  user?: { username: string; email: string }
  package?: { name: string }
}

interface PaymentConfig {
  epay: { enabled: boolean; apiUrl: string; pid: string }
  alipay: { enabled: boolean; appId: string; notifyUrl: string }
  wechat: { enabled: boolean; appId: string; mchId: string; notifyUrl: string }
  usdt: { enabled: boolean; address: string; network: string; minConfirm: number }
}

export default function Orders() {
  const { t } = useTranslation()
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [payModalVisible, setPayModalVisible] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [statusForm] = Form.useForm()
  const [paymentConfig, setPaymentConfig] = useState<PaymentConfig | null>(null)
  const [checkingPayment, setCheckingPayment] = useState(false)

  useEffect(() => {
    fetchOrders()
    fetchPaymentConfig()
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

  const isAdmin = () => {
    const authStorage = localStorage.getItem('auth-storage')
    if (authStorage) {
      try {
        const { state } = JSON.parse(authStorage)
        return state?.user?.role === 'admin'
      } catch (e) {
        return false
      }
    }
    return false
  }

  const fetchOrders = async () => {
    try {
      const url = isAdmin() ? '/api/commerce/orders' : '/api/orders'
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setOrders(result.data || [])
    } catch (error) {
      message.error(t('orders.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const fetchPaymentConfig = async () => {
    try {
      const response = await fetch('/api/commerce/payment-config', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setPaymentConfig(result.data)
    } catch (error) {
      console.error('Failed to fetch payment config:', error)
    }
  }

  const handleViewOrder = (order: Order) => {
    setSelectedOrder(order)
    statusForm.setFieldsValue({
      status: order.status,
      paidAmount: order.amount,
      transactionId: order.tradeNo
    })
    setModalVisible(true)
  }

  const handlePayOrder = (order: Order) => {
    setSelectedOrder(order)
    setPayModalVisible(true)
  }

  const handleEpay = async () => {
    if (!selectedOrder) return
    try {
      const response = await fetch(`/api/payment/epay/${selectedOrder.orderNo}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.data?.payUrl) {
        window.open(result.data.payUrl, '_blank')
      }
    } catch (error) {
      message.error(t('orders.paymentError'))
    }
  }

  const handleAlipay = async () => {
    if (!selectedOrder) return
    try {
      const response = await fetch(`/api/payment/alipay/${selectedOrder.orderNo}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.data?.payUrl) {
        window.open(result.data.payUrl, '_blank')
      }
    } catch (error) {
      message.error(t('orders.paymentError'))
    }
  }

  const handleCheckPayment = async () => {
    if (!selectedOrder) return
    setCheckingPayment(true)
    try {
      const response = await fetch(`/api/orders/${selectedOrder.orderNo}`, {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.data?.status === 'paid') {
        message.success(t('orders.paymentSuccess'))
        setPayModalVisible(false)
        fetchOrders()
      } else {
        message.info(t('orders.paymentPending'))
      }
    } catch (error) {
      message.error(t('orders.checkError'))
    } finally {
      setCheckingPayment(false)
    }
  }

  const handleUpdateStatus = async (values: any) => {
    if (!selectedOrder) return

    try {
      const response = await fetch(`/api/commerce/orders/${selectedOrder.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify(values)
      })

      if (response.ok) {
        message.success(t('orders.updateSuccess'))
        setModalVisible(false)
        fetchOrders()
      } else {
        const data = await response.json()
        message.error(data.error || t('orders.updateError'))
      }
    } catch (error) {
      message.error(t('orders.updateError'))
    }
  }

  const getStatusTag = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'orange',
      paid: 'green',
      cancelled: 'red',
      refunded: 'purple'
    }
    const labels: Record<string, string> = {
      pending: t('orders.statusPending'),
      paid: t('orders.statusPaid'),
      cancelled: t('orders.statusCancelled'),
      refunded: t('orders.statusRefunded')
    }
    return <Tag color={colors[status] || 'default'}>{labels[status] || status.toUpperCase()}</Tag>
  }

  const getPayMethodTag = (method: string) => {
    const labels: Record<string, string> = {
      alipay: 'Alipay',
      wechat: 'WeChat',
      usdt: 'USDT'
    }
    return <Tag>{labels[method] || method}</Tag>
  }

  const columns = [
    {
      title: t('orders.orderNo'),
      dataIndex: 'orderNo',
      key: 'orderNo',
      render: (no: string) => <code style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>{no}</code>
    },
    ...(isAdmin() ? [{
      title: t('orders.user'),
      key: 'user',
      render: (_: any, record: Order) => (
        <span style={{ color: 'var(--text-primary)' }}>{record.user?.username || record.userId}</span>
      )
    }] : []),
    {
      title: t('orders.package'),
      key: 'package',
      render: (_: any, record: Order) => (
        <span style={{ color: 'var(--text-primary)' }}>{record.package?.name || `Package #${record.packageId}`}</span>
      )
    },
    {
      title: t('orders.amount'),
      key: 'amount',
      render: (_: any, record: Order) => (
        <span style={{ color: '#22c55e', fontWeight: 600 }}>
          {record.currency} {record.amount.toFixed(2)}
          {record.discount > 0 && (
            <span style={{ color: 'var(--text-tertiary)', fontSize: '12px', marginLeft: '4px' }}>(-{record.discount})</span>
          )}
        </span>
      )
    },
    {
      title: t('orders.payMethod'),
      dataIndex: 'payMethod',
      key: 'payMethod',
      render: (method: string) => getPayMethodTag(method)
    },
    {
      title: t('orders.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status)
    },
    {
      title: t('orders.createdAt'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm')
    },
    {
      title: t('orders.actions'),
      key: 'actions',
      render: (_: any, record: Order) => (
        <Space>
          <Button 
            icon={<EyeOutlined />} 
            onClick={() => handleViewOrder(record)} 
            size="small"
          >
            {t('common.view')}
          </Button>
          {record.status === 'pending' && (
            <Button 
              type="primary"
              icon={<CreditCardOutlined />} 
              onClick={() => handlePayOrder(record)} 
              size="small"
            >
              {t('orders.payNow')}
            </Button>
          )}
        </Space>
      )
    }
  ]

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600, color: 'var(--text-primary)' }}>{t('orders.title')}</h1>
      </div>

      <Card style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
        <Table
          dataSource={orders}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 20 }}
        />
      </Card>

      <Modal
        title={t('orders.orderDetail')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        {selectedOrder && (
          <div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', fontSize: '14px' }}>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.orderNo')}:</span>
                <code style={{ marginLeft: '8px', color: 'var(--text-primary)' }}>{selectedOrder.orderNo}</code>
              </div>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.status')}:</span>
                <span style={{ marginLeft: '8px' }}>{getStatusTag(selectedOrder.status)}</span>
              </div>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.amount')}:</span>
                <span style={{ marginLeft: '8px', color: '#22c55e', fontWeight: 600 }}>
                  {selectedOrder.currency} {selectedOrder.amount.toFixed(2)}
                </span>
              </div>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.discount')}:</span>
                <span style={{ marginLeft: '8px', color: 'var(--text-primary)' }}>{selectedOrder.discount || '-'}</span>
              </div>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.payMethod')}:</span>
                <span style={{ marginLeft: '8px' }}>{getPayMethodTag(selectedOrder.payMethod)}</span>
              </div>
              <div>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.createdAt')}:</span>
                <span style={{ marginLeft: '8px', color: 'var(--text-primary)' }}>{dayjs(selectedOrder.createdAt).format('YYYY-MM-DD HH:mm:ss')}</span>
              </div>
              {selectedOrder.payTime && (
                <div>
                  <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.payTime')}:</span>
                  <span style={{ marginLeft: '8px', color: 'var(--text-primary)' }}>{dayjs(selectedOrder.payTime).format('YYYY-MM-DD HH:mm:ss')}</span>
                </div>
              )}
              {selectedOrder.tradeNo && (
                <div>
                  <span style={{ color: 'var(--text-tertiary)' }}>{t('orders.tradeNo')}:</span>
                  <code style={{ marginLeft: '8px', color: 'var(--text-primary)' }}>{selectedOrder.tradeNo}</code>
                </div>
              )}
            </div>

            {isAdmin() && selectedOrder.status === 'pending' && (
              <div style={{ borderTop: '1px solid var(--border-color)', paddingTop: '16px', marginTop: '16px' }}>
                <h4 style={{ color: 'var(--text-primary)', marginBottom: '12px' }}>{t('orders.updateStatus')}</h4>
                <Form form={statusForm} onFinish={handleUpdateStatus} layout="vertical">
                  <Form.Item name="status" label={t('orders.status')} rules={[{ required: true }]}>
                    <Select>
                      <Select.Option value="pending">{t('orders.statusPending')}</Select.Option>
                      <Select.Option value="paid">{t('orders.statusPaid')}</Select.Option>
                      <Select.Option value="cancelled">{t('orders.statusCancelled')}</Select.Option>
                      <Select.Option value="refunded">{t('orders.statusRefunded')}</Select.Option>
                    </Select>
                  </Form.Item>
                  <Form.Item name="paidAmount" label={t('orders.paidAmount')}>
                    <InputNumber min={0} style={{ width: '100%' }} />
                  </Form.Item>
                  <Form.Item name="transactionId" label={t('orders.transactionId')}>
                    <input style={{ width: '100%', backgroundColor: 'var(--bg-secondary)', border: '1px solid var(--border-color)', borderRadius: '4px', padding: '8px 12px', color: 'var(--text-primary)' }} />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit">
                      {t('common.update')}
                    </Button>
                  </Form.Item>
                </Form>
              </div>
            )}
          </div>
        )}
      </Modal>

      <Modal
        title={t('orders.payOrder')}
        open={payModalVisible}
        onCancel={() => setPayModalVisible(false)}
        footer={null}
        width={500}
      >
        {selectedOrder && paymentConfig && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <p style={{ color: 'var(--text-secondary)', marginBottom: '8px' }}>{t('orders.orderNo')}: {selectedOrder.orderNo}</p>
              <p style={{ fontSize: '24px', fontWeight: 600, color: '#22c55e' }}>
                {selectedOrder.currency} {selectedOrder.amount.toFixed(2)}
              </p>
            </div>

            {selectedOrder.payMethod === 'alipay' && paymentConfig.alipay.enabled && (
              <div style={{ textAlign: 'center' }}>
                <Button type="primary" size="large" onClick={handleAlipay}>
                  {t('orders.goToAlipay')}
                </Button>
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.alipayHint')}
                </p>
              </div>
            )}

            {selectedOrder.payMethod === 'epay' && paymentConfig.epay?.enabled && (
              <div style={{ textAlign: 'center' }}>
                <Button type="primary" size="large" onClick={handleEpay}>
                  {t('orders.goToEpay')}
                </Button>
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.epayHint')}
                </p>
              </div>
            )}

            {selectedOrder.payMethod === 'wechat' && paymentConfig.wechat.enabled && (
              <div style={{ textAlign: 'center' }}>
                <QRCode value={`weixin://wxpay/bizpayurl?pr=${selectedOrder.orderNo}`} size={200} />
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.wechatHint')}
                </p>
              </div>
            )}

            {selectedOrder.payMethod === 'usdt' && paymentConfig.usdt.enabled && (
              <div style={{ textAlign: 'center' }}>
                <QRCode value={paymentConfig.usdt.address} size={200} />
                <p style={{ marginTop: '16px', color: 'var(--text-primary)', fontFamily: 'monospace', fontSize: '14px', wordBreak: 'break-all' }}>
                  {paymentConfig.usdt.address}
                </p>
                <p style={{ color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.usdtHint')} ({paymentConfig.usdt.network})
                </p>
                <p style={{ color: '#f59e0b', fontSize: '12px' }}>
                  {t('orders.usdtConfirm')}: {paymentConfig.usdt.minConfirm}
                </p>
              </div>
            )}

            <div style={{ textAlign: 'center', marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '16px' }}>
              <Button icon={<ReloadOutlined />} onClick={handleCheckPayment} loading={checkingPayment}>
                {t('orders.checkPayment')}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
