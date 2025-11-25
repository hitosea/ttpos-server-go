# Shop 端分批类型管理功能增强 需求文档

> 本文档定义 Shop 端分批类型管理功能增强的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                          |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/shop-batch-tag-enhancement.md](../../../../team/proposals/2025-11/shop-batch-tag-enhancement.md) |
| **创建日期**      | 2025-11-20                                                                                                                    |
| **负责人**        | xiezhihuan                                                                                                                    |
| **目标 Sprint**   | Sprint 待定                                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                    |
| **关联 Dootask**  | #36921 - 新管理端/点餐助手/收银端-分批送厨功能                                                                                |

---

## 📋 概述

在 Shop 端（商家后台）的分批类型管理功能中，增加名称缩写字段，以优化收银端和点餐助手的界面展示。

**注意**：分批类型的多语言支持已在 v2.9.0 版本中实现，本次无需再次实现。

## 🎯 产品对齐

本功能是 v2.10.0 版本中分批送厨功能的重要组成部分，支持：

- 多语言环境下的分批类型管理
- 收银端和点餐助手界面的优化展示
- 与系统其他多语言功能保持一致的设计模式

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 Shop 端管理分批类型时，能够设置名称缩写  
**以便于** 优化收银端和点餐助手的界面展示

---

## 功能需求

### Requirement 1: 分批类型名称缩写字段

**注意**：多语言名称支持已在 v2.9.0 版本中实现，本次无需再次实现。

**用户故事**: 作为商户管理员，我想在创建和编辑分批类型时设置名称缩写，以便于收银端和点餐助手界面展示

#### 验收标准

1. **WHEN** 商户管理员创建分批类型 **THEN** 系统 **SHALL** 要求填写名称缩写字段（必填）
2. **WHEN** 商户管理员编辑分批类型 **THEN** 系统 **SHALL** 允许修改名称缩写字段
3. **WHEN** 商户管理员查看分批类型详情 **THEN** 系统 **SHALL** 显示名称缩写信息
4. **IF** 商户管理员未填写缩写字段 **THEN** 系统 **SHALL** 提示必填并阻止提交
5. **IF** 商户管理员填写的缩写超过长度限制（10 个字符） **THEN** 系统 **SHALL** 提示错误并阻止提交

#### 具体要求

- [ ] 1.1 数据库表 `ttpos_batch_tag` 增加 `abbreviation` 字段（VARCHAR(10), NOT NULL）
- [ ] 1.2 创建分批类型接口支持缩写字段输入（必填，长度 1-10 个字符）
- [ ] 1.3 编辑分批类型接口支持缩写字段修改
- [ ] 1.4 分批类型详情接口返回缩写字段
- [ ] 1.5 缩写字段验证：必填、长度 1-10 个字符

---

### Requirement 2: 数据迁移和兼容性

**用户故事**: 作为系统管理员，我想确保现有分批类型数据能够平滑迁移，以便于不影响现有功能

#### 验收标准

1. **WHEN** 执行数据库迁移 **THEN** 系统 **SHALL** 为现有分批类型设置默认缩写（从多语言名称中提取或使用默认值）
2. **IF** 现有分批类型已有 `multi_language_name_uuid` **THEN** 系统 **SHALL** 从多语言名称中提取中文名称作为默认缩写

#### 具体要求

- [ ] 2.1 编写数据库迁移脚本，增加 `abbreviation` 字段
- [ ] 2.2 编写数据迁移脚本，为现有分批类型设置默认缩写（从多语言名称中提取中文名称，或使用名称的前几个字符）
- [ ] 2.3 确保迁移脚本可以安全回滚

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/shop/batch/tag/add`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（如需要）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] SQL 注入防护（使用参数化查询）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）

---

## 验收标准

### 功能验收

1. **缩写字段支持**: 创建、编辑、查看分批类型时，缩写字段功能正常
2. **数据迁移**: 现有数据能够平滑迁移，不影响现有功能
3. **接口兼容性**: 新接口向后兼容，不影响现有前端

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **数据迁移测试**: 迁移脚本测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 必须保证现有分批类型数据不丢失
- 必须保证接口向后兼容（但缩写字段为必填，前端需要更新）
- 缩写字段不强制唯一性，允许商户自定义
- 多语言名称功能已在 v2.9.0 实现，本次无需修改

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - 数据库 ORM
- `ttpos-server-go/app/model` - 数据模型
- `ttpos-server-go/app/dto` - 数据传输对象

### 服务依赖

- **Main → Main**: 使用现有的多语言名称服务

### 业务依赖

- 依赖现有的分批类型管理功能
- 依赖多语言名称表（`ttpos_multi_language_name`）- 已在 v2.9.0 实现

---

## 风险和缓解

### 风险 1: 数据迁移风险

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 编写详细的迁移脚本，包含回滚逻辑
- 在测试环境充分测试迁移脚本
- 为现有数据设置合理的默认值（从多语言名称中提取中文名称）

### 风险 2: 向后兼容性

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 保持现有接口结构，新增字段作为可选字段（但实际必填）
- 提供数据转换层，确保旧版本前端仍可使用
- 在 API 文档中明确标注新字段

### 风险 3: 缩写字段唯一性

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 不强制唯一性，允许商户自定义
- 在 UI 层面提示用户建议使用唯一缩写
- 在 API 文档中说明缩写字段的使用建议

---

## 时间表

- **Phase 1 - 数据库设计和迁移**: 0.5 天
- **Phase 2 - 核心实现**: 1 天
- **Phase 3 - 测试和优化**: 0.5 天
- **总计**: 2 天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关代码

- `main/app/model/product.go` - BatchTag 模型
- `main/app/repository/batch_tag.go` - BatchTag Repository
- `main/app/service/product.go` - Product Service（包含分批类型相关方法）
- `main/app/api/v1/shop/shop_batch_product.go` - 分批类型 API
- `main/app/dto/req/product.go` - 分批类型请求 DTO
- `main/app/dto/resp/product_resp/product.go` - 分批类型响应 DTO

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: xiezhihuan  
**审核者**: 待定
