# ERPNext Document Template（文档模板）使用指南

> 📖 **用途**: 说明如何在 ERPNext 中创建和使用文档模板，以及如何通过 API 使用模板创建盘点单等文档

---

## 一、Document Template 概述

### 1.1 什么是 Document Template

**Document Template（文档模板）**是 ERPNext 提供的一个功能，允许你预设文档的默认字段值，以便快速创建具有相同配置的文档。

### 1.2 适用场景

- **盘点单模板**：预设默认仓库、盘点目的、成本中心等
- **采购订单模板**：预设供应商、默认仓库、付款条件等
- **销售订单模板**：预设客户、默认仓库、价格表等
- **其他单据模板**：任何需要重复使用相同配置的文档

### 1.3 核心优势

- ✅ **提高效率**：减少重复输入
- ✅ **保证一致性**：确保相同类型的文档使用相同的配置
- ✅ **降低错误**：减少手动输入错误

---

## 二、在 ERPNext 中创建 Document Template

### 2.1 创建盘点单模板

#### Step 1: 访问模板列表

1. 登录 ERPNext 系统
2. 导航至：**主页 > 库存 > 工具 > 文档模板**
3. 或者直接搜索：`Document Template`

#### Step 2: 创建新模板

1. 点击 **"新建"** 按钮
2. 填写模板信息：

| 字段 | 说明 | 示例 |
|------|------|------|
| **模板标题** | 模板名称 | `日常盘点模板` |
| **文档类型** | 选择 `Stock Reconciliation` | `Stock Reconciliation` |
| **公司** | 选择适用的公司 | `Company A` |
| **分支** | 选择分支（可选） | `Branch 1` |
| **默认仓库** | 预设仓库 | `WH-001` |
| **盘点目的** | 预设目的 | `Stock Reconciliation` |
| **成本中心** | 预设成本中心（可选） | `Cost Center 1` |
| **费用科目** | 预设费用科目（可选） | `Expense Account` |

#### Step 3: 保存模板

点击 **"保存"** 按钮，模板创建完成。

---

### 2.2 模板字段说明

**盘点单模板支持的字段**：

| 字段 | ERPNext 字段名 | 是否必填 | 说明 |
|------|---------------|---------|------|
| 公司 | `company` | ✅ | 公司名称 |
| 分支 | `branch` | ❌ | 分支名称 |
| 默认仓库 | `set_warehouse` | ❌ | 默认仓库编码 |
| 盘点目的 | `purpose` | ❌ | `Stock Reconciliation` 或 `Opening Stock` |
| 成本中心 | `cost_center` | ❌ | 成本中心编码 |
| 费用科目 | `expense_account` | ❌ | 费用科目编码 |
| 过账日期 | `posting_date` | ❌ | 默认过账日期（通常使用当前日期） |
| 过账时间 | `posting_time` | ❌ | 默认过账时间 |

**注意**：
- 模板中设置的字段值会在创建文档时自动填充
- 创建文档后仍可以修改这些字段值
- 模板不支持设置明细行（items），明细行需要在创建文档后手动添加

---

## 三、在 ERPNext Web 界面中使用模板

### 3.1 使用模板创建盘点单

#### Step 1: 创建新盘点单

1. 导航至：**主页 > 库存 > 工具 > 库存盘点**
2. 点击 **"新建"** 按钮

#### Step 2: 选择模板

1. 在盘点单表单中找到 **"来自模板"** 字段（`from_template`）
2. 从下拉列表中选择之前创建的模板
3. 系统会自动填充模板中预设的字段值

#### Step 3: 添加明细并保存

1. 添加盘点明细（物品和数量）
2. 根据需要修改字段值
3. 点击 **"保存"** 或 **"提交"**

---

## 四、通过 API 使用模板创建文档

### 4.1 ERPNext API 方式

#### 方式一：使用 `from_template` 参数

**API 端点**：
```
POST /api/v2/document/Stock Reconciliation
```

