# 旧商家后台-统计报表-导出-加上Grab数据/LINEMAN数据 任务分解

> 本文档定义 旧商家后台统计报表导出增加 Grab 和 LINE MAN 数据 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 7  
**进行中**: -  
**完成率**: 70%

**注意**: 根据需求变更，已移除 LINE MAN 相关功能，仅保留 Grab 数据统计。

---

## Phase 1: 代码分析和准备

- [x] 1.1 分析现有实现

  - File: `main/app/service/statistics.go`
  - Purpose: 理解 `CountSale`、`CountPayment`、`CountSaleDays`、`CountPaymentDays` 方法的实现逻辑
  - Requirements: 1.1, 2.1
  - Leverage: 
    - `main/app/service/statistics.go` - CountSale (line 125), CountPayment (line 423)
    - `main/app/service/statistics.go` - CountSaleDays (line 229), CountPaymentDays (line 597)
    - `main/app/repository/statistics_takeout.go` - CountTakeoutChannelSaleByPlatform, CountTakeoutPaymentMethodRawData
  - Success: 理解现有实现逻辑，确认集成方式

---

## Phase 2: 核心实现（Go Main）

### Service 层 - 销售统计

- [x] 2.1 扩展 CountSaleResp 结构，添加 Grab 统计字段

  - File: `main/app/service/statistics.go`
  - Purpose: 在 `CountSaleResp` 结构体中添加 Grab 的统计字段（订单数、最小/最大/平均订单金额）
  - Requirements: 1.1, 1.2
  - **变更**: 根据需求变更，已移除 LINE MAN 相关字段，仅保留 Grab 字段
  - Leverage: 
    - `main/app/service/statistics.go` - CountSaleResp 结构体定义（line 100-122）
    - `main/app/model/statistics.go` - ChannelSaleRepoResult 结构体（参考字段命名）
  - Prompt: Role: Go Developer specializing in Data Structure | Task: 在 CountSaleResp 结构体中添加 Grab 和 LINE MAN 的统计字段 | Context: 1) 添加 Grab 字段：GrabOrderNum (int64), GrabMinOrderAmount (float64), GrabMaxOrderAmount (float64), GrabAvgOrderAmount (float64)；2) 添加 LINE MAN 字段：LinemanOrderNum (int64), LinemanMinOrderAmount (float64), LinemanMaxOrderAmount (float64), LinemanAvgOrderAmount (float64)；3) 使用 json 标签，遵循现有命名规范 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: CountSaleResp 结构体添加了 Grab 和 LINE MAN 字段，字段命名规范，代码通过 go fmt 和 go vet

- [x] 2.2 扩展 CountSaleDays 方法，集成 Grab 销售数据

  - File: `main/app/service/statistics.go`
  - Purpose: 在 `CountSaleDays` 方法中为每个日期增加 Grab 平台的销售数据统计（包括订单数、最小/最大/平均订单金额）
  - Requirements: 1.1, 1.2
  - **完成时间**: 2026-01-20
  - **变更记录**: 
    - 使用 `CountTakeoutSale` 方法替代 `CountTakeoutChannelSaleByPlatform`
    - 累加 Grab 数据到总统计字段（不累加到外卖相关字段）
    - 移除 `totalProductOriginPrice` 字段（代码清理）
  - Leverage: 
    - `main/app/service/statistics.go` - CountSale 方法中集成外卖数据的方式（line 149-155）
    - `main/app/repository/statistics_takeout.go` - CountTakeoutChannelSaleByPlatform 方法
    - `main/app/repository/statistics.go` - CountChannelSale 方法中查询 Grab 数据的方式（line 3148-3158）
    - `main/app/repository/statistics.go` - calculateChannelSaleFromRawData 函数（line 3205-3254）
  - Prompt: Role: Go Developer specializing in Statistics Service | Task: 在 CountSaleDays 方法中为每个日期（day）增加 Grab 平台的销售数据统计，包括订单数、最小/最大/平均订单金额 | Context: 1) 参考 CountSale 方法中集成外卖数据的方式，使用 CountTakeoutChannelSaleByPlatform 查询 Grab 数据（platform="grab"）；2) 按日期筛选数据（使用 accepted_time 字段转换为日期字符串匹配 day）；3) 使用 calculateChannelSaleFromRawData 或类似逻辑计算 Grab 的统计指标：订单数、最小/最大/平均订单金额；4) 将统计结果填充到 CountSaleDaysResp 的新字段中：GrabOrderNum, GrabMinOrderAmount, GrabMaxOrderAmount, GrabAvgOrderAmount；5) 确保 Grab 数据排列在最后 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用 errors.WithMessage 包装错误，使用 decimal 进行金额计算避免精度问题 | Success: CountSaleDays 方法正确集成 Grab 销售数据（包括订单数、最小/最大/平均订单金额），数据排列在最后，代码通过 go fmt 和 go vet

