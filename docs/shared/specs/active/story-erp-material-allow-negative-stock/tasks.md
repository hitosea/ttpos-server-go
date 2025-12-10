# 物品负库存控制 任务分解

> 本文档定义物品负库存控制功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 8  
**进行中**: -  
**完成率**: 44%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_allow_negative_stock_to_material_table.php`
  - Purpose: 在 `ttpos_material` 表中添加 `allow_negative_stock` 字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有迁移文件: `admin/database/migrations/20251121081848_add_allow_substore_visible_to_material_table.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_material 表中添加 allow_negative_stock 字段 | Context: 字段类型 INT(1) NOT NULL DEFAULT 0，注释"是否允许负库存：1-允许，0-不允许"，添加在 origin_country_code 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/material.go`
  - Purpose: 在 Material 结构体中添加 AllowNegativeStock 字段
  - Requirements: 1.3
  - Leverage: 现有 Model: `main/app/model/material.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 Material 结构体中添加 AllowNegativeStock 字段 | Context: 字段类型 int，gorm 标签 column:allow_negative_stock，default:0，comment:'是否允许负库存：1-允许，0-不允许' | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

- [ ] 1.4 更新 PHP Model（可选）

  - File: `admin/app/common/model/product/Material.php`
  - Purpose: 在 PHP Material 模型中添加 allow_negative_stock 字段
  - Requirements: 1.4
  - Leverage: 现有 Model: `admin/app/common/model/product/Material.php`
  - Success: PHP Model 更新成功

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [x] 2.1 添加 Repository 方法

  - File: `main/app/repository/i_material_repo.go`, `main/app/repository/material_repo.go`
  - Purpose: 添加 UpdateMaterialAllowNegativeStock 方法
  - Requirements: 3.2
  - Leverage: 现有 Repository: `main/app/repository/material_repo.go`，参考 `UpdateMaterialAllowSubstoreVisible` 方法
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 IMaterialRepo 接口和 MaterialRepoImpl 实现中添加 UpdateMaterialAllowNegativeStock 方法 | Context: 方法签名 func UpdateMaterialAllowNegativeStock(uuid uint64, allowNegativeStock bool) error，将 bool 转换为 int（true=1, false=0）后更新数据库 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 方法添加成功，实现正确

### DTO 层

- [x] 2.2 更新 MaterialEditErpReq

  - File: `main/app/dto/req/material.go`
  - Purpose: 在 MaterialEditErpReq 结构体中添加 AllowNegativeStock 字段
  - Requirements: 3.3
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`，参考 MaterialAddErpReq 中的 AllowNegativeStock 字段
  - Prompt: Role: Go Developer | Task: 在 MaterialEditErpReq 结构体中添加 AllowNegativeStock *bool 字段 | Context: 字段类型 *bool，json 标签 allow_negative_stock | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [x] 2.3 验证 MaterialAddReq 和 MaterialEditReq

  - File: `main/app/dto/req/material.go`
  - Purpose: 验证 AllowNegativeStock 字段是否存在且类型正确
  - Requirements: 2.1, 3.1
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`
  - Success: 字段已存在且类型正确（MaterialAddReq: *bool, MaterialEditReq: bool）

### Service 层

- [x] 2.4 更新 AddMaterial Service 方法

  - File: `main/app/service/material.go`
  - Purpose: 确保 AddMaterial 方法正确保存 AllowNegativeStock 字段
  - Requirements: 2.2
  - Leverage: 现有 Service: `main/app/service/material.go`，addMaterial 函数
  - Prompt: Role: Go Developer with business logic expertise | Task: 验证 AddMaterial 方法中 AllowNegativeStock 字段的保存逻辑 | Context: 检查 addMaterial 函数中是否正确将 AllowNegativeStock 保存到数据库，检查 material 结构体是否包含该字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 保存逻辑正确，字段被正确写入数据库

- [x] 2.5 更新 EditMaterial Service 方法

  - File: `main/app/service/material.go`
  - Purpose: 在 EditMaterial 方法中添加 AllowNegativeStock 字段的更新逻辑
  - Requirements: 3.2
  - Leverage: 现有 Service: `main/app/service/material.go`，参考 UpdateMaterialAllowSubstoreVisible 的调用方式
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 EditMaterial 方法中添加 AllowNegativeStock 字段的更新逻辑 | Context: 在更新物品后，调用 materialRepo.UpdateMaterialAllowNegativeStock 方法更新字段，将 bool 值转换为 int（true=1, false=0） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 更新逻辑正确，字段被正确更新

- [x] 2.6 更新 EditMaterial ERP 同步逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在 EditMaterial 的 ERP 同步逻辑中添加 AllowNegativeStock 字段
  - Requirements: 3.4
  - Leverage: 现有 Service: `main/app/service/material.go`，EditMaterial 方法中的 ERP 同步部分
  - Prompt: Role: Go Developer | Task: 在 EditMaterial 方法的 ERP 同步逻辑中，将 AllowNegativeStock 字段添加到 MaterialEditErpReq | Context: 在构建 MaterialEditErpReq 时，添加 AllowNegativeStock: &request.AllowNegativeStock（需要将 bool 转换为 *bool） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: ERP 同步逻辑更新成功，字段正确传递

- [x] 2.7 更新 UpdateMaterialByEprItem Service 方法

  - File: `main/app/service/material.go`
  - Purpose: 在 UpdateMaterialByEprItem 方法中添加 AllowNegativeStock 字段的同步逻辑
  - Requirements: 4.1
  - Leverage: 现有 Service: `main/app/service/material.go`，UpdateMaterialByEprItem 方法
  - Prompt: Role: Go Developer | Task: 在 UpdateMaterialByEprItem 方法中添加 AllowNegativeStock 字段的更新逻辑 | Context: 在 updateData map 中添加 allow_negative_stock 字段，将 *bool 转换为 int（true=1, false=0, nil=0） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 同步逻辑更新成功，字段正确保存

