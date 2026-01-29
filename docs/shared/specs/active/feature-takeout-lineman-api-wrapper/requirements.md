# Lineman API 包装实现 需求文档

> 本文档定义 Lineman API 包装服务的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/lineman-api-wrapper.md](../../../../team/proposals/2026-01/lineman-api-wrapper.md) |
| **创建日期**      | 2026-01-07                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-07             |
| **审核意见** | 技术实现类需求，参考现有 Grab/Skootar 实现，需求明确，批准进入设计阶段 |

---

## 📋 概述

实现 Lineman 外卖平台 API 的包装服务，为 ttpos-takeout 模块提供统一的第三方平台 API 调用接口。参考现有的 GrabFood 和 Skootar API 包装实现模式，在 `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman` 目录下实现完整的 Lineman API 包装层。

该功能不对外直接提供 gRPC 服务，仅在内部 Service 层实现业务逻辑，供 ttpos-takeout 模块内部调用。

## 🎯 产品对齐

- **扩展外卖平台覆盖**: 支持泰国重要的 Lineman 外卖配送平台，完善系统的外卖平台生态
- **统一集成模式**: 遵循现有的 Grab/Skootar 集成模式，降低维护成本
- **提升市场竞争力**: 增强系统对泰国市场的支持能力

## 📝 用户故事

**作为** 外卖系统开发者  
**我想** 实现 Lineman API 的包装服务  
**以便于** 业务层能够统一调用 Lineman 平台的订单、菜单、门店等功能

---

## 功能需求

### Requirement 1: 认证与配置管理

**用户故事**: 作为开发者，我想配置 Lineman API 的认证信息，以便于系统能够正常调用 Lineman API

#### 验收标准

1. **WHEN** 系统启动时 **THEN** 系统 **SHALL** 从配置文件中正确读取 Lineman API 配置（Endpoint、API Key、Secret Key 等）
2. **IF** 配置缺失或无效 **THEN** 系统 **SHALL** 在首次调用时返回明确的错误信息
3. **WHEN** 调用需要认证的 API **THEN** 系统 **SHALL** 正确携带认证信息（Token/签名）

#### 具体要求

- [x] 1.1 定义 Lineman 配置结构体（`internal/model/conf/lineman.go`）
- [x] 1.2 实现配置读取逻辑（参考 Grab 的 `MustConf()` 方法）
- [x] 1.3 实现认证信息管理（Token 管理、请求签名）
- [x] 1.4 实现请求超时配置（支持自定义超时时间）

---

### Requirement 2: 订单管理 API

**用户故事**: 作为开发者，我想调用 Lineman 订单管理 API，以便于处理订单接收、接受/拒绝、取消、状态更新等操作

#### 验收标准

1. **WHEN** 接收到 Lineman 订单 Webhook **THEN** 系统 **SHALL** 正确解析订单数据并返回结构化响应
2. **WHEN** 调用接受订单 API **THEN** 系统 **SHALL** 向 Lineman 发送正确的请求并处理响应
3. **WHEN** 调用拒绝订单 API **THEN** 系统 **SHALL** 向 Lineman 发送拒绝原因并记录日志
4. **WHEN** 调用取消订单 API **THEN** 系统 **SHALL** 向 Lineman 发送取消原因并处理响应
5. **WHEN** 调用更新订单状态 API **THEN** 系统 **SHALL** 正确更新订单状态并记录日志
6. **WHEN** 标记订单准备完成 **THEN** 系统 **SHALL** 通知 Lineman 订单可以配送

#### 具体要求

- [x] 2.1 实现订单接收 Webhook 处理（`order.go::ReceiveOrder`）
- [x] 2.2 实现接受订单 API（`order.go::AcceptOrder`）
- [x] 2.3 实现拒绝订单 API（`order.go::RejectOrder`）
- [x] 2.4 实现取消订单 API（`order.go::CancelOrder`）
- [x] 2.5 实现更新订单状态 API（`order.go::UpdateOrderStatus`）
- [x] 2.6 实现标记订单准备完成 API（`order.go::MarkOrderReady`）
- [x] 2.7 定义订单相关 DTO（`internal/model/dto/lineman/order.go`）

---

### Requirement 3: 菜单管理 API

**用户故事**: 作为开发者，我想调用 Lineman 菜单管理 API，以便于同步菜单、更新菜单状态

#### 验收标准

