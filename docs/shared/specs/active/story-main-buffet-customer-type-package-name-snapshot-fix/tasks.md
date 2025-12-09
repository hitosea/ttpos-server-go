# 自助餐顾客类型套餐名称快照修复 任务分解

> 本文档定义自助餐顾客类型套餐名称快照修复功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 13  
**进行中**: -  
**完成率**: 72%

---

## Phase 1: 数据库迁移

### 1.1 创建数据库迁移文件

- [x] 1.1 创建数据库迁移文件 - 添加 buffet_package_name 字段

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_buffet_package_name_to_sale_order_buffet_customer_type.php`
  - Purpose: 在 `ttpos_sale_order_buffet_customer_type` 表添加 `buffet_package_name` 快照字段（JSON 格式，包含所有语言）
  - Requirements: Requirement 1（数据库结构变更）
  - Leverage: 现有迁移文件: `admin/database/migrations/20251208152048_add_order_source_name_to_sale_bill.php`
  - SQL:
    ```sql
    ALTER TABLE `ttpos_sale_order_buffet_customer_type` 
    ADD COLUMN `buffet_package_name` TEXT NOT NULL DEFAULT '' 
    COMMENT '自助餐套餐名称快照（JSON），不随后台更新' 
    AFTER `buffet_package_uuid`;
    ```
  - Note: **JSON 方案** - 快照包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_sale_order_buffet_customer_type 表添加 buffet_package_name 字段 | Context: 字段类型为 TEXT，默认值为空字符串，注释说明为快照字段（JSON 格式），迁移脚本需要检查字段是否已存在，如果已存在则跳过（幂等性） | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行 | Success: 迁移文件创建成功，字段类型为 TEXT，注释正确
  - Success: 迁移文件创建成功，字段类型为 TEXT，注释正确

- [ ] 1.2 执行数据库迁移（测试环境）

  - File: -
  - Purpose: 在测试环境数据库中添加 `buffet_package_name` 字段
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加，测试环境验证通过

---

## Phase 2: Go Model 修改

### 2.1 修改 SaleOrderBuffetCustomerType 模型字段

- [x] 2.1 添加 BuffetPackageName 字段到 SaleOrderBuffetCustomerType 结构体

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 在 `SaleOrderBuffetCustomerType` 结构体添加 `BuffetPackageName` 字段，存储 JSON 格式快照
  - Requirements: Requirement 2（数据模型修改）
  - Leverage: 现有字段定义: `main/app/model/sale_order_buffet_customer_type.go:37` - `BuffetPackageUuid` 字段
  - Code:
    ```go
    BuffetPackageUuid           uint64 `gorm:"column:buffet_package_uuid;comment:自助餐套餐ID" json:"buffet_package_uuid"`
    BuffetPackageName           string `gorm:"column:buffet_package_name;type:text" json:"buffet_package_name"` // 新增快照字段（JSON）
    BuffetCustomerTypePriceUuid uint64 `gorm:"column:buffet_customer_type_price_uuid;comment:顾客类型定价ID" json:"buffet_customer_type_price_uuid"`
    ```
  - Position: 紧跟 `BuffetPackageUuid` 字段之后
  - Note: **JSON 方案** - 字段存储完整多语言 JSON
  - Prompt: Role: Go Developer | Task: 在 SaleOrderBuffetCustomerType 结构体添加 BuffetPackageName 字段，gorm 标签类型为 text，json 标签为 buffet_package_name | Context: 字段定义在 sale_order_buffet_customer_type.go 第 37 行附近，紧跟 BuffetPackageUuid 字段之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段添加成功，GORM 和 JSON 标签正确，编译通过
  - Success: 字段添加成功，GORM 和 JSON 标签正确，编译通过

### 2.2 实现快照方法

- [x] 2.2 实现 GetLocaleBuffetPackageName() 方法（JSON 方案）

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 实现自助餐套餐名称获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 3（查询逻辑修改）
  - Leverage: 参考 `GetLocaleOrderSourceName()` 方法的实现（`main/app/model/sale_bill.go:789`）
  - Key Logic (JSON 方案):
    1. 优先使用 `BuffetPackageName` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `BuffetPackage.MultiLanguageName`
    4. 都为空时返回空响应
  - Code Reference:
    ```go
    // 参考 GetLocaleOrderSourceName() 的实现
    func (model *SaleOrderBuffetCustomerType) GetLocaleBuffetPackageName() dto.LocaleResponse {
        // 优先使用快照字段
        snapshotName := model.BuffetPackageName

        // 如果快照字段不为空，尝试反序列化为多语言数据
        if snapshotName != "" {
            var snapshotLocale dto.LocaleResponse
            if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
                if !snapshotLocale.IsNull() {
                    return snapshotLocale
                }
            }
        }

        // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
        if !model.BuffetPackage.MultiLanguageName.IsNullName() {
            return model.BuffetPackage.MultiLanguageName.GetNames()
        }

        return dto.LocaleResponse{}
    }
    ```
  - Import: 需要添加 `encoding/json` 和 `ttpos-server-go/app/dto` 导入
  - Prompt: Role: Go Developer | Task: 实现 GetLocaleBuffetPackageName() 方法，优先使用快照字段，降级使用关联表数据 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，快照字段为 JSON 格式，需要反序列化为 LocaleResponse | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确，逻辑完整，编译通过
  - Success: 方法实现正确，逻辑完整，编译通过

- [x] 2.3 实现 SetBuffetPackageNameSnapshot() 方法（JSON 方案）

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 实现快照保存方法，从 MultiLanguageName 序列化为 JSON
  - Requirements: Requirement 4（下单逻辑修改）
  - Leverage: 参考 `SetOrderSourceNameSnapshot()` 方法的实现（`main/app/model/sale_bill.go:815`）
  - Key Logic (JSON 方案):
    1. 如果 `MultiLanguageName` 为空，设置为空字符串
    2. 从 `MultiLanguageName` 获取完整多语言数据（`GetNames()`）
    3. 序列化为 JSON 字符串
    4. 保存到 `BuffetPackageName` 字段
  - Code Reference:
    ```go
    // 参考 SetOrderSourceNameSnapshot() 的实现
    func (model *SaleOrderBuffetCustomerType) SetBuffetPackageNameSnapshot(multiLangName MultiLanguageName) error {
        // 如果多语言名称为空，设置为空字符串
        if multiLangName.IsNullName() {
            model.BuffetPackageName = ""
            return nil
        }

        // 构建 LocaleResponse
        localeResp := multiLangName.GetNames()

        // 序列化为 JSON
        jsonData, err := json.Marshal(localeResp)
        if err != nil {
            return err
        }

        model.BuffetPackageName = string(jsonData)
        return nil
    }
    ```
  - Import: 需要添加 `encoding/json` 导入
  - Prompt: Role: Go Developer | Task: 实现 SetBuffetPackageNameSnapshot() 方法，从 MultiLanguageName 序列化为 JSON 保存 | Context: 参考 SetOrderSourceNameSnapshot() 的实现模式，参数为 MultiLanguageName（值类型），需要处理空值情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确，序列化逻辑完整，编译通过
  - Success: 方法实现正确，序列化逻辑完整，编译通过

---

## Phase 3: 查询逻辑修改

### 3.1 替换现有查询方法

- [x] 3.1 修改 GetOrderInfos() 方法 - 使用 GetLocaleBuffetPackageName()

  - File: `main/app/service/order_manage.go`
  - Purpose: 替换 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleBuffetPackageName()`
  - Requirements: Requirement 5（查询逻辑修改）
  - Leverage: 现有代码: `main/app/service/order_manage.go:459` - `GetOrderInfos()` 方法
  - Key Logic:
    ```go
    // 修改前
    buffetLocaleName := saleBill.GetLocaleBuffetPackageNameByUuid(
        orderBuffetCustomer.BuffetPackageUuid,
        orderBuffetCustomer.BuffetPackage.MultiLanguageName,
    )
    
    // 修改后
    buffetLocaleName := orderBuffetCustomer.GetLocaleBuffetPackageName()
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetOrderInfos() 方法，使用 orderBuffetCustomer.GetLocaleBuffetPackageName() 替换 SaleBill.GetLocaleBuffetPackageNameByUuid() 调用 | Context: 方法在 order_manage.go 第 459 行附近，需要替换自助餐名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [x] 3.2 修改 checkBuffetCustomerTypePriceChanged() 方法 - 使用 GetLocaleBuffetPackageName()

  - File: `main/app/service/order.go`
  - Purpose: 替换 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleBuffetPackageName()`
  - Requirements: Requirement 5
  - Leverage: 现有代码: `main/app/service/order.go:2777` - `checkBuffetCustomerTypePriceChanged()` 方法
  - Key Logic:
    ```go
    // 修改前
    buffetLocaleName := saleBill.GetLocaleBuffetPackageNameByUuid(
        buffetCustomer.BuffetPackageUuid,
        buffetCustomer.BuffetPackage.MultiLanguageName,
    )
    
    // 修改后
    buffetLocaleName := buffetCustomer.GetLocaleBuffetPackageName()
    ```
  - Prompt: Role: Go Developer | Task: 修改 checkBuffetCustomerTypePriceChanged() 方法，使用 buffetCustomer.GetLocaleBuffetPackageName() 替换 SaleBill.GetLocaleBuffetPackageNameByUuid() 调用 | Context: 方法在 order.go 第 2777 行附近，需要替换自助餐名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [x] 3.3 修改 GetCustomerList() 方法 - 使用 GetLocaleBuffetPackageName()

  - File: `main/app/model/sale_order.go`
  - Purpose: 替换 `SaleBill.GetLocaleBuffetPackageNameByUuid()` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleBuffetPackageName()`
  - Requirements: Requirement 5
  - Leverage: 现有代码: `main/app/model/sale_order.go:670` - `GetCustomerList()` 方法
  - Key Logic:
    ```go
    // 修改前
    buffetLocaleName := model.SaleBill.GetLocaleBuffetPackageNameByUuid(
        orderBuffetCustomer.BuffetPackageUuid,
        orderBuffetCustomer.BuffetPackage.MultiLanguageName,
    )
    
    // 修改后
    buffetLocaleName := orderBuffetCustomer.GetLocaleBuffetPackageName()
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetCustomerList() 方法，使用 orderBuffetCustomer.GetLocaleBuffetPackageName() 替换 SaleBill.GetLocaleBuffetPackageNameByUuid() 调用 | Context: 方法在 sale_order.go 第 670 行附近，需要替换自助餐名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [x] 3.4 搜索并修改其他使用 SaleOrderBuffetCustomerType 的地方

  - File: 根据搜索结果确定
  - Purpose: 确保所有使用 `SaleOrderBuffetCustomerType` 的地方都使用快照方法
  - Requirements: Requirement 5
  - Leverage: 使用 grep 搜索: `grep -r "SaleOrderBuffetCustomerType\|sale_order_buffet_customer_type" main/app/service/ main/app/model/`
  - Search Command:
    ```bash
    cd main && grep -r "SaleBill\.GetLocaleBuffetPackageNameByUuid\|orderBuffetCustomer\|buffetCustomer" app/service/ app/model/ | grep -i "buffet"
    ```
  - Prompt: Role: Go Developer | Task: 搜索所有使用 SaleOrderBuffetCustomerType 的地方，确保都使用 GetLocaleBuffetPackageName() 方法 | Context: 搜索代码库中使用 SaleBill.GetLocaleBuffetPackageNameByUuid() 且参数包含 SaleOrderBuffetCustomerType 的地方 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有相关代码都已修改，统一使用快照方法
  - Success: 所有相关代码都已修改，统一使用快照方法

