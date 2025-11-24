# 套餐分组类型和加价功能 任务分解

> 本文档定义 套餐分组类型和加价功能 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件 - product_package_group 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_group_type_and_optional_count_to_product_package_group_table.php`
  - Purpose: 在 product_package_group 表中增加 group_type 和 optional_count 字段
  - Requirements: 8.1, 8.3, 8.4, 8.5
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_product_package_group 表中增加 group_type (TINYINT DEFAULT 0) 和 optional_count (INT DEFAULT 1) 字段 | Context: group_type 表示分组类型 0-固定 1-可选，optional_count 表示可选数量，迁移前检查字段是否存在，迁移时设置现有数据的默认值 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移文件支持回滚 | Success: 迁移文件创建成功，字段定义正确，支持回滚

- [ ] 1.2 创建数据库迁移文件 - product_package_group_item 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_add_price_to_product_package_group_item_table.php`
  - Purpose: 在 product_package_group_item 表中增加 add_price 字段
  - Requirements: 8.2, 8.3, 8.4, 8.5
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_product_package_group_item 表中增加 add_price (DECIMAL(10,2) DEFAULT 0.00) 字段 | Context: add_price 表示加价金额，迁移前检查字段是否存在，迁移时设置现有数据的默认值为 0.00 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移文件支持回滚 | Success: 迁移文件创建成功，字段定义正确，支持回滚

- [ ] 1.3 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建字段
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1, 1.2 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已创建，现有数据默认值正确

- [ ] 1.4 更新 Go Model - ProductPackageGroup

  - File: `main/app/model/product_package_group.go`
  - Purpose: 在 ProductPackageGroup 结构体中增加 GroupType 和 OptionalCount 字段
  - Requirements: 1.5, 1.6
  - Leverage: 现有 Model: `main/app/model/product_package_group.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 ProductPackageGroup 结构体中增加 GroupType (int) 和 OptionalCount (int) 字段 | Context: 使用 gorm 标签，GroupType 类型为 tinyint，OptionalCount 类型为 int，默认值为 0 和 1，添加注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确，gorm 标签正确

- [ ] 1.5 更新 Go Model - ProductPackageGroupItem

  - File: `main/app/model/product_package_group_item.go`
  - Purpose: 在 ProductPackageGroupItem 结构体中增加 AddPrice 字段
  - Requirements: 3.5
  - Leverage: 现有 Model: `main/app/model/product_package_group_item.go`，迁移文件: Task 1.2
  - Prompt: Role: Go Developer | Task: 在 ProductPackageGroupItem 结构体中增加 AddPrice (float64) 字段 | Context: 使用 gorm 标签，类型为 decimal(10,2)，默认值为 0.00，添加注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确，gorm 标签正确

- [ ] 1.6 更新 PHP Model - ProductPackageGroup

  - File: `admin/app/common/model/product/ProductPackageGroup.php`
  - Purpose: 在 ProductPackageGroup 模型中增加 group_type 和 optional_count 字段映射
  - Requirements: 1.5, 1.6
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroup.php`，迁移文件: Task 1.1
  - Prompt: Role: PHP Developer | Task: 在 ProductPackageGroup 模型的 $field 数组中增加 group_type 和 optional_count 字段 | Context: 字段名与数据库一致，保持字段顺序 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: DTO 层扩展

