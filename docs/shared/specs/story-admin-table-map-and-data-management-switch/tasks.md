# 云平台商家管理增加桌台地图和数据管理开关 任务分解

> 本文档定义云平台商家管理增加桌台地图和数据管理开关的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_table_map_and_data_management_fields_to_company_setting.php`
  - Purpose: 在 company_setting 表中添加两个新字段
  - Requirements: 1.3, 2.3
  - Leverage: 现有迁移文件: `admin/database/migrations/20251022094844_add_is_open_advanced_ticket_print_field_to_company_setting.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 company_setting 表中添加 is_open_table_map 和 is_open_data_management 两个字段 | Context: 字段类型为 int(11)，默认值为 0，放在 is_open_advanced_ticket_print 字段后面 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 1.3, 2.3
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [ ] 1.3 更新 Seeds 文件（如需要）

  - File: `admin/database/seeds/saas.sql`, `admin/database/seeds/shop_01.sql`
  - Purpose: 更新 Seeds 文件，包含新字段
  - Requirements: 1.3, 2.3
  - Leverage: 现有 Seeds 文件
  - Success: Seeds 文件更新成功

---

## Phase 2: 后端实现（PHP）

### Model 层

- [ ] 2.1 扩展 Supplier Model，添加新字段处理

  - File: `admin/app/admin/model/supplier/Supplier.php`
  - Purpose: 在 add() 和 edit() 方法中处理新字段
  - Requirements: 1.3, 2.3, 3.3
  - Leverage: 现有的 `is_open_advanced_ticket_print` 字段处理方式
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 Supplier Model 的 add() 和 edit() 方法中添加 is_open_table_map 和 is_open_data_management 字段的处理逻辑 | Context: 参考 is_open_advanced_ticket_print 字段的处理方式，默认值为 0 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 扩展完成，新字段正确保存

- [ ] 2.2 扩展 App Model，添加新字段查询

  - File: `admin/app/admin/model/app/App.php`
  - Purpose: 在 getList() 方法中添加新字段到查询字段列表
  - Requirements: 1.3, 2.3, 3.4
  - Leverage: 现有的 `is_open_advanced_ticket_print` 字段查询方式
  - Prompt: Role: PHP Developer | Task: 在 App Model 的 getList() 方法中添加 is_open_table_map 和 is_open_data_management 字段到查询字段列表 | Context: 参考 is_open_advanced_ticket_print 字段的查询方式 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 扩展完成，新字段正确查询

- [ ] 2.3 更新验证器（如需要）

  - File: `admin/app/admin/validate/AppValidate.php`
  - Purpose: 添加新字段的验证规则（如需要）
  - Requirements: 1.3, 2.3
  - Leverage: 现有的验证规则
  - Success: 验证器更新成功

---

## Phase 3: 前端实现（Vue）

### API 封装

- [ ] 3.1 扩展商家 API 类型定义

  - File: `admin/views/admin/src/api/merchant/index.ts`
  - Purpose: 在商家数据类型中添加新字段
  - Requirements: 1.1, 2.1
  - Leverage: 现有的 `is_open_advanced_ticket_print` 字段定义
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 在商家 API 类型定义中添加 is_open_table_map 和 is_open_data_management 字段 | Context: 参考 is_open_advanced_ticket_print 字段的定义方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 类型定义完成，字段类型正确

### 页面组件

- [ ] 3.2 扩展商家编辑页面，添加桌台地图开关

  - File: `admin/views/admin/src/pages/merchant/components/dialog-edit.vue`
  - Purpose: 在商家编辑表单中添加桌台地图开关
  - Requirements: 1.1, 1.2, 1.4, 3.1
  - Leverage: 现有的高级票据开关实现（第143-148行）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在商家编辑表单中添加桌台地图开关，放在高级票据打印开关后面 | Context: 使用 el-radio-group，参考 is_open_advanced_ticket_print 的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 开关添加成功，位置正确，功能正常

- [ ] 3.3 扩展商家编辑页面，添加数据管理开关

  - File: `admin/views/admin/src/pages/merchant/components/dialog-edit.vue`
  - Purpose: 在商家编辑表单中添加数据管理开关
  - Requirements: 2.1, 2.2, 2.4, 3.1
  - Leverage: 桌台地图开关的实现（Task 3.2）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在商家编辑表单中添加数据管理开关，放在桌台地图开关后面 | Context: 使用 el-radio-group，参考桌台地图开关的实现方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 开关添加成功，位置正确，功能正常

- [ ] 3.4 添加表单数据初始化和验证规则

  - File: `admin/views/admin/src/pages/merchant/components/dialog-edit.vue`
  - Purpose: 在 formData 和 formRules 中添加新字段
  - Requirements: 1.1, 2.1, 3.3
  - Leverage: 现有的 formData 和 formRules 定义（第360-400行）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在 formData 中添加 is_open_table_map 和 is_open_data_management 字段（默认值为 0），在 formRules 中添加验证规则 | Context: 参考 is_open_advanced_ticket_print 字段的定义方式 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 表单数据初始化正确，验证规则正确

- [ ] 3.5 添加多语言支持

  - File: `admin/views/admin/src/locales/*.json`（所有语言文件）
  - Purpose: 添加"桌台地图"和"数据管理"的多语言翻译
  - Requirements: 1.5, 2.5
  - Leverage: 现有的多语言文件，参考"高级票据打印"的翻译
  - Prompt: Role: Frontend Developer | Task: 在所有语言文件中添加"桌台地图"和"数据管理"的翻译 | Context: 参考"高级票据打印"的翻译方式，支持中文、英文、日语、韩语等10种语言 | Restrictions: 遵循多语言规范 | Success: 所有语言文件更新成功

---

## Phase 4: 测试和优化

- [ ] 4.1 功能测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试两个开关的所有功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写功能测试用例，覆盖所有功能需求 | Context: 测试开关开启/关闭、位置显示、设置保存和读取 | Restrictions: 测试覆盖完整 | Success: 所有功能测试通过

- [ ] 4.2 集成测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试商家信息保存和读取的完整流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程：打开商家编辑页面 → 修改开关状态 → 保存 → 重新打开页面验证设置 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.3 浏览器兼容性测试

  - File: -
  - Purpose: 确保在主流浏览器中正常工作
  - Requirements: 浏览器兼容性要求
  - Leverage: 浏览器测试工具
  - Success: Chrome 90+, Safari 14+, Firefox 88+, Edge 90+ 测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
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

- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-admin-table-map-and-data-management-switch/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-admin-table-map-and-data-management-switch/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-admin-table-map-and-data-management-switch/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-admin-table-map-and-data-management-switch/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-admin-table-map-and-data-management-switch/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: PHP 代码格式化，Vue 代码 ESLint 检查
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
**最后更新**: 2025-11-19  
**维护者**: 后端开发组

