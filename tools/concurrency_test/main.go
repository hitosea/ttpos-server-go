package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ── JWT Claims（复用 gen_token 结构）──

type Assistant struct {
	DeviceId  string `json:"device_id"`
	StaffUuid uint64 `json:"staff_uuid"`
}

type Claims struct {
	Source         string    `json:"source"`
	CompanyUuid    uint64    `json:"company_uuid"`
	StaffUuid      uint64    `json:"staff_uuid"`
	MemberUuid     uint64    `json:"member_uuid"`
	DeviceUuid     uint64    `json:"device_uuid"`
	DeviceId       string    `json:"device_id"`
	Assistant      Assistant `json:"assistant"`
	IsRefreshToken bool      `json:"is_refresh_token"`
	Brand          string    `json:"brand"`
	jwt.RegisteredClaims
}

// ── 请求结果 ──

type Result struct {
	Index    int
	Duration time.Duration
	HTTPCode int
	APICode  int
	Error    string
}

// ── 配置 ──

type Config struct {
	Method     string
	Host       string
	Path       string
	Body       string
	Concurrent int
	Total      int
	Verbose    bool
	SlowMs     int

	// JWT
	Source     string
	Company   uint64
	Staff     uint64
	Device    string
	DeviceUuid uint64
	Secret    string
}

func main() {
	cfg := parseFlags()

	token, err := generateToken(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成 token 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== API 并发测试 ===\n")
	fmt.Printf("方法: %s\n", cfg.Method)
	fmt.Printf("路径: %s\n", cfg.Path)
	fmt.Printf("并发: %d | 总请求: %d\n", cfg.Concurrent, cfg.Total)
	if cfg.Body != "" {
		body := cfg.Body
		if len(body) > 80 {
			body = body[:80] + "..."
		}
		fmt.Printf("Body: %s\n", body)
	}
	fmt.Println()

	results := runBenchmark(cfg, token)
	printReport(results, cfg.SlowMs)
}

func parseFlags() Config {
	var cfg Config
	var bodyFile string

	flag.StringVar(&cfg.Method, "method", "POST", "HTTP 方法 (GET/POST/DELETE)")
	flag.StringVar(&cfg.Host, "host", "http://localhost:8080", "目标主机")
	flag.StringVar(&cfg.Path, "path", "", "API 路径 (必填)")
	flag.StringVar(&cfg.Body, "body", "", "请求体 JSON")
	flag.StringVar(&bodyFile, "body-file", "", "请求体 JSON 文件路径 (与 -body 互斥)")
	flag.IntVar(&cfg.Concurrent, "c", 5, "并发数")
	flag.IntVar(&cfg.Total, "n", 20, "总请求数")
	flag.BoolVar(&cfg.Verbose, "v", false, "详细输出")
	flag.IntVar(&cfg.SlowMs, "slow", 1000, "慢请求阈值 (ms)")

	flag.StringVar(&cfg.Source, "source", "cashier", "Token source")
	flag.Uint64Var(&cfg.Company, "company", 8267304538112000, "Company UUID")
	flag.Uint64Var(&cfg.Staff, "staff", 1724054105, "Staff UUID")
	flag.StringVar(&cfg.Device, "device", "claude-test-device", "Device ID")
	flag.Uint64Var(&cfg.DeviceUuid, "device-uuid", 0, "Device UUID")
	flag.StringVar(&cfg.Secret, "secret", "", "JWT secret (默认读 JWT_SECRET 环境变量)")

	flag.Parse()

	if cfg.Path == "" {
		fmt.Fprintln(os.Stderr, "错误: -path 参数必填")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "用法示例:")
		fmt.Fprintln(os.Stderr, `  ./concurrency_test -path "/api/v1/cashier/instant/order/cart/info" -method GET -c 5 -n 20`)
		fmt.Fprintln(os.Stderr, `  ./concurrency_test -path "/api/v1/cashier/instant/order/cart/product/num" -body '{"num":100,...}' -c 5 -n 20`)
		os.Exit(1)
	}

	if bodyFile != "" && cfg.Body != "" {
		fmt.Fprintln(os.Stderr, "错误: -body 和 -body-file 不能同时使用")
		os.Exit(1)
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取 body 文件失败: %v\n", err)
			os.Exit(1)
		}
		cfg.Body = string(data)
	}

	if cfg.Secret == "" {
		cfg.Secret = os.Getenv("JWT_SECRET")
	}
	if cfg.Secret == "" {
		cfg.Secret = "dkjhd00a08"
	}

	return cfg
}

