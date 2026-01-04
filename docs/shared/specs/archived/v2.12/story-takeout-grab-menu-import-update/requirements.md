> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Grab 外卖导入商品优化 需求文档

> 本文档定义 Grab 外卖导入商品优化功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-grab-menu-import-update.md](../../../../team/proposals/2025-12/v2.11.0-grab-menu-import-update.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
| **负责人**        | weifashi                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | weifashi             |
| **审核日期** | 2025-12-11             |
| **审核意见** | 需求清晰，验收标准明确，技术方案可行，批准进入设计阶段         |

---

## 📋 概述

Grab 外卖导入功能是 TTPOS 与外卖平台集成的核心能力，允许商户从 Grab 平台导入商品到 TTPOS 系统。当前实现存在语言映射、价格处理、属性映射等多个问题，导致商户使用体验不佳，需要大量人工调整。

本需求旨在优化 Grab 外卖导入功能，通过完善语言映射与翻译、统一价格处理规则、优化商品属性映射（单位、税率、规格、属性组）、简化外卖开关逻辑、提供配置跳转链接等措施，提升商户接入效率，降低使用门槛，提高数据一致性。

**核心价值**：
- 减少商户手动调整工作量 60%
- 提升商户接入效率 50%
- 提高多语言数据一致性
- 简化外卖配置流程

## 🎯 产品对齐

该功能支持 TTPOS 的外卖平台集成战略，通过降低商户接入门槛、提升数据质量、优化用户体验，增强 TTPOS 在外卖场景的竞争力，为后续接入更多外卖平台（如 FoodPanda、Deliveroo 等）奠定基础。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 TTPOS 中便捷地导入 Grab 外卖商品，并自动处理语言、价格、单位、税率、规格等属性  
**以便于** 快速完成外卖平台配置，减少人工调整工作量，提升运营效率

---

## 功能需求

### Requirement 1: 多语言映射与翻译

**用户故事**: 作为商户管理员，我想在导入 Grab 商品时自动处理多语言数据，以便于确保商品在不同语言环境下显示正确。

#### 验收标准

1. **WHEN** 从 Grab 导入商品时 **THEN** 系统 **SHALL** 根据语言映射表将 Grab 语言字段映射到 TTPOS 对应的语言字段
2. **WHEN** Grab 提供的语言在 TTPOS 中不存在时 **THEN** 系统 **SHALL** 调用翻译服务自动翻译该语言
3. **WHEN** Grab 和 TTPOS 都支持某语言（如英文）时 **THEN** 系统 **SHALL** 直接使用 Grab 提供的原始值，不进行翻译
4. **WHEN** 翻译服务失败时 **THEN** 系统 **SHALL** 使用英文值作为降级方案，并记录警告日志
5. **WHEN** 将 TTPOS 商品同步到 Grab 时 **THEN** 系统 **SHALL** 使用相同的语言映射表进行反向映射

#### 具体要求

- [x] 1.1 建立 TTPOS 和 Grab 语言映射表（如：TTPOS `zh` ↔ Grab `zh-CN`）
- [x] 1.2 实现翻译服务调用接口（支持第三方翻译 API 或内部翻译服务）
- [x] 1.3 增加翻译结果缓存机制，避免重复翻译相同内容
- [x] 1.4 提供翻译失败降级策略（使用英文或原始值）
- [x] 1.5 记录语言映射和翻译日志，便于问题排查
- [x] 1.6 支持双向语言映射（TTPOS → Grab 和 Grab → TTPOS）

---

### Requirement 2: 外卖开关逻辑优化

**用户故事**: 作为商户管理员，我想在未完成 Grab 配置时也能开启外卖功能，以便于提前准备外卖分类和商品。

#### 验收标准

1. **WHEN** 商户未完成 Grab 平台配置时 **THEN** 系统 **SHALL** 仍允许开启外卖功能
2. **WHEN** 商户开启外卖功能后 **THEN** 系统 **SHALL** 允许创建和管理外卖分类
3. **WHEN** 商户开启外卖功能后 **THEN** 系统 **SHALL** 允许添加和编辑外卖商品
4. **WHEN** 商户未完成 Grab 配置时 **THEN** 系统 **SHALL** 显示提示信息，引导商户完成配置

#### 具体要求

- [x] 2.1 移除外卖开关与 Grab 配置的强关联
- [x] 2.2 开启外卖后，允许访问外卖分类管理功能
- [x] 2.3 开启外卖后，允许访问外卖商品管理功能
- [x] 2.4 在外卖管理页面显示 Grab 配置状态提示
- [x] 2.5 提供 Grab 配置引导链接

---

### Requirement 3: 配置流程优化

**用户故事**: 作为商户管理员，我想通过便捷的跳转链接访问 Grab 配置页面，以便于快速完成外卖平台配置。

#### 验收标准

1. **WHEN** 商户完成基础外卖配置后 **THEN** 系统 **SHALL** 显示 Grab 配置页面跳转链接
2. **WHEN** 商户点击跳转链接时 **THEN** 系统 **SHALL** 在新窗口打开 Grab 配置页面
3. **WHEN** 商户在 Grab 配置页面时 **THEN** 系统 **SHALL** 允许跳过商品导入步骤
4. **WHEN** 商户选择跳过导入时 **THEN** 系统 **SHALL** 仅建立平台关联，不同步商品数据
5. **WHEN** 店内存在商品时 **THEN** 系统 **SHALL NOT** 自动将店内商品同步到 Grab

#### 具体要求

- [x] 3.1 在外卖配置页面添加 Grab 配置跳转按钮
- [x] 3.2 生成带参数的 Grab 配置页面链接（包含商户信息）
- [x] 3.3 在 Grab 配置流程中增加"跳过导入"选项
- [x] 3.4 明确说明店内商品不会自动同步到 Grab
- [x] 3.5 提供配置状态显示（已配置/未配置/配置中）

---

### Requirement 4: 价格处理规则统一

**用户故事**: 作为商户管理员，我想在 TTPOS 和 Grab 之间同步商品时价格不被错误转换，以便于保持价格的准确性。

#### 验收标准

1. **WHEN** 将 TTPOS 商品价格同步到 Grab 时 **THEN** 系统 **SHALL** 直接传输价格数字，不进行汇率换算
2. **WHEN** 从 Grab 导入商品价格时 **THEN** 系统 **SHALL** 直接使用原始价格数字，不进行汇率换算
3. **WHEN** 显示 Grab 订单时 **THEN** 系统 **SHALL** 使用 Grab 的货币单位（如 $）显示价格
4. **WHEN** 显示 TTPOS 店内订单时 **THEN** 系统 **SHALL** 使用 TTPOS 的货币单位（如 ¥）显示价格
5. **WHEN** 价格精度不一致时 **THEN** 系统 **SHALL** 按目标平台的精度要求进行格式化（四舍五入）

#### 具体要求

- [x] 4.1 移除价格同步过程中的汇率换算逻辑
- [x] 4.2 统一价格传输格式（使用分为单位，整数类型）
- [x] 4.3 在订单显示时根据来源平台选择货币单位
- [x] 4.4 增加价格精度处理逻辑（小数位数统一）
- [x] 4.5 在价格传输前进行格式验证和边界检查
- [x] 4.6 记录价格转换日志，便于问题追溯

---

### Requirement 5: 商品属性映射优化

**用户故事**: 作为商户管理员，我想从 Grab 导入商品时自动创建和关联正确的单位、税率、规格和属性组，以便于减少手动配置工作量。

#### 验收标准

1. **WHEN** 从 Grab 导入商品时 **AND** 商户不存在 "Grab" 单位时 **THEN** 系统 **SHALL** 自动创建 "Grab" 单位
2. **WHEN** 从 Grab 导入商品时 **THEN** 系统 **SHALL** 默认使用 "Grab" 单位
3. **WHEN** 门店开启税率功能时 **AND** 从 Grab 导入商品时 **AND** 不存在 "0%" 税率时 **THEN** 系统 **SHALL** 自动创建税率名称为 "Grab"、税率为 "0%" 的税率配置
4. **WHEN** 从 Grab 导入商品时 **THEN** 系统 **SHALL** 默认使用 "0%" / "Grab" 税率
5. **WHEN** 从 Grab 导入商品时 **AND** 不存在 "默认" 规格时 **THEN** 系统 **SHALL** 自动创建 "默认" 规格
6. **WHEN** 从 Grab 导入商品时 **THEN** 系统 **SHALL** 使用商品价格作为 "默认" 规格的价格
7. **WHEN** 从 Grab 导入商品的属性组时 **THEN** 系统 **SHALL** 将 Grab 的 `SelectionRangeMin` 和 `SelectionRangeMax` 映射到 TTPOS 的属性组选择范围
8. **WHEN** Grab 属性组的 `SelectionRangeMin` > 0 时 **THEN** 系统 **SHALL** 设置 TTPOS 属性组为必选（`IsMust = 1`）
9. **WHEN** Grab 属性组的 `SelectionRangeMax` > 1 时 **THEN** 系统 **SHALL** 开启 TTPOS 属性组的选择数量输入（`IsOpenInput = true`）

#### 具体要求

- [x] 5.1 实现单位自动创建逻辑（检查是否存在 "Grab" 单位，不存在则创建）
- [x] 5.2 实现税率自动创建逻辑（检查是否存在 "0%" / "Grab" 税率，不存在则创建）
- [x] 5.3 实现规格自动创建逻辑（检查是否存在 "默认" 规格，不存在则创建）
- [x] 5.4 在商品导入时默认关联 "Grab" 单位、"0%" / "Grab" 税率、"默认" 规格
- [x] 5.5 实现属性组选择范围映射逻辑（`SelectionRangeMin` 和 `SelectionRangeMax` → `IsMust` 和 `IsOpenInput`）
- [x] 5.6 在属性组同步时正确设置 `MaxSelection` 字段
- [x] 5.7 增加属性组选择范围的边界值校验（如最大选择数量不能超过属性数量）
- [x] 5.8 提供默认值的多语言支持（"Grab" 单位、"Grab" 税率名称、"默认" 规格）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/takeout/grab/import_menu`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 请求参数验证完整（必填项、格式、范围）
- [x] 错误信息国际化支持
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 新增字段：`ttpos_product_unit.source`（来源平台）、`ttpos_product_unit.source_id`（来源平台ID）
- [x] 新增字段：`ttpos_tax.source`（来源平台）、`ttpos_tax.source_id`（来源平台ID）
- [x] 新增字段：`ttpos_product_flavor.source`（来源平台）、`ttpos_product_flavor.source_id`（来源平台ID）
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms（单商品导入）
- [x] 批量导入 100 个商品 < 10s
- [x] 翻译服务超时设置：5s
- [x] 数据库查询优化（使用索引）
- [x] 缓存策略（单位、税率、规格查询结果缓存 1 小时）
- [x] 并发处理（使用 UUID 锁避免重复导入）

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] 单元测试：语言映射、价格处理、属性映射逻辑
- [x] 集成测试：完整的商品导入流程（从 Grab 获取数据到 TTPOS 保存）
- [x] API 测试：导入接口的正常场景和异常场景
- [x] 翻译服务降级测试：翻译失败时的降级逻辑
- [x] 并发测试：多个商户同时导入商品

### 国际化要求

- [x] 支持 10 种语言（中文、英文、泰语、日语、韩语等）
- [x] 所有前端文案使用多语言实现
- [x] 错误提示信息支持多语言
- [x] 单位、税率、规格默认名称支持多语言
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（JWT Token）
- [x] 导入商品前验证商户权限
- [x] Grab API 调用使用 OAuth 2.0 认证
- [x] 敏感配置信息加密存储（Grab API Key）
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验和输出转义）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级（翻译服务失败时使用英文）
- [x] 事务管理（保证商品、分类、属性组的数据一致性）
- [x] 错误日志记录（使用 Logger，记录导入失败详情）
- [x] 导入失败时提供详细的错误信息和失败原因
- [x] 支持导入结果统计（成功数量、失败数量）

---

## 验收标准

### 功能验收

1. **语言映射与翻译**: 从 Grab 导入商品时，正确映射语言字段，并对缺失语言自动翻译
2. **外卖开关逻辑**: 未配置 Grab 时可开启外卖，开启后可管理外卖分类和商品
3. **配置流程**: 提供 Grab 配置跳转链接，支持跳过导入，店内商品不自动同步到 Grab
4. **价格处理**: 价格在 TTPOS 和 Grab 之间传输时不进行汇率换算，订单显示使用对应平台货币单位
5. **商品属性映射**: 自动创建和关联 "Grab" 单位、"0%" / "Grab" 税率、"默认" 规格，正确映射属性组选择范围

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%，Repository ≥ 80%）
2. **API 测试**: 导入接口的正常场景和异常场景测试通过
3. **集成测试**: 完整的商品导入流程测试通过（从 Grab 获取到 TTPOS 保存）
4. **翻译服务降级测试**: 翻译失败时使用英文降级方案
5. **并发测试**: 多个商户同时导入商品无数据冲突
6. **手动测试**: 浏览器兼容性测试通过（Chrome、Safari、Firefox、Edge）

### 文档验收

1. **技术文档**: design.md 完整且准确，包含架构设计和实现细节
2. **API 文档**: 导入接口文档完整（请求参数、响应格式、错误码）
3. **数据库文档**: 迁移脚本和表结构变更文档完整
4. **测试文档**: tasks.md 中的测试任务完成，测试用例文档完整

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以小写字母开头（如 `ITakeoutSrv` → `takeoutSrv`）
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

#### Vue 模块

- 必须使用 Vue 3 + TypeScript + Vite
- 使用 Element Plus 组件库
- 组件命名使用 PascalCase
- 遵循 `.cursor/rules/vue.mdc`

### 业务约束

- 商品导入必须在事务中执行，保证数据一致性
- 翻译服务失败不应阻塞商品导入，应使用降级方案
- 属性组选择范围必须符合业务规则（最大选择数量不能超过属性数量）
- 单位、税率、规格创建必须支持多语言
- 店内商品不会自动同步到 Grab（单向同步）

### 资源约束

- 开发时间: 5-7 天
- Story Point: 5-8 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- **翻译服务**: 第三方翻译 API（如 Google Translate API、Microsoft Translator API）或内部翻译服务
- **Grab API**: Grab 平台提供的商品获取接口（OAuth 2.0 认证）
- **语言映射表**: TTPOS 和 Grab 语言代码映射配置
- **多语言模块**: `main/pkg/language/language.go`

### 服务依赖

- **Takeout Service**: 处理外卖相关业务逻辑
- **Product Service**: 处理商品、单位、税率、规格、属性组相关业务逻辑
- **Translation Service**: 调用翻译服务接口（待实现）

### 业务依赖

- **外卖开关**: 商户必须开启外卖功能才能使用导入功能
- **Grab 配置**: 商户必须完成 Grab 平台配置才能实际同步订单
- **商品分类**: 导入商品前必须存在对应的商品分类

---

## 风险和缓解

### 风险 1: 翻译服务不稳定

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 增加翻译结果缓存，减少翻译服务调用次数
- 实现翻译失败降级方案（使用英文或原始值）
- 增加重试机制（最多重试 3 次）
- 提供手动翻译入口，允许商户手动修改翻译结果

### 风险 2: 语言映射不完全对应

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 建立详细的语言映射表，覆盖 Grab 和 TTPOS 支持的所有语言
- 对于不对应的语言，使用最接近的语言代码
- 记录语言映射日志，便于后续优化
- 提供语言映射配置界面，允许运营人员调整映射规则

### 风险 3: 价格精度不一致

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 统一使用分为单位进行价格存储和传输
- 在显示时根据货币单位进行格式化
- 增加价格精度校验，确保不超过支持的小数位数
- 记录价格转换日志，便于问题追溯

### 风险 4: 属性组映射边界情况

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 增加选择范围的边界值校验（最大选择数量不能超过属性数量）
- 对于不符合规则的选择范围，记录警告日志并使用默认值
- 提供属性组映射规则配置，允许运营人员调整映射逻辑
- 在导入前进行数据校验，提前发现问题

### 风险 5: 并发导入数据冲突

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 使用 UUID 锁避免同一商户同时导入
- 在事务中执行导入操作，保证数据一致性
- 增加幂等性检查，避免重复导入相同商品
- 记录导入日志，便于问题追溯和恢复

---

## 时间表

- **Phase 1 - 语言映射与翻译**: 1.5 天
  - 建立语言映射表
  - 实现翻译服务调用接口
  - 实现翻译结果缓存和降级逻辑
  
- **Phase 2 - 外卖开关逻辑优化**: 0.5 天
  - 移除外卖开关与 Grab 配置的强关联
  - 调整前端页面逻辑
  
- **Phase 3 - 配置流程优化**: 1 天
  - 实现 Grab 配置跳转链接生成
  - 增加"跳过导入"选项
  - 调整前端配置页面
  
- **Phase 4 - 价格处理规则**: 0.5 天
  - 移除汇率换算逻辑
  - 统一价格传输格式
  - 调整订单显示逻辑
  
- **Phase 5 - 商品属性映射**: 2.5 天
  - 实现单位、税率、规格自动创建逻辑
  - 实现属性组选择范围映射逻辑
  - 调整商品导入流程
  
- **Phase 6 - 测试与联调**: 1 天
  - 单元测试
  - 集成测试
  - API 测试
  - 前后端联调
  
- **总计**: 7 天（SP = 8）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范
- `.cursor/rules/vue.mdc` - Vue 开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关代码

- `main/app/service/takeout.go` - 外卖服务实现
- `main/app/service/product.go` - 商品服务实现
- `main/pkg/language/language.go` - 语言处理模块
- `main/app/modules/takeout/domain/menu/valueobject/` - 外卖菜单领域对象

### 外部参考

- Grab API 文档: 待补充
- Google Translate API: https://cloud.google.com/translate/docs
- Microsoft Translator API: https://docs.microsoft.com/azure/cognitive-services/translator/

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2025-12/2025-12-11.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: weifashi  
**审核者**: 待审核

