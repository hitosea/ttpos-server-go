# MessageSendLog 实体模型说明

## 基本信息

- **实体名称**: MessageSendLog
- **表名**: message_send_log
- **所属模块**: ttpos-message
- **描述**: 消息发送日志实体，用于记录每次消息发送的详细日志

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | uint64 | 日志ID | 主键 |
| MessageUuid | string | 消息UUID | 关联消息记录 |
| SendTime | int | 发送时间 | 时间戳 |
| SendResult | int | 发送结果 | 0-失败 1-成功 |
| ErrorMessage | string | 错误信息 | |
| RequestData | string | 请求数据 | JSON格式 |
| ResponseData | string | 响应数据 | JSON格式 |
| CreatedAt | int | 创建时间 | 时间戳 |

## 关联关系

### 关联实体
- **MessageUuid** → MessageRecord.Uuid（关联消息记录）

## 数据流分析

### 数据来源
- 消息发送过程中的日志记录
- 每次发送尝试都会创建一条日志

### 数据流向
1. **日志记录流程**:
   - 消息发送时创建日志记录
   - 记录请求数据（RequestData）和响应数据（ResponseData）
   - 记录发送结果（SendResult）
   - 如果失败，记录错误信息（ErrorMessage）

2. **日志查询流程**:
   - 通过 MessageUuid 查询某条消息的所有发送日志
   - 用于分析发送失败原因
   - 用于审计和问题排查

### 业务场景
- 消息发送详细日志
- 发送失败原因分析
- 消息发送审计
- 问题排查和调试

## 索引建议

- 主键索引: Id
- 普通索引: MessageUuid（消息查询）
- 普通索引: SendTime（时间范围查询）
- 普通索引: SendResult（结果查询）

## 注意事项

1. 一条消息可能有多条发送日志（重试场景）
2. RequestData 和 ResponseData 存储完整的请求和响应数据，用于问题排查
3. SendTime 记录实际发送时间，可能与 MessageRecord.SendTime 不同（重试场景）
4. 日志记录不删除，用于长期审计

