# 批量重新生成订单材料消耗和POS发票 任务分解

> 本文档定义批量重新生成订单材料消耗和POS发票功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 25  
**已完成**: 20  
**进行中**: -  
**完成率**: 80%

---

## Phase 1: 任务清单管理（Task Manager）

### DTO 定义

- [x] 1.1 创建任务清单DTO结构

  - File: `main/app/dto/batch_regenerate_task.go`
  - Purpose: 定义任务清单的JSON结构体，包括公司、日期、订单、步骤四层结构
  - Requirements: 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有DTO: `main/app/dto/req/`, `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 BatchRegenerateTask 相关的DTO结构体，包括 BatchRegenerateTask, CompanyTask, DateTask, OrderTask, StepTask, TaskSummary | Context: 参考 requirements.md 中的JSON结构，使用 json 标签，Status 字段使用 string 类型（pending/running/completed/failed） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 结构体创建成功，字段定义正确，JSON标签正确

- [x] 1.2 创建进度响应DTO

  - File: `main/app/dto/resp/batch_regenerate_resp.go`
  - Purpose: 定义进度显示的响应结构体
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: 现有响应DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 BatchRegenerateProgressResp 相关的响应结构体，包括 BatchRegenerateProgressResp, CompanyProgress, DateProgress, OrderProgress | Context: 包含总体进度、完成步骤数、失败步骤数、待执行步骤数、预计剩余时间等字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 响应DTO创建成功，字段定义完整

### Service 接口和实现

- [x] 1.3 创建 Task Manager Service 接口

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 定义任务清单管理的Service接口
  - Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 4.1
  - Leverage: 现有Service接口: `main/app/service/sales_outbound_summary_service.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 IBatchRegenerateTaskManager 接口，定义 GenerateTaskList, LoadTaskList, SaveTaskList, ExecuteTaskList, GetProgress 方法 | Context: 接口方法需要支持任务清单的生成、加载、保存、执行和进度查询 | Restrictions: 遵循 .cursor/rules/go-main.mdc，接口以 I 开头 | Success: 接口定义完整，方法签名正确

- [x] 1.4 实现任务清单生成逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现根据公司列表和起始日期生成任务清单的逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有Repository: `main/app/repository/company.go`, `main/app/repository/order.go`
  - Prompt: Role: Go Developer | Task: 实现 GenerateTaskList 方法，查询订单、按日期分组、生成四层JSON结构 | Context: 查询符合条件的订单（status=1已结账 AND created_at >= startDate AND delete_time=0），按 created_at 的日期部分分组，为每个日期生成日期级别步骤，为每个订单生成3个订单步骤 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用事务确保数据一致性 | Success: 任务清单生成成功，结构正确，包含所有必需字段

- [x] 1.5 实现任务清单加载和保存逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现任务清单的JSON文件读写逻辑
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: Go标准库: `encoding/json`, `os`, `utils.CreateFile`, `utils.IsFileExist`
  - Prompt: Role: Go Developer | Task: 实现 LoadTaskList 和 SaveTaskList 方法，支持JSON文件的读取和写入 | Context: LoadTaskList 需要解析JSON文件并验证格式，SaveTaskList 需要序列化任务清单并写入文件 | Restrictions: 验证JSON格式和必需字段，使用格式化JSON输出便于阅读 | Success: 任务清单加载和保存成功，格式验证正确

- [x] 1.6 实现任务清单状态更新逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现任务清单中步骤状态的更新逻辑
  - Requirements: 2.6, 2.7, 2.8
  - Leverage: Task 1.5 的实现
  - Prompt: Role: Go Developer | Task: 实现更新任务清单中步骤状态的逻辑，包括更新 status, start_time, end_time, error 字段 | Context: 需要支持更新指定公司、日期、订单、步骤的状态，更新后保存到文件 | Restrictions: 使用文件锁防止并发操作，定期保存状态 | Success: 状态更新逻辑正确，文件保存成功

---

## Phase 2: 任务执行引擎

