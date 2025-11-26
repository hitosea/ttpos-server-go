# Setup 设置服务说明文档

## 📋 服务概览

Setup 设置服务是 ttpos-erp 模块的初始化服务，负责店铺初始化、用户管理、POS Profile 创建、文档初始化等系统配置功能。该服务是系统启动和店铺开通的核心服务。

## 🎯 主要功能

### 店铺初始化
- **初始化店铺**: 完整的店铺初始化流程（分支、用户、仓库、POS Profile）
- **创建分支**: 创建店铺分支
- **创建用户**: 创建收银员用户账户
- **创建 POS Profile**: 创建默认 POS 配置文件

### 文档初始化
- **自定义字段初始化**: 从目录批量创建自定义字段
- **客户初始化**: 批量创建客户文档
- **支付方式初始化**: 批量创建支付方式
- **通用文档初始化**: 支持从目录批量初始化任意 DocType

### 时间管理
- **时区转换**: 获取默认时区（Asia/Shanghai）
- **本地时间转换**: UTC 时间转换为本地时间

## 📁 文件结构

```
internal/logic/setup/
└── setup.go                   # 设置服务主逻辑
```

## 🔧 接口定义

### ISetup 接口

```go
type ISetup interface {
    // InitShop 初始化店铺
    InitShop(ctx context.Context, req *setup.InitShopReq) (*setup.InitShopResp, error)
    
    // CreateBranch 创建分店
    CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, error)
    
    // CreateUser 创建网站用户
    CreateUser(ctx context.Context, req *setup2.CreateUserInp) error
    
    // CreateDefaultPosProfile 创建默认的POS配置文件
    CreateDefaultPosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posProfileName string, error)
    
    // GetUserApiKeySecret 获取用户的API密钥和密钥
    GetUserApiKeySecret(ctx context.Context, userEmail string) (apiKey, apiSecret string, error)
    
    // InitCustomFields 初始化自定义字段
    InitCustomFields(ctx context.Context, dirBase string) error
    
    // InitCustomers 初始化客户
    InitCustomers(ctx context.Context, dirBase string) error
    
    // InitModeOfPayment 初始化支付方式
    InitModeOfPayment(ctx context.Context, dirBase string) error
    
    // InitErpDocTypeWithDirname 从目录初始化 ERP DocType
    InitErpDocTypeWithDirname(ctx context.Context, dirBase string) error
    
    // GetDefaultTimeZone 获取默认时区
    GetDefaultTimeZone(ctx context.Context) string
    
    // MustGetLocalDateTime 获取本地时间
    MustGetLocalDateTime(ctx context.Context, utcTime *gtime.Time) *gtime.Time
}
```

## 🏗️ 实现细节

### 店铺初始化流程

完整的店铺初始化流程包括：

1. **创建分支**
2. **创建收银员用户**
3. **获取用户 API Key/Secret**
4. **记录门店管理员关联关系**
5. **创建默认仓库**
6. **创建在途仓库**
7. **创建支付方式账户**
8. **创建默认 POS Profile**
9. **创建内部供应商**（连锁店模式）

```go
func (s *sSetup) InitShop(ctx context.Context, req *setup.InitShopReq) (*setup.InitShopResp, error) {
    // 1. 创建分支
    branchName, err := s.CreateBranch(ctx, req)
    
    // 2. 创建收银员用户
    adminEmail := fmt.Sprintf("%s@ttpos-user.com", req.AdminUuid)
    err = s.CreateUser(ctx, &setup2.CreateUserInp{
        UserEmail: adminEmail,
        FirstName: req.ShopName,
    })
    
    // 3. 获取 API Key/Secret
    apiKey, apiSecret, err := s.GetUserApiKeySecret(ctx, adminEmail)
    
    // 4. 记录门店管理员关联关系
    _, err = dao.ShopCashier.Ctx(ctx).Insert(&do.ShopCashier{
        ShopUuid:     req.ShopUuid,
        AdminUuid:    req.AdminUuid,
        CashierEmail: adminEmail,
        ApiKey:       apiKey,
        ApiSecret:    apiSecret,
        CompanyAbbr:  req.CompanyAbbr,
        Branch:       branchName,
        SiteCode:     siteCode,
    })
    
    // 5. 创建默认仓库
    _, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
        Branch:      branchName,
        WhType:      "Normal",
        AliasName:   "Default",
        CompanyAbbr: req.CompanyAbbr,
    })
    
    // 6. 创建在途仓库
    _, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
        Branch:      branchName,
        WhType:      "Transit",
        AliasName:   "Transit",
        CompanyAbbr: req.CompanyAbbr,
    })
    
    // 7. 创建支付方式账户
    err = s.createPaymentAccounts(ctx, req.CompanyAbbr)
    
    // 8. 创建默认 POS Profile
    posProfileName, err := s.CreateDefaultPosProfile(ctx, &setup.CreateDefaultPosProfileReq{
        Name:        fmt.Sprintf("%s - POS", branchName),
        CompanyAbbr: req.CompanyAbbr,
        Branch:      branchName,
        Cashiers: []*setup.PosProfileCashier{{User: adminEmail}},
    })
    
    // 9. 创建内部供应商（连锁店模式）
    if siteCode != consts.SiteCodeTtpos {
        // 添加供应商交易对象
        err = service.Supplier().AddSupplerTransactCompany(ctx, &dto.AddSupplerTransactCompanyReq{
            Supplier:        erp.HeadquartersSupplier,
            WithCompanyAbbr: req.CompanyAbbr,
        })
        
        // 创建内部供应商
        _, err = service.Supplier().CreateSupplier(ctx, &buying.CreateSupplierReq{
            Supplier: &buying.SupplierData{
                AliasName:          branchName,
                Branch:             branchName,
                CompanyAbbr:        req.CompanyAbbr,
                IsInternalSupplier: true,
            },
        })
    }
    
    return &setup.InitShopResp{
        BranchName: branchName,
        AdminEmail: adminEmail,
        PosProfile: posProfileName,
    }, nil
}
```

