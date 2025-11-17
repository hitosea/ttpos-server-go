# Takeout 服务说明文档

## 📋 服务概览

Takeout 服务是 TTPOS 外卖系统中的数据访问服务，负责外卖订单数据的查询和管理。该服务作为数据层组件，为其他业务服务提供统一的外卖订单数据访问接口，支持多供应商的订单数据管理。

## 🎯 主要功能

### 订单数据查询
- **订单查询**: 根据商店订单 UUID 查询外卖订单信息
- **数据访问**: 提供统一的订单数据访问接口
- **供应商支持**: 支持多个外卖供应商的订单数据

### 服务路由
- **供应商路由**: 根据供应商名称路由到对应的服务实现
- **接口统一**: 为不同供应商提供统一的服务接口
- **扩展支持**: 预留新供应商的扩展能力

## 📁 文件结构

```
internal/service/takeout.go           # 服务接口定义
internal/logic/takeout/takeout.go   # 服务实现
internal/model/entity/job.go          # 订单实体定义
internal/consts/consts.go           # 供应商常量定义
```

## 🔧 接口定义

### ITakeout 接口

```go
type ITakeout interface {
    Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error)
}
```

#### 方法说明

##### Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error)
- **功能**: 根据商店订单 UUID 查询外卖订单信息
- **参数**: 
  - `ctx context.Context`: Go 上下文对象
  - `shopOrderUuid string`: 商店订单 UUID
- **返回值**: 
  - `*entity.Job`: 外卖订单实体信息
  - `error`: 错误信息
- **说明**: 
  - 从数据库查询指定 UUID 的外卖订单
  - 返回完整的订单信息，包括配送状态、司机信息等
  - 如果订单不存在或查询失败，返回相应的错误信息

### 扩展接口定义

在实现中还定义了扩展的业务接口：

```go
type ITakeout interface {
    // 核心数据查询
    Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error)
    
    // 业务功能接口（由具体供应商实现）
    EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error)
    CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error)
    ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error)
    CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error)
    GetDriverInfo(ctx context.Context, req *dto.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error)
}
```

## 🏗️ 实现细节

### sTakeout 结构体

```go
type sTakeout struct {}
```

### 核心实现逻辑

1. **服务初始化**
   ```go
   var Takeout = new(sTakeout)
   ```

2. **订单查询实现**
   ```go
   func (s *sTakeout) Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error) {
       var takeoutJob *entity.Job
       err := dao.Job.Ctx(ctx).Where(dao.Job.Columns().ShopRefNo, shopOrderUuid).Scan(&takeoutJob)
       if err != nil {
           if e, ok := err.(*gerror.Error); ok {
               return nil, gerror.Wrap(e.Cause(), "获取外送订单失败")
           }
           return nil, gerror.Wrap(err, "获取外送订单失败")
       }
       return takeoutJob, nil
   }
   ```

3. **供应商服务路由**
   ```go
   func GetService(name consts.ProviderName) ITakeout {
       switch name {
       case consts.ProviderSkootar:
           return service.Skootar()
       default:
           return service.Skootar()
       }
   }
   ```

4. **服务注册**
   ```go
   func RegisterTakeout(i ITakeout) {
       localTakeout = i
   }
   ```

5. **服务获取**
   ```go
   func Takeout() ITakeout {
       if localTakeout == nil {
           panic("implement not found for interface ITakeout, forgot register?")
       }
       return localTakeout
   }
   ```

## 📊 数据模型

### Job 实体

外卖订单的核心数据结构，包含完整的订单信息：

```go
type Job struct {
    // 基础信息
    Id                   int64       `json:"id"`                   // 主键ID
    Uuid                 string      `json:"uuid"`                 // 外送订单UUID
    ShopRefNo            string      `json:"shopRefNo"`            // 餐馆订单参考UUID
    
    // 客户信息
    CustomerMobile       string      `json:"customerMobile"`       // 下单客户电话
    CustomerEmail        string      `json:"customerEmail"`        // 下单客户联系邮件
    
    // 供应商信息
    ProviderName         string      `json:"providerName"`         // 外送供应商：skootar,grab
    TakeoutRefNo         string      `json:"takeoutRefNo"`         // 外送系统订单号
    
    // 位置信息
    ShopLocationUuid     string      `json:"shopLocationUuid"`     // 餐馆位置信息
    ConsumerLocationUuid string      `json:"consumerLocationUuid"` // 消费者位置信息
    
    // 时间信息
    JobDate              string      `json:"jobDate"`              // 订单日期：YYYY-MM-DD
    StartTime            string      `json:"startTime"`            // 开始时间或"now"
    FinishTime           string      `json:"finishTime"`           // 订单结束时间
    
    // 业务信息
    PaymentType          string      `json:"paymentType"`          // 支付类型
    JobStatus            string      `json:"jobStatus"`            // 外送订单状态
    Remark               string      `json:"remark"`               // 订单备注
    
    // 扩展字段
    Reserved1            string      `json:"reserved1"`            // 保留字段1
    Reserved2            string      `json:"reserved2"`            // 保留字段2
    
    // 系统字段
    CreatedAt            *gtime.Time `json:"createdAt"`            // 创建时间
    UpdatedAt            *gtime.Time `json:"updatedAt"`            // 更新时间
    DeletedAt            *gtime.Time `json:"deletedAt"`            // 软删除
    
    // 回调信息
    CallbackUrl          string      `json:"callbackUrl"`          // 订单状态更新回调
    
    // 司机信息
    SkootarId            string      `json:"skootarId"`            // 骑手Id
    SkootarName          string      `json:"skootarName"`          // 骑手名称
    SkootarPhone         string      `json:"skootarPhone"`         // 骑手电话
    SkootarImageUrl      string      `json:"skootarImageUrl"`      // 骑手头像
    SkootarRating        float64     `json:"skootarRating"`        // 骑手评分
}
```

### 供应商常量

```go
type ProviderName string

const (
    ProviderSkootar ProviderName = "skootar"
    ProviderGrab    ProviderName = "grab"
)
```

## 🔄 使用流程

### 1. 基本订单查询
```go
// 获取服务实例
takeoutService := service.Takeout()

// 查询订单信息
job, err := takeoutService.Get(ctx, "shop-order-uuid-123")
if err != nil {
    // 处理错误
    return err
}

// 使用订单数据
fmt.Printf("订单UUID: %s\n", job.Uuid)
fmt.Printf("供应商: %s\n", job.ProviderName)
fmt.Printf("订单状态: %s\n", job.JobStatus)
```

### 2. 供应商服务路由
```go
// 根据供应商名称获取对应的服务
skootarService := takeout.GetService(consts.ProviderSkootar)

// 使用供应商特定的功能
resp, err := skootarService.CreateOrder(ctx, createOrderReq)
if err != nil {
    return err
}
```

### 3. 业务服务集成
```go
// 在其他业务服务中使用 Takeout 服务
func (s *OrderService) ProcessOrder(ctx context.Context, shopOrderUuid string) error {
    // 查询外卖订单
    takeoutJob, err := service.Takeout().Get(ctx, shopOrderUuid)
    if err != nil {
        return err
    }
    
    // 根据供应商处理订单
    providerService := takeout.GetService(consts.ProviderName(takeoutJob.ProviderName))
    
    // 执行业务逻辑
    return s.processProviderOrder(ctx, providerService, takeoutJob)
}
```

## ⚠️ 注意事项

1. **数据库依赖**: 服务依赖数据库中的 Job 表，确保表结构正确
2. **错误处理**: 需要妥善处理数据库查询错误和业务逻辑错误
3. **数据一致性**: 确保查询到的订单数据与实际业务状态一致
4. **供应商配置**: 供应商服务路由需要正确的配置支持
5. **软删除**: 查询时需要考虑软删除的数据过滤
6. **性能考虑**: 在高频查询场景下需要考虑数据库性能优化
7. **并发安全**: 注意并发访问时的数据一致性

