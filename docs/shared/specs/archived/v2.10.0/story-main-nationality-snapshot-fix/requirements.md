> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 国籍信息快照修复 需求文档

> 本文档定义国籍信息快照修复功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/nationality-snapshot-fix.md](../../../../team/proposals/2025-12/nationality-snapshot-fix.md) |
| **创建日期**      | 2025-12-02                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | 产品组             |
| **审核日期** | 2025-12-02             |
| **审核意见** | 需求明确，设计合理，可进入技术设计阶段         |

---

## 📋 概述

当前订单查询时，国籍信息直接从 `ttpos_nationality` 关联表获取，导致后台删除或修改国籍配置时，历史订单显示会变化。本功能将为订单表添加国籍名称快照字段，确保订单历史信息准确反映下单时的真实状态，满足财务、税务对订单历史记录的合规性要求。

**核心价值**：
- 确保订单国籍信息准确反映下单时的状态
- 满足财务、税务对订单历史记录的合规性要求
- 支持订单历史查询和问题追溯
- 避免因数据变更导致的业务逻辑错误

## 🎯 产品对齐

本功能是"订单商品信息快照修复"（`order-attribute-snapshot-fix.md`）系列需求的一部分，通过建立完整的订单快照机制，确保：
- **数据一致性**：订单信息作为历史快照，不随后台配置变更而改变
- **业务可靠性**：支持订单对账、报表、审计等关键业务场景
- **合规性**：满足餐饮行业对历史订单数据的监管要求

## 📝 用户故事

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实国籍信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的国籍信息  
**以便于** 准确处理客户咨询

---

## 功能需求

### Requirement 1: 数据库结构变更 - 添加国籍名称快照字段

**用户故事**: 作为开发者，我想在订单表中添加国籍名称快照字段，以便于保存下单时的国籍信息。

#### 验收标准

1. **WHEN** 执行数据库迁移脚本 **THEN** 系统 **SHALL** 在 `ttpos_sale_bill` 表添加 `nationality_name` 字段
2. **WHEN** 添加字段后 **THEN** 字段类型 **SHALL** 为 `VARCHAR(255)`，默认值为空字符串
3. **WHEN** 添加字段后 **THEN** 字段注释 **SHALL** 为 "国籍名称快照（单语言），不随后台更新"
4. **WHEN** 执行迁移 **THEN** 系统 **SHALL** 不影响现有数据和业务功能

#### 具体要求

- [x] 1.1 在 `ttpos_sale_bill` 表添加 `nationality_name` 字段（VARCHAR(255)）
- [x] 1.2 字段位置：`AFTER nationality_uuid`
- [x] 1.3 字段默认值：空字符串 `''`
- [x] 1.4 字段注释：明确说明为快照字段，不随后台更新
- [x] 1.5 迁移脚本支持可重复执行（幂等性）

---

### Requirement 2: 数据模型修改 - 添加 NationalityName 字段

**用户故事**: 作为开发者，我想在 `SaleBill` 模型中添加 `NationalityName` 字段，以便于在代码中操作快照数据。

#### 验收标准

1. **WHEN** 修改 `SaleBill` 结构体 **THEN** 系统 **SHALL** 添加 `NationalityName` 字段
2. **WHEN** 字段定义完成 **THEN** GORM 标签 **SHALL** 正确映射到数据库字段 `nationality_name`
3. **WHEN** 字段定义完成 **THEN** JSON 标签 **SHALL** 为 `nationality_name`
4. **WHEN** 编译代码 **THEN** 系统 **SHALL** 无编译错误

#### 具体要求

- [x] 2.1 在 `SaleBill` 结构体添加 `NationalityName string` 字段
- [x] 2.2 添加 GORM 标签：`gorm:"column:nationality_name"`
- [x] 2.3 添加 JSON 标签：`json:"nationality_name"`
- [x] 2.4 字段位置：紧跟 `NationalityUuid` 字段之后

---

### Requirement 3: 查询逻辑修改 - 实现国籍名称获取方法

**用户故事**: 作为开发者，我想实现国籍名称获取方法，以便于优先使用快照数据，降级使用关联表数据。

#### 验收标准

1. **WHEN** 调用 `GetLocaleNationalityName()` 方法 **THEN** 系统 **SHALL** 优先返回快照字段 `NationalityName`
2. **IF** 快照字段为空 **THEN** 系统 **SHALL** 降级使用 `Nationality.MultiLanguageName` 关联表数据
3. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）
4. **IF** 快照字段有值且关联表存在 **THEN** 主语言（ZH）**SHALL** 使用快照，其他语言 **SHALL** 使用关联表
5. **IF** 快照字段有值但关联表不存在 **THEN** 所有语言 **SHALL** 使用快照的主语言填充

#### 具体要求

