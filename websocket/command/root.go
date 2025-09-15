package command

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"websocket/api"
	"websocket/config"
	"websocket/pkg/cache"
	"websocket/pkg/database"
	"websocket/pkg/logger"
	"websocket/service"
	"websocket/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

// internalAuthMiddleware 内部访问验证中间件
func internalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查X-API-KEY头部
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" {
			http.Error(w, "404 Unauthorized", http.StatusUnauthorized)
			return
		}

		// 验证API Key是否匹配配置中的JWT Secret
		if apiKey != config.JWT.Secret {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// 验证通过，继续处理请求
		next(w, r)
	}
}

var rootCommand = &cobra.Command{
	Use:   "ttpos",
	Short: "启动服务",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		// 初始化数据库管理器
		database.GetDBManager(config.Database)

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cacheConfig)

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		utils.InitIdGenerator()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// redis 订阅
		go service.RedisSubscribe()
		// 启动WebSocket服务器
		http.HandleFunc("/ws", service.HandleConnections)
		// 开放推送接口（仅限内部访问）
		http.HandleFunc("/ws/push", internalAuthMiddleware(api.PushClient))
		// 启动HTTP服务器
		fmt.Println("WebSocket server started on :8099")
		err := http.ListenAndServe(":8099", nil)
		if err != nil {
			fmt.Println("Error starting server:", err)
		}

	},
}

func Execute() {
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
