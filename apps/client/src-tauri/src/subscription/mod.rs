use crate::models::{ServerInfo, SubscriptionInfo};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::PathBuf;
use tauri::AppHandle;
use tokio::fs;
use uuid::Uuid;

pub struct SubscriptionManager {
    app_handle: AppHandle,
    subscriptions: HashMap<String, Subscription>,
    servers: HashMap<String, Server>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Subscription {
    id: String,
    name: String,
    url: String,
    servers: Vec<Server>,
    last_update: Option<String>,
    expires_at: Option<String>,
    upload: u64,
    download: u64,
    total: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Server {
    id: String,
    name: String,
    address: String,
    port: u16,
    protocol: String,
    subscription_id: String,
    config: serde_json::Value,
}

impl SubscriptionManager {
    pub fn new(app_handle: AppHandle) -> Self {
        Self {
            app_handle,
            subscriptions: HashMap::new(),
            servers: HashMap::new(),
        }
    }
    
    pub async fn list(&self) -> anyhow::Result<Vec<SubscriptionInfo>> {
        let subs: Vec<SubscriptionInfo> = self.subscriptions
            .values()
            .map(|sub| SubscriptionInfo {
                id: sub.id.clone(),
                name: sub.name.clone(),
                url: sub.url.clone(),
                server_count: sub.servers.len(),
                last_update: sub.last_update.clone(),
                expires_at: sub.expires_at.clone(),
                upload: sub.upload,
                download: sub.download,
                total: sub.total,
            })
            .collect();
        
        Ok(subs)
    }
    
    pub async fn get_servers(&self) -> anyhow::Result<Vec<ServerInfo>> {
        let servers: Vec<ServerInfo> = self.servers
            .values()
            .map(|s| ServerInfo {
                id: s.id.clone(),
                name: s.name.clone(),
                address: s.address.clone(),
                port: s.port,
                protocol: s.protocol.clone(),
                latency: None,
                subscription_id: s.subscription_id.clone(),
            })
            .collect();
        
        Ok(servers)
    }
    
    pub async fn add(&mut self, url: &str, name: &str) -> anyhow::Result<()> {
        let id = Uuid::new_v4().to_string();
        
        let content = reqwest::get(url)
            .await?
            .text()
            .await?;
        
        let decoded = base64_decode(&content);
        let servers = parse_subscription_content(&decoded, &id);
        
        let subscription = Subscription {
            id: id.clone(),
            name: name.to_string(),
            url: url.to_string(),
            servers: servers.clone(),
            last_update: Some(chrono::Utc::now().to_rfc3339()),
            expires_at: None,
            upload: 0,
            download: 0,
            total: 0,
        };
        
        for server in servers {
            self.servers.insert(server.id.clone(), server);
        }
        
        self.subscriptions.insert(id, subscription);
        
        self.save().await?;
        
        Ok(())
    }
    
    pub async fn update(&mut self, id: &str) -> anyhow::Result<()> {
        let sub = self.subscriptions.get(id)
            .ok_or_else(|| anyhow::anyhow!("Subscription not found"))?
            .clone();
        
        let content = reqwest::get(&sub.url)
            .await?
            .text()
            .await?;
        
        let decoded = base64_decode(&content);
        let new_servers = parse_subscription_content(&decoded, id);
        
        for old_server in &sub.servers {
            self.servers.remove(&old_server.id);
        }
        
        for server in &new_servers {
            self.servers.insert(server.id.clone(), server.clone());
        }
        
        if let Some(sub) = self.subscriptions.get_mut(id) {
            sub.servers = new_servers;
            sub.last_update = Some(chrono::Utc::now().to_rfc3339());
        }
        
        self.save().await?;
        
        Ok(())
    }
    
    pub async fn remove(&mut self, id: &str) -> anyhow::Result<()> {
        if let Some(sub) = self.subscriptions.remove(id) {
            for server in sub.servers {
                self.servers.remove(&server.id);
            }
        }
        
        self.save().await?;
        
        Ok(())
    }
    
    pub async fn test_latency(&self, server_id: &str) -> anyhow::Result<u64> {
        let server = self.servers.get(server_id)
            .ok_or_else(|| anyhow::anyhow!("Server not found"))?;
        
        let start = std::time::Instant::now();
        let addr = format!("{}:{}", server.address, server.port);
        
        match tokio::net::TcpStream::connect(&addr).await {
            Ok(_) => Ok(start.elapsed().as_millis() as u64),
            Err(_) => Ok(9999),
        }
    }
    
    pub fn get_server_config(&self, server_id: &str) -> Option<serde_json::Value> {
        self.servers.get(server_id).map(|s| s.config.clone())
    }
    
    async fn save(&self) -> anyhow::Result<()> {
        let data_path = self.get_data_path()?;
        
        if let Some(parent) = data_path.parent() {
            fs::create_dir_all(parent).await?;
        }
        
        let data = SubscriptionData {
            subscriptions: self.subscriptions.clone(),
            servers: self.servers.clone(),
        };
        
        let content = serde_json::to_string_pretty(&data)?;
        fs::write(&data_path, content).await?;
        
        Ok(())
    }
    
    fn get_data_path(&self) -> anyhow::Result<PathBuf> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| anyhow::anyhow!("Cannot find config directory"))?;
        Ok(config_dir.join("wui-client").join("subscriptions.json"))
    }
}

#[derive(Serialize, Deserialize)]
struct SubscriptionData {
    subscriptions: HashMap<String, Subscription>,
    servers: HashMap<String, Server>,
}