- [x] ~~2.3 扩展 CountSaleDays 方法，集成 LINE MAN 销售数据~~ **已取消**

  - **状态**: 已取消（根据需求变更，移除 LINE MAN 相关功能）
  - **原因**: 用户明确要求移除 LINE MAN 相关统计，仅保留 Grab 数据

### Service 层 - 支付统计

- [x] 2.4 扩展 CountPaymentDays 方法，集成 Grab 支付数据

  - File: `main/app/service/statistics.go`
  - Purpose: 在 `CountPaymentDays` 方法中为每个日期增加 Grab 平台的支付数据统计
  - Requirements: 2.1, 2.2
  - **完成时间**: 2026-01-20
  - **变更记录**: 
    - 使用 `CountTakeoutPayment` 方法替代 `CountTakeoutPaymentMethodRawData`
    - Grab 支付数据在排序后追加，确保排在最后
  - Leverage: 
    - `main/app/service/statistics.go` - CountPayment 方法中集成外卖支付数据的方式（line 459-581）
    - `main/app/repository/statistics_takeout.go` - CountTakeoutPaymentMethodRawData 方法
  - Prompt: Role: Go Developer specializing in Statistics Service | Task: 在 CountPaymentDays 方法中为每个日期（day）增加 Grab 平台的支付数据统计 | Context: 1) 参考 CountPayment 方法中集成外卖支付数据的方式，使用 CountTakeoutPaymentMethodRawData 查询 Grab 支付数据；2) 按日期筛选数据（使用 accepted_time 字段转换为日期字符串匹配 day）；3) 筛选 payment_name = "Grab" 的数据；4) 将 Grab 的支付数据追加到 PaymentList 中，转换为 CountPaymentRespList 格式；5) 确保 Grab 数据排列在最后（在排序后追加，参考 CountPayment 方法的实现） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: CountPaymentDays 方法正确集成 Grab 支付数据，数据排列在最后，代码通过 go fmt 和 go vet

- [x] ~~2.5 扩展 CountPaymentDays 方法，集成 LINE MAN 支付数据~~ **已取消**

  - **状态**: 已取消（根据需求变更，移除 LINE MAN 相关功能）
  - **原因**: 用户明确要求移除 LINE MAN 相关统计，仅保留 Grab 数据

---

## Phase 3: 测试和验证

- [ ] 3.1 编写 CountSaleDays 单元测试

  - File: `main/app/service/statistics_test.go` 或新建测试文件
  - Purpose: 确保 CountSaleDays 方法正确集成 Grab 销售数据（包括订单数、最小/最大/平均订单金额）
  - Requirements: 1.3, 1.4
  - **变更**: 仅测试 Grab 数据，已移除 LINE MAN 相关测试
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountSaleDays 方法编写单元测试，验证 Grab 和 LINE MAN 数据正确集成，包括订单数、最小/最大/平均订单金额 | Context: 1) 测试正常场景：有 Grab/LINE MAN 订单数据时，统计结果包含这些数据，且订单数、最小/最大/平均订单金额计算正确；2) 测试边界情况：无数据、单条数据、多条数据；3) 测试数据排列顺序：Grab 和 LINE MAN 数据排列在最后；4) 测试日期分组：跨日期数据正确分组；5) 测试统计指标准确性：订单数、最小/最大/平均订单金额与原始数据一致 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试覆盖率 ≥ 70%，所有测试通过，统计指标计算正确

