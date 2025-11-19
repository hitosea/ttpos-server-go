# 云平台商家管理增加桌台地图和数据管理开关 需求文档

> 本文档定义云平台商家管理增加桌台地图和数据管理开关的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-19-table-map-and-data-management-switch.md](../../../team/proposals/2025-11-19-table-map-and-data-management-switch.md) |
| **创建日期**      | 2025-11-19                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

---

## 📋 概述

在云平台的商家管理页面（添加/编辑商家）中新增两个功能开关："桌台地图"和"数据管理"。两个开关默认关闭，管理员可以根据需要开启或关闭。**开关位置：放在"高级票据"选项后面。** 本功能仅包含云平台管理端的开关配置功能，不包含桌台地图和数据管理功能本身的实现。

## 🎯 产品对齐

该功能支持云平台管理需求：
- **提供功能控制能力**：为云平台管理员提供商家功能权限的配置入口
- **灵活管理商家功能**：根据商家需求或套餐类型控制功能可用性
- **降低运营成本**：可以按需开启功能，避免不必要的资源消耗
- **符合 SaaS 平台管理规范**：支持功能开关和权限控制

## 📝 用户故事

**作为** 云平台管理员  
**我想** 在添加/编辑商家时控制桌台地图和数据管理功能的开启/关闭  
**以便于** 灵活管理不同商家的功能权限

---

## 功能需求

### Requirement 1: 桌台地图开关

**用户故事**: 作为云平台管理员，我想在商家管理页面中配置桌台地图功能开关，以便于控制该功能的可用性

#### 验收标准

1. **WHEN** 在添加/编辑商家页面打开"桌台地图"开关 **THEN** 系统 **SHALL** 保存开关状态到数据库，默认状态为关闭
2. **WHEN** 在添加/编辑商家页面关闭"桌台地图"开关 **THEN** 系统 **SHALL** 保存关闭状态到数据库
3. **WHEN** 重新打开商家编辑页面 **THEN** 系统 **SHALL** 正确显示之前保存的开关状态

#### 具体要求

- [ ] 1.1 在商家管理页面（`admin/views/admin/src/pages/merchant/components/dialog-edit.vue`）增加桌台地图开关
- [ ] 1.2 **开关位置：放在"高级票据打印"选项后面**
- [ ] 1.3 开关状态保存到 `company_setting` 表的 `is_open_table_map` 字段
- [ ] 1.4 前端页面显示开关控件，支持开启/关闭（使用 el-radio-group）
- [ ] 1.5 开关状态支持多语言显示

---

### Requirement 2: 数据管理开关

**用户故事**: 作为云平台管理员，我想在商家管理页面中配置数据管理功能开关，以便于控制该功能的可用性

#### 验收标准

1. **WHEN** 在添加/编辑商家页面打开"数据管理"开关 **THEN** 系统 **SHALL** 保存开关状态到数据库，默认状态为关闭
2. **WHEN** 在添加/编辑商家页面关闭"数据管理"开关 **THEN** 系统 **SHALL** 保存关闭状态到数据库
3. **WHEN** 重新打开商家编辑页面 **THEN** 系统 **SHALL** 正确显示之前保存的开关状态

#### 具体要求

- [ ] 2.1 在商家管理页面增加数据管理开关
- [ ] 2.2 **开关位置：放在"桌台地图"开关后面**
- [ ] 2.3 开关状态保存到 `company_setting` 表的 `is_open_data_management` 字段
- [ ] 2.4 前端页面显示开关控件，支持开启/关闭（使用 el-radio-group）
- [ ] 2.5 开关状态支持多语言显示

---

### Requirement 3: 商家管理集成

**用户故事**: 作为云平台管理员，我想在添加/编辑商家时看到并保存这两个开关的状态，以便于管理商家功能权限

#### 验收标准

1. **WHEN** 查看添加/编辑商家页面 **THEN** 系统 **SHALL** 在"高级票据打印"选项后面显示"桌台地图"和"数据管理"两个开关
2. **IF** 新建商家时未设置开关状态 **THEN** 系统 **SHALL** 默认两个开关都为关闭状态
3. **WHEN** 保存商家信息 **THEN** 系统 **SHALL** 同时保存两个开关的状态

