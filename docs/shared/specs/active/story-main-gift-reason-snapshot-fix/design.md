# 赠菜原因快照修复 设计文档

> 本文档定义赠菜原因快照修复功能的技术设计和实现方案。

## 📋 概述

本功能使用 `ttpos_sale_order_product_reason` 表现有的 `name` 字段（TEXT 类型，JSON 快照），修复 `GetGiftReason()` 方法优先使用快照字段，确保订单历史信息准确反映下单时的真实状态，不随后台配置变更而改变。采用 JSON 方案保存完整多语言数据，既保证快照完整性，又提供完整的多语言支持。

**核心特性**：
- 无需数据库结构变更（快照字段已存在）
- 查询逻辑优化：优先使用快照数据，降级使用关联表
- 多语言支持：快照保存完整多语言 JSON，直接返回无需补充
- 兼容性处理：历史数据通过降级逻辑正常显示
- 下单逻辑修改：确保创建赠菜原因时保存快照字段

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计严格遵循 Go Main 开发规范：

- ✅ 不修改 Service 层（无需新增 Service）
- ✅ Model 层修改方法（`SaleOrderProduct.GetGiftReason()`）
- ✅ 修改下单逻辑，使用快照数据
- ✅ 不使用 panic，返回 error
- ✅ 遵循单一职责原则

### 数据库规范 (database.mdc)

数据库设计遵循规范：

- ✅ 使用现有字段 `name`（TEXT 类型，存储 JSON）
- ✅ 字段默认值为空字符串
- ✅ 字段注释明确说明"不随后台更新"
- ✅ 无需数据库迁移（字段已存在）

---

## 🔄 代码复用分析

### 可复用的现有组件

本功能主要是修改现有逻辑，无需新增 Service 或 Repository：

1. **SaleOrderProductReason 模型**: `main/app/model/order.go`
   - 已有 `Name` 字段（TEXT 类型，JSON 快照）
   - 用于存储赠菜原因快照（JSON 格式）

2. **FreeReason 模型**: `main/app/model/reason.go`
   - 已有的关联表模型（赠菜原因使用 FreeReason 表）
   - 用于降级查询

3. **MultiLanguageName 模型**: `main/app/model/multi_language_name.go`
   - 已有的多语言模型
   - 用于多语言数据获取

4. **DTO LocaleResponse**: `main/app/dto/common_resp.go`
   - 已有的多语言响应结构
   - 用于返回多语言数据

5. **参考实现**: 
   - `main/app/model/sale_order_product.go:988` - `GetCancelReason()` 方法（退菜原因快照逻辑）
   - `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()` 方法（JSON 方案快照逻辑）
   - 可直接复用实现模式

### 集成点

**下单逻辑集成**：
- 在创建 `SaleOrderProductReason` 时（赠菜原因），从 `FreeReason.MultiLanguageName` 获取完整多语言数据
- 序列化为 JSON 保存到 `SaleOrderProductReason.Name` 字段
- 涉及所有赠菜入口（`CreateSaleOrderProductReasons` 方法）

**查询逻辑集成**：
- 在订单查询时，使用快照字段（JSON）获取赠菜原因
- 修改 `SaleOrderProduct.GetGiftReason()` 方法，优先使用快照字段
- 涉及所有订单查询接口（订单详情、订单列表、报表等）

---

## 🗄️ 数据库设计

### 数据表说明

#### 表: ttpos_sale_order_product_reason

**字段说明**：使用现有 `name` 字段（TEXT 类型），用于保存赠菜原因名称快照（JSON 格式，包含所有语言）。

**字段详情**：

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| name | TEXT | 原因名称快照（JSON），不随后台更新 | DEFAULT '' |

**JSON 格式示例**：

```json
{"zh":"会员生日福利","th":"","en":"Member Birthday Benefit","zhtw":"","ja":"","ko":"","my":"","tr":"","sv":""}
```

**注意**：
- 字段已存在，无需数据库迁移
- 该字段同时用于免单、退菜、赠菜原因的快照
- 字段类型为 TEXT，支持存储 JSON 数据

---

## 📊 数据模型

### Go Model

#### SaleOrderProductReason 模型

**文件**: `main/app/model/order.go`

**字段说明**：`Name` 字段已存在，无需修改

```go
// SaleOrderProductReason 销售订单产品各种原因 `ttpos_sale_order_product_reason`
type SaleOrderProductReason struct {
	// ... 其他字段
	GiftReasonUuid uint64 `gorm:"column:gift_reason_uuid;type:bigint(20) unsigned;not null;default:0;comment:赠菜原因ID" json:"gift_reason_uuid"`
	
	// 快照字段（JSON 方案）
	// Requirement: story-main-reason-snapshot-fix
	Name string `gorm:"column:name;type:text;default:'';comment:原因名称快照（JSON），不随后台更新" json:"name"`
	
	// 关联对象
	MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
```

