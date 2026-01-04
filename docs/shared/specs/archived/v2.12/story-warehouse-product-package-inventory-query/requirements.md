> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# 仓库模块商品包库存查询功能 需求文档

> 本文档定义 仓库模块商品包库存查询功能 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/warehouse-product-package-inventory-query.md](../../../../team/proposals/2025-12/warehouse-product-package-inventory-query.md) |
| **创建日期**      | 2025-12-11                                                                                                 |
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

当前仓库模块已实现了商品BOM（ProductBom）的库存查询功能，但缺少商品包（ProductPackage）的库存查询接口。在实际业务场景中，一个商品包对应多个商品BOM，需要根据商品包下所有BOM的库存情况来计算商品包的整体库存。

**核心价值**：
- **统一库存管理**：提供商品包级别的库存查询接口，完善库存管理体系
- **业务准确性**：明确商品包库存计算规则（取最小值），确保库存计算的准确性
- **代码复用**：基于已有的商品BOM库存查询功能，复用现有策略和逻辑
- **性能优化**：通过批量查询和缓存机制，提升商品包库存查询性能

## 🎯 产品对齐

该功能支持产品在库存管理方面的核心需求，在已有商品BOM库存查询功能的基础上，扩展商品包级别的库存查询能力，为商品包相关的业务场景（如商品列表展示、库存预警等）提供可靠的库存数据支持。

## 📝 用户故事

**作为** 仓库管理员  
**我想** 查询商品包的实时库存  
**以便于** 准确了解商品包可售数量，避免超卖或库存积压

**作为** 系统开发者  
**我想** 使用统一的商品包库存查询接口  
**以便于** 简化业务代码，提高代码可维护性

---

## 功能需求

### Requirement 1: 商品包库存查询领域服务方法

**用户故事**: 作为 系统开发者，我想 使用统一的商品包库存查询接口，以便于 简化业务代码，提高代码可维护性

#### 验收标准

1. **WHEN** 调用商品包库存查询接口 **AND** 传入商品包UUID **THEN** 系统 **SHALL** 查询商品包下所有BOM并返回最小值库存
2. **WHEN** 商品包不存在 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **WHEN** 商品包下没有BOM **THEN** 系统 **SHALL** 返回0或抛出明确的错误信息
4. **WHEN** 查询库存时发生异常 **THEN** 系统 **SHALL** 记录错误日志并返回友好的错误提示

#### 具体要求

- [ ] 1.1 在 `IProductInventoryDomainService` 接口中新增 `GetProductPackageInventory(productPackageUuid)` 方法
- [ ] 1.2 通过 `FindByProductPackageUuid` 查询商品包下所有BOM
- [ ] 1.3 遍历每个BOM，调用 `GetProductInventory` 获取库存
- [ ] 1.4 计算所有BOM库存的最小值
- [ ] 1.5 返回最小值作为商品包库存（float64 类型）

---

### Requirement 2: 商品包库存查询应用服务方法

**用户故事**: 作为 系统开发者，我想 使用带缓存的商品包库存查询接口，以便于 提升查询性能

#### 验收标准

1. **WHEN** 查询商品包的库存 **AND** 缓存中存在该商品包的库存数据 **THEN** 系统 **SHALL** 直接返回缓存数据
2. **WHEN** 查询商品包的库存 **AND** 缓存中不存在该商品包的库存数据 **THEN** 系统 **SHALL** 计算库存并写入缓存
3. **WHEN** 缓存过期（5分钟） **THEN** 系统 **SHALL** 重新计算库存并更新缓存
4. **WHEN** 调用缓存失效方法 **THEN** 系统 **SHALL** 删除对应商品包的缓存数据

#### 具体要求

- [ ] 2.1 在 `ProductInventoryAppService` 中新增 `GetProductPackageInventory(productPackageUuid)` 方法
- [ ] 2.2 实现缓存逻辑：
  - [ ] 2.2.1 缓存键格式：`product_package_inventory:{company_uuid}:{product_package_uuid}`
  - [ ] 2.2.2 缓存过期时间：5分钟（与BOM库存缓存保持一致）
  - [ ] 2.2.3 优先从缓存读取，缓存未命中时调用领域服务计算
