import { useState, from 'react'
import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom'
import { Layout, Card, Table, Button, Modal, Form, Input, Select, InputNumber, DatePicker, message, Statistic, Row, Col } from 'antd'
import { KeyOutlined, UserOutlined, DashboardOutlined, License, createFromIconfont } from '@ant-design/icons'

import 'antd/dist/antd.css'
import './App.css'

interface License {
  id: number
  key: string
  type: string
  plan: string
  status: string
  maxTunnels: number
  maxUsers: number
  expiresAt: string
  createdAt: string
}

interface Stats {
  totalLicenses: number
  activeLicenses: number
  expiredLicenses: number
}

function App() {
  const [token, setToken] = useState('')
  const [licenses, setLicenses] = useState<License[]>([])
  const [stats, setStats] = useState<Stats | null)
  const [createModal, setCreateModal] = useState(false)
  const [form] = Form.useForm()

  const api = (path: string, options: RequestInit = {}) => {
    return fetch('http://localhost:8081' + path, {
      ...options,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    }).then(r => r.json())
  }

  const login = async (password: string) => {
    const res = await api('/api/admin/login', {
      method: 'POST',
      body: JSON.stringify({ password }),
    })
    if (res.token) {
      setToken(res.token)
  }
  }

  const loadLicenses = async () => {
    const res = await api('/api/admin/licenses')
    setLicenses(res.data || [])
  }

  const loadStats = async () => {
    const res = await api('/api/admin/stats')
    setStats(res)
  }

  const createLicense = async (values: any) => {
    await api('/api/admin/licenses/create', {
      method: 'POST',
      body: JSON.stringify(values),
    })
    message.success('License created')
    setCreateModal(false)
    loadLicenses()
    loadStats()
  }

  useEffect(() => {
    if (token) {
      loadLicenses()
      loadStats()
    }
  }, [token])

  if (!token) {
    return (
      <div className="login-container">
        <Card className="login-card">
          <h2>WUI Admin</h2>
          <Form onFinish={(v) => login(v.password)}>
            <Form.Item name="password" label="Password">
              <Input.Password />
            </Form.Item>
            <Button type="primary" htmlType="submit" block>
              Login
            </Button>
          </Form>
        </Card>
      </div>
    )
  }

  return (
    <div className="admin-container">
      <Layout className="layout">
        <Layout.Sider className="sidebar">
          <div className="logo">WUI Admin</div>
          <Menu
            items={[
              { key: 'dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
              { key: 'licenses', icon: <KeyOutlined />, label: 'Licenses' },
              { key: 'customers', icon: <UserOutlined />, label: 'Customers' },
            ]}
          />
        </Layout.Sider>
        <Layout.Content className="content">
          <h1>Dashboard</h1>
          
          <Row gutter={16}>
            <Col span={8}>
              <Statistic title="Total Licenses" value={stats?.totalLicenses || 0} />
            </Col>
            <Col span={8}>
              <Statistic title="Active Licenses" value={stats?.activeLicenses || 0} />
            </Col>
          </Row>

          <Card title="Licenses" extra={
            <Button type="primary" onClick={() => setCreateModal(true)}>
              Create License
            </Button>
          }>
            <Table
              dataSource={licenses}
              columns={[
                { title: 'Key', dataIndex: 'key', key: 'key' },
                { title: 'Plan', dataIndex: 'plan', key: 'plan' },
                { title: 'Status', dataIndex: 'status', key: 'status' },
                { title: 'Max Tunnels', dataIndex: 'maxTunnels', key: 'maxTunnels' },
                { title: 'Expires', dataIndex: 'expiresAt', key: 'expiresAt' },
              ]}
            />
          </Card>

          <Modal
            title="Create License"
            open={createModal}
            onCancel={() => setCreateModal(false)}
            onOk={() => form.submit()}
          >
            <Form form={form} onFinish={createLicense} layout="vertical">
              <Form.Item name="plan" label="Plan" rules={[{ required: true }]}>
                <Select>
                  <Select.Option value="basic">Basic</Select.Option>
                  <Select.Option value="pro">Pro</Select.Option>
                  <Select.Option value="enterprise">Enterprise</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item name="maxTunnels" label="Max Tunnels" rules={[{ required: true }]}>
                <InputNumber min={1} max={999} />
              </Form.Item>
              <Form.Item name="maxUsers" label="Max Users" rules={[{ required: true }]}>
                <InputNumber min={1} max={999} />
              </Form.Item>
              <Form.Item name="days" label="Valid for (days)" rules={[{ required: true }]}>
                <InputNumber min={1} max={365} />
              </Form.Item>
            </Form>
          </Modal>
        </Layout.Content>
      </Layout>
    </div>
  )
}

export default App
