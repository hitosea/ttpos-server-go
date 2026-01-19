# 每日库存汇总表查看与展示方案

> 明确 BOM 消耗与每日盘点数据的查看方式及展示方案

---

## 一、需求背景

华莱士客户近期已持续进行库存盘点，并计划在本周一开始正式核对相关数据。

为避免临时排查、信息不清，需提前明确 BOM 消耗与每日盘点数据的查看方式及展示方案。

---

## 二、库存汇总表结构

### 2.1 表头字段

| 字段名 | 字段说明 | 数据类型 | 示例值 |
|--------|----------|----------|--------|
| `item_code` | 物料编码 | string | F001 |
| `item_name` | 物料名称 | string | 香辣鸡腿 |
| `item_group` | 物料分组 | string | 食材 |
| `opening_qty` | 初始库存数量 | decimal | 120 |
| `sales_qty` | 销售额_数量（累计销量） | decimal | 85 |
| `theoretical_qty` | 理论库存数量 | decimal | 35 |
| `actual_qty` | 实际库存数量 | decimal | 32 |
| `diff_qty` | 差异数量 | decimal | -3 |
| `stock_uom` | 物料单位 | string | 个 |

### 2.2 字段计算公式

- **理论库存数量** (`theoretical_qty`) = 初始库存数量 (`opening_qty`) - 累计销量 (`sales_qty`)
- **差异数量** (`diff_qty`) = 实际库存数量 (`actual_qty`) - 理论库存数量 (`theoretical_qty`)

### 2.3 示例数据

| item_code | item_name | item_group | opening_qty | sales_qty | theoretical_qty | actual_qty | diff_qty | stock_uom |
|-----------|-----------|------------|-------------|-----------|-----------------|------------|----------|-----------|
| F001 | 香辣鸡腿 | 食材 | 120 | 85 | 35 | 32 | -3 | 个 |
| F002 | 汉堡面包 | 食材 | 200 | 130 | 70 | 68 | -2 | 个 |
| F003 | 薯条（大包） | 食材 | 50 | 38 | 12 | 10 | -2 | 包 |
| P001 | 可乐杯（大） | 包装 | 500 | 320 | 180 | 175 | -5 | 个 |
| P002 | 纸袋 | 包装 | 300 | 210 | 90 | 85 | -5 | 个 |
| C001 | 餐厅清洁剂 | 清洁 | 20 | 5 | 15 | 14 | -1 | 瓶 |
| F004 | 鸡翅（对） | 食材 | 80 | 60 | 20 | 18 | -2 | 对 |
| P003 | 吸管 | 包装 | 1000 | 750 | 250 | 240 | -10 | 根 |

---

## 三、数据来源说明

### 3.1 初始库存数量 (`opening_qty`)

**数据来源**：当日营业开始时的仓库物品库存数量

**获取方式**：
- 从 `warehouse_item` 表获取指定日期开始时的库存数量
- 或从当日第一个盘点单的账面库存数量获取

**相关代码位置**：
- `main/app/model/warehouse_item.go` - 仓库物品库存模型
- `main/app/repository/warehouse_item.go` - 仓库物品库存查询

### 3.2 累计销量 (`sales_qty`)

**数据来源**：当日通过 BOM 消耗计算得出的物料消耗总量

**计算逻辑**：
1. 查询当日所有已完成的销售订单（`sale_order` 表，`status = 1`，`finish_time` 在当日范围内）
2. 遍历订单中的商品（`sale_order_product`），排除已删除、已取消、未送厨、套餐商品、未接单商品
3. 根据商品的 BOM 配置（`product_bom`）计算物料消耗：
   - 如果商品有成本卡（`product_bom_card`），使用成本卡的关联材料（`related_materials`）
   - 如果商品是规格商品（`is_flavor`），使用规格关联材料（`flavor_materials`）
   - 如果商品是小料（`is_sauce`），使用小料关联材料（`sauce_materials`）
4. 累加所有订单中该物料的消耗数量（已换算为基准单位）

**计算公式**：
```
物料消耗量 = Σ(商品数量 × BOM 中该物料的用量)
```

**相关代码位置**：
- `main/app/model/sale_order.go::GetValidSaleOrderProductMaterialList()` - 获取订单物料消耗列表
- `main/app/model/product.go::GetDecreaseNum()` - 计算物料减少数量
- `main/app/service/order.go` - 订单服务中的库存扣减逻辑

