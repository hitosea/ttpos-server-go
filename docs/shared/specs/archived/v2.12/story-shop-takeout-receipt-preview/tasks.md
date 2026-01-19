# 新管理端外卖票据预览功能 任务分解

> 本文档定义新管理端外卖票据预览功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 21  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 后端 API 开发（1-2 天）

### DTO 层

- [ ] 1.1 创建 Request DTO

  - File: `main/app/dto/req/printer_req.go`
  - Purpose: 定义预览外卖票据的请求参数
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有 DTO: `main/app/dto/req/printer_req.go`（PreviewPrinterCustomizeReq）
  - Prompt: Role: Go Developer | Task: 在 printer_req.go 中创建 PreviewTakeoutReceiptReq 结构体，包含 template_type 字段（string，必填，枚举值：takeout_customer_receipt, takeout_merchant_receipt） | Context: 复用现有 DTO 风格，使用 binding 标签验证参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc，遵循 .cursor/rules/api.mdc | Success: DTO 创建成功，validation 正确

- [ ] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/printer_resp.go`
  - Purpose: 定义预览外卖票据的响应数据
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有 DTO: `main/app/dto/resp/printer_resp.go`（PreviewPrinterCustomizeResp）
  - Prompt: Role: Go Developer | Task: 在 printer_resp.go 中创建 PreviewTakeoutReceiptResp 结构体，包含 image_url（string）、template_type（string）、is_example_data（bool） | Context: 复用现有 DTO 风格 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，响应格式正确

### 模板文件准备

- [ ] 1.3 创建外卖顾客联模板 JSON

  - File: `main/app/printer/pkg/template/takeout_customer_receipt_tmp.json`
  - Purpose: 定义外卖顾客联的打印模板
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有模板: `main/app/printer/pkg/template/statement_order_tmp.json`（外卖订单模板参考）
  - Prompt: Role: Template Designer | Task: 创建外卖顾客联打印模板 JSON，包含店铺信息、订单信息、商品列表（使用平台商品名）、金额信息 | Context: 参考 statement_order_tmp.json 的结构，使用 {{}} 占位符 | Restrictions: 遵循现有模板格式 | Success: 模板创建成功，占位符正确

- [ ] 1.4 创建外卖商家联模板 JSON

  - File: `main/app/printer/pkg/template/takeout_merchant_receipt_tmp.json`
  - Purpose: 定义外卖商家联的打印模板
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有模板: `main/app/printer/pkg/template/statement_order_tmp.json`（外卖订单模板参考）
  - Prompt: Role: Template Designer | Task: 创建外卖商家联打印模板 JSON，包含店铺信息、订单信息、商品列表（使用店内商品名）、金额信息 | Context: 参考 statement_order_tmp.json 的结构，与顾客联的区别是商品名称字段 | Restrictions: 遵循现有模板格式 | Success: 模板创建成功，占位符正确

- [ ] 1.5 创建示例数据 JSON

  - File: `main/app/printer/pkg/template/takeout_receipt_data.json`
  - Purpose: 定义外卖票据的示例数据（当没有真实订单时使用）
  - Requirements: 4.2, 4.3
  - Leverage: 现有示例数据: `main/app/printer/pkg/template/statement_order_data.json`
  - Prompt: Role: Data Engineer | Task: 创建外卖票据示例数据 JSON，包含完整的订单信息、商品列表（带规格、加料、备注）、金额信息 | Context: 参考 statement_order_data.json 的结构 | Restrictions: 数据真实可信，商品种类丰富 | Success: 示例数据创建成功，数据完整

### Service 层（TakeoutOrderSrv）

- [ ] 1.6 实现 GetLatestOrderForPreview()

  - File: `main/app/service/takeout/takeout_order.go`
  - Purpose: 查询最近的外卖订单（用于预览）
  - Requirements: 4.1, 4.2
  - Leverage: 现有 Service: `main/app/service/takeout/takeout_order.go`，现有 Repository: `main/app/repository/takeout_order_repo.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 takeout_order.go 中实现 GetLatestOrderForPreview() 方法，查询最近 1 条外卖订单（按创建时间倒序），预加载订单商品 | Context: 使用 DBManager，调用 TakeoutOrderRepo，使用 WithOrderItems() 预加载，使用 Limit(1) | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现完整，查询逻辑正确，预加载正确

