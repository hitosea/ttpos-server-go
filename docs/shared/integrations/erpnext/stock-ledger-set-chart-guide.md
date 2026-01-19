# ERPNext Stock Ledger Set Chart 使用指南

> 详细说明如何在 ERPNext 中使用 Set Chart（Dashboard Chart）展示库存汇总表数据

---

## 📋 快速参考

### 创建 Set Chart 的两种方式

| 方式 | 创建位置 | 操作步骤 | 推荐度 |
|------|---------|---------|--------|
| **方式一** | Stock Ledger 报表页面 | 打开报表 → 设置筛选条件 → 创建图表 | ⭐⭐⭐⭐⭐ |
| **方式二** | Dashboard 页面 | 打开 Dashboard → 添加图表 → 选择 Stock Ledger | ⭐⭐⭐ |

### 关键字段映射

| 库存汇总表字段 | Stock Ledger 字段 | 说明 |
|---------------|------------------|------|
| `opening_qty` | 计算得出 | `theoretical_qty + sales_qty` |
| `sales_qty` | `out_qty`（累计） | 筛选 `voucher_type` = `Sales Invoice` |
| `theoretical_qty` | `qty_after_transaction` | 交易后数量 |
| `actual_qty` | `actual_qty` | 筛选 `voucher_type` = `Stock Reconciliation` |
| `diff_qty` | 计算得出 | `actual_qty - theoretical_qty` |

### 常用筛选条件 JSON

```json
{
  "company": "Company A",
  "from_date": "2025-01-16",
  "to_date": "2025-01-16",
  "warehouse": "WH-001",
  "item_group": "食材"
}
```

---

## 一、概述

### 1.1 什么是 Set Chart

**Set Chart（Dashboard Chart）** 是 ERPNext 中的图表功能，可以在 Dashboard（仪表盘）中创建可视化图表，用于展示报表数据。

### 1.2 库存汇总表字段映射

| 库存汇总表字段 | Stock Ledger 字段 | 说明 |
|---------------|------------------|------|
| `item_code` | `item_code` | 物料编码 |
| `item_name` | `item_name` | 物料名称 |
| `item_group` | `item_group` | 物料分组 |
| `opening_qty` | 计算得出 | 初始库存数量 = 理论库存 + 累计销量 |
| `sales_qty` | `out_qty`（累计） | 累计销量（销售出库数量） |
| `theoretical_qty` | `qty_after_transaction` | 理论库存数量（交易后数量） |
| `actual_qty` | `actual_qty` | 实际库存数量（盘点单中的实盘数量） |
| `diff_qty` | 计算得出 | 差异数量 = 实际库存 - 理论库存 |
| `stock_uom` | `stock_uom` | 物料单位 |

---

## 二、创建 Set Chart 的步骤

### 2.1 Set Chart 的位置说明

**Set Chart** 是 ERPNext 中基于报表数据创建图表的功能，有两种创建方式：

1. **方式一：在 Stock Ledger 报表页面中直接创建**（推荐）
   - 打开 Stock Ledger 报表后，在报表页面中直接创建图表
   - 图表会自动关联当前报表的筛选条件

2. **方式二：在 Dashboard（仪表盘）中创建**
   - 在 Dashboard 中创建图表，选择 Stock Ledger 作为数据源
   - 需要手动配置筛选条件

**推荐使用方式一**，因为更直观且操作更简单。

---

### 2.2 方式一：在 Stock Ledger 报表页面中创建 Set Chart（推荐）

#### Step 1: 打开 Stock Ledger 报表

1. 登录 ERPNext 系统
2. 导航至：**主页 > 库存（Stock） > 报表（Reports） > Stock Ledger（库存分类账）**
   - 或者直接搜索：`Stock Ledger`
   - 或者直接访问：`/app/query-report/Stock Ledger`

#### Step 2: 设置报表筛选条件

