# WUI License Server

独立的 License 验证服务器，用于集中管理 WUI 实例的授权验证。

## 功能

- License Key 验证
- License 激活/解绑
- 心跳检测
- 使用统计
- 多实例管理

## 快速开始

### 安装

```bash
curl -fsSL https://your-domain.com/license/install.sh | bash
```

### 手动部署

```bash
# 编译
cd apps/license-server
go build -o wui-license cmd/main.go

# 运行
export LICENSE_SERVER_PORT=8080
export LICENSE_DB_PATH=/opt/wui-license/license.db
./wui-license
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| LICENSE_SERVER_PORT | 8080 | 服务端口 |
| LICENSE_DB_PATH | /opt/wui-license/license.db | 数据库路径 |

## API 接口

### 验证 License

```http
POST /api/v1/license/validate
Content-Type: application/json

{
  "licenseKey": "WUI-XXXX-XXXX-XXXX",
  "instanceId": "unique-instance-id",
  "machineId": "optional-machine-id",
  "domain": "optional-domain.com",
  "ipAddress": "1.2.3.4"
}
```

响应:
```json
{
  "valid": true,
  "message": "License is valid",
  "type": "personal",
  "plan": "pro",
  "maxTunnels": 10,
  "maxUsers": 1,
  "maxTraffic": 107374182400,
  "features": "[\"websocket\",\"cluster\"]",
  "expiresAt": "2025-12-31T23:59:59Z"
}
```

### 激活 License

```http
POST /api/v1/license/activate
Content-Type: application/json

{
  "licenseKey": "WUI-XXXX-XXXX-XXXX",
  "instanceId": "unique-instance-id",
  "machineId": "optional-machine-id",
  "domain": "optional-domain.com",
  "ipAddress": "1.2.3.4"
}
```

### 心跳上报

```http
POST /api/v1/license/heartbeat
Content-Type: application/json

{
  "licenseKey": "WUI-XXXX-XXXX-XXXX",
  "instanceId": "unique-instance-id",
  "version": "0.1.0",
  "tunnelCount": 5,
  "userCount": 10,
  "cpuUsage": 25.5,
  "memUsage": 60.2,
  "diskUsage": 45.8,
  "ipAddress": "1.2.3.4",
  "domain": "example.com"
}
```

### 获取 License 信息

```http
GET /api/v1/license/info?licenseKey=WUI-XXXX-XXXX-XXXX
```

### 解绑 License

```http
POST /api/v1/license/deactivate
Content-Type: application/json

{
  "licenseKey": "WUI-XXXX-XXXX-XXXX",
  "instanceId": "unique-instance-id"
}
```

## License 类型

| 类型 | Max Tunnels | Max Users | Max Traffic | 特性 |
|------|-------------|-----------|-------------|------|
| trial | 3 | 1 | 10GB | 基础 |
| personal | 10 | 1 | 100GB | WebSocket |
| team | 50 | 10 | 500GB | WebSocket, Cluster |
| enterprise | Unlimited | Unlimited | Unlimited | 全部 |

## 安全性

- License Key 使用 SHA256 哈希存储
- 支持实例绑定（Machine ID / IP / Domain）
- 验证频率限制
- HTTPS 传输（推荐）

## 管理

### 创建 License

```bash
# 使用 SQLite
sqlite3 /opt/wui-license/license.db

INSERT INTO licenses (
  license_key, license_hash, type, plan, 
  max_tunnels, max_users, max_traffic, 
  status, customer_name, customer_email
) VALUES (
  'WUI-XXXX-XXXX-XXXX',
  '<sha256-hash>',
  'personal',
  'pro',
  10, 1, 107374182400,
  'inactive',
  'Customer Name',
  'customer@example.com'
);
```

### 查看所有 License

```bash
sqlite3 /opt/wui-license/license.db "SELECT * FROM licenses;"
```

### 查看 Heartbeat 记录

```bash
sqlite3 /opt/wui-license/license.db "SELECT * FROM heartbeats ORDER BY created_at DESC LIMIT 100;"
```

## 日志

```bash
# 查看日志
journalctl -u wui-license -f

# 应用日志
tail -f /opt/wui-license/logs/wui-license.log
```

## 架构

```
┌──────────────┐     HTTP/HTTPS     ┌──────────────────┐
│              │ ──────────────────>│                  │
│  WUI 主服务   │                   │  License Server  │
│              │<────────────────── │                  │
└──────────────┘                   └──────────────────┘
                                          │
                                          ▼
                                   ┌──────────────────┐
                                   │    SQLite DB     │
                                   └──────────────────┘
```

## 高可用

对于生产环境，建议：

1. 使用 PostgreSQL 替代 SQLite
2. 部署多个实例 + 负载均衡
3. 添加 Redis 缓存
4. 启用 HTTPS

## 故障排查

### License 验证失败

1. 检查 License Key 是否正确
2. 检查实例绑定是否匹配
3. 检查是否过期
4. 检查网络连接

### 服务无法启动

1. 检查端口是否被占用
2. 检查数据库文件权限
3. 检查日志文件

## License

MIT