- [x] 2.1 实现任务执行引擎主逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现任务执行引擎，按公司→日期→订单→步骤的顺序执行
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Task 1.4-1.6 的实现，现有Service: `main/app/service/sales_outbound_summary_service.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 ExecuteTaskList 方法，遍历公司→日期→订单→步骤，调用对应的Service方法执行步骤 | Context: 按顺序遍历，跳过状态为 completed 的步骤，执行 pending 或 failed 的步骤，每个步骤执行完成后更新状态 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用文件锁防止并发，定期保存状态 | Success: 任务执行引擎工作正常，执行顺序正确

- [x] 2.2 实现订单步骤执行逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现订单级别步骤的执行逻辑（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice）
  - Requirements: 2.3, 2.4
  - Leverage: 现有Service方法: `salesOutboundSummarySrv.RegenerateOrderMaterial()`, `salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound()`, `salesOutboundSummarySrv.RegenerateOrderPosInvoice()`
  - Prompt: Role: Go Developer | Task: 实现订单步骤执行逻辑，根据步骤名称调用对应的Service方法 | Context: Step 1 调用 RegenerateOrderMaterial，Step 2 调用 RegenerateSaleBillMaterialOutbound，Step 3 调用 RegenerateOrderPosInvoice，传入 companyUuid 和 saleOrderUuid | Restrictions: 捕获错误并记录到步骤的 error 字段，更新步骤状态 | Success: 订单步骤执行成功，错误处理正确

- [x] 2.3 实现日期级别步骤执行时机检查

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现检查该日期下所有订单的所有步骤是否都已完成（completed或failed）
  - Requirements: 2.4, 2.5
  - Leverage: Task 2.1, 2.2 的实现
  - Prompt: Role: Go Developer | Task: 实现检查日期下所有订单步骤完成状态的逻辑，判断是否所有订单的所有步骤都已完成（completed或failed） | Context: 遍历该日期下的所有订单，检查每个订单的所有步骤状态，如果所有步骤都是 completed 或 failed，则返回 true | Restrictions: 逻辑正确，性能优化 | Success: 检查逻辑正确，执行时机准确

- [x] 2.4 实现日期级别步骤执行逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现日期级别步骤的执行逻辑（regenerate-sales-outbound）
  - Requirements: 2.4, 2.5
  - Leverage: 现有Service方法: `salesOutboundSummarySrv.RegenerateSalesOutboundSummary()`
  - Prompt: Role: Go Developer | Task: 实现日期级别步骤执行逻辑，调用 RegenerateSalesOutboundSummary 方法 | Context: 传入 companyUuid 和 date，执行完成后更新日期级别步骤的状态 | Restrictions: 捕获错误并记录，更新步骤状态 | Success: 日期级别步骤执行成功，错误处理正确

- [x] 2.5 实现文件锁机制

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现文件锁机制，防止多个实例同时操作同一任务清单文件
  - Requirements: 2.4
  - Leverage: 现有锁机制: `pkg/lock/system_lock.go`
  - Prompt: Role: Go Developer | Task: 实现文件锁机制，使用 TryLockUuidString 和 UnlockUuidString 方法 | Context: 锁Key使用 `batch_regenerate_task:{taskFilePath}`，在 ExecuteTaskList 开始时获取锁，结束时释放锁 | Restrictions: 确保锁的正确获取和释放，防止死锁 | Success: 文件锁机制工作正常，防止并发操作

---

## Phase 3: 断点续传功能

- [ ] 3.1 实现任务清单加载和验证逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现从现有任务清单文件加载并验证的逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 1.5 的实现
  - Prompt: Role: Go Developer | Task: 实现任务清单加载和验证逻辑，验证JSON格式、必需字段、状态值有效性 | Context: LoadTaskList 方法需要验证文件存在性、JSON格式正确性、必需字段存在性、状态值有效性（pending/running/completed/failed） | Restrictions: 验证失败时返回明确的错误信息 | Success: 任务清单加载和验证成功，错误处理正确

- [ ] 3.2 实现已完成步骤跳过逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现跳过状态为 completed 的步骤的逻辑
  - Requirements: 3.2, 3.4
  - Leverage: Task 2.1 的实现
  - Prompt: Role: Go Developer | Task: 实现跳过已完成步骤的逻辑，在执行步骤前检查状态，如果为 completed 则跳过 | Context: 在 ExecuteTaskList 中，执行每个步骤前检查步骤状态，如果为 completed 则跳过，继续下一个步骤 | Restrictions: 逻辑正确，不影响其他步骤执行 | Success: 已完成步骤跳过逻辑正确，不影响执行流程

- [ ] 3.3 实现失败步骤重新执行逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现重新执行状态为 failed 的步骤的逻辑
  - Requirements: 3.3, 3.4
  - Leverage: Task 2.1, 2.2, 2.4 的实现
  - Prompt: Role: Go Developer | Task: 实现失败步骤重新执行逻辑，遇到状态为 failed 的步骤时重新执行 | Context: 在 ExecuteTaskList 中，遇到状态为 failed 的步骤时，重置状态为 pending 并重新执行 | Restrictions: 重新执行前清除之前的错误信息 | Success: 失败步骤重新执行逻辑正确

---

## Phase 4: 进度显示功能（可选）

- [x] 4.1 实现进度计算逻辑

  - File: `main/app/service/batch_regenerate_task_manager.go`
  - Purpose: 实现计算各级别进度和总体统计的逻辑
  - Requirements: 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: Task 1.1, 1.2 的实现
  - Prompt: Role: Go Developer | Task: 实现 GetProgress 方法，计算公司/日期/订单/步骤级别的进度和总体统计 | Context: 遍历任务清单，统计已完成、失败、待执行的步骤数，计算完成百分比，计算预计剩余时间（基于已完成步骤的平均耗时） | Restrictions: 计算逻辑正确，性能优化 | Success: 进度计算逻辑正确，统计信息准确

- [x] 4.2 实现进度显示格式化

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现进度信息的格式化输出
  - Requirements: 4.3, 4.4
  - Leverage: Task 4.1 的实现，参考 `regenerate_order_material.go` 的彩色输出
  - Prompt: Role: Go Developer | Task: 实现进度信息的格式化输出，使用颜色和格式化展示进度信息 | Context: 输出总体进度、公司进度、日期进度、订单进度、步骤统计等信息，使用颜色区分不同状态（已完成/进行中/未开始） | Restrictions: 输出格式清晰，易于阅读 | Success: 进度显示格式化正确，输出清晰

- [x] 4.3 实现定时刷新机制

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现进度信息的定时刷新机制
  - Requirements: 4.7
  - Leverage: Go标准库: `time.Ticker`
  - Prompt: Role: Go Developer | Task: 实现进度信息的定时刷新机制，使用 Ticker 定期更新进度显示 | Context: 默认每5秒刷新一次，可通过 progressInterval 参数配置，使用 goroutine 异步更新，不阻塞主执行流程 | Restrictions: 异步更新不阻塞主流程，支持配置刷新间隔 | Success: 定时刷新机制工作正常，不阻塞主流程

---

## Phase 5: 命令行工具

- [x] 5.1 创建命令文件框架

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 创建命令行工具的基础框架，包括命令定义、参数解析和初始化逻辑
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有命令: `main/command/regenerate_order_material.go`, `main/command/regenerate_sales_outbound.go`
  - Prompt: Role: Go Developer specializing in CLI tools | Task: 创建 batch-regenerate-order-material-and-invoice 命令的基础框架，参考 regenerate_order_material.go 的结构 | Context: 使用 Cobra 框架，定义命令名称、描述、参数（company-uuids, start-date, task-file, resume, dry-run, show-progress, progress-interval），实现 PreRun 初始化逻辑（配置、日志、数据库、缓存、锁等） | Restrictions: 遵循 .cursor/rules/go-main.mdc，命令文件放在 main/command/ 目录 | Success: 命令框架创建成功，参数解析正确，PreRun 初始化完整

- [x] 5.2 实现参数解析和验证

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现命令行参数的解析和验证逻辑
  - Requirements: 5.1, 5.2
  - Leverage: Task 5.1 的命令框架
  - Prompt: Role: Go Developer | Task: 实现参数解析和验证逻辑，验证 company-uuids 和 start-date 必填，解析 company-uuids 为数组，验证日期格式 | Context: 解析 company-uuids 字符串（逗号分隔）为 uint64 数组，验证 start-date 格式为 YYYY-MM-DD，验证 task-file 路径有效性（如果提供） | Restrictions: 参数验证失败时输出明确的错误信息并退出 | Success: 参数解析和验证正确，错误处理完善

- [x] 5.3 实现任务清单生成和执行逻辑

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现任务清单的生成和执行逻辑
  - Requirements: 1.1, 1.2, 2.1, 2.2, 3.1
  - Leverage: Task 5.1, 5.2 的实现，Task Manager Service
  - Prompt: Role: Go Developer | Task: 实现任务清单生成和执行逻辑，集成 Task Manager Service | Context: 如果不是 resume 模式，调用 GenerateTaskList 生成任务清单；如果是 resume 模式，调用 LoadTaskList 加载任务清单；如果不是 dry-run 模式，调用 ExecuteTaskList 执行任务 | Restrictions: 错误处理完善，输出友好的提示信息 | Success: 任务清单生成和执行逻辑正确，错误处理完善

- [x] 5.4 实现 dry-run 预览模式

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现 dry-run 预览模式，仅生成任务清单不执行
  - Requirements: 1.7
  - Leverage: Task 5.3 的实现
  - Prompt: Role: Go Developer | Task: 实现 dry-run 预览模式，生成任务清单后输出统计信息并退出 | Context: 如果使用 --dry-run 参数，生成任务清单后输出统计信息（总公司数、总日期数、总订单数、总步骤数等），不实际执行任务 | Restrictions: 输出格式清晰，统计信息准确 | Success: dry-run 预览模式工作正常，输出信息准确

- [x] 5.5 实现进度显示集成

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 集成进度显示功能到命令行工具
  - Requirements: 4.1, 4.7
  - Leverage: Task 4.1-4.3 的实现，Task 5.3 的实现
  - Prompt: Role: Go Developer | Task: 集成进度显示功能到命令行工具，如果启用 --show-progress 则显示进度信息 | Context: 在执行任务时，如果启用 --show-progress，调用 GetProgress 获取进度信息并格式化输出，按照 --progress-interval 指定的间隔刷新 | Restrictions: 进度显示不阻塞主执行流程，支持实时更新 | Success: 进度显示集成成功，实时更新正常

- [x] 5.6 实现日志输出和错误处理

  - File: `main/command/batch_regenerate_order_material_and_invoice.go`
  - Purpose: 实现详细的日志输出和错误处理
  - Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6
  - Leverage: Task 5.1-5.5 的实现，参考 `regenerate_order_material.go` 的日志输出
  - Prompt: Role: Go Developer | Task: 实现详细的日志输出和错误处理，包括控制台输出和文件日志 | Context: 使用彩色输出（blueColor, yellowColor, redColor, greenColor）区分不同状态，记录详细日志到文件（logs/batch-regenerate-{timestamp}.log），输出统计信息（总公司数、总日期数、总订单数、完成数、失败数、耗时等） | Restrictions: 使用 logger.Logger 记录日志，错误信息清晰友好 | Success: 日志输出和错误处理完善，统计信息准确

---

## Phase 6: 测试

- [ ] 6.1 编写 Task Manager Service 单元测试

  - File: `main/app/service/batch_regenerate_task_manager_test.go`
  - Purpose: 确保 Task Manager Service 的业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 BatchRegenerateTaskManager 编写单元测试，覆盖率 ≥ 70% | Context: 测试任务清单生成、加载、保存、执行、进度计算等逻辑，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 6.2 编写命令行工具集成测试

  - File: `main/command/batch_regenerate_order_material_and_invoice_test.go`
  - Purpose: 测试命令行工具的端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/command/`
  - Prompt: Role: QA Automation Engineer | Task: 为命令行工具编写集成测试，测试端到端流程 | Context: 测试任务清单生成、任务执行、断点续传、进度显示等功能，测试各种参数组合 | Restrictions: 测试真实场景，覆盖边界情况 | Success: 集成测试通过，覆盖核心流程

