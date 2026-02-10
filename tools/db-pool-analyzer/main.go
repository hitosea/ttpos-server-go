package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
)

// 命令行参数
var (
	dsn          string // 数据库连接字符串
	reportType   string // 报告类型
	startTime    string // 开始时间
	endTime      string // 结束时间
	instanceID   string // 实例ID（可选）
	dbName       string // 数据库名称（可选）
	topN         int    // Top N 数量
	outputJSON   bool   // 是否输出JSON格式
	compareStart string // 对比时间段开始
	compareEnd   string // 对比时间段结束
)

// DBPoolStats 数据库连接池统计记录
type DBPoolStats struct {
	ID                uint64
	UUID              uint64
	InstanceID        string
	DBName            string
	MaxOpenConns      int
	OpenConns         int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDurationMs    int64
	MaxIdleClosed     int64
	MaxIdleTimeClosed int64
	MaxLifetimeClosed int64
	SampleTime        int64
	CreateTime        int64
}

// Analyzer 分析器
type Analyzer struct {
	db               *sql.DB
	startTime        int64
	endTime          int64
	instanceID       string
	dbName           string
	topN             int
	compareStartTime int64
	compareEndTime   int64
}

func init() {
	flag.StringVar(&dsn, "dsn", "", "MySQL DSN (例: user:pass@tcp(host:3306)/saas)")
	flag.StringVar(&reportType, "report", "summary", "报告类型: summary|history|health|top|trend|instance|config|leak|compare|efficiency|anomaly|recommend")
	flag.StringVar(&startTime, "start", "", "开始时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.StringVar(&endTime, "end", "", "结束时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.StringVar(&instanceID, "instance", "", "实例ID过滤 (可选)")
	flag.StringVar(&dbName, "db", "", "数据库名称过滤 (可选)")
	flag.IntVar(&topN, "top", 20, "Top N 数量")
	flag.BoolVar(&outputJSON, "json", false, "输出JSON格式")
	flag.StringVar(&compareStart, "compare-start", "", "对比时间段开始 (用于compare报告)")
	flag.StringVar(&compareEnd, "compare-end", "", "对比时间段结束 (用于compare报告)")
}

func main() {
	flag.Parse()

	if dsn == "" {
		printUsage()
		os.Exit(1)
	}

	// 设置默认时间范围（最近24小时）
	if startTime == "" {
		startTime = time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	}
	if endTime == "" {
		endTime = time.Now().Format("2006-01-02 15:04:05")
	}

	// 解析时间
	startTs, endTs := parseTimeRange(startTime, endTime)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("数据库连接测试失败: %v\n", err)
		os.Exit(1)
	}

	// 设置时区
	if _, err := db.Exec("SET time_zone = '+08:00'"); err != nil {
		fmt.Printf("设置时区失败: %v\n", err)
		os.Exit(1)
	}

	// 解析对比时间段
	var compareStartTs, compareEndTs int64
	if compareStart != "" && compareEnd != "" {
		compareStartTs, compareEndTs = parseTimeRange(compareStart, compareEnd)
	}

	analyzer := &Analyzer{
		db:               db,
		startTime:        startTs,
		endTime:          endTs,
		instanceID:       instanceID,
		dbName:           dbName,
		topN:             topN,
		compareStartTime: compareStartTs,
		compareEndTime:   compareEndTs,
	}

	// 执行报告
	switch reportType {
	case "summary":
		analyzer.SummaryReport()
	case "history":
		analyzer.HistoryReport()
	case "health":
		analyzer.HealthReport()
	case "top":
		analyzer.TopProblemsReport()
	case "trend":
		analyzer.TrendReport()
	case "instance":
		analyzer.InstanceReport()
	case "config":
		analyzer.ConfigAnalysisReport()
	case "leak":
		analyzer.LeakDetectionReport()
	case "compare":
		if compareStart == "" || compareEnd == "" {
			fmt.Println("错误: compare 报告需要 -compare-start 和 -compare-end 参数")
			os.Exit(1)
		}
		analyzer.CompareReport()
	case "efficiency":
		analyzer.EfficiencyReport()
	case "anomaly":
		analyzer.AnomalyReport()
	case "recommend":
		analyzer.RecommendReport()
	default:
		fmt.Printf("未知的报告类型: %s\n", reportType)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("数据库连接池统计分析工具 (只读)")
	fmt.Println(strings.Repeat("═", 90))
	fmt.Println("\n参数说明:")
	flag.PrintDefaults()

	fmt.Println("\n" + strings.Repeat("═", 90))
	fmt.Println("                              命令详解与示例")
	fmt.Println(strings.Repeat("═", 90))

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 1. summary - 总体概览（默认）                                                           │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 快速了解连接池整体状况，包括实例数、数据库数、最大使用量等                        │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report summary                                        │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report summary -start '2026-02-01' -end '2026-02-10'  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 2. history - 历史记录查看                                                               │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 查看连接池历史详细记录，支持按实例/数据库过滤                                     │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report history                                        │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report history -db shop123456                         │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report history -instance 'instance-1'                 │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report history -db shop123456 -instance 'instance-1'  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 3. health - 健康状况分析                                                                │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 自动识别连接池问题（使用率过高、等待过多、频繁关闭等）                            │")
	fmt.Println("│ 检测: 🔴严重(使用率>90%/等待>100次) 🟡警告(使用率>70%/等待>10次)                        │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report health                                         │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report health -instance 'instance-1'                  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 4. top - Top N 问题数据库                                                               │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 按问题分数排序，快速定位最需要关注的数据库                                        │")
	fmt.Println("│ 评分: 使用率(50分) + 等待次数×0.5 + 等待时长×0.01                                       │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report top                                            │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report top -top 10                                    │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report top -top 50 -instance 'instance-1'             │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 5. trend - 趋势分析（按小时）                                                           │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 按小时聚合统计，发现高峰时段和时间规律                                            │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report trend                                          │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report trend -start '2026-02-09' -end '2026-02-10'    │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report trend -instance 'instance-1'                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report trend -db shop123456                           │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 6. instance - 实例负载分析                                                              │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 对比各服务实例的负载情况，发现负载不均衡问题                                      │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report instance                                       │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report instance -start '2026-02-09'                   │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 7. config - 连接池配置分析                                                              │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 检查 MaxOpenConns/ConnMaxLifetime/MaxIdleConns 等配置是否合理                     │")
	fmt.Println("│ 检测: 配置过大(浪费资源)、配置过小(容易等待)、连接频繁关闭等                            │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report config                                         │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report config -db shop123456                          │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 8. leak - 连接泄漏检测                                                                  │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 检测连接泄漏特征（InUse持续增长、连接耗尽、使用率异常高）                         │")
	fmt.Println("│ 特征: 持续增长(70%以上采样点InUse增加) / 连接耗尽(InUse≈MaxOpen且Idle=0)                │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report leak                                           │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report leak -start '2026-02-09 10:00:00'              │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report leak -db shop123456                            │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 9. compare - 时间段对比                                                                 │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 对比两个时间段的连接池指标，用于升级前后/今天vs昨天对比                           │")
	fmt.Println("│ 必需: -compare-start 和 -compare-end 参数                                               │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   # 今天 vs 昨天                                                                        │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report compare \\                                      │")
	fmt.Println("│       -start '2026-02-10' -end '2026-02-11' \\                                           │")
	fmt.Println("│       -compare-start '2026-02-09' -compare-end '2026-02-10'                             │")
	fmt.Println("│                                                                                         │")
	fmt.Println("│   # 升级后 vs 升级前（精确到小时）                                                      │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report compare \\                                      │")
	fmt.Println("│       -start '2026-02-10 14:00:00' -end '2026-02-10 16:00:00' \\                         │")
	fmt.Println("│       -compare-start '2026-02-10 10:00:00' -compare-end '2026-02-10 12:00:00'           │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 10. efficiency - 连接效率分析                                                           │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 分析连接池资源利用效率，找出资源浪费的数据库                                      │")
	fmt.Println("│ 指标: 利用率(InUse/Open) / 配置效率(MaxInUse/MaxOpen) / 资源浪费率                      │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report efficiency                                     │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report efficiency -top 30                             │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report efficiency -instance 'instance-1'              │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 11. anomaly - 异常检测                                                                  │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 检测异常值（连接使用突增、等待次数突增、连接池满）                                │")
	fmt.Println("│ 阈值: InUse超过平均值3倍 / WaitCount超过平均值5倍 / InUse达到MaxOpen                    │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report anomaly                                        │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report anomaly -top 50                                │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report anomaly -db shop123456                         │")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n┌─────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 12. recommend - 配置优化建议                                                            │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 功能: 根据历史数据给出具体的配置优化建议                                                │")
	fmt.Println("│ 建议: MaxOpenConns调整值 / ConnMaxLifetime调整 / MaxIdleConns调整                       │")
	fmt.Println("│ 示例:                                                                                   │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report recommend                                      │")
	fmt.Println("│   ./db-pool-analyzer -dsn '$DSN' -report recommend -start '2026-02-01' -end '2026-02-10'│")
	fmt.Println("└─────────────────────────────────────────────────────────────────────────────────────────┘")

	fmt.Println("\n" + strings.Repeat("═", 90))
	fmt.Println("                                 常用诊断流程")
	fmt.Println(strings.Repeat("═", 90))

	fmt.Println(`
  用户反馈下单慢
       │
       ▼
  ┌─────────────────────────────────────────────────────┐
  │ ./db-pool-analyzer -dsn '$DSN' -report health       │  ← 快速扫描问题
  └─────────────────────────────────────────────────────┘
       │
       ▼
  ┌─────────────────────────────────────────────────────┐
  │ ./db-pool-analyzer -dsn '$DSN' -report anomaly      │  ← 找异常时间点
  └─────────────────────────────────────────────────────┘
       │
       ▼
  ┌─────────────────────────────────────────────────────┐
  │ ./db-pool-analyzer -dsn '$DSN' -report leak         │  ← 排查连接泄漏
  └─────────────────────────────────────────────────────┘
       │
       ▼
  ┌─────────────────────────────────────────────────────┐
  │ ./db-pool-analyzer -dsn '$DSN' -report recommend    │  ← 获取优化建议
  └─────────────────────────────────────────────────────┘
`)

	fmt.Println(strings.Repeat("═", 90))
	fmt.Println("提示: 将 $DSN 替换为实际的数据库连接字符串")
	fmt.Println("      例如: 'readonly:password@tcp(127.0.0.1:3306)/saas'")
	fmt.Println(strings.Repeat("═", 90))
}

func parseTimeRange(start, end string) (int64, int64) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var startTs, endTs int64

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, start, time.Local); err == nil {
			startTs = t.Unix()
			break
		}
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, end, time.Local); err == nil {
			if layout == "2006-01-02" {
				// 如果只有日期，结束时间设为当天23:59:59
				t = t.Add(24*time.Hour - time.Second)
			}
			endTs = t.Unix()
			break
		}
	}

	return startTs, endTs
}

