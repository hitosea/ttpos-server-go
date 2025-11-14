# 会员API模块文档

## 概述

会员API模块是TTPOS系统中的核心模块之一，负责处理所有面向会员端用户的API请求。该模块基于Gin框架开发，采用RESTful API设计规范，提供完整的会员生命周期管理功能。

## 模块结构

会员API模块位于 `main/app/api/v1/member/` 目录下，包含8个核心文件：

```
member/
├── member_address.go      # 会员地址管理
├── member_auth.go         # 会员认证相关
├── member_base.go         # 会员基础信息管理
├── member_benefit.go      # 会员权益管理
├── member_callback.go     # 回调处理
├── member_marketing.go    # 营销活动管理
├── member_order.go        # 会员端订单管理
└── member_product.go      # 会员端产品管理
```

## 功能分类

### 1. 会员认证 (member_auth.go)
负责会员的身份验证、注册、登录等基础认证功能。

#### 主要API接口：
- **POST /member/login_info**: 获取登录前信息
- **POST /member/send_code**: 发送登录验证码
- **POST /member/send_register_code**: 发送注册验证码
- **POST /member/login**: 登录
- **POST /member/register**: 注册
- **POST /member/visitor/login**: 游客登录

#### 认证流程：
1. 游客登录：无需注册，直接生成token
2. 手机号登录：发送验证码 → 验证登录
3. 注册流程：发送验证码 → 验证注册 → 获取token

### 2. 会员基础信息 (member_base.go)
管理会员的基础信息维护。

#### 主要API接口：
- **GET /member/base**: 获取基础信息
- **POST /member/base/nickname**: 修改会员昵称

### 3. 会员地址管理 (member_address.go)
提供完整的地址管理功能，支持谷歌地图地址搜索。

#### 主要API接口：
- **GET /member/address/search**: 搜索谷歌地图地址
- **GET /member/address**: 获取地址列表
- **GET /member/address/detail**: 获取地址详情
- **POST /member/address/add**: 添加地址
- **POST /member/address/update**: 更新地址
- **DELETE /member/address/delete**: 删除地址
- **POST /member/address/auth**: 认证地址手机号

#### 地址认证流程：
1. 搜索并选择地址
2. 提交手机号认证
3. 系统自动注册会员并返回token

### 4. 会员权益管理 (member_benefit.go)
管理会员的优惠券和积分权益。

#### 主要API接口：
- **GET /member/coupon/list**: 获取我的优惠券
- **GET /member/coupon/history/list**: 获取我的历史优惠券
- **GET /member/points/record/list**: 获取我的积分记录

### 5. 营销活动管理 (member_marketing.go)
提供营销活动相关的功能。

#### 主要API接口：
- **GET /member/marketing_activity**: 获取营销活动（已废弃）
- **GET /member/marketing_activity_list**: 获取营销活动列表
- **GET /member/marketing_activity_detail**: 获取营销活动详情

### 6. 会员端订单管理 (member_order.go)
完整的订单生命周期管理，从创建到完成。

#### 主要API接口：
- **POST /member/order/create**: 创建会员端订单
- **GET /member/order/form/info**: 获取订单提交表单信息
- **POST /member/order/address**: 设置订单地址
- **POST /member/order/pay**: 提交支付
- **GET /member/order/pay/info**: 获取支付信息
- **GET /member/order/pay/status**: 获取支付状态
- **GET /member/order/list**: 获取会员端订单列表
- **GET /member/order/detail**: 获取会员端订单详情
- **GET /member/order/payment/method/list**: 获取支付方式列表
- **POST /member/order/cancel**: 取消订单
- **GET /member/order/rider**: 获取骑手信息
- **POST /member/order/send_auth_code**: 发送认证验证码

#### 订单流程：
1. 获取表单信息
2. 创建订单
3. 设置配送地址
4. 选择支付方式并支付
5. 跟踪支付状态和骑手位置
6. 查看订单详情

### 7. 会员端产品管理 (member_product.go)
提供产品浏览和搜索功能。

