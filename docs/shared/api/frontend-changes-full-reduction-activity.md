# 满减营销功能 - 前端对接文档

> 本文档说明后端新增的满减营销活动功能，供前端开发人员对接使用。

**更新日期**: 2025-11-24  
**版本**: v1.0.0

---

## 📋 概述

后端新增了满减营销活动功能，支持在结账页面选择满减/每满减活动，自动计算活动抵扣金额。该功能已集成到现有的结账流程中。

---

## 🔄 API 变更

### 1. 新增 API：选择或取消满减活动

#### 收银端

**接口地址**: `POST /api/v1/cashier/desk/order/payment/activity`

**请求头**:
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求参数**:
```json
{
  "sale_bill_uuid": 123456,              // 销售账单UUID（必填）
  "sale_order_uuid": 789012,             // 销售订单UUID（必填）
  "full_reduction_activity_uuid": 345678  // 满减活动UUID，0表示取消活动
}
```

**响应格式**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "member_info": {...},
    "coupon_list": {...},
    "payment_orders": {...},
    "payment_methods": {...},
    "amounts": {...},
    "points_exchange": {...},
    "activity_list": {
      "list": [...]
    }
  }
}
```

**错误响应**:
```json
{
  "code": 0,
  "message": "活动信息已经变更，请重新确认",
  "data": {}
}
```

#### 助手端

**接口地址**: `POST /api/v1/assistant/desk/order/payment/activity`

**请求参数和响应格式与收银端相同**

---

### 2. 修改 API：获取结账页面信息

**接口地址**: `GET /api/v1/cashier/desk/order/payment/info`（收银端）  
**接口地址**: `GET /api/v1/assistant/desk/order/payment/info`（助手端）

**响应格式变更**:

在原有响应中新增了 `activity_list` 字段，并在 `amounts.list` 中的每个 `PaymentMethodAmount` 对象中新增了 `activity_amount` 字段：

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "member_info": {...},
    "coupon_list": {...},
    "payment_orders": {...},
    "payment_methods": {...},
    "amounts": {...},
    "points_exchange": {...},
    "activity_list": {                    // ⭐ 新增字段
      "list": [
        {
          "uuid": 123456,
          "locale_name": {
            "zh": "满200减20",
            "en": "200 off 20",
            "th": "...",
            "zhtw": "...",
            "ja": "...",
            "ko": "...",
            "my": "...",
            "tr": "...",
            "sv": "..."
          },
          "activity_type": 1,             // 活动类型：1-阶梯满减，2-循环满减
          "start_date": "2025-11-24",
          "end_date": "2025-12-31",
          "start_time": "09:00",
          "end_time": "22:00",
          "is_all_day": false,
          "rules": [
            {
              "threshold": 200,           // 满减阈值
              "discount": 20              // 减价金额
            }
          ],
          "is_available": true,           // 是否可用（是否在适用时段内、订单金额是否达到满减条件）
          "is_selected": false,           // 是否已选中
          "discount_amount": 0            // 抵扣金额（如果已选中，显示实际抵扣金额）
        }
      ]
    }
  }
}
```

---

## 📊 数据结构说明

### FullReductionActivityList

```typescript
interface FullReductionActivityList {
  list: FullReductionActivityItem[];
}
```

### FullReductionActivityItem

```typescript
interface FullReductionActivityItem {
  uuid: number;                          // 活动UUID
  locale_name: LocaleResponse;           // 活动多语言名称
  activity_type: number;                 // 活动类型：1-阶梯满减，2-循环满减
  start_date: string;                    // 开始日期（YYYY-MM-DD）
  end_date: string;                       // 结束日期（YYYY-MM-DD）
  start_time: string;                     // 开始时间（HH:mm）
  end_time: string;                       // 结束时间（HH:mm）
  is_all_day: boolean;                    // 是否全天
  rules: ActivityRule[];                 // 活动规则列表
  is_available: boolean;                  // 是否可用
  is_selected: boolean;                   // 是否已选中
  discount_amount: number;                // 抵扣金额（如果已选中）
}
```

### ActivityRule

```typescript
interface ActivityRule {
  threshold: number;                      // 满减阈值
  discount: number;                       // 减价金额
}
```

### PaymentMethodAmount（已修改）

在 `amounts.list` 中的 `PaymentMethodAmount` 对象中新增了 `activity_amount` 字段：

