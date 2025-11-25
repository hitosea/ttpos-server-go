# API 响应格式示例

> Go Main 模块中 API 响应格式的正确和错误示例

---

## 规范说明

**data 必须是对象，不能是 null 或数组**

---

## ✅ 正确示例

```go
// ✅ 正确
helper.Success(c, gin.H{})
helper.Success(c, gin.H{"list": []})
```

---

## ❌ 错误示例

```go
// ❌ 错误
helper.Success(c, nil)
helper.Success(c, []string{})
```

**问题：**
- `data` 返回 `null` 会导致前端无法统一处理
- `data` 返回数组不符合项目规范，必须是对象

---

## 切片初始化规范

**响应体中的切片必须使用 `make` 初始化，避免返回 null**

### ✅ 正确示例

```go
// ✅ 正确：使用 make 初始化切片
type ExampleResp struct {
    List     []Item   `json:"list"`      // 切片字段
    Tags     []string `json:"tags"`      // 字符串切片
    Ids      []uint64 `json:"ids"`      // 数字切片
}

func GetExample() *ExampleResp {
    resp := &ExampleResp{
        List: make([]Item, 0),      // ✅ 初始化为空切片，JSON 序列化为 []
        Tags: make([]string, 0),     // ✅ 初始化为空切片，JSON 序列化为 []
        Ids:  make([]uint64, 0),    // ✅ 初始化为空切片，JSON 序列化为 []
    }
    return resp
}

// JSON 输出：
// {
//   "list": [],
//   "tags": [],
//   "ids": []
// }
```

### ❌ 错误示例

```go
// ❌ 错误：未初始化的切片会序列化为 null
type ExampleResp struct {
    List []Item `json:"list"`
}

func GetExample() *ExampleResp {
    resp := &ExampleResp{
        List: nil,  // ❌ nil 切片，JSON 序列化为 null
    }
    return resp
}

// JSON 输出：
// {
//   "list": null  // ❌ 错误：应该是 []
// }
```

### 常见场景

**1. 列表查询接口**

```go
// ✅ 正确
func (h *OrderHandler) GetList(c *gin.Context) {
    ctx := helper.GetContext(c)
    
    result, err := h.orderSrv.GetList(ctx, &req.OrderListReq{})
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }
    
    // 确保 List 字段已初始化
    if result.List == nil {
        result.List = make([]OrderItem, 0)
    }
    
    helper.Success(c, result, "获取成功")
}
```

**2. 响应结构体定义**

```go
// ✅ 正确：在结构体定义时就初始化
type OrderListResp struct {
    List []OrderItem `json:"list"`
    Meta OrderMeta   `json:"meta"`
}

func NewOrderListResp() *OrderListResp {
    return &OrderListResp{
        List: make([]OrderItem, 0),  // ✅ 初始化
        Meta: OrderMeta{},
    }
}

// ❌ 错误：未初始化
func NewOrderListResp() *OrderListResp {
    return &OrderListResp{
        List: nil,  // ❌ 错误
    }
}
```

**3. 条件赋值**

```go
// ✅ 正确：即使有条件判断，也要初始化
func GetOrderItems(orderId uint64) []OrderItem {
    items, err := orderRepo.GetItems(orderId)
    if err != nil {
        return make([]OrderItem, 0)  // ✅ 错误时返回空切片，不是 nil
    }
    
    if len(items) == 0 {
        return make([]OrderItem, 0)  // ✅ 空结果返回空切片
    }
    
    return items
}
```

### 规则总结

- ✅ **必须使用 `make([]Type, 0)` 初始化所有响应体中的切片字段**
- ✅ **即使切片为空，也要初始化为空切片，不要使用 nil**
- ✅ **在结构体初始化时就进行切片初始化**
- ✅ **函数返回切片时，确保返回的是空切片而不是 nil**
- ❌ **禁止在响应体中使用 nil 切片**
- ❌ **禁止依赖 Go 的零值初始化（nil slice）**

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - API 响应格式规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

