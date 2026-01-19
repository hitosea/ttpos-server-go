> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 单品备注原因管理 需求文档

> 本文档定义单品备注原因管理的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/item-remark-reason-management.md](../../../../team/proposals/2025-12/item-remark-reason-management.md) |
| **创建日期**      | 2025-12-05                                                                                                                   |
| **负责人**        | {姓名}                                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-05               |
| **审核意见** | 需求明确，可以开始开发   |

---

## 📋 概述

参考"整单备注"的实现逻辑，为"单品备注"添加原因管理功能。在旧商户后台（PHP）和新管理端（Go）的业务设置中，新增"单品备注"原因管理模块，支持多语言、增删改查等操作，与"整单备注"逻辑保持一致。

**本次仅实现后端 API，前端界面后续实现。**

## 🎯 产品对齐

该功能支持订单备注管理需求：

- **提升收银效率**：收银员可通过选择预设原因快速添加单品备注
- **统一备注格式**：通过预设原因管理，确保备注信息格式统一
- **改善后厨体验**：统一的备注格式便于后厨理解订单需求
- **数据统计支持**：为后续分析常用备注原因提供数据基础

## 📝 用户故事

**作为** 商户管理员  
**我想** 在业务设置中管理单品备注原因列表  
**以便于** 收银员可以快速选择预设原因添加单品备注，提升效率和统一性

**作为** 收银员  
**我想** 在添加单品备注时选择预设原因  
**以便于** 快速完成备注操作，避免手动输入错误

---

## 功能需求

### Requirement 1: 单品备注原因管理（新管理端 Go）

**用户故事**: 作为商户管理员，我想在新管理端管理单品备注原因列表，以便于统一管理常用的单品备注原因

#### 验收标准

1. **WHEN** 调用获取单品备注原因列表 API (`GET /shop/setting/order_item_remark`) **THEN** 系统 **SHALL** 返回当前商户的所有有效原因列表（支持多语言）
2. **WHEN** 调用新增单品备注原因 API (`POST /shop/setting/order_item_remark/add`) **THEN** 系统 **SHALL** 创建新的单品备注原因记录
3. **WHEN** 调用编辑单品备注原因 API (`POST /shop/setting/order_item_remark/edit`) **THEN** 系统 **SHALL** 更新指定原因的多语言名称信息
4. **WHEN** 调用删除单品备注原因 API (`DELETE /shop/setting/order_item_remark`) **THEN** 系统 **SHALL** 软删除记录，不影响历史订单数据
5. **IF** 单品备注原因数量已达到 100 个 **THEN** 系统 **SHALL** 拒绝新增操作并提示错误信息"单品备注数量不能超过100个"
6. **WHEN** 新增或编辑单品备注原因 **THEN** 系统 **SHALL** 验证多语言名称的完整性（根据门店语言设置）
7. **WHEN** 新增或编辑单品备注原因 **THEN** 系统 **SHALL** 验证每个语言名称的字数限制（非字符）不超过 100 字
8. **IF** 多语言名称不完整 **THEN** 系统 **SHALL** 提示"多语言名称不完整"
9. **IF** 名称长度超过 100 字（非字符） **THEN** 系统 **SHALL** 提示"名称长度不能超过100个字"
10. **WHEN** 调用 API **THEN** 系统 **SHALL** 验证用户权限，权限处理逻辑与"整单备注"一致

#### 具体要求

- [ ] 1.1 创建 `OrderItemRemark` 模型（`main/app/model/order_item_remark.go`）
- [ ] 1.2 创建 `OrderItemRemarkRepo` 仓库（`main/app/repository/base/order_item_remark.go`），实现增删改查接口
- [ ] 1.3 在 Service（`main/app/service/other.go`）中实现 `AddOrderItemRemark`、`EditOrderItemRemark`、`DeleteOrderItemRemark`、`GetOrderItemRemarkList` 方法
- [ ] 1.4 在 API（`main/app/api/v1/shop/shop_setting.go`）中添加 4 个接口：`GetOrderItemRemark`、`AddOrderItemRemark`、`EditOrderItemRemark`、`DeleteOrderItemRemark`
- [ ] 1.5 实现数量限制验证（最多 100 个）
- [ ] 1.6 实现多语言名称验证（完整性、字数限制）
- [ ] 1.7 实现权限验证逻辑（与整单备注一致）

---

### Requirement 2: 单品备注原因管理（旧商户后台 PHP）

**用户故事**: 作为商户管理员，我想在旧商户后台管理单品备注原因列表，以便于统一管理常用的单品备注原因

#### 验收标准

1. **WHEN** 调用 `GET /index.php/shop/setting.Business/orderItemRemark` **THEN** 系统 **SHALL** 返回当前商户的所有有效原因列表（支持多语言）
2. **WHEN** 调用 `POST /index.php/shop/setting.Business/orderItemRemark` 并传入 `order_item_remark` 数组 **THEN** 系统 **SHALL** 支持批量增删改操作
3. **WHEN** 新增单品备注原因（action='add'） **THEN** 系统 **SHALL** 创建新的单品备注原因记录
4. **WHEN** 编辑单品备注原因（action='edit'） **THEN** 系统 **SHALL** 更新指定原因的多语言名称信息
5. **WHEN** 删除单品备注原因（action='delete'） **THEN** 系统 **SHALL** 软删除记录，不影响历史订单数据
6. **IF** 单品备注原因数量已达到 100 个 **THEN** 系统 **SHALL** 拒绝新增操作并提示错误信息"单品备注数量不能超过100个"
7. **WHEN** 新增或编辑单品备注原因 **THEN** 系统 **SHALL** 验证多语言名称的完整性（根据门店语言设置）
8. **WHEN** 新增或编辑单品备注原因 **THEN** 系统 **SHALL** 验证每个语言名称的字数限制（非字符）不超过 100 字
9. **IF** 多语言名称不完整 **THEN** 系统 **SHALL** 提示"多语言名称不完整"
10. **IF** 名称长度超过 100 字（非字符） **THEN** 系统 **SHALL** 提示"名称长度不能超过100个字"
11. **WHEN** 调用接口 **THEN** 系统 **SHALL** 验证用户权限，权限处理逻辑与"整单备注"一致

