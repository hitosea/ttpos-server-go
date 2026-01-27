# SaveMaterialRequestResp 新增 sales_order 字段

## 📋 基本信息

| 项目              | 内容                                                                         |
| ----------------- | ---------------------------------------------------------------------------- |
| **来源 Proposal** | [erp-material-request-sales-order](../../../team/proposals/2026-01/erp-material-request-sales-order.md) |
| **创建日期**      | 2026-01-26                                                                   |
| **负责人**        | rikugun                                                                      |
| **目标版本**      | v2.16                                                                        |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已完成     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-26 |

---

## 📝 用户故事

**作为** 后端服务调用方
**我想** 在调用 SaveMaterialRequest 后获取完整的订单信息（包括 sales_order）
**以便于** 减少 API 调用次数，无需二次查询即可获取关联的内部销售订单号

---

## 功能需求

### Requirement 1: SaveMaterialRequestResp 新增 sales_order 字段

**用户故事**: 作为后端服务调用方，我想在 SaveMaterialRequest 响应中获取 sales_order 字段，以便于无需额外查询即可获取内部销售订单号

#### 验收标准

1. **WHEN** 调用 SaveMaterialRequest 成功创建物品申请单、采购订单和内部销售订单 **THEN** 系统 **SHALL** 在响应中返回 `sales_order` 字段，值为创建的内部销售订单名称
2. **WHEN** 创建内部销售订单失败（整体流程回滚） **THEN** 系统 **SHALL** 返回错误信息，不返回 `sales_order` 字段
3. **WHEN** 请求 purpose 非采购类型（不创建内部销售订单） **THEN** 系统 **SHALL** 返回空的 `sales_order` 字段

---

## 非功能需求

### 测试要求

- [ ] Protobuf 生成代码编译通过
- [ ] gRPC 接口调用测试通过

### 兼容性

- [x] 向后兼容：Protobuf 新增字段不影响现有调用方
- [x] 现有调用方无需修改即可继续工作

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/go-rules.mdc 规范
- 必须使用 `make pb` 重新生成 Protobuf 代码
- **禁止修改自动生成的 dao/entity/do 文件**

### 资源约束

- Story Point: 1

---

## 实现范围

### 需要修改的文件

| 文件路径 | 修改内容 |
|---------|---------|
| `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto` | 在 `SaveMaterialRequestResp` 中新增 `sales_order` 字段 |
| `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go` | 将 `CreateInnerSaleOrderFromPurchaseOrder` 返回值赋给 `resp.SalesOrder` |

### Protobuf 变更

```protobuf
message SaveMaterialRequestResp {
  string material_request_name = 1;      // 物品申请单名称
  string purchase_order = 2;             // 采购订单号
  string sales_order = 3;                // 内部销售订单号（新增）
}
```

### 代码变更

```go
// Before (当前代码)
if _, err = service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{...}); err != nil {
    // 错误处理
}

// After (修改后)
saleOrder, err := service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{...})
if err != nil {
    // 错误处理
}
resp.SalesOrder = saleOrder.Name
```

---

## 风险和缓解

### 风险 1: 现有调用方未处理新字段

**影响**: 低
**缓解措施**: Protobuf 新增字段天然向后兼容，现有调用方会自动忽略未知字段

---

**版本**: v1.0.0
**创建日期**: 2026-01-26
