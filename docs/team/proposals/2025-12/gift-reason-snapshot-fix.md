# 赠菜原因快照修复 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。  
> 本提案是 [订单商品信息快照修复](../2025-01/order-attribute-snapshot-fix.md) 的子提案，专门针对赠菜原因快照功能。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiezhihuan |
| **日期**   | 2025-12-09 |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | -      |
| **关联 Spec** | [story-main-gift-reason-snapshot-fix](../../shared/specs/active/story-main-gift-reason-snapshot-fix/requirements.md) |
| **关联提案** | [订单商品信息快照修复](../2025-01/order-attribute-snapshot-fix.md) |

---

## 🎯 背景和动机

### 问题描述

当前订单查询时，赠菜原因会随后台数据变更而改变，导致订单历史信息不准确。

**具体场景**：

1. **赠菜原因被删除**：
   - 订单赠菜原因："会员生日福利"（下单时选择了该原因）
   - 后台删除了"会员生日福利"原因
   - 查询订单时显示：空或错误信息（赠菜原因信息丢失）

2. **赠菜原因被改名**：
   - 订单赠菜原因："会员生日福利"（下单时选择了该原因）
   - 后台将"会员生日福利"改名为"生日优惠"
   - 查询订单时显示："生日优惠"（显示的是新名称，而非下单时的名称）

**问题影响**：

- ❌ 订单历史信息不准确，无法还原下单时的真实状态
- ❌ 影响对账、退款等业务场景的准确性
- ❌ 违反数据一致性原则：订单信息应该作为历史快照，不应随数据变更而改变
- ❌ 赠菜原因被删除后，历史订单可能无法正常显示
- ❌ 影响订单追溯和审计功能

### 业务价值

**解决这个问题能带来什么业务价值？**

- ✅ **数据准确性**：确保订单信息准确反映下单时的状态
- ✅ **合规性**：满足财务、税务对订单历史记录的要求
- ✅ **可追溯性**：支持订单历史查询和问题追溯
- ✅ **用户体验**：用户查看历史订单时看到的是下单时的真实信息
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

**核心思路**：订单赠菜原因应该使用快照数据，而不是从关联表实时获取。

**现状分析**：

1. **数据库设计已支持快照**：
   - `ttpos_sale_order_product_reason` 表已有 `name` 字段（TEXT 类型），注释："原因名称快照（JSON），不随后台更新"
   - 该字段同时用于免单、退菜、赠菜原因的快照
   - 说明设计时已考虑快照需求

2. **代码实现未使用快照**：
   - **赠菜原因**：从 `GiftReason.MultiLanguageName` 获取，未使用 `SaleOrderProductReason.Name` 快照字段
   - `GetGiftReason()` 方法（`main/app/model/sale_order_product.go:1073`）直接从关联表获取数据
   - 导致赠菜原因被删除或改名时，订单显示信息变化

**解决方案**：

1. **修复查询逻辑**（使用现有快照字段）：
   - **赠菜原因**：优先使用 `SaleOrderProductReason.Name` 字段（JSON 快照），如果为空则使用关联表数据（兼容历史数据）

2. **修复下单逻辑**（保存快照）：
   - 在创建 `SaleOrderProductReason` 时，如果 `GiftReasonUuid` 不为空，保存赠菜原因的快照字段（JSON 格式）

3. **实施策略**：
   - **新订单**：下单时保存赠菜原因快照字段
   - **历史订单**：快照字段为空时，降级使用关联表数据（兼容性处理）
   - **渐进式实施**：不需要强制迁移所有历史数据，新订单自动使用快照机制

4. **多语言支持**：
   - **快照字段**：保存完整的多语言 JSON（包含所有语言：ZH、TH、EN、ZHTW、JA、KO、MY、TR、SV）
   - **查询逻辑**：优先使用快照字段，如果快照为空或解析失败，降级使用关联表数据
   - 这样既保证了快照完整性（即使数据被删除也能显示），又尽可能提供了多语言支持

### 核心功能点

1. **修复赠菜原因获取逻辑**
   - 修改 `SaleOrderProduct.GetGiftReason()` 方法
   - 优先使用 `SaleOrderProductReason.Name` 字段（JSON 快照）
   - 如果 `Name` 为空或解析失败，降级使用 `GiftReason.MultiLanguageName`
   - 支持多语言返回（`dto.LocaleResponse`）

2. **修复下单逻辑 - 保存快照**
   - 在创建赠菜原因时，保存快照字段
   - 确保 `SaleOrderProductReason.Name` 字段正确保存多语言 JSON

