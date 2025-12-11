# 仓库模块商品包库存查询功能 任务分解

> 本文档定义 仓库模块商品包库存查询功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 4  
**进行中**: -  
**完成率**: 50%

---

## Phase 1: Domain Layer（领域层）

### Domain Service 扩展

- [x] 1.1 扩展 IProductInventoryDomainService 接口，新增 GetProductPackageInventory 方法

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Purpose: 定义商品包库存查询接口
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有接口: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Prompt: Role: Go Developer specializing in DDD Domain Service | Task: 在 IProductInventoryDomainService 接口中新增 GetProductPackageInventory 方法 | Context: 方法签名: GetProductPackageInventory(ctx context.Context, productPackageUuid uint64) (float64, error) | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context | Success: 接口方法定义完整，方法签名正确

- [x] 1.2 实现 GetProductPackageInventory 方法，实现最小值计算逻辑

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Purpose: 实现商品包库存计算逻辑，返回所有BOM库存的最小值
  - Requirements: 1.1, 1.2, 1.3, 3.1, 3.2, 3.3
  - Leverage: 
    - 现有方法: `ProductBomRepository.FindByProductPackageUuid()` - 查询商品包下所有BOM
    - 现有方法: `GetProductInventory()` - 查询单个BOM库存
    - 标准库: `math.MaxFloat64`, `math.Min` - 最小值计算
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetProductPackageInventory 方法，计算商品包下所有BOM库存的最小值 | Context: 1) 调用 FindByProductPackageUuid 查询BOM列表 2) 遍历每个BOM，调用 GetProductInventory 获取库存 3) 使用 math.Min 或循环比较获取最小值 4) 处理边界情况（没有BOM、部分BOM查询失败） | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context，不使用 panic，返回 error，使用 errors.WithMessage 包装错误 | Success: 方法实现完整，最小值计算正确，边界情况处理完善

---

## Phase 2: Application Layer（应用层）

### Application Service 扩展

- [x] 2.1 扩展 ProductInventoryAppService，新增 GetProductPackageInventory 方法（带缓存）

  - File: `main/app/modules/inventory/application/product_inventory_app_service.go`
  - Purpose: 实现带缓存的商品包库存查询方法
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 
    - 现有方法: `GetProductInventory()` - 带缓存的BOM库存查询方法（参考实现）
    - 现有常量: `ProductInventoryCacheKeyPrefix`, `ProductInventoryCacheTTL` - 缓存键前缀和过期时间
    - 领域服务: `domainService.GetProductPackageInventory()` - Task 1.2 的实现
  - Prompt: Role: Go Developer specializing in Application Service with caching | Task: 在 ProductInventoryAppService 中新增 GetProductPackageInventory 方法，实现缓存逻辑 | Context: 1) 定义缓存键: product_package_inventory:{company_uuid}:{product_package_uuid} 2) 缓存过期时间: 5分钟 3) 优先从缓存读取，缓存未命中时调用领域服务计算 4) 计算完成后写入缓存 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context，缓存键格式与BOM库存缓存保持一致 | Success: 方法实现完整，缓存逻辑正确，缓存键格式正确

- [x] 2.2 实现 InvalidateProductPackageInventoryCache 方法

  - File: `main/app/modules/inventory/application/product_inventory_app_service.go`
  - Purpose: 实现商品包库存缓存失效方法
  - Requirements: 2.1, 2.4
  - Leverage: 
    - 现有方法: `InvalidateProductInventoryCache()` - BOM库存缓存失效方法（参考实现）
    - 缓存键格式: `ProductPackageInventoryCacheKeyPrefix` - Task 2.1 定义的常量
  - Prompt: Role: Go Developer specializing in Application Service | Task: 实现 InvalidateProductPackageInventoryCache 方法，使商品包库存缓存失效 | Context: 1) 构建缓存键: product_package_inventory:{company_uuid}:{product_package_uuid} 2) 调用 cache.Del() 删除缓存 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 pkg/context.Context | Success: 方法实现完整，缓存失效逻辑正确

---

## Phase 3: 单元测试

### Domain Service 测试