**注意事项**：
- 所有数量需要换算为物料的基准单位（`base_unit`）
- 如果物料由成本卡管理（`unit_uuid != 0`），需要考虑单位换算率（`base_unit_conversion_rate`）

### 3.3 理论库存数量 (`theoretical_qty`)

**计算公式**：
```
理论库存数量 = 初始库存数量 - 累计销量
```

**说明**：
- 理论库存数量 = 当日开始时库存 - 当日消耗的物料数量
- 如果计算结果为负数，表示理论库存不足（可能存在库存异常）

### 3.4 实际库存数量 (`actual_qty`)

**数据来源**：当日盘点单中的实盘数量

**获取方式**：
1. 查询指定日期已审核的盘点单（`stock_reconciliation` 表，`status = 2`）
2. 从盘点单明细（`stock_reconciliation_item`）中获取实盘数量（`counted_quantity`）
3. 如果当日有多个盘点单，取最后一个已审核的盘点单数据

**相关代码位置**：
- `main/app/model/stock_reconciliation.go` - 盘点单模型
- `main/app/service/stock_reconciliation.go` - 盘点单服务

**注意事项**：
- 实盘数量已换算为基准单位
- 如果当日没有盘点单，实际库存数量为空或使用账面库存数量

### 3.5 差异数量 (`diff_qty`)

**计算公式**：
```
差异数量 = 实际库存数量 - 理论库存数量
```

**说明**：
- 正数表示盘盈（实际库存 > 理论库存）
- 负数表示盘亏（实际库存 < 理论库存）
- 零表示账实相符

---

## 四、数据查看方式

### 4.1 通过 API 查询

#### 4.1.1 查询每日库存汇总表

**接口路径**（待实现）：
```
GET /shop/stock/daily_summary
```

**请求参数**：
```json
{
  "date": "2025-01-16",           // 查询日期（YYYY-MM-DD）
  "warehouse_uuid": 123,          // 仓库UUID（可选）
  "item_group": "食材",            // 物料分组（可选）
  "page_no": 1,                   // 页码
  "page_size": 20                 // 每页数量
}
```

**响应数据**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "meta": {
      "total": 100,
      "page_no": 1,
      "page_size": 20
    },
    "list": [
      {
        "item_code": "F001",
        "item_name": "香辣鸡腿",
        "item_group": "食材",
        "opening_qty": 120.0000,
        "sales_qty": 85.0000,
        "theoretical_qty": 35.0000,
        "actual_qty": 32.0000,
        "diff_qty": -3.0000,
        "stock_uom": "个"
      }
    ]
  }
}
```

#### 4.1.2 查询 BOM 消耗明细

**接口路径**（待实现）：
```
GET /shop/stock/bom_consumption_detail
```

**请求参数**：
```json
{
  "date": "2025-01-16",           // 查询日期
  "material_uuid": 456,           // 物料UUID（可选）
  "warehouse_uuid": 123           // 仓库UUID（可选）
}
```

**响应数据**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "material_uuid": 456,
    "material_code": "F001",
    "material_name": "香辣鸡腿",
    "total_consumption": 85.0000,
    "details": [
      {
        "sale_order_no": "SO202501160001",
        "product_name": "香辣鸡腿堡",
        "product_num": 10,
        "bom_consumption": 10.0000,
        "consumption_time": 1705392000
      }
    ]
  }
}
```

### 4.2 通过数据库查询

#### 4.2.1 查询初始库存数量

```sql
-- 查询指定日期开始时的库存数量
SELECT 
    m.code AS item_code,
    m.name AS item_name,
    m.item_group,
    wi.stock_num AS opening_qty,
    m.base_unit AS stock_uom
FROM warehouse_item wi
INNER JOIN material m ON wi.material_uuid = m.uuid
WHERE wi.warehouse_uuid = ?
  AND wi.delete_time = 0
  AND m.delete_time = 0
ORDER BY m.code;
```

#### 4.2.2 查询 BOM 消耗数量

```sql
-- 查询当日物料消耗总量（通过销售出库记录）
SELECT 
    m.code AS item_code,
    m.name AS item_name,
    m.item_group,
    SUM(wil.num) AS sales_qty,
    m.base_unit AS stock_uom
FROM warehouse_in_out_log wil
INNER JOIN material m ON wil.material_uuid = m.uuid
WHERE wil.warehouse_uuid = ?
  AND wil.log_type = 1  -- 出库
  AND wil.scene = 1     -- 销售出库
  AND DATE(FROM_UNIXTIME(wil.create_time)) = ?
  AND wil.delete_time = 0
  AND m.delete_time = 0
GROUP BY m.uuid, m.code, m.name, m.item_group, m.base_unit
ORDER BY m.code;
```

