> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 赠菜原因快照修复 需求文档

> 本文档定义赠菜原因快照修复功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/gift-reason-snapshot-fix.md](../../../../team/proposals/2025-12/gift-reason-snapshot-fix.md) |
| **创建日期**      | 2025-12-09                                                                                                 |
| **负责人**        | xiezhihuan                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
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

当前订单查询时，赠菜原因会随后台数据变更而改变，导致订单历史信息不准确。本功能将使用 `ttpos_sale_order_product_reason` 表现有的 `name` 字段（TEXT 类型，JSON 快照），修复 `GetGiftReason()` 方法优先使用快照字段，确保订单历史信息准确反映下单时的真实状态，满足财务、税务对订单历史记录的合规性要求。

**核心价值**：
- 确保订单赠菜原因信息准确反映下单时的状态
- 支持多语言快照，满足国际化需求
- 满足财务、税务对订单历史记录的合规性要求
- 支持订单历史查询和问题追溯
- 避免因数据变更导致的业务逻辑错误

## 🎯 产品对齐

本功能是"订单商品信息快照修复"（`order-attribute-snapshot-fix.md`）系列需求的一部分，通过建立完整的订单快照机制，确保：
- **数据一致性**：订单信息作为历史快照，不随后台配置变更而改变
- **多语言支持**：支持多语言快照，满足国际化需求
- **业务可靠性**：支持订单对账、报表、审计等关键业务场景
- **合规性**：满足餐饮行业对历史订单数据的监管要求

## 📝 用户故事

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实赠菜原因信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的赠菜原因  
**以便于** 准确处理客户咨询

---

## 功能需求

### Requirement 1: 查询逻辑修改 - 修复赠菜原因获取方法

**用户故事**: 作为开发者，我想修复 `GetGiftReason()` 方法，以便于优先使用快照字段，降级使用关联表数据。

#### 验收标准

1. **WHEN** 查询包含赠菜原因的订单 **THEN** 系统 **SHALL** 优先使用 `SaleOrderProductReason.Name` 字段（JSON 快照）
2. **IF** 快照字段为空或 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用 `GiftReason.MultiLanguageName`（兼容历史数据）
3. **WHEN** 查询订单赠菜原因 **THEN** 系统 **SHALL** 返回多语言格式（`dto.LocaleResponse`）
4. **IF** 后台删除了某个赠菜原因 **THEN** 历史订单 **SHALL** 仍然显示该原因的原始名称（从快照字段获取）

#### 具体要求

- [ ] 1.1 修改 `SaleOrderProduct.GetGiftReason()` 方法（`main/app/model/sale_order_product.go:1073`）
- [ ] 1.2 优先使用 `SaleOrderProductReason.Name` 字段（JSON 快照）
- [ ] 1.3 解析 JSON 为 `dto.LocaleResponse`（包含所有语言：ZH、TH、EN、ZHTW、JA、KO、MY、TR、SV）
- [ ] 1.4 快照字段为空或解析失败时，降级使用 `GiftReason.MultiLanguageName`
- [ ] 1.5 支持自定义赠菜原因（`SaleOrderProduct.GiftReason` 字段）
- [ ] 1.6 返回多语言格式（`dto.LocaleResponse`），多个原因用"、"分隔

---

### Requirement 2: 下单逻辑修改 - 确保保存快照

**用户故事**: 作为开发者，我想在创建赠菜原因时保存快照字段，以便于新订单使用快照机制。

#### 验收标准

1. **WHEN** 创建新订单并选择赠菜原因 **THEN** 系统 **SHALL** 正确保存赠菜原因快照字段（JSON 格式）
2. **WHEN** 保存快照字段 **THEN** 系统 **SHALL** 将多语言数据序列化为 JSON 字符串
3. **IF** 多语言名称为空 **THEN** 系统 **SHALL** 将快照字段设置为空字符串
4. **WHEN** 保存快照失败 **THEN** 系统 **SHALL** 记录错误日志但不中断流程

#### 具体要求

- [ ] 2.1 检查所有创建 `SaleOrderProductReason` 的地方，确保 `GiftReasonUuid` 不为空时保存快照
- [ ] 2.2 在创建赠菜原因时，从 `GiftReason.MultiLanguageName` 获取多语言数据
- [ ] 2.3 将多语言数据序列化为 JSON 字符串，保存到 `SaleOrderProductReason.Name` 字段
- [ ] 2.4 如果多语言名称为空，将快照字段设置为空字符串
- [ ] 2.5 如果序列化失败，记录错误日志但不中断流程

---

