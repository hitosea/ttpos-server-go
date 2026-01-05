# 总店删除资源时子店使用情况验证 任务分解（后端）

> 本文档定义**总店**删除规格、属性组、属性、加料、单位、商品时，后端检查子店使用情况的详细执行任务清单。

**🎯 核心要点**：
1. **只有总店删除时**才检查子店使用情况
2. **返回所有**使用的子店名称（不限制数量）
3. 子店删除自己的资源无需检查

**📦 范围说明**：本任务清单仅包含 Go Main 后端实现，前端和 ERP 部分由其他开发者负责。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 27（仅后端）  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 0: 独立查询接口（0.5 天）

### 新增独立接口

- [ ] 0.1 创建 AttributeCheckReq DTO

  - File: `main/app/dto/req/shop_product_req.go`
  - Purpose: 定义属性使用情况查询请求参数
  - Requirements: 0.1, 0.4
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 AttributeCheckReq 结构体 | Context: 包含 Uuid 字段，同时支持 POST (json) 和 GET (form) | Restrictions: 使用 binding:"required" 验证 | Success: DTO 定义正确

- [ ] 0.2 实现 CheckAttributeUsage Service 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 实现独立的属性使用情况查询（不执行删除）
  - Requirements: 0.1, 0.2, 0.3, 0.4
  - Leverage: Task 2.5 的 CheckAttributeUsageBeforeDelete 实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 CheckAttributeUsage 方法 | Context: 1) 检查总店身份（子店返回错误） 2) 跨数据库查询所有子店 3) 返回所有使用的子店名称 | Restrictions: companySetting.IsSubShop() 返回错误，使用跨数据库查询模式 | Success: 总店检查正确，子店调用返回错误

- [ ] 0.3 创建 CheckAttributeUsage API 接口

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 创建独立的属性使用情况查询接口
  - Requirements: 0.1, 0.2, 0.3, 0.4
  - Leverage: Task 0.2 的 Service 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 CheckAttributeUsage API，路由为 POST /api/v1/shop/attribute/check_usage | Context: 调用 CheckAttributeUsage Service 方法，返回使用情况数据 | Restrictions: 使用 helper.Success 返回成功，data 包含 is_used, used_by_shops, total_count | Success: API 创建正确，响应格式准确

- [ ] 0.4 注册路由

  - File: `main/router/router.go`
  - Purpose: 注册属性使用情况查询接口路由
  - Requirements: 0.1
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功

- [ ] 0.5 创建 AttributesBatchCheckReq DTO

  - File: `main/app/dto/req/shop_product_req.go`
  - Purpose: 定义批量检查属性使用情况的请求参数
  - Requirements: 0.6
  - Leverage: Task 0.1 的实现
  - Prompt: Role: Go Developer | Task: 创建 AttributesBatchCheckReq 结构体 | Context: 包含 Uuids []uint64 字段，使用 binding:"required,min=1" 验证 | Restrictions: 数组至少包含 1 个元素 | Success: DTO 定义正确

- [ ] 0.6 实现 CheckAttributesUsageBatch Service 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 实现批量检查属性使用情况
  - Requirements: 0.6
  - Leverage: Task 0.2 的 CheckAttributeUsage 方法
  - Prompt: Role: Go Developer | Task: 实现 CheckAttributesUsageBatch 方法 | Context: 循环调用 CheckAttributeUsage，返回 map[uint64]*ResourceUsageResp | Restrictions: 检查总店身份 | Success: 批量查询正确

- [ ] 0.7 创建批量检查 API 接口

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 创建批量检查属性使用情况接口
  - Requirements: 0.6
  - Leverage: Task 0.3 的实现
  - Prompt: Role: Go Developer | Task: 创建 CheckAttributesUsageBatch API，路由为 POST /api/v1/shop/attribute/check_usage_batch | Context: 返回 map 结构，key 为属性UUID，value 为使用情况 | Restrictions: data 必须是对象 | Success: API 创建正确

