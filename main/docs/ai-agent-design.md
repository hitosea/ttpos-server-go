# TTPOS AI Agent 设计方案

## 一、概述

### 1.1 项目背景

TTPOS 餐饮系统需要集成 AI 能力，让用户通过自然语言描述操作需求，由 AI 引擎分析后自动调用系统功能完成任务。

### 1.2 目标

- 提供自然语言交互接口，降低系统操作门槛
- 基于 Claude Tool Use 实现智能工具调用
- 可扩展的工具注册机制，便于接入现有业务服务
- 支持多轮对话和上下文理解

### 1.3 技术选型

| 组件 | 技术方案 | 说明 |
|------|----------|------|
| AI 引擎 | Claude API (Anthropic) | 支持 Tool Use，理解能力强 |
| 后端框架 | Go + Gin | 与现有系统一致 |
| 通信协议 | HTTP/JSON | RESTful API |
| 流式响应 | SSE (Server-Sent Events) | 可选，提升用户体验 |

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              前端应用                                        │
│                      (Flutter / Web / 小程序)                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼ HTTP POST /api/v1/ai/chat
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API Gateway                                     │
│                           (Gin + Middleware)                                 │
│                      认证、限流、日志、链路追踪                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Agent Service                                     │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         Orchestrator (调度器)                         │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐   │   │
│  │  │ 消息管理    │───▶│ Claude API  │───▶│ 响应解析 & 工具路由     │   │   │
│  │  │ (历史记录)  │    │   调用      │    │                         │   │   │
│  │  └─────────────┘    └─────────────┘    └───────────┬─────────────┘   │   │
│  │                                                     │                 │   │
│  │                          ┌──────────────────────────┘                 │   │
│  │                          ▼                                            │   │
│  │               ┌─────────────────────┐                                 │   │
│  │               │   Tool Executor     │                                 │   │
│  │               │   (工具执行器)       │                                 │   │
│  │               └──────────┬──────────┘                                 │   │
│  │                          │                                            │   │
│  └──────────────────────────┼────────────────────────────────────────────┘   │
│                             ▼                                                │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                       Tool Registry (工具注册表)                       │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │   │
│  │  │ 订单工具 │ │ 会员工具 │ │ 商品工具 │ │ 报表工具 │ │ 系统工具 │   │   │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘   │   │
│  │       │            │            │            │            │          │   │
│  └───────┼────────────┼────────────┼────────────┼────────────┼──────────┘   │
└──────────┼────────────┼────────────┼────────────┼────────────┼──────────────┘
           │            │            │            │            │
           ▼            ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         现有业务服务层 (Service)                             │
│     OrderSrv │ MemberSrv │ ProductSrv │ StatisticsSrv │ SettingSrv │ ...   │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              数据层                                          │
│                    MySQL (多租户) │ Redis │ RocketMQ                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 核心流程

```
┌──────┐      ┌──────────┐      ┌─────────────┐      ┌──────────────┐
│ 用户 │─────▶│ Agent API │─────▶│ Orchestrator│─────▶│ Claude API   │
└──────┘      └──────────┘      └─────────────┘      └──────┬───────┘
                                       ▲                     │
                                       │                     ▼
                                       │              ┌──────────────┐
                                       │              │ 返回 tool_use │
                                       │              └──────┬───────┘
                                       │                     │
                                       │                     ▼
                                ┌──────┴───────┐      ┌──────────────┐
                                │ 工具执行结果  │◀─────│ Tool Executor│
                                └──────────────┘      └──────────────┘
                                       │
                                       ▼
                                ┌──────────────┐
                                │ 返回给 Claude │
                                └──────┬───────┘
                                       │
                                       ▼
                                ┌──────────────┐
                                │ 最终回复用户  │
                                └──────────────┘
```

### 2.3 多轮工具调用流程

```
Round 1:
  User: "帮我查一下订单 ORD001 的详情，并计算总金额的税费（8%）"

  Claude 分析 → 需要先查询订单
  Claude 返回: tool_use(query_order, {order_no: "ORD001"})

  Executor 执行 → 返回订单信息 {total_amount: 500}

Round 2:
  Claude 收到订单信息 → 需要计算税费
  Claude 返回: tool_use(calculator, {operation: "multiply", a: 500, b: 0.08})

  Executor 执行 → 返回计算结果 {result: 40}

Round 3:
  Claude 收到计算结果 → 生成最终回复
  Claude 返回: "订单 ORD001 总金额 500 元，8% 税费为 40 元。"

  → 返回给用户
```

