# LINE MAN OAuth Token 配置指南

## 概述

LINE MAN OAuth Token 功能实现了 OAuth 2.0 Client Credentials 认证流程,用于 LINE MAN 平台与 TTPOS 系统之间的安全通信。

> ⚠️ **临时方案声明**
> 
> 本实现为**临时过渡方案**,目的是快速支持 LINE MAN 平台接入。
> 
> **未来规划**:
> - 团队计划建设统一的**权限中心单点登录系统 (SSO)**
> - 届时将整合所有第三方平台认证逻辑 (包括 LINE MAN、Grab、Skootar 等)
> - 实现集中化的认证管理、统一的安全策略和更好的可维护性
> - 本模块的功能将迁移至统一权限中心
> 
> **开发者注意**:
> - 新增功能时请考虑未来迁移的兼容性
> - 避免过度复杂的定制化逻辑
> - 保持代码模块化,便于后续迁移

## 配置步骤

### 1. 环境变量配置

在 `.env` 文件中添加以下配置:

```bash
# LINE MAN OAuth 配置
# Platform 配置（用于平台级别的认证）
LINEMAN_PLATFORM_CLIENT_ID=your_platform_client_id_here
LINEMAN_PLATFORM_CLIENT_SECRET=your_platform_client_secret_here
LINEMAN_PLATFORM_SECRET_KEY=your_jwt_secret_key_here

# Partner 配置（用于商户级别的认证，支持多商户）
LINEMAN_PARTNER_DEFAULT_CLIENT_ID=your_partner_client_id_here
LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET=your_partner_client_secret_here

# 环境配置
LINEMAN_ENV=staging  # 或 production
```

### 2. 配置文件

配置会自动从 `manifest/config/config.tpl.yaml` 模板生成,包含以下内容:

```yaml
app:
  provider:
    lineman:
      # Platform 配置（平台级别）
      platform:
        clientId: "$LINEMAN_PLATFORM_CLIENT_ID"
        clientSecret: "$LINEMAN_PLATFORM_CLIENT_SECRET"
        secretKey: "$LINEMAN_PLATFORM_SECRET_KEY"
        environment: "$LINEMAN_ENV"
        timeout: "30s"
      # Partner 配置（商户级别，支持多商户）
      partner:
        $LINEMAN_PARTNER_DEFAULT_CLIENT_ID:
          clientId: "$LINEMAN_PARTNER_DEFAULT_CLIENT_ID"
          clientSecret: "$LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET"
          environment: "$LINEMAN_ENV"
          timeout: "60s"
```

**配置说明**:
- **platform**: 平台级别配置,用于 LINE MAN 平台与 TTPOS 系统的认证
- **partner**: 商户级别配置,支持多商户场景,每个商户使用独立的 client_id/client_secret
- 商户配置以 client_id 为 key,支持动态添加多个商户

## API 使用

### OAuth Token 接口

**请求**:

```http
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET
```

**响应**:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**错误响应**:

```json
{
  "error": "invalid_client",
  "error_description": "client_id 不匹配"
}
```

### 使用 Access Token

在后续 API 调用中,将 Token 添加到 Authorization Header:

```http
GET /api/v1/menu
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 技术实现

### 架构设计

```
┌─────────────────┐
│  LINE MAN       │
│  Platform       │
└────────┬────────┘
         │ POST /oauth2/token
         │ (client_id, client_secret)
         ▼
┌─────────────────────────────────────┐
│  Controller Layer                   │
│  lineman_v1_o_auth_token.go         │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Service Layer                      │
│  service.LinemanToken()             │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Logic Layer                        │
│  lineman_token/lineman_token.go     │
│  - GenerateToken()                  │
│  - ParseToken()                     │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Configuration                      │
│  lineman_token/config.go            │
│  - MustConfig()                     │
└─────────────────────────────────────┘
```

### 文件结构

```
ttpos-bmp/app/ttpos-takeout/
├── api/lineman/v1/
│   └── oauth.go                      # API 定义
├── internal/
│   ├── controller/lineman/
│   │   └── lineman_v1_o_auth_token.go  # Controller 实现
│   ├── logic/lineman_token/
│   │   ├── config.go                 # Platform 配置管理
│   │   ├── partner_config.go         # Partner 配置加载器
│   │   └── lineman_token.go          # Token 生成和验证
│   ├── service/
│   │   └── lineman_token.go          # Service 接口
│   └── model/
│       ├── conf/
│       │   └── provider.go           # 配置结构体
│       └── dto/lineman/
│           └── lineman.go            # DTO 定义
└── manifest/config/
    └── config.tpl.yaml               # 配置模板
