# 旧后台商品管理-成品/套餐可选范围调整 任务分解

> 本文档定义商品管理中属性、加料、套餐分组可选范围调整的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 21（已排除数据库迁移任务）  
**已完成**: 8  
**进行中**: 2.9  
**完成率**: 38%

---

## ⚠️ 前置说明

**数据库字段已在任务 #37946 中添加**（迁移文件：`20251222145027_add_selection_range_fields.php`），因此本任务**无需执行数据库迁移**，直接从后端逻辑调整开始。

**已完成的数据库变更**：
- ✅ `ttpos_product_package` 表已添加 `sauce_min_selection` 字段
- ✅ `ttpos_product_package_attribute_group` 表已添加 `min_selection` 字段
- ✅ `ttpos_product_package_group` 表已添加 `optional_min_count` 字段
- ✅ 旧数据已自动迁移

---

## Phase 1: 数据验证（可选）

### 验证数据库字段

- [ ] 1.1 验证数据库字段是否已添加

  - File: -
  - Purpose: 确认任务 #37946 的数据库迁移已执行
  - Requirements: 1.1, 2.1, 3.1
  - Command: `SHOW COLUMNS FROM ttpos_product_package LIKE 'sauce_min_selection';`
  - Success: 字段存在且旧数据已迁移（sauce_required=1 的记录 sauce_min_selection=1）

---

## Phase 2: 后端逻辑调整

### 商品模型调整

- [x] 2.1 调整商品模型 - 属性验证逻辑（add 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 将"必选+最大可选"验证逻辑调整为"最小-最大可选范围"验证
  - Requirements: 1.2, 1.3, 1.4, 1.5, 1.6, 1.7
  - Leverage: 现有 add() 方法第113-128行，参考 design.md 中的调整逻辑
  - Success: 验证逻辑正确，支持最小-最大可选范围，错误提示准确

- [x] 2.2 调整商品模型 - 加料验证逻辑（add 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 将加料"必选+最大可选"验证逻辑调整为"最小-最大可选范围"验证
  - Requirements: 2.2, 2.3, 2.4, 2.5, 2.6, 2.7
  - Leverage: 现有 add() 方法第189-214行，参考 design.md 中的调整逻辑
  - Success: 验证逻辑正确，加料数量限制从10调整为100，支持最小-最大可选范围

- [x] 2.3 调整商品模型 - 套餐分组验证逻辑（add 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 添加套餐分组可选范围验证逻辑
  - Requirements: 3.2, 3.3, 3.4, 3.5, 3.6
  - Leverage: 现有 add() 方法第129-187行，参考 design.md 中的调整逻辑
  - Success: 分组数量限制从5调整为100，支持可选范围验证，允许为0

- [x] 2.4 调整商品模型 - 数据映射（add 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: 调整字段映射，新字段映射到新数据库字段，同时兼容旧字段
  - Requirements: 1.1, 2.1
  - Leverage: 现有 add() 方法第254-257行，参考 design.md 中的调整逻辑
  - Success: 字段映射正确，sauce_min_selection 和 sauce_max_selection 保存正确

- [x] 2.5 调整商品模型 - 属性验证逻辑（edit 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: edit 方法的属性验证逻辑调整（同 add 方法）
  - Requirements: 1.2, 1.3, 1.4, 1.5, 1.6, 1.7
  - Leverage: Task 2.1 的调整逻辑，应用到 edit() 方法第565-581行
  - Success: 验证逻辑正确，与 add 方法一致

- [x] 2.6 调整商品模型 - 加料验证逻辑（edit 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: edit 方法的加料验证逻辑调整（同 add 方法）
  - Requirements: 2.2, 2.3, 2.4, 2.5, 2.6, 2.7
  - Leverage: Task 2.2 的调整逻辑，应用到 edit() 方法第647-676行
  - Success: 验证逻辑正确，与 add 方法一致

- [x] 2.7 调整商品模型 - 套餐分组验证逻辑（edit 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: edit 方法的套餐分组验证逻辑调整（同 add 方法）
  - Requirements: 3.2, 3.3, 3.4, 3.5, 3.6
  - Leverage: Task 2.3 的调整逻辑，应用到 edit() 方法第582-645行
  - Success: 验证逻辑正确，与 add 方法一致

- [x] 2.8 调整商品模型 - 数据映射（updateProductPackage 方法）

  - File: `admin/app/shop/model/product/Product.php`
  - Purpose: updateProductPackage 方法的字段映射调整
  - Requirements: 1.1, 2.1
  - Leverage: 现有 updateProductPackage() 方法第797-799行
  - Success: 字段映射正确，sauce_min_selection 更新正确

### 属性模型调整