```typescript
interface PaymentMethodAmount {
  sale_order_origin_amount: number;       // 订单原价
  sale_order_cart_amount: number;         // 购物车应收金额
  sale_order_amount: number;              // 应收金额
  unpaid_amount: number;                  // 未收金额
  zero_amount: number;                    // 抹零金额
  zero_rule: number;                     // 结账抹零规格
  is_auto_zero: boolean;                  // 是否是自动抹零
  payment_method_uuid: number;            // 支付方式UUID
  code: number;                          // 支付方式代码
  coupon_exchange_amount: number;         // 优惠券抵扣金额
  activity_amount: number;                // ⭐ 新增：满减活动抵扣金额
  commission_fee: number;                 // 已付款的手续费
}
```

### LocaleResponse

```typescript
interface LocaleResponse {
  zh: string;                             // 中文
  en: string;                             // 英文
  th: string;                             // 泰文
  zhtw: string;                           // 繁体中文
  ja: string;                             // 日文
  ko: string;                             // 韩文
  my: string;                             // 缅甸文
  tr: string;                             // 土耳其文
  sv: string;                             // 瑞典文
}
```

---

## 🎯 业务规则

### 活动可用性判断

活动是否可用（`is_available`）由后端判断，基于以下条件：

1. ✅ **活动在有效期内**（开始日期 ≤ 当前日期 ≤ 结束日期）
2. ✅ **活动在适用时段内**（如果设置了时段限制）
3. ✅ **订单金额达到满减条件**（订单金额 ≥ 活动最小阈值）
4. ✅ **订单未部分支付**（如果已部分支付，则不可选择活动）
5. ✅ **积分抵扣后最终应收不为0**（如果为0，则不可选择活动）

### 活动类型说明

1. **阶梯满减** (`activity_type = 1`):
   - 根据订单金额，找到满足条件的最大规则
   - 例如：满100减10，满200减25，满300减40
   - 如果订单金额为250，则使用"满200减25"规则

2. **循环满减** (`activity_type = 2`):
   - 每满一定金额减一定金额，可循环使用
   - 例如：每满100减10
   - 如果订单金额为250，则使用2次，共减20

### 互斥规则

1. **活动与优惠券互斥**:
   - 如果已使用优惠券，则不能选择活动
   - 如果已选择活动，则不能使用优惠券
   - 选择活动时，如果已使用优惠券，会返回错误："活动与优惠券只能二选一"

2. **活动与积分抵扣互斥**:
   - 选择活动后，积分不再自动抵扣（`auto_points_exchange` 设置为 0）
   - 积分抵扣后最终应收为0时，不可选择活动
   - 选择活动时，如果积分抵扣后最终应收为0，会返回错误："积分抵扣后最终应收为0，不可选择满减活动"

### 活动选择/取消

- **选择活动**: 传入活动UUID（`full_reduction_activity_uuid > 0`）
- **取消活动**: 传入0（`full_reduction_activity_uuid = 0`）
- **替换活动**: 先取消当前活动，再选择新活动（或直接传入新活动UUID，后端会自动处理）

---

## ⚠️ 错误处理

### 常见错误码和提示

| 错误信息 | 说明 | 处理建议 |
|---------|------|---------|
| "活动信息已经变更，请重新确认" | 活动已失效或不在有效期内 | 刷新活动列表，重新选择 |
| "活动不在适用时段内" | 当前时间不在活动的适用时段内 | 提示用户活动当前不可用 |
| "订单金额未达到满减条件" | 订单金额小于活动最小阈值 | 提示用户需要达到一定金额才能使用 |
| "活动与优惠券只能二选一" | 已使用优惠券，不能选择活动 | 提示用户先取消优惠券 |
| "积分抵扣后最终应收为0，不可选择满减活动" | 积分抵扣后没有应收金额 | 提示用户无法使用活动 |
| "订单已部分支付，不可选择活动" | 订单已部分支付 | 提示用户无法修改活动 |

---

## 🔄 前端对接建议

### 1. 活动列表展示

```typescript
// 获取结账页面信息时，会返回活动列表
const response = await getPaymentInfo(saleBillUuid, saleOrderUuid);
const activityList = response.data.activity_list.list;

// 根据 is_available 显示/置灰活动
activityList.forEach(activity => {
  if (activity.is_available) {
    // 显示为可选状态
  } else {
    // 显示为不可选状态（置灰）
  }
});

// 根据 is_selected 显示选中状态
activityList.forEach(activity => {
  if (activity.is_selected) {
    // 显示为已选中状态
    // 显示 discount_amount 抵扣金额
  }
});
```

### 2. 选择活动

