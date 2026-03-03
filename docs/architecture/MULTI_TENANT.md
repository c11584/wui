# WUI 多租户架构设计

## 概述

本文档描述 WUI 商业化的多租户架构设计，采用**共享数据库 + UserID 隔离**模式。

## 架构决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 数据隔离 | 共享数据库 | 简单高效，适合中小规模 |
| License服务 | 独立服务 | 集中管理多实例，支持集群部署 |
| 验证频率 | 启动+每日 | 平衡安全和性能 |
| 邮件服务 | 163 SMTP | 国内可用，成本低 |

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户浏览器                               │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    WUI 主服务 (Go + Gin)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Auth API     │  │ Tunnel API   │  │ User API     │          │
│  │ (JWT+RBAC)   │  │ (CRUD)       │  │ (管理)       │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              License 验证客户端                           │   │
│  │         (启动验证 + 每日后台验证)                          │   │
│  └─────────────────────────┬───────────────────────────────┘   │
└────────────────────────────┼────────────────────────────────────┘
                             │ HTTP/HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                 License 验证服务 (独立部署)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ License API  │  │ Heartbeat    │  │ Admin API    │          │
│  │ (验证/激活)  │  │ (心跳检测)   │  │ (密钥管理)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                            │                                    │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              License 数据库 (SQLite/PostgreSQL)          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## 数据模型设计

### 1. User 模型（扩展）

```go
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    // 基本信息
    Username  string `gorm:"uniqueIndex;not null" json:"username"`
    Email     string `gorm:"uniqueIndex;not null" json:"email"`
    Password  string `gorm:"not null" json:"-"`
    
    // 权限和状态
    Role      string `gorm:"default:'user';index" json:"role"` // admin, user
    Status    string `gorm:"default:'active';index" json:"status"` // active, suspended, deleted
    
    // 配额
    MaxTunnels    int   `gorm:"default:5" json:"maxTunnels"`
    MaxTraffic    int64 `gorm:"default:107374182400" json:"maxTraffic"` // 100GB
    
    // 时间戳
    LastLoginAt   *time.Time `json:"lastLoginAt"`
    LastLoginIP   string     `gorm:"type:varchar(45)" json:"lastLoginIp"`
    
    // 2FA
    TwoFactorEnabled bool   `gorm:"default:false" json:"twoFactorEnabled"`
    TwoFactorSecret  string `gorm:"type:varchar(32)" json:"-"`
    
    // 密码重置
    ResetToken     string     `gorm:"type:varchar(64)" json:"-"`
    ResetTokenExp  *time.Time `json:"-"`
    
    // License 关联
    LicenseKey    string `gorm:"type:varchar(64);index" json:"licenseKey"`
    LicenseExpire *time.Time `json:"licenseExpire"`
}
```

### 2. Tunnel 模型（扩展）

```go
type Tunnel struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    // 所属用户（多租户）
    UserID uint `gorm:"not null;index" json:"userId"`
    
    // ... 其他字段保持不变 ...
}
```

### 3. AuditLog 模型（新增）

```go
type AuditLog struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `gorm:"index" json:"createdAt"`
    
    // 用户信息
    UserID   uint   `gorm:"index" json:"userId"`
    Username string `gorm:"type:varchar(255)" json:"username"`
    
    // 操作信息
    Action   string `gorm:"type:varchar(100);index" json:"action"` // login, create_tunnel, delete_tunnel, etc.
    Resource string `gorm:"type:varchar(255)" json:"resource"` // tunnels, users, settings
    ResourceID uint `json:"resourceId"`
    
    // 详情
    Detail   string `gorm:"type:text" json:"detail"` // JSON 格式的详细信息
    
    // 请求信息
    IPAddress string `gorm:"type:varchar(45)" json:"ipAddress"`
    UserAgent string `gorm:"type:text" json:"userAgent"`
    
    // 结果
    Success bool `gorm:"default:true" json:"success"`
    Error   string `gorm:"type:text" json:"error"`
}
```

### 4. License 模型（License 服务）

```go
type License struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    
    // 密钥信息
    LicenseKey  string `gorm:"uniqueIndex;not null" json:"licenseKey"` // 格式: WUI-XXXX-XXXX-XXXX
    LicenseHash string `gorm:"uniqueIndex" json:"-"` // SHA256 哈希
    
    // 类型
    Type string `gorm:"not null;index" json:"type"` // trial, personal, team, enterprise
    Plan string `gorm:"not null" json:"plan"` // basic, pro, enterprise
    
    // 功能限制
    MaxTunnels    int   `gorm:"default:5" json:"maxTunnels"`
    MaxUsers      int   `gorm:"default:1" json:"maxUsers"`
    MaxTraffic    int64 `gorm:"default:107374182400" json:"maxTraffic"` // bytes
    Features      string `gorm:"type:text" json:"features"` // JSON array: ["websocket", "cluster"]
    
    // 有效期
    ExpiresAt     *time.Time `json:"expiresAt"`
    Lifetime      bool       `gorm:"default:false" json:"lifetime"`
    
    // 绑定信息
    InstanceID    string `gorm:"type:varchar(64);index" json:"instanceId"` // 绑定的实例 ID
    BindIP        string `gorm:"type:varchar(45)" json:"bindIp"`
    BindDomain    string `gorm:"type:varchar(255)" json:"bindDomain"`
    BindMachineID string `gorm:"type:varchar(64)" json:"bindMachineId"` // 机器唯一标识
    
    // 状态
    Status        string `gorm:"default:'inactive';index" json:"status"` // inactive, active, suspended, expired
    ActivatedAt   *time.Time `json:"activatedAt"`
    LastCheckAt   *time.Time `json:"lastCheckAt"`
    
    // 客户信息
    CustomerName  string `gorm:"type:varchar(255)" json:"customerName"`
    CustomerEmail string `gorm:"type:varchar(255)" json:"customerEmail"`
    OrderID       string `gorm:"type:varchar(100)" json:"orderId"`
    
    // 备注
    Remark        string `gorm:"type:text" json:"remark"`
}
```

