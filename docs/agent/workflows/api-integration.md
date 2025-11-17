# API 对接工作流（后端版）

> 本文档定义第三方 API 集成的完整流程

---

## 📋 概述

### 适用场景

- 对接第三方支付（微信支付、支付宝）
- 集成云服务（OSS、SMS、推送）
- 接入 SaaS 平台（ERP、CRM）
- 开发内部 API 接口（Go REST/gRPC）

### 预计时间

- 简单 API (RESTful): 1-2 天
- 复杂 API (OAuth/Webhook): 2-3 天
- gRPC 服务: 2-4 天

---

## 完整流程

```
收到对接需求 → API 调研 → 技术方案设计 → 选择技术栈 →
创建 API 封装 → 实现认证逻辑 → 实现业务接口 →
编写测试 → 文档更新 → 上线监控
```

---

## 执行流程

### Step 1: API 调研 (2-4 小时)

#### 阅读官方文档

- [ ] API 总览和架构
- [ ] 认证方式（OAuth2/API Key/JWT）
- [ ] 接口列表和参数说明
- [ ] 响应格式和错误码
- [ ] 速率限制和配额
- [ ] SDK/示例代码

#### 搜索历史经验

```
query: "{第三方服务名} API integration Go PHP"
group_id: "ttpos-integrations"
```

#### 整理 API 清单

```markdown
## 需要对接的接口

| 接口名称 | HTTP 方法 | 路径           | 用途         | 优先级 | 语言 |
| -------- | --------- | -------------- | ------------ | ------ | ---- |
| 创建订单 | POST      | /v1/orders     | 创建支付订单 | P0     | Go   |
| 查询订单 | GET       | /v1/orders/:id | 查询订单状态 | P0     | Go   |
```

---

### Step 2: 技术方案设计 (1-2 小时)

#### 确定技术栈

```yaml
IF 主业务服务 (订单/支付/会员) THEN
  语言: Go
  位置: main/app/service/
  框架: Gin + GORM

ELSE IF 管理功能/微服务 THEN
  语言: Go (GoFrame)
  位置: ttpos-bmp/app/ttpos-*/
  框架: GoFrame + gRPC

ELSE IF 后台管理 THEN
  语言: PHP
  位置: admin/app/
  框架: ThinkPHP
```

#### 设计类结构（Go 示例）

```
main/app/service/integrations/{service_name}/
├── {service_name}_service.go      # 服务类
├── {service_name}_auth.go         # 认证逻辑
├── {service_name}_config.go       # 配置管理
├── dto/                           # 数据传输对象
│   ├── request/
│   └── response/
└── errors/                        # 异常定义
```

#### 定义接口规范（Go 示例）

```go
type IPaymentService interface {
    // 创建支付订单
    CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.OrderResp, error)

    // 查询订单状态
    QueryOrder(ctx context.Context, orderNo string) (*resp.OrderStatusResp, error)

    // 申请退款
    Refund(ctx context.Context, req req.RefundReq) (bool, error)
}
```

---

### Step 3: 创建 API 封装

#### Go 服务示例

```go
// {service_name}_service.go
package payment

import (
    "context"
    "net/http"
    "ttpos-server-go/pkg/logger"
)

type PaymentServiceImpl struct {
    config *Config
    client *http.Client
}

func NewPaymentService(config *Config) IPaymentService {
    return &PaymentServiceImpl{
        config: config,
        client: &http.Client{
            Timeout: time.Duration(config.Timeout) * time.Second,
        },
    }
}
```

#### PHP 服务示例

```php
// {ServiceName}Service.php
namespace app\service\integrations\payment;

class PaymentService {
    private $config;
    private $client;

    public function __construct($config) {
        $this->config = $config;
        $this->client = new \GuzzleHttp\Client([
            'timeout' => $config['timeout'] ?? 30,
            'base_uri' => $config['base_url'],
        ]);
    }
}
```

---

### Step 4: 实现认证逻辑

#### API Key 认证（Go）

```go
func (s *PaymentServiceImpl) addAuthHeaders(req *http.Request) {
    req.Header.Set("X-API-Key", s.config.APIKey)
    req.Header.Set("X-API-Secret", s.config.APISecret)
}
```

#### OAuth2 认证（Go）

```go
func (s *PaymentServiceImpl) GetAccessToken() (string, error) {
    // 检查 Token 是否过期
    if s.accessToken != "" && time.Now().Before(s.expiresAt) {
        return s.accessToken, nil
    }

    // 刷新 Token
    return s.refreshAccessToken()
}
```

---

### Step 5: 实现业务接口

#### Go 实现示例

```go
func (s *PaymentServiceImpl) CreateOrder(ctx context.Context, req req.CreateOrderReq) (*resp.OrderResp, error) {
    // 验证参数
    if err := req.Validate(); err != nil {
        return nil, errors.WithMessage(err, "参数验证失败")
    }

    // 构造请求
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/v1/orders", bytes.NewBuffer(body))
    s.addAuthHeaders(httpReq)

    // 发送请求
    httpResp, err := s.client.Do(httpReq)
    if err != nil {
        logger.Logger.Error("CreateOrder Error", zap.Error(err))
        return nil, errors.WithMessage(err, "请求失败")
    }
    defer httpResp.Body.Close()

    // 解析响应
    var orderResp resp.OrderResp
    if err := json.NewDecoder(httpResp.Body).Decode(&orderResp); err != nil {
        return nil, errors.WithMessage(err, "解析响应失败")
    }

    return &orderResp, nil
}
```