1. **WHEN** 调用获取菜单 API **THEN** 系统 **SHALL** 正确获取当前菜单数据
2. **WHEN** 调用同步菜单 API **THEN** 系统 **SHALL** 正确构造菜单数据并发送到 Lineman
3. **WHEN** 调用更新菜单状态 API **THEN** 系统 **SHALL** 正确更新菜单状态（启用/禁用）

#### 具体要求

- [x] 3.1 实现获取菜单 API（`menu.go::GetMenu`）
- [x] 3.2 实现同步菜单 API（`menu.go::SyncMenu`）
- [x] 3.3 实现更新菜单状态 API（`menu.go::UpdateMenuStatus`）
- [x] 3.4 定义菜单相关 DTO（`internal/model/dto/lineman/menu.go`）

---

### Requirement 4: 门店管理 API

**用户故事**: 作为开发者，我想调用 Lineman 门店管理 API，以便于管理门店状态（营业/暂停）

#### 验收标准

1. **WHEN** 调用获取门店状态 API **THEN** 系统 **SHALL** 正确获取当前门店状态
2. **WHEN** 调用暂停门店营业 API **THEN** 系统 **SHALL** 向 Lineman 发送暂停请求并记录日志
3. **WHEN** 调用恢复门店营业 API **THEN** 系统 **SHALL** 向 Lineman 发送恢复请求并记录日志
4. **WHEN** 调用更新门店信息 API **THEN** 系统 **SHALL** 向 Lineman 同步最新的门店信息

#### 具体要求

- [x] 4.1 实现获取门店状态 API（`store.go::GetStoreStatus`）
- [x] 4.2 实现暂停门店营业 API（`store.go::PauseStore`）
- [x] 4.3 实现恢复门店营业 API（`store.go::ResumeStore`）
- [x] 4.4 实现更新门店信息 API（`store.go::UpdateStoreInfo`）
- [x] 4.5 定义门店相关 DTO（`internal/model/dto/lineman/store.go`）

---

### Requirement 5: 通用工具和错误处理

**用户故事**: 作为开发者，我想使用统一的 HTTP 请求封装和错误处理机制，以便于提高代码质量和可维护性

#### 验收标准

1. **WHEN** API 调用失败 **THEN** 系统 **SHALL** 使用 `gerror` 包装错误并记录详细日志
2. **WHEN** HTTP 请求超时 **THEN** 系统 **SHALL** 返回明确的超时错误信息
3. **WHEN** 响应解析失败 **THEN** 系统 **SHALL** 记录原始响应并返回解析错误
4. **WHEN** 关键操作执行时 **THEN** 系统 **SHALL** 使用 GoFrame 日志系统记录操作日志

#### 具体要求

- [x] 5.1 封装 HTTP 请求工具（支持 GET/POST/PUT/DELETE）
- [x] 5.2 实现统一的响应解析逻辑
- [x] 5.3 实现统一的错误处理（使用 `gerror.Wrapf`）
- [x] 5.4 实现关键操作日志记录（使用 `g.Log()`）
- [x] 5.5 实现 Webhook 签名验证（如 Lineman API 要求）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 GoFrame 的架构模式（Service → DTO）
- **单一职责原则**: 每个文件应有单一、明确的目的（order/menu/store 独立实现）
- **模块化设计**: Service 应独立且可复用，便于其他模块调用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - Takeout 模块规范

### API 设计要求

- [x] 不对外提供 gRPC 服务，仅内部 Service 层实现
- [x] DTO 定义在 `internal/model/dto/lineman/` 目录
- [x] 使用 JSON 标签进行序列化/反序列化
- [x] 参考 Grab 和 Skootar 的 API 包装实现模式

### 性能要求

- [x] HTTP 请求超时时间可配置（默认 30 秒）
- [x] 使用连接池复用 HTTP 连接
- [x] 避免不必要的数据转换和复制

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] 单元测试覆盖核心方法（订单/菜单/门店 API）
- [x] Mock Lineman API 响应进行测试
- [x] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 安全要求

- [x] 所有 API 调用需要携带认证信息
- [x] Webhook 签名验证（防止伪造请求）
- [x] 敏感配置（API Key/Secret Key）不得硬编码
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时返回明确的错误信息
- [x] HTTP 请求超时重试机制（可选，根据 Lineman API 特性）
- [x] 错误日志记录（使用 `g.Log().Errorf`）
- [x] 关键操作日志记录（订单接受/拒绝、菜单同步等）

