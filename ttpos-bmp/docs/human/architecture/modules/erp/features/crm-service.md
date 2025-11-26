# CRM 服务说明文档

## 📋 服务概览

CRM 服务是 ttpos-erp 模块的客户关系管理服务，负责联系人和地址的管理。该服务提供联系人和地址的 CRUD 操作，支持动态关联到其他文档类型（如客户、供应商等）。

## 🎯 主要功能

### 联系人管理
- **联系人列表**: 查询联系人信息列表
- **联系人详情**: 获取联系人完整信息
- **创建联系人**: 创建新的联系人记录
- **更新联系人**: 更新联系人信息
- **删除联系人**: 删除联系人记录

### 地址管理
- **地址列表**: 查询地址信息列表
- **地址详情**: 获取地址完整信息
- **创建地址**: 创建新的地址记录
- **更新地址**: 更新地址信息
- **删除地址**: 删除地址记录

## 📁 文件结构

```
internal/logic/crm/
├── crm.go                    # CRM 服务主逻辑
├── contact.go                # 联系人管理
└── address.go                # 地址管理

api/crm/
├── crm.proto                 # CRM 服务 Protobuf 定义
```

## 🔧 接口定义

### ICrm 接口

```go
type ICrm interface {
    // GetContactList 获取联系人列表
    GetContactList(ctx context.Context, req *crm.GetContactListReq) (*crm.GetContactListResp, error)
    
    // GetContact 获取联系人详情
    GetContact(ctx context.Context, req *crm.GetContactReq) (*crm.GetContactResp, error)
    
    // CreateContact 创建联系人
    CreateContact(ctx context.Context, req *crm.CreateContactReq) (*crm.CreateContactResp, error)
    
    // UpdateContact 更新联系人
    UpdateContact(ctx context.Context, req *crm.UpdateContactReq) (*crm.UpdateContactResp, error)
    
    // DeleteContact 删除联系人
    DeleteContact(ctx context.Context, req *crm.DeleteContactReq) (*crm.DeleteContactResp, error)
    
    // GetAddressList 获取地址列表
    GetAddressList(ctx context.Context, req *crm.GetAddressListReq) (*crm.GetAddressListResp, error)
    
    // GetAddress 获取地址详情
    GetAddress(ctx context.Context, req *crm.GetAddressReq) (*crm.GetAddressResp, error)
    
    // CreateAddress 创建地址
    CreateAddress(ctx context.Context, req *crm.CreateAddressReq) (*crm.CreateAddressResp, error)
    
    // UpdateAddress 更新地址
    UpdateAddress(ctx context.Context, req *crm.UpdateAddressReq) (*crm.UpdateAddressResp, error)
    
    // DeleteAddress 删除地址
    DeleteAddress(ctx context.Context, req *crm.DeleteAddressReq) (*crm.DeleteAddressResp, error)
}
```

## 🏗️ 实现细节

### 联系人创建

联系人支持动态关联到其他文档类型：

```go
func (s *sCrm) CreateContact(ctx context.Context, req *crm.CreateContactReq) (*crm.CreateContactResp, error) {
    // 构建联系人数据
    contactData := s.buildContactCreateData(req)
    
    // 调用ERP API创建联系人
    resp, err := service.Document().Create(ctx, erp.DocTypeContact, contactData)
    
    // 解析响应获取创建的联系人名称
    contactName := resp.Get("data.name").String()
    
    return &crm.CreateContactResp{
        Name:    contactName,
        Success: true,
        Message: "联系人创建成功",
    }, nil
}
```

### 地址创建

地址支持动态关联到其他文档类型：

```go
func (s *sCrm) CreateAddress(ctx context.Context, req *crm.CreateAddressReq) (*crm.CreateAddressResp, error) {
    // 构建地址数据
    addressData := s.buildAddressCreateData(req)
    
    // 调用ERP API创建地址
    resp, err := service.Document().Create(ctx, erp.DocTypeAddress, addressData)
    
    // 解析响应获取创建的地址名称
    addressName := resp.Get("data.name").String()
    
    return &crm.CreateAddressResp{
        Name: addressName,
    }, nil
}
```

