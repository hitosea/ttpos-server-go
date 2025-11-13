# 业务中台模块

## 技术栈

- 基础框架GoFrame [开发手册](https://goframe.org/docs/cli)
- 服务发现，配置管理 Nacos
- 消息中间件 RocketMQ , 支持延迟消息
- 微服务通讯及远程调用 gRPC
- 数据库迁移工具 golang-migrate
- 部分工具参考了框架 HotGo(https://gitee.com/bufanyun/hotgo)

## INSTALL

```shell
go mod tidy
#生成配置，依赖上级目录生成的 .env 文件, 每次.env 变更后执行
make conf
```

## 运行

```shell
# 在容器中启动中间件
make mid

# 在容器中构建并启动全部中台服务
make up

# 本地单独启动各模块
# 管理模块
make run.manager
# 门店模块
make run.shop

#外送模块
make run.takeout

#erp模块
make run.erp

```

### ERP 数据迁移

ERP 历史数据/结构迁移脚本位于 `app/ttpos-erp/manifest/erp-migrate`：

```bash
# 单版本目录迁移（按指定目录执行）
make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5

# 全量版本迁移（遍历根目录下所有版本子目录）
make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate
```

## 数据库迁移

本项目使用 `golang-migrate` 进行数据库迁移管理，所有迁移脚本都放在 `manifest/sql` 目录中。

### 迁移目录结构

```
ttpos-bmp/
├── app/
│   ├── ttpos-manager/           # 管理模块
│   │   └── manifest/
│   │       └── sql/           # 管理端数据库迁移脚本
│   │   ├── ttpos-erp/           # erp模块
│   │   │   └── manifest/
│   │   │   └── sql/           # ERP系统数据库迁移脚本
│   │   └── ttpos-shop/
│   │       └── manifest/
│   │       └── sql/           # 商城系统数据库迁移脚本
│   │   └── ttpos-takout/        #外送模块
│   │       └── manifest/
│   │       └── sql/           # 外送系统数据库迁移脚本
```

## gRpc 服务开发流程

参考工程 [https://github.com/gogf/examples/blob/main/registry/nacos/README.ZH.MD]

1. 定义 proto 文件，在 manifest.proto 中编写 或通过 gf gen pbentity 生成，须注意 manifest/config.yaml 中的gen 配置
2. 通过 gf gen pb 编译 proto 文件，生成对应的 protobuf go 文件以及对应的控制器文件。
3. 新增自定义gRpc 服务，proto文件目录 manifest/protobuf/svc
4. 执行 gf gen pb 生成 api/ctrl
5. 调整 生成的服务代码,在 controller 中实现服务接口
6. 在 logic/rpc 中注册控制器

## 部署步骤

1. 确认.env 包含了外送/中台相关配置 ，参考 .env.example
2. ttpos-bmp 目录下执行 make conf 重新生成各app下的 manifest/config/config.yaml，注意数据库配置，开发环境用的是root，生产测试环境根据情况改一下
3. 服务端口目前使用,需要对外开放， 可根据部署情况调整 
    - 管理 manager模块 14001:http 14002:grpc
    - 门店 shop 模块 14011:http 14012:grpc
    - erp 模块 14021:http 14022:grpc
    - 外送 takeout模块 14031:http 14032:grpc
    - 消息 message 模块 14041:http 14042:grpc

4. 如果有数据库变更,可以通过本各app下的 make db_up.docker 或者主工程的 make migrate 升级迁移数据库变更
5. 构建容器
    - bmp-takeout: 执行 docker compose build bmp-takeout
    - bmp-manager: 执行 docker compose build bmp-manager
    - bmp-shop: 执行 docker compose build bmp-shop
    - bmp-erp: 执行 docker compose build bmp-erp
6. 全量构建  docker compose build
7. 检查容器部署时挂载生成好的 manifest/config/config.yaml 到容器的 的/app/config/config.yaml
8. 启动容器检查日志，可通过 /api/v1/[模块名称]/hello 检查服务是否存活
9. 修改 docker/nginx/conf.d/ttpos-bmp.conf 配置中对应模块的反向代理地址，使用各服务的http 端口，测试环境使用部署主机的服务IP 如192.168.30.24, 生产环境使用slb ip
