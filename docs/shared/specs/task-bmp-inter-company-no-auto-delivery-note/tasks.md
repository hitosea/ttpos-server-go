# ttpos-bmp 内部采购不自动创建发货单 任务分解

> 本文档定义移除 ERPNext 内部销售订单自动创建发货单逻辑的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求
- **遵循规范**: 严格遵循 `erpnext.mdc` - 不修改 ERPNext 源码，不使用 Server Scripts

## 📊 进度总览

**总任务数**: 11  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 调研和分析（0.5 天）

### 目标

深入了解 ERPNext Inter Company Transaction 自动创建 Delivery Note 的机制，确定最优的禁用方案。

---

- [ ] 1.1 调研 ERPNext Inter Company Transaction 自动创建机制

  - File: -
  - Purpose: 了解 ERPNext 如何在 Sales Order 提交后自动创建 Delivery Note
  - Requirements: Requirement 1.1
  - Leverage: ERPNext 官方文档，ERPNext 源码（仅阅读，不修改）
  - Prompt: Role: ERPNext Integration Engineer | Task: 调研 ERPNext Inter Company Transaction 中 Sales Order 提交后自动创建 Delivery Note 的触发机制 | Context: 阅读 ERPNext 文档和源码，查找 `on_submit` 事件、工作流配置、Inter Company Settings | Restrictions: 只阅读分析，不修改 ERPNext 源码 | Success: 明确自动创建的触发点（工作流/代码逻辑/系统设置）

- [ ] 1.2 确认 ERPNext 系统设置中的配置项

  - File: -
  - Purpose: 确认是否存在系统级配置可以禁用自动创建 Delivery Note
  - Requirements: Requirement 1.2
  - Leverage: Task 1.1 的调研结果，ERPNext 文档
  - Prompt: Role: ERPNext Administrator | Task: 查找 ERPNext 系统设置中是否有禁用 Inter Company Transaction 自动创建 Delivery Note 的配置项 | Context: 检查 Setup > Settings > Inter Company Transaction Settings，查看是否有 "auto_create_delivery_note" 或类似配置 | Restrictions: 不修改任何配置，只确认是否存在 | Success: 明确是否有配置项，记录配置项名称和位置

- [ ] 1.3 分析 ttpos-bmp 中相关代码逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 分析当前 `CreateInnerSaleOrderFromPurchaseOrder` 的实现逻辑
  - Requirements: Requirement 1.2, 1.3
  - Leverage: Task 1.1, 1.2 的调研结果，现有代码 (lines 86-151)
  - Prompt: Role: Go Developer with ERPNext integration expertise | Task: 分析 ttpos-bmp 中 `CreateInnerSaleOrderFromPurchaseOrder` 的实现，识别是否有可以插入禁用逻辑的地方 | Context: 查看销售订单的创建、提交流程，确认在哪个环节可以禁用或删除自动创建的 Delivery Note | Restrictions: 遵循 `erpnext.mdc` - 不修改 ERPNext 源码 | Success: 明确在 ttpos-bmp 代码中的修改点

- [ ] 1.4 确定最优实现方案

  - File: `design.md`
  - Purpose: 根据调研结果，确定最优的实现方案（方案 A/B/C）
  - Requirements: Requirement 1.3
  - Leverage: Task 1.1, 1.2, 1.3 的分析结果
  - Success: 在 design.md 中标记选定的方案，说明选择理由

---

## Phase 2: 核心实现（1 天）

### 目标

实现禁用自动创建 Delivery Note 的逻辑，保留手动创建接口。

---

### 2.1 扩展 ERPNext 服务（如需要）

- [ ] 2.1.1 实现 `Document.Delete` 方法（如不存在）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go`
  - Purpose: 实现删除 ERPNext 文档的通用方法
  - Requirements: Requirement 1.3
  - Leverage: 现有 `service.Document()` 实现，ERPNext API 文档
  - Prompt: Role: Go Developer with ERPNext API expertise | Task: 在 `service.Document()` 中实现 `Delete(ctx, doctype, name)` 方法 | Context: 调用 ERPNext 的删除 API，传入 doctype 和 name，处理响应和错误 | Restrictions: 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`，与 ERPNext 交互通过 `service.Rpc()` | Success: Delete 方法实现完成，可成功删除 ERPNext 文档