#### 主要API接口：
- **GET /member/product/category/list**: 获取产品类别列表
- **GET /member/product/list**: 获取产品列表
- **GET /member/product/recommend/list**: 获取推荐商品列表
- **GET /member/product/search**: 搜索商品

### 8. 回调处理 (member_callback.go)
处理第三方支付回调。

#### 主要API接口：
- **POST /member/order/callback**: 处理支付回调

## 认证机制

### JWT Token认证
- 大部分API接口需要JWT Token认证
- 通过 `middleware.MemberAuth` 中间件验证
- Token包含会员UUID、公司UUID等信息

### 特殊认证
- 回调接口使用自定义Token验证：`MD5(订单号 + secret)`
- 游客登录无需预先认证

## 数据流设计

### 请求处理流程
1. **路由匹配**: 根据URL路径匹配到对应处理器
2. **参数绑定**: 使用Gin的ShouldBind自动绑定请求参数
3. **参数验证**: 自定义验证器或binding标签验证
4. **上下文获取**: 通过helper.GetContext获取业务上下文
5. **业务处理**: 调用相应的服务层方法
6. **错误处理**: 统一的错误处理和响应格式
7. **响应返回**: 统一的JSON响应格式

### 响应格式规范
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [...],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

### 错误响应格式
```json
{
  "code": 0,
  "message": "错误信息",
  "data": null
}
```

## 服务依赖

### 核心服务依赖
- **AuthSrv**: 认证服务
- **MemberSrv**: 会员服务
- **OrderSrv**: 订单服务
- **ProductSrv**: 产品服务
- **SettingSrv**: 设置服务

### 基础设施依赖
- **DBManager**: 数据库管理器
- **Cache**: 缓存服务
- **Logger**: 日志服务

## API文档规范

### Swagger注解
所有API接口都使用Swaggo注解生成API文档：
- `@Summary`: 接口简述
- `@Description`: 接口详细描述
- `@Tags`: 接口分组标签
- `@Security`: 安全认证方式
- `@Param`: 参数说明
- `@Success/@Failure`: 响应格式说明

### 接口分组标签
- `会员端.认证`: 认证相关接口
- `会员端.地址`: 地址管理接口
- `会员端.基础信息`: 基础信息管理接口
- `会员端.优惠券`: 优惠券管理接口
- `会员端.积分`: 积分管理接口
- `会员端.营销活动`: 营销活动接口
- `会员端-订单`: 订单管理接口
- `会员端-订单列表`: 订单列表接口
- `会员端.产品列表`: 产品管理接口
- `会员端.回调`: 回调处理接口

## 性能优化

### 缓存策略
- 热点数据缓存（如商品信息、用户信息）
- 验证码缓存（带过期时间）
- 会话信息缓存

### 数据库优化
- 分页查询优化
- 索引优化
- 预加载关联数据

### 并发控制
- UUID锁机制防止重复操作
- 分布式锁保证数据一致性

## 安全考虑

### 输入验证
- 参数绑定验证
- 自定义验证器
- XSS防护

### 权限控制
- JWT Token验证
- 公司级别数据隔离
- 操作权限检查

### 敏感信息保护
- 密码加密存储
- API密钥保护
- 日志脱敏

## 扩展性设计

### 模块化架构
- 每个功能模块独立文件
- 服务层接口化设计
- 依赖注入模式

### 配置化支持
- 动态配置支持
- 多环境配置
- 功能开关控制

### 国际化支持
- 多语言响应
- 错误信息国际化
- 业务文案国际化

## 监控和日志

### 日志记录
- 请求响应日志
- 错误异常日志
- 业务操作日志

### 性能监控
- API响应时间监控
- 错误率统计
- 业务指标监控

## 维护建议

### 代码规范
- 中文注释
- 统一错误处理
- 代码格式化

### 测试策略
- 单元测试覆盖核心逻辑
- 集成测试验证API接口
- 压力测试验证性能表现

### 部署考虑
- 水平扩展支持
- 灰度发布能力
- 回滚方案准备
