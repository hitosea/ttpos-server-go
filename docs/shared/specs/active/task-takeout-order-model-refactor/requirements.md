# TakeoutOrder 模型重构 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | task-takeout-order-model-refactor                                    |
| **Level**         | task                                                                 |
| **来源 Proposal** | [takeout-order-model-refactor](../../../team/proposals/2026-01/takeout-order-model-refactor.md) |
| **创建日期**      | 2026-01-23                                                           |
| **负责人**        | rikugun                                                              |
| **目标 Sprint**   | 待定                                                                 |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-23 |

---

## 📝 用户故事

**作为** 后端开发者/集成开发者
**我想** 让 TakeoutOrder 结构体与 Grab SDK 完全对齐，并通过 AdditionalProperties 支持其他平台扩展
**以便于** 降低维护成本、提升扩展性、减少转换错误、提高代码清晰度

---

## 功能需求

### Requirement 1: 重构 TakeoutOrder 主结构体

**用户故事**: 作为后端开发者，我想让 TakeoutOrder 字段与 Grab SDK Order 完全一致，以便于减少转换代码和维护成本

#### 验收标准

1. **WHEN** TakeoutOrder 结构体定义完成 **THEN** 系统 **SHALL** 包含与 Grab SDK Order 完全一致的字段（名称、类型、必选/可选）
2. **WHEN** 接收到 Grab 订单 **THEN** 系统 **SHALL** 能直接映射到 TakeoutOrder，无需额外类型转换
3. **WHEN** 需要支持平台特有字段 **THEN** 系统 **SHALL** 通过 AdditionalProperties map[string]interface{} 存储

### Requirement 2: 重构关联结构体

**用户故事**: 作为后端开发者，我想让所有关联结构体（OrderItem、OrderPrice 等）也对齐 Grab SDK，以便于保持一致性

#### 验收标准

1. **WHEN** OrderItem 结构体定义完成 **THEN** 系统 **SHALL** 字段对齐 Grab SDK OrderItem
2. **WHEN** OrderPrice 结构体定义完成 **THEN** 系统 **SHALL** 字段对齐 Grab SDK OrderPrice
3. **WHEN** 其他关联结构体定义完成 **THEN** 系统 **SHALL** 包括 Currency、FeatureFlags、Campaign、Promo、DineIn、Receiver、OrderReadyEstimation 等

### Requirement 3: 统一类型规范

**用户故事**: 作为后端开发者，我想统一字段类型规范，以便于减少类型转换错误

#### 验收标准

1. **WHEN** 价格字段定义 **THEN** 系统 **SHALL** 使用 int64（最小货币单位，如泰铢分）
2. **WHEN** 时间字段定义 **THEN** 系统 **SHALL** 使用 *time.Time 或 RFC3339 字符串（与 Grab SDK 一致）
3. **WHEN** 必需字段定义 **THEN** 系统 **SHALL** 使用非指针类型
4. **WHEN** 可选字段定义 **THEN** 系统 **SHALL** 使用指针类型

### Requirement 4: 保持 Lineman 兼容性

**用户故事**: 作为集成开发者，我想在重构后 Lineman 订单仍能正常处理，以便于不影响现有业务

#### 验收标准

1. **WHEN** 接收到 Lineman 订单 **THEN** 系统 **SHALL** 能正确转换到 TakeoutOrder 结构
2. **WHEN** Lineman 特有字段无法映射到 Grab 字段 **THEN** 系统 **SHALL** 存入 AdditionalProperties
3. **WHEN** 重构完成 **THEN** 系统 **SHALL** 现有 Grab 和 Lineman 订单处理流程正常工作

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 覆盖 Grab 订单转换场景
- [ ] 覆盖 Lineman 订单转换场景
- [ ] 边界值测试（空值、极值等）

### 兼容性要求

- [ ] Main 模块调用方正常工作
- [ ] BMP 模块调用方正常工作
- [ ] 回归测试通过

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 参考实现: `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order.go`
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- 结构体定义位置: `ttpos-api/ttpos-takeout/message/`

### 资源约束

- Story Point: 5（结构体重构 + 关联结构体 + 类型统一）

---

## 风险和缓解

### 风险 1: 兼容性风险

**影响**: 高
**描述**: 重构后可能影响现有订单处理逻辑
**缓解措施**:
- 编写完整的单元测试覆盖各平台订单转换场景
- 进行全面的回归测试

### 风险 2: 调用方适配

**影响**: 中
**描述**: Main/BMP 模块中使用旧字段的代码需要同步修改
**缓解措施**:
- 列出所有调用方并逐一检查
- 可考虑保留旧字段别名支持渐进式迁移

---

## 技术参考

### Grab SDK Order 结构体

**位置**: `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order.go`

**核心字段**:
```go
type Order struct {
    OrderID              string              `json:"orderID"`
    ShortOrderNumber     string              `json:"shortOrderNumber"`
    MerchantID           string              `json:"merchantID"`
    PartnerMerchantID    *string             `json:"partnerMerchantID,omitempty"`
    PaymentType          string              `json:"paymentType"`
    Cutlery              bool                `json:"cutlery"`
    OrderTime            string              `json:"orderTime"`
    SubmitTime           *time.Time          `json:"submitTime,omitempty"`
    CompleteTime         *time.Time          `json:"completeTime,omitempty"`
    ScheduledTime        *string             `json:"scheduledTime,omitempty"`
    OrderState           *string             `json:"orderState,omitempty"`
    Currency             Currency            `json:"currency"`
    FeatureFlags         OrderFeatureFlags   `json:"featureFlags"`
    Items                []OrderItem         `json:"items"`
    Campaigns            []OrderCampaign     `json:"campaigns,omitempty"`
    Promos               []OrderPromo        `json:"promos,omitempty"`
    Price                OrderPrice          `json:"price"`
    DineIn               NullableDineIn      `json:"dineIn,omitempty"`
    Receiver             NullableReceiver    `json:"receiver,omitempty"`
    OrderReadyEstimation *OrderReadyEstimation `json:"orderReadyEstimation,omitempty"`
    MembershipID         *string             `json:"membershipID,omitempty"`
    AdditionalProperties map[string]interface{}
}
```

### 当前 TakeoutOrder 位置

**位置**: `ttpos-api/ttpos-takeout/message/takeout_order.go`

---

**版本**: v1.0.0
**创建日期**: 2026-01-23
