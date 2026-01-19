# ERPNext 库存差异报表创建指南

> 在 ERPNext v15 中创建库存差异分析报表的详细操作步骤（适用于 Frappe Cloud）

---

## 一、报表需求说明

### 1.1 报表功能

显示指定门店（仓库）、指定日期的物料库存差异分析，包括：

- **初始库存**：当日该物料的初始库存数量
- **销售数量**：当日该物料的累计销量
- **理论库存**：初始库存 - 累计销量
- **实际库存**：当日该物料的最后盘点数量
- **差异数量**：实际库存 - 理论库存

### 1.2 数据源

- **主表**：`tabStock Ledger Entry`（库存分类账）
- **关联表**：`tabItem`（物料主数据）

### 1.3 限制条件

- ✅ 使用 Frappe Cloud（官方云）
- ❌ 不能使用 Server Scripts
- ✅ 可以使用 Query Report（SQL 报表）

---

## 二、创建步骤

### 2.1 进入报表模块

1. 登录 ERPNext 系统
2. 在主菜单中找到 **"报表"**（Report）模块
3. 点击进入报表列表页面

### 2.2 创建新报表

1. 点击右上角的 **"新建"**（New）按钮
2. 选择 **"查询报表"**（Query Report）
3. 系统会打开报表编辑页面

### 2.3 填写报表基本信息

在报表编辑页面填写以下信息：

| 字段 | 值 | 说明 |
|------|-----|------|
| **报表名称** | `库存差异分析报表` | 报表的显示名称 |
| **模块** | `库存`（Stock） | 所属模块 |
| **报表类型** | `查询报表`（Query Report） | 报表类型 |

### 2.4 编写 SQL 查询

在 **"查询"**（Query）字段中，输入以下 SQL 语句：

