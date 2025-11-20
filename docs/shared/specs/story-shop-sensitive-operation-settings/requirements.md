# 商家后台业务设置增加敏感操作设置 需求文档

> 本文档定义商家后台业务设置增加敏感操作设置的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-19-sensitive-operation-settings.md](../../../team/proposals/2025-11-19-sensitive-operation-settings.md) |
| **创建日期**      | 2025-11-19                                                                                                 |
| **负责人**        | 曾振华、何翔                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [x] Vue (admin/views/)                                   |

---

## 📋 概述

在业务设置中新增"敏感操作设置"模块，支持配置折扣和退款操作是否需要权限密码验证，以及选择授权验证人。**本功能需要在两个终端实现：1) 新管理端（Go项目的shop目录：`main/app/api/v1/shop/`）- 仅实现后端接口，无前端；2) (旧)商家后台（PHP项目的admin/shop目录：`admin/app/shop/`）- 实现后端接口和Vue前端。** 本功能仅包含业务设置页面的配置功能，密码验证逻辑将在其他任务中实现。

## 🎯 产品对齐

该功能支持门店管理需求：
- **提升财务安全性**：为门店经理/店长提供敏感操作权限配置的入口
- **增强管理控制**：门店经理/店长可以指定授权验证人
- **为后续验证功能打基础**：为密码验证功能提供配置数据支持
- **符合餐饮行业管理规范**：支持重要操作需要上级授权的管理需求

## 📝 用户故事

**作为** 门店经理/店长  
**我想** 在业务设置中配置折扣/退款操作的权限验证  
**以便于** 控制普通员工进行敏感操作时的权限，降低财务风险

---

## 功能需求

### Requirement 1: 折扣操作权限设置（新管理端Go）

**用户故事**: 作为门店经理/店长，我想通过新管理端的业务设置接口配置折扣操作是否需要密码验证，以便于控制折扣操作的权限

#### 验收标准

1. **WHEN** 调用新管理端业务设置保存接口，传入折扣权限验证开关参数 **THEN** 系统 **SHALL** 保存开关状态到数据库
2. **WHEN** 调用新管理端业务设置查询接口 **THEN** 系统 **SHALL** 正确返回之前保存的折扣权限验证开关状态
3. **WHEN** 传入无效的开关值（非0或1） **THEN** 系统 **SHALL** 返回参数验证错误

#### 具体要求

- [ ] 1.1 在 Go API（`main/app/api/v1/shop/shop_setting.go`）中扩展业务设置保存接口，支持折扣权限验证开关字段
- [ ] 1.2 在 Go Service（`main/app/service/setting/setting.go`）中添加折扣权限验证开关处理逻辑
- [ ] 1.3 在 Go DTO（`main/app/dto/req/base.go`）中添加折扣权限验证开关字段，包含参数验证
- [ ] 1.4 开关状态保存到 `setting` 表的 `business` 配置中（JSON 格式）

---

### Requirement 1B: 折扣操作权限设置（商家后台PHP）

**用户故事**: 作为门店经理/店长，我想在商家后台的业务设置中配置折扣操作是否需要密码验证，以便于控制折扣操作的权限

#### 验收标准

1. **WHEN** 在商家后台业务设置页面打开"折扣"权限验证开关 **THEN** 系统 **SHALL** 保存开关状态到数据库
2. **WHEN** 在商家后台业务设置页面关闭"折扣"权限验证开关 **THEN** 系统 **SHALL** 保存关闭状态到数据库
3. **WHEN** 重新打开设置页面 **THEN** 系统 **SHALL** 正确显示之前保存的开关状态

#### 具体要求

- [ ] 1B.1 在业务设置页面（`admin/app/shop/controller/setting/Business.php`）增加折扣权限验证开关
- [ ] 1B.2 开关状态保存到 `setting` 表的 `business` 配置中（JSON 格式）
- [ ] 1B.3 前端页面（Vue）显示开关控件，支持开启/关闭
- [ ] 1B.4 开关状态支持多语言显示

---

### Requirement 2: 退款操作权限设置（新管理端Go）

**用户故事**: 作为门店经理/店长，我想通过新管理端的业务设置接口配置退款操作是否需要密码验证，以便于控制退款操作的权限

#### 验收标准

1. **WHEN** 调用新管理端业务设置保存接口，传入退款权限验证开关参数 **THEN** 系统 **SHALL** 保存开关状态到数据库
2. **WHEN** 调用新管理端业务设置查询接口 **THEN** 系统 **SHALL** 正确返回之前保存的退款权限验证开关状态
3. **WHEN** 传入无效的开关值（非0或1） **THEN** 系统 **SHALL** 返回参数验证错误

