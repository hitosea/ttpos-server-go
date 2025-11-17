# 微服务集成工作流

> 本文档定义 ttpos-bmp 微服务开发和集成的完整流程

---

## 📋 概述

### 适用场景
- 开发 gRPC 微服务
- 注册服务到 Nacos
- 主服务调用微服务
- 微服务间通信

### 预计时间
- 简单 gRPC 服务: 2-3 天
- 复杂微服务（含业务逻辑）: 3-5 天

### 技术栈
- **框架**: GoFrame v2.x
- **协议**: gRPC + REST
- **服务发现**: Nacos
- **消息队列**: RocketMQ

---

## 完整流程

```
定义 Protobuf → 生成代码 → 实现 gRPC 服务 → 注册到 Nacos →
编写客户端调用 → 测试服务 → 创建 API 文档 → 部署验证
```

---

## 执行流程

### Step 1: 定义 Protobuf (30-60 分钟)

#### 创建 Proto 文件

```bash
cd /Users/ben/projects/ttpos-server-go/ttpos-bmp/app/ttpos-erp

# 创建 Protobuf 文件
mkdir -p manifest/protobuf
touch manifest/protobuf/order.proto
```

#### Proto 文件示例

```protobuf
syntax = "proto3";

package order.v1;

option go_package = "ttpos-bmp/app/ttpos-erp/api/order/v1";

// 订单服务
service OrderService {
  // 创建订单
  rpc CreateOrder(CreateOrderRequest) returns (OrderResponse);
  // 查询订单
  rpc GetOrder(GetOrderRequest) returns (OrderResponse);
  // 更新订单
  rpc UpdateOrder(UpdateOrderRequest) returns (OrderResponse);
}

// 创建订单请求
message CreateOrderRequest {
  string order_no = 1;  // 订单号
  double amount = 2;    // 金额
  int32 status = 3;     // 状态
}

// 订单响应
message OrderResponse {
  int64 id = 1;         // 订单ID
  string order_no = 2;  // 订单号
  double amount = 3;    // 金额
  int32 status = 4;     // 状态
  int64 create_time = 5;// 创建时间
}

// 查询订单请求
message GetOrderRequest {
  int64 id = 1;         // 订单ID
}

// 更新订单请求
message UpdateOrderRequest {
  int64 id = 1;         // 订单ID
  int32 status = 2;     // 状态
}
```

#### Protobuf 规范
- 使用 snake_case 命名字段
- 消息名使用 PascalCase
- 服务名以 Service 结尾
- 添加中文注释说明

参考: `ttpos-bmp/.cursor/rules/proto-rules.mdc`

---

### Step 2: 生成代码 (5-10 分钟)

#### 执行代码生成

```bash
cd /Users/ben/projects/ttpos-server-go/ttpos-bmp

# 生成 Protobuf 代码
gf gen pb

# 生成 DAO (数据访问层)
gf gen dao

# 生成 Service 接口
gf gen service
```

#### 生成的文件

```
ttpos-bmp/app/ttpos-erp/
├── api/order/v1/               # gRPC API
│   ├── order.pb.go            # Protobuf 生成
│   └── order_grpc.pb.go       # gRPC 生成
├── internal/
│   ├── controller/
│   │   ├── http/              # HTTP 控制器
│   │   └── rpc/               # gRPC 控制器
│   ├── logic/                 # 业务逻辑
│   ├── dao/                   # 数据访问层（自动生成）
│   └── model/
│       ├── entity/            # 数据实体（自动生成）
│       ├── do/                # 数据对象（自动生成）
│       └── dto/               # 数据传输对象
└── internal/service/          # 服务接口（自动生成）
```

---

### Step 3: 实现 gRPC 服务 (2-4 小时)

#### 创建 gRPC 控制器

```go
// internal/controller/rpc/order.go
package rpc

import (
    "context"
    "ttpos-bmp/app/ttpos-erp/api/order/v1"
    "ttpos-bmp/app/ttpos-erp/internal/logic"
)

type OrderController struct {
    order.UnimplementedOrderServiceServer
    logic *logic.OrderLogic
}

func NewOrderController() *OrderController {
    return &OrderController{
        logic: logic.NewOrderLogic(),
    }
}

// CreateOrder 创建订单
func (c *OrderController) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.OrderResponse, error) {
    // 调用 logic 层
    result, err := c.logic.CreateOrder(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &order.OrderResponse{
        Id:         result.Id,
        OrderNo:    result.OrderNo,
        Amount:     result.Amount,
        Status:     result.Status,
        CreateTime: result.CreateTime,
    }, nil
}

// GetOrder 查询订单
func (c *OrderController) GetOrder(ctx context.Context, req *order.GetOrderRequest) (*order.OrderResponse, error) {
    result, err := c.logic.GetOrder(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return &order.OrderResponse{
        Id:         result.Id,
        OrderNo:    result.OrderNo,
        Amount:     result.Amount,
        Status:     result.Status,
        CreateTime: result.CreateTime,
    }, nil
}
```

