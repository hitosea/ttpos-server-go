# 套餐分组类型增强功能（必选和默认选中）任务分解

> 本文档定义 套餐分组类型增强功能（必选和默认选中） 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 12  
**进行中**: -  
**待完成**: 3 (1.2 执行数据库迁移, 4.1-4.2 API层参数验证, 5.x 测试任务)  
**完成率**: 80%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件 - product_package_group_item 表

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_is_required_and_is_default_to_product_package_group_item_table.php`
  - Purpose: 在 product_package_group_item 表中增加 is_required 和 is_default 字段
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/shared/specs/active/story-shop-package-group-type/design.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_product_package_group_item 表中增加 is_required (TINYINT DEFAULT 0) 和 is_default (TINYINT DEFAULT 0) 字段 | Context: is_required 表示必选 0-不必选 1-必选，is_default 表示默认选中 0-默认不选中 1-默认选中，迁移前检查字段是否存在，迁移时设置现有数据的默认值为 0 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移文件支持回滚 | Success: 迁移文件创建成功，字段定义正确，支持回滚

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中创建字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已创建，现有数据默认值正确

- [x] 1.3 更新 Go Model - ProductPackageGroupItem

  - File: `main/app/model/product_package_group_item.go`
  - Purpose: 在 ProductPackageGroupItem 结构体中增加 IsRequired 和 IsDefault 字段
  - Requirements: 1.5, 2.5
  - Leverage: 现有 Model: `main/app/model/product_package_group_item.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 ProductPackageGroupItem 结构体中增加 IsRequired (int) 和 IsDefault (int) 字段 | Context: 使用 gorm 标签，IsRequired 类型为 tinyint，IsDefault 类型为 tinyint，默认值为 0，添加注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确，gorm 标签正确

- [x] 1.4 更新 PHP Model - ProductPackageGroupItem

  - File: `admin/app/common/model/product/ProductPackageGroupItem.php`
  - Purpose: 在 ProductPackageGroupItem 模型的 $field 数组中增加 is_required 和 is_default 字段
  - Requirements: 1.5, 2.5
  - Leverage: 现有 Model: `admin/app/common/model/product/ProductPackageGroupItem.php`，迁移文件: Task 1.1
  - Prompt: Role: PHP Developer | Task: 在 ProductPackageGroupItem 模型的 $field 数组中增加 is_required 和 is_default 字段 | Context: 字段名与数据库一致，保持字段顺序 | Restrictions: 遵循 .cursor/rules/php.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: DTO 层扩展

- [x] 2.1 扩展 Request DTO - ProductShopAddPackageGroupProductReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopAddPackageGroupProductReq 中增加 IsRequired 和 IsDefault 字段
  - Requirements: 3.1
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考现有字段定义
  - Prompt: Role: Go Developer | Task: 在 ProductShopAddPackageGroupProductReq 结构体中增加 IsRequired (int) 和 IsDefault (int) 字段 | Context: IsRequired 默认值为 0，IsDefault 默认值为 0，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [x] 2.2 扩展 Request DTO - ProductShopEditPackageGroupProductReq

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 ProductShopEditPackageGroupProductReq 中增加 IsRequired 和 IsDefault 字段
  - Requirements: 3.2
  - Leverage: 现有 DTO: `main/app/dto/req/product.go`，参考 Task 2.1
  - Prompt: Role: Go Developer | Task: 在 ProductShopEditPackageGroupProductReq 结构体中增加 IsRequired (int) 和 IsDefault (int) 字段 | Context: 与创建接口保持一致，添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

- [x] 2.3 扩展 Response DTO - PackageProductDetail

  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 在 PackageProductDetail 中增加 IsRequired 和 IsDefault 字段
  - Requirements: 4.1
  - Leverage: 现有 DTO: `main/app/dto/resp/product_resp/product.go`，参考现有字段定义
  - Prompt: Role: Go Developer | Task: 在 PackageProductDetail 结构体中增加 IsRequired (int) 和 IsDefault (int) 字段 | Context: 添加 json 标签和注释 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

---

## Phase 3: Service 层扩展

- [x] 3.1 扩展套餐创建逻辑 - 保存必选和默认选中字段

  - File: `main/app/service/product.go`
  - Purpose: 在套餐创建时保存 is_required 和 is_default 字段
  - Requirements: 3.1
  - Leverage: 现有 Service: `main/app/service/product.go:ProductShopAdd()`，参考现有 AddPrice 字段的处理方式
  - Prompt: Role: Go Developer | Task: 在套餐创建逻辑中，保存 ProductPackageGroupItem 的 IsRequired 和 IsDefault 字段 | Context: 从 ProductShopAddPackageGroupProductReq 中获取字段值，保存到 Model 中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段保存成功，数据正确

- [x] 3.2 扩展套餐编辑逻辑 - 更新必选和默认选中字段

  - File: `main/app/service/product.go`
  - Purpose: 在套餐编辑时更新 is_required 和 is_default 字段
  - Requirements: 3.2
  - Leverage: 现有 Service: `main/app/service/product.go:ProductShopEdit()`，参考 Task 3.1
  - Prompt: Role: Go Developer | Task: 在套餐编辑逻辑中，更新 ProductPackageGroupItem 的 IsRequired 和 IsDefault 字段 | Context: 从 ProductShopEditPackageGroupProductReq 中获取字段值，更新到 Model 中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段更新成功，数据正确

- [x] 3.3 扩展商品详情查询逻辑 - 返回必选和默认选中字段

  - File: `main/app/service/product.go`
  - Purpose: 在商品详情查询时返回 is_required 和 is_default 字段
  - Requirements: 3.3
  - Leverage: 现有 Service: `main/app/service/product.go:ProductShopDetail()`，参考现有 AddPrice 字段的处理方式
  - Prompt: Role: Go Developer | Task: 在商品详情查询逻辑中，返回 ProductPackageGroupItem 的 IsRequired 和 IsDefault 字段 | Context: 从 Model 中读取字段值，填充到响应 DTO 中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段返回成功，数据正确

