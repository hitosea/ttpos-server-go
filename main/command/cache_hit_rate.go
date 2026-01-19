package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

var (
	cacheHitRateTopNFlag      int
	cacheHitRateSortByFlag    string
	cacheHitRateKeyFlag       string
	cacheHitRateWatchFlag     bool
	cacheHitRateWatchInterval int
)

func init() {
	rootCommand.AddCommand(cacheHitRateCmd)
	cacheHitRateCmd.Flags().IntVar(&cacheHitRateTopNFlag, "top", 10, "显示 Top N 的 key（默认 10）")
	cacheHitRateCmd.Flags().StringVar(&cacheHitRateSortByFlag, "sort", "total", "排序方式：total（总访问次数）、hits（命中次数）、hitrate（命中率），默认 total")
	cacheHitRateCmd.Flags().StringVar(&cacheHitRateKeyFlag, "key", "", "查询指定 key 的统计信息")
	cacheHitRateCmd.Flags().BoolVar(&cacheHitRateWatchFlag, "watch", false, "实时监控模式（持续更新）")
	cacheHitRateCmd.Flags().IntVar(&cacheHitRateWatchInterval, "interval", 5, "监控模式下的更新间隔（秒），默认 5 秒")
}

var cacheHitRateCmd = &cobra.Command{
	Use:   "cache-hit-rate",
	Short: "查询对象存储缓存的命中率统计",
	Long: `查询对象存储缓存的命中率统计信息，支持以下功能：
1. 显示整体统计信息（命中次数、未命中次数、命中率、总请求数）
2. 显示 Top N 的 key 统计（--top N）
3. 查询指定 key 的统计信息（--key {key}）
4. 重置统计信息（--reset）
5. 实时监控模式（--watch）`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s\n", redColor, err, resetColor)
			os.Exit(1)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("%sFailed to initialize logger: %v%s\n", redColor, err, resetColor)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 如果指定了 key，显示该 key 的统计信息
		if cacheHitRateKeyFlag != "" {
			if err := displayKeyStatsViaHTTP(cacheHitRateKeyFlag); err != nil {
				fmt.Printf("%s错误: 查询 key 统计信息失败: %v%s\n", redColor, err, resetColor)
				return
			}
			return
		}

		// 如果指定了 watch，进入实时监控模式
		if cacheHitRateWatchFlag {
			watchModeViaHTTP()
			return
		}

		// 默认显示整体统计和 Top N
		if err := displayStatsViaHTTP(); err != nil {
			fmt.Printf("%s错误: 查询统计信息失败: %v%s\n", redColor, err, resetColor)
			return
		}
	},
}

// displayStatsViaHTTP 通过 HTTP 接口显示统计信息
func displayStatsViaHTTP() error {
	// 构建请求 URL
	serverPort := config.Server.Port
	if serverPort == "" {
		serverPort = "8080" // 默认端口
	}
	u, err := url.Parse(fmt.Sprintf("http://localhost:%s/api/v1/internal/cache/hit-rate/stats", serverPort))
	if err != nil {
		return fmt.Errorf("构建请求 URL 失败: %w", err)
	}

	// 添加查询参数
	q := u.Query()
	q.Set("top", strconv.Itoa(cacheHitRateTopNFlag))
	q.Set("sort", cacheHitRateSortByFlag)
	u.RawQuery = q.Encode()

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", u.String(), nil)
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
	httpResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 请求失败，状态码: %d, 响应: %s", httpResp.StatusCode, string(body))
	}

	// 解析响应
	var apiResponse dto.Response
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应中的 code
	if apiResponse.Code != 0 {
		msg := apiResponse.Message
		if msg == "" {
			msg = "未知错误"
		}
		return fmt.Errorf("接口返回错误: %s", msg)
	}

	// 解析数据
	var statsResp resp.CacheHitRateStatsResp
	statsDataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return fmt.Errorf("序列化响应数据失败: %w", err)
	}
	if err := json.Unmarshal(statsDataBytes, &statsResp); err != nil {
		return fmt.Errorf("解析统计数据失败: %w", err)
	}

	// 显示统计信息
	separator := strings.Repeat("=", 80)
	fmt.Println(separator)
	fmt.Println("                   对象存储缓存命中率统计")
	fmt.Printf("更新时间: %s\n", statsResp.Stats.LastUpdate.Format("2006-01-02 15:04:05"))
	fmt.Println(separator)

	// 整体统计
	fmt.Println("\n📊 整体统计:")
	fmt.Printf("  命中次数: %s%d%s\n", greenColor, statsResp.Stats.Hits, resetColor)
	fmt.Printf("  未命中次数: %s%d%s\n", yellowColor, statsResp.Stats.Misses, resetColor)
	fmt.Printf("  总请求数: %s%d%s\n", blueColor, statsResp.Stats.Total, resetColor)
	if statsResp.Stats.Total > 0 {
		hitRateColor := greenColor
		if statsResp.Stats.HitRate < 70 {
			hitRateColor = redColor
		} else if statsResp.Stats.HitRate < 80 {
			hitRateColor = yellowColor
		}
		fmt.Printf("  命中率: %s%.2f%%%s\n", hitRateColor, statsResp.Stats.HitRate, resetColor)
	} else {
		fmt.Printf("  命中率: %sN/A%s (暂无数据)\n", yellowColor, resetColor)
	}

	// Key 级别统计
	if statsResp.Stats.KeyCount > 0 {
		fmt.Printf("\n📈 Key 级别统计 (共 %d 个 key):\n", statsResp.Stats.KeyCount)

		// 显示 Top Keys
		if len(statsResp.TopKeys) > 0 {
			sortByLabel := map[string]string{
				"total":   "总访问次数",
				"hits":    "命中次数",
				"hitrate": "命中率",
			}
			sortBy := cacheHitRateSortByFlag
			if sortBy != "total" && sortBy != "hits" && sortBy != "hitrate" {
				sortBy = "total"
			}
			fmt.Printf("\n🔥 Top %d 最热 Key (按%s排序):\n", len(statsResp.TopKeys), sortByLabel[sortBy])
			fmt.Printf("%-50s %8s %8s %8s %10s %20s\n", "Key", "命中", "未命中", "总计", "命中率", "最后访问")
			fmt.Println(strings.Repeat("-", 80))
			for _, keyStats := range statsResp.TopKeys {
				hitRateColor := greenColor
				if keyStats.Total > 0 {
					if keyStats.HitRate < 70 {
						hitRateColor = redColor
					} else if keyStats.HitRate < 80 {
						hitRateColor = yellowColor
					}
				}
				fmt.Printf("%-50s %8d %8d %8d %s%9.2f%%%s %20s\n",
					truncateString(keyStats.Key, 50),
					keyStats.Hits,
					keyStats.Misses,
					keyStats.Total,
					hitRateColor,
					keyStats.HitRate,
					resetColor,
					keyStats.LastAccess.Format("2006-01-02 15:04:05"),
				)
			}
		}
	} else {
		fmt.Printf("\n%s提示: 暂无 Key 级别统计数据%s\n", yellowColor, resetColor)
	}

	fmt.Println(separator)
	return nil
}

