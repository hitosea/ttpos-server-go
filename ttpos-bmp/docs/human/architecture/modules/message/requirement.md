# TTPOS 消息中心服务需求文档

## 1. 项目概述

### 1.1 项目简介
ttpos-message 是 TTPOS 中台工程的核心模块之一，使用 GoFrame v2.x 框架开发的 MonoRepoApp。负责统一管理和发送各类消息（邮件、短信等），为其他业务模块提供标准化的消息发送能力。

### 1.2 技术栈
- **框架**: GoFrame v2.x
- **通信协议**: gRPC
- **消息队列**: RocketMQ
- **数据库**: MySQL (独立 schema: `messages`)
- **邮件服务**: Mailgun API
- **工程结构**: 遵循 GoFrame 标准目录结构，与 `app/ttpos-erp` 保持一致

### 1.3 核心特性
- 统一的消息发送接口
- 异步消息处理机制
- 多渠道消息发送支持（邮件、短信）
- 消息发送记录追踪
- 消息模板管理
- 高可用、高并发设计

---

## 2. 功能需求

### 2.1 gRPC 服务接口

#### 2.1.1 SendMessage - 发送消息接口

**接口定义**:
```protobuf
// 发送消息请求
message SendMessageReq {
  string message_uuid = 1;    // 消息唯一标识（由调用方生成，用于幂等性）
  uint64 template_id = 2;     // 消息模板ID
  string message_args = 3;    // 消息参数（JSON格式）
  string message_type = 4;    // 消息类型（email/sms）
  string recipient = 5;       // 接收人（邮箱地址或手机号）
  string subject = 6;         // 消息主题（可选，邮件使用）
  string company_uuid = 7;    // 公司UUID
  string operator_uuid = 8;   // 操作人UUID
}

// 发送消息响应
message SendMessageResp {
  bool success = 1;           // 是否成功提交
  string message = 2;         // 响应消息
  string message_uuid = 3;    // 消息UUID
}
```

**功能说明**:
- 接收消息发送请求，进行参数验证
- 创建消息记录（状态：待发送）
- 将消息发送任务提交到 RocketMQ
- 返回提交结果（不等待实际发送完成）
- 支持幂等性（通过 `message_uuid` 去重）

#### 2.1.2 GetMessageStatus - 查询消息状态接口

**接口定义**:
```protobuf
// 查询消息状态请求
message GetMessageStatusReq {
  string message_uuid = 1;    // 消息UUID
}

// 查询消息状态响应
message GetMessageStatusResp {
  bool success = 1;
  string message = 2;
  MessageInfo message_info = 3;
}

// 消息信息
message MessageInfo {
  uint64 message_id = 1;        // 消息ID
  string message_uuid = 2;      // 消息UUID
  uint64 template_id = 3;       // 模板ID
  string message_type = 4;      // 消息类型
  string recipient = 5;         // 接收人
  int32 status = 6;             // 消息状态（0-待发送，1-发送中，2-发送成功，3-发送失败）
  string error_message = 7;     // 错误信息
  int64 created_at = 8;         // 创建时间
  int64 send_time = 9;          // 发送时间
}
```

**功能说明**:
- 根据 `message_uuid` 查询消息发送状态
- 返回消息的详细信息和当前状态

#### 2.1.3 ResendMessage - 重发消息接口

**接口定义**:
```protobuf
// 重发消息请求
message ResendMessageReq {
  string message_uuid = 1;    // 消息UUID
}

// 重发消息响应
message ResendMessageResp {
  bool success = 1;
  string message = 2;
}
```

**功能说明**:
- 针对发送失败的消息，支持重新发送
- 更新消息状态并重新提交到 RocketMQ

---

### 2.2 邮件发送功能

#### 2.2.1 Mailgun 集成
- 使用 Mailgun API 发送邮件
- 支持 HTML 和纯文本格式
- 支持附件发送（预留功能）
- 支持发送状态回调（webhook）

#### 2.2.2 邮件配置
- Mailgun API Key 配置
- Mailgun Domain 配置
- 发件人地址配置
- 通过配置文件管理，支持环境变量

#### 2.2.3 邮件模板
- 支持模板变量替换（基于 `message_args`）
- 模板内容存储在数据库
- 支持 HTML 模板渲染

---

### 2.3 消息队列处理

#### 2.3.1 RocketMQ 生产者
- 接收到 gRPC 请求后，将消息发送任务提交到 RocketMQ
- Topic: `message-send-topic`
- Tag: `email`, `sms`
- 消息体包含完整的消息信息

#### 2.3.2 RocketMQ 消费者
- 订阅 `message-send-topic`
- 根据 Tag 区分消息类型
- 并发消费，支持失败重试
- 消费逻辑：
  1. 更新消息状态为"发送中"
  2. 根据消息类型调用对应的发送服务
  3. 更新消息状态（成功/失败）
  4. 记录发送时间和错误信息

