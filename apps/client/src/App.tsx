import { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
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

function App() {
  const [status, setStatus] = useState<ProxyStatus | null>(null);
  const [servers, setServers] = useState<ServerInfo[]>([]);
  const [subscriptions, setSubscriptions] = useState<SubscriptionInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'dashboard' | 'servers' | 'subscriptions' | 'settings'>('dashboard');
  const [addSubModal, setAddSubModal] = useState(false);
  const [newSubUrl, setNewSubUrl] = useState('');
  const [newSubName, setNewSubName] = useState('');

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
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

  const handleSelectServer = async (serverId: string) => {
    try {
      await invoke('select_server', { serverId });
      fetchData();
    } catch (error) {
      console.error('Failed to select server:', error);
    }
  };

  const handleTestLatency = async (serverId: string) => {
    try {
      const latency = await invoke<number>('test_latency', { serverId });
      setServers(servers.map(s => 
        s.id === serverId ? { ...s, latency } : s
      ));
    } catch (error) {
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
    } catch (error) {
      console.error('Failed to add subscription:', error);
    }
  };

  const handleUpdateSubscription = async (id: string) => {
    try {
      await invoke('update_subscription', { id });
      fetchData();
    } catch (error) {
      console.error('Failed to update subscription:', error);
    }
  };

  const handleRemoveSubscription = async (id: string) => {
    try {
      await invoke('remove_subscription', { id });
      fetchData();
    } catch (error) {
      console.error('Failed to remove subscription:', error);
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
    { value: 'Rule', label: '规则', icon: '📋' },
    { value: 'Global', label: '全局', icon: '🌐' },
    { value: 'Direct', label: '直连', icon: '🔗' },
    { value: 'Tun', label: 'TUN', icon: '🚇' },
  ];

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="logo">
          <span className="logo-icon">⚡</span>
          <span className="logo-text">WUI Client</span>
        </div>
        
        <nav className="nav">
          <button 
            className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`}
            onClick={() => setActiveTab('dashboard')}
          >
            <span className="nav-icon">📊</span>
            <span>仪表板</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'servers' ? 'active' : ''}`}
            onClick={() => setActiveTab('servers')}
          >
            <span className="nav-icon">🖥️</span>
            <span>节点</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'subscriptions' ? 'active' : ''}`}
            onClick={() => setActiveTab('subscriptions')}
          >
            <span className="nav-icon">📦</span>
            <span>订阅</span>
          </button>
          <button 
            className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
            onClick={() => setActiveTab('settings')}
          >
            <span className="nav-icon">⚙️</span>
            <span>设置</span>
          </button>
        </nav>
      </aside>

      <main className="main">
        {activeTab === 'dashboard' && (
          <div className="dashboard">
            <div className="status-card">
              <div className="status-header">
                <div className={`status-indicator ${status?.connected ? 'connected' : 'disconnected'}`} />
                <h2>{status?.connected ? '已连接' : '未连接'}</h2>
              </div>
              {status?.current_server_name && (
                <p className="current-server">{status.current_server_name}</p>
              )}
              {status?.latency && (
                <p className="latency">延迟: {status.latency}ms</p>
              )}
              <button 
                className={`connect-btn ${status?.connected ? 'disconnect' : ''}`}
                onClick={handleConnect}
                disabled={loading}
              >
                {loading ? '处理中...' : status?.connected ? '断开连接' : '连接'}
              </button>
            </div>

            <div className="mode-selector">
              <h3>代理模式</h3>
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
                <span className="stat-label">上传</span>
                <span className="stat-value upload">{formatBytes(status?.upload || 0)}</span>
              </div>
              <div className="stat-card">
                <span className="stat-label">下载</span>
                <span className="stat-value download">{formatBytes(status?.download || 0)}</span>
              </div>
            </div>

            <div className="quick-servers">
              <h3>快速切换</h3>
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
            <h2>节点列表</h2>
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
                      测试延迟
                    </button>
                    <button 
                      className="select-btn"
                      onClick={() => handleSelectServer(server.id)}
                    >
                      选择
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
              <h2>订阅管理</h2>
              <button 
                className="add-btn"
                onClick={() => setAddSubModal(true)}
              >
                + 添加订阅
              </button>
            </div>
            <div className="subscription-list">
              {subscriptions.map(sub => (
                <div key={sub.id} className="subscription-card">
                  <div className="sub-info">
                    <h4>{sub.name}</h4>
                    <p>{sub.server_count} 个节点</p>
                    {sub.last_update && (
                      <p className="last-update">
                        更新于: {new Date(sub.last_update).toLocaleString()}
                      </p>
                    )}
                  </div>
                  <div className="sub-actions">
                    <button onClick={() => handleUpdateSubscription(sub.id)}>
                      更新
                    </button>
                    <button 
                      className="delete-btn"
                      onClick={() => handleRemoveSubscription(sub.id)}
                    >
                      删除
                    </button>
                  </div>
                </div>
              ))}
            </div>

            {addSubModal && (
              <div className="modal-overlay" onClick={() => setAddSubModal(false)}>
                <div className="modal" onClick={e => e.stopPropagation()}>
                  <h3>添加订阅</h3>
                  <input
                    type="text"
                    placeholder="订阅名称"
                    value={newSubName}
                    onChange={e => setNewSubName(e.target.value)}
                  />
                  <input
                    type="text"
                    placeholder="订阅链接"
                    value={newSubUrl}
                    onChange={e => setNewSubUrl(e.target.value)}
                  />
                  <div className="modal-actions">
                    <button onClick={() => setAddSubModal(false)}>取消</button>
                    <button className="primary" onClick={handleAddSubscription}>添加</button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="settings-page">
            <h2>设置</h2>
            <div className="settings-section">
              <h3>常规</h3>
              <div className="setting-item">
                <label>开机自启</label>
                <input type="checkbox" />
              </div>
              <div className="setting-item">
                <label>自动连接</label>
                <input type="checkbox" />
              </div>
              <div className="setting-item">
                <label>启动时最小化</label>
                <input type="checkbox" />
              </div>
            </div>
            <div className="settings-section">
              <h3>本地端口</h3>
              <div className="setting-item">
                <label>HTTP 代理端口</label>
                <input type="number" defaultValue={7890} />
              </div>
              <div className="setting-item">
                <label>SOCKS 代理端口</label>
                <input type="number" defaultValue={7891} />
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
