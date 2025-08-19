# 点餐门店业务模块

## INSTALL
```shell
#首先修改配置
make conf

#修改 hack/config.yaml, manifest/config/config.yaml 中的数据库名

```

## 开发参考
   1. 设计表结构
   2. 使用gf gen dao 生成对应的 dao/do/model
   3. 编写api层： **定义业务侧数据结构，提供对外接口的输入/输出数据结构**
   4. 编写model层： **定义数据侧数据结构，提供对内的数据处理的输入/输出数据结构**
   5. 编写logic层，自动生成service层代码。（ **通过配置goland File Watcher自动生成** ，也可以通过gf gen service手动执行脚本生成，建议前者）
   6. 在service层代码生成RegisterXX()方法后，在对应的logic模块注册服务(每个模块只需要写一次)
   7. 编写controller层，接收/解析用户输入的参数，调用service层的服务。
   8. 注册路由，对外暴露接口，比如这个项目是编写cmd.go文件。
   9. 在main.go中 加入一行 \_ "project-name/internal/logic" （只需写一次）
   10. 在main.go中加入一行 \_ "github.com/gogf/gf/contrib/drivers/mysql/v2" （如果你使用的是mysql；只需写一次）
