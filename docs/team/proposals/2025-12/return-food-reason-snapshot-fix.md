# 原因信息快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。  
> 本提案是「订单商品信息快照修复」的子任务，聚焦于免单原因和退菜原因快照修复。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-08 |
| **目标版本** | v2.11.0 |
| **状态**   | 已批准 → Spec 已创建   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-reason-snapshot-fix](../../shared/specs/active/story-main-reason-snapshot-fix/requirements.md) |
| **父提案** | [订单商品信息快照修复](./2025-01/order-attribute-snapshot-fix.md) |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，免单原因和退菜原因信息会随后台数据变更而改变，导致订单历史信息不准确。

**具体场景**：

1. **免单原因被删除**：
   - 订单免单原因："员工福利"
   - 后台删除了"员工福利"原因
   - 查询订单时显示错误信息或空值

2. **免单原因被改名**：
   - 订单免单原因："员工福利"
   - 后台将"员工福利"改名为"内部测试"
   - 查询订单时显示："内部测试"（显示的是新名称，而非下单时的名称）

3. **退菜原因被删除**：
   - 订单退菜原因："菜品质量问题"
   - 后台删除了"菜品质量问题"原因
   - 查询订单时显示错误信息或空值

4. **退菜原因被改名**：
   - 订单退菜原因："菜品质量问题"
   - 后台将"菜品质量问题"改名为"质量问题"
   - 查询订单时显示："质量问题"（显示的是新名称，而非下单时的名称）

**问题影响**：

- ❌ 订单历史信息不准确，无法还原免单/退菜时的真实原因
- ❌ 影响对账、退款等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 可能导致法律风险（如发票、账单与实际订单不符）
- ❌ 免单/退菜原因被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单免单/退菜原因准确反映免单/退菜时的状态
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **用户体验**：用户查看历史订单时看到的是免单/退菜时的真实原因
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

**核心思路**：订单免单原因和退菜原因应该使用快照数据，而不是从关联表实时获取。

**现状分析**：

1. **数据库设计未支持快照**：
   - `ttpos_sale_order_product_reason` 表只有 `multi_language_name_uuid` 字段，通过 UUID 关联到 `ttpos_multi_language_name` 表
   - 没有 `name` 快照字段，导致免单/退菜原因被删除或改名时，订单显示信息变化

2. **代码实现未使用快照**：
   - **免单原因**：从 `FreeReason.MultiLanguageName` 获取（通过 UUID 关联，可能被修改或删除）
   - `SaleOrder.GetFreeReason()` 方法直接使用 `reason.MultiLanguageName`，未使用快照字段
   - **退菜原因**：从 `ReturnFoodReason.MultiLanguageName` 获取（通过 UUID 关联，可能被修改或删除）
   - `SaleOrderProduct.GetCancelReason()` 方法直接使用 `reason.MultiLanguageName`，未使用快照字段
   - 导致免单/退菜原因被删除或改名时，订单显示信息变化

**解决方案**：

1. **数据库结构变更**：
   - 在 `ttpos_sale_order_product_reason` 表添加 `name` 字段（TEXT 类型），用于快照免单/退菜原因名称（JSON 格式，包含所有语言）

2. **修复查询逻辑**：
   - **免单原因**：修改 `SaleOrder.GetFreeReason()` 方法
     - 优先使用 `SaleOrderProductReason.Name` 字段（JSON 格式）
     - 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
     - 如果 `Name` 为空或 JSON 解析失败，降级使用 `FreeReason.MultiLanguageName`（兼容历史数据）
   - **退菜原因**：修改 `SaleOrderProduct.GetCancelReason()` 方法
     - 优先使用 `SaleOrderProductReason.Name` 字段（JSON 格式）
     - 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
     - 如果 `Name` 为空或 JSON 解析失败，降级使用 `ReturnFoodReason.MultiLanguageName`（兼容历史数据）

3. **修复下单逻辑**：
   - **免单原因**：修改 `SaleOrder.NewFreeOrderReason()` 方法
     - 确保创建免单原因时保存 `SaleOrderProductReason.Name` 快照字段
     - 从 `FreeReason.MultiLanguageName` 获取完整多语言数据，序列化为 JSON 保存到快照字段
   - **退菜原因**：修改 `SaleOrderProduct.NewSaleOrderProductReasonList()` 方法
     - 确保创建退菜原因时保存 `SaleOrderProductReason.Name` 快照字段
     - 从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据，序列化为 JSON 保存到快照字段

