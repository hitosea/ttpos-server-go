# ERPNext POS Invoice (Grab外卖) 需求文档

> Grab 外卖订单接单后同步到 ERPNext

## 📋 基本信息

| 项目 | 内容 |
|------|------|
| **来源 Proposal** | [v2.12.0-erpnext-pos-invoice-grab.md](../../../../team/proposals/2025-12/v2.12.0-erpnext-pos-invoice-grab.md) |
| **创建日期** | 2025-12-29 |
| **负责人** | weifashi |
| **目标 Sprint** | v2.12.0 |
| **涉及技术栈** | [x] Go (main/) ~~[x] Go (ttpos-bmp/)~~ |
| **关联任务** | DooTask #38169 |
| **说明** | ERPNext 配置和 BMP 模块已由其他同事完成，本需求只涉及 main 模块 |

## 📋 审核状态

| 项目 | 内容 |
|------|------|
| **审核状态** | 待审核 |
| **审核人** | 待指定 |

---

## 📋 概述

Grab 外卖订单在**商家接单后**，系统应同步到 ERPNext 创建 POS Invoice。

### 关键架构

**订单类型**：Grab 外卖平台订单（TakeoutOrder）
- 数据表：`ttpos_takeout_order`
- 特点：与 SaleBill **完全独立**（不创建 SaleBill）
- 触发事件：`OrderAcceptedEvent` (`"takeout.order.accepted"`)

**流程**：
```
顾客在 Grab App 下单并支付
  ↓
Grab 推送订单到 TTPOS（BMP 接收 webhook）
  ↓
创建 TakeoutOrder（独立订单系统）
  ↓
商家接单（触发 OrderAcceptedEvent）
  ↓
✨ 直接从 TakeoutOrder 同步到 ERP（本需求）
```

---

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 ERPNext POS Invoice 中查看 Grab 订单信息（订单来源、Grab 订单号、配送费）  
**以便于** 准确核对账单、追溯订单来源、进行财务分析

---

## 功能需求

### 1. 数据同步

#### 触发时机
- **事件**：`OrderAcceptedEvent` (`"takeout.order.accepted"`)
- **位置**：`main/app/event/takeout/takeout_order_accept_event_handler.go`

#### 数据映射

| TTPOS 字段 | 来源 | ERP 字段 | 说明 |
|-----------|------|---------|------|
| 订单来源 | `TakeoutOrder.Platform` = "grab" | `custom_order_source_name` | 查询 OrderSource 表获取名称 |
| 第三方订单号 | `TakeoutOrder.PlatformOrderId` | `custom_related_order_no` | Grab 订单ID |
| 订单类型 | `TakeoutOrder.Platform` | `custom_related_order_type` | "grab" |
| 配送费 | `TakeoutOrder.DeliveryFee` | Invoice Item | 单独商品项 |
| 订单商品 | `TakeoutOrderItem[]` | `items[]` | 商品列表 |
| 税费 | `TakeoutOrder.Tax` | `taxes[]` | 税费 |
| 订单金额 | `TakeoutOrder.EaterPayment` | `grand_total` | 实付金额 |

#### 同步条件
```go
company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != ""
```

---

### 2. 配送费处理

**规则**：
- `DeliveryFee > 0`：添加配送费商品项
- `DeliveryFee = 0`：不添加

**商品项结构**：
```json
{
  "item_code": "DELIVERY_FEE",
  "item_name": "配送费",
  "qty": 1,
  "rate": 25.00,  // TakeoutOrder.DeliveryFee
  "amount": 25.00
}
```

---

### 3. ERP 自定义字段（已完成）

✅ ERPNext `POS Invoice` 已添加以下自定义字段（其他同事已完成）：

| 字段名 | 类型 | 说明 |
|-------|------|------|
| `custom_order_source_name` | Data | 订单来源（"Grab"） |
| `custom_related_order_no` | Data | Grab 订单ID |
| `custom_related_order_type` | Data | "grab" |

✅ 配送费虚拟商品 `DELIVERY_FEE` 已创建

✅ BMP 模块的 Protobuf 和 SavePosInvoice 方法已扩展

---

## 验收标准

### AC1：接单后触发同步

```gherkin
Given Grab 订单待接单 + 公司已开启 ERP Phase 3
When 商家接单
Then 触发 OrderAcceptedEvent
  And 调用 ERP 同步逻辑
  And 从 TakeoutOrder 构建 POS Invoice
  And 调用 BMP gRPC 接口
```

### AC2：ERP 包含订单来源

```gherkin
Given Grab 订单已同步
When 查看 ERPNext POS Invoice
Then 包含 custom_order_source_name = "Grab"
  And 包含 custom_related_order_type = "grab"
```

### AC3：ERP 包含 Grab 订单号

```gherkin
Given PlatformOrderId = "T-bIpDLSJyY4DkuI"
When 查看 ERPNext POS Invoice
Then custom_related_order_no = "T-bIpDLSJyY4DkuI"
```

### AC4：ERP 包含配送费

```gherkin
Given DeliveryFee = 25.00
When 查看商品明细
Then 包含配送费商品项
  And 数量 = 1, 单价 = 25.00
```

### AC5：配送费为 0 时不创建商品项

```gherkin
Given DeliveryFee = 0
When 查看商品明细
Then 不包含配送费商品项
```

### AC6：同步失败不阻塞流程

```gherkin
Given ERP 服务异常
When 接单触发同步失败
Then 记录错误日志
  And 不阻塞出库/送厨/打印流程
```

### AC7：未开启 ERP 不同步

```gherkin
Given 公司未开启 ERP Phase 3
When 接单
Then 不触发 ERP 同步
  And 其他流程正常
```

---

## 技术约束

### 数据完整性
- 必须验证 TakeoutOrder 数据完整性
- 必须处理 DeliveryFee = 0 的情况
- 必须处理商品规格（Modifiers）

### 事务处理
- ERP 同步失败不应影响核心流程
- 异步处理（独立 goroutine）
- 记录详细日志

### 性能要求
- 不阻塞接单流程
- 超时时间：30 秒
- 失败不自动重试

### 兼容性
- 不影响店内订单 ERP 同步
- 不影响外送订单（MemberSaleOrder）ERP 同步
- 支持未来扩展其他外卖平台

---

## 数据字典

### TakeoutOrder 核心字段

| 字段 | 类型 | 说明 | 示例 |
|-----|------|------|------|
| `Platform` | string | 平台名称 | "grab" |
| `PlatformOrderId` | string | Grab 订单ID | "T-bIpDLSJyY4DkuI" |
| `ShortOrderNumber` | string | 短订单号 | "BG7J" |
| `DeliveryFee` | float64 | 配送费 | 25.00 |
| `EaterPayment` | float64 | 实付金额 | 125.50 |
| `Tax` | float64 | 税费 | 8.77 |
| `PaymentType` | string | 支付方式 | "ONLINE" |

---

## 相关文档

- **设计文档**：[design.md](./design.md)
- **任务拆解**：[tasks.md](./tasks.md)
- **提案文档**：[../../proposals/2025-12/v2.12.0-erpnext-pos-invoice-grab.md](../../../../team/proposals/2025-12/v2.12.0-erpnext-pos-invoice-grab.md)

---

**文档版本**: v2.0  
**最后更新**: 2025-12-29  
**维护者**: weifashi
