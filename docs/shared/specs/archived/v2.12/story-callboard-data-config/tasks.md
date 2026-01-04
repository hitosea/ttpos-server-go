# 叫号系统返回配置信息 任务分解

> 本文档定义 叫号系统返回配置信息 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 4  
**已完成**: 3  
**进行中**: -  
**完成率**: 75%

---

## Phase 1: DTO 扩展

- [x] 1.1 扩展 QueueDataResp 响应结构

  - File: `main/app/dto/resp/callboard.go`
  - Purpose: 在 `QueueDataResp` 结构中新增配置信息字段（必返字段）
  - Requirements: 1.1
  - Leverage: 现有 `QueueDataResp` 结构: `main/app/dto/resp/callboard.go`，参考 `DeviceItem` 结构中的配置字段定义
  - Prompt: Role: Go Developer | Task: 扩展 QueueDataResp 结构，新增 name、background_image_url、timeout_limit、voice_call_enabled、call_count 字段 | Context: 字段为必返字段（不使用 omitempty），TimeoutLimit 和 VoiceCallEnabled 使用指针类型，其他字段使用值类型 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段命名使用驼峰命名 | Success: 响应结构扩展成功，字段定义正确

---

## Phase 2: Service 层实现

- [x] 2.1 修改 GetQueueData 方法读取配置信息

  - File: `main/app/service/callboard/service.go`
  - Purpose: 在 `GetQueueData` 方法中从 `bindInfo` 读取配置信息并填充到响应结构
  - Requirements: 1.2
  - Leverage: 现有 `GetQueueData` 方法: `main/app/service/callboard/service.go:211`，`DeviceBindInfo` 结构已包含配置字段: `main/app/service/callboard/service.go:422`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 GetQueueData 方法，从 bindInfo 读取配置信息并填充到 QueueDataResp | Context: bindInfo 已包含 Name、BackgroundImageUrl、TimeoutLimit、VoiceCallEnabled、CallCount 字段，需要设置默认值 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 方法修改成功，配置信息正确填充

- [x] 2.2 实现默认值逻辑

  - File: `main/app/service/callboard/service.go`
  - Purpose: 实现配置信息缺失时的默认值处理逻辑
  - Requirements: 1.3
  - Leverage: 现有默认值处理: `main/app/service/callboard/service.go:310-326`（GetDeviceList 方法中的默认值处理）
  - Prompt: Role: Go Developer | Task: 在 GetQueueData 方法中实现默认值逻辑 | Context: name 为空时设置为 "WALLACE"，background_image_url 为空时设置为空字符串，timeout_limit 为 nil 时设置为 0，voice_call_enabled 为 nil 时设置为 false，call_count 为 0 时设置为 1 | Restrictions: 默认值必须明确且一致，参考 GetDeviceList 方法的实现 | Success: 默认值逻辑实现正确，所有边界情况已处理

---

## Phase 3: 测试和验证

- [ ] 3.1 编写 Service 单元测试

  - File: `main/app/service/callboard/service_test.go`
  - Purpose: 测试 GetQueueData 方法的配置信息返回和默认值逻辑
  - Requirements: 1.4, 1.5
  - Leverage: 现有测试: `main/app/service/callboard/service_test.go`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetQueueData 方法编写单元测试，覆盖配置信息返回和默认值逻辑 | Context: 测试配置信息存在时的返回，测试配置信息为空时的默认值处理，测试配置信息为 nil 时的默认值处理，测试配置信息读取失败时的错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试覆盖率 ≥ 70%，所有测试通过，边界情况已覆盖

- [ ] 3.2 API 集成测试

  - File: `main/app/api/v1/callboard/handler_test.go`（或集成测试文件）
  - Purpose: 测试 `/callboard/data` 接口返回配置信息
  - Requirements: 所有功能需求
  - Leverage: 现有 API 测试: `main/app/api/v1/callboard/`（如有）
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 /callboard/data 接口编写集成测试，验证配置字段返回 | Context: 测试接口响应包含配置字段，测试默认值返回，测试配置信息更新后的返回 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过，响应格式正确

---

## Phase 4: 文档更新

- [x] 4.1 更新 API 文档（Swagger 注释）

  - File: `main/app/api/v1/callboard/handler.go`
  - Purpose: 更新 `/callboard/data` 接口的 Swagger 注释，说明新增配置字段
  - Requirements: 文档要求
  - Leverage: 现有 Swagger 注释: `main/app/api/v1/callboard/handler.go:80-90`
  - Prompt: Role: Technical Writer | Task: 更新 GetQueueData 方法的 Swagger 注释，说明新增配置字段 | Context: 在 @Success 注释中说明新增字段，说明默认值逻辑 | Restrictions: 遵循 Swagger 注释规范 | Success: Swagger 注释更新完成，文档准确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - 接口响应包含所有配置字段（必返）
  - 默认值处理正确
  - 降级处理正确
  - 向后兼容

### 文档同步

- [ ] API 文档已更新（Swagger 注释）
- [ ] 代码注释已更新（默认值逻辑说明）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-callboard-data-config/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-callboard-data-config/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-callboard-data-config/tasks.md
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

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-11.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-11  
**维护者**: 后端开发组
