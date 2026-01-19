# 销售订单成本卡价格不一致问题解决方案

> 📖 **问题描述**: TTPOS销售含有成本卡（物品）的商品时，ERP会产生两份销售单，导致应收金额和实收金额不一致

---

## 一、问题分析

### 1.1 问题现象

当TTPOS销售含有成本卡（BOM）的商品时：

- **商品已配置价格**：例如商品A售价100元
- **成本卡中的物品已配置价格**：例如物品B售价50元
- **销售时ERP产生两份销售单**：
  - 销售单1：商品A，价格100元
  - 销售单2：物品B，价格50元
- **财务数据不一致**：
  - 应收金额 = 商品价格 + 物品价格 = 100 + 50 = 150元
  - 实收金额 = 商品价格 = 100元
  - **差额**：50元（财务数据不一致）

### 1.2 根本原因

**成本卡（BOM）的定位混淆**：

1. **成本卡的作用**：
   - 成本卡（BOM）用于定义商品的**成本结构**（原材料组成）
   - 成本卡中的物品（Material）是用于**成本核算**和**库存管理**的
   - 成本卡中的物品**不应该**作为独立的销售商品

2. **当前实现的问题**：
   - 在ERPNext中，商品和成本卡中的物品都被当作**可销售的商品**处理
   - 同步订单时，既创建了商品的销售单，又创建了成本卡中物品的销售单
   - 导致价格重复计算

### 1.3 业务逻辑分析

**业务场景区分**：

#### 场景一：TTPOS销售给客户（需要解决的问题）

```
销售场景：
- 客户购买：商品A（售价100元）
- 商品A有成本卡，包含物品B（成本50元）

当前ERPNext处理：
- ✅ 创建销售单：商品A，价格100元
- ❌ 错误创建：物品B的销售单，价格50元（导致价格重复计算）

正确的处理方式：
- ✅ 创建销售单：商品A，价格100元
- ❌ 不应该创建：物品B的销售单（物品B只用于成本核算）

财务数据：
- 应收金额 = 100元（商品A的价格）
- 实收金额 = 100元（客户实际支付）
- ✅ 数据一致
```

#### 场景二：品牌采购、调拨单（总店卖给子店）

```
业务场景：
- 总店卖给子店：物品B（售价50元）
- 在ERPNext上以销售的形式完成
- 物品必须要有价格（因为需要作为销售商品）

ERPNext处理：
- ✅ 创建销售单：物品B，价格50元（这是正确的，因为总店卖给子店）

财务数据：
- 应收金额 = 50元（物品B的价格）
- 实收金额 = 50元（子店实际支付）
- ✅ 数据一致
```

**关键区别**：

| 场景 | 销售对象 | 物品是否需要价格 | 是否创建物品销售单 |
|------|---------|----------------|------------------|
| TTPOS销售给客户 | 客户 | ❌ 不需要 | ❌ 不应该创建 |
| 品牌采购、调拨单 | 子店 | ✅ 需要 | ✅ 应该创建 |

---

## 二、解决方案

### ⚠️ 重要说明

由于**品牌采购、调拨单**的业务需求（总店卖给子店，物品必须要有价格），**不能**将成本卡中的物品设置为"不可销售"。

因此，解决方案需要**区分业务场景**：
- **TTPOS销售给客户**：只创建商品的销售单，不创建成本卡中物品的销售单
- **品牌采购、调拨单**：正常创建物品的销售单（因为总店卖给子店）

---

### 方案一：代码修改方案（推荐）⭐

**核心思路**：在订单同步到ERPNext时，根据业务场景区分处理：
- **TTPOS销售订单**：只同步商品，不同步成本卡中的物品
- **品牌采购、调拨单**：正常同步物品（因为总店卖给子店）

#### 2.1 实现步骤

1. **定位订单同步代码**

   - 查找订单同步到ERPNext的代码位置
   - 确认如何处理商品和成本卡中的物品

2. **添加业务场景标识**

   ```go
   // 伪代码示例
   type OrderSyncContext struct {
       OrderType string // "customer_sale" 或 "transfer" 或 "brand_purchase"
       // ... 其他字段
   }
   ```

