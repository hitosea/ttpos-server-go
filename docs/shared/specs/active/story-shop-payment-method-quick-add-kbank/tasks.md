# 支付方式快捷添加（Kbank渠道）任务分解

> 本文档定义支付方式快捷添加（Kbank渠道）功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 12  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 常量和DTO扩展

- [x] 1.1 定义Kbank支付方式source常量

  - File: `main/app/constant/payment.go`
  - Purpose: 定义Kbank支付方式的source值常量
  - Requirements: 1.1, 2.3
  - Leverage: 现有常量定义: `main/app/constant/payment.go`，参考 `PaymentMethodSourceLianLianPay = 2`
  - Prompt: Role: Go Developer | Task: 在 constant/payment.go 中定义 PaymentMethodSourceKbank = 3，并在 PaymentMethodSourceTextMap 中添加映射 | Context: 参考现有 PaymentMethodSourceLianLianPay 的定义方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 常量定义成功，映射关系正确

- [x] 1.2 扩展 DefaultPaymentMethodResp 响应结构

  - File: `main/app/dto/resp/payment_method.go`
  - Purpose: 在响应结构中增加 can_add、source 字段（name字段即为payment_name，无需新增）
  - Requirements: 1.3
  - Leverage: 现有响应结构: `main/app/dto/resp/payment_method.go`
  - Prompt: Role: Go Developer | Task: 扩展 DefaultPaymentMethodResp 结构体，增加 CanAdd bool、Source int 字段 | Context: name字段即为payment_name，无需新增payment_name字段，保持向后兼容 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体扩展成功，字段定义正确

- [x] 1.3 扩展 PaymentMethodCreateItem 请求结构

  - File: `main/app/dto/req/payment_method_req.go`
  - Purpose: 在请求结构中增加 source 字段
  - Requirements: 2.1
  - Leverage: 现有请求结构: `main/app/dto/req/payment_method_req.go`
  - Prompt: Role: Go Developer | Task: 扩展 PaymentMethodCreateItem 结构体，增加 Source int 字段（可选字段） | Context: source字段为可选，默认值为0（后端会处理为默认值1） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体扩展成功，字段定义正确

---

## Phase 2: Repository层扩展

- [x] 2.1 新增 WhereSource 选项方法

  - File: `main/app/repository/payment_method.go`
  - Purpose: 实现按source字段查询的选项方法
  - Requirements: 1.4, 2.4
  - Leverage: 现有选项方法: `main/app/repository/payment_method.go`，参考 `WherePaymentName`、`WhereCode` 方法
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 PaymentMethodRepo 中新增 WhereSource(source int) DBOption 方法 | Context: 使用选项模式，返回 DBOption 函数，查询 source = ? 条件 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 方法实现正确，选项模式正确

