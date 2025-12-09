# 自助餐顾客类型名称快照修复 任务分解

> 本文档定义自助餐顾客类型名称快照修复功能的详细执行任务清单。

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

## Phase 1: 数据库迁移

### 1.1 创建数据库迁移文件

- [ ] 1.1 创建数据库迁移文件 - 修改 name 字段类型为 TEXT

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_modify_sale_order_buffet_customer_type_name_to_text.php`
  - Purpose: 将 `ttpos_sale_order_buffet_customer_type.name` 字段类型从 `VARCHAR(255)` 修改为 `TEXT`
  - Requirements: Requirement 1（数据库结构变更）
  - Leverage: 现有迁移文件: `admin/database/migrations/20251209094516_modify_sale_order_product_name_to_text.php`
  - SQL:
    ```sql
    ALTER TABLE `ttpos_sale_order_buffet_customer_type` 
    MODIFY COLUMN `name` TEXT NOT NULL DEFAULT '' 
    COMMENT '顾客类型名称快照（JSON），不随后台更新';
    ```
  - Note: **JSON 方案** - 快照包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV），所有语言使用相同值
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，将 ttpos_sale_order_buffet_customer_type.name 字段类型改为 TEXT | Context: 字段注释更新为"顾客类型名称快照（JSON），不随后台更新"，迁移脚本需要检查字段当前类型，如果已经是 TEXT，则跳过（幂等性） | Restrictions: 遵循 .cursor/rules/database.mdc，迁移脚本支持可重复执行 | Success: 迁移文件创建成功，字段类型修改正确
  - Success: 迁移文件创建成功，字段类型修改正确

- [ ] 1.2 执行数据库迁移（测试环境）

  - File: -
  - Purpose: 在测试环境数据库中将 `name` 字段类型修改为 `TEXT`
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段类型已修改为 TEXT，测试环境验证通过

---

## Phase 2: Go Model 修改

### 2.1 修改 SaleOrderBuffetCustomerType 模型字段

- [ ] 2.1 修改 Name 字段类型为 TEXT

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 将 `Name` 字段的 gorm 标签类型从 `varchar(255)` 改为 `text`
  - Requirements: Requirement 2（数据模型修改）
  - Leverage: 现有字段定义: `main/app/model/sale_order_buffet_customer_type.go:17` - `Name` 字段
  - Code:
    ```go
    Name string `gorm:"column:name;type:text" json:"name"` // 修改为 TEXT 类型，存储 JSON 快照
    ```
  - Note: **JSON 方案** - 字段存储完整多语言 JSON
  - Prompt: Role: Go Developer | Task: 修改 SaleOrderBuffetCustomerType.Name 字段的 gorm 标签，将类型改为 text，注释更新为"顾客类型名称快照（JSON），不随后台更新" | Context: 字段定义在 sale_order_buffet_customer_type.go 第 17 行附近 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段类型修改正确，注释更新正确，编译通过
  - Success: 字段类型修改正确，注释更新正确，编译通过

### 2.2 实现快照方法

- [ ] 2.2 实现 GetLocaleName() 方法（JSON 方案）

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 实现顾客类型名称获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 3（查询逻辑修改）
  - Leverage: 参考 `GetLocaleOrderSourceName()` 方法的实现（`main/app/model/sale_bill.go:789`）
  - Key Logic (JSON 方案):
    1. 优先使用 `Name` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `BuffetCustomerTypePrice.BuffetCustomerType.Name`（单语言）
    4. 将单语言名称转换为多语言格式（所有语言使用相同值）
    5. 都为空时返回空响应
  - Code Reference:
    ```go
    // 参考 GetLocaleOrderSourceName() 的实现
    func (model *SaleOrderBuffetCustomerType) GetLocaleName() dto.LocaleResponse {
        // 优先使用快照字段
        snapshotName := model.Name

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
        if model.BuffetCustomerTypePrice.BuffetCustomerType.Name != "" {
            // BuffetCustomerType.Name 是单语言字段，转换为多语言格式
            return dto.LocaleResponse{
                ZH:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                TH:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                EN:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                ZHTW: model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                JA:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                KO:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                MY:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                TR:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
                SV:   model.BuffetCustomerTypePrice.BuffetCustomerType.Name,
            }
        }

        return dto.LocaleResponse{}
    }
    ```
  - Import: 需要添加 `encoding/json` 和 `ttpos-server-go/app/dto` 导入
  - Prompt: Role: Go Developer | Task: 实现 GetLocaleName() 方法，优先使用快照字段，降级使用关联表数据 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，快照字段为 JSON 格式，需要反序列化为 LocaleResponse，降级时注意 BuffetCustomerType.Name 是单语言字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确，逻辑完整，编译通过
  - Success: 方法实现正确，逻辑完整，编译通过

- [ ] 2.3 实现 SetNameSnapshot() 方法（JSON 方案）

  - File: `main/app/model/sale_order_buffet_customer_type.go`
  - Purpose: 实现快照保存方法，从单语言名称转换为多语言 JSON
  - Requirements: Requirement 4（下单逻辑修改）
  - Leverage: 参考 `SetOrderSourceNameSnapshot()` 方法的实现（`main/app/model/sale_bill.go:815`）
  - Key Logic (JSON 方案):
    1. 如果 `customerTypeName` 为空，设置为空字符串
    2. 将单语言名称转换为多语言格式（所有语言使用相同值）
    3. 序列化为 JSON 字符串
    4. 保存到 `Name` 字段
  - Code Reference:
    ```go
    // 参考 SetOrderSourceNameSnapshot() 的实现
    func (model *SaleOrderBuffetCustomerType) SetNameSnapshot(customerTypeName string) error {
        // 如果名称为空，设置为空字符串
        if customerTypeName == "" {
            model.Name = ""
            return nil
        }

        // 构建多语言响应（所有语言使用相同值）
        localeResp := dto.LocaleResponse{
            ZH:   customerTypeName,
            TH:   customerTypeName,
            EN:   customerTypeName,
            ZHTW: customerTypeName,
            JA:   customerTypeName,
            KO:   customerTypeName,
            MY:   customerTypeName,
            TR:   customerTypeName,
            SV:   customerTypeName,
        }

        // 序列化为 JSON
        jsonData, err := json.Marshal(localeResp)
        if err != nil {
            return err
        }

        model.Name = string(jsonData)
        return nil
    }
    ```
  - Import: 需要添加 `encoding/json` 和 `ttpos-server-go/app/dto` 导入
  - Prompt: Role: Go Developer | Task: 实现 SetNameSnapshot() 方法，从单语言名称转换为多语言 JSON 保存 | Context: 参考 SetOrderSourceNameSnapshot() 的实现模式，参数为单语言字符串，需要转换为多语言格式（所有语言使用相同值） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确，序列化逻辑完整，编译通过
  - Success: 方法实现正确，序列化逻辑完整，编译通过

---

## Phase 3: 查询逻辑修改

### 3.1 替换现有查询方法

- [ ] 3.1 修改 GetOrderInfos() 方法 - 使用 GetLocaleName()

  - File: `main/app/service/order_manage.go`
  - Purpose: 替换 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleName()`
  - Requirements: Requirement 5（查询逻辑修改）
  - Leverage: 现有代码: `main/app/service/order_manage.go:462` - `GetOrderInfos()` 方法
  - Key Logic:
    ```go
    // 修改前
    LocaleAttributeName: dto.LocaleResponse{
        ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    },
    
    // 修改后
    LocaleAttributeName: orderBuffetCustomer.GetLocaleName(),
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetOrderInfos() 方法，使用 orderBuffetCustomer.GetLocaleName() 替换直接使用 BuffetCustomerTypePrice.BuffetCustomerType.Name 的代码 | Context: 方法在 order_manage.go 第 462 行附近，需要替换顾客类型名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [ ] 3.2 修改 GetOrderInfos() 方法 - 订单导出功能

  - File: `main/app/service/order_manage.go`
  - Purpose: 替换订单导出功能中的 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 调用
  - Requirements: Requirement 5
  - Leverage: 现有代码: `main/app/service/order_manage.go:274` - 订单导出相关代码
  - Key Logic:
    ```go
    // 修改前
    AttrName: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    
    // 修改后
    AttrName: orderBuffetCustomer.GetLocaleName().GetLocale(language),
    ```
  - Prompt: Role: Go Developer | Task: 修改订单导出功能，使用 orderBuffetCustomer.GetLocaleName() 获取多语言名称 | Context: 方法在 order_manage.go 第 274 行附近，需要根据语言获取对应的名称 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [ ] 3.3 修改 checkBuffetCustomerTypePriceChanged() 方法 - 使用 GetLocaleName()

  - File: `main/app/service/order.go`
  - Purpose: 替换 `buffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleName()`
  - Requirements: Requirement 5
  - Leverage: 现有代码: `main/app/service/order.go:2783` - `checkBuffetCustomerTypePriceChanged()` 方法
  - Key Logic:
    ```go
    // 修改前
    LocaleAttributeName: dto.LocaleResponse{
        ZH:   buffetCustomer.Name,
        TH:   buffetCustomer.Name,
        EN:   buffetCustomer.Name,
        ZHTW: buffetCustomer.Name,
        JA:   buffetCustomer.Name,
        KO:   buffetCustomer.Name,
        MY:   buffetCustomer.Name,
        TR:   buffetCustomer.Name,
        SV:   buffetCustomer.Name,
    },
    
    // 修改后
    LocaleAttributeName: buffetCustomer.GetLocaleName(),
    ```
  - Prompt: Role: Go Developer | Task: 修改 checkBuffetCustomerTypePriceChanged() 方法，使用 buffetCustomer.GetLocaleName() 替换直接使用 Name 字段的代码 | Context: 方法在 order.go 第 2783 行附近，需要替换顾客类型名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [ ] 3.4 修改 GetCustomerList() 方法 - 使用 GetLocaleName()

  - File: `main/app/model/sale_order.go`
  - Purpose: 替换 `orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name` 调用，使用 `SaleOrderBuffetCustomerType.GetLocaleName()`
  - Requirements: Requirement 5
  - Leverage: 现有代码: `main/app/model/sale_order.go:673` - `GetCustomerList()` 方法
  - Key Logic:
    ```go
    // 修改前
    LocaleAttributeName: dto.LocaleResponse{
        ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
        SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
    },
    
    // 修改后
    LocaleAttributeName: orderBuffetCustomer.GetLocaleName(),
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetCustomerList() 方法，使用 orderBuffetCustomer.GetLocaleName() 替换直接使用 BuffetCustomerTypePrice.BuffetCustomerType.Name 的代码 | Context: 方法在 sale_order.go 第 673 行附近，需要替换顾客类型名称获取逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法修改正确，功能不变，编译通过
  - Success: 方法修改正确，功能不变，编译通过