- [ ] 2.3 新增 `InvalidateProductPackageInventoryCache(productPackageUuid)` 方法用于缓存失效
- [ ] 2.4 在BOM库存更新时，同步失效对应商品包的缓存

---

### Requirement 3: 商品包库存计算逻辑

**用户故事**: 作为 仓库管理员，我想 查询商品包的实时库存，以便于 准确了解商品包可售数量

#### 验收标准

1. **WHEN** 查询商品包的库存 **AND** 商品包下存在多个商品BOM **THEN** 系统 **SHALL** 返回所有BOM库存中的最小值
2. **WHEN** 查询商品包的库存 **AND** 商品包下只有一个商品BOM **THEN** 系统 **SHALL** 返回该BOM的库存
3. **WHEN** 查询商品包的库存 **AND** 商品包下没有商品BOM **THEN** 系统 **SHALL** 返回0或抛出明确的错误信息
4. **WHEN** 查询商品包的库存 **AND** 某个BOM库存查询失败 **THEN** 系统 **SHALL** 记录错误日志但继续计算其他BOM的库存

#### 具体要求

- [ ] 3.1 实现最小值计算逻辑：
  - [ ] 3.1.1 遍历商品包下所有BOM
  - [ ] 3.1.2 对每个BOM调用 `GetProductInventory` 获取库存
  - [ ] 3.1.3 使用 `math.Min` 或循环比较获取最小值
- [ ] 3.2 处理边界情况：
  - [ ] 3.2.1 商品包下没有BOM：返回0或抛出错误（需明确业务规则）
  - [ ] 3.2.2 某个BOM查询失败：记录错误日志，继续计算其他BOM
  - [ ] 3.2.3 所有BOM查询失败：返回错误
