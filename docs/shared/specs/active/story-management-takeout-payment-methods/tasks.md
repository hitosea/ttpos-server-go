# Grab/LINE MAN 外卖支付方式自动创建 任务分解

> 本文档定义 Grab/LINE MAN 外卖支付方式自动创建功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 10  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 常量定义

- [x] 1.1 添加 Grab 和 LINE MAN 支付方式 code 常量和名称常量

  - File: `main/app/constant/payment.go`
  - Purpose: 定义 Grab 和 LINE MAN 支付方式的固定 code 值和名称常量
  - Requirements: 1.4, 2.4
  - Leverage: 现有常量定义: `main/app/constant/payment.go`，参考 LianLianPay 的定义方式
  - Prompt: Role: Go Developer | Task: 在 payment.go 中添加 PaymentMethodCodeGrab = 91100、PaymentMethodCodeLineMan = 91200、PaymentMethodNameGrab = "Grab" 和 PaymentMethodNameLineMan = "LINE MAN" 常量 | Context: 参考 PaymentMethodCodeLianLianWechatPay = 90111 的定义方式，code 常量放在 LianLianPay 常量之后，名称常量单独定义一组 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 常量定义正确，注释清晰

---

## Phase 2: Service 层实现（Go Main）

- [x] 2.1 实现 SaveGrabPaymentMethod 方法

  - File: `main/app/service/payment_method.go`
  - Purpose: 保存 Grab 支付方式，如果已存在则跳过（幂等性）。供外部服务调用，支持事务。
  - Requirements: 1.1-1.6
  - Leverage: 现有 Create 方法: `main/app/service/payment_method.go:354-457`，ERP 同步逻辑: `main/app/service/payment_method.go:420-443`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 实现 SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error 方法，创建 Grab 支付方式（code=constant.PaymentMethodCodeGrab）| Context: 1) 方法接收 tx *gorm.DB 参数，供外部服务在事务中调用；2) 检查支付方式是否已存在（通过 payment_name=constant.PaymentMethodNameGrab 或 code=constant.PaymentMethodCodeGrab）；3) 如果不存在，创建支付方式（Source=0, IsShowCashier=0等，DefaultImg=""，使用常量 PaymentMethodNameGrab 和 PaymentMethodCodeGrab）；4) 如果开启ERP，同步到ERP，ERP 同步失败时返回错误"创建 Grab 支付方式失败"，阻塞流程；5) 记录日志；6) 所有数据库操作使用传入的 tx，不内部开启事务；7) 错误检查使用 strings.Contains(err.Error(), "record not found") 而不是 errors.Is | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，ERP 同步失败必须返回错误并阻塞流程，DefaultImg 为空字符串，使用常量而非硬编码字符串 | Success: 方法实现完整，幂等性正确，ERP 同步失败时返回错误并阻塞流程，支持事务调用，使用常量

- [x] 2.2 实现 SaveLineManPaymentMethod 方法

  - File: `main/app/service/payment_method.go`
  - Purpose: 保存 LINE MAN 支付方式，如果已存在则跳过（幂等性）。供外部服务调用，支持事务。
  - Requirements: 2.1-2.6
  - Leverage: Task 2.1 的实现，现有 Create 方法
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 实现 SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error 方法，创建 LINE MAN 支付方式（code=constant.PaymentMethodCodeLineMan）| Context: 参考 SaveGrabPaymentMethod 的实现，但 code=constant.PaymentMethodCodeLineMan，PaymentName=constant.PaymentMethodNameLineMan，DefaultImg=""，ERP 同步失败时返回错误"创建 LINE MAN 支付方式失败"，阻塞流程，错误检查使用 strings.Contains(err.Error(), "record not found") | Restrictions: 遵循 .cursor/rules/go-main.mdc，方法接收 tx *gorm.DB 参数，ERP 同步失败必须返回错误并阻塞流程，使用常量而非硬编码字符串 | Success: 方法实现完整，与 SaveGrabPaymentMethod 逻辑一致，ERP 同步失败时返回错误并阻塞流程，支持事务调用，使用常量

- [x] 2.3 修改 GetList 方法，过滤 Grab/LINE MAN 支付方式

  - File: `main/app/service/payment_method.go`
  - Purpose: 在获取支付方式列表时过滤掉 Grab 和 LINE MAN 支付方式
  - Requirements: 1.3, 2.3
  - Leverage: 现有 GetList 方法: `main/app/service/payment_method.go:85-174`
  - Prompt: Role: Go Developer | Task: 修改 GetList 方法，在返回结果前过滤掉 code=constant.PaymentMethodCodeGrab 或 constant.PaymentMethodCodeLineMan 的支付方式 | Context: 在 paymentMethodItems 构建循环中，添加过滤逻辑，跳过 code=constant.PaymentMethodCodeGrab 或 constant.PaymentMethodCodeLineMan 的支付方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用常量而非硬编码数字 | Success: 过滤逻辑正确，不影响其他支付方式显示，使用常量

