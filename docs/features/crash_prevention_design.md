# 异常捕获防止程序崩溃掉线处理方案设计文档

### 1. 简介

在 Go 语言应用程序中，为了应对不可预知的运行时异常（即 `panic`），并防止程序因未捕获的 `panic` 而崩溃掉线，本项目设计了一套健壮的 `panic` 捕获和恢复方案。

本方案旨在：
*   **`panic` 捕获与恢复**: 在关键执行路径中（如 HTTP 请求处理、Goroutine 入口），通过 `defer` 和 `recover` 机制捕获 `panic`，防止程序崩溃，并进行优雅降级或错误报告。
*   **错误堆栈追踪**: 增强 `panic` 错误信息，包含调用堆栈，便于问题定位。

### 2. 设计原则

*   **Go 语言哲学**: 优先使用 `error` 类型返回错误，而不是 `panic`。`panic` 应仅用于表示不可恢复的程序错误或异常情况。
*   **全局 `panic` 恢复**: 在应用程序的入口点（例如 HTTP 服务器中间件、独立的 Goroutine）设置 `panic` 恢复机制。
*   **友好错误响应**: 捕获 `panic` 后，向客户端返回统一且友好的错误响应，避免暴露敏感的内部信息。
*   **详细日志**: 捕获的 `panic` 及其堆栈信息应详细记录到日志中，便于事后分析和排查。

### 3. `panic` 捕获与恢复机制

尽管项目鼓励返回 `error` 而非 `panic`，但为了应对不可预测的运行时异常（例如数组越界、空指针解引用或第三方库抛出的 `panic`），必须建立一个健壮的 `panic` 捕获和恢复机制。

**实现方案**:

1.  **全局中间件捕获**: 在 Gin Web 框架中，通常会实现一个全局的中间件来捕获所有 HTTP 请求处理过程中的 `panic`。

    ```go
    // 示例：Gin panic 恢复中间件 (伪代码)
    import (
        "fmt"
        "runtime/debug"
        "github.com/gin-gonic/gin"
        "go.uber.org/zap"
        "ttpos-server-go/app/errors"
        "ttpos-server-go/pkg/helper"
        "ttpos-server-go/pkg/logger"
    )

    func RecoveryMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
            defer func() {
                if r := recover(); r != nil {
                    // 捕获 panic
                    err, ok := r.(error)
                    if !ok {
                        err = fmt.Errorf("%v", r)
                    }

                    // 记录 panic 堆栈信息
                    stack := debug.Stack() // 获取完整的堆栈信息
                    logger.Logger.Error("Panic recovered",
                        zap.Error(err),
                        zap.ByteString("stack", stack),
                        zap.Any("request_info", c.Request.URL.Path), // 记录请求信息
                    )

                    // 返回统一的错误响应给客户端
                    helper.Fail(c, errors.ErrInternal.GetCode(), errors.ErrInternal.Error())
                    c.Abort() // 终止当前请求
                }
            }()
            c.Next() // 继续处理请求
        }
    }
    // 在 Gin 路由中注册: router.Use(RecoveryMiddleware())
    ```

    **关键点**:
    *   使用 `defer` 语句确保在函数返回前执行 `recover()`。
    *   `recover()` 返回 `nil` 表示没有 `panic` 发生，否则返回 `panic` 的值。
    *   捕获 `panic` 后，应立即记录详细的堆栈信息 (`debug.Stack()`) 到日志系统 (如 Zap)。
    *   向客户端返回一个统一的、不暴露内部细节的错误响应 (`errors.ErrInternal`)。
    *   `c.Abort()` 终止当前请求的处理链。

2.  **Goroutine 独立捕获**: 对于应用程序中启动的非请求相关的独立 Goroutine (例如后台任务、事件处理器)，也应在其入口处添加 `defer` + `recover` 机制，以防止单个 Goroutine 的 `panic` 导致整个应用程序崩溃。

    ```go
    // 示例：独立 Goroutine panic 恢复 (伪代码)
    import (
        "fmt"
        "runtime/debug"
        "go.uber.org/zap"
        "ttpos-server-go/pkg/logger"
    )

    func runBackgroundTask() {
        defer func() {
            if r := recover(); r != nil {
                err, ok := r.(error)
                if !ok {
                    err = fmt.Errorf("%v", r)
                }
                stack := debug.Stack()
                logger.Logger.Error("Background task panic recovered",
                    zap.Error(err),
                    zap.ByteString("stack", stack),
                )
                // 可以选择重启 Goroutine 或发送告警
            }
        }()
        // 实际的后台任务逻辑
        // ...
    }

    // go runBackgroundTask()
    ```

