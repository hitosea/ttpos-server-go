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
	dsn         string // 数据库连接字符串
	reportType  string // 报告类型
	startTime   string // 开始时间
	endTime     string // 结束时间
	companyUuid uint64 // 商户UUID（可选）
	action      string // 操作类型（可选）
	source      string // 来源终端（可选）
	topN        int    // Top N 数量
	outputJSON  bool   // 是否输出JSON格式
)

func init() {
	flag.StringVar(&dsn, "dsn", "", "MySQL DSN (例: user:pass@tcp(host:3306)/dbname)")
	flag.StringVar(&reportType, "report", "summary", "报告类型: summary|slow|error|trend|company|action|terminal")
	flag.StringVar(&startTime, "start", "", "开始时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.StringVar(&endTime, "end", "", "结束时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.Uint64Var(&companyUuid, "company", 0, "商户UUID (可选，0表示全部)")
	flag.StringVar(&action, "action", "", "操作类型过滤 (可选)")
	flag.StringVar(&source, "source", "", "来源终端过滤 (可选)")
	flag.IntVar(&topN, "top", 20, "Top N 数量")
	flag.BoolVar(&outputJSON, "json", false, "输出JSON格式")
}

func main() {
	flag.Parse()

	if dsn == "" {
		fmt.Println("错误: 必须提供数据库连接字符串 (-dsn)")
		fmt.Println("\n使用方法:")
		flag.PrintDefaults()
		fmt.Println("\n报告类型说明:")
		fmt.Println("  summary  - 总体概览报告")
		fmt.Println("  slow     - 慢请求分析 (>3秒)")
		fmt.Println("  error    - 错误分析报告")
		fmt.Println("  trend    - 趋势分析 (按小时)")
		fmt.Println("  company  - 门店性能对比")
		fmt.Println("  action   - 接口性能排名")
		fmt.Println("  terminal - 终端性能对比")
		fmt.Println("\n示例:")
		fmt.Println("  # 查看今天的总体概览")
		fmt.Println("  ./duration-analyzer -dsn 'user:pass@tcp(host:3306)/saas' -report summary -start 2026-02-05")
		fmt.Println("")
		fmt.Println("  # 查看慢请求")
		fmt.Println("  ./duration-analyzer -dsn 'user:pass@tcp(host:3306)/saas' -report slow -start 2026-02-05 -top 50")
		fmt.Println("")
		fmt.Println("  # 查看某个门店的错误")
		fmt.Println("  ./duration-analyzer -dsn 'user:pass@tcp(host:3306)/saas' -report error -company 123456 -start 2026-02-05")
		os.Exit(1)
	}

	// 设置默认时间范围（今天）
	if startTime == "" {
		startTime = time.Now().Format("2006-01-02")
	}
	if endTime == "" {
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	// 解析时间
	startMs, endMs := parseTimeRange(startTime, endTime)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		fmt.Printf("数据库连接测试失败: %v\n", err)
		os.Exit(1)
	}

	// 创建分析器
	analyzer := &Analyzer{
		db:          db,
		startTimeMs: startMs,
		endTimeMs:   endMs,
		companyUuid: companyUuid,
		action:      action,
		source:      source,
		topN:        topN,
	}

	// 执行报告
	switch reportType {
	case "summary":
		analyzer.SummaryReport()
	case "slow":
		analyzer.SlowRequestReport()
	case "error":
		analyzer.ErrorReport()
	case "trend":
		analyzer.TrendReport()
	case "company":
		analyzer.CompanyReport()
	case "action":
		analyzer.ActionReport()
	case "terminal":
		analyzer.TerminalReport()
	default:
		fmt.Printf("未知的报告类型: %s\n", reportType)
		os.Exit(1)
	}
}

// parseTimeRange 解析时间范围
func parseTimeRange(start, end string) (int64, int64) {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}

	var startTime, endTime time.Time
	var err error

	for _, layout := range layouts {
		startTime, err = time.ParseInLocation(layout, start, time.Local)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Printf("解析开始时间失败: %v\n", err)
		os.Exit(1)
	}

	for _, layout := range layouts {
		endTime, err = time.ParseInLocation(layout, end, time.Local)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Printf("解析结束时间失败: %v\n", err)
		os.Exit(1)
	}

	return startTime.UnixMilli(), endTime.UnixMilli()
}

