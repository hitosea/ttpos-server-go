# 品牌采购流程 ERPNext 实现方案

## 📋 概述

本文档基于品牌采购流程图，提供完整的 ERPNext 实现方案，并说明 TTPOS 侧需要做的调整。

## 🎯 流程图关键节点

根据流程图，品牌采购流程包含以下关键节点：

1. **门店**：创建 MR 申请
2. **采购部门**：
   - 审批 MR
   - **根据物品的默认供应商自动判断**：
     - **集采路径**：物品未勾选默认供应商（或默认供应商是总部）→ 自动创建 BOI（集采订单）→ 审批 BOI
     - **直采路径**：物品勾选了默认供应商（外部供应商）→ 自动创建 PO（直采订单）→ 审批 PO → 打印 PDF → 提交供应商
3. **仓库**：生成物料收货单 → 物品入库码检查 → 处理退货/换货/退款/报废/销售退货
4. **外部供应商**：接收 PO → 配送 → 处理换货/退款
5. **财务部门**：处理付款单/收款单 → 审批 → 完成付款/收款

**关键说明**：
- MR 审批后，系统**自动**根据物品的默认供应商判断走集采还是直采
- 一个 MR 可能同时包含集采和直采物品，系统会自动分别创建 BOI 和 PO
- BOI 和 PO 是**并行**创建的，不是串行流程

## 📦 品牌采购分类

品牌采购分为两个部分：

### 1. 集采部分（集中采购）
- **业务模式**：总部集中采购后，通过销售订单配送给门店
- **触发条件**：MR 审批后，物品**未勾选默认供应商**（或默认供应商是总部）
- **ERPNext 单据**：Sales Order（Inter Company Sales Order）
- **配送方式**：总部发货给门店
- **收货单据**：Delivery Note（发货单）
- **财务处理**：总部创建 Sales Invoice（销售发票），门店创建 Purchase Invoice（采购发票）

### 2. 直采部分（直接采购）
- **业务模式**：外部供应商直接配送给门店
- **触发条件**：MR 审批后，物品**勾选了默认供应商**（外部供应商）
- **ERPNext 单据**：Purchase Order
- **配送方式**：外部供应商直接发货给门店
- **收货单据**：Purchase Receipt（采购收货单）
- **财务处理**：门店创建 Purchase Invoice（采购发票）和付款单

### 3. 判断逻辑

**MR 审批后的判断流程**：

```
MR 审批通过
   ↓
遍历 MR 中的每个物品
   ↓
判断物品是否勾选了默认供应商？
   ├─ 是（外部供应商）→ 走直采流程，创建 Purchase Order
   └─ 否（或默认供应商是总部）→ 走集采流程，创建 Sales Order（BOI）
```

**关键字段**：
- Material Request Item 中的 `supplier` 字段（默认供应商）
- 如果 `supplier` 有值且不是总部 → 直采
- 如果 `supplier` 为空或是总部 → 集采

## 🔄 ERPNext 实现方案

### 1. 单据类型映射

#### 1.1 集采部分（集中采购）

| TTPOS 单据 | ERPNext 单据 | 说明 |
|-----------|-------------|------|
| BOI（内部采购订单/集采订单） | **Sales Order**（Inter Company Sales Order） | **集采订单走销售订单业务线，总部销售给门店。不再创建 Purchase Order** |
| 物料收货单（集采） | **Delivery Note** | **集采收货单，总部发货给门店** |
| 集采退货单 - 换货 | **Sales Invoice (Credit Note)** | **集采退货单（换货），门店退货给总部，总部配送换货** |
| 集采退货单 - 退款 | **Sales Invoice (Credit Note)** | **集采退货单（退款），门店退货给总部，总部退款** |

#### 1.2 直采部分（直接采购）

| TTPOS 单据 | ERPNext 单据 | 说明 |
|-----------|-------------|------|
| PO（外部采购订单/直采订单） | Purchase Order | **直采订单，外部供应商直接配送给门店** |
| 物料收货单（直采） | Purchase Receipt | **直采收货单，外部供应商发货给门店** |
| 直采退货单（DN）- 换货 | Purchase Invoice (Debit Note) | **直采退货单（换货），门店退货给外部供应商，供应商配送换货** |
| 直采退货单（DN）- 退款 | Purchase Invoice (Debit Note) | **直采退货单（退款），门店退货给外部供应商，供应商退款** |

#### 1.3 通用单据

| TTPOS 单据 | ERPNext 单据 | 说明 |
|-----------|-------------|------|
| MR（Material Request） | Material Request | 物料申请单 |
| 销售退货单 | Sales Invoice (Credit Note) | 销售退货单 |
| 付款单 | Payment Entry | 付款单 |
| 收款单 | Payment Entry | 收款单 |

**重要说明**：
- **集采部分（BOI）**：在 ERPNext 中走销售订单业务线，**直接创建 Sales Order，不再创建 Purchase Order**。总部作为销售方，门店作为购买方
- **直采部分（PO）**：在 ERPNext 中走采购订单业务线，创建 Purchase Order。外部供应商作为供应商，门店作为购买方
- **集采收货**：创建 Delivery Note（发货单），而不是 Purchase Receipt
- **直采收货**：创建 Purchase Receipt（采购收货单）
- **集采财务**：总部创建 Sales Invoice（销售发票），门店创建 Purchase Invoice（采购发票）
- **直采财务**：门店创建 Purchase Invoice（采购发票）和付款单

### 2. ERPNext 单据流程

#### 2.1 MR 申请流程

**关键逻辑**：审批后的 MR，根据物品是否勾选了默认供应商来区分集采和直采。

```
1. 门店创建 MR 申请
   ↓
2. ERPNext 创建 Material Request
   - 单据类型：Purchase
   - 状态：Draft
   - 物品可以设置默认供应商（可选）
   ↓
3. 采购部门审批
   - 状态：Submitted
   ↓
4. 根据物品的默认供应商判断走集采还是直采：
   - 如果物品勾选了默认供应商（外部供应商）→ 走直采流程，创建 Purchase Order
   - 如果物品未勾选默认供应商（或默认供应商是总部）→ 走集采流程，创建 Sales Order
```

**ERPNext API 调用**：

```python
# 创建 Material Request
POST /api/resource/Material Request
{
    "material_request_type": "Purchase",
    "transaction_date": "2025-01-15",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "uom": "Nos",
            "warehouse": "Stores - Company",
            "supplier": "Supplier A - Company"  // 可选：默认供应商（如果设置，则走直采）
        },
        {
            "item_code": "ITEM-002",
            "qty": 50,
            "uom": "Nos",
            "warehouse": "Stores - Company"
            // 未设置默认供应商，走集采
        }
    ]
}

# 提交 Material Request
POST /api/resource/Material Request/{name}
{
    "action": "submit"
}
```

**判断逻辑**：

```python
# 审批 MR 后，根据物品的默认供应商判断流程
def process_material_request(mr_name):
    mr = get_material_request(mr_name)
    
    for item in mr.items:
        if item.supplier:  # 如果物品勾选了默认供应商
            # 走直采流程：创建 Purchase Order
            create_purchase_order(
                supplier=item.supplier,
                items=[item]
            )
        else:  # 如果物品未勾选默认供应商
            # 走集采流程：创建 Sales Order
            create_sales_order(
                customer=store_branch,
                items=[item]
            )
```

#### 2.2 BOI（集采订单）流程

**业务逻辑**：BOI 对应**集采部分**，走销售订单业务线，总部集中采购后通过销售订单配送给门店。**不再创建 Purchase Order，直接创建 Sales Order**。

**触发条件**：MR 审批后，物品**未勾选默认供应商**（或默认供应商是总部）时，**自动创建** BOI（集采订单）。

```
1. MR 审批通过
   ↓
2. 系统检测到物品未勾选默认供应商（或默认供应商是总部）
   ↓
3. 自动创建 BOI（集采订单）
   - 关联 MR 申请
   - 选择内部供应商（总部）
   ↓
4. ERPNext 直接创建 Inter Company Sales Order（销售订单）
   - 客户：门店（作为客户）
   - 公司：总部（作为销售方）
   - 状态：Draft
   ↓
5. 审批 BOI
   - Sales Order 状态：Submitted
```

**ERPNext API 调用**：

```python
# 直接创建 Inter Company Sales Order（销售订单）
POST /api/resource/Sales Order
{
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "transaction_date": "2025-01-15",
    "delivery_date": "2025-01-20",
    "set_warehouse": "Headquarters Warehouse - Company",
    "selling_price_list": "Standard Selling",  // 销售价格表
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse - Company"
        }
    ]
}

# 提交 Sales Order
POST /api/resource/Sales Order/{sales_order_name}
{
    "action": "submit"
}
```

**关键点**：
- BOI 在 ERPNext 中**只创建 Sales Order**，不再创建 Purchase Order
- 总部作为销售方，门店作为客户
- 门店收货时，创建的是 **Delivery Note**（发货单），而不是 Purchase Receipt

#### 2.3 PO（直采订单）流程

**业务逻辑**：PO 对应**直采部分**，走采购订单业务线，外部供应商直接配送给门店。

**触发条件**：MR 审批后，物品**勾选了默认供应商**（外部供应商）时，创建 PO（直采订单）。

```
1. MR 审批后，检测到物品勾选了默认供应商（外部供应商）
   ↓
2. 采购部门创建直采 PO
   - 关联 MR 申请
   - 供应商：物品的默认供应商
   ↓
3. ERPNext 创建 Purchase Order
   - 供应商：外部供应商（从物品的默认供应商获取）
   - 状态：Draft
   ↓
4. 审批 PO（优化：提交后自动根据金额进入对应审批状态）
   - 金额 < 100,000：自动进入 Pending PMA（采购经理审批）
   - 金额 ≥ 100,000：自动进入 Pending VP（VP 审批）
   - 审批通过后状态：Approved（docstatus = 1）
   ↓
5. 打印采购 PDF
   ↓
6. 提交供应商（外部供应商直接配送给门店）
```

**ERPNext API 调用**：

```python
# 创建外部采购订单（PO）
POST /api/resource/Purchase Order
{
    "supplier": "External Supplier - Company",
    "transaction_date": "2025-01-15",
    "schedule_date": "2025-01-25",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Stores - Company"
        }
    ],
    "grand_total": 1000.00
}

# 提交 Purchase Order
POST /api/resource/Purchase Order/{name}
{
    "action": "submit"
}

# 获取采购订单 PDF
GET /api/method/frappe.utils.print_format.download_pdf
{
    "doctype": "Purchase Order",
    "name": "PO-00001",
    "format": "Standard"
}
```

#### 2.4 物料收货流程（关键：按供应商拆分）

##### 2.4.1 外部采购收货流程

```
1. 外部供应商配送货物到门店
   ↓
2. 门店收货时，根据供应商拆分收货单
   - 一个采购订单可能包含多个供应商的物品
   - 需要为每个供应商创建独立的 Purchase Receipt
   ↓
3. ERPNext 创建 Purchase Receipt
   - 关联 Purchase Order
   - 状态：Draft
   ↓
4. 物品入库码检查
   - 检查物品编码是否正确
   - 检查数量是否匹配
   ↓
5. 确认收货
   - 状态：Submitted
   - 自动更新库存
```

**ERPNext API 调用（按供应商拆分）**：

