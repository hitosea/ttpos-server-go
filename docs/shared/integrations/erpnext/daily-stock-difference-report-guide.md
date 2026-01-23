# ERPNext 每日库存差异报表创建指南

> 在 ERPNext v15（Frappe Cloud）中创建每日库存差异分析报表的详细操作步骤

---

## 一、需求说明

### 1.1 报表功能

显示指定门店（仓库）、指定日期的物料库存差异分析，包括：

- **初始库存**（opening_qty）：当日该物料的初始库存数量
- **销售数量**（sales_qty）：当日该物料的累计销量
- **理论库存**（theoretical_qty）：初始库存 - 累计销量
- **实际库存**（actual_qty）：当日该物料的最后盘点数量
- **差异数量**（diff_qty）：实际库存 - 理论库存

### 1.2 数据源

- **主表**：`tabStock Ledger Entry`（库存分类账）
- **关联表**：`tabItem`（物料主数据）

### 1.3 限制条件

- ✅ 使用 Frappe Cloud（官方云）
- ❌ 不能使用 Server Scripts
- ✅ 可以使用 Query Report（SQL 报表）

### 1.4 报表字段说明

| 字段名 | 中文名称 | 说明 | 计算公式 |
|--------|----------|------|----------|
| `item_code` | 物料编码 | 物料的唯一标识代码 | - |
| `item_name` | 物料名称 | 物料的描述性名称 | - |
| `item_group` | 物料分组 | 物料所属的分类或组别 | - |
| `opening_qty` | 初始库存数量 | 当日该物料的初始库存 | 前一天最后一条记录的 `qty_after_transaction` |
| `sales_qty` | 销售数量 | 当日该物料的累计销量 | 当天所有 `Sales Invoice` 类型的 `actual_qty` 绝对值之和 |
| `theoretical_qty` | 理论库存数量 | 理论上的库存数量 | `初始库存数量 - 累计销量` |
| `actual_qty` | 实际库存数量 | 当日该物料的最后盘点数量 | 当天最后一次 `Stock Reconciliation` 的 `qty_after_transaction` 或 `actual_qty` |
| `diff_qty` | 差异数量 | 实际库存与理论库存的差异 | `实际库存数量 - 理论库存数量` |
| `stock_uom` | 物料单位 | 物料的计量单位 | - |

---

## 二、创建 Query Report 详细步骤

### 2.1 进入报表模块

1. 登录 ERPNext 系统
2. 在主菜单中找到 **"报表"**（Report）模块
   - 或者直接搜索：`Report`
   - 或者直接访问：`/app/query-report`
3. 点击进入报表列表页面

### 2.2 创建新报表

1. 点击右上角的 **"新建"**（New）按钮
2. 选择 **"查询报表"**（Query Report）
3. 系统会打开报表编辑页面

### 2.3 填写报表基本信息

在报表编辑页面填写以下信息：

| 字段 | 值 | 说明 |
|------|-----|------|
| **报表名称** | `每日库存差异分析报表` | 报表的显示名称 |
| **模块** | `库存`（Stock） | 所属模块 |
| **报表类型** | `查询报表`（Query Report） | 报表类型 |

### 2.4 编写 SQL 查询

在 **"查询"**（Query）字段中，输入以下 SQL 语句：

