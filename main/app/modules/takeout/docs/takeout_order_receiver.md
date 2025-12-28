# 外卖订单收货人信息表使用说明

## 概述

`TakeoutOrderReceiver` 表用于存储外卖订单的收货人信息，包括姓名、电话、详细地址、坐标等信息。

## 表结构

### 基础字段
- `id` - 主键ID
- `uuid` - UUID（唯一标识）
- `takeout_order_uuid` - 外卖订单UUID（唯一索引，关联订单表）
- `platform` - 平台名称（grab/lineman/foodpanda等）

### 收货人信息
- `receiver_name` - 收货人姓名
- `receiver_phones` - 收货人电话

### 地址信息
- `unit_number` - 单元号/门牌号
- `delivery_instruction` - 配送说明
- `poi_source` - POI来源（GRAB/GOOGLE/FACEBOOK等）
- `poi_id` - POI ID
- `address` - 完整地址
- `postcode` - 邮政编码

### 坐标信息
- `latitude` - 纬度（decimal(10,7)）
- `longitude` - 经度（decimal(10,7)）

### 审计字段
- `create_time` - 创建时间
- `update_time` - 更新时间
- `delete_time` - 删除时间（软删除）

## 使用示例

### 1. 在订单创建时保存收货人信息

```go
// 在 Application Service 中
func (s *takeoutOrderAppService) SyncNewOrder(ctx context.Context, platform string, takeoutOrderUuid string, rawData map[string]interface{}) error {
    // ... 转换订单数据 ...
    
    // 转换收货人信息
    receiverInfo, err := s.converters[platform].ConvertReceiverInfo(orderUuid, platform, submitOrderReq, currentTime)
    if err != nil {
        return errors.WithMessage(err, "转换收货人信息失败")
    }
    
    // 保存订单和收货人信息（在事务中）
    db := ctx.GetDB()
    return db.Transaction(func(tx *gorm.DB) error {
        // 1. 保存订单
        orderRepo := persistence.NewTakeoutOrderRepo(tx)
        if err := orderRepo.Create(order); err != nil {
            return err
        }
        
        // 2. 保存收货人信息（如果存在）
        if receiverInfo != nil {
            receiverRepo := persistence.NewTakeoutOrderReceiverRepo(tx)
            if err := receiverRepo.Create(receiverInfo); err != nil {
                return errors.WithMessage(err, "保存收货人信息失败")
            }
        }
        
        return nil
    })
}
```

### 2. 查询订单收货人信息

```go
// 根据订单UUID查询收货人信息
receiverRepo := persistence.NewTakeoutOrderReceiverRepo(db)
receiver, err := receiverRepo.GetByOrderUuid(orderUuid)
if err != nil {
    return err
}

if receiver != nil {
    fmt.Printf("收货人: %s\n", receiver.ReceiverName)
    fmt.Printf("电话: %s\n", receiver.ReceiverPhones)
    fmt.Printf("地址: %s\n", receiver.Address)
    fmt.Printf("坐标: (%f, %f)\n", receiver.Latitude, receiver.Longitude)
}
```

### 3. 更新收货人信息

```go
receiverRepo := persistence.NewTakeoutOrderReceiverRepo(db)

// 获取现有信息
receiver, err := receiverRepo.GetByOrderUuid(orderUuid)
if err != nil {
    return err
}

if receiver != nil {
    // 更新信息
    receiver.ReceiverPhones = "新电话号码"
    receiver.DeliveryInstruction = "新的配送说明"
    
    if err := receiverRepo.Update(receiver); err != nil {
        return err
    }
}
```

## 数据来源

### Grab SDK 结构映射

```go
// SDK 结构
type SubmitOrderRequest struct {
    Receiver *Receiver `json:"receiver,omitempty"`
    // ...
}

type Receiver struct {
    Name     *string  `json:"name,omitempty"`
    Phones   *string  `json:"phones,omitempty"`
    Address  *Address `json:"address,omitempty"`
}

type Address struct {
    UnitNumber          *string      `json:"unitNumber,omitempty"`
    DeliveryInstruction *string      `json:"deliveryInstruction,omitempty"`
    PoiSource           *string      `json:"poiSource,omitempty"`
    PoiID               *string      `json:"poiID,omitempty"`
    Address             *string      `json:"address,omitempty"`
    Postcode            *string      `json:"postcode,omitempty"`
    Coordinates         *Coordinates `json:"coordinates,omitempty"`
}

type Coordinates struct {
    Latitude  *float64 `json:"latitude,omitempty"`
    Longitude *float64 `json:"longitude,omitempty"`
}
```

## 注意事项

1. **可选字段处理**
   - Receiver 是可选的，不是所有订单都有收货人信息
   - 地址的各个字段也都是可选的
   - 使用 `Has` 方法检查字段是否存在

2. **唯一性约束**
   - `takeout_order_uuid` 有唯一索引
   - 每个订单只能有一条收货人信息记录

3. **坐标精度**
   - 纬度和经度使用 `decimal(10,7)` 类型
   - 精度约为 1.1 厘米（满足地图定位需求）

4. **事务处理**
   - 订单和收货人信息应在同一事务中保存
   - 确保数据一致性

5. **软删除**
   - 使用 `delete_time` 字段实现软删除
   - 查询时需要添加 `delete_time = 0` 条件

## 相关文件

- Model: `app/modules/takeout/domain/model/takeout_order_receiver.go`
- Repository: `app/modules/takeout/infrastructure/persistence/takeout_order_receiver_repo.go`
- Migration: `admin/database/migrations/20251222233328_create_table_takeout_order_receiver.php`
- Converter: `app/modules/takeout/infrastructure/adapter/grab/grab_order_converter.go` (ConvertReceiverInfo方法)

## 数据库迁移

### 执行迁移

在 admin 目录下执行：

```bash
# 进入 admin 目录
cd admin

# 执行迁移
php think migrate:run
```

### 回滚迁移

```bash
# 回滚最后一次迁移
php think migrate:rollback

# 回滚到指定版本
php think migrate:rollback -t 20251222233328
```

### 查看迁移状态

```bash
# 查看迁移状态
php think migrate:status
```

