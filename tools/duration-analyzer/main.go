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
	companyUuid  uint64 // 商户UUID（可选）
	action       string // 操作类型（可选）
	source       string // 来源终端（可选）
	topN         int    // Top N 数量
	outputJSON   bool   // 是否输出JSON格式
	compareStart string // 对比时间段开始（用于compare报告）
	compareEnd   string // 对比时间段结束（用于compare报告）
	billUuid     uint64 // 账单UUID（用于bill-trace报告）
	staffUuid    uint64 // 员工UUID（可选）
)

func init() {
	flag.StringVar(&dsn, "dsn", "", "MySQL DSN (例: user:pass@tcp(host:3306)/dbname)")
	flag.StringVar(&reportType, "report", "summary", "报告类型: summary|slow|error|trend|company|action|terminal|nodes|peak|device|staff|instance|bill-trace|compare|distribution|anomaly|path|concurrency")
	flag.StringVar(&startTime, "start", "", "开始时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.StringVar(&endTime, "end", "", "结束时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)")
	flag.Uint64Var(&companyUuid, "company", 0, "商户UUID (可选，0表示全部)")
	flag.StringVar(&action, "action", "", "操作类型过滤 (可选)")
	flag.StringVar(&source, "source", "", "来源终端过滤 (可选)")
	flag.IntVar(&topN, "top", 20, "Top N 数量")
	flag.BoolVar(&outputJSON, "json", false, "输出JSON格式")
	flag.StringVar(&compareStart, "compare-start", "", "对比时间段开始 (用于compare报告)")
	flag.StringVar(&compareEnd, "compare-end", "", "对比时间段结束 (用于compare报告)")
	flag.Uint64Var(&billUuid, "bill", 0, "账单UUID (用于bill-trace报告)")
	flag.Uint64Var(&staffUuid, "staff", 0, "员工UUID (可选)")
}

