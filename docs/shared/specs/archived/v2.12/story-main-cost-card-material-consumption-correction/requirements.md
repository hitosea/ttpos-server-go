> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 成本卡材料消耗修正 需求文档

> 本文档定义 成本卡材料消耗修正 功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/cost-card-material-consumption-correction.md](../../../../team/proposals/2025-12/cost-card-material-consumption-correction.md) |
| **创建日期**      | 2025-12-12                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概念说明

为便于理解，明确以下概念：

- **商品（Product）**：TTPOS 中售卖给顾客的商品，如菜品、套餐等
- **物品/材料（Material）**：同一个事物的不同称呼，指原材料、配料等库存物品
- **成本卡（BOM Card）**：定义商品与材料的关系，一个商品可以根据成本卡由多个材料组成
- **商品库存**：基于成本卡计算的商品可售数量，计算公式为：材料库存 / 材料用量（取最小值）

---

## 📋 概述

商户的某些**已完成订单**因为成本卡设置的材料消耗量错误，导致材料出库记录错误、材料库存不准确、商品库存错误、ERP 数据不一致以及每日销售出库统计错误。本功能旨在提供一个批量修正工具，能够退回错误扣减的材料、修正材料库存和商品库存、重新计算材料消耗、重新生成出库记录、重新同步 ERP 数据，并修正受影响的每日销售出库记录，确保数据准确性和一致性。

**重要约束**：只能选择已经完成的订单（订单状态为已结账且完成时间 > 0）进行修正。

## 🎯 产品对齐

本功能支持产品愿景中的**数据准确性**和**系统一致性**目标，确保 POS 系统与 ERP 系统的数据一致，为商户提供准确的库存管理和经营分析数据，符合财务和审计要求。

## 📝 用户故事

**作为** 商户管理员  
**我想** 批量修正因成本卡错误导致的订单材料消耗、出库记录以及商品库存错误  
**以便于** 恢复准确的材料库存和商品库存数据，确保 POS 与 ERP 系统数据一致，获得准确的经营统计

---

## 功能需求

### Requirement 1: 订单选择与识别

**用户故事**: 作为商户管理员，我想自主选择需要修正的订单，以便于精确修正有问题的订单

#### 验收标准

1. **WHEN** 管理员自主选择需要修正的订单 **THEN** 系统 **SHALL** 只能选择已经完成的订单（订单状态为已结账且完成时间 > 0），并能够识别并列出订单中使用错误成本卡的商品
2. **WHEN** 选择订单后 **THEN** 系统 **SHALL** 显示订单的基本信息（订单号、完成时间、商品列表等）
3. **WHEN** 识别到错误成本卡商品 **THEN** 系统 **SHALL** 显示该商品使用的成本卡信息和材料消耗详情

#### 具体要求

- [ ] 1.1 提供订单选择界面，支持用户手动选择订单（订单号列表或订单详情页）
- [ ] 1.2 **约束条件**：只能选择已经完成的订单（订单状态为已结账且完成时间 > 0）
- [ ] 1.3 系统自动识别订单中使用成本卡的商品（规格商品、加料等）
- [ ] 1.4 显示每个订单中使用错误成本卡的商品列表
- [ ] 1.5 显示订单的创建时间、完成时间等关键信息
- [ ] 1.6 注：后期可考虑添加按订单号、时间范围、商品等条件的过滤功能

---

### Requirement 2: 材料退回处理

**用户故事**: 作为商户管理员，我想将错误扣减的材料退回，以便于修正材料库存和商品库存

#### 验收标准

1. **WHEN** 执行修正操作 **THEN** 系统 **SHALL** 先将错误扣减的材料退回，修正材料库存
2. **WHEN** 退回材料时 **THEN** 系统 **SHALL** 根据订单历史出库记录计算需要退回的材料数量
3. **WHEN** 材料退回完成 **THEN** 系统 **SHALL** 更新材料库存，确保库存数据准确
4. **WHEN** 材料库存修正完成 **THEN** 系统 **SHALL** 自动更新相关商品的库存（基于成本卡计算：材料库存/材料用量）

