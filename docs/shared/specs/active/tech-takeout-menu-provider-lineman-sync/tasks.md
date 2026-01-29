# Menu Provider 多平台支持与 Lineman 状态同步 任务分解

> 本文档定义 Menu Provider 多平台支持与 Lineman 状态同步功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 12  
**进行中**: Phase 4 (测试)  
**完成率**: 67%

---

## Phase 1: Protobuf 扩展和代码生成

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（Requirement 1, 2, etc.）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）
- **Success**: 成功标准

---

- [x] 1.1 修改 Protobuf 定义，增加 provider_name 字段

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 在 UpdateMenuItemReq 中增加 provider_name 字段，支持平台标识
  - Requirements: Requirement 1
  - Leverage: 现有 Protobuf 文件: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Task Details:
    - 在 `UpdateMenuItemReq` 消息中增加字段: `optional string provider_name = 9;`
    - 添加中文注释: `// 平台名称: grab (默认), lineman，为未来平台预留扩展`
    - 更新 `available_status` 字段注释，增加 `SOLD_OUT_TODAY` 说明
  - Prompt:
    ```
    Role: Protobuf Developer
    
    Task: 在 UpdateMenuItemReq 中增加 provider_name 字段
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto
    - Leverage: 现有 UpdateMenuItemReq 消息定义
    - Requirements: 需要支持多平台标识（grab, lineman）
    
    Restrictions:
    - 字段序号使用 9（避免冲突）
    - 使用 optional 修饰
    - 添加完整的中文注释
    - 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc
    
    Success Criteria:
    - provider_name 字段正确添加
    - 注释完整且清晰
    - Proto 文件格式正确
    ```
  - Success: `provider_name` 字段正确添加，注释完整

- [x] 1.2 执行 make pb 重新生成代码

  - File: -
  - Purpose: 根据修改后的 Protobuf 文件生成 Go 代码
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的 Protobuf 文件
  - Command:
    ```bash
    cd ttpos-bmp/app/ttpos-takeout
    make proto
    ```
  - Success: 代码生成成功，`api/menu/menu.pb.go` 包含 `ProviderName` 字段

- [x] 1.3 验证生成的代码

  - File: `ttpos-bmp/app/ttpos-takeout/api/menu/menu.pb.go`
  - Purpose: 确保生成的 Go 代码正确且可编译
  - Requirements: Requirement 1
  - Leverage: Task 1.2 生成的代码
  - Task Details:
    - 检查 `UpdateMenuItemReq` 结构体包含 `ProviderName *string` 字段
    - 检查字段序号为 9
    - 执行 `go build` 确保可编译
  - Success: 生成的代码正确，可编译通过

---

## Phase 2: DTO 和状态映射实现

- [x] 2.1 创建 Lineman Menu Status DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/menu_status.go`
  - Purpose: 定义 Lineman 菜单状态更新的请求和响应 DTO
  - Requirements: Requirement 2, 4
  - Leverage: 现有 Lineman DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/base.go`
  - Task Details:
    - 创建 `MenuStatusUpdateReq` 结构体（包含 MenuItems 数组）
    - 创建 `MenuItem` 结构体（包含 ID 和 MenuStatus）
    - 创建 `MenuStatusUpdateResp` 结构体（包含 Status, Code, Message）
  - Prompt:
    ```
    Role: Go Developer with DTO expertise
    
    Task: 创建 Lineman Menu Status DTO
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/menu_status.go
    - Leverage: ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/base.go
    - Requirements: 定义 Lineman API 的请求和响应格式
    
    Request Format:
    {
      "menuItems": [
        {
          "id": "partner-item-id",
          "menuStatus": "SUSPENDED"
        }
      ]
    }
    
    Response Format:
    {
      "status": "ok",
      "code": "SUCCESS",
      "message": "Menu status updated"
    }
    
    Restrictions:
    - 使用 JSON 标签
    - 添加中文注释
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Success Criteria:
    - DTO 定义完整
    - 字段映射正确
    - 可编译通过
    ```
  - Success: DTO 定义完整，字段映射正确

