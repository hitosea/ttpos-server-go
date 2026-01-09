# 新管理端-品牌采购-单据编号命名规则优化 任务分解

> 本文档定义 品牌采购单据编号命名规则优化 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 9  
**进行中**: -  
**完成率**: 60%

---

## Phase 1: 常量定义和基础准备

- [x] 1.1 扩展编号类型常量

  - File: `main/app/constant/number_sequence.go`
  - Purpose: 添加新的编号类型常量，用于区分不同类型的单据
  - Requirements: 1.1, 2.1-2.6
  - Leverage: 现有常量: `main/app/constant/number_sequence.go`
  - Prompt: Role: Go Developer | Task: 在 constant/number_sequence.go 中添加新的编号类型常量 | Context: 需要添加采购申请、采购收货、品牌采购、品采收货、盘点单、调拨单的编号类型常量 | Restrictions: 遵循现有命名规范，使用小写字母和下划线 | Success: 常量定义完成，格式正确

---

## Phase 2: Helper 层修改（编号生成逻辑）

### 采购订单 Helper

- [x] 2.1 修改采购申请/品牌采购编号生成方法

  - File: `main/app/service/purchase_order/helper.go`
  - Purpose: 将 generateOrderNo 方法改为使用 ttpos_number_sequence 表，时间戳格式改为 yyyyMMddHHmmss
  - Requirements: 1.1, 1.3, 1.4, 3.1
  - Leverage: 现有方法: `main/app/service/purchase_order/helper.go:34-82`，参考示例: `main/app/service/order.go:5504-5518`，Repository: `main/app/repository/number_sequence.go`
  - Prompt: Role: Go Developer specializing in Helper Layer | Task: 修改 generateOrderNo 方法，使用 ttpos_number_sequence 表生成序列号，时间戳格式改为 yyyyMMddHHmmss | Context: 需要传入 saasDB、companyUuid、prefix、numberType、timezone 参数，使用 NumberSequenceRepo.GetNextSequence 获取序列号，格式为 prefix + yyyyMMddHHmmss + 序列号（4位） | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 errors.WithMessage 包装错误，不使用 panic | Success: 方法修改完成，编号格式正确，错误处理完善

- [x] 2.2 修改采购收货/品采收货编号生成方法

  - File: `main/app/service/purchase_order/helper.go`
  - Purpose: 将 generateReceiptNo 方法改为使用 ttpos_number_sequence 表，时间戳格式改为 yyyyMMddHHmmss
  - Requirements: 1.2, 1.4, 3.1
  - Leverage: 现有方法: `main/app/service/purchase_order/helper.go:84-134`，Task 2.1 的实现
  - Prompt: Role: Go Developer specializing in Helper Layer | Task: 修改 generateReceiptNo 方法，使用 ttpos_number_sequence 表生成序列号，时间戳格式改为 yyyyMMddHHmmss | Context: 实现逻辑同 generateOrderNo，前缀为 PRC 或 TPHY | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改完成，编号格式正确

- [ ] 2.3 删除旧的序列号生成方法

  - File: `main/app/service/purchase_order/helper.go`
  - Purpose: 删除 generatePurchaseOrderSerialNo 和 generateReceiptOrderSerialNo 方法，不再需要
  - Requirements: 1.1, 1.2
  - Leverage: Task 2.1, 2.2 完成后删除旧方法
  - Success: 旧方法已删除，代码清理完成

### 盘点单 Helper

- [x] 2.4 修改盘点单编号生成方法

  - File: `main/app/service/stock_reconciliation.go`
  - Purpose: 将 generateOrderNo 方法改为使用 ttpos_number_sequence 表，时间戳格式改为 yyyyMMddHHmmss
  - Requirements: 1.5, 3.1
  - Leverage: 现有方法: `main/app/service/stock_reconciliation.go:855-894`，参考示例: `main/app/service/order.go:5504-5518`，Repository: `main/app/repository/number_sequence.go`
  - Prompt: Role: Go Developer specializing in Helper Layer | Task: 修改 generateOrderNo 方法，使用 ttpos_number_sequence 表生成序列号，时间戳格式改为 yyyyMMddHHmmss | Context: 需要传入 saasDB、companyUuid、timezone 参数，前缀为 ST，格式为 ST + yyyyMMddHHmmss + 序列号（4位） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改完成，编号格式正确

