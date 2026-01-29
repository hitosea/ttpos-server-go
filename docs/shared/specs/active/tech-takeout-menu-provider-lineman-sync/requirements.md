# Menu Provider 多平台支持与 Lineman 状态同步 需求文档

> 本文档定义 Menu Provider 多平台支持与 Lineman 状态同步功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.14.0-menu-provider-lineman-status-sync.md](../../../../team/proposals/2026-01/v2.14.0-menu-provider-lineman-status-sync.md) |
| **创建日期**      | 2026-01-13                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-13             |
| **审核意见** | 技术任务，审核通过，进入设计阶段         |

---

## 📋 概述

当前 TTPOS 外卖系统的菜单更新协议仅支持 Grab 平台，随着业务拓展至 Lineman 平台，需要支持多平台的差异化处理。本功能通过在 Protobuf 协议中增加 `provider_name` 字段来标识平台，并实现针对不同平台的状态映射和字段校验逻辑，确保商户可以在 TTPOS 后台统一管理商品状态并自动同步到各外卖平台。

**核心价值**：
- 提升平台扩展性，为接入更多外卖平台奠定基础
- 确保 TTPOS 与各平台的商品状态实时一致
- 优化用户体验，一处操作多平台同步
- 降低状态映射错误率，避免商品错误上下架

## 🎯 产品对齐

本功能支持 TTPOS 的多平台外卖战略，通过统一的后台管理实现多平台商品状态同步，提升商户运营效率。随着泰国市场 Lineman 平台的接入，以及未来可能接入 FoodPanda、ShopeeFood 等平台，本功能为平台扩展提供了可复用的技术架构。

**业务目标对齐**：
- 支持泰国市场多平台外卖业务
- 提升商户管理效率 30%（减少多平台重复操作）
- 降低状态不一致导致的客诉率

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 TTPOS Shop 后台统一管理商品状态（上架、下架、售罄）  
**以便于** 无需在 Grab、Lineman 等多个平台分别操作，自动同步到所有平台，提高管理效率并减少状态不一致问题

---

## 功能需求

### Requirement 1: Protobuf 协议扩展 - 增加平台标识字段

**用户故事**: 作为系统开发者，我想在菜单更新协议中增加平台标识字段，以便于区分不同平台的处理逻辑

#### 验收标准

1. **WHEN** 定义 `UpdateMenuItemReq` 消息 **THEN** 系统 **SHALL** 包含 `provider_name` 字段（类型：string，可选）
2. **WHEN** `provider_name` 未指定 **THEN** 系统 **SHALL** 默认使用 `grab` 值
3. **WHEN** `provider_name` 指定为 `lineman` **THEN** 系统 **SHALL** 执行 Lineman 专用处理逻辑
4. **WHEN** 修改 Protobuf 文件 **THEN** 系统 **SHALL** 重新生成对应的 Go 代码文件

#### 具体要求

- [x] 1.1 在 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 的 `UpdateMenuItemReq` 中新增 `provider_name` 字段
- [x] 1.2 字段序号使用 `9`（避免与现有字段冲突）
- [x] 1.3 字段类型为 `optional string`，默认值为 `grab`
- [x] 1.4 可选值包括：`grab`、`lineman`（为未来平台预留扩展空间）
- [x] 1.5 在注释中说明字段用途和可选值
- [x] 1.6 执行 `make proto` 重新生成 `menu.pb.go` 文件

**Protobuf 定义**：
```protobuf
// 更新菜单项请求
message UpdateMenuItemReq {
  string merchant_id = 1;                         // Grab MerchantID (必填)
  string item_id = 2;                             // 商品ID (partner item id, 必填)
  optional int64 price = 3;                       // 价格 (minor unit，单位：分)
  optional string available_status = 4;           // 可用状态: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE, SOLD_OUT_TODAY
  optional int64 max_stock = 5;                   // 库存数量
  repeated AdvancedPricing advanced_pricings = 6; // 高级定价配置
  repeated Purchasability purchasabilities = 7;   // 购买能力配置
  string request_id = 8;                          // 请求 ID (可选，用于追踪)
  optional string provider_name = 9;              // 平台名称: grab (默认), lineman，为未来平台预留扩展
}
```

