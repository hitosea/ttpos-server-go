# SaleBill 结构体 JSON 文档

本文档详细说明了 `SaleBill` 结构体的所有字段及其关联对象的 JSON 格式。

## 完整 JSON 示例

```json
{
  "uuid": 1234567890123456789,
  "create_time": 1704067200,
  "update_time": 1704067200,
  "delete_time": 0,
  "order_no": "SB20240101001",
  "duty_no": "D20240101001",
  "serial_no": "S20240101001",
  "status": 1,
  "bill_type": 0,
  "dining_method": 0,
  "order_source_uuid": 0,
  "nationality_uuid": 0,
  "is_buffet": 0,
  "buffet_duration": 0,
  "buffet_start_time": 0,
  "delay_duration": 0,
  "delay_start_time": 0,
  "non_ordering_time": 0,
  "reminder_order_time": 0,
  "meal_num": 2,
  "remark": "开台备注",
  "reason": "",
  "amount": 150.00,
  "origin_amount": 180.00,
  "product_amount": 120.00,
  "product_original_amount": 150.00,
  "payment_amount": 150.00,
  "payment_commission_fee": 0.00,
  "service_fee": 10.00,
  "tax_fee": 5.00,
  "discount_fee": 20.00,
  "member_discount_fee": 10.00,
  "gift_amount": 0.00,
  "free_amount": 0.00,
  "lock_time": 0,
  "finish_time": 1704067800,
  "hide_bill_time": 0,
  "production_time": 1704067300,
  "cashier_name": "张三",
  "consumer_uuid": 9876543210987654321,
  "cashier_uuid": 1111111111111111111,
  "desk_uuid": 2222222222222222222,
  "device_uuid": 3333333333333333333,
  "auto_add_must_product": 1,
  "is_kitchen_confirm": 0,
  "sale_orders": [
    {
      "uuid": 4444444444444444444,
      "create_time": 1704067200,
      "update_time": 1704067200,
      "delete_time": 0,
      "order_no": "SO20240101001",
      "status": 1,
      "is_free": 0,
      "free_reason": "",
      "cashier_name": "张三",
      "device_id": "DEV001",
      "consumer_uuid": 9876543210987654321,
      "cashier_uuid": 1111111111111111111,
      "sale_bill_uuid": 1234567890123456789,
      "staff_shift_log_uuid": 5555555555555555555,
      "product_amount": 120.00,
      "product_original_amount": 150.00,
      "service_fee": 10.00,
      "tax_fee": 5.00,
      "custom_discount_fee": 20.00,
      "member_discount_fee": 10.00,
      "origin_amount": 180.00,
      "amount": 150.00,
      "custom_amount": -1,
      "finish_time": 1704067800,
      "member_discount_rate": 0.90,
      "member_card_discount_rate": 1.00,
      "custom_discount_rate": 0.90,
      "zero_rule": 0,
      "zero_fee": 0.00,
      "zero_checkout_rule": 0,
      "pay_points": 0.00,
      "pay_points_amount": 0.00,
      "points_exchange_rate": 0.00,
      "auto_points_exchange": 0,
      "coupon_amount": 0.00,
      "payment_amount": 150.00,
      "change_amount": 0.00,
      "zero_checkout_fee": 0.00,
      "final_price": 150.00,
      "payment_commission_fee": 0.00,
      "gift_amount": 0.00,
      "gift_points": 0.00,
      "gift_points_rate": 0.00,
      "gift_points_type": 0,
      "member_level_name": "",
      "member_balance": 0.00,
      "unit": "￥",
      "erp_products_invoice_name": "",
      "erp_material_invoice_name": "",
      "erp_discount_amount": 0.00,
      "sale_order_products": [
        {
          "uuid": 8888888888888888888,
          "create_time": 1704067200,
          "update_time": 1704067200,
          "delete_time": 0,
          "name": "宫保鸡丁",
          "flavor_name": "中辣",
          "num": 1.0,
          "unit_num": 0.0,
          "num_type": 0,
          "remark": "少放花生",
          "is_buffet": 0,
          "device_id": "DEV001",
          "status": 1,
          "is_require": 0,
          "is_accept_order": 1,
          "flavor_price": 50.00,
          "sauce_price": 0.00,
          "product_price": 50.00,
          "sale_price": 50.00,
          "sale_price_no_tax": 50.00,
          "price": 45.00,
          "total_price": 55.00,
          "origin_total_price": 60.00,
          "change_price_time": 0,
          "open_member_discount": 1,
          "open_overall_discount": 1,
          "member_discount_rate": 0.90,
          "member_card_discount_rate": 1.00,
          "member_order_discount_rate": 1.00,
          "custom_discount_rate": 0.90,
          "discount_fee": 5.00,
          "member_discount_fee": 5.00,
          "custom_discount_fee": 0.00,
          "tax_rate": 7.00,
          "service_tax_fee": 0.00,
          "tax_fee": 3.50,
          "service_fee": 5.00,
          "deduct_stock_type": 0,
          "deduct_stock_time": 0,
          "gift_time": 0,
          "wrap_time": 0,
          "cancel_time": 0,
          "gift_reason": "",
          "cancel_reason": "",
          "multi_language_name_uuid": 9999999999999999999,
          "image_file_uuid": 1010101010101010101,
          "production_order_uuid": 0,
          "product_package_uuid": 1212121212121212121,
          "sale_bill_uuid": 1234567890123456789,
          "sale_order_uuid": 4444444444444444444,
          "must_plan_uuid": 0,
          "desk_uuid": 2222222222222222222,
          "sign": "sign123456",
          "h5_order_product_uuid": 0,
          "h5_order_uuid": 0,
          "package_uuid": 0,
          "package_group_uuid": 0,
          "product_type": 0,
          "package_sub_product_params": "",
          "send_kitchen_time": 1704067300,
          "erp_code": "ERP001",
          "batch_tag_uuid": 0,
          "batch_time": 0,
          "is_batch": 0,
          "sale_order_product_boms": [
            {
              "uuid": 1313131313131313131,
              "create_time": 1704067200,
              "update_time": 1704067200,
              "delete_time": 0,
              "name": "中辣",
              "price": 0.00,
              "is_flavor_bom": 1,
              "sale_order_uuid": 4444444444444444444,
              "sale_order_product_uuid": 8888888888888888888,
              "product_bom_uuid": 1414141414141414141
            },
            {
              "uuid": 1515151515151515151,
              "create_time": 1704067200,
              "update_time": 1704067200,
              "delete_time": 0,
              "name": "加花生",
              "price": 5.00,
              "is_flavor_bom": 0,
              "sale_order_uuid": 4444444444444444444,
              "sale_order_product_uuid": 8888888888888888888,
              "product_bom_uuid": 1616161616161616161
            }
          ]
        }
      ]
    }
  ],
  "sale_bill_setting": {
    "uuid": 6666666666666666666,
    "create_time": 1704067200,
    "update_time": 1704067200,
    "delete_time": 0,
    "sale_bill_uuid": 1234567890123456789,
    "service_fee_type": 2,
    "service_fee_value": 10.00,
    "tax_fee_type": 1,
    "service_apply": 1,
    "service_fee_base": 0,
    "discount_type": 0,
    "zero_rule": 0,
    "zero_checkout_rule": 0,
    "is_stat_gift": 0,
    "is_stat_free": 0,
    "open_points_exchange": 0,
    "points_exchange_rate": 0.00,
    "auto_points_exchange": 0
  },
  "desk": {
    "uuid": 2222222222222222222,
    "desk_no": "A01",
    "region_uuid": 8888888888888888888,
    "type_uuid": 9999999999999999999,
  }
}
```