```python
# 场景：一个采购订单包含多个供应商的物品
# 需要为每个供应商创建独立的 Purchase Receipt

# 供应商 A 的收货单
POST /api/resource/Purchase Receipt
{
    "supplier": "Supplier A - Company",
    "purchase_order": "PO-00001",
    "posting_date": "2025-01-20",
    "posting_time": "10:00:00",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 50,
            "uom": "Nos",
            "warehouse": "Stores - Company",
            "purchase_order_item": "PO-ITEM-001"
        }
    ]
}

# 供应商 B 的收货单
POST /api/resource/Purchase Receipt
{
    "supplier": "Supplier B - Company",
    "purchase_order": "PO-00001",
    "posting_date": "2025-01-20",
    "posting_time": "14:00:00",
    "set_warehouse": "Stores - Company",
    "items": [
        {
            "item_code": "ITEM-002",
            "qty": 30,
            "uom": "Nos",
            "warehouse": "Stores - Company",
            "purchase_order_item": "PO-ITEM-002"
        }
    ]
}

# 提交 Purchase Receipt
POST /api/resource/Purchase Receipt/{name}
{
    "action": "submit"
}
```

##### 2.4.2 集采收货流程（走销售订单业务线）

**业务逻辑**：集采部分由总部仓库人员创建 Delivery Note 给司机配送，门店收货时确认收货，门店可以针对 Delivery Note 进行退货。

**⚠️ 关键：按总部仓库拆分 Delivery Note**

一个 Sales Order（集采订单）可能包含来自不同总部仓库的物品，需要**按仓库拆分**创建多个 Delivery Note，方便门店人员根据不同的仓库发货进行收货：

- **仓库 A 的物品** → 创建 Delivery Note A
- **仓库 B 的物品** → 创建 Delivery Note B
- **仓库 C 的物品** → 创建 Delivery Note C

**拆分规则**：
1. 从 Sales Order 中获取所有待发货的物品
2. 按物品的 `warehouse` 字段（源仓库）进行分组
3. 为每个仓库创建一个独立的 Delivery Note
4. 每个 Delivery Note 只包含来自同一仓库的物品

```
1. 总部仓库人员创建 Delivery Note（发货单）
   - 关联 Sales Order（Inter Company Sales Order，集采订单）
   - ⚠️ 关键：系统自动按仓库拆分，为每个仓库创建独立的 Delivery Note
   - 状态：Draft
   ↓
2. 提交 Delivery Note
   - 状态：Submitted
   - 自动更新库存（从对应的总部仓库扣减）
   ↓
3. 司机配送货物到门店
   - 不同仓库的物品可能由不同的司机或批次配送
   ↓
4. 门店收货时，确认收货
   - 选择对应的 Delivery Note（按仓库区分）
   - 物品入库码检查
   - 检查物品编码是否正确
   - 检查数量是否匹配
   ↓
5. 门店确认收货
   - 更新库存（门店仓库增加）
   - 门店可以针对 Delivery Note 进行退货
```

**ERPNext API 调用（按仓库拆分）**：

```python
# 场景：一个 Sales Order 包含来自不同总部仓库的物品
# 需要为每个仓库创建独立的 Delivery Note

# 步骤1：获取 Sales Order 信息，按仓库分组物品
sales_order = get_sales_order("SO-00001")

# 按仓库分组物品
warehouse_items = {}
for item in sales_order.items:
    warehouse = item.warehouse  # 源仓库（总部仓库）
    if warehouse not in warehouse_items:
        warehouse_items[warehouse] = []
    warehouse_items[warehouse].append(item)

# 步骤2：为每个仓库创建独立的 Delivery Note

# 仓库 A 的 Delivery Note
POST /api/resource/Delivery Note
{
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "delivery_date": "2025-01-20",
    "set_warehouse": "Headquarters Warehouse A - Company",  // 源仓库（总部仓库A）
    "set_target_warehouse": "Store Warehouse - Company",   // 目标仓库（门店）
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 50,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse A - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-001"
        }
    ]
}

# 仓库 B 的 Delivery Note
POST /api/resource/Delivery Note
{
    "customer": "Store Branch - Company",
    "company": "Headquarters - Company",
    "delivery_date": "2025-01-20",
    "set_warehouse": "Headquarters Warehouse B - Company",  // 源仓库（总部仓库B）
    "set_target_warehouse": "Store Warehouse - Company",
    "items": [
        {
            "item_code": "ITEM-002",
            "qty": 30,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse B - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-002"
        }
    ]
}

# 步骤3：提交每个 Delivery Note（总部仓库人员操作）
POST /api/resource/Delivery Note/{name}
{
    "action": "submit"
}
# 提交后，库存从对应的总部仓库扣减

# 步骤4：门店收货确认（门店操作）
# 门店在 TTPOS 中按 Delivery Note 分别确认收货
# 每个 Delivery Note 对应一个仓库的发货，方便门店人员区分
# 门店可以针对每个 Delivery Note 进行退货
```

**拆分逻辑伪代码**：

```python
def create_delivery_notes_by_warehouse(sales_order_name):
    """
    从 Sales Order 按仓库拆分创建多个 Delivery Note
    """
    # 1. 获取 Sales Order 信息
    sales_order = get_sales_order(sales_order_name)
    
    # 2. 按仓库分组物品
    warehouse_groups = {}
    for item in sales_order.items:
        source_warehouse = item.warehouse  # 源仓库（总部仓库）
        if source_warehouse not in warehouse_groups:
            warehouse_groups[source_warehouse] = []
        warehouse_groups[source_warehouse].append(item)
    
    # 3. 为每个仓库创建独立的 Delivery Note
    delivery_notes = []
    for warehouse, items in warehouse_groups.items():
        delivery_note = create_delivery_note({
            "customer": sales_order.customer,
            "company": sales_order.company,
            "set_warehouse": warehouse,  # 源仓库
            "set_target_warehouse": sales_order.set_target_warehouse,  # 目标仓库（门店）
            "items": items  # 只包含该仓库的物品
        })
        delivery_notes.append(delivery_note)
    
    return delivery_notes
```

**关键点**：
- **总部仓库人员**创建 Delivery Note（发货单），给到司机配送
- Delivery Note 关联的是 **Sales Order**（集采订单），而不是 Purchase Order
- ⚠️ **按仓库拆分**：一个 Sales Order 包含多个仓库的物品时，需要按仓库创建多个 Delivery Note
- 每个 Delivery Note 只包含来自同一仓库的物品，方便门店人员区分不同仓库的发货
- 提交 Delivery Note 后，库存从对应的总部仓库扣减
- **门店收货时**：按 Delivery Note 分别确认收货，更新库存（门店仓库增加）
- **门店可以针对每个 Delivery Note 进行退货**（创建 Sales Invoice Credit Note）

#### 2.5 采购退货/换货/退款流程

##### 2.5.1 直采退货/换货/退款流程

**业务逻辑**：直采部分退货给外部供应商，分为换货和退款两种处理方式。

**流程分支**：

```
仓库发现需要退货的物品
   ↓
判断处理方式：
   ├─ 换货处理
   │    ↓
   │   创建采购退货单（DN）
   │    ↓
   │   ERPNext 创建 Purchase Invoice（Debit Note）
   │   - 类型：Debit Note
   │   - 关联 Purchase Receipt
   │   - 状态：Draft
   │    ↓
   │   审批退货单
   │   - 状态：Submitted
   │    ↓
   │   供应商配送换货
   │
   └─ 退款处理
        ↓
       创建采购退货单（DN）
        ↓
       ERPNext 创建 Purchase Invoice（Debit Note）
       - 类型：Debit Note
       - 关联 Purchase Receipt
       - 状态：Draft
        ↓
       审批退货单
       - 状态：Submitted
        ↓
       供应商退款
```

**ERPNext API 调用**：

```python
# 创建采购退货单（Debit Note）- 换货或退款
POST /api/resource/Purchase Invoice
{
    "is_return": 1,
    "is_debit_note": 1,
    "supplier": "Supplier A - Company",
    "purchase_receipt": "PR-00001",
    "posting_date": "2025-01-22",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": -10,  # 负数表示退货
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Stores - Company"
        }
    ]
}

# 提交 Purchase Invoice
POST /api/resource/Purchase Invoice/{name}
{
    "action": "submit"
}

# 换货处理：供应商配送换货后，创建新的 Purchase Receipt
# 退款处理：供应商退款，财务处理退款单
```

**关键点**：
- **换货处理**：创建采购退货单（DN）→ 审批 → 供应商配送换货 → 创建新的 Purchase Receipt
- **退款处理**：创建采购退货单（DN）→ 审批 → 供应商退款 → 财务处理退款单

##### 2.5.2 集采退货/换货/退款流程（走销售订单业务线）

**业务逻辑**：集采部分退货给总部，门店针对 Delivery Note 进行退货，分为换货和退款两种处理方式。

**流程分支**：

```
门店发现需要退货给总部的物品（集采退货）
   ↓
判断处理方式：
   ├─ 换货处理
   │    ↓
   │   创建集采退货单
   │    ↓
   │   ERPNext 创建 Sales Invoice（Credit Note）
   │   - 类型：Credit Note（销售退货）
   │   - 关联 Delivery Note
   │   - 状态：Draft
   │    ↓
   │   审批退货单
   │   - 状态：Submitted
   │   - 库存流转：从门店仓库扣减，总部仓库增加
   │    ↓
   │   总部配送换货
   │   - 创建新的 Delivery Note
   │
   └─ 退款处理
        ↓
       创建集采退货单
        ↓
       ERPNext 创建 Sales Invoice（Credit Note）
       - 类型：Credit Note（销售退货）
       - 关联 Delivery Note
       - 状态：Draft
        ↓
       审批退货单
       - 状态：Submitted
       - 库存流转：从门店仓库扣减，总部仓库增加
        ↓
       总部退款
       - 财务处理退款单
```

**ERPNext API 调用**：

```python
# 创建集采退货单（Credit Note，销售退货）- 换货或退款
POST /api/resource/Sales Invoice
{
    "is_return": 1,
    "is_credit_note": 1,
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "delivery_note": "DN-00001",           // 关联 Delivery Note
    "posting_date": "2025-01-22",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": -10,  # 负数表示退货
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Store Warehouse - Company"  // 从门店仓库扣减
        }
    ]
}

# 提交 Sales Invoice
POST /api/resource/Sales Invoice/{name}
{
    "action": "submit"
}

# 换货处理：总部配送换货后，创建新的 Delivery Note
# 退款处理：总部退款，财务处理退款单
```

**关键点**：
- 集采退货创建的是 **Sales Invoice（Credit Note）**，而不是 Purchase Invoice（Debit Note）
- 因为集采走销售订单业务线，退货也是销售退货
- 库存流转：从门店仓库扣减，总部仓库增加
- **换货处理**：创建退货单 → 审批 → 总部配送换货 → 创建新的 Delivery Note
- **退款处理**：创建退货单 → 审批 → 总部退款 → 财务处理退款单

#### 2.6 销售退货流程

```
1. 仓库发现需要销售退货的物品
   ↓
2. 创建销售退货单
   ↓
3. ERPNext 创建 Sales Return
   - 类型：Credit Note
   - 关联 Sales Invoice
   - 状态：Draft
   ↓
4. 审批退货单
   - 状态：Submitted
```

**ERPNext API 调用**：

