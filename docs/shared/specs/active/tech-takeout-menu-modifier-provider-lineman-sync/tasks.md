# Menu Modifier Provider 多平台支持与 Lineman 状态同步 - 任务分解

> 本文档定义 Menu Modifier Provider 多平台支持的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 13  
**进行中**: -  
**完成率**: 72%

---

## Phase 1: Protobuf 修改和代码生成

- [x] 1.1 修改 Protobuf 定义

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 在 `UpdateMenuModifierReq` 中新增 `provider_name` 字段，支持多平台
  - Requirements: Requirement 1
  - Leverage: 现有 Protobuf 定义: `UpdateMenuItemReq` 中的 `provider_name` 字段（参考 [tech-takeout-menu-provider-lineman-sync](../tech-takeout-menu-provider-lineman-sync/design.md)）
  - Changes:
    ```protobuf
    message UpdateMenuModifierReq {
      string merchant_id = 1;
      string modifier_id = 2;
      string modifier_name = 3;
      optional int64 price = 4;
      optional string available_status = 5;
      optional bool is_free = 6;
      repeated AdvancedPricing advanced_pricings = 7;
      string request_id = 8;
      optional string provider_name = 9;  // 新增字段：平台名称 (grab, lineman)
    }
    ```
  - Success: Protobuf 定义修改完成，字段编号正确（9）

- [x] 1.2 生成 Protobuf 代码

  - File: -
  - Purpose: 生成 Go 代码
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make pb`
  - Success: ✅ 代码生成成功，`UpdateMenuModifierReq` 中包含 `ProviderName *string` 字段，Controller 已启用该字段

---

## Phase 2: DTO 层

- [x] 2.1 创建 Lineman 修饰符状态 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/modifier_status.go`
  - Purpose: 定义 Lineman 修饰符状态更新请求/响应 DTO
  - Requirements: Requirement 3
  - Leverage: 现有 Lineman DTO: `internal/model/dto/lineman/menu_status.go`
  - Prompt: Role: Go Developer with DTO expertise | Task: 创建 Lineman 修饰符状态 DTO | Context: 参考 menu_status.go 的实现模式，定义 ModifierStatusUpdateReq, ModifierPropertyValue（使用 int 类型的 Status 字段）, ModifierStatusUpdateResp | Restrictions: Status 字段必须为 int 类型（1, 2, 3） | Success: DTO 创建成功，JSON 标签正确
  - Success: `modifier_status.go` 创建完成，包含 `ModifierStatusUpdateReq`, `ModifierPropertyValue`, `ModifierStatusUpdateResp`

---

## Phase 3: Client 层

- [x] 3.1 创建 ModifierStatusClient

  - File: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/modifier_status_client.go`
  - Purpose: 实现调用 Lineman 修饰符状态更新 API 的客户端
  - Requirements: Requirement 4
  - Leverage: 
    - 现有 Client: `internal/client/lineman/menu_status_client.go`（MenuStatusClient）
    - 认证逻辑: `internal/client/lineman/token_client.go`（OAuthTokenClient）
    - 重试机制: `internal/client/lineman/retry.go`（WithRetry）
    - 配置加载: `internal/client/lineman/config.go`（MustConfig）
  - Prompt: Role: Go Developer with HTTP Client expertise | Task: 创建 ModifierStatusClient，实现 UpdateModifierStatus 和 UpdateModifierStatusWithRetry 方法 | Context: 复用 OAuthTokenClient.GetAuthorizationHeader() 获取 Bearer Token，使用 ghttp.Client 发送 PUT 请求到 /v1/partners/{partnerId}/stores/{storeId}/menu/property/values/status，实现重试机制 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，复用现有认证逻辑，HTTP 200 为成功，status=ok 为成功 | Success: Client 创建成功，API 调用正确，重试机制正常
  - Success: `modifier_status_client.go` 创建完成，包含 `ModifierStatusClient`, `UpdateModifierStatus`, `UpdateModifierStatusWithRetry`, `getAuthorizationHeader`

- [x] 3.2 编写 ModifierStatusClient 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/modifier_status_client_test.go`
  - Purpose: 测试 ModifierStatusClient 的 API 调用逻辑
  - Requirements: Requirement 4
  - Leverage: 现有测试: `internal/client/lineman/menu_status_client_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 ModifierStatusClient 编写单元测试 | Context: 测试成功场景、API 错误场景、HTTP 错误场景、重试机制 | Restrictions: 使用 Mock 客户端，避免真实 API 调用 | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试创建完成，覆盖 4+ 测试用例

---

## Phase 4: Logic 层

- [x] 4.1 创建状态映射函数

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/modifier_status.go`
  - Purpose: 实现 TTPOS 状态（string）到 Lineman 状态（int）的映射函数
  - Requirements: Requirement 2
  - Leverage: 现有映射函数: `internal/logic/lineman/menu_status.go`（MapStatusToLineman，string → string）
  - Prompt: Role: Go Developer | Task: 实现 MapStatusToLinemanModifier(ttposStatus string) (int, error) 函数 | Context: 映射规则：AVAILABLE → 1, UNAVAILABLE → 3, SOLD_OUT_TODAY → 2，其他返回错误 | Restrictions: 严格校验枚举值，未知状态返回明确错误 | Success: 映射函数实现完成，边界情况处理正确
  - Success: `MapStatusToLinemanModifier` 函数创建完成，映射逻辑正确

