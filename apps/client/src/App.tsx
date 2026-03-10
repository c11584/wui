import { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';
import { useTranslation } from 'react-i18next';
import import toast from 'react-hot-toast';
    import './App.css';

interface ProxyStatus {
  connected: boolean;
  mode: 'Global' | 'Rule' | 'Direct' | 'Tun';
  current_server: string | null;
  current_server_name: string | null;
  latency: number | null;
  upload: number;
  download: number;
}

interface ServerInfo {
  id: string;
  name: string;
  address: string;
  port: number;
  protocol: string;
  latency: number | null;
  subscription_id: string;
}

interface SubscriptionInfo {
  id: string;
  name: string;
  url: string;
  server_count: number;
  last_update: string | null;
}

type ProxyMode = 'Global' | 'Rule' | 'Direct' | 'Tun';

interface AppSettings {
  autoStart: boolean;
  autoConnect: boolean;
  startMinimized: boolean;
  httpPort: number;
  socksPort: number;
  autoUpdateSubscriptions: boolean;
  autoUpdateInterval: number;
}

function App() {
  const { t, i18n } = useTranslation();
  const [status, setStatus] = useState<ProxyStatus | null>(null);
  const [servers, setServers] = useState<ServerInfo[]>([]);
  const [subscriptions, setSubscriptions] = useState<SubscriptionInfo[]>([]);
  const [settings, setSettings] = useState<AppSettings>({
    autoStart: false,
    autoConnect: false,
    startMinimized: false,
    httpPort: 7890,
    socksPort: 7891,
    autoUpdateSubscriptions: false,
    autoUpdateInterval: 60,
  });

  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'dashboard' | 'servers' | 'subscriptions' | 'settings'>('dashboard');
  const [addSubModal, setAddSubModal] = useState(false);
  const [newSubUrl, setNewSubUrl] = useState('');
  const [newSubName, setNewSubName] = useState('');

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const isAutostartEnabled = await invoke<boolean>('is_autostart_enabled');
      setSettings(prev => ({ ...prev, autoStart: isAutostartEnabled }));
    } catch (error) {
      console.error('Failed to fetch settings:', error);
    }
  };

  useEffect(() => {
    fetchData();
    loadSettings();
    const interval = setInterval(fetchData, 5000);

    const unlistenConnect = listen('tray-connect', () => {
      handleConnect();
    });

    const unlistenDisconnect = listen('tray-disconnect', () => {
      handleDisconnect();
    });

    const unlistenSetMode = listen<string>('tray-set-mode', (event) => {
      handleModeChange(event.payload as ProxyMode);
    });

    return () => {
      clearInterval(interval);
      unlistenConnect.then(fn => fn());
      unlistenDisconnect.then(fn => fn());
      unlistenSetMode.then(fn => fn());
    };
  }, []);

  const fetchData = async () => {
    try {
      const [proxyStatus, serverList, subList] = await Promise.all([
        invoke<ProxyStatus>('get_proxy_status'),
        invoke<ServerInfo[]>('get_servers'),
        invoke<SubscriptionInfo[]>('get_subscriptions'),
      ]);
      setStatus(proxyStatus);
      setServers(serverList);
      setSubscriptions(subList);
    } catch (error) {
      console.error('Failed to fetch data:', error);
    }
  };

  const handleModeChange = async (mode: ProxyMode) => {
    setLoading(true);
    try {
      await invoke('set_proxy_mode', { mode });
      fetchData();
    } catch (error) {
      console.error('Failed to set mode:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleConnect = async () => {
    setLoading(true);
    try {
      if (status?.connected) {
        await invoke('stop_proxy');
      } else {
        await invoke('start_proxy');
      }
      fetchData();
    } catch (error) {
      console.error('Failed to toggle connection:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDisconnect = async () => {
    setLoading(true);
    try {
      await invoke('stop_proxy');
      fetchData();
    } catch (error) {
      console.error('Failed to disconnect:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectServer = async (serverId: string) => {
    try {
      await invoke('select_server', { serverId });
      fetchData();
    } catch (error) {
      console.error('Failed to select server:', error);
    }
  };

  const handleTestLatency = async (serverId: string) {
    try {
      const latency = await invoke<number>('test_latency', { serverId });
      setServers(servers.map(s => 
        s.id === serverId ? { ...s, latency } : s
      ));
    } } catch (error) {
      console.error('Failed to test latency:', error);
    }
  };

  const handleAddSubscription = async () => {
    if (!newSubUrl || !newSubName) return;
    try {
      await invoke('add_subscription', { url: newSubUrl, name: newSubName });
      setNewSubUrl('');
      setNewSubName('');
      setAddSubModal(false);
      fetchData();
      toast.success(t('subscription.addSuccess'));
    } catch (error) {
      console.error('Failed to add subscription:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(t('subscription.addFailed', { error: errorMessage }));
    }
  };

  const handleUpdateSubscription = async (id: string) => {
    try {
      await invoke('update_subscription', { id });
      fetchData();
      toast.success(t('subscription.updateSuccess'));
    } catch (error) {
      console.error('Failed to update subscription:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(t('subscription.updateFailed', { error: errorMessage }));
    }
  };

  const handleRemoveSubscription = async (id: string) {
    try {
      await invoke('remove_subscription', { id });
      fetchData();
      toast.success(t('subscription.deleteSuccess'));
    } catch (error) {
      console.error('Failed to update subscription:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      toast.error(t('subscription.deleteFailed', { error: errorMessage }));
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const modeOptions: { value: ProxyMode; label: string; icon: string }[] = [
    { value: 'Rule', label: t('mode.rule'), icon: '📋' },
    { value: 'Global', label: t('mode.global'), icon: '🌐' },
    { value: 'Direct', label: t('mode.direct'), icon: '🔗' },
    { value: 'Tun', label: t('mode.tun'), icon: '🚇' },
  ];

  const handleLanguageChange = (lng: string) => {
    i18n.changeLanguage(lng);
  };

  const loadSettings = async () => {
    try {
      const config = await invoke<AppSettings>('get_config');
      setSettings(config);
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  };

  const toggleAutoStart = async (enabled: boolean) => {
    try {
      if (enabled) {
        await invoke('enable_autostart');
      } else {
        await invoke('disable_autostart');
      }
      setSettings(prev => ({ ...prev, autoStart: enabled }));
    } catch (error) {
      console.error('Failed to toggle auto-start:', error);
    }
  };

  const toggleAutoConnect = async (enabled: boolean) => {
    try {
      await invoke('update_config', { 
        newConfig: { ...settings, autoConnect: enabled } 
      });
      setSettings(prev => ({ ...prev, autoConnect: enabled }));
    } catch (error) {
      console.error('Failed to toggle auto-connect:', error);
    }
  };

 
  return (
    <div className="app">
      <aside className="sidebar">
        <div className="logo">
          <span className="logo-icon">⚡</span>
          <span className="logo-text">{t('app.title')}</span>
        </div>
        
        <nav className="nav">
          <button 
            className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`}
            onClick={() => setActiveTab('dashboard')}
          >
            <span className="nav-icon">📊</span>
            <span>{t('nav.dashboard')}</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'servers' ? 'active' : ''}`}
            onClick={() => setActiveTab('servers')}
          >
            <span className="nav-icon">🖥️</span>
            <span>{t('nav.servers')}</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'subscriptions' ? 'active' : ''}`}
            onClick={() => setActiveTab('subscriptions')}
          >
            <span className="nav-icon">📦</span>
            <span>{t('nav.subscriptions')}</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
            onClick={() => setActiveTab('settings')}
          >
            <span className="nav-icon">⚙️</span>
            <span>{t('nav.settings')}</span>
          </button>
        </nav>
      </aside>

      <main className="main">
        {activeTab === 'dashboard' && (
          <div className="dashboard">
            <div className="status-card">
              <div className="status-header">
                <div className={`status-indicator ${status?.connected ? 'connected' : 'disconnected'}`} />
                <h2>{status?.connected ? t('status.connected') : t('status.disconnected')}</h2>
              </div>
              {status?.current_server_name && (
                <p className="current-server">{status.current_server_name}</p>
              )}
              {status?.latency && (
                <p className="latency">{t('server.latency')}: {status.latency}ms</p>
              )}
              <button 
                className={`connect-btn ${status?.connected ? 'disconnect' : ''}`}
                onClick={handleConnect}
                disabled={loading}
              >
                {loading ? t('status.connecting') : status?.connected ? t('actions.disconnect') : t('actions.connect')}
              </button>
            </div>

            <div className="mode-selector">
              <h3>{t('mode.title')}</h3>
              <div className="mode-options">
                {modeOptions.map(option => (
                  <button
                    key={option.value}
                    className={`mode-btn ${status?.mode === option.value ? 'active' : ''}`}
                    onClick={() => handleModeChange(option.value)}
                    disabled={loading}
                  >
                    <span className="mode-icon">{option.icon}</span>
                    <span>{option.label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="traffic-stats">
              <div className="stat-card">
                <span className="stat-label">{t('traffic.upload')}</span>
                <span className="stat-value upload">{formatBytes(status?.upload || 0)}</span>
              </div>
              <div className="stat-card">
                <span className="stat-label">{t('traffic.download')}</span>
                <span className="stat-value download">{formatBytes(status?.download || 0)}</span>
              </div>
            </div>

            <div className="quick-servers">
              <h3>{t('actions.select')}</h3>
              <div className="server-list">
                {servers.slice(0, 5).map(server => (
                  <button
                    key={server.id}
                    className={`server-item ${status?.current_server === server.id ? 'active' : ''}`}
                    onClick={() => handleSelectServer(server.id)}
                  >
                    <span className="server-name">{server.name}</span>
                    {server.latency && (
                      <span className={`server-latency ${server.latency < 100 ? 'fast' : server.latency < 300 ? 'medium' : 'slow'}`}>
                        {server.latency}ms
                      </span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'servers' && (
          <div className="servers-page">
            <h2>{t('server.title')}</h2>
            <div className="server-grid">
              {servers.map(server => (
                <div 
                  key={server.id} 
                  className={`server-card ${status?.current_server === server.id ? 'active' : ''}`}
                >
                  <div className="server-info">
                    <h4>{server.name}</h4>
                    <p>{server.address}:{server.port}</p>
                    <span className="protocol-tag">{server.protocol}</span>
                  </div>
                  <div className="server-actions">
                    <button 
                      className="latency-btn"
                      onClick={() => handleTestLatency(server.id)}
                    >
                      {t('actions.testLatency')}
                    </button>
                    <button 
                      className="select-btn"
                    onClick={() => handleSelectServer(server.id)}
                    >
                      {t('actions.select')}
                    </button>
                  </div>
                  {server.latency && (
                    <div className={`latency-badge ${server.latency < 100 ? 'fast' : server.latency < 300 ? 'medium' : 'slow'}`}>
                      {server.latency}ms
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'subscriptions' && (
          <div className="subscriptions-page">
            <div className="page-header">
              <h2>{t('subscription.title')}</h2>
              <button 
                className="add-btn"
                onClick={() => setAddSubModal(true)}
              >
                + {t('actions.add')}
              </button>
            </div>
            <div className="subscription-list">
              {subscriptions.map(sub => (
                <div key={sub.id} className="subscription-card">
                  <div className="sub-info">
                    <h4>{sub.name}</h4>
                    <p>{sub.server_count} {t('subscription.serverCount')}</p>
                    {sub.last_update && (
                      <p className="last-update">
                        {t('subscription.lastUpdate')}: {new Date(sub.last_update).toLocaleString()}
                      </p>
                    )}
                  </div>
                  <div className="sub-actions">
                    <button onClick={() => handleUpdateSubscription(sub.id)}>
                      {t('actions.update')}
                    </button>
                    <button 
                      className="delete-btn"
                      onClick={() => handleRemoveSubscription(sub.id)}
                    >
                      {t('actions.delete')}
                    </button>
                  </div>
                </div>
              ))}
            </div>

            {addSubModal && (
              <div className="modal-overlay" onClick={() => setAddSubModal(false)}>
                <div className="modal" onClick={e => e.stopPropagation()}>
                  <h3>{t('subscription.addTitle')}</h3>
                  <input
                    type="text"
                    placeholder={t('subscription.namePlaceholder')}
                    value={newSubName}
                    onChange={e => setNewSubName(e.target.value)}
                  />
                  <input
                    type="text"
                    placeholder={t('subscription.urlPlaceholder')}
                    value={newSubUrl}
                    onChange={e => setNewSubUrl(e.target.value)}
                  />
                  <div className="modal-actions">
                    <button onClick={() => setAddSubModal(false)}>{t('actions.cancel')}</button>
                    <button className="primary" onClick={handleAddSubscription}>{t('actions.add')}</button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="settings-page">
            <h2>{t('settings.title')}</h2>
            <div className="settings-section">
              <h3>{t('settings.language')}</h3>
              <div className="setting-item">
                <select 
                  value={i18n.language} 
                  onChange={(e) => handleLanguageChange(e.target.value)}
                  className="language-select"
                >
                  <option value="zh">{t('language.zh')}</option>
                  <option value="en">{t('language.en')}</option>
                </select>
              </div>
            </div>
            <div className="settings-section">
              <h3>{t('settings.autoStart')}</h3>
              <div className="setting-item">
                <label>{t('settings.autoStart')}</label>
                <input 
                  type="checkbox" 
                  checked={settings.autoStart}
                  onChange={(e) => toggleAutoStart(e.target.checked)}
                />
              </div>
              <div className="setting-item">
                <label>{t('settings.autoConnect')}</label>
                <input 
                  type="checkbox" 
                  checked={settings.autoConnect}
                  onChange={(e) => toggleAutoConnect(e.target.checked)}
                />
              </div>
              <div className="setting-item">
                <label>{t('settings.minimizeOnStart')}</label>
                <input 
                  type="checkbox" 
                  checked={settings.startMinimized}
                  onChange={(e) => setSettings(prev => ({ ...prev, startMinimized: e.target.checked }))}
                />
              </div>
            </div>
            <div className="settings-section">
              <h3>{t('settings.autoUpdateSubscriptions')}</h3>
              <div className="setting-item">
                <label>{t('settings.autoUpdateSubscriptions')}</label>
                <input 
                  type="checkbox" 
                  checked={settings.autoUpdateSubscriptions}
                  onChange={(e) => setSettings(prev => ({ ...prev, autoUpdateSubscriptions: e.target.checked }))}
                />
              </div>
              <div className="setting-item">
                <label>{t('settings.autoUpdateInterval')}</label>
                <input 
                  type="number" 
                  min={5}
                  max={1440}
                  value={settings.autoUpdateInterval}
                  onChange={(e) => setSettings(prev => ({ ...prev, autoUpdateInterval: parseInt(e.target.value) || 5 }))}
                />
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
