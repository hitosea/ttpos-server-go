# grabfood-create-self-serve-journey-grpc 设计文档

> 本文档定义 grabfood-create-self-serve-journey-grpc 的技术设计和实现方案。

## 📋 概述

本需求旨在为 Main 模块提供 GrabFood 自助点餐链接生成能力。实现将部署在 `ttpos-bmp` 微服务集群中的 `ttpos-takeout` 服务内，通过 gRPC 暴露给 Main 模块或其他服务调用。

核心流程：
1.  Main 模块调用 gRPC 接口 `CreateSelfServeJourney`。
2.  `ttpos-takeout` 服务接收请求，根据 Brand/Store 标识从配置中获取 Grab 凭证。
3.  服务调用 GrabFood 官方 Go SDK (`grabfood-api-sdk-go`) 的 `CreateSelfServeJourney` 接口。
4.  处理响应并映射错误，返回统一格式结果。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- **微服务架构**: 遵循 `ttpos-bmp` 的微服务架构，服务注册到 Nacos。
- **代码结构**: 遵循 Controller (RPC) -> Logic -> DAO/SDK 分层。
- **SDK 集成**: Grab SDK 封装在 Logic 层或独立的 Adapter 层，不直接在 Controller 调用。
- **错误处理**: 统一错误码映射，使用 `errors` 包包装错误。

### API 设计规范 (api.mdc)

- **gRPC 定义**: 使用 Protobuf v3 定义接口，方法名 PascalCase，字段 snake_case。
- **幂等性**: 接口支持幂等键 `request_id`。

---

## 🔄 代码复用分析

### 可复用的现有组件

- **Grab SDK Wrapper**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/sdk_wrapper.go` (如果存在，复用现有 SDK 初始化和调用逻辑)。
- **配置管理**: `ttpos-bmp/app/ttpos-takeout/internal/config` (复用现有配置加载机制)。
- **Grab 业务逻辑**: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab` (复用现有 Grab 相关逻辑，如 Token 管理)。

### 集成点

- **Proto 定义**: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` (新增 RPC 方法)。
- **Takeout 服务**: `ttpos-bmp/app/ttpos-takeout` (业务实现载体)。

---

## 🏗️ 架构设计

### 分层设计原则

**Go BMP 架构**:

```
RPC Controller (Handler)
  ↓
Logic 层 (Grab Service)
  ↓
SDK Wrapper / Grab SDK
```

### 模块划分

#### Go BMP 模块 (`ttpos-bmp/app/ttpos-takeout`)

- **RPC Controller**: `internal/controller/rpc/grab_v1_create_self_serve_journey.go`
  - 负责参数校验、Trace ID 提取、调用 Logic。
- **Logic 层**: `internal/logic/grab/self_serve_journey.go`
  - 负责业务逻辑：配置获取、SDK 调用、错误映射、重试策略。
- **SDK Wrapper**: `internal/logic/grab/sdk_wrapper.go`
  - 封装 `grabfood-api-sdk-go` 的具体调用。
- **Protobuf**: `manifest/protobuf/grab/grab.proto`
  - 定义 `CreateSelfServeJourney` 接口。

---

## 🔌 API 设计

### gRPC API

#### Protobuf 定义

```protobuf
// ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto

service Grab {
  // ... existing rpcs ...

  // 创建自助激活链接
  rpc CreateSelfServeJourney (CreateSelfServeJourneyReq) returns (CreateSelfServeJourneyResp);
}

message CreateSelfServeJourneyReq {
  string provider_name = 1;      // 外卖渠道, grab,lineman
  string shop_uuid = 2;      // 门店 ID (Grab Store ID)
  string request_id = 3; // 追踪 ID,可选
  // 无需传入 environment，由后端配置决定
}

message CreateSelfServeJourneyResp {
    string provider_name = 1;      // 外卖渠道, grab,lineman
  string self_serve_url = 2; // 自助点餐链接
  string request_id = 3;       // 追踪 ID
}
```

---

## 🧩 组件和接口

### Logic 层

#### Logic 接口

```go
// ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go (interface definition)

type sGrab struct {
    // ...
}

func (s *sGrab) CreateSelfServeJourney(ctx context.Context, req *grab.CreateSelfServeJourneyReq) (*grab.CreateSelfServeJourneyResp, error) {
    // 1. 获取渠道配置 (包含 Environment, Credentials)
    // 2. 初始化或获取 SDK Client
    // 3. 调用 SDK
    // 4. 处理响应与错误
}
```

### 错误处理

- **授权错误**: 映射为 `CodeAuthFailed`。
- **网络超时**: 映射为 `CodeNetworkError`，触发重试。
- **业务错误**: 映射为 `CodeBizError`，透传 Grab 错误信息。

---

## 🔒 安全设计

- **凭证管理**: Grab `ClientID` / `ClientSecret` 存储在 takeout 服务配置或数据库中，不通过 gRPC 请求传输。
- **环境隔离**: 根据配置自动选择 Sandbox 或 Production 环境，避免生产环境误操作。

---

## 🧪 测试策略

### 单元测试

- **Logic 层**: Mock SDK Wrapper，测试不同 Grab 响应（成功、失败、超时）下的处理逻辑。覆盖率 ≥ 70%。

### 集成测试

- **Sandbox 环境**: 使用 Grab Sandbox 凭证进行真实调用，验证链接生成是否成功。

---

## 📚 实现清单

### Phase 1: 协议与接口定义

- [ ] 更新 `grab.proto` 定义 gRPC 接口
- [ ] 生成 gRPC 代码 (`make dao` 或 `make proto`)

### Phase 2: 核心实现

- [ ] 实现 SDK Wrapper 方法 `CreateSelfServeJourney`
- [ ] 实现 Logic 层 `CreateSelfServeJourney`
- [ ] 实现 RPC Controller
- [ ] 注册 RPC 服务

### Phase 3: 测试与集成

- [ ] 编写 Logic 单元测试
- [ ] 编写集成测试
- [ ] 部署到测试环境验证

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-11.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-11  
**作者**: rikugun  
**审核者**: -
