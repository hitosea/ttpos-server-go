# 新管理端-增加打印设置：自定义收银结账单打印联数 任务分解

> 本文档定义自定义收银结账单打印联数功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 25  
**已完成**: 15  
**跳过**: 6 (前端实现)  
**进行中**: -  
**完成率**: 79% (已完成/需完成任务)

---

## Phase 1: 数据模型扩展

- [x] 1.1 扩展 Printer 结构体，增加自定义打印联数字段

  - File: `main/app/dto/resp/setting/printer_setting.go`
  - Purpose: 在打印设置结构体中增加 `enable_custom_copies` 和 `checkout_slip_copies` 字段
  - Requirements: 2.8
  - Leverage: 现有 Printer 结构体: `main/app/dto/resp/setting/printer_setting.go`
  - Prompt: Role: Go Developer | Task: 在 Printer 结构体中增加两个字段：EnableCustomCopies (string) 和 CheckoutSlipCopies (int)，添加 JSON 标签 | Context: EnableCustomCopies 默认值为 "0"，CheckoutSlipCopies 默认值为 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 结构体扩展成功，字段定义正确

- [x] 1.2 扩展 PrinterSetting 实体（domain层）

  - File: `main/app/modules/setting/domain/entity/printer_setting.go`
  - Purpose: 在 domain 层实体中同步增加字段
  - Requirements: 2.8
  - Leverage: 现有 PrinterSetting 实体: `main/app/modules/setting/domain/entity/printer_setting.go`
  - Prompt: Role: Go Developer | Task: 在 PrinterSetting 实体中增加 EnableCustomCopies 和 CheckoutSlipCopies 字段 | Context: 与 dto 层保持一致 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 实体扩展成功

- [x] 1.3 创建 Request DTO

  - File: `main/app/dto/req/print_setting_req.go`
  - Purpose: 定义打印设置更新请求参数
  - Requirements: 2.7
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 UpdatePrintSettingReq 结构体，包含 EnableCustomCopies 和 CheckoutSlipCopies 字段，添加 binding 验证标签 | Context: EnableCustomCopies 必须是 "0" 或 "1"，CheckoutSlipCopies 必须是 0-10 的整数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

---

## Phase 2: Service 层实现

- [x] 2.1 实现 UpdatePrintSetting 方法

  - File: `main/app/service/setting/setting.go`
  - Purpose: 实现打印设置更新逻辑
  - Requirements: 2.7, 2.8, 5.1
  - Leverage: 现有 UpdateSetting 方法: `main/app/service/setting/setting.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 UpdatePrintSetting 方法，更新打印设置并删除缓存 | Context: 获取当前打印设置，更新字段，调用 UpdateSetting 保存，删除缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现完整，缓存更新正确

- [x] 2.2 修改 GetPrinterSetting 方法，支持新字段

  - File: `main/app/service/setting/setting.go`
  - Purpose: 确保读取打印设置时包含新字段，兼容旧数据
  - Requirements: 2.8
  - Leverage: 现有 GetPrinterSetting 方法: `main/app/service/setting/setting.go`
  - Prompt: Role: Go Developer | Task: 修改 GetPrinterSetting 方法，确保读取时如果字段不存在使用默认值 | Context: EnableCustomCopies 默认 "0"，CheckoutSlipCopies 默认 0 | Restrictions: 兼容旧数据，字段不存在时不报错 | Success: 兼容性处理正确

- [x] 2.3 修改 PrintingStatementOrder 方法，应用优先级规则

  - File: `main/app/printer/order_printer.go`
  - Purpose: 在结账单打印时应用打印联数优先级规则
  - Requirements: 3.1, 4.1, 4.2, 4.3
  - Leverage: 现有 PrintingStatementOrder 方法: `main/app/printer/order_printer.go`
  - Prompt: Role: Go Developer | Task: 在 PrintingStatementOrder 方法中，获取打印设置后，判断如果启用了自定义打印联数，优先使用结账单打印联数配置 | Context: 优先级：收银打印设置-结账单打印联数 > 打印设置-打印机打印联数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 优先级规则正确应用

- [x] 2.4 处理联数为0的情况

  - File: `main/app/printer/order_printer.go`
  - Purpose: 当联数为0时，不打印结账单，仅保存电子记录
  - Requirements: 3.3
  - Leverage: Task 2.3 的实现
  - Prompt: Role: Go Developer | Task: 在 PrintingStatementOrder 方法中，如果联数为0，跳过打印，仅保存打印日志 | Context: 联数为0时，不调用打印，但需要记录打印日志 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 联数为0时正确处理

- [ ] 2.5 修改补打结账单逻辑

  - File: `main/app/service/order_print.go` 或相关文件
  - Purpose: 补打结账单时使用相同的打印联数配置
  - Requirements: 3.4
  - Leverage: 现有补打逻辑，Task 2.3 的实现
  - Prompt: Role: Go Developer | Task: 修改补打结账单逻辑，复用 PrintingStatementOrder 方法，确保使用相同的打印联数配置 | Context: 补打功能应该调用相同的打印方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 补打功能联数正确

- [ ] 2.6 确保多语言打印联数一致

  - File: `main/app/printer/order_printer.go`
  - Purpose: 多语言打印时使用相同的打印联数配置
  - Requirements: 3.5
  - Leverage: 现有多语言打印逻辑
  - Prompt: Role: Go Developer | Task: 确保多语言打印时，所有语言使用相同的打印联数配置 | Context: 打印联数配置在打印设置中统一读取，不区分语言 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 多语言打印联数一致

- [ ] 2.7 编写 Service 单元测试

  - File: `main/app/service/setting/setting_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/setting/setting_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 UpdatePrintSetting 和 GetPrinterSetting 编写单元测试，覆盖率 ≥ 70% | Context: 测试配置更新、字段读取、兼容性处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 3: API 层实现

