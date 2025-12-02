# 套餐分组可选份数支持 需求文档

> 本文档定义套餐分组可选份数支持的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11/package-group-copy-num.md](../../../../team/proposals/2025-11/package-group-copy-num.md) |
| **创建日期**      | 2025-11-27                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [x] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

为满足套餐分组可选份数的业务需求，在 `sale_order_product` 表中新增 `copy_num` 字段，用于记录套餐子商品在分组中被选择的份数。该功能主要涉及数据库表结构变更、数据模型更新和业务逻辑适配，确保订单数据能够准确记录套餐分组的选择情况，为后续的统计、对账、退菜等业务提供数据基础。

该功能主要影响订单商品相关的数据模型和业务逻辑，涉及 main（Go）、admin（PHP）、ttpos-bmp（Go）三个模块的模型同步更新。

## 🎯 产品对齐

该功能支持公司2025年Q4的核心目标：
- **提升数据准确性**：完整记录套餐分组中每个子商品的选择份数，确保订单数据的准确性
- **支持复杂套餐场景**：满足套餐分组可选份数的业务需求，支持更灵活的套餐配置
- **优化业务处理**：为退菜、统计、对账等功能提供准确的数据基础，提升系统扩展性

## 📝 用户故事

**作为** 收银员/商户管理员  
**我想** 在订单中准确记录套餐分组中每个子商品的选择份数  
**以便于** 后续的统计、对账、退菜等业务能够正确处理套餐分组可选份数的场景

---

## 功能需求

### Requirement 1: 数据库字段新增

**用户故事**: 作为 系统管理员，我想 在 `sale_order_product` 表中新增 `copy_num` 字段，以便于 记录套餐子商品在分组中被选择的份数

#### 验收标准

1. **WHEN** 执行数据库迁移 **THEN** 系统 **SHALL** 在 `ttpos_sale_order_product` 表中成功添加 `copy_num` 字段
2. **IF** 字段类型为 `DECIMAL(12,4)` **THEN** 系统 **SHALL** 支持小数份数记录
3. **IF** 字段默认值为 `0` **THEN** 系统 **SHALL** 确保现有数据兼容
4. **WHEN** 字段位置在 `unit_num` 之后 **THEN** 系统 **SHALL** 保持表结构的一致性
5. **IF** 同步更新 `shop_01.sql` **THEN** 系统 **SHALL** 确保种子数据与最新表结构一致

#### 具体要求

- [x] 1.1 创建数据库迁移文件，添加 `copy_num` 字段
- [x] 1.2 字段类型：`DECIMAL(12,4)`，默认值：`0`，注释：表示该子商品在分组中被选择多少份
- [x] 1.3 字段位置：在 `unit_num` 字段之后
- [x] 1.4 同步更新 `admin/database/seeds/shop_01.sql` 文件
- [x] 1.5 为现有数据设置默认值 0

---

### Requirement 2: 数据模型更新

**用户故事**: 作为 开发人员，我想 更新 Go 和 PHP 的数据模型，以便于 在代码中正确使用 `copy_num` 字段

#### 验收标准

1. **WHEN** 更新 Go Model（main 模块） **THEN** 系统 **SHALL** 在 `SaleOrderProduct` 结构体中添加 `CopyNum` 字段
2. **IF** Go Model 字段类型为 `float64` **THEN** 系统 **SHALL** 正确映射到数据库的 `DECIMAL(12,4)` 类型
3. **WHEN** 更新 PHP Model（admin 模块） **THEN** 系统 **SHALL** 在对应的模型类中添加 `copy_num` 字段
4. **WHEN** 更新 BMP Model（ttpos-bmp 模块） **THEN** 系统 **SHALL** 同步更新 entity/do 结构体
5. **IF** 所有模型更新完成 **THEN** 系统 **SHALL** 确保三个模块的模型定义一致

#### 具体要求

- [x] 2.1 更新 `main/app/model/sale_order_product.go`，添加 `CopyNum float64` 字段
- [ ] 2.2 更新 admin 模块的 PHP Model，添加 `copy_num` 字段
- [ ] 2.3 更新 ttpos-bmp 模块的 entity/do 结构体（如需要）
- [ ] 2.4 确保所有模型的字段类型、默认值、注释保持一致

---

### Requirement 3: 业务逻辑适配

**用户故事**: 作为 收银员，我想 在创建订单时系统自动记录套餐子商品的份数，以便于 订单数据准确反映套餐选择情况

#### 验收标准

1. **WHEN** 创建包含套餐分组可选份数的订单 **THEN** 系统 **SHALL** 在 `sale_order_product` 表中正确记录每个子商品的 `copy_num` 字段
2. **IF** 套餐子商品属于可选分组 **THEN** 系统 **SHALL** 记录该子商品在分组中被选择的份数
3. **IF** 普通商品或套餐主商品 **THEN** 系统 **SHALL** 设置 `copy_num` 为 0 或默认值
4. **WHEN** 退菜操作涉及套餐子商品 **THEN** 系统 **SHALL** 正确处理 `copy_num` 字段
5. **WHEN** 统计报表查询订单商品 **THEN** 系统 **SHALL** 能够正确使用 `copy_num` 字段

