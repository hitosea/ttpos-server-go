# ERPNext 对接 - 物品管理增加默认销售单位 任务分解

> 本文档定义 ERPNext 对接 - 物品管理增加默认销售单位的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 9  
**进行中**: -  
**完成率**: 50.0%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_default_sales_unit_to_material_table.php`
  - Purpose: 在物品表中增加 `default_sales_unit` 字段
  - Requirements: 1.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_material 表中增加 default_sales_unit 字段 | Context: 字段类型 bigint unsigned，默认值 0，位置在 cost_unit_uuid 之后，添加索引 idx_default_sales_unit | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中执行迁移，添加字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/material.go`
  - Purpose: 在 Material 结构体中增加 DefaultSalesUnitUuid 字段和关联关系
  - Requirements: 1.1
  - Leverage: 现有 Model: `main/app/model/material.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 Material 结构体中增加 DefaultSalesUnitUuid 字段和 DefaultSalesUnit 关联关系 | Context: 使用 gorm 标签，字段名 default_sales_unit，类型 uint64，默认值 0，添加关联关系 MaterialUnit | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: 核心实现（Go Main）

### DTO 层

- [x] 2.1 更新 Request DTO

  - File: `main/app/dto/req/material.go`
  - Purpose: 在 MaterialCreateReq 和 MaterialUpdateReq 中增加 DefaultSalesUnitUuid 字段
  - Requirements: 4.1, 4.2
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`
  - Prompt: Role: Go Developer | Task: 在 MaterialCreateReq 和 MaterialUpdateReq 中增加 DefaultSalesUnitUuid 字段（可选，指针类型） | Context: 字段名 default_sales_unit_uuid，类型 *uint64，JSON 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [x] 2.2 更新 MaterialEditErpReq DTO

  - File: `main/app/dto/req/material.go`
  - Purpose: 在 MaterialEditErpReq 中增加 DefaultSalesUnit 字段（ERPNext UOM）
  - Requirements: 1.1
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`
  - Prompt: Role: Go Developer | Task: 在 MaterialEditErpReq 中增加 DefaultSalesUnit 字段（string 类型，ERPNext UOM） | Context: 字段名 default_sales_unit，类型 string，JSON 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功

- [x] 2.3 更新 Response DTO

  - File: `main/app/dto/resp/material_resp/material.go`
  - Purpose: 在 MaterialResp 中增加 DefaultSalesUnitUuid 和 DefaultSalesUnit 字段
  - Requirements: 2.1
  - Leverage: 现有 DTO: `main/app/dto/resp/material_resp/material.go`
  - Prompt: Role: Go Developer | Task: 在 MaterialResp 中增加 DefaultSalesUnitUuid 和 DefaultSalesUnit 字段 | Context: DefaultSalesUnitUuid 类型 uint64，DefaultSalesUnit 类型 *MaterialUnitResp（可选） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，响应格式正确

### Service 层

- [x] 2.4 修改 ERPNext 同步逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在 SyncMaterial 和 UpdateMaterialByEprItem 方法中同步默认销售单位字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Service: `main/app/service/material.go`，ERPNext Service: `main/app/service/rpc/erp/material.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 SyncMaterial 和 UpdateMaterialByEprItem 方法，同步 ERPNext 的 DefaultSalesUnit 字段 | Context: 从 ItemInfo 中获取 DefaultSalesUnit（UOM），通过 UOM 查找对应的 MaterialUnit UUID，设置到 Material.DefaultSalesUnitUuid | Restrictions: 遵循 .cursor/rules/go-main.mdc，如果单位不存在则记录日志并设置为 0 | Success: 同步逻辑正确，错误处理完善

- [x] 2.5 修改物品创建逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在 CreateMaterial 方法中支持设置默认销售单位
  - Requirements: 4.1, 4.4
  - Leverage: 现有 Service: `main/app/service/material.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 CreateMaterial 方法，支持设置默认销售单位 | Context: 如果传入了 DefaultSalesUnitUuid，验证该单位是否属于该物品（在物品创建后设置单位时验证），设置到 Material.DefaultSalesUnitUuid | Restrictions: 遵循 .cursor/rules/go-main.mdc，验证单位必须存在 | Success: 创建逻辑正确，验证完善

- [x] 2.6 修改物品更新逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在 UpdateMaterial 方法中支持更新默认销售单位，并实现权限控制
  - Requirements: 3.1, 3.3, 4.2, 4.4
  - Leverage: 现有 Service: `main/app/service/material.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 UpdateMaterial 方法，支持更新默认销售单位，实现权限控制 | Context: 如果传入了 DefaultSalesUnitUuid，检查物品是否为总部来源，如果是总部来源且用户无权限则拒绝修改，验证单位是否属于该物品，更新 Material.DefaultSalesUnitUuid | Restrictions: 遵循 .cursor/rules/go-main.mdc，权限控制正确 | Success: 更新逻辑正确，权限控制完善

