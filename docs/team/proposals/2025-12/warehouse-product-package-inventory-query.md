# 仓库模块商品包库存查询功能 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-11   |
| **目标版本** | {版本号} |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-warehouse-product-package-inventory-query](../../../shared/specs/archived/v2.12/story-warehouse-product-package-inventory-query/requirements.md)      |
| **关联提案** | [warehouse-product-inventory-query](./warehouse-product-inventory-query.md) |

---

## 🎯 背景和动机

### 问题描述

当前仓库模块已实现了商品BOM（ProductBom）的库存查询功能，但缺少商品包（ProductPackage）的库存查询接口。在实际业务场景中，一个商品包对应多个商品BOM，需要根据商品包下所有BOM的库存情况来计算商品包的整体库存。

**现状问题**：
- 商品包库存查询功能缺失，无法直接查询商品包的库存
- 商品包与商品BOM是一对多关系，需要遍历所有BOM才能计算商品包库存
- 商品包库存计算规则不明确：应该取最小值还是其他计算方式
- 缺少统一的商品包库存查询领域服务接口

### 业务价值

- **统一库存管理**：提供商品包级别的库存查询接口，完善库存管理体系
- **业务准确性**：明确商品包库存计算规则（取最小值），确保库存计算的准确性
- **代码复用**：基于已有的商品BOM库存查询功能，复用现有策略和逻辑
- **性能优化**：通过批量查询和缓存机制，提升商品包库存查询性能

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 仓库管理员、系统开发者

---

## 💡 解决方案概述

### 方案描述

在现有的商品BOM库存查询功能基础上，扩展商品包库存查询功能。商品包库存等于该商品包下所有商品BOM库存中的最小值。

**核心设计思路**：
1. **扩展领域服务**：在 `ProductInventoryDomainService` 中新增 `GetProductPackageInventory` 方法
2. **复用现有逻辑**：通过 `FindByProductPackageUuid` 查询商品包下所有BOM，然后调用现有的 `GetProductInventory` 方法获取每个BOM的库存
3. **最小值计算**：遍历所有BOM的库存，返回最小值作为商品包库存
4. **缓存优化**：在应用服务层添加商品包库存缓存，提升查询性能

### 核心功能点

1. **商品包库存查询领域服务方法**
   - 提供 `GetProductPackageInventory(productPackageUuid)` 方法
   - 查询商品包下所有商品BOM
   - 遍历计算每个BOM的库存
   - 返回所有BOM库存中的最小值

2. **商品包库存查询应用服务方法**
   - 提供带缓存的商品包库存查询接口
   - 缓存键格式：`product_package_inventory:{company_uuid}:{product_package_uuid}`
   - 缓存过期时间：5分钟（与BOM库存缓存保持一致）
   - 提供缓存失效方法

3. **库存计算规则**
   - **商品包库存** = min(商品包下所有商品BOM的库存)
   - 如果商品包下没有BOM，返回0或抛出错误
   - 如果某个BOM查询失败，记录错误但继续计算其他BOM

4. **错误处理**
   - 商品包不存在：返回错误
   - 商品包下没有BOM：返回0或抛出错误（需明确业务规则）
   - BOM库存查询失败：记录错误日志，但不中断整个计算流程

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [x] 其他: 仓库模块领域服务、应用服务

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 基于已有的商品BOM库存查询功能，复用现有策略和逻辑
- 需要扩展领域服务接口，新增商品包库存计算方法
- 需要实现批量查询和最小值计算逻辑
- 需要添加缓存机制，提升性能
- 需要处理边界情况（无BOM、查询失败等）

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3 SP（待技术评审确认）

**分解**：
- 领域服务扩展：0.5 天
  - 新增 `GetProductPackageInventory` 方法
  - 实现最小值计算逻辑
- 应用服务扩展：0.5 天
  - 新增带缓存的商品包库存查询方法
  - 实现缓存失效方法
- 单元测试和集成测试：1 天
  - 测试商品包库存计算逻辑
  - 测试缓存机制
  - 测试边界情况
- 文档和代码审查：0.5 天

### 风险识别