#### 具体要求

- [ ] 2.1 查询订单的历史出库记录（WarehouseOutFormItem）
- [ ] 2.2 根据历史出库记录计算每个材料需要退回的数量
- [ ] 2.3 执行材料退回操作，增加材料库存（WarehouseItem.Stock）
- [ ] 2.4 记录材料退回日志（WarehouseInOutLog）
- [ ] 2.5 更新规格/加料关联材料库存（RelatedMaterial）
- [ ] 2.6 重新计算相关商品的库存（根据成本卡计算：材料库存/材料用量）
- [ ] 2.7 使用数据库事务确保数据一致性

---

### Requirement 3: 重新计算材料消耗

**用户故事**: 作为商户管理员，我想根据当前正确的成本卡重新计算材料消耗，以便于生成正确的出库记录

#### 验收标准

1. **WHEN** 材料退回完成 **THEN** 系统 **SHALL** 根据当前正确的成本卡重新计算材料消耗并生成出库记录
2. **WHEN** 重新计算时 **THEN** 系统 **SHALL** 使用订单商品当前关联的成本卡配置
3. **WHEN** 成本卡不存在或无效 **THEN** 系统 **SHALL** 提示错误并中止修正流程

#### 具体要求

- [ ] 3.1 获取订单商品当前关联的成本卡（ProductBomCard）
- [ ] 3.2 根据成本卡配置计算材料消耗量（参考 flavorUseCard 逻辑）
- [ ] 3.3 考虑商品数量、成本卡加工份数等因素
- [ ] 3.4 生成新的材料消耗记录（SaleOrderProductBom）
- [ ] 3.5 验证成本卡配置的有效性（材料是否存在、单位是否正确等）

---

### Requirement 4: 重新生成出库记录

**用户故事**: 作为商户管理员，我想基于正确的材料消耗量重新生成出库记录，以便于更新出库单和出库日志，并修正商品库存

#### 验收标准

1. **WHEN** 出库记录生成完成 **THEN** 系统 **SHALL** 重新向 ERP 发送 POS invoice 数据
2. **WHEN** 生成出库记录时 **THEN** 系统 **SHALL** 基于正确的材料消耗量创建出库单和出库日志
3. **WHEN** 材料库存更新完成 **THEN** 系统 **SHALL** 重新计算相关商品的库存（基于成本卡：材料库存/材料用量）
4. **WHEN** 出库记录生成失败 **THEN** 系统 **SHALL** 能够回滚已执行的操作

#### 具体要求

- [ ] 4.1 基于重新计算的材料消耗量创建出库单（WarehouseOutForm）
- [ ] 4.2 创建出库单明细（WarehouseOutFormItem）
- [ ] 4.3 记录出库日志（WarehouseInOutLog）
- [ ] 4.4 扣减材料库存（WarehouseItem.Stock）
- [ ] 4.5 更新规格/加料关联材料库存（RelatedMaterial）
- [ ] 4.6 重新计算相关商品的库存（根据成本卡计算：材料库存/材料用量）
- [ ] 4.7 关联订单商品（SaleOrderProductUuid）

---

### Requirement 5: ERP 数据同步

**用户故事**: 作为商户管理员，我想重新向 ERP 发送 POS invoice 数据，以便于确保 ERP 系统中的数据与 POS 系统一致

#### 验收标准

1. **WHEN** ERP 数据同步完成 **THEN** 系统 **SHALL** 重新统计并修正受影响的每日销售出库记录
2. **WHEN** 重新发送 ERP 数据时 **THEN** 系统 **SHALL** 处理可能的重复或冲突情况
3. **WHEN** ERP 同步失败 **THEN** 系统 **SHALL** 记录错误并允许重试

