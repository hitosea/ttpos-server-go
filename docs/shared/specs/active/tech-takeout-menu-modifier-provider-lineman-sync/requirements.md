# Menu Modifier Provider 多平台支持与 Lineman 状态同步 - 需求文档

> 本文档定义功能的业务需求和验收标准。

---

## 📋 基本信息

| 项目         | 内容                                                                 |
| ------------ | -------------------------------------------------------------------- |
| **功能名称** | Menu Modifier Provider 多平台支持与 Lineman 状态同步                 |
| **Spec ID**  | tech-takeout-menu-modifier-provider-lineman-sync                     |
| **类型**     | 技术任务                                                             |
| **模块**     | ttpos-takeout（BMP 微服务）                                          |
| **版本**     | v2.14.0                                                              |
| **创建日期** | 2026-01-13                                                           |
| **负责人**   | rikugun                                                              |
| **优先级**   | 高                                                                   |
| **状态**     | 待审核                                                               |
| **来源 Proposal** | [v2.14.0-menu-modifier-provider-lineman-support](../../../team/proposals/2026-01/v2.14.0-menu-modifier-provider-lineman-support.md) |

---

## 🎯 需求概述

### 背景

当前 `UpdateMenuModifierReq` 协议只支持 Grab 平台的修饰符（Modifier/Property Values）更新逻辑。随着业务拓展到 Lineman 平台，需要支持不同平台的差异化处理。

**核心问题**：
1. **平台识别缺失**：无法区分当前更新请求是针对哪个平台
2. **状态类型不一致**：TTPOS 使用 `string` 类型，Lineman API 使用 `int` 类型
3. **状态语义映射不同**：不同平台的状态枚举值和语义不同
4. **API 调用差异**：不同平台需要调用不同的第三方 API

### 目标

在 `UpdateMenuModifierReq` 中增加多平台支持，实现：
1. 通过 `provider_name` 字段识别平台（默认 `grab`）
2. 当 `provider_name=lineman` 时，支持 TTPOS 状态（string）到 Lineman 状态（int）的转换
3. 调用 Lineman 专用的修饰符状态更新 API
4. 保持向后兼容，现有 Grab 逻辑不受影响

### 业务价值

- **平台扩展性**：为后续接入更多外卖平台奠定基础
- **状态一致性**：确保 TTPOS 与各平台的修饰符状态实时同步
- **用户体验**：商户在一个后台操作，自动同步到所有平台
- **降低错误率**：避免状态映射错误导致的修饰符上下架问题

---

## 👥 目标用户

### 主要用户

- **商户管理员**：在 TTPOS Shop 后台管理菜单修饰符状态
- **外卖运营人员**：监控修饰符状态同步情况

### 使用场景

1. **场景 1：修饰符临时售罄**
   - 商户在 Shop 后台将"加辣"修饰符设置为"今日售罄"
   - 系统自动同步到 Lineman 平台（明天自动恢复）
   - 系统自动同步到 Grab 平台（需要手动恢复）

2. **场景 2：修饰符永久下架**
   - 商户将"特大份"修饰符设置为"不可用"
   - 系统同步到 Lineman（状态：SUSPENDED）
   - 系统同步到 Grab（状态：UNAVAILABLE）

3. **场景 3：修饰符重新上架**
   - 商户将修饰符恢复为"可用"
   - 系统同步到所有平台

---

## 📝 核心功能需求

### Requirement 1: 增加平台识别字段

**描述**：在 `UpdateMenuModifierReq` 中增加 `provider_name` 字段，用于标识目标平台。

**详细说明**：
- 字段名称：`provider_name`
- 类型：`optional string`
- 默认值：`grab`
- 可选值：`grab`, `lineman`
- 位置：Protobuf 字段编号 9

**验收标准**：
- [x] Protobuf 定义中包含 `provider_name` 字段
- [x] 未指定 `provider_name` 时，默认值为 `grab`
- [x] 支持 `grab` 和 `lineman` 两种值
- [x] 生成的 Go 代码中包含 `ProviderName *string` 字段

---

### Requirement 2: TTPOS 状态到 Lineman 状态的映射

