# 快捷支付功能 任务分解

> 本文档定义快捷支付功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/20251117100000_add_quick_payment_fields.php`
  - Purpose: 为 ttpos_company 和 ttpos_order 表添加快捷支付相关字段
  - Requirements: 2.1, 3.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，design.md 中的 SQL 设计
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_company 表添加 default_payment_method 字段，为 ttpos_order 表添加 is_quick_payment 字段 | Context: 使用 PHP Phinx，必须检查字段是否存在，参考 design.md | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，包含字段存在性检查

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 2.1, 3.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [ ] 1.3 更新 Go Model

  - File: `main/app/model/company.go`, `main/app/model/order.go`
  - Purpose: 在 Go Model 中添加新字段
  - Requirements: 2.1, 3.1
  - Leverage: 现有 Model，design.md 中的 Model 定义
  - Prompt: Role: Go Developer | Task: 在 Company 和 Order 结构体中添加新字段 | Context: DefaultPaymentMethod uint8, IsQuickPayment uint8，使用 gorm 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: 核心实现（Go Main）

### DTO 层

- [ ] 2.1 创建 Request DTO

  - File: `main/app/dto/req/quick_payment_req.go`
  - Purpose: 定义快捷支付API请求参数
  - Requirements: 1.1, 1.2
  - Leverage: 现有 DTO: `main/app/dto/req/`，design.md 中的 DTO 定义
  - Prompt: Role: Go Developer | Task: 创建 QuickPaymentReq 结构体 | Context: 包含 OrderUuid (required) 和 PaymentMethod (optional) | Restrictions: 使用 binding 标签，遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 2.2 创建 Response DTO

  - File: `main/app/dto/resp/quick_payment_resp.go`
  - Purpose: 定义快捷支付API响应数据
  - Requirements: 1.3
  - Leverage: 现有 DTO: `main/app/dto/resp/`，design.md 中的 DTO 定义
  - Success: DTO 创建成功，响应格式正确

### Service 层

- [ ] 2.3 创建 Service 接口

  - File: `main/app/service/i_quick_payment_srv.go`
  - Purpose: 定义快捷支付业务逻辑接口
  - Requirements: 1.1-1.7
  - Leverage: 现有 Service 接口: `main/app/service/i_*_srv.go`
  - Prompt: Role: Go Developer | Task: 创建 IQuickPaymentSrv 接口，定义 QuickPay 方法 | Context: 方法签名: QuickPay(ctx *gin.Context, req *dto_req.QuickPaymentReq) (*dto_resp.QuickPaymentResp, error) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整

- [ ] 2.4 实现 Service 业务逻辑

  - File: `main/app/service/quick_payment_srv.go`
  - Purpose: 实现快捷支付核心逻辑
  - Requirements: 1.1-1.7
  - Leverage: 现有 Service: `main/app/service/order_srv.go`, `main/app/service/payment_srv.go`，design.md 中的完整实现代码
  - Prompt: Role: Go Developer with payment system expertise | Task: 实现 quickPaymentSrv，包含完整支付流程 | Context: 依赖 IOrderService 和 IPaymentService，使用 SystemLock 防并发，发布事件，实现缓存逻辑 | Restrictions: 不直接依赖 Repository，使用事务，不使用 panic | Success: Service 实现完整，业务逻辑正确，包含并发控制和缓存

- [ ] 2.5 编写 Service 单元测试

  - File: `main/app/service/quick_payment_srv_test.go`
  - Purpose: 测试快捷支付业务逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer | Task: 为 QuickPaymentService 编写单元测试，覆盖率 ≥ 70% | Context: 测试正常流程、订单状态异常、并发场景、缓存场景 | Restrictions: Payment 相关覆盖率 100% | Success: 测试覆盖率达标，所有测试通过

### API 层

- [ ] 2.6 创建 API Controller

  - File: `main/app/api/quick_payment_api.go`
  - Purpose: 实现快捷支付HTTP接口
  - Requirements: 1.1-1.7
  - Leverage: 现有 API: `main/app/api/*_api.go`，design.md 中的 API 实现
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 QuickPaymentAPI，实现 QuickPay 方法 | Context: URL: /api/v1/order/quick_payment, 使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 创建成功，响应格式正确

- [ ] 2.7 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册快捷支付路由
  - Requirements: 1.1
  - Leverage: 现有路由配置
  - Code: 在 order 路由组中添加: `order.POST("/quick_payment", quickPaymentAPI.QuickPay)`
  - Success: 路由注册成功

- [ ] 2.8 编写 API 集成测试

  - File: `main/app/api/quick_payment_api_test.go`
  - Purpose: 测试快捷支付API接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer | Task: 为 QuickPaymentAPI 编写集成测试 | Context: 测试正常流程、参数验证、错误处理 | Success: 所有 API 测试通过

---

## Phase 3: 测试和优化

- [ ] 3.1 端到端集成测试

  - File: `test/integration/quick_payment_test.go`
  - Purpose: 测试完整支付流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端测试 | Context: 创建订单 → 快捷支付 → 验证订单状态 → 验证支付记录 → 验证事件发布 | Success: 集成测试通过

- [ ] 3.2 并发测试

  - File: -
  - Purpose: 验证并发场景下的正确性
  - Requirements: 1.7
  - Test: 10个goroutine同时对同一订单调用快捷支付API，验证只有1次成功
  - Success: 并发测试通过，无重复支付

- [ ] 3.3 性能测试

  - File: -
  - Purpose: 验证性能指标
  - Requirements: 非功能需求
  - Tool: wrk 或 ab
  - Command: `wrk -t10 -c100 -d30s http://localhost:8080/api/v1/order/quick_payment`
  - Success: 本地响应时间 < 200ms，QPS > 1000

- [ ] 3.4 缓存验证

  - File: -
  - Purpose: 验证缓存策略
  - Requirements: 2.2
  - Test: 查看 Redis 缓存命中率，验证 > 80%
  - Success: 缓存命中率达标

- [ ] 3.5 文档更新

  - File: `docs/shared/api/order_api.md`, `CHANGELOG.md`
  - Purpose: 更新API文档和变更日志
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - QuickPaymentService: ≥ 70%
  - PaymentService: 100%（高风险）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 执行流程

1. **选择任务**: 从 Phase 1 开始，按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的实现代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### 预计时间

- Phase 1: 0.5 天（4 小时）
- Phase 2: 1.5 天（12 小时）
- Phase 3: 1 天（8 小时）
- **总计**: 3 天（24 小时）= **SP 5**

---

## 附录：AI Prompt 示例

### Service 实现

```
Role: Go Developer with payment system expertise

Task: 实现 QuickPaymentService，包含完整的快捷支付逻辑

Context:
- File: main/app/service/quick_payment_srv.go
- Leverage: design.md 中的完整实现代码
- Requirements: requirements.md Requirement 1
- Dependencies: IOrderService, IPaymentService, SystemLock

Implementation Steps:
1. 使用 SystemLock.LockUuid() 加锁
2. 调用 OrderService 获取订单
3. 验证订单状态（只允许待支付状态）
4. 获取商户配置（先查缓存，未命中查数据库）
5. 调用 PaymentService 创建支付记录
6. 调用 OrderService 更新订单状态
7. 发布订单支付完成事件（异步）
8. 返回支付结果

Restrictions:
- 不直接依赖 Repository
- 使用事务管理
- 不使用 panic
- 使用 errors.WithMessage 包装错误

Success Criteria:
- 代码通过 go fmt 和 go vet
- 业务逻辑正确
- 并发安全
- 缓存逻辑正确
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-17.md`
- 当执行任务中形成复盘/优化建议时，及时沉淀 Episode 并在本节更新名称。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-17  
**维护者**: 后端开发组