- [x] 2.4 修改 GetManagementList 方法，过滤 Grab/LINE MAN 支付方式

  - File: `main/app/service/payment_method.go`
  - Purpose: 在获取管理端支付方式列表时过滤掉 Grab 和 LINE MAN 支付方式
  - Requirements: 1.3, 2.3
  - Leverage: 现有 GetManagementList 方法: `main/app/service/payment_method.go:176-237`，Repository WhereNotCode 方法: Task 2.5
  - Prompt: Role: Go Developer | Task: 修改 GetManagementList 方法，在查询时过滤掉 code=constant.PaymentMethodCodeGrab 或 constant.PaymentMethodCodeLineMan 的支付方式 | Context: 在 opts 中添加 WhereNotCode([]int{constant.PaymentMethodCodeGrab, constant.PaymentMethodCodeLineMan}) 选项 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用常量而非硬编码数字 | Success: 过滤逻辑正确，使用 Repository 选项方法，使用常量

---

## Phase 3: Repository 层扩展（Go Main）

- [x] 3.1 添加 WherePaymentName 选项方法

  - File: `main/app/repository/payment_method.go`
  - Purpose: 添加按支付方式名称查询的选项方法
  - Requirements: 2.1, 2.2
  - Leverage: 现有选项方法: `main/app/repository/payment_method.go:115-161`
  - Prompt: Role: Go Developer with GORM expertise | Task: 添加 WherePaymentName 选项方法，用于按 payment_name 字段查询支付方式 | Context: 参考现有的 Where 方法实现，如 WhereStatus、WhereCashier 等 | Restrictions: 遵循选项模式，返回 DBOption | Success: 方法实现正确，可以用于查询过滤

- [x] 3.2 添加 WhereCode 选项方法

  - File: `main/app/repository/payment_method.go`
  - Purpose: 添加按支付方式 code 查询的选项方法
  - Requirements: 2.1, 2.2
  - Leverage: 现有选项方法: `main/app/repository/payment_method.go:115-161`
  - Prompt: Role: Go Developer with GORM expertise | Task: 添加 WhereCode 选项方法，用于按 code 字段查询支付方式 | Context: 参考现有的 Where 方法实现 | Restrictions: 遵循选项模式，返回 DBOption | Success: 方法实现正确，可以用于查询过滤

- [x] 3.3 添加 WhereNotCode 选项方法（如不存在）

  - File: `main/app/repository/payment_method.go`
  - Purpose: 添加排除指定 code 的查询选项方法
  - Requirements: 2.4
  - Leverage: 现有选项方法: `main/app/repository/payment_method.go:115-161`，检查是否已存在 WhereNotCode 方法
  - Prompt: Role: Go Developer with GORM expertise | Task: 检查并添加 WhereNotCode 选项方法（如果不存在），用于排除指定 code 列表的支付方式 | Context: 先检查方法是否已存在，如果不存在则添加 | Restrictions: 遵循选项模式，返回 DBOption | Success: 方法实现正确，可以用于查询过滤

---

## Phase 4: PHP Admin 模块

- [x] 4.1 修改旧后台支付方式列表，过滤 Grab/LINE MAN

  - File: `admin/app/shop/controller/setting/Paytype.php`
  - Purpose: 在旧后台获取支付方式列表时过滤掉 Grab 和 LINE MAN 支付方式
  - Requirements: 1.3, 2.3
  - Leverage: 现有 index 方法: `admin/app/shop/controller/setting/Paytype.php:26-39`，PayType Model: `admin/app/common/model/store/PayType.php`
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 修改 Paytype::index 方法，过滤掉 Grab 和 LINE MAN 支付方式 | Context: 在 PayTypeModel::list 调用后，过滤掉 code=91100 或 91200 的支付方式（Grab 和 LINE MAN），或在 Model 层添加过滤逻辑 | Restrictions: 遵循 .cursor/rules/php.mdc，Controller 不写业务逻辑 | Success: 过滤逻辑正确，不影响其他支付方式显示

---

## Phase 5: 测试

- [x] 5.1 编写 Service 单元测试

  - File: `main/app/service/payment_method_test.go`
  - Purpose: 测试 SaveGrabPaymentMethod 和 SaveLineManPaymentMethod 方法
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaveGrabPaymentMethod 和 SaveLineManPaymentMethod 编写单元测试 | Context: 测试支付方式创建、幂等性、ERP 同步、错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [x] 5.2 编写列表过滤测试

  - File: `main/app/service/payment_method_test.go`
  - Purpose: 测试 GetList 和 GetManagementList 方法的过滤逻辑
  - Requirements: 1.3, 2.3
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 测试支付方式列表过滤功能 | Context: 测试 GetList 和 GetManagementList 方法是否正确过滤掉 Grab/LINE MAN 支付方式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 过滤测试通过，Grab/LINE MAN 支付方式不显示

- [ ] 5.3 集成测试

  - File: `test/integration/payment_method_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试配置 Grab/LINE MAN 外卖后自动创建支付方式，测试支付方式列表不显示 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
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
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-management-takeout-payment-methods/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-management-takeout-payment-methods/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-management-takeout-payment-methods/tasks.md
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
**最后更新**: 2025-12-22  
**维护者**: 后端开发组

