# TTPOS 盘点单同步到 ERPNext 接口说明

> 当前盘点单同步到 ERPNext 的接口调用链路和参数说明

---

## 一、调用链路

```
TTPOS Main 模块
    ↓ (gRPC)
BMP 模块（ttpos-erp）
    ↓ (HTTP REST API)
ERPNext 系统
```

---

## 二、接口详情

### 2.1 提交盘点单（创建盘点单到 ERPNext）

#### TTPOS Main → BMP 模块（gRPC）

**接口路径**：
```
gRPC Service: stock.StockService
RPC Method: /stock.StockService/SaveStockReconciliation
```

**调用位置**：
- 文件：`main/app/service/rpc/erp/stock.go`
- 方法：`SubmitStockReconciliation()`

**请求参数**：
```go
type SaveStockReconciliationReq struct {
    CompanyAbbr string                        // 公司缩写
    Branch      string                        // 分支名称
    PostingDate string                        // 过账日期（格式：2006-01-02）
    PostingTime string                        // 过账时间（格式：15:04:05）
    Warehouse   string                        // 仓库编码（ERP Code）
    Items       []*StockReconciliationItem     // 盘点明细列表
}

type StockReconciliationItem struct {
    ItemCode string   // 物品编码
    Qty      float64  // 实盘数量（基准单位）
}
```

**响应参数**：
```go
type SaveStockReconciliationResp struct {
    StockReconciliationName string  // ERPNext 盘点单号（格式：MAT-RECO-YYYY-XXXXX）
}
```

#### BMP 模块 → ERPNext（HTTP REST API）

**接口路径**：
```
POST /api/v2/document/Stock Reconciliation
```

**调用位置**：
- 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
- 方法：`SaveStockReconciliation()`
- 实现：`ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go::Create()`

**请求头**：
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求体**：
```json
{
  "naming_series": "MAT-RECO-.YYYY.-",
  "company": "Company A",
  "posting_date": "2025-01-16",
  "posting_time": "14:30:00",
  "set_posting_time": 1,
  "purpose": "Stock Reconciliation",
  "set_warehouse": "WH-001",
  "items": [
    {
      "item_code": "MAT-001",
      "item_name": "大米",
      "qty": 100.000,
      "warehouse": "WH-001",
      "valuation_rate": 1.0,
      "doctype": "Stock Reconciliation Item"
    }
  ],
  "doctype": "Stock Reconciliation"
}
```

**响应体**：
```json
{
  "data": {
    "name": "MAT-RECO-2025-00001",
    "company": "Company A",
    "posting_date": "2025-01-16",
    "status": "Draft"
  }
}
```

---

### 2.2 审核盘点单（提交盘点单到 ERPNext）

#### TTPOS Main → BMP 模块（gRPC）

**接口路径**：
```
gRPC Service: stock.StockService
RPC Method: /stock.StockService/SubmitStockReconciliation
```

**调用位置**：
- 文件：`main/app/service/rpc/erp/stock.go`
- 方法：`ApproveStockReconciliation()`

**请求参数**：
```go
type SubmitStockReconciliationReq struct {
    StockReconciliationName string  // ERPNext 盘点单号
}
```

**响应参数**：
```go
type SubmitStockReconciliationResp struct {
    Message string  // 操作结果消息
}
```

#### BMP 模块 → ERPNext（HTTP REST API）

**接口路径**：
```
PUT /api/v2/document/Stock Reconciliation/{name}
```

**调用位置**：
- 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
- 方法：`SubmitStockReconciliation()`
- 实现：`ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go::ChangeDocStatus()`

**请求头**：
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求体**：
```json
{
  "docstatus": 1
}
```

**说明**：
- `docstatus = 0`：Draft（草稿）
- `docstatus = 1`：Submitted（已提交）
- `docstatus = 2`：Cancelled（已取消）

**响应体**：
```json
{
  "data": {
    "name": "MAT-RECO-2025-00001",
    "docstatus": 1,
    "status": "Submitted"
  }
}
```

---

## 三、代码调用示例

### 3.1 TTPOS Main 模块调用

**提交盘点单**：
```go
// 文件：main/app/service/stock_reconciliation.go
// 方法：submitStockReconciliation()

erpSrv := erp.NewIErpSrv(s.dbm)
erpReq, err := erpSrv.SubmitStockReconciliation(ctx, companySetting, &stock.SaveStockReconciliationReq{
    CompanyAbbr: companySetting.ErpnextCompanyAbbr,
    Branch:      companySetting.ErpnextBranchName,
    PostingDate: now.Format("2006-01-02"),
    PostingTime: now.Format("15:04:05"),
    Warehouse:   stockReconciliation.Warehouse.ErpCode,
    Items:       erpItems,  // []*stock.StockReconciliationItem
})
```

