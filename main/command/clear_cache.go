package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	objectStoragePersistence "ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	clearCacheAllFlag         bool
	clearCacheCompanyUuidFlag uint64
	clearCacheObjectTypeFlag  string
	clearCacheKeyFlag         string
	clearCacheForceFlag       bool
	clearCacheL1OnlyFlag      bool
)

func init() {
	rootCommand.AddCommand(clearCacheCmd)
	clearCacheCmd.Flags().BoolVar(&clearCacheAllFlag, "all", false, "清空所有系统缓存 (ttpos:*)")
	clearCacheCmd.Flags().Uint64Var(&clearCacheCompanyUuidFlag, "company-uuid", 0, "指定商家 UUID，清空该商家的所有缓存 (ttpos:{company_uuid}:*)")
	clearCacheCmd.Flags().StringVar(&clearCacheObjectTypeFlag, "object-type", "", "指定对象类型（需配合 --company-uuid），清空该类型缓存 (ttpos:{company_uuid}:{object_type}:*)")
	clearCacheCmd.Flags().StringVar(&clearCacheKeyFlag, "key", "", "指定单个 key，清空该 key 的缓存")
	clearCacheCmd.Flags().BoolVar(&clearCacheForceFlag, "force", false, "跳过确认提示（谨慎使用）")
	clearCacheCmd.Flags().BoolVar(&clearCacheL1OnlyFlag, "l1-only", false, "仅清除 L1 本地缓存（内存缓存），不清除 L2 Redis 缓存")
}

