# 员工账号增加权限密码 任务分解

> 本文档定义员工账号增加权限密码的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 11  
**进行中**: -  
**完成率**: 73%

---

## Phase 1: 数据库迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/20251121014418_add_permission_password_field_to_staff_table.php`
  - Purpose: 在 ttpos_staff 表中添加 permission_password 字段
  - Requirements: 1.1
  - Leverage: 现有的 password 字段，参考其他字段迁移文件
  - Status: ✅ 已完成，字段类型为 varchar(255) NOT NULL DEFAULT ''，包含默认值设置逻辑

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Seeds 文件（如需要）

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 更新 Seeds 文件，包含新字段
  - Requirements: 1.1
  - Leverage: 现有 Seeds 文件
  - Status: ✅ 已完成，字段定义为 VARCHAR(255) NOT NULL DEFAULT ''

---

## Phase 2: Go新管理端实现

### Model 层

- [x] 2.1 扩展 Staff Model，添加权限密码字段

  - File: `main/app/model/staff.go`
  - Purpose: 在 Staff 结构体中添加 PermissionPassword 字段
  - Requirements: 1.2
  - Leverage: 现有的 Password 字段定义
  - Status: ✅ 已完成，字段类型为 string，gorm 标签为 permission_password，json 标签为 "-"（不返回给前端）

### DTO 层

- [x] 2.2 扩展 DTO，添加权限密码字段

  - File: `main/app/dto/req/staff.go`
  - Purpose: 在 AddStaffReq 和 UpdateStaffReq 中添加 PermissionPassword 字段
  - Requirements: 1.3, 4.3, 4.4
  - Leverage: 现有的 Password 字段定义
  - Status: ✅ 已完成，AddStaffReq 中必填，UpdateStaffReq 中非必填

- [x] 2.3 创建自定义验证器（4-8位数字验证）

  - File: `main/pkg/validator/validation.go` 和 `main/pkg/validator/validator.go`
  - Purpose: 创建权限密码格式验证器
  - Requirements: 1.3, 3.1
  - Leverage: 现有的验证器实现
  - Status: ✅ 已完成，验证器 `permissionPassword` 已创建并注册

### Service 层

- [x] 2.4 扩展 AddStaff 方法，添加权限密码处理

  - File: `main/app/service/staff.go`
  - Purpose: 在 AddStaff 方法中处理权限密码字段
  - Requirements: 1.4, 4.4
  - Leverage: 现有的 Password 字段处理方式
  - Status: ✅ 已完成，权限密码使用 utils.EncryptPassword() 加密后存储

- [x] 2.5 扩展 UpdateStaff 方法，添加权限密码处理

  - File: `main/app/service/staff.go`
  - Purpose: 在 UpdateStaff 方法中处理权限密码字段
  - Requirements: 1.4, 4.4
  - Leverage: 现有的 Password 字段处理方式
  - Status: ✅ 已完成，权限密码仅在设置了新值时才更新

### API 层

- [x] 2.6 验证 API 接口（无需修改）

  - File: `main/app/api/v1/shop/shop_staff.go`
  - Purpose: 验证现有 API 接口是否支持新字段
  - Requirements: 1.5
  - Leverage: 现有的 AddStaff 和 UpdateStaff API
  - Status: ✅ 已验证，API 接口通过 DTO 绑定自动支持新字段

---

## Phase 3: PHP商家后台实现

### Model 层

- [x] 3.1 扩展 User Model，添加权限密码字段处理（add方法）

  - File: `admin/app/shop/model/auth/User.php`
  - Purpose: 在 add() 方法中处理权限密码字段
  - Requirements: 2.1, 4.4
  - Leverage: 现有的 password 字段处理方式
  - Status: ✅ 已完成，权限密码使用 salt_hash() 函数加密后存储

- [x] 3.2 扩展 User Model，添加权限密码字段处理（edit方法）

  - File: `admin/app/shop/model/auth/User.php`
  - Purpose: 在 edit() 方法中处理权限密码字段
  - Requirements: 2.1, 4.4
  - Leverage: 现有的 password 字段处理方式
  - Status: ✅ 已完成，权限密码仅在设置了新值时才更新

- [x] 3.3 添加权限密码格式验证

  - File: `admin/extend/help/ValidateHelp.php`
  - Purpose: 验证权限密码格式（4-8位数字）
  - Requirements: 2.2, 3.2
  - Leverage: 现有的密码验证逻辑
  - Status: ✅ 已完成，添加了 validatePermissionPassword() 方法，使用正则表达式 `/^\d{4,8}$/` 验证

