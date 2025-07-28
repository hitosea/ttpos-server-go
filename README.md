# TTPOS 收银系统服务端

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-6.0+-DC382D?style=flat&logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)

**现代化的餐饮收银系统后端服务**

支持多端点餐、桌台管理、会员服务、外送配送等完整的餐饮业务场景

</div>

## 🏗️ 项目架构

### 系统概述
TTPOS是一套完整的餐饮收银解决方案，采用微服务架构设计，支持：
- 💰 **收银管理**：订单处理、支付结算、打印小票
- 🍽️ **桌台服务**：开台、点餐、转台、合台
- 👥 **会员系统**：会员管理、积分、优惠券
- 🚚 **外送服务**：在线订餐、配送管理
- 📱 **多端支持**：收银机、平板、H5扫码点餐
- 🏪 **多店管理**：支持连锁店统一管理

### 技术栈
- **后端框架**：Gin + GORM
- **数据库**：MySQL 8.0+ / SQLite
- **缓存**：Redis 6.0+
- **消息队列**：Redis + DelayQueue
- **微服务**：gRPC (外送服务)
- **实时通信**：WebSocket
- **API文档**：Swagger
- **监控**：SkyWalking
- **容器化**：Docker + Docker Compose

### 模块结构
```
ttpos-server-go/
├── main/                    # 主要业务服务
│   ├── app/                # 应用层
│   │   ├── controller/     # 控制器
│   │   ├── service/        # 业务逻辑
│   │   ├── repository/     # 数据访问
│   │   ├── model/          # 数据模型
│   │   ├── dto/            # 数据传输对象
│   │   ├── constant/       # 常量定义
│   │   └── errors/         # 错误处理
│   ├── config/             # 配置管理
│   ├── pkg/                # 基础设施
│   ├── middleware/         # 中间件
│   └── router/             # 路由配置
├── websocket/              # WebSocket服务
├── ttpos-bmp/                # 业务中台服务
├── redis-proxy/            # Redis代理服务
└── admin/                  # 管理后台（PHP）
```

## 🚀 快速开始

### 环境要求
- Go 1.23+
- MySQL 8.0+ 或 SQLite
- Redis 6.0+
- Docker & Docker Compose (可选)

### 安装部署

#### 1. 克隆项目
```bash
git clone https://github.com/your-org/ttpos-server-go.git
cd ttpos-server-go
```

#### 2. 配置环境变量
```bash
# 复制环境配置文件
cp .env.example .env

# 编辑配置文件
vim .env
```

主要配置项：
```env
# 数据库配置
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_DATABASE=ttpos

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 服务配置
SERVER_PORT=8080
SERVER_MODE=debug
JWT_SECRET=your-secret-key
```

#### 3. 安装依赖
```bash
# 进入主服务目录
cd main
go mod tidy

# 安装其他服务依赖
cd ../websocket && go mod tidy
cd ../takeout && go mod tidy
cd ../redis-proxy && go mod tidy
```

#### 4. 数据库迁移
```bash
# 运行迁移
cd admin
php think migrate:run
```

#### 5. 启动服务

**方式一：直接运行**
```bash
# 启动主服务
cd main && go run main.go

# 启动WebSocket服务
cd websocket && go run main.go

# 启动外送服务
cd takeout && go run main.go

# 启动Redis代理
cd redis-proxy && go run main.go
```

**方式二：Docker Compose**
```bash
# 一键启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 访问服务
- **API文档**: http://localhost:8080/swagger/index.html
- **主服务**: http://localhost:8080
- **WebSocket**: ws://localhost:8099/ws
- **外送服务**: localhost:14031 (gRPC)

## 📊 服务端口

| 服务 | 端口 | 协议 | 描述 |
|------|------|------|------|
| 主服务 | 8080 | HTTP | 核心业务API |
| WebSocket | 8099 | WebSocket | 实时通信 |
| 外送服务 | 14031 | gRPC | 外送业务 |
| Redis代理 | 16379 | Redis | Redis代理 |

## 🛠️ 开发指南

### 项目规范

#### 1. 命名规范
```go
// ✅ 结构体：除了ID大写，其他驼峰命名
type Staff struct {
    StaffId   uint64  // ID字段大写
    StaffName string  // 其他字段驼峰命名
}

// ✅ URL：使用snake_case
"/api/v1/passport/server_public_key"

// ✅ 包名和文件名：使用snake_case
package member_service  // member_service.go
```

#### 2. API响应格式
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [{"foo": "bar"}],
    "options": {
      "list": [{"key": "k1", "value": "v1"}]
    },
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

#### 3. 错误处理
```go
// ✅ 返回错误，不使用panic
func GetUser(id uint64) (*User, error) {
    if id == 0 {
        return nil, errors.New("用户ID不能为空")
    }
    return user, nil
}
```

#### 4. 包导入顺序
```go
import (
    // 标准库
    "context"
    "fmt"
    
    // 第三方包
    "github.com/gin-gonic/gin"
    
    // 项目包
    "ttpos-server-go/app/model"
)
```

### 架构模式

#### 1. 服务层设计
```go
// 接口定义
type IOrderSrv interface {
    CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.OrderResp, error)
}

// 实现
type orderSrv struct {
    dbm        *database.DBManager
    memberSrv  IMemberSrv
}

func NewOrderSrv(dbm *database.DBManager, memberSrv IMemberSrv) IOrderSrv {
    return &orderSrv{dbm: dbm, memberSrv: memberSrv}
}
```

#### 2. Repository层设计
```go
type IProductRepo interface {
    GetProduct(opts ...DBOption) (model.Product, error)
}

type productRepo struct {
    db *gorm.DB  // 只传入db实例，不传入dbm实例
}
```

#### 3. 事件总线使用

**定义事件**（在`pkg/eventbus/event/`目录）：
```go
const EventOrderCreated EventName = "Event_Order_Created"

type OrderCreatedPayload struct {
    BasePayload
    OrderId uint64 `json:"order_id"`
}

func (system *SystemEventBus) PublishOrderCreatedEvent(msg OrderCreatedPayload) {
    system.bus.Publish(eventbus.Event{Name: string(EventOrderCreated), Payload: msg})
}
```

**发布事件**：
```go
event.NewSystemBus().PublishOrderCreatedEvent(event.OrderCreatedPayload{
    OrderId: orderId,
})
```

**订阅事件**（在`app/event/`目录）：
```go
event.NewSystemBus().SubscribeOrderCreatedEvent(func(msg event.OrderCreatedPayload) {
    // 处理逻辑
})
```

#### 4. 并发控制
```go
type Service struct {
    systemLock lock.Lock
}

func (s *Service) OpenDesk(deskUuid uint64) error {
    s.systemLock.LockUuid(deskUuid)
    defer s.systemLock.UnlockUuid(deskUuid)
    
    // 业务逻辑
    return nil
}
```

### API文档
使用 [swaggo/swag](https://github.com/swaggo/swag) 生成API文档：

```bash
# 安装swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
cd main && swag init

# 访问文档
open http://localhost:8080/swagger/index.html
```

### 性能要求
- **API响应时间**：本地环境 < 200ms
- **数据库查询**：使用索引、预加载、分页
- **并发处理**：使用UUID锁控制并发

## 🗄️ 数据库

### 支持的数据库
- **MySQL 8.0+**：生产环境推荐
- **SQLite**：开发测试环境

### 迁移管理
```bash
# 运行迁移
cd admin
php think migrate:run

# 创建新迁移
php think migrate:create CreateUsersTable

# 回滚迁移
php think migrate:rollback
```

### 数据库规范
```go
// 模型定义规范
// 商品分类表 ttpos_product_category
type ProductCategory struct {
    ID   uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
    Name string `gorm:"column:name;type:varchar(100);comment:'分类名称'"`
}
```

## 🔧 配置说明

### 环境变量配置
| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `DB_TYPE` | 数据库类型 | mysql | ✅ |
| `DB_HOST` | 数据库主机 | localhost | ✅ |
| `DB_PORT` | 数据库端口 | 3306 | ✅ |
| `DB_USER` | 数据库用户 | root | ✅ |
| `DB_PASSWORD` | 数据库密码 | - | ✅ |
| `REDIS_HOST` | Redis主机 | localhost | ✅ |
| `JWT_SECRET` | JWT密钥 | - | ✅ |
| `SERVER_MODE` | 运行模式 | debug | ❌ |

### 配置文件结构
```go
type Config struct {
    Server   ServerConf   // 服务配置
    Database DatabaseConf // 数据库配置
    Redis    RedisConf    // Redis配置
    JWT      JWTConf      // JWT配置
    Log      LogConf      // 日志配置
}
```

## 🚀 部署指南

### Docker部署
```bash
# 构建镜像
docker build -t ttpos-server-go .

# 运行容器
docker run -d \
  --name ttpos-server \
  -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  ttpos-server-go
```

### Docker Compose
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: ttpos
    ports:
      - "3306:3306"

  redis:
    image: redis:6.2-alpine
    ports:
      - "6379:6379"
```

### 生产环境部署

#### 1. 构建优化
```bash
# 构建生产版本
./scripts/build.sh

# 设置版本信息
VERSION=v1.0.0 COMMIT_SHA=$(git rev-parse HEAD) ./scripts/build.sh
```

#### 2. 服务监控
- **健康检查**：`GET /health`
- **指标监控**：SkyWalking Agent
- **日志管理**：结构化日志 + 日志轮转

#### 3. 高可用部署
- **负载均衡**：Nginx + 多实例部署
- **数据库**：MySQL主从 + Redis Cluster
- **容器编排**：Kubernetes + Helm Charts

## 📝 API文档

### 核心接口

#### 认证接口
```
POST /api/v1/auth/login      # 用户登录
POST /api/v1/auth/logout     # 用户登出
POST /api/v1/auth/refresh    # 刷新Token
```

#### 订单接口
```
POST /api/v1/order/create           # 创建订单
GET  /api/v1/order/list             # 订单列表
GET  /api/v1/order/{id}             # 订单详情
PUT  /api/v1/order/{id}/cancel      # 取消订单
```

#### 桌台接口
```
POST /api/v1/desk/open              # 开台
PUT  /api/v1/desk/{id}/close        # 关台
PUT  /api/v1/desk/{id}/change       # 转台
```

### 接口规范
- **请求格式**：JSON
- **响应格式**：统一JSON结构
- **认证方式**：JWT Token
- **错误处理**：标准错误码

完整API文档：http://localhost:8080/swagger/index.html

## 🧪 测试

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行指定包测试
go test ./app/service

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 集成测试
```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行集成测试
go test -tags=integration ./tests/...
```

### API测试
推荐使用 Postman 或 Insomnia 导入API文档进行测试。

## 📊 监控运维

### 日志管理
- **日志级别**：DEBUG、INFO、WARN、ERROR
- **日志格式**：JSON结构化日志
- **日志轮转**：按大小和时间自动轮转
- **日志位置**：`./log/` 目录

### 性能监控
- **SkyWalking**：分布式链路追踪
- **Prometheus**：指标收集
- **Grafana**：可视化监控面板

### 健康检查
```bash
# 服务健康检查
curl http://localhost:8080/health

# 数据库连接检查
curl http://localhost:8080/health/db

# Redis连接检查
curl http://localhost:8080/health/redis
```

## 🔒 安全说明

### 认证授权
- **JWT认证**：无状态Token认证
- **权限控制**：基于角色的访问控制(RBAC)
- **接口加密**：支持RSA请求加密

### 数据安全
- **SQL注入防护**：使用参数化查询
- **XSS防护**：输入输出过滤
- **敏感数据**：密码加密存储

### 网络安全
- **HTTPS**：生产环境强制HTTPS
- **CORS**：跨域请求控制
- **限流**：API频率限制

## 🤝 贡献指南

### 代码提交规范
```bash
# 提交格式
<type>(<scope>): <subject>

# 示例
feat(order): 添加订单取消功能
fix(payment): 修复支付回调问题
docs(readme): 更新部署文档
```

### 开发流程
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

### 代码检查清单
- [ ] 命名规范符合项目要求
- [ ] API响应格式统一
- [ ] 错误处理使用return error
- [ ] 添加必要的单元测试
- [ ] 更新相关文档
- [ ] 通过所有CI检查

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源协议。

## 📞 联系方式

- **项目主页**：https://github.com/your-org/ttpos-server-go
- **问题反馈**：https://github.com/your-org/ttpos-server-go/issues
- **技术文档**：https://docs.ttpos.com

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个星标支持！**

Made with ❤️ by TTPOS Team

</div>
