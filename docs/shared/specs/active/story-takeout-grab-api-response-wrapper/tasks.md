# grab-api-response-wrapper 任务分解

> 本文档定义 将 Grab 服务 API 响应格式统一为 takeout.ApiResponse 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6
**已完成**: 6
**进行中**: -
**完成率**: 100%

---

## Phase 1: 接口定义修改

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 grab.proto 文件

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
  - Purpose: 将两个方法的返回值改为 takeout.ApiResponse
  - Requirements: 1.1, 2.1
  - Leverage: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 中的实现方式
  - Prompt: Role: Protobuf Developer | Task: 修改 grab.proto 中的 CreateSelfServeJourney 和 GetShopProviderCfg 方法返回值 | Context: 参考 menu.proto 的实现，将自定义响应消息改为 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: protobuf 文件修改完成，语法正确

- [x] 1.2 重新生成 protobuf Go 代码

  - File: -
  - Purpose: 生成更新后的 gRPC Go 代码
  - Requirements: 1.1, 2.1
  - Leverage: 现有 Makefile: `ttpos-bmp/app/ttpos-takeout/Makefile`
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao`
  - Success: 代码生成成功，无编译错误

- [x] 1.3 验证代码编译通过

  - File: -
  - Purpose: 确保修改后的代码可以正常编译
  - Requirements: 1.1, 2.1
  - Leverage: -
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go build ./...`
  - Success: 编译通过，无错误

---

## Phase 2: 代码实现更新

### Controller 层

- [x] 2.1 修改 Grab Controller 响应格式

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab_controller.go`
  - Purpose: 适配新的 ApiResponse 返回格式
  - Requirements: 1.2, 2.2
  - Leverage: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/` 中的其他 controller 实现
  - Prompt: Role: Go Developer with gRPC expertise | Task: 修改 GrabController 的两个方法，将返回值改为 takeout.ApiResponse | Context: 使用 takeout.ApiSuccessWithData() 和 takeout.ApiError() 返回响应 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Controller 修改完成，响应格式正确

### Logic 层

- [x] 2.2 验证 Logic 层业务逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/`
  - Purpose: 确保业务逻辑返回的数据结构与新响应格式兼容
  - Requirements: 1.4, 2.4
  - Leverage: 现有 Logic 实现
  - Success: Logic 层无需修改，数据结构兼容

---

## Phase 3: 测试验证

- [x] 3.1 编写单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab_controller_test.go`
  - Purpose: 测试新的响应格式转换逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为修改后的 GrabController 编写单元测试 | Context: 测试成功响应和错误响应格式 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [x] 3.2 执行集成测试

  - File: -
  - Purpose: 验证端到端功能正常工作
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Success: 集成测试通过，API 响应格式正确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [x] 测试覆盖率达标
  - Controller: ≥ 70%
- [x] 所有测试通过

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] 验收标准已达成

### 文档同步

- [x] protobuf 文件变更已记录

### 规范遵循

- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [x] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [x] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-grab-api-response-wrapper/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-grab-api-response-wrapper/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-grab-api-response-wrapper/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-grab-api-response-wrapper/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-grab-api-response-wrapper/tasks.md)" | bc
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

### Go BMP 微服务开发

```
Role: Go Developer specializing in gRPC and GoFrame

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录
- gRPC 服务必须注册到 Nacos
- 使用统一的 ApiResponse 格式
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
- 响应格式验证测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 响应格式正确
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
**维护者**: rikugun
