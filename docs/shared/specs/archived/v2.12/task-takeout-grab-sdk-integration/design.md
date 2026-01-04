# Takeout 模块集成 GrabFood 官方 SDK 技术设计

> 本文档定义 Grab 官方 SDK 集成的技术实现方案。

## 📋 基本信息

| 项目           | 内容                                        |
| -------------- | ------------------------------------------- |
| **关联需求**   | [requirements.md](./requirements.md)        |
| **设计日期**   | 2025-12-08                                  |
| **设计者**     | rikugun                                     |
| **技术栈**     | Go 1.23 + GoFrame 2.x + Grab SDK v1.0.2     |

---

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        ttpos-takeout                            │
├─────────────────────────────────────────────────────────────────┤
│  Controller Layer (api/grab/v1/)                                │
│  ├── submit_order.go          (Webhook: 订单提交)               │
│  ├── push_order_state.go      (Webhook: 状态变更)               │
│  ├── get_menu.go              (Webhook: 菜单拉取)               │
│  └── ...                                                        │
├─────────────────────────────────────────────────────────────────┤
│  Service Layer (internal/service/)                              │
│  └── grab.go                  (业务编排)                        │
├─────────────────────────────────────────────────────────────────┤
│  Logic Layer (internal/logic/grab/)                             │
│  ├── sdk_wrapper.go     [NEW] (SDK Wrapper - 统一入口)          │
│  ├── auth.go                  (签名验证 - 保留)                 │
│  ├── order_service.go         (订单业务)                        │
│  ├── store_service.go         (门店业务)                        │
│  ├── menu_service.go          (菜单业务)                        │
│  └── client.go          [DEP] (旧实现 - 废弃)                   │
├─────────────────────────────────────────────────────────────────┤
│  Model Layer                                                    │
│  ├── dto/grab/          [DEP] (自定义 DTO - 逐步废弃)           │
│  └── SDK Models         [NEW] (grabfood.SubmitOrderRequest 等)  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│              github.com/grab/grabfood-api-sdk-go                │
│  ├── APIClient          (HTTP 客户端)                           │
│  ├── GetOauthGrabAPI    (OAuth Token)                           │
│  ├── AcceptRejectOrderAPI, CancelOrderAPI, ...  (订单操作)      │
│  ├── PauseStoreAPI, GetStoreStatusAPI, ...      (门店管理)      │
│  └── Model Types        (SubmitOrderRequest, OrderStateRequest) │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 SDK Wrapper 层设计

**核心原则**：业务代码不直接依赖 SDK，通过 Wrapper 层隔离。

```go
// internal/logic/grab/sdk_wrapper.go

package grab

import (
    "context"
    "sync"
    
    grabfood "github.com/grab/grabfood-api-sdk-go"
    "github.com/gogf/gf/v2/frame/g"
)

// SDKWrapper 封装 Grab 官方 SDK
type SDKWrapper struct {
    client    *grabfood.APIClient
    config    *SDKConfig
    tokenLock sync.RWMutex
}

// SDKConfig SDK 配置
type SDKConfig struct {
    ClientID     string
    ClientSecret string
    Environment  string // "staging" | "production"
}

// NewSDKWrapper 创建 SDK Wrapper
func NewSDKWrapper(cfg *SDKConfig) *SDKWrapper {
    config := grabfood.NewConfiguration()
    return &SDKWrapper{
        client: grabfood.NewAPIClient(config),
        config: cfg,
    }
}

// GetContext 获取带环境配置的 Context
func (w *SDKWrapper) GetContext(ctx context.Context) context.Context {
    if w.config.Environment == "staging" {
        return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.StgEnv)
    }
    return context.WithValue(ctx, grabfood.ContextServerIndex, grabfood.PrdEnv)
}

// GetAccessToken 获取 Access Token (带 Redis 缓存)
func (w *SDKWrapper) GetAccessToken(ctx context.Context) (string, error) {
    // 1. 尝试从 Redis 读取
    redisKey := RedisKeyTokenPrefix + w.config.ClientID
    cachedToken, err := g.Redis().Get(ctx, redisKey)
    if err == nil && !cachedToken.IsEmpty() {
        return cachedToken.String(), nil
    }
    
    // 2. 通过 SDK 获取新 Token
    authReq := grabfood.NewGrabOauthRequest(
        w.config.ClientID,
        w.config.ClientSecret,
        "client_credentials",
        "food.partner_api",
    )
    
    resp, _, err := w.client.GetOauthGrabAPI.
        GetOauthGrab(w.GetContext(ctx)).
        GrabOauthRequest(*authReq).
        Execute()
    if err != nil {
        return "", fmt.Errorf("SDK OAuth failed: %w", err)
    }
    
    token := *resp.AccessToken
    expiresIn := int(*resp.ExpiresIn)
    
    // 3. 写入 Redis 缓存
    ttl := expiresIn - TokenExpireBuffer
    if ttl > 0 {
        _ = g.Redis().SetEX(ctx, redisKey, token, int64(ttl))
    }
    
    return token, nil
}

// GetAuthorizationHeader 获取 Authorization 请求头
func (w *SDKWrapper) GetAuthorizationHeader(ctx context.Context) (string, error) {
    token, err := w.GetAccessToken(ctx)
    if err != nil {
        return "", err
    }
    return "Bearer " + token, nil
}
```

