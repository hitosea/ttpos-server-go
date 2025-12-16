# 原因信息快照修复 需求文档

> 本文档定义原因信息快照修复功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/return-food-reason-snapshot-fix.md](../../../../team/proposals/2025-12/return-food-reason-snapshot-fix.md) |
| **创建日期**      | 2025-12-08                                                                                                 |
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

当前订单查询时，免单原因和退菜原因信息会随后台数据变更而改变，导致订单历史信息不准确。本功能将为 `ttpos_sale_order_product_reason` 表添加原因名称快照字段（JSON 格式），确保订单历史信息准确反映免单/退菜时的真实状态，满足财务、税务对订单历史记录的合规性要求。

**核心价值**：
- 确保订单免单/退菜原因信息准确反映免单/退菜时的状态
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
**我想** 查看历史订单时看到免单/退菜时的真实原因信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到免单/退菜时的真实原因  
**以便于** 准确处理退款和客户咨询

---

## 功能需求

### Requirement 1: 数据库结构变更 - 添加原因名称快照字段

**用户故事**: 作为开发者，我想在 `ttpos_sale_order_product_reason` 表中添加原因名称快照字段，以便于保存免单/退菜时的原因信息。

#### 验收标准

1. **WHEN** 执行数据库迁移脚本 **THEN** 系统 **SHALL** 在 `ttpos_sale_order_product_reason` 表添加 `name` 字段
2. **WHEN** 添加字段后 **THEN** 字段类型 **SHALL** 为 `TEXT`，默认值为空字符串
3. **WHEN** 添加字段后 **THEN** 字段注释 **SHALL** 为 "原因名称快照（JSON），不随后台更新"
4. **WHEN** 执行迁移 **THEN** 系统 **SHALL** 不影响现有数据和业务功能

#### 具体要求

- [ ] 1.1 在 `ttpos_sale_order_product_reason` 表添加 `name` 字段（TEXT 类型，存储 JSON）
- [ ] 1.2 字段位置：在 `gift_reason_uuid` 字段之后
- [ ] 1.3 字段默认值：空字符串 `''`
- [ ] 1.4 字段注释：明确说明为快照字段（JSON 格式），不随后台更新
- [ ] 1.5 迁移脚本支持可重复执行（幂等性）

---

### Requirement 2: 数据模型修改 - 添加 Name 字段

**用户故事**: 作为开发者，我想在 `SaleOrderProductReason` 模型中添加 `Name` 字段，以便于在代码中操作快照数据。

#### 验收标准

1. **WHEN** 修改 `SaleOrderProductReason` 结构体 **THEN** 系统 **SHALL** 添加 `Name` 字段
2. **WHEN** 字段定义完成 **THEN** GORM 标签 **SHALL** 正确映射到数据库字段 `name`
3. **WHEN** 字段定义完成 **THEN** JSON 标签 **SHALL** 为 `name`
4. **WHEN** 编译代码 **THEN** 系统 **SHALL** 无编译错误

#### 具体要求

- [ ] 2.1 在 `SaleOrderProductReason` 结构体添加 `Name string` 字段
- [ ] 2.2 添加 GORM 标签：`gorm:"column:name;type:text"`
- [ ] 2.3 添加 JSON 标签：`json:"name"`
- [ ] 2.4 字段位置：紧跟 `GiftReasonUuid` 字段之后

---

### Requirement 3: 查询逻辑修改 - 实现免单原因获取方法（JSON 方案）

**用户故事**: 作为开发者，我想实现免单原因获取方法，以便于优先使用快照数据，降级使用关联表数据。

#### 验收标准

1. **WHEN** 调用 `GetFreeReason()` 方法 **THEN** 系统 **SHALL** 优先返回快照字段 `SaleOrderProductReason.Name`（JSON 格式）
2. **IF** 快照字段为空或 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用 `FreeReason.MultiLanguageName` 关联表数据
3. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）
4. **IF** 快照字段有值（JSON）**THEN** 系统 **SHALL** 解析 JSON 并返回完整多语言数据（所有语言）

