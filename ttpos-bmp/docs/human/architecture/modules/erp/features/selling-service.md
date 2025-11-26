# Selling 销售服务说明文档

## 📋 服务概览

Selling 销售服务是 ttpos-erp 模块的核心业务服务，负责 POS 销售全流程管理。该服务与 ERPNext 的 POS 模块深度对接，提供开账、关账、发票管理、支付方式管理等完整的 POS 销售能力。

## 🎯 主要功能

### POS 配置管理
- **配置列表**: 查询 POS Profile 列表
- **创建配置**: 创建新的 POS Profile
- **用户绑定**: 配置可用用户

### POS 开关账
- **开账**: 创建 POS Opening Entry
- **关账**: 创建 POS Closing Entry
- **状态查询**: 查询开账状态

### POS 发票管理
- **保存发票**: 创建销售发票
- **退货发票**: 创建退货发票
- **取消发票**: 取消已提交发票
- **发票查询**: 查询期间发票

### 支付方式
- **支付列表**: 获取支付方式列表
- **创建支付**: 创建支付方式账户

### 客户管理
- **客户列表**: 查询客户信息列表（支持分页和过滤）
- **客户详情**: 获取客户完整信息
- **创建客户**: 创建新客户（支持内部客户）
- **更新客户**: 更新客户信息
- **客户统计**: 统计符合条件的客户数量
- **添加交易公司**: 为客户添加允许交易的公司

### 销售订单管理
- **创建销售订单**: 创建新的销售订单
- **更新销售订单**: 更新销售订单信息
- **获取销售订单**: 查询销售订单详情
- **提交销售订单**: 提交销售订单
- **取消销售订单**: 取消销售订单

## 📁 文件结构

```
internal/logic/selling/
├── selling.go           # 销售服务主逻辑
├── sale_order.go        # 销售订单
├── selling_customer.go  # 客户管理
└── async_selling.go     # 异步销售处理

api/selling/
├── selling.proto        # 销售服务 Protobuf 定义
├── selling.pb.go        # 生成的 Go 代码
└── selling_grpc.pb.go   # 生成的 gRPC 代码
```

## 🔧 接口定义

### ISelling 接口

```go
type ISelling interface {
    // GetPosProfileList 获取 POS 配置文件列表
    GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (*selling.PosProfileListResp, error)
    
    // CreatePosProfile 创建 POS 配置文件
    CreatePosProfile(ctx context.Context, req *setup.CreatePosProfileInp) (*erp.POSProfile, error)
    
    // OpenPosEntry POS 开账
    OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error)
    
    // ClosePosEntry POS 关账
    ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
    
    // SavePosInvoice 保存 POS 发票
    SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error)
    
    // ReturnPosInvoice 退货 POS 发票
    ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
    
    // CancelPosInvoice 取消 POS 发票
    CancelPosInvoice(ctx context.Context, invoiceName string) error
    
    // GetModeOfPaymentList 获取支付方式列表
    GetModeOfPaymentList(ctx context.Context, req *selling.GetModeOfPaymentListReq) (*selling.GetModeOfPaymentListResp, error)
    
    // CreateModePaymentAccount 创建支付方式账户
    CreateModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) error
    
    // IsProfileOpening 查询配置是否已开账
    IsProfileOpening(ctx context.Context, posProfile, user string) (bool, error)
    
    // GetPosOpeningEntry 获取开账记录
    GetPosOpeningEntry(ctx context.Context, name string) (*erp.POSOpeningEntry, error)
    
    // GetPosInvoiceList 获取 POS 发票列表
    GetPosInvoiceList(ctx context.Context, req *dtoSelling.GetPosInvoiceListReq) ([]dtoSelling.SimplePosInvoice, error)
    
    // 客户管理
    // CountCustomer 统计客户数量
    CountCustomer(ctx context.Context, filter *erp.Customer) (int, error)
    
    // CreateCustomer 创建客户
    CreateCustomer(ctx context.Context, req *erp.Customer) (*erp.Customer, error)
    
    // UpdateCustomer 更新客户
    UpdateCustomer(ctx context.Context, name string, req *erp.Customer) (*erp.Customer, error)
    
    // GetCustomer 获取客户信息
    GetCustomer(ctx context.Context, name string) (*erp.Customer, error)
    
    // ListCustomers 获取客户列表
    ListCustomers(ctx context.Context, req *dtoSelling.ListCustomersReq) ([]*erp.Customer, error)
    
    // AddCompanyToCustomer 将公司添加到客户的允许交易公司列表
    AddCompanyToCustomer(ctx context.Context, customer *erp.Customer, companyName string) error
    
    // 销售订单管理
    // CreateSalesOrder 创建销售订单
    CreateSalesOrder(ctx context.Context, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error)
    
    // UpdateSalesOrder 更新销售订单
    UpdateSalesOrder(ctx context.Context, name string, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error)
    
    // GetSalesOrder 获取销售订单信息
    GetSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error)
    
    // SubmitSalesOrder 提交销售订单
    SubmitSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error)
    
    // CancelSalesOrder 取消销售订单
    CancelSalesOrder(ctx context.Context, name string) error
}
```