```sql
-- ERPNext 每日库存差异分析报表 SQL 查询
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
    COALESCE(
        (
            SELECT qty_after_transaction
            FROM `tabStock Ledger Entry` sle_opening
            WHERE sle_opening.item_code = sle.item_code
                AND sle_opening.warehouse = sle.warehouse
                AND sle_opening.company = sle.company
                AND sle_opening.posting_date < %(date)s
            ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC, sle_opening.creation DESC
            LIMIT 1
        ),
        0
    ) AS '初始库存数量',
    
    -- 销售数量：汇总当天所有 Sales Invoice 类型的出库数量
    -- 注意：actual_qty < 0 表示出库，需要取绝对值
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
                ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC, sle_opening.creation DESC
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
    
    -- 实际库存：获取当天最后一次 Stock Reconciliation 后的库存
    -- 优先使用 qty_after_transaction（交易后数量），如果没有则使用 actual_qty
    COALESCE(
        (
            SELECT COALESCE(qty_after_transaction, actual_qty)
            FROM `tabStock Ledger Entry` sle_recon
            WHERE sle_recon.item_code = sle.item_code
                AND sle_recon.warehouse = sle.warehouse
                AND sle_recon.company = sle.company
                AND sle_recon.posting_date = %(date)s
                AND sle_recon.voucher_type = 'Stock Reconciliation'
            ORDER BY sle_recon.posting_time DESC, sle_recon.creation DESC
            LIMIT 1
        ),
        -- 如果没有盘点记录，使用当天最后一条记录的 qty_after_transaction
        (
            SELECT qty_after_transaction
            FROM `tabStock Ledger Entry` sle_last
            WHERE sle_last.item_code = sle.item_code
                AND sle_last.warehouse = sle.warehouse
                AND sle_last.company = sle.company
                AND sle_last.posting_date = %(date)s
            ORDER BY sle_last.posting_time DESC, sle_last.creation DESC
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
                ORDER BY sle_recon.posting_time DESC, sle_recon.creation DESC
                LIMIT 1
            ),
            (
                SELECT qty_after_transaction
                FROM `tabStock Ledger Entry` sle_last
                WHERE sle_last.item_code = sle.item_code
                    AND sle_last.warehouse = sle_last.warehouse
                    AND sle_last.company = sle_last.company
                    AND sle_last.posting_date = %(date)s
                ORDER BY sle_last.posting_time DESC, sle_last.creation DESC
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
                    ORDER BY sle_opening.posting_date DESC, sle_opening.posting_time DESC, sle_opening.creation DESC
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

**重要说明**：

1. **初始库存计算**：
   - 查询前一天（`posting_date < 指定日期`）最后一条记录的 `qty_after_transaction`
   - 排序：`ORDER BY posting_date DESC, posting_time DESC, creation DESC`
   - 如果不存在（当天新增物料），返回 0

2. **销售数量计算**：
   - 汇总当天所有 `voucher_type = 'Sales Invoice'` 且 `actual_qty < 0` 的记录
   - 使用 `ABS(SUM(actual_qty))` 取绝对值
   - **注意**：数据库中 `actual_qty` 负数表示出库，界面显示的 "Out Qty" 是计算字段

3. **实际库存计算**：
   - 优先：当天最后一次 `Stock Reconciliation` 记录的 `qty_after_transaction` 或 `actual_qty`
   - 备选：如果没有盘点记录，使用当天最后一条记录的 `qty_after_transaction`
   - 排序：`ORDER BY posting_time DESC, creation DESC`

4. **字段映射**：
   - ERPNext 界面显示的 "Out Qty" 在数据库中**不存在**
   - 使用 `actual_qty` 字段，负数表示出库，正数表示入库
   - `qty_after_transaction` 是交易后的库存余额

### 2.5 设置查询参数

在报表编辑页面的 **"查询参数"**（Query Parameters）部分，添加以下参数：

| 参数名 | 标签 | 字段类型 | 选项 | 默认值 | 必填 |
|--------|------|----------|------|--------|------|
| `date` | 日期 | Date | - | `today` | ✅ 是 |
| `warehouse` | 仓库 | Link | Warehouse | - | ✅ 是 |
| `company` | 公司 | Link | Company | - | ✅ 是 |

**设置步骤**：

1. 在报表编辑页面找到 **"查询参数"**（Query Parameters）部分
2. 点击 **"添加行"**（Add Row）按钮
3. 依次添加三个参数：
   - **参数名**：`date`，**标签**：`日期`，**字段类型**：选择 `Date`，**默认值**：输入 `today`，**必填**：勾选
   - **参数名**：`warehouse`，**标签**：`仓库`，**字段类型**：选择 `Link`，**选项**：输入 `Warehouse`，**必填**：勾选
   - **参数名**：`company`，**标签**：`公司`，**字段类型**：选择 `Link`，**选项**：输入 `Company`，**必填**：勾选

### 2.6 保存报表

1. 检查所有信息填写无误
2. 点击右上角的 **"保存"**（Save）按钮
3. 系统会提示保存成功

---

## 三、使用报表

### 3.1 打开报表

1. 在报表列表中，找到 **"每日库存差异分析报表"**
2. 点击报表名称，打开报表页面

### 3.2 设置查询条件

在报表页面顶部，会显示查询参数输入框：

1. **日期**：选择要查询的日期（默认是今天）
2. **仓库**：选择要查询的仓库（门店）
3. **公司**：选择要查询的公司

### 3.3 运行报表

1. 填写完所有必填参数后
2. 点击 **"运行"**（Run）或 **"刷新"**（Refresh）按钮
3. 系统会执行 SQL 查询并显示结果

### 3.4 查看结果

报表会以表格形式显示所有物料的库存差异数据，包括：

- 物料基本信息（编码、名称、分组、单位）
- 初始库存数量
- 销售数量
- 理论库存数量
- 实际库存数量
- 差异数量

---

## 四、添加图表

### 4.1 创建图表

ERPNext 允许在报表中添加图表，以直观展示数据。

**步骤**：

1. **在报表页面创建图表**：
   - 运行报表后，在报表页面找到 **"图表"**（Chart）标签页
   - 或者点击报表右上角的 **"图表"**（Chart）按钮

2. **如果图表区域不存在，创建新图表**：
   - 点击报表页面右上角的 **"菜单"**（Menu）按钮（三个点图标）
   - 选择 **"添加图表"**（Add Chart）或 **"创建图表"**（Create Chart）
   - 或者导航至：**主页 > 设置 > 图表（Charts） > 新建（New）**

### 4.2 配置图表参数

在图表配置界面中：

#### 图表 1：库存对比柱状图（推荐）

**用途**：展示各物料的初始库存、理论库存、实际库存对比

**配置项**：

1. **基本信息**：
   - **图表名称**：`库存对比图表`
   - **图表标题**：`库存对比（理论 vs 实际）`

2. **数据源配置**：
   - **报表类型**：`Report`
   - **报表名称**：选择 `每日库存差异分析报表`
   - **筛选条件**：会自动继承当前报表的筛选条件

3. **图表类型**：
   - 选择：`Bar`（柱状图）
   - 子类型：`Grouped Bar`（分组柱状图）

4. **X 轴配置**：
   - **字段**：`item_name`（物料名称）
   - **标签**：`物料名称`
   - **排序**：`Ascending`（升序）

5. **Y 轴配置**：
   - **字段**：`理论库存数量`（theoretical_qty）
   - **标签**：`库存数量`
   - **聚合方式**：`Sum`（求和）

6. **系列配置**（多系列图表）：
   - **系列 1**：
     - **名称**：`理论库存`
     - **字段**：`理论库存数量`
     - **颜色**：`#1890ff`（蓝色）
   - **系列 2**：
     - **名称**：`实际库存`
     - **字段**：`实际库存数量`
     - **颜色**：`#52c41a`（绿色）