3. **兼容性处理**
   - 当快照字段为空时，降级使用关联表数据
   - 确保历史订单正常显示

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
- [x] 业务逻辑（`GetGiftReason()` 方法）
- [ ] 第三方集成
- [ ] 数据库迁移（无需新增字段，使用现有字段）

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要修改业务逻辑，涉及数据一致性
- [ ] **高**：涉及数据库结构变更、业务逻辑修改、数据迁移

**说明**：
- 无需数据库结构变更（快照字段已存在）
- 需要修改 `GetGiftReason()` 方法
- 需要处理兼容性和数据迁移
- 需要修改下单逻辑，确保保存快照数据
- 需要充分测试确保不影响现有功能

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 2-3 SP（待技术评审确认）

**任务分解**：
1. **代码修改 - 查询逻辑**（0.5-1 天）
   - 修改 `GetGiftReason()` 方法
   - 添加快照字段解析逻辑
   - 添加兼容性处理

2. **代码修改 - 下单逻辑**（0.5 天）
   - 确保创建赠菜原因时保存快照字段
   - 验证所有下单入口都保存快照

3. **测试验证**（0.5 天）
   - 单元测试（覆盖快照有值/无值、JSON 有效/无效等场景）
   - 集成测试（验证订单查询、下单保存快照）
   - 回归测试（确保不影响现有功能）

### 风险识别

**潜在风险**：

1. **历史数据不完整**
   - 风险：部分历史订单的快照字段可能为空
   - 影响：需要降级处理
   - 缓解：实现降级逻辑，历史数据可以逐步迁移

2. **下单逻辑修改风险**
   - 风险：需要修改下单逻辑，确保保存快照字段
   - 影响：可能遗漏某些下单场景，导致快照数据不完整
   - 缓解：全面梳理所有下单入口，确保都保存快照数据

3. **JSON 解析失败**
   - 风险：快照字段可能包含无效 JSON
   - 影响：需要降级处理
   - 缓解：实现 JSON 解析错误处理，降级使用关联表数据

4. **回归风险**
   - 风险：修改核心方法可能影响其他功能（订单查询、打印、导出等）
   - 影响：需要充分测试，特别是订单相关的所有功能

**缓解措施**：

1. **兼容性处理**：
   - 实现降级逻辑，确保历史订单正常显示
   - 逐步迁移，不强制要求所有数据立即完整

2. **充分测试**：
   - 编写单元测试覆盖所有修改的方法和场景
   - 测试快照逻辑（快照有值/无值、JSON 有效/无效的情况）
   - 进行回归测试确保不影响现有功能（订单查询、打印、导出等）
   - 在测试环境充分验证后再上线

3. **全面梳理**：
   - 梳理所有使用赠菜原因的地方
   - 确保所有相关方法都使用快照数据

---

## 🔗 相关资源

### 参考需求

- 主提案: [订单商品信息快照修复](../2025-01/order-attribute-snapshot-fix.md)
- 类似功能: 免单原因快照、退菜原因快照（使用相同的快照字段）

### 相关文档

- 数据模型定义: `main/app/model/order.go` - `SaleOrderProductReason`
- 赠菜原因获取方法: `main/app/model/sale_order_product.go:1073` - `GetGiftReason()`

### 代码位置

**问题代码**：
- `main/app/model/sale_order_product.go:1073` - `GetGiftReason()` 方法（赠菜原因）

**数据模型**：
- `main/app/model/order.go:466-485` - `SaleOrderProductReason` 模型定义（`Name` 字段已存在）
- `main/app/model/reason.go` - `GiftReason` 模型定义

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

- [x] 创建 Spec：`story-main-gift-reason-snapshot-fix`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看历史订单时看到下单时的真实赠菜原因信息  
**以便于** 准确对账和追溯订单历史

**作为** 收银员  
**我想** 查看订单详情时看到下单时的赠菜原因  
**以便于** 准确处理客户咨询

### AC 验收标准（初稿）

1. **WHEN** 查询包含赠菜原因的订单 **THEN** 系统 **SHALL** 显示下单时保存的赠菜原因快照
2. **IF** 后台删除了某个赠菜原因 **THEN** 历史订单 **SHALL** 仍然显示该原因的原始名称
3. **IF** 后台修改了某个赠菜原因的名称 **THEN** 历史订单 **SHALL** 显示修改前的原始名称
4. **IF** 订单快照数据为空（历史数据） **THEN** 系统 **SHALL** 降级使用关联表数据（兼容性）
5. **WHEN** 创建新订单并选择赠菜原因 **THEN** 系统 **SHALL** 正确保存赠菜原因快照字段（JSON 格式）
6. **WHEN** 查询订单赠菜原因 **THEN** 系统 **SHALL** 返回多语言格式（`LocaleResponse`）
7. **IF** 快照字段 JSON 解析失败 **THEN** 系统 **SHALL** 降级使用关联表数据