- [ ] 3.5 搜索并修改其他使用 SaleOrderBuffetCustomerType 的地方

  - File: 根据搜索结果确定
  - Purpose: 确保所有使用 `SaleOrderBuffetCustomerType` 的地方都使用快照方法
  - Requirements: Requirement 5
  - Leverage: 使用 grep 搜索: `grep -r "BuffetCustomerTypePrice\.BuffetCustomerType\.Name\|orderBuffetCustomer\.Name\|buffetCustomer\.Name" main/app/service/ main/app/model/`
  - Search Command:
    ```bash
    cd main && grep -r "BuffetCustomerTypePrice\.BuffetCustomerType\.Name\|orderBuffetCustomer\.Name\|buffetCustomer\.Name" app/service/ app/model/ | grep -i "buffet"
    ```
  - Prompt: Role: Go Developer | Task: 搜索所有使用 SaleOrderBuffetCustomerType 的地方，确保都使用 GetLocaleName() 方法 | Context: 搜索代码库中直接使用 BuffetCustomerTypePrice.BuffetCustomerType.Name 或 Name 字段的地方 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有相关代码都已修改，统一使用快照方法
  - Success: 所有相关代码都已修改，统一使用快照方法

---

## Phase 4: 下单逻辑修改

### 4.1 修改订单创建逻辑 - 保存快照