在 Stock Ledger 报表页面顶部，设置筛选条件：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| **Company（公司）** | 选择公司名称（必填） | `Company A` |
| **From Date（开始日期）** | 选择开始日期 | `2025-01-16` |
| **To Date（结束日期）** | 选择结束日期 | `2025-01-16`（与开始日期相同，查询单日数据） |
| **Warehouse（仓库）** | 选择仓库（可选） | `WH-001` |
| **Item Group（物料分组）** | 选择物料分组（可选） | `食材` |
| **Item Code（物料编码）** | 输入物料编码（可选） | `F001` |

3. 点击 **"Update（更新）"** 或 **"刷新"** 按钮，查看报表数据

#### Step 3: 创建 Set Chart

在 Stock Ledger 报表页面中：

1. **找到图表区域**：
   - 报表页面通常有一个 **"Chart（图表）"** 或 **"Visualization（可视化）"** 标签页
   - 或者点击报表右上角的 **"Chart"** 按钮
   - 或者点击报表工具栏中的 **"Show Chart（显示图表）"** 图标

2. **如果图表区域不存在，创建新图表**：
   - 点击报表页面右上角的 **"Menu（菜单）"** 按钮（三个点图标）
   - 选择 **"Add Chart（添加图表）"** 或 **"Create Chart（创建图表）"**
   - 或者导航至：**主页 > 设置 > 图表（Charts） > 新建（New）**

#### Step 4: 配置图表参数

在图表配置界面中：

1. **图表名称（Chart Name）**：
   - 输入图表名称，例如：`库存对比图表`

2. **数据源（Data Source）**：
   - **报表类型（Report Type）**：选择 `Report`
   - **报表名称（Report Name）**：选择 `Stock Ledger`
   - **筛选条件（Filters）**：会自动继承当前报表的筛选条件

3. **图表类型（Chart Type）**：
   - 选择图表类型，例如：`Bar`（柱状图）、`Line`（折线图）、`Pie`（饼图）等

4. **X 轴配置（X Axis）**：
   - **字段（Field）**：选择 `item_name`（物料名称）或 `item_code`（物料编码）
   - **标签（Label）**：输入 X 轴标签，例如：`物料名称`

5. **Y 轴配置（Y Axis）**：
   - **字段（Field）**：选择数值字段，例如：`qty_after_transaction`（理论库存）
   - **标签（Label）**：输入 Y 轴标签，例如：`库存数量`
   - **聚合方式（Aggregation）**：选择 `Sum`（求和）或 `Average`（平均值）

6. **系列配置（Series）**（如果需要多个系列）：
   - 点击 **"Add Series（添加系列）"** 按钮
   - 添加多个系列，例如：
     - 系列 1：`qty_after_transaction`（理论库存）
     - 系列 2：`actual_qty`（实际库存）

7. **图表标题（Chart Title）**：
   - 输入图表标题，例如：`库存对比（理论 vs 实际）`

8. **其他配置**：
   - **颜色（Colors）**：自定义图表颜色
   - **图例（Legend）**：显示/隐藏图例
   - **工具提示（Tooltip）**：配置鼠标悬停提示

9. 点击 **"Save（保存）"** 按钮保存图表

#### Step 5: 查看图表

保存后，图表会显示在 Stock Ledger 报表页面的图表区域中。

---

### 2.3 方式二：在 Dashboard（仪表盘）中创建 Set Chart

#### Step 1: 访问 Dashboard

1. 登录 ERPNext 系统
2. 导航至：**主页 > Dashboard（仪表盘）**
   - 或者直接访问：`/app/dashboard`

#### Step 2: 创建或编辑 Dashboard

1. **如果 Dashboard 不存在**：
   - 点击 **"New Dashboard（新建仪表盘）"** 按钮
   - 输入 Dashboard 名称，例如：`库存汇总仪表盘`
   - 点击 **"Save（保存）"**

2. **如果 Dashboard 已存在**：
   - 选择要编辑的 Dashboard
   - 点击右上角的 **"Edit（编辑）"** 或 **"Customize（自定义）"** 按钮

#### Step 3: 添加图表