---

## Phase 4: 下单逻辑修改

### 4.1 修改订单创建逻辑 - 保存快照

- [x] 4.1 修改 NewSaleOrderBuffetCustomerType() 方法 - 保存快照

  - File: `main/app/model/sale_order.go`
  - Purpose: 在 `NewSaleOrderBuffetCustomerType()` 方法中，创建 `SaleOrderBuffetCustomerType` 后保存自助餐套餐名称快照
  - Requirements: Requirement 6（下单逻辑修改）
  - Leverage: 现有方法: `main/app/model/sale_order.go:1178` - `NewSaleOrderBuffetCustomerType()` 方法
  - Key Logic:
    ```go
    saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
        // ... 现有字段
    }
    
    // 计算金额
    saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    // 注意：需要确保 BuffetPackage 已加载，否则无法获取 MultiLanguageName
    // 如果 BuffetPackage 未加载，需要在调用此方法前先加载，或者在调用后单独设置快照
    ```
  - Note: 此方法可能无法直接访问 `BuffetPackage`，需要在调用方设置快照
  - Prompt: Role: Go Developer | Task: 在 NewSaleOrderBuffetCustomerType() 方法中添加注释，说明需要在调用方设置快照 | Context: 方法在 sale_order.go 第 1178 行附近，需要确保 BuffetPackage 已加载才能设置快照 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 注释添加正确，说明清晰
  - Success: 注释添加正确，说明清晰

