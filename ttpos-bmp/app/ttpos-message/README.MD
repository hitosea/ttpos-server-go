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

### 创建环境变量文件

在项目根目录创建 `.env` 文件（注意：`.env` 文件不会被提交到 Git）：

```bash
# 在 ttpos-bmp 根目录创建 .env 文件
cd /path/to/ttpos-bmp
touch .env
```

编辑 `.env` 文件，添加必需的环境变量（参考下面的"环境变量配置"章节）。

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

# Nacos 配置
nacos:
  server:
    ip: $NACOS_SERVER_IP
    port: $NACOS_SERVER_PORT
  config:
    dataId: ttpos-message.yaml
    group: DEFAULT_GROUP

# RocketMQ 配置
queue:
  switch: true
  driver: "rocketmq"
  groupName: "ttpos-message"
  rocketmq:
    endpoint: "$ROCKETMQ_ENDPOINT"
    nameSrvAdders: ["$ROCKETMQ_NAME_SRV_ADDR"]
    accessKey: "$ROCKETMQ_ACCESS_KEY"
    secretKey: "$ROCKETMQ_SECRET_KEY"

# Mailgun 配置
mailgun:
  domain: "$MAILGUN_DOMAIN"
  apiKey: "$MAILGUN_API_KEY"
  fromEmail: "$MAILGUN_FROM_EMAIL"
  fromName: "TTPOS System"
```

### 环境变量配置

在项目根目录创建 `.env` 文件（或在部署环境中设置环境变量）：

```bash
# .env 文件示例

# ==================== 数据库配置 ====================
DB_USERNAME=root
DB_PASSWORD=your_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_PORT_OPEN=3306
DB_NAME=messages

# ==================== Nacos 配置 ====================
NACOS_SERVER_IP=127.0.0.1
NACOS_SERVER_PORT=8848

# ==================== RocketMQ 配置 ====================
ROCKETMQ_ENDPOINT=127.0.0.1:8081
ROCKETMQ_NAME_SRV_ADDR=127.0.0.1:9876
ROCKETMQ_ACCESS_KEY=your_access_key      # 可选，如果启用 ACL
ROCKETMQ_SECRET_KEY=your_secret_key      # 可选，如果启用 ACL
ROCKETMQ_BROKER_ADDR=127.0.0.1:10911     # 可选

# ==================== Mailgun 配置 ====================
# 重要：这些配置必须设置，否则邮件发送功能无法使用
MAILGUN_DOMAIN=your-domain.mailgun.org
MAILGUN_API_KEY=key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
MAILGUN_FROM_EMAIL=noreply@yourdomain.com

# ==================== gRPC 配置（可选）====================
# 手动指定服务注册地址（多个地址用逗号分隔）
GRPC_ENDPOINTS=192.168.1.100:14032,192.168.1.101:14032
```

#### Mailgun 配置获取步骤

1. **注册 Mailgun 账号**
   - 访问 [Mailgun 官网](https://www.mailgun.com/)
   - 注册并验证账号

2. **获取 API Key**
   - 登录 [Mailgun 控制台](https://app.mailgun.com/)
   - 进入 "Settings" → "API Keys"
   - 复制 "Private API key"（以 `key-` 开头）
   - 设置到环境变量 `MAILGUN_API_KEY`

3. **配置域名**
   - 进入 "Sending" → "Domains"
   - 选择或添加域名
   - 复制域名（如 `sandbox12345.mailgun.org` 或 `yourdomain.com`）
   - 设置到环境变量 `MAILGUN_DOMAIN`

4. **设置发件人邮箱**
   - 发件人邮箱必须是已验证域名下的邮箱
   - 例如：`noreply@yourdomain.com`
   - 设置到环境变量 `MAILGUN_FROM_EMAIL`

#### 环境变量优先级

配置读取优先级：**环境变量 > 配置文件**

例如，配置文件中写：
```yaml
mailgun:
  domain: "$MAILGUN_DOMAIN"
```

GoFrame 会自动从环境变量中读取 `MAILGUN_DOMAIN` 的值。

#### 开发环境配置

开发环境可以直接修改 `config.yaml` 文件：

```yaml
mailgun:
  domain: "sandbox12345.mailgun.org"        # 直接写死值
  apiKey: "key-xxxxxx"                      # 直接写死值（不推荐）
  fromEmail: "test@sandbox12345.mailgun.org"