3. **修改同步逻辑**

   ```go
   // 伪代码示例
   func SyncOrderToErpNext(ctx *OrderSyncContext, order *SaleOrder) error {
       // 遍历订单商品
       for _, product := range order.Products {
           // 创建商品的销售单
           createSalesOrderItem(product.ItemCode, product.Price)
           
           // 根据业务场景决定是否同步成本卡中的物品
           if product.HasBomCard() {
               switch ctx.OrderType {
               case "customer_sale":
                   // TTPOS销售给客户：不创建成本卡中物品的销售单
                   // 但需要通过库存调整或生产订单记录物品出库
                   // 方案A：通过库存调整（Stock Entry）记录物品出库
                   createStockEntryForBomMaterials(product.BomCard.Materials, order)
                   
                   // 方案B：通过生产订单（Work Order）记录物品出库（如果商品是通过生产制作的）
                   // createWorkOrderForProduct(product, order)
                   break
               case "transfer", "brand_purchase":
                   // 品牌采购、调拨单：正常创建物品的销售单
                   for _, material := range product.BomCard.Materials {
                       createSalesOrderItem(material.ItemCode, material.Price)
                   }
                   break
               }
           }
       }
       return nil
   }
   
   // 创建库存调整单，记录成本卡物品出库
   func createStockEntryForBomMaterials(materials []*BomMaterial, order *SaleOrder) error {
       // 创建库存调整单（Material Issue），记录物品出库
       stockEntry := &StockEntry{
           StockEntryType: "Material Issue", // 物料出库
           Purpose:        "Material Issue",
           Company:        order.Company,
           PostingDate:    order.OrderDate,
           Items:          make([]*StockEntryItem, 0),
       }
       
       for _, material := range materials {
           stockEntry.Items = append(stockEntry.Items, &StockEntryItem{
               ItemCode:  material.ItemCode,
               Qty:       material.Qty * order.ProductQty, // 物品数量 = 成本卡数量 × 商品数量
               Uom:       material.Uom,
               SrcWarehouse: order.Warehouse, // 从销售仓库出库
           })
       }
       
       return createStockEntry(stockEntry)
   }
   ```

4. **添加配置项（可选）**

   - 添加配置项：`是否同步成本卡中的物品到销售单（TTPOS销售订单）`
   - 默认值：`否`
   - 允许管理员根据业务需求配置

#### 2.2 物品出库处理方案

**问题**：如果不同步成本卡中的物品到销售单，就没有物品出库记录，库存管理会有问题。

**解决方案**：通过**库存调整（Stock Entry）**或**生产订单（Work Order）**记录物品出库。

##### 方案A：库存调整（Stock Entry）- 推荐 ⭐

**核心思路**：创建库存调整单（Material Issue），记录成本卡中物品的出库。

**实现步骤**：

1. **创建库存调整单**

   ```go
   // 伪代码示例
   func createStockEntryForBomMaterials(materials []*BomMaterial, order *SaleOrder) error {
       stockEntry := &StockEntry{
           StockEntryType: "Material Issue", // 物料出库
           Purpose:        "Material Issue",
           Company:        order.Company,
           PostingDate:    order.OrderDate,
           Items:          make([]*StockEntryItem, 0),
       }
       
       for _, material := range materials {
           stockEntry.Items = append(stockEntry.Items, &StockEntryItem{
               ItemCode:     material.ItemCode,
               Qty:          material.Qty * order.ProductQty, // 物品数量 = 成本卡数量 × 商品数量
               Uom:          material.Uom,
               SrcWarehouse: order.Warehouse, // 从销售仓库出库
           })
       }
       
       return createStockEntry(stockEntry)
   }
   ```

2. **关联销售单和库存调整单**

   为了便于追踪，可以在库存调整单中记录关联的销售单号：

   ```go
   // 伪代码示例
   stockEntry := &StockEntry{
       StockEntryType: "Material Issue",
       Purpose:        "Material Issue",
       Company:        order.Company,
       PostingDate:    order.OrderDate,
       ReferenceNo:    order.SalesOrderNo, // 关联销售单号
       ReferenceType:  "Sales Order",       // 关联单据类型
       Items:          make([]*StockEntryItem, 0),
   }
   ```

3. **优点**：
   - ✅ **记录物品出库**：通过库存调整单记录物品出库，不影响销售单
   - ✅ **财务数据一致**：销售单只有商品，应收金额 = 实收金额
   - ✅ **库存管理正确**：物品库存正确扣减
   - ✅ **符合ERPNext逻辑**：库存调整单是ERPNext标准的库存管理方式
   - ✅ **可追溯性**：通过关联字段可以追溯销售单和库存调整单的关系

4. **缺点**：
   - ❌ **需要额外创建单据**：每个订单需要创建销售单 + 库存调整单
   - ❌ **可能增加复杂度**：需要确保库存调整单和销售单的关联关系

##### 方案B：生产订单（Work Order）

**核心思路**：如果商品是通过生产制作的，可以通过生产订单记录物品消耗。

**适用场景**：
- 商品是通过生产制作的（如：制作菜品需要消耗原材料）
- ERPNext中商品配置了BOM（Bill of Materials）