#### 具体要求

- [ ] 2.1 创建 `OrderItemRemark` 模型（`admin/app/shop/model/setting/OrderItemRemark.php`）
- [ ] 2.2 在 Controller（`admin/app/shop/controller/setting/Business.php`）中新增 `orderItemRemark()` 方法
- [ ] 2.3 实现批量增删改操作逻辑（参考 `orderRemark()` 方法）
- [ ] 2.4 实现数量限制验证（最多 100 个）
- [ ] 2.5 实现多语言名称验证（完整性、字数限制）
- [ ] 2.6 实现权限验证逻辑（与整单备注一致）

---

### Requirement 3: 数据模型设计

**用户故事**: 作为系统，我想存储单品备注原因数据，以便于管理和查询

#### 验收标准

1. **WHEN** 创建数据表 `order_item_remark` **THEN** 系统 **SHALL** 包含以下字段：id（主键）, uuid, name, multi_language_name_uuid, create_time, update_time, delete_time
2. **WHEN** 创建数据表 **THEN** 系统 **SHALL** 设置 id 为主键，uuid 为唯一索引
3. **WHEN** 创建数据表 **THEN** 系统 **SHALL** 使用软删除机制（delete_time 字段）

#### 具体要求

- [ ] 3.1 创建数据库迁移文件（`admin/database/migrations/YYYYMMDDHHMMSS_create_order_item_remark_table.php`）
- [ ] 3.2 表结构参考 `order_remark` 表，但添加 id 主键，移除 app_id 和 shop_supplier_id 字段
- [ ] 3.3 字段类型和约束：id 为自增主键，uuid 为 bigint unsigned 唯一索引，其他字段类型与 `order_remark` 表保持一致

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/shop/setting/order_item_remark`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀（`ttpos_order_item_remark`）
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 并发处理（使用 UUID 锁，如需要）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语、泰语、繁体中文、缅甸语、土耳其语、瑞典语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] SQL 注入防护（使用参数化查询）
- [x] 权限验证（与整单备注权限逻辑一致）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **新管理端 API 功能**: 所有 4 个 API 接口正常工作，支持增删改查操作
2. **旧商户后台 API 功能**: `orderItemRemark()` 方法正常工作，支持批量增删改操作
3. **数量限制**: 最多添加 100 个单品备注原因，超过限制时提示错误
4. **字数限制**: 每个语言名称字数（非字符）不超过 100 字，超过限制时提示错误
5. **多语言支持**: 支持所有门店配置的语言，验证多语言名称完整性
6. **权限验证**: 权限处理逻辑与"整单备注"一致
7. **软删除**: 删除操作使用软删除，不影响历史订单数据

### 测试验收

1. **单元测试**: Service 和 Repository 层测试覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **边界测试**: 数量限制、字数限制等边界条件测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 参考 `OrderRemark` 相关实现

#### PHP 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除
- 参考 `orderRemark()` 方法实现

### 业务约束

- 最多添加 100 个单品备注原因
- 每个语言名称字数（非字符）限制为 100 字
- 权限处理逻辑与"整单备注"一致
- 仅实现后端 API，前端界面后续实现

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `OrderRemark` 相关实现 - 参考实现逻辑
- `MultiLanguageName` 模型 - 多语言名称支持

### 服务依赖

- **Main → Main**: 复用现有的多语言验证逻辑
- **Admin → Admin**: 复用现有的多语言验证逻辑

### 业务依赖

- 依赖"整单备注"功能的实现逻辑
- 依赖门店语言配置

---

## 风险和缓解

### 风险 1: API 设计与整单备注不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 严格参考 `OrderRemark` 相关 API 的实现逻辑
- 代码审查时重点检查一致性

### 风险 2: 多语言验证逻辑错误

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 复用现有的多语言验证逻辑
- 单元测试覆盖多语言验证场景

### 风险 3: 权限验证逻辑不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 明确权限处理逻辑与"整单备注"一致
- 代码审查时重点检查权限验证

---

## 时间表

- **Phase 1 - 数据模型设计**: 0.5 天
- **Phase 2 - Go Main 模块 API 实现**: 1 天
- **Phase 3 - PHP Admin 模块 API 实现**: 0.5 天
- **Phase 4 - 单元测试和联调**: 0.5-1 天
- **总计**: 2-3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- **整单备注功能**:
  - Go API: `main/app/api/v1/shop/shop_setting.go` (GetOrderRemark, AddOrderRemark, EditOrderRemark, DeleteOrderRemark)
  - Go Service: `main/app/service/other.go` (AddOrderRemark, EditOrderRemark, DeleteOrderRemark)
  - Go Repository: `main/app/repository/base/order_remark.go`
  - PHP API: `admin/app/shop/controller/setting/Business.php` (orderRemark 方法)
  - 数据库迁移: `admin/database/migrations/20251020134645_create_order_remark_table.php`

### 外部参考

- [无]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: 王昱  
**审核者**: {审核者}