---

## 2. Outbound API 迁移设计

### 2.1 API 映射表

| 功能 | 旧实现 (client.go) | SDK API | 方法签名 |
|------|-------------------|---------|----------|
| OAuth Token | `fetchTokenFromGrab()` | `GetOauthGrabAPI.GetOauthGrab()` | `Execute()` |
| 接受订单 | `AcceptOrder()` | `AcceptRejectOrderAPI.AcceptRejectOrder()` | `Execute()` |
| 拒绝订单 | `RejectOrder()` | `AcceptRejectOrderAPI.AcceptRejectOrder()` | `Execute()` |
| 取消订单 | `CancelOrder()` | `CancelOrderAPI.CancelOrder()` | `Execute()` |
| 标记就绪 | `MarkOrderReady()` | `MarkOrderReadyAPI.MarkOrderReady()` | `Execute()` |
| 更新配送状态 | `UpdateDeliveryState()` | `UpdateDeliveryStateAPI.UpdateDeliveryState()` | `Execute()` |
| 暂停门店 | `UpdateStoreStatus()` | `PauseStoreAPI.PauseStore()` | `Execute()` |
| 获取门店状态 | `GetStoreStatus()` | `GetStoreStatusAPI.GetStoreStatus()` | `Execute()` |
| 获取营业时间 | `GetStoreHours()` | `GetStoreHourAPI.GetStoreHour()` | `Execute()` |
| 菜单通知 | `NotifyMenuUpdate()` | `UpdateMenuNotificationAPI.UpdateMenuNotification()` | `Execute()` |

### 2.2 迁移示例：订单接受

**旧实现** (`client.go`):

```go
func (c *APIClient) AcceptOrder(ctx context.Context, orderID string) error {
    path := fmt.Sprintf("/partner/v1/orders/%s", orderID)
    req := grab.AcceptOrderRequest{
        OrderID: orderID,
        State:   "ACCEPTED",
    }
    _, err := c.doRequest(ctx, "POST", path, req)
    return err
}
```

**新实现** (`sdk_wrapper.go`):

```go
func (w *SDKWrapper) AcceptOrder(ctx context.Context, orderID string) error {
    auth, err := w.GetAuthorizationHeader(ctx)
    if err != nil {
        return err
    }
    
    req := grabfood.NewAcceptOrderRequest(orderID, grabfood.ACCEPTED)
    
    _, _, err = w.client.AcceptRejectOrderAPI.
        AcceptRejectOrder(w.GetContext(ctx), orderID).
        Authorization(auth).
        AcceptOrderRequest(*req).
        Execute()
    
    return err
}
```

---

## 3. Inbound DTO 替换设计

### 3.1 DTO 映射表

| Webhook 端点 | 现有 DTO (dto/grab/) | SDK Model | 替换方案 |
|-------------|---------------------|-----------|----------|
| `POST /partner/orders` | `SubmitOrderRequest` | `grabfood.SubmitOrderRequest` | 直接使用 SDK Model |
| `PUT /partner/orders/state` | `OrderStateRequest` | `grabfood.OrderStateRequest` | 直接使用 SDK Model |
| `GET /partner/menu` | `GetMenuResponse` | `grabfood.GetMenuNewResponse` | 使用 SDK 构造响应 |
| `POST /partner/menu/sync/state` | `MenuSyncStateRequest` | `grabfood.MenuSyncWebhookRequest` | 直接使用 SDK Model |
| `POST /oauth/partner` | `PartnerTokenRequest` | `grabfood.PartnerOauthRequest` | 直接使用 SDK Model |

### 3.2 Controller 层适配

**当前实现** (`api/grab/v1/submit_order.go`):

```go
type SubmitOrderReq struct {
    g.Meta `path:"/partner/orders" method:"post"`
    *grab.SubmitOrderRequest  // 自定义 DTO
}
```

**迁移后**:

```go
import grabfood "github.com/grab/grabfood-api-sdk-go"

type SubmitOrderReq struct {
    g.Meta `path:"/partner/orders" method:"post"`
    *grabfood.SubmitOrderRequest  // SDK Model
}
```

### 3.3 签名验证保留

SDK **不提供** Webhook 签名验证功能，需保留 `auth.go` 中的实现：

```go
// internal/logic/grab/auth.go - 保持不变

// SignatureVerifier 签名验证器
type SignatureVerifier struct {
    secretKey string
}

// VerifySignature 验证 Grab Webhook 签名
// signature: X-Grab-Signature 请求头值
// timestamp: X-Grab-Timestamp 请求头值
// body: 请求体原始字节
func (v *SignatureVerifier) VerifySignature(signature, timestamp string, body []byte) error {
    // ... HMAC-SHA256 验证逻辑保持不变
}
```