---

## 🔌 核心方法设计

### 查询逻辑修改

#### SaleOrderProduct.GetGiftReason() 方法修改

**文件**: `main/app/model/sale_order_product.go:1073`

**变更说明**：修改赠菜原因获取方法，优先使用快照字段（JSON），降级使用关联表数据。

**实现逻辑**：

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
		if reason.IsDelete() {
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
					zhtwNames = append(zhtwNames, snapshotLocale.ZHTW)
					jaNames = append(jaNames, snapshotLocale.JA)
					koNames = append(koNames, snapshotLocale.KO)
					myNames = append(myNames, snapshotLocale.MY)
					trNames = append(trNames, snapshotLocale.TR)
					svNames = append(svNames, snapshotLocale.SV)
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
			zhtwNames = append(zhtwNames, multiLang.ZHTW)
			jaNames = append(jaNames, multiLang.JA)
			koNames = append(koNames, multiLang.KO)
			myNames = append(myNames, multiLang.MY)
			trNames = append(trNames, multiLang.TR)
			svNames = append(svNames, multiLang.SV)
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

**关键逻辑**：
1. 遍历 `model.CancelReasons`，筛选出 `IsGiftReason()` 为 true 的原因
2. 优先使用 `reason.Name` 快照字段（JSON）
3. 解析 JSON 为 `dto.LocaleResponse`
4. 快照为空或解析失败时，降级使用 `reason.MultiLanguageName`
5. 处理自定义赠菜原因（`model.GiftReason` 字段）
6. 返回多语言格式，多个原因用"、"分隔

---

### 下单逻辑修改

#### CreateSaleOrderProductReasons() 方法修改

**文件**: `main/app/repository/sale_order_product.go:209`

**变更说明**：修改批量创建销售订单商品原因方法，在创建赠菜原因时保存快照字段（JSON 格式）。

**实现逻辑**：

```go
// CreateSaleOrderProductReasons 批量创建销售订单商品原因
func (r *saleOrderProductRepo) CreateSaleOrderProductReasons(
	saleOrderUuid uint64,
	saleOrderProductUuid uint64,
	source string,
	returnFoodReasons [][2]uint64,
) error {
	if len(returnFoodReasons) == 0 {
		return nil
	}
	db := r.db
	
	// 如果是赠菜原因，需要加载 FreeReason 数据以获取多语言名称
	var giftReasonsMap map[uint64]*model.FreeReason
	if source == constant.ProductReasonTypeGift {
		giftReasonUuids := make([]uint64, 0, len(returnFoodReasons))
		for _, reason := returnFoodReasons {
			giftReasonUuids = append(giftReasonUuids, reason[0])
		}
		giftReasons, err := base.NewGiftOrFreeOrderReasonRepo(db).GetFreeOrderReasonListByUuids(giftReasonUuids)
		if err == nil {
			giftReasonsMap = make(map[uint64]*model.FreeReason)
			for _, reason := range giftReasons {
				giftReasonsMap[reason.Uuid] = reason
			}
		}
	}
	
	// 构建批量插入数据
	reasons := make([]*model.SaleOrderProductReason, len(returnFoodReasons))
	for i, reason := range returnFoodReasons {
		reasons[i] = &model.SaleOrderProductReason{
			SaleOrderUuid:         saleOrderUuid,
			SaleOrderProductUuid:  saleOrderProductUuid,
			MultiLanguageNameUuid: reason[1],
		}
		
		// 序列化多语言数据为 JSON（仅赠菜原因）
		var nameJSON string
		if source == constant.ProductReasonTypeGift {
			if giftReason, ok := giftReasonsMap[reason[0]]; ok {
				if giftReason.MultiLanguageName != nil && !giftReason.MultiLanguageName.IsNullName() {
					localeResp := giftReason.MultiLanguageName.GetNames()
					jsonData, err := json.Marshal(localeResp)
					if err == nil {
						nameJSON = string(jsonData)
					}
				}
			}
		}
		
		if source == constant.ProductReasonTypeReturnFood {
			reasons[i].ReturnFoodReasonUuid = reason[0]
		}
		if source == constant.ProductReasonTypeGift {
			reasons[i].GiftReasonUuid = reason[0]
			// 保存快照字段（JSON 格式，包含所有语言）
			reasons[i].Name = nameJSON
		}
		if source == constant.ProductReasonTypeFree {
			reasons[i].FreeReasonUuid = reason[0]
		}
	}
	// 批量创建
	return db.Create(&reasons).Error
}
```

**关键逻辑**：
1. 如果是赠菜原因（`source == constant.ProductReasonTypeGift`），加载 `FreeReason` 数据
2. 从 `FreeReason.MultiLanguageName` 获取完整多语言数据
3. 序列化为 JSON 字符串
4. 保存到 `SaleOrderProductReason.Name` 字段
5. 如果序列化失败，`Name` 字段为空（降级使用关联表）

**注意**：
- 需要加载 `FreeReason` 数据以获取多语言名称
- 序列化失败时，`Name` 字段为空，降级使用关联表数据
- 不影响现有退菜和免单原因的创建逻辑

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
if snapshotJSON != "" {
	if err := json.Unmarshal([]byte(snapshotJSON), &snapshotLocale); err == nil {
		if !snapshotLocale.IsNull() {
			// 使用快照数据
			continue
		}
	} else {
		// JSON 解析失败，记录日志但不中断流程
		logger.Logger.Warn("赠菜原因快照 JSON 解析失败", zap.Error(err), zap.String("snapshot", snapshotJSON))
	}
}
// 降级使用关联表数据
```

### 关联表数据不存在

**场景**：快照字段为空，且关联表数据也不存在（已删除）

**处理方式**：
- 跳过该原因
- 继续处理其他原因
- 返回已处理的原因列表

---

## 🔍 影响范围分析

### 修改的文件

1. **Model 层**：
   - `main/app/model/sale_order_product.go` - `GetGiftReason()` 方法

2. **Repository 层**：
   - `main/app/repository/sale_order_product.go` - `CreateSaleOrderProductReasons()` 方法

### 影响的接口

1. **订单查询接口**：
   - 订单详情查询（使用 `GetGiftReason()`）
   - 订单列表查询（使用 `GetGiftReason()`）
   - 订单打印（使用 `GetGiftReason()`）
   - 订单导出（使用 `GetGiftReason()`）

2. **下单接口**：
   - 创建订单（使用 `CreateSaleOrderProductReasons()`）
   - 修改订单（使用 `CreateSaleOrderProductReasons()`）

### 测试影响

- 需要测试所有使用赠菜原因的订单查询场景
- 需要测试所有创建赠菜原因的订单创建场景
- 需要测试快照字段为空的历史订单兼容性

---

## 📝 实现细节

### Import 依赖

```go
import (
	"encoding/json"
	"strings"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/logger"
	"go.uber.org/zap"
)
```

### 错误处理策略

1. **JSON 解析失败**：降级使用关联表数据，记录警告日志
2. **序列化失败**：`Name` 字段为空，降级使用关联表数据，记录错误日志
3. **关联表数据不存在**：跳过该原因，继续处理其他原因

### 性能考虑

1. **批量查询优化**：在 `CreateSaleOrderProductReasons()` 中批量加载 `FreeReason` 数据，避免 N+1 查询
2. **JSON 解析缓存**：无需缓存，每次查询都解析（数据量小）
3. **降级查询**：仅在快照字段为空时查询关联表，减少数据库查询

---

## 🧪 测试策略

### 单元测试

**测试文件**: `main/app/model/sale_order_product_test.go`

**测试场景**：
1. 快照字段存在且 JSON 有效 → 返回快照数据
2. 快照字段存在但 JSON 无效 → 降级使用关联表数据
3. 快照字段为空 → 降级使用关联表数据
4. 关联表数据不存在 → 跳过该原因
5. 多个赠菜原因组合 → 返回组合结果
6. 自定义赠菜原因 → 返回自定义原因

### 集成测试

**测试场景**：
1. 创建订单（包含赠菜原因） → 验证快照字段保存成功（JSON 格式）
2. 删除赠菜原因配置 → 查询订单仍显示快照数据
3. 修改赠菜原因名称 → 查询订单仍显示修改前的名称

---

## 📚 参考实现

### 类似功能实现

1. **退菜原因快照**: `main/app/model/sale_order_product.go:988` - `GetCancelReason()` 方法
2. **免单原因快照**: `main/app/model/sale_order_ext_getset.go` - `GetFreeReason()` 方法
3. **外卖来源快照**: `main/app/model/sale_bill.go:789` - `GetLocaleOrderSourceName()` 方法

### 实现模式

所有快照功能都遵循相同的实现模式：
1. 优先使用快照字段（JSON）
2. 解析 JSON 为 `dto.LocaleResponse`
3. 快照为空或解析失败时，降级使用关联表数据
4. 返回多语言格式

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: xiezhihuan  
**审核者**: {审核者}

