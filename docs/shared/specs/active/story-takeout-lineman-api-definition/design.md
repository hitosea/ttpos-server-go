# LINE MAN API 定义 设计文档

> 本文档定义 LINE MAN Webhook API 的技术设计和实现方案。

## 📋 概述

本设计文档描述如何在 ttpos-takeout 模块中使用 GoFrame 框架定义 LINE MAN Webhook API 的数据结构。这是一个纯 API 定义任务，不涉及业务逻辑实现，只创建符合 GoFrame 规范的 Request 和 Response 结构体。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

本设计严格遵循 GoFrame 开发规范：

- ✅ 使用 GoFrame 2.x 框架标签系统
- ✅ API 定义放在 `api/lineman/v1/` 目录
- ✅ 请求结构体以 `Req` 结尾
- ✅ 响应结构体以 `Resp` 结尾
- ✅ 使用 `g.Meta` 标签定义路由
- ✅ 使用 `v` 标签定义验证规则
- ✅ 使用 `json` 标签定义 JSON 映射
- ✅ 使用 `dc` 标签添加字段描述
- ✅ 所有注释使用中文

### API 设计规范 (api.mdc)

- ✅ URL 使用 snake_case
- ✅ 响应格式统一：`{status, code, message}`
- ✅ 路径参数使用 `:paramName` 格式

---

## 🔄 代码复用分析

### 参考文档

本API定义基于以下 LINE MAN 官方文档：

- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/io-auth.md`
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-place-order.md`
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-order-status-update-notification.md`
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-order-update-notification.md`
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-menu-sync-notification.md`
- `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-trigger-sync-menu.md`

### 可参考的现有代码

- `ttpos-bmp/app/ttpos-takeout/api/grab/v1/` - GrabFood API 定义（结构参考）
- GoFrame 标签使用示例

---

## 🏗️ 架构设计

### 文件结构

```
ttpos-bmp/app/ttpos-takeout/api/lineman/v1/
├── oauth.go              # OAuth 认证 API 定义（1个API）
├── order.go              # 订单相关 API 定义（3个API）
├── menu.go               # 菜单同步 API 定义（2个API）
└── common.go             # 通用响应格式
```

### 模块划分说明

#### oauth.go

**职责**: OAuth 认证相关的 API 定义

**内容**:
- `OAuthTokenReq` - OAuth 令牌请求
- `OAuthTokenResp` - OAuth 令牌响应

**特点**:
- Content-Type: `application/x-www-form-urlencoded`
- 固定授权类型: `client_credentials`

#### order.go

**职责**: 订单相关的 API 定义

**内容**:
- `PlaceOrderReq` - 订单创建请求
- `OrderStatusUpdateReq` - 订单状态更新请求
- `OrderUpdateReq` - 订单更新通知请求
- `OrderItem` - 订单商品（嵌套结构）
- `OrderItemProperty` - 商品属性（嵌套结构）
- `OrderItemPropertyValue` - 属性值（嵌套结构）
- `OrderAdditionalItem` - 订单附加项（嵌套结构）

**特点**:
- Content-Type: `application/json`
- 包含复杂的嵌套结构
- 路径参数: `partnerId`, `storeId`

#### menu.go

**职责**: 菜单同步相关的 API 定义

**内容**:
- `MenuSyncNotificationReq` - 菜单同步通知请求
- `TriggerSyncMenuReq` - 菜单同步触发请求

**特点**:
- Content-Type: `application/json`
- 相对简单的数据结构

#### common.go

**职责**: 通用响应格式定义

**内容**:
- `LinemanCommonResp` - LINE MAN 统一响应格式

**特点**:
- 所有 API 返回此格式
- `status`: "ok" 或 "fail"
- `code`: 结果代码
- `message`: 结果描述（可选）

---

## 📊 数据模型

### GoFrame 标签系统

```go
type ExampleReq struct {
    // g.Meta 标签 - 定义路由和 HTTP 方法
    g.Meta    `path:"/api/v1/example" method:"post" tags:"Example API" summary:"示例接口"`
    
    // v 标签 - 定义验证规则（中文错误提示）
    FieldName string `json:"field_name" v:"required|length:1,20#字段不能为空|字段长度为1-20个字符" dc:"字段描述"`
}
```

### 标签说明

| 标签 | 用途 | 示例 |
|------|------|------|
| `g.Meta` | 路由定义 | `path:"/api/v1/xxx" method:"post"` |
| `v` | 验证规则 | `required\|length:1,20#错误提示1\|错误提示2` |
| `json` | JSON 映射 | `json:"field_name"` |
| `dc` | 字段描述 | `dc:"字段的详细说明"` |