---

## 三、模块设计

### 3.1 目录结构

```
app/modules/agent/
├── api/                          # API 层
│   └── agent_api.go              # HTTP 接口处理
│
├── service/                      # 服务层
│   ├── agent_service.go          # Agent 核心服务
│   └── conversation_service.go   # 会话管理服务（可选）
│
├── engine/                       # Agent 引擎
│   └── tool/                     # Tool 框架
│       ├── interface.go          # Tool 接口定义
│       ├── registry.go           # 工具注册表
│       └── executor.go           # 工具执行器
│
├── tools/                        # 工具实现
│   ├── order/                    # 订单工具
│   │   ├── query_order.go
│   │   ├── cancel_order.go
│   │   └── modify_order.go
│   ├── member/                   # 会员工具
│   │   ├── query_member.go
│   │   └── member_points.go
│   ├── product/                  # 商品工具
│   │   └── query_product.go
│   ├── report/                   # 报表工具
│   │   └── sales_report.go
│   ├── demo_tools.go             # 演示工具
│   └── register.go               # 统一注册
│
├── dto/                          # 数据传输对象
│   ├── req/
│   │   └── agent_req.go
│   └── resp/
│       └── agent_resp.go
│
├── model/                        # 数据模型（可选）
│   ├── conversation.go           # 会话模型
│   └── message.go                # 消息模型
│
├── repository/                   # 数据访问层（可选）
│   └── conversation_repo.go
│
└── agent.go                      # 模块入口

pkg/claude/                       # Claude SDK（公共包）
├── client.go                     # API 客户端
├── types.go                      # 类型定义
└── stream.go                     # 流式处理（可选）
```

### 3.2 核心接口定义

#### 3.2.1 Tool 接口

```go
// Tool 定义了一个可被 AI 调用的工具接口
type Tool interface {
    // Name 返回工具名称，用于 Claude API 注册
    // 命名规范：snake_case，如 query_order, cancel_order
    Name() string

    // Description 返回工具描述
    // 要求：清晰说明功能、参数、使用场景
    // AI 根据此描述决定何时调用该工具
    Description() string

    // InputSchema 返回 JSON Schema
    // 定义工具的输入参数结构
    InputSchema() map[string]any

    // Execute 执行工具逻辑
    // ctx: 包含用户信息、权限、数据库连接等
    // input: AI 传入的参数
    Execute(ctx context.Context, input map[string]any) (any, error)
}
```

#### 3.2.2 Agent Service 接口

```go
// IAgentSrv Agent 服务接口
type IAgentSrv interface {
    // Chat 同步对话
    Chat(ctx context.Context, message string) (*resp.ChatResp, error)

    // ChatStream 流式对话（可选）
    ChatStream(ctx context.Context, message string, ch chan<- *StreamEvent) error
}
```

### 3.3 数据结构

#### 3.3.1 请求结构

```go
// ChatReq 对话请求
type ChatReq struct {
    ConversationId string `json:"conversation_id,omitempty"` // 会话ID（多轮对话）
    Message        string `json:"message" binding:"required"` // 用户消息
}
```

#### 3.3.2 响应结构

```go
// ChatResp 对话响应
type ChatResp struct {
    ConversationId string               `json:"conversation_id,omitempty"`
    Content        string               `json:"content"`
    ToolCalls      []tool.ExecuteResult `json:"tool_calls,omitempty"`
    FinishReason   string               `json:"finish_reason"`
    Usage          *UsageInfo           `json:"usage,omitempty"`
}

// ExecuteResult 工具执行结果
type ExecuteResult struct {
    ToolName string `json:"tool_name"`
    Input    any    `json:"input"`
    Output   any    `json:"output"`
    Error    string `json:"error,omitempty"`
    Success  bool   `json:"success"`
}
```

---

## 四、API 设计

### 4.1 接口列表

| Method | Path | 说明 | 认证 |
|--------|------|------|------|
| POST | `/api/v1/ai/chat` | 对话接口 | 需要 |
| GET | `/api/v1/ai/chat/stream` | SSE 流式对话 | 需要 |
| GET | `/api/v1/ai/health` | 健康检查 | 不需要 |
| GET | `/api/v1/ai/conversations` | 会话列表 | 需要 |
| GET | `/api/v1/ai/conversation/:id` | 会话详情 | 需要 |
| DELETE | `/api/v1/ai/conversation/:id` | 删除会话 | 需要 |