func generateToken(cfg Config) (string, error) {
	claims := Claims{
		Source:      cfg.Source,
		CompanyUuid: cfg.Company,
		StaffUuid:   cfg.Staff,
		DeviceId:    cfg.Device,
		DeviceUuid:  cfg.DeviceUuid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func runBenchmark(cfg Config, token string) []Result {
	sem := make(chan struct{}, cfg.Concurrent)
	resultsCh := make(chan Result, cfg.Total)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 30 * time.Second}
	url := cfg.Host + cfg.Path

	for i := 0; i < cfg.Total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := sendRequest(client, cfg.Method, url, cfg.Body, token, idx)

			// 实时输出慢请求
			ms := r.Duration.Milliseconds()
			if ms > int64(cfg.SlowMs) {
				fmt.Printf("  [SLOW] #%d: %dms (HTTP:%d code:%d)\n", r.Index, ms, r.HTTPCode, r.APICode)
			} else if cfg.Verbose {
				fmt.Printf("  #%d: %dms (HTTP:%d code:%d)\n", r.Index, ms, r.HTTPCode, r.APICode)
			}

			resultsCh <- r
		}(i)
	}

	wg.Wait()
	close(resultsCh)

	results := make([]Result, 0, cfg.Total)
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func sendRequest(client *http.Client, method, url, body, token string, idx int) Result {
	r := Result{Index: idx}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := client.Do(req)
	r.Duration = time.Since(start)

	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer resp.Body.Close()

	r.HTTPCode = resp.StatusCode

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Error = err.Error()
		return r
	}

	var apiResp struct {
		Code int `json:"code"`
	}
	if json.Unmarshal(respBody, &apiResp) == nil {
		r.APICode = apiResp.Code
	} else {
		r.APICode = -1
	}

	return r
}

func printReport(results []Result, slowMs int) {
	if len(results) == 0 {
		fmt.Println("无结果")
		return
	}

	fmt.Printf("\n=== 测试结果 ===\n")

	// 统计成功/失败
	success := 0
	var errResults []Result
	for _, r := range results {
		if r.HTTPCode == 200 && r.APICode == 0 && r.Error == "" {
			success++
		} else if r.Error != "" {
			errResults = append(errResults, r)
		}
	}
	failed := len(results) - success

	// 按耗时排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Duration < results[j].Duration
	})

	n := len(results)
	minMs := results[0].Duration.Milliseconds()
	maxMs := results[n-1].Duration.Milliseconds()

	var totalMs int64
	for _, r := range results {
		totalMs += r.Duration.Milliseconds()
	}
	avgMs := totalMs / int64(n)

	p50 := results[percentileIndex(n, 50)].Duration.Milliseconds()
	p95 := results[percentileIndex(n, 95)].Duration.Milliseconds()
	p99 := results[percentileIndex(n, 99)].Duration.Milliseconds()

	slow1s := 0
	slow3s := 0
	for _, r := range results {
		ms := r.Duration.Milliseconds()
		if ms > 1000 {
			slow1s++
		}
		if ms > 3000 {
			slow3s++
		}
	}

	fmt.Printf("请求数: %d | 成功: %d | 失败: %d\n", n, success, failed)
	fmt.Printf("Min: %dms | Max: %dms | Avg: %dms\n", minMs, maxMs, avgMs)
	fmt.Printf("P50: %dms | P95: %dms | P99: %dms\n", p50, p95, p99)
	fmt.Printf(">1s: %d | >3s: %d\n", slow1s, slow3s)

	// Top 10 最慢
	fmt.Printf("\n--- Top 10 最慢 ---\n")
	start := n - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < n; i++ {
		r := results[i]
		line := fmt.Sprintf("  #%d: %dms (HTTP:%d code:%d)", r.Index, r.Duration.Milliseconds(), r.HTTPCode, r.APICode)
		if r.Error != "" {
			line += fmt.Sprintf(" err:%s", r.Error)
		}
		fmt.Println(line)
	}

	// 显示错误详情
	if len(errResults) > 0 {
		fmt.Printf("\n--- 错误详情 ---\n")
		for _, r := range errResults {
			fmt.Printf("  #%d: %s\n", r.Index, r.Error)
		}
	}

	fmt.Println()
}

func percentileIndex(n, p int) int {
	idx := (n*p + 99) / 100
	if idx >= n {
		idx = n - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}
