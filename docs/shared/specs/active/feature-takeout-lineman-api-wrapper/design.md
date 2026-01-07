# Lineman API 包装实现 设计文档

> 本文档定义 Lineman API 包装服务的技术设计和实现方案。

## 📋 概述

实现 Lineman 外卖平台 API 的包装服务，为 ttpos-takeout 模块提供统一的第三方平台 API 调用接口。仅在 `internal/logic/lineman` 目录下实现业务逻辑层（Service），不对外提供 gRPC 服务。

参考现有的 GrabFood 和 Skootar API 包装实现模式，遵循 GoFrame 框架规范和项目编码标准。

---

## 🎯 规范对齐

### Go BMP 规范 (go-rules.mdc)

- 遵循 GoFrame v2.x 框架规范
- 使用 `gerror` 进行错误处理
- 使用 `g.Log()` 记录日志
- 不对外提供 gRPC 服务，仅内部 Service 层实现
- 日志描述使用中文
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### ttpos-takeout 子模块规范 (go-ttpos-takeout.mdc)

- Logic/Service 层应尽量复用已有的逻辑，避免重复实现
- 返回参数类型不能是 `takeout.ApiResponse`
- 返回具体的业务数据类型（如 DTO）
- 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`

### 参考实现

- **必须**参考 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/` 的实现模式
- **必须**参考 `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/` 的实现模式
- 保持代码风格和目录结构的一致性

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Grab 配置管理模式**: `internal/logic/grab/config.go` - 参考配置读取和懒加载单例模式
- **Grab 认证管理模式**: `internal/logic/grab/grab.go` - 参考 Token 管理、Redis 缓存、双重检查锁
- **Grab 签名验证**: `internal/logic/grab/auth.go` - 参考 Webhook 签名验证实现
- **Grab HTTP 请求封装**: `internal/logic/grab/grab.go` - 参考统一错误处理模式
- **Grab DTO 定义**: `internal/model/dto/grab/` - 参考 DTO 结构和命名规范

### 集成点

- **配置管理**: 在 `manifest/config/config.yaml` 中添加 Lineman 配置节点
- **配置结构**: 在 `internal/model/conf/lineman.go` 中定义 Lineman 配置结构
- **DTO 定义**: 在 `internal/model/dto/lineman/` 中定义订单/菜单/门店相关 DTO
- **服务注册**: 在 `internal/logic/lineman/lineman.go` 中实现服务并注册到服务容器

---

## 🏗️ 架构设计

### 分层设计原则

**ttpos-bmp 内部 Service 层架构**:

```
Controller 层 (其他模块调用)
  ↓ 依赖
Logic 层 (Lineman Service)
  ↓ 依赖
HTTP 请求封装 (Lineman API)
  ↓ 依赖
外部 API (Lineman 平台)
```

**依赖规则**:

- ✅ 其他模块通过 `service.Lineman()` 调用
- ✅ Lineman Service 封装 HTTP 请求
- ❌ 不对外提供 gRPC 接口
- ✅ 使用 `gerror` 统一错误处理
- ✅ 使用 `g.Log()` 记录日志

### 模块划分

#### ttpos-takeout/internal/logic/lineman 目录结构

```
ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/
├── lineman.go          # 主服务入口，配置管理，服务注册
├── config.go           # 配置结构定义，配置读取
├── auth.go             # 认证相关（Token 管理、请求签名）
├── order.go            # 订单管理 API 实现
├── menu.go             # 菜单管理 API 实现
├── store.go            # 门店管理 API 实现
├── http_client.go      # HTTP 请求工具封装（可选）
└── lineman_test.go     # 单元测试
```

#### DTO 目录结构

```
ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/
├── order.go            # 订单相关 DTO
├── menu.go             # 菜单相关 DTO
├── store.go            # 门店相关 DTO
└── auth.go             # 认证相关 DTO（如需要）
```

#### 配置结构定义

```
ttpos-bmp/app/ttpos-takeout/internal/model/conf/lineman.go
```

---

## 📊 数据模型

### 配置结构