#### 具体要求

- [ ] 3.1 修改 `SaleOrder.GetFreeReason()` 方法
- [ ] 3.2 实现快照优先逻辑：优先使用 `SaleOrderProductReason.Name` 字段（JSON）
- [ ] 3.3 实现 JSON 解析：将快照字段反序列化为 `LocaleResponse`
- [ ] 3.4 实现降级逻辑：快照为空或解析失败时使用关联表数据
- [ ] 3.5 方法返回类型：`dto.LocaleResponse`
- [ ] 3.6 添加方法注释，说明逻辑和用途

---

### Requirement 4: 查询逻辑修改 - 实现退菜原因获取方法（JSON 方案）

**用户故事**: 作为开发者，我想实现退菜原因获取方法，以便于优先使用快照数据，降级使用关联表数据。

#### 验收标准

1. **WHEN** 调用 `GetCancelReason()` 方法 **THEN** 系统 **SHALL** 优先返回快照字段 `SaleOrderProductReason.Name`（JSON 格式）
2. **IF** 快照字段为空或 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用 `ReturnFoodReason.MultiLanguageName` 关联表数据
3. **WHEN** 返回数据 **THEN** 格式 **SHALL** 为 `dto.LocaleResponse`（多语言）
4. **IF** 快照字段有值（JSON）**THEN** 系统 **SHALL** 解析 JSON 并返回完整多语言数据（所有语言）

#### 具体要求

- [ ] 4.1 修改 `SaleOrderProduct.GetCancelReason()` 方法
- [ ] 4.2 实现快照优先逻辑：优先使用 `SaleOrderProductReason.Name` 字段（JSON）
- [ ] 4.3 实现 JSON 解析：将快照字段反序列化为 `LocaleResponse`
- [ ] 4.4 实现降级逻辑：快照为空或解析失败时使用关联表数据
- [ ] 4.5 方法返回类型：`dto.LocaleResponse`
- [ ] 4.6 添加方法注释，说明逻辑和用途

---

### Requirement 5: 下单逻辑修改 - 保存免单原因快照（JSON 方案）

**用户故事**: 作为开发者，我想在创建免单原因时保存快照字段，以便于新订单自动使用快照机制。

#### 验收标准

1. **WHEN** 创建新免单原因 **THEN** 系统 **SHALL** 正确保存快照字段（`SaleOrderProductReason.Name`）
2. **WHEN** 保存快照字段 **THEN** 数据格式 **SHALL** 为 JSON 字符串（包含所有语言）
3. **WHEN** 序列化多语言数据 **THEN** 系统 **SHALL** 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
4. **IF** 多语言数据为空或无效 **THEN** 系统 **SHALL** 保存空字符串（降级处理）

#### 具体要求

- [ ] 5.1 修改 `SaleOrder.NewFreeOrderReason()` 方法
- [ ] 5.2 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
- [ ] 5.3 序列化多语言数据为 JSON 字符串
- [ ] 5.4 保存 JSON 字符串到 `SaleOrderProductReason.Name` 字段
- [ ] 5.5 添加错误处理：序列化失败时记录日志，但不中断流程

---

### Requirement 6: 下单逻辑修改 - 保存退菜原因快照（JSON 方案）

**用户故事**: 作为开发者，我想在创建退菜原因时保存快照字段，以便于新订单自动使用快照机制。

#### 验收标准

1. **WHEN** 创建新退菜原因 **THEN** 系统 **SHALL** 正确保存快照字段（`SaleOrderProductReason.Name`）
2. **WHEN** 保存快照字段 **THEN** 数据格式 **SHALL** 为 JSON 字符串（包含所有语言）
3. **WHEN** 序列化多语言数据 **THEN** 系统 **SHALL** 从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据
4. **IF** 多语言数据为空或无效 **THEN** 系统 **SHALL** 保存空字符串（降级处理）

#### 具体要求

