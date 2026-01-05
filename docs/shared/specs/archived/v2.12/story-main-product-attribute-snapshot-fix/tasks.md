# 商品属性信息快照修复 任务分解

> 本文档定义商品属性信息快照修复功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 23  
**已完成**: 21  
**进行中**: -  
**完成率**: 91%

---

## Phase 1: 数据库迁移

### 1.1 创建数据库迁移文件 - 修改字段类型为 TEXT

- [x] 1.1 创建数据库迁移文件 - ttpos_sale_order_product.name

  - File: `admin/database/migrations/20251209094516_modify_sale_order_product_name_to_text.php`
  - Purpose: 将 `ttpos_sale_order_product.name` 字段类型从 VARCHAR(255) 改为 TEXT
  - Requirements: Requirement 1
  - Leverage: 现有迁移文件: `admin/database/migrations/20251205090001_modify_sign_to_nullable_in_sale_order_product.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，将 ttpos_sale_order_product.name 字段类型改为 TEXT | Context: 字段注释更新为"商品名称快照（JSON），不随后台更新"，迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行（幂等性） | Success: 迁移文件创建成功，字段类型修改正确

- [x] 1.2 创建数据库迁移文件 - ttpos_sale_order_product.flavor_name

  - File: `admin/database/migrations/20251209094517_modify_sale_order_product_flavor_name_to_text.php`
  - Purpose: 将 `ttpos_sale_order_product.flavor_name` 字段类型从 VARCHAR(255) 改为 TEXT（如果当前不是 TEXT）
  - Requirements: Requirement 2
  - Leverage: 现有迁移文件: `admin/database/migrations/20251205090001_modify_sign_to_nullable_in_sale_order_product.php`（注意：该文件已修改 flavor_name 为 text，需要检查）
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，将 ttpos_sale_order_product.flavor_name 字段类型改为 TEXT（如果当前不是 TEXT） | Context: 字段注释更新为"规格名称快照（JSON），不随后台更新"，迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行（幂等性） | Success: 迁移文件创建成功，字段类型修改正确

- [x] 1.3 创建数据库迁移文件 - ttpos_sale_order_product_bom.name

  - File: `admin/database/migrations/20251209094518_modify_sale_order_product_bom_name_to_text.php`
  - Purpose: 将 `ttpos_sale_order_product_bom.name` 字段类型从 VARCHAR(255) 改为 TEXT
  - Requirements: Requirement 3
  - Leverage: 现有迁移文件: `admin/database/migrations/20251205090001_modify_sign_to_nullable_in_sale_order_product.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，将 ttpos_sale_order_product_bom.name 字段类型改为 TEXT | Context: 字段注释更新为"规格或小料名称快照（JSON），不随后台更新"，迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行（幂等性） | Success: 迁移文件创建成功，字段类型修改正确

- [x] 1.4 创建数据库迁移文件 - ttpos_sale_order_product_attribute.name

  - File: `admin/database/migrations/20251209094519_modify_sale_order_product_attribute_name_to_text.php`
  - Purpose: 将 `ttpos_sale_order_product_attribute.name` 字段类型从 VARCHAR(255) 改为 TEXT
  - Requirements: Requirement 4
  - Leverage: 现有迁移文件: `admin/database/migrations/20251205090001_modify_sign_to_nullable_in_sale_order_product.php`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，将 ttpos_sale_order_product_attribute.name 字段类型改为 TEXT | Context: 字段注释更新为"商品属性名称快照（JSON），不随后台更新"，迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行（幂等性） | Success: 迁移文件创建成功，字段类型修改正确

- [ ] 1.5 执行数据库迁移（测试环境）

  - File: -
  - Purpose: 在测试环境执行迁移，验证字段类型修改
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: Task 1.1, 1.2, 1.3, 1.4 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段类型已修改为 TEXT

---

## Phase 2: Go Model 修改

### 2.1 修改 SaleOrderProduct 模型字段类型

- [x] 2.1 修改 SaleOrderProduct.Name 字段类型

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 将 `Name` 字段的 gorm 标签类型从 `varchar(255)` 改为 `text`
  - Requirements: Requirement 1
  - Leverage: 现有字段定义: `main/app/model/sale_order_product.go:28`
  - Prompt: Role: Go Developer | Task: 修改 SaleOrderProduct.Name 字段的 gorm 标签，将类型改为 text，注释更新为"商品名称快照（JSON），不随后台更新" | Context: 字段定义在 sale_order_product.go 第 28 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段类型修改正确，注释更新正确

