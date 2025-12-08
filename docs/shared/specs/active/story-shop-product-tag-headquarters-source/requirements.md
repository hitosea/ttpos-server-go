# 新管理端-菜品标签-来源总部数据的商品 需求文档

> 本文档定义 新管理端-菜品标签-来源总部数据的商品 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-shop-product-tag-headquarters-source.md](../../../../team/proposals/2025-12/v2.11.0-shop-product-tag-headquarters-source.md) |
| **创建日期**      | 2025-12-08                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

在分店管理端，当商品来源总部且已关联菜品标签时，分店操作商品关联本店菜品标签时，系统需要给出明确的提示，避免标签冲突和数据不一致的问题。

**核心价值**：
- 保证标签数据的一致性和准确性
- 避免分店与总部标签冲突
- 提升分店操作体验，明确提示规则
- 维护总部统一管理的商品标签体系

## 🎯 产品对齐

该功能支持总部-分店统一管理的业务模式，确保总部统一管理的商品标签不被分店覆盖，维护数据一致性和业务规则的统一性。

## 📝 用户故事

**作为** 分店管理员  
**我想** 在关联商品标签时，系统能提示我哪些商品已被总部标签关联  
**以便于** 避免标签冲突，保证数据一致性

---

## 功能需求

### Requirement 1: 创建商品标签时的冲突检测

**用户故事**: 作为分店管理员，我想在创建商品标签时，系统能检测并提示哪些商品已被总部标签关联，以便于避免标签冲突

#### 验收标准

1. **WHEN** 分店管理员创建商品标签 **AND** 关联商品中包含已被总部标签关联的商品 **THEN** 系统 **SHALL** 阻止保存并提示冲突信息
2. **WHEN** 分店管理员创建商品标签 **AND** 关联商品中不包含已被总部标签关联的商品 **THEN** 系统 **SHALL** 正常保存标签
3. **IF** 商品已被总部标签关联 **THEN** 系统 **SHALL** 提示：`商品A、B已经被来源总部的标签名称1关联，无法被当前标签关联`

#### 具体要求

- [ ] 1.1 在 `AddProductLabel` 方法中，保存标签前检查关联商品是否已被总部标签关联
- [ ] 1.2 查询关联商品的 `product_label_uuid` 字段，判断对应的标签 `headquarter_uuid > 0`
- [ ] 1.3 如果存在冲突，返回错误信息，包含冲突的商品名称和总部标签名称
- [ ] 1.4 错误信息格式：`商品{商品名称列表}已经被来源总部的标签{标签名称}关联，无法被当前标签关联`

---

### Requirement 2: 编辑商品标签时的冲突检测

**用户故事**: 作为分店管理员，我想在编辑商品标签时，系统能检测并提示新增关联商品中哪些已被总部标签关联，以便于避免标签冲突

#### 验收标准

1. **WHEN** 分店管理员编辑商品标签 **AND** 新增关联商品中包含已被总部标签关联的商品 **THEN** 系统 **SHALL** 阻止保存并提示冲突信息
2. **WHEN** 分店管理员编辑商品标签 **AND** 新增关联商品中不包含已被总部标签关联的商品 **THEN** 系统 **SHALL** 正常保存标签
3. **IF** 商品已被总部标签关联 **THEN** 系统 **SHALL** 提示：`商品A、B已经被来源总部的标签名称1关联，无法被当前标签关联`

#### 具体要求

- [ ] 2.1 在 `EditProductLabel` 方法中，更新关联商品前检查新增商品是否已被总部标签关联
- [ ] 2.2 查询新增关联商品的 `product_label_uuid` 字段，判断对应的标签 `headquarter_uuid > 0`
- [ ] 2.3 如果存在冲突，返回错误信息，包含冲突的商品名称和总部标签名称
- [ ] 2.4 错误信息格式：`商品{商品名称列表}已经被来源总部的标签{标签名称}关联，无法被当前标签关联`

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
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**涉及表结构**：
- `ttpos_product_label`: 商品标签表（已有 `headquarter_uuid` 字段）
- `ttpos_product_package`: 商品包表（已有 `product_label_uuid` 字段）

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
  - 需要确保 `ttpos_product_package.product_label_uuid` 有索引
  - 需要确保 `ttpos_product_label.headquarter_uuid` 有索引
