# POS 价格表规则服务使用说明

## 📋 功能概述

POS 价格表规则服务用于管理采购价格表和销售价格表的映射关系，为不同公司配置默认的价格表规则。

## 🔧 已实现功能

### 1. 创建价格表规则

```go
priceList, err := service.PosPriceList().CreatePosPriceList(ctx, &erp.PosPriceList{
    RuleCode:         "RULE-001",
    Company:          "Company A",
    BuyingPriceList:  "Standard Buying",
    SellingPriceList: "Standard Selling",
    Disabled:         0,
})
```

### 2. 获取价格表规则详情

```go
priceList, err := service.PosPriceList().GetPosPriceList(ctx, "RULE-001")
```

### 3. 更新价格表规则

```go
priceList, err := service.PosPriceList().UpdatePosPriceList(ctx, &erp.PosPriceList{
    Name:             "RULE-001",
    BuyingPriceList:  "Premium Buying",
    SellingPriceList: "Premium Selling",
})
```

### 4. 删除价格表规则

```go
err := service.PosPriceList().DeletePosPriceList(ctx, "RULE-001")
```

### 5. 获取价格表规则列表

```go
list, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    CompanyAbbr: "CA",     // 按公司缩写查询
    Disabled:    0,        // 只查询启用的规则
    PageNo:      1,        // 页码
    PageSize:    10,       // 每页数量
})
```

### 6. 从配置文件获取默认价格表

```go
priceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
if err != nil {
    // 处理错误
}
// priceList.BuyingPriceList: "Buying - External"
// priceList.SellingPriceList: "Selling - Internal"
```

### 7. 根据公司获取默认价格表规则

```go
priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, "Company A")
```

## 📦 数据结构

### PosPriceList

```go
type PosPriceList struct {
    Name             string // 文档名称（ERPNext自动生成）
    RuleCode         string // 规则代码（必填）
    Company          string // 公司（必填）
    BuyingPriceList  string // 采购价格表（必填）
    SellingPriceList string // 销售价格表（必填）
    Disabled         int    // 是否禁用（0-启用，1-禁用）
    Owner            string // 创建人
    Creation         string // 创建时间
    Modified         string // 修改时间
    ModifiedBy       string // 修改人
    Docstatus        int    // 文档状态（0-草稿，1-已提交，2-已取消）
    Idx              int    // 索引
}
```

### ListPosPriceListsReq

```go
type ListPosPriceListsReq struct {
    CompanyAbbr string // 公司缩写编码（自动转换为公司名称）
    Company     string // 公司名称（直接使用）
    RuleCode    string // 规则代码
    Disabled    int    // 是否禁用（-1-全部，0-启用，1-禁用）
    PageNo      int    // 页码（必填，最小为1）
    PageSize    int    // 每页数量（必填，1-100）
}
```

## 🔍 字段说明

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Name | string | - | ERPNext 自动生成的文档名称 |
| RuleCode | string | ✅ | 规则代码，用于唯一标识规则 |
| Company | string | ✅ | 公司名称 |
| BuyingPriceList | string | ✅ | 采购价格表名称 |
| SellingPriceList | string | ✅ | 销售价格表名称 |
| Disabled | int | - | 0-启用，1-禁用 |

## ⚙️ 配置文件

在 `manifest/config/config.yaml` 中配置默认价格表：

```yaml
app:
  erpnext:
    core:
      pos_price_list:
        default:
          buying_price_list: "Buying - External"
          selling_price_list: "Selling - Internal"
```

配置说明：
- `buying_price_list`: 默认采购价格表名称（必须是 ERPNext 中已存在的价格表）
- `selling_price_list`: 默认销售价格表名称（必须是 ERPNext 中已存在的价格表）

## 🎯 使用场景

### 场景1：使用配置文件的默认价格表

```go
// 获取配置的默认价格表
defaultPriceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
if err != nil {
    return err
}

// 创建订单时使用默认价格表
order := &erp.PurchaseOrder{
    BuyingPriceList: defaultPriceList.BuyingPriceList,
    // ... 其他字段
}
```

### 场景2：为新公司配置默认价格表

```go
// 创建价格表规则
priceList, err := service.PosPriceList().CreatePosPriceList(ctx, &erp.PosPriceList{
    RuleCode:         "DEFAULT-COMPANY-A",
    Company:          "Company A",
    BuyingPriceList:  "Standard Buying",
    SellingPriceList: "Standard Selling",
    Disabled:         0,
})
if err != nil {
    // 错误处理
}
```

### 场景3：获取公司的价格表配置（优先公司配置，否则使用默认）

```go
var priceList *erp.PosPriceList

// 先尝试获取公司的价格表规则
priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, "Company A")
if err != nil {
    // 公司未配置价格表规则，使用默认配置
    g.Log().Warning(ctx, "公司未配置价格表规则，使用默认配置", err)
    
    priceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
    if err != nil {
        return gerror.Wrap(err, "获取默认价格表失败")
    }
}

// 使用价格表创建订单
order := &erp.PurchaseOrder{
    Company:          "Company A",
    BuyingPriceList:  priceList.BuyingPriceList,
    SellingPriceList: priceList.SellingPriceList,
    // ... 其他字段
}
```

### 场景4：按公司缩写查询价格表规则列表

```go
list, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    CompanyAbbr: "CA",     // 公司缩写，自动转换为公司名称
    Disabled:    0,        // 只查询启用的规则
    PageNo:      1,
    PageSize:    20,
})
if err != nil {
    return err
}

for _, priceList := range list {
    fmt.Printf("规则: %s, 公司: %s\n", priceList.RuleCode, priceList.Company)
}
fmt.Printf("共查询到 %d 条记录\n", len(list))
```

