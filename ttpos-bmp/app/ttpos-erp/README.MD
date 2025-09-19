# ERPNEXT 对接模块

## INSTALL
```shell
#首先修改配置
make conf

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


## 部署注意
1. 新增Role Profile: ShopCashier, 权限包括 Accounts Manager，Accounts User，Sales Manager，Sales Master Manager，Sales User，Stock Manager
2. 修改数据库 erp.erp_site 各erpnext site 地址，以及管理员账号的 api_key , api_secret
3. 连锁模式的 site 新增一个总部供应商，名称固定 "Headquarters - Supplier" ,代表公司选择总部公司
4. 执行 `main  migrate --siteCode 1 --dirBase ./manifest/erp-migrate/v2.5` 在erpnext中迁移脚本
5. 执行 `main  migrate --siteCode 1 --dirBase ./manifest/erp-migrate/payment` 在erpnext中迁移支付方式
