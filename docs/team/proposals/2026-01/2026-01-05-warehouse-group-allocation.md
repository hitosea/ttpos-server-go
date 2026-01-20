# 仓库组（母仓库）子仓库分配方案

> 本文档说明如何在 ERPNext 中实现"母仓库=仓库组，子仓库=真实库存，门店只选母仓库，总部决定从哪个子仓库发货"的业务需求。

---

## 📋 业务需求

### 核心需求

1. **仓库层级结构**
   - **母仓库**：仓库组（`is_group = 1`），类似文件夹名称，不存放实物
   - **子仓库**：真实仓库（`is_group = 0`），存放实物，有实际库存

2. **门店操作**
   - 门店创建 **Material Request（材料申请单）** 时，只选择**母仓库（仓库组）**
   - 门店不需要知道具体从哪个子仓库发货
   - MR 审批后，系统自动创建 Sales Order（集采订单）时，继承母仓库信息

3. **总部操作**
   - 总部审批 Material Request
   - 总部在 ERPNext 上能看到从 MR 创建的 Sales Order（集采订单）
   - 总部需要**选择具体子仓库**并正常发货
   - 总部需要根据**分配策略**（先进先出、优先级仓库、批次效期）自动或手动分配子仓库

4. **分配策略**
   - 先进先出（FIFO）
   - 优先级仓库
   - 批次效期
   - 规则可能是单一，也可能是组合（且的关系）

---

## 🎯 ERPNext 实现方案

### 方案概述

**核心思路**：利用 ERPNext 的**仓库组（Warehouse Group）**和**自定义字段/脚本**实现。

**关键点**：
- ERPNext 的 `Warehouse` 支持 `is_group = 1`（仓库组）和 `is_group = 0`（真实仓库）
- **库存台账只能发生在真实仓库（`is_group = 0`）**
- Material Request 和 Sales Order 的 `warehouse` 字段**不能直接选择仓库组**（ERPNext 限制）
- 需要通过**自定义字段**或**脚本**实现：
  - Material Request 阶段：门店选择母仓库（仓库组）
  - Sales Order 阶段：从 MR 继承母仓库，总部分配子仓库

---

## 📐 方案 A：自定义字段 + 脚本分配（推荐）

### 1. 仓库结构设置

#### 1.1 在 ERPNext 中创建仓库层级

```
母仓库（仓库组）
├── 子仓库-01（真实仓库）
├── 子仓库-02（真实仓库）
└── 子仓库-03（真实仓库）
```

**ERPNext 操作步骤**：

1. **创建母仓库（仓库组）**
   - Stock → Warehouse → New
   - Warehouse Name: `母仓库-总部`
   - **勾选 `Is Group`**（关键！）
   - Company: 总部公司
   - 保存

2. **创建子仓库（真实仓库）**
   - Stock → Warehouse → New
   - Warehouse Name: `母仓库-总部-01`
   - **不勾选 `Is Group`**（默认）
   - Parent Warehouse: `母仓库-总部`（选择母仓库作为父级）
   - Company: 总部公司
   - 保存

3. **重复创建其他子仓库**
   - `母仓库-总部-02`
   - `母仓库-总部-03`
   - 都设置 `Parent Warehouse = 母仓库-总部`

---

### 2. Material Request 字段设计

#### 2.1 添加自定义字段

**在 Material Request 单据类型中添加自定义字段**：

1. **Customize Form** → Material Request
2. **添加字段**：

| 字段名 | 字段类型 | 标签 | 说明 |
|--------|---------|------|------|
| `warehouse_group` | Link → Warehouse | 仓库组（母仓库） | 门店选择，过滤条件：`is_group = 1` |

**字段配置**：

```python
# warehouse_group 字段配置
{
    "fieldname": "warehouse_group",
    "fieldtype": "Link",
    "label": "仓库组（母仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 1]
    ],
    "reqd": 1,  # 门店必填
    "read_only": 0
}
```

**Material Request Item 字段配置**：

```python
# 在 Material Request Item 子表中，warehouse 字段可以留空或设置为默认值
# 实际仓库信息从父单据的 warehouse_group 继承
```

### 3. Sales Order 字段设计

