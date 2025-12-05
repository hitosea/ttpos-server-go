# ERP 迁移命令分离优化 设计文档

> 本文档定义 ERP 迁移命令分离优化的技术设计和实现方案。

## 📋 概述

ERP 迁移命令分离优化通过调整命令注册结构，将迁移命令从 Main 的子命令改为独立的顶级命令，使得执行迁移命令时不需要启动 HTTP 服务器。

核心实现包括：
- 修改 `main.go` 中的命令注册逻辑
- 使用 GoFrame 的命令行参数解析机制
- 确保命令独立执行，不相互影响

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- ✅ 遵循 GoFrame 命令行工具规范
- ✅ 保持代码结构清晰
- ✅ 错误处理使用 gerror
- ✅ 不使用 panic，返回 error

---

## 🔄 代码复用分析

### 可复用的现有组件

- **命令定义**: `ttpos-bmp/app/ttpos-erp/internal/cmd/cmd.go` - 复用现有命令定义
- **迁移逻辑**: `service.Setup().InitErpDocTypeWithDirname()` - 复用现有迁移逻辑
- **GoFrame 命令框架**: 使用 GoFrame 的 `gcmd` 包

### 集成点

- **命令注册**: 在 `main.go` 中注册命令
- **命令执行**: 通过 GoFrame 的命令行参数解析执行对应命令

---

## 🏗️ 架构设计

### 当前架构

```
main.go
  └── Main (HTTP 服务器)
      ├── ErpMigrate (子命令)
      └── ErpAllMigrate (子命令)
```

**问题**：执行子命令时可能触发父命令的初始化逻辑

### 目标架构

**方案一（推荐）**：独立命令注册

```
main.go
  ├── Main (HTTP 服务器) - 独立命令
  ├── ErpMigrate - 独立命令
  └── ErpAllMigrate - 独立命令
```

**方案二**：根命令统一管理

```
main.go
  └── Root (根命令)
      ├── Main (HTTP 服务器)
      ├── ErpMigrate
      └── ErpAllMigrate
```

---

## 🔧 实现方案

### 方案一：独立命令注册（推荐）

**实现思路**：根据命令行参数判断执行哪个命令，分别注册和执行。

**代码实现**：

```go
// main.go
package main

import (
	_ "ttpos-bmp/app/ttpos-erp/internal/packed"
	_ "ttpos-bmp/app/ttpos-erp/internal/boot"
	_ "ttpos-bmp/app/ttpos-erp/internal/logic"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gcmd"

	"ttpos-bmp/app/ttpos-erp/internal/cmd"
)

func main() {
	ctx := gctx.GetInitCtx()
	
	// 解析命令行参数
	parser, err := gcmd.Parse(nil)
	if err != nil {
		panic(err)
	}
	
	// 根据命令名称执行对应命令
	commandName := parser.GetArg(0).String()
	
	switch commandName {
	case "migrate":
		// 执行迁移命令
		if err := cmd.ErpMigrate.Run(ctx); err != nil {
			panic(err)
		}
		return
	case "migrate-all":
		// 执行全量迁移命令
		if err := cmd.ErpAllMigrate.Run(ctx); err != nil {
			panic(err)
		}
		return
	default:
		// 默认执行 Main 命令（HTTP 服务器）
		if err := cmd.Main.Run(ctx); err != nil {
			panic(err)
		}
	}
}
```

**优点**：
- 实现简单，代码清晰
- 命令完全独立，互不影响
- 向后兼容性好

**缺点**：
- 需要手动解析命令行参数

---

### 方案二：根命令统一管理

**实现思路**：创建根命令，将所有命令添加为根命令的子命令。

**代码实现**：

```go
// cmd.go 中新增根命令
var Root = &gcmd.Command{
	Name:  "ttpos-erp",
	Usage: "ttpos-erp [command] [options]",
	Brief: "TTPOS ERP 服务命令行工具",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		// 显示帮助信息
		parser.PrintHelp()
		return nil
	},
}

// main.go
func main() {
	// 将所有命令添加为根命令的子命令
	cmd.Root.AddCommand(cmd.Main)
	cmd.Root.AddCommand(cmd.ErpMigrate)
	cmd.Root.AddCommand(cmd.ErpAllMigrate)
	
	// 运行根命令
	if err := cmd.Root.Run(gctx.GetInitCtx()); err != nil {
		panic(err)
	}
}
```