- [x] 4.2 修改 NewSaleOrderBuffetCustomerType() 函数 - 保存快照

  - File: `main/app/model/sale_order.go`
  - Purpose: 在 `NewSaleOrderBuffetCustomerType()` 函数中，创建 `SaleOrderBuffetCustomerType` 后保存自助餐套餐名称快照
  - Requirements: Requirement 6
  - Leverage: 现有函数: `main/app/model/sale_order.go:1291` - `NewSaleOrderBuffetCustomerType()` 函数
  - Key Logic:
    ```go
    saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
        // ... 现有字段
    }
    
    // 计算金额
    saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    // 注意：需要确保 BuffetPackage 已加载，否则无法获取 MultiLanguageName
    // 如果 BuffetPackage 未加载，需要在调用此函数前先加载，或者在调用后单独设置快照
    ```
  - Note: 此函数可能无法直接访问 `BuffetPackage`，需要在调用方设置快照
  - Prompt: Role: Go Developer | Task: 在 NewSaleOrderBuffetCustomerType() 函数中添加注释，说明需要在调用方设置快照 | Context: 函数在 sale_order.go 第 1291 行附近，需要确保 BuffetPackage 已加载才能设置快照 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 注释添加正确，说明清晰
  - Success: 注释添加正确，说明清晰

