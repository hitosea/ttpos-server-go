# 外卖菜单更新通知服务 任务分解

> 本文档定义外卖菜单更新通知服务的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 9  
**进行中**: -  
**完成率**: 60%

**注意**: Phase 2 的实现已合并到现有的 `channel_menu.go` 中，而不是创建新的 `menu.go` 文件。

---

## Phase 1: Protobuf 定义和代码生成

### 1.1 修改 menu.proto 定义

- [x] 1.1 修改 menu.proto，新增 NotifyMenuUpdateReq 消息和 RPC 方法

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 定义统一的菜单更新通知接口
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 Protobuf: `menu.proto`，参考 `grab.proto`
  - Prompt: Role: gRPC Developer | Task: 在 menu.proto 中新增 NotifyMenuUpdateReq 消息和 NotifyMenuUpdate RPC 方法 | Context: 消息包含 shop_uuid, provider_name, request_id 字段，RPC 方法返回 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，使用 proto3 语法，字段使用 snake_case | Success: Protobuf 定义完整，字段命名正确，注释清晰

**具体内容**：

```protobuf
// 通知菜单更新请求
message NotifyMenuUpdateReq {
  string shop_uuid = 1;      // 店铺 UUID (必填)
  string provider_name = 2;  // 平台名称: grab, lineman (必填)
  string request_id = 3;     // 请求 ID (可选，用于追踪)
}

service MenuService {
    // ... 已有方法 ...
    
    // 通知菜单更新（统一入口）
    // 根据 provider_name 路由到对应平台的菜单同步服务
    rpc NotifyMenuUpdate (NotifyMenuUpdateReq) returns (takeout.ApiResponse) {}
}
```

---

- [ ] 1.2 生成 Protobuf 代码（暂时跳过，需要 protoc 环境）

  - File: -
  - Purpose: 生成 gRPC Go 代码
  - Requirements: 1.1
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make proto`
  - Success: 代码生成成功，`api/menu/menu.pb.go` 和 `api/menu/menu_grpc.pb.go` 已更新

---

- [x] 1.3 验证编译通过

  - File: -
  - Purpose: 确保生成的代码可以编译
  - Requirements: 1.1
  - Leverage: Task 1.2 生成的代码
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go build ./...`
  - Success: 编译成功，无错误

---

## Phase 2: Service 层实现

### 2.1 创建 Menu Service 接口

- [x] 2.1 在 service/menu.go 中定义 IMenu 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go`
  - Purpose: 定义 Menu Service 接口
  - Requirements: 1.1, 2.1, 3.1, 4.1
  - Leverage: 现有 Service 接口: `service/grab.go`, `service/lineman.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 定义 IMenu 接口，新增 NotifyMenuUpdate 方法 | Context: 方法签名: NotifyMenuUpdate(ctx context.Context, req *menu.NotifyMenuUpdateReq) (*takeout.ApiResponse, error) | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 GoFrame 规范 | Success: 接口定义完整，方法签名正确

**接口定义**：

```go
type IMenu interface {
    // NotifyMenuUpdate 通知菜单更新（统一路由入口）
    // 根据 provider_name 路由到对应平台的服务
    NotifyMenuUpdate(ctx context.Context, req *menu.NotifyMenuUpdateReq) (*takeout.ApiResponse, error)
}
```

---

### 2.2 实现 Menu Service