1. 在 Dashboard 编辑模式下，点击 **"Add Chart（添加图表）"** 或 **"Add Widget（添加组件）"** 按钮
2. 选择 **"Chart（图表）"** 类型

#### Step 4: 配置图表数据源

在图表配置界面中：

1. **报表类型（Report Type）**：选择 `Report`
2. **报表名称（Report Name）**：选择 `Stock Ledger`
3. **筛选条件（Filters）**：配置筛选条件

**筛选条件配置方式**：

**方式 A：使用 JSON 格式**（推荐）
```json
{
  "company": "Company A",
  "from_date": "2025-01-16",
  "to_date": "2025-01-16",
  "warehouse": "WH-001",
  "item_group": "食材"
}
```

**方式 B：使用筛选器界面**
- 在筛选条件区域，逐个添加筛选条件：
  - `company` = `Company A`
  - `from_date` = `2025-01-16`
  - `to_date` = `2025-01-16`
  - `warehouse` = `WH-001`（可选）
  - `item_group` = `食材`（可选）

#### Step 5: 配置图表参数

参考 **2.2 方式一** 的 Step 4 配置图表参数。

#### Step 6: 保存并查看

1. 点击 **"Save（保存）"** 按钮保存图表
2. 图表会显示在 Dashboard 中
3. 可以拖拽调整图表位置和大小

---

### 2.4 两种方式对比

| 对比项 | 方式一（报表页面） | 方式二（Dashboard） |
|--------|------------------|-------------------|
| **创建位置** | Stock Ledger 报表页面 | Dashboard 页面 |
| **筛选条件** | 自动继承报表筛选条件 | 需要手动配置 |
| **操作复杂度** | ⭐⭐ 简单 | ⭐⭐⭐ 中等 |
| **适用场景** | 查看报表时快速创建图表 | 在 Dashboard 中集中展示多个图表 |
| **推荐度** | ⭐⭐⭐⭐⭐ 推荐 | ⭐⭐⭐ 可选 |

**建议**：优先使用方式一，如果需要多个图表组合展示，再使用方式二。

---

## 三、配置不同类型的图表

### 3.1 图表 1：库存对比柱状图（推荐）

**用途**：展示各物料的初始库存、理论库存、实际库存对比

#### 详细配置步骤

**前提条件**：已完成 **二、创建 Set Chart 的步骤**，进入图表配置界面。

**配置项说明**：

1. **基本信息**：
   - **图表名称（Chart Name）**：`库存对比图表`
   - **图表标题（Chart Title）**：`库存对比（理论 vs 实际）`

2. **数据源配置**：
   - **报表类型（Report Type）**：`Report`
   - **报表名称（Report Name）**：`Stock Ledger`
   - **筛选条件（Filters）**：
     ```json
     {
       "company": "Company A",
       "from_date": "2025-01-16",
       "to_date": "2025-01-16",
       "warehouse": "WH-001"
     }
     ```

3. **图表类型（Chart Type）**：
   - 选择：`Bar`（柱状图）
   - 子类型：`Grouped Bar`（分组柱状图）或 `Stacked Bar`（堆叠柱状图）

4. **X 轴配置（X Axis）**：
   - **字段（Field）**：`item_name`（物料名称）
   - **标签（Label）**：`物料名称`
   - **排序（Sort）**：`Ascending`（升序）或 `Descending`（降序）

5. **Y 轴配置（Y Axis）**：
   - **字段（Field）**：`qty_after_transaction`（理论库存）
   - **标签（Label）**：`库存数量`
   - **聚合方式（Aggregation）**：`Sum`（求和）
   - **单位（Unit）**：`个`（可选）

6. **系列配置（Series）**（多系列图表）：
   - **系列 1**：
     - **名称（Name）**：`理论库存`
     - **字段（Field）**：`qty_after_transaction`
     - **颜色（Color）**：`#1890ff`（蓝色）
   - **系列 2**：
     - **名称（Name）**：`实际库存`
     - **字段（Field）**：`actual_qty`
     - **颜色（Color）**：`#52c41a`（绿色）