### 4.2 对话接口详情

#### POST /api/v1/ai/chat

**请求：**
```json
{
  "conversation_id": "conv_123",  // 可选，续接对话
  "message": "帮我查一下今天的订单"
}
```

**响应：**
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "conversation_id": "conv_123",
    "content": "今天共有 15 笔订单，总销售额 3,256.50 元。其中已完成 12 笔，待处理 3 笔。",
    "tool_calls": [
      {
        "tool_name": "query_order",
        "input": {
          "start_time": 1708819200,
          "end_time": 1708905600
        },
        "output": {
          "total_count": 15,
          "total_amount": 3256.50,
          "orders": [...]
        },
        "success": true
      }
    ],
    "finish_reason": "end_turn",
    "usage": {
      "input_tokens": 256,
      "output_tokens": 128,
      "total_tokens": 384
    }
  }
}
```

### 4.3 SSE 流式响应格式

```
event: thinking
data: {"content": "让我查询一下订单信息..."}

event: tool_start
data: {"tool_name": "query_order", "input": {...}}

event: tool_end
data: {"tool_name": "query_order", "success": true, "output": {...}}

event: content_delta
data: {"delta": "今天共有"}

event: content_delta
data: {"delta": " 15 笔订单"}

event: done
data: {"conversation_id": "conv_123", "finish_reason": "end_turn"}
```

---

## 五、Tool 设计规范

### 5.1 Tool 命名规范

| 类型 | 命名格式 | 示例 |
|------|----------|------|
| 查询类 | `query_{entity}` | `query_order`, `query_member` |
| 创建类 | `create_{entity}` | `create_order` |
| 修改类 | `update_{entity}` | `update_order_status` |
| 删除类 | `cancel_{entity}` / `delete_{entity}` | `cancel_order` |
| 操作类 | `{action}_{entity}` | `add_member_points` |
| 统计类 | `{entity}_report` / `{entity}_statistics` | `sales_report` |

### 5.2 Description 编写规范

Description 是 AI 判断是否调用工具的关键，需要：

1. **明确说明功能**：工具做什么
2. **列出参数说明**：每个参数的含义和格式
3. **说明使用场景**：什么情况下应该使用
4. **给出示例**（可选）：典型的调用场景

**示例：**
```go
func (t *QueryOrderTool) Description() string {
    return `查询订单信息。可以根据订单号、时间范围、状态等条件查询。

支持的查询条件：
- order_no: 订单号（精确匹配，查询单个订单）
- start_time/end_time: 时间范围（Unix 时间戳）
- status: 订单状态（pending/paid/completed/cancelled）
- customer_name: 客户姓名（模糊匹配）
- limit: 返回数量限制，默认10，最大50

使用场景：
- 用户询问"查一下订单xxx"时使用 order_no 参数
- 用户询问"今天的订单"时使用时间范围参数
- 用户询问"已完成的订单"时使用 status 参数`
}
```

### 5.3 InputSchema 规范

使用标准 JSON Schema 格式：

```go
func (t *QueryOrderTool) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "order_no": map[string]any{
                "type":        "string",
                "description": "订单号，精确匹配",
            },
            "start_time": map[string]any{
                "type":        "integer",
                "description": "开始时间（Unix 时间戳）",
            },
            "end_time": map[string]any{
                "type":        "integer",
                "description": "结束时间（Unix 时间戳）",
            },
            "status": map[string]any{
                "type":        "string",
                "enum":        []string{"pending", "paid", "completed", "cancelled"},
                "description": "订单状态",
            },
            "limit": map[string]any{
                "type":        "integer",
                "description": "返回数量限制",
                "default":     10,
                "minimum":     1,
                "maximum":     50,
            },
        },
        // 必填参数
        // "required": []string{"order_no"},
    }
}
```

### 5.4 Execute 实现规范

```go
func (t *QueryOrderTool) Execute(ctx context.Context, input map[string]any) (any, error) {
    // 1. 参数提取和验证
    orderNo, _ := input["order_no"].(string)
    limit := 10
    if l, ok := input["limit"].(float64); ok {  // JSON 数字默认是 float64
        limit = int(l)
    }

    // 2. 调用业务服务
    orders, err := t.orderSrv.QueryOrders(ctx, QueryOptions{
        OrderNo: orderNo,
        Limit:   limit,
    })
    if err != nil {
        return nil, fmt.Errorf("查询订单失败: %w", err)
    }

    // 3. 返回结构化结果（会序列化为 JSON 返回给 Claude）
    return map[string]any{
        "total_count": len(orders),
        "orders":      orders,
    }, nil
}
```

---

## 六、业务 Tool 规划

### 6.1 第一期（MVP）

| Tool 名称 | 功能 | 依赖服务 |
|-----------|------|----------|
| `query_order` | 查询订单 | OrderSrv |
| `query_member` | 查询会员 | MemberSrv |
| `get_current_time` | 获取当前时间 | - |
| `calculator` | 数学计算 | - |

### 6.2 第二期

| Tool 名称 | 功能 | 依赖服务 |
|-----------|------|----------|
| `cancel_order` | 取消订单 | OrderSrv |
| `add_member_points` | 添加会员积分 | MemberSrv |
| `deduct_member_points` | 扣除会员积分 | MemberSrv |
| `query_product` | 查询商品 | ProductSrv |
| `update_product_stock` | 更新库存 | ProductSrv |

### 6.3 第三期

| Tool 名称 | 功能 | 依赖服务 |
|-----------|------|----------|
| `sales_report` | 销售报表 | StatisticsSrv |
| `member_report` | 会员报表 | StatisticsSrv |
| `inventory_report` | 库存报表 | StatisticsSrv |
| `query_desk` | 查询桌台 | DeskSrv |
| `update_desk_status` | 更新桌台状态 | DeskSrv |

---

## 七、安全设计

### 7.1 认证授权

```go
// 复用现有的认证中间件
aiGroup := apiV1.Group("/ai")
aiGroup.Use(middleware.ShopAuth())  // 或其他认证中间件
{
    aiGroup.POST("/chat", agentApi.Chat)
}
```

### 7.2 权限控制

每个 Tool 可以定义所需权限：

```go
type Tool interface {
    // ... 其他方法

    // RequiredPermissions 返回执行此工具需要的权限
    RequiredPermissions() []string
}