**3. 定时任务中的 `panic` 捕获**:

在应用程序的定时任务中，每一项定时任务的启动方法都明确地集成了 `panic` 捕获机制，以防止单个定时任务的异常导致整个服务崩溃。通过在任务执行函数的入口处使用 `defer` 和 `recover`，确保了即使任务内部发生不可预期的运行时错误，也能被妥善处理并记录，而不会中断其他任务或主程序的运行。

例如，在 `main/app/tasks/translate_when_sync.go` 文件中，可以看到如下的 `panic` 捕获逻辑：

```go
// ... existing code ...
func RunTranslateWhenSyncTask() {
	defer func() {
		if err := recover(); err != nil {
			// 捕获 panic 并记录日志
			logger.Logger.Error("定时任务 [RunTranslateWhenSyncTask] 发生 panic", zap.Any("panic_value", err), zap.Stack("stack_trace"))
		}
	}()
	// 定时任务的实际业务逻辑
	// ...
}
// ... existing code ...
```

**关键点**:
*   **任务隔离**: 每个定时任务的 `panic` 捕获机制确保了任务之间的独立性，一个任务的失败不会连锁影响其他任务。
*   **服务健壮性**: 即使定时任务内部发生 `panic`，也不会影响主程序的运行和稳定性。
*   **详细日志**: 定时任务中捕获的 `panic` 会被记录到日志中，包含堆栈信息，便于排查。

**4. `utils.Go` 方法的安全协程启动**:

本项目在 `pkg/utils/goroutine.go` 中提供了 `Go` 方法，这是一个便捷的安全协程启动工具。它自动集成了 `defer` + `recover` 机制，确保任何通过 `utils.Go` 启动的协程在发生 `panic` 时，能够被捕获并记录日志，而不会导致整个应用程序崩溃。

```go
// ... existing code ...
package utils

import (
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// ExecuteSafeGoroutine 以安全模式启动一个协程，内部会捕获 panic 并记录日志。
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 捕获 panic 并记录日志
				logger.Logger.Error("独立协程发生 panic", zap.Any("panic_value", r), zap.Stack("stack_trace"))
			}
		}()
		fn()
	}()
}
// ... existing code ...
```

**使用示例**:

业务逻辑中，如果需要启动一个独立的 Goroutine 来执行某个任务，应优先使用 `utils.Go`：

```go
// 替代传统的 go func() { ... }()
utils.Go(func() {
    // 您的协程业务逻辑，即使发生 panic 也不会导致主程序崩溃
    // 例如：可能引发 panic 的操作
    var s []int
    _ = s[0] // 访问越界，将引发 panic
})
```

**关键点**:
*   **自动化 `panic` 捕获**: 开发者无需手动在每个 Goroutine 中编写 `defer` + `recover` 代码。
*   **统一日志记录**: 所有捕获的 `panic` 都会通过 `logger.Logger.Error` 记录，包含 `panic` 值和完整的堆栈信息。
*   **提高开发效率与健壮性**: 简化了安全协程的编写，降低了因遗漏 `recover` 而导致程序崩溃的风险。

### 4. 日志记录

详细的日志记录是 `panic` 捕获不可或缺的一部分。本项目应使用结构化日志库 (如 `go.uber.org/zap`) 记录：
*   **`panic` 捕获**: `panic` 的具体值、完整的堆栈信息以及相关的请求上下文或 Goroutine 信息。

### 5. 总结

本项目通过全局的 Gin 中间件和独立的 Goroutine 入口处的 `defer` + `recover` 模式（包括 `utils.Go` 方法），实现了对 `panic` 的健壮捕获和恢复，从而有效防止程序崩溃掉线。这套方案确保了在程序发生任何不可预期的 `panic` 时，系统都能够优雅地处理，提供友好的用户体验，并为开发者提供详细的错误诊断信息，保障系统的稳定性和可靠性。