package service

import (
	"context"
	"encoding/json"
	"fmt"
	"ttpos-ai/config"
	"ttpos-ai/pkg/database"
	"ttpos-ai/pkg/logger"
	"ttpos-ai/pkg/tools"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// IAgentService Agent服务接口
type IAgentService interface {
	Query(ctx context.Context, question string) (*AgentResult, error)
	QueryWithHistory(ctx context.Context, messages []ChatMessage) (*AgentResult, error)
}

// AgentService Agent服务实现
type AgentService struct {
	chatModel model.ChatModel
	sqlTool   *tools.SQLTool
	schema    string // 数据库schema缓存
}

// AgentResult Agent查询结果
type AgentResult struct {
	Answer   string `json:"answer"`            // 最终回答
	SQLUsed  string `json:"sql_used,omitempty"` // 使用的SQL
	SQLData  any    `json:"sql_data,omitempty"` // SQL查询结果
	Thinking string `json:"thinking,omitempty"` // 思考过程
}

// NewAgentService 创建Agent服务
func NewAgentService() (IAgentService, error) {
	if !config.LLM.Enabled {
		return nil, fmt.Errorf("LLM功能未启用")
	}

	// 检查数据库连接
	if !database.IsConnected() {
		return nil, fmt.Errorf("数据库未连接")
	}

	cfg := &config.LLM
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxTokens:   &cfg.MaxTokens,
		Temperature: &cfg.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建ChatModel失败: %w", err)
	}

	sqlTool := tools.NewSQLTool()

	// 获取数据库schema
	schemaInfo, err := sqlTool.GetSchemaInfo(context.Background())
	if err != nil {
		logger.Logger.Warn("获取数据库Schema失败", zap.Error(err))
		schemaInfo = "无法获取数据库结构信息"
	}

	return &AgentService{
		chatModel: chatModel,
		sqlTool:   sqlTool,
		schema:    schemaInfo,
	}, nil
}

// Query 查询（单轮）
func (s *AgentService) Query(ctx context.Context, question string) (*AgentResult, error) {
	return s.QueryWithHistory(ctx, []ChatMessage{
		{Role: "user", Content: question},
	})
}

// QueryWithHistory 带历史的查询
func (s *AgentService) QueryWithHistory(ctx context.Context, messages []ChatMessage) (*AgentResult, error) {
	logger.Logger.Info("Agent查询", zap.Int("message_count", len(messages)))

	// 构建系统提示
	systemPrompt := s.buildSystemPrompt()

	// 构建消息
	schemaMessages := make([]*schema.Message, 0, len(messages)+1)
	schemaMessages = append(schemaMessages, &schema.Message{
		Role:    schema.System,
		Content: systemPrompt,
	})

	for _, msg := range messages {
		var role schema.RoleType
		switch msg.Role {
		case "system":
			continue // 跳过用户的system消息
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		default:
			role = schema.User
		}
		schemaMessages = append(schemaMessages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 定义工具
	toolInfo := &schema.ToolInfo{
		Name: "execute_sql",
		Desc: "执行SQL查询语句，用于查询数据库中的数据。只支持SELECT语句。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"sql": {
				Type: schema.String,
				Desc: "要执行的SQL SELECT语句，表名需要加上前缀 ttpos_",
			},
		}),
	}

	// 绑定工具
	if err := s.chatModel.BindTools([]*schema.ToolInfo{toolInfo}); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %w", err)
	}

	result := &AgentResult{}
	maxIterations := 5 // 最大迭代次数，防止无限循环

	for i := 0; i < maxIterations; i++ {
		// 调用LLM
		resp, err := s.chatModel.Generate(ctx, schemaMessages)
		if err != nil {
			return nil, fmt.Errorf("LLM调用失败: %w", err)
		}

		// 检查是否有工具调用
		if len(resp.ToolCalls) > 0 {
			// 添加助手消息
			schemaMessages = append(schemaMessages, resp)

			// 处理工具调用
			for _, toolCall := range resp.ToolCalls {
				if toolCall.Function.Name == "execute_sql" {
					// 解析SQL参数
					var args struct {
						SQL string `json:"sql"`
					}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						logger.Logger.Error("解析工具参数失败", zap.Error(err))
						continue
					}

					result.SQLUsed = args.SQL
					logger.Logger.Info("执行SQL", zap.String("sql", args.SQL))

					// 执行SQL
					sqlResult, err := s.sqlTool.Execute(ctx, &tools.SQLQueryInput{SQL: args.SQL})
					if err != nil {
						logger.Logger.Error("SQL执行错误", zap.Error(err))
					}

					// 构建工具响应
					toolResultJSON, _ := json.Marshal(sqlResult)
					result.SQLData = sqlResult

					// 添加工具结果消息
					schemaMessages = append(schemaMessages, schema.ToolMessage(
						string(toolResultJSON),
						toolCall.ID,
					))
				}
			}
		} else {
			// 没有工具调用，返回最终结果
			result.Answer = resp.Content
			break
		}
	}

	// 如果没有最终答案，说明达到了最大迭代次数
	if result.Answer == "" {
		result.Answer = "抱歉，无法完成查询，请尝试简化您的问题。"
	}

	logger.Logger.Info("Agent查询完成", zap.String("sql", result.SQLUsed))
	return result, nil
}

// buildSystemPrompt 构建系统提示
func (s *AgentService) buildSystemPrompt() string {
	return fmt.Sprintf(`你是一个专业的数据库查询助手，可以帮助用户查询餐饮系统的数据。

## 你的能力
- 理解用户的自然语言问题
- 将问题转换为SQL查询
- 解释查询结果

## 数据库信息
%s

## 规则
1. 只能使用SELECT语句查询数据
2. 表名都有前缀 ttpos_，查询时必须带上
3. 查询结果最多返回%d条记录
4. 如果用户问题不需要查询数据库，直接回答即可
5. 查询后要用通俗易懂的语言解释结果
6. 金额字段通常以分为单位，展示时需要转换为元
7. 时间字段通常是Unix时间戳，展示时需要转换为可读格式

## 常用表说明
- ttpos_order: 订单表
- ttpos_order_product: 订单商品表
- ttpos_product_package: 商品表
- ttpos_product_category: 商品分类表
- ttpos_member: 会员表
- ttpos_desk: 桌位表

请根据用户的问题，决定是否需要查询数据库，并给出准确的回答。`, s.schema, config.Database.MaxRows)
}
