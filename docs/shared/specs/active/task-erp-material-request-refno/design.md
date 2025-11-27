# SaveMaterialRequestReq 增加 RefNo 字段 设计文档

> 本文档定义 ttpos-erp stock 模块 SaveMaterialRequestReq 新增 ref_no 字段的技术设计和实现方案。

## 📋 概述

为 `stock.SaveMaterialRequestReq` protobuf 消息新增 `ref_no` 字段，用于存储 ttpos 传递的原始订单号，支持跨系统问题排查和数据追溯。

**技术特点**：
- 仅涉及 protobuf 字段新增，无业务逻辑变更
- 字段可选，完全向后兼容
- 无数据库变更，无性能影响

---

## 🎯 规范对齐

### Go BMP 规范 (ttpos-bmp/.cursor/rules/go-rules.mdc)

- ✅ protobuf 修改后执行 `gf gen pb` 重新生成
- ✅ 禁止手动修改 dao/entity/do/ 目录（不涉及）
- ✅ 遵循 GoFrame 项目结构

### Protobuf 规范 (ttpos-bmp/.cursor/rules/proto-rules.mdc)

- ✅ 字段名使用 snake_case（`ref_no`）
- ✅ 字段编号使用下一个可用编号（10）
- ✅ 字段包含中文注释说明用途

---

## 🔄 代码复用分析

### 现有代码

| 文件 | 说明 |
|------|------|
| `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto` | 现有 protobuf 定义 |
| `ttpos-bmp/app/ttpos-erp/api/stock/stock.pb.go` | 生成的 Go 代码 |
| `ttpos-bmp/app/ttpos-erp/api/stock/stock_grpc.pb.go` | 生成的 gRPC 代码 |
| `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go` | RPC Controller |

### 集成点

- **ttpos → ttpos-erp**: gRPC 调用，调用方可选传入 `ref_no`

---

## 🏗️ 架构设计

### 变更范围

```
stock.proto (修改)
    ↓ gf gen pb
stock.pb.go (自动生成)
stock_grpc.pb.go (自动生成)
```

### 数据流

```
ttpos 调用方
    ↓ SaveMaterialRequest(ref_no: "ORDER-123")
ttpos-erp RPC Controller
    ↓ 接收 ref_no 字段
业务逻辑 (可选日志记录)
```

---

## 📊 数据模型

### Protobuf 定义变更

**修改前** (`stock.proto` 第 28-38 行)：

```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;         // 公司缩写,必填
  string branch = 3;               // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  string source_warehouse = 5;     // 来源仓库，必填
  string target_warehouse = 6;     // 目标仓库，必填
  string purpose = 7;              // 申请目的,可选 默认 Purchase
  string supplier = 8;             // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
}
```

**修改后**：

```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;         // 公司缩写,必填
  string branch = 3;               // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  string source_warehouse = 5;     // 来源仓库，必填
  string target_warehouse = 6;     // 目标仓库，必填
  string purpose = 7;              // 申请目的,可选 默认 Purchase
  string supplier = 8;             // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
  string ref_no = 10;              // 来源单据号，可选，用于跟踪 ttpos 原始订单号
}
```

### 生成的 Go 结构体

```go
// 自动生成，位于 api/stock/stock.pb.go
type SaveMaterialRequestReq struct {
    // ... 现有字段 ...
    RefNo string `protobuf:"bytes,10,opt,name=ref_no,json=refNo,proto3" json:"ref_no,omitempty" dc:"来源单据号，可选，用于跟踪 ttpos 原始订单号"`
}

func (x *SaveMaterialRequestReq) GetRefNo() string {
    if x != nil {
        return x.RefNo
    }
    return ""
}
```

---

## 🔌 API 设计

### gRPC API

**接口**: `StockService.SaveMaterialRequest`

**请求参数变更**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ref_no | string | 否 | 来源单据号，用于跟踪 ttpos 原始订单号 |

**向后兼容性**：
- 字段为可选（proto3 默认可选）
- 不传时默认为空字符串
- 现有调用方无需修改

---

## 🧩 组件和接口

### 无需修改的组件

| 组件 | 原因 |
|------|------|
| RPC Controller | 字段自动映射，无需额外处理 |
| Logic 层 | 字段仅用于追溯，不参与业务逻辑 |
| Service 层 | 无业务逻辑变更 |

### 可选增强（后续迭代）

如需在日志中记录 `ref_no`，可在 Controller 层添加：

```go
func (*Controller) SaveMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *api.ResponseInfo, err error) {
    // 可选：记录来源单据号
    if req.RefNo != "" {
        g.Log().Infof(ctx, "SaveMaterialRequest ref_no: %s", req.RefNo)
    }
    // ... 现有逻辑 ...
}
```

---

## 🧪 测试策略

### 验证测试

1. **字段传递测试**: 验证 `ref_no` 能正确传递和接收
2. **兼容性测试**: 验证不传 `ref_no` 时接口正常工作
3. **序列化测试**: 验证字段序列化/反序列化正确

### 测试方法

使用 gRPC 客户端或单元测试验证：

```go
// 测试示例
req := &stock.SaveMaterialRequestReq{
    CompanyAbbr: "TEST",
    Branch:      "test-branch",
    RefNo:       "ORDER-123456",  // 新增字段
    // ... 其他字段 ...
}
```

---

## 📚 实现清单

### Phase 1: Protobuf 修改

- [ ] 修改 `stock.proto`，新增 `ref_no` 字段
- [ ] 执行 `gf gen pb` 重新生成 Go 代码
- [ ] 验证生成的代码正确

### Phase 2: 验证测试

- [ ] 验证新字段能正确传递和接收
- [ ] 验证向后兼容性

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: -

