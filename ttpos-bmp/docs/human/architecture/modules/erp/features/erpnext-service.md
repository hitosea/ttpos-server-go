# ERPNext 集成服务说明文档

## 📋 服务概览

ERPNext 集成服务是 ttpos-erp 模块的核心基础设施，负责封装所有与 ERPNext 系统的交互。该服务提供统一的 API 调用接口，支持多站点切换、身份模拟、错误处理等功能。

## 🎯 主要功能

### HTTP 客户端管理
- **站点切换**: 根据上下文动态切换 ERPNext 站点
- **授权管理**: 自动处理 API Key/Secret 授权
- **身份模拟**: 支持以指定用户身份执行操作
- **请求调试**: 可配置的请求/响应日志

### Document 操作
- **列表查询**: 支持分页、过滤、字段选择
- **单个查询**: 根据名称获取文档详情
- **创建文档**: 创建新的 DocType 实例
- **更新文档**: 更新现有文档
- **删除文档**: 删除文档（软删除）
- **状态变更**: 提交、取消文档

### RPC 调用
- **方法执行**: 调用 ERPNext 自定义 API 方法
- **报表查询**: 执行 ERPNext 报表

## 📁 文件结构

```
internal/logic/erpnext/
├── erpnext.go      # RPC 服务，HTTP 客户端封装
├── document.go     # Document CRUD 操作
├── doctype.go      # DocType 统计查询
├── report.go       # 报表查询服务
├── token.go        # 授权管理
├── resource.go     # 资源文件处理
└── print_format.go # 打印格式管理
```

## 🔧 接口定义

### IRpc 接口

```go
type IRpc interface {
    // Execute 执行 ERPNext 远程方法
    Execute(ctx context.Context, req *ErpReq, params interface{}) (*gjson.Json, error)
    
    // GetSiteCode 获取当前站点编码
    GetSiteCode(ctx context.Context) string
    
    // GetAndProcessSiteAuthorization 获取站点授权信息
    GetAndProcessSiteAuthorization(ctx context.Context, siteCode string) (*SiteAuth, error)
    
    // GetAndProcessCashierAuthorization 获取收银员授权信息
    GetAndProcessCashierAuthorization(ctx context.Context, userEmail string) (string, error)
}
```

### IDocument 接口

```go
type IDocument interface {
    // List 查询文档列表
    List(ctx context.Context, req *ErpReq, params *RequestParams) (*gjson.Json, error)
    
    // Get 获取单个文档
    Get(ctx context.Context, req *ErpReq, params interface{}) (*gjson.Json, error)
    
    // Create 创建文档
    Create(ctx context.Context, doctype string, data interface{}) (*gjson.Json, error)
    
    // Update 更新文档
    Update(ctx context.Context, req *ErpReq, data interface{}) (*gjson.Json, error)
    
    // Delete 删除文档
    Delete(ctx context.Context, req *ErpReq) (*gjson.Json, error)
    
    // ChangeDocStatus 变更文档状态
    ChangeDocStatus(ctx context.Context, doctype, name, status string) (*gjson.Json, error)
}
```

### IDoctype 接口

```go
type IDoctype interface {
    // Count 统计文档数量
    Count(ctx context.Context, req *ErpReq, params *RequestParams) (int, error)
}
```

### IReport 接口

```go
type IReport interface {
    // Run 执行报表查询
    Run(ctx context.Context, params *ReportParams) (*gjson.Json, error)
}
```

## 🏗️ 实现细节

### HTTP 客户端获取

```go
func GetClient(ctx context.Context) *gclient.Client {
    var c = g.Client()
    m := grpcx.Ctx.IncomingMap(ctx)
    
    // 根据站点编码设置授权
    if m.Contains(consts.ContextSiteCode) {
        serviceAuthorization, err := Rpc.GetAndProcessSiteAuthorization(ctx, m.GetVar(consts.ContextSiteCode).String())
        if err == nil {
            c.SetPrefix(serviceAuthorization.SiteUrl)
            c.SetHeader("Authorization", serviceAuthorization.Authorization)
        }
    } else {
        // 使用默认配置
        c.SetPrefix(g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.serviceUrl").String())
    }
    
    // 收银员身份模拟
    if ctx.Value(consts.ContextFakeUser) != nil {
        cashierAuthorization, err := Rpc.GetAndProcessCashierAuthorization(ctx, ctx.Value(consts.ContextFakeUser).(string))
        if err == nil {
            c.SetHeader("Authorization", cashierAuthorization)
        }
    }
    
    return c
}
```

### 错误检测

```go
func detectError(resp *gvar.Var) (*gjson.Json, error) {
    if resp == nil || resp.IsEmpty() {
        return nil, gerror.New("调用erp接口返回空")
    }
    
    if j, err := gjson.DecodeToJson(resp); err == nil {
        // 检查异常类型
        if j.Contains("exc_type") {
            return nil, gerror.Newf("调用erp接口返回异常,exc_type:%s,exception:%s", 
                j.Get("exc_type").String(), j.Get("exception").String())
        }
        
        // 检查错误列表
        if j.Contains("errors") {
            errMsgList := make([]string, 0)
            for _, errItem := range j.GetJsons("errors") {
                if errItem.Contains("message") {
                    errMsgList = append(errMsgList, errItem.Get("message").String())
                }
            }
            return nil, gerror.Newf("调用erp接口返回异常:%s", strings.Join(errMsgList, ";"))
        }
        
        return j, nil
    }
    
    return nil, gerror.Wrapf(err, "调用erp接口返回解析异常")
}
```

