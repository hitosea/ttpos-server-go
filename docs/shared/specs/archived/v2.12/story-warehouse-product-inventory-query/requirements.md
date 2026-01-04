> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 仓库模块商品库存查询功能 需求文档

> 本文档定义 仓库模块商品库存查询功能 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/warehouse-product-inventory-query.md](../../../../team/proposals/2025-12/warehouse-product-inventory-query.md) |
| **创建日期**      | 2025-12-10                                                                                                 |
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

## 📋 概述

当前仓库模块缺少统一的商品库存查询功能，无法根据商品类型（有成本卡/无成本卡）灵活计算库存。不同商品类型的库存计算逻辑不同，需要统一封装为领域服务，遵循 DDD 设计原则，提供清晰的库存查询接口。

**核心价值**：
- **统一库存管理**：提供统一的库存查询接口，简化业务调用
- **灵活扩展**：基于 DDD 设计，便于后续扩展新的商品类型或库存计算规则
- **代码可维护性**：清晰的领域模型和策略模式，提升代码可读性和可维护性
- **业务准确性**：确保不同商品类型的库存计算逻辑正确执行

## 🎯 产品对齐

该功能支持产品在库存管理方面的核心需求，确保不同商品类型（有成本卡/无成本卡）的库存计算逻辑统一、准确，为后续的库存管理、订单处理等功能提供可靠的基础服务。

## 📝 用户故事

**作为** 仓库管理员  
**我想** 查询商品的实时库存  
**以便于** 准确了解商品可售数量，避免超卖或库存积压

**作为** 系统开发者  
**我想** 使用统一的库存查询接口  
**以便于** 简化业务代码，提高代码可维护性

---

## 功能需求

### Requirement 1: 商品库存查询领域服务

**用户故事**: 作为 系统开发者，我想 使用统一的库存查询接口，以便于 简化业务代码，提高代码可维护性

#### 验收标准

1. **WHEN** 调用库存查询接口 **AND** 传入商品ID **THEN** 系统 **SHALL** 自动识别商品类型（有成本卡/无成本卡）并返回对应库存
2. **WHEN** 商品不存在 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **WHEN** 查询库存时发生异常 **THEN** 系统 **SHALL** 记录错误日志并返回友好的错误提示

#### 具体要求

- [ ] 1.1 创建 `ProductInventoryService` 领域服务，提供统一的 `GetProductInventory(productId)` 方法
- [ ] 1.2 自动识别商品类型（通过 `ProductBom.HasProductBomCard()` 判断是否有成本卡）
- [ ] 1.3 根据商品类型选择对应的库存计算策略（有成本卡策略/无成本卡策略）
- [ ] 1.4 返回库存值（float64 类型）

---

### Requirement 2: 有成本卡商品库存计算策略

**用户故事**: 作为 仓库管理员，我想 查询有成本卡商品的实时库存，以便于 准确了解商品可售数量

#### 验收标准

1. **WHEN** 查询有成本卡商品的库存 **AND** 成本卡控制已开启（`UseBomCardStock == 1`） **THEN** 系统 **SHALL** 根据成本卡计算材料用量得到库存
2. **WHEN** 查询有成本卡商品的库存 **AND** 成本卡控制未开启（`UseBomCardStock == 0`） **AND** 商品标记售罄（`IsSoldOut == 1`） **THEN** 系统 **SHALL** 返回库存为 0
3. **WHEN** 查询有成本卡商品的库存 **AND** 成本卡控制未开启 **AND** 商品未标记售罄 **AND** 设置了可售量（`SellableQuantity > 0`） **THEN** 系统 **SHALL** 返回设置的可售量值（`SellableQuantity`）
4. **WHEN** 查询有成本卡商品的库存 **AND** 成本卡控制未开启 **AND** 商品未标记售罄 **AND** 未设置可售量（`SellableQuantity == 0`） **THEN** 系统 **SHALL** 返回 99999999（无限库存）

#### 具体要求

