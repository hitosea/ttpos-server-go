# GrabFood 外卖平台对接 (v1.1.3) 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | User (AI Agent 代填)   |
| **日期**   | 2025-12-04   |
| **目标版本** | v2.10.x |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

目前 `ttpos-takeout` 模块主要支持配送服务（如 Skootar），尚未对接主流的外卖平台（如 GrabFood）。商家无法直接通过 POS 系统接收和管理 GrabFood 订单，导致需要人工手动录单，容易出错且效率低下。

### 业务价值

1.  **拓展销售渠道**：通过对接 GrabFood，帮助商户获取更多线上流量和订单。
2.  **自动化订单处理**：实现订单自动下发到 POS/KDS，减少人工干预，降低漏单/错单率。
3.  **统一管理**：在 POS 系统内统一管理堂食与外卖订单，简化操作流程。

### 目标用户

- [x] 收银员 (处理外卖接单/拒单)
- [x] 商户管理员 (管理菜单映射、查看报表)
- [x] 厨房人员 (通过 KDS 查看外卖制作单)

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-takeout` 模块中新增 `Grab` Provider 实现，对接 GrabFood API v1.1.3。采用 **Webhook + MQ** 的异步架构，将 GrabFood 的订单推送转换为内部标准化消息，通过 MQ 解耦业务逻辑，确保系统的稳定性与可扩展性，以便后续接入其他外卖平台（如 Foodpanda, Lineman）。

### 核心功能点

1.  **菜单同步 (Menu Sync)**
    - 将 POS 商品/套餐数据转换为 GrabFood 格式并推送。
    - 处理 Webhook 回调以确认同步状态。
2.  **订单处理 (Order Integration)**
    - **Webhook 接收**：实现 GrabFood `Submit Order` 等 Webhook 接口。
    - **MQ 解耦**：收到订单后，将原始数据转换为标准化的 `ThirdPartyOrderEvent` 并投递到 RocketMQ。
    - **订单落库**：消费 MQ 消息，在 ERP/POS 系统中创建对应的销售订单。
3.  **门店状态管理 (Store Status)**
    - 支持通过 POS 切换 Grab 门店的营业/暂停状态。

### 架构设计 (MQ 解耦)

为支持后续对接更多平台，采用统一的消息模型：

```go
// 伪代码：统一外卖订单事件结构
type ThirdPartyOrderEvent struct {
    Platform    string      // "grab", "foodpanda", ...
    Action      string      // "create", "cancel", "status_update"
    RawOrderID  string      // 平台原始订单号
    OrderData   interface{} // 标准化后的订单数据
    OccurredAt  int64
}
```

流程：
`Grab Webhook` -> `ttpos-takeout (Http Handler)` -> `RocketMQ (Topic: takeout_order)` -> `ttpos-erp (Consumer)` -> `Create Order`

### 影响范围

**涉及终端**：
- [x] POS 收银端 (接单操作、订单列表标识)
- [x] Shop 商家管理端 (授权配置、菜单映射)
- [x] KDS 厨显端 (显示外卖渠道标识)

**涉及模块**：
- [x] `ttpos-takeout`: 核心对接逻辑、Webhook 处理、MQ 生产。
- [x] `ttpos-erp`: MQ 消费、订单创建逻辑适配。
- [x] 数据模型: 完善 `related_order_no`, `related_order_type` 等字段的使用。

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**
- [ ] **中**
- [x] **高**：涉及第三方 API 签名验证、复杂的菜单结构映射、MQ 异步处理及分布式事务一致性考虑。

### 工作量预估

- **预计天数**: 7-10 天
- **预估 SP**: 13 SP (TBD)

### 风险识别

1.  **API 变更风险**：Grab API v1.1.3 若有非兼容性变更需及时跟进。
2.  **消息丢失/重复**：需确保 MQ 消息的可靠投递与幂等消费。
3.  **菜单映射复杂性**：Grab 的 Modifier Group 与 POS 的加料/套餐结构可能不完全一致，需设计合理的映射规则。

---

## 🔗 相关资源

### 参考需求

- [GrabFood API Documentation v1.1.3](https://developer.grab.com/docs/grabfood/api/v1-1-3/)

### 相关文档

- `ttpos-bmp/app/ttpos-takeout/README.MD`
- `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go` (已有 Provider 枚举)

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |

### 评审结论

- [ ] ✅ **批准**
- [ ] 🔄 **修改后批准**
- [ ] ❌ **拒绝**

---

**版本**: v1.0.0
**创建日期**: 2025-12-04


