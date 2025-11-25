# 业务设置增加敏感操作设置 任务分解

> 本文档定义业务设置增加敏感操作设置的详细执行任务清单。**本功能需要在两个终端实现：1) 新管理端（Go项目的shop目录）；2) (旧)商家后台（PHP项目的admin/shop目录）。**

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 13  
**进行中**: -  
**完成率**: 87%

**说明**: 新管理端无前端，仅实现后端接口；商家后台实现后端接口和Vue前端。

---

## Phase 1: Go新管理端后端实现

### DTO 层

- [x] 1.1 扩展 UpdateBusinessSetting DTO，添加敏感操作设置字段

  - File: `main/app/dto/req/base.go`
  - Purpose: 在业务设置 DTO 中添加敏感操作设置字段
  - Requirements: 1.1, 2.1, 3.3, 4.3
  - Leverage: 现有的 `UpdateBusinessSetting` 结构体，参考 `IsNeedPassword` 字段的定义
  - Prompt: Role: Go Developer | Task: 在 UpdateBusinessSetting 结构体中添加敏感操作设置字段 | Context: 需要添加 DiscountNeedPassword (string, oneof=0 1), DiscountAuthorizedStaffIds ([]uint64), RefundNeedPassword (string, oneof=0 1), RefundAuthorizedStaffIds ([]uint64) 四个字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 binding tag 进行验证 | Success: DTO 扩展完成，字段定义正确，验证规则正确

### Service 层

- [x] 1.2 扩展 Setting Service，支持敏感操作设置字段

  - File: `main/app/service/setting/setting.go`
  - Purpose: 在 EditBusinessSetting 方法中处理新增的敏感操作设置字段
  - Requirements: 1.2, 2.2, 3.3, 4.3
  - Leverage: 现有的 `EditBusinessSetting` 方法，使用 `copier.CopyWithOption` 自动复制字段
  - Prompt: Role: Go Developer | Task: 确保 EditBusinessSetting 方法正确处理新增的敏感操作设置字段 | Context: 新增字段会通过 copier.CopyWithOption 自动复制到 businessSetting 中，无需额外处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口 | Success: Service 扩展完成，新字段正确保存到数据库

- [x] 1.3 验证授权员工ID有效性（Go）

  - File: `main/app/service/setting/setting.go`
  - Purpose: 确保保存的授权员工ID在系统中存在
  - Requirements: 3.3, 4.3
  - Leverage: `main/app/service/staff.go` - 员工 Service，或 `main/app/model/staff.go` - 员工模型
  - Prompt: Role: Go Developer | Task: 在保存设置前验证授权员工ID是否存在 | Context: 遍历 DiscountAuthorizedStaffIds 和 RefundAuthorizedStaffIds，检查每个ID是否在 staff 表中存在 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 Repository 查询数据库 | Success: 无效的员工ID被过滤，仅保存有效的ID

---

## Phase 2: PHP商家后台后端实现

### Controller 层

- [x] 1.1 扩展 Business Controller，添加敏感操作设置字段处理

  - File: `admin/app/shop/controller/setting/Business.php`
  - Purpose: 在业务设置保存时处理新增的敏感操作设置字段
  - Requirements: 1.1, 2.1, 3.3, 4.3
  - Leverage: 现有 Business Controller 的 `index()` 方法，参考 `is_need_password` 字段的处理方式
  - Prompt: Role: PHP Developer with ThinkPHP expertise | Task: 在 Business Controller 的 index() 方法中添加敏感操作设置字段的处理逻辑 | Context: 需要处理 discount_need_password, discount_authorized_staff_ids, refund_need_password, refund_authorized_staff_ids 四个字段，参考现有的 is_need_password 字段处理方式 | Restrictions: 遵循 .cursor/rules/php.mdc，Controller 不写业务逻辑，只做参数处理和调用 Model | Success: Controller 扩展完成，新字段正确保存到数据库

