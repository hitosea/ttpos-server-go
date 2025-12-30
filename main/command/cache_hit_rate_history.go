package command

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

var (
	cacheHitRateHistoryInstanceFlag string
	cacheHitRateHistoryStartTime    string
	cacheHitRateHistoryEndTime      string
	cacheHitRateHistoryAggregate    bool
	cacheHitRateHistoryKeyFlag      string
	cacheHitRateHistoryTopNFlag     int
	cacheHitRateHistorySortByFlag   string
)

// KeyStatsData 单个 key 的统计数据
// 支持大小写字段名兼容（优先使用大写，如果不存在则使用小写）
type KeyStatsData struct {
	// 大写字段名（优先）
	Hits       *int64   `json:"Hits,omitempty"`
	Misses     *int64   `json:"Misses,omitempty"`
	Total      *int64   `json:"Total,omitempty"`
	HitRate    *float64 `json:"HitRate,omitempty"`
	LastAccess *string  `json:"LastAccess,omitempty"`
	// 小写字段名（兼容）
	HitsLower       *int64   `json:"hits,omitempty"`
	MissesLower     *int64   `json:"misses,omitempty"`
	TotalLower      *int64   `json:"total,omitempty"`
	HitRateLower    *float64 `json:"hit_rate,omitempty"`
	LastAccessLower *string  `json:"last_access,omitempty"`
}

// GetHits 获取命中次数（兼容大小写，优先使用大写字段）
func (k *KeyStatsData) GetHits() int64 {
	if k.Hits != nil {
		return *k.Hits
	}
	if k.HitsLower != nil {
		return *k.HitsLower
	}
	return 0
}

// GetMisses 获取未命中次数（兼容大小写，优先使用大写字段）
func (k *KeyStatsData) GetMisses() int64 {
	if k.Misses != nil {
		return *k.Misses
	}
	if k.MissesLower != nil {
		return *k.MissesLower
	}
	return 0
}

// GetTotal 获取总访问次数（兼容大小写，优先使用大写字段）
func (k *KeyStatsData) GetTotal() int64 {
	if k.Total != nil {
		return *k.Total
	}
	if k.TotalLower != nil {
		return *k.TotalLower
	}
	return 0
}

// GetHitRate 获取命中率（兼容大小写，优先使用大写字段）
func (k *KeyStatsData) GetHitRate() float64 {
	if k.HitRate != nil {
		return *k.HitRate
	}
	if k.HitRateLower != nil {
		return *k.HitRateLower
	}
	return 0
}

// GetLastAccess 获取最后访问时间字符串（兼容大小写，优先使用大写字段）
func (k *KeyStatsData) GetLastAccess() string {
	if k.LastAccess != nil {
		return *k.LastAccess
	}
	if k.LastAccessLower != nil {
		return *k.LastAccessLower
	}
	return ""
}

// KeyStatsMap KeyStats JSON 的完整结构（key -> KeyStatsData）
type KeyStatsMap map[string]*KeyStatsData

