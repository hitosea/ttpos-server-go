# MessageRecord 实体模型说明

## 基本信息

- **实体名称**: MessageRecord
- **表名**: message_record
- **所属模块**: ttpos-message
- **描述**: 消息记录实体，用于记录所有发送的消息（邮件/短信）

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | uint64 | 消息ID | 主键 |
| Uuid | string | 消息UUID | 唯一标识 |
| TemplateId | uint64 | 模板ID | 关联模板 |
| TemplateUuid | string | 模板UUID | 关联模板 |
| MessageType | string | 消息类型 | email/sms |
| Recipient | string | 接收人 | 邮箱或手机号 |
| Subject | string | 消息主题 | |
| Content | string | 消息内容 | 渲染后 |
| MessageArgs | string | 消息参数 | JSON格式 |
| Status | int | 状态 | 0-待发送 1-发送中 2-发送成功 3-发送失败 |
| ErrorMessage | string | 错误信息 | |
| RetryCount | int | 重试次数 | |
| CompanyUuid | string | 公司UUID | |
| OperatorUuid | string | 操作人UUID | |
| SendTime | int | 发送时间 | 时间戳 |
| CreatedAt | int | 创建时间 | 时间戳 |
| UpdatedAt | int | 更新时间 | 时间戳 |
| DeletedAt | int | 删除时间 | 时间戳，软删除 |

## 关联关系

### 关联实体
- **TemplateUuid** → MessageTemplate.Uuid（关联消息模板）
- **CompanyUuid** → 公司（外部关联）
- **OperatorUuid** → 操作人（外部关联）

### 关联记录
- **Uuid** → MessageSendLog.MessageUuid（消息发送日志）

## 数据流分析

### 数据来源
- 业务系统触发的消息发送请求
- 通过消息服务API创建

### 数据流向
1. **消息创建流程**:
   - 业务系统调用消息服务API
   - 选择消息模板（TemplateUuid）
   - 传入消息参数（MessageArgs）
   - 创建 MessageRecord，状态为"待发送"
   - 渲染模板内容（Content）

2. **消息发送流程**:
   - 消息队列或定时任务处理待发送消息
   - 更新状态为"发送中"
   - 调用邮件或短信服务发送
   - 更新状态为"发送成功"或"发送失败"
   - 记录发送时间（SendTime）
   - 如果失败，记录错误信息（ErrorMessage）

3. **重试机制**:
   - 发送失败时，增加重试次数（RetryCount）
   - 根据重试策略决定是否重新发送
   - 记录每次发送的日志（MessageSendLog）

### 业务场景
- 邮件发送记录
- 短信发送记录
- 消息发送状态跟踪
- 消息发送失败重试

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: TemplateUuid（模板查询）
- 普通索引: MessageType（类型查询）
- 普通索引: Status（状态查询）
- 普通索引: CompanyUuid（公司查询）
- 普通索引: Recipient（接收人查询）
- 普通索引: CreatedAt（时间范围查询）

## 注意事项

1. Content 字段存储渲染后的最终内容
2. MessageArgs 存储原始参数，用于重试和审计
3. Status 字段跟踪消息发送状态
4. 支持重试机制，通过 RetryCount 控制重试次数
5. 使用软删除机制（DeletedAt）