- [ ] 6.3 编写断点续传功能测试

  - File: `test/integration/batch_regenerate_resume_test.go`
  - Purpose: 测试断点续传功能的各种场景
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有集成测试
  - Prompt: Role: QA Engineer | Task: 编写断点续传功能的测试用例，覆盖各种场景 | Context: 测试从已完成任务清单继续执行、跳过已完成步骤、重新执行失败步骤等场景 | Restrictions: 测试真实场景，覆盖边界情况 | Success: 断点续传功能测试通过

- [ ] 6.4 编写进度显示功能测试（可选）

  - File: `test/integration/batch_regenerate_progress_test.go`
  - Purpose: 测试进度显示功能的正确性
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7
  - Leverage: 现有集成测试
  - Prompt: Role: QA Engineer | Task: 编写进度显示功能的测试用例，验证进度计算的正确性 | Context: 测试进度计算逻辑、进度显示格式化、定时刷新机制等 | Restrictions: 测试真实场景 | Success: 进度显示功能测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Command: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 命令行文档已更新（使用说明和参数说明）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-batch-regenerate-order-material-and-invoice/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-batch-regenerate-order-material-and-invoice/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-batch-regenerate-order-material-and-invoice/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-batch-regenerate-order-material-and-invoice/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-batch-regenerate-order-material-and-invoice/tasks.md)" | bc
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
**最后更新**: 2025-12-17  
**维护者**: xiezhihuan