## 字段说明

### BaseModel 基础字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `id` | uint | 自增ID | 数据库自动生成，从1开始递增 |
| `uuid` | uint64 | 销售账单唯一标识 | 雪花算法生成的64位整数，如：1234567890123456789 |
| `create_time` | int64 | 创建时间（时间戳） | Unix时间戳，如：1704067200（2024-01-01 00:00:00） |
| `update_time` | int64 | 更新时间（时间戳） | Unix时间戳，每次更新时自动更新 |
| `delete_time` | int64 | 删除时间（时间戳） | 0表示未删除，非0表示已删除（软删除） |

### 主键和标识字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `order_no` | string | 销售账单编号 | 业务订单编号，格式如："SB20240101001" |
| `duty_no` | string | 当班编号 | 用于标记该账单属于哪个当班，格式如："D20240101001" |
| `serial_no` | string | 桌位编号（点餐流水号） | 桌台订单的流水号，格式如："S20240101001" |

### 状态相关字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `status` | uint | 订单状态 | 0-待付款、1-已完成、2-已取消 |
| `is_lock` | uint | 是否锁单 | 0-否、1-是 |
| `is_split_order` | uint | 是否拆单 | 0-否、1-是 |

### 订单类型字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `bill_type` | uint | 账单类型 | 0-桌台订单、1-点餐订单(包括会员端的堂食订单)、2-会员端订单(仅外送) |
| `dining_method` | uint | 用餐方式 | 0-堂食、1-打包 |
| `order_source_uuid` | uint64 | 订单来源UUID | 0=店内，>0=外卖（关联OrderSource表） |
| `nationality_uuid` | uint64 | 国籍UUID | 0=未记录，>0=关联Nationality表 |
| `is_buffet` | uint | 是否自助餐 | 0-否、1-是 |
| `buffet_duration` | uint | 自助餐可用时长（秒） | 0为不限时，>0为限时时长 |
| `buffet_start_time` | int64 | 自助餐开始时间（秒） | Unix时间戳 |
| `delay_duration` | uint | 总延迟时长（秒） | 自助餐加钟的总时长 |
| `delay_start_time` | int64 | 总延迟时长开始时间（秒） | Unix时间戳 |
| `non_ordering_time` | uint | 不可下单时间（分钟） | 自助餐不可下单的时间段 |
| `reminder_order_time` | uint | 提醒下单时间（分钟） | 自助餐提醒下单的时间点 |

