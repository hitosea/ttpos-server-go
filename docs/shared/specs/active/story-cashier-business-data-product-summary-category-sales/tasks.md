# （优化）收银机-营业数据按商品【合计】、按商品分类【销售额】取值调整 任务分解

> 本文档定义（优化）收银机-营业数据按商品【合计】、按商品分类【销售额】取值调整功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 6  
**进行中**: -  
**完成率**: 75%

---

## Phase 1: Repository 层调整

### Requirement 1: 调整按商品统计【合计】字段取值

- [x] 1.1 调整 CountProduct 方法的 sale_amount 计算逻辑

  - File: `main/app/repository/statistics.go`
  - Purpose: 将按商品统计的合计字段从原价销售额改为实际销售额（不包含退款）
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有计算逻辑: `CountProductSale` 方法中的 `actual_sale_amount` 计算（921行）
  - Prompt: Role: Go Developer specializing in SQL and Statistics | Task: 调整 CountProduct 方法中的 sale_amount 计算逻辑，从 `SUM(sp.product_final_price * sp.product_num)` 改为 `SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num)))` | Context: 参考 CountProductSale 方法中的 actual_sale_amount 计算逻辑，排除免单和赠菜，扣减退款数量 | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保 SQL 语法正确 | Success: 计算逻辑调整成功，排除免单和赠菜，扣减退款数量

- [x] 1.2 验证 CountProduct 方法的退款扣减逻辑

  - File: `main/app/repository/statistics.go`
  - Purpose: 验证退款扣减逻辑正确，包括单品退款和整单退款场景
  - Requirements: 1.2, 1.3, 1.4
  - Leverage: Task 1.1 的实现
  - Prompt: Role: QA Engineer | Task: 验证 CountProduct 方法的退款扣减逻辑 | Context: 测试单品退款场景（refund_num < product_num），测试整单退款场景（refund_num = product_num），验证计算结果与商家后台一致 | Restrictions: 确保退款数量不会超过商品数量 | Success: 退款扣减逻辑正确，与商家后台数据一致

---

### Requirement 2: 调整按商品分类统计【销售额】字段取值

- [x] 2.1 调整 CountCategory 方法的 sale_amount 计算逻辑

  - File: `main/app/repository/statistics.go`
  - Purpose: 将按商品分类统计的销售额字段从原价销售额改为实际销售额（不包含退款）
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: Task 1.1 的实现，现有计算逻辑: `CountProductSale` 方法中的 `actual_sale_amount` 计算
  - Prompt: Role: Go Developer specializing in SQL and Statistics | Task: 调整 CountCategory 方法中的 sale_amount 计算逻辑，从 `SUM(sp.product_final_price * sp.product_num)` 改为 `SUM(IF(sp.free_num > 0 OR sp.give_num > 0, 0, sp.product_final_price * (sp.product_num - sp.refund_num)))` | Context: 与 CountProduct 方法使用相同的计算逻辑，确保数据一致性，需要调整两个分支（categoryType != 2 和 categoryType == 2） | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保 SQL 语法正确 | Success: 计算逻辑调整成功，与 CountProduct 方法保持一致

- [x] 2.2 验证 CountCategory 方法的退款扣减逻辑

  - File: `main/app/repository/statistics.go`
  - Purpose: 验证退款扣减逻辑正确，包括单品退款和整单退款场景
  - Requirements: 2.2, 2.3, 2.4
  - Leverage: Task 2.1 的实现
  - Prompt: Role: QA Engineer | Task: 验证 CountCategory 方法的退款扣减逻辑 | Context: 测试单品退款场景，测试整单退款场景，验证计算结果与商家后台一致，验证与 CountProduct 方法的数据一致性 | Restrictions: 确保退款数量不会超过商品数量 | Success: 退款扣减逻辑正确，与商家后台数据一致，与 CountProduct 方法数据一致

---

## Phase 2: 测试验证

- [x] 3.1 编写 Repository 层单元测试

  - File: `main/app/repository/statistics_test.go`
  - Purpose: 确保统计计算逻辑正确，测试覆盖率 ≥ 80%
  - Requirements: 1.5, 2.5, 测试要求
  - Leverage: 现有测试: `main/app/repository/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountProduct 和 CountCategory 方法编写单元测试，覆盖率 ≥ 80% | Context: 测试正常商品无退款场景，测试正常商品有退款场景，测试免单商品场景，测试赠菜商品场景，测试退款数量边界情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 3.2 API 集成测试

  - File: `main/app/api/v1/cashier/cashier_statistics_test.go`（如存在）或手动测试
  - Purpose: 测试统计接口的数据准确性
  - Requirements: 功能验收 1, 2
  - Leverage: 现有 API 测试
  - Prompt: Role: QA Automation Engineer | Task: 测试统计接口的数据准确性 | Context: 测试 `/cashier/statistics/product` 接口，验证 subtotal 字段为实际销售额，测试 `/cashier/statistics/product_category` 接口，验证 prices 字段为实际销售额 | Restrictions: 确保数据与商家后台一致 | Success: 所有 API 测试通过，数据准确

- [ ] 3.3 数据对比测试（与商家后台）

  - File: -
  - Purpose: 验证统计数据与商家后台数据一致性
  - Requirements: 1.4, 1.5, 2.4, 2.5
  - Leverage: 商家后台统计接口
  - Prompt: Role: QA Engineer | Task: 对比收银机统计数据和商家后台统计数据 | Context: 创建测试订单并支付，执行单品退款，对比收银机和商家后台的统计数据，执行整单退款，对比收银机和商家后台的统计数据 | Restrictions: 确保数据完全一致 | Success: 统计数据与商家后台数据一致

---

## Phase 3: 打印功能验证

- [x] 4.1 检查打印模板数据来源

  - File: `main/app/printer/template/business_data_*.go`
  - Purpose: 确保打印模板使用调整后的统计接口数据
  - Requirements: 3.1
  - Leverage: 现有打印模板: `main/app/printer/template/business_data_*.go`
  - Prompt: Role: Go Developer | Task: 检查打印模板的数据来源 | Context: 检查 business_data_xprinter.go、business_data_sunmi.go 等打印模板，确认数据来源是调整后的统计接口 | Restrictions: 确保打印数据使用调整后的接口 | Success: 打印模板数据来源正确

- [ ] 4.2 验证打印数据一致性

  - File: `main/app/printer/template/business_data_*.go`
  - Purpose: 验证打印小票显示的合计和销售额与调整后的接口数据一致
  - Requirements: 3.2, 3.3
  - Leverage: Task 4.1 的检查结果
  - Prompt: Role: QA Engineer | Task: 验证打印数据与接口数据一致 | Context: 打印按商品统计的营业数据，验证显示的合计与接口数据一致，打印按商品分类统计的营业数据，验证显示的销售额与接口数据一致 | Restrictions: 确保打印数据正确反映退款扣减后的实际销售额 | Success: 打印数据与接口数据一致

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Repository: ≥ 80%
  - 统计计算逻辑: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] 统计数据与商家后台数据一致

### 文档同步

- [ ] API 文档已更新（如有变更）
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
grep -c "^- \[" docs/shared/specs/active/story-cashier-business-data-product-summary-category-sales/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-cashier-business-data-product-summary-category-sales/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-cashier-business-data-product-summary-category-sales/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-cashier-business-data-product-summary-category-sales/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-cashier-business-data-product-summary-category-sales/tasks.md)" | bc
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
**最后更新**: 2025-12-08  
**维护者**: 后端开发组

