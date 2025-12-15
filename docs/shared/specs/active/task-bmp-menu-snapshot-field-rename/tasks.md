# 调整 GetMenuSnapshotResp.content 为 GetMenuSnapshotResp.menu_data 任务分解

> 本文档定义调整菜单快照响应字段命名的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 3  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: Protobuf 定义修改

- [ ] 1.1 修改 Protobuf 字段定义

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 将 `GetMenuSnapshotResp.content` 字段重命名为 `menu_data`
  - Requirements: Requirement 1
  - Leverage: 现有 Protobuf 文件: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Prompt: Role: Protobuf Developer | Task: 将 GetMenuSnapshotResp 消息中的 `string content = 2;` 改为 `string menu_data = 2;` | Context: 字段编号保持为 2，字段类型保持为 string，注释保持清晰 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，字段命名使用 snake_case | Success: 字段定义修改成功，与 SaveMenuSnapshotReq.menu_data 保持一致

---

## Phase 2: 代码生成和验证

- [ ] 2.1 重新生成 Protobuf Go 代码

  - File: `ttpos-bmp/app/ttpos-takeout/api/menu/menu.pb.go`（自动生成）
  - Purpose: 根据新的 Protobuf 定义生成 Go 代码
  - Requirements: Requirement 2
  - Leverage: Task 1.1 的 Protobuf 文件
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao`（或根据项目规范执行 protobuf 代码生成命令）
  - Success: 代码生成成功，包含 `MenuData` 字段和 `GetMenuData()` 方法

- [ ] 2.2 验证生成的代码

  - File: `ttpos-bmp/app/ttpos-takeout/api/menu/menu.pb.go`
  - Purpose: 确认生成的代码正确，字段名和方法名已更新
  - Requirements: Requirement 2
  - Leverage: Task 2.1 生成的代码
  - Success: 生成的代码包含 `MenuData` 字段和 `GetMenuData()` 方法，不再包含 `Content` 字段和 `GetContent()` 方法

---

## Phase 3: 业务代码更新

- [ ] 3.1 更新 Logic 层字段引用

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go`
  - Purpose: 更新 `GetMenuSnapshot` 方法中的字段引用，使用 `MenuData` 替代 `Content`
  - Requirements: Requirement 3
  - Leverage: 现有 Logic 实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go`，Task 2.1 生成的代码
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 更新 GetMenuSnapshot 方法，将 `resp.Content = content` 改为 `resp.MenuData = menuData`，同时将变量名从 `content` 改为 `menuData` 以提高可读性 | Context: 方法位于 channel_menu.go 第 92-132 行，需要修改第 115 行和第 126 行 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，保持业务逻辑不变 | Success: 字段引用更新成功，代码编译通过

- [ ] 3.2 检查其他字段引用

  - File: `ttpos-bmp/app/ttpos-takeout/`
  - Purpose: 搜索项目中所有对 `GetMenuSnapshotResp.Content` 或 `GetContent()` 的引用，确保全部更新
  - Requirements: Requirement 3
  - Leverage: 使用 grep 搜索: `grep -r "\.Content\|GetContent" ttpos-bmp/app/ttpos-takeout`
  - Command: `grep -r "\.Content\|GetContent\|GetMenuSnapshotResp" ttpos-bmp/app/ttpos-takeout`
  - Success: 确认所有引用已更新，无遗漏

---

## Phase 4: 测试和验证

- [ ] 4.1 编译验证

  - File: `ttpos-bmp/app/ttpos-takeout/`
  - Purpose: 确保项目编译成功，无语法错误
  - Requirements: 所有功能需求
  - Leverage: Task 1.1, 2.1, 3.1 的实现
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go build ./...`
  - Success: 编译成功，无错误

- [ ] 4.2 单元测试（如有）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu_test.go`
  - Purpose: 运行相关单元测试，确保功能正常
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件（如有）
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go test ./internal/logic/channel_menu/...`
  - Success: 所有测试通过

- [ ] 4.3 手动测试验证

  - File: -
  - Purpose: 手动调用 `GetMenuSnapshot` 接口，验证响应字段为 `menu_data`
  - Requirements: 所有功能需求
  - Leverage: gRPC 客户端工具或 Postman
  - Success: 接口调用成功，响应字段名为 `menu_data`，字段值正确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Protobuf 文件格式正确
- [ ] 所有测试通过（如有）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 代码注释已更新（如需要）
- [ ] 相关文档已更新（如有）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-bmp-menu-snapshot-field-rename/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-menu-snapshot-field-rename/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-bmp-menu-snapshot-field-rename/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-menu-snapshot-field-rename/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-bmp-menu-snapshot-field-rename/tasks.md)" | bc
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

### Protobuf 开发

```
Role: Protobuf Developer

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 字段命名使用 snake_case
- 字段编号保持连续
- 字段类型正确
- 注释清晰

Success Criteria:
- {成功标准1}
- Protobuf 文件格式正确
- 代码生成成功
```

### Go BMP 开发

```
Role: Go Developer with GoFrame expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录
- 遵循 GoFrame 项目结构
- 不使用 panic，返回 error

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 编译成功
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-15  
**维护者**: 后端开发组