- [x] 2.9 调整属性模型 - addAttribute 方法

  - File: `admin/app/shop/model/product/ProductAttribute.php`
  - Purpose: 将属性组字段从 is_must + max_selection 调整为 min_selection + max_selection
  - Requirements: 1.1
  - Leverage: 现有 addAttribute() 方法第15-31行，参考 design.md 中的调整逻辑
  - Success: 属性组保存时 min_selection 和 max_selection 字段正确，兼容 is_must 字段
  - **实施记录**: 已完成，添加对 attribute_min_select 和 attribute_max_select 的支持

- [x] 2.10 调整属性模型 - updateAttribute 方法

  - File: `admin/app/shop/model/product/ProductAttribute.php`
  - Purpose: updateAttribute 方法的字段调整（同 addAttribute）
  - Requirements: 1.1
  - Leverage: Task 2.9 的调整逻辑，应用到 updateAttribute() 方法第52-77行
  - Success: 属性组更新时 min_selection 和 max_selection 字段正确
  - **实施记录**: 已完成，移除旧的 attribute_open_max_select 逻辑

### 套餐分组模型调整

- [x] 2.11 调整套餐分组模型 - addPackageGroup 方法

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 添加 min_selection 和 max_selection 字段保存
  - Requirements: 3.1
  - Leverage: 现有 addPackageGroup() 方法，参考 design.md 中的调整逻辑
  - Success: 套餐分组保存时 min_selection 和 max_selection 字段正确
  - **实施记录**: 已完成，添加 optional_min_count 字段支持和验证逻辑

- [x] 2.12 调整套餐分组模型 - updatePackageGroup 方法

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: updatePackageGroup 方法添加字段更新
  - Requirements: 3.1
  - Leverage: Task 2.11 的调整逻辑，应用到 updatePackageGroup() 方法
  - Success: 套餐分组更新时 min_selection 和 max_selection 字段正确
  - **实施记录**: 已完成，添加 optional_min_count 字段支持和验证逻辑

### 通用模型调整

- [x] 2.13 调整通用商品模型字段定义

  - File: `admin/app/common/model/product/Product.php`
  - Purpose: 添加新字段到 Model 字段列表
  - Requirements: 1.1, 2.1, 3.1
  - Leverage: 现有字段定义第54-56行
  - Success: 字段列表包含 sauce_min_selection
  - **实施记录**: 已完成，添加 feed_min_select 兼容字段和访问器

### 订单验证逻辑调整

- [x] 2.14 调整订单验证逻辑 - 属性验证

  - File: `admin/app/common/model/BaseModelOrder.php`
  - Purpose: 订单提交时的属性可选范围验证调整
  - Requirements: 1.1
  - Leverage: 现有验证逻辑第1229-1248行
  - Success: 验证逻辑使用 min_selection 和 max_selection，兼容旧字段
  - **实施记录**: 已完成，使用 attribute_min_select 和 attribute_max_select 进行范围验证

- [x] 2.15 调整订单验证逻辑 - 加料验证

  - File: `admin/app/common/model/BaseModelOrder.php`
  - Purpose: 订单提交时的加料可选范围验证调整
  - Requirements: 2.1
  - Leverage: 现有验证逻辑第1268-1286行
  - Success: 验证逻辑使用 sauce_min_selection 和 sauce_max_selection，兼容旧字段
  - **实施记录**: 已完成，使用 feed_min_select 和 feed_max_select 进行范围验证

---

## Phase 3: 前端界面调整

### 属性表单调整

- [ ] 3.1 调整属性表单字段

  - File: `admin/views/shop/src/views/product/store/product/part/Attr.vue`
  - Purpose: 将"必选+最大可选"字段调整为"最小可选-最大可选"范围输入
  - Requirements: 1.2, 1.5
  - Leverage: 现有属性表单组件
  - Success: 表单显示"最小可选"和"最大可选"两个输入框，验证逻辑正确

- [ ] 3.2 调整属性表单验证

  - File: 同 Task 3.1
  - Purpose: 前端验证最大可选 ≥ 最小可选
  - Requirements: 1.5
  - Leverage: 现有前端验证逻辑
  - Success: 前端验证正确，错误提示友好

- [ ] 3.3 调整属性组数量限制提示

  - File: 同 Task 3.1
  - Purpose: 将属性组数量限制提示从原有限制调整为100
  - Requirements: 1.7
  - Leverage: 现有数量限制逻辑
  - Success: 超过100个属性组时提示"属性组数量不能超过100"

### 加料表单调整

- [ ] 3.4 调整加料表单字段

  - File: `admin/views/shop/src/views/product/store/product/part/Ingredients.vue`
  - Purpose: 将"必选+最大可选"字段调整为"最小可选-最大可选"范围输入
  - Requirements: 2.3, 2.5
  - Leverage: 现有加料表单组件
  - Success: 表单显示"最小可选"和"最大可选"两个输入框，验证逻辑正确

- [ ] 3.5 调整加料表单验证

  - File: 同 Task 3.4
  - Purpose: 前端验证最大可选 ≥ 最小可选
  - Requirements: 2.5
  - Leverage: 现有前端验证逻辑
  - Success: 前端验证正确，错误提示友好

