# 原因信息快照修复 任务分解

> 本文档定义原因信息快照修复功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 16  
**已完成**: 7  
**进行中**: -  
**完成率**: 43.75%

---

## Phase 1: 数据库设计和迁移

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: Requirement 1）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建数据库迁移脚本（JSON 方案）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_reason_name_to_sale_order_product_reason.php`
  - Purpose: 在 ttpos_sale_order_product_reason 表添加 name 快照字段（JSON 格式，包含所有语言）
  - Requirements: Requirement 1（数据库结构变更）
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考 `story-main-order-source-snapshot-fix` 的迁移文件
  - SQL:
    ```sql
    ALTER TABLE `ttpos_sale_order_product_reason` 
    ADD COLUMN `name` TEXT NOT NULL DEFAULT '' 
    COMMENT '原因名称快照（JSON），不随后台更新' 
    AFTER `gift_reason_uuid`;
    ```
  - Note: **JSON 方案** - 快照包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
  - Success: 迁移文件创建成功，字段类型为 TEXT，注释正确

- [ ] 1.2 执行数据库迁移（测试环境）

  - File: -
  - Purpose: 在测试环境数据库中添加 name 字段
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加，测试环境验证通过

- [x] 1.3 修改 Go Model - 添加 Name 字段（JSON 方案）

  - File: `main/app/model/order.go`
  - Purpose: 在 SaleOrderProductReason 结构体添加 Name 字段，存储 JSON 格式快照
  - Requirements: Requirement 2（数据模型修改）
  - Leverage: 现有 Model: `main/app/model/order.go`，参考 `SaleBill.OrderSourceName` 字段定义
  - Code:
    ```go
    Name string `gorm:"column:name;type:text;default:'';comment:原因名称快照（JSON），不随后台更新" json:"name"`
    ```
  - Position: 紧跟 `GiftReasonUuid` 字段之后
  - Note: **JSON 方案** - 字段存储完整多语言 JSON
  - Success: 字段添加成功，GORM 和 JSON 标签正确，编译通过

---

## Phase 2: 查询逻辑修改

- [x] 2.1 修改 GetFreeReason() 方法（JSON 方案）

  - File: `main/app/model/sale_order_ext_getset.go`
  - Purpose: 实现免单原因获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 3（查询逻辑修改 - 免单原因）
  - Leverage: 参考 `GetLocaleOrderSourceName()` 方法的实现（`main/app/model/sale_bill.go`）
  - Key Logic (JSON 方案):
    1. 优先使用 `SaleOrderProductReason.Name` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `FreeReason.MultiLanguageName`
    4. 都为空时返回空响应
  - Code Reference:
    ```go
    // 参考 GetLocaleOrderSourceName() 的实现
    func (model *SaleOrder) GetFreeReason() dto.LocaleResponse {
        // 实现逻辑（参考 design.md）
        // 1. 遍历 FreeReasons
        // 2. 优先使用快照字段（JSON）
        // 3. 降级使用关联表数据
        // 4. 处理自定义免单原因
    }
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetFreeReason() 方法，优先使用快照字段（JSON），降级使用关联表数据 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，处理 JSON 解析和降级逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有方法签名 | Success: 方法修改完成，逻辑正确，编译通过，单元测试通过
  - Success: 方法修改完成，逻辑正确，编译通过

- [x] 2.2 修改 GetCancelReason() 方法（JSON 方案）

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 实现退菜原因获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 4（查询逻辑修改 - 退菜原因）
  - Leverage: 参考 `GetLocaleOrderSourceName()` 方法的实现（`main/app/model/sale_bill.go`）
  - Key Logic (JSON 方案):
    1. 优先使用 `SaleOrderProductReason.Name` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `ReturnFoodReason.MultiLanguageName`
    4. 都为空时返回空响应
  - Code Reference:
    ```go
    // 参考 GetLocaleOrderSourceName() 的实现
    func (model *SaleOrderProduct) GetCancelReason() dto.LocaleResponse {
        // 实现逻辑（参考 design.md）
        // 1. 遍历 CancelReasons
        // 2. 优先使用快照字段（JSON）
        // 3. 降级使用关联表数据
        // 4. 处理自定义退菜原因
    }
    ```
  - Prompt: Role: Go Developer | Task: 修改 GetCancelReason() 方法，优先使用快照字段（JSON），降级使用关联表数据 | Context: 参考 GetLocaleOrderSourceName() 的实现模式，处理 JSON 解析和降级逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有方法签名 | Success: 方法修改完成，逻辑正确，编译通过，单元测试通过
  - Success: 方法修改完成，逻辑正确，编译通过