#### 实现业务逻辑层

```go
// internal/logic/order.go
package logic

import (
    "context"
    "github.com/gogf/gf/v2/errors/gerror"
    "ttpos-bmp/app/ttpos-erp/api/order/v1"
    "ttpos-bmp/app/ttpos-erp/internal/dao"
    "ttpos-bmp/app/ttpos-erp/internal/model/do"
)

type OrderLogic struct {
    dao *dao.OrderDao
}

func NewOrderLogic() *OrderLogic {
    return &OrderLogic{
        dao: dao.Order,
    }
}

// CreateOrder 创建订单
func (l *OrderLogic) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.OrderResponse, error) {
    // 验证参数
    if req.OrderNo == "" {
        return nil, gerror.New("订单号不能为空")
    }
    
    // 创建订单
    id, err := l.dao.Insert(ctx, do.Order{
        OrderNo: req.OrderNo,
        Amount:  req.Amount,
        Status:  req.Status,
    })
    
    if err != nil {
        return nil, gerror.Wrap(err, "创建订单失败")
    }
    
    // 查询返回
    result, err := l.dao.FindById(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return &order.OrderResponse{
        Id:         result.Id,
        OrderNo:    result.OrderNo,
        Amount:     result.Amount,
        Status:     int32(result.Status),
        CreateTime: result.CreateTime,
    }, nil
}
```

---

### Step 4: 注册到 Nacos (15-30 分钟)

#### 配置 Nacos

```yaml
# manifest/config/config.yaml
nacos:
  address: "127.0.0.1:8848"
  namespace: "ttpos"
  group: "DEFAULT_GROUP"

grpc:
  name: "ttpos-erp"         # 服务名称
  address: ":9000"          # gRPC 端口
```

#### 注册服务

```go
// internal/boot/grpc.go
package boot

import (
    "github.com/gogf/gf/v2/frame/g"
    "google.golang.org/grpc"
    "ttpos-bmp/app/ttpos-erp/api/order/v1"
    "ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
)

func InitGrpc() {
    // 创建 gRPC 服务器
    s := grpc.NewServer()
    
    // 注册服务
    order.RegisterOrderServiceServer(s, rpc.NewOrderController())
    
    // 启动服务
    address := g.Cfg().MustGet(ctx, "grpc.address").String()
    lis, err := net.Listen("tcp", address)
    if err != nil {
        g.Log().Fatalf(ctx, "failed to listen: %v", err)
    }
    
    g.Log().Infof(ctx, "gRPC server listening on %s", address)
    
    // 注册到 Nacos
    registerToNacos()
    
    if err := s.Serve(lis); err != nil {
        g.Log().Fatalf(ctx, "failed to serve: %v", err)
    }
}

func registerToNacos() {
    // 实现 Nacos 注册逻辑
}
```

---

### Step 5: 编写客户端调用 (30-60 分钟)

#### 主服务调用微服务

```go
// main/app/service/order_client.go
package service

import (
    "context"
    "google.golang.org/grpc"
    "ttpos-bmp/app/ttpos-erp/api/order/v1"
)

type OrderClient struct {
    client order.OrderServiceClient
}

func NewOrderClient() *OrderClient {
    // 从 Nacos 获取服务地址
    address := discoverService("ttpos-erp")
    
    // 创建 gRPC 连接
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    
    return &OrderClient{
        client: order.NewOrderServiceClient(conn),
    }
}

// CreateOrder 调用微服务创建订单
func (c *OrderClient) CreateOrder(ctx context.Context, orderNo string, amount float64) (*order.OrderResponse, error) {
    req := &order.CreateOrderRequest{
        OrderNo: orderNo,
        Amount:  amount,
        Status:  1,
    }
    
    resp, err := c.client.CreateOrder(ctx, req)
    if err != nil {
        logger.Logger.Error("CreateOrder Error", zap.Error(err))
        return nil, err
    }
    
    return resp, nil
}
```

---

### Step 6: 测试服务 (1-2 小时)

#### 单元测试