```sql
-- ERPNext 库存差异分析报表 SQL 查询
-- 功能：显示指定门店、指定日期的物料库存差异
-- 数据源：Stock Ledger Entry（库存分类账）
-- 注意：Stock Ledger 界面显示的 "Out Qty" 在数据库中对应 actual_qty 字段（负数表示出库）

SELECT
    -- 基础字段
    sle.item_code AS '物料编码',
    i.item_name AS '物料名称',
    i.item_group AS '物料分组',
    i.stock_uom AS '物料单位',
    
    -- 初始库存：获取当天开始时的库存
    -- 方法：查询前一天最后一条记录的 qty_after_transaction（交易后数量）
    COALESCE(opening.qty_after_transaction, 0) AS '初始库存数量',
    
    -- 销售数量：汇总当天所有 Sales Invoice 类型的出库数量
    -- 注意：actual_qty < 0 表示出库，需要取绝对值
    COALESCE(ABS(sales.total_sales_qty), 0) AS '销售数量',
    
    -- 理论库存：初始库存 - 累计销量
    COALESCE(opening.qty_after_transaction, 0) - COALESCE(ABS(sales.total_sales_qty), 0) AS '理论库存数量',
    
    -- 实际库存：获取当天最后一次 Stock Reconciliation 后的库存
    -- 优先使用 qty_after_transaction（交易后数量），如果没有则使用 actual_qty
    COALESCE(
        COALESCE(recon.qty_after_transaction, recon.actual_qty),
        last_record.qty_after_transaction,
        0
    ) AS '实际库存数量',
    
    -- 差异数量：实际库存 - 理论库存
    (
        COALESCE(
            COALESCE(recon.qty_after_transaction, recon.actual_qty),
            last_record.qty_after_transaction,
            0
        )
        -
        (COALESCE(opening.qty_after_transaction, 0) - COALESCE(ABS(sales.total_sales_qty), 0))
    ) AS '差异数量'

FROM
    -- 主表：获取当天有库存变动的所有物料
    (
        SELECT DISTINCT
            item_code,
            warehouse,
            company
        FROM `tabStock Ledger Entry`
        WHERE posting_date = %(date)s
            AND warehouse = %(warehouse)s
            AND company = %(company)s
    ) sle
    
    -- 关联物料主数据
    INNER JOIN `tabItem` i ON sle.item_code = i.name
    
    -- 获取初始库存（前一天最后一条记录）
    LEFT JOIN (
        SELECT
            sle1.item_code,
            sle1.warehouse,
            sle1.company,
            sle1.qty_after_transaction
        FROM `tabStock Ledger Entry` sle1
        INNER JOIN (
            SELECT
                item_code,
                warehouse,
                company,
                MAX(CONCAT(posting_date, ' ', posting_time, ' ', creation)) AS max_datetime
            FROM `tabStock Ledger Entry`
            WHERE posting_date < %(date)s
                AND warehouse = %(warehouse)s
                AND company = %(company)s
            GROUP BY item_code, warehouse, company
        ) sle2 ON sle1.item_code = sle2.item_code
            AND sle1.warehouse = sle2.warehouse
            AND sle1.company = sle2.company
            AND CONCAT(sle1.posting_date, ' ', sle1.posting_time, ' ', sle1.creation) = sle2.max_datetime
    ) opening ON sle.item_code = opening.item_code
        AND sle.warehouse = opening.warehouse
        AND sle.company = opening.company
    
    -- 计算销售数量（当天所有 Sales Invoice 的出库数量）
    LEFT JOIN (
        SELECT
            item_code,
            warehouse,
            company,
            SUM(actual_qty) AS total_sales_qty
        FROM `tabStock Ledger Entry`
        WHERE posting_date = %(date)s
            AND warehouse = %(warehouse)s
            AND company = %(company)s
            AND voucher_type = 'Sales Invoice'
            AND actual_qty < 0  -- 出库为负数
        GROUP BY item_code, warehouse, company
    ) sales ON sle.item_code = sales.item_code
        AND sle.warehouse = sales.warehouse
        AND sle.company = sales.company
    
    -- 获取实际库存（当天最后一次 Stock Reconciliation）
    LEFT JOIN (
        SELECT
            sle1.item_code,
            sle1.warehouse,
            sle1.company,
            sle1.qty_after_transaction,
            sle1.actual_qty
        FROM `tabStock Ledger Entry` sle1
        INNER JOIN (
            SELECT
                item_code,
                warehouse,
                company,
                MAX(CONCAT(posting_time, ' ', creation)) AS max_datetime
            FROM `tabStock Ledger Entry`
            WHERE posting_date = %(date)s
                AND warehouse = %(warehouse)s
                AND company = %(company)s
                AND voucher_type = 'Stock Reconciliation'
            GROUP BY item_code, warehouse, company
        ) sle2 ON sle1.item_code = sle2.item_code
            AND sle1.warehouse = sle2.warehouse
            AND sle1.company = sle2.company
            AND sle1.voucher_type = 'Stock Reconciliation'
            AND CONCAT(sle1.posting_time, ' ', sle1.creation) = sle2.max_datetime
    ) recon ON sle.item_code = recon.item_code
        AND sle.warehouse = recon.warehouse
        AND sle.company = recon.company
    
    -- 如果没有盘点记录，获取当天最后一条记录
    LEFT JOIN (
        SELECT
            sle1.item_code,
            sle1.warehouse,
            sle1.company,
            sle1.qty_after_transaction
        FROM `tabStock Ledger Entry` sle1
        INNER JOIN (
            SELECT
                item_code,
                warehouse,
                company,
                MAX(CONCAT(posting_time, ' ', creation)) AS max_datetime
            FROM `tabStock Ledger Entry`
            WHERE posting_date = %(date)s
                AND warehouse = %(warehouse)s
                AND company = %(company)s
            GROUP BY item_code, warehouse, company
        ) sle2 ON sle1.item_code = sle2.item_code
            AND sle1.warehouse = sle2.warehouse
            AND sle1.company = sle2.company
            AND CONCAT(sle1.posting_time, ' ', sle1.creation) = sle2.max_datetime
        WHERE NOT EXISTS (
            SELECT 1
            FROM `tabStock Ledger Entry` sle_recon
            WHERE sle_recon.item_code = sle1.item_code
                AND sle_recon.warehouse = sle1.warehouse
                AND sle_recon.company = sle1.company
                AND sle_recon.posting_date = %(date)s
                AND sle_recon.voucher_type = 'Stock Reconciliation'
        )
    ) last_record ON sle.item_code = last_record.item_code
        AND sle.warehouse = last_record.warehouse
        AND sle.company = last_record.company

-- 按物料编码排序
ORDER BY
    sle.item_code
```

**重要说明**：
- 此 SQL 使用了 LEFT JOIN 和子查询优化，避免重复计算
- 使用 `actual_qty` 字段（负数表示出库）来计算销售数量，而不是 `out_qty`（该字段可能不存在于数据库中）
- 初始库存通过查询前一天最后一条记录获取，确保准确性
- 实际库存优先使用盘点记录，如果没有则使用当天最后一条记录