```python
# 创建销售退货单（Credit Note）
POST /api/resource/Sales Invoice
{
    "is_return": 1,
    "is_credit_note": 1,
    "customer": "Customer - Company",
    "posting_date": "2025-01-22",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": -5,  # 负数表示退货
            "rate": 15.00,
            "uom": "Nos",
            "warehouse": "Stores - Company"
        }
    ]
}

# 提交 Sales Invoice
POST /api/resource/Sales Invoice/{name}
{
    "action": "submit"
}
```

#### 2.7 财务付款/收款流程

##### 2.7.1 外部采购付款流程

```
1. 根据采购收货单生成 Purchase Invoice（采购发票）
   ↓
2. 根据 Purchase Invoice 生成付款单
   ↓
3. ERPNext 创建 Payment Entry
   - 类型：Pay（付款）
   - 对方类型：Supplier（供应商）
   - 关联 Purchase Invoice
   - 状态：Draft
   ↓
4. 审批付款单
   - 状态：Submitted
   ↓
5. 完成付款
```

**ERPNext API 调用**：

```python
# 步骤1：创建 Purchase Invoice（采购发票）
POST /api/resource/Purchase Invoice
{
    "supplier": "Supplier A - Company",
    "purchase_receipt": "PR-00001",
    "posting_date": "2025-01-25",
    "due_date": "2025-02-25",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos"
        }
    ]
}

# 步骤2：创建付款单（Payment Entry）
POST /api/resource/Payment Entry
{
    "payment_type": "Pay",
    "party_type": "Supplier",
    "party": "Supplier A - Company",
    "posting_date": "2025-01-25",
    "paid_amount": 1000.00,
    "received_amount": 1000.00,
    "references": [
        {
            "reference_doctype": "Purchase Invoice",
            "reference_name": "PI-00001",
            "allocated_amount": 1000.00
        }
    ]
}

# 提交 Payment Entry
POST /api/resource/Payment Entry/{name}
{
    "action": "submit"
}
```

##### 2.7.2 集采付款/收款流程（走销售订单业务线）

**业务逻辑**：集采部分走销售订单业务线，总部是销售方，门店是购买方。

**总部视角（收款）**：
```
1. 根据集采收货单（Delivery Note）生成 Sales Invoice（销售发票）
   ↓
2. 根据 Sales Invoice 生成收款单
   ↓
3. ERPNext 创建 Payment Entry
   - 类型：Receive（收款）
   - 对方类型：Customer（客户，即门店）
   - 关联 Sales Invoice
   - 状态：Draft
   ↓
4. 审批收款单
   - 状态：Submitted
   ↓
5. 完成收款
```

**门店视角（付款）**：
```
1. 根据集采收货单（Delivery Note）生成 Purchase Invoice（采购发票）
   ↓
2. 根据 Purchase Invoice 生成付款单
   ↓
3. ERPNext 创建 Payment Entry
   - 类型：Pay（付款）
   - 对方类型：Supplier（供应商，即总部）
   - 关联 Purchase Invoice
   - 状态：Draft
   ↓
4. 审批付款单
   - 状态：Submitted
   ↓
5. 完成付款
```

**ERPNext API 调用**：

```python
# 总部视角：创建 Sales Invoice（销售发票）
POST /api/resource/Sales Invoice
{
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",    // 总部作为销售方
    "delivery_note": "DN-00001",
    "posting_date": "2025-01-25",
    "due_date": "2025-02-25",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos"
        }
    ]
}

# 总部视角：创建收款单（Payment Entry）
POST /api/resource/Payment Entry
{
    "payment_type": "Receive",
    "party_type": "Customer",
    "party": "Store Branch - Company",  // 门店作为客户
    "posting_date": "2025-01-25",
    "paid_amount": 1000.00,
    "received_amount": 1000.00,
    "references": [
        {
            "reference_doctype": "Sales Invoice",
            "reference_name": "SI-00001",
            "allocated_amount": 1000.00
        }
    ]
}

# 门店视角：创建 Purchase Invoice（采购发票）
POST /api/resource/Purchase Invoice
{
    "supplier": "Headquarters - Company",  // 总部作为供应商
    "company": "Store Branch - Company",    // 门店作为购买方
    "delivery_note": "DN-00001",
    "posting_date": "2025-01-25",
    "due_date": "2025-02-25",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 100,
            "rate": 10.00,
            "uom": "Nos"
        }
    ]
}

# 门店视角：创建付款单（Payment Entry）
POST /api/resource/Payment Entry
{
    "payment_type": "Pay",
    "party_type": "Supplier",
    "party": "Headquarters - Company",  // 总部作为供应商
    "posting_date": "2025-01-25",
    "paid_amount": 1000.00,
    "received_amount": 1000.00,
    "references": [
        {
            "reference_doctype": "Purchase Invoice",
            "reference_name": "PI-00001",
            "allocated_amount": 1000.00
        }
    ]
}
```

**关键点**：
- 内部采购走销售订单业务线，财务处理也分为两个视角：
  - **总部**：创建 Sales Invoice（销售发票）和收款单
  - **门店**：创建 Purchase Invoice（采购发票）和付款单
- 两个视角的发票和付款单是相互关联的，确保财务数据一致

### 3. 审批流程配置

ERPNext 支持自定义审批流程，可以通过以下方式实现：

#### 3.1 金额审批规则（优化版）

**优化说明**：
- 草稿提交后，系统根据金额自动判断下一个审批状态
- 金额 < 100,000：直接进入 PM 审核（Pending PMA）
- 金额 ≥ 100,000：直接进入 VP 审核（Pending VP）
- 移除中间状态"Pending review"，简化流程，减少用户操作步骤

```python
# 在 ERPNext 中配置 Workflow
# 金额 < 100,000：采购经理审批
# 金额 ≥ 100,000：VP 审批

# Workflow 配置示例（JSON）
{
    "workflow_name": "Purchase Order Approval",
    "document_type": "Purchase Order",
    "workflow_state_field": "workflow_state",
    "states": [
        {
            "state": "Draft",
            "allow_edit": "All"
        },
        {
            "state": "Pending PMA",
            "allow_edit": "Purchase Manager"
        },
        {
            "state": "Pending VP",
            "allow_edit": "VP"
        },
        {
            "state": "Approved",
            "allow_edit": "System Manager"
        },
        {
            "state": "Rejected",
            "allow_edit": "Purchase User"
        }
    ],
    "transitions": [
        {
            "state": "Draft",
            "action": "Submit for Approval",
            "next_state": "Pending PMA",
            "allowed": "Purchase User",
            "condition": "doc.grand_total < 100000"
        },
        {
            "state": "Draft",
            "action": "Submit for Approval",
            "next_state": "Pending VP",
            "allowed": "Purchase User",
            "condition": "doc.grand_total >= 100000"
        },
        {
            "state": "Pending PMA",
            "action": "Approve",
            "next_state": "Approved",
            "allowed": "Purchase Manager"
        },
        {
            "state": "Pending PMA",
            "action": "Reject",
            "next_state": "Rejected",
            "allowed": "Purchase Manager"
        },
        {
            "state": "Pending VP",
            "action": "Approve",
            "next_state": "Approved",
            "allowed": "VP"
        },
        {
            "state": "Pending VP",
            "action": "Reject",
            "next_state": "Rejected",
            "allowed": "VP"
        },
        {
            "state": "Rejected",
            "action": "Submit for Approval",
            "next_state": "Pending PMA",
            "allowed": "Purchase User",
            "condition": "doc.grand_total < 100000"
        },
        {
            "state": "Rejected",
            "action": "Submit for Approval",
            "next_state": "Pending VP",
            "allowed": "Purchase User",
            "condition": "doc.grand_total >= 100000"
        }
    ]
}
```

**工作流状态流转图**：

```
Draft（草稿）
    │
    │ [Submit for Approval]
    │ 根据金额自动判断：
    │
    ├─→ [金额 < 100,000]
    │         ▼
    │   ┌─────────────────────┐
    │   │ Pending PMA         │
    │   │ （待采购经理审批）   │
    │   └───────────┬─────────┘
    │               │
    │               │ [Approve] / [Reject]
    │               ▼
    │   ┌─────────────────────┐
    │   │ Approved / Rejected  │
    │   └─────────────────────┘
    │
    └─→ [金额 ≥ 100,000]
              ▼
        ┌─────────────────────┐
        │ Pending VP           │
        │ （待VP审批）         │
        └───────────┬─────────┘
                    │
                    │ [Approve] / [Reject]
                    ▼
        ┌─────────────────────┐
        │ Approved / Rejected  │
        └─────────────────────┘
```

**Transition Rules 配置表格**：

| No. | State* | Action* | Next State* | Allowed* | Condition |
|-----|--------|---------|-------------|----------|-----------|
| 1 | `Draft` | `Submit for Approval` | `Pending PMA` | `Purchase User` | `doc.grand_total < 100000` |
| 2 | `Draft` | `Submit for Approval` | `Pending VP` | `Purchase User` | `doc.grand_total >= 100000` |
| 3 | `Pending PMA` | `Approve` | `Approved` | `Purchase Manager` | （留空） |
| 4 | `Pending PMA` | `Reject` | `Rejected` | `Purchase Manager` | （留空） |
| 5 | `Pending VP` | `Approve` | `Approved` | `VP` | （留空） |
| 6 | `Pending VP` | `Reject` | `Rejected` | `VP` | （留空） |
| 7 | `Rejected` | `Submit for Approval` | `Pending PMA` | `Purchase User` | `doc.grand_total < 100000` |
| 8 | `Rejected` | `Submit for Approval` | `Pending VP` | `Purchase User` | `doc.grand_total >= 100000` |

**关键优化点**：
1. ✅ **移除中间状态**：不再需要"Pending review"状态，提交后直接根据金额进入对应审批状态
2. ✅ **自动判断**：提交时系统根据金额自动选择审批路径，无需用户手动选择
3. ✅ **简化流程**：减少用户操作步骤，提升效率
4. ✅ **条件判断**：在 Transition Rules 中使用 Condition 字段实现金额判断逻辑

## 🔧 TTPOS 侧需要做的调整

### 0. MR 审批后的集采/直采判断逻辑（新增）

#### 0.1 实现逻辑

**关键点**：MR 审批后，根据物品是否勾选了默认供应商来判断走集采还是直采。