---

### Requirement 2: Lineman 状态映射实现

**用户故事**: 作为系统，我想将 TTPOS 内部状态正确映射为 Lineman 平台的状态枚举，以便于确保状态同步的正确性

#### 验收标准

1. **WHEN** `available_status=AVAILABLE` 且 `provider_name=lineman` **THEN** 系统 **SHALL** 映射为 Lineman 的 `AVAILABLE`
2. **WHEN** `available_status=UNAVAILABLE` 且 `provider_name=lineman` **THEN** 系统 **SHALL** 映射为 Lineman 的 `SUSPENDED`
3. **WHEN** `available_status=SOLD_OUT_TODAY` 且 `provider_name=lineman` **THEN** 系统 **SHALL** 映射为 Lineman 的 `SOLD_OUT_TODAY`
4. **WHEN** `available_status` 包含不支持的状态值（如 `UNAVAILABLEHIDE`）且 `provider_name=lineman` **THEN** 系统 **SHALL** 返回错误提示
5. **WHEN** `provider_name=grab` 或未指定 **THEN** 系统 **SHALL** 使用 Grab 状态映射逻辑（保持现有逻辑不变）

#### 具体要求

- [x] 2.1 在 `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/` 创建状态映射函数 `MapMenuStatus`
- [x] 2.2 实现状态映射表：
  - `AVAILABLE` → `AVAILABLE`
  - `UNAVAILABLE` → `SUSPENDED`
  - `SOLD_OUT_TODAY` → `SOLD_OUT_TODAY`
- [x] 2.3 对于不支持的状态（如 `UNAVAILABLEHIDE`），返回明确的错误信息
- [x] 2.4 添加单元测试覆盖所有状态映射场景
- [x] 2.5 在代码中添加注释说明映射规则和 Lineman API 文档链接

**状态映射表**：

| TTPOS 状态 (available_status) | Grab 状态         | Lineman 状态 (menuStatus) | 说明 |
|-------------------------------|-------------------|---------------------------|------|
| AVAILABLE                     | AVAILABLE         | AVAILABLE                 | 商品可售 |
| UNAVAILABLE                   | UNAVAILABLE       | SUSPENDED                 | 商品暂停销售 |
| UNAVAILABLEHIDE               | UNAVAILABLEHIDE   | -（不支持，返回错误）      | Lineman 不支持隐藏状态 |
| SOLD_OUT_TODAY                | -（不支持）       | SOLD_OUT_TODAY            | 今日售罄，明天自动恢复为 AVAILABLE |

---

### Requirement 3: Lineman 字段校验逻辑

**用户故事**: 作为系统，我想校验 Lineman 平台的更新请求只包含支持的字段，以便于避免不支持的字段导致 API 调用失败

#### 验收标准

1. **WHEN** `provider_name=lineman` 且请求包含 `price` 字段 **THEN** 系统 **SHALL** 返回错误 "Lineman 平台仅支持更新 available_status 字段"
2. **WHEN** `provider_name=lineman` 且请求包含 `max_stock` 字段 **THEN** 系统 **SHALL** 返回错误 "Lineman 平台仅支持更新 available_status 字段"
3. **WHEN** `provider_name=lineman` 且请求包含 `advanced_pricings` 字段 **THEN** 系统 **SHALL** 返回错误 "Lineman 平台仅支持更新 available_status 字段"
4. **WHEN** `provider_name=lineman` 且请求包含 `purchasabilities` 字段 **THEN** 系统 **SHALL** 返回错误 "Lineman 平台仅支持更新 available_status 字段"
5. **WHEN** `provider_name=lineman` 且请求仅包含 `available_status` 字段 **THEN** 系统 **SHALL** 通过校验，继续执行更新
6. **WHEN** `provider_name=grab` 或未指定 **THEN** 系统 **SHALL** 跳过此校验（Grab 支持所有字段）

#### 具体要求

