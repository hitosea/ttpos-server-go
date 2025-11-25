# 新管理端-物品可见性 任务分解

> 本文档定义新管理端-物品可见性功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 52  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_allow_substore_visible_to_material_table.php`
  - Purpose: 在 ttpos_material 表中添加 allow_substore_visible 字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `admin/database/migrations/20250919095649_add_warehouse_item_table.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_material 表中添加 allow_substore_visible 字段（tinyint(1)，默认值 1） | Context: 必须遵循 .cursor/rules/database.mdc，字段名使用 snake_case，添加索引 idx_allow_substore_visible | Restrictions: 迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 1.2
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/material.go`
  - Purpose: 在 Material 结构体中添加 AllowSubstoreVisible 字段
  - Requirements: 1.3
  - Leverage: 现有 Model: `main/app/model/material.go`
  - Prompt: Role: Go Developer | Task: 在 Material 结构体中添加 AllowSubstoreVisible 字段，映射到 allow_substore_visible 列 | Context: 使用 gorm 标签，类型为 int，默认值 1 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

- [ ] 1.4 更新 PHP Model

  - File: `admin/app/common/model/product/Material.php`
  - Purpose: 在 PHP Material 模型中添加对应字段
  - Requirements: 1.4
  - Leverage: 现有 Model: `admin/app/common/model/product/Material.php`
  - Prompt: Role: PHP Developer | Task: 在 Material 模型中添加 allowSubstoreVisible 字段 | Context: 字段类型 tinyint，默认值 1 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [x] 2.1 添加可见性过滤选项方法

  - File: `main/app/repository/material_repo.go`
  - Purpose: 添加 WhereAllowSubstoreVisible 选项方法
  - Requirements: 3.1-3.10
  - Leverage: 现有 Repository: `main/app/repository/material_repo.go`，使用选项模式
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 MaterialRepoImpl 中添加 WhereAllowSubstoreVisible 选项方法 | Context: 使用选项模式(DBOption)，返回过滤条件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 选项方法添加成功

### Service 层

- [x] 2.2 修改同步逻辑，同步可见性字段

  - File: `main/app/service/material.go`
  - Purpose: 在 SyncMaterial 方法中同步 allow_substore_visible 字段
  - Requirements: 2.1, 2.2
  - Leverage: 现有同步逻辑: `main/app/service/material.go` 的 `SyncMaterial` 方法（第 2778 行）
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 SyncMaterial 方法，在同步总部物品到子店时，同步 allow_substore_visible 字段 | Context: 在同步逻辑中，将总部的 allow_substore_visible 值同步到子店 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有同步逻辑 | Success: 同步逻辑修改成功，可见性字段正确同步

- [x] 2.3 添加可见性过滤逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在物品查询方法中添加可见性过滤逻辑（子店自动过滤）
  - Requirements: 3.1-3.10
  - Leverage: 现有 Service: `main/app/service/material.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 在物品查询方法中添加可见性过滤逻辑，子店查询时自动过滤 allow_substore_visible = 0 的物品 | Context: 使用 companySetting.IsSubShop() 判断是否为子店，子店查询时添加过滤条件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 过滤逻辑添加成功，子店查询正确过滤

### API 层

- [x] 2.4 修改物品列表查询接口，自动应用可见性过滤

  - File: `main/app/api/material_api.go`
  - Purpose: 物品列表查询接口自动应用可见性过滤（子店）
  - Requirements: 3.1
  - Leverage: 现有 API: `main/app/api/material_api.go`，Task 2.3 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 修改物品列表查询接口，子店查询时自动过滤不可见物品 | Context: Service 层已添加过滤逻辑，API 层无需额外处理 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 接口正确应用过滤

