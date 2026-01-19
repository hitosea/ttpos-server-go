# 自助餐名称信息快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-08 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-buffet-package-name-snapshot-fix](../../shared/specs/archived/v2.12/story-main-buffet-package-name-snapshot-fix/) |
| **父提案** | `order-attribute-snapshot-fix.md` |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，自助餐名称信息会随后台数据变更而改变，导致订单历史信息不准确。这是订单商品信息快照修复需求（`order-attribute-snapshot-fix.md`）的子任务。

**具体场景**：

1. **自助餐名称被删除**：
   - 订单自助餐："豪华自助餐"
   - 后台删除了"豪华自助餐"配置
   - 查询订单时自助餐名称信息丢失或显示错误

2. **自助餐名称被改名**：
   - 订单自助餐："豪华自助餐"
   - 后台将"豪华自助餐"改名为"超值自助餐"
   - 查询订单时显示："超值自助餐"（显示的是新名称，而非下单时的名称）

3. **自助餐套餐组合变化**：
   - 订单包含两个自助餐套餐："豪华自助餐" + "儿童自助餐"
   - 后台删除了其中一个套餐或改名
   - 查询订单时显示错误信息或新名称

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、统计报表等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 自助餐名称被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单自助餐名称信息准确反映下单时的状态
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **用户体验**：用户查看历史订单时看到的是下单时的真实自助餐名称信息
- ✅ **业务可靠性**：避免因数据变更导致的业务逻辑错误

### 目标用户

- [x] 收银员（查看历史订单）
- [x] 商户管理员（对账、报表）
- [ ] 厨房人员
- [ ] 顾客
- [x] 财务人员（对账、审计）

---

## 💡 解决方案概述

### 方案描述

**核心思路**：订单自助餐名称信息应该使用快照数据，而不是从关联表实时获取。

**现状分析**：

1. **数据库设计缺少快照字段**：
   - `ttpos_sale_bill` 表只有 `buffet_package1_uuid` 和 `buffet_package2_uuid` 字段，关联到 `ttpos_buffet_package` 表
   - 没有 `buffet_package1_name` 和 `buffet_package2_name` 快照字段
   - 查询时从 `BuffetPackage1.MultiLanguageName` 和 `BuffetPackage2.MultiLanguageName` 获取名称
   - 导致自助餐数据被删除或改名时，订单显示信息变化

2. **代码实现依赖关联表**：
   - `SaleBill.GetBuffetName()` 方法直接从 `BuffetPackage1.MultiLanguageName` 和 `BuffetPackage2.MultiLanguageName` 获取
   - `SaleBill.GetBuffetNames()` 方法从 `SaleOrderBuffetCustomerTypes` 中获取（订单级别的自助餐信息）
   - `SaleOrder.GetBuffetNames()` 方法也从 `SaleOrderBuffetCustomerTypes` 中获取
   - 没有快照机制

**解决方案**：

1. **新增快照字段**（需要数据库结构变更）：
   - 在 `ttpos_sale_bill` 表添加 `buffet_package1_name` 字段（TEXT 类型，存储 JSON），用于快照自助餐套餐1名称（多语言 JSON）
   - 在 `ttpos_sale_bill` 表添加 `buffet_package2_name` 字段（TEXT 类型，存储 JSON），用于快照自助餐套餐2名称（多语言 JSON）

2. **修复查询逻辑**：
   - 修改 `SaleBill.GetBuffetName()` 方法：优先使用 `BuffetPackage1Name`/`BuffetPackage2Name` 字段（快照）
   - 修改 `SaleBill.GetBuffetNames()` 方法：优先使用快照字段
   - 修改 `SaleOrder.GetBuffetNames()` 方法：优先使用快照字段
   - 如果快照字段为空，降级使用 `BuffetPackage.MultiLanguageName`（兼容历史数据）
   - 实现多语言支持：快照保存完整多语言 JSON，直接返回无需补充

3. **实施策略**：
   - **新订单**：下单时保存自助餐名称快照到 `buffet_package1_name`/`buffet_package2_name` 字段
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **数据迁移**：可选，从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

4. **多语言支持**：
   - **现状**：快照字段保存完整多语言 JSON，包含所有语言
   - **解决方案**：采用 JSON 方案
     - 快照字段保存完整多语言 JSON（包含所有语言：ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
     - 查询时直接解析 JSON 返回，无需从关联表补充
     - 如果快照为空或解析失败，降级使用关联表数据
     - 这样既保证了快照完整性（即使数据被删除也能显示），又提供了完整的多语言支持