- [x] 2.2 修改 SaleOrderProduct.FlavorName 字段类型

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 将 `FlavorName` 字段的 gorm 标签类型从 `varchar(255)` 改为 `text`
  - Requirements: Requirement 2
  - Leverage: 现有字段定义: `main/app/model/sale_order_product.go:29`
  - Prompt: Role: Go Developer | Task: 修改 SaleOrderProduct.FlavorName 字段的 gorm 标签，将类型改为 text，注释更新为"规格名称快照（JSON），不随后台更新" | Context: 字段定义在 sale_order_product.go 第 29 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段类型修改正确，注释更新正确

- [x] 2.3 修改 SaleOrderProductBom.Name 字段类型

  - File: `main/app/model/order.go`
  - Purpose: 将 `SaleOrderProductBom.Name` 字段的 gorm 标签类型从 `varchar(255)` 改为 `text`
  - Requirements: Requirement 3
  - Leverage: 现有字段定义: `main/app/model/order.go:94`
  - Prompt: Role: Go Developer | Task: 修改 SaleOrderProductBom.Name 字段的 gorm 标签，将类型改为 text，注释更新为"规格或小料名称快照（JSON），不随后台更新" | Context: 字段定义在 order.go 第 94 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段类型修改正确，注释更新正确

- [x] 2.4 修改 SaleOrderProductAttribute.Name 字段类型

  - File: `main/app/model/order.go`
  - Purpose: 将 `SaleOrderProductAttribute.Name` 字段的 gorm 标签类型从 `varchar(255)` 改为 `text`
  - Requirements: Requirement 4
  - Leverage: 现有字段定义: `main/app/model/order.go:63`
  - Prompt: Role: Go Developer | Task: 修改 SaleOrderProductAttribute.Name 字段的 gorm 标签，将类型改为 text，注释更新为"商品属性名称快照（JSON），不随后台更新" | Context: 字段定义在 order.go 第 63 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段类型修改正确，注释更新正确

### 2.2 添加快照方法

- [x] 2.5 添加 SaleOrderProduct.GetLocaleName() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 添加获取商品名称（多语言）的方法，优先使用快照字段
  - Requirements: Requirement 1
  - Leverage: 参考实现: `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProduct 中添加 GetLocaleName() 方法，优先使用 Name 快照字段（JSON），降级使用 MultiLanguageName 关联表 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，使用 json.Unmarshal 解析快照字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，优先使用快照字段，降级逻辑正确

- [x] 2.6 添加 SaleOrderProduct.SetNameSnapshot() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 添加设置商品名称快照（JSON）的方法
  - Requirements: Requirement 1
  - Leverage: 参考实现: `main/app/model/sale_bill.go:818` - `SetOrderSourceNameSnapshot()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProduct 中添加 SetNameSnapshot() 方法，从 MultiLanguageName 获取完整多语言数据并序列化为 JSON | Context: 参考 SetOrderSourceNameSnapshot() 的实现模式，使用 json.Marshal 序列化 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，JSON 序列化正确

- [x] 2.7 添加 SaleOrderProduct.GetLocaleFlavorName() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 添加获取规格名称（多语言）的方法，优先使用快照字段
  - Requirements: Requirement 2
  - Leverage: 参考实现: `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProduct 中添加 GetLocaleFlavorName() 方法，优先使用 FlavorName 快照字段（JSON），降级使用关联表 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，使用 json.Unmarshal 解析快照字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，优先使用快照字段，降级逻辑正确

- [x] 2.8 添加 SaleOrderProductBom.GetLocaleName() 方法

  - File: `main/app/model/order.go`
  - Purpose: 添加获取规格或小料名称（多语言）的方法，优先使用快照字段
  - Requirements: Requirement 3
  - Leverage: 参考实现: `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProductBom 中添加 GetLocaleName() 方法，优先使用 Name 快照字段（JSON），降级使用关联表（根据 IsFlavor() 判断是规格还是小料） | Context: 参考 GetLocaleOrderSourceName() 的实现模式，使用 json.Unmarshal 解析快照字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，优先使用快照字段，降级逻辑正确

