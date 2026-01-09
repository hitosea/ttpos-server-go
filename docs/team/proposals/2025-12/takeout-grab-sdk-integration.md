# Takeout 模块集成 GrabFood 官方 SDK 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-08   |
| **目标版本** | - |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | story-takeout-grab-integration |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-bmp/app/ttpos-takeout` 模块中的 Grab API 集成采用**自实现方案**，涉及两个方向的 API：

**1. 调用 Grab 的 API（Outbound）** - `internal/logic/grab/client.go`
**2. 被 Grab 调用的 API（Inbound/Webhook）** - `api/grab/v1/` + `model/dto/grab/`

存在以下问题：

1. **维护成本高**：需要手动跟踪 Grab API 变更，自行更新请求/响应结构
2. **API 覆盖不完整**：仅实现了核心功能（OAuth、订单、门店、菜单通知），缺少 Campaign、DineIn、Membership 等高级功能
3. **DTO 同步风险**：自定义的 `model/dto/grab/` 结构（如 `SubmitOrderRequest`、`OrderStateRequest`）可能与 Grab 官方 Schema 不一致
4. **Webhook Model 重复定义**：`dto/grab/order.go` 等文件手动定义了 Webhook 请求结构，与 SDK 提供的 Model 重复
5. **重复造轮子**：Grab 已发布官方 Go SDK（[github.com/grab/grabfood-api-sdk-go](https://github.com/grab/grabfood-api-sdk-go)），包含完整类型定义和 API 封装

### 业务价值

- **降低维护成本**：官方 SDK 由 Grab 团队维护，API 变更时自动适配
- **提升开发效率**：直接使用官方 Model，减少 DTO 转换代码
- **增强功能覆盖**：可快速接入 Campaign（营销活动）、DineIn（堂食）等新功能
- **提高可靠性**：官方 SDK 经过测试验证，减少集成错误风险

### 目标用户

- [x] 商户管理员（通过 Shop 端管理 Grab 门店）
- [x] 收银员（POS 端接收/处理 Grab 订单）
- [x] 厨房人员（KDS 端查看 Grab 订单）
- [x] 其他: 后端开发工程师

---

## 💡 解决方案概述

### 方案描述

引入 Grab 官方 SDK `github.com/grab/grabfood-api-sdk-go`，**渐进式替换**现有自实现方案，覆盖**双向 API**：

#### 1. Outbound（调用 Grab API）
- 使用 SDK 的 `APIClient` 替换 `client.go` 中的自实现
- 复用 SDK 内置的 OAuth Token 管理（可选保留 Redis 缓存策略）

#### 2. Inbound（被 Grab 调用的 Webhook）
- 使用 SDK 的 Model 定义（如 `SubmitOrderRequest`、`OrderStateRequest`）替换 `dto/grab/` 下的自定义结构
- 保留现有 `auth.go` 中的签名验证逻辑（SDK 不提供 Webhook 验证）
- GoFrame 的 API 定义（`api/grab/v1/`）保持不变，仅调整内部使用的 DTO

#### 实施阶段
1. **Phase 1**：引入 SDK 依赖，新增 SDK Wrapper 层
2. **Phase 2**：迁移 Outbound 功能（OAuth、订单操作、门店管理）到 SDK
3. **Phase 3**：迁移 Inbound DTO（用 SDK Model 替换 `dto/grab/` 自定义结构）
4. **Phase 4**：扩展新功能（Campaign、DineIn、Menu 批量更新等）

### 核心功能点

1. **SDK 集成**：添加 `github.com/grab/grabfood-api-sdk-go` 依赖
2. **Outbound Client**：封装 SDK Client，统一配置管理（环境切换、Token 缓存）
3. **Outbound 迁移**：将 `client.go` 中的 API 调用迁移到 SDK 实现
4. **Inbound DTO 替换**：使用 SDK 的 Webhook Model 替换自定义 DTO
5. **签名验证保留**：`auth.go` 中的 HMAC-SHA256 验证逻辑保留
6. **新功能启用**：接入 SDK 提供的 Campaign、Menu 批量更新等 API

### 影响范围

**涉及终端**：
- [x] POS 收银端（Grab 订单接收/状态更新）
- [x] Shop 商家管理端（门店状态管理）
- [x] KDS 厨显端（订单展示）
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型（DTO 可能简化）
- [x] 业务逻辑（client.go 重构）
- [x] 第三方集成（核心变更）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [x] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 3-5 天（分阶段实施）
- **预估 SP**: 5-8（待技术评审确认）

**分阶段工作量**：
| 阶段 | 工作内容 | 预估时间 |
|------|----------|----------|
| Phase 1 | SDK 引入、Wrapper 层搭建 | 0.5 天 |
| Phase 2 | 现有功能迁移（OAuth、订单、门店） | 1-2 天 |
| Phase 3 | 测试验证、DTO 清理 | 1 天 |
| Phase 4 | 新功能接入（可选） | 1-2 天 |

### 风险识别

**潜在风险**：
1. **SDK 版本兼容性**：官方 SDK 版本较新（v1.0.2），需验证与当前 Go 1.23 的兼容性
2. **Token 管理差异**：SDK 内置 Token 管理可能与现有 Redis 缓存策略不同
3. **响应结构变化**：SDK Model 结构可能与现有业务代码期望的字段不完全匹配

**缓解措施**：
1. 本地验证 SDK 依赖兼容性，确认无冲突
2. 封装 Wrapper 层统一 Token 获取逻辑，保持 Redis 缓存策略
3. 分阶段迁移，每个 API 单独测试验证后再上线

---

## 🔗 相关资源

### 参考需求

- 官方 SDK: [github.com/grab/grabfood-api-sdk-go](https://github.com/grab/grabfood-api-sdk-go)
- SDK 文档: [developer.grab.com](https://developer.grab.com)

### 相关文档

- 现有实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
- 现有 DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/`
- 关联 Spec: `docs/shared/specs/archived/v2.12/story-takeout-grab-integration/`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |  |           |
| 技术负责人   |  |           |
| 开发代表     | rikugun |           |
| 测试代表     |  |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审]
```

**下一步行动**：

- [x] 创建 Spec：[task-takeout-grab-sdk-integration](../../../shared/specs/archived/v2.12/task-takeout-grab-sdk-integration/requirements.md)
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### 现有实现与 SDK 功能对照

#### Outbound（调用 Grab API）

| 功能 | 现有实现 | SDK 支持 | 备注 |
|------|----------|----------|------|
| OAuth Token | ✅ `client.go` | ✅ `GetOauthGrabAPI` | SDK 内置 |
| 接受/拒绝订单 | ✅ | ✅ `AcceptRejectOrderAPI` | |
| 取消订单 | ✅ | ✅ `CancelOrderAPI` | |
| 标记订单就绪 | ✅ | ✅ `MarkOrderReadyAPI` | |
| 更新配送状态 | ✅ | ✅ `UpdateDeliveryStateAPI` | |
| 门店暂停/恢复 | ✅ | ✅ `PauseStoreAPI` | |
| 获取门店状态 | ✅ | ✅ `GetStoreStatusAPI` | |
| 获取营业时间 | ✅ | ✅ `GetStoreHourAPI` | |
| 菜单通知 | ✅ | ✅ `UpdateMenuNotificationAPI` | |
| 订单列表 | ❌ | ✅ `ListOrdersAPI` | **新增** |
| 菜单批量更新 | ❌ | ✅ `BatchUpdateMenu` | **新增** |
| 营销活动管理 | ❌ | ✅ `CreateCampaignAPI` 等 | **新增** |
| 堂食优惠券 | ❌ | ✅ `GetDineinVoucherAPI` 等 | **新增** |

#### Inbound（被 Grab 调用的 Webhook）

| Webhook 端点 | 现有 DTO | SDK Model | 迁移方案 |
|-------------|----------|-----------|----------|
| `POST /partner/orders` (提交订单) | ✅ `dto/grab/order.go` → `SubmitOrderRequest` | ✅ `SubmitOrderRequest` | 用 SDK Model 替换 |
| `PUT /partner/orders/state` (状态变更) | ✅ `dto/grab/order.go` → `OrderStateRequest` | ✅ `OrderStateRequest` | 用 SDK Model 替换 |
| `GET /partner/menu` (菜单拉取) | ✅ `dto/grab/menu.go` → `GetMenuResponse` | ✅ `GetMenuNewResponse` | 用 SDK Model 替换 |
| `POST /partner/menu/sync/state` (菜单同步) | ✅ 自定义解析 | ✅ `MenuSyncWebhookRequest` | 用 SDK Model 替换 |
| `POST /oauth/partner` (Partner Token) | ✅ `api/v1/oauth_partner_webhook.go` | ✅ `PartnerOauthRequest/Response` | 用 SDK Model 替换 |
| `POST /integration/status` (集成状态) | ✅ 自定义 | ✅ `PushIntegrationStatusWebhookRequest` | 用 SDK Model 替换 |
| Webhook 签名验证 | ✅ `auth.go` | ❌ 不提供 | **保留现有实现** |

### SDK 使用示例

```go
import grabfood "github.com/grab/grabfood-api-sdk-go"

