# Takeout 模块新增 Grab ListOrders 服务 需求提案

## 📋 提案信息

| 项目          | 内容             |
| ------------- | ---------------- |
| **提案人**    | rikugun          |
| **日期**      | 2026-01-23       |
| **目标版本**  | v2.15            |
| **状态**      | 待评审           |
| **关联 Spec** | -                |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-takeout` 模块已集成 GrabFood 官方 SDK（`github.com/grab/grabfood-api-sdk-go`），实现了订单接收（Webhook）、订单状态更新、接受/拒绝订单、标记就绪、取消订单等核心功能。但**缺少主动查询订单列表的能力**。

现有问题：
1. **无法主动拉取订单**：只能被动通过 Webhook 接收订单，如果 Webhook 漏单或系统重启期间有订单，无法补捕
2. **无法对账**：商户无法核对本地订单数据与 Grab 平台的一致性
3. **报表数据不完整**：缺少从 Grab 平台直接获取订单数据的能力，影响订单分析和统计

### 业务价值

- **数据完整性保障**：支持订单补捕和同步，确保本地数据与 Grab 平台一致
- **订单对账支持**：商户可核对订单金额与结算数据
- **报表数据补全**：为运营报表提供完整的 Grab 订单数据源

### 目标用户

- [x] 后端服务（定时任务、数据同步服务）
- [x] 商户管理员（通过后台系统查询）
- [x] 运营人员（订单分析和报表）

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-takeout` 模块的 `grab.proto` 中新增 `ListOrders` gRPC 服务方法，封装 GrabFood SDK 的 `ListOrdersAPI`。采用**完整参数透传**模式，将 SDK 支持的所有查询参数（merchantID、date、orderIDs、page）暴露给调用方，提供灵活的查询能力。

参考现有 `grab_order.go` 中的实现模式：
- 使用 `service.Grab()` 获取 Grab 服务实例
- 通过 SDK Client 调用 API
- 统一返回 `takeout.ApiResponse` 格式

### 核心功能点

1. **Proto 定义**：在 `grab.proto` 中新增 `ListOrders` RPC 方法和请求/响应消息
2. **参数透传**：支持 merchantID（必填）、date（可选）、orderIDs（可选）、page（可选）
3. **SDK 调用**：在 `internal/logic/grab/` 中实现对 `grabfood.ListOrdersAPI` 的调用封装
4. **响应格式**：统一返回 `takeout.ApiResponse`，data 字段包含订单列表和分页信息

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
- [x] 内部服务（gRPC 调用）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（新增 gRPC 方法）
- [x] 数据模型（新增 Proto 消息定义）
- [x] 业务逻辑（SDK 调用封装）
- [x] 第三方集成（GrabFood SDK）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整,无业务逻辑变更
- [x] **中**：需要前后端联调,基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预估 SP**: 2-3（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [x] **否**：单模块，SP ≤ 5，可直接创建 1 个 Spec
- [ ] **是**：需要拆分为多个 Spec

**预估 Spec 数量**：1 个

**预估 Spec 列表**：
1. `task-takeout-grab-list-orders` - 实现 Grab ListOrders gRPC 服务

### 风险识别

**潜在风险**：
1. **API 调用频率限制**：Grab API 可能有调用频率限制，需确认 Rate Limit 策略
2. **分页数据量**：单次返回订单数量未知，大量订单时需考虑分页循环策略

**缓解措施**：
1. 查阅 Grab 官方文档确认 Rate Limit，必要时在服务层增加限流逻辑
2. 返回 `more` 字段告知调用方是否有更多数据，由调用方决定是否继续分页

---

## 🔗 相关资源

### 参考文档

- GrabFood SDK: [github.com/grab/grabfood-api-sdk-go](https://github.com/grab/grabfood-api-sdk-go)
- SDK ListOrders API: [pkg.go.dev/github.com/grab/grabfood-api-sdk-go](https://pkg.go.dev/github.com/grab/grabfood-api-sdk-go)
- Grab API 文档: [developer.grab.com](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/list-orders/operation/list-orders)

### 相关代码

- Proto 定义: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
- 订单逻辑参考: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go`
- 关联提案: `docs/team/proposals/2025-12/takeout-grab-sdk-integration.md`

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名    | 签名/日期 |
| ---------- | ------- | --------- |
| 产品经理   |         |           |
| 技术负责人 |         |           |
| 开发代表   | rikugun |           |
| 测试代表   |         |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审]
```

**下一步行动**：

- [ ] 创建 Spec：`task-takeout-grab-list-orders`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### API 设计（初稿）

#### Proto 定义

```protobuf
// 新增 ListOrders RPC 方法
service Grab {
  // ... 现有方法 ...

  // 查询订单列表
  // 参数：merchant_id（必填）、date（可选）、order_ids（可选）、page（可选）
  // 返回：统一的 ApiResponse 格式，包含订单列表和分页信息
  rpc ListOrders (ListOrdersReq) returns (takeout.ApiResponse);
}

// 查询订单列表请求消息
message ListOrdersReq {
  string merchant_id = 1;         // Grab 商户 ID（必填）
  string date = 2;                // 日期过滤，格式 YYYY-MM-DD（可选）
  repeated string order_ids = 3;  // 订单 ID 列表过滤（可选）
  int32 page = 4;                 // 分页页码（可选，默认 1）
}

// 查询订单列表响应消息
message ListOrdersResp {
  repeated GrabOrder orders = 1;  // 订单列表
  bool more = 2;                  // 是否有更多数据
}

// Grab 订单信息（简化版，完整字段参考 SDK Order 结构）
message GrabOrder {
  string order_id = 1;            // 订单 ID
  string merchant_id = 2;         // 商户 ID
  string order_state = 3;         // 订单状态
  string order_time = 4;          // 下单时间
  string short_order_number = 5;  // 短订单号
  // ... 其他必要字段
}
```

### User Story（初稿）

**作为** 后端服务/定时任务
**我想** 调用 ListOrders gRPC 接口查询 Grab 订单列表
**以便于** 同步订单数据、补捕漏单、支持订单对账和报表分析

### AC 验收标准（初稿）

1. **WHEN** 调用 `ListOrders` 并传入有效 merchant_id **THEN** 系统 **SHALL** 返回该商户的订单列表
2. **WHEN** 传入 date 参数 **THEN** 系统 **SHALL** 只返回指定日期的订单
3. **WHEN** 传入 order_ids 参数 **THEN** 系统 **SHALL** 只返回指定 ID 的订单
4. **WHEN** 返回 more=true **THEN** 调用方 **SHALL** 可通过增加 page 参数获取更多数据
5. **WHEN** Grab API 返回错误 **THEN** 系统 **SHALL** 返回统一错误格式并记录日志

---

**版本**: v1.0.0
**创建日期**: 2026-01-23
**维护者**: rikugun
