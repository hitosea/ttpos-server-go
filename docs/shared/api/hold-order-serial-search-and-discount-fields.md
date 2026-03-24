# API 变更文档：挂单列表流水号搜索 & 会员订单折扣字段

> 分支: `feature/order-first-submit`
> 日期: 2026-03-23

---

## 一、挂单列表 — 新增流水号搜索

### 接口信息

| 项目 | 值 |
|------|-----|
| URL | `GET /api/v1/cashier/instant/order/list` |
| 认证 | `Authorization: Bearer {cashier_token}` |

### 请求参数 (Query)

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page_no | int | 是 | 页码 |
| page_size | int | 是 | 每页条数 |
| **keyword** | string | 否 | **[新增]** 搜索关键词，按流水号模糊匹配 |

### 说明

- `serial_no` 字段原已存在于响应中，本次无响应结构变更
- `keyword` 参数为空或不传时，行为与之前一致（返回所有挂单）
- 搜索为模糊匹配（LIKE），支持部分流水号搜索

---

## 二、会员端堂食订单详情 — 折扣字段

### 接口信息

| 项目 | 值 |
|------|-----|
| URL | `GET /api/v1/member/order/dine_in/detail` |
| 认证 | `Authorization: Bearer {member_token}` |

### `amount_info` 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `discount_amount` | float64 | 整单折扣金额（与收银机购物车 `discount_amount` 一致，仅 `custom_discount_fee`） |
| `service_fee` | float64 | 服务费 |
| `tax_fee` | float64 | 税费 |
| `amount` | float64 | 应付金额（已扣退款） |
| `payment_method_name` | string | 支付方式名称 |

> `discount_amount` 仅包含整单折扣（custom_discount_fee），与收银机购物车口径一致。