```

**注意**：生产环境建议使用环境变量，避免敏感信息泄露。

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

### 1. 环境变量未生效

**问题**：配置了环境变量但服务仍然报错配置未找到

**解决方案**：
```bash
# 1. 确认 .env 文件位置正确（项目根目录）
ls -la .env

# 2. 确认环境变量格式正确（无空格）
# ✅ 正确
MAILGUN_API_KEY=key-xxxx
# ❌ 错误
MAILGUN_API_KEY = key-xxxx

# 3. 重启服务使环境变量生效
make restart

# 4. 验证环境变量是否加载
curl http://localhost:14031/debug/config
```

### 2. 消息发送失败怎么办？

- 检查消息状态：使用 `GetMessageStatus` 接口查询
- 查看错误信息：`error_message` 字段包含详细错误
- 重试发送：使用 `ResendMessage` 接口（最多重试5次）
- 检查日志：查看 `./log/message/` 目录下的日志文件

### 3. Mailgun 配置错误

**常见错误信息**：
- `Mailgun domain 未配置` - 检查 `MAILGUN_DOMAIN` 环境变量
- `Mailgun apiKey 未配置` - 检查 `MAILGUN_API_KEY` 环境变量
- `Mailgun fromEmail 未配置` - 检查 `MAILGUN_FROM_EMAIL` 环境变量
- `Forbidden` - API Key 无效或权限不足
- `Domain not found` - 域名未在 Mailgun 中验证

**排查步骤**：
```bash
# 1. 验证配置是否正确加载
curl http://localhost:14031/debug/config

# 2. 查看 Mailgun 配置
# 输出应该包含 domain、fromEmail、fromName

# 3. 测试 Mailgun 连接
# 发送一条测试消息并查看日志
```

**确保以下配置正确**：
- `MAILGUN_DOMAIN`: Mailgun 验证过的域名（如 `sandbox12345.mailgun.org`）
- `MAILGUN_API_KEY`: 从 Mailgun 控制台获取（以 `key-` 开头）
- `MAILGUN_FROM_EMAIL`: 必须是验证域名下的邮箱

### 4. RocketMQ 连接失败

**常见错误信息**：
- `connect timeout` - 网络不通或地址错误
- `no route to host` - 防火墙阻止或 IP 地址错误
- `authentication failed` - ACL 认证失败

**检查配置**：
- `ROCKETMQ_NAME_SRV_ADDR`: 确保 RocketMQ NameServer 地址正确
- 网络连通性：确保服务能访问 RocketMQ
- 权限：如果启用了 ACL，需要配置 `ROCKETMQ_ACCESS_KEY` 和 `ROCKETMQ_SECRET_KEY`

**测试连接**：
```bash
# 测试 NameServer 连通性
telnet 127.0.0.1 9876

# 查看 RocketMQ 日志
tail -f ./log/queue/*.log
```

### 5. Nacos 注册失败

**常见错误**：
- `connect refused` - Nacos 服务未启动或地址错误
- `service register failed` - 服务注册失败

**解决方案**：
```bash
# 1. 检查 Nacos 服务状态
curl http://127.0.0.1:8848/nacos/v1/console/health/readiness

# 2. 验证 Nacos 配置
echo $NACOS_SERVER_IP
echo $NACOS_SERVER_PORT

# 3. 查看服务注册情况
# 访问 Nacos 控制台：http://127.0.0.1:8848/nacos
# 查看服务列表中是否有 ttpos-message
```

### 6. gRPC 服务无法访问

**问题**：其他服务无法调用 ttpos-message 的 gRPC 接口

**排查步骤**：
```bash
# 1. 检查 gRPC 端口是否监听
netstat -tuln | grep 14032
# 或
lsof -i :14032

# 2. 使用 grpcurl 测试
grpcurl -plaintext localhost:14032 list

# 3. 检查防火墙
sudo firewall-cmd --list-ports

# 4. 查看服务日志
tail -f ./log/message/*.log
```

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