#### 3.1 添加自定义字段

**在 Sales Order 单据类型中添加自定义字段**：

1. **Customize Form** → Sales Order
2. **添加字段**：

| 字段名 | 字段类型 | 标签 | 说明 |
|--------|---------|------|------|
| `warehouse_group` | Link → Warehouse | 仓库组（母仓库） | 门店选择，过滤条件：`is_group = 1` |
| `source_warehouse` | Link → Warehouse | 发货仓库（子仓库） | 总部选择，过滤条件：`is_group = 0` 且 `parent_warehouse = warehouse_group` |
| `warehouse_allocation_strategy` | Select | 分配策略 | 可选值：FIFO、优先级仓库、批次效期、组合策略 |

**字段配置**：

```python
# warehouse_group 字段配置
{
    "fieldname": "warehouse_group",
    "fieldtype": "Link",
    "label": "仓库组（母仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 1]
    ],
    "reqd": 1,  # 门店必填
    "read_only": 0
}

# source_warehouse 字段配置
{
    "fieldname": "source_warehouse",
    "fieldtype": "Link",
    "label": "发货仓库（子仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 0],
        ["parent_warehouse", "=", "warehouse_group"]  # 动态过滤
    ],
    "reqd": 0,  # 总部填写
    "read_only": 0
}
```

---

### 4. 门店创建 Material Request 流程

#### 4.1 门店操作步骤

1. **门店创建 Material Request**
   - Material Request Type: `Purchase`
   - Transaction Date: 当前日期
   - **`warehouse_group`**: 选择"母仓库-总部"（仓库组）
   - Items: 添加物品和数量
   - 保存并提交

2. **关键点**：
   - 门店**只填写 `warehouse_group`**（母仓库）
   - Material Request Item 的 `warehouse` 字段可以留空或设置为默认值
   - 门店不需要知道具体从哪个子仓库发货

---

### 5. 从 Material Request 创建 Sales Order 流程

#### 5.1 自动创建 Sales Order

**触发条件**：Material Request 审批后，物品**未勾选默认供应商**（或默认供应商是总部）时，**自动创建** Sales Order（集采订单）。

**创建流程**：

1. **MR 审批通过**
   - Material Request 状态：Submitted

2. **系统检测到物品未勾选默认供应商（或默认供应商是总部）**

3. **自动创建 Sales Order**
   - 从 Material Request 创建 Sales Order
   - **继承 `warehouse_group`**（从 MR 的 `warehouse_group` 字段继承）
   - Customer: 门店（作为客户）
   - Company: 总部（作为销售方）
   - 状态：Draft

4. **关键点**：
   - Sales Order 创建时，**自动继承 MR 的 `warehouse_group`**
   - Sales Order 的 `warehouse` 字段（标准字段）可以留空或设置为默认值
   - 等待总部分配子仓库

#### 5.2 从 Material Request 创建 Sales Order 的脚本实现

**实现方式**：在创建 Sales Order 时，通过 Server Script 或 API 调用继承 MR 的 `warehouse_group`。

**方案 1：使用 ERPNext Server Script（推荐）**

```python
# ERPNext Server Script
# DocType: Sales Order
# Script Type: Before Insert / Before Save

import frappe

def set_warehouse_group_from_mr(doc):
    """
    从 Material Request 继承 warehouse_group
    """
    if doc.material_request and not doc.warehouse_group:
        # 从 Material Request 获取 warehouse_group
        mr = frappe.get_doc("Material Request", doc.material_request)
        if mr.warehouse_group:
            doc.warehouse_group = mr.warehouse_group
            frappe.msgprint(f"已从 Material Request 继承仓库组：{mr.warehouse_group}")
```

**方案 2：在创建 Sales Order 的 API 调用中设置**

