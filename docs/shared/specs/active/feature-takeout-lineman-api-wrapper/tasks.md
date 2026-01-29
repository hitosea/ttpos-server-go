# Lineman API 包装实现 任务分解

> 本文档定义 Lineman API 包装服务的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 22  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 基础设施和配置管理

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 创建配置结构定义

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/conf/lineman.go`
  - Purpose: 定义 Lineman API 配置结构，用于读取配置文件
  - Requirements: 1.1（定义 Lineman 配置结构体）
  - Leverage: 现有配置结构: `internal/model/conf/grab.go`
  - Prompt: Role: Go Developer | Task: 创建 Lineman 配置结构体，包含 Endpoint、ApiKey、SecretKey、Environment、Timeout 字段 | Context: 参考 Grab 配置结构，使用 json 标签 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 配置结构创建成功，字段定义完整

- [ ] 1.2 实现配置读取逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/config.go`
  - Purpose: 实现 Lineman 配置读取，支持懒加载
  - Requirements: 1.2（实现配置读取逻辑）
  - Leverage: 现有配置读取: `internal/logic/grab/config.go`
  - Prompt: Role: Go Developer | Task: 实现 MustConfig() 方法，读取 app.provider.lineman.platform 配置节点 | Context: 使用 g.Cfg().MustGet(ctx, "app.provider.lineman.platform").Scan(&lineman)，配置缺失时 panic | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 配置读取成功，错误处理正确

- [ ] 1.3 创建主服务入口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman.go`
  - Purpose: 创建 Lineman Service 主入口，实现配置懒加载和服务注册
  - Requirements: 1.1, 1.2（配置管理）
  - Leverage: 现有服务实现: `internal/logic/grab/grab.go`，参考懒加载单例模式和服务注册
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 创建 sLineman 结构体，实现 MustConf() 方法（懒加载单例），实现 init() 注册服务 | Context: 使用 sync.RWMutex 控制并发，双重检查锁，调用 service.RegisterLineman(Lineman) | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，日志使用中文 | Success: 服务创建成功，配置懒加载正确，服务注册成功

- [ ] 1.4 添加 Lineman 配置到配置文件模板

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
  - Purpose: 在配置文件模板中添加 Lineman 配置节点
  - Requirements: 1.1（配置结构）
  - Leverage: 现有配置: `app.provider.grab.platform` 节点
  - Success: 配置节点添加成功，格式正确

---

## Phase 2: 认证模块

- [ ] 2.1 实现认证管理基础结构

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/auth.go`
  - Purpose: 实现 Lineman API 认证管理（Token/API Key）
  - Requirements: 1.3（认证信息管理）
  - Leverage: 现有认证实现: `internal/logic/grab/auth.go`，参考 Token 管理和 Redis 缓存
  - Prompt: Role: Go Developer with authentication expertise | Task: 实现 GetAccessToken() 方法，根据 Lineman API 认证方式（API Key 或 OAuth）获取授权信息 | Context: 如果使用 API Key，直接返回 conf.ApiKey；如果使用 OAuth，实现 Token 获取和 Redis 缓存 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 处理错误 | Success: 认证方法实现成功，Token 缓存正确（如需要）

- [ ] 2.2 实现 Webhook 签名验证

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/auth.go`
  - Purpose: 实现 Webhook 签名验证，防止伪造请求
  - Requirements: 5.5（Webhook 签名验证）
  - Leverage: 现有签名验证: `internal/logic/grab/auth.go`，参考签名算法和验证逻辑
  - Prompt: Role: Security Developer with cryptography expertise | Task: 实现 VerifyWebhookSignature() 方法，验证 Lineman Webhook 签名 | Context: 根据 Lineman API Spec 实现签名算法，验证 X-Lineman-Signature 和 X-Lineman-Timestamp 请求头 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 处理验证失败 | Success: 签名验证实现成功，安全性正确

---

## Phase 3: DTO 定义

- [ ] 3.1 定义订单相关 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/order.go`
  - Purpose: 定义订单相关的数据传输对象（DTO）
  - Requirements: 2.7（定义订单相关 DTO）
  - Leverage: 现有 DTO: `internal/model/dto/grab/order.go`，参考 Grab 订单 DTO 结构
  - Prompt: Role: Go Developer with API design expertise | Task: 根据 Lineman API Spec 定义订单 DTO，包含 ReceiveOrderReq、AcceptOrderReq、RejectOrderReq、CancelOrderReq、UpdateOrderStatusReq、MarkOrderReadyReq 等结构体 | Context: 使用 json 标签，字段命名使用小驼峰 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 定义完整，字段类型正确