### 核心功能点

#### 一、数据库结构变更

1. **新增快照字段**
   - 在 `ttpos_sale_bill` 表添加 `buffet_package1_name` 字段
   - 在 `ttpos_sale_bill` 表添加 `buffet_package2_name` 字段
   - 类型：TEXT（存储 JSON 格式）
   - 注释：自助餐套餐名称快照（JSON），不随后台更新
   - 默认值：空字符串

#### 二、代码修改

2. **修改数据模型**
   - 在 `SaleBill` 结构体添加 `BuffetPackage1Name` 字段
   - 在 `SaleBill` 结构体添加 `BuffetPackage2Name` 字段
   - 字段类型：`string`
   - JSON 标签：`json:"buffet_package1_name"` / `json:"buffet_package2_name"`
   - GORM 标签：`gorm:"column:buffet_package1_name;type:text"` / `gorm:"column:buffet_package2_name;type:text"`

3. **修复自助餐名称获取逻辑**
   - 新增 `SaleBill.GetLocaleBuffetPackage1Name()` 方法
   - 新增 `SaleBill.GetLocaleBuffetPackage2Name()` 方法
   - 修改 `SaleBill.GetBuffetName()` 方法：优先使用快照字段
   - 修改 `SaleBill.GetBuffetNames()` 方法：优先使用快照字段
   - 修改 `SaleOrder.GetBuffetNames()` 方法：优先使用快照字段
   - 实现逻辑：
     - 优先使用 `BuffetPackage1Name`/`BuffetPackage2Name` 字段（JSON）
     - 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
     - 快照为空或解析失败时，降级使用 `BuffetPackage.MultiLanguageName`
     - 构建多语言响应

4. **修改下单逻辑**
   - 在创建订单时，保存自助餐名称快照到 `BuffetPackage1Name`/`BuffetPackage2Name` 字段
   - 从 `BuffetPackage.MultiLanguageName` 获取完整多语言数据
   - 序列化为 JSON 保存
   - 确保所有下单入口都保存快照数据

5. **修改订单查询逻辑**
   - 在订单详情查询中，使用新的快照方法获取自助餐名称
   - 替换原有的直接从 `BuffetPackage.MultiLanguageName` 获取的逻辑

#### 三、数据迁移和兼容性

6. **数据完整性检查**
   - 检查历史订单的 `buffet_package1_name`/`buffet_package2_name` 字段填充情况
   - 统计需要迁移的订单数量

7. **兼容性处理**
   - 当快照字段为空时，降级使用关联表数据
   - 确保历史订单正常显示

8. **数据迁移**（可选）
   - 编写数据迁移脚本，从关联表补充历史订单的快照字段
   - 仅迁移关联表数据存在的记录
   - 支持可重复执行（幂等性）

#### 四、测试验证

9. **测试验证**
   - 验证自助餐删除后订单显示
   - 验证自助餐改名后订单显示
   - 验证自助餐套餐组合变化后订单显示
   - 验证历史订单兼容性
   - 验证多语言显示

### 影响范围

**涉及终端**：
- [x] POS 收银端（订单查询）
- [x] Shop 商家管理端（订单管理、对账）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端（历史订单）

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型（`SaleBill`）
- [x] 业务逻辑（名称获取方法）
- [ ] 第三方集成
- [x] 数据库迁移（新增快照字段）
- [x] 下单逻辑（保存快照数据）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要修改业务逻辑，涉及数据一致性
- [x] **高**：涉及数据库结构变更、业务逻辑修改、数据迁移

**说明**：
- 需要数据库结构变更（新增快照字段）
- 需要修改多个模型方法和查询逻辑
- 需要处理兼容性和数据迁移
- 需要修改下单逻辑，确保保存快照数据
- 需要充分测试确保不影响现有功能

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**任务分解**：
1. **数据库结构变更**（0.5 天）
   - 设计数据库迁移脚本
   - 添加快照字段到 `ttpos_sale_bill` 表
   - 执行迁移并验证

2. **代码修改 - 模型和方法**（1-1.5 天）
   - 修改 `SaleBill` 模型，添加快照字段
   - 实现 `GetLocaleBuffetPackage1Name()` 和 `GetLocaleBuffetPackage2Name()` 方法
   - 修改 `GetBuffetName()` 和 `GetBuffetNames()` 方法
   - 添加兼容性处理