#### 具体要求

- [ ] 2.1 在 Go API（`main/app/api/v1/shop/shop_setting.go`）中扩展业务设置保存接口，支持退款权限验证开关字段
- [ ] 2.2 在 Go Service（`main/app/service/setting/setting.go`）中添加退款权限验证开关处理逻辑
- [ ] 2.3 在 Go DTO（`main/app/dto/req/base.go`）中添加退款权限验证开关字段，包含参数验证
- [ ] 2.4 开关状态保存到 `setting` 表的 `business` 配置中（JSON 格式）

---

### Requirement 2B: 退款操作权限设置（商家后台PHP）

**用户故事**: 作为门店经理/店长，我想在商家后台的业务设置中配置退款操作是否需要密码验证，以便于控制退款操作的权限

#### 验收标准

1. **WHEN** 在商家后台业务设置页面打开"退款"权限验证开关 **THEN** 系统 **SHALL** 保存开关状态到数据库
2. **WHEN** 在商家后台业务设置页面关闭"退款"权限验证开关 **THEN** 系统 **SHALL** 保存关闭状态到数据库
3. **WHEN** 重新打开设置页面 **THEN** 系统 **SHALL** 正确显示之前保存的开关状态

#### 具体要求

- [ ] 2B.1 在业务设置页面增加退款权限验证开关
- [ ] 2B.2 开关状态保存到 `setting` 表的 `business` 配置中（JSON 格式）
- [ ] 2B.3 前端页面显示开关控件，支持开启/关闭
- [ ] 2B.4 开关状态支持多语言显示

---

### Requirement 3: 授权员工选择（折扣）

**用户故事**: 作为门店经理/店长，我想为折扣操作选择授权验证人，以便于指定哪些员工可以授权折扣操作

#### 验收标准

1. **IF** 调用业务设置保存接口，传入授权员工ID列表 **THEN** 系统 **SHALL** 将选中的员工ID列表保存到数据库
2. **IF** 调用业务设置查询接口 **THEN** 系统 **SHALL** 正确返回之前保存的授权员工ID列表
3. **WHEN** 传入无效的员工ID **THEN** 系统 **SHALL** 过滤无效ID，仅保存有效的员工ID

#### 具体要求

- [ ] 3.1 在商家后台的业务设置页面增加"折扣授权员工"选择器（多选，Vue前端）
- [ ] 3.2 选择器支持搜索和筛选员工（Vue前端）
- [ ] 3.3 选中的员工ID列表保存到 `setting` 表的 `business` 配置中（JSON 数组格式）
- [ ] 3.4 前端页面显示已选择的授权员工列表（商家后台Vue前端）
- [ ] 3.5 Go API 和 PHP API 都需要支持授权员工ID列表的保存和查询
- [ ] 3.6 Go API 和 PHP API 都需要验证授权员工ID的有效性

---

### Requirement 4: 授权员工选择（退款）

**用户故事**: 作为门店经理/店长，我想为退款操作选择授权验证人，以便于指定哪些员工可以授权退款操作

#### 验收标准

1. **IF** 调用业务设置保存接口，传入授权员工ID列表 **THEN** 系统 **SHALL** 将选中的员工ID列表保存到数据库
2. **IF** 调用业务设置查询接口 **THEN** 系统 **SHALL** 正确返回之前保存的授权员工ID列表
3. **WHEN** 传入无效的员工ID **THEN** 系统 **SHALL** 过滤无效ID，仅保存有效的员工ID

#### 具体要求