```python
# TTPOS 侧调用 ERPNext API 创建 Sales Order 时
# 从 Material Request 获取 warehouse_group 并传递

def create_sales_order_from_mr(mr_name):
    """
    从 Material Request 创建 Sales Order
    """
    # 获取 Material Request
    mr = get_material_request(mr_name)
    
    # 创建 Sales Order
    sales_order = {
        "doctype": "Sales Order",
        "customer": store_branch,  # 门店作为客户
        "company": headquarters_company,  # 总部作为销售方
        "material_request": mr_name,
        "warehouse_group": mr.warehouse_group,  # 继承 warehouse_group
        "items": []
    }
    
    # 添加 Items
    for item in mr.items:
        if not item.supplier or item.supplier == headquarters_supplier:
            # 只处理集采物品（未勾选默认供应商或默认供应商是总部）
            sales_order["items"].append({
                "item_code": item.item_code,
                "qty": item.qty,
                "uom": item.uom,
                "warehouse": "",  # 留空，等待总部分配
                # warehouse_group 会从父单据继承
            })
    
    # 调用 ERPNext API 创建
    create_sales_order(sales_order)
```

---

### 6. 总部分配子仓库流程

#### 6.1 手动分配（方案 A-1）

**总部操作步骤**：

1. **总部打开 Sales Order**
   - 查看 `warehouse_group`（门店已选择）
   - 查看 Items 列表

2. **为每个 Item 选择子仓库**
   - 在 Sales Order Item 行上添加自定义字段 `source_warehouse_item`
   - 总部人员手动选择子仓库（从 `warehouse_group` 的子仓库中选择）

3. **保存并提交 Sales Order**

**Sales Order Item 自定义字段**：

```python
# 在 Sales Order Item 子表添加字段
{
    "fieldname": "source_warehouse_item",
    "fieldtype": "Link",
    "label": "发货仓库（子仓库）",
    "options": "Warehouse",
    "filters": [
        ["is_group", "=", 0],
        ["parent_warehouse", "=", "warehouse_group"]  # 从父单据获取
    ],
    "reqd": 0
}
```

---

#### 6.2 自动分配（方案 A-2，推荐）

**使用 ERPNext Script 实现自动分配策略**。

**Script 位置**：Sales Order → Client Script 或 Server Script

**分配策略实现**：

