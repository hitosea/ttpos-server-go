# Grab 批量更新菜单 API 集成 任务分解

> 本文档定义 Grab 批量更新菜单 API 集成 的详细执行任务清单。

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

## Phase 1: Protobuf 定义和代码生成

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3, 4.1）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 定义 Protobuf 消息和服务方法

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 定义批量更新菜单的 gRPC 接口和消息结构
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: 现有 Protobuf 定义: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`（`UpdateMenuItemReq`, `AdvancedPricing`, `Purchasability`）
  - Prompt: Role: gRPC Developer | Task: 在 menu.proto 中新增 MenuEntity、BatchUpdateMenuReq、BatchUpdateMenuResp、MenuEntityError 消息定义，并在 MenuService 中添加 BatchUpdateMenu RPC 方法 | Context: 复用现有的 AdvancedPricing 和 Purchasability 消息，MenuEntity 包含 id、price、available_status、max_stock、advanced_pricings、purchasabilities 字段，BatchUpdateMenuReq 包含 merchant_id、field、menu_entities、request_id 字段 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，字段使用 snake_case，请求消息以 Req 结尾，响应消息以 Resp 结尾，RPC 方法返回 takeout.ApiResponse | Success: Protobuf 定义完成，字段类型正确，注释完整

- [x] 1.2 生成 Protobuf Go 代码

  - File: -
  - Purpose: 根据 Protobuf 定义生成 Go 代码
  - Requirements: 4.1, 4.5
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Success: Go 代码生成成功，`api/menu/menu.pb.go` 文件已更新

---

## Phase 2: DTO 定义和数据模型

- [x] 2.1 创建批量更新 DTO 定义

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`
  - Purpose: 定义批量更新菜单的 Go 数据传输对象
  - Requirements: 1.1, 1.4, 1.5, 2.1, 2.3, 3.1, 3.4
  - Leverage: 现有 DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`（`UpdateAdvancedPricingReq`, `UpdatePurchasabilityReq`），GrabFood SDK 结构: `github.com/grab/grabfood-api-sdk-go@v1.0.2`（`BatchUpdateMenuItem`, `MenuEntity`, `BatchUpdateMenuResponse`, `MenuEntityError`）
  - Prompt: Role: Go Developer specializing in DTO design | Task: 在 menu_update.go 中新增 BatchUpdateMenuReq、MenuEntity、BatchUpdateMenuResp、MenuEntityError 结构体 | Context: 参考 GrabFood SDK 的结构定义，BatchUpdateMenuReq 包含 MerchantID、Field、MenuEntities 字段，MenuEntity 包含 ID、Price、AvailableStatus、MaxStock、AdvancedPricings、Purchasabilities 字段，复用现有的 UpdateAdvancedPricingReq 和 UpdatePurchasabilityReq | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 json 标签和 v 标签（GoFrame 验证），Field 必须是 ITEM 或 MODIFIER，MenuEntities 数量必须在 1-100 之间 | Success: DTO 定义完整，验证标签正确，注释清晰

- [x] 2.2 新增菜单日志类型常量

  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - Purpose: 定义批量更新的日志类型常量
  - Requirements: 3.3
  - Leverage: 现有日志类型常量
  - Prompt: Role: Go Developer | Task: 新增 MenuSyncTypeBatchUpdateItem 和 MenuSyncTypeBatchUpdateModifier 常量 | Context: 常量值分别为 "BATCH_UPDATE_ITEM" 和 "BATCH_UPDATE_MODIFIER"，用于 menu_log 表的 sync_type 字段 | Restrictions: 遵循命名规范 | Success: 常量定义完成，命名清晰

---

## Phase 3: Service 层实现

- [x] 3.1 实现 Grab 服务的 BatchUpdateMenu 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go`
  - Purpose: 在 Grab 服务接口中新增批量更新菜单方法
  - Requirements: 1.1, 2.1
  - Leverage: 现有 Grab 服务: `ttpos-bmp/app/ttpos-takeout/internal/service/grab.go`（`UpdateMenuItem`, `UpdateMenuModifier`），GrabFood SDK: `github.com/grab/grabfood-api-sdk-go@v1.0.2/api_update_menu_record.go`（`BatchUpdateMenu` 方法）
  - Prompt: Role: Go Developer with API integration expertise | Task: 在 sGrab 结构体中实现 BatchUpdateMenu 方法，调用 GrabFood SDK 的 BatchUpdateMenu API | Context: 方法签名为 `func (s *sGrab) BatchUpdateMenu(ctx context.Context, merchantID string, req *grabfood.BatchUpdateMenuItem) (*grabfood.BatchUpdateMenuResponse, error)`，使用 s.client.UpdateMenuRecordAPI.BatchUpdateMenu 调用 SDK | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 包处理错误，记录详细日志（中文） | Success: 方法实现完成，SDK 调用正确，错误处理完善