### 技术方案要点（初稿）

#### 修改赠菜原因获取方法

```go
// GetGiftReason 获取赠菜原因（多语言）
// 优先使用快照字段，降级使用关联表数据，支持多语言
// 快照字段保存多语言（JSON）
// Requirement: story-main-gift-reason-snapshot-fix
func (model *SaleOrderProduct) GetGiftReason() dto.LocaleResponse {
	zhNames := make([]string, 0)
	thNames := make([]string, 0)
	enNames := make([]string, 0)
	zhtwNames := make([]string, 0)
	jaNames := make([]string, 0)
	koNames := make([]string, 0)
	myNames := make([]string, 0)
	trNames := make([]string, 0)
	svNames := make([]string, 0)
	
	// 遍历选择的赠品原因
	for _, reason := range model.CancelReasons {
		if !reason.IsGiftReason() {
			continue
		}
		
		// 优先使用快照字段（JSON）
		snapshotName := reason.Name
		if snapshotName != "" {
			var snapshotLocale dto.LocaleResponse
			if err := json.Unmarshal([]byte(snapshotName), &snapshotLocale); err == nil {
				if !snapshotLocale.IsNull() {
					// 快照数据有效，使用快照数据
					zhNames = append(zhNames, snapshotLocale.ZH)
					thNames = append(thNames, snapshotLocale.TH)
					enNames = append(enNames, snapshotLocale.EN)
					zhtwNames = append(zhtwNames, snapshotLocale.ZHTW)
					jaNames = append(jaNames, snapshotLocale.JA)
					koNames = append(koNames, snapshotLocale.KO)
					myNames = append(myNames, snapshotLocale.MY)
					trNames = append(trNames, snapshotLocale.TR)
					svNames = append(svNames, snapshotLocale.SV)
					continue
				}
			}
		}
		
		// 降级：如果快照字段为空或解析失败，使用关联表（兼容历史数据）
		if reason.MultiLanguageName != nil {
			zhNames = append(zhNames, reason.MultiLanguageName.ZhName)
			thNames = append(thNames, reason.MultiLanguageName.ThName)
			enNames = append(enNames, reason.MultiLanguageName.EnName)
			zhtwNames = append(zhtwNames, reason.MultiLanguageName.ZhTwName)
			jaNames = append(jaNames, reason.MultiLanguageName.JaName)
			koNames = append(koNames, reason.MultiLanguageName.KoName)
			myNames = append(myNames, reason.MultiLanguageName.MyName)
			trNames = append(trNames, reason.MultiLanguageName.TrName)
			svNames = append(svNames, reason.MultiLanguageName.SvName)
		}
	}
	
	// 添加自定义的赠菜原因
	if model.GiftReason != "" {
		zhNames = append(zhNames, model.GiftReason)
		thNames = append(thNames, model.GiftReason)
		enNames = append(enNames, model.GiftReason)
		zhtwNames = append(zhtwNames, model.GiftReason)
		jaNames = append(jaNames, model.GiftReason)
		koNames = append(koNames, model.GiftReason)
		myNames = append(myNames, model.GiftReason)
		trNames = append(trNames, model.GiftReason)
		svNames = append(svNames, model.GiftReason)
	}
	
	reasonDto := dto.LocaleResponse{
		ZH:   strings.Join(zhNames, "、"),
		TH:   strings.Join(thNames, "、"),
		EN:   strings.Join(enNames, "、"),
		ZHTW: strings.Join(zhtwNames, "、"),
		JA:   strings.Join(jaNames, "、"),
		KO:   strings.Join(koNames, "、"),
		MY:   strings.Join(myNames, "、"),
		TR:   strings.Join(trNames, "、"),
		SV:   strings.Join(svNames, "、"),
	}
	return reasonDto
}
```

#### 下单时保存快照

在创建 `SaleOrderProductReason` 时，如果 `GiftReasonUuid` 不为空，需要保存快照字段：

```go
// 创建赠菜原因时保存快照
if giftReasonUuid != 0 {
	// 序列化多语言数据为 JSON
	var nameJSON string
	if !giftReason.MultiLanguageName.IsNullName() {
		localeResp := giftReason.MultiLanguageName.GetNames()
		jsonData, err := json.Marshal(localeResp)
		if err == nil {
			nameJSON = string(jsonData)
		}
	}
	
	saleOrderProductReason := &SaleOrderProductReason{
		// ... 其他字段
		GiftReasonUuid: giftReasonUuid,
		// 保存快照字段（JSON 格式，包含所有语言）
		Name: nameJSON,
	}
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