7. **其他配置**：
   - **图例**：✅ 显示
   - **工具提示**：✅ 显示
   - **数据标签**：❌ 不显示（避免图表拥挤）

8. 点击 **"保存"**（Save）保存图表

#### 图表 2：差异数量柱状图

**用途**：展示各物料的差异数量（盘盈/盘亏）

**配置项**：

1. **图表类型**：`Bar`（柱状图）

2. **X 轴**：
   - **字段**：`item_name`（物料名称）
   - **标签**：`物料名称`

3. **Y 轴**：
   - **字段**：`差异数量`（diff_qty）
   - **标签**：`差异数量`
   - **聚合方式**：`Sum`

4. **图表标题**：`库存差异分析`

5. **颜色配置**：
   - 正数（盘盈）：绿色
   - 负数（盘亏）：红色

6. 点击 **"保存"**（Save）保存图表

#### 图表 3：物料分组饼图

**用途**：展示不同物料分组的库存占比

**配置项**：

1. **图表类型**：`Pie`（饼图）

2. **分组字段**：
   - **字段**：`item_group`（物料分组）

3. **数值字段**：
   - **字段**：`实际库存数量`（actual_qty）
   - **聚合方式**：`Sum`

4. **图表标题**：`物料分组库存占比`

5. 点击 **"保存"**（Save）保存图表

### 4.3 查看图表

保存后，图表会显示在报表页面的图表区域中。

---

## 五、在 Dashboard 中展示

### 5.1 创建 Dashboard

1. 导航至：**主页 > Dashboard（仪表盘）**
2. 点击 **"新建仪表盘"**（New Dashboard）按钮
3. 输入 Dashboard 名称，例如：`库存差异分析仪表盘`
4. 点击 **"保存"**（Save）

### 5.2 添加报表和图表

1. 在 Dashboard 编辑模式下，点击 **"添加组件"**（Add Widget）按钮
2. 选择 **"报表"**（Report）类型
3. 选择 **"每日库存差异分析报表"**
4. 配置筛选条件（日期、仓库、公司）
5. 保存后，报表会显示在 Dashboard 中

6. 继续添加图表：
   - 点击 **"添加组件"**（Add Widget）按钮
   - 选择 **"图表"**（Chart）类型
   - 选择之前创建的图表
   - 保存后，图表会显示在 Dashboard 中

### 5.3 调整布局

- 拖拽报表和图表调整位置
- 调整组件大小
- 保存 Dashboard

---

## 六、测试和验证

### 6.1 测试 SQL 查询

在创建报表之前，可以先在 ERPNext 的数据库中直接测试 SQL 查询：