```

## Token 规范

### JWT Claims

```json
{
  "client_id": "your_client_id",
  "iat": 1704614400,
  "exp": 1704618000
}
```

### Token 有效期

- 默认: 3600 秒 (1 小时)
- 可通过配置调整

### 签名算法

- HS256 (HMAC with SHA-256)

## 安全注意事项

1. **密钥管理**
   - `LINEMAN_SECRET_KEY` 必须保密
   - 定期轮换密钥
   - 不要将密钥提交到版本控制

2. **Token 验证**
   - 始终验证 Token 签名
   - 检查 Token 过期时间
   - 验证 client_id 匹配

3. **HTTPS**
   - 生产环境必须使用 HTTPS
   - 防止 Token 被窃取

## 测试

### 本地测试

```bash
# 启动服务
cd ttpos-bmp/app/ttpos-takeout
make run

# 测试 OAuth Token 接口
curl -X POST http://localhost:14031/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET"
```

### 单元测试

```bash
# 运行测试
go test ./internal/logic/lineman_token/... -v
```

## 故障排查

### 常见错误

1. **"根据 client_id 获取配置失败"**
   - 检查 Partner 配置中是否包含该 client_id
   - 确认环境变量 `LINEMAN_PARTNER_*_CLIENT_ID` 已正确设置
   - 检查配置文件中 partner 节点是否正确配置

2. **"client_secret 不匹配"**
   - 检查环境变量 `LINEMAN_PARTNER_*_CLIENT_SECRET` 是否正确
   - 确认请求中的 client_secret 与配置一致

3. **"获取 LINE MAN 平台配置失败"**
   - 检查配置文件是否正确生成
   - 确认环境变量 `LINEMAN_PLATFORM_*` 已正确设置
   - 检查配置路径 `app.provider.lineman.platform` 是否存在

4. **"Token 解析失败"**
   - 检查 Token 格式是否正确
   - 确认签名密钥一致
   - 验证 Token 是否已过期

5. **"未找到 LINE MAN Partner 配置"**
   - 确认 Partner 配置已添加到 `app.provider.lineman.partner` 节点
   - 检查 client_id 是否正确
   - 验证配置文件格式是否正确

## 参考资料

- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [GoFrame 文档](https://goframe.org)

## 版本历史

- **v2.13.1** (2026-01-07)
  - 初始实现 (临时方案)
  - 支持 OAuth 2.0 Client Credentials
  - JWT Token 生成和验证

## 维护者

- rikugun

## 迁移计划

### 统一权限中心迁移路线图

**Phase 1: 临时方案 (当前版本)**
- ✅ 实现 LINE MAN OAuth Token 功能
- ✅ 参考 Grab OAuth 架构
- ✅ 支持基本的认证流程

**Phase 2: 统一权限中心设计 (规划中)**
- 📋 设计统一 SSO 架构
- 📋 定义标准化认证接口
- 📋 规划多租户支持

**Phase 3: 迁移实施 (未来版本)**
- 📋 迁移 LINE MAN 认证到 SSO
- 📋 迁移 Grab 认证到 SSO
- 📋 迁移其他平台认证到 SSO
- 📋 废弃临时实现代码

**Phase 4: 优化完善**
- 📋 统一日志和监控
- 📋 增强安全策略
- 📋 性能优化

### 迁移注意事项

1. **数据兼容性**: 确保现有 Token 在迁移期间仍然有效
2. **API 兼容性**: 保持外部 API 接口不变,避免影响 LINE MAN 平台
3. **配置迁移**: 平滑迁移现有配置到统一配置中心
4. **监控告警**: 设置迁移监控,及时发现问题

## 相关文档

- [Grab OAuth 实现](../internal/logic/grab_token/README.md)
- [外送服务 README](../README.MD)
- [统一权限中心架构设计](./sso-architecture.md) (待创建)

