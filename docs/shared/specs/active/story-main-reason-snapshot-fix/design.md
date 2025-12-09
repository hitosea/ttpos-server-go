# 原因信息快照修复 设计文档

> 本文档定义原因信息快照修复功能的技术设计和实现方案。

## 📋 概述

本功能为 `ttpos_sale_order_product_reason` 表添加原因名称快照字段（`name`），确保订单历史信息准确反映免单/退菜时的真实原因状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 数据库结构变更：添加 `name` 快照字段（TEXT 类型，存储 JSON）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 统一处理：免单原因和退菜原因共用同一个快照字段

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ 不修改 Service 层（无需新增 Service）
- ✅ Model 层添加字段和方法（`SaleOrderProductReason.Name` 和快照相关方法）
- ✅ 修改下单和查询逻辑，使用快照数据
- ✅ 不使用 panic，返回 error
- ✅ 遵循单一职责原则

### 数据库规范 (database.mdc)

数据库设计遵循规范：

- ✅ 字段使用 TEXT 类型（存储 JSON）
- ✅ 字段默认值为空字符串
- ✅ 字段注释明确说明"不随后台更新"
- ✅ 迁移脚本支持可重复执行（幂等性）
- ✅ 使用 `ALTER TABLE ADD COLUMN` 安全添加字段

---

## 🔄 代码复用分析

### 可复用的现有组件

本功能主要是修改现有逻辑，无需新增 Service 或 Repository：

1. **SaleOrderProductReason 模型**: `main/app/model/order.go`
   - 添加 `Name` 字段
   - 用于存储免单/退菜原因快照（JSON 格式）

2. **FreeReason 和 ReturnFoodReason 模型**: `main/app/model/reason.go`
   - 已有的关联表模型
   - 用于降级查询

3. **MultiLanguageName 模型**: `main/app/model/multi_language_name.go`
   - 已有的多语言模型
   - 用于多语言数据获取

4. **DTO LocaleResponse**: `main/app/dto/locale.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

5. **参考实现**: `main/app/model/sale_bill.go` - `GetLocaleOrderSourceName()` 和 `SetOrderSourceNameSnapshot()`
   - 已实现的 JSON 方案快照逻辑
   - 可直接复用实现模式

### 集成点

**下单逻辑集成**：
- 在创建 `SaleOrderProductReason` 时（免单原因），从 `FreeReason.MultiLanguageName` 获取完整多语言数据
- 在创建 `SaleOrderProductReason` 时（退菜原因），从 `ReturnFoodReason.MultiLanguageName` 获取完整多语言数据
- 序列化为 JSON 保存到 `SaleOrderProductReason.Name` 字段
- 涉及所有免单/退菜入口

**查询逻辑集成**：
- 在订单查询时，使用快照字段（JSON）获取免单/退菜原因
- 修改 `SaleOrder.GetFreeReason()` 方法，优先使用快照字段
- 修改 `SaleOrderProduct.GetCancelReason()` 方法，优先使用快照字段
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表变更

#### 修改表: ttpos_sale_order_product_reason

**变更说明**：在 `ttpos_sale_order_product_reason` 表添加 `name` 字段，用于保存免单/退菜原因名称快照（JSON 格式，包含所有语言）。

**迁移 SQL**：

```sql
-- 添加原因名称快照字段（JSON 方案）
ALTER TABLE `ttpos_sale_order_product_reason` 
ADD COLUMN `name` TEXT NOT NULL DEFAULT '' 
COMMENT '原因名称快照（JSON），不随后台更新' 
AFTER `gift_reason_uuid`;
```

**字段说明**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| name | TEXT | 原因名称快照（JSON），不随后台更新 | DEFAULT '' |

**JSON 格式示例**：

```json
{"zh":"员工福利","th":"","en":"Employee Benefit","zhtw":"","ja":"","ko":"","my":"","tr":"","sv":""}
```

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_reason_name_to_sale_order_product_reason.php`

**迁移脚本要点**：

- 检查字段是否已存在（幂等性）
- 字段类型为 TEXT
- 默认值为空字符串
- 字段位置在 `gift_reason_uuid` 之后

---

## 📊 数据模型

### Go Model

#### SaleOrderProductReason 模型修改

**文件**: `main/app/model/order.go`

**变更**：添加 `Name` 字段

```go
// SaleOrderProductReason 销售订单产品各种原因 `ttpos_sale_order_product_reason`
type SaleOrderProductReason struct {
	// 基础字段
	BaseModel
	// 关联ID字段
	SaleOrderUuid         uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	SaleOrderProductUuid  uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单商品ID" json:"sale_order_product_uuid"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;not null;default:0;comment:多语言名称ID" json:"multi_language_name_uuid"`
	// 三选一。
	ReturnFoodReasonUuid uint64 `gorm:"column:return_food_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:退菜原因ID" json:"return_food_reason_uuid"`
	FreeReasonUuid       uint64 `gorm:"column:free_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:免单原因ID" json:"free_reason_uuid"`
	GiftReasonUuid       uint64 `gorm:"column:gift_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:赠菜原因ID" json:"gift_reason_uuid"`
	
	// 快照字段（JSON 方案）
	Name string `gorm:"column:name;type:text;default:'';comment:原因名称快照（JSON），不随后台更新" json:"name"`

	// 关联对象
	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
