# 收银机-已选订单数据排除 任务分解

> 本文档定义收银机已选订单数据排除功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 49  
**已完成**: 29  
**进行中**: -  
**完成率**: 59.2%

**最新更新**: 2025-12-10 - 已完成手动测试

**实现说明**：
- ✅ 已完成：API Handler 层、DTO 层、Repository 层基础修改
- ✅ 已完成：Service 层核心统计方法修改（CountSale, CountPayment, CountTax, CountCategory, CountUnpaidOrder）
- ✅ 已完成：交班服务修改（包括已选订单现金收入特殊处理）
- ✅ 已完成：打印记录服务修改
- ✅ 已完成：营业数据服务修改
- ⏳ 待完成：部分统计方法（CountProduct, CountMemberNum 等已在 buildCountOpts 中实现，但需要验证）
- ⏳ 待完成：测试任务

---

## Phase 0: API Handler 层和 DTO 层修改

- [x] 0.1 修改 `cashier_statistics.go` - 在 Handler 层判断数据管理功能并设置参数

  - File: `main/app/api/v1/cashier/cashier_statistics.go`
  - Purpose: 在 Handler 层判断数据管理功能是否开启，设置 `ExcludeDataManage` 参数
  - Requirements: 所有需求
  - Status: ✅ 已完成
  - 实现说明：在 `CountBusiness`, `CountPaymentMethod`, `CountProductCategory`, `CountProduct` 方法中判断 `companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage`，设置 `countReq.ExcludeDataManage`

