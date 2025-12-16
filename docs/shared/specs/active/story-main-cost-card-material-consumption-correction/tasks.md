# 成本卡材料消耗修正 任务分解

> 本文档定义 成本卡材料消耗修正 功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 45  
**已完成**: 25  
**进行中**: -  
**完成率**: 55.6%

---

## Phase 1: DTO 和 Service 接口定义

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/cost_card_correction_req.go`
  - Purpose: 定义 API 请求参数结构体
  - Requirements: 1.1, 7.3
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 CostCardCorrectionReq 和 CostCardCorrectionPreviewReq 结构体 | Context: 包含 order_uuids 字段，使用 binding 标签验证 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [x] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/cost_card_correction_resp.go`
  - Purpose: 定义 API 响应数据结构体
  - Requirements: 1.1, 7.3
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 CostCardCorrectionPreviewResp, CostCardCorrectionResp, CostCardCorrectionLogsResp 等响应结构体 | Context: 包含订单信息、商品信息、材料信息、修正结果等 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

- [x] 1.3 创建 Service 接口

  - File: `main/app/service/i_cost_card_correction_service.go`
  - Purpose: 定义业务逻辑接口
  - Requirements: 所有功能需求
  - Leverage: 现有 Service 接口: `main/app/service/i_*_service.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 ICostCardCorrectionSrv 接口，定义 PreviewCorrection, ExecuteCorrection, GetCorrectionLogs 方法 | Context: 接口以 I 开头，方法签名包含 ctx 和 req 参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

---

## Phase 2: 订单选择与识别（Requirement 1）

- [x] 2.1 实现订单查询和识别逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 查询订单并识别使用成本卡的商品
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: `main/app/repository/sale_order_repo.go`, `main/app/model/sale_order.go`
  - Prompt: Role: Go Developer | Task: 实现查询订单列表和识别使用成本卡商品的逻辑 | Context: 查询 SaleOrder，识别 SaleOrderProduct 中使用 ProductBomCard 的商品 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 能够正确识别使用成本卡的商品

- [x] 2.2 实现预览修正影响逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 预览修正操作的影响范围
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: Task 2.1 的实现
  - Prompt: Role: Go Developer | Task: 实现 PreviewCorrection 方法，计算修正影响 | Context: 计算每个订单的材料退回数量、新消耗量、受影响的日期范围 | Restrictions: 不实际执行修正，只计算影响 | Success: 预览结果准确，包含所有必要信息

---

## Phase 3: 材料退回处理（Requirement 2）

- [x] 3.1 实现查询历史出库记录逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 查询订单的历史出库记录
  - Requirements: 2.1, 2.2
  - Leverage: `main/app/repository/warehouse_form_repo.go`, `main/app/model/warehouse_form.go`
  - Prompt: Role: Go Developer | Task: 实现查询订单历史出库记录的逻辑 | Context: 查询 WarehouseOutFormItem，按订单UUID和材料UUID汇总 | Restrictions: 只查询已出库的记录（status=1, reduce_stock=1） | Success: 能够正确查询历史出库记录

- [x] 3.2 实现材料退回逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 执行材料退回操作，增加材料库存
  - Requirements: 2.3, 2.4
  - Leverage: `main/app/repository/warehouse_item_repo.go`, `main/app/service/purchase_order/helper.go` (参考材料库存操作)
  - Prompt: Role: Go Developer | Task: 实现 returnMaterials 方法，退回错误扣减的材料 | Context: 增加 WarehouseItem.Stock，记录 WarehouseInOutLog，更新 RelatedMaterial | Restrictions: 使用事务，确保数据一致性 | Success: 材料退回成功，库存正确增加

- [x] 3.3 实现商品库存重新计算逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 根据修正后的材料库存重新计算商品库存
  - Requirements: 2.6
  - Leverage: `main/app/modules/inventory/domain/service/bom_card_product_inventory_strategy.go`, `main/app/model/product.go` (CalculateExpectedProductionNum)
  - Prompt: Role: Go Developer | Task: 实现 recalculateProductInventory 方法，重新计算商品库存 | Context: 根据成本卡计算：材料库存/材料用量（取最小值） | Restrictions: 更新所有使用该材料的成本卡关联的商品库存 | Success: 商品库存重新计算正确

---

## Phase 4: 重新计算材料消耗（Requirement 3）

- [x] 4.1 实现删除旧材料消耗记录逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 删除订单的旧材料消耗记录
  - Requirements: 3.4
  - Leverage: `main/app/repository/sale_order_product_bom_repo.go`
  - Prompt: Role: Go Developer | Task: 实现删除旧材料消耗记录的逻辑 | Context: 删除 SaleOrderProductBom 中关联该订单的记录 | Restrictions: 使用软删除或物理删除（根据业务需求） | Success: 旧记录删除成功

- [x] 4.2 实现重新计算材料消耗逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 根据正确的成本卡重新计算材料消耗量
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: `main/app/model/sale_order.go` (flavorUseCard), `main/app/model/product.go` (ProductBomCard)
  - Prompt: Role: Go Developer | Task: 实现 recalculateMaterialConsumption 方法，重新计算材料消耗 | Context: 获取订单商品当前关联的成本卡，使用 flavorUseCard 逻辑计算消耗量，考虑商品数量和成本卡加工份数 | Restrictions: 验证成本卡有效性，材料是否存在 | Success: 材料消耗重新计算正确，生成新的消耗记录

---

## Phase 5: 重新生成出库记录（Requirement 4）

- [x] 5.1 实现创建出库单逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 创建新的出库单
  - Requirements: 4.1
  - Leverage: `main/app/repository/warehouse_form_repo.go`, `main/app/service/order.go` (参考出库单创建)
  - Prompt: Role: Go Developer | Task: 实现创建 WarehouseOutForm 的逻辑 | Context: 生成出库单号，设置场景为销售出库 | Restrictions: 遵循现有出库单创建规范 | Success: 出库单创建成功

- [x] 5.2 实现创建出库单明细和扣减库存逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 创建出库单明细并扣减材料库存
  - Requirements: 4.2, 4.3, 4.4, 4.5
  - Leverage: `main/app/repository/warehouse_form_repo.go`, `main/app/repository/warehouse_item_repo.go`, `main/app/service/purchase_order/helper.go`
  - Prompt: Role: Go Developer | Task: 实现创建 WarehouseOutFormItem 和扣减材料库存的逻辑 | Context: 按材料汇总出库数量，创建出库单明细，扣减 WarehouseItem.Stock，记录 WarehouseInOutLog | Restrictions: 使用事务，确保数据一致性 | Success: 出库单明细创建成功，库存正确扣减

- [x] 5.3 实现更新关联材料库存逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 更新规格/加料关联材料库存
  - Requirements: 4.5
  - Leverage: `main/app/repository/material_repo.go` (UpdateRelatedMaterialStock)
  - Prompt: Role: Go Developer | Task: 实现更新 RelatedMaterial 库存的逻辑 | Context: 调用 MaterialRepo.UpdateRelatedMaterialStock | Restrictions: 更新所有相关材料的关联库存 | Success: 关联材料库存更新成功

- [x] 5.4 实现重新计算商品库存逻辑（出库后）

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 出库后重新计算商品库存
  - Requirements: 4.6
  - Leverage: Task 3.3 的实现
  - Prompt: Role: Go Developer | Task: 在出库后重新计算商品库存 | Context: 调用 recalculateProductInventory 方法 | Restrictions: 确保商品库存准确 | Success: 商品库存重新计算正确

---

## Phase 6: ERP 数据同步（Requirement 5）

- [x] 6.1 实现重新生成 POS Invoice 数据逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 重新生成 POS Invoice 数据
  - Requirements: 5.1
  - Leverage: `main/app/service/order_pay.go` (SavePosInvoice), `main/app/service/rpc/erp/selling.go`
  - Prompt: Role: Go Developer | Task: 实现重新生成 POS Invoice 数据的逻辑 | Context: 包含订单商品和材料消耗，参考 order_pay.go 中的 SavePosInvoice 逻辑 | Restrictions: 使用正确的成本卡数据 | Success: POS Invoice 数据生成正确

- [x] 6.2 实现调用 ERP 接口同步数据逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 调用 ERP 接口保存 POS Invoice
  - Requirements: 5.2, 5.3
  - Leverage: `main/app/service/rpc/erp/selling.go` (SavePosInvoice), `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Prompt: Role: Go Developer | Task: 实现 resyncErpData 方法，调用 ERP 接口同步数据 | Context: 调用 ErpSellingService.SavePosInvoice，处理返回结果，更新订单的 ERP invoice 名称 | Restrictions: 处理 ERP 返回的错误，如需要先删除旧数据 | Success: ERP 数据同步成功