---

## Phase 4: Logic 层实现

- [x] 4.1 实现批量更新菜单 Logic 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 实现批量更新菜单的核心业务逻辑
  - Requirements: 1.1-1.7, 2.1-2.6, 3.1-3.5
  - Leverage: 现有 Logic 实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`（`UpdateMenuItem`, `UpdateMenuModifier`），Task 2.1 的 DTO 定义，Task 3.1 的 Service 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 sGrabMenu 结构体中实现 BatchUpdateMenu 方法 | Context: 方法签名为 `func (s *sGrabMenu) BatchUpdateMenu(ctx context.Context, req *grabDto.BatchUpdateMenuReq) (*grabDto.BatchUpdateMenuResp, error)`，包含参数验证、SDK 请求构建、调用 Grab API、错误处理、日志记录 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，Logic 层返回类型不能是 takeout.ApiResponse，使用 gerror 包处理错误，日志使用中文 | Success: Logic 方法实现完整，业务逻辑正确，错误处理完善，日志记录清晰

- [x] 4.2 实现 DTO 到 SDK 结构体转换逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 实现 DTO MenuEntity 到 SDK MenuEntity 的转换
  - Requirements: 1.4, 1.5, 2.3
  - Leverage: 现有转换逻辑（`convertAdvancedPricings`, `convertPurchasabilities`），Task 4.1 的 BatchUpdateMenu 方法
  - Prompt: Role: Go Developer | Task: 在 BatchUpdateMenu 方法中实现 DTO MenuEntity 到 SDK MenuEntity 的转换逻辑 | Context: 遍历 req.MenuEntities，创建 SDK MenuEntity 对象，设置 ID、Price、AvailableStatus、MaxStock（仅商品）、AdvancedPricings、Purchasabilities（仅商品），复用现有的 convertAdvancedPricings 和 convertPurchasabilities 方法 | Restrictions: 修饰符不支持 maxStock 和 purchasabilities，需根据 req.Field 判断 | Success: 转换逻辑正确，字段映射完整

- [x] 4.3 实现批量更新日志记录方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 实现批量更新的日志记录方法
  - Requirements: 3.3, 3.4, 3.5
  - Leverage: 现有日志记录逻辑，Task 2.2 的日志类型常量
  - Prompt: Role: Go Developer | Task: 实现 logBatchUpdate 方法，记录批量更新日志到 menu_log 表 | Context: 方法签名为 `func (s *sGrabMenu) logBatchUpdate(ctx context.Context, merchantID, field string, count int, success bool, errMsg string)`，sync_type 格式为 "BATCH_UPDATE_{field}"（BATCH_UPDATE_ITEM 或 BATCH_UPDATE_MODIFIER），status 根据 success 判断，error_msg 存储详细错误信息 | Restrictions: 使用 dao.MenuLog.Ctx(ctx).Data(logDo).Insert() 插入日志 | Success: 日志记录方法实现完成，日志格式正确

---

## Phase 5: Controller 层实现

- [x] 5.1 实现 gRPC Controller 的 BatchUpdateMenu 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 gRPC 接口层，处理批量更新菜单请求
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.6
  - Leverage: 现有 Controller 实现: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`，Task 4.1 的 Logic 方法
  - Prompt: Role: Go Developer with gRPC expertise | Task: 在 Controller 结构体中实现 BatchUpdateMenu 方法 | Context: 方法签名为 `func (c *Controller) BatchUpdateMenu(ctx context.Context, req *api.BatchUpdateMenuReq) (*takeout.ApiResponse, error)`，包含参数验证（merchantID、field、menuEntities 数量）、Protobuf 到 DTO 转换、调用 Logic 层、DTO 到 Protobuf 转换、包装为 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc，Controller 层响应必须使用 takeout.ApiResponse 包装，使用 rpc.ApiError 和 rpc.ApiSuccessWithData 辅助方法，使用 anypb.New 包装业务数据 | Success: Controller 方法实现完整，参数验证正确，响应包装正确

