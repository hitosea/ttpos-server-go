# Go BMP 模块开发指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: Go BMP 模块（GoFrame）的详细开发规范和最佳实践

---

## 📋 重要说明

**ttpos-bmp 项目有自己的专用开发规范**，详见：
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go代码开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf开发规范

本文档是对这些规范的补充和详细说明。

---

## 框架说明

本项目使用 [GoFrame](https://github.com/gogf/gf) v2.x 框架开发。

**参考资料**:
- GoFrame 官方文档：https://goframe.org.cn
- GoFrame API 文档：https://pkg.go.dev/github.com/gogf/gf/v2

---

## 开发规范

### 1. 服务实现模式

```go
// 服务实现（在 logic/ 目录）
type sUser struct{}

// 单例模式
var (
    insUser = sUser{}
)

func User() *sUser {
    return &insUser
}

func init() {
    service.RegisterUser(User())  // 注册到服务容器
}

// 实现业务逻辑
func (s *sUser) GetUser(ctx context.Context, req *user.GetUserReq) (*user.GetUserResp, error) {
    // 业务逻辑实现
    var user *entity.User
    err := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, req.UserId).Scan(&user)
    if err != nil {
        return nil, err
    }
    
    return &user.GetUserResp{
        UserId:   user.Id,
        Username: user.Username,
    }, nil
}
```

### 2. 依赖注入

```go
type sOrder struct {
    userLogic *sUser
}

func Order() *sOrder {
    return &sOrder{
        userLogic: User(),
    }
}
```

---

## 数据库操作

### 1. 使用 DAO 层

```go
// 查询操作
var user *entity.User
err := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, userId).Scan(&user)

// 插入操作
userId, err := dao.User.Ctx(ctx).Data(do.User{
    Username: "test",
    Email:    "test@example.com",
}).InsertAndGetId()

// 更新操作
_, err := dao.User.Ctx(ctx).
    Where(dao.User.Columns().Id, userId).
    Update(do.User{
        Username: "updated",
    })

// 删除操作（软删除）
_, err := dao.User.Ctx(ctx).
    Where(dao.User.Columns().Id, userId).
    Delete()
```

### 2. 事务处理

```go
err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
    // 在事务中执行多个操作
    _, err := dao.User.Ctx(ctx).TX(tx).Data(userData).Insert()
    if err != nil {
        return err
    }
    
    _, err = dao.UserProfile.Ctx(ctx).TX(tx).Data(profileData).Insert()
    if err != nil {
        return err
    }
    return nil
})
```

---

## 错误处理

### 使用 gerror 包

```go
import "github.com/gogf/gf/v2/errors/gerror"

// 创建新错误
err := gerror.New("用户不存在")

// 包装已有错误
err = gerror.Wrap(originalErr, "获取用户失败")

// 格式化错误信息
err = gerror.Newf("用户ID %d 不存在", userID)

// 包装并格式化
err = gerror.Wrapf(originalErr, "更新用户 %d 失败", userID)
```

**最佳实践**:
- ✅ 使用 `gerror` 包，不要使用标准库的 `errors`
- ✅ 错误信息使用中文
- ✅ 在业务逻辑层包装错误，添加业务含义
- ❌ 避免使用 `panic`

---

## 配置管理

### 1. 配置文件结构

```yaml
# manifest/config/config.tpl.yaml
server:
  address: ":8080"
  logPath: "./log"
  
database:
  default:
    link: "mysql:$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME"
    
logger:
  level: "all"
  stdout: true
```

### 2. 配置读取

```go
import "github.com/gogf/gf/v2/frame/g"

// 读取配置
serverAddr := g.Cfg().MustGet(ctx, "server.address").String()
dbConfig := g.Cfg().MustGet(ctx, "database.default").Map()
```

---

## 日志规范

### 日志使用

```go
import "github.com/gogf/gf/v2/frame/g"

// 不同级别的日志
g.Log().Debug(ctx, "调试信息")
g.Log().Info(ctx, "普通信息")
g.Log().Warning(ctx, "警告信息")
g.Log().Error(ctx, "错误信息", err)
g.Log().Fatal(ctx, "致命错误", err)

// 格式化日志
g.Log().Infof(ctx, "用户 %d 执行了 %s 操作", userId, action)
```

**最佳实践**:
- ✅ 必须传递 `context` 参数，便于链路追踪
- ✅ 错误日志包含完整的错误信息
- ✅ 使用结构化日志记录关键业务操作
- ❌ 敏感信息不要记录到日志中

---

## 数据库迁移

### 迁移脚本示例

```sql
-- xxx_up.sql
-- 创建用户表
CREATE TABLE IF NOT EXISTS `user` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一标识',
    `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名',
    `email` varchar(100) NOT NULL DEFAULT '' COMMENT '邮箱',
    `create_time` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `update_time` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    `delete_time` int NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

---

## gRPC 服务开发

### gRPC 控制器实现

```go
// internal/controller/rpc/user.go
package rpc

import (
    "context"
    "ttpos-bmp/app/ttpos-erp/api/user"
    "ttpos-bmp/app/ttpos-erp/internal/service"
)

type cUser struct {
    user.UnimplementedUserServiceServer
}

var User = new(cUser)

func (c *cUser) GetUserInfo(ctx context.Context, req *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
    return service.User().GetUser(ctx, req)
}
```

---

## 性能优化

### 1. 数据库查询优化

```go
// ✅ 使用索引字段查询
dao.User.Ctx(ctx).Where(dao.User.Columns().Username, username).One()

// ❌ 避免全表扫描
dao.User.Ctx(ctx).Where("email LIKE ?", "%@example.com%").All()  // 慢
dao.User.Ctx(ctx).Where(dao.User.Columns().Email, email).All()   // 快
```

### 2. 使用缓存

```go
import "github.com/gogf/gf/v2/os/gcache"

// 设置缓存
gcache.Set(ctx, "user:"+userId, userData, time.Hour)

// 获取缓存
value, err := gcache.Get(ctx, "user:"+userId)
```

---

## 安全规范

### 1. 输入验证

```go
type CreateUserReq struct {
    Username string `v:"required|length:2,20#用户名不能为空|用户名长度为2-20个字符"`
    Email    string `v:"required|email#邮箱不能为空|邮箱格式不正确"`
    Age      int    `v:"required|between:1,150#年龄不能为空|年龄必须在1-150之间"`
}
```

### 2. SQL注入防护

```go
// ✅ 正确：使用参数化查询
dao.User.Ctx(ctx).Where("username = ?", username).One()

// ❌ 错误：字符串拼接
dao.User.Ctx(ctx).Where("username = '" + username + "'").One()
```

---

## 注意事项

### 开发参考文档优先级

1. `ttpos-bmp/README.MD` - 项目说明
2. `ttpos-bmp/MIGRATION_QUICK_START.md` - 迁移快速开始
3. `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go开发规范
4. `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf规范
5. GoFrame 官方文档

### 重要提示

- ❌ 分析处理 ttpos-bmp 代码时，不要参考 `main/` 目录下的代码
- ✅ 遵循 GoFrame 框架的标准结构和最佳实践
- ✅ 使用框架提供的工具和命令

---

## 相关文档

- [Go BMP 架构设计](../architecture/go-bmp-architecture.md) - 深入理解架构
- [Protobuf 规范](../../ttpos-bmp/.cursor/rules/proto-rules.mdc) - Protobuf 开发规范
- [GoFrame 官方文档](https://goframe.org.cn) - 框架详细文档
- [微服务集成工作流](../../agent/workflows/microservice-integration.md) - 微服务开发流程

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