- [x] 4.2 实现 ModifierStatusLogic

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/modifier_status.go`
  - Purpose: 实现 Lineman 修饰符状态业务逻辑
  - Requirements: Requirement 6
  - Leverage: 
    - 现有 Logic: `internal/logic/lineman/menu_status.go`（MenuStatusLogic）
    - Task 3.1 的 ModifierStatusClient
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 ModifierStatusLogic，包含 UpdateModifierStatus 方法 | Context: 参数校验（storeId, modifierId, status），调用 ModifierStatusClient，检查响应 status=ok | Restrictions: status 必须为 1, 2, 或 3 | Success: Logic 实现完成，业务逻辑正确
  - Success: `ModifierStatusLogic` 创建完成，包含 `UpdateModifierStatus` 方法

- [x] 4.3 编写状态映射函数单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/modifier_status_test.go`
  - Purpose: 测试状态映射函数
  - Requirements: Requirement 2
  - Leverage: Task 4.1 的映射函数
  - Prompt: Role: QA Engineer | Task: 为 MapStatusToLinemanModifier 编写单元测试 | Context: 测试 AVAILABLE, UNAVAILABLE, SOLD_OUT_TODAY, 空字符串, 无效状态 | Restrictions: 覆盖所有映射场景 | Success: 测试覆盖率 100%，所有测试通过
  - Success: 测试创建完成，覆盖 5+ 测试用例

- [x] 4.4 编写 ModifierStatusLogic 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/modifier_status_test.go`
  - Purpose: 测试 ModifierStatusLogic 业务逻辑
  - Requirements: Requirement 6
  - Leverage: Task 4.2 的 ModifierStatusLogic
  - Prompt: Role: QA Engineer | Task: 为 ModifierStatusLogic 编写单元测试 | Context: 测试成功场景、空 storeId、空 modifierId、无效 status | Restrictions: 使用 Mock Client | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试创建完成，覆盖 4+ 测试用例

---

## Phase 5: Controller 层

- [x] 5.1 实现平台路由逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 在 `UpdateMenuModifier` 方法中根据 `provider_name` 进行平台路由
  - Requirements: Requirement 5
  - Leverage: 
    - 现有 Controller: `internal/controller/rpc/menu/menu.go`（UpdateMenuItem 的平台路由逻辑）
    - Task 4.2 的 ModifierStatusLogic
    - Task 4.1 的 MapStatusToLinemanModifier
  - Prompt: Role: Go Developer with gRPC expertise | Task: 修改 UpdateMenuModifier 方法，增加平台路由逻辑 | Context: 获取 provider_name（默认 grab），switch 路由到 handleGrabModifierUpdate 或 handleLinemanModifierUpdate | Restrictions: 保持 Grab 现有逻辑不变（向后兼容） | Success: 平台路由实现完成，Grab 逻辑不受影响
  - Changes:
    - 在 `UpdateMenuModifier` 方法开头获取 `provider_name`
    - 使用 `switch` 语句路由到 `handleGrabModifierUpdate` 或 `handleLinemanModifierUpdate`
    - 默认路由到 `handleGrabModifierUpdate`（Grab 逻辑）
  - Success: 平台路由逻辑实现完成

- [x] 5.2 实现 Lineman 字段校验函数

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 `validateLinemanModifierRequest` 字段校验函数
  - Requirements: Requirement 5
  - Leverage: Task 5.1 的 Controller 代码
  - Prompt: Role: Go Developer | Task: 实现 validateLinemanModifierRequest 字段校验函数 | Context: 检查 available_status 不为空，检查 price, is_free, advanced_pricings 字段不存在 | Restrictions: Lineman 仅支持 available_status 字段 | Success: 字段校验函数实现完成，边界情况处理正确
  - Success: `validateLinemanModifierRequest` 函数创建完成

- [x] 5.3 实现 handleLinemanModifierUpdate 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 Lineman 修饰符更新处理逻辑
  - Requirements: Requirement 5, Requirement 6
  - Leverage: 
    - Task 5.2 的 validateLinemanModifierRequest
    - Task 4.1 的 MapStatusToLinemanModifier
    - Task 4.2 的 ModifierStatusLogic
  - Prompt: Role: Go Developer with gRPC expertise | Task: 实现 handleLinemanModifierUpdate 方法 | Context: 1) 调用 validateLinemanModifierRequest 校验字段，2) 调用 MapStatusToLinemanModifier 映射状态，3) 调用 ModifierStatusLogic.UpdateModifierStatus | Restrictions: 错误处理使用 gerror.Wrap | Success: 方法实现完成，逻辑正确
  - Success: `handleLinemanModifierUpdate` 方法创建完成

- [ ] 5.4 编写 Controller 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu_test.go`
  - Purpose: 测试 Controller 的平台路由和 Lineman 处理逻辑
  - Requirements: Requirement 5
  - Leverage: 现有测试: `internal/controller/rpc/menu/menu_test.go`
  - Prompt: Role: QA Engineer | Task: 为 UpdateMenuModifier 编写单元测试 | Context: 测试 Lineman 成功场景、字段校验失败场景、Grab 向后兼容场景 | Restrictions: 使用 Mock Logic | Success: 测试覆盖率 ≥ 70%，所有测试通过
  - Success: 测试创建完成，覆盖 3+ 测试用例