- [ ] 4.1 修改 NewSaleOrderBuffetCustomerType() 方法 - 保存快照

  - File: `main/app/model/sale_order.go`
  - Purpose: 在 `NewSaleOrderBuffetCustomerType()` 方法中，创建 `SaleOrderBuffetCustomerType` 后保存顾客类型名称快照
  - Requirements: Requirement 6（下单逻辑修改）
  - Leverage: 现有方法: `main/app/model/sale_order.go:1160` - `NewSaleOrderBuffetCustomerType()` 方法
  - Key Logic:
    ```go
    saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
        // ... 现有字段
    }
    
    // 计算金额
    saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
    
    // 设置顾客类型名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-name-snapshot-fix
    // 注意：需要确保 BuffetCustomerTypePrice.BuffetCustomerType 已加载，否则无法获取 Name
    // 如果未加载，需要在调用此方法前先加载，或者在调用后单独设置快照
    ```
  - Note: 此方法可能无法直接访问 `BuffetCustomerTypePrice.BuffetCustomerType`，需要在调用方设置快照
  - Prompt: Role: Go Developer | Task: 在 NewSaleOrderBuffetCustomerType() 方法中添加注释，说明需要在调用方设置快照 | Context: 方法在 sale_order.go 第 1160 行附近，需要确保 BuffetCustomerTypePrice.BuffetCustomerType 已加载才能设置快照 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 注释添加正确，说明清晰
  - Success: 注释添加正确，说明清晰

