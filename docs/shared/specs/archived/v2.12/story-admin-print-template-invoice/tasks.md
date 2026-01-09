# 新管理端-自定义打印模板增加发票 任务分解

> 本文档定义新管理端发票打印模板的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 29  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 后端基础实现（Go Main）

### Repository 层扩展

- [ ] 1.1 扩展 Repository 接口

  - File: `main/app/repository/i_printer_template_repo.go`
  - Purpose: 新增发票模板相关接口方法
  - Requirements: Requirement 1, 4
  - Leverage: 现有 Repository 接口: `main/app/repository/printer_template.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 扩展 IPrinterTemplateRepo 接口，新增 GetInvoiceTemplates() 和 GetPrinterTemplateByUuid() 方法 | Context: 复用现有表 ttpos_printer_template，发票类型 template=3 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [ ] 1.2 实现 Repository 扩展方法

  - File: `main/app/repository/printer_template.go`
  - Purpose: 实现发票模板数据访问逻辑
  - Requirements: Requirement 1, 4
  - Leverage: 现有 Repository 实现: `main/app/repository/printer_template.go`
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 GetInvoiceTemplates() 和 GetPrinterTemplateByUuid() 方法 | Context: 查询 template=3 的记录，软删除(delete_time=0) | Restrictions: 只持有 db \*gorm.DB，使用 GORM | Success: Repository 实现完整，软删除正确

### DTO 层

- [ ] 1.3 创建发票模板 Request DTO

  - File: `main/app/dto/req/printer_req.go`
  - Purpose: 定义发票模板 API 请求参数
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有 DTO: `main/app/dto/req/printer_req.go`
  - Prompt: Role: Go Developer | Task: 创建发票模板 Request DTO，包含 Create, Update, Get, Delete, Use, RestoreDefault, Preview 请求结构体 | Context: 使用 binding 标签验证参数，模板名称限制1-20字符 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 1.4 创建发票模板 Response DTO

  - File: `main/app/dto/resp/printer_resp.go`
  - Purpose: 定义发票模板 API 响应数据
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有 DTO: `main/app/dto/resp/printer_resp.go`
  - Prompt: Role: Go Developer | Task: 创建发票模板 Response DTO，包含单条和列表响应结构体 | Context: 包含 uuid, name, is_default, is_advanced, is_using, tmp_data等字段 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

- [ ] 1.5 创建发票模板配置结构体

  - File: `main/app/printer/pkg/template/invoice_template.go`
  - Purpose: 定义发票模板配置的数据结构
  - Requirements: Requirement 1, 2, 3
  - Leverage: 现有模板结构: `main/app/printer/pkg/template/`
  - Prompt: Role: Go Developer | Task: 创建 InvoiceTemplateConfig 和相关结构体，用于解析和生成 JSON 配置 | Context: 支持发票字段、自定义文字/图片、分割线、空行等 | Restrictions: 遵循 JSON 序列化规范 | Success: 结构体定义完整，JSON 序列化正确

### Service 层

- [ ] 1.6 扩展 Service 接口

  - File: `main/app/service/i_printer_srv.go`
  - Purpose: 定义发票模板业务逻辑接口
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有 Service 接口: `main/app/service/printer.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 扩展 IPrinterSrv 接口，新增发票模板相关方法 | Context: 包含 GetList, Create, Update, Delete, Use, RestoreDefault, Preview 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 1.7 实现 GetInvoiceTemplateList

  - File: `main/app/service/printer.go`
  - Purpose: 实现获取发票模板列表逻辑
  - Requirements: Requirement 1, 4
  - Leverage: 现有 Service 实现: `main/app/service/printer.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetInvoiceTemplateList 方法 | Context: 查询 template=3 的记录，解析 tmp_data JSON，判断 is_using 状态 | Restrictions: 持有 DBManager，不使用 panic，返回 error | Success: 列表查询正确，JSON 解析正确，is_using 判断正确

- [ ] 1.8 实现 CreateInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现创建发票模板逻辑
  - Requirements: Requirement 1, 3
  - Leverage: 现有 Service 实现: `main/app/service/printer.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 CreateInvoiceTemplate 方法 | Context: 验证模板名称长度(1-20)，验证 JSON 格式，检查重复添加限制(最多5次)，生成 UUID | Restrictions: 不使用 panic，返回 error | Success: 创建成功，验证逻辑正确

- [ ] 1.9 实现 UpdateInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现更新发票模板逻辑
  - Requirements: Requirement 1, 3, 4
  - Leverage: Task 1.8
  - Success: 更新成功，验证逻辑正确

