# 在 /assistant/desk/ping 接口中返回已选国旗ID 任务分解

> 本文档定义「在 /assistant/desk/ping 接口中返回已选国旗ID」功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 4  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 响应结构扩展

- [ ] 1.1 修改 DeskPing 响应结构体

  - File: `main/app/dto/resp/desk.go`
  - Purpose: 在 `DeskPing` 结构体中添加 `NationalityUuid` 字段
  - Requirements: 1.1
  - Leverage: 现有响应结构: `main/app/dto/resp/desk.go`，参考其他字段定义
  - Prompt: Role: Go Developer | Task: 在 DeskPing 结构体中添加 NationalityUuid uint64 字段，JSON 标签为 nationality_uuid，添加注释说明 | Context: 字段位置放在 OrderRemark 字段之后，类型为 uint64，默认值为 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段命名使用驼峰命名 | Success: 字段添加成功，JSON 标签正确，注释清晰

---

## Phase 2: Service 层实现

- [ ] 2.1 修改 GetDeskPing 方法添加国籍UUID赋值逻辑

  - File: `main/app/service/desk.go`
  - Purpose: 在 `GetDeskPing` 方法中从 `desk.SaleBill.NationalityUuid` 读取并赋值给响应
  - Requirements: 1.2, 1.3, 1.4
  - Leverage: 现有 Service 方法: `main/app/service/desk.go`，参考其他字段赋值逻辑
  - Prompt: Role: Go Developer with Service Layer expertise | Task: 在 GetDeskPing 方法中添加国籍UUID赋值逻辑 | Context: 在获取 desk.SaleBill 后，检查 SaleBill 是否为 nil，如果不为 nil，则赋值 res.NationalityUuid = desk.SaleBill.NationalityUuid，如果为 nil 或 NationalityUuid 为 0，则保持默认值 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，添加空值检查避免 panic，不使用 panic，返回 error | Success: 赋值逻辑正确，空值检查完善，代码通过 go fmt 和 go vet

---

## Phase 3: 文档更新

- [ ] 3.1 更新 Swagger 文档

  - File: `main/docs/swagger.yaml`
  - Purpose: 在 `/assistant/desk/ping` 接口的响应定义中添加 `nationality_uuid` 字段说明
  - Requirements: 1.5
  - Leverage: 现有 Swagger 文档: `main/docs/swagger.yaml`，参考其他字段定义
  - Prompt: Role: Technical Writer | Task: 更新 Swagger 文档，在 DeskPing 响应模型中添加 nationality_uuid 字段 | Context: 字段类型为 integer(uint64)，说明为"国籍UUID（0=未设置）"，示例值为 0 或大于 0 的 UUID | Restrictions: 遵循 Swagger 规范，字段说明清晰 | Success: Swagger 文档更新成功，字段定义正确

---

## Phase 4: 测试验证

- [ ] 4.1 编写 API 测试

  - File: `main/app/api/v1/assistant/assistant_desk_test.go` 或新建测试文件
  - Purpose: 测试 `/assistant/desk/ping` 接口返回的 `nationality_uuid` 字段
  - Requirements: 所有功能需求
  - Leverage: 现有 API 测试: `main/app/api/v1/assistant/`，参考其他接口测试
  - Prompt: Role: QA Engineer with Go API testing expertise | Task: 为 /assistant/desk/ping 接口编写测试，验证 nationality_uuid 字段 | Context: 测试场景包括：1) 未开台场景（返回 0），2) 已开台但未设置国籍（返回 0），3) 已设置国籍（返回对应 UUID），4) 设置国籍后轮询（返回更新后的 UUID） | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用测试框架 | Success: 所有测试场景通过，测试覆盖率达标

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（Swagger）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-assistant-desk-ping-nationality/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-assistant-desk-ping-nationality/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-assistant-desk-ping-nationality/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-assistant-desk-ping-nationality/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-assistant-desk-ping-nationality/tasks.md)" | bc
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

### Go 后端开发

```
Role: Go Developer specializing in Service Layer

Task: 在 GetDeskPing 方法中添加国籍UUID赋值逻辑

Context:
- Current file: main/app/service/desk.go
- Leverage code: main/app/service/desk.go (现有 GetDeskPing 方法)
- Requirements: 1.2, 1.3, 1.4 - 从 desk.SaleBill.NationalityUuid 读取并赋值
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc

Restrictions:
- Service 只依赖其他 Service 接口
- 不使用 panic，返回 error
- 添加空值检查避免 panic
- 使用 errors.WithMessage 包装错误

Success Criteria:
- 赋值逻辑正确
- 空值检查完善
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-25  
**维护者**: TTPOS Team

