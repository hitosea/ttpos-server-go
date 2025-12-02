# hack 目录说明

## 📁 目录用途

`hack` 目录包含 GoFrame CLI 工具的配置文件，用于开发环境中的代码生成和构建工具。

## 📄 文件说明

### config.tpl.yaml

GoFrame CLI 工具的配置模板文件，包含以下配置：

#### 1. 代码生成配置 (gfcli.gen)

##### pbentity - Protobuf 实体生成
```yaml
pbentity:
  link: "mysql:$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT_OPEN)/websocket"
  path: "manifest/protobuf/websocket"
  removePrefix: "ttpos_"
  package: "ttpos-bmp/app/ttpos-websocket/api/websocket"
```
- 从数据库表结构生成 Protobuf 消息定义
- 自动移除表名前缀 `ttpos_`

##### pb - Protobuf 代码生成
```yaml
pb:
  api: "api"
  ctrl: "internal/controller/rpc"
```
- 从 `.proto` 文件生成 Go 代码
- API 代码生成到 `api/` 目录
- 控制器代码生成到 `internal/controller/rpc/` 目录

##### dao - 数据访问层生成
```yaml
dao:
  - link: "mysql:$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT_OPEN)/websocket"
    descriptionTag: true
    removePrefix: "ttpos_"
    tablesEx: "schema_migrations"
```
- 从数据库表结构生成 DAO 层代码
- 排除 `schema_migrations` 表
- 包含字段描述标签

#### 2. 运行配置 (gfcli.run)

```yaml
run:
  path: "./bin"
  watchPaths:
    - api/*.go
    - internal/*.go
```
- 热重载监听文件变化
- 自动重新编译和运行

#### 3. 构建配置 (gfcli.build)

```yaml
build:
  path: "./bin"
  system: linux
  arch: amd64
```
- 构建目标平台：Linux/AMD64
- 输出目录：`./bin`

#### 4. Docker 配置 (gfcli.docker)

```yaml
docker:
  build: "-a amd64 -s linux -p temp -ew"
  tagPrefixes:
    - hub.hitosea.com/ttpos-bmp-websocket
```
- Docker 镜像标签前缀
- 自动构建配置

#### 5. 数据库迁移配置

```yaml
migrate-db-link: mysql://$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/websocket
migrate-db-link-open: mysql://$DB_USERNAME:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT_OPEN)/websocket
```
- 数据库迁移连接字符串
- 支持内网和外网端口

## 🚀 常用命令

### 代码生成

```bash
# 生成 DAO 代码（数据访问层）
gf gen dao

# 生成 Protobuf 代码
gf gen pb

# 生成 PB Entity（从数据库生成 Protobuf）
gf gen pbentity

# 生成所有代码
gf gen dao && gf gen pb
```

### 运行和构建

```bash
# 开发模式运行（支持热重载）
gf run

# 构建可执行文件
gf build

# 构建 Docker 镜像
gf docker
```

### 数据库操作

```bash
# 使用迁移工具（如果配置了）
# 注意：需要额外的数据库迁移工具
```

## 🔧 环境变量

配置文件中使用的环境变量：

| 变量名 | 说明 | 示例值 |
|-------|------|--------|
| `$DB_USERNAME` | 数据库用户名 | `root` |
| `$DB_PASSWORD` | 数据库密码 | `password` |
| `$DB_HOST` | 数据库主机 | `localhost` |
| `$DB_PORT` | 数据库端口（内网） | `3306` |
| `$DB_PORT_OPEN` | 数据库端口（外网） | `3307` |

## 📝 注意事项

### 1. 自动生成的代码不要手动修改

以下目录的代码由工具自动生成，**禁止手动修改**：
- `internal/dao/` - 数据访问层
- `internal/model/do/` - 数据对象
- `internal/model/entity/` - 数据实体
- `internal/service/` - 服务接口

### 2. 修改数据库后重新生成代码

当数据库表结构变化时：
```bash
# 1. 运行数据库迁移
# 2. 重新生成 DAO 代码
gf gen dao
```

### 3. 修改 Protobuf 后重新生成代码

当 `.proto` 文件变化时：
```bash
gf gen pb
```

## 🔗 相关文档

- [GoFrame CLI 工具文档](https://goframe.org/docs/cli)
- [GoFrame 代码生成](https://goframe.org/docs/cli/gen)
- [GoFrame DAO 生成](https://goframe.org/docs/cli/gen-dao)
- [GoFrame Protobuf 生成](https://goframe.org/docs/cli/gen-pb)

## 🛠️ 故障排查

### 问题：找不到 gf 命令

```bash
# 安装 GoFrame CLI 工具
go install github.com/gogf/gf/cmd/gf/v2@latest
```

### 问题：数据库连接失败

1. 检查环境变量是否正确设置
2. 检查数据库服务是否运行
3. 检查网络连接

### 问题：生成的代码有错误

1. 确保数据库表结构正确
2. 确保 `.proto` 文件语法正确
3. 删除旧的生成代码后重新生成

## 📦 项目结构示例

```
ttpos-websocket/
├── hack/
│   ├── config.tpl.yaml       # CLI 工具配置（本文件）
│   └── README.md             # 本说明文档
├── api/                      # API 接口定义（pb 生成）
├── internal/
│   ├── dao/                  # 数据访问层（自动生成）
│   ├── model/
│   │   ├── do/              # 数据对象（自动生成）
│   │   ├── entity/          # 数据实体（自动生成）
│   │   └── dto/             # 数据传输对象（手动）
│   └── service/             # 服务接口（自动生成）
├── manifest/
│   └── protobuf/            # Protobuf 定义文件
└── bin/                     # 构建输出目录
```

---

**文档版本**：1.0  
**最后更新**：2025-11-13  
**维护者**：ttpos-server-go 团队