```go
// 在 MR 审批后，根据物品的默认供应商判断流程
func (s *materialRequestSrv) ProcessMaterialRequestAfterApprove(
    ctx context.Context,
    mrUuid uint64,
) error {
    // 获取 MR 及其明细
    mr, err := s.repo.GetByUuid(mrUuid, s.repo.WithItems())
    if err != nil {
        return err
    }
    
    // 按默认供应商分组物品
    directPurchaseItems := make([]model.MaterialRequestItem, 0)  // 直采物品
    centralizedPurchaseItems := make([]model.MaterialRequestItem, 0)  // 集采物品
    
    for _, item := range mr.Items {
        // 判断物品是否有默认供应商（外部供应商）
        if item.DefaultSupplierErpCode != "" && !s.isHeadquartersSupplier(item.DefaultSupplierErpCode) {
            // 有默认供应商且不是总部 → 直采
            directPurchaseItems = append(directPurchaseItems, item)
        } else {
            // 无默认供应商或默认供应商是总部 → 集采
            centralizedPurchaseItems = append(centralizedPurchaseItems, item)
        }
    }
    
    // 为直采物品创建 Purchase Order
    if len(directPurchaseItems) > 0 {
        err := s.createDirectPurchaseOrder(ctx, mr, directPurchaseItems)
        if err != nil {
            return err
        }
    }
    
    // 为集采物品创建 Sales Order（BOI）
    if len(centralizedPurchaseItems) > 0 {
        err := s.createCentralizedPurchaseOrder(ctx, mr, centralizedPurchaseItems)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// 创建直采订单（Purchase Order）
func (s *materialRequestSrv) createDirectPurchaseOrder(
    ctx context.Context,
    mr *model.MaterialRequest,
    items []model.MaterialRequestItem,
) error {
    // 按供应商分组物品
    supplierGroups := make(map[string][]model.MaterialRequestItem)
    for _, item := range items {
        supplierCode := item.DefaultSupplierErpCode
        supplierGroups[supplierCode] = append(supplierGroups[supplierCode], item)
    }
    
    // 为每个供应商创建独立的 Purchase Order
    for supplierCode, supplierItems := range supplierGroups {
        // 创建 Purchase Order
        poReq := req.PurchaseOrderCreateReq{
            SupplierErpCode: supplierCode,
            Items:            s.convertItemsToPurchaseOrderItems(supplierItems),
            PurchaseType:    constant.PurchaseTypeExternal,  // 外部采购（直采）
            // ... 其他字段
        }
        
        _, err := s.purchaseOrderSrv.CreatePurchaseOrder(ctx, poReq)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// 创建集采订单（Sales Order，BOI）
func (s *materialRequestSrv) createCentralizedPurchaseOrder(
    ctx context.Context,
    mr *model.MaterialRequest,
    items []model.MaterialRequestItem,
) error {
    // 创建 BOI（集采订单）
    boiReq := req.PurchaseOrderCreateReq{
        PurchaseType:      constant.PurchaseTypeInternal,  // 内部采购（集采）
        SupplierErpCode:   constant.ErpHeadquartersSupplierCode,  // 总部供应商
        Items:             s.convertItemsToPurchaseOrderItems(items),
        // ... 其他字段
    }
    
    _, err := s.purchaseOrderSrv.CreatePurchaseOrder(ctx, boiReq)
    if err != nil {
        return err
    }
    
    return nil
}
```

#### 0.2 Material Request Item 模型调整

```go
// Material Request Item 需要添加默认供应商字段
type MaterialRequestItem struct {
    // ... 现有字段 ...
    
    // 新增：默认供应商编码（用于判断直采/集采）
    // 如果设置了默认供应商（外部供应商），则走直采流程
    // 如果未设置默认供应商（或默认供应商是总部），则走集采流程
    DefaultSupplierErpCode string `gorm:"column:default_supplier_erp_code;type:varchar(255);default:'';comment:默认供应商编码（如果设置且不是总部，则走直采）" json:"default_supplier_erp_code"`
    DefaultSupplierName    string `gorm:"column:default_supplier_name;type:varchar(255);default:'';comment:默认供应商名称" json:"default_supplier_name"`
}
```

#### 0.3 数据库迁移脚本

```sql
-- 在 Material Request Item 表中添加默认供应商字段
ALTER TABLE `ttpos_material_request_item`
ADD COLUMN `default_supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '默认供应商编码（如果设置且不是总部，则走直采）' AFTER `material_uuid`,
ADD COLUMN `default_supplier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '默认供应商名称' AFTER `default_supplier_erp_code`,
ADD INDEX `idx_default_supplier_erp_code` (`default_supplier_erp_code`);
```

### 1. 采购订单模型调整

#### 1.1 支持多供应商采购订单

**当前问题**：
- 一个采购订单只能关联一个供应商
- 收货时无法按供应商拆分收货单

**需要调整**：

```go
// 在 PurchaseOrder 模型中添加字段
type PurchaseOrder struct {
    // ... 现有字段 ...
    
    // 新增：是否多供应商采购订单
    IsMultiSupplier int `gorm:"column:is_multi_supplier;type:tinyint(1);default:0;comment:是否多供应商：0-否；1-是" json:"is_multi_supplier"`
}

// 新增：采购订单供应商关联表
type PurchaseOrderSupplier struct {
    BaseModel
    PurchaseOrderUuid uint64 `gorm:"column:purchase_order_uuid;type:bigint(20);not null;index" json:"purchase_order_uuid"`
    SupplierErpCode   string `gorm:"column:supplier_erp_code;type:varchar(255);not null" json:"supplier_erp_code"`
    SupplierName      string `gorm:"column:supplier_name;type:varchar(255);not null" json:"supplier_name"`
}

// 在 PurchaseOrderItem 中添加供应商字段
type PurchaseOrderItem struct {
    // ... 现有字段 ...
    
    // 新增：供应商编码（用于多供应商场景）
    SupplierErpCode string `gorm:"column:supplier_erp_code;type:varchar(255);default:'';comment:供应商编码（多供应商场景）" json:"supplier_erp_code"`
    SupplierName    string `gorm:"column:supplier_name;type:varchar(255);default:'';comment:供应商名称（多供应商场景）" json:"supplier_name"`
}
```

#### 1.2 收货单按供应商拆分

**当前问题**：
- 收货单基于采购订单创建，一个采购订单只能创建一个收货单
- 无法按供应商拆分收货单

**需要调整**：

```go
// 修改 CreatePurchaseReceiptOrder 方法
// 支持按供应商拆分收货单

func (s *purchaseReceiptOrderSrv) CreatePurchaseReceiptOrder(
    ctx context.Context,
    req req.PurchaseReceiptCreateReq,
) (resp.PurchaseReceiptOrderCreateResp, error) {
    // ... 现有验证逻辑 ...
    
    // 按供应商分组收货明细
    supplierGroups := make(map[string][]req.PurchaseReceiptItemReq)
    for _, item := range req.Items {
        // 获取物品对应的供应商
        supplierCode := s.getSupplierCodeForItem(ctx, item.PurchaseOrderItemUuid)
        supplierGroups[supplierCode] = append(supplierGroups[supplierCode], item)
    }
    
    // 为每个供应商创建独立的收货单
    var receiptOrderUuids []uint64
    for supplierCode, items := range supplierGroups {
        receiptReq := req.PurchaseReceiptCreateReq{
            PurchaseOrderUuid: req.PurchaseOrderUuid,
            Items:             items,
            ReceiveTime:       req.ReceiveTime,
            IsConfirm:         req.IsConfirm,
        }
        
        receiptOrder, err := s.createReceiptOrderForSupplier(ctx, receiptReq, supplierCode)
        if err != nil {
            return resp.PurchaseReceiptOrderCreateResp{}, err
        }
        
        receiptOrderUuids = append(receiptOrderUuids, receiptOrder.Uuid)
    }
    
    return resp.PurchaseReceiptOrderCreateResp{
        ReceiptOrderUuids: receiptOrderUuids,
    }, nil
}

// 新增：为特定供应商创建收货单
func (s *purchaseReceiptOrderSrv) createReceiptOrderForSupplier(
    ctx context.Context,
    req req.PurchaseReceiptCreateReq,
    supplierCode string,
) (*model.PurchaseReceiptOrder, error) {
    // 获取供应商信息
    supplier, err := repository.NewSupplierRepo(ctx.GetDB()).GetByErpCode(supplierCode)
    if err != nil {
        return nil, errors.WithMessage(errors.New("供应商不存在"), err.Error())
    }
    
    // 创建收货单（与现有逻辑相同，但指定供应商）
    receiptOrder := &model.PurchaseReceiptOrder{
        // ... 现有字段 ...
        SupplierErpCode: supplierCode,
        SupplierName:    supplier.Name,
    }
    
    // ... 创建收货单逻辑 ...
    
    return receiptOrder, nil
}
```

### 2. 采购退货功能

#### 2.1 新增采购退货单模型

```go
// 新增：采购退货单模型
type PurchaseReturnOrder struct {
    BaseModel
    OrderNo                string  `gorm:"column:order_no;type:varchar(255);not null;comment:退货单号" json:"order_no"`
    ErpOrderNo             string  `gorm:"column:erp_order_no;type:varchar(255);default:'';comment:ERP退货单号" json:"erp_order_no"`
    PurchaseReceiptUuid    uint64  `gorm:"column:purchase_receipt_uuid;type:bigint(20);not null;index;comment:采购收货单UUID（直采）或Delivery Note UUID（集采）" json:"purchase_receipt_uuid"`
    PurchaseReceiptNo      string  `gorm:"column:purchase_receipt_no;type:varchar(255);not null;comment:采购收货单号（直采）或Delivery Note号（集采）" json:"purchase_receipt_no"`
    DeliveryNoteUuid       uint64  `gorm:"column:delivery_note_uuid;type:bigint(20);default:0;index;comment:Delivery Note UUID（集采）" json:"delivery_note_uuid"`
    DeliveryNoteNo         string  `gorm:"column:delivery_note_no;type:varchar(255);default:'';comment:Delivery Note号（集采）" json:"delivery_note_no"`
    SupplierErpCode        string  `gorm:"column:supplier_erp_code;type:varchar(255);not null;comment:供应商编码" json:"supplier_erp_code"`
    SupplierName           string  `gorm:"column:supplier_name;type:varchar(255);not null;comment:供应商名称" json:"supplier_name"`
    Status                 int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态：0-待提交 1-待审核 2-已通过 3-已驳回" json:"status"`
    ReturnType             int     `gorm:"column:return_type;type:int(10);not null;default:1;comment:退货类型：1-换货 2-退款" json:"return_type"`
    ProcessType            int     `gorm:"column:process_type;type:int(10);not null;default:1;comment:处理类型：1-直采退货 2-集采退货" json:"process_type"`
    ExchangeDeliveryNoteNo string  `gorm:"column:exchange_delivery_note_no;type:varchar(255);default:'';comment:换货Delivery Note号（集采换货时使用）" json:"exchange_delivery_note_no"`
    ReturnReason           string  `gorm:"column:return_reason;type:varchar(255);default:'';comment:退货原因" json:"return_reason"`
    ReturnTime             int64   `gorm:"column:return_time;type:int(10);default:0;comment:退货时间" json:"return_time"`
    ApproverUuid           uint64  `gorm:"column:approver_uuid;type:bigint(20);default:0;comment:审批人UUID" json:"approver_uuid"`
    ApproverName           string  `gorm:"column:approver_name;type:varchar(255);default:'';comment:审批人姓名" json:"approver_name"`
    
    // 关联关系
    Items []PurchaseReturnOrderItem `gorm:"foreignKey:ReturnOrderUuid;references:Uuid" json:"items,omitempty"`
}