### 调拨单 Helper

- [x] 2.5 修改调拨单编号生成方法

  - File: `main/app/service/transfer_order/helper.go`
  - Purpose: 将 GenerateOrderNo 方法改为使用 ttpos_number_sequence 表，时间戳格式改为 yyyyMMddHHmmss，移除 Redis 依赖
  - Requirements: 1.6, 3.1
  - Leverage: 现有方法: `main/app/service/transfer_order/helper.go:35-80`，参考示例: `main/app/service/order.go:5504-5518`，Repository: `main/app/repository/number_sequence.go`
  - Prompt: Role: Go Developer specializing in Helper Layer | Task: 修改 GenerateOrderNo 方法，使用 ttpos_number_sequence 表生成序列号，时间戳格式改为 yyyyMMddHHmmss，移除 Redis 依赖 | Context: 需要传入 saasDB、companyUuid、timezone 参数，前缀为 TR，格式为 TR + yyyyMMddHHmmss + 序列号（4位） | Restrictions: 遵循 .cursor/rules/go-main.mdc，移除 Redis 相关代码 | Success: 方法修改完成，编号格式正确，Redis 依赖已移除

---

## Phase 3: Service 层修改（调用 Helper）

### 采购订单 Service

- [x] 3.1 修改采购申请/品牌采购创建方法

  - File: `main/app/service/purchase_order/purchase_order.go`
  - Purpose: 修改 CreatePurchaseOrder 方法，调用新的 generateOrderNo 方法，传入正确的参数
  - Requirements: 1.1, 1.3, 3.1
  - Leverage: 现有方法: `main/app/service/purchase_order/purchase_order.go:293-351`，Task 2.1 的 Helper
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 修改 CreatePurchaseOrder 方法，使用新的 generateOrderNo 方法 | Context: 需要获取 saasDB、companyUuid（使用总部 UUID 或当前公司 UUID），确定 prefix 和 numberType，调用 helper.generateOrderNo | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理完善 | Success: Service 方法修改完成，参数传递正确

- [x] 3.2 修改采购收货/品采收货创建方法

  - File: `main/app/service/purchase_order/receipt_order.go`
  - Purpose: 修改创建收货单的方法，调用新的 generateReceiptNo 方法
  - Requirements: 1.2, 1.4, 3.1
  - Leverage: 现有方法: `main/app/service/purchase_order/receipt_order.go`，Task 2.2 的 Helper
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 修改创建收货单的方法，使用新的 generateReceiptNo 方法 | Context: 需要获取 saasDB、companyUuid，确定 prefix 和 numberType，调用 helper.generateReceiptNo | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 方法修改完成

### 盘点单 Service

- [x] 3.3 修改盘点单创建方法

  - File: `main/app/service/stock_reconciliation.go`
  - Purpose: 修改 SaveStockReconciliation 方法，调用新的 generateOrderNo 方法
  - Requirements: 1.5, 3.1
  - Leverage: 现有方法: `main/app/service/stock_reconciliation.go:339-380`，Task 2.4 的 Helper
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 修改 SaveStockReconciliation 方法，使用新的 generateOrderNo 方法 | Context: 需要获取 saasDB、companyUuid，调用 generateOrderNo | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 方法修改完成

### 调拨单 Service

- [x] 3.4 修改调拨单创建方法

  - File: `main/app/service/transfer_order/transfer_order.go`
  - Purpose: 修改 CreateTransferOrder 方法，调用新的 GenerateOrderNo 方法，传入正确的参数
  - Requirements: 1.6, 3.1
  - Leverage: 现有方法: `main/app/service/transfer_order/transfer_order.go:542-696`，Task 2.5 的 Helper
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 修改 CreateTransferOrder 方法，使用新的 GenerateOrderNo 方法 | Context: 需要获取 saasDB、companyUuid，调用 helper.GenerateOrderNo，移除 Redis 锁相关代码 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 方法修改完成，Redis 锁已移除

