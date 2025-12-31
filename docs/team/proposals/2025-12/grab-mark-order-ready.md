# 实现 GrabFood Mark Order as Ready API 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-23   |
| **目标版本** | v2.12.0 |
| **状态**   | ✅ 已批准 - 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-bmp-grab-mark-order-ready](../../shared/specs/active/story-bmp-grab-mark-order-ready/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前 TTPOS 系统已经实现了 GrabFood 的订单接收和接受/拒绝功能，但缺少订单准备完成的通知机制。

根据 GrabFood API 文档要求，商户需要在订单准备完成后调用 "Mark order as ready" API 通知 Grab 平台，以便：
- 通知配送员可以开始取餐
- 更新订单状态为准备完成
- 优化配送时效和用户体验

当前系统无法主动通知 Grab 订单已准备完成，导致：
- 配送员无法及时知道订单状态
- 影响整体配送效率
- 用户体验不佳

### 业务价值

- 完善 GrabFood 订单流程闭环
- 提高配送效率和准确性
- 提升用户满意度
- 符合 GrabFood 平台对接规范

### 目标用户

- [x] 厨房人员（订单准备完成后）
- [x] POS 收银员（管理订单状态）
- [ ] 商户管理员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

在 ttpos-takeout 模块中实现 GrabFood 的 "Mark order as ready" API 集成：

1. **业务逻辑层**：在 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go` 中新增 `MarkOrderReady` 方法
   - 参考现有的 `PrepareOrder` 方法实现
   - 调用 GrabFood SDK 的 MarkOrderReady API
   - 处理成功/失败响应
   - 记录操作日志

2. **gRPC 接口层**：在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto` 中新增接口定义
   - 定义 `MarkOrderReadyReq` 请求消息
   - 定义 `MarkOrderReadyResp` 响应消息
   - 在 `OrderService` 中添加 `MarkOrderReady` RPC 方法

3. **集成方式**：
   - 使用已有的 GrabFood SDK (`github.com/grab/grabfood-api-sdk-go`)
   - 复用现有的配置和认证机制
   - 遵循现有的错误处理规范

### 核心功能点

1. **订单准备完成通知**
   - 接收 TTPOS 订单 UUID
   - 查询对应的 Grab 订单信息
   - 调用 GrabFood Mark Order as Ready API
   - 返回操作结果

2. **错误处理**
   - API 调用失败的重试机制
   - 完善的错误日志记录
   - 向调用方返回明确的错误信息

3. **状态同步**
   - 更新本地订单状态
   - 记录状态变更日志

### 影响范围

**涉及终端**：
- [x] POS 收银端（触发订单准备完成）
- [x] KDS 厨显端（触发订单准备完成）
- [ ] Shop 商家管理端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（gRPC）
- [ ] 数据模型
- [x] 业务逻辑（grab_order logic）
- [x] 第三方集成（GrabFood SDK）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：基于现有 SDK 和架构，实现逻辑清晰
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 2-3（待技术评审确认）

**工作项**：
1. 实现 grab_order.go 中的 MarkOrderReady 方法（0.5 天）
2. 定义 protobuf 接口并生成代码（0.5 天）
3. 实现 gRPC Controller 层（0.5 天）
4. 单元测试和联调测试（0.5 天）

### 风险识别

**潜在风险**：
1. GrabFood API 调用限制或超时
2. 订单状态流转不一致
3. SDK 版本兼容性问题

**缓解措施**：
1. 实现重试机制和超时处理
2. 完善的状态日志记录和校验
3. 参考现有 PrepareOrder 实现，使用相同的 SDK 版本

---

## 🔗 相关资源

### 参考需求

- GrabFood API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/mark-order-ready
- 类似功能实现: `PrepareOrder` 方法（接受/拒绝订单）

### 相关文档

- GrabFood SDK: https://github.com/grab/grabfood-api-sdk-go
- 现有实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
- Protobuf 规范: `.cursor/rules/proto-rules.mdc`
- Go BMP 规范: `.cursor/rules/go-bmp.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |  |           |
| 技术负责人   |  |           |
| 开发代表     | rikugun |           |
| 测试代表     |  |           |
| UI/UX 设计师 | - |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-bmp-grab-mark-order-ready`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** 厨房人员/收银员  
**我想** 在订单准备完成后通知 Grab 平台  
**以便于** 配送员能及时取餐，提高配送效率

### AC 验收标准（初稿）

1. **WHEN** 调用 MarkOrderReady API 且订单存在 **THEN** 系统 **SHALL** 成功调用 GrabFood API 并返回成功响应
2. **WHEN** 调用 MarkOrderReady API 且订单不存在 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **WHEN** GrabFood API 调用失败 **THEN** 系统 **SHALL** 记录详细的错误日志并返回错误响应
4. **IF** 订单已标记为 ready **THEN** 系统 **SHALL** 允许重复调用（幂等性）

### 技术实现参考

#### 1. grab_order.go 新增方法

```go
// MarkOrderReady 标记订单为准备完成
// 参数：
//   - ctx: 上下文对象
//   - orderEntity: 订单实体
//
// 返回：
//   - err: 错误信息
func (s *sGrabOrder) MarkOrderReady(ctx context.Context, orderEntityInterface interface{}) error {
    // 类型断言获取订单实体
    orderEntity, ok := orderEntityInterface.(*entity.Order)
    if !ok {
        return gerror.New("订单实体类型错误")
    }

    // 调用 GrabFood SDK 标记订单准备完成
    // markStatus 默认为 1
    err := service.Grab().MarkOrderReady(ctx, orderEntity.ProviderOrderId, 1)
    if err != nil {
        g.Log().Errorf(ctx, "标记订单准备完成失败: %v", err)
        return gerror.Wrap(err, "标记订单准备完成失败")
    }

    g.Log().Infof(ctx, "订单 %s 已标记为准备完成", orderEntity.Uuid)
    return nil
}
```

#### 2. order.proto 新增定义

```protobuf
// 标记订单准备完成请求
message MarkOrderReadyReq {
  string takeout_order_uuid = 1; // TTPOS订单UUID
  string request_id = 2;         // 请求追踪ID (可选)
}

// 标记订单准备完成响应
message MarkOrderReadyResp {
  string order_uuid = 1; // 订单UUID
}

// 在 OrderService 中添加
service OrderService {
  // ... 现有方法 ...
  
  // 标记订单准备完成 (markStatus 默认为 1)
  rpc MarkOrderReady(MarkOrderReadyReq) returns (takeout.ApiResponse);
}
```

### 线框图/原型（可选）

无需 UI 界面，仅为后端 API 实现。

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

