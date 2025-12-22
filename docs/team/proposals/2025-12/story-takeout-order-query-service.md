# 外卖订单信息查询服务 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-19   |
| **目标版本** | - |
| **状态**   | 已评审   |
| **关联任务** | - |
| **关联 Spec** | [docs/shared/specs/active/story-takeout-order-query-service/requirements.md](../../../shared/specs/active/story-takeout-order-query-service/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 ttpos-takeout 模块缺少订单信息查询接口，无法通过 gRPC 服务查询外卖订单的状态和原始数据。

业务方需要能够根据 TTPOS 订单 UUID (`orderUuid`) 快速查询订单的当前状态、类型以及原始平台数据，用于订单追踪、问题排查和状态同步。

### 业务价值

- 支持订单状态实时查询，提升运维效率
- 提供原始数据查询能力，便于问题排查
- 精简返回字段，降低网络传输开销
- 支持 requestId 跟踪，便于日志追踪和调试

### 目标用户

- [x] 商户管理员
- [x] 其他: TTPOS 内部系统（Main 模块）

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-takeout` 模块新增订单查询 gRPC 服务，支持通过 `shopUuid` + `orderUuid` 查询订单信息。

返回字段精简为业务必需的 4 个字段：`shopUuid`、`orderStatus`、`orderType`、`rawData`，避免返回冗余数据。

### 核心功能点

1. 新增 `GetOrderInfo` gRPC 接口
2. 输入参数：`shopUuid`、`orderUuid`、`requestId`（跟踪用）
3. 返回精简字段：`shopUuid`、`orderStatus`、`orderType`、`rawData`、`providerName`
4. 支持 requestId 透传，便于链路追踪

### 影响范围

**涉及终端**：
- [x] 其他: gRPC 微服务调用

**涉及模块**：
- [x] API 接口 (Protobuf 定义)
- [x] 业务逻辑 (Logic 层)

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：标准 CRUD 查询，无复杂业务逻辑

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. 查询无索引字段可能影响性能

**缓解措施**：
1. 确保 `shop_uuid` + `order_uuid` 组合索引存在

---

## 🔗 相关资源

### 参考文件

- Protobuf 定义: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
- Order Entity: `ttpos-bmp/app/ttpos-takeout/internal/model/entity/order.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 技术负责人   |        |           |
| 开发代表     |        |           |

### 评审结论

- [x] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
✅ 技术评审通过：标准 CRUD 查询，无复杂业务逻辑
✅ 工作量评估合理：0.5 天，SP=1
✅ 风险可控：确保索引存在即可

自动创建 Spec：story-takeout-order-query-service
```

**下一步行动**：

- [x] 创建 Spec：`story-takeout-order-query-service`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** TTPOS 内部系统（Main 模块）  
**我想** 通过 gRPC 接口查询外卖订单信息  
**以便于** 获取订单状态和原始数据进行业务处理

### AC 验收标准（初稿）

1. **WHEN** 调用 `GetOrderInfo` 接口传入有效的 `shopUuid` 和 `orderUuid` **THEN** 系统 **SHALL** 返回订单的 `shopUuid`、`orderStatus`、`orderType`、`rawData`、`providerName`
2. **WHEN** 查询不存在的订单 **THEN** 系统 **SHALL** 返回空结果或适当的错误码
3. **WHEN** 传入 `requestId` **THEN** 系统 **SHALL** 在日志中记录该 requestId 便于追踪

### 接口设计（初稿）

**Protobuf 定义**:

```protobuf
// 获取订单信息请求
message GetOrderInfoReq {
  string shop_uuid = 1;    // TTPOS店铺UUID
  string order_uuid = 2;  // TTPOS订单UUID
  string request_id = 3;   // 请求追踪ID
}

// 获取订单信息响应
message GetOrderInfoResp {
  string shop_uuid = 1;     // TTPOS店铺UUID
  string order_status = 2;  // 订单状态
  string order_type = 3;    // 订单类型
  string raw_data = 4;      // 原始JSON数据
  string provider_name = 5; // 渠道名称: grab, foodpanda
}

service OrderService {
  // 获取订单信息
  rpc GetOrderInfo(GetOrderInfoReq) returns (takeout.ApiResponse);
}
```

---

**版本**: v1.0.0
**创建日期**: 2025-12-19
**维护者**: rikugun
**最后更新**: 2025-12-22 (方案调整: shopRefNo → orderUuid)