- [x] 2.2 编写 Repository 单元测试

  - File: `main/app/repository/payment_method_test.go`
  - Purpose: 测试 WhereSource 方法
  - Requirements: 1.4, 2.4
  - Leverage: 现有测试: `main/app/repository/payment_method_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 WhereSource 方法编写单元测试 | Context: 测试查询source=3的支付方式，测试与其他选项方法组合使用 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 3: Service层扩展

- [x] 3.1 扩展 GetDefaultPayList 方法 - 增加Kbank支付方式定义

  - File: `main/app/service/payment_method.go`
  - Purpose: 在 GetDefaultPayList 方法中增加5种Kbank支付方式定义
  - Requirements: 1.1, 1.2
  - Leverage: 现有实现: `main/app/service/payment_method.go`，参考现有 defaultPayments 定义
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 GetDefaultPayList 方法开头增加Kbank支付方式定义，包含5种支付方式（code: 93000-93400） | Context: Kbank支付方式在最前面（sort值最小：0-4），使用 constant.PaymentMethodSourceKbank | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Kbank支付方式定义正确，sort值正确

- [x] 3.2 扩展 GetDefaultPayList 方法 - 实现重复检测逻辑

  - File: `main/app/service/payment_method.go`
  - Purpose: 查询已添加的Kbank支付方式，标记can_add字段
  - Requirements: 1.4, 1.5
  - Leverage: Task 2.1 的 WhereSource 方法，现有查询逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 GetDefaultPayList 中实现重复检测逻辑，查询已添加的Kbank支付方式，构建map用于快速查找，标记can_add字段 | Context: 使用 paymentMethodRepo.GetPaymentMethodList + WhereSource，构建 existingPaymentNames map，遍历Kbank支付方式时检查是否已添加 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 重复检测逻辑正确，can_add字段标记正确

- [x] 3.3 扩展 GetDefaultPayList 方法 - 构建响应数据

  - File: `main/app/service/payment_method.go`
  - Purpose: 构建包含can_add、source字段的响应数据（name字段即为payment_name）
  - Requirements: 1.3
  - Leverage: Task 1.2 的响应结构扩展，Task 3.1-3.2 的实现
  - Prompt: Role: Go Developer | Task: 在 GetDefaultPayList 中构建 DefaultPaymentMethodResp 响应数据，包含所有新增字段 | Context: 使用 baseUrl 构建完整图片URL，设置 CanAdd、Source 字段，Name 字段设置为 PaymentName（即payment_name） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 响应数据构建正确，所有字段填充正确

- [x] 3.4 扩展 Create 方法 - 处理source字段

  - File: `main/app/service/payment_method.go`
  - Purpose: 在Create方法中处理source字段，支持传入source值
  - Requirements: 2.2, 2.3
  - Leverage: 现有Create实现: `main/app/service/payment_method.go`，Task 1.3 的请求结构扩展
  - Prompt: Role: Go Developer with business logic expertise | Task: 在Create方法中处理source字段，如果传入source=0则使用默认值1，否则使用传入的值 | Context: 在创建paymentMethod时，使用 item.Source（如果提供），否则使用 constant.PaymentMethodSourceDefault | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: source字段处理正确，默认值逻辑正确

- [x] 3.5 扩展 Create 方法 - 实现重复检测

  - File: `main/app/service/payment_method.go`
  - Purpose: 创建前检查是否已存在相同的payment_name和source组合
  - Requirements: 2.4
  - Leverage: Task 2.1 的 WhereSource 方法，现有 WherePaymentName 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在Create方法开头实现重复检测逻辑，遍历所有items，检查是否存在相同的payment_name和source组合 | Context: 使用 paymentMethodRepo.GetPaymentMethod + WherePaymentName + WhereSource，如果已存在则返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误信息清晰 | Success: 重复检测逻辑正确，错误提示清晰

- [x] 3.6 编写 Service 单元测试

  - File: `main/app/service/payment_method_test.go`
  - Purpose: 测试GetDefaultPayList和Create的扩展功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/payment_method_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetDefaultPayList 和 Create 的扩展功能编写单元测试，覆盖率100% | Context: 测试Kbank支付方式返回、重复检测、can_add字段、source字段处理、批量创建 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Payment相关模块100%覆盖 | Success: 测试覆盖率100%，所有测试通过

---

## Phase 4: API测试和集成测试

- [x] 4.1 编写 GetDefaultPayList API 集成测试

  - File: `main/app/api/v1/shop/shop_payment_method_test.go`
  - Purpose: 测试扩展后的 GetDefaultPayList API
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有API测试: `main/app/api/v1/shop/shop_payment_method_test.go`
  - Prompt: Role: QA Automation Engineer | Task: 编写 GetDefaultPayList API 集成测试 | Context: 测试响应格式、Kbank支付方式在最前面、can_add字段正确性、source字段存在、name字段即为payment_name | Restrictions: 测试真实API调用 | Success: 所有API测试通过

- [x] 4.2 编写 Create API 集成测试

  - File: `main/app/api/v1/shop/shop_payment_method_test.go`
  - Purpose: 测试扩展后的 Create API（支持source参数）
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有API测试: `main/app/api/v1/shop/shop_payment_method_test.go`
  - Prompt: Role: QA Automation Engineer | Task: 编写 Create API 集成测试，测试source参数、重复检测、批量创建 | Context: 测试传入source=3创建Kbank支付方式、测试重复创建返回错误、测试批量创建多个支付方式 | Restrictions: 测试真实API调用 | Success: 所有API测试通过

- [x] 4.3 端到端集成测试

  - File: `test/integration/payment_method_kbank_test.go`
  - Purpose: 测试完整的Kbank支付方式快捷添加流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试流程：1. 调用GetDefaultPayList获取Kbank支付方式列表 2. 选择can_add=true的支付方式 3. 调用Create批量创建 4. 再次调用GetDefaultPayList验证can_add=false | Restrictions: 测试真实用户场景 | Success: 集成测试通过

---

## Phase 5: 性能优化和文档

- [ ] 5.1 数据库索引优化（可选）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_index_payment_method_payment_name_source.php`
  - Purpose: 添加复合索引优化重复检测查询
  - Requirements: 性能要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，添加 idx_payment_name_source 复合索引 | Context: 索引字段：payment_name, source，用于优化重复检测查询 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 索引创建成功，查询性能提升

- [ ] 5.2 更新API文档

  - File: `docs/shared/api/payment_method_api.md`（如有）
  - Purpose: 更新API文档，说明新增字段和功能
  - Requirements: 文档要求
  - Leverage: 现有API文档
  - Prompt: Role: Technical Writer | Task: 更新支付方式API文档，说明GetDefaultPayList和Create接口的扩展 | Context: 说明新增字段can_add、source、payment_name，说明Kbank支付方式code和source值 | Restrictions: 文档准确完整 | Success: API文档已更新

- [ ] 5.3 更新CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有CHANGELOG格式
  - Success: CHANGELOG已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: 100%（Payment相关）
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-payment-method-quick-add-kbank/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-payment-method-quick-add-kbank/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-payment-method-quick-add-kbank/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-payment-method-quick-add-kbank/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-payment-method-quick-add-kbank/tasks.md)" | bc
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

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-29  
**维护者**: 后端开发组