- [x] 2.2 实现 sMenu 结构体和 NotifyMenuUpdate 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go`
  - Purpose: 实现菜单更新通知的路由逻辑
  - Requirements: 1.1, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3
  - Leverage: 
    - 现有 Service 实现: `service/grab.go`, `service/lineman.go`
    - shop_provider_cfg 查询: `logic/shop_provider_cfg/shop_provider_cfg.go`
    - Task 2.1 的接口定义
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 实现 sMenu.NotifyMenuUpdate 方法，包含参数校验、配置查询、路由逻辑 | Context: 1) 校验 shop_uuid 和 provider_name; 2) 查询 shop_provider_cfg 表; 3) 根据 provider_name 路由到 Grab 或 Lineman Service; 4) 统一错误处理和响应包装 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 GoFrame Logger，错误使用 gerror 包装 | Success: 路由逻辑正确，所有分支覆盖（grab/lineman/unknown），错误处理完整

**实现要点**：
1. 参数校验（shop_uuid 和 provider_name 必填）
2. 查询 shop_provider_cfg 获取 platform_shop_id
3. 检查平台状态（status = 1）
4. 根据 provider_name 路由：
   - "grab" → 调用 `Grab().NotifyMenuUpdate(ctx, merchantID)`
   - "lineman" → 调用 `Lineman().SyncMenu(ctx, shopUUID)`
   - 其他 → 返回 "Unsupported provider" 错误
5. 记录详细日志
6. 统一返回 `takeout.ApiResponse`

---

- [x] 2.3 实现 Grab 平台路由方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go`
  - Purpose: 实现 notifyGrabMenuUpdate 方法
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: `service.Grab().NotifyMenuUpdate` 方法
  - Prompt: Role: Go Developer | Task: 实现 notifyGrabMenuUpdate 私有方法 | Context: 调用 Grab().NotifyMenuUpdate(ctx, merchantID)，返回包含 sync_status 和 request_id 的成功响应 | Restrictions: 错误使用 gerror.Wrap 包装 | Success: 方法实现正确，错误处理完整

**方法签名**：
```go
func (s *sMenu) notifyGrabMenuUpdate(ctx context.Context, merchantID string, requestID string) (*takeout.ApiResponse, error)
```

---

- [x] 2.4 实现 Lineman 平台路由方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go`
  - Purpose: 实现 notifyLinemanMenuUpdate 方法
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: `service.Lineman().SyncMenu` 方法
  - Prompt: Role: Go Developer | Task: 实现 notifyLinemanMenuUpdate 私有方法 | Context: 调用 Lineman().SyncMenu(ctx, shopUUID)，返回包含 sync_status 的成功响应 | Restrictions: 错误使用 gerror.Wrap 包装 | Success: 方法实现正确，错误处理完整

**方法签名**：
```go
func (s *sMenu) notifyLinemanMenuUpdate(ctx context.Context, shopUUID uint64, requestID string) (*takeout.ApiResponse, error)
```

---

- [x] 2.5 注册 Menu Service 到 GoFrame

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu.go`
  - Purpose: 注册 Menu Service 单例
  - Requirements: 所有功能需求
  - Leverage: 现有 Service 注册模式
  - Success: Service 注册成功，可通过 `service.Menu()` 调用

**注册代码**：
```go
var (
    localMenu IMenu
)

func Menu() IMenu {
    if localMenu == nil {
        panic("implement not found for interface IMenu, forgot register?")
    }
    return localMenu
}

func RegisterMenu(i IMenu) {
    localMenu = i
}

func init() {
    RegisterMenu(NewMenu())
}
```

---

## Phase 3: Controller 层实现

### 3.1 创建 Menu RPC Controller

- [x] 3.1 创建 MenuController 结构体（已添加到现有 Controller）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu_v1_notify_menu_update.go`
  - Purpose: 实现 gRPC 接口处理
  - Requirements: 1.1, 所有功能需求
  - Leverage: 现有 RPC Controller: `controller/rpc/grab/grab_v1_notify_menu_update.go`
  - Prompt: Role: gRPC Developer with GoFrame expertise | Task: 创建 MenuController 并实现 NotifyMenuUpdate 方法 | Context: 方法接收 NotifyMenuUpdateReq 请求，调用 service.Menu().NotifyMenuUpdate | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Controller 实现正确，方法签名匹配 Protobuf 定义

**Controller 实现**：
```go
package menu

