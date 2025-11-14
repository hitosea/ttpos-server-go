# JobStatusLog 实体模型说明

## 基本信息

- **实体名称**: JobStatusLog
- **表名**: job_status_log
- **所属模块**: ttpos-takeout
- **描述**: 外送订单状态日志实体，用于记录订单状态变更历史

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 主键ID |
| Uuid | string | 全局唯一ID | 唯一标识 |
| JobUuid | string | 外送订单uuid | 关联订单 |
| StatusBefore | string | 变更前状态 | |
| StatusAfter | string | 变更后状态 | |
| CreatedAt | *gtime.Time | 创建时间 | |
| UpdatedAt | *gtime.Time | 更新时间 | |
| DeletedAt | *gtime.Time | 软删除 | |

## 关联关系

### 关联实体
- **JobUuid** → Job.Uuid（关联外送订单）

## 数据流分析

### 数据来源
- 订单状态变更时自动创建
- 通过外送供应商API或回调触发

### 数据流向
1. **状态变更记录流程**:
   - 订单状态发生变更时创建日志
   - 记录变更前状态（StatusBefore）和变更后状态（StatusAfter）
   - 记录变更时间（CreatedAt）

2. **状态历史查询流程**:
   - 通过 JobUuid 查询订单的所有状态变更历史
   - 用于订单状态追踪
   - 用于问题排查和审计

### 业务场景
- 订单状态变更历史
- 订单状态追踪
- 订单状态审计
- 问题排查

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: JobUuid（订单查询）
- 普通索引: StatusAfter（状态查询）
- 普通索引: CreatedAt（时间范围查询）

## 注意事项

1. 每次状态变更都会创建一条日志记录
2. StatusBefore 和 StatusAfter 记录完整的状态变更信息
3. 日志记录不删除，用于长期审计
4. 使用软删除机制（DeletedAt）

