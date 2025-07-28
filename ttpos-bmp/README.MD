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
#生成配置，依赖上级目录生成的 .env 文件
make conf
```

## 运行
```shell
# 管理模块
make run.manager
# 门店模块
make run.shop

#外送模块
make run.takeout

#erp模块
make run.erp

```

## 数据库迁移

本项目使用 `golang-migrate` 进行数据库迁移管理，所有迁移脚本都放在 `manifest/sql` 目录中。


### 详细文档
- [数据库迁移规则](./DATABASE_MIGRATION_RULES.md) - 详细的迁移规范和最佳实践
- [快速开始指南](./MIGRATION_QUICK_START.md) - 快速上手指南
- [迁移脚本模板](./migrate_template.sql) - 常用迁移脚本模板

### 迁移目录结构
```
ttpos-bmp/
├── app/
│   ├── ttpos-manager/
│   │   └── manifest/
│   │       └── sql/           # 管理端数据库迁移脚本
│   │   ├── ttpos-erp/
│   │   │   └── manifest/
│   │   │   └── sql/           # ERP系统数据库迁移脚本
│   │   └── ttpos-shop/
│   │       └── manifest/
│   │       └── sql/           # 商城系统数据库迁移脚本
```


## gRpc 服务开发流程
参考工程 [https://github.com/gogf/examples/blob/main/registry/nacos/README.ZH.MD]
1. 定义 proto 文件，在 manifest.proto 中编写 或通过 gf gen pbentity 生成，须注意 manifest/config.yaml 中的gen 配置
2. 通过 gf gen pb 编译 proto 文件，生成对应的 protobuf go 文件以及对应的控制器文件。
3. 新增自定义gRpc 服务，proto文件目录 manifest/protobuf/svc 
4. 执行 gf gen pb 生成 api/ctrl 
5. 调整 生成的服务代码,在 controller 中实现服务接口
6. 在 logic/rpc 中注册控制器