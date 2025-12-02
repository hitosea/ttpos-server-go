# ERP Print Format Doctype 通用服务支持 设计文档

> 本文档定义 ERP Print Format Doctype 通用服务支持的技术设计和实现方案。

## 📋 概述

ERP Print Format Doctype 通用服务支持通过封装 ERPNext Print Format API，提供统一的打印格式管理服务。核心实现包括：

- 新增 PrintFormatService 封装 Print Format 逻辑
- 复用现有的 Document 和 Doctype 服务
- 遵循现有的服务注册机制

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- ✅ PrintFormatService 依赖 IDocument 和 IDoctype 接口
- ✅ URL 使用 snake_case
- ✅ data 字段返回对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 gerror.Wrapf 包装错误

### API 设计规范 (api.mdc)

- ✅ 响应格式统一: `{code, message, data{}}`
- ✅ data 不能为 null 或数组

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Document Service**: `ttpos-bmp/app/ttpos-erp/internal/service/erpnext.go` - 复用文档 CRUD 逻辑
- **Doctype Service**: `ttpos-bmp/app/ttpos-erp/internal/service/erpnext.go` - 复用元数据查询逻辑
- **ERPNext Client**: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` - 复用客户端实现
- **错误处理**: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` - 复用 detectError 函数

### 集成点

- **Document 模块**: 调用 Document Service 进行 CRUD 操作
- **Doctype 模块**: 调用 Doctype Service 查询元数据
- **ERPNext Client**: 使用现有的客户端和认证机制

---

## 🏗️ 架构设计

### 分层设计

```
PrintFormatLogic (Logic 层)
  ↓ 调用
IDocument + IDoctype (Service 层)
  ↓ 调用
ERPNext Client (基础设施层)
  ↓ 调用
ERPNext API
```

**依赖规则**：

- ✅ PrintFormatLogic 依赖 IDocument 和 IDoctype 接口
- ✅ 复用现有的 ERPNext Client 实现
- ✅ 遵循现有的错误处理机制

---

## 🗄️ 数据结构设计

### Print Format 数据结构

参考 ERPNext Print Format 标准结构：

```go
// Print Format 基本信息
type PrintFormat struct {
    Name        string `json:"name"`         // Print Format 名称
    DocType     string `json:"doc_type"`     // 关联的 DocType
    Standard    int    `json:"standard"`      // 是否标准格式：0-否，1-是
    HTML        string `json:"html"`         // HTML 模板内容
    CSS         string `json:"css"`          // CSS 样式
    PrintFormat string `json:"print_format"` // 打印格式类型：Standard, Jinja, JS
    PrintFormatType string `json:"print_format_type"` // 格式类型
    // ... 其他字段
}
```

### Request DTO

```go
// Print Format 列表查询请求
type PrintFormatListReq struct {
    DocType    string `json:"doc_type"`     // DocType 过滤
    Limit      int    `json:"limit"`        // 分页大小
    LimitStart int    `json:"limit_start"` // 分页起始位置
}

// Print Format 创建/更新请求
type PrintFormatCreateUpdateReq struct {
    Name        string `json:"name" binding:"required"`
    DocType     string `json:"doc_type" binding:"required"`
    HTML        string `json:"html"`
    CSS         string `json:"css"`
    PrintFormat string `json:"print_format"`
    // ... 其他字段
}
```

### Response DTO

```go
// Print Format 详情响应
type PrintFormatDetailResp struct {
    Name        string `json:"name"`
    DocType     string `json:"doc_type"`
    HTML        string `json:"html"`
    CSS         string `json:"css"`
    PrintFormat string `json:"print_format"`
    // ... 其他字段
}

// Print Format 列表响应
type PrintFormatListResp struct {
    List  []*PrintFormatDetailResp `json:"list"`
    Total int                      `json:"total"`
}
```

---

## 🔌 API 设计

### Print Format Meta API

**功能**: 获取 Print Format 的元数据信息

**实现方式**: 调用 `service.Doctype().Meta(ctx, &erp.ErpReq{DocType: "Print Format"})`