- [ ] 3.1 编写 GetProductPackageInventory 单元测试

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service_test.go`
  - Purpose: 确保商品包库存计算逻辑正确
  - Requirements: 1.1, 1.2, 1.3, 3.1, 3.2, 3.3
  - Leverage: 现有测试: `main/app/modules/inventory/domain/service/product_inventory_domain_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetProductPackageInventory 编写单元测试，覆盖率 ≥ 90% | Context: 测试场景: 1) 商品包下多个BOM的最小值计算 2) 商品包下单个BOM的库存计算 3) 商品包下没有BOM的边界情况 4) 部分BOM查询失败的场景 5) 所有BOM查询失败的场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 Mock Repository | Success: 测试覆盖率 ≥ 90%，所有测试通过，边界情况已覆盖

### Application Service 测试

- [ ] 3.2 编写 GetProductPackageInventory 单元测试（带缓存）

  - File: `main/app/modules/inventory/application/product_inventory_app_service_test.go`
  - Purpose: 确保带缓存的商品包库存查询逻辑正确
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有测试: `main/app/modules/inventory/application/product_inventory_app_service_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetProductPackageInventory 编写单元测试，测试缓存机制 | Context: 测试场景: 1) 缓存命中场景 2) 缓存未命中场景 3) 缓存写入场景 4) 缓存失效场景 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 Mock Cache 和 Mock Domain Service | Success: 测试覆盖率 ≥ 80%，所有测试通过，缓存机制测试完整

- [ ] 3.3 编写 InvalidateProductPackageInventoryCache 单元测试

  - File: `main/app/modules/inventory/application/product_inventory_app_service_test.go`
  - Purpose: 确保缓存失效方法正确
  - Requirements: 2.1, 2.4
  - Leverage: Task 3.2 的测试结构
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 InvalidateProductPackageInventoryCache 编写单元测试 | Context: 测试缓存删除逻辑 | Restrictions: 遵循 .cursor/rules/go-modules.mdc，使用 Mock Cache | Success: 测试通过，缓存删除逻辑正确

---

## Phase 4: 集成测试

### 端到端测试

- [ ] 4.1 编写集成测试

  - File: `test/integration/product_package_inventory_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/product_inventory_test.go`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整的商品包库存查询流程，包括: 1) 多个BOM的最小值计算 2) 缓存机制 3) 缓存失效 4) 边界情况处理 | Restrictions: 测试真实用户场景，使用真实数据库和缓存 | Success: 集成测试通过，端到端流程正确

---

## Phase 5: 性能优化和文档

### 性能优化

- [ ] 5.1 性能测试和优化

  - File: `main/app/modules/inventory/application/product_inventory_app_service.go`, `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Prompt: Role: Performance Engineer | Task: 进行性能测试和优化 | Context: 1) 测试响应时间 < 200ms 2) 测试缓存命中率 > 80% 3) 优化数据库查询（使用索引）4) 优化缓存策略 | Restrictions: 遵循性能要求 | Success: 响应时间 < 200ms，缓存命中率 > 80%

### 文档更新

- [ ] 5.2 更新代码注释和文档

  - File: `main/app/modules/inventory/domain/service/product_inventory_domain_service.go`, `main/app/modules/inventory/application/product_inventory_app_service.go`
  - Purpose: 确保代码注释完整
  - Requirements: 文档要求
  - Leverage: 现有代码注释风格
  - Prompt: Role: Technical Writer | Task: 更新代码注释 | Context: 为新增方法添加中文注释，说明方法用途、参数、返回值、错误情况 | Restrictions: 注释清晰完整 | Success: 所有新增方法都有完整的中文注释

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Domain Service: ≥ 90%
  - Application Service: ≥ 80%
  - 商品包库存计算逻辑: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 代码注释完整
- [ ] design.md 已更新（如有调整）
- [ ] tasks.md 中的任务已完成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/go-modules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-warehouse-product-package-inventory-query/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-warehouse-product-package-inventory-query/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-warehouse-product-package-inventory-query/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-warehouse-product-package-inventory-query/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-warehouse-product-package-inventory-query/tasks.md)" | bc
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
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/go-modules.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口或 Repository 接口
- Repository 只持有 db 实例
- 使用 pkg/context.Context
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率达标
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 90% (Domain Service) 或 ≥ 80% (Application Service)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

Restrictions:
- 遵循 .cursor/rules/go-modules.mdc
- 必须包含边界情况测试
- 使用 Mock 对象

Success Criteria:
- 测试覆盖率达标
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
**最后更新**: 2025-12-11  
**维护者**: 后端开发组