---

## Phase 4: 测试

### 单元测试

- [ ] 4.1 编写采购订单 Helper 单元测试

  - File: `main/app/service/purchase_order/helper_test.go`
  - Purpose: 测试 generateOrderNo 和 generateReceiptNo 方法的正确性
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/purchase_order/`，参考: `main/app/service/order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 purchaseOrderHelper 编写单元测试，覆盖率 100% | Context: 测试编号格式（前缀 + yyyyMMddHHmmss + 序列号），测试序列号递增，测试并发场景，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

- [ ] 4.2 编写盘点单 Helper 单元测试

  - File: `main/app/service/stock_reconciliation_test.go`
  - Purpose: 测试 generateOrderNo 方法的正确性
  - Requirements: 1.5
  - Leverage: 现有测试: `main/app/service/stock_reconciliation_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 stockReconciliationSrv.generateOrderNo 编写单元测试 | Context: 测试编号格式，测试序列号递增，测试并发场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

- [ ] 4.3 编写调拨单 Helper 单元测试

  - File: `main/app/service/transfer_order/helper_test.go`
  - Purpose: 测试 GenerateOrderNo 方法的正确性
  - Requirements: 1.6
  - Leverage: 现有测试: `main/app/service/transfer_order/`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 transferOrderHelper.GenerateOrderNo 编写单元测试 | Context: 测试编号格式，测试序列号递增，测试并发场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

### 集成测试

- [ ] 4.4 编写编号唯一性集成测试

  - File: `test/integration/document_number_test.go`
  - Purpose: 测试同一秒内生成多个单据时，编号的唯一性
  - Requirements: 3.1, 3.2
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 编写编号唯一性集成测试 | Context: 测试同一秒内创建多个单据，验证编号不重复，测试并发场景 | Restrictions: 测试真实并发场景 | Success: 集成测试通过，编号唯一性验证通过

- [ ] 4.5 编写历史数据兼容性测试

  - File: `test/integration/document_number_compatibility_test.go`
  - Purpose: 测试历史数据使用旧格式编号，系统能正确识别
  - Requirements: 4.3
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 编写历史数据兼容性测试 | Context: 测试搜索功能支持新旧格式，测试前端展示兼容新旧格式 | Restrictions: 测试真实历史数据场景 | Success: 兼容性测试通过

---

## Phase 5: 代码审查和优化

- [ ] 5.1 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范遵循
  - Requirements: 所有需求
  - Leverage: `.cursor/rules/go-main.mdc`, `.cursor/rules/api.mdc`
  - Success: 代码审查通过，无规范问题

- [ ] 5.2 性能测试

  - File: -
  - Purpose: 验证编号生成性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 编号生成时间 < 50ms，并发测试通过

- [ ] 5.3 文档更新

  - File: `CHANGELOG.md`, `docs/shared/api/`（如有需要）
  - Purpose: 更新变更日志和 API 文档
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新 CHANGELOG 和相关文档 | Context: 记录编号格式变更，说明影响范围 | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Helper: 100%
  - Service: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - 编号格式正确（前缀 + yyyyMMddHHmmss + 序列号）
  - 编号唯一性验证通过
  - 历史数据兼容性验证通过

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] 设计文档已更新（如有调整）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-brand-purchase-document-number-naming-optimization/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-brand-purchase-document-number-naming-optimization/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-brand-purchase-document-number-naming-optimization/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-brand-purchase-document-number-naming-optimization/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-brand-purchase-document-number-naming-optimization/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的技术设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go Helper 开发

```
Role: Go Developer specializing in Helper Layer

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc

Restrictions:
- 使用 errors.WithMessage 包装错误
- 不使用 panic，返回 error
- 参数验证完善
- 错误处理完善

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 100%
```

### Go Service 开发

```
Role: Go Developer with Service Layer expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc

Restrictions:
- Service 只依赖其他 Service 接口
- 使用事务管理
- 错误处理完善
- 不使用 panic，返回 error

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
- Coverage target: 100% (Helper) 或 ≥ 70% (Service)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-24  
**维护者**: 后端开发组

