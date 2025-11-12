# ERPNext 服务功能说明

## 概述
ERPNext 服务提供与 ERPNext 系统交互的底层接口，包括文档类型、文档操作、RPC 调用、报表和资源管理等核心功能。

## 服务接口

### IDoctype - 文档类型管理
- **Meta**: 获取文档类型的元数据信息
- **Count**: 统计指定文档类型的记录数量

### IDocument - 文档操作
- **List**: 获取文档列表
- **Get**: 获取单个文档详情
- **Create**: 创建新文档
- **Update**: 更新文档
- **Delete**: 删除文档
- **Copy**: 复制文档
- **Execute**: 执行文档方法
- **ChangeDocStatus**: 更改文档状态（草稿/提交/取消）

### IRpc - 远程过程调用
- **Execute**: 执行 RPC 方法调用
- **GetSiteCode**: 获取站点编码

### IReport - 报表管理
- **Run**: 运行报表并获取结果

### IResource - 资源管理
- **List**: 获取资源列表
- **Get**: 获取单个资源
- **Post**: 创建资源
- **Put**: 更新资源
- **Delete**: 删除资源

## 业务场景

### 文档操作
- 所有 ERP 业务单据的 CRUD 操作
- 文档状态流转（草稿→提交→取消）
- 文档复制功能

### RPC 调用
- 调用 ERPNext 服务端方法
- 执行自定义业务逻辑
- 获取系统配置信息

### 报表查询
- 执行标准报表
- 获取业务数据统计
- 支持自定义报表参数

### 资源访问
- RESTful 风格的资源操作
- 批量数据处理
- 资源过滤和查询

## 数据结构

### ErpReq - ERP 请求参数
包含文档类型和文档名称等基本信息

### RequestParams - 请求参数
包含过滤条件、字段选择、排序、分页等查询参数

### ReportParams - 报表参数
包含报表名称和报表执行参数

## 使用说明

### 服务注册
```go
service.RegisterDoctype(doctypeImpl)
service.RegisterDocument(documentImpl)
service.RegisterRpc(rpcImpl)
service.RegisterReport(reportImpl)
service.RegisterResource(resourceImpl)
```

### 服务调用
```go
// 获取文档服务实例
doc := service.Document()

// 创建文档
result, err := doc.Create(ctx, "Sales Order", orderData)

// 更改文档状态
result, err := doc.ChangeDocStatus(ctx, "Sales Order", "SO-001", "1")

// 执行 RPC
rpc := service.Rpc()
result, err := rpc.Execute(ctx, req, params)
```

## 文档状态说明
- **0**: 草稿（Draft）
- **1**: 已提交（Submitted）
- **2**: 已取消（Cancelled）

## 注意事项
1. 所有操作都需要有效的 ERPNext 认证
2. 文档状态变更需要遵循 ERPNext 的工作流规则
3. RPC 调用需要确保方法在 ERPNext 端已定义
4. 报表执行可能耗时较长，建议异步处理
5. 资源操作需要注意权限控制
6. 返回结果统一使用 gjson.Json 格式，便于灵活解析
