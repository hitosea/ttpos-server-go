# 旧管理端-商品管理-套餐 需求文档

> 本文档定义旧管理端商品管理中套餐功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                  |
| ----------------- | --------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/旧管理端-商品管理-套餐.md](../../../../team/proposals/2025-11/旧管理端-商品管理-套餐.md) |
| **创建日期**      | 2025-11-25                                                                                                            |
| **负责人**        | 待分配                                                                                                                |
| **目标 Sprint**   | Sprint {N}                                                                                                            |
| **涉及技术栈**    | [ ] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                            |
| **关联任务**      | DooTask #36907                                                                                                        |

---

## 📋 概述

在旧管理端的商品管理模块中，增强套餐商品管理功能的 PHP 后端 API，支持更灵活的分组和选择机制。**注意：数据库字段已添加，本次仅实现 PHP 后端 API 功能。** 通过后端 API 支持分组类型（固定/可选）、可选数量控制、必选/默认选项配置、价格加价模式等功能。

## 🎯 产品对齐

该功能支持产品在套餐商品配置上的灵活性需求，满足不同业务场景的套餐组合规则，提升商户使用体验，降低配置复杂度。

## 📝 用户故事

**作为** 商户管理员  
**我想** 通过后端 API 配置套餐的分组类型、可选数量、必选/默认选项，以及商品的加价和数量  
**以便于** 创建更灵活、更符合业务需求的套餐商品，提升套餐配置的灵活性和用户体验

---

## 功能需求

### Requirement 1: 分组类型配置（PHP 后端 API 支持）

**用户故事**: 作为商户管理员，我想通过后端 API 为套餐商品配置分组类型（固定/可选），以便于控制套餐商品的选择规则

#### 验收标准

1. **WHEN** 调用保存套餐组 API **THEN** 系统 **SHALL** 支持 `group_type` 字段（0-固定，1-可选）
2. **WHEN** `group_type` 为 1（可选）时 **THEN** 系统 **SHALL** 支持 `optional_count` 字段
3. **WHEN** `optional_count` 未提供 **THEN** 系统 **SHALL** 默认值为 1

#### 具体要求

- [x] 1.1 数据库字段已存在：`ttpos_product_package_group.group_type` 字段（0-固定，1-可选）
- [x] 1.2 数据库字段已存在：`ttpos_product_package_group.optional_count` 字段
- [ ] 1.3 PHP 后端 API 支持接收和处理 `group_type` 和 `optional_count` 字段
- [ ] 1.4 更新 PHP Model：`admin/app/common/model/product/ProductPackageGroup.php`

---

### Requirement 2: 可选分组配置（PHP 后端 API 支持）

**用户故事**: 作为商户管理员，我想通过后端 API 为可选分组配置可选数量、必选和默认选项，以便于控制用户的选择行为

#### 验收标准

1. **WHEN** 调用保存套餐组 API **THEN** 系统 **SHALL** 支持 `optional_count` 字段（默认 1）
2. **WHEN** 调用保存套餐组商品 API **THEN** 系统 **SHALL** 支持 `is_required` 和 `is_default` 字段
3. **WHEN** 必选+默认商品总数大于可选数量 **THEN** 系统 **SHALL** 返回错误："必选+默认商品总数不可大于可选数量"

#### 具体要求

- [x] 2.1 数据库字段已存在：`ttpos_product_package_group_item.is_required` 字段（0-否 1-是，默认 0）
- [x] 2.2 数据库字段已存在：`ttpos_product_package_group_item.is_default` 字段（0-否 1-是，默认 0）
- [ ] 2.3 PHP 后端 API 实现数据校验：必选+默认商品总数不可大于可选数量（如可选2，A商品必选+默认，B商品默认，C商品必选）
- [ ] 2.4 更新 PHP Model：`admin/app/common/model/product/ProductPackageGroupItem.php`（支持新字段）
- [ ] 2.5 更新 `ProductPackageGroup::updatePackageGroup` 方法，支持新字段

---

### Requirement 3: 价格模式调整（PHP 后端 API 支持）

**用户故事**: 作为商户管理员，我想通过后端 API 使用加价模式设置套餐商品价格，以便于更灵活地设置套餐商品的价格

#### 验收标准

