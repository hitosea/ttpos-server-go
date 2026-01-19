# Manufacturing 制造服务说明文档

## 📋 服务概览

Manufacturing 制造服务是 ttpos-erp 模块的制造管理服务，负责 BOM（物料清单）的管理。该服务与 ERPNext 的制造模块对接，提供 BOM 的创建、查询、更新等功能。

## 🎯 主要功能

### BOM 管理
- **BOM 列表**: 支持多条件过滤查询（公司、分支、商品、激活状态等）
- **BOM 详情**: 获取 BOM 完整信息
- **创建 BOM**: 创建新的 BOM（包含明细项目）
- **权限过滤**: 支持子公司权限过滤

## 📁 文件结构

```
internal/logic/manufacturing/
└── bom.go                    # BOM 管理逻辑

api/manufacturing/
├── manufacturing.proto       # 制造服务 Protobuf 定义
```

## 🔧 接口定义

### IBom 接口

```go
type IBom interface {
    // GetBomList 获取BOM列表
    GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (*manufacturing.GetBomListResp, error)
    
    // GetBom 根据BOM名称获取单个BOM详细信息
    GetBom(ctx context.Context, req *manufacturing.GetBomReq) (*erp.Bom, error)
    
    // SaveBom 保存BOM信息
    SaveBom(ctx context.Context, req *manufacturing.SaveBomReq) (*manufacturing.SaveBomResp, error)
}
```

## 🏗️ 实现细节

### BOM 列表查询

支持的过滤条件：
- `CompanyAbbr` - 公司缩写
- `Branch` - 分支机构
- `ItemCode` - 商品编码
- `IsActive` - 是否激活
- `IsDefault` - 是否默认
- `SubCompanyAbbr` - 子公司缩写（权限过滤）

```go
func (s *sBom) GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (*manufacturing.GetBomListResp, error) {
    // 构建查询过滤条件
    filters, err := s.buildBomListFilters(ctx, req)
    
    // 调用ERP接口获取BOM列表
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: erp.DocTypeBom,
    }, &erp.RequestParams{
        Fields:  []string{"name", "item", "uom", "quantity", "is_active", "is_default"},
        Filters: filters,
        Limit:   consts.Limit9999,
    })
    
    // 解析响应数据并构建BOM列表
    bomList, err := s.parseBomListResponse(ctx, resp.MustToJson(), req)
    
    return &manufacturing.GetBomListResp{BomList: bomList}, nil
}
```

### BOM 创建流程

创建 BOM 后会自动提交：

```go
func (s *sBom) SaveBom(ctx context.Context, req *manufacturing.SaveBomReq) (*manufacturing.SaveBomResp, error) {
    // 1. 获取公司名称
    companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
    
    // 2. 构建BOM数据结构
    bomInfo := s.buildBomData(req, companyName)
    
    // 3. 调用ERP接口创建BOM
    resp, err := service.Document().Create(ctx, "BOM", bomInfo)
    
    // 4. 解析创建BOM的响应结果
    bomName, err := s.parseSaveBomResponse(ctx, resp.MustToJson())
    
    // 5. 提交BOM（创建后是草稿状态）
    _, err = service.Document().ChangeDocStatus(ctx, "BOM", bomName, erp.DocstatusSubmitted)
    
    return &manufacturing.SaveBomResp{BomName: bomName}, nil
}
```

### 权限过滤

BOM 列表查询支持子公司权限过滤：

```go
// 遍历数据数组，构建BOM信息列表
for _, item := range dataArray {
    // 如果指定了子公司，需要检查权限
    if len(req.SubCompanyAbbr) > 0 {
        bomInfo, err := s.GetBom(ctx, &manufacturing.GetBomReq{
            BomName: item.Get("name").String(),
        })
        
        // 检查权限
        hasPermission, err := service.Permission().CheckPermission(ctx, bomInfo.CustomPermissionRule, subCompanyName)
        if !hasPermission {
            continue // 当前子公司无权限，跳过该BOM
        }
    }
    
    // 添加到结果列表
    bomList = append(bomList, bomInfo)
}
```

## 📊 数据模型

### BomInfo BOM 信息

```go
type BomInfo struct {
    ItemCode  string  // 商品编码
    BomName   string   // BOM 名称
    Uom       string   // 单位
    Quantity  float64  // 数量
    IsActive  bool     // 是否激活
    IsDefault bool     // 是否默认
    Company   string   // 公司缩写
}
```

### Bom BOM 详情

```go
type Bom struct {
    Name       string      // BOM 名称
    Item       string      // 商品编码
    Company    string      // 公司
    Uom        string      // 单位
    Quantity   float64     // 数量
    IsActive   bool        // 是否激活
    IsDefault  bool        // 是否默认
    Items      []BomItem   // BOM 明细项目
    CustomPermissionRule []PermissionRule // 权限规则
}
```

### BomItem BOM 明细项目

```go
type BomItem struct {
    ItemCode string  // 物品编码
    Rate     float64 // 单价
    Qty      float64 // 数量
    Uom      string  // 单位
}
```

## 🔄 使用流程

### 1. 查询 BOM 列表

```go
resp, err := bomService.GetBomList(ctx, &manufacturing.GetBomListReq{
    CompanyAbbr: "CFG",
    ItemCode:    "SP20231201001",
    IsActive:    true,
})

for _, bom := range resp.BomList {
    fmt.Printf("BOM: %s - %s\n", bom.BomName, bom.ItemCode)
}
```

### 2. 获取 BOM 详情

```go
bom, err := bomService.GetBom(ctx, &manufacturing.GetBomReq{
    BomName: "BOM-2023-00001",
})

fmt.Printf("商品: %s, 数量: %.2f %s\n", bom.Item, bom.Quantity, bom.Uom)
```

### 3. 创建 BOM

```go
resp, err := bomService.SaveBom(ctx, &manufacturing.SaveBomReq{
    CompanyAbbr: "CFG",
    ItemCode:    "SP20231201001",
    Uom:         "Nos",
    Quantity:    1,
    IsActive:    true,
    IsDefault:   true,
    Items: []*manufacturing.BomItem{
        {ItemCode: "WPR20231201001", Qty: 0.5, Uom: "Nos", Rate: 10},
        {ItemCode: "WPR20231201002", Qty: 0.3, Uom: "Nos", Rate: 5},
    },
})

fmt.Printf("BOM 创建成功: %s\n", resp.BomName)
```

## ⚠️ 注意事项

1. **BOM 提交**: 创建 BOM 后会自动提交
2. **权限过滤**: 列表查询支持子公司权限过滤
3. **默认 BOM**: 一个商品可以有多个 BOM，但只有一个默认 BOM
4. **激活状态**: 只有激活的 BOM 才会在生产中使用

## 📝 总结

Manufacturing 制造服务提供了 BOM 管理能力。

### 技术特点

- **权限过滤**: 支持子公司权限过滤
- **自动提交**: 创建后自动提交 BOM
- **灵活查询**: 支持多种过滤条件

### 设计优势

- **简单易用**: 接口简洁，易于使用
- **权限控制**: 完善的权限过滤机制

