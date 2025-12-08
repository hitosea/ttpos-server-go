# 收银机沽清商品设置（多终端统一沽清判断） 任务分解

> 本文档定义 收银机沽清商品设置（多终端统一沽清判断） 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 25  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件添加新字段到 product_bom

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_sold_out_fields_to_product_bom_table.php`
  - Purpose: 添加 use_bom_card_stock, has_sellable_quantity, sellable_quantity 字段到 product_bom 表
  - Requirements: 2.3, 2.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件添加三个新字段到 ttpos_product_bom 表 | Context: use_bom_card_stock INT(1) default 1, has_sellable_quantity INT(1) default 0, sellable_quantity DECIMAL(22,4) default 0.0000 | Restrictions: 遵循 .cursor/rules/database.mdc，添加适当索引 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中应用新字段
  - Requirements: 2.3
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，表已更新

- [x] 1.3 更新 Go Model 添加新字段

  - File: `main/app/model/product.go`
  - Purpose: 在 ProductBom 结构体中添加新字段和 gorm 标签
  - Requirements: 2.3
  - Leverage: 现有 Model: `main/app/model/product.go` (ProductBom 结构体)
  - Prompt: Role: Go Developer | Task: 更新 ProductBom 结构体添加三个新字段 | Context: 使用 gorm 标签，类型 int 和 float64 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，编译通过

---

## Phase 2: DTO 和 API 更新

- [x] 2.1 更新 Request DTO 添加新字段

  - File: `main/app/dto/req/sold_out.go`
  - Purpose: 在 SoldOutItem 中添加 use_bom_card_stock, has_sellable_quantity, sellable_quantity 字段
  - Requirements: 2.1
  - Leverage: 现有 DTO: `main/app/dto/req/sold_out.go`
  - Prompt: Role: Go Developer | Task: 更新 SoldOutItem 添加可选新字段 | Context: 使用 pointer 类型和 json 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，binding 正确

- [x] 2.2 新增 GetSoldOutSettingsReq DTO

  - File: `main/app/dto/req/sold_out.go`
  - Purpose: 定义获取沽清设置的请求参数
  - Requirements: 1.4
  - Leverage: 现有 req DTOs
  - Prompt: Role: Go Developer | Task: 创建 GetSoldOutSettingsReq 结构体 | Context: 包含 product_package_uuid 字段，binding required | Restrictions: 遵循规范 | Success: DTO 创建成功

- [x] 2.3 新增 Response DTOs

  - File: `main/app/dto/resp/sold_out.go`
  - Purpose: 定义 SoldOutSettingsResp 和 SoldOutSetting 结构体
  - Requirements: 1.5
  - Leverage: 现有 resp DTOs
  - Prompt: Role: Go Developer | Task: 创建响应 DTOs | Context: 包含所有设置字段 | Restrictions: data 是对象 | Success: DTOs 创建成功

- [x] 2.4 新增 API: 获取沽清设置

  - File: `main/app/api/v1/cashier/cashier_sold_out.go`
  - Purpose: 实现 GET /cashier/sold_out/settings 接口
  - Requirements: Requirement 1 (1.1, 1.2)
  - Leverage: 现有 sold_out API
  - Prompt: Role: Go Developer with Gin expertise | Task: 添加 GetSettings 方法 | Context: BindQuery req, 调用 Service, 使用 helper.Success | Restrictions: snake_case URL | Success: API 添加成功，响应格式正确

- [ ] 2.5 更新现有 API: 添加沽清商品

  - File: `main/app/api/v1/cashier/cashier_sold_out.go`
  - Purpose: 修改 AddSoldOut 支持新参数（DTO已更新，无需修改API逻辑）
  - Requirements: Requirement 2
  - Leverage: 现有 Add 接口
  - Success: API 已支持新字段（通过 DTO 自动绑定）

- [x] 2.6 注册新 API 路由

  - File: `main/app/api/v1/cashier/cashier_sold_out.go`
  - Purpose: 在 RegisterSoldOutHandlers 中注册 GetSettings 路由
  - Requirements: Requirement 1
  - Leverage: 现有路由注册
  - Success: 路由注册成功

---

## Phase 3: Repository 和 Service 更新

- [ ] 3.1 更新 ProductRepo 支持查询商品所有规格（如需要）

  - File: `main/app/repository/product_repo.go`
  - Purpose: 添加 GetBomsByProductPackageUuid 方法（如不存在）
  - Requirements: Requirement 1
  - Leverage: 现有 Repo 方法
  - Prompt: Role: Go Developer | Task: 添加 GetBomsByProductPackageUuid 方法 | Context: 根据 product_package_uuid 查询所有规格 | Restrictions: 只持有 db | Success: 方法添加成功

- [ ] 3.2 更新 ISoldOutSrv 接口

  - File: `main/app/service/sold_out.go`
  - Purpose: 添加 GetSettings 方法到接口
  - Requirements: Requirement 1
  - Leverage: 现有接口
  - Prompt: Role: Go Developer | Task: 添加接口方法 | Context: 签名正确 | Success: 接口更新

- [ ] 3.3 实现 GetSettings Service 方法

  - File: `main/app/service/sold_out.go`
  - Purpose: 查询并计算沽清设置，包括成本卡库存
  - Requirements: 1.3, 1.2
  - Leverage: 现有 Service, ProductRepo
  - Prompt: Role: Go Developer | Task: 实现 GetSettings | Context: 查询 boms, 计算 stock（暂时返回0，后续实现成本卡计算） | Restrictions: 依赖 ProductRepo | Success: 方法实现，逻辑正确

- [ ] 3.4 更新 AddSoldOut Service 方法

  - File: `main/app/service/sold_out.go`
  - Purpose: 保存新沽清字段
  - Requirements: 2.2
  - Leverage: 现有 Add 方法
  - Prompt: Role: Go Developer | Task: 更新 AddSoldOut | Context: 更新 product_bom 新字段 | Restrictions: 保持 WebSocket 推送 | Success: 方法更新

---

## Phase 4: 统一沽清判断逻辑集成

- [ ] 4.1 实现统一沽清判断逻辑（CheckSoldOut）

  - File: `main/app/service/sold_out.go`
  - Purpose: 添加 CheckSoldOut 方法，根据设置判断商品是否可售
  - Requirements: Requirement 3 (3.1, 3.2, 3.3)
  - Leverage: 现有订单库存检查逻辑
  - Prompt: Role: Go Developer | Task: 添加 CheckSoldOut 方法 | Context: 根据设置判断可售，支持负库存 | Restrictions: 支持负库存场景 | Success: 逻辑实现

- [ ] 4.2 集成到订单创建逻辑

  - File: `main/app/service/order.go` (或相关)
  - Purpose: 在下单逻辑中调用 CheckSoldOut
  - Requirements: 3.1
  - Leverage: 现有订单 Service
  - Prompt: Role: Go Developer | Task: 集成 CheckSoldOut 到订单创建 | Context: 在下单前检查沽清 | Success: 逻辑集成

- [ ] 4.3 集成到送厨逻辑

  - File: `main/app/service/kitchen.go` (或相关)
  - Purpose: 应用沽清判断到送厨
  - Requirements: 3.2
  - Leverage: 现有送厨逻辑
  - Prompt: Role: Go Developer | Task: 集成 CheckSoldOut 到送厨 | Context: 类似下单逻辑 | Success: 更新完成

- [ ] 4.4 集成到结账逻辑

  - File: `main/app/service/payment.go` (或相关)
  - Purpose: 应用沽清判断到结账
  - Requirements: 3.3
  - Leverage: 现有结账逻辑
  - Prompt: Role: Go Developer | Task: 集成 CheckSoldOut 到结账 | Context: 类似下单逻辑 | Success: 更新完成

- [ ] 4.5 实现可售量扣减逻辑

  - File: `main/app/service/sold_out.go`
  - Purpose: 下单时扣减 sellable_quantity
  - Requirements: 3.4
  - Leverage: Task 4.1
  - Prompt: Role: Go Developer | Task: 添加 DeductSellableQuantity 方法 | Context: 原子更新 | Restrictions: 并发锁 | Success: 扣减逻辑实现

- [ ] 4.6 实现成本卡库存计算

  - File: `main/app/service/inventory.go` (或相关)
  - Purpose: 计算 bom_card_stock_num
  - Requirements: 1.3
  - Leverage: 现有库存计算
  - Prompt: Role: Go Developer | Task: 实现或更新 CalculateBomStock | Context: 基于 product_bom_card | Success: 计算正确

- [ ] 4.7 实现负库存判断

  - File: `main/app/service/sold_out.go`
  - Purpose: 根据设置允许/禁止负库存
  - Requirements: 3.6
  - Leverage: Task 4.1
  - Prompt: Role: Go Developer | Task: 在判断中添加负库存逻辑 | Context: 如果允许负库存，继续售卖 | Success: 判断逻辑完整

- [ ] 4.8 更新商品库存查询显示可售量

  - File: `main/app/service/product.go` (或相关)
  - Purpose: 查询时包含 sellable_quantity
  - Requirements: 3.7
  - Leverage: 现有查询
  - Prompt: Role: Go Developer | Task: 更新 GetProductStock | Context: Join product_bom | Success: 查询更新

---

## Phase 5: 优化和测试

- [ ] 5.1 实现缓存策略

  - File: `main/app/service/sold_out.go`
  - Purpose: 缓存沽清设置
  - Requirements: 非功能 - 性能
  - Leverage: 现有 Redis 工具
  - Prompt: Role: Go Developer | Task: 添加 Cache-Aside 到 GetSettings | Context: Key: ttpos:cashier:sold_out:settings:{product_package_uuid} | Success: 缓存实现

- [ ] 5.2 添加并发控制（UUID 锁）

  - File: `main/app/service/sold_out.go`
  - Purpose: 防止并发更新沽清
  - Requirements: 非功能 - 可靠性
  - Leverage: `pkg/lock/system_lock.go`
  - Prompt: Role: Go Developer | Task: 在 Update 中添加锁 | Context: 使用 UUID 锁 | Success: 并发安全

- [ ] 5.3 Repository 单元测试

  - File: `main/app/repository/product_repo_test.go`
  - Purpose: 测试新字段更新和查询
  - Requirements: 测试验收
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 更新测试覆盖新字段 | Context: 覆盖率 ≥ 80% | Success: 测试通过

- [ ] 5.4 Service 单元测试

  - File: `main/app/service/sold_out_service_test.go`
  - Purpose: 测试 GetSettings 和 AddSoldOut
  - Requirements: 测试验收
  - Leverage: 现有测试
  - Prompt: Role: QA Engineer | Task: 添加测试 for 新方法 | Context: Mock ProductRepo, 覆盖率 ≥ 70% | Success: 测试通过

- [ ] 5.5 API 测试

  - File: `main/app/api/v1/cashier/cashier_sold_out_test.go`
  - Purpose: 测试新/更新 API
  - Requirements: 2.1, 测试验收
  - Leverage: 现有 API 测试
  - Prompt: Role: QA Engineer | Task: 测试接口 | Context: 正常/错误场景 | Success: API 测试通过

- [ ] 5.6 集成测试：沽清判断流程

  - File: `test/integration/sold_out_test.go`
  - Purpose: 测试下单/送厨/结账集成
  - Requirements: Requirement 3
  - Leverage: 现有集成测试
  - Prompt: Role: QA Engineer | Task: 端到端测试 | Context: 设置沽清 → 下单 → 扣减 | Success: 流程测试通过

- [ ] 5.7 并发测试

  - File: `test/integration/concurrent_sold_out_test.go`
  - Purpose: 测试并发下单场景
  - Requirements: 非功能 - 并发
  - Leverage: Go 测试工具
  - Success: 无超卖，性能达标

---

## Phase 6: 文档和清理

- [ ] 6.1 更新 API 文档

  - File: `docs/shared/api/cashier_api.md`
  - Purpose: 文档化新/更新接口
  - Requirements: 非功能 - 文档
  - Leverage: 现有 API 文档
  - Prompt: Role: Technical Writer | Task: 添加接口描述 | Context: 基于 design.md | Success: 文档更新

- [ ] 6.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录变更
  - Requirements: version.mdc
  - Leverage: 现有日志
  - Success: 更新完成

- [ ] 6.3 代码审查和格式化

  - File: 所有修改文件
  - Purpose: 确保代码质量
  - Requirements: 非功能
  - Command: `go fmt ./...`, `go vet ./...`
  - Success: 检查通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - 订单相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] 数据库文档已更新
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
grep -c "^- \[" docs/shared/specs/active/story-shop-sold-out-settings/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-sold-out-settings/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-sold-out-settings/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-sold-out-settings/tasks.md)" | bc
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
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc, .cursor/rules/database.mdc

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
- 测试覆盖率 ≥ 70% (Service) 或 ≥ 80% (Repository)
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试

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

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-08  
**维护者**: 后端开发组

