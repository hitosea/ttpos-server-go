# ERP 迁移命令分离优化 任务分解

> 本文档定义 ERP 迁移命令分离优化的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 5  
**已完成**: 2  
**进行中**: -  
**完成率**: 40%

---

## Phase 1: 命令结构调整

- [x] 1.1 修改 main.go 实现命令路由逻辑

  - File: `ttpos-bmp/app/ttpos-erp/main.go`
  - Purpose: 根据命令行参数选择执行对应命令
  - Requirements: Requirement 1
  - Leverage: 现有命令定义 `ttpos-bmp/app/ttpos-erp/internal/cmd/cmd.go`，GoFrame 命令文档
  - Prompt: Role: Go Developer | Task: 修改 main.go，实现命令路由逻辑，根据命令行参数选择执行 Main、ErpMigrate 或 ErpAllMigrate | Context: 使用 gcmd.Parse() 解析命令行参数，根据第一个参数判断命令名称 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，保持向后兼容 | Success: 命令路由逻辑正确，迁移命令执行时不启动 HTTP 服务器
  - Code Example:
    ```go
    func main() {
        ctx := gctx.GetInitCtx()
        parser, err := gcmd.Parse(nil)
        if err != nil {
            panic(err)
        }
        commandName := parser.GetArg(0).String()
        switch commandName {
        case "migrate":
            cmd.ErpMigrate.Run(ctx)
            return
        case "migrate-all":
            cmd.ErpAllMigrate.Run(ctx)
            return
        default:
            cmd.Main.Run(ctx)
        }
    }
    ```

- [ ] 1.2 测试命令执行逻辑

  - File: `ttpos-bmp/app/ttpos-erp/main.go`
  - Purpose: 验证命令路由和执行逻辑正确
  - Requirements: Requirement 1, Requirement 2
  - Leverage: 现有测试环境
  - Prompt: Role: QA Engineer | Task: 测试命令执行逻辑，验证迁移命令不启动 HTTP 服务器 | Context: 执行 migrate 和 migrate-all 命令，检查是否启动 HTTP 服务器 | Restrictions: 使用测试环境 | Success: 迁移命令执行时不启动 HTTP 服务器，Main 命令正常启动 HTTP 服务器

- [ ] 1.3 验证向后兼容性

  - File: `ttpos-bmp/app/ttpos-erp/main.go`
  - Purpose: 确保现有命令行参数格式继续有效
  - Requirements: Requirement 3
  - Leverage: 现有 CI/CD 脚本和运维脚本
  - Prompt: Role: QA Engineer | Task: 验证向后兼容性，确保现有命令行参数格式继续有效 | Context: 使用现有参数格式执行命令，验证功能正常 | Restrictions: 保持参数格式不变 | Success: 现有命令行参数格式继续有效，功能正常

---

## Phase 2: 测试和文档

- [x] 2.1 代码质量检查

  - File: `ttpos-bmp/app/ttpos-erp/main.go`
  - Purpose: 确保代码符合规范
  - Requirements: 所有功能需求
  - Leverage: Go 代码规范
  - Prompt: Role: Go Developer | Task: 执行代码质量检查，确保代码符合规范 | Context: 执行 go fmt、go vet 检查 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc | Success: 代码通过格式化、静态检查和规范检查

- [ ] 2.2 更新相关文档

  - File: `docs/shared/api/erp_api.md`（如果存在）、README.md
  - Purpose: 更新文档，说明新的命令执行方式
  - Requirements: Requirement 3
  - Leverage: 现有文档
  - Prompt: Role: Technical Writer | Task: 更新相关文档，说明新的命令执行方式 | Context: 更新命令使用说明，说明迁移命令独立执行 | Restrictions: 保持文档格式一致 | Success: 文档已更新，说明清晰

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码符合 Go BMP 规范

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 向后兼容性验证通过

---

## 进度追踪

### 执行流程

1. **选择任务**: 从 Phase 1 开始，按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的实现方案
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### 预计时间

- Phase 1: 0.5 天（4 小时）
- Phase 2: 0.5 天（4 小时）
- **总计**: 1 天（8 小时）= **SP 1**

---

## 附录：AI Prompt 示例

### 命令路由实现

```
Role: Go Developer with GoFrame expertise

Task: 修改 main.go，实现命令路由逻辑

Context:
- File: ttpos-bmp/app/ttpos-erp/main.go
- Leverage: design.md 中的方案一实现代码
- Requirements: requirements.md Requirement 1
- Dependencies: gcmd.Parse(), cmd.Main, cmd.ErpMigrate, cmd.ErpAllMigrate

Implementation Steps:
1. 使用 gcmd.Parse() 解析命令行参数
2. 获取第一个参数（命令名称）
3. 根据命令名称选择执行对应命令
4. 默认执行 Main 命令

Restrictions:
- 保持向后兼容
- 遵循 Go BMP 规范
- 错误处理完善

Success Criteria:
- 代码通过 go fmt 和 go vet
- 命令路由逻辑正确
- 迁移命令执行时不启动 HTTP 服务器
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当执行任务中形成复盘/优化建议时，及时沉淀 Episode 并在本节更新名称。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

