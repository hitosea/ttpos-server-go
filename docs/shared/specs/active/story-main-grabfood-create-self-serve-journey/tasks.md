# grabfood-create-self-serve-journey-grpc 任务分解

> 本文档定义 grabfood-create-self-serve-journey-grpc 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8
**已完成**: 5
**进行中**: -
**完成率**: 62.5%

---

## Phase 1: 协议与接口定义

- [x] 1.1 定义 Protobuf 接口
  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
  - Purpose: 在 Grab 服务中新增 `CreateSelfServeJourney` RPC 方法定义
  - Requirements: 1.1, 1.2
  - Leverage: 现有 `grab/grab.proto`
  - Prompt: Role: gRPC Developer | Task: 更新 grab.proto，添加 CreateSelfServeJourney 接口 | Context: 包含 provider_name, shop_uuid, request_id，返回 provider_name, self_serve_url, request_id | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: proto 文件定义正确

- [x] 1.2 生成 gRPC 代码
  - File: -
  - Purpose: 生成 Go 客户端和服务端存根代码
  - Requirements: 1.1
  - Leverage: Makefile
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao` (或相应的 proto 生成命令)
  - Success: `api/` 目录下生成新的 go 文件

---

## Phase 2: 核心实现 (Logic & SDK)

- [x] 2.1 实现 SDK Wrapper
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/sdk_wrapper.go`
  - Purpose: 封装 Grab SDK 的 `CreateSelfServeJourney` 调用
  - Requirements: 2.1
  - Leverage: 现有 `sdk_wrapper.go`
  - Prompt: Role: Go Developer | Task: 在 SdkWrapper 中添加 CreateSelfServeJourney 方法 | Context: 调用 grabfood-api-sdk-go，处理请求参数组装 | Restrictions: 遵循 go-bmp.mdc | Success: Wrapper 方法实现完成

- [x] 2.2 实现 Logic 业务逻辑
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/self_serve_journey.go`
  - Purpose: 实现核心业务流程：配置获取、环境选择、SDK调用、错误处理
  - Requirements: 1.1, 2.1, 2.2, 3.2, 3.3
  - Leverage: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go`
  - Prompt: Role: Go Developer | Task: 实现 CreateSelfServeJourney logic | Context: 从配置获取凭证（区分沙箱/生产），调用 Wrapper，处理错误映射 | Restrictions: 不依赖 Controller，返回 error | Success: 业务逻辑实现完成

- [x] 2.3 实现 RPC Controller
  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab_v1_create_self_serve_journey.go`
  - Purpose: 实现 gRPC Handler，负责参数校验和调用 Logic
  - Requirements: 1.1
  - Leverage: 现有 RPC Controller
  - Prompt: Role: Go Developer | Task: 实现 Grab.CreateSelfServeJourney RPC handler | Context: 校验 provider_name, shop_uuid，传递 request_id，调用 Logic | Restrictions: 遵循 go-bmp.mdc | Success: Controller 实现完成

- [x] 2.4 注册 RPC 服务
  - File: `ttpos-bmp/app/ttpos-takeout/internal/cmd/cmd.go` (或相关注册文件)
  - Purpose: 确保新方法被注册
  - Requirements: 1.1
  - Leverage: 现有注册逻辑
  - Success: 服务启动无报错

---

## Phase 3: 测试与集成

- [ ] 3.1 编写 Logic 单元测试
  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/self_serve_journey_test.go`
  - Purpose: 验证业务逻辑，覆盖成功、失败、超时场景
  - Requirements: 2.1, 2.2
  - Prompt: Role: QA Engineer | Task: 编写单元测试 | Context: Mock SDK Wrapper，测试环境选择逻辑 | Success: 覆盖率 >= 70%

- [ ] 3.2 编写集成测试
  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/grab_self_serve_test.go`
  - Purpose: 验证端到端调用（Sandbox）
  - Requirements: 1.1, 1.2
  - Success: Sandbox 调用成功返回链接

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-11.md`
