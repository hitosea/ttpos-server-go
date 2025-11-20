# 订单数据导入 任务分解

> 本文档定义订单数据导入功能的详细执行任务清单。
> 
> **💡 MVP 方案**：最小可执行版本，快速验证可行性

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 16  
**已完成**: 1  
**进行中**: -  
**完成率**: 6.3%

**优先级分布**：
- **P0 核心任务**: 6 个（必须实现，核心功能）🔥
- **P1 重要任务**: 5 个（重要功能，完善体验）⚡
- **P2 可选任务**: 5 个（后续扩展，优化体验）🎨

**当前阶段**: P0 核心任务（Excel 解析 + 数据库写入）

---

## 🎯 任务优先级说明

- **P0 - 核心任务**：必须实现，核心功能（Excel 解析 + 数据库写入）
- **P1 - 重要任务**：重要功能，完善体验（DTO、路由、基础测试）
- **P2 - 可选任务**：后续扩展，优化体验（前端、PHP、完整测试）

**实现顺序**：先完成所有 P0 任务，再实现 P1，最后实现 P2

---

## Phase 1: 需求确认和 Excel 格式定义

- [x] 1.1 确认 Excel 格式规范 ✅

  - File: `docs/shared/specs/story-order-import-data/excel-format.md`
  - Purpose: 与华莱士确认 Excel 文件格式，定义字段映射关系
  - Requirements: Requirement 2
  - Leverage: 参考现有订单数据结构
  - Success: Excel 格式规范已定义，包含字段映射、校验规则和示例数据

---

## 🔥 P0 - 核心任务：Excel 解析和数据库写入

### 核心目标
实现 Excel 文件解析、数据校验和批量写入数据库的核心逻辑，确保数据能够成功导入。

### Service 层（核心业务逻辑）