#### gRPC 服务实现（ttpos-bmp）

```go
// 定义 Protobuf (manifest/protobuf/payment.proto)
service PaymentService {
    rpc CreateOrder(CreateOrderRequest) returns (OrderResponse);
    rpc QueryOrder(QueryOrderRequest) returns (OrderStatusResponse);
}

// 实现 gRPC 服务 (internal/controller/rpc/payment.go)
func (s *PaymentController) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
    // 调用 logic 层
    order, err := s.logic.CreateOrder(ctx, req)
    if err != nil {
        return nil, err
    }

    return &pb.OrderResponse{
        OrderNo: order.OrderNo,
        Status:  order.Status,
    }, nil
}
```

---

### Step 6: 编写测试

#### Go 单元测试

```go
func TestPaymentService_CreateOrder(t *testing.T) {
    // Given
    service := NewPaymentService(&Config{
        APIKey:    "test_key",
        APISecret: "test_secret",
        BaseURL:   "https://api.test.com",
    })

    req := req.CreateOrderReq{
        Amount:  10000,
        OrderID: "ORDER_123",
    }

    // When
    order, err := service.CreateOrder(context.Background(), req)

    // Then
    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.NotEmpty(t, order.OrderNo)
}
```

#### 测试覆盖率要求

- Go Service: ≥70%
- PHP Service: ≥70%
- gRPC Service: ≥70%

---

### Step 7: 文档更新

#### 创建集成文档

```bash
mkdir -p docs/shared/integrations/{service_name}
cd docs/shared/integrations/{service_name}
touch README.md setup.md api-reference.md troubleshooting.md
```

#### README.md 结构

```markdown
# {Service Name} 集成

## 概述

集成 {Service Name} 提供 {功能描述}。

## 技术栈

- **主服务**: Go (main/)
- **微服务**: Go (ttpos-bmp/)
- **后台**: PHP (admin/)

## 快速开始

### Go 配置

\`\`\`go
config := &payment.Config{
APIKey: os.Getenv("PAYMENT_API_KEY"),
APISecret: os.Getenv("PAYMENT_API_SECRET"),
BaseURL: "https://api.payment.com",
}

service := payment.NewPaymentService(config)
\`\`\`

### PHP 配置

\`\`\`php
$config = [
'api_key' => env('PAYMENT_API_KEY'),
'api_secret' => env('PAYMENT_API_SECRET'),
'base_url' => 'https://api.payment.com',
];

$service = new PaymentService($config);
\`\`\`

## 相关文档

- [配置指南](./setup.md)
- [API 参考](./api-reference.md)
- [问题排查](./troubleshooting.md)
```

---

### Step 8: 上线和监控

#### 配置监控（Go）

```go
// 记录关键指标
logger.Logger.Info("API Call",
    zap.String("service", "payment"),
    zap.String("method", "CreateOrder"),
    zap.Int64("duration_ms", duration.Milliseconds()),
    zap.String("status", "success"),
)
```

#### 记录到 Graphiti

```yaml
name: "integration-{service-name}-{YYYY-MM}"
group_id: "ttpos-integrations"
episode_body: |
  服务: {Service Name}

  集成内容:
  - 创建订单
  - 查询订单

  技术栈:
  - Go (main/)
  - 认证方式: OAuth2

  注意事项:
  - API 限流 1000次/分钟
  - Token 有效期 2 小时

  相关文档: docs/shared/integrations/{service_name}/
```

---

## 检查清单

### 调研阶段

- [ ] API 文档已阅读
- [ ] 认证方式已确认
- [ ] 接口清单已整理
- [ ] 技术栈已确定（Go/PHP）

### 开发阶段

- [ ] 目录结构已创建
- [ ] 配置类已实现
- [ ] 认证逻辑已实现
- [ ] 业务接口已实现
- [ ] 错误处理完整
- [ ] 重试机制已配置（如需要）

### 测试阶段

- [ ] 单元测试已编写
- [ ] 测试覆盖率 ≥ 70%
- [ ] 沙箱环境测试通过

### 文档阶段

- [ ] 集成文档已创建
- [ ] 配置指南已完成
- [ ] API 参考已完成

### 上线阶段

- [ ] 灰度发布完成
- [ ] 监控配置完成
- [ ] 经验已记录

---

## 常见问题

### Q: Go 和 PHP 如何选择？

**A**:

- 主业务服务（订单、支付、会员）→ Go (main/)
- 微服务模块 → Go (ttpos-bmp/)
- 后台管理功能 → PHP (admin/)

### Q: 如何处理 API 限流？

**A**:

1. 查看 API 文档的限流说明
2. 添加请求频率控制
3. 实现指数退避重试

### Q: gRPC 和 REST 如何选择？

**A**:

- 内部服务间调用 → gRPC (ttpos-bmp)
- 外部 API 对接 → REST
- 第三方服务 → 遵循官方文档

---

## 相关资源

### 规范文件

- `.cursor/rules/go-main.mdc` - Go 规范
- `.cursor/rules/php.mdc` - PHP 规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 规范

### 工作流

- [微服务集成工作流](./microservice-integration.md) - gRPC 服务开发

### 文档

- `docs/shared/integrations/` - 集成文档目录
- `docs/shared/api/` - API 文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 建议在完成第三方对接/联调复盘后沉淀 Episode，记录认证、限流、回调等经验。

---

**最后更新**: 2025-11-16  
**维护者**: 后端开发组
