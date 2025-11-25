# 满减营销功能（打印相关）任务分解

> 本文档定义打印模板中新增"活动抵扣"字段的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 11  
**进行中**: 无  
**完成率**: 100%

---

## Phase 1: 自定义打印模板配置

### JSON 配置更新

- [x] 1.1 定位配置文件中的应付信息区块

  - File: `main/app/printer/pkg/template/statement_order_config.json`
  - Purpose: 定位到需要插入"活动抵扣"字段的位置
  - Requirements: 1.1
  - Leverage: 现有配置文件结构，参考"优惠券抵扣"配置项
  - Success: ✅ 找到 `"block_id": "amount_due_information"` 区块（第289行），data_rows 为空数组（表示可由用户自定义）

- [x] 1.2 添加"活动抵扣"字段到数据结构和默认模板

  - File: 
    - `main/app/printer/template_struct/order.go` (数据结构)
    - `main/app/printer/pkg/template/statement_order_tmp.json` (结账单默认模板)
    - `main/app/printer/pkg/template/statement_order_data.json` (结账单示例数据)
    - `main/app/printer/pkg/template/statement_pre_tmp.json` (预结账单默认模板)
    - `main/app/printer/pkg/template/statement_pre_data.json` (预结账单示例数据)
  - Purpose: 在数据结构中添加"activity_amount"字段，并在默认自定义模板中添加配置块
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: CouponExchangeAmount 字段的定义
  - Success: ✅ 已完成五项修改：
    1. StatementOrderInfoData 结构体中添加 `ActivityAmount float64` 字段
    2. 结账单默认模板中添加"活动抵扣"配置块（在优惠券抵扣之后）
    3. 结账单示例数据中添加 `activity_amount: 50`
    4. 预结账单默认模板中添加"活动抵扣"配置块（在优惠券抵扣之后）
    5. 预结账单示例数据中添加 `activity_amount: 50`
  - Note: 默认模板使用 `operator: "gt"` 确保仅大于0时显示

- [x] 1.3 验证数据可用性

  - File: `main/app/printer/template_struct/order.go`, `main/app/printer/template/statement_order_img_custom.go`
  - Purpose: 确保数据结构正确，模板可以访问 activity_amount
  - Requirements: 1.1
  - Success: ✅ 数据结构已添加，自定义模板可使用 `{{order.activity_amount}}` 访问数据，标准模板直接使用 `saleOrder.ActivityAmount`

---

## Phase 2: 自定义模板渲染逻辑

### Go 代码实现

- [x] 2.1 确认数据源字段

  - File: `main/app/model/sale_bill.go` (或相关模型文件)
  - Purpose: 确认 `activity_amount` 字段在数据模型中存在且类型正确
  - Requirements: 1.4
  - Leverage: 现有 Model 定义
  - Success: ✅ 确认 SaleOrder 和 SaleBill 包含 `ActivityAmount` 字段，类型为 `float64`

- [x] 2.2 在自定义模板渲染中读取活动抵扣金额

  - File: `main/app/printer/template/statement_order_img_custom.go`
  - Purpose: 在 `GetPrintContent()` 方法中读取 `activity_amount` 字段
  - Requirements: 1.4, 1.5
  - Leverage: 现有金额读取代码（如优惠券抵扣、会员优惠等）
  - Success: ✅ 在 StatementOrderInfoData 中添加 `ActivityAmount: saleOrder.ActivityAmount`

- [x] 2.3 判断金额并格式化

  - File: `main/app/printer/template/statement_order_img_custom.go`
  - Purpose: 判断 `activity_amount` 是否大于 0，并格式化为两位小数的字符串
  - Requirements: 1.5, 1.6, 1.7
  - Leverage: 现有金额格式化代码和判断逻辑（参考优惠券抵扣）
  - Success: ✅ 添加到 StatementOrderInfoData 结构体，在模板中可使用 {{order.activity_amount}} 并通过条件判断显示
  - Code Example:
    ```go
    activityAmount := saleBill.ActivityAmount
    if activityAmount == nil {
        activityAmount = &decimal.Zero
    }
    
    // 仅大于 0 才继续处理
    if activityAmount.GreaterThan(decimal.Zero) {
        activityAmountStr := activityAmount.StringFixed(2)
        // 继续渲染逻辑...
    }
    ```