---

## Phase 4: 前端实现

### 新管理端前端（如适用）

- [ ] 4.1 添加权限密码输入框（新管理端）

  - File: `admin/views/admin/src/pages/`（员工管理页面，如适用）
  - Purpose: 在员工编辑表单中添加权限密码输入框
  - Requirements: 1.1, 3.1, 4.1, 4.2, 4.3
  - Leverage: 现有的密码输入框实现
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在员工编辑表单中添加权限密码输入框 | Context: 使用 el-input，type 为 password，新建时必填（显示必填标识），编辑时非必填（显示为空，不回显原密码），placeholder 根据编辑模式显示不同提示（新建："请输入权限密码（必填）"，编辑："留空则不修改原权限密码"），添加格式提示"密码必须为 4 - 8 位数字" | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 输入框添加成功，新建时必填，编辑时非必填

- [ ] 4.2 添加密码格式验证（新管理端）

  - File: `admin/views/admin/src/pages/`（员工管理页面，如适用）
  - Purpose: 在表单验证规则中添加权限密码格式验证
  - Requirements: 3.1
  - Leverage: 现有的表单验证规则
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 添加权限密码格式验证规则 | Context: 使用正则表达式 `/^\d{4,8}$/` 验证，验证失败提示"密码必须为 4 - 8 位数字" | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 验证规则正确，错误提示友好

### 商家后台前端

- [x] 4.3 添加权限密码输入框（商家后台）

  - File: `admin/views/shop/src/views/auth/user/dialog/Add.vue` 和 `Edit.vue`
  - Purpose: 在员工编辑表单中添加权限密码输入框
  - Requirements: 2.1, 3.1, 4.1, 4.2, 4.3
  - Leverage: 现有的密码输入框实现
  - Status: ✅ 已完成，新建时必填，编辑时非必填，不回显原密码

- [x] 4.4 添加密码格式验证（商家后台）

  - File: `admin/views/shop/src/views/auth/user/dialog/Add.vue` 和 `Edit.vue`
  - Purpose: 在表单验证规则中添加权限密码格式验证
  - Requirements: 3.1
  - Leverage: 现有的表单验证规则
  - Status: ✅ 已完成，使用正则表达式 `/^\d{4,8}$/` 验证

- [ ] 4.5 添加多语言支持

  - File: `admin/views/shop/src/locales/*.json`（所有语言文件）
  - Purpose: 添加"权限密码"和"密码必须为 4 - 8 位数字"的多语言翻译
  - Requirements: 1.5, 2.5
  - Leverage: 现有的多语言文件
  - Prompt: Role: Frontend Developer | Task: 在所有语言文件中添加"权限密码"和"密码必须为 4 - 8 位数字"的翻译 | Context: 支持中文、英文、日语、韩语等10种语言 | Restrictions: 遵循多语言规范 | Success: 所有语言文件更新成功

---

## Phase 5: 测试和优化

- [ ] 5.1 Go新管理端功能测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试新管理端的权限密码功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写功能测试用例，覆盖新管理端的所有功能需求 | Context: 测试添加员工、编辑员工、密码格式验证、默认值处理 | Restrictions: 测试覆盖完整 | Success: 所有功能测试通过

- [ ] 5.2 PHP商家后台功能测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试商家后台的权限密码功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写功能测试用例，覆盖商家后台的所有功能需求 | Context: 测试添加员工、编辑员工、密码格式验证、默认值处理 | Restrictions: 测试覆盖完整 | Success: 所有功能测试通过

- [ ] 5.3 集成测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试两个终端的完整流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程：打开员工编辑页面 → 设置权限密码 → 保存 → 重新打开页面验证（不显示明文） | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.4 两个终端一致性测试

  - File: -
  - Purpose: 确保新管理端和商家后台的行为一致
  - Requirements: 所有功能需求
  - Leverage: 功能测试用例
  - Success: 两个终端的行为一致，数据格式一致

- [ ] 5.5 浏览器兼容性测试

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
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 测试覆盖率达标

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 两个终端（新管理端和商家后台）功能一致

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/php.mdc`
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-shop-staff-permission-password/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-shop-staff-permission-password/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-shop-staff-permission-password/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-shop-staff-permission-password/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-shop-staff-permission-password/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: Go 代码格式化，PHP 代码格式化，Vue 代码 ESLint 检查
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