### 动态关联

联系人和地址都支持通过 `Links` 字段关联到其他文档类型：

```go
type DynamicLink struct {
    LinkDoctype string // 关联文档类型（如 "Customer", "Supplier"）
    LinkName    string // 关联文档名称
    LinkTitle   string // 关联文档标题
}
```

## 📊 数据模型

### ContactInfo 联系人信息

```go
type ContactInfo struct {
    Name             string              // 联系人名称
    FirstName        string              // 名字
    MiddleName       string              // 中间名
    LastName         string              // 姓氏
    FullName         string              // 全名
    EmailId          string              // 邮箱
    Phone            string              // 电话
    MobileNo         string              // 手机
    Status           string              // 状态
    Salutation       string              // 称谓
    Designation      string              // 职位
    Gender           string              // 性别
    CompanyName      string              // 公司名称
    Department       string              // 部门
    IsPrimaryContact bool                // 是否主要联系人
    IsBillingContact bool                // 是否账单联系人
    PhoneNos         []*ContactPhoneInfo // 电话号码列表
}
```

### AddressInfo 地址信息

```go
type AddressInfo struct {
    Name              string // 地址名称
    AddressTitle      string // 地址标题
    AddressType       string // 地址类型（Billing/Shipping/Office等）
    AddressLine1      string // 地址行1
    AddressLine2      string // 地址行2
    City              string // 城市
    County            string // 县
    State             string // 州/省
    Country           string // 国家
    Pincode           string // 邮编
    EmailId           string // 邮箱
    Phone             string // 电话
    Fax               string // 传真
    IsPrimaryAddress  bool   // 是否主要地址
    IsShippingAddress bool   // 是否发货地址
    Disabled          bool   // 是否禁用
}
```

## 🔄 使用流程

### 1. 创建联系人

```go
resp, err := crmService.CreateContact(ctx, &crm.CreateContactReq{
    FirstName: "张三",
    LastName:  "张",
    EmailId:   "zhangsan@example.com",
    Phone:     "1234567890",
    Links: []*crm.DynamicLinkInfo{
        {
            LinkDoctype: "Customer",
            LinkName:    "Customer-001",
            LinkTitle:   "客户A",
        },
    },
})
```

### 2. 创建地址

```go
resp, err := crmService.CreateAddress(ctx, &crm.CreateAddressReq{
    AddressTitle: "总部地址",
    AddressType:  "Billing",
    AddressLine1: "123 Main Street",
    City:         "Bangkok",
    Country:      "Thailand",
    Links: []*crm.DynamicLinkInfo{
        {
            LinkDoctype: "Customer",
            LinkName:    "Customer-001",
            LinkTitle:   "客户A",
        },
    },
})
```

### 3. 查询联系人列表

```go
resp, err := crmService.GetContactList(ctx, &crm.GetContactListReq{
    Status: "Open",
    Limit:  100,
})

for _, contact := range resp.ContactList {
    fmt.Printf("联系人: %s - %s\n", contact.FullName, contact.EmailId)
}
```

## ⚠️ 注意事项

1. **动态关联**: 联系人和地址通过 Links 字段关联到其他文档
2. **必填字段**: 联系人至少需要 FirstName，地址至少需要 AddressTitle、AddressLine1、City、Country
3. **主要联系人**: 可以设置主要联系人和账单联系人
4. **地址类型**: 地址类型包括 Billing、Shipping、Office 等

## 📝 总结

CRM 服务提供了联系人和地址的管理能力。

### 技术特点

- **动态关联**: 支持关联到多种文档类型
- **灵活查询**: 支持多种过滤条件
- **完整 CRUD**: 提供完整的增删改查操作

### 设计优势

- **通用性**: 可以关联到客户、供应商等多种文档类型
- **易于使用**: 接口简洁，易于使用