func main() {
	flag.Parse()

	if dsn == "" {
		fmt.Println("错误: 必须提供数据库连接字符串 (-dsn)")
		fmt.Println("\n使用方法:")
		flag.PrintDefaults()
		fmt.Println("\n报告类型说明:")
		fmt.Println("  summary      - 总体概览报告")
		fmt.Println("  slow         - 慢请求分析 (>3秒)")
		fmt.Println("  error        - 错误分析报告")
		fmt.Println("  trend        - 趋势分析 (按小时)")
		fmt.Println("  company      - 门店性能对比")
		fmt.Println("  action       - 接口性能排名")
		fmt.Println("  terminal     - 终端性能对比")
		fmt.Println("  nodes        - 时间节点分析 (各阶段耗时)")
		fmt.Println("  peak         - 高峰时段分析 (识别业务高峰)")
		fmt.Println("  device       - 设备性能分析 (识别问题设备)")
		fmt.Println("  staff        - 员工效率分析")
		fmt.Println("  instance     - 服务实例负载分析")
		fmt.Println("  bill-trace   - 账单操作追踪 (需要 -bill 参数)")
		fmt.Println("  compare      - 时间段对比 (需要 -compare-start/-compare-end)")
		fmt.Println("  distribution - 耗时分布直方图")
		fmt.Println("  anomaly      - 异常检测 (P99以上的请求)")
		fmt.Println("  path         - 请求路径分析")
		fmt.Println("  concurrency  - 并发分析")
		fmt.Println("\n" + strings.Repeat("─", 60))
		fmt.Println("使用示例:")
		fmt.Println(strings.Repeat("─", 60))

		fmt.Println("\n【场景1】日常巡检 - 快速了解系统整体状况")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report summary           # 总体概览")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report trend             # 按小时趋势")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report distribution      # 耗时分布直方图")

		fmt.Println("\n【场景2】性能问题排查 - 找出慢接口和慢请求")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report slow -top 50      # 慢请求详情")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report action            # 接口性能排名")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report path              # 请求路径分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report nodes -action checkout  # 某接口耗时节点")

		fmt.Println("\n【场景3】高峰期和并发分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report peak              # 识别高峰时段")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report concurrency       # 并发/QPS分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report peak -company 123456  # 某门店高峰时段")

		fmt.Println("\n【场景4】设备和实例分析 - 找出问题设备/实例")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report device            # 设备性能分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report device -company 123456  # 某门店设备对比")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report instance          # 服务实例负载分析")

		fmt.Println("\n【场景5】员工效率分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report staff             # 员工操作效率排名")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report staff -company 123456  # 某门店员工分析")

		fmt.Println("\n【场景6】错误和异常分析")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report error             # 错误类型统计")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report anomaly           # 异常检测(超P99)")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report error -company 123456  # 某门店错误")

		fmt.Println("\n【场景7】多维度对比")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report company           # 门店性能对比")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report terminal          # 终端类型对比")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report action -source pos  # 只看pos终端")

		fmt.Println("\n【场景8】时间段对比 - 升级前后/今天vs昨天")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report compare \\")
		fmt.Println("      -start 2026-02-06 -end 2026-02-07 \\")
		fmt.Println("      -compare-start 2026-02-05 -compare-end 2026-02-06")

		fmt.Println("\n【场景9】账单追踪 - 追踪单个账单的完整操作链")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report bill-trace -bill 123456789")

		fmt.Println("\n【场景10】精确时间范围查询")
		fmt.Println("  ./duration-analyzer -dsn '$DSN' -report summary \\")
		fmt.Println("      -start '2026-02-05 10:00:00' -end '2026-02-05 12:00:00'")

		fmt.Println("\n" + strings.Repeat("─", 60))
		fmt.Println("提示: 将 $DSN 替换为实际的数据库连接字符串")
		fmt.Println("      例如: 'user:pass@tcp(127.0.0.1:3306)/saas'")
		fmt.Println(strings.Repeat("─", 60))
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

	// 设置 session 时区为东八区（北京时间）
	if _, err := db.Exec("SET time_zone = '+08:00'"); err != nil {
		fmt.Printf("设置时区失败: %v\n", err)
		os.Exit(1)
	}

	// 解析对比时间段
	var compareStartMs, compareEndMs int64
	if compareStart != "" && compareEnd != "" {
		compareStartMs, compareEndMs = parseTimeRange(compareStart, compareEnd)
	}

	// 创建分析器
	analyzer := &Analyzer{
		db:             db,
		startTimeMs:    startMs,
		endTimeMs:      endMs,
		companyUuid:    companyUuid,
		action:         action,
		source:         source,
		topN:           topN,
		compareStartMs: compareStartMs,
		compareEndMs:   compareEndMs,
		billUuid:       billUuid,
		staffUuid:      staffUuid,
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
	case "nodes":
		analyzer.NodesReport()
	case "peak":
		analyzer.PeakReport()
	case "device":
		analyzer.DeviceReport()
	case "staff":
		analyzer.StaffReport()
	case "instance":
		analyzer.InstanceReport()
	case "bill-trace":
		analyzer.BillTraceReport()
	case "compare":
		analyzer.CompareReport()
	case "distribution":
		analyzer.DistributionReport()
	case "anomaly":
		analyzer.AnomalyReport()
	case "path":
		analyzer.PathReport()
	case "concurrency":
		analyzer.ConcurrencyReport()
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
	db             *sql.DB
	startTimeMs    int64
	endTimeMs      int64
	companyUuid    uint64
	action         string
	source         string
	topN           int
	compareStartMs int64
	compareEndMs   int64
	billUuid       uint64
	staffUuid      uint64
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

// TimeNode 时间节点
type TimeNode struct {
	Name     string `json:"name"`
	OffsetMs int64  `json:"offset_ms"`
}

// NodesReport 时间节点分析报告
func (a *Analyzer) NodesReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            时间节点分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 查询包含时间节点的记录
	query := fmt.Sprintf(`
		SELECT action, time_nodes, duration_ms
		FROM ttpos_order_operation_duration
		WHERE %s AND time_nodes IS NOT NULL AND time_nodes != '' AND time_nodes != '[]'
		LIMIT %d
	`, where, a.topN*10) // 多查一些用于统计

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	// 按节点名称聚合统计
	nodeStats := make(map[string][]int64) // nodeName -> []offsetMs
	actionNodes := make(map[string]map[string][]int64) // action -> nodeName -> []offsetMs

	for rows.Next() {
		var actionName, timeNodesJSON string
		var durationMs int
		rows.Scan(&actionName, &timeNodesJSON, &durationMs)

		var nodes []TimeNode
		if err := json.Unmarshal([]byte(timeNodesJSON), &nodes); err != nil {
			continue
		}

		for _, node := range nodes {
			nodeStats[node.Name] = append(nodeStats[node.Name], node.OffsetMs)

			if actionNodes[actionName] == nil {
				actionNodes[actionName] = make(map[string][]int64)
			}
			actionNodes[actionName][node.Name] = append(actionNodes[actionName][node.Name], node.OffsetMs)
		}
	}

	if len(nodeStats) == 0 {
		fmt.Println("\n暂无时间节点数据")
		fmt.Println("提示: 需要在代码中使用 ctx.Mark(\"node_name\") 记录时间节点")
		return
	}

	// 1. 全局节点统计
	fmt.Printf("\n📊 全局节点耗时统计\n")
	a.printNodeStats(nodeStats)

	// 2. 按 Action 分组的节点统计
	fmt.Printf("\n📋 各接口节点耗时明细\n")
	for actionName, nodes := range actionNodes {
		fmt.Printf("\n  [%s]\n", actionName)
		a.printNodeStatsIndent(nodes, "    ")
	}
}

// printNodeStats 打印节点统计
func (a *Analyzer) printNodeStats(nodeStats map[string][]int64) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"节点名称", "样本数", "平均偏移", "最小", "最大", "P50", "P95"})
	table.SetBorder(false)

	// 按平均值排序
	type nodeStat struct {
		name string
		data []int64
	}
	var stats []nodeStat
	for name, data := range nodeStats {
		stats = append(stats, nodeStat{name, data})
	}
	sort.Slice(stats, func(i, j int) bool {
		return avg(stats[i].data) < avg(stats[j].data)
	})

	for _, s := range stats {
		data := s.data
		sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
		n := len(data)

		table.Append([]string{
			s.name,
			fmt.Sprintf("%d", n),
			fmt.Sprintf("%d ms", avg(data)),
			fmt.Sprintf("%d ms", data[0]),
			fmt.Sprintf("%d ms", data[n-1]),
			fmt.Sprintf("%d ms", data[percentileIndex(n, 50)]),
			fmt.Sprintf("%d ms", data[percentileIndex(n, 95)]),
		})
	}
	table.Render()
}

// printNodeStatsIndent 打印节点统计（带缩进）
func (a *Analyzer) printNodeStatsIndent(nodeStats map[string][]int64, indent string) {
	// 按平均值排序
	type nodeStat struct {
		name string
		data []int64
	}
	var stats []nodeStat
	for name, data := range nodeStats {
		stats = append(stats, nodeStat{name, data})
	}
	sort.Slice(stats, func(i, j int) bool {
		return avg(stats[i].data) < avg(stats[j].data)
	})

	for _, s := range stats {
		data := s.data
		sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
		n := len(data)
		fmt.Printf("%s%-20s: avg=%4dms, min=%4dms, max=%4dms, p95=%4dms (n=%d)\n",
			indent, s.name, avg(data), data[0], data[n-1], data[percentileIndex(n, 95)], n)
	}
}

// avg 计算平均值
func avg(data []int64) int64 {
	if len(data) == 0 {
		return 0
	}
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum / int64(len(data))
}