### Service 层（PrinterSrv）

- [ ] 1.7 实现 GetTakeoutTestData()

  - File: `main/app/service/printer.go`
  - Purpose: 获取外卖订单测试数据（优先真实订单，降级示例数据）
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有 Service: `main/app/service/printer.go`（GetTestData() 方法参考），Task 1.6 的 GetLatestOrderForPreview()，Task 1.5 的示例数据
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 printer.go 中实现 GetTakeoutTestData() 方法，调用 TakeoutOrderSrv.GetLatestOrderForPreview() 获取订单，如果没有订单则读取 takeout_receipt_data.json 示例数据，根据 template_type 处理商品名称（顾客联用平台名，商家联用店内名） | Context: 依赖 TakeoutOrderSrv 接口，使用 convertOrderToTestData() 转换数据，使用 getExampleTakeoutData() 读取示例数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc，降级策略正确 | Success: 方法实现完整，降级逻辑正确，商品名称处理正确

- [ ] 1.8 实现 getExampleTakeoutData()（私有方法）

  - File: `main/app/service/printer.go`
  - Purpose: 读取示例数据 JSON 文件并返回
  - Requirements: 4.2, 4.3
  - Leverage: Task 1.5 的示例数据文件，现有文件读取逻辑
  - Prompt: Role: Go Developer | Task: 在 printer.go 中实现 getExampleTakeoutData() 私有方法，读取 takeout_receipt_data.json 文件，解析为 map[string]interface{}，根据 template_type 调整商品名称字段 | Context: 使用 os.ReadFile() 和 json.Unmarshal()，使用 Redis 缓存示例数据（1 小时过期） | Restrictions: 遵循 .cursor/rules/go-main.mdc，文件不存在时返回错误 | Success: 方法实现完整，文件读取正确，缓存正确

- [ ] 1.9 实现 convertOrderToTestData()（私有方法）

  - File: `main/app/service/printer.go`
  - Purpose: 将外卖订单转换为测试数据格式
  - Requirements: 4.1, 2.2, 3.2
  - Leverage: 现有的 MenuDataRepository（商品名称映射逻辑），`main/app/modules/takeout/infrastructure/persistence/menu_data_repository_impl.go`
  - Prompt: Role: Go Developer | Task: 在 printer.go 中实现 convertOrderToTestData() 私有方法，将 TakeoutOrder 转换为测试数据格式，根据 template_type 处理商品名称（顾客联：GetMenuNamesByPlatformItemIds，商家联：GetProductNamesByUuids） | Context: 提取订单基本信息、商品列表、金额信息，构建测试数据格式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现完整，数据转换正确，商品名称映射正确

- [ ] 1.10 实现 PreviewTakeoutReceipt()

  - File: `main/app/service/printer.go`
  - Purpose: 预览外卖票据（核心业务逻辑）
  - Requirements: 1.1, 2.1, 3.1, 4.1, 5.1
  - Leverage: Task 1.7-1.9 的实现，现有 Service: `main/app/service/printer.go`（ParserConcurrent(), GetTemplateJSONStr()）
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 printer.go 中实现 PreviewTakeoutReceipt() 方法，验证 template_type，调用 GetTakeoutTestData() 获取测试数据，调用 GetTemplateJSONStr() 获取模板 JSON，调用 ParserConcurrent() 解析模板生成 base64 图片，返回 PreviewTakeoutReceiptResp | Context: 依赖 PrinterSrv 的其他方法，使用 DBManager，使用 errors.WithMessage 包装错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现完整，业务逻辑正确，错误处理正确

