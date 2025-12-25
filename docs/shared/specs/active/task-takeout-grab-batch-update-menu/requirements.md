# Grab 批量更新菜单 API 集成 需求文档

> 本文档定义 Grab 批量更新菜单 API 集成 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-grab-batch-update-menu.md](../../../../team/proposals/2025-12/v2.11.0-grab-batch-update-menu.md) |
| **创建日期**      | 2025-12-23                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2025-12-23             |
| **审核意见** | 需求明确，技术方案可行，可进入设计阶段         |

---

## 📋 概述

集成 GrabFood 官方提供的 Batch Update Menu API（`PUT /partner/v1/batch/menu`），实现批量更新菜单项（商品）和修饰符功能，提升菜单同步效率，降低 API 调用成本，减少限流风险。

**当前痛点**：
- 批量更新 50 个商品需要发起 50 次 API 调用，耗时 10-15 秒
- 容易触发 API 限流（Rate Limit）
- 部分请求失败导致菜单数据不一致
- 每次单项更新产生独立日志记录，批量操作产生大量日志

**改进目标**：
- 批量更新效率提升 80%+（从 10-15 秒缩短至 1-2 秒）
- 降低 API 调用次数，减少限流风险
- 提高数据一致性（单次批量请求保证原子性）
- 优化日志管理（批量操作产生单条日志记录）

## 🎯 产品对齐

本功能支持以下产品目标：

1. **提升商户体验**：批量菜单调整操作响应更快，提升商户后台操作效率
2. **降低运营成本**：减少 API 调用次数，降低第三方平台集成成本
3. **提高系统稳定性**：减少因 API 限流导致的菜单同步失败问题

## 📝 用户故事

**作为** 系统管理员  
**我想** 通过单次 API 调用批量更新多个商品/修饰符的价格和库存  
**以便于** 提升菜单同步效率，减少 API 调用次数，降低限流风险

---

## 功能需求

### Requirement 1: 批量更新菜单项（商品）

**用户故事**: 作为 系统管理员，我想 批量更新多个商品的价格、库存和可用状态，以便于 提升菜单同步效率

#### 验收标准

1. **WHEN** 调用批量更新 API 传入 10 个商品更新请求（`field=ITEM`）**THEN** 系统 **SHALL** 在 2 秒内完成所有商品的更新
2. **WHEN** 单次批量更新包含 50 个商品 **THEN** 系统 **SHALL** 成功调用 GrabFood API 并返回更新结果
3. **IF** 批量更新超过 100 条记录 **THEN** 系统 **SHALL** 返回参数验证错误 "菜单实体数量必须在1-100之间"
4. **WHEN** 商品包含高级定价配置（Advanced Pricing）**THEN** 系统 **SHALL** 正确转换并传递给 GrabFood API
5. **WHEN** 商品包含购买能力配置（Purchasability）**THEN** 系统 **SHALL** 正确转换并传递给 GrabFood API

#### 具体要求

- [x] 1.1 支持批量更新商品价格（`price` 字段，minor unit 格式）
- [x] 1.2 支持批量更新商品库存（`maxStock` 字段）
- [x] 1.3 支持批量更新商品可用状态（`availableStatus`: AVAILABLE, UNAVAILABLE, UNAVAILABLEHIDE）
- [x] 1.4 支持批量更新商品高级定价配置（`advancedPricings` 数组）
- [x] 1.5 支持批量更新商品购买能力配置（`purchasabilities` 数组）
- [x] 1.6 单次批量更新最多支持 100 个商品
- [x] 1.7 必须使用 `field=ITEM` 标识商品类型

---

### Requirement 2: 批量更新修饰符

**用户故事**: 作为 系统管理员，我想 批量更新多个修饰符的价格和可用状态，以便于 提升菜单同步效率

#### 验收标准