- [x] 6.3 实现 ERP 同步错误处理和重试逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 处理 ERP 同步失败的情况
  - Requirements: 5.3, 5.5
  - Leverage: Task 6.2 的实现
  - Prompt: Role: Go Developer | Task: 实现 ERP 同步错误处理和重试逻辑 | Context: 记录同步日志，支持重试机制 | Restrictions: 错误不中断整个修正流程，记录到失败列表 | Success: 错误处理正确，支持重试

---

## Phase 7: 每日销售出库修正（Requirement 6）

- [x] 7.1 实现识别受影响日期范围逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 识别受影响的日期范围
  - Requirements: 6.1
  - Leverage: `main/app/model/sale_order.go` (SaleOrder)
  - Prompt: Role: Go Developer | Task: 实现识别受影响日期范围的逻辑 | Context: 根据订单的营业日期（business_date）识别受影响的日期 | Restrictions: 去重，排序 | Success: 日期范围识别正确

- [x] 7.2 实现重新统计每日销售出库记录逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 重新统计每日销售出库记录
  - Requirements: 6.2, 6.3
  - Leverage: `admin/app/common/model/product/Product.php` (salesOutInventoryRecord), `admin/app/common/model/erp/ErpInventoryRecord.php`
  - Prompt: Role: Go Developer | Task: 实现重新统计每日销售出库记录的逻辑 | Context: 根据修正后的订单数据，重新统计每日的销售出库记录 | Restrictions: 更新或重新生成 ErpInventoryRecord | Success: 每日销售出库记录修正正确

