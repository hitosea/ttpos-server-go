# 常量定义示例

> Go Main 模块中常量定义的正确和错误示例

---

## 规范说明

- ✅ **所有常量必须在 `constant` 包中定义**
- ✅ **禁止在代码中使用字面量（magic numbers/strings）**
- ✅ **使用常量或枚举替代字面量**

---

## ✅ 正确示例

```go
// ✅ 正确 - 在 constant/order.go 中定义
const (
    OrderStatusPending  = 0
    OrderStatusComplete = 1
    OrderStatusCanceled = 2
)

// 在代码中使用
if order.Status == constant.OrderStatusPending {
    // ...
}

// ✅ 正确 - 使用常量
if order.Source == constant.OrderSourceInstant {
    // ...
}
```

---

## ❌ 错误示例

```go
// ❌ 错误 - 使用字面量
if order.Status == 0 {
    // ...
}

// ❌ 错误 - 字符串字面量
if order.Source == "instant" {
    // ...
}
```

**问题：**
- 使用字面量（magic numbers/strings）会导致代码可读性差
- 字面量分散在代码中，难以维护和修改
- 不符合项目规范，必须在 constant 包中定义

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 常量定义规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