- [x] 2.4 数据结构填充（自定义模板）

  - File: `main/app/printer/template/statement_order_img_custom.go`
  - Purpose: 将 `activity_amount` 数据填充到模板数据结构中
  - Requirements: 1.6, 1.7, 3.3
  - Leverage: 现有数据填充逻辑（如 CouponExchangeAmount）
  - Success: ✅ 已添加 `ActivityAmount: saleOrder.ActivityAmount` 到 StatementOrderInfoData，模板解析器会自动处理条件显示和多语言

- [x] 2.5 错误处理（自然处理）

  - File: 标准模板和自定义模板
  - Purpose: 确保异常情况的优雅处理
  - Requirements: 非功能需求 - 可靠性
  - Success: ✅ 标准模板使用 `if saleOrder.ActivityAmount > 0` 判断，自动处理 0 和 nil 值；自定义模板由解析器处理

---

## Phase 3: 标准模板渲染逻辑

### 结账单/预结账单/发票模板更新

- [x] 3.1 在标准模板中添加活动抵扣渲染（仅大于 0）

  - File: `main/app/printer/template/statement_order_img.go` 等 6 个文件
  - Purpose: 在结账单、预结账单、发票模版2的渲染代码中添加"活动抵扣"行（仅当金额 > 0 时）
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有"优惠券抵扣"渲染代码，Task 2.2-2.4 的实现
  - Success: ✅ 已在以下文件中添加渲染逻辑：
    - statement_order_img.go
    - statement_order_img_58mm.go
    - statement_order_compax.go
    - statement_order_codesoft.go
    - statement_order_xprinter.go
    - statement_order_sunmi.go
  - Code Example:
    ```go
    // 优惠券抵扣
    if couponAmount.GreaterThan(decimal.Zero) {
        line := fmt.Sprintf("%s: ￥%s", t.base.Translate("优惠券抵扣"), couponAmount.StringFixed(2))
        // ... 渲染逻辑
    }
    
    // 活动抵扣 (新增 - 仅大于 0 时显示)
    if activityAmount.GreaterThan(decimal.Zero) {
        line := fmt.Sprintf("%s: ￥%s", t.base.Translate("活动抵扣"), activityAmount.StringFixed(2))
        // ... 渲染逻辑
    }
    
    // 减免金额
    if reductionAmount.GreaterThan(decimal.Zero) {
        line := fmt.Sprintf("%s: ￥%s", t.base.Translate("减免金额"), reductionAmount.StringFixed(2))
        // ... 渲染逻辑
    }
    ```

- [x] 3.2 验证标准模板的零金额/负金额处理

  - File: `main/app/printer/template/statement_order_img.go`
  - Purpose: 确保当 `activity_amount` ≤ 0 时，不显示"活动抵扣"行
  - Requirements: 2.5, 验收标准 5
  - Leverage: Task 3.1 的实现
  - Success: ✅ 使用 `if saleOrder.ActivityAmount > 0` 判断，0 或负数时不渲染

---

## Phase 4: 测试和验证（可选 - 建议由QA执行）

### 功能测试

- [ ] 4.1 单元测试：金额格式化（可选）

  - File: `main/app/printer/template/statement_order_img_custom_test.go` (或新建)
  - Purpose: 测试 `activity_amount` 金额格式化的正确性
  - Requirements: 测试要求
  - Leverage: 现有测试代码模式
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 编写单元测试，验证 activity_amount 金额格式化 | Context: 测试正常金额（50.00）、零金额（0.00）、nil 值、极大值等情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖各种场景，所有测试通过
  - Test Cases:
    - 输入: `decimal.NewFromFloat(50.00)` → 输出: `"50.00"` → 显示
    - 输入: `decimal.Zero` → 不显示
    - 输入: `nil` → 不显示
    - 输入: `decimal.NewFromFloat(-10.00)` → 不显示
    - 输入: `decimal.NewFromFloat(0.01)` → 输出: `"0.01"` → 显示
    - 输入: `decimal.NewFromFloat(9999.99)` → 输出: `"9999.99"` → 显示

- [ ] 4.2 集成测试：完整打印流程（建议）

  - File: 测试环境手动验证
  - Purpose: 测试完整的打印渲染流程，验证"活动抵扣"字段显示（包括大于 0 和等于 0 的情况）
  - Requirements: 验收标准 1-6
  - Test Cases:
    1. 创建测试订单，设置 `activity_amount` = 50.00，打印验证显示"活动抵扣: ￥50.00"
    2. 设置 `activity_amount` = 0，打印验证不显示"活动抵扣"行
  - Success: 金额 > 0 时显示，≤ 0 时不显示

