# Core 核心服务说明文档

## 📋 服务概览

Core 核心服务是 ttpos-erp 模块的基础核心服务，提供用户管理和价格表管理等核心功能。该服务为其他服务提供基础能力支持。

## 🎯 主要功能

### 用户管理
- **获取用户信息**: 根据用户名获取用户详细信息
- **获取用户时区**: 获取用户配置的时区（默认 Asia/Bangkok）

### 价格表管理
- **创建价格表规则**: 创建 POS 价格表规则
- **获取价格表规则**: 获取价格表规则详情
- **更新价格表规则**: 更新价格表规则
- **删除价格表规则**: 删除价格表规则
- **价格表规则列表**: 查询价格表规则列表
- **默认价格表**: 从配置文件获取默认价格表
- **公司价格表**: 根据公司获取价格表规则

## 📁 文件结构

```
internal/logic/core/
├── user.go                   # 用户管理
└── pos_price_list.go         # 价格表管理
```

## 🔧 接口定义

### IUser 接口（用户管理）

```go
type IUser interface {
    // GetUserByUsername 根据用户名获取用户信息
    GetUserByUsername(ctx context.Context, userEmail string) (*erp.User, error)
    
    // MustGetUserTimeZone 获取用户时区
    MustGetUserTimeZone(ctx context.Context, userEmail string) string
}
```

### IPosPriceList 接口（价格表管理）

```go
type IPosPriceList interface {
    // CreatePosPriceList 创建 POS 价格表规则
    CreatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error)
    
    // GetPosPriceList 获取 POS 价格表规则详情
    GetPosPriceList(ctx context.Context, name string) (*erp.PosPriceList, error)
    
    // UpdatePosPriceList 更新 POS 价格表规则
    UpdatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error)
    
    // DeletePosPriceList 删除 POS 价格表规则
    DeletePosPriceList(ctx context.Context, name string) error
    
    // ListPosPriceLists 获取 POS 价格表规则列表
    ListPosPriceLists(ctx context.Context, req *erp.ListPosPriceListsReq) ([]*erp.PosPriceList, error)
    
    // GetDefaultPosPriceList 从配置文件获取默认价格表
    GetDefaultPosPriceList(ctx context.Context) (*erp.PosPriceList, error)
    
    // GetPosPriceListByCompany 根据公司获取默认价格表规则
    GetPosPriceListByCompany(ctx context.Context, company string) (*erp.PosPriceList, error)
}
```

## 🏗️ 实现细节

### 用户信息获取

```go
func (s *sUser) GetUserByUsername(ctx context.Context, userEmail string) (*erp.User, error) {
    // 调用 Document 服务获取用户信息
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: "User",
        Name:    userEmail,
    }, nil)
    
    // 解析响应数据
    userData := resp.GetJson("data")
    var user = &erp.User{}
    userData.Scan(&user)
    
    return user, nil
}
```

### 用户时区获取

```go
func (s *sUser) MustGetUserTimeZone(ctx context.Context, userEmail string) string {
    user, err := s.GetUserByUsername(ctx, userEmail)
    if err != nil {
        // 默认返回曼谷时区
        return "Asia/Bangkok"
    }
    return user.TimeZone
}
```

### 价格表规则创建

```go
func (s *sPosPriceList) CreatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error) {
    // 参数验证
    if err := s.validateCreateReq(req); err != nil {
        return nil, err
    }
    
    // 构建创建数据
    data := s.buildPosPriceListData(ctx, req)
    
    // 调用ERP接口创建价格表规则
    resp, err := service.Document().Create(ctx, erp.DocTypePosPriceList, data)
    
    // 解析响应数据
    result, err := s.parsePosPriceListResponse(resp.MustToJson())
    
    return result, nil
}
```

### 默认价格表获取

从配置文件读取默认价格表：

