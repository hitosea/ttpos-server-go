# 收银机分批送厨前置关联模式 设计文档

> 本文档定义收银机分批送厨前置关联模式的技术设计和实现方案。

## 📋 概述

收银机分批送厨前置关联模式通过优化购物车签名计算、加购接口和分批送厨弹窗，实现在选购商品时就关联分批类型。核心实现包括：

- 在业务设置中增加分批送厨模式字段
- 优化购物车商品签名计算（包含 batch_tag_uuid）
- 加购接口支持 batch_tag_uuid 参数
- 分批送厨弹窗支持更换类型功能
- 新增更换类型接口

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ Service 依赖 Repository 接口
- ✅ URL 使用 snake_case: `/cashier/desk/order/cart/batch/change_tag`
- ✅ data 字段返回对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 errors.WithMessage 包装错误

### API 设计规范 (api.mdc)

- ✅ 响应格式统一: `{code, message, data{}}`
- ✅ data 不能为 null 或数组

---

## 🔄 代码复用分析

### 可复用的现有组件

- **OrderService**: `main/app/service/order.go` - 复用订单相关逻辑
- **OrderCookingService**: `main/app/service/order_cooking.go` - 复用分批送厨逻辑
- **BatchTagRepository**: `main/app/repository/batch_tag.go` - 复用分批类型查询
- **SettingService**: `main/app/service/setting/setting.go` - 复用业务设置查询

### 集成点

- **订单模块**: 调用 OrderService 处理加购和送厨
- **分批类型模块**: 调用 BatchTagRepository 查询分批类型
- **业务设置模块**: 调用 SettingService 获取分批送厨模式
- **购物车模块**: 优化签名计算逻辑

---

## 🏗️ 架构设计

### 分层设计

```
CashierDeskHandler (API 层)
  ↓ 调用
OrderService (业务层)
  ↓ 依赖
IBatchTagRepo + IOrderRepo (数据层)
```

**依赖规则**：

- ✅ OrderService 依赖 IBatchTagRepo 和 IOrderRepo 接口
- ✅ 使用事务管理保证数据一致性

---

## 🗄️ 数据模型设计

### 响应结构变更

#### 1. CashierBase 响应结构（新增字段）

**文件**: `main/app/dto/resp/base.go`

```go
type CashierBase struct {
    // ... 现有字段 ...
    Business   setting.Business    `json:"business"`   // 门店业务设置
    // ... 其他字段 ...
}
```

#### 2. Business 业务设置（新增字段）

**文件**: `main/app/dto/resp/setting/business_setting.go`

```go
type Business struct {
    // ... 现有字段 ...
    BatchCookingMode string `json:"batch_cooking_mode"` // 分批送厨模式: "pre" 前置 / "post" 后置，默认 "post"
}
```

#### 3. OrderCartProductAddReq 请求结构（新增字段）

**文件**: `main/app/dto/req/instant.go`

```go
type InstantOrderAddProductReq struct {
    SaleBillUuid  uint64     `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
    SaleOrderUuid uint64     `json:"sale_order_uuid"` // 销售订单UUID, 必填
    Product       AddProduct `json:"product"`         // 商品, 必填
    BatchTagUuid  uint64     `json:"batch_tag_uuid"`  // 分批类型UUID, 可选（前置模式时使用）
}
```

#### 4. ChangeBatchTagReq 请求结构（新增）

**文件**: `main/app/dto/req/instant.go`

```go
type ChangeBatchTagReq struct {
    SaleBillUuid          uint64   `json:"sale_bill_uuid" binding:"required"`           // 销售账单UUID
    SaleOrderProductUuids []uint64 `json:"sale_order_product_uuids" binding:"required"` // 销售订单商品UUID列表
    BatchTagUuid          uint64   `json:"batch_tag_uuid" binding:"required"`          // 分批类型UUID
}
```

#### 5. BatchTagListResp 响应结构（新增或修改）

**文件**: `main/app/dto/resp/shop_cart.go` 或新建文件

```go
type BatchTagListResp struct {
    List []BatchTagItem `json:"list"` // 分批类型列表
}

type BatchTagItem struct {
    Uuid        uint64             `json:"uuid"`         // 分批类型UUID
    LocaleName  dto.LocaleResponse `json:"locale_name"`  // 分批类型名称（多语言）
    Color       string             `json:"color"`        // 颜色
    Sort        int                `json:"sort"`        // 排序
    Abbreviation string            `json:"abbreviation"` // 缩写
}
```

---

## 🔌 API 设计

### 1. 获取基础信息接口（修改）

**接口**: `GET /cashier/base`

**响应变更**: 在 `Business` 对象中增加 `batch_cooking_mode` 字段

**响应示例**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "business": {
      "batch_cooking_mode": "pre"
    }
  }
}
```

### 2. 加购商品接口（修改）

**接口**: `POST /cashier/desk/order/cart/product/add`

**请求变更**: 在 `InstantOrderAddProductReq` 中增加 `batch_tag_uuid` 字段（可选）

**请求示例**:
```json
{
  "sale_bill_uuid": 123,
  "sale_order_uuid": 456,
  "product": {
    "uuid": 789,
    "flavor_uuid": 101,
    "sauce_uuids": [111, 112],
    "attributes": []
  },
  "batch_tag_uuid": 999
}
```

### 3. 更换分批类型接口（新增）

**接口**: `POST /cashier/desk/order/cart/batch/change_tag`

**请求参数**:
```json
{
  "sale_bill_uuid": 123,
  "sale_order_product_uuids": [111, 112, 113],
  "batch_tag_uuid": 999
}
```

