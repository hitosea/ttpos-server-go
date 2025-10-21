# TTPOS 消息中心服务

TTPOS 消息中心服务（ttpos-message）是 TTPOS 中台工程的核心模块之一，负责统一管理和发送各类消息（邮件、短信等），为其他业务模块提供标准化的消息发送能力。

## 📋 功能特性

- ✅ 统一的消息发送接口（gRPC）
- ✅ 异步消息处理机制（RocketMQ）
- ✅ 多渠道消息发送支持（邮件、短信）
- ✅ 消息发送记录追踪
- ✅ 消息模板管理
- ✅ 消息发送失败重试
- ✅ 幂等性保证
- ✅ 高可用、高并发设计

## 🛠️ 技术栈

- **框架**: GoFrame v2.x
- **通信协议**: gRPC
- **消息队列**: RocketMQ
- **数据库**: MySQL（独立 schema: messages）
- **邮件服务**: Mailgun API

## 📦 项目结构

```
ttpos-message/
├── api/                    # gRPC API 定义（自动生成）
│   └── message/
│       └── v1/
├── internal/               # 内部实现代码
│   ├── boot/              # 启动初始化
│   ├── cmd/               # 命令行入口
│   ├── consts/            # 常量定义
│   ├── controller/        # 控制器层
│   │   └── rpc/
│   │       └── message/
│   ├── dao/               # 数据访问层（自动生成）
│   ├── logic/             # 业务逻辑层
│   │   ├── message/       # 消息服务
│   │   ├── queue/         # 队列服务
│   │   └── sender/        # 发送服务
│   ├── model/             # 数据模型
│   │   ├── do/            # 数据对象（自动生成）
│   │   ├── dto/           # 数据传输对象
│   │   └── entity/        # 数据实体（自动生成）
│   └── service/           # 服务接口（自动生成）
├── manifest/              # 配置清单
│   ├── config/            # 配置文件
│   ├── protobuf/          # Protobuf 定义
│   └── sql/               # 数据库迁移脚本
├── docs/                  # 文档
│   └── requirement.md     # 需求文档
├── main.go                # 程序入口
└── Makefile               # 构建脚本
```

## 🚀 快速开始

### 环境要求

- Go 1.20+
- MySQL 5.7+
- RocketMQ 4.9+
- Mailgun 账号

### 安装依赖

```bash
cd ttpos-bmp/app/ttpos-message
go mod tidy
```

### 配置文件

复制配置模板并修改：

```bash
cp manifest/config/config.tpl.yaml manifest/config/config.yaml
```

配置项说明：

```yaml
# 数据库配置
database:
  default:
    link: "mysql:username:password@tcp(host:port)/messages"

# gRPC 服务配置
grpc:
  name: "ttpos-message"
  address: ":14032"

# RocketMQ 配置
queue:
  switch: true
  driver: "rocketmq"
  groupName: "ttpos-message"
  rocketmq:
    endpoint: "your-rocketmq-endpoint"
    nameSrvAdders: ["your-rocketmq-addr:9876"]

# Mailgun 配置
mailgun:
  domain: "your-domain.mailgun.org"
  apiKey: "your-api-key"
  fromEmail: "noreply@yourdomain.com"
  fromName: "TTPOS System"
```

### 数据库初始化

执行数据库迁移脚本：

```bash
# 依次执行 manifest/sql 目录下的 up.sql 文件
mysql -u username -p messages < manifest/sql/20250121100001_create_message_template.up.sql
mysql -u username -p messages < manifest/sql/20250121100002_create_message_record.up.sql
mysql -u username -p messages < manifest/sql/20250121100003_create_message_send_log.up.sql
```

### 代码生成

```bash
# 生成 Protobuf 代码
make proto

# 生成 DAO 代码（需要先配置数据库）
make dao

# 生成 Service 接口代码
make service
```

### 运行服务

```bash
# 开发模式运行
make run

# 编译并运行
make build
./bin/ttpos-message
```

## 📡 API 使用

### gRPC 接口