**优点**：
- 符合 GoFrame 命令框架的最佳实践
- 命令结构清晰，便于扩展
- 自动支持帮助信息

**缺点**：
- 需要修改命令定义文件
- 可能影响现有命令的执行方式

---

## 🎯 推荐方案

**推荐使用方案一**，原因：
1. **实现简单**：只需修改 `main.go`，不需要修改命令定义
2. **向后兼容**：保持现有命令参数格式不变
3. **风险较低**：改动范围小，易于测试和验证
4. **符合需求**：完全满足"迁移命令不启动 HTTP 服务器"的需求

---

## 🔍 实现细节

### 命令执行流程

**方案一执行流程**：

```
1. main() 函数启动
2. 解析命令行参数
3. 判断命令名称
   - 如果是 "migrate" → 执行 ErpMigrate.Run()
   - 如果是 "migrate-all" → 执行 ErpAllMigrate.Run()
   - 否则 → 执行 Main.Run()
4. 命令执行完成
```

### 命令行参数解析

使用 GoFrame 的 `gcmd.Parse()` 解析命令行参数：

```go
parser, err := gcmd.Parse(nil)
if err != nil {
    panic(err)
}

// 获取第一个参数（命令名称）
commandName := parser.GetArg(0).String()

// 获取选项参数
siteCode := parser.GetOpt("siteCode", "1").String()
dirBase := parser.GetOpt("dirBase", "./manifest/erp-migrate/v2.5").String()
```

### 错误处理

- 使用 `panic` 处理致命错误（如命令解析失败）
- 命令执行错误由命令内部处理（如 `ErpMigrate` 中的错误处理）

---

## 🧪 测试策略

### 单元测试

**测试用例**：
1. 测试命令参数解析逻辑
2. 测试命令路由逻辑（根据参数选择命令）
3. 测试向后兼容性（现有参数格式）

### 集成测试

**测试场景**：
1. 执行 `gf run main.go --args "migrate --siteCode 1 --dirBase ./manifest/erp-migrate/v2.9"`
   - 验证：仅执行迁移逻辑，不启动 HTTP 服务器
2. 执行 `gf run main.go --args "migrate-all --siteCode 1 --dirBase ./manifest/erp-migrate"`
   - 验证：仅执行全量迁移逻辑，不启动 HTTP 服务器
3. 执行 `gf run main.go`
   - 验证：正常启动 HTTP 服务器

---

## 📈 性能优化

### 优化措施

1. **避免不必要的初始化**：迁移命令执行时不初始化 HTTP 服务器
2. **快速退出**：迁移命令执行完成后立即退出，不保持进程运行

### 性能指标

- **迁移命令启动时间**: < 1 秒（不启动 HTTP 服务器）
- **命令执行响应**: 立即执行，无延迟

---

## 🚨 错误处理

### 主要错误场景

1. **命令解析失败**: 返回错误信息，显示帮助信息
2. **命令执行失败**: 由命令内部处理，返回错误码

### 错误处理策略

```go
// 命令解析错误
parser, err := gcmd.Parse(nil)
if err != nil {
    g.Log().Error(ctx, "命令解析失败", err)
    parser.PrintHelp()
    os.Exit(1)
}

// 命令执行错误（在命令内部处理）
if err := cmd.ErpMigrate.Run(ctx); err != nil {
    g.Log().Error(ctx, "迁移命令执行失败", err)
    os.Exit(1)
}
```

---

## 📚 实现清单

### Phase 1: 命令结构调整（参见 tasks.md）

- [ ] 修改 `main.go`，实现命令路由逻辑
- [ ] 测试命令执行逻辑
- [ ] 验证向后兼容性

### Phase 2: 测试和文档（参见 tasks.md）

- [ ] 单元测试
- [ ] 集成测试
- [ ] 文档更新

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当设计结论可复用或踩坑较多时，沉淀 Episode 并在此更新名称，保持 Spec ↔ Graphiti 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**作者**: 后端开发组  
**审核者**: {待分配}