### 身份模拟

```go
func SetFakeUser(ctx context.Context, userEmail string) context.Context {
    ctx = context.WithValue(ctx, consts.ContextFakeUser, userEmail)
    return ctx
}
```

## 📊 数据模型

### ErpReq 请求结构

```go
type ErpReq struct {
    DocType string `json:"doctype"` // DocType 名称
    Name    string `json:"name"`    // 文档名称
    Method  string `json:"method"`  // API 方法名
}
```

### RequestParams 查询参数

```go
type RequestParams struct {
    Fields  []string     `json:"fields"`   // 查询字段
    Filters [][]string   `json:"filters"`  // 过滤条件
    Limit   int          `json:"limit"`    // 限制数量
    OrderBy string       `json:"order_by"` // 排序
}
```

### ReportParams 报表参数

```go
type ReportParams struct {
    ReportName           string `json:"report_name"`            // 报表名称
    Filters              string `json:"filters"`                // 过滤条件 JSON
    IgnorePreparedReport bool   `json:"ignore_prepared_report"` // 忽略预生成报表
}
```

### SiteAuth 站点授权

```go
type SiteAuth struct {
    SiteCode      string // 站点编码
    SiteUrl       string // 站点 URL
    Authorization string // 授权头
}
```

## 🔄 使用流程

### 1. 文档列表查询

```go
// 查询物品列表
resp, err := service.Document().List(ctx, &erp.ErpReq{
    DocType: erp.DocTypeItem,
}, &erp.RequestParams{
    Fields:  g.ArrayStr{"item_name", "item_code", "item_group"},
    Filters: [][]string{{"item_group", "=", "CFG Products"}},
    Limit:   100,
})
if err != nil {
    return nil, gerror.Wrapf(err, "查询物品列表失败")
}

// 解析结果
dataArray := resp.GetJsons("data")
for _, item := range dataArray {
    fmt.Printf("物品: %s\n", item.Get("item_name").String())
}
```

### 2. 创建文档

```go
// 创建物品
newItem := g.Map{
    "item_code":  "SP20231201001",
    "item_name":  "测试商品",
    "item_group": "CFG Products",
    "stock_uom":  "Nos",
}

resp, err := service.Document().Create(ctx, erp.DocTypeItem, &newItem)
if err != nil {
    return nil, gerror.Wrapf(err, "创建物品失败")
}

itemCode := resp.Get("data.name").String()
```

### 3. 执行远程方法

```go
// 创建变体商品
resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
    Method: erp.ApiMethodCreateVariantItem,
}, g.Map{
    "item": "SP20231201001",
    "args": `{"Size": "Large"}`,
})
if err != nil {
    return nil, gerror.Wrapf(err, "创建变体商品失败")
}
```

### 4. 以收银员身份操作

```go
// 设置收银员身份
ctx = erpnext.SetFakeUser(ctx, "cashier@example.com")

// 创建 POS 发票（以收银员身份）
resp, err := service.Document().Create(ctx, erp.DocTypePosInvoice, posInvoice)
```

## ⚠️ 注意事项

1. **站点配置**: 确保 `erp_site` 表配置正确的站点信息
2. **API 权限**: API Key 需要有足够的权限访问目标 DocType
3. **并发限制**: 注意 ERPNext 的 API 并发限制
4. **错误处理**: 所有 ERPNext 调用都需要处理错误
5. **超时设置**: 建议设置合理的请求超时时间
6. **日志记录**: 生产环境建议关闭 `dump` 选项

## 🔮 扩展性

### 新 DocType 支持

1. 在 `consts.go` 添加 DocType 常量
2. 在 `dto/erp/` 定义数据结构
3. 在业务逻辑层调用 Document 服务

### 新 API 方法支持

1. 在 `consts.go` 添加方法常量
2. 调用 RPC Execute 方法

### 缓存机制

可在 Document 层添加缓存：

```go
func (s *sDocument) GetWithCache(ctx context.Context, req *ErpReq, params interface{}) (*gjson.Json, error) {
    cacheKey := fmt.Sprintf("erp:%s:%s", req.DocType, req.Name)
    
    // 尝试从缓存获取
    if cached, _ := g.Redis().Get(ctx, cacheKey); !cached.IsNil() {
        return gjson.DecodeToJson(cached.Bytes())
    }
    
    // 从 ERPNext 获取
    result, err := s.Get(ctx, req, params)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    g.Redis().SetEX(ctx, cacheKey, result.String(), 300)
    
    return result, nil
}
```

## 📝 总结

ERPNext 集成服务作为 ttpos-erp 的基础设施层，提供了统一、可靠的 ERPNext 交互能力。

### 技术特点

- **多站点支持**: 动态切换不同 ERPNext 站点
- **身份模拟**: 支持以指定用户身份执行操作
- **统一封装**: 标准化的请求/响应处理
- **错误处理**: 完善的错误检测和转换

### 设计优势

- **解耦合**: 业务层与 ERPNext API 解耦
- **可扩展**: 易于添加新的 DocType 和方法
- **可维护**: 集中的 API 调用管理
- **可测试**: 接口抽象便于 Mock 测试

