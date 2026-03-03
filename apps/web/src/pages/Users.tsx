import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Table, Button, Modal, Form, Input, Select, InputNumber, Tag, Space, message, Popconfirm, Card } from 'antd'
import { UserOutlined, MailOutlined, LockOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

interface User {
  id: number
  username: string
  email: string
  role: string
  status: string
  maxTunnels: number
  maxTraffic: number
  createdAt: string
  lastLoginAt: string | null
}

export default function Users() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await fetch('/api/users', {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      const data = await response.json()
      setUsers(data)
    } catch (error) {
      message.error(t('users.fetchError'))
    } finally {
      setLoading(false)
    }
  }

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

  const handleCreate = () => {
    setEditingUser(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue({
      email: user.email,
      role: user.role,
      status: user.status,
      maxTunnels: user.maxTunnels,
      maxTraffic: user.maxTraffic / (1024 * 1024 * 1024)
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/users/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      })
      if (response.ok) {
        message.success(t('users.deleteSuccess'))
        fetchUsers()
      } else {
        message.error(t('users.deleteError'))
      }
    } catch (error) {
      message.error(t('users.deleteError'))
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      const url = editingUser ? `/api/users/${editingUser.id}` : '/api/auth/register'
      const method = editingUser ? 'PUT' : 'POST'

      const body: any = {
        email: values.email,
        maxTunnels: values.maxTunnels,
        maxTraffic: values.maxTraffic * 1024 * 1024 * 1024
      }

      if (!editingUser) {
        body.username = values.username
        body.password = values.password
      } else {
        body.status = values.status
        if (values.password) {
          body.password = values.password
        }
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
        message.success(editingUser ? t('users.updateSuccess') : t('users.createSuccess'))
        setModalVisible(false)
        fetchUsers()
      } else {
        const data = await response.json()
        message.error(data.error || t('users.saveError'))
      }
    } catch (error) {
      message.error(t('users.saveError'))
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return 'Unlimited'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const columns = [
    {
      title: t('users.username'),
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: t('users.email'),
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: t('users.role'),
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => <Tag color={role === 'admin' ? 'gold' : 'blue'}>{role}</Tag>
    },
    {
      title: t('users.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={status === 'active' ? 'green' : 'red'}>{status}</Tag>
    },
    {
      title: t('users.maxTunnels'),
      dataIndex: 'maxTunnels',
      key: 'maxTunnels',
      render: (v: number) => v || 'Unlimited'
    },
    {
      title: t('users.maxTraffic'),
      dataIndex: 'maxTraffic',
      key: 'maxTraffic',
      render: (v: number) => formatBytes(v)
    },
    {
      title: t('users.createdAt'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm')
    },
    {
      title: t('users.lastLogin'),
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
      render: (v: string | null) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'
    },
    {
      title: t('users.actions'),
      key: 'actions',
      render: (_: any, record: User) => (
        <Space>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)} size="small">
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('users.deleteConfirm')}
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
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-white">{t('users.title')}</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          {t('users.createUser')}
        </Button>
      </div>

      <Card className="bg-gray-800 border-gray-700">
        <Table
          dataSource={users}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title={editingUser ? t('users.editUser') : t('users.createUser')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          {!editingUser && (
            <Form.Item name="username" label={t('users.username')} rules={[{ required: true, min: 3 }]}>
              <Input prefix={<UserOutlined />} />
            </Form.Item>
          )}
          <Form.Item name="email" label={t('users.email')} rules={[{ required: true, type: 'email' }]}>
            <Input prefix={<MailOutlined />} />
          </Form.Item>
          <Form.Item name="password" label={t('users.password')} rules={editingUser ? [] : [{ required: true, min: 6 }]}>
            <Input.Password prefix={<LockOutlined />} placeholder={editingUser ? t('users.passwordPlaceholder') : ''} />
          </Form.Item>
          {editingUser && (
            <Form.Item name="status" label={t('users.status')} rules={[{ required: true }]}>
              <Select>
                <Select.Option value="active">{t('users.statusActive')}</Select.Option>
                <Select.Option value="suspended">{t('users.statusSuspended')}</Select.Option>
              </Select>
            </Form.Item>
          )}
          <Form.Item name="maxTunnels" label={t('users.maxTunnels')} rules={[{ required: true }]}>
            <InputNumber min={1} className="w-full" />
          </Form.Item>
          <Form.Item name="maxTraffic" label={t('users.maxTrafficGB')} rules={[{ required: true }]}>
            <InputNumber min={1} className="w-full" addonAfter="GB" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingUser ? t('common.update') : t('common.create')}
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