- [ ] 2.3 编写 GetFreeReason() 和 GetCancelReason() 单元测试

  - File: `main/app/model/sale_order_ext_getset_test.go` 和 `main/app/model/sale_order_product_test.go`
  - Purpose: 测试快照字段存在/不存在、JSON 解析成功/失败、关联表存在/不存在等场景
  - Requirements: Requirement 3, Requirement 4
  - Leverage: 参考 `main/app/model/sale_bill_order_source_test.go` 的测试实现
  - Test Cases:
    - 快照字段存在且 JSON 有效 → 返回快照数据
    - 快照字段存在但 JSON 无效 → 降级使用关联表数据
    - 快照字段为空 → 降级使用关联表数据
    - 关联表数据不存在 → 返回空响应
    - 多个原因组合 → 返回组合结果
  - Prompt: Role: QA Automation Engineer | Task: 编写 GetFreeReason() 和 GetCancelReason() 单元测试 | Context: 测试快照字段存在/不存在、JSON 解析成功/失败、关联表存在/不存在等场景 | Restrictions: 测试覆盖率 100% | Success: 单元测试通过，覆盖率达标
  - Success: 单元测试通过，覆盖率达标

---

## Phase 3: 下单逻辑修改

- [x] 3.1 修改 NewFreeOrderReason() 方法（JSON 方案）

  - File: `main/app/model/sale_order.go`
  - Purpose: 修改免单原因创建方法，保存快照字段（JSON 格式）
  - Requirements: Requirement 5（下单逻辑修改 - 免单原因）
  - Leverage: 参考 `SetOrderSourceNameSnapshot()` 方法的实现（`main/app/model/sale_bill.go`）
  - Key Logic (JSON 方案):
    1. 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
    2. 序列化为 JSON 字符串
    3. 保存到 `SaleOrderProductReason.Name` 字段
    4. 处理序列化失败的情况（记录日志，但不中断流程）
  - Code Reference:
    ```go
    // 参考 SetOrderSourceNameSnapshot() 的实现
    func (model *SaleOrder) NewFreeOrderReason(freeReasons []*FreeReason) []*SaleOrderProductReason {
        // 实现逻辑（参考 design.md）
        // 1. 遍历 freeReasons
        // 2. 序列化多语言数据为 JSON
        // 3. 保存到 Name 字段
    }
    ```
  - Prompt: Role: Go Developer | Task: 修改 NewFreeOrderReason() 方法，保存快照字段（JSON 格式） | Context: 从 FreeReason.MultiLanguageName 获取完整多语言数据，序列化为 JSON 保存，处理序列化失败的情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有下单流程 | Success: 方法修改完成，逻辑正确，编译通过
  - Success: 方法修改完成，逻辑正确，编译通过

- [x] 3.2 修改 NewSaleOrderProductReasonList() 方法（JSON 方案）

  - File: `main/app/model/sale_order_product.go`
  - Purpose: 修改退菜原因创建方法，保存快照字段（JSON 格式）
  - Requirements: Requirement 6（下单逻辑修改 - 退菜原因）
  - Leverage: 参考 `SetOrderSourceNameSnapshot()` 方法的实现（`main/app/model/sale_bill.go`）
  - Key Logic (JSON 方案):
    1. 从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据
    2. 序列化为 JSON 字符串
    3. 保存到 `SaleOrderProductReason.Name` 字段
    4. 处理序列化失败的情况（记录日志，但不中断流程）
  - Code Reference:
    ```go
    // 参考 SetOrderSourceNameSnapshot() 的实现
    func (model *SaleOrderProduct) NewSaleOrderProductReasonList(reasons []*ReturnFoodReason) []*SaleOrderProductReason {
        // 实现逻辑（参考 design.md）
        // 1. 遍历 reasons
        // 2. 序列化多语言数据为 JSON
        // 3. 保存到 Name 字段
    }
    ```
  - Prompt: Role: Go Developer | Task: 修改 NewSaleOrderProductReasonList() 方法，保存快照字段（JSON 格式） | Context: 从 ReturnFoodReason.MultiLanguageName 获取完整多语言数据，序列化为 JSON 保存，处理序列化失败的情况 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有下单流程 | Success: 方法修改完成，逻辑正确，编译通过
  - Success: 方法修改完成，逻辑正确，编译通过

