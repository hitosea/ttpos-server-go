# 查找没有估值率或成本价的物品

> 查找所有门店仓库中哪些物品没有估值率或成本价，以及如何通过 ERPNext 盘点解决这个问题

---

## 一、问题说明

### 1.1 ERPNext 约束

ERPNext 在创建库存盘点单时，**默认要求估值率（Valuation Rate）不能为 0**。

### 1.2 当前系统处理

TTPOS 系统在创建盘点单时，如果物品的估值率为 0，会自动设置 `AllowZeroValuationRate = 1`，以便 ERPNext 接受该盘点记录。

但是，**最佳实践是为所有物品设置合理的估值率或成本价**，以便：
- 准确计算库存价值
- 生成准确的财务报表
- 避免盘点时的额外配置

---

## 二、查找没有估值率的物品

### 2.1 方法一：通过 ERPNext 界面查询

1. 登录 ERPNext 系统
2. 进入 **Items** 模块
3. 点击 **Reports** > **Item Price List**
4. 查看哪些物品的 **Valuation Rate**、**Standard Rate**、**Last Purchase Rate** 都为 0

### 2.2 方法二：通过 ERPNext API 查询

使用 ERPNext API 查询所有物品，检查估值率：

```python
# Python 示例
import requests

# ERPNext API 配置
erpnext_url = "https://your-erpnext-instance.com"
api_key = "your-api-key"
api_secret = "your-api-secret"

# 查询所有物品
response = requests.get(
    f"{erpnext_url}/api/resource/Item",
    auth=(api_key, api_secret),
    params={
        "fields": '["name", "item_name", "valuation_rate", "standard_rate", "last_purchase_rate"]',
        "filters": '[["is_stock_item", "=", 1]]',
        "limit_page_length": 9999
    }
)

items = response.json()["data"]

# 查找没有估值率的物品
items_without_valuation = []
for item in items:
    valuation_rate = item.get("valuation_rate", 0) or 0
    standard_rate = item.get("standard_rate", 0) or 0
    last_purchase_rate = item.get("last_purchase_rate", 0) or 0
    
    if valuation_rate == 0 and standard_rate == 0 and last_purchase_rate == 0:
        items_without_valuation.append({
            "item_code": item["name"],
            "item_name": item.get("item_name", ""),
            "valuation_rate": valuation_rate,
            "standard_rate": standard_rate,
            "last_purchase_rate": last_purchase_rate
        })

print(f"发现 {len(items_without_valuation)} 个物品没有估值率或成本价")
for item in items_without_valuation:
    print(f"  - {item['item_code']}: {item['item_name']}")
```

### 2.3 方法三：通过 SQL 查询（如果可访问 ERPNext 数据库）

```sql
-- 查询所有没有估值率、标准价格、最近采购价的物品
SELECT 
    name AS item_code,
    item_name,
    valuation_rate,
    standard_rate,
    last_purchase_rate,
    CASE 
        WHEN valuation_rate > 0 THEN '有估值率'
        WHEN standard_rate > 0 THEN '有标准价格'
        WHEN last_purchase_rate > 0 THEN '有采购价格'
        ELSE '无价格信息'
    END AS price_status
FROM 
    `tabItem`
WHERE 
    is_stock_item = 1
    AND disabled = 0
    AND (valuation_rate IS NULL OR valuation_rate = 0)
    AND (standard_rate IS NULL OR standard_rate = 0)
    AND (last_purchase_rate IS NULL OR last_purchase_rate = 0)
ORDER BY 
    item_name;
```

### 2.4 方法四：通过 TTPOS 盘点单检查

在创建盘点单时，系统会自动检查物品的估值率。如果物品没有估值率，会在日志中记录警告信息。

---

## 三、解决方案

### 3.1 方案一：在 ERPNext 中直接设置估值率（推荐）

**步骤：**

1. 登录 ERPNext 系统
2. 进入 **Items** 模块
3. 打开需要设置估值率的物品
4. 在 **Pricing** 部分设置：
   - **Valuation Rate**：估值率（优先使用）
   - **Standard Rate**：标准价格（备选）
   - **Last Purchase Rate**：最近采购价（备选）

**优点：**
- 一次设置，永久生效
- 所有盘点单都会自动使用该估值率
- 数据准确，便于财务管理

**缺点：**
- 需要逐个物品设置，工作量大
- 如果物品很多，需要批量操作

### 3.2 方案二：通过盘点单设置估值率（临时解决）

**步骤：**

1. 在创建盘点单时，为每个物品明细指定 `valuation_rate` 字段
2. 系统会优先使用请求中的估值率
3. 如果未指定，则从 Item 获取