```go
// internal/model/conf/lineman.go
package conf

// Lineman Lineman 平台配置
type Lineman struct {
    Endpoint     string `json:"endpoint"`     // API 端点 (staging/production)
    ApiKey       string `json:"apiKey"`       // API Key
    SecretKey    string `json:"secretKey"`    // Secret Key（用于签名验证）
    Environment  string `json:"environment"`  // staging/production
    Timeout      int    `json:"timeout"`      // 请求超时时间（秒），默认 30
}
```

### DTO 定义

#### 订单相关 DTO

```go
// internal/model/dto/lineman/order.go
package lineman

// ReceiveOrderReq 接收订单 Webhook 请求
type ReceiveOrderReq struct {
    OrderID       string  `json:"orderId"`       // 订单ID
    MerchantID    string  `json:"merchantId"`    // 商户ID
    OrderTime     int64   `json:"orderTime"`     // 下单时间
    TotalAmount   float64 `json:"totalAmount"`   // 订单总金额
    Items         []OrderItem `json:"items"`     // 订单商品列表
    DeliveryInfo  DeliveryInfo `json:"deliveryInfo"` // 配送信息
}

type OrderItem struct {
    ItemID   string  `json:"itemId"`   // 商品ID
    Name     string  `json:"name"`     // 商品名称
    Quantity int     `json:"quantity"` // 数量
    Price    float64 `json:"price"`    // 单价
}

type DeliveryInfo struct {
    Address    string `json:"address"`    // 配送地址
    Phone      string `json:"phone"`      // 联系电话
    CustomerName string `json:"customerName"` // 顾客姓名
}

// AcceptOrderReq 接受订单请求
type AcceptOrderReq struct {
    OrderID string `json:"orderId"` // 订单ID
}

// RejectOrderReq 拒绝订单请求
type RejectOrderReq struct {
    OrderID      string `json:"orderId"`      // 订单ID
    RejectReason string `json:"rejectReason"` // 拒绝原因
}

// CancelOrderReq 取消订单请求
type CancelOrderReq struct {
    OrderID      string `json:"orderId"`      // 订单ID
    CancelReason string `json:"cancelReason"` // 取消原因
}

// UpdateOrderStatusReq 更新订单状态请求
type UpdateOrderStatusReq struct {
    OrderID string `json:"orderId"` // 订单ID
    Status  string `json:"status"`  // 订单状态
}

// MarkOrderReadyReq 标记订单准备完成请求
type MarkOrderReadyReq struct {
    OrderID string `json:"orderId"` // 订单ID
}
```

#### 菜单相关 DTO

```go
// internal/model/dto/lineman/menu.go
package lineman

// GetMenuReq 获取菜单请求
type GetMenuReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
}

// GetMenuResp 获取菜单响应
type GetMenuResp struct {
    MenuItems []MenuItem `json:"menuItems"` // 菜单项列表
}

type MenuItem struct {
    ItemID      string  `json:"itemId"`      // 商品ID
    Name        string  `json:"name"`        // 商品名称
    Description string  `json:"description"` // 商品描述
    Price       float64 `json:"price"`       // 价格
    Available   bool    `json:"available"`   // 是否可用
}

// SyncMenuReq 同步菜单请求
type SyncMenuReq struct {
    MerchantID string     `json:"merchantId"` // 商户ID
    MenuItems  []MenuItem `json:"menuItems"`  // 菜单项列表
}

// UpdateMenuStatusReq 更新菜单状态请求
type UpdateMenuStatusReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
    ItemID     string `json:"itemId"`     // 商品ID
    Available  bool   `json:"available"`  // 是否可用
}
```

#### 门店相关 DTO

```go
// internal/model/dto/lineman/store.go
package lineman

// GetStoreStatusReq 获取门店状态请求
type GetStoreStatusReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
}

// GetStoreStatusResp 获取门店状态响应
type GetStoreStatusResp struct {
    MerchantID string `json:"merchantId"` // 商户ID
    Status     string `json:"status"`     // 门店状态 (open/closed/paused)
}

// PauseStoreReq 暂停门店营业请求
type PauseStoreReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
    Duration   int    `json:"duration"`   // 暂停时长（分钟）
}

// ResumeStoreReq 恢复门店营业请求
type ResumeStoreReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
}

// UpdateStoreInfoReq 更新门店信息请求
type UpdateStoreInfoReq struct {
    MerchantID string `json:"merchantId"` // 商户ID
    StoreName  string `json:"storeName"`  // 门店名称
    Address    string `json:"address"`    // 门店地址
    Phone      string `json:"phone"`      // 联系电话
}
```

