# 旧管理端-商品管理-套餐 任务分解

> 本文档定义旧管理端商品管理中套餐功能的 PHP 后端 API 详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 8  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: Model 更新

- [x] 1.1 更新 ProductPackageGroup Model - updatePackageGroup 方法

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在 updatePackageGroup 方法中支持 group_type 和 optional_count 字段
  - Requirements: 1.3, 1.4
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroup.php`，updatePackageGroup 方法
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 更新 ProductPackageGroup::updatePackageGroup 方法，支持接收和处理 group_type 和 optional_count 字段 | Context: group_type 默认值为 0（固定），optional_count 默认值为 0，当 group_type 为 1（可选）时使用 optional_count | Restrictions: 遵循 .cursor/rules/php.mdc，使用 ThinkPHP ORM | Success: Model 更新成功，字段处理正确

- [x] 1.2 更新 ProductPackageGroupItem Model

- [x] 1.3 更新 ProductPackageGroup Model - addPackageGroup 方法

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在 addPackageGroup 方法中支持所有新字段和数据校验
  - Requirements: 1.3, 1.4, 2.3, 2.4, 2.5, 3.2, 4.2, 5.1
  - Leverage: 现有方法: `admin/app/common/model/product/ProductPackageGroup.php`，addPackageGroup 方法
  - Success: addPackageGroup 方法支持所有新字段，包含数据校验逻辑

  - File: `admin/app/common/model/product/ProductPackageGroupItem.php`
  - Purpose: 在 updatePackageGroup 方法中支持 is_required、is_default、add_price 字段，num 字段默认值为 1
  - Requirements: 2.4, 2.5, 3.2, 4.2
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroupItem.php`，ProductPackageGroup::updatePackageGroup 方法
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 更新 ProductPackageGroup::updatePackageGroup 方法中的商品保存逻辑，支持 is_required、is_default、add_price 字段，num 字段默认值为 1 | Context: is_required 和 is_default 为 int 类型（0-否 1-是），默认值为 0；add_price 默认值为 0；num 默认值为 1 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功，字段处理正确

---

## Phase 2: 核心实现（PHP）

### 数据校验

- [x] 2.1 实现数据校验逻辑

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在 updatePackageGroup 方法中添加数据校验：必选数量不可大于可选数量
  - Requirements: 2.3, 5.1
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroup.php`，updatePackageGroup 方法
  - Prompt: Role: PHP Developer with business logic expertise | Task: 在 updatePackageGroup 方法中添加数据校验逻辑：当 group_type 为可选时，检查必选数量是否大于可选数量 | Context: 遍历 groupItemList，统计每个可选组的必选商品数量（is_required=1），与 optional_count 比较 | Restrictions: 遵循 .cursor/rules/php.mdc，抛出异常处理错误 | Success: 数据校验逻辑正确，错误提示清晰

- [x] 2.2 实现参数验证

  - File: `admin/app/shop/model/product/Product.php` (验证逻辑在 Model 层)
  - Purpose: 参数验证已在 Model 层的 add 和 edit 方法中实现
  - Requirements: 5.2
  - Leverage: 现有验证逻辑: `admin/app/shop/model/product/Product.php`，使用 ValidateHelp 进行验证
  - Note: 验证逻辑已在 Model 层实现，无需创建独立验证器

---

## Phase 3: 商品详情回显

- [x] 3.1 更新商品详情回显

  - File: `admin/app/common/model/product/Product.php`
  - Purpose: 在 getPackageGroup 方法中添加新字段的回显
  - Requirements: 所有功能需求
  - Leverage: 现有方法: `admin/app/common/model/product/Product.php`，getPackageGroup 方法
  - Success: 商品详情接口返回新字段：group_type、optional_count、is_required、is_default、add_price

---

## Phase 4: 测试和优化

- [x] 4.1 测试指南文档

  - File: `docs/shared/specs/active/story-shop-package-management/test-guide.md`
  - Purpose: 创建手动测试指南和测试用例
  - Requirements: 所有功能需求
  - Leverage: 测试最佳实践
  - Success: 测试指南文档已创建，包含完整的测试用例

- [x] 4.2 API 文档更新

  - File: `docs/shared/api/shop-package-management-enhancement.md`
  - Purpose: 创建 API 文档，说明接口变更
  - Requirements: 所有功能需求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档已创建，包含接口变更说明和示例

- [ ] 4.3 单元测试（可选）

  - File: `admin/app/common/model/product/ProductPackageGroupTest.php` (或相关测试文件)
  - Purpose: 为 updatePackageGroup 方法编写单元测试
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `admin/tests/`（如存在）
  - Note: 项目暂无标准测试框架，已创建测试指南文档
  - Success: 测试指南文档已创建

- [ ] 4.4 集成测试（可选）

  - File: `admin/tests/integration/PackageGroupTest.php` (或相关测试文件)
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Note: 项目暂无标准测试框架，已创建测试指南文档
  - Success: 测试指南文档已创建

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-package-management/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-package-management/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-package-management/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-package-management/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-package-management/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: PHP 代码格式化检查
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### PHP 后端开发

```
Role: PHP Developer with ThinkPHP expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/php.mdc, .cursor/rules/api.mdc, .cursor/rules/database.mdc

Restrictions:
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除
- 遵循 PSR-2 代码风格
- 使用 ThinkPHP ORM

Success Criteria:
- {成功标准1}
- 代码通过 PSR-2 格式检查
- 测试通过
```

### 测试工程师

```
Role: QA Engineer with PHP testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

Restrictions:
- 遵循 .cursor/rules/php.mdc
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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-25  
**维护者**: 开发组