// Analyzer 分析器
type Analyzer struct {
	db          *sql.DB
	startTimeMs int64
	endTimeMs   int64
	companyUuid uint64
	action      string
	source      string
	topN        int
}

// buildWhereClause 构建WHERE条件
func (a *Analyzer) buildWhereClause() (string, []any) {
	conditions := []string{"start_time >= ? AND start_time < ?"}
	args := []any{a.startTimeMs, a.endTimeMs}

	if a.companyUuid > 0 {
		conditions = append(conditions, "company_uuid = ?")
		args = append(args, a.companyUuid)
	}
	if a.action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, a.action)
	}
	if a.source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, a.source)
	}

	return strings.Join(conditions, " AND "), args
}

// SummaryReport 总体概览报告
func (a *Analyzer) SummaryReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    操作耗时分析报告 - 总体概览")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 60))

	// 基础统计
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MIN(duration_ms) as min_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow_count
		FROM ttpos_order_operation_duration
		WHERE %s
	`, where)

	var total, success, fail, avgMs, minMs, maxMs, slowCount int64
	err := a.db.QueryRow(query, args...).Scan(&total, &success, &fail, &avgMs, &minMs, &maxMs, &slowCount)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	errorRate := float64(0)
	if total > 0 {
		errorRate = float64(fail) * 100 / float64(total)
	}

	fmt.Printf("\n📊 基础统计\n")
	fmt.Printf("   总请求数:    %d\n", total)
	fmt.Printf("   成功请求:    %d\n", success)
	fmt.Printf("   失败请求:    %d\n", fail)
	fmt.Printf("   错误率:      %.2f%%\n", errorRate)
	fmt.Printf("   慢请求数:    %d (>3秒)\n", slowCount)

	fmt.Printf("\n⏱️  耗时统计\n")
	fmt.Printf("   平均耗时:    %d ms\n", avgMs)
	fmt.Printf("   最小耗时:    %d ms\n", minMs)
	fmt.Printf("   最大耗时:    %d ms\n", maxMs)

	// 百分位数
	a.printPercentiles(where, args)

	// Top 5 最慢接口
	fmt.Printf("\n🐢 最慢接口 Top 5\n")
	a.printTopSlow(where, args, 5)

	// Top 5 错误最多接口
	fmt.Printf("\n❌ 错误最多接口 Top 5\n")
	a.printTopErrors(where, args, 5)
}

// printPercentiles 打印百分位数
func (a *Analyzer) printPercentiles(where string, args []any) {
	query := fmt.Sprintf(`
		SELECT duration_ms
		FROM ttpos_order_operation_duration
		WHERE %s
		ORDER BY duration_ms ASC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	var durations []int
	for rows.Next() {
		var d int
		rows.Scan(&d)
		durations = append(durations, d)
	}

	if len(durations) == 0 {
		return
	}

	sort.Ints(durations)
	n := len(durations)

	p50 := durations[percentileIndex(n, 50)]
	p90 := durations[percentileIndex(n, 90)]
	p95 := durations[percentileIndex(n, 95)]
	p99 := durations[percentileIndex(n, 99)]

	fmt.Printf("\n📈 百分位数\n")
	fmt.Printf("   P50:         %d ms\n", p50)
	fmt.Printf("   P90:         %d ms\n", p90)
	fmt.Printf("   P95:         %d ms\n", p95)
	fmt.Printf("   P99:         %d ms\n", p99)
}

func percentileIndex(n, p int) int {
	idx := (n * p / 100) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return idx
}

