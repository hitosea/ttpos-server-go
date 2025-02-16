# ttpos-server-go

1. 结构体，除了ID是大写，其他都是大小写，比如StaffId；传参时，id用小写，其他用大小写，比如staffId
2. url 使用snake(蛇形)写法，不使用kebab(烤串)写法，举例如下：
```html
http://your.domain.com/api/v1/passport/server_public_key
```
3. 接口响应返回统一格式，data只能返回对象，不能返回null或数组；如果有分页，则分页信息使用meta返回，如果有其他属性，则属性也不能返回数组，举例如下：
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "options": {
      "list": [{"key": "k1", "value": "v1"}]
    },
    "list": [{"foo": "bar"}],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```
4. 业务逻辑中，不要使用panic，使用返回错误的方式，错误可以预先定义，也可以使用app error
5. 包名和文件名，使用snake写法
6. 如果需要将gin.Context作为参数传递，可以使用copy方法获取副本，在协程中使用，注意不能对该对象进行写操作
7. 导入的包按照自带包-第三方包-项目包的顺序书写，用空行分隔
8. 自动验证的错误消息，使用对应结构+Message的map，传递参数，方便国际化，参考收银端登录接口
9. api文档使用https://github.com/swaggo/swag 
10. API接口响应时间要求本地响应200ms以内
11. repository层的接口实现不能传入dbm实例，只能传入db实例
12. 系统事件总线。使用前先`定义事件`，才可以`发布事件`和`订阅事件`

```
# 定义事件，参考：
main/pkg/eventbus/event/sample_event.go

定义事件统一在/pkg/eventbus/event目录下，一个文件对应一个事件的定义，文件名为事件名称，格式为xxx_event.go

例如：sample_event.go表示Sample事件的定义，定义在/pkg/eventbus/event/sample_event.go
```
```
# 发布事件
引入这个包	     "ttpos-server-go/pkg/eventbus/event"
发布一个Sample事件 event.NewSystemBus().PublishSampleEvent()
```
```
# 订阅事件
引入这个包	     "ttpos-server-go/pkg/eventbus/event"
订阅Sample事件 event.NewSystemBus().SubscribeSampleEvent(func(msg event.SamplePayload) {})

订阅事件的事件处理器统一下在/app/event目录下，一个文件对应一个事件的处理。文件名为事件名称，格式为xxx_event_handler.go，

举例如下cancel_order_event_handler.go表示取消订单事件的处理器，定义在/app/event/cancel_order_event_handler.go
```
13. 并发控制uuid锁。在并发场景下，当操作同一个uuid资源时需要先获取uuid锁，先得到锁的协程先执行而其他协程等待锁。
```
使用步骤：
1. 导入包 ttpos-server-go/pkg/lock
2. 获取单例系统锁。将系统锁作为Service的属性之一
type Service struct {
	systemLock Lock
}
func NewService(systemLock Lock) *Service {
	return &Service{systemLock: lock.NewSystemLock()}
}
3. 使用锁。在需要并发控制的方法中添加如下两行代码
func (s *Service) OpenDesk(deskUuid uint64) error {
	s.systemLock.LockUuid(deskUuid)
	defer s.systemLock.UnlockUuid(deskUuid)
	return nil
}
4. 注意uuid资源的状态终止时，手动删除uuid锁资源
比如，在桌台被删除时，执行ClearUuidLock方法删除锁资源以清空内存占用
s.systemLock.ClearUuidLock(deskUuid)
```
14. 服务A依赖服务B，通过传参的形式；服务A不能直接依赖服务B所依赖的repo，方便后期按照服务划分模块
15. 接口声明使用I开头，实现使用Impl结尾；Repository简写成Repo，Service简写成Srv
```code
// 声明repository接口 IProductCategoryRepo
type IProductCategoryRepo interface {}

// ProductCategoryRepoImpl 实现 IProductCategoryRepo
type ProductCategoryRepoImpl struct {}

// 声明service接口 ICashierSrv
type ICashierSrv interface {}

// CashierSrvImpl 实现 ICashierSrv
type CashierSrvImpl struct {}
```
16. 所有model结构体，使用gorm的column选项，指定数据库表字段名，举例如下：
```code
type ProductCategory struct {
	ID uint `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
}
```
17. 所有model结构体，使用注释，指定数据库表名，举例如下：
```code
// 商品分类表 ttpos_product_category
type ProductCategory struct {
	ID uint `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
}
```
18. func方法的参数如果超过3个，要定义参数结构体，3个以下包括3个可以直接写参数名