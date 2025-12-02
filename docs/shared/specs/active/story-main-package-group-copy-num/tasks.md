# 套餐分组可选份数支持 任务分解

> 本文档定义套餐分组可选份数支持的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 5  
**进行中**: -  
**完成率**: 33%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/20251127140650_add_copy_num_to_sale_order_product_table.php`
  - Purpose: 在 `ttpos_sale_order_product` 表中添加 `copy_num` 字段
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有迁移文件: `admin/database/migrations/20250127000000_add_client_version_to_sale_bill.php`，参考模板: `.cursor/rules/database.mdc`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 sale_order_product 表中添加 copy_num 字段 | Context: 字段类型 DECIMAL(12,4)，默认值 0，位置在 unit_num 之后，注释：表示该子商品在分组中被选择多少份 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中添加字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 同步更新 shop_01.sql

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 同步更新种子数据文件，确保新商户初始化时包含新字段
  - Requirements: 1.4
  - Leverage: 现有 Seeds: `admin/database/seeds/shop_01.sql`，迁移文件: Task 1.1
  - Prompt: Role: Database Engineer | Task: 在 shop_01.sql 的 ttpos_sale_order_product 表定义中添加 copy_num 字段 | Context: 字段类型 DECIMAL(12,4)，默认值 0，位置在 unit_num 之后，注释：表示该子商品在分组中被选择多少份 | Restrictions: 遵循 .cursor/rules/database.mdc，确保与迁移文件一致 | Success: Seeds 文件更新成功，字段定义正确

---

## Phase 2: 数据模型更新

- [x] 2.1 更新 Go Model - SaleOrderProduct

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 在 SaleOrderProduct 结构体中增加 CopyNum 字段
  - Requirements: 2.1, 2.4
  - Leverage: 现有 Model: `main/app/model/sale_order_product.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProduct 结构体中增加 CopyNum (float64) 字段 | Context: 使用 gorm 标签，类型为 decimal(12,4)，默认值为 0.00，添加注释说明用途，位置在 UnitNum 字段之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确，gorm 标签正确