- [ ] 0.8 修改编辑属性组 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在编辑属性组时检查被移除的属性是否被子店使用
  - Requirements: 0.5, 0.6
  - Leverage: Task 0.6 的批量检查方法
  - Prompt: Role: Go Developer | Task: 修改 UpdateAttributeGroup API，在更新前检查被移除的属性 | Context: 1) 获取原属性组的属性列表 2) 对比找出被移除的属性 3) 批量检查使用情况 4) 如果有被使用的，阻止更新并返回详细信息 | Restrictions: 返回被使用的属性名称和子店列表 | Success: 编辑检查正确，阻止移除被使用的属性

- [ ] 0.9 编写 API 测试

  - File: `main/app/api/v1/shop/shop_product_test.go`
  - Purpose: 测试独立查询接口和编辑属性组检查
  - Requirements: 0.1-0.6
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 测试 CheckAttributeUsage、CheckAttributesUsageBatch、UpdateAttributeGroup API | Context: 测试总店查询成功、子店查询失败、批量查询、编辑时移除被使用的属性 | Restrictions: 验证响应格式 | Success: 所有测试通过

---

## Phase 1: Repository 层实现（0.5 天）

### 扩展 Product Repository

⚠️ **架构说明**：本项目每个门店有独立数据库，需采用跨数据库查询方案

- [ ] 1.1 扩展 Product Repository 接口

  - File: `main/app/repository/i_product_repo.go`
  - Purpose: 定义检查单个子店使用情况的接口方法（单数据库查询）
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有 Repository 接口: `main/app/repository/i_product_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 IProductRepo 接口中添加以下方法：CheckFlavorUsageInShop, CheckAttributeGroupUsageInShop, CheckAttributeUsageInShop, CheckSauceUsageInShop, CheckUnitUsageInShop, CheckProductUsageInPackage | Context: 每个方法返回 (bool, error)，检查当前数据库是否使用指定资源 | Restrictions: 遵循 .cursor/rules/go-main.mdc，接口以 I 开头 | Success: 接口定义完整，方法签名正确

- [ ] 1.2 实现 CheckFlavorUsageInShop 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店是否使用了指定规格（Flavor）
  - Requirements: 1.1, 3.1
  - Leverage: 现有 Product Repository 实现，参考 design.md 中的 SQL 查询
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 CheckFlavorUsageInShop 方法，查询 ttpos_product_bom 表检查 product_flavor_uuid | Context: 使用 INNER JOIN ttpos_product_package，检查软删除，返回 count > 0 | Restrictions: 只持有 db *gorm.DB，使用 GORM | Success: 方法实现正确，查询优化

- [ ] 1.3 实现 CheckAttributeGroupUsageInShop 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店是否使用了指定属性组
  - Requirements: 1.2, 3.3
  - Leverage: Task 1.2 的实现，查询 ttpos_product_package_attribute_group 表
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 CheckAttributeGroupUsageInShop 方法 | Context: 通过 ttpos_product_package_attribute_group 表检查 product_attribute_group_uuid，JOIN ttpos_product_package 检查商品 | Restrictions: 软删除检查 | Success: 方法实现正确
  
- [ ] 1.4 实现 CheckAttributeUsageInShop 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店是否使用了指定属性
  - Requirements: 1.3, 3.1
  - Leverage: Task 1.2 的实现，查询 ttpos_product_package_attribute 表
  - Success: 方法实现正确

- [ ] 1.5 实现 CheckSauceUsageInShop 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店是否使用了指定加料（Sauce）
  - Requirements: 1.4, 3.1
  - Leverage: Task 1.2 的实现，查询 product_sauce_uuid
  - Success: 方法实现正确

- [ ] 1.6 实现 CheckUnitUsageInShop 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店是否使用了指定单位
  - Requirements: 1.5, 3.1
  - Leverage: Task 1.2 的实现，查询 ttpos_product_package.unit_uuid
  - Success: 方法实现正确

- [ ] 1.7 实现 CheckProductUsageInPackage 方法

  - File: `main/app/repository/product_repo.go`
  - Purpose: 检查当前子店的套餐是否使用了指定商品
  - Requirements: 1.6, 3.2
  - Leverage: Task 1.2 的实现，查询套餐商品关联
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 CheckProductUsageInPackage 方法，查询套餐商品关联 | Context: 通过 ttpos_product_bom 查询套餐（product_type=1）是否包含指定商品 | Restrictions: 软删除检查 | Success: 方法实现正确

- [ ] 1.8 编写 Repository 单元测试

  - File: `main/app/repository/product_repo_test.go`
  - Purpose: 测试子店使用情况查询方法
  - Requirements: 所有 Repository 需求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为新增的 6 个查询方法编写单元测试 | Context: 测试使用和未使用两种情况、软删除过滤 | Restrictions: 测试覆盖率 ≥ 80% | Success: 所有测试通过，覆盖率达标

---

## Phase 2: Service 层实现（1 天）

### 扩展 Product Service

- [ ] 2.1 扩展 Product Service 接口

  - File: `main/app/service/i_product_srv.go`
  - Purpose: 定义检查资源使用情况的接口方法（跨数据库版本）
  - Requirements: 0.1-0.4, 1.1-1.6, 2.1-2.3
  - Leverage: 现有 Service 接口: `main/app/service/i_product_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IProductSrv 接口中添加：1) 独立查询接口 CheckAttributeUsage 2) 删除前检查接口 CheckFlavorUsageBeforeDelete, CheckAttributeGroupUsageBeforeDelete, CheckAttributeUsageBeforeDelete, CheckSauceUsageBeforeDelete, CheckUnitUsageBeforeDelete, CheckProductUsageBeforeDelete | Context: 每个方法返回 *dto_resp.ResourceUsageResp 和 error | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，包含独立查询和删除前检查