- [ ] 2.5 添加物品可见性更新接口（总店）

  - File: `main/app/api/material_api.go`
  - Purpose: 添加更新物品可见性设置的接口（仅总店可用）
  - Requirements: 1.5
  - Leverage: 现有 API: `main/app/api/material_api.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 添加更新物品可见性设置的接口，检查用户权限（仅总店可用） | Context: URL: /api/v1/material/update_visible，POST 方法，参数: uuid, allow_substore_visible | Restrictions: 遵循 .cursor/rules/api.mdc，检查权限 | Success: API 接口创建成功，权限检查正确

- [x] 2.6 添加批量更新物品可见性接口（总店）

  - File: `main/app/api/material_api.go`
  - Purpose: 添加批量更新物品可见性设置的接口（仅总店可用）
  - Requirements: 4.1, 4.2
  - Leverage: 现有 API: `main/app/api/material_api.go`
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 添加批量更新物品可见性设置的接口，检查用户权限（仅总店可用） | Context: URL: /api/v1/material/batch_update_visible，POST 方法，参数: uuids[], allow_substore_visible | Restrictions: 遵循 .cursor/rules/api.mdc，检查权限 | Success: API 接口创建成功，批量更新正确

---

## Phase 3: PHP Admin 模块

- [ ] 3.1 更新 Material Service，添加可见性更新方法

  - File: `admin/app/{admin|shop}/service/MaterialService.php`
  - Purpose: 添加更新物品可见性的业务逻辑方法
  - Requirements: 1.5
  - Leverage: 现有 Service: `admin/app/{admin|shop}/service/MaterialService.php`
  - Success: Service 方法添加成功

- [ ] 3.2 更新 Material Controller，添加可见性更新接口

  - File: `admin/app/{admin|shop}/controller/MaterialController.php`
  - Purpose: 添加更新物品可见性的控制器方法
  - Requirements: 1.5
  - Leverage: 现有 Controller: `admin/app/{admin|shop}/controller/MaterialController.php`
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 MaterialController 中添加更新物品可见性的方法 | Context: 调用 Service 层方法，检查权限（仅总店可用） | Restrictions: 遵循 .cursor/rules/php.mdc，Controller 不写业务逻辑 | Success: Controller 方法添加成功

- [ ] 3.3 添加批量更新可见性接口

  - File: `admin/app/{admin|shop}/controller/MaterialController.php`
  - Purpose: 添加批量更新物品可见性的控制器方法
  - Requirements: 4.1, 4.2
  - Leverage: 现有 Controller: `admin/app/{admin|shop}/controller/MaterialController.php`
  - Success: Controller 方法添加成功

---

## Phase 4: Vue 前端模块

- [ ] 4.1 在物品管理页面添加可见性设置开关

  - File: `admin/views/{admin|shop}/pages/material/index.vue`
  - Purpose: 在物品列表或编辑页面添加"允许子店可见"开关控件
  - Requirements: 1.5, 1.6, 1.7
  - Leverage: 现有页面: `admin/views/{admin|shop}/pages/material/index.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在物品管理页面添加"允许子店可见"开关，仅总店显示 | Context: 使用 Element Plus Switch 组件，默认开启，仅总店可见 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 开关控件添加成功，权限控制正确

- [ ] 4.2 添加操作提示弹窗

  - File: `admin/views/{admin|shop}/pages/material/index.vue`
  - Purpose: 在开启/关闭可见性时显示确认弹窗
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: 现有页面: `admin/views/{admin|shop}/pages/material/index.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 添加确认弹窗，开启/关闭时显示影响范围提示 | Context: 使用 Element Plus MessageBox，提示内容明确告知同步前后状态差异 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 弹窗添加成功，提示内容正确

- [ ] 4.3 添加批量操作功能

  - File: `admin/views/{admin|shop}/pages/material/index.vue`
  - Purpose: 添加批量设置可见/不可见功能
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有页面: `admin/views/{admin|shop}/pages/material/index.vue`，参考批量停用/开启功能
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 添加批量设置可见/不可见功能，逻辑与批量停用/开启保持一致 | Context: 使用表格多选，批量操作按钮，确认提示 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 批量操作功能添加成功

- [ ] 4.4 添加筛选功能

  - File: `admin/views/{admin|shop}/pages/material/index.vue`
  - Purpose: 添加按可见性筛选物品的功能
  - Requirements: 4.3
  - Leverage: 现有页面: `admin/views/{admin|shop}/pages/material/index.vue`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 添加筛选功能，可按"允许子店可见"和"不允许子店可见"筛选物品 | Context: 使用 Element Plus Select 组件，筛选条件传递给后端 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 筛选功能添加成功

- [ ] 4.5 更新 API 封装

  - File: `admin/views/{admin|shop}/api/material.ts`
  - Purpose: 封装更新物品可见性的 API 调用
  - Requirements: 1.5, 4.1
  - Leverage: 现有 API: `admin/views/{admin|shop}/api/material.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 封装更新物品可见性的 API 调用 | Context: 定义 TypeScript 类型，使用 axios 调用接口 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