- [ ] 2.1 实现有成本卡商品库存计算策略类（`BomCardProductInventoryStrategy`）
- [ ] 2.2 判断是否开启成本卡控制（检查 `ProductBom.UseBomCardStock` 字段）
- [ ] 2.3 若开启成本卡控制，根据成本卡计算材料用量得到库存
  - [ ] 2.3.1 获取成本卡关联的所有材料（`RelatedMaterial`）
  - [ ] 2.3.2 计算每个材料的可用库存（材料库存 / 材料用量）
  - [ ] 2.3.3 取最小值作为商品库存
- [ ] 2.4 若未开启成本卡控制，执行无成本卡商品的库存计算逻辑
  - [ ] 2.4.1 判断是否标记售罄（`IsSoldOut == 1`），售罄返回 0
  - [ ] 2.4.2 判断是否设置可售量（`SellableQuantity > 0`），设置则返回该值
  - [ ] 2.4.3 否则返回 99999999（无限库存）

---

### Requirement 3: 无成本卡商品库存计算策略

**用户故事**: 作为 仓库管理员，我想 查询无成本卡商品的实时库存，以便于 准确了解商品可售数量

#### 验收标准

1. **WHEN** 查询无成本卡商品的库存 **AND** 商品标记售罄（`IsSoldOut == 1`） **THEN** 系统 **SHALL** 返回库存为 0
2. **WHEN** 查询无成本卡商品的库存 **AND** 商品未标记售罄 **AND** 设置了可售量（`SellableQuantity > 0`） **THEN** 系统 **SHALL** 返回设置的可售量值（`SellableQuantity`）
3. **WHEN** 查询无成本卡商品的库存 **AND** 商品未标记售罄 **AND** 未设置可售量（`SellableQuantity == 0`） **THEN** 系统 **SHALL** 返回 99999999（无限库存）

#### 具体要求

- [ ] 3.1 实现无成本卡商品库存计算策略类（`NonBomCardProductInventoryStrategy`）
- [ ] 3.2 判断是否标记售罄（`IsSoldOut == 1`），售罄返回 0
- [ ] 3.3 判断是否设置可售量（`SellableQuantity > 0`），设置则返回该值
- [ ] 3.4 否则返回 99999999（无限库存）

---

### Requirement 4: 成本卡材料用量计算

**用户故事**: 作为 系统开发者，我想 根据成本卡计算材料用量得到库存，以便于 准确计算有成本卡商品的库存

#### 验收标准

1. **WHEN** 计算成本卡材料用量库存 **AND** 成本卡关联多个材料 **THEN** 系统 **SHALL** 计算每个材料的可用库存并取最小值
2. **WHEN** 计算成本卡材料用量库存 **AND** 某个材料库存不足 **THEN** 系统 **SHALL** 返回该材料限制的库存值
3. **WHEN** 计算成本卡材料用量库存 **AND** 材料用量为 0 或负数 **THEN** 系统 **SHALL** 返回 0 或抛出错误

#### 具体要求

- [ ] 4.1 获取成本卡关联的所有材料（通过 `ProductBomCard` 关联 `RelatedMaterial`）
- [ ] 4.2 遍历每个材料，计算可用库存：
  - [ ] 4.2.1 获取材料库存（`Material.GetStockNum()`）
  - [ ] 4.2.2 获取材料用量（`RelatedMaterial.Num`）
  - [ ] 4.2.3 计算可用库存 = 材料库存 / 材料用量（向下取整）