- [ ] 4.1 在商家后台的业务设置页面增加"退款授权员工"选择器（多选，Vue前端）
- [ ] 4.2 选择器支持搜索和筛选员工（Vue前端）
- [ ] 4.3 选中的员工ID列表保存到 `setting` 表的 `business` 配置中（JSON 数组格式）
- [ ] 4.4 前端页面显示已选择的授权员工列表（商家后台Vue前端）
- [ ] 4.5 Go API 和 PHP API 都需要支持授权员工ID列表的保存和查询
- [ ] 4.6 Go API 和 PHP API 都需要验证授权员工ID的有效性

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层（Go）和 Controller → Model 分层（PHP）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Model 应独立且可复用
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 使用现有的 `setting` 表，key 为 `business`
- [ ] values 字段使用 JSON 格式存储配置
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 设置页面加载时间 < 500ms
- [ ] 设置保存响应时间 < 200ms

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Go Service 层测试覆盖率 ≥ 70%
- [ ] PHP Controller 层测试覆盖核心逻辑
- [ ] 前端组件测试覆盖交互逻辑
- [ ] 集成测试覆盖设置保存和读取流程
- [ ] 两个终端（新管理端和商家后台）的一致性测试

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `admin/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 设置保存需要权限验证（仅门店管理员）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 设置保存失败时提供错误提示
- [ ] 数据验证失败时提供友好提示
- [ ] 错误日志记录

---

## 验收标准

### 功能验收

1. **新管理端折扣权限设置接口**: 可以通过API接口保存和查询折扣操作密码验证开关，状态正确保存和返回
2. **商家后台折扣权限设置**: 可以开启/关闭折扣操作密码验证开关，状态正确保存和显示（前端+后端）
3. **新管理端退款权限设置接口**: 可以通过API接口保存和查询退款操作密码验证开关，状态正确保存和返回
4. **商家后台退款权限设置**: 可以开启/关闭退款操作密码验证开关，状态正确保存和显示（前端+后端）
5. **授权员工选择（折扣）**: 可以多选员工作为折扣操作的授权验证人，员工列表正确保存和显示（商家后台前端+后端接口）
6. **授权员工选择（退款）**: 可以多选员工作为退款操作的授权验证人，员工列表正确保存和显示（商家后台前端+后端接口）
7. **设置持久化**: 重新打开设置页面（商家后台）或调用查询接口（新管理端）时，之前保存的配置正确显示
8. **两个终端一致性**: 新管理端和商家后台的API行为一致，数据格式一致

### 测试验收

1. **功能测试**: 所有功能需求测试通过
2. **集成测试**: 设置保存和读取流程测试通过
3. **浏览器测试**: 主流浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 本功能需要在两个终端（新管理端Go和商家后台PHP）都实现
- 密码验证逻辑将在其他任务中实现
- 设置数据需要与后续的验证逻辑任务对接，数据结构设计要预留扩展性
- 两个终端使用相同的数据模型和格式，确保数据一致性

### 资源约束

- 开发时间: 3-4 天
- Story Point: 3-4 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- Go: Gin 框架、GORM、现有的 Setting Service
- PHP: ThinkPHP 6.0
- Vue 3 + Element Plus
- 现有的 Setting 模型和业务设置 Controller/Service

### 服务依赖

- **Frontend → Go API**: HTTP API 调用（新管理端）
- **Frontend → PHP API**: HTTP API 调用（商家后台）

### 业务依赖

- 依赖现有的业务设置模块
  - Go: `main/app/api/v1/shop/shop_setting.go`、`main/app/service/setting/setting.go`
  - PHP: `admin/app/shop/controller/setting/Business.php`
- 依赖员工管理模块（获取员工列表）
  - Go: `main/app/api/v1/shop/shop_staff.go`
  - PHP: `admin/app/shop/model/auth/User.php`

---

## 风险和缓解

### 风险 1: 授权员工选择器需要支持多选和搜索，UI 交互需要优化

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 参考现有的员工选择组件，复用成熟的交互模式
- 使用 Element Plus 的 Select 组件（支持多选和搜索）

### 风险 2: 设置数据需要与后续的验证逻辑任务对接，数据结构设计要预留扩展性

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 在数据模型设计中考虑后续验证逻辑的需求，预留必要字段
- 使用 JSON 格式存储配置，便于扩展

### 风险 3: 需要在两个终端（新管理端Go和商家后台PHP）都实现，需要确保一致性

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 统一两个终端的实现方式，确保数据格式一致，API 接口行为一致
- 使用相同的数据模型和存储格式

---

## 时间表

- **Phase 1 - 数据库和模型**: 0.5 天（使用现有 setting 表，无需新建表）
- **Phase 2 - Go新管理端后端实现**: 0.5 天（API 和 Service）
- **Phase 3 - PHP商家后台后端实现**: 0.5 天（Controller 和 Model）
- **Phase 4 - 前端实现**: 1-1.5 天（新管理端和商家后台的 Vue 页面和组件）
- **Phase 5 - 测试和联调**: 0.5-1 天
- **总计**: 3-4 天（SP = 3-4）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 类似功能参考

- **Go新管理端**:
  - `main/app/api/v1/shop/shop_setting.go` - 业务设置 API
  - `main/app/service/setting/setting.go` - 业务设置 Service（已有业务设置处理逻辑）
- **PHP商家后台**:
  - `admin/app/shop/controller/setting/Business.php` - 业务设置 Controller（已有 `is_need_password` 类似实现）
  - `admin/app/common/enum/settings/BusinessEnum.php` - 业务设置枚举

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