#### 4.2.3 查询实际库存数量

```sql
-- 查询当日已审核盘点单的实盘数量
SELECT 
    m.code AS item_code,
    m.name AS item_name,
    m.item_group,
    sri.counted_quantity AS actual_qty,
    m.base_unit AS stock_uom
FROM stock_reconciliation_item sri
INNER JOIN stock_reconciliation sr ON sri.stock_reconciliation_uuid = sr.uuid
INNER JOIN material m ON sri.material_uuid = m.uuid
WHERE sr.warehouse_uuid = ?
  AND sr.status = 2  -- 已审核
  AND DATE(FROM_UNIXTIME(sr.create_time)) = ?
  AND sri.delete_time = 0
  AND sr.delete_time = 0
  AND m.delete_time = 0
ORDER BY m.code;
```

---

## 五、前端展示方案

### 5.1 页面布局

**页面路径**（待实现）：
```
/shop/inventory/daily-summary
```

**页面结构**：
1. **查询条件区域**
   - 日期选择器（默认当天）
   - 仓库选择器（可选）
   - 物料分组筛选（可选）
   - 查询按钮

2. **汇总信息卡片**
   - 总物料数
   - 盘盈物料数
   - 盘亏物料数
   - 账实相符物料数

3. **数据表格**
   - 显示库存汇总表的所有字段
   - 支持排序（按差异数量、物料编码等）
   - 支持导出 Excel

### 5.2 表格列配置

| 列名 | 字段 | 宽度 | 对齐方式 | 说明 |
|------|------|------|----------|------|
| 物料编码 | `item_code` | 120px | 左对齐 | - |
| 物料名称 | `item_name` | 200px | 左对齐 | - |
| 物料分组 | `item_group` | 100px | 左对齐 | - |
| 初始库存 | `opening_qty` | 120px | 右对齐 | 显示单位 |
| 累计销量 | `sales_qty` | 120px | 右对齐 | 显示单位 |
| 理论库存 | `theoretical_qty` | 120px | 右对齐 | 显示单位 |
| 实际库存 | `actual_qty` | 120px | 右对齐 | 显示单位 |
| 差异数量 | `diff_qty` | 120px | 右对齐 | 显示单位，负数标红 |
| 单位 | `stock_uom` | 80px | 居中 | - |

### 5.3 数据展示规则

1. **差异数量颜色标识**：
   - 负数（盘亏）：红色显示
   - 正数（盘盈）：绿色显示
   - 零（账实相符）：黑色显示

2. **数量格式化**：
   - 保留 4 位小数
   - 显示物料单位（如：120.0000 个）

3. **空值处理**：
   - 如果当日没有盘点单，实际库存数量显示为 "-"
   - 如果当日没有销售，累计销量显示为 "0.0000"

### 5.4 交互功能

1. **点击物料编码**：跳转到物料详情页
2. **点击累计销量**：弹出 BOM 消耗明细弹窗
3. **点击实际库存**：跳转到盘点单详情页（如果有）
4. **导出 Excel**：导出当前查询结果到 Excel 文件

---

## 六、实现建议

### 6.1 后端实现

#### 6.1.1 创建 Service 方法

**文件位置**：`main/app/service/stock.go`

**方法签名**：
```go
// GetDailyStockSummary 获取每日库存汇总表
func (s *stockSrv) GetDailyStockSummary(ctx context.Context, req req.DailyStockSummaryReq) (resp.DailyStockSummaryResp, error)
```

**实现步骤**：
1. 查询初始库存数量（从 `warehouse_item` 表或当日第一个盘点单）
2. 查询累计销量（从 `warehouse_in_out_log` 表，`log_type = 1`, `scene = 1`）
3. 计算理论库存数量
4. 查询实际库存数量（从当日已审核盘点单）
5. 计算差异数量
6. 合并数据并返回

#### 6.1.2 创建 API 接口

**文件位置**：`main/app/api/v1/shop/shop_stock.go`

**接口定义**：
```go
// GetDailyStockSummary 获取每日库存汇总表
// @Summary 获取每日库存汇总表
// @Description 查询指定日期的库存汇总数据，包含初始库存、累计销量、理论库存、实际库存、差异数量
// @Tags 商家端.库存管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param date query string true "查询日期（YYYY-MM-DD）"
// @Param warehouse_uuid query int false "仓库UUID"
// @Param item_group query string false "物料分组"
// @Param page_no query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} dto.Response{data=resp.DailyStockSummaryResp} "成功"
// @Router /shop/stock/daily_summary [get]
func (h *StockHandler) GetDailyStockSummary(c *gin.Context)
```

