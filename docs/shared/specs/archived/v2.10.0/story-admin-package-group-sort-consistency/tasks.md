# 确保套餐分组在各个端排序一致 任务分解

> 本文档定义确保套餐分组在各个端排序一致功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 8  
**进行中**: -  
**完成率**: 72.7%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_sort_to_product_package_group_table.php`
  - Purpose: 在 ttpos_product_package_group 表中添加 sort 字段
  - Requirements: 1.3
  - Leverage: 现有迁移文件: `admin/database/migrations/20251131980000_add_group_type_and_optional_count_to_product_package_group_table.php`
  - Prompt: Role: Database Engineer | Task: 创建添加 sort 字段到 ttpos_product_package_group 表的迁移文件 | Context: sort 字段类型为 int，NOT NULL DEFAULT 0，添加在 optional_count 字段之后，添加索引 idx_product_package_uuid_sort | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加 sort 字段
  - Requirements: 1.3
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/product_package_group.go`
  - Purpose: 在 ProductPackageGroup 结构体中添加 Sort 字段
  - Requirements: 1.3
  - Leverage: 现有 Model: `main/app/model/product_package_group.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 ProductPackageGroup 结构体中添加 Sort 字段 | Context: Sort 字段类型为 int，gorm 标签为 `type:int;not null;default:0;comment:排序字段，数值越小越靠前`，添加在 OptionalCount 字段之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: 核心实现（Go Main）

### Repository 层

- [x] 2.1 更新 Repository 查询方法（添加排序）

  - File: `main/app/repository/product_package_group.go`
  - Purpose: 在查询套餐分组时按 sort 字段排序
  - Requirements: 2.1, 2.2
  - Leverage: 现有 Repository: `main/app/repository/product_package_group.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 更新 ProductPackageGroupRepo 的查询方法，添加按 sort 字段升序排序 | Context: GetProductPackageGroup 方法添加 `Order("sort ASC, id ASC")`，WithProductPackageGroup 预加载方法在 Preload 中添加排序 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 查询方法更新成功，排序正确

- [x] 2.2 更新预加载方法（添加排序）

  - File: `main/app/repository/product_package_group.go`
  - Purpose: 在预加载套餐分组时按 sort 字段排序
  - Requirements: 2.1, 2.2
  - Leverage: 现有预加载方法: `main/app/repository/product_package_group.go`
  - Prompt: Role: Go Developer with GORM expertise | Task: 更新 WithProductPackageGroup 预加载方法，在 Preload 回调中添加排序 | Context: 在 Preload 回调中添加 `Order("sort ASC, id ASC")`，相同 sort 值时按 id 升序排序 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 预加载方法更新成功，排序正确

### Service 层

- [x] 2.3 更新 Service 保存方法（设置排序值）

  - File: `main/app/service/product.go`
  - Purpose: 在保存套餐分组时根据数组索引设置 sort 值
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Service: `main/app/service/product.go`，SaveProductPackageGroup 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 更新 SaveProductPackageGroup 方法，根据数组索引设置 sort 值 | Context: 遍历 groupList 时，使用 index + 1 作为 sort 值（或直接使用 index），在创建和更新分组时设置 sort 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 更新成功，排序值设置正确

---

## Phase 3: PHP Admin 模块实现

- [x] 3.1 更新 ProductPackageGroup Model（addPackageGroup 方法设置排序）

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在添加套餐分组时根据数组索引设置 sort 值
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroup.php`，addPackageGroup 方法
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 更新 addPackageGroup 方法，根据数组索引设置 sort 值 | Context: 遍历 $packageGroup 时，使用 $index + 1 作为 sort 值，在 $insertGroups 数组中添加 'sort' => $sortValue | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功，排序值设置正确

- [x] 3.2 更新 ProductPackageGroup Model（updatePackageGroup 方法设置排序）

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在更新套餐分组时根据数组索引设置 sort 值
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroup.php`，updatePackageGroup 方法
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 更新 updatePackageGroup 方法，根据数组索引设置 sort 值 | Context: 遍历 $groupList 时，使用 $index + 1 作为 sort 值，在 $groupData 数组中添加 'sort' => $sortValue，在创建和更新分组时都设置 sort 字段 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功，排序值设置正确

- [x] 3.3 更新 Product Model（关联查询添加排序）

  - File: `admin/app/common/model/product/Product.php`
  - Purpose: 在查询套餐分组时按 sort 字段排序
  - Requirements: 2.1, 2.2
  - Leverage: 现有 Model: `admin/app/common/model/product/Product.php`，productPackageGroup 关联方法
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 更新 productPackageGroup 关联方法，添加按 sort 字段排序 | Context: 在 hasMany 关联中添加 ->order('sort', 'asc')->order('id', 'asc')，或者在查询时使用 with(['productPackageGroup' => function($q) { $q->order('sort', 'asc')->order('id', 'asc'); }])，相同 sort 值时按 id 升序排序 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 关联查询更新成功，排序正确

---

## Phase 4: 测试和优化

- [ ] 4.1 编写 Repository 单元测试

  - File: `main/app/repository/product_package_group_test.go`
  - Purpose: 确保 Repository 查询时排序正确
  - Requirements: 2.1, 2.2
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 ProductPackageGroupRepo 编写单元测试，测试排序功能 | Context: 测试查询时按 sort 字段排序，测试预加载时排序正确 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.2 编写 Service 单元测试

  - File: `main/app/service/product_test.go`
  - Purpose: 确保 Service 保存时排序值设置正确
  - Requirements: 1.1, 1.2
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SaveProductPackageGroup 方法编写单元测试，测试排序值设置 | Context: 测试保存时根据数组索引设置 sort 值，测试多个分组排序正确 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.3 API 集成测试

  - File: `test/integration/package_group_sort_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试保存套餐分组时排序正确，测试查询时排序正确，测试多终端排序一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.4 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms，数据库查询 < 50ms

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
- [ ] 数据库文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-package-group-sort-consistency/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-package-group-sort-consistency/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-package-group-sort-consistency/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-package-group-sort-consistency/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-package-group-sort-consistency/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: Go 代码 `go fmt`, `go vet`, `go test`；PHP 代码 PSR-2 格式检查
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
**最后更新**: 2025-12-02  
**维护者**: 后端开发组

