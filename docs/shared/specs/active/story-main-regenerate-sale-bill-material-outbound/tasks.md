# 重新生成销售账单材料出库记录 任务分解

> 本文档定义重新生成销售账单材料出库记录功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 13  
**已完成**: 10  
**进行中**: -  
**完成率**: 77%

---

## Phase 1: DTO 和响应结构

- [x] 1.1 创建响应结构

  - File: `main/app/dto/resp/sales_outbound_summary_resp.go`
  - Purpose: 定义 `RegenerateSaleBillMaterialOutboundResp` 响应结构
  - Requirements: 6.6
  - Leverage: 现有响应结构: `RegenerateOrderMaterialResp`, `RegenerateSalesOutboundSummaryResp`
  - Prompt: Role: Go Developer | Task: 在 sales_outbound_summary_resp.go 中新增 RegenerateSaleBillMaterialOutboundResp 结构体 | Context: 包含 DeletedCount（删除记录数）、InsertedCount（新增记录数）、DurationMs（执行耗时）字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段使用 json tag | Success: 响应结构定义完整，字段类型正确

---

## Phase 2: 服务接口实现

- [x] 2.1 在接口中新增方法定义

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 在 `ISalesOutboundSummarySrv` 接口中新增 `RegenerateSaleBillMaterialOutbound` 方法
  - Requirements: 6.1
  - Leverage: 现有接口方法: `RegenerateOrderMaterial`, `RegenerateSalesOutboundSummary`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 ISalesOutboundSummarySrv 接口中新增 RegenerateSaleBillMaterialOutbound 方法 | Context: 方法签名：`RegenerateSaleBillMaterialOutbound(ctx *gin.Context, companyUuid uint64, saleBillUuid uint64) (*resp.RegenerateSaleBillMaterialOutboundResp, error)` | Restrictions: 遵循 .cursor/rules/go-main.mdc，接口以 I 开头 | Success: 接口方法定义完整，方法签名正确

- [x] 2.2 实现查询原记录逻辑

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 查询销售账单的所有材料出库记录，并按 warehouse_out_form_uuid 分组
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有 Repository: `main/app/repository/warehouse_form.go` (`GetWarehouseOutFormItemBySaleBillUuid` 方法)
  - Prompt: Role: Go Developer | Task: 实现查询原记录逻辑，查询销售账单的材料出库记录并分组 | Context: 使用 WarehouseFormRepo.GetWarehouseOutFormItemBySaleBillUuid() 查询记录，过滤 material_uuid != 0 和 delete_time = 0，按 warehouse_out_form_uuid 分组 | Restrictions: 使用 map[uint64][]*model.WarehouseOutFormItem 存储分组结果 | Success: 查询逻辑正确，分组结果正确

- [x] 2.3 实现软删除原记录逻辑

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 软删除销售账单的旧材料出库记录
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有 Repository: `main/app/repository/warehouse_form.go`，参考 `RegenerateOrderMaterial` 的删除逻辑
  - Prompt: Role: Go Developer | Task: 实现软删除原记录逻辑，更新 delete_time 字段 | Context: 使用事务更新 `sale_bill_uuid = ? AND material_uuid != 0 AND delete_time = 0` 的记录，设置 delete_time 为当前时间戳，返回删除的记录数 | Restrictions: 使用 repository.CommonRepo.Transaction() 确保原子性，记录操作日志 | Success: 软删除逻辑正确，返回删除记录数

- [x] 2.4 实现重新计算材料消耗逻辑

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 根据订单当前的成本卡配置重新计算材料消耗
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6
  - Leverage: 现有方法: `RegenerateOrderMaterial()` 中的材料消耗计算逻辑，`OrderRepo.GetSaleBillAllInfo()`
  - Prompt: Role: Go Developer | Task: 实现重新计算材料消耗逻辑，复用 RegenerateOrderMaterial 的计算逻辑 | Context: 使用 OrderRepo.GetSaleBillAllInfo() 获取订单信息，遍历订单中的所有订单，调用 saleOrder.GetValidSaleOrderProductMaterialList() 计算材料消耗 | Restrictions: 仅统计有效售出的商品，材料消耗精度保留 4 位小数 | Success: 材料消耗计算正确，返回 MaterialStock 列表

- [x] 2.5 实现创建新记录并关联原出库单UUID逻辑

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 创建新的材料出库记录，并关联到原有的 warehouse_out_form_uuid
  - Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
  - Leverage: 现有 Repository: `main/app/repository/warehouse_form.go` (`CreateWarehouseOutFormItemRecords` 方法)，参考 `Model.NewWarehouseOutForm()` 的创建逻辑
  - Prompt: Role: Go Developer | Task: 实现创建新记录逻辑，按 warehouse_out_form_uuid 分组创建记录 | Context: 遍历分组后的原记录，为每个 warehouse_out_form_uuid 创建对应的新记录，验证出库单是否存在，继承原记录的关键字段（warehouse_out_form_uuid, warehouse_uuid, sale_bill_uuid, sale_order_uuid, staff_shift_log_uuid），设置正确的字段值（material_uuid, num, scene=0, status=1, reduce_stock=0） | Restrictions: 使用事务批量创建，如果原出库单不存在记录警告但允许继续执行，记录操作日志 | Success: 新记录创建成功，正确关联原出库单UUID