- [x] 7.3 实现批量修正多天数据逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 支持批量修正多天的每日销售出库记录
  - Requirements: 6.4, 6.5
  - Leverage: Task 7.2 的实现
  - Prompt: Role: Go Developer | Task: 实现批量修正多天数据的逻辑 | Context: 遍历受影响的日期，逐个修正每日销售出库记录 | Restrictions: 使用事务，确保数据一致性 | Success: 多天数据批量修正成功

---

## Phase 8: 操作日志与审计（Requirement 7）

- [x] 8.1 设计修正日志表结构（可选）

  - File: `main/app/model/cost_card_correction_log.go`
  - Purpose: 定义修正日志表结构
  - Requirements: 7.1, 7.2
  - Leverage: 现有日志表: `main/app/model/purchase_order.go` (PurchaseOrderLog)
  - Prompt: Role: Database Engineer | Task: 设计修正日志表结构（如需要） | Context: 记录修正操作UUID、订单UUID、操作时间、操作人、状态等信息 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 表结构设计合理

- [x] 8.2 实现记录修正日志逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 记录修正操作的详细日志
  - Requirements: 7.1, 7.2
  - Leverage: 现有日志表或新创建的日志表
  - Prompt: Role: Go Developer | Task: 实现 recordCorrectionLog 方法，记录修正日志 | Context: 记录操作时间、操作人、订单列表、材料列表、数量、状态等信息 | Restrictions: 记录每一步操作的执行结果 | Success: 日志记录完整准确

- [x] 8.3 实现操作日志查询逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 提供操作日志查询功能
  - Requirements: 7.3
  - Leverage: Task 8.2 的实现
  - Prompt: Role: Go Developer | Task: 实现 GetCorrectionLogs 方法，查询修正日志 | Context: 支持按修正UUID、订单UUID查询，支持分页 | Restrictions: 遵循分页规范 | Success: 日志查询功能正常

- [ ] 8.4 实现操作回滚功能（可选）

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 支持操作回滚
  - Requirements: 7.4, 7.5
  - Leverage: Task 8.2 的日志记录
  - Prompt: Role: Go Developer | Task: 实现操作回滚功能（如需要） | Context: 根据日志记录，反向执行修正操作 | Restrictions: 使用事务，确保回滚成功 | Success: 回滚功能正常

---

## Phase 9: API 层实现