- [ ] 3.3 确保计算逻辑的准确性：
  - [ ] 3.3.1 正确处理无限库存（99999999）的情况
  - [ ] 3.3.2 确保最小值计算的正确性

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 DDD 分层架构（Domain → Application → Infrastructure）
- **领域服务扩展**: 在 `main/app/modules/inventory/domain/service/` 中扩展 `ProductInventoryDomainService`
- **应用服务扩展**: 在 `main/app/modules/inventory/application/` 中扩展 `ProductInventoryAppService`
- **代码复用**: 复用现有的 `GetProductInventory` 方法和策略模式
- **单一职责原则**: 商品包库存计算逻辑独立封装，不与其他逻辑耦合
- **依赖管理**: Service 只能依赖 Repository 接口，不能直接依赖 Infrastructure
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/go-modules.mdc` - Go Modules 模块开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] 如果后续需要提供 HTTP API，URL 使用 snake_case 命名（如：`/api/v1/product_package_inventory`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{inventory: float64}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 不涉及新增数据库表，使用现有表结构
- [ ] 使用现有字段和关联关系：
  - `ttpos_product_package.uuid` - 商品包UUID
  - `ttpos_product_bom.product_package_uuid` - 商品BOM关联的商品包UUID
  - `ttpos_product_bom.uuid` - 商品BOM UUID
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化：
  - [ ] 使用 `FindByProductPackageUuid` 批量查询BOM，减少数据库查询次数
  - [ ] 使用索引 `product_package_uuid` 优化查询性能
  - [ ] 避免 N+1 查询问题
- [ ] 缓存策略（Redis）：
  - [ ] 对商品包库存计算结果进行缓存，减少重复计算
  - [ ] 缓存过期时间：5分钟
  - [ ] 缓存失效：在BOM库存更新时，同步失效对应商品包的缓存
- [ ] 并发处理：使用 UUID 锁确保并发安全

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] **商品包库存计算逻辑测试覆盖率 100%**（高风险业务逻辑）
- [ ] 集成测试覆盖核心流程：
  - [ ] 测试商品包下多个BOM的库存计算
  - [ ] 测试商品包下单个BOM的库存计算
  - [ ] 测试商品包下没有BOM的边界情况
  - [ ] 测试缓存机制
- [ ] 单元测试覆盖所有新增方法
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有错误提示使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 输入参数校验（商品包UUID必须为正整数）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 错误日志记录（使用 Logger）
- [ ] 部分BOM查询失败时，不影响其他BOM的计算
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **商品包库存计算**: 正确计算商品包下所有BOM库存的最小值
2. **边界情况处理**: 正确处理商品包下没有BOM、部分BOM查询失败等边界情况
3. **缓存机制**: 缓存正常工作，缓存失效机制正确
4. **统一接口**: 提供统一的商品包库存查询接口，复用现有BOM库存查询逻辑

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%，Repository ≥ 80%，计算逻辑 100%）
2. **集成测试**: 端到端流程测试通过
3. **性能测试**: 响应时间 < 200ms，支持并发查询
4. **缓存测试**: 缓存读写和失效机制测试通过

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
- 复用现有的 `ProductInventoryDomainService` 和 `ProductInventoryAppService`

### 业务约束

- 商品包库存计算规则：取所有BOM库存的最小值
- 商品包下没有BOM时的处理规则：返回0或抛出错误（需明确业务规则）
- 不能影响现有BOM库存查询功能
- 需要保持向后兼容

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `main/app/modules/inventory/domain/service/product_inventory_domain_service.go` - 商品库存领域服务（已实现）
- `main/app/modules/inventory/application/product_inventory_app_service.go` - 商品库存应用服务（已实现）
- `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl.go` - 商品BOM仓储实现（已实现）
- `main/app/model/product.go` - ProductPackage 和 ProductBom 模型

### 服务依赖

- **Main → Main**: 内部服务调用（无需跨服务）

### 业务依赖

- **前置条件**: 商品BOM库存查询功能已实现（story-warehouse-product-inventory-query）
- **数据依赖**: 
  - 商品包数据（ProductPackage）
  - 商品BOM数据（ProductBom）
  - 商品BOM与商品包的关联关系（product_package_uuid）

---

## 风险和缓解

### 风险 1: 性能问题

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 使用 `FindByProductPackageUuid` 批量查询BOM，减少数据库查询次数
- 对商品包库存进行缓存，减少重复计算
- 优化数据库查询，使用索引优化查询性能

### 风险 2: 数据一致性

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 商品包库存依赖多个BOM的库存数据，需要确保数据一致性
- 在BOM库存更新时，同步失效对应商品包的缓存
- 定期校验库存数据的准确性

### 风险 3: 边界情况处理

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 明确商品包下没有BOM时的处理规则
- 部分BOM查询失败时，记录错误日志但不中断整个计算流程
- 编写详细的单元测试覆盖各种边界情况

### 风险 4: 缓存失效策略

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在BOM库存更新时，同步失效对应商品包的缓存
- 实现缓存失效方法，确保缓存更新的及时性
- 设置合理的缓存过期时间（5分钟）

---

## 时间表

- **Phase 1 - 领域服务扩展**: 0.5 天
  - 新增 `GetProductPackageInventory` 方法
  - 实现最小值计算逻辑
- **Phase 2 - 应用服务扩展**: 0.5 天
  - 新增带缓存的商品包库存查询方法
  - 实现缓存失效方法
- **Phase 3 - 单元测试和集成测试**: 1 天
  - 测试商品包库存计算逻辑
  - 测试缓存机制
  - 测试边界情况
- **Phase 4 - 文档和代码审查**: 0.5 天
- **总计**: 2-3 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/go-modules.mdc` - Go Modules 模块开发规范
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

- `main/app/modules/inventory/domain/service/product_inventory_domain_service.go` - 商品库存领域服务
- `main/app/modules/inventory/application/product_inventory_app_service.go` - 商品库存应用服务
- `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl.go` - 商品BOM仓储实现
- `main/app/model/product.go` - ProductPackage 和 ProductBom 模型

### 相关 Spec

- `docs/shared/specs/active/story-warehouse-product-inventory-query/` - 商品BOM库存查询功能（前置依赖）

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
**创建日期**: 2025-12-11  
**作者**: xiezhihuan  
**审核者**: {审核者}