### 6.2 前端实现

#### 6.2.1 创建页面组件

**文件位置**：`admin/views/shop/src/views/inventory/daily-summary/index.vue`

**主要功能**：
- 日期选择器
- 仓库选择器
- 物料分组筛选
- 数据表格展示
- 导出 Excel 功能

#### 6.2.2 创建 API 调用方法

**文件位置**：`admin/views/shop/src/api/inventory.ts`

**方法定义**：
```typescript
// 获取每日库存汇总表
export function getDailyStockSummary(params: {
  date: string;
  warehouse_uuid?: number;
  item_group?: string;
  page_no?: number;
  page_size?: number;
}) {
  return request.get('/shop/stock/daily_summary', { params });
}
```

---

## 七、注意事项

### 7.1 数据准确性

1. **初始库存数量**：
   - 需要明确是当日营业开始时的库存，还是当日第一个盘点单的账面库存
   - 建议使用当日第一个盘点单的账面库存，确保数据一致性

2. **累计销量**：
   - 必须使用销售出库记录（`warehouse_in_out_log`），而不是直接计算 BOM 消耗
   - 确保只统计已完成的订单（`sale_order.status = 1`）

3. **实际库存数量**：
   - 如果当日有多个盘点单，需要明确使用哪个盘点单的数据
   - 建议使用最后一个已审核的盘点单数据

### 7.2 性能优化

1. **数据缓存**：
   - 对于历史日期的数据，可以缓存计算结果
   - 当日数据实时计算

2. **查询优化**：
   - 使用索引优化查询性能
   - 避免全表扫描

3. **分页查询**：
   - 支持分页查询，避免一次性加载大量数据

### 7.3 边界情况处理

1. **当日没有盘点单**：
   - 实际库存数量显示为空或使用账面库存数量
   - 差异数量无法计算，显示为 "-"

2. **当日没有销售**：
   - 累计销量为 0
   - 理论库存数量 = 初始库存数量

3. **物料单位不一致**：
   - 确保所有数量都换算为基准单位
   - 显示时使用物料的基准单位

---

## 八、通过 Stock Ledger 和 Set Chart 实现方案

### 8.1 方案概述

**✅ 推荐方案**：通过 ERPNext 的 **Stock Ledger（库存分类账）** 查询数据，并通过 **Set Chart（图表）** 或 **ECharts** 进行可视化展示。

**优势**：
- ✅ 利用 ERPNext 现有的 Stock Ledger 报表，数据准确可靠
- ✅ 无需重复开发数据查询逻辑
- ✅ 支持按日期范围、仓库、物品编码等条件灵活查询
- ✅ 可以通过图表直观展示库存趋势和差异

### 8.2 Stock Ledger 数据映射

#### 8.2.1 Stock Ledger 字段说明

ERPNext 的 Stock Ledger 报表包含以下关键字段：

| Stock Ledger 字段 | 说明 | 对应需求字段 |
|------------------|------|-------------|
| `item_code` | 物料编码 | `item_code` |
| `item_name` | 物料名称 | `item_name` |
| `item_group` | 物料分组 | `item_group` |
| `warehouse` | 仓库 | - |
| `posting_date` | 过账日期 | - |
| `posting_time` | 过账时间 | - |
| `in_qty` | 入库数量 | - |
| `out_qty` | 出库数量 | `sales_qty`（累计出库） |
| `qty_after_transaction` | 交易后数量 | `theoretical_qty`（当日最后一笔交易后数量） |
| `actual_qty` | 实际数量 | `actual_qty`（盘点单中的实盘数量） |
| `stock_uom` | 库存单位 | `stock_uom` |

#### 8.2.2 数据计算逻辑

**通过 Stock Ledger 计算库存汇总表**：

1. **初始库存数量** (`opening_qty`)：
   - 查询当日第一笔交易前的库存数量
   - 或查询上一日最后一笔交易的 `qty_after_transaction`

2. **累计销量** (`sales_qty`)：
   - 汇总当日所有 `out_qty`（出库数量）
   - 过滤 `voucher_type` 为销售相关的凭证类型（如 "Sales Invoice", "Delivery Note" 等）