- [x] 2.9 添加 SaleOrderProductBom.SetNameSnapshot() 方法

  - File: `main/app/model/order.go`
  - Purpose: 添加设置规格或小料名称快照（JSON）的方法
  - Requirements: Requirement 3
  - Leverage: 参考实现: `main/app/model/sale_bill.go:818` - `SetOrderSourceNameSnapshot()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProductBom 中添加 SetNameSnapshot() 方法，从 MultiLanguageName 获取完整多语言数据并序列化为 JSON | Context: 参考 SetOrderSourceNameSnapshot() 的实现模式，使用 json.Marshal 序列化 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，JSON 序列化正确

- [x] 2.10 添加 SaleOrderProductAttribute.GetLocaleName() 方法

  - File: `main/app/model/order.go`
  - Purpose: 添加获取属性名称（多语言）的方法，优先使用快照字段
  - Requirements: Requirement 4
  - Leverage: 参考实现: `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProductAttribute 中添加 GetLocaleName() 方法，优先使用 Name 快照字段（JSON），降级使用关联表 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，使用 json.Unmarshal 解析快照字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，优先使用快照字段，降级逻辑正确

- [x] 2.11 添加 SaleOrderProductAttribute.SetNameSnapshot() 方法

  - File: `main/app/model/order.go`
  - Purpose: 添加设置属性名称快照（JSON）的方法
  - Requirements: Requirement 4
  - Leverage: 参考实现: `main/app/model/sale_bill.go:818` - `SetOrderSourceNameSnapshot()`
  - Prompt: Role: Go Developer | Task: 在 SaleOrderProductAttribute 中添加 SetNameSnapshot() 方法，从 MultiLanguageName 获取完整多语言数据并序列化为 JSON | Context: 参考 SetOrderSourceNameSnapshot() 的实现模式，使用 json.Marshal 序列化 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 方法实现正确，JSON 序列化正确

---

## Phase 3: 修改查询逻辑

### 3.1 修改商品名称获取方法

- [x] 3.1 修改 GetProductNameAttributes() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改商品名称属性组合方法，使用快照数据
  - Requirements: Requirement 5
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:255` - `GetProductNameAttributes()`
  - Prompt: Role: Go Developer | Task: 修改 GetProductNameAttributes() 方法，使用 GetLocaleName() 和 GetLocaleFlavorName() 获取商品名称和规格名称，使用 GetLocaleName() 获取属性名称 | Context: 现有方法在第 255 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

- [x] 3.2 修改 GetNameAndFlavorName() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改商品名称规格组合方法，使用快照数据
  - Requirements: Requirement 5
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1299` - `GetNameAndFlavorName()`
  - Prompt: Role: Go Developer | Task: 修改 GetNameAndFlavorName() 方法，使用 GetLocaleName() 和 GetLocaleFlavorName() 获取商品名称和规格名称 | Context: 现有方法在第 1299 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

### 3.2 修改规格名称获取方法

- [x] 3.3 修改 GetFlavorName() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改规格名称获取方法，使用快照数据
  - Requirements: Requirement 2
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1482` - `GetFlavorName()`
  - Prompt: Role: Go Developer | Task: 修改 GetFlavorName() 方法，直接调用 GetLocaleFlavorName() | Context: 现有方法在第 1482 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

### 3.3 修改小料名称获取方法

- [x] 3.4 修改 GetSauceNamesList() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改小料名称获取方法，使用快照数据
  - Requirements: Requirement 3
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1496` - `GetSauceNamesList()`
  - Prompt: Role: Go Developer | Task: 修改 GetSauceNamesList() 方法，使用 SaleOrderProductBom.GetLocaleName() 获取小料名称 | Context: 现有方法在第 1496 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

### 3.4 修改属性名称获取方法

- [x] 3.5 修改 GetAttributeNameList() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改属性名称列表获取方法，使用快照数据
  - Requirements: Requirement 4
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1423` - `GetAttributeNameList()`
  - Prompt: Role: Go Developer | Task: 修改 GetAttributeNameList() 方法，使用 GetLocaleFlavorName()、SaleOrderProductBom.GetLocaleName()、SaleOrderProductAttribute.GetLocaleName() 获取规格、小料、属性名称 | Context: 现有方法在第 1423 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

- [x] 3.6 修改 GetPureAttributeNameList() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改纯属性名称列表获取方法，使用快照数据
  - Requirements: Requirement 4
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1468` - `GetPureAttributeNameList()`
  - Prompt: Role: Go Developer | Task: 修改 GetPureAttributeNameList() 方法，使用 SaleOrderProductAttribute.GetLocaleName() 获取属性名称 | Context: 现有方法在第 1468 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

