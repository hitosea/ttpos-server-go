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

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - API 响应格式规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