```

---

## 🔌 核心方法设计

### 查询逻辑修改

#### 1. SaleOrder.GetFreeReason() 方法修改

**文件**: `main/app/model/sale_order_ext_getset.go`

**变更说明**：修改免单原因获取方法，优先使用快照字段（JSON），降级使用关联表数据。

**实现逻辑**：

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

#### 2. SaleOrderProduct.GetCancelReason() 方法修改

**文件**: `main/app/model/sale_order_product.go`

**变更说明**：修改退菜原因获取方法，优先使用快照字段（JSON），降级使用关联表数据。

**实现逻辑**：

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

### 下单逻辑修改

#### 1. SaleOrder.NewFreeOrderReason() 方法修改

**文件**: `main/app/model/sale_order.go`

**变更说明**：修改免单原因创建方法，保存快照字段（JSON 格式）。

**实现逻辑**：

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
		
		list = append(list, &SaleOrderProductReason{
			BaseModel: BaseModel{
				Uuid: reasonUuid,
			},
			SaleOrderUuid:         model.Uuid,
			MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
			FreeReasonUuid:        reason.Uuid,
			// 保存快照字段（JSON 格式，包含所有语言）
			Name: nameJSON,
		})
	}
	return list
}
```

#### 2. SaleOrderProduct.NewSaleOrderProductReasonList() 方法修改

**文件**: `main/app/model/sale_order_product.go`

**变更说明**：修改退菜原因创建方法，保存快照字段（JSON 格式）。

**实现逻辑**：

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
		
		list = append(list, &SaleOrderProductReason{
			BaseModel: BaseModel{
				Uuid: reasonUuid,
			},
			SaleOrderUuid:         model.SaleOrderUuid,
			SaleOrderProductUuid:  model.Uuid,
			ReturnFoodReasonUuid:  reason.Uuid,
			MultiLanguageNameUuid: reason.MultiLanguageNameUuid,
			// 保存快照字段（JSON 格式，包含所有语言）
			Name: nameJSON,
		})
	}
	return list
}
```

---

## 🚨 错误处理

### JSON 解析失败

**场景**：快照字段包含无效的 JSON 数据

**处理方式**：
- 记录错误日志（不中断流程）
- 降级使用关联表数据
- 返回多语言响应

**代码示例**：

```go
if err := json.Unmarshal([]byte(snapshotJSON), &snapshotLocale); err != nil {
	// 记录错误日志，但不中断流程
	logger.Logger.Warn("解析原因快照 JSON 失败", zap.Error(err), zap.String("snapshot", snapshotJSON))
	// 继续降级逻辑
}
```

### 多语言数据为空

**场景**：关联表数据不存在或已删除

**处理方式**：
- 快照字段为空时，降级使用关联表数据
- 关联表数据也不存在时，返回空响应
- 确保历史订单正常显示

---

## 🔒 安全设计

### JSON 解析安全

- 使用标准库 `encoding/json` 进行解析
- 限制 JSON 大小（TEXT 类型有数据库限制）
- 解析失败时降级处理，不抛出异常

### 数据完整性

- 快照字段保存完整多语言数据
- 历史数据通过降级逻辑兼容
- 新订单自动使用快照机制

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**：

- Model 层方法: 80%+
- **Order 相关模块: 100%**（高风险）

**测试内容**：

- `GetFreeReason()` 方法：快照字段存在/不存在、JSON 解析成功/失败
- `GetCancelReason()` 方法：快照字段存在/不存在、JSON 解析成功/失败
- `NewFreeOrderReason()` 方法：序列化成功/失败
- `NewSaleOrderProductReasonList()` 方法：序列化成功/失败

### 集成测试

**测试流程**：

- 创建免单原因，验证快照字段保存
- 创建退菜原因，验证快照字段保存
- 查询订单，验证快照字段使用
- 删除关联表数据，验证降级逻辑

---

## 📈 性能优化

### JSON 解析优化

- 避免重复解析（缓存解析结果）
- 使用标准库 `encoding/json`（性能已优化）
- 快照字段优先，减少关联查询

### 数据库优化

- 快照字段使用 TEXT 类型（支持大 JSON）
- 不需要额外索引（查询时通过主键关联）

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 修改 Go Model（添加 `Name` 字段）

### Phase 2: 核心实现

- [ ] 修改 `GetFreeReason()` 方法
- [ ] 修改 `GetCancelReason()` 方法
- [ ] 修改 `NewFreeOrderReason()` 方法
- [ ] 修改 `NewSaleOrderProductReasonList()` 方法

### Phase 3: 测试

- [ ] 单元测试
- [ ] 集成测试
- [ ] 回归测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: xiezhihuan  
**审核者**: {审核者}