// 示例
func (t *CancelOrderTool) RequiredPermissions() []string {
    return []string{"order:cancel"}
}
```

执行器检查权限：

```go
func (e *Executor) Execute(ctx context.Context, toolName string, input json.RawMessage) *ExecuteResult {
    tool, _ := e.registry.Get(toolName)

    // 检查权限
    permissions := tool.RequiredPermissions()
    if !hasPermissions(ctx, permissions) {
        return &ExecuteResult{
            Error: "权限不足",
            Success: false,
        }
    }

    // 执行工具
    // ...
}
```

### 7.3 数据隔离

- 从 context 中获取 company_uuid
- Tool 执行时只能访问当前商户的数据
- 使用现有的 `ctx.GetDB()` 获取对应商户的数据库连接

### 7.4 输入验证

- Tool 的 InputSchema 定义参数类型和约束
- Execute 方法中进行额外的业务验证
- 防止 SQL 注入、XSS 等攻击

### 7.5 日志审计

```go
func (s *AgentSrv) Chat(ctx context.Context, message string) (*resp.ChatResp, error) {
    // 记录请求日志
    logger.Info("AI chat request",
        zap.String("company_uuid", ctx.GetCompanyUuid()),
        zap.String("staff_uuid", ctx.GetStaffUuid()),
        zap.String("message", message),
    )

    // ... 执行逻辑

    // 记录工具调用日志
    for _, toolCall := range result.ToolCalls {
        logger.Info("AI tool call",
            zap.String("company_uuid", ctx.GetCompanyUuid()),
            zap.String("tool_name", toolCall.ToolName),
            zap.Any("input", toolCall.Input),
            zap.Bool("success", toolCall.Success),
        )
    }
}
```

### 7.6 限流保护

```go
// 使用现有的限流中间件或 Redis 实现
aiGroup.Use(middleware.RateLimit(RateLimitConfig{
    Key:      "ai_chat",
    Limit:    100,        // 每分钟最多 100 次
    Window:   time.Minute,
    ByUser:   true,       // 按用户限流
}))
```

---

## 八、配置管理

### 8.1 环境变量

```bash
# Claude API 配置
CLAUDE_API_KEY=sk-ant-xxxxx           # API Key（必填）
CLAUDE_BASE_URL=https://api.anthropic.com  # API 地址（可选，用于代理）
CLAUDE_MODEL=claude-sonnet-4-20250514          # 模型（可选）
CLAUDE_MAX_TOKENS=4096                 # 最大 Token（可选）
CLAUDE_TIMEOUT=120s                    # 超时时间（可选）

