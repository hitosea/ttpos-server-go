package command

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// 命令行参数
var (
	batchCompanies string // 逗号分隔的 company_uuid 列表
	batchWorkers   int    // 并发数
	batchVerbose   bool   // 是否输出详细日志
)

// SyncResult 同步结果
type SyncResult struct {
	CompanyUuid uint64
	CompanyName string
	Success     bool
	Skipped     bool // 是否因锁冲突被跳过
	Error       error
	Duration    time.Duration
}

func init() {
	syncErpDataBatchCmd.Flags().StringVarP(&batchCompanies, "companies", "c", "", "门店 UUID 列表，逗号分隔（必填）")
	syncErpDataBatchCmd.Flags().IntVarP(&batchWorkers, "workers", "w", 5, "最大并发数（默认 5）")
	syncErpDataBatchCmd.Flags().BoolVarP(&batchVerbose, "verbose", "v", false, "输出详细日志到控制台")
	_ = syncErpDataBatchCmd.MarkFlagRequired("companies")
	rootCommand.AddCommand(syncErpDataBatchCmd)
}

// syncErpDataBatchCmd 批量同步 ERP 数据命令
var syncErpDataBatchCmd = &cobra.Command{
	Use:   "sync-erp-data-batch",
	Short: "批量同步多门店 ERP 数据",
	Long:  `批量同步多门店 ERP 数据，支持并发控制。使用示例：./main sync-erp-data-batch -c 1234,5678,9012 -w 3`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化 Redis 分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}
		// 根据 verbose 参数决定是否输出详细日志到控制台
		if batchVerbose {
			setupConsoleLogger()
		}
		// 门店级别的进度通过 fmt.Printf 输出到控制台

		// 初始化 id 生成器
		utils.InitIdGenerator()

		// 初始化服务发现
		cloud.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		startTime := time.Now()

		// 解析门店列表
		companyUuids, err := parseBatchSyncCompanyUuids(batchCompanies)
		if err != nil {
			fmt.Printf("%s错误: %v%s\n", redColor, err, resetColor)
			return
		}

		if len(companyUuids) == 0 {
			fmt.Printf("%s错误: 未指定门店 UUID%s\n", redColor, resetColor)
			return
		}

		fmt.Printf("%s========== 批量同步 ERP 数据 ==========%s\n", blueColor, resetColor)
		fmt.Printf("门店数量: %d, 并发数: %d\n", len(companyUuids), batchWorkers)
		fmt.Printf("门店列表: %v\n\n", companyUuids)

		// 初始化 DBManager（统一管理所有数据库连接）
		dbm := database.GetDBManager(config.Database)

		// 验证所有门店
		fmt.Printf("%s[1/3] 验证门店有效性...%s\n", blueColor, resetColor)
		invalidUuids := validateCompanyUuids(companyUuids, dbm)
		if len(invalidUuids) > 0 {
			fmt.Printf("%s错误: 发现无效的 company_uuid，终止执行%s\n", redColor, resetColor)
			fmt.Printf("无效列表:\n")
			for _, item := range invalidUuids {
				fmt.Printf("  - %d: %s\n", item.uuid, item.reason)
			}
			return
		}
		fmt.Printf("%s✓ 所有门店验证通过%s\n\n", greenColor, resetColor)

		// 初始化服务依赖
		fmt.Printf("%s[2/3] 初始化服务...%s\n", blueColor, resetColor)
		syncSrv := initSyncService(dbm)
		fmt.Printf("%s✓ 服务初始化完成%s\n\n", greenColor, resetColor)

		// 执行批量同步
		fmt.Printf("%s[3/3] 开始同步...%s\n", blueColor, resetColor)
		results := runBatchSync(companyUuids, batchWorkers, dbm, syncSrv)

		// 输出汇总
		printSummary(results, time.Since(startTime))
	},
}

// parseBatchSyncCompanyUuids 解析逗号分隔的 company_uuid 列表
func parseBatchSyncCompanyUuids(input string) ([]uint64, error) {
	if input == "" {
		return nil, fmt.Errorf("companies 参数不能为空")
	}

	parts := strings.Split(input, ",")
	uuids := make([]uint64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		uuid, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的 company_uuid: %s", part)
		}
		uuids = append(uuids, uuid)
	}

	return uuids, nil
}

