# Job 实体模型说明

## 基本信息

- **实体名称**: Job
- **表名**: job
- **所属模块**: ttpos-takeout
- **描述**: 外送订单实体，用于管理外送订单信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | ID | 主键 |
| Uuid | string | 外送订单UUID | 唯一标识 |
| ShopRefNo | string | 餐馆订单参考 | 如UUID |
| CustomerMobile | string | 下单客户电话 | |
| CustomerEmail | string | 下单客户联系邮件 | |
| ProviderName | string | 外送供应商 | skootar,grab |
| TakeoutRefNo | string | 外送系统订单号 | |
| ShopLocationUuid | string | 餐馆位置信息 | |
| ConsumerLocationUuid | string | 消费者位置信息 | |
| JobDate | string | 订单日期 | YYYY-MM-DD |
| StartTime | string | 开始时间 | 24小时制或"now" |
| FinishTime | string | 订单结束时间 | |
| PaymentType | string | 支付类型 | invoice/cash/creditcard/prepaid |
| JobStatus | string | 外送订单状态 | |
| Remark | string | 订单备注 | |
| Reserved1 | string | 保留字段1 | |
| Reserved2 | string | 保留字段2 | |
| CreatedAt | *gtime.Time | 创建时间 | |
| UpdatedAt | *gtime.Time | 更新时间 | |
| DeletedAt | *gtime.Time | 软删除 | |
| CallbackUrl | string | 订单状态更新回调 | |
| SkootarId | string | 骑手Id | |
| SkootarName | string | 骑手名称 | |
| SkootarPhone | string | 骑手电话 | |
| SkootarImageUrl | string | 骑手头像 | |
| SkootarRating | float64 | 骑手评分 | |

## 关联关系

### 关联实体
- **ShopLocationUuid** → JobLocation.Uuid（餐馆位置）
- **ConsumerLocationUuid** → JobLocation.Uuid（消费者位置）
- **Uuid** → JobDriver.JobUuid（骑手信息）
- **Uuid** → JobStatusLog.JobUuid（状态日志）
- **TakeoutRefNo** → CallbackMsg.TakeoutRefNo（回调消息）

## 数据流分析

### 数据来源
- 从餐馆系统（TTPOS）创建的外送订单
- 通过外送供应商API创建订单

### 数据流向
1. **订单创建流程**:
   - 餐馆系统创建外送订单
   - 调用外送服务API创建Job
   - 设置订单信息（客户信息、位置信息、支付方式等）
   - 状态初始化为待处理

2. **订单处理流程**:
   - 外送供应商分配骑手
   - 更新骑手信息（SkootarId、SkootarName等）
   - 更新订单状态（JobStatus）
   - 记录状态变更日志（JobStatusLog）

3. **订单状态回调**:
   - 外送供应商通过CallbackUrl回调订单状态
   - 更新订单状态
   - 记录回调消息（CallbackMsg）

4. **订单完成流程**:
   - 订单配送完成
   - 更新FinishTime
   - 记录最终状态

### 业务场景
- 外送订单管理
- 多外送供应商支持（skootar、grab等）
- 骑手信息跟踪
- 订单状态实时更新

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 唯一索引: TakeoutRefNo（外送系统订单号唯一）
- 普通索引: ShopRefNo（餐馆订单查询）
- 普通索引: ProviderName（供应商查询）
- 普通索引: JobStatus（状态查询）
- 普通索引: JobDate（日期查询）
- 普通索引: CreatedAt（时间范围查询）

## 注意事项

1. 支持多个外送供应商（ProviderName）
2. ShopLocationUuid 和 ConsumerLocationUuid 关联到 JobLocation 表
3. 骑手信息可能为空（未分配时）
4. CallbackUrl 用于接收外送供应商的状态回调
5. 使用软删除机制（DeletedAt）

