# 协程使用示例

> Go Main 模块中协程使用的正确和错误示例

---

## 规范说明

- ✅ **所有 go func 协程都必须使用 `utils.Go` 方法**
- ✅ **`utils.Go` 方法已内置 recover 机制，自动捕获 panic 并记录日志**
- ❌ **禁止直接使用 `go func() { ... }()`**
- ❌ **禁止在协程中手动编写 recover（`utils.Go` 已包含）**

---

## ✅ 正确示例

### 示例 1: 使用 utils.Go 启动协程发布事件

```go
// ✅ 正确 - 使用 utils.Go 启动协程
import "ttpos-server-go/pkg/utils"

utils.Go(func() {
    event.NewSystemBus().PublishOrderCreatedEvent(
        event.OrderCreatedPayload{
            BasePayload: event.BasePayload{
                Ctx: ctx,
                CompanyUuid: ctx.GetCompanyUuid(),
            },
            OrderUuid: orderUuid,
        },
    )
})
```

### 示例 2: 异步处理任务

```go
// ✅ 正确 - 异步处理任务
utils.Go(func() {
    // 业务逻辑，即使发生 panic 也不会导致主程序崩溃
    doSomething()
})
```

---

## ❌ 错误示例

### 示例 1: 直接使用 go func()

```go
// ❌ 错误 - 直接使用 go func()
go func() {
    // 如果这里发生 panic，可能导致整个程序崩溃
    doSomething()
}()
```

**问题：**
- 没有 panic 恢复机制
- 协程中的 panic 可能导致整个程序崩溃
- 不符合项目规范

### 示例 2: 手动编写 recover

```go
// ❌ 错误 - 手动编写 recover（utils.Go 已包含）
go func() {
    defer func() {
        if r := recover(); r != nil {
            // 手动处理...
        }
    }()
    doSomething()
}()
```

**问题：**
- `utils.Go` 已经内置了 recover 机制
- 手动编写 recover 是重复代码
- 日志记录可能不一致

---

## 关键点

- **自动化 panic 捕获**：`utils.Go` 自动集成 `defer` + `recover` 机制，无需手动编写
- **统一日志记录**：所有捕获的 panic 都会通过 `logger.Logger.Error` 记录，包含 panic 值和完整的堆栈信息
- **提高健壮性**：确保单个协程的 panic 不会导致整个应用程序崩溃
- **简化开发**：开发者无需在每个协程中手动编写 recover 代码

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 协程使用规范
- [Go Main 开发指南](../../../../docs/human/guides/go-main-development.md) - 详细开发规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

