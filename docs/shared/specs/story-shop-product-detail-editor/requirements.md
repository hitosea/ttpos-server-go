# 商品管理 增加商品详情（新管理端） 需求文档

> 本文档定义商品详情功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-25-shop-product-detail-editor.md](../../../team/proposals/2025-11-25-shop-product-detail-editor.md) |
| **创建日期**      | 2025-11-25                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #36939                                                                                              |

---

## 📋 概述

为商品管理模块提供商品详情字段的后端支持。在数据模型中为单规格、多规格、套餐商品增加详情字段，提供 API 接口支持商品详情的保存和查询。前端可通过 API 获取和更新商品详情内容（富文本格式）。

## 🎯 产品对齐

- 提升商品信息完整性，帮助顾客更好地了解商品
- 增强商户的商品展示能力
- 提升用户体验和购买转化率
- 满足商户对商品详情编辑的常见需求

## 📝 用户故事

**作为** 商户管理员  
**我想** 在商品管理中为单规格、多规格、套餐商品添加详情描述  
**以便于** 向顾客展示更完整的商品信息，提升商品吸引力和购买转化率

---

## 功能需求

### Requirement 1: 数据模型扩展

**用户故事**: 作为商户管理员，我想为商品添加详情字段，以便于存储富文本格式的商品描述信息

#### 验收标准

1. **WHEN** 数据库迁移执行完成 **THEN** 系统 **SHALL** 在 `ttpos_product_package` 表中增加 `detail` 字段（LONGTEXT 类型）
2. **WHEN** 查询商品数据 **THEN** 系统 **SHALL** 返回 `detail` 字段内容（可为空）
3. **IF** `detail` 字段为空 **THEN** 系统 **SHALL** 返回空字符串，不影响现有功能

#### 具体要求

- [ ] 1.1 为 `ttpos_product_package` 表添加 `detail` 字段（LONGTEXT 类型，支持存储富文本 HTML 内容）
- [ ] 1.2 字段默认值为空字符串 `''`
- [ ] 1.3 字段注释为 `'商品详情（富文本）'`
- [ ] 1.4 更新 Go Model `ProductPackage`，添加 `Detail` 字段
- [ ] 1.5 确保字段为可选字段，不影响现有查询和更新逻辑

---

### Requirement 2: 商品详情查询接口

**用户故事**: 作为前端开发者，我想通过 API 获取商品详情，以便于在页面中展示商品详情内容

#### 验收标准

1. **WHEN** 调用商品详情查询接口（单规格商品） **THEN** 系统 **SHALL** 返回商品的 `detail` 字段内容
2. **WHEN** 调用商品详情查询接口（多规格商品） **THEN** 系统 **SHALL** 返回商品的 `detail` 字段内容
3. **WHEN** 调用商品详情查询接口（套餐商品） **THEN** 系统 **SHALL** 返回商品的 `detail` 字段内容
4. **IF** 商品 `detail` 字段为空 **THEN** 系统 **SHALL** 返回空字符串 `""`
5. **WHEN** 查询不存在的商品 **THEN** 系统 **SHALL** 返回错误信息

#### 具体要求

- [ ] 2.1 在现有商品查询接口中增加 `detail` 字段返回
- [ ] 2.2 支持通过商品 UUID 查询商品详情
- [ ] 2.3 响应格式遵循 API 设计规范（data 必须是对象）
- [ ] 2.4 空值处理：`detail` 为空时返回空字符串

---

### Requirement 3: 商品详情保存接口

**用户故事**: 作为前端开发者，我想通过 API 保存商品详情，以便于商户编辑商品详情后保存到数据库

#### 验收标准

1. **WHEN** 调用商品详情保存接口（单规格商品） **THEN** 系统 **SHALL** 正确保存 `detail` 内容到数据库
2. **WHEN** 调用商品详情保存接口（多规格商品） **THEN** 系统 **SHALL** 正确保存 `detail` 内容到数据库
3. **WHEN** 调用商品详情保存接口（套餐商品） **THEN** 系统 **SHALL** 正确保存 `detail` 内容到数据库
4. **WHEN** 保存商品详情后再次查询 **THEN** 系统 **SHALL** 返回最新保存的 `detail` 内容
5. **IF** 传入的 `detail` 为空字符串 **THEN** 系统 **SHALL** 保存为空字符串
6. **WHEN** 保存不存在的商品 **THEN** 系统 **SHALL** 返回错误信息