// printTopSlow 打印最慢接口
func (a *Analyzer) printTopSlow(where string, args []any, limit int) {
	query := fmt.Sprintf(`
		SELECT action, source,
			   COUNT(*) as cnt,
			   ROUND(AVG(duration_ms)) as avg_ms,
			   MAX(duration_ms) as max_ms
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY action, source
		ORDER BY avg_ms DESC
		LIMIT %d
	`, where, limit)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"操作", "来源", "请求数", "平均耗时", "最大耗时"})
	table.SetBorder(false)

	for rows.Next() {
		var act, src string
		var cnt, avgMs, maxMs int
		rows.Scan(&act, &src, &cnt, &avgMs, &maxMs)
		table.Append([]string{act, src, fmt.Sprintf("%d", cnt), fmt.Sprintf("%d ms", avgMs), fmt.Sprintf("%d ms", maxMs)})
	}
	table.Render()
}

// printTopErrors 打印错误最多接口
func (a *Analyzer) printTopErrors(where string, args []any, limit int) {
	query := fmt.Sprintf(`
		SELECT action, source,
			   COUNT(*) as fail_cnt,
			   (SELECT COUNT(*) FROM ttpos_order_operation_duration t2
			    WHERE t2.action = t1.action AND t2.source = t1.source
			    AND t2.start_time >= ? AND t2.start_time < ?) as total_cnt
		FROM ttpos_order_operation_duration t1
		WHERE %s AND status = 0
		GROUP BY action, source
		ORDER BY fail_cnt DESC
		LIMIT %d
	`, where, limit)

	// 添加时间参数用于子查询
	fullArgs := append(args, a.startTimeMs, a.endTimeMs)

	rows, err := a.db.Query(query, fullArgs...)
	if err != nil {
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"操作", "来源", "失败数", "总数", "错误率"})
	table.SetBorder(false)

	for rows.Next() {
		var act, src string
		var failCnt, totalCnt int
		rows.Scan(&act, &src, &failCnt, &totalCnt)
		errorRate := float64(failCnt) * 100 / float64(totalCnt)
		table.Append([]string{act, src, fmt.Sprintf("%d", failCnt), fmt.Sprintf("%d", totalCnt), fmt.Sprintf("%.2f%%", errorRate)})
	}
	table.Render()
}

// SlowRequestReport 慢请求报告
func (a *Analyzer) SlowRequestReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                         慢请求分析报告 (>3秒)")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	query := fmt.Sprintf(`
		SELECT action, source, company_uuid, sale_bill_uuid,
			   duration_ms, request_path, status, error_msg,
			   FROM_UNIXTIME(start_time/1000) as start_at
		FROM ttpos_order_operation_duration
		WHERE %s AND duration_ms > 3000
		ORDER BY duration_ms DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "操作", "来源", "门店", "账单", "耗时", "状态", "路径"})
	table.SetBorder(false)
	table.SetColWidth(40)

	for rows.Next() {
		var act, src, path, errMsg, startAt string
		var companyUuid, saleBillUuid uint64
		var durationMs, status int
		rows.Scan(&act, &src, &companyUuid, &saleBillUuid, &durationMs, &path, &status, &errMsg, &startAt)

		statusStr := "✓"
		if status == 0 {
			statusStr = "✗"
		}

		table.Append([]string{
			startAt,
			act,
			src,
			fmt.Sprintf("%d", companyUuid),
			fmt.Sprintf("%d", saleBillUuid),
			fmt.Sprintf("%d ms", durationMs),
			statusStr,
			truncate(path, 30),
		})
	}
	table.Render()
}