**响应**: 返回更新后的购物车信息

**响应示例**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "sale_bill_uuid": 123,
    "sale_orders": [...]
  }
}
```

### 4. 获取分批类型列表接口（新增或修改）

**接口**: `GET /cashier/desk/batch_tag/list`

**响应示例**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 999,
        "locale_name": {
          "zh": "主食",
          "en": "Main Course"
        },
        "color": "#FF0000",
        "sort": 1,
        "abbreviation": "ZS"
      }
    ]
  }
}
```

### 5. 获取分批送厨弹窗列表接口（修改）

**接口**: `POST /cashier/desk/order/cart/batch/cooking`

**响应变更**: 分批类型列表按 `sort` 排序，优先级高的在前

---

## 💻 核心实现逻辑

### 1. 购物车商品签名计算

**位置**: `main/app/service/order_product.go`

**逻辑**:
```go
// 计算商品签名，当前置模式时包含 batch_tag_uuid
func calculateProductSignature(product *AddProduct, batchTagUuid uint64, isPreMode bool) string {
    signature := fmt.Sprintf("%d_%d_%v_%v", 
        product.Uuid, 
        product.FlavorUuid, 
        product.SauceUuids, 
        product.Attributes)
    
    // 前置模式时，签名包含 batch_tag_uuid
    if isPreMode && batchTagUuid > 0 {
        signature += fmt.Sprintf("_%d", batchTagUuid)
    }
    
    return signature
}
```

### 2. 加购商品逻辑

**位置**: `main/app/service/order_product.go` - `InstantOrderCartProductAdd`

**逻辑流程**:
1. 获取业务设置，判断是否为前置模式
2. 如果是前置模式，从请求中获取 `batch_tag_uuid`（如果未提供，使用默认类型）
3. 验证 `batch_tag_uuid` 的有效性
4. 计算商品签名（包含 `batch_tag_uuid`）
5. 查找购物车中是否有相同签名的商品
6. 如果有，合并数量；如果没有，新增商品
7. 保存商品时关联 `batch_tag_uuid`

### 3. 更换类型逻辑

**位置**: `main/app/service/order_product.go` - 新增方法 `ChangeBatchTag`

**逻辑流程**:
1. 验证请求参数
2. 获取销售账单信息
3. 获取要更换的商品列表
4. 验证商品是否已送厨（已送厨则不允许修改）
5. 验证新的 `batch_tag_uuid` 的有效性
6. 更新商品的分批类型关联
7. 如果订单有关联的点餐助手端，同步更新
8. 返回更新后的购物车信息

### 4. 分批送厨弹窗逻辑

**位置**: `main/app/service/order_cooking.go` - `GetOrderCartProductBatchCookingList`

**逻辑变更**:
1. 获取分批类型列表时，按 `sort` 排序，优先级高的在前
2. 返回的商品信息中包含 `batch_tag_uuid` 和类型颜色
3. 前端根据 `batch_tag_uuid` 显示对应的类型边框颜色

---

## 🔧 关键代码位置

### Service 层

- **加购商品**: `main/app/service/order_product.go` - `InstantOrderCartProductAdd`
- **更换类型**: `main/app/service/order_product.go` - `ChangeBatchTag` (新增)
- **分批送厨弹窗**: `main/app/service/order_cooking.go` - `GetOrderCartProductBatchCookingList`
- **获取基础信息**: `main/app/service/auth.go` - `CashierBase`

### Repository 层

- **分批类型查询**: `main/app/repository/batch_tag.go` - `GetBatchTagList`
- **订单查询**: `main/app/repository/order.go` - `GetSaleBillAllInfo`
- **商品更新**: `main/app/repository/sale_order_product.go` - `UpdateSaleOrderProductList`

### API 层

- **基础信息**: `main/app/api/v1/cashier/cashier_base.go` - `GetBase`
- **加购商品**: `main/app/api/v1/cashier/cashier_desk.go` - `InstantOrderCartProductAdd`
- **更换类型**: `main/app/api/v1/cashier/cashier_desk.go` - `ChangeBatchTag` (新增)
- **分批送厨弹窗**: `main/app/api/v1/cashier/cashier_desk.go` - `GetOrderCartProductBatchCookingList`

---

## 🧪 测试策略

### 单元测试

1. **购物车签名计算测试**
   - 测试相同商品不同分批类型能正确分开
   - 测试相同商品相同分批类型能正确合并
   - 测试后置模式下签名计算不受影响

2. **加购商品测试**
   - 测试前置模式下加购商品关联分批类型
   - 测试未提供 batch_tag_uuid 时使用默认类型
   - 测试无效的 batch_tag_uuid 返回错误

3. **更换类型测试**
   - 测试未送厨商品可以更换类型
   - 测试已送厨商品不允许更换类型
   - 测试更换类型后购物车正确更新

### 集成测试

1. **完整加购流程测试**
   - 前置模式下加购多个商品
   - 验证购物车中商品正确关联类型
   - 验证相同商品不同类型分开显示

2. **分批送厨流程测试**
   - 前置模式下加购商品
   - 点击送厨，验证弹窗显示正确
   - 更换类型后送厨，验证正确性

---

## 📝 注意事项

1. **向后兼容**: 默认使用后置模式，确保现有功能不受影响
2. **数据一致性**: 更换类型操作需要同步到点餐助手端
3. **性能优化**: 购物车签名计算需要高效，避免影响加购性能
4. **错误处理**: 所有接口都需要完善的错误处理和日志记录

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: 后端开发组  
**审核者**: 待定