1. **WHEN** 调用保存套餐组商品 API **THEN** 系统 **SHALL** 支持 `add_price` 字段（加价金额）
2. **WHEN** `add_price` 未提供 **THEN** 系统 **SHALL** 默认值为 0

#### 具体要求

- [x] 3.1 数据库字段已存在：`ttpos_product_package_group_item.add_price` 字段（默认值为 0）
- [ ] 3.2 PHP 后端 API 支持接收和处理 `add_price` 字段

---

### Requirement 4: 数量默认值（PHP 后端 API 支持）

**用户故事**: 作为商户管理员，我想通过后端 API 设置商品数量，系统默认值为 1，以便于简化配置流程

#### 验收标准

1. **WHEN** 调用保存套餐组商品 API **AND** `num` 未提供 **THEN** 系统 **SHALL** 默认值为 1

#### 具体要求

- [x] 4.1 数据库字段已存在：`ttpos_product_package_group_item.num` 字段（默认值为 1）
- [ ] 4.2 PHP 后端 API 支持接收和处理 `num` 字段，未提供时使用默认值 1

---

### Requirement 5: 数据校验逻辑

**用户故事**: 作为商户管理员，我想后端 API 能够正确校验套餐配置数据，以便于确保数据正确性

#### 验收标准

1. **WHEN** 调用保存套餐组 API **AND** 必选+默认商品总数大于可选数量 **THEN** 系统 **SHALL** 返回错误信息
2. **WHEN** 调用保存套餐组 API **AND** 数据格式不正确 **THEN** 系统 **SHALL** 返回参数错误信息

#### 具体要求

- [ ] 5.1 实现数据校验：必选+默认商品总数不可大于可选数量
- [ ] 5.2 实现参数验证：确保必填字段存在，数据类型正确

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Model 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Model 应独立且可复用
- **遵循规范**:
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] 响应格式统一
- [ ] 错误信息清晰明确
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 使用软删除（delete_time）
- [ ] 时间字段使用 int 类型
- [ ] 金额字段使用 decimal
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 事务管理（保证数据一致性）

### 测试要求

- [ ] 单元测试覆盖核心逻辑
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程
- [ ] 参考: `.cursor/rules/php.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（输入校验）
- [ ] CSRF 防护（Token 验证）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **分组类型配置**: PHP 后端 API 支持"固定"和"可选"两种分组类型
2. **可选数量控制**: PHP 后端 API 支持可选数量字段，默认 1
3. **必选/默认选项**: PHP 后端 API 支持商品的必选和默认选项
4. **价格加价模式**: PHP 后端 API 支持加价字段，默认 0
5. **数量默认值**: PHP 后端 API 支持数量字段，默认 1
6. **数据校验**: PHP 后端 API 校验必选+默认商品总数不可大于可选数量

### 测试验收

1. **单元测试**: 核心逻辑测试通过
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
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

### 业务约束

- 数据库字段已添加，无需数据迁移
- 业务逻辑：可选/必选逻辑可能影响订单计算，需要充分测试
- 仅实现 PHP 后端 API，前端实现不在本次范围

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `admin/app/common/model/product/ProductPackageGroup.php` - 套餐组模型
- `admin/app/common/model/product/ProductPackageGroupItem.php` - 套餐组商品模型
- `admin/app/shop/service/ProductService.php` - 商品服务
- `admin/app/shop/controller/product/store/Product.php` - 商品控制器

### 服务依赖

- **Frontend → Admin**: HTTP API 调用（本次仅实现 Admin 端 API）

### 业务依赖

- 依赖现有的套餐商品管理功能
- 依赖商品管理模块的基础功能

---

## 风险和缓解

### 风险 1: 业务逻辑影响

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 充分测试订单计算逻辑，确保价格计算正确
- 与订单模块开发人员沟通，确认影响范围
- 编写完整的集成测试用例

### 风险 2: 数据兼容性

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 确保新字段有默认值，不影响现有数据
- 在测试环境充分验证

---

## 时间表

- **Phase 1 - Model 更新**: 0.5 天
- **Phase 2 - PHP 后端核心实现**: 1.5 天
- **Phase 3 - 测试和优化**: 1 天
- **总计**: 3 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/php-architecture.md` - PHP 架构

### 开发指南

- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- DooTask #36907 - 产品需求文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**作者**: 产品组 + 开发组  
**审核者**: 待审核