- [x] 4.3 修改 GetSaleOrderBuffetCustomerTypes() 方法 - 保存快照

  - File: `main/app/model/sale_order_ext_getset.go`
  - Purpose: 在 `GetSaleOrderBuffetCustomerTypes()` 方法中，创建 `SaleOrderBuffetCustomerType` 后保存自助餐套餐名称快照
  - Requirements: Requirement 6
  - Leverage: 现有方法: `main/app/model/sale_order_ext_getset.go:628` - `GetSaleOrderBuffetCustomerTypes()` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(...)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    if buffet, ok := buffetMap[buffetUuid]; ok && !buffet.MultiLanguageName.IsNullName() {
        if err := saleOrderBuffetCustomerType.SetBuffetPackageNameSnapshot(buffet.MultiLanguageName); err != nil {
            // 记录错误日志，但不中断流程
            logger.Logger.Error("保存自助餐套餐名称快照失败", zap.Error(err), zap.Uint64("buffet_package_uuid", buffetUuid))
        }
    }
    ```
  - Prompt: Role: Go Developer | Task: 在 GetSaleOrderBuffetCustomerTypes() 方法中，创建 SaleOrderBuffetCustomerType 后调用 SetBuffetPackageNameSnapshot() 保存快照 | Context: 方法在 sale_order_ext_getset.go 第 628 行附近，buffetList 已加载，可以从 buffetMap 中获取 BuffetPackage | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录错误日志但不中断流程 | Success: 快照保存逻辑添加正确，错误处理完善
  - Success: 快照保存逻辑添加正确，错误处理完善

- [x] 4.4 修改 CreateDeskOrder 下单逻辑 - 保存快照

  - File: `main/app/service/order_base.go`
  - Purpose: 在 `CreateDeskOrder` 下单逻辑中，创建 `SaleOrderBuffetCustomerType` 后保存自助餐套餐名称快照
  - Requirements: Requirement 6
  - Leverage: 现有代码: `main/app/service/order_base.go:177` - `CreateDeskOrder` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderBuffetCustomerTypes, _, _, maxTimeLimit, nonOrderingTime, reminderOrderTime := saleOrder.GetSaleOrderBuffetCustomerTypes(...)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    // 注意：GetSaleOrderBuffetCustomerTypes() 方法内部已设置快照，此处无需重复设置
    // 但如果 GetSaleOrderBuffetCustomerTypes() 未设置快照，需要在此处设置
    ```
  - Note: 如果 `GetSaleOrderBuffetCustomerTypes()` 方法已设置快照（Task 4.3），则此处无需重复设置
  - Prompt: Role: Go Developer | Task: 检查 CreateDeskOrder 下单逻辑，确保 SaleOrderBuffetCustomerType 的快照已保存 | Context: 方法在 order_base.go 第 177 行附近，调用 GetSaleOrderBuffetCustomerTypes() 创建 SaleOrderBuffetCustomerType | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 快照保存逻辑正确，所有下单入口都保存快照
  - Success: 快照保存逻辑正确，所有下单入口都保存快照