**代码示例：**

```go
// 创建盘点单时指定估值率
items := []*stock.StockReconciliationItem{
    {
        ItemCode:      "MAT-001",
        ItemName:      "大米",
        Qty:           100.0,
        ValuationRate: 5.50, // 明确指定估值率
    },
    {
        ItemCode:      "MAT-002",
        ItemName:      "面粉",
        Qty:           50.0,
        ValuationRate: 3.20, // 明确指定估值率
    },
}

req := &stock.SaveStockReconciliationReq{
    CompanyAbbr: "COMPANY-A",
    Branch:      "BRANCH-1",
    Warehouse:   "WH-001",
    PostingDate: "2025-01-16",
    Items:       items,
}
```

**优点：**
- 可以临时解决盘点问题
- 不需要修改 ERPNext 中的物品主数据

**缺点：**
- 每次盘点都需要指定
- 如果忘记指定，仍然会使用 Item 中的估值率（可能为 0）

### 3.3 方案三：批量更新估值率

如果物品很多，可以通过 ERPNext API 批量更新：

```python
# Python 示例：批量更新估值率
import requests

erpnext_url = "https://your-erpnext-instance.com"
api_key = "your-api-key"
api_secret = "your-api-secret"

# 物品估值率映射（从采购订单或其他来源获取）
item_valuation_map = {
    "MAT-001": 5.50,
    "MAT-002": 3.20,
    "MAT-003": 8.00,
    # ... 更多物品
}

# 批量更新
for item_code, valuation_rate in item_valuation_map.items():
    data = {
        "valuation_rate": valuation_rate
    }
    
    response = requests.put(
        f"{erpnext_url}/api/resource/Item/{item_code}",
        auth=(api_key, api_secret),
        json=data
    )
    
    if response.status_code == 200:
        print(f"✅ 更新成功: {item_code} = {valuation_rate}")
    else:
        print(f"❌ 更新失败: {item_code} - {response.text}")
```

---

## 四、估值率获取优先级

在创建盘点单时，系统会按以下优先级获取估值率：

```
优先级 1: 请求值（如果 > 0）
    ↓
优先级 2: Item.ValuationRate（如果 > 0）
    ↓
优先级 3: Item.StandardRate（如果 > 0）
    ↓
优先级 4: Item.LastPurchaseRate（如果 > 0）
    ↓
优先级 5: 如果所有价格都是 0，返回 0 并设置 AllowZeroValuationRate = 1
```

**代码位置：**
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go::getValuationRate()`

---

## 五、最佳实践

### 5.1 新物品创建时

在创建新物品时，**必须设置估值率或成本价**：

1. 如果有采购价格，设置 **Last Purchase Rate**
2. 如果有标准价格，设置 **Standard Rate**
3. 如果有估值率，设置 **Valuation Rate**

### 5.2 盘点前检查

在创建盘点单之前，建议先检查物品的估值率：

1. 查询没有估值率的物品列表
2. 为这些物品设置估值率
3. 然后再创建盘点单

### 5.3 定期维护

建议定期（如每月）检查并更新物品的估值率：

1. 查询所有物品的估值率
2. 根据最新的采购价格更新估值率
3. 确保所有物品都有合理的估值率

---

## 六、常见问题

### 6.1 为什么 ERPNext 要求估值率不能为 0？

估值率用于计算库存价值，如果为 0，会导致：
- 库存价值计算不准确
- 财务报表数据错误
- 成本核算出现问题

### 6.2 AllowZeroValuationRate 的作用是什么？

`AllowZeroValuationRate = 1` 告诉 ERPNext 允许该盘点记录使用零估值率。这是临时解决方案，**不建议长期使用**。

### 6.3 如何获取物品的合理估值率？

可以从以下来源获取：
1. **最近采购价格**：从采购订单中获取
2. **标准价格**：物品的标准定价
3. **市场价格**：当前市场价格
4. **成本价格**：物品的实际成本

### 6.4 如果物品确实没有成本（如免费样品），怎么办？

对于确实没有成本的物品（如免费样品、赠品等）：
1. 可以设置一个很小的估值率（如 0.01）
2. 或者使用 `AllowZeroValuationRate = 1`（不推荐）
3. 建议在物品名称或描述中标注"免费样品"

---

## 七、相关文档

- [盘点单 TTPOS 与 ERPNext 数据同步机制](../human/business/stock-reconciliation-erp-sync.md)
- [ERPNext 库存差异报表创建指南](./stock-difference-report-guide.md)
- [ERPNext 文档模板指南](./document-template-guide.md)

---

**最后更新**：2025-01-16  
**维护者**：TTPOS Team
