- [ ] 4.2 修改 NewSaleOrderBuffetCustomerType() 函数 - 保存快照

  - File: `main/app/model/sale_order.go`
  - Purpose: 在 `NewSaleOrderBuffetCustomerType()` 函数中，创建 `SaleOrderBuffetCustomerType` 后保存顾客类型名称快照
  - Requirements: Requirement 6
  - Leverage: 现有函数: `main/app/model/sale_order.go:1279` - `NewSaleOrderBuffetCustomerType()` 函数
  - Key Logic:
    ```go
    saleOrderBuffetCustomerType := &SaleOrderBuffetCustomerType{
        Name: customerName, // 当前直接设置单语言名称
        // ... 其他字段
    }
    
    // 计算金额
    saleOrderBuffetCustomerType.CalcSaleOrderBuffetCustomerType(setting)
    
    // 设置顾客类型名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-name-snapshot-fix
    // 注意：customerName 参数是单语言名称，需要转换为多语言 JSON
    if err := saleOrderBuffetCustomerType.SetNameSnapshot(customerName); err != nil {
        // 记录错误日志，但不中断流程
        logger.Logger.Error("保存顾客类型名称快照失败", zap.Error(err), zap.String("customer_name", customerName))
    }
    ```
  - Prompt: Role: Go Developer | Task: 在 NewSaleOrderBuffetCustomerType() 函数中，创建 SaleOrderBuffetCustomerType 后调用 SetNameSnapshot() 保存快照 | Context: 函数在 sale_order.go 第 1279 行附近，customerName 参数是单语言名称，需要转换为多语言 JSON | Restrictions: 遵循 .cursor/rules/go-main.mdc，记录错误日志但不中断流程 | Success: 快照保存逻辑添加正确，错误处理完善
  - Success: 快照保存逻辑添加正确，错误处理完善

- [ ] 4.3 修改 GetSaleOrderBuffetCustomerTypes() 方法 - 保存快照

  - File: `main/app/model/sale_order_ext_getset.go`
  - Purpose: 在 `GetSaleOrderBuffetCustomerTypes()` 方法中，创建 `SaleOrderBuffetCustomerType` 后保存顾客类型名称快照
  - Requirements: Requirement 6
  - Leverage: 现有方法: `main/app/model/sale_order_ext_getset.go:706` - `GetSaleOrderBuffetCustomerTypes()` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderBuffetCustomerType := NewSaleOrderBuffetCustomerType(customerTypePrice.Name, ...)
    
    // 设置顾客类型名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-name-snapshot-fix
    // 注意：NewSaleOrderBuffetCustomerType() 函数内部已设置快照（Task 4.2），此处无需重复设置
    // 但如果 NewSaleOrderBuffetCustomerType() 未设置快照，需要在此处设置
    ```
  - Note: 如果 `NewSaleOrderBuffetCustomerType()` 函数已设置快照（Task 4.2），则此处无需重复设置
  - Prompt: Role: Go Developer | Task: 检查 GetSaleOrderBuffetCustomerTypes() 方法，确保 SaleOrderBuffetCustomerType 的快照已保存 | Context: 方法在 sale_order_ext_getset.go 第 706 行附近，调用 NewSaleOrderBuffetCustomerType() 创建 SaleOrderBuffetCustomerType | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 快照保存逻辑正确，所有创建入口都保存快照
  - Success: 快照保存逻辑正确，所有创建入口都保存快照

- [ ] 4.4 修改 CreateDeskOrder 下单逻辑 - 保存快照

  - File: `main/app/service/order_base.go`
  - Purpose: 在 `CreateDeskOrder` 下单逻辑中，创建 `SaleOrderBuffetCustomerType` 后保存顾客类型名称快照
  - Requirements: Requirement 6
  - Leverage: 现有代码: `main/app/service/order_base.go:180` - `CreateDeskOrder` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderBuffetCustomerTypes, _, _, maxTimeLimit, nonOrderingTime, reminderOrderTime := saleOrder.GetSaleOrderBuffetCustomerTypes(...)
    
    // 设置顾客类型名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-name-snapshot-fix
    // 注意：GetSaleOrderBuffetCustomerTypes() 方法内部已设置快照，此处无需重复设置
    // 但如果 GetSaleOrderBuffetCustomerTypes() 未设置快照，需要在此处设置
    ```
  - Note: 如果 `GetSaleOrderBuffetCustomerTypes()` 方法已设置快照（Task 4.3），则此处无需重复设置
  - Prompt: Role: Go Developer | Task: 检查 CreateDeskOrder 下单逻辑，确保 SaleOrderBuffetCustomerType 的快照已保存 | Context: 方法在 order_base.go 第 180 行附近，调用 GetSaleOrderBuffetCustomerTypes() 创建 SaleOrderBuffetCustomerType | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 快照保存逻辑正确，所有下单入口都保存快照
  - Success: 快照保存逻辑正确，所有下单入口都保存快照