**描述**：实现 TTPOS 的 `available_status`（string）到 Lineman `status`（int）的转换逻辑。

**状态映射表**：

| TTPOS 状态 (string) | Lineman 状态 (int) | Lineman 语义       | 说明                   |
| ------------------- | ------------------ | ------------------ | ---------------------- |
| `AVAILABLE`         | `1`                | AVAILABLE          | 可用                   |
| `UNAVAILABLE`       | `3`                | SUSPENDED          | 暂停供应               |
| `SOLD_OUT_TODAY`    | `2`                | SOLD_OUT_TODAY     | 今日售罄（明天自动恢复） |

**验收标准**：
- [x] 实现状态映射函数 `MapStatusToLinemanModifier(ttposStatus string) (int, error)`
- [x] `AVAILABLE` → `1`
- [x] `UNAVAILABLE` → `3`
- [x] `SOLD_OUT_TODAY` → `2`
- [x] 未知状态返回错误
- [x] 单元测试覆盖所有映射场景

---

### Requirement 3: Lineman 修饰符状态更新 DTO

**描述**：创建 Lineman 专用的修饰符状态更新请求/响应 DTO。

**DTO 定义**：

```go
// ModifierStatusUpdateReq Lineman 修饰符状态更新请求
type ModifierStatusUpdateReq struct {
    PropertyValues []ModifierPropertyValue `json:"propertyValues"`
}

// ModifierPropertyValue 修饰符属性值
type ModifierPropertyValue struct {
    ID     string `json:"id"`     // Partner Modifier ID
    Status int    `json:"status"` // 1=AVAILABLE, 2=SOLD_OUT_TODAY, 3=SUSPENDED
}

// ModifierStatusUpdateResp Lineman 修饰符状态更新响应
type ModifierStatusUpdateResp struct {
    Status  string `json:"status"`  // ok, fail
    Code    string `json:"code"`    // SUCCESS, ERROR
    Message string `json:"message"` // 响应消息
}
```

**验收标准**：
- [x] 创建 `internal/model/dto/lineman/modifier_status.go`
- [x] 定义 `ModifierStatusUpdateReq` 结构体
- [x] 定义 `ModifierPropertyValue` 结构体（使用 `int` 类型的 `Status`）
- [x] 定义 `ModifierStatusUpdateResp` 结构体
- [x] JSON 标签正确

---

### Requirement 4: Lineman 修饰符状态更新 API 客户端

**描述**：实现调用 Lineman 修饰符状态更新 API 的客户端。

**API 规范**：
- **Method**: `PUT`
- **URL**: `/v1/partners/{partnerId}/stores/{storeId}/menu/property/values/status`
- **Headers**: `Authorization: Bearer {token}`, `Content-Type: application/json`
- **Request Body**: `ModifierStatusUpdateReq`
- **Response Body**: `ModifierStatusUpdateResp`

**验收标准**：
- [x] 创建 `internal/client/lineman/modifier_status_client.go`
- [x] 实现 `UpdateModifierStatus(ctx, storeId, req) (*ModifierStatusUpdateResp, error)` 方法
- [x] 复用现有的 `AuthClient` 获取 Access Token
- [x] 使用 GoFrame 的 `ghttp.Client` 发送请求
- [x] HTTP 200 视为成功
- [x] 响应中 `status=ok` 视为成功
- [x] 记录请求和响应日志
- [x] 支持重试机制（复用 `WithRetry`）

---

### Requirement 5: Controller 层字段校验与平台路由

**描述**：在 Controller 层根据 `provider_name` 进行平台路由，并对 Lineman 请求进行字段校验。

**字段校验规则（Lineman）**：
- ✅ 允许字段：`merchant_id`, `modifier_id`, `modifier_name`, `available_status`, `provider_name`
- ❌ 禁止字段：`price`, `is_free`, `advanced_pricings`

**路由逻辑**：
```go
switch providerName {
case "grab":
    return c.handleGrabUpdate(ctx, req)
case "lineman":
    return c.handleLinemanUpdate(ctx, req)
default:
    return error("不支持的平台")
}
```