- [x] 2.2 实现状态映射函数

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_status.go`
  - Purpose: 实现 TTPOS 状态到 Lineman 状态的映射逻辑
  - Requirements: Requirement 2
  - Leverage: 现有 Lineman Logic: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync.go`
  - Task Details:
    - 创建 `MapStatusToLineman` 函数
    - 实现状态映射：AVAILABLE → AVAILABLE, UNAVAILABLE → SUSPENDED, SOLD_OUT_TODAY → SOLD_OUT_TODAY
    - 对不支持的状态（UNAVAILABLEHIDE）返回错误
  - Prompt:
    ```
    Role: Go Developer with business logic expertise
    
    Task: 实现状态映射函数 MapStatusToLineman
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_status.go
    - Leverage: ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync.go
    - Requirements: 映射 TTPOS 状态到 Lineman 状态
    
    Mapping Table:
    | TTPOS 状态     | Lineman 状态    |
    |----------------|-----------------|
    | AVAILABLE      | AVAILABLE       |
    | UNAVAILABLE    | SUSPENDED       |
    | SOLD_OUT_TODAY | SOLD_OUT_TODAY  |
    | UNAVAILABLEHIDE| 不支持（返回错误）|
    
    Restrictions:
    - 使用 gerror 处理错误
    - 添加中文注释
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Success Criteria:
    - 所有状态映射正确
    - 不支持的状态返回明确错误
    - 可编译通过
    ```
  - Success: 状态映射函数实现完整，所有场景覆盖

- [x] 2.3 编写状态映射单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_status_test.go`
  - Purpose: 测试状态映射逻辑的正确性
  - Requirements: Requirement 2
  - Leverage: Task 2.2 的状态映射函数
  - Task Details:
    - 测试 AVAILABLE → AVAILABLE
    - 测试 UNAVAILABLE → SUSPENDED
    - 测试 SOLD_OUT_TODAY → SOLD_OUT_TODAY
    - 测试 UNAVAILABLEHIDE 返回错误
    - 测试空状态返回错误
  - Success: 所有测试用例通过，覆盖率 ≥ 80%

---

## Phase 3: Lineman API 客户端实现

- [x] 3.1 创建 Lineman Menu Status Client

  - File: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_status_client.go`
  - Purpose: 实现调用 Lineman 菜单状态更新 API 的客户端
  - Requirements: Requirement 4
  - Leverage: 现有 Lineman Client: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_sync_client.go`
  - Task Details:
    - 创建 `MenuStatusClient` 结构体
    - 实现 `UpdateMenuStatus` 方法
    - 调用 `PUT /v1/partners/{partnerId}/stores/{storeId}/menu/items/status` 接口
    - 复用现有的 `AuthClient` 获取 Access Token
    - 添加请求/响应日志
  - Prompt:
    ```
    Role: Go Developer with HTTP Client expertise
    
    Task: 创建 Lineman Menu Status Client
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_status_client.go
    - Leverage: ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_sync_client.go
    - Requirements: 调用 Lineman API 更新菜单状态
    - API Doc: https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633
    
    API Details:
    - Method: PUT
    - URL: /v1/partners/{partnerId}/stores/{storeId}/menu/items/status
    - Headers: Authorization: Bearer {token}, Content-Type: application/json
    - Request Body: MenuStatusUpdateReq
    - Response Body: MenuStatusUpdateResp
    
    Restrictions:
    - 使用 GoFrame 的 ghttp.Client
    - 复用 AuthClient 获取 Token
    - 使用 gerror 处理错误
    - 记录请求和响应日志
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Success Criteria:
    - Client 实现完整
    - API 调用成功
    - 错误处理正确
    - 日志记录完整
    ```
  - Success: Client 实现完整，API 调用成功

- [x] 3.2 实现 Lineman Menu Status Logic

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_status.go` (扩展)
  - Purpose: 实现 Lineman 菜单状态更新的业务逻辑
  - Requirements: Requirement 4
  - Leverage: Task 3.1 的 Client，Task 2.2 的状态映射
  - Task Details:
    - 创建 `MenuStatusLogic` 结构体
    - 实现 `UpdateMenuStatus` 方法
    - 调用 `MenuStatusClient`
    - 添加参数校验（最多 100 个商品）
    - 添加错误处理和日志
  - Success: Logic 实现完整，业务逻辑正确