#### 具体要求

- [ ] 3.1 在商家编辑表单中正确显示两个开关的位置（高级票据后面）
- [ ] 3.2 开关状态与商家信息一起保存
- [ ] 3.3 新建商家时，两个开关默认值为 0（关闭）
- [ ] 3.4 编辑商家时，正确加载已保存的开关状态

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Model 分层（PHP）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Model 应独立且可复用
- **遵循规范**:
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 在 `company_setting` 表中新增两个字段：`is_open_table_map` 和 `is_open_data_management`
- [ ] 字段类型：`int(11)`，默认值：0（关闭），1（开启）
- [ ] 字段位置：放在 `is_open_advanced_ticket_print` 字段后面
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 商家编辑页面加载时间 < 500ms
- [ ] 商家信息保存响应时间 < 200ms

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] PHP Controller 层测试覆盖核心逻辑
- [ ] 前端组件测试覆盖交互逻辑
- [ ] 集成测试覆盖商家信息保存和读取流程

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `admin/views/admin/src/locales/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 商家信息保存需要权限验证（仅云平台管理员）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 商家信息保存失败时提供错误提示
- [ ] 数据验证失败时提供友好提示
- [ ] 错误日志记录

---

## 验收标准

### 功能验收

1. **桌台地图开关**: 可以开启/关闭桌台地图功能开关，状态正确保存和显示，位置在高级票据后面
2. **数据管理开关**: 可以开启/关闭数据管理功能开关，状态正确保存和显示，位置在桌台地图后面
3. **开关位置**: 两个开关正确显示在"高级票据打印"选项后面
4. **设置持久化**: 重新打开商家编辑页面时，之前保存的配置正确显示
5. **默认值**: 新建商家时，两个开关默认都为关闭状态

### 测试验收

1. **功能测试**: 所有功能需求测试通过
2. **集成测试**: 商家信息保存和读取流程测试通过
3. **浏览器测试**: 主流浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 本功能仅包含云平台管理端的开关配置功能
- 桌台地图和数据管理功能本身的实现将在其他任务中处理
- 开关位置必须放在"高级票据打印"选项后面

### 资源约束

- 开发时间: 1-2 天
- Story Point: 1-2 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- ThinkPHP 6.0
- Vue 3 + Element Plus
- 现有的商家管理模块（`admin/app/admin/controller/Shop.php`）

### 服务依赖

- **Frontend → Admin**: HTTP API 调用

### 业务依赖

- 依赖现有的商家管理功能（`admin/app/admin/controller/Shop.php`）
- 依赖现有的商家编辑页面（`admin/views/admin/src/pages/merchant/components/dialog-edit.vue`）

---

## 风险和缓解

### 风险 1: 需要确认商家表结构，可能需要新增字段

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 先查看现有 `company_setting` 表结构，确定字段添加方式
- 参考现有的 `is_open_advanced_ticket_print` 字段实现方式

### 风险 2: 需要确认现有商家管理页面的代码结构，确保集成顺利

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 参考现有的商家管理功能，复用成熟的代码模式
- 参考高级票据开关的实现方式

---

## 时间表

- **Phase 1 - 数据库迁移**: 0.5 天（创建迁移文件，添加两个字段）
- **Phase 2 - 后端实现**: 0.5 天（PHP Controller 和 Model）
- **Phase 3 - 前端实现**: 0.5 天（Vue 页面和组件）
- **Phase 4 - 测试和联调**: 0.5 天
- **总计**: 1-2 天（SP = 1-2）

---

## 参考资料

### 核心规范

- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 类似功能参考

- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `admin/app/admin/model/supplier/Supplier.php` - 商家 Model（已有 `is_open_advanced_ticket_print` 类似实现）
- `admin/views/admin/src/pages/merchant/components/dialog-edit.vue` - 商家编辑页面（已有高级票据开关实现）
- `admin/database/migrations/20251022094844_add_is_open_advanced_ticket_print_field_to_company_setting.php` - 高级票据字段迁移文件（参考）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: 产品组 + 开发组  
**审核者**: {审核者}