**请求示例**：
```json
{
  "from_template": "日常盘点模板",
  "items": [
    {
      "item_code": "MAT-001",
      "qty": 100.0,
      "warehouse": "WH-001",
      "valuation_rate": 1.0,
      "doctype": "Stock Reconciliation Item"
    }
  ],
  "doctype": "Stock Reconciliation"
}
```

**说明**：
- `from_template` 字段值为模板标题（Template Title）
- ERPNext 会根据模板标题查找对应的模板
- 模板中的字段值会自动填充到新文档中
- 可以在请求中覆盖模板的字段值

#### 方式二：先获取模板，再创建文档

**Step 1: 获取模板数据**

```
GET /api/v2/document/Document Template?filters=[["template_title", "=", "日常盘点模板"]]
```

**响应示例**：
```json
{
  "data": [
    {
      "name": "TMP-00001",
      "template_title": "日常盘点模板",
      "doc_type": "Stock Reconciliation",
      "company": "Company A",
      "set_warehouse": "WH-001",
      "purpose": "Stock Reconciliation",
      "cost_center": "Cost Center 1"
    }
  ]
}
```

**Step 2: 使用模板数据创建文档**

```json
{
  "company": "Company A",
  "set_warehouse": "WH-001",
  "purpose": "Stock Reconciliation",
  "cost_center": "Cost Center 1",
  "items": [
    {
      "item_code": "MAT-001",
      "qty": 100.0,
      "warehouse": "WH-001",
      "valuation_rate": 1.0,
      "doctype": "Stock Reconciliation Item"
    }
  ],
  "doctype": "Stock Reconciliation"
}
```

---

### 4.2 TTPOS 代码集成方式

#### 当前实现（未使用模板）

当前 TTPOS 代码中创建盘点单时，字段值是硬编码的：

```go
// 文件：ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go
data := &erp.StockReconciliation{
    NamingSeries: erp.DefaultStockReconciliationSeries,
    Company:      company.CompanyName,
    PostingDate:  req.PostingDate,
    DocType:      erp.DocTypeStockReconciliation,
    Purpose:      erp.StockReconciliationPurposeStockReconciliation,
    SetWarehouse: warehouseName,
    // ... 其他字段
}
```

#### 改进方案：支持使用模板

**方案一：在请求中传递模板名称**

```go
// 修改请求结构
type SaveStockReconciliationReq struct {
    CompanyAbbr string
    Branch      string
    PostingDate string
    PostingTime string
    Warehouse   string
    TemplateTitle string  // 新增：模板标题
    Items       []*StockReconciliationItem
}

// 修改创建逻辑
func (s *sStock) SaveStockReconciliation(ctx context.Context, req *stock.SaveStockReconciliationReq) (res *stock.SaveStockReconciliationResp, err error) {
    // ... 获取公司信息等 ...
    
    // 构建数据
    data := &erp.StockReconciliation{
        DocType: erp.DocTypeStockReconciliation,
    }
    
    // 如果指定了模板，使用模板
    if len(req.TemplateTitle) > 0 {
        // 方式1：直接传递 from_template 参数
        data.FromTemplate = req.TemplateTitle
        
        // 方式2：或者先获取模板数据，再合并字段
        // templateData, err := s.getTemplateData(ctx, req.TemplateTitle)
        // if err == nil {
        //     // 合并模板字段到 data
        // }
    } else {
        // 使用默认值
        data.NamingSeries = erp.DefaultStockReconciliationSeries
        data.Purpose = erp.StockReconciliationPurposeStockReconciliation
        // ... 其他默认字段
    }
    
    // 设置必填字段（覆盖模板值）
    data.Company = company.CompanyName
    data.PostingDate = req.PostingDate
    
    // ... 添加明细等 ...
    
    // 创建文档
    resp, err := service.Document().Create(ctx, erp.DocTypeStockReconciliation, data)
    // ...
}
```

**方案二：在公司配置中设置默认模板**