```python
# ERPNext Server Script
# DocType: Sales Order
# Script Type: Before Save

import frappe
from frappe.utils import flt

def allocate_warehouse_by_strategy(doc):
    """
    根据分配策略自动分配子仓库
    """
    if not doc.warehouse_group:
        return
    
    # 获取母仓库下的所有子仓库
    child_warehouses = frappe.get_all(
        "Warehouse",
        filters={
            "parent_warehouse": doc.warehouse_group,
            "is_group": 0
        },
        fields=["name", "warehouse_name"]
    )
    
    if not child_warehouses:
        frappe.throw(f"仓库组 {doc.warehouse_group} 下没有子仓库")
    
    # 获取分配策略
    strategy = doc.warehouse_allocation_strategy or "FIFO"
    
    # 为每个 Item 分配子仓库
    for item in doc.items:
        if item.source_warehouse_item:
            # 如果已手动分配，跳过
            continue
        
        # 根据策略分配
        allocated_warehouse = allocate_by_strategy(
            item.item_code,
            item.qty,
            child_warehouses,
            strategy,
            doc.company
        )
        
        if allocated_warehouse:
            item.source_warehouse_item = allocated_warehouse
            item.warehouse = allocated_warehouse  # 更新标准字段


def allocate_by_strategy(item_code, qty, warehouses, strategy, company):
    """
    根据策略分配仓库
    """
    if strategy == "FIFO":
        return allocate_by_fifo(item_code, qty, warehouses, company)
    elif strategy == "优先级仓库":
        return allocate_by_priority(item_code, qty, warehouses, company)
    elif strategy == "批次效期":
        return allocate_by_batch_expiry(item_code, qty, warehouses, company)
    elif strategy == "组合策略":
        return allocate_by_combined(item_code, qty, warehouses, company)
    else:
        # 默认 FIFO
        return allocate_by_fifo(item_code, qty, warehouses, company)


def allocate_by_fifo(item_code, qty, warehouses, company):
    """
    先进先出策略：选择最早入库的批次所在的仓库
    """
    from frappe.utils import getdate
    
    # 查询每个子仓库的可用库存和最早批次
    warehouse_stock = []
    
    for wh in warehouses:
        # 查询该仓库的库存
        stock_qty = frappe.db.sql("""
            SELECT SUM(actual_qty) as qty
            FROM `tabBin`
            WHERE item_code = %s
            AND warehouse = %s
        """, (item_code, wh.name), as_dict=True)
        
        available_qty = stock_qty[0].qty if stock_qty else 0
        
        if available_qty >= qty:
            # 查询最早批次
            earliest_batch = frappe.db.sql("""
                SELECT sle.batch_no, MIN(sle.posting_date) as earliest_date
                FROM `tabStock Ledger Entry` sle
                INNER JOIN `tabBatch` b ON sle.batch_no = b.name
                WHERE sle.item_code = %s
                AND sle.warehouse = %s
                AND sle.actual_qty > 0
                GROUP BY sle.batch_no
                ORDER BY earliest_date ASC
                LIMIT 1
            """, (item_code, wh.name), as_dict=True)
            
            if earliest_batch:
                warehouse_stock.append({
                    "warehouse": wh.name,
                    "qty": available_qty,
                    "earliest_date": earliest_batch[0].earliest_date
                })
    
    if not warehouse_stock:
        return None
    
    # 按最早日期排序，选择最早的
    warehouse_stock.sort(key=lambda x: x["earliest_date"])
    return warehouse_stock[0]["warehouse"]


def allocate_by_priority(item_code, qty, warehouses, company):
    """
    优先级仓库策略：按仓库优先级分配
    """
    # 假设在 Warehouse 自定义字段中添加了 priority 字段
    # 或者在 Warehouse 表中添加 priority 字段
    
    warehouse_priorities = []
    
    for wh in warehouses:
        # 查询优先级（需要自定义字段）
        priority = frappe.db.get_value("Warehouse", wh.name, "priority") or 999
        
        # 查询可用库存
        stock_qty = frappe.db.sql("""
            SELECT SUM(actual_qty) as qty
            FROM `tabBin`
            WHERE item_code = %s
            AND warehouse = %s
        """, (item_code, wh.name), as_dict=True)
        
        available_qty = stock_qty[0].qty if stock_qty else 0
        
        if available_qty >= qty:
            warehouse_priorities.append({
                "warehouse": wh.name,
                "priority": priority,
                "qty": available_qty
            })
    
    if not warehouse_priorities:
        return None
    
    # 按优先级排序（数字越小优先级越高）
    warehouse_priorities.sort(key=lambda x: x["priority"])
    return warehouse_priorities[0]["warehouse"]


def allocate_by_batch_expiry(item_code, qty, warehouses, company):
    """
    批次效期策略：优先选择即将过期的批次所在的仓库
    """
    warehouse_expiry = []
    
    for wh in warehouses:
        # 查询该仓库的批次和效期
        batches = frappe.db.sql("""
            SELECT 
                sle.batch_no,
                b.expiry_date,
                SUM(sle.actual_qty) as qty
            FROM `tabStock Ledger Entry` sle
            INNER JOIN `tabBatch` b ON sle.batch_no = b.name
            WHERE sle.item_code = %s
            AND sle.warehouse = %s
            AND sle.actual_qty > 0
            AND b.expiry_date IS NOT NULL
            GROUP BY sle.batch_no
            ORDER BY b.expiry_date ASC
        """, (item_code, wh.name), as_dict=True)
        
        if batches:
            # 计算总可用库存
            total_qty = sum(b["qty"] for b in batches)
            
            if total_qty >= qty:
                # 选择最早过期的批次
                earliest_expiry = batches[0]["expiry_date"]
                warehouse_expiry.append({
                    "warehouse": wh.name,
                    "expiry_date": earliest_expiry,
                    "qty": total_qty
                })
    
    if not warehouse_expiry:
        return None
    
    # 按效期排序，选择最早过期的
    warehouse_expiry.sort(key=lambda x: x["expiry_date"])
    return warehouse_expiry[0]["warehouse"]


def allocate_by_combined(item_code, qty, warehouses, company):
    """
    组合策略：优先级 + 批次效期（且的关系）
    1. 先按优先级筛选
    2. 在优先级相同的仓库中，选择最早过期的批次
    """
    warehouse_candidates = []
    
    for wh in warehouses:
        # 查询优先级
        priority = frappe.db.get_value("Warehouse", wh.name, "priority") or 999
        
        # 查询批次和效期
        batches = frappe.db.sql("""
            SELECT 
                sle.batch_no,
                b.expiry_date,
                SUM(sle.actual_qty) as qty
            FROM `tabStock Ledger Entry` sle
            INNER JOIN `tabBatch` b ON sle.batch_no = b.name
            WHERE sle.item_code = %s
            AND sle.warehouse = %s
            AND sle.actual_qty > 0
            AND b.expiry_date IS NOT NULL
            GROUP BY sle.batch_no
            ORDER BY b.expiry_date ASC
        """, (item_code, wh.name), as_dict=True)
        
        if batches:
            total_qty = sum(b["qty"] for b in batches)
            
            if total_qty >= qty:
                earliest_expiry = batches[0]["expiry_date"]
                warehouse_candidates.append({
                    "warehouse": wh.name,
                    "priority": priority,
                    "expiry_date": earliest_expiry,
                    "qty": total_qty
                })
    
    if not warehouse_candidates:
        return None
    
    # 先按优先级排序，再按效期排序
    warehouse_candidates.sort(key=lambda x: (x["priority"], x["expiry_date"]))
    return warehouse_candidates[0]["warehouse"]
```

