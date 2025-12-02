# API 层示例

> Go Main 模块中 API 层开发的正确和错误示例

---

## 规范说明

- ✅ **只能调用 Service 接口**
- ✅ **只负责参数校验、调用 Service、返回响应**
- ❌ **严禁引用 repository 包**（违反分层架构，必须通过 Service 层）
- ❌ **严禁编写业务逻辑，所有业务逻辑必须在 service 层实现**
- ❌ **禁止在 app/api/v1 目录下的任何子目录中引用 repository 包**

---

## ✅ 正确示例

```go
// ✅ 正确 - API 层只负责调用 Service
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req dto.CreateOrderReq
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.Error(c, err)
        return
    }

    result, err := h.orderSrv.CreateOrder(c, req)
    if err != nil {
        helper.Error(c, err)
        return
    }

    helper.Success(c, result)
}
```

---

## ❌ 错误示例

### 示例 1: API 层引用 repository

```go
// ❌ 错误 - API 层引用 repository
import "app/repository"  // ❌ 禁止
```

**问题：** API 层严禁引用 repository 包，必须通过 Service 层

### 示例 2: API 层编写业务逻辑

```go
// ❌ 错误 - API 层编写业务逻辑
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // ❌ 不应该在这里写业务逻辑
    order := &model.Order{...}
    db.Create(order)  // ❌ 错误
    // 应该调用 h.orderSrv.CreateOrder()
}
```

**问题：** API 层严禁编写业务逻辑，所有业务逻辑必须在 service 层实现

---

## API 层请求参数解析规范

**在 `app/api/v1` 目录下的 handler 中，必须根据 HTTP 方法使用对应的参数解析方式：**

### GET 请求

- ✅ **必须使用 `ShouldBindQuery` 解析查询参数**
- ✅ **Req 结构体字段必须使用 `form` tag**
- ✅ **错误处理使用 `helper.HandleValidationError(c, err, params, nil)`**

```go
// ✅ 正确 - GET 请求参数解析
func (h *OrderHandler) GetPaymentQrcode(c *gin.Context) {
    ctx := helper.NewContext(c)
    
    params := req.InstantOrderPaymentQrcodeReq{}
    if err := c.ShouldBindQuery(&params); err != nil {
        helper.HandleValidationError(c, err, params, nil)
        return
    }
    
    // 调用 Service
    res, err := h.orderSrv.InstantOrderPaymentQrcode(ctx, params)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, res)
}

// Req 结构体定义（在 dto/req 包中）
type InstantOrderPaymentQrcodeReq struct {
    SaleBillUuid      uint64  `form:"sale_bill_uuid" json:"sale_bill_uuid"`           // ✅ 使用 form tag
    SaleOrderUuid     uint64  `form:"sale_order_uuid" json:"sale_order_uuid"`       // ✅ 使用 form tag
    PaymentMethodUuid uint64  `form:"payment_method_uuid" json:"payment_method_uuid"` // ✅ 使用 form tag
    PaymentAmount     float64 `form:"payment_amount" json:"payment_amount"`           // ✅ 使用 form tag
}
```

### POST 请求

- ✅ **必须使用 `ShouldBindJSON` 解析 JSON 请求体**
- ✅ **Req 结构体字段必须使用 `json` tag**
- ✅ **错误处理使用 `helper.HandleValidationError(c, err, params, nil)`**

```go
// ✅ 正确 - POST 请求参数解析
func (h *OrderHandler) SetPaymentZeroRule(c *gin.Context) {
    ctx := helper.NewContext(c)
    
    params := req.InstantOrderPaymentZeroRuleReq{}
    if err := c.ShouldBindJSON(&params); err != nil {
        helper.HandleValidationError(c, err, params, nil)
        return
    }
    
    // 调用 Service
    res, err := h.orderSrv.SetPaymentZeroRule(ctx, params)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, res)
}

// Req 结构体定义（在 dto/req 包中）
type InstantOrderPaymentZeroRuleReq struct {
    SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // ✅ 使用 json tag
    SaleOrderUuid uint64 `json:"sale_order_uuid"` // ✅ 使用 json tag
    ZeroRule      int    `json:"zero_rule"`       // ✅ 使用 json tag
}
```

### DELETE 请求

- ✅ **必须使用 `ShouldBindQuery` 解析查询参数**
- ✅ **Req 结构体字段必须使用 `form` tag**
- ✅ **错误处理使用 `helper.HandleValidationError(c, err, params, nil)`**

```go
// ✅ 正确 - DELETE 请求参数解析
func (h *OrderHandler) DeleteSaleOrder(c *gin.Context) {
    ctx := helper.NewContext(c)
    
    params := req.InstantOrderSaleOrderDeleteReq{}
    if err := c.ShouldBindQuery(&params); err != nil {
        helper.HandleValidationError(c, err, params, nil)
        return
    }
    
    // 调用 Service
    res, err := h.orderSrv.InstantOrderSaleOrderDelete(ctx, params)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
        return
    }
    
    helper.Success(c, res)
}

// Req 结构体定义（在 dto/req 包中）
type InstantOrderSaleOrderDeleteReq struct {
    SaleBillUuid  uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"`  // ✅ 使用 form tag
    SaleOrderUuid uint64 `form:"sale_order_uuid" json:"sale_order_uuid"` // ✅ 使用 form tag
}
```

**总结：**

| HTTP 方法 | 解析方法          | Req 结构体 tag | 错误处理                          |
| --------- | ----------------- | -------------- | --------------------------------- |
| GET       | `ShouldBindQuery` | `form`         | `helper.HandleValidationError`    |
| POST      | `ShouldBindJSON`  | `json`         | `helper.HandleValidationError`    |
| DELETE    | `ShouldBindQuery` | `form`         | `helper.HandleValidationError`    |

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - API 层规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

