package command

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/otlp"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().StringVar(&batchRegenerateCompanyUuidsFlag, "company-uuids", "", "公司UUID列表，逗号分隔")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().StringVar(&batchRegenerateStartDateFlag, "start-date", "", "起始日期，格式：YYYY-MM-DD")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().StringVar(&batchRegenerateOpenPosEntryNameFlag, "open-pos-entry-name", "", "OpenPosEntryName")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().StringVar(&batchRegenerateTaskFileFlag, "task-file", "", "任务清单文件路径（可选，默认：./batch-regenerate-task-{timestamp}.json）")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().BoolVar(&batchRegenerateResumeFlag, "resume", false, "从现有任务清单继续执行")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().BoolVar(&batchRegenerateDryRunFlag, "dry-run", false, "预览模式，仅生成任务清单不执行")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().BoolVar(&batchRegenerateShowProgressFlag, "show-progress", false, "显示详细进度信息")
	batchRegenerateOrderMaterialAndInvoiceCmd.Flags().IntVar(&batchRegenerateProgressIntervalFlag, "progress-interval", 5, "进度信息刷新间隔（秒数，默认：5秒）")
	batchRegenerateOrderMaterialAndInvoiceCmd.MarkFlagRequired("company-uuids")
	batchRegenerateOrderMaterialAndInvoiceCmd.MarkFlagRequired("start-date")
	batchRegenerateOrderMaterialAndInvoiceCmd.MarkFlagRequired("open-pos-entry-name")
	rootCommand.AddCommand(batchRegenerateOrderMaterialAndInvoiceCmd)
}

var (
	batchRegenerateCompanyUuidsFlag      string
	batchRegenerateStartDateFlag         string
	batchRegenerateOpenPosEntryNameFlag  string
	batchRegenerateTaskFileFlag          string
	batchRegenerateResumeFlag            bool
	batchRegenerateDryRunFlag            bool
	batchRegenerateShowProgressFlag      bool
	batchRegenerateProgressIntervalFlag  int
)

// 颜色常量（ANSI转义序列）- 使用字符串字面量避免包级别冲突
const (
	batchRegenerateBlueColor   = "\033[34m" // 蓝色
	batchRegenerateYellowColor = "\033[33m" // 黄色
	batchRegenerateRedColor    = "\033[31m" // 红色
	batchRegenerateGreenColor  = "\033[32m" // 绿色
	batchRegenerateResetColor  = "\033[0m"  // 重置颜色
)

