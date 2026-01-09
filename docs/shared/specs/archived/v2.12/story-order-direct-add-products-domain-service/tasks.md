# 订单直接添加商品领域服务 任务分解

> 本文档定义订单直接添加商品领域服务的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 领域服务接口和数据结构定义

- [ ] 1.1 创建领域服务接口文件

  - File: `main/app/modules/order/domain/service/i_order_direct_add_products_domain_service.go`
  - Purpose: 定义订单直接添加商品领域服务接口
  - Requirements: 1.1, 2.1
  - Leverage: 现有领域服务接口: `main/app/modules/order/domain/service/order_domain_service.go`，`main/app/modules/inventory/domain/service/warehouse_domain_service.go`
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 创建 IOrderDirectAddProductsDomainService 接口，定义 AddProductsToOrder 方法 | Context: 使用 pkg/context.Context，方法签名参考 requirements.md 和 design.md | Restrictions: 遵循 .cursor/rules/go-modules.mdc，接口以 I 开头 | Success: 接口定义完整，方法签名正确

- [ ] 1.2 定义 AddToOrderProduct 结构体和 ProductType 枚举

  - File: `main/app/modules/order/domain/service/i_order_direct_add_products_domain_service.go`
  - Purpose: 定义添加到订单的商品/实体数据结构
  - Requirements: 2.1
  - Leverage: 现有模型: `main/app/model/sale_order_product.go`，`main/app/model/sale_order_buffet_customer_type.go`，`main/app/model/sale_order_buffet_delay_product.go`
  - Prompt: Role: Go Developer | Task: 定义 AddToOrderProduct 结构体和 ProductType 枚举 | Context: 支持 Normal, Package, BuffetCustomer, BuffetDelay 四种类型，根据 Type 使用对应的字段 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 结构体定义完整，类型枚举正确

- [ ] 1.3 定义 AddToOrderOption 和选项函数

  - File: `main/app/modules/order/domain/service/i_order_direct_add_products_domain_service.go`
  - Purpose: 定义选项模式和选项函数
  - Requirements: 3.1
  - Leverage: 现有选项模式: `main/app/service/order_action.go:ActionAddOption`
  - Prompt: Role: Go Developer | Task: 定义 AddToOrderOption 结构体和选项函数（WithH5Product, WithMemberAdd, WithTableAdd, WithBuffetContext, WithBatchCooking） | Context: 使用函数式选项模式 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 选项结构体和函数定义完整

---

## Phase 2: 领域服务实现

- [ ] 2.1 创建领域服务实现文件

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 实现订单直接添加商品领域服务
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有领域服务实现: `main/app/modules/inventory/domain/service/warehouse_domain_service.go`，现有数据写入逻辑: `main/app/service/order.go:newSaleOrderProduct`，事务管理: `main/app/service/order_base.go`
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 实现 orderDirectAddProductsDomainService，包含 AddProductsToOrder 方法 | Context: 使用 repository.CommonRepo.Transaction 进行事务管理，根据 Type 调用不同的 persist 方法 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，不使用 panic，返回 error | Success: 实现完整，事务管理正确

- [ ] 2.2 实现 persistProduct 方法

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 实现普通商品/套餐的数据写入逻辑
  - Requirements: 1.6, 3.4, 3.5, 3.6
  - Leverage: 现有写入逻辑: `main/app/repository/sale_order_product.go:CreateSaleOrderProductAndBomAndAttribute`，`main/app/service/order.go:newSaleOrderProduct`
  - Prompt: Role: Go Developer | Task: 实现 persistProduct 方法，写入 sale_order_product、sale_order_product_bom、sale_order_product_attribute、sale_order_product_reason 表 | Context: 根据选项设置 is_accept_order 字段，写入商品及关联数据 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，确保外键关联正确 | Success: 方法实现完整，数据写入正确

- [ ] 2.3 实现 persistBuffetCustomer 方法

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 实现自助餐顾客的数据写入逻辑
  - Requirements: 1.6
  - Leverage: 现有写入逻辑: `main/app/repository/sale_order_buffet_customer_type.go:CreateSaleOrderBuffetCustomerTypeRecord`，`main/app/service/order_base.go`
  - Prompt: Role: Go Developer | Task: 实现 persistBuffetCustomer 方法，写入 sale_order_buffet_customer_type 表 | Context: 设置 sale_order_uuid，调用 Repository 创建记录 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 方法实现完整，数据写入正确

- [ ] 2.4 实现 persistBuffetDelay 方法

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 实现自助餐加钟的数据写入逻辑
  - Requirements: 1.6
  - Leverage: 现有写入逻辑: `main/app/repository/order.go:CreateSaleOrderBuffetDelayProduct`，`main/app/service/order_buffet.go`
  - Prompt: Role: Go Developer | Task: 实现 persistBuffetDelay 方法，写入 sale_order_buffet_delay_product 表 | Context: 设置 sale_order_uuid，调用 Repository 创建记录 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 方法实现完整，数据写入正确

- [ ] 2.5 实现 persistOperationRecord 方法

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 实现操作记录的数据写入逻辑
  - Requirements: 1.6
  - Leverage: 现有写入逻辑: `main/app/repository/sale_order_operation_record.go:CreateSaleOrderOperationRecord`
  - Prompt: Role: Go Developer | Task: 实现 persistOperationRecord 方法，写入 sale_order_operation_record 表 | Context: 从 context 获取操作来源、操作员等信息，记录添加操作 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 方法实现完整，操作记录正确

