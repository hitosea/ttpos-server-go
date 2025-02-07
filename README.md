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
      "total": 100 // 总记录数
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
11. 服务A依赖服务B，通过传参的形式；服务A不能直接依赖服务B所依赖的repo，方便后期按照服务划分模块