- [ ] 1.10 实现 DeleteInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现删除发票模板逻辑
  - Requirements: Requirement 4
  - Leverage: Task 1.7
  - Prompt: Role: Go Developer | Task: 实现 DeleteInvoiceTemplate 方法 | Context: 默认模板不可删除，正在使用的模板不可删除，使用软删除 | Restrictions: 检查 is_default 和 is_using 状态 | Success: 删除成功，业务规则正确

- [ ] 1.11 实现 UseInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现使用发票模板逻辑
  - Requirements: Requirement 4
  - Leverage: Task 1.7
  - Prompt: Role: Go Developer | Task: 实现 UseInvoiceTemplate 方法 | Context: 更新配置中的 invoice_template_uuid，仅可选择一个模板使用 | Restrictions: 使用 SettingSrv 更新配置 | Success: 使用成功，配置更新正确

- [ ] 1.12 实现 RestoreDefaultInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现恢复默认发票模板逻辑
  - Requirements: Requirement 4
  - Leverage: Task 1.8
  - Success: 恢复成功，模板数据正确

- [ ] 1.13 实现 PreviewInvoiceTemplate

  - File: `main/app/service/printer.go`
  - Purpose: 实现预览发票模板逻辑
  - Requirements: Requirement 4, 5
  - Leverage: 现有打印预览逻辑: `main/app/printer/template/`
  - Success: 预览生成成功，图片/HTML正确

### API 层

- [ ] 1.14 实现发票模板 API 接口

  - File: `main/app/api/v1/shop/shop_print.go`
  - Purpose: 实现发票模板 HTTP API 接口
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_print.go`，Task 1.6-1.13 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 shop_print.go 中新增发票模板相关接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 创建成功，响应格式正确

- [ ] 1.15 注册发票模板 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册发票模板 API 路由
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功

---

## Phase 2: Vue 前端可视化编辑器

### API 封装

- [ ] 2.1 创建发票模板 API 封装

  - File: `admin/views/shop/api/printTemplate.ts`
  - Purpose: 封装发票模板后端 API 调用
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有 API: `admin/views/shop/api/`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装发票模板 API 调用，包含 getList, create, update, delete, use, restoreDefault, preview 方法 | Context: 使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

### 可视化编辑器组件

- [ ] 2.2 创建发票模板列表页面

  - File: `admin/views/shop/pages/print-template/invoice/index.vue`
  - Purpose: 实现发票模板列表页面
  - Requirements: Requirement 1, 4
  - Leverage: 现有页面: `admin/views/shop/pages/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建发票模板列表页面，显示模板列表，支持新增/编辑/删除/使用/预览 | Context: 使用 Element Plus，使用 Composition API | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面创建成功，列表显示正确

- [ ] 2.3 创建左侧模板选择区组件

  - File: `admin/views/shop/components/print-template-editor/TemplateSelector.vue`
  - Purpose: 实现左侧模板选择项列表
  - Requirements: Requirement 1, 2
  - Leverage: Element Plus Tree 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建左侧模板选择区组件 | Context: 显示可添加的字段列表，点击后在编辑区显示配置项并高亮 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 组件创建成功，交互正确

- [ ] 2.4 创建右侧样式编辑区组件

  - File: `admin/views/shop/components/print-template-editor/StyleEditor.vue`
  - Purpose: 实现右侧样式编辑区
  - Requirements: Requirement 2, 3
  - Leverage: Element Plus Form 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建右侧样式编辑区组件 | Context: 显示选中项的配置，支持分割线/空行插入，支持删除 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 组件创建成功，功能完整

- [ ] 2.5 实现拖拽排序功能

  - File: `admin/views/shop/components/print-template-editor/DraggableList.vue`
  - Purpose: 实现拖拽排序功能
  - Requirements: Requirement 3
  - Leverage: Vuedraggable 库
  - Prompt: Role: Frontend Developer with Vue 3 + Vuedraggable expertise | Task: 使用 Vuedraggable 实现拖拽排序功能 | Context: 右侧编辑区所有项/组可拖动调整位置，默认模板不可拖动 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 拖拽功能正常，排序保存正确

- [ ] 2.6 实现重复添加限制和提示

  - File: `admin/views/shop/components/print-template-editor/ItemManager.ts`
  - Purpose: 实现项目重复添加限制和提示逻辑
  - Requirements: Requirement 3
  - Leverage: Element Plus Message 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现重复添加限制逻辑 | Context: 单项最多可重复添加5次，重复添加第三次时弹窗提示确认 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 限制逻辑正确，提示显示正确

- [ ] 2.7 实现自定义文字编辑

  - File: `admin/views/shop/components/print-template-editor/CustomTextEditor.vue`
  - Purpose: 实现自定义文字编辑功能
  - Requirements: Requirement 2
  - Leverage: Element Plus Input 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现自定义文字编辑 | Context: 支持多语言输入，最多500字符，未翻译的语言不打印 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 编辑功能正常，字符限制正确