// cacheHitRateHistoryCmd 缓存命中率历史查询命令
var cacheHitRateHistoryCmd = &cobra.Command{
	Use:   "cache-hit-rate-history",
	Short: "查询历史缓存命中率统计",
	Long:  "查询历史缓存命中率统计，支持按实例分组显示和聚合统计",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s\n", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化日志系统（必须在数据库初始化之前）
		if err := logger.Init(); err != nil {
			log.Fatalf("%sFailed to initialize logger: %v%s\n", redColor, err, resetColor)
		}

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 初始化数据库连接
		dbm := database.GetDBManager(config.Database)

		// 解析时间范围
		now := time.Now()
		startTime := now.AddDate(0, 0, -7) // 默认最近一周
		endTime := now

		if cacheHitRateHistoryStartTime != "" {
			parsedStartTime, err := time.Parse("2006-01-02 15:04:05", cacheHitRateHistoryStartTime)
			if err != nil {
				fmt.Printf("%s错误: 开始时间格式错误，请使用格式: 2006-01-02 15:04:05%s\n", redColor, resetColor)
				return
			}
			startTime = parsedStartTime
		}

		if cacheHitRateHistoryEndTime != "" {
			parsedEndTime, err := time.Parse("2006-01-02 15:04:05", cacheHitRateHistoryEndTime)
			if err != nil {
				fmt.Printf("%s错误: 结束时间格式错误，请使用格式: 2006-01-02 15:04:05%s\n", redColor, resetColor)
				return
			}
			endTime = parsedEndTime
		}

		// 查询快照
		repo := repository.NewCacheHitRateSnapshotRepo(dbm.GetDB(0)) // saas 库的 index 是 0
		snapshots, err := repo.GetSnapshotsByTimeRange(cacheHitRateHistoryInstanceFlag, startTime, endTime)
		if err != nil {
			fmt.Printf("%s错误: 查询快照失败: %v%s\n", redColor, err, resetColor)
			return
		}

		if len(snapshots) == 0 {
			fmt.Printf("%s提示: 在指定时间范围内没有找到快照数据%s\n", yellowColor, resetColor)
			return
		}

		// 如果指定了 --key，显示指定 key 的历史统计
		if cacheHitRateHistoryKeyFlag != "" {
			displayKeyHistoryStats(snapshots, cacheHitRateHistoryKeyFlag)
			return
		}

		// 如果指定了 --top，显示 Top N 的 key 历史统计
		if cacheHitRateHistoryTopNFlag > 0 {
			displayTopKeysHistoryStats(snapshots, cacheHitRateHistoryTopNFlag, cacheHitRateHistorySortByFlag)
			return
		}

		// 按实例分组
		instanceSnapshots := make(map[string][]*model.CacheHitRateSnapshot)
		for _, snapshot := range snapshots {
			instanceSnapshots[snapshot.InstanceID] = append(instanceSnapshots[snapshot.InstanceID], snapshot)
		}

		// 显示结果
		separator := strings.Repeat("=", 100)
		fmt.Println(separator)
		fmt.Printf("                   缓存命中率历史统计 (%s 至 %s)\n",
			startTime.Format("2006-01-02 15:04:05"),
			endTime.Format("2006-01-02 15:04:05"))
		fmt.Println(separator)

		// 按实例分组显示
		var instanceIDs []string
		for instanceID := range instanceSnapshots {
			instanceIDs = append(instanceIDs, instanceID)
		}
		sort.Strings(instanceIDs)

		var aggregateStats struct {
			Hits    int64
			Misses  int64
			Total   int64
			HitRate float64
		}

		for _, instanceID := range instanceIDs {
			snapshots := instanceSnapshots[instanceID]
			sort.Slice(snapshots, func(i, j int) bool {
				return snapshots[i].SnapshotTime.Before(snapshots[j].SnapshotTime)
			})

			fmt.Printf("\n📊 实例: %s%s%s\n", blueColor, instanceID, resetColor)
			fmt.Printf("%-20s %10s %10s %10s %10s %10s\n", "快照时间", "命中", "未命中", "总计", "命中率", "重启标记")
			fmt.Println(strings.Repeat("-", 100))

			var instanceStats struct {
				Hits    int64
				Misses  int64
				Total   int64
				HitRate float64
			}

			var previousSnapshot *model.CacheHitRateSnapshot
			for _, snapshot := range snapshots {
				// 计算增量
				if previousSnapshot != nil {
					if snapshot.IsRestart == 1 || snapshot.Total < previousSnapshot.Total {
						// 重启后的第一条快照，累加统计
						instanceStats.Hits += snapshot.Hits
						instanceStats.Misses += snapshot.Misses
						instanceStats.Total += snapshot.Total
					} else {
						// 正常情况，使用增量
						instanceStats.Hits += (snapshot.Hits - previousSnapshot.Hits)
						instanceStats.Misses += (snapshot.Misses - previousSnapshot.Misses)
						instanceStats.Total += (snapshot.Total - previousSnapshot.Total)
					}
				} else {
					// 第一条快照
					instanceStats.Hits += snapshot.Hits
					instanceStats.Misses += snapshot.Misses
					instanceStats.Total += snapshot.Total
				}

				restartMark := ""
				if snapshot.IsRestart == 1 {
					restartMark = fmt.Sprintf("%s重启%s", yellowColor, resetColor)
				}

				hitRateColor := greenColor
				if snapshot.HitRate < 70 {
					hitRateColor = redColor
				} else if snapshot.HitRate < 80 {
					hitRateColor = yellowColor
				}

				fmt.Printf("%-20s %10d %10d %10d %s%9.2f%%%s %10s\n",
					snapshot.SnapshotTime.Format("2006-01-02 15:04:05"),
					snapshot.Hits,
					snapshot.Misses,
					snapshot.Total,
					hitRateColor,
					snapshot.HitRate,
					resetColor,
					restartMark,
				)

				previousSnapshot = snapshot
			}

			// 计算实例总体命中率
			if instanceStats.Total > 0 {
				instanceStats.HitRate = float64(instanceStats.Hits) / float64(instanceStats.Total) * 100
			}

			// 显示实例汇总
			instanceHitRateColor := greenColor
			if instanceStats.HitRate < 70 {
				instanceHitRateColor = redColor
			} else if instanceStats.HitRate < 80 {
				instanceHitRateColor = yellowColor
			}

			fmt.Printf("%-20s %10d %10d %10d %s%9.2f%%%s\n",
				"实例汇总",
				instanceStats.Hits,
				instanceStats.Misses,
				instanceStats.Total,
				instanceHitRateColor,
				instanceStats.HitRate,
				resetColor,
			)

			// 累加到聚合统计
			aggregateStats.Hits += instanceStats.Hits
			aggregateStats.Misses += instanceStats.Misses
			aggregateStats.Total += instanceStats.Total
		}

		// 显示聚合统计
		if cacheHitRateHistoryAggregate && len(instanceIDs) > 1 {
			if aggregateStats.Total > 0 {
				aggregateStats.HitRate = float64(aggregateStats.Hits) / float64(aggregateStats.Total) * 100
			}

			aggregateHitRateColor := greenColor
			if aggregateStats.HitRate < 70 {
				aggregateHitRateColor = redColor
			} else if aggregateStats.HitRate < 80 {
				aggregateHitRateColor = yellowColor
			}

			fmt.Printf("\n%s总体统计（所有实例聚合）%s\n", blueColor, resetColor)
			fmt.Println(strings.Repeat("-", 100))
			fmt.Printf("%-20s %10d %10d %10d %s%9.2f%%%s\n",
				"总计",
				aggregateStats.Hits,
				aggregateStats.Misses,
				aggregateStats.Total,
				aggregateHitRateColor,
				aggregateStats.HitRate,
				resetColor,
			)
		}

		fmt.Println(separator)
	},
}

