# 外卖来源信息快照修复 任务分解

> 本文档定义外卖来源信息快照修复功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 14  
**已完成**: 9  
**进行中**: -  
**完成率**: 64.3%

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

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_order_source_name_to_sale_bill.php`
  - Purpose: 在 ttpos_sale_bill 表添加 order_source_name 快照字段（JSON 格式，包含所有语言）
  - Requirements: Requirement 1（数据库结构变更）
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考 `story-main-nationality-snapshot-fix` 的迁移文件
  - SQL:
    ```sql
    ALTER TABLE `ttpos_sale_bill` 
    ADD COLUMN `order_source_name` TEXT NOT NULL DEFAULT '' 
    COMMENT '外卖来源名称快照（JSON），不随后台更新' 
    AFTER `order_source_uuid`;
    ```
  - Note: **JSON 方案** - 快照包含所有语言（ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
  - Success: 迁移文件创建成功，字段类型为 TEXT，注释正确

- [ ] 1.2 执行数据库迁移（测试环境）

  - File: -
  - Purpose: 在测试环境数据库中添加 order_source_name 字段
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加，测试环境验证通过

- [x] 1.3 修改 Go Model - 添加 OrderSourceName 字段（JSON 方案）

  - File: `main/app/model/sale_bill.go`
  - Purpose: 在 SaleBill 结构体添加 OrderSourceName 字段，存储 JSON 格式快照
  - Requirements: Requirement 2（数据模型修改）
  - Leverage: 现有 Model: `main/app/model/sale_bill.go`，参考 `NationalityName` 字段定义
  - Code:
    ```go
    OrderSourceName string `gorm:"column:order_source_name;type:text" json:"order_source_name"`
    ```
  - Position: 紧跟 `OrderSourceUuid` 字段之后
  - Note: **JSON 方案** - 字段存储完整多语言 JSON
  - Success: 字段添加成功，GORM 和 JSON 标签正确，编译通过

---

## Phase 2: 核心实现 - Model 层

- [x] 2.1 实现 GetLocaleOrderSourceName() 方法（JSON 方案）

  - File: `main/app/model/sale_bill.go`
  - Purpose: 实现外卖来源名称获取方法，解析 JSON 快照，支持降级兼容
  - Requirements: Requirement 3（查询逻辑修改）
  - Leverage: 参考 `GetLocaleNationalityName()` 方法的实现（`main/app/model/sale_bill.go:702-730`）
  - Key Logic (JSON 方案):
    1. 优先使用 `OrderSourceName` 快照字段（JSON）
    2. 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
    3. 快照为空或解析失败时，降级使用 `OrderSource.MultiLanguageName`
    4. 都为空时返回空响应
  - Code Reference:
    ```go
    // 参考 GetLocaleNationalityName() 的实现
    func (model *SaleBill) GetLocaleOrderSourceName() dto.LocaleResponse {
        // 实现逻辑（参考 design.md）
    }
    ```
  - Success: 方法实现完成，逻辑正确，编译通过

- [x] 2.2 实现 SetOrderSourceNameSnapshot() 方法（JSON 方案）

  - File: `main/app/model/sale_bill.go`
  - Purpose: 实现设置外卖来源名称快照方法，序列化多语言数据为 JSON
  - Requirements: Requirement 4（下单逻辑修改）
  - Leverage: 参考 `SetNationalityNameSnapshot()` 方法的实现（`main/app/model/sale_bill.go:732-753`）
  - Key Logic:
    1. 从 `MultiLanguageName` 获取完整多语言数据
    2. 构建 `LocaleResponse`
    3. 序列化为 JSON 保存到 `OrderSourceName` 字段
  - Code Reference:
    ```go
    // 参考 SetNationalityNameSnapshot() 的实现
    func (model *SaleBill) SetOrderSourceNameSnapshot(multiLangName MultiLanguageName) error {
        // 实现逻辑（参考 design.md）
    }
    ```
  - Success: 方法实现完成，逻辑正确，编译通过

- [x] 2.3 编写 GetLocaleOrderSourceName() 和 SetOrderSourceNameSnapshot() 单元测试（JSON 方案）

  - File: `main/app/model/sale_bill_order_source_test.go`（新建文件）
  - Purpose: 确保 GetLocaleOrderSourceName() 和 SetOrderSourceNameSnapshot() 方法逻辑正确，覆盖所有场景
  - Requirements: Requirement 3, Requirement 4
  - Leverage: 参考 `sale_bill_nationality_test.go` 的测试用例
  - Test Cases (JSON 方案):
    - GetLocaleOrderSourceName():
      - 场景1: 快照有值（JSON）+ 关联表存在 → 使用快照数据（所有语言）
      - 场景2: 快照有值（JSON）+ 关联表不存在 → 使用快照数据（所有语言）
      - 场景3: 快照为空 + 关联表存在 → 降级使用关联表数据
      - 场景4: 快照为空 + 关联表不存在 → 返回空响应
      - 场景5: 快照 JSON 格式错误 → 降级使用关联表数据
    - SetOrderSourceNameSnapshot():
      - 场景1: 正常设置快照 → JSON 格式正确
      - 场景2: 多语言名称为空 → 设置为空字符串
  - Success: 单元测试创建完成，所有测试用例通过，覆盖率 100%

---

## Phase 3: 下单逻辑修改

- [x] 3.1 梳理所有下单入口

  - File: -
  - Purpose: 梳理所有创建订单的地方，确保都需要保存快照
  - Requirements: Requirement 4（下单逻辑修改）
  - Leverage: 搜索代码库中创建 `SaleBill` 的地方
  - Search Command:
    ```bash
    cd main && grep -r "SaleBill{" app/service/ app/api/
    grep -r "OrderSourceUuid" app/service/ app/api/
    ```
  - Success: 所有下单入口已梳理，列出清单

- [x] 3.2 修改下单逻辑 - 保存外卖来源名称快照

  - File: 根据 Task 3.1 梳理的结果，逐个修改
  - Purpose: 在创建 SaleBill 时，从 OrderSource.MultiLanguageName 获取完整多语言数据，序列化为 JSON 保存到 OrderSourceName 字段
  - Requirements: Requirement 4
  - Leverage: 参考 `SetOrderSourceNameSnapshot()` 方法，参考国籍快照修复的下单逻辑修改
  - Key Logic:
    ```go
    // 在创建 SaleBill 时添加
    saleBill := &model.SaleBill{
        // ... 其他字段
        OrderSourceUuid: orderSourceUuid,
        // ... 其他字段
    }
    
    // 如果外卖来源存在，设置快照
    if orderSourceUuid != 0 && orderSource != nil && orderSource.MultiLanguageName != nil {
        if err := saleBill.SetOrderSourceNameSnapshot(orderSource.MultiLanguageName); err != nil {
            // 记录错误日志，但不影响下单流程
            ctx.Log().Error("保存外卖来源名称快照失败", zap.Error(err))
        }
    }
    ```
  - Prompt: Role: Go Developer | Task: 修改下单逻辑，在创建 SaleBill 时保存外卖来源名称快照 | Context: 使用 SetOrderSourceNameSnapshot() 方法，从 OrderSource.MultiLanguageName 获取完整多语言数据，序列化为 JSON 保存，处理边界情况（外卖来源不存在、为空） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有下单流程 | Success: 所有下单入口都保存快照，边界情况处理正确

- [ ] 3.3 编写下单集成测试

  - File: `test/integration/order_test.go` 或相关测试文件
  - Purpose: 测试下单流程是否正确保存外卖来源名称快照
  - Requirements: Requirement 4
  - Leverage: 现有集成测试，参考国籍快照修复的集成测试
  - Test Cases:
    - 创建订单 → 验证快照字段保存成功（JSON 格式）
    - 删除外卖来源配置 → 查询订单仍显示快照数据
  - Prompt: Role: QA Automation Engineer | Task: 编写下单集成测试，验证外卖来源名称快照保存 | Context: 创建订单，验证 SaleBill.OrderSourceName 字段保存成功（JSON 格式），删除外卖来源配置后订单仍显示快照数据 | Restrictions: 测试真实下单场景 | Success: 集成测试通过

---

## Phase 4: 查询逻辑修改

- [x] 4.1 梳理所有订单查询接口

  - File: -
  - Purpose: 梳理所有查询订单外卖来源名称的地方
  - Requirements: Requirement 5（订单查询逻辑修改）
  - Leverage: 搜索代码库中使用 `OrderSource.MultiLanguageName` 或 `OrderSourceName` 的地方
  - Search Command:
    ```bash
    cd main && grep -r "OrderSource.MultiLanguageName" app/service/ app/api/
    grep -r "OrderSourceName" app/service/ app/api/
    ```
  - Success: 所有查询接口已梳理，列出清单

- [x] 4.2 修改订单查询逻辑 - 使用 GetLocaleOrderSourceName()

  - File: 根据 Task 4.1 梳理的结果，逐个修改
  - Purpose: 替换原有的直接从 OrderSource.MultiLanguageName 获取的逻辑，使用 GetLocaleOrderSourceName() 方法
  - Requirements: Requirement 5
  - Leverage: Task 2.1 实现的 GetLocaleOrderSourceName() 方法
  - Key Logic:
    ```go
    // 原有逻辑（错误）
    if saleBill.OrderSource != nil && saleBill.OrderSource.MultiLanguageName != nil {
        orderSourceName = saleBill.OrderSource.MultiLanguageName.GetNames()
    }
    
    // 新逻辑（正确）
    orderSourceName = saleBill.GetLocaleOrderSourceName()
    ```
  - Prompt: Role: Go Developer | Task: 修改订单查询逻辑，使用 GetLocaleOrderSourceName() 方法获取外卖来源名称 | Context: 替换原有的直接从 OrderSource.MultiLanguageName 获取的逻辑，确保所有查询接口都使用快照数据 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有查询接口都使用快照数据，历史订单正常显示

- [ ] 4.3 编写查询集成测试

  - File: `test/integration/order_test.go` 或相关测试文件
  - Purpose: 测试查询流程是否正确使用快照数据
  - Requirements: Requirement 5
  - Leverage: 现有集成测试，参考国籍快照修复的集成测试
  - Test Cases:
    - 创建订单 → 查询订单 → 验证使用快照数据
    - 删除外卖来源配置 → 查询订单 → 验证仍显示快照数据
    - 修改外卖来源名称 → 查询订单 → 验证仍显示修改前的名称
    - 历史订单（快照为空） → 查询订单 → 验证降级逻辑正常
  - Prompt: Role: QA Automation Engineer | Task: 编写查询集成测试，验证外卖来源名称快照查询 | Context: 测试多个场景（外卖来源删除、外卖来源改名、历史订单兼容），验证查询逻辑使用快照数据 | Restrictions: 测试真实查询场景 | Success: 集成测试通过

---

## Phase 5: 数据迁移和兼容性

- [ ] 5.1 编写数据检查脚本

  - File: `scripts/check_order_source_snapshot.sql`
  - Purpose: 检查历史订单的 order_source_name 字段填充情况
  - Requirements: Requirement 6（数据迁移和兼容性处理）
  - Leverage: 参考 `scripts/check_nationality_snapshot.sql`
  - SQL:
    ```sql
    -- 统计需要迁移的订单数量
    SELECT COUNT(*) AS total_need_migrate
    FROM ttpos_sale_bill sb
    WHERE sb.order_source_name = '' 
      AND sb.order_source_uuid != 0 
      AND sb.deleted_at IS NULL;
    
    -- 统计关联表数据存在的记录数
    SELECT COUNT(*) AS total_can_migrate
    FROM ttpos_sale_bill sb
    INNER JOIN ttpos_order_source os ON sb.order_source_uuid = os.uuid
    INNER JOIN ttpos_multi_language_name mln ON os.multi_language_name_uuid = mln.uuid
    WHERE sb.order_source_name = '' 
      AND sb.order_source_uuid != 0 
      AND mln.zh_name != ''
      AND sb.deleted_at IS NULL;
    ```
  - Success: 检查脚本创建成功，统计结果准确

- [ ] 5.2 编写数据迁移脚本（可选）

  - File: `scripts/migrate_order_source_snapshot.sql`
  - Purpose: 从关联表补充历史订单的快照字段（JSON 格式）
  - Requirements: Requirement 6
  - Leverage: 参考 design.md 中的迁移 SQL
  - SQL:
    ```sql
    -- 补充历史订单的外卖来源名称快照（仅迁移关联表数据存在的记录）
    -- 注意：需要序列化为 JSON 格式
    UPDATE ttpos_sale_bill sb
    INNER JOIN ttpos_order_source os ON sb.order_source_uuid = os.uuid
    INNER JOIN ttpos_multi_language_name mln ON os.multi_language_name_uuid = mln.uuid
    SET sb.order_source_name = JSON_OBJECT(
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
    WHERE sb.order_source_name = '' 
      AND sb.order_source_uuid != 0 
      AND mln.zh_name != ''
      AND sb.deleted_at IS NULL;
    ```
  - Note: 使用 MySQL 的 `JSON_OBJECT()` 函数生成 JSON 格式
  - Success: 迁移脚本创建成功，支持幂等性

- [ ] 5.3 执行数据迁移（可选，测试环境）

  - File: -
  - Purpose: 在测试环境执行数据迁移，验证迁移结果
  - Requirements: Requirement 6
  - Leverage: Task 5.2 的迁移脚本
  - Command: 执行 SQL 脚本
  - Success: 迁移执行成功，迁移结果正确，JSON 格式验证通过

---

## Phase 6: 测试和验证

- [ ] 6.1 回归测试 - 订单查询接口

  - File: -
  - Purpose: 确保订单查询接口不受影响
  - Requirements: 所有功能需求
  - Leverage: 现有测试用例
  - Test Cases:
    - 订单详情查询
    - 订单列表查询
    - 订单筛选查询
  - Success: 所有测试通过

- [ ] 6.2 回归测试 - 订单打印/导出/报表

  - File: -
  - Purpose: 确保订单打印、导出、报表功能不受影响
  - Requirements: 所有功能需求
  - Leverage: 现有测试用例
  - Test Cases:
    - 订单打印（小票、发票）
    - 订单导出（Excel）
    - 订单报表（日报、月报）
  - Success: 所有测试通过

- [ ] 6.3 执行生产环境迁移

  - File: -
  - Purpose: 在生产环境执行数据库迁移
  - Requirements: Requirement 1
  - Leverage: Task 1.1 的迁移文件
  - Pre-check:
    - [ ] 备份生产数据库
    - [ ] 测试环境验证通过
    - [ ] 准备回滚方案
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，生产环境验证通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 单元测试覆盖率达标:
  - GetLocaleOrderSourceName(): 100%
  - SetOrderSourceNameSnapshot(): 100%
- [ ] 所有测试通过（单元测试、集成测试、回归测试）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足:
  - Requirement 1: 数据库结构变更 ✅
  - Requirement 2: 数据模型修改 ✅
  - Requirement 3: 查询逻辑修改 ✅
  - Requirement 4: 下单逻辑修改 ✅
  - Requirement 5: 订单查询逻辑修改 ✅
  - Requirement 6: 数据迁移和兼容性处理 ✅
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成（10 项功能验收）

### 文档同步

- [ ] 数据库文档已更新（表结构说明）
- [ ] CHANGELOG.md 已更新
- [ ] API 文档已更新（如订单查询接口响应有变化）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-order-source-snapshot-fix/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-order-source-snapshot-fix/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-order-source-snapshot-fix/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-order-source-snapshot-fix/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-order-source-snapshot-fix/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/xiezhihuan/2025-12/2025-12-02.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-02  
**维护者**: xiezhihuan