- [ ] 2.1 扩展 Request DTO - ProductShopAddPackageGroupReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopAddPackageGroupReq 中增加 GroupType 和 OptionalCount 字段
  - Requirements: 7.1
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考现有字段定义
  - Prompt: Role: Go Developer | Task: 在 ProductShopAddPackageGroupReq 结构体中增加 GroupType (int) 和 OptionalCount (int) 字段 | Context: GroupType 默认值为 0，OptionalCount 默认值为 1，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [ ] 2.2 扩展 Request DTO - ProductShopAddPackageGroupProductReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopAddPackageGroupProductReq 中增加 AddPrice 字段
  - Requirements: 7.2
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考现有字段定义
  - Prompt: Role: Go Developer | Task: 在 ProductShopAddPackageGroupProductReq 结构体中增加 AddPrice (float64) 字段 | Context: AddPrice 默认值为 0，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [ ] 2.3 扩展 Request DTO - ProductShopEditPackageGroupReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopEditPackageGroupReq 中增加 GroupType 和 OptionalCount 字段
  - Requirements: 7.3
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考 Task 2.1
  - Prompt: Role: Go Developer | Task: 在 ProductShopEditPackageGroupReq 结构体中增加 GroupType (int) 和 OptionalCount (int) 字段 | Context: 与创建接口保持一致，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [ ] 2.4 扩展 Request DTO - ProductShopEditPackageGroupProductReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopEditPackageGroupProductReq 中增加 AddPrice 字段
  - Requirements: 7.4
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考 Task 2.2
  - Prompt: Role: Go Developer | Task: 在 ProductShopEditPackageGroupProductReq 结构体中增加 AddPrice (float64) 字段 | Context: 与创建接口保持一致，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [ ] 2.5 扩展 Response DTO - ProductPackageGroupResp

  - File: `main/app/dto/resp/product.go`
  - Purpose: 在 ProductPackageGroupResp 中增加 GroupType、OptionalCount 字段，在 ProductPackageGroupItemResp 中增加 AddPrice 字段
  - Requirements: 7.5
  - Leverage: 现有 DTO: `main/app/dto/resp/product.go`，参考现有响应结构
  - Prompt: Role: Go Developer | Task: 在 ProductPackageGroupResp 中增加 GroupType 和 OptionalCount 字段，在 ProductPackageGroupItemResp 中增加 AddPrice 字段 | Context: 字段类型与 Request DTO 一致，添加 json 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc，data 必须是对象 | Success: DTO 更新成功，响应格式正确

---

## Phase 3: Service 层业务逻辑

- [ ] 3.1 实现分组类型验证逻辑

  - File: `main/app/service/product.go`
  - Purpose: 实现分组类型验证，确保 group_type 为 0 或 1
  - Requirements: 1.1, 1.2, 2.1
  - Leverage: 现有 Service: `main/app/service/product.go`，参考现有验证逻辑
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 validatePackageGroupType 方法，验证 group_type 必须为 0（固定）或 1（可选） | Context: 在创建和编辑套餐时调用，验证失败返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑正确，错误提示清晰

- [ ] 3.2 实现可选数量验证逻辑

  - File: `main/app/service/product.go`
  - Purpose: 实现可选数量验证，确保 optional_count >= 1 且 <= 分组内商品总数
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有 Service: `main/app/service/product.go`，参考 Task 3.1
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 validateOptionalCount 方法，验证 optional_count >= 1 且 <= 分组内商品总数，固定分组时自动设置为商品总数 | Context: 在创建和编辑套餐时调用，验证失败返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑正确，固定分组时自动计算可选数量

- [ ] 3.3 实现加价金额验证逻辑

  - File: `main/app/service/product.go`
  - Purpose: 实现加价金额验证，确保 add_price >= 0
  - Requirements: 3.2, 3.4
  - Leverage: 现有 Service: `main/app/service/product.go`，参考 Task 3.1
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 validateAddPrice 方法，验证 add_price >= 0 | Context: 在创建和编辑套餐时调用，验证失败返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑正确，错误提示清晰

- [ ] 3.4 实现必选数量验证逻辑

  - File: `main/app/service/product.go`
  - Purpose: 实现必选数量验证，确保必选数量 <= 可选数量
  - Requirements: 4.4
  - Leverage: 现有 Service: `main/app/service/product.go`，参考 Task 3.2
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 validateRequiredCount 方法，验证必选数量 <= 可选数量且 <= 分组内商品总数 | Context: 在创建和编辑套餐时调用，验证失败返回错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑正确，错误提示清晰

- [ ] 3.5 集成验证逻辑到套餐创建流程

  - File: `main/app/service/product.go`
  - Purpose: 在套餐创建时调用所有验证逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有 Service: `main/app/service/product.go:ProductShopAdd()`，参考 Task 3.1-3.4
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 ProductShopAdd 方法中集成分组类型、可选数量、加价、必选数量验证逻辑 | Context: 在保存数据前进行验证，验证失败返回错误，不保存数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑集成正确，错误处理完善

- [ ] 3.6 集成验证逻辑到套餐编辑流程

  - File: `main/app/service/product.go`
  - Purpose: 在套餐编辑时调用所有验证逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有 Service: `main/app/service/product.go:ProductShopEdit()`，参考 Task 3.5
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 ProductShopEdit 方法中集成分组类型、可选数量、加价、必选数量验证逻辑 | Context: 在更新数据前进行验证，验证失败返回错误，不更新数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 验证逻辑集成正确，错误处理完善