- [ ] 2.2 创建 ResourceUsageResp DTO

  - File: `main/app/dto/resp/shop_product_resp.go`
  - Purpose: 定义资源使用情况响应结构
  - Requirements: 1.7
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 ResourceUsageResp 结构体 | Context: 包含 IsUsed (bool), UsedByShops ([]string), TotalCount (int) | Restrictions: 使用 json tag | Success: DTO 定义正确

- [ ] 2.3 实现 CheckFlavorUsageBeforeDelete 方法（跨数据库版本）

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查规格删除前的子店使用情况（遍历所有子店数据库）
  - Requirements: 1.1, 1.2, 3.1, 3.6
  - Leverage: Task 1.2 的 Repository 方法，参考 design.md 中的跨数据库实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 CheckFlavorUsageBeforeDelete 方法（跨数据库） | Context: 0) 检查是否为总店（只有总店才检查） 1) 从 SAAS 数据库获取子店列表 2) 并发遍历每个子店数据库 3) 使用 goroutine + 信号量限流(20并发) 4) 汇总结果返回**所有**子店名称（不限制数量） | Restrictions: 使用 constant.DefaultDB 获取 SAAS DB，使用 sync.WaitGroup 和 sync.Mutex，companySetting.IsSubShop() 判断总店 | Success: 总店身份检查正确，跨数据库查询正确，返回所有子店名称，并发安全

- [ ] 2.4 实现 CheckAttributeGroupUsageBeforeDelete 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查属性组删除前的子店使用情况（跨数据库）
  - Requirements: 1.2, 3.3
  - Leverage: Task 2.3 的跨数据库实现模式
  - Success: 方法实现正确

- [ ] 2.5 实现 CheckAttributeUsageBeforeDelete 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查属性删除前的子店使用情况（跨数据库）
  - Requirements: 1.3, 1.7
  - Leverage: Task 2.3 的跨数据库实现模式
  - Success: 方法实现正确

