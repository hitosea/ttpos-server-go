# 收银端满减营销功能 任务分解

> 本文档定义 收银端满减营销功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 28  
**已完成**: 19  
**进行中**: -  
**完成率**: 67.9%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件 - sale_order 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_full_reduction_activity_to_sale_order_table.php`
  - Purpose: 在 sale_order 表中增加满减活动相关字段
  - Requirements: 4.2, 4.3, 4.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_sale_order 表中增加 activity_amount, full_reduction_activity_uuid, full_reduction_activity_message 字段 | Context: activity_amount 使用 decimal(20,8)，full_reduction_activity_uuid 使用 bigint unsigned，full_reduction_activity_message 使用 varchar(255)，所有字段默认值为 0 或空字符串，添加索引 idx_full_reduction_activity_uuid | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [x] 1.2 创建数据库迁移文件 - sale_bill 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_full_reduction_activity_to_sale_bill_table.php`
  - Purpose: 在 sale_bill 表中增加满减活动抵扣金额字段
  - Requirements: 4.2
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_sale_bill 表中增加 activity_amount 字段 | Context: activity_amount 使用 decimal(20,8)，默认值为 0 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功

- [ ] 1.3 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建字段
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1, 1.2 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已创建

- [x] 1.4 更新 Go Model - SaleOrder

  - File: `main/app/model/sale_order.go`
  - Purpose: 在 SaleOrder 结构体中增加满减活动相关字段
  - Requirements: 4.2, 4.3, 4.4
  - Leverage: 现有 Model: `main/app/model/sale_order.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 SaleOrder 结构体中增加 ActivityAmount, FullReductionActivityUuid, FullReductionActivityMessage 字段 | Context: 使用 gorm 标签，字段类型与数据库一致，添加注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

- [x] 1.5 更新 Go Model - SaleBill

  - File: `main/app/model/sale_bill.go`
  - Purpose: 在 SaleBill 结构体中增加满减活动抵扣金额字段
  - Requirements: 4.2
  - Leverage: 现有 Model: `main/app/model/sale_bill.go`，迁移文件: Task 1.2
  - Prompt: Role: Go Developer | Task: 在 SaleBill 结构体中增加 ActivityAmount 字段 | Context: 使用 gorm 标签，字段类型与数据库一致 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功

---

## Phase 2: 核心实现（Go Main）

### DTO 层

- [x] 2.1 创建 Request DTO - 活动选择

  - File: `main/app/dto/req/order.go`
  - Purpose: 定义活动选择请求参数
  - Requirements: 2.1, 2.2
  - Leverage: 现有 DTO: `main/app/dto/req/order.go`，参考 `InstantOrderPaymentCouponReq`
  - Prompt: Role: Go Developer | Task: 在 order.go 中创建 InstantOrderPaymentActivityReq 结构体 | Context: 包含 SaleOrderUuid 和 FullReductionActivityUuid 字段，使用 binding 标签验证参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [x] 2.2 创建 Response DTO - 活动列表

  - File: `main/app/dto/resp/order.go`
  - Purpose: 定义活动列表响应结构
  - Requirements: 1.1, 1.6
  - Leverage: 现有 DTO: `main/app/dto/resp/order.go`，参考优惠券列表结构
  - Prompt: Role: Go Developer | Task: 在 order.go 中创建 FullReductionActivityItem 和 ActivityRule 结构体，扩展 InstantOrderPaymentInfoResp | Context: FullReductionActivityItem 包含活动所有信息，ActivityRule 包含阈值和减价金额，InstantOrderPaymentInfoResp 增加 ActivityList 字段 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

### Service 层

- [x] 2.3 实现活动列表查询逻辑

  - File: `main/app/service/order_pay.go`
  - Purpose: 实现获取满减活动列表的业务逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有 Service: `main/app/service/order_pay.go:InstantOrderPaymentInfo()`，参考优惠券列表查询逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 getFullReductionActivityList 方法，查询有效日期内的活动，判断活动是否在适用时段内，判断订单金额是否达到满减条件，判断活动是否已选中，排序：可用时间范围内显示在前 | Context: 需要查询活动数据（可能需要通过 Repository 或 API），判断活动可用性，计算活动抵扣金额（如果已选中），返回活动列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 活动列表查询逻辑正确，可用性判断正确，排序正确

- [x] 2.4 扩展 OrderPaymentInfo 接口

  - File: `main/app/service/order_pay.go`
  - Purpose: 在获取结账页面信息时返回活动列表
  - Requirements: 1.1
  - Leverage: 现有 Service: `main/app/service/order_pay.go:InstantOrderPaymentInfo()`，Task 2.3
  - Prompt: Role: Go Developer | Task: 在 InstantOrderPaymentInfo 方法中调用 getFullReductionActivityList，将活动列表添加到响应中 | Context: 在返回响应前调用活动列表查询，将结果添加到 resp.ActivityList | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口扩展成功，活动列表正确返回

- [x] 2.5 实现活动选择/取消逻辑

  - File: `main/app/service/order_pay.go`
  - Purpose: 实现选择或取消满减活动的业务逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6
  - Leverage: 现有 Service: `main/app/service/order_pay.go:OrderPaymentCoupon()`，参考优惠券选择逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 OrderPaymentActivity 方法，处理活动选择/取消/替换逻辑 | Context: 1. 加锁（参考优惠券实现），2. 验证活动有效性（有效期、时段、满减条件），3. 处理活动选择/取消/替换（类似优惠券逻辑），4. 计算活动抵扣金额，5. 处理与优惠券的互斥（活动与优惠券只能二选一），6. 处理与积分抵扣的互斥（选择活动后积分不再自动抵扣），7. 重新计算订单金额，8. 返回结账页面信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务，不使用 panic，返回 error | Success: 活动选择逻辑正确，互斥规则正确，金额计算正确

- [x] 2.6 实现活动抵扣金额计算

  - File: `main/app/service/order_pay.go`
  - Purpose: 实现活动抵扣金额的计算逻辑
  - Requirements: 3.1, 3.2
  - Leverage: 现有 Service: `main/app/service/order_pay.go`，参考优惠券抵扣金额计算
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 calculateActivityDiscount 方法，根据活动类型（阶梯满减/循环满减）计算抵扣金额 | Context: 1. 查询活动详情和规则，2. 根据活动类型计算抵扣金额（阶梯满减：找到满足条件的最大规则；循环满减：计算循环次数），3. 如果扣减金额大于订单金额，则最终扣减金额为订单金额，4. 返回抵扣金额和活动规则信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 decimal 进行金额计算 | Success: 金额计算逻辑正确，边界情况处理正确

- [x] 2.7 扩展结账完成逻辑 - 活动核销

  - File: `main/app/service/order_pay.go`
  - Purpose: 在结账完成时核销活动并记录活动信息
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: 现有 Service: `main/app/service/order_pay.go:InstantOrderPaymentFinish()`，参考优惠券核销逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 InstantOrderPaymentFinish 方法中增加活动核销逻辑 | Context: 1. 检查所选活动是否有效（如果无效则提示"活动信息已经变更，请重新确认"），2. 计算活动抵扣金额，3. 记录活动抵扣金额到订单表（activity_amount），4. 记录活动UUID到订单表（full_reduction_activity_uuid），5. 记录活动规则信息到订单表（full_reduction_activity_message），6. 活动抵扣金额不计入会员累计消费，7. 在订单操作日志中记录活动使用信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务，不使用 panic，返回 error | Success: 活动核销逻辑正确，活动信息记录正确

- [x] 2.8 扩展账单金额计算逻辑

  - File: `main/app/service/order_manage.go` 或相关文件
  - Purpose: 在计算账单金额时汇总活动抵扣金额
  - Requirements: 4.2
  - Leverage: 现有 Service: `main/app/service/order_manage.go:CalcAndSaveSaleBill()`，参考优惠券金额汇总逻辑
  - Prompt: Role: Go Developer | Task: 在计算账单金额时，汇总所有订单的活动抵扣金额到 sale_bill.activity_amount | Context: 在 CalcAndSaveSaleBill 方法中，汇总所有 sale_order 的 activity_amount 到 sale_bill.activity_amount | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 账单金额计算正确

- [x] 2.9 实现反结账逻辑 - 清空活动字段

  - File: `main/app/service/order_manage.go` 或相关文件
  - Purpose: 反结账时清空活动相关字段
  - Requirements: 4.8
  - Leverage: 现有 Service: 反结账相关方法
  - Prompt: Role: Go Developer | Task: 在反结账逻辑中清空活动相关字段 | Context: 清空 full_reduction_activity_uuid, full_reduction_activity_message, activity_amount 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 反结账逻辑正确

- [x] 2.10 实现拆单和免单场景处理

  - File: `main/app/service/order_pay.go` 或相关文件
  - Purpose: 处理拆单和免单场景下的活动逻辑
  - Requirements: 5.1, 5.2
  - Leverage: 现有 Service: 拆单和免单相关方法
  - Prompt: Role: Go Developer | Task: 实现拆单和免单场景下的活动处理逻辑 | Context: 拆单场景：每个拆单可以独立使用满减活动；免单场景：免单时清空活动选择，正常完成免单结账 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 拆单和免单场景处理正确

### API 层

- [x] 2.11 新增 API - 选择或取消满减活动

  - File: `main/app/api/v1/cashier/cashier_desk.go`
  - Purpose: 实现选择或取消满减活动的 API 接口
  - Requirements: 2.1, 2.2
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_desk.go:OrderPaymentCoupon()`，Task 2.5
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 OrderPaymentActivity API 方法，实现选择或取消满减活动接口 | Context: URL 为 /cashier/desk/order/payment/activity，使用 POST 方法，接收 InstantOrderPaymentActivityReq 参数，调用 OrderPaymentActivity Service 方法，返回 InstantOrderPaymentInfoResp | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 2.12 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册活动选择 API 路由
  - Requirements: 2.11
  - Leverage: 现有路由: `main/router/router.go`，参考优惠券路由注册
  - Prompt: Role: Go Developer | Task: 在 router.go 中注册 OrderPaymentActivity 路由 | Context: 路由路径为 /cashier/desk/order/payment/activity，使用 POST 方法，需要 JWT 认证 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功