- [ ] **P0-1** 创建 Service 接口

  - File: `main/app/service/i_order_import_service.go`
  - Purpose: 定义订单导入业务逻辑接口
  - Requirements: Requirement 2, Requirement 3
  - Leverage: 现有 Service 接口: `main/app/service/i_*_service.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 IOrderImportSrv 接口，定义 Import 方法 | Context: 接收文件，返回导入结果 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确
  - Success: 接口定义完整，方法签名正确

- [ ] **P0-2** 实现 Excel 解析逻辑

  - File: `main/app/service/order_import_service.go` (部分)
  - Purpose: 解析 Excel 文件，提取订单基本信息和明细数据
  - Requirements: Requirement 2
  - Leverage: 使用 `github.com/xuri/excelize/v2` 库
  - Prompt: Role: Go Developer with Excel parsing expertise | Task: 实现 Excel 解析函数，读取 Sheet1（订单基本信息）和 Sheet2（订单明细） | Context: 使用 excelize 库，解析表头和数据行，跳过空行 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 能够正确解析 Excel 文件，提取所有字段数据
  - Success: 能够正确解析 Excel 文件，提取所有字段数据

- [ ] **P0-3** 实现数据校验逻辑

  - File: `main/app/service/order_import_service.go` (部分)
  - Purpose: 校验必填字段、数据格式和关联数据存在性
  - Requirements: Requirement 2
  - Leverage: 现有订单相关 Service（Shop、Product、Member）
  - Prompt: Role: Go Developer with validation expertise | Task: 实现数据校验函数，校验必填字段、日期格式、金额格式、门店/商品/会员是否存在 | Context: 校验订单基本信息和订单明细，返回详细的错误信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 能够正确校验所有数据，返回明确的错误提示
  - Success: 能够正确校验所有数据，返回明确的错误提示

- [ ] **P0-4** 实现数据库写入逻辑

  - File: `main/app/service/order_import_service.go` (部分)
  - Purpose: 将校验通过的数据批量写入数据库
  - Requirements: Requirement 3
  - Leverage: 现有订单相关 Repository 和 Service
  - Prompt: Role: Go Developer with database expertise | Task: 实现数据库写入函数，使用事务批量写入订单数据 | Context: 创建 SaleBill、SaleOrder、SaleOrderProduct 记录，使用事务保证一致性，订单号重复则跳过 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务管理 | Success: 能够正确写入数据库，事务管理正确，数据一致性保证
  - Success: 能够正确写入数据库，事务管理正确，数据一致性保证

- [ ] **P0-5** 整合 Service 业务逻辑

  - File: `main/app/service/order_import_service.go`
  - Purpose: 整合 Excel 解析、数据校验和数据库写入逻辑
  - Requirements: Requirement 2, Requirement 3
  - Leverage: Task P0-2, P0-3, P0-4 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 整合所有核心逻辑，实现完整的 Import 方法 | Context: 调用解析、校验、写入函数，处理错误，返回导入结果（成功数量、失败数量、失败列表） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，错误处理完善
  - Success: Service 实现完整，业务逻辑正确，错误处理完善

- [ ] **P0-6** 创建基础 API 接口（最小可用）

  - File: `main/app/api/order_import_api.go`
  - Purpose: 实现文件上传和导入 HTTP API 接口（最小可用版本）
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有 API: `main/app/api/*_api.go`，Task P0-5 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 OrderImportAPI，实现文件上传接口（最小可用版本） | Context: URL 使用 snake_case (/api/v1/order/import)，使用 helper.Success() 返回响应，data 必须是对象，校验文件格式和大小 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确
  - Success: API 创建成功，响应格式正确，参数验证正确

---

## ⚡ P1 - 重要任务：完善功能和体验

### DTO 层

- [ ] **P1-1** 创建 Request DTO

  - File: `main/app/dto/req/order_import_req.go`
  - Purpose: 定义文件上传请求参数
  - Requirements: Requirement 1
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 OrderImportReq，包含文件上传字段 | Context: 使用 form 标签，支持 multipart/form-data | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确
  - Success: DTO 创建成功，validation 正确

- [ ] **P1-2** 创建 Response DTO

  - File: `main/app/dto/resp/order_import_resp.go`
  - Purpose: 定义导入结果响应数据
  - Requirements: Requirement 4
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 OrderImportResp，包含成功数量、失败数量和失败列表 | Context: data 必须是对象，不能是 null 或数组 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: DTO 创建成功，响应格式正确
  - Success: DTO 创建成功，响应格式正确

### API 层

- [ ] **P1-3** 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册订单导入 API 路由
  - Requirements: Requirement 1
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功

- [ ] **P1-4** 完善 API 接口（使用 DTO）

  - File: `main/app/api/order_import_api.go`
  - Purpose: 使用 DTO 完善 API 接口
  - Requirements: Requirement 1, Requirement 4
  - Leverage: Task P1-1, P1-2 的 DTO
  - Success: API 接口完善，使用 DTO 进行参数和响应处理

### 测试

- [ ] **P1-5** 编写 Service 单元测试（核心逻辑）

  - File: `main/app/service/order_import_service_test.go`
  - Purpose: 确保 Service 核心业务逻辑正确
  - Requirements: Requirement 2, Requirement 3
  - Leverage: 现有测试: `main/app/service/*_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 OrderImportSrv 编写单元测试，重点测试 Excel 解析、数据校验、数据库写入 | Context: 测试 Excel 解析，测试数据校验，测试批量导入，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 核心逻辑测试通过，覆盖率 ≥ 70%
  - Success: 核心逻辑测试通过，覆盖率 ≥ 70%

---

## 🎨 P2 - 可选任务：前端和完整测试

### PHP Admin 模块

- [ ] **P2-1** 创建 PHP Controller

  - File: `admin/app/shop/controller/store/order/ImportController.php`
  - Purpose: 实现后台管理文件上传接口
  - Requirements: Requirement 1
  - Leverage: 现有 Controller: `admin/app/shop/controller/`
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 创建 ImportController，实现文件上传接口，调用 Go Main API | Context: 遵循 MVC 分层，Controller 不写业务逻辑 | Restrictions: 遵循 .cursor/rules/php.mdc，使用验证器 | Success: Controller 创建成功，接口正确
  - Success: Controller 创建成功，接口正确

- [ ] **P2-2** 创建 PHP Service

  - File: `admin/app/shop/service/order/ImportService.php`
  - Purpose: 调用 Go Main API
  - Requirements: Requirement 1
  - Leverage: 现有 Service: `admin/app/shop/service/`
  - Success: Service 创建成功

### Vue 前端模块

- [ ] **P2-3** 创建导入页面

  - File: `admin/views/shop/src/views/order/import/index.vue`
  - Purpose: 实现导入页面 UI
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有页面: `admin/views/shop/src/views/product/store/product/importProduct.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建订单导入页面，实现文件上传和结果展示 | Context: 使用 Element Plus，使用 Composition API，参考商品导入页面 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 页面创建成功，功能完整
  - Success: 页面创建成功，功能完整

- [ ] **P2-4** 创建 API 封装

  - File: `admin/views/shop/src/api/order/import.ts`
  - Purpose: 封装后端 API 调用
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有 API: `admin/views/shop/src/api/`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装订单导入 API 调用 | Context: 使用 axios，定义 TypeScript 类型 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确
  - Success: API 封装完成，类型定义正确

### 完整测试

- [ ] **P2-5** 编写 API 集成测试

  - File: `main/app/api/order_import_api_test.go`
  - Purpose: 测试 API 接口
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 OrderImportAPI 编写集成测试 | Context: 测试文件上传接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过
  - Success: 所有 API 测试通过

- [ ] **P2-6** 集成测试和性能测试

  - File: `test/integration/order_import_test.go`（可选）
  - Purpose: 测试端到端功能和性能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试和性能测试 | Context: 测试完整导入流程，测试数据一致性，测试性能（5000条 < 30秒） | Restrictions: 测试真实用户场景 | Success: 集成测试通过，性能达标
  - Success: 集成测试通过，性能达标

---

## Phase 2: 后端核心实现（Go Main）- 已按优先级重组

> ⚠️ **注意**：Phase 2 的任务已按优先级重组到上面的 P0/P1/P2 部分，此处保留作为参考。

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
  - Order: 100%
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
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-order-import-data/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-order-import-data/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-order-import-data/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-order-import-data/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-order-import-data/tasks.md)" | bc
```

### 执行流程

**优先级执行顺序**：

1. **P0 核心任务**（必须完成）：
   - 先实现 P0-1 到 P0-6，确保核心功能可用
   - 可以使用临时 DTO 和简单 API，先让功能跑通
   - 重点：Excel 解析 + 数据校验 + 数据库写入

2. **P1 重要任务**（完善体验）：
   - 完成 P1-1 到 P1-5，使用 DTO，完善 API，添加测试
   - 确保代码质量和规范性

3. **P2 可选任务**（后续扩展）：
   - 完成 P2-1 到 P2-6，添加前端、PHP 接口和完整测试
   - 优化用户体验和系统完整性

**每个任务的执行步骤**：

1. **选择任务**: 按优先级选择下一个未完成任务
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

---

## 📋 快速开始（核心任务）

### 第一步：实现 Service 接口和核心逻辑

```bash
# 1. 创建 Service 接口
# File: main/app/service/i_order_import_service.go

# 2. 实现 Excel 解析
# File: main/app/service/order_import_service.go
# - 使用 excelize 解析 Excel
# - 提取订单基本信息和明细数据

# 3. 实现数据校验
# - 校验必填字段
# - 校验数据格式
# - 校验关联数据（门店、商品、会员）

# 4. 实现数据库写入
# - 使用事务批量写入
# - 创建 SaleBill、SaleOrder、SaleOrderProduct 记录
```

### 第二步：创建最小可用 API

```bash
# 1. 创建基础 API（可以使用临时结构体，不用 DTO）
# File: main/app/api/order_import_api.go

# 2. 注册路由
# File: main/router/router.go
```

### 第三步：测试核心功能

```bash
# 使用 Postman 或 curl 测试 API
# 上传示例 Excel 文件
# 验证数据是否正确写入数据库
```

---

**模板版本**: v2.0.0（按优先级重组）  
**最后更新**: 2025-11-19  
**维护者**: xiezhihuan

