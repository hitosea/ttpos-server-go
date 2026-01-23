# Takeout 模块新增 Grab ListOrders 服务 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | task-takeout-grab-list-orders                                        |
| **来源 Proposal** | [takeout-grab-list-orders](../../../team/proposals/2026-01/takeout-grab-list-orders.md) |
| **创建日期**      | 2026-01-23                                                           |
| **负责人**        | rikugun                                                              |
| **目标版本**      | v2.15                                                                |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-23 |

---

## 📝 用户故事

**作为** 后端服务/定时任务
**我想** 通过 gRPC 接口调用 GrabFood ListOrders API 查询订单列表
**以便于** 同步订单数据、补捕漏单、支持订单对账和数据完整性保障

---

## 功能需求

### Requirement 1: Proto 定义 ListOrders RPC 方法

**用户故事**: 作为后端开发者，我想在 grab.proto 中定义 ListOrders 服务方法，以便于提供标准化的 gRPC 接口

#### 验收标准

1. **WHEN** 编译 proto 文件 **THEN** 系统 **SHALL** 成功生成 Go 代码，包含 ListOrders 方法签名
2. **WHEN** 查看生成的代码 **THEN** 系统 **SHALL** 包含 ListOrdersReq 和 ListOrdersResp 消息结构

---

### Requirement 2: 完整参数透传支持

**用户故事**: 作为调用方，我想使用灵活的查询参数，以便于按不同维度查询订单

#### 验收标准

1. **WHEN** 调用 ListOrders 并传入有效 merchant_id **THEN** 系统 **SHALL** 返回该商户的订单列表
2. **WHEN** 传入 date 参数（格式 YYYY-MM-DD）**THEN** 系统 **SHALL** 只返回指定日期的订单
3. **WHEN** 传入 order_ids 参数 **THEN** 系统 **SHALL** 只返回指定 ID 的订单
4. **WHEN** 传入 page 参数 **THEN** 系统 **SHALL** 返回对应页的订单数据
5. **IF** merchant_id 为空 **THEN** 系统 **SHALL** 返回参数错误

---

### Requirement 3: SDK 调用封装

**用户故事**: 作为后端服务，我想通过封装好的接口调用 Grab SDK，以便于统一管理 API 调用逻辑

#### 验收标准

1. **WHEN** 调用 ListOrders **THEN** 系统 **SHALL** 使用 grabfood-api-sdk-go 的 ListOrdersAPI 发起请求
2. **WHEN** SDK 调用成功 **THEN** 系统 **SHALL** 将 SDK 响应转换为 Proto 定义的响应格式
3. **WHEN** SDK 调用失败 **THEN** 系统 **SHALL** 返回统一错误格式并记录详细日志

---

### Requirement 4: 统一响应格式和分页支持

**用户故事**: 作为调用方，我想获取标准化的响应和分页信息，以便于处理查询结果

#### 验收标准

1. **WHEN** 查询成功 **THEN** 系统 **SHALL** 返回 takeout.ApiResponse 格式，data 字段包含订单列表
2. **WHEN** 返回 more=true **THEN** 调用方 **SHALL** 可通过增加 page 参数获取更多数据
3. **WHEN** 返回 more=false **THEN** 表示 **SHALL** 已返回所有匹配的订单

---

## 非功能需求

### 测试要求

- [ ] Logic 层单元测试覆盖率 ≥ 80%
- [ ] 包含 SDK Mock 测试用例

### 平台兼容性

- [x] ttpos-takeout gRPC 服务
- [x] 内部服务调用

### 日志要求

- [ ] 记录请求参数（merchant_id, date, page）
- [ ] 记录响应订单数量
- [ ] 错误时记录详细错误信息

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 依赖: github.com/grab/grabfood-api-sdk-go
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/ 规范
- 禁止修改自动生成文件（dao/, model/entity/, model/do/）

### 资源约束

- Story Point: 2-3

---

## 风险和缓解

### 风险 1: 测试环境不可用（重要）

**影响**: 高
**缓解措施**:
- Grab 平台限制，此 API **仅生产环境支持**
- 单元测试使用 Mock SDK
- 集成测试需使用生产凭据
- 上线前进行有限的生产环境验证

### 风险 2: Grab API 调用频率限制

**影响**: 中
**缓解措施**: 查阅 Grab 官方文档确认 Rate Limit，必要时在服务层增加限流逻辑

### 风险 3: 大量订单时的分页处理

**影响**: 低
**缓解措施**: 返回 more 字段告知调用方是否有更多数据，由调用方决定是否继续分页，避免服务端超时

---

## API 设计

### Proto 定义（待实现）

```protobuf
// 在 grab.proto 中新增

// 查询订单列表
rpc ListOrders (ListOrdersReq) returns (takeout.ApiResponse);

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
```

### 实现位置

| 文件 | 说明 |
|------|------|
| `manifest/protobuf/grab/grab.proto` | Proto 定义 |
| `internal/logic/grab/grab_order.go` | SDK 调用封装（新增 ListOrders 方法） |
| `internal/controller/rpc/grab.go` | gRPC Controller（自动生成后实现） |

---

## 相关资源

- 关联 Proposal: [takeout-grab-list-orders](../../../team/proposals/2026-01/takeout-grab-list-orders.md)
- GrabFood SDK: [github.com/grab/grabfood-api-sdk-go](https://github.com/grab/grabfood-api-sdk-go)
- 参考实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go`

---

**版本**: v1.0.0
**创建日期**: 2026-01-23