- [x] 2.13 扩展 Assistant 端 API（如需要）

  - File: `main/app/api/v1/assistant/assistant_desk.go` 或相关文件
  - Purpose: 在助手端也支持活动选择功能
  - Requirements: 2.1, 2.2
  - Leverage: 现有 API: `main/app/api/v1/assistant/assistant_desk.go`，参考收银端实现
  - Prompt: Role: Go Developer | Task: 在助手端添加活动选择 API，复用收银端的 Service 方法 | Context: 创建 OrderPaymentActivity 方法，调用相同的 Service 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 助手端 API 创建成功

### 订单操作日志

- [x] 2.14 扩展订单操作日志 - 记录活动使用

  - File: `main/app/service/order_manage.go` 或相关文件
  - Purpose: 在订单操作日志中记录活动使用信息
  - Requirements: 4.6
  - Leverage: 现有 Service: 订单操作日志相关方法
  - Prompt: Role: Go Developer | Task: 在订单操作日志中记录活动使用信息 | Context: 记录活动名称和抵扣金额，参考优惠券使用记录的实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 订单操作日志记录正确

---

## Phase 3: 集成和优化

- [ ] 3.1 实现缓存策略

  - File: `main/app/service/order_pay.go`
  - Purpose: 使用 Redis 缓存活动列表
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/pkg/cache/`，参考优惠券缓存实现
  - Prompt: Role: Go Developer | Task: 实现活动列表缓存，Key 为 ttpos:full_reduction_activity:{company_uuid}:list，过期时间为 5 分钟 | Context: 使用 Cache-Aside Pattern，查询前先查缓存，缓存未命中时查询数据库并写入缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 缓存实现完成，命中率 > 80%

- [ ] 3.2 实现并发控制

  - File: `main/app/service/order_pay.go`
  - Purpose: 使用 UUID 锁防止并发冲突
  - Requirements: 可靠性要求
  - Leverage: `main/pkg/lock/system_lock.go`，参考优惠券并发控制
  - Prompt: Role: Go Developer | Task: 在活动选择时使用 UUID 锁，防止并发操作 | Context: 对 SaleBillUuid 加锁，参考优惠券选择的加锁逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 并发场景测试通过

- [ ] 3.3 数据库查询优化

  - File: `main/app/repository/` 或相关文件
  - Purpose: 优化活动列表查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，现有索引
  - Prompt: Role: Database Engineer | Task: 优化活动列表查询，添加必要的索引 | Context: 确保活动查询使用索引，避免全表扫描 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 查询时间 < 50ms

---

## Phase 4: 测试

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/order_pay_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 OrderPaymentActivity 和相关方法编写单元测试，覆盖率 ≥ 70% | Context: 测试活动列表查询，测试活动选择/取消，测试金额计算，测试互斥规则，测试活动核销 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `main/app/api/v1/cashier/cashier_desk_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/v1/cashier/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 OrderPaymentActivity API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 4.3 编写集成测试

  - File: `test/integration/full_reduction_activity_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程：进入结账页面 -> 选择活动 -> 计算金额 -> 完成结账，测试活动与优惠券互斥，测试活动与积分抵扣互斥，测试拆单场景，测试免单场景 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.4 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

---

## Phase 5: 文档更新

- [ ] 5.1 更新 API 文档

  - File: `docs/shared/api/cashier_api.md` 或相关文件
  - Purpose: 确保 API 文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新 API 文档，添加活动选择接口说明 | Context: 记录接口路径、参数、响应格式、错误码等 | Restrictions: 文档准确完整 | Success: API 文档已更新

- [x] 5.2 更新数据库文档

  - File: `docs/shared/specs/active/story-pos-full-reduction-marketing/design.md`
  - Purpose: 确保数据库文档完整
  - Requirements: 文档要求
  - Leverage: design.md
  - Success: 数据库文档已更新

- [ ] 5.3 更新 CHANGELOG

  - File: `CHANGELOG.md` 或相关文件
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Payment/Order: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] 数据库文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-pos-full-reduction-marketing/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-pos-full-reduction-marketing/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-pos-full-reduction-marketing/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-pos-full-reduction-marketing/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-pos-full-reduction-marketing/tasks.md)" | bc
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

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