func init() {
	cacheHitRateHistoryCmd.Flags().StringVar(&cacheHitRateHistoryInstanceFlag, "instance", "", "指定实例ID（可选，不指定则查询所有实例）")
	cacheHitRateHistoryCmd.Flags().StringVar(&cacheHitRateHistoryStartTime, "start-time", "", "开始时间（格式：2006-01-02 15:04:05，默认：当前时间往前推7天）")
	cacheHitRateHistoryCmd.Flags().StringVar(&cacheHitRateHistoryEndTime, "end-time", "", "结束时间（格式：2006-01-02 15:04:05，默认：当前时间）")
	cacheHitRateHistoryCmd.Flags().BoolVar(&cacheHitRateHistoryAggregate, "aggregate", false, "是否聚合所有实例的数据（默认 false）")
	cacheHitRateHistoryCmd.Flags().StringVar(&cacheHitRateHistoryKeyFlag, "key", "", "查询指定 key 的历史统计信息")
	cacheHitRateHistoryCmd.Flags().IntVar(&cacheHitRateHistoryTopNFlag, "top", 0, "显示 Top N 的 key 历史统计（默认 0，不显示）")
	cacheHitRateHistoryCmd.Flags().StringVar(&cacheHitRateHistorySortByFlag, "sort", "total", "排序方式：total（总访问次数）、hits（命中次数）、hitrate（命中率），默认 total")

	rootCommand.AddCommand(cacheHitRateHistoryCmd)
}