#### 具体要求

- [ ] 5.1 重新生成 POS invoice 数据（包含订单商品和材料消耗）
- [ ] 5.2 调用 ERP 接口保存 POS invoice（参考 ttpos-bmp selling.go）
- [ ] 5.3 处理 ERP 返回的结果（成功/失败/冲突）
- [ ] 5.4 如需要，先删除旧数据再插入新数据
- [ ] 5.5 记录 ERP 同步日志

---

### Requirement 6: 每日销售出库修正

**用户故事**: 作为商户管理员，我想重新统计并修正受影响的每日销售出库记录，以便于获得准确的经营统计

#### 验收标准

1. **WHEN** ERP 数据同步完成 **THEN** 系统 **SHALL** 重新统计并修正受影响的每日销售出库记录
2. **WHEN** 修正每日销售出库时 **THEN** 系统 **SHALL** 识别受影响的日期范围
3. **WHEN** 涉及多天数据时 **THEN** 系统 **SHALL** 支持批量修正多天的数据

#### 具体要求

- [ ] 6.1 识别受影响的日期范围（根据订单的营业日期）
- [ ] 6.2 重新统计每日营业结束生成的销售出库记录（参考 salesOutInventoryRecord）
- [ ] 6.3 更新或重新生成每日销售出库记录（ErpInventoryRecord）
- [ ] 6.4 支持批量修正多天的数据
- [ ] 6.5 确保每日销售出库记录的准确性

---

### Requirement 7: 操作日志与审计

**用户故事**: 作为商户管理员，我想查看修正操作的详细日志，以便于审计和追踪

#### 验收标准

1. **WHEN** 修正操作完成 **THEN** 系统 **SHALL** 记录详细的操作日志，包括修正的订单、材料、数量等信息
2. **IF** 修正过程中出现错误 **THEN** 系统 **SHALL** 能够回滚已执行的操作，恢复到修正前的状态
3. **WHEN** 需要回滚时 **THEN** 系统 **SHALL** 根据操作日志执行回滚操作

#### 具体要求

