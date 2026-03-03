# WUI - 快速开始指南

## 一键安装

### 默认安装（推荐）

```bash
curl -fsSL https://your-domain.com/install.sh | bash
```

默认配置：
- 端口：32451
- 用户名：admin
- 密码：admin

### 自定义安装

```bash
curl -fsSL https://your-domain.com/install.sh | bash -s -- \
  --port 32451 \
  --username admin \
  --password admin \
  --install-dir /opt/wui
```

### 安装参数

- `--port`: Web 面板端口（默认：32451）
- `--username`: 管理员用户名（默认：admin）
- `--password`: 管理员密码（默认：admin）
- `--install-dir`: 安装目录（默认：/opt/wui）
- `--help`: 显示帮助信息

## 支持的操作系统

- ✅ Ubuntu 18.04+
- ✅ Debian 10+
- ✅ CentOS 7+
- ✅ RHEL 7+
- ✅ Arch Linux

## 安装后访问

1. 浏览器访问：`http://YOUR_SERVER_IP:32451`
2. 使用默认账号登录：
   - 用户名：admin
   - 密码：admin
3. **立即修改默认密码！**

## 常用命令

```bash
# 启动服务
systemctl start wui

# 停止服务
systemctl stop wui

# 重启服务
systemctl restart wui

# 查看状态
systemctl status wui

# 查看日志
journalctl -u wui -f
```

## 创建第一个隧道

1. 登录面板
2. 点击 "Create Tunnel"
3. 配置入站：
   - 协议：SOCKS5
   - 端口：32000
   - 启用 UDP
4. 保存后添加出站配置
5. 启动隧道

## 开发环境

### 前置要求

- Node.js 18+
- Go 1.21+
- pnpm 8+

### 启动开发服务器

```bash
# 安装依赖
pnpm install

# 启动前端
cd apps/web
pnpm dev

# 启动后端（另一个终端）
cd apps/server
go run cmd/main.go
```

## 卸载

```bash
systemctl stop wui
systemctl disable wui
rm -rf /opt/wui
rm /etc/systemd/system/wui.service
systemctl daemon-reload
```

## 故障排查

### 端口被占用

```bash
# 检查端口
netstat -tulpn | grep 32451

# 杀掉占用进程
kill -9 <PID>
```

### 权限问题

```bash
# 确保 wui 进程有执行权限
chmod +x /opt/wui/wui
chmod +x /opt/wui/bin/xray
```

### 查看详细日志

```bash
journalctl -u wui -n 100 --no-pager
```