- [ ] 缓存策略（Redis）- 本功能暂不需要缓存
- [ ] 并发处理（使用 UUID 锁）- 本功能暂不需要

### 浏览器兼容性（管理后台）

- [x] Chrome 90+
- [x] Safari 14+
- [x] Firefox 88+
- [x] Edge 90+

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [ ] **Payment/Order 相关模块测试覆盖率 100%**（高风险）- 本功能不涉及
- [ ] 集成测试覆盖核心流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

**测试场景**：
1. 创建标签时，关联商品已被总部标签关联 - 应返回错误
2. 创建标签时，关联商品未被总部标签关联 - 应成功创建
3. 编辑标签时，新增商品已被总部标签关联 - 应返回错误
4. 编辑标签时，新增商品未被总部标签关联 - 应成功更新

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 错误提示信息需要支持多语言
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
- [x] 错误日志记录（使用 Logger）
- [x] 故障恢复机制

---

## 验收标准

### 功能验收

1. **创建标签冲突检测**: 创建商品标签时，如果关联商品已被总部标签关联，系统应阻止保存并提示冲突信息
2. **编辑标签冲突检测**: 编辑商品标签时，如果新增关联商品已被总部标签关联，系统应阻止保存并提示冲突信息
3. **正常流程验证**: 创建/编辑标签时，如果关联商品未被总部标签关联，系统应正常保存

### 测试验收

1. **单元测试**: Service 层测试覆盖率 ≥ 70%
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整（如有）
3. **数据库文档**: 迁移脚本和表结构文档完整（无需新增表）
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

- 不涉及

#### PHP 模块

- 不涉及

#### Vue 模块

- 不涉及（前端仅需显示错误提示）

### 业务约束

- 总部标签优先级高于分店标签
- 分店无法覆盖总部已关联的商品标签
- 商品只能关联一个标签（总部或分店）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - 数据库 ORM
- `ttpos-server-go/app/model` - 数据模型
- `ttpos-server-go/app/repository` - 数据访问层

### 服务依赖

- **Main → BMP**: 无
- **Admin → Main**: 无
- **Frontend → Admin**: 无

### 业务依赖

- 商品标签管理功能（已存在）
- 商品包（ProductPackage）与标签的关联关系（已存在）
- 总部-分店同步机制（已存在）

---

## 风险和缓解

### 风险 1: 商品标签关联关系查询性能问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 确保 `ttpos_product_package.product_label_uuid` 字段有索引
- 确保 `ttpos_product_label.headquarter_uuid` 字段有索引
- 使用批量查询，避免 N+1 查询问题
- 如果商品数量很大，考虑分批检查

### 风险 2: 总部标签与分店标签的权限边界需要明确

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 明确权限规则：总部标签优先级高于分店标签
- 在代码中明确注释业务规则
- 错误提示信息清晰说明原因

---

## 时间表

- **Phase 1 - 需求分析**: 0.5 天
- **Phase 2 - 技术设计**: 0.5 天
- **Phase 3 - 开发实现**: 1-1.5 天
- **Phase 4 - 测试验证**: 0.5 天
- **总计**: 2-3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/vue.mdc` - Vue 开发规范
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

- DooTask #37432 - 原始需求任务

### 相关代码

- `main/app/service/product_label.go` - 商品标签服务
- `main/app/repository/product_label.go` - 商品标签数据访问层
- `main/app/model/product_label.go` - 商品标签模型
- `main/app/model/product.go` - 商品包模型（ProductPackage）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: TTPOS Team  
**审核者**: {审核者}

