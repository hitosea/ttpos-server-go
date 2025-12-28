# 外卖系统数据库架构说明

> **版本**: v1.0  
> **最后更新**: 2025-12-29  
> **维护者**: TTPOS Team

---

## 📋 目录

- [1. 概述](#1-概述)
- [2. 核心表结构](#2-核心表结构)
  - [2.1 平台管理表](#21-平台管理表)
  - [2.2 菜单数据表](#22-菜单数据表)
  - [2.3 订单主表](#23-订单主表)
  - [2.4 订单商品表](#24-订单商品表)
  - [2.5 订单修饰符表](#25-订单修饰符表)
  - [2.6 订单收货人表](#26-订单收货人表)
  - [2.7 订单活动表](#27-订单活动表)
  - [2.8 订单促销表](#28-订单促销表)
  - [2.9 订单原料表](#29-订单原料表)
  - [2.10 平台设置表](#210-平台设置表)
  - [2.11 导入日志表](#211-导入日志表)
- [3. 表关系图](#3-表关系图)
- [4. 数据流程](#4-数据流程)
- [5. 关键字段说明](#5-关键字段说明)
- [6. 索引策略](#6-索引策略)
- [7. 最佳实践](#7-最佳实践)

---

## 1. 概述

### 1.1 系统定位

TTPOS 外卖系统是一个多平台外卖订单管理系统，支持：
- **Grab**（东南亚主流平台）
- **Lineman**（泰国本地平台）
- **FoodPanda**（国际平台）
- 其他第三方平台（可扩展）

### 1.2 架构特点

- ✅ **多平台支持**：统一数据模型适配多个外卖平台
- ✅ **数据冗余设计**：订单创建时保存关键名称，减少查询依赖
- ✅ **分离关注点**：平台数据与 TTPOS 核心数据分别存储
- ✅ **可追溯性**：保存原始订单数据和映射关系
- ✅ **高性能**：合理的索引和查询优化

### 1.3 表分类

| 分类 | 表名 | 用途 |
|------|------|------|
| **平台管理** | `ttpos_takeout` | 平台开关和菜单管理 |
| **菜单数据** | `ttpos_product_package_takeout`<br>`ttpos_product_bom_takeout` | 外卖商品和规格价格 |
| **订单核心** | `ttpos_takeout_order`<br>`ttpos_takeout_order_item`<br>`ttpos_takeout_order_item_modifier` | 订单主表、商品、修饰符 |
| **订单扩展** | `ttpos_takeout_order_receiver`<br>`ttpos_takeout_order_campaign`<br>`ttpos_takeout_order_promo` | 收货人、活动、促销 |
| **业务支持** | `ttpos_takeout_order_material`<br>`ttpos_takeout_settings`<br>`ttpos_takeout_import_log` | 原料扣减、配置、日志 |

---

## 2. 核心表结构

### 2.1 平台管理表

#### `ttpos_takeout` - 外卖平台状态管理表

**表名**: `ttpos_takeout`  
**用途**: 管理外卖平台的开关状态、菜单数据和绑定状态

**字段说明**:

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| `uuid` | bigint | 唯一标识 | 主键 |
| `platform` | varchar(50) | 外卖平台 | grab/lineman/foodpanda |
| `enabled` | int | 是否开启 | 1=开启, 0=关闭 |
| `import_status` | int | 导入状态 | 0=未导入, 1=导入中, 2=成功, 3=失败 |
| `menu` | json | 平台菜单数据 | 从平台获取的原始菜单 |
| `ttpos_menu` | json | TTPOS 导出的菜单 | 推送到平台的菜单数据 |
| `is_bound` | int | 是否已绑定平台 | 1=已绑定, 0=未绑定 |
| `skip` | int | 是否跳过绑定 | 1=跳过, 0=不跳过 |
| `binding_link` | varchar(500) | 平台绑定链接 | 缓存用 |

**索引**:
- `UNIQUE KEY uk_platform (platform, delete_time)` - 平台唯一性
- `KEY idx_platform (platform)` - 平台查询
- `KEY idx_enabled (enabled)` - 开启状态查询

**业务规则**:
1. 每个平台只能有一条记录（排除软删除）
2. 关闭平台不影响已有订单
3. 菜单数据以 JSON 格式存储，支持动态结构

---

### 2.2 菜单数据表

#### `ttpos_product_package_takeout` - 外卖商品表

**表名**: `ttpos_product_package_takeout`  
**用途**: 存储商品的外卖专属信息（价格、描述、图片等）

**核心字段**:

| 字段名 | 类型 | 说明 | 关联 |
|--------|------|------|------|
| `product_package_uuid` | bigint | 商品包UUID | → `ttpos_product_package.uuid` |
| `name` | text | 商品名称 | 外卖专用名称 |
| `multi_language_name_uuid` | bigint | 多语言名称ID | → `ttpos_multi_language_name.uuid` |
| `product_type` | int | 商品类型 | 0=商品, 1=套餐 |
| `takeout_type` | int | 外卖类型 | 1=Grab, 2=FoodPanda, 3=其他 |
| `status` | int | 外卖状态 | 0=下架, 1=上架 |
| `price` | decimal(22,4) | 外卖价格 | 可能与店内价格不同 |
| `category_uuid` | bigint | 外卖分类UUID | 外卖专用分类 |
| `source` | varchar(50) | 来源平台 | grab/foodpanda/lineman |
| `source_product_id` | varchar(500) | 平台商品ID | 平台唯一标识 |

**关键特性**:
- ✅ 支持店内商品与外卖商品价格分离
- ✅ 支持多语言名称和描述
- ✅ 与平台商品ID双向映射
- ✅ 支持外卖专用分类体系

#### `ttpos_product_bom_takeout` - 外卖规格价格表

**表名**: `ttpos_product_bom_takeout`  
**用途**: 存储规格/加料在外卖平台的专属价格

**核心字段**:

| 字段名 | 类型 | 说明 | 关联 |
|--------|------|------|------|
| `product_package_takeout_uuid` | bigint | 外卖商品UUID | → `ttpos_product_package_takeout.uuid` |
| `product_bom_uuid` | bigint | BOM UUID | → `ttpos_product_bom.uuid` |
| `grab_modifier_id` | varchar(500) | Grab修饰符ID | 平台专用ID |
| `price` | decimal(22,4) | 外卖规格价格 | 可能与店内不同 |

---

### 2.3 订单主表

#### `ttpos_takeout_order` - 外卖订单表

**表名**: `ttpos_takeout_order`  
**用途**: 存储外卖订单的主要信息

**核心字段分组**:

##### A. 平台标识
| 字段名 | 类型 | 说明 |
|--------|------|------|
| `platform` | varchar(20) | 外卖平台 (grab/foodpanda/lineman) |
| `platform_order_id` | varchar(255) | 平台订单ID |
| `platform_order_state` | varchar(50) | 平台订单状态 |
| `short_order_number` | varchar(50) | 短订单号 (GF-123) |
| `merchant_id` | varchar(100) | 商户ID |

##### B. 订单状态
| 字段名 | 类型 | 说明 | 枚举值 |
|--------|------|------|--------|
| `order_state` | int | 订单状态 | 0=待接单, 1=已接单配餐中, 2=待骑手接单, 3=骑手配送中, 4=已完成, 5=已拒单 |
| `is_abnormal` | int | 是否异常 | 0=正常, 1=异常 |
| `abnormal_detail` | text | 异常详情 | JSON格式 |
| `stock_status` | int | 库存状态 | 1=充足, 2=不足 |

##### C. 价格信息
| 字段名 | 类型 | 说明 | 单位 |
|--------|------|------|------|
| `subtotal` | bigint | 小计金额 | 分（最小货币单位） |
| `delivery_fee` | bigint | 配送费 | 分 |
| `eater_payment` | bigint | 顾客实付 | 分 |
| `platform_discount` | bigint | 平台优惠 | 分 |
| `merchant_discount` | bigint | 商户优惠 | 分 |
| `basket_promo` | bigint | 购物车优惠 | 分 |
| `tax` | bigint | 税费 | 分 |
| `merchant_charge_fee` | bigint | 商户服务费 | 分 |

##### D. 货币信息
| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| `currency_code` | varchar(10) | 货币代码 | THB/VND/USD |
| `currency_symbol` | varchar(10) | 货币符号 | ฿/$/ |
| `currency_exponent` | int | 货币指数 | 2 (即 1元=100分) |

##### E. 时间字段
| 字段名 | 类型 | 说明 |
|--------|------|------|
| `order_time` | int | 下单时间 |
| `submit_time` | int | 提交时间 |
| `scheduled_time` | int | 预定时间 |
| `accepted_time` | int | 接单时间 |
| `completed_time` | int | 完成时间 |
| `rejected_time` | int | 拒单时间 |
| `estimated_ready_time` | int | 预计完成时间 |
| `max_ready_time` | int | 最大完成时间 |

##### F. 订单类型
| 字段名 | 类型 | 说明 | 枚举值 |
|--------|------|------|--------|
| `order_type` | varchar(50) | 订单类型 | TAKEAWAY/DELIVERY/DINEIN |
| `payment_type` | varchar(20) | 支付方式 | CASH/ONLINE |
| `order_accepted_type` | varchar(20) | 接单类型 | AUTO/MANUAL |
| `cutlery` | int | 是否需要餐具 | 0=否, 1=是 |

##### G. 操作记录
| 字段名 | 类型 | 说明 |
|--------|------|------|
| `accepted_by` | bigint | 接单人UUID |
| `rejected_by` | bigint | 拒单人UUID |
| `reject_reason_code` | varchar(50) | 拒单原因代码 |
| `reject_reason` | varchar(255) | 拒单原因 |

##### H. 原始数据
| 字段名 | 类型 | 说明 |
|--------|------|------|
| `raw_data` | mediumtext | 平台原始订单数据 (JSON) |

**索引策略**:
```sql
UNIQUE KEY uk_platform_order (platform, platform_order_id, delete_time)  -- 防止重复订单
KEY idx_platform (platform, delete_time)                                  -- 平台查询
KEY idx_order_state (order_state, delete_time)                            -- 状态筛选
KEY idx_order_time (order_time, delete_time)                              -- 时间范围查询
KEY idx_short_order_number (short_order_number, delete_time)              -- 订单号搜索
```

**业务规则**:
1. **货币处理**: 所有金额以最小货币单位（分）存储，避免浮点数精度问题
2. **时间存储**: Unix 时间戳（秒），方便跨时区处理
3. **原始数据保留**: `raw_data` 保存平台完整数据，用于问题追溯
4. **软删除**: 使用 `delete_time` 实现软删除，不物理删除订单

---

### 2.4 订单商品表

#### `ttpos_takeout_order_item` - 外卖订单商品表

**表名**: `ttpos_takeout_order_item`  
**用途**: 存储订单中的商品信息

**核心字段**:

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| `takeout_order_uuid` | bigint | 外卖订单UUID | → `ttpos_takeout_order.uuid` |
| `platform` | varchar(20) | 外卖平台 | grab/foodpanda/lineman |
| **平台商品信息** ||||
| `platform_item_id` | varchar(100) | 平台商品ID | TTPOS-ITEM-{uuid} |
| `item_name` | text | 商品名称 | 平台名称（优先外卖表） |
| **TTPOS 映射信息** ||||
| `ttpos_product_uuid` | bigint | TTPOS商品UUID | → `ttpos_product_package.uuid` |
| `ttpos_product_type` | int | TTPOS商品类型 | 0=商品, 1=套餐 |
| **✨ 新增：TTPOS 核心表名称** ||||
| `ttpos_item_name` | text | TTPOS商品名称 | **来自 `ttpos_product_package`（核心表）** |
| **订单信息** ||||
| `quantity` | int | 数量 | |
| `price` | decimal(20,4) | 单价 | 元，4位小数 |
| `tax` | decimal(20,4) | 税费 | 元，4位小数 |
| `specifications` | varchar(500) | 规格说明 | |
| `is_mapped` | int | 是否已关联 | 0=异常, 1=正常 |

**关键特性**:
- ✅ **双名称设计**: `item_name`（显示用）+ `ttpos_item_name`（标识用）
- ✅ **平台前缀识别**: `platform_item_id` 以 `TTPOS-ITEM-` 开头表示已映射
- ✅ **类型区分**: 普通商品和套餐分开处理
- ✅ **异常标记**: `is_mapped=0` 标记未映射商品

**索引**:
```sql
UNIQUE KEY uk_uuid (uuid)
KEY idx_takeout_order_uuid (takeout_order_uuid, delete_time)  -- 查询订单商品
KEY idx_platform_item (platform, platform_item_id, delete_time) -- 平台商品查询
```

**关联关系**:
```
ttpos_takeout_order (1) ----< (N) ttpos_takeout_order_item
ttpos_product_package (1) ----< (N) ttpos_takeout_order_item
```

---

### 2.5 订单修饰符表

#### `ttpos_takeout_order_item_modifier` - 外卖订单商品修饰符表

**表名**: `ttpos_takeout_order_item_modifier`  
**用途**: 存储订单商品的规格、加料、属性、套餐子商品等修饰符信息

**核心字段**:

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| `takeout_order_item_uuid` | bigint | 订单商品UUID | → `ttpos_takeout_order_item.uuid` |
| `platform` | varchar(50) | 平台 | grab/foodpanda/lineman |
| **平台修饰符信息** ||||
| `platform_modifier_id` | varchar(255) | 平台修饰符ID | 平台唯一标识 |
| `modifier_name` | text | 修饰符名称 | 平台名称（优先外卖表） |
| **TTPOS 映射信息** ||||
| `ttpos_modifier_uuid` | bigint | TTPOS修饰符UUID | 根据类型关联不同表 |
| `ttpos_modifier_type` | varchar(20) | TTPOS修饰符类型 | **flavor/sauce/attr/commodity** |
| **✨ 新增：TTPOS 核心表名称** ||||
| `ttpos_modifier_name` | text | TTPOS修饰符名称 | **来自核心表（各类型对应的表）** |
| **✨ 新增：商品规格信息（commodity 专用）** ||||
| `ttpos_flavor_uuid` | bigint | TTPOS规格UUID | **对应 `product_bom_uuid`** |
| `ttpos_flavor_name` | text | TTPOS规格名称 | **来自 `ttpos_product_bom`** |
| **订单信息** ||||
| `quantity` | int | 数量 | commodity 类型的数量 |
| `price` | decimal(20,4) | 价格 | 元，4位小数 |
| `tax` | decimal(20,4) | 税费 | 元，4位小数 |
| `is_mapped` | int | 是否已映射 | 0=未映射, 1=已映射 |

**修饰符类型详解**:

| Type | 说明 | ttpos_modifier_uuid 关联 | 数据来源 |
|------|------|--------------------------|----------|
| **flavor** | 规格（大/中/小杯） | → `ttpos_product_bom.uuid` | `ttpos_product_flavor` |
| **sauce** | 加料（珍珠/椰果） | → `ttpos_product_bom.uuid` | `ttpos_product_sauce` |
| **attr** | 属性（冰度/糖度） | → `ttpos_product_package_attribute.uuid` | `ttpos_attribute` |
| **commodity** | 套餐商品 | → `ttpos_product_package_group_item.uuid` | `ttpos_product_package` + `ttpos_product_bom`（规格） |

**commodity 类型特殊处理**:

对于 `commodity` 类型，存储了完整的商品和规格信息：

```
套餐商品: 珍珠奶茶 (大杯) x2
├─ ttpos_modifier_uuid  → product_package_group_item.uuid (套餐项ID)
├─ ttpos_modifier_name  → "珍珠奶茶" (商品名称，来自 ttpos_product_package)
├─ ttpos_flavor_uuid    → product_bom.uuid (规格ID)
├─ ttpos_flavor_name    → "大杯" (规格名称，来自 ttpos_product_bom)
└─ quantity             → 2 (已计算好的数量: groupItem.Num * item.Quantity)
```

**索引**:
```sql
UNIQUE KEY idx_uuid (uuid)
KEY idx_order_item_uuid (takeout_order_item_uuid)              -- 查询商品修饰符
KEY idx_platform_modifier (platform, platform_modifier_id)    -- 平台修饰符查询
KEY idx_delete_time (delete_time)                             -- 软删除过滤
```

**关联关系**:
```
ttpos_takeout_order_item (1) ----< (N) ttpos_takeout_order_item_modifier

修饰符类型关联:
├─ flavor  → ttpos_product_bom.uuid
├─ sauce   → ttpos_product_bom.uuid
├─ attr    → ttpos_product_package_attribute.uuid
└─ commodity → ttpos_product_package_group_item.uuid
             ├─ related_uuid → ttpos_product_package.uuid (商品)
             └─ product_bom_uuid → ttpos_product_bom.uuid (规格)
```

**业务规则**:
1. **名称冗余设计**: 订单创建时保存名称，避免后续查询依赖
2. **商家联/客户联区分**:
   - 商家联打印：使用 `ttpos_xxx_name`（TTPOS 标准名称）
   - 客户联打印：使用 `xxx_name`（平台名称）
3. **数量计算**: commodity 类型的 `quantity` 在创建订单时已计算好（`groupItem.Num * item.Quantity`）
4. **规格完整性**: commodity 类型包含完整的商品名称和规格名称

---

### 2.6 订单收货人表

#### `ttpos_takeout_order_receiver` - 外卖订单收货人信息表

**表名**: `ttpos_takeout_order_receiver`  
**用途**: 存储订单的收货人和配送地址信息

**核心字段**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `takeout_order_uuid` | bigint | 外卖订单UUID（唯一） |
| `platform` | varchar(50) | 平台名称 |
| **收货人信息** |||
| `receiver_name` | varchar(100) | 收货人姓名 |
| `receiver_phones` | varchar(50) | 收货人电话 |
| `unit_number` | varchar(50) | 单元号/门牌号 |
| `delivery_instruction` | varchar(255) | 配送说明 |
| **地址信息** |||
| `address` | varchar(255) | 详细地址 |
| `postcode` | varchar(20) | 邮政编码 |
| `latitude` | decimal(10,7) | 纬度 |
| `longitude` | decimal(10,7) | 经度 |
| **POI 信息** |||
| `poi_source` | varchar(50) | POI来源 (GRAB/GOOGLE/FACEBOOK) |
| `poi_id` | varchar(100) | POI ID |

**索引**:
```sql
UNIQUE KEY uk_takeout_order_uuid (takeout_order_uuid)  -- 一对一关系
```

**关联关系**:
```
ttpos_takeout_order (1) ---- (1) ttpos_takeout_order_receiver
```

---

### 2.7 订单活动表

#### `ttpos_takeout_order_campaign` - 外卖订单活动表

**表名**: `ttpos_takeout_order_campaign`  
**用途**: 存储订单参与的平台活动信息（折扣、赠品等）

**核心字段**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `takeout_order_uuid` | bigint | 外卖订单UUID |
| `campaign_id` | varchar(100) | 活动ID |
| `campaign_name` | varchar(255) | 活动名称 |
| `campaign_name_for_mex` | varchar(255) | 商户提供的活动名称 |
| `campaign_level` | varchar(50) | 活动级别 (order/item) |
| `campaign_type` | varchar(50) | 活动类型 (discount/bundle/free_item) |
| `usage_count` | int | 活动使用次数 |
| `mex_funded_ratio` | int | 商户资金占比(%) |
| `deducted_amount` | bigint | 折扣金额(分) |
| `deducted_part` | varchar(50) | 折扣部分 (subtotal/delivery_fee) |
| **赠品信息** |||
| `free_item_id` | varchar(100) | 赠品ID |
| `free_item_name` | varchar(255) | 赠品名称 |
| `free_item_quantity` | int | 赠品数量 |
| `free_item_price` | bigint | 赠品价格(分) |
| **应用商品** |||
| `applied_item_ids` | text | 应用的商品ID列表 (JSON数组) |

**关联关系**:
```
ttpos_takeout_order (1) ----< (N) ttpos_takeout_order_campaign
```

---

### 2.8 订单促销表

#### `ttpos_takeout_order_promo` - 外卖订单促销表

**表名**: `ttpos_takeout_order_promo`  
**用途**: 存储订单使用的促销码信息

**核心字段**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `takeout_order_uuid` | varchar(255) | 外卖订单UUID |
| `platform` | varchar(20) | 外卖平台 |
| `promo_code` | varchar(100) | 促销代码 |
| `promo_name` | varchar(255) | 促销名称 |
| `promo_description` | varchar(500) | 促销描述 |
| `promo_amount` | bigint | 促销金额(分) |
| `mex_funded_ratio` | int | 商户承担比例(%) |
| `mex_funded_amount` | bigint | 商户承担金额(分) |
| `targeted_price` | bigint | 目标价格-订单小计(分) |

**关联关系**:
```
ttpos_takeout_order (1) ----< (N) ttpos_takeout_order_promo
```

---

### 2.9 订单原料表

#### `ttpos_takeout_order_material` - 外卖订单原料表

**表名**: `ttpos_takeout_order_material`  
**用途**: 记录外卖订单的原料消耗，用于库存管理和成本核算

**核心字段**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `takeout_order_uuid` | bigint | 外卖订单UUID |
| `material_uuid` | bigint | 原料ID |
| `warehouse_uuid` | bigint | 仓库ID |
| `num` | decimal(20,4) | 实际使用数量 |
| `staff_shift_log_uuid` | bigint | 员工班次记录ID |
| `is_summarized` | int | 是否已统计 (0=未统计, 1=已统计) |

**索引**:
```sql
KEY idx_takeout_order_uuid (takeout_order_uuid)
KEY idx_material_uuid (material_uuid)
KEY idx_warehouse_uuid (warehouse_uuid)
KEY idx_is_summarized_create_time (is_summarized, create_time)  -- 统计查询优化
```

**关联关系**:
```
ttpos_takeout_order (1) ----< (N) ttpos_takeout_order_material
ttpos_material (1) ----< (N) ttpos_takeout_order_material
ttpos_warehouse (1) ----< (N) ttpos_takeout_order_material
```

---

### 2.10 平台设置表

#### `ttpos_takeout_settings` - 外卖平台配置表

**表名**: `ttpos_takeout_settings`  
**用途**: 存储各外卖平台的配置信息

**核心字段**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `platform` | varchar(20) | 外卖平台 (grab/foodpanda/lineman) |
| `is_enabled` | int | 是否启用 (0=关闭, 1=开启) |
| `auto_accept` | int | 自动接单开关 (0=关闭, 1=开启) |
| `max_amount` | bigint | 自动接单金额上限(分) |
| `platform_config` | text | 平台特定配置 (JSON) |

**索引**:
```sql
UNIQUE KEY uk_platform (platform, delete_time)  -- 每个平台唯一配置
KEY idx_platform (platform, delete_time)
```

**业务规则**:
1. 每个平台只能有一条配置记录
2. `auto_accept` 控制是否自动接单
3. `max_amount` 限制自动接单的订单金额上限
4. `platform_config` 存储平台特定的 JSON 配置

---

### 2.11 导入日志表

#### `ttpos_takeout_import_log` - 外卖导入日志表

**表名**: `ttpos_takeout_import_log`  
**用途**: 记录菜单导入/导出的日志信息

**核心字段**:

| 字段名 | 类型 | 说明 | 枚举值 |
|--------|------|------|--------|
| `platform` | varchar(50) | 外卖平台 | |
| `import_type` | int | 导入类型 | 1=TTPOS推送到平台, 2=平台推送到TTPOS |
| `import_direction` | varchar(200) | 导入方向描述 | |
| `status` | int | 导入状态 | 0=进行中, 1=成功, 2=失败 |
| `progress` | int | 进度百分比 | 0-100 |
| `success_count` | int | 成功数量 | |
| `failure_count` | int | 失败数量 | |
| `total_count` | int | 总数量 | |
| `error_message` | text | 错误信息 | |
| `start_time` | int | 开始时间 | |
| `end_time` | int | 结束时间 | |
| `duration` | int | 耗时(秒) | |

**索引**:
```sql
KEY idx_platform (platform)
KEY idx_import_type (import_type)
KEY idx_status (status)
KEY idx_create_time (create_time)  -- 时间范围查询
```

**业务规则**:
1. 记录每次菜单导入/导出的完整过程
2. 支持进度追踪（实时更新 `progress`）
3. 区分成功和失败的记录数
4. 保存错误信息用于问题排查

---

## 3. 表关系图

### 3.1 核心关系图

```
┌─────────────────────────┐
│  ttpos_takeout          │  平台管理
│  (平台开关和菜单)         │
└─────────────────────────┘
            │
            │ platform
            ▼
┌─────────────────────────┐
│  ttpos_takeout_order    │  订单主表
│  (订单核心信息)           │
└─────────────────────────┘
            │
            ├─────────────────────┐
            │                     │
            ▼                     ▼
┌──────────────────────┐  ┌───────────────────────────┐
│ ttpos_takeout_order_ │  │ ttpos_takeout_order_      │
│ receiver             │  │ item                      │
│ (收货人信息) 1:1     │  │ (订单商品) 1:N            │
└──────────────────────┘  └───────────────────────────┘
                                      │
                                      │ takeout_order_item_uuid
                                      ▼
                          ┌───────────────────────────────┐
                          │ ttpos_takeout_order_item_     │
                          │ modifier                      │
                          │ (商品修饰符) 1:N               │
                          │ - flavor (规格)                │
                          │ - sauce (加料)                 │
                          │ - attr (属性)                  │
                          │ - commodity (套餐商品)         │
                          └───────────────────────────────┘
            │
            ├─────────────────────────────────┐
            │                                 │
            ▼                                 ▼
┌──────────────────────┐        ┌──────────────────────┐
│ ttpos_takeout_order_ │        │ ttpos_takeout_order_ │
│ campaign             │        │ promo                │
│ (订单活动) 1:N       │        │ (订单促销) 1:N       │
└──────────────────────┘        └──────────────────────┘
            │
            │
            ▼
┌──────────────────────┐
│ ttpos_takeout_order_ │
│ material             │
│ (订单原料) 1:N       │
└──────────────────────┘
```

### 3.2 菜单数据关系

```
┌───────────────────────────┐
│ ttpos_product_package     │  TTPOS 核心商品表
└───────────────────────────┘
            │
            │ product_package_uuid
            ▼
┌───────────────────────────┐
│ ttpos_product_package_    │  外卖商品表
│ takeout                   │  (外卖专用信息)
└───────────────────────────┘
            │
            │ product_package_takeout_uuid
            ▼
┌───────────────────────────┐
│ ttpos_product_bom_        │  外卖规格价格表
│ takeout                   │  (规格/加料价格)
└───────────────────────────┘
            │
            │ product_bom_uuid
            ▼
┌───────────────────────────┐
│ ttpos_product_bom         │  TTPOS 核心 BOM 表
└───────────────────────────┘
```

### 3.3 名称字段关系（新架构）

```
订单创建时的数据流：

1. 商品名称:
   ttpos_product_package_takeout (外卖表)
        ↓ 优先
   item_name (显示用)
   
   ttpos_product_package (核心表)
        ↓ 始终
   ttpos_item_name (标识用)

2. 修饰符名称:
   各类型对应的核心表
        ↓ 根据类型
   ├─ ttpos_product_flavor → ttpos_modifier_name
   ├─ ttpos_product_sauce → ttpos_modifier_name
   ├─ ttpos_attribute → ttpos_modifier_name
   └─ ttpos_product_package → ttpos_modifier_name (commodity)
   
   modifier_name (显示用, 外卖表优先)
   ttpos_modifier_name (标识用, 核心表)

3. 规格名称 (commodity 专用):
   ttpos_product_bom → ProductFlavor
        ↓
   ttpos_flavor_uuid (规格UUID)
   ttpos_flavor_name (规格名称)
```

---

## 4. 数据流程

### 4.1 订单创建流程

```
1. 接收平台订单
   ↓
2. 解析订单数据
   ├─ 订单基本信息 → ttpos_takeout_order
   ├─ 收货人信息 → ttpos_takeout_order_receiver
   ├─ 活动信息 → ttpos_takeout_order_campaign
   └─ 促销信息 → ttpos_takeout_order_promo
   ↓
3. 处理订单商品
   ├─ 创建商品记录 → ttpos_takeout_order_item
   │   ├─ 设置 item_name (平台/外卖表名称)
   │   └─ 设置 ttpos_item_name (核心表名称) ✨
   └─ 处理修饰符 → ttpos_takeout_order_item_modifier
       ├─ flavor: 设置 ttpos_modifier_name ✨
       ├─ sauce: 设置 ttpos_modifier_name ✨
       ├─ attr: 设置 ttpos_modifier_name ✨
       └─ commodity: 
           ├─ 设置 ttpos_modifier_name ✨
           ├─ 设置 ttpos_flavor_uuid ✨
           ├─ 设置 ttpos_flavor_name ✨
           └─ 计算 quantity (groupItem.Num * item.Quantity) ✨
   ↓
4. 库存检查
   ├─ 计算 BOM 出库数量
   └─ 检查库存是否充足
   ↓
5. 扣减库存
   └─ 创建原料记录 → ttpos_takeout_order_material
   ↓
6. 创建送厨单
   └─ ttpos_production_order
```

### 4.2 菜单同步流程

```
TTPOS → 平台 (推送)
   ↓
1. 读取 ttpos_product_package (核心商品)
   ↓
2. 读取 ttpos_product_package_takeout (外卖信息)
   ↓
3. 组装平台菜单格式
   ↓
4. 推送到平台 API
   ↓
5. 记录日志 → ttpos_takeout_import_log

平台 → TTPOS (拉取)
   ↓
1. 从平台 API 拉取菜单
   ↓
2. 解析菜单数据
   ↓
3. 更新 ttpos_takeout.menu
   ↓
4. 记录日志 → ttpos_takeout_import_log
```

### 4.3 打印流程

```
打印请求
   ↓
判断打印类型
   ├─ 商家联 (isMerchantReceipt = true)
   │   ├─ 商品名称: 使用 ttpos_item_name ✨
   │   ├─ 修饰符名称: 使用 ttpos_modifier_name ✨
   │   └─ 规格名称: 使用 ttpos_flavor_name ✨
   │
   └─ 客户联 (isMerchantReceipt = false)
       ├─ 商品名称: 使用 item_name
       ├─ 修饰符名称: 使用 modifier_name
       └─ 规格名称: 使用 ttpos_flavor_name ✨
```

---

## 5. 关键字段说明

### 5.1 UUID 字段

所有表都使用 `bigint` 类型的 `uuid` 作为主键，优点：
- ✅ 分布式环境友好（雪花算法生成）
- ✅ 无需数据库自增，支持批量插入
- ✅ 可预生成 UUID，方便关联

### 5.2 时间字段

| 字段类型 | 格式 | 存储值 | 示例 |
|---------|------|--------|------|
| `xxx_time` | Unix 时间戳 | 秒 | 1703836800 |
| `create_time` | Unix 时间戳 | 秒 | 1703836800 |
| `update_time` | Unix 时间戳 | 秒 | 1703836800 |
| `delete_time` | Unix 时间戳 | 秒，0表示未删除 | 0 / 1703836800 |

### 5.3 金额字段

| 字段类型 | 类型 | 单位 | 说明 |
|---------|------|------|------|
| `price` | decimal(20,4) | 元 | 商品价格，4位小数 |
| `subtotal` | bigint | 分 | 订单金额，最小货币单位 |
| `xxx_fee` | bigint | 分 | 各类费用，最小货币单位 |

**为什么使用两种类型？**
- `decimal(20,4)`: 商品级别价格，需要精确计算
- `bigint`: 订单级别金额，避免浮点数误差

### 5.4 平台字段

| 平台代码 | 说明 | 支持状态 |
|---------|------|----------|
| `grab` | Grab (东南亚) | ✅ 已支持 |
| `lineman` | Lineman (泰国) | ✅ 已支持 |
| `foodpanda` | FoodPanda | 🚧 规划中 |

### 5.5 新增名称字段 ✨

| 表名 | 字段名 | 用途 | 数据来源 |
|------|--------|------|----------|
| `ttpos_takeout_order_item` | `ttpos_item_name` | TTPOS 标识用商品名称 | `ttpos_product_package` (核心表) |
| `ttpos_takeout_order_item_modifier` | `ttpos_modifier_name` | TTPOS 标识用修饰符名称 | 核心表（根据类型不同） |
| `ttpos_takeout_order_item_modifier` | `ttpos_flavor_uuid` | 规格UUID（commodity专用） | `product_package_group_item.product_bom_uuid` |
| `ttpos_takeout_order_item_modifier` | `ttpos_flavor_name` | 规格名称（commodity专用） | `ttpos_product_bom.ProductFlavor` |

**设计理念**:
1. **双名称策略**: 显示名称（客户看到的）+ 标识名称（TTPOS 标准的）
2. **订单时快照**: 创建订单时保存名称，避免后续修改影响历史订单
3. **减少查询**: 打印、送厨单等操作直接使用保存的名称，无需额外查询
4. **数据一致性**: 订单的名称始终保持创建时的状态

---

## 6. 索引策略

### 6.1 索引类型

| 索引类型 | 前缀 | 用途 | 示例 |
|---------|------|------|------|
| 主键 | `PRIMARY KEY` | 唯一标识 | `PRIMARY KEY (id)` |
| 唯一索引 | `UNIQUE KEY uk_` | 防重复 | `uk_platform_order` |
| 普通索引 | `KEY idx_` | 查询优化 | `idx_order_state` |

### 6.2 常用索引模式

#### A. 平台+ID 联合唯一索引
```sql
UNIQUE KEY uk_platform_order (platform, platform_order_id, delete_time)
```
**用途**: 防止同一平台的订单重复创建

#### B. 外键+软删除 联合索引
```sql
KEY idx_takeout_order_uuid (takeout_order_uuid, delete_time)
```
**用途**: 查询订单关联数据，同时过滤软删除

#### C. 状态+时间 联合索引
```sql
KEY idx_order_state (order_state, delete_time)
KEY idx_order_time (order_time, delete_time)
```
**用途**: 按状态和时间范围查询订单

#### D. 统计查询优化索引
```sql
KEY idx_is_summarized_create_time (is_summarized, create_time)
```
**用途**: 原料统计时快速筛选未统计记录

### 6.3 索引使用建议

✅ **推荐做法**:
- 查询条件中的字段建立索引
- 关联查询的外键建立索引
- `delete_time` 组合到常用索引中
- 避免过多单列索引，优先联合索引

❌ **避免做法**:
- 在低选择性字段（如 `is_mapped`）单独建索引
- 索引列过多（超过5个）
- 频繁更新的字段建索引

---

## 7. 最佳实践

### 7.1 订单处理

#### ✅ 推荐做法

1. **使用事务**: 订单创建涉及多表操作，必须使用事务
   ```go
   tx := db.Begin()
   defer func() {
       if r := recover(); r != nil {
           tx.Rollback()
       }
   }()
   
   // 创建订单
   tx.Create(&order)
   // 创建商品
   tx.Create(&items)
   // 创建修饰符
   tx.Create(&modifiers)
   
   tx.Commit()
   ```

2. **批量插入**: 使用批量插入提升性能
   ```go
   db.CreateInBatches(&items, 100)  // 每批100条
   ```

3. **名称冗余**: 订单创建时保存名称快照
   ```go
   item.ItemName = productInfo.Name          // 显示用
   item.TtposItemName = productInfo.TtposName // 标识用
   
   modifier.ModifierName = modifierInfo.Name  // 显示用
   modifier.TtposModifierName = modifierInfo.TtposName // 标识用
   
   // commodity 类型额外保存规格
   if modifier.TtposModifierType == "commodity" {
       modifier.TtposFlavorUuid = modifierInfo.TtposFlavorUuid
       modifier.TtposFlavorName = modifierInfo.TtposFlavorName
   }
   ```

4. **幂等性保证**: 使用唯一索引防止重复创建
   ```sql
   UNIQUE KEY uk_platform_order (platform, platform_order_id, delete_time)
   ```

#### ❌ 避免做法

1. ❌ 不要在订单处理中进行复杂计算
2. ❌ 不要在订单处理中调用外部 API
3. ❌ 不要遗漏事务处理
4. ❌ 不要在循环中执行数据库操作

### 7.2 查询优化

#### ✅ 推荐做法

1. **使用索引**: 查询条件利用索引
   ```go
   db.Where("platform = ? AND order_state = ? AND delete_time = 0", platform, state)
   ```

2. **预加载关联**: 使用 Preload 避免 N+1 问题
   ```go
   db.Preload("TakeoutOrderItems.TakeoutOrderItemModifiers").
      Preload("TakeoutOrderReceiver").
      Find(&orders)
   ```

3. **分页查询**: 大数据量使用分页
   ```go
   db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&orders)
   ```

4. **选择字段**: 只查询需要的字段
   ```go
   db.Select("uuid, platform, order_state, order_time").Find(&orders)
   ```

### 7.3 打印优化

#### ✅ 推荐做法

1. **直接使用订单数据**: 不需要额外查询
   ```go
   // 商家联
   if isMerchantReceipt {
       itemName = item.TtposItemName       // 核心表名称
       modifierName = modifier.TtposModifierName
   } else {
       itemName = item.ItemName             // 平台/外卖表名称
       modifierName = modifier.ModifierName
   }
   
   // commodity 规格（两种联都显示）
   if modifier.TtposModifierType == "commodity" {
       flavorName = modifier.TtposFlavorName
   }
   ```

2. **批量查询分类**: 只查询必要的关联数据
   ```go
   // 只查询分类信息（用于送厨单路由）
   productPackages := productPackageRepo.GetProductPackageList(
       repository.CommonRepo.WhereInUuids(packageUuids),
       productPackageRepo.WithProductCategory(),
   )
   ```

### 7.4 数据一致性

#### ✅ 推荐做法

1. **软删除**: 使用 `delete_time` 实现软删除
   ```go
   db.Model(&order).Update("delete_time", time.Now().Unix())
   ```

2. **乐观锁**: 使用版本号防止并发更新
   ```go
   db.Model(&order).
      Where("uuid = ? AND version = ?", uuid, version).
      Update("order_state", newState)
   ```

3. **原始数据保留**: 保存 `raw_data` 用于问题追溯
   ```go
   order.RawData = string(rawJSON)
   ```

---

## 附录

### A. 相关文档

- [外卖订单流程](./takeout-order-flow.md)
- [菜单同步指南](./menu-sync-guide.md)
- [打印模板开发](./printer-template-guide.md)

### B. 更新记录

| 日期 | 版本 | 变更内容 | 作者 |
|------|------|----------|------|
| 2025-12-29 | v1.0 | 初始版本，完整的数据库架构说明 | TTPOS Team |
| 2025-12-29 | v1.0 | 新增名称字段说明（ttpos_item_name等） | TTPOS Team |
| 2025-12-29 | v1.0 | 新增 commodity 规格字段说明 | TTPOS Team |

### C. 联系方式

如有问题或建议，请联系：
- **项目仓库**: ttpos-server-go
- **文档位置**: `docs/shared/takeout/database-architecture.md`

---

**最后更新**: 2025-12-29  
**文档版本**: v1.0