### 用户创建

创建收银员用户账户：

```go
func (s *sSetup) CreateUser(ctx context.Context, req *setup2.CreateUserInp) error {
    userPayload := g.Map{
        "email":              req.UserEmail,
        "first_name":         req.FirstName,
        "user_type":          "Website User",
        "enabled":            1,
        "send_welcome_email": false,
        "role_profile_name":  "ShopCashier",
        "time_zone":          "Asia/Bangkok",
    }
    
    _, err := service.Document().Create(ctx, "User", userPayload)
    return err
}
```

### 文档初始化

支持从目录批量初始化文档，自动处理数字前缀：

```go
func (s *sSetup) InitErpDocTypeWithDirname(ctx context.Context, dirBase string) error {
    // 读取目录中的所有文件
    files, err := os.ReadDir(dirBase)
    
    // 过滤出目录并按名称排序
    var dirs []string
    for _, file := range files {
        if file.IsDir() {
            dirs = append(dirs, file.Name())
        }
    }
    sort.Strings(dirs)
    
    // 按照排序后的顺序处理每个目录
    for _, dirName := range dirs {
        // 删除数字前缀（如 "00_"）
        prunedDirName := s.pruneNumericPrefix(dirName)
        
        // 转换为 Title Case 作为 DocType
        docType := utility.ConvertToTitleCase(prunedDirName)
        
        // 初始化文档
        err := s.initDocumentsFromDir(ctx, DocumentInitConfig{
            DirBase:  dirBase,
            DirName:  dirName,
            DocType:  docType,
            ItemName: docType,
        })
    }
    
    return nil
}
```

### 时区管理

默认时区为 `Asia/Shanghai`（东八区）：

```go
func (s *sSetup) GetDefaultTimeZone(ctx context.Context) string {
    return "Asia/Shanghai"
}

func (s *sSetup) MustGetLocalDateTime(ctx context.Context, utcTime *gtime.Time) *gtime.Time {
    localTime, err := utcTime.ToZone(s.GetDefaultTimeZone(ctx))
    if err != nil {
        g.Log().Error(ctx, "转换时区失败", err)
        return utcTime // 转换失败返回 UTC 时间
    }
    return localTime
}
```

## 📊 数据模型

### InitShopReq 初始化店铺请求

```go
type InitShopReq struct {
    ShopUuid    string // 店铺 UUID
    ShopName    string // 店铺名称
    AdminUuid   string // 管理员 UUID
    CompanyAbbr string // 公司缩写
}
```

### InitShopResp 初始化店铺响应

```go
type InitShopResp struct {
    BranchName string // 分支名称
    AdminEmail string // 管理员邮箱
    PosProfile string // POS Profile 名称
}
```

### CreateUserInp 创建用户输入

```go
type CreateUserInp struct {
    UserEmail string // 用户邮箱
    FirstName string // 名字
    AdminUuid string // 管理员 UUID
    ShopUuid  string // 店铺 UUID
}
```

## 🔄 使用流程

### 1. 初始化店铺

```go
resp, err := setupService.InitShop(ctx, &setup.InitShopReq{
    ShopUuid:    "shop-uuid-123",
    ShopName:    "Wallace Burger (CFG)",
    AdminUuid:   "admin-uuid-123",
    CompanyAbbr: "CFG",
})

fmt.Printf("分支: %s\n", resp.BranchName)
fmt.Printf("管理员邮箱: %s\n", resp.AdminEmail)
fmt.Printf("POS Profile: %s\n", resp.PosProfile)
```

### 2. 初始化自定义字段

```go
err := setupService.InitCustomFields(ctx, "manifest/erp-migrate/v2.5")
```

### 3. 初始化客户

```go
err := setupService.InitCustomers(ctx, "manifest/erp-migrate/v2.5")
```

### 4. 时区转换

```go
utcTime := gtime.Now()
localTime := setupService.MustGetLocalDateTime(ctx, utcTime)
fmt.Printf("本地时间: %s\n", localTime.Format("Y-m-d H:i:s"))
```

## ⚠️ 注意事项

1. **店铺 UUID**: 店铺 UUID 必须唯一
2. **重复初始化**: 店铺可以重复初始化，已存在的资源会被跳过
3. **API Key/Secret**: 用户创建后会自动生成 API Key/Secret
4. **目录结构**: 文档初始化目录需要按数字前缀排序
5. **时区**: 默认使用 Asia/Shanghai 时区
6. **连锁店模式**: 连锁店模式下会自动创建内部供应商

## 📝 总结

Setup 设置服务是系统初始化的核心服务，提供了完整的店铺初始化能力。

### 技术特点

- **完整流程**: 从分支创建到 POS Profile 配置的完整流程
- **批量初始化**: 支持从目录批量初始化文档
- **时区管理**: 统一的时区转换管理
- **自动化**: 自动创建所需的基础配置

### 设计优势

- **一键初始化**: 通过一个接口完成所有初始化工作
- **容错处理**: 支持重复初始化，已存在的资源会被跳过
- **灵活扩展**: 支持自定义文档类型的初始化

