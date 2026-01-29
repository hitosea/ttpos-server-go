# LINE MAN API 定义 任务分解

> 本文档定义 LINE MAN Webhook API 定义的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: API 定义文件创建

### OAuth 认证 API

- [ ] 1.1 创建 oauth.go 文件

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/oauth.go`
  - Purpose: 定义 OAuth 认证相关的 API 数据结构
  - Requirements: Requirement 1 (OAuth 认证 API 定义)
  - Leverage: 参考文档 `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/io-auth.md`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 创建 LINE MAN OAuth 认证 API 定义 | Context: 使用 GoFrame 标签系统（g.Meta, v, json, dc），Content-Type 为 application/x-www-form-urlencoded，g.Meta path 为 /oauth2/token（不包含 /v1/lmwn/ 前缀），method 为 post，参考 ttpos-bmp/app/ttpos-takeout/api/grab/v1/ 的格式 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，请求结构体以 Req 结尾，响应结构体以 Res 结尾，响应结构体必须包含 g.Meta 标签使用 mime:"application/json"，所有注释使用中文，path 不包含 /v1/lmwn/ 前缀 | Success: oauth.go 文件创建成功，包含 OAuthTokenReq 和 OAuthTokenRes 结构体，响应包含 g.Meta 标签，字段定义完整，验证规则正确

### 订单相关 API

- [ ] 1.2 创建 order.go 文件 - 基础结构

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
  - Purpose: 定义订单相关 API 的基础数据结构
  - Requirements: Requirement 2, 3, 4 (订单创建、状态更新、更新通知)
  - Leverage: 参考文档 `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-place-order.md`, `i-order-status-update-notification.md`, `i-order-update-notification.md`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 创建 LINE MAN 订单 API 定义的基础结构 | Context: 使用 GoFrame 标签系统，路径参数使用 :partnerId 和 :storeId 格式，g.Meta path 不包含 /v1/lmwn/ 前缀（如 /partners/:partnerId/stores/:storeId/orders） | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，请求结构体以 Req 结尾，path 不包含 /v1/lmwn/ 前缀 | Success: order.go 文件创建成功，包含 package 声明和 import 语句

- [ ] 1.3 定义订单嵌套结构体

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
  - Purpose: 定义订单商品、属性等嵌套结构体
  - Requirements: Requirement 2 (订单创建 API 定义)
  - Leverage: Task 1.2 的基础结构，参考 `i-place-order.md` 中的数据结构
  - Prompt: Role: Go Developer | Task: 在 order.go 中定义订单相关的嵌套结构体 | Context: 包含 OrderItem（订单商品）、OrderItemProperty（商品属性）、OrderItemPropertyValue（属性值）、OrderAdditionalItem（附加项）四个结构体，每个字段包含 json 标签、v 验证标签和 dc 描述标签 | Restrictions: 所有注释使用中文，字段验证规则完整 | Success: 嵌套结构体定义完成，字段映射正确，验证规则完整

- [ ] 1.4 定义 PlaceOrderReq 结构体

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
  - Purpose: 定义订单创建请求结构体
  - Requirements: Requirement 2 (订单创建 API 定义)
  - Leverage: Task 1.3 的嵌套结构体
  - Prompt: Role: Go Developer | Task: 定义 PlaceOrderReq 结构体 | Context: g.Meta path 为 /partners/:partnerId/stores/:storeId/orders（不包含 /v1/lmwn/ 前缀），method 为 post，包含所有订单字段，使用 Task 1.3 定义的嵌套结构体 | Restrictions: 必填字段包含验证规则，中文错误提示，path 不包含 /v1/lmwn/ 前缀 | Success: PlaceOrderReq 结构体定义完成，路由和验证规则正确

- [ ] 1.5 定义 OrderStatusUpdateReq 结构体

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
  - Purpose: 定义订单状态更新请求结构体
  - Requirements: Requirement 3 (订单状态更新 API 定义)
  - Leverage: Task 1.2 的基础结构
  - Prompt: Role: Go Developer | Task: 定义 OrderStatusUpdateReq 结构体 | Context: g.Meta path 为 /partners/:partnerId/stores/:storeId/order/status（不包含 /v1/lmwn/ 前缀），method 为 post，orderStatus 字段支持 FINISH 和 CANCELED 枚举值 | Restrictions: 使用 in 验证规则限制枚举值，path 不包含 /v1/lmwn/ 前缀 | Success: OrderStatusUpdateReq 结构体定义完成，枚举验证正确

- [ ] 1.6 定义 OrderUpdateReq 结构体

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/order.go`
  - Purpose: 定义订单更新通知请求结构体
  - Requirements: Requirement 4 (订单更新通知 API 定义)
  - Leverage: Task 1.4 的 PlaceOrderReq 结构体（复用大部分字段）
  - Prompt: Role: Go Developer | Task: 定义 OrderUpdateReq 结构体 | Context: g.Meta path 为 /partners/:partnerId/stores/:storeId/orders（不包含 /v1/lmwn/ 前缀），method 为 put，类似 PlaceOrderReq 但额外包含 orderUpdatedTime 字段 | Restrictions: 复用 OrderItem 等嵌套结构体，path 不包含 /v1/lmwn/ 前缀 | Success: OrderUpdateReq 结构体定义完成，包含更新时间字段