- [ ] 4.5 修改 OrderChangeBuffet 修改逻辑 - 保存快照

  - File: `main/app/service/order_buffet.go`
  - Purpose: 在 `OrderChangeBuffet` 修改逻辑中，创建 `SaleOrderBuffetCustomerType` 后保存顾客类型名称快照
  - Requirements: Requirement 6
  - Leverage: 现有代码: `main/app/service/order_buffet.go:110` - `OrderChangeBuffet` 方法
  - Key Logic:
    ```go
    // 创建 SaleOrderBuffetCustomerType
    saleOrderCustomerTypes, buffetUuids, mealNum, maxTimeLimit, _, _ := saleOrder.GetSaleOrderBuffetCustomerTypes(...)
    
    // 设置顾客类型名称快照（JSON 方案）
    // Requirement: story-main-buffet-customer-type-name-snapshot-fix
    // 注意：GetSaleOrderBuffetCustomerTypes() 方法内部已设置快照，此处无需重复设置
    // 但如果 GetSaleOrderBuffetCustomerTypes() 未设置快照，需要在此处设置
    ```
  - Note: 如果 `GetSaleOrderBuffetCustomerTypes()` 方法已设置快照（Task 4.3），则此处无需重复设置
  - Prompt: Role: Go Developer | Task: 检查 OrderChangeBuffet 修改逻辑，确保 SaleOrderBuffetCustomerType 的快照已保存 | Context: 方法在 order_buffet.go 第 110 行附近，调用 GetSaleOrderBuffetCustomerTypes() 创建 SaleOrderBuffetCustomerType | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 快照保存逻辑正确，所有修改入口都保存快照
  - Success: 快照保存逻辑正确，所有修改入口都保存快照

---

## Phase 5: 测试验证

### 5.1 单元测试

- [ ] 5.1 编写 GetLocaleName() 和 SetNameSnapshot() 单元测试

  - File: `main/app/model/sale_order_buffet_customer_type_test.go`
  - Purpose: 测试顾客类型名称快照方法的正确性
  - Requirements: Requirement 3, 4
  - Leverage: 现有测试: `main/app/model/sale_bill_order_source_test.go`
  - Test Cases:
    - GetLocaleName() - 快照字段有值且有效 JSON
    - GetLocaleName() - 快照字段为空
    - GetLocaleName() - 快照字段无效 JSON
    - GetLocaleName() - 关联表数据为空
    - SetNameSnapshot() - 正常序列化
    - SetNameSnapshot() - 名称为空
  - Prompt: Role: QA Engineer | Task: 为 GetLocaleName() 和 SetNameSnapshot() 编写单元测试，覆盖快照有值/无值、JSON 有效/无效、关联表有数据/无数据等场景 | Context: 参考 sale_bill_order_source_test.go 的测试模式 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 测试覆盖率 ≥ 80%，所有测试通过

### 5.2 集成测试