1. **进入数据库查询工具**：
   - 在 ERPNext 中，进入 **"开发者工具"**（Developer Tools）
   - 选择 **"查询"**（Query）或 **"控制台"**（Console）

2. **替换参数**：
   - 将 SQL 中的 `%(date)s` 替换为实际日期，如 `'2025-01-17'`
   - 将 `%(warehouse)s` 替换为实际仓库编码，如 `'WH-001'`
   - 将 `%(company)s` 替换为实际公司名称，如 `'Company A'`

3. **执行查询**：
   - 复制 SQL 查询
   - 替换参数后执行
   - 检查结果是否正确

### 6.2 验证数据准确性

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
ORDER BY posting_date DESC, posting_time DESC, creation DESC
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
ORDER BY posting_time DESC, creation DESC
LIMIT 1;
```

---

## 七、常见问题

### Q1: SQL 语法错误

**问题**：保存报表时提示 SQL 语法错误

**解决方案**：
1. 检查 SQL 语句中的表名是否正确（注意空格：`tabStock Ledger Entry`）
2. 检查字段名是否正确（注意大小写）
3. 检查参数占位符是否正确（`%(date)s`、`%(warehouse)s`、`%(company)s`）
4. 检查 ERPNext 版本是否支持相关 SQL 语法

### Q2: 查询结果为空

**问题**：运行报表后没有数据显示

**解决方案**：
1. 检查指定日期是否有库存变动记录
2. 检查指定仓库是否存在
3. 检查指定公司是否正确
4. 检查物料是否有对应的 Item 记录
5. 使用测试 SQL（6.2 节）验证数据是否存在

### Q3: 初始库存计算错误

**问题**：初始库存显示为 0，但实际应该有库存

**解决方案**：
1. 检查前一天是否有库存变动记录（使用 6.2 节的验证 SQL）
2. 如果物料是当天新增的，初始库存确实为 0
3. 检查排序逻辑是否正确（`ORDER BY posting_date DESC, posting_time DESC, creation DESC`）
4. 如果前一天有多条记录，确保获取的是最后一条

### Q4: 销售数量计算错误

**问题**：销售数量不准确

**解决方案**：
1. 检查 `voucher_type` 是否正确（应该是 `Sales Invoice`）
2. 检查 `actual_qty` 是否为负数（出库为负数）
3. 如果 ERPNext 中使用了其他销售单据类型（如 `Delivery Note`），需要修改 SQL 添加相应条件：
   ```sql
   AND sle_sales.voucher_type IN ('Sales Invoice', 'Delivery Note')
   ```
4. 使用测试 SQL（6.2 节）验证销售记录是否正确

### Q5: 实际库存计算错误

**问题**：实际库存显示不正确

**解决方案**：
1. 检查当天是否有 `Stock Reconciliation` 记录（使用 6.2 节的验证 SQL）
2. 检查 `qty_after_transaction` 字段是否有值
3. 如果没有盘点记录，系统会使用当天最后一条记录，可能不准确
4. 检查排序逻辑是否正确（`ORDER BY posting_time DESC, creation DESC`）

### Q6: 字段不存在错误

**问题**：提示字段 `out_qty` 不存在

**解决方案**：
- ERPNext 数据库中**没有** `out_qty` 字段
- 使用 `actual_qty` 字段，通过正负值判断出入库：
  - `actual_qty < 0`：出库
  - `actual_qty > 0`：入库
- 参考字段映射说明（1.4 节）

### Q7: 图表显示"无数据"

**问题**：图表显示"无数据"

**解决方案**：
1. 检查筛选条件是否过于严格
2. 检查报表数据是否存在
3. 检查字段映射是否正确
4. 确认 X 轴和 Y 轴字段是否存在

---

## 八、性能优化建议

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

## 九、总结

通过以上步骤，您可以在 ERPNext v15（Frappe Cloud）中创建一个完整的每日库存差异分析报表，包括：

✅ **表格展示**：清晰的物料库存差异数据  
✅ **图表展示**：直观的差异数量可视化  
✅ **参数过滤**：灵活的日期、仓库、公司筛选  
✅ **导出功能**：支持 Excel、CSV 导出  
✅ **Dashboard 集成**：可以在 Dashboard 中集中展示

**注意事项**：
- Frappe Cloud 不支持 Server Scripts，只能使用 Query Report
- SQL 查询性能取决于数据量，建议添加索引优化
- 如果查询结果不准确，需要根据实际业务逻辑调整 SQL
- 初始库存通过查询前一天最后一条记录获取，确保准确性

---

**最后更新**：2025-01-17  
**维护者**：TTPOS Team