### 菜单同步 API

- [ ] 1.7 创建 menu.go 文件

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/menu.go`
  - Purpose: 定义菜单同步相关的 API 数据结构
  - Requirements: Requirement 5, 6 (菜单同步通知、菜单同步触发)
  - Leverage: 参考文档 `ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/i-menu-sync-notification.md`, `i-trigger-sync-menu.md`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 创建 LINE MAN 菜单同步 API 定义 | Context: 包含 MenuSyncNotificationReq（g.Meta path: /partners/:partnerId/stores/:storeId/menus/notification, method: post）和 TriggerSyncMenuReq（g.Meta path: /partners/:partnerId/stores/:storeId/menus/trigger-sync, method: post）两个结构体，使用 GoFrame 标签系统，path 不包含 /v1/lmwn/ 前缀，status 字段支持 SUCCESS 和 FAILED 枚举值 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，所有注释使用中文，path 不包含 /v1/lmwn/ 前缀 | Success: menu.go 文件创建成功，两个结构体定义完整，验证规则正确，path 不包含前缀

### 响应结构体定义

**说明**：不需要创建独立的 common.go 文件，每个 API 都有自己的响应结构体（以 `Res` 结尾），但它们遵循统一的格式：
- 包含 `g.Meta` 标签，使用 `mime:"application/json"`
- 包含 `status`、`code`、`message` 三个标准字段
- 已在 oauth.go、order.go、menu.go 中分别定义

---

## Phase 2: 文档编写

- [ ] 2.1 更新集成说明文档

  - File: `ttpos-bmp/docs/shared/integrations/lineman.md`
  - Purpose: 提供 LINE MAN API 的完整集成说明
  - Requirements: 所有功能需求
  - Leverage: 提案中的"集成说明文档大纲"章节，参考 `docs/shared/integrations/grabfood.md`（如存在）
  - Prompt: Role: Technical Writer with API documentation expertise | Task: 编写 LINE MAN 平台集成说明文档 | Context: 包含概述、API 定义、认证机制、API 端点说明、请求/响应示例、错误处理、部署配置、测试说明、常见问题等章节 | Restrictions: 文档清晰、完整、易于理解，包含具体的代码示例 | Success: 集成说明文档完成，内容覆盖所有 6 个 API，示例完整

---

## Phase 3: 验证和优化

- [ ] 3.1 代码格式检查

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/*.go`
  - Purpose: 确保代码格式符合 Go 标准
  - Requirements: 代码质量要求
  - Leverage: go fmt 工具
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go fmt ./api/lineman/v1/...`
  - Success: 所有文件通过 go fmt 格式化，无格式问题

- [ ] 3.2 静态代码检查

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/*.go`
  - Purpose: 发现潜在的代码问题
  - Requirements: 代码质量要求
  - Leverage: go vet 工具
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go vet ./api/lineman/v1/...`
  - Success: 所有文件通过 go vet 检查，无警告或错误

- [ ] 3.3 与 LINE MAN 文档对比验证

  - File: -
  - Purpose: 确保 API 定义与 LINE MAN 官方文档一致
  - Requirements: 所有功能需求
  - Leverage: LINE MAN API 详细文档（6 个 md 文件）
  - Prompt: Role: QA Engineer | Task: 逐字段对比 API 定义与 LINE MAN 官方文档 | Context: 验证字段名称、类型、必填性、长度限制、枚举值等是否与文档一致 | Restrictions: 任何不一致都需要修正 | Success: 所有 API 定义与 LINE MAN 文档完全一致，无遗漏字段

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt`
- [ ] Go 代码通过 `go vet`
- [ ] 所有结构体包含完整的中文注释
- [ ] 所有字段包含 `dc` 标签描述

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 6 个 API 定义全部完成
- [ ] 通用响应格式已定义
- [ ] 验收标准已达成