3. **代码修改 - 下单和查询逻辑**（0.5 天）
   - 修改下单逻辑，确保保存快照数据
   - 修改订单查询逻辑，使用快照数据

4. **数据检查与迁移**（0.5 天）
   - 检查历史数据完整性
   - 编写数据迁移脚本（可选）
   - 执行数据迁移

5. **测试验证**（0.5 天）
   - 单元测试（覆盖所有修改的方法）
   - 集成测试（验证订单查询、下单保存快照）
   - 回归测试（确保不影响现有功能）

### 风险识别

**潜在风险**：

1. **数据库结构变更风险**
   - 风险：需要新增快照字段，涉及数据库迁移
   - 影响：需要设计迁移脚本，确保不影响现有数据
   - 缓解：使用 `ALTER TABLE ADD COLUMN` 添加可空字段，不影响现有数据

2. **历史数据不完整**
   - 风险：部分历史订单的快照字段可能为空
   - 影响：需要降级处理或数据迁移
   - 缓解：实现降级逻辑，历史数据可以逐步迁移

3. **多语言支持问题** ⚠️ **关键问题**
   - **问题**：快照字段需要保存完整多语言 JSON，但需要确保格式正确
   - **影响**：需要设计合理的多语言快照方案
   - **解决方案**：采用 JSON 方案
     - 快照字段保存完整多语言 JSON（包含所有语言）
     - 查询时直接解析 JSON 返回，无需从关联表补充
     - 如果快照为空或解析失败，降级使用关联表数据
     - 这样既保证了快照完整性（数据被删除也能显示），又提供了完整的多语言支持

4. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存所有快照字段
   - 影响：可能遗漏某些下单场景，导致快照数据不完整
   - 缓解：全面梳理所有下单入口，确保都保存快照数据

5. **性能影响**
   - 风险：如果快照字段为空，降级查询可能增加数据库查询
   - 影响：需要优化查询逻辑，优先使用快照数据

6. **回归风险**
   - 风险：修改核心方法可能影响其他功能（订单查询、打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能

**缓解措施**：

1. **数据检查**：
   - 先检查历史数据的所有快照字段填充情况
   - 根据检查结果决定是否需要数据迁移

2. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

3. **多语言处理**：
   - 采用 JSON 方案，快照保存完整多语言数据
   - 查询时直接解析 JSON，无需从关联表补充
   - 快照为空或解析失败时，降级使用关联表数据

4. **性能优化**：
   - 优先使用快照数据，减少关联查询
   - 如果必须降级查询，使用索引优化

5. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
   - 在测试环境充分验证后再上线

6. **全面梳理**：
   - 梳理所有使用自助餐名称的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 父提案: `order-attribute-snapshot-fix.md` - 订单商品信息快照修复
