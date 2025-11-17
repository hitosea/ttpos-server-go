# GoFrame 开发指南

> 基于 GoFrame v2.x 的开发规范和最佳实践

## 📋 概述

本指南介绍在 TTPOS 业务中台项目中使用 GoFrame v2.x 框架的开发规范、架构模式和最佳实践。

## 🏗️ 项目架构

### 目录结构
```
app/ttpos-[module]/
├── api/                    # API 层（自动生成）
├── internal/               # 内部代码
│   ├── cmd/               # 命令行工具
│   ├── consts/            # 常量定义
│   ├── controller/        # 控制器层
│   ├── logic/             # 业务逻辑层
│   ├── model/             # 数据模型
│   │   ├── dao/          # 数据访问对象
│   │   ├── do/           # 数据对象
│   │   └── entity/       # 实体对象
│   └── service/           # 服务层
├── manifest/               # 配置文件和资源
│   ├── config/            # 配置文件
│   ├── protobuf/          # Protobuf 定义
│   └── sql/               # 数据库迁移脚本
└── test/                  # 测试文件
```

### 分层架构
1. **API 层**: HTTP/gRPC 接口定义（自动生成）
2. **Controller 层**: 请求处理和响应
3. **Logic 层**: 业务逻辑处理
4. **Service 层**: 通用服务组件
5. **DAO 层**: 数据访问对象（自动生成）

## 🚀 开发流程

### 1. 定义数据模型
```go
// 使用 gf gen dao 生成基础 DAO 代码
// 表结构定义在 manifest/sql/ 中
```

### 2. 定义 API 接口
```protobuf
// manifest/protobuf/svc/[service].proto
service UserService {
  rpc GetUser(GetUserReq) returns (GetUserRes);
}
```

### 3. 生成代码
```bash
# 生成 DAO
gf gen dao

# 生成 Protobuf 代码
gf gen pb

# 生成控制器
gf gen ctrl
```

### 4. 实现业务逻辑
```go
// internal/logic/[module]/[service].go
type sUser struct{}

func (s *sUser) GetUser(ctx context.Context, req *v1.GetUserReq) (*v1.GetUserRes, error) {
    // 业务逻辑实现
    return &v1.GetUserRes{User: user}, nil
}
```

## 📝 编码规范

### 命名规范
- **包名**: 全小写，使用下划线分隔，如 `user_service`
- **结构体**: PascalCase，如 `UserInfo`
- **方法**: PascalCase，如 `GetUserInfo`
- **变量**: camelCase，如 `userName`

### 错误处理
```go
// 使用 gerror 进行错误包装
return nil, gerror.New("用户不存在")

// 使用自定义错误码
return nil, gerror.NewCode(gcode.CodeNotFound, "用户不存在")
```

### 日志记录
```go
// 使用内置日志
g.Log().Info(ctx, "用户登录", g.Map{
    "userId": userId,
    "ip": clientIP,
})

// 不同日志级别
g.Log().Debug(ctx, "调试信息")
g.Log().Info(ctx, "普通信息")
g.Log().Warning(ctx, "警告信息")
g.Log().Error(ctx, "错误信息")
```

## 🗄️ 数据库操作

### DAO 使用
```go
// 查询单个用户
user, err := dao.User.Ctx(ctx).Where("id", userId).One()

// 分页查询
users, err := dao.User.Ctx(ctx).Page(page, size).All()

// 条件查询
users, err := dao.User.Ctx(ctx).Where(g.Map{
    "status": 1,
    "create_time >": startTime,
}).All()
```

### 事务处理
```go
err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
    // 在事务中执行数据库操作
    _, err := dao.User.Ctx(ctx).TX(tx).Insert(user)
    if err != nil {
        return err
    }

    _, err = dao.UserRole.Ctx(ctx).TX(tx).Insert(userRole)
    return err
})
```

## 🔧 配置管理

