# 优化 RedoPosConsumer 增加 SiteCode 过滤 任务分解

> 本文档定义优化 RedoPosConsumer 增加 SiteCode 过滤的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 5  
**进行中**: -  
**完成率**: 45%

---

## Phase 1: 代码修改

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 RedoPosConsumer.Handle 方法 - SavePosInvoice 分支

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
  - Purpose: 为 `MsgTypeSavePosInvoice` 消息类型的查询添加 `SiteCode` 过滤条件
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 
    - 现有代码: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go:393-414`
    - 参考实现: `SavePosInvoiceConsumer` (同文件:24-90) 已实现 SiteCode 过滤
  - Prompt: Role: Go Developer specializing in GoFrame Consumer | Task: 修改 RedoPosConsumer.Handle 方法中的 MsgTypeSavePosInvoice 分支，添加 SiteCode 过滤条件 | Context: 当前代码只使用 OpenPosEntryName 和 Docstatus 过滤，需要添加 SiteCode 过滤。当 SiteCode 为空时，记录警告日志但不中断处理（向后兼容） | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 g.Log().Warningf() 记录警告，使用 gerror.Wrapf() 包装错误 | Success: SiteCode 过滤已添加，向后兼容性已处理，代码通过 go fmt 和 go vet

- [x] 1.2 修改 RedoPosConsumer.Handle 方法 - CancelPosInvoice 分支

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
  - Purpose: 为 `MsgTypeCancelPosInvoice` 消息类型的查询添加 `SiteCode` 过滤条件
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 
    - 现有代码: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go:415-436`
    - 参考实现: `CancelPosInvoice` (同文件:190-285) 已实现 SiteCode 过滤
  - Prompt: Role: Go Developer specializing in GoFrame Consumer | Task: 修改 RedoPosConsumer.Handle 方法中的 MsgTypeCancelPosInvoice 分支，添加 SiteCode 过滤条件 | Context: 当前代码只使用 OpenPosEntryName 和 Docstatus 过滤，需要添加 SiteCode 过滤。当 SiteCode 为空时，记录警告日志但不中断处理（向后兼容） | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 g.Log().Warningf() 记录警告，使用 gerror.Wrapf() 包装错误 | Success: SiteCode 过滤已添加，向后兼容性已处理，代码通过 go fmt 和 go vet

- [x] 1.3 修改 RedoPosConsumer.Handle 方法 - ReturnPosInvoice 分支

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
  - Purpose: 为 `MsgTypeReturnPosInvoice` 消息类型的查询添加 `SiteCode` 过滤条件
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 
    - 现有代码: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go:437-458`
    - 参考实现: `ReturnPosInvoiceConsumer` (同文件:92-188) 已实现 SiteCode 过滤
  - Prompt: Role: Go Developer specializing in GoFrame Consumer | Task: 修改 RedoPosConsumer.Handle 方法中的 MsgTypeReturnPosInvoice 分支，添加 SiteCode 过滤条件 | Context: 当前代码只使用 OpenPosEntryName 和 Docstatus 过滤，需要添加 SiteCode 过滤。当 SiteCode 为空时，记录警告日志但不中断处理（向后兼容） | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 g.Log().Warningf() 记录警告，使用 gerror.Wrapf() 包装错误 | Success: SiteCode 过滤已添加，向后兼容性已处理，代码通过 go fmt 和 go vet

- [x] 1.4 修改 RedoPosConsumer.Handle 方法 - ClosePosEntry 分支

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
  - Purpose: 为 `MsgTypeClosePosEntry` 消息类型的查询添加 `SiteCode` 过滤条件
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 
    - 现有代码: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go:459-480`
    - 参考实现: `ClosePosEntryConsumer` (同文件:287-363) 已实现 SiteCode 过滤
  - Prompt: Role: Go Developer specializing in GoFrame Consumer | Task: 修改 RedoPosConsumer.Handle 方法中的 MsgTypeClosePosEntry 分支，添加 SiteCode 过滤条件 | Context: 当前代码只使用 PosOpenEntryName 和 Docstatus 过滤，需要添加 SiteCode 过滤。当 SiteCode 为空时，记录警告日志但不中断处理（向后兼容） | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 g.Log().Warningf() 记录警告，使用 gerror.Wrapf() 包装错误 | Success: SiteCode 过滤已添加，向后兼容性已处理，代码通过 go fmt 和 go vet