- [x] 3.3 编写 Client 和 Logic 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_status_client_test.go`, `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_status_test.go` (扩展)
  - Purpose: 测试 Client 和 Logic 的正确性
  - Requirements: Requirement 4
  - Leverage: Task 3.1 的 Client，Task 3.2 的 Logic
  - Task Details:
    - Mock Lineman API 成功响应（HTTP 200）
    - Mock Lineman API 失败响应（HTTP 400/500）
    - 测试参数校验（menuItems 为空、超过 100 个）
    - 测试错误处理
  - Success: 所有测试用例通过，覆盖率 ≥ 80%

---

## Phase 4: Controller 层集成

- [x] 4.1 实现字段校验函数

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 校验 Lineman 请求只包含 available_status 字段
  - Requirements: Requirement 3
  - Leverage: 现有 Menu Controller: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Task Details:
    - 创建 `validateLinemanRequest` 方法
    - 检查 price, max_stock, advanced_pricings, purchasabilities 字段
    - 如果包含不支持的字段，返回明确错误
    - 检查 available_status 不能为空
  - Prompt:
    ```
    Role: Go Developer with validation expertise
    
    Task: 实现 Lineman 请求字段校验函数
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go
    - Leverage: 现有 Menu Controller
    - Requirements: Lineman 平台仅支持更新 available_status 字段
    
    Validation Rules:
    - price 不能存在（返回错误："Lineman 平台仅支持更新 available_status 字段，不支持 price 字段"）
    - max_stock 不能存在
    - advanced_pricings 不能存在
    - purchasabilities 不能存在
    - available_status 必须存在且不为空
    
    Restrictions:
    - 使用 gerror 处理错误
    - 错误信息清晰明确
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Success Criteria:
    - 校验函数实现完整
    - 所有不支持字段正确识别
    - 错误信息准确
    ```
  - Success: 字段校验函数实现完整，所有场景覆盖

- [x] 4.2 实现平台路由分发逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 根据 provider_name 将请求路由到对应平台处理逻辑
  - Requirements: Requirement 5
  - Leverage: 现有 Menu Controller，Task 4.1 的字段校验
  - Task Details:
    - 修改 `UpdateMenuItem` 方法，增加平台路由逻辑
    - 获取 `provider_name`，默认为 `grab`
    - 使用 switch-case 分发请求：grab → `handleGrabUpdate`, lineman → `handleLinemanUpdate`
    - 对不支持的平台返回错误
  - Success: 平台路由逻辑实现完整，Grab 逻辑保持不变

- [x] 4.3 实现 handleLinemanUpdate 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 Lineman 平台的菜单更新处理逻辑
  - Requirements: Requirement 2, 3, 4, 5
  - Leverage: Task 4.1 的字段校验，Task 2.2 的状态映射，Task 3.2 的 Logic
  - Task Details:
    - 调用 `validateLinemanRequest` 校验字段
    - 调用 `mapToLinemanStatus` 映射状态
    - 调用 `MenuStatusLogic.UpdateMenuStatus` 更新状态
    - 返回统一响应格式
  - Success: Lineman 处理逻辑实现完整，所有功能正常

- [ ] 4.4 编写 Controller 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu_test.go`
  - Purpose: 测试 Controller 层的平台路由和字段校验
  - Requirements: Requirement 3, 5
  - Leverage: Task 4.1-4.3 的 Controller 实现
  - Task Details:
    - 测试 `provider_name=grab` 路由到 Grab 逻辑
    - 测试 `provider_name=lineman` 路由到 Lineman 逻辑
    - 测试未指定 `provider_name` 默认路由到 Grab
    - 测试不支持的 `provider_name` 返回错误
    - 测试 Lineman 字段校验（所有不支持字段场景）
    - 测试状态映射（所有状态场景）
  - Success: 所有测试用例通过，覆盖率 ≥ 80%