---

## 🔌 API 设计

### 内部 Service 接口

```go
// internal/service/lineman.go (框架自动生成)
type ILineman interface {
    // 配置管理
    MustConf() *conf.Lineman
    
    // 认证管理
    GetAccessToken(ctx context.Context) (string, error)
    VerifyWebhookSignature(ctx context.Context, signature, timestamp string, body []byte) error
    
    // 订单管理
    HandleReceiveOrder(ctx context.Context, req *dto.ReceiveOrderReq) error
    AcceptOrder(ctx context.Context, orderID string) error
    RejectOrder(ctx context.Context, orderID string, reason string) error
    CancelOrder(ctx context.Context, orderID string, reason string) error
    UpdateOrderStatus(ctx context.Context, orderID string, status string) error
    MarkOrderReady(ctx context.Context, orderID string) error
    
    // 菜单管理
    GetMenu(ctx context.Context, merchantID string) (*dto.GetMenuResp, error)
    SyncMenu(ctx context.Context, merchantID string, menuItems []dto.MenuItem) error
    UpdateMenuStatus(ctx context.Context, merchantID string, itemID string, available bool) error
    
    // 门店管理
    GetStoreStatus(ctx context.Context, merchantID string) (*dto.GetStoreStatusResp, error)
    PauseStore(ctx context.Context, merchantID string, duration int) error
    ResumeStore(ctx context.Context, merchantID string) error
    UpdateStoreInfo(ctx context.Context, merchantID string, storeName, address, phone string) error
}
```

### HTTP API 调用设计

#### 认证方式

根据 Lineman API Spec 文档，可能采用以下认证方式之一：

1. **API Key 认证**: 在请求头中携带 `X-API-Key: {apiKey}`
2. **OAuth 2.0**: 使用 `Authorization: Bearer {token}`
3. **签名认证**: 使用 `X-Signature: {signature}`

**实现策略**（参考 Grab）：

- 如果需要 Token，使用 Redis 缓存，提前刷新避免过期
- 如果需要签名，实现签名算法和验证逻辑
- 封装 `getAuthorizationHeader()` 方法，统一添加认证信息

#### HTTP 请求工具

```go
// internal/logic/lineman/http_client.go (可选)
package lineman

import (
    "context"
    "net/http"
    "time"
    "github.com/gogf/gf/v2/frame/g"
)

// doRequest 封装 HTTP 请求
func (s *sLineman) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
    client := &http.Client{
        Timeout: time.Duration(s.MustConf().Timeout) * time.Second,
    }
    
    // 构造请求
    // 添加认证信息
    // 发送请求
    // 处理响应
    // 统一错误处理
    
    return nil, nil
}
```

---

## 🧩 组件和接口

### 主服务实现

```go
// internal/logic/lineman/lineman.go
package lineman

import (
    "context"
    "sync"
    
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    
    "ttpos-bmp/app/ttpos-takeout/internal/model/conf"
    "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
    "ttpos-bmp/app/ttpos-takeout/internal/service"
}

var (
    // Lineman Lineman 服务实例
    Lineman  = new(sLineman)
    config   *conf.Lineman
    configMu sync.RWMutex
)

type sLineman struct {
    // 可以添加需要的字段，如 HTTP 客户端、认证信息等
}

func init() {
    service.RegisterLineman(Lineman)
}

// MustConf 获取 Lineman 配置（懒加载单例）
func (s *sLineman) MustConf() *conf.Lineman {
    configMu.RLock()
    if config != nil {
        defer configMu.RUnlock()
        return config
    }
    configMu.RUnlock()
    
    configMu.Lock()
    defer configMu.Unlock()
    
    // 双重检查
    if config != nil {
        return config
    }
    
    config = MustConfig()
    return config
}

// handleError 统一错误处理
func (s *sLineman) handleError(ctx context.Context, err error, operation string) error {
    if err == nil {
        return nil
    }
    g.Log().Errorf(ctx, "[Lineman] %s 失败: %v", operation, err)
    return gerror.Wrapf(err, "[Lineman] %s 失败", operation)
}
```

