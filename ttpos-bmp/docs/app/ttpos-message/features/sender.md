### 模块：Sender（发送通道服务）

定位：对外部发送渠道进行抽象与封装，当前实现 Mailgun 邮件发送，SMS 预留扩展点。

主要接口（IMailgun）
- Init：读取并校验 Mailgun 相关配置（域名、API Key、发件人等）
- SendEmail：调用 Mailgun API 发送邮件，支持 HTML 正文与主题
- ValidateConfig：启动前或运行期校验配置
- GetConfig：输出当前生效配置（仅用于调试）

发送流程
1) 构建 HTTP 请求（鉴权、Headers、Body）
2) 发起调用并接收响应
3) 写入发送日志（请求/响应/错误）
4) 返回发送结果，由上层更新消息状态

错误处理
- 网络错误、超时、非 2xx 响应、API 业务错误
- 对错误进行归档与可观测输出，便于定位问题

配置要点（manifest/config/config.yaml → mailgun）
- domain、apiKey、fromEmail、fromName、apiBase、timeout
- 敏感信息通过环境变量注入

依赖
- 外部 Mailgun 服务
- Message 模块（用于记录发送日志与状态）