- [x] 3.1 在 `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go` 的 `UpdateMenuItem` 方法中添加字段校验逻辑
- [x] 3.2 创建校验函数 `validateLinemanRequest`，检查请求是否只包含 `available_status`
- [x] 3.3 对于不支持的字段，返回明确的错误信息（包含字段名称）
- [x] 3.4 错误信息支持国际化（i18n）
- [x] 3.5 添加单元测试覆盖所有字段校验场景
- [x] 3.6 在日志中记录校验失败的请求详情（用于调试）

---

### Requirement 4: Lineman API 客户端实现

**用户故事**: 作为系统，我想调用 Lineman 的菜单状态更新 API，以便于将状态变更同步到 Lineman 平台

#### 验收标准

1. **WHEN** 调用 Lineman API 更新菜单状态 **THEN** 系统 **SHALL** 使用 `PUT /v1/partners/{partnerId}/stores/{storeId}/menu/items/status` 接口
2. **WHEN** 构造请求体 **THEN** 系统 **SHALL** 按照 Lineman API 规范格式化 JSON（包含 `menuItems` 数组）
3. **WHEN** API 调用成功（HTTP 200）**THEN** 系统 **SHALL** 解析响应并返回成功
4. **WHEN** API 调用失败（HTTP 4xx/5xx）**THEN** 系统 **SHALL** 记录错误日志并返回明确的错误信息
5. **WHEN** 网络异常或超时 **THEN** 系统 **SHALL** 执行重试机制（最多 3 次）
6. **WHEN** 所有重试均失败 **THEN** 系统 **SHALL** 返回失败并记录到错误日志

#### 具体要求

- [x] 4.1 在 `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/` 创建 `menu_status_client.go`
- [x] 4.2 实现 `UpdateMenuItemStatus` 方法，调用 Lineman API
- [x] 4.3 使用现有的 HTTP 客户端配置（认证、超时、重试）
- [x] 4.4 请求体格式：
  ```json
  {
    "menuItems": [
      {
        "id": "partner-item-id",
        "menuStatus": "SUSPENDED"
      }
    ]
  }
  ```
- [x] 4.5 响应体格式：
  ```json
  {
    "status": "ok",
    "code": "SUCCESS",
    "message": "Menu status updated"
  }
  ```
- [x] 4.6 支持批量更新（一次请求最多 100 个商品）
- [x] 4.7 添加请求/响应日志（包含 request_id 用于追踪）
- [x] 4.8 集成测试覆盖 API 调用成功和失败场景
- [x] 4.9 参考 Lineman API 文档：[Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633)

---

### Requirement 5: Controller 层路由分发逻辑

**用户故事**: 作为系统，我想根据 `provider_name` 字段将请求路由到对应的平台处理逻辑，以便于实现多平台支持

#### 验收标准

1. **WHEN** `provider_name=grab` 或未指定 **THEN** 系统 **SHALL** 调用现有的 Grab 处理逻辑
2. **WHEN** `provider_name=lineman` **THEN** 系统 **SHALL** 调用 Lineman 专用处理逻辑
3. **WHEN** `provider_name` 为不支持的值（如 `foodpanda`）**THEN** 系统 **SHALL** 返回错误 "不支持的平台: {provider_name}"
4. **WHEN** 处理成功 **THEN** 系统 **SHALL** 返回统一的成功响应格式
5. **WHEN** 处理失败 **THEN** 系统 **SHALL** 返回统一的错误响应格式（包含错误码和错误信息）

#### 具体要求

- [x] 5.1 在 `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go` 的 `UpdateMenuItem` 方法中添加平台路由逻辑
- [x] 5.2 使用 switch-case 根据 `provider_name` 分发请求
- [x] 5.3 Grab 处理逻辑保持现有代码不变（向后兼容）
- [x] 5.4 Lineman 处理逻辑包括：
  - 字段校验（调用 Requirement 3）
  - 状态映射（调用 Requirement 2）
  - API 调用（调用 Requirement 4）
- [x] 5.5 添加集成测试覆盖 Grab 和 Lineman 两种场景
- [x] 5.6 在日志中记录平台类型和处理结果

---

### Requirement 6: available_status 枚举扩展

**用户故事**: 作为系统，我想支持 `SOLD_OUT_TODAY` 状态，以便于 Lineman 平台实现"今日售罄，明天自动恢复"的功能

#### 验收标准