### 订单基本信息

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `meal_num` | uint | 就餐人数 | 正整数，如：2、4、6 |
| `remark` | string | 备注（开台备注） | 字符串，如："开台备注" |
| `order_remark` | string | 整单备注JSON | JSON字符串，格式：`{"list":[{"remark":"备注内容","is_latest":true}]}` |
| `reason` | string | 原因 | 取消订单的原因说明 |

### 金额字段 - 主要金额

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `amount` | float64 | 订单总金额 | 关联销售订单的总金额之和，如：150.00 |
| `origin_amount` | float64 | 订单金额（折前价） | 商品未含税时=商品金额+服务费+税费；商品已含税时=商品金额（含税）+服务费+税费，如：180.00 |
| `product_amount` | float64 | 商品金额 | 关联销售订单的商品金额之和，如：120.00 |
| `product_original_amount` | float64 | 原始商品金额 | 商品原始金额=（订单.原始商品金额）之和，如：150.00 |

### 金额字段 - 支付相关

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `payment_amount` | float64 | 支付金额 | 支付金额-订单总金额=支付手续费，如：150.00 |
| `payment_commission_fee` | float64 | 支付手续费 | 多次支付的支付手续费之和，如：0.00 |

### 金额字段 - 费用相关

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `service_fee` | float64 | 服务费 | 关联销售订单的服务费之和，如：10.00 |
| `tax_fee` | float64 | 税费 | 关联销售订单的税费之和，如：5.00 |

### 金额字段 - 优惠相关

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `discount_fee` | float64 | 折扣费用 | 关联销售订单的折扣费用之和，如：20.00 |
| `member_discount_fee` | float64 | 会员折扣费用 | 关联销售订单的会员折扣费用之和，如：10.00 |
| `gift_amount` | float64 | 赠菜金额 | 关联销售订单的赠菜金额之和，如：0.00 |
| `free_amount` | float64 | 免单金额 | 关联销售订单的免单金额之和，如：0.00 |

### 时间相关字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `lock_time` | int64 | 锁单时间 | Unix时间戳，0表示未锁单 |
| `finish_time` | int64 | 完成时间（时间戳） | Unix时间戳，0表示未完成 |
| `hide_bill_time` | int64 | 隐藏账单时间（时间戳） | Unix时间戳，0表示未隐藏 |
| `production_time` | int64 | 首次送厨时间（时间戳） | Unix时间戳，0表示未送厨 |