// PeakReport 高峰时段分析报告
func (a *Analyzer) PeakReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            高峰时段分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 按小时统计请求量，找出高峰时段
	fmt.Printf("\n📊 每小时请求量分布\n")
	query := fmt.Sprintf(`
		SELECT
			HOUR(FROM_UNIXTIME(start_time/1000)) as hour,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
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

	type hourStat struct {
		hour   int
		total  int
		avgMs  int
		fail   int
		slow   int
	}
	var stats []hourStat
	var totalRequests int
	var maxTotal int

	for rows.Next() {
		var h hourStat
		rows.Scan(&h.hour, &h.total, &h.avgMs, &h.fail, &h.slow)
		stats = append(stats, h)
		totalRequests += h.total
		if h.total > maxTotal {
			maxTotal = h.total
		}
	}

	if len(stats) == 0 {
		fmt.Println("暂无数据")
		return
	}

	avgPerHour := totalRequests / len(stats)
	peakThreshold := avgPerHour * 2 // 超过平均值2倍视为高峰

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时段", "请求数", "占比", "平均耗时", "失败", "慢请求", "状态"})
	table.SetBorder(false)

	var peakHours []int
	var offPeakAvgMs, peakAvgMs int
	var offPeakCount, peakCount int

	for _, h := range stats {
		percent := float64(h.total) * 100 / float64(totalRequests)
		isPeak := h.total >= peakThreshold

		status := ""
		if isPeak {
			status = "🔥 高峰"
			peakHours = append(peakHours, h.hour)
			peakAvgMs += h.avgMs * h.total
			peakCount += h.total
		} else {
			offPeakAvgMs += h.avgMs * h.total
			offPeakCount += h.total
		}

		// 简单的柱状图
		barLen := h.total * 20 / maxTotal
		bar := strings.Repeat("█", barLen)

		table.Append([]string{
			fmt.Sprintf("%02d:00-%02d:00", h.hour, (h.hour+1)%24),
			fmt.Sprintf("%d %s", h.total, bar),
			fmt.Sprintf("%.1f%%", percent),
			fmt.Sprintf("%d ms", h.avgMs),
			fmt.Sprintf("%d", h.fail),
			fmt.Sprintf("%d", h.slow),
			status,
		})
	}
	table.Render()

	// 2. 高峰时段汇总
	fmt.Printf("\n🔥 高峰时段识别 (请求量 >= 平均值的2倍)\n")
	if len(peakHours) > 0 {
		fmt.Printf("   高峰时段: ")
		for i, h := range peakHours {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%02d:00", h)
		}
		fmt.Printf("\n")

		if peakCount > 0 {
			peakAvgMs = peakAvgMs / peakCount
		}
		if offPeakCount > 0 {
			offPeakAvgMs = offPeakAvgMs / offPeakCount
		}

		fmt.Printf("\n⚡ 高峰期 vs 非高峰期性能对比\n")
		compTable := tablewriter.NewWriter(os.Stdout)
		compTable.SetHeader([]string{"时段类型", "请求数", "平均耗时", "性能差异"})
		compTable.SetBorder(false)

		diff := ""
		if offPeakAvgMs > 0 {
			diffPercent := float64(peakAvgMs-offPeakAvgMs) * 100 / float64(offPeakAvgMs)
			if diffPercent > 0 {
				diff = fmt.Sprintf("+%.1f%% (高峰期更慢)", diffPercent)
			} else {
				diff = fmt.Sprintf("%.1f%%", diffPercent)
			}
		}

		compTable.Append([]string{"高峰期", fmt.Sprintf("%d", peakCount), fmt.Sprintf("%d ms", peakAvgMs), diff})
		compTable.Append([]string{"非高峰期", fmt.Sprintf("%d", offPeakCount), fmt.Sprintf("%d ms", offPeakAvgMs), "-"})
		compTable.Render()
	} else {
		fmt.Printf("   未检测到明显高峰时段（请求量分布较均匀）\n")
	}

	// 3. 按分钟粒度找出最繁忙的时刻
	fmt.Printf("\n🎯 最繁忙时刻 Top %d (按分钟)\n", min(a.topN, 10))
	query = fmt.Sprintf(`
		SELECT
			FROM_UNIXTIME(FLOOR(start_time/60000)*60, '%%H:%%i') as minute,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY minute
		ORDER BY total DESC
		LIMIT %d
	`, where, min(a.topN, 10))

	rows2, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows2.Close()

	minTable := tablewriter.NewWriter(os.Stdout)
	minTable.SetHeader([]string{"时刻", "请求数", "平均耗时"})
	minTable.SetBorder(false)

	for rows2.Next() {
		var minute string
		var total, avgMs int
		rows2.Scan(&minute, &total, &avgMs)
		minTable.Append([]string{minute, fmt.Sprintf("%d", total), fmt.Sprintf("%d ms", avgMs)})
	}
	minTable.Render()
}

// DeviceReport 设备性能分析报告
func (a *Analyzer) DeviceReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            设备性能分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 全局设备性能排名（最慢的设备）
	fmt.Printf("\n🐢 最慢设备 Top %d\n", a.topN)
	query := fmt.Sprintf(`
		SELECT
			device_sn,
			company_uuid,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s AND device_sn != ''
		GROUP BY device_sn, company_uuid
		HAVING total >= 10
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
	table.SetHeader([]string{"设备SN", "门店UUID", "请求数", "平均耗时", "最大耗时", "失败", "慢请求"})
	table.SetBorder(false)

	for rows.Next() {
		var deviceSn string
		var companyUuid uint64
		var total, avgMs, maxMs, fail, slow int
		rows.Scan(&deviceSn, &companyUuid, &total, &avgMs, &maxMs, &fail, &slow)
		table.Append([]string{
			truncate(deviceSn, 20),
			fmt.Sprintf("%d", companyUuid),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()

	// 2. 同门店设备性能对比（找出异常设备）
	fmt.Printf("\n🔍 门店内设备性能异常检测\n")
	fmt.Println("   (找出同一门店内性能明显差于其他设备的设备)")

	// 先获取每个门店的平均性能
	query = fmt.Sprintf(`
		SELECT
			company_uuid,
			device_sn,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms
		FROM ttpos_order_operation_duration
		WHERE %s AND device_sn != ''
		GROUP BY company_uuid, device_sn
		HAVING total >= 5
	`, where)

	rows2, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows2.Close()

	// 按门店聚合
	type deviceStat struct {
		deviceSn string
		total    int
		avgMs    int
	}
	companyDevices := make(map[uint64][]deviceStat)

	for rows2.Next() {
		var companyUuid uint64
		var ds deviceStat
		rows2.Scan(&companyUuid, &ds.deviceSn, &ds.total, &ds.avgMs)
		companyDevices[companyUuid] = append(companyDevices[companyUuid], ds)
	}

	// 找出异常设备（性能比同门店平均值差50%以上）
	type anomalyDevice struct {
		companyUuid  uint64
		deviceSn     string
		deviceAvgMs  int
		companyAvgMs int
		diffPercent  float64
		total        int
	}
	var anomalies []anomalyDevice

	for companyUuid, devices := range companyDevices {
		if len(devices) < 2 {
			continue // 只有一个设备的门店跳过
		}

		// 计算门店平均耗时
		var totalMs, totalCount int
		for _, d := range devices {
			totalMs += d.avgMs * d.total
			totalCount += d.total
		}
		companyAvgMs := totalMs / totalCount

		// 找出异常设备
		for _, d := range devices {
			if companyAvgMs > 0 {
				diffPercent := float64(d.avgMs-companyAvgMs) * 100 / float64(companyAvgMs)
				if diffPercent > 50 { // 比平均值慢50%以上
					anomalies = append(anomalies, anomalyDevice{
						companyUuid:  companyUuid,
						deviceSn:     d.deviceSn,
						deviceAvgMs:  d.avgMs,
						companyAvgMs: companyAvgMs,
						diffPercent:  diffPercent,
						total:        d.total,
					})
				}
			}
		}
	}

	if len(anomalies) == 0 {
		fmt.Println("\n   ✅ 未检测到明显异常的设备")
	} else {
		// 按差异百分比排序
		sort.Slice(anomalies, func(i, j int) bool {
			return anomalies[i].diffPercent > anomalies[j].diffPercent
		})

		anomalyTable := tablewriter.NewWriter(os.Stdout)
		anomalyTable.SetHeader([]string{"设备SN", "门店UUID", "设备耗时", "门店平均", "差异", "请求数", "建议"})
		anomalyTable.SetBorder(false)

		for i, ad := range anomalies {
			if i >= a.topN {
				break
			}
			suggestion := "检查设备"
			if ad.diffPercent > 100 {
				suggestion = "⚠️ 建议更换"
			} else if ad.diffPercent > 80 {
				suggestion = "需要维护"
			}

			anomalyTable.Append([]string{
				truncate(ad.deviceSn, 20),
				fmt.Sprintf("%d", ad.companyUuid),
				fmt.Sprintf("%d ms", ad.deviceAvgMs),
				fmt.Sprintf("%d ms", ad.companyAvgMs),
				fmt.Sprintf("+%.1f%%", ad.diffPercent),
				fmt.Sprintf("%d", ad.total),
				suggestion,
			})
		}
		anomalyTable.Render()
	}

	// 3. 指定门店的设备对比（如果指定了company参数）
	if a.companyUuid > 0 {
		fmt.Printf("\n📊 门店 %d 设备性能对比\n", a.companyUuid)

		query = fmt.Sprintf(`
			SELECT
				device_sn,
				source,
				COUNT(*) as total,
				ROUND(AVG(duration_ms)) as avg_ms,
				MIN(duration_ms) as min_ms,
				MAX(duration_ms) as max_ms,
				SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail
			FROM ttpos_order_operation_duration
			WHERE %s AND device_sn != ''
			GROUP BY device_sn, source
			ORDER BY avg_ms DESC
		`, where)

		rows3, err := a.db.Query(query, args...)
		if err != nil {
			return
		}
		defer rows3.Close()

		compTable := tablewriter.NewWriter(os.Stdout)
		compTable.SetHeader([]string{"设备SN", "终端", "请求数", "平均", "最小", "最大", "失败"})
		compTable.SetBorder(false)

		for rows3.Next() {
			var deviceSn, source string
			var total, avgMs, minMs, maxMs, fail int
			rows3.Scan(&deviceSn, &source, &total, &avgMs, &minMs, &maxMs, &fail)
			compTable.Append([]string{
				truncate(deviceSn, 20),
				source,
				fmt.Sprintf("%d", total),
				fmt.Sprintf("%d ms", avgMs),
				fmt.Sprintf("%d ms", minMs),
				fmt.Sprintf("%d ms", maxMs),
				fmt.Sprintf("%d", fail),
			})
		}
		compTable.Render()
	}
}

// StaffReport 员工效率分析报告
func (a *Analyzer) StaffReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            员工效率分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 员工操作量排名
	fmt.Printf("\n📊 员工操作量排名 Top %d\n", a.topN)
	query := fmt.Sprintf(`
		SELECT
			staff_uuid,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			MIN(duration_ms) as min_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s AND staff_uuid > 0
		GROUP BY staff_uuid
		ORDER BY total DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"员工UUID", "操作数", "平均耗时", "最小", "最大", "失败", "慢操作"})
	table.SetBorder(false)

	for rows.Next() {
		var staffUuid uint64
		var total, avgMs, minMs, maxMs, fail, slow int
		rows.Scan(&staffUuid, &total, &avgMs, &minMs, &maxMs, &fail, &slow)
		table.Append([]string{
			fmt.Sprintf("%d", staffUuid),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", minMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%d", slow),
		})
	}
	table.Render()

	// 2. 效率最低的员工（平均耗时最长）
	fmt.Printf("\n🐢 平均耗时最长的员工 Top %d\n", a.topN)
	query = fmt.Sprintf(`
		SELECT
			staff_uuid,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail
		FROM ttpos_order_operation_duration
		WHERE %s AND staff_uuid > 0
		GROUP BY staff_uuid
		HAVING total >= 10
		ORDER BY avg_ms DESC
		LIMIT %d
	`, where, a.topN)

	rows2, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows2.Close()

	table2 := tablewriter.NewWriter(os.Stdout)
	table2.SetHeader([]string{"员工UUID", "操作数", "平均耗时", "失败数", "错误率"})
	table2.SetBorder(false)

	for rows2.Next() {
		var staffUuid uint64
		var total, avgMs, fail int
		rows2.Scan(&staffUuid, &total, &avgMs, &fail)
		errorRate := float64(fail) * 100 / float64(total)
		table2.Append([]string{
			fmt.Sprintf("%d", staffUuid),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
		})
	}
	table2.Render()

	// 3. 错误率最高的员工
	fmt.Printf("\n❌ 错误率最高的员工 Top %d\n", a.topN)
	query = fmt.Sprintf(`
		SELECT
			staff_uuid,
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail
		FROM ttpos_order_operation_duration
		WHERE %s AND staff_uuid > 0
		GROUP BY staff_uuid
		HAVING total >= 10 AND fail > 0
		ORDER BY (fail/total) DESC, fail DESC
		LIMIT %d
	`, where, a.topN)

	rows3, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows3.Close()

	table3 := tablewriter.NewWriter(os.Stdout)
	table3.SetHeader([]string{"员工UUID", "操作数", "失败数", "错误率"})
	table3.SetBorder(false)

	for rows3.Next() {
		var staffUuid uint64
		var total, fail int
		rows3.Scan(&staffUuid, &total, &fail)
		errorRate := float64(fail) * 100 / float64(total)
		table3.Append([]string{
			fmt.Sprintf("%d", staffUuid),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
		})
	}
	table3.Render()
}

// InstanceReport 服务实例负载分析报告
func (a *Analyzer) InstanceReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                          服务实例负载分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 实例负载概览
	fmt.Printf("\n📊 实例负载概览\n")
	query := fmt.Sprintf(`
		SELECT
			instance_id,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			MIN(duration_ms) as min_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s AND instance_id != ''
		GROUP BY instance_id
		ORDER BY total DESC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	type instanceStat struct {
		instanceId string
		total      int
		avgMs      int
		minMs      int
		maxMs      int
		fail       int
		slow       int
	}
	var stats []instanceStat
	var totalRequests int

	for rows.Next() {
		var s instanceStat
		rows.Scan(&s.instanceId, &s.total, &s.avgMs, &s.minMs, &s.maxMs, &s.fail, &s.slow)
		stats = append(stats, s)
		totalRequests += s.total
	}

	if len(stats) == 0 {
		fmt.Println("暂无实例数据")
		return
	}

	avgPerInstance := totalRequests / len(stats)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"实例ID", "请求数", "负载", "平均耗时", "最大耗时", "失败", "慢请求", "状态"})
	table.SetBorder(false)

	for _, s := range stats {
		loadPercent := float64(s.total) * 100 / float64(totalRequests)
		status := ""
		if s.total > avgPerInstance*2 {
			status = "⚠️ 过载"
		} else if s.total < avgPerInstance/2 {
			status = "💤 空闲"
		} else {
			status = "✅ 正常"
		}

		table.Append([]string{
			truncate(s.instanceId, 25),
			fmt.Sprintf("%d", s.total),
			fmt.Sprintf("%.1f%%", loadPercent),
			fmt.Sprintf("%d ms", s.avgMs),
			fmt.Sprintf("%d ms", s.maxMs),
			fmt.Sprintf("%d", s.fail),
			fmt.Sprintf("%d", s.slow),
			status,
		})
	}
	table.Render()

	// 2. 负载均衡分析
	fmt.Printf("\n⚖️ 负载均衡分析\n")
	fmt.Printf("   实例数量:     %d\n", len(stats))
	fmt.Printf("   总请求数:     %d\n", totalRequests)
	fmt.Printf("   平均每实例:   %d\n", avgPerInstance)

	// 计算标准差
	var variance float64
	for _, s := range stats {
		diff := float64(s.total - avgPerInstance)
		variance += diff * diff
	}
	stdDev := int(sqrt(variance / float64(len(stats))))
	fmt.Printf("   负载标准差:   %d (越小越均衡)\n", stdDev)

	if stdDev > avgPerInstance/2 {
		fmt.Printf("   ⚠️ 负载分布不均衡，建议检查负载均衡配置\n")
	} else {
		fmt.Printf("   ✅ 负载分布较均衡\n")
	}
}

// sqrt 简单的平方根计算
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// BillTraceReport 账单操作追踪报告
func (a *Analyzer) BillTraceReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            账单操作追踪报告")
	fmt.Println(strings.Repeat("=", 80))

	if a.billUuid == 0 {
		fmt.Println("\n❌ 错误: 需要指定 -bill 参数")
		fmt.Println("   用法: ./duration-analyzer -dsn '$DSN' -report bill-trace -bill 123456789")
		return
	}

	fmt.Printf("账单UUID: %d\n", a.billUuid)
	fmt.Println(strings.Repeat("-", 80))

	// 查询该账单的所有操作
	query := `
		SELECT
			action, source, staff_uuid, device_sn,
			duration_ms, status, error_msg,
			FROM_UNIXTIME(start_time/1000) as start_at,
			time_nodes
		FROM ttpos_order_operation_duration
		WHERE sale_bill_uuid = ?
		ORDER BY start_time ASC
	`

	rows, err := a.db.Query(query, a.billUuid)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("\n📋 操作时间线\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "操作", "来源", "员工", "耗时", "状态", "错误"})
	table.SetBorder(false)

	var totalDuration int
	var operationCount int

	for rows.Next() {
		var action, source, deviceSn, startAt, timeNodes string
		var staffUuid uint64
		var durationMs, status int
		var errorMsg sql.NullString
		rows.Scan(&action, &source, &staffUuid, &deviceSn, &durationMs, &status, &errorMsg, &startAt, &timeNodes)

		statusStr := "✓"
		if status == 0 {
			statusStr = "✗"
		}

		errStr := ""
		if errorMsg.Valid && errorMsg.String != "" {
			errStr = truncate(errorMsg.String, 30)
		}

		table.Append([]string{
			startAt,
			action,
			source,
			fmt.Sprintf("%d", staffUuid),
			fmt.Sprintf("%d ms", durationMs),
			statusStr,
			errStr,
		})

		totalDuration += durationMs
		operationCount++
	}
	table.Render()

	if operationCount > 0 {
		fmt.Printf("\n📊 汇总\n")
		fmt.Printf("   操作次数:   %d\n", operationCount)
		fmt.Printf("   总耗时:     %d ms\n", totalDuration)
		fmt.Printf("   平均耗时:   %d ms\n", totalDuration/operationCount)
	} else {
		fmt.Println("\n未找到该账单的操作记录")
	}
}

// CompareReport 时间段对比报告
func (a *Analyzer) CompareReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            时间段对比报告")
	fmt.Println(strings.Repeat("=", 80))

	if a.compareStartMs == 0 || a.compareEndMs == 0 {
		fmt.Println("\n❌ 错误: 需要指定 -compare-start 和 -compare-end 参数")
		fmt.Println("   用法: ./duration-analyzer -dsn '$DSN' -report compare \\")
		fmt.Println("         -start 2026-02-05 -end 2026-02-06 \\")
		fmt.Println("         -compare-start 2026-02-04 -compare-end 2026-02-05")
		return
	}

	fmt.Printf("时间段A: %s ~ %s\n", startTime, endTime)
	fmt.Printf("时间段B: %s ~ %s\n", compareStart, compareEnd)
	fmt.Println(strings.Repeat("-", 80))

	// 获取时间段A的统计
	statsA := a.getTimeRangeStats(a.startTimeMs, a.endTimeMs)
	// 获取时间段B的统计
	statsB := a.getTimeRangeStats(a.compareStartMs, a.compareEndMs)

	fmt.Printf("\n📊 整体对比\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "时间段A", "时间段B", "变化"})
	table.SetBorder(false)

	addCompareRow := func(name string, a, b int64) {
		diff := ""
		if b > 0 {
			diffPercent := float64(a-b) * 100 / float64(b)
			if diffPercent > 0 {
				diff = fmt.Sprintf("+%.1f%% ↑", diffPercent)
			} else if diffPercent < 0 {
				diff = fmt.Sprintf("%.1f%% ↓", diffPercent)
			} else {
				diff = "0%"
			}
		}
		table.Append([]string{name, fmt.Sprintf("%d", a), fmt.Sprintf("%d", b), diff})
	}

	addCompareRow("请求数", statsA.total, statsB.total)
	addCompareRow("失败数", statsA.fail, statsB.fail)
	addCompareRow("平均耗时(ms)", statsA.avgMs, statsB.avgMs)
	addCompareRow("最大耗时(ms)", statsA.maxMs, statsB.maxMs)
	addCompareRow("慢请求数", statsA.slow, statsB.slow)

	table.Append([]string{
		"错误率",
		fmt.Sprintf("%.2f%%", statsA.errorRate),
		fmt.Sprintf("%.2f%%", statsB.errorRate),
		fmt.Sprintf("%.2f%%", statsA.errorRate-statsB.errorRate),
	})

	table.Render()

	// 性能变化评估
	fmt.Printf("\n📈 变化评估\n")
	if statsA.avgMs > statsB.avgMs {
		degradation := float64(statsA.avgMs-statsB.avgMs) * 100 / float64(statsB.avgMs)
		if degradation > 20 {
			fmt.Printf("   ⚠️ 性能明显下降 (+%.1f%%)\n", degradation)
		} else {
			fmt.Printf("   📉 性能略有下降 (+%.1f%%)\n", degradation)
		}
	} else if statsA.avgMs < statsB.avgMs {
		improvement := float64(statsB.avgMs-statsA.avgMs) * 100 / float64(statsB.avgMs)
		fmt.Printf("   ✅ 性能提升 (-%.1f%%)\n", improvement)
	} else {
		fmt.Printf("   ➖ 性能持平\n")
	}
}

type timeRangeStats struct {
	total     int64
	fail      int64
	avgMs     int64
	maxMs     int64
	slow      int64
	errorRate float64
}

func (a *Analyzer) getTimeRangeStats(startMs, endMs int64) timeRangeStats {
	conditions := []string{"start_time >= ? AND start_time < ?"}
	args := []any{startMs, endMs}

	if a.companyUuid > 0 {
		conditions = append(conditions, "company_uuid = ?")
		args = append(args, a.companyUuid)
	}
	if a.action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, a.action)
	}

	where := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail,
			ROUND(AVG(duration_ms)) as avg_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN duration_ms > 3000 THEN 1 ELSE 0 END) as slow
		FROM ttpos_order_operation_duration
		WHERE %s
	`, where)

	var stats timeRangeStats
	a.db.QueryRow(query, args...).Scan(&stats.total, &stats.fail, &stats.avgMs, &stats.maxMs, &stats.slow)

	if stats.total > 0 {
		stats.errorRate = float64(stats.fail) * 100 / float64(stats.total)
	}

	return stats
}

// DistributionReport 耗时分布报告
func (a *Analyzer) DistributionReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            耗时分布直方图")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 定义分布区间
	buckets := []struct {
		name   string
		minMs  int
		maxMs  int
	}{
		{"0-50ms", 0, 50},
		{"50-100ms", 50, 100},
		{"100-200ms", 100, 200},
		{"200-500ms", 200, 500},
		{"500ms-1s", 500, 1000},
		{"1s-2s", 1000, 2000},
		{"2s-3s", 2000, 3000},
		{"3s-5s", 3000, 5000},
		{"5s-10s", 5000, 10000},
		{">10s", 10000, 999999999},
	}

	// 查询每个区间的数量
	var results []struct {
		name    string
		count   int
		percent float64
	}
	var total int

	for _, b := range buckets {
		query := fmt.Sprintf(`
			SELECT COUNT(*) FROM ttpos_order_operation_duration
			WHERE %s AND duration_ms >= ? AND duration_ms < ?
		`, where)
		fullArgs := append(args, b.minMs, b.maxMs)

		var count int
		a.db.QueryRow(query, fullArgs...).Scan(&count)
		total += count
		results = append(results, struct {
			name    string
			count   int
			percent float64
		}{name: b.name, count: count})
	}

	// 计算百分比
	maxCount := 0
	for i := range results {
		if total > 0 {
			results[i].percent = float64(results[i].count) * 100 / float64(total)
		}
		if results[i].count > maxCount {
			maxCount = results[i].count
		}
	}

	fmt.Printf("\n📊 耗时分布 (共 %d 条记录)\n\n", total)

	// 绘制直方图
	for _, r := range results {
		barLen := 0
		if maxCount > 0 {
			barLen = r.count * 40 / maxCount
		}
		bar := strings.Repeat("█", barLen)
		fmt.Printf("  %10s |%-40s| %6d (%5.1f%%)\n", r.name, bar, r.count, r.percent)
	}

	// 分析结果
	fmt.Printf("\n📈 分布分析\n")
	fast := 0    // <200ms
	normal := 0  // 200ms-1s
	slow := 0    // 1s-3s
	verySlow := 0 // >3s

	for i, r := range results {
		switch {
		case i <= 2:
			fast += r.count
		case i <= 4:
			normal += r.count
		case i <= 6:
			slow += r.count
		default:
			verySlow += r.count
		}
	}

	if total > 0 {
		fmt.Printf("   快速 (<200ms):    %d (%.1f%%)\n", fast, float64(fast)*100/float64(total))
		fmt.Printf("   正常 (200ms-1s):  %d (%.1f%%)\n", normal, float64(normal)*100/float64(total))
		fmt.Printf("   较慢 (1s-3s):     %d (%.1f%%)\n", slow, float64(slow)*100/float64(total))
		fmt.Printf("   很慢 (>3s):       %d (%.1f%%)\n", verySlow, float64(verySlow)*100/float64(total))
	}
}

// AnomalyReport 异常检测报告
func (a *Analyzer) AnomalyReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            异常检测报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 计算 P99 阈值
	query := fmt.Sprintf(`
		SELECT duration_ms FROM ttpos_order_operation_duration
		WHERE %s ORDER BY duration_ms ASC
	`, where)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
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
		fmt.Println("暂无数据")
		return
	}

	n := len(durations)
	p99 := durations[percentileIndex(n, 99)]
	p95 := durations[percentileIndex(n, 95)]

	fmt.Printf("\n📊 基准值\n")
	fmt.Printf("   P95: %d ms\n", p95)
	fmt.Printf("   P99: %d ms\n", p99)
	fmt.Printf("   异常阈值 (P99): %d ms\n", p99)

	// 2. 查询超过 P99 的请求
	fmt.Printf("\n🚨 异常请求 (>P99) Top %d\n", a.topN)
	query = fmt.Sprintf(`
		SELECT action, source, company_uuid, duration_ms, status, error_msg,
		       FROM_UNIXTIME(start_time/1000) as start_at
		FROM ttpos_order_operation_duration
		WHERE %s AND duration_ms > ?
		ORDER BY duration_ms DESC
		LIMIT %d
	`, where, a.topN)

	fullArgs := append(args, p99)
	rows2, err := a.db.Query(query, fullArgs...)
	if err != nil {
		return
	}
	defer rows2.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "操作", "来源", "门店", "耗时", "状态", "错误"})
	table.SetBorder(false)

	for rows2.Next() {
		var action, source, startAt string
		var companyUuid uint64
		var durationMs, status int
		var errorMsg sql.NullString
		rows2.Scan(&action, &source, &companyUuid, &durationMs, &status, &errorMsg, &startAt)

		statusStr := "✓"
		if status == 0 {
			statusStr = "✗"
		}

		errStr := ""
		if errorMsg.Valid {
			errStr = truncate(errorMsg.String, 25)
		}

		table.Append([]string{
			startAt,
			action,
			source,
			fmt.Sprintf("%d", companyUuid),
			fmt.Sprintf("%d ms", durationMs),
			statusStr,
			errStr,
		})
	}
	table.Render()

	// 3. 异常接口统计
	fmt.Printf("\n📋 异常频发接口 (超过P99次数最多)\n")
	query = fmt.Sprintf(`
		SELECT action, COUNT(*) as cnt, ROUND(AVG(duration_ms)) as avg_ms
		FROM ttpos_order_operation_duration
		WHERE %s AND duration_ms > ?
		GROUP BY action
		ORDER BY cnt DESC
		LIMIT 10
	`, where)

	rows3, err := a.db.Query(query, fullArgs...)
	if err != nil {
		return
	}
	defer rows3.Close()

	table2 := tablewriter.NewWriter(os.Stdout)
	table2.SetHeader([]string{"操作", "异常次数", "平均耗时"})
	table2.SetBorder(false)

	for rows3.Next() {
		var action string
		var cnt, avgMs int
		rows3.Scan(&action, &cnt, &avgMs)
		table2.Append([]string{action, fmt.Sprintf("%d", cnt), fmt.Sprintf("%d ms", avgMs)})
	}
	table2.Render()
}

// PathReport 请求路径分析报告
func (a *Analyzer) PathReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                            请求路径分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 请求路径性能排名
	fmt.Printf("\n🐢 最慢请求路径 Top %d\n", a.topN)
	query := fmt.Sprintf(`
		SELECT
			request_path,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms,
			MAX(duration_ms) as max_ms,
			SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as fail
		FROM ttpos_order_operation_duration
		WHERE %s AND request_path != ''
		GROUP BY request_path
		HAVING total >= 5
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
	table.SetHeader([]string{"请求路径", "请求数", "平均耗时", "最大耗时", "失败", "错误率"})
	table.SetBorder(false)

	for rows.Next() {
		var path string
		var total, avgMs, maxMs, fail int
		rows.Scan(&path, &total, &avgMs, &maxMs, &fail)
		errorRate := float64(fail) * 100 / float64(total)
		table.Append([]string{
			truncate(path, 40),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d ms", avgMs),
			fmt.Sprintf("%d ms", maxMs),
			fmt.Sprintf("%d", fail),
			fmt.Sprintf("%.2f%%", errorRate),
		})
	}
	table.Render()

	// 2. 请求量最高的路径
	fmt.Printf("\n📊 请求量最高路径 Top %d\n", a.topN)
	query = fmt.Sprintf(`
		SELECT
			request_path,
			COUNT(*) as total,
			ROUND(AVG(duration_ms)) as avg_ms
		FROM ttpos_order_operation_duration
		WHERE %s AND request_path != ''
		GROUP BY request_path
		ORDER BY total DESC
		LIMIT %d
	`, where, a.topN)

	rows2, err := a.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows2.Close()

	table2 := tablewriter.NewWriter(os.Stdout)
	table2.SetHeader([]string{"请求路径", "请求数", "平均耗时"})
	table2.SetBorder(false)

	for rows2.Next() {
		var path string
		var total, avgMs int
		rows2.Scan(&path, &total, &avgMs)
		table2.Append([]string{
			truncate(path, 50),
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d ms", avgMs),
		})
	}
	table2.Render()
}