4. **多语言支持（JSON 方案）**：
   - **快照字段**：保存完整多语言 JSON 字符串（包含 ZH、TH、EN、ZHTW、JA、KO、MY、TR、SV 所有语言）
   - **查询逻辑**：
     - 优先使用快照字段（JSON），解析后返回完整多语言数据
     - 如果快照字段为空或 JSON 解析失败，降级使用关联表数据（兼容历史数据）
     - 这样既保证了快照完整性（即使数据被删除也能显示），又提供了完整的多语言支持

5. **实施策略**：
   - **新订单**：免单/退菜时保存快照字段
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **数据迁移**：可选，从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

### 核心功能点

1. **数据库结构变更**
   - 在 `ttpos_sale_order_product_reason` 表添加 `name` 字段（TEXT 类型），用于快照免单/退菜原因名称（JSON 格式，包含所有语言）

2. **修复免单原因获取逻辑（JSON 方案）**
   - 修改 `SaleOrder.GetFreeReason()` 方法
   - 优先使用 `SaleOrderProductReason.Name` 字段（JSON 格式）
   - 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
   - 如果 `Name` 为空或 JSON 解析失败，降级使用 `FreeReason.MultiLanguageName`

3. **修复退菜原因获取逻辑（JSON 方案）**
   - 修改 `SaleOrderProduct.GetCancelReason()` 方法
   - 优先使用 `SaleOrderProductReason.Name` 字段（JSON 格式）
   - 解析 JSON 为 `dto.LocaleResponse`（包含所有语言）
   - 如果 `Name` 为空或 JSON 解析失败，降级使用 `ReturnFoodReason.MultiLanguageName`

4. **修复下单逻辑（JSON 方案）**
   - **免单原因**：修改 `SaleOrder.NewFreeOrderReason()` 方法
     - 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
     - 序列化为 JSON 保存到 `SaleOrderProductReason.Name` 快照字段
   - **退菜原因**：修改 `SaleOrderProduct.NewSaleOrderProductReasonList()` 方法
     - 从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据
     - 序列化为 JSON 保存到 `SaleOrderProductReason.Name` 快照字段

5. **数据迁移和兼容性**
   - 检查历史订单的快照字段填充情况
   - 提供数据迁移脚本（可选）
   - 确保历史订单正常显示（降级逻辑）

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
- [x] 数据模型（`SaleOrderProductReason`）
- [x] 业务逻辑（免单原因和退菜原因获取方法）
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
- 需要修改免单原因和退菜原因获取方法
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
   - 添加快照字段到 `ttpos_sale_order_product_reason` 表
   - 执行迁移并验证

2. **代码修改 - 免单原因**（0.5-1 天）
   - 修改 `GetFreeReason()` 方法
   - 修改 `NewFreeOrderReason()` 方法
   - 添加兼容性处理

3. **代码修改 - 退菜原因**（0.5-1 天）
   - 修改 `GetCancelReason()` 方法
   - 修改 `NewSaleOrderProductReasonList()` 方法
   - 添加兼容性处理

4. **数据检查与迁移**（0.5 天）
   - 检查历史数据完整性
   - 编写数据迁移脚本（可选）
   - 执行数据迁移

5. **测试验证**（0.5-1 天）
   - 单元测试（覆盖修改的方法）
   - 集成测试（验证订单查询、免单/退菜保存快照）
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

3. **多语言支持问题（JSON 方案）**
   - **问题**：快照字段需要保存多语言数据，接口返回需要 `dto.LocaleResponse` 格式（多语言）
   - **影响**：需要设计合理的多语言快照方案
   - **解决方案**：采用"JSON 快照"方案
     - 快照字段保存完整多语言 JSON 字符串（包含 ZH、TH、EN、ZHTW、JA、KO、MY、TR、SV 所有语言）
     - 查询时，优先使用快照字段（JSON），解析后返回完整多语言数据
     - 如果快照字段为空或 JSON 解析失败，降级使用关联表数据（兼容历史数据）
     - 这样既保证了快照完整性（即使数据被删除也能显示），又提供了完整的多语言支持

4. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存快照字段
   - 影响：可能遗漏某些免单/退菜场景，导致快照数据不完整
   - 缓解：全面梳理所有免单/退菜入口，确保都保存快照数据

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

3. **多语言处理（JSON 方案）**：
   - 采用"JSON 快照"方案
   - 快照字段保存完整多语言 JSON 字符串（包含所有语言）
   - 查询时优先使用快照字段（JSON），解析后返回完整多语言数据
   - 如果快照字段为空或 JSON 解析失败，降级使用关联表数据（兼容历史数据）

4. **性能优化**：
   - 优先使用快照数据，减少关联查询
   - 如果必须降级查询，使用索引优化

5. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试多语言快照逻辑（关联表存在/不存在的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出、退款等）
   - 在测试环境充分验证后再上线

6. **全面梳理**：
   - 梳理所有使用免单原因和退菜原因的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 父提案: [订单商品信息快照修复](./2025-01/order-attribute-snapshot-fix.md)
- 类似功能: 订单商品快照机制（商品名称、价格等已有快照）
- 竞品分析: 主流餐饮系统都采用订单快照机制

### 相关文档

- 订单信息获取逻辑分析: `docs/shared/api/cashier-order-info-analysis.md`
- 数据模型定义: `main/app/model/order.go`
- 免单原因获取方法: `main/app/model/sale_order_ext_getset.go`
- 退菜原因获取方法: `main/app/model/sale_order_product.go`

### 代码位置

**问题代码**：
- `main/app/model/sale_order_ext_getset.go:29` - `GetFreeReason()` 方法（免单原因）
- `main/app/model/sale_order.go:1173` - `NewFreeOrderReason()` 方法（创建免单原因）
- `main/app/model/sale_order_product.go:959` - `GetCancelReason()` 方法（退菜原因）
- `main/app/model/sale_order_product.go:926` - `NewSaleOrderProductReasonList()` 方法（创建退菜原因）

**数据模型**：
- `main/app/model/order.go:350-365` - `SaleOrderProductReason` 模型定义（需要新增 `Name` 字段）
- `main/app/model/reason.go:14-51` - `FreeReason` 模型定义
- `main/app/model/reason.go:5-51` - `ReturnFoodReason` 模型定义
- `main/app/model/sale_order.go:102` - `SaleOrder.FreeReasons` 字段定义
- `main/app/model/sale_order_product.go:125` - `SaleOrderProduct.CancelReasons` 字段定义

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