- [ ] 2.8 实现自定义图片上传

  - File: `admin/views/shop/components/print-template-editor/CustomImageUploader.vue`
  - Purpose: 实现自定义图片上传功能
  - Requirements: Requirement 2
  - Leverage: Element Plus Upload 组件
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现自定义图片上传 | Context: 验证图片格式和大小，未上传时不打印 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 上传功能正常，验证逻辑正确

- [ ] 2.9 实现模板预览功能

  - File: `admin/views/shop/components/print-template-editor/TemplatePreview.vue`
  - Purpose: 实现模板预览功能
  - Requirements: Requirement 4, 5
  - Leverage: Task 2.1 的 API 封装
  - Success: 预览功能正常，显示正确

---

## Phase 3: PHP 旧商家后台兼容

- [ ] 3.1 扩展 PHP Controller

  - File: `admin/app/shop/controller/Print.php`
  - Purpose: 新增发票模板预览接口
  - Requirements: Requirement 5
  - Leverage: 现有 Controller: `admin/app/shop/controller/Print.php`
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 Print Controller 中新增发票模板预览方法 | Context: 遵循 MVC 分层，Controller 不写业务逻辑，调用 Go Main API | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Controller 创建成功，接口正确

- [ ] 3.2 创建预览页面

  - File: `admin/views/shop/print/invoice.html`
  - Purpose: 实现发票模板预览页面
  - Requirements: Requirement 5
  - Leverage: 现有页面: `admin/views/shop/print/`
  - Success: 页面创建成功，预览正确

---

## Phase 4: 测试和优化

- [ ] 4.1 单元测试 - Repository

  - File: `main/app/repository/printer_template_test.go`
  - Purpose: 测试 Repository 数据访问逻辑
  - Requirements: 所有需求
  - Leverage: 现有测试: `main/app/repository/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为发票模板 Repository 编写单元测试，覆盖率 ≥ 80% | Context: 测试 GetInvoiceTemplates, GetPrinterTemplateByUuid 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.2 单元测试 - Service

  - File: `main/app/service/printer_test.go`
  - Purpose: 测试 Service 业务逻辑
  - Requirements: 所有需求
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为发票模板 Service 编写单元测试，覆盖率 ≥ 70% | Context: 测试 GetList, Create, Update, Delete, Use, RestoreDefault 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.3 API 集成测试

  - File: `main/app/api/v1/shop/shop_print_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 所有需求
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Success: 所有 API 测试通过

- [ ] 4.4 E2E 测试

  - File: `test/e2e/print_template_invoice_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有需求
  - Leverage: 现有 E2E 测试
  - Success: E2E 测试通过

- [ ] 4.5 缓存优化

  - File: `main/app/service/printer.go`
  - Purpose: 实现 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Success: 缓存实现完成，命中率 > 80%

- [ ] 4.6 浏览器兼容性测试

  - File: -
  - Purpose: 测试浏览器兼容性
  - Requirements: 浏览器兼容性要求
  - Success: Chrome/Safari/Firefox/Edge 测试通过

- [ ] 4.7 打印机兼容性测试

  - File: -
  - Purpose: 测试不同打印机的兼容性
  - Requirements: 可靠性要求
  - Success: 主流打印机测试通过

- [ ] 4.8 文档更新

  - File: `docs/shared/api/print_template_api.md`, `CHANGELOG.md`
  - Purpose: 更新文档
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新发票模板相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/go-printer.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-print-template-invoice/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-print-template-invoice/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-print-template-invoice/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-print-template-invoice/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-print-template-invoice/tasks.md)" | bc
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
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/go-printer.mdc, .cursor/rules/api.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service) 或 ≥ 80% (Repository)
```

### PHP 后端开发

```
Role: PHP Developer with ThinkPHP expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/php.mdc

Restrictions:
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除
- 遵循 PSR-2 代码风格

Success Criteria:
- {成功标准1}
- 代码通过 PSR-2 格式检查
```

### Vue 前端开发

```
Role: Frontend Developer with Vue 3 + TypeScript expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/vue.mdc

Restrictions:
- 使用 Vue 3 Composition API
- 使用 TypeScript
- 使用 Element Plus 组件库
- 遵循命名规范

Success Criteria:
- {成功标准1}
- 代码通过 ESLint 检查
```

### 测试工程师

```
Role: QA Engineer with {Go/PHP/Vue} testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc 或 .cursor/rules/php.mdc
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
- 活动日志：`docs/team/activities/2025-12/2025-12-05.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-05  
**维护者**: 后端开发组