### 配置管理实现

```go
// internal/logic/lineman/config.go
package lineman

import (
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gctx"
    
    "ttpos-bmp/app/ttpos-takeout/internal/model/conf"
)

// MustConfig 获取 Lineman 配置（平台维度）
// 读取 app.provider.lineman.platform 节点
func MustConfig() *conf.Lineman {
    ctx := gctx.New()
    
    var lineman conf.Lineman
    if err := g.Cfg().MustGet(ctx, "app.provider.lineman.platform").Scan(&lineman); err != nil {
        g.Log().Fatal(ctx, err)
        panic(gerror.Newf("获取 Lineman 平台配置失败: %v", err))
    }
    
    return &lineman
}
```

### 认证管理实现

```go
// internal/logic/lineman/auth.go
package lineman

import (
    "context"
    
    "github.com/gogf/gf/v2/errors/gerror"
    "github.com/gogf/gf/v2/frame/g"
)

// GetAccessToken 获取或刷新 Access Token（如 Lineman 使用 OAuth 2.0）
// 参考 Grab 的实现，使用 Redis 缓存
func (s *sLineman) GetAccessToken(ctx context.Context) (string, error) {
    conf := s.MustConf()
    // 根据 Lineman API Spec 实现 Token 获取逻辑
    // 如果使用 API Key，直接返回 conf.ApiKey
    // 如果使用 OAuth，实现 Token 获取和缓存
    
    g.Log().Infof(ctx, "[Lineman] 获取授权信息")
    return conf.ApiKey, nil
}

// VerifyWebhookSignature 验证 Lineman Webhook 签名（公开方法，供其他服务调用）
// signature: X-Lineman-Signature 请求头值
// timestamp: X-Lineman-Timestamp 请求头值
// body: 请求体原始字节
func (s *sLineman) VerifyWebhookSignature(ctx context.Context, signature, timestamp string, body []byte) error {
    conf := s.MustConf()
    // 根据 Lineman API Spec 实现签名验证逻辑
    // 参考 Grab 的签名验证实现
    
    g.Log().Infof(ctx, "[Lineman] 验证 Webhook 签名")
    // TODO: 实现签名验证
    return nil
}
```

### 订单管理实现

```go
// internal/logic/lineman/order.go
package lineman

import (
    "context"
    
    "github.com/gogf/gf/v2/frame/g"
    
    dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// HandleReceiveOrder 处理 Lineman 订单接收 Webhook
// 签名验证已由中间件完成
func (s *sLineman) HandleReceiveOrder(ctx context.Context, req *dto.ReceiveOrderReq) error {
    g.Log().Infof(ctx, "[Lineman] 接收订单: %s", req.OrderID)
    // TODO: 实现订单接收逻辑
    return nil
}

// AcceptOrder 接受订单
func (s *sLineman) AcceptOrder(ctx context.Context, orderID string) error {
    g.Log().Infof(ctx, "[Lineman] 接受订单: %s", orderID)
    // TODO: 调用 Lineman API 接受订单
    return nil
}

// RejectOrder 拒绝订单
func (s *sLineman) RejectOrder(ctx context.Context, orderID string, reason string) error {
    g.Log().Infof(ctx, "[Lineman] 拒绝订单: %s, 原因: %s", orderID, reason)
    // TODO: 调用 Lineman API 拒绝订单
    return nil
}

// CancelOrder 取消订单
func (s *sLineman) CancelOrder(ctx context.Context, orderID string, reason string) error {
    g.Log().Infof(ctx, "[Lineman] 取消订单: %s, 原因: %s", orderID, reason)
    // TODO: 调用 Lineman API 取消订单
    return nil
}

// UpdateOrderStatus 更新订单状态
func (s *sLineman) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
    g.Log().Infof(ctx, "[Lineman] 更新订单状态: %s, 状态: %s", orderID, status)
    // TODO: 调用 Lineman API 更新订单状态
    return nil
}

// MarkOrderReady 标记订单准备完成
func (s *sLineman) MarkOrderReady(ctx context.Context, orderID string) error {
    g.Log().Infof(ctx, "[Lineman] 订单已准备完成: %s", orderID)
    // TODO: 调用 Lineman API 标记订单准备完成
    return nil
}
```

