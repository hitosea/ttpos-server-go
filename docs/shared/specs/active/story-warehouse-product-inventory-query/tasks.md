# 仓库模块商品库存查询功能 任务分解

> 本文档定义 仓库模块商品库存查询功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 15  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: Domain Layer（领域层）

### Repository 接口

- [x] 1.1 创建 ProductBomRepository 接口

  - File: `main/app/modules/inventory/domain/repository/product_bom_repository.go`
  - Purpose: 定义商品BOM仓储接口
  - Requirements: 1.1, 2.3
  - Leverage: 现有 Repository 接口: `main/app/modules/inventory/domain/repository/warehouse_repository.go`
  - Prompt: Role: Go Developer specializing in DDD Repository Pattern | Task: 创建 IProductBomRepository 接口，定义 FindByUuid 和 FindByProductPackageUuid 方法 | Context: 使用 pkg/context.Context，返回 *entity.ProductBom（但实际使用 model.ProductBom） | Restrictions: 遵循 .cursor/rules/go-modules.mdc，接口定义在 domain/repository/ | Success: 接口定义完整，方法签名正确

### Strategy 接口和实现

- [x] 1.2 创建 InventoryStrategy 接口

  - File: `main/app/modules/inventory/domain/service/inventory_strategy.go`
  - Purpose: 定义库存计算策略接口
  - Requirements: 1.2, 2.1, 3.1
  - Leverage: 策略模式设计，参考现有领域服务: `main/app/modules/inventory/domain/service/warehouse_item_domain_service.go`
  - Prompt: Role: Go Developer specializing in Strategy Pattern | Task: 创建 IInventoryStrategy 接口，定义 CalculateInventory 方法 | Context: 使用 pkg/context.Context，接收 productBom interface{}（使用现有 Model） | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 接口定义完整，方法签名正确

- [x] 1.3 实现 BomCardProductInventoryStrategy（有成本卡商品库存计算策略）

  - File: `main/app/modules/inventory/domain/service/bom_card_product_inventory_strategy.go`
  - Purpose: 实现有成本卡商品的库存计算逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.4, 4.1, 4.2, 4.3
  - Leverage: 
    - 现有方法: `main/app/model/product.go:902` - `ProductBomCard.CalculateExpectedProductionNum()`
    - 现有常量: `main/app/constant/product.go:22` - `ProductBomInfiniteStock`
    - 现有字段: `ProductBom.UseBomCardStock`, `ProductBom.IsSoldOut`, `ProductBom.SellableQuantity`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 BomCardProductInventoryStrategy，计算有成本卡商品的库存 | Context: 判断 UseBomCardStock，开启则调用 ProductBomCard.CalculateExpectedProductionNum()，未开启则执行无成本卡逻辑 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context，不使用 panic，返回 error | Success: 策略实现完整，逻辑正确，处理所有边界情况

- [x] 1.4 实现 NonBomCardProductInventoryStrategy（无成本卡商品库存计算策略）

  - File: `main/app/modules/inventory/domain/service/non_bom_card_product_inventory_strategy.go`
  - Purpose: 实现无成本卡商品的库存计算逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 
    - 现有字段: `ProductBom.IsSoldOut`, `ProductBom.SellableQuantity`
    - 现有常量: `main/app/constant/product.go:22` - `ProductBomInfiniteStock`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 NonBomCardProductInventoryStrategy，计算无成本卡商品的库存 | Context: 判断 IsSoldOut，售罄返回0，否则判断 SellableQuantity，设置则返回该值，否则返回 ProductBomInfiniteStock | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context，不使用 panic，返回 error | Success: 策略实现完整，逻辑正确

### Domain Service 实现

- [x] 1.5 实现 ProductInventoryDomainService（商品库存领域服务）

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Purpose: 实现统一的商品库存查询接口
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 
    - 现有领域服务: `main/app/modules/inventory/domain/service/warehouse_item_domain_service.go`
    - Task 1.1 的 Repository 接口
    - Task 1.2-1.4 的策略实现
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 实现 ProductInventoryDomainService，提供统一的 GetProductInventory 方法 | Context: 查询商品BOM，判断是否有成本卡，选择对应策略，计算库存 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context，不使用 panic，返回 error，使用 errors.WithMessage 包装错误 | Success: Service 实现完整，逻辑正确，错误处理完善

---

## Phase 2: Infrastructure Layer（基础设施层）

### Repository 实现

- [x] 2.1 实现 ProductBomRepository（适配现有 Model）

  - File: `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl.go`
  - Purpose: 实现商品BOM仓储，适配现有 Model 层
  - Requirements: 1.1
  - Leverage: 
    - 现有 Repository 实现: `main/app/modules/inventory/infrastructure/persistence/warehouse_repository_impl.go`
    - 现有 Model: `main/app/model/product.go` - `ProductBom`
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 ProductBomRepository，适配现有 ProductBom Model | Context: 使用 GORM 查询，Preload 关联数据（ProductBomCard.RelatedMaterials.Material），使用 pkg/context.Context.GetDB() 获取数据库连接 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，Repository 只持有 db 实例，软删除(delete_time=0) | Success: Repository 实现完整，查询正确，关联数据预加载正确

---

## Phase 3: 单元测试

### Domain Service 测试

