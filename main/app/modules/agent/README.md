# AI Agent 模块

基于 Claude API 的 AI Agent 模块，支持自然语言理解和工具调用。

## 快速开始

### 1. 设置环境变量

```bash
# 在 .env 文件中添加
CLAUDE_API_KEY=your_api_key_here

# 可选配置
CLAUDE_BASE_URL=https://api.anthropic.com  # 自定义 API 地址（如使用代理）
CLAUDE_MODEL=claude-sonnet-4-20250514           # 使用的模型
```

### 2. 在 router 中注册

```go
// router/router.go

import (
    "ttpos-server-go/app/modules/agent"
    "ttpos-server-go/pkg/claude"
)

func Setup(r *gin.Engine, dbm *database.DBManager, cache cache.Cache) {
    // ... 其他路由 ...

    apiV1 := r.Group("api/v1")
    {
        // ... 其他路由组 ...

        // AI Agent 模块
        apiKey := os.Getenv("CLAUDE_API_KEY")
        if apiKey != "" {
            agentModule := agent.NewModule(apiKey,
                // 可选：自定义配置
                // claude.WithBaseURL("https://your-proxy.com"),
                // claude.WithModel("claude-sonnet-4-20250514"),
            )
            agentModule.RegisterRoutes(apiV1)
        }
    }
}
```

### 3. 测试接口

```bash
# 健康检查
curl http://localhost:8080/api/v1/ai/health

# 对话测试
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "现在几点了？"}'

# 查询订单
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "帮我查一下今天的订单"}'

# 计算
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "计算 123 乘以 456"}'
```

## API 接口

### POST /api/v1/ai/chat

对话接口，AI 会根据需要自动调用工具。

**请求体：**
```json
{
  "message": "用户消息"
}
```

**响应：**
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "content": "AI 回复内容",
    "tool_calls": [
      {
        "tool_name": "query_order",
        "input": {"status": "completed"},
        "output": {...},
        "success": true
      }
    ],
    "finish_reason": "end_turn",
    "usage": {
      "input_tokens": 100,
      "output_tokens": 200,
      "total_tokens": 300
    }
  }
}
```

## 添加自定义工具

### 1. 实现 Tool 接口

```go
// app/modules/agent/tools/my_tool.go

package tools

import (
    "context"
    "ttpos-server-go/app/modules/agent/engine/tool"
)

type MyCustomTool struct {
    // 可以注入服务依赖
    // mySrv service.IMyService
}

func (t *MyCustomTool) Name() string {
    return "my_custom_tool"
}

func (t *MyCustomTool) Description() string {
    return "工具描述，帮助 AI 理解何时使用此工具"
}

func (t *MyCustomTool) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param1": map[string]any{
                "type":        "string",
                "description": "参数1描述",
            },
        },
        "required": []string{"param1"},
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, input map[string]any) (any, error) {
    param1 := input["param1"].(string)
    // 执行逻辑...
    return map[string]any{"result": "success"}, nil
}
```

### 2. 注册工具

```go
// 方式1：在 tools/demo_tools.go 的 RegisterDemoTools 中添加
registry.MustRegister(&MyCustomTool{})

// 方式2：在路由注册时动态添加
agentModule := agent.NewModule(apiKey)
agentModule.GetRegistry().MustRegister(&MyCustomTool{})
agentModule.RegisterRoutes(apiV1)
```

## 目录结构

```
app/modules/agent/
├── api/                    # API 层
│   └── agent_api.go
├── dto/                    # 数据传输对象
│   ├── req/
│   │   └── agent_req.go
│   └── resp/
│       └── agent_resp.go
├── engine/                 # Agent 引擎
│   └── tool/               # Tool 框架
│       ├── interface.go    # Tool 接口定义
│       ├── registry.go     # 工具注册表
│       └── executor.go     # 工具执行器
├── service/                # 服务层
│   └── agent_service.go    # Agent 核心服务
├── tools/                  # 工具实现
│   └── demo_tools.go       # 演示工具
├── agent.go                # 模块入口
└── README.md
```

## 架构说明

```
用户消息 → AgentApi → AgentSrv → Claude API
                         ↓
                    [如果需要调用工具]
                         ↓
              Claude 返回 tool_use
                         ↓
              Executor 执行工具
                         ↓
              工具结果返回 Claude
                         ↓
              Claude 生成最终回复
                         ↓
                    返回给用户
```

## 注意事项

1. **API Key 安全**：不要将 API Key 提交到代码仓库
2. **Token 消耗**：每次对话都会消耗 Token，注意监控用量
3. **超时设置**：默认 120 秒超时，复杂操作可能需要更长时间
4. **错误处理**：工具执行失败会返回错误信息给 AI，AI 会尝试其他方式
5. **最大轮次**：默认最多 10 轮工具调用，防止无限循环