- [ ] 1.11 扩展 GetTemplateJSONStr() 支持外卖票据类型

  - File: `main/app/service/printer.go`
  - Purpose: 扩展现有方法支持外卖票据模板类型
  - Requirements: 2.1, 3.1
  - Leverage: 现有 Service: `main/app/service/printer.go`（GetTemplateJSONStr() 方法），Task 1.3-1.4 的模板文件
  - Prompt: Role: Go Developer | Task: 在 printer.go 中扩展 GetTemplateJSONStr() 方法，添加对 takeout_customer_receipt 和 takeout_merchant_receipt 模板类型的支持，读取对应的模板 JSON 文件 | Context: 使用 switch 语句判断模板类型，调用 os.ReadFile() 读取模板文件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 扩展成功，模板读取正确

### API 层

- [ ] 1.12 创建 API Controller 方法

  - File: `main/app/api/v1/shop/shop_print.go`
  - Purpose: 实现预览外卖票据 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_print.go`，Task 1.10 的 Service 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 shop_print.go 中创建 PreviewTakeoutReceipt() API 方法，绑定 JSON 参数，调用 PrinterSrv.PreviewTakeoutReceipt()，使用 helper.Success() 返回响应，data 必须是对象 | Context: URL 使用 snake_case，参数验证，错误处理 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 1.13 注册 API 路由

  - File: `main/router/router.go`（或对应的路由文件）
  - Purpose: 注册预览外卖票据 API 路由
  - Requirements: 所有功能需求
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 在路由文件中注册 POST /api/v1/shop/printer_template/preview_takeout_receipt 路由，绑定到 ShopPrintAPI.PreviewTakeoutReceipt 方法，添加身份验证中间件和权限检查中间件 | Context: 使用 router.POST()，添加 middleware.Auth() 和 middleware.Permission("printer_template:view") | Restrictions: 遵循现有路由注册风格 | Success: 路由注册成功，中间件配置正确

### 测试

- [ ] 1.14 编写 Service 单元测试

  - File: `main/app/service/printer_test.go`
  - Purpose: 测试 PrinterSrv 的外卖票据预览功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/printer_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 在 printer_test.go 中添加测试用例：TestPreviewTakeoutReceipt()（测试预览外卖顾客联、商家联、不支持的模板类型、没有订单数据时使用示例数据） | Context: 使用 mock，测试业务逻辑，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 1.15 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_print_test.go`
  - Purpose: 测试预览外卖票据 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/v1/shop/shop_print_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 在 shop_print_test.go 中添加测试用例：TestPreviewTakeoutReceipt()（测试 API 接口，测试参数验证，测试响应格式，测试权限控制） | Context: 使用 httptest，测试 HTTP 请求和响应 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 2: 前端开发（1-2 天）

### API 封装

- [ ] 2.1 封装 API 调用

  - File: `admin/views/shop/src/api/printer.ts`
  - Purpose: 封装预览外卖票据 API 调用
  - Requirements: 所有功能需求
  - Leverage: 现有 API: `admin/views/shop/src/api/printer.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 在 printer.ts 中添加 previewTakeoutReceipt() 方法，调用 POST /api/v1/shop/printer_template/preview_takeout_receipt 接口，定义 TypeScript 类型 PreviewTakeoutReceiptRequest 和 PreviewTakeoutReceiptResponse | Context: 使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

### 页面组件

- [ ] 2.2 票据样式设置页面新增预览入口

  - File: `admin/views/shop/src/views/settings/printer/index.vue`
  - Purpose: 在票据样式设置页面中新增"外卖顾客联"和"外卖商家联"预览入口
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有页面: `admin/views/shop/src/views/settings/printer/index.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在 printer/index.vue 中的票据类型列表中新增"外卖顾客联"和"外卖商家联"两个预览入口，点击时调用 previewTakeoutReceipt() API，打开预览弹窗 | Context: 使用 Element Plus，使用 Composition API，权限控制（检查用户是否有 printer_template:view 权限） | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 入口创建成功，功能正确，权限控制正确

