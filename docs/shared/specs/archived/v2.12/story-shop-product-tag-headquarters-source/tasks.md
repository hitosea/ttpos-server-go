# 新管理端-菜品标签-来源总部数据的商品 任务分解

> 本文档定义 新管理端-菜品标签-来源总部数据的商品 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 5  
**进行中**: -  
**完成率**: 45%

---

## Phase 1: Repository 层扩展

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 扩展 Repository 接口，添加冲突检测方法

  - File: `main/app/repository/product_label.go`
  - Purpose: 在 `IProductLabelRepo` 接口中添加 `CheckHeadquarterLabelConflict` 方法，用于检查商品是否已被总部标签关联
  - Requirements: 1.2, 2.2
  - Leverage: 现有 Repository 接口: `main/app/repository/product_label.go`，参考其他查询方法
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 IProductLabelRepo 接口中添加 CheckHeadquarterLabelConflict 方法签名 | Context: 方法接收商品包UUID列表，返回冲突的商品包列表和标签列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口方法签名定义正确

- [x] 1.2 实现 Repository 冲突检测方法

  - File: `main/app/repository/product_label.go`
  - Purpose: 实现 `CheckHeadquarterLabelConflict` 方法，查询商品包及其关联的标签，过滤出 headquarter_uuid > 0 的标签
  - Requirements: 1.2, 2.2
  - Leverage: 现有 Repository 实现: `main/app/repository/product_label.go`，参考 `WithProductPackages` 方法的 Preload 用法
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 CheckHeadquarterLabelConflict 方法，查询商品包及其关联的标签（headquarter_uuid > 0） | Context: 使用 Preload 预加载 MultiLanguageName 和 ProductLabel，过滤条件：product_label_uuid > 0 且标签的 headquarter_uuid > 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用参数化查询，软删除检查 | Success: 方法实现完整，查询逻辑正确，返回冲突的商品包和标签列表

- [ ] 1.3 编写 Repository 单元测试

  - File: `main/app/repository/product_label_test.go`
  - Purpose: 为 `CheckHeadquarterLabelConflict` 方法编写单元测试，确保查询逻辑正确
  - Requirements: 1.2, 2.2
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CheckHeadquarterLabelConflict 方法编写单元测试，覆盖率 ≥ 80% | Context: 测试场景：1) 商品已被总部标签关联 2) 商品未被总部标签关联 3) 商品无标签关联 4) 空列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 2: Service 层扩展

- [x] 2.1 添加 Service 冲突检测私有方法

  - File: `main/app/service/product_label.go`
  - Purpose: 在 `ProductLabelSrvImpl` 中添加 `checkHeadquarterLabelConflict` 私有方法，调用 Repository 方法并组装错误信息
  - Requirements: 1.3, 2.3
  - Leverage: 现有 Service 实现: `main/app/service/product_label.go`，参考错误处理方式
  - Prompt: Role: Go Developer with business logic expertise | Task: 添加 checkHeadquarterLabelConflict 私有方法，调用 Repository 的 CheckHeadquarterLabelConflict，组装商品名称列表和标签名称 | Context: 从冲突的商品包中提取 MultiLanguageName，组装错误信息格式：`商品[商品A、商品B]已经被来源总部的标签[标签名称1]关联，无法被当前标签关联` | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 errors.WithMessage 包装错误 | Success: 方法实现完整，错误信息格式正确

- [x] 2.2 在 AddProductLabel 方法中添加冲突检测

  - File: `main/app/service/product_label.go`
  - Purpose: 在 `AddProductLabel` 方法中，保存标签前调用冲突检测方法，如果存在冲突则返回错误
  - Requirements: 1.1, 1.3, 1.4
  - Leverage: 现有 Service 方法: `main/app/service/product_label.go` 的 `AddProductLabel` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 AddProductLabel 方法中，保存标签前添加冲突检测逻辑 | Context: 如果 req.ProductPackageUuids 不为空，调用 checkHeadquarterLabelConflict 检查冲突，如果存在冲突则返回错误，阻止保存 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 冲突检测逻辑正确，错误信息正确，不影响正常流程

- [x] 2.3 在 EditProductLabel 方法中添加冲突检测

  - File: `main/app/service/product_label.go`
  - Purpose: 在 `EditProductLabel` 方法中，更新关联商品前调用冲突检测方法，如果存在冲突则返回错误
  - Requirements: 2.1, 2.3, 2.4
  - Leverage: 现有 Service 方法: `main/app/service/product_label.go` 的 `EditProductLabel` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 EditProductLabel 方法中，更新关联商品前添加冲突检测逻辑 | Context: 如果 req.ProductPackageUuids 不为空，调用 checkHeadquarterLabelConflict 检查冲突，如果存在冲突则返回错误，阻止保存 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 冲突检测逻辑正确，错误信息正确，不影响正常流程

- [ ] 2.4 编写 Service 单元测试

  - File: `main/app/service/product_label_test.go`
  - Purpose: 为 `AddProductLabel` 和 `EditProductLabel` 方法编写单元测试，覆盖冲突检测场景
  - Requirements: 1.1, 1.3, 2.1, 2.3
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 AddProductLabel 和 EditProductLabel 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试场景：1) 创建标签时，关联商品已被总部标签关联 - 应返回错误 2) 创建标签时，关联商品未被总部标签关联 - 应成功创建 3) 编辑标签时，新增商品已被总部标签关联 - 应返回错误 4) 编辑标签时，新增商品未被总部标签关联 - 应成功更新 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 3: 测试和优化

- [ ] 3.1 集成测试

  - File: `test/integration/product_label_conflict_test.go`
  - Purpose: 测试端到端功能，验证 API 接口的冲突检测逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，测试商品标签冲突检测功能 | Context: 测试用户完整流程：1) 创建总部标签并关联商品 2) 分店尝试创建标签关联相同商品 - 应返回错误 3) 分店创建标签关联其他商品 - 应成功 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 3.2 性能测试

  - File: -
  - Purpose: 确保冲突检测查询性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 本地响应时间 < 200ms，数据库查询 < 50ms

- [ ] 3.3 错误信息国际化

  - File: `main/app/service/product_label.go`, `main/i18n/`
  - Purpose: 将错误提示信息支持多语言
  - Requirements: 国际化要求
  - Leverage: 现有国际化实现: `main/i18n/`
  - Prompt: Role: Go Developer | Task: 将冲突检测错误信息支持多语言 | Context: 使用 i18n 包，错误信息需要支持 10 种语言 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 错误信息支持多语言，所有语言测试通过

- [ ] 3.4 文档更新

  - File: `docs/shared/api/shop_product_label_api.md`（如有）
  - Purpose: 更新 API 文档，说明冲突检测错误响应
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Success: API 文档已更新，错误响应说明完整

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
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
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-product-tag-headquarters-source/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-product-tag-headquarters-source/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-product-tag-headquarters-source/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-product-tag-headquarters-source/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-product-tag-headquarters-source/tasks.md)" | bc
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
**最后更新**: 2025-12-08  
**维护者**: TTPOS Team