- [ ] 7.1 记录修正操作的详细信息（操作时间、操作人、订单列表、材料列表、数量等）
- [ ] 7.2 记录每一步操作的执行结果（成功/失败）
- [ ] 7.3 提供操作日志查询界面
- [ ] 7.4 支持操作回滚功能（如需要）
- [ ] 7.5 记录回滚操作的日志

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/cost_card_correction`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 分页信息统一放在 meta 中
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] 金额字段使用 decimal(20,8)
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 批量处理时响应时间 < 30s（每批 100 个订单）
- [ ] 数据库查询优化（使用索引）
- [ ] 分批处理大批量订单，避免一次性处理过多数据
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **材料出库相关模块测试覆盖率 100%**（高风险）
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
- [ ] 故障恢复机制（支持回滚）

---

## 验收标准

### 功能验收

1. **订单选择**: 管理员能够自主选择需要修正的订单（只能选择已完成的订单：订单状态为已结账且完成时间 > 0），系统能够识别并列出使用错误成本卡的商品
2. **材料退回**: 系统能够根据订单历史出库记录计算并退回错误扣减的材料，修正材料库存
3. **商品库存修正**: 系统能够根据修正后的材料库存重新计算相关商品的库存（基于成本卡：材料库存/材料用量）
4. **材料消耗重新计算**: 系统能够根据当前正确的成本卡重新计算材料消耗量
5. **出库记录重新生成**: 系统能够基于正确的材料消耗量重新生成出库记录，并更新商品库存
6. **ERP 数据同步**: 系统能够重新向 ERP 发送 POS invoice 数据，确保数据一致
7. **每日销售出库修正**: 系统能够重新统计并修正受影响的每日销售出库记录
8. **操作日志**: 系统能够记录详细的操作日志，支持审计和回滚

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%, 材料出库相关 100%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过（订单选择 → 材料退回 → 重新计算 → 出库记录 → ERP 同步 → 每日销售出库修正）
4. **手动测试**: 浏览器兼容性测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整（如有）
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

- 修正操作必须在事务中执行，确保数据一致性
- 修正过程中需要锁定相关订单，避免并发操作
- 支持分批处理大批量订单，避免性能问题
- ERP 同步失败时需要能够重试

### 资源约束

- 开发时间: 10-15 天
- Story Point: 13-21 (必须 ≤ 5，需要拆分)

---

## 依赖关系

### 技术依赖

- `main/app/model/product.go` - ProductBomCard 成本卡模型
- `main/app/model/warehouse_form.go` - WarehouseOutForm 出库单模型
- `main/app/model/sale_order.go` - SaleOrder 订单模型
- `main/app/service/material.go` - MaterialService 材料服务

### 服务依赖

- **Main → BMP**: gRPC 调用 ERP 同步接口（ttpos-bmp selling.go）

### 业务依赖

- 订单必须已结账（有出库记录）
- 成本卡配置必须正确（修正时使用）
- ERP 系统必须可用（同步时）

### 数据库表依赖

- `ttpos_sale_bill` - 销售单据表
- `ttpos_sale_order` - 销售订单表
- `ttpos_sale_order_product` - 销售订单商品表
- `ttpos_sale_order_product_bom` - 销售订单商品BOM表（材料消耗记录）
- `ttpos_sale_order_material` - 销售订单材料表
- `ttpos_warehouse_out_form` - 出库单表
- `ttpos_warehouse_out_form_item` - 出库单明细表
- `ttpos_warehouse_in_out_log` - 仓库出入库日志表
- `ttpos_warehouse_item` - 仓库物品库存表
- `ttpos_product_bom` - 商品BOM表（关联成本卡）
- `ttpos_product_bom_card` - 成本卡表
- `ttpos_related_material` - 成本卡关联材料表（定义成本卡的材料组成）

---

## 风险和缓解

### 风险 1: 数据一致性风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用数据库事务确保数据一致性
- 必要时使用分布式事务
- 详细记录操作日志，支持回滚

### 风险 2: 性能风险

**影响**: 中  
**概率**: 高  
**缓解措施**:

- 分批处理大批量订单（每批 100 个）
- 优化数据库查询（使用索引）
- 异步处理非关键步骤

### 风险 3: ERP 同步风险

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 与 ERP 系统确认接口支持
- 必要时先删除旧数据再插入新数据
- 实现重试机制
- 记录同步日志

### 风险 4: 回滚风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 详细记录每一步操作
- 实现操作回滚功能
- 支持手动回滚

### 风险 5: 并发风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 修正过程中锁定相关订单
- 使用 UUID 锁避免并发操作
- 检查订单状态，避免重复修正

---

## 时间表

- **Phase 1 - 需求分析与技术方案设计**: 2-3 天
- **Phase 2 - 订单选择与材料退回逻辑**: 2-3 天
- **Phase 3 - 重新计算材料消耗与出库**: 2-3 天
- **Phase 4 - ERP 数据同步**: 2-3 天
- **Phase 5 - 每日销售出库修正**: 2-3 天
- **Phase 6 - 测试与优化**: 2-3 天
- **总计**: 10-15 天（SP = 13-21，需要拆分）

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
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- 订单反结账功能（salesOutReverse）- 类似功能参考
- ERP POS invoice 接口文档

### 相关代码位置

- 成本卡相关：`main/app/model/product.go` (ProductBomCard)
- 材料出库相关：`main/app/model/warehouse_form.go` (WarehouseOutForm)
- 订单材料消耗：`main/app/model/sale_order.go` (flavorUseCard)
- ERP 同步：`ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
- 销售出库记录：`admin/app/common/model/product/Product.php` (salesOutInventoryRecord)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-12  
**作者**: xiezhihuan  
**审核者**: {审核者}
