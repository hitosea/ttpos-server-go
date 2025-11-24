# Repository 层示例

> Go Main 模块中 Repository 层开发的正确和错误示例

---

## 规范说明

- ✅ **只能持有 db 实例**
- ❌ **不能持有 DBManager**
- ✅ **使用选项模式**

---

## ✅ 正确示例

```go
type OrderRepoImpl struct {
    db  *gorm.DB           // ✅ 正确
    dbm *database.DBManager // ❌ 错误
}
```

---

## ❌ 错误示例

```go
type OrderRepoImpl struct {
    dbm *database.DBManager // ❌ 错误：Repository 不能持有 DBManager
}
```

**问题：**
- Repository 层只能持有 db 实例
- 持有 DBManager 违反了分层架构原则

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - Repository 层规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