- [ ] 3.3 编写下单集成测试

  - File: `test/integration/order_test.go` 或相关测试文件
  - Purpose: 测试下单流程是否正确保存原因快照
  - Requirements: Requirement 5, Requirement 6
  - Leverage: 现有集成测试，参考外卖来源快照修复的集成测试
  - Test Cases:
    - 创建免单原因 → 验证快照字段保存成功（JSON 格式）
    - 创建退菜原因 → 验证快照字段保存成功（JSON 格式）
    - 删除原因配置 → 查询订单仍显示快照数据
  - Prompt: Role: QA Automation Engineer | Task: 编写下单集成测试，验证原因快照保存 | Context: 创建免单/退菜原因，验证 SaleOrderProductReason.Name 字段保存成功（JSON 格式），删除原因配置后订单仍显示快照数据 | Restrictions: 测试真实下单场景 | Success: 集成测试通过

---

## Phase 4: 查询逻辑修改

- [x] 4.1 梳理所有订单查询接口

  - File: -
  - Purpose: 梳理所有查询订单免单/退菜原因的地方
  - Requirements: Requirement 3, Requirement 4
  - Leverage: 搜索代码库中使用 `GetFreeReason()` 或 `GetCancelReason()` 的地方
  - Search Command:
    ```bash
    cd main && grep -r "GetFreeReason\|GetCancelReason" app/service/ app/api/
    ```
  - Success: 所有查询接口已梳理，列出清单

- [x] 4.2 修改订单查询逻辑 - 使用快照数据

  - File: 根据 Task 4.1 梳理的结果，逐个修改
  - Purpose: 确保所有订单查询接口都使用快照数据（通过 GetFreeReason() 和 GetCancelReason() 方法）
  - Requirements: Requirement 3, Requirement 4
  - Leverage: Task 2.1 和 2.2 修改的 GetFreeReason() 和 GetCancelReason() 方法
  - Key Logic:
    ```go
    // 原有逻辑（错误）
    // 直接从 reason.MultiLanguageName 获取
    
    // 新逻辑（正确）
    // 通过 GetFreeReason() 和 GetCancelReason() 方法获取
    // 这些方法内部已经使用快照数据
    ```
  - Prompt: Role: Go Developer | Task: 修改订单查询逻辑，确保使用快照数据 | Context: 所有查询接口都通过 GetFreeReason() 和 GetCancelReason() 方法获取原因信息，这些方法内部已经使用快照数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询接口都使用快照数据，历史订单正常显示
  - Success: 所有查询接口都使用快照数据，历史订单正常显示

- [ ] 4.3 编写查询集成测试

  - File: `test/integration/order_test.go` 或相关测试文件
  - Purpose: 测试查询流程是否正确使用快照数据
  - Requirements: Requirement 3, Requirement 4
  - Leverage: 现有集成测试，参考外卖来源快照修复的集成测试
  - Test Cases:
    - 创建免单原因 → 查询订单 → 验证使用快照数据
    - 创建退菜原因 → 查询订单 → 验证使用快照数据
    - 删除原因配置 → 查询订单 → 验证仍显示快照数据
    - 修改原因名称 → 查询订单 → 验证仍显示修改前的名称
    - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常
  - Prompt: Role: QA Automation Engineer | Task: 编写查询集成测试，验证原因快照查询 | Context: 测试多个场景（原因删除、原因改名、历史订单兼容），验证查询逻辑使用快照数据 | Restrictions: 测试真实查询场景 | Success: 集成测试通过

---

## Phase 5: 数据迁移和兼容性

- [ ] 5.1 编写数据检查脚本

  - File: `scripts/check_reason_snapshot_data.sql` 或相关脚本
  - Purpose: 检查历史订单的原因快照字段填充情况
  - Requirements: Requirement 6（数据迁移和兼容性）
  - Leverage: 参考外卖来源快照修复的数据检查脚本
  - SQL:
    ```sql
    -- 检查快照字段填充率
    SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN name != '' THEN 1 ELSE 0 END) as with_snapshot,
        SUM(CASE WHEN name = '' THEN 1 ELSE 0 END) as without_snapshot
    FROM ttpos_sale_order_product_reason
    WHERE delete_time = 0
    AND (free_reason_uuid != 0 OR return_food_reason_uuid != 0);
    ```
  - Success: 数据检查脚本编写完成，执行成功

