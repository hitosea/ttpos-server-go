# Logto 集成方案详细指南

> 👤 **受众**: 架构师、后端开发工程师  
> 📖 **用途**: Logto 集成到 TTPOS 项目的详细技术方案和实现指南

---

## 📋 目录

- [Logto 工作原理](#logto-工作原理)
- [架构对比](#架构对比)
- [集成方案设计](#集成方案设计)
- [实施步骤](#实施步骤)
- [代码示例](#代码示例)
- [多租户实现](#多租户实现)
- [第三方服务接入](#第三方服务接入)
- [注意事项](#注意事项)

---

## Logto 工作原理

### 核心概念

Logto 是一个基于 **OIDC (OpenID Connect)** 标准的身份认证服务，采用 **授权码流程 (Authorization Code Flow)** 实现用户认证。

### 认证流程

```mermaid
sequenceDiagram
    participant Client as 客户端应用
    participant Logto as Logto 服务
    participant User as 用户
    participant API as TTPOS API

    Client->>Logto: 1. 重定向到登录页
    Logto->>User: 2. 显示登录界面
    User->>Logto: 3. 输入账号密码
    Logto->>Logto: 4. 验证用户身份
    Logto->>Client: 5. 重定向回应用（带授权码）
    Client->>Logto: 6. 用授权码换取 Token
    Logto->>Client: 7. 返回 Access Token + ID Token
    Client->>API: 8. 携带 Token 访问 API
    API->>Logto: 9. 验证 Token（可选）
    Logto->>API: 10. 返回 Token 信息
    API->>Client: 11. 返回业务数据
```

### Token 类型

1. **ID Token**：包含用户身份信息（JWT 格式）
   - `sub`: 用户 ID
   - `email`: 邮箱
   - `organizations`: 用户所属组织列表（多租户）
   - `organization_roles`: 组织角色映射

2. **Access Token**：用于访问受保护的 API 资源
   - 可以是 JWT 或 opaque token
   - 包含权限范围（scopes）

3. **Refresh Token**：用于刷新 Access Token
   - 长期有效，用于获取新的 Access Token

### 多租户支持

Logto 通过 **Organization（组织）** 概念实现多租户：

- 每个商家对应一个 Organization
- 用户可以被添加到多个 Organization
- ID Token 中包含 `organizations` 声明，列出用户所属的所有组织
- 可以通过 `organization_id` 参数切换组织上下文

---

## 架构对比

### 当前架构（自建 JWT）

```
┌─────────────┐
│   客户端    │
└──────┬──────┘
       │ 1. 登录请求（username + password）
       ▼
┌─────────────────┐
│  TTPOS API      │
│  /api/v1/login  │
└──────┬──────────┘
       │ 2. 查询 company_staff
       ▼
┌─────────────────┐
│   SaaS 数据库   │
│ company_staff   │
└──────┬──────────┘
       │ 3. 查询 staff
       ▼
┌─────────────────┐
│   商家数据库    │
│     staff       │
└──────┬──────────┘
       │ 4. 生成 JWT Token
       ▼
┌─────────────────┐
│   客户端        │
│  (Token 存储)   │
└─────────────────┘
```

**特点**：
- ✅ 简单直接，无外部依赖
- ✅ 完全控制 Token 结构
- ❌ 需要自行实现多租户逻辑
- ❌ 需要自行管理用户和密码

### Logto 架构

```
┌─────────────┐
│   客户端    │
└──────┬──────┘
       │ 1. 重定向到 Logto
       ▼
┌─────────────────┐
│   Logto 服务    │
│  (独立部署)     │
│  - 用户管理     │
│  - 组织管理     │
│  - 认证流程     │
└──────┬──────────┘
       │ 2. 用户登录
       ▼
┌─────────────────┐
│   Logto 数据库  │
│  - users        │
│  - organizations│
│  - relations    │
└──────┬──────────┘
       │ 3. 返回 Token
       ▼
┌─────────────┐
│   客户端    │
│ (ID Token + │
│ Access Token)│
└──────┬──────┘
       │ 4. 携带 Token 访问 API
       ▼
┌─────────────────┐
│  TTPOS API      │
│  (验证 Token)   │
└──────┬──────────┘
       │ 5. 查询业务数据
       ▼
┌─────────────────┐
│   商家数据库    │
└─────────────────┘
```

**特点**：
- ✅ 标准化认证流程（OIDC）
- ✅ 内置多租户支持
- ✅ 用户管理功能完善
- ✅ 支持社交登录、MFA 等高级功能
- ⚠️ 需要独立部署 Logto 服务
- ⚠️ 需要同步用户和组织数据

---

## 集成方案设计

### 方案一：完全替换（推荐用于新项目）

**架构**：
- 所有认证由 Logto 处理
- TTPOS API 只验证 Logto Token
- 用户和组织数据存储在 Logto

**优点**：
- ✅ 完全标准化
- ✅ 功能丰富
- ✅ 维护成本低

**缺点**：
- ❌ 需要大量改造现有代码
- ❌ 需要数据迁移
- ❌ 用户数据在 Logto，业务数据在 TTPOS

### 方案二：混合方案（推荐用于现有项目）

**架构**：
- Logto 负责用户认证和 Token 发放
- TTPOS 保留业务用户数据（staff 表）
- 通过 `user_id` 关联 Logto 用户和 TTPOS 员工

**数据同步**：
```
Logto User (user_id: "logto_xxx")
    ↓ (通过 user_id 关联)
TTPOS Staff (logto_user_id: "logto_xxx")
```

**优点**：
- ✅ 改动相对较小
- ✅ 保留现有业务逻辑
- ✅ 可以逐步迁移

**缺点**：
- ⚠️ 需要维护两套用户数据
- ⚠️ 需要同步机制

### 方案三：桥接方案（最推荐）

**架构**：
- Logto 作为认证服务，处理登录和 Token 发放
- TTPOS 保留完整的用户和组织数据
- 登录后，TTPOS 根据 Logto 用户信息查询本地数据，生成自定义 Token

**流程**：
```
1. 用户通过 Logto 登录
2. Logto 返回 ID Token（包含 user_id, organizations）
3. TTPOS 接收 ID Token，验证后：
   - 根据 user_id 查询本地 staff 表
   - 根据 organizations 查询用户可访问的商家列表
   - 生成 TTPOS 自定义 Token（保持现有 Token 结构）
4. 客户端使用 TTPOS Token 访问 API
```

**优点**：
- ✅ 最小改动现有代码
- ✅ 保持现有 Token 结构
- ✅ 完全控制业务逻辑
- ✅ 可以逐步迁移

**缺点**：
- ⚠️ 需要维护 Logto 和 TTPOS 的用户关联

---

## 实施步骤

### 阶段一：部署 Logto（1-2天）

#### 1.1 Docker 部署

```bash
# 创建 docker-compose.yml
cat > docker-compose.logto.yml <<EOF
version: '3.8'

services:
  logto:
    image: ghcr.io/logto-io/logto:latest
    ports:
      - "3001:3001"
    environment:
      - DB_URL=postgresql://logto:password@postgres:5432/logto
      - ENDPOINT=http://localhost:3001
      - ADMIN_ENDPOINT=http://localhost:3002
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: logto
      POSTGRES_PASSWORD: password
      POSTGRES_DB: logto
    volumes:
      - logto_db:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - logto_redis:/data

volumes:
  logto_db:
  logto_redis:
EOF

# 启动服务
docker-compose -f docker-compose.logto.yml up -d
```

#### 1.2 配置 Logto

1. 访问管理后台：`http://localhost:3002`
2. 创建初始管理员账号
3. 配置应用：
   - 应用名称：TTPOS
   - 应用类型：Traditional Web Application
   - 重定向 URI：`http://localhost:8080/api/v1/auth/logto/callback`
   - 登出重定向 URI：`http://localhost:8080/api/v1/auth/logout`

### 阶段二：集成 Logto SDK（2-3天）

#### 2.1 安装 Go SDK

```bash
go get github.com/logto-io/go
```

#### 2.2 创建 Logto 客户端

```go
// main/pkg/logto/client.go
package logto

import (
    "github.com/logto-io/go/client"
)

type Client struct {
    client *logto.Client
}

func NewClient(config Config) (*Client, error) {
    logtoClient, err := logto.NewClient(
        logto.Config{
            Endpoint:        config.Endpoint,        // http://localhost:3001
            AppId:           config.AppId,           // 从 Logto 控制台获取
            AppSecret:       config.AppSecret,       // 从 Logto 控制台获取
            RedirectUri:     config.RedirectUri,     // http://localhost:8080/api/v1/auth/logto/callback
            PostLogoutRedirectUri: config.PostLogoutRedirectUri,
            Scopes:          []string{"openid", "profile", "email", "organizations"},
        },
    )
    if err != nil {
        return nil, err
    }
    
    return &Client{client: logtoClient}, nil
}
```

#### 2.3 创建认证处理器

```go
// main/app/api/v1/auth/logto_auth.go
package auth

import (
    "github.com/gin-gonic/gin"
    "ttpos-server-go/app/api/helper"
    "ttpos-server-go/app/constant"
    "ttpos-server-go/pkg/logto"
)

type LogtoAuthHandler struct {
    logtoClient *logto.Client
    authSrv     service.IAuthSrv
}

// InitiateLogin 发起登录（重定向到 Logto）
func (h *LogtoAuthHandler) InitiateLogin(c *gin.Context) {
    // 获取重定向 URL
    redirectUrl, err := h.logtoClient.GetSignInUrl(c.Request.Context())
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeServerError, err)
        return
    }
    
    // 重定向到 Logto 登录页
    c.Redirect(302, redirectUrl)
}

// Callback 登录回调
func (h *LogtoAuthHandler) Callback(c *gin.Context) {
    // 获取授权码
    code := c.Query("code")
    if code == "" {
        helper.Fail(c, constant.CodeParamError, "缺少授权码")
        return
    }
    
    // 用授权码换取 Token
    tokenResponse, err := h.logtoClient.SignIn(code, c.Request.Context())
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeServerError, err)
        return
    }
    
    // 解析 ID Token
    idTokenClaims, err := h.logtoClient.VerifyIdToken(tokenResponse.IdToken)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeTokenInvalid, err)
        return
    }
    
    // 获取用户信息
    logtoUserId := idTokenClaims.Subject
    organizations := idTokenClaims.Organizations // 用户所属的组织列表
    
    // 查询 TTPOS 本地用户数据
    // ... 业务逻辑 ...
    
    // 生成 TTPOS Token（保持现有 Token 结构）
    // ... 生成 Token ...
    
    // 返回响应
    helper.Success(c, gin.H{
        "token":         ttpToken,
        "refresh_token": ttpRefreshToken,
        "companies":     companies, // 可访问的商家列表
    })
}
```

### 阶段三：多租户实现（2-3天）

#### 3.1 在 Logto 中创建组织

```go
// main/app/service/logto_org.go
package service

import (
    "github.com/logto-io/go/client"
)

type LogtoOrgService struct {
    logtoClient *logto.Client
}

// CreateOrganization 为商家创建 Logto 组织
func (s *LogtoOrgService) CreateOrganization(companyUuid uint64, companyName string) (string, error) {
    // 调用 Logto Management API 创建组织
    org, err := s.logtoClient.CreateOrganization(companyName, map[string]string{
        "company_uuid": fmt.Sprintf("%d", companyUuid),
    })
    if err != nil {
        return "", err
    }
    
    return org.Id, nil
}

// AddUserToOrganization 将用户添加到组织
func (s *LogtoOrgService) AddUserToOrganization(logtoUserId string, organizationId string, role string) error {
    return s.logtoClient.AddUserToOrganization(logtoUserId, organizationId, role)
}
```

#### 3.2 用户-组织关联

```go
// main/app/model/logto_user.go
package model

// LogtoUser TTPOS 用户与 Logto 用户的关联表
type LogtoUser struct {
    BaseModel
    LogtoUserId   string `gorm:"column:logto_user_id;type:varchar(255);uniqueIndex;comment:Logto 用户ID" json:"logto_user_id"`
    StaffUuid     uint64 `gorm:"column:staff_uuid;type:bigint(20);index;comment:TTPOS 员工ID" json:"staff_uuid"`
    CompanyUuid    uint64 `gorm:"column:company_uuid;type:bigint(20);index;comment:商家ID" json:"company_uuid"`
    LogtoOrgId    string `gorm:"column:logto_org_id;type:varchar(255);index;comment:Logto 组织ID" json:"logto_org_id"`
}
```

### 阶段四：Token 验证中间件改造（1-2天）

#### 4.1 支持两种 Token

```go
// main/middleware/auth.go

func ParseJwt(c *gin.Context, authHeader string, authSrv service.IAuthSrv, dbm *database.DBManager) {
    parts := strings.SplitN(authHeader, " ", 2)
    if !(len(parts) == 2 && parts[0] == "Bearer") {
        helper.Fail(c, constant.CodeTokenInvalid, "登录失效，请重新登录")
        c.Abort()
        return
    }
    
    tokenString := parts[1]
    
    // 尝试解析为 TTPOS Token（保持向后兼容）
    claims, err := auth.ParseToken(tokenString, config.JWT.Secret)
    if err == nil {
        // 使用现有逻辑
        parseTTPOSToken(c, claims, authSrv, dbm)
        return
    }
    
    // 尝试解析为 Logto Token
    logtoClaims, err := logtoClient.VerifyIdToken(tokenString)
    if err == nil {
        // 使用 Logto Token 逻辑
        parseLogtoToken(c, logtoClaims, authSrv, dbm)
        return
    }
    
    helper.Fail(c, constant.CodeTokenInvalid, "登录失效，请重新登录")
    c.Abort()
}

func parseLogtoToken(c *gin.Context, logtoClaims *logto.IdTokenClaims, authSrv service.IAuthSrv, dbm *database.DBManager) {
    // 从 Logto Token 中获取用户信息
    logtoUserId := logtoClaims.Subject
    organizations := logtoClaims.Organizations
    
    // 查询 TTPOS 本地用户数据
    logtoUserRepo := repository.NewLogtoUserRepo(dbm.GetDB(constant.DefaultDB))
    logtoUser := logtoUserRepo.GetByLogtoUserId(logtoUserId)
    
    if logtoUser.LogtoUserId == "" {
        helper.Fail(c, constant.CodeTokenInvalid, "用户不存在")
        c.Abort()
        return
    }
    
    // 获取当前组织（从请求头或 Token 中获取）
    currentOrgId := c.GetHeader("X-Organization-Id")
    if currentOrgId == "" && len(organizations) > 0 {
        currentOrgId = organizations[0] // 默认使用第一个组织
    }
    
    // 查询对应的商家和员工信息
    // ... 业务逻辑 ...
    
    // 设置上下文
    c.Set(jwt.CompanyUuid, companyUuid)
    c.Set(jwt.StaffUuid, staffUuid)
    // ... 其他字段 ...
}
```

---

## 代码示例

### 完整的登录流程

```go
// main/app/api/v1/auth/logto_auth.go

package auth

import (
    "context"
    "github.com/gin-gonic/gin"
    "ttpos-server-go/app/api/helper"
    "ttpos-server-go/app/constant"
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/logto"
)

type LogtoAuthHandler struct {
    logtoClient *logto.Client
    authSrv     service.IAuthSrv
    logtoOrgSrv service.ILogtoOrgSrv
}

// InitiateLogin 发起登录
func (h *LogtoAuthHandler) InitiateLogin(c *gin.Context) {
    // 可选：指定组织 ID（用于直接登录到特定商家）
    orgId := c.Query("organization_id")
    
    redirectUrl, err := h.logtoClient.GetSignInUrl(c.Request.Context(), logto.SignInOptions{
        OrganizationId: orgId,
    })
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeServerError, err)
        return
    }
    
    c.Redirect(302, redirectUrl)
}

// Callback 登录回调
func (h *LogtoAuthHandler) Callback(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        helper.Fail(c, constant.CodeParamError, "缺少授权码")
        return
    }
    
    // 1. 用授权码换取 Token
    tokenResponse, err := h.logtoClient.SignIn(code, c.Request.Context())
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeServerError, err)
        return
    }
    
    // 2. 验证 ID Token
    idTokenClaims, err := h.logtoClient.VerifyIdToken(tokenResponse.IdToken)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeTokenInvalid, err)
        return
    }
    
    logtoUserId := idTokenClaims.Subject
    organizations := idTokenClaims.Organizations
    
    // 3. 查询 TTPOS 用户关联
    saasDB := h.authSrv.GetDBManager().GetDB(constant.DefaultDB)
    logtoUserRepo := repository.NewLogtoUserRepo(saasDB)
    logtoUsers := logtoUserRepo.GetByLogtoUserId(logtoUserId)
    
    if len(logtoUsers) == 0 {
        helper.Fail(c, constant.CodeTokenInvalid, "用户未关联到任何商家")
        return
    }
    
    // 4. 构建可访问的商家列表
    companies := make([]resp.CompanyInfo, 0)
    for _, lu := range logtoUsers {
        // 验证组织是否在 Logto Token 中
        if !contains(organizations, lu.LogtoOrgId) {
            continue
        }
        
        // 查询商家信息
        companyDB := h.authSrv.GetDBManager().GetDB(lu.CompanyUuid)
        companyRepo := repository.NewCompanyRepo(companyDB)
        company, _ := companyRepo.GetCompanyInfoByUuid(lu.CompanyUuid)
        
        if company != nil && !company.IsExpired() && !company.IsException() {
            companies = append(companies, resp.CompanyInfo{
                CompanyUuid: lu.CompanyUuid,
                CompanyName: company.CompanyName,
                StaffUuid:   lu.StaffUuid,
                LogtoOrgId:  lu.LogtoOrgId,
            })
        }
    }
    
    if len(companies) == 0 {
        helper.Fail(c, constant.CodeTokenInvalid, "没有可访问的商家")
        return
    }
    
    // 5. 生成 TTPOS Token（使用第一个商家）
    currentCompany := companies[0]
    staffRepo := repository.NewStaffRepo(h.authSrv.GetDBManager().GetDB(currentCompany.CompanyUuid))
    staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(currentCompany.StaffUuid))
    
    claims := auth.Claims{
        Source:         c.Query("source"), // 从查询参数获取
        CompanyUuid:    currentCompany.CompanyUuid,
        StaffUuid:      currentCompany.StaffUuid,
        DeviceId:       c.Query("device_id"),
        // ... 其他字段
    }
    
    token, _ := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
    refreshToken, _ := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
    
    // 6. 返回响应
    helper.Success(c, resp.LoginResp{
        Token:        token,
        RefreshToken: refreshToken,
        UserInfo: resp.UserInfo{
            LogtoUserId: logtoUserId,
            Username:    staff.Username,
        },
        Companies: companies,
    })
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

---

## 多租户实现

### 组织切换

```go
// main/app/api/v1/auth/logto_auth.go

// SwitchOrganization 切换组织（商家）
func (h *LogtoAuthHandler) SwitchOrganization(c *gin.Context) {
    var req req.SwitchOrgReq
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.Fail(c, constant.CodeParamError, err.Error())
        return
    }
    
    // 获取当前用户信息
    logtoUserId := c.GetString(jwt.LogtoUserId)
    if logtoUserId == "" {
        helper.Fail(c, constant.CodeTokenInvalid, "用户未登录")
        return
    }
    
    // 验证用户是否有权限访问该组织
    saasDB := h.authSrv.GetDBManager().GetDB(constant.DefaultDB)
    logtoUserRepo := repository.NewLogtoUserRepo(saasDB)
    logtoUser := logtoUserRepo.GetByLogtoUserIdAndOrgId(logtoUserId, req.OrganizationId)
    
    if logtoUser.LogtoUserId == "" {
        helper.Fail(c, constant.CodeAccessDenied, "无权访问该商家")
        return
    }
    
    // 重新生成 Token（切换到新商家）
    claims := auth.Claims{
        Source:      c.GetString(jwt.Source),
        CompanyUuid: logtoUser.CompanyUuid,
        StaffUuid:   logtoUser.StaffUuid,
        DeviceId:    c.GetString(jwt.DeviceId),
        // ... 其他字段
    }
    
    token, _ := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
    refreshToken, _ := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
    
    helper.Success(c, resp.SwitchOrgResp{
        Token:        token,
        RefreshToken: refreshToken,
    })
}
```

---

## 第三方服务接入

### 场景说明

当第三方服务（如 ERP 系统、BI 分析平台、支付服务等）需要接入 TTPOS API 时，需要使用 **OAuth 2.0 Client Credentials Flow（客户端凭证流程）** 进行服务到服务（Service-to-Service）认证。

### 认证流程

```mermaid
sequenceDiagram
    participant ThirdParty as 第三方服务
    participant Logto as Logto 服务
    participant TTPOS as TTPOS API

    ThirdParty->>Logto: 1. 使用 Client ID + Secret 请求 Token
    Note over ThirdParty,Logto: POST /oidc/token<br/>grant_type=client_credentials
    Logto->>Logto: 2. 验证客户端凭证
    Logto->>ThirdParty: 3. 返回 Access Token
    ThirdParty->>TTPOS: 4. 携带 Access Token 访问 API
    TTPOS->>Logto: 5. 验证 Access Token
    Logto->>TTPOS: 6. 返回 Token 信息（包含 scopes）
    TTPOS->>TTPOS: 7. 验证权限范围（scopes）
    TTPOS->>ThirdParty: 8. 返回业务数据
```

### 实施步骤

#### 步骤一：在 Logto 中注册第三方应用

1. **登录 Logto 管理后台**
   - 访问：`http://localhost:3002`

2. **创建 Machine-to-Machine (M2M) 应用**
   - 应用类型：选择 "Machine-to-Machine"
   - 应用名称：例如 "ERP Integration Service"
   - 应用描述：第三方服务的描述

3. **配置 API 资源权限**
   - 在应用详情页，导航至 "API Resources" 选项卡
   - 添加 TTPOS API 资源
   - 选择所需的权限范围（scopes）：
     - `read:orders` - 读取订单
     - `write:orders` - 创建/更新订单
     - `read:products` - 读取商品
     - `write:products` - 创建/更新商品
     - `read:reports` - 读取报表
     - `read:companies` - 读取商家信息

4. **获取客户端凭证**
   - Client ID：应用 ID
   - Client Secret：客户端密钥（保存好，只显示一次）

#### 步骤二：在 TTPOS 中注册 API 资源

```go
// main/app/service/logto_api_resource.go
package service

import (
    "github.com/logto-io/go/client"
)

type LogtoAPIResourceService struct {
    logtoClient *logto.Client
}

// RegisterAPIResource 注册 API 资源
func (s *LogtoAPIResourceService) RegisterAPIResource() error {
    // 在 Logto 中注册 TTPOS API 资源
    apiResource := logto.APIResource{
        Name:        "TTPOS API",
        Identifier:  "https://api.ttpos.com",
        Scopes: []logto.Scope{
            {Name: "read:orders", Description: "读取订单"},
            {Name: "write:orders", Description: "创建/更新订单"},
            {Name: "read:products", Description: "读取商品"},
            {Name: "write:products", Description: "创建/更新商品"},
            {Name: "read:reports", Description: "读取报表"},
            {Name: "read:companies", Description: "读取商家信息"},
        },
    }
    
    return s.logtoClient.CreateAPIResource(apiResource)
}
```

#### 步骤三：创建第三方应用管理表

```go
// main/app/model/third_party_app.go
package model

// ThirdPartyApp 第三方应用表
type ThirdPartyApp struct {
    BaseModel
    AppName        string `gorm:"column:app_name;type:varchar(255);comment:应用名称;NOT NULL" json:"app_name"`
    LogtoClientId  string `gorm:"column:logto_client_id;type:varchar(255);uniqueIndex;comment:Logto 客户端ID;NOT NULL" json:"logto_client_id"`
    LogtoClientSecret string `gorm:"column:logto_client_secret;type:varchar(255);comment:Logto 客户端密钥" json:"-"` // 加密存储
    CompanyUuid    uint64 `gorm:"column:company_uuid;type:bigint(20);index;comment:关联的商家ID（可选）" json:"company_uuid"`
    Scopes         string `gorm:"column:scopes;type:text;comment:权限范围（JSON数组）" json:"scopes"`
    Status         int    `gorm:"column:status;type:tinyint(4);default:1;comment:状态：1-启用，0-禁用" json:"status"`
    Description    string `gorm:"column:description;type:text;comment:应用描述" json:"description"`
    CallbackUrl    string `gorm:"column:callback_url;type:varchar(500);comment:回调地址" json:"callback_url"`
}
```

#### 步骤四：实现 Token 验证中间件

```go
// main/middleware/api_auth.go
package middleware

import (
    "fmt"
    "strings"
    "ttpos-server-go/app/api/helper"
    "ttpos-server-go/app/constant"
    "ttpos-server-go/pkg/logto"
)

// APIResourceAuth API 资源保护中间件
func APIResourceAuth(logtoClient *logto.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        
        if authHeader == "" {
            helper.Fail(c, constant.CodeTokenInvalid, "缺少认证信息")
            c.Abort()
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if !(len(parts) == 2 && parts[0] == "Bearer") {
            helper.Fail(c, constant.CodeTokenInvalid, "无效的认证格式")
            c.Abort()
            return
        }
        
        accessToken := parts[1]
        
        // 验证 Access Token
        tokenInfo, err := logtoClient.IntrospectToken(accessToken)
        if err != nil {
            helper.Fail(c, constant.CodeTokenInvalid, "Token 验证失败")
            c.Abort()
            return
        }
        
        // 检查 Token 是否有效
        if !tokenInfo.Active {
            helper.Fail(c, constant.CodeTokenInvalid, "Token 已失效")
            c.Abort()
            return
        }
        
        // 检查 Token 类型（必须是 Access Token）
        if tokenInfo.TokenType != "access_token" {
            helper.Fail(c, constant.CodeTokenInvalid, "无效的 Token 类型")
            c.Abort()
            return
        }
        
        // 获取权限范围（scopes）
        scopes := tokenInfo.Scope
        
        // 验证是否有权限访问当前资源
        requiredScope := getRequiredScope(c.Request.URL.Path, c.Request.Method)
        if requiredScope != "" && !hasScope(scopes, requiredScope) {
            helper.Fail(c, constant.CodeAccessDenied, "权限不足")
            c.Abort()
            return
        }
        
        // 将 Token 信息存储到上下文
        c.Set("client_id", tokenInfo.ClientId)
        c.Set("scopes", scopes)
        c.Set("access_token", accessToken)
        
        c.Next()
    }
}

// getRequiredScope 根据路径和方法获取所需的权限范围
func getRequiredScope(path string, method string) string {
    // 定义路由与权限的映射
    scopeMap := map[string]string{
        "GET /api/v1/orders":     "read:orders",
        "POST /api/v1/orders":    "write:orders",
        "PUT /api/v1/orders":     "write:orders",
        "GET /api/v1/products":   "read:products",
        "POST /api/v1/products":  "write:products",
        "PUT /api/v1/products":   "write:products",
        "GET /api/v1/reports":    "read:reports",
        "GET /api/v1/companies":  "read:companies",
    }
    
    key := method + " " + path
    return scopeMap[key]
}

// hasScope 检查是否包含所需的权限范围
func hasScope(scopes []string, requiredScope string) bool {
    for _, scope := range scopes {
        if scope == requiredScope {
            return true
        }
    }
    return false
}
```

#### 步骤五：第三方服务获取 Token（示例）

```go
// 第三方服务示例代码
package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "strings"
)

type TokenRequest struct {
    GrantType    string `json:"grant_type"`
    ClientId     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    Scope        string `json:"scope"`
}

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope"`
}

// GetAccessToken 获取 Access Token
func GetAccessToken(logtoEndpoint, clientId, clientSecret string, scopes []string) (*TokenResponse, error) {
    tokenReq := TokenRequest{
        GrantType:    "client_credentials",
        ClientId:     clientId,
        ClientSecret: clientSecret,
        Scope:        strings.Join(scopes, " "),
    }
    
    jsonData, _ := json.Marshal(tokenReq)
    
    resp, err := http.Post(
        logtoEndpoint+"/oidc/token",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    
    var tokenResp TokenResponse
    json.Unmarshal(body, &tokenResp)
    
    return &tokenResp, nil
}

// 使用示例
func main() {
    tokenResp, err := GetAccessToken(
        "http://localhost:3001",
        "your_client_id",
        "your_client_secret",
        []string{"read:orders", "read:products"},
    )
    if err != nil {
        panic(err)
    }
    
    // 使用 Access Token 访问 TTPOS API
    req, _ := http.NewRequest("GET", "https://api.ttpos.com/api/v1/orders", nil)
    req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    // ... 处理响应
}
```

#### 步骤六：在 TTPOS 中应用中间件

```go
// main/app/router/v1/api.go
package v1

import (
    "github.com/gin-gonic/gin"
    "ttpos-server-go/main/middleware"
    "ttpos-server-go/pkg/logto"
)

func RegisterAPIHandlers(router gin.IRouter, logtoClient *logto.Client) {
    // API 资源保护路由组
    apiGroup := router.Group("/api/v1")
    
    // 应用 API 资源保护中间件
    apiGroup.Use(middleware.APIResourceAuth(logtoClient))
    
    {
        // 订单相关 API
        apiGroup.GET("/orders", orderController.GetOrders)
        apiGroup.POST("/orders", orderController.CreateOrder)
        apiGroup.PUT("/orders/:id", orderController.UpdateOrder)
        
        // 商品相关 API
        apiGroup.GET("/products", productController.GetProducts)
        apiGroup.POST("/products", productController.CreateProduct)
        
        // 报表相关 API
        apiGroup.GET("/reports", reportController.GetReports)
        
        // 商家信息 API
        apiGroup.GET("/companies", companyController.GetCompanies)
    }
}
```

### 多租户支持

#### 方案一：通过 Organization 限制访问

```go
// 在 Token 验证时，检查第三方应用是否有权限访问特定商家
func validateOrganizationAccess(c *gin.Context, companyUuid uint64) bool {
    clientId := c.GetString("client_id")
    
    // 查询第三方应用配置
    thirdPartyAppRepo := repository.NewThirdPartyAppRepo(db)
    app := thirdPartyAppRepo.GetByClientId(clientId)
    
    // 如果应用关联了特定商家，检查是否匹配
    if app.CompanyUuid > 0 && app.CompanyUuid != companyUuid {
        return false
    }
    
    return true
}
```

#### 方案二：通过 Scope 限制访问范围

```go
// 在 Token 中传递 organization_id
// 第三方服务请求 Token 时，指定组织 ID
tokenReq := TokenRequest{
    GrantType:    "client_credentials",
    ClientId:     clientId,
    ClientSecret: clientSecret,
    Scope:        "read:orders organization:company_uuid_123",
}

// 在 TTPOS API 中验证组织权限
func validateOrganizationScope(c *gin.Context, companyUuid uint64) bool {
    scopes := c.GetStringSlice("scopes")
    
    // 检查是否有组织限制
    for _, scope := range scopes {
        if strings.HasPrefix(scope, "organization:") {
            orgId := strings.TrimPrefix(scope, "organization:")
            if orgId != fmt.Sprintf("%d", companyUuid) {
                return false
            }
        }
    }
    
    return true
}
```

### 权限管理

#### 1. 权限范围定义

```go
// main/app/constant/scopes.go
package constant

const (
    // 订单相关权限
    ScopeReadOrders   = "read:orders"
    ScopeWriteOrders  = "write:orders"
    ScopeDeleteOrders = "delete:orders"
    
    // 商品相关权限
    ScopeReadProducts   = "read:products"
    ScopeWriteProducts  = "write:products"
    ScopeDeleteProducts = "delete:products"
    
    // 报表相关权限
    ScopeReadReports = "read:reports"
    
    // 商家信息权限
    ScopeReadCompanies = "read:companies"
    ScopeWriteCompanies = "write:companies"
)
```

#### 2. 权限验证辅助函数

```go
// main/middleware/scope_check.go
package middleware

func RequireScope(requiredScope string) gin.HandlerFunc {
    return func(c *gin.Context) {
        scopes := c.GetStringSlice("scopes")
        
        if !hasScope(scopes, requiredScope) {
            helper.Fail(c, constant.CodeAccessDenied, "权限不足：需要 "+requiredScope)
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 使用示例
router.GET("/orders", 
    middleware.APIResourceAuth(logtoClient),
    middleware.RequireScope(constant.ScopeReadOrders),
    orderController.GetOrders,
)
```

### 第三方服务接入流程

#### 1. 申请接入

第三方服务提供商需要：
1. 填写接入申请表
2. 说明接入目的和使用场景
3. 列出需要的权限范围

#### 2. 审核和配置

TTPOS 管理员：
1. 审核接入申请
2. 在 Logto 中创建 M2M 应用
3. 配置权限范围
4. 生成客户端凭证
5. 将凭证安全地提供给第三方服务

#### 3. 第三方服务集成

第三方服务：
1. 使用客户端凭证获取 Access Token
2. 携带 Token 访问 TTPOS API
3. 处理 Token 过期（使用 Refresh Token 或重新获取）

### 安全最佳实践

1. **客户端密钥保护**
   - 使用环境变量存储客户端密钥
   - 不要将密钥提交到代码仓库
   - 定期轮换密钥

2. **Token 安全**
   - 使用 HTTPS 传输 Token
   - Token 存储在安全位置（内存或加密存储）
   - 设置合理的 Token 过期时间

3. **权限最小化**
   - 只授予必要的权限范围
   - 定期审查第三方应用的权限
   - 支持权限撤销

4. **审计日志**
   - 记录所有第三方 API 访问
   - 监控异常访问模式
   - 定期审查访问日志

### 示例：完整的第三方服务集成

```go
// 第三方服务完整示例
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

type TTPOSClient struct {
    logtoEndpoint string
    clientId      string
    clientSecret  string
    accessToken   string
    tokenExpiry   time.Time
    httpClient    *http.Client
}

func NewTTPOSClient(logtoEndpoint, clientId, clientSecret string) *TTPOSClient {
    return &TTPOSClient{
        logtoEndpoint: logtoEndpoint,
        clientId:      clientId,
        clientSecret:  clientSecret,
        httpClient:    &http.Client{Timeout: 30 * time.Second},
    }
}

// ensureToken 确保有有效的 Access Token
func (c *TTPOSClient) ensureToken(ctx context.Context) error {
    // 如果 Token 还有效，直接返回
    if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
        return nil
    }
    
    // 获取新的 Token
    tokenResp, err := GetAccessToken(
        c.logtoEndpoint,
        c.clientId,
        c.clientSecret,
        []string{"read:orders", "read:products"},
    )
    if err != nil {
        return fmt.Errorf("获取 Token 失败: %w", err)
    }
    
    c.accessToken = tokenResp.AccessToken
    c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
    
    return nil
}

// GetOrders 获取订单列表
func (c *TTPOSClient) GetOrders(ctx context.Context, companyUuid uint64) ([]Order, error) {
    if err := c.ensureToken(ctx); err != nil {
        return nil, err
    }
    
    req, _ := http.NewRequestWithContext(
        ctx,
        "GET",
        fmt.Sprintf("https://api.ttpos.com/api/v1/orders?company_uuid=%d", companyUuid),
        nil,
    )
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API 请求失败: %d", resp.StatusCode)
    }
    
    // 解析响应
    var orders []Order
    // ... 解析逻辑
    
    return orders, nil
}
```

---

## 注意事项

### 1. 数据同步

**问题**：Logto 中的用户和组织数据需要与 TTPOS 业务数据同步

**解决方案**：
- 创建员工时，同步创建 Logto 用户和组织
- 更新员工信息时，同步更新 Logto 用户信息
- 删除员工时，同步删除 Logto 用户关联

```go
// main/app/service/staff.go

func (s *staffSrv) CreateStaff(ctx context.Context, req req.AddStaffReq) error {
    // 1. 创建 TTPOS 员工
    staff := model.Staff{
        // ... 字段赋值
    }
    err := staffRepo.Create(&staff)
    if err != nil {
        return err
    }
    
    // 2. 创建 Logto 用户
    logtoUser, err := s.logtoClient.CreateUser(staff.Username, staff.Phone, staff.Email)
    if err != nil {
        // 回滚：删除已创建的员工
        staffRepo.Delete(staff.Uuid)
        return err
    }
    
    // 3. 创建关联记录
    logtoUserRepo.Create(&model.LogtoUser{
        LogtoUserId: logtoUser.Id,
        StaffUuid:   staff.Uuid,
        CompanyUuid: staff.CompanyUuid,
        LogtoOrgId:  logtoOrgId, // 从商家配置获取
    })
    
    // 4. 将用户添加到组织
    err = s.logtoOrgSrv.AddUserToOrganization(logtoUser.Id, logtoOrgId, "member")
    if err != nil {
        // 处理错误
    }
    
    return nil
}
```

### 2. 密码管理

**问题**：用户密码存储在 Logto，但业务逻辑可能需要验证密码

**解决方案**：
- 方案 A：所有密码验证都通过 Logto API
- 方案 B：首次登录时，将 Logto 密码同步到 TTPOS（不推荐，安全性低）
- 方案 C：使用 Logto 的密码重置功能，不维护本地密码

### 3. Token 兼容性

**问题**：现有代码依赖自定义 Token 结构

**解决方案**：
- 保持现有 Token 结构不变
- Logto 登录后，生成 TTPOS Token
- 中间件同时支持两种 Token（向后兼容）

### 4. 性能考虑

**问题**：每次请求都需要验证 Logto Token 可能影响性能

**解决方案**：
- 使用缓存存储已验证的 Token
- 只在必要时调用 Logto API 验证
- 考虑使用本地 JWT 验证（如果 Logto 使用 JWT）

---

## 总结

### 集成优势

1. ✅ **标准化**：遵循 OIDC 标准，易于集成和维护
2. ✅ **多租户**：内置组织管理，支持多商家场景
3. ✅ **功能丰富**：支持社交登录、MFA 等高级功能
4. ✅ **安全性**：专业的身份认证服务，安全性高

### 实施建议

1. **分阶段实施**：
   - 第一阶段：部署 Logto，实现基础登录
   - 第二阶段：实现多租户支持
   - 第三阶段：迁移现有用户

2. **保持兼容**：
   - 同时支持 Logto 和现有认证方式
   - 逐步迁移用户

3. **数据同步**：
   - 建立用户和组织同步机制
   - 处理同步失败的情况

---

**最后更新**: 2025-11-20  
**维护者**: TTPOS Team  
**版本**: v1.0

