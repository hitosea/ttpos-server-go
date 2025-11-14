# Skootar 服务说明文档

## 📋 服务概览

Skootar 服务是 TTPOS 外卖系统中的核心配送服务，负责与第三方配送平台 Skootar 的集成。该服务提供了完整的外卖配送管理功能，包括订单创建、确认、取消、距离预估、司机信息查询等，是外卖业务的核心组件。

## 🎯 主要功能

### 订单管理
- **创建订单**: 向 Skootar 平台创建新的配送订单
- **确认订单**: 商家确认配送订单
- **取消订单**: 取消已创建的配送订单
- **订单详情**: 查询配送订单的详细信息

### 信息查询
- **距离预估**: 预估配送距离和时间
- **司机信息**: 获取配送司机的实时信息
- **订单状态**: 查询订单的当前状态

### 状态管理
- **状态变更**: 处理 Skootar 平台的订单状态回调
- **回调处理**: 验证和处理第三方平台的状态通知

### 配置管理
- **配置获取**: 获取 Skootar API 配置信息
- **URL 构建**: 构建 API 请求地址
- **认证参数**: 生成请求所需的认证参数

## 📁 文件结构

```
internal/service/skootar.go           # 服务接口定义
internal/logic/skootar/               # 服务实现目录
│   ├── skootar.go                   # 核心实现
│   ├── create_order.go               # 创建订单逻辑
│   ├── confirm_order.go              # 确认订单逻辑
│   ├── cancel_order.go               # 取消订单逻辑
│   ├── estimate_distance.go          # 距离预估逻辑
│   ├── get_driver_info.go           # 司机信息查询逻辑
│   └── job_detail.go               # 订单详情查询逻辑
internal/model/dto/skootar/          # Skootar 相关 DTO
│   ├── skootar.go                  # 基础数据结构
│   ├── create_order.go              # 创建订单 DTO
│   ├── confirm_order.go             # 确认订单 DTO
│   ├── cancel_order.go              # 取消订单 DTO
│   ├── estimate_price.go            # 价格预估 DTO
│   ├── get_driver_location.go       # 司机位置 DTO
│   └── job_detail.go               # 订单详情 DTO
internal/model/conf/skootar.go       # Skootar 配置模型
```

## 🔧 接口定义

### ISkootar 接口

```go
type ISkootar interface {
    // 订单管理
    CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error)
    ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error)
    CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error)
    
    // 信息查询
    EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error)
    GetDriverInfo(ctx context.Context, req *dto.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error)
    JobDetail4Food(ctx context.Context, req *skootar.JobDetailInp) (jobDetail *skootar.JobDetail, err error)
    
    // 状态管理
    JobStatusChange(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error)
    
    // 配置管理
    MustConf() *conf.Skootar
    GetUrl(apiPath string) string
    ReqBase() skootar.ReqBase
}
```

#### 方法说明

##### 订单管理方法

###### CancelOrder(ctx, req) (res, err)
- **功能**: 取消 Skootar 配送订单
- **参数**: 取消订单请求参数
- **返回**: 取消订单响应结果

###### ConfirmOrder(ctx, req) (res, err)
- **功能**: 商家确认配送订单
- **参数**: 确认订单请求参数
- **返回**: 确认订单响应结果

###### CreateOrder(ctx, req) (res, err)
- **功能**: 创建新的配送订单
- **参数**: 创建订单请求参数
- **返回**: 创建订单响应结果

##### 信息查询方法

###### EstimateDistance(ctx, req) (res, err)
- **功能**: 预估配送距离和时间
- **参数**: 距离预估请求参数
- **返回**: 距离预估响应结果

###### GetDriverInfo(ctx, req) (res, err)
- **功能**: 获取配送司机信息
- **参数**: 司机信息查询参数
- **返回**: 司机信息响应结果

###### JobDetail4Food(ctx, req) (jobDetail, err)
- **功能**: 查询外卖订单详情
- **参数**: 订单详情查询参数
- **返回**: 订单详情信息

##### 状态管理方法

###### JobStatusChange(ctx, req) (res, err)
- **功能**: 处理订单状态变更回调
- **参数**: 状态变更请求参数
- **返回**: 状态变更响应结果

##### 配置管理方法

###### MustConf() *conf.Skootar
- **功能**: 获取 Skootar 配置信息
- **返回**: Skootar 配置对象

###### GetUrl(apiPath string) string
- **功能**: 构建 API 请求 URL
- **参数**: API 路径
- **返回**: 完整的 API URL

###### ReqBase() skootar.ReqBase
- **功能**: 获取请求基础参数
- **返回**: 请求基础参数对象

## 🏗️ 实现细节

### sSkootar 结构体

```go
type sSkootar struct {}
```

### 核心实现逻辑

1. **服务初始化**
   ```go
   func init() {
       service.RegisterSkootar(Skootar)
   }
   ```

2. **配置管理**
   ```go
   func (s *sSkootar) MustConf() *conf.Skootar {
       if config == nil {
           config = &conf.Skootar{}
           ctx := gctx.New()
           if err := g.Cfg().MustGet(gctx.New(), "app.provider.skootar").Scan(config); err != nil {
               g.Log().Fatal(ctx, err)
               panic(gerror.Newf("获取Skootar配置失败"))
           }
       }
       return config
   }
   ```

3. **URL 构建**
   ```go
   func (s *sSkootar) GetUrl(apiPath string) string {
       return fmt.Sprintf("%v%v", s.MustConf().Endpoint, apiPath)
   }
   ```

4. **请求参数构建**
   ```go
   func (s *sSkootar) ReqBase() skootar.ReqBase {
       return skootar.ReqBase{
           ApiKey:   s.MustConf().ApiKey,
           UserName: s.MustConf().UserName,
           Channel:  s.MustConf().Channel,
       }
   }
   ```

5. **回调认证**
   ```go
   func (*sSkootar) getCallBackAuth(shopRefNo string) string {
       if res, err := gmd5.EncryptString(shopRefNo + g.Cfg().MustGet(gctx.GetInitCtx(), "app.callbackSecret").String()); err == nil {
           return res
       } else {
           g.Log().Error(gctx.GetInitCtx(), "获取回调Auth失败", err)
       }
       return ""
   }
   ```

## 📊 数据模型

### 请求基础参数 (ReqBase)
```go
type ReqBase struct {
    ApiKey        string `json:"apiKey"`                  // API 密钥
    UserName      string `json:"userName"`                // 用户名
    Channel       string `json:"channel"`                 // 渠道标识
    CustomerEmail string `json:"customerEmail,omitempty"`   // 客户邮箱（可选）
    CustomerMoble string `json:"customerMoble,omitempty"`   // 客户手机（可选）
}
```

### 响应基础参数 (RespBase)
```go
type RespBase struct {
    ResponseCode string `json:"responseCode"` // 响应代码
    ResponseDesc string `json:"responseDesc"` // 响应描述
}
```

### 位置信息 (Location)
```go
type Location struct {
    AddressName  string         `json:"addressName,omitempty"`  // 地址名称
    Address      string         `json:"address,omitempty"`       // 详细地址
    Lat          string         `json:"lat,omitempty"`           // 纬度
    Lng          string         `json:"lng,omitempty"`           // 经度
    ContactName  string         `json:"contactName,omitempty"`   // 联系人姓名
    ContactPhone string         `json:"contactPhone,omitempty"`  // 联系电话
    CashFee      consts.CashFee `json:"cashFee,omitempty"`      // 现金费用标识
    Seq          int            `json:"seq"`                    // 序列号
    Remark       string         `json:"remark,omitempty"`        // 备注
}
```

## 🔄 使用流程

### 1. 创建配送订单
```go
// 获取服务实例
skootarService := service.Skootar()

// 构建创建订单请求
req := &api.CreateOrderReq{
    // 填充订单信息...
}

// 调用创建订单接口
resp, err := skootarService.CreateOrder(ctx, req)
if err != nil {
    // 处理错误
    return err
}

// 处理响应
fmt.Printf("订单创建成功，订单号: %s\n", resp.TakeoutRefNo)
```

