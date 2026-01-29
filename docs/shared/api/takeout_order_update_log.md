# 外卖订单更新日志

## 概述

`ttpos_takeout_order_update_log` 表用于记录外卖订单的更新历史，以 JSON 格式完整保存更新前后的订单数据。

## 表结构

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `id` | INT | 自增ID |
| `uuid` | BIGINT | 日志UUID |
| `takeout_order_uuid` | BIGINT | 外卖订单UUID |
| `old_data` | TEXT | 更新前订单数据(JSON格式) |
| `new_data` | TEXT | 更新后订单数据(JSON格式) |
| `create_time` | INT | 创建时间(时间戳) |
| `update_time` | INT | 更新时间(时间戳) |
| `delete_time` | INT | 删除时间(时间戳) |

## 索引

- `unique_uuid`: UUID 唯一索引
- `idx_takeout_order_uuid`: 订单UUID索引
- `idx_create_time`: 创建时间索引

## 数据结构

### old_data / new_data JSON 格式

```json
{
  "order": {
    "uuid": 123456,
    "platform": "lineman",
    "platform_order_id": "LM20260126001",
    "subtotal": 350.50,
    "delivery_fee": 30.00,
    "small_order_fee": 0.00,
    "eater_payment": 380.50,
    "platform_discount": 0.00,
    "merchant_discount": 20.00,
    "basket_promo": 0.00,
    "tax": 0.00,
    "merchant_charge_fee": 50.00,
    "platform_total": 380.50,
    "order_state": 10
  },
  "items": [
    {
      "uuid": 789,
      "platform_item_id": "item_001",
      "item_name": "{\"en\":\"Fried Rice\",\"th\":\"ข้าวผัด\"}",
      "quantity": 2,
      "price": 100.00,
      "ttpos_product_type": 0,
      "ttpos_item_name": "{\"en\":\"Fried Rice\",\"th\":\"ข้าวผัด\"}",
      "ttpos_item_erp_code": "F001",
      "specifications": "",
      "TakeoutOrderItemModifiers": [
        {
          "uuid": 1001,
          "platform_modifier_id": "mod_001",
          "modifier_name": "{\"en\":\"Extra Egg\",\"th\":\"ไข่เพิ่ม\"}",
          "ttpos_modifier_type": "sauce",
          "ttpos_modifier_name": "{\"en\":\"Extra Egg\",\"th\":\"ไข่เพิ่ม\"}"
        }
      ]
    }
  ]
}
```

## 使用场景

### 1. 订单更新审计
追踪订单的完整变更历史。

```sql
-- 查询某个订单的所有更新历史
SELECT 
    uuid,
    takeout_order_uuid,
    FROM_UNIXTIME(create_time) as update_time,
    old_data,
    new_data
FROM ttpos_takeout_order_update_log 
WHERE takeout_order_uuid = 123456 
ORDER BY create_time DESC;
```

### 2. 对比订单变化
使用 JSON 函数提取和对比字段变化。

```sql
-- 对比价格变化
SELECT 
    takeout_order_uuid,
    JSON_EXTRACT(old_data, '$.order.subtotal') as old_subtotal,
    JSON_EXTRACT(new_data, '$.order.subtotal') as new_subtotal,
    JSON_EXTRACT(old_data, '$.order.platform_total') as old_platform_total,
    JSON_EXTRACT(new_data, '$.order.platform_total') as new_platform_total,
    FROM_UNIXTIME(create_time) as update_time
FROM ttpos_takeout_order_update_log 
WHERE JSON_EXTRACT(old_data, '$.order.subtotal') != JSON_EXTRACT(new_data, '$.order.subtotal')
ORDER BY create_time DESC;
```

### 3. 商品变化追踪
查询商品数量变化的订单。

```sql
-- 查询商品数量有变化的订单
SELECT 
    takeout_order_uuid,
    JSON_LENGTH(JSON_EXTRACT(old_data, '$.items')) as old_item_count,
    JSON_LENGTH(JSON_EXTRACT(new_data, '$.items')) as new_item_count,
    FROM_UNIXTIME(create_time) as update_time
FROM ttpos_takeout_order_update_log 
WHERE JSON_LENGTH(JSON_EXTRACT(old_data, '$.items')) != JSON_LENGTH(JSON_EXTRACT(new_data, '$.items'))
ORDER BY create_time DESC;
```

### 4. 统计更新频率
统计订单更新的频率。

```sql
-- 统计每日订单更新次数
SELECT 
    DATE(FROM_UNIXTIME(create_time)) as update_date,
    COUNT(*) as update_count,
    COUNT(DISTINCT takeout_order_uuid) as unique_orders
FROM ttpos_takeout_order_update_log 
GROUP BY DATE(FROM_UNIXTIME(create_time))
ORDER BY update_date DESC;
```

## 代码示例

### 记录订单更新日志

```go
// 构建更新前的订单数据
oldOrderData := map[string]interface{}{
    "order": map[string]interface{}{
        "uuid":                existingOrder.Uuid,
        "platform":            existingOrder.Platform,
        "platform_order_id":   existingOrder.PlatformOrderId,
        "subtotal":            existingOrder.Subtotal,
        "delivery_fee":        existingOrder.DeliveryFee,
        "platform_total":      existingOrder.PlatformTotal,
        // ... 其他字段
    },
    "items": existingOrder.TakeoutOrderItems,
}
oldDataJSON, _ := json.Marshal(oldOrderData)

// 构建更新后的订单数据
newOrderData := map[string]interface{}{
    "order": map[string]interface{}{
        "uuid":                updatedOrder.Uuid,
        "platform":            updatedOrder.Platform,
        "platform_order_id":   updatedOrder.PlatformOrderId,
        "subtotal":            updatedOrder.Subtotal,
        "delivery_fee":        updatedOrder.DeliveryFee,
        "platform_total":      updatedOrder.PlatformTotal,
        // ... 其他字段
    },
    "items": updatedOrder.TakeoutOrderItems,
}
newDataJSON, _ := json.Marshal(newOrderData)

// 创建日志记录
updateLog := &model.TakeoutOrderUpdateLog{
    BaseModel: model.BaseModel{
        Uuid: logUuid,
    },
    TakeoutOrderUuid: orderUuid,
    OldData:          string(oldDataJSON),
    NewData:          string(newDataJSON),
}

if err := tx.Create(updateLog).Error; err != nil {
    logger.Logger.Warn("记录订单更新日志失败", zap.Error(err))
}
```

## 注意事项

1. **完整数据**: 记录包含订单主表、商品列表、修饰符等完整信息
2. **JSON 格式**: 使用 JSON 格式存储，便于查询和分析
3. **不影响主流程**: 日志记录失败不影响订单更新
4. **事务一致性**: 在订单更新事务中记录，保证数据一致性
5. **数据归档**: 建议定期归档历史数据，避免表过大

## 迁移说明

### 执行迁移

```bash
cd admin
php think migrate:run
```

### 回滚迁移

```bash
cd admin
php think migrate:rollback
```

## 相关文件

- 模型: `main/app/modules/takeout/domain/model/takeout_order_update_log.go`
- 迁移: `admin/database/migrations/20260126170000_add_takeout_order_operation_log_table.php`
- 使用: `main/app/modules/takeout/application/takeout_order_service.go`

---

**最后更新**: 2026-01-26  
**维护者**: TTPOS Team
