# 订单分批送厨模式快照 任务分解

> 本文档定义订单分批送厨模式快照功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 11  
**进行中**: -  
**完成率**: 73%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_batch_cooking_mode_to_sale_bill_setting.php`
  - Purpose: 为 `ttpos_sale_bill_setting` 表新增 `batch_cooking_mode` 字段
  - Requirements: 1.1, 1.3
  - Leverage: 现有迁移文件: `admin/database/migrations/20251126201148_add_source_to_sale_bill.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，为 ttpos_sale_bill_setting 表新增 batch_cooking_mode 字段 | Context: 字段类型 VARCHAR(10)，默认值 'post'，NOT NULL，注释为"分批送厨模式: pre-前置 / post-后置，默认 post"，放在 auto_points_exchange 字段之后 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中新增字段
  - Requirements: 1.1, 1.3
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已新增

- [x] 1.3 为历史数据设置默认值

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_update_batch_cooking_mode_default.php` 或直接在 Task 1.1 中包含
  - Purpose: 为历史订单的 `batch_cooking_mode` 字段设置默认值 "post"
  - Requirements: 1.4
  - Leverage: Task 1.1 的迁移文件
  - SQL: `UPDATE ttpos_sale_bill_setting SET batch_cooking_mode = 'post' WHERE batch_cooking_mode IS NULL OR batch_cooking_mode = ''`
  - Success: 历史数据已设置默认值

- [x] 1.4 更新 Go Model

  - File: `main/app/model/order.go`
  - Purpose: 在 `SaleBillSetting` 结构体中添加 `BatchCookingMode` 字段
  - Requirements: 1.2
  - Leverage: 现有 Model: `main/app/model/order.go` (SaleBillSetting 结构体)
  - Prompt: Role: Go Developer | Task: 在 SaleBillSetting 结构体中添加 BatchCookingMode 字段 | Context: 字段类型 string，gorm 标签 column:batch_cooking_mode，type:varchar(10)，default:'post'，comment 为"分批送厨模式: pre-前置 / post-后置，默认 post"，json 标签为 batch_cooking_mode | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段放在 AutoPointsExchange 字段之后 | Success: Model 更新成功，字段映射正确

---

## Phase 2: 订单创建逻辑修改

- [x] 2.1 修改 NewSaleBillSetting 函数

  - File: `main/app/service/order.go`
  - Purpose: 在创建订单时从 `business_setting` 读取分批送厨模式并保存到 `sale_bill_setting`
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有函数: `main/app/service/order.go` (NewSaleBillSetting，第 408-544 行)，常量: `main/app/constant/setting.go` (BatchCookingModePre, BatchCookingModePost)
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 NewSaleBillSetting 函数，在创建 SaleBillSetting 时保存 batch_cooking_mode 字段 | Context: 从 businessSetting.BatchCookingMode 读取值，如果为空则使用默认值 constant.BatchCookingModePost，保存到 saleBillSetting.BatchCookingMode | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 函数修改成功，订单创建时正确保存 batch_cooking_mode

- [ ] 2.2 测试订单创建逻辑

  - File: `main/app/service/order_test.go` (如存在)
  - Purpose: 确保订单创建时正确保存 `batch_cooking_mode`
  - Requirements: 2.3
  - Leverage: 现有测试文件
  - Test Cases:
    - 测试前置模式：business_setting.BatchCookingMode = "pre"，验证保存为 "pre"
    - 测试后置模式：business_setting.BatchCookingMode = "post"，验证保存为 "post"
    - 测试默认值：business_setting.BatchCookingMode = ""，验证保存为 "post"
  - Success: 所有测试用例通过

---

## Phase 3: 送厨逻辑修改

- [x] 3.1 修改 ActionCooking 函数

  - File: `main/app/service/order_action.go`
  - Purpose: 送厨时从 `sale_bill_setting.batch_cooking_mode` 读取，而不是从 `business_setting` 读取
  - Requirements: 3.1, 3.2
  - Leverage: 现有函数: `main/app/service/order_action.go` (ActionCooking，第 120-154 行)，Repository: `main/app/repository/order.go` (GetSaleBillAllInfo 已预加载 SaleBillSetting)
  - Prompt: Role: Go Developer with business logic expertise | Task: 修改 ActionCooking 函数，从 saleBill.SaleBillSetting.BatchCookingMode 读取分批送厨模式 | Context: 移除从 business_setting 读取的逻辑，改为从 saleBill.SaleBillSetting.BatchCookingMode 读取，如果为空则使用默认值 constant.BatchCookingModePost | Restrictions: 遵循 .cursor/rules/go-main.mdc，注意兼容历史数据（字段为空时使用默认值） | Success: 函数修改成功，送厨逻辑使用快照值

- [x] 3.2 修改 InstantOrderCartProductCooking 函数

  - File: `main/app/service/order_product.go`
  - Purpose: 助手端送厨时从 `sale_bill_setting.batch_cooking_mode` 读取
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有函数: `main/app/service/order_product.go` (InstantOrderCartProductCooking，第 501-551 行)
  - Prompt: Role: Go Developer | Task: 修改 InstantOrderCartProductCooking 函数的 defer 逻辑，从 sale_bill_setting 读取 batch_cooking_mode | Context: 在 defer 函数中，先获取 saleBill，然后从 saleBill.SaleBillSetting.BatchCookingMode 读取，如果为空则使用默认值 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数修改成功，助手端送厨使用快照值

- [x] 3.3 修改 updateProductBatchFlagToZero 函数调用

  - File: `main/app/service/order_product.go`
  - Purpose: 修改 `updateProductBatchFlagToZero` 函数调用，传入 `sale_bill_setting.batch_cooking_mode`
  - Requirements: 3.1, 3.2
  - Leverage: 现有函数: `main/app/service/order_product.go` (updateProductBatchFlagToZero，第 588 行)
  - Prompt: Role: Go Developer | Task: 修改 updateProductBatchFlagToZero 函数调用，传入 sale_bill_setting.batch_cooking_mode 而不是 business_setting.BatchCookingMode | Context: 需要先获取 saleBill，然后从 saleBill.SaleBillSetting.BatchCookingMode 读取 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数调用修改成功

- [x] 3.4 修改订单商品创建逻辑

  - File: `main/app/service/order.go`
  - Purpose: 订单商品创建时使用 `sale_bill_setting.batch_cooking_mode`（如果已创建）或 `business_setting.BatchCookingMode`（创建时）
  - Requirements: 3.1, 3.2
  - Leverage: 现有代码: `main/app/service/order.go` (第 1774 行)
  - Prompt: Role: Go Developer | Task: 修改订单商品创建逻辑，优先使用 sale_bill_setting.batch_cooking_mode，如果不存在则使用 business_setting.BatchCookingMode | Context: 在创建订单商品时，如果 sale_bill_setting 已创建，使用其 batch_cooking_mode，否则使用 business_setting.BatchCookingMode（仅用于创建时） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 逻辑修改成功

- [x] 3.5 修改 updateProductBatchFlagToZero 函数实现

  - File: `main/app/service/order_action.go`
  - Purpose: `updateProductBatchFlagToZero` 函数内部使用传入的 `batch_cooking_mode` 参数
  - Requirements: 3.1, 3.2
  - Leverage: 现有函数: `main/app/service/order_action.go` (updateProductBatchFlagToZero，第 911-924 行)
  - Prompt: Role: Go Developer | Task: 确保 updateProductBatchFlagToZero 函数使用传入的 batch_cooking_mode 参数，而不是从 business_setting 读取 | Context: 函数签名已包含 modeType 参数，确保函数内部使用该参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数实现正确

- [x] 3.6 修改响应 DTO 数据来源

  - File: `main/app/service/order.go` 或相关 Service 文件
  - Purpose: `shop_cart.BatchCookingMode` 字段的值来源从 `business_setting` 改为 `sale_bill_setting.batch_cooking_mode`
  - Requirements: 3.1, 3.2
  - Leverage: 现有 DTO: `main/app/dto/resp/shop_cart.go` (BatchCookingMode 字段，第 167 行)
  - Prompt: Role: Go Developer | Task: 修改 shop_cart.BatchCookingMode 字段的赋值逻辑，从 sale_bill_setting.batch_cooking_mode 读取 | Context: 查找所有设置 shop_cart.BatchCookingMode 的地方，改为从 saleBill.SaleBillSetting.BatchCookingMode 读取 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 响应数据来源修改成功

---

## Phase 4: 测试和优化

- [ ] 4.1 编写 NewSaleBillSetting 单元测试

  - File: `main/app/service/order_test.go` (如存在) 或新建
  - Purpose: 测试订单创建时正确保存 `batch_cooking_mode`
  - Requirements: 2.3
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 NewSaleBillSetting 编写单元测试，覆盖率 100% | Context: 测试前置模式、后置模式、默认值场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

- [ ] 4.2 编写送厨逻辑单元测试

  - File: `main/app/service/order_action_test.go`, `main/app/service/order_product_test.go`
  - Purpose: 测试送厨逻辑使用快照值
  - Requirements: 3.1, 3.2
  - Leverage: 现有测试文件
  - Test Cases:
    - 测试前置模式订单送厨
    - 测试后置模式订单送厨
    - 测试历史订单兼容性（字段为空时使用默认值）
    - 测试配置变更不影响已创建订单
  - Success: 测试覆盖率达标，所有测试通过

- [ ] 4.3 集成测试

  - File: `test/integration/order_batch_cooking_mode_test.go` (如存在)
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Test Flow:
    1. 创建订单（前置模式）→ 验证 `batch_cooking_mode` = "pre"
    2. 修改全局配置为后置模式
    3. 送厨已创建订单 → 验证仍使用前置模式
    4. 创建新订单 → 验证使用后置模式
  - Success: 集成测试通过

- [ ] 4.4 回归测试

  - File: -
  - Purpose: 确保现有功能不受影响
  - Requirements: 可靠性要求
  - Test Cases:
    - 订单创建流程正常
    - 送厨流程正常
    - 分批送厨功能正常（前置/后置模式）
    - POS、Assistant、KDS 等终端功能正常
  - Success: 所有回归测试通过

- [ ] 4.5 性能测试

  - File: -
  - Purpose: 确保性能不受影响
  - Requirements: 性能要求
  - Test Metrics:
    - 订单创建响应时间 < 200ms
    - 送厨响应时间不受影响
  - Success: 性能指标达标

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Order 相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] API 文档已更新（如有变更）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-order-batch-cooking-mode-snapshot/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-order-batch-cooking-mode-snapshot/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-order-batch-cooking-mode-snapshot/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-order-batch-cooking-mode-snapshot/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-order-batch-cooking-mode-snapshot/tasks.md)" | bc
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
**最后更新**: 2025-12-02  
**维护者**: xiezhihuan

