# Bug-251128-001 修复任务清单

> **当前状态**: 🟡 规划中
> **开始时间**: 2025-11-28
> **预计完成**: 2025-11-28

---

## 📋 任务列表

### 1. 数据库迁移

- [x] **创建数据库迁移文件 - multi_language_name 表** `admin/database/migrations/20251128094716_alter_multi_language_name_fields_to_varchar_1000.php`
  - 需求: 将 `ttpos_multi_language_name` 表的所有语言字段从 `VARCHAR(255)` 改为 `VARCHAR(1000)`
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 
    - 字段：`en_name`, `zh_name`, `zh_tw_name`, `th_name`, `my_name`, `ja_name`, `ko_name`, `tr_name`, `sv_name`
    - 参考迁移文件格式：`20251125100000_add_describe_multi_language_uuid_to_product_package_table.php`
  - ✅ **已完成**: 2025-11-28

- [x] **创建数据库迁移文件 - product_package 表** `admin/database/migrations/20251128094717_alter_product_package_describe_to_varchar_1000.php`
  - 需求: 将 `ttpos_product_package` 表的 `describe` 字段从 `VARCHAR(255)` 改为 `VARCHAR(1000)`
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 
    - 字段：`describe`
    - 参考迁移文件格式：`20251125100000_add_describe_multi_language_uuid_to_product_package_table.php`
  - ✅ **已完成**: 2025-11-28

- [x] **更新种子文件** `admin/database/seeds/shop_01.sql`
  - 需求: 同步更新 `ttpos_multi_language_name` 表和 `ttpos_product_package` 表的字段定义
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 
    - `multi_language_name` 表字段长度从 `VARCHAR(255)` 改为 `VARCHAR(1000)`
    - `product_package` 表的 `describe` 字段长度从 `VARCHAR(255)` 改为 `VARCHAR(1000)`
  - ✅ **已完成**: 2025-11-28（已更新，包含 product_package.describe 字段）

- [ ] **执行数据库迁移** 
  - 需求: 在开发环境执行迁移脚本，验证迁移成功
  - 预计时间: 0.5小时
  - 负责人: 
  - 说明: 使用 `make migrate` 命令执行迁移

### 2. 代码修复

- [x] **验证逻辑保持不变** `main/app/service/product_check.go`
  - 需求: 保持 `CheckProductSellingPoint` 方法的验证长度为 500（不修改）
  - 预计时间: 0小时
  - 负责人: 
  - 代码位置: 第 87 行
  - 说明: 验证长度（500）小于数据库字段长度（1000），确保不会超出数据库限制
  - ✅ **已确认**: 2025-11-28

- [x] **验证逻辑保持不变** `admin/app/shop/model/product/Product.php`
  - 需求: 保持 `applySellingPointLocales` 方法的验证长度为 500（不修改）
  - 预计时间: 0小时
  - 负责人: 
  - 代码位置: 第 371 行
  - 说明: 验证长度（500）小于数据库字段长度（1000），确保不会超出数据库限制
  - ✅ **已确认**: 2025-11-28

### 3. 测试验证

- [ ] **编写单元测试** `main/test/service/product_check_test.go`
  - 需求: 测试 `CheckProductSellingPoint` 方法的长度验证逻辑
  - 预计时间: 1小时
  - 负责人: 
  - 测试用例:
    - 测试 499 字符（应该通过）
    - 测试 500 字符（应该通过）
    - 测试 501 字符（应该失败，验证逻辑限制）

- [ ] **集成测试**
  - 需求: 测试商品保存功能，验证卖点长度限制
  - 预计时间: 1小时
  - 负责人: 
  - 测试步骤:
    1. 创建商品，卖点内容 500 字符，验证保存成功（数据库字段足够大）
    2. 创建商品，卖点内容 501 字符，验证保存失败（验证逻辑限制）
    3. 编辑商品，修改卖点内容，验证长度限制（500 字符）

- [ ] **手动验证**
  - 需求: 在商品编辑页面手动测试卖点保存功能
  - 预计时间: 0.5小时
  - 负责人: 
  - 测试步骤:
    1. 进入商品编辑页面（shop 终端）
    2. 输入 500 字符的卖点内容，保存，验证成功（数据库字段足够大，不会报错）
    3. 输入 501 字符的卖点内容，保存，验证提示错误（验证逻辑限制）
    4. 测试多语言卖点内容

### 4. 文档更新

- [ ] **更新数据库设计文档**（如需要）
  - 需求: 更新 `ttpos_multi_language_name` 表的字段长度说明
  - 预计时间: 0.5小时
  - 负责人: 

- [ ] **更新 API 文档**（如需要）
  - 需求: 更新卖点字段的长度限制说明
  - 预计时间: 0.5小时
  - 负责人: 

### 5. 代码审查

- [ ] **代码审查**
  - 需求: 通过 Code Review
  - 预计时间: 0.5小时
  - 负责人: 
  - 审查重点:
    - 数据库迁移脚本正确性
    - 验证逻辑修改正确性
    - 代码风格一致性

### 6. 部署上线

- [ ] **发布到测试环境**
  - 需求: 部署并验证
  - 预计时间: 0.5小时
  - 负责人: 
  - 步骤:
    1. 执行数据库迁移
    2. 部署代码
    3. 验证功能正常

- [ ] **发布到生产环境**
  - 需求: 生产发布并监控
  - 预计时间: 1小时
  - 负责人: 
  - 步骤:
    1. 选择低峰期执行数据库迁移
    2. 部署代码
    3. 监控商品保存成功率
    4. 验证功能正常

---

## 📊 任务统计

- **总任务数**: 12
- **已完成**: 5
- **进行中**: 0
- **未开始**: 7
- **完成率**: 42%

---

## 🔗 相关链接

- Bug 报告: `bug.md`
- 修复方案: `solution.md`
- 相关代码:
  - `main/app/service/product_check.go:87` - Go 模块验证逻辑
  - `admin/app/shop/model/product/Product.php:371` - PHP 模块验证逻辑
  - `admin/database/migrations/20251128094716_alter_multi_language_name_fields_to_varchar_1000.php` - multi_language_name 表迁移文件
  - `admin/database/migrations/20251128094717_alter_product_package_describe_to_varchar_1000.php` - product_package 表迁移文件
  - `admin/database/seeds/shop_01.sql` - 种子文件

---

**创建时间**: 2025-11-28 09:44  
**最后更新**: 2025-11-28 09:44