- [x] 3.1 新增 `SaleBill.GetLocaleNationalityName()` 方法
- [x] 3.2 方法返回类型：`dto.LocaleResponse`
- [x] 3.3 实现快照优先逻辑：优先使用 `NationalityName` 字段
- [x] 3.4 实现降级逻辑：快照为空时使用关联表数据
- [x] 3.5 实现多语言支持：主语言用快照，其他语言从关联表补充
- [x] 3.6 实现兜底逻辑：关联表不存在时，所有语言用快照填充
- [x] 3.7 添加方法注释，说明逻辑和用途

---

### Requirement 4: 下单逻辑修改 - 保存国籍名称快照

**用户故事**: 作为开发者，我想在创建订单时保存国籍名称快照，以便于新订单自动使用快照机制。

#### 验收标准

1. **WHEN** 创建订单（`SaleBill`）**THEN** 系统 **SHALL** 保存国籍名称快照到 `NationalityName` 字段
2. **WHEN** 保存快照 **THEN** 系统 **SHALL** 从 `Nationality.MultiLanguageName.ZhName` 获取中文名称
3. **IF** 国籍信息不存在 **THEN** 快照字段 **SHALL** 为空字符串
4. **WHEN** 保存订单 **THEN** 系统 **SHALL** 不影响现有下单流程

#### 具体要求

- [x] 4.1 梳理所有下单入口（POS、扫码点餐、外卖等）
- [x] 4.2 在创建 `SaleBill` 时，从 `Nationality.MultiLanguageName.ZhName` 获取中文名称
- [x] 4.3 将中文名称保存到 `SaleBill.NationalityName` 字段
- [x] 4.4 确保所有下单入口都正确保存快照
- [x] 4.5 处理边界情况：国籍信息不存在或为空

---

### Requirement 5: 订单查询逻辑修改 - 使用快照数据

**用户故事**: 作为开发者，我想在订单查询时使用快照数据，以便于显示下单时的真实国籍信息。

#### 验收标准

1. **WHEN** 查询订单详情 **THEN** 系统 **SHALL** 使用 `GetLocaleNationalityName()` 方法获取国籍名称
2. **WHEN** 返回订单详情 **THEN** 国籍名称 **SHALL** 为多语言格式（`LocaleResponse`）
3. **IF** 后台删除了国籍配置 **THEN** 历史订单 **SHALL** 仍然显示下单时的国籍名称
4. **IF** 后台修改了国籍名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称

#### 具体要求

- [x] 5.1 梳理所有订单查询接口（订单详情、订单列表、报表等）
- [x] 5.2 替换原有的直接从 `Nationality.MultiLanguageName` 获取的逻辑
- [x] 5.3 使用 `SaleBill.GetLocaleNationalityName()` 方法获取国籍名称
- [x] 5.4 确保所有查询接口都使用快照数据
- [x] 5.5 验证历史订单兼容性（快照字段为空的情况）

---

### Requirement 6: 数据迁移和兼容性处理

**用户故事**: 作为开发者，我想为历史订单补充快照数据，以便于历史订单也能正常显示国籍信息。

#### 验收标准

1. **WHEN** 检查历史订单 **THEN** 系统 **SHALL** 统计 `nationality_name` 字段的填充情况
2. **WHEN** 执行数据迁移 **THEN** 系统 **SHALL** 仅迁移关联表数据存在的记录
3. **IF** 关联表数据已删除 **THEN** 快照字段 **SHALL** 保持为空（通过降级逻辑兼容）
4. **WHEN** 迁移完成 **THEN** 系统 **SHALL** 验证迁移结果的正确性

#### 具体要求