3. **理论库存数量** (`theoretical_qty`)：
   - 使用当日最后一笔交易的 `qty_after_transaction`
   - 或计算：`opening_qty - sales_qty`

4. **实际库存数量** (`actual_qty`)：
   - 查询当日盘点单（`voucher_type = "Stock Reconciliation"`）的 `actual_qty`
   - 如果没有盘点单，使用当日最后一笔交易的 `qty_after_transaction`

5. **差异数量** (`diff_qty`)：
   - 计算：`actual_qty - theoretical_qty`

### 8.3 通过 Stock Ledger API 查询

#### 8.3.1 调用 BMP 模块 Stock Ledger 接口

**接口路径**：
```
gRPC: /stock.StockService/GetStockLedger
```

**请求参数**：
```go
type GetStockLedgerReq struct {
    CompanyAbbr string  // 公司缩写，必填
    FromDate    string  // 开始日期（YYYY-MM-DD），必填
    ToDate      string  // 结束日期（YYYY-MM-DD），必填
    Branch      string  // 分支机构（可选）
    Warehouse   string  // 仓库名称（可选）
    ItemCode    string  // 物品编码（可选）
    VoucherNo   string  // 凭证编号（可选）
    Limit       int32   // 查询限制数量（可选，默认100，最大1000）
}
```

**响应数据**：
```go
type GetStockLedgerResp struct {
    StockLedgerList []*StockLedger  // 库存分类账列表
}

type StockLedger struct {
    ItemCode             string   // 物品编码
    ItemName             string   // 物品名称
    ItemGroup            string   // 物品分组
    Warehouse            string   // 仓库
    PostingDate          string   // 过账日期
    PostingTime          string   // 过账时间
    InQty                float64  // 入库数量
    OutQty               float64  // 出库数量
    QtyAfterTransaction  float64  // 交易后数量
    ActualQty            float64  // 实际数量
    StockUom             string   // 库存单位
    VoucherType          string   // 凭证类型
    VoucherNo            string   // 凭证编号
    // ... 其他字段
}
```

**相关代码位置**：
- BMP 模块：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go::GetStockLedger()`
- Protobuf 定义：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`

#### 8.3.2 数据聚合处理

**实现步骤**：

1. **调用 Stock Ledger API**：
   ```go
   // 查询当日库存分类账
   req := &stock.GetStockLedgerReq{
       CompanyAbbr: companyAbbr,
       FromDate:    "2025-01-16",
       ToDate:      "2025-01-16",
       Warehouse:   warehouseName,
   }
   resp, err := erpStockClient.GetStockLedger(ctx, req)
   ```

2. **数据聚合**：
   ```go
   // 按物料分组聚合数据
   itemMap := make(map[string]*DailyStockSummaryItem)
   
   for _, ledger := range resp.StockLedgerList {
       itemCode := ledger.ItemCode
       
       if itemMap[itemCode] == nil {
           itemMap[itemCode] = &DailyStockSummaryItem{
               ItemCode:  itemCode,
               ItemName:  ledger.ItemName,
               ItemGroup: ledger.ItemGroup,
               StockUom:  ledger.StockUom,
           }
       }
       
       item := itemMap[itemCode]
       
       // 累计出库数量（销售出库）
       if isSalesOutbound(ledger.VoucherType) {
           item.SalesQty += ledger.OutQty
       }
       
       // 更新交易后数量（理论库存）
       if ledger.QtyAfterTransaction != 0 {
           item.TheoreticalQty = ledger.QtyAfterTransaction
       }
       
       // 更新实际库存（盘点单）
       if ledger.VoucherType == "Stock Reconciliation" {
           item.ActualQty = ledger.ActualQty
       }
   }
   
   // 计算初始库存和差异数量
   for _, item := range itemMap {
       item.OpeningQty = item.TheoreticalQty + item.SalesQty
       item.DiffQty = item.ActualQty - item.TheoreticalQty
   }
   ```

### 8.4 通过 Set Chart 展示

#### 8.4.1 ERPNext Set Chart 方案

**Set Chart** 是 ERPNext 中的 Dashboard Chart 功能，可以在 ERPNext 界面中直接创建图表。

**详细操作指南**：请参考 [ERPNext Stock Ledger Set Chart 使用指南](../../shared/integrations/erpnext/stock-ledger-set-chart-guide.md)

**快速步骤**：

1. **在 ERPNext 中创建 Set Chart**：
   - 进入 ERPNext Dashboard
   - 创建新的 Chart
   - 选择数据源为 "Stock Ledger"
   - 配置图表类型（柱状图、折线图、饼图等）

