use tauri::{
    menu::{Menu, MenuItem},
    tray::{TrayIcon, TrayIconBuilder},
    AppHandle, Emitter, Manager, Runtime,
};

pub fn setup_tray<R: Runtime>(
    app: &AppHandle<R>,
) -> Result<TrayIcon<R>, Box<dyn std::error::Error>> {
    let show_item = MenuItem::with_id(app, "show", "显示主窗口", true, None::<&str>)?;
    let hide_item = MenuItem::with_id(app, "hide", "隐藏主窗口", true, None::<&str>)?;
    let sep1 = MenuItem::with_id(app, "sep1", "-", true, None::<&str>)?;
    let connect_item = MenuItem::with_id(app, "connect", "连接", true, None::<&str>)?;
    let disconnect_item = MenuItem::with_id(app, "disconnect", "断开连接", true, None::<&str>)?;
    let sep2 = MenuItem::with_id(app, "sep2", "-", true, None::<&str>)?;
    let mode_global = MenuItem::with_id(app, "mode_global", "全局模式", true, None::<&str>)?;
    let mode_rule = MenuItem::with_id(app, "mode_rule", "规则模式", true, None::<&str>)?;
    let mode_direct = MenuItem::with_id(app, "mode_direct", "直连模式", true, None::<&str>)?;
    let mode_tun = MenuItem::with_id(app, "mode_tun", "TUN 模式", true, None::<&str>)?;
    let sep3 = MenuItem::with_id(app, "sep3", "-", true, None::<&str>)?;
    let quit_item = MenuItem::with_id(app, "quit", "退出", true, None::<&str>)?;

    let menu = Menu::with_items(
        app,
        &[
            &show_item,
            &hide_item,
            &sep1,
            &connect_item,
            &disconnect_item,
            &sep2,
            &mode_global,
            &mode_rule,
            &mode_direct,
            &mode_tun,
            &sep3,
            &quit_item,
        ],
    )?;

    let tray = TrayIconBuilder::new()
        .icon(app.default_window_icon().unwrap().clone())
        .menu(&menu)
        .show_menu_on_left_click(true)
        .tooltip("WUI Client")
        .on_menu_event(|app, event| match event.id.as_ref() {
            "show" => {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.show();
                    let _ = window.set_focus();
                }
            }
            "hide" => {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.hide();
                }
            }
            "connect" => {
                let _ = app.emit("tray-connect", ());
            }
            "disconnect" => {
                let _ = app.emit("tray-disconnect", ());
            }
            "mode_global" => {
                let _ = app.emit("tray-set-mode", "Global");
            }
            "mode_rule" => {
                let _ = app.emit("tray-set-mode", "Rule");
            }
            "mode_direct" => {
                let _ = app.emit("tray-set-mode", "Direct");
            }
            "mode_tun" => {
                let _ = app.emit("tray-set-mode", "Tun");
            }
            "quit" => {
                app.exit(0);
            }
            _ => {}
        })
        .build(app)?;

    Ok(tray)
}

#[allow(dead_code)]
pub fn update_tray_icon<R: Runtime>(
    tray: &TrayIcon<R>,
    _connected: bool,
) -> Result<(), Box<dyn std::error::Error>> {
    tray.set_icon(Some(
        tray.app_handle().default_window_icon().unwrap().clone(),
    ))?;
    Ok(())
}

#[allow(dead_code)]
pub fn update_tray_tooltip<R: Runtime>(
    tray: &TrayIcon<R>,
    connected: bool,
    server_name: Option<&str>,
) -> Result<(), Box<dyn std::error::Error>> {
    let tooltip = if connected {
        if let Some(name) = server_name {
            format!("WUI Client - 已连接: {}", name)
        } else {
            "WUI Client - 已连接".to_string()
        }
    } else {
        "WUI Client - 未连接".to_string()
    };
    tray.set_tooltip(Some(&tooltip))?;
    Ok(())
}
