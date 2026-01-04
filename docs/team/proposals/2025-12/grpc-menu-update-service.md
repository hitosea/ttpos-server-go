# gRPC 菜单更新服务 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目         | 内容                                |
| ------------ | ----------------------------------- |
| **提案人**   | AI Agent                            |
| **日期**     | 2025-12-16                          |
| **目标版本** | v2.10.x                             |
| **状态**     | 已创建 Spec                         |
| **关联任务** | -                                   |
| **关联 Spec** | [task-bmp-grpc-menu-update-service](../../../shared/specs/archived/v2.12/task-bmp-grpc-menu-update-service/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `grab_menu.go` 中的 `UpdateMenuItem` 和 `UpdateMenuModifier` 方法仅作为内部 Service 方法存在，无法被其他微服务（如 TTPOS 主服务）通过 gRPC 调用。这限制了菜单项实时更新能力的跨服务集成。

### 业务价值

- 允许 TTPOS 主服务实时同步菜单项/修饰符的价格、库存、可用状态到 GrabFood
- 减少菜单全量同步的需求，提升更新效率
- 支持商家在 TTPOS 系统中直接管理 GrabFood 商品状态

### 目标用户

- [x] 商户管理员
- [x] 其他: TTPOS 主服务、BMP 后台

---

## 💡 解决方案概述

### 方案描述

在现有的 `menu.proto` 中新增两个 RPC 方法：`UpdateMenuItem` 和 `UpdateMenuModifier`，保持与现有 `GetMenuSnapshot`、`SaveMenuSnapshot` 相同的风格。Controller 层调用现有的 `service.GrabMenu()` 方法实现业务逻辑。

### 核心功能点

1. **UpdateMenuItem RPC**：更新单个菜单项（价格、库存、可用状态、高级定价、购买能力）
2. **UpdateMenuModifier RPC**：更新单个修饰符（价格、可用状态、是否免费、高级定价）
3. **统一响应格式**：使用 `takeout.ApiResponse` 包装结果

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [x] Shop 商家管理端

**涉及模块**：
- [x] API 接口
- [x] 业务逻辑
- [x] 第三方集成（GrabFood API）

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯 API 层包装，业务逻辑已存在

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. Proto 结构需要支持可选字段（价格、库存等）
2. 需确保与现有 HTTP API 的一致性

**缓解措施**：
1. 使用 `optional` 或包装类型处理可选字段
2. 复用现有 DTO 转换逻辑

---

## 🔗 相关资源

### 参考需求

- 现有实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
- Proto 风格参考: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`

### 相关文档

- GrabFood API 文档: [Update Menu Record API](https://developer.grab.com/docs/grabfood/api-reference/update-menu-record)

---

## 🛠 技术方案

### 1. Proto 定义更新

**文件**: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`

```proto
syntax = "proto3";
package menu;
import "takeout_api.proto" ;
option go_package = "ttpos-bmp/app/ttpos-takeout/api/menu";

// ============= 现有定义保持不变 =============

message GetMenuSnapshotReq {
  string provider_name = 1; // 渠道名称: grab,lineman
  string shop_uuid = 2;     // 店铺 UUID
  string request_id = 3;    // 请求 ID,可选
}

message GetMenuSnapshotResp {
  string menu_data = 2;        // Provider 侧原始菜单 JSON
  int64 updated_at = 3;        // 快照更新时间
  string sync_state = 4;       // 同步状态  QUEUED/PROCESSING/SUCCESS/FAIL
}

message SaveMenuSnapshotReq {
  string provider_name = 1; // 渠道名称: grab,lineman
  string shop_uuid = 2;     // 店铺 UUID
  string menu_data = 3;     // 菜单数据 JSON 字符串
  string request_id = 4;    // 请求 ID
}

message SaveMenuSnapshotResp {
}

// ============= 新增定义 =============

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
  optional string available_status = 4;           // 可用状态: ✅ 已完成 - 已发布 v2.12
  optional int64 max_stock = 5;                   // 库存数量
  repeated AdvancedPricing advanced_pricings = 6; // 高级定价配置
  repeated Purchasability purchasabilities = 7;   // 购买能力配置
  string request_id = 8;                          // 请求 ID (可选)
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
  optional string available_status = 5;           // 可用状态: ✅ 已完成 - 已发布 v2.12
  optional bool is_free = 6;                      // 是否免费
  repeated AdvancedPricing advanced_pricings = 7; // 高级定价配置
  string request_id = 8;                          // 请求 ID (可选)
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

// 外卖菜单服务
service MenuService {
    // 获取菜单快照
    rpc GetMenuSnapshot (GetMenuSnapshotReq) returns (takeout.ApiResponse) {}
    // 保存ttpos 菜单快照数据
    rpc SaveMenuSnapshot (SaveMenuSnapshotReq) returns (takeout.ApiResponse) {}
    // 更新菜单项 (商品)
    rpc UpdateMenuItem (UpdateMenuItemReq) returns (takeout.ApiResponse) {}
    // 更新菜单修饰符
    rpc UpdateMenuModifier (UpdateMenuModifierReq) returns (takeout.ApiResponse) {}
}
```

### 2. Controller 实现

**文件**: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`

新增两个方法：

```go
// UpdateMenuItem 更新菜单项
func (c *Controller) UpdateMenuItem(ctx context.Context, req *api.UpdateMenuItemReq) (*takeout.ApiResponse, error) {
    // 参数校验
    if req.MerchantId == "" {
        return &takeout.ApiResponse{
            Code:    "4001",
            Message: "merchant_id 不能为空",
        }, nil
    }
    if req.ItemId == "" {
        return &takeout.ApiResponse{
            Code:    "4001",
            Message: "item_id 不能为空",
        }, nil
    }

    // 转换请求参数
    updateReq := &grabDto.UpdateMenuItemReq{
        MerchantID:      req.MerchantId,
        ItemID:          req.ItemId,
        AvailableStatus: req.GetAvailableStatus(),
    }
    if req.Price != nil {
        price := *req.Price
        updateReq.Price = &price
    }
    if req.MaxStock != nil {
        stock := *req.MaxStock
        updateReq.MaxStock = &stock
    }
    // 转换高级定价配置
    if len(req.AdvancedPricings) > 0 {
        for _, ap := range req.AdvancedPricings {
            updateReq.AdvancedPricings = append(updateReq.AdvancedPricings, grabDto.UpdateAdvancedPricingReq{
                Key:   ap.Key,
                Price: ap.Price,
            })
        }
    }
    // 转换购买能力配置
    if len(req.Purchasabilities) > 0 {
        for _, p := range req.Purchasabilities {
            updateReq.Purchasabilities = append(updateReq.Purchasabilities, grabDto.UpdatePurchasabilityReq{
                Key:         p.Key,
                Purchasable: p.Purchasable,
            })
        }
    }

    // 调用 Service 层
    result, err := service.GrabMenu().UpdateMenuItem(ctx, updateReq)
    if err != nil {
        g.Log().Errorf(ctx, "[Menu] UpdateMenuItem failed: %v", err)
        return &takeout.ApiResponse{
            Code:    "5001",
            Message: "更新菜单项失败: " + err.Error(),
        }, nil
    }

    // 转换响应
    resp := &api.UpdateMenuItemResp{
        Success:      result.Success,
        MerchantId:   result.MerchantID,
        RecordId:     result.RecordID,
        RecordType:   result.RecordType,
        ErrorCode:    result.ErrorCode,
        ErrorMessage: result.ErrorMessage,
    }

    dataAny, err := anypb.New(resp)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    "5001",
            Message: "序列化响应数据失败",
        }, nil
    }

    g.Log().Debugf(ctx, "[Menu] UpdateMenuItem success: %+v", resp)
    return &takeout.ApiResponse{
        Code:    "0",
        Message: "success",
        Data:    dataAny,
    }, nil
}

