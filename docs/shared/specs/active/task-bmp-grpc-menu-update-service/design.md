# gRPC 菜单更新服务 设计文档

> 本文档定义 gRPC 菜单项/修饰符更新服务的技术设计和实现方案。

## 📋 概述

将现有的 `UpdateMenuItem` 和 `UpdateMenuModifier` 内部服务方法暴露为 gRPC 服务，扩展 `MenuService` 服务定义，允许 TTPOS 主服务通过 gRPC 调用实时更新 GrabFood 菜单属性。

**设计原则**：
- 复用现有业务逻辑（`service.GrabMenu()`）
- 保持与现有 `MenuService` RPC 风格一致
- 使用统一的 `takeout.ApiResponse` 响应格式

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

本设计严格遵循 Go BMP 微服务规范：

- ✅ 使用 GoFrame 2.x 框架
- ✅ Proto 文件定义在 `manifest/protobuf/` 目录
- ✅ Controller 实现在 `internal/controller/rpc/` 目录
- ✅ 复用现有 `internal/logic/` 业务逻辑
- ✅ 不修改 dao/entity/do/ 目录（自动生成）

### API 设计规范 (api.mdc)

- ✅ 响应格式统一：`{code, message, data{}}`
- ✅ 使用 `takeout.ApiResponse` 包装响应
- ✅ 错误码规范：4001（参数错误）、5001（服务错误）、0（成功）

---

## 🔄 代码复用分析

### 可复用的现有组件

| 组件 | 路径 | 复用方式 |
|------|------|----------|
| **GrabMenu Service** | `internal/service/grab_menu.go` | 直接调用 `UpdateMenuItem()` 和 `UpdateMenuModifier()` |
| **DTO 定义** | `internal/model/dto/grab/menu_update.go` | 复用 `UpdateMenuItemReq`、`UpdateMenuModifierReq`、`UpdateMenuResult` |
| **Menu Controller** | `internal/controller/rpc/menu/menu.go` | 扩展添加新方法 |
| **Proto 定义** | `manifest/protobuf/menu/menu.proto` | 扩展添加新 RPC |

### 集成点

- **现有 gRPC 服务**: `MenuService` 已注册，只需添加新方法
- **统一响应**: 使用 `takeout.ApiResponse` + `anypb.Any`
- **日志记录**: 复用 `menu_log` 表记录操作日志

---

## 🏗️ 架构设计

### 分层设计

```
gRPC Client (TTPOS Main)
        │
        ▼
┌─────────────────────────────────────┐
│  Controller Layer (RPC)             │
│  menu.go - UpdateMenuItem()         │
│           - UpdateMenuModifier()    │
│  ┌─────────────────────────────┐    │
│  │ • 参数校验                   │    │
│  │ • Proto → DTO 转换          │    │
│  │ • 调用 Service              │    │
│  │ • DTO → Proto 响应转换      │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  Service Layer (Logic)              │
│  service.GrabMenu()                 │
│  ┌─────────────────────────────┐    │
│  │ • 业务逻辑（已存在）         │    │
│  │ • 调用 GrabFood SDK         │    │
│  │ • 记录操作日志              │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  External API                       │
│  GrabFood API                       │
│  PUT /partner/v1/merchants/menu/    │
│      record                         │
└─────────────────────────────────────┘
```

### 数据流

```
1. gRPC Request (UpdateMenuItemReq)
       │
       ▼
2. Controller 参数校验
       │
       ▼
3. Proto → DTO 转换 (api.UpdateMenuItemReq → grabDto.UpdateMenuItemReq)
       │
       ▼
4. 调用 service.GrabMenu().UpdateMenuItem()
       │
       ▼
5. GrabFood API 调用
       │
       ▼
6. 返回 UpdateMenuResult
       │
       ▼
7. DTO → Proto 响应转换 (grabDto.UpdateMenuResult → api.UpdateMenuItemResp)
       │
       ▼
8. 包装为 takeout.ApiResponse
       │
       ▼
9. gRPC Response
```

---

## 🔌 API 设计

### gRPC API

#### Protobuf 定义

**文件**: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`

```protobuf
syntax = "proto3";
package menu;
import "takeout_api.proto";
option go_package = "ttpos-bmp/app/ttpos-takeout/api/menu";

// ============= 新增消息定义 =============

// 高级定价配置
message AdvancedPricing {
  string key = 1;    // 定价键 (格式: serviceType.orderType.channel)
  int64 price = 2;   // 价格 (minor unit，单位：分)
}

// 购买能力配置
message Purchasability {
  string key = 1;          // 购买能力键 (格式: serviceType.orderType.channel)
  bool purchasable = 2;    // 是否可购买
}