- [x] 3.1 编写 ProductInventoryDomainService 单元测试

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service_test.go`
  - Purpose: 确保领域服务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/modules/inventory/domain/service/warehouse_domain_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 ProductInventoryDomainService 编写单元测试，覆盖率 ≥ 90% | Context: 测试有成本卡商品（成本卡控制开启/未开启）、无成本卡商品、商品不存在、成本卡不存在等场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 测试覆盖率 ≥ 90%，所有测试通过

### Strategy 测试

- [x] 3.2 编写 BomCardProductInventoryStrategy 单元测试

  - File: `main/app/modules/inventory/domain/service/bom_card_product_inventory_strategy_test.go`
  - Purpose: 确保有成本卡商品库存计算逻辑正确
  - Requirements: 2.1, 2.2, 2.3, 2.4, 4.1, 4.2, 4.3
  - Leverage: Task 3.1 的测试结构
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 BomCardProductInventoryStrategy 编写单元测试，覆盖率 100% | Context: 测试成本卡控制开启（材料库存充足/不足/为0）、成本卡控制未开启（售罄/可售量/无限库存）等场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 测试覆盖率 100%，所有测试通过

- [x] 3.3 编写 NonBomCardProductInventoryStrategy 单元测试

  - File: `main/app/modules/inventory/domain/service/non_bom_card_product_inventory_strategy_test.go`
  - Purpose: 确保无成本卡商品库存计算逻辑正确
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 3.1 的测试结构
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 NonBomCardProductInventoryStrategy 编写单元测试，覆盖率 100% | Context: 测试售罄、可售量设置、无限库存等场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 测试覆盖率 100%，所有测试通过

### Repository 测试

- [x] 3.4 编写 ProductBomRepository 单元测试

  - File: `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 1.1
  - Leverage: 现有测试: `main/app/modules/inventory/infrastructure/persistence/warehouse_repository_impl_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 ProductBomRepository 编写单元测试，覆盖率 ≥ 80% | Context: 测试 FindByUuid（存在/不存在）、关联数据预加载等场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 4: 集成和优化

### 集成测试

- [x] 4.1 编写集成测试

  - File: `test/integration/product_inventory_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整的库存查询流程，包括有成本卡和无成本卡商品的各种场景 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

### 缓存优化

- [x] 4.2 实现 Redis 缓存

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Purpose: 对库存查询结果进行缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/app/service/`（参考其他 Service 的缓存实现）
  - Prompt: Role: Go Developer with Redis expertise | Task: 在 ProductInventoryDomainService 中实现 Redis 缓存 | Context: Key 格式: `ttpos:inventory:product:{product_bom_uuid}`，过期时间 5 分钟，使用 Cache-Aside Pattern | Restrictions: 遵循 .cursor/rules/go-modules.mdc | Success: 缓存实现完成，命中率 > 80%

### 性能优化

- [x] 4.3 数据库查询优化

  - File: `main/app/modules/inventory/infrastructure/persistence/product_bom_repository_impl.go`
  - Purpose: 优化 SQL 查询性能
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，现有优化经验
  - Prompt: Role: Database Engineer | Task: 优化 ProductBomRepository 的查询性能 | Context: 使用 Preload 预加载关联数据，添加必要的索引，避免 N+1 查询 | Restrictions: 查询时间 < 50ms | Success: 查询时间 < 50ms

- [x] 4.4 添加数据库索引（如需要）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_index_to_product_bom.sql`
  - Purpose: 为商品BOM表添加索引，提升查询性能
  - Requirements: 性能要求
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 为 ttpos_product_bom 表添加索引 | Context: 为 product_bom_card_uuid, uuid 字段添加索引（如不存在） | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 索引创建成功，查询性能提升

---

## Phase 5: 文档和代码审查

### 文档更新

- [x] 5.1 更新库存模块 README

  - File: `main/app/modules/inventory/README.md`
  - Purpose: 更新文档，说明新增的商品库存查询功能
  - Requirements: 文档要求
  - Leverage: 现有 README 结构
  - Prompt: Role: Technical Writer | Task: 更新库存模块 README，添加 ProductInventoryDomainService 说明 | Context: 说明接口、使用示例、业务规则 | Restrictions: 文档准确完整 | Success: 文档已更新

- [x] 5.2 更新设计文档实现细节

  - File: `docs/shared/specs/active/story-warehouse-product-inventory-query/design.md`
  - Purpose: 记录实现过程中的关键决策和变更
  - Requirements: 文档要求
  - Leverage: 设计文档模板
  - Success: 设计文档已更新

### 代码审查

- [x] 5.3 代码审查和重构

  - File: 所有新增文件
  - Purpose: 确保代码质量，符合规范
  - Requirements: 代码质量要求
  - Leverage: `.cursor/rules/go-main.mdc`, `.cursor/rules/go-modules.mdc`
  - Command: `go fmt ./... && go vet ./... && go test ./...`
  - Success: 代码通过格式化和静态检查，测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Domain Service: ≥ 90%
  - Strategy: 100%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 库存模块 README 已更新
- [ ] design.md 已更新实现细节
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/go-modules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`（如提供 API）

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-warehouse-product-inventory-query/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-warehouse-product-inventory-query/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-warehouse-product-inventory-query/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-warehouse-product-inventory-query/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-warehouse-product-inventory-query/tasks.md)" | bc
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

### Go 领域服务开发

```
Role: Go Developer specializing in DDD Domain Service

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-modules.mdc, .cursor/rules/go-main.mdc

Restrictions:
- 使用 pkg/context.Context，不使用标准库 context.Context
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖 Repository 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率达标
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-10  
**维护者**: 后端开发组