- [x] 9.1 创建 API Controller

  - File: `main/app/api/cost_card_correction_api.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有 API: `main/app/api/*_api.go`, Phase 1-8 的 Service 实现
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 CostCardCorrectionAPI，实现 Preview, Execute, GetLogs 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 9.2 注册 API 路由

  - File: `main/router/router.go`, `main/app/api/v1/shop/shop_cost_card_correction.go`
  - Purpose: 注册 API 路由
  - Requirements: 所有功能需求
  - Leverage: 现有路由: `main/router/router.go`
  - Prompt: Role: Go Developer | Task: 注册成本卡修正相关的 API 路由 | Context: POST /api/v1/shop/cost_card_correction/preview, POST /api/v1/shop/cost_card_correction/execute, GET /api/v1/shop/cost_card_correction/logs | Restrictions: 遵循路由注册规范 | Success: 路由注册成功

- [ ] 9.3 注册 Service 依赖

  - File: `main/app/service/service.go` 或依赖注入文件
  - Purpose: 注册 Service 依赖
  - Requirements: 所有功能需求
  - Leverage: 现有 Service 注册: `main/app/service/service.go`
  - Prompt: Role: Go Developer | Task: 注册 CostCardCorrectionService 及其依赖 | Context: 依赖 MaterialService, OrderService, WarehouseService, ErpSellingService, InventoryService | Restrictions: 遵循依赖注入规范 | Success: Service 注册成功

---

## Phase 10: 测试

- [ ] 10.1 编写 Service 单元测试

  - File: `main/app/service/cost_card_correction_service_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CostCardCorrectionService 编写单元测试，覆盖率 ≥ 70% | Context: 测试材料退回、重新计算、出库记录生成、ERP同步等业务逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 10.2 编写 Repository 单元测试（如需要）

  - File: `main/app/repository/{name}_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 如创建新 Repository
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 编写 Repository 单元测试，覆盖率 ≥ 80% | Context: 测试 CRUD 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 10.3 编写 API 集成测试

  - File: `main/app/api/cost_card_correction_api_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 CostCardCorrectionAPI 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 10.4 编写端到端集成测试

  - File: `test/integration/cost_card_correction_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整修正流程：订单选择 → 材料退回 → 重新计算 → 出库记录 → ERP 同步 → 每日销售出库修正 | Restrictions: 测试真实用户场景，测试数据一致性 | Success: 集成测试通过

- [ ] 10.5 编写并发测试

  - File: `test/integration/cost_card_correction_concurrent_test.go`
  - Purpose: 测试并发修正场景
  - Requirements: 可靠性要求
  - Leverage: 现有并发测试
  - Prompt: Role: QA Engineer | Task: 编写并发修正测试 | Context: 测试多个订单并发修正，测试 UUID 锁机制 | Restrictions: 确保数据一致性 | Success: 并发测试通过

---

## Phase 11: 性能优化和错误处理

- [ ] 11.1 实现分批处理逻辑

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 分批处理大批量订单
  - Requirements: 性能要求
  - Leverage: Task 9.1 的实现
  - Prompt: Role: Go Developer | Task: 实现分批处理逻辑，每批 100 个订单 | Context: 避免一次性处理过多数据，提高性能 | Restrictions: 确保每批处理的事务独立性 | Success: 分批处理正常，性能达标

- [ ] 11.2 实现 UUID 锁机制

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 防止并发修正同一订单
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Prompt: Role: Go Developer | Task: 实现 UUID 锁机制，防止并发修正 | Context: 使用系统锁锁定订单UUID，修正完成后释放 | Restrictions: 确保锁的正确获取和释放 | Success: 并发场景测试通过

- [ ] 11.3 优化数据库查询

  - File: `main/app/repository/*_repo.go`
  - Purpose: 优化 SQL 查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，添加索引
  - Prompt: Role: Database Engineer | Task: 优化数据库查询，添加必要索引 | Context: 为 order_uuid, material_uuid 等字段添加索引 | Restrictions: 遵循数据库优化规范 | Success: 查询时间 < 50ms

- [ ] 11.4 实现错误回滚机制

  - File: `main/app/service/cost_card_correction_service.go`
  - Purpose: 确保修正失败时能够回滚
  - Requirements: 可靠性要求
  - Leverage: 数据库事务
  - Prompt: Role: Go Developer | Task: 完善错误回滚机制 | Context: 使用事务确保数据一致性，任何步骤失败自动回滚 | Restrictions: 确保回滚完整 | Success: 错误回滚测试通过

---

## Phase 12: 文档和代码审查

- [ ] 12.1 更新 API 文档

  - File: `docs/shared/api/cost_card_correction_api.md`
  - Purpose: 确保 API 文档完整
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新成本卡修正 API 文档 | Context: 包含所有 API 接口的说明、参数、响应格式 | Restrictions: 文档准确完整 | Success: API 文档已更新

- [ ] 12.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Prompt: Role: Technical Writer | Task: 更新 CHANGELOG，记录成本卡材料消耗修正功能 | Context: 按照版本管理规范记录 | Restrictions: 遵循 .cursor/rules/version.mdc | Success: CHANGELOG 已更新

- [ ] 12.3 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量
  - Requirements: 代码质量要求
  - Leverage: 代码审查清单
  - Prompt: Role: Code Reviewer | Task: 审查所有代码修改 | Context: 检查代码规范、错误处理、测试覆盖率 | Restrictions: 遵循所有开发规范 | Success: 代码审查通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - 材料出库相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 技术文档完整

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-cost-card-material-consumption-correction/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-cost-card-material-consumption-correction/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-cost-card-material-consumption-correction/tasks.md
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
**最后更新**: 2025-12-12  
**维护者**: 后端开发组

