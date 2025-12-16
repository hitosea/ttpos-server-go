# 点餐端快速选择备注 任务分解

> 本文档定义点餐端快速选择备注功能的详细执行任务清单。

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

## Phase 0: 数据库扩展（如需要）

- [ ] 0.1 创建数据库迁移文件，添加 `order_item_remark_uuid` 字段

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_order_item_remark_uuid_to_sale_order_product_reason.php`
  - Purpose: 在 `ttpos_sale_order_product_reason` 表中新增 `order_item_remark_uuid` 字段
  - Requirements: 数据库设计要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考 `20251208174755_add_reason_name_to_sale_order_product_reason.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，添加 order_item_remark_uuid 字段和索引 | Context: 字段类型为 BIGINT UNSIGNED，默认值为 0，添加索引 idx_order_item_remark_uuid | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 迁移文件创建成功，字段和索引定义正确

- [ ] 0.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 数据库设计要求
  - Leverage: Task 0.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [ ] 0.3 更新 Go Model，添加 `OrderItemRemarkUuid` 字段

  - File: `main/app/model/order.go`
  - Purpose: 在 `SaleOrderProductReason` 结构体中添加 `OrderItemRemarkUuid` 字段
  - Requirements: 数据模型设计要求
  - Leverage: 现有 Model: `main/app/model/order.go`（参考 `ReturnFoodReasonUuid` 字段）
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProductReason 结构体中添加 OrderItemRemarkUuid 字段 | Context: 字段类型为 uint64，gorm 标签为 `column:order_item_remark_uuid`，添加 IsOrderItemRemark() 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段和方法添加正确

---

## Phase 1: DTO 和 Service 层开发

### DTO 层

- [ ] 1.1 创建 Request DTO

  - File: `main/app/dto/req/cashier.go` 或新建 `main/app/dto/req/order_item_remark.go`
  - Purpose: 定义获取菜品备注预设的请求参数
  - Requirements: 1.1, 1.2
  - Leverage: 现有 DTO: `main/app/dto/req/shop.go`（参考 `AddOrderItemRemarkReq`）
  - Prompt: Role: Go Developer | Task: 创建 GetOrderItemRemarkPresetReq 结构体，包含 product_uuid 字段 | Context: 使用 binding 标签验证参数，product_uuid 为必填 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [ ] 1.2 扩展 Response DTO

  - File: `main/app/dto/resp/order_item_remark.go`
  - Purpose: 扩展响应结构体，支持根据菜品获取备注预设
  - Requirements: 1.1, 1.2
  - Leverage: 现有 DTO: `main/app/dto/resp/order_item_remark.go`（已存在 `OrderItemRemarkResp`）
  - Prompt: Role: Go Developer | Task: 在 OrderItemRemarkResp 基础上，确保响应格式正确 | Context: data 必须是对象，不能是 null 或数组 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: DTO 扩展成功，响应格式正确

### Service 层

- [ ] 1.3 在 Service 接口中新增方法

  - File: `main/app/service/i_other_service.go`
  - Purpose: 定义根据菜品获取备注预设的业务方法
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 Service 接口: `main/app/service/i_other_service.go`（参考 `GetReturnFoodReasonList`）
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IOtherSrv 接口中新增 GetOrderItemRemarkPresetByProductUuid 方法 | Context: 方法签名参考 GetReturnFoodReasonList，参数为 product_uuid | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [ ] 1.4 实现 Service 业务逻辑

  - File: `main/app/service/other.go`
  - Purpose: 实现根据菜品获取备注预设的业务逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 Service 实现: `main/app/service/other.go`（参考 `GetReturnFoodReasonList` 和 `GetOrderItemRemarkList`）
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetOrderItemRemarkPresetByProductUuid 方法，参考退菜逻辑实现，使用 ttpos_sale_order_product_reason 表 | Context: 1. 根据 product_uuid 查询 ttpos_sale_order_product_reason 表中 order_item_remark_uuid > 0 的记录 2. 通过 sale_order_product_uuid 关联到 sale_order_product 表，再通过 product_uuid 关联到 product 表 3. 筛选出该商品在历史订单中使用过的备注预设 UUID 列表 4. 如果菜品有关联备注预设，返回菜品专属预设列表 5. 如果菜品无关联备注预设，返回全局备注预设列表（查询 ttpos_order_item_remark 表） 6. 只返回启用状态（delete_time = 0）的备注预设 7. 按排序顺序返回 8. 实现 Redis 缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，缓存逻辑正确