**实现步骤**：

1. **创建生产订单**

   ```go
   // 伪代码示例
   func createWorkOrderForProduct(product *Product, order *SaleOrder) error {
       workOrder := &WorkOrder{
           ItemCode:     product.ItemCode,
           Qty:          order.ProductQty,
           Company:      order.Company,
           PlannedDate:  order.OrderDate,
           SourceWarehouse: order.Warehouse,
       }
       
       return createWorkOrder(workOrder)
   }
   ```

2. **优点**：
   - ✅ **符合生产逻辑**：如果商品是通过生产制作的，生产订单更符合业务逻辑
   - ✅ **自动消耗物品**：ERPNext会根据BOM自动消耗物品

3. **缺点**：
   - ❌ **只适用于生产商品**：如果商品不是通过生产制作的，不适用
   - ❌ **需要配置BOM**：需要在ERPNext中配置商品的BOM

#### 2.3 优点

- ✅ **区分业务场景**：正确处理不同业务场景的需求
- ✅ **灵活可控**：可以通过配置控制是否同步成本卡中的物品
- ✅ **财务数据一致**：TTPOS销售订单的应收金额 = 实收金额 = 商品价格
- ✅ **库存管理正确**：通过库存调整单或生产订单记录物品出库
- ✅ **不影响调拨单**：品牌采购、调拨单正常创建物品的销售单

#### 2.4 缺点

- ❌ **需要修改代码**：需要修改订单同步逻辑
- ❌ **需要充分测试**：确保不影响品牌采购、调拨单功能
- ❌ **可能增加复杂度**：需要创建额外的库存调整单或生产订单

---

### 方案二：ERPNext价格策略方案

**核心思路**：在ERPNext中，为TTPOS销售订单和品牌采购、调拨单使用不同的价格表：
- **TTPOS销售订单**：使用价格表A，成本卡中的物品价格为0
- **品牌采购、调拨单**：使用价格表B，成本卡中的物品价格为正常价格

#### 2.1 实现步骤

1. **在ERPNext中创建两个价格表**

   - **价格表A（TTPOS销售订单）**：
     - 商品价格：正常价格
     - 成本卡中的物品价格：0（不参与销售价格计算）
   
   - **价格表B（品牌采购、调拨单）**：
     - 商品价格：正常价格
     - 成本卡中的物品价格：正常价格（用于总店卖给子店）

2. **修改订单同步逻辑**

   ```go
   // 伪代码示例
   func SyncOrderToErpNext(ctx *OrderSyncContext, order *SaleOrder) error {
       // 根据业务场景选择价格表
       var priceList string
       switch ctx.OrderType {
       case "customer_sale":
           priceList = "TTPOS销售订单价格表" // 成本卡物品价格为0
       case "transfer", "brand_purchase":
           priceList = "品牌采购调拨价格表" // 成本卡物品价格为正常价格
       }
       
       // 创建销售订单，使用对应的价格表
       createSalesOrder(order, priceList)
       return nil
   }
   ```

#### 2.2 优点

- ✅ **保留成本核算功能**：成本卡中的物品仍然可以用于成本核算
- ✅ **财务数据一致**：TTPOS销售订单的应收金额 = 商品价格
- ✅ **不影响调拨单**：品牌采购、调拨单正常使用物品价格

#### 2.3 缺点

- ❌ **逻辑复杂**：需要维护两个价格表
- ❌ **可能造成混淆**：需要确保使用正确的价格表
- ❌ **需要修改代码**：需要修改订单同步逻辑，指定价格表

---


---

## 三、推荐方案

### 3.1 推荐方案：方案一（代码修改方案 + 库存调整）

**理由**：

1. ✅ **区分业务场景**：正确处理TTPOS销售订单和品牌采购、调拨单的不同需求
2. ✅ **财务数据一致**：TTPOS销售订单的应收金额 = 实收金额 = 商品价格
3. ✅ **库存管理正确**：通过库存调整单记录物品出库，确保库存正确扣减
4. ✅ **不影响调拨单**：品牌采购、调拨单正常创建物品的销售单
5. ✅ **符合业务逻辑**：TTPOS销售给客户时，成本卡中的物品只用于成本核算和库存管理

**物品出库方案**：推荐使用**库存调整（Stock Entry）**方案
- 创建销售单：只包含商品（解决财务问题）
- 创建库存调整单：记录成本卡中物品的出库（解决库存问题）

### 3.2 实施步骤

1. **第一步：确认业务场景**

   - **TTPOS销售订单**：客户购买商品，不应该创建成本卡中物品的销售单
   - **品牌采购、调拨单**：总店卖给子店，需要创建物品的销售单

