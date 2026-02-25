package agent

import (
	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/modules/agent/api"
	"ttpos-server-go/app/modules/agent/engine/tool"
	"ttpos-server-go/app/modules/agent/service"
	"ttpos-server-go/app/modules/agent/tools"
	"ttpos-server-go/pkg/claude"
)

// Module Agent 模块
type Module struct {
	client   *claude.Client
	registry *tool.Registry
	agentSrv *service.AgentSrv
	agentApi *api.AgentApi
}

// NewModule 创建 Agent 模块
// apiKey: Claude API Key
// opts: Claude Client 配置选项
func NewModule(apiKey string, opts ...claude.ClientOption) *Module {
	// 创建 Claude 客户端
	client := claude.NewClient(apiKey, opts...)

	// 创建工具注册表并注册工具
	registry := tool.NewRegistry()
	tools.RegisterDemoTools(registry)

	// 创建服务
	agentSrv := service.NewAgentSrv(client, registry)

	// 创建 API
	agentApi := api.NewAgentApi(agentSrv)

	return &Module{
		client:   client,
		registry: registry,
		agentSrv: agentSrv,
		agentApi: agentApi,
	}
}

// RegisterRoutes 注册路由
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	aiGroup := r.Group("/ai")
	{
		aiGroup.POST("/chat", m.agentApi.Chat)
		aiGroup.GET("/health", m.agentApi.Health)
	}
}

// GetRegistry 获取工具注册表（用于注册自定义工具）
func (m *Module) GetRegistry() *tool.Registry {
	return m.registry
}

// GetAgentSrv 获取 Agent 服务（用于直接调用）
func (m *Module) GetAgentSrv() service.IAgentSrv {
	return m.agentSrv
}