### 收银员信息

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `cashier_name` | string | 收银员名称 | 字符串，如："张三" |

### 关联ID字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `consumer_uuid` | uint64 | 消费者ID | 关联会员表，0表示非会员订单 |
| `cashier_uuid` | uint64 | 收银员ID | 关联员工表，系统自动创建的账单为0 |
| `desk_uuid` | uint64 | 餐桌ID | 关联桌台表，0表示点餐订单 |
| `buffet_package1_uuid` | uint64 | 自助餐套餐1ID | 关联自助餐套餐表，0表示未选择 |
| `buffet_package2_uuid` | uint64 | 自助餐套餐2ID | 关联自助餐套餐表，0表示未选择 |
| `device_uuid` | uint64 | 设备ID | 标识账单创建设备，点餐账单通过设备uuid查询 |
| `member_sale_order_uuid` | uint64 | 会员销售订单ID | 关联会员端订单，0表示非会员端订单 |
| `batch_tag_uuid` | uint64 | 分批类型UUID | 关联分批类型表，0表示未设置 |

### 必点方案相关字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `show_must_plan` | uint | 是否显示必点方案 | 0-不显示、1-显示 |
| `auto_add_must_product` | uint | 是否自动加购必点商品 | 0-不自动加购、1-自动加购 |

### 其他字段

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `is_kitchen_confirm` | uint | 厨显端是否确认退菜整单 | 0-否、1-是（2.4.0版本新增） |
| `reverse_settle_count` | uint | 反结账次数 | 正整数，0表示未反结账（2.5.0版本新增） |

## 关联对象说明

### SaleOrders（销售订单列表）

`SaleOrders` 是一个数组，包含该账单下的所有销售订单。每个 `SaleOrder` 对象包含：

- **基础标识字段**：订单编号、状态、是否免单等
- **金额字段**：商品金额、服务费、税费、折扣等
- **支付信息**：支付金额、找零金额、优惠券金额等
- **关联对象**：支付订单列表、会员信息、订单商品列表等
- **sale_order_products**：销售订单商品列表（数组）

详细字段说明请参考 `SaleOrder` 结构体文档。

#### SaleOrderProducts（销售订单商品列表）

`sale_order_products` 是 `SaleOrder` 对象中的一个数组字段，包含该订单下的所有商品。每个 `SaleOrderProduct` 对象包含：

| 字段类别 | 主要字段 | 说明 |
|---------|---------|------|
| **基础信息** | `name`, `flavor_name`, `num`, `remark` | 商品名称、规格名称、数量、备注 |
| **价格信息** | `flavor_price`, `sauce_price`, `product_price`, `sale_price`, `price`, `total_price` | 规格原价、小料价、原始单价、销售价、最终单价、应收金额 |
| **折扣信息** | `member_discount_rate`, `custom_discount_rate`, `discount_fee` | 会员折扣率、自定义折扣率、折扣金额 |
| **税费和服务费** | `tax_rate`, `tax_fee`, `service_fee` | 税率、商品税费、服务费 |
| **状态信息** | `status`, `is_require`, `is_accept_order` | 送厨状态、是否必点、是否已接单 |
| **关联ID** | `product_package_uuid`, `sale_order_uuid`, `sale_bill_uuid` | 商品包ID、销售订单ID、销售账单ID |
| **套餐相关** | `package_uuid`, `package_group_uuid`, `product_type` | 套餐UUID、套餐分组UUID、商品类型 |
| **分批相关** | `is_batch`, `batch_tag_uuid`, `batch_time` | 是否分批商品、分批类型UUID、分批时间 |
| **时间信息** | `send_kitchen_time`, `gift_time`, `cancel_time` | 送厨时间、赠菜时间、退菜时间 |

**主要字段取值参考：**

- `status`: 0-未送厨、1-已送厨
- `is_require`: 0-否、1-是（是否必点商品）
- `is_accept_order`: 0-否、1-是（是否已接单）
- `product_type`: 0-商品、1-套餐、2-套餐子商品
- `is_batch`: 0-否、1-是（是否是分批商品）
- `num_type`: 0-整数、1-小数（数量类型）

#### SaleOrderProductBoms（销售订单商品BOM列表）

