use crate::models::{ProxyMode, ProxyStatus};
use std::process::{Child, Command};
use std::path::PathBuf;
use std::sync::atomic::{AtomicU64, Ordering};
use tauri::AppHandle;
use tokio::fs;

const CLASH_API_URL: &str = "http://127.0.0.1:9090";

pub struct CoreManager {
    #[allow(dead_code)]
    app_handle: AppHandle,
    process: Option<Child>,
    mode: ProxyMode,
    current_server: Option<String>,
    current_server_name: Option<String>,
    upload: AtomicU64,
    download: AtomicU64,
}

impl CoreManager {
    pub fn new(app_handle: AppHandle) -> Self {
        Self {
            app_handle,
            process: None,
            mode: ProxyMode::default(),
            current_server: None,
            current_server_name: None,
            upload: AtomicU64::new(0),
            download: AtomicU64::new(0),
        }
    }
    
    pub fn get_status(&self) -> ProxyStatus {
        let connected = self.current_server.is_some() && self.process.is_some();
        
        ProxyStatus {
            connected,
            mode: self.mode,
            current_server: self.current_server.clone(),
            current_server_name: self.current_server_name.clone(),
            latency: None,
            upload: self.upload.load(Ordering::Relaxed),
            download: self.download.load(Ordering::Relaxed),
        }
    }
    
    pub async fn set_mode(&mut self, mode: ProxyMode) -> anyhow::Result<()> {
        self.mode = mode;
        if self.process.is_some() {
            self.restart().await?;
        }
        Ok(())
    }
    
    pub async fn select_server(&mut self, server_id: &str) -> anyhow::Result<()> {
        self.current_server = Some(server_id.to_string());
        if self.process.is_some() {
            self.restart().await?;
        }
        Ok(())
    }
    
    pub async fn start(&mut self) -> anyhow::Result<()> {
        if self.process.is_some() {
            return Ok(());
        }
        
        if self.current_server.is_none() {
            return Err(anyhow::anyhow!("No server selected. Please add a subscription and select a server first."));
        }
        
        let config = self.generate_config().await?;
        let config_path = self.get_config_path()?;
        
        if let Some(parent) = config_path.parent() {
            fs::create_dir_all(parent).await?;
        }
        
        let content = serde_json::to_string_pretty(&config)?;
        fs::write(&config_path, &content).await?;
        
        let core_path = self.get_core_path()?;
        
        if !core_path.exists() {
            self.download_core().await?;
        }
        
        let child = Command::new(&core_path)
            .arg("run")
            .arg("-c")
            .arg(&config_path)
            .spawn()?;
        
        self.process = Some(child);
        
        tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
        
        if !self.check_proxy_alive().await {
            self.process = None;
            return Err(anyhow::anyhow!("Failed to start proxy core. Please check if the server configuration is valid."));
        }
        
        Ok(())
    }
    
    pub async fn stop(&mut self) -> anyhow::Result<()> {
        if let Some(mut process) = self.process.take() {
            let _ = process.kill();
        }
        Ok(())
    }
    
    pub async fn restart(&mut self) -> anyhow::Result<()> {
        self.stop().await?;
        tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
        self.start().await
    }
    
    pub async fn get_traffic(&self) -> anyhow::Result<(u64, u64)> {
        if self.process.is_some() && self.mode != ProxyMode::Direct {
            if let Ok(response) = reqwest::Client::new()
                .get(format!("{}/traffic", CLASH_API_URL))
                .timeout(std::time::Duration::from_secs(2))
                .send()
                .await
            {
                if let Ok(text) = response.text().await {
                    if let Ok(json) = serde_json::from_str::<serde_json::Value>(&text) {
                        let up = json.get("up").and_then(|v| v.as_u64()).unwrap_or(0);
                        let down = json.get("down").and_then(|v| v.as_u64()).unwrap_or(0);
                        self.upload.store(up, Ordering::Relaxed);
                        self.download.store(down, Ordering::Relaxed);
                    }
                }
            }
        }
        Ok((self.upload.load(Ordering::Relaxed), self.download.load(Ordering::Relaxed)))
    }
    
