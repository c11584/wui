import { useState, useEffect } from 'react'
import { tunnelApi } from '../api/tunnels'
import type { Tunnel, Outbound } from '../types'
import { 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Select, 
  InputNumber, 
  Switch, 
  Space, 
  Tag, 
  Popconfirm,
  DatePicker,
  Card
} from 'antd'
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  PlayCircleOutlined, 
  PauseCircleOutlined,
  ReloadOutlined,
  EyeOutlined,
  SearchOutlined
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { generateRandomPort, generateRandomName } from '../utils'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

export default function Tunnels() {
  const { t } = useTranslation()
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [loading, setLoading] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [editingTunnel, setEditingTunnel] = useState<Tunnel | null>(null)
  const [showConfig, setShowConfig] = useState<string | null>(null)
  const [form] = Form.useForm()
  const [outbounds, setOutbounds] = useState<Partial<Outbound>[]>([])
  
  const [searchName, setSearchName] = useState('')
  const [searchProtocol, setSearchProtocol] = useState('')
  const [searchEnabled, setSearchEnabled] = useState('')

  useEffect(() => {
    fetchTunnels()
  }, [])

  const fetchTunnels = async (params?: { name?: string; protocol?: string; enabled?: string }) => {
    setLoading(true)
    try {
      const data = await tunnelApi.list({
        name: params?.name ?? searchName,
        protocol: params?.protocol ?? searchProtocol,
        enabled: params?.enabled ?? searchEnabled,
      })
      setTunnels(data)
    } catch (error) {
      console.error('Failed to fetch tunnels:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    fetchTunnels({
      name: searchName,
      protocol: searchProtocol,
      enabled: searchEnabled,
    })
  }

  const handleReset = () => {
    setSearchName('')
    setSearchProtocol('')
    setSearchEnabled('')
    fetchTunnels({ name: '', protocol: '', enabled: '' })
  }

  const handleCreate = async (values: any) => {
    try {
      const submitData = {
        ...values,
        outbounds,
        expireTime: values.expireTime ? values.expireTime.toISOString() : null,
        allowDomains: linesToArray(values.allowDomains),
        allowIps: linesToArray(values.allowIps),
        denyDomains: linesToArray(values.denyDomains),
        denyIps: linesToArray(values.denyIps),
      }
      await tunnelApi.create(submitData)
      setShowForm(false)
      form.resetFields()
      setOutbounds([])
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.createFailed'),
      })
    }
  }

  const handleUpdate = async (values: any) => {
    if (!editingTunnel) return
    try {
      const submitData = {
        ...values,
        outbounds,
        expireTime: values.expireTime ? values.expireTime.toISOString() : null,
        allowDomains: linesToArray(values.allowDomains),
        allowIps: linesToArray(values.allowIps),
        denyDomains: linesToArray(values.denyDomains),
        denyIps: linesToArray(values.denyIps),
      }
      await tunnelApi.update(editingTunnel.id, submitData)
      setEditingTunnel(null)
      form.resetFields()
      setOutbounds([])
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.updateFailed'),
      })
    }
  }

  const linesToArray = (text: string): string => {
    if (!text) return ''
    const lines = text.split('\n').map(line => line.trim()).filter(line => line.length > 0)
    return lines.length > 0 ? JSON.stringify(lines) : ''
  }

  const arrayToLines = (jsonStr: string): string => {
    if (!jsonStr) return ''
    try {
      const arr = JSON.parse(jsonStr)
      return Array.isArray(arr) ? arr.join('\n') : ''
    } catch {
      return jsonStr
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await tunnelApi.delete(id)
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.deleteFailed'),
      })
    }
  }

  const handleStart = async (id: number) => {
    try {
      await tunnelApi.start(id)
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.startFailed'),
      })
    }
  }

  const handleStop = async (id: number) => {
    try {
      await tunnelApi.stop(id)
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.stopFailed'),
      })
    }
  }

  const handleRestart = async (id: number) => {
    try {
      await tunnelApi.restart(id)
      fetchTunnels()
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.restartFailed'),
      })
    }
  }

  const handleViewConfig = async (id: number) => {
    try {
      const response = await tunnelApi.getConfig(id)
      setShowConfig(response.config)
    } catch (error: any) {
      Modal.error({
        content: error.message || t('tunnel.messages.getConfigFailed'),
      })
    }
  }

  const openCreateModal = () => {
    form.setFieldsValue({
      name: generateRandomName(),
      remark: '',
      enabled: true,
      inboundProtocol: 'socks5',
      inboundPort: generateRandomPort(),
      inboundListen: '0.0.0.0',
      inboundAuth: false,
      udpEnabled: true,
      trafficLimit: 0,
      speedLimit: 0,
      trafficResetCycle: 'monthly',
      aclEnabled: false,
      aclMode: 'blacklist',
      allowDomains: '',
      allowIps: '',
      denyDomains: '',
      denyIps: '',
    })
    setOutbounds([])
    setShowForm(true)
  }

  const openEditModal = (tunnel: Tunnel) => {
    form.setFieldsValue({
      ...tunnel,
      expireTime: tunnel.expireTime ? dayjs(tunnel.expireTime) : null,
      allowDomains: arrayToLines(tunnel.allowDomains),
      allowIps: arrayToLines(tunnel.allowIps),
      denyDomains: arrayToLines(tunnel.denyDomains),
      denyIps: arrayToLines(tunnel.denyIps),
    })
    setOutbounds(tunnel.outbounds || [])
    setEditingTunnel(tunnel)
  }

  const addOutbound = () => {
    const newOutbound: Partial<Outbound> = {
      name: `${t('tunnel.outbound')} ${(outbounds.length || 0) + 1}`,
      protocol: 'vless',
      address: '',
      port: 443,
      config: JSON.stringify({ uuid: '', encryption: 'none', flow: '' }, null, 2),
      weight: 1,
      healthCheckEnabled: false,
    }
    setOutbounds([...outbounds, newOutbound])
  }

  const removeOutbound = (index: number) => {
    const newOutbounds = [...outbounds]
    newOutbounds.splice(index, 1)
    setOutbounds(newOutbounds)
  }

  const columns: ColumnsType<Tunnel> = [
    {
      title: t('tunnel.tunnelName'),
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text: string, record: Tunnel) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          {record.remark && (
            <div style={{ fontSize: '13px', color: 'var(--text-tertiary)', marginTop: '4px' }}>
              {record.remark}
            </div>
          )}
        </div>
      ),
    },
    {
      title: t('tunnel.protocol'),
      dataIndex: 'inboundProtocol',
      key: 'inboundProtocol',
      width: 140,
      render: (text: string, record: Tunnel) => (
        <Tag color="blue">
          {text.toUpperCase()}:{record.inboundPort}
        </Tag>
      ),
    },
    {
      title: t('tunnel.outbound'),
      key: 'outbounds',
      width: 220,
      render: (_: any, record: Tunnel) => (
        record.outbounds && record.outbounds.length > 0 ? (
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            {record.outbounds.map((out, i) => (
              <div key={out.id || i} style={{ marginBottom: '4px' }}>
                {out.protocol.toUpperCase()} → {out.address}:{out.port}
              </div>
            ))}
          </div>
        ) : (
          <span style={{ color: 'var(--text-tertiary)' }}>{t('tunnel.directConnect')}</span>
        )
      ),
    },
    {
      title: t('tunnel.udp'),
      dataIndex: 'udpEnabled',
      key: 'udpEnabled',
      width: 70,
      align: 'center',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? t('tunnel.on') : t('tunnel.off')}
        </Tag>
      ),
    },
    {
      title: t('dashboard.connections'),
      dataIndex: 'connections',
      key: 'connections',
      width: 90,
      align: 'center',
      render: (count: number) => (
        <span style={{ fontWeight: 500 }}>
          {count}
        </span>
      ),
    },
    {
      title: t('tunnel.status.active'),
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      align: 'center',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? t('tunnel.running') : t('tunnel.stopped')}
        </Tag>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 280,
      fixed: 'right',
      render: (_: any, record: Tunnel) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={record.enabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            onClick={() => record.enabled ? handleStop(record.id) : handleStart(record.id)}
          >
            {record.enabled ? t('tunnel.stopBtn') : t('tunnel.startBtn')}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => handleRestart(record.id)}
          >
            {t('tunnel.restartBtn')}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            {t('tunnel.editBtn')}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewConfig(record.id)}
          >
            {t('tunnel.configBtn')}
          </Button>
          <Popconfirm
            title={t('tunnel.confirmDelete')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('tunnel.confirmText')}
            cancelText={t('tunnel.cancelText')}
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              {t('tunnel.deleteBtn')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px', background: 'var(--bg-primary)', minHeight: '100vh' }}>
      <Card>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 500 }}>{t('tunnel.title')}</h2>
            <p style={{ margin: '8px 0 0 0', color: 'var(--text-tertiary)', fontSize: '14px' }}>{t('tunnel.subtitle')}</p>
          </div>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={openCreateModal}
          >
            {t('tunnel.createTunnel')}
          </Button>
        </div>

        <div style={{ marginBottom: '16px', padding: '16px', background: 'var(--bg-tertiary)', borderRadius: '4px' }}>
          <Space size="middle">
            <Input
              placeholder={t('tunnel.tunnelName')}
              value={searchName}
              onChange={(e) => setSearchName(e.target.value)}
              onPressEnter={handleSearch}
              style={{ width: 200 }}
            />
            <Select
              placeholder={t('tunnel.protocol')}
              value={searchProtocol}
              onChange={(value) => setSearchProtocol(value)}
              allowClear
              style={{ width: 140 }}
            >
              <Select.Option value="socks5">SOCKS5</Select.Option>
              <Select.Option value="http">HTTP</Select.Option>
            </Select>
            <Select
              placeholder={t('tunnel.status.active')}
              value={searchEnabled}
              onChange={(value) => setSearchEnabled(value)}
              allowClear
              style={{ width: 120 }}
            >
              <Select.Option value="true">{t('common.enabled')}</Select.Option>
              <Select.Option value="false">{t('common.disabled')}</Select.Option>
            </Select>
            <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
              {t('common.search')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              {t('common.reset')}
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={tunnels}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `${t('tunnel.total')} ${total} ${t('tunnel.items')}`,
            showQuickJumper: true,
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title={editingTunnel ? t('tunnel.editTunnel') : t('tunnel.createTunnel')}
        open={showForm || editingTunnel !== null}
        onCancel={() => {
          setShowForm(false)
          setEditingTunnel(null)
          form.resetFields()
          setOutbounds([])
        }}
        footer={null}
        width={800}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={editingTunnel ? handleUpdate : handleCreate}
        >
          <Form.Item name="name" label={t('tunnel.tunnelName')}>
            <Input placeholder={t('tunnel.tunnelNamePlaceholder')} />
          </Form.Item>
          <Form.Item name="remark" label={t('tunnel.remark')}>
            <Input placeholder={t('tunnel.remarkPlaceholder')} />
          </Form.Item>

          <Form.Item label={t('tunnel.inboundConfig')} required>
            <Space.Compact style={{ width: '100%' }}>
              <Form.Item name="inboundProtocol" noStyle>
                <Select style={{ width: 140 }}>
                  <Select.Option value="socks5">SOCKS5</Select.Option>
                  <Select.Option value="http">HTTP</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item name="inboundPort" noStyle>
                <InputNumber placeholder={t('tunnel.port')} style={{ width: 120 }} />
              </Form.Item>
              <Form.Item name="inboundListen" noStyle>
                <Input placeholder={t('tunnel.listenAddress')} />
              </Form.Item>
            </Space.Compact>
          </Form.Item>

          <Space style={{ marginBottom: '16px' }}>
            <Form.Item name="udpEnabled" valuePropName="checked" label={t('tunnel.udp')}>
              <Switch checkedChildren={t('tunnel.on')} unCheckedChildren={t('tunnel.off')} />
            </Form.Item>
            <Form.Item name="inboundAuth" valuePropName="checked" label={t('tunnel.authConfig')}>
              <Switch checkedChildren={t('tunnel.on')} unCheckedChildren={t('tunnel.off')} />
            </Form.Item>
          </Space>

          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.inboundAuth !== cur.inboundAuth}>
            {({ getFieldValue }) => 
              getFieldValue('inboundAuth') ? (
                <Space style={{ width: '100%', marginBottom: '16px' }}>
                  <Form.Item name="inboundUsername" label={t('tunnel.username')} style={{ width: '200px', marginBottom: 0 }}>
                    <Input placeholder="username" />
                  </Form.Item>
                  <Form.Item name="inboundPassword" label={t('tunnel.password')} style={{ width: '200px', marginBottom: 0 }}>
                    <Input placeholder="password" />
                  </Form.Item>
                </Space>
              ) : null
            }
          </Form.Item>

          <Space style={{ width: '100%', marginBottom: '16px' }}>
            <Form.Item name="trafficLimit" label={t('tunnel.trafficLimitGb')} style={{ width: '200px', marginBottom: 0 }}>
              <InputNumber min={0} placeholder={t('tunnel.trafficLimitPlaceholder')} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="trafficResetCycle" label={t('tunnel.fields.resetCycle')} style={{ width: '160px', marginBottom: 0 }}>
              <Select style={{ width: '100%' }}>
                <Select.Option value="daily">{t('tunnel.fields.daily')}</Select.Option>
                <Select.Option value="weekly">{t('tunnel.fields.weekly')}</Select.Option>
                <Select.Option value="monthly">{t('tunnel.fields.monthly')}</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="speedLimit" label={t('tunnel.speedLimitKbps')} style={{ width: '200px', marginBottom: 0 }}>
              <InputNumber min={0} placeholder={t('tunnel.trafficLimitPlaceholder')} style={{ width: '100%' }} />
            </Form.Item>
          </Space>

          <Form.Item name="expireTime" label={t('tunnel.fields.expireTime')}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label={t('tunnel.aclConfig')}>
            <Space style={{ marginBottom: '12px' }}>
              <Form.Item name="aclEnabled" valuePropName="checked" noStyle>
                <Switch checkedChildren={t('tunnel.on')} unCheckedChildren={t('tunnel.off')} />
              </Form.Item>
              <span style={{ color: 'var(--text-tertiary)' }}>{t('tunnel.aclDesc')}</span>
            </Space>
            
            <Form.Item noStyle shouldUpdate={(prev, cur) => prev.aclEnabled !== cur.aclEnabled}>
              {({ getFieldValue }) => 
                getFieldValue('aclEnabled') ? (
                  <div style={{ marginTop: '12px', padding: '16px', background: 'var(--bg-tertiary)', borderRadius: '4px' }}>
                    <Form.Item name="aclMode" label={t('tunnel.fields.resetCycle').split(' ')[0]} style={{ marginBottom: '12px' }}>
                      <Select style={{ width: 200 }}>
                        <Select.Option value="blacklist">{t('tunnel.blacklistMode')}</Select.Option>
                        <Select.Option value="whitelist">{t('tunnel.whitelistMode')}</Select.Option>
                      </Select>
                    </Form.Item>
                    
                    <Form.Item noStyle shouldUpdate={(prev, cur) => prev.aclMode !== cur.aclMode}>
                      {({ getFieldValue }) => {
                        const mode = getFieldValue('aclMode')
                        return mode === 'blacklist' ? (
                          <>
                            <Form.Item name="denyDomains" label={t('tunnel.denyDomains')} style={{ marginBottom: '12px' }}
                              extra={t('tunnel.denyDomainsHint')}>
                              <Input.TextArea rows={3} placeholder="example.com&#10;*.ads.com" />
                            </Form.Item>
                            <Form.Item name="denyIps" label={t('tunnel.denyIps')} style={{ marginBottom: 0 }}
                              extra={t('tunnel.denyIpsHint')}>
                              <Input.TextArea rows={3} placeholder="192.168.1.1&#10;10.0.0.0/8" />
                            </Form.Item>
                          </>
                        ) : (
                          <>
                            <Form.Item name="allowDomains" label={t('tunnel.allowDomains')} style={{ marginBottom: '12px' }}
                              extra={t('tunnel.allowDomainsHint')}>
                              <Input.TextArea rows={3} placeholder="google.com&#10;*.google.com" />
                            </Form.Item>
                            <Form.Item name="allowIps" label={t('tunnel.allowIps')} style={{ marginBottom: 0 }}
                              extra={t('tunnel.allowIpsHint')}>
                              <Input.TextArea rows={3} placeholder="8.8.8.8&#10;1.1.1.0/24" />
                            </Form.Item>
                          </>
                        )
                      }}
                    </Form.Item>
                  </div>
                ) : null
              }
            </Form.Item>
          </Form.Item>

          <Form.Item label={t('tunnel.outboundConfig')}>
            <div style={{ marginBottom: '12px' }}>
              <Button 
                type="dashed" 
                onClick={addOutbound} 
                icon={<PlusOutlined />}
                block
              >
                {t('tunnel.addOutbound')}
              </Button>
            </div>
            
            {outbounds.length === 0 && (
              <div style={{
                textAlign: 'center',
                padding: '24px',
                color: 'var(--text-tertiary)',
                background: 'var(--bg-tertiary)',
                borderRadius: '4px'
              }}>
                {t('tunnel.noOutboundsHint')}
              </div>
            )}
            
            {outbounds.map((outbound, index) => (
              <Card 
                key={index} 
                size="small" 
                style={{ marginBottom: '12px' }}
                extra={
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => removeOutbound(index)}
                  />
                }
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Space style={{ width: '100%' }}>
                    <Form.Item style={{ flex: 1, marginBottom: 0 }}>
                      <Input
                        placeholder={t('tunnel.tunnelName')}
                        value={outbound.name}
                        onChange={(e) => {
                          const newOutbounds = [...outbounds]
                          newOutbounds[index] = { ...newOutbounds[index], name: e.target.value }
                          setOutbounds(newOutbounds)
                        }}
                      />
                    </Form.Item>
                    <Form.Item style={{ width: '140px', marginBottom: 0 }}>
                      <Select
                        value={outbound.protocol}
                        onChange={(value) => {
                          const newOutbounds = [...outbounds]
                          newOutbounds[index] = { ...newOutbounds[index], protocol: value }
                          setOutbounds(newOutbounds)
                        }}
                        style={{ width: '100%' }}
                      >
                        <Select.Option value="socks5">SOCKS5</Select.Option>
                        <Select.Option value="http">HTTP</Select.Option>
                        <Select.Option value="vless">VLESS</Select.Option>
                        <Select.Option value="vmess">VMess</Select.Option>
                        <Select.Option value="trojan">Trojan</Select.Option>
                        <Select.Option value="shadowsocks">Shadowsocks</Select.Option>
                      </Select>
                    </Form.Item>
                  </Space>
                  <Space style={{ width: '100%' }}>
                    <Form.Item style={{ flex: 1, marginBottom: 0 }}>
                      <Input
                        placeholder={t('tunnel.serverAddress')}
                        value={outbound.address}
                        onChange={(e) => {
                          const newOutbounds = [...outbounds]
                          newOutbounds[index] = { ...newOutbounds[index], address: e.target.value }
                          setOutbounds(newOutbounds)
                        }}
                      />
                    </Form.Item>
                    <Form.Item style={{ width: '120px', marginBottom: 0 }}>
                      <InputNumber
                        placeholder={t('tunnel.port')}
                        value={outbound.port}
                        onChange={(value) => {
                          const newOutbounds = [...outbounds]
                          newOutbounds[index] = { ...newOutbounds[index], port: value || 443 }
                          setOutbounds(newOutbounds)
                        }}
                        style={{ width: '100%' }}
                      />
                    </Form.Item>
                  </Space>
                  <Form.Item style={{ marginBottom: 0 }}>
                    <Input.TextArea
                      placeholder={t('tunnel.protocolConfig')}
                      value={outbound.config}
                      onChange={(e) => {
                        const newOutbounds = [...outbounds]
                        newOutbounds[index] = { ...newOutbounds[index], config: e.target.value }
                        setOutbounds(newOutbounds)
                      }}
                      rows={3}
                      style={{ fontFamily: 'monospace' }}
                    />
                  </Form.Item>
                </Space>
              </Card>
            ))}
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: '24px' }}>
            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => {
                setShowForm(false)
                setEditingTunnel(null)
                form.resetFields()
                setOutbounds([])
              }}>
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit">
                {editingTunnel ? t('common.save') : t('common.create')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('tunnel.xrayConfig')}
        open={showConfig !== null}
        onCancel={() => setShowConfig(null)}
        footer={null}
        width={900}
      >
        <pre style={{
          background: 'var(--bg-tertiary)',
          padding: '16px',
          borderRadius: '4px',
          maxHeight: '60vh',
          overflow: 'auto',
          fontSize: '13px',
        }}>
          {showConfig}
        </pre>
      </Modal>
    </div>
  )
}
