# 外送服务


## INSTALL
```shell
#安装依赖
go mod tidy

#生成配置，依赖上级目录生成的 .env 文件
make conf

#本地编译运行, **注意改下 manifest/config/config.yaml 里的数据库端口为容器代理出来的端口 **
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


## 外送模块相关的配置

```ini

#外送服务相关配置
#外送服务grpc 接入点, 本地调试用 localhost:14032 ,部署时使用 takeout 服务的地址
TAKEOUT_ENDPOINT=localhost:14032
#外送服务 http 接入点, 本地调试用 localhost:14031 ,部署时使用 takeout 服务的地址
TAKEOUT_SERVICE_URL=http://localhost:14031
#测试环境参考 TAKEOUT_SERVICE_URL=https://ttpos-test1.ttpos.com/api/v1/takeout


#外送服务调用点餐系统回调的加密密钥，目前和 JWT_SECRET 一致
TAKEOUT_CALLBACK_SECRET=$JWT_SECRET  

#SKOOTAR API KEY
SKOOTAR_API_KEY=$SKOOTAR_API_KEY
SKOOTAR_ENDPOINT=https://hw.skootar.com/skootar_api_dev
SKOOTAR_CHANNEL=diylink
SKOOTAR_USERNAME=victor@diylink.net


```