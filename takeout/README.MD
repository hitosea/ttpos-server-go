# 外送服务


## INSTALL
```shell
#安装依赖
go mod tidy

#生成配置，依赖上级目录生成的 .env 文件
make conf

#编译运行
make run

#升级 db
make db_up.docker

#新增db变更
make db_add NAME='comment'

```

## gRpc 服务开发流程

参考工程 [https://github.com/gogf/examples/blob/main/registry/nacos/README.ZH.MD]

1. 定义 proto 文件，在 manifest.proto 中编写 或通过 gf gen pbentity 生成，须注意 manifest/config.yaml 中的gen 配置
2. 新增自定义gRpc 服务，proto文件目录 manifest/protobuf/svc
3. 通过 gf gen pb 编译 proto 文件，生成对应的 protobuf go 文件以及对应的控制器文件。
4. 执行 gf gen pb 生成 api/ctrl
5. 调整 生成的服务代码,在 controller 中实现服务接口, 目前会打包为 api和放在 controller/api 中需要改下名
6. 在 logic/rpc 中注册控制器


## 部署

部署需要注意几个点
1. .env.example 新增了配置， .env 需要同步增加
2. 新增环境相关的 manifest/config/config.yaml ，模板manifest/config/config.tpl.yaml，修改对应的账号密码
3. 新增的config.yaml 部署时 挂载到容器 的/app/config/config.yaml 下
4. 数据库初始化需要新增 一个takeout 库， 后续更新可以通过本工程 make db_up.docker 或者主工程的 make migrate 升级迁移数据库变更