### 5. Heartbeat 模型（License 服务）

```go
type Heartbeat struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `gorm:"index" json:"createdAt"`
    
    LicenseKey  string `gorm:"index" json:"licenseKey"`
    InstanceID  string `gorm:"index" json:"instanceId"`
    
    // 状态
    Status      string `json:"status"` // ok, warning, error
    Version     string `json:"version"`
    TunnelCount int    `json:"tunnelCount"`
    UserCount   int    `json:"userCount"`
    
    // 资源使用
    CpuUsage    float64 `json:"cpuUsage"`
    MemUsage    float64 `json:"memUsage"`
    DiskUsage   float64 `json:"diskUsage"`
    
    // 网络
    IPAddress   string `gorm:"type:varchar(45)" json:"ipAddress"`
    Domain      string `json:"domain"`
}
```

## License 密钥格式

```
WUI-XXXX-XXXX-XXXX

示例: WUI-A1B2-C3D4-E5F6
```

**生成规则**:
- 前缀: `WUI`
- 3 组 4 字符，用 `-` 分隔
- 字符集: `ABCDEFGHJKLMNPQRSTUVWXYZ23456789` (去除易混淆字符)
- 总长度: 15 字符

## License 验证流程

### 1. 启动验证

```
WUI 主服务启动
    ↓
检查本地缓存的 License
    ↓ (不存在或过期)
调用 License 服务验证 API
    ↓
License 服务验证:
  - 密钥是否有效
  - 是否过期
  - 是否绑定到当前实例
  - 状态是否正常
    ↓
返回验证结果 + 功能限制
    ↓
主服务缓存结果 (24小时)
    ↓
启动完成
```

### 2. 每日验证

```
后台定时任务 (每 24 小时)
    ↓
调用 License 服务心跳 API
    ↓
发送:
  - License Key
  - Instance ID
  - 使用统计
    ↓
License 服务更新:
  - LastCheckAt
  - 使用统计
    ↓
返回:
  - 验证状态
  - 最新的功能限制
    ↓
主服务更新缓存
```

### 3. 功能检查

```
用户请求创建隧道
    ↓
检查 License 限制:
  - MaxTunnels
  - MaxTraffic
  - Features
    ↓
未超限 → 允许操作
超限 → 返回错误提示
```

## API 端点设计

### License 服务 API

```
POST /api/v1/license/validate     # 验证 License
POST /api/v1/license/activate     # 激活 License
POST /api/v1/license/heartbeat    # 心跳上报
POST /api/v1/license/deactivate   # 解绑 License
GET  /api/v1/license/info         # 获取 License 信息
```

### 主服务 API (新增)

```
POST /api/auth/register           # 用户注册
POST /api/auth/forgot-password    # 忘记密码
POST /api/auth/reset-password     # 重置密码
GET  /api/users                   # 用户列表 (admin)
GET  /api/users/:id               # 用户详情 (admin)
PUT  /api/users/:id               # 更新用户 (admin)
DELETE /api/users/:id             # 删除用户 (admin)
GET  /api/audit-logs              # 审计日志 (admin)
POST /api/license/activate        # 激活 License
GET  /api/license/info            # 获取当前 License 信息
```

## 安全考虑

1. **License 密钥安全**
   - 不在日志中明文记录
   - 传输时使用 HTTPS
   - 存储时使用 SHA256 哈希

2. **实例绑定**
   - 支持绑定机器 ID、IP、域名
   - 绑定后其他实例无法使用
   - 提供解绑接口（需验证）

3. **频率限制**
   - 验证 API: 1 次/分钟
   - 心跳 API: 1 次/小时
   - 防止暴力破解

4. **离线模式**
   - 本地缓存 24 小时
   - 网络故障时允许运行
   - 超过 7 天未验证则降级

## 数据库迁移策略

1. **添加新字段**
   - 使用 GORM AutoMigrate
   - 默认值保证兼容性

2. **数据迁移**
   - 现有 admin 用户保留
   - 添加默认配额
   - Tunnel 添加 UserID (默认为 admin)

3. **回滚方案**
   - 保留原字段
   - 新增字段可选
   - 分阶段迁移

## 实施顺序

1. ✅ 创建数据模型
2. ✅ 实现 License 服务
3. ✅ 扩展 User 模型
4. ✅ 添加 RBAC 中间件
5. ✅ 实现用户管理 API
6. ✅ 实现审计日志
7. ✅ 前端页面开发
8. ✅ 集成测试
9. ✅ 部署上线