#### 1. 发送消息

```protobuf
rpc SendMessage(SendMessageReq) returns (SendMessageResp);
```

请求示例：

```go
import (
    v1 "ttpos-bmp/app/ttpos-message/api/message/v1"
)

req := &v1.SendMessageReq{
    MessageUuid: "msg-uuid-12345",
    TemplateId:  1,
    MessageArgs: `{"username":"张三","order_no":"ORD001"}`,
    MessageType: "email",
    Recipient:   "user@example.com",
    Subject:     "订单确认",
    CompanyId:   100,
    OperatorId:  1,
}

resp, err := messageClient.SendMessage(ctx, req)
```

#### 2. 查询消息状态

```protobuf
rpc GetMessageStatus(GetMessageStatusReq) returns (GetMessageStatusResp);
```

#### 3. 重发消息

```protobuf
rpc ResendMessage(ResendMessageReq) returns (ResendMessageResp);
```

## 🔧 开发指南

### 添加新的消息模板

在数据库中添加模板记录：

```sql
INSERT INTO message_template (
    uuid, template_name, template_type, template_subject, 
    template_content, template_args, status, created_at, updated_at
) VALUES (
    'tpl_welcome', '欢迎邮件', 'email', '欢迎加入 {{company_name}}',
    '<html><body><h2>欢迎 {{username}}</h2><p>感谢您注册 {{company_name}}！</p></body></html>',
    '{"username":"用户名","company_name":"公司名称"}',
    1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
);
```

### 消息状态说明

- `0` - 待发送：消息已创建，等待发送
- `1` - 发送中：消息正在发送
- `2` - 发送成功：消息已成功发送
- `3` - 发送失败：消息发送失败

### 消息类型说明

- `email` - 邮件消息（已实现）
- `sms` - 短信消息（预留）

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/logic/message/...
```

## 📊 监控与日志

### 健康检查

```bash
curl http://localhost:14031/health
```

### 配置查询（调试）

```bash
curl http://localhost:14031/debug/config
```

### 日志位置

- 应用日志：`./log/message/`
- 队列日志：`./log/queue/`

## 🔄 部署

### Docker 部署

```bash
# 构建镜像
docker build -t ttpos-message:latest .

# 运行容器
docker run -d \
  --name ttpos-message \
  -p 14031:14031 \
  -p 14032:14032 \
  -v /path/to/config.yaml:/app/manifest/config/config.yaml \
  ttpos-message:latest
```

### Kubernetes 部署

参考 `manifest/deploy/kustomize/` 目录下的配置文件。

## 📝 常见问题

### 1. 消息发送失败怎么办？

- 检查消息状态：使用 `GetMessageStatus` 接口查询
- 查看错误信息：`error_message` 字段包含详细错误
- 重试发送：使用 `ResendMessage` 接口（最多重试5次）

### 2. Mailgun 配置错误

确保以下配置正确：
- `domain`: Mailgun 验证过的域名
- `apiKey`: 从 Mailgun 控制台获取
- `fromEmail`: 必须是验证域名下的邮箱

### 3. RocketMQ 连接失败

检查配置：
- `nameSrvAdders`: 确保 RocketMQ NameServer 地址正确
- 网络连通性：确保服务能访问 RocketMQ
- 权限：如果启用了 ACL，需要配置 `accessKey` 和 `secretKey`

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 开发规范

请遵循项目的 GoFrame 开发规范：
- 参考 `.cursorrules` 文件中的 `go-rules` 和 `proto-rules`
- 使用中文注释
- 遵循目录结构规范
- 自动生成的代码不要手动修改

## 📚 相关文档

- [需求文档](docs/requirement.md)
- [GoFrame 官方文档](https://goframe.org)
- [Mailgun API 文档](https://documentation.mailgun.com/)
- [RocketMQ 官方文档](https://rocketmq.apache.org/)

## 📧 联系方式

如有问题或建议，请联系项目维护者。

## 📜 许可证

本项目遵循公司内部开源协议。
