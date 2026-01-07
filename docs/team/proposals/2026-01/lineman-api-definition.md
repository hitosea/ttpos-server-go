# LINE MAN API 定义 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2026-01-07   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

TTPOS 外卖系统需要与 LINE MAN 平台对接，接收 LINE MAN 的订单通知和管理请求。目前 `ttpos-bmp/app/ttpos-takeout` 模块缺少完整的 LINE MAN Webhook API 定义，需要基于 LINE MAN 官方文档创建标准的 RESTful API 定义（使用 GoFrame 框架）。

### 业务价值

- 支持泰国市场主流外卖平台 LINE MAN 的订单接收
- 实现订单自动化接收和处理，减少人工操作
- 接收 LINE MAN 的菜单同步通知，及时响应同步结果
- 支持 LINE MAN 触发的菜单同步请求
- 为后续扩展其他外卖平台（如 GrabFood）奠定基础

### 目标用户

- [x] 商户管理员（配置 LINE MAN 对接）
- [x] 厨房人员（接收和处理订单）
- [x] 系统运维人员（监控对接状态）
- [ ] 收银员
- [ ] 顾客
- [x] 其他: 外卖平台运营人员

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-bmp/app/ttpos-takeout/api/lineman/v1` 目录下创建完整的 LINE MAN Webhook API 定义（RESTful 风格），包括：

1. **OAuth 认证 API**：接收 LINE MAN 的认证请求，返回访问令牌
2. **订单创建 Webhook API**：接收 LINE MAN 的订单创建通知
3. **订单状态更新 Webhook API**：接收 LINE MAN 的订单状态更新通知
4. **订单更新 Webhook API**：接收 LINE MAN 的订单更新通知
5. **菜单同步通知 Webhook API**：接收 LINE MAN 的菜单同步结果通知
6. **菜单同步触发 Webhook API**：接收 LINE MAN 的菜单同步触发请求

所有 API 定义严格遵循 GoFrame 开发规范：
- 请求结构体以 `Req` 结尾
- 响应结构体以 `Resp` 结尾
- 使用 Go struct tag 定义验证规则和 JSON 映射
- 返回统一的响应格式

### 核心功能点

1. **OAuth 认证服务**
   - 接收 LINE MAN 的认证请求
   - 实现 `client_credentials` 授权类型验证
   - 返回访问令牌、令牌类型、有效期

2. **订单 Webhook 接收**
   - 接收订单创建（Place Order）
   - 接收订单状态更新（Order Status Update）
   - 接收订单更新通知（Order Update Notification）
   - 解析订单商品、属性、备注等详细信息
   - 返回统一的成功/失败响应

3. **菜单同步 Webhook 接收**
   - 接收菜单同步结果通知（Menu Sync Notification）
   - 接收菜单同步触发请求（Trigger Sync Menu）
   - 处理同步成功/失败状态
   - 返回统一的响应格式

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端（配置和监控）
- [x] KDS 厨显端（订单展示）
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（核心）
- [x] 数据模型（Protobuf 定义）
- [x] 业务逻辑（后续实现）
- [x] 第三方集成（LINE MAN 平台）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：API 定义，无复杂业务逻辑
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3 SP（待技术评审确认）

**工作分解**：
1. 创建 API 定义文件（Go struct）（1 天）
   - oauth.go, order.go, menu.go, common.go
   - 包含完整的验证规则和中文注释
2. 编写 API 文档和注释（0.5 天）
   - 代码内注释
   - 字段说明（`dc` tag）
3. 更新集成说明文档（0.5 天）
   - 更新 `ttpos-bmp/docs/shared/integrations/lineman.md`
   - 包含使用说明、示例、配置说明
4. 代码审查和格式验证（0.5-1 天）
   - go fmt 格式化
   - 数据结构验证

### 风险识别

**潜在风险**：
1. LINE MAN API 文档可能存在更新或不一致
2. 时间格式（ISO 8601）和时区（UTC+7）需要特别处理
3. OAuth 认证逻辑需要与 LINE MAN 的授权机制匹配
4. 订单数据结构较复杂（嵌套商品、属性等）

**缓解措施**：
1. 参考现有文档 `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/*.md`
2. 使用 Go `time.Time` 类型和标准时间解析函数处理时间字段
3. 在 API 定义中添加详细的中文注释说明字段含义和限制
4. 使用 GoFrame 的验证规则（`v` tag）确保数据完整性

---

## 🔗 相关资源

### 参考需求

- LINE MAN 服务文档: `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-service.md`
- LINE MAN Inbound API 文档: `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-*.md`（5 个）
- LINE MAN 双向 API 文档: `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/io-*.md`（1 个）
- GrabFood 对接文档: `ttpos-bmp/docs/shared/integrations/grabfood.md`

### 相关文档

- GoFrame Go 代码开发规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`
- ttpos-takeout 模块规则: `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`
- GoFrame API 定义文档: https://goframe.org/pages/viewpage.action?pageId=1114367

### 输出文档

完成本需求后，需要更新以下文档：
- **集成说明文档**: `ttpos-bmp/docs/shared/integrations/lineman.md`
  - 包含 API 定义的使用说明
  - 包含认证机制说明
  - 包含请求/响应示例
  - 包含错误处理说明
  - 包含部署配置说明

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | rikugun |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-lineman-api-definition`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint 1
- [ ] 完成后更新集成文档：`ttpos-bmp/docs/shared/integrations/lineman.md`

---

## 📝 附录

### API 清单

根据 LINE MAN 服务文档，需要定义以下 Webhook API（LINE MAN → TTPOS）：

| 序号 | API 名称 | 端点 | HTTP 方法 | 方向 | 参考文档 |
| --- | --- | --- | --- | --- | --- |
| 1 | OAuth 认证 | `/v1/lmwn/oauth2/token` | POST | LINE MAN ← TTPOS | `io-auth.md` |
| 2 | 订单创建 | `/v1/lmwn/partners/{partnerId}/stores/{storeId}/orders` | POST | LINE MAN → TTPOS | `i-place-order.md` |
| 3 | 订单状态更新 | `/v1/lmwn/partners/{partnerId}/stores/{storeId}/order/status` | POST | LINE MAN → TTPOS | `i-order-status-update-notification.md` |
| 4 | 订单更新通知 | `/v1/lmwn/partners/{partnerId}/stores/{storeId}/orders` | PUT | LINE MAN → TTPOS | `i-order-update-notification.md` |
| 5 | 菜单同步通知 | `/v1/lmwn/partners/{partnerId}/stores/{storeId}/menus/notification` | POST | LINE MAN → TTPOS | `i-menu-sync-notification.md` |
| 6 | 菜单同步触发 | `/v1/lmwn/partners/{partnerId}/stores/{storeId}/menus/trigger-sync` | POST | LINE MAN → TTPOS | `i-trigger-sync-menu.md` |

**文件命名规则**：
- `i-*.md`：Inbound API，LINE MAN → TTPOS（5 个）
- `io-*.md`：双向 API，LINE MAN ← TTPOS（1 个，OAuth 认证）
- `o-*.md`：Outbound API，TTPOS → LINE MAN（不在本提案范围内）

**说明**：
- 所有 API 都是 LINE MAN 主动调用 TTPOS 的 Webhook 接口
- OAuth 认证是双向的：LINE MAN 调用此接口获取令牌用于后续请求
- 需要实现 OAuth Bearer Token 认证验证
- 返回统一的 JSON 响应格式：`{"status": "ok|fail", "code": "...", "message": "..."}`

### API 定义文件结构

```
ttpos-bmp/app/ttpos-takeout/api/lineman/v1/
├── oauth.go              # OAuth 认证 API 定义
├── order.go              # 订单相关 API 定义（创建、状态更新、更新通知）
├── menu.go               # 菜单同步 API 定义（同步通知、触发同步）
└── common.go             # 通用数据结构和响应格式
```

**文件说明**：
- 每个文件包含对应功能的 Request 和 Response 结构体
- 使用 GoFrame 的 `g.Meta` 标签定义路由和请求方法
- 使用 `v` 标签定义验证规则
- 使用 `json` 标签定义 JSON 字段映射

### User Story（初稿）

**作为** 外卖系统开发者  
**我想** 在 ttpos-takeout 模块中定义完整的 LINE MAN Webhook API（RESTful 风格）  
**以便于** 接收 LINE MAN 平台的订单通知和管理请求，实现标准化对接

### AC 验收标准（初稿）

1. **WHEN** 创建 API 定义文件 **THEN** 系统 **SHALL** 在 `api/lineman/v1/` 目录下包含 6 个 Webhook API 的定义
2. **WHEN** 查看 API 定义 **THEN** 系统 **SHALL** 符合 GoFrame 开发规范（请求以 Req 结尾，响应以 Resp 结尾）
3. **WHEN** 查看结构体定义 **THEN** 系统 **SHALL** 包含完整的中文注释、验证规则（`v` tag）和 JSON 映射（`json` tag）
4. **WHEN** 验证数据结构 **THEN** 系统 **SHALL** 与 LINE MAN 官方文档保持一致
5. **WHEN** 执行 `go fmt` 命令 **THEN** 系统 **SHALL** 代码格式符合 Go 标准
6. **WHEN** 需求完成后 **THEN** 系统 **SHALL** 更新集成说明文档到 `ttpos-bmp/docs/shared/integrations/lineman.md`

### API 定义示例

#### 示例 1：OAuth 认证 API（io-auth.md）

```go
// OAuth 认证 API 定义
package v1

import "github.com/gogf/gf/v2/frame/g"

// OAuthTokenReq OAuth 令牌请求
// Content-Type: application/x-www-form-urlencoded
type OAuthTokenReq struct {
	g.Meta       `path:"/v1/lmwn/oauth2/token" method:"post" tags:"LINE MAN OAuth" summary:"OAuth 认证接口"`
	GrantType    string `json:"grant_type" v:"required|in:client_credentials#授权类型不能为空|授权类型必须为client_credentials" dc:"OAuth 授权类型，固定值：client_credentials"`
	ClientId     string `json:"client_id" v:"required#客户端ID不能为空" dc:"LINE MAN 分配的客户端 ID"`
	ClientSecret string `json:"client_secret" v:"required#客户端密钥不能为空" dc:"LINE MAN 分配的客户端密钥"`
}

// OAuthTokenResp OAuth 令牌响应
type OAuthTokenResp struct {
	AccessToken string `json:"access_token" dc:"访问令牌，用于后续 API 调用"`
	TokenType   string `json:"token_type" dc:"令牌类型，固定值：Bearer"`
	ExpiresIn   int    `json:"expires_in" dc:"令牌有效期（秒），通常为 3600"`
}
```

#### 示例 2：订单创建 API（i-place-order.md）

```go
// PlaceOrderReq 订单创建请求
type PlaceOrderReq struct {
	g.Meta            `path:"/v1/lmwn/partners/:partnerId/stores/:storeId/orders" method:"post" tags:"LINE MAN Order" summary:"接收订单创建通知"`
	PartnerId         string               `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID"`
	StoreId           string               `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID"`
	OrderId           string               `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度不能超过20" dc:"订单唯一 ID，格式：LMF-yyMMdd-{generated number}"`
	OrderShortCode    string               `json:"orderShortCode" v:"required|length:4,4#短订单ID不能为空|短订单ID必须为4位" dc:"短订单 ID，为 orderId 的后四位"`
	RestaurantRevenue float64              `json:"restaurantRevenue" v:"required#商户收入不能为空" dc:"商户收入总额（已扣除合作伙伴补贴折扣）"`
	OrderAcceptedTime string               `json:"orderAcceptedTime" v:"required#订单接受时间不能为空" dc:"订单接受时间，ISO 8601 格式"`
	Items             []OrderItem          `json:"items" v:"required#订单商品不能为空" dc:"订单商品列表"`
	AdditionalItems   []OrderAdditionalItem `json:"additionalItems" dc:"订单附加项列表"`
	MemberId          string               `json:"memberId" dc:"绑定 LINE MAN 账号的会员 ID"`
	CustomerType      string               `json:"customerType" v:"required|in:DELIVERY,PICKUP#订单类型不能为空|订单类型必须为DELIVERY或PICKUP" dc:"订单类型：DELIVERY（外送）或 PICKUP（自取）"`
}

// OrderItem 订单商品
type OrderItem struct {
	Id          string              `json:"id" v:"required#商品ID不能为空" dc:"菜单项 ID"`
	Quantity    int                 `json:"quantity" v:"required|min:1#商品数量不能为空|商品数量至少为1" dc:"商品数量"`
	UnitPrice   float64             `json:"unitPrice" v:"required#商品单价不能为空" dc:"商品单价（THB），包含额外选项费用"`
	Memo        string              `json:"memo" dc:"顾客备注"`
	PromotionId string              `json:"promotionId" dc:"促销活动 ID"`
	Discount    float64             `json:"discount" dc:"促销折扣金额"`
	Properties  []OrderItemProperty `json:"properties" dc:"商品选项"`
}
```

#### 通用响应格式

```go
// LinemanCommonResp LINE MAN 通用响应格式
type LinemanCommonResp struct {
	Status  string `json:"status" dc:"结果状态：ok 表示成功，fail 表示失败"`
	Code    string `json:"code" dc:"结果代码"`
	Message string `json:"message,omitempty" dc:"结果描述"`
}
```

**说明**：
- 使用 `g.Meta` 标签定义路由路径、HTTP 方法、分组和描述
- 使用 `v` 标签定义验证规则和错误提示（中文）
- 使用 `json` 标签定义 JSON 字段映射
- 使用 `dc` 标签添加字段描述（用于文档生成）
- OAuth 认证使用 `application/x-www-form-urlencoded`，其他接口使用 `application/json`
- 路径参数使用 `:paramName` 格式（如 `:partnerId`、`:storeId`）

### 集成说明文档大纲

完成需求后，需要更新 `ttpos-bmp/docs/shared/integrations/lineman.md`，建议包含以下内容：

```markdown
# LINE MAN 平台集成说明

## 概述
- LINE MAN 平台简介
- 集成架构图
- 支持的功能列表

## API 定义
- 代码位置：ttpos-bmp/app/ttpos-takeout/api/lineman/v1/
- API 列表和说明

## 认证机制
- OAuth 2.0 Client Credentials 流程
- 令牌获取和刷新
- Bearer Token 使用方式

## API 端点说明
### 1. OAuth 认证
### 2. 订单管理
### 3. 菜单同步

## 请求/响应示例
- 每个 API 的完整示例
- 成功和失败场景

## 错误处理
- 错误码说明
- HTTP 状态码说明
- 常见错误和解决方案

## 部署配置
- 环境变量配置
- 路由配置
- 中间件配置

## 测试说明
- 测试环境配置
- 测试用例
- Mock 数据

## 常见问题（FAQ）
```

### 线框图/原型（可选）

无需 UI 设计，本需求为纯 API 定义。

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2026-01-07  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