---

## Phase 5: 集成测试和优化

- [ ] 5.1 集成测试 - Grab 平台向后兼容

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/menu_test.go`
  - Purpose: 确保 Grab 平台功能未受影响
  - Requirements: Requirement 5（向后兼容）
  - Leverage: 现有 Grab 集成测试
  - Task Details:
    - 使用现有 Grab 测试用例
    - 调用 `UpdateMenuItem` gRPC 接口（`provider_name=grab` 或未指定）
    - 验证 Grab API 调用成功
    - 验证响应格式正确
  - Success: 所有现有 Grab 测试通过

- [ ] 5.2 集成测试 - Lineman 平台端到端

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/menu_lineman_test.go`
  - Purpose: 测试 Lineman 平台的端到端流程
  - Requirements: Requirement 2, 3, 4, 5
  - Leverage: Task 4.3 的 Controller 实现
  - Task Details:
    - 调用 `UpdateMenuItem` gRPC 接口（`provider_name=lineman`）
    - 测试字段校验（传入不支持的字段，验证返回错误）
    - 测试状态映射（传入 AVAILABLE, UNAVAILABLE, SOLD_OUT_TODAY）
    - 验证 Lineman API 调用成功（可以 Mock）
    - 验证响应格式正确
  - Success: Lineman 端到端测试通过

- [ ] 5.3 实现 HTTP 重试机制

  - File: `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_status_client.go`
  - Purpose: 在 API 调用失败时自动重试
  - Requirements: 非功能需求 - 可靠性
  - Leverage: Task 3.1 的 Client，GoFrame 的 ghttp.Client 重试功能
  - Task Details:
    - 配置 HTTP Client 重试次数（最多 3 次）
    - 配置重试间隔（指数退避：1s, 2s, 4s）
    - 记录重试日志
  - Success: 重试机制实现完整，测试通过

- [ ] 5.4 添加监控和日志

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`, `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/menu_status_client.go`
  - Purpose: 添加监控指标和详细日志，便于调试和监控
  - Requirements: 非功能需求 - 可靠性
  - Leverage: GoFrame 的 glog 日志库
  - Task Details:
    - 记录平台路由信息（provider_name）
    - 记录状态映射信息（TTPOS 状态 → Lineman 状态）
    - 记录 Lineman API 请求和响应（包含 request_id）
    - 记录错误详情（包含错误码和错误信息）
    - 敏感信息（如 Token）不记录
  - Success: 日志记录完整，便于调试

---

## Phase 6: 文档更新

- [ ] 6.1 更新 API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 更新 API 文档，说明新增的 provider_name 参数
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Task Details:
    - 在 `UpdateMenuItem` 接口文档中增加 `provider_name` 参数说明
    - 说明 Lineman 平台的限制（仅支持 available_status）
    - 说明状态映射规则
    - 添加示例（Grab 和 Lineman）
  - Success: API 文档更新完整，准确

- [ ] 6.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录本次功能更新
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Task Details:
    - 在 `v2.14.0` 版本下添加本次功能
    - 说明新增功能：多平台支持、Lineman 状态同步
    - 说明向后兼容性
  - Success: CHANGELOG 更新完整

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（≥ 80%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足（Requirement 1-6）
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] Protobuf 注释完整

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/tech-takeout-menu-provider-lineman-sync/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-menu-provider-lineman-sync/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/tech-takeout-menu-provider-lineman-sync/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-menu-provider-lineman-sync/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/tech-takeout-menu-provider-lineman-sync/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的技术设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame 2.x

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录
- 使用 gerror 处理错误
- 添加中文注释
- Controller → Logic → Client 分层清晰

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### Protobuf 开发

```
Role: Protobuf Developer

Task: {具体任务描述}

Context:
- File: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- 字段序号不能冲突
- 使用 optional 修饰可选字段
- 添加完整的中文注释

Success Criteria:
- Protobuf 定义正确
- 注释完整
- 可生成 Go 代码
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- Mock 第三方 API 测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充 - 开发完成后记录]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-13  
**维护者**: rikugun
