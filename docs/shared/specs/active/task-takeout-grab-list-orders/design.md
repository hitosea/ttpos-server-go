# task-takeout-grab-list-orders 技术设计

## 📋 概述

| 项目 | 内容 |
|------|------|
| **Spec ID** | task-takeout-grab-list-orders |
| **设计人** | rikugun |
| **设计日期** | 2026-01-23 |
| **总 SP** | 2 |

---

## 🔄 代码复用分析

### 可复用代码

| 文件 | 说明 | 复用方式 |
|------|------|---------|
| `internal/client/grab/client.go` | SDK Client 封装 | 直接调用 `grab.Default()` |
| `internal/client/grab/client.go` | 授权头获取 | 直接调用 `GetAuthorizationHeader()` |
| `internal/client/grab/client.go` | SDK Context | 直接调用 `GetSDKContext()` |
| `internal/logic/grab/grab_api.go` | SDK 错误处理 | 参考 `HandleSDKError()` 调用模式 |

### 需要新建/修改

| 文件 | 说明 |
|------|------|
| `internal/logic/grab/grab_api.go` | 新增 ListOrders 方法 |
| `internal/service/grab.go` | 更新服务接口 |

---

## 🏗️ 架构设计

### 架构图

```mermaid
graph TD
    A[内部服务调用] -->|service.Grab().ListOrders| B[Logic Layer<br/>internal/logic/grab/grab_api.go]
    B -->|GetAuthorizationHeader| C[Client Layer<br/>internal/client/grab/]
    B -->|ListOrdersAPI.Execute| D[GrabFood SDK<br/>grabfood-api-sdk-go]
    D -->|HTTP| E[Grab API<br/>partner-api.grab.com]
    B -->|返回| F[SDK Response<br/>*grabfood.ListOrdersResponse]
```

### 分层说明

| 层级 | 路径 | 职责 |
|------|------|------|
| **Service** | `internal/service/grab.go` | 接口定义（自动生成） |
| **Logic** | `internal/logic/grab/grab_api.go` | SDK 调用封装，直接返回 SDK 类型 |
| **Client** | `internal/client/grab/` | SDK Client 管理，Token 管理 |

### 设计决策

**使用自定义 DTO + 原始 JSON 解析**：不直接使用 SDK 返回类型，而是解析原始 HTTP 响应到自定义 `ListOrdersResponse` DTO，内部使用 `TakeoutOrder` 类型。

**理由**：
1. **字段完整性**：SDK 的 `OrderPrice` 缺少 `Total` 字段，但 Grab API 实际返回该字段
2. **统一数据模型**：使用 `ttpos-api/ttpos-takeout/message.TakeoutOrder`，与其他外卖平台保持一致
3. **简化架构**：不暴露 gRPC 接口，仅供内部服务调用
4. **内部服务专用**：此接口仅供内部服务调用（订单同步、数据对账）

### 数据流

```
1. 内部服务调用 service.Grab().ListOrders(...)
2. Logic 层参数校验（merchant_id 必填）
3. Logic 层获取 Authorization Header
4. 构建 SDK ApiListOrdersRequest
5. 执行 SDK ListOrdersAPI.Execute()
6. 读取原始 HTTP 响应体
7. 解析 JSON 到自定义 ListOrdersResponse（含 TakeoutOrder）
8. 返回 *grabDto.ListOrdersResponse
```

---

## 🧩 组件和接口

### DTO 定义

**位置**: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/list_orders.go`

```go
// ListOrdersResponse Grab ListOrders API 响应
// 使用 TakeoutOrder 以支持 Price.Total 字段（SDK 缺失该字段）
type ListOrdersResponse struct {
    Orders []message.TakeoutOrder `json:"orders"`
    More   bool                   `json:"more"`
}
```

### Service 接口

**位置**: `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go`

```go
// ListOrders 查询 Grab 订单列表
// 封装 GrabFood SDK ListOrdersAPI，支持按商户、日期、订单ID等维度查询
// 注意：此接口在 Grab 测试环境下不可用，仅生产环境支持
// 参数：
//   - merchantID: Grab 商户 ID（必填）
//   - date: 日期过滤，格式 YYYY-MM-DD（可选）
//   - orderIDs: 订单 ID 列表过滤（可选）
//   - page: 分页页码（可选）
// 返回：
//   - resp: ListOrdersResponse（使用 TakeoutOrder 以获取完整 Price 含 Total）
//   - err: 错误信息
ListOrders(ctx context.Context, merchantID string, date string, orderIDs []string, page int32) (*grabDto.ListOrdersResponse, error)
```

### Logic 层实现

**位置**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_api.go`

```go
// ListOrders 查询 Grab 订单列表
// 封装 GrabFood SDK ListOrdersAPI，支持按商户、日期、订单ID等维度查询
// 注意：此接口在 Grab 测试环境下不可用，仅生产环境支持
func (s *sGrab) ListOrders(ctx context.Context, merchantID string, date string, orderIDs []string, page int32) (*grabDto.ListOrdersResponse, error) {
    // 1. 参数验证
    if merchantID == "" {
        return nil, gerror.NewCode(gcode.CodeInvalidParameter, "merchant_id 不能为空")
    }

    // 2. 获取 Client 和授权
    client := grab.Default()
    auth, err := client.GetAuthorizationHeader(ctx)
    if err != nil {
        return nil, gerror.Wrap(err, "获取授权信息失败")
    }

    // 3. 构建请求并执行
    req := client.GetClient().ListOrdersAPI.
        ListOrders(client.GetSDKContext(ctx)).
        Authorization(auth).
        MerchantID(merchantID)
    // ... 可选参数设置 ...

    // 4. 执行请求（获取原始 HTTP 响应以解析完整字段）
    _, httpResp, err := req.Execute()
    if err = client.HandleSDKError(ctx, err, "ListOrders"); err != nil {
        return nil, err
    }
    defer httpResp.Body.Close()

    // 5. 读取原始响应体并解析为自定义结构（SDK 的 OrderPrice 缺少 Total 字段）
    bodyBytes, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return nil, gerror.Wrap(err, "读取响应体失败")
    }

    var resp grabDto.ListOrdersResponse
    if err := json.Unmarshal(bodyBytes, &resp); err != nil {
        return nil, gerror.Wrap(err, "解析 ListOrders 响应失败")
    }

    return &resp, nil
}
```