// 更新菜单项请求
message UpdateMenuItemReq {
  string merchant_id = 1;                         // Grab MerchantID (必填)
  string item_id = 2;                             // 商品ID (partner item id, 必填)
  optional int64 price = 3;                       // 价格 (minor unit，单位：分)
  optional string available_status = 4;           // 可用状态: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE
  optional int64 max_stock = 5;                   // 库存数量
  repeated AdvancedPricing advanced_pricings = 6; // 高级定价配置
  repeated Purchasability purchasabilities = 7;   // 购买能力配置
  string request_id = 8;                          // 请求 ID (可选，用于追踪)
}

// 更新菜单项响应
message UpdateMenuItemResp {
  bool success = 1;          // 是否成功
  string merchant_id = 2;    // 商户ID
  string record_id = 3;      // 记录ID (ItemID)
  string record_type = 4;    // 记录类型: ITEM
  string error_code = 5;     // 错误码 (失败时)
  string error_message = 6;  // 错误信息 (失败时)
}

// 更新菜单修饰符请求
message UpdateMenuModifierReq {
  string merchant_id = 1;                         // Grab MerchantID (必填)
  string modifier_id = 2;                         // 修饰符ID (partner modifier id, 必填)
  string modifier_name = 3;                       // 修饰符名称 (用于定位记录, 必填)
  optional int64 price = 4;                       // 价格 (minor unit，单位：分)
  optional string available_status = 5;           // 可用状态: AVAILABLE, UNAVAILABLE
  optional bool is_free = 6;                      // 是否免费
  repeated AdvancedPricing advanced_pricings = 7; // 高级定价配置
  string request_id = 8;                          // 请求 ID (可选，用于追踪)
}

// 更新菜单修饰符响应
message UpdateMenuModifierResp {
  bool success = 1;          // 是否成功
  string merchant_id = 2;    // 商户ID
  string record_id = 3;      // 记录ID (ModifierID)
  string record_type = 4;    // 记录类型: MODIFIER
  string error_code = 5;     // 错误码 (失败时)
  string error_message = 6;  // 错误信息 (失败时)
}

// ============= 扩展 MenuService =============

service MenuService {
    // 获取菜单快照 (已有)
    rpc GetMenuSnapshot (GetMenuSnapshotReq) returns (takeout.ApiResponse) {}
    // 保存ttpos 菜单快照数据 (已有)
    rpc SaveMenuSnapshot (SaveMenuSnapshotReq) returns (takeout.ApiResponse) {}
    // 更新菜单项 (新增)
    rpc UpdateMenuItem (UpdateMenuItemReq) returns (takeout.ApiResponse) {}
    // 更新菜单修饰符 (新增)
    rpc UpdateMenuModifier (UpdateMenuModifierReq) returns (takeout.ApiResponse) {}
}
```

**生成代码**:

```bash
cd ttpos-bmp/app/ttpos-takeout
make proto
```

---

## 🧩 组件和接口

### Controller 层实现

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`