- [x] 3.7 修改 GetAttributeNamesByLang() 方法

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改属性名称获取方法（单语言），使用快照数据
  - Requirements: Requirement 4
  - Leverage: 现有方法: `main/app/model/sale_order_product.go:1545` - `GetAttributeNamesByLangs()`
  - Prompt: Role: Go Developer | Task: 修改 GetAttributeNamesByLangs() 方法，使用 GetLocaleFlavorName()、SaleOrderProductBom.GetLocaleName()、SaleOrderProductAttribute.GetLocaleName() 获取规格、小料、属性名称，然后使用 GetLocale() 获取单语言 | Context: 现有方法在第 1545 行附近，需要替换直接使用关联表的逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，使用快照数据

---

## Phase 4: 修改下单逻辑

### 4.1 修改订单创建逻辑 - 保存快照

- [x] 4.1 修改 NewDefaultSaleOrderProduct 设置商品名称快照

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 在 NewDefaultSaleOrderProduct 中，保存商品名称快照（JSON）
  - Requirements: Requirement 1
  - Leverage: 参考实现: `main/app/service/order_base.go:152` - 自助餐名称快照保存逻辑
  - Prompt: Role: Go Developer | Task: 在 NewDefaultSaleOrderProduct 方法中，创建 SaleOrderProduct 后调用 SetNameSnapshot() 保存商品名称快照 | Context: 方法在 sale_order_product.go 第 1720 行附近，productPackage.MultiLanguageName 已加载 | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录错误日志但不中断流程 | Success: 商品名称快照保存成功

- [x] 4.2 修改 newSaleOrderProduct 设置规格/小料/属性快照

  - File: `main/app/service/order.go`
  - Purpose: 在 newSaleOrderProduct 中，保存规格、小料、属性名称快照（JSON）
  - Requirements: Requirement 2, 3, 4
  - Leverage: 参考实现: `main/app/service/order_base.go:152` - 自助餐名称快照保存逻辑
  - Prompt: Role: Go Developer | Task: 在 newSaleOrderProduct 方法中，创建 SaleOrderProduct 后，遍历 SaleOrderProductBoms 和 SaleOrderProductAttributes，调用 SetNameSnapshot() 保存快照 | Context: 方法在 order.go 第 1648 行附近，flavorProductBom、sauceProductBoms、productAttributes 已加载 | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录错误日志但不中断流程 | Success: 规格、小料、属性名称快照保存成功

- [x] 4.3 修改 EditProduct 设置规格/小料/属性快照

  - File: `main/app/service/order.go`
  - Purpose: 在 EditProduct 中，保存规格、小料、属性名称快照（JSON）
  - Requirements: Requirement 2, 3, 4
  - Leverage: 参考实现: `main/app/service/order_base.go:152` - 自助餐名称快照保存逻辑
  - Prompt: Role: Go Developer | Task: 在 EditProduct 方法中，创建 SaleOrderProductBom 和 SaleOrderProductAttribute 后，调用 SetNameSnapshot() 保存快照 | Context: 方法在 order.go 第 1481 行附近，flavorProductBom、sauceProductBoms、productAttributes 已加载 | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录错误日志但不中断流程 | Success: 规格、小料、属性名称快照保存成功

---

## Phase 5: 测试验证

### 5.1 单元测试

- [ ] 5.1 编写 GetLocaleName() 和 SetNameSnapshot() 单元测试

  - File: `main/app/model/sale_order_product_test.go`
  - Purpose: 测试商品名称快照方法的正确性
  - Requirements: Requirement 1
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Prompt: Role: QA Engineer | Task: 为 GetLocaleName() 和 SetNameSnapshot() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据等场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 5.2 编写 GetLocaleFlavorName() 单元测试

  - File: `main/app/model/sale_order_product_test.go`
  - Purpose: 测试规格名称快照方法的正确性
  - Requirements: Requirement 2
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Prompt: Role: QA Engineer | Task: 为 GetLocaleFlavorName() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据等场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 5.3 编写 SaleOrderProductBom 快照方法单元测试

  - File: `main/app/model/order_test.go`
  - Purpose: 测试规格/小料名称快照方法的正确性
  - Requirements: Requirement 3
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Prompt: Role: QA Engineer | Task: 为 SaleOrderProductBom.GetLocaleName() 和 SetNameSnapshot() 编写单元测试，覆盖规格和小料两种场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 5.4 编写 SaleOrderProductAttribute 快照方法单元测试

  - File: `main/app/model/order_test.go`
  - Purpose: 测试属性名称快照方法的正确性
  - Requirements: Requirement 4
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Prompt: Role: QA Engineer | Task: 为 SaleOrderProductAttribute.GetLocaleName() 和 SetNameSnapshot() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据等场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