**验收标准**：
- [x] 在 `UpdateMenuModifier` 方法中获取 `provider_name`（默认 `grab`）
- [x] 根据 `provider_name` 路由到对应的处理方法
- [x] 实现 `handleLinemanUpdate(ctx, req)` 方法
- [x] 实现 `validateLinemanModifierRequest(req)` 字段校验函数
- [x] Lineman 请求中包含 `price`、`is_free`、`advanced_pricings` 字段时返回错误
- [x] Lineman 请求中 `available_status` 为空时返回错误
- [x] Grab 现有逻辑不受影响（向后兼容）

---

### Requirement 6: Service 层集成

**描述**：在 Service 层实现 Lineman 修饰符状态更新业务逻辑。

**业务流程**：
1. 接收 `merchantId`（对应 Lineman storeId）、`modifierId`、`linemanStatus`（int）
2. 构造 `ModifierStatusUpdateReq`
3. 调用 Lineman Client 进行 API 调用
4. 检查响应状态
5. 返回结果

**验收标准**：
- [x] 在 `internal/logic/lineman/lineman.go` 中实现 `UpdateModifierStatus` 方法
- [x] 方法签名：`UpdateModifierStatus(ctx, merchantId, modifierId, status int) error`
- [x] 参数校验：`merchantId`、`modifierId` 不能为空，`status` 必须为 1、2 或 3
- [x] 调用 Lineman Client 进行 API 调用
- [x] 检查响应 `status=ok`
- [x] 错误处理：记录日志，返回明确错误信息

---

## ⚠️ 非功能性需求

### 性能要求

- **API 响应时间**: ≤ 2s（P99）
- **并发支持**: 支持 100 QPS
- **重试机制**: 最多重试 3 次，间隔 1s、2s、4s（指数退避）

### 安全要求

- **认证**: 使用 Lineman OAuth 2.0 Token
- **数据校验**: 严格校验 `status` 枚举值（1, 2, 3）
- **错误处理**: 不暴露内部实现细节

### 可维护性

- **代码复用**: 复用现有的 Lineman Client 认证逻辑
- **日志记录**: 记录请求/响应、错误、重试信息
- **测试覆盖**: 单元测试 + 集成测试

---

## 🎨 用户体验

### User Story

**作为** 商户管理员  
**我想** 在 TTPOS 后台统一管理修饰符状态，自动同步到 Grab 和 Lineman 平台  
**以便于** 无需在多个平台分别操作，提高管理效率，减少状态不一致

### 验收标准（Acceptance Criteria）

#### AC 1: 平台识别与路由

**GIVEN** 调用 `UpdateMenuModifierReq`  
**WHEN** `provider_name` 未指定  
**THEN** 系统 **SHALL** 默认使用 `grab` 处理逻辑

**GIVEN** 调用 `UpdateMenuModifierReq`  
**WHEN** `provider_name=lineman`  
**THEN** 系统 **SHALL** 使用 Lineman 处理逻辑

#### AC 2: 状态映射

**GIVEN** `provider_name=lineman` 且 `available_status="AVAILABLE"`  
**WHEN** 调用 `UpdateMenuModifierReq`  
**THEN** 系统 **SHALL** 将状态转换为 Lineman `status=1` (AVAILABLE)

**GIVEN** `provider_name=lineman` 且 `available_status="UNAVAILABLE"`  
**WHEN** 调用 `UpdateMenuModifierReq`  
**THEN** 系统 **SHALL** 将状态转换为 Lineman `status=3` (SUSPENDED)

**GIVEN** `provider_name=lineman` 且 `available_status="SOLD_OUT_TODAY"`  
**WHEN** 调用 `UpdateMenuModifierReq`  
**THEN** 系统 **SHALL** 将状态转换为 Lineman `status=2` (SOLD_OUT_TODAY)

#### AC 3: 字段校验

**GIVEN** `provider_name=lineman`  
**WHEN** 请求中包含 `price` 字段  
**THEN** 系统 **SHALL** 返回错误 "Lineman 平台仅支持更新 available_status 字段，不支持 price 字段"

**GIVEN** `provider_name=lineman`  
**WHEN** 请求中 `available_status` 为空  
**THEN** 系统 **SHALL** 返回错误 "available_status 字段为必填"

#### AC 4: 向后兼容