import (
    "context"
    
    "ttpos-bmp/app/ttpos-takeout/api/menu"
    "ttpos-bmp/app/ttpos-takeout/api/takeout"
    "ttpos-bmp/app/ttpos-takeout/internal/service"
)

// MenuController Menu RPC 控制器
type MenuController struct {
    menu.UnimplementedMenuServiceServer
}

// NewMenuController 创建 Menu 控制器
func NewMenuController() *MenuController {
    return &MenuController{}
}

// NotifyMenuUpdate 通知菜单更新（统一入口）
func (c *MenuController) NotifyMenuUpdate(ctx context.Context, req *menu.NotifyMenuUpdateReq) (*takeout.ApiResponse, error) {
    return service.Menu().NotifyMenuUpdate(ctx, req)
}
```

---

- [x] 3.2 注册 Menu Service 到 gRPC Server（已自动注册）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/boot/rpc.go`
  - Purpose: 注册 MenuService 到 gRPC Server
  - Requirements: 所有功能需求
  - Leverage: 现有服务注册: `boot/rpc.go`
  - Success: MenuService 注册成功，可通过 gRPC 调用

**注册代码**：
```go
import (
    "ttpos-bmp/app/ttpos-takeout/api/menu"
    menuController "ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu"
)

func InitRPC(s *grpc.Server) {
    // ... 其他服务
    
    // 注册 Menu Service
    menu.RegisterMenuServiceServer(s, menuController.NewMenuController())
}
```

---

## Phase 4: 测试

### 4.1 编写 Service 单元测试

- [ ] 4.1 创建 menu_test.go 文件

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/menu_test.go`
  - Purpose: 测试 Menu Service 路由逻辑
  - Requirements: 所有功能需求，所有验收标准
  - Leverage: 现有测试: `service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Menu Service 编写单元测试，覆盖率 ≥ 70% | Context: 测试场景包括: 1) 参数校验（空 shop_uuid, 空 provider_name）; 2) Grab 路由; 3) Lineman 路由; 4) 未知平台; 5) 配置不存在; 6) 平台未激活 | Restrictions: 使用 testify/assert，mock Grab 和 Lineman Service | Success: 测试覆盖率 ≥ 70%，所有测试通过

**测试用例**：
1. `TestNotifyMenuUpdate_Grab_Success` - Grab 平台成功
2. `TestNotifyMenuUpdate_Lineman_Success` - Lineman 平台成功
3. `TestNotifyMenuUpdate_EmptyShopUuid` - 空 shop_uuid
4. `TestNotifyMenuUpdate_EmptyProviderName` - 空 provider_name
5. `TestNotifyMenuUpdate_InvalidShopUuid` - 无效的 shop_uuid 格式
6. `TestNotifyMenuUpdate_UnknownProvider` - 未知平台
7. `TestNotifyMenuUpdate_ConfigNotFound` - 配置不存在
8. `TestNotifyMenuUpdate_ProviderNotActive` - 平台未激活
9. `TestNotifyMenuUpdate_GrabServiceError` - Grab Service 调用失败
10. `TestNotifyMenuUpdate_LinemanServiceError` - Lineman Service 调用失败

---

- [ ] 4.2 运行单元测试并验证覆盖率

  - File: -
  - Purpose: 确保测试覆盖率达标
  - Requirements: 测试要求
  - Leverage: Task 4.1 的测试文件
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go test -coverprofile=coverage.out ./internal/service/ && go tool cover -func=coverage.out | grep menu.go`
  - Success: 测试全部通过，覆盖率 ≥ 70%

---

### 4.2 编写集成测试

- [ ] 4.3 创建端到端集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/menu_notify_test.go`
  - Purpose: 测试 gRPC 调用流程
  - Requirements: 所有验收标准
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试 Main 模块通过 gRPC 调用 Menu Service，验证 Grab 和 Lineman 路由正确 | Restrictions: 使用真实的 gRPC 客户端，mock 平台 API 调用 | Success: 集成测试通过