## 🔮 扩展性

### 供应商扩展
1. **新供应商支持**: 通过实现 ITakeout 接口添加新的配送供应商
2. **动态路由**: 支持运行时动态配置供应商路由
3. **负载均衡**: 支持多供应商的负载均衡和故障转移
4. **供应商切换**: 支持订单在不同供应商间的切换

### 功能扩展
1. **批量查询**: 支持批量查询多个订单
2. **条件查询**: 支持更复杂的查询条件
3. **分页查询**: 支持大数据量的分页查询
4. **缓存机制**: 添加查询结果缓存提升性能
5. **数据同步**: 支持与外部系统的数据同步

### 监控扩展
1. **查询监控**: 监控查询性能和成功率
2. **数据统计**: 订单数据统计和分析
3. **异常告警**: 查询异常和错误告警
4. **性能优化**: 基于监控数据的性能优化

## 🎯 使用场景

### 订单查询场景
1. **订单详情**: 用户查看订单详细信息
2. **状态查询**: 查询订单当前状态
3. **历史订单**: 查询历史订单记录
4. **订单统计**: 基于订单数据的统计分析

### 业务集成场景
1. **订单处理**: 在订单处理流程中查询订单信息
2. **状态同步**: 与外部系统同步订单状态
3. **客户服务**: 客服查询订单信息提供服务
4. **财务结算**: 基于订单数据进行财务结算

### 运营分析场景
1. **销售分析**: 分析外卖销售数据
2. **供应商分析**: 分析不同供应商的表现
3. **用户行为**: 分析用户订餐行为
4. **效率优化**: 基于数据优化运营效率

## 🔧 服务设计模式

### 适配器模式
Takeout 服务使用了适配器模式，为不同的外卖供应商提供统一的接口：

```go
// 统一接口
type ITakeout interface {
    Get(ctx context.Context, shopOrderUuid string) (*entity.Job, error)
    // 其他统一方法...
}

// 适配器路由
func GetService(name consts.ProviderName) ITakeout {
    switch name {
    case consts.ProviderSkootar:
        return service.Skootar()  // Skootar 适配器
    case consts.ProviderGrab:
        return service.Grab()     // Grab 适配器（预留）
    default:
        return service.Skootar()   // 默认适配器
    }
}
```

### 服务定位器模式
通过服务定位器模式管理不同供应商的服务实例：

```go
// 服务注册
func RegisterSkootar(i ISkootar) {
    localSkootar = i
}

// 服务获取
func Skootar() ISkootar {
    if localSkootar == nil {
        panic("implement not found for interface ISkootar, forgot register?")
    }
    return localSkootar
}
```

## 📝 总结

Takeout 服务作为外卖系统的数据访问层，提供了统一、可靠的订单数据管理能力。

### 技术特点
- **统一接口**: 为不同供应商提供统一的数据访问接口
- **灵活路由**: 支持多供应商的服务路由和切换
- **错误处理**: 完善的错误处理和日志记录
- **扩展性强**: 预留了新供应商的扩展能力

### 设计优势
- **解耦合**: 业务逻辑与具体供应商实现解耦
- **可维护**: 清晰的接口定义和实现分离
- **可测试**: 接口抽象便于单元测试和模拟
- **高性能**: 直接数据库访问，性能优异

### 业务价值
- **数据统一**: 提供统一的订单数据访问入口
- **供应商无关**: 业务逻辑不依赖特定供应商
- **易于扩展**: 新增供应商成本低
- **运维友好**: 统一的监控和运维接口

该服务是外卖系统数据层的核心组件，为整个系统提供了稳定、高效的数据访问能力，是系统架构中的重要基础设施。
