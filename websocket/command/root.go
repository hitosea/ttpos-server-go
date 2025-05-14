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

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

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

	},
	Run: func(cmd *cobra.Command, args []string) {
		// redis 订阅
		go service.RedisSubscribe()
		// 启动WebSocket服务器
		http.HandleFunc("/ws", service.HandleConnections)
		// 开放推送接口
		http.HandleFunc("/ws/push", api.PushClient)
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