**路由前缀说明**:
- API 定义中的 `path` 不包含 `/v1/lmwn/` 前缀
- 实际路由绑定时通过路由分组手动添加前缀
- 例如：`path:"/oauth2/token"` 最终路由为 `/v1/lmwn/oauth2/token`

### 路径参数处理

```go
type ExampleReq struct {
    g.Meta     `path:"/partners/:partnerId/stores/:storeId" method:"post"`
    PartnerId  string `json:"partnerId" v:"required#合作伙伴ID不能为空"`
    StoreId    string `json:"storeId" v:"required#门店ID不能为空"`
    // ... 其他字段
}
```

**注意**: path 中不包含 `/v1/lmwn/` 前缀，路由绑定时通过路由分组添加。

---

## 🔌 API 设计

### API 1: OAuth 认证

**端点**: `POST /v1/lmwn/oauth2/token`  
**API 路径**: `/oauth2/token`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`OAuthTokenReq`):
```go
type OAuthTokenReq struct {
    g.Meta       `path:"/oauth2/token" method:"post" tags:"LINE MAN OAuth" summary:"OAuth 认证接口"`
    GrantType    string `json:"grant_type" v:"required|in:client_credentials#授权类型不能为空|授权类型必须为client_credentials" dc:"OAuth 授权类型，固定值：client_credentials"`
    ClientId     string `json:"client_id" v:"required#客户端ID不能为空" dc:"LINE MAN 分配的客户端 ID"`
    ClientSecret string `json:"client_secret" v:"required#客户端密钥不能为空" dc:"LINE MAN 分配的客户端密钥"`
}
```

**Response** (`OAuthTokenRes`):
```go
type OAuthTokenRes struct {
    g.Meta      `mime:"application/json"`
    AccessToken string `json:"access_token" dc:"访问令牌，用于后续 API 调用"`
    TokenType   string `json:"token_type" dc:"令牌类型，固定值：Bearer"`
    ExpiresIn   int    `json:"expires_in" dc:"令牌有效期（秒），通常为 3600"`
}
```

**特点**:
- Content-Type: `application/x-www-form-urlencoded`
- 双向 API: LINE MAN ← TTPOS

---

### API 2: 订单创建