- [ ] 6.1 修改 `SaleOrderProduct.NewSaleOrderProductReasonList()` 方法
- [ ] 6.2 从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据
- [ ] 6.3 序列化多语言数据为 JSON 字符串
- [ ] 6.4 保存 JSON 字符串到 `SaleOrderProductReason.Name` 字段
- [ ] 6.5 添加错误处理：序列化失败时记录日志，但不中断流程

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 快照字段使用 TEXT 类型，存储 JSON 格式
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] JSON 解析性能优化（避免重复解析）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **Order 相关模块测试覆盖率 100%**（高风险）
- [ ] 集成测试覆盖核心流程（免单/退菜保存快照）
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 快照字段保存完整多语言 JSON（包含所有语言）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] JSON 解析安全（防止恶意 JSON）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] JSON 解析失败时降级使用关联表数据

---

## 验收标准

### 功能验收

1. **数据库迁移**: 成功添加 `name` 字段，不影响现有数据
2. **免单原因快照**: 创建免单原因时正确保存 JSON 快照
3. **退菜原因快照**: 创建退菜原因时正确保存 JSON 快照
4. **免单原因查询**: 查询时优先使用快照字段，降级使用关联表
5. **退菜原因查询**: 查询时优先使用快照字段，降级使用关联表
6. **多语言支持**: 快照字段包含完整多语言数据（所有语言）
7. **历史数据兼容**: 历史订单（快照字段为空）正常显示（降级逻辑）

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（免单/退菜保存和查询）
4. **回归测试**: 确保不影响现有功能（订单查询、打印、导出等）

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **数据库文档**: 迁移脚本和表结构文档完整
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
- JSON 序列化/反序列化使用标准库 `encoding/json`

### 业务约束

- 快照字段必须保存完整多语言数据（JSON 格式）
- 历史订单必须兼容（降级逻辑）
- 不影响现有订单查询、打印、导出等功能

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `encoding/json` - JSON 序列化/反序列化
- `dto.LocaleResponse` - 多语言响应格式

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖 `FreeReason` 和 `ReturnFoodReason` 模型
- 依赖 `MultiLanguageName` 模型
- 依赖订单创建和查询流程

---

## 风险和缓解

### 风险 1: 数据库结构变更风险

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用 `ALTER TABLE ADD COLUMN` 添加可空字段，不影响现有数据
- 迁移脚本支持可重复执行（幂等性）
- 在测试环境充分验证后再上线

### 风险 2: 历史数据不完整

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 实现降级逻辑，历史订单正常显示（使用关联表数据）
- 逐步迁移，不强制要求所有数据立即完整
- 新订单自动使用快照机制

### 风险 3: JSON 解析失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 实现 JSON 解析错误处理
- 解析失败时降级使用关联表数据
- 记录错误日志，便于排查问题

### 风险 4: 下单逻辑修改风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 全面梳理所有免单/退菜入口，确保都保存快照数据
- 添加错误处理，序列化失败时记录日志但不中断流程
- 充分测试，特别是订单相关的所有功能

### 风险 5: 回归风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 编写单元测试覆盖所有修改的方法和场景
- 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
- 在测试环境充分验证后再上线

---

## 时间表

- **Phase 1 - 数据库结构变更**: 0.5 天
- **Phase 2 - 代码修改（免单原因）**: 0.5-1 天
- **Phase 3 - 代码修改（退菜原因）**: 0.5-1 天
- **Phase 4 - 数据检查与迁移**: 0.5 天
- **Phase 5 - 测试验证**: 0.5-1 天
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
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 相关 Spec

- `story-main-order-source-snapshot-fix` - 外卖来源快照修复（参考实现）
- `story-main-buffet-package-name-snapshot-fix` - 自助餐名称快照修复（参考实现）
- `story-main-nationality-snapshot-fix` - 国籍信息快照修复（参考实现）

### 外部参考

- [JSON 序列化最佳实践](https://golang.org/pkg/encoding/json/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: xiezhihuan  
**审核者**: {审核者}