- [ ] 2.1.2 实现 `Document.List` 方法（如不存在）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go`
  - Purpose: 实现查询 ERPNext 文档列表的通用方法
  - Requirements: Requirement 1.3
  - Leverage: 现有 `service.Document()` 实现，ERPNext API 文档
  - Prompt: Role: Go Developer with ERPNext API expertise | Task: 在 `service.Document()` 中实现 `List(ctx, doctype, params)` 方法 | Context: 调用 ERPNext 的查询 API，支持 filters 和 fields 参数，处理响应和错误 | Restrictions: 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`，返回 *gjson.Json 类型 | Success: List 方法实现完成，可成功查询 ERPNext 文档列表

### 2.2 实现删除自动创建 Delivery Note 的逻辑

- [ ] 2.2.1 实现 `removeAutoCreatedDeliveryNote` 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 查询并删除销售订单自动创建的 Draft 状态 Delivery Note
  - Requirements: Requirement 1.3
  - Leverage: Task 2.1.1, 2.1.2 的实现，`service.Document()` 服务
  - Prompt: Role: Go Developer with ERPNext integration expertise | Task: 实现 `removeAutoCreatedDeliveryNote(ctx, salesOrderName)` 方法，查询并删除指定销售订单自动创建的 Draft 状态 Delivery Note | Context: 使用 `service.Document().List()` 查询，使用 `service.Document().Delete()` 删除，只删除 docstatus=0 (Draft) 的发货单 | Restrictions: 遵循 `erpnext.mdc` - 不修改 ERPNext 源码，遵循 `go-rules.mdc`，记录详细日志 | Success: 方法实现完成，可成功删除自动创建的 Draft Delivery Note

- [ ] 2.2.2 修改 `CreateInnerSaleOrderFromPurchaseOrder` 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 在销售订单提交后，调用删除逻辑
  - Requirements: Requirement 1.1, 1.3, 1.4
  - Leverage: Task 2.2.1 的实现，现有代码 (lines 86-151)
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 `CreateInnerSaleOrderFromPurchaseOrder`，在提交销售订单后，异步调用 `removeAutoCreatedDeliveryNote` 删除自动创建的 Delivery Note | Context: 使用 `go func()` 异步执行，等待 2 秒后删除（确保 ERPNext 完成自动创建），删除失败只记录警告日志，不影响主流程 | Restrictions: 遵循 `erpnext.mdc` 和 `go-rules.mdc`，不使用 panic，返回 error | Success: 修改完成，提交销售订单后不会有自动创建的 Draft Delivery Note

### 2.3 验证手动创建接口

- [ ] 2.3.1 验证 `CreateDeliveryNoteFromInnerSaleOrder` 接口

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 确认手动创建发货单的接口保持不变且正常工作
  - Requirements: Requirement 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有代码 (lines 153-186)
  - Success: 接口保持不变，测试调用成功

---

## Phase 3: 测试（0.5 天）

### 目标

确保修改后的流程正常工作，现有功能不受影响。

---