// displayKeyStatsViaHTTP 通过 HTTP 接口显示指定 key 的统计信息
func displayKeyStatsViaHTTP(key string) error {
	// 构建请求 URL
	serverPort := config.Server.Port
	if serverPort == "" {
		serverPort = "8080" // 默认端口
	}
	u, err := url.Parse(fmt.Sprintf("http://localhost:%s/api/v1/internal/cache/hit-rate/stats", serverPort))
	if err != nil {
		return fmt.Errorf("构建请求 URL 失败: %w", err)
	}

	// 添加查询参数
	q := u.Query()
	q.Set("key", key)
	u.RawQuery = q.Encode()

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", u.String(), nil)
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
	httpResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 请求失败，状态码: %d, 响应: %s", httpResp.StatusCode, string(body))
	}

	// 解析响应
	var apiResponse dto.Response
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应中的 code
	if apiResponse.Code != 0 {
		msg := apiResponse.Message
		if msg == "" {
			msg = "未知错误"
		}
		return fmt.Errorf("接口返回错误: %s", msg)
	}

	// 解析数据
	var keyStatsResp resp.CacheHitRateKeyStatsResp
	keyStatsDataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return fmt.Errorf("序列化响应数据失败: %w", err)
	}
	if err := json.Unmarshal(keyStatsDataBytes, &keyStatsResp); err != nil {
		return fmt.Errorf("解析 Key 统计数据失败: %w", err)
	}

	// 显示统计信息
	separator := strings.Repeat("=", 80)
	fmt.Println(separator)
	fmt.Printf("                   Key 统计信息: %s\n", key)
	fmt.Println(separator)

	fmt.Printf("\n📊 统计详情:\n")
	fmt.Printf("  Key: %s%s%s\n", blueColor, keyStatsResp.Key, resetColor)
	fmt.Printf("  命中次数: %s%d%s\n", greenColor, keyStatsResp.Hits, resetColor)
	fmt.Printf("  未命中次数: %s%d%s\n", yellowColor, keyStatsResp.Misses, resetColor)
	fmt.Printf("  总请求数: %s%d%s\n", blueColor, keyStatsResp.Total, resetColor)
	if keyStatsResp.Total > 0 {
		hitRateColor := greenColor
		if keyStatsResp.HitRate < 70 {
			hitRateColor = redColor
		} else if keyStatsResp.HitRate < 80 {
			hitRateColor = yellowColor
		}
		fmt.Printf("  命中率: %s%.2f%%%s\n", hitRateColor, keyStatsResp.HitRate, resetColor)
	} else {
		fmt.Printf("  命中率: %sN/A%s (暂无数据)\n", yellowColor, resetColor)
	}
	fmt.Printf("  最后访问时间: %s%s%s\n", blueColor, keyStatsResp.LastAccess.Format("2006-01-02 15:04:05"), resetColor)

	fmt.Println(separator)
	return nil
}

// watchModeViaHTTP 通过 HTTP 接口实时监控模式
func watchModeViaHTTP() {
	interval := time.Duration(cacheHitRateWatchInterval) * time.Second
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}

	fmt.Println("\n开始实时监控缓存命中率统计... (按 Ctrl+C 退出)")
	fmt.Printf("更新间隔: %v\n", interval)
	fmt.Println(strings.Repeat("=", 80))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 立即显示一次
	if err := displayStatsViaHTTP(); err != nil {
		fmt.Printf("%s错误: 查询统计信息失败: %v%s\n", redColor, err, resetColor)
		return
	}

	// 循环更新
	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n正在退出监控...")
			return
		case <-ticker.C:
			// 清屏并显示新数据
			fmt.Print("\033[2J\033[H") // 清屏
			if err := displayStatsViaHTTP(); err != nil {
				fmt.Printf("%s错误: 查询统计信息失败: %v%s\n", redColor, err, resetColor)
				continue
			}
		}
	}
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