### 2.5 简化版 SQL（如果上述 SQL 报错，可使用此版本）

如果上述 SQL 查询在 ERPNext 中报错（可能是日期时间排序问题），可以使用以下简化版本：

```sql
-- ERPNext 库存差异分析报表 SQL 查询（简化版）
-- 适用于 ERPNext Query Report，兼容性更好

SELECT
    -- 基础字段
    sle.item_code AS '物料编码',
    i.item_name AS '物料名称',
    i.item_group AS '物料分组',
    i.stock_uom AS '物料单位',
    
    -- 初始库存：前一天最后一条记录的 qty_after_transaction
    COALESCE(
        (
            SELECT qty_after_transaction
            FROM `tabStock Ledger Entry` sle_opening
            WHERE sle_opening.item_code = sle.item_code
                AND sle_opening.warehouse = sle.warehouse
                AND sle_opening.company = sle.company
                AND sle_opening.posting_date < %(date)s
            ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC
            LIMIT 1
        ),
        0
    ) AS '初始库存数量',
    
    -- 销售数量：当天所有 Sales Invoice 的出库数量（actual_qty 负数取绝对值）
    COALESCE(
        (
            SELECT ABS(SUM(actual_qty))
            FROM `tabStock Ledger Entry` sle_sales
            WHERE sle_sales.item_code = sle.item_code
                AND sle_sales.warehouse = sle.warehouse
                AND sle_sales.company = sle.company
                AND sle_sales.posting_date = %(date)s
                AND sle_sales.voucher_type = 'Sales Invoice'
                AND sle_sales.actual_qty < 0
        ),
        0
    ) AS '销售数量',
    
    -- 理论库存：初始库存 - 销售数量
    (
        COALESCE(
            (
                SELECT qty_after_transaction
                FROM `tabStock Ledger Entry` sle_opening
                WHERE sle_opening.item_code = sle.item_code
                    AND sle_opening.warehouse = sle.warehouse
                    AND sle_opening.company = sle.company
                    AND sle_opening.posting_date < %(date)s
                ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC
                LIMIT 1
            ),
            0
        )
        -
        COALESCE(
            (
                SELECT ABS(SUM(actual_qty))
                FROM `tabStock Ledger Entry` sle_sales
                WHERE sle_sales.item_code = sle.item_code
                    AND sle_sales.warehouse = sle.warehouse
                    AND sle_sales.company = sle.company
                    AND sle_sales.posting_date = %(date)s
                    AND sle_sales.voucher_type = 'Sales Invoice'
                    AND sle_sales.actual_qty < 0
            ),
            0
        )
    ) AS '理论库存数量',
    
    -- 实际库存：当天最后一次 Stock Reconciliation 的 qty_after_transaction
    COALESCE(
        (
            SELECT COALESCE(qty_after_transaction, actual_qty)
            FROM `tabStock Ledger Entry` sle_recon
            WHERE sle_recon.item_code = sle.item_code
                AND sle_recon.warehouse = sle.warehouse
                AND sle_recon.company = sle.company
                AND sle_recon.posting_date = %(date)s
                AND sle_recon.voucher_type = 'Stock Reconciliation'
            ORDER BY sle_recon.posting_time DESC
            LIMIT 1
        ),
        -- 如果没有盘点记录，使用当天最后一条记录
        (
            SELECT qty_after_transaction
            FROM `tabStock Ledger Entry` sle_last
            WHERE sle_last.item_code = sle.item_code
                AND sle_last.warehouse = sle.warehouse
                AND sle_last.company = sle.company
                AND sle_last.posting_date = %(date)s
            ORDER BY sle_last.posting_time DESC
            LIMIT 1
        ),
        0
    ) AS '实际库存数量',
    
    -- 差异数量：实际库存 - 理论库存
    (
        COALESCE(
            (
                SELECT COALESCE(qty_after_transaction, actual_qty)
                FROM `tabStock Ledger Entry` sle_recon
                WHERE sle_recon.item_code = sle.item_code
                    AND sle_recon.warehouse = sle.warehouse
                    AND sle_recon.company = sle.company
                    AND sle_recon.posting_date = %(date)s
                    AND sle_recon.voucher_type = 'Stock Reconciliation'
                ORDER BY sle_recon.posting_time DESC
                LIMIT 1
            ),
            (
                SELECT qty_after_transaction
                FROM `tabStock Ledger Entry` sle_last
                WHERE sle_last.item_code = sle.item_code
                    AND sle_last.warehouse = sle.warehouse
                    AND sle_last.company = sle.company
                    AND sle_last.posting_date = %(date)s
                ORDER BY sle_last.posting_time DESC
                LIMIT 1
            ),
            0
        )
        -
        (
            COALESCE(
                (
                    SELECT qty_after_transaction
                    FROM `tabStock Ledger Entry` sle_opening
                    WHERE sle_opening.item_code = sle.item_code
                        AND sle_opening.warehouse = sle.warehouse
                        AND sle_opening.company = sle.company
                        AND sle_opening.posting_date < %(date)s
                    ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC
                    LIMIT 1
                ),
                0
            )
            -
            COALESCE(
                (
                    SELECT ABS(SUM(actual_qty))
                    FROM `tabStock Ledger Entry` sle_sales
                    WHERE sle_sales.item_code = sle.item_code
                        AND sle_sales.warehouse = sle.warehouse
                        AND sle_sales.company = sle.company
                        AND sle_sales.posting_date = %(date)s
                        AND sle_sales.voucher_type = 'Sales Invoice'
                        AND sle_sales.actual_qty < 0
                ),
                0
            )
        )
    ) AS '差异数量'

FROM
    `tabStock Ledger Entry` sle
    INNER JOIN `tabItem` i ON sle.item_code = i.name
    
WHERE
    sle.posting_date = %(date)s
    AND sle.warehouse = %(warehouse)s
    AND sle.company = %(company)s

GROUP BY
    sle.item_code,
    i.item_name,
    i.item_group,
    i.stock_uom

ORDER BY
    sle.item_code
```