```go
func (s *sPosPriceList) GetDefaultPosPriceList(ctx context.Context) (*erp.PosPriceList, error) {
    // 从配置文件读取默认价格表
    buyingPriceList := g.Cfg().MustGet(ctx, 
        "app.erpnext.core.pos_price_list.default.buying_price_list", 
        "Buying - External").String()
    sellingPriceList := g.Cfg().MustGet(ctx, 
        "app.erpnext.core.pos_price_list.default.selling_price_list", 
        "Selling - Internal").String()
    
    return &erp.PosPriceList{
        RuleCode:         "DEFAULT",
        Company:          "Default",
        BuyingPriceList:  buyingPriceList,
        SellingPriceList: sellingPriceList,
        Disabled:         0,
    }, nil
}
```

### 公司价格表获取

根据公司名称获取价格表规则：

```go
func (s *sPosPriceList) GetPosPriceListByCompany(ctx context.Context, company string) (*erp.PosPriceList, error) {
    // 构建过滤条件
    filters := [][]string{
        {"company", "=", company},
        {"disabled", "=", "0"},
    }
    
    // 查询价格表规则列表
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypePosPriceList,
    }, &erp.RequestParams{
        Fields:  g.ArrayStr{"name", "rule_code", "company", "buying_price_list", "selling_price_list", "disabled"},
        Filters: filters,
        Limit:   1,
    })
    
    // 解析响应数据
    list, err := s.parsePosPriceListListResponse(resp.MustToJson())
    if len(list) == 0 {
        return nil, gerror.Newf("公司 %s 未配置价格表规则", company)
    }
    
    return list[0], nil
}
```

## 📊 数据模型

### User 用户信息

```go
type User struct {
    Email    string // 邮箱
    FirstName string // 名字
    LastName  string // 姓氏
    TimeZone  string // 时区
    Enabled   int    // 是否启用
    // ... 更多字段
}
```

### PosPriceList 价格表规则

```go
type PosPriceList struct {
    Name             string // 价格表规则名称
    RuleCode         string // 规则代码
    Company          string // 公司
    BuyingPriceList  string // 采购价格表
    SellingPriceList string // 销售价格表
    Disabled         int    // 是否禁用
}
```

## 🔄 使用流程

### 1. 获取用户信息

```go
user, err := userService.GetUserByUsername(ctx, "cashier@example.com")
fmt.Printf("用户时区: %s\n", user.TimeZone)
```

### 2. 获取用户时区

```go
timeZone := userService.MustGetUserTimeZone(ctx, "cashier@example.com")
// 返回: "Asia/Bangkok" 或用户配置的时区
```

### 3. 创建价格表规则

```go
priceList, err := posPriceListService.CreatePosPriceList(ctx, &erp.PosPriceList{
    RuleCode:         "CFG-PRICE",
    Company:          "CFG Company",
    BuyingPriceList:  "Buying - External",
    SellingPriceList: "Selling - Internal",
})
```

### 4. 获取公司价格表

```go
priceList, err := posPriceListService.GetPosPriceListByCompany(ctx, "CFG Company")
fmt.Printf("采购价格表: %s\n", priceList.BuyingPriceList)
fmt.Printf("销售价格表: %s\n", priceList.SellingPriceList)
```

### 5. 获取默认价格表

```go
defaultPriceList, err := posPriceListService.GetDefaultPosPriceList(ctx)
```

## ⚠️ 注意事项

1. **用户时区**: 如果获取用户信息失败，默认返回 "Asia/Bangkok"
2. **价格表规则**: 每个公司可以配置一个价格表规则
3. **默认价格表**: 从配置文件读取，如果未配置则使用默认值
4. **禁用状态**: 禁用的价格表规则不会被查询到

## 📝 总结

Core 核心服务提供了用户管理和价格表管理等核心功能。

### 技术特点

- **用户管理**: 提供用户信息查询和时区获取
- **价格表管理**: 完整的价格表规则 CRUD 操作
- **默认配置**: 支持从配置文件读取默认价格表

### 设计优势

- **基础服务**: 为其他服务提供基础能力支持
- **简单高效**: 接口简洁，易于使用