**潜在风险**：
1. **性能问题**：商品包下BOM数量较多时，需要多次调用BOM库存查询，可能影响性能
2. **数据一致性**：商品包库存依赖多个BOM的库存数据，需要确保数据一致性
3. **边界情况处理**：商品包下没有BOM、部分BOM查询失败等边界情况的处理规则需要明确
4. **缓存失效策略**：商品包库存缓存需要在BOM库存更新时同步失效

**缓解措施**：
1. **批量查询优化**：使用 `FindByProductPackageUuid` 批量查询BOM，减少数据库查询次数
2. **缓存策略**：对商品包库存进行缓存，减少重复计算
3. **错误处理**：明确边界情况的处理规则，记录错误日志但不中断流程
4. **缓存失效**：在BOM库存更新时，同步失效对应商品包的库存缓存

---

## 🔗 相关资源

### 参考需求

- 基础功能: [warehouse-product-inventory-query](./warehouse-product-inventory-query.md) - 商品BOM库存查询功能
- 类似功能: 商品库存管理相关功能

### 相关文档

- 产品需求文档 (PRD): 待补充
- 用户调研报告: 待补充
- 技术预研文档: 待补充
- 相关代码: 
  - `main/app/modules/inventory/domain/service/product_inventory_domain_service.go` - 商品库存领域服务
  - `main/app/modules/inventory/application/product_inventory_app_service.go` - 商品库存应用服务
  - `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl.go` - 商品BOM仓储实现
  - `main/app/model/product.go` - ProductPackage 和 ProductBom 模型定义

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-warehouse-product-package-inventory-query`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 仓库管理员  
**我想** 查询商品包的实时库存  
**以便于** 准确了解商品包可售数量，避免超卖或库存积压

**作为** 系统开发者  
**我想** 使用统一的商品包库存查询接口  
**以便于** 简化业务代码，提高代码可维护性

### AC 验收标准（初稿）

1. **WHEN** 查询商品包的库存 **AND** 商品包下存在多个商品BOM **THEN** 系统 **SHALL** 返回所有BOM库存中的最小值
2. **WHEN** 查询商品包的库存 **AND** 商品包下只有一个商品BOM **THEN** 系统 **SHALL** 返回该BOM的库存
3. **WHEN** 查询商品包的库存 **AND** 商品包下没有商品BOM **THEN** 系统 **SHALL** 返回0或抛出明确的错误信息
4. **WHEN** 查询商品包的库存 **AND** 商品包不存在 **THEN** 系统 **SHALL** 返回错误信息
5. **WHEN** 查询商品包的库存 **AND** 某个BOM库存查询失败 **THEN** 系统 **SHALL** 记录错误日志但继续计算其他BOM的库存
6. **WHEN** 查询商品包的库存 **AND** 缓存中存在该商品包的库存数据 **THEN** 系统 **SHALL** 直接返回缓存数据
7. **WHEN** 查询商品包的库存 **AND** 缓存中不存在该商品包的库存数据 **THEN** 系统 **SHALL** 计算库存并写入缓存

### 技术实现要点

1. **领域服务扩展**
   ```go
   // IProductInventoryDomainService 接口新增方法
   GetProductPackageInventory(ctx context.Context, productPackageUuid uint64) (float64, error)
   ```

2. **应用服务扩展**
   ```go
   // ProductInventoryAppService 新增方法
   GetProductPackageInventory(ctx context.Context, productPackageUuid uint64) (float64, error)
   InvalidateProductPackageInventoryCache(ctx context.Context, productPackageUuid uint64) error
   ```

3. **库存计算逻辑**
   - 通过 `FindByProductPackageUuid` 查询商品包下所有BOM
   - 遍历每个BOM，调用 `GetProductInventory` 获取库存
   - 使用 `math.Min` 或循环比较获取最小值
   - 返回最小值作为商品包库存

4. **缓存策略**
   - 缓存键：`product_package_inventory:{company_uuid}:{product_package_uuid}`
   - 缓存过期时间：5分钟
   - 缓存失效：在BOM库存更新时，同步失效对应商品包的缓存

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