- [x] 6.1 编写数据检查脚本，统计需要迁移的订单数量
- [x] 6.2 编写数据迁移脚本，从关联表补充快照字段
- [x] 6.3 迁移脚本只处理关联表数据存在的记录
- [x] 6.4 迁移脚本支持可重复执行（幂等性）
- [x] 6.5 验证迁移结果，确保数据正确性

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/database.mdc` - 数据库开发规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 快照字段使用 VARCHAR(255) 类型
- [x] 快照字段默认值为空字符串
- [x] 快照字段注释明确说明"不随后台更新"
- [x] 迁移脚本支持可重复执行（幂等性）
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 优先使用快照数据，减少关联查询
- [x] 降级查询时使用索引优化（`nationality_uuid` 已有索引）
- [x] 本地响应时间 < 200ms

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] **Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程（下单、查询、多语言）
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、泰语、日语、韩语等）
- [x] 快照字段保存主语言（中文），其他语言从关联表补充
- [x] 关联表不存在时，所有语言使用快照的主语言填充
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] SQL 注入防护（使用参数化查询）
- [x] 数据迁移脚本执行前需要备份
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 兼容性处理：快照字段为空时降级使用关联表数据
- [x] 错误日志记录（使用 Logger）
- [x] 数据迁移失败时回滚机制

---

## 验收标准

### 功能验收

1. **数据库变更**: `ttpos_sale_bill` 表成功添加 `nationality_name` 字段，字段类型和注释符合要求
2. **模型修改**: `SaleBill` 结构体添加 `NationalityName` 字段，GORM 和 JSON 标签正确
3. **查询逻辑**: `GetLocaleNationalityName()` 方法正确实现快照优先、降级兼容、多语言支持
4. **下单逻辑**: 所有下单入口都正确保存国籍名称快照到 `NationalityName` 字段
5. **订单查询**: 所有订单查询接口都使用快照数据，历史订单正常显示
6. **国籍删除测试**: 删除国籍配置后，历史订单仍然显示下单时的国籍名称
7. **国籍改名测试**: 修改国籍名称后，历史订单仍然显示修改前的原始名称
8. **多语言测试**: 关联表存在时其他语言正常显示，关联表不存在时所有语言使用快照填充
9. **历史订单兼容**: 快照字段为空的历史订单，通过降级逻辑正常显示
10. **数据迁移**: 数据迁移脚本正确执行，仅迁移关联表数据存在的记录

### 测试验收

1. **单元测试**: 
   - `GetLocaleNationalityName()` 方法测试覆盖率 100%
   - 测试场景：快照有值+关联表存在、快照有值+关联表不存在、快照为空+关联表存在、快照为空+关联表不存在
2. **集成测试**: 
   - 下单流程测试：验证保存快照
   - 查询流程测试：验证使用快照
   - 多语言测试：验证多语言显示逻辑
3. **回归测试**: 
   - 订单查询接口测试通过
   - 订单打印测试通过
   - 订单导出测试通过
   - 订单报表测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（包含数据库设计、代码实现、多语言处理）
2. **数据库文档**: 
   - 迁移脚本完整（添加字段、数据迁移）
   - 表结构文档更新（`ttpos_sale_bill` 表）
3. **测试文档**: tasks.md 中的测试任务完成（单元测试、集成测试、回归测试）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

### 业务约束

- 快照字段只保存主语言（中文），不保存多语言
- 数据迁移不强制要求所有历史数据立即完整（渐进式实施）
- 历史订单通过降级逻辑兼容（快照为空时使用关联表）

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - ORM 框架
- `main/app/dto` - DTO 定义（`LocaleResponse`）
- `main/app/model` - 数据模型（`SaleBill`, `Nationality`, `MultiLanguageName`）

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖 `Nationality` 和 `MultiLanguageName` 模型
- 依赖订单创建和查询流程

---

## 风险和缓解

### 风险 1: 数据库结构变更风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用 `ALTER TABLE ADD COLUMN` 添加可空字段，不影响现有数据
- 迁移脚本在测试环境充分验证后再上生产
- 执行迁移前备份数据库

### 风险 2: 历史数据不完整

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 实现降级逻辑，快照为空时使用关联表数据
- 渐进式实施，不强制要求所有历史数据立即完整
- 可选执行数据迁移脚本，补充历史数据

### 风险 3: 多语言支持问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 采用"主语言快照 + 关联表补充"的混合方案
- 快照字段保存主语言（中文），其他语言从关联表补充
- 关联表不存在时，所有语言使用快照的主语言填充
- 充分测试多语言场景（关联表存在/不存在）

### 风险 4: 下单逻辑修改风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 全面梳理所有下单入口（POS、扫码点餐、外卖等）
- 确保所有入口都保存快照数据
- 编写集成测试覆盖所有下单场景
- 在测试环境充分验证后再上线

### 风险 5: 回归风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 编写单元测试覆盖所有修改的方法
- 进行回归测试确保不影响现有功能（订单查询、打印、导出、报表等）
- 在测试环境充分验证后再上线
- 准备回滚方案

---

## 时间表

- **Phase 1 - 数据库变更和模型修改**: 0.5 天
  - 编写数据库迁移脚本
  - 修改 `SaleBill` 模型
  - 执行迁移并验证
- **Phase 2 - 查询和下单逻辑修改**: 1-1.5 天
  - 新增 `GetLocaleNationalityName()` 方法
  - 修改订单查询逻辑
  - 修改下单逻辑
  - 添加兼容性处理和多语言支持
- **Phase 3 - 数据检查与迁移**: 0.5 天
  - 检查历史数据完整性
  - 编写数据迁移脚本（可选）
  - 执行数据迁移
- **Phase 4 - 测试验证**: 0.5 天
  - 单元测试（覆盖新增的方法）
  - 集成测试（验证订单查询、下单保存快照）
  - 回归测试（确保不影响现有功能）
- **总计**: 2.5-3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关文档

- `docs/team/proposals/2025-01/order-attribute-snapshot-fix.md` - 父提案（订单商品信息快照修复）
- `docs/shared/api/cashier-order-info-analysis.md` - 订单信息获取逻辑分析
- `main/app/model/sale_bill.go` - SaleBill 模型定义
- `main/app/model/nationality.go` - Nationality 模型定义

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-12/2025-12-02.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-02  
**作者**: xiezhihuan  
**审核者**: {待分配}