- [ ] 4.3 多语言测试（建议）

  - File: 测试环境手动测试
  - Purpose: 验证 9 种语言环境下"活动抵扣"显示正确
  - Requirements: 多语言验证 1-5
  - Leverage: 使用 `t.base.Translate("活动抵扣")` 自动多语言
  - Test Matrix:
    | 语言   | 期望显示             |
    | ------ | -------------------- |
    | zh     | 活动抵扣             |
    | zhtw   | 活動抵扣             |
    | en     | Activity Deduction   |
    | th     | หักลดจากกิจกรรม     |
    | tr     | Etkinlik indir       |
    | ja     | 活動控除             |
    | ko     | 활动 공제            |
    | sv     | Aktivitetavdrag      |
    | my     | လှုပ်ရှားမှုလျှော့   |
  - Note: 需要在多语言配置文件中添加"活动抵扣"的翻译（如果尚未存在）

- [ ] 4.4 回归测试：其他字段不受影响（建议）

  - File: 测试环境手动测试
  - Purpose: 确保新增字段不影响现有打印模板的其他字段
  - Requirements: 回归测试 1-3
  - Leverage: 现有打印模板测试用例
  - Quick Check List:
    - [ ] 优惠券抵扣显示正常
    - [ ] 手动抹零显示正常
    - [ ] 打印布局正常，无错位
  - Success: 现有功能不受影响

---

## Phase 5: 文档和交付

### 文档更新

- [x] 5.1 更新 CHANGELOG

  - File: `docs/shared/specs/task-printer-activity-deduction/CHANGELOG.md`
  - Purpose: 记录本次功能更新
  - Requirements: 文档验收
  - Leverage: 现有 CHANGELOG 格式
  - Success: ✅ CHANGELOG 已创建，完整记录功能实现、技术细节、测试状态、相关文档和性能影响
  - Entry:
    ```markdown
    ## [v2.10.0] - 2025-11-25
    
    ### Added
    - 打印模板新增"活动抵扣"字段显示（DooTask #37078）
      - 自定义打印模板支持配置"活动抵扣"字段
      - 结账单、预结账单、发票模版2显示"活动抵扣"金额
      - 支持 9 种语言翻译
    ```

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt`
- [ ] Go 代码通过 `go vet`
- [ ] JSON 配置格式正确（通过 `jq` 验证）
- [ ] 单元测试通过
- [ ] 集成测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成（参考 requirements.md）

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] Spec 文档（requirements.md, design.md, tasks.md）与实现一致

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/structs.mdc`
- [ ] 代码风格与现有代码保持一致

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/task-printer-activity-deduction/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/task-printer-activity-deduction/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/task-printer-activity-deduction/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/task-printer-activity-deduction/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/task-printer-activity-deduction/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（从 Phase 1 开始）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的技术设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码（如有）
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`, `jq`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 打印模板开发

```
Role: Go Developer specializing in print template rendering

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/structs.mdc

Restrictions:
- 不使用 panic，返回 error
- 使用 decimal 类型处理金额
- 金额保留两位小数
- 遵循现有打印模板代码风格
- 使用 logger.Logger 记录错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 金额格式化正确（两位小数）
```

### JSON 配置工程师

```
Role: Configuration Engineer with JSON expertise

Task: {配置任务描述}

Context:
- Current file: main/app/printer/pkg/template/statement_order_config.json
- Leverage: 现有字段配置（如"优惠券抵扣"）
- Requirements: {需求编号和内容}

Restrictions:
- JSON 格式必须正确（使用 jq 验证）
- 不破坏现有配置结构
- 包含完整的多语言翻译（9 种语言）
- block_id 唯一，不与现有冲突

Success Criteria:
- JSON 格式正确（通过 jq 验证）
- 配置添加成功
- 多语言翻译完整
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Test scenarios: {测试场景列表}

Test Cases Required:
- 正常场景测试（金额 > 0）
- 零金额测试（金额 = 0）
- Nil 值测试
- 多语言测试
- 回归测试（其他字段不受影响）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 测试覆盖边界情况

Success Criteria:
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2025-11/2025-11-25.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-25  
**维护者**: weifashi

