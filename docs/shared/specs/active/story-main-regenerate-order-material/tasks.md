# 重新生成订单材料用料命令 任务分解

> 本文档定义重新生成订单材料用料命令功能的详细执行任务清单。

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

## Phase 1: 命令框架搭建

- [x] 1.1 创建命令文件框架

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 创建命令行工具的基础框架，包括命令定义、参数解析和初始化逻辑
  - Requirements: 4.1, 4.2, 4.4
  - Leverage: 现有命令: `main/command/regenerate_sales_outbound.go`
  - Prompt: Role: Go Developer specializing in CLI tools | Task: 创建 regenerate-order-material 命令的基础框架，参考 regenerate_sales_outbound.go 的结构 | Context: 使用 Cobra 框架，定义命令名称、描述、参数（company-uuid, sale-order-uuid, dry-run），实现 PreRun 初始化逻辑（配置、日志、数据库、缓存、锁等） | Restrictions: 遵循 .cursor/rules/go-main.mdc，命令文件放在 main/command/ 目录 | Success: 命令框架创建成功，参数解析正确，PreRun 初始化完整

- [x] 1.2 实现参数验证和 dry-run 预览模式

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 实现参数验证逻辑和 dry-run 预览模式输出
  - Requirements: 4.3, 4.5
  - Leverage: Task 1.1 的命令框架，参考 `regenerate_sales_outbound.go:95-131`
  - Prompt: Role: Go Developer | Task: 实现参数验证（company-uuid 和 sale-order-uuid 必填）和 dry-run 预览模式 | Context: 验证参数不能为空，dry-run 模式下输出预览信息（将要执行的操作），不实际执行 | Restrictions: 使用彩色输出（blueColor, yellowColor, redColor, greenColor），参考 regenerate_sales_outbound.go 的输出格式 | Success: 参数验证正确，dry-run 预览模式工作正常

---

## Phase 2: 业务逻辑实现

- [x] 2.1 实现订单信息获取逻辑

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 根据订单 UUID 获取订单完整信息，包含商品、BOM、材料关联等
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Repository: `main/app/repository/order.go` (`GetSaleBillAllInfo` 方法)
  - Prompt: Role: Go Developer | Task: 实现订单信息获取逻辑，使用 OrderRepo.GetSaleBillAllInfo() 获取订单信息 | Context: 根据 sale-order-uuid 找到对应的 sale-bill-uuid，然后调用 GetSaleBillAllInfo()，该方法已预加载必要的 BOM 和材料关联数据 | Restrictions: 订单不存在时输出错误信息并退出，订单存在但 saleOrder 为 nil 时也输出错误 | Success: 订单信息获取成功，包含完整的商品和 BOM 数据

- [x] 2.2 实现材料用量计算逻辑

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 调用现有方法计算订单的材料用量
  - Requirements: 1.3, 1.4, 1.5
  - Leverage: 现有 Model 方法: `main/app/model/sale_order.go` (`GetValidSaleOrderProductMaterialList` 方法)
  - Prompt: Role: Go Developer | Task: 实现材料用量计算逻辑，调用 saleOrder.GetValidSaleOrderProductMaterialList() 方法 | Context: 该方法会自动计算有效售出商品的材料用量，支持成本卡和关联材料两种方式，返回 MaterialStock 列表 | Restrictions: 订单未完成（finish_time=0）时输出警告但允许继续执行 | Success: 材料用量计算成功，返回正确的 MaterialStock 列表

- [x] 2.3 实现删除旧记录逻辑

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 删除订单的旧材料记录（软删除）
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有 Repository: `main/app/repository/sale_order_material.go` (`DeleteSaleOrderMaterial` 方法)
  - Prompt: Role: Go Developer | Task: 实现删除旧材料记录逻辑，使用 SaleOrderMaterialRepo.DeleteSaleOrderMaterial() 方法 | Context: 传入 sale_bill_uuid，该方法会软删除该账单的所有材料记录（更新 delete_time 字段） | Restrictions: 使用事务确保原子性，记录删除的记录数 | Success: 旧记录删除成功，返回删除的记录数