- [x] 2.7 修改物品详情响应

  - File: `main/app/service/material.go`
  - Purpose: 在 GetMaterial 方法中返回默认销售单位信息
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有 Service: `main/app/service/material.go`
  - Prompt: Role: Go Developer | Task: 修改 GetMaterial 方法，返回默认销售单位信息 | Context: 使用 Preload 预加载 DefaultSalesUnit 关联关系，在响应中返回 DefaultSalesUnitUuid 和 DefaultSalesUnit 信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 响应包含默认销售单位信息

- [ ] 2.8 编写 Service 单元测试

  - File: `main/app/service/material_test.go`
  - Purpose: 测试默认销售单位相关的业务逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/material_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为默认销售单位功能编写单元测试，覆盖率 ≥ 70% | Context: 测试创建物品时设置默认销售单位，测试更新物品时修改默认销售单位，测试总部物品不允许子店修改，测试验证默认销售单位必须是该物品已配置的单位，测试 ERPNext 同步时处理默认销售单位 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [ ] 2.9 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_material_test.go`
  - Purpose: 测试物品 API 接口（包含默认销售单位）
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/v1/shop/`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为物品 API 编写集成测试，测试默认销售单位功能 | Context: 测试创建物品 API（包含默认销售单位），测试更新物品 API（修改默认销售单位），测试获取物品详情 API（返回默认销售单位），测试权限控制（总部物品只读） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: Vue 前端模块

- [ ] 3.1 更新 API 封装

  - File: `admin/views/shop/api/material.ts`
  - Purpose: 在 API 封装中增加默认销售单位字段的类型定义
  - Requirements: 4.1, 4.2
  - Leverage: 现有 API: `admin/views/shop/api/material.ts`
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 在 Material API 的类型定义中增加 default_sales_unit_uuid 字段 | Context: 在创建和更新接口的请求类型中增加可选字段 default_sales_unit_uuid，在响应类型中增加 default_sales_unit_uuid 和 default_sales_unit 字段 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 类型定义更新成功

- [ ] 3.2 物品详情页显示默认销售单位

  - File: `admin/views/shop/pages/material/detail.vue` 或相关组件
  - Purpose: 在物品详情页的基本信息区域显示"默认销售单位（ERPNext）"字段
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有页面: `admin/views/shop/pages/material/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在物品详情页的基本信息区域添加"默认销售单位（ERPNext）"字段显示 | Context: 有值时显示单位名称，无值时显示"无"，字段标签清晰标识来源为 ERPNext | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Element Plus 组件 | Success: 详情页显示正确

- [ ] 3.3 创建物品表单添加默认销售单位字段

  - File: `admin/views/shop/pages/material/create.vue` 或相关组件
  - Purpose: 在创建物品表单中添加"默认销售单位"字段
  - Requirements: 4.1, 4.3, 4.4
  - Leverage: 现有页面: `admin/views/shop/pages/material/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在创建物品表单中添加"默认销售单位"下拉选择字段 | Context: 下拉选项动态加载该物品的所有单位（基准单位 + 非基准单位），如果只有基准单位则只显示基准单位，提交时包含 default_sales_unit_uuid 字段 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Element Plus Select 组件 | Success: 创建表单功能完整

- [ ] 3.4 编辑物品表单添加默认销售单位字段

  - File: `admin/views/shop/pages/material/edit.vue` 或相关组件
  - Purpose: 在编辑物品表单中添加"默认销售单位"字段，并实现权限控制
  - Requirements: 3.1, 3.2, 4.2, 4.3, 4.4
  - Leverage: 现有页面: `admin/views/shop/pages/material/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在编辑物品表单中添加"默认销售单位"字段，实现权限控制 | Context: 下拉选项动态加载该物品的所有单位，如果物品来源于总部则字段为只读（disabled），提交时包含 default_sales_unit_uuid 字段 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Element Plus Select 组件，权限控制正确 | Success: 编辑表单功能完整，权限控制正确

---

## Phase 4: 测试和优化

- [ ] 4.1 集成测试

  - File: `test/integration/material_default_sales_unit_test.go`
  - Purpose: 测试端到端功能流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试 ERPNext 同步 → 查看详情 → 编辑 → 保存的完整流程，测试创建物品 → 设置默认销售单位 → 查看详情的流程 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 4.3 文档更新

  - File: `docs/shared/api/material_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档（如有），数据库文档，CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

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

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-erpnext-default-sales-unit/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-erpnext-default-sales-unit/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-erpnext-default-sales-unit/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-erpnext-default-sales-unit/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-erpnext-default-sales-unit/tasks.md)" | bc
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
- 活动日志：`docs/team/activities/xiezhihuan/2026-01/2026-01-19.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-19  
**维护者**: xiezhihuan
