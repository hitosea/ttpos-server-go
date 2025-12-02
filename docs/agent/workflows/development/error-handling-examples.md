# 错误处理示例

> Go Main 模块中错误处理的正确和错误示例

---

## 规范说明

- ✅ **必须返回 error**，不使用 panic
- ✅ **使用** `errors.WithMessage` 包装错误

---

## ✅ 正确示例

```go
// ✅ 正确
return nil, errors.WithMessage(err, "查询失败")
```

---

## ❌ 错误示例

```go
// ❌ 错误
panic(err)
```

**问题：**
- 使用 `panic` 会导致程序崩溃
- 不符合项目规范，必须返回 error

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 错误处理规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