// ErrorReport 错误分析报告
func (a *Analyzer) ErrorReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                              错误分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 错误类型统计
	fmt.Printf("\n📊 错误类型统计\n")
	query := fmt.Sprintf(`
		SELECT error_msg, COUNT(*) as cnt
		FROM ttpos_order_operation_duration
		WHERE %s AND status = 0 AND error_msg != ''
		GROUP BY error_msg
		ORDER BY cnt DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"错误信息", "出现次数"})
	table.SetBorder(false)
	table.SetColWidth(80)

	for rows.Next() {
		var errMsg string
		var cnt int
		rows.Scan(&errMsg, &cnt)
		table.Append([]string{truncate(errMsg, 70), fmt.Sprintf("%d", cnt)})
	}
	table.Render()

	// 错误最多的接口
	fmt.Printf("\n❌ 错误最多的接口\n")
	a.printTopErrors(where, args, a.topN)
}

// TrendReport 趋势分析报告
func (a *Analyzer) TrendReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                              趋势分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	query := fmt.Sprintf(`
		SELECT
			FROM_UNIXTIME(FLOOR(start_time/3600000)*3600, '%%Y-%%m-%%d %%H:00') as hour,
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY hour
		ORDER BY hour
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "请求数", "失败数", "错误率", "平均耗时", "最大耗时", "慢请求"})
	table.SetBorder(false)

	for rows.Next() {
		var hour string
		var total, fail, avgMs, maxMs, slow int
		rows.Scan(&hour, &total, &fail, &avgMs, &maxMs, &slow)
		errorRate := float64(fail) * 100 / float64(total)
		table.Append([]string{
			hour,
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()
}

// CompanyReport 门店性能对比报告
func (a *Analyzer) CompanyReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            门店性能对比报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	query := fmt.Sprintf(`
		SELECT
			company_uuid,
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY company_uuid
		ORDER BY avg_ms DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"门店UUID", "请求数", "失败数", "错误率", "平均耗时", "最大耗时", "慢请求"})
	table.SetBorder(false)

	for rows.Next() {
		var companyUuid uint64
		var total, fail, avgMs, maxMs, slow int
		rows.Scan(&companyUuid, &total, &fail, &avgMs, &maxMs, &slow)
		errorRate := float64(fail) * 100 / float64(total)
		table.Append([]string{
			fmt.Sprintf("%d", companyUuid),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()
}

// ActionReport 接口性能排名报告
func (a *Analyzer) ActionReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            接口性能排名报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	query := fmt.Sprintf(`
		SELECT
			action,
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MIN(duration_ms) as min_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY action
		ORDER BY avg_ms DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"操作类型", "请求数", "失败", "错误率", "平均", "最小", "最大", "慢请求"})
	table.SetBorder(false)

	for rows.Next() {
		var action string
		var total, fail, avgMs, minMs, maxMs, slow int
		rows.Scan(&action, &total, &fail, &avgMs, &minMs, &maxMs, &slow)
		errorRate := float64(fail) * 100 / float64(total)
		table.Append([]string{
			action,
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.1f%%", errorRate),
			fmt.Sprintf("%d", avgMs),
			fmt.Sprintf("%d", minMs),
			fmt.Sprintf("%d", maxMs),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()
}

// TerminalReport 终端性能对比报告
func (a *Analyzer) TerminalReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            终端性能对比报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	query := fmt.Sprintf(`
		SELECT
			source,
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MIN(duration_ms) as min_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY source
		ORDER BY total DESC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"终端", "请求数", "失败数", "错误率", "平均耗时", "最小", "最大", "慢请求"})
	table.SetBorder(false)

	for rows.Next() {
		var src string
		var total, fail, avgMs, minMs, maxMs, slow int
		rows.Scan(&src, &total, &fail, &avgMs, &minMs, &maxMs, &slow)
		errorRate := float64(fail) * 100 / float64(total)
		table.Append([]string{
			src,
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", minMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// 用于JSON输出的结构体
type SummaryResult struct {
	TimeRange   string `json:"time_range"`
	Total       int64  `json:"total"`
	Success     int64  `json:"success"`
	Fail        int64  `json:"fail"`
	ErrorRate   float64 `json:"error_rate"`
	AvgMs       int64  `json:"avg_ms"`
	MinMs       int64  `json:"min_ms"`
	MaxMs       int64  `json:"max_ms"`
	SlowCount   int64  `json:"slow_count"`
	Percentiles struct {
		P50 int `json:"p50"`
		P90 int `json:"p90"`
		P95 int `json:"p95"`
		P99 int `json:"p99"`
	} `json:"percentiles"`
}

func outputAsJSON(data any) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}
