# JobDriver 实体模型说明

## 基本信息

- **实体名称**: JobDriver
- **表名**: job_driver
- **所属模块**: ttpos-takeout
- **描述**: 外送订单骑手信息实体，用于管理订单分配的骑手信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 主键ID |
| Uuid | string | 全局唯一uuid | 唯一标识 |
| JobUuid | string | 外送订单uuid | 关联订单 |
| DriverId | string | 骑手ID | |
| DriverName | string | 骑手名称 | |
| DriverPhone | string | 骑手电话 | |
| DriverImageUrl | string | 骑手头像 | |
| Lat | string | 骑手当前纬度 | |
| Lng | string | 骑手当前经度 | |
| CreatedAt | *gtime.Time | 创建时间 | |
| UpdatedAt | *gtime.Time | 更新时间 | |
| DeletedAt | *gtime.Time | 软删除 | |

## 关联关系

### 关联实体
- **JobUuid** → Job.Uuid（关联外送订单）

## 数据流分析

### 数据来源
- 外送供应商分配骑手时创建
- 通过外送供应商API获取骑手信息

### 数据流向
1. **骑手分配流程**:
   - 外送供应商分配骑手给订单
   - 创建 JobDriver 记录
   - 记录骑手基本信息（ID、名称、电话、头像）

2. **骑手位置更新流程**:
   - 骑手位置实时更新（Lat、Lng）
   - 用于订单跟踪和地图显示
   - 通过外送供应商API或回调更新

3. **骑手信息查询流程**:
   - 通过 JobUuid 查询订单的骑手信息
   - 用于订单详情展示
   - 用于客户联系骑手

### 业务场景
- 骑手信息管理
- 骑手位置实时跟踪
- 订单配送跟踪

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: JobUuid（订单查询）
- 普通索引: DriverId（骑手查询）
- 普通索引: Lat + Lng（位置查询）

## 注意事项

1. 一个订单可能只有一个骑手（一对一关系）
2. Lat 和 Lng 实时更新，用于位置跟踪
3. DriverImageUrl 用于前端展示骑手头像
4. 使用软删除机制（DeletedAt）