2. **第二步：定位代码位置**

   - 查找订单同步到ERPNext的代码位置
   - 确认如何处理商品和成本卡中的物品
   - 确认如何区分TTPOS销售订单和品牌采购、调拨单

3. **第三步：修改同步逻辑**

   - 添加业务场景标识（订单类型）
   - 根据订单类型决定是否同步成本卡中的物品
   - **TTPOS销售订单**：
     - 创建销售单：只包含商品（不包含成本卡中的物品）
     - 创建库存调整单：记录成本卡中物品的出库（Material Issue）
   - **品牌采购、调拨单**：正常同步物品（创建物品的销售单）

4. **第四步：测试验证**

   - **测试TTPOS销售订单**：
     - 创建测试订单，包含有成本卡的商品
     - 验证ERPNext中只创建商品的销售单（不包含成本卡中的物品）
     - 验证ERPNext中创建库存调整单（记录成本卡中物品的出库）
     - 验证财务数据一致性（应收 = 实收 = 商品价格）
     - 验证库存数据正确性（物品库存正确扣减）
   
   - **测试品牌采购、调拨单**：
     - 创建测试调拨单
     - 验证ERPNext中正常创建物品的销售单
     - 验证财务数据一致性

5. **第五步：生产环境部署**

   - 在测试环境验证通过后，部署到生产环境
   - 监控订单同步情况
   - 监控财务数据一致性

---

## 四、相关文档

- [ERPNext销售订单审批工作流配置指南](../integrations/erpnext/sales-order-approval-workflow.md)
- [成本卡管理规范](../../human/business/workflows/cost-card-management.md)（如果存在）

---

## 四、实施注意事项

### 4.1 业务场景识别

**关键问题**：如何区分TTPOS销售订单和品牌采购、调拨单？

可能的识别方式：
1. **订单类型字段**：检查订单是否有类型字段（如 `order_type`）
2. **订单来源**：检查订单来源（如 `order_source`）
3. **客户类型**：检查客户类型（客户 vs 子店）
4. **订单编号前缀**：检查订单编号前缀（如 `SO-` vs `TR-`）

### 4.2 代码修改建议

1. **添加订单类型常量**

   ```go
   const (
       OrderTypeCustomerSale = "customer_sale"  // TTPOS销售订单
       OrderTypeTransfer     = "transfer"       // 调拨单
       OrderTypeBrandPurchase = "brand_purchase" // 品牌采购
   )
   ```

2. **修改订单同步函数签名**

   ```go
   func SyncOrderToErpNext(ctx context.Context, order *SaleOrder, orderType string) error {
       // 根据订单类型决定是否同步成本卡中的物品
       // ...
   }
   ```

3. **添加配置项（可选）**

   ```go
   type OrderSyncConfig struct {
       SyncBomMaterialsForCustomerSale bool // TTPOS销售订单是否同步成本卡物品
       SyncBomMaterialsForTransfer     bool // 调拨单是否同步成本卡物品
   }
   ```

### 4.3 测试用例

**测试用例1：TTPOS销售订单（有成本卡的商品）**

```
输入：
- 订单类型：customer_sale
- 商品：商品A（售价100元，数量2，有成本卡，包含物品B，数量1）

预期结果：
- ERPNext创建销售单：商品A，价格100元，数量2
- ERPNext不创建：物品B的销售单
- ERPNext创建库存调整单（Material Issue）：
  - 物品B，数量2（成本卡数量1 × 商品数量2）
  - 从销售仓库出库
- 财务数据：应收 = 实收 = 100元 × 2 = 200元
- 库存数据：物品B库存减少2
```

**测试用例2：品牌采购、调拨单（有成本卡的商品）**

```
输入：
- 订单类型：transfer
- 商品：商品A（售价100元，有成本卡，包含物品B，价格50元）

预期结果：
- ERPNext创建销售单：商品A，价格100元
- ERPNext创建销售单：物品B，价格50元
- 财务数据：应收 = 实收 = 150元
```

---

## 五、问题记录

**问题发现时间**：2025-12-22  
**问题描述**：TTPOS销售含有成本卡的商品时，ERP会产生两份销售单，导致财务数据不一致  
**业务场景**：品牌采购、调拨单需要物品价格（总店卖给子店），不能将物品设置为"不可销售"  
**解决方案**：方案一（代码修改方案）- 区分业务场景，TTPOS销售订单不同步成本卡物品，调拨单正常同步  
**状态**：待实施

---

**最后更新**：2025-12-22  
**维护者**：TTPOS Team