// invalidUuid 无效 UUID 信息
type invalidUuid struct {
	uuid   uint64
	reason string
}

// validateCompanyUuids 验证所有 company_uuid 有效性
func validateCompanyUuids(uuids []uint64, dbm *database.DBManager) []invalidUuid {
	var invalidList []invalidUuid

	for _, uuid := range uuids {
		// 从 DBManager 获取数据库连接（统一管理）
		companyDB := dbm.GetDB(uuid)
		if companyDB == nil {
			invalidList = append(invalidList, invalidUuid{uuid: uuid, reason: "数据库连接失败: DBManager 返回 nil"})
			continue
		}

		// 获取公司信息
		companyRepo := repository.NewCompanyRepo(companyDB)
		company, err := companyRepo.GetCompanyInfoByUuid(uuid)
		if err != nil {
			invalidList = append(invalidList, invalidUuid{uuid: uuid, reason: fmt.Sprintf("获取公司信息失败: %v", err)})
			continue
		}

		// 检查 ERP 开启状态
		if !company.IsOpenErp() {
			invalidList = append(invalidList, invalidUuid{uuid: uuid, reason: "公司未开启 ERP"})
			continue
		}
	}

	return invalidList
}

// initSyncService 初始化同步服务及其依赖
func initSyncService(dbm *database.DBManager) service.ISyncSrv {
	cacheInstance := cache.Global
	localeSrv := service.NewLocaleSrv()
	settingSrv := setting.NewSrv(dbm, cacheInstance)
	translateSrv := service.NewTranslateSrv(dbm, cacheInstance)
	messageSrv := message.NewIMessageSrv(dbm)
	supplierSrv := service.NewSupplierSrv(dbm)
	productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cacheInstance, translateSrv)
	materialSrv := service.NewMaterialSrv(dbm, localeSrv, settingSrv, translateSrv, messageSrv)
	warehouseSrv := service.NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)

	return service.NewSyncSrv(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv, paymentMethodSrv)
}

// runBatchSync 执行批量同步
func runBatchSync(companyUuids []uint64, workers int, dbm *database.DBManager, syncSrv service.ISyncSrv) []SyncResult {
	total := len(companyUuids)
	jobs := make(chan uint64, total)
	results := make(chan SyncResult, total)

	// 进度计数器
	var completed int32

	// 启动 workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		utils.Go(func() {
			defer wg.Done()
			for companyUuid := range jobs {
				result := syncOneCompany(companyUuid, dbm, syncSrv)
				results <- result

				// 更新进度并打印
				current := atomic.AddInt32(&completed, 1)
				printBatchSyncProgress(int(current), total, result)
			}
		})
	}

	// 分发任务
	for _, uuid := range companyUuids {
		jobs <- uuid
	}
	close(jobs)

	// 等待完成
	wg.Wait()
	close(results)

	// 收集结果
	allResults := make([]SyncResult, 0, total)
	for r := range results {
		allResults = append(allResults, r)
	}

	return allResults
}

// syncOneCompany 同步单个门店
func syncOneCompany(companyUuid uint64, dbm *database.DBManager, syncSrv service.ISyncSrv) SyncResult {
	start := time.Now()
	result := SyncResult{CompanyUuid: companyUuid}

	logger.Logger.Info("开始同步门店",
		zap.Uint64("company_uuid", companyUuid),
	)

	// 从 DBManager 获取门店数据库连接（统一管理，无需手动关闭）
	companyDB := dbm.GetDB(companyUuid)
	if companyDB == nil {
		result.Error = fmt.Errorf("数据库连接失败: DBManager 返回 nil")
		logger.Logger.Error("同步失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(result.Error),
		)
		return result
	}

	// 获取公司信息
	companyRepo := repository.NewCompanyRepo(companyDB)
	company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		result.Error = fmt.Errorf("获取公司信息失败: %v", err)
		logger.Logger.Error("同步失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(result.Error),
		)
		return result
	}
	result.CompanyName = company.Name

	// 检查 CompanySetting 是否存在
	if company.CompanySetting == nil {
		result.Error = fmt.Errorf("公司设置不存在")
		logger.Logger.Error("同步失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.String("company_name", company.Name),
			zap.Error(result.Error),
		)
		return result
	}

	// 设置上下文
	ctx := ttposContext.NewContext()
	ctx.SetDB(companyDB)
	ctx.SetCompanyUuid(companyUuid)
	ctx.SetCompany(*company)
	ctx.SetCompanySetting(*company.CompanySetting)
	ctx.SetLanguage("zh")

	// 执行同步（Sync 内部已有分布式锁，会返回 "数据同步中，请稍后再试" 错误）
	_, err = syncSrv.Sync(ctx, req.SyncReq{IsSyncExecute: true})
	if err != nil {
		// 检查是否因为正在同步中而被跳过
		if strings.Contains(err.Error(), "数据同步中") {
			result.Skipped = true
			result.Error = fmt.Errorf("门店正在同步中，已跳过")
			result.Duration = time.Since(start)
			logger.Logger.Warn("跳过同步：门店正在同步中",
				zap.Uint64("company_uuid", companyUuid),
				zap.String("company_name", company.Name),
			)
			return result
		}
		result.Error = fmt.Errorf("同步执行失败: %v", err)
		result.Duration = time.Since(start)
		logger.Logger.Error("同步失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.String("company_name", company.Name),
			zap.Error(result.Error),
		)
		return result
	}

	result.Success = true
	result.Duration = time.Since(start)

	logger.Logger.Info("同步成功",
		zap.Uint64("company_uuid", companyUuid),
		zap.String("company_name", company.Name),
		zap.Duration("duration", result.Duration),
	)

	return result
}