**GIVEN** 调用 `UpdateMenuModifierReq` 且未指定 `provider_name`  
**WHEN** 请求包含任意字段组合  
**THEN** 系统 **SHALL** 按现有 Grab 逻辑处理（不受影响）

#### AC 5: 错误处理

**GIVEN** `provider_name=lineman` 且 `available_status="INVALID_STATUS"`  
**WHEN** 调用 `UpdateMenuModifierReq`  
**THEN** 系统 **SHALL** 返回错误 "不支持的状态: INVALID_STATUS"

**GIVEN** Lineman API 调用失败（如网络错误）  
**WHEN** 调用 `UpdateMenuModifierReq`  
**THEN** 系统 **SHALL** 自动重试最多 3 次，并记录错误日志

---

## 🔧 技术约束

### 技术栈

- **Protobuf**: gRPC 协议定义
- **Go**: GoFrame 2.x 框架
- **Lineman API**: RESTful API（OAuth 2.0 认证）

### 模块依赖

- **ttpos-bmp/ttpos-takeout**: BMP 微服务
- **Protobuf**: `manifest/protobuf/menu/menu.proto`
- **Lineman Client**: 复用现有的 `internal/client/lineman/`

### 开发规范

- **Protobuf 规范**: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- **Go 开发规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`
- **API 设计规范**: `.cursor/rules/api.mdc`

---

## 📊 影响分析

### 涉及终端

- [x] **Shop 商家管理端**：菜单管理功能
- [ ] POS 收银端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

### 涉及模块

- [x] **Protobuf 协议**：`menu.proto`
- [x] **BMP 微服务**：`ttpos-takeout`
- [x] **Lineman Client**：API 客户端
- [x] **状态映射**：Business Logic
- [ ] UI 组件
- [ ] 数据库

### 数据变更

- [ ] 无数据库变更
- [ ] 无数据迁移

---

## 🚧 风险与限制

### 潜在风险

1. **类型转换错误**：string → int 映射错误导致 Lineman API 调用失败
   - **缓解**: 使用 map 进行严格枚举校验，未知状态返回错误

2. **Lineman API 调用失败**：网络问题、认证失败、限流
   - **缓解**: 实现重试机制，记录详细日志

3. **状态映射错误**：导致修饰符错误上下架
   - **缓解**: 单元测试覆盖所有状态映射场景

4. **向后兼容问题**：现有 Grab 调用方受影响
   - **缓解**: 默认值为 `grab`，现有调用方无需修改代码

### 已知限制

1. **Lineman 仅支持状态更新**：不支持价格、高级定价等字段
2. **状态枚举差异**：需要维护 TTPOS → Lineman 的状态映射表
3. **API 调用延迟**：第三方 API 调用可能影响响应时间

---

## 📚 参考资料

### API 文档

- **Lineman API**: [Google Sheets - Update Menu Property Values Status API](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=1934684079#gid=1934684079)
- **Grab Menu API**: ttpos-bmp 现有实现

### 相关文档

- **Protobuf 规范**: `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- **Go 开发规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`
- **API 设计规范**: `.cursor/rules/api.mdc`
- **Story Point 规范**: `.cursor/rules/scrum_story_point.mdc`

### 相关 Spec

- **参考 Spec**: [tech-takeout-menu-provider-lineman-sync](../tech-takeout-menu-provider-lineman-sync/requirements.md)（Menu Item 多平台支持）

---

## 📝 审核记录

### 审核状态

- **当前状态**: ✅ 已通过 - 进入技术设计阶段
- **创建日期**: 2026-01-13
- **最后更新**: 2026-01-13
- **审核通过日期**: 2026-01-13

### 审核意见

| 日期 | 审核人 | 角色 | 意见 | 状态 |
|------|--------|------|------|------|
| - | - | - | - | - |

### 审核通过标准

- [ ] **产品审核**：需求清晰，业务价值明确
- [ ] **技术审核**：技术方案可行，风险可控
- [ ] **测试审核**：验收标准明确，可测试

---

**版本**: v1.0.0  
**创建日期**: 2026-01-13  
**最后更新**: 2026-01-13  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/specs.mdc`, `.cursor/rules/scrum_story_point.mdc`