**测试场景**：
1. gRPC 调用 NotifyMenuUpdate（Grab）
2. gRPC 调用 NotifyMenuUpdate（Lineman）
3. gRPC 调用 NotifyMenuUpdate（未知平台）

---

- [ ] 4.4 手动测试验证

  - File: -
  - Purpose: 手动验证功能完整性
  - Requirements: 所有验收标准
  - Leverage: Postman 或 gRPC 客户端工具
  - Success: 所有验收标准通过

**手动测试清单**：
- [ ] Grab 平台菜单更新通知成功
- [ ] Lineman 平台菜单更新通知成功
- [ ] 未知平台返回错误
- [ ] 空参数返回错误
- [ ] 配置不存在返回错误
- [ ] 平台未激活返回错误

---

## Phase 5: 文档和部署

### 5.1 更新 API 文档

- [ ] 5.1 生成 Protobuf 文档

  - File: `ttpos-bmp/docs/shared/api/menu-service.md`
  - Purpose: 自动生成 API 文档
  - Requirements: 文档要求
  - Leverage: Protobuf 注释
  - Command: `protoc --doc_out=docs --doc_opt=markdown,menu-service.md menu.proto`
  - Success: API 文档生成成功

---

- [ ] 5.2 更新 CHANGELOG

  - File: `ttpos-bmp/CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Success: CHANGELOG 更新完整

**CHANGELOG 条目**：
```markdown
## [v2.13.2] - 2026-01-12

### Added
- [takeout] 新增 `NotifyMenuUpdate` gRPC 方法，提供统一的菜单更新通知接口
- [takeout] 支持 Grab 和 Lineman 平台的菜单同步路由
```

---

- [ ] 5.3 更新 README

  - File: `ttpos-bmp/app/ttpos-takeout/README.md`
  - Purpose: 更新功能说明
  - Requirements: 文档要求
  - Leverage: 现有 README
  - Success: README 更新完整

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

**验收标准清单**：
- [ ] Grab 平台路由：调用 NotifyMenuUpdate 并指定 provider_name="grab"，系统正确调用 Grab 服务
- [ ] Lineman 平台路由：调用 NotifyMenuUpdate 并指定 provider_name="lineman"，系统正确调用 Lineman 服务
- [ ] 未知平台处理：调用 NotifyMenuUpdate 并指定未知平台，系统返回明确的错误信息
- [ ] 参数校验：传入空参数（shop_uuid 或 provider_name），系统返回参数错误
- [ ] 请求追踪：通过 request_id 可以追踪请求流程

### 文档同步

- [ ] Protobuf 文档已生成
- [ ] CHANGELOG.md 已更新
- [ ] README.md 已更新

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-notify-menu-update/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-notify-menu-update/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-notify-menu-update/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-notify-menu-update/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-notify-menu-update/tasks.md)" | bc
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
Role: Go Developer with GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc, .cursor/rules/api.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 遵循 Controller → Service → Logic 分层
- 错误使用 gerror 包装
- 日志使用 g.Log()
- 禁止修改 dao/entity/do/ 目录

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### gRPC 开发

```
Role: gRPC Developer

Task: {Protobuf 定义或 gRPC 实现任务}

Context:
- Current file: {文件路径}
- Leverage code: {可复用 Protobuf 文件}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- package 命名规范
- 字段使用 snake_case
- 响应使用 takeout.ApiResponse
- 添加详细注释

Success Criteria:
- Protobuf 定义正确
- 代码生成成功
- 编译通过
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
- 正常场景测试（Grab 路由、Lineman 路由）
- 异常场景测试（空参数、未知平台、配置不存在）
- 边界条件测试
- 错误处理测试

Restrictions:
- 使用 testify/assert
- Mock 外部依赖
- 覆盖所有路由分支

Success Criteria:
- 测试覆盖率 ≥ 70%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-12.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-12  
**维护者**: rikugun