// printBatchSyncProgress 打印单个门店同步进度
func printBatchSyncProgress(completed, total int, result SyncResult) {
	var status string
	if result.Skipped {
		status = yellowColor + "⊘" + resetColor // 跳过
	} else if result.Success {
		status = greenColor + "✓" + resetColor // 成功
	} else {
		status = redColor + "✗" + resetColor // 失败
	}

	name := result.CompanyName
	if name == "" {
		name = "-"
	}

	errMsg := ""
	if result.Error != nil {
		if result.Skipped {
			errMsg = fmt.Sprintf(" | %s%v%s", yellowColor, result.Error, resetColor)
		} else {
			errMsg = fmt.Sprintf(" | %s%v%s", redColor, result.Error, resetColor)
		}
	}

	fmt.Printf("[%d/%d] %s %d (%s) %v%s\n",
		completed, total, status,
		result.CompanyUuid, name,
		result.Duration.Round(time.Millisecond),
		errMsg,
	)
}

// printSummary 打印同步结果汇总
func printSummary(results []SyncResult, totalElapsed time.Duration) {
	var success, failed, skipped int
	var failedList, skippedList []SyncResult
	var totalDuration time.Duration

	for _, r := range results {
		if r.Skipped {
			skipped++
			skippedList = append(skippedList, r)
		} else if r.Success {
			success++
			totalDuration += r.Duration
		} else {
			failed++
			failedList = append(failedList, r)
		}
	}

	fmt.Printf("\n%s========== 同步完成 ==========%s\n", blueColor, resetColor)
	fmt.Printf("总数: %d, %s成功: %d%s, %s失败: %d%s, %s跳过: %d%s\n",
		len(results),
		greenColor, success, resetColor,
		redColor, failed, resetColor,
		yellowColor, skipped, resetColor,
	)
	fmt.Printf("总耗时: %v\n", totalElapsed.Round(time.Millisecond))

	if success > 0 {
		avgDuration := totalDuration / time.Duration(success)
		fmt.Printf("平均耗时: %v\n", avgDuration.Round(time.Millisecond))
	}

	if skipped > 0 {
		fmt.Printf("\n%s跳过门店列表（正在同步中）:%s\n", yellowColor, resetColor)
		for _, r := range skippedList {
			name := r.CompanyName
			if name == "" {
				name = "-"
			}
			fmt.Printf("  - %d (%s)\n", r.CompanyUuid, name)
		}
		fmt.Printf("\n%s提示: 被跳过的门店正在由其他进程同步，无需重试%s\n", yellowColor, resetColor)
	}

	if failed > 0 {
		fmt.Printf("\n%s失败门店列表:%s\n", redColor, resetColor)
		for _, r := range failedList {
			name := r.CompanyName
			if name == "" {
				name = "-"
			}
			fmt.Printf("  - %d (%s): %v\n", r.CompanyUuid, name, r.Error)
		}
		fmt.Printf("\n%s提示: 可将失败的 UUID 重新执行同步%s\n", yellowColor, resetColor)
	}
}
