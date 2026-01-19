package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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
			patterns := []string{
				fmt.Sprintf("%s:*", objectStoragePersistence.SystemPrefix), // ttpos5:*
			}

			fmt.Printf("%s模式: 清空所有系统缓存%s\n", blueColor, resetColor)
			fmt.Printf("%s扫描模式: %s%s\n", blueColor, strings.Join(patterns, ", "), resetColor)

			// 合并所有模式的 keys
			allKeys := make(map[string]bool) // 使用 map 去重
			var scanErr error
			for _, pattern := range patterns {
				fmt.Printf("%s正在扫描模式: %s...%s\n", blueColor, pattern, resetColor)
				patternKeys, patternErr := cache.ScanRedisKeys(ctx, client, pattern, 5000) // 使用更大的 count 值
				if patternErr != nil {
					fmt.Printf("%s警告: 扫描模式 %s 失败: %v%s\n", yellowColor, pattern, patternErr, resetColor)
					logger.Logger.Warn("扫描 Redis keys 失败", zap.String("pattern", pattern), zap.Error(patternErr))
					scanErr = patternErr // 记录错误，但不中断扫描其他模式
					continue
				}
				for _, key := range patternKeys {
					allKeys[key] = true
				}
				fmt.Printf("%s  找到 %d 个 keys%s\n", blueColor, len(patternKeys), resetColor)
			}

			// 转换为切片
			keys = make([]string, 0, len(allKeys))
			for key := range allKeys {
				keys = append(keys, key)
			}

			// 如果所有模式都扫描失败，设置 err
			if len(keys) == 0 && scanErr != nil {
				err = scanErr
			}
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

		// 显示所有 keys（按字母顺序排序，便于查看）
		sortedKeys := make([]string, len(keys))
		copy(sortedKeys, keys)
		sort.Strings(sortedKeys)

		fmt.Printf("%s所有 keys 列表：%s\n", yellowColor, resetColor)
		for i := 0; i < len(sortedKeys); i++ {
			fmt.Printf("  %d. %s\n", i+1, sortedKeys[i])
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

		// 验证 Global 是否已初始化
		if cache.Global == nil {
			fmt.Printf("%s错误: cache.Global 未初始化%s\n", redColor, resetColor)
			logger.Logger.Error("cache.Global 未初始化")
			return
		}

		// 验证 client 是否已获取（在之前已经获取过了）
		if client == nil {
			fmt.Printf("%s错误: Redis 客户端未初始化%s\n", redColor, resetColor)
			logger.Logger.Error("Redis 客户端未初始化")
			return
		}

		// 判断是否为集群模式
		isCluster := false
		if _, ok := client.(*redis.ClusterClient); ok {
			isCluster = true
		}

		// 批量删除，每次删除一批（避免一次性删除过多）
		// 注意：Redis 集群模式下，一次 DEL 只能删除同一个 slot 的 keys
		// 如果是集群模式，需要逐个删除或按 slot 分组删除
		batchSize := 1000
		if isCluster {
			// 集群模式下，逐个删除以避免 CROSSSLOT 错误
			batchSize = 1
			fmt.Printf("%s检测到 Redis 集群模式，将逐个删除 keys（避免 CROSSSLOT 错误）%s\n", blueColor, resetColor)
		}

		deletedCount := 0
		failedCount := 0
		for i := 0; i < len(keys); i += batchSize {
			end := i + batchSize
			if end > len(keys) {
				end = len(keys)
			}
			batch := keys[i:end]

			// 使用 Redis 客户端直接删除，并检查结果
			delResult := client.Del(ctx, batch...)
			if err := delResult.Err(); err != nil {
				// 如果是集群模式且是 CROSSSLOT 错误，尝试逐个删除
				if isCluster && strings.Contains(err.Error(), "CROSSSLOT") {
					// 逐个删除这个批次中的 keys
					for _, key := range batch {
						singleDelResult := client.Del(ctx, key)
						if singleErr := singleDelResult.Err(); singleErr != nil {
							fmt.Printf("%s警告: 删除 key 失败: %s, 错误: %v%s\n", yellowColor, key, singleErr, resetColor)
							logger.Logger.Warn("删除单个 key 失败", zap.String("key", key), zap.Error(singleErr))
							failedCount++
						} else {
							deletedCount++
						}
					}
				} else {
					fmt.Printf("%s警告: 删除批次失败: %v%s\n", yellowColor, err, resetColor)
					logger.Logger.Warn("删除缓存批次失败", zap.Error(err), zap.Int("batch_start", i), zap.Int("batch_end", end), zap.Strings("keys", batch))
					failedCount += len(batch)
				}
			} else {
				// 检查实际删除的数量
				actualDeleted := delResult.Val()
				deletedCount += int(actualDeleted)
				if int(actualDeleted) != len(batch) {
					fmt.Printf("%s警告: 批次删除数量不匹配，期望 %d，实际 %d%s\n", yellowColor, len(batch), actualDeleted, resetColor)
					logger.Logger.Warn("删除数量不匹配", zap.Int("expected", len(batch)), zap.Int64("actual", actualDeleted))
					failedCount += len(batch) - int(actualDeleted)
				}
			}

			// 每删除 100 个 keys 显示一次进度（避免输出过多）
			if deletedCount%100 == 0 || deletedCount == len(keys) {
				fmt.Printf("%s已删除 %d/%d 个 keys...%s\r", blueColor, deletedCount, len(keys), resetColor)
			}
		}
		fmt.Printf("\n")

		if failedCount > 0 {
			fmt.Printf("%s警告: %d 个 keys 删除失败%s\n", yellowColor, failedCount, resetColor)
		}

		// 同时清除 L1 本地缓存（通过 HTTP 接口）
		fmt.Printf("%s正在清除 L1 本地缓存...%s\n", blueColor, resetColor)
		l1ClearedCount := 0
		if err := clearL1CacheViaHTTP(); err != nil {
			fmt.Printf("%s警告: 清除 L1 缓存失败: %v%s\n", yellowColor, err, resetColor)
			logger.Logger.Warn("清除 L1 缓存失败", zap.Error(err))
		} else {
			l1ClearedCount = 1 // HTTP 接口会返回清除的 cacheGroup 数量
			fmt.Printf("%sL1 缓存清除成功%s\n", greenColor, resetColor)
		}

		// 显示删除结果
		fmt.Printf("%s========================================%s\n", greenColor, resetColor)
		fmt.Printf("%s操作成功完成！%s\n", greenColor, resetColor)
		fmt.Printf("%s删除的 L2 Redis key 数量: %d%s\n", greenColor, deletedCount, resetColor)
		if l1ClearedCount > 0 {
			fmt.Printf("%s清除的 L1 缓存组数量: %d%s\n", greenColor, l1ClearedCount, resetColor)
		}
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