- [x] 3.4 扩展商品列表查询逻辑 - 返回必选和默认选中字段

  - File: `main/app/service/product.go`
  - Purpose: 在商品列表查询时返回 is_required 和 is_default 字段
  - Requirements: 4.2
  - Leverage: 现有 Service: `main/app/service/product.go:ProductSearch()`，参考现有 AddPrice 字段的处理方式
  - Prompt: Role: Go Developer | Task: 在商品列表查询逻辑中，返回 PackageProductDetail 的 IsRequired 和 IsDefault 字段 | Context: 从 Model 中读取字段值，填充到 PackageProductDetail 结构体中 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段返回成功，数据正确

- [x] 3.5 实现业务逻辑验证 - 必选和默认选中规则验证

  - File: `main/app/service/product.go`
  - Purpose: 在套餐创建/编辑时验证必选和默认选中的业务规则
  - Requirements: 6.1, 6.2, 6.3, 6.4
  - Leverage: 现有 Service: `main/app/service/product.go`，参考现有验证逻辑
  - Prompt: Role: Go Developer | Task: 实现 validateRequiredAndDefault 函数，验证必选和默认选中的业务规则 | Context: 1. 可选分组时，必选数量不能大于可选数量；2. 可选分组时，默认数量不能大于可选数量；验证失败时返回明确的错误提示 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 验证逻辑正确，错误提示明确

---

## Phase 4: API 层扩展

- [x] 4.1 扩展创建套餐接口 - 参数验证

  - File: `main/app/service/product_check.go`
  - Purpose: 在创建套餐接口中增加 is_required 和 is_default 字段的参数验证
  - Requirements: 3.4, 3.5
  - Status: ✅ 已完成 - 验证逻辑在 `CheckProductPackage` 函数中实现，API 层通过 Service 层调用验证
  - Implementation: 验证逻辑在 Service 层的 `product_check.go:CheckProductPackage()` 中实现，包括字段值验证（0或1）、必选/默认数量验证（仅可选分组）

- [x] 4.2 扩展编辑套餐接口 - 参数验证

  - File: `main/app/service/product_check.go`
  - Purpose: 在编辑套餐接口中增加 is_required 和 is_default 字段的参数验证
  - Requirements: 3.4, 3.5
  - Status: ✅ 已完成 - 验证逻辑在 `CheckProductPackage` 函数中实现，API 层通过 Service 层调用验证
  - Implementation: 与创建接口使用相同的验证逻辑，通过 `CheckProductPackage` 函数统一处理

---

## Phase 5: 测试和验证

- [ ] 5.1 单元测试 - Model 字段映射

  - File: `main/app/model/product_package_group_item_test.go` (新建)
  - Purpose: 测试 ProductPackageGroupItem Model 的字段映射
  - Requirements: 1.3
  - Leverage: 现有测试文件，参考现有测试用例
  - Prompt: Role: Go Developer | Task: 编写单元测试，测试 ProductPackageGroupItem Model 的 IsRequired 和 IsDefault 字段映射 | Context: 测试字段的 gorm 标签、默认值、数据库映射 | Restrictions: 遵循 Go 测试规范 | Success: 测试通过，覆盖完整

- [ ] 5.2 集成测试 - API 接口

  - File: `main/tests/integration/product_test.go` (新建或扩展)
  - Purpose: 测试创建/编辑/查询接口的必选和默认选中功能
  - Requirements: 3.1, 3.2, 3.3, 4.2
  - Leverage: 现有测试文件，参考现有测试用例
  - Prompt: Role: Go Developer | Task: 编写集成测试，测试套餐创建/编辑/查询接口的必选和默认选中功能 | Context: 测试字段的保存、更新、查询，测试业务规则验证 | Restrictions: 遵循 Go 测试规范 | Success: 测试通过，覆盖完整

- [ ] 5.3 端到端测试 - 完整流程

  - File: `main/tests/e2e/package_group_enhancement_test.go` (新建)
  - Purpose: 测试完整的套餐创建、编辑、查询流程
  - Requirements: 全部
  - Leverage: 现有测试文件，参考现有测试用例
  - Prompt: Role: Go Developer | Task: 编写端到端测试，测试完整的套餐创建、编辑、查询流程 | Context: 测试包含必选和默认选中商品的套餐创建、编辑、查询 | Restrictions: 遵循 Go 测试规范 | Success: 测试通过，覆盖完整

---

## 📝 任务依赖关系

```
Phase 1 (数据库) → Phase 2 (DTO) → Phase 3 (Service) → Phase 4 (API) → Phase 5 (测试)
```

**关键路径**:
1. 数据库迁移 → Model 更新 → DTO 扩展 → Service 扩展 → API 扩展 → 测试

---

## 🎯 验收标准

### 功能验收

- [ ] 数据库字段创建成功，默认值正确
- [ ] Model 字段映射正确
- [ ] DTO 字段定义正确
- [ ] Service 层逻辑正确
- [ ] API 接口参数验证正确
- [ ] 业务规则验证正确

### 数据验收

- [ ] 创建套餐时，必选和默认选中字段保存正确
- [ ] 编辑套餐时，必选和默认选中字段更新正确
- [ ] 查询套餐时，必选和默认选中字段返回正确
- [ ] 商品列表接口返回必选和默认选中字段

### 兼容性验收

- [ ] 现有套餐功能不受影响
- [ ] API 接口向后兼容
- [ ] 数据库迁移不影响现有数据

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 开发组