## 🏗️ 实现细节

### POS 开账流程

```go
func (s *sSelling) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error) {
    // 1. 获取公司名称
    companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
    
    // 2. 构建开账信息
    openDetails := s.buildOpeningEntryDetails(req.OpenPosEntryDetail)
    reqInfo := s.buildOpeningEntryRequest(req, companyName, openDetails)
    
    // 3. 创建开账记录（草稿）
    resp, err := service.Document().Create(ctx, erp.DocTypePosOpeningEntry, reqInfo)
    
    // 4. 提交开账记录
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosOpeningEntry, openEntry.Name, erp.DocstatusSubmitted)
    
    return &selling.OpenPosEntryResp{...}, nil
}
```

### POS 关账流程

```go
func (s *sSelling) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error) {
    // 1. 获取开账信息
    openEntry, err := s.GetPosOpeningEntry(ctx, req.PosOpenEntryName)
    
    // 2. 查询期间发票
    invoices, err := s.GetPosInvoiceList(ctx, &dtoSelling.GetPosInvoiceListReq{
        PosProfile:            openEntry.PosProfile,
        User:                  openEntry.User,
        Docstatus:             erp.DocstatusSubmitted,
        CustomPosOpeningEntry: req.PosOpenEntryName,
    })
    
    // 3. 构建关账请求
    closeDetails := s.buildClosingEntryDetails(req.ClosePosEntryDetail)
    reqInfo := s.buildClosingEntryRequest(req, closeDetails)
    reqInfo.PosTransactions = s.buildPosTransactions(invoices)
    
    // 4. 创建关账记录
    resp, err := service.Document().Create(ctx, erp.DocTypePosClosingEntry, reqInfo)
    
    // 5. 提交关账记录
    _, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosClosingEntry, closeEntry.Name, erp.DocstatusSubmitted)
    
    return &selling.ClosePosEntryResp{...}, nil
}
```

### POS 发票保存流程

```go
func (s *sSelling) SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
    // 1. 获取开账记录
    openingEntry, err := s.GetPosOpeningEntry(ctx, req.OpenPosEntryName)
    
    res := &selling.SavePosInvoiceResp{}
    
    // 2. 保存物品发票（如果有原材料）
    if len(req.MaterialItems) > 0 {
        invoiceName, err := s.SavePosInvoiceStep(ctx, req, openingEntry, true)
        if err != nil {
            // 失败时取消已创建的发票
            if len(invoiceName) > 0 {
                s.CancelPosInvoice(ctx, invoiceName)
            }
            return nil, err
        }
        res.MaterialInvoiceName = invoiceName
    }
    
    // 3. 保存商品发票
    invoiceName, err := s.SavePosInvoiceStep(ctx, req, openingEntry, false)
    if err != nil {
        // 失败时取消已创建的发票
        if res.MaterialInvoiceName != "" {
            s.CancelPosInvoice(ctx, res.MaterialInvoiceName)
        }
        return nil, err
    }
    res.ProductsInvoiceName = invoiceName
    
    return res, nil
}
```

### 发票构建

