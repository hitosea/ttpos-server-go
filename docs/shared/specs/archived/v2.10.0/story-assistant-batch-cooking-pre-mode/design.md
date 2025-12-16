# 点餐助手端分批送厨前置模式 设计文档

> 本文档定义点餐助手端分批送厨前置模式的技术设计和实现方案。

## 📋 概述

点餐助手端分批送厨前置模式通过优化购物车签名计算、加购接口和智能送厨逻辑，实现在选购商品时就关联分批类型，下单后自动按优先级送厨。核心实现包括：

- 在业务设置中增加分批送厨模式字段
- 新增分批类型列表接口
- 优化购物车商品签名计算（包含 batch_tag_uuid）
- 加购接口支持 batch_tag_uuid 参数
- 智能送厨逻辑（自动按优先级送厨）

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ Service 依赖 Repository 接口
- ✅ URL 使用 snake_case: `/assistant/batch_tag/list`
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
- **ProductService**: `main/app/service/product.go` - 复用分批类型列表查询（GetBatchTagList 方法）
- **BatchTagRepository**: `main/app/repository/batch_tag.go` - 复用分批类型查询（由 ProductService 内部调用）
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
AssistantHandler (API 层)
  ↓ 调用
ProductService + OrderService (业务层)
  ↓ 依赖
IBatchTagRepo + IOrderRepo (数据层)
```

**依赖规则**：

- ✅ API 层通过 Service 层访问数据，不直接调用 Repository
- ✅ ProductService 提供 GetBatchTagList 方法，供 API 层调用
- ✅ OrderService 依赖 IBatchTagRepo 和 IOrderRepo 接口
- ✅ 使用事务管理保证数据一致性

---

## 🗄️ 数据模型设计

### 响应结构变更

#### 1. AssistantBase 响应结构（新增字段）

**文件**: `main/app/dto/resp/base.go`

```go
type AssistantBase struct {
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

#### 4. BatchTagList 响应结构（复用）

**文件**: `main/app/dto/resp/product_resp/product.go`

**说明**: 复用现有的 `BatchTagList` 和 `BatchTag` 结构，无需新建

```go
type BatchTagList struct {
    List []BatchTag `json:"list"` // 分批类型列表
}

type BatchTag struct {
    Uuid         uint64             `json:"uuid"`         // 分批类型UUID
    LocaleName   dto.LocaleResponse `json:"locale_name"`  // 分批类型名称（多语言）
    Color        string             `json:"color"`        // 颜色
    Sort         int                `json:"sort"`         // 排序
    Abbreviation string            `json:"abbreviation"`  // 缩写
}
```

---

## 🔌 API 设计

### 1. 获取基础信息接口（修改）

**接口**: `GET /assistant/base`

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

### 2. 获取分批类型列表接口（新增）

**接口**: `GET /assistant/batch_tag/list`

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

**排序规则**: 按 `sort` 字段升序排序，优先级高的在前

### 3. 加购商品接口（修改）

**接口**: `POST /assistant/desk/order/cart/product/add`

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

### 3. 智能送厨逻辑

**位置**: `main/app/service/order_action.go` - 修改 `ActionSubmit` 方法

**逻辑流程**:
1. 下单成功后，获取所有预送厨商品
2. 按分批类型分组，找出优先级最高的类型（sort 最小的类型）
3. 自动送厨该类型的所有商品
4. 如果还有预送厨商品，继续送优先级最高的类型
5. 重复此过程，直到没有可送厨商品

**伪代码**:
```go
// 下单后智能送厨
func (s *orderSrv) autoSendCookingByPriority(ctx context.Context, saleBillUuid uint64) error {
    for {
        // 获取所有预送厨商品
        preCookingProducts := getPreCookingProducts(saleBillUuid)
        if len(preCookingProducts) == 0 {
            break
        }
        
        // 按分批类型分组
        typeGroups := groupByBatchTag(preCookingProducts)
        
        // 找出优先级最高的类型（sort 最小的类型）
        highestPriorityType := findHighestPriorityType(typeGroups)
        
        // 送厨该类型的所有商品
        err := s.sendCookingByBatchTag(ctx, saleBillUuid, highestPriorityType)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 4. 分批类型列表查询

**位置**: `main/app/api/v1/assistant/assistant_base.go` - 新增方法 `GetBatchTagList`

**逻辑流程**:
1. 调用 ProductService.GetBatchTagList 方法（通过 Service 层）
2. ProductService 内部调用 BatchTagRepository 查询所有分批类型
3. 按 sort 字段升序排序
4. 转换为响应结构，包含多语言名称、颜色、缩写等信息
5. 返回排序后的列表

**实现说明**:
- 遵循分层架构：API 层 → Service 层 → Repository 层
- 复用 `ProductService.GetBatchTagList` 方法，与收银端保持一致
- 在 `BaseHandler` 中注入 `productSrv service.IProductSrv`

---

## 🔧 关键代码位置

### Service 层

- **加购商品**: `main/app/service/order_product.go` - `InstantOrderCartProductAdd`
- **智能送厨**: `main/app/service/order_action.go` - `ActionSubmit` (修改)
- **获取基础信息**: `main/app/service/auth.go` - `AssistantBase`

### Repository 层

- **分批类型查询**: `main/app/repository/batch_tag.go` - `GetBatchTagList`
- **订单查询**: `main/app/repository/order.go` - `GetSaleBillAllInfo`
- **商品更新**: `main/app/repository/sale_order_product.go` - `UpdateSaleOrderProductList`

### API 层

- **基础信息**: `main/app/api/v1/assistant/assistant_base.go` - `GetBase`
- **分批类型列表**: `main/app/api/v1/assistant/assistant_base.go` - `GetBatchTagList` (新增，通过 ProductService 调用)
- **加购商品**: `main/app/api/v1/assistant/assistant_desk.go` - `OrderCartProductAdd`
- **下单**: `main/app/api/v1/assistant/assistant_desk.go` - `OrderSubmit` (修改)

### Service 层

- **分批类型列表**: `main/app/service/product.go` - `GetBatchTagList` (复用，与收银端一致)

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

3. **智能送厨测试**
   - 测试按优先级正确送厨
   - 测试每次点击下单都送优先级最高的类型
   - 测试没有可送厨商品时正常完成下单

### 集成测试

1. **完整加购流程测试**
   - 前置模式下加购多个商品
   - 验证购物车中商品正确关联类型
   - 验证相同商品不同类型分开显示

2. **智能送厨流程测试**
   - 前置模式下加购多个分批类型商品
   - 点击下单，验证自动按优先级送厨
   - 验证每次点击下单都送优先级最高的类型

---

## 📝 注意事项

1. **向后兼容**: 默认使用后置模式，确保现有功能不受影响
2. **性能优化**: 购物车签名计算需要高效，避免影响加购性能
3. **智能送厨**: 确保送厨逻辑不影响正常下单流程
4. **错误处理**: 所有接口都需要完善的错误处理和日志记录
5. **事务管理**: 智能送厨逻辑需要使用事务保证数据一致性

---

## 🔗 与收银端实现的差异

### 相同点

- 购物车签名计算逻辑相同
- 加购接口支持 batch_tag_uuid 参数
- 业务设置中增加 batch_cooking_mode 字段

### 不同点

- **助手端**: 不需要更换类型功能（根据需求，助手端不支持更换类型）
- **助手端**: 需要智能送厨逻辑（自动按优先级送厨）
- **助手端**: 需要分批类型列表接口
- **收银端**: 需要更换类型接口
- **收银端**: 分批送厨弹窗支持更换类型功能

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: 后端开发组  
**审核者**: 待定

