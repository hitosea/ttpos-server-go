# task-all-purchase-order-multi-name-query 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 1 |
| 总任务数 | 4 |
| 已完成 | 4 |
| 完成率 | 100% |

---

## Phase 1: 核心实现

### 1.1 修改 Protobuf 定义

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/buying/buying.proto` |
| Purpose | 添加 name 字段到 GetPurchaseOrderListReq |
| Requirements | REQ1: name 字段支持 IN 查询 |
| Leverage | 参考现有字段定义格式 |

**具体改动**:
```protobuf
message GetPurchaseOrderListReq {
  // ... 现有字段
  string name = 7; // 新增：采购订单名称过滤，支持逗号分隔多个值
}
```

- [x] 完成

### 1.2 生成 Protobuf 代码

| 项目 | 内容 |
|------|------|
| Command | `cd ttpos-bmp/app/ttpos-erp && make pb` |
| Purpose | 根据 proto 文件生成 Go 代码 |
| Output | `api/buying/buying.pb.go`, `api/buying/buying_grpc.pb.go` |

- [x] 完成

### 1.3 修改过滤逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go` |
| Function | `buildPurchaseOrderListFilters` |
| Purpose | 添加 name 字段的 IN 查询支持 |
| Requirements | REQ1: 多值使用 IN 查询<br/>REQ2: 单值保持等值查询 |
| Leverage | 复用现有过滤模式 |

**具体改动**:
```go
// 在 buildPurchaseOrderListFilters 方法中添加
if len(req.Name) > 0 {
    if strings.Contains(req.Name, ",") {
        filters = append(filters, g.ArrayStr{"name", "in", req.Name})
    } else {
        filters = append(filters, g.ArrayStr{"name", "=", req.Name})
    }
}
```

- [x] 完成

### 1.4 验证测试

| 项目 | 内容 |
|------|------|
| Command | `cd ttpos-bmp/app/ttpos-erp && go build ./...` |
| Purpose | 验证代码编译通过 |

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] 编译通过: `go build ./...`

### 功能完整性
- [x] 多 name 查询（逗号分隔）使用 IN 查询
- [x] 单 name 查询保持原有行为
- [x] 空 name 不应用过滤条件

### BMP 模块特殊检查
- [x] `make pb` 代码生成成功
- [x] 不修改自动生成的文件（dao/entity/do/service）

---

## 验收标准追踪

| AC | 状态 | 验证方式 |
|----|------|---------|
| WHEN name 含逗号 THEN 使用 IN 查询 | ✅ 通过 | 代码审查 |
| WHEN name 单值 THEN 保持等值查询 | ✅ 通过 | 代码审查 |
| WHEN name 为空 THEN 不应用过滤 | ✅ 通过 | 代码审查 |

---

**版本**: v1.0.0
**创建日期**: 2026-01-29
**完成日期**: 2026-01-29