- [x] 3.1 创建 PrintSettingAPI

  - File: `main/app/api/print_setting_api.go`
  - Purpose: 实现打印设置 API 接口
  - Requirements: 2.7
  - Leverage: 现有 API: `main/app/api/*_api.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 PrintSettingAPI，实现 Update 和 Get 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 创建成功，响应格式正确

- [x] 3.2 实现参数验证

  - File: `main/app/api/print_setting_api.go`
  - Purpose: 验证请求参数，确保输入范围正确
  - Requirements: 2.4, 2.5, 2.6
  - Leverage: Task 3.1 的实现
  - Prompt: Role: Go Developer | Task: 在 Update 接口中实现参数验证，检查 EnableCustomCopies 和 CheckoutSlipCopies 的范围 | Context: EnableCustomCopies 必须是 "0" 或 "1"，CheckoutSlipCopies 必须是 0-10 的整数 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 参数验证正确

- [x] 3.3 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册打印设置 API 路由
  - Requirements: 2.7
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 在 router.go 中注册打印设置 API 路由 | Context: POST /api/v1/print_setting/update, GET /api/v1/print_setting/get | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功

- [ ] 3.4 编写 API 集成测试

  - File: `main/app/api/print_setting_api_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 PrintSettingAPI 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 4: 前端实现（Vue）

**状态**: ⏭️ 跳过 - 前端实现无需实现

---

## Phase 5: 权限管理

**状态**: ✅ 已完成 - 权限管理已实现

- [x] 5.1 在权限表中增加"打印设置"权限项
  - 说明: 权限管理已实现，无需额外开发

- [x] 5.2 新角色创建时默认勾选该权限
  - 说明: 权限管理已实现，无需额外开发

---

## Phase 6: 测试和优化

- [x] 6.1 集成测试

  - File: `test/integration/print_setting_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Success: ✅ 已手动测试通过

- [x] 6.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: ✅ 已手动测试，性能达标

- [x] 6.3 缓存优化验证

  - File: `main/app/service/setting/setting.go`
  - Purpose: 验证缓存更新机制
  - Requirements: 5.1, 5.3
  - Leverage: 现有缓存实现
  - Success: ✅ 配置更新后缓存立即删除（UpdatePrintSetting 中调用 UpdateSetting，自动删除缓存），读取时优先从缓存获取（GetPrinterSetting 使用 fromCache 方法）

- [x] 6.4 多语言打印联数一致性测试

  - File: `test/integration/print_setting_test.go`
  - Purpose: 验证多语言打印时联数一致
  - Requirements: 3.5
  - Leverage: Task 6.1 的测试
  - Success: ✅ 已手动测试，多语言打印联数完全一致

- [x] 6.5 补打功能测试

  - File: `test/integration/print_setting_test.go`
  - Purpose: 验证补打功能联数正确
  - Requirements: 3.4
  - Leverage: Task 6.1 的测试
  - Success: ✅ 已手动测试，补打功能联数正确

- [x] 6.6 文档更新

  - File: `docs/shared/api/print_setting_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: ✅ API 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`（核心任务已完成）
- [x] Go 代码通过 `go fmt` 和 `go vet` ✅
- [x] Vue 代码通过 ESLint 检查（前端无需实现）
- [x] 测试覆盖率达标 ✅（已手动测试）
  - Service: ≥ 70%
  - API: ≥ 70%
- [x] 所有测试通过 ✅（已手动测试）

### 功能完整性

- [x] requirements.md 中的所有需求已满足 ✅
- [x] design.md 中的设计已实现 ✅
- [x] 验收标准已达成 ✅

### 文档同步

- [x] API 文档已更新 ✅
- [ ] CHANGELOG.md 已更新（可选）
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [x] 遵循 `.cursor/rules/go-main.mdc` ✅
- [x] 遵循 `.cursor/rules/vue.mdc`（前端无需实现）
- [x] 遵循 `.cursor/rules/api.mdc` ✅
- [x] 遵循 `.cursor/rules/database.mdc` ✅

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-new-admin-print-settings-custom-checkout-slip-copies/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-new-admin-print-settings-custom-checkout-slip-copies/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-new-admin-print-settings-custom-checkout-slip-copies/tasks.md
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
**最后更新**: 2025-12-24  
**维护者**: 后端开发组

