# task-erp-material-request-sales-order 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 1 |
| 总任务数 | 3 |
| 已完成 | 3 |
| 完成率 | 100% |

---

## Phase 1: 实现

### 1.1 修改 Protobuf 定义

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto` |
| Purpose | 在 SaveMaterialRequestResp 中新增 sales_order 字段 |
| Requirements | 字段编号为 3，类型为 string |
| Leverage | 参考现有 purchase_order 字段定义 |

**变更内容**:
```protobuf
message SaveMaterialRequestResp {
  string material_request_name = 1;      // 物品申请单名称
  string purchase_order = 2;             // 采购订单号
  string sales_order = 3;                // 内部销售订单号
}
```

- [x] 完成

---

### 1.2 生成 Protobuf 代码

| 项目 | 内容 |
|------|------|
| Command | `cd ttpos-bmp/app/ttpos-erp && make pb` |
| Purpose | 重新生成 Go 代码，包含新的 SalesOrder 字段 |
| Output | `ttpos-bmp/app/ttpos-erp/api/stock/stock.pb.go` |

- [x] 完成

---

### 1.3 修改 Controller 代码

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go` |
| Purpose | 接收 CreateInnerSaleOrderFromPurchaseOrder 返回值并赋给 resp.SalesOrder |
| Requirements | 仅在采购类型 (purpose == erp.StockEntryTypePurchase) 时设置 |
| Leverage | 使用 saleOrder.Name 获取订单名称 |

**变更位置**: 第 75-87 行

**Before**:
```go
if _, err = service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
    SourceName:      purchaseOrder.Name,
    DeliveryDate:    requiredBy,
    SourceWarehouse: req.SourceWarehouse,
}); err != nil {
    // 错误处理...
}
```

**After**:
```go
saleOrder, err := service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
    SourceName:      purchaseOrder.Name,
    DeliveryDate:    requiredBy,
    SourceWarehouse: req.SourceWarehouse,
})
if err != nil {
    // 错误处理（保持不变）...
}
resp.SalesOrder = saleOrder.Name
```

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `make pb` 执行成功
- [x] `make build` 编译通过
- [x] 代码格式化完成

### 功能完整性
- [x] 采购类型请求返回 sales_order 字段
- [x] 非采购类型请求返回空 sales_order
- [x] 现有调用方正常工作（向后兼容）

### 验收标准
- [x] AC1: 成功创建时响应包含 sales_order 字段
- [x] AC2: 失败时返回错误信息
- [x] AC3: 非采购类型返回空 sales_order

---

**版本**: v1.0.0
**创建日期**: 2026-01-26
**完成日期**: 2026-01-26