- [ ] 4.3 取所有材料可用库存的最小值作为商品库存
- [ ] 4.4 处理边界情况（材料用量为 0、材料库存为 0 等）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 DDD 分层架构（Domain → Application → Infrastructure）
- **领域服务**: 在 `main/app/modules/inventory/domain/service/` 下创建 `ProductInventoryService`
- **策略模式**: 使用策略模式处理不同商品类型的库存计算逻辑
- **单一职责原则**: 每个策略类只负责一种商品类型的库存计算
- **依赖管理**: Service 只能依赖 Repository 接口，不能直接依赖 Infrastructure
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 如果后续需要提供 HTTP API，URL 使用 snake_case 命名（如：`/api/v1/product_inventory`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{inventory: float64}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 不涉及新增数据库表，使用现有表结构
- [ ] 使用现有字段：
  - `ttpos_product_bom.use_bom_card_stock` - 是否使用成本卡库存
  - `ttpos_product_bom.is_sold_out` - 是否沽清
  - `ttpos_product_bom.sellable_quantity` - 可售数量
  - `ttpos_product_bom.product_bom_card_uuid` - 成本卡ID
  - `ttpos_related_material` - 关联材料表
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引，避免 N+1 查询）
- [ ] 缓存策略（Redis）：对库存计算结果进行缓存，减少重复计算
- [ ] 并发处理：使用 UUID 锁确保并发安全

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **库存计算逻辑测试覆盖率 100%**（高风险业务逻辑）
- [ ] 集成测试覆盖核心流程
- [ ] 单元测试覆盖所有策略类
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有错误提示使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 输入参数校验（商品ID必须为正整数）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **有成本卡商品库存计算**: 正确识别成本卡控制开关，根据成本卡计算材料用量得到库存
2. **无成本卡商品库存计算**: 正确判断售罄状态和可售量设置，返回对应库存值
3. **统一接口**: 提供统一的库存查询接口，自动识别商品类型并选择对应策略
4. **边界情况处理**: 正确处理材料库存为 0、材料用量为 0、商品不存在等边界情况

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%，Repository ≥ 80%，策略类 100%）
2. **集成测试**: 端到端流程测试通过
3. **性能测试**: 响应时间 < 200ms，支持并发查询

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **代码注释**: 关键逻辑有中文注释
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口或 Repository 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 遵循 DDD 分层架构

### 业务约束

- 库存计算逻辑必须与现有业务逻辑保持一致
- 不能影响现有库存相关功能
- 需要保持向后兼容

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/model/product.go` - ProductBom、ProductBomCard、RelatedMaterial 模型
- `main/app/modules/inventory/` - 库存模块现有结构
- `main/app/repository/` - Repository 层接口

### 服务依赖

- **Main → Main**: 内部服务调用（无需跨服务）

### 业务依赖

- 商品数据（ProductBom）
- 成本卡数据（ProductBomCard）
- 材料数据（Material）
- 关联材料数据（RelatedMaterial）

---

## 风险和缓解

### 风险 1: 成本卡材料用量计算复杂度

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 在 Spec 设计阶段详细分析成本卡材料用量计算逻辑
- 必要时进行技术预研，验证计算逻辑的正确性
- 编写详细的单元测试覆盖各种边界情况

### 风险 2: 性能影响

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 对库存计算结果进行缓存（Redis），减少重复计算
- 优化数据库查询，避免 N+1 查询问题
- 使用索引优化查询性能

### 风险 3: 数据一致性

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 使用数据库事务确保数据一致性
- 在查询时使用适当的锁机制
- 定期校验库存数据的准确性

### 风险 4: 现有代码兼容性

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 先实现新功能，逐步替换现有库存查询逻辑
- 保持向后兼容，不破坏现有功能
- 充分测试现有功能的回归

---

## 时间表

- **Phase 1 - 领域服务设计**: 1 天
- **Phase 2 - 有成本卡商品库存计算策略**: 1-2 天
- **Phase 3 - 无成本卡商品库存计算策略**: 0.5 天
- **Phase 4 - 单元测试和集成测试**: 1 天
- **Phase 5 - 文档和代码审查**: 0.5 天
- **总计**: 3-5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/features/warehouse.md` - 仓库服务架构
- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `main/app/modules/inventory/README.md` - 库存模块说明

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 相关代码

- `main/app/model/product.go` - ProductBom、ProductBomCard、RelatedMaterial 模型
- `main/app/service/warehouse.go` - 仓库服务
- `main/app/modules/inventory/` - 库存模块（DDD 结构）

### 外部参考

- DDD（领域驱动设计）最佳实践
- 策略模式设计模式

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-10  
**作者**: xiezhihuan  
**审核者**: {审核者}