### 2.6 设置查询参数

在报表编辑页面的 **"查询参数"**（Query Parameters）部分，添加以下参数：

| 参数名 | 标签 | 字段类型 | 选项 | 默认值 | 必填 |
|--------|------|----------|------|--------|------|
| `date` | 日期 | Date | - | `today` | ✅ 是 |
| `warehouse` | 仓库 | Link | Warehouse | - | ✅ 是 |
| `company` | 公司 | Link | Company | - | ✅ 是 |

**设置步骤**：

1. 点击 **"添加行"**（Add Row）按钮
2. 依次添加三个参数：
   - **date**：类型选择 `Date`，默认值输入 `today`
   - **warehouse**：类型选择 `Link`，选项输入 `Warehouse`
   - **company**：类型选择 `Link`，选项输入 `Company`
3. 将 `date` 和 `warehouse` 设置为必填（Required）

### 2.7 保存报表

1. 检查所有信息填写无误
2. 点击右上角的 **"保存"**（Save）按钮
3. 系统会提示保存成功

---

## 三、优化版 SQL（性能优化）

如果上述 SQL 查询性能较慢，可以使用以下优化版本：

```sql
-- ERPNext 库存差异分析报表 SQL 查询（优化版）
-- 使用 CTE（公共表表达式）提高查询性能

WITH 
-- 获取初始库存（前一天最后一条记录）
opening_stock AS (
    SELECT
        item_code,
        warehouse,
        company,
        qty_after_transaction AS opening_qty
    FROM `tabStock Ledger Entry` sle1
    WHERE sle1.posting_date < %(date)s
        AND sle1.warehouse = %(warehouse)s
        AND sle1.company = %(company)s
        AND (sle1.item_code, sle1.warehouse, sle1.company, sle1.posting_date, sle1.posting_time, sle1.creation) IN (
            SELECT 
                item_code,
                warehouse,
                company,
                posting_date,
                posting_time,
                creation
            FROM `tabStock Ledger Entry` sle2
            WHERE sle2.posting_date < %(date)s
                AND sle2.warehouse = %(warehouse)s
                AND sle2.company = %(company)s
                AND sle2.item_code = sle1.item_code
            ORDER BY sle2.posting_date DESC, sle2.posting_time DESC, sle2.creation DESC
            LIMIT 1
        )
),

-- 计算销售数量（当天所有 Sales Invoice）
sales_qty AS (
    SELECT
        item_code,
        warehouse,
        company,
        ABS(SUM(COALESCE(actual_qty, 0))) AS sales_qty
    FROM `tabStock Ledger Entry`
    WHERE posting_date = %(date)s
        AND warehouse = %(warehouse)s
        AND company = %(company)s
        AND voucher_type = 'Sales Invoice'
        AND actual_qty < 0
    GROUP BY item_code, warehouse, company
),

-- 获取实际库存（当天最后一次盘点记录）
actual_stock AS (
    SELECT
        item_code,
        warehouse,
        company,
        COALESCE(qty_after_transaction, actual_qty) AS actual_qty
    FROM `tabStock Ledger Entry` sle1
    WHERE sle1.posting_date = %(date)s
        AND sle1.warehouse = %(warehouse)s
        AND sle1.company = %(company)s
        AND sle1.voucher_type = 'Stock Reconciliation'
        AND (sle1.item_code, sle1.warehouse, sle1.company, sle1.posting_time, sle1.creation) IN (
            SELECT 
                item_code,
                warehouse,
                company,
                posting_time,
                creation
            FROM `tabStock Ledger Entry` sle2
            WHERE sle2.posting_date = %(date)s
                AND sle2.warehouse = %(warehouse)s
                AND sle2.company = %(company)s
                AND sle2.voucher_type = 'Stock Reconciliation'
                AND sle2.item_code = sle1.item_code
            ORDER BY sle2.posting_time DESC, sle2.creation DESC
            LIMIT 1
        )
),

-- 如果没有盘点记录，获取当天最后一条记录
last_stock AS (
    SELECT
        item_code,
        warehouse,
        company,
        qty_after_transaction AS actual_qty
    FROM `tabStock Ledger Entry` sle1
    WHERE sle1.posting_date = %(date)s
        AND sle1.warehouse = %(warehouse)s
        AND sle1.company = %(company)s
        AND NOT EXISTS (
            SELECT 1
            FROM actual_stock ac
            WHERE ac.item_code = sle1.item_code
                AND ac.warehouse = sle1.warehouse
                AND ac.company = sle1.company
        )
        AND (sle1.item_code, sle1.warehouse, sle1.company, sle1.posting_time, sle1.creation) IN (
            SELECT 
                item_code,
                warehouse,
                company,
                posting_time,
                creation
            FROM `tabStock Ledger Entry` sle2
            WHERE sle2.posting_date = %(date)s
                AND sle2.warehouse = %(warehouse)s
                AND sle2.company = %(company)s
                AND sle2.item_code = sle1.item_code
            ORDER BY sle2.posting_time DESC, sle2.creation DESC
            LIMIT 1
        )
),

-- 合并实际库存（优先使用盘点记录，否则使用最后一条记录）
final_actual_stock AS (
    SELECT * FROM actual_stock
    UNION ALL
    SELECT * FROM last_stock
    WHERE NOT EXISTS (
        SELECT 1 FROM actual_stock ac
        WHERE ac.item_code = last_stock.item_code
            AND ac.warehouse = last_stock.warehouse
            AND ac.company = last_stock.company
    )
),

-- 获取所有有库存变动的物料
all_items AS (
    SELECT DISTINCT
        sle.item_code,
        i.item_name,
        i.item_group,
        i.stock_uom
    FROM `tabStock Ledger Entry` sle
    INNER JOIN `tabItem` i ON sle.item_code = i.name
    WHERE sle.posting_date = %(date)s
        AND sle.warehouse = %(warehouse)s
        AND sle.company = %(company)s
)

-- 主查询：汇总所有数据
SELECT
    ai.item_code AS '物料编码',
    ai.item_name AS '物料名称',
    ai.item_group AS '物料分组',
    ai.stock_uom AS '物料单位',
    COALESCE(os.opening_qty, 0) AS '初始库存数量',
    COALESCE(sq.sales_qty, 0) AS '销售数量',
    COALESCE(os.opening_qty, 0) - COALESCE(sq.sales_qty, 0) AS '理论库存数量',
    COALESCE(fas.actual_qty, 0) AS '实际库存数量',
    COALESCE(fas.actual_qty, 0) - (COALESCE(os.opening_qty, 0) - COALESCE(sq.sales_qty, 0)) AS '差异数量'
FROM
    all_items ai
    LEFT JOIN opening_stock os ON ai.item_code = os.item_code 
        AND ai.item_code IN (SELECT item_code FROM opening_stock WHERE warehouse = %(warehouse)s AND company = %(company)s)
    LEFT JOIN sales_qty sq ON ai.item_code = sq.item_code
    LEFT JOIN final_actual_stock fas ON ai.item_code = fas.item_code
ORDER BY
    ai.item_code
```