---

## 验收标准

### 功能验收

1. **认证配置**: 系统能够正确读取 Lineman API 配置并携带认证信息
2. **订单管理**: 能够正确处理订单接收、接受/拒绝、取消、状态更新等操作
3. **菜单管理**: 能够正确获取、同步、更新菜单状态
4. **门店管理**: 能够正确获取、暂停/恢复门店营业状态
5. **错误处理**: API 调用失败时能够返回明确的错误信息并记录日志

### 测试验收

1. **单元测试**: Service 层测试覆盖率 ≥ 70%
2. **集成测试**: 使用 Mock 数据测试所有 API 调用流程
3. **手动测试**: 使用 Lineman Staging 环境验证实际 API 调用

### 文档验收

1. **技术文档**: design.md 完整且准确，包含目录结构、DTO 定义、API 调用示例
2. **API 文档**: 内部 Service 方法文档完整（godoc 注释）
3. **配置文档**: 配置项说明完整（参数说明、默认值、环境要求）
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x 框架
- 遵循 GoFrame 的目录结构和命名规范
- 使用 `gerror` 进行错误处理
- 使用 `g.Log()` 记录日志
- 不对外提供 gRPC 服务，仅内部 Service 层实现
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### 参考实现

- 必须参考 `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/` 的实现模式
- 必须参考 `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/` 的实现模式
- 保持代码风格和目录结构的一致性

### 业务约束

- 仅支持 Lineman 平台 API（不涉及其他平台）
- API Spec 来源于 `docs/others/lineman/API Spec (Master).xlsx`
- 仅实现必要的 API（订单/菜单/门店管理），其他功能按需扩展

### 资源约束

- 开发时间: 3-5 天
- Story Point: 5-8 (待技术评审确认)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2`: GoFrame 框架
- `github.com/gogf/gf/v2/errors/gerror`: 错误处理
- `github.com/gogf/gf/v2/frame/g`: 日志系统
- `net/http`: HTTP 客户端

### 服务依赖

- 无 gRPC 服务依赖（仅内部 Service 层实现）
- 依赖 Lineman API 外部服务

### 业务依赖

- 需要 Lineman API 的认证信息（API Key、Secret Key）
- 需要 Lineman Staging/Production 环境的访问权限
- 依赖 `docs/others/lineman/API Spec (Master).xlsx` 中的 API 规格说明

---

## 风险和缓解

### 风险 1: API Spec 文档理解偏差

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 仔细阅读 `API Spec (Master).xlsx` 文档，必要时与产品/业务确认
- 参考 Grab 和 Skootar 的实现，推断 Lineman API 的常见模式
- 在 Staging 环境充分测试，验证理解的正确性

### 风险 2: Lineman API 认证机制复杂

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 参考 Grab 的认证实现，设计灵活的认证接口
- 预留扩展接口，支持 Token、签名等多种认证方式
- 在 Staging 环境充分测试认证流程

### 风险 3: Webhook 签名验证机制不明确

**影响**: 低  
**概率**: 中  
**缓解措施**:

- 查阅 Lineman API 文档，了解签名算法和验证流程
- 预留 Webhook 签名验证的扩展接口
- 在 Staging 环境测试 Webhook 签名验证

### 风险 4: 缺少官方 SDK

**影响**: 低  
**概率**: 高  
**缓解措施**:

- 封装通用的 HTTP 请求工具，便于维护和测试
- 参考 Grab/Skootar 的 HTTP 请求封装实现
- 使用 Mock 数据进行单元测试，降低对外部 API 的依赖

---

## 时间表

- **Phase 1 - API 规格分析和 DTO 定义**: 0.5 天
- **Phase 2 - 配置管理和认证模块**: 0.5 天
- **Phase 3 - 订单管理 API 实现**: 1-1.5 天
- **Phase 4 - 菜单管理 API 实现**: 0.5-1 天
- **Phase 5 - 门店管理 API 实现**: 0.5 天
- **Phase 6 - 单元测试和集成测试**: 1 天
- **总计**: 3-5 天（SP = 5-8）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc` - Takeout 模块规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/` - Grab API 包装实现
- `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/` - Skootar API 包装实现

### API 文档

- `docs/others/lineman/API Spec (Master).xlsx` - Lineman API 规格说明

### 外部参考

- GoFrame 官方文档: https://goframe.org.cn

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: 待审核