- [ ] 2.3 实现预览弹窗组件

  - File: `admin/views/shop/src/components/printer/TakeoutReceiptPreview.vue`（可选，或直接在页面中实现）
  - Purpose: 实现外卖票据预览弹窗，显示 base64 图片
  - Requirements: 2.1, 3.1, 5.1
  - Leverage: 现有预览组件或弹窗
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建 TakeoutReceiptPreview.vue 组件，显示外卖票据预览弹窗，包含关闭按钮，显示 base64 图片，如果 is_example_data 为 true 则显示"示例数据"标注 | Context: 使用 Element Plus Dialog，使用 Composition API，使用 <img :src="imageUrl" /> 显示图片 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 组件创建成功，功能完整，示例数据标注正确

- [ ] 2.4 前端权限控制

  - File: `admin/views/shop/src/views/settings/printer/index.vue`
  - Purpose: 检查用户权限，无权限用户不显示预览入口
  - Requirements: 1.1, 1.5
  - Leverage: 现有权限控制逻辑
  - Prompt: Role: Frontend Developer | Task: 在 printer/index.vue 中添加权限检查逻辑，使用 v-if 或 computed 判断用户是否有 printer_template:view 权限，无权限时隐藏"外卖顾客联"和"外卖商家联"预览入口 | Context: 使用 Vuex 或 Pinia 获取用户权限 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 权限控制正确，无权限用户不显示入口

---

## Phase 3: 测试和优化（1 天）

### 功能测试

- [ ] 3.1 手动测试：预览外卖顾客联

  - File: -
  - Purpose: 验证外卖顾客联预览功能
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 浏览器开发者工具
  - Success: 预览成功，商品名称使用平台商品名，预览格式正确

- [ ] 3.2 手动测试：预览外卖商家联

  - File: -
  - Purpose: 验证外卖商家联预览功能
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 浏览器开发者工具
  - Success: 预览成功，商品名称使用店内商品名，预览格式正确

- [ ] 3.3 手动测试：无订单数据时使用示例数据

  - File: -
  - Purpose: 验证降级逻辑
  - Requirements: 4.2, 4.3
  - Leverage: 浏览器开发者工具
  - Success: 预览成功，使用示例数据，页面标注"示例数据"

- [ ] 3.4 手动测试：权限控制

  - File: -
  - Purpose: 验证权限控制逻辑
  - Requirements: 1.5, 2.1, 3.1
  - Leverage: 不同权限的测试账号
  - Success: 无权限用户不显示预览入口，有权限用户正常使用

### 性能测试

- [ ] 3.5 性能测试：响应时间

  - File: -
  - Purpose: 确保预览响应时间 < 500ms
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 500ms

### 文档更新

- [ ] 3.6 更新 API 文档

  - File: `docs/shared/api/printer_api.md`
  - Purpose: 更新 API 文档，添加预览外卖票据接口
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档已更新，接口说明完整

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
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
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-receipt-preview/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-receipt-preview/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-takeout-receipt-preview/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-takeout-receipt-preview/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-takeout-receipt-preview/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 后端开发（Service 层）

```
Role: Go Developer specializing in business logic

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Design: 参考 design.md 中的设计方案
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/go-printer.mdc, .cursor/rules/api.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### Vue 前端开发

```
Role: Frontend Developer with Vue 3 + TypeScript expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Design: 参考 design.md 中的设计方案
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
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service)

Test Cases Required:
- 正常场景测试
- 异常场景测试（不支持的模板类型、文件不存在等）
- 边界条件测试（没有订单数据时使用示例数据）
- 权限测试

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
- 活动日志：`docs/team/activities/weifashi/2025-12/2025-12-26.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-26  
**维护者**: 后端开发组