7. **筛选条件（Filters）**（在图表层面进一步筛选）：
   - `voucher_type` = `Stock Reconciliation`（仅显示盘点单数据）
   - 或者：不设置，显示所有数据

8. **其他配置**：
   - **图例（Legend）**：✅ 显示
   - **工具提示（Tooltip）**：✅ 显示
   - **数据标签（Data Labels）**：❌ 不显示（避免图表拥挤）

9. 点击 **"Save（保存）"** 保存图表

#### 配置示例截图说明

```
图表配置界面布局：
┌─────────────────────────────────────┐
│ Chart Name: 库存对比图表              │
│ Chart Title: 库存对比（理论 vs 实际） │
├─────────────────────────────────────┤
│ Report Type: [Report ▼]              │
│ Report Name: [Stock Ledger ▼]       │
│ Filters: {JSON 格式筛选条件}         │
├─────────────────────────────────────┤
│ Chart Type: [Bar ▼]                 │
│   └─ [Grouped Bar] [Stacked Bar]    │
├─────────────────────────────────────┤
│ X Axis:                              │
│   Field: [item_name ▼]              │
│   Label: 物料名称                     │
│   Sort: [Ascending ▼]               │
├─────────────────────────────────────┤
│ Y Axis:                              │
│   Field: [qty_after_transaction ▼]   │
│   Label: 库存数量                     │
│   Aggregation: [Sum ▼]              │
├─────────────────────────────────────┤
│ Series:                              │
│   [+ Add Series]                     │
│   Series 1: 理论库存                  │
│     Field: qty_after_transaction     │
│     Color: #1890ff                   │
│   Series 2: 实际库存                  │
│     Field: actual_qty                │
│     Color: #52c41a                   │
├─────────────────────────────────────┤
│ [Save] [Cancel]                      │
└─────────────────────────────────────┘
```

#### 预期效果

图表会显示：
- X 轴：物料名称（如：香辣鸡腿、汉堡面包等）
- Y 轴：库存数量
- 两个系列：理论库存（蓝色柱状图）和实际库存（绿色柱状图）
- 可以直观对比每个物料的理论库存和实际库存差异

### 3.2 图表 2：差异数量折线图

**用途**：展示各物料的差异数量（盘盈/盘亏）

#### 配置步骤

1. **图表类型（Chart Type）**：选择 `Line`（折线图）

2. **X 轴（X Axis）**：
   - 字段：`item_name`（物料名称）

3. **Y 轴（Y Axis）**：
   - 字段：自定义计算字段 `actual_qty - qty_after_transaction`（差异数量）

4. **图表标题**：`库存差异分析`

**注意**：ERPNext 的 Set Chart 可能不支持直接计算字段，需要：
- 方案 A：在 Stock Ledger 报表中添加计算列
- 方案 B：使用 `actual_qty` 和 `qty_after_transaction` 两个系列，手动对比

### 3.3 图表 3：物料分组饼图

**用途**：展示不同物料分组的库存占比

#### 配置步骤

1. **图表类型（Chart Type）**：选择 `Pie`（饼图）

2. **分组字段（Group By）**：
   - 字段：`item_group`（物料分组）

3. **数值字段（Value Field）**：
   - 字段：`qty_after_transaction`（理论库存）
   - 或者：`actual_qty`（实际库存）

4. **图表标题**：`物料分组库存占比`

### 3.4 图表 4：累计销量柱状图

**用途**：展示各物料的累计销量（销售出库数量）

#### 配置步骤

1. **图表类型（Chart Type）**：选择 `Bar`（柱状图）

2. **筛选条件（Filters）**：
   - `voucher_type`: `Sales Invoice`（销售发票）
   - 或者：`Delivery Note`（交货单）

3. **X 轴（X Axis）**：
   - 字段：`item_name`（物料名称）

4. **Y 轴（Y Axis）**：
   - 字段：`out_qty`（出库数量）
   - 聚合方式：`Sum`（求和）