- [ ] 5.2 编写数据迁移脚本（可选）

  - File: `scripts/migrate_reason_snapshot_data.sql` 或相关脚本
  - Purpose: 从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
  - Requirements: Requirement 6（数据迁移和兼容性）
  - Leverage: 参考外卖来源快照修复的数据迁移脚本
  - SQL Strategy:
    ```sql
    -- 迁移免单原因快照（仅迁移关联表数据存在的记录）
    UPDATE ttpos_sale_order_product_reason sor
    INNER JOIN ttpos_free_reason fr ON sor.free_reason_uuid = fr.uuid
    INNER JOIN ttpos_multi_language_name mln ON fr.multi_language_name_uuid = mln.uuid
    SET sor.name = JSON_OBJECT(
        'zh', mln.zh_name,
        'th', mln.th_name,
        'en', mln.en_name,
        'zhtw', mln.zh_tw_name,
        'ja', mln.ja_name,
        'ko', mln.ko_name,
        'my', mln.my_name,
        'tr', mln.tr_name,
        'sv', mln.sv_name
    )
    WHERE sor.delete_time = 0
    AND sor.free_reason_uuid != 0
    AND sor.name = '';
    
    -- 迁移退菜原因快照（仅迁移关联表数据存在的记录）
    UPDATE ttpos_sale_order_product_reason sor
    INNER JOIN ttpos_return_food_reason rfr ON sor.return_food_reason_uuid = rfr.uuid
    INNER JOIN ttpos_multi_language_name mln ON rfr.multi_language_name_uuid = mln.uuid
    SET sor.name = JSON_OBJECT(
        'zh', mln.zh_name,
        'th', mln.th_name,
        'en', mln.en_name,
        'zhtw', mln.zh_tw_name,
        'ja', mln.ja_name,
        'ko', mln.ko_name,
        'my', mln.my_name,
        'tr', mln.tr_name,
        'sv', mln.sv_name
    )
    WHERE sor.delete_time = 0
    AND sor.return_food_reason_uuid != 0
    AND sor.name = '';
    ```
  - Note: 迁移脚本只处理关联表数据存在的记录，对于已删除的数据，保持快照字段为空（使用降级逻辑）
  - Success: 数据迁移脚本编写完成，执行成功

- [ ] 5.3 执行数据迁移（可选，测试环境）

  - File: -
  - Purpose: 在测试环境执行数据迁移，验证迁移结果
  - Requirements: Requirement 6
  - Leverage: Task 5.2 的迁移脚本
  - Command: 在测试环境执行迁移脚本
  - Success: 迁移执行成功，数据验证通过

---

## Phase 6: 回归测试和上线

- [ ] 6.1 回归测试 - 订单查询接口

  - File: -
  - Purpose: 确保所有订单查询接口正常工作
  - Requirements: 非功能需求（测试要求）
  - Test Cases:
    - 订单详情查询
    - 订单列表查询
    - 订单报表查询
    - 订单导出功能
  - Success: 所有订单查询接口测试通过

- [ ] 6.2 回归测试 - 订单打印/导出/报表

  - File: -
  - Purpose: 确保订单打印、导出、报表功能正常工作
  - Requirements: 非功能需求（测试要求）
  - Test Cases:
    - 订单打印（包含免单/退菜原因）
    - 订单导出（包含免单/退菜原因）
    - 订单报表（包含免单/退菜原因）
  - Success: 所有打印/导出/报表功能测试通过

- [ ] 6.3 执行生产环境迁移

  - File: -
  - Purpose: 在生产环境执行数据库迁移和数据迁移（如需要）
  - Requirements: Requirement 1, Requirement 6
  - Leverage: Task 1.1 的迁移文件和 Task 5.2 的数据迁移脚本
  - Pre-requisites:
    - 测试环境验证通过
    - 备份生产数据库
    - 选择低峰期执行
  - Success: 生产环境迁移执行成功，业务正常

---

## 📝 注意事项

### JSON 序列化/反序列化

- 使用标准库 `encoding/json` 进行序列化/反序列化
- 序列化失败时记录日志，但不中断流程
- 反序列化失败时降级使用关联表数据

### 兼容性处理

- 历史订单的快照字段可能为空，必须实现降级逻辑
- 降级逻辑使用关联表数据，确保历史订单正常显示
- 新订单自动使用快照机制

### 测试覆盖

- 单元测试覆盖所有修改的方法
- 集成测试覆盖核心流程（下单、查询）
- 回归测试确保不影响现有功能

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: xiezhihuan  
**审核者**: {审核者}

