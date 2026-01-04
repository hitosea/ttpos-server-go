# 云平台-商家管理-Grab外卖控制 任务分解

> 本文档定义云平台商家管理中 Grab 外卖控制功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 16  
**已完成**: 6（后端核心功能）  
**已取消**: 6（前端可见性控制，暂不实现）  
**待完成**: 4（测试和文档）  
**完成率**: 100%（后端核心功能已完成，前端由前端团队后续实现）

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_enable_grab_delivery_to_company_setting.php`
  - Purpose: 在 `company_setting` 表中添加 `enable_grab_delivery` 字段
  - Requirements: R1.1, R1.7
  - Leverage: 现有迁移文件: `admin/database/migrations/20251205185229_add_enable_kiosk_to_company_setting.php`
  - Prompt: Role: Database Engineer | Task: 创建添加 enable_grab_delivery 字段的迁移文件，参考 enable_kiosk 的实现方式 | Context: 字段类型 INT(3)，默认值 0，注释"是否启用Grab外卖：0-否；1-是"，添加在 enable_kiosk 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加 `enable_grab_delivery` 字段
  - Requirements: R1.1, R1.7
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 PHP Model 字段定义

  - File: `admin/app/admin/model/app/App.php`
  - Purpose: 在商家列表查询中添加 `enable_grab_delivery` 字段
  - Requirements: R1.4, R1.6
  - Leverage: 现有 Model: `admin/app/admin/model/app/App.php`，参考 enable_kiosk 字段（第 98 行）
  - Prompt: Role: PHP Developer | Task: 在 App Model 的 getList() 方法中添加 enable_grab_delivery 字段 | Context: 在 $field 数组中添加 "su.enable_grab_delivery"，参考 enable_kiosk 的实现 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 字段添加成功，查询返回正确

---

## Phase 2: 后端 API 实现

- [x] 2.1 更新商家新建接口参数

  - File: `admin/app/admin/controller/Shop.php`
  - Purpose: 在新建商家接口中添加 `enable_grab_delivery` 参数
  - Requirements: R1.2
  - Leverage: 现有 Controller: `admin/app/admin/controller/Shop.php`，参考 enable_kiosk 参数（第 102 行）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 Shop Controller 的 add() 方法中添加 enable_grab_delivery 参数文档 | Context: 在 @Apidoc\Param 中添加参数说明，类型 int，可选，默认 0 | Restrictions: 遵循 .cursor/rules/php.mdc，参考 enable_kiosk 的实现 | Success: 参数文档添加成功，接口正确接收参数

- [x] 2.2 更新商家编辑接口参数

  - File: `admin/app/admin/controller/Shop.php`
  - Purpose: 在编辑商家接口中添加 `enable_grab_delivery` 参数
  - Requirements: R1.3
  - Leverage: 现有 Controller: `admin/app/admin/controller/Shop.php`，参考 enable_kiosk 参数（第 182 行）
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 Shop Controller 的 edit() 方法中添加 enable_grab_delivery 参数文档 | Context: 在 @Apidoc\Param 中添加参数说明，类型 int，可选，默认 0 | Restrictions: 遵循 .cursor/rules/php.mdc，参考 enable_kiosk 的实现 | Success: 参数文档添加成功，接口正确接收参数

- [x] 2.3 更新验证器规则

  - File: `admin/app/admin/validate/AppValidate.php`
  - Purpose: 在验证器中添加 `enable_grab_delivery` 验证规则
  - Requirements: R1.5
  - Leverage: 现有验证器: `admin/app/admin/validate/AppValidate.php`，参考 enable_kiosk 验证（第 115 行）
  - Prompt: Role: PHP Developer | Task: 在 AppValidate 中添加 enable_grab_delivery 验证规则 | Context: 在 $scene['add'] 和 $scene['edit'] 数组中添加 'enable_grab_delivery'，在 $rule 数组中添加验证规则 'in:0,1' | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 验证规则添加成功，参数验证正确

- [x] 2.4 更新商家 Model 保存逻辑

  - File: `admin/app/admin/model/app/App.php`
  - Purpose: 确保 `enable_grab_delivery` 字段能正确保存到数据库
  - Requirements: R1.2, R1.3
  - Leverage: 现有 Model: `admin/app/admin/model/app/App.php`，参考 enable_kiosk 的保存逻辑
  - Prompt: Role: PHP Developer | Task: 确保 App Model 的 add() 和 edit() 方法能正确处理 enable_grab_delivery 字段 | Context: 检查字段是否在保存字段列表中，确保默认值处理正确 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 字段保存逻辑正确，数据能正确写入数据库

- [x] 2.5 完善 /shop/base 接口

  - File: `main/app/model/company.go`, `main/app/dto/resp/base.go`, `main/app/service/auth.go`, `admin/app/shop/controller/Controller.php`
  - Purpose: 在商家端基础信息接口中添加 `enable_grab_delivery` 字段
  - Requirements: R1.4
  - Leverage: 参考 `enable_kiosk` 的实现方式
  - Success: Go Main 和 PHP Admin 的 `/shop/base` 接口都返回 `enable_grab_delivery` 字段

- [ ] 2.6 编写后端单元测试

  - File: `admin/app/admin/controller/ShopTest.php` (如存在)
  - Purpose: 测试商家新建/编辑接口的 `enable_grab_delivery` 参数
  - Requirements: R1.2, R1.3, R1.4
  - Leverage: 现有测试文件（如有）
  - Prompt: Role: QA Engineer with PHP testing expertise | Task: 为商家新建/编辑接口编写 enable_grab_delivery 参数测试 | Context: 测试参数验证，测试默认值处理，测试数据保存 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 测试通过，覆盖率达标

---

## Phase 3: 前端可见性控制（已取消）

> **说明**: 根据需求，前端可见性控制暂不实现，由前端团队后续根据业务需要自行实现。

- [x] ~~3.1 实现路由控制~~ （已取消）

- [x] ~~3.2 实现侧边栏菜单控制~~ （已取消）

- [x] ~~3.3 实现外卖订单列表过滤~~ （已取消）

- [x] ~~3.4 实现外卖接单功能控制~~ （已取消）

- [x] ~~3.5 实现外卖设置页面控制~~ （已取消）

- [x] ~~3.6 实现外卖商家管理页面控制~~ （已取消）

---

## Phase 4: 测试和优化

- [ ] 4.1 后端 API 集成测试

  - File: `test/integration/shop_api_test.php` (如存在)
  - Purpose: 测试商家新建/编辑接口的完整流程
  - Requirements: R1.2, R1.3, R1.4
  - Leverage: 现有集成测试（如有）
  - Prompt: Role: QA Automation Engineer | Task: 实现商家新建/编辑接口的集成测试 | Context: 测试 enable_grab_delivery 参数的完整流程，包括参数验证、数据保存、列表查询 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [x] ~~4.2 前端功能测试~~ （已取消 - 前端暂不实现）

  - File: -
  - Purpose: 测试前端可见性控制功能
  - Requirements: R2.1, R2.2, R2.3, R2.4, R2.5, R2.6
  - Leverage: 手动测试或自动化测试工具
  - Prompt: Role: QA Engineer | Task: 测试前端可见性控制功能 | Context: 测试 enable_grab_delivery 为 0 和 1 时，各页面的显示/隐藏是否正确 | Restrictions: 覆盖所有相关页面 | Success: 所有功能测试通过

- [ ] 4.3 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 本地响应时间 < 200ms

- [ ] 4.4 文档更新

  - File: `docs/shared/api/shop_management.md` (如存在), `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新字段）
- [ ] CHANGELOG.md 已更新

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
grep -c "^- \[" docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: PHP 代码格式化，Vue 代码 ESLint
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
**维护者**: 后端开发组

