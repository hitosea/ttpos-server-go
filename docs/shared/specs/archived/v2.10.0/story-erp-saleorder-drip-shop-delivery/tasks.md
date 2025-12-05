# ERP销售订单dripShop交付逻辑 任务分解

> 本文档定义ERP销售订单dripShop交付逻辑的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11
**已完成**: 8
**进行中**: -
**完成率**: 73%

---

## Phase 1: 数据结构确认

### 数据结构分析

- [x] 1.1 确认dripShop/dropship字段定义（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go`
  - Purpose: 确认dripShop字段在ERP系统中的存储位置和命名
  - Requirements: 1.1, 1.2, 1.3
  - **Result**: 字段为 **`Item.DeliveredBySupplier`**（`delivered_by_supplier`，int类型，1=true）
  - Success: ✅ 已确认字段位置和字段名

- [x] 1.2 确认供应商数据结构（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go`
  - Purpose: 确认商品SupplierItems字段的数据结构格式
  - Requirements: 3.1, 3.2, 3.3
  - **Result**: `SupplierItems` 是 `[]interface{}` 类型，每个元素包含 `supplier` 字段
  - Success: ✅ 使用 gjson 解析第一个供应商

- [x] 1.3 确认SaleOrderItem是否需要Supplier字段（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/buying.go`
  - Purpose: 根据需求确认SaleOrderItem是否需要添加Supplier字段
  - Requirements: 3.1
  - **Result**: SaleOrderItem 中没有 Supplier 字段，只有 `DeliveredBySupplier bool` 字段
  - Success: ✅ 无需添加 Supplier 字段，只需设置 DeliveredBySupplier = true

---

## Phase 2: 核心业务逻辑实现

### Buying 模块集成 dripShop 逻辑

- [x] 2.1 添加processDripShopItems辅助方法（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 实现dripShop商品识别和供应商选择的业务逻辑
  - Requirements: 1.1-3.3
  - **Result**: 已实现，遍历商品调用 service.Item().GetItem()，检查 DeliveredBySupplier 属性
  - Success: ✅ 业务逻辑正确，异常处理完善

- [x] 2.2 添加isDripShopItem辅助方法（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 判断商品是否为dropship商品
  - Requirements: 1.1, 1.2
  - **Result**: 已实现，检查 Item.DeliveredBySupplier == 1
  - Success: ✅ 正确识别dropship商品

- [x] 2.3 添加selectFirstSupplier辅助方法（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 从商品中选择第一个供应商
  - Requirements: 3.1, 3.2, 3.3
  - **Result**: 已实现，使用 gjson 解析 SupplierItems[0].supplier
  - Success: ✅ 正确选择供应商

- [x] 2.4 修改CreateInnerSaleOrderFromPurchaseOrder集成dripShop逻辑（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 在内部销售订单创建流程中调用dripShop处理逻辑
  - Requirements: 1.1-3.3
  - **Result**: 已实现，在设置DeliveryDate之后调用 processDripShopItems
  - Success: ✅ 内部销售订单创建流程正确集成dripShop逻辑

- [x] 2.5 添加必要的导入（已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go`
  - Purpose: 添加item包和gjson包的导入
  - Requirements: 2.1
  - **Result**: 已添加 `"ttpos-bmp/app/ttpos-erp/api/item"` 和 `"github.com/gogf/gf/v2/encoding/gjson"`
  - Success: ✅ 导入正确，代码可编译

- [ ] 2.6 编写单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go`
  - Purpose: 测试dripShop业务逻辑的正确性
  - Requirements: 1.1-3.3
  - Leverage: 现有测试框架和Mock
  - Prompt: Role: QA Engineer | Task: 为processDripShopItems相关方法编写单元测试，覆盖率100% | Context: 测试正常流程、异常场景（无供应商、无效供应商等） | Restrictions: 使用GoFrame测试框架，Mock依赖的Service | Success: 测试覆盖率100%，所有测试通过

- [x] 2.7 更新gRPC Controller（无需修改，已完成）

  - File: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - Purpose: 确保gRPC控制器正确调用Buying服务
  - Requirements: 所有功能需求
  - **Result**: 无需修改，gRPC控制器已正确调用 service.Buying().CreateInnerSaleOrderFromPurchaseOrder
  - Success: ✅ gRPC接口正确调用业务逻辑

---

## Phase 3: 测试和优化

### 测试完善

- [ ] 3.1 端到端集成测试

  - File: `ttpos-bmp/test/integration/drip_shop_inner_sale_order_test.go`
  - Purpose: 测试完整的内部销售订单创建流程，包括dripShop逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端测试 | Context: 从采购订单创建包含dripShop和非dripShop商品的内部销售订单，验证订单创建成功，验证dripShop商品的DeliveredBySupplier设置正确 | Success: 集成测试通过

- [ ] 3.2 异常场景测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying_test.go`
  - Purpose: 测试各种异常场景的处理
  - Requirements: 所有功能需求
  - Test: dripShop字段不存在、供应商不存在、SupplierItems为空等场景
  - Success: 异常处理正确，错误日志完整

