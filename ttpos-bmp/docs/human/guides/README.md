# 开发指南

> 📚 开发规范、最佳实践和工具使用指南

本目录包含 TTPOS 业务中台项目的开发指南，帮助开发者快速上手和遵循最佳实践。

## 📖 指南列表

### [GoFrame 开发指南](./goframe-development-guide.md)
**内容：** 基于 GoFrame v2.x 的完整开发指南
- 项目架构和分层设计
- 开发流程和编码规范
- 数据库操作和配置管理
- HTTP/gRPC 接口开发
- 测试编写和性能优化

### [Protobuf 开发规范](./protobuf-development-guide.md)
**内容：** gRPC 服务接口定义规范
- 文件组织和命名规范
- 消息和服务定义语法
- 版本管理和兼容性
- 代码生成和使用示例

### [数据库设计规范](./database-design-guide.md)
**内容：** 数据库设计和迁移规范
- 表结构设计原则
- 索引和约束使用
- 迁移脚本编写
- 性能优化建议

## 🚀 快速开始

### 新人入职
1. 阅读 [GoFrame开发指南](./goframe-development-guide.md)
2. 查看 [项目主文档](../../../README.MD)
3. 运行 `make help` 了解可用命令

### 功能开发
1. 参考 [开发流程](./goframe-development-guide.md#开发流程)
2. 使用 `gf gen` 命令生成代码
3. 遵循 [编码规范](./goframe-development-guide.md#编码规范)

### 接口开发
1. 阅读 [Protobuf规范](./protobuf-development-guide.md)
2. 定义 `.proto` 文件
3. 使用 `gf gen pb` 生成代码

## 🛠️ 常用命令

```bash
# 项目帮助
make help

# 生成配置文件
make conf

# 生成 DAO 代码
gf gen dao

# 生成 Protobuf 代码
gf gen pb

# 生成控制器代码
gf gen ctrl

# 运行项目
gf run main.go

# 构建项目
gf build
```

## 📋 开发规范

### 代码规范
- 遵循 Go 官方编码规范
- 使用 `gofmt` 格式化代码
- 使用 `go vet` 检查代码
- 提交前运行 `go mod tidy`

### 提交规范
- 使用清晰的提交信息
- 关联相关 issue
- 小步提交，逻辑完整

### 分支管理
- `main`: 主分支，稳定版本
- `develop`: 开发分支
- `feature/*`: 功能分支
- `hotfix/*`: 热修复分支

## 🔧 环境配置

### 开发环境
```bash
# 1. 安装依赖
go mod download

# 2. 配置环境变量
cp .env.example .env

# 3. 生成配置
make conf

# 4. 启动中间件
make mid

# 5. 运行项目
make run.manager
```

### 容器环境
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 🧪 测试

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/logic/...

# 生成测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 集成测试
```bash
# 使用测试数据库
export DB_DATABASE=ttpos_test

# 运行集成测试
go test -tags=integration ./test/...
```

## 📊 性能监控

### 应用监控
- 使用 SkyWalking 进行分布式追踪
- 配置应用指标监控
- 设置健康检查端点

### 数据库监控
- 监控慢查询
- 检查连接池状态
- 定期分析索引使用情况

## 🆘 问题排查

### 常见问题
1. **配置不生效**: 检查 `manifest/config/config.yaml`
2. **数据库连接失败**: 验证 `.env` 中的数据库配置
3. **gRPC 调用失败**: 检查服务注册和网络连接

### 调试技巧
- 使用 `g.Log().Debug()` 输出调试信息
- 查看 SkyWalking 追踪信息
- 检查 Nacos 服务注册状态

## 📚 扩展阅读

- [GoFrame 官方文档](https://goframe.org)
- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 指南](https://developers.google.com/protocol-buffers)
- [Go 语言规范](https://golang.org/ref/spec)

---

**最后更新:** 2025-11-17