- [ ] 3.2 定义菜单相关 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/menu.go`
  - Purpose: 定义菜单相关的数据传输对象（DTO）
  - Requirements: 3.4（定义菜单相关 DTO）
  - Leverage: 现有 DTO: `internal/model/dto/grab/menu.go`，参考 Grab 菜单 DTO 结构
  - Prompt: Role: Go Developer with API design expertise | Task: 根据 Lineman API Spec 定义菜单 DTO，包含 GetMenuReq、GetMenuResp、SyncMenuReq、UpdateMenuStatusReq 等结构体 | Context: 使用 json 标签，字段命名使用小驼峰 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 定义完整，字段类型正确

- [ ] 3.3 定义门店相关 DTO

  - File: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/lineman/store.go`
  - Purpose: 定义门店相关的数据传输对象（DTO）
  - Requirements: 4.5（定义门店相关 DTO）
  - Leverage: 现有 DTO: `internal/model/dto/grab/store.go`，参考 Grab 门店 DTO 结构
  - Prompt: Role: Go Developer with API design expertise | Task: 根据 Lineman API Spec 定义门店 DTO，包含 GetStoreStatusReq、GetStoreStatusResp、PauseStoreReq、ResumeStoreReq、UpdateStoreInfoReq 等结构体 | Context: 使用 json 标签，字段命名使用小驼峰 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 定义完整，字段类型正确

---

## Phase 4: 订单管理实现

- [ ] 4.1 实现订单接收 Webhook 处理

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 实现 Lineman 订单接收 Webhook 处理逻辑
  - Requirements: 2.1（实现订单接收 Webhook 处理）
  - Leverage: 现有订单处理: `internal/logic/grab/grab.go::HandleSubmitOrder`，参考 Webhook 处理流程
  - Prompt: Role: Go Developer with webhook expertise | Task: 实现 HandleReceiveOrder() 方法，处理 Lineman 订单接收 Webhook | Context: 解析订单数据，记录日志，返回结构化响应 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 处理错误 | Success: Webhook 处理成功，日志记录完整

- [ ] 4.2 实现接受订单 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 调用 Lineman API 接受订单
  - Requirements: 2.2（实现接受订单 API）
  - Leverage: 现有订单操作: `internal/logic/grab/grab.go::AcceptOrder`，参考 HTTP 请求封装
  - Prompt: Role: Go Developer with HTTP client expertise | Task: 实现 AcceptOrder() 方法，调用 Lineman API 接受订单 | Context: 构造 HTTP 请求，携带认证信息，处理响应，记录日志 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 gerror 包装错误 | Success: API 调用成功，错误处理正确

- [ ] 4.3 实现拒绝订单 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 调用 Lineman API 拒绝订单
  - Requirements: 2.3（实现拒绝订单 API）
  - Leverage: 现有订单操作: `internal/logic/grab/grab.go::RejectOrder`
  - Success: API 调用成功，日志记录完整

- [ ] 4.4 实现取消订单 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 调用 Lineman API 取消订单
  - Requirements: 2.4（实现取消订单 API）
  - Leverage: 现有订单操作: `internal/logic/grab/grab.go::CancelOrder`
  - Success: API 调用成功，日志记录完整

- [ ] 4.5 实现更新订单状态 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 调用 Lineman API 更新订单状态
  - Requirements: 2.5（实现更新订单状态 API）
  - Leverage: 现有订单操作: `internal/logic/grab/grab.go::UpdateDeliveryState`
  - Success: API 调用成功，状态更新正确

- [ ] 4.6 实现标记订单准备完成 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/order.go`
  - Purpose: 调用 Lineman API 标记订单准备完成
  - Requirements: 2.6（实现标记订单准备完成 API）
  - Leverage: 现有订单操作: `internal/logic/grab/grab.go::MarkOrderReady`
  - Success: API 调用成功，日志记录完整

---

## Phase 5: 菜单管理实现

- [ ] 5.1 实现获取菜单 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu.go`
  - Purpose: 调用 Lineman API 获取菜单
  - Requirements: 3.1（实现获取菜单 API）
  - Leverage: 现有菜单操作: `internal/logic/grab/grab.go::HandleGetMenu`，参考菜单数据处理
  - Prompt: Role: Go Developer | Task: 实现 GetMenu() 方法，调用 Lineman API 获取菜单数据 | Context: 构造 HTTP 请求，解析菜单响应，返回结构化数据 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: API 调用成功，菜单数据解析正确

