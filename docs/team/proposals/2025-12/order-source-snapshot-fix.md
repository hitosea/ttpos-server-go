# 外卖来源信息快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-02 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-order-source-snapshot-fix](../../shared/specs/active/story-main-order-source-snapshot-fix/) |
| **父提案** | `order-attribute-snapshot-fix.md` |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，外卖来源信息会随后台数据变更而改变，导致订单历史信息不准确。这是订单商品信息快照修复需求（`order-attribute-snapshot-fix.md`）的子任务。

**具体场景**：

1. **外卖来源信息被删除**：
   - 订单来源："美团外卖"
   - 后台删除了"美团外卖"配置
   - 查询订单时外卖来源信息丢失或显示错误

2. **外卖来源信息被改名**：
   - 订单来源："美团外卖"
   - 后台将"美团外卖"改名为"美团"
   - 查询订单时显示："美团"（显示的是新名称，而非下单时的名称）

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、统计报表等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 外卖来源信息被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单外卖来源信息准确反映下单时的状态
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **用户体验**：用户查看历史订单时看到的是下单时的真实外卖来源信息
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

**核心思路**：订单外卖来源信息应该使用快照数据，而不是从关联表实时获取。

**现状分析**：

1. **数据库设计缺少快照字段**：
   - `ttpos_sale_bill` 表只有 `order_source_uuid` 字段，关联到 `ttpos_order_source` 表
   - 没有 `order_source_name` 快照字段
   - 查询时从 `OrderSource.MultiLanguageName` 获取名称
   - 导致外卖来源数据被删除或改名时，订单显示信息变化

2. **代码实现依赖关联表**：
   - 订单详情查询中的 `OrderSourceName` 直接从 `OrderSource.MultiLanguageName` 获取
   - 没有快照机制

**解决方案**：

1. **新增快照字段**（需要数据库结构变更）：
   - 在 `ttpos_sale_bill` 表添加 `order_source_name` 字段（VARCHAR(255)），用于快照外卖来源名称（单语言）

2. **修复查询逻辑**：
   - 优先使用 `SaleBill.OrderSourceName` 字段（快照）
   - 如果快照字段为空，降级使用 `OrderSource.MultiLanguageName`（兼容历史数据）
   - 实现多语言支持：主语言使用快照，其他语言从关联表补充

3. **实施策略**：
   - **新订单**：下单时保存外卖来源名称快照到 `order_source_name` 字段
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **数据迁移**：可选，从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

4. **多语言支持**：
   - **现状**：快照字段只保存单语言（中文），但接口需要返回多语言格式
   - **解决方案**：采用"主语言快照 + 关联表补充"的混合方案
     - 快照字段保存主语言（中文）作为历史快照
     - 查询时构建 `LocaleResponse`：
       - `ZH`（中文）：优先使用快照字段，如果为空则使用关联表
       - 其他语言（TH、EN等）：优先使用关联表数据，如果关联表不存在或已删除，则使用快照的主语言填充
     - 这样既保证了快照完整性（即使数据被删除也能显示），又尽可能提供了多语言支持

### 核心功能点

#### 一、数据库结构变更

1. **新增快照字段**
   - 在 `ttpos_sale_bill` 表添加 `order_source_name` 字段
   - 类型：VARCHAR(255)
   - 注释：外卖来源名称快照（单语言），不随后台更新
   - 默认值：空字符串

#### 二、代码修改

2. **修改数据模型**
   - 在 `SaleBill` 结构体添加 `OrderSourceName` 字段
   - 字段类型：`string`
   - JSON 标签：`json:"order_source_name"`
   - GORM 标签：`gorm:"column:order_source_name"`

3. **修复外卖来源获取逻辑**
   - 新增 `SaleBill.GetLocaleOrderSourceName()` 方法
   - 实现逻辑：
     - 优先使用 `SaleBill.OrderSourceName` 字段（快照）
     - 如果快照字段为空，降级使用 `OrderSource.MultiLanguageName`
     - 构建多语言响应 `dto.LocaleResponse`

4. **修改下单逻辑**
   - 在创建订单时，保存外卖来源名称快照到 `SaleBill.OrderSourceName`
   - 从 `OrderSource.MultiLanguageName.ZhName` 获取中文名称
   - 确保所有下单入口都保存快照数据

5. **修改订单查询逻辑**
   - 在订单详情查询中，使用 `GetLocaleOrderSourceName()` 方法获取外卖来源名称
   - 替换原有的直接从关联表获取的逻辑

#### 三、数据迁移和兼容性

6. **数据完整性检查**
   - 检查历史订单的 `order_source_name` 字段填充情况
   - 统计需要迁移的订单数量

7. **兼容性处理**
   - 当快照字段为空时，降级使用关联表数据
   - 确保历史订单正常显示

8. **下单时保存快照**
   - 确保创建订单时正确保存外卖来源名称快照
   - 梳理所有下单入口（POS、扫码点餐、外卖等）

#### 四、测试验证