```go
package menu

import (
    "context"
    
    api "ttpos-bmp/app/ttpos-takeout/api/menu"
    "ttpos-bmp/app/ttpos-takeout/api/takeout"
    "ttpos-bmp/app/ttpos-takeout/internal/consts"
    grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
    "ttpos-bmp/app/ttpos-takeout/internal/service"

    "github.com/gogf/gf/v2/frame/g"
    "google.golang.org/protobuf/types/known/anypb"
)

// Controller 已存在，扩展添加新方法

// UpdateMenuItem 更新菜单项
// 参数：merchant_id、item_id 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuItem(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
    // 1. 参数校验
    if req.MerchantId == "" {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeInvalidParam),
            Message: consts.MsgMerchantIDEmpty,
        }, nil
    }
    if req.ItemId == "" {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeInvalidParam),
            Message: consts.MsgItemIDEmpty,
        }, nil
    }

    // 2. Proto → DTO 转换
    updateReq := &grabDto.UpdateMenuItemReq{
        MerchantID: req.MerchantId,
        ItemID:     req.ItemId,
    }
    
    // 处理可选字段
    if req.Price != nil {
        price := *req.Price
        updateReq.Price = &price
    }
    if req.AvailableStatus != nil {
        updateReq.AvailableStatus = *req.AvailableStatus
    }
    if req.MaxStock != nil {
        stock := *req.MaxStock
        updateReq.MaxStock = &stock
    }
    
    // 转换高级定价配置
    for _, ap := range req.AdvancedPricings {
        updateReq.AdvancedPricings = append(updateReq.AdvancedPricings, 
            grabDto.UpdateAdvancedPricingReq{
                Key:   ap.Key,
                Price: ap.Price,
            })
    }
    
    // 转换购买能力配置
    for _, p := range req.Purchasabilities {
        updateReq.Purchasabilities = append(updateReq.Purchasabilities,
            grabDto.UpdatePurchasabilityReq{
                Key:         p.Key,
                Purchasable: p.Purchasable,
            })
    }

    // 3. 调用 Service 层
    result, err := service.GrabMenu().UpdateMenuItem(ctx, updateReq)
    if err != nil {
        g.Log().Errorf(ctx, "[Menu] UpdateMenuItem failed: merchantID=%s, itemID=%s, error=%v",
            req.MerchantId, req.ItemId, err)
        return &takeout.ApiResponse{
            Code:    string(consts.CodeServiceError),
            Message: "更新菜单项失败: " + err.Error(),
        }, nil
    }

    // 4. DTO → Proto 响应转换
    resp := &api.UpdateMenuItemResp{
        Success:      result.Success,
        MerchantId:   result.MerchantID,
        RecordId:     result.RecordID,
        RecordType:   result.RecordType,
        ErrorCode:    result.ErrorCode,
        ErrorMessage: result.ErrorMessage,
    }

    // 5. 包装为 ApiResponse
    dataAny, err := anypb.New(resp)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeSerializeError),
            Message: consts.MsgSerializeFailed,
        }, nil
    }

    g.Log().Infof(ctx, "[Menu] UpdateMenuItem success: merchantID=%s, itemID=%s",
        req.MerchantId, req.ItemId)
    return &takeout.ApiResponse{
        Code:    string(consts.CodeSuccess),
        Message: consts.MsgSuccess,
        Data:    dataAny,
    }, nil
}

// UpdateMenuModifier 更新菜单修饰符
// 参数：merchant_id、modifier_id、modifier_name 必填，其他字段可选
// 返回：takeout.ApiResponse 统一响应格式
func (c *Controller) UpdateMenuModifier(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
    // 1. 参数校验
    if req.MerchantId == "" {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeInvalidParam),
            Message: consts.MsgMerchantIDEmpty,
        }, nil
    }
    if req.ModifierId == "" {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeInvalidParam),
            Message: consts.MsgModifierIDEmpty,
        }, nil
    }
    if req.ModifierName == "" {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeInvalidParam),
            Message: consts.MsgModifierNameEmpty,
        }, nil
    }

    // 2. Proto → DTO 转换
    updateReq := &grabDto.UpdateMenuModifierReq{
        MerchantID:   req.MerchantId,
        ModifierID:   req.ModifierId,
        ModifierName: req.ModifierName,
    }
    
    // 处理可选字段
    if req.Price != nil {
        price := *req.Price
        updateReq.Price = &price
    }
    if req.AvailableStatus != nil {
        updateReq.AvailableStatus = *req.AvailableStatus
    }
    if req.IsFree != nil {
        isFree := *req.IsFree
        updateReq.IsFree = &isFree
    }
    
    // 转换高级定价配置
    for _, ap := range req.AdvancedPricings {
        updateReq.AdvancedPricings = append(updateReq.AdvancedPricings,
            grabDto.UpdateAdvancedPricingReq{
                Key:   ap.Key,
                Price: ap.Price,
            })
    }

    // 3. 调用 Service 层
    result, err := service.GrabMenu().UpdateMenuModifier(ctx, updateReq)
    if err != nil {
        g.Log().Errorf(ctx, "[Menu] UpdateMenuModifier failed: merchantID=%s, modifierID=%s, error=%v",
            req.MerchantId, req.ModifierId, err)
        return &takeout.ApiResponse{
            Code:    string(consts.CodeServiceError),
            Message: "更新菜单修饰符失败: " + err.Error(),
        }, nil
    }

    // 4. DTO → Proto 响应转换
    resp := &api.UpdateMenuModifierResp{
        Success:      result.Success,
        MerchantId:   result.MerchantID,
        RecordId:     result.RecordID,
        RecordType:   result.RecordType,
        ErrorCode:    result.ErrorCode,
        ErrorMessage: result.ErrorMessage,
    }

    // 5. 包装为 ApiResponse
    dataAny, err := anypb.New(resp)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    string(consts.CodeSerializeError),
            Message: consts.MsgSerializeFailed,
        }, nil
    }

    g.Log().Infof(ctx, "[Menu] UpdateMenuModifier success: merchantID=%s, modifierID=%s",
        req.MerchantId, req.ModifierId)
    return &takeout.ApiResponse{
        Code:    string(consts.CodeSuccess),
        Message: consts.MsgSuccess,
        Data:    dataAny,
    }, nil
}
```

---