// parseKeyStats 解析 KeyStats JSON 字符串
func parseKeyStats(keyStatsJSON string) (KeyStatsMap, error) {
	if keyStatsJSON == "" {
		return nil, nil
	}

	var keyStatsMap KeyStatsMap
	if err := json.Unmarshal([]byte(keyStatsJSON), &keyStatsMap); err != nil {
		return nil, err
	}

	return keyStatsMap, nil
}

// parseLastAccessTime 解析最后访问时间字符串
func parseLastAccessTime(lastAccessStr string, fallbackTime time.Time) time.Time {
	if lastAccessStr == "" {
		return fallbackTime
	}

	timeFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, lastAccessStr); err == nil {
			return t
		}
	}

	return fallbackTime
}

// displayKeyHistoryStats 显示指定 key 的历史统计信息
func displayKeyHistoryStats(snapshots []*model.CacheHitRateSnapshot, key string) {
	separator := strings.Repeat("=", 100)
	fmt.Println(separator)
	fmt.Printf("                   Key 历史统计: %s\n", key)
	fmt.Println(separator)

	// 按快照时间排序
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].SnapshotTime.Before(snapshots[j].SnapshotTime)
	})

	fmt.Printf("\n%-20s %10s %10s %10s %10s %20s\n", "快照时间", "命中", "未命中", "总计", "命中率", "最后访问")
	fmt.Println(strings.Repeat("-", 100))

	found := false
	for _, snapshot := range snapshots {
		if snapshot.KeyStats == "" {
			continue
		}

		// 解析 KeyStats JSON
		keyStatsMap, err := parseKeyStats(snapshot.KeyStats)
		if err != nil {
			continue
		}

		// 查找指定的 key
		keyStatsData, ok := keyStatsMap[key]
		if !ok || keyStatsData == nil {
			continue
		}

		found = true

		// 获取字段值
		ksHits := keyStatsData.GetHits()
		ksMisses := keyStatsData.GetMisses()
		ksTotal := keyStatsData.GetTotal()
		ksHitRate := keyStatsData.GetHitRate()
		lastAccessStr := keyStatsData.GetLastAccess()

		// 解析最后访问时间
		lastAccess := parseLastAccessTime(lastAccessStr, snapshot.SnapshotTime)

		hitRateColor := greenColor
		if ksTotal > 0 {
			if ksHitRate < 70 {
				hitRateColor = redColor
			} else if ksHitRate < 80 {
				hitRateColor = yellowColor
			}
		}

		fmt.Printf("%-20s %10d %10d %10d %s%9.2f%%%s %20s\n",
			snapshot.SnapshotTime.Format("2006-01-02 15:04:05"),
			ksHits,
			ksMisses,
			ksTotal,
			hitRateColor,
			ksHitRate,
			resetColor,
			lastAccess.Format("2006-01-02 15:04:05"),
		)
	}

	if !found {
		fmt.Printf("%s提示: 在指定时间范围内没有找到 key '%s' 的统计数据%s\n", yellowColor, key, resetColor)
	}

	fmt.Println(separator)
}