1. **WHEN** `available_status=SOLD_OUT_TODAY` **THEN** 系统 **SHALL** 识别为合法状态
2. **WHEN** `available_status=SOLD_OUT_TODAY` 且 `provider_name=lineman` **THEN** 系统 **SHALL** 映射为 Lineman 的 `SOLD_OUT_TODAY`
3. **WHEN** `available_status=SOLD_OUT_TODAY` 且 `provider_name=grab` **THEN** 系统 **SHALL** 返回错误 "Grab 平台不支持 SOLD_OUT_TODAY 状态"
4. **WHEN** 前端或调用方传入 `SOLD_OUT_TODAY` **THEN** 系统 **SHALL** 正确解析和处理

#### 具体要求

- [x] 6.1 在 Protobuf `UpdateMenuItemReq` 的 `available_status` 字段注释中添加 `SOLD_OUT_TODAY`
- [x] 6.2 在状态映射逻辑中支持 `SOLD_OUT_TODAY`（仅 Lineman）
- [x] 6.3 在字段校验逻辑中允许 `SOLD_OUT_TODAY`
- [x] 6.4 添加单元测试覆盖 `SOLD_OUT_TODAY` 状态的处理
- [x] 6.5 在文档中说明 `SOLD_OUT_TODAY` 的语义和平台支持情况

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → Client 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 和 Client 应独立且可复用
- **依赖管理**: Controller 依赖 Logic，Logic 依赖 Client
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] 遵循 Lineman API 规范（参考文档链接）
- [x] HTTP 请求使用 JSON 格式
- [x] 响应格式统一：`{status, code, message}`
- [x] 错误码明确且可追溯
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 性能要求

- [x] API 调用超时时间 30 秒
- [x] 失败重试最多 3 次，间隔 1 秒
- [x] 批量更新最多支持 100 个商品
- [x] 日志记录不影响主流程性能

### 测试要求

- [x] 单元测试覆盖率 ≥ 80%（状态映射、字段校验、路由分发）
- [x] 集成测试覆盖 Grab 和 Lineman 两种平台场景
- [x] API Mock 测试（模拟 Lineman API 响应）
- [x] 向后兼容测试（确保现有 Grab 调用不受影响）
- [x] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 国际化要求

- [x] 错误信息支持多语言（中文、英文、泰语）
- [x] 使用 i18n 配置文件管理文案
- [x] 参考: `ttpos-bmp/i18n/` - 国际化配置

### 安全要求

- [x] Lineman API 调用使用 OAuth 2.0 认证
- [x] 敏感信息（如 access_token）不记录到日志
- [x] 参数校验防止注入攻击
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时自动重试（指数退避）
- [x] 失败请求记录到错误日志（包含 request_id）
- [x] 使用 request_id 追踪请求链路
- [x] 监控 API 调用成功率和响应时间

---

## 验收标准

### 功能验收

1. **Protobuf 扩展**: `provider_name` 字段正确添加且生成代码无误
2. **状态映射**: 所有状态映射场景测试通过（AVAILABLE, UNAVAILABLE, SOLD_OUT_TODAY）
3. **字段校验**: Lineman 请求只包含 `available_status` 时通过，包含其他字段时返回明确错误
4. **API 调用**: 成功调用 Lineman API 并正确解析响应
5. **路由分发**: 根据 `provider_name` 正确路由到 Grab 或 Lineman 处理逻辑
6. **向后兼容**: 现有 Grab 调用方无需修改代码，功能正常运行

### 测试验收

1. **单元测试**: 覆盖率 ≥ 80%，所有测试用例通过
2. **集成测试**: Grab 和 Lineman 场景测试通过
3. **API Mock 测试**: 模拟 Lineman API 成功和失败场景
4. **向后兼容测试**: 使用现有测试用例验证 Grab 功能未受影响

### 文档验收

1. **技术文档**: 创建 design.md，包含详细技术方案
2. **API 文档**: 更新 Lineman API 调用说明
3. **Protobuf 文档**: 在 proto 文件中添加完整注释
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- Protobuf 文件位于 `manifest/protobuf/` 目录
- 使用 `make proto` 重新生成代码

