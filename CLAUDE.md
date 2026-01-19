# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

TTPOS 是一个现代化的餐饮收银系统后端，支持多终端（pos/shop/kds/qds/assistant/tablet/mobile/menu/member）和业务场景（点餐、桌位管理、会员、外卖、报表等）。采用微服务架构设计，包含以下模块：

- **Main 模块** (`main/`)：Go 1.23+ + Gin 核心业务服务
- **BMP 模块** (`ttpos-bmp/`)：Go 1.23+ + GoFrame 2.x 业务中台微服务
- **Shared API** (`ttpos-api/`)：gRPC/HTTP 接口定义和共享类型
- **Admin 模块** (`admin/`)：**Legacy** - PHP 8.0+ + ThinkPHP 6.0（仅用于修复旧功能或运维）
- **Vue Admin** (`admin/views/`)：**Legacy** - Vue 3 + TypeScript + Vite + Element Plus

> **注意**：前端实现已迁移至独立仓库 `ttpos-flutter`。PHP Admin 和内置 Vue3 Admin 仅作为遗留模块，不进行新功能开发。

## 技术栈

| 模块 | 技术栈 |
|------|--------|
| Go Main | Go 1.23+, Gin, GORM, MySQL 8.0+, Redis 6.0+, RocketMQ, OpenTelemetry, Swagger |
| Go BMP | Go 1.23+, GoFrame v2.x, MySQL, Redis, gRPC, Nacos, OTel |
| Legacy PHP | PHP 8.0+, ThinkPHP 6.0 |
| Legacy Vue | Vue 3, TypeScript, Vite, Element Plus |
| 基础设施 | Docker, Nginx, Redis Cluster, MySQL 主从, SkyWalking, Prometheus + Grafana |

## 跨仓库协作

当需要查阅前端代码、接口定义、交互实现或跨仓库引用时：

1. 读取根目录下的 `.agents` 文件
2. 获取 `FRONTEND_PATH` 等配置的绝对路径
3. 使用 `read_file` 或 `list_dir` 或其他 MCP 服务的工具访问该绝对路径下的文件

## 常用命令

### Main 模块 (main/)
```bash
cd main && go run main.go           # 启动主服务
cd main && go test ./...            # 运行所有测试
cd main && go test ./app/service    # 运行指定包测试
cd main && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out  # 测试覆盖率
cd main && go mod tidy              # 整理依赖
cd main && go fmt ./...             # 格式化代码
cd main && go vet ./...             # 静态检查
cd main && swag init                # 生成 Swagger 文档（需先安装 swag）
```

### BMP 模块 (ttpos-bmp/)
```bash
cd ttpos-bmp && make help           # 查看所有可用命令
cd ttpos-bmp && make conf           # 生成配置文件
cd ttpos-bmp && make mid            # 启动中间件（Nacos、RocketMQ）
cd ttpos-bmp && make run.erp        # 启动 ERP 服务 (14021:http 14022:grpc)
cd ttpos-bmp && make run.takeout    # 启动外送服务 (14031:http 14032:grpc)
cd ttpos-bmp && make run.message    # 启动消息服务 (14041:http 14042:grpc)
cd ttpos-bmp && make run.websocket  # 启动 WebSocket 服务 (14051:http 14052:grpc)
cd ttpos-bmp && make migrate        # 执行数据库迁移
cd ttpos-bmp && make up             # 构建并启动 BMP 服务容器
```

### BMP 子模块代码生成 (ttpos-bmp/app/ttpos-xxx/)
```bash
make dao      # 生成 DAO/DO/Entity 数据访问层代码
make ctrl     # 解析 API 并生成控制器/SDK 代码
make service  # 生成服务接口代码
make pb       # 解析 protobuf 并生成 Go 代码
make build    # 编译构建二进制文件
make run      # 运行服务
```

### Admin 模块 (admin/) - Legacy
```bash
cd admin && php think run           # 启动 PHP 服务
cd admin && php think migrate:run   # 运行数据库迁移
cd admin && php think migrate:create CreateTableName  # 创建迁移文件
cd admin && php think unit          # 运行单元测试
```

### Docker 部署
```bash
docker-compose -f docker-compose.dev.yml up -d       # 启动开发环境
docker-compose -f docker-compose.dev.redis.yml up -d # 启动 Redis 集群
docker-compose -f docker-compose.production.yml up -d # 生产环境部署
```

## 代码架构

### Main 模块分层架构 (main/app/)
```
api/v1/         # API 层：路由处理，只负责参数校验和调用 Service
service/        # Service 层：业务逻辑，接口定义和实现在同一文件
repository/     # Repository 层：数据访问，使用选项模式
model/          # 数据模型
dto/req/        # 请求参数对象
dto/resp/       # 响应数据对象
constant/       # 常量定义
errors/         # 错误定义
event/          # 事件处理
queue/          # 队列处理
tasks/          # 定时任务
modules/        # 功能模块
cloud/          # 云服务集成
```

### Main 公共包 (main/pkg/)
```
database/       # 数据库管理（DBManager）
cache/          # Redis 缓存
eventbus/       # 事件总线
lock/           # 分布式锁
language/       # 多语言（i18n）
logger/         # 日志
auth/           # 认证
context/        # 上下文扩展
validator/      # 参数校验
utils/          # 工具函数（包含 utils.Go 协程方法）
rocketmq/       # RocketMQ 集成
websocket/      # WebSocket
nacos/          # Nacos 配置中心
otlp/           # OpenTelemetry 追踪
storage/        # 存储服务
sms/            # 短信服务
encrypt/        # 加密工具
```

### BMP 模块结构 (ttpos-bmp/)
```
app/
  ttpos-erp/        # ERP 服务
  ttpos-takeout/    # 外送服务
  ttpos-message/    # 消息服务
  ttpos-websocket/  # WebSocket 服务
  ttpos-shop/       # 商户服务
  ttpos-manager/    # 管理服务

每个子模块结构：
  api/                    # API 接口定义（protobuf）
  internal/controller/    # 控制器（http/ 和 rpc/）
  internal/logic/         # 业务逻辑实现（手动编写）
  internal/dao/           # 数据访问层（自动生成，禁止修改）
  internal/model/entity/  # 数据实体（自动生成，禁止修改）
  internal/model/do/      # 数据对象（自动生成，禁止修改）
  internal/model/dto/     # 数据传输对象（手动编写）
  internal/consts/        # 常量定义
  manifest/sql/           # 数据库迁移脚本
  manifest/protobuf/      # protobuf 定义
```

### 服务通信
- Main ↔ BMP：gRPC
- Admin ↔ Main：HTTP API
- 跨服务异步：RocketMQ
- 实时推送：WebSocket
- 配置管理：Nacos

## 核心编码规范

### 命名规范
| 类型 | 规则 | 示例 |
|------|------|------|
| URL | snake_case | `/api/v1/order_info` |
| 结构体 | 大驼峰，ID 字段大写 | `StaffId`, `OrderUuid` |
| 接口 | `I` 开头 | `IOrderSrv`, `IUserRepo` |
| 包名/文件名 | snake_case | `member_service.go` |
| 外键 UUID 字段 | 完整表名_uuid | `product_uuid`（非 `prod_uuid`）|

### API 响应格式
```json
{
  "code": 1,
  "message": "success",
  "data": {}  // 必须是对象，不能是 null 或数组
}
```

**切片初始化规范**：响应体中的切片必须使用 `make` 初始化，避免返回 null。

### HTTP 方法规范
| 方法 | 用途 | 参数解析 | Req Tag |
|------|------|----------|---------|
| GET | 获取信息（查询、列表、详情） | `ShouldBindQuery` | `form` |
| POST | 创建/修改数据 | `ShouldBindJSON` | `json` |
| DELETE | 删除数据 | `ShouldBindJSON` | `json` |
| PUT | **禁止使用** | - | - |

### 分层依赖规则
- API 层只能调用 Service，**严禁引用 repository 包**
- Service 可依赖其他 Service，只能依赖自己领域的 Repository
- Repository 只持有 db 实例，不持有 DBManager
- **禁止跨层调用和循环依赖**

### Go Main 关键约束
- 使用 `any` 替代 `interface{}`
- 使用 `return error`，不使用 `panic`
- 协程必须使用 `utils.Go` 方法（内置 recover）
- 多语言字段必须使用 `dto.LocaleResponse` 结构，字段名使用 `LocaleName` 或 `LocaleXXXName` 格式
- 常量必须在 `constant` 包中定义
- Service 接口定义（`I{Name}Srv`）和实现（`{Name}Srv`）必须在同一文件
- 通过 `ctx.GetDB()` 获取数据库连接
- 数据库操作必须通过 Repository 层，禁止直接使用 GORM 方法
- 多个数据库操作必须使用事务包裹（`repository.CommonRepo.Transaction`）
- **所有日志必须包含 `company_uuid` 字段**

### 事务使用规范
```go
// 正确示例
err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    orderRepo := repository.NewOrderRepo(tx)  // 使用 tx 创建 Repository
    // ... 数据库操作
    return nil  // 自动提交
})
// ❌ 禁止在事务中手动调用 tx.Commit() 或 tx.Rollback()
// ❌ 禁止在事务中使用 db 而不是 tx 创建 Repository
```

### Go BMP 关键约束
- **禁止修改自动生成文件**：`dao/`, `model/entity/`, `model/do/`, `service/`
- 使用 `gerror` 处理错误（不用标准库 errors）
- 业务逻辑写在 `internal/logic/` 目录
- 使用 `dao` 层操作数据库
- 事务使用 `g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error { ... })`

### 数据库规范

#### 多租户架构
每个商户一个独立数据库（如 `shop8267304538112000`），业务表无需 `company_uuid` 字段。

#### 表设计规范
- 表名使用 `ttpos_` 前缀
- **迁移文件中表名不要带 `ttpos_` 前缀**
- 每表必需字段：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- 时间字段用 `int`（Unix 时间戳）
- 金额字段用 `decimal`
- 布尔字段用 `int`（0/1），禁止使用 boolean

#### 迁移文件 TARGET 常量
| 值 | 说明 |
|---|---|
| `const TARGET = 'all';` | 应用到所有商户数据库和 saas 主库（company/company_setting 表必须）|
| `const TARGET = 'main';` | 仅应用到 saas 主库 |
| 不设置 | 仅应用到所有商户数据库（普通业务表）|

#### 多语言表规范
需要多语言的业务表必须包含两个字段：
- `name` VARCHAR(1000) - 存储 JSON 格式的多语言数据
- `multi_language_name_uuid` BIGINT - 关联 `ttpos_multi_language_name` 表

**创建迁移文件后必须同步更新**：
- `admin/database/seeds/shop_01.sql`
- `main/app/model/` 对应模型文件

### Git 提交规范
```
<type>(<scope>): <subject>
```
类型：feat, fix, docs, style, refactor, perf, test, build, ci, chore

## 开发检查清单

### 编码时
- [ ] URL 使用 snake_case，使用名词单数 + 操作后缀
- [ ] 结构体 ID 字段大写
- [ ] 使用 `any` 而不是 `interface{}`
- [ ] Service 接口和实现在同一文件
- [ ] 通过 `ctx.GetDB()` 获取数据库连接
- [ ] 数据库操作通过 Repository 层
- [ ] 多表操作使用事务包裹
- [ ] 日志包含 `company_uuid`
- [ ] 协程使用 `utils.Go`
- [ ] 多语言字段使用 `dto.LocaleResponse`
- [ ] API 层不引用 repository 包
- [ ] 响应切片使用 `make` 初始化
- [ ] GET 用 `ShouldBindQuery`，POST/DELETE 用 `ShouldBindJSON`

### 提交前
- [ ] `go mod tidy` 整理依赖
- [ ] `go fmt ./...` 格式化代码
- [ ] `go vet ./...` 静态检查
- [ ] 测试通过
- [ ] 迁移文件已同步更新 `shop_01.sql`

## 文档索引

### 规则文件 (.cursor/rules/)
| 文件 | 说明 |
|------|------|
| `go-main.mdc` | Go Main 核心约束 |
| `go-bmp.mdc` | Go BMP 模块规范 |
| `database.mdc` | 数据库开发规范 |
| `api.mdc` | API 设计规范 |
| `php.mdc` | PHP 开发规范（Legacy）|
| `vue.mdc` | Vue 开发规范（Legacy）|
| `security.mdc` | 安全开发规范 |
| `version.mdc` | 版本提交规范 |
| `workflows.mdc` | 工作流导航 |

### 文档目录 (docs/)
| 目录 | 说明 |
|------|------|
| `docs/human/architecture/` | 架构设计文档 |
| `docs/human/guides/` | 详细开发指南 |
| `docs/shared/specs/` | 需求规格文档 |
| `docs/shared/api/` | API 定义文档 |
| `docs/agent/workflows/` | Agent 工作流程 |

### BMP 专用规范
| 文件 | 说明 |
|------|------|
| `ttpos-bmp/.cursor/rules/go-rules.mdc` | Go 代码规范 |
| `ttpos-bmp/.cursor/rules/proto-rules.mdc` | Protobuf 规范 |