#### 具体要求

- [x] 3.1 在订单创建逻辑中，识别套餐子商品并记录 `copy_num` 字段
- [x] 3.2 对于可选分组中的子商品，正确计算并记录份数（使用 `product.Num` 设置 `CopyNum`）
- [x] 3.3 对于普通商品和套餐主商品，设置 `copy_num` 为 0（默认值）
- [ ] 3.4 检查退菜逻辑，确保 `copy_num` 字段不影响现有退菜功能
- [ ] 3.5 检查统计逻辑，确保 `copy_num` 字段不影响现有统计功能

---

### Requirement 4: API 接口更新

**用户故事**: 作为 前端开发人员，我想 在订单查询接口中获取 `copy_num` 字段，以便于 前端展示套餐子商品的份数信息

#### 验收标准

1. **WHEN** 查询订单详情 **THEN** API **SHALL** 返回 `copy_num` 字段信息
2. **IF** 订单商品为套餐子商品 **THEN** API **SHALL** 返回该商品在分组中的份数
3. **IF** 订单商品为普通商品或套餐主商品 **THEN** API **SHALL** 返回 `copy_num` 为 0
4. **WHEN** 订单列表查询 **THEN** API **SHALL** 在商品信息中包含 `copy_num` 字段

#### 具体要求

- [ ] 4.1 更新订单详情接口，在响应中包含 `copy_num` 字段
- [ ] 4.2 更新订单列表接口，在商品信息中包含 `copy_num` 字段
- [ ] 4.3 确保 DTO 结构体包含 `copy_num` 字段
- [ ] 4.4 更新 API 文档（如有），说明 `copy_num` 字段的含义

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 数量字段使用 `DECIMAL(12,4)` 类型（与 `unit_num` 保持一致）
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 敏感数据加密存储
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（前端输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **数据库字段新增**: 迁移脚本执行成功，`copy_num` 字段已添加到表中
2. **数据模型更新**: Go、PHP、BMP 三个模块的模型已同步更新
3. **业务逻辑适配**: 订单创建时正确记录 `copy_num` 字段
4. **API 接口更新**: 订单查询接口返回 `copy_num` 字段信息
5. **数据兼容性**: 现有订单数据的 `copy_num` 字段默认值为 0，不影响现有功能

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（订单创建、查询、退菜）
4. **手动测试**: 套餐分组可选份数场景测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

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

- **向后兼容**: 现有订单数据的 `copy_num` 字段默认值为 0，不影响现有功能
- **数据完整性**: 所有新创建的订单商品必须设置 `copy_num` 字段，不能为 NULL
- **字段用途**: `copy_num` 字段主要用于套餐子商品，普通商品和套餐主商品该字段值为 0

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP（待技术评审确认，必须 ≤ 5）

---

## 依赖关系

### 技术依赖

- `main/app/model/sale_order_product.go` - 订单商品模型
- `main/app/model/product_package_group.go` - 套餐分组模型
- `admin/database/seeds/shop_01.sql` - 数据库表结构定义
- `admin/database/migrations/` - 数据库迁移脚本目录

### 服务依赖

- **Main → BMP**: 无依赖（如需要同步 BMP 模块模型）
- **Admin → Main**: 无依赖
- **Frontend → Admin**: 无依赖

### 业务依赖

- **套餐分组功能**: 依赖现有的套餐分组功能正常工作
- **订单创建流程**: 依赖现有的订单创建流程正常工作

---

## 风险和缓解

### 风险 1: 数据兼容性

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 新字段设置默认值 0，确保现有数据兼容
- 数据迁移脚本为现有数据设置默认值
- 充分测试现有订单查询、统计等功能

### 风险 2: 业务逻辑影响

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 全面检查所有使用 `sale_order_product` 表的地方
- 确保新字段不影响现有逻辑
- 代码审查时重点检查订单创建、退菜、统计等核心流程

### 风险 3: 多模块同步

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 严格按照数据库开发规范执行
- 确保 main、admin、ttpos-bmp 三个模块的模型定义一致
- 代码审查时检查所有模块的模型更新

### 风险 4: 遗漏业务逻辑适配

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 全面搜索代码库中所有创建 `SaleOrderProduct` 的地方
- 代码审查时重点检查订单创建逻辑
- 添加单元测试覆盖所有创建路径

---

## 时间表

- **Phase 1 - 数据库和模型更新**: 1.5 天
  - 数据库迁移脚本编写（0.5 天）
  - Go Model 更新（main 模块）（0.5 天）
  - PHP Model 更新（admin 模块）（0.5 天）
  - BMP Model 更新（ttpos-bmp 模块）（0.5 天）
- **Phase 2 - 业务逻辑适配**: 1 天
  - 订单创建业务逻辑适配（1 天）
- **Phase 3 - API 接口更新和测试**: 0.5 天
  - API 接口更新和测试（0.5 天）
- **总计**: 2-3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/php-architecture.md` - PHP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- 套餐分组类型和加价功能：`docs/shared/api/frontend-changes-package-group-type.md`
- 套餐管理增强功能：`docs/shared/api/shop-package-management-enhancement.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: xiezhihuan  
**审核者**: {审核者}