---

### 7. Delivery Note 创建和发货

#### 7.1 从 Sales Order 创建 Delivery Note

**总部操作步骤**：

1. **打开 Sales Order**
   - 确认 `source_warehouse_item` 已分配（自动或手动）

2. **创建 Delivery Note**
   - 点击 "Create" → "Delivery Note"
   - ERPNext 会自动从 Sales Order 创建 Delivery Note

3. **关键点**：
   - Delivery Note 的 `warehouse` 字段会从 Sales Order Item 的 `source_warehouse_item` 或 `warehouse` 字段继承
   - 如果 Sales Order Item 的 `warehouse` 字段已设置为子仓库，Delivery Note 会自动使用该子仓库

4. **提交 Delivery Note**
   - 提交后，ERPNext 会从指定的子仓库扣减库存

---

### 8. 跨公司（Inter-Company）处理

#### 8.1 跨公司 Sales Order 配置

**关键配置**：

1. **Company 设置**
   - 总部公司：`Headquarters - Company`
   - 门店公司：`Store Branch - Company`

2. **Customer 设置**
   - 门店作为总部的客户
   - Customer Type: `Company`
   - Company: `Store Branch - Company`

3. **Sales Order 创建**
   - Company: `Headquarters - Company`（总部）
   - Customer: `Store Branch - Company`（门店）
   - Transaction Type: `Internal`（内部交易）

4. **Delivery Note 创建**
   - Source Warehouse: 子仓库（总部）
   - Target Warehouse: 门店仓库（门店）

---

## 📐 方案 B：Workflow + 审批分配（备选）

如果方案 A 不满足需求，可以考虑使用 **Workflow** 实现审批分配流程。

### 流程设计

```
1. 门店创建 Material Request（只选母仓库）
   ↓
2. Material Request 审批通过
   ↓
3. 系统自动创建 Sales Order（继承母仓库）
   ↓
4. Sales Order 状态：Draft，触发 Workflow：进入"待分配仓库"状态
   ↓
5. 总部人员分配子仓库（手动或自动）
   ↓
6. 审批通过，状态变为 Submitted
   ↓
7. 创建 Delivery Note 并发货
```

**Workflow 配置**：

1. **Workflow State**：
   - Draft
   - Pending Warehouse Allocation（待分配仓库）
   - Submitted

2. **Workflow Action**：
   - Allocate Warehouse（分配仓库）
   - Submit（提交）

3. **Workflow Transition**：
   - Draft → Pending Warehouse Allocation（自动）
   - Pending Warehouse Allocation → Submitted（需要分配仓库后）

---

## 🔧 技术实现细节

### 1. 仓库优先级字段

**在 Warehouse 中添加自定义字段**：

```python
{
    "fieldname": "priority",
    "fieldtype": "Int",
    "label": "优先级",
    "description": "数字越小优先级越高，用于仓库分配策略",
    "default": 999
}
```