5. **图表标题**：`累计销量统计`

---

## 四、创建多个图表组合展示

### 4.1 推荐布局

在 Dashboard 中创建多个图表，组合展示完整的库存汇总信息：

| 图表位置 | 图表类型 | 展示内容 |
|---------|---------|---------|
| 左上 | 柱状图 | 库存对比（理论 vs 实际） |
| 右上 | 折线图 | 差异数量趋势 |
| 左下 | 饼图 | 物料分组占比 |
| 右下 | 柱状图 | 累计销量统计 |

### 4.2 创建步骤

1. **创建第一个图表**（库存对比）
   - 按照 3.1 节的配置步骤创建

2. **创建第二个图表**（差异数量）
   - 按照 3.2 节的配置步骤创建

3. **创建第三个图表**（物料分组）
   - 按照 3.3 节的配置步骤创建

4. **创建第四个图表**（累计销量）
   - 按照 3.4 节的配置步骤创建

5. **调整布局**
   - 拖拽图表调整位置
   - 调整图表大小

---

## 五、数据计算说明

### 5.1 初始库存数量（opening_qty）

**计算方法**：
```
初始库存数量 = 理论库存数量 + 累计销量
```

**在 Stock Ledger 中获取**：
- 查询当日第一笔交易前的库存数量
- 或查询上一日最后一笔交易的 `qty_after_transaction`

**在 Set Chart 中展示**：
- 由于 ERPNext Set Chart 不支持直接计算，建议：
  - 方案 A：创建自定义报表，添加计算列
  - 方案 B：在 Dashboard 中创建两个图表，分别展示理论库存和累计销量

### 5.2 累计销量（sales_qty）

**计算方法**：
```
累计销量 = Σ(out_qty) WHERE voucher_type IN ('Sales Invoice', 'Delivery Note')
```

**在 Set Chart 中展示**：
- 筛选条件：`voucher_type` = `Sales Invoice` 或 `Delivery Note`
- Y 轴：`out_qty`（出库数量）
- 聚合方式：`Sum`（求和）
- 分组：`item_code`（物料编码）

### 5.3 理论库存数量（theoretical_qty）

**计算方法**：
```
理论库存数量 = qty_after_transaction（当日最后一笔交易后数量）
```

**在 Set Chart 中展示**：
- Y 轴：`qty_after_transaction`（交易后数量）
- 筛选条件：按日期范围查询
- 分组：`item_code`（物料编码）

### 5.4 实际库存数量（actual_qty）

**计算方法**：
```
实际库存数量 = actual_qty WHERE voucher_type = 'Stock Reconciliation'
```

**在 Set Chart 中展示**：
- 筛选条件：`voucher_type` = `Stock Reconciliation`
- Y 轴：`actual_qty`（实际数量）
- 分组：`item_code`（物料编码）

### 5.5 差异数量（diff_qty）

**计算方法**：
```
差异数量 = 实际库存数量 - 理论库存数量
```

**在 Set Chart 中展示**：
- ERPNext Set Chart 不支持直接计算，需要：
  - 方案 A：创建自定义报表，添加计算列 `diff_qty = actual_qty - qty_after_transaction`
  - 方案 B：创建两个系列，分别展示实际库存和理论库存，手动对比

---

## 六、创建自定义报表（推荐方案）

### 6.1 为什么需要自定义报表

ERPNext 的 Set Chart 基于报表数据，但 Stock Ledger 报表可能不包含所有需要的计算字段（如 `opening_qty`、`diff_qty`）。

**解决方案**：创建自定义报表，添加计算列。

### 6.2 创建自定义报表步骤

#### Step 1: 创建报表脚本

1. 导航至：**主页 > 设置 > 自定义 > 报表脚本**
2. 点击 **"新建"**
3. 填写报表信息：
   - **报表名称**：`Daily Stock Summary`
   - **报表类型**：`Script Report`
   - **报表脚本**：编写 Python 脚本

#### Step 2: 编写报表脚本

