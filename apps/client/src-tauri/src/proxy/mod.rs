use std::io;

pub struct SystemProxy {
    original_settings: Option<ProxySettings>,
}

#[derive(Debug, Clone)]
struct ProxySettings {
    enabled: bool,
    server: Option<String>,
    port: Option<u16>,
}

impl SystemProxy {
    pub fn new() -> Self {
        Self {
            original_settings: None,
        }
    }
    
    pub async fn enable(&mut self) -> anyhow::Result<()> {
        let settings = self.get_current_settings().await?;
        self.original_settings = Some(settings);
        
        self.set_proxy(true, "127.0.0.1", 7890).await
    }
    
    pub async fn disable(&mut self) -> anyhow::Result<()> {
        self.set_proxy(false, "", 0).await
    }
    
    pub async fn restore(&mut self) -> anyhow::Result<()> {
        if let Some(settings) = &self.original_settings {
            if settings.enabled {
                if let (Some(server), Some(port)) = (&settings.server, settings.port) {
                    self.set_proxy(true, server, port).await?;
                }
            } else {
                self.set_proxy(false, "", 0).await?;
            }
        }
        Ok(())
    }
    
    #[cfg(target_os = "macos")]
    async fn get_current_settings(&self) -> io::Result<ProxySettings> {
        use std::process::Command;
        
        let output = Command::new("networksetup")
            .args(["-getwebproxy", "Wi-Fi"])
            .output()?;
        
        let stdout = String::from_utf8_lossy(&output.stdout);
        let enabled = stdout.contains("Enabled: Yes");
        
        let server = if enabled {
            stdout.lines()
                .find(|line| line.starts_with("Server:"))
                .and_then(|line| line.split(':').nth(1))
                .map(|s| s.trim().to_string())
        } else {
            None
        };
        
        let port = if enabled {
            stdout.lines()
                .find(|line| line.starts_with("Port:"))
                .and_then(|line| line.split(':').nth(1))
                .and_then(|s| s.trim().parse().ok())
        } else {
            None
        };
        
        Ok(ProxySettings { enabled, server, port })
    }
    
    #[cfg(target_os = "macos")]
    async fn set_proxy(&self, enabled: bool, server: &str, port: u16) -> anyhow::Result<()> {
        use std::process::Command;
        
        let network_services = Self::get_network_services()?;
        
        for service in network_services {
            if enabled {
                Command::new("networksetup")
                    .args(["-setwebproxy", &service, server, &port.to_string()])
                    .status()?;
                Command::new("networksetup")
                    .args(["-setsecurewebproxy", &service, server, &port.to_string()])
                    .status()?;
                Command::new("networksetup")
                    .args(["-setwebproxystate", &service, "on"])
                    .status()?;
                Command::new("networksetup")
                    .args(["-setsecurewebproxystate", &service, "on"])
                    .status()?;
            } else {
                Command::new("networksetup")
                    .args(["-setwebproxystate", &service, "off"])
                    .status()?;
                Command::new("networksetup")
                    .args(["-setsecurewebproxystate", &service, "off"])
                    .status()?;
            }
        }
        
        Ok(())
    }
    
    #[cfg(target_os = "macos")]
    fn get_network_services() -> io::Result<Vec<String>> {
        use std::process::Command;
        
        let output = Command::new("networksetup")
            .args(["-listallnetworkservices"])
            .output()?;
        
        let stdout = String::from_utf8_lossy(&output.stdout);
        let services: Vec<String> = stdout
            .lines()
            .skip(1)
            .filter(|line| !line.starts_with('*') && !line.is_empty())
            .map(|s| s.to_string())
            .collect();
        
        Ok(services)
    }
    
    #[cfg(target_os = "windows")]
    async fn get_current_settings(&self) -> io::Result<ProxySettings> {
        use winreg::RegKey;
        use winreg::enums::*;
        
        let hkcu = RegKey::predef(HKEY_CURRENT_USER);
        let path = r"Software\Microsoft\Windows\CurrentVersion\Internet Settings";
        let key = hkcu.open_subkey(path)?;
        
        let enabled: u32 = key.get_value("ProxyEnable")?;
        let server: String = key.get_value("ProxyServer").unwrap_or_default();
        
        let (host, port) = if !server.is_empty() {
            let parts: Vec<&str> = server.split(':').collect();
            let host = parts.get(0).unwrap_or(&"").to_string();
            let port = parts.get(1).and_then(|p| p.parse().ok());
            (Some(host), port)
        } else {
            (None, None)
        };
        
        Ok(ProxySettings {
            enabled: enabled != 0,
            server: host,
            port,
        })
    }
    
    #[cfg(target_os = "windows")]
    async fn set_proxy(&self, enabled: bool, server: &str, port: u16) -> anyhow::Result<()> {
        use winreg::RegKey;
        use winreg::enums::*;
        
        let hkcu = RegKey::predef(HKEY_CURRENT_USER);
        let path = r"Software\Microsoft\Windows\CurrentVersion\Internet Settings";
        let (key, _) = hkcu.create_subkey(path)?;
        
        key.set_value("ProxyEnable", &(if enabled { 1u32 } else { 0u32 }))?;
        
        if enabled {
            let proxy_server = format!("{}:{}", server, port);
            key.set_value("ProxyServer", &proxy_server)?;
        }
        
        Self::refresh_proxy_settings();
        
        Ok(())
    }
    
    #[cfg(target_os = "windows")]
    fn refresh_proxy_settings() {
        use std::process::Command;
        
        let _ = Command::new("netsh")
            .args(["winhttp", "import", "proxy", "source=ie"])
            .status();
    }
    
    #[cfg(target_os = "linux")]
    async fn get_current_settings(&self) -> io::Result<ProxySettings> {
        Ok(ProxySettings {
            enabled: false,
            server: None,
            port: None,
        })
    }
    
    #[cfg(target_os = "linux")]
    async fn set_proxy(&self, enabled: bool, server: &str, port: u16) -> anyhow::Result<()> {
        use std::process::Command;
        
        if enabled {
            let proxy = format!("http://{}:{}", server, port);
            std::env::set_var("http_proxy", &proxy);
            std::env::set_var("https_proxy", &proxy);
            std::env::set_var("HTTP_PROXY", &proxy);
            std::env::set_var("HTTPS_PROXY", &proxy);
            
            let _ = Command::new("gsettings")
                .args(["set", "org.gnome.system.proxy", "mode", "manual"])
                .status();
            let _ = Command::new("gsettings")
                .args(["set", "org.gnome.system.proxy.http", "host", server])
                .status();
            let _ = Command::new("gsettings")
                .args(["set", "org.gnome.system.proxy.http", "port", &port.to_string()])
                .status();
        } else {
            std::env::remove_var("http_proxy");
            std::env::remove_var("https_proxy");
            std::env::remove_var("HTTP_PROXY");
            std::env::remove_var("HTTPS_PROXY");
            
            let _ = Command::new("gsettings")
                .args(["set", "org.gnome.system.proxy", "mode", "none"])
                .status();
        }
        
        Ok(())
    }
}

impl Default for SystemProxy {
    fn default() -> Self {
        Self::new()
    }
}
