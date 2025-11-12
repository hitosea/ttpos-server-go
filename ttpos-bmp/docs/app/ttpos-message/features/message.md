### 模块：Message（消息服务）

定位：对外提供标准化的消息服务能力（gRPC），负责消息提交、状态查询与失败重发，落库记录并与队列解耦。

主要接口
- SendMessage：接收发送请求，参数校验 → 幂等校验（message_uuid）→ 模板查询与渲染 → 创建消息记录（状态：待发送）→ 发布队列 → 快速返回提交结果
- GetMessageStatus：根据 message_uuid 查询消息记录与状态信息，返回错误信息与时间戳等详情
- ResendMessage：仅允许对失败消息重发，校验最大重试次数 → 置为待发送 → 重新发布队列

核心数据结构（参考 internal/model/dto）
- SendMessageInput/Output：发送请求与返回
- GetMessageStatusInput/Output：状态查询请求与返回
- ResendMessageInput/Output：重发请求与返回
- MessageTemplateDTO：消息模板（主题/内容/变量）
- MessageRecordDTO：消息记录（类型、接收人、内容、状态、错误信息、重试次数、时间）
- MessageSendLogDTO：发送日志（请求/响应快照、错误信息、时间）

关键流程
1) 幂等性：以 message_uuid 作为唯一键避免重复提交
2) 模板渲染：支持 {{var}} 变量替换，主题/正文均可渲染
3) 状态管理：待发送 → 发送中 → 发送成功/失败；失败时记录错误原因
4) 发送日志：记录每次发送的请求/响应数据，便于审计与排错

错误与边界
- 参数错误：模板不存在、模板禁用、变量 JSON 不合法、邮箱/手机号格式不合法
- 状态约束：仅失败状态允许 Resend，限制最大重试次数
- 幂等冲突：重复 message_uuid 直接返回已提交结果

依赖
- Queue 模块：提交 RocketMQ 消息
- Sender 模块：由消费者触发实际发送