- [x] 1.2 添加参数验证

  - File: `admin/app/shop/controller/setting/Business.php`
  - Purpose: 验证敏感操作设置字段的合法性
  - Requirements: 1.1, 2.1, 3.3, 4.3
  - Leverage: 现有参数验证逻辑
  - Prompt: Role: PHP Developer | Task: 添加敏感操作设置字段的参数验证 | Context: discount_need_password 和 refund_need_password 必须是 0 或 1，discount_authorized_staff_ids 和 refund_authorized_staff_ids 必须是数组 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 参数验证正确，无效数据被过滤

- [x] 1.3 验证授权员工ID有效性

  - File: `admin/app/shop/controller/setting/Business.php`
  - Purpose: 确保保存的授权员工ID在系统中存在
  - Requirements: 3.3, 4.3
  - Leverage: `admin/app/shop/model/auth/User.php` - 员工模型
  - Prompt: Role: PHP Developer | Task: 在保存设置前验证授权员工ID是否存在 | Context: 遍历 discount_authorized_staff_ids 和 refund_authorized_staff_ids，检查每个ID是否在 User 表中存在 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: 无效的员工ID被过滤，仅保存有效的ID

---

## Phase 3: 前端实现（Vue - 仅商家后台）

**注意**: 新管理端无前端，仅实现后端接口。本阶段仅涉及商家后台的Vue前端开发。

### API 封装

- [x] 3.1 扩展商家后台业务设置 API 封装

  - File: `admin/views/shop/src/api/setting.js`（或相应文件）
  - Purpose: 封装商家后台业务设置获取和保存的 API 调用
  - Requirements: 1B.1, 2B.1, 3.1, 4.1
  - Leverage: 现有的商家后台业务设置 API 封装
  - Prompt: Role: Frontend Developer with TypeScript expertise | Task: 扩展商家后台业务设置 API，支持敏感操作设置字段 | Context: 在 getBusinessSetting 和 saveBusinessSetting 方法中处理新字段 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: API 封装完成，类型定义正确

### 页面组件（商家后台）

- [x] 3.2 扩展商家后台业务设置页面，添加敏感操作设置区域

  - File: `admin/views/shop/src/pages/setting/business/index.vue`（或相应文件）
  - Purpose: 在商家后台业务设置页面中添加敏感操作设置区域
  - Requirements: 1B.3, 2B.3, 3.1, 4.1
  - Leverage: 现有的业务设置页面结构
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在业务设置页面中添加敏感操作设置卡片区域 | Context: 使用 Element Plus 的 Card 组件，参考现有设置项的布局 | Restrictions: 遵循 .cursor/rules/vue.mdc，使用 Composition API | Success: 页面结构正确，样式统一

- [x] 3.3 实现折扣权限开关（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 实现折扣操作权限验证开关
  - Requirements: 1B.3
  - Leverage: Element Plus 的 Radio 组件，参考 `is_need_password` 的实现
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现折扣权限验证开关，支持开启/关闭 | Context: 使用 el-radio-group，选项为"需要密码"和"无需密码"，参考 is_need_password 的实现 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 开关功能正常，状态正确保存和显示

- [x] 3.4 实现退款权限开关（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 实现退款操作权限验证开关
  - Requirements: 2B.3
  - Leverage: Element Plus 的 Radio 组件，参考折扣权限开关的实现
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现退款权限验证开关，支持开启/关闭 | Context: 使用 el-radio-group，选项为"需要密码"和"无需密码"，参考折扣权限开关的实现 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 开关功能正常，状态正确保存和显示

- [x] 3.5 实现折扣授权员工选择器（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 实现折扣操作授权员工多选选择器
  - Requirements: 3.1, 3.2, 3.4
  - Leverage: Element Plus 的 Select 组件（支持多选和搜索）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现折扣授权员工多选选择器，支持搜索和筛选 | Context: 使用 el-select 组件，设置 multiple 和 filterable 属性，仅在 discount_need_password 为 1 时显示 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 选择器功能正常，支持多选和搜索，已选员工正确显示