**注意**：优化版 SQL 使用了 CTE，但 ERPNext 的 Query Report 可能不支持 CTE。如果报错，请使用第一个版本的 SQL。

---

## 四、字段说明

### 4.1 报表字段

| 字段名 | 中文名称 | 说明 | 计算公式 |
|--------|----------|------|----------|
| `item_code` | 物料编码 | 物料的唯一标识代码 | - |
| `item_name` | 物料名称 | 物料的描述性名称 | - |
| `item_group` | 物料分组 | 物料所属的分类或组别 | - |
| `stock_uom` | 物料单位 | 物料的计量单位 | - |
| `opening_qty` | 初始库存数量 | 当日该物料的初始库存 | 前一天最后一条记录的 `qty_after_transaction` |
| `sales_qty` | 销售数量 | 当日该物料的累计销量 | 当天所有 `Sales Invoice` 类型的 `actual_qty` 绝对值之和 |
| `theoretical_qty` | 理论库存数量 | 理论上的库存数量 | `初始库存数量 - 累计销量` |
| `actual_qty` | 实际库存数量 | 当日该物料的最后盘点数量 | 当天最后一次 `Stock Reconciliation` 的 `qty_after_transaction` 或 `actual_qty` |
| `diff_qty` | 差异数量 | 实际库存与理论库存的差异 | `实际库存数量 - 理论库存数量` |