```go
func (s *sSelling) buildPosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq, openingEntry *erp.POSOpeningEntry) *erp.POSInvoice {
    posInvoice := &erp.POSInvoice{
        PosProfile:            openingEntry.PosProfile,
        Company:               openingEntry.Company,
        PostingDate:           postingDatetime.Format("Y-m-d"),
        PostingTime:           postingDatetime.Format("H:i:s"),
        Currency:              req.Currency,
        UpdateStock:           0,
        CustomerOrder:         req.OrderNo,
        SetPostingTime:        1,
        CustomPosOpeningEntry: req.OpenPosEntryName,
    }
    
    // 设置客户
    if len(req.CustomerUuid) > 0 {
        posInvoice.Customer = "Member"
        posInvoice.CustomerUUID = req.CustomerUuid
    } else {
        posInvoice.Customer = "Default"
    }
    
    // 构建项目
    posInvoice.Items = s.buildInvoiceItems(req.Items)
    
    // 构建税费
    posInvoice.Taxes = s.buildInvoiceTaxes(req.Taxes, req.CompanyAbbr)
    
    // 构建支付
    posInvoice.Payments = s.buildInvoicePayments(req.Payments)
    
    return posInvoice
}
```

## 📊 数据模型

### OpenPosEntryReq 开账请求

```go
type OpenPosEntryReq struct {
    PosProfileName     string                 // POS 配置名称
    CashierEmail       string                 // 收银员邮箱
    CompanyAbbr        string                 // 公司缩写
    PeriodStartDate    int64                  // 开账时间戳
    OpenPosEntryDetail []*OpenPosEntryDetail  // 开账明细
}

type OpenPosEntryDetail struct {
    ModeOfPayment string  // 支付方式
    OpeningAmount float64 // 开账金额
}
```

### ClosePosEntryReq 关账请求

```go
type ClosePosEntryReq struct {
    PosOpenEntryName    string                  // 开账记录名称
    PeriodEndDate       int64                   // 关账时间戳
    ClosePosEntryDetail []*ClosePosEntryDetail  // 关账明细
}

type ClosePosEntryDetail struct {
    ModeOfPayment string  // 支付方式
    OpeningAmount float64 // 开账金额
    ClosingAmount float64 // 关账金额
}
```

### SavePosInvoiceReq 保存发票请求

```go
type SavePosInvoiceReq struct {
    OpenPosEntryName string             // 开账记录名称
    CompanyAbbr      string             // 公司缩写
    OrderNo          string             // 订单号
    CustomerUuid     string             // 客户 UUID
    PostingDatetime  int64              // 记账时间
    Currency         string             // 币种
    PriceListCurrency string            // 价格列表币种
    UpdateStock      int32              // 是否更新库存
    Items            []*PosInvoiceItem  // 商品项目
    MaterialItems    []*PosInvoiceItem  // 物品项目
    Taxes            []*PosInvoiceTax   // 税费
    Payments         []*PosInvoicePayment // 支付
}
```

### PosInvoiceItem 发票项目

```go
type PosInvoiceItem struct {
    ItemCode    string  // 物品编码
    Qty         float64 // 数量
    Rate        float64 // 单价
    Amount      float64 // 金额
    Description string  // 描述
    Uom         string  // 单位
    IsFreeItem  bool    // 是否赠品
}
```

### PosInvoicePayment 发票支付

```go
type PosInvoicePayment struct {
    ModeOfPayment string  // 支付方式
    Amount        float64 // 金额
}
```

## 🔄 使用流程

### 1. POS 开账

```go
resp, err := sellingService.OpenPosEntry(ctx, &selling.OpenPosEntryReq{
    PosProfileName:  "Wallace Burger POS",
    CashierEmail:    "cashier@example.com",
    CompanyAbbr:     "CFG",
    PeriodStartDate: time.Now().Unix(),
    OpenPosEntryDetail: []*selling.OpenPosEntryDetail{
        {ModeOfPayment: "Cash", OpeningAmount: 1000},
        {ModeOfPayment: "Balance", OpeningAmount: 0},
    },
})

openEntryName := resp.OpenPosEntryInfo.OpenPosEntryName
```

### 2. 保存销售发票

```go
resp, err := sellingService.SavePosInvoice(ctx, &selling.SavePosInvoiceReq{
    OpenPosEntryName: openEntryName,
    CompanyAbbr:      "CFG",
    OrderNo:          "ORD20231201001",
    PostingDatetime:  time.Now().Unix(),
    Currency:         "THB",
    Items: []*selling.PosInvoiceItem{
        {ItemCode: "SP20231201001_00", Qty: 2, Rate: 100, Amount: 200},
    },
    Payments: []*selling.PosInvoicePayment{
        {ModeOfPayment: "Cash", Amount: 200},
    },
})
```