- [ ] 3.3 性能测试

  - File: -
  - Purpose: 验证dripShop逻辑的性能表现
  - Requirements: 非功能需求
  - Test: 批量创建内部销售订单测试，验证响应时间 < 200ms
  - Success: 性能指标达标

### 文档和部署

- [ ] 3.4 更新API文档（如需要）

  - File: `docs/shared/api/erp-api.md`
  - Purpose: 更新ERP API文档（如接口有变更）
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API文档已更新（通常无需更新，因为不涉及接口变更）

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `gofmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - DripShopLogic: 100%
  - SaleOrder: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] dripShop逻辑正确识别和处理商品
- [ ] 供应商交付设置自动完成
- [ ] 错误场景处理完善

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 执行流程

1. **选择任务**: 从 Phase 1 开始，按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的实现代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `gofmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### 预计时间

- Phase 1: 0.5 天（4 小时）
- Phase 2: 1.5 天（12 小时）
- Phase 3: 1 天（8 小时）
- **总计**: 3 天（24 小时）= **SP 3-5**

---

## 附录：AI Prompt 示例

### processDripShopItems 实现

```
Role: Go Developer with GoFrame expertise

Task: 在sBuying结构体中添加processDripShopItems辅助方法

Context:
- File: ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go
- Leverage: design.md 中的完整实现代码
- Requirements: requirements.md Requirement 1,2,3
- 现有代码: sBuying结构体已有CreateInnerSaleOrderFromPurchaseOrder方法
- 导入: 需要添加 "ttpos-bmp/app/ttpos-erp/api/item" 导入

Implementation Steps:
1. 在buying.go中添加processDripShopItems方法
2. 遍历订单商品items，调用service.Item().GetItem()获取商品详情（使用item.GetItemReq）
3. 调用isDripShopItem检查商品的 **Item.DeliveredBySupplier** 字段是否为1
4. 如果是dropship商品，设置 **SaleOrderItem.DeliveredBySupplier = true**
5. 调用selectFirstSupplier选择供应商（如需要）
6. 返回错误（如有）

Restrictions:
- 使用GoFrame框架和gerror
- 通过service.Item()调用服务
- 不直接依赖DAO
- 完善的错误处理和日志
- 遵循现有代码风格（参考buying.go中的其他方法）
- 遵循go-bmp.mdc规范

Success Criteria:
- 代码通过gofmt和go vet
- 业务逻辑正确实现
- 异常处理完善
- 包含详细注释
- 遵循BMP开发规范
```

### CreateInnerSaleOrderFromPurchaseOrder 集成

```
Role: Go Developer with GoFrame ERP expertise

Task: 在CreateInnerSaleOrderFromPurchaseOrder方法中集成dripShop逻辑

Context:
- File: ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go
- Leverage: 现有CreateInnerSaleOrderFromPurchaseOrder方法，design.md中的集成代码
- Requirements: requirements.md 所有需求

Implementation Steps:
1. 在CreateInnerSaleOrderFromPurchaseOrder方法中，在设置DeliveryDate之后
2. 调用s.processDripShopItems(ctx, salesOrder.Items)处理dripShop逻辑
3. 如果返回错误，使用gerror.Wrapf包装并返回
4. 继续执行原有的订单创建逻辑（设置仓库、价格表等）

Restrictions:
- 不破坏现有内部销售订单创建逻辑
- 保持向后兼容性
- 使用gerror处理错误
- 遵循现有代码风格
- 遵循go-bmp.mdc规范

Success Criteria:
- dripShop商品正确处理
- 非dripShop商品不受影响
- 错误处理正确
- 代码可维护性好
- 遵循BMP规范
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-26.md`
- 当执行任务中形成复盘/优化建议时，及时沉淀 Episode 并在本节更新名称。

---

**模板版本**: v1.0.0
**最后更新**: 2025-11-26
**维护者**: 后端开发组