-- ============================================================================
-- 查询没有估值率或成本价的物品
-- 
-- 用途：查找所有门店仓库中哪些物品没有估值率（Valuation Rate）、
--       标准价格（Standard Rate）或最近采购价（Last Purchase Rate）
-- 
-- 使用方法：
--   1. 连接到 ERPNext 数据库
--   2. 执行此 SQL 查询
--   3. 查看结果，为没有价格的物品设置估值率
-- 
-- 注意：此查询需要访问 ERPNext 数据库，如果没有权限，请使用 ERPNext API
-- ============================================================================

-- 查询所有没有估值率、标准价格、最近采购价的物品
SELECT 
    name AS item_code,
    item_name,
    item_group,
    stock_uom AS unit,
    valuation_rate,
    standard_rate,
    last_purchase_rate,
    CASE 
        WHEN valuation_rate > 0 THEN '有估值率'
        WHEN standard_rate > 0 THEN '有标准价格'
        WHEN last_purchase_rate > 0 THEN '有采购价格'
        ELSE '无价格信息'
    END AS price_status,
    disabled,
    is_stock_item
FROM 
    `tabItem`
WHERE 
    is_stock_item = 1  -- 只查询库存物品
    AND disabled = 0   -- 只查询未禁用的物品
    AND (
        -- 估值率为空或0
        (valuation_rate IS NULL OR valuation_rate = 0)
        -- 标准价格为空或0
        AND (standard_rate IS NULL OR standard_rate = 0)
        -- 最近采购价为空或0
        AND (last_purchase_rate IS NULL OR last_purchase_rate = 0)
    )
ORDER BY 
    item_group, item_name;

-- ============================================================================
-- 按仓库查询没有估值率的物品（需要关联库存表）
-- ============================================================================

SELECT 
    i.name AS item_code,
    i.item_name,
    i.item_group,
    i.stock_uom AS unit,
    i.valuation_rate,
    i.standard_rate,
    i.last_purchase_rate,
    sle.warehouse,
    SUM(sle.actual_qty) AS current_stock,
    CASE 
        WHEN i.valuation_rate > 0 THEN '有估值率'
        WHEN i.standard_rate > 0 THEN '有标准价格'
        WHEN i.last_purchase_rate > 0 THEN '有采购价格'
        ELSE '无价格信息'
    END AS price_status
FROM 
    `tabItem` i
INNER JOIN 
    `tabStock Ledger Entry` sle ON i.name = sle.item_code
WHERE 
    i.is_stock_item = 1
    AND i.disabled = 0
    AND (
        (i.valuation_rate IS NULL OR i.valuation_rate = 0)
        AND (i.standard_rate IS NULL OR i.standard_rate = 0)
        AND (i.last_purchase_rate IS NULL OR i.last_purchase_rate = 0)
    )
    AND sle.actual_qty > 0  -- 只查询有库存的物品
GROUP BY 
    i.name, i.item_name, i.item_group, i.stock_uom, 
    i.valuation_rate, i.standard_rate, i.last_purchase_rate, sle.warehouse
ORDER BY 
    sle.warehouse, i.item_group, i.item_name;

-- ============================================================================
-- 统计没有估值率的物品数量（按公司分组）
-- ============================================================================

SELECT 
    i.custom_company AS company,
    i.custom_branch AS branch,
    COUNT(DISTINCT i.name) AS items_without_price,
    COUNT(DISTINCT CASE WHEN sle.actual_qty > 0 THEN i.name END) AS items_with_stock
FROM 
    `tabItem` i
LEFT JOIN 
    `tabStock Ledger Entry` sle ON i.name = sle.item_code
WHERE 
    i.is_stock_item = 1
    AND i.disabled = 0
    AND (
        (i.valuation_rate IS NULL OR i.valuation_rate = 0)
        AND (i.standard_rate IS NULL OR i.standard_rate = 0)
        AND (i.last_purchase_rate IS NULL OR i.last_purchase_rate = 0)
    )
GROUP BY 
    i.custom_company, i.custom_branch
ORDER BY 
    items_without_price DESC;

-- ============================================================================
-- 查询没有估值率的物品，并显示最近一次采购价格（如果有）
-- ============================================================================

SELECT 
    i.name AS item_code,
    i.item_name,
    i.valuation_rate,
    i.standard_rate,
    i.last_purchase_rate,
    poi.rate AS last_purchase_order_rate,
    po.transaction_date AS last_purchase_date,
    po.name AS last_purchase_order
FROM 
    `tabItem` i
LEFT JOIN (
    SELECT 
        item_code,
        rate,
        parent,
        MAX(idx) AS max_idx
    FROM 
        `tabPurchase Order Item`
    GROUP BY 
        item_code, parent
) poi ON i.name = poi.item_code
LEFT JOIN 
    `tabPurchase Order` po ON poi.parent = po.name
WHERE 
    i.is_stock_item = 1
    AND i.disabled = 0
    AND (
        (i.valuation_rate IS NULL OR i.valuation_rate = 0)
        AND (i.standard_rate IS NULL OR i.standard_rate = 0)
        AND (i.last_purchase_rate IS NULL OR i.last_purchase_rate = 0)
    )
ORDER BY 
    i.item_name;
















