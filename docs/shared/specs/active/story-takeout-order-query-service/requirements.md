# 外卖订单信息查询服务 需求规格

> **状态**: 已审核  
> **版本**: v1.0.0  
> **创建日期**: 2025-12-22  
> **更新日期**: 2025-12-22

---

## 📋 基本信息

| 项目       | 内容                              |
| ---------- | --------------------------------- |
| **Spec 名称** | story-takeout-order-query-service |
| **模块**   | ttpos-takeout (BMP)             |
| **类型**   | story (用户故事)                 |
| **提案人** | rikugun                          |
| **负责人** | 待定                             |
| **目标版本** | 待定                             |
| **技术栈** | Go + GoFrame + gRPC + MySQL     |
| **复杂度** | 低                               |
| **预估工作量** | 0.5 天 (SP: 1)                  |

**来源 Proposal**: [docs/team/proposals/2025-12/story-takeout-order-query-service.md](../../../team/proposals/2025-12/story-takeout-order-query-service.md)

---

## 🎯 需求描述

### 问题背景

当前 `ttpos-takeout` 模块缺少订单信息查询接口，无法通过 gRPC 服务查询外卖订单的状态和原始数据。

业务方需要能够根据 TTPOS 内部订单号 (`shopRefNo`) 快速查询订单的当前状态、类型以及原始平台数据，用于订单追踪、问题排查和状态同步。

### 业务价值

- 支持订单状态实时查询，提升运维效率
- 提供原始数据查询能力，便于问题排查
- 精简返回字段，降低网络传输开销
- 支持 requestId 跟踪，便于日志追踪和调试

### 目标用户

- [x] 商户管理员
- [x] 其他: TTPOS 内部系统（Main 模块）

---

## 💡 功能需求

### 核心功能点

1. **新增 `GetOrderInfo` gRPC 接口**
   - 支持通过 `shopUuid` + `orderUuid` 查询订单信息
   - 返回精简字段：`shopUuid`、`orderStatus`、`orderType`、`rawData`

2. **输入参数设计**
   - `shopUuid`: TTPOS 店铺 UUID
   - `orderUuid`: TTPOS 订单 UUID
   - `requestId`: 请求追踪 ID（可选，用于日志追踪）

3. **输出结果设计**
   - `shopUuid`: TTPOS 店铺 UUID
   - `orderStatus`: 订单状态
   - `orderType`: 订单类型
   - `rawData`: 原始 JSON 数据
   - `providerName`: 渠道名称 (grab, foodpanda)

4. **错误处理**
   - 查询不存在的订单时返回空结果或适当错误码
   - 支持 requestId 透传，便于链路追踪

### 非功能需求

- **性能要求**: 查询响应时间 < 100ms
- **可用性要求**: 99.9% SLA
- **安全性要求**: 通过 gRPC 内部调用，无需额外安全措施

---

## 🔍 验收标准

### 功能验收标准

1. **WHEN** 调用 `GetOrderInfo` 接口传入有效的 `shopUuid` 和 `orderUuid`
   **THEN** 系统 **SHALL** 返回订单的 `shopUuid`、`orderStatus`、`orderType`、`rawData`、`providerName`

2. **WHEN** 查询不存在的订单
   **THEN** 系统 **SHALL** 返回空结果或适当的错误码

3. **WHEN** 传入 `requestId`
   **THEN** 系统 **SHALL** 在日志中记录该 requestId 便于追踪

### 接口验收标准

- [ ] Protobuf 定义正确
- [ ] gRPC 服务实现完成
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过

---

## 📊 技术方案

### 技术栈选择

- **框架**: GoFrame v2.x
- **协议**: gRPC + Protobuf
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+ (如需要)

### 接口设计

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

### 数据库设计

**涉及表**: `ttpos_takeout.order`

**查询条件**:
- `shop_uuid` + `order_uuid` (需确保组合索引存在)

**返回格式**:
- 使用 `takeout.ApiResponse` 统一响应格式
- `ApiResponse.code`: 响应码（"0" 表示成功）
- `ApiResponse.message`: 响应消息
- `ApiResponse.data`: 订单信息（`GetOrderInfoResp` 类型）

**订单信息字段**（在 `ApiResponse.data` 中）:
- `shop_uuid`
- `order_status`
- `order_type`
- `raw_data` (JSON 格式)
- `provider_name` (渠道名称)

### 影响范围

**涉及终端**:
- [x] 其他: gRPC 微服务调用

**涉及模块**:
- [x] API 接口 (Protobuf 定义)
- [x] 业务逻辑 (Logic 层)
- [x] 数据访问 (DAO 层)

---

## 🧪 测试策略

### 单元测试

- Logic 层查询逻辑测试
- DAO 层数据库操作测试
- 参数验证测试

### 集成测试

- gRPC 接口调用测试
- 数据库查询性能测试
- 错误场景测试

### 验收测试

- 功能验收测试 (UAT)
- 性能验收测试

---

## 📈 风险评估

### 技术风险

| 风险项 | 概率 | 影响 | 缓解措施 |
| ------ | ---- | ---- | -------- |
| 查询无索引影响性能 | 中 | 中 | 确保 `shop_uuid` + `shop_ref_no` 组合索引存在 |
| 数据量大导致查询慢 | 低 | 低 | 考虑添加缓存策略 |

### 业务风险

| 风险项 | 概率 | 影响 | 缓解措施 |
| ------ | ---- | ---- | -------- |
| 返回字段不满足业务需求 | 低 | 中 | 产品审核时确认字段需求 |

---

## 📋 审核清单

### 产品审核

- [ ] 需求描述清晰准确
- [ ] 验收标准明确可验证
- [ ] 业务价值合理
- [ ] 用户场景覆盖完整

### 技术审核

- [ ] 接口设计合理
- [ ] 数据库设计优化
- [ ] 性能影响评估
- [ ] 安全风险评估

---

## 🔄 审核状态

**当前状态**: ✅ 已通过

### 审核历史

| 日期 | 审核类型 | 审核人 | 结果 | 意见 |
| ---- | -------- | ------ | ---- | ---- |
| 2025-12-22 | 创建 | 系统 | 待审核 | 基于 Proposal 自动创建 |
| 2025-12-22 | 审核 | 系统 | 已通过 | 自动审核通过，准备进入设计阶段 |

### 下一步行动

1. **产品审核**: 确认需求和验收标准
2. **技术审核**: 确认技术方案和实现细节
3. **审核通过后**: ✅ 已执行 `/spec-design` 创建技术设计文档

---

**维护者**: rikugun  
**最后更新**: 2025-12-22 (方案调整: shopRefNo → orderUuid)