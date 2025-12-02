# 收银端套餐商品选择功能 任务分解

> 本文档定义收银端套餐商品选择功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 9  
**进行中**: -  
**完成率**: 60%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件 - sale_order_product 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_add_price_to_sale_order_product_table.php`
  - Purpose: 在 sale_order_product 表中增加 add_price 字段
  - Requirements: 6.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/shared/specs/active/story-shop-package-group-type/design.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_sale_order_product 表中增加 add_price (DECIMAL(22,4) DEFAULT 0.00) 字段 | Context: add_price 表示加价金额，子商品记录单商品加价金额，套餐主商品记录所有子商品加价总和，迁移前检查字段是否存在，迁移时设置现有数据的默认值为 0.00 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移文件支持回滚 | Success: 迁移文件创建成功，字段定义正确，支持回滚

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已创建，现有数据默认值正确

- [x] 1.3 更新 Go Model - SaleOrderProduct

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 在 SaleOrderProduct 结构体中增加 AddPrice 字段
  - Requirements: 6.1, 6.2, 6.3
  - Leverage: 现有 Model: `main/app/model/sale_order_product.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProduct 结构体中增加 AddPrice (float64) 字段 | Context: 使用 gorm 标签，类型为 decimal(22,4)，默认值为 0.00，添加注释说明子商品和套餐主商品的用途 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确，gorm 标签正确

---

## Phase 2: DTO 层扩展

- [x] 2.1 扩展 Request DTO - ProductRequest

  - File: `main/app/dto/req/shop_cart.go`
  - Purpose: 在 ProductRequest 结构体中增加 AddPrice 字段
  - Requirements: 7.1
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go`，参考现有字段定义
  - Prompt: Role: Go Developer | Task: 在 ProductRequest 结构体中增加 AddPrice (float64) 字段 | Context: AddPrice 默认值为 0，添加 json 标签 "add_price" 和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [x] 2.2 扩展 Request DTO - ProductParams

  - File: `main/app/dto/req/shop_cart.go`
  - Purpose: 在 ProductParams 结构体中增加 AddPrice 字段
  - Requirements: 7.2
  - Leverage: 现有 DTO: `main/app/dto/req/shop_cart.go`，参考 Task 2.1
  - Prompt: Role: Go Developer | Task: 在 ProductParams 结构体中增加 AddPrice (float64) 字段 | Context: AddPrice 默认值为 0，添加 json 标签 "add_price" 和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

---

## Phase 3: 业务逻辑实现

- [x] 3.1 实现分组选择验证逻辑

  - File: `main/app/service/order_product.go`
  - Purpose: 在 OrderCartProductPackageAdd 方法中增加分组选择验证
  - Requirements: 8.1, 8.2, 8.3, 8.4, 8.5
  - Leverage: 现有方法: `main/app/service/order_product.go:OrderCartProductPackageAdd()`，参考设计文档分组验证逻辑
  - Prompt: Role: Go Developer | Task: 在 OrderCartProductPackageAdd 方法中增加分组选择验证逻辑 | Context: 查询套餐分组配置，验证固定分组是否包含所有商品，验证可选分组已选数量是否等于 optional_count，验证失败时返回明确的错误提示 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑实现成功，错误提示明确，测试通过

- [x] 3.2 实现加价参数传递

  - File: `main/app/service/order_product.go`
  - Purpose: 在构建子商品参数时传递加价金额
  - Requirements: 7.3
  - Leverage: 现有方法: `main/app/service/order_product.go:OrderCartProductPackageAdd()`，参考设计文档加价参数传递
  - Prompt: Role: Go Developer | Task: 在构建 subProduct 时，从 productReq.AddPrice 获取加价金额并传递给 ProductParams | Context: 在构建 subProducts 的循环中，设置 subProduct.AddPrice = productReq.AddPrice | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 加价参数正确传递

- [x] 3.3 实现价格计算逻辑 - 子商品

  - File: `main/app/service/order.go`
  - Purpose: 在创建子商品时保存加价金额
  - Requirements: 6.2, 7.4
  - Leverage: 现有方法: `main/app/service/order.go:newSaleOrderProductForPackageSubProduct()`，参考设计文档价格计算
  - Prompt: Role: Go Developer | Task: 在创建子商品时，从 product.AddPrice 获取加价金额并保存到 saleOrderProduct.AddPrice | Context: 在 newSaleOrderProductForPackageSubProduct 方法中，设置 saleOrderProduct.AddPrice = product.AddPrice | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 子商品加价金额正确保存

- [x] 3.4 实现价格计算逻辑 - 套餐主商品

  - File: `main/app/service/order.go`
  - Purpose: 在创建套餐主商品时计算所有子商品加价总和
  - Requirements: 6.3, 6.4
  - Leverage: 现有方法: `main/app/service/order.go:newSaleOrderProduct()`，参考设计文档价格计算
  - Prompt: Role: Go Developer | Task: 在创建套餐主商品时，计算所有子商品的加价总和并保存到 saleOrderProduct.AddPrice | Context: 在 newSaleOrderProduct 方法中，当 ProductType 为套餐时，遍历 subProducts 计算 totalAddPrice = Σ(subProduct.AddPrice × subProduct.Num)，设置 saleOrderProduct.AddPrice = totalAddPrice | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 套餐主商品加价总和正确计算和保存

- [x] 3.5 实现套餐单价计算

  - File: `main/app/service/order.go`
  - Purpose: 在计算商品价格时考虑加价金额
  - Requirements: 6.4, 6.5
  - Leverage: 现有方法: `main/app/service/order.go:CalcSaleOrderProduct()`，参考设计文档价格计算
  - Prompt: Role: Go Developer | Task: 在计算套餐价格时，将加价金额加入到套餐单价计算中 | Context: 套餐单价 = 套餐原始定价 + AddPrice，在 CalcSaleOrderProduct 方法中，如果是套餐商品，将 AddPrice 加入到价格计算 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 套餐单价计算正确，包含加价金额

---

## Phase 4: API 接口扩展

- [ ] 4.1 扩展购物车详情接口 - 返回加价信息

  - File: `main/app/service/order_product.go`
  - Purpose: 在购物车详情接口中返回套餐商品的加价信息
  - Requirements: 6.5
  - Leverage: 现有方法: `main/app/service/order_product.go:GetOrderCartInfo()`，参考响应结构
  - Prompt: Role: Go Developer | Task: 在购物车详情接口的响应中，为套餐商品和子商品增加 add_price 字段 | Context: 在构建购物车商品响应时，设置 add_price 字段，子商品返回单商品加价，套餐主商品返回加价总和 | Restrictions: 遵循 .cursor/rules/go-main.mdc，响应格式统一 | Success: 购物车详情接口返回正确的加价信息

---

## Phase 5: 测试和验证

- [ ] 5.1 单元测试 - 分组选择验证

  - File: `main/app/service/order_product_test.go`
  - Purpose: 编写分组选择验证的单元测试
  - Requirements: 8.1, 8.2, 8.3
  - Leverage: 现有测试文件，参考测试要点
  - Prompt: Role: Go Developer | Task: 编写分组选择验证的单元测试，覆盖固定分组和可选分组的各种场景 | Context: 测试固定分组选择所有商品（通过）、缺少商品（失败），测试可选分组已选数量等于可选数量（通过）、不等于可选数量（失败） | Restrictions: 遵循 Go 测试规范 | Success: 单元测试通过，覆盖所有场景

- [ ] 5.2 单元测试 - 加价参数传递

  - File: `main/app/service/order_product_test.go`
  - Purpose: 编写加价参数传递的单元测试
  - Requirements: 7.3, 7.4
  - Leverage: 现有测试文件
  - Prompt: Role: Go Developer | Task: 编写加价参数传递的单元测试，验证加价参数正确传递和保存 | Context: 测试加价参数从请求传递到商品创建流程，验证数据库中的加价金额正确 | Restrictions: 遵循 Go 测试规范 | Success: 单元测试通过

- [ ] 5.3 单元测试 - 价格计算

  - File: `main/app/service/order_test.go`
  - Purpose: 编写价格计算的单元测试
  - Requirements: 6.4, 6.5
  - Leverage: 现有测试文件
  - Prompt: Role: Go Developer | Task: 编写价格计算的单元测试，验证套餐单价计算正确 | Context: 测试套餐单价 = 套餐原始定价 + 加价总和，验证子商品和套餐主商品的加价金额正确 | Restrictions: 遵循 Go 测试规范 | Success: 单元测试通过

- [ ] 5.4 集成测试 - 套餐加购流程

  - File: `main/app/api/v1/cashier/cashier_desk_test.go`
  - Purpose: 编写套餐加购的集成测试
  - Requirements: AC6, AC7, AC8
  - Leverage: 现有测试文件
  - Prompt: Role: Go Developer | Task: 编写套餐加购的集成测试，覆盖固定分组、可选分组、包含加价的场景 | Context: 测试固定分组套餐加购、可选分组套餐加购、包含加价的套餐加购，验证购物车详情接口返回正确的价格 | Restrictions: 遵循 Go 测试规范 | Success: 集成测试通过

- [ ] 5.5 手动测试 - 前端界面

  - File: -
  - Purpose: 手动测试前端各终端的套餐选择界面
  - Requirements: AC1, AC2, AC3, AC4, AC5
  - Leverage: 前端代码
  - Steps:
    1. 测试固定分组：验证所有商品默认选中，不可取消
    2. 测试可选分组：验证选择状态显示、步进器功能
    3. 测试加价显示：验证只有加价 > 0 的商品显示价格
    4. 测试选择验证：验证分组已选满和未选满的提示
  - Success: 前端界面功能正常，交互流畅

---

## Phase 6: 文档更新

- [ ] 6.1 更新 API 文档

  - File: `docs/shared/api/package-add-to-cart-business-logic.md`
  - Purpose: 更新套餐加购业务逻辑文档，增加加价参数和分组验证说明
  - Requirements: -
  - Leverage: 现有文档
  - Prompt: Role: Technical Writer | Task: 更新套餐加购业务逻辑文档，增加加价参数传递和分组选择验证的说明 | Context: 在文档中增加加价参数传递流程、分组选择验证逻辑、价格计算逻辑的说明 | Restrictions: 保持文档格式一致 | Success: 文档更新完成，内容准确

- [ ] 6.2 更新前端对接文档

  - File: `docs/shared/api/frontend-changes-package-group-type.md`
  - Purpose: 更新前端对接文档，增加加购接口参数说明
  - Requirements: -
  - Leverage: 现有文档
  - Prompt: Role: Technical Writer | Task: 更新前端对接文档，增加加购接口的 add_price 参数说明和分组选择验证说明 | Context: 在文档中增加加购接口的请求参数变更、分组选择验证规则、错误提示说明 | Restrictions: 保持文档格式一致 | Success: 文档更新完成，内容准确

---

## 📝 执行提示

### 开发顺序建议

1. **Phase 1**: 数据库迁移和 Model 更新（基础）
2. **Phase 2**: DTO 扩展（接口层）
3. **Phase 3**: 业务逻辑实现（核心功能）
4. **Phase 4**: API 接口扩展（接口层）
5. **Phase 5**: 测试和验证（质量保证）
6. **Phase 6**: 文档更新（文档完善）

### 关键注意事项

1. **数据库迁移**: 迁移前检查字段是否存在，避免重复添加
2. **分组验证**: 验证逻辑应在构建商品参数之前执行
3. **价格计算**: 确保加价金额正确累加，避免重复计算
4. **错误提示**: 错误信息要明确，包含分组名称和具体错误原因
5. **测试覆盖**: 确保覆盖所有场景，包括边界情况

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 技术组

