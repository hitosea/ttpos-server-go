# ERP销售订单dripShop交付逻辑 需求文档

> 本文档定义ERP销售订单dripShop交付逻辑的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-26-erp-saleorder-drip-shop-delivery.md](../../../team/proposals/2025-11-26-erp-saleorder-drip-shop-delivery.md) |
| **创建日期**      | 2025-11-26                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint 25                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

---

## 📋 概述

ERP销售订单dripShop交付逻辑优化功能旨在提升ttpos-bmp中台ERP模块内部销售订单创建效率。通过在 `CreateInnerSaleOrderFromPurchaseOrder` 方法中增加对item的处理，自动识别dripShop类型的商品，并自动设置供应商交付相关信息，减少手工操作，降低错误率。

该功能主要涉及ttpos-erp微服务的buying模块逻辑优化，**不涉及数据库表修改和创建**，无需UI界面变更，纯后端业务逻辑调整。

## 🎯 产品对齐

该功能支持公司2025年Q4的核心目标：
- **提升运营效率**: 减少ERP操作的重复手工输入
- **降低错误率**: 通过自动化逻辑减少人为失误
- **优化用户体验**: 提升ERP系统的智能化程度

## 📝 用户故事

**作为** ERP系统操作员  
**我想** 在ttpos-bmp中台ERP模块通过 `CreateInnerSaleOrderFromPurchaseOrder` 创建内部销售订单时，对于dripShop商品能够自动设置供应商交付相关信息  
**以便于** 减少手工操作时间并确保设置准确性

---

## 功能需求

### Requirement 1: dripShop商品自动识别

**用户故事**: 作为ttpos-erp微服务buying模块，我想在 `CreateInnerSaleOrderFromPurchaseOrder` 创建内部销售订单时自动识别dripShop类型的商品，以便于应用特殊的交付逻辑

#### 验收标准

1. **WHEN** 通过 `CreateInnerSaleOrderFromPurchaseOrder` 创建内部销售订单时包含dripShop=true的商品 **THEN** 系统自动识别该商品为dripShop类型
2. **WHEN** 商品的dripShop属性为false或不存在 **THEN** 系统不应用特殊交付逻辑
3. **WHEN** 订单中同时包含dripShop和非dripShop商品 **THEN** 系统仅对dripShop商品应用特殊逻辑

#### 具体要求

- [ ] 1.1 在 `CreateInnerSaleOrderFromPurchaseOrder` 方法中检查商品的dropship属性
- [ ] 1.2 通过 `service.Item().GetItem()` 获取商品信息，检查 **`Item.DeliveredBySupplier`** 字段（`delivered_by_supplier`，int类型，1=true）
- [ ] 1.3 记录dropship商品的识别日志以便调试

---

### Requirement 2: 自动设置供应商交付标识

**用户故事**: 作为ttpos-erp微服务buying模块，我想在识别dripShop商品后自动设置供应商交付标识，以便于简化订单创建流程

#### 验收标准

1. **WHEN** 内部销售订单中存在dripShop商品 **THEN** 系统自动将这些商品的 `DeliveredBySupplier` 设置为 true
2. **WHEN** dripShop商品的供应商交付标识已被手动设置 **THEN** 系统不覆盖现有设置（在设置DeliveryDate之后处理）
3. **WHEN** 非dripShop商品的供应商交付标识 **THEN** 系统保持原有逻辑不变

#### 具体要求

- [ ] 2.1 在 `processDripShopItems` 方法中实现供应商交付标识的自动设置逻辑
- [ ] 2.2 设置 `SaleOrderItem.DeliveredBySupplier = true`（bool类型）
- [ ] 2.3 支持批量商品的自动设置

---

### Requirement 3: 自动选择供应商

**用户故事**: 作为ttpos-erp微服务buying模块，我想在设置供应商交付标识时自动选择供应商，以便于完成订单创建

#### 验收标准

1. **WHEN** dripShop商品设置为供应商交付 **THEN** 系统自动从商品的 `SupplierItems` 列表中选择第一个供应商
2. **WHEN** 商品没有供应商列表或列表为空 **THEN** 系统记录错误并终止订单创建
3. **WHEN** 选择的供应商无效或不存在 **THEN** 系统记录错误并终止订单创建

#### 具体要求

- [ ] 3.1 实现从商品 `SupplierItems` 字段中解析并选择第一个供应商的逻辑
- [ ] 3.2 使用 gjson 解析 SupplierItems（[]interface{}类型）
- [ ] 3.3 处理供应商数据异常的情况（空列表、格式错误等）
- [ ] 3.4 确认SaleOrderItem是否需要Supplier字段，或通过其他方式处理供应商信息

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 销售订单创建逻辑专注订单创建，dripShop逻辑单独处理
- **模块化设计**: dripShop逻辑可独立测试和复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
  - `.cursor/rules/go-rules.mdc` - BMP Go 代码开发规范
  - `.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### API 设计要求

- [x] 使用 gRPC 接口设计
- [x] data 字段必须是对象：`{"data": {"order_uuid": 123, "status": "created"}}`
- [x] 响应格式：统一gRPC响应格式
- [x] 参考: `.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### 数据库设计要求