// batch-regenerate-order-material-and-invoice 批量重新生成订单材料消耗和POS发票
var batchRegenerateOrderMaterialAndInvoiceCmd = &cobra.Command{
	Use:   "batch-regenerate-order-material-and-invoice",
	Short: "批量重新生成订单材料消耗和POS发票",
	Long:  `批量重新生成指定公司和日期范围内的所有订单的材料消耗和POS发票`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		config.Server.Mode = "release"

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化国际化
		i18n.Init()

		// 自定义验证规则
		validator.Init()

		// 初始化id生成器
		utils.InitIdGenerator()

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁（必须在调用 NewSystemLock 之前）
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化短信客户端
		sms.InitClient(config.SMS.APIKey, config.SMS.BaseURL, config.SMS.ProjectName)
		// 检查短信客户端配置
		if err := sms.GetSMSClient().CheckConfig(); err != nil {
			fmt.Printf("[FATAL] Failed to check SMS client config: %v\n", err)
			logger.Logger.Info("Failed to check SMS client config", zap.Error(err))
		}

		//初始化服务发现
		cloud.Init()

		// 初始化 OTLP 调用链跟踪
		if err := otlp.Init(context.Background(), config.Otlp); err != nil {
			fmt.Printf("[FATAL] Failed to initialize OpenTelemetry: %v\n", err)
			logger.Logger.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 记录命令开始执行
		logger.Logger.Info("批量重新生成订单材料消耗和POS发票命令开始执行",
			zap.String("company_uuids", batchRegenerateCompanyUuidsFlag),
			zap.String("start_date", batchRegenerateStartDateFlag),
			zap.String("task_file", batchRegenerateTaskFileFlag),
			zap.Bool("resume", batchRegenerateResumeFlag),
			zap.Bool("dry_run", batchRegenerateDryRunFlag),
			zap.Bool("show_progress", batchRegenerateShowProgressFlag),
			zap.Int("progress_interval", batchRegenerateProgressIntervalFlag))

		// 解析参数
		companyUuids, err := parseCompanyUuids(batchRegenerateCompanyUuidsFlag)
		if err != nil {
			fmt.Printf("%s错误: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
			logger.Logger.Error("解析公司UUID列表失败",
				zap.String("company_uuids", batchRegenerateCompanyUuidsFlag),
				zap.Error(err))
			return
		}
		logger.Logger.Info("公司UUID列表解析成功", zap.Int("count", len(companyUuids)), zap.Any("uuids", companyUuids))

		// 验证日期格式
		if err := validateDate(batchRegenerateStartDateFlag); err != nil {
			fmt.Printf("%s错误: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
			logger.Logger.Error("日期格式验证失败",
				zap.String("start_date", batchRegenerateStartDateFlag),
				zap.Error(err))
			return
		}

		// 验证 OpenPosEntryName
		if batchRegenerateOpenPosEntryNameFlag == "" {
			fmt.Printf("%s错误: OpenPosEntryName不能为空%s\n", batchRegenerateRedColor, batchRegenerateResetColor)
			logger.Logger.Error("OpenPosEntryName不能为空")
			return
		}

		// 确定任务清单文件路径
		taskFilePath := batchRegenerateTaskFileFlag
		if taskFilePath == "" {
			timestamp := time.Now().Unix()
			taskFilePath = fmt.Sprintf("./batch-regenerate-task-%d.json", timestamp)
		}
		logger.Logger.Info("任务清单文件路径", zap.String("task_file", taskFilePath))

		// 初始化数据库管理器
		dbm := database.GetDBManager(config.Database)

		// 初始化Service
		settingSrv := setting.NewSrv(dbm, cache.Global)
		salesOutboundSummarySrv := service.NewSalesOutboundSummarySrv(dbm, settingSrv, cache.Global)
		taskManager := service.NewBatchRegenerateTaskManager(dbm, salesOutboundSummarySrv, batchRegenerateOpenPosEntryNameFlag)
		logger.Logger.Info("服务初始化完成")

		// 如果是resume模式，加载现有任务清单
		var task *dto.BatchRegenerateTask
		if batchRegenerateResumeFlag {
			fmt.Printf("%s正在加载任务清单: %s%s\n", batchRegenerateBlueColor, taskFilePath, batchRegenerateResetColor)
			logger.Logger.Info("从现有任务清单继续执行", zap.String("task_file", taskFilePath))
			task, err = taskManager.LoadTaskList(taskFilePath)
			if err != nil {
				fmt.Printf("%s错误: 加载任务清单失败: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
				logger.Logger.Error("加载任务清单失败",
					zap.String("task_file", taskFilePath),
					zap.Error(err))
				return
			}
			fmt.Printf("%s✓ 任务清单加载成功%s\n", batchRegenerateGreenColor, batchRegenerateResetColor)
			logger.Logger.Info("任务清单加载成功",
				zap.String("task_file", taskFilePath),
				zap.Int("total_companies", task.Summary.TotalCompanies),
				zap.Int("total_dates", task.Summary.TotalDates),
				zap.Int("total_orders", task.Summary.TotalOrders),
				zap.Int("completed_steps", task.Summary.CompletedSteps),
				zap.Int("failed_steps", task.Summary.FailedSteps),
				zap.Int("pending_steps", task.Summary.PendingSteps))
		} else {
			// 生成任务清单
			fmt.Printf("%s正在生成任务清单...%s\n", batchRegenerateBlueColor, batchRegenerateResetColor)
			logger.Logger.Info("开始生成任务清单",
				zap.Int("company_count", len(companyUuids)),
				zap.String("start_date", batchRegenerateStartDateFlag))
			task, err = taskManager.GenerateTaskList(companyUuids, batchRegenerateStartDateFlag, taskFilePath)
			if err != nil {
				fmt.Printf("%s错误: 生成任务清单失败: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
				logger.Logger.Error("生成任务清单失败",
					zap.Int("company_count", len(companyUuids)),
					zap.String("start_date", batchRegenerateStartDateFlag),
					zap.Error(err))
				return
			}

			// 保存任务清单
			if err := taskManager.SaveTaskList(task, taskFilePath); err != nil {
				fmt.Printf("%s错误: 保存任务清单失败: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
				logger.Logger.Error("保存任务清单失败",
					zap.String("task_file", taskFilePath),
					zap.Error(err))
				return
			}
			fmt.Printf("%s✓ 任务清单生成成功: %s%s\n", batchRegenerateGreenColor, taskFilePath, batchRegenerateResetColor)
			logger.Logger.Info("任务清单生成成功",
				zap.String("task_file", taskFilePath),
				zap.Int("total_companies", task.Summary.TotalCompanies),
				zap.Int("total_dates", task.Summary.TotalDates),
				zap.Int("total_orders", task.Summary.TotalOrders),
				zap.Int("total_steps", task.Summary.TotalDateSteps+task.Summary.TotalOrderSteps))
		}

		// 输出任务清单统计信息
		printTaskSummary(task)

		// 如果是dry-run模式，只生成任务清单不执行
		if batchRegenerateDryRunFlag {
			fmt.Printf("%s预览模式，任务清单已生成，不执行实际操作%s\n", batchRegenerateYellowColor, batchRegenerateResetColor)
			return
		}

		// 执行任务清单
		fmt.Printf("%s开始执行任务清单...%s\n", batchRegenerateBlueColor, batchRegenerateResetColor)
		logger.Logger.Info("开始执行任务清单",
			zap.String("task_file", taskFilePath),
			zap.Bool("show_progress", batchRegenerateShowProgressFlag),
			zap.Int("progress_interval", batchRegenerateProgressIntervalFlag))
		startTime := time.Now()

		// 如果启用进度显示，启动进度显示 goroutine
		var progressTicker *time.Ticker
		var progressStopChan chan bool
		if batchRegenerateShowProgressFlag {
			if batchRegenerateProgressIntervalFlag <= 0 {
				batchRegenerateProgressIntervalFlag = 5 // 默认5秒
			}
			progressTicker = time.NewTicker(time.Duration(batchRegenerateProgressIntervalFlag) * time.Second)
			progressStopChan = make(chan bool)
			go func() {
				for {
					select {
					case <-progressTicker.C:
						// 重新加载任务清单以获取最新进度
						currentTask, err := taskManager.LoadTaskList(taskFilePath)
						if err == nil {
							progress, err := taskManager.GetProgress(currentTask)
							if err == nil {
								printProgress(progress)
							}
						}
					case <-progressStopChan:
						return
					}
				}
			}()
		}

		err = taskManager.ExecuteTaskList(task, taskFilePath, false, 0) // 不在 service 层显示进度
		duration := time.Since(startTime)

		// 停止进度显示
		if progressTicker != nil {
			progressTicker.Stop()
		}
		if progressStopChan != nil {
			close(progressStopChan)
			// 等待进度显示 goroutine 退出
			time.Sleep(100 * time.Millisecond)
		}

		// 如果启用了进度显示，显示最终进度
		if batchRegenerateShowProgressFlag {
			// 重新加载任务清单以获取最新进度
			finalTask, loadErr := taskManager.LoadTaskList(taskFilePath)
			if loadErr == nil {
				finalProgress, progressErr := taskManager.GetProgress(finalTask)
				if progressErr == nil {
					printProgress(finalProgress)
				}
			}
		}

		if err != nil {
			fmt.Printf("%s错误: 执行任务清单失败: %s%s\n", batchRegenerateRedColor, err.Error(), batchRegenerateResetColor)
			logger.Logger.Error("执行任务清单失败",
				zap.String("task_file", taskFilePath),
				zap.Duration("duration", duration),
				zap.Error(err))
			return
		}

		// 重新加载任务清单以获取最新统计信息
		finalTask, loadErr := taskManager.LoadTaskList(taskFilePath)
		if loadErr != nil {
			logger.Logger.Warn("重新加载任务清单失败，使用旧数据",
				zap.String("task_file", taskFilePath),
				zap.Error(loadErr))
			finalTask = task
		}

		// 输出最终统计信息
		fmt.Printf("%s✓ 任务执行完成，耗时: %v%s\n", batchRegenerateGreenColor, duration, batchRegenerateResetColor)
		printTaskSummary(finalTask)

		// 记录执行完成日志
		logger.Logger.Info("批量重新生成任务执行完成",
			zap.String("task_file", taskFilePath),
			zap.Duration("duration", duration),
			zap.Int("total_companies", finalTask.Summary.TotalCompanies),
			zap.Int("total_dates", finalTask.Summary.TotalDates),
			zap.Int("total_orders", finalTask.Summary.TotalOrders),
			zap.Int("completed_steps", finalTask.Summary.CompletedSteps),
			zap.Int("failed_steps", finalTask.Summary.FailedSteps),
			zap.Int("pending_steps", finalTask.Summary.PendingSteps))
	},
}

// parseCompanyUuids 解析公司UUID列表（逗号分隔）
func parseCompanyUuids(companyUuidsStr string) ([]uint64, error) {
	if companyUuidsStr == "" {
		return nil, fmt.Errorf("公司UUID列表不能为空")
	}

	parts := strings.Split(companyUuidsStr, ",")
	companyUuids := make([]uint64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		uuid, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的公司UUID: %s", part)
		}
		if uuid == 0 {
			return nil, fmt.Errorf("公司UUID不能为0")
		}
		companyUuids = append(companyUuids, uuid)
	}

	if len(companyUuids) == 0 {
		return nil, fmt.Errorf("公司UUID列表不能为空")
	}

	return companyUuids, nil
}

// validateDate 验证日期格式
func validateDate(dateStr string) error {
	if dateStr == "" {
		return fmt.Errorf("日期不能为空")
	}

	// 验证日期格式 YYYY-MM-DD
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %s", dateStr)
	}

	return nil
}

// printTaskSummary 输出任务清单统计信息
func printTaskSummary(task *dto.BatchRegenerateTask) {
	fmt.Printf("\n%s========== 任务清单统计 ==========%s\n", batchRegenerateBlueColor, batchRegenerateResetColor)
	fmt.Printf("任务清单文件: %s\n", batchRegenerateTaskFileFlag)
	fmt.Printf("创建时间: %s\n", task.CreatedAt)
	fmt.Printf("更新时间: %s\n", task.UpdatedAt)
	fmt.Printf("\n%s统计信息:%s\n", batchRegenerateYellowColor, batchRegenerateResetColor)
	fmt.Printf("  总公司数: %d\n", task.Summary.TotalCompanies)
	fmt.Printf("  总日期数: %d\n", task.Summary.TotalDates)
	fmt.Printf("  总订单数: %d\n", task.Summary.TotalOrders)
	fmt.Printf("  总步骤数: %d (日期级别: %d, 订单级别: %d)\n",
		task.Summary.TotalDateSteps+task.Summary.TotalOrderSteps,
		task.Summary.TotalDateSteps,
		task.Summary.TotalOrderSteps)
	fmt.Printf("  已完成: %d\n", task.Summary.CompletedSteps)
	fmt.Printf("  失败: %d\n", task.Summary.FailedSteps)
	fmt.Printf("  待执行: %d\n", task.Summary.PendingSteps)
	fmt.Printf("%s=====================================%s\n\n", batchRegenerateBlueColor, batchRegenerateResetColor)
}

// printProgress 格式化输出进度信息
func printProgress(progress *resp.BatchRegenerateProgressResp) {
	// 清屏并移动光标到开头（使用 ANSI 转义序列）
	fmt.Print("\033[2J\033[H") // 清屏并移动光标到 (0,0)

	// 输出总体进度
	fmt.Printf("%s========== 批量重新生成进度 ==========%s\n", batchRegenerateBlueColor, batchRegenerateResetColor)
	fmt.Printf("%s总体进度: %.2f%%%s\n", batchRegenerateGreenColor, progress.OverallProgress, batchRegenerateResetColor)

	// 进度条
	barLen := 50
	filled := int(progress.OverallProgress / 100 * float64(barLen))
	if filled > barLen {
		filled = barLen
	}
	barFilled := strings.Repeat("█", filled)
	barEmpty := strings.Repeat("░", barLen-filled)
	fmt.Printf("[%s%s] %.2f%%\n", batchRegenerateGreenColor+barFilled+batchRegenerateResetColor, barEmpty, progress.OverallProgress)

	// 当前执行的命令
	if progress.CurrentStep.StepName != "" {
		fmt.Printf("\n%s当前执行:%s\n", batchRegenerateYellowColor, batchRegenerateResetColor)
		stepDisplayName := getStepDisplayName(progress.CurrentStep.StepName)
		if progress.CurrentStep.StepType == "date" {
			fmt.Printf("  命令: %s%s%s\n", batchRegenerateBlueColor, stepDisplayName, batchRegenerateResetColor)
			fmt.Printf("  公司: %s%s%s\n", batchRegenerateBlueColor, progress.CurrentStep.CompanyName, batchRegenerateResetColor)
			fmt.Printf("  日期: %s%s%s\n", batchRegenerateBlueColor, progress.CurrentStep.Date, batchRegenerateResetColor)
		} else if progress.CurrentStep.StepType == "order" {
			fmt.Printf("  命令: %s%s%s\n", batchRegenerateBlueColor, stepDisplayName, batchRegenerateResetColor)
			fmt.Printf("  公司: %s%s%s\n", batchRegenerateBlueColor, progress.CurrentStep.CompanyName, batchRegenerateResetColor)
			fmt.Printf("  日期: %s%s%s\n", batchRegenerateBlueColor, progress.CurrentStep.Date, batchRegenerateResetColor)
			fmt.Printf("  订单: %s%s%s\n", batchRegenerateBlueColor, progress.CurrentStep.OrderNo, batchRegenerateResetColor)
		}
	}

	// 步骤统计
	fmt.Printf("\n%s步骤统计:%s\n", batchRegenerateYellowColor, batchRegenerateResetColor)
	fmt.Printf("  已完成: %s%d%s\n", batchRegenerateGreenColor, progress.CompletedSteps, batchRegenerateResetColor)
	fmt.Printf("  失败: %s%d%s\n", batchRegenerateRedColor, progress.FailedSteps, batchRegenerateResetColor)
	fmt.Printf("  待执行: %s%d%s\n", batchRegenerateYellowColor, progress.PendingSteps, batchRegenerateResetColor)
	if progress.EstimatedTimeLeft != "" {
		fmt.Printf("  预计剩余时间: %s%s%s\n", batchRegenerateBlueColor, progress.EstimatedTimeLeft, batchRegenerateResetColor)
	}

	// 公司级别进度
	fmt.Printf("\n%s公司进度:%s\n", batchRegenerateYellowColor, batchRegenerateResetColor)
	fmt.Printf("  总数: %d\n", progress.CompanyProgress.TotalCompanies)
	fmt.Printf("  待处理: %d\n", progress.CompanyProgress.PendingCompanies)
	if len(progress.CompanyProgress.CompletedCompanies) > 0 {
		fmt.Printf("  已完成: %s%s%s\n", batchRegenerateGreenColor, strings.Join(progress.CompanyProgress.CompletedCompanies, ", "), batchRegenerateResetColor)
	}
	if progress.CompanyProgress.CurrentCompany != "" {
		fmt.Printf("  当前处理: %s%s%s\n", batchRegenerateBlueColor, progress.CompanyProgress.CurrentCompany, batchRegenerateResetColor)
	}

	// 日期级别进度
	if progress.CompanyProgress.CurrentCompany != "" {
		fmt.Printf("\n%s日期进度 (%s):%s\n", batchRegenerateYellowColor, progress.CompanyProgress.CurrentCompany, batchRegenerateResetColor)
		fmt.Printf("  总数: %d\n", progress.DateProgress.TotalDates)
		fmt.Printf("  待处理: %d\n", progress.DateProgress.PendingDates)
		if len(progress.DateProgress.CompletedDates) > 0 {
			fmt.Printf("  已完成: %s%s%s\n", batchRegenerateGreenColor, strings.Join(progress.DateProgress.CompletedDates, ", "), batchRegenerateResetColor)
		}
		if progress.DateProgress.CurrentDate != "" {
			fmt.Printf("  当前处理: %s%s%s\n", batchRegenerateBlueColor, progress.DateProgress.CurrentDate, batchRegenerateResetColor)
		}
	}

	// 订单级别进度
	if progress.DateProgress.CurrentDate != "" {
		fmt.Printf("\n%s订单进度 (%s):%s\n", batchRegenerateYellowColor, progress.DateProgress.CurrentDate, batchRegenerateResetColor)
		fmt.Printf("  总数: %d\n", progress.OrderProgress.TotalOrders)
		fmt.Printf("  待处理: %d\n", progress.OrderProgress.PendingOrders)
		if len(progress.OrderProgress.CompletedOrders) > 0 {
			// 只显示最后5个已完成的订单
			completedOrders := progress.OrderProgress.CompletedOrders
			if len(completedOrders) > 5 {
				completedOrders = completedOrders[len(completedOrders)-5:]
			}
			fmt.Printf("  已完成: %s%s%s\n", batchRegenerateGreenColor, strings.Join(completedOrders, ", "), batchRegenerateResetColor)
		}
		if progress.OrderProgress.CurrentOrder != "" {
			fmt.Printf("  当前处理: %s%s%s\n", batchRegenerateBlueColor, progress.OrderProgress.CurrentOrder, batchRegenerateResetColor)
		}
	}

	fmt.Printf("%s=====================================%s\n", batchRegenerateBlueColor, batchRegenerateResetColor)
	fmt.Printf("更新时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

// getStepDisplayName 获取步骤的显示名称
func getStepDisplayName(stepName string) string {
	stepNameMap := map[string]string{
		"regenerate-order-material":               "重新生成订单材料消耗",
		"regenerate-sale-order-material-outbound": "重新生成订单材料出库",
		"regenerate-order-pos-invoice":            "重新生成订单POS发票",
		"regenerate-sales-outbound":               "重新生成销售出库汇总",
	}
	if displayName, ok := stepNameMap[stepName]; ok {
		return displayName
	}
	return stepName
}