- [x] 1.5 添加 SiteCode 验证和警告日志

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go`
  - Purpose: 在消息处理前验证 SiteCode，当为空时记录警告日志
  - Requirements: 2.1, 2.2
  - Leverage: 
    - 现有代码: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer.go:377-390`
    - 参考实现: 其他 Consumer 的消息验证逻辑
  - Prompt: Role: Go Developer specializing in GoFrame Consumer | Task: 在 RedoPosConsumer.Handle 方法开始处添加 SiteCode 验证，当 SiteCode 为空时记录警告日志 | Context: 在解析消息后、处理消息前添加验证逻辑，记录警告但不中断处理流程（向后兼容） | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 g.Log().Warningf() 记录警告，包含消息内容以便排查 | Success: SiteCode 验证已添加，警告日志已记录，向后兼容性已处理

---

## Phase 2: 测试

- [ ] 2.1 编写 RedoPosConsumer 单元测试 - SavePosInvoice

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer_test.go`
  - Purpose: 测试 `MsgTypeSavePosInvoice` 消息类型的 SiteCode 过滤功能
  - Requirements: 1.1, 1.2, 1.3, 2.1, 2.2
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RedoPosConsumer.Handle 方法的 MsgTypeSavePosInvoice 分支编写单元测试 | Context: 测试 SiteCode 过滤功能，测试向后兼容性（SiteCode 为空），测试错误处理 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 GoFrame 测试框架 | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

- [ ] 2.2 编写 RedoPosConsumer 单元测试 - CancelPosInvoice

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer_test.go`
  - Purpose: 测试 `MsgTypeCancelPosInvoice` 消息类型的 SiteCode 过滤功能
  - Requirements: 1.1, 1.2, 1.3, 2.1, 2.2
  - Leverage: Task 2.1 的测试代码
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RedoPosConsumer.Handle 方法的 MsgTypeCancelPosInvoice 分支编写单元测试 | Context: 测试 SiteCode 过滤功能，测试向后兼容性（SiteCode 为空），测试错误处理 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 GoFrame 测试框架 | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

- [ ] 2.3 编写 RedoPosConsumer 单元测试 - ReturnPosInvoice

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer_test.go`
  - Purpose: 测试 `MsgTypeReturnPosInvoice` 消息类型的 SiteCode 过滤功能
  - Requirements: 1.1, 1.2, 1.3, 2.1, 2.2
  - Leverage: Task 2.1 的测试代码
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RedoPosConsumer.Handle 方法的 MsgTypeReturnPosInvoice 分支编写单元测试 | Context: 测试 SiteCode 过滤功能，测试向后兼容性（SiteCode 为空），测试错误处理 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 GoFrame 测试框架 | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

- [ ] 2.4 编写 RedoPosConsumer 单元测试 - ClosePosEntry

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer_test.go`
  - Purpose: 测试 `MsgTypeClosePosEntry` 消息类型的 SiteCode 过滤功能
  - Requirements: 1.1, 1.2, 1.3, 2.1, 2.2
  - Leverage: Task 2.1 的测试代码
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RedoPosConsumer.Handle 方法的 MsgTypeClosePosEntry 分支编写单元测试 | Context: 测试 SiteCode 过滤功能，测试向后兼容性（SiteCode 为空），测试错误处理 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，使用 GoFrame 测试框架 | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

- [ ] 2.5 编写集成测试 - 多站点场景

  - File: `ttpos-bmp/app/ttpos-erp/internal/consumer/selling/selling_consumer_integration_test.go`
  - Purpose: 测试多站点场景下的数据隔离功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试（如有）
  - Prompt: Role: QA Automation Engineer | Task: 实现多站点场景的集成测试 | Context: 测试不同站点的订单不会相互影响，测试 SiteCode 过滤的正确性 | Restrictions: 测试真实多站点场景 | Success: 集成测试通过，多站点数据隔离正确

---

## Phase 3: 数据库索引检查（可选）

- [ ] 3.1 检查并添加 site_code 索引

  - File: `admin/database/migrations/`（如需要）
  - Purpose: 确保 `site_code` 字段有索引，优化查询性能
  - Requirements: 性能要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Command: 检查现有索引，如无则创建迁移文件添加索引
  - Success: `site_code` 字段有索引，查询性能达标

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（Consumer: ≥ 70%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 代码注释已更新（关键逻辑有清晰的中文注释）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-bmp-redo-pos-consumer-site-code-filter/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-redo-pos-consumer-site-code-filter/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-bmp-redo-pos-consumer-site-code-filter/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-redo-pos-consumer-site-code-filter/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-bmp-redo-pos-consumer-site-code-filter/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP Consumer 开发

```
Role: Go Developer specializing in GoFrame Consumer

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-bmp.mdc, ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 g.Log() 记录日志
- 使用 gerror.Wrap() 包装错误
- 不使用 panic，返回 error
- 遵循 GoFrame 项目结构

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试（SiteCode 为空的情况）
- 多站点场景测试

Restrictions:
- 遵循 .cursor/rules/go-bmp.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率 ≥ 70%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-01  
**维护者**: 后端开发组