- [ ] 2.6 实现 CheckSauceUsageBeforeDelete 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查加料删除前的子店使用情况（跨数据库）
  - Requirements: 1.4, 1.7
  - Leverage: Task 2.3 的跨数据库实现模式
  - Success: 方法实现正确

- [ ] 2.7 实现 CheckUnitUsageBeforeDelete 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查单位删除前的子店使用情况（跨数据库）
  - Requirements: 1.5, 1.7
  - Leverage: Task 2.3 的跨数据库实现模式
  - Success: 方法实现正确

- [ ] 2.8 实现 CheckProductUsageBeforeDelete 方法

  - File: `main/app/service/product_srv.go`
  - Purpose: 检查商品删除前的子店使用情况（跨数据库）
  - Requirements: 1.6, 1.7
  - Leverage: Task 2.3 的跨数据库实现模式
  - Success: 方法实现正确

- [ ] 2.9 编写 Service 单元测试

  - File: `main/app/service/product_srv_test.go`
  - Purpose: 测试子店使用情况检查逻辑
  - Requirements: 所有 Service 需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为新增的 Check 方法编写单元测试 | Context: Mock Repository，测试被使用和未被使用两种情况 | Restrictions: 测试覆盖率 ≥ 70% | Success: 所有测试通过，覆盖率达标

---

## Phase 3: API 层实现（0.5 天）

### 修改删除 API

- [ ] 3.1 修改规格删除 API（Flavor）

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除规格前检查子店使用情况（只有总店删除时检查）
  - Requirements: 1.1, 1.2, 1.9, 1.10
  - Leverage: 现有删除 API，Task 2.3 的 Service 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 修改 DeleteFlavor API，在删除前调用 CheckFlavorUsageBeforeDelete | Context: Service 层已处理总店身份检查，如果 IsUsed=true，使用 helper.ErrorWithData 返回错误，data 中包含所有子店名称（不限制数量） | Restrictions: data 必须是对象，包含 is_used, used_by_shops（所有子店）, total_count | Success: API 修改正确，返回所有子店名称

- [ ] 3.2 修改属性组删除 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除属性组前检查子店使用情况
  - Requirements: 1.2, 3.3, 1.9
  - Leverage: Task 3.1 的实现
  - Success: API 修改正确

- [ ] 3.3 修改属性删除 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除属性前检查子店使用情况
  - Requirements: 1.3, 1.8, 1.9
  - Leverage: Task 3.1 的实现
  - Success: API 修改正确

- [ ] 3.4 修改加料删除 API（Sauce）

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除加料前检查子店使用情况
  - Requirements: 1.4, 1.8, 1.9
  - Leverage: Task 3.1 的实现
  - Success: API 修改正确

- [ ] 3.5 修改单位删除 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除单位前检查子店使用情况
  - Requirements: 1.5, 1.8, 1.9
  - Leverage: Task 3.1 的实现
  - Success: API 修改正确

- [ ] 3.6 修改商品删除 API

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在删除商品前检查子店使用情况
  - Requirements: 1.6, 1.8, 1.9
  - Leverage: Task 3.1 的实现
  - Success: API 修改正确

- [ ] 3.7 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_product_test.go`
  - Purpose: 测试删除 API 的验证逻辑
  - Requirements: 所有 API 需求
  - Leverage: 现有测试: `main/app/api/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 测试 6 种资源删除 API 的两种情况：被使用（失败）和未使用（成功） | Context: 验证响应格式、提示文案、HTTP 状态码 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 4: 测试和优化（0.5 天）

### 集成测试

- [ ] 4.1 端到端集成测试

  - File: `test/integration/shop_product_delete_test.go`
  - Purpose: 测试完整的删除流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试场景1（子店使用-删除失败-返回使用情况）、场景2（子店未使用-删除成功） | Restrictions: 使用真实数据库，测试跨数据库查询 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 测试 100 个子店时的查询性能
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk）、模拟 100 个子店数据库
  - Prompt: Role: Performance Engineer | Task: 测试 100 个子店场景下的删除 API 性能 | Context: 使用 wrk 压测，验证并发查询优化效果 | Success: 响应时间 < 500ms（100 个子店）

