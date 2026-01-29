# LINE MAN OAuth Access Token 缓存功能 需求文档

> 本文档定义 LINE MAN OAuth Access Token 缓存功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.13.1-lineman-access-token.md](../../../../team/proposals/2026-01/v2.13.1-lineman-access-token.md) |
| **创建日期**      | 2026-01-07                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint N                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | rikugun             |
| **审核日期** | 2026-01-07             |
| **审核意见** | 需求定义清晰，技术方案参考 Grab 实现降低风险，可进入设计阶段         |

---

## 📋 概述

实现 LINE MAN OAuth Access Token 的自动获取与 Redis 缓存机制，参考 Grab 平台的成熟实现，减少不必要的网络请求，提升系统性能和稳定性。

**核心价值**：
- 通过 Redis 缓存减少 90% 以上的 Token 请求次数
- 缓存命中时响应时间从 100-300ms 降低到 1-5ms
- 避免频繁请求导致的限流风险
- 统一化 Token 管理，与 Grab 平台保持架构一致性

## 🎯 产品对齐

该功能支持 LINE MAN 外卖平台集成的性能优化目标，通过技术手段提升系统效率，为业务扩展提供坚实的技术基础。与 Grab 平台集成保持相同的架构模式，降低团队维护成本。

## 📝 用户故事

**作为** 外卖平台集成开发人员  
**我想** 在调用 LINE MAN API 时自动获取并缓存 Access Token  
**以便于** 减少网络请求次数，提升系统性能和稳定性

---

## 功能需求

### Requirement 1: OAuth Access Token 获取

**用户故事**: 作为系统，我想从 LINE MAN OAuth 服务器获取 Access Token，以便于调用 LINE MAN API

#### 验收标准

1. **WHEN** 系统首次调用 `FetchTokenFromAPI()` **THEN** 系统 **SHALL** 向 LINE MAN OAuth 服务器发送 POST 请求，携带 client_id、client_secret 和 grant_type
2. **WHEN** OAuth 服务器返回成功响应 **THEN** 系统 **SHALL** 解析并返回 access_token 和 expires_in
3. **WHEN** OAuth 请求失败 **THEN** 系统 **SHALL** 返回包含详细错误信息的 error
4. **WHEN** 响应缺少必需字段 **THEN** 系统 **SHALL** 返回明确的错误信息

#### 具体要求

- [x] 1.1 实现 `FetchTokenFromAPI()` 方法，调用 LINE MAN OAuth API
- [x] 1.2 使用配置的 endpoint 构建 Token URL：`{endpoint}/oauth/token`
- [x] 1.3 发送 POST 请求，Content-Type 为 application/json
- [x] 1.4 请求体包含：grant_type=client_credentials, client_id, client_secret, scope
- [x] 1.5 解析响应：access_token, expires_in, token_type
- [x] 1.6 记录日志：成功时记录 INFO 日志，失败时记录 ERROR 日志
- [x] 1.7 使用 GoFrame 的 g.Client() 发送 HTTP 请求
- [x] 1.8 错误处理：网络错误、响应解析错误、字段缺失错误

---

### Requirement 2: Redis Token 缓存

**用户故事**: 作为系统，我想将获取的 Access Token 缓存到 Redis，以便于减少重复请求

#### 验收标准

1. **WHEN** 成功获取 Token **THEN** 系统 **SHALL** 将 Token 缓存到 Redis，Key 为 `lineman:oauth:token:{client_id}`
2. **WHEN** 设置缓存 TTL **THEN** 系统 **SHALL** 使用 `expires_in - 60` 秒作为 TTL
3. **WHEN** Redis 写入失败 **THEN** 系统 **SHALL** 记录警告日志但不影响业务流程
4. **WHEN** TTL 计算结果 ≤ 0 **THEN** 系统 **SHALL** 跳过缓存写入

#### 具体要求

- [x] 2.1 Redis Key 格式：`lineman:oauth:token:{client_id}`
- [x] 2.2 使用 `g.Redis().SetEX()` 设置带 TTL 的缓存
- [x] 2.3 TTL 缓冲时间：60 秒（避免边界情况）
- [x] 2.4 Redis 故障不影响业务，仅记录日志
- [x] 2.5 记录缓存写入成功的 INFO 日志

---

### Requirement 3: Token 缓存读取与双重检查锁

**用户故事**: 作为系统，我想从 Redis 读取缓存的 Token，以便于避免重复请求

#### 验收标准

1. **WHEN** 调用 `GetAccessToken()` **THEN** 系统 **SHALL** 首先尝试从 Redis 读取 Token
2. **WHEN** Redis 缓存命中且 Token 有效 **THEN** 系统 **SHALL** 直接返回缓存的 Token，不发起网络请求
3. **WHEN** Redis 缓存未命中 **THEN** 系统 **SHALL** 获取互斥锁，执行双重检查后调用 OAuth API
4. **WHEN** 并发请求同时获取 Token **THEN** 系统 **SHALL** 只有一个请求调用 OAuth API，其他请求等待并从缓存获取
5. **WHEN** 获取锁后再次检查缓存 **THEN** 系统 **SHALL** 如果缓存已存在则直接返回，避免重复请求

