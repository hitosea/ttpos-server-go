# ERP 迁移命令分离优化 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                                                                             |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **提案人**    | rikugun                                                                                                                                                          |
| **日期**      | 2025-11-24                                                                                                                                                       |
| **目标版本**  | v2.9                                                                                                                                                             |
| **状态**      | 已创建 Spec                                                                                                                                                      |
| **关联任务**  | -                                                                                                                                                                |
| **关联 Spec** | [docs/shared/specs/archived/v2.10.0/task-erp-migrate-command-separation/requirements.md](../../../shared/specs/active/task-erp-migrate-command-separation/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-bmp/app/ttpos-erp/internal/cmd/cmd.go` 中定义了三个命令：

- `Main` - 启动 HTTP 服务器
- `ErpMigrate` - 执行 ERP 数据迁移（单版本目录）
- `ErpAllMigrate` - 执行 ERP 全量迁移（遍历所有版本目录）

在 `main.go` 中，`ErpMigrate` 和 `ErpAllMigrate` 被添加为 `Main` 的子命令：

```go
err := cmd.Main.AddCommand(cmd.ErpMigrate)
err = cmd.Main.AddCommand(cmd.ErpAllMigrate)
cmd.Main.Run(gctx.GetInitCtx())
```

**问题**：

1. 执行迁移命令时，可能触发 Main 命令的初始化逻辑（如 HTTP 服务器启动）
2. 迁移命令执行完成后会调用 `os.Exit(0)`，但在此之前可能已经启动了不必要的服务
3. 命令结构不够清晰，迁移命令应该作为独立的命令行工具，而非 HTTP 服务的子命令

### 业务价值

- **提升执行效率**：迁移命令执行时不需要启动 HTTP 服务器，减少资源消耗
- **降低耦合度**：迁移工具与 HTTP 服务解耦，便于独立维护和测试
- **改善用户体验**：命令执行更快，无需等待 HTTP 服务初始化
- **符合最佳实践**：命令行工具应该独立于服务进程

### 目标用户

- [x] 开发人员
- [x] 运维人员
- [ ] 商户管理员
- [ ] 其他: **\_\_\_\_**

---

## 💡 解决方案概述

### 方案描述

将 `ErpMigrate` 和 `ErpAllMigrate` 从 `Main` 的子命令改为独立的顶级命令，使得它们可以独立执行，无需启动 HTTP 服务器。

**方案一（推荐）**：将迁移命令注册为独立的顶级命令

```go
// main.go
func main() {
    // 迁移命令作为独立命令
    cmd.ErpMigrate.Run(gctx.GetInitCtx())
    cmd.ErpAllMigrate.Run(gctx.GetInitCtx())

    // Main 命令独立运行
    cmd.Main.Run(gctx.GetInitCtx())
}
```

**方案二**：创建根命令，统一管理所有命令

```go
// cmd.go 中新增根命令
var Root = &gcmd.Command{
    Name:  "ttpos-erp",
    Brief: "TTPOS ERP 服务命令行工具",
    Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
        // 显示帮助信息
        return nil
    },
}

// main.go
func main() {
    cmd.Root.AddCommand(cmd.Main)
    cmd.Root.AddCommand(cmd.ErpMigrate)
    cmd.Root.AddCommand(cmd.ErpAllMigrate)
    cmd.Root.Run(gctx.GetInitCtx())
}
```

### 核心功能点

1. **命令分离**：将迁移命令从 Main 的子命令改为独立命令
2. **执行优化**：迁移命令执行时不启动 HTTP 服务器
3. **代码重构**：优化 `main.go` 中的命令注册逻辑
4. **向后兼容**：保持命令行参数和功能不变

### 影响范围

**涉及终端**：

- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：

- [x] 命令行工具 (`cmd.go`, `main.go`)
- [ ] API 接口
- [ ] 数据模型
- [ ] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: **\_\_\_\_**

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯命令结构调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：

1. **命令执行方式变更**：可能影响现有的 CI/CD 脚本或运维脚本
2. **命令参数解析**：需要确保参数解析逻辑正确

**缓解措施**：

1. **保持向后兼容**：确保命令行参数格式不变，仅调整命令结构
2. **充分测试**：在测试环境验证迁移命令的独立执行
3. **文档更新**：更新相关文档，说明新的命令执行方式

---

## 🔗 相关资源

### 参考需求

- GoFrame 命令文档: https://goframe.org/pages/viewpage.action?pageId=1114369
- 现有迁移命令实现: `ttpos-bmp/app/ttpos-erp/internal/cmd/cmd.go`

### 相关文档

- GoFrame 命令行工具规范: `.cursor/rules/go-bmp.mdc`
- 数据库迁移规范: `.cursor/rules/database.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`task-erp-migrate-command-separation`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 开发/运维人员  
**我想** 执行 ERP 迁移命令时不需要启动 HTTP 服务器  
**以便于** 提高执行效率，减少资源消耗，加快迁移速度

### AC 验收标准（初稿）

1. **WHEN** 执行 `gf run main.go --args "migrate --siteCode 1 --dirBase ./manifest/erp-migrate/v2.9"` **THEN** 系统 **SHALL** 仅执行迁移逻辑，不启动 HTTP 服务器
2. **WHEN** 执行 `gf run main.go --args "migrate-all --siteCode 1 --dirBase ./manifest/erp-migrate"` **THEN** 系统 **SHALL** 仅执行全量迁移逻辑，不启动 HTTP 服务器
3. **WHEN** 执行 `gf run main.go` **THEN** 系统 **SHALL** 正常启动 HTTP 服务器
4. **IF** 迁移命令执行成功 **THEN** 系统 **SHALL** 正常退出（`os.Exit(0)`）

### 技术方案（初稿）

**推荐方案**：使用 GoFrame 的命令行参数解析机制，根据命令行参数决定执行哪个命令。

**实现步骤**：

1. 修改 `main.go`，根据命令行参数选择执行命令
2. 确保迁移命令独立执行，不触发 Main 命令
3. 测试验证命令执行逻辑
4. 更新相关文档

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组  
**相关规范**: `.cursor/rules/go-bmp.mdc`