- [x] 2.6 实现分布式锁和事务管理

  - File: `main/app/service/sales_outbound_summary_service.go`
  - Purpose: 实现分布式锁防止并发操作，使用事务确保数据一致性
  - Requirements: 6.2, 6.3, 6.4, 6.5
  - Leverage: 现有工具: `lock.NewSystemLock()`，`repository.CommonRepo.Transaction()`
  - Prompt: Role: Go Developer | Task: 实现分布式锁和事务管理，防止并发操作并确保数据一致性 | Context: 使用 lock.NewSystemLock().TryLockUuidString() 获取锁，锁Key格式：`regenerate_sale_bill_material_outbound:{companyUuid}:{saleBillUuid}`，使用 defer 释放锁，将软删除和创建操作放在同一事务中 | Restrictions: 锁获取失败时返回错误，事务失败时回滚所有操作，记录操作日志 | Success: 分布式锁工作正常，事务管理正确

---

## Phase 3: 命令行工具实现

- [x] 3.1 创建命令文件框架

  - File: `main/command/regenerate_sale_bill_material_outbound.go`
  - Purpose: 创建命令行工具的基础框架，包括命令定义、参数解析和初始化逻辑
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有命令: `main/command/regenerate_order_material.go`
  - Prompt: Role: Go Developer specializing in CLI tools | Task: 创建 regenerate-sale-bill-material-outbound 命令的基础框架，参考 regenerate_order_material.go 的结构 | Context: 使用 Cobra 框架，定义命令名称、描述、参数（company-uuid, sale-bill-uuid, dry-run），实现 PreRun 初始化逻辑（配置、日志、数据库、缓存、锁等） | Restrictions: 遵循 .cursor/rules/go-main.mdc，命令文件放在 main/command/ 目录 | Success: 命令框架创建成功，参数解析正确，PreRun 初始化完整

- [x] 3.2 实现参数验证和 dry-run 预览模式

  - File: `main/command/regenerate_sale_bill_material_outbound.go`
  - Purpose: 实现参数验证逻辑和 dry-run 预览模式输出
  - Requirements: 5.2, 5.3, 5.7
  - Leverage: Task 3.1 的命令框架，参考 `regenerate_order_material.go:96-176`
  - Prompt: Role: Go Developer | Task: 实现参数验证（company-uuid 和 sale-bill-uuid 必填）和 dry-run 预览模式 | Context: 验证参数不能为空，dry-run 模式下输出预览信息（将要执行的操作：删除记录数、新增记录数），不实际执行 | Restrictions: 使用彩色输出（blueColor, yellowColor, redColor, greenColor），参考 regenerate_order_material.go 的输出格式 | Success: 参数验证正确，dry-run 预览模式工作正常

- [x] 3.3 实现用户确认机制和调用服务

  - File: `main/command/regenerate_sale_bill_material_outbound.go`
  - Purpose: 实现用户确认机制和调用服务方法
  - Requirements: 5.4, 5.5, 5.6
  - Leverage: Task 3.1, 3.2 的实现，参考 `regenerate_order_material.go:178-207`
  - Prompt: Role: Go Developer | Task: 实现用户确认机制和调用服务方法 | Context: 非 dry-run 模式下要求用户输入 'yes' 确认，调用 ISalesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound() 方法，输出操作结果（删除记录数、新增记录数、耗时） | Restrictions: 用户取消时退出，操作失败时输出错误信息，使用友好的输出格式 | Success: 用户确认机制工作正常，服务调用成功，输出格式友好

---

## Phase 4: 测试和文档

- [ ] 4.1 编写 Service 层单元测试

  - File: `main/app/service/sales_outbound_summary_service_test.go`
  - Purpose: 为 `RegenerateSaleBillMaterialOutbound` 方法编写单元测试
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/sales_outbound_summary_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 RegenerateSaleBillMaterialOutbound 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试正常流程、参数验证、分布式锁、事务回滚、错误处理等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 mock 数据 | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写命令行工具测试

  - File: `main/command/regenerate_sale_bill_material_outbound_test.go`
  - Purpose: 为命令行工具编写测试
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/command/regenerate_order_material_test.go`（如有）
  - Prompt: Role: QA Engineer | Task: 为命令行工具编写测试，覆盖所有参数组合 | Context: 测试必填参数缺失、无效UUID、dry-run模式、用户确认等场景 | Restrictions: 使用测试数据库，清理测试数据 | Success: 所有参数组合测试通过

- [ ] 4.3 编写集成测试

  - File: `main/tests/integration/regenerate_sale_bill_material_outbound_test.go`
  - Purpose: 编写端到端集成测试
  - Requirements: 测试要求
  - Leverage: 现有集成测试: `main/tests/integration/`
  - Prompt: Role: QA Engineer | Task: 编写端到端集成测试，测试完整的重新生成流程 | Context: 创建测试数据（销售账单、出库单、材料出库记录），执行重新生成操作，验证原记录已软删除，验证新记录已创建并关联原出库单UUID | Restrictions: 使用测试数据库，清理测试数据 | Success: 集成测试通过，验证逻辑正确

---

## 📝 任务执行顺序

1. **Phase 1**: DTO 和响应结构（1 个任务）
2. **Phase 2**: 服务接口实现（6 个任务）
3. **Phase 3**: 命令行工具实现（3 个任务）
4. **Phase 4**: 测试和文档（3 个任务）

**建议执行顺序**：
- 先完成 Phase 1（响应结构）
- 然后完成 Phase 2（服务接口实现）
- 接着完成 Phase 3（命令行工具）
- 最后完成 Phase 4（测试）

---

## 🔗 相关文档

- [需求文档](./requirements.md)
- [设计文档](./design.md)
- [销售出库单明细业务逻辑文档](../../api/warehouse-out-form-item-sales.md)

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: xiezhihuan  
**审核者**: {审核者}

