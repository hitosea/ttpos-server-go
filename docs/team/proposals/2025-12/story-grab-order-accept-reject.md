# Grab 订单接受/拒绝功能

## 基本信息

- **提案人**: AI Assistant
- **日期**: 2025-12-22
- **版本**: v1.0.0
- **状态**: 待评审
- **优先级**: 中等

## 背景和动机

当前 TTPOS 外卖系统已支持 GrabFood 订单的接收和状态同步，但缺少对订单的主动管理能力。商户需要能够主动接受或拒绝 GrabFood 平台推送的订单，以更好地控制订单处理流程和提升运营效率。

## 解决方案概述

在现有的 GrabFood 集成基础上，新增订单接受/拒绝功能：

1. **新增 Prepare 服务**: 在 `order.proto` 中定义新的 gRPC 接口
2. **实现 Grab SDK 调用**: 在 `grab_order.go` 中集成 GrabFood API 的 accept-reject-order 功能
3. **多平台扩展**: 设计支持未来扩展到其他外卖平台（如 FoodPanda）的架构
4. **状态同步**: 确保本地状态与平台状态的一致性

## 核心功能点

### 1. Prepare 服务接口
- **接口名称**: `PrepareOrder`
- **请求参数**:
  - `takeout_order_uuid`: TTPOS 订单 UUID
  - `to_state`: 目标状态 (Accepted/Rejected)
- **响应**: 标准 API 响应格式

### 2. GrabFood 集成
- 调用 GrabFood API: `POST /grabfood/partner/v1/order/accept-reject`
- 支持接受和拒绝两种操作
- 包含必要的验证和错误处理

### 3. 多平台架构设计
- 根据 `provider_name` 字段路由到对应平台的处理逻辑
- 目前实现 GrabFood，后续可扩展 FoodPanda 等平台

### 4. 业务逻辑
- 订单状态验证（只能处理待确认状态的订单）
- 本地状态更新
- MQ 事件推送
- 完整的错误处理和日志记录

## 技术实现方案

### API 接口定义
```protobuf
// 新增 PrepareOrder 请求消息
message PrepareOrderReq {
  string takeout_order_uuid = 1;  // TTPOS 订单 UUID
  string to_state = 2;            // 目标状态: Accepted/Rejected
  string request_id = 3;          // 请求追踪ID (可选)
}

// 新增 PrepareOrder 响应消息
message PrepareOrderResp {
  string order_uuid = 1;     // 订单 UUID
}
```

### 核心实现逻辑
```go
// 在 grab_order.go 中新增方法
func (s *sGrabOrder) PrepareOrder(ctx context.Context, orderUUID string, toState string) error {
    // 1. 查询订单信息
    // 2. 验证订单状态
    // 3. 调用 GrabFood SDK 接受/拒绝订单
    // 4. 更新本地状态
    // 5. 发送 MQ 事件
}
```

### 集成方式
- 通过现有的 `OrderService` 暴露新接口
- 复用现有的数据库事务和错误处理逻辑
- 遵循现有的日志和监控规范

## 验收标准

### 功能验收
- [ ] 能够成功接受 GrabFood 订单
- [ ] 能够成功拒绝 GrabFood 订单
- [ ] 本地订单状态正确更新
- [ ] MQ 事件正确推送
- [ ] 错误情况正确处理和记录

### 集成验收
- [ ] gRPC 接口正常调用
- [ ] 与现有订单查询接口兼容
- [ ] 日志记录完整
- [ ] 性能满足要求（响应时间 < 2s）

### 异常处理验收
- [ ] 订单不存在时返回明确错误
- [ ] 订单状态不允许操作时返回错误
- [ ] SDK 调用失败时本地状态不更新
- [ ] 网络超时等异常情况正确处理

## 风险评估

### 技术风险
- **低**: 基于现有架构，复用 GrabFood SDK
- **低**: 遵循现有错误处理和日志规范

### 业务风险  
- **中**: 需要确保与 GrabFood 平台的接口兼容性
- **低**: 拒绝订单可能影响用户体验，需谨慎处理

### 运维风险
- **低**: 新增功能对现有系统影响小
- **低**: 完整的日志和监控覆盖

## 实施计划

### Phase 1: 设计与开发 (1-2 天)
- 设计 protobuf 接口
- 实现 GrabFood SDK 集成
- 编写单元测试

### Phase 2: 测试与验证 (1 天)  
- 功能测试
- 异常测试
- 集成测试

### Phase 3: 部署上线 (0.5 天)
- 代码审查
- 部署到测试环境
- 生产环境上线

## 关联任务

- 无关联 DooTask 任务

## 相关文档

- **Spec 需求文档**: [docs/shared/specs/active/story-grab-order-accept-reject/requirements.md](../../../../shared/specs/active/story-grab-order-accept-reject/requirements.md)

## 评审清单

### 产品评审
- [ ] 功能需求是否完整？
- [ ] 用户体验是否合理？
- [ ] 是否存在更好的解决方案？

### 技术评审  
- [ ] 架构设计是否合理？
- [ ] 代码实现是否符合规范？
- [ ] 性能和安全是否满足要求？

### 测试评审
- [ ] 测试用例是否覆盖所有场景？
- [ ] 异常处理是否完善？
- [ ] 集成测试是否通过？
