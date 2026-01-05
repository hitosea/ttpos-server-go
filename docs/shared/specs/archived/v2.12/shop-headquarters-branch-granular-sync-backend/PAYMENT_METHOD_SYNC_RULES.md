# 支付方式同步特殊规则说明

> 本文档详细说明支付方式同步的复杂业务规则，供开发人员参考

---

## 📋 概述

支付方式同步与其他数据类型（优惠券、满额减、菜品标签等）有显著不同，需要特殊处理。

**关键差异**：
1. ❌ **不删除**未勾选的总部数据
2. 🎯 **同名判断**基于 `payment_name`
3. 🔢 **特殊 code** 有不同的处理逻辑
4. 🆕 **新增时** code 需要自动生成

---

## 🔍 规则详解

### 规则1：获取可同步列表时过滤

**过滤条件**：不显示 `code = 40` 和 `code = 10` 的支付方式

```sql
SELECT * FROM ttpos_payment_method 
WHERE headquarter_uuid = 0 
  AND delete_time = 0
  AND code NOT IN (40, 10)  -- 关键：过滤这两个code
```

**原因**：这两个 code 是系统保留的支付方式，不允许同步。

---

### 规则1.1：已同步状态判断（⚠️ 特殊）

**与其他数据类型的区别**：

| 数据类型 | 已同步判断方式 |
|---------|---------------|
| 优惠券、满额减、菜品标签等 | 查询分店 `headquarter_uuid = 总部uuid` |
| **支付方式** | **通过 `payment_name` 匹配** |

**原因**：支付方式同步规则复杂
- 同名且为普通code时，跳过同步（不会设置 `headquarter_uuid`）
- 同名且为特殊code时，只更新 `headquarter_uuid`
- 因此不能简单通过 `headquarter_uuid` 判断，需要通过名称匹配

**实现逻辑**：

```go
func (s *SyncSrv) getPaymentMethodGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
    // 1. 查询总部支付方式（过滤 code=40 和 code=10）
    var hqPayments []model.PaymentMethod
    err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
        Where("code NOT IN (?)", []int{40, 10}).
        Find(&hqPayments).Error
    
    // 2. 查询分店所有支付方式的名称列表（不限 headquarter_uuid）
    var subShopPaymentNames []string
    err = subShopDB.Model(&model.PaymentMethod{}).
        Where("delete_time = 0").
        Pluck("payment_name", &subShopPaymentNames).Error
    
    // 3. 构建分店已有名称的map
    subShopNameMap := make(map[string]bool)
    for _, name := range subShopPaymentNames {
        subShopNameMap[name] = true
    }
    
    // 4. 匹配总部支付方式，找出已同步的uuid
    var syncedUuids []uint64
    for _, hqPayment := range hqPayments {
        if subShopNameMap[hqPayment.PaymentName] {
            // 分店有同名支付方式，视为已同步
            syncedUuids = append(syncedUuids, hqPayment.Uuid)
        }
    }
    
    // 5. 组装数据项
    var items []resp.DataItem
    for _, hqPayment := range hqPayments {
        items = append(items, resp.DataItem{
            Uuid:        hqPayment.Uuid,
            Name:        hqPayment.PaymentName,
            RelatedData: []resp.RelatedData{},
            AdditionalInfo: map[string]any{
                "code": hqPayment.Code,
            },
        })
    }
    
    return resp.DataGroup{
        Type:        constant.SyncDataTypePaymentMethod,
        TypeName:    constant.SyncDataTypeNames[constant.SyncDataTypePaymentMethod],
        Items:       items,
        SyncedUuids: syncedUuids, // 通过名称匹配得到的uuid列表
    }, nil
}
```

**关键点**：
1. 查询分店所有支付方式的名称（不限 `headquarter_uuid`）
2. 遍历总部支付方式，通过名称匹配判断是否已同步
3. 已同步的总部 uuid 加入 `syncedUuids` 列表

---

### 规则2：删除策略（不删除）

**与其他数据类型的区别**：

| 数据类型 | 删除策略 |
|---------|---------|
| 优惠券、满额减、菜品标签等 | ✅ 删除未勾选的总部数据 |
| **支付方式** | ❌ **不删除**未勾选的总部数据 |

**实现**：

```go
deleteTasks := []struct {
    TableName  string
    Uuids      []uint64
    SkipDelete bool
}{
    {"ttpos_marketing_coupon", syncData.Coupon, false},
    {"ttpos_payment_method", syncData.PaymentMethod, true}, // 标记为true，跳过删除
}

for _, task := range deleteTasks {
    if task.SkipDelete {
        continue // 支付方式不删除
    }
    // ...删除逻辑
}
```

---

### 规则3：同名判断依据

**判断字段**：`payment_name`（不是 `name`，不是 `code`）

```go
// 检查分店是否已有同名支付方式
var existPayment model.PaymentMethod
err := subShopDB.Where("payment_name = ? AND delete_time = 0", hqPayment.PaymentName).
    First(&existPayment).Error
```

---