### 4.2 计算逻辑说明

#### 初始库存计算

- **方法**：查询前一天（`posting_date < 指定日期`）最后一条记录的 `qty_after_transaction`
- **排序**：按 `posting_date DESC, posting_time DESC` 排序
- **字段说明**：`qty_after_transaction` 是交易后的库存数量，表示该笔交易完成后的库存余额
- **如果不存在**：返回 0（表示该物料是当天新增的，初始库存为 0）

#### 销售数量计算

- **方法**：汇总当天所有 `voucher_type = 'Sales Invoice'` 且 `actual_qty < 0` 的记录
- **字段说明**：
  - `actual_qty` 是实际数量变动（正数表示入库，负数表示出库）
  - ERPNext 界面显示的 "Out Qty" 是计算字段，数据库中对应的是 `actual_qty`（负数）
  - 因此使用 `actual_qty < 0` 来筛选出库记录，然后取绝对值
- **计算**：`ABS(SUM(actual_qty))`（取绝对值）
- **如果不存在**：返回 0（表示当天没有销售）

#### 实际库存计算

- **优先**：当天最后一次 `Stock Reconciliation` 记录的 `qty_after_transaction` 或 `actual_qty`
  - `qty_after_transaction` 优先，如果没有则使用 `actual_qty`
- **备选**：如果没有盘点记录，使用当天最后一条记录的 `qty_after_transaction`
- **字段说明**：
  - `qty_after_transaction`：交易后的库存数量（推荐使用）
  - `actual_qty`：实际数量（盘点单中通常等于实盘数量）
- **如果不存在**：返回 0

### 4.3 字段映射说明

**重要**：ERPNext 界面显示的字段名与数据库字段名的对应关系：

| 界面显示 | 数据库字段 | 说明 |
|---------|-----------|------|
| Out Qty | `actual_qty`（负数） | 出库数量，数据库中 `actual_qty < 0` 表示出库 |
| In Qty | `actual_qty`（正数） | 入库数量，数据库中 `actual_qty > 0` 表示入库 |
| Balance Qty | `qty_after_transaction` | 交易后的库存余额 |
| Actual Qty | `actual_qty` | 实际数量（盘点单中为实盘数量） |

**注意**：`out_qty` 和 `in_qty` 在数据库中**不存在**，它们是 ERPNext 界面根据 `actual_qty` 的正负值计算显示的。因此 SQL 查询中需要使用 `actual_qty` 字段，并通过正负值判断出入库。

---

## 五、使用报表

### 5.1 打开报表

1. 在报表列表中，找到 **"库存差异分析报表"**
2. 点击报表名称，打开报表页面

### 5.2 设置查询条件

在报表页面顶部，会显示查询参数输入框：

1. **日期**：选择要查询的日期（默认是今天）
2. **仓库**：选择要查询的仓库（门店）
3. **公司**：选择要查询的公司

### 5.3 运行报表

1. 填写完所有必填参数后
2. 点击 **"运行"**（Run）或 **"刷新"**（Refresh）按钮
3. 系统会执行 SQL 查询并显示结果

### 5.4 查看结果

报表会以表格形式显示所有物料的库存差异数据，包括：

- 物料基本信息（编码、名称、分组、单位）
- 初始库存数量
- 销售数量
- 理论库存数量
- 实际库存数量
- 差异数量

---