#### 具体要求

- [ ] 3.1 在现有商品更新接口中增加 `detail` 字段更新支持
- [ ] 3.2 支持通过商品 UUID 更新商品详情
- [ ] 3.3 请求参数验证：`detail` 字段为可选字段（string 类型）
- [ ] 3.4 响应格式遵循 API 设计规范
- [ ] 3.5 更新时自动更新 `update_time` 字段

---

### Requirement 4: 商品新增接口支持详情

**用户故事**: 作为商户管理员，我想在创建新商品时同步填写详情内容，以便商品创建完成即可对外展示完整信息

#### 验收标准

1. **WHEN** 调用商品新增接口（`/shop/product/add`） **THEN** 系统 **SHALL** 接收 `detail` 字段并写入数据库
2. **WHEN** 创建单规格、多规格、套餐商品 **THEN** 系统 **SHALL** 正确保存 `detail` 内容
3. **IF** `detail` 字段未传或为空 **THEN** 系统 **SHALL** 保存为空字符串
4. **WHEN** 新增商品后调用详情接口 **THEN** 系统 **SHALL** 返回新增的 `detail` 内容

#### 具体要求

- [ ] 4.1 在 `ProductShopAddReq` 中新增 `detail` 字段（string，可选）
- [ ] 4.2 商品新增 Service 逻辑将 `detail` 字段写入 `ttpos_product_package.detail`
- [ ] 4.3 请求参数校验：`detail` 最长不超过数据库允许的大小（由前端约束）
- [ ] 4.4 创建成功后返回与现有接口一致的响应

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

- [x] URL 使用 snake_case 命名（如：`/api/v1/shop/product_detail`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 富文本内容较大时不影响查询性能

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（富文本内容需要前端校验和过滤）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **数据模型扩展**: `ttpos_product_package` 表成功添加 `detail` 字段，Go Model 已更新
2. **查询接口**: 商品查询接口正确返回 `detail` 字段，空值处理正确
3. **保存接口**: 商品更新接口正确保存 `detail` 字段，更新后查询返回最新值
4. **兼容性**: 现有功能不受影响，向后兼容

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **兼容性测试**: 现有功能测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
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

- 商品详情字段为可选字段，不影响现有业务逻辑
- 富文本内容由前端负责渲染，后端仅存储
- 支持 HTML 格式的富文本内容（字段类型 LONGTEXT）

### 资源约束

- 开发时间: 2-3 天（仅后端）
- Story Point: 3-5 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - 数据库 ORM
- `github.com/gin-gonic/gin` - Web 框架

### 服务依赖

- **Admin → Main**: HTTP API 调用（前端通过 Admin 调用 Main API）

### 业务依赖

- 依赖现有的商品管理功能
- 依赖现有的商品查询和更新接口

---

## 风险和缓解

### 风险 1: 富文本内容较大影响性能

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用 LONGTEXT 类型存储，支持较大内容
- 查询时仅返回必要字段，避免全量查询
- 考虑内容长度限制（前端限制）

### 风险 2: 数据库迁移兼容性

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 制定完善的数据库迁移脚本
- 提供回滚方案
- 在测试环境充分验证

### 风险 3: API 向后兼容性

**影响**: 中  
**概率**: 低  
**缓解措施**:

- `detail` 字段为可选字段，不影响现有接口
- 空值返回空字符串，不返回 null
- 充分测试现有功能

---

## 时间表

- **Phase 1 - 数据库迁移**: 0.5 天
- **Phase 2 - 核心实现**: 1-1.5 天
- **Phase 3 - 测试和优化**: 0.5-1 天
- **总计**: 2-3 天（SP = 3-5）

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

### 外部参考

- [DooTask #36939](https://doo.work/task/36939) - 原始需求任务

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
**审核者**: {审核者}