# Agent 配置
AGENT_MAX_TOOL_ITERATIONS=10          # 最大工具调用轮次
AGENT_ENABLE_CONVERSATION=true        # 是否启用会话管理
```

### 8.2 Nacos 配置（可选）

```yaml
ai:
  claude:
    api_key: ${CLAUDE_API_KEY}
    base_url: https://api.anthropic.com
    model: claude-sonnet-4-20250514
    max_tokens: 4096
    timeout: 120s
  agent:
    max_tool_iterations: 10
    system_prompt: |
      你是 TTPOS 餐饮系统的智能助手...
```

---

## 九、监控运维

### 9.1 监控指标

| 指标 | 类型 | 说明 |
|------|------|------|
| `ai_chat_requests_total` | Counter | 对话请求总数 |
| `ai_chat_duration_seconds` | Histogram | 对话耗时 |
| `ai_tool_calls_total` | Counter | 工具调用总数（按工具名分组） |
| `ai_tool_errors_total` | Counter | 工具调用失败数 |
| `ai_tokens_used_total` | Counter | Token 消耗总数 |

### 9.2 链路追踪

```go
func (s *AgentSrv) Chat(ctx context.Context, message string) (*resp.ChatResp, error) {
    ctx, span := tracer.Start(ctx, "AgentSrv.Chat")
    defer span.End()

    span.SetAttributes(
        attribute.String("message_length", strconv.Itoa(len(message))),
    )

    // ... 执行逻辑

    span.SetAttributes(
        attribute.Int("tool_calls_count", len(result.ToolCalls)),
        attribute.Int("tokens_used", result.Usage.TotalTokens),
    )
}
```

### 9.3 告警规则

- API 调用失败率 > 5%
- 平均响应时间 > 30s
- Token 消耗异常增长
- 工具执行失败率 > 10%

---

## 十、扩展性设计

### 10.1 添加新 Tool

1. 实现 Tool 接口
2. 在 register.go 中注册
3. 无需修改 Agent 核心代码

```go
// tools/my_tool.go
type MyTool struct {
    mySrv service.IMyService
}

func (t *MyTool) Name() string { return "my_tool" }
// ... 实现其他方法

// tools/register.go
func RegisterAllTools(registry *tool.Registry, ...) {
    registry.MustRegister(&MyTool{mySrv: mySrv})
}
```

### 10.2 切换 AI 模型

Claude Client 支持配置不同模型：

```go
agent.NewModule(apiKey,
    claude.WithModel("claude-opus-4-20250514"),  // 更强大的模型
)
```

### 10.3 支持其他 AI 提供商

定义通用接口，可扩展支持 OpenAI、国内大模型等：

```go
type LLMClient interface {
    CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error)
}

// claude.Client 实现此接口
// openai.Client 实现此接口
```

---

## 十一、实施计划

### Phase 1: MVP（1-2 周）

- [x] Claude API Client
- [x] Tool 框架（接口、注册表、执行器）
- [x] 基础 Demo Tools
- [x] Agent Service
- [x] HTTP API
- [ ] 集成到路由
- [ ] 基础测试

### Phase 2: 业务集成（2-3 周）

- [ ] 订单查询 Tool
- [ ] 会员查询 Tool
- [ ] 权限控制
- [ ] 日志审计
- [ ] 监控指标

### Phase 3: 增强功能（2-3 周）

- [ ] SSE 流式响应
- [ ] 会话管理（多轮对话）
- [ ] 更多业务 Tool
- [ ] 性能优化

### Phase 4: 生产就绪（1-2 周）

- [ ] 完善监控告警
- [ ] 文档完善
- [ ] 压力测试
- [ ] 灰度发布

---

## 十二、附录

### A. Claude API 参考

- [Messages API](https://docs.anthropic.com/claude/reference/messages_post)
- [Tool Use](https://docs.anthropic.com/claude/docs/tool-use)

### B. 示例对话

**用户：** 帮我查一下今天销售额超过100元的订单

**AI 思考过程：**
1. 理解用户需求：查询今天的订单，筛选条件是销售额 > 100
2. 调用 query_order 工具，传入时间范围和金额筛选条件
3. 获取结果后，用自然语言总结

**AI 回复：** 今天销售额超过100元的订单共有8笔，分别是...