### Requirement 3: 兼容性处理

**用户故事**: 作为开发者，我想实现兼容性处理，以便于历史订单正常显示。

#### 验收标准

1. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
2. **IF** 快照字段 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用关联表数据
3. **WHEN** 查询历史订单 **THEN** 系统 **SHALL** 正常显示赠菜原因（即使快照字段为空）

#### 具体要求

- [ ] 3.1 实现降级逻辑：快照字段为空时，使用关联表数据
- [ ] 3.2 实现 JSON 解析错误处理：解析失败时，降级使用关联表数据
- [ ] 3.3 确保历史订单正常显示，不影响现有功能

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 使用现有字段 `ttpos_sale_order_product_reason.name`（TEXT 类型）
- [ ] 字段存储 JSON 格式的多语言数据
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] JSON 解析性能优化
- [ ] 避免不必要的数据库查询（优先使用快照数据）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] `GetGiftReason()` 方法测试覆盖率 ≥ 80%
- [ ] 测试场景包括：快照有值/无值、JSON 有效/无效、关联表有数据/无数据
- [ ] 集成测试覆盖订单查询场景
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 快照字段保存完整的多语言 JSON
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] JSON 解析安全（防止恶意 JSON）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] JSON 解析失败时优雅降级
- [ ] 错误日志记录（使用 Logger）
- [ ] 确保历史订单正常显示（兼容性）

---

## 验收标准

### 功能验收

1. **赠菜原因查询**: 查询包含赠菜原因的订单时，系统显示下单时保存的赠菜原因快照
2. **数据删除兼容**: 后台删除了某个赠菜原因后，历史订单仍然显示该原因的原始名称
3. **数据改名兼容**: 后台修改了某个赠菜原因的名称后，历史订单显示修改前的原始名称
4. **历史数据兼容**: 订单快照数据为空（历史数据）时，系统降级使用关联表数据
5. **新订单快照**: 创建新订单并选择赠菜原因时，系统正确保存赠菜原因快照字段（JSON 格式）
6. **多语言支持**: 查询订单赠菜原因时，系统返回多语言格式（`LocaleResponse`）
7. **JSON 解析容错**: 快照字段 JSON 解析失败时，系统降级使用关联表数据

### 测试验收

1. **单元测试**: `GetGiftReason()` 方法测试覆盖率 ≥ 80%
2. **集成测试**: 订单查询场景测试通过
3. **回归测试**: 确保不影响现有功能（订单查询、打印、导出等）

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **代码注释**: 关键方法添加注释说明快照逻辑
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

### 业务约束

- 必须兼容历史数据（快照字段为空的情况）
- 必须支持多语言（10 种语言）
- 必须保证数据一致性（快照不随后台变更）

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2-3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `encoding/json` - JSON 序列化/反序列化
- `ttpos-server-go/app/dto` - LocaleResponse 类型定义

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖 `SaleOrderProductReason` 模型（已存在）
- 依赖 `GiftReason` 模型（已存在）
- 依赖 `MultiLanguageName` 模型（已存在）

---

## 风险和缓解

### 风险 1: 历史数据不完整

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 实现降级逻辑，确保历史订单正常显示
- 逐步迁移，不强制要求所有数据立即完整

### 风险 2: JSON 解析失败

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 实现 JSON 解析错误处理，降级使用关联表数据
- 添加错误日志记录，便于排查问题

### 风险 3: 下单逻辑遗漏

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 全面梳理所有创建 `SaleOrderProductReason` 的地方
- 确保所有下单入口都保存快照数据
- 添加单元测试验证快照保存逻辑

### 风险 4: 回归风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 充分测试，特别是订单查询、打印、导出等功能
- 在测试环境充分验证后再上线

---

## 时间表

- **Phase 1 - 查询逻辑修改**: 0.5-1 天
- **Phase 2 - 下单逻辑修改**: 0.5 天
- **Phase 3 - 测试验证**: 0.5 天
- **总计**: 1-2 天（SP = 2-3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关文档

- 主提案: [订单商品信息快照修复](../../../../team/proposals/2025-01/order-attribute-snapshot-fix.md)
- 类似功能: [原因信息快照修复](../story-main-reason-snapshot-fix/requirements.md)（免单原因、退菜原因）

### 代码位置

- **问题代码**: `main/app/model/sale_order_product.go:1073` - `GetGiftReason()` 方法
- **数据模型**: `main/app/model/order.go:466-485` - `SaleOrderProductReason` 模型定义
- **关联模型**: `main/app/model/reason.go` - `GiftReason` 模型定义

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}