---

## Phase 4: API 层接口更新

- [ ] 4.1 更新创建套餐接口

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在创建套餐接口中支持新字段
  - Requirements: 7.1, 7.2
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go:ProductShopAdd()`，参考 Task 2.1, 2.2
  - Prompt: Role: Go Developer | Task: 更新 ProductShopAdd 接口，支持接收 group_type、optional_count、add_price 字段 | Context: 参数通过 DTO 传递，调用 Service 层验证逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，URL 使用 snake_case | Success: 接口更新成功，参数接收正确

- [ ] 4.2 更新编辑套餐接口

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在编辑套餐接口中支持新字段
  - Requirements: 7.3, 7.4
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go:ProductShopEdit()`，参考 Task 4.1
  - Prompt: Role: Go Developer | Task: 更新 ProductShopEdit 接口，支持接收 group_type、optional_count、add_price 字段 | Context: 参数通过 DTO 传递，调用 Service 层验证逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，URL 使用 snake_case | Success: 接口更新成功，参数接收正确

- [ ] 4.3 更新商品详情接口

  - File: `main/app/api/v1/shop/shop_product.go`
  - Purpose: 在商品详情接口中返回新字段
  - Requirements: 7.5
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_product.go:ProductShopDetail()`，参考 Task 2.5
  - Prompt: Role: Go Developer | Task: 更新 ProductShopDetail 接口，返回数据中包含 group_type、optional_count、add_price 字段 | Context: 从数据库查询数据，组装响应 DTO | Restrictions: 遵循 .cursor/rules/go-main.mdc，data 必须是对象 | Success: 接口更新成功，响应数据包含新字段

---

## Phase 5: 测试

- [ ] 5.1 单元测试 - 验证逻辑

  - File: `main/app/service/product_test.go`
  - Purpose: 测试分组类型、可选数量、加价、必选数量验证逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有测试文件，参考测试模板
  - Prompt: Role: Go Developer | Task: 编写单元测试，测试分组类型、可选数量、加价、必选数量验证逻辑 | Context: 测试正常情况和异常情况，验证错误提示正确 | Restrictions: 遵循 Go 测试规范 | Success: 测试覆盖完整，所有测试通过

- [ ] 5.2 集成测试 - API 接口

  - File: `main/tests/integration/shop_product_test.go`
  - Purpose: 测试创建、编辑、查询套餐接口
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有测试文件，参考测试模板
  - Prompt: Role: Go Developer | Task: 编写集成测试，测试创建、编辑、查询套餐接口，验证新字段功能 | Context: 测试正常情况和异常情况，验证数据保存和返回正确 | Restrictions: 遵循 Go 测试规范 | Success: 测试覆盖完整，所有测试通过

- [ ] 5.3 数据库迁移测试

  - File: `admin/tests/migration_test.php`
  - Purpose: 测试数据库迁移文件
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有测试文件，参考测试模板
  - Prompt: Role: PHP Developer | Task: 编写数据库迁移测试，验证迁移文件执行和回滚 | Context: 测试字段创建、默认值设置、回滚功能 | Restrictions: 遵循 PHP 测试规范 | Success: 测试覆盖完整，所有测试通过

---

## 📝 实现检查清单

### 数据库层
- [ ] 迁移文件创建完成
- [ ] 迁移文件支持回滚
- [ ] 迁移执行成功
- [ ] 现有数据默认值正确

### Model 层
- [ ] Go Model 更新完成
- [ ] PHP Model 更新完成
- [ ] 字段映射正确

### DTO 层
- [ ] Request DTO 扩展完成
- [ ] Response DTO 扩展完成
- [ ] 字段定义正确

### Service 层
- [ ] 验证逻辑实现完成
- [ ] 验证逻辑集成到创建流程
- [ ] 验证逻辑集成到编辑流程

### API 层
- [ ] 创建接口更新完成
- [ ] 编辑接口更新完成
- [ ] 详情接口更新完成

### 测试
- [ ] 单元测试完成
- [ ] 集成测试完成
- [ ] 数据库迁移测试完成

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 开发组

