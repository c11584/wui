use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use tokio::fs;
use crate::models::ProxyMode;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub server_url: String,
    pub proxy_mode: ProxyMode,
    pub auto_start: bool,
    pub auto_connect: bool,
    pub start_minimized: bool,
    pub selected_server: Option<String>,
    pub local_http_port: u16,
    pub local_socks_port: u16,
    pub theme: String,
    pub language: String,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            server_url: String::new(),
            proxy_mode: ProxyMode::default(),
            auto_start: false,
            auto_connect: false,
            start_minimized: false,
            selected_server: None,
            local_http_port: 7890,
            local_socks_port: 7891,
            theme: "system".to_string(),
            language: "zh".to_string(),
        }
    }
}

impl AppConfig {
    pub async fn load() -> anyhow::Result<Self> {
        let config_path = Self::get_config_path()?;
        
        if config_path.exists() {
            let content = fs::read_to_string(&config_path).await?;
            let config: AppConfig = serde_json::from_str(&content)?;
            Ok(config)
        } else {
            let config = Self::default();
            config.save().await?;
            Ok(config)
        }
    }
    
    pub async fn save(&self) -> anyhow::Result<()> {
        let config_path = Self::get_config_path()?;
        
        if let Some(parent) = config_path.parent() {
            fs::create_dir_all(parent).await?;
        }
        
        let content = serde_json::to_string_pretty(self)?;
        fs::write(&config_path, content).await?;
        
        Ok(())
    }
    
    fn get_config_path() -> anyhow::Result<PathBuf> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| anyhow::anyhow!("Cannot find config directory"))?;
        Ok(config_dir.join("wui-client").join("config.json"))
    }
}