// 采购退货单明细
type PurchaseReturnOrderItem struct {
    BaseModel
    ReturnOrderUuid         uint64  `gorm:"column:return_order_uuid;type:bigint(20);not null;index" json:"return_order_uuid"`
    PurchaseReceiptItemUuid uint64  `gorm:"column:purchase_receipt_item_uuid;type:bigint(20);not null;index" json:"purchase_receipt_item_uuid"`
    MaterialCode            string  `gorm:"column:material_code;type:varchar(255);not null" json:"material_code"`
    MaterialName            string  `gorm:"column:material_name;type:text;not null" json:"material_name"`
    MaterialUuid            uint64  `gorm:"column:material_uuid;type:bigint(20);not null" json:"material_uuid"`
    ReturnNum               float64 `gorm:"column:return_num;type:decimal(14,4);not null;comment:退货数量" json:"return_num"`
    UnitUuid                uint64  `gorm:"column:unit_uuid;type:bigint(20);not null" json:"unit_uuid"`
    UnitName                string  `gorm:"column:unit_name;type:varchar(255);not null" json:"unit_name"`
    Valuation               float64 `gorm:"column:valuation;type:decimal(14,2);default:0;comment:估值单价" json:"valuation"`
    TotalPrice              float64 `gorm:"column:total_price;type:decimal(14,2);default:0;comment:总价" json:"total_price"`
}
```

#### 2.2 新增采购退货服务接口

```go
type IPurchaseReturnOrderSrv interface {
    // 创建采购退货单
    CreatePurchaseReturnOrder(ctx context.Context, req req.PurchaseReturnOrderCreateReq) (resp.PurchaseReturnOrderCreateResp, error)
    
    // 提交采购退货单
    SubmitPurchaseReturnOrder(ctx context.Context, req req.PurchaseReturnOrderSubmitReq) error
    
    // 审批采购退货单
    ApprovePurchaseReturnOrder(ctx context.Context, req req.PurchaseReturnOrderApproveReq) error
    
    // 获取采购退货单列表
    GetPurchaseReturnOrderList(ctx context.Context, req req.PurchaseReturnOrderListReq) (resp.PurchaseReturnOrderListResp, error)
    
    // 获取采购退货单详情
    GetPurchaseReturnOrderDetail(ctx context.Context, req req.PurchaseReturnOrderDetailReq) (resp.PurchaseReturnOrderDetailResp, error)
}
```

### 3. 销售退货功能增强

#### 3.1 支持从采购收货单创建销售退货

```go
// 在销售退货服务中添加方法
func (s *salesReturnOrderSrv) CreateSalesReturnFromPurchaseReceipt(
    ctx context.Context,
    req req.SalesReturnFromPurchaseReceiptReq,
) (resp.SalesReturnOrderCreateResp, error) {
    // 1. 验证采购收货单
    // 2. 创建销售退货单
    // 3. 调用 ERPNext 创建 Sales Return
    // 4. 更新库存
}
```

### 4. 财务付款/收款功能

#### 4.1 新增付款单/收款单模型

```go
// 付款单模型
type PaymentEntry struct {
    BaseModel
    EntryNo          string  `gorm:"column:entry_no;type:varchar(255);not null;comment:付款单号" json:"entry_no"`
    ErpEntryNo       string  `gorm:"column:erp_entry_no;type:varchar(255);default:'';comment:ERP付款单号" json:"erp_entry_no"`
    PaymentType      int     `gorm:"column:payment_type;type:int(10);not null;comment:付款类型：1-付款 2-收款" json:"payment_type"`
    PartyType        string  `gorm:"column:party_type;type:varchar(50);not null;comment:对方类型：Supplier/Customer" json:"party_type"`
    PartyUuid        uint64  `gorm:"column:party_uuid;type:bigint(20);not null;index;comment:对方UUID" json:"party_uuid"`
    PartyName        string  `gorm:"column:party_name;type:varchar(255);not null;comment:对方名称" json:"party_name"`
    PaidAmount       float64 `gorm:"column:paid_amount;type:decimal(14,2);not null;comment:付款金额" json:"paid_amount"`
    Status           int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态：0-待提交 1-待审核 2-已通过 3-已驳回 4-已完成" json:"status"`
    PaymentDate      int64   `gorm:"column:payment_date;type:int(10);not null;comment:付款日期" json:"payment_date"`
    ApproverUuid     uint64  `gorm:"column:approver_uuid;type:bigint(20);default:0;comment:审批人UUID" json:"approver_uuid"`
    
    // 关联关系
    References []PaymentEntryReference `gorm:"foreignKey:PaymentEntryUuid;references:Uuid" json:"references,omitempty"`
}

// 付款单关联单据
type PaymentEntryReference struct {
    BaseModel
    PaymentEntryUuid uint64 `gorm:"column:payment_entry_uuid;type:bigint(20);not null;index" json:"payment_entry_uuid"`
    ReferenceType    string `gorm:"column:reference_type;type:varchar(50);not null;comment:关联单据类型" json:"reference_type"`
    ReferenceUuid    uint64 `gorm:"column:reference_uuid;type:bigint(20);not null;index;comment:关联单据UUID" json:"reference_uuid"`
    ReferenceNo       string `gorm:"column:reference_no;type:varchar(255);not null;comment:关联单据号" json:"reference_no"`
    AllocatedAmount   float64 `gorm:"column:allocated_amount;type:decimal(14,2);not null;comment:分配金额" json:"allocated_amount"`
}
```

#### 4.2 新增付款单/收款单服务接口

```go
type IPaymentEntrySrv interface {
    // 创建付款单
    CreatePaymentEntry(ctx context.Context, req req.PaymentEntryCreateReq) (resp.PaymentEntryCreateResp, error)
    
    // 创建收款单
    CreateReceiptEntry(ctx context.Context, req req.ReceiptEntryCreateReq) (resp.ReceiptEntryCreateResp, error)
    
    // 审批付款单/收款单
    ApprovePaymentEntry(ctx context.Context, req req.PaymentEntryApproveReq) error
    
    // 完成付款/收款
    CompletePaymentEntry(ctx context.Context, req req.PaymentEntryCompleteReq) error
}
```

### 5. ERPNext 集成调整

#### 5.1 收货单按供应商拆分调用 ERPNext

##### 5.1.1 外部采购收货单

```go
// 在创建外部采购收货单时，按供应商分组调用 ERPNext
func (s *purchaseReceiptOrderSrv) createErpPurchaseReceipt(
    ctx context.Context,
    receiptOrder *model.PurchaseReceiptOrder,
) error {
    // 外部采购收货单，创建 Purchase Receipt
    if receiptOrder.ReceiptType == constant.ReceiptTypeExternal {
        // 按供应商分组物品
        supplierItems := make(map[string][]model.PurchaseReceiptOrderItem)
        for _, item := range receiptOrder.Items {
            supplierCode := receiptOrder.SupplierErpCode
            supplierItems[supplierCode] = append(supplierItems[supplierCode], item)
        }
        
        // 为每个供应商创建独立的 Purchase Receipt
        for supplierCode, items := range supplierItems {
            erpReq := &buying.SavePurchaseReceiptReq{
                PurchaseOrderName: receiptOrder.PurchaseOrder.ErpOrderNo,
                Items:             s.convertItemsToErpFormat(items),
            }
            
            _, err := s.erpClient.SavePurchaseReceipt(ctx, erpReq)
            if err != nil {
                return errors.WithMessage(err, fmt.Sprintf("创建ERP收货单失败（供应商：%s）", supplierCode))
            }
        }
    }
    
    return nil
}
```

##### 5.1.2 集采收货单（走销售订单业务线）

**重要说明**：集采收货流程分为两个阶段：
1. **总部仓库人员**创建 Delivery Note（发货单），给到司机配送
2. **门店**确认收货，可以针对 Delivery Note 进行退货

```go
// 阶段1：总部仓库人员创建 Delivery Note（发货单）
// 这个操作应该在总部系统中完成，不在门店收货时创建
func (s *warehouseSrv) CreateDeliveryNoteForCentralizedPurchase(
    ctx context.Context,
    salesOrderName string,
    req req.CreateDeliveryNoteReq,
) error {
    // 总部仓库人员创建 Delivery Note
    erpReq := &buying.CreateDeliveryNoteFromInnerSaleOrderReq{
        SourceName:      salesOrderName,
        SourceWarehouse: req.SourceWarehouseErpCode,
        TargetWarehouse: req.TargetWarehouseErpCode,
    }
    
    deliveryNote, err := s.erpClient.CreateDeliveryNoteFromInnerSaleOrder(ctx, erpReq)
    if err != nil {
        return errors.WithMessage(err, "创建ERP发货单失败")
    }
    
    // 提交 Delivery Note，库存从总部仓库扣减
    _, err = s.erpClient.SubmitDeliveryNote(ctx, deliveryNote.Name)
    if err != nil {
        return errors.WithMessage(err, "提交ERP发货单失败")
    }
    
    return nil
}

