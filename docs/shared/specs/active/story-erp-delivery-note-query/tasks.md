# story-erp-delivery-note-query 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 7 |
| 已完成 | 6 |
| 完成率 | 86% |

---

## Phase 1: Proto 定义与代码生成

### 1.1 新建 delivery_note.proto

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/delivery_note_pb/delivery_note.proto` |
| Purpose | 定义独立的 DeliveryNoteService 和相关 message |
| Requirements | 需求 R1 - 送货单列表查询接口 |
| Leverage | 参考 `selling/selling.proto` 结构 |

**实现要点**：
1. 添加 `DeliveryNote` message（送货单信息）
2. 添加 `DeliveryNoteItem` message（送货单明细项）
3. 添加 `GetDeliveryNoteListReq` message（包含 po_no 字段）
4. 添加 `GetDeliveryNoteListResp` message
5. 定义独立的 `DeliveryNoteService` 服务
6. `go_package` 设为 `ttpos-bmp/app/ttpos-erp/api/delivery_note_pb`

**实现说明**：
- 为避免与 `api/selling/` 中现有 `selling` 包冲突，proto 文件放置在独立目录 `delivery_note_pb/`
- 生成的 Go 代码位于 `api/delivery_note_pb/`

- [x] 完成

### 1.2 执行 protobuf 代码生成

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/api/delivery_note_pb/` |
| Purpose | 生成 Go 代码 |
| Requirements | - |
| Leverage | - |

**执行命令**：
```bash
cd ttpos-bmp/app/ttpos-erp && make pb
```

- [x] 完成

---

## Phase 2: Logic 迁移与 DTO 扩展

### 2.1 迁移 Logic 到 selling 目录

| 项目 | 内容 |
|------|------|
| From | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go` |
| To | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/delivery_note.go` |
| Purpose | 将 DeliveryNote Logic 迁移到 selling 领域 |
| Requirements | 保持现有功能不变 |

**实现要点**：
1. 将文件从 `logic/stock/` 移动到 `logic/selling/`
2. 修改 package 名称为 `selling`
3. 更新 `internal/service/delivery_note.go` 中的注册引用
4. 确保原有调用方正常工作（如有）

**实现说明**：
- 新建文件在 `internal/logic/selling/delivery_note.go`
- 包含完整的 po_no 过滤逻辑实现
- 删除原 `logic/stock/delivery_note.go` 文件

- [x] 完成

### 2.2 扩展 GetDeliveryNoteListReq DTO

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/api/delivery_note/delivery_note.go` |
| Purpose | 添加 PoNo 字段支持采购订单号查询 |
| Requirements | 支持 po_no 查询 |
| Leverage | - |

**实现要点**：
```go
type GetDeliveryNoteListReq struct {
    // ... 现有字段 ...
    PoNo string `json:"po_no"` // 采购订单号，可选
}
```

- [x] 完成

### 2.3 扩展 GetDeliveryNoteList Logic

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/delivery_note.go` |
| Purpose | 支持 po_no 过滤逻辑 |
| Requirements | 通过 PO → SO → DN 链路查询 |
| Leverage | 现有 GetDeliveryNoteList 方法 |

**实现要点**：
1. 如果 `req.PoNo` 不为空，先查询关联的销售订单
2. 将销售订单号加入过滤条件（`against_sales_order`）
3. 可能需要通过子查询或 DeliveryNoteItem 表关联

**实现说明**：
- 新增 `getSalesOrdersByPoNo` 方法：通过 po_no 查询关联的销售订单
- 新增 `filterByPoNo` 方法：在应用层按 po_no 过滤送货单
- 查询链路：PO(po_no) -> Sales Order(name) -> Delivery Note Item(against_sales_order)

- [x] 完成

---

## Phase 3: Controller 实现与注册

### 3.1 新建 DeliveryNote Controller

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/delivery_note/delivery_note.go` |
| Purpose | 暴露独立的 gRPC 服务 |
| Requirements | 需求 R1 验收标准 1-3 |
| Leverage | 参考 `rpc/stock/stock.go` 实现模式 |

**实现要点**：
1. 定义 `Controller` 结构体，嵌入 `UnimplementedDeliveryNoteServiceServer`
2. 实现 `Register(s *grpcx.GrpcServer)` 函数
3. 实现 `GetDeliveryNoteList` 方法：
   - 参数验证（company_abbr 必填）
   - 转换 proto 请求到 DTO
   - 调用 `service.DeliveryNote().GetDeliveryNoteList()`
   - 包装返回 `rpc.ApiSuccessWithData()`

- [x] 完成

### 3.2 注册 DeliveryNoteService

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/boot/rpc.go` |
| Purpose | 在 gRPC 服务器启动时注册 DeliveryNoteService |
| Requirements | - |
| Leverage | 参考现有 Service 注册方式 |

**实现要点**：
```go
import "ttpos-bmp/app/ttpos-erp/internal/controller/rpc/delivery_note"

// 在 gRPC 服务注册处添加
delivery_note.Register(service.RpcServer.GRpc)
```

**实现说明**：
- 注册位置在 `internal/boot/rpc.go` 的 `initRpcServer()` 函数中
- 与其他服务（stock, selling 等）并列注册

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [x] `make pb` 代码生成完成

### 功能完整性
- [x] GetDeliveryNoteList gRPC 接口可调用
- [x] 返回符合筛选条件的送货单分页列表
- [x] include_items=true 时包含商品明细
- [x] 支持 po_no 查询
- [x] 参数缺失时返回明确错误信息

### 测试验证
- [ ] 单元测试覆盖率 >= 80%
- [ ] gRPC 接口集成测试通过

---

## 实现文件清单

| 操作 | 文件路径 |
|------|----------|
| 新建 | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/delivery_note_pb/delivery_note.proto` |
| 生成 | `ttpos-bmp/app/ttpos-erp/api/delivery_note_pb/delivery_note.pb.go` |
| 生成 | `ttpos-bmp/app/ttpos-erp/api/delivery_note_pb/delivery_note_grpc.pb.go` |
| 新建 | `ttpos-bmp/app/ttpos-erp/internal/logic/selling/delivery_note.go` |
| 删除 | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/delivery_note.go` |
| 修改 | `ttpos-bmp/app/ttpos-erp/api/delivery_note/delivery_note.go` |
| 新建 | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/delivery_note/delivery_note.go` |
| 修改 | `ttpos-bmp/app/ttpos-erp/internal/boot/rpc.go` |