---

## Phase 6: 集成测试

- [ ] 6.1 编写端到端集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/menu_modifier_lineman_test.go`（新建）
  - Purpose: 测试完整的 Lineman 修饰符更新流程
  - Requirements: 所有功能需求
  - Leverage: 
    - Task 1.2 的 gRPC 接口
    - 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 编写端到端集成测试 | Context: 构造 UpdateMenuModifierReq（provider_name=lineman），调用 gRPC 接口，验证响应成功，验证 Lineman API 被正确调用（使用 Mock 或实际 API） | Restrictions: 测试真实用户场景 | Success: 集成测试通过
  - Success: 集成测试创建完成，测试通过

---

## Phase 7: 文档和代码检查

- [ ] 7.1 更新 API 文档

  - File: `docs/shared/api/ttpos-takeout-api.md`（如存在）
  - Purpose: 更新 gRPC API 文档
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档更新完成

- [x] 7.2 代码格式化和静态检查

  - File: -
  - Purpose: 确保代码质量
  - Requirements: 代码质量要求
  - Command: 
    ```bash
    cd ttpos-bmp/app/ttpos-takeout
    go fmt ./...
    go vet ./...
    ```
  - Success: ✅ 代码格式化完成，静态检查通过（已验证修改的包）

- [ ] 7.3 运行所有测试

  - File: -
  - Purpose: 确保所有测试通过
  - Requirements: 测试要求
  - Command: 
    ```bash
    cd ttpos-bmp/app/ttpos-takeout
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    ```
  - Success: 所有测试通过，覆盖率达标

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Go 代码通过 `golangci-lint` 检查
- [ ] 测试覆盖率达标
  - Client: ≥ 80%
  - Logic: ≥ 80%
  - Controller: ≥ 70%
- [ ] 所有测试通过（包括单元测试和集成测试）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成（参考 requirements.md 的 AC）

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/tech-takeout-menu-modifier-provider-lineman-sync/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-menu-modifier-provider-lineman-sync/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/tech-takeout-menu-modifier-provider-lineman-sync/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-menu-modifier-provider-lineman-sync/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/tech-takeout-menu-modifier-provider-lineman-sync/tasks.md)" | bc
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

### GoFrame 后端开发

```
Role: Go Developer specializing in GoFrame framework

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录（自动生成）
- DTO 手动编写在 internal/model/dto/ 中
- 使用 GoFrame 的 ghttp.Client 发送 HTTP 请求
- 使用 gerror.Wrap/gerror.Wrapf 包装错误
- 使用 g.Log() 记录日志

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80% (Client, Logic) 或 ≥ 70% (Controller)
```

### Protobuf 开发

```
Role: Protobuf Developer

Task: {具体任务描述}

Context:
- Current file: manifest/protobuf/menu/menu.proto
- Leverage code: {现有 Protobuf 定义}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 字段编号不能冲突
- 新增字段使用 optional
- 字段注释说明默认值和可选值

Success Criteria:
- Protobuf 定义正确
- 字段编号无冲突
- 注释完整
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80% (Client, Logic) 或 ≥ 70% (Controller)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- Mock 外部依赖

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试
- 使用 Mock 避免真实 API 调用

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
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-13  
**维护者**: rikugun