### 场景5：查询所有启用的价格表规则

```go
list, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    Disabled: 0,    // 只查询启用的规则
    PageNo:   1,
    PageSize: 100,
})
```

## ⚠️ 注意事项

1. **规则代码唯一性**：`RuleCode` 应该是唯一的，建议使用有意义的命名规则
2. **公司名称**：必须是 ERPNext 中已存在的公司名称
3. **价格表名称**：`BuyingPriceList` 和 `SellingPriceList` 必须是 ERPNext 中已存在的价格表
4. **禁用状态**：禁用的规则不会被 `GetPosPriceListByCompany` 方法返回
5. **删除规则**：删除前需要确保没有关联的业务数据引用该规则
6. **公司缩写**：使用 `CompanyAbbr` 时会自动查询对应的公司名称，如果查询失败会记录日志但不影响其他条件

## 🔄 与其他服务的集成

### 与采购订单集成

```go
// 获取公司的价格表规则（如果没有则使用默认）
priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, "Company A")
if err != nil {
    // 使用默认配置
    priceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
    if err != nil {
        return err
    }
}

// 创建采购订单时使用价格表
purchaseOrder := &erp.PurchaseOrder{
    Company:         "Company A",
    BuyingPriceList: priceList.BuyingPriceList,
    // ... 其他字段
}
```

### 与销售订单集成

```go
// 获取公司的价格表规则（如果没有则使用默认）
priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, "Company A")
if err != nil {
    // 使用默认配置
    priceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
    if err != nil {
        return err
    }
}

// 创建销售订单时使用价格表
salesOrder := &erp.SalesOrder{
    Company:          "Company A",
    SellingPriceList: priceList.SellingPriceList,
    // ... 其他字段
}
```

## 📝 错误处理

```go
priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, "Company A")
if err != nil {
    if strings.Contains(err.Error(), "未配置价格表规则") {
        // 公司未配置价格表规则，使用默认配置或提示用户配置
        return gerror.New("请先为该公司配置价格表规则")
    }
    // 其他错误
    return err
}
```

## 🎉 完成状态

- ✅ 创建价格表规则
- ✅ 获取价格表规则详情
- ✅ 更新价格表规则
- ✅ 删除价格表规则
- ✅ 获取价格表规则列表（支持 company_abbr 查询）
- ✅ 从配置文件获取默认价格表
- ✅ 根据公司获取默认价格表规则
- ✅ 参数验证
- ✅ 错误处理
- ✅ 日志记录
- ✅ 配置文件支持

所有核心功能已实现，可以直接使用！

## 📖 API 总览

| 方法 | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `CreatePosPriceList` | 创建价格表规则 | `*erp.PosPriceList` | `*erp.PosPriceList, error` |
| `GetPosPriceList` | 获取价格表规则详情 | `name string` | `*erp.PosPriceList, error` |
| `UpdatePosPriceList` | 更新价格表规则 | `*erp.PosPriceList` | `*erp.PosPriceList, error` |
| `DeletePosPriceList` | 删除价格表规则 | `name string` | `error` |
| `ListPosPriceLists` | 获取价格表规则列表 | `*erp.ListPosPriceListsReq` | `[]*erp.PosPriceList, error` |
| `GetDefaultPosPriceList` | 从配置文件获取默认价格表 | - | `*erp.PosPriceList, error` |
| `GetPosPriceListByCompany` | 根据公司获取价格表规则 | `company string` | `*erp.PosPriceList, error` |

## 🔍 ListPosPriceListsReq 查询条件说明

| 字段 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| CompanyAbbr | string | - | 公司缩写编码（自动转换为公司名称） | "CA" |
| Company | string | - | 公司名称（直接使用） | "Company A" |
| RuleCode | string | - | 规则代码 | "RULE-001" |
| Disabled | int | - | 是否禁用（-1-全部，0-启用，1-禁用） | 0 |
| PageNo | int | ✅ | 页码（最小为1） | 1 |
| PageSize | int | ✅ | 每页数量（1-100） | 20 |

**查询逻辑**：
- 如果同时提供 `CompanyAbbr` 和 `Company`，优先使用 `CompanyAbbr`
- `Disabled` 设置为 -1 或不设置时查询全部状态
- 分页参数必填，超出范围会自动调整

## 💡 使用示例

### 示例1：按公司缩写查询

```go
// 查询公司缩写为 "HQ" 的所有启用规则
list, total, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    CompanyAbbr: "HQ",
    Disabled:    0,
    PageNo:      1,
    PageSize:    20,
})
```

### 示例2：按规则代码查询

```go
// 查询特定规则代码
list, total, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    RuleCode: "DEFAULT-RULE",
    PageNo:   1,
    PageSize: 10,
})
```

### 示例3：查询所有规则

```go
// 查询所有价格表规则（不限制状态）
list, total, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    Disabled: -1,  // -1 表示查询全部
    PageNo:   1,
    PageSize: 100,
})
```

### 示例4：组合查询

```go
// 查询特定公司的启用规则
list, total, err := service.PosPriceList().ListPosPriceLists(ctx, &erp.ListPosPriceListsReq{
    CompanyAbbr: "CA",
    RuleCode:    "VIP",
    Disabled:    0,
    PageNo:      1,
    PageSize:    10,
})
```