var clearCacheCmd = &cobra.Command{
	Use:   "clear-cache",
	Short: "清理 Redis 中的对象存储缓存",
	Long: `清理 Redis 中的对象存储缓存，支持以下清理模式：
1. 清空所有系统缓存：--all
2. 清空某个商家的缓存：--company-uuid {uuid}
3. 清除某类型数据的缓存：--company-uuid {uuid} --object-type {type}
4. 清除指定 key 的缓存：--key {key}
5. 仅清除 L1 本地缓存（内存缓存）：--l1-only`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			fmt.Printf("%sFailed to initialize config: %v%s\n", redColor, err, resetColor)
			os.Exit(1)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			fmt.Printf("%sFailed to initialize logger: %v%s\n", redColor, err, resetColor)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 如果指定了 --l1-only，通过 HTTP 接口清除所有 L1 缓存
		if clearCacheL1OnlyFlag {
			fmt.Printf("%s模式: 仅清除 L1 本地缓存（内存缓存）%s\n", blueColor, resetColor)
			fmt.Printf("%s注意: 此操作将通过 HTTP 接口清空服务进程的所有 cacheGroup 单例的 L1 缓存%s\n", yellowColor, resetColor)
			fmt.Printf("%sL2 Redis 缓存不受影响，L1 缓存会在下次访问时自动从 L2 回填%s\n", yellowColor, resetColor)

			// 确认提示（除非使用 --force）
			if !clearCacheForceFlag {
				fmt.Printf("%s========================================%s\n", blueColor, resetColor)
				fmt.Printf("%s警告：此操作将清空服务进程的所有 L1 本地缓存%s\n", redColor, resetColor)
				fmt.Printf("%s输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
				var confirmation string
				fmt.Scanln(&confirmation)
				if strings.ToLower(strings.TrimSpace(confirmation)) != "yes" {
					fmt.Printf("%s操作已取消%s\n", yellowColor, resetColor)
					return
				}
			}

			// 通过 HTTP 接口清除 L1 缓存
			fmt.Printf("%s开始通过 HTTP 接口清除 L1 缓存...%s\n", blueColor, resetColor)
			if err := clearL1CacheViaHTTP(); err != nil {
				fmt.Printf("%s错误: 清除 L1 缓存失败: %v%s\n", redColor, err, resetColor)
				logger.Logger.Error("清除 L1 缓存失败", zap.Error(err))
				return
			}

			// 显示结果
			fmt.Printf("%s========================================%s\n", greenColor, resetColor)
			fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
			fmt.Printf("%s已通过 HTTP 接口清空服务进程的 L1 缓存%s\n", greenColor, resetColor)
			fmt.Printf("%s========================================%s\n", greenColor, resetColor)
			return
		}

		// 验证参数合法性
		modeCount := 0
		if clearCacheAllFlag {
			modeCount++
		}
		if clearCacheCompanyUuidFlag > 0 {
			modeCount++
		}
		if clearCacheKeyFlag != "" {
			modeCount++
		}

		if modeCount == 0 {
			fmt.Printf("%s错误: 必须指定一种清理模式（--all, --company-uuid, 或 --key）%s\n", redColor, resetColor)
			cmd.Help()
			return
		}

		if modeCount > 1 {
			fmt.Printf("%s错误: 不能同时使用多个清理模式%s\n", redColor, resetColor)
			cmd.Help()
			return
		}

		if clearCacheObjectTypeFlag != "" && clearCacheCompanyUuidFlag == 0 {
			fmt.Printf("%s错误: --object-type 必须配合 --company-uuid 使用%s\n", redColor, resetColor)
			cmd.Help()
			return
		}

		// 获取 Redis 客户端
		var client redis.UniversalClient
		if clusterClient := cache.Global.GetClusterClient(); clusterClient != nil {
			client = clusterClient
		} else if redisClient := cache.Global.GetClient(); redisClient != nil {
			client = redisClient
		} else {
			fmt.Printf("%s错误: 无法获取 Redis 客户端，当前缓存类型不支持此操作%s\n", redColor, resetColor)
			return
		}

		// 构建匹配模式和 keys
		var pattern string
		var keys []string
		var err error
		ctx := context.Background()

		if clearCacheAllFlag {
			// 清空所有系统缓存
			pattern = fmt.Sprintf("%s:*", objectStoragePersistence.SystemPrefix)
			fmt.Printf("%s模式: 清空所有系统缓存%s\n", blueColor, resetColor)
			fmt.Printf("%s匹配模式: %s%s\n", blueColor, pattern, resetColor)
			keys, err = cache.ScanRedisKeysDefault(ctx, client, pattern)
		} else if clearCacheKeyFlag != "" {
			// 清除指定 key
			keys = []string{clearCacheKeyFlag}
			fmt.Printf("%s模式: 清除指定 key%s\n", blueColor, resetColor)
			fmt.Printf("%sKey: %s%s\n", blueColor, clearCacheKeyFlag, resetColor)
		} else if clearCacheCompanyUuidFlag > 0 {
			if clearCacheObjectTypeFlag != "" {
				// 清除指定商家的指定类型缓存
				pattern = fmt.Sprintf("%s:%d:%s:*", objectStoragePersistence.SystemPrefix, clearCacheCompanyUuidFlag, clearCacheObjectTypeFlag)
				fmt.Printf("%s模式: 清除指定商家的指定类型缓存%s\n", blueColor, resetColor)
				fmt.Printf("%s商家 UUID: %d%s\n", blueColor, clearCacheCompanyUuidFlag, resetColor)
				fmt.Printf("%s对象类型: %s%s\n", blueColor, clearCacheObjectTypeFlag, resetColor)
				fmt.Printf("%s匹配模式: %s%s\n", blueColor, pattern, resetColor)
			} else {
				// 清空指定商家的所有缓存
				pattern = fmt.Sprintf("%s:%d:*", objectStoragePersistence.SystemPrefix, clearCacheCompanyUuidFlag)
				fmt.Printf("%s模式: 清空指定商家的所有缓存%s\n", blueColor, resetColor)
				fmt.Printf("%s商家 UUID: %d%s\n", blueColor, clearCacheCompanyUuidFlag, resetColor)
				fmt.Printf("%s匹配模式: %s%s\n", blueColor, pattern, resetColor)
			}
			keys, err = cache.ScanRedisKeysDefault(ctx, client, pattern)
		}

		if err != nil {
			fmt.Printf("%s错误: 扫描 Redis keys 失败: %v%s\n", redColor, err, resetColor)
			logger.Logger.Error("扫描 Redis keys 失败", zap.Error(err))
			return
		}

		if len(keys) == 0 {
			fmt.Printf("%s提示: 没有找到匹配的缓存 key%s\n", yellowColor, resetColor)
			return
		}

		// 显示预览信息
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)
		fmt.Printf("%s找到 %d 个匹配的缓存 key%s\n", blueColor, len(keys), resetColor)
		fmt.Printf("%s========================================%s\n", blueColor, resetColor)

		// 显示前10个 keys 预览
		previewCount := 10
		if len(keys) < previewCount {
			previewCount = len(keys)
		}
		fmt.Printf("%s前 %d 个 keys 预览：%s\n", yellowColor, previewCount, resetColor)
		for i := 0; i < previewCount; i++ {
			fmt.Printf("  %d. %s\n", i+1, keys[i])
		}
		if len(keys) > previewCount {
			fmt.Printf("%s  ... 还有 %d 个 keys%s\n", yellowColor, len(keys)-previewCount, resetColor)
		}

		// 确认提示（除非使用 --force）
		if !clearCacheForceFlag {
			fmt.Printf("%s========================================%s\n", blueColor, resetColor)
			fmt.Printf("%s警告：此操作将删除上述 %d 个缓存 key%s\n", redColor, len(keys), resetColor)
			fmt.Printf("%s输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
			var confirmation string
			fmt.Scanln(&confirmation)
			if strings.ToLower(strings.TrimSpace(confirmation)) != "yes" {
				fmt.Printf("%s操作已取消%s\n", yellowColor, resetColor)
				return
			}
		}

		// 执行删除
		fmt.Printf("%s开始删除缓存...%s\n", blueColor, resetColor)

		// 批量删除，每次删除一批（避免一次性删除过多）
		batchSize := 1000
		deletedCount := 0
		for i := 0; i < len(keys); i += batchSize {
			end := i + batchSize
			if end > len(keys) {
				end = len(keys)
			}
			batch := keys[i:end]
			cache.Global.Del(batch...)
			deletedCount += len(batch)
			fmt.Printf("%s已删除 %d/%d 个 keys...%s\r", blueColor, deletedCount, len(keys), resetColor)
		}
		fmt.Printf("\n")

		// 显示删除结果
		fmt.Printf("%s========================================%s\n", greenColor, resetColor)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s删除的 key 数量: %d%s\n", greenColor, deletedCount, resetColor)
		fmt.Printf("%s========================================%s\n", greenColor, resetColor)
	},
}

// clearL1CacheViaHTTP 通过 HTTP 接口清除 L1 缓存
func clearL1CacheViaHTTP() error {
	// 构建请求 URL
	serverPort := config.Server.Port
	if serverPort == "" {
		serverPort = "8080" // 默认端口
	}
	url := fmt.Sprintf("http://localhost:%s/api/v1/internal/cache/l1/clear", serverPort)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置 API Key header，使用 CommitSHA 作为身份验证
	apiKey := config.CommitSHA
	if apiKey == "" {
		return fmt.Errorf("CommitSHA 未配置，无法调用内部接口")
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应中的 code
	if code, ok := result["code"].(float64); ok && code != 0 {
		msg := "未知错误"
		if message, ok := result["message"].(string); ok {
			msg = message
		}
		return fmt.Errorf("接口返回错误: %s", msg)
	}

	// 显示清除结果
	if data, ok := result["data"].(map[string]interface{}); ok {
		if clearedCount, ok := data["cleared_count"].(float64); ok {
			fmt.Printf("%s清除的 cacheGroup 数量: %.0f%s\n", blueColor, clearedCount, resetColor)
		}
		if message, ok := data["message"].(string); ok {
			fmt.Printf("%s%s%s\n", blueColor, message, resetColor)
		}
	}

	return nil
}
