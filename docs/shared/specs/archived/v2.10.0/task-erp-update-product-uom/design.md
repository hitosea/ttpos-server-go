# UpdateProduct 增加 UOM 字段支持 设计文档

> 本文档定义 ERP UpdateProduct 接口增加 UOM 字段更新支持的技术设计和实现方案。

## 📋 概述

扩展 `ttpos-erp` 模块的 `UpdateProduct` gRPC 接口，在请求和响应消息中增加 `stock_uom` 字段，并在业务逻辑层处理该字段的更新。

**技术栈**: Go (ttpos-bmp/) - GoFrame 2.x

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- ✅ 禁止修改 dao/entity/do/ 目录（本任务无需修改）
- ✅ 修改 protobuf 后执行 `gf gen pb` 重新生成 API 代码
- ✅ Logic 层实现业务逻辑

### Protobuf 规范 (proto-rules.mdc)

- ✅ 字段名使用 snake_case（`stock_uom`）
- ✅ 字段编号递增（使用 6）
- ✅ 添加中文注释

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Item DTO**: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go` - 已包含 `StockUom` 字段
- **Document Service**: `ttpos-bmp/app/ttpos-erp/internal/service/document.go` - 通用文档更新服务
- **UpdateProduct Logic**: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go` - 现有实现

### 集成点

- **ERPNext Item**: 通过 Document.Update 更新 ERPNext 的 Item 文档

---

## 🏗️ 架构设计

### 分层设计

```
gRPC Controller (product.go)
  ↓ 调用
Logic 层 (sProduct.UpdateProduct)
  ↓ 调用
Document Service (Update)
  ↓ HTTP
ERPNext API
```

### 模块划分

#### Go BMP 模块

- **RPC Controller**: `ttpos-erp/internal/controller/rpc/item/product.go` - gRPC 接口（无需修改）
- **Logic 层**: `ttpos-erp/internal/logic/stock/product.go` - 业务逻辑（需修改）
- **DTO**: `ttpos-erp/internal/model/dto/erp/item.go` - Item DTO（已包含字段）
- **API 定义**: `ttpos-erp/api/item/product.pb.go` - 自动生成（需重新生成）

---

## 🗄️ 数据库设计

**无数据库变更** - 本任务仅涉及 API 接口扩展，数据存储在 ERPNext。

---

## 📊 数据模型

### 现有 Item DTO

```go
// ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/item.go
type Item struct {
    // ... 其他字段
    StockUom string `json:"stock_uom,omitempty"` // 库存计量单位 ✅ 已存在
    // ... 其他字段
}
```

---

## 🔌 API 设计

### gRPC API

#### Protobuf 定义变更

```protobuf
// ttpos-bmp/app/ttpos-erp/manifest/protobuf/item/product.proto

message UpdateProductReq {
  string item_code = 1;                    // 物品编码，必填
  bool not_for_sale = 2;                   // 是否禁售，可选
  string internal_code = 3;                // 内部编码，可选
  bool disabled = 4;                       // 是否禁用，可选
  repeated ProductAttribute attributes = 5; // 更新规格值
  string stock_uom = 6;                    // 库存单位，可选 ✅ 新增
}

message UpdateProductResp {
  string item_code = 1;                    // 物品编码
  bool not_for_sale = 2;                   // 是否禁售
  string internal_code = 3;                // 内部编码
  bool disabled = 4;                       // 是否禁用
  repeated ProductAttribute attributes = 5; // 规格值
  string stock_uom = 6;                    // 库存单位 ✅ 新增
}
```

#### 生成代码

```bash
cd ttpos-bmp/app/ttpos-erp
gf gen pb
```

---

## 🧩 组件和接口

### Logic 层变更

```go
// ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go

func (s *sProduct) UpdateProduct(ctx context.Context, req *item.UpdateProductReq) (*item.UpdateProductResp, error) {
    itemInfo := &erp.Item{
        CustomNotForSale: req.NotForSale,
        Disabled:         req.Disabled,
    }

    // ✅ 新增：处理 stock_uom
    if len(req.StockUom) > 0 {
        itemInfo.StockUom = req.StockUom
    }

    if len(req.Attributes) > 0 {
        attributes := make([]erp.ItemVariantAttribute, 0)
        for _, attr := range req.Attributes {
            attributes = append(attributes, erp.ItemVariantAttribute{
                Attribute:      attr.Attribute,
                AttributeValue: attr.AttributeValue,
            })
        }
        itemInfo.Attributes = attributes
    }

    if len(req.InternalCode) > 0 {
        itemInfo.CustomInternalCode = req.InternalCode
    }

    _, err := service.Document().Update(ctx, &erp.ErpReq{
        DocType: erp.DocTypeItem,
        Name:    req.ItemCode,
    }, itemInfo)
    if err != nil {
        return nil, gerror.Wrapf(err, "更新商品信息失败")
    }

    return &item.UpdateProductResp{
        ItemCode:     req.ItemCode,
        NotForSale:   req.NotForSale,
        InternalCode: req.InternalCode,
        Disabled:     req.Disabled,
        StockUom:     req.StockUom, // ✅ 新增
    }, nil
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: ERPNext 更新失败

- **处理方式**: 包装错误并返回
- **用户影响**: 返回错误信息 "更新商品信息失败"
- **代码示例**:
  ```go
  if err != nil {
      return nil, gerror.Wrapf(err, "更新商品信息失败")
  }
  ```

#### 场景 2: UOM 值在 ERPNext 中不存在

- **处理方式**: ERPNext 会返回错误
- **用户影响**: 返回 ERPNext 的错误信息

---

## 🧪 测试策略

### 测试场景

1. **正常更新**: 传入有效的 `stock_uom` 值，验证更新成功
2. **空值兼容**: 不传入 `stock_uom`，验证不影响其他字段更新
3. **空字符串**: 传入空字符串，验证不修改现有值

### 测试命令

```bash
cd ttpos-bmp/app/ttpos-erp
go test -v ./internal/logic/stock/...
```

---

## 📈 性能优化

**无性能影响** - 仅增加一个字段的映射，无额外开销。

---

## 📚 实现清单

### Phase 1: Protobuf 定义

- [ ] 修改 `product.proto`，增加 `stock_uom` 字段
- [ ] 执行 `gf gen pb` 重新生成代码

### Phase 2: Logic 实现

- [ ] 修改 `sProduct.UpdateProduct` 方法
- [ ] 处理 `req.StockUom` 字段
- [ ] 更新响应返回

### Phase 3: 测试

- [ ] 手动测试接口
- [ ] 验证 ERPNext 数据更新

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-11/2025-11-27.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**作者**: rikugun  
**审核者**: 待定

