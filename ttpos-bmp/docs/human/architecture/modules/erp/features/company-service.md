# Company 公司服务说明文档

## 📋 服务概览

Company 公司服务是 ttpos-erp 模块的基础服务，负责公司信息的查询和管理。该服务提供公司列表查询、公司详情获取、子公司管理等功能，是其他服务的基础依赖。

## 🎯 主要功能

### 公司查询
- **公司列表**: 支持多条件过滤查询（公司名称、缩写、父公司）
- **公司详情**: 获取公司完整信息
- **公司名称转换**: 根据公司缩写获取公司名称

### 子公司管理
- **查询子公司**: 判断公司是否有子公司
- **递归查询**: 递归查询所有层级的子公司
- **子公司列表**: 获取直接子公司列表

## 📁 文件结构

```
internal/logic/company/
└── company.go                 # 公司服务主逻辑

api/company/
├── company.proto              # 公司服务 Protobuf 定义
```

## 🔧 接口定义

### ICompany 接口

```go
type ICompany interface {
    // GetCompanyList 获取公司列表
    GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (*company.GetCompanyListResp, error)
    
    // GetCompany 根据公司名称获取公司信息
    GetCompany(ctx context.Context, name string) (*erp.Company, error)
    
    // GetCompanyWithAbbr 根据公司缩写获取公司信息
    GetCompanyWithAbbr(ctx context.Context, abbr string) (*company.CompanyInfo, error)
    
    // GetCompanyNameWithAbbr 根据公司简称获取公司名称
    GetCompanyNameWithAbbr(ctx context.Context, companyAbbr string) (string, error)
    
    // HasSubCompany 判断公司是否有子公司
    HasSubCompany(ctx context.Context, companyName string) (bool, error)
    
    // GetAllSubCompanies 递归查询指定公司下的所有子公司
    GetAllSubCompanies(ctx context.Context, companyName string) ([]*company.CompanyInfo, error)
}
```

## 🏗️ 实现细节

### 公司列表查询

支持的过滤条件：
- `CompanyName` - 公司名称（支持精确匹配和模糊匹配）
- `CompanyAbbr` - 公司缩写（支持精确匹配和模糊匹配）
- `ParentCompany` - 父公司名称（支持精确匹配和模糊匹配）

```go
func (s *sCompany) GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (*company.GetCompanyListResp, error) {
    filters := make([][]string, 0)
    
    // 按公司名称过滤（支持 % 通配符）
    if len(req.CompanyName) > 0 {
        if strings.Contains(req.CompanyName, "%") {
            filters = append(filters, g.ArrayStr{"name", "like", req.CompanyName})
        } else {
            filters = append(filters, g.ArrayStr{"name", "=", req.CompanyName})
        }
    }
    
    // 按公司缩写过滤
    if len(req.CompanyAbbr) > 0 {
        if strings.Contains(req.CompanyAbbr, "%") {
            filters = append(filters, g.ArrayStr{"abbr", "like", req.CompanyAbbr})
        } else {
            filters = append(filters, g.ArrayStr{"abbr", "=", req.CompanyAbbr})
        }
    }
    
    // 查询公司列表
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: "Company",
    }, &erp.RequestParams{
        Fields:  g.ArrayStr{"name", "abbr", "parent_company"},
        Filters: filters,
        Limit:   1000,
    })
    
    // 解析并返回结果
    // ...
}
```

### 递归查询子公司

递归查询所有层级的子公司，支持多层级公司结构：

```go
func (s *sCompany) GetAllSubCompanies(ctx context.Context, companyName string) ([]*company.CompanyInfo, error) {
    allSubCompanies := make([]*company.CompanyInfo, 0)
    
    // 递归查询子公司
    err := s.recursiveQuerySubCompanies(ctx, companyName, &allSubCompanies)
    if err != nil {
        return nil, gerror.Wrapf(err, "递归查询子公司失败")
    }
    
    return allSubCompanies, nil
}

func (s *sCompany) recursiveQuerySubCompanies(ctx context.Context, parentCompanyName string, result *[]*company.CompanyInfo) error {
    // 查询直接子公司
    filters := [][]string{{"parent_company", "=", parentCompanyName}}
    
    erpCompanyList, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: "Company",
    }, &erp.RequestParams{
        Fields:  g.ArrayStr{"name", "abbr", "parent_company"},
        Filters: filters,
        Limit:   1000,
    })
    
    // 遍历直接子公司
    for _, item := range dataArray {
        companyInfo := &company.CompanyInfo{
            CompanyName:   item.Get("name").String(),
            CompanyAbbr:   item.Get("abbr").String(),
            ParentCompany: item.Get("parent_company").String(),
        }
        
        // 添加到结果集
        *result = append(*result, companyInfo)
        
        // 递归查询当前子公司的子公司
        err = s.recursiveQuerySubCompanies(ctx, companyInfo.CompanyName, result)
    }
    
    return nil
}
```

## 📊 数据模型

### CompanyInfo 公司信息

```go
type CompanyInfo struct {
    CompanyName   string // 公司名称
    CompanyAbbr   string // 公司缩写
    ParentCompany string // 父公司名称
}
```

### Company 公司详情

```go
type Company struct {
    Name          string // 公司名称
    Abbr          string // 公司缩写
    ParentCompany string // 父公司名称
    // ... 更多字段
}
```

## 🔄 使用流程

### 1. 查询公司列表

```go
resp, err := companyService.GetCompanyList(ctx, &company.GetCompanyListReq{
    CompanyAbbr: "CFG",
})

for _, company := range resp.CompanyList {
    fmt.Printf("公司: %s (%s)\n", company.CompanyName, company.CompanyAbbr)
}
```

### 2. 根据缩写获取公司名称

```go
companyName, err := companyService.GetCompanyNameWithAbbr(ctx, "CFG")
// 返回: "CFG Company"
```

### 3. 查询所有子公司

```go
subCompanies, err := companyService.GetAllSubCompanies(ctx, "CFG Company")
for _, subCompany := range subCompanies {
    fmt.Printf("子公司: %s\n", subCompany.CompanyName)
}
```

### 4. 判断是否有子公司

```go
hasSub, err := companyService.HasSubCompany(ctx, "CFG Company")
if hasSub {
    fmt.Println("该公司有子公司")
}
```

## ⚠️ 注意事项

1. **公司缩写唯一性**: 公司缩写在 ERPNext 中应该唯一
2. **递归查询**: 递归查询子公司时注意性能，避免过深的层级
3. **模糊匹配**: 使用 `%` 通配符进行模糊匹配
4. **父公司关系**: 通过 `parent_company` 字段建立公司层级关系

## 📝 总结

Company 公司服务是其他服务的基础依赖，提供了公司信息的查询和管理能力。

### 技术特点

- **灵活查询**: 支持精确匹配和模糊匹配
- **层级管理**: 支持多层级公司结构
- **递归查询**: 递归查询所有层级的子公司

### 设计优势

- **基础服务**: 为其他服务提供公司信息查询能力
- **简单高效**: 接口简洁，易于使用