9. **测试验证**
   - 验证外卖来源删除后订单显示
   - 验证外卖来源改名后订单显示
   - 验证多语言显示（关联表存在/不存在的情况）
   - 验证历史订单兼容性
   - 验证所有下单入口都正确保存快照

### 影响范围

**涉及终端**：
- [x] POS 收银端（订单查询）
- [x] Shop 商家管理端（订单管理、对账）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [x] Mobile 扫码端（订单历史）
- [ ] Menu 电子菜单端
- [x] Member 会员端（历史订单）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（订单查询接口）
- [x] 数据模型（`SaleBill`）
- [x] 业务逻辑（外卖来源名称获取方法、下单逻辑）
- [ ] 第三方集成
- [x] 数据库迁移（新增快照字段）
- [x] 下单逻辑（保存快照数据）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要修改业务逻辑，涉及数据一致性
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 需要数据库结构变更（新增快照字段）
- 需要修改订单查询逻辑和下单逻辑
- 需要处理兼容性和多语言支持
- 需要充分测试确保不影响现有功能

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**任务分解**：
1. **数据库结构变更**（0.5 天）
   - 设计数据库迁移脚本
   - 添加 `order_source_name` 字段到 `ttpos_sale_bill` 表
   - 执行迁移并验证

2. **代码修改**（1-1.5 天）
   - 修改 `SaleBill` 模型，添加 `OrderSourceName` 字段
   - 新增 `GetLocaleOrderSourceName()` 方法
   - 修改订单查询逻辑，使用快照数据
   - 修改下单逻辑，保存快照数据
   - 添加兼容性处理和多语言支持

3. **数据检查与迁移**（0.5 天）
   - 检查历史数据完整性
   - 编写数据迁移脚本（可选）
   - 执行数据迁移（如需要）

4. **测试验证**（0.5 天）
   - 单元测试（覆盖新增的方法）
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

3. **多语言支持问题**
   - 问题：快照字段只保存单语言（中文），但接口返回需要 `dto.LocaleResponse` 格式（多语言）
   - 影响：需要设计合理的多语言快照方案
   - 解决方案：采用"主语言快照 + 关联表补充"的混合方案
     - 快照字段保存主语言（中文）
     - 查询时，优先使用快照字段填充主语言（ZH）
     - 如果关联表数据存在且未删除，使用关联表数据填充其他语言（TH、EN等）
     - 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充

4. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存快照字段
   - 影响：可能遗漏某些下单场景，导致快照数据不完整
   - 缓解：全面梳理所有下单入口，确保都保存快照数据

5. **性能影响**
   - 风险：如果快照字段为空，降级查询可能增加数据库查询
   - 影响：需要优化查询逻辑，优先使用快照数据
   - 缓解：优先使用快照数据，减少关联查询

6. **回归风险**
   - 风险：修改订单查询逻辑可能影响其他功能（订单打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能
   - 缓解：编写单元测试和集成测试，进行回归测试

**缓解措施**：

1. **数据检查**：
   - 先检查历史数据的 `order_source_name` 字段填充情况
   - 根据检查结果决定是否需要数据迁移

2. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

3. **多语言处理**：
   - 采用"主语言快照 + 关联表补充"方案
   - 快照字段保存主语言（中文）
   - 查询时优先使用快照字段填充主语言
   - 关联表存在时使用关联表数据填充其他语言
   - 关联表不存在时所有语言都用快照的主语言填充

4. **性能优化**：
   - 优先使用快照数据，减少关联查询
   - 如果必须降级查询，使用索引优化

5. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出等）
   - 在测试环境充分验证后再上线

6. **全面梳理**：
   - 梳理所有使用外卖来源信息的地方
   - 确保所有相关方法都使用快照数据
   - 梳理所有下单入口，确保都保存快照数据

---

## 🔗 相关资源

### 参考需求