```go
// internal/logic/order_test.go
package logic

import (
    "context"
    "testing"
    "ttpos-bmp/app/ttpos-erp/api/order/v1"
)

func TestOrderLogic_CreateOrder(t *testing.T) {
    logic := NewOrderLogic()
    
    req := &order.CreateOrderRequest{
        OrderNo: "TEST001",
        Amount:  100.50,
        Status:  1,
    }
    
    resp, err := logic.CreateOrder(context.Background(), req)
    
    if err != nil {
        t.Errorf("CreateOrder failed: %v", err)
    }
    
    if resp.OrderNo != "TEST001" {
        t.Errorf("Expected order_no TEST001, got %s", resp.OrderNo)
    }
}
```

#### 集成测试

```bash
# 启动 ttpos-bmp 服务
cd ttpos-bmp/app/ttpos-erp
gf run main.go

# 测试 gRPC 调用
cd main
go test ./app/service/order_client_test.go -v
```

---

### Step 7: 创建 API 文档 (30-60 分钟)

#### 创建文档目录

```bash
mkdir -p /Users/ben/projects/ttpos-server-go/docs/shared/api/grpc
touch docs/shared/api/grpc/order-service.md
```

#### API 文档模板

```markdown
# Order Service API

## 概述
订单微服务，提供订单的创建、查询、更新功能。

## 服务信息
- **服务名**: ttpos-erp
- **协议**: gRPC
- **端口**: 9000
- **Nacos 注册**: ttpos/DEFAULT_GROUP

## 接口列表

### CreateOrder - 创建订单

**请求**:
\`\`\`protobuf
message CreateOrderRequest {
  string order_no = 1;  // 订单号（必填）
  double amount = 2;    // 金额（必填，>0）
  int32 status = 3;     // 状态（必填，1=待支付）
}
\`\`\`

**响应**:
\`\`\`protobuf
message OrderResponse {
  int64 id = 1;         // 订单ID
  string order_no = 2;  // 订单号
  double amount = 3;    // 金额
  int32 status = 4;     // 状态
  int64 create_time = 5;// 创建时间
}
\`\`\`

**Go 调用示例**:
\`\`\`go
client := order.NewOrderServiceClient(conn)
resp, err := client.CreateOrder(ctx, &order.CreateOrderRequest{
    OrderNo: "ORD001",
    Amount:  100.50,
    Status:  1,
})
\`\`\`

**错误码**:
- `InvalidArgument`: 参数错误
- `AlreadyExists`: 订单号已存在
- `Internal`: 内部错误

## 相关文档
- [Protobuf 定义](../../../ttpos-bmp/app/ttpos-erp/manifest/protobuf/order.proto)
- [服务实现](../../../ttpos-bmp/app/ttpos-erp/internal/logic/order.go)
```

---

### Step 8: 部署验证 (30-60 分钟)

#### 部署检查清单
- [ ] Protobuf 已生成
- [ ] gRPC 服务已实现
- [ ] Nacos 注册成功
- [ ] 客户端调用成功
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] API 文档已创建

#### 监控配置
```go
// 记录 gRPC 调用指标
g.Log().Info(ctx, "gRPC Call",
    g.Map{
        "service": "OrderService",
        "method":  "CreateOrder",
        "duration_ms": duration.Milliseconds(),
        "status": "success",
    },
)
```

---

## 检查清单

### Protobuf 定义
- [ ] Proto 文件已创建
- [ ] 字段使用 snake_case
- [ ] 添加了中文注释
- [ ] 代码已生成

### gRPC 服务
- [ ] gRPC 控制器已实现
- [ ] 业务逻辑层已实现
- [ ] 数据访问层已实现
- [ ] 错误处理完整

### 服务注册
- [ ] Nacos 配置正确
- [ ] 服务注册成功
- [ ] 服务发现正常

### 客户端调用
- [ ] 客户端代码已实现
- [ ] 服务发现正常
- [ ] 调用测试成功

### 测试和文档
- [ ] 单元测试已编写
- [ ] 集成测试通过
- [ ] API 文档已创建

---

## 常见问题

### Q: gRPC 和 REST 如何选择？
**A**: 
- 内部服务间调用 → gRPC (高性能)
- 外部 API 调用 → REST (兼容性好)

### Q: 如何调试 gRPC 服务？
**A**: 
1. 使用 grpcurl 工具
2. 使用 Postman（支持 gRPC）
3. 使用 BloomRPC 客户端

### Q: Nacos 注册失败怎么办？
**A**: 
1. 检查 Nacos 服务是否启动
2. 检查配置是否正确
3. 检查网络连接
4. 查看日志排查错误

---

## 相关资源

### 规范文件
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 规范 ⭐⭐⭐
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范 ⭐⭐⭐

### 工作流
- [API 对接工作流](./api-integration.md)

### 模板
- `docs/agent/templates/grpc-service-template.md` - gRPC 服务模板

---

**最后更新**: 2025-11-16  
**维护者**: 后端开发组