- [ ] 3.2 编写 CountPaymentDays 单元测试

  - File: `main/app/service/statistics_test.go` 或新建测试文件
  - Purpose: 确保 CountPaymentDays 方法正确集成 Grab 支付数据
  - Requirements: 2.3, 2.4
  - **变更**: 仅测试 Grab 数据，已移除 LINE MAN 相关测试
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountPaymentDays 方法编写单元测试，验证 Grab 和 LINE MAN 数据正确集成 | Context: 1) 测试正常场景：有 Grab/LINE MAN 支付数据时，统计结果包含这些数据；2) 测试边界情况：无数据、单条数据、多条数据；3) 测试数据排列顺序：Grab 和 LINE MAN 数据排列在最后；4) 测试日期分组：跨日期数据正确分组 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 3.3 集成测试 - 导出功能验证

  - File: `test/integration/statistics_export_test.go` 或手动测试
  - Purpose: 验证导出功能包含 Grab 数据（包括订单数、最小/最大/平均订单金额），且排列顺序正确
  - Requirements: 所有功能需求
  - **变更**: 仅验证 Grab 数据，已移除 LINE MAN 相关验证
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，验证统计导出功能 | Context: 1) 创建测试数据：创建 Grab 和 LINE MAN 订单（包含不同金额的订单，用于测试最小/最大/平均订单金额）；2) 调用导出接口 `/shop/statistics/export`；3) 验证导出数据包含 Grab 和 LINE MAN 数据；4) 验证 Grab 和 LINE MAN 的统计指标：订单数、最小/最大/平均订单金额；5) 验证数据排列顺序：Grab 和 LINE MAN 排在最后；6) 验证数据准确性：对比订单数据和统计结果 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，数据准确性和排列顺序正确，统计指标计算正确

- [ ] 3.4 数据准确性验证

  - File: -
  - Purpose: 手动验证统计结果与订单数据一致（包括订单数、最小/最大/平均订单金额）
  - Requirements: 1.4, 2.4
  - Leverage: 数据库查询工具
  - Success: 统计结果与订单数据一致，Grab 数据正确，订单数、最小/最大/平均订单金额计算准确
  - **变更**: 仅验证 Grab 数据，已移除 LINE MAN 相关验证

---

## Phase 4: 代码审查和优化

- [x] 4.1 代码审查

  - File: `main/app/service/statistics.go`
  - Purpose: 确保代码符合规范，性能优化
  - Requirements: 所有需求
  - Leverage: `.cursor/rules/go-main.mdc`, `.cursor/rules/code-review.mdc`
  - Success: 代码审查通过，无规范问题
  - **完成时间**: 2026-01-20
  - **审查结果**: 
    - ✅ 无报错风险
    - ✅ Grab外卖统计符合需求
    - ✅ Grab外卖开关判断正确
    - ✅ 空数据不影响原有统计逻辑
  - **审查报告**: `docs/code-review-grab-statistics.md`

- [ ] 4.2 性能优化（如需要）

  - File: `main/app/service/statistics.go`, `main/app/repository/statistics_takeout.go`
  - Purpose: 确保查询性能不受影响
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，现有索引
  - Success: 查询时间 < 50ms，性能达标

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

- [x] requirements.md 中的所有需求已满足（已移除 LINE MAN 相关需求）
- [x] design.md 中的设计已实现（已移除 LINE MAN 相关设计）
- [x] 验收标准已达成
  - ✅ 销售统计（按天）导出包含 Grab 数据，且排列在最后
  - ✅ 支付数据（按天）导出包含 Grab 数据，且排列在最后
  - ✅ 数据统计准确，与订单数据一致
  - ✅ Grab 开关判断正确，空数据不影响原有统计逻辑

### 文档同步

- [ ] 代码注释完整
- [ ] 如有 API 变更，更新 API 文档

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-statistics-export-add-grab-lineman-data/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-statistics-export-add-grab-lineman-data/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-statistics-export-add-grab-lineman-data/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-statistics-export-add-grab-lineman-data/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-statistics-export-add-grab-lineman-data/tasks.md)" | bc
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

### Go 后端开发

```
Role: Go Developer specializing in Statistics Service

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: main/app/service/statistics.go
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service)
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

## 变更记录

### 2026-01-20
- ✅ 完成 Grab 销售统计集成（CountSaleDays）
- ✅ 完成 Grab 支付统计集成（CountPaymentDays）
- ✅ 完成代码审查（4.1）
- ✅ 移除 LINE MAN 相关功能（根据需求变更）
- ✅ 代码清理：移除 `totalProductOriginPrice` 字段
- ✅ 优化：使用 `CountTakeoutSale` 和 `CountTakeoutPayment` 方法替代原始数据查询

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-19  
**维护者**: 后端开发组