fn base64_decode(input: &str) -> String {
    use base64::{Engine as _, engine::general_purpose};
    
    let cleaned = input
        .replace('-', "+")
        .replace('_', "/")
        .trim_end_matches('=')
        .to_string();
    
    let padding = 4 - (cleaned.len() % 4);
    let padded = if padding < 4 {
        format!("{}{}", cleaned, "=".repeat(padding))
    } else {
        cleaned
    };
    
    general_purpose::STANDARD
        .decode(padded)
        .map(|bytes| String::from_utf8_lossy(&bytes).to_string())
        .unwrap_or_default()
}

fn parse_subscription_content(content: &str, subscription_id: &str) -> Vec<Server> {
    content
        .lines()
        .filter(|line| !line.is_empty())
        .enumerate()
        .filter_map(|(i, line)| parse_proxy_url(line, subscription_id, i))
        .collect()
}

fn parse_proxy_url(url: &str, subscription_id: &str, index: usize) -> Option<Server> {
    let url = url.trim();
    
    if url.starts_with("vmess://") {
        parse_vmess(url, subscription_id, index)
    } else if url.starts_with("vless://") {
        parse_vless(url, subscription_id, index)
    } else if url.starts_with("trojan://") {
        parse_trojan(url, subscription_id, index)
    } else if url.starts_with("ss://") {
        parse_shadowsocks(url, subscription_id, index)
    } else {
        None
    }
}

fn parse_vmess(url: &str, subscription_id: &str, _index: usize) -> Option<Server> {
    let encoded = url.strip_prefix("vmess://")?;
    let json = base64_decode(encoded);
    let config: serde_json::Value = serde_json::from_str(&json).ok()?;
    
    Some(Server {
        id: uuid::Uuid::new_v4().to_string(),
        name: config.get("ps")?.as_str()?.to_string(),
        address: config.get("add")?.as_str()?.to_string(),
        port: config.get("port")?.as_str()?.parse().ok()?,
        protocol: "vmess".to_string(),
        subscription_id: subscription_id.to_string(),
        config,
    })
}

fn parse_vless(url: &str, subscription_id: &str, index: usize) -> Option<Server> {
    let url_without_prefix = url.strip_prefix("vless://")?;
    let parts: Vec<&str> = url_without_prefix.split('?').collect();
    let userinfo_host: Vec<&str> = parts.get(0)?.split('@').collect();
    
    let (uuid, host_port) = (
        userinfo_host.get(0)?,
        userinfo_host.get(1)?
    );
    
    let host_port_parts: Vec<&str> = host_port.split(':').collect();
    let (host, port) = (
        host_port_parts.get(0)?.to_string(),
        host_port_parts.get(1)?.parse::<u16>().ok()?
    );
    
    let name = url.split('#').nth(1)
        .map(|s| urlencoding_decode(s))
        .unwrap_or_else(|| format!("Server {}", index + 1));
    
    Some(Server {
        id: uuid::Uuid::new_v4().to_string(),
        name,
        address: host,
        port,
        protocol: "vless".to_string(),
        subscription_id: subscription_id.to_string(),
        config: serde_json::json!({
            "uuid": uuid,
            "encryption": "none",
        }),
    })
}

fn parse_trojan(url: &str, subscription_id: &str, index: usize) -> Option<Server> {
    let url_without_prefix = url.strip_prefix("trojan://")?;
    let parts: Vec<&str> = url_without_prefix.split('?').collect();
    let userinfo_host: Vec<&str> = parts.get(0)?.split('@').collect();
    
    let (password, host_port) = (
        userinfo_host.get(0)?,
        userinfo_host.get(1)?
    );
    
    let host_port_parts: Vec<&str> = host_port.split(':').collect();
    let (host, port) = (
        host_port_parts.get(0)?.to_string(),
        host_port_parts.get(1)?.parse::<u16>().ok()?
    );
    
    let name = url.split('#').nth(1)
        .map(|s| urlencoding_decode(s))
        .unwrap_or_else(|| format!("Server {}", index + 1));
    
    Some(Server {
        id: uuid::Uuid::new_v4().to_string(),
        name,
        address: host,
        port,
        protocol: "trojan".to_string(),
        subscription_id: subscription_id.to_string(),
        config: serde_json::json!({
            "password": password,
        }),
    })
}

fn parse_shadowsocks(url: &str, subscription_id: &str, index: usize) -> Option<Server> {
    let url_without_prefix = url.strip_prefix("ss://")?;
    
    let (encoded_part, name) = if url_without_prefix.contains('#') {
        let parts: Vec<&str> = url_without_prefix.splitn(2, '#').collect();
        (*parts.get(0)?, urlencoding_decode(parts.get(1)?))
    } else {
        (url_without_prefix, format!("Server {}", index + 1))
    };
    
    let decoded = base64_decode(encoded_part);
    let parts: Vec<&str> = decoded.split('@').collect();
    
    let method_password: Vec<&str> = parts.get(0)?.split(':').collect();
    let host_port: Vec<&str> = parts.get(1)?.split(':').collect();
    
    Some(Server {
        id: uuid::Uuid::new_v4().to_string(),
        name,
        address: host_port.get(0)?.to_string(),
        port: host_port.get(1)?.parse().ok()?,
        protocol: "shadowsocks".to_string(),
        subscription_id: subscription_id.to_string(),
        config: serde_json::json!({
            "method": method_password.get(0)?,
            "password": method_password.get(1)?,
        }),
    })
}

fn urlencoding_decode(s: &str) -> String {
    urlencoding::decode(s).unwrap_or_default().to_string()
}