### 文档同步

- [ ] 集成说明文档已更新 (`ttpos-bmp/docs/shared/integrations/lineman.md`)
- [ ] 代码注释完整清晰
- [ ] 字段描述（dc 标签）准确

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`
- [ ] 请求结构体以 `Req` 结尾
- [ ] 响应结构体以 `Res` 结尾
- [ ] 响应结构体包含 `g.Meta` 标签，使用 `mime:"application/json"`
- [ ] 使用 GoFrame 标签系统（g.Meta, v, json, dc）
- [ ] 所有注释使用中文
- [ ] 参考格式：`ttpos-bmp/app/ttpos-takeout/api/grab/v1/`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数（Phase 1 + Phase 2 + Phase 3）
grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-api-definition/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-api-definition/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-lineman-api-definition/tasks.md
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的参考文档和代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### GoFrame API 定义

```
Role: Go Developer with GoFrame 2.x expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Reference docs: {LINE MAN API 文档路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc

GoFrame 标签系统:
- g.Meta: 定义路由和 HTTP 方法
  示例: `path:"/api/v1/xxx" method:"post" tags:"API Group" summary:"API 描述"`
- v: 定义验证规则（中文错误提示）
  示例: `v:"required|length:1,20#字段不能为空|字段长度为1-20个字符"`
- json: 定义 JSON 字段映射
  示例: `json:"field_name"`
- dc: 添加字段描述（用于文档生成）
  示例: `dc:"字段的详细说明"`

Restrictions:
- 请求结构体以 Req 结尾，响应结构体以 Res 结尾
- 响应结构体必须包含 g.Meta 标签，使用 mime:"application/json"
- 所有注释使用中文
- 路径参数使用 :paramName 格式（如 :partnerId）
- 必填字段包含验证规则
- 枚举值使用 in 验证（如 `v:"in:VALUE1,VALUE2#错误提示"`）
- 参考 ttpos-bmp/app/ttpos-takeout/api/grab/v1/ 的格式

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 字段定义与 LINE MAN 文档一致
```

### 文档编写

```
Role: Technical Writer with API documentation expertise

Task: 编写 LINE MAN 平台集成说明文档

Context:
- Target file: ttpos-bmp/docs/shared/integrations/lineman.md
- API definitions: ttpos-bmp/app/ttpos-takeout/api/lineman/v1/*.go
- Reference: 提案中的"集成说明文档大纲"章节
- LINE MAN API docs: ttpos-bmp/docs/human/architecture/modules/takeout/features/lineman-api/*.md

Document Structure:
1. 概述（LINE MAN 平台简介、集成架构、功能列表）
2. API 定义（代码位置、API 列表）
3. 认证机制（OAuth 2.0 流程）
4. API 端点说明（每个 API 的详细说明）
5. 请求/响应示例（完整示例）
6. 错误处理（错误码、HTTP 状态码）
7. 部署配置（环境变量、路由、中间件）
8. 测试说明（测试环境、测试用例、Mock 数据）
9. 常见问题（FAQ）

Restrictions:
- 使用中文编写
- 包含具体的代码示例
- 示例数据真实可用
- 格式清晰易读

Success Criteria:
- 文档完整覆盖所有 6 个 API
- 包含完整的请求/响应示例
- 部署配置说明清晰
- 示例代码可直接使用
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**目标版本**: v2.13.1  
**最后更新**: 2026-01-07  
**维护者**: rikugun

