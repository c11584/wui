import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { 
  Card, 
  Form, 
  Input, 
  Switch, 
  Button, 
  Space, 
  InputNumber,
  Table,
  Modal,
  Tag,
  message,
  Progress
} from 'antd'
import { SaveOutlined, ReloadOutlined, PlusOutlined, DeleteOutlined, SyncOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import api from '../api/client'

interface SystemSettings {
  registrationEnabled: boolean
  inviteOnly: boolean
  ipWhitelistEnabled: boolean
  ipWhitelist: string
  smtpHost: string
  smtpPort: number
  smtpUsername: string
  smtpPassword: string
  smtpFrom: string
  trafficAlertPercent: number
  licenseAlertDays: number
}

interface InviteCode {
  id: number
  code: string
  maxUses: number
  usedCount: number
  expiresAt: string | null
  createdAt: string
}

interface VersionInfo {
  version: string
  buildDate: string
  goVersion: string
}

interface UpdateInfo {
  hasUpdate: boolean
  currentVersion: string
  latestVersion: string
  downloadUrl: string
}

export default function SystemSettings() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [settings, setSettings] = useState<SystemSettings>({
    registrationEnabled: false,
    inviteOnly: false,
    ipWhitelistEnabled: false,
    ipWhitelist: '',
    smtpHost: '',
    smtpPort: 587,
    smtpUsername: '',
    smtpPassword: '',
    smtpFrom: '',
    trafficAlertPercent: 80,
    licenseAlertDays: 7,
  })
  const [inviteCodes, setInviteCodes] = useState<InviteCode[]>([])
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [newInvite, setNewInvite] = useState({ maxUses: 1, expiresIn: 7 })
  const [showTestEmailModal, setShowTestEmailModal] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null)
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [updateProgress, setUpdateProgress] = useState<{status: string, progress: number, message: string} | null>(null)

  useEffect(() => {
    fetchSettings()
    fetchInviteCodes()
    fetchVersion()
  }, [])

  const fetchSettings = async () => {
    setLoading(true)
    try {
      const res = await api.get('/settings')
      if (res.data.success) {
        setSettings(res.data.data)
      }
    } catch (error) {
      console.error('Failed to fetch settings:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchInviteCodes = async () => {
    try {
      const res = await api.get('/invite-codes')
      if (res.data.success) {
        setInviteCodes(res.data.data || [])
      }
    } catch (error) {
      console.error('Failed to fetch invite codes:', error)
    }
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      await api.put('/settings', settings)
      message.success(t('settings.settingsSaved'))
    } catch (error: any) {
      message.error(error.response?.data?.error || t('settings.saveFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleCreateInvite = async () => {
    try {
      const expiresAt = new Date()
      expiresAt.setDate(expiresAt.getDate() + newInvite.expiresIn)
      
      await api.post('/invite-codes', {
        maxUses: newInvite.maxUses,
        expiresAt: expiresAt.toISOString()
      })
      message.success(t('settings.inviteCreated'))
      setShowInviteModal(false)
      fetchInviteCodes()
    } catch (error: any) {
      message.error(error.response?.data?.error || t('settings.inviteCreateFailed'))
    }
  }

  const handleDeleteInvite = async (id: number) => {
    try {
      await api.delete(`/invite-codes/${id}`)
      message.success(t('settings.inviteDeleted'))
      fetchInviteCodes()
    } catch (error: any) {
      message.error(t('settings.inviteDeleteFailed'))
    }
  }

  const handleSendTestEmail = async () => {
    if (!testEmail) {
      message.error(t('settings.enterEmail'))
      return
    }
    try {
      await api.post('/test-email', { email: testEmail })
      message.success(t('settings.testEmailSent'))
      setShowTestEmailModal(false)
      setTestEmail('')
    } catch (error: any) {
      message.error(error.response?.data?.error || t('settings.testEmailFailed'))
    }
  }

  const handleBackup = async () => {
    try {
      const res = await api.get('/backup', { responseType: 'blob' })
      const url = window.URL.createObjectURL(new Blob([res.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `wui-backup-${Date.now()}.json`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      message.success(t('settings.backupDownloaded'))
    } catch (error) {
      message.error(t('settings.backupFailed'))
    }
  }

  const fetchVersion = async () => {
    try {
      const res = await api.get('/system/version')
      if (res.data.success) {
        setVersionInfo(res.data.data)
      }
    } catch (error) {
      console.error('Failed to fetch version:', error)
    }
  }

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true)
    try {
      const [versionRes, updateRes] = await Promise.all([
        api.get('/system/version'),
        api.get('/system/check-update')
      ])
      
      if (versionRes.data.success) {
        setVersionInfo(versionRes.data.data)
      }
      if (updateRes.data.success) {
        setUpdateInfo(updateRes.data)
      }
    } catch (error) {
      message.error(t('version.checkFailed'))
    } finally {
      setCheckingUpdate(false)
    }
  }

  const handleDoUpdate = async () => {
    setUpdating(true)
    setUpdateProgress({ status: 'starting', progress: 0, message: 'Starting update...' })
    
    try {
      await api.post('/system/update/do')
      pollUpdateProgress()
    } catch (error: any) {
      message.error(error.response?.data?.message || t('version.updateFailed'))
      setUpdating(false)
    }
  }

  const pollUpdateProgress = async () => {
    const poll = async () => {
      try {
        const res = await api.get('/system/update/progress')
        if (res.data.success) {
          const progress = res.data.data
          setUpdateProgress(progress)
          
          if (progress.status === 'updating') {
            setTimeout(poll, 1000)
          } else if (progress.status === 'completed') {
            message.success(t('version.updateCompleted'))
            setTimeout(() => {
              window.location.reload()
            }, 2000)
          } else if (progress.status === 'failed') {
            message.error(progress.message || t('version.updateFailed'))
            setUpdating(false)
          }
        }
      } catch (error) {
        console.error('Failed to poll progress:', error)
        setTimeout(poll, 2000)
      }
    }
    poll()
  }

  const inviteColumns = [
    { title: t('settings.code'), dataIndex: 'code', key: 'code' },
    { title: t('settings.maxUses'), dataIndex: 'maxUses', key: 'maxUses' },
    { title: t('settings.usedCount'), dataIndex: 'usedCount', key: 'usedCount' },
    { 
      title: t('common.status'), 
      key: 'status',
      render: (_: any, record: InviteCode) => {
        if (record.expiresAt && new Date(record.expiresAt) < new Date()) {
          return <Tag color="red">{t('settings.expired')}</Tag>
        }
        if (record.usedCount >= record.maxUses) {
          return <Tag color="orange">{t('settings.usedUp')}</Tag>
        }
        return <Tag color="green">{t('settings.active')}</Tag>
      }
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_: any, record: InviteCode) => (
        <Button 
          type="link" 
          danger 
          icon={<DeleteOutlined />}
          onClick={() => handleDeleteInvite(record.id)}
        >
          {t('common.delete')}
        </Button>
      )
    }
  ]

  return (
    <div style={{ padding: '24px' }}>
      <Card style={{ marginBottom: '24px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}>
        <div style={{ marginBottom: '16px' }}>
          <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 500, color: 'var(--text-primary)' }}>
            {t('settings.systemSettings')}
          </h2>
          <p style={{ margin: '8px 0 0 0', color: 'var(--text-tertiary)', fontSize: '14px' }}>
            {t('settings.subtitle')}
          </p>
        </div>

        <Form layout="vertical">
          <Card 
            title={t('settings.registration')} 
            style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
          >
            <Space style={{ width: '100%' }}>
              <Form.Item label={t('settings.enableRegistration')}>
                <Switch
                  checked={settings.registrationEnabled}
                  onChange={(checked) => setSettings({ ...settings, registrationEnabled: checked })}
                />
              </Form.Item>
              <Form.Item label={t('settings.inviteOnly')}>
                <Switch
                  checked={settings.inviteOnly}
                  onChange={(checked) => setSettings({ ...settings, inviteOnly: checked })}
                />
              </Form.Item>
            </Space>
          </Card>

          <Card 
            title={t('common.security')} 
            style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
          >
            <Space style={{ width: '100%', marginBottom: '16px' }}>
              <Form.Item label={t('settings.ipWhitelist')}>
                <Switch
                  checked={settings.ipWhitelistEnabled}
                  onChange={(checked) => setSettings({ ...settings, ipWhitelistEnabled: checked })}
                />
              </Form.Item>
            </Space>
            {settings.ipWhitelistEnabled && (
              <Form.Item label={t('settings.allowedIps')}>
                <Input.TextArea
                  rows={4}
                  value={settings.ipWhitelist}
                  onChange={(e) => setSettings({ ...settings, ipWhitelist: e.target.value })}
                  placeholder="192.168.1.0/24&#10;10.0.0.1"
                />
              </Form.Item>
            )}
          </Card>

          <Card 
            title={t('settings.emailSmtp')} 
            style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
          >
            <Space style={{ width: '100%' }}>
              <Form.Item label={t('settings.host')} style={{ flex: 1 }}>
                <Input
                  value={settings.smtpHost}
                  onChange={(e) => setSettings({ ...settings, smtpHost: e.target.value })}
                  placeholder="smtp.example.com"
                />
              </Form.Item>
              <Form.Item label={t('settings.port')} style={{ width: 100 }}>
                <InputNumber
                  value={settings.smtpPort}
                  onChange={(value) => setSettings({ ...settings, smtpPort: value || 587 })}
                />
              </Form.Item>
            </Space>
            <Space style={{ width: '100%' }}>
              <Form.Item label={t('settings.smtpUsername')} style={{ flex: 1 }}>
                <Input
                  value={settings.smtpUsername}
                  onChange={(e) => setSettings({ ...settings, smtpUsername: e.target.value })}
                />
              </Form.Item>
              <Form.Item label={t('settings.smtpPassword')} style={{ flex: 1 }}>
                <Input.Password
                  value={settings.smtpPassword}
                  onChange={(e) => setSettings({ ...settings, smtpPassword: e.target.value })}
                />
              </Form.Item>
            </Space>
            <Form.Item label={t('settings.fromAddress')}>
              <Input
                value={settings.smtpFrom}
                onChange={(e) => setSettings({ ...settings, smtpFrom: e.target.value })}
                placeholder="noreply@example.com"
              />
            </Form.Item>
            <Button onClick={() => setShowTestEmailModal(true)}>{t('settings.sendTestEmail')}</Button>
          </Card>

          <Card 
            title={t('settings.alerts')} 
            style={{ marginBottom: '16px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
          >
            <Space>
              <Form.Item label={t('settings.trafficAlert')}>
                <InputNumber
                  min={1}
                  max={100}
                  value={settings.trafficAlertPercent}
                  onChange={(value) => setSettings({ ...settings, trafficAlertPercent: value || 80 })}
                  addonAfter="%"
                />
              </Form.Item>
              <Form.Item label={t('settings.licenseAlert')}>
                <InputNumber
                  min={1}
                  max={30}
                  value={settings.licenseAlertDays}
                  onChange={(value) => setSettings({ ...settings, licenseAlertDays: value || 7 })}
                  addonAfter={t('store.days')}
                />
              </Form.Item>
            </Space>
          </Card>

          <Space style={{ marginBottom: '24px' }}>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={loading}>
              {t('settings.saveSettings')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={fetchSettings}>
              {t('settings.reset')}
            </Button>
            <Button onClick={handleBackup}>
              {t('settings.downloadBackup')}
            </Button>
          </Space>
        </Form>
      </Card>

      <Card 
        title={t('settings.inviteCodes')} 
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setShowInviteModal(true)}>
            {t('settings.createInviteCode')}
          </Button>
        }
        style={{ backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
      >
        <Table
          columns={inviteColumns}
          dataSource={inviteCodes}
          rowKey="id"
          pagination={false}
        />
      </Card>

      <Modal
        title={t('settings.createInviteCode')}
        open={showInviteModal}
        onCancel={() => setShowInviteModal(false)}
        onOk={handleCreateInvite}
      >
        <Form layout="vertical">
          <Form.Item label={t('settings.maxUses')}>
            <InputNumber
              min={1}
              value={newInvite.maxUses}
              onChange={(value) => setNewInvite({ ...newInvite, maxUses: value || 1 })}
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item label={t('settings.expiresInDays')}>
            <InputNumber
              min={1}
              value={newInvite.expiresIn}
              onChange={(value) => setNewInvite({ ...newInvite, expiresIn: value || 7 })}
              style={{ width: '100%' }}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('settings.sendTestEmail')}
        open={showTestEmailModal}
        onCancel={() => setShowTestEmailModal(false)}
        onOk={handleSendTestEmail}
      >
        <Form layout="vertical">
          <Form.Item label={t('settings.testEmailAddress')}>
            <Input
              type="email"
              value={testEmail}
              onChange={(e) => setTestEmail(e.target.value)}
              placeholder="test@example.com"
            />
          </Form.Item>
        </Form>
      </Modal>

      <Card 
        title={t('version.title') ?? 'Version'}
        extra={
          <Button 
            icon={<SyncOutlined spin={checkingUpdate} />} 
            onClick={handleCheckUpdate}
          >
            {checkingUpdate ? t('version.checking') : t('version.checkUpdate')}
          </Button>
        }
        style={{ marginTop: '24px', backgroundColor: 'var(--card-bg)', borderColor: 'var(--border-color)' }}
      >
        <div style={{ marginBottom: '16px' }}>
          <p style={{ margin: 0, fontSize: '16px', color: 'var(--text-primary)' }}>
            {t('version.currentVersion')}: <strong>{versionInfo?.version || '-'}</strong>
          </p>
          <p style={{ margin: '4px 0 0 0', fontSize: '14px', color: 'var(--text-tertiary)' }}>
            {t('version.buildDate')}: {versionInfo?.buildDate || '-'}
          </p>
        </div>
        {updateProgress && updating ? (
          <div style={{ marginBottom: '16px', padding: '16px', backgroundColor: 'var(--bg-tertiary)', borderRadius: '8px' }}>
            <p style={{ margin: 0, marginBottom: '12px', fontWeight: 600 }}>
              {updateProgress.message}
            </p>
            <Progress 
              percent={updateProgress.progress} 
              status={updateProgress.status === 'failed' ? 'exception' : updateProgress.status === 'completed' ? 'success' : 'active'}
            />
            {updateProgress.status === 'completed' && (
              <p style={{ margin: '8px 0 0 0', color: '#52c41e' }}>
                {t('version.restartHint')}
              </p>
            )}
          </div>
        ) : updateInfo && updateInfo.hasUpdate ? (
          <div style={{ marginBottom: '16px', padding: '16px', backgroundColor: 'rgba(250, 173, 0, 0.1)', borderRadius: '8px' }}>
            <p style={{ margin: 0, marginBottom: '8px', color: '#52c41e', fontWeight: 600 }}>
              <ExclamationCircleOutlined style={{ marginRight: '4px', color: '#faad0b' }} />
              {t('version.hasUpdate')}: {updateInfo.latestVersion}
            </p>
            <p style={{ margin: 0, marginBottom: '12px', color: 'var(--text-tertiary)' }}>
              {t('version.updateHint')}
            </p>
            <Space>
              <Button 
                type="primary"
                loading={updating}
                onClick={handleDoUpdate}
              >
                {t('version.updateNow')}
              </Button>
              <Button 
                href={updateInfo.downloadUrl}
                target="_blank"
              >
                {t('version.downloadManual')}
              </Button>
            </Space>
          </div>
        ) : updateInfo ? (
          <Tag color="green">{t('version.noUpdate')}</Tag>
        ) : null}
      </Card>
    </div>
  )
}
