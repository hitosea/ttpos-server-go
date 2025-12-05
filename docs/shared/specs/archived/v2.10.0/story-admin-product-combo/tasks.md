# 旧管理端-商品管理-套餐组可选数量校验 任务分解

> 本文档定义套餐组可选数量校验功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 3  
**已完成**: 2  
**进行中**: -  
**完成率**: 67%

---

## Phase 1: 代码实现

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 在 addPackageGroup 方法中添加可选数量校验

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在新增套餐分组时校验可选数量 >= 1
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有方法: `admin/app/common/model/product/ProductPackageGroup.php::addPackageGroup()`，已有部分校验逻辑（检查必选+默认是否大于可选数量）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 ProductPackageGroup::addPackageGroup() 方法中添加可选数量校验，确保 optional_count >= 1 | Context: 在保存套餐组数据前，遍历 package_group 数组，对每个套餐组检查 optional_count 字段，如果小于 1 则抛出异常"套餐组可选数量不能小于 1" | Restrictions: 遵循 .cursor/rules/php.mdc，校验应在保存数据库前执行，使用异常处理 | Success: 校验逻辑添加成功，当 optional_count < 1 时抛出异常

- [x] 1.2 在 updatePackageGroup 方法中添加可选数量校验

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在更新套餐分组时校验可选数量 >= 1
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有方法: `admin/app/common/model/product/ProductPackageGroup.php::updatePackageGroup()`，已有部分校验逻辑（检查必选+默认是否大于可选数量）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 ProductPackageGroup::updatePackageGroup() 方法中添加可选数量校验，确保 optional_count >= 1 | Context: 在保存套餐组数据前，遍历 package_group 数组，对每个套餐组检查 optional_count 字段，如果小于 1 则抛出异常"套餐组可选数量不能小于 1" | Restrictions: 遵循 .cursor/rules/php.mdc，校验应在保存数据库前执行，使用异常处理 | Success: 校验逻辑添加成功，当 optional_count < 1 时抛出异常

- [ ] 1.3 编写单元测试

  - File: `admin/app/common/model/product/ProductPackageGroupTest.php` (如需要)
  - Purpose: 确保校验逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试文件（如有）
  - Prompt: Role: QA Engineer with PHP testing expertise | Task: 为 ProductPackageGroup 的校验逻辑编写单元测试 | Context: 测试 optional_count = 0、负数、null、正常值等场景，测试多套餐组场景 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 测试覆盖率 100%，所有测试通过

---

## Phase 2: 测试验证

- [ ] 2.1 API 测试 - 创建套餐场景

  - File: -
  - Purpose: 测试创建套餐时可选数量校验是否生效
  - Requirements: 所有功能需求
  - Leverage: API 测试工具（Postman/curl）
  - Test Cases:
    - 创建套餐，optional_count = 0 → 应返回错误响应
    - 创建套餐，optional_count = 1 → 应返回成功响应
    - 创建套餐，optional_count = -1 → 应返回错误响应
    - 创建套餐，optional_count = null → 应返回错误响应
    - 创建套餐，多个套餐组，其中一个 optional_count = 0 → 应返回错误响应
  - Success: 所有测试用例通过

- [ ] 2.2 API 测试 - 更新套餐场景

  - File: -
  - Purpose: 测试更新套餐时可选数量校验是否生效
  - Requirements: 所有功能需求
  - Leverage: API 测试工具（Postman/curl）
  - Test Cases:
    - 更新套餐，optional_count = 0 → 应返回错误响应
    - 更新套餐，optional_count = 1 → 应返回成功响应
    - 更新套餐，optional_count = -1 → 应返回错误响应
    - 更新套餐，多个套餐组，其中一个 optional_count = 0 → 应返回错误响应
  - Success: 所有测试用例通过

- [ ] 2.3 集成测试

  - File: -
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Success: 集成测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 测试覆盖率达标（校验逻辑 100%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有需要）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-product-combo/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-product-combo/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-product-combo/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-product-combo/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-product-combo/tasks.md)" | bc
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

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-28  
**维护者**: 后端开发组