## 六、添加图表（可选）

### 6.1 创建图表

1. 在报表页面，点击右上角的 **"图表"**（Chart）按钮
2. 选择图表类型：
   - **柱状图**（Bar Chart）：适合显示差异数量
   - **折线图**（Line Chart）：适合显示库存趋势
   - **饼图**（Pie Chart）：适合显示物料分组占比

### 6.2 配置图表

**示例：差异数量柱状图**

- **图表类型**：柱状图（Bar Chart）
- **X 轴**：物料编码（`item_code`）
- **Y 轴**：差异数量（`diff_qty`）
- **图表标题**：库存差异分析

### 6.3 保存图表

1. 配置完成后，点击 **"保存"**（Save）按钮
2. 图表会显示在报表页面下方

---

## 七、测试 SQL 查询

### 7.1 在数据库中直接测试

在创建报表之前，可以先在 ERPNext 的数据库中直接测试 SQL 查询：

1. **进入数据库查询工具**：
   - 在 ERPNext 中，进入 **"开发者工具"**（Developer Tools）
   - 选择 **"查询"**（Query）或 **"控制台"**（Console）

2. **替换参数**：
   - 将 SQL 中的 `%(date)s` 替换为实际日期，如 `'2025-01-17'`
   - 将 `%(warehouse)s` 替换为实际仓库编码，如 `'WH-001'`
   - 将 `%(company)s` 替换为实际公司名称，如 `'Company A'`

3. **执行查询**：
   - 复制简化版 SQL（2.5 节）
   - 替换参数后执行
   - 检查结果是否正确

### 7.2 验证数据准确性

**验证初始库存**：
```sql
-- 检查某个物料的前一天最后一条记录
SELECT 
    item_code,
    posting_date,
    posting_time,
    qty_after_transaction
FROM `tabStock Ledger Entry`
WHERE item_code = 'MAT-001'  -- 替换为实际物料编码
    AND warehouse = 'WH-001'  -- 替换为实际仓库
    AND company = 'Company A'  -- 替换为实际公司
    AND posting_date < '2025-01-17'  -- 替换为实际日期
ORDER BY posting_date DESC, posting_time DESC
LIMIT 1;
```

**验证销售数量**：
```sql
-- 检查某个物料当天的销售记录
SELECT 
    item_code,
    voucher_type,
    actual_qty,
    ABS(actual_qty) AS sales_qty
FROM `tabStock Ledger Entry`
WHERE item_code = 'MAT-001'
    AND warehouse = 'WH-001'
    AND company = 'Company A'
    AND posting_date = '2025-01-17'
    AND voucher_type = 'Sales Invoice'
    AND actual_qty < 0;
```

**验证实际库存**：
```sql
-- 检查某个物料当天的盘点记录
SELECT 
    item_code,
    voucher_type,
    qty_after_transaction,
    actual_qty
FROM `tabStock Ledger Entry`
WHERE item_code = 'MAT-001'
    AND warehouse = 'WH-001'
    AND company = 'Company A'
    AND posting_date = '2025-01-17'
    AND voucher_type = 'Stock Reconciliation'
ORDER BY posting_time DESC
LIMIT 1;
```

## 八、常见问题

### 8.1 SQL 语法错误

**问题**：保存报表时提示 SQL 语法错误

**解决方案**：
1. 检查 SQL 语句中的表名是否正确（注意空格：`tabStock Ledger Entry`）
2. 检查字段名是否正确（注意大小写）
3. 检查参数占位符是否正确（`%(date)s`、`%(warehouse)s`、`%(company)s`）
4. 如果使用简化版 SQL 仍然报错，检查 ERPNext 版本是否支持相关 SQL 语法

### 8.2 查询结果为空

**问题**：运行报表后没有数据显示

**解决方案**：
1. 检查指定日期是否有库存变动记录
2. 检查指定仓库是否存在
3. 检查指定公司是否正确
4. 检查物料是否有对应的 Item 记录
5. 使用测试 SQL（7.2 节）验证数据是否存在

### 8.3 初始库存计算错误

**问题**：初始库存显示为 0，但实际应该有库存

**解决方案**：
1. 检查前一天是否有库存变动记录（使用 7.2 节的验证 SQL）
2. 如果物料是当天新增的，初始库存确实为 0
3. 检查排序逻辑是否正确（`ORDER BY posting_date DESC, posting_time DESC`）
4. 如果前一天有多条记录，确保获取的是最后一条

### 8.4 销售数量计算错误