**返回**: ERPNext 元数据 JSON 结构

---

### Print Format 列表查询

**功能**: 根据 DocType 查询 Print Format 列表

**实现方式**: 调用 `service.Document().List(ctx, &erp.ErpReq{DocType: "Print Format"}, params)`

**参数**:
- DocType: 过滤条件（可选）
- Limit: 分页大小
- LimitStart: 分页起始位置

**返回**: Print Format 列表

---

### Print Format 详情查询

**功能**: 根据名称查询 Print Format 详情

**实现方式**: 调用 `service.Document().Get(ctx, &erp.ErpReq{DocType: "Print Format", Name: name}, nil)`

**参数**:
- Name: Print Format 名称（必填）

**返回**: Print Format 详细信息（包括 HTML 模板）

---

### Print Format 创建

**功能**: 创建新的 Print Format

**实现方式**: 调用 `service.Document().Create(ctx, "Print Format", data)`

**参数**: Print Format 数据结构

**返回**: 创建后的 Print Format 信息

---

### Print Format 更新

**功能**: 更新现有的 Print Format

**实现方式**: 调用 `service.Document().Update(ctx, &erp.ErpReq{DocType: "Print Format", Name: name}, data)`

**参数**: Print Format 名称和更新数据

**返回**: 更新后的 Print Format 信息

---

### Print Format 删除

**功能**: 删除 Print Format（软删除）

**实现方式**: 调用 `service.Document().Delete(ctx, &erp.ErpReq{DocType: "Print Format", Name: name})`

**参数**: Print Format 名称

**返回**: 删除确认信息

---

## 🧩 核心组件实现

### Logic 接口

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go
package erpnext

import (
    "context"
    "ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
    "ttpos-bmp/app/ttpos-erp/internal/service"
    
    "github.com/gogf/gf/v2/encoding/gjson"
    "github.com/gogf/gf/v2/errors/gerror"
)

const printFormatDocType = "Print Format"

var PrintFormat = new(sPrintFormat)

type sPrintFormat struct {
}

func init() {
    // 注册到 service（如果需要单独的服务接口）
    // service.RegisterPrintFormat(PrintFormat)
}

// Meta 获取 Print Format 元数据
func (s *sPrintFormat) Meta(ctx context.Context, req *erp.ErpReq) (*gjson.Json, error) {
    resp := service.Doctype().Meta(ctx, &erp.ErpReq{
        DocType: printFormatDocType,
    })
    return resp, nil
}

// List 查询 Print Format 列表
func (s *sPrintFormat) List(ctx context.Context, req *erp.PrintFormatListReq) ([]*erp.PrintFormatDetailResp, error) {
    filters := make([][]string, 0)
    
    // 按 DocType 过滤
    if req.DocType != "" {
        filters = append(filters, []string{"doc_type", "=", req.DocType})
    }
    
    // 查询列表
    resp, err := service.Document().List(ctx, &erp.ErpReq{
        DocType: printFormatDocType,
    }, &erp.RequestParams{
        Fields:  []string{"name", "doc_type", "standard", "print_format_type"},
        Filters: filters,
        Limit:   req.Limit,
        LimitStart: req.LimitStart,
    })
    if err != nil {
        return nil, gerror.Wrapf(err, "查询 Print Format 列表失败")
    }
    
    // 解析响应
    list := make([]*erp.PrintFormatDetailResp, 0)
    dataArray := resp.GetJsons("data")
    for _, data := range dataArray {
        item := &erp.PrintFormatDetailResp{}
        if err := data.Scan(item); err != nil {
            continue
        }
        list = append(list, item)
    }
    
    return list, nil
}

// Get 查询 Print Format 详情
func (s *sPrintFormat) Get(ctx context.Context, name string) (*erp.PrintFormatDetailResp, error) {
    if name == "" {
        return nil, gerror.New("Print Format 名称不能为空")
    }
    
    resp, err := service.Document().Get(ctx, &erp.ErpReq{
        DocType: printFormatDocType,
        Name:    name,
    }, nil)
    if err != nil {
        return nil, gerror.Wrapf(err, "查询 Print Format 详情失败")
    }
    
    // 解析响应
    detail := &erp.PrintFormatDetailResp{}
    if err := resp.GetJson("data").Scan(detail); err != nil {
        return nil, gerror.Wrapf(err, "解析 Print Format 详情失败")
    }
    
    return detail, nil
}