// 阶段2：门店确认收货（针对已创建的 Delivery Note）
func (s *purchaseReceiptOrderSrv) confirmCentralizedPurchaseReceipt(
    ctx context.Context,
    receiptOrder *model.PurchaseReceiptOrder,
    deliveryNoteName string,
) error {
    // 集采收货单，门店确认收货
    if receiptOrder.ReceiptType == constant.ReceiptTypeInternal {
        // 获取 Delivery Note 信息（Delivery Note 已在总部创建并提交）
        deliveryNote, err := s.erpClient.GetDeliveryNote(ctx, deliveryNoteName)
        if err != nil {
            return errors.WithMessage(err, "获取ERP发货单失败")
        }
        
        // 门店确认收货，更新库存（门店仓库增加）
        // 注意：Delivery Note 已经在总部提交，这里只是确认收货
        // 门店可以针对 Delivery Note 进行退货
        
        // 更新收货单状态
        receiptOrder.Status = constant.ReceiptOrderStatusReceived
        receiptOrder.ErpOrderNo = deliveryNoteName  // 记录 Delivery Note 名称
        
        // 更新库存（门店仓库增加）
        err = s.updateMaterialStock(ctx, receiptOrder)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// 集采换货：总部创建换货 Delivery Note
func (s *warehouseSrv) CreateExchangeDeliveryNoteForCentralizedPurchase(
    ctx context.Context,
    originalSalesOrderName string,  // 原始 Sales Order 名称
    returnOrderUuid uint64,          // 退货单 UUID
    req req.CreateExchangeDeliveryNoteReq,
) error {
    // 获取退货单信息
    returnOrder, err := s.returnOrderRepo.GetByUuid(returnOrderUuid)
    if err != nil {
        return errors.WithMessage(err, "获取退货单失败")
    }
    
    // 验证退货类型是否为换货
    if returnOrder.ReturnType != constant.ReturnTypeExchange {
        return errors.New("退货类型不是换货")
    }
    
    // 总部仓库人员创建换货 Delivery Note
    erpReq := &buying.CreateDeliveryNoteFromInnerSaleOrderReq{
        SourceName:      originalSalesOrderName,  // 关联原始 Sales Order
        SourceWarehouse: req.SourceWarehouseErpCode,
        TargetWarehouse: req.TargetWarehouseErpCode,
    }
    
    deliveryNote, err := s.erpClient.CreateDeliveryNoteFromInnerSaleOrder(ctx, erpReq)
    if err != nil {
        return errors.WithMessage(err, "创建ERP换货发货单失败")
    }
    
    // 提交 Delivery Note，库存从总部仓库扣减
    _, err = s.erpClient.SubmitDeliveryNote(ctx, deliveryNote.Name)
    if err != nil {
        return errors.WithMessage(err, "提交ERP换货发货单失败")
    }
    
    // 更新退货单，记录换货 Delivery Note
    returnOrder.ExchangeDeliveryNoteNo = deliveryNote.Name
    err = s.returnOrderRepo.Update(returnOrder)
    if err != nil {
        return errors.WithMessage(err, "更新退货单失败")
    }
    
    return nil
}

// 集采换货：门店确认收货换货物品
func (s *purchaseReceiptOrderSrv) confirmCentralizedPurchaseExchangeReceipt(
    ctx context.Context,
    receiptOrder *model.PurchaseReceiptOrder,
    exchangeDeliveryNoteName string,
) error {
    // 集采换货收货单，门店确认收货
    if receiptOrder.ReceiptType == constant.ReceiptTypeInternal {
        // 获取换货 Delivery Note 信息（已在总部创建并提交）
        deliveryNote, err := s.erpClient.GetDeliveryNote(ctx, exchangeDeliveryNoteName)
        if err != nil {
            return errors.WithMessage(err, "获取ERP换货发货单失败")
        }
        
        // 门店确认收货换货物品，更新库存（门店仓库增加）
        receiptOrder.Status = constant.ReceiptOrderStatusReceived
        receiptOrder.ErpOrderNo = exchangeDeliveryNoteName  // 记录换货 Delivery Note 名称
        receiptOrder.IsExchange = true  // 标记为换货收货单
        
        // 更新库存（门店仓库增加）
        err = s.updateMaterialStock(ctx, receiptOrder)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// 集采换货：总部创建换货 Delivery Note
func (s *warehouseSrv) CreateExchangeDeliveryNoteForCentralizedPurchase(
    ctx context.Context,
    originalSalesOrderName string,  // 原始 Sales Order 名称
    returnOrderUuid uint64,          // 退货单 UUID
    req req.CreateExchangeDeliveryNoteReq,
) error {
    // 获取退货单信息
    returnOrder, err := s.returnOrderRepo.GetByUuid(returnOrderUuid)
    if err != nil {
        return errors.WithMessage(err, "获取退货单失败")
    }
    
    // 验证退货类型是否为换货
    if returnOrder.ReturnType != constant.ReturnTypeExchange {
        return errors.New("退货类型不是换货")
    }
    
    // 总部仓库人员创建换货 Delivery Note
    erpReq := &buying.CreateDeliveryNoteFromInnerSaleOrderReq{
        SourceName:      originalSalesOrderName,  // 关联原始 Sales Order
        SourceWarehouse: req.SourceWarehouseErpCode,
        TargetWarehouse: req.TargetWarehouseErpCode,
    }
    
    deliveryNote, err := s.erpClient.CreateDeliveryNoteFromInnerSaleOrder(ctx, erpReq)
    if err != nil {
        return errors.WithMessage(err, "创建ERP换货发货单失败")
    }
    
    // 提交 Delivery Note，库存从总部仓库扣减
    _, err = s.erpClient.SubmitDeliveryNote(ctx, deliveryNote.Name)
    if err != nil {
        return errors.WithMessage(err, "提交ERP换货发货单失败")
    }
    
    // 更新退货单，记录换货 Delivery Note
    returnOrder.ExchangeDeliveryNoteNo = deliveryNote.Name
    err = s.returnOrderRepo.Update(returnOrder)
    if err != nil {
        return errors.WithMessage(err, "更新退货单失败")
    }
    
    return nil
}

// 集采换货：门店确认收货换货物品
func (s *purchaseReceiptOrderSrv) confirmCentralizedPurchaseExchangeReceipt(
    ctx context.Context,
    receiptOrder *model.PurchaseReceiptOrder,
    exchangeDeliveryNoteName string,
) error {
    // 集采换货收货单，门店确认收货
    if receiptOrder.ReceiptType == constant.ReceiptTypeInternal {
        // 获取换货 Delivery Note 信息（已在总部创建并提交）
        deliveryNote, err := s.erpClient.GetDeliveryNote(ctx, exchangeDeliveryNoteName)
        if err != nil {
            return errors.WithMessage(err, "获取ERP换货发货单失败")
        }
        
        // 门店确认收货换货物品，更新库存（门店仓库增加）
        receiptOrder.Status = constant.ReceiptOrderStatusReceived
        receiptOrder.ErpOrderNo = exchangeDeliveryNoteName  // 记录换货 Delivery Note 名称
        receiptOrder.IsExchange = true  // 标记为换货收货单
        
        // 更新库存（门店仓库增加）
        err = s.updateMaterialStock(ctx, receiptOrder)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

#### 5.2 新增采购退货 ERPNext 接口

##### 5.2.1 外部采购退货接口

```go
// 在 ttpos-bmp 中添加外部采购退货接口
service BuyingService {
    // 创建外部采购退货单（Debit Note）
    rpc CreatePurchaseReturn(CreatePurchaseReturnReq) returns (CreatePurchaseReturnResp);
}

message CreatePurchaseReturnReq {
    string purchase_receipt_name = 1; // 采购收货单名称，必填
    string supplier = 2; // 供应商，必填
    string company_abbr = 3; // 公司缩写，必填
    string posting_date = 4; // 过账日期 Y-m-d，必填
    repeated PurchaseReturnItem items = 5; // 退货物品列表，必填
    string return_reason = 6; // 退货原因，可选
}

message PurchaseReturnItem {
    string item_code = 1; // 物品编码，必填
    double qty = 2; // 退货数量（负数），必填
    double rate = 3; // 单价，可选
    string uom = 4; // 单位，可选
    string warehouse = 5; // 仓库，可选
}
```

##### 5.2.2 内部采购退货接口（走销售订单业务线）

```go
// 在 ttpos-bmp 中添加内部采购退货接口
service SellingService {
    // 创建内部采购退货单（Credit Note，销售退货）
    rpc CreateInternalPurchaseReturn(CreateInternalPurchaseReturnReq) returns (CreateInternalPurchaseReturnResp);
}

message CreateInternalPurchaseReturnReq {
    string delivery_note_name = 1; // Delivery Note 名称，必填
    string customer = 2; // 客户（门店），必填
    string company_abbr = 3; // 公司缩写（总部），必填
    string posting_date = 4; // 过账日期 Y-m-d，必填
    repeated PurchaseReturnItem items = 5; // 退货物品列表，必填
    string return_reason = 6; // 退货原因，可选
}
```

#### 5.3 新增付款单/收款单 ERPNext 接口

```go
// 在 ttpos-bmp 中添加付款单/收款单接口
service AccountingService {
    // 创建付款单
    rpc CreatePaymentEntry(CreatePaymentEntryReq) returns (CreatePaymentEntryResp);
    
    // 创建收款单
    rpc CreateReceiptEntry(CreateReceiptEntryReq) returns (CreateReceiptEntryResp);
}

message CreatePaymentEntryReq {
    string payment_type = 1; // 付款类型：Pay/Receive，必填
    string party_type = 2; // 对方类型：Supplier/Customer，必填
    string party = 3; // 对方名称，必填
    string company_abbr = 4; // 公司缩写，必填
    string posting_date = 5; // 过账日期 Y-m-d，必填
    double paid_amount = 6; // 付款金额，必填
    repeated PaymentEntryReference references = 7; // 关联单据列表，必填
}

message PaymentEntryReference {
    string reference_doctype = 1; // 关联单据类型，必填
    string reference_name = 2; // 关联单据名称，必填
    double allocated_amount = 3; // 分配金额，必填
}
```

## 📝 操作流程说明

### 完整流程示例

#### 场景：门店采购多供应商物品

**步骤 1：门店创建 MR 申请**

```
1. 门店在 TTPOS 创建采购申请
   - 选择物品
   - 可以为物品设置默认供应商（可选）
   - 提交申请
   ↓
2. TTPOS 调用 ERPNext 创建 Material Request
   - 状态：Draft
   - 物品可以包含默认供应商信息
   ↓
3. 采购部门审批 MR
   - ERPNext 状态：Submitted
   ↓
4. 系统根据物品的默认供应商自动判断：
   - 有默认供应商（外部供应商）→ 走直采流程
   - 无默认供应商（或默认供应商是总部）→ 走集采流程
```

**步骤 2：MR 审批后自动创建 BOI（集采订单）或 PO（直采订单）**

**业务逻辑**：MR 审批后，系统根据物品的默认供应商自动判断走集采还是直采。

##### 2.1 集采订单（BOI）创建

**触发条件**：MR 审批后，物品**未勾选默认供应商**（或默认供应商是总部）。

```
1. MR 审批通过
   ↓
2. 系统检测到物品未勾选默认供应商（或默认供应商是总部）
   ↓
3. 自动创建集采订单（BOI）
   - 关联 MR 申请
   - 选择内部供应商（总部）
   - 标记为集采订单
   ↓
4. TTPOS 调用 ERPNext 直接创建 Inter Company Sales Order（销售订单）
   - 客户：门店（作为客户）
   - 公司：总部（作为销售方）
   - 状态：Draft
   ↓
5. 审批 BOI
   - Sales Order 状态：Submitted
```

##### 2.2 直采订单（PO）创建

**触发条件**：MR 审批后，物品**勾选了默认供应商**（外部供应商）。

```
1. MR 审批通过
   ↓
2. 系统检测到物品勾选了默认供应商（外部供应商）
   ↓
3. 自动创建直采订单（PO）
   - 关联 MR 申请
   - 供应商：物品的默认供应商
   - 标记为直采订单
   ↓
4. TTPOS 调用 ERPNext 创建 Purchase Order
   - 供应商：外部供应商（从物品的默认供应商获取）
   - 状态：Draft
   ↓
5. 审批 PO（优化：提交后自动根据金额进入对应审批状态）
   - 金额 < 100,000：自动进入 Pending PMA（采购经理审批）
   - 金额 ≥ 100,000：自动进入 Pending VP（VP 审批）
   - 审批通过后 ERPNext 状态：Approved（docstatus = 1）
   ↓
6. 打印采购 PDF
   - 调用 ERPNext API 获取 PDF
   ↓
7. 提交外部供应商（外部供应商直接配送给门店）
```

**注意**：一个 MR 可能同时包含集采和直采物品，系统会自动分别创建 BOI 和 PO。

**步骤 4：门店收货（关键：按仓库和供应商拆分）**

**拆分规则总结**：

1. **集采部分（BOI）**：按**总部仓库**拆分 Delivery Note
   - 一个 Sales Order 包含多个仓库的物品时，按仓库创建多个 Delivery Note
   - 每个 Delivery Note 只包含来自同一仓库的物品

2. **直采部分（PO）**：按**外部供应商**拆分 Purchase Receipt
   - 一个 Purchase Order 包含多个供应商的物品时，按供应商创建多个 Purchase Receipt
   - 每个 Purchase Receipt 只包含来自同一供应商的物品

**拆分目的**：
- 方便门店人员根据不同的仓库发货和供应商发货进行收货
- 每个单据对应一个明确的发货来源（仓库或供应商）
- 便于对账和问题追溯

##### 4.1 直采收货

**业务逻辑**：直采部分由外部供应商直接配送给门店。

```
1. 外部供应商直接配送货物到门店
   ↓
2. 门店在 TTPOS 创建收货单
   - 选择直采订单（PO）
   - 选择收货物品
   ↓
3. TTPOS 自动按供应商拆分收货单
   - 供应商 A 的物品 → 收货单 A
   - 供应商 B 的物品 → 收货单 B
   ↓
4. 为每个供应商创建独立的 ERPNext Purchase Receipt
   - Purchase Receipt A（供应商 A）
   - Purchase Receipt B（供应商 B）
   ↓
5. 物品入库码检查
   - 检查物品编码
   - 检查数量
   ↓
6. 确认收货
   - ERPNext 状态：Submitted
   - 自动更新库存（门店仓库增加）
```

##### 4.2 集采收货（走销售订单业务线）

**业务逻辑**：集采部分由总部仓库人员创建 Delivery Note 给司机配送，门店收货时确认收货，门店可以针对 Delivery Note 进行退货。

**⚠️ 关键：按总部仓库拆分 Delivery Note**

一个 Sales Order（集采订单）可能包含来自不同总部仓库的物品，系统需要**按仓库自动拆分**，为每个仓库创建独立的 Delivery Note：

- **仓库 A 的物品** → Delivery Note A
- **仓库 B 的物品** → Delivery Note B
- **外部供应商的物品** → Purchase Receipt（直采部分）

这样门店人员可以清楚地知道：
- 哪些物品来自总部仓库 A
- 哪些物品来自总部仓库 B
- 哪些物品来自外部供应商

**阶段1：总部仓库人员创建 Delivery Note（按仓库拆分）**

```
1. 总部仓库人员在总部系统中创建 Delivery Note（发货单）
   - 关联 Sales Order（Inter Company Sales Order，集采订单）
   - ⚠️ 系统自动按仓库拆分：
     * 从 Sales Order 获取所有待发货物品
     * 按物品的 warehouse 字段（源仓库）分组
     * 为每个仓库创建独立的 Delivery Note
   - 状态：Draft
   ↓
2. 提交 Delivery Note
   - 每个 Delivery Note 独立提交
   - 状态：Submitted
   - 自动更新库存（从对应的总部仓库扣减）
   ↓
3. 给到司机配送货物到门店
   - 不同仓库的物品可能由不同的司机或批次配送
   - 每个 Delivery Note 对应一个仓库的发货
```

**阶段2：门店确认收货（按 Delivery Note 分别确认）**

```
1. 司机配送货物到门店
   - 可能同时到达多个 Delivery Note 的货物（来自不同仓库）
   ↓
2. 门店在 TTPOS 确认收货
   - 选择对应的 Delivery Note（按仓库区分）
   - 系统显示该 Delivery Note 对应的仓库信息
   - 选择收货物品
   ↓
3. 物品入库码检查
   - 检查物品编码
   - 检查数量
   - 确认物品来自正确的仓库
   ↓
4. 确认收货
   - 更新库存（门店仓库增加）
   - 门店可以针对该 Delivery Note 进行退货
   - 每个 Delivery Note 独立处理，互不影响
```

**步骤 5：处理退货**

##### 5.1 直采退货

**业务逻辑**：直采部分退货给外部供应商。

```
1. 门店发现需要退货给外部供应商的物品（直采退货）
   ↓
2. 创建直采退货单（DN）
   - 关联直采收货单（Purchase Receipt）
   - 选择退货类型：换货/退款
   ↓
3. TTPOS 调用 ERPNext 创建 Purchase Invoice（Debit Note）
   - 状态：Draft
   ↓
4. 审批退货单
   - ERPNext 状态：Submitted
   ↓
5. 外部供应商接收退货或退款
```

##### 5.2 集采退货/换货/退款（走销售订单业务线）

**业务逻辑**：集采部分退货给总部，门店针对 Delivery Note 进行退货，分为换货和退款两种处理方式。

**5.2.1 集采换货处理**

```
1. 门店发现需要退货给总部的物品（集采退货）
   ↓
2. 选择"换货处理"
   ↓
3. 创建集采退货单
   - 关联 Delivery Note（发货单）
   - 退货类型：换货
   - 选择需要换货的物品和数量
   ↓
4. TTPOS 调用 ERPNext 创建 Sales Invoice（Credit Note）
   - 类型：Credit Note（销售退货）
   - 关联 Delivery Note
   - 状态：Draft
   ↓
5. 审批退货单
   - ERPNext 状态：Submitted
   - 库存流转：从门店仓库扣减，总部仓库增加
   ↓
6. 总部接收退货
   - 总部仓库人员确认收到退货物品
   ↓
7. 总部配送换货
   - 总部仓库人员创建新的 Delivery Note（换货）
   - 关联原始 Sales Order（集采订单）
   - 状态：Draft
   ↓
8. 提交新的 Delivery Note
   - 状态：Submitted
   - 库存流转：从总部仓库扣减
   ↓
9. 司机配送换货物品到门店
   ↓
10. 门店收货换货物品
    - 门店在 TTPOS 确认收货
    - 选择新的 Delivery Note（换货）
    - 确认换货物品和数量
    ↓
11. 物品入库码检查
    - 检查物品编码
    - 检查数量
    ↓
12. 确认收货
    - 更新库存（门店仓库增加）
    - 换货流程完成
```

**ERPNext API 调用**：

```python
# 步骤1：创建集采退货单（Credit Note）
POST /api/resource/Sales Invoice
{
    "is_return": 1,
    "is_credit_note": 1,
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "delivery_note": "DN-00001",           // 关联原始 Delivery Note
    "posting_date": "2025-01-22",
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": -10,  # 负数表示退货
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Store Warehouse - Company"  // 从门店仓库扣减
        }
    ]
}

# 步骤2：提交退货单
POST /api/resource/Sales Invoice/{credit_note_name}
{
    "action": "submit"
}
# 提交后，库存从门店仓库扣减，总部仓库增加

# 步骤3：总部创建换货 Delivery Note
POST /api/resource/Delivery Note
{
    "customer": "Store Branch - Company",  // 门店作为客户
    "company": "Headquarters - Company",   // 总部作为销售方
    "sales_order": "SO-00001",             // 关联原始 Sales Order
    "delivery_date": "2025-01-25",
    "set_warehouse": "Headquarters Warehouse - Company",  // 源仓库（总部）
    "set_target_warehouse": "Store Warehouse - Company",   // 目标仓库（门店）
    "items": [
        {
            "item_code": "ITEM-001",
            "qty": 10,  # 换货数量
            "rate": 10.00,
            "uom": "Nos",
            "warehouse": "Headquarters Warehouse - Company",
            "target_warehouse": "Store Warehouse - Company",
            "sales_order": "SO-00001",
            "sales_order_item": "SO-ITEM-001"
        }
    ]
}

# 步骤4：提交换货 Delivery Note
POST /api/resource/Delivery Note/{exchange_dn_name}
{
    "action": "submit"
}
# 提交后，库存从总部仓库扣减

# 步骤5：门店确认收货（在 TTPOS 中操作）
# 门店确认收货后，库存增加到门店仓库
```

**关键点**：
- 集采换货需要创建两个 Delivery Note：
  1. **原始 Delivery Note**：用于退货（通过 Credit Note 关联）
  2. **换货 Delivery Note**：用于配送换货物品
- 换货 Delivery Note 关联的是**原始 Sales Order**，而不是退货单
- 库存流转：
  - 退货时：门店仓库扣减 → 总部仓库增加
  - 换货时：总部仓库扣减 → 门店仓库增加

**5.2.2 集采退款处理**

```
1. 门店发现需要退货给总部的物品（集采退货）
   ↓
2. 选择"退款处理"
   ↓
3. 创建集采退货单
   - 关联 Delivery Note（发货单）
   - 退货类型：退款
   ↓
4. TTPOS 调用 ERPNext 创建 Sales Invoice（Credit Note）
   - 类型：Credit Note（销售退货）
   - 关联 Delivery Note
   - 状态：Draft
   ↓
5. 审批退货单
   - ERPNext 状态：Submitted
   - 库存流转：从门店仓库扣减，总部仓库增加
   ↓
6. 总部接收退货
   ↓
7. 总部退款
   ↓
8. 财务处理退款单
   - 总部：创建 Payment Entry（退款给门店）
   - 门店：创建 Payment Entry（收到退款）
```

**步骤 6：财务处理付款/收款**

##### 6.1 直采付款

**业务逻辑**：直采部分由门店付款给外部供应商。

```
1. 根据直采收货单（Purchase Receipt）生成 Purchase Invoice（采购发票）
   ↓
2. 根据 Purchase Invoice 生成付款单
   ↓
3. TTPOS 调用 ERPNext 创建 Payment Entry
   - 类型：Pay（付款）
   - 对方类型：Supplier（外部供应商）
   - 关联 Purchase Invoice
   - 状态：Draft
   ↓
4. 审批付款单
   - ERPNext 状态：Submitted
   ↓
5. 完成付款
   - ERPNext 状态：Paid
```

##### 6.2 集采付款/收款（走销售订单业务线）

**业务逻辑**：集采部分走销售订单业务线，总部收款，门店付款。

**总部视角（收款）**：
```
1. 根据 Delivery Note 生成 Sales Invoice（销售发票）
   ↓
2. 根据 Sales Invoice 生成收款单
   ↓
3. TTPOS 调用 ERPNext 创建 Payment Entry
   - 类型：Receive（收款）
   - 对方类型：Customer（客户，即门店）
   - 关联 Sales Invoice
   - 状态：Draft
   ↓
4. 审批收款单
   - ERPNext 状态：Submitted
   ↓
5. 完成收款
   - ERPNext 状态：Paid
```

**门店视角（付款）**：
```
1. 根据 Delivery Note 生成 Purchase Invoice（采购发票）
   ↓
2. 根据 Purchase Invoice 生成付款单
   ↓
3. TTPOS 调用 ERPNext 创建 Payment Entry
   - 类型：Pay（付款）
   - 对方类型：Supplier（供应商，即总部）
   - 关联 Purchase Invoice
   - 状态：Draft
   ↓
4. 审批付款单
   - ERPNext 状态：Submitted
   ↓
5. 完成付款
   - ERPNext 状态：Paid
```

## 🎯 关键实现要点

### 1. 按供应商拆分收货单

**实现逻辑**：

```go
// 在创建收货单时，按供应商分组
func groupItemsBySupplier(items []req.PurchaseReceiptItemReq) map[string][]req.PurchaseReceiptItemReq {
    supplierGroups := make(map[string][]req.PurchaseReceiptItemReq)
    
    for _, item := range items {
        // 从采购订单明细中获取供应商信息
        supplierCode := getSupplierCodeForItem(item.PurchaseOrderItemUuid)
        supplierGroups[supplierCode] = append(supplierGroups[supplierCode], item)
    }
    
    return supplierGroups
}

// 为每个供应商创建独立的收货单
func createReceiptOrdersBySupplier(ctx context.Context, groups map[string][]req.PurchaseReceiptItemReq) ([]uint64, error) {
    var receiptOrderUuids []uint64
    
    for supplierCode, items := range groups {
        receiptOrder, err := createReceiptOrderForSupplier(ctx, items, supplierCode)
        if err != nil {
            return nil, err
        }
        receiptOrderUuids = append(receiptOrderUuids, receiptOrder.Uuid)
    }
    
    return receiptOrderUuids, nil
}
```

### 2. 审批流程集成

**实现逻辑**：

**注意**：工作流优化后，审批流程由 ERPNext 工作流自动处理。提交订单时，系统根据金额自动进入对应审批状态：
- 金额 < 100,000：自动进入 Pending PMA（采购经理审批）
- 金额 ≥ 100,000：自动进入 Pending VP（VP 审批）

TTPOS 侧只需要同步 ERPNext 的工作流状态即可：

```go
// 提交采购订单时，ERPNext 工作流会自动根据金额判断审批路径
func (s *purchaseOrderSrv) SubmitPurchaseOrder(
    ctx context.Context,
    req req.PurchaseOrderSubmitReq,
) error {
    // ... 现有逻辑 ...
    
    // 调用 ERPNext API 提交订单
    // ERPNext 工作流会自动根据金额判断：
    // - 金额 < 100,000 → 进入 Pending PMA
    // - 金额 ≥ 100,000 → 进入 Pending VP
    err := s.erpClient.SubmitPurchaseOrder(ctx, purchaseOrder.ErpOrderNo)
    if err != nil {
        return err
    }
    
    // 同步 ERPNext 工作流状态
    err = s.syncWorkflowState(ctx, purchaseOrder)
    if err != nil {
        return err
    }
    
    return nil
}

// 同步工作流状态
func (s *purchaseOrderSrv) syncWorkflowState(
    ctx context.Context,
    purchaseOrder *model.PurchaseOrder,
) error {
    // 查询 ERPNext 工作流状态
    erpOrder, err := s.erpClient.GetPurchaseOrder(ctx, purchaseOrder.ErpOrderNo)
    if err != nil {
        return err
    }
    
    // 根据工作流状态更新 TTPOS 状态
    workflowState := erpOrder.WorkflowState
    switch workflowState {
    case "Pending PMA":
        purchaseOrder.Status = constant.PurchaseOrderStatusPending
        // 通知采购经理审批
    case "Pending VP":
        purchaseOrder.Status = constant.PurchaseOrderStatusPending
        // 通知 VP 审批
    case "Approved":
        purchaseOrder.Status = constant.PurchaseOrderStatusApproved
        // 审批通过，可以继续后续流程
    case "Rejected":
        purchaseOrder.Status = constant.PurchaseOrderStatusRejected
        // 审批被拒绝
    }
    
    return s.repo.Update(purchaseOrder)
}
```

### 3. ERPNext 状态同步

**实现逻辑**：

```go
// 同步 ERPNext 状态到 TTPOS
func (s *purchaseOrderSrv) syncErpStatus(ctx context.Context, purchaseOrder *model.PurchaseOrder) error {
    // 查询 ERPNext 采购订单状态
    erpOrder, err := s.erpClient.GetPurchaseOrder(ctx, purchaseOrder.ErpOrderNo)
    if err != nil {
        return err
    }
    
    // 更新 TTPOS 状态
    switch erpOrder.Status {
    case "Draft":
        purchaseOrder.Status = constant.PurchaseOrderStatusDraft
    case "Submitted":
        purchaseOrder.Status = constant.PurchaseOrderStatusPending
    case "Completed":
        purchaseOrder.Status = constant.PurchaseOrderStatusApproved
    }
    
    return s.repo.Update(purchaseOrder)
}
```

## 📊 数据库迁移脚本

### 1. 采购订单表调整

```sql
-- 添加多供应商支持字段
ALTER TABLE `ttpos_purchase_order`
ADD COLUMN `is_multi_supplier` TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否多供应商：0-否；1-是' AFTER `purchase_type`;

-- 创建采购订单供应商关联表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_order_supplier` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `purchase_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购订单UUID',
    `supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商编码',
    `supplier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_purchase_order_uuid` (`purchase_order_uuid`),
    INDEX `idx_supplier_erp_code` (`supplier_erp_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采购订单供应商关联表';

-- 在采购订单明细表中添加供应商字段
ALTER TABLE `ttpos_purchase_order_item`
ADD COLUMN `supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商编码（多供应商场景）' AFTER `base_erpnext_uom`,
ADD COLUMN `supplier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商名称（多供应商场景）' AFTER `supplier_erp_code`;
```

### 2. 采购退货单表

```sql
-- 创建采购退货单表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_return_order` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单UUID',
    `order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退货单号',
    `erp_order_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP退货单号',
    `purchase_receipt_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购收货单UUID',
    `purchase_receipt_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '采购收货单号',
    `supplier_erp_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商编码',
    `supplier_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '供应商名称',
    `status` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0-待提交 1-待审核 2-已通过 3-已驳回',
    `return_type` INT(10) UNSIGNED NOT NULL DEFAULT 1 COMMENT '退货类型：1-换货 2-退款',
    `process_type` INT(10) UNSIGNED NOT NULL DEFAULT 1 COMMENT '处理类型：1-直采退货 2-集采退货',
    `delivery_note_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Delivery Note UUID（集采）',
    `delivery_note_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Delivery Note号（集采）',
    `exchange_delivery_note_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '换货Delivery Note号（集采换货时使用）',
    `return_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '退货原因',
    `return_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货时间',
    `approver_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审批人UUID',
    `approver_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审批人姓名',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_purchase_receipt_uuid` (`purchase_receipt_uuid`),
    INDEX `idx_supplier_erp_code` (`supplier_erp_code`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采购退货单表';

-- 创建采购退货单明细表
CREATE TABLE IF NOT EXISTS `ttpos_purchase_return_order_item` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `return_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '退货单UUID',
    `purchase_receipt_item_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '采购收货单明细UUID',
    `material_code` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '物品编码',
    `material_name` TEXT NOT NULL COMMENT '物品名称JSON',
    `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品UUID',
    `return_num` DECIMAL(14,4) NOT NULL DEFAULT 0.0000 COMMENT '退货数量',
    `unit_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '单位UUID',
    `unit_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '单位名称',
    `valuation` DECIMAL(14,2) NOT NULL DEFAULT 0.00 COMMENT '估值单价',
    `total_price` DECIMAL(14,2) NOT NULL DEFAULT 0.00 COMMENT '总价',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_return_order_uuid` (`return_order_uuid`),
    INDEX `idx_purchase_receipt_item_uuid` (`purchase_receipt_item_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='采购退货单明细表';
```

### 3. 付款单/收款单表

```sql
-- 创建付款单表
CREATE TABLE IF NOT EXISTS `ttpos_payment_entry` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '付款单UUID',
    `entry_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '付款单号',
    `erp_entry_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'ERP付款单号',
    `payment_type` INT(10) UNSIGNED NOT NULL DEFAULT 1 COMMENT '付款类型：1-付款 2-收款',
    `party_type` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '对方类型：Supplier/Customer',
    `party_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '对方UUID',
    `party_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '对方名称',
    `paid_amount` DECIMAL(14,2) NOT NULL DEFAULT 0.00 COMMENT '付款金额',
    `status` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态：0-待提交 1-待审核 2-已通过 3-已驳回 4-已完成',
    `payment_date` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '付款日期',
    `approver_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审批人UUID',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_party_uuid` (`party_uuid`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='付款单表';

-- 创建付款单关联单据表
CREATE TABLE IF NOT EXISTS `ttpos_payment_entry_reference` (
    `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'UUID',
    `payment_entry_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '付款单UUID',
    `reference_type` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '关联单据类型',
    `reference_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联单据UUID',
    `reference_no` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '关联单据号',
    `allocated_amount` DECIMAL(14,2) NOT NULL DEFAULT 0.00 COMMENT '分配金额',
    `create_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
    INDEX `idx_payment_entry_uuid` (`payment_entry_uuid`),
    INDEX `idx_reference_uuid` (`reference_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='付款单关联单据表';
```

## ✅ 总结

### TTPOS 侧需要做的调整总结

1. **采购订单模型调整**
   - 支持多供应商采购订单
   - 在采购订单明细中添加供应商字段
   - 创建采购订单供应商关联表

2. **收货单按供应商拆分**
   - 修改创建收货单逻辑，按供应商分组
   - 为每个供应商创建独立的收货单
   - **直采**：调用 ERPNext 时为每个供应商创建独立的 Purchase Receipt
   - **集采**：
     - **总部仓库人员**：在总部系统中创建 Delivery Note（发货单），给到司机配送
     - **门店**：确认收货，可以针对 Delivery Note 进行退货

3. **品牌采购分类：集采和直采**
   - **集采部分（BOI）**：
     - 在 ERPNext 中**直接创建 Sales Order（Inter Company Sales Order）**，**不再创建 Purchase Order**
     - 总部集中采购后，通过销售订单配送给门店
     - 总部作为销售方，门店作为购买方
     - **总部仓库人员**创建 Delivery Note（发货单），给到司机配送；**门店**确认收货，可以针对 Delivery Note 进行退货
     - 退货时创建 Sales Invoice（Credit Note），而不是 Purchase Invoice（Debit Note）
     - 财务处理：总部创建 Sales Invoice 和收款单，门店创建 Purchase Invoice 和付款单
   - **直采部分（PO）**：
     - 在 ERPNext 中创建 Purchase Order
     - 外部供应商直接配送给门店
     - 外部供应商作为供应商，门店作为购买方
     - 收货时创建 Purchase Receipt
     - 退货时创建 Purchase Invoice（Debit Note）
     - 财务处理：门店创建 Purchase Invoice 和付款单

4. **新增采购退货功能**
   - 创建采购退货单模型和服务
   - **外部采购退货**：集成 ERPNext Purchase Invoice（Debit Note）
   - **内部采购退货**：集成 ERPNext Sales Invoice（Credit Note）
   - 支持换货和退款两种退货类型

5. **新增财务付款/收款功能**
   - 创建付款单/收款单模型和服务
   - 集成 ERPNext Payment Entry
   - 支持审批流程
   - **内部采购**：区分总部收款和门店付款两个视角

6. **ERPNext 集成增强**
   - 在 ttpos-bmp 中添加外部采购退货接口
   - 在 ttpos-bmp 中添加内部采购退货接口（Sales Invoice Credit Note）
   - 在 ttpos-bmp 中添加付款单/收款单接口
   - 支持按供应商拆分收货单的 ERPNext 调用
   - 支持内部采购创建 Sales Order 和 Delivery Note

### 关键实现要点

1. **按仓库和供应商拆分单据**：这是最关键的调整
   - **集采部分**：按**总部仓库**拆分 Delivery Note
     - 一个 Sales Order 包含多个仓库的物品时，按仓库创建多个 Delivery Note
     - 每个 Delivery Note 只包含来自同一仓库的物品
     - 方便门店人员根据不同的仓库发货进行收货
   - **直采部分**：按**外部供应商**拆分 Purchase Receipt
     - 一个 Purchase Order 包含多个供应商的物品时，按供应商创建多个 Purchase Receipt
     - 每个 Purchase Receipt 只包含来自同一供应商的物品
     - 方便门店人员根据不同的供应商发货进行收货

2. **拆分逻辑实现**：
   - 从 Sales Order 获取所有待发货物品，按 `warehouse` 字段（源仓库）分组
   - 从 Purchase Order 获取所有待收货物品，按 `supplier` 字段分组
   - 为每个分组创建独立的单据（Delivery Note 或 Purchase Receipt）

3. **审批流程集成**：根据金额判断审批人（采购经理或 VP）
4. **ERPNext 状态同步**：确保 TTPOS 和 ERPNext 状态一致
5. **多仓库和多供应商支持**：一个订单可以包含多个仓库或供应商的物品

### 下一步行动

1. 创建数据库迁移脚本
2. 实现采购订单多供应商支持
3. **实现集采 Delivery Note 按仓库拆分逻辑**（新增）
4. **实现直采 Purchase Receipt 按供应商拆分逻辑**（已有）
5. 实现采购退货功能
6. 实现财务付款/收款功能
7. 在 ttpos-bmp 中添加相应的 ERPNext 接口
8. 编写单元测试和集成测试