---

## 📊 数据模型

### 响应类型

使用自定义 `ListOrdersResponse` DTO，内部包含 `TakeoutOrder` 类型：

```go
// grabDto.ListOrdersResponse
type ListOrdersResponse struct {
    Orders []message.TakeoutOrder `json:"orders"`
    More   bool                   `json:"more"`
}

// message.TakeoutOrderPrice（关键：包含 Total 字段）
type TakeoutOrderPrice struct {
    Subtotal          int64  `json:"subtotal"`
    Tax               *int64 `json:"tax,omitempty"`
    MerchantChargeFee *int64 `json:"merchantChargeFee,omitempty"`
    // ... 其他字段 ...
    Total             *int64 `json:"total,omitempty"`  // SDK 缺失此字段
}
```

**设计优势**：
1. **字段完整**：包含 Grab API 返回但 SDK 未定义的 `Price.Total` 字段
2. **统一模型**：使用 `TakeoutOrder`，与 Lineman 等其他平台保持一致
3. **类型安全**：编译时检查，无需手动解析 JSON 字段

---

## 🔌 API 设计

### 内部服务调用

| 项目 | 内容 |
|------|------|
| **Service** | `service.Grab()` |
| **Method** | `ListOrders` |
| **返回类型** | `*grabDto.ListOrdersResponse` |

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| merchantID | string | 是 | Grab 商户 ID |
| date | string | 否 | 日期过滤，格式 YYYY-MM-DD |
| orderIDs | []string | 否 | 订单 ID 列表 |
| page | int32 | 否 | 分页页码，默认 1 |

### 调用示例

```go
import (
    "ttpos-bmp/app/ttpos-takeout/internal/service"
)

// 查询商户所有订单
resp, err := service.Grab().ListOrders(ctx, "merchant-123", "", nil, 0)
if err != nil {
    return err
}

// 按日期查询
resp, err := service.Grab().ListOrders(ctx, "merchant-123", "2026-01-23", nil, 0)

// 按订单 ID 查询
resp, err := service.Grab().ListOrders(ctx, "merchant-123", "", []string{"order-1", "order-2"}, 0)

// 分页查询
resp, err := service.Grab().ListOrders(ctx, "merchant-123", "", nil, 2)

// 处理响应（使用 TakeoutOrder 结构体字段访问）
for _, order := range resp.Orders {
    fmt.Printf("Order ID: %s, State: %v\n", order.OrderID, order.OrderState)
    // 访问 Price.Total（SDK 不支持此字段）
    if order.Price.Total != nil {
        fmt.Printf("  Total: %d\n", *order.Price.Total)
    }
}
if resp.More {
    // 继续查询下一页
}
```

---

## ⚠️ 风险识别

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| **测试环境不可用** | 高 | Grab 平台限制，此 API 仅生产环境支持，测试需使用生产凭据 |
| Grab API Rate Limit | 中 | 记录调用日志，监控调用频率 |
| 网络超时 | 低 | 复用现有 HandleSDKError 错误处理 |

### 重要提示：测试环境限制

> **Grab 平台限制**：`ListOrders` API 在 Grab 测试环境（Staging）下**不可用**，仅生产环境支持。
>
> 开发和测试时请注意：
> - 单元测试需要 Mock SDK 调用
> - 集成测试需要使用生产环境凭据
> - 建议在上线前进行有限的生产环境验证

---

## 🧪 测试策略

### 测试范围

| 层级 | 文件 | 覆盖率目标 |
|------|------|-----------|
| Logic | `grab_api_test.go` | 80%+ |

### 测试用例

1. **正常查询**：传入有效 merchant_id，验证返回订单列表
2. **日期过滤**：传入 date 参数，验证只返回指定日期订单
3. **订单 ID 过滤**：传入 order_ids，验证只返回指定订单
4. **分页查询**：验证 more 字段正确返回
5. **参数校验**：merchant_id 为空时返回错误
6. **SDK 错误**：模拟 SDK 返回错误，验证错误处理

### 测试命令

```bash
cd ttpos-bmp/app/ttpos-takeout && go test -v ./internal/logic/grab/... -run TestListOrders
```

---

## 📝 实现检查清单

### 代码规范

- [x] 使用 `gerror` 处理错误
- [x] 日志包含关键参数（merchant_id, date, page）
- [x] 日志记录响应订单数量
- [x] 遵循现有 SDK 调用模式
- [x] 返回 SDK 原生类型

### BMP 规范

- [x] Logic 层代码在 internal/logic/grab/
- [x] 禁止修改自动生成文件
- [x] Service 接口更新

---

**版本**: v2.0.0
**创建日期**: 2026-01-23
**更新日期**: 2026-01-23
**更新说明**: 移除 Proto/gRPC 设计，改为直接使用 SDK 类型
