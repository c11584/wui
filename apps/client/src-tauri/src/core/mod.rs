use crate::models::{ProxyMode, ProxyStatus};
use std::process::{Child, Command};
use std::path::PathBuf;
use tauri::AppHandle;
use tokio::fs;

pub struct CoreManager {
    app_handle: AppHandle,
    process: Option<Child>,
    mode: ProxyMode,
    current_server: Option<String>,
    current_server_name: Option<String>,
    upload: u64,
    download: u64,
}

impl CoreManager {
    pub fn new(app_handle: AppHandle) -> Self {
        Self {
            app_handle,
            process: None,
            mode: ProxyMode::default(),
            current_server: None,
            current_server_name: None,
            upload: 0,
            download: 0,
        }
    }
    
    pub fn get_status(&self) -> ProxyStatus {
        ProxyStatus {
            connected: self.process.is_some(),
            mode: self.mode.clone(),
            current_server: self.current_server.clone(),
            current_server_name: self.current_server_name.clone(),
            latency: None,
            upload: self.upload,
            download: self.download,
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
        Ok((self.upload, self.download))
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