---

## 4. 配置管理

### 4.1 配置结构

```yaml
# config.tpl.yaml
grab:
  clientId: ${GRAB_CLIENT_ID}
  clientSecret: ${GRAB_CLIENT_SECRET}
  environment: ${GRAB_ENV:staging}  # staging | production
  webhookSecret: ${GRAB_WEBHOOK_SECRET}
```

### 4.2 配置加载

```go
// internal/model/conf/provider_grab.go

type GrabConfig struct {
    ClientID      string `yaml:"clientId"`
    ClientSecret  string `yaml:"clientSecret"`
    Environment   string `yaml:"environment"`
    WebhookSecret string `yaml:"webhookSecret"`
}

func NewGrabSDKWrapper() *grab.SDKWrapper {
    cfg := &grab.SDKConfig{
        ClientID:     g.Cfg().MustGet(ctx, "grab.clientId").String(),
        ClientSecret: g.Cfg().MustGet(ctx, "grab.clientSecret").String(),
        Environment:  g.Cfg().MustGet(ctx, "grab.environment").String(),
    }
    return grab.NewSDKWrapper(cfg)
}
```

---

## 5. 迁移策略

### 5.1 渐进式迁移

采用 **Feature Flag** 控制迁移进度，新旧实现可并存：

```go
// internal/logic/grab/grab.go

func (s *sGrab) AcceptOrder(ctx context.Context, orderID string) error {
    if g.Cfg().MustGet(ctx, "grab.useSDK").Bool() {
        // 新实现: SDK
        return s.sdkWrapper.AcceptOrder(ctx, orderID)
    }
    // 旧实现: client.go
    return s.client.AcceptOrder(ctx, orderID)
}
```

### 5.2 迁移顺序

```
Phase 1: SDK 引入 + Wrapper 层
    ↓
Phase 2: Outbound 迁移 (OAuth → 订单 → 门店 → 菜单)
    ↓  每个 API 迁移后单独测试
Phase 3: Inbound DTO 替换
    ↓  保持 API 定义不变，仅替换内部 Model
Phase 4: 清理旧代码 + 新功能评估
```

---

## 6. 文件变更清单

### 6.1 新增文件

| 文件路径 | 说明 |
|----------|------|
| `internal/logic/grab/sdk_wrapper.go` | SDK Wrapper 封装层 |

### 6.2 修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `go.mod` | 添加 SDK 依赖 |
| `internal/logic/grab/grab.go` | 集成 SDKWrapper |
| `internal/logic/grab/order_service.go` | 使用 SDK 调用 |
| `internal/logic/grab/store_service.go` | 使用 SDK 调用 |
| `internal/logic/grab/menu_service.go` | 使用 SDK 调用 |
| `api/grab/v1/*.go` | 使用 SDK Model |
| `internal/model/conf/provider.go` | 添加 SDK 配置 |

### 6.3 废弃文件

| 文件路径 | 处理方式 |
|----------|----------|
| `internal/logic/grab/client.go` | 标记废弃，Phase 4 删除 |
| `internal/model/dto/grab/order.go` | 保留部分常量，Model 废弃 |
| `internal/model/dto/grab/menu.go` | 保留部分常量，Model 废弃 |
| `internal/model/dto/grab/store.go` | 保留部分常量，Model 废弃 |

### 6.4 保留文件

| 文件路径 | 原因 |
|----------|------|
| `internal/logic/grab/auth.go` | SDK 不提供签名验证 |
| `internal/model/dto/grab/auth.go` | 签名相关常量 |

---

## 7. 测试策略

### 7.1 单元测试

```go
// internal/logic/grab/sdk_wrapper_test.go

func TestSDKWrapper_GetAccessToken(t *testing.T) {
    // Mock Redis
    // Mock SDK API
    // 验证 Token 缓存逻辑
}

func TestSDKWrapper_AcceptOrder(t *testing.T) {
    // Mock SDK API
    // 验证请求参数正确性
}
```

### 7.2 集成测试

- 使用 Grab Staging 环境进行端到端测试
- 验证所有 Outbound API 功能正常
- 验证所有 Inbound Webhook 处理正常

### 7.3 回归测试

- 现有功能无退化
- 订单流程完整性验证
- 门店状态管理验证

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| SDK Model 字段与业务代码不兼容 | 中 | 创建 Adapter 层转换，逐步适配 |
| SDK 内部 Bug | 低 | 可回退到旧实现 (Feature Flag) |
| Token 管理行为差异 | 中 | 保持 Redis 缓存策略不变 |
| 性能差异 | 低 | Benchmark 对比测试 |

---

## 9. 参考资料

- [Grab SDK GitHub](https://github.com/grab/grabfood-api-sdk-go)
- [Grab SDK Go Doc](https://pkg.go.dev/github.com/grab/grabfood-api-sdk-go)
- [GrabFood API Documentation](https://developer.grab.com)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: rikugun
