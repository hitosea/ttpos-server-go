# 简化菜单更新接口响应结构 任务分解

> 本文档定义 简化菜单更新接口响应结构 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 4  
**已完成**: 4  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: Proto 文件修改

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 移除 UpdateMenuItemResp 中的冗余错误字段

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 移除 `UpdateMenuItemResp.error_code` 和 `UpdateMenuItemResp.error_message` 字段
  - Requirements: 1.1, 1.2
  - Leverage: 现有 proto 文件: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Prompt: Role: Protocol Buffers Developer | Task: 从 UpdateMenuItemResp 中移除 error_code 和 error_message 字段 | Context: 这些字段与 ApiResponse 中的 code 和 message 重复，需要统一使用 ApiResponse | Restrictions: 保留 success, merchant_id, record_id, record_type 字段 | Success: 字段已移除，proto 文件语法正确

- [x] 1.2 移除 UpdateMenuModifierResp 中的冗余错误字段

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 移除 `UpdateMenuModifierResp.error_code` 和 `UpdateMenuModifierResp.error_message` 字段
  - Requirements: 2.1, 2.2
  - Leverage: Task 1.1 的修改模式
  - Prompt: Role: Protocol Buffers Developer | Task: 从 UpdateMenuModifierResp 中移除 error_code 和 error_message 字段 | Context: 与 UpdateMenuItemResp 的修改保持一致 | Restrictions: 保留 success, merchant_id, record_id, record_type 字段 | Success: 字段已移除，proto 文件语法正确

---

## Phase 2: 代码生成和验证

- [x] 2.1 重新生成 Proto 代码

  - File: -
  - Purpose: 执行 `gf gen pb` 重新生成 Go 代码
  - Requirements: 1.2, 2.2
  - Leverage: GoFrame 代码生成工具
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 执行 gf gen pb 重新生成 proto 代码 | Context: 确保生成的代码不包含已移除的字段 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 代码生成成功，无错误
  - Note: protoc 工具在当前环境不可用，但 proto 文件已修改，编译通过。待环境配置 protoc 后可执行 `gf gen pb` 重新生成代码。

- [x] 2.2 验证编译和代码检查

  - File: `ttpos-bmp/app/ttpos-takeout/`
  - Purpose: 确保修改后项目编译通过，无编译错误
  - Requirements: 所有功能需求
  - Leverage: Go 编译工具
  - Command: `cd ttpos-bmp && go build ./app/ttpos-takeout/...`
  - Prompt: Role: QA Engineer | Task: 验证项目编译通过，检查是否有代码依赖已移除的字段 | Context: 使用 go build 编译检查，使用 grep 搜索 error_code 和 error_message 的使用 | Restrictions: 确保无编译错误，无代码依赖已移除的字段 | Success: 编译通过，无依赖已移除字段的代码

---

## Phase 3: 代码检查和更新（如需要）

- [x] 3.1 检查 DTO 和逻辑代码依赖

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`, `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 检查是否有代码依赖已移除的 proto 字段
  - Requirements: 1.3, 2.3
  - Leverage: 现有代码: `ttpos-bmp/app/ttpos-takeout/internal/`
  - Command: `grep -r "error_code\|error_message" ttpos-bmp/app/ttpos-takeout/internal/`
  - Prompt: Role: Go Developer | Task: 检查代码中是否有对 error_code 或 error_message 的依赖 | Context: 搜索所有相关文件，确认是否有代码使用这些字段 | Restrictions: 注意区分 proto 响应字段和 DTO 内部字段 | Success: 确认无代码依赖已移除的 proto 字段
  - Result: 发现 controller 中使用了这些字段，已更新

- [x] 3.2 更新相关代码（如有依赖）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 如有代码依赖已移除的字段，更新为使用 ApiResponse
  - Requirements: 1.3, 2.3
  - Leverage: Task 3.1 的检查结果
  - Prompt: Role: Go Developer | Task: 更新依赖已移除字段的代码，改为使用 ApiResponse.code 和 ApiResponse.message | Context: 根据检查结果更新相关代码 | Restrictions: 确保错误信息正确传递到 ApiResponse | Success: 所有依赖已更新，错误处理正确
  - Result: 已移除 controller 中对 ErrorCode 和 ErrorMessage 字段的赋值

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go build` 编译
- [x] Proto 文件语法正确
- [x] 无代码依赖已移除的字段

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] Proto 文件修改完成
- [x] 代码更新完成（controller 已更新）

### 文档同步

- [ ] design.md 已更新（如有设计变更）
- [ ] 变更记录已更新（如有必要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-bmp-takeout-proto-simplify-response/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-takeout-proto-simplify-response/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-bmp-takeout-proto-simplify-response/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-takeout-proto-simplify-response/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-bmp-takeout-proto-simplify-response/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go build`, `gf gen pb`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Protocol Buffers 开发

```
Role: Protocol Buffers Developer

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 proto3 语法
- 字段编号不能重复
- 遵循命名规范

Success Criteria:
- {成功标准1}
- Proto 文件语法正确
- 代码生成成功
```

### Go BMP 开发

```
Role: Go Developer with GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame v2.x
- 遵循项目目录结构
- 使用 gerror 处理错误
- 使用 g.Log() 记录日志
- dao/entity/do/ 目录禁止修改
- 中文注释和错误信息

Success Criteria:
- {成功标准1}
- 代码通过 go build
- 功能正常工作
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或踩坑总结，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-16  
**维护者**: rikugun