// ConcurrencyReport 并发分析报告
func (a *Analyzer) ConcurrencyReport() {
	where, args := a.buildWhereClause()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                              并发分析报告")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("时间范围: %s ~ %s\n", startTime, endTime)
	fmt.Println(strings.Repeat("-", 80))

	// 1. 按秒统计并发量
	fmt.Printf("\n📊 每秒请求量分布\n")
	query := fmt.Sprintf(`
		SELECT
			FROM_UNIXTIME(FLOOR(start_time/1000), '%%Y-%%m-%%d %%H:%%i:%%s') as sec,
			COUNT(*) as qps
		FROM ttpos_order_operation_duration
		WHERE %s
		GROUP BY sec
		ORDER BY qps DESC
		LIMIT %d
	`, where, a.topN)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	type secStat struct {
		sec string
		qps int
	}
	var stats []secStat
	var maxQps int

	for rows.Next() {
		var s secStat
		rows.Scan(&s.sec, &s.qps)
		stats = append(stats, s)
		if s.qps > maxQps {
			maxQps = s.qps
		}
	}

	fmt.Printf("\n🔥 峰值 QPS Top %d\n", min(len(stats), a.topN))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"时间", "QPS", "可视化"})
	table.SetBorder(false)

	for _, s := range stats {
		barLen := 0
		if maxQps > 0 {
			barLen = s.qps * 30 / maxQps
		}
		bar := strings.Repeat("█", barLen)
		table.Append([]string{s.sec, fmt.Sprintf("%d", s.qps), bar})
	}
	table.Render()

	// 2. 并发统计汇总
	query = fmt.Sprintf(`
		SELECT
			COUNT(*) as total_seconds,
			AVG(qps) as avg_qps,
			MAX(qps) as max_qps,
			MIN(qps) as min_qps
		FROM (
			SELECT COUNT(*) as qps
			FROM ttpos_order_operation_duration
			WHERE %s
			GROUP BY FLOOR(start_time/1000)
		) t
	`, where)

	var totalSeconds int
	var avgQps, peakQps, minQps float64
	a.db.QueryRow(query, args...).Scan(&totalSeconds, &avgQps, &peakQps, &minQps)

	fmt.Printf("\n📈 并发统计\n")
	fmt.Printf("   统计秒数:   %d 秒\n", totalSeconds)
	fmt.Printf("   平均 QPS:   %.1f\n", avgQps)
	fmt.Printf("   峰值 QPS:   %.0f\n", peakQps)
	fmt.Printf("   最低 QPS:   %.0f\n", minQps)

	// 3. 高并发时段的性能影响
	fmt.Printf("\n⚡ 高并发对性能的影响\n")
	// 找出高并发时段 (QPS > 平均值*2)
	query = fmt.Sprintf(`
		SELECT
			AVG(duration_ms) as high_conc_avg,
			(SELECT AVG(duration_ms) FROM ttpos_order_operation_duration WHERE %s) as overall_avg
		FROM ttpos_order_operation_duration
		WHERE %s AND FLOOR(start_time/1000) IN (
			SELECT sec FROM (
				SELECT FLOOR(start_time/1000) as sec, COUNT(*) as qps
				FROM ttpos_order_operation_duration
				WHERE %s
				GROUP BY sec
				HAVING qps > ?
			) high_qps
		)
	`, where, where, where)

	fullArgs := append(args, args...)
	fullArgs = append(fullArgs, args...)
	fullArgs = append(fullArgs, int(avgQps*2))

	var highConcAvg, overallAvg sql.NullFloat64
	a.db.QueryRow(query, fullArgs...).Scan(&highConcAvg, &overallAvg)

	if highConcAvg.Valid && overallAvg.Valid && overallAvg.Float64 > 0 {
		diff := (highConcAvg.Float64 - overallAvg.Float64) * 100 / overallAvg.Float64
		fmt.Printf("   整体平均耗时:     %.0f ms\n", overallAvg.Float64)
		fmt.Printf("   高并发时平均耗时: %.0f ms\n", highConcAvg.Float64)
		if diff > 0 {
			fmt.Printf("   性能下降:         +%.1f%%\n", diff)
		} else {
			fmt.Printf("   性能变化:         %.1f%%\n", diff)
		}
	}
}

