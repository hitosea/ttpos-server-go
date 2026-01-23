# TakeoutOrder 模型重构 需求提案

## 📋 提案信息

| 项目          | 内容                          |
| ------------- | ----------------------------- |
| **提案人**    | rikugun                       |
| **日期**      | 2026-01-23                    |
| **目标版本**  | 待定                          |
| **状态**      | 待评审                        |
| **关联 Spec** | -                             |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-api/ttpos-takeout/message/takeout_order.go` 中的 `TakeoutOrder` 结构体存在以下问题：

1. **与 Grab SDK 不一致**：字段名、类型与官方 `grabfood-api-sdk-go` 的 `Order` 结构体存在差异，导致转换代码复杂且容易出错
2. **Lineman 字段混杂**：Grab 和 Lineman 平台特有字段混在一起，结构不清晰，难以维护
3. **类型不统一**：如价格字段使用 `float64` 而 Grab SDK 使用 `int64`（最小货币单位），增加了转换复杂度
4. **扩展性差**：新增外卖平台时需要修改核心结构体，违反开闭原则

### 业务价值

- **降低维护成本**：与 Grab SDK 字段一一对应，减少转换代码和维护工作量
- **提升扩展性**：新增平台时无需修改核心结构，只需实现转换逻辑
- **减少 Bug**：类型统一后减少数据转换错误
- **代码清晰度**：Grab 核心字段 vs 扩展字段分层清晰，提升可读性

### 目标用户

- [x] 后端开发者（维护外卖订单处理逻辑）
- [x] 集成开发者（对接新外卖平台）

---

## 💡 解决方案概述

### 方案描述

采用「**完全复制 Grab SDK + 扩展字段**」的重构方案：

1. `TakeoutOrder` 结构体字段完全对齐 Grab SDK 的 `Order` 结构体（字段名、类型、必选/可选）
2. Lineman 等其他平台无法映射到 Grab 字段的数据，统一放入 `AdditionalProperties map[string]interface{}`
3. 同步重构所有关联结构体：`TakeoutOrderItem`、`TakeoutOrderPrice`、`TakeoutModifier` 等

### 核心功能点

1. **重构 TakeoutOrder 主结构体**：完全对齐 Grab SDK `Order` 的字段定义（类型、可选性、JSON tag）
2. **重构关联结构体**：同步重构 `OrderItem`、`OrderPrice`、`Currency`、`FeatureFlags`、`Campaign`、`Promo`、`DineIn`、`Receiver`、`OrderReadyEstimation` 等
3. **统一类型规范**：价格类型对齐 Grab（int64 最小货币单位），时间类型使用 `time.Time` 或 RFC3339 字符串
4. **扩展字段支持**：通过 `AdditionalProperties` 支持 Lineman 等平台特有字段，无需修改核心结构
5. **更新转换逻辑**：重构 Grab/Lineman 订单到 TakeoutOrder 的转换函数

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [ ] Kiosk 自助点餐机
- [x] 无终端（后端共享层重构）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（ttpos-api 共享层）
- [x] 数据模型（TakeoutOrder 及关联结构体）
- [x] 业务逻辑（Grab/Lineman 订单转换逻辑）
- [x] 其他：Main 模块和 BMP 模块的调用方

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整,无业务逻辑变更
- [ ] **中**：需要前后端联调,基础业务逻辑
- [x] **高**：涉及架构调整、需要同步修改多个模块的调用代码

### 工作量预估

- **预估 SP**: 8（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [ ] **否**：单终端，SP ≤ 5，可直接创建 1 个 Spec
- [x] **是**：需要拆分为多个 Spec

**拆分维度**（如需拆分）：
- [ ] 按终端拆分
- [x] 按复杂度拆分：预估 SP > 5
- [x] 按功能模块拆分：结构体定义、转换逻辑、调用方适配
- [ ] 按 Phase 拆分
- [ ] 其他

**预估 Spec 数量**：2-3 个

**预估 Spec 列表**：
1. `story-takeout-order-model-refactor` - 重构 TakeoutOrder 及关联结构体定义
2. `story-takeout-order-converter-refactor` - 重构 Grab/Lineman 订单转换逻辑
3. `story-takeout-order-caller-adapt` - 更新 Main/BMP 模块调用方代码（可选，可合并到前两个）

### 风险识别

**潜在风险**：
1. **兼容性风险**：重构后可能影响现有订单处理逻辑
2. **回归测试**：需要全面测试 Grab 和 Lineman 订单流程

**缓解措施**：
1. 保留旧结构体别名或提供迁移工具，支持渐进式迁移
2. 编写完整的单元测试覆盖各平台订单转换场景

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名   | 签名/日期 |
| ---------- | ------ | --------- |
| 产品经理   |        |           |
| 技术负责人 |        |           |
| 开发代表   |        |           |
| 测试代表   |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-order-model-refactor`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 后端开发者
**我想** 让 TakeoutOrder 结构体与 Grab SDK 完全对齐，并通过 AdditionalProperties 支持其他平台扩展
**以便于** 降低维护成本、减少转换错误、提升新平台集成效率

### AC 验收标准（初稿）

1. **WHEN** 接收到 Grab 订单 **THEN** 系统 **SHALL** 直接映射到 TakeoutOrder，无需类型转换
2. **WHEN** 接收到 Lineman 订单 **THEN** 系统 **SHALL** 将公共字段映射到 TakeoutOrder，平台特有字段放入 AdditionalProperties
3. **WHEN** 新增外卖平台 **THEN** 开发者 **SHALL** 仅需实现转换函数，无需修改 TakeoutOrder 核心结构

### 技术参考

**Grab SDK Order 结构体位置**：
`github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order.go`

**当前 TakeoutOrder 位置**：
`ttpos-api/ttpos-takeout/message/takeout_order.go`

**关键字段对比**：

| 字段           | Grab SDK                | 当前 TakeoutOrder           | 重构后          |
| -------------- | ----------------------- | --------------------------- | --------------- |
| Cutlery        | `bool`                  | `*bool`                     | `bool`          |
| SubmitTime     | `*time.Time`            | `*string`                   | `*time.Time`    |
| Currency       | `Currency` (必需)       | `*TakeoutCurrency` (可选)   | `Currency`      |
| FeatureFlags   | `OrderFeatureFlags`     | `*TakeoutFeatureFlags`      | `OrderFeatureFlags` |
| Price.Subtotal | `int64` (最小货币单位)  | `float64`                   | `int64`         |

---

**版本**: v1.0.0
