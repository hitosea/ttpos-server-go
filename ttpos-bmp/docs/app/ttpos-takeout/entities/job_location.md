# JobLocation 实体模型说明

## 基本信息

- **实体名称**: JobLocation
- **表名**: job_location
- **所属模块**: ttpos-takeout
- **描述**: 外送订单位置实体，用于管理餐馆和消费者的位置信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 主键ID |
| Uuid | string | 全局唯一uuid | 唯一标识 |
| LocationType | int | 位置类型 | 0-餐馆 1-消费者 |
| AddressName | string | 地址说明 | |
| Address | string | 详细地址 | |
| Lat | string | 纬度 | |
| Lng | string | 经度 | |
| ContactName | string | 联系人名称 | |
| ContactPhone | string | 联系人号码 | |
| Seq | int | 地址序列 | 1开始 |
| Remark | string | 备注 | |
| CreatedAt | *gtime.Time | 创建时间 | |
| UpdatedAt | *gtime.Time | 更新时间 | |
| DeletedAt | *gtime.Time | 软删除 | |

## 关联关系

### 关联实体
- **Uuid** → Job.ShopLocationUuid（餐馆位置）
- **Uuid** → Job.ConsumerLocationUuid（消费者位置）

## 数据流分析

### 数据来源
- 外送订单创建时的位置信息
- 餐馆位置从系统配置获取
- 消费者位置从订单信息获取

### 数据流向
1. **位置创建流程**:
   - 创建外送订单时创建位置记录
   - 餐馆位置（LocationType=0）从系统配置获取
   - 消费者位置（LocationType=1）从订单信息获取
   - 记录详细地址和经纬度

2. **位置使用流程**:
   - Job 实体通过 ShopLocationUuid 和 ConsumerLocationUuid 关联位置
   - 外送供应商使用位置信息进行配送
   - 支持多个位置点（通过 Seq 排序）

### 业务场景
- 餐馆位置管理
- 消费者配送地址管理
- 地理位置信息存储
- 多地址点支持

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: LocationType（类型查询）
- 普通索引: Lat + Lng（地理位置查询）

## 注意事项

1. LocationType 区分餐馆位置和消费者位置
2. Seq 字段用于支持多个地址点（如多个取餐点或配送点）
3. Lat 和 Lng 用于地理位置计算和地图显示
4. 使用软删除机制（DeletedAt）

