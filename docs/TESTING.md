# WUI - 开发与测试指南

## 本地开发

### 1. 启动后端

```bash
cd apps/server

# 安装依赖
go mod download

# 设置测试配置
export WUI_CONFIG=/tmp/wui-test/config.json

# 运行
go run cmd/main.go
```

后端将在 `http://localhost:32451` 启动

### 2. 启动前端

```bash
cd apps/web

# 安装依赖
pnpm install

# 开发模式
pnpm dev
```

前端将在 `http://localhost:3000` 启动，自动代理到后端

## 生产构建

### 1. 构建所有平台

```bash
./scripts/build.sh
```

生成的文件在 `build/` 目录：
- `wui-linux-amd64-0.1.0.tar.gz` - Linux x86_64
- `wui-linux-arm64-0.1.0.tar.gz` - Linux ARM64
- `wui-linux-arm-0.1.0.tar.gz` - Linux ARM

### 2. 单平台构建

```bash
# 前端
cd apps/web
pnpm build

# 后端
cd apps/server
go build -o wui cmd/main.go
```

## 测试流程

### 1. 登录测试

- 访问：http://localhost:32451
- 用户名：admin
- 密码：admin

### 2. 创建隧道

1. 点击 "Create Tunnel"
2. 填写基本信息：
   - Name: Test Tunnel
   - Inbound Protocol: SOCKS5
   - Inbound Port: 32000
   - Enable UDP: ✓

3. 添加出站：
   - Protocol: VLESS
   - Address: your-server.com
   - Port: 443
   - Config:
   ```json
   {
     "uuid": "your-uuid-here",
     "encryption": "none",
     "flow": "xtls-rprx-vision",
     "streamSettings": {
       "network": "tcp",
       "security": "tls",
       "tlsSettings": {
         "serverName": "your-server.com"
       }
     }
   }
   ```

4. 点击 "Create Tunnel"

### 3. 启动隧道

1. 在隧道列表中找到刚创建的隧道
2. 点击 "Start" 按钮
3. 查看状态是否变为 "Active"

### 4. 测试连接

```bash
# 使用 SOCKS5 代理测试
curl --socks5 127.0.0.1:32000 http://www.google.com

# 或配置浏览器代理
# SOCKS5: 127.0.0.1:32000
```

### 5. 查看配置

点击 "View Config" 查看生成的 Xray 配置

### 6. 停止隧道

点击 "Stop" 按钮停止隧道

## API 测试

### 使用 curl 测试

```bash
# 登录
curl -X POST http://localhost:32451/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 使用返回的 token 测试其他接口
TOKEN="your-token-here"

# 获取隧道列表
curl http://localhost:32451/api/tunnels \
  -H "Authorization: Bearer $TOKEN"

# 创建隧道
curl -X POST http://localhost:32451/api/tunnels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test",
    "inboundProtocol": "socks5",
    "inboundPort": 32000,
    "udpEnabled": true,
    "outbounds": []
  }'
```

## 服务器安装测试

### 1. 准备服务器

- Ubuntu 20.04+ / Debian 10+ / CentOS 7+
- Root 权限
- 至少 512MB 内存

### 2. 上传文件

```bash
# 上传构建的文件到服务器
scp build/wui-linux-amd64-0.1.0.tar.gz root@your-server:/root/
```

### 3. 安装

```bash
# SSH 到服务器
ssh root@your-server

# 解压
tar -xzf wui-linux-amd64-0.1.0.tar.gz

# 运行安装脚本
cd wui-linux-amd64-0.1.0
chmod +x install/install.sh
./install/install.sh
```

### 4. 访问面板

浏览器访问：`http://your-server-ip:32451`

## 常见问题

### 端口被占用

```bash
# 检查端口
netstat -tulpn | grep 32451

# 杀掉进程
kill -9 <PID>
```

### 权限问题

```bash
chmod +x /opt/wui/wui
chmod +x /opt/wui/bin/xray
```

### 查看日志

```bash
# 系统日志
journalctl -u wui -f

# 应用日志
tail -f /opt/wui/logs/wui.log
```

### 数据库问题

```bash
# 数据库位置
/opt/wui/data/wui.db

# 备份数据库
cp /opt/wui/data/wui.db /opt/wui/data/wui.db.backup
```

## 性能测试

### 并发测试

```bash
# 使用 ab (Apache Bench)
ab -n 1000 -c 100 http://localhost:32451/api/tunnels

# 使用 wrk
wrk -t4 -c100 -d30s http://localhost:32451/api/tunnels
```

### 流量测试

```bash
# 使用 iperf3 测试带宽
iperf3 -c your-server -p 32000
```
