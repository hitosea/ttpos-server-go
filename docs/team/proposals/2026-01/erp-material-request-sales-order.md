# SaveMaterialRequestResp 新增 sales_order 字段

## 📋 提案信息

| 项目          | 内容                  |
| ------------- | --------------------- |
| **提案人**    | rikugun               |
| **日期**      | 2026-01-26            |
| **目标版本**  | v2.16                 |
| **状态**      | 待评审                |
| **关联 Spec** | -                     |

---

## 🎯 背景和动机

### 问题描述

在 `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go` 的 `SaveMaterialRequest` 方法中，调用 `CreateInnerSaleOrderFromPurchaseOrder` 创建内部销售订单后，返回的 `SaleOrder` 对象被丢弃（使用 `_` 忽略返回值），导致调用方无法获取关联的内部销售订单号。

当前代码（第 76-86 行）：
```go
// saleOrder := &erp.SaleOrder{}
if _, err = service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
    SourceName:      purchaseOrder.Name,
    DeliveryDate:    requiredBy,
    SourceWarehouse: req.SourceWarehouse,
}); err != nil {
    // 错误处理...
}
```

当前 `SaveMaterialRequestResp` 结构只返回两个字段：
- `material_request_name`：物品申请单名称
- `purchase_order`：采购订单号

缺少 `sales_order` 字段，导致后端服务调用方需要二次查询才能获取销售订单信息。

### 业务价值

- 提升开发效率：一次请求返回完整信息，避免额外的 API 调用
- 减少网络开销：无需二次查询获取销售订单号
- 保证数据一致性：返回的订单号与创建时一致，避免时序问题

### 目标用户

- [x] 后端服务调用方
- [x] 其他微服务集成方

---

## 💡 解决方案概述

### 方案描述

在 `SaveMaterialRequestResp` 消息中新增 `sales_order` 字段，用于返回 `CreateInnerSaleOrderFromPurchaseOrder` 方法创建的内部销售订单名称。同时修改 `stock.go` 中的代码，将返回的 `saleOrder.Name` 赋值给响应对象。

### 核心功能点

1. 修改 `stock.proto` 文件，在 `SaveMaterialRequestResp` 中新增 `sales_order` 字段
2. 执行 `make pb` 重新生成 Go 代码
3. 修改 `stock.go` 中的 `SaveMaterialRequest` 方法，将 `CreateInnerSaleOrderFromPurchaseOrder` 的返回值赋值给 `resp.SalesOrder`

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [ ] Kiosk 自助点餐机
- [x] 后端服务（ERP/BMP）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：Protobuf 字段新增 + 返回值赋值，无业务逻辑变更
- [ ] **中**：需要前后端联调,基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预估 SP**: 1（待技术评审确认）

### 拆分预估

**是否需要拆分**：
- [x] **否**：单模块，SP ≤ 5，可直接创建 1 个 Spec
- [ ] **是**：需要拆分为多个 Spec

### 风险识别

**潜在风险**：
1. 现有调用方可能未处理新增字段（向后兼容，风险低）

**缓解措施**：
1. Protobuf 新增字段向后兼容，不影响现有调用方

---

## 🤝 需求评审

### 评审参与人

| 角色       | 姓名   | 签名/日期 |
| ---------- | ------ | --------- |
| 产品经理   |        |           |
| 技术负责人 |        |           |
| 开发代表   |        |           |
| 测试代表   |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-erp-material-request-sales-order`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 后端服务调用方
**我想** 在调用 SaveMaterialRequest 后获取完整的订单信息（包括 sales_order）
**以便于** 无需二次查询即可获取关联的内部销售订单号

### AC 验收标准（初稿）

1. **WHEN** 调用 SaveMaterialRequest 成功创建物品申请单和内部销售订单 **THEN** 响应 **SHALL** 包含 `sales_order` 字段，值为创建的内部销售订单名称
2. **WHEN** 创建内部销售订单失败 **THEN** 响应 **SHALL** 不包含 `sales_order` 字段或为空字符串

### 代码变更参考

**stock.proto 变更**：
```protobuf
message SaveMaterialRequestResp {
  string material_request_name = 1;      // 物品申请单名称
  string purchase_order = 2;             // 采购订单号
  string sales_order = 3;                // 内部销售订单号（新增）
}
```

**stock.go 变更**：
```go
saleOrder, err := service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
    SourceName:      purchaseOrder.Name,
    DeliveryDate:    requiredBy,
    SourceWarehouse: req.SourceWarehouse,
})
if err != nil {
    // 错误处理...
}
resp.SalesOrder = saleOrder.Name
```

---

**版本**: v1.0.0