- [ ] 2.6 实现参数验证和类型检查

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service.go`
  - Purpose: 验证输入参数和类型一致性
  - Requirements: 1.7, 2.3
  - Leverage: 现有验证逻辑: `main/app/service/order_action.go:actionAdd`
  - Prompt: Role: Go Developer | Task: 在 AddProductsToOrder 方法中添加参数验证和类型检查 | Context: 验证 products 不为空，验证 Type 字段与对应数据字段的一致性 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，返回明确的错误信息 | Success: 验证逻辑完整，错误信息明确

---

## Phase 3: 单元测试

- [ ] 3.1 创建领域服务单元测试文件

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service_test.go`
  - Purpose: 编写领域服务单元测试
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/modules/inventory/domain/service/warehouse_domain_service_test.go`（如有）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 OrderDirectAddProductsDomainService 编写单元测试，覆盖率 ≥ 80% | Context: Mock Repository 层，测试正常场景、异常场景、事务回滚 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 3.2 测试正常场景（添加普通商品、套餐、自助餐顾客、自助餐加钟）

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service_test.go`
  - Purpose: 测试各种类型的商品/实体添加
  - Requirements: 1.1, 1.2, 2.1, 2.2, 2.3, 2.4
  - Leverage: Task 3.1 的测试文件
  - Success: 所有正常场景测试通过

- [ ] 3.3 测试批量添加场景（混合添加多种类型）

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service_test.go`
  - Purpose: 测试批量添加多种类型的商品/实体
  - Requirements: 1.3, 2.4
  - Leverage: Task 3.1 的测试文件
  - Success: 批量添加测试通过

- [ ] 3.4 测试异常场景（参数为空、类型不匹配、数据写入失败）

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service_test.go`
  - Purpose: 测试异常处理和错误返回
  - Requirements: 1.7, 2.3, 4.4
  - Leverage: Task 3.1 的测试文件
  - Success: 所有异常场景测试通过

- [ ] 3.5 测试事务回滚场景

  - File: `main/app/modules/order/domain/service/order_direct_add_products_domain_service_test.go`
  - Purpose: 测试数据写入失败时事务回滚
  - Requirements: 1.5, 4.1, 4.2
  - Leverage: Task 3.1 的测试文件
  - Success: 事务回滚测试通过

---

## Phase 4: 集成和适配

- [ ] 4.1 在应用服务中集成领域服务

  - File: `main/app/service/order.go` 或新建应用服务文件
  - Purpose: 在应用服务中调用领域服务
  - Requirements: 集成需求
  - Leverage: 现有应用服务: `main/app/modules/order/application/order_app_service.go`
  - Prompt: Role: Go Developer | Task: 在应用服务中集成 OrderDirectAddProductsDomainService | Context: 创建领域服务实例，在加购逻辑中调用 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 集成完成，调用正确

- [ ] 4.2 重构现有加购逻辑（可选，渐进式迁移）

  - File: `main/app/service/order_action.go:actionAdd`
  - Purpose: 重构现有加购逻辑，调用领域服务
  - Requirements: 向后兼容需求
  - Leverage: 现有逻辑: `main/app/service/order_action.go:actionAdd`，`main/app/service/order.go:newSaleOrderProduct`
  - Prompt: Role: Go Developer | Task: 重构 actionAdd 方法，在业务规则验证后调用领域服务 | Context: 保留业务规则验证逻辑，数据写入部分调用领域服务 | Restrictions: 保证向后兼容，保留旧方法作为 fallback | Success: 重构完成，向后兼容

- [ ] 4.3 验证向后兼容性

  - File: 所有调用加购逻辑的文件
  - Purpose: 确保重构不影响现有功能
  - Requirements: 向后兼容需求
  - Leverage: 现有测试用例
  - Success: 所有现有功能测试通过

---

## Phase 5: 集成测试和优化

- [ ] 5.1 编写集成测试

  - File: `test/integration/order_direct_add_products_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试从应用服务调用到数据库写入的完整流程，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.2 性能测试和优化

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Success: 数据写入操作 < 500ms，事务提交 < 100ms

- [ ] 5.3 代码审查和优化

  - File: 所有实现文件
  - Purpose: 代码质量检查和优化
  - Requirements: 代码质量要求
  - Leverage: 代码审查清单
  - Success: 代码通过审查，符合规范

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（领域服务层 ≥ 80%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 代码注释完整清晰
- [ ] design.md 与实现一致
- [ ] tasks.md 任务状态已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-modules.mdc`
- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-order-direct-add-products-domain-service/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-order-direct-add-products-domain-service/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-order-direct-add-products-domain-service/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-order-direct-add-products-domain-service/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-order-direct-add-products-domain-service/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的技术设计
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 领域服务开发

```
Role: Go Developer specializing in Domain Service

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Design: design.md 中的技术设计
- Project specs: 遵循 .cursor/rules/go-modules.mdc, .cursor/rules/go-main.mdc

Restrictions:
- 接口以 I 开头，实现以小写开头
- 所有方法第一个参数必须是 context.Context（pkg/context）
- 领域服务依赖 Repository 接口，不直接依赖数据库
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误
- 事务管理使用 repository.CommonRepo.Transaction

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80%

Test Cases Required:
- 正常场景测试（添加普通商品、套餐、自助餐顾客、自助餐加钟）
- 批量添加测试（混合添加多种类型）
- 异常场景测试（参数为空、类型不匹配、数据写入失败）
- 事务回滚测试

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- Mock Repository 层
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率 ≥ 80%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-19  
**维护者**: xiezhihuan