### 5.2 集成测试

- [ ] 5.5 编写下单集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试下单时保存快照数据
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Prompt: Role: QA Engineer | Task: 编写下单集成测试，验证创建订单时商品名称、规格名称、小料名称、属性名称快照都正确保存为 JSON | Context: 测试所有下单入口（POS、扫码点餐、外卖等） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有下单场景的快照数据保存正确

- [ ] 5.6 编写查询集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试查询时使用快照数据
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Prompt: Role: QA Engineer | Task: 编写查询集成测试，验证订单查询时优先使用快照数据，后台删除商品/规格/小料/属性后，历史订单仍能正常显示 | Context: 测试订单详情、订单列表、订单打印、订单导出等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询场景的快照逻辑正确

### 5.3 回归测试

- [ ] 5.7 回归测试 - 订单查询接口

  - File: -
  - Purpose: 确保订单查询功能不受影响
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单查询接口回归测试，确保所有订单查询接口正常工作 | Context: 测试订单详情、订单列表、订单搜索等接口 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有订单查询接口测试通过

- [ ] 5.8 回归测试 - 订单打印/导出/报表

  - File: -
  - Purpose: 确保订单打印、导出、报表功能不受影响
  - Requirements: Requirement 1, 2, 3, 4, 5
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单打印、导出、报表回归测试，确保所有功能正常工作 | Context: 测试订单打印、订单导出、订单报表等功能 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有功能测试通过

---

## Phase 6: 数据迁移（可选）

### 6.1 数据检查脚本

- [ ] 6.1 编写数据检查脚本

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_check_product_attribute_snapshot_data.php`
  - Purpose: 检查历史订单的快照字段填充情况
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 现有迁移文件
  - Prompt: Role: Database Engineer | Task: 编写数据检查脚本，统计历史订单中快照字段为空的记录数量 | Context: 检查 name、flavor_name、sale_order_product_bom.name、sale_order_product_attribute.name 字段 | Restrictions: 遵循 .cursor/rules/database.mdc | Success: 检查脚本执行成功，统计数据准确

### 6.2 数据迁移脚本（可选）

- [ ] 6.2 编写数据迁移脚本（可选）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_migrate_product_attribute_snapshot_data.php`
  - Purpose: 补充历史订单的快照字段（从关联表迁移）
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: 参考实现: `docs/shared/specs/active/story-main-order-source-snapshot-fix/design.md` - 数据迁移 SQL
  - Prompt: Role: Database Engineer | Task: 编写数据迁移脚本，从关联表补充历史订单的快照字段（JSON 格式） | Context: 使用 JSON_OBJECT() 函数序列化多语言数据，只迁移关联表数据存在的记录 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行 | Success: 迁移脚本执行成功，历史数据补充正确

- [ ] 6.3 执行数据迁移（可选，测试环境）

  - File: -
  - Purpose: 在测试环境执行数据迁移，验证迁移脚本正确性
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: Task 6.2 的迁移脚本
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，历史数据补充正确

---

## Phase 7: 生产环境部署

### 7.1 生产环境迁移

- [ ] 7.1 执行生产环境迁移

  - File: -
  - Purpose: 在生产环境执行数据库迁移
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: Phase 1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 生产环境迁移执行成功，字段类型已修改

- [ ] 7.2 执行生产环境数据迁移（可选）

  - File: -
  - Purpose: 在生产环境执行数据迁移（如果选择执行）
  - Requirements: Requirement 1, 2, 3, 4
  - Leverage: Task 6.2 的迁移脚本
  - Command: `cd admin && php think migrate:run`
  - Success: 生产环境数据迁移执行成功

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan

