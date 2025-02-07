package command

import (
	"context"
	"fmt"
	"log"
	"websocket/config"
	"websocket/pkg/cache"
	"websocket/pkg/database"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(testsCmd)
}

var testsCmd = &cobra.Command{
	Use:   "test",
	Short: "开始测试",
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

	},
	Run: func(_ *cobra.Command, _ []string) {
		// 订阅
		err := cache.GlobalRedis.Client.Publish(context.Background(), "websocket_msg_push", "您的消息内容").Err()
		if err != nil {
			fmt.Println("Error publishing message:", err)
		} else {
			fmt.Println("Message published successfully")
		}
	},
}
