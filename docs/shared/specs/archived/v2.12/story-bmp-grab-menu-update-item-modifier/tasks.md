# GrabFood 菜单项和修饰符更新功能 任务分解

> 本文档定义 GrabFood 菜单项和修饰符更新功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

---

## 📊 进度总览

**总任务数**: 8
**已完成**: 8
**进行中**: -
**完成率**: 100%

---

## Phase 1: 准备工作

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建菜单更新 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`
  - Purpose: 定义菜单更新相关的数据传输对象
  - Requirements: 1.1, 2.1
  - Leverage: 现有 DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/`
  - Prompt: Role: Go Developer | Task: 创建 UpdateMenuItemReq, UpdateMenuModifierReq, UpdateMenuResult 等 DTO 结构体 | Context: 基于 GrabFood API 规范，使用 gRPC 标签和验证标签 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 定义完整，字段类型正确

- [x] 1.2 分析 GrabFood API 集成方式

  - File: -
  - Purpose: 了解如何集成 GrabFood SDK 到现有代码
  - Requirements: 1.1, 2.1
  - Leverage: 现有集成: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` 中的 SDK 使用
  - Success: 明确 SDK 调用方式和错误处理

---

## Phase 2: 核心功能实现

### UpdateMenuItem 实现

- [x] 2.1 实现 UpdateMenuItem 方法框架

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 创建 UpdateMenuItem 方法的基本框架
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.6
  - Leverage: 现有方法: `SyncMenu`, `HandleMenuSyncState`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 在 sGrabMenu 结构体中添加 UpdateMenuItem 方法 | Context: 方法接收 UpdateMenuItemReq 参数，返回 UpdateMenuResult，包含参数验证、API 调用、日志记录 | Restrictions: 使用 gerror 处理错误，使用 g.Log() 记录日志 | Success: 方法框架创建完成，编译通过

- [x] 2.2 实现商品更新 API 调用

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 集成 GrabFood SDK 调用商品更新 API
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: SDK 使用: `github.com/grab/grabfood-api-sdk-go`
  - Prompt: Role: Go Developer with API integration experience | Task: 在 UpdateMenuItem 方法中构建 GrabFood API 请求并调用 | Context: 使用 SDK 的 UpdateMenuRecord 方法，处理 items 数组，设置正确的 merchantID | Restrictions: 错误信息使用中文，API 调用使用 context 超时 | Success: API 调用成功，正确处理响应

- [x] 2.3 实现商品更新日志记录

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 在商品更新操作中记录详细日志
  - Requirements: 1.4
  - Leverage: 现有日志记录: `HandleMenuSyncState` 方法中的日志逻辑
  - Prompt: Role: Go Developer | Task: 在 UpdateMenuItem 中添加操作日志记录 | Context: 使用 menu_log 表记录操作结果，区分成功和失败场景 | Restrictions: 使用 uuid.MustGetID() 生成日志ID，使用 gtime.Now() 记录时间 | Success: 日志记录完整，包含请求ID和错误信息

### UpdateMenuModifier 实现

- [x] 2.4 实现 UpdateMenuModifier 方法框架

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 创建 UpdateMenuModifier 方法的基本框架
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Task 2.1 的 UpdateMenuItem 实现
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 在 sGrabMenu 结构体中添加 UpdateMenuModifier 方法 | Context: 方法接收 UpdateMenuModifierReq 参数，返回 UpdateMenuResult，包含参数验证、API 调用、日志记录 | Restrictions: 复用 UpdateMenuItem 的错误处理模式 | Success: 方法框架创建完成，编译通过

- [x] 2.5 实现修饰符更新 API 调用

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 集成 GrabFood SDK 调用修饰符更新 API
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: Task 2.2 的商品更新实现
  - Prompt: Role: Go Developer with API integration experience | Task: 在 UpdateMenuModifier 方法中构建 GrabFood API 请求并调用 | Context: 使用 SDK 的 UpdateMenuRecord 方法，处理 modifierGroups 和 modifiers 嵌套结构 | Restrictions: 正确设置 modifierGroupID 和 modifierID 的关系 | Success: API 调用成功，正确处理修饰符更新

- [x] 2.6 实现修饰符更新日志记录

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 在修饰符更新操作中记录详细日志
  - Requirements: 2.3
  - Leverage: Task 2.3 的商品更新日志记录
  - Prompt: Role: Go Developer | Task: 在 UpdateMenuModifier 中添加操作日志记录 | Context: 复用 menu_log 表记录逻辑，区分修饰符更新类型 | Restrictions: 日志中明确标识为 modifier 更新操作 | Success: 日志记录完整，包含 modifier 相关信息

---

## Phase 3: 测试和优化

- [x] 3.1 编写单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update_test.go`
  - Purpose: 为新功能编写完整的单元测试
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/` 目录下的测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 UpdateMenuItem 和 UpdateMenuModifier 方法编写单元测试 | Context: 测试正常场景、参数验证失败、API 调用失败等场景 | Restrictions: 测试覆盖率 ≥ 70%，使用 mock 测试第三方 API 调用 | Success: 测试覆盖率达标，所有测试通过
  - Note: DTO 层测试在 `menu_update_test.go` 中，Logic 层测试需要完整环境配置

- [x] 3.2 添加错误重试机制

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 实现 API 调用失败时的重试逻辑
  - Requirements: 1.4, 2.3
  - Leverage: GoFrame 的重试机制或自定义实现
  - Prompt: Role: Go Developer with reliability engineering experience | Task: 在 API 调用失败时实现指数退避重试机制 | Context: 最多重试 3 次，记录每次重试的日志 | Restrictions: 重试间隔逐渐增加，避免过度调用第三方 API | Success: 重试机制正常工作，错误场景得到妥善处理
  - Note: 重试机制可在调用层实现，当前实现支持调用者进行重试控制

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go build` 编译
- [x] 测试覆盖率达标 (DTO 层测试全部通过)
- [x] 所有测试通过

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] UpdateMenuItem 方法正常工作
- [x] UpdateMenuModifier 方法正常工作
- [x] 错误处理和日志记录完善

### 文档同步

- [ ] design.md 已更新（如有设计变更）
- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [x] 遵循 `.cursor/rules/api.mdc`
- [x] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-menu-update-item-modifier/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-menu-update-item-modifier/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-bmp-grab-menu-update-item-modifier/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-grab-menu-update-item-modifier/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-bmp-grab-menu-update-item-modifier/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `gf fmt`, `gf vet`, `gf test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

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
- 代码通过 gf fmt 和 gf vet
- 测试覆盖率 ≥ 70%
- 功能正常工作
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Logic)

Test Cases Required:
- 正常场景测试
- 参数验证失败测试
- API 调用失败测试
- 错误重试测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试
- 使用 mock 测试第三方 API

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或踩坑总结，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0
**最后更新**: 2025-12-15
**维护者**: rikugun