1. **WHEN** 调用批量更新 API 传入 10 个修饰符更新请求（`field=MODIFIER`）**THEN** 系统 **SHALL** 在 2 秒内完成所有修饰符的更新
2. **WHEN** 修饰符包含高级定价配置 **THEN** 系统 **SHALL** 正确转换并传递给 GrabFood API
3. **IF** `field` 不是 `ITEM` 或 `MODIFIER` **THEN** 系统 **SHALL** 返回参数验证错误
4. **WHEN** 批量更新修饰符时传入 `maxStock` 或 `purchasabilities` 字段 **THEN** 系统 **SHALL** 忽略这些字段（修饰符不支持）

#### 具体要求

- [x] 2.1 支持批量更新修饰符价格（`price` 字段）
- [x] 2.2 支持批量更新修饰符可用状态（`availableStatus`: AVAILABLE, UNAVAILABLE）
- [x] 2.3 支持批量更新修饰符高级定价配置（`advancedPricings` 数组）
- [x] 2.4 修饰符不支持 `maxStock` 和 `purchasabilities` 字段
- [x] 2.5 单次批量更新最多支持 100 个修饰符
- [x] 2.6 必须使用 `field=MODIFIER` 标识修饰符类型

---

### Requirement 3: 类型隔离与错误处理

**用户故事**: 作为 系统管理员，我想 系统能正确处理批量更新的部分失败场景，以便于 快速定位和修复失败的菜单项

#### 验收标准

1. **WHEN** 批量更新部分失败（`status=partial_success`）**THEN** 系统 **SHALL** 在响应的 `errors` 数组中记录失败实体的 ID 和错误原因
2. **WHEN** 批量更新全部失败（`status=fail`）**THEN** 系统 **SHALL** 在响应中返回所有失败实体的详细错误信息
3. **WHEN** 批量更新全部成功（`status=success`）**THEN** 系统 **SHALL** 在 `menu_log` 表中生成单条日志记录
4. **WHEN** 调用 GrabFood API 失败 **THEN** 系统 **SHALL** 在 `menu_log` 表中记录失败日志，包含错误原因
5. **IF** 单次请求同时包含商品和修饰符 **THEN** 系统 **SHALL** 返回参数验证错误（必须分两次调用）

#### 具体要求

- [x] 3.1 单次请求只能更新同一类型（`field` 为 `ITEM` 或 `MODIFIER`）
- [x] 3.2 如需同时更新商品和修饰符，需分别调用两次 API
- [x] 3.3 批量操作生成单条 `menu_log` 记录，`sync_type` 为 `BATCH_UPDATE_ITEM` 或 `BATCH_UPDATE_MODIFIER`
- [x] 3.4 部分失败时，在 `error_msg` 字段中记录 JSON 格式的详细错误信息
- [x] 3.5 支持通过 `request_id` 追踪批量更新状态

---

### Requirement 4: gRPC 接口定义

**用户故事**: 作为 其他服务模块，我想 通过 gRPC 调用批量更新菜单接口，以便于 集成到业务流程中

#### 验收标准

1. **WHEN** 调用 gRPC `BatchUpdateMenu` 方法 **THEN** 系统 **SHALL** 返回 `takeout.ApiResponse` 格式的响应
2. **WHEN** gRPC 请求参数验证失败 **THEN** 系统 **SHALL** 返回 `ApiResponse` 中的 `code` 和 `message` 错误信息
3. **WHEN** gRPC 调用成功 **THEN** 系统 **SHALL** 在 `ApiResponse.data` 中返回 `BatchUpdateMenuResp` 结构

#### 具体要求

- [x] 4.1 在 `manifest/protobuf/menu/menu.proto` 中定义 `BatchUpdateMenuReq` 消息
- [x] 4.2 在 `manifest/protobuf/menu/menu.proto` 中定义 `BatchUpdateMenuResp` 消息
- [x] 4.3 在 `manifest/protobuf/menu/menu.proto` 中定义 `MenuEntity` 消息
- [x] 4.4 在 `manifest/protobuf/menu/menu.proto` 中定义 `MenuEntityError` 消息
- [x] 4.5 在 `MenuService` 中新增 `BatchUpdateMenu` rpc 方法
- [x] 4.6 响应统一使用 `takeout.ApiResponse` 包装

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → SDK 分层
- **单一职责原则**: 每个方法应有单一、明确的目的
- **模块化设计**: Logic 层方法应独立且可复用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go 代码开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - ttpos-takeout 子模块规范