### 2. 查询司机信息
```go
// 构建查询请求
req := &dto.GetDriverInfoInp{
    TakeoutRefNo: "SK123456",
}

// 调用查询接口
resp, err := skootarService.GetDriverInfo(ctx, req)
if err != nil {
    return err
}

// 处理司机信息
fmt.Printf("司机姓名: %s, 电话: %s\n", resp.SkootarName, resp.SkootarPhone)
```

### 3. 处理状态回调
```go
// 处理 Skootar 平台的状态回调
req := &v1.SkootarStatusReq{
    JobId:     "123456",
    JobStatus:  "completed",
    // 其他状态信息...
}

res, err := skootarService.JobStatusChange(ctx, req)
if err != nil {
    return err
}

// 处理回调响应
fmt.Printf("状态处理完成: %s\n", res.Message)
```

## ⚠️ 注意事项

1. **配置依赖**: 服务依赖正确的 Skootar API 配置，包括 Endpoint、ApiKey、UserName 等
2. **网络连接**: 需要稳定的网络连接访问 Skootar API
3. **错误处理**: 需要妥善处理 API 调用失败的情况
4. **数据一致性**: 确保本地订单状态与 Skootar 平台状态同步
5. **安全性**: 妥善保管 API 密钥和回调密钥
6. **限流处理**: 注意 API 调用频率限制
7. **回调验证**: 验证回调请求的合法性

## 🔮 扩展性

### 功能扩展
1. **多供应商支持**: 扩展支持其他配送平台（如 Grab）
2. **批量操作**: 支持批量创建和查询订单
3. **智能调度**: 基于历史数据的智能配送建议
4. **实时追踪**: WebSocket 实时订单状态推送
5. **费用优化**: 自动选择最优配送方案

### 监控扩展
1. **性能监控**: API 调用时间和成功率监控
2. **异常告警**: 配送失败和异常情况告警
3. **数据分析**: 配送数据统计和分析
4. **SLA 监控**: 配送时效和服务质量监控

## 🎯 使用场景

### 订单创建场景
1. **即时配送**: 用户下单后立即创建配送订单
2. **预约配送**: 用户预约特定时间的配送服务
3. **批量配送**: 商家批量处理多个配送订单

### 状态管理场景
1. **实时追踪**: 用户实时查看配送状态
2. **异常处理**: 配送异常时的状态更新和处理
3. **完成确认**: 配送完成后的状态确认

### 运营分析场景
1. **效率分析**: 分析配送效率和时效
2. **成本控制**: 监控和控制配送成本
3. **服务质量**: 评估配送服务质量

## 🔧 配置管理

### 配置结构
```go
type Skootar struct {
    Endpoint string `json:"endpoint"` // API 端点
    ApiKey   string `json:"apiKey"`   // API 密钥
    UserName string `json:"userName"` // 用户名
    Channel  string `json:"channel"`  // 渠道标识
}
```

### 配置获取
服务提供了懒加载的配置获取机制：
- 首次调用时从配置文件加载
- 后续调用直接返回缓存配置
- 配置加载失败时会记录日志并抛出异常

## 📝 总结

Skootar 服务是外卖系统的核心业务组件，提供了完整的第三方配送平台集成能力。

### 技术特点
- **完整集成**: 与 Skootar 平台的完整 API 集成
- **配置管理**: 灵活的配置管理和加载机制
- **错误处理**: 完善的错误处理和日志记录
- **安全认证**: 支持回调请求的安全验证

### 设计优势
- **模块化**: 清晰的模块划分和接口设计
- **可扩展**: 预留了多供应商扩展能力
- **可维护**: 标准化的代码结构和完整的中文注释
- **高性能**: 优化的网络调用和数据处理

### 业务价值
- **配送能力**: 提供专业的配送服务能力
- **用户体验**: 实时的配送状态和司机信息
- **运营效率**: 自动化的配送流程管理
- **成本控制**: 灵活的配送方案选择

该服务是外卖系统实现配送功能的关键组件，为用户提供了可靠的配送服务体验。