- [ ] 5.2 实现同步菜单 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu.go`
  - Purpose: 调用 Lineman API 同步菜单数据
  - Requirements: 3.2（实现同步菜单 API）
  - Leverage: 现有菜单操作: `internal/logic/grab/grab.go::SyncMenu`，参考菜单同步逻辑
  - Success: API 调用成功，菜单同步正确

- [ ] 5.3 实现更新菜单状态 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu.go`
  - Purpose: 调用 Lineman API 更新菜单状态（启用/禁用）
  - Requirements: 3.3（实现更新菜单状态 API）
  - Leverage: 现有菜单操作: `internal/logic/grab/grab.go::UpdateMenuRecord`
  - Success: API 调用成功，状态更新正确

---

## Phase 6: 门店管理实现

- [ ] 6.1 实现获取门店状态 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/store.go`
  - Purpose: 调用 Lineman API 获取门店状态
  - Requirements: 4.1（实现获取门店状态 API）
  - Leverage: 现有门店操作: `internal/logic/grab/grab.go::GetStoreStatus`，参考门店状态查询
  - Prompt: Role: Go Developer | Task: 实现 GetStoreStatus() 方法，调用 Lineman API 获取门店状态 | Context: 构造 HTTP 请求，解析门店状态响应 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: API 调用成功，门店状态解析正确

- [ ] 6.2 实现暂停门店营业 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/store.go`
  - Purpose: 调用 Lineman API 暂停门店营业
  - Requirements: 4.2（实现暂停门店营业 API）
  - Leverage: 现有门店操作: `internal/logic/grab/grab.go::PauseStore`，参考门店状态更新
  - Success: API 调用成功，门店暂停正确

- [ ] 6.3 实现恢复门店营业 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/store.go`
  - Purpose: 调用 Lineman API 恢复门店营业
  - Requirements: 4.3（实现恢复门店营业 API）
  - Leverage: 现有门店操作: `internal/logic/grab/grab.go::ResumeStore`
  - Success: API 调用成功，门店恢复正确

- [ ] 6.4 实现更新门店信息 API

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/store.go`
  - Purpose: 调用 Lineman API 更新门店信息
  - Requirements: 4.4（实现更新门店信息 API）
  - Leverage: 现有门店操作，参考门店信息更新逻辑
  - Success: API 调用成功，门店信息更新正确

---

## Phase 7: 测试

- [ ] 7.1 编写单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/lineman_test.go`
  - Purpose: 确保 Lineman Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `internal/logic/grab/*_test.go`，参考 Mock 数据和测试方法
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Lineman Service 编写单元测试，覆盖率 ≥ 70% | Context: 测试配置读取、认证管理、订单/菜单/门店 API 调用，使用 Mock HTTP 客户端 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 7.2 Staging 环境集成测试

  - File: -
  - Purpose: 使用 Lineman Staging 环境验证实际 API 调用
  - Requirements: 所有功能需求
  - Leverage: Lineman Staging 环境配置
  - Success: 所有 API 调用测试通过，日志记录完整

---

## Phase 8: 服务注册和文档

- [ ] 8.1 重新生成 Service 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/lineman.go`
  - Purpose: 生成 Lineman Service 接口定义
  - Requirements: 所有功能需求
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen service`
  - Success: Service 接口生成成功，方法定义完整

- [ ] 8.2 更新 API 文档（如需要）

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 记录 Lineman API 包装服务的使用方法
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] requirements.md 和 design.md 保持同步

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`
- [ ] 参考 Grab 和 Skootar 实现模式
- [ ] 日志描述使用中文
- [ ] 错误处理使用 `gerror`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/feature-takeout-lineman-api-wrapper/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/feature-takeout-lineman-api-wrapper/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/feature-takeout-lineman-api-wrapper/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/feature-takeout-lineman-api-wrapper/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/feature-takeout-lineman-api-wrapper/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc

Restrictions:
- 使用 GoFrame v2.x 框架
- 使用 gerror 进行错误处理
- 使用 g.Log() 记录日志
- 日志描述使用中文
- 不对外提供 gRPC 服务
- 返回参数类型不能是 takeout.ApiResponse
- 参考 Grab 和 Skootar 实现模式

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- Mock HTTP 客户端测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-07  
**维护者**: rikugun