```python
import frappe
from frappe import _

def execute(filters=None):
    columns = [
        {
            "fieldname": "item_code",
            "label": _("物料编码"),
            "fieldtype": "Link",
            "options": "Item",
            "width": 120
        },
        {
            "fieldname": "item_name",
            "label": _("物料名称"),
            "fieldtype": "Data",
            "width": 200
        },
        {
            "fieldname": "item_group",
            "label": _("物料分组"),
            "fieldtype": "Link",
            "options": "Item Group",
            "width": 120
        },
        {
            "fieldname": "opening_qty",
            "label": _("初始库存数量"),
            "fieldtype": "Float",
            "width": 120,
            "precision": 4
        },
        {
            "fieldname": "sales_qty",
            "label": _("累计销量"),
            "fieldtype": "Float",
            "width": 120,
            "precision": 4
        },
        {
            "fieldname": "theoretical_qty",
            "label": _("理论库存数量"),
            "fieldtype": "Float",
            "width": 120,
            "precision": 4
        },
        {
            "fieldname": "actual_qty",
            "label": _("实际库存数量"),
            "fieldtype": "Float",
            "width": 120,
            "precision": 4
        },
        {
            "fieldname": "diff_qty",
            "label": _("差异数量"),
            "fieldtype": "Float",
            "width": 120,
            "precision": 4
        },
        {
            "fieldname": "stock_uom",
            "label": _("物料单位"),
            "fieldtype": "Data",
            "width": 80
        }
    ]
    
    data = get_data(filters)
    
    return columns, data

def get_data(filters):
    company = filters.get("company")
    from_date = filters.get("from_date")
    to_date = filters.get("to_date")
    warehouse = filters.get("warehouse")
    item_group = filters.get("item_group")
    
    # 构建查询条件
    conditions = {
        "company": company,
        "posting_date": ["between", [from_date, to_date]]
    }
    
    if warehouse:
        conditions["warehouse"] = warehouse
    
    if item_group:
        conditions["item_group"] = item_group
    
    # 查询 Stock Ledger 数据
    stock_ledger = frappe.db.sql("""
        SELECT 
            item_code,
            item_name,
            item_group,
            warehouse,
            posting_date,
            voucher_type,
            voucher_no,
            out_qty,
            qty_after_transaction,
            actual_qty,
            stock_uom
        FROM `tabStock Ledger Entry`
        WHERE company = %(company)s
            AND posting_date BETWEEN %(from_date)s AND %(to_date)s
            AND (warehouse = %(warehouse)s OR %(warehouse)s IS NULL)
            AND (item_group = %(item_group)s OR %(item_group)s IS NULL)
        ORDER BY item_code, posting_date, posting_time
    """, {
        "company": company,
        "from_date": from_date,
        "to_date": to_date,
        "warehouse": warehouse or None,
        "item_group": item_group or None
    }, as_dict=True)
    
    # 按物料分组聚合数据
    item_map = {}
    
    for entry in stock_ledger:
        item_code = entry.item_code
        
        if item_code not in item_map:
            item_map[item_code] = {
                "item_code": entry.item_code,
                "item_name": entry.item_name,
                "item_group": entry.item_group,
                "stock_uom": entry.stock_uom,
                "sales_qty": 0.0,
                "theoretical_qty": 0.0,
                "actual_qty": 0.0
            }
        
        item = item_map[item_code]
        
        # 累计销量（销售出库）
        if entry.voucher_type in ["Sales Invoice", "Delivery Note"]:
            item["sales_qty"] += entry.out_qty or 0.0
        
        # 更新理论库存（交易后数量）
        if entry.qty_after_transaction is not None:
            item["theoretical_qty"] = entry.qty_after_transaction
        
        # 更新实际库存（盘点单）
        if entry.voucher_type == "Stock Reconciliation":
            if entry.actual_qty is not None:
                item["actual_qty"] = entry.actual_qty
    
    # 计算初始库存和差异数量
    result = []
    for item_code, item in item_map.items():
        # 初始库存 = 理论库存 + 累计销量
        opening_qty = item["theoretical_qty"] + item["sales_qty"]
        
        # 差异数量 = 实际库存 - 理论库存
        diff_qty = item["actual_qty"] - item["theoretical_qty"]
        
        result.append({
            "item_code": item["item_code"],
            "item_name": item["item_name"],
            "item_group": item["item_group"],
            "opening_qty": opening_qty,
            "sales_qty": item["sales_qty"],
            "theoretical_qty": item["theoretical_qty"],
            "actual_qty": item["actual_qty"],
            "diff_qty": diff_qty,
            "stock_uom": item["stock_uom"]
        })
    
    return result
```