- [x] 4.5 修改 OrderChangeBuffet 修改逻辑 - 保存快照

  - File: `main/app/service/order_buffet.go`
  - Purpose: 在 `OrderChangeBuffet` 修改逻辑中，创建 `SaleOrderBuffetCustomerType` 后保存自助餐套餐名称快照
  - Requirements: Requirement 6
  - Leverage: 现有代码: `main/app/service/order_buffet.go:110` - `OrderChangeBuffet` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderCustomerTypes, buffetUuids, mealNum, maxTimeLimit, _, _ := saleOrder.GetSaleOrderBuffetCustomerTypes(...)
    
    // 设置自助餐套餐名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-package-name-snapshot-fix
    // 注意：GetSaleOrderBuffetCustomerTypes() 方法内部已设置快照，此处无需重复设置
    // 但如果 GetSaleOrderBuffetCustomerTypes() 未设置快照，需要在此处设置
    ```
  - Note: 如果 `GetSaleOrderBuffetCustomerTypes()` 方法已设置快照（Task 4.3），则此处无需重复设置
  - Prompt: Role: Go Developer | Task: 检查 OrderChangeBuffet 修改逻辑，确保 SaleOrderBuffetCustomerType 的快照已保存 | Context: 方法在 order_buffet.go 第 110 行附近，调用 GetSaleOrderBuffetCustomerTypes() 创建 SaleOrderBuffetCustomerType | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 快照保存逻辑正确，所有修改入口都保存快照
  - Success: 快照保存逻辑正确，所有修改入口都保存快照

---

## Phase 5: 测试验证

### 5.1 单元测试

- [ ] 5.1 编写 GetLocaleBuffetPackageName() 和 SetBuffetPackageNameSnapshot() 单元测试

  - File: `main/app/model/sale_order_buffet_customer_type_test.go`
  - Purpose: 测试自助餐套餐名称快照方法的正确性
  - Requirements: Requirement 3, 4
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Test Cases:
    - GetLocaleBuffetPackageName() - 快照字段有值且有效 JSON
    - GetLocaleBuffetPackageName() - 快照字段为空
    - GetLocaleBuffetPackageName() - 快照字段无效 JSON
    - GetLocaleBuffetPackageName() - 关联表数据为空
    - SetBuffetPackageNameSnapshot() - 正常序列化
    - SetBuffetPackageNameSnapshot() - 多语言名称为空
  - Prompt: Role: QA Engineer | Task: 为 GetLocaleBuffetPackageName() 和 SetBuffetPackageNameSnapshot() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据等场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试覆盖率 ≥ 80%，所有测试通过

### 5.2 集成测试

- [ ] 5.2 编写下单集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试下单时保存快照数据
  - Requirements: Requirement 6
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单（包含自助餐） → 验证 `SaleOrderBuffetCustomerType.BuffetPackageName` 字段保存成功（JSON 格式）
    - 删除自助餐配置 → 查询订单仍显示快照数据
  - Prompt: Role: QA Engineer | Task: 编写下单集成测试，验证创建订单时自助餐套餐名称快照正确保存为 JSON | Context: 测试所有下单入口（POS、扫码点餐、外卖等） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有下单场景的快照数据保存正确
  - Success: 所有下单场景的快照数据保存正确

- [ ] 5.3 编写查询集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试查询时使用快照数据
  - Requirements: Requirement 5
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单 → 查询订单 → 验证使用快照数据
    - 删除自助餐配置 → 查询订单 → 验证仍显示快照数据
    - 修改自助餐名称 → 查询订单 → 验证仍显示修改前的名称
    - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常
  - Prompt: Role: QA Engineer | Task: 编写查询集成测试，验证订单查询时优先使用快照数据，后台删除自助餐套餐后，历史订单仍能正常显示 | Context: 测试订单详情、订单列表、订单打印、订单导出等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询场景的快照逻辑正确
  - Success: 所有查询场景的快照逻辑正确

### 5.3 回归测试

- [ ] 5.4 回归测试 - 订单查询接口

  - File: -
  - Purpose: 确保订单查询功能不受影响
  - Requirements: Requirement 5
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单查询接口回归测试，确保所有订单查询接口正常工作 | Context: 测试订单详情、订单列表、订单搜索等接口 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有订单查询接口测试通过
  - Success: 所有订单查询接口测试通过

- [ ] 5.5 回归测试 - 订单打印/导出/报表

  - File: -
  - Purpose: 确保订单打印、导出、报表功能不受影响
  - Requirements: Requirement 5
  - Leverage: 现有测试用例
  - Prompt: Role: QA Engineer | Task: 执行订单打印、导出、报表回归测试，确保所有功能正常工作 | Context: 测试订单打印、订单导出、订单报表等功能 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有功能测试通过
  - Success: 所有功能测试通过

---

## Phase 6: 数据迁移（可选）

### 6.1 数据检查脚本

- [ ] 6.1 编写数据检查脚本

  - File: `scripts/check_buffet_customer_type_package_snapshot.sql`
  - Purpose: 检查历史订单的 `buffet_package_name` 字段填充情况
  - Requirements: Requirement 1
  - Leverage: 参考 `scripts/check_order_source_snapshot.sql`
  - SQL:
    ```sql
    -- 统计需要迁移的订单数量
    SELECT COUNT(*) AS total_need_migrate
    FROM ttpos_sale_order_buffet_customer_type sobct
    WHERE sobct.buffet_package_name = '' 
      AND sobct.buffet_package_uuid != 0 
      AND sobct.deleted_at IS NULL;
    
    -- 统计关联表数据存在的记录数
    SELECT COUNT(*) AS total_can_migrate
    FROM ttpos_sale_order_buffet_customer_type sobct
    INNER JOIN ttpos_buffet_package bp ON sobct.buffet_package_uuid = bp.uuid
    INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
    WHERE sobct.buffet_package_name = '' 
      AND sobct.buffet_package_uuid != 0 
      AND mln.zh_name != ''
      AND sobct.deleted_at IS NULL;
    ```
  - Success: 检查脚本创建成功，统计结果准确

### 6.2 数据迁移脚本（可选）

- [ ] 6.2 编写数据迁移脚本（可选）

  - File: `scripts/migrate_buffet_customer_type_package_snapshot.sql`
  - Purpose: 从关联表补充历史订单的 `buffet_package_name` 快照字段
  - Requirements: Requirement 1
  - Leverage: Task 6.1 的检查脚本
  - SQL:
    ```sql
    -- 补充历史订单的自助餐套餐名称快照（仅迁移关联表数据存在的记录）
    UPDATE ttpos_sale_order_buffet_customer_type sobct
    INNER JOIN ttpos_buffet_package bp ON sobct.buffet_package_uuid = bp.uuid
    INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
    SET sobct.buffet_package_name = JSON_OBJECT(
        'zh', mln.zh_name,
        'th', IFNULL(mln.th_name, ''),
        'en', IFNULL(mln.en_name, ''),
        'zhtw', IFNULL(mln.zhtw_name, ''),
        'ja', IFNULL(mln.ja_name, ''),
        'ko', IFNULL(mln.ko_name, ''),
        'my', IFNULL(mln.my_name, ''),
        'tr', IFNULL(mln.tr_name, ''),
        'sv', IFNULL(mln.sv_name, '')
    )
    WHERE sobct.buffet_package_name = '' 
      AND sobct.buffet_package_uuid != 0 
      AND mln.zh_name != ''
      AND sobct.deleted_at IS NULL;
    ```
  - Note: 可选执行，不强制要求
  - Success: 迁移脚本创建成功，执行后数据补充正确

- [ ] 6.3 执行数据迁移（可选，测试环境）

  - File: -
  - Purpose: 在测试环境执行数据迁移脚本，验证迁移逻辑
  - Requirements: Requirement 1
  - Leverage: Task 6.2 的迁移脚本
  - Command: 在测试环境执行 SQL 脚本
  - Success: 迁移执行成功，数据补充正确，测试环境验证通过

---

## Phase 7: 生产环境部署

- [ ] 7.1 执行生产环境迁移

  - File: -
  - Purpose: 在生产环境执行数据库迁移
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 生产环境迁移执行成功，字段已添加

- [ ] 7.2 执行生产环境数据迁移（可选）

  - File: -
  - Purpose: 在生产环境执行数据迁移脚本（可选）
  - Requirements: Requirement 1
  - Leverage: Task 6.2 的迁移脚本
  - Command: 在生产环境执行 SQL 脚本
  - Success: 生产环境数据迁移执行成功（如执行）

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}