```go
// 在公司配置表中添加字段
// company_setting.stock_reconciliation_template_title

// 创建盘点单时自动使用公司默认模板
func (s *sStock) SaveStockReconciliation(ctx context.Context, req *stock.SaveStockReconciliationReq) (res *stock.SaveStockReconciliationResp, err error) {
    // 获取公司配置
    companySetting, err := service.Company().GetCompanySetting(ctx, req.CompanyAbbr)
    if err != nil {
        return nil, err
    }
    
    // 如果公司配置了默认模板，使用模板
    templateTitle := ""
    if len(companySetting.StockReconciliationTemplateTitle) > 0 {
        templateTitle = companySetting.StockReconciliationTemplateTitle
    } else if len(req.TemplateTitle) > 0 {
        templateTitle = req.TemplateTitle
    }
    
    // ... 使用模板创建文档 ...
}
```

---

## 五、模板使用最佳实践

### 5.1 模板命名规范

建议使用清晰的命名规则：

- **按用途命名**：`日常盘点模板`、`期初盘点模板`
- **按仓库命名**：`主仓库盘点模板`、`分仓盘点模板`
- **按公司命名**：`Company A 盘点模板`、`Company B 盘点模板`

### 5.2 模板字段设置建议

**必填字段**：
- ✅ `company`：公司（必须设置）
- ✅ `doc_type`：文档类型（自动设置）

**推荐设置字段**：
- ✅ `set_warehouse`：默认仓库（如果所有盘点都使用同一仓库）
- ✅ `purpose`：盘点目的（如果固定）
- ✅ `cost_center`：成本中心（如果需要成本核算）

**可选字段**：
- ❌ `posting_date`：通常使用当前日期，不需要在模板中设置
- ❌ `posting_time`：通常使用当前时间，不需要在模板中设置

### 5.3 模板权限管理

- **创建权限**：只有管理员或授权用户可以创建模板
- **使用权限**：所有用户可以查看和使用模板
- **修改权限**：只有创建者或管理员可以修改模板

---

## 六、常见问题

### Q1: 模板中的字段值可以修改吗？

**A**: 可以。使用模板创建文档后，所有字段值都可以修改。模板只是提供初始值。

### Q2: 模板支持设置明细行吗？

**A**: 不支持。ERPNext 的 Document Template 不支持预设明细行（items），明细行需要在创建文档后手动添加。

### Q3: 如何删除模板？

**A**: 
1. 导航至：**主页 > 库存 > 工具 > 文档模板**
2. 找到要删除的模板
3. 点击 **"删除"** 按钮
4. 确认删除

**注意**：删除模板不会影响已创建的文档。

### Q4: 模板可以设置多个仓库吗？

**A**: 不可以。模板中的 `set_warehouse` 字段只能设置一个默认仓库。如果需要为不同仓库创建盘点单，可以：
- 创建多个模板（每个仓库一个模板）
- 或者创建文档后手动修改仓库

### Q5: 通过 API 使用模板时，如何覆盖模板字段值？

**A**: 在 API 请求中直接设置字段值即可覆盖模板值：

```json
{
  "from_template": "日常盘点模板",
  "set_warehouse": "WH-002",  // 覆盖模板中的仓库
  "purpose": "Opening Stock",  // 覆盖模板中的目的
  "items": [...]
}
```

---

## 七、总结

### 7.1 Document Template 的优势

- ✅ **提高效率**：减少重复输入，快速创建文档
- ✅ **保证一致性**：确保相同类型的文档使用相同的配置
- ✅ **降低错误**：减少手动输入错误

### 7.2 适用场景

- ✅ 需要频繁创建相同配置的文档
- ✅ 需要确保文档字段值的一致性
- ✅ 需要简化文档创建流程

### 7.3 注意事项

- ⚠️ 模板不支持设置明细行
- ⚠️ 模板字段值可以在创建文档后修改
- ⚠️ 删除模板不会影响已创建的文档

---

## 八、相关文档

- [ERPNext 官方文档 - Document Template](https://docs.erpnext.com/docs/user/manual/en/customize-erpnext/document-template)
- [盘点单同步到 ERPNext 接口说明](./stock-reconciliation-erpnext-api.md)
- [盘点单 TTPOS 与 ERPNext 数据同步机制](../../human/business/stock-reconciliation-erp-sync.md)

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team