## 🚨 错误处理

### 错误码常量定义

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/consts/response.go`

```go
package consts

// ResponseCode gRPC 响应码常量
// 用于 takeout.ApiResponse 的 Code 字段
type ResponseCode string

const (
    // CodeSuccess 成功
    CodeSuccess ResponseCode = "0"
    
    // CodeInvalidParam 参数校验失败 (4xxx 系列)
    CodeInvalidParam ResponseCode = "4001"
    
    // CodeServiceError 服务内部错误 (5xxx 系列)
    CodeServiceError ResponseCode = "5001"
    
    // CodeSerializeError 序列化错误
    CodeSerializeError ResponseCode = "5002"
    
    // CodeExternalAPIError 外部 API 调用错误
    CodeExternalAPIError ResponseCode = "5003"
)

// ResponseMessage 响应消息常量
const (
    MsgSuccess           = "success"
    MsgSerializeFailed   = "序列化响应数据失败"
    MsgMerchantIDEmpty   = "merchant_id 不能为空"
    MsgItemIDEmpty       = "item_id 不能为空"
    MsgModifierIDEmpty   = "modifier_id 不能为空"
    MsgModifierNameEmpty = "modifier_name 不能为空"
)
```

### 错误码规范

| 错误码 | 常量名 | 场景 | 示例消息 |
|--------|--------|------|----------|
| `0` | `CodeSuccess` | 成功 | `success` |
| `4001` | `CodeInvalidParam` | 参数校验失败 | `merchant_id 不能为空` |
| `5001` | `CodeServiceError` | 服务调用失败 | `更新菜单项失败: {error}` |
| `5002` | `CodeSerializeError` | 响应序列化失败 | `序列化响应数据失败` |
| `5003` | `CodeExternalAPIError` | 外部 API 错误 | `GrabFood API 调用失败: {error}` |

### 使用示例

```go
// 参数错误
return &takeout.ApiResponse{
    Code:    string(consts.CodeInvalidParam),
    Message: consts.MsgMerchantIDEmpty,
}, nil

// 成功响应
return &takeout.ApiResponse{
    Code:    string(consts.CodeSuccess),
    Message: consts.MsgSuccess,
    Data:    dataAny,
}, nil

// 服务错误
return &takeout.ApiResponse{
    Code:    string(consts.CodeServiceError),
    Message: "更新菜单项失败: " + err.Error(),
}, nil
```

### 错误处理策略

1. **参数错误**: 返回 4001，不记录日志
2. **Service 错误**: 返回 5001，记录 Error 日志，Service 层会自动记录到 `menu_log` 表
3. **序列化错误**: 返回 5001，记录 Error 日志

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**: 70%+

**测试内容**:
- Controller 参数校验（空值、格式错误）
- Proto → DTO 转换逻辑
- 可选字段处理

**测试文件**: `internal/controller/rpc/menu/menu_test.go`

### 集成测试

**测试方式**: 使用 grpcurl

```bash
# 测试 UpdateMenuItem
grpcurl -plaintext -d '{
  "merchant_id": "1-ABCD1234EFGH5678",
  "item_id": "item-001",
  "price": 1990,
  "available_status": "AVAILABLE"
}' localhost:9001 menu.MenuService/UpdateMenuItem

# 测试 UpdateMenuModifier
grpcurl -plaintext -d '{
  "merchant_id": "1-ABCD1234EFGH5678",
  "modifier_id": "mod-001",
  "modifier_name": "Extra Cheese",
  "price": 500,
  "is_free": false
}' localhost:9001 menu.MenuService/UpdateMenuModifier
```

---

## 📈 性能考虑

### 响应时间

- **Controller 处理**: < 10ms
- **Service 处理**: < 50ms
- **GrabFood API**: 100-500ms（外部依赖）
- **总响应时间**: 约 200-600ms

### 日志记录

- **成功**: Info 级别日志
- **失败**: Error 级别日志 + `menu_log` 表记录

---

## 📚 实现清单

### Phase 1: Proto 定义

- [ ] 1.1 更新 `menu.proto` 添加新消息类型
- [ ] 1.2 更新 `menu.proto` 添加新 RPC 方法
- [ ] 1.3 执行 `make proto` 生成代码

### Phase 2: Controller 实现

- [ ] 2.1 实现 `UpdateMenuItem` 方法
- [ ] 2.2 实现 `UpdateMenuModifier` 方法
- [ ] 2.3 添加必要的 import

### Phase 3: 测试验证

- [ ] 3.1 编写单元测试
- [ ] 3.2 使用 grpcurl 进行集成测试
- [ ] 3.3 验证日志记录

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: AI Agent  
**审核者**: 待定