// buildWhereClause 构建查询条件
func (a *Analyzer) buildWhereClause() (string, []any) {
	where := "WHERE create_time >= ? AND create_time <= ? AND delete_time = 0"
	args := []any{a.startTime, a.endTime}

	if a.instanceID != "" {
		where += " AND instance_id = ?"
		args = append(args, a.instanceID)
	}
	if a.dbName != "" {
		where += " AND db_name = ?"
		args = append(args, a.dbName)
	}

	return where, args
}

// queryRecords 查询记录
func (a *Analyzer) queryRecords() ([]DBPoolStats, error) {
	where, args := a.buildWhereClause()
	query := fmt.Sprintf(`
		SELECT id, uuid, instance_id, db_name, max_open_conns, open_conns,
		       in_use, idle, wait_count, wait_duration_ms,
		       max_idle_closed, max_idle_time_closed, max_lifetime_closed,
		       sample_time, create_time
		FROM ttpos_db_pool_stats
		%s
		ORDER BY sample_time ASC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DBPoolStats
	for rows.Next() {
		var r DBPoolStats
		err := rows.Scan(&r.ID, &r.UUID, &r.InstanceID, &r.DBName,
			&r.MaxOpenConns, &r.OpenConns, &r.InUse, &r.Idle,
			&r.WaitCount, &r.WaitDurationMs,
			&r.MaxIdleClosed, &r.MaxIdleTimeClosed, &r.MaxLifetimeClosed,
			&r.SampleTime, &r.CreateTime)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, nil
}

// SummaryReport 总体概览报告
func (a *Analyzer) SummaryReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 统计
	instanceSet := make(map[string]bool)
	dbSet := make(map[string]bool)
	var totalWaitCount, totalWaitDuration int64
	var maxInUse, maxOpenConns int

	for _, r := range records {
		instanceSet[r.InstanceID] = true
		dbSet[r.DBName] = true
		if r.WaitCount > totalWaitCount {
			totalWaitCount = r.WaitCount
		}
		if r.WaitDurationMs > totalWaitDuration {
			totalWaitDuration = r.WaitDurationMs
		}
		if r.InUse > maxInUse {
			maxInUse = r.InUse
		}
		if r.MaxOpenConns > maxOpenConns {
			maxOpenConns = r.MaxOpenConns
		}
	}

	// 输出
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("                    数据库连接池统计概览")
	fmt.Println(strings.Repeat("═", 70))

	fmt.Printf("\n时间范围: %s 至 %s\n",
		time.Unix(a.startTime, 0).Format("2006-01-02 15:04:05"),
		time.Unix(a.endTime, 0).Format("2006-01-02 15:04:05"))

	fmt.Println(strings.Repeat("─", 70))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "值"})
	table.SetBorder(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})

	table.Append([]string{"总记录数", fmt.Sprintf("%d", len(records))})
	table.Append([]string{"实例数量", fmt.Sprintf("%d", len(instanceSet))})
	table.Append([]string{"数据库数量", fmt.Sprintf("%d", len(dbSet))})
	table.Append([]string{"最大连接使用数", fmt.Sprintf("%d / %d", maxInUse, maxOpenConns)})
	table.Append([]string{"最大等待次数", fmt.Sprintf("%d", totalWaitCount)})
	table.Append([]string{"最大等待时长", fmt.Sprintf("%d ms", totalWaitDuration)})

	table.Render()

	// 显示实例列表
	fmt.Println("\n实例列表:")
	for inst := range instanceSet {
		fmt.Printf("  • %s\n", inst)
	}

	fmt.Println(strings.Repeat("═", 70))
}

// HistoryReport 历史记录报告
func (a *Analyzer) HistoryReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	fmt.Println(strings.Repeat("═", 130))
	fmt.Println("                                     连接池历史记录")
	fmt.Println(strings.Repeat("═", 130))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "实例", "数据库", "最大", "打开", "使用", "空闲", "等待次数", "等待时长(ms)", "生命周期关闭"})
	table.SetBorder(false)

	for _, r := range records {
		sampleTime := time.UnixMilli(r.SampleTime).Format("01-02 15:04")
		table.Append([]string{
			sampleTime,
			truncate(r.InstanceID, 15),
			truncate(r.DBName, 15),
			fmt.Sprintf("%d", r.MaxOpenConns),
			fmt.Sprintf("%d", r.OpenConns),
			fmt.Sprintf("%d", r.InUse),
			fmt.Sprintf("%d", r.Idle),
			fmt.Sprintf("%d", r.WaitCount),
			fmt.Sprintf("%d", r.WaitDurationMs),
			fmt.Sprintf("%d", r.MaxLifetimeClosed),
		})
	}

	table.Render()
	fmt.Printf("\n共 %d 条记录\n", len(records))
	fmt.Println(strings.Repeat("═", 130))
}

// HealthReport 健康状况分析
func (a *Analyzer) HealthReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 按数据库聚合
	type dbHealth struct {
		InstanceID        string
		DBName            string
		MaxInUse          int
		MaxOpenConns      int
		MaxWaitCount      int64
		MaxWaitDuration   int64
		MaxLifetimeClosed int64
		SampleCount       int
		Issues            []string
	}

	dbHealthMap := make(map[string]*dbHealth)

	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if dbHealthMap[key] == nil {
			dbHealthMap[key] = &dbHealth{
				InstanceID:   r.InstanceID,
				DBName:       r.DBName,
				MaxOpenConns: r.MaxOpenConns,
			}
		}

		h := dbHealthMap[key]
		h.SampleCount++

		if r.InUse > h.MaxInUse {
			h.MaxInUse = r.InUse
		}
		if r.WaitCount > h.MaxWaitCount {
			h.MaxWaitCount = r.WaitCount
		}
		if r.WaitDurationMs > h.MaxWaitDuration {
			h.MaxWaitDuration = r.WaitDurationMs
		}
		if r.MaxLifetimeClosed > h.MaxLifetimeClosed {
			h.MaxLifetimeClosed = r.MaxLifetimeClosed
		}
	}

	// 分析问题
	type issue struct {
		Level   string
		DB      string
		Problem string
		Details string
	}

	var issues []issue

	for _, h := range dbHealthMap {
		db := h.InstanceID + ":" + h.DBName

		// 连接使用率
		if h.MaxOpenConns > 0 {
			usage := float64(h.MaxInUse) / float64(h.MaxOpenConns) * 100
			if usage > 90 {
				issues = append(issues, issue{
					Level:   "🔴 严重",
					DB:      db,
					Problem: "连接使用率过高",
					Details: fmt.Sprintf("%.1f%% (%d/%d)", usage, h.MaxInUse, h.MaxOpenConns),
				})
			} else if usage > 70 {
				issues = append(issues, issue{
					Level:   "🟡 警告",
					DB:      db,
					Problem: "连接使用率较高",
					Details: fmt.Sprintf("%.1f%% (%d/%d)", usage, h.MaxInUse, h.MaxOpenConns),
				})
			}
		}

		// 等待次数
		if h.MaxWaitCount > 100 {
			issues = append(issues, issue{
				Level:   "🔴 严重",
				DB:      db,
				Problem: "连接等待次数过多",
				Details: fmt.Sprintf("%d 次, 累计 %d ms", h.MaxWaitCount, h.MaxWaitDuration),
			})
		} else if h.MaxWaitCount > 10 {
			issues = append(issues, issue{
				Level:   "🟡 警告",
				DB:      db,
				Problem: "存在连接等待",
				Details: fmt.Sprintf("%d 次, 累计 %d ms", h.MaxWaitCount, h.MaxWaitDuration),
			})
		}

		// 连接频繁关闭
		if h.MaxLifetimeClosed > 1000 {
			issues = append(issues, issue{
				Level:   "🟡 警告",
				DB:      db,
				Problem: "连接频繁因生命周期关闭",
				Details: fmt.Sprintf("累计 %d 次", h.MaxLifetimeClosed),
			})
		}
	}

	// 输出
	fmt.Println(strings.Repeat("═", 120))
	fmt.Println("                              连接池健康分析报告")
	fmt.Println(strings.Repeat("═", 120))

	if len(issues) == 0 {
		fmt.Println("\n✅ 所有数据库连接池状态正常")
	} else {
		fmt.Printf("\n发现 %d 个问题:\n\n", len(issues))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"级别", "数据库", "问题", "详情"})
		table.SetBorder(false)
		table.SetColWidth(40)

		for _, i := range issues {
			table.Append([]string{i.Level, truncate(i.DB, 35), i.Problem, i.Details})
		}

		table.Render()
	}

	// 统计摘要
	fmt.Printf("\n分析摘要:\n")
	fmt.Printf("  • 分析数据库数量: %d\n", len(dbHealthMap))
	fmt.Printf("  • 分析记录总数: %d\n", len(records))
	fmt.Printf("  • 发现问题数量: %d\n", len(issues))

	fmt.Println(strings.Repeat("═", 120))
}

// TopProblemsReport Top N 问题数据库报告
func (a *Analyzer) TopProblemsReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 计算问题分数
	type dbScore struct {
		InstanceID   string
		DBName       string
		Score        float64
		MaxInUse     int
		MaxOpenConns int
		WaitCount    int64
		WaitDuration int64
	}

	scores := make(map[string]*dbScore)

	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if scores[key] == nil {
			scores[key] = &dbScore{
				InstanceID:   r.InstanceID,
				DBName:       r.DBName,
				MaxOpenConns: r.MaxOpenConns,
			}
		}

		s := scores[key]
		if r.InUse > s.MaxInUse {
			s.MaxInUse = r.InUse
		}
		if r.WaitCount > s.WaitCount {
			s.WaitCount = r.WaitCount
		}
		if r.WaitDurationMs > s.WaitDuration {
			s.WaitDuration = r.WaitDurationMs
		}
	}

	// 计算分数
	for _, s := range scores {
		if s.MaxOpenConns > 0 {
			s.Score += float64(s.MaxInUse) / float64(s.MaxOpenConns) * 50
		}
		s.Score += float64(s.WaitCount) * 0.5
		s.Score += float64(s.WaitDuration) * 0.01
	}

	// 排序
	var scoreList []*dbScore
	for _, s := range scores {
		scoreList = append(scoreList, s)
	}
	sort.Slice(scoreList, func(i, j int) bool {
		return scoreList[i].Score > scoreList[j].Score
	})

	// 输出
	fmt.Println(strings.Repeat("═", 110))
	fmt.Printf("                           Top %d 问题数据库\n", a.topN)
	fmt.Println(strings.Repeat("═", 110))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"排名", "数据库", "问题分数", "最大使用", "连接上限", "等待次数", "等待时长(ms)"})
	table.SetBorder(false)

	count := a.topN
	if count > len(scoreList) {
		count = len(scoreList)
	}

	for i := 0; i < count; i++ {
		s := scoreList[i]
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			truncate(s.InstanceID+":"+s.DBName, 40),
			fmt.Sprintf("%.1f", s.Score),
			fmt.Sprintf("%d", s.MaxInUse),
			fmt.Sprintf("%d", s.MaxOpenConns),
			fmt.Sprintf("%d", s.WaitCount),
			fmt.Sprintf("%d", s.WaitDuration),
		})
	}

	table.Render()
	fmt.Println(strings.Repeat("═", 110))
}

// TrendReport 趋势分析报告
func (a *Analyzer) TrendReport() {
	where, args := a.buildWhereClause()
	query := fmt.Sprintf(`
		SELECT
			FROM_UNIXTIME(create_time, '%%Y-%%m-%%d %%H:00') as hour,
			COUNT(*) as sample_count,
			AVG(in_use) as avg_in_use,
			MAX(in_use) as max_in_use,
			AVG(idle) as avg_idle,
			MAX(wait_count) as max_wait_count,
			MAX(wait_duration_ms) as max_wait_duration
		FROM ttpos_db_pool_stats
		%s
		GROUP BY hour
		ORDER BY hour ASC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println(strings.Repeat("═", 100))
	fmt.Println("                              连接池趋势分析")
	fmt.Println(strings.Repeat("═", 100))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间段", "采样数", "平均使用", "最大使用", "平均空闲", "最大等待次数", "最大等待时长(ms)"})
	table.SetBorder(false)

	count := 0
	for rows.Next() {
		var hour string
		var sampleCount int
		var avgInUse, maxInUse, avgIdle float64
		var maxWaitCount, maxWaitDuration int64

		err := rows.Scan(&hour, &sampleCount, &avgInUse, &maxInUse, &avgIdle, &maxWaitCount, &maxWaitDuration)
		if err != nil {
			continue
		}

		table.Append([]string{
			hour,
			fmt.Sprintf("%d", sampleCount),
			fmt.Sprintf("%.1f", avgInUse),
			fmt.Sprintf("%.0f", maxInUse),
			fmt.Sprintf("%.1f", avgIdle),
			fmt.Sprintf("%d", maxWaitCount),
			fmt.Sprintf("%d", maxWaitDuration),
		})
		count++
	}

	if count == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	table.Render()
	fmt.Println(strings.Repeat("═", 100))
}

// InstanceReport 实例负载报告
func (a *Analyzer) InstanceReport() {
	where, args := a.buildWhereClause()
	query := fmt.Sprintf(`
		SELECT
			instance_id,
			COUNT(DISTINCT db_name) as db_count,
			COUNT(*) as sample_count,
			AVG(in_use) as avg_in_use,
			MAX(in_use) as max_in_use,
			SUM(wait_count) as total_wait_count,
			SUM(wait_duration_ms) as total_wait_duration
		FROM ttpos_db_pool_stats
		%s
		GROUP BY instance_id
		ORDER BY total_wait_count DESC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println(strings.Repeat("═", 110))
	fmt.Println("                              实例负载分析")
	fmt.Println(strings.Repeat("═", 110))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"实例", "数据库数", "采样数", "平均使用", "最大使用", "总等待次数", "总等待时长(ms)"})
	table.SetBorder(false)

	count := 0
	for rows.Next() {
		var instanceID string
		var dbCount, sampleCount int
		var avgInUse, maxInUse float64
		var totalWaitCount, totalWaitDuration int64

		err := rows.Scan(&instanceID, &dbCount, &sampleCount, &avgInUse, &maxInUse, &totalWaitCount, &totalWaitDuration)
		if err != nil {
			continue
		}

		table.Append([]string{
			truncate(instanceID, 30),
			fmt.Sprintf("%d", dbCount),
			fmt.Sprintf("%d", sampleCount),
			fmt.Sprintf("%.1f", avgInUse),
			fmt.Sprintf("%.0f", maxInUse),
			fmt.Sprintf("%d", totalWaitCount),
			fmt.Sprintf("%d", totalWaitDuration),
		})
		count++
	}

	if count == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	table.Render()
	fmt.Println(strings.Repeat("═", 110))
}

// OutputJSON 输出JSON格式
func OutputJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ConfigAnalysisReport 连接池配置分析报告
func (a *Analyzer) ConfigAnalysisReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 按数据库聚合
	type configStats struct {
		DBName            string
		InstanceID        string
		MaxOpenConns      int
		MaxInUse          int
		AvgInUse          float64
		MaxIdle           int
		AvgIdle           float64
		MaxIdleClosed     int64
		MaxIdleTimeClosed int64
		MaxLifetimeClosed int64
		SampleCount       int
	}

	statsMap := make(map[string]*configStats)
	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if statsMap[key] == nil {
			statsMap[key] = &configStats{
				DBName:       r.DBName,
				InstanceID:   r.InstanceID,
				MaxOpenConns: r.MaxOpenConns,
			}
		}
		s := statsMap[key]
		s.SampleCount++
		s.AvgInUse += float64(r.InUse)
		s.AvgIdle += float64(r.Idle)
		if r.InUse > s.MaxInUse {
			s.MaxInUse = r.InUse
		}
		if r.Idle > s.MaxIdle {
			s.MaxIdle = r.Idle
		}
		if r.MaxIdleClosed > s.MaxIdleClosed {
			s.MaxIdleClosed = r.MaxIdleClosed
		}
		if r.MaxIdleTimeClosed > s.MaxIdleTimeClosed {
			s.MaxIdleTimeClosed = r.MaxIdleTimeClosed
		}
		if r.MaxLifetimeClosed > s.MaxLifetimeClosed {
			s.MaxLifetimeClosed = r.MaxLifetimeClosed
		}
	}

	// 计算平均值
	for _, s := range statsMap {
		if s.SampleCount > 0 {
			s.AvgInUse /= float64(s.SampleCount)
			s.AvgIdle /= float64(s.SampleCount)
		}
	}

	fmt.Println(strings.Repeat("═", 130))
	fmt.Println("                                    连接池配置分析")
	fmt.Println(strings.Repeat("═", 130))

	type configIssue struct {
		DB      string
		Type    string
		Current string
		Issue   string
		Suggest string
	}

	var issues []configIssue

	for _, s := range statsMap {
		db := s.InstanceID + ":" + s.DBName

		// 检查 MaxOpenConns 是否过大（浪费资源）
		if s.MaxOpenConns > 0 && float64(s.MaxInUse) < float64(s.MaxOpenConns)*0.3 {
			issues = append(issues, configIssue{
				DB:      db,
				Type:    "MaxOpenConns",
				Current: fmt.Sprintf("%d", s.MaxOpenConns),
				Issue:   fmt.Sprintf("设置过大，最大仅使用 %.0f%%", float64(s.MaxInUse)/float64(s.MaxOpenConns)*100),
				Suggest: fmt.Sprintf("建议设为 %d", max(s.MaxInUse*2, 10)),
			})
		}

		// 检查 MaxOpenConns 是否过小（容易等待）
		if s.MaxOpenConns > 0 && float64(s.MaxInUse) > float64(s.MaxOpenConns)*0.9 {
			issues = append(issues, configIssue{
				DB:      db,
				Type:    "MaxOpenConns",
				Current: fmt.Sprintf("%d", s.MaxOpenConns),
				Issue:   fmt.Sprintf("设置过小，最大使用 %.0f%%", float64(s.MaxInUse)/float64(s.MaxOpenConns)*100),
				Suggest: fmt.Sprintf("建议设为 %d", s.MaxOpenConns*2),
			})
		}

		// 检查 ConnMaxLifetime（如果关闭次数过多）
		if s.MaxLifetimeClosed > 500 {
			issues = append(issues, configIssue{
				DB:      db,
				Type:    "ConnMaxLifetime",
				Current: "未知",
				Issue:   fmt.Sprintf("生命周期关闭连接 %d 次，过于频繁", s.MaxLifetimeClosed),
				Suggest: "建议增大 ConnMaxLifetime 值",
			})
		}

		// 检查 MaxIdleConns（如果空闲关闭次数过多）
		if s.MaxIdleClosed > 500 {
			issues = append(issues, configIssue{
				DB:      db,
				Type:    "MaxIdleConns",
				Current: "未知",
				Issue:   fmt.Sprintf("超过空闲上限关闭 %d 次", s.MaxIdleClosed),
				Suggest: "建议增大 MaxIdleConns 值",
			})
		}

		// 检查 ConnMaxIdleTime
		if s.MaxIdleTimeClosed > 500 {
			issues = append(issues, configIssue{
				DB:      db,
				Type:    "ConnMaxIdleTime",
				Current: "未知",
				Issue:   fmt.Sprintf("空闲超时关闭 %d 次", s.MaxIdleTimeClosed),
				Suggest: "建议增大 ConnMaxIdleTime 值",
			})
		}
	}

	if len(issues) == 0 {
		fmt.Println("\n✅ 所有数据库连接池配置正常")
	} else {
		fmt.Printf("\n发现 %d 个配置问题:\n\n", len(issues))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"数据库", "配置项", "当前值", "问题", "建议"})
		table.SetBorder(false)
		table.SetColWidth(30)

		for _, i := range issues {
			table.Append([]string{truncate(i.DB, 30), i.Type, i.Current, i.Issue, i.Suggest})
		}
		table.Render()
	}

	fmt.Println(strings.Repeat("═", 130))
}

// LeakDetectionReport 连接泄漏检测报告
func (a *Analyzer) LeakDetectionReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 按数据库分组，检测 InUse 是否持续增长
	type dbTimeSeries struct {
		InstanceID string
		DBName     string
		Records    []DBPoolStats
	}

	seriesMap := make(map[string]*dbTimeSeries)
	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if seriesMap[key] == nil {
			seriesMap[key] = &dbTimeSeries{
				InstanceID: r.InstanceID,
				DBName:     r.DBName,
			}
		}
		seriesMap[key].Records = append(seriesMap[key].Records, r)
	}

	fmt.Println(strings.Repeat("═", 120))
	fmt.Println("                              连接泄漏检测")
	fmt.Println(strings.Repeat("═", 120))

	type leakSuspect struct {
		DB           string
		Trend        string
		StartInUse   int
		EndInUse     int
		MaxInUse     int
		MaxOpenConns int
		Severity     string
	}

	var suspects []leakSuspect

	for _, series := range seriesMap {
		if len(series.Records) < 3 {
			continue
		}

		// 计算趋势
		first := series.Records[0]
		last := series.Records[len(series.Records)-1]
		maxInUse := 0
		increasingCount := 0
		prevInUse := first.InUse

		for _, r := range series.Records {
			if r.InUse > maxInUse {
				maxInUse = r.InUse
			}
			if r.InUse > prevInUse {
				increasingCount++
			}
			prevInUse = r.InUse
		}

		// 检测泄漏特征
		db := series.InstanceID + ":" + series.DBName

		// 特征1：InUse 持续增长
		if float64(increasingCount) > float64(len(series.Records))*0.7 && last.InUse > first.InUse*2 {
			suspects = append(suspects, leakSuspect{
				DB:           db,
				Trend:        "持续增长",
				StartInUse:   first.InUse,
				EndInUse:     last.InUse,
				MaxInUse:     maxInUse,
				MaxOpenConns: first.MaxOpenConns,
				Severity:     "🔴 严重",
			})
		}

		// 特征2：InUse 接近 MaxOpenConns 且 Idle 为 0
		if last.InUse > 0 && last.MaxOpenConns > 0 &&
			float64(last.InUse) > float64(last.MaxOpenConns)*0.9 && last.Idle == 0 {
			suspects = append(suspects, leakSuspect{
				DB:           db,
				Trend:        "连接耗尽",
				StartInUse:   first.InUse,
				EndInUse:     last.InUse,
				MaxInUse:     maxInUse,
				MaxOpenConns: first.MaxOpenConns,
				Severity:     "🔴 严重",
			})
		}

		// 特征3：InUse 异常高（超过正常水平）
		if maxInUse > 50 && first.MaxOpenConns > 0 && float64(maxInUse) > float64(first.MaxOpenConns)*0.8 {
			suspects = append(suspects, leakSuspect{
				DB:           db,
				Trend:        "使用率过高",
				StartInUse:   first.InUse,
				EndInUse:     last.InUse,
				MaxInUse:     maxInUse,
				MaxOpenConns: first.MaxOpenConns,
				Severity:     "🟡 警告",
			})
		}
	}

	if len(suspects) == 0 {
		fmt.Println("\n✅ 未检测到连接泄漏迹象")
	} else {
		fmt.Printf("\n发现 %d 个疑似泄漏:\n\n", len(suspects))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"级别", "数据库", "特征", "起始InUse", "结束InUse", "最大InUse", "连接上限"})
		table.SetBorder(false)

		for _, s := range suspects {
			table.Append([]string{
				s.Severity,
				truncate(s.DB, 35),
				s.Trend,
				fmt.Sprintf("%d", s.StartInUse),
				fmt.Sprintf("%d", s.EndInUse),
				fmt.Sprintf("%d", s.MaxInUse),
				fmt.Sprintf("%d", s.MaxOpenConns),
			})
		}
		table.Render()

		fmt.Println("\n提示: 连接泄漏通常表现为 InUse 持续增长不释放，请检查是否有未关闭的数据库连接")
	}

	fmt.Println(strings.Repeat("═", 120))
}

// CompareReport 时间段对比报告
func (a *Analyzer) CompareReport() {
	// 查询当前时间段
	records1, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询当前时间段失败: %v\n", err)
		return
	}

	// 查询对比时间段
	originalStart := a.startTime
	originalEnd := a.endTime
	a.startTime = a.compareStartTime
	a.endTime = a.compareEndTime
	records2, err := a.queryRecords()
	a.startTime = originalStart
	a.endTime = originalEnd

	if err != nil {
		fmt.Printf("查询对比时间段失败: %v\n", err)
		return
	}

	// 聚合统计
	type periodStats struct {
		SampleCount   int
		AvgInUse      float64
		MaxInUse      int
		TotalWait     int64
		TotalWaitMs   int64
		DBCount       int
		InstanceCount int
	}

	calcStats := func(records []DBPoolStats) periodStats {
		var stats periodStats
		dbSet := make(map[string]bool)
		instSet := make(map[string]bool)

		for _, r := range records {
			stats.SampleCount++
			stats.AvgInUse += float64(r.InUse)
			if r.InUse > stats.MaxInUse {
				stats.MaxInUse = r.InUse
			}
			if r.WaitCount > stats.TotalWait {
				stats.TotalWait = r.WaitCount
			}
			if r.WaitDurationMs > stats.TotalWaitMs {
				stats.TotalWaitMs = r.WaitDurationMs
			}
			dbSet[r.DBName] = true
			instSet[r.InstanceID] = true
		}

		if stats.SampleCount > 0 {
			stats.AvgInUse /= float64(stats.SampleCount)
		}
		stats.DBCount = len(dbSet)
		stats.InstanceCount = len(instSet)

		return stats
	}

	stats1 := calcStats(records1)
	stats2 := calcStats(records2)

	fmt.Println(strings.Repeat("═", 100))
	fmt.Println("                              时间段对比分析")
	fmt.Println(strings.Repeat("═", 100))

	fmt.Printf("\n当前时间段: %s 至 %s\n",
		time.Unix(originalStart, 0).Format("2006-01-02 15:04"),
		time.Unix(originalEnd, 0).Format("2006-01-02 15:04"))
	fmt.Printf("对比时间段: %s 至 %s\n",
		time.Unix(a.compareStartTime, 0).Format("2006-01-02 15:04"),
		time.Unix(a.compareEndTime, 0).Format("2006-01-02 15:04"))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "当前", "对比", "变化", "变化率"})
	table.SetBorder(false)

	addRow := func(name string, curr, comp float64, format string) {
		diff := curr - comp
		var rate float64
		if comp != 0 {
			rate = diff / comp * 100
		}

		rateStr := fmt.Sprintf("%.1f%%", rate)
		if rate > 10 {
			rateStr = "📈 +" + rateStr
		} else if rate < -10 {
			rateStr = "📉 " + rateStr
		}

		table.Append([]string{
			name,
			fmt.Sprintf(format, curr),
			fmt.Sprintf(format, comp),
			fmt.Sprintf(format, diff),
			rateStr,
		})
	}

	fmt.Println()
	addRow("采样数", float64(stats1.SampleCount), float64(stats2.SampleCount), "%.0f")
	addRow("数据库数", float64(stats1.DBCount), float64(stats2.DBCount), "%.0f")
	addRow("平均使用连接", stats1.AvgInUse, stats2.AvgInUse, "%.1f")
	addRow("最大使用连接", float64(stats1.MaxInUse), float64(stats2.MaxInUse), "%.0f")
	addRow("最大等待次数", float64(stats1.TotalWait), float64(stats2.TotalWait), "%.0f")
	addRow("最大等待时长(ms)", float64(stats1.TotalWaitMs), float64(stats2.TotalWaitMs), "%.0f")

	table.Render()
	fmt.Println(strings.Repeat("═", 100))
}

// EfficiencyReport 连接效率分析报告
func (a *Analyzer) EfficiencyReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 按数据库聚合
	type effStats struct {
		DBName       string
		InstanceID   string
		MaxOpenConns int
		AvgInUse     float64
		AvgIdle      float64
		AvgOpen      float64
		MaxInUse     int
		SampleCount  int
	}

	statsMap := make(map[string]*effStats)
	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if statsMap[key] == nil {
			statsMap[key] = &effStats{
				DBName:       r.DBName,
				InstanceID:   r.InstanceID,
				MaxOpenConns: r.MaxOpenConns,
			}
		}
		s := statsMap[key]
		s.SampleCount++
		s.AvgInUse += float64(r.InUse)
		s.AvgIdle += float64(r.Idle)
		s.AvgOpen += float64(r.OpenConns)
		if r.InUse > s.MaxInUse {
			s.MaxInUse = r.InUse
		}
	}

	// 计算平均值和效率
	type effResult struct {
		DB            string
		MaxOpenConns  int
		AvgOpen       float64
		AvgInUse      float64
		AvgIdle       float64
		Utilization   float64 // 利用率 = InUse / Open
		Waste         float64 // 浪费率 = (MaxOpen - MaxInUse) / MaxOpen
		ConfigEffRate float64 // 配置效率 = MaxInUse / MaxOpen
	}

	var results []effResult
	for _, s := range statsMap {
		if s.SampleCount > 0 {
			s.AvgInUse /= float64(s.SampleCount)
			s.AvgIdle /= float64(s.SampleCount)
			s.AvgOpen /= float64(s.SampleCount)
		}

		r := effResult{
			DB:           s.InstanceID + ":" + s.DBName,
			MaxOpenConns: s.MaxOpenConns,
			AvgOpen:      s.AvgOpen,
			AvgInUse:     s.AvgInUse,
			AvgIdle:      s.AvgIdle,
		}

		if s.AvgOpen > 0 {
			r.Utilization = s.AvgInUse / s.AvgOpen * 100
		}
		if s.MaxOpenConns > 0 {
			r.Waste = float64(s.MaxOpenConns-s.MaxInUse) / float64(s.MaxOpenConns) * 100
			r.ConfigEffRate = float64(s.MaxInUse) / float64(s.MaxOpenConns) * 100
		}

		results = append(results, r)
	}

	// 按配置效率排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].ConfigEffRate < results[j].ConfigEffRate
	})

	fmt.Println(strings.Repeat("═", 130))
	fmt.Println("                                    连接效率分析")
	fmt.Println(strings.Repeat("═", 130))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"数据库", "最大连接", "平均打开", "平均使用", "平均空闲", "利用率", "配置效率", "资源浪费"})
	table.SetBorder(false)

	count := a.topN
	if count > len(results) {
		count = len(results)
	}

	for i := 0; i < count; i++ {
		r := results[i]
		table.Append([]string{
			truncate(r.DB, 35),
			fmt.Sprintf("%d", r.MaxOpenConns),
			fmt.Sprintf("%.1f", r.AvgOpen),
			fmt.Sprintf("%.1f", r.AvgInUse),
			fmt.Sprintf("%.1f", r.AvgIdle),
			fmt.Sprintf("%.1f%%", r.Utilization),
			fmt.Sprintf("%.1f%%", r.ConfigEffRate),
			fmt.Sprintf("%.1f%%", r.Waste),
		})
	}

	table.Render()

	// 统计摘要
	var totalWaste float64
	for _, r := range results {
		totalWaste += r.Waste
	}
	avgWaste := totalWaste / float64(len(results))

	fmt.Printf("\n摘要:\n")
	fmt.Printf("  • 分析数据库: %d 个\n", len(results))
	fmt.Printf("  • 平均资源浪费: %.1f%%\n", avgWaste)
	if avgWaste > 50 {
		fmt.Println("  • 💡 建议: 整体配置偏高，可适当减少 MaxOpenConns 节省资源")
	}

	fmt.Println(strings.Repeat("═", 130))
}

// AnomalyReport 异常检测报告
func (a *Analyzer) AnomalyReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 按数据库分组计算基线
	type dbBaseline struct {
		InstanceID  string
		DBName      string
		AvgInUse    float64
		StdInUse    float64
		AvgWait     float64
		MaxInUse    int
		MaxWait     int64
		Records     []DBPoolStats
		SampleCount int
	}

	baselineMap := make(map[string]*dbBaseline)
	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if baselineMap[key] == nil {
			baselineMap[key] = &dbBaseline{
				InstanceID: r.InstanceID,
				DBName:     r.DBName,
			}
		}
		b := baselineMap[key]
		b.Records = append(b.Records, r)
		b.SampleCount++
		b.AvgInUse += float64(r.InUse)
		b.AvgWait += float64(r.WaitCount)
		if r.InUse > b.MaxInUse {
			b.MaxInUse = r.InUse
		}
		if r.WaitCount > b.MaxWait {
			b.MaxWait = r.WaitCount
		}
	}

	// 计算平均值和标准差
	for _, b := range baselineMap {
		if b.SampleCount > 0 {
			b.AvgInUse /= float64(b.SampleCount)
			b.AvgWait /= float64(b.SampleCount)
		}

		// 计算标准差
		var sumSq float64
		for _, r := range b.Records {
			diff := float64(r.InUse) - b.AvgInUse
			sumSq += diff * diff
		}
		if b.SampleCount > 1 {
			b.StdInUse = sumSq / float64(b.SampleCount-1)
			if b.StdInUse > 0 {
				b.StdInUse = b.StdInUse // sqrt would be needed for real std
			}
		}
	}

	fmt.Println(strings.Repeat("═", 120))
	fmt.Println("                              异常检测报告")
	fmt.Println(strings.Repeat("═", 120))

	type anomaly struct {
		Time     string
		DB       string
		Type     string
		Value    string
		Baseline string
		Severity string
	}

	var anomalies []anomaly

	for _, b := range baselineMap {
		db := b.InstanceID + ":" + b.DBName

		for _, r := range b.Records {
			sampleTime := time.UnixMilli(r.SampleTime).Format("01-02 15:04")

			// 检测 InUse 异常（超过平均值 3 倍）
			if b.AvgInUse > 0 && float64(r.InUse) > b.AvgInUse*3 {
				anomalies = append(anomalies, anomaly{
					Time:     sampleTime,
					DB:       db,
					Type:     "连接使用突增",
					Value:    fmt.Sprintf("%d", r.InUse),
					Baseline: fmt.Sprintf("平均 %.1f", b.AvgInUse),
					Severity: "🔴 严重",
				})
			}

			// 检测等待次数突增
			if b.AvgWait > 0 && float64(r.WaitCount) > b.AvgWait*5 {
				anomalies = append(anomalies, anomaly{
					Time:     sampleTime,
					DB:       db,
					Type:     "等待次数突增",
					Value:    fmt.Sprintf("%d", r.WaitCount),
					Baseline: fmt.Sprintf("平均 %.1f", b.AvgWait),
					Severity: "🔴 严重",
				})
			}

			// 检测连接池满
			if r.MaxOpenConns > 0 && r.InUse >= r.MaxOpenConns {
				anomalies = append(anomalies, anomaly{
					Time:     sampleTime,
					DB:       db,
					Type:     "连接池已满",
					Value:    fmt.Sprintf("%d/%d", r.InUse, r.MaxOpenConns),
					Baseline: "-",
					Severity: "🔴 严重",
				})
			}
		}
	}

	if len(anomalies) == 0 {
		fmt.Println("\n✅ 未检测到异常")
	} else {
		fmt.Printf("\n发现 %d 个异常:\n\n", len(anomalies))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"级别", "时间", "数据库", "异常类型", "当前值", "基线"})
		table.SetBorder(false)

		// 只显示 top N
		count := a.topN
		if count > len(anomalies) {
			count = len(anomalies)
		}

		for i := 0; i < count; i++ {
			a := anomalies[i]
			table.Append([]string{a.Severity, a.Time, truncate(a.DB, 30), a.Type, a.Value, a.Baseline})
		}
		table.Render()

		if len(anomalies) > count {
			fmt.Printf("\n... 还有 %d 个异常未显示 (使用 -top 参数查看更多)\n", len(anomalies)-count)
		}
	}

	fmt.Println(strings.Repeat("═", 120))
}

// RecommendReport 配置优化建议报告
func (a *Analyzer) RecommendReport() {
	records, err := a.queryRecords()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("在指定时间范围内没有找到统计数据")
		return
	}

	// 聚合统计
	type dbStats struct {
		DBName            string
		InstanceID        string
		MaxOpenConns      int
		MaxInUse          int
		AvgInUse          float64
		MaxWaitCount      int64
		MaxWaitDuration   int64
		MaxLifetimeClosed int64
		MaxIdleClosed     int64
		SampleCount       int
	}

	statsMap := make(map[string]*dbStats)
	for _, r := range records {
		key := r.InstanceID + ":" + r.DBName
		if statsMap[key] == nil {
			statsMap[key] = &dbStats{
				DBName:       r.DBName,
				InstanceID:   r.InstanceID,
				MaxOpenConns: r.MaxOpenConns,
			}
		}
		s := statsMap[key]
		s.SampleCount++
		s.AvgInUse += float64(r.InUse)
		if r.InUse > s.MaxInUse {
			s.MaxInUse = r.InUse
		}
		if r.WaitCount > s.MaxWaitCount {
			s.MaxWaitCount = r.WaitCount
		}
		if r.WaitDurationMs > s.MaxWaitDuration {
			s.MaxWaitDuration = r.WaitDurationMs
		}
		if r.MaxLifetimeClosed > s.MaxLifetimeClosed {
			s.MaxLifetimeClosed = r.MaxLifetimeClosed
		}
		if r.MaxIdleClosed > s.MaxIdleClosed {
			s.MaxIdleClosed = r.MaxIdleClosed
		}
	}

	// 计算平均值
	for _, s := range statsMap {
		if s.SampleCount > 0 {
			s.AvgInUse /= float64(s.SampleCount)
		}
	}

	fmt.Println(strings.Repeat("═", 130))
	fmt.Println("                                    配置优化建议")
	fmt.Println(strings.Repeat("═", 130))

	type recommendation struct {
		DB          string
		Config      string
		Current     string
		Recommended string
		Reason      string
		Priority    string
	}

	var recs []recommendation

	for _, s := range statsMap {
		db := s.InstanceID + ":" + s.DBName

		// MaxOpenConns 建议
		if s.MaxOpenConns > 0 {
			usageRate := float64(s.MaxInUse) / float64(s.MaxOpenConns)

			if usageRate < 0.3 && s.MaxOpenConns > 20 {
				// 配置过大
				recommended := max(s.MaxInUse*2, 10)
				recs = append(recs, recommendation{
					DB:          db,
					Config:      "MaxOpenConns",
					Current:     fmt.Sprintf("%d", s.MaxOpenConns),
					Recommended: fmt.Sprintf("%d", recommended),
					Reason:      fmt.Sprintf("使用率仅 %.0f%%，可节省资源", usageRate*100),
					Priority:    "🟢 低",
				})
			} else if usageRate > 0.8 || s.MaxWaitCount > 10 {
				// 配置过小
				recommended := s.MaxOpenConns * 2
				recs = append(recs, recommendation{
					DB:          db,
					Config:      "MaxOpenConns",
					Current:     fmt.Sprintf("%d", s.MaxOpenConns),
					Recommended: fmt.Sprintf("%d", recommended),
					Reason:      fmt.Sprintf("使用率 %.0f%%，等待 %d 次", usageRate*100, s.MaxWaitCount),
					Priority:    "🔴 高",
				})
			}
		}

		// ConnMaxLifetime 建议
		if s.MaxLifetimeClosed > 500 {
			recs = append(recs, recommendation{
				DB:          db,
				Config:      "ConnMaxLifetime",
				Current:     "未知",
				Recommended: "增大 50%",
				Reason:      fmt.Sprintf("生命周期关闭 %d 次过多", s.MaxLifetimeClosed),
				Priority:    "🟡 中",
			})
		}

		// MaxIdleConns 建议
		if s.MaxIdleClosed > 500 {
			recs = append(recs, recommendation{
				DB:          db,
				Config:      "MaxIdleConns",
				Current:     "未知",
				Recommended: fmt.Sprintf(">= %d", int(s.AvgInUse)+5),
				Reason:      fmt.Sprintf("空闲上限关闭 %d 次", s.MaxIdleClosed),
				Priority:    "🟡 中",
			})
		}
	}

	if len(recs) == 0 {
		fmt.Println("\n✅ 当前配置已优化，暂无建议")
	} else {
		// 按优先级排序
		sort.Slice(recs, func(i, j int) bool {
			return recs[i].Priority > recs[j].Priority
		})

		fmt.Printf("\n共 %d 条优化建议:\n\n", len(recs))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"优先级", "数据库", "配置项", "当前值", "建议值", "原因"})
		table.SetBorder(false)
		table.SetColWidth(25)

		for _, r := range recs {
			table.Append([]string{r.Priority, truncate(r.DB, 25), r.Config, r.Current, r.Recommended, truncate(r.Reason, 30)})
		}
		table.Render()
	}

	// 全局建议
	fmt.Println("\n📋 通用最佳实践:")
	fmt.Println("  • MaxOpenConns: 建议设为预期最大并发数的 1.5-2 倍")
	fmt.Println("  • MaxIdleConns: 建议设为 MaxOpenConns 的 50%-80%")
	fmt.Println("  • ConnMaxLifetime: 建议设为 5-10 分钟，避免使用过期连接")
	fmt.Println("  • ConnMaxIdleTime: 建议设为 1-3 分钟，及时释放空闲资源")

	fmt.Println(strings.Repeat("═", 130))
}