    pub async fn test_connection_latency(&self) -> anyhow::Result<u64> {
        if self.process.is_none() || self.mode == ProxyMode::Direct {
            return Err(anyhow::anyhow!("Proxy not running or in direct mode"));
        }
        
        let start = std::time::Instant::now();
        
        let proxy_url = if self.mode == ProxyMode::Tun {
            None
        } else {
            Some(reqwest::Proxy::http("http://127.0.0.1:7890")?)
        };
        
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(5));
        
        let client = if let Some(proxy) = proxy_url {
            client.proxy(proxy).build()?
        } else {
            client.build()?
        };
        
        let test_urls = [
            "http://www.gstatic.com/generate_204",
            "http://cp.cloudflare.com",
            "http://connectivitycheck.gstatic.com/generate_204",
        ];
        
        for url in &test_urls {
            match client.get(*url).send().await {
                Ok(response) => {
                    if response.status().is_success() || response.status().as_u16() == 204 {
                        return Ok(start.elapsed().as_millis() as u64);
                    }
                }
                Err(_) => continue,
            }
        }
        
        Ok(9999)
    }
    
    async fn check_proxy_alive(&self) -> bool {
        tokio::time::sleep(tokio::time::Duration::from_millis(500)).await;
        
        if self.mode == ProxyMode::Direct {
            return true;
        }
        
        let check_port = |port: u16| async move {
            tokio::net::TcpStream::connect(format!("127.0.0.1:{}", port))
                .await
                .is_ok()
        };
        
        let http_alive = check_port(7890).await;
        let socks_alive = check_port(7891).await;
        
        http_alive || socks_alive
    }
    
    async fn generate_config(&self) -> anyhow::Result<serde_json::Value> {
        let mode_config = match self.mode {
            ProxyMode::Global => self.generate_global_config(),
            ProxyMode::Rule => self.generate_rule_config(),
            ProxyMode::Direct => self.generate_direct_config(),
            ProxyMode::Tun => self.generate_tun_config(),
        };
        
        Ok(mode_config)
    }
    
    fn generate_global_config(&self) -> serde_json::Value {
        serde_json::json!({
            "log": {
                "level": "info"
            },
            "experimental": {
                "clash_api": {
                    "external_controller": "127.0.0.1:9090",
                    "secret": ""
                }
            },
            "inbounds": [
                {
                    "type": "http",
                    "tag": "http-in",
                    "listen": "127.0.0.1",
                    "listen_port": 7890
                },
                {
                    "type": "socks",
                    "tag": "socks-in",
                    "listen": "127.0.0.1",
                    "listen_port": 7891
                }
            ],
            "outbounds": [
                {
                    "type": "selector",
                    "tag": "proxy",
                    "outbounds": ["auto"],
                    "default": "auto"
                },
                {
                    "type": "urltest",
                    "tag": "auto",
                    "outbounds": []
                },
                {
                    "type": "direct",
                    "tag": "direct"
                }
            ],
            "route": {
                "rules": [
                    {
                        "outbound": "proxy"
                    }
                ],
                "final": "proxy"
            }
        })
    }
    
    fn generate_rule_config(&self) -> serde_json::Value {
        serde_json::json!({
            "log": {
                "level": "info"
            },
            "experimental": {
                "clash_api": {
                    "external_controller": "127.0.0.1:9090",
                    "secret": ""
                }
            },
            "inbounds": [
                {
                    "type": "http",
                    "tag": "http-in",
                    "listen": "127.0.0.1",
                    "listen_port": 7890
                },
                {
                    "type": "socks",
                    "tag": "socks-in",
                    "listen": "127.0.0.1",
                    "listen_port": 7891
                }
            ],
            "outbounds": [
                {
                    "type": "selector",
                    "tag": "proxy",
                    "outbounds": ["auto", "direct"],
                    "default": "auto"
                },
                {
                    "type": "urltest",
                    "tag": "auto",
                    "outbounds": []
                },
                {
                    "type": "direct",
                    "tag": "direct"
                }
            ],
            "route": {
                "rules": [
                    {
                        "protocol": "dns",
                        "outbound": "dns-out"
                    },
                    {
                        "ip_is_private": true,
                        "outbound": "direct"
                    }
                ],
                "final": "proxy",
                "auto_detect_interface": true
            }
        })
    }
    
    fn generate_direct_config(&self) -> serde_json::Value {
        serde_json::json!({
            "log": {
                "level": "info"
            },
            "inbounds": [],
            "outbounds": [
                {
                    "type": "direct",
                    "tag": "direct"
                }
            ],
            "route": {
                "final": "direct"
            }
        })
    }
    
    fn generate_tun_config(&self) -> serde_json::Value {
        serde_json::json!({
            "log": {
                "level": "info"
            },
            "experimental": {
                "clash_api": {
                    "external_controller": "127.0.0.1:9090",
                    "secret": ""
                }
            },
            "inbounds": [
                {
                    "type": "tun",
                    "tag": "tun-in",
                    "inet4_address": "172.19.0.1/30",
                    "auto_route": true,
                    "strict_route": true,
                    "stack": "system"
                }
            ],
            "outbounds": [
                {
                    "type": "selector",
                    "tag": "proxy",
                    "outbounds": ["auto"],
                    "default": "auto"
                },
                {
                    "type": "urltest",
                    "tag": "auto",
                    "outbounds": []
                },
                {
                    "type": "direct",
                    "tag": "direct"
                }
            ],
            "route": {
                "rules": [
                    {
                        "protocol": "dns",
                        "outbound": "dns-out"
                    }
                ],
                "final": "proxy",
                "auto_detect_interface": true
            }
        })
    }
    
    fn get_config_path(&self) -> anyhow::Result<PathBuf> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| anyhow::anyhow!("Cannot find config directory"))?;
        Ok(config_dir.join("wui-client").join("sing-box-config.json"))
    }
    
    fn get_core_path(&self) -> anyhow::Result<PathBuf> {
        let data_dir = dirs::data_dir()
            .ok_or_else(|| anyhow::anyhow!("Cannot find data directory"))?;
        
        #[cfg(target_os = "macos")]
        let binary_name = "sing-box-darwin-arm64";
        
        #[cfg(target_os = "windows")]
        let binary_name = "sing-box-windows-amd64.exe";
        
        #[cfg(target_os = "linux")]
        let binary_name = "sing-box-linux-amd64";
        
        Ok(data_dir.join("wui-client").join("core").join(binary_name))
    }
    
    async fn download_core(&self) -> anyhow::Result<()> {
        let core_path = self.get_core_path()?;
        
        if let Some(parent) = core_path.parent() {
            fs::create_dir_all(parent).await?;
        }
        
        let version = "1.10.7";
        
        #[cfg(target_os = "macos")]
        let platform = "darwin-arm64";
        
        #[cfg(target_os = "windows")]
        let platform = "windows-amd64";
        
        #[cfg(target_os = "linux")]
        let platform = "linux-amd64";
        
        let url = format!(
            "https://github.com/SagerNet/sing-box/releases/download/v{}/sing-box-{}-{}.tar.gz",
            version, version, platform
        );
        
        let response = reqwest::get(&url).await?;
        let bytes = response.bytes().await?;
        
        let temp_path = core_path.with_extension("tar.gz");
        fs::write(&temp_path, &bytes).await?;
        
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            let _output = Command::new("tar")
                .args(["-xzf", temp_path.to_str().unwrap(), "-C", core_path.parent().unwrap().to_str().unwrap()])
                .output()?;
            
            let extracted_name = format!("sing-box-{}-{}", version, platform);
            let extracted_path = core_path.parent().unwrap().join(&extracted_name).join("sing-box");
            if extracted_path.exists() {
                fs::rename(&extracted_path, &core_path).await?;
                fs::remove_dir_all(core_path.parent().unwrap().join(&extracted_name)).await?;
            }
            
            fs::set_permissions(&core_path, std::fs::Permissions::from_mode(0o755)).await?;
        }
        
        #[cfg(windows)]
        {
            let _output = Command::new("tar")
                .args(["-xzf", temp_path.to_str().unwrap(), "-C", core_path.parent().unwrap().to_str().unwrap()])
                .output()?;
        }
        
        fs::remove_file(&temp_path).await?;
        
        Ok(())
    }
}

impl Drop for CoreManager {
    fn drop(&mut self) {
        if let Some(mut process) = self.process.take() {
            let _ = process.kill();
        }
    }
}