### 规则4：特殊 code 处理

**特殊 code 列表**：`90111`, `90222`, `90333`

**处理逻辑**：

```go
specialCodes := map[int]bool{
    90111: true,
    90222: true,
    90333: true,
}

if err == nil {
    // 分店已有同名支付方式
    if specialCodes[existPayment.Code] {
        // 特殊code：只更新 headquarter_uuid，不跳过
        subShopDB.Model(&model.PaymentMethod{}).
            Where("id = ?", existPayment.ID).
            Update("headquarter_uuid", headquarterUuid)
    } else {
        // 普通code：跳过同步
        logger.Logger.Info("支付方式已存在，跳过同步", 
            zap.String("name", hqPayment.PaymentName))
        continue
    }
}
```

**流程图**：

```
查询分店是否有同名支付方式（payment_name）
    |
    ├─ 不存在 ─────────────> 创建新支付方式（见规则5）
    |
    └─ 存在
        |
        ├─ code = 90111, 90222, 90333 ──> 更新 headquarter_uuid
        |
        └─ 其他 code ──────────────────> 跳过同步
```

---

### 规则5：创建新支付方式

当分店不存在同名支付方式时，创建新记录：

**字段赋值规则**：

| 字段 | 赋值规则 | 说明 |
|------|---------|------|
| `payment_name` | 总部值 | 复制总部 |
| `code` | **自动生成** | 与手动添加（source=1）规则一致 |
| `headquarter_uuid` | 总部uuid | 标记来源 |
| `logo_file_uuid` | **固定为 0** | 不使用总部logo |
| `source` | 1 | 标记为手动添加类型 |
| `qrcode_file_uuid` | **数据库默认值** | 不显式设置 |
| `fee_percent` | **数据库默认值** | 不显式设置 |
| `is_show_cashier` | **数据库默认值** | 不显式设置 |
| `is_show_assistant` | **数据库默认值** | 不显式设置 |
| `is_show_member_recharge` | **数据库默认值** | 不显式设置 |
| `status` | **数据库默认值** | 不显式设置 |
| `sort` | **数据库默认值** | 不显式设置 |
| `default_img` | **数据库默认值** | 不显式设置 |
| `erpnext_payment` | **数据库默认值** | 不显式设置 |

**代码示例**：

```go
newPayment := model.PaymentMethod{
    PaymentName:     hqPayment.PaymentName,
    Code:            generatePaymentCode(subShopDB), // 自动生成
    HeadquarterUuid: headquarterUuid,
    LogoFileUuid:    0, // 固定为0
    Source:          1, // 手动添加类型
    // 以下字段不设置，使用数据库默认值
}
subShopDB.Create(&newPayment)
```

---

### 规则6：code 生成规则

**规则**：与手动添加（`source = 1`）一致

**通常实现**：
- 查询当前最大的自定义 code（范围：90000-99999）
- 新 code = 最大 code + 1

**代码示例**：

```go
func generatePaymentCode(db *gorm.DB) int {
    var maxCode int
    db.Model(&model.PaymentMethod{}).
        Where("code >= 90000 AND code < 100000 AND delete_time = 0").
        Select("COALESCE(MAX(code), 89999)").
        Scan(&maxCode)
    
    return maxCode + 1
}
```

**注意**：
- 需要查询现有代码确认实际的 code 生成规则
- 可能有其他范围或逻辑，需要与手动添加保持一致

---

## 🎯 获取可同步列表流程

```
1. 查询总部支付方式列表（总部DB）
   └─ WHERE headquarter_uuid = 0 AND code NOT IN (40, 10)
   
2. 查询分店中从总部同步的支付方式名称列表（分店DB）
   └─ WHERE headquarter_uuid = 总部uuid
   └─ PLUCK payment_name
   
3. 遍历总部支付方式
   └─ 通过 payment_name 匹配分店已同步名称
       |
       ├─ 匹配成功 ──> 加入 synced_uuids 列表（总部uuid）
       └─ 匹配失败 ──> 不加入
       
4. 返回
   ├─ items: 总部支付方式列表
   └─ synced_uuids: 已同步的总部uuid列表（通过名称匹配）
```

**关键**：
- 查询条件：分店DB中 `headquarter_uuid = 总部uuid` 的支付方式名称
- 匹配逻辑：分店已同步名称 vs 总部支付方式名称
- 返回结果：匹配到的**总部uuid**列表

---

## 🎯 同步操作流程

```
1. 获取勾选的总部支付方式列表
   └─ 过滤 code=40 和 code=10
   
2. 遍历每个总部支付方式
   |
   ├─ 查询分店是否有同名（payment_name）
   |
   ├─ 不存在 ──> 创建新支付方式
   |   |
   |   ├─ code = generatePaymentCode()
   |   ├─ logo_file_uuid = 0
   |   ├─ source = 1
   |   └─ 其他字段用数据库默认值
   |
   └─ 存在
       |
       ├─ code = 90111, 90222, 90333
       |   └─ 更新 headquarter_uuid
       |
       └─ 其他 code
           └─ 跳过（保持分店原有）

3. 不删除分店中未勾选的总部支付方式
```

