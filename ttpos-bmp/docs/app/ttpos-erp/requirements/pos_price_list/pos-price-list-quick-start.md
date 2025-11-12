# POS 价格表快速入门

## 🚀 快速开始

### 1. 配置默认价格表

编辑 `manifest/config/config.yaml`：

```yaml
app:
  erpnext:
    core:
      pos_price_list:
        default:
          buying_price_list: "Buying - External"
          selling_price_list: "Selling - Internal"
```

### 2. 获取默认价格表

```go
import "ttpos-bmp/app/ttpos-erp/internal/service"

// 获取配置的默认价格表
priceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
if err != nil {
    return err
}

fmt.Printf("采购价格表: %s\n", priceList.BuyingPriceList)  // "Buying - External"
fmt.Printf("销售价格表: %s\n", priceList.SellingPriceList) // "Selling - Internal"
```

### 3. 智能获取价格表（推荐）

优先使用公司配置，如果没有则使用默认配置：

```go
// 封装的辅助函数
func GetPriceListForCompany(ctx context.Context, company string) (*erp.PosPriceList, error) {
    // 尝试获取公司配置
    priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, company)
    if err == nil {
        return priceList, nil
    }
    
    // 如果公司没有配置，使用默认配置
    return service.PosPriceList().GetDefaultPosPriceList(ctx)
}

// 使用
priceList, err := GetPriceListForCompany(ctx, "Company A")
if err != nil {
    return err
}
// 使用 priceList.BuyingPriceList 和 priceList.SellingPriceList
```

## 📝 常用场景

### 场景 A: 创建采购订单

```go
// 获取价格表
priceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
if err != nil {
    return err
}

// 创建采购订单
purchaseOrder := &erp.PurchaseOrder{
    Company:         "Company A",
    Supplier:        "Supplier X",
    BuyingPriceList: priceList.BuyingPriceList,  // 使用默认采购价格表
    // ... 其他字段
}
```

### 场景 B: 创建销售订单

```go
// 获取价格表
priceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
if err != nil {
    return err
}

// 创建销售订单
salesOrder := &erp.SalesOrder{
    Company:          "Company A",
    Customer:         "Customer Y",
    SellingPriceList: priceList.SellingPriceList,  // 使用默认销售价格表
    // ... 其他字段
}
```

### 场景 C: 为公司配置专属价格表

```go
// 第1步：创建价格表规则
priceList, err := service.PosPriceList().CreatePosPriceList(ctx, &erp.PosPriceList{
    RuleCode:         "VIP-COMPANY",
    Company:          "VIP Company",
    BuyingPriceList:  "Premium Buying",
    SellingPriceList: "Premium Selling",
    Disabled:         0,
})

// 第2步：使用时会自动获取该公司的配置
priceList, err = service.PosPriceList().GetPosPriceListByCompany(ctx, "VIP Company")
// priceList.BuyingPriceList  = "Premium Buying"
// priceList.SellingPriceList = "Premium Selling"
```

## ⚡ 性能提示

### 缓存默认配置

由于默认配置很少改变，可以考虑缓存：

```go
var (
    defaultPriceList  *erp.PosPriceList
    defaultConfigOnce sync.Once
)

func GetDefaultPriceLists(ctx context.Context) (*erp.PosPriceList, error) {
    var err error
    defaultConfigOnce.Do(func() {
        defaultPriceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
    })
    return defaultPriceList, err
}
```

## ⚠️ 注意事项

1. **配置检查**：启动时验证配置是否存在
```go
func init() {
    ctx := context.Background()
    _, _, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
    if err != nil {
        panic("默认价格表配置错误: " + err.Error())
    }
}
```

2. **价格表存在性**：确保 ERPNext 中存在配置的价格表
3. **公司优先**：生产环境建议优先使用公司配置，默认配置作为兜底
4. **日志记录**：使用默认配置时记录日志，便于监控

## 🔍 故障排查

### 错误：配置文件中未设置默认采购价格表

**原因**：配置文件中缺少配置项

**解决**：在 `config.yaml` 中添加：
```yaml
app:
  erpnext:
    core:
      pos_price_list:
        default:
          buying_price_list: "Standard Buying"
          selling_price_list: "Standard Selling"
```

### 错误：公司未配置价格表规则

**原因**：该公司没有创建价格表规则

**解决方案 1**：使用默认配置（推荐）
```go
priceList, err := service.PosPriceList().GetDefaultPosPriceList(ctx)
```

**解决方案 2**：为公司创建价格表规则
```go
service.PosPriceList().CreatePosPriceList(ctx, &erp.PosPriceList{
    RuleCode:         "COMPANY-001",
    Company:          "Company A",
    BuyingPriceList:  "Standard Buying",
    SellingPriceList: "Standard Selling",
})
```

## 🎯 最佳实践

```go
// 推荐的价格表获取函数
func GetPriceListsForOrder(ctx context.Context, company string) (*erp.PosPriceList, error) {
    // 1. 优先获取公司配置
    priceList, err := service.PosPriceList().GetPosPriceListByCompany(ctx, company)
    if err == nil && priceList.Disabled == 0 {
        g.Log().Debug(ctx, "使用公司配置的价格表", g.Map{
            "company": company,
            "buying":  priceList.BuyingPriceList,
            "selling": priceList.SellingPriceList,
        })
        return priceList, nil
    }
    
    // 2. 公司没有配置或已禁用，使用默认配置
    g.Log().Warning(ctx, "公司未配置价格表，使用默认配置", g.Map{
        "company": company,
        "error":   err,
    })
    
    priceList, err = service.PosPriceList().GetDefaultPosPriceList(ctx)
    if err != nil {
        return nil, gerror.Wrap(err, "获取默认价格表失败")
    }
    
    g.Log().Info(ctx, "使用默认价格表", g.Map{
        "company": company,
        "buying":  priceList.BuyingPriceList,
        "selling": priceList.SellingPriceList,
    })
    
    return priceList, nil
}
```