- [x] 2.4 实现插入新记录逻辑

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 批量插入新计算的材料记录
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有 Repository: `main/app/repository/sale_order_material.go` (`BatchInsertSaleOrderMaterial` 方法)，参考 `main/app/event/order/order_checkout_event_handler.go:241-252`
  - Prompt: Role: Go Developer | Task: 实现批量插入新材料记录逻辑，构建 SaleOrderMaterial 对象列表并批量插入 | Context: 遍历 MaterialStock 列表，构建 SaleOrderMaterial 对象（包含 SaleOrderUuid, SaleBillUuid, MaterialUuid, WarehouseUuid, Num, StaffShiftLogUuid, CreateTime），使用 BatchInsertSaleOrderMaterial() 批量插入 | Restrictions: CreateTime 使用 saleOrder.FinishTime，Num 保留 4 位小数，使用事务确保原子性 | Success: 新记录插入成功，返回插入的记录数

- [x] 2.5 实现事务管理和用户确认机制

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 实现数据库事务管理和用户确认机制
  - Requirements: 2.5, 3.4, 4.3
  - Leverage: 现有工具: `main/app/repository/common.go` (`CommonRepo.Transaction` 方法)，参考 `regenerate_sales_outbound.go:133-141`
  - Prompt: Role: Go Developer | Task: 实现事务管理和用户确认机制，将删除和插入操作放在同一事务中，非 dry-run 模式下要求用户输入 'yes' 确认 | Context: 使用 repository.CommonRepo.Transaction() 包装删除和插入操作，非 dry-run 模式下使用 fmt.Scanln() 获取用户输入，输入非 'yes' 时取消操作 | Restrictions: 事务失败时回滚，用户取消时退出，操作成功时输出统计信息（删除记录数、新增记录数、耗时） | Success: 事务管理正确，用户确认机制工作正常

---

## Phase 3: 错误处理和日志

- [x] 3.1 实现错误处理逻辑

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 完善错误处理，包括订单不存在、订单未完成、数据库操作失败等场景
  - Requirements: 1.2, 1.4, 2.4, 3.4
  - Leverage: Task 2.1-2.5 的实现
  - Prompt: Role: Go Developer | Task: 完善错误处理逻辑，处理订单不存在、订单未完成、数据库操作失败等错误场景 | Context: 订单不存在时输出错误并退出，订单未完成时输出警告但继续，数据库操作失败时回滚事务并输出错误 | Restrictions: 使用 errors.WithMessage 包装错误，输出友好的错误信息 | Success: 所有错误场景处理正确，错误信息清晰

- [x] 3.2 实现日志输出和操作结果统计

  - File: `main/command/regenerate_order_material.go`
  - Purpose: 实现彩色日志输出和操作结果统计
  - Requirements: 4.5, 4.6
  - Leverage: 参考 `regenerate_sales_outbound.go:115-160` 的输出格式
  - Prompt: Role: Go Developer | Task: 实现彩色日志输出和操作结果统计，包括操作信息展示、成功/失败提示、统计信息（删除记录数、新增记录数、耗时） | Context: 使用 blueColor, greenColor, redColor, yellowColor 输出彩色日志，记录操作开始时间，计算耗时，输出统计信息 | Restrictions: 日志格式清晰，统计信息准确 | Success: 日志输出完整，统计信息准确

---

## Phase 4: 测试和文档

- [ ] 4.1 编写命令测试用例

  - File: `main/command/regenerate_order_material_test.go`
  - Purpose: 编写单元测试，确保命令功能正确
  - Requirements: 测试验收
  - Leverage: 现有测试: `main/command/` 目录下的测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 regenerate-order-material 命令编写单元测试，覆盖率 ≥ 70% | Context: 测试参数验证、订单信息获取、材料用量计算、数据库操作、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 更新命令使用文档

  - File: `docs/shared/api/` 或 README
  - Purpose: 更新命令使用说明文档
  - Requirements: 文档验收
  - Leverage: 现有文档: `docs/shared/api/` 或项目 README
  - Prompt: Role: Technical Writer | Task: 更新命令使用说明文档，包括命令格式、参数说明、使用示例、注意事项等 | Context: 参考 regenerate-sales-outbound 命令的文档格式 | Restrictions: 文档准确完整 | Success: 文档更新完成，使用说明清晰

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（Command 层 ≥ 70%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 命令使用文档已更新
- [ ] 代码注释完整（关键逻辑有中文注释）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/structs.mdc`
- [ ] 命令结构符合项目规范

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-regenerate-order-material/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-regenerate-order-material/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-regenerate-order-material/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-regenerate-order-material/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-regenerate-order-material/tasks.md)" | bc
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
**最后更新**: 2025-12-16  
**维护者**: xiezhihuan

