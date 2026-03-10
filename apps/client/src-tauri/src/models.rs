use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, Default, Serialize, Deserialize, PartialEq)]
pub enum ProxyMode {
    Global,
    #[default]
    Rule,
    Direct,
    Tun,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ProxyStatus {
    pub connected: bool,
    pub mode: ProxyMode,
    pub current_server: Option<String>,
    pub current_server_name: Option<String>,
    pub latency: Option<u64>,
    pub upload: u64,
    pub download: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerInfo {
    pub id: String,
    pub name: String,
    pub address: String,
    pub port: u16,
    pub protocol: String,
    pub latency: Option<u64>,
    pub subscription_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscriptionInfo {
    pub id: String,
    pub name: String,
    pub url: String,
    pub server_count: usize,
    pub last_update: Option<String>,
    pub expires_at: Option<String>,
    pub upload: u64,
    pub download: u64,
    pub total: u64,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[allow(dead_code)]
pub struct TrafficStats {
    pub upload: u64,
    pub download: u64,
}
