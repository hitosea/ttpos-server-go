# CallbackMsg 实体模型说明

## 基本信息

- **实体名称**: CallbackMsg
- **表名**: callback_msg
- **所属模块**: ttpos-takeout
- **描述**: 外送订单回调消息实体，用于记录外送供应商的回调消息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 主键ID |
| CreatedAt | *gtime.Time | 创建时间 | |
| UpdatedAt | *gtime.Time | 修改时间 | |
| DeletedAt | *gtime.Time | 软删除 | |
| Uuid | string | 全局唯一ID | 唯一标识 |
| TakeoutRefNo | string | 外送系统订单号 | 如skootar.jobId |
| Content | string | 消息内容 | JSON格式 |
| StatusDatetime | *gtime.Time | 状态变更时间 | |

## 关联关系

### 关联实体
- **TakeoutRefNo** → Job.TakeoutRefNo（关联外送订单）

## 数据流分析

### 数据来源
- 外送供应商通过回调接口发送的消息
- 订单状态变更时的回调消息

### 数据流向
1. **回调接收流程**:
   - 外送供应商通过回调接口发送消息
   - 创建 CallbackMsg 记录
   - 解析消息内容（Content）
   - 根据 TakeoutRefNo 关联到对应的订单（Job）

2. **回调处理流程**:
   - 解析消息内容，提取订单状态信息
   - 更新订单状态（Job.JobStatus）
   - 记录状态变更日志（JobStatusLog）
   - 更新状态变更时间（StatusDatetime）

### 业务场景
- 外送订单状态回调
- 回调消息记录和审计
- 订单状态自动更新

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: TakeoutRefNo（订单号查询）
- 普通索引: StatusDatetime（时间范围查询）
- 普通索引: CreatedAt（时间范围查询）

## 注意事项

1. Content 字段存储完整的回调消息内容（JSON格式）
2. TakeoutRefNo 用于关联外送订单
3. StatusDatetime 记录状态变更的实际时间
4. 使用软删除机制（DeletedAt）

