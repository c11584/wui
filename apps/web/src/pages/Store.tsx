import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Button, Modal, Form, Input, Radio, Space, message, Tag, Spin, QRCode } from 'antd'
import { ShoppingCartOutlined, CheckOutlined, StarOutlined, TagOutlined, ReloadOutlined } from '@ant-design/icons'

interface Package {
  id: number
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

interface CouponValidation {
  valid: boolean
  discount: number
  isPercent: boolean
}

interface PaymentConfig {
  epay: { enabled: boolean; apiUrl: string; pid: string }
  alipay: { enabled: boolean; appId: string; notifyUrl: string }
  wechat: { enabled: boolean; appId: string; mchId: string; notifyUrl: string }
  usdt: { enabled: boolean; address: string; network: string; minConfirm: number }
}

interface Order {
  id: number
  orderNo: string
  amount: number
  currency: string
  status: string
  payMethod: string
}

export default function Store() {
  const { t } = useTranslation()
  const [packages, setPackages] = useState<Package[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [payModalVisible, setPayModalVisible] = useState(false)
  const [selectedPackage, setSelectedPackage] = useState<Package | null>(null)
  const [createdOrder, setCreatedOrder] = useState<Order | null>(null)
  const [couponCode, setCouponCode] = useState('')
  const [couponValidation, setCouponValidation] = useState<CouponValidation | null>(null)
  const [validatingCoupon, setValidatingCoupon] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [checkingPayment, setCheckingPayment] = useState(false)
  const [paymentConfig, setPaymentConfig] = useState<PaymentConfig | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchPackages()
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

  const fetchPackages = async () => {
    try {
      const response = await fetch('/api/packages', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setPackages(result.data || [])
    } catch (error) {
      message.error(t('store.fetchError'))
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

  const handleSelectPackage = (pkg: Package) => {
    setSelectedPackage(pkg)
    setCouponCode('')
    setCouponValidation(null)
    const defaultPayMethod = paymentConfig?.epay?.enabled ? 'epay' 
      : paymentConfig?.alipay?.enabled ? 'alipay' 
      : paymentConfig?.wechat?.enabled ? 'wechat' 
      : paymentConfig?.usdt?.enabled ? 'usdt' 
      : 'epay'
    form.setFieldsValue({
      payMethod: defaultPayMethod
    })
    setModalVisible(true)
  }

  const validateCoupon = async () => {
    if (!couponCode || !selectedPackage) return

    setValidatingCoupon(true)
    try {
      const response = await fetch(
        `/api/coupons/verify?code=${encodeURIComponent(couponCode)}&amount=${selectedPackage.price}`,
        {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        }
      )
      const data = await response.json()
      if (data.success) {
        setCouponValidation(data.data)
        message.success(t('store.couponValid'))
      } else {
        setCouponValidation(null)
        message.error(data.error || t('store.couponInvalid'))
      }
    } catch (error) {
      message.error(t('store.couponValidateError'))
    } finally {
      setValidatingCoupon(false)
    }
  }

  const handleSubmit = async (values: any) => {
    if (!selectedPackage) return

    setSubmitting(true)
    try {
      const response = await fetch('/api/orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({
          packageId: selectedPackage.id,
          couponCode: couponCode || undefined,
          payMethod: values.payMethod
        })
      })

      if (response.ok) {
        const data = await response.json()
        message.success(t('store.orderCreated'))
        setModalVisible(false)
        
        if (data.data) {
          setCreatedOrder(data.data)
          setPayModalVisible(true)
        }
      } else {
        const data = await response.json()
        message.error(data.error || t('store.orderError'))
      }
    } catch (error) {
      message.error(t('store.orderError'))
    } finally {
      setSubmitting(false)
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return 'Unlimited'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(0)) + ' ' + sizes[i]
  }

  const getFinalPrice = () => {
    if (!selectedPackage) return 0
    if (!couponValidation || !couponValidation.valid) return selectedPackage.price
    return selectedPackage.price - couponValidation.discount
  }

  const handleEpay = async () => {
    if (!createdOrder) return
    try {
      const response = await fetch(`/api/payment/epay/${createdOrder.orderNo}`, {
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
    if (!createdOrder) return
    try {
      const response = await fetch(`/api/payment/alipay/${createdOrder.orderNo}`, {
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
    if (!createdOrder) return
    setCheckingPayment(true)
    try {
      const response = await fetch(`/api/orders/${createdOrder.orderNo}`, {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.data?.status === 'paid') {
        message.success(t('orders.paymentSuccess'))
        setPayModalVisible(false)
      } else {
        message.info(t('orders.paymentPending'))
      }
    } catch (error) {
      message.error(t('orders.checkError'))
    } finally {
      setCheckingPayment(false)
    }
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '16rem' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <h1 style={{
          margin: '0 0 0.5rem 0',
          fontSize: '1.875rem',
          fontWeight: 700,
          color: 'var(--text-primary)'
        }}>
          {t('store.title')}
        </h1>
        <p style={{
          margin: 0,
          fontSize: '1rem',
          color: 'var(--text-tertiary)'
        }}>
          {t('store.subtitle')}
        </p>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
        gap: '1.5rem'
      }}>
        {packages.map(pkg => (
          <Card
            key={pkg.id}
            style={{
              backgroundColor: 'var(--card-bg)',
              borderColor: 'var(--border-color)',
              position: 'relative',
              boxShadow: pkg.isPopular ? '0 0 0 2px var(--accent)' : 'var(--shadow)'
            }}
          >
            {pkg.isPopular && (
              <div style={{
                position: 'absolute',
                top: '-12px',
                left: '50%',
                transform: 'translateX(-50%)'
              }}>
                <Tag icon={<StarOutlined />} color="gold">{t('store.popular')}</Tag>
              </div>
            )}
            
            <div style={{ textAlign: 'center', marginBottom: '1rem' }}>
              <h3 style={{
                margin: '0 0 0.25rem 0',
                fontSize: '1.25rem',
                fontWeight: 700,
                color: 'var(--text-primary)'
              }}>
                {pkg.name}
              </h3>
              <p style={{
                margin: 0,
                fontSize: '0.875rem',
                color: 'var(--text-tertiary)'
              }}>
                {pkg.description}
              </p>
            </div>

            <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
              <span style={{
                fontSize: '2.25rem',
                fontWeight: 700,
                color: 'var(--text-primary)'
              }}>
                {pkg.currency}
              </span>
              <span style={{
                fontSize: '2.25rem',
                fontWeight: 700,
                color: '#22c55e'
              }}>
                {pkg.price.toFixed(2)}
              </span>
              <span style={{
                fontSize: '0.875rem',
                color: 'var(--text-tertiary)'
              }}>
                /{pkg.duration} {t('store.days')}
              </span>
            </div>

            <ul style={{
              listStyle: 'none',
              padding: 0,
              margin: '0 0 1.5rem 0'
            }}>
              <li style={{
                display: 'flex',
                alignItems: 'center',
                marginBottom: '0.5rem',
                color: 'var(--text-secondary)'
              }}>
                <CheckOutlined style={{ color: '#22c55e', marginRight: '0.5rem' }} />
                {pkg.maxTunnels} {t('store.tunnels')}
              </li>
              <li style={{
                display: 'flex',
                alignItems: 'center',
                marginBottom: '0.5rem',
                color: 'var(--text-secondary)'
              }}>
                <CheckOutlined style={{ color: '#22c55e', marginRight: '0.5rem' }} />
                {formatBytes(pkg.maxTraffic)} {t('store.traffic')}
              </li>
              {pkg.maxSpeed > 0 && (
                <li style={{
                  display: 'flex',
                  alignItems: 'center',
                  marginBottom: '0.5rem',
                  color: 'var(--text-secondary)'
                }}>
                  <CheckOutlined style={{ color: '#22c55e', marginRight: '0.5rem' }} />
                  {pkg.maxSpeed / 1024 / 1024} Mbps {t('store.speed')}
                </li>
              )}
              {pkg.features && (() => {
                try {
                  const features = JSON.parse(pkg.features)
                  return features.map((f: string, i: number) => (
                    <li key={i} style={{
                      display: 'flex',
                      alignItems: 'center',
                      marginBottom: '0.5rem',
                      color: 'var(--text-secondary)'
                    }}>
                      <CheckOutlined style={{ color: '#22c55e', marginRight: '0.5rem' }} />
                      {f}
                    </li>
                  ))
                } catch {
                  return null
                }
              })()}
            </ul>

            <Button
              type="primary"
              block
              size="large"
              icon={<ShoppingCartOutlined />}
              onClick={() => handleSelectPackage(pkg)}
            >
              {t('store.buyNow')}
            </Button>
          </Card>
        ))}
      </div>

      <Modal
        title={t('store.confirmOrder')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={500}
      >
        {selectedPackage && (
          <Form form={form} onFinish={handleSubmit} layout="vertical">
            <div style={{
              backgroundColor: 'var(--bg-tertiary)',
              borderRadius: '0.375rem',
              padding: '1rem',
              marginBottom: '1rem'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('store.package')}</span>
                <span style={{ color: 'var(--text-primary)' }}>{selectedPackage.name}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('store.duration')}</span>
                <span style={{ color: 'var(--text-primary)' }}>{selectedPackage.duration} {t('store.days')}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-tertiary)' }}>{t('store.originalPrice')}</span>
                <span style={{ color: 'var(--text-primary)' }}>{selectedPackage.currency} {selectedPackage.price.toFixed(2)}</span>
              </div>
            </div>

            <Form.Item label={t('store.couponCode')}>
              <Space.Compact style={{ width: '100%' }}>
                <Input 
                  value={couponCode}
                  onChange={e => {
                    setCouponCode(e.target.value)
                    setCouponValidation(null)
                  }}
                  placeholder={t('store.couponPlaceholder')}
                />
                <Button 
                  onClick={validateCoupon}
                  loading={validatingCoupon}
                  icon={<TagOutlined />}
                >
                  {t('store.apply')}
                </Button>
              </Space.Compact>
            </Form.Item>

            {couponValidation && couponValidation.valid && (
              <div style={{
                backgroundColor: 'rgba(34, 197, 94, 0.1)',
                border: '1px solid #22c55e',
                borderRadius: '0.375rem',
                padding: '0.75rem',
                marginBottom: '1rem'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', color: '#22c55e' }}>
                  <span>{t('store.discount')}</span>
                  <span>-{selectedPackage.currency} {couponValidation.discount.toFixed(2)}</span>
                </div>
              </div>
            )}

            <div style={{
              backgroundColor: 'var(--card-bg)',
              borderRadius: '0.375rem',
              padding: '1rem',
              marginBottom: '1rem',
              border: '1px solid var(--border-color)'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{t('store.finalPrice')}</span>
                <span style={{ color: '#22c55e', fontSize: '1.25rem', fontWeight: 600 }}>
                  {selectedPackage.currency} {getFinalPrice().toFixed(2)}
                </span>
              </div>
            </div>

            <Form.Item 
              name="payMethod" 
              label={t('store.payMethod')} 
              rules={[{ required: true }]}
            >
              <Radio.Group style={{ width: '100%' }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  {paymentConfig?.epay?.enabled && (
                    <Radio.Button value="epay" style={{ width: '100%', textAlign: 'center' }}>
                      {t('store.epay')}
                    </Radio.Button>
                  )}
                  {paymentConfig?.alipay?.enabled && (
                    <Radio.Button value="alipay" style={{ width: '100%', textAlign: 'center' }}>
                      Alipay
                    </Radio.Button>
                  )}
                  {paymentConfig?.wechat?.enabled && (
                    <Radio.Button value="wechat" style={{ width: '100%', textAlign: 'center' }}>
                      WeChat Pay
                    </Radio.Button>
                  )}
                  {paymentConfig?.usdt?.enabled && (
                    <Radio.Button value="usdt" style={{ width: '100%', textAlign: 'center' }}>
                      USDT
                    </Radio.Button>
                  )}
                  {!paymentConfig?.epay?.enabled && !paymentConfig?.alipay?.enabled && !paymentConfig?.wechat?.enabled && !paymentConfig?.usdt?.enabled && (
                    <div style={{ color: 'var(--text-tertiary)', textAlign: 'center', padding: '16px' }}>
                      {t('store.noPaymentMethod')}
                    </div>
                  )}
                </Space>
              </Radio.Group>
            </Form.Item>

            <Form.Item>
              <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
                <Button onClick={() => setModalVisible(false)}>
                  {t('common.cancel')}
                </Button>
                <Button type="primary" htmlType="submit" loading={submitting}>
                  {t('store.placeOrder')}
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title={t('orders.payOrder')}
        open={payModalVisible}
        onCancel={() => setPayModalVisible(false)}
        footer={null}
        width={500}
      >
        {createdOrder && paymentConfig && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <p style={{ color: 'var(--text-secondary)', marginBottom: '8px' }}>{t('orders.orderNo')}: {createdOrder.orderNo}</p>
              <p style={{ fontSize: '24px', fontWeight: 600, color: '#22c55e' }}>
                {createdOrder.currency} {createdOrder.amount.toFixed(2)}
              </p>
            </div>

            {createdOrder.payMethod === 'alipay' && paymentConfig.alipay.enabled && (
              <div style={{ textAlign: 'center' }}>
                <Button type="primary" size="large" onClick={handleAlipay}>
                  {t('orders.goToAlipay')}
                </Button>
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.alipayHint')}
                </p>
              </div>
            )}

            {createdOrder.payMethod === 'epay' && paymentConfig.epay?.enabled && (
              <div style={{ textAlign: 'center' }}>
                <Button type="primary" size="large" onClick={handleEpay}>
                  {t('orders.goToEpay')}
                </Button>
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.epayHint')}
                </p>
              </div>
            )}

            {createdOrder.payMethod === 'wechat' && paymentConfig.wechat.enabled && (
              <div style={{ textAlign: 'center' }}>
                <QRCode value={`weixin://wxpay/bizpayurl?pr=${createdOrder.orderNo}`} size={200} />
                <p style={{ marginTop: '16px', color: 'var(--text-tertiary)', fontSize: '14px' }}>
                  {t('orders.wechatHint')}
                </p>
              </div>
            )}

            {createdOrder.payMethod === 'usdt' && paymentConfig.usdt.enabled && (
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