### 3. 退货处理

```go
resp, err := sellingService.ReturnPosInvoice(ctx, &selling.ReturnPosInvoiceReq{
    InvoiceName:      "ACC-PSINV-2023-00001",
    OpenPosEntryName: openEntryName,
    CompanyAbbr:      "CFG",
    OrderNo:          "ORD20231201001-R",
    Items: []*selling.PosInvoiceItem{
        {ItemCode: "SP20231201001_00", Qty: -1, Rate: 100, Amount: -100},
    },
    Payments: []*selling.PosInvoicePayment{
        {ModeOfPayment: "Cash", Amount: -100},
    },
})
```

### 4. POS 关账

```go
resp, err := sellingService.ClosePosEntry(ctx, &selling.ClosePosEntryReq{
    PosOpenEntryName: openEntryName,
    PeriodEndDate:    time.Now().Unix(),
    ClosePosEntryDetail: []*selling.ClosePosEntryDetail{
        {ModeOfPayment: "Cash", OpeningAmount: 1000, ClosingAmount: 1200},
        {ModeOfPayment: "Balance", OpeningAmount: 0, ClosingAmount: 0},
    },
})
```

### 5. 创建客户

```go
customer, err := sellingService.CreateCustomer(ctx, &erp.Customer{
    CustomerName:       "客户A",
    CustomerType:       "Company",
    IsInternalCustomer: 1,
    RepresentsCompany:  "CFG Company",
    CustomerGroup:      "Default Customer Group",
})
```

### 6. 创建销售订单

```go
salesOrder, err := sellingService.CreateSalesOrder(ctx, &dtoSelling.SalesOrder{
    Customer:         "客户A",
    Company:          "CFG Company",
    DeliveryDate:     "2023-12-31",
    SetWarehouse:     "Wallace Burger (CFG)-Normal-Default",
    SellingPriceList: "Selling - Internal",
    Items: []dtoSelling.SalesOrderItem{
        {ItemCode: "SP20231201001", Qty: 10, Uom: "Nos", Rate: 100, Warehouse: "Wallace Burger (CFG)-Normal-Default"},
    },
})

// 提交销售订单
salesOrder, err = sellingService.SubmitSalesOrder(ctx, salesOrder.Name)
```

## ⚠️ 注意事项

1. **开账状态**: 一个 POS Profile + 用户 同时只能有一个开账状态
2. **发票创建者**: POS 发票使用开账收银员身份创建（Fake User）
3. **物品/商品分离**: 有原材料时会创建两张发票
4. **时区处理**: 记账时间需要转换为用户时区
5. **关账发票**: 关账时会查询期间所有已提交发票
6. **失败回滚**: 发票创建失败时需要取消已创建的发票

## 🔮 扩展性

### 异步销售模式

支持将销售记录异步推送到队列：

```go
// 同步模式
resp, err := sellingService.SavePosInvoice(ctx, req)

// 异步模式
queue.Push(string(consts.TopicSellingSave), req)
```

### 批量关账

可扩展批量关账功能：

```go
func (s *sSelling) BatchClosePosEntry(ctx context.Context, openEntryNames []string, endDate int64) error {
    for _, name := range openEntryNames {
        _, err := s.ClosePosEntry(ctx, &selling.ClosePosEntryReq{
            PosOpenEntryName: name,
            PeriodEndDate:    endDate,
        })
        if err != nil {
            g.Log().Errorf(ctx, "关账失败: %s, %v", name, err)
        }
    }
    return nil
}
```

## 📝 总结

Selling 销售服务提供了完整的 POS 销售流程管理能力。

### 技术特点

- **完整流程**: 开账 → 销售 → 退货 → 关账
- **身份模拟**: 以收银员身份创建发票
- **双发票**: 物品和商品分离处理
- **时区转换**: 支持用户时区

### 设计优势

- **事务保证**: 失败时自动回滚
- **灵活配置**: 支持多种支付方式
- **可追溯**: 完整的操作记录