**端点**: `POST /v1/lmwn/partners/:partnerId/stores/:storeId/orders`  
**API 路径**: `/partners/:partnerId/stores/:storeId/orders`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`PlaceOrderReq`):
```go
type PlaceOrderReq struct {
    g.Meta            `path:"/partners/:partnerId/stores/:storeId/orders" method:"post" tags:"LINE MAN Order" summary:"接收订单创建通知"`
    PartnerId         string               `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID"`
    StoreId           string               `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID"`
    OrderId           string               `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度不能超过20" dc:"订单唯一 ID"`
    OrderShortCode    string               `json:"orderShortCode" v:"required|length:4,4#短订单ID不能为空|短订单ID必须为4位" dc:"短订单 ID"`
    RestaurantRevenue float64              `json:"restaurantRevenue" v:"required#商户收入不能为空" dc:"商户收入总额"`
    OrderAcceptedTime string               `json:"orderAcceptedTime" v:"required#订单接受时间不能为空" dc:"订单接受时间，ISO 8601 格式"`
    Items             []OrderItem          `json:"items" v:"required#订单商品不能为空" dc:"订单商品列表"`
    AdditionalItems   []OrderAdditionalItem `json:"additionalItems" dc:"订单附加项列表"`
    MemberId          string               `json:"memberId" dc:"绑定 LINE MAN 账号的会员 ID"`
    CustomerType      string               `json:"customerType" v:"required|in:DELIVERY,PICKUP#订单类型不能为空|订单类型必须为DELIVERY或PICKUP" dc:"订单类型"`
}
```

**嵌套结构体**: `OrderItem`, `OrderItemProperty`, `OrderItemPropertyValue`, `OrderAdditionalItem`

**Response**: `LinemanCommonResp`

---

### API 3: 订单状态更新

**端点**: `POST /v1/lmwn/partners/:partnerId/stores/:storeId/order/status`  
**API 路径**: `/partners/:partnerId/stores/:storeId/order/status`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`OrderStatusUpdateReq`):
```go
type OrderStatusUpdateReq struct {
    g.Meta      `path:"/partners/:partnerId/stores/:storeId/order/status" method:"post" tags:"LINE MAN Order" summary:"接收订单状态更新通知"`
    PartnerId   string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID"`
    StoreId     string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID"`
    OrderId     string `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度不能超过20" dc:"订单唯一 ID"`
    OrderStatus string `json:"orderStatus" v:"required|in:FINISH,CANCELED#订单状态不能为空|订单状态必须为FINISH或CANCELED" dc:"订单状态"`
}
```

**Response**: `LinemanCommonResp`

---

### API 4: 订单更新通知

**端点**: `PUT /v1/lmwn/partners/:partnerId/stores/:storeId/orders`  
**API 路径**: `/partners/:partnerId/stores/:storeId/orders`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`OrderUpdateReq`):
- 类似 `PlaceOrderReq`
- 额外包含 `orderUpdatedTime` 字段
- `g.Meta` 使用 `method:"put"`

**Response** (`OrderUpdateRes`):
```go
type OrderUpdateRes struct {
    g.Meta  `mime:"application/json"`
    Status  string `json:"status" dc:"结果状态：ok 表示成功，fail 表示失败"`
    Code    string `json:"code" dc:"结果代码"`
    Message string `json:"message,omitempty" dc:"结果描述"`
}
```

---

### API 5: 菜单同步通知

**端点**: `POST /v1/lmwn/partners/:partnerId/stores/:storeId/menus/notification`  
**API 路径**: `/partners/:partnerId/stores/:storeId/menus/notification`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`MenuSyncNotificationReq`):
```go
type MenuSyncNotificationReq struct {
    g.Meta            `path:"/partners/:partnerId/stores/:storeId/menus/notification" method:"post" tags:"LINE MAN Menu" summary:"接收菜单同步通知"`
    PartnerId         string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID"`
    StoreId           string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID"`
    MenuSyncRequestId string `json:"menuSyncRequestId" v:"required#菜单同步请求ID不能为空" dc:"菜单同步请求唯一 ID"`
    UpdatedAt         string `json:"updatedAt" v:"required#更新时间不能为空" dc:"更新时间，ISO 8601 格式"`
    Status            string `json:"status" v:"required|in:SUCCESS,FAILED#状态不能为空|状态必须为SUCCESS或FAILED" dc:"菜单同步结果状态"`
    Error             string `json:"error" dc:"错误信息，状态为 SUCCESS 时为空"`
}
```

**Response**: `LinemanCommonResp`

---

### API 6: 菜单同步触发

**端点**: `POST /v1/lmwn/partners/:partnerId/stores/:storeId/menus/trigger-sync`  
**API 路径**: `/partners/:partnerId/stores/:storeId/menus/trigger-sync`（路由绑定时添加 `/v1/lmwn/` 前缀）

**Request** (`TriggerSyncMenuReq`):
```go
type TriggerSyncMenuReq struct {
    g.Meta    `path:"/partners/:partnerId/stores/:storeId/menus/trigger-sync" method:"post" tags:"LINE MAN Menu" summary:"接收菜单同步触发请求"`
    PartnerId string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID"`
    StoreId   string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID"`
}
```

**Response**: `LinemanCommonResp`

---

## 🧩 通用响应格式

所有 API 使用统一的响应格式：

```go
// LinemanCommonResp LINE MAN 通用响应格式
type LinemanCommonResp struct {
    Status  string `json:"status" dc:"结果状态：ok 表示成功，fail 表示失败"`
    Code    string `json:"code" dc:"结果代码"`
    Message string `json:"message,omitempty" dc:"结果描述"`
}
```

**状态值**:
- `"ok"` - 成功
- `"fail"` - 失败

---

## 🚨 错误处理

### 验证错误

所有验证错误通过 GoFrame 的 `v` 标签自动处理，返回中文错误提示。

**示例**:
```go
`v:"required#字段不能为空"`
`v:"required|length:1,20#字段不能为空|字段长度为1-20个字符"`
`v:"required|in:VALUE1,VALUE2#字段不能为空|字段必须为VALUE1或VALUE2"`
```

---

## 📈 性能优化

本需求仅涉及 API 定义，不涉及性能优化。

---

## 🌐 国际化

本需求中的错误提示使用中文，符合项目要求。

---

## 📚 实现清单

### Phase 1: API 定义

- [x] 创建 oauth.go - OAuth 认证 API
- [x] 创建 order.go - 订单相关 API（3个）
- [x] 创建 menu.go - 菜单同步 API（2个）
- [x] 创建 common.go - 通用响应格式

### Phase 2: 文档

- [ ] 更新集成说明文档 `ttpos-bmp/docs/shared/integrations/lineman.md`
- [ ] 完善代码注释

### Phase 3: 验证

- [ ] 执行 `go fmt` 格式化
- [ ] 执行 `go vet` 检查
- [ ] 与 LINE MAN 文档对比验证

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`

---

**版本**: v2.13.1  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: rikugun

