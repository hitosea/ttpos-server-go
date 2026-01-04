# 叫号系统-订单已完成自动消失（时间）任务分解

> 本文档定义 叫号系统-订单已完成自动消失（时间） 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 5  
**已完成**: 4  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: Service 层实现

### 修改 GetQueueData 方法

- [x] 1.1 修改 `GetQueueData` 方法，传递 `timeout_limit` 参数

  - File: `main/app/service/callboard/service.go`
  - Purpose: 在获取 `PreparedQueue` 时传递 `timeout_limit` 配置，实现订单过滤
  - Requirements: 1.1, 1.2, 2.1, 2.2
  - Leverage: 现有代码: `main/app/service/callboard/service.go:210-265`
  - Status: ✅ 已完成 - 已修改 `GetQueueData` 方法，从 `bindInfo.TimeoutLimit` 获取超时时间配置，处理 nil 情况（默认视为 0），调用 `getCallBoardQueue` 时传递 `timeout_limit` 参数。`PreparingQueue` 传递 `timeoutLimit = 0`（不过滤），`PreparedQueue` 传递实际的 `timeout_limit` 值

- [x] 1.2 修改 `getCallBoardQueue` 方法，增加 `timeoutLimit` 参数并实现过滤逻辑

  - File: `main/app/service/callboard/service.go`
  - Purpose: 在 `getCallBoardQueue` 方法中实现根据 `timeout_limit` 过滤订单的逻辑
  - Requirements: 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有代码: `main/app/service/callboard/service.go:617-653`
  - Status: ✅ 已完成 - 已修改 `getCallBoardQueue` 方法，增加 `timeoutLimit int` 参数（单位：分钟）。如果 `timeoutLimit > 0`，在遍历 Redis 结果时，根据 score（订单完成时间）进行过滤：`当前时间 - score > timeoutLimit * 60`（秒）时跳过该订单。使用 `time.Now().Unix()` 获取当前时间（Unix 时间戳，单位：秒）

---

## Phase 2: 测试

### 单元测试

- [ ] 2.1 编写 Service 层单元测试

  - File: `main/app/service/callboard/service_test.go`（如不存在则创建）
  - Purpose: 确保过滤逻辑正确实现
  - Requirements: 1.2, 1.3, 1.4, 2.1, 2.2, 2.3
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 `GetQueueData` 和 `getCallBoardQueue` 方法编写单元测试，测试以下场景：1) `timeout_limit` 为 0 时不过滤；2) `timeout_limit > 0` 时过滤超时订单；3) `timeout_limit` 为 nil 时不过滤；4) `PreparingQueue` 不受 `timeout_limit` 影响；5) 边界情况（订单完成时间刚好等于 `timeout_limit` 分钟，`PreparedQueue` 为空）| Context: 使用 mock Redis 客户端，创建测试数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 单元测试编写完成，所有测试通过，覆盖率达标

### 集成测试

- [ ] 2.2 编写 API 集成测试

  - File: `main/app/api/v1/callboard/handler_test.go`（如不存在则创建）
  - Purpose: 测试 `/callboard/data` 接口的过滤功能
  - Requirements: 1.2, 1.4, 2.1, 2.2, 2.3
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 `/callboard/data` 接口编写集成测试，测试以下场景：1) 不同 `timeout_limit` 配置下的过滤效果；2) `PreparedQueue` 中订单的过滤逻辑；3) `PreparingQueue` 不受影响；4) 配置更新后立即生效 | Context: 使用测试数据库和 Redis，创建测试订单数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 集成测试编写完成，所有测试通过

---

## Phase 3: 文档和验证

### 文档更新

- [x] 3.1 更新 API 文档（Swagger）

  - File: `main/docs/swagger.json` 和 `main/docs/swagger.yaml`
  - Purpose: 更新 API 文档，说明 `prepared_queue` 会根据 `timeout_limit` 过滤
  - Requirements: 1.2, 1.4
  - Leverage: 现有 Swagger 文档
  - Status: ⏭️ 无需更新 - API 文档无需更新，响应结构未变化，只是返回内容会根据 `timeout_limit` 过滤

### 手动测试

- [x] 3.2 手动测试验证

  - File: -
  - Purpose: 在开发环境手动测试功能
  - Requirements: 所有需求
  - Leverage: 开发环境
  - Status: ✅ 已完成 - 手动测试完成，所有测试场景通过

---

## 📝 实现检查清单

### 代码实现

- [x] `GetQueueData` 方法修改完成，正确传递 `timeout_limit` 参数
- [x] `getCallBoardQueue` 方法修改完成，实现过滤逻辑
- [x] 处理 `timeout_limit` 为 nil 或 0 的情况
- [x] 确保 `PreparingQueue` 不受过滤影响

### 测试

- [ ] 单元测试编写完成，覆盖率 ≥ 70%
- [ ] 集成测试编写完成，所有测试通过
- [x] 手动测试完成，所有场景验证通过

### 文档

- [x] API 文档更新完成（无需更新）
- [x] 代码注释完善

---

## 🎯 关键实现点

### 1. 时间单位转换

- `timeout_limit` 配置单位：**分钟**
- Redis score 单位：**秒**（Unix 时间戳）
- 过滤条件：`当前时间(秒) - 订单完成时间(秒) > timeout_limit(分钟) * 60`

### 2. 向后兼容

- `timeout_limit` 为 nil 或 0 时，不进行过滤
- `PreparingQueue` 始终不进行过滤（传递 `timeoutLimit = 0`）
- 现有设备配置行为保持不变

### 3. 性能考虑

- 过滤逻辑在内存中进行，不影响 Redis 查询性能
- 使用 `ZRevRangeByScoreWithScores` 一次性获取所有数据
- 过滤操作在获取数据后立即进行

---

## 📚 参考资料

### 设计文档

- `design.md` - 技术设计文档
- `requirements.md` - 需求文档

### 代码参考

- `main/app/service/callboard/service.go` - 叫号系统服务实现
- `docs/shared/specs/active/story-callboard-data-config/design.md` - 相关功能设计文档

### 规范文档

- `.cursor/rules/go-main.mdc` - Go Main 开发规范
- `.cursor/rules/api.mdc` - API 设计规范

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: 王昱  
**审核者**: {审核者}