- 参考子任务: `nationality-snapshot-fix.md` - 国籍信息快照修复
- 参考子任务: `order-source-snapshot-fix.md` - 外卖来源信息快照修复

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/sale_bill.go`
- 自助餐名称获取方法: `main/app/model/sale_bill_ext_getset.go`

### 代码位置

**问题代码**：
- `main/app/model/sale_bill_ext_getset.go:328` - `GetBuffetName()` 方法（自助餐名称）
- `main/app/model/sale_bill_ext_getset.go:458` - `GetBuffetNames()` 方法（所有自助餐名称）
- `main/app/model/sale_order_ext_getset.go:82` - `SaleOrder.GetBuffetNames()` 方法

**数据模型**：
- `main/app/model/sale_bill.go:87-88` - `SaleBill` 模型定义（`BuffetPackage1Uuid`、`BuffetPackage2Uuid` 字段）
- `main/app/model/sale_bill.go:111-112` - `SaleBill` 关联定义（`BuffetPackage1`、`BuffetPackage2`）

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [x] 创建 Spec：`story-main-buffet-package-name-snapshot-fix` ✅ 2025-12-08
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}
- [ ] 产品审核：需求文档（requirements.md）
- [ ] 技术设计：使用 `/spec-design` 创建设计文档和任务分解

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实自助餐名称信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的自助餐名称  
**以便于** 准确处理客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含自助餐的订单 **THEN** 系统 **SHALL** 显示下单时保存的自助餐名称快照
2. **IF** 后台删除了某个自助餐配置 **THEN** 历史订单 **SHALL** 仍然显示该自助餐的原始名称
3. **IF** 后台修改了某个自助餐的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
4. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
5. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 正确保存自助餐名称快照（`BuffetPackage1Name`/`BuffetPackage2Name`）
6. **WHEN** 查询订单自助餐信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
7. **IF** 订单包含两个自助餐套餐 **THEN** 系统 **SHALL** 正确显示两个套餐的名称快照

### 技术方案要点（初稿）

#### JSON 快照策略

**核心思路**：完整多语言 JSON 快照

- **快照字段**：保存完整多语言 JSON（包含所有语言：ZH, EN, TH, ZHTW, JA, KO, MY, TR, SV）
- **查询逻辑**：
  - 优先使用快照字段（JSON）
  - 解析 JSON 为 `LocaleResponse`（包含所有语言）
  - 快照为空或解析失败时，降级使用关联表数据

#### 具体实现方案

1. **数据库迁移脚本**：
   ```sql
   -- 添加自助餐套餐1名称快照字段（JSON 方案）
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `buffet_package1_name` TEXT NOT NULL DEFAULT '' 
   COMMENT '自助餐套餐1名称快照（JSON），不随后台更新' 
   AFTER `buffet_package2_uuid`;
   
   -- 添加自助餐套餐2名称快照字段（JSON 方案）
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `buffet_package2_name` TEXT NOT NULL DEFAULT '' 
   COMMENT '自助餐套餐2名称快照（JSON），不随后台更新' 
   AFTER `buffet_package1_name`;
   ```

2. **修改数据模型**：
   ```go
   // main/app/model/sale_bill.go
   type SaleBill struct {
       // ... 其他字段
       BuffetPackage1Uuid uint64 `gorm:"column:buffet_package1_uuid" json:"buffet_package1_uuid"`
       BuffetPackage1Name string `gorm:"column:buffet_package1_name;type:text" json:"buffet_package1_name"` // 新增快照字段（JSON）
       BuffetPackage2Uuid uint64 `gorm:"column:buffet_package2_uuid" json:"buffet_package2_uuid"`
       BuffetPackage2Name string `gorm:"column:buffet_package2_name;type:text" json:"buffet_package2_name"` // 新增快照字段（JSON）
       // ... 其他字段
   }
   ```

3. **新增自助餐名称获取方法**：
   ```go
   // main/app/model/sale_bill.go
   
   // GetLocaleBuffetPackage1Name 获取自助餐套餐1名称（多语言）
   // 优先使用快照字段，降级使用关联表数据，支持多语言
   // 快照字段保存多语言（JSON）
   // Requirement: story-main-buffet-package-name-snapshot-fix
   func (model *SaleBill) GetLocaleBuffetPackage1Name() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.BuffetPackage1Name
       
       // 如果快照字段不为空，尝试反序列化为多语言数据
       if snapshotName != "" {
           var snapshotLocale dto.LocaleResponse
           if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
               // 反序列化成功，检查是否有主语言数据
               if !snapshotLocale.IsNull() {
                   // 如果快照数据完整，直接返回
                   return snapshotLocale
               }
           }
           // 如果反序列化失败或数据不完整，继续后续降级逻辑
       }
       
       // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
       if model.BuffetPackage1 != nil && !model.BuffetPackage1.MultiLanguageName.IsNullName() {
           return model.BuffetPackage1.MultiLanguageName.GetNames()
       }
       
       // 兜底：如果关联表也没有数据，返回空的多语言响应
       return dto.LocaleResponse{}
   }
   
   // GetLocaleBuffetPackage2Name 获取自助餐套餐2名称（多语言）
   // 类似 GetLocaleBuffetPackage1Name() 的实现
   func (model *SaleBill) GetLocaleBuffetPackage2Name() dto.LocaleResponse {
       // 实现逻辑类似 GetLocaleBuffetPackage1Name()
   }
   
   // SetBuffetPackage1NameSnapshot 设置自助餐套餐1名称快照（JSON）
   // 从 MultiLanguageName 获取完整多语言数据并序列化为 JSON
   // Requirement: story-main-buffet-package-name-snapshot-fix (JSON 方案)
   func (model *SaleBill) SetBuffetPackage1NameSnapshot(multiLangName MultiLanguageName) error {
       // 如果多语言名称为空，设置为空字符串
       if multiLangName.IsNullName() {
           model.BuffetPackage1Name = ""
           return nil
       }
       
       // 构建 LocaleResponse
       localeResp := multiLangName.GetNames()
       
       // 序列化为 JSON
       jsonData, err := json.Marshal(localeResp)
       if err != nil {
           return err
       }
       
       model.BuffetPackage1Name = string(jsonData)
       return nil
   }
   
   // SetBuffetPackage2NameSnapshot 设置自助餐套餐2名称快照（JSON）
   // 类似 SetBuffetPackage1NameSnapshot() 的实现
   func (model *SaleBill) SetBuffetPackage2NameSnapshot(multiLangName MultiLanguageName) error {
       // 实现逻辑类似 SetBuffetPackage1NameSnapshot()
   }
   ```

4. **修改 GetBuffetName() 方法**：
   ```go
   // main/app/model/sale_bill_ext_getset.go
   
   // 获取自助餐名称
   func (model *SaleBill) GetBuffetName() (name dto.LocaleResponse) {
       name1 := model.GetLocaleBuffetPackage1Name()
       name2 := model.GetLocaleBuffetPackage2Name()
       
       if !name1.IsNull() && !name2.IsNull() {
           // 两个套餐都存在
           name = dto.LocaleResponse{
               ZH:   fmt.Sprintf("%s+%s", name1.ZH, name2.ZH),
               TH:   fmt.Sprintf("%s+%s", name1.TH, name2.TH),
               EN:   fmt.Sprintf("%s+%s", name1.EN, name2.EN),
               ZHTW: fmt.Sprintf("%s+%s", name1.ZHTW, name2.ZHTW),
               JA:   fmt.Sprintf("%s+%s", name1.JA, name2.JA),
               KO:   fmt.Sprintf("%s+%s", name1.KO, name2.KO),
               MY:   fmt.Sprintf("%s+%s", name1.MY, name2.MY),
               TR:   fmt.Sprintf("%s+%s", name1.TR, name2.TR),
               SV:   fmt.Sprintf("%s+%s", name1.SV, name2.SV),
           }
           return
       }
       
       // 只有一个套餐时都是只填在BuffetPackage1
       if !name1.IsNull() {
           name = name1
           return
       }
       
       return name
   }
   ```

5. **修改 GetBuffetNames() 方法**：
   ```go
   // main/app/model/sale_bill_ext_getset.go
   
   // 获取所有自助餐名称
   func (model *SaleBill) GetBuffetNames(language string) string {
       buffets := make([]string, 0)
       
       // 优先使用快照字段
       name1 := model.GetLocaleBuffetPackage1Name()
       name2 := model.GetLocaleBuffetPackage2Name()
       
       if !name1.IsNull() {
           buffets = append(buffets, name1.GetLocale(language))
       }
       if !name2.IsNull() {
           buffets = append(buffets, name2.GetLocale(language))
       }
       
       // 如果快照字段都为空，降级使用关联表（兼容历史数据）
       if len(buffets) == 0 {
           for _, order := range model.SaleOrders {
               for _, buffet := range order.SaleOrderBuffetCustomerTypes {
                   name := buffet.BuffetPackage.MultiLanguageName.GetNameByLang(language)
                   if !slices.Contains(buffets, name) {
                       buffets = append(buffets, name)
                   }
               }
           }
       }
       
       return strings.Join(buffets, "+")
   }
   ```

6. **数据迁移脚本**（可选）：
   ```sql
   -- 补充历史订单的自助餐套餐1名称快照（仅迁移关联表数据存在的记录）
   UPDATE ttpos_sale_bill sb
   INNER JOIN ttpos_buffet_package bp ON sb.buffet_package1_uuid = bp.uuid
   INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
   SET sb.buffet_package1_name = JSON_OBJECT(
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
   WHERE sb.buffet_package1_name = '' 
     AND sb.buffet_package1_uuid != 0 
     AND mln.zh_name != ''
     AND sb.deleted_at IS NULL;
   
   -- 补充历史订单的自助餐套餐2名称快照（类似逻辑）
   UPDATE ttpos_sale_bill sb
   INNER JOIN ttpos_buffet_package bp ON sb.buffet_package2_uuid = bp.uuid
   INNER JOIN ttpos_multi_language_name mln ON bp.multi_language_name_uuid = mln.uuid
   SET sb.buffet_package2_name = JSON_OBJECT(
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
   WHERE sb.buffet_package2_name = '' 
     AND sb.buffet_package2_uuid != 0 
     AND mln.zh_name != ''
     AND sb.deleted_at IS NULL;
   ```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