```typescript
// 选择活动
const selectActivity = async (activityUuid: number) => {
  try {
    const response = await selectActivity({
      sale_bill_uuid: saleBillUuid,
      sale_order_uuid: saleOrderUuid,
      full_reduction_activity_uuid: activityUuid
    });
    
    // 更新结账页面信息
    updatePaymentInfo(response.data);
  } catch (error) {
    // 处理错误，显示错误提示
    showError(error.message);
  }
};
```

### 3. 取消活动

```typescript
// 取消活动
const cancelActivity = async () => {
  try {
    const response = await selectActivity({
      sale_bill_uuid: saleBillUuid,
      sale_order_uuid: saleOrderUuid,
      full_reduction_activity_uuid: 0  // 传入0表示取消
    });
    
    // 更新结账页面信息
    updatePaymentInfo(response.data);
  } catch (error) {
    showError(error.message);
  }
};
```

### 4. 活动规则展示

```typescript
// 展示活动规则
const formatActivityRule = (activity: FullReductionActivityItem): string => {
  if (activity.activity_type === 1) {
    // 阶梯满减：显示所有规则
    return activity.rules.map(rule => 
      `满${rule.threshold}减${rule.discount}`
    ).join('，');
  } else if (activity.activity_type === 2) {
    // 循环满减：显示循环规则
    const rule = activity.rules[0];
    return `每满${rule.threshold}减${rule.discount}`;
  }
  return '';
};
```

### 5. 金额计算

活动抵扣金额已由后端计算好，前端无需重新计算：

- 如果活动已选中（`is_selected = true`），`discount_amount` 字段包含实际抵扣金额
- 活动抵扣金额已包含在 `amounts` 中的订单金额计算中
- **`amounts.list` 中每个 `PaymentMethodAmount` 对象的 `activity_amount` 字段包含满减活动抵扣金额**
- 前端可以同时显示：
  - `activity_list` 中选中活动的 `discount_amount`（用于活动列表展示）
  - `amounts.list` 中每个支付方式的 `activity_amount`（用于金额明细展示）

---

## 📝 注意事项

1. **响应格式**: `activity_list` 是一个对象，包含 `list` 数组，不是直接的数组
   ```json
   // ✅ 正确
   "activity_list": {
     "list": [...]
   }
   
   // ❌ 错误
   "activity_list": [...]
   ```

2. **多语言字段**: `locale_name` 使用 `LocaleResponse` 结构，包含所有支持的语言
   - 前端应根据当前语言环境显示对应的名称
   - 如果当前语言不存在，可回退到中文（`zh`）

3. **活动可用性**: `is_available` 由后端判断，前端无需重复判断
   - 如果 `is_available = false`，应置灰显示，不可选择
   - 如果 `is_available = true`，显示为可选状态

4. **活动选中状态**: `is_selected` 表示活动是否已选中
   - 如果 `is_selected = true`，应显示选中状态和 `discount_amount` 抵扣金额
   - 如果 `is_selected = false`，显示为未选中状态

5. **活动时段**: 活动可能设置了时段限制（`is_all_day = false`）
   - 前端可以显示活动的适用时段（`start_time` - `end_time`）
   - 但可用性判断由后端完成，前端无需判断

6. **订单金额更新**: 选择或取消活动后，订单金额会自动更新
   - 前端应刷新结账页面信息，获取最新的金额数据
   - `amounts` 中的金额已包含活动抵扣金额
   - **`amounts.list` 中每个 `PaymentMethodAmount` 对象的 `activity_amount` 字段会显示满减活动抵扣金额**

7. **互斥规则**: 活动与优惠券、积分抵扣互斥
   - 选择活动时，如果已使用优惠券，会返回错误
   - 选择活动后，积分不再自动抵扣
   - 前端应提示用户这些互斥规则

8. **部分支付**: 如果订单已部分支付，则不可选择或修改活动
   - 前端应在选择活动前检查订单支付状态
   - 如果已部分支付，应禁用活动选择功能

---

## 🔗 相关接口

### 收银端

- `GET /api/v1/cashier/desk/order/payment/info` - 获取结账页面信息（已修改）
- `POST /api/v1/cashier/desk/order/payment/activity` - 选择或取消满减活动（新增）
- `POST /api/v1/cashier/desk/order/payment/finish` - 完成结账（已修改，增加活动核销逻辑）

### 助手端

- `GET /api/v1/assistant/desk/order/payment/info` - 获取结账页面信息（已修改）
- `POST /api/v1/assistant/desk/order/payment/activity` - 选择或取消满减活动（新增）
- `POST /api/v1/assistant/desk/order/payment/finish` - 完成结账（已修改，增加活动核销逻辑）

---

## 📞 联系方式

如有疑问，请联系后端开发团队。

---

**文档版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