- [ ] 4.3 数据库查询优化验证

  - File: -
  - Purpose: 验证单个子店的 SQL 查询性能
  - Requirements: 性能要求
  - Leverage: MySQL EXPLAIN 分析
  - Prompt: Role: Database Engineer | Task: 使用 EXPLAIN 分析查询计划，验证索引使用 | Context: 检查 CheckFlavorUsageInShop 等查询的执行计划 | Success: 查询使用索引，扫描行数少

### 文档更新

- [ ] 4.4 更新 API 文档

  - File: `docs/shared/api/shop_product_api.md`
  - Purpose: 更新删除接口文档
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 响应格式文档完整准确

- [ ] 4.5 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG 格式
  - Success: 变更记录完整


---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有后端任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过
- [ ] 性能测试达标（100 个子店 < 500ms）

### 功能完整性

- [ ] requirements.md 中的所有后端需求已满足
- [ ] design.md 中的后端设计已实现
- [ ] 后端验收标准已达成
- [ ] 6 种资源类型删除验证全部实现（规格、属性组、属性、加料、单位、商品）
- [ ] API 响应格式符合前端约定

### 文档同步

- [ ] API 响应格式文档已更新
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

### 集成接口

- [ ] API 响应格式与前端约定一致（包含 is_used, used_by_shops, total_count）
- [ ] 预留 ERP 同步扩展点（由 ERP 集成开发者后续实现）

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-product-parent-shop-delete-validation/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-product-parent-shop-delete-validation/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-product-parent-shop-delete-validation/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-product-parent-shop-delete-validation/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-product-parent-shop-delete-validation/tasks.md)" | bc
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

### Go Repository 层开发

```
Role: Go Developer specializing in Repository Pattern with GORM expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/database.mdc

Restrictions:
- 只持有 db *gorm.DB 实例
- 使用 GORM 查询
- 软删除检查（delete_time = 0）
- 使用 INNER JOIN 优化查询
- 返回 bool（是否使用）
- 不使用 panic，返回 error

Success Criteria:
- {成功标准}
- 查询使用索引
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### Go Service 层开发（跨数据库）

```
Role: Go Developer with business logic expertise and multi-database query experience

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc
- Architecture: 每个门店独立数据库，需要跨数据库并发查询

Restrictions:
- 持有 DBManager，通过 GetDB(shopUuid) 连接不同门店数据库
- 从 constant.DefaultDB 获取 SAAS 数据库连接
- 使用 goroutine 并发查询，信号量限流（20 并发）
- 使用 sync.WaitGroup 和 sync.Mutex 保证并发安全
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误
- 使用 logger.Logger 记录日志

Success Criteria:
- {成功标准}
- 跨数据库查询正确
- 并发安全
- 性能达标（100 个子店 < 500ms）
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### Go API 层开发

```
Role: Go Developer with Gin framework expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc

Restrictions:
- 使用 helper.Success() 返回成功响应
- 使用 helper.Error() 返回错误
- data 字段必须是对象，包含 is_used, used_by_shops, total_count
- 不直接使用 c.JSON()
- URL 使用 snake_case

Success Criteria:
- {成功标准}
- API 响应格式正确
- data 字段包含使用情况数据
- 代码通过 go fmt 和 go vet
```

### Go 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景：子店使用资源，删除失败，返回使用情况
- 正常场景：子店未使用资源，删除成功
- 边界条件：无子店
- 边界条件：大量子店（100 个）
- 边界条件：超过 10 个子店使用（验证只返回前 10 个）
- 异常场景：数据库连接失败
- 异常场景：并发查询中部分子店查询失败

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 使用 testify/assert
- Mock DBManager 和多个数据库连接
- 测试跨数据库并发查询

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
- 并发场景已测试
```

---

## Graphiti & 活动日志

- Related Episode: `待补充`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-04  
**维护者**: 后端开发组