// displayTopKeysHistoryStats 显示 Top N 的 key 历史统计
func displayTopKeysHistoryStats(snapshots []*model.CacheHitRateSnapshot, topN int, sortBy string) {
	separator := strings.Repeat("=", 100)
	fmt.Println(separator)
	fmt.Printf("                   Top %d Key 历史统计 (按%s排序)\n", topN, getSortByLabel(sortBy))
	fmt.Println(separator)

	// 聚合所有快照中的 key 统计数据
	keyStatsAggregated := make(map[string]*keyHistoryStats)

	for _, snapshot := range snapshots {
		if snapshot.KeyStats == "" {
			continue
		}

		// 解析 KeyStats JSON
		keyStatsMap, err := parseKeyStats(snapshot.KeyStats)
		if err != nil {
			continue
		}

		// 遍历所有 key
		for keyName, keyStatsData := range keyStatsMap {
			if keyStatsData == nil {
				continue
			}

			// 初始化或获取聚合统计
			if keyStatsAggregated[keyName] == nil {
				keyStatsAggregated[keyName] = &keyHistoryStats{
					Key:        keyName,
					Hits:       0,
					Misses:     0,
					Total:      0,
					LastAccess: time.Time{},
				}
			}
			stats := keyStatsAggregated[keyName]

			// 累加统计数据（使用快照中的累计值）
			ksHits := keyStatsData.GetHits()
			ksMisses := keyStatsData.GetMisses()
			ksTotal := keyStatsData.GetTotal()

			// 使用最大值（因为快照是累计值）
			if ksTotal > stats.Total {
				stats.Total = ksTotal
				stats.Hits = ksHits
				stats.Misses = ksMisses
				if stats.Total > 0 {
					stats.HitRate = float64(stats.Hits) / float64(stats.Total) * 100
				}
			}

			// 更新最后访问时间
			lastAccessStr := keyStatsData.GetLastAccess()
			if lastAccessStr != "" {
				lastAccess := parseLastAccessTime(lastAccessStr, snapshot.SnapshotTime)
				if lastAccess.After(stats.LastAccess) {
					stats.LastAccess = lastAccess
				}
			}
		}
	}

	// 转换为切片并排序
	keyStatsList := make([]*keyHistoryStats, 0, len(keyStatsAggregated))
	for _, stats := range keyStatsAggregated {
		keyStatsList = append(keyStatsList, stats)
	}

	// 排序
	switch sortBy {
	case "hits":
		sort.Slice(keyStatsList, func(i, j int) bool {
			return keyStatsList[i].Hits > keyStatsList[j].Hits
		})
	case "hitrate":
		sort.Slice(keyStatsList, func(i, j int) bool {
			return keyStatsList[i].HitRate > keyStatsList[j].HitRate
		})
	default: // "total"
		sort.Slice(keyStatsList, func(i, j int) bool {
			return keyStatsList[i].Total > keyStatsList[j].Total
		})
	}

	// 显示 Top N
	if len(keyStatsList) == 0 {
		fmt.Printf("%s提示: 没有找到 Key 级别统计数据%s\n", yellowColor, resetColor)
		fmt.Println(separator)
		return
	}

	if topN > len(keyStatsList) {
		topN = len(keyStatsList)
	}

	fmt.Printf("\n%-50s %8s %8s %8s %10s %20s\n", "Key", "命中", "未命中", "总计", "命中率", "最后访问")
	fmt.Println(strings.Repeat("-", 100))

	for i := 0; i < topN; i++ {
		stats := keyStatsList[i]
		hitRateColor := greenColor
		if stats.Total > 0 {
			if stats.HitRate < 70 {
				hitRateColor = redColor
			} else if stats.HitRate < 80 {
				hitRateColor = yellowColor
			}
		}

		lastAccessStr := "N/A"
		if !stats.LastAccess.IsZero() {
			lastAccessStr = stats.LastAccess.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%-50s %8d %8d %8d %s%9.2f%%%s %20s\n",
			truncateString(stats.Key, 50),
			stats.Hits,
			stats.Misses,
			stats.Total,
			hitRateColor,
			stats.HitRate,
			resetColor,
			lastAccessStr,
		)
	}

	fmt.Println(separator)
}

// keyHistoryStats key 历史统计结构
type keyHistoryStats struct {
	Key        string
	Hits       int64
	Misses     int64
	Total      int64
	HitRate    float64
	LastAccess time.Time
}

// getSortByLabel 获取排序方式标签
func getSortByLabel(sortBy string) string {
	labels := map[string]string{
		"total":   "总访问次数",
		"hits":    "命中次数",
		"hitrate": "命中率",
	}
	if label, ok := labels[sortBy]; ok {
		return label
	}
	return "总访问次数"
}
