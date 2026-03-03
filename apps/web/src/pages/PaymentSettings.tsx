import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, Form, Input, Switch, Button, message, Spin, InputNumber } from 'antd'
import { CreditCardOutlined, SaveOutlined } from '@ant-design/icons'

interface PaymentConfig {
  epay: {
    enabled: boolean
    apiUrl: string
    pid: string
    notifyUrl: string
    returnUrl: string
  }
  alipay: {
    enabled: boolean
    appId: string
    notifyUrl: string
  }
  wechat: {
    enabled: boolean
    appId: string
    mchId: string
    notifyUrl: string
  }
  usdt: {
    enabled: boolean
    address: string
    network: string
    minConfirm: number
  }
}

export default function PaymentSettings() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchConfig()
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

  const fetchConfig = async () => {
    try {
      const response = await fetch('/api/commerce/payment-config', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      if (result.success) {
        form.setFieldsValue(result.data)
      }
    } catch (error) {
      message.error(t('settings.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async (values: PaymentConfig) => {
    setSaving(true)
    try {
      const response = await fetch('/api/commerce/payment-config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify(values)
      })

      if (response.ok) {
        message.success(t('settings.settingsSaved'))
        fetchConfig()
      } else {
        const data = await response.json()
        message.error(data.error || t('settings.saveFailed'))
      }
    } catch (error) {
      message.error(t('settings.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '300px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', maxWidth: '800px' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 600, color: 'var(--text-primary)' }}>
          <CreditCardOutlined style={{ marginRight: '12px' }} />
          {t('paymentSettings.title')}
        </h1>
      </div>

      <Form form={form} onFinish={handleSave} layout="vertical">
        {/* Epay */}
        <Card style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h3 style={{ margin: 0, color: 'var(--text-primary)' }}>{t('paymentSettings.epay')}</h3>
            <Form.Item name={['epay', 'enabled']} valuePropName="checked" noStyle>
              <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
            </Form.Item>
          </div>
          <Form.Item name={['epay', 'apiUrl']} label={t('paymentSettings.apiUrl')}>
            <Input placeholder="https://pay.example.com/submit.php" />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name={['epay', 'pid']} label={t('paymentSettings.pid')}>
              <Input placeholder="1000" />
            </Form.Item>
            <Form.Item name={['epay', 'key']} label={t('paymentSettings.key')}>
              <Input.Password placeholder="your-secret-key" />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name={['epay', 'notifyUrl']} label={t('paymentSettings.notifyUrl')}>
              <Input placeholder="https://your-domain.com/api/payment/epay/notify" />
            </Form.Item>
            <Form.Item name={['epay', 'returnUrl']} label={t('paymentSettings.returnUrl')}>
              <Input placeholder="https://your-domain.com/orders" />
            </Form.Item>
          </div>
        </Card>

        {/* Alipay */}
        <Card style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h3 style={{ margin: 0, color: 'var(--text-primary)' }}>{t('paymentSettings.alipay')}</h3>
            <Form.Item name={['alipay', 'enabled']} valuePropName="checked" noStyle>
              <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
            </Form.Item>
          </div>
          <Form.Item name={['alipay', 'appId']} label={t('paymentSettings.appId')}>
            <Input placeholder="2021000000000000" />
          </Form.Item>
          <Form.Item name={['alipay', 'notifyUrl']} label={t('paymentSettings.notifyUrl')}>
            <Input placeholder="https://your-domain.com/api/payment/alipay/notify" />
          </Form.Item>
        </Card>

        {/* WeChat */}
        <Card style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h3 style={{ margin: 0, color: 'var(--text-primary)' }}>{t('paymentSettings.wechat')}</h3>
            <Form.Item name={['wechat', 'enabled']} valuePropName="checked" noStyle>
              <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name={['wechat', 'appId']} label={t('paymentSettings.appId')}>
              <Input placeholder="wx000000000000" />
            </Form.Item>
            <Form.Item name={['wechat', 'mchId']} label={t('paymentSettings.mchId')}>
              <Input placeholder="0000000000" />
            </Form.Item>
          </div>
          <Form.Item name={['wechat', 'notifyUrl']} label={t('paymentSettings.notifyUrl')}>
            <Input placeholder="https://your-domain.com/api/payment/wechat/notify" />
          </Form.Item>
        </Card>

        {/* USDT */}
        <Card style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h3 style={{ margin: 0, color: 'var(--text-primary)' }}>{t('paymentSettings.usdt')}</h3>
            <Form.Item name={['usdt', 'enabled']} valuePropName="checked" noStyle>
              <Switch checkedChildren={t('common.enabled')} unCheckedChildren={t('common.disabled')} />
            </Form.Item>
          </div>
          <Form.Item name={['usdt', 'address']} label={t('paymentSettings.usdtAddress')}>
            <Input placeholder="TJxKfQ..." />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <Form.Item name={['usdt', 'network']} label={t('paymentSettings.usdtNetwork')}>
              <Input placeholder="TRC20" />
            </Form.Item>
            <Form.Item name={['usdt', 'minConfirm']} label={t('paymentSettings.usdtMinConfirm')}>
              <InputNumber min={1} max={100} style={{ width: '100%' }} placeholder="3" />
            </Form.Item>
          </div>
        </Card>

        <Form.Item>
          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving} size="large">
            {t('settings.saveSettings')}
          </Button>
        </Form.Item>
      </Form>
    </div>
  )
}