- [ ] 3.6 调整加料数量限制提示

  - File: 同 Task 3.4
  - Purpose: 将加料数量限制从10调整为100
  - Requirements: 2.7
  - Leverage: 现有数量限制逻辑
  - Success: 超过100个加料时提示"最多可添加100个加料"

### 套餐分组表单调整

- [ ] 3.7 调整套餐分组表单字段

  - File: `admin/views/shop/src/views/product/store/product/part/Package.vue`
  - Purpose: 添加"最小可选-最大可选"范围输入
  - Requirements: 3.2, 3.4
  - Leverage: 现有套餐分组表单组件
  - Success: 表单显示"最小可选"和"最大可选"两个输入框，默认值为 1-1

- [ ] 3.8 调整套餐分组表单验证

  - File: 同 Task 3.7
  - Purpose: 前端验证最大可选 ≥ 最小可选，允许为0
  - Requirements: 3.4
  - Leverage: 现有前端验证逻辑
  - Success: 前端验证正确，允许最小值和最大值为0

- [ ] 3.9 调整套餐分组数量限制提示

  - File: 同 Task 3.7
  - Purpose: 将套餐分组数量限制从5调整为100
  - Requirements: 3.6
  - Leverage: 现有数量限制逻辑
  - Success: 超过100个分组时提示"分组数量不能超过100"

---

## Phase 4: 测试和修复

### 数据迁移测试

- [ ] 4.1 测试旧数据迁移

  - File: -
  - Purpose: 确保旧数据迁移后显示和功能正常
  - Requirements: 所有需求
  - Leverage: 生产环境旧数据副本
  - Success: 旧数据迁移后，属性、加料、套餐分组可选范围显示正确

### 边界值测试

- [ ] 4.2 测试边界值场景

  - File: -
  - Purpose: 测试最小值=最大值、最小值>最大值等边界情况
  - Requirements: 1.3, 2.3, 3.2
  - Leverage: 测试环境
  - Success: 边界值场景处理正确，错误提示准确

### 兼容性测试

- [ ] 4.3 测试向后兼容性

  - File: -
  - Purpose: 测试旧版前端调用新接口是否正常
  - Requirements: 所有需求
  - Leverage: 旧版前端代码
  - Success: 旧版前端调用新接口不报错，功能正常

### 功能测试

- [ ] 4.4 测试添加商品功能

  - File: -
  - Purpose: 测试添加商品时属性、加料、套餐分组可选范围设置正确
  - Requirements: 1.2, 2.3, 3.2
  - Leverage: 测试环境
  - Success: 添加商品成功，可选范围保存正确

- [ ] 4.5 测试编辑商品功能

  - File: -
  - Purpose: 测试编辑商品时属性、加料、套餐分组可选范围显示和修改正确
  - Requirements: 1.4, 2.4, 3.3
  - Leverage: 测试环境
  - Success: 编辑商品成功，可选范围显示和保存正确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] 数据库迁移脚本测试通过
- [ ] 所有验证逻辑正确
- [ ] 向后兼容性处理完善

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 旧数据迁移正确

### 文档同步

- [ ] 数据库文档已更新
- [ ] API 文档已更新（如有变化）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-product-package-attribute-sauce-selection/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-product-package-attribute-sauce-selection/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-product-package-attribute-sauce-selection/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-product-package-attribute-sauce-selection/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-product-package-attribute-sauce-selection/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **查看设计**: 查看 design.md 中的设计方案
5. **实现代码**: 按照规范实现功能
6. **运行检查**: 测试功能，确保正确
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：关键文件速查

### 后端文件

| 文件 | 用途 | 主要修改点 |
|------|------|-----------|
| `admin/app/shop/model/product/Product.php` | 商品模型 | add/edit 方法的验证逻辑 |
| `admin/app/shop/model/product/ProductAttribute.php` | 属性模型 | addAttribute/updateAttribute 方法 |
| `admin/app/common/model/product/ProductPackageGroup.php` | 套餐分组模型 | addPackageGroup/updatePackageGroup 方法 |
| `admin/app/common/model/BaseModelOrder.php` | 订单验证 | 属性和加料验证逻辑 |
| `admin/app/common/model/product/Product.php` | 通用商品模型 | 字段定义 |

### 数据库文件

| 文件 | 用途 |
|------|------|
| `admin/database/migrations/{timestamp}_add_selection_range_to_product_tables.php` | 数据迁移脚本 |

### 前端文件（待确认）

| 文件 | 用途 | 主要修改点 |
|------|------|-----------|
| `admin/views/shop/pages/product/components/AttributeForm.vue` | 属性表单 | 字段调整 |
| `admin/views/shop/pages/product/components/SauceForm.vue` | 加料表单 | 字段调整 |
| `admin/views/shop/pages/product/components/PackageGroupForm.vue` | 套餐分组表单 | 字段调整 |

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-23  
**维护者**: 后端开发组

