### 项目介绍

TTPOS 业务中台（ttpos-bmp）是为餐饮零售场景打造的 Go 微服务集群，基于 GoFrame v2.x + gRPC 架构，统一承载管理后台、门店、ERP、外送等域的能力，并通过 Nacos 做服务发现/配置管理，RocketMQ 做异步解耦与延迟消息，配合数据库迁移体系保障各模块独立演进与平滑升级。

### 模块组成
- 管理模块（ttpos-manager）：用户与权限、系统配置、组织与角色等能力
- 门店模块（ttpos-shop）：商品、订单、会员、门店经营相关能力
- ERP 模块（ttpos-erp）：进销存、生产制造、价格体系、仓库与库存
- 外送模块（ttpos-takeout）：第三方外送对接、订单派送、回调处理
- 消息模块（ttpos-message）：统一消息中心（邮件/短信），异步队列发送、模板渲染、状态追踪

上述模块均为独立应用，具备各自的 HTTP/gRPC 服务与部署产物，可单独开发与发布。

### 技术栈
- GoFrame v2.x：应用框架、ORM、代码生成工具链
- gRPC：微服务通讯与远程调用
- Nacos：服务发现与配置管理
- RocketMQ：消息队列（支持延迟消息）
- golang-migrate：数据库迁移与版本管理

### ttpos-message 模块简介
- 定位：为各业务域提供标准化的消息发送能力（gRPC），统一模板管理与发送日志，依赖 RocketMQ 实现异步解耦与失败重试
- 能力：
  - 发送接口：SendMessage（幂等提交，快速返回）
  - 查询接口：GetMessageStatus（按 message_uuid 查询状态）
  - 重发接口：ResendMessage（失败消息重试，限制重试次数）
  - 渠道支持：email（已实现 Mailgun），sms（预留扩展）
  - 模板渲染：支持 {{var}} 变量替换，主题/正文渲染
  - 发送追踪：记录请求/响应与错误日志，便于审计与排错
- 端口（默认，可按环境调整）：http:14041，grpc:14042

### 目录结构概览（节选）
- app/ttpos-erp、app/ttpos-manager、app/ttpos-shop、app/ttpos-takeout：各业务模块
  - api/：接口定义与 gRPC 生成代码
  - internal/：模块内部实现（boot、controller、logic、dao、model、service 等）
  - manifest/：模块配置模板、部署清单、protobuf、sql 迁移脚本
  - main.go、Makefile：模块入口与构建脚本
- internal/pkg：中台通用能力（加解密、队列封装、Nacos、OTLP、缓存、中间件
- Makefile：根级开发与部署命令集合
- README.MD：总览与使用说明

### 启动与本地开发
1) 初始化依赖与配置
```bash
go mod tidy
make conf        # 生成各模块 manifest/config/config.yaml（依赖上级 .env）
```
2) 启动中间件与应用
```bash
make mid         # 启动 Nacos、RocketMQ 等中间件容器
make up          # 构建并启动全部中台服务容器
```
3) 单模块本地运行（便于开发调试）
```bash
make run.manager   # http:14001  grpc:14002
make run.shop      # http:14011  grpc:14012
make run.erp       # http:14021  grpc:14022
make run.takeout   # http:14031  grpc:14032
```
4) 健康检查
```bash
curl http://localhost:1400x/api/v1/[模块名]/hello
```

### 数据库迁移
本项目使用 golang-migrate 管理数据库版本，迁移脚本按模块放在各自的 `app/*/manifest/sql` 目录。
- 新增迁移文件（在对应模块目录执行）：`make db_add NAME=your_change`
- 本地迁移：`make db_up`
- Docker 环境迁移：`make db_up.docker`
- 根目录全量迁移：`make migrate`

ERP 历史数据/结构迁移另见 `app/ttpos-erp/manifest/erp-migrate`，可按需执行：
```bash
make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5
```

### gRPC 开发流程
1) 在 `manifest/protobuf`（或模块内 manifest/protobuf）定义/更新 proto  
2) 执行 `make pb` 或 `gf gen pb` 生成 `api/` 与 `controller` 代码骨架  
3) 在 `internal/controller` 实现服务接口，在 `internal/logic` 编写业务逻辑  
4) 在 `internal/boot` 完成服务注册、Nacos 注册与启动逻辑  

注意：`dao/`、`model/do`、`model/entity`、`service/` 等由工具生成的文件不要手工修改。

### 配置与端口
- 通过 `make conf` 基于根目录 `.env` 生成各模块的 `manifest/config/config.yaml`
- 默认端口（可按环境调整）：
  - manager：14001/http，14002/grpc
  - shop：14011/http，14012/grpc
  - erp：14021/http，14022/grpc
  - takeout：14031/http，14032/grpc
  - message：14041/http，14042/grpc

### 部署指引（简要）
1) 准备 `.env` 并执行 `make conf` 生成配置  
2) 可选：`make migrate` 升级数据库  
3) 构建镜像：
   - `docker compose build bmp-manager|bmp-shop|bmp-erp|bmp-takeout`
   - 或 `docker compose build` 全量构建
4) 确认容器将生成的 `manifest/config/config.yaml` 挂载到容器 `/app/config/config.yaml`
5) 配置网关/反向代理（示例：`docker/nginx/conf.d/ttpos-bmp.conf`）指向各模块 HTTP 端口
6) 启动后通过 `/api/v1/[模块名]/hello` 验证

### 运维与排错
- 日志位置：各模块 `log/` 目录（如 `app/ttpos-erp/log`）
- 队列观测：RocketMQ 的 Topic 与消费日志参考模块 `internal/consts` 与 `internal/pkg/queue`
- 常见问题：
  - Nacos 未注册：检查 `make mid` 是否正常、配置是否生成、容器网络
  - gRPC 连接失败：校验端口暴露、服务是否完成 `internal/boot` 注册启动
  - 迁移失败：确认数据库连接配置、迁移脚本顺序与依赖

### 相关文档
- 根级 `README.MD`：总览与命令说明
- `MIGRATION_QUICK_START.md`：迁移快速上手
- GoFrame 文档：`https://goframe.org.cn`
- 每个模块名称目录包含了模块的详细功能说明, 新增的服务描述记录在 features 目录下 , 修订的记录在 changelog 目录下