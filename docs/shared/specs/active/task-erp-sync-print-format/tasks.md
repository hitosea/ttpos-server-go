# task-erp-sync-print-format 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 4 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: 核心实现

### 1.1 扩展 Service 接口

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/service/setup.go` |
| Purpose | 新增 SyncPrintFormat 接口声明 |
| Requirements | R1: 执行指令获取打印格式文档 |
| Leverage | 现有 ISetup 接口模式 |

**实现要点**:
- 新增 `SyncPrintFormat(ctx context.Context, dirBase string) (*SyncPrintFormatResult, error)` 方法声明
- 新增 `SyncPrintFormatResult` 结构体定义

- [ ] 完成

---

### 1.2 实现 Logic 业务逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` |
| Purpose | 实现打印格式同步核心逻辑 |
| Requirements | R1, R2, R3 |
| Leverage | initDocumentsFromDir 错误处理模式, service.Rpc() |

**实现要点**:
```go
func (s *sSetup) SyncPrintFormat(ctx context.Context, dirBase string) (*SyncPrintFormatResult, error) {
    // 1. 调用 ERP API 获取 Print Format 列表
    //    使用 service.Doctype().List() 或 service.Rpc().Execute()
    //    过滤条件: name LIKE "Wallace%"

    // 2. 遍历文档列表，获取每个文档详情
    //    使用 service.Document().Get()

    // 3. 保存为 JSON 文件
    //    文件名: {文档名称}.json
    //    目录: dirBase (默认 ./manifest/printformat/html/)

    // 4. 统计结果
    //    记录成功/失败数量和错误详情
}
```

**关键点**:
- 使用 `gerror` 处理错误
- 单文档失败不影响其他文档
- 日志包含详细上下文

- [ ] 完成

---

### 1.3 新增 CLI 指令

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/cmd/cmd.go` |
| Purpose | 新增 sync-print-format 命令 |
| Requirements | R1, R3 |
| Leverage | ErpMigrate 指令模式 |

**实现要点**:
```go
var ErpSyncPrintFormat = &gcmd.Command{
    Name:  "sync-print-format",
    Usage: "sync-print-format --siteCode 1 [--dirBase ./manifest/printformat/html/]",
    Brief: "从 ERP 同步 Wallace 打印格式到本地",
    Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
        // 参数解析
        siteCode := parser.GetOpt("siteCode", "").String()
        dirBase := parser.GetOpt("dirBase", "./manifest/printformat/html/").String()

        // 参数校验
        if siteCode == "" {
            g.Log().Error(ctx, "siteCode 参数必填")
            return gerror.New("siteCode 参数必填")
        }

        // 设置 context
        ctx = grpcx.Ctx.NewIncoming(ctx, g.Map{
            consts.ContextSiteCode: siteCode,
        })

        // 调用 service
        result, err := service.Setup().SyncPrintFormat(ctx, dirBase)
        if err != nil {
            g.Log().Error(ctx, "同步打印格式失败", err)
            return err
        }

        // 输出结果摘要
        g.Log().Infof(ctx, "同步完成! 成功: %d, 失败: %d, 总计: %d",
            result.Success, result.Failed, result.Total)

        os.Exit(0)
        return nil
    },
}
```

- [ ] 完成

---

### 1.4 注册命令到 main.go

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/main.go` |
| Purpose | 将新命令注册到 CLI |
| Requirements | - |
| Leverage | 现有命令注册模式 |

**实现要点**:
- 在 main.go 中添加 `cmd.ErpSyncPrintFormat` 到命令列表

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [ ] AC1: 执行指令获取 Wallace 文档列表
- [ ] AC2: 文档保存为 JSON 文件
- [ ] AC3: 输出同步结果摘要
- [ ] 参数缺失时输出明确错误

### BMP 规范
- [ ] 使用 `gerror` 处理错误
- [ ] 日志包含上下文信息
- [ ] 遵循 GoFrame 编码规范

---

## 验收测试

### 手动测试命令

```bash
# 进入 ttpos-erp 目录
cd ttpos-bmp/app/ttpos-erp

# 测试同步命令
go run main.go sync-print-format --siteCode 1

# 检查输出文件
ls -la manifest/printformat/html/*.json

# 测试参数缺失
go run main.go sync-print-format
# 预期: 输出 "siteCode 参数必填" 错误
```

---

**版本**: v1.0.0
**创建日期**: 2026-01-15