**问题**：销售数量不准确

**解决方案**：
1. 检查 `voucher_type` 是否正确（应该是 `Sales Invoice`）
2. 检查 `actual_qty` 是否为负数（出库为负数）
3. 如果 ERPNext 中使用了其他销售单据类型（如 `Delivery Note`），需要修改 SQL 添加相应条件：
   ```sql
   AND sle_sales.voucher_type IN ('Sales Invoice', 'Delivery Note')
   ```
4. 使用测试 SQL（7.2 节）验证销售记录是否正确

### 8.5 实际库存计算错误

**问题**：实际库存显示不正确

**解决方案**：
1. 检查当天是否有 `Stock Reconciliation` 记录（使用 7.2 节的验证 SQL）
2. 检查 `qty_after_transaction` 字段是否有值
3. 如果没有盘点记录，系统会使用当天最后一条记录，可能不准确
4. 检查排序逻辑是否正确（`ORDER BY posting_time DESC`）

### 8.6 字段不存在错误

**问题**：提示字段 `out_qty` 不存在

**解决方案**：
- ERPNext 数据库中**没有** `out_qty` 字段
- 使用 `actual_qty` 字段，通过正负值判断出入库：
  - `actual_qty < 0`：出库
  - `actual_qty > 0`：入库
- 参考 4.3 节的字段映射说明

---

## 九、性能优化建议

### 8.1 添加索引

如果报表查询较慢，可以在数据库中为以下字段添加索引：

```sql
-- 在 ERPNext 数据库中执行（需要数据库管理员权限）
CREATE INDEX idx_stock_ledger_item_warehouse_date 
ON `tabStock Ledger Entry` (item_code, warehouse, posting_date, voucher_type);

CREATE INDEX idx_stock_ledger_warehouse_date 
ON `tabStock Ledger Entry` (warehouse, posting_date, posting_time);
```

### 8.2 限制查询范围

如果数据量很大，可以：

1. **添加物料分组过滤**：在 SQL 中添加 `item_group` 过滤条件
2. **添加日期范围**：改为查询日期范围而不是单日
3. **添加物料编码过滤**：只查询特定物料

### 8.3 使用缓存

ERPNext 的 Query Report 支持缓存，可以在报表设置中启用：

1. 打开报表编辑页面
2. 找到 **"缓存"**（Cache）选项
3. 设置缓存时间（如 1 小时）

---

## 十、扩展功能

### 9.1 添加物料分组汇总

可以在报表底部添加汇总行，按物料分组统计：

```sql
-- 在 SQL 末尾添加 UNION ALL 汇总查询
UNION ALL

SELECT
    '小计' AS '物料编码',
    CONCAT('分组：', i.item_group) AS '物料名称',
    i.item_group AS '物料分组',
    '' AS '物料单位',
    SUM(opening_qty) AS '初始库存数量',
    SUM(sales_qty) AS '销售数量',
    SUM(theoretical_qty) AS '理论库存数量',
    SUM(actual_qty) AS '实际库存数量',
    SUM(diff_qty) AS '差异数量'
FROM
    -- 主查询结果
GROUP BY
    i.item_group
```

### 9.2 添加差异预警

可以在 SQL 中添加条件，只显示差异较大的物料：

```sql
-- 在主查询的 WHERE 子句中添加
AND ABS(
    COALESCE(fas.actual_qty, 0) - (COALESCE(os.opening_qty, 0) - COALESCE(sq.sales_qty, 0))
) > 10  -- 差异大于 10 的物料
```

### 9.3 导出功能

ERPNext 的 Query Report 支持导出为 Excel、CSV 等格式：

1. 运行报表后，点击右上角的 **"导出"**（Export）按钮
2. 选择导出格式（Excel、CSV 等）
3. 下载文件

---

## 十一、总结

通过以上步骤，您可以在 ERPNext v15（Frappe Cloud）中创建一个完整的库存差异分析报表，包括：

✅ **表格展示**：清晰的物料库存差异数据  
✅ **图表展示**：直观的差异数量可视化  
✅ **参数过滤**：灵活的日期、仓库、公司筛选  
✅ **导出功能**：支持 Excel、CSV 导出  

**注意事项**：
- Frappe Cloud 不支持 Server Scripts，只能使用 Query Report
- SQL 查询性能取决于数据量，建议添加索引优化
- 如果查询结果不准确，需要根据实际业务逻辑调整 SQL

---

**最后更新**：2025-01-17  
**维护者**：TTPOS Team