---

## Phase 5: 业务模块可见性过滤

- [x] 5.1 修改成本卡相关接口，添加可见性过滤

  - File: `main/app/api/cost_card_api.go` 或相关文件
  - Purpose: 子店查询成本卡时过滤不可见物品
  - Requirements: 3.2
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 成本卡接口过滤成功

- [x] 5.2 修改采购申请相关接口，添加可见性过滤

  - File: `main/app/api/purchase_apply_api.go` 或相关文件
  - Purpose: 子店查询采购申请时过滤不可见物品
  - Requirements: 3.3
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 采购申请接口过滤成功

- [x] 5.3 修改采购收货相关接口，添加可见性过滤

  - File: `main/app/api/purchase_receive_api.go` 或相关文件
  - Purpose: 子店查询采购收货时过滤不可见物品
  - Requirements: 3.4
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 采购收货接口过滤成功

- [x] 5.4 修改品牌采购相关接口，添加可见性过滤

  - File: `main/app/api/brand_purchase_api.go` 或相关文件
  - Purpose: 子店查询品牌采购时过滤不可见物品
  - Requirements: 3.5
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 品牌采购接口过滤成功（已通过采购申请接口覆盖）

- [x] 5.5 修改品采收货相关接口，添加可见性过滤

  - File: `main/app/api/brand_purchase_receive_api.go` 或相关文件
  - Purpose: 子店查询品采收货时过滤不可见物品
  - Requirements: 3.6
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 品采收货接口过滤成功（已通过采购收货接口覆盖）

- [x] 5.6 修改盘点单相关接口，添加可见性过滤

  - File: `main/app/api/inventory_check_api.go` 或相关文件
  - Purpose: 子店查询盘点单时过滤不可见物品
  - Requirements: 3.7
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 盘点单接口过滤成功

- [x] 5.7 修改调拨单相关接口，添加可见性过滤

  - File: `main/app/api/transfer_api.go` 或相关文件
  - Purpose: 子店查询调拨单时过滤不可见物品
  - Requirements: 3.8
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 调拨单接口过滤成功

- [x] 5.8 修改库存查询相关接口，添加可见性过滤

  - File: `main/app/api/stock_query_api.go` 或相关文件
  - Purpose: 子店查询库存时过滤不可见物品
  - Requirements: 3.9
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 库存查询接口过滤成功

- [x] 5.9 修改出入库明细表相关接口，添加可见性过滤

  - File: `main/app/api/stock_detail_api.go` 或相关文件
  - Purpose: 子店查询出入库明细时过滤不可见物品
  - Requirements: 3.10
  - Leverage: 现有接口，Task 2.3 的过滤逻辑
  - Success: 出入库明细接口过滤成功

---

## Phase 6: 测试和优化

- [ ] 6.1 编写 Repository 单元测试

  - File: `main/app/repository/material_repo_test.go`
  - Purpose: 测试可见性过滤选项方法
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 MaterialRepo 的可见性过滤选项方法编写单元测试 | Context: 测试 WhereAllowSubstoreVisible 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 6.2 编写 Service 单元测试

  - File: `main/app/service/material_srv_test.go`
  - Purpose: 测试同步逻辑和可见性过滤逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 MaterialSrv 的同步逻辑和可见性过滤逻辑编写单元测试 | Context: 测试同步可见性字段，测试子店过滤逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 6.3 编写 API 集成测试

  - File: `main/app/api/material_api_test.go`
  - Purpose: 测试物品可见性相关 API
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为物品可见性相关 API 编写集成测试 | Context: 测试总店设置可见性，测试子店查询过滤，测试批量操作 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 6.4 集成测试

  - File: `test/integration/material_visibility_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试总店设置可见性 → 子店同步 → 子店查询过滤的完整流程 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 6.5 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 6.6 数据库查询优化

  - File: `main/app/repository/material_repo.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 6.7 文档更新

  - File: `docs/shared/api/material_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, 数据库文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

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

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

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
grep -c "^- \[" docs/shared/specs/active/story-admin-item-visibility/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-item-visibility/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-item-visibility/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-item-visibility/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-item-visibility/tasks.md)" | bc
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
**最后更新**: 2025-11-21  
**维护者**: 后端开发组