### 配置文件结构
```yaml
# manifest/config/config.yaml
server:
  address: ":14001"
  serverRoot: "resource/public"

database:
  default:
    link: "mysql:root:password@tcp(127.0.0.1:3306)/ttpos"
    debug: true

redis:
  default:
    address: "127.0.0.1:6379"
    db: 0
```

### 配置使用
```go
// 获取配置
port := g.Cfg().MustGet(ctx, "server.address").String()

// 获取数据库配置
dbConfig := g.Cfg().MustGet(ctx, "database.default").Map()
```

## 🌐 HTTP 接口开发

### 控制器实现
```go
type cUser struct {
    g.Meta `path:"/user" tags:"用户管理" method:"get,post" summary:"用户接口"`
}

// 获取用户列表
func (c *cUser) GetList(ctx context.Context, req *v1.GetUserListReq) (res *v1.GetUserListRes, err error) {
    res = &v1.GetUserListRes{}

    // 业务逻辑调用
    res.Users, err = logic.User().GetList(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}
```

### 中间件使用
```go
// 注册中间件
s := g.Server()
s.Use(ghttp.MiddlewareHandlerResponse)
s.Use(middleware.Auth)
s.Use(middleware.CORS)
```

## 📡 gRPC 服务开发

### 服务定义
```protobuf
syntax = "proto3";

service UserService {
  rpc GetUser(GetUserReq) returns (GetUserRes);
}

message GetUserReq {
  int64 user_id = 1;
}

message GetUserRes {
  User user = 1;
}

message User {
  int64 id = 1;
  string username = 2;
  string email = 3;
}
```

### 服务实现
```go
type sUserService struct{}

func (s *sUserService) GetUser(ctx context.Context, req *v1.GetUserReq) (*v1.GetUserRes, error) {
    user, err := logic.User().GetById(ctx, req.UserId)
    if err != nil {
        return nil, err
    }

    return &v1.GetUserRes{User: user}, nil
}
```

## 🧪 测试编写

### 单元测试
```go
func TestUser_GetById(t *testing.T) {
    // 创建测试服务
    s := &sUser{}

    // 准备测试数据
    ctx := context.Background()
    userId := int64(1)

    // 执行测试
    user, err := s.GetById(ctx, userId)

    // 断言结果
    assert.Nil(t, err)
    assert.Equal(t, userId, user.Id)
}
```

### 集成测试
```go
func TestUserAPI_GetList(t *testing.T) {
    // 启动测试服务器
    // 发送 HTTP 请求
    // 验证响应
}
```

## 🔒 安全考虑

### 输入验证
```go
// 使用 gvalid 进行参数验证
type GetUserReq struct {
    g.Meta `valid:"GetUserReq"`
    UserId int64 `v:"required|min:1" dc:"用户ID"`
}

func (c *cUser) Get(ctx context.Context, req *v1.GetUserReq) (res *v1.GetUserRes, err error) {
    // 自动进行参数验证
    // ...
}
```

### 权限控制
```go
// 实现权限中间件
func Auth(r *ghttp.Request) {
    token := r.GetHeader("Authorization")
    if token == "" {
        r.Response.WriteJson(g.Map{"code": 401, "message": "未授权"})
        r.Exit()
    }

    // 验证 token 并设置用户信息到上下文
    // ...
}
```

## 📊 性能优化

### 缓存使用
```go
// Redis 缓存
cache := g.Redis()
err := cache.Set(ctx, "user:"+userId, user, time.Hour)

// 缓存查询
user, err := cache.Get(ctx, "user:"+userId)
```

### 数据库优化
- 使用索引
- 避免 N+1 查询
- 使用分页查询
- 合理使用连接池

## 🔧 常用命令

```bash
# 生成 DAO
gf gen dao

# 生成 Protobuf
gf gen pb

# 生成控制器
gf gen ctrl

# 运行项目
gf run main.go

# 构建项目
gf build
```

## 📚 参考资料

- [GoFrame 官方文档](https://goframe.org)
- [Protobuf 指南](https://developers.google.com/protocol-buffers)
- [gRPC 文档](https://grpc.io/docs/)

---

**最后更新:** 2025-11-17