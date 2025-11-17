### 模块：Queue（队列服务）

定位：封装消息队列能力（RocketMQ），提供初始化、可用性检测与消息发布，支撑异步发送与削峰解耦。

主要接口
- Init：读取配置并初始化 RocketMQ 客户端、消费线程池等资源
- PublishMessage：将消息发送任务发布至 Topic，包含消息体序列化、Tag 区分（email/sms）与错误处理
- IsEnabled：根据配置开关决定是否启用队列能力

消息体（参考 internal/model/dto.RocketMQMessage）
- 基本字段：message_uuid、template_id、message_type（email/sms）、recipient、subject、content、message_args 等
- 用途：供消费者解析并驱动 Sender 进行实际发送

重试与可靠性
- 发送端：失败重试（可配置），记录告警日志
- 消费端：RocketMQ 自动重试 + 业务重试策略（见 Consumer）

配置要点（manifest/config/config.yaml）
- queue.switch：是否启用
- queue.driver：驱动选择（rocketmq）
- queue.rocketmq：nameSrv、broker、重试次数、日志等级、超时等

依赖
- RocketMQ 集群
- 日志系统（记录发布与失败信息）