- [ ] 5.2 编写下单集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试下单时保存快照数据
  - Requirements: Requirement 6
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单（包含自助餐） → 验证 `SaleOrderBuffetCustomerType.Name` 字段保存成功（JSON 格式）
    - 删除顾客类型配置 → 查询订单仍显示快照数据
  - Prompt: Role: QA Engineer | Task: 编写下单集成测试，验证创建订单时顾客类型名称快照正确保存为 JSON | Context: 测试所有下单入口（POS、扫码点餐、外卖等） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有下单场景的快照数据保存正确
  - Success: 所有下单场景的快照数据保存正确

- [ ] 5.3 编写查询集成测试

  - File: `main/app/service/order*_test.go`
  - Purpose: 测试查询时使用快照数据
  - Requirements: Requirement 5
  - Leverage: 现有测试: `main/app/service/order*_test.go`
  - Test Cases:
    - 创建订单 → 查询订单 → 验证使用快照数据
    - 删除顾客类型配置 → 查询订单 → 验证仍显示快照数据
    - 修改顾客类型名称 → 查询订单 → 验证仍显示修改前的名称
    - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常
  - Prompt: Role: QA Engineer | Task: 编写查询集成测试，验证订单查询时优先使用快照数据，后台删除顾客类型后，历史订单仍能正常显示 | Context: 测试订单详情、订单列表、订单打印、订单导出等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询场景的快照逻辑正确
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

  - File: `scripts/check_buffet_customer_type_name_snapshot.sql`
  - Purpose: 检查历史订单的 `name` 字段填充情况
  - Requirements: Requirement 1
  - Leverage: 参考 `scripts/check_order_source_snapshot.sql`
  - SQL:
    ```sql
    -- 统计需要迁移的订单数量
    SELECT COUNT(*) AS total_need_migrate
    FROM ttpos_sale_order_buffet_customer_type sobct
    WHERE (sobct.name = '' OR sobct.name NOT LIKE '{%')
      AND sobct.buffet_customer_type_price_uuid != 0 
      AND sobct.deleted_at IS NULL;
    
    -- 统计关联表数据存在的记录数
    SELECT COUNT(*) AS total_can_migrate
    FROM ttpos_sale_order_buffet_customer_type sobct
    INNER JOIN ttpos_buffet_customer_type_price bctp ON sobct.buffet_customer_type_price_uuid = bctp.uuid
    INNER JOIN ttpos_buffet_customer_type bct ON bctp.customer_type_uuid = bct.uuid
    WHERE (sobct.name = '' OR sobct.name NOT LIKE '{%')
      AND sobct.buffet_customer_type_price_uuid != 0 
      AND bct.name != ''
      AND sobct.deleted_at IS NULL;
    ```
  - Success: 检查脚本创建成功，统计结果准确

### 6.2 数据迁移脚本（可选）

- [ ] 6.2 编写数据迁移脚本（可选）

  - File: `scripts/migrate_buffet_customer_type_name_snapshot.sql`
  - Purpose: 从关联表补充历史订单的 `name` 快照字段
  - Requirements: Requirement 1
  - Leverage: Task 6.1 的检查脚本
  - SQL:
    ```sql
    -- 补充历史订单的顾客类型名称快照（仅迁移关联表数据存在的记录）
    -- 注意：BuffetCustomerType.Name 是单语言字段，需要转换为多语言 JSON
    UPDATE ttpos_sale_order_buffet_customer_type sobct
    INNER JOIN ttpos_buffet_customer_type_price bctp ON sobct.buffet_customer_type_price_uuid = bctp.uuid
    INNER JOIN ttpos_buffet_customer_type bct ON bctp.customer_type_uuid = bct.uuid
    SET sobct.name = JSON_OBJECT(
        'zh', IFNULL(bct.name, ''),
        'th', IFNULL(bct.name, ''),
        'en', IFNULL(bct.name, ''),
        'zhtw', IFNULL(bct.name, ''),
        'ja', IFNULL(bct.name, ''),
        'ko', IFNULL(bct.name, ''),
        'my', IFNULL(bct.name, ''),
        'tr', IFNULL(bct.name, ''),
        'sv', IFNULL(bct.name, '')
    )
    WHERE (sobct.name = '' OR sobct.name NOT LIKE '{%')
      AND sobct.buffet_customer_type_price_uuid != 0 
      AND bct.name != ''
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
  - Success: 生产环境迁移执行成功，字段类型已修改为 TEXT

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