- [ ] 创建 Spec：`story-main-reason-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到免单/退菜时的真实原因信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到免单/退菜时的真实原因  
**以便于** 准确处理退款和客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含免单原因的订单 **THEN** 系统 **SHALL** 显示免单时保存的免单原因快照
2. **WHEN** 查询包含退菜原因的订单 **THEN** 系统 **SHALL** 显示退菜时保存的退菜原因快照
3. **IF** 后台删除了某个免单原因 **THEN** 历史订单 **SHALL** 仍然显示该原因的原始名称
4. **IF** 后台删除了某个退菜原因 **THEN** 历史订单 **SHALL** 仍然显示该原因的原始名称
5. **IF** 后台修改了某个免单原因的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
6. **IF** 后台修改了某个退菜原因的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
7. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
8. **WHEN** 创建新免单原因 **THEN** 系统 **SHALL** 正确保存快照字段（`SaleOrderProductReason.Name`）
9. **WHEN** 创建新退菜原因 **THEN** 系统 **SHALL** 正确保存快照字段（`SaleOrderProductReason.Name`）
10. **WHEN** 查询订单免单原因信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
11. **WHEN** 查询订单退菜原因信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
12. **WHEN** 查询订单免单/退菜原因信息 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
13. **IF** 快照字段有值（JSON）**THEN** 系统 **SHALL** 解析 JSON 并返回完整多语言数据（所有语言）
14. **IF** 快照字段为空或 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用关联表数据（兼容历史数据）

### 技术方案要点（初稿）

#### 多语言快照策略（JSON 方案）

**核心思路**：JSON 快照 + 关联表降级

- **快照字段**：保存完整多语言 JSON 字符串（包含 ZH、TH、EN、ZHTW、JA、KO、MY、TR、SV 所有语言）
- **查询逻辑**：
  - 优先使用快照字段（JSON），解析后返回完整多语言数据
  - 如果快照字段为空或 JSON 解析失败，降级使用关联表数据（兼容历史数据）

#### 具体实现方案

1. **修改免单原因获取方法（JSON 方案）**：
   ```go
   func (model *SaleOrder) GetFreeReason() dto.LocaleResponse {
       zhNames := make([]string, 0)
       thNames := make([]string, 0)
       enNames := make([]string, 0)
       // ... 其他语言
       
       // 遍历选择的免单原因
       for _, reason := range model.FreeReasons {
           if !reason.IsFreeReason() || reason.IsDelete() {
               continue
           }
           
           // 优先使用快照字段（JSON）
           snapshotJSON := reason.Name
           var snapshotLocale dto.LocaleResponse
           
           // 如果快照字段不为空，尝试反序列化为多语言数据
           if snapshotJSON != "" {
               if err := json.Unmarshal([]byte(snapshotJSON), &snapshotLocale); err == nil {
                   // 反序列化成功，检查是否有主语言数据
                   if !snapshotLocale.IsNull() {
                       // 使用快照数据（所有语言）
                       zhNames = append(zhNames, snapshotLocale.ZH)
                       thNames = append(thNames, snapshotLocale.TH)
                       enNames = append(enNames, snapshotLocale.EN)
                       // ... 其他语言
                       continue
                   }
               }
               // 如果反序列化失败或数据不完整，继续后续降级逻辑
           }
           
           // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
           if reason.MultiLanguageName != nil && !reason.MultiLanguageName.IsNullName() {
               multiLang := reason.MultiLanguageName.GetNames()
               zhNames = append(zhNames, multiLang.ZH)
               thNames = append(thNames, multiLang.TH)
               enNames = append(enNames, multiLang.EN)
               // ... 其他语言
           }
       }
       
       // 添加自定义的免单原因
       if model.FreeReason != "" {
           zhNames = append(zhNames, model.FreeReason)
           thNames = append(thNames, model.FreeReason)
           enNames = append(enNames, model.FreeReason)
           // ... 所有语言都用自定义原因
       }
       
       return dto.LocaleResponse{
           ZH:   strings.Join(zhNames, "、"),
           TH:   strings.Join(thNames, "、"),
           EN:   strings.Join(enNames, "、"),
           // ... 其他语言
       }
   }
   ```

2. **修改创建免单原因方法（JSON 方案）**：
   ```go
   func (model *SaleOrder) NewFreeOrderReason(freeReasons []*FreeReason) []*SaleOrderProductReason {
       list := make([]*SaleOrderProductReason, 0)
       for _, reason := range freeReasons {
           reasonUuid, _ := utils.GetID()
           
           // 序列化多语言数据为 JSON
           var nameJSON string
           if reason.MultiLanguageName != nil && !reason.MultiLanguageName.IsNullName() {
               localeResp := reason.MultiLanguageName.GetNames()
               jsonData, err := json.Marshal(localeResp)
               if err == nil {
                   nameJSON = string(jsonData)
               }
           }
           
           reasonRecord := &SaleOrderProductReason{
               BaseModel: BaseModel{
                   Uuid: reasonUuid,
               },
               SaleOrderUuid:         model.Uuid,
               MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
               FreeReasonUuid:        reason.Uuid,
               // 保存快照字段（JSON 格式，包含所有语言）
               Name: nameJSON,
           }
           list = append(list, reasonRecord)
       }
       return list
   }
   ```

3. **修改退菜原因获取方法（JSON 方案）**：
   ```go
   func (model *SaleOrderProduct) GetCancelReason() dto.LocaleResponse {
       zhNames := make([]string, 0)
       thNames := make([]string, 0)
       enNames := make([]string, 0)
       // ... 其他语言
       
       // 遍历选择的退菜原因
       for _, reason := range model.CancelReasons {
           if !reason.IsReturnFoodReason() || reason.IsDelete() {
               continue
           }
           
           // 优先使用快照字段（JSON）
           snapshotJSON := reason.Name
           var snapshotLocale dto.LocaleResponse
           
           // 如果快照字段不为空，尝试反序列化为多语言数据
           if snapshotJSON != "" {
               if err := json.Unmarshal([]byte(snapshotJSON), &snapshotLocale); err == nil {
                   // 反序列化成功，检查是否有主语言数据
                   if !snapshotLocale.IsNull() {
                       // 使用快照数据（所有语言）
                       zhNames = append(zhNames, snapshotLocale.ZH)
                       thNames = append(thNames, snapshotLocale.TH)
                       enNames = append(enNames, snapshotLocale.EN)
                       // ... 其他语言
                       continue
                   }
               }
               // 如果反序列化失败或数据不完整，继续后续降级逻辑
           }
           
           // 降级：如果快照字段为空或反序列化失败，使用关联表（兼容历史数据）
           if reason.MultiLanguageName != nil && !reason.MultiLanguageName.IsNullName() {
               multiLang := reason.MultiLanguageName.GetNames()
               zhNames = append(zhNames, multiLang.ZH)
               thNames = append(thNames, multiLang.TH)
               enNames = append(enNames, multiLang.EN)
               // ... 其他语言
           }
       }
       
       // 添加自定义的退菜原因
       if model.CancelReason != "" {
           zhNames = append(zhNames, model.CancelReason)
           thNames = append(thNames, model.CancelReason)
           enNames = append(enNames, model.CancelReason)
           // ... 所有语言都用自定义原因
       }
       
       return dto.LocaleResponse{
           ZH:   strings.Join(zhNames, "、"),
           TH:   strings.Join(thNames, "、"),
           EN:   strings.Join(enNames, "、"),
           // ... 其他语言
       }
   }
   ```

4. **修改创建退菜原因方法（JSON 方案）**：
   ```go
   func (model *SaleOrderProduct) NewSaleOrderProductReasonList(reasons []*ReturnFoodReason) []*SaleOrderProductReason {
       list := make([]*SaleOrderProductReason, 0)
       for _, reason := range reasons {
           reasonUuid, _ := utils.GetID()
           
           // 序列化多语言数据为 JSON
           var nameJSON string
           if reason.MultiLanguageName != nil && !reason.MultiLanguageName.IsNullName() {
               localeResp := reason.MultiLanguageName.GetNames()
               jsonData, err := json.Marshal(localeResp)
               if err == nil {
                   nameJSON = string(jsonData)
               }
           }
           
           reasonRecord := &SaleOrderProductReason{
               BaseModel: BaseModel{
                   Uuid: reasonUuid,
               },
               SaleOrderUuid:         model.SaleOrderUuid,
               SaleOrderProductUuid:  model.Uuid,
               ReturnFoodReasonUuid:  reason.Uuid,
               MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
               // 保存快照字段（JSON 格式，包含所有语言）
               Name: nameJSON,
           }
           list = append(list, reasonRecord)
       }
       return list
   }
   ```

5. **数据库迁移脚本（JSON 方案）**：
   ```sql
   -- 添加免单/退菜原因快照字段（JSON 格式，包含所有语言）
   ALTER TABLE `ttpos_sale_order_product_reason` 
   ADD COLUMN `name` TEXT NOT NULL DEFAULT '' COMMENT '原因名称快照（JSON），不随后台更新' AFTER `gift_reason_uuid`;
   ```
   
   **JSON 格式示例**：
   ```json
   {"zh":"员工福利","th":"","en":"Employee Benefit","zhtw":"","ja":"","ko":"","my":"","tr":"","sv":""}
   ```

6. **数据检查脚本**：
   - 检查 `ttpos_sale_order_product_reason.name` 字段的填充率（免单原因和退菜原因）
   - 识别需要补充数据的订单

7. **数据迁移脚本**（可选）：
   - **策略**：只对之后的订单做处理，历史订单通过降级逻辑兼容
   - **可选迁移**：从关联表补充历史订单的快照字段（仅迁移关联表数据存在的记录）
   - **迁移范围**：只迁移关联表数据存在的记录，对于已删除的数据，保持快照字段为空（使用降级逻辑）
   - **迁移时机**：可以在系统空闲时执行，不影响正常业务

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`