// 配置 SDK
config := grabfood.NewConfiguration()
apiClient := grabfood.NewAPIClient(config)
ctx := context.WithValue(context.Background(), grabfood.ContextServerIndex, grabfood.StgEnv)

// 获取 OAuth Token
authResp, _, _ := apiClient.GetOauthGrabAPI.GetOauthGrab(ctx).
    GrabOauthRequest(*grabfood.NewGrabOauthRequest(
        clientID, clientSecret, "client_credentials", "food.partner_api",
    )).Execute()

// 获取门店营业时间
authorization := "Bearer " + *authResp.AccessToken
resp, _, _ := apiClient.GetStoreHourAPI.GetStoreHour(ctx, merchantID).
    Authorization(authorization).Execute()
```

### User Story（初稿）

**作为** 后端开发工程师  
**我想** 使用 Grab 官方 SDK 进行 API 集成  
**以便于** 减少维护成本，快速接入新功能，提高代码可靠性

### AC 验收标准（初稿）

1. **WHEN** 调用 Grab API（订单、门店等）**THEN** 系统 **SHALL** 通过官方 SDK 发起请求
2. **WHEN** 需要新增 Campaign/DineIn 功能 **THEN** 开发者 **SHALL** 直接使用 SDK API
3. **WHEN** SDK 升级 **THEN** 系统 **SHALL** 通过 `go get -u` 获取最新 API 支持
4. **IF** 现有业务测试通过 **THEN** 迁移后 **SHALL** 保持功能一致性

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**维护者**: rikugun