**审核盘点单**：
```go
// 文件：main/app/service/stock_reconciliation.go
// 方法：ApproveStockReconciliation()

erpSrv := erp.NewIErpSrv(s.dbm)
_, err := erpSrv.ApproveStockReconciliation(ctx, companySetting, &stock.SubmitStockReconciliationReq{
    StockReconciliationName: stockReconciliation.ErpCode,
})
```

### 3.2 BMP 模块实现

**创建盘点单**：
```go
// 文件：ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go
// 方法：SaveStockReconciliation()

resp, err := service.Document().Create(ctx, erp.DocTypeStockReconciliation, data)
// 实际调用：POST /api/v2/document/Stock Reconciliation
```

**提交盘点单**：
```go
// 文件：ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go
// 方法：SubmitStockReconciliation()

_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypeStockReconciliation, req.StockReconciliationName, erp.DocstatusSubmitted)
// 实际调用：PUT /api/v2/document/Stock Reconciliation/{name}
// 请求体：{"docstatus": 1}
```

---

## 四、接口映射关系

| TTPOS 操作 | TTPOS 方法 | BMP gRPC 方法 | ERPNext API | ERPNext 操作 |
|------------|------------|---------------|-------------|--------------|
| **提交盘点单** | `SubmitStockReconciliation()` | `SaveStockReconciliation` | `POST /api/v2/document/Stock Reconciliation` | 创建盘点单（Draft） |
| **审核盘点单** | `ApproveStockReconciliation()` | `SubmitStockReconciliation` | `PUT /api/v2/document/Stock Reconciliation/{name}` | 提交盘点单（Submitted） |

**注意**：
- TTPOS 的"提交"对应 ERPNext 的"创建"（Draft 状态）
- TTPOS 的"审核"对应 ERPNext 的"提交"（Submitted 状态）

---

## 五、数据映射

### 5.1 提交时的数据映射

| TTPOS 字段 | ERPNext 字段 | 说明 |
|------------|--------------|------|
| `company.erpnext_company_abbr` | `company` | 公司名称 |
| `company.erpnext_branch_name` | `branch` | 分支名称 |
| `now().Format("2006-01-02")` | `posting_date` | 过账日期 |
| `now().Format("15:04:05")` | `posting_time` | 过账时间 |
| `warehouse.erp_code` | `set_warehouse` | 仓库编码 |
| `material.code` | `item_code` | 物品编码 |
| `item.counted_quantity` | `qty` | 实盘数量 |
| - | `purpose` | 盘点目的（默认：Stock Reconciliation） |
| - | `naming_series` | 编号系列（MAT-RECO-.YYYY.-） |

### 5.2 审核时的数据映射

| TTPOS 字段 | ERPNext 字段 | 说明 |
|------------|--------------|------|
| `stock_reconciliation.erp_code` | `name` | ERPNext 盘点单号 |
| - | `docstatus` | 文档状态（1 = Submitted） |

---

## 六、错误处理

### 6.1 常见错误

**仓库禁用错误**：
```
错误信息：Disabled Warehouse
处理方式：返回错误"仓库状态已关闭，请修改仓库状态"
```

**物品禁用错误**：
```
错误信息：Item XXX is disabled
处理方式：返回错误"物品XXX状态已关闭，请修改物品状态"
```

**其他错误**：
```
错误信息：ERPNext API 返回的错误信息
处理方式：返回通用错误"提交盘点单失败"或"审核盘点单失败"
```

### 6.2 错误处理代码

```go
// 文件：main/app/service/stock_reconciliation.go

if err != nil {
    // 检查是否是仓库禁用错误
    if strings.Contains(err.Error(), "Disabled Warehouse") {
        return errors.New("仓库状态已关闭，请修改仓库状态")
    }
    // 提取物品名称
    itemName := s.extractName("Item", "is disabled", err.Error())
    if itemName != "" {
        return errors.New("物品" + itemName + "状态已关闭，请修改物品状态")
    }
    return errors.WithMessage(errors.New("提交盘点单失败"), err.Error())
}
```

---

## 七、认证方式

### 7.1 BMP 模块认证