#### Protobuf 规范

- 字段序号不能冲突
- 使用 `optional` 修饰可选字段
- 注释必须完整且清晰
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

#### Lineman API 约束

- 必须遵循 Lineman API 规范（参考 Google Sheets 文档）
- 仅支持更新 `available_status` 字段
- 批量更新最多 100 个商品
- 状态枚举：`AVAILABLE`, `SUSPENDED`, `SOLD_OUT_TODAY`

### 业务约束

- 不影响现有 Grab 平台的菜单更新功能
- Lineman 平台仅支持状态更新，不支持价格、库存等字段更新
- `SOLD_OUT_TODAY` 状态仅适用于 Lineman 平台

### 资源约束

- 开发时间: 3-5 天
- Story Point: 3 (待技术评审确认)
- 单人开发，需要与 Lineman API 测试环境联调

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2`: GoFrame 框架
- `google.golang.org/protobuf`: Protobuf 库
- Lineman OAuth 2.0 认证服务
- Lineman Menu API 服务

### 服务依赖

- **ttpos-takeout → Lineman API**: HTTP 调用（菜单状态更新）
- **Shop 后台 → ttpos-takeout**: gRPC 调用（触发菜单更新）

### 业务依赖

- Lineman 平台已完成商户注册和授权
- Lineman API 测试环境可用
- 商品数据已在 Lineman 平台完成初次同步（Menu Sync）

---

## 风险和缓解

### 风险 1: Lineman API 调用失败

**影响**: 高（导致状态无法同步）  
**概率**: 中（网络问题、限流、服务故障）  
**缓解措施**:

- 实现自动重试机制（最多 3 次，指数退避）
- 记录失败请求到错误日志，包含 request_id
- 监控 API 调用成功率，设置告警阈值
- 提供手动重试入口（Shop 后台）

### 风险 2: 状态映射错误

**影响**: 高（导致商品错误上下架，影响订单）  
**概率**: 低（通过测试可避免）  
**缓解措施**:

- 单元测试覆盖所有状态映射场景
- 集成测试验证端到端流程
- Code Review 重点检查映射逻辑
- 在日志中记录映射前后的状态值

### 风险 3: 向后兼容问题

**影响**: 高（影响现有 Grab 平台功能）  
**概率**: 低（通过默认值和测试可避免）  
**缓解措施**:

- `provider_name` 默认值为 `grab`（未指定时使用 Grab 逻辑）
- 向后兼容测试验证现有功能未受影响
- 灰度发布，先在测试环境验证再上线

### 风险 4: 字段校验逻辑遗漏

**影响**: 中（导致 Lineman API 返回错误）  
**概率**: 中（新增字段可能遗漏校验）  
**缓解措施**:

- 使用白名单方式（只允许 `available_status`）
- 单元测试覆盖所有不支持字段的校验
- 在代码中添加明确注释说明 Lineman 支持的字段

---

## 时间表

- **Phase 1 - Protobuf 扩展 + 代码生成**: 0.5 天
  - 修改 `menu.proto`
  - 执行 `make proto` 生成代码
  - 验证生成代码无误

- **Phase 2 - 核心逻辑实现**: 2 天
  - 状态映射函数（0.5 天）
  - 字段校验逻辑（0.5 天）
  - Lineman API 客户端（1 天）

- **Phase 3 - Controller 集成 + 测试**: 1.5 天
  - Controller 路由分发逻辑（0.5 天）
  - 单元测试（0.5 天）
  - 集成测试（0.5 天）

- **Phase 4 - 联调 + 文档**: 1 天
  - Lineman API 联调测试（0.5 天）
  - 文档编写（design.md, tasks.md）（0.5 天）

- **总计**: 5 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/security.mdc` - 安全开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/shared/integrations/lineman/` - Lineman 集成文档

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南

### 外部参考

- [Lineman API 文档 - Google Sheets](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=585076633#gid=585076633)
- [Lineman Partner Platform](https://partner.lineman.io/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充 - 开发完成后记录]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：开发过程中的技术决策和踩坑经验应同步更新 Episode 并在 Spec 中互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-13  
**作者**: rikugun  
**审核者**: 待审核
