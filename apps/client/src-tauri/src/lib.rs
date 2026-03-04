mod config;
mod core;
mod models;
mod proxy;
mod subscription;
mod tray;

use std::sync::Arc;
use tauri::Manager;
use tokio::sync::RwLock;
use crate::config::AppConfig;
use crate::core::CoreManager;
use crate::models::{ProxyMode, ProxyStatus, ServerInfo, SubscriptionInfo};
use crate::proxy::SystemProxy;
use crate::subscription::SubscriptionManager;
use crate::tray::setup_tray;

pub struct AppState {
    pub config: Arc<RwLock<AppConfig>>,
    pub core_manager: Arc<RwLock<CoreManager>>,
    pub subscription_manager: Arc<RwLock<SubscriptionManager>>,
    pub system_proxy: Arc<RwLock<SystemProxy>>,
}

#[tauri::command]
async fn get_proxy_status(state: tauri::State<'_, Arc<AppState>>) -> Result<ProxyStatus, String> {
    let core = state.core_manager.read().await;
    Ok(core.get_status())
}

#[tauri::command]
async fn set_proxy_mode(
    mode: ProxyMode,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut core = state.core_manager.write().await;
    core.set_mode(mode.clone()).await.map_err(|e| e.to_string())?;
    
    let mut proxy = state.system_proxy.write().await;
    if mode == ProxyMode::Direct {
        proxy.disable().await.map_err(|e| e.to_string())?;
    } else {
        proxy.enable().await.map_err(|e| e.to_string())?;
    }
    
    Ok(())
}

#[tauri::command]
async fn get_servers(state: tauri::State<'_, Arc<AppState>>) -> Result<Vec<ServerInfo>, String> {
    let sub = state.subscription_manager.read().await;
    sub.get_servers().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn select_server(
    server_id: String,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut core = state.core_manager.write().await;
    core.select_server(&server_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn test_latency(
    server_id: String,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<u64, String> {
    let sub = state.subscription_manager.read().await;
    sub.test_latency(&server_id).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_subscriptions(
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<Vec<SubscriptionInfo>, String> {
    let sub = state.subscription_manager.read().await;
    sub.list().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn add_subscription(
    url: String,
    name: String,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut sub = state.subscription_manager.write().await;
    sub.add(&url, &name).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn update_subscription(
    id: String,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut sub = state.subscription_manager.write().await;
    sub.update(&id).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn remove_subscription(
    id: String,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut sub = state.subscription_manager.write().await;
    sub.remove(&id).await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_config(state: tauri::State<'_, Arc<AppState>>) -> Result<AppConfig, String> {
    let config = state.config.read().await;
    Ok(config.clone())
}

#[tauri::command]
async fn update_config(
    new_config: AppConfig,
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(), String> {
    let mut config = state.config.write().await;
    *config = new_config;
    config.save().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn start_proxy(state: tauri::State<'_, Arc<AppState>>) -> Result<(), String> {
    let mut core = state.core_manager.write().await;
    core.start().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn stop_proxy(state: tauri::State<'_, Arc<AppState>>) -> Result<(), String> {
    let mut core = state.core_manager.write().await;
    core.stop().await.map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_traffic_stats(
    state: tauri::State<'_, Arc<AppState>>,
) -> Result<(u64, u64), String> {
    let core = state.core_manager.read().await;
    core.get_traffic().await.map_err(|e| e.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let rt = tokio::runtime::Runtime::new().expect("Failed to create tokio runtime");
    
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_autostart::init(
            tauri_plugin_autostart::MacosLauncher::LaunchAgent,
            None,
        ))
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .setup(move |app| {
            let handle = app.handle();
            
            match setup_tray(&app.handle()) {
                Ok(_tray) => {
                    #[cfg(debug_assertions)]
                    eprintln!("Tray icon setup successful");
                }
                Err(e) => {
                    eprintln!("Failed to setup tray: {}", e);
                }
            }
            
            rt.block_on(async {
                let config = AppConfig::load().await.unwrap_or_default();
                let core_manager = CoreManager::new(handle.clone());
                let subscription_manager = SubscriptionManager::new(handle.clone());
                let system_proxy = SystemProxy::new();
                
                let state = Arc::new(AppState {
                    config: Arc::new(RwLock::new(config)),
                    core_manager: Arc::new(RwLock::new(core_manager)),
                    subscription_manager: Arc::new(RwLock::new(subscription_manager)),
                    system_proxy: Arc::new(RwLock::new(system_proxy)),
                });
                
                app.manage(state);
            });
            
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            get_proxy_status,
            set_proxy_mode,
            get_servers,
            select_server,
            test_latency,
            get_subscriptions,
            add_subscription,
            update_subscription,
            remove_subscription,
            get_config,
            update_config,
            start_proxy,
            stop_proxy,
            get_traffic_stats,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