2. **配置图表参数**：
   - X 轴：物料编码或物料分组
   - Y 轴：库存数量（初始库存、累计销量、理论库存、实际库存、差异数量）
   - 筛选条件：日期范围、仓库、物料分组等

3. **图表类型建议**：
   - **柱状图**：展示各物料的库存对比（初始库存、理论库存、实际库存）
   - **折线图**：展示库存趋势变化（按日期）
   - **饼图**：展示物料分组占比
   - **组合图**：同时展示多个指标（初始库存、累计销量、差异数量）

**注意**：由于 ERPNext Set Chart 的限制，建议创建自定义报表来展示完整的库存汇总表数据（包含计算字段如 `opening_qty`、`diff_qty`）。详细步骤请参考上述指南文档。

#### 8.4.2 前端 ECharts 方案（推荐）

**如果需要在 TTPOS 前端展示**，可以使用 ECharts 进行可视化：

**实现步骤**：

1. **获取数据**：
   ```typescript
   // 调用后端 API 获取库存汇总数据
   const data = await getDailyStockSummary({
     date: '2025-01-16',
     warehouse_uuid: 123
   });
   ```

2. **配置 ECharts 图表**：
   ```typescript
   const option = {
     title: {
       text: '每日库存汇总表'
     },
     tooltip: {
       trigger: 'axis',
       axisPointer: {
         type: 'shadow'
       }
     },
     legend: {
       data: ['初始库存', '累计销量', '理论库存', '实际库存', '差异数量']
     },
     xAxis: {
       type: 'category',
       data: data.list.map(item => item.item_name)
     },
     yAxis: {
       type: 'value'
     },
     series: [
       {
         name: '初始库存',
         type: 'bar',
         data: data.list.map(item => item.opening_qty)
       },
       {
         name: '累计销量',
         type: 'bar',
         data: data.list.map(item => item.sales_qty)
       },
       {
         name: '理论库存',
         type: 'bar',
         data: data.list.map(item => item.theoretical_qty)
       },
       {
         name: '实际库存',
         type: 'bar',
         data: data.list.map(item => item.actual_qty)
       },
       {
         name: '差异数量',
         type: 'line',
         data: data.list.map(item => item.diff_qty),
         itemStyle: {
           color: function(params) {
             // 负数标红，正数标绿
             return params.value < 0 ? '#ff4d4f' : '#52c41a';
           }
         }
       }
     ]
   };
   ```

3. **图表类型建议**：
   - **多柱状图**：展示初始库存、累计销量、理论库存、实际库存的对比
   - **折线图**：展示差异数量趋势（按日期）
   - **散点图**：展示理论库存 vs 实际库存的分布
   - **仪表盘**：展示盘盈/盘亏物料占比

**相关代码位置**：
- 前端图表组件示例：`admin/views/shop/src/views/home/part/product/Transaction.vue`
- 报损图表示例：`admin/views/shop/src/views/inventory/wastage/echartsVue.vue`

### 8.5 方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **Stock Ledger + Set Chart** | ✅ 利用 ERPNext 现有功能<br>✅ 数据准确可靠<br>✅ 无需额外开发 | ❌ 需要在 ERPNext 中配置<br>❌ 展示方式受限 | ERPNext 用户直接查看 |
| **Stock Ledger + ECharts** | ✅ 数据准确可靠<br>✅ 展示方式灵活<br>✅ 集成到 TTPOS 前端 | ❌ 需要前端开发 | TTPOS 前端展示 |
| **直接数据库查询** | ✅ 完全可控<br>✅ 性能优化空间大 | ❌ 需要重复开发<br>❌ 数据一致性风险 | 特殊需求场景 |

**推荐方案**：**Stock Ledger + ECharts**，既保证了数据准确性，又提供了灵活的展示方式。

---

## 九、相关文档

- [盘点单 TTPOS 与 ERPNext 数据同步机制](./stock-reconciliation-erp-sync.md)
- [盘点单功能产品概述](./stock-reconciliation-product-overview.md)
- [BOM 消耗计算逻辑](../architecture/features/product-bom.md)
- [ERPNext Stock Ledger Set Chart 使用指南](../../shared/integrations/erpnext/stock-ledger-set-chart-guide.md) ⭐ 详细操作步骤
- [Stock Ledger API 文档](../../shared/api/stock-ledger-api.md)（待补充）

---

**最后更新**：2025-01-17  
**维护者**：TTPOS Team

