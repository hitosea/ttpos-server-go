# LINE MAN PlaceOrder 订单接收功能 任务分解

> 本文档定义 LINE MAN PlaceOrder 订单接收功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 14  
**已完成**: 10  
**进行中**: -  
**完成率**: 71%

---

## Phase 1: Logic 层实现（核心业务逻辑）

### 任务说明

本阶段实现订单数据转换和保存的核心逻辑，参考 Grab 订单处理架构。

---

- [x] 1.1 创建 Logic 文件和基础结构

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 创建 LINE MAN 订单处理 Logic 层文件
  - Requirements: 2.1 (订单数据接收), 2.2 (数据模型转换)
  - Leverage: `internal/logic/grab/grab_order.go` - Grab 订单处理逻辑作为参考
  - Completed: 2026-01-12 ✅

- [x] 1.2 实现 HandlePlaceOrder 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 实现订单处理入口方法
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: `internal/logic/grab/grab_order.go` 的 `HandleSubmitOrder` 方法
  - Completed: 2026-01-12 ✅

- [x] 1.3 实现订单数据转换逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 将 LINE MAN 订单数据转换为 TTPOS 订单模型
  - Requirements: 2.2 (数据模型转换)
  - Leverage: `internal/logic/grab/grab_order.go` 的 `saveOrderFromSDK` 方法, [字段映射文档](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
  - Completed: 2026-01-12 ✅

- [x] 1.4 实现订单保存逻辑（事务）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 使用事务保存订单主表和明细表
  - Requirements: 2.3 (订单数据持久化)
  - Leverage: `internal/logic/grab/grab_order.go` 的事务处理代码
  - Completed: 2026-01-12 ✅

- [x] 1.5 实现门店配置查询

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 查询 storeId → shop_uuid 映射关系
  - Requirements: 2.2 (数据模型转换 - shop_uuid 查询)
  - Leverage: `internal/logic/grab/grab_order.go` 的门店配置查询代码
  - Completed: 2026-01-12 ✅

- [x] 1.6 实现订单 ID 去重逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go`
  - Purpose: 防止重复订单（HTTP 409）
  - Requirements: 2.3 (订单数据持久化 - 去重)
  - Leverage: 数据库唯一索引（provider_name + provider_order_id）
  - Completed: 2026-01-12 ✅

---

## Phase 2: Service 和 Controller 实现

### 任务说明

注册 Service 接口，完成 Controller 业务逻辑调用。

---

- [x] 2.1 创建 Service 接口定义

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman_order.go`
  - Purpose: 定义 LINE MAN 订单处理服务接口
  - Requirements: 所有功能需求
  - Leverage: `internal/service/grab.go` - Grab Service 接口定义
  - Completed: 2026-01-12 ✅

- [x] 2.2 注册 Logic 到 Service

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order.go` (init 函数)
  - Purpose: 将 Logic 层实现注册到 Service 接口
  - Requirements: 所有功能需求
  - Leverage: `internal/logic/grab/grab.go` - 现有 Service 注册代码
  - Completed: 2026-01-12 ✅

- [x] 2.3 实现 Controller 业务逻辑调用

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_place_order.go`
  - Purpose: 在 Controller 中调用 Logic 层处理订单
  - Requirements: 所有功能需求
  - Leverage: `internal/controller/grab/grab_v1_submit_order.go` - Grab Controller 实现
  - Completed: 2026-01-12 ✅

---

## Phase 3: 测试

### 任务说明

编写单元测试、API 测试和集成测试。

---

- [x] 3.1 编写订单数据转换单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go`
  - Purpose: 测试 LINE MAN → TTPOS 字段映射
  - Requirements: 2.2 (数据模型转换)
  - Leverage: `internal/logic/grab/grab_order_test.go` - Grab 订单测试用例
  - Completed: 2026-01-12 ✅ (基础测试已创建，集成测试需要测试环境)

- [ ] 3.2 编写订单保存单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_order_test.go`
  - Purpose: 测试订单保存逻辑和事务回滚
  - Requirements: 2.3 (订单数据持久化)
  - Leverage: `internal/logic/grab/grab_order_test.go`
  - Note: 需要测试环境（数据库连接、测试数据）

- [ ] 3.3 编写 API 集成测试（Postman / curl）

  - File: `docs/shared/api/lineman_place_order_test.md`（测试文档）
  - Purpose: 测试 PlaceOrder Webhook 接口
  - Requirements: 所有功能需求
  - Leverage: LINE MAN API 文档
  - Note: 需要测试环境和 LINE MAN 平台联调

- [ ] 3.4 编写端到端集成测试

  - File: `test/integration/lineman_order_test.go`
  - Purpose: 测试完整订单流程（Webhook → 数据库 → MQ → Main 模块）
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Note: 需要完整的测试环境

---

## Phase 4: 文档和部署

### 任务说明

更新相关文档，准备部署。

---

- [x] 4.1 更新 API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 补充 LINE MAN PlaceOrder 接口说明
  - Requirements: 文档验收标准
  - Completed: 2026-01-12 ✅ (已在 design.md 中详细说明)

- [ ] 4.2 更新 CHANGELOG

  - File: 项目根目录暂无 CHANGELOG.md，记录在 Spec 文档中
  - Purpose: 记录功能变更
  - Requirements: 文档验收标准
  - Note: 可在项目版本发布时统一更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码通过 `golangci-lint` 检查
- [ ] 测试覆盖率达标
  - Logic: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - [x] Requirement 1: 订单数据接收
  - [x] Requirement 2: 数据模型转换
  - [x] Requirement 3: 订单数据持久化
  - [x] Requirement 4: 消息队列通知
  - [x] Requirement 5: 错误处理和日志
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - [x] 订单保存成功，返回 HTTP 200
  - [x] MQ 事件发送成功
  - [x] 参数验证失败返回 HTTP 400
  - [x] 订单 ID 重复返回 HTTP 409
  - [x] 数据库失败返回 HTTP 500
  - [x] 能够区分 LINE MAN 和 Grab 订单

### 文档同步

- [ ] API 文档已更新（lineman_api.md）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-place-order/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-place-order/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-lineman-place-order/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-place-order/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-place-order/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码（Grab 订单处理）
4. **参考映射文档**: [Lineman API定义及TTPOS 映射](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165)
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame microservices

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
  - ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order.go - Grab 订单处理参考
  - https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=182890165#gid=182890165 - 字段映射文档
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 GoFrame ORM 进行数据库操作
- 使用 g.Log() 记录日志
- 使用 gerror.Wrap 包装错误
- Logic 层专注业务逻辑，不包含 HTTP 响应处理

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80%
- Reference: ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_order_test.go

Test Cases Required:
- 正常场景测试（订单保存成功）
- 异常场景测试（参数错误、数据库错误）
- 边界条件测试（订单 ID 重复、门店不存在）
- 数据转换测试（字段映射正确性）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试
- 使用测试数据库，测试后清理数据

Success Criteria:
- 测试覆盖率 ≥ 80%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-12.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-12  
**维护者**: rikugun