**站点认证**：
- 从 gRPC Context 中获取 `erp_site_code`
- 根据站点编码获取站点授权信息
- 设置 HTTP 请求头：`Authorization: Bearer {token}`

**实现位置**：
- 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go`
- 方法：`GetClient()`

### 7.2 ERPNext API 认证

**认证方式**：
- Token 认证（Bearer Token）
- Token 从站点配置中获取

---

## 八、同步时机

### 8.1 提交盘点单

**触发时机**：
- 用户在 TTPOS 中点击"提交盘点单"
- 盘点单状态：已保存 → 已提交

**同步内容**：
- 创建 ERPNext 盘点单（Draft 状态）
- 返回 ERPNext 盘点单号
- 更新 TTPOS 盘点单的 `erp_code` 字段

### 8.2 审核盘点单

**触发时机**：
- 用户在 TTPOS 中点击"审核通过"
- 盘点单状态：已提交 → 已审核

**同步内容**：
- 提交 ERPNext 盘点单（Submitted 状态）
- ERPNext 自动更新库存

---

## 九、接口调用流程

### 9.1 提交盘点单流程

```
1. TTPOS 用户操作
   ↓
2. TTPOS Main: submitStockReconciliation()
   - 构建 erpItems（盘点明细）
   - 调用 erpSrv.SubmitStockReconciliation()
   ↓
3. TTPOS ERP 服务: SubmitStockReconciliation()
   - 创建 gRPC 客户端
   - 调用 BMP 模块 gRPC 接口
   ↓
4. BMP 模块: SaveStockReconciliation()
   - 构建 ERPNext 数据
   - 调用 service.Document().Create()
   ↓
5. ERPNext API: POST /api/v2/document/Stock Reconciliation
   - 创建盘点单（Draft）
   - 返回盘点单号
   ↓
6. 更新 TTPOS 盘点单
   - erp_code = ERPNext 盘点单号
   - status = 已提交
```

### 9.2 审核盘点单流程

```
1. TTPOS 用户操作
   ↓
2. TTPOS Main: ApproveStockReconciliation()
   - 更新 TTPOS 库存
   - 调用 erpSrv.ApproveStockReconciliation()
   ↓
3. TTPOS ERP 服务: ApproveStockReconciliation()
   - 创建 gRPC 客户端
   - 调用 BMP 模块 gRPC 接口
   ↓
4. BMP 模块: SubmitStockReconciliation()
   - 调用 service.Document().ChangeDocStatus()
   ↓
5. ERPNext API: PUT /api/v2/document/Stock Reconciliation/{name}
   - 更新 docstatus = 1（Submitted）
   - ERPNext 自动更新库存
   ↓
6. 更新 TTPOS 盘点单状态
   - status = 已审核
```

---

## 十、关键代码位置

### 10.1 TTPOS Main 模块

**ERP 服务接口**：
- `main/app/service/rpc/erp/stock.go::SubmitStockReconciliation()`
- `main/app/service/rpc/erp/stock.go::ApproveStockReconciliation()`

**盘点单服务**：
- `main/app/service/stock_reconciliation.go::submitStockReconciliation()`
- `main/app/service/stock_reconciliation.go::ApproveStockReconciliation()`

### 10.2 BMP 模块

**gRPC 控制器**：
- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go::SaveStockReconciliation()`
- `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go::SubmitStockReconciliation()`

**业务逻辑**：
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go::SaveStockReconciliation()`
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go::SubmitStockReconciliation()`

**ERPNext API 调用**：
- `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go::Create()`
- `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go::ChangeDocStatus()`

### 10.3 Protobuf 定义

**接口定义**：
- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`

---

## 十一、总结

### 11.1 接口调用链

```
TTPOS Main (Go)
    ↓ gRPC
BMP 模块 (Go)
    ↓ HTTP REST API
ERPNext (Python)
```

### 11.2 关键接口

1. **提交盘点单**：
   - TTPOS → BMP：`/stock.StockService/SaveStockReconciliation`
   - BMP → ERPNext：`POST /api/v2/document/Stock Reconciliation`

2. **审核盘点单**：
   - TTPOS → BMP：`/stock.StockService/SubmitStockReconciliation`
   - BMP → ERPNext：`PUT /api/v2/document/Stock Reconciliation/{name}`

### 11.3 数据流向

- **提交时**：TTPOS 盘点单数据 → ERPNext 创建盘点单（Draft）
- **审核时**：ERPNext 盘点单号 → ERPNext 提交盘点单（Submitted）

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team