#### 具体要求

- [x] 3.1 实现 `GetAccessToken()` 方法，优先从 Redis 读取
- [x] 3.2 使用 `sync.Mutex` 实现互斥锁
- [x] 3.3 双重检查锁（DCL）模式：锁前检查一次，锁后再检查一次
- [x] 3.4 缓存命中时记录 DEBUG 日志
- [x] 3.5 缓存未命中时记录 INFO 日志
- [x] 3.6 获取 Token 后更新 Redis 缓存
- [x] 3.7 返回 Token 字符串

---

### Requirement 4: Authorization Header 生成

**用户故事**: 作为业务代码，我想获取格式化的 Authorization 请求头，以便于调用 LINE MAN API

#### 验收标准

1. **WHEN** 调用 `GetAuthorizationHeader()` **THEN** 系统 **SHALL** 返回格式为 `Bearer {token}` 的字符串
2. **WHEN** Token 获取失败 **THEN** 系统 **SHALL** 返回错误信息
3. **WHEN** Token 为空 **THEN** 系统 **SHALL** 返回错误信息

#### 具体要求

- [x] 4.1 实现 `GetAuthorizationHeader()` 方法
- [x] 4.2 调用 `GetAccessToken()` 获取 Token
- [x] 4.3 返回 `"Bearer " + token` 格式的字符串
- [x] 4.4 错误传递：将 `GetAccessToken()` 的错误向上传递

---

### Requirement 5: 配置支持

**用户故事**: 作为运维人员，我想通过配置管理 LINE MAN 平台的 endpoint 和凭证，以便于不同环境使用不同配置

#### 验收标准

1. **WHEN** 加载配置 **THEN** 系统 **SHALL** 从配置中读取 endpoint、client_id、client_secret、secretKey
2. **WHEN** 配置缺失必需字段 **THEN** 系统 **SHALL** 返回明确的错误信息
3. **WHEN** 环境变量变更 **THEN** 系统 **SHALL** 在重启后使用新配置

#### 具体要求

- [x] 5.1 配置模板：`manifest/config/config.tpl.yaml`
- [x] 5.2 新增配置项：`app.provider.lineman.platform.endpoint`
- [x] 5.3 数据结构：在 `Lineman` 结构体中添加 `Endpoint` 字段
- [x] 5.4 环境变量：`$LINEMAN_PLATFORM_ENDPOINT`
- [x] 5.5 Staging 默认值：`https://beta-partner-order.wndv.co`
- [x] 5.6 Production 默认值：待确认
- [x] 5.7 配置懒加载：使用 `MustConf()` 获取配置
- [x] 5.8 配置文档：创建 `lineman-env-example.md` 和 `CHANGELOG-v2.13.1.md`

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Logic → Service 分层（GoFrame 架构）
- **单一职责原则**: Token 管理逻辑集中在 `lineman_token` 包中
- **模块化设计**: 可独立使用，不依赖其他业务模块
- **依赖管理**: 仅依赖 GoFrame 框架和标准库
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] 方法命名遵循 Go 命名规范（大驼峰）
- [x] 错误处理使用 `gerror` 包
- [x] 日志记录使用 `g.Log()`
- [x] 上下文传递：所有方法接收 `context.Context` 参数

### 性能要求

- [x] Redis 缓存命中率 > 90%
- [x] 缓存命中时响应时间 < 10ms
- [x] API 调用响应时间 < 500ms
- [x] 并发安全：支持多协程并发调用
- [x] 内存优化：Token 仅缓存在 Redis，不占用应用内存

### 浏览器兼容性（管理后台）

- N/A（纯后端功能）

### 测试要求

- [x] Logic 层单元测试覆盖率 ≥ 80%
- [x] 测试覆盖场景：
  - OAuth API 调用成功/失败
  - Redis 缓存命中/未命中
  - Redis 故障降级
  - 并发安全测试
  - 配置加载测试

### 国际化要求

- N/A（后端服务，日志使用中文）

### 安全要求

- [x] client_secret 敏感信息不记录到日志
- [x] Token 存储在 Redis，设置合理的 TTL
- [x] 使用 HTTPS 调用 LINE MAN OAuth API
- [x] 配置通过环境变量注入，不硬编码

### 可靠性要求

- [x] 网络异常时返回明确的错误信息
- [x] Redis 故障时降级到直接调用 API
- [x] 错误日志记录详细的调用栈
- [x] 双重检查锁避免并发重复请求

---

## 验收标准

### 功能验收