### Response DTO 层

- [x] 2.8 更新 Material Response DTO

  - File: `main/app/dto/resp/material_resp/material.go`
  - Purpose: 在响应结构体中添加 AllowNegativeStock 字段
  - Requirements: 2.3, 3.5
  - Leverage: 现有 Response DTO: `main/app/dto/resp/material_resp/material.go`
  - Prompt: Role: Go Developer | Task: 在 Material 响应结构体中添加 AllowNegativeStock 字段 | Context: 字段类型 int，json 标签 allow_negative_stock | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Response DTO 更新成功，字段正确返回

---

## Phase 3: PHP Admin 模块（可选）

- [ ] 3.1 更新 PHP Model

  - File: `admin/app/common/model/product/Material.php`
  - Purpose: 在 PHP Material 模型中添加 allow_negative_stock 字段
  - Requirements: 1.4
  - Leverage: 现有 Model: `admin/app/common/model/product/Material.php`
  - Success: PHP Model 更新成功

---

## Phase 4: Vue 前端模块

- [ ] 4.1 更新添加物品表单

  - File: `admin/views/shop/pages/material/add.vue` 或相关组件
  - Purpose: 在添加物品表单中添加"允许负库存"开关组件
  - Requirements: 5.1
  - Leverage: 现有页面: `admin/views/shop/pages/material/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在添加物品表单中添加"允许负库存"开关组件 | Context: 使用 Element Plus 的 el-switch 组件，绑定到 allow_negative_stock 字段，添加字段说明文案 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 表单更新成功，开关组件正确绑定

- [ ] 4.2 更新编辑物品表单

  - File: `admin/views/shop/pages/material/edit.vue` 或相关组件
  - Purpose: 在编辑物品表单中添加"允许负库存"开关组件，并绑定现有值
  - Requirements: 5.2
  - Leverage: 现有页面: `admin/views/shop/pages/material/`
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在编辑物品表单中添加"允许负库存"开关组件，并绑定现有值 | Context: 使用 Element Plus 的 el-switch 组件，从 API 响应中获取 allow_negative_stock 值（int 转 bool），绑定到表单字段 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 表单更新成功，开关组件正确绑定现有值

- [ ] 4.3 更新 API 封装（如需要）

  - File: `admin/views/shop/api/material.ts`
  - Purpose: 确保 API 封装正确传递 allow_negative_stock 字段
  - Requirements: 5.4
  - Leverage: 现有 API: `admin/views/shop/api/material.ts`
  - Success: API 封装正确传递字段

---

## Phase 5: 测试和验证

- [ ] 5.1 Repository 单元测试

  - File: `main/app/repository/material_repo_test.go`
  - Purpose: 测试 UpdateMaterialAllowNegativeStock 方法
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 UpdateMaterialAllowNegativeStock 方法编写单元测试 | Context: 测试 true/false 值的更新，测试 uuid 不存在的情况，测试数据库错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 5.2 Service 单元测试

  - File: `main/app/service/material_test.go`
  - Purpose: 测试 AddMaterial 和 EditMaterial 方法中 AllowNegativeStock 字段的处理逻辑
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 AddMaterial 和 EditMaterial 方法中 AllowNegativeStock 字段的处理逻辑编写单元测试 | Context: 测试添加物品时保存 AllowNegativeStock，测试编辑物品时更新 AllowNegativeStock，测试 ERP 同步逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 5.3 API 集成测试

  - File: `main/app/api/v1/shop/shop_material_test.go` 或集成测试文件
  - Purpose: 测试添加和编辑物品接口的 AllowNegativeStock 功能
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Automation Engineer | Task: 为添加和编辑物品接口编写集成测试 | Context: 测试添加物品时传递 allow_negative_stock 参数，测试编辑物品时修改 allow_negative_stock 参数，验证响应中包含该字段 | Restrictions: 测试真实 API 调用 | Success: 所有 API 测试通过

- [ ] 5.4 ERP 同步测试

  - File: 集成测试文件
  - Purpose: 测试 ERP 同步时 AllowNegativeStock 字段的正确传递
  - Requirements: 4.2
  - Leverage: 现有 ERP 同步测试
  - Success: ERP 同步测试通过，字段正确传递

- [ ] 5.5 前端功能测试

  - File: -
  - Purpose: 手动测试前端添加/编辑物品表单的 AllowNegativeStock 功能
  - Requirements: 5.3
  - Success: 前端功能测试通过，表单正确显示和提交

---

## Phase 6: 文档更新

- [ ] 6.1 更新 API 文档

  - File: `docs/shared/api/` 或相关 API 文档
  - Purpose: 更新物品管理 API 文档，添加 allow_negative_stock 字段说明
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Success: API 文档已更新

- [ ] 6.2 更新数据库文档

  - File: 数据库文档
  - Purpose: 更新数据库表结构文档，添加 allow_negative_stock 字段说明
  - Requirements: 文档要求
  - Leverage: 现有数据库文档
  - Success: 数据库文档已更新

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
- [ ] 遵循 `.cursor/rules/php.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/vue.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-material-allow-negative-stock/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-material-allow-negative-stock/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-material-allow-negative-stock/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-material-allow-negative-stock/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-material-allow-negative-stock/tasks.md)" | bc
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
**最后更新**: 2025-12-10  
**维护者**: 后端开发组

