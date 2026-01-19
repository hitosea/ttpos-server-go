# 新管理端商品管理增加外卖商品模块 任务分解

> 本文档定义外卖商品管理功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 10  
**进行中**: -  
**完成率**: 67%

---

## Phase 1: 数据库设计和迁移 ✅ 已完成

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/20251208232558_create_product_package_takeout_table.php`
  - Purpose: 定义外卖商品表结构
  - Requirements: 1.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`

- [x] 1.2 创建 Go Model

  - File: `main/app/model/product_package_takeout.go`
  - Purpose: 定义 Go 数据模型
  - Requirements: 1.1
  - Leverage: 现有 Model: `main/app/model/product_package.go`

- [x] 1.3 更新 Seeds 文件

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 添加表结构到种子文件
  - Requirements: 1.1

---

## Phase 2: 后端核心实现（Go Main）✅ 已完成

### Repository 层

- [x] 2.1 创建 Repository 接口和实现

  - File: `main/app/repository/product_package_takeout.go`
  - Purpose: 实现数据访问层
  - Requirements: 1.1, 1.2, 2.1-2.4
  - Leverage: 现有 Repository: `main/app/repository/product_package.go`

### DTO 层

- [x] 2.2 创建 Request DTO

  - File: `main/app/dto/req/product_takeout.go`
  - Purpose: 定义 API 请求参数
  - Requirements: 1.2, 2.1-2.3
  - Leverage: 现有 DTO: `main/app/dto/req/product_shop.go`

- [x] 2.3 创建 Response DTO

  - File: `main/app/dto/resp/product_takeout.go`
  - Purpose: 定义 API 响应数据
  - Requirements: 2.2
  - Leverage: 现有 DTO: `main/app/dto/resp/product.go`

### Service 层

- [x] 2.4 创建 Service 接口和实现

  - File: `main/app/service/product_takeout.go`
  - Purpose: 实现业务逻辑
  - Requirements: 1.2, 2.1-2.4
  - Leverage: 现有 Service: `main/app/service/product.go`

### API 层

- [x] 2.5 创建 API Handler

  - File: `main/app/api/v1/shop/shop_product_takeout.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: 1.2, 2.1-2.4
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go`

- [x] 2.6 注册 API 路由

  - File: `main/router/shop.go`
  - Purpose: 注册外卖商品 API 路由
  - Requirements: 1.2, 2.1-2.4

- [x] 2.7 添加外卖类型常量

  - File: `main/app/constant/product.go`
  - Purpose: 定义外卖平台类型常量
  - Requirements: 多平台支持

---

## Phase 3: Vue 前端模块 🚧 待开发

### API 封装

- [ ] 3.1 创建前端 API 封装

  - File: `admin/views/shop/api/product_takeout.ts`
  - Purpose: 封装后端 API 调用
  - Requirements: 1.3, 2.5
  - Leverage: 现有 API: `admin/views/shop/api/product.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装外卖商品 API 调用 | Context: 使用 axios，定义 TypeScript 类型，包含 add, edit, detail, delete, status 接口

### 页面组件

- [ ] 3.2 商品添加页面增加外卖 Tab

  - File: `admin/views/shop/pages/product/add.vue` (修改)
  - Purpose: 在商品添加页面增加外卖 Tab 切换
  - Requirements: 1.3
  - Leverage: 现有页面结构
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在商品添加页面增加外卖 Tab | Context: 使用 Element Plus Tabs，切换到外卖 Tab 时显示外卖配置表单

- [ ] 3.3 商品编辑页面增加外卖 Tab

  - File: `admin/views/shop/pages/product/edit.vue` (修改)
  - Purpose: 在商品编辑页面增加外卖 Tab 切换
  - Requirements: 2.5
  - Leverage: 现有页面结构
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在商品编辑页面增加外卖 Tab | Context: 使用 Element Plus Tabs，切换到外卖 Tab 时加载已有的外卖配置

- [ ] 3.4 外卖商品表单组件

  - File: `admin/views/shop/components/product/TakeoutForm.vue`
  - Purpose: 外卖商品配置表单组件
  - Requirements: 1.3, 1.4, 2.5
  - Leverage: 现有组件: `admin/views/shop/components/product/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 创建外卖商品配置表单 | Context: 包含外卖类型选择、分类选择、状态开关、图片上传等字段，使用 Element Plus 表单组件

---

## Phase 4: 测试 ⏳ 待开发

- [ ] 4.1 编写 Repository 单元测试

  - File: `main/app/repository/product_package_takeout_test.go`
  - Purpose: 测试数据访问层
  - Requirements: 测试验收
  - Leverage: 现有测试: `main/app/repository/*_test.go`

- [ ] 4.2 编写 Service 单元测试

  - File: `main/app/service/product_takeout_test.go`
  - Purpose: 测试业务逻辑层
  - Requirements: 测试验收
  - Leverage: 现有测试: `main/app/service/*_test.go`

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 后端代码通过 `go fmt` 和 `go vet`
- [ ] 前端代码通过 ESLint 检查
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [x] requirements.md 中的后端需求已满足
- [ ] requirements.md 中的前端需求已满足
- [x] design.md 中的后端设计已实现
- [ ] design.md 中的前端设计已实现

### 文档同步

- [x] API 接口已在 requirements.md 中记录
- [x] 数据库表已在 requirements.md 中记录

### 规范遵循

- [x] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`（前端）
- [x] 遵循 `.cursor/rules/api.mdc`
- [x] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-grab-integration/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-grab-integration/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-grab-integration/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **实现代码**: 按照规范实现功能
5. **运行检查**: `go fmt`, `go vet`, `go test` / ESLint
6. **标记完成**: 将 `[ ]` 改为 `[x]`
7. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-09  
**维护者**: weifashi