### 菜单管理实现

```go
// internal/logic/lineman/menu.go
package lineman

import (
    "context"
    
    "github.com/gogf/gf/v2/frame/g"
    
    dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// GetMenu 获取菜单
func (s *sLineman) GetMenu(ctx context.Context, merchantID string) (*dto.GetMenuResp, error) {
    g.Log().Infof(ctx, "[Lineman] 获取菜单: %s", merchantID)
    // TODO: 调用 Lineman API 获取菜单
    return &dto.GetMenuResp{}, nil
}

// SyncMenu 同步菜单到 Lineman
func (s *sLineman) SyncMenu(ctx context.Context, merchantID string, menuItems []dto.MenuItem) error {
    g.Log().Infof(ctx, "[Lineman] 同步菜单: %s, 菜单项数量: %d", merchantID, len(menuItems))
    // TODO: 调用 Lineman API 同步菜单
    return nil
}

// UpdateMenuStatus 更新菜单状态（启用/禁用）
func (s *sLineman) UpdateMenuStatus(ctx context.Context, merchantID string, itemID string, available bool) error {
    g.Log().Infof(ctx, "[Lineman] 更新菜单状态: merchant=%s, item=%s, available=%v", merchantID, itemID, available)
    // TODO: 调用 Lineman API 更新菜单状态
    return nil
}
```

### 门店管理实现