1. **Token 获取**: 成功调用 LINE MAN OAuth API 并获取 Token
2. **缓存机制**: Token 成功缓存到 Redis，设置正确的 TTL
3. **缓存读取**: 后续请求从 Redis 读取 Token，减少 API 调用
4. **并发安全**: 并发场景下只发起一次 OAuth 请求
5. **容错处理**: Redis 故障时业务不受影响

### 测试验收

1. **单元测试**: 覆盖率达标（≥ 80%）
2. **集成测试**: 与 LINE MAN OAuth API 集成测试通过
3. **性能测试**: 缓存命中率和响应时间达标
4. **并发测试**: 并发场景下 Token 正确性和唯一性验证

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **配置文档**: `lineman-env-example.md` 和 `CHANGELOG-v2.13.1.md` 完整
3. **代码注释**: 所有导出方法包含中文注释
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 使用 `g.Redis()` 访问 Redis
- 使用 `g.Client()` 发送 HTTP 请求
- 使用 `g.Log()` 记录日志
- 使用 `gerror` 包处理错误
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### Redis 约束

- Key 命名规范：`lineman:oauth:token:{client_id}`
- TTL 计算：`expires_in - 60` 秒
- 使用 `SetEX()` 原子性设置 Key 和 TTL
- Redis 故障不影响业务

#### LINE MAN OAuth API 约束

- Grant Type: client_credentials
- Content-Type: application/json
- 响应字段：access_token, expires_in, token_type
- Token 有效期：由 LINE MAN 服务器返回（通常 1 小时）

### 业务约束

- 仅用于系统主动调用 LINE MAN API，不用于接收回调验证
- Token 仅缓存在 Redis，不持久化到数据库
- Token 过期后自动重新获取，无需手动刷新

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 核心框架
- `github.com/golang-jwt/jwt/v5` - JWT 库（已有，用于 Partner Token）
- Redis 6.0+ - Token 缓存存储

### 服务依赖

- **LINE MAN OAuth API**: `{LINEMAN_PLATFORM_ENDPOINT}/oauth/token`
- **Redis Cluster**: Token 缓存服务

### 业务依赖

- 需要 LINE MAN 提供有效的 client_id 和 client_secret
- 需要配置正确的 endpoint 地址（staging/production）

---

## 风险和缓解

### 风险 1: LINE MAN OAuth API 文档不完整

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 参考 Grab OAuth 文档和标准 OAuth 2.0 规范推断实现
- 与 LINE MAN 技术支持联系确认 API 细节
- 实现灵活的请求体和响应解析，支持字段扩展

### 风险 2: Token 提前失效

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 设置 60 秒缓冲时间，提前刷新 Token
- 如果 API 调用返回 401 Unauthorized，自动重新获取 Token
- 记录详细的日志便于排查问题

### 风险 3: Redis 故障导致服务不可用

**影响**: 低  
**概率**: 低  
**缓解措施**:

- Redis 故障时降级到直接调用 OAuth API
- 仅记录警告日志，不阻塞业务流程
- Redis 恢复后自动恢复缓存功能

### 风险 4: 并发场景下 Token 重复获取

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用双重检查锁（DCL）机制
- 互斥锁保证同一时间只有一个请求调用 OAuth API
- 锁后再次检查缓存，避免重复请求

---

## 时间表

- **Phase 1 - 配置和基础结构**: 0.25 天
  - 添加 endpoint 配置
  - 更新数据模型
  - 创建配置文档

- **Phase 2 - 核心功能实现**: 1 天
  - 实现 `FetchTokenFromAPI()` 方法
  - 实现 `GetAccessToken()` 方法（Redis 缓存 + DCL）
  - 实现 `GetAuthorizationHeader()` 方法

- **Phase 3 - 测试和文档**: 0.75 天
  - 单元测试（覆盖率 ≥ 80%）
  - 集成测试（与 LINE MAN API）
  - 更新技术文档

- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/knowledge_management.mdc` - 知识管理规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构（如有）
- `ttpos-bmp/README.MD` - BMP 模块说明

### 参考实现

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go` - Grab OAuth Token 实现
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_token/grab_token.go` - Grab Token 服务

### 配置文档

- `ttpos-bmp/app/ttpos-takeout/manifest/config/lineman-env-example.md` - 环境变量配置示例
- `ttpos-bmp/app/ttpos-takeout/manifest/config/CHANGELOG-v2.13.1.md` - 配置变更日志

### 外部参考

- OAuth 2.0 Client Credentials Flow: https://oauth.net/2/grant-types/client-credentials/
- GoFrame 官方文档: https://goframe.org.cn
- GoFrame Redis 文档: https://goframe.org/docs/components/contrib-nosql-redis

---

## Graphiti & 活动日志

- Related Episode: `[待补充 - 实现完成后记录经验]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`
- 提醒：实现完成后，将 Token 缓存策略和双重检查锁的实践经验记录到 Graphiti

---

**版本**: v1.0.0  
**创建日期**: 2026-01-07  
**作者**: rikugun  
**审核者**: 待指定

