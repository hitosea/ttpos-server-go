# 自助点餐机管理 任务分解

> 本文档定义自助点餐机管理功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 12  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_enable_kiosk_to_company_setting.php`
  - Purpose: 在 company_setting 表中添加 enable_kiosk 字段
  - Requirements: 1.1, 1.7
  - Leverage: 现有迁移文件: `admin/database/migrations/20251120013811_add_table_map_fields_to_company_setting.php`
  - Prompt: Role: Database Engineer | Task: 创建数据库迁移文件，在 company_setting 表中添加 enable_kiosk 字段（INT(3)，默认值 0，注释：是否启用自助点餐机：0-否；1-是）| Context: 参考 enable_data_management 字段的实现方式，字段位置在 enable_data_management 之后 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [x] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加 enable_kiosk 字段
  - Requirements: 1.1, 1.7
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 App Model 查询字段

  - File: `admin/app/admin/model/app/App.php`
  - Purpose: 在商家列表查询中添加 enable_kiosk 字段
  - Requirements: 1.4, 1.6
  - Leverage: 现有 Model: `admin/app/admin/model/app/App.php` (第 96 行)
  - Prompt: Role: PHP Developer | Task: 在 App Model 的 getList() 方法的 $field 数组中添加 "su.enable_kiosk" | Context: 参考 enable_data_management 字段的添加方式（第 96 行）| Restrictions: 遵循 .cursor/rules/php.mdc | Success: 字段添加成功，列表查询返回 enable_kiosk 字段

---

## Phase 2: 验证器和 Controller（商家管理）

- [x] 2.1 在验证器中添加 enable_kiosk 验证规则

  - File: `admin/app/admin/validate/AppValidate.php`
  - Purpose: 添加 enable_kiosk 参数的验证规则
  - Requirements: 1.5
  - Leverage: 现有验证器: `admin/app/admin/validate/AppValidate.php` (第 54-55 行，第 110-111 行，第 151-152 行)
  - Prompt: Role: PHP Developer | Task: 在 AppValidate 中添加 enable_kiosk 验证规则 | Context: 在 $rule 数组中添加 'enable_kiosk|是否启用自助点餐机' => 'in:0,1'，在 $scene['add'] 和 $scene['edit'] 数组中添加 'enable_kiosk' | Restrictions: 遵循 .cursor/rules/php.mdc，参考 enable_data_management 的实现方式 | Success: 验证规则添加成功，参数验证正确

- [x] 2.2 更新商家新建接口文档注释

  - File: `admin/app/admin/controller/Shop.php`
  - Purpose: 在新建商家接口的 API 文档注释中添加 enable_kiosk 参数说明
  - Requirements: 1.2
  - Leverage: 现有 Controller: `admin/app/admin/controller/Shop.php` (add 方法，第 101 行)
  - Prompt: Role: PHP Developer | Task: 在 Shop Controller 的 add() 方法的 API 文档注释中添加 enable_kiosk 参数说明 | Context: 参考 enable_data_management 参数的文档注释格式（第 180 行）| Restrictions: 遵循 API 文档规范 | Success: 文档注释添加成功，参数说明正确

- [x] 2.3 更新商家编辑接口文档注释

  - File: `admin/app/admin/controller/Shop.php`
  - Purpose: 在编辑商家接口的 API 文档注释中添加 enable_kiosk 参数说明
  - Requirements: 1.3
  - Leverage: 现有 Controller: `admin/app/admin/controller/Shop.php` (edit 方法，第 180 行)
  - Prompt: Role: PHP Developer | Task: 在 Shop Controller 的 edit() 方法的 API 文档注释中添加 enable_kiosk 参数说明 | Context: 参考 enable_data_management 参数的文档注释格式（第 180 行）| Restrictions: 遵循 API 文档规范 | Success: 文档注释添加成功，参数说明正确

---

## Phase 3: 客户端版本管理

- [x] 3.1 更新客户端版本列表接口文档注释

  - File: `admin/app/admin/controller/client/Client.php`
  - Purpose: 在客户端版本列表接口的 API 文档注释中添加自助点餐机类型说明（type=6）
  - Requirements: 2.1, 2.5
  - Leverage: 现有 Controller: `admin/app/admin/controller/client/Client.php` (index 方法，第 47 行)
  - Prompt: Role: PHP Developer | Task: 在 Client Controller 的 index() 方法的 API 文档注释中更新 type 参数说明，添加"6自助点餐机" | Context: 当前说明为"类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端"，需要添加"6自助点餐机" | Restrictions: 遵循 API 文档规范 | Success: 文档注释更新成功，类型说明完整

- [x] 3.2 更新客户端版本添加接口文档注释

  - File: `admin/app/admin/controller/client/Client.php`
  - Purpose: 在客户端版本添加接口的 API 文档注释中添加自助点餐机类型说明（type=6）
  - Requirements: 2.2, 2.5
  - Leverage: 现有 Controller: `admin/app/admin/controller/client/Client.php` (add 方法)
  - Prompt: Role: PHP Developer | Task: 在 Client Controller 的 add() 方法的 API 文档注释中更新 type 参数说明，添加"6自助点餐机" | Context: 参考 index() 方法的类型说明更新方式 | Restrictions: 遵循 API 文档规范 | Success: 文档注释更新成功，类型说明完整

- [x] 3.3 更新客户端版本查询接口文档注释

  - File: `admin/app/admin/controller/client/Client.php`
  - Purpose: 在客户端版本查询接口的 API 文档注释中添加自助点餐机类型说明（type=6）
  - Requirements: 2.4, 2.5
  - Leverage: 现有 Controller: `admin/app/admin/controller/client/Client.php` (getNewVersion 方法)
  - Prompt: Role: PHP Developer | Task: 在 Client Controller 的 getNewVersion() 方法的 API 文档注释中更新 type 参数说明，添加"6自助点餐机" | Context: 参考 index() 方法的类型说明更新方式 | Restrictions: 遵循 API 文档规范 | Success: 文档注释更新成功，类型说明完整

- [x] 3.4 验证客户端版本管理功能支持 type=6

  - File: `admin/app/admin/controller/client/Client.php`
  - Purpose: 验证现有客户端版本管理功能已支持 type=6（自助点餐机）
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有 Controller: `admin/app/admin/controller/client/Client.php`
  - Prompt: Role: QA Engineer | Task: 验证客户端版本管理功能（列表、添加、发布、查询）已支持 type=6 | Context: 检查现有代码逻辑，确认 type 参数为动态值，无需修改代码逻辑 | Restrictions: 确保所有版本管理功能正常工作 | Success: 功能验证通过，type=6 正常工作

---

## Phase 4: 测试和验证

- [x] 4.1 API 测试 - 商家管理接口

  - File: `test/api/shop_test.php` (如存在)
  - Purpose: 测试商家新建/编辑接口的 enable_kiosk 参数
  - Requirements: 1.2, 1.3, 1.4
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer | Task: 编写商家管理接口的 API 测试 | Context: 测试新建商家时 enable_kiosk 参数（默认值、0、1），测试编辑商家时 enable_kiosk 参数，测试商家列表返回 enable_kiosk 字段 | Restrictions: 遵循测试规范 | Success: 所有测试通过

- [x] 4.2 API 测试 - 客户端版本管理接口

  - File: `test/api/client_test.php` (如存在)
  - Purpose: 测试客户端版本管理接口支持 type=6
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer | Task: 编写客户端版本管理接口的 API 测试 | Context: 测试 type=6 的版本列表查询，测试 type=6 的版本添加，测试 type=6 的版本发布，测试 type=6 的版本查询 | Restrictions: 遵循测试规范 | Success: 所有测试通过

- [x] 4.3 手动测试 - 商家管理功能

  - File: -
  - Purpose: 手动测试商家管理功能
  - Requirements: 1.2, 1.3, 1.4
  - Success: 新建商家时可设置 enable_kiosk，编辑商家时可修改 enable_kiosk，商家列表显示 enable_kiosk 字段

- [x] 4.4 手动测试 - 客户端版本管理功能

  - File: -
  - Purpose: 手动测试客户端版本管理功能
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Success: type=6 的版本列表查询正常，版本添加正常，版本发布正常，版本查询正常

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

- [ ] API 文档已更新（接口注释）
- [ ] 数据库文档已更新（迁移脚本）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-kiosk-management/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-kiosk-management/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-kiosk-management/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-kiosk-management/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-kiosk-management/tasks.md)" | bc
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
**最后更新**: 2025-12-05  
**维护者**: 后端开发组