#### 2.3.3 失败重试机制
- 支持自动重试（RocketMQ 重试机制）
- 最大重试次数：3次
- 重试间隔：递增（1分钟、5分钟、15分钟）
- 超过重试次数后标记为失败

---

### 2.4 消息模板管理

#### 2.4.1 模板表结构
```sql
CREATE TABLE `message_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '模板UUID',
  `template_name` varchar(100) NOT NULL COMMENT '模板名称',
  `template_type` varchar(20) NOT NULL COMMENT '模板类型(email/sms)',
  `template_subject` varchar(200) DEFAULT NULL COMMENT '模板主题(邮件用)',
  `template_content` text NOT NULL COMMENT '模板内容(支持变量)',
  `template_args` json DEFAULT NULL COMMENT '模板参数定义',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态(0-禁用,1-启用)',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` int NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_template_type` (`template_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息模板表';
```

#### 2.4.2 模板变量
- 模板内容支持占位符，如：`{{username}}`, `{{order_no}}`
- `message_args` 为 JSON 格式，包含所有变量值
- 发送时进行变量替换

---

## 3. 数据库设计

### 3.1 消息记录表

```sql
CREATE TABLE `message_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '消息UUID',
  `template_id` bigint unsigned NOT NULL COMMENT '模板ID',
  `message_type` varchar(20) NOT NULL COMMENT '消息类型(email/sms)',
  `recipient` varchar(100) NOT NULL COMMENT '接收人',
  `subject` varchar(200) DEFAULT NULL COMMENT '消息主题',
  `content` text COMMENT '消息内容(渲染后)',
  `message_args` json DEFAULT NULL COMMENT '消息参数',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态(0-待发送,1-发送中,2-发送成功,3-发送失败)',
  `error_message` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT '重试次数',
  `company_uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '公司UUID',
  `operator_uuid` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人UUID',
  `send_time` int NOT NULL DEFAULT 0 COMMENT '发送时间',
  `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` int NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` int NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_template_id` (`template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息记录表';
```

### 3.2 消息发送日志表

```sql
CREATE TABLE `message_send_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `message_uuid` varchar(64) NOT NULL COMMENT '消息UUID',
  `send_time` int NOT NULL COMMENT '发送时间',
  `send_result` tinyint NOT NULL COMMENT '发送结果(0-失败,1-成功)',
  `error_message` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `request_data` json DEFAULT NULL COMMENT '请求数据',
  `response_data` json DEFAULT NULL COMMENT '响应数据',
  `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_message_uuid` (`message_uuid`),
  KEY `idx_send_time` (`send_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息发送日志表';
```

---

## 4. 技术架构

### 4.1 架构图

```
┌─────────────┐
│  其他业务服务 │
│ (ttpos-erp) │
└──────┬──────┘
       │ gRPC
       ▼
┌─────────────────────┐
│  ttpos-message      │
│  ┌───────────────┐  │
│  │ gRPC Server   │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼────────┐ │
│  │ Logic Layer    │ │
│  └───────┬────────┘ │
│          │          │
│  ┌───────▼────────┐ │
│  │ RocketMQ       │ │
│  │ Producer       │ │
│  └───────┬────────┘ │
└──────────┼──────────┘
           │
           ▼
    ┌──────────────┐
    │  RocketMQ    │
    │  Broker      │
    └──────┬───────┘
           │
           ▼
┌─────────────────────┐
│  ttpos-message      │
│  ┌───────────────┐  │
│  │ RocketMQ      │  │
│  │ Consumer      │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼────────┐ │
│  │ Message        │ │
│  │ Sender         │ │
│  └───────┬────────┘ │
│          │          │
│  ┌───────▼────────┐ │
│  │ Mailgun API    │ │
│  └────────────────┘ │
└─────────────────────┘
```

### 4.2 核心组件

#### 4.2.1 gRPC 服务层
- 位置：`internal/controller/rpc/message/`
- 职责：接收外部请求，参数验证，调用业务逻辑层

#### 4.2.2 业务逻辑层
- 位置：`internal/logic/message/`
- 职责：
  - 消息发送逻辑
  - 消息状态管理
  - 模板渲染
  - 消息队列集成

#### 4.2.3 数据访问层
- 位置：`internal/dao/`（自动生成）
- 职责：数据库操作

#### 4.2.4 消息发送服务
- 位置：`internal/logic/sender/`
- 职责：
  - 邮件发送（Mailgun）
  - 短信发送（预留）

#### 4.2.5 消息队列服务
- 位置：`internal/logic/queue/`
- 职责：
  - RocketMQ 生产者
  - RocketMQ 消费者

---

## 5. 配置管理

### 5.1 配置文件示例

```yaml
# manifest/config/config.tpl.yaml
server:
  address: ":8084"
  logPath: "./log"

# gRPC 服务配置
grpc:
  name: "ttpos-message"
  address: ":9084"
  logPath: "./log"

# 数据库配置
database:
  default:
    link: "mysql:$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/messages"
    maxIdle: 10
    maxOpen: 100
    maxLifetime: 30

# RocketMQ 配置
rocketmq:
  nameServers:
    - "$ROCKETMQ_HOST:9876"
  groupName: "ttpos-message-group"
  topic: "message-send-topic"
  retryTimes: 3

# Mailgun 配置
mailgun:
  domain: "$MAILGUN_DOMAIN"
  apiKey: "$MAILGUN_API_KEY"
  fromEmail: "$MAILGUN_FROM_EMAIL"
  fromName: "TTPOS System"

# Nacos 配置
nacos:
  serverAddr: "$NACOS_SERVER_ADDR"
  namespace: "$NACOS_NAMESPACE"
  group: "DEFAULT_GROUP"
  dataId: "ttpos-message.yaml"

# 日志配置
logger:
  level: "all"
  stdout: true
```

---

## 6. 开发计划

### 6.1 Phase 1 - 基础架构搭建
- [ ] 创建项目目录结构
- [ ] 配置文件模板
- [ ] 数据库迁移脚本
- [ ] gRPC protobuf 定义
- [ ] 基础工具类

### 6.2 Phase 2 - 核心功能开发
- [ ] gRPC 服务实现
  - [ ] SendMessage 接口
  - [ ] GetMessageStatus 接口
  - [ ] ResendMessage 接口
- [ ] 数据库 DAO 层生成
- [ ] 业务逻辑层实现
- [ ] Mailgun 邮件发送服务

### 6.3 Phase 3 - 消息队列集成
- [ ] RocketMQ 生产者实现
- [ ] RocketMQ 消费者实现
- [ ] 异步发送流程
- [ ] 失败重试机制

### 6.4 Phase 4 - 完善功能
- [ ] 消息模板管理
- [ ] 模板变量渲染
- [ ] 发送日志记录
- [ ] 监控和告警

### 6.5 Phase 5 - 测试与部署
- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能测试
- [ ] Docker 镜像构建
- [ ] 部署文档

---

## 7. 非功能性需求

### 7.1 性能要求
- gRPC 接口响应时间 < 100ms（不含实际发送）
- 支持 1000 TPS 消息提交
- 消息发送成功率 > 99%

### 7.2 可用性要求
- 服务可用性 > 99.9%
- 支持水平扩展
- 消息队列支持持久化

### 7.3 安全要求
- API Key 加密存储
- 敏感信息不记录日志
- 支持访问权限控制

### 7.4 可维护性
- 完善的日志记录
- 清晰的错误信息
- 代码注释完整
- 遵循 GoFrame 开发规范

---

## 8. 依赖服务

### 8.1 必需服务
- MySQL 数据库
- RocketMQ 消息队列
- Nacos 服务注册与配置中心

### 8.2 外部服务
- Mailgun API（邮件发送）

---

## 9. 接口调用示例

### 9.1 发送邮件

```go
// 调用方代码示例
req := &message.SendMessageReq{
    MessageUuid: guuid.S(),
    TemplateId:  1001,
    MessageArgs: `{"username":"张三","order_no":"ORD20250101001"}`,
    MessageType:  "email",
    Recipient:    "zhangsan@example.com",
    Subject:      "订单确认通知",
    CompanyUuid:  "company-uuid-100",
    OperatorUuid: "operator-uuid-1",
}

resp, err := messageClient.SendMessage(ctx, req)
if err != nil {
    g.Log().Error(ctx, "发送消息失败", err)
    return
}

g.Log().Info(ctx, "消息提交成功", resp.MessageUuid)
```

### 9.2 查询消息状态

```go
req := &message.GetMessageStatusReq{
    MessageUuid: "message-uuid-here",
}

resp, err := messageClient.GetMessageStatus(ctx, req)
if err != nil {
    g.Log().Error(ctx, "查询消息状态失败", err)
    return
}

g.Log().Info(ctx, "消息状态", resp.MessageInfo.Status)
```

---

## 10. 附录

### 10.1 消息状态说明
- `0` - 待发送：消息已创建，等待发送
- `1` - 发送中：消息正在发送
- `2` - 发送成功：消息已成功发送
- `3` - 发送失败：消息发送失败

### 10.2 消息类型说明
- `email` - 邮件消息
- `sms` - 短信消息（预留）

### 10.3 Mailgun API 文档
- 官方文档：https://documentation.mailgun.com/
- API 参考：https://documentation.mailgun.com/en/latest/api-intro.html
