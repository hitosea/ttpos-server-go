---
description: Go Main 模块编码规范 — 编写或修改 main/ 目录下的 Go 代码时自动加载
globs:
  - "main/**/*.go"
---

# Go Main 编码规范

## 分层架构

API → Service → Repository → DB，**严禁跨层调用和循环依赖**。

- API 层只调用 Service，**禁止引用 repository 包**
- Service 可依赖其他 Service，只能依赖自己领域的 Repository
- Repository 只持有 db 实例，不持有 DBManager

## 命名规范

| 类型 | 规则 | 示例 |
|------|------|------|
| 结构体 | 大驼峰，ID 字段大写 | `StaffId`, `OrderUuid` |
| 接口 | `I` 开头 | `IOrderSrv`, `IUserRepo` |
| 包名/文件名 | snake_case | `member_service.go` |
| 外键 UUID 字段 | 完整表名_uuid | `product_uuid`（非 `prod_uuid`）|

## 关键约束

- 使用 `any` 替代 `interface{}`
- 使用 `return error`，不使用 `panic`
- 协程必须使用 `utils.Go`（内置 recover）
- 多语言字段必须使用 `dto.LocaleResponse`，字段名用 `LocaleName` 或 `LocaleXXXName`
- 常量必须在 `constant` 包中定义
- Service 接口定义（`I{Name}Srv`）和实现（`{Name}Srv`）在同一文件
- 通过 `ctx.GetDB()` 获取数据库连接
- 数据库操作必须通过 Repository 层，禁止直接用 GORM
- **所有日志必须包含 `company_uuid` 字段**

## 事务规范

```go
err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    orderRepo := repository.NewOrderRepo(tx)  // 用 tx 创建 Repository
    // ... 数据库操作
    return nil  // 自动提交
})
// 禁止：手动 tx.Commit()/tx.Rollback()
// 禁止：事务中用 db 而非 tx 创建 Repository
```