`sale_order_product_boms` 是 `SaleOrderProduct` 对象中的一个数组字段，包含该商品的规格和小料信息。每个 `SaleOrderProductBom` 对象包含：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `uuid` | uint64 | BOM唯一标识 | 雪花算法生成的64位整数 |
| `create_time` | int64 | 创建时间（时间戳） | Unix时间戳 |
| `update_time` | int64 | 更新时间（时间戳） | Unix时间戳 |
| `delete_time` | int64 | 删除时间（时间戳） | 0表示未删除，非0表示已删除 |
| `name` | string | 规格或小料规格名称 | 不随后台更新，如："中辣"、"加花生" |
| `price` | float64 | 单价 | 不随后台更新，记录加购时的价格，如：5.00 |
| `is_flavor_bom` | uint | 是否为规格商品BOM | 0-否（加料商品）、1-是（规格商品） |
| `sale_order_uuid` | uint64 | 销售订单ID | 关联销售订单 |
| `sale_order_product_uuid` | uint64 | 销售订单商品ID | 关联销售订单商品 |
| `product_bom_uuid` | uint64 | 商品BOM ID | 关联商品BOM表 |

**说明：**
- `is_flavor_bom = 1` 表示这是规格（如：中辣、大辣）
- `is_flavor_bom = 0` 表示这是小料/加料（如：加花生、加香菜）
- 一个商品可以有多个规格和多个小料

### H5OrderProducts（H5订单商品列表）

`H5OrderProducts` 是一个数组，包含扫码点餐的商品列表。每个 `H5OrderProduct` 对象包含：

- **快照信息**：商品名称、价格、数量、属性文本等（接单后不再改变）
- **关联UUID**：销售订单商品UUID、H5订单UUID、销售账单UUID

### SaleBillSetting（销售账单设置）

`sale_bill_setting` 对象包含该账单的费用计算和优惠设置：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `service_fee_type` | uint | 服务费类型 | 0-免服务费、1-按固定金额、2-按比例-不收取税费、3-按比例-收取税费 |
| `service_fee_value` | float64 | 服务费值 | 类型为1时是固定金额，类型为2和3时是%比例 |
| `tax_fee_type` | uint | 税费类型 | 0-关闭消费税、1-商品未含税、2-商品已含税 |
| `service_apply` | uint | 是否收取服务费 | 0-不收取、1-收取 |
| `service_fee_base` | uint | 服务费计算基准 | 0-商品惠后价、1-商品价格合计 |
| `discount_type` | uint | 打折类型 | 0-百分比打折%、1-百分比直接减免% off |
| `zero_rule` | uint | 优惠折扣抹零 | 0-实款实收、1-抹分、2-抹角、3-四舍五入保留一位小数、4-四舍五入保留整数 |
| `zero_checkout_rule` | uint | 结账抹零 | 0-实款实收、1-抹分、2-抹角、5-抹元 |
| `is_stat_gift` | uint | 是否统计赠菜金额 | 0-不计入总销售额、优惠折扣、1-计入总销售额、优惠折扣 |
| `is_stat_free` | uint | 是否统计免单金额 | 0-不计入总销售额、优惠折扣、服务费、税费、1-计入 |
| `open_points_exchange` | uint | 是否开启积分抵扣 | 0-不开启、1-开启 |
| `points_exchange_rate` | float64 | 积分汇率 | 每积分抵扣的金额，如：0.01 |
| `auto_points_exchange` | uint | 积分抵扣类型 | 0-手动抵扣、1-自动抵扣 |

### Cashier（收银员）

`cashier` 对象包含收银员的基本信息：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `company_uuid` | uint64 | 集团ID | 关联公司表 |
| `username` | string | 用户名 | 登录用户名 |
| `phone` | string | 手机号 | 如："13800138000" |
| `real_name` | string | 姓名 | 如："张三" |
| `is_super` | int | 是否为超级管理员 | 0-不是、1-是 |
| `user_type` | int | 账号类型 | 0-总台、1-门店 |
| `is_disable` | int | 是否禁用 | 0-未禁用、1-禁用 |
| `cashier_online` | int | 收银员当班 | 0-不在线、1-在线 |
| `cashier_login_time` | int64 | 收银员当班登录时间 | Unix时间戳 |
| `duty_no` | string | 当班编号 | 如："D20240101001" |