#### Step 3: 配置报表筛选条件

在报表脚本中添加筛选条件配置：

```python
def get_filters():
    return [
        {
            "fieldname": "company",
            "label": _("公司"),
            "fieldtype": "Link",
            "options": "Company",
            "reqd": 1
        },
        {
            "fieldname": "from_date",
            "label": _("开始日期"),
            "fieldtype": "Date",
            "reqd": 1
        },
        {
            "fieldname": "to_date",
            "label": _("结束日期"),
            "fieldtype": "Date",
            "reqd": 1
        },
        {
            "fieldname": "warehouse",
            "label": _("仓库"),
            "fieldtype": "Link",
            "options": "Warehouse"
        },
        {
            "fieldname": "item_group",
            "label": _("物料分组"),
            "fieldtype": "Link",
            "options": "Item Group"
        }
    ]
```

#### Step 4: 在 Set Chart 中使用自定义报表

1. 创建新的 Chart
2. **报表类型**：选择 `Report`
3. **报表名称**：选择 `Daily Stock Summary`（自定义报表）
4. **图表类型**：选择 `Bar`（柱状图）
5. **X 轴**：`item_name`（物料名称）
6. **Y 轴**：`opening_qty`, `sales_qty`, `theoretical_qty`, `actual_qty`, `diff_qty`（多个系列）

---

## 七、常见问题

### Q1: 在 Stock Ledger 报表页面找不到"Chart"按钮怎么办？

**A**: 
- **可能原因 1**：ERPNext 版本不支持在报表页面直接创建图表
  - **解决方案**：使用方式二，在 Dashboard 中创建图表
  
- **可能原因 2**：需要先运行报表，生成数据后才能创建图表
  - **解决方案**：先设置筛选条件，点击"Update"运行报表，然后再创建图表
  
- **可能原因 3**：用户权限不足
  - **解决方案**：联系系统管理员，确保有创建图表的权限

### Q2: Set Chart 不支持计算字段怎么办？

**A**: 
- **问题**：无法直接计算 `diff_qty = actual_qty - theoretical_qty`
- **解决方案**：
  1. **方案 A（推荐）**：创建自定义报表，在报表脚本中添加计算列（详见第六章）
  2. **方案 B**：创建两个系列（实际库存和理论库存），手动对比差异
  3. **方案 C**：使用 ERPNext 的"Number Card"功能，显示单个计算值

### Q3: 如何展示初始库存数量（opening_qty）？

**A**: 
- **问题**：Stock Ledger 报表中没有 `opening_qty` 字段
- **解决方案**：
  1. **方案 A（推荐）**：创建自定义报表，添加计算列 `opening_qty = theoretical_qty + sales_qty`
  2. **方案 B**：创建两个图表：
     - 图表 1：展示理论库存（`qty_after_transaction`）
     - 图表 2：展示累计销量（`out_qty`，筛选销售出库）
     - 手动计算：初始库存 = 理论库存 + 累计销量

### Q4: 如何筛选销售出库数据（sales_qty）？

**A**: 
- **在筛选条件中添加**：
  ```json
  {
    "voucher_type": ["in", ["Sales Invoice", "Delivery Note"]]
  }
  ```
- **或者使用多个筛选条件**：
  - `voucher_type` = `Sales Invoice`
  - 或者：`voucher_type` = `Delivery Note`

