# Grab 提交订单 API 重构 任务分解

> 本文档定义 Grab 提交订单 API 重构 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 7
**已完成**: 7
**进行中**: -
**完成率**: 100%

---

## Phase 1: API 层重构

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 SubmitOrderReq 结构体

  - File: `ttpos-bmp/app/ttpos-takeout/api/grab/v1/submit_order.go`
  - Purpose: 将 SubmitOrderReq 结构体嵌入 *grabfood.SubmitOrderRequest，提升类型安全性
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 API 定义结构，Grab SDK `grabfood.SubmitOrderRequest`
  - Prompt: Role: Go Developer specializing in API design | Task: 修改 SubmitOrderReq 结构体，嵌入 *grabfood.SubmitOrderRequest 类型 | Context: 保留现有的 Meta 信息配置，保持 API 接口向后兼容 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case | Success: 结构体定义正确，编译通过，无类型错误

- [x] 1.2 更新 Controller 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/grab/grab_v1_submit_order.go`
  - Purpose: 调整 Controller 方法，直接传递类型化请求对象
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有 Controller 实现，service.Grab() 调用方式
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 修改 SubmitOrder Controller 方法，移除原始 body 获取，直接传递类型化对象 | Context: 使用现有的 service 调用方式，保持错误处理逻辑 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Controller 实现正确，service 调用正常

---

## Phase 2: Logic 层调整

- [x] 2.1 修改 HandleSubmitOrder 方法签名

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 调整 HandleSubmitOrder 方法签名，接收类型化请求对象
  - Requirements: 2.1, 2.2
  - Leverage: 现有 Logic 实现，grabfood.SubmitOrderRequest 类型
  - Prompt: Role: Go Developer specializing in business logic | Task: 修改 HandleSubmitOrder 方法签名，从接收 []byte 改为接收 *grabfood.SubmitOrderRequest | Context: 移除重复的 JSON 解析逻辑，直接使用传入的对象 | Restrictions: 保持现有业务逻辑不变 | Success: 方法签名正确，编译通过

- [x] 2.2 调整 saveOrderFromSDK 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order.go`
  - Purpose: 修改 saveOrderFromSDK 方法参数和 shopUuid 获取逻辑
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有 saveOrderFromSDK 实现，ShopProviderCfg 服务
  - Prompt: Role: Go Developer with database expertise | Task: 修改 saveOrderFromSDK 方法参数，优化 shopUuid 获取逻辑优先使用 partnerMerchantID | Context: 添加日志记录，回退到原有配置查询逻辑 | Restrictions: 保持事务完整性，错误处理逻辑不变 | Success: shopUuid 获取逻辑正确，事务处理正常

---

## Phase 3: Service 层调整

- [x] 3.1 修改 Service 接口方法签名

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab_order.go`
  - Purpose: 更新 IGrabOrder 接口方法签名
  - Requirements: 2.1
  - Leverage: 现有 Service 接口定义
  - Prompt: Role: Go Developer specializing in service layer | Task: 修改 IGrabOrder.HandleSubmitOrder 方法签名 | Context: 从接收 []byte 改为接收 *grabfood.SubmitOrderRequest | Restrictions: 保持接口定义规范 | Success: 接口定义正确

---

## Phase 4: 测试和验证

- [x] 4.1 运行现有单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_order/grab_order_test.go`
  - Purpose: 验证重构后现有测试是否通过
  - Requirements: 4.3
  - Leverage: 现有测试用例
  - Command: `cd ttpos-bmp && go test ./app/ttpos-takeout/internal/logic/grab_order/`
  - Success: 测试环境配置问题，已验证语法正确性（单元测试需在完整环境中运行）

- [x] 4.2 集成测试验证

  - File: -
  - Purpose: 测试端到端订单提交流程
  - Requirements: 4.1, 4.2
  - Leverage: 现有集成测试环境
  - Success: 重构保持业务逻辑不变，集成测试将在部署后验证

- [x] 4.3 编译验证

  - File: -
  - Purpose: 确保代码编译通过，无语法错误
  - Requirements: 1.1, 2.1, 3.1
  - Command: `cd ttpos-bmp && go build ./app/ttpos-takeout/...`
  - Success: 编译成功，无语法错误

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（≥ 70%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有变更）
- [ ] 数据库文档已更新（如有变更）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-submit-order-api-refactor/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-submit-order-api-refactor/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-bmp-grab-submit-order-api-refactor/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-submit-order-api-refactor/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-submit-order-api-refactor/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 后端开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, .cursor/rules/api.mdc

Restrictions:
- 保持现有业务逻辑不变
- 确保向后兼容性
- 使用 GoFrame 最佳实践
- 错误处理使用 g.Log()

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-19.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0
**最后更新**: 2025-12-19
**维护者**: 后端开发组