// UpdateMenuModifier 更新菜单修饰符
func (c *Controller) UpdateMenuModifier(ctx context.Context, req *api.UpdateMenuModifierReq) (*takeout.ApiResponse, error) {
    // 参数校验
    if req.MerchantId == "" {
        return &takeout.ApiResponse{
            Code:    "4001",
            Message: "merchant_id 不能为空",
        }, nil
    }
    if req.ModifierId == "" {
        return &takeout.ApiResponse{
            Code:    "4001",
            Message: "modifier_id 不能为空",
        }, nil
    }
    if req.ModifierName == "" {
        return &takeout.ApiResponse{
            Code:    "4001",
            Message: "modifier_name 不能为空",
        }, nil
    }

    // 转换请求参数
    updateReq := &grabDto.UpdateMenuModifierReq{
        MerchantID:      req.MerchantId,
        ModifierID:      req.ModifierId,
        ModifierName:    req.ModifierName,
        AvailableStatus: req.GetAvailableStatus(),
    }
    if req.Price != nil {
        price := *req.Price
        updateReq.Price = &price
    }
    if req.IsFree != nil {
        isFree := *req.IsFree
        updateReq.IsFree = &isFree
    }
    // 转换高级定价配置
    if len(req.AdvancedPricings) > 0 {
        for _, ap := range req.AdvancedPricings {
            updateReq.AdvancedPricings = append(updateReq.AdvancedPricings, grabDto.UpdateAdvancedPricingReq{
                Key:   ap.Key,
                Price: ap.Price,
            })
        }
    }

    // 调用 Service 层
    result, err := service.GrabMenu().UpdateMenuModifier(ctx, updateReq)
    if err != nil {
        g.Log().Errorf(ctx, "[Menu] UpdateMenuModifier failed: %v", err)
        return &takeout.ApiResponse{
            Code:    "5001",
            Message: "更新菜单修饰符失败: " + err.Error(),
        }, nil
    }

    // 转换响应
    resp := &api.UpdateMenuModifierResp{
        Success:      result.Success,
        MerchantId:   result.MerchantID,
        RecordId:     result.RecordID,
        RecordType:   result.RecordType,
        ErrorCode:    result.ErrorCode,
        ErrorMessage: result.ErrorMessage,
    }

    dataAny, err := anypb.New(resp)
    if err != nil {
        return &takeout.ApiResponse{
            Code:    "5001",
            Message: "序列化响应数据失败",
        }, nil
    }

    g.Log().Debugf(ctx, "[Menu] UpdateMenuModifier success: %+v", resp)
    return &takeout.ApiResponse{
        Code:    "0",
        Message: "success",
        Data:    dataAny,
    }, nil
}
```

### 3. 实现步骤

1. **更新 Proto 文件**
   - 在 `menu.proto` 中添加新的消息类型和 RPC 方法
   - 执行 `make proto` 生成 Go 代码

2. **实现 Controller 方法**
   - 在 `menu.go` 中添加 `UpdateMenuItem` 和 `UpdateMenuModifier` 方法
   - 添加必要的 import：`grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"`

3. **测试验证**
   - 编写单元测试
   - 使用 grpcurl 进行集成测试

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名 | 签名/日期 |
| ------------ | ---- | --------- |
| 技术负责人   |      |           |
| 开发代表     |      |           |
| 测试代表     |      |           |

### 评审结论

- [ ] ✅ **批准**：进入开发阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待填写]
```

**下一步行动**：

- [ ] 更新 Proto 文件
- [ ] 实现 Controller 方法
- [ ] 编写测试用例

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**维护者**: 开发组