### Q5: 如何展示差异数量（盘盈/盘亏）？

**A**: 
- **方案 A（推荐）**：创建自定义报表，添加计算列 `diff_qty = actual_qty - theoretical_qty`
- **方案 B**：创建两个系列：
  - 系列 1：`actual_qty`（实际库存）
  - 系列 2：`qty_after_transaction`（理论库存）
  - 通过柱状图对比，手动查看差异
- **方案 C**：创建折线图，使用两个系列，差异通过两条线的距离判断

### Q6: 图表数据不准确怎么办？

**A**: 
1. **检查筛选条件**：
   - 确认公司、日期、仓库等筛选条件是否正确
   - 检查日期格式是否正确（YYYY-MM-DD）
   
2. **检查数据源**：
   - 确认 Stock Ledger 报表数据是否完整
   - 检查是否有数据权限限制
   
3. **检查字段映射**：
   - 确认字段名称是否正确（区分大小写）
   - 确认字段类型是否匹配（数值字段不能用于文本字段）
   
4. **检查聚合方式**：
   - 如果数据需要汇总，确认聚合方式是否正确（Sum、Average 等）
   
5. **清除缓存**：
   - 清除浏览器缓存
   - 在 ERPNext 中清除报表缓存

### Q7: 如何在图表中显示物料单位（stock_uom）？

**A**: 
- **方案 A**：在 Y 轴标签中添加单位，例如：`库存数量（个）`
- **方案 B**：在工具提示（Tooltip）中显示单位信息
- **方案 C**：创建自定义报表，将数量和单位合并为一个字段

### Q8: 图表显示"无数据"怎么办？

**A**: 
1. **检查筛选条件**：
   - 确认筛选条件是否过于严格
   - 尝试放宽筛选条件（如不设置仓库、物料分组等）
   
2. **检查数据是否存在**：
   - 在 Stock Ledger 报表中直接查询，确认是否有数据
   - 检查日期范围是否正确
   
3. **检查字段映射**：
   - 确认 X 轴和 Y 轴字段是否存在
   - 确认字段名称拼写是否正确

### Q9: 如何导出图表数据？

**A**: 
- **方案 A**：在 Stock Ledger 报表页面导出数据（Excel、CSV 等）
- **方案 B**：在图表配置中添加导出功能（如果 ERPNext 版本支持）
- **方案 C**：截图保存图表

### Q10: 如何分享图表给其他用户？

**A**: 
- **方案 A**：将图表添加到共享 Dashboard，其他用户访问 Dashboard 即可查看
- **方案 B**：导出图表配置，其他用户导入配置即可使用
- **方案 C**：截图分享（最简单但无法交互）

---

## 八、最佳实践

### 8.1 图表设计建议

1. **图表类型选择**：
   - 对比数据：使用柱状图（Bar）
   - 趋势数据：使用折线图（Line）
   - 占比数据：使用饼图（Pie）

2. **数据筛选**：
   - 始终设置公司筛选条件
   - 使用日期范围筛选，避免查询过多数据
   - 按仓库筛选，提高数据准确性

3. **图表布局**：
   - 在 Dashboard 中创建多个图表，组合展示
   - 使用清晰的图表标题
   - 添加图表说明文字

### 8.2 性能优化

1. **数据量控制**：
   - 使用日期范围筛选，避免查询过多历史数据
   - 按仓库或物料分组筛选，减少数据量

2. **缓存设置**：
   - 对于历史数据，可以设置报表缓存
   - 当日数据实时查询

---

## 九、相关文档

- [每日库存汇总表查看与展示方案](../../human/business/daily-stock-summary-report.md)
- [盘点单 TTPOS 与 ERPNext 数据同步机制](../../human/business/stock-reconciliation-erp-sync.md)
- [ERPNext 官方文档 - Dashboard Chart](https://docs.erpnext.com/docs/user/manual/en/customize-erpnext/dashboard-chart)

---

**最后更新**：2025-01-17  
**维护者**：TTPOS Team