- [ ] 1.5 编写 Service 单元测试

  - File: `main/app/service/other_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/other_test.go`（参考 `TestOtherSrv_GetOrderItemRemarkList`）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetOrderItemRemarkPresetByProductUuid 编写单元测试，覆盖率 ≥ 70% | Context: 测试菜品有关联预设、无关联预设、全局预设、缓存等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 2: API 层开发

### API 层

- [ ] 2.1 创建收银端 API

  - File: `main/app/api/v1/cashier/cashier_base.go`
  - Purpose: 实现收银端获取菜品备注预设的 HTTP API 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_base.go`（参考 `GetReturnReason`）
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 GetOrderItemRemarkPreset API，参考退菜原因 API 实现 | Context: URL 使用 snake_case `/api/v1/cashier/item_remark_preset`，使用 GET 方法，接收 product_uuid 查询参数，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，data 必须是对象 | Success: API 创建成功，响应格式正确，参数验证正确

- [ ] 2.2 创建助手端 API

  - File: `main/app/api/v1/assistant/assistant_base.go`
  - Purpose: 实现助手端获取菜品备注预设的 HTTP API 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 API: `main/app/api/v1/assistant/assistant_base.go`（参考 `GetReturnReason`）
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 GetOrderItemRemarkPreset API，参考收银端实现 | Context: URL 使用 snake_case `/api/v1/assistant/item_remark_preset`，使用 GET 方法，接收 product_uuid 查询参数 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 创建成功，响应格式正确

- [ ] 2.3 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册收银端和助手端的 API 路由
  - Requirements: 1.1, 1.2
  - Leverage: 现有路由: `main/router/router.go`（参考退菜原因路由注册）
  - Success: 路由注册成功

- [ ] 2.4 编写 API 集成测试

  - File: `main/app/api/v1/cashier/cashier_base_test.go` 或新建测试文件
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/cashier/` 或 `main/app/api/v1/assistant/`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 GetOrderItemRemarkPreset API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 缓存和性能优化

- [ ] 3.1 实现 Redis 缓存逻辑

  - File: `main/app/service/other.go`
  - Purpose: 实现备注预设列表的 Redis 缓存
  - Requirements: 1.5, 性能要求
  - Leverage: 现有缓存实现: `main/app/service/`（参考其他 Service 的缓存实现）
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 GetOrderItemRemarkPresetByProductUuid 方法中实现 Redis 缓存 | Context: Key 命名: `ttpos:order_item_remark:product:{product_uuid}`，过期时间: 5 分钟，使用 Cache-Aside Pattern | Restrictions: 遵循缓存最佳实践 | Success: 缓存实现完成，命中率 > 80%

- [ ] 3.2 实现缓存失效策略

  - File: `main/app/service/other.go`（备注预设更新时）
  - Purpose: 备注预设更新时清除相关缓存
  - Requirements: 1.5, 性能要求
  - Leverage: 现有缓存失效实现
  - Success: 缓存失效策略实现完成

- [ ] 3.3 数据库查询优化

  - File: `main/app/repository/base/order_item_remark.go`（如需要新增方法）
  - Purpose: 优化 SQL 查询，添加索引
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

---

## Phase 4: 测试和文档

- [ ] 4.1 集成测试

  - File: `test/integration/order_item_remark_preset_test.go` 或现有测试文件
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程，测试数据一致性，测试多终端（POS/Assistant） | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 4.3 文档更新

  - File: `docs/shared/api/order_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
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
grep -c "^- \[" docs/shared/specs/active/story-order-item-remark-preset/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-order-item-remark-preset/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-order-item-remark-preset/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-order-item-remark-preset/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-order-item-remark-preset/tasks.md)" | bc
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
**最后更新**: 2025-12-09  
**维护者**: 后端开发组