---

## ⚠️ 关键注意事项

1. **已同步状态判断（⚠️ 特殊）**：
   - 查询条件：分店DB中 `headquarter_uuid = 总部uuid` 的支付方式名称列表
   - 匹配逻辑：通过 `payment_name` 匹配总部支付方式
   - 返回结果：匹配到的**总部uuid**列表（用于前端默认勾选）

2. **不要删除**：
   - 支付方式不参与删除流程
   - `SkipDelete = true`

3. **同名判断**：
   - 使用 `payment_name` 字段
   - 大小写敏感（取决于数据库配置）

4. **特殊 code**：
   - 90111, 90222, 90333 需要特殊处理
   - 只更新 `headquarter_uuid`，不创建新记录

5. **code 生成**：
   - 必须与手动添加规则一致
   - 需要查询现有代码确认实际逻辑

6. **字段默认值**：
   - 很多字段使用数据库默认值
   - 不要显式设置为0或空字符串

---

## 🧪 测试用例

### 用例 1：首次同步支付方式

**前置条件**：
- 总部有支付方式 "支付宝"（code=90001）
- 分店没有同名支付方式

**预期结果**：
- 分店创建新支付方式 "支付宝"
- code 自动生成（如 90001）
- logo_file_uuid = 0
- headquarter_uuid = 总部uuid

### 用例 2：同名普通 code

**前置条件**：
- 总部有支付方式 "微信支付"（code=90002）
- 分店已有 "微信支付"（code=90002）

**预期结果**：
- 跳过同步
- 分店支付方式保持不变

### 用例 3：同名特殊 code

**前置条件**：
- 总部有支付方式 "现金"（code=90111）
- 分店已有 "现金"（code=90111, headquarter_uuid=0）

**预期结果**：
- 不跳过
- 只更新分店支付方式的 headquarter_uuid = 总部uuid
- 其他字段不变

### 用例 4：过滤特定 code

**前置条件**：
- 总部有支付方式 "系统支付1"（code=40）
- 总部有支付方式 "系统支付2"（code=10）

**预期结果**：
- 这两个支付方式不出现在可同步列表中
- 不会被同步到分店

### 用例 5：已同步状态判断（名称匹配）

**前置条件**：
- 总部数据库：
  - 支付方式 "微信支付"（uuid=111111, code=90002, headquarter_uuid=0）
  - 支付方式 "支付宝"（uuid=222222, code=90003, headquarter_uuid=0）
  - 支付方式 "银行卡"（uuid=333333, code=90004, headquarter_uuid=0）

- 分店数据库：
  - 支付方式 "微信支付"（uuid=999991, code=90002, **headquarter_uuid=总部uuid**）
  - 支付方式 "银行卡"（uuid=999992, code=90004, **headquarter_uuid=总部uuid**）
  - 支付方式 "现金"（uuid=999993, code=90111, headquarter_uuid=0）- 分店自建

**调用接口**：
- 分店调用 `GetHeadquartersDataList`

**处理流程**：
1. 查询总部DB：获取总部支付方式列表（111111, 222222, 333333）
2. 查询分店DB：`WHERE headquarter_uuid = 总部uuid`，获取名称列表 ["微信支付", "银行卡"]
3. 遍历总部支付方式，通过名称匹配：
   - "微信支付" ✅ 匹配 → 总部uuid 111111 加入 synced_uuids
   - "支付宝" ❌ 不匹配
   - "银行卡" ✅ 匹配 → 总部uuid 333333 加入 synced_uuids

**预期结果**：
```json
{
  "type": "payment_method",
  "synced_uuids": [111111, 333333],  // ✅ 已同步的总部uuid
  "items": [
    {"uuid": 111111, "name": "微信支付"},
    {"uuid": 222222, "name": "支付宝"},
    {"uuid": 333333, "name": "银行卡"}
  ]
}
```

**说明**：
- 只查询分店中 `headquarter_uuid = 总部uuid` 的支付方式名称
- 这些名称代表"已从总部同步的支付方式"
- 通过名称匹配总部DB，得到对应的总部uuid列表
- 前端根据 synced_uuids 默认勾选对应的总部支付方式

---

## 📝 数据库字段默认值

需要在数据库迁移文件或表结构中确认以下字段的默认值：

```sql
CREATE TABLE ttpos_payment_method (
    -- ...
    qrcode_file_uuid bigint DEFAULT 0,
    fee_percent decimal(5,2) DEFAULT 0.00,
    is_show_cashier tinyint DEFAULT 1,
    is_show_assistant tinyint DEFAULT 1,
    is_show_member_recharge tinyint DEFAULT 0,
    status tinyint DEFAULT 1,
    sort int DEFAULT 0,
    default_img varchar(255) DEFAULT '',
    erpnext_payment varchar(100) DEFAULT NULL,
    -- ...
);
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: 曾振华  
**关联任务**: DooTask #37462