- [x] **不涉及数据库表修改和创建**，所有字段均已存在于ERP系统中
- [x] 使用现有的 `tabItem` 表中的 **`delivered_by_supplier`** 字段（dropship属性）
- [x] 使用现有的 `tabSales Order Item` 表中的 `delivered_by_supplier` 字段
- [x] 使用现有的 `tabItem` 表中的 `supplier_items` 字段
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] dripShop逻辑处理时间 < 50ms
- [x] 不影响销售订单创建的整体性能
- [x] 支持并发订单创建场景

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] dripShop逻辑相关测试覆盖率 100%（高风险）
- [x] 集成测试覆盖完整的订单创建流程
- [x] 异常场景测试（供应商不存在、无效商品等）
- [x] 参考: `.cursor/rules/go-bmp.mdc` - 测试规范

### 安全要求

- [x] gRPC接口需要身份验证
- [x] 商品和供应商数据访问需要权限验证
- [x] 防止恶意修改dripShop属性的数据篡改
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] dripShop逻辑失败时事务回滚，保证数据一致性
- [x] 详细记录错误日志（使用GoFrame日志）
- [x] 供应商数据异常时的降级处理

---

## 验收标准

### 功能验收

1. **dripShop识别**: 在 `CreateInnerSaleOrderFromPurchaseOrder` 中能正确识别dripShop商品并应用特殊逻辑
2. **自动设置**: dripShop商品的 `DeliveredBySupplier` 字段自动设置为 true
3. **供应商选择**: 自动从商品 `SupplierItems` 列表中选择第一个供应商（如需要）
4. **兼容性**: 不影响非dripShop商品的正常内部销售订单创建流程
5. **不涉及表修改**: 确认没有数据库表结构变更

### 测试验收

1. **单元测试**: dripShop逻辑相关代码覆盖率100%（processDripShopItems、isDripShopItem、selectFirstSupplier）
2. **gRPC测试**: CreateInnerSaleOrderFromPurchaseOrder接口在各种场景下正确响应
3. **集成测试**: 完整的内部销售订单创建流程测试通过（从采购订单创建）
4. **异常测试**: 供应商异常、无效商品、SupplierItems为空等场景处理正确

### 文档验收

1. **技术文档**: design.md 包含完整的实现设计
2. **API文档**: gRPC接口文档已更新（如有变更）
3. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame v2.x 框架
- 接口以 `I` 开头，实现以 `s` 开头
- Service 只能依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 使用 gerror 处理错误

### 业务约束

- 仅在 `CreateInnerSaleOrderFromPurchaseOrder` 创建内部销售订单时应用dripShop逻辑
- 不影响其他销售订单创建方法
- dripShop属性的判断基于商品数据，不支持动态修改
- 供应商选择固定为列表中的第一个，不支持自定义选择
- **不涉及数据库表修改和创建**

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 (符合 ≤ 5 标准)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 核心框架
- `google.golang.org/grpc` - gRPC 框架

### 服务依赖

- **无外部微服务依赖**: 纯 ttpos-erp 模块实现

### 业务依赖

- 依赖现有的buying模块（sBuying）
- 依赖现有的商品模块（service.Item()）
- 依赖现有的ERP文档服务（service.Document()）

---

## 风险和缓解

### 风险 1: dripShop/dropship字段定义（已确认）

**影响**: 低  
**概率**: 低  
**状态**: ✅ 已确认

- 字段名称：**`Item.DeliveredBySupplier`**（JSON: `delivered_by_supplier`）
- 数据类型：int（1 表示 true，商品由供应商直接交付）
- 位置：`ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go`

### 风险 2: SupplierItems数据结构不明确

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 查看现有代码中如何使用SupplierItems字段
- 确认SupplierItems的数据结构格式（JSON格式或特定结构）
- 使用gjson解析供应商信息

### 风险 3: SaleOrderItem缺少Supplier字段

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 确认需求中"supplier从item的supplier中获取第一个"的具体含义
- 如果SaleOrderItem需要Supplier字段，需要确认是否可以通过其他方式处理
- 与产品团队确认供应商信息的处理方式

---

## 时间表

- **Phase 1 - 需求分析和设计**: 0.5 天
- **Phase 2 - 代码实现和测试**: 1.5 天
- **Phase 3 - 集成测试和优化**: 1 天
- **总计**: 3 天（SP = 3-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 核心约束
- `.cursor/rules/go-rules.mdc` - BMP Go 代码开发规范
- `.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/modules.md` - 模块关系说明

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/protobuf-development-guide.md` - Protobuf 开发指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-26.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 后端开发组  
**审核者**: CTO