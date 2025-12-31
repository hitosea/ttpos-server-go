# 厨显（Grab/LINE MAN外卖相关） 任务分解

> 本文档定义厨显系统中 Grab/LINE MAN 外卖订单标识和商品名称统一显示的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 27  
**已完成**: 12  
**进行中**: -  
**完成率**: 44%

---

## Phase 1: 数据库迁移和模型修改（1 天）

### 数据库迁移

- [x] 1.1 创建数据库迁移脚本，添加 `takeout_order_uuid` 字段

  - File: `admin/database/migrations/20251225145527_add_takeout_order_uuid_to_production_order_table.php`
  - Purpose: 在 `ttpos_production_order` 表中添加 `takeout_order_uuid` 字段，用于关联外卖订单
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有迁移脚本: `main/manifest/sql/`
  - Prompt: Role: Database Engineer | Task: 创建数据库迁移脚本，在 ttpos_production_order 表中添加 takeout_order_uuid 字段 | Context: 字段类型为 BIGINT UNSIGNED，默认值为 0，添加注释说明关联到 ttpos_takeout_order.uuid | Restrictions: 遵循 .cursor/rules/database.mdc，确保迁移脚本可重复执行（idempotent） | Success: 迁移脚本创建完成，语法正确，可重复执行
  - SQL 示例:
    ```sql
    ALTER TABLE `ttpos_production_order` 
    ADD COLUMN IF NOT EXISTS `takeout_order_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 
    COMMENT '外卖订单UUID（关联 ttpos_takeout_order.uuid）' 
    AFTER `sale_bill_uuid`;
    ```

- [x] 1.2 为 `takeout_order_uuid` 字段添加索引

  - File: `admin/database/migrations/20251225145527_add_takeout_order_uuid_to_production_order_table.php`
  - Purpose: 为 `takeout_order_uuid` 字段添加索引，优化查询性能
  - Requirements: Requirement 1.1, 2.1
  - Leverage: Task 1.1 的迁移脚本
  - Prompt: Role: Database Engineer | Task: 在迁移脚本中添加索引创建语句 | Context: 为 takeout_order_uuid 字段创建索引，索引名称为 idx_takeout_order_uuid | Restrictions: 遵循 .cursor/rules/database.mdc，检查索引是否已存在（idempotent） | Success: 索引创建语句添加完成，语法正确
  - SQL 示例:
    ```sql
    CREATE INDEX IF NOT EXISTS `idx_takeout_order_uuid` 
    ON `ttpos_production_order` (`takeout_order_uuid`);
    ```

### Model 层修改

- [x] 1.3 更新 `ProductionOrder` Model，添加 `TakeoutOrderUuid` 字段

  - File: `main/app/model/production.go`
  - Purpose: 在 ProductionOrder 结构体中添加 TakeoutOrderUuid 字段
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 ProductionOrder 结构体定义
  - Prompt: Role: Go Developer | Task: 在 ProductionOrder 结构体中添加 TakeoutOrderUuid 字段 | Context: 字段类型为 uint64，gorm 标签为 column:takeout_order_uuid，添加注释说明 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段命名使用 PascalCase | Success: 字段添加完成，gorm 标签正确

- [x] 1.4 在 `ProductionOrder` Model 中添加 `IsTakeoutOrder()` 方法

  - File: `main/app/model/production.go`
  - Purpose: 添加判断是否为外卖订单的方法
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 ProductionOrder 方法，参考 SaleBill.IsTakeoutBill() 方法
  - Prompt: Role: Go Developer | Task: 在 ProductionOrder 结构体中添加 IsTakeoutOrder() 方法 | Context: 方法返回 bool，当 TakeoutOrderUuid > 0 时返回 true | Restrictions: 遵循 .cursor/rules/go-main.mdc，方法命名清晰 | Success: 方法实现完成，逻辑正确
  - Code 示例:
    ```go
    // 判断是否为外卖订单
    func (p *ProductionOrder) IsTakeoutOrder() bool {
        return p.TakeoutOrderUuid > 0
    }
    ```

- [x] 1.5 更新 `ProductionOrder.Source` 字段注释

  - File: `main/app/model/production.go`
  - Purpose: 更新 Source 字段的注释，说明支持 grab, lineman 值
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 ProductionOrder.Source 字段定义
  - Prompt: Role: Go Developer | Task: 更新 ProductionOrder.Source 字段的注释 | Context: 在注释中说明 source 字段支持的值包括 shop, cashier, tablet, kitchen, assistant, h5, grab, lineman | Restrictions: 遵循 .cursor/rules/go-main.mdc，注释清晰准确 | Success: 注释更新完成，说明完整

- [x] 1.6 在 `ProductionOrderProduct` Model 中添加 `ProductionOrder` 关联

  - File: `main/app/model/production.go`
  - Purpose: 在 ProductionOrderProduct 结构体中添加 ProductionOrder 关联，便于查询时预加载
  - Requirements: Requirement 3.1
  - Leverage: 现有 ProductionOrderProduct 结构体，参考 SaleBill 关联
  - Prompt: Role: Go Developer | Task: 在 ProductionOrderProduct 结构体中添加 ProductionOrder 关联字段 | Context: 使用 gorm foreignKey:ProductionOrderUuid;references:Uuid 标签，字段类型为 ProductionOrder | Restrictions: 遵循 .cursor/rules/go-main.mdc，关联定义正确 | Success: 关联字段添加完成，gorm 标签正确
  - Code 示例:
    ```go
    ProductionOrder ProductionOrder `gorm:"foreignKey:ProductionOrderUuid;references:Uuid" json:"production_order"`
    ```

---

## Phase 2: 外卖订单接单流程修改（1-2 天）

### 接单逻辑修改

- [x] 2.1 修改外卖订单接单逻辑，创建 `ProductionOrder`

  - File: `main/app/modules/takeout/domain/service/takeout_order_service.go`
  - Purpose: 在 AcceptOrder 方法中，创建 ProductionOrder 时设置 takeout_order_uuid 和 source 字段
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 AcceptOrder 方法，参考店内订单创建 ProductionOrder 的逻辑
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 修改 AcceptOrder 方法，创建 ProductionOrder 时设置外卖订单相关字段 | Context: 设置 takeout_order_uuid = TakeoutOrder.uuid，source = TakeoutOrder.platform（grab, lineman），sale_bill_uuid = 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保外卖订单不关联 SaleBill | Success: 接单逻辑修改完成，ProductionOrder 创建正确
  - Key Points:
    - 设置 `takeout_order_uuid = TakeoutOrder.uuid`
    - 设置 `source = TakeoutOrder.platform`（grab, lineman）
    - 设置 `sale_bill_uuid = 0`（外卖订单不关联 SaleBill）

- [x] 2.2 创建 `ProductionOrderProduct`，关联到 `ProductionOrder`

  - File: `main/app/modules/takeout/domain/service/takeout_order_service.go`
  - Purpose: 在接单后创建 ProductionOrderProduct，关联到外卖订单的 ProductionOrder
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有创建 ProductionOrderProduct 的逻辑，参考店内订单的送厨商品创建
  - Prompt: Role: Go Developer specializing in Domain Service | Task: 在 AcceptOrder 方法中创建 ProductionOrderProduct | Context: 根据 TakeoutOrderItems 创建 ProductionOrderProduct，关联到新创建的 ProductionOrder | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保商品信息正确映射 | Success: ProductionOrderProduct 创建完成，关联正确

- [ ] 2.3 验证外卖订单接单后正确创建送厨单和送厨商品

  - File: `main/app/modules/takeout/domain/service/takeout_order_service.go`
  - Purpose: 验证接单流程，确保 ProductionOrder 和 ProductionOrderProduct 创建正确
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有测试用例或手动测试
  - Test Cases:
    - 创建 Grab 订单，接单后验证 ProductionOrder.takeout_order_uuid 正确
    - 创建 LINE MAN 订单，接单后验证 ProductionOrder.source 为 "lineman"
    - 验证 ProductionOrder.sale_bill_uuid 为 0
    - 验证 ProductionOrderProduct 正确创建并关联到 ProductionOrder
  - Success: 所有测试用例通过，数据创建正确

---

## Phase 3: 厨显接口修改和验证（1-2 天）

### Repository 层修改

- [x] 3.1 修改 `GetProducts` 方法，预加载 `ProductionOrder` 关联

  - File: `main/app/repository/production_order.go`
  - Purpose: 在查询 ProductionOrderProduct 时预加载 ProductionOrder 关联
  - Requirements: Requirement 1.1, 2.1, 3.1
  - Leverage: 现有 GetProducts 方法，参考 SaleBill 预加载逻辑
  - Prompt: Role: Go Developer specializing in Repository Layer | Task: 在 GetProducts 方法中添加 ProductionOrder 预加载 | Context: 在现有 Preload 链中添加 Preload("ProductionOrder") | Restrictions: 遵循 .cursor/rules/go-main.mdc，预加载顺序合理 | Success: 预加载添加完成，查询时 ProductionOrder 数据正确加载
  - Code 示例:
    ```go
    db.Preload("ProductionOrder").  // ✅ 新增：预加载 ProductionOrder
        Preload("SaleBill").
        Preload("BatchTag.MultiLanguageName").
        // ... 其他预加载
    ```

### Service 层修改

- [x] 3.2 更新 `ProductionGroup` DTO，添加 `IsTakeoutOrder` 字段

  - File: `main/app/dto/resp/production.go`
  - Purpose: 在 ProductionGroup 结构体中添加 IsTakeoutOrder 字段，用于标识第三方平台外卖订单
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 ProductionGroup 结构体，参考 IsTakeoutBill 字段
  - Prompt: Role: Go Developer | Task: 在 ProductionGroup 结构体中添加 IsTakeoutOrder 字段 | Context: 字段类型为 bool，json 标签为 is_takeout_order，添加注释说明 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段命名清晰 | Success: 字段添加完成，json 标签正确

- [x] 3.3 修改 `groupByOrder` 方法，区分传统外送和第三方平台外卖

  - File: `main/app/service/production.go`
  - Purpose: 在 groupByOrder 方法中，通过 ProductionOrder.IsTakeoutOrder() 判断外卖订单，设置 is_takeout_bill 和 is_takeout_order 字段
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有 groupByOrder 方法，参考 SaleBill.IsTakeoutBill() 判断逻辑
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 groupByOrder 方法，区分传统外送和第三方平台外卖 | Context: 传统外送订单通过 SaleBill.IsTakeoutBill() 判断，第三方平台外卖订单通过 ProductionOrder.IsTakeoutOrder() 判断 | Restrictions: 遵循 .cursor/rules/go-main.mdc，逻辑清晰准确 | Success: groupByOrder 方法修改完成，字段设置正确
  - Code 示例:
    ```go
    // 传统外送订单：通过 SaleBill.IsTakeoutBill() 判断
    group.IsTakeoutBill = product.SaleBill.IsTakeoutBill()
    // 第三方平台外卖订单：通过 ProductionOrder.IsTakeoutOrder() 判断
    group.IsTakeoutOrder = product.ProductionOrder.IsTakeoutOrder()
    ```

- [x] 3.4 修改 `groupByCategory` 方法，区分传统外送和第三方平台外卖

  - File: `main/app/service/production.go`
  - Purpose: 在 groupByCategory 方法中，通过 ProductionOrder.IsTakeoutOrder() 判断外卖订单，设置 is_takeout_bill 和 is_takeout_order 字段
  - Requirements: Requirement 2.1
  - Leverage: 现有 groupByCategory 方法，参考 Task 3.3 的实现
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 groupByCategory 方法，区分传统外送和第三方平台外卖 | Context: 使用与 groupByOrder 相同的判断逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持代码一致性 | Success: groupByCategory 方法修改完成，字段设置正确

### 验证和测试

- [ ] 3.5 验证 `GetProductListByOrder` 返回的字段正确

  - File: `main/app/service/production.go`
  - Purpose: 验证 GetProductListByOrder 接口返回的 is_takeout_bill 和 is_takeout_order 字段正确
  - Requirements: Requirement 1.1, 2.1
  - Test Cases:
    - 传统外送订单：`is_takeout_bill = true`, `is_takeout_order = false`
    - 第三方平台外卖订单（Grab）：`is_takeout_bill = false`, `is_takeout_order = true`
    - 第三方平台外卖订单（LINE MAN）：`is_takeout_bill = false`, `is_takeout_order = true`
    - 堂食订单：`is_takeout_bill = false`, `is_takeout_order = false`
  - Success: 所有测试用例通过，字段值正确

- [ ] 3.6 验证 `GetProductListByCategory` 返回的字段正确

  - File: `main/app/service/production.go`
  - Purpose: 验证 GetProductListByCategory 接口返回的 is_takeout_bill 和 is_takeout_order 字段正确
  - Requirements: Requirement 2.1
  - Test Cases: 同 Task 3.5
  - Success: 所有测试用例通过，字段值正确

- [ ] 3.7 编写单元测试（测试 `IsTakeoutOrder()` 方法）

  - File: `main/app/model/production_test.go`
  - Purpose: 为 IsTakeoutOrder() 方法编写单元测试
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有测试文件结构，参考 SaleBill.IsTakeoutBill() 测试
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 IsTakeoutOrder() 方法编写单元测试 | Context: 测试 takeout_order_uuid = 0 时返回 false，takeout_order_uuid > 0 时返回 true | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 80% | Success: 测试完成，覆盖率达标，所有测试通过
  - Test Cases:
    - `takeout_order_uuid = 0`: 返回 `false`
    - `takeout_order_uuid > 0`: 返回 `true`

- [ ] 3.8 编写单元测试（测试外卖订单判断逻辑）

  - File: `main/app/service/production_test.go`
  - Purpose: 为 groupByOrder 和 groupByCategory 方法中的外卖订单判断逻辑编写单元测试
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有测试文件结构
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为外卖订单判断逻辑编写单元测试 | Context: 测试传统外送订单和第三方平台外卖订单的判断逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试完成，覆盖率达标，所有测试通过

- [ ] 3.9 编写 API 集成测试（测试 Grab/LINE MAN 订单在厨显中的显示）

  - File: `main/test/integration/kitchen_production_test.go`（或新建文件）
  - Purpose: 编写端到端集成测试，验证 Grab/LINE MAN 订单在厨显中的显示
  - Requirements: Requirement 1.1, 2.1
  - Leverage: 现有集成测试结构
  - Prompt: Role: QA Automation Engineer | Task: 实现 Grab/LINE MAN 订单在厨显中的集成测试 | Context: 测试完整业务流程，创建外卖订单、接单、查询厨显接口，验证返回的字段正确 | Restrictions: 使用真实数据库，测试真实集成场景 | Success: 集成测试通过，端到端流程验证成功
  - Test Cases:
    - 创建 Grab 订单，接单后调用 `/api/v1/kitchen/product/list_by_order`，验证 `is_takeout_order = true`
    - 创建 LINE MAN 订单，接单后调用 `/api/v1/kitchen/product/list_by_category`，验证 `is_takeout_order = true`

---

## Phase 4: 前端实现（2-3 天）

> **注意**: 前端实现在前端仓库 `all-kds-grab-order-display` 中完成

- [ ] 4.1 前端根据 `is_takeout_bill` 和 `is_takeout_order` 字段显示相应标识

  - File: 前端仓库相关文件
  - Purpose: 根据后端返回的字段显示相应的标识
  - Requirements: Requirement 1.1, 2.1
  - Logic:
    - `is_takeout_bill = true`: 显示"外送"标识（传统店内外送）
    - `is_takeout_order = true`: 显示"外卖"标识（第三方平台外卖：Grab/LINE MAN）
  - Success: 标识显示正确，逻辑清晰

- [ ] 4.2 按订单显示模式：在订单列标题或订单卡片上显示相应的标识

  - File: 前端仓库相关文件
  - Purpose: 在按订单显示模式下，在订单列标题或订单卡片上显示外送/外卖标识
  - Requirements: Requirement 1.1
  - Success: 标识显示位置合理，易于识别

- [ ] 4.3 按分类显示模式：在商品卡片上显示相应的标识

  - File: 前端仓库相关文件
  - Purpose: 在按分类显示模式下，在商品卡片上显示外送/外卖标识
  - Requirements: Requirement 2.1
  - Success: 标识显示位置合理，易于识别

- [ ] 4.4 标识的视觉样式设计（颜色、图标等）

  - File: 前端仓库相关文件
  - Purpose: 设计外送和外卖标识的视觉样式
  - Requirements: Requirement 1.1, 2.1
  - Design:
    - 外送订单：使用通用外送标识
    - 外卖订单：使用第三方平台标识（Grab/LINE MAN）
  - Success: 视觉样式清晰醒目，易于区分

- [ ] 4.5 前端测试

  - File: 前端仓库相关文件
  - Purpose: 编写前端测试，验证标识显示正确
  - Requirements: Requirement 1.1, 2.1
  - Success: 前端测试通过，标识显示正确

---

## Phase 5: 集成测试和优化（1 天）

- [ ] 5.1 端到端集成测试

  - File: 测试文件
  - Purpose: 执行端到端集成测试，验证完整业务流程
  - Requirements: 所有功能需求
  - Test Scenarios:
    - 创建 Grab 订单 → 接单 → 查询厨显接口 → 验证标识显示
    - 创建 LINE MAN 订单 → 接单 → 查询厨显接口 → 验证标识显示
    - 创建传统外送订单 → 查询厨显接口 → 验证标识显示
  - Success: 所有测试场景通过

- [ ] 5.2 性能测试

  - File: 测试文件
  - Purpose: 测试厨显接口的性能，确保查询性能不受影响
  - Requirements: 性能优化
  - Metrics:
    - 本地响应时间: < 200ms（目标）
    - 数据库查询: < 50ms（目标）
    - 并发能力: 1000+ QPS（目标）
  - Success: 性能指标达标

- [ ] 5.3 Bug 修复

  - File: 相关文件
  - Purpose: 修复测试过程中发现的所有 Bug
  - Requirements: 所有功能需求
  - Success: 所有 Bug 已修复

- [ ] 5.4 文档更新

  - File: `docs/shared/api/kitchen_production.md`（如存在）
  - Purpose: 更新 API 文档，说明新增的 `is_takeout_order` 字段
  - Requirements: 文档规范
  - Success: 文档更新完成，说明清晰

---

## 代码优化记录

> 在执行任务过程中若发现优化点，请在此记录。

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Model: ≥ 80%
  - Service: ≥ 70%
- [ ] 所有测试通过
- [ ] 数据库迁移脚本可重复执行（idempotent）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有）
- [ ] 数据库文档已更新
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-kds-grab-line-man-display/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-kds-grab-line-man-display/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-kds-grab-line-man-display/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-kds-grab-line-man-display/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-kds-grab-line-man-display/tasks.md)" | bc
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

### Go Main 后端开发

```
Role: Go Developer with Gin expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/database.mdc

Restrictions:
- 使用 Go 1.23+
- 遵循 Go Main 三层架构（API → Service → Repository）
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- URL 使用 snake_case

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

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
- 在执行任务过程中若总结出经验或踩坑总结，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-25  
**维护者**: weifashi

