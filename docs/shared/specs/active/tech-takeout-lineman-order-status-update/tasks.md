# Lineman 订单状态更新 Webhook 任务分解

> 本文档定义 LINE MAN 订单状态更新 Webhook 的详细执行任务清单。

## 📊 进度总览

**总任务数**: 18  
**已完成**: 11 (Phase 1-3 全部完成 + CHANGELOG + API 文档)  
**剩余待完成**: 7 (单元测试、集成测试、Troubleshooting)  
**完成率**: 61%

---

## Phase 1: 常量和接口定义

- [x] 1.1 添加订单状态更新常量
  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - Purpose: 定义 `OrderActionStatusUpdate` 和 LINE MAN 状态常量
  - Requirements: Requirement 2
  - Leverage: 现有常量 `OrderActionCreate`, `OrderStatusCompleted`
  - ✅ 常量已存在于 consts.go

- [x] 1.2 添加 Service 接口方法
  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman.go`
  - Purpose: 在 `ILinemanOrder` 中添加 `HandleOrderStatusUpdate` 方法
  - Requirements: Requirement 1
  - Leverage: 现有方法 `HandlePlaceOrder`, `HandleOrderUpdate`
  - ✅ Service 接口已添加

---

## Phase 2: Logic 层实现

- [x] 2.1 实现状态映射函数
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 实现 `mapLinemanStatusToTTPOS`（FINISH → COMPLETED）
  - Requirements: Requirement 2
  - Leverage: Task 1.1 的常量定义
  - ✅ 状态映射函数已实现

- [x] 2.2 实现订单查询逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 根据 `provider_order_id` 查询订单
  - Requirements: Requirement 3
  - Leverage: `HandlePlaceOrder` 中的订单查询逻辑
  - ✅ 订单查询已实现

- [x] 2.3 实现幂等性检查逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 检查当前状态是否与目标状态相同
  - Requirements: Requirement 3
  - Leverage: `HandleOrderUpdate` 中的幂等性检查
  - ✅ 幂等性检查已实现

- [x] 2.4 实现订单状态更新逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 更新 `order_status` 和 `updated_at` 字段
  - Requirements: Requirement 3
  - Leverage: `HandleOrderUpdate` 中的订单更新逻辑
  - ✅ 订单状态更新已实现

- [x] 2.5 实现 RocketMQ 事件发送逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 构造 `OrderEvent` 并发送到 RocketMQ
  - Requirements: Requirement 4
  - Leverage: `HandlePlaceOrder` 中的 MQ 发送逻辑
  - ✅ RocketMQ 事件发送已实现

- [x] 2.6 实现完整的 HandleOrderStatusUpdate 方法
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 整合 Task 2.1-2.5 的逻辑
  - Requirements: Requirement 1-4
  - Leverage: Task 2.1-2.5 的实现
  - ✅ 完整方法已实现并通过静态检查

---

## Phase 3: Controller 层实现

- [x] 3.1 实现 Controller 的 OrderStatusUpdate 方法
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_order_status_update.go`
  - Purpose: 调用 Service 层，封装响应格式
  - Requirements: Requirement 1, 5
  - Leverage: `lineman_v1_order_update.go` 中的 `OrderUpdate` 方法
  - ✅ Controller 已实现并通过静态检查

---

## Phase 4: 测试和文档

- [ ] 4.1 编写状态映射函数单元测试
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go`
  - Purpose: 测试 FINISH → COMPLETED, CANCELED → CANCELED
  - Requirements: Requirement 2
  - Test Cases: 
    - `TestMapLinemanStatusToTTPOS_Finish`
    - `TestMapLinemanStatusToTTPOS_Canceled`
    - `TestMapLinemanStatusToTTPOS_Unknown`

- [ ] 4.2 编写 HandleOrderStatusUpdate 单元测试
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go`
  - Purpose: 测试订单状态更新逻辑
  - Requirements: Requirement 3, 4
  - Test Cases:
    - `TestHandleOrderStatusUpdate_Success`
    - `TestHandleOrderStatusUpdate_OrderNotFound`
    - `TestHandleOrderStatusUpdate_Idempotent`
  - Success: 测试覆盖率 ≥ 70%

- [ ] 4.3 集成测试（手动测试）
  - Purpose: 使用 Postman 测试完整流程
  - Requirements: 所有功能需求
  - Test Cases:
    - TC-1: 订单完成（FINISH → COMPLETED）
    - TC-2: 订单取消（CANCELED → CANCELED）
    - TC-3: 订单不存在（返回 404）
    - TC-4: 重复请求（幂等性）

- [ ] 4.4 验证 RocketMQ 消息发送
  - Purpose: 确保 Main 模块接收到事件
  - Requirements: Requirement 4
  - Verification: 检查 RocketMQ 控制台和 Main 模块日志

- [x] 4.5 更新 API 文档
  - File: `docs/shared/integrations/lineman/lineman-webhook-api.md`
  - Purpose: 添加订单状态更新 Webhook 说明
  - Content: 端点、参数、响应格式、状态映射、错误码
  - ✅ API 文档已创建（包含完整的 Webhook API 规范）

- [x] 4.6 更新 CHANGELOG
  - File: `ttpos-bmp/CHANGELOG.md`
  - Purpose: 记录功能新增
  - Content: 
    ```
    ### Added
    - LINE MAN 订单状态更新 Webhook
    ```
  - ✅ CHANGELOG 已更新

- [ ] 4.7 创建 Troubleshooting 文档（可选）
  - File: `docs/shared/troubleshooting/lineman-order-status-update.md`
  - Purpose: 故障排查指南
  - Content: 常见问题和解决方案

---

## 提交清单

### 代码质量
- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（Logic 层 ≥ 70%）
- [ ] 所有测试通过

### 功能完整性
- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步
- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪命令

```bash
# 查看总任务数
grep -c "^- \[" tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" tasks.md) * 100 / $(grep -c "^- \[" tasks.md)" | bc
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-13.md`

---

**版本**: v1.0.0  
**创建日期**: 2026-01-13  
**维护者**: rikugun