- [x] 5.2 实现 Protobuf 到 DTO 转换逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 Protobuf MenuEntity 到 DTO MenuEntity 的转换
  - Requirements: 1.4, 1.5, 2.3, 4.1
  - Leverage: Task 5.1 的 Controller 方法
  - Prompt: Role: Go Developer | Task: 在 BatchUpdateMenu 方法中实现 Protobuf MenuEntity 到 DTO MenuEntity 的转换逻辑 | Context: 遍历 req.MenuEntities，创建 DTO MenuEntity 对象，设置 ID、Price、AvailableStatus、MaxStock、AdvancedPricings、Purchasabilities，处理可选字段（使用 entity.Price != nil 判断） | Restrictions: 字段映射要完整，可选字段处理要正确 | Success: 转换逻辑正确，字段映射完整

- [x] 5.3 实现 DTO 到 Protobuf 转换逻辑（响应）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 DTO BatchUpdateMenuResp 到 Protobuf BatchUpdateMenuResp 的转换
  - Requirements: 3.4, 4.2, 4.6
  - Leverage: Task 5.1 的 Controller 方法
  - Prompt: Role: Go Developer | Task: 在 BatchUpdateMenu 方法中实现 DTO BatchUpdateMenuResp 到 Protobuf BatchUpdateMenuResp 的转换逻辑 | Context: 创建 api.BatchUpdateMenuResp 对象，设置 MerchantId、Status、Errors，遍历 resp.Errors 转换为 api.MenuEntityError | Restrictions: 字段映射要完整 | Success: 转换逻辑正确，errors 数组转换正确

---

## Phase 6: 测试实现

**注**: 测试实现已略过，可在后续需要时补充

- [ ] 6.1 编写 Logic 层单元测试 - 成功场景（已略过）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试批量更新菜单成功场景
  - Requirements: 1.1, 2.1
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 BatchUpdateMenu 方法编写单元测试，测试批量更新商品和修饰符成功场景 | Context: 测试用例包括批量更新 10 个商品（status=success）、批量更新 10 个修饰符（status=success）、测试高级定价和购买能力转换 | Restrictions: 使用 mock 替代 Grab API 调用，测试覆盖率 ≥ 80% | Success: 测试用例完整，所有测试通过

- [ ] 6.2 编写 Logic 层单元测试 - 部分失败场景

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试批量更新部分失败场景
  - Requirements: 3.2, 3.4
  - Leverage: Task 6.1 的测试用例
  - Prompt: Role: QA Engineer | Task: 编写测试用例，测试批量更新部分失败（status=partial_success）和全部失败（status=fail）场景 | Context: 模拟 Grab API 返回 partial_success 和 fail 状态，验证 errors 数组是否正确返回，验证日志记录是否包含错误信息 | Restrictions: 使用 mock 替代 Grab API 调用 | Success: 测试用例完整，错误处理逻辑正确

- [ ] 6.3 编写 Logic 层单元测试 - 参数验证场景

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试参数验证逻辑
  - Requirements: 1.6, 2.5, 3.1
  - Leverage: Task 6.1 的测试用例
  - Prompt: Role: QA Engineer | Task: 编写测试用例，测试参数验证逻辑 | Context: 测试 merchantID 为空、field 不是 ITEM 或 MODIFIER、menuEntities 数量为 0、menuEntities 数量超过 100 等场景 | Restrictions: 验证错误消息是否正确 | Success: 测试用例完整，参数验证逻辑正确