```go
// internal/logic/lineman/store.go
package lineman

import (
    "context"
    
    "github.com/gogf/gf/v2/frame/g"
    
    dto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman"
)

// GetStoreStatus 获取门店状态
func (s *sLineman) GetStoreStatus(ctx context.Context, merchantID string) (*dto.GetStoreStatusResp, error) {
    g.Log().Infof(ctx, "[Lineman] 获取门店状态: %s", merchantID)
    // TODO: 调用 Lineman API 获取门店状态
    return &dto.GetStoreStatusResp{}, nil
}

// PauseStore 暂停门店营业
func (s *sLineman) PauseStore(ctx context.Context, merchantID string, duration int) error {
    g.Log().Infof(ctx, "[Lineman] 暂停门店营业: merchant=%s, duration=%d 分钟", merchantID, duration)
    // TODO: 调用 Lineman API 暂停门店
    return nil
}

// ResumeStore 恢复门店营业
func (s *sLineman) ResumeStore(ctx context.Context, merchantID string) error {
    g.Log().Infof(ctx, "[Lineman] 恢复门店营业: %s", merchantID)
    // TODO: 调用 Lineman API 恢复门店
    return nil
}

// UpdateStoreInfo 更新门店信息
func (s *sLineman) UpdateStoreInfo(ctx context.Context, merchantID string, storeName, address, phone string) error {
    g.Log().Infof(ctx, "[Lineman] 更新门店信息: %s", merchantID)
    // TODO: 调用 Lineman API 更新门店信息
    return nil
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: API 调用失败

- **处理方式**: 使用 `gerror.Wrapf` 包装错误，记录详细日志
- **用户影响**: 返回明确的错误信息，便于排查问题
- **代码示例**:
  ```go
  if err != nil {
      g.Log().Errorf(ctx, "[Lineman] AcceptOrder 失败: orderID=%s, error=%v", orderID, err)
      return gerror.Wrapf(err, "[Lineman] 接受订单失败: %s", orderID)
  }
  ```

#### 场景 2: 配置缺失或无效

- **处理方式**: 在首次调用时返回明确的错误信息，记录配置内容
- **用户影响**: 提示配置错误，便于快速修复
- **代码示例**:
  ```go
  conf := s.MustConf()
  if conf.ApiKey == "" {
      return gerror.New("[Lineman] API Key 配置缺失")
  }
  ```

#### 场景 3: HTTP 请求超时

- **处理方式**: 返回超时错误，记录请求详情
- **用户影响**: 提示网络异常或 API 响应慢
- **代码示例**:
  ```go
  client := &http.Client{
      Timeout: time.Duration(s.MustConf().Timeout) * time.Second,
  }
  // 如果超时，返回明确的错误信息
  ```

#### 场景 4: Webhook 签名验证失败

- **处理方式**: 返回验证失败错误，记录签名和请求体
- **用户影响**: 拒绝非法请求，保护系统安全
- **代码示例**:
  ```go
  if err := s.VerifyWebhookSignature(ctx, signature, timestamp, body); err != nil {
      g.Log().Warningf(ctx, "[Lineman] Webhook 签名验证失败: %v", err)
      return gerror.Wrap(err, "[Lineman] Webhook 签名验证失败")
  }
  ```

---

## 🔒 安全设计

### 身份验证

- **API Key 认证**: 所有 API 调用需要携带 API Key
- **Webhook 签名验证**: 验证 Lineman 发送的 Webhook 请求签名，防止伪造

### 权限控制

- 仅内部模块可以调用 Lineman Service，不对外暴露
- 通过 `service.Lineman()` 统一入口调用

### 数据安全

- **敏感配置**: API Key、Secret Key 使用环境变量，不硬编码
- **日志脱敏**: 日志中不记录完整的 API Key 和 Secret Key
- **HTTPS**: 与 Lineman API 通信使用 HTTPS

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- Service 层: 70%+

**测试内容**:

- 配置读取和懒加载单例
- 认证信息获取和缓存
- 订单/菜单/门店 API 调用逻辑
- 错误处理和日志记录

**示例**:

```go
// internal/logic/lineman/lineman_test.go
func TestLinemanService_AcceptOrder(t *testing.T) {
    // Mock HTTP 客户端
    // 测试 AcceptOrder 方法
    // 验证返回值和日志
}
```

### 集成测试

**测试内容**:

- 使用 Mock 数据测试所有 API 调用流程
- 使用 Lineman Staging 环境验证实际 API 调用

### Webhook 测试

**测试内容**:

- 测试签名验证逻辑
- 测试订单接收 Webhook 处理
- 测试菜单同步状态回调

---

## 📈 性能优化

### 优化策略

1. **HTTP 连接池**:
   - 使用连接池复用 HTTP 连接，减少连接开销

2. **Token 缓存**:
   - 使用 Redis 缓存 Token，提前刷新避免过期
   - 参考 Grab 的 Token 管理实现

3. **并发控制**:
   - 使用 `sync.RWMutex` 控制配置读取并发
   - 避免重复初始化

4. **超时配置**:
   - HTTP 请求超时时间可配置（默认 30 秒）
   - 避免长时间阻塞

### 性能指标

- HTTP 请求超时时间: < 30s（可配置）
- Token 缓存命中率: > 80%
- 并发调用: 支持多线程安全

---

## 📚 实现清单

### Phase 1: 基础设施

- [x] 创建目录结构
- [x] 定义配置结构 (`conf/lineman.go`)
- [x] 实现配置读取 (`logic/lineman/config.go`)
- [x] 实现主服务入口 (`logic/lineman/lineman.go`)
- [x] 注册服务到服务容器

### Phase 2: 认证模块

- [ ] 实现 Token 管理（如需要）
- [ ] 实现 Webhook 签名验证
- [ ] 实现认证信息缓存（Redis）

### Phase 3: DTO 定义

- [ ] 定义订单相关 DTO (`dto/lineman/order.go`)
- [ ] 定义菜单相关 DTO (`dto/lineman/menu.go`)
- [ ] 定义门店相关 DTO (`dto/lineman/store.go`)

### Phase 4: 订单管理

- [ ] 实现订单接收 Webhook 处理
- [ ] 实现接受订单 API
- [ ] 实现拒绝订单 API
- [ ] 实现取消订单 API
- [ ] 实现更新订单状态 API
- [ ] 实现标记订单准备完成 API

### Phase 5: 菜单管理

- [ ] 实现获取菜单 API
- [ ] 实现同步菜单 API
- [ ] 实现更新菜单状态 API

### Phase 6: 门店管理

- [ ] 实现获取门店状态 API
- [ ] 实现暂停门店营业 API
- [ ] 实现恢复门店营业 API
- [ ] 实现更新门店信息 API

### Phase 7: 测试

- [ ] 编写单元测试
- [ ] 编写集成测试（Mock）
- [ ] Staging 环境测试（真实 API）

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: 待审核