// Create 创建 Print Format
func (s *sPrintFormat) Create(ctx context.Context, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error) {
    if req.Name == "" {
        return nil, gerror.New("Print Format 名称不能为空")
    }
    if req.DocType == "" {
        return nil, gerror.New("DocType 不能为空")
    }
    
    resp, err := service.Document().Create(ctx, printFormatDocType, req)
    if err != nil {
        return nil, gerror.Wrapf(err, "创建 Print Format 失败")
    }
    
    // 解析响应
    detail := &erp.PrintFormatDetailResp{}
    if err := resp.GetJson("data").Scan(detail); err != nil {
        return nil, gerror.Wrapf(err, "解析 Print Format 响应失败")
    }
    
    return detail, nil
}

// Update 更新 Print Format
func (s *sPrintFormat) Update(ctx context.Context, name string, req *erp.PrintFormatCreateUpdateReq) (*erp.PrintFormatDetailResp, error) {
    if name == "" {
        return nil, gerror.New("Print Format 名称不能为空")
    }
    
    _, err := service.Document().Update(ctx, &erp.ErpReq{
        DocType: printFormatDocType,
        Name:    name,
    }, req)
    if err != nil {
        return nil, gerror.Wrapf(err, "更新 Print Format 失败")
    }
    
    // 获取更新后的信息
    return s.Get(ctx, name)
}

// Delete 删除 Print Format
func (s *sPrintFormat) Delete(ctx context.Context, name string) error {
    if name == "" {
        return gerror.New("Print Format 名称不能为空")
    }
    
    _, err := service.Document().Delete(ctx, &erp.ErpReq{
        DocType: printFormatDocType,
        Name:    name,
    })
    if err != nil {
        return gerror.Wrapf(err, "删除 Print Format 失败")
    }
    
    return nil
}
```

### 注册到 Logic

在 `logic.go` 中导入：

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/logic.go
import (
    _ "ttpos-bmp/app/ttpos-erp/internal/logic/erpnext"
    // ... 其他导入
)
```

---

## 🚨 错误处理

### 主要错误场景

1. **参数验证失败**: 返回 "参数不能为空" 等错误
2. **ERPNext API 调用失败**: 使用 detectError 处理，返回详细错误信息
3. **Print Format 不存在**: 返回 "Print Format 不存在"
4. **数据解析失败**: 返回 "解析响应失败"

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**: PrintFormatLogic ≥ 70%

**测试用例**:

- Meta 查询成功
- 列表查询（带/不带过滤条件）
- 详情查询（存在/不存在）
- 创建 Print Format
- 更新 Print Format
- 删除 Print Format
- 参数验证

### 集成测试

**测试流程**:

- 创建 Print Format → 查询详情 → 更新 → 查询列表 → 删除

---

## 📈 性能优化

### 优化措施

1. **复用现有服务**: 直接使用 Document 和 Doctype 服务，无需额外封装
2. **错误处理**: 复用现有的 detectError 函数
3. **客户端复用**: 使用现有的 ERPNext Client

### 性能指标

- 查询响应时间: < 500ms
- 创建/更新响应时间: < 1000ms

---

## 📚 实现清单

### Phase 1: 数据结构定义（参见 tasks.md）

- [ ] 定义 Print Format DTO
- [ ] 定义 Request/Response 结构

### Phase 2: Logic 实现（参见 tasks.md）

- [ ] 实现 PrintFormatLogic
- [ ] 实现各个方法
- [ ] 注册到 logic.go

### Phase 3: 测试和优化（参见 tasks.md）

- [ ] 单元测试
- [ ] 集成测试

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当设计结论可复用或踩坑较多时，沉淀 Episode 并在此更新名称，保持 Spec ↔ Graphiti 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**作者**: 后端开发组  
**审核者**: {待分配}