- [x] 3.6 实现退款授权员工选择器（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 实现退款操作授权员工多选选择器
  - Requirements: 4.1, 4.2, 4.4
  - Leverage: Element Plus 的 Select 组件，参考折扣授权员工选择器的实现
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现退款授权员工多选选择器，支持搜索和筛选 | Context: 使用 el-select 组件，设置 multiple 和 filterable 属性，仅在 refund_need_password 为 1 时显示，参考折扣授权员工选择器的实现 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 选择器功能正常，支持多选和搜索，已选员工正确显示

- [x] 3.7 加载员工列表（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 在页面加载时获取员工列表，供选择器使用
  - Requirements: 3.2, 4.2
  - Leverage: 现有的员工列表 API（`admin/views/shop/src/api/auth.js` 或相应文件）
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 在页面加载时调用员工列表 API，获取员工数据 | Context: 使用 onMounted 钩子，调用员工列表 API，将数据存储到响应式变量中 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 员工列表正确加载，选择器可以正常使用

- [x] 3.8 实现设置保存逻辑（商家后台）

  - File: `admin/views/shop/src/pages/setting/business/index.vue`
  - Purpose: 实现敏感操作设置的保存逻辑
  - Requirements: 1B.1, 2B.1, 3.3, 4.3
  - Leverage: 现有的业务设置保存逻辑
  - Prompt: Role: Frontend Developer with Vue 3 expertise | Task: 实现敏感操作设置的保存逻辑，包含表单验证和错误处理 | Context: 在保存时收集所有设置字段（包括新增的敏感操作设置字段），调用保存 API，处理成功和失败情况 | Restrictions: 遵循 .cursor/rules/vue.mdc | Success: 设置保存成功，错误处理正确

---

## Phase 4: 测试和优化

- [ ] 4.1 功能测试（新管理端API和商家后台）

  - File: `test/`（或相应测试目录）
  - Purpose: 测试敏感操作设置的所有功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写功能测试用例，覆盖所有功能需求 | Context: 测试新管理端API接口（保存和查询），测试商家后台前端+后端（开关开启/关闭、授权员工选择、设置保存和读取） | Restrictions: 测试覆盖完整 | Success: 所有功能测试通过

- [ ] 4.2 集成测试（新管理端API和商家后台）

  - File: `test/`（或相应测试目录）
  - Purpose: 测试设置保存和读取的完整流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试新管理端API：调用保存接口 → 调用查询接口验证；测试商家后台：打开设置页面 → 修改敏感操作设置 → 保存 → 重新打开页面验证设置 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.3 两个终端一致性测试

  - File: `test/`（或相应测试目录）
  - Purpose: 确保新管理端API和商家后台的行为一致，数据格式一致
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 测试两个终端的一致性 | Context: 通过新管理端API保存设置后，在商家后台读取应显示相同数据；在商家后台保存设置后，通过新管理端API读取应显示相同数据 | Restrictions: 测试数据一致性 | Success: 两个终端数据格式一致，API行为一致

- [ ] 4.4 浏览器兼容性测试

  - File: -
  - Purpose: 确保在主流浏览器中正常工作
  - Requirements: 浏览器兼容性要求
  - Leverage: 浏览器测试工具
  - Success: Chrome 90+, Safari 14+, Firefox 88+, Edge 90+ 测试通过

- [ ] 4.5 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 设置页面加载时间 < 500ms，设置保存响应时间 < 200ms

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 gofmt 格式化，通过 golangci-lint 检查
- [ ] PHP 代码通过 PSR-2 格式化
- [ ] Vue 代码通过 ESLint 检查
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`（Go新管理端）
- [ ] 遵循 `.cursor/rules/php.mdc`（PHP商家后台）
- [ ] 遵循 `.cursor/rules/vue.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-sensitive-operation-settings/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-sensitive-operation-settings/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-sensitive-operation-settings/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-sensitive-operation-settings/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-sensitive-operation-settings/tasks.md)" | bc
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

