import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Table, Button, Modal, Form, Input, InputNumber, Space, message, Popconfirm, Card, Tag, Typography } from 'antd'
import { PlusOutlined, DeleteOutlined, CopyOutlined, KeyOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

const { Text } = Typography

interface APIToken {
  id: number
  createdAt: string
  userId: number
  name: string
  token: string
  permissions: string
  lastUsedAt: string | null
  expiresAt: string | null
  enabled: boolean
}

export default function APITokens() {
  const { t } = useTranslation()
  const [tokens, setTokens] = useState<APIToken[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [newToken, setNewToken] = useState<string | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchTokens()
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

  const fetchTokens = async () => {
    try {
      const response = await fetch('/api/api-tokens', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const result = await response.json()
      setTokens(result.data || [])
    } catch (error) {
      message.error(t('apiTokens.fetchError'))
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    form.resetFields()
    form.setFieldsValue({
      permissions: 'read'
    })
    setNewToken(null)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/api-tokens/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      if (response.ok) {
        message.success(t('apiTokens.deleteSuccess'))
        fetchTokens()
      } else {
        message.error(t('apiTokens.deleteError'))
      }
    } catch (error) {
      message.error(t('apiTokens.deleteError'))
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      const body: any = {
        name: values.name,
        permissions: values.permissions
      }
      if (values.expiresIn) {
        body.expiresIn = values.expiresIn
      }

      const response = await fetch('/api/api-tokens', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify(body)
      })

      if (response.ok) {
        const data = await response.json()
        setNewToken(data.token)
        message.success(t('apiTokens.createSuccess'))
        fetchTokens()
        // Auto-close modal after 5 seconds to allow user to copy the token
        setTimeout(() => {
          setModalVisible(false)
          setNewToken(null)
        }, 5000)
      } else {
        const data = await response.json()
        message.error(data.error || t('apiTokens.createError'))
      }
    } catch (error) {
      message.error(t('apiTokens.createError'))
    }
  }

  const copyToken = (token: string) => {
    navigator.clipboard.writeText(token)
    message.success(t('apiTokens.copied'))
  }

  const isExpired = (expiresAt: string | null) => {
    if (!expiresAt) return false
    return dayjs(expiresAt).isBefore(dayjs())
  }

  const columns = [
    {
      title: t('apiTokens.name'),
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <KeyOutlined />
          {name}
        </Space>
      )
    },
    {
      title: t('apiTokens.token'),
      dataIndex: 'token',
      key: 'token',
      render: (token: string) => (
        <Text copyable={{ text: token }}>
          {token.substring(0, 12)}...
        </Text>
      )
    },
    {
      title: t('apiTokens.permissions'),
      dataIndex: 'permissions',
      key: 'permissions',
      render: (perms: string) => (
        <Tag color="blue">{perms}</Tag>
      )
    },
    {
      title: t('apiTokens.lastUsed'),
      dataIndex: 'lastUsedAt',
      key: 'lastUsedAt',
      render: (v: string | null) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'
    },
    {
      title: t('apiTokens.expires'),
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      render: (v: string | null) => {
        if (!v) return <Tag color="green">{t('apiTokens.neverExpires')}</Tag>
        const expired = isExpired(v)
        return (
          <Tag color={expired ? 'red' : 'blue'}>
            {expired ? t('apiTokens.expired') : dayjs(v).format('YYYY-MM-DD')}
          </Tag>
        )
      }
    },
    {
      title: t('apiTokens.status'),
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: APIToken) => (
        <Tag color={enabled && !isExpired(record.expiresAt) ? 'green' : 'red'}>
          {enabled && !isExpired(record.expiresAt) ? t('common.active') : t('common.inactive')}
        </Tag>
      )
    },
    {
      title: t('apiTokens.createdAt'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm')
    },
    {
      title: t('apiTokens.actions'),
      key: 'actions',
      render: (_: any, record: APIToken) => (
        <Popconfirm
          title={t('apiTokens.deleteConfirm')}
          onConfirm={() => handleDelete(record.id)}
          okText={t('common.confirm')}
          cancelText={t('common.cancel')}
        >
          <Button danger icon={<DeleteOutlined />} size="small">
            {t('common.delete')}
          </Button>
        </Popconfirm>
      )
    }
  ]

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">{t('apiTokens.title')}</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          {t('apiTokens.createToken')}
        </Button>
      </div>

      <Card className="bg-gray-800 border-gray-700">
        <Table
          dataSource={tokens}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={newToken ? t('apiTokens.tokenCreated') : t('apiTokens.createToken')}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false)
          setNewToken(null)
        }}
        footer={null}
        width={500}
      >
        {newToken ? (
          <div className="space-y-4">
            <div className="bg-yellow-900/30 border border-yellow-600 rounded p-4">
              <p className="text-yellow-400 font-bold mb-2">{t('apiTokens.saveWarning')}</p>
              <p className="text-gray-300 text-sm">{t('apiTokens.saveWarningDesc')}</p>
            </div>
            <div className="bg-gray-700 rounded p-3 flex items-center justify-between">
              <code className="text-sm break-all">{newToken}</code>
              <Button 
                icon={<CopyOutlined />} 
                onClick={() => copyToken(newToken)}
                className="ml-2"
              >
                {t('common.copy')}
              </Button>
            </div>
            <Button block onClick={() => {
              setModalVisible(false)
              setNewToken(null)
            }}>
              {t('common.close')}
            </Button>
          </div>
        ) : (
          <Form form={form} onFinish={handleSubmit} layout="vertical">
            <Form.Item name="name" label={t('apiTokens.name')} rules={[{ required: true }]}>
              <Input placeholder={t('apiTokens.namePlaceholder')} />
            </Form.Item>
            <Form.Item name="permissions" label={t('apiTokens.permissions')} rules={[{ required: true }]}>
              <Input placeholder="read,write,admin" />
            </Form.Item>
            <Form.Item name="expiresIn" label={t('apiTokens.expiresInHours')} extra={t('apiTokens.expiresInHint')}>
              <InputNumber min={1} className="w-full" placeholder="720" />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit">
                  {t('common.create')}
                </Button>
                <Button onClick={() => setModalVisible(false)}>
                  {t('common.cancel')}
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </Modal>
    </div>
  )
}