### API 设计要求

- [x] 遵循 GrabFood SDK 结构定义（`BatchUpdateMenuItem`, `MenuEntity`, `BatchUpdateMenuResponse`）
- [x] 响应格式统一使用 `takeout.ApiResponse`
- [x] Controller 层负责将业务数据包装为 `ApiResponse`
- [x] Logic 层返回具体的业务数据类型（`*BatchUpdateMenuResp`）

### 数据库设计要求

- [x] 复用现有 `menu_log` 表存储批量更新日志
- [x] `sync_type` 字段使用 `BATCH_UPDATE_ITEM` 或 `BATCH_UPDATE_MODIFIER`
- [x] `error_msg` 字段存储 JSON 格式的详细错误信息
- [x] 不需要新增数据库表

### 性能要求

- [x] 批量更新 50 个商品的响应时间 < 2 秒
- [x] 单次批量更新最多 100 条记录（遵循 GrabFood API 限制）
- [x] 建议分批处理，单次批量不超过 50 条记录，避免触发限流

### 测试要求

- [x] Logic 层 `BatchUpdateMenu` 方法单元测试覆盖率 ≥ 80%
- [x] 测试覆盖成功、部分失败、全部失败三种场景
- [x] 测试参数验证逻辑（超过 100 条、`field` 值错误）
- [x] 测试 DTO 转换为 SDK 结构的正确性
- [x] 集成测试覆盖 gRPC 调用流程

### 可靠性要求

- [x] GrabFood API 调用失败时记录详细错误日志
- [x] 部分失败时返回详细的失败实体列表
- [x] 支持通过 `request_id` 追踪批量更新状态
- [x] 日志记录包含 merchantID、field、实体数量等关键信息

---

## 验收标准

### 功能验收

1. **批量更新商品**: 单次更新 50 个商品，响应时间 < 2 秒，全部成功
2. **批量更新修饰符**: 单次更新 50 个修饰符，响应时间 < 2 秒，全部成功
3. **部分失败处理**: 模拟部分商品更新失败，系统正确返回失败实体的 ID 和错误原因
4. **参数验证**: 传入超过 100 条记录，系统返回参数验证错误
5. **日志记录**: 批量更新成功后，`menu_log` 表中生成单条日志记录

### 测试验收

1. **单元测试**: Logic 层方法覆盖率 ≥ 80%
2. **集成测试**: gRPC 调用端到端流程测试通过
3. **SDK 集成测试**: 验证与 GrabFood SDK 的正确集成
4. **错误场景测试**: 测试部分失败、全部失败、参数验证错误等场景

### 文档验收

1. **技术文档**: design.md 完整且准确（待 `/spec-design` 创建）
2. **Protobuf 文档**: Protobuf 定义完整且符合规范
3. **日志文档**: 日志记录格式和字段定义清晰
4. **测试文档**: tasks.md 中的测试任务完成（待 `/spec-design` 创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- Controller 层响应必须使用 `takeout.ApiResponse` 包装
- Logic 层返回类型不能是 `takeout.ApiResponse`

#### GrabFood SDK 约束

- 必须使用 `github.com/grab/grabfood-api-sdk-go@v1.0.2` 或更高版本
- 严格遵循 SDK 结构定义（`BatchUpdateMenuItem`, `MenuEntity`, `BatchUpdateMenuResponse`）
- 单次请求只能更新同一类型（`field` 为 `ITEM` 或 `MODIFIER`）
- 单次批量更新最多 100 条记录（GrabFood API 限制）