- [x] 2.2 更新 Request DTO - ProductParams

  - File: `main/app/dto/req/shop_cart.go`
  - Purpose: 在 ProductParams 结构体中增加 CopyNum 字段，用于传递套餐子商品的份数
  - Requirements: 3.1, 3.2
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go`，参考现有字段定义（如 UnitNum）
  - **实际实现**: 代码中使用 `product.Num` 来设置 `CopyNum`，因为 `product.Num` 在套餐子商品的情况下就是该子商品在分组中被选择的份数。无需单独添加 `CopyNum` 字段到 `ProductParams`。
  - Success: DTO 逻辑正确，`product.Num` 正确传递到 `CopyNum`

- [ ] 2.3 更新 PHP Model（如需要）

  - File: `admin/app/{admin|shop}/model/SaleOrderProduct.php`（根据实际路径）
  - Purpose: 在 PHP Model 中添加 copy_num 字段
  - Requirements: 2.2, 2.4
  - Leverage: 现有 PHP Model，迁移文件: Task 1.1
  - Prompt: Role: PHP Developer | Task: 在 SaleOrderProduct Model 中添加 copy_num 字段 | Context: 字段类型 decimal(12,4)，默认值 0 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: PHP Model 更新成功

- [ ] 2.4 更新 BMP Model（如需要）

  - File: `ttpos-bmp/app/ttpos-shop/internal/model/`（entity/do 目录）
  - Purpose: 通过重新生成代码更新 BMP Model
  - Requirements: 2.3, 2.4
  - Leverage: 数据库迁移完成后，使用 GoFrame CLI 重新生成
  - Command: `cd ttpos-bmp/app/ttpos-shop && make dao`
  - Success: BMP Model 重新生成成功，包含 copy_num 字段

---

## Phase 3: 业务逻辑适配

- [x] 3.1 修改订单创建逻辑 - newPackageSubProducts

  - File: `main/app/service/order.go`
  - Purpose: 在创建套餐子商品时，设置 copy_num 字段
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有方法: `main/app/service/order.go` 的 `newPackageSubProducts` 方法，Task 2.1-2.2 的实现
  - **实际实现**:
    - 套餐主商品创建：`order.go:1816` - `CopyNum: product.Num`
    - 套餐子商品创建：`order.go:2229` - `CopyNum: product.Num`
    - 套餐主商品数量计算：`order.go:1843` - 使用 `CopyNum` 计算 `Num`：`saleOrderProduct.Num = decimal.NewFromFloat(saleOrderProduct.GetUnitNum()).Mul(decimal.NewFromFloat(saleOrderProduct.CopyNum)).Round(4).InexactFloat64()`
  - **说明**: 使用 `product.Num` 来设置 `CopyNum`，因为 `product.Num` 在套餐子商品的情况下就是该子商品在分组中被选择的份数
  - Success: 业务逻辑正确，copy_num 字段正确设置

- [ ] 3.2 检查退菜逻辑兼容性

  - File: `main/app/service/order_product.go`
  - Purpose: 检查退菜逻辑，确保 copy_num 字段不影响现有退菜功能
  - Requirements: 3.4
  - Leverage: 现有退菜逻辑: `main/app/service/order_product.go`
  - Prompt: Role: Go Developer | Task: 检查退菜逻辑，确保 copy_num 字段不影响现有功能 | Context: 退菜时创建新的退菜记录，copy_num 字段应该被正确复制 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 退菜逻辑兼容，copy_num 字段正确处理

- [ ] 3.3 检查统计逻辑兼容性

  - File: `main/app/service/`（统计相关文件）
  - Purpose: 检查统计逻辑，确保 copy_num 字段不影响现有统计功能
  - Requirements: 3.5
  - Leverage: 现有统计逻辑
  - Prompt: Role: Go Developer | Task: 检查统计逻辑，确保 copy_num 字段不影响现有统计功能 | Context: 统计查询订单商品时，copy_num 字段应该被正确读取，但不影响统计计算 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 统计逻辑兼容，copy_num 字段不影响统计

---

## Phase 4: API 接口验证

- [ ] 4.1 验证订单详情接口返回 copy_num

  - File: `main/app/api/order_api.go`（或相关文件）
  - Purpose: 验证订单详情接口自动返回 copy_num 字段
  - Requirements: 4.1
  - Leverage: 现有 API: `main/app/api/order_api.go`，Task 2.1 的 Model 更新
  - Prompt: Role: QA Engineer | Task: 验证订单详情接口返回 copy_num 字段 | Context: 由于 Model 中已包含 json tag，响应应该自动包含 copy_num 字段 | Restrictions: 验证响应格式正确 | Success: API 返回 copy_num 字段，字段值正确

- [ ] 4.2 验证订单列表接口返回 copy_num

  - File: `main/app/api/order_api.go`（或相关文件）
  - Purpose: 验证订单列表接口自动返回 copy_num 字段
  - Requirements: 4.2
  - Leverage: 现有 API，Task 2.1 的 Model 更新
  - Prompt: Role: QA Engineer | Task: 验证订单列表接口返回 copy_num 字段 | Context: 订单列表中的商品信息应该包含 copy_num 字段 | Restrictions: 验证响应格式正确 | Success: API 返回 copy_num 字段，字段值正确

---

## Phase 5: 测试

- [ ] 5.1 编写 Service 单元测试

  - File: `main/app/service/order_test.go`（或新建测试文件）
  - Purpose: 测试订单创建时 copy_num 字段正确设置
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有测试: `main/app/service/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为订单创建逻辑编写单元测试，验证 copy_num 字段 | Context: 测试套餐子商品的 copy_num 正确记录，测试普通商品和套餐主商品的 copy_num 为 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，测试覆盖率 ≥ 70% | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 5.2 编写 API 集成测试

  - File: `main/app/api/order_api_test.go`（或相关测试文件）
  - Purpose: 测试订单查询接口返回 copy_num 字段
  - Requirements: 4.1, 4.2
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为订单查询接口编写集成测试，验证 copy_num 字段 | Context: 测试订单详情和订单列表接口返回 copy_num 字段，验证字段值正确 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 5.3 端到端集成测试

  - File: `test/integration/package_group_copy_num_test.go`（或相关文件）
  - Purpose: 测试完整的套餐分组可选份数流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试创建包含套餐分组可选份数的订单，验证 copy_num 字段正确记录，验证订单查询返回 copy_num | Restrictions: 测试真实用户场景 | Success: 集成测试通过

---

## Phase 6: 文档和代码审查

- [ ] 6.1 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范遵循
  - Requirements: 所有需求
  - Leverage: `.cursor/rules/go-main.mdc`, `.cursor/rules/database.mdc`
  - Success: 代码审查通过，符合规范

- [ ] 6.2 更新文档（如有需要）

  - File: `docs/shared/api/`（如有相关 API 文档）
  - Purpose: 更新 API 文档，说明 copy_num 字段
  - Requirements: 4.4
  - Leverage: 现有 API 文档
  - Success: 文档更新完成

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] PHP 代码通过 PSR-2 格式化（如修改）
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Order 相关: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 数据库迁移脚本已创建并执行
- [ ] shop_01.sql 已同步更新
- [ ] API 文档已更新（如有新接口说明）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/go-bmp.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/php.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-package-group-copy-num/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-package-group-copy-num/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-package-group-copy-num/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-package-group-copy-num/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-package-group-copy-num/tasks.md)" | bc
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
**最后更新**: 2025-11-27  
**维护者**: 后端开发组