### 2. 分配策略配置

**在 Sales Order 或 Company 中添加分配策略配置**：

```python
# Sales Order 自定义字段
{
    "fieldname": "warehouse_allocation_strategy",
    "fieldtype": "Select",
    "label": "分配策略",
    "options": "\nFIFO\n优先级仓库\n批次效期\n组合策略",
    "default": "FIFO"
}
```

### 3. 批量分配按钮

**在 Sales Order 表单添加"自动分配仓库"按钮**：

```python
# Client Script
frappe.ui.form.on('Sales Order', {
    refresh: function(frm) {
        if (frm.doc.warehouse_group && frm.doc.docstatus === 0) {
            frm.add_custom_button(__('自动分配仓库'), function() {
                // 调用 Server Script 自动分配
                frappe.call({
                    method: 'your_app.warehouse_allocation.auto_allocate_warehouse',
                    args: {
                        sales_order: frm.doc.name
                    },
                    callback: function(r) {
                        if (r.message) {
                            frappe.show_alert({
                                message: __('仓库分配完成'),
                                indicator: 'green'
                            });
                            frm.reload_doc();
                        }
                    }
                });
            });
        }
    }
});
```

---

## ✅ 验收标准

### 功能验收

1. **门店操作**
   - ✅ 门店创建 Material Request 时，可以选择母仓库（仓库组）
   - ✅ 门店不需要选择子仓库
   - ✅ Material Request 可以正常保存和提交
   - ✅ 从 Material Request 创建 Sales Order 时，自动继承母仓库信息

2. **总部操作**
   - ✅ 总部审批 Material Request
   - ✅ 总部人员可以在 ERPNext 上看到从 MR 创建的 Sales Order
   - ✅ 总部可以手动选择子仓库
   - ✅ 总部可以使用自动分配策略分配子仓库
   - ✅ 分配策略支持：FIFO、优先级仓库、批次效期、组合策略

3. **发货流程**
   - ✅ 从 Sales Order 创建 Delivery Note 时，自动使用分配的子仓库
   - ✅ Delivery Note 提交后，从正确的子仓库扣减库存
   - ✅ 跨公司发货流程正常

4. **数据一致性**
   - ✅ 库存扣减发生在正确的子仓库
   - ✅ 库存台账记录正确
   - ✅ 跨公司对账数据正确

---

## 📝 注意事项

### ERPNext 限制

1. **仓库组不能直接用于库存记账**
   - ERPNext 的 `Bin` 表（库存台账）只能关联真实仓库（`is_group = 0`）
   - Material Request 和 Sales Order 的 `warehouse` 字段不能选择仓库组
   - 需要通过自定义字段和脚本实现：
     - Material Request 阶段：使用自定义字段 `warehouse_group` 选择母仓库
     - Sales Order 阶段：从 MR 继承 `warehouse_group`，总部分配子仓库

2. **Delivery Note 仓库继承**
   - Delivery Note 的 `warehouse` 字段会从 Sales Order Item 的 `warehouse` 字段继承
   - 需要确保 Sales Order Item 的 `warehouse` 字段已设置为子仓库

3. **跨公司配置**
   - 需要正确配置 Inter-Company 关系
   - 需要配置 Customer 和 Supplier 的 Inter-Company 关系

### 性能考虑

1. **分配策略查询**
   - 如果子仓库数量多，分配策略查询可能较慢
   - 建议添加索引：`tabBin(item_code, warehouse)`、`tabStock Ledger Entry(item_code, warehouse, batch_no)`

2. **批量分配**
   - 如果 Sales Order 的 Items 数量多，批量分配可能较慢
   - 建议使用后台任务或异步处理

---

## 🔗 相关文档

- ERPNext Warehouse 文档: https://docs.erpnext.com/docs/user/manual/en/stock/warehouse
- ERPNext Customization 文档: https://docs.erpnext.com/docs/user/manual/en/customize-erpnext
- ERPNext Scripting 文档: https://docs.erpnext.com/docs/user/manual/en/scripting

---

**版本**: v1.0.0  
**创建日期**: 2026-01-05  
**维护者**: TTPOS Team

