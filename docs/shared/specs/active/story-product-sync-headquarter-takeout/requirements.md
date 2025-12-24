# 同步总部外卖商品功能 需求文档

> 本文档定义同步总部外卖商品到子店的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/sync-headquarter-takeout-product.md](../../../../team/proposals/2025-12/sync-headquarter-takeout-product.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | 曾振华                                                                                                       |
| **目标 Sprint**   | Sprint 当前                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #37953                                                                                              |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | -             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

在商品同步功能中，当前只同步了店内商品数据，缺少外卖商品数据的同步。本需求旨在扩展 `SyncProduct` 方法，在同步完总部店内商品后，增加同步总部外卖商品的逻辑，包括外卖商品基本信息、外卖规格价格、外卖属性价格、外卖套餐子商品价格四个维度的数据。

同步策略：
- 首次同步：外卖商品状态默认设置为下架（0），规格价格和套餐子商品价格使用总部数据
- 再次同步：保留子店已配置的外卖商品状态（status）、规格价格（price）和套餐子商品价格（add_price），避免覆盖子店的个性化配置

## 🎯 产品对齐

该功能支持连锁商户的外卖商品统一管理，使总部可以批量下发外卖商品配置到子店，同时保留子店的定价和上架自主权，提升外卖业务的管理效率。

## 📝 用户故事

**作为** 总部管理员  
**我想** 将外卖商品配置同步到所有子店  
**以便于** 实现外卖商品的统一管理和批量下发

**作为** 子店管理员  
**我想** 在首次同步时商品默认下架  
**以便于** 我可以先检查和调整配置后再上架销售

**作为** 子店管理员  
**我想** 在再次同步时保留我自定义的价格和状态  
**以便于** 总部的更新不会影响我已经配置好的外卖商品

---

## 功能需求

### Requirement 1: 同步总部外卖商品基本信息

**用户故事**: 作为系统，我想同步总部的外卖商品基本信息到子店，以便于子店可以使用总部配置的外卖商品

#### 验收标准

1. **WHEN** 执行商品同步 **AND** 公司是子店 **AND** 开启同步总部数据 **THEN** 系统 **SHALL** 查询总部的所有外卖商品（包含软删除）
2. **WHEN** 查询到总部外卖商品 **THEN** 系统 **SHALL** 包含以下关联数据：
   - 外卖规格价格（`ProductBomTakeouts`）
   - 外卖属性价格（`ProductPackageAttributeTakeouts`）
   - 外卖套餐子商品价格（`ProductPackageGroupItemTakeouts`）
3. **WHEN** 子店首次同步某个外卖商品 **THEN** 系统 **SHALL** 将该外卖商品的 `status` 设置为 0（下架）
4. **WHEN** 子店再次同步某个外卖商品 **AND** 子店已存在该商品 **THEN** 系统 **SHALL** 保留子店原有的 `status` 和 `price` 值

#### 具体要求

- [x] 1.1 查询总部外卖商品时，使用 `WithProductBomTakeouts`、`WithProductPackageAttributeTakeouts` 和 `WithProductPackageGroupItemTakeouts` 预加载关联数据
- [x] 1.2 同步时复制总部外卖商品的所有字段，除了需要特殊处理的字段（status、price、HeadquarterUuid）
- [x] 1.3 设置子店外卖商品的 `HeadquarterUuid` 为总部 UUID
- [x] 1.4 保留总部外卖商品的时间戳（CreateTime、UpdateTime、DeleteTime）
- [x] 1.5 首次同步时 status 设置为 0，再次同步保留子店的 status 和 price

---

### Requirement 2: 同步外卖规格价格

**用户故事**: 作为系统，我想同步总部的外卖规格价格到子店，以便于子店可以使用总部配置的规格定价

#### 验收标准

1. **WHEN** 同步外卖商品 **AND** 总部外卖商品有规格价格 **THEN** 系统 **SHALL** 同步所有规格价格到子店
2. **WHEN** 子店首次同步某个规格价格 **THEN** 系统 **SHALL** 使用总部的 `price` 值
3. **WHEN** 子店再次同步某个规格价格 **AND** 子店已存在该规格价格 **THEN** 系统 **SHALL** 保留子店原有的 `price` 值
4. **WHEN** 规格价格同步失败 **THEN** 系统 **SHALL** 记录错误日志但不中断事务

#### 具体要求

- [x] 2.1 查询子店已有的外卖规格价格，根据 `uuid` 和 `product_package_takeout_uuid` 判断是否已同步
- [x] 2.2 对于已同步的规格价格，读取子店的 `price` 值并在新数据中使用
- [x] 2.3 复制总部规格价格的其他所有字段（ProductBomUuid、GrabModifierId、时间戳等）
- [x] 2.4 设置子店规格价格的 `HeadquarterUuid` 为总部 UUID

---

### Requirement 3: 同步外卖属性价格

**用户故事**: 作为系统，我想同步总部的外卖属性价格到子店，以便于子店可以使用总部配置的属性定价

#### 验收标准

1. **WHEN** 同步外卖商品 **AND** 总部外卖商品有属性价格 **THEN** 系统 **SHALL** 同步所有属性价格到子店
2. **WHEN** 子店首次同步某个属性价格 **THEN** 系统 **SHALL** 使用总部的 `price` 值
3. **WHEN** 子店再次同步某个属性价格 **AND** 子店已存在该属性价格 **THEN** 系统 **SHALL** 保留子店原有的 `price` 值
4. **WHEN** 属性价格同步失败 **THEN** 系统 **SHALL** 记录错误日志但不中断事务

#### 具体要求

- [x] 3.1 查询子店已有的外卖属性价格，根据 `uuid` 判断是否已同步
- [x] 3.2 对于已同步的属性价格，读取子店的 `price` 值并在新数据中使用
- [x] 3.3 复制总部属性价格的其他所有字段（ProductPackageAttributeUuid、时间戳等）
- [x] 3.4 设置子店属性价格的 `HeadquarterUuid` 为总部 UUID

---

### Requirement 4: 同步外卖套餐子商品价格

**用户故事**: 作为系统，我想同步总部的外卖套餐子商品价格到子店，以便于子店可以使用总部配置的套餐子商品定价

#### 验收标准

1. **WHEN** 同步外卖商品 **AND** 总部外卖商品有套餐子商品价格 **THEN** 系统 **SHALL** 同步所有套餐子商品价格到子店
2. **WHEN** 子店首次同步某个套餐子商品价格 **THEN** 系统 **SHALL** 使用总部的 `add_price` 值
3. **WHEN** 子店再次同步某个套餐子商品价格 **AND** 子店已存在该价格记录 **THEN** 系统 **SHALL** 保留子店原有的 `add_price` 值
4. **WHEN** 套餐子商品价格同步失败 **THEN** 系统 **SHALL** 记录错误日志但不中断事务

#### 具体要求

- [x] 4.1 查询子店已有的外卖套餐子商品价格，根据 `uuid`、`product_package_takeout_uuid` 和 `product_package_group_item_uuid` 判断是否已同步
- [x] 4.2 对于已同步的套餐子商品价格，读取子店的 `add_price` 值并在新数据中使用
- [x] 4.3 复制总部套餐子商品价格的其他所有字段（ProductPackageGroupUuid、ProductPackageGroupItemUuid、时间戳等）
- [x] 4.4 设置子店套餐子商品价格的 `HeadquarterUuid` 为总部 UUID

---

### Requirement 5: 批量删除和批量插入

**用户故事**: 作为系统，我想先删除子店现有的外卖商品数据，再批量插入总部数据，以便于保证数据的完整性和一致性

#### 验收标准

1. **WHEN** 开始同步外卖商品 **THEN** 系统 **SHALL** 先查询子店所有来自总部的外卖商品（`HeadquarterUuid = 总部UUID`）
2. **WHEN** 查询到子店现有外卖商品 **THEN** 系统 **SHALL** 批量物理删除这些商品及其关联数据（规格价格、属性价格、套餐子商品价格）
3. **WHEN** 删除完成 **THEN** 系统 **SHALL** 批量插入新的外卖商品数据
4. **WHEN** 批量插入失败 **THEN** 系统 **SHALL** 回滚整个事务

#### 具体要求

- [x] 5.1 使用 `DestroyProductPackageTakeout` 物理删除外卖商品
- [x] 5.2 使用 `DestroyProductBomTakeout` 物理删除外卖规格价格
- [x] 5.3 使用 `DestroyProductPackageAttributeTakeout` 物理删除外卖属性价格
- [x] 5.4 使用 `DestroyProductPackageGroupItemTakeout` 物理删除外卖套餐子商品价格
- [x] 5.5 批量插入时，外卖商品使用批量创建，关联数据逐条创建
- [x] 5.6 使用数据库事务保证数据一致性

---

### Requirement 6: 错误处理和日志

**用户故事**: 作为开发者，我想在同步过程中有完善的错误处理和日志记录，以便于问题排查和数据追溯

#### 验收标准

1. **WHEN** 同步外卖商品失败 **THEN** 系统 **SHALL** 回滚事务并返回错误信息
2. **WHEN** 创建单个关联数据失败 **THEN** 系统 **SHALL** 记录错误日志但继续处理其他数据
3. **WHEN** 事务提交失败 **THEN** 系统 **SHALL** 返回包含详细错误信息的错误对象

#### 具体要求

- [x] 6.1 使用 `errors.WithMessage` 包装错误信息
- [x] 6.2 使用 `logger.Logger.Error` 记录关键错误
- [x] 6.3 错误日志包含必要的上下文信息（UUID、表名等）
- [x] 6.4 整体事务失败时返回明确的错误提示

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 同步逻辑封装在 Service 层
- **模块化设计**: Repository 提供独立的查询和写入方法
- **依赖管理**: Service 只依赖 Repository 接口
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### 数据库设计要求

- [x] 外卖商品表已包含标准字段: `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 价格字段使用 decimal(22,4)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 使用批量操作减少数据库交互次数
- [x] 预加载关联数据避免 N+1 查询
- [x] 使用事务保证数据一致性
- [x] 大批量数据时使用分批处理（如有需要）

### 可靠性要求

- [x] 使用数据库事务保证数据一致性
- [x] 错误日志记录（使用 Logger）
- [x] 关键步骤失败时回滚事务
- [x] 非关键步骤失败时记录日志但继续执行

---

## 验收标准

### 功能验收

1. **总部到子店同步**: 子店执行商品同步后，外卖商品、规格价格、属性价格、套餐子商品价格全部同步成功
2. **首次同步状态**: 首次同步的外卖商品状态为下架（0）
3. **再次同步保留**: 再次同步时，子店的外卖商品 price/status、规格价格、属性价格和套餐子商品价格保持不变
4. **数据完整性**: 同步后的数据包含所有必要字段，关联关系正确

### 测试验收

1. **单元测试**: Service 层方法的测试覆盖率 ≥ 70%
2. **集成测试**: 端到端同步流程测试通过
3. **手动测试**: 
   - 首次同步：验证外卖商品状态为下架
   - 再次同步：验证子店的外卖商品 price/status、规格 price、属性 price 和套餐子商品 add_price 保持不变
   - 错误场景：验证事务回滚和错误日志

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **任务文档**: tasks.md 任务分解完成
3. **代码注释**: 关键逻辑有中文注释

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架（当前已使用）
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例
- 不使用 panic，返回 error
- 遵循 `.cursor/rules/go-main.mdc`

### 业务约束

- 只同步 `HeadquarterUuid = 0` 的总部外卖商品
- 子店的外卖商品必须标记 `HeadquarterUuid = 总部UUID`
- 首次同步时外卖商品必须下架
- 再次同步时必须保留子店的外卖商品 price/status、规格 price、属性 price 和套餐子商品 add_price

### 资源约束

- 开发时间: 2 天
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `gorm.io/gorm` - ORM 框架
- `github.com/pkg/errors` - 错误处理
- `go.uber.org/zap` - 日志记录

### 内部依赖

- `repository.ProductPackageTakeoutRepo` - 外卖商品 Repository
- `repository.ProductBomTakeoutRepo` - 外卖规格价格 Repository
- `repository.ProductPackageAttributeTakeoutRepo` - 外卖属性价格 Repository
- `repository.ProductPackageGroupItemTakeoutRepo` - 外卖套餐子商品价格 Repository
- `repository.CommonRepo` - 通用查询条件

### 业务依赖

- 必须先同步店内商品（`ProductPackage`、`ProductBom`）
- 外卖商品必须关联已存在的店内商品（`product_package_uuid`）
- 公司必须是子店（`IsSubShop() = true`）

---

## 风险和缓解

### 风险 1: 子店配置被意外覆盖

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 首次同步前查询子店是否已存在该外卖商品
- 明确定义保留字段列表（外卖商品 price/status、规格 price、属性 price、套餐子商品 add_price）
- 增加单元测试覆盖保留逻辑

### 风险 2: 大数据量同步性能问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用批量操作减少数据库交互
- 预加载关联数据避免 N+1 查询
- 如需要可实现分批同步

### 风险 3: 事务超时导致同步失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 合理设置事务超时时间
- 关联数据创建失败时记录日志但不中断
- 提供手动重试机制

---

## 时间表

- **Phase 1 - 代码实现**: 1 天
  - 修改 `SyncProduct` 方法
  - 实现外卖商品同步逻辑
  - 实现保留字段逻辑
- **Phase 2 - 测试验证**: 0.5 天
  - 编写单元测试
  - 手动测试验证
- **Phase 3 - 文档完善**: 0.5 天
  - 编写 design.md
  - 编写 tasks.md
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 相关代码

- `main/app/service/product.go:7754-8089` - 店内商品同步逻辑（参考实现）
- `main/app/model/product_package_takeout.go` - 外卖商品模型
- `main/app/model/product_bom_takeout.go` - 外卖规格价格模型
- `main/app/model/product_package_attribute_takeout.go` - 外卖属性价格模型
- `main/app/model/product_package_group_item_takeout.go` - 外卖套餐子商品价格模型
- `main/app/repository/product_package_takeout.go` - 外卖商品 Repository
- `main/app/repository/product_bom_takeout.go` - 外卖规格价格 Repository
- `main/app/repository/product_package_attribute_takeout.go` - 外卖属性价格 Repository
- `main/app/repository/product_package_group_item_takeout.go` - 外卖套餐子商品价格 Repository

### 外部参考

- DooTask #37953 - 原始需求

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/曾振华/2025-12/2025-12-18.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: 曾振华  
**审核者**: 待审核
