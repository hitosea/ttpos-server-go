# ERP UpdateProduct 增加 UOM 字段更新支持 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目         | 内容                           |
| ------------ | ------------------------------ |
| **提案人**   | rikugun                        |
| **日期**     | 2025-11-27                     |
| **目标版本** | 待定                           |
| **状态**     | 已创建 Spec                    |
| **关联任务** | -                              |
| **关联 Spec** | [task-erp-update-product-uom](../../shared/specs/archived/v2.12/task-erp-update-product-uom/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-erp` 模块的 `UpdateProduct` 接口仅支持更新以下字段：
- `not_for_sale` - 是否禁售
- `internal_code` - 内部编码
- `disabled` - 是否禁用
- `attributes` - 规格值

**无法更新商品的计量单位（UOM）字段**，如 `stock_uom`（库存单位）。

当商户需要修改商品的计量单位时（例如从"个"改为"份"），目前无法通过 API 完成，需要直接在 ERPNext 后台操作，增加了运维成本和出错风险。

### 业务价值

- 支持商户灵活调整商品计量单位
- 减少后台运维操作，降低人工干预风险
- 完善商品管理 API 功能，提升系统自动化程度
- 与 ERPNext 的 Item 数据模型保持一致

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 系统运维人员、ERP 管理员

---

## 💡 解决方案概述

### 方案描述

在 `UpdateProductReq` 消息中增加 `stock_uom` 字段，允许调用方传入新的计量单位。业务逻辑层在更新商品时，将 `stock_uom` 映射到 ERPNext 的 Item 文档对应字段进行更新。

### 核心功能点

1. **Protobuf 定义扩展**：在 `UpdateProductReq` 和 `UpdateProductResp` 中增加 `stock_uom` 字段
2. **业务逻辑更新**：修改 `sProduct.UpdateProduct` 方法，支持 `stock_uom` 字段的更新
3. **参数验证**：验证传入的 UOM 是否有效（可选，根据业务需要）

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯接口扩展，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. ERPNext 对 `stock_uom` 字段更新可能有业务约束（如已有库存记录时不允许修改）
2. UOM 字段值需要与 ERPNext 系统中已定义的 UOM 匹配

**缓解措施**：
1. 查阅 ERPNext 文档确认 `stock_uom` 更新的业务规则
2. 调用前验证 UOM 是否存在于系统中（可选）

---

## 🔗 相关资源

### 参考需求

- 现有 UpdateProduct 接口: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto`
- ERPNext Item DocType: https://github.com/frappe/erpnext/blob/develop/erpnext/stock/doctype/item/item.json

### 相关文档

- Item DTO 定义: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go`
- UpdateProduct 逻辑实现: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |
| 测试代表     |        |           |
| UI/UX 设计师 | N/A    |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`task-erp-update-product-uom`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 通过 API 更新商品的计量单位（UOM）  
**以便于** 灵活管理商品属性，无需登录 ERPNext 后台操作

### AC 验收标准（初稿）

1. **WHEN** 调用 `UpdateProduct` 接口并传入有效的 `stock_uom` 值 **THEN** 系统 **SHALL** 成功更新商品的计量单位
2. **WHEN** 调用 `UpdateProduct` 接口不传入 `stock_uom` 字段 **THEN** 系统 **SHALL** 不修改商品的计量单位（保持现有行为）
3. **IF** 传入的 `stock_uom` 在 ERPNext 中不存在 **THEN** 系统 **SHALL** 返回明确的错误信息（可选）

### 技术实现要点

1. **Protobuf 修改** (`product.proto`):

```protobuf
message UpdateProductReq {
  string item_code = 1;
  bool not_for_sale = 2;
  string internal_code = 3;
  bool disabled = 4;
  repeated ProductAttribute attributes = 5;
  string stock_uom = 6; // 新增：库存单位
}

message UpdateProductResp {
  string item_code = 1;
  bool not_for_sale = 2;
  string internal_code = 3;
  bool disabled = 4;
  repeated ProductAttribute attributes = 5;
  string stock_uom = 6; // 新增：库存单位
}
```

2. **Logic 层修改** (`product.go`):

```go
func (s *sProduct) UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*item.UpdateProductResp, error) {
    itemInfo := &erp.Item{
        CustomNotForSale: req.NotForSale,
        Disabled:         req.Disabled,
    }
    
    // 新增：处理 stock_uom
    if len(req.StockUom) > 0 {
        itemInfo.StockUom = req.StockUom
    }
    
    // ... 其余逻辑不变
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: rikugun