- 父提案: `docs/team/proposals/2025-01/order-attribute-snapshot-fix.md`
- 类似功能: 订单商品快照机制（商品名称、价格等已有快照）
- 参考子任务: `docs/team/proposals/2025-12/nationality-snapshot-fix.md`（国籍快照修复）
- 竞品分析: 主流餐饮系统都采用订单快照机制

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/sale_bill.go`
- 外卖来源模型: `main/app/model/order_source.go`

### 代码位置

**问题代码**：
- 订单详情查询中的 `OrderSourceName` 获取逻辑（从 `OrderSource.MultiLanguageName` 获取）

**数据模型**：
- `main/app/model/sale_bill.go:15-112` - `SaleBill` 模型定义（需要新增 `OrderSourceName` 字段）
- `main/app/model/order_source.go` - `OrderSource` 模型定义

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

- [x] 创建 Spec：`story-main-order-source-snapshot-fix` ✅ 2025-12-02
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}
- [ ] 产品审核：需求文档（requirements.md）
- [ ] 技术设计：使用 `/spec-design` 创建设计文档和任务分解

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实外卖来源信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的外卖来源信息  
**以便于** 准确处理客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含外卖来源信息的订单 **THEN** 系统 **SHALL** 显示下单时保存的外卖来源名称快照
2. **IF** 后台删除了某个外卖来源配置 **THEN** 历史订单 **SHALL** 仍然显示该外卖来源的原始名称
3. **IF** 后台修改了某个外卖来源的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
4. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
5. **WHEN** 创建新订单 **THEN** 系统 **SHALL** 正确保存外卖来源名称快照到 `SaleBill.OrderSourceName` 字段
6. **WHEN** 查询订单外卖来源信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
7. **IF** 关联表数据存在 **THEN** 系统 **SHALL** 使用关联表数据填充其他语言（TH、EN等）
8. **IF** 关联表数据不存在（已删除） **THEN** 系统 **SHALL** 使用快照的主语言填充所有语言字段

### 技术方案要点（初稿）

#### 多语言快照策略

**核心思路**：主语言快照 + 关联表补充

- **快照字段**：保存主语言（中文）作为历史快照
- **查询逻辑**：
  - 主语言（ZH）：优先使用快照字段
  - 其他语言：优先使用关联表，如果关联表不存在或已删除，则使用快照的主语言填充

#### 具体实现方案

1. **数据库迁移脚本**：
   ```sql
   -- 添加外卖来源名称快照字段
   ALTER TABLE `ttpos_sale_bill` 
   ADD COLUMN `order_source_name` VARCHAR(255) NOT NULL DEFAULT '' 
   COMMENT '外卖来源名称快照（单语言），不随后台更新' 
   AFTER `order_source_uuid`;
   ```

2. **修改数据模型**：
   ```go
   // main/app/model/sale_bill.go
   type SaleBill struct {
       // ... 其他字段
       OrderSourceUuid uint64 `gorm:"column:order_source_uuid" json:"order_source_uuid"`
       OrderSourceName string `gorm:"column:order_source_name" json:"order_source_name"` // 新增快照字段
       // ... 其他字段
   }
   ```

3. **新增外卖来源获取方法**：
   ```go
   // main/app/model/sale_bill.go
   // GetLocaleOrderSourceName 获取外卖来源名称（多语言）
   func (model *SaleBill) GetLocaleOrderSourceName() dto.LocaleResponse {
       // 优先使用快照字段
       snapshotName := model.OrderSourceName
       
       // 如果快照字段为空，降级使用关联表（兼容历史数据）
       if snapshotName == "" && model.OrderSource != nil && model.OrderSource.MultiLanguageName != nil {
           return model.OrderSource.MultiLanguageName.GetNames()
       }
       
       // 如果快照字段有值，构建多语言响应
       result := dto.LocaleResponse{ZH: snapshotName}
       
       // 如果关联表数据存在且未删除，使用关联表数据填充其他语言
       if model.OrderSource != nil && model.OrderSource.MultiLanguageName != nil && !model.OrderSource.MultiLanguageName.IsNullName() {
           multiLang := model.OrderSource.MultiLanguageName.GetNames()
           result.TH = multiLang.TH
           result.EN = multiLang.EN
           result.ZHTW = multiLang.ZHTW
           result.JA = multiLang.JA
           result.KO = multiLang.KO
           result.MY = multiLang.MY
           result.TR = multiLang.TR
           result.SV = multiLang.SV
       } else {
           // 如果关联表数据不存在（已删除），所有语言都用快照的主语言填充
           result.TH = snapshotName
           result.EN = snapshotName
           result.ZHTW = snapshotName
           result.JA = snapshotName
           result.KO = snapshotName
           result.MY = snapshotName
           result.TR = snapshotName
           result.SV = snapshotName
       }
       
       return result
   }
   ```

4. **修改下单逻辑**：
   - 在创建 `SaleBill` 时，从 `OrderSource.MultiLanguageName.ZhName` 获取中文名称
   - 保存到 `SaleBill.OrderSourceName` 字段
   - 梳理所有下单入口（POS、扫码点餐、外卖等）

5. **修改订单查询逻辑**：
   - 在订单详情查询接口中，使用 `SaleBill.GetLocaleOrderSourceName()` 获取外卖来源名称
   - 替换原有的直接从 `OrderSource.MultiLanguageName` 获取的逻辑

6. **数据检查脚本**：
   - 检查 `ttpos_sale_bill.order_source_name` 字段的填充率
   - 统计需要迁移的订单数量

7. **数据迁移脚本**（可选）：
   - **策略**：只对关联表数据存在的历史订单做补充，其他通过降级逻辑兼容
   - **迁移SQL**：
     ```sql
     -- 补充历史订单的外卖来源名称快照（仅迁移关联表数据存在的记录）
     UPDATE ttpos_sale_bill sb
     INNER JOIN ttpos_order_source os ON sb.order_source_uuid = os.uuid
     INNER JOIN ttpos_multi_language_name mln ON os.multi_language_name_uuid = mln.uuid
     SET sb.order_source_name = mln.zh_name
     WHERE sb.order_source_name = '' 
       AND sb.order_source_uuid != 0 
       AND mln.zh_name != ''
       AND sb.deleted_at IS NULL;
     ```

### 线框图/原型（可选）

无需 UI 变更，主要是后端逻辑修复。

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-02  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