### Desk（桌台）

`desk` 对象包含桌台的基本信息：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `desk_no` | string | 桌位编号 | 如："A01" |
| `region_uuid` | uint64 | 桌台区域ID | 关联桌台区域表 |
| `type_uuid` | uint64 | 桌台类型ID | 关联桌台类型表 |
| `sort` | uint | 排序序号 | 正整数 |
| `status` | uint | 状态 | 0-未开台、1-已开台 |
| `is_disable` | uint | 是否禁用 | 0-否、1-是 |
| `qrcode_token` | string | 二维码token | 用于判断二维码链接是否有效 |
| `sale_bill_uuid` | uint64 | 销售账单UUID | 当前绑定的销售账单 |
| `device_uuid` | uint64 | 平板设备uuid | 0-未绑定 |
| `default_people_num` | uint | 默认人数 | 正整数 |
| `is_open_default_people_num` | uint | 是否开启默认人数 | 0-否、1-是 |

### BuffetPackage1 / BuffetPackage2（自助餐套餐）

`buffet_package1` 和 `buffet_package2` 对象包含自助餐套餐信息（可为 null）：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `name` | string | 自助餐名称 | 字符串 |
| `multi_language_name_uuid` | uint64 | 多语言名称ID | 关联多语言名称表 |
| `sort` | uint | 排序顺序 | 正整数 |
| `tax_uuid` | uint64 | 税率ID | 关联税率表 |
| `is_limit_time` | uint | 是否限时 | 0-否、1-是 |
| `limit_time` | uint | 限时时间 | 分钟数 |
| `can_combined` | uint | 是否可组合 | 0-否、1-是 |
| `non_ordering_time` | uint | 不可下单时间（分钟） | 正整数 |
| `reminder_order_time` | uint | 提醒下单时间（分钟） | 正整数 |
| `actual_sale_num` | float64 | 实际销量 | 浮点数 |
| `open_overall_discount` | uint | 是否开启整单折扣 | 0-否、1-是 |

### BatchTag（分批类型）

`batch_tag` 对象包含分批类型信息（可为 null）：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `name` | string | 名称 | 字符串 |
| `multi_language_name_uuid` | uint64 | 多语言名称UUID | 关联多语言名称表 |
| `abbreviation` | string | 名称缩写 | 字符串 |
| `color` | string | 颜色值 | 如："#FF0000" |
| `sort` | int | 排序 | 数字越小越靠前 |

### OrderSource（订单来源）

`order_source` 对象包含订单来源信息（可为 null）：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `multi_language_name_uuid` | uint64 | 多语言名称UUID | 关联多语言名称表 |
| `sort` | int | 排序 | 正整数 |
| `status` | int | 状态 | 1-启用、0-禁用 |

### Nationality（国籍）

`nationality` 对象包含国籍信息（可为 null）：

| 字段名 | 类型 | 说明 | 取值参考 |
|--------|------|------|----------|
| `multi_language_name_uuid` | uint64 | 多语言名称UUID | 关联多语言名称表 |
| `sort` | int | 排序 | 正整数 |
| `status` | int | 状态 | 1-启用、0-禁用 |

## 注意事项

1. **时间戳字段**：所有时间字段均为 Unix 时间戳（秒级），需要转换为日期时间格式时使用相应的时间函数。

2. **金额字段**：所有金额字段均为 `float64` 类型，保留两位小数，单位为元。

3. **UUID 字段**：所有 UUID 字段均为 `uint64` 类型，使用雪花算法生成。

4. **状态字段**：大部分状态字段使用 `uint` 类型，0 通常表示"否"或"未启用"，1 表示"是"或"已启用"。

5. **关联对象**：关联对象可能为 `null`，需要在使用前进行空值检查。

6. **软删除**：`delete_time` 为 0 表示未删除，非 0 表示已删除（软删除）。

7. **JSON 字段**：`order_remark` 字段存储的是 JSON 字符串，需要解析后才能使用。

## 版本说明

- `is_kitchen_confirm`：2.4.0 版本新增
- `reverse_settle_count`：2.5.0 版本新增