### 业务约束

- 批量更新商品和修饰符需分两次调用（GrabFood API 限制）
- 修饰符不支持 `maxStock` 和 `purchasabilities` 字段
- 必须复用现有的 `menu_log` 表存储日志
- 日志记录的 `sync_type` 必须包含 `BATCH_UPDATE_` 前缀

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3-5 SP (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `github.com/grab/grabfood-api-sdk-go: v1.0.2` - GrabFood SDK，提供批量更新 API 结构
- `github.com/gogf/gf/v2` - GoFrame 框架
- `internal/service/grab.go` - Grab 服务接口，提供 `BatchUpdateMenu` 方法
- `internal/model/dto/grab/menu_update.go` - 现有的菜单更新 DTO（复用 `UpdateAdvancedPricingReq`, `UpdatePurchasabilityReq`）

### 服务依赖

- **ttpos-takeout → GrabFood API**: HTTP PUT `/partner/v1/batch/menu`
- **Shop 管理端 → ttpos-takeout**: gRPC 调用 `MenuService.BatchUpdateMenu`

### 业务依赖

- 现有的单项更新接口（`UpdateMenuItem`, `UpdateMenuModifier`）作为参考实现
- 现有的 `menu_log` 表结构
- 现有的 Grab 认证和签名机制

---

## 风险和缓解

### 风险 1: GrabFood API 限流风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 在业务层实现分批逻辑，单次批量建议不超过 50 条记录
- 监控批量更新的 API 调用频率
- 实现重试机制，失败时自动重试（带指数退避）

### 风险 2: 部分失败处理复杂度

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 详细记录失败实体的 ID 和错误原因（`errors` 数组）
- 在日志中记录 JSON 格式的详细错误信息
- 提供清晰的错误消息，便于排查问题

### 风险 3: 类型混合限制

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 在参数验证层明确提示"单次请求只能更新同一类型"
- 在 gRPC Controller 层或业务层实现自动分组逻辑（可选）
- 在文档中明确说明限制原因（GrabFood API 设计限制）

### 风险 4: SDK 版本兼容性

**影响**: 低  
**概率**: 低  
**缓解措施**:

- ✅ 已确认当前使用的 `github.com/grab/grabfood-api-sdk-go@v1.0.2` 支持 Batch Update Menu API
- SDK 结构定义已验证（`BatchUpdateMenuItem`, `MenuEntity`, `BatchUpdateMenuResponse`）
- 无需升级 SDK 版本或手动实现 HTTP 请求

---

## 时间表

- **Phase 1 - DTO 定义和转换逻辑**: 0.5 天
- **Phase 2 - Logic 层批量更新方法实现**: 0.5 天
- **Phase 3 - Protobuf 定义和 gRPC Controller**: 0.5 天
- **Phase 4 - 日志记录和错误处理**: 0.5 天
- **Phase 5 - 单元测试和集成测试**: 1 天
- **总计**: 3 天（SP = 3-5，待技术评审确认）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go 代码开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - ttpos-takeout 子模块规范

### 现有实现参考

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` - UpdateMenuItem, UpdateMenuModifier 实现
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go` - DTO 定义
- `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` - Protobuf 定义

### GrabFood SDK 参考

- `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_batch_update_menu_item.go` - BatchUpdateMenuItem 结构
- `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_batch_update_menu_response.go` - BatchUpdateMenuResponse 结构
- `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_menu_entity.go` - MenuEntity 结构
- `github.com/grab/grabfood-api-sdk-go@v1.0.2/model_menu_entity_error.go` - MenuEntityError 结构
- `github.com/grab/grabfood-api-sdk-go@v1.0.2/api_update_menu_record.go` - BatchUpdateMenu API 方法

### 外部参考

- [GrabFood API 官方文档](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-record/operation/update-menu)
- [Batch Update Menu API](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/update-menu-record/operation/batch-update-menu)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-23.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**审核者**: 待审核