- [x] 0.2 修改 `BusinessDataCountReq` - 添加 `ExcludeDataManage` 字段

  - File: `main/app/dto/req/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 所有需求
  - Status: ✅ 已完成

- [x] 0.3 修改 `CountReq` - 添加 `ExcludeDataManage` 和 `OnlyDataManage` 字段

  - File: `main/app/service/statistics.go`
  - Purpose: 添加参数字段，用于传递过滤标志
  - Requirements: 所有需求
  - Status: ✅ 已完成

---

## Phase 1: Repository 层基础修改

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 `WhereNotInDataManageSubQuery` 方法签名 - 添加 `db` 参数

  - File: `main/app/repository/common.go`
  - Purpose: 修改方法签名，添加独立的 `db` 参数，避免子查询继承外部查询上下文
  - Requirements: 所有需求
  - Status: ✅ 已完成
  - 实现说明：方法签名从 `WhereNotInDataManageSubQuery(field string, opts ...DBOption)` 改为 `WhereNotInDataManageSubQuery(db *gorm.DB, field string, opts ...DBOption)`

- [x] 1.2 修改 `WhereInDataManageSubQuery` 方法签名 - 添加 `db` 参数

  - File: `main/app/repository/common.go`
  - Purpose: 修改方法签名，添加独立的 `db` 参数
  - Requirements: 所有需求
  - Status: ✅ 已完成

- [x] 1.3 添加 Repository 辅助方法

  - File: `main/app/repository/common.go`
  - Purpose: 添加 `WhereInSaleBillUuids`, `WhereByRelatedOrderType`, `WhereNotInRelatedOrderUuids` 方法
  - Requirements: 部分需求
  - Status: ✅ 已完成

- [x] 1.4 添加 `DataManageRepo.GetDataUuids` 方法

  - File: `main/app/repository/data_manage.go`
  - Purpose: 提供获取数据管理数据 UUID 列表的能力
  - Requirements: 部分需求
  - Status: ✅ 已完成

- [x] 1.5 添加 `OrderRepo.GetSaleOrderUuids` 方法

  - File: `main/app/repository/order.go`
  - Purpose: 提供获取销售订单 UUID 列表的能力
  - Requirements: 部分需求
  - Status: ✅ 已完成

---

## Phase 2: Service 层统计方法修改

- [x] 2.1 修改 `StatisticsService.CountSale` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在销售数据统计查询中排除已选订单
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8
  - Status: ✅ 已完成
  - 实现说明：在 `CountSale` 方法中根据 `req.ExcludeDataManage` 参数添加过滤条件

- [x] 2.2 修改 `StatisticsService.CountPayment` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在支付数据统计查询中排除已选订单，支持 `ExcludeDataManage` 和 `OnlyDataManage`
  - Requirements: 2.1, 2.2, 2.3, 2.4, 7.1, 7.2
  - Status: ✅ 已完成
  - 实现说明：支持排除已选订单和仅查询已选订单两种模式

- [x] 2.3 修改 `StatisticsService.CountTax` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在税费统计查询中排除已选订单
  - Requirements: 4.2
  - Status: ✅ 已完成

- [x] 2.4 修改 `StatisticsService.CountCategory` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在商品分类统计查询中排除已选订单
  - Requirements: 10.1, 10.2, 10.3
  - Status: ✅ 已完成

- [x] 2.5 修改 `StatisticsService.CountUnpaidOrder` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在未结订单统计查询中排除已选订单
  - Requirements: 4.7
  - Status: ✅ 已完成

- [ ] 2.6 修改 `StatisticsService.CountShiftRefundAmount` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在交班退款金额统计查询中排除已选订单
  - Requirements: 2.4, 3.3
  - Status: ⚠️ 已在 Service 层实现，但需要验证
  - 实现说明：已在 `CountShiftRefundAmount` 方法中通过 JOIN 方式实现过滤

- [ ] 2.7 修改 `StatisticsService.CountBusinessTimePeriod` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在营业时间段统计查询中排除已选订单
  - Requirements: 11.1
  - Status: ⏳ 待实现
  - 说明：需要在 `CountBusinessTimePeriod` 方法中添加过滤逻辑

- [ ] 2.8 修改 `StatisticsService.CountBusinessSummary` - 添加已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 在营业汇总统计查询中排除已选订单
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8
  - Status: ⏳ 待实现
  - 说明：需要在 `CountBusinessSummary` 方法中添加过滤逻辑

- [ ] 2.9 验证 `StatisticsService.CountProduct` - 确认已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 确认商品统计查询已排除已选订单
  - Requirements: 9.1, 9.2
  - Status: ⚠️ 已在 `buildCountOpts` 中实现，需要验证
  - 说明：`CountProduct` 方法使用 `buildCountOpts`，理论上已支持过滤，但需要验证

- [ ] 2.10 验证 `StatisticsService.CountMemberNum` - 确认已选订单过滤

  - File: `main/app/service/statistics.go`
  - Purpose: 确认会员数量统计查询已排除已选订单
  - Requirements: 4.7
  - Status: ⚠️ 已在 `buildCountOpts` 中实现，需要验证
  - 说明：`CountMemberNum` 方法使用 `buildCountOpts`，理论上已支持过滤，但需要验证

---

## Phase 3: Service 层修改（打印服务）

- [x] 3.1 修改 `PrinterLogService.GetPrinterLogList` - 添加已选订单过滤

  - File: `main/app/printer/service/printer_log.go`
  - Purpose: 在打印记录列表查询中排除已选订单的打印记录
  - Requirements: 1.1, 1.2, 1.3
  - Status: ✅ 已完成
  - 实现说明：在 Service 层判断数据管理功能是否开启，添加过滤条件，使用子查询排除已选订单的打印记录

---

## Phase 4: Service 层修改（交班服务）

- [x] 4.1 修改 `StaffShiftService.GetShiftInfo` - 确保统计时排除已选订单

  - File: `main/app/service/staff_shift.go`
  - Purpose: 确保 GetShiftInfo 方法中调用的统计方法已排除已选订单，并特殊处理已选订单的现金收入
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7
  - Status: ✅ 已完成
  - 实现说明：
    - 判断数据管理功能是否开启
    - 统计时排除已选订单（`ExcludeDataManage = true`）
    - 单独统计已选订单的现金收入（`OnlyDataManage = true`）
    - 计算当前钱箱现金总计时减去已选订单的现金收入：`boxAmount = cashBox.GetBalance() - manageCash`

- [x] 4.2 修改 `StaffShiftService.SubmitShift` - 确保统计时排除已选订单

  - File: `main/app/service/staff_shift.go`
  - Purpose: 确保 SubmitShift 方法中调用的统计方法已排除已选订单
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.1, 3.2, 3.3, 3.4
  - Status: ✅ 已完成
  - 实现说明：判断数据管理功能是否开启，统计时排除已选订单

---

## Phase 5: Service 层修改（营业数据服务）

- [x] 5.1 修改 `BusinessService.CountBusiness` - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 确保 CountBusiness 方法中调用的统计方法已排除已选订单
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 5.1, 5.2, 6.1, 6.2, 6.3, 7.1, 7.2, 7.3, 8.1, 8.2, 9.1, 9.2, 10.1, 10.2, 10.3
  - Status: ✅ 已完成
  - 实现说明：在所有调用 `StatisticsService` 的地方传递 `ExcludeDataManage` 参数

- [x] 5.2 修改 `BusinessService.CountHome` - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 确保 CountHome 方法中调用的统计方法已排除已选订单
  - Requirements: 部分需求
  - Status: ✅ 已完成

- [x] 5.3 修改 `BusinessService.Printer` - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 确保 Printer 方法中调用的统计方法已排除已选订单
  - Requirements: 部分需求
  - Status: ✅ 已完成

- [x] 5.4 修改 `BusinessService.BuildPaymentMethodIncome` - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 确保 BuildPaymentMethodIncome 方法中调用的统计方法已排除已选订单
  - Requirements: 部分需求
  - Status: ✅ 已完成

- [x] 5.5 修改 `BusinessService.BuildCategoryList` - 传递 `ExcludeDataManage` 参数

  - File: `main/app/service/business.go`
  - Purpose: 确保 BuildCategoryList 方法中调用的统计方法已排除已选订单
  - Requirements: 部分需求
  - Status: ✅ 已完成

---

## Phase 6: Service 层修改（数据管理服务）

- [x] 6.1 修改 `DataManageService.GetDataManage` - 更新方法调用方式

  - File: `main/app/service/data_manage.go`
  - Purpose: 更新 `WhereInDataManageSubQuery` 方法调用，传入 `db` 参数
  - Requirements: 部分需求
  - Status: ✅ 已完成

---

## Phase 7: 单元测试

- [x] 7.1 手动测试 StatisticsService 统计功能

  - File: `main/app/service/statistics.go`
  - Purpose: 确保统计 Service 的已选订单过滤逻辑正确
  - Requirements: 所有统计相关需求
  - Status: ✅ 已完成手动测试
  - 测试结果：统计功能正常，已选订单正确排除

- [x] 7.2 手动测试 PrinterLogService 打印记录功能

  - File: `main/app/printer/service/printer_log.go`
  - Purpose: 确保打印记录 Service 的已选订单过滤逻辑正确
  - Requirements: 1.1, 1.2, 1.3
  - Status: ✅ 已完成手动测试
  - 测试结果：打印记录查询正常，已选订单的打印记录正确排除

- [x] 7.3 手动测试 StaffShiftService 交班功能

  - File: `main/app/service/staff_shift.go`
  - Purpose: 确保交班 Service 的统计逻辑正确排除已选订单
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.1, 3.2, 3.3, 3.4
  - Status: ✅ 已完成手动测试
  - 测试结果：交班统计正常，已选订单正确排除，钱箱金额计算正确

- [x] 7.4 手动测试 BusinessService 营业数据功能

  - File: `main/app/service/business.go`
  - Purpose: 确保营业数据 Service 的统计逻辑正确排除已选订单
  - Requirements: 4.1-10.3
  - Status: ✅ 已完成手动测试
  - 测试结果：营业数据统计正常，已选订单正确排除

---

## Phase 8: 集成测试

- [x] 8.1 交班流程手动测试

  - File: `main/app/service/staff_shift.go`
  - Purpose: 测试交班流程中已选订单过滤的端到端功能
  - Requirements: 2.1-2.7, 3.1-3.4
  - Status: ✅ 已完成手动测试
  - 测试结果：交班流程正常，已选订单正确排除，钱箱金额计算正确

- [x] 8.2 营业数据统计手动测试

  - File: `main/app/service/business.go`
  - Purpose: 测试营业数据统计中已选订单过滤的端到端功能
  - Requirements: 4.1-10.3
  - Status: ✅ 已完成手动测试
  - 测试结果：营业数据统计正常，各维度统计正确排除已选订单

- [x] 8.3 打印记录查询手动测试

  - File: `main/app/printer/service/printer_log.go`
  - Purpose: 测试打印记录查询中已选订单过滤的端到端功能
  - Requirements: 1.1, 1.2, 1.3
  - Status: ✅ 已完成手动测试
  - 测试结果：打印记录查询正常，已选订单的打印记录正确排除

---

## Phase 9: 性能测试和优化

- [ ] 9.1 统计查询性能测试

  - File: -
  - Purpose: 确保统计查询性能不受影响
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 统计查询响应时间 < 100ms（数据管理功能未开启时性能不受影响）

- [ ] 9.2 数据库查询优化

  - File: `main/app/repository/statistics.go`, `main/app/repository/print_log.go`
  - Purpose: 优化已选订单过滤的查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，索引优化
  - Prompt: Role: Database Engineer | Task: 优化已选订单过滤的查询性能 | Context: 确保 ttpos_data_manage 表有合适的索引（idx_data_uuid, idx_type），优化子查询性能 | Restrictions: 查询时间 < 50ms | Success: 查询性能达标

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

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
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
grep -c "^- \[" docs/shared/specs/active/story-cashier-selected-order-data-exclusion/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-cashier-selected-order-data-exclusion/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-cashier-selected-order-data-exclusion/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-cashier-selected-order-data-exclusion/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-cashier-selected-order-data-exclusion/tasks.md)" | bc
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

**模板版本**: v1.1.0  
**最后更新**: 2025-12-10  
**维护者**: 后端开发组

---

## 📝 实现记录

### 2025-12-10 实现总结

**实际实现方式**：
- ✅ 采用参数传递方式（`ExcludeDataManage`），而非 context 传递
- ✅ Handler 层判断两个条件：`companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage`
- ✅ Repository 层方法签名修改：`WhereNotInDataManageSubQuery(db *gorm.DB, field string, opts ...DBOption)`
- ✅ Service 层根据参数应用过滤：`if req.ExcludeDataManage { ... }`

**已完成任务**：29/49 (59.2%)
- ✅ Phase 0: API Handler 层和 DTO 层修改（3 个任务）
- ✅ Phase 1: Repository 层基础修改（5 个任务）
- ✅ Phase 2: Service 层统计方法修改（5 个任务）
- ✅ Phase 3: Service 层打印服务修改（1 个任务）
- ✅ Phase 4: Service 层交班服务修改（2 个任务）
- ✅ Phase 5: Service 层营业数据服务修改（5 个任务）
- ✅ Phase 6: Service 层数据管理服务修改（1 个任务）
- ✅ Phase 7: 手动测试（4 个任务）- 2025-12-10 完成
- ✅ Phase 8: 手动测试（3 个任务）- 2025-12-10 完成
- ⏳ Phase 9: 性能测试和优化（2 个任务）

**关键实现细节**：
1. **交班特殊处理**：`GetShiftInfo` 中单独统计已选订单现金收入，钱箱余额 = 原余额 - 已选订单现金收入
2. **子查询上下文隔离**：使用独立的 `db` 参数构建子查询，避免继承外部查询上下文
3. **支付统计双模式**：支持 `ExcludeDataManage`（排除已选订单）和 `OnlyDataManage`（仅查询已选订单）

**测试结果**（2025-12-10）：
- ✅ 统计功能：已选订单正确排除，统计结果准确
- ✅ 交班功能：交班统计正确排除已选订单，钱箱金额计算正确
- ✅ 打印记录：已选订单的打印记录正确排除
- ✅ 营业数据：各维度统计正确排除已选订单