- [ ] 3.1 编写集成测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go`
  - Purpose: 测试完整的内部采购流程，确认不自动创建 Delivery Note
  - Requirements: 所有功能需求
  - Leverage: 现有测试代码，Task 2.2 的实现
  - Prompt: Role: QA Automation Engineer with Go testing expertise | Task: 编写集成测试，测试内部采购流程（Material Request → Purchase Order → Sales Order → 手动创建 Delivery Note） | Context: 测试用例包括：1) 创建销售订单后无自动 Delivery Note，2) 手动创建 Delivery Note 成功 | Restrictions: 遵循 `go-rules.mdc`，使用 testify/assert | Success: 集成测试通过，验收标准满足

- [ ] 3.2 执行回归测试

  - File: -
  - Purpose: 确保其他 ERP 功能不受影响
  - Requirements: Requirement 1.4, 1.5
  - Leverage: 现有测试套件
  - Success: 所有回归测试通过，Material Request、Purchase Order、外部销售订单功能正常

- [ ] 3.3 性能测试

  - File: -
  - Purpose: 确保接口响应时间不受影响
  - Requirements: 非功能需求 - 性能要求
  - Leverage: 性能测试工具
  - Success: API 响应时间与修改前一致

---

## Phase 4: 文档更新（0.5 天）

### 目标

更新相关文档，确保开发人员和运营人员了解新的流程。

---

- [ ] 4.1 更新 API 文档

  - File: `ttpos-bmp/docs/shared/api/erp_api.md` 或类似文件
  - Purpose: 说明 `CreateDeliveryNoteFromInnerSaleOrder` 接口的用法
  - Requirements: Requirement 3.1, 3.2, 3.3
  - Leverage: 现有 API 文档，`docs/agent/templates/api-doc-template.md`
  - Success: API 文档更新完成，手动创建发货单流程说明清晰

- [ ] 4.2 更新内部采购流程文档

  - File: `docs/human/architecture/features/purchase_order.md` 或类似文件
  - Purpose: 说明新的内部采购流程（销售订单不自动创建发货单）
  - Requirements: Requirement 3.1
  - Leverage: 现有架构文档
  - Success: 流程文档更新完成，新流程说明清晰

- [ ] 4.3 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录此次流程变更
  - Requirements: Requirement 3.4
  - Leverage: 现有 CHANGELOG 格式
  - Success: CHANGELOG 更新完成，变更记录清晰

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc` - 未修改 ERPNext 源码，未使用 Server Scripts
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - [ ] 内部销售订单提交不自动创建发货单
  - [ ] 手动创建发货单正常工作
  - [ ] 现有数据和流程不受影响

### 文档同步

- [ ] API 文档已更新
- [ ] 内部采购流程文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] **遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc`**（核心约束）
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/task-bmp-inter-company-no-auto-delivery-note/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/task-bmp-inter-company-no-auto-delivery-note/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/task-bmp-inter-company-no-auto-delivery-note/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/task-bmp-inter-company-no-auto-delivery-note/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/task-bmp-inter-company-no-auto-delivery-note/tasks.md)" | bc
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

### ERPNext 集成开发

```
Role: Go Developer specializing in ERPNext Integration

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/erpnext.mdc

Restrictions (CRITICAL - erpnext.mdc):
- ❌ 不修改 ERPNext 源代码
- ❌ 不使用 ERPNext Server Scripts 功能
- ✅ 通过 ttpos-bmp/ttpos-erp 模块代码实现
- ✅ 与 ERPNext 交互通过 ttpos-erp/internal/logic/erpnext 下的通用服务
- ✅ JSON struct 定义在 model/dto 包中
- 遵循 GoFrame 2.x 规范
- 不使用 panic，返回 error
- 使用 gerror.Wrapf 包装错误
- 记录详细日志（g.Log()）

Success Criteria:
- {成功标准1}
- 未修改 ERPNext 源码
- 未使用 ERPNext Server Scripts
- 代码通过 go fmt 和 go vet
```

### Go 测试工程师

```
Role: QA Engineer with Go testing and ERPNext expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Test scenarios: {测试场景列表}

Test Cases Required:
- 正常场景测试（销售订单创建成功，无自动 Delivery Note）
- 手动创建发货单测试（成功创建并关联）
- 异常场景测试（ERPNext API 调用失败）
- 回归测试（其他 ERP 功能不受影响）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 使用 testify/assert
- 必须包含边界情况测试

Success Criteria:
- 所有测试通过
- 验收标准满足
- 边界情况已覆盖
```

### 文档更新

```
Role: Technical Writer with ERP domain knowledge

Task: 更新内部采购流程文档，说明新的流程（不自动创建发货单）

Context:
- Current file: {文档路径}
- Changes: 销售订单提交后不自动创建发货单，需手动创建
- Leverage: 现有文档格式

Requirements:
- 清晰说明流程变更
- 提供手动创建发货单的步骤
- 说明变更的原因和价值

Success Criteria:
- 文档准确完整
- 新旧流程对比清晰
- 操作步骤易于理解
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-19  
**维护者**: rikugun