- [ ] 6.4 编写 Controller 层集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu_test.go`
  - Purpose: 测试 gRPC 接口端到端流程
  - Requirements: 4.1-4.6
  - Leverage: 现有 Controller 测试
  - Prompt: Role: QA Automation Engineer | Task: 编写集成测试，测试 gRPC BatchUpdateMenu 接口 | Context: 测试成功场景、部分失败场景、参数验证场景，验证 ApiResponse 格式是否正确 | Restrictions: 使用 mock 替代 Grab API 调用 | Success: 集成测试通过，ApiResponse 格式正确

- [ ] 6.5 编写 SDK 集成测试（可选）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/grab_test.go`
  - Purpose: 测试与 GrabFood SDK 的集成
  - Requirements: 1.1, 2.1
  - Leverage: GrabFood SDK 文档和示例
  - Prompt: Role: Integration Test Engineer | Task: 编写 SDK 集成测试，验证与 GrabFood SDK 的正确集成 | Context: 测试 BatchUpdateMenu 方法调用 SDK 是否正确，测试 SDK 响应解析是否正确 | Restrictions: 可以使用 staging 环境或 mock 数据 | Success: SDK 集成测试通过

---

## Phase 7: 文档和优化

- [x] 7.1 更新 Proposal 状态

  - File: `docs/team/proposals/2025-12/v2.11.0-grab-batch-update-menu.md`
  - Purpose: 更新 Proposal 的状态和关联信息
  - Requirements: -
  - Leverage: -
  - Success: Proposal 状态更新为"已完成"，关联 Spec

- [x] 7.2 更新活动日志

  - File: `docs/team/activities/rikugun/2025-12/2025-12-24.md`（或当前日期）
  - Purpose: 记录任务完成情况
  - Requirements: -
  - Leverage: -
  - Success: 活动日志已更新

- [x] 7.3 代码审查和优化

  - File: -
  - Purpose: 进行代码审查和性能优化
  - Requirements: 性能要求
  - Leverage: `.cursor/rules/code-review.mdc`
  - Success: 代码审查通过，linter 错误已修复

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic 层: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - 批量更新 50 个商品响应时间 < 2 秒
  - 部分失败场景正确处理
  - 参数验证逻辑正确
  - 日志记录完整

### 文档同步

- [ ] Protobuf 注释完整
- [ ] 代码注释清晰（中文）
- [ ] Proposal 状态已更新
- [ ] 活动日志已更新

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`
- [ ] Controller 层响应使用 `takeout.ApiResponse` 包装
- [ ] Logic 层不返回 `takeout.ApiResponse`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-takeout-grab-batch-update-menu/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-takeout-grab-batch-update-menu/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-takeout-grab-batch-update-menu/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-takeout-grab-batch-update-menu/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-takeout-grab-batch-update-menu/tasks.md)" | bc
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

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc, ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录（自动生成）
- Controller 层响应必须使用 takeout.ApiResponse 包装
- Logic 层返回类型不能是 takeout.ApiResponse
- 使用 gerror 包处理错误，不使用 panic
- 日志使用中文描述

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80% (Logic 层)
```

### Protobuf 开发

```
Role: gRPC Developer

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- 请求消息以 Req 结尾
- 响应消息以 Resp 结尾
- 字段使用 snake_case
- 服务使用 PascalCase
- RPC 方法返回 takeout.ApiResponse

Success Criteria:
- {成功标准1}
- Protobuf 定义正确
- 注释完整
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80% (Logic 层)

Test Cases Required:
- 正常场景测试（批量更新成功）
- 异常场景测试（部分失败、全部失败）
- 边界条件测试（参数验证）
- 性能测试（响应时间 < 2 秒）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 使用 mock 替代 Grab API 调用
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标（≥ 80%）
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-23.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-23  
**维护者**: rikugun

