package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IBatchRegenerateTaskManager 批量重新生成任务管理器接口
type IBatchRegenerateTaskManager interface {
	// GenerateTaskList 生成任务清单
	// companyUuids: 公司UUID列表
	// startDate: 起始日期，格式：YYYY-MM-DD
	// endDate: 结束日期，格式：YYYY-MM-DD（可选，为空时不限制结束日期）
	// taskFilePath: 任务清单文件路径（可选，为空时自动生成）
	GenerateTaskList(companyUuids []uint64, startDate string, endDate string, taskFilePath string) (*dto.BatchRegenerateTask, error)

	// LoadTaskList 加载任务清单
	// taskFilePath: 任务清单文件路径
	LoadTaskList(taskFilePath string) (*dto.BatchRegenerateTask, error)

	// SaveTaskList 保存任务清单
	// task: 任务清单对象
	// taskFilePath: 任务清单文件路径
	SaveTaskList(task *dto.BatchRegenerateTask, taskFilePath string) error

	// ExecuteTaskList 执行任务清单
	// task: 任务清单对象
	// taskFilePath: 任务清单文件路径
	// showProgress: 是否显示进度
	// progressInterval: 进度刷新间隔（秒）
	ExecuteTaskList(task *dto.BatchRegenerateTask, taskFilePath string, showProgress bool, progressInterval int) error

	// GetProgress 获取进度信息
	// task: 任务清单对象
	GetProgress(task *dto.BatchRegenerateTask) (*resp.BatchRegenerateProgressResp, error)
}

// batchRegenerateTaskManager 批量重新生成任务管理器实现
type batchRegenerateTaskManager struct {
	dbm                     *database.DBManager
	salesOutboundSummarySrv ISalesOutboundSummarySrv
	openPosEntryName        string
	fileLock                *sync.Mutex // 文件锁，防止并发操作任务清单文件
}

// NewBatchRegenerateTaskManager 创建批量重新生成任务管理器实例
func NewBatchRegenerateTaskManager(
	dbm *database.DBManager,
	salesOutboundSummarySrv ISalesOutboundSummarySrv,
	openPosEntryName string,
) IBatchRegenerateTaskManager {
	return &batchRegenerateTaskManager{
		dbm:                     dbm,
		salesOutboundSummarySrv: salesOutboundSummarySrv,
		openPosEntryName:        openPosEntryName,
		fileLock:                &sync.Mutex{},
	}
}

// GenerateTaskList 生成任务清单
func (m *batchRegenerateTaskManager) GenerateTaskList(
	companyUuids []uint64,
	startDate string,
	endDate string,
	taskFilePath string,
) (*dto.BatchRegenerateTask, error) {
	now := time.Now().UTC()
	createdAt := now.Format(time.RFC3339)

	task := &dto.BatchRegenerateTask{
		Companies: make([]dto.CompanyTask, 0),
		Summary:   dto.TaskSummary{},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	// 获取saas数据库连接（用于查询公司信息）
	saasDB := m.dbm.GetDB(0)
	companyRepo := repository.NewCompanyRepo(saasDB)

	// 遍历每个公司
	for _, companyUuid := range companyUuids {
		// 获取公司信息
		company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("获取公司信息失败: company_uuid=%d", companyUuid))
		}

		// 获取门店时区
		timezone := "Asia/Shanghai" // 默认时区
		if company.CompanySetting != nil {
			timezone = company.CompanySetting.GetTimezone()
		}
		timeUtil := utils.Timezone(timezone)

		// 使用商家时区解析起始日期
		startTime, err := timeUtil.FormatTimeToTime(startDate)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("日期格式错误，应为 YYYY-MM-DD: company_uuid=%d", companyUuid))
		}
		startTimestamp := startTime.Unix()

		// 使用商家时区解析结束日期（如果提供）
		var endTimestamp int64 = 0
		if endDate != "" {
			endTime, err := timeUtil.FormatTimeToTime(endDate)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("结束日期格式错误，应为 YYYY-MM-DD: company_uuid=%d", companyUuid))
			}
			// 设置为当天的 23:59:59（使用商家时区）
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
			endTimestamp = endTime.Unix()
		}

		// 获取该公司的数据库连接
		db := m.dbm.GetDB(companyUuid)

		// 查询符合条件的订单（status=1已结账，created_at >= startDate，未删除）
		opts := []repository.DBOption{
			func(db *gorm.DB) *gorm.DB {
				return db.Where("status = ?", constant.SaleOrderStatusFinish).
					Where("create_time >= ?", startTimestamp).
					Where("delete_time = ?", constant.NotDeleted)
			},
		}
		if endTimestamp > 0 {
			opts = append(opts, func(db *gorm.DB) *gorm.DB {
				return db.Where("create_time <= ?", endTimestamp)
			})
		}
		opts = append(opts, func(db *gorm.DB) *gorm.DB {
			return db.Order("create_time ASC")
		})
		saleOrders, err := repository.NewSaleOrderRepo(db).GetSaleOrderList(opts...)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("查询订单失败: company_uuid=%d", companyUuid))
		}

		// 按日期分组订单（使用门店时区转换时间）
		dateOrderMap := make(map[string][]model.SaleOrder) // key: 日期字符串 YYYY-MM-DD
		for _, order := range saleOrders {
			orderDate := timeUtil.FormatUnixTime(order.CreateTime, "2006-01-02")
			dateOrderMap[orderDate] = append(dateOrderMap[orderDate], order)
		}

		// 生成日期任务列表
		dateTasks := make([]dto.DateTask, 0)
		for dateStr, orders := range dateOrderMap {
			// 生成订单任务列表
			orderTasks := make([]dto.OrderTask, 0)
			for _, order := range orders {
				orderTask := dto.OrderTask{
					SaleOrderUuid: order.Uuid,
					OrderNo:       order.OrderNo,
					OrderDate:     dateStr,
					Steps: []dto.StepTask{
						{
							Step:      1,
							Name:      "regenerate-order-material",
							Status:    dto.StepStatusPending,
							StartTime: "",
							EndTime:   "",
							Error:     "",
						},
						{
							Step:      2,
							Name:      "regenerate-sale-order-material-outbound",
							Status:    dto.StepStatusPending,
							StartTime: "",
							EndTime:   "",
							Error:     "",
						},
						{
							Step:      3,
							Name:      "regenerate-order-pos-invoice",
							Status:    dto.StepStatusPending,
							StartTime: "",
							EndTime:   "",
							Error:     "",
						},
					},
				}
				orderTasks = append(orderTasks, orderTask)
			}

			// 生成日期级别步骤
			dateTask := dto.DateTask{
				Date: dateStr,
				DateStep: dto.StepTask{
					Step:      1,
					Name:      "regenerate-sales-outbound",
					Status:    dto.StepStatusPending,
					StartTime: "",
					EndTime:   "",
					Error:     "",
				},
				Orders: orderTasks,
			}
			dateTasks = append(dateTasks, dateTask)
		}

		// 添加到公司任务列表
		companyTask := dto.CompanyTask{
			CompanyUuid: companyUuid,
			CompanyName: company.Name,
			Dates:       dateTasks,
		}
		task.Companies = append(task.Companies, companyTask)
	}

	// 计算统计信息
	task.Summary = m.calculateSummary(task)

	return task, nil
}

// calculateSummary 计算任务统计信息
func (m *batchRegenerateTaskManager) calculateSummary(task *dto.BatchRegenerateTask) dto.TaskSummary {
	summary := dto.TaskSummary{
		TotalCompanies:  len(task.Companies),
		TotalDates:      0,
		TotalOrders:     0,
		TotalDateSteps:  0,
		TotalOrderSteps: 0,
		CompletedSteps:  0,
		FailedSteps:     0,
		PendingSteps:    0,
	}

	for _, company := range task.Companies {
		summary.TotalDates += len(company.Dates)
		for _, dateTask := range company.Dates {
			summary.TotalDateSteps++
			summary.TotalOrders += len(dateTask.Orders)
			for _, orderTask := range dateTask.Orders {
				summary.TotalOrderSteps += len(orderTask.Steps)
				for _, step := range orderTask.Steps {
					switch step.Status {
					case dto.StepStatusCompleted:
						summary.CompletedSteps++
					case dto.StepStatusFailed:
						summary.FailedSteps++
					case dto.StepStatusPending, dto.StepStatusRunning:
						summary.PendingSteps++
					}
				}
			}
			// 日期级别步骤
			switch dateTask.DateStep.Status {
			case dto.StepStatusCompleted:
				summary.CompletedSteps++
			case dto.StepStatusFailed:
				summary.FailedSteps++
			case dto.StepStatusPending, dto.StepStatusRunning:
				summary.PendingSteps++
			}
		}
	}

	return summary
}

// UpdateStepStatus 更新步骤状态（辅助方法）
func (m *batchRegenerateTaskManager) UpdateStepStatus(
	task *dto.BatchRegenerateTask,
	companyIndex int,
	dateIndex int,
	orderIndex int,
	stepIndex int,
	isDateStep bool,
	status string,
	errorMsg string,
) error {
	now := time.Now().UTC().Format(time.RFC3339)

	if isDateStep {
		// 更新日期级别步骤
		if companyIndex < 0 || companyIndex >= len(task.Companies) {
			return errors.New("公司索引超出范围")
		}
		if dateIndex < 0 || dateIndex >= len(task.Companies[companyIndex].Dates) {
			return errors.New("日期索引超出范围")
		}
		step := &task.Companies[companyIndex].Dates[dateIndex].DateStep
		step.Status = status
		if status == dto.StepStatusRunning || status == dto.StepStatusCompleted || status == dto.StepStatusFailed {
			if step.StartTime == "" {
				step.StartTime = now
			}
		}
		if status == dto.StepStatusCompleted || status == dto.StepStatusFailed {
			step.EndTime = now
		}
		if errorMsg != "" {
			step.Error = errorMsg
		}
	} else {
		// 更新订单级别步骤
		if companyIndex < 0 || companyIndex >= len(task.Companies) {
			return errors.New("公司索引超出范围")
		}
		if dateIndex < 0 || dateIndex >= len(task.Companies[companyIndex].Dates) {
			return errors.New("日期索引超出范围")
		}
		if orderIndex < 0 || orderIndex >= len(task.Companies[companyIndex].Dates[dateIndex].Orders) {
			return errors.New("订单索引超出范围")
		}
		if stepIndex < 0 || stepIndex >= len(task.Companies[companyIndex].Dates[dateIndex].Orders[orderIndex].Steps) {
			return errors.New("步骤索引超出范围")
		}
		step := &task.Companies[companyIndex].Dates[dateIndex].Orders[orderIndex].Steps[stepIndex]
		step.Status = status
		if status == dto.StepStatusRunning || status == dto.StepStatusCompleted || status == dto.StepStatusFailed {
			if step.StartTime == "" {
				step.StartTime = now
			}
		}
		if status == dto.StepStatusCompleted || status == dto.StepStatusFailed {
			step.EndTime = now
		}
		if errorMsg != "" {
			step.Error = errorMsg
		}
	}

	return nil
}

// LoadTaskList 加载任务清单
func (m *batchRegenerateTaskManager) LoadTaskList(taskFilePath string) (*dto.BatchRegenerateTask, error) {
	// 检查文件是否存在
	if !utils.IsFileExist(taskFilePath) {
		return nil, errors.New(fmt.Sprintf("任务清单文件不存在: %s", taskFilePath))
	}

	// 读取文件内容
	content, err := os.ReadFile(taskFilePath)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("读取任务清单文件失败: %s", taskFilePath))
	}

	// 解析JSON
	var task dto.BatchRegenerateTask
	if err := json.Unmarshal(content, &task); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("解析任务清单JSON失败: %s", taskFilePath))
	}

	// 验证任务清单结构
	if err := m.validateTaskList(&task); err != nil {
		return nil, errors.WithMessage(err, "任务清单验证失败")
	}

	return &task, nil
}

// SaveTaskList 保存任务清单
func (m *batchRegenerateTaskManager) SaveTaskList(task *dto.BatchRegenerateTask, taskFilePath string) error {
	// 更新更新时间
	task.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// 重新计算统计信息
	task.Summary = m.calculateSummary(task)

	// 序列化为JSON（格式化输出，便于阅读）
	jsonBytes, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return errors.WithMessage(err, "序列化任务清单失败")
	}

	// 写入文件
	if err := utils.CreateFile(taskFilePath, jsonBytes, 0644); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("保存任务清单文件失败: %s", taskFilePath))
	}

	return nil
}

// validateTaskList 验证任务清单结构
func (m *batchRegenerateTaskManager) validateTaskList(task *dto.BatchRegenerateTask) error {
	// 验证必需字段
	if task.CreatedAt == "" {
		return errors.New("任务清单缺少 created_at 字段")
	}
	if task.UpdatedAt == "" {
		return errors.New("任务清单缺少 updated_at 字段")
	}

	// 验证公司列表
	for i, company := range task.Companies {
		if company.CompanyUuid == 0 {
			return errors.New(fmt.Sprintf("公司任务[%d]缺少 company_uuid 字段", i))
		}
		if company.CompanyName == "" {
			return errors.New(fmt.Sprintf("公司任务[%d]缺少 company_name 字段", i))
		}

		// 验证日期列表
		for j, dateTask := range company.Dates {
			if dateTask.Date == "" {
				return errors.New(fmt.Sprintf("公司[%d]日期任务[%d]缺少 date 字段", i, j))
			}

			// 验证日期级别步骤状态
			if !dto.IsValidStatus(dateTask.DateStep.Status) {
				return errors.New(fmt.Sprintf("公司[%d]日期[%s]的日期级别步骤状态无效: %s", i, dateTask.Date, dateTask.DateStep.Status))
			}

			// 验证订单列表
			for k, orderTask := range dateTask.Orders {
				if orderTask.SaleOrderUuid == 0 {
					return errors.New(fmt.Sprintf("公司[%d]日期[%s]订单[%d]缺少 sale_order_uuid 字段", i, dateTask.Date, k))
				}
				if orderTask.OrderNo == "" {
					return errors.New(fmt.Sprintf("公司[%d]日期[%s]订单[%d]缺少 order_no 字段", i, dateTask.Date, k))
				}

				// 验证订单步骤
				if len(orderTask.Steps) != 3 {
					return errors.New(fmt.Sprintf("公司[%d]日期[%s]订单[%d]的步骤数量不正确，应为3个", i, dateTask.Date, k))
				}
				for l, step := range orderTask.Steps {
					if !dto.IsValidStatus(step.Status) {
						return errors.New(fmt.Sprintf("公司[%d]日期[%s]订单[%d]步骤[%d]状态无效: %s", i, dateTask.Date, k, l, step.Status))
					}
				}
			}
		}
	}

	return nil
}

// ExecuteTaskList 执行任务清单
func (m *batchRegenerateTaskManager) ExecuteTaskList(
	task *dto.BatchRegenerateTask,
	taskFilePath string,
	showProgress bool,
	progressInterval int,
) error {
	// 1. 获取文件锁（防止多个实例同时操作同一任务清单文件）
	lockKey := fmt.Sprintf("batch_regenerate_task:%s", taskFilePath)
	systemLock := lock.NewSystemLock()
	if !systemLock.TryLockUuidString(lockKey) {
		return errors.New("任务清单正在被其他进程使用，请稍后再试")
	}
	defer systemLock.UnlockUuidString(lockKey)

	// 2. 启动进度显示（如果启用）
	var progressTicker *time.Ticker
	var progressStopChan chan bool
	if showProgress {
		if progressInterval <= 0 {
			progressInterval = 5 // 默认5秒
		}
		progressTicker = time.NewTicker(time.Duration(progressInterval) * time.Second)
		progressStopChan = make(chan bool)
		go m.startProgressDisplay(task, progressTicker, progressStopChan)
		defer func() {
			progressTicker.Stop()
			close(progressStopChan)
		}()
	}

	// 3. 遍历公司→日期→订单→步骤执行
	for companyIndex, company := range task.Companies {
		for dateIndex, dateTask := range company.Dates {
			// 3.1 执行该日期下的所有订单步骤
			for orderIndex, orderTask := range dateTask.Orders {
				for stepIndex, step := range orderTask.Steps {
					// 跳过已完成的步骤
					if step.Status == dto.StepStatusCompleted {
						continue
					}

					// 执行订单步骤
					err := m.executeOrderStep(
						task,
						companyIndex,
						dateIndex,
						orderIndex,
						stepIndex,
						company.CompanyUuid,
						orderTask.SaleOrderUuid,
						step.Name,
						taskFilePath,
					)
					if err != nil {
						// 错误已记录到步骤中，继续执行下一个步骤
						continue
					}
				}
			}

			// 3.2 检查该日期下所有订单的所有步骤是否都已完成（completed或failed）
			if m.isAllOrderStepsCompleted(task, companyIndex, dateIndex) {
				// 执行日期级别步骤
				if dateTask.DateStep.Status != dto.StepStatusCompleted {
					err := m.executeDateStep(
						task,
						companyIndex,
						dateIndex,
						company.CompanyUuid,
						dateTask.Date,
						taskFilePath,
					)
					if err != nil {
						// 错误已记录到步骤中，继续执行下一个日期
						continue
					}
				}
			}
		}
	}

	return nil
}

// executeOrderStep 执行订单级别步骤
func (m *batchRegenerateTaskManager) executeOrderStep(
	task *dto.BatchRegenerateTask,
	companyIndex int,
	dateIndex int,
	orderIndex int,
	stepIndex int,
	companyUuid uint64,
	saleOrderUuid uint64,
	stepName string,
	taskFilePath string,
) error {
	// 更新步骤状态为 running
	m.fileLock.Lock()
	err := m.UpdateStepStatus(task, companyIndex, dateIndex, orderIndex, stepIndex, false, dto.StepStatusRunning, "")
	if err == nil {
		m.SaveTaskList(task, taskFilePath) // 保存状态
	}
	m.fileLock.Unlock()
	if err != nil {
		return errors.WithMessage(err, "更新步骤状态失败")
	}

	// 执行步骤
	var stepErr error
	ctx := gin.Context{} // 命令行环境，ctx可以为nil
	switch stepName {
	case "regenerate-order-material":
		_, stepErr = m.salesOutboundSummarySrv.RegenerateOrderMaterial(&ctx, companyUuid, saleOrderUuid)
	case "regenerate-sale-order-material-outbound":
		_, stepErr = m.salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound(&ctx, companyUuid, saleOrderUuid)
	case "regenerate-order-pos-invoice":
		_, stepErr = m.salesOutboundSummarySrv.RegenerateOrderPosInvoice(&ctx, companyUuid, saleOrderUuid, m.openPosEntryName)
	default:
		stepErr = errors.New(fmt.Sprintf("未知的步骤名称: %s", stepName))
	}

	// 更新步骤状态
	m.fileLock.Lock()
	if stepErr != nil {
		m.UpdateStepStatus(task, companyIndex, dateIndex, orderIndex, stepIndex, false, dto.StepStatusFailed, stepErr.Error())
	} else {
		m.UpdateStepStatus(task, companyIndex, dateIndex, orderIndex, stepIndex, false, dto.StepStatusCompleted, "")
	}
	m.SaveTaskList(task, taskFilePath) // 保存状态
	m.fileLock.Unlock()

	// time.Sleep(10 * time.Second)

	return stepErr
}

// executeDateStep 执行日期级别步骤
func (m *batchRegenerateTaskManager) executeDateStep(
	task *dto.BatchRegenerateTask,
	companyIndex int,
	dateIndex int,
	companyUuid uint64,
	date string,
	taskFilePath string,
) error {
	// 更新步骤状态为 running
	m.fileLock.Lock()
	err := m.UpdateStepStatus(task, companyIndex, dateIndex, -1, -1, true, dto.StepStatusRunning, "")
	if err == nil {
		m.SaveTaskList(task, taskFilePath) // 保存状态
	}
	m.fileLock.Unlock()
	if err != nil {
		return errors.WithMessage(err, "更新步骤状态失败")
	}

	// 执行日期级别步骤
	ctx := gin.Context{} // 命令行环境，ctx可以为nil
	_, stepErr := m.salesOutboundSummarySrv.RegenerateSalesOutboundSummary(&ctx, companyUuid, date)

	// 更新步骤状态
	m.fileLock.Lock()
	if stepErr != nil {
		m.UpdateStepStatus(task, companyIndex, dateIndex, -1, -1, true, dto.StepStatusFailed, stepErr.Error())
	} else {
		m.UpdateStepStatus(task, companyIndex, dateIndex, -1, -1, true, dto.StepStatusCompleted, "")
	}
	m.SaveTaskList(task, taskFilePath) // 保存状态
	m.fileLock.Unlock()

	return stepErr
}

// isAllOrderStepsCompleted 检查该日期下所有订单的所有步骤是否都已完成（completed或failed）
func (m *batchRegenerateTaskManager) isAllOrderStepsCompleted(task *dto.BatchRegenerateTask, companyIndex int, dateIndex int) bool {
	if companyIndex < 0 || companyIndex >= len(task.Companies) {
		return false
	}
	if dateIndex < 0 || dateIndex >= len(task.Companies[companyIndex].Dates) {
		return false
	}

	dateTask := task.Companies[companyIndex].Dates[dateIndex]
	for _, orderTask := range dateTask.Orders {
		for _, step := range orderTask.Steps {
			// 如果还有pending或running状态的步骤，说明未全部完成
			if step.Status == dto.StepStatusPending || step.Status == dto.StepStatusRunning {
				return false
			}
		}
	}

	return true
}

// startProgressDisplay 启动进度显示（异步）
func (m *batchRegenerateTaskManager) startProgressDisplay(
	task *dto.BatchRegenerateTask,
	ticker *time.Ticker,
	stopChan chan bool,
) {
	for {
		select {
		case <-ticker.C:
			progress, err := m.GetProgress(task)
			if err == nil {
				// TODO: 格式化输出进度信息（在命令行工具中实现）
				_ = progress
			}
		case <-stopChan:
			return
		}
	}
}

// GetProgress 获取进度信息
func (m *batchRegenerateTaskManager) GetProgress(task *dto.BatchRegenerateTask) (*resp.BatchRegenerateProgressResp, error) {
	summary := m.calculateSummary(task)

	// 计算总体完成百分比
	totalSteps := summary.CompletedSteps + summary.FailedSteps + summary.PendingSteps
	var overallProgress float64
	if totalSteps > 0 {
		overallProgress = float64(summary.CompletedSteps) / float64(totalSteps) * 100
	}

	// 计算预计剩余时间（基于已完成步骤的平均耗时）
	estimatedTimeLeft := m.calculateEstimatedTimeLeft(task, summary.CompletedSteps)

	// 计算公司级别进度
	companyProgress := m.calculateCompanyProgress(task)

	// 计算日期级别进度（当前公司）
	dateProgress := m.calculateDateProgress(task, companyProgress.CurrentCompany)

	// 计算订单级别进度（当前日期）
	orderProgress := m.calculateOrderProgress(task, companyProgress.CurrentCompany, dateProgress.CurrentDate)

	// 查找当前正在执行的步骤
	currentStep := m.findCurrentStep(task)

	return &resp.BatchRegenerateProgressResp{
		OverallProgress:   overallProgress,
		CompletedSteps:    summary.CompletedSteps,
		FailedSteps:       summary.FailedSteps,
		PendingSteps:      summary.PendingSteps,
		EstimatedTimeLeft: estimatedTimeLeft,
		CurrentStep:       currentStep,
		CompanyProgress:   companyProgress,
		DateProgress:      dateProgress,
		OrderProgress:     orderProgress,
	}, nil
}

// calculateEstimatedTimeLeft 计算预计剩余时间
func (m *batchRegenerateTaskManager) calculateEstimatedTimeLeft(task *dto.BatchRegenerateTask, completedSteps int) string {
	if completedSteps == 0 {
		return "计算中..."
	}

	// 收集所有已完成步骤的耗时
	var totalDuration time.Duration
	var stepCount int

	for _, company := range task.Companies {
		for _, dateTask := range company.Dates {
			// 日期级别步骤
			if dateTask.DateStep.Status == dto.StepStatusCompleted && dateTask.DateStep.StartTime != "" && dateTask.DateStep.EndTime != "" {
				startTime, err1 := time.Parse(time.RFC3339, dateTask.DateStep.StartTime)
				endTime, err2 := time.Parse(time.RFC3339, dateTask.DateStep.EndTime)
				if err1 == nil && err2 == nil {
					totalDuration += endTime.Sub(startTime)
					stepCount++
				}
			}

			// 订单级别步骤
			for _, orderTask := range dateTask.Orders {
				for _, step := range orderTask.Steps {
					if step.Status == dto.StepStatusCompleted && step.StartTime != "" && step.EndTime != "" {
						startTime, err1 := time.Parse(time.RFC3339, step.StartTime)
						endTime, err2 := time.Parse(time.RFC3339, step.EndTime)
						if err1 == nil && err2 == nil {
							totalDuration += endTime.Sub(startTime)
							stepCount++
						}
					}
				}
			}
		}
	}

	if stepCount == 0 {
		return "计算中..."
	}

	// 计算平均耗时
	avgDuration := totalDuration / time.Duration(stepCount)

	// 计算预计剩余时间
	summary := m.calculateSummary(task)
	remainingSteps := summary.PendingSteps
	if remainingSteps == 0 {
		return "即将完成"
	}

	estimatedDuration := avgDuration * time.Duration(remainingSteps)
	estimatedMinutes := int(estimatedDuration.Minutes())
	if estimatedMinutes < 1 {
		return "约1分钟"
	}
	return fmt.Sprintf("约%d分钟", estimatedMinutes)
}

// calculateCompanyProgress 计算公司级别进度
func (m *batchRegenerateTaskManager) calculateCompanyProgress(task *dto.BatchRegenerateTask) resp.CompanyProgress {
	progress := resp.CompanyProgress{
		TotalCompanies:     len(task.Companies),
		PendingCompanies:   0,
		CompletedCompanies: make([]string, 0),
		CurrentCompany:     "",
	}

	var currentCompanyIndex = -1

	for i, company := range task.Companies {
		// 检查公司是否完成（所有日期和订单的所有步骤都完成）
		isCompleted := true
		hasRunning := false

		for _, dateTask := range company.Dates {
			// 检查日期级别步骤
			if dateTask.DateStep.Status == dto.StepStatusRunning {
				hasRunning = true
				isCompleted = false
			} else if dateTask.DateStep.Status != dto.StepStatusCompleted {
				isCompleted = false
			}

			// 检查订单步骤
			for _, orderTask := range dateTask.Orders {
				for _, step := range orderTask.Steps {
					if step.Status == dto.StepStatusRunning {
						hasRunning = true
						isCompleted = false
					} else if step.Status != dto.StepStatusCompleted && step.Status != dto.StepStatusFailed {
						isCompleted = false
					}
				}
			}
		}

		if isCompleted {
			progress.CompletedCompanies = append(progress.CompletedCompanies, company.CompanyName)
		} else {
			progress.PendingCompanies++
			if hasRunning && currentCompanyIndex == -1 {
				currentCompanyIndex = i
				progress.CurrentCompany = company.CompanyName
			}
		}
	}

	// 如果没有正在处理的，找第一个未完成的
	if progress.CurrentCompany == "" && progress.PendingCompanies > 0 {
		for _, company := range task.Companies {
			isCompleted := false
			for _, completedName := range progress.CompletedCompanies {
				if company.CompanyName == completedName {
					isCompleted = true
					break
				}
			}
			if !isCompleted {
				progress.CurrentCompany = company.CompanyName
				break
			}
		}
	}

	return progress
}

// calculateDateProgress 计算日期级别进度（当前公司）
func (m *batchRegenerateTaskManager) calculateDateProgress(task *dto.BatchRegenerateTask, currentCompanyName string) resp.DateProgress {
	progress := resp.DateProgress{
		TotalDates:     0,
		PendingDates:   0,
		CompletedDates: make([]string, 0),
		CurrentDate:    "",
	}

	// 找到当前公司
	var currentCompany *dto.CompanyTask
	for i := range task.Companies {
		if task.Companies[i].CompanyName == currentCompanyName {
			currentCompany = &task.Companies[i]
			break
		}
	}

	if currentCompany == nil {
		return progress
	}

	progress.TotalDates = len(currentCompany.Dates)
	var currentDateIndex = -1

	for i, dateTask := range currentCompany.Dates {
		// 检查日期是否完成（所有订单步骤完成 + 日期级别步骤完成）
		isCompleted := true
		hasRunning := false

		// 检查订单步骤
		for _, orderTask := range dateTask.Orders {
			for _, step := range orderTask.Steps {
				if step.Status == dto.StepStatusRunning {
					hasRunning = true
					isCompleted = false
				} else if step.Status != dto.StepStatusCompleted && step.Status != dto.StepStatusFailed {
					isCompleted = false
				}
			}
		}

		// 检查日期级别步骤
		if dateTask.DateStep.Status == dto.StepStatusRunning {
			hasRunning = true
			isCompleted = false
		} else if dateTask.DateStep.Status != dto.StepStatusCompleted {
			isCompleted = false
		}

		if isCompleted {
			progress.CompletedDates = append(progress.CompletedDates, dateTask.Date)
		} else {
			progress.PendingDates++
			if hasRunning && currentDateIndex == -1 {
				currentDateIndex = i
				progress.CurrentDate = dateTask.Date
			}
		}
	}

	// 如果没有正在处理的，找第一个未完成的
	if progress.CurrentDate == "" && progress.PendingDates > 0 {
		for _, dateTask := range currentCompany.Dates {
			isCompleted := false
			for _, completedDate := range progress.CompletedDates {
				if dateTask.Date == completedDate {
					isCompleted = true
					break
				}
			}
			if !isCompleted {
				progress.CurrentDate = dateTask.Date
				break
			}
		}
	}

	return progress
}

// calculateOrderProgress 计算订单级别进度（当前日期）
func (m *batchRegenerateTaskManager) calculateOrderProgress(task *dto.BatchRegenerateTask, currentCompanyName string, currentDate string) resp.OrderProgress {
	progress := resp.OrderProgress{
		TotalOrders:     0,
		PendingOrders:   0,
		CompletedOrders: make([]string, 0),
		CurrentOrder:    "",
	}

	// 找到当前公司和日期
	var currentDateTask *dto.DateTask
	for i := range task.Companies {
		if task.Companies[i].CompanyName == currentCompanyName {
			for j := range task.Companies[i].Dates {
				if task.Companies[i].Dates[j].Date == currentDate {
					currentDateTask = &task.Companies[i].Dates[j]
					break
				}
			}
			break
		}
	}

	if currentDateTask == nil {
		return progress
	}

	progress.TotalOrders = len(currentDateTask.Orders)
	var currentOrderIndex = -1

	for i, orderTask := range currentDateTask.Orders {
		// 检查订单是否完成（所有步骤完成）
		isCompleted := true
		hasRunning := false

		for _, step := range orderTask.Steps {
			if step.Status == dto.StepStatusRunning {
				hasRunning = true
				isCompleted = false
			} else if step.Status != dto.StepStatusCompleted && step.Status != dto.StepStatusFailed {
				isCompleted = false
			}
		}

		if isCompleted {
			progress.CompletedOrders = append(progress.CompletedOrders, orderTask.OrderNo)
		} else {
			progress.PendingOrders++
			if hasRunning && currentOrderIndex == -1 {
				currentOrderIndex = i
				progress.CurrentOrder = orderTask.OrderNo
			}
		}
	}

	// 如果没有正在处理的，找第一个未完成的
	if progress.CurrentOrder == "" && progress.PendingOrders > 0 {
		for _, orderTask := range currentDateTask.Orders {
			isCompleted := false
			for _, completedOrder := range progress.CompletedOrders {
				if orderTask.OrderNo == completedOrder {
					isCompleted = true
					break
				}
			}
			if !isCompleted {
				progress.CurrentOrder = orderTask.OrderNo
				break
			}
		}
	}

	return progress
}

// findCurrentStep 查找当前正在执行的步骤
// 优先查找 running 状态的步骤，如果没有则返回下一个 pending 状态的步骤
// 注意：执行顺序是先订单级别步骤，后日期级别步骤，所以查找时也按此顺序
func (m *batchRegenerateTaskManager) findCurrentStep(task *dto.BatchRegenerateTask) resp.CurrentStep {
	currentStep := resp.CurrentStep{}
	var nextPendingStep resp.CurrentStep
	foundNextPending := false

	// 遍历查找状态为 running 的步骤，同时记录第一个 pending 步骤
	// 按照执行顺序：先订单级别步骤，后日期级别步骤
	for _, company := range task.Companies {
		for _, dateTask := range company.Dates {
			// 先检查订单级别步骤（优先级更高）
			for _, orderTask := range dateTask.Orders {
				for _, step := range orderTask.Steps {
					if step.Status == dto.StepStatusRunning {
						currentStep.StepName = step.Name
						currentStep.StepType = "order"
						currentStep.CompanyName = company.CompanyName
						currentStep.Date = dateTask.Date
						currentStep.OrderNo = orderTask.OrderNo
						currentStep.SaleOrderUuid = orderTask.SaleOrderUuid
						return currentStep
					}
					// 如果没有找到 running 的，记录第一个 pending 的订单级别步骤
					if !foundNextPending && step.Status == dto.StepStatusPending {
						nextPendingStep.StepName = step.Name
						nextPendingStep.StepType = "order"
						nextPendingStep.CompanyName = company.CompanyName
						nextPendingStep.Date = dateTask.Date
						nextPendingStep.OrderNo = orderTask.OrderNo
						nextPendingStep.SaleOrderUuid = orderTask.SaleOrderUuid
						foundNextPending = true
					}
				}
			}

			// 再检查日期级别步骤
			if dateTask.DateStep.Status == dto.StepStatusRunning {
				currentStep.StepName = dateTask.DateStep.Name
				currentStep.StepType = "date"
				currentStep.CompanyName = company.CompanyName
				currentStep.Date = dateTask.Date
				return currentStep
			}
			// 如果没有找到 running 的，且还没有找到 pending 的订单级别步骤，记录第一个 pending 的日期级别步骤
			if !foundNextPending && dateTask.DateStep.Status == dto.StepStatusPending {
				nextPendingStep.StepName = dateTask.DateStep.Name
				nextPendingStep.StepType = "date"
				nextPendingStep.CompanyName = company.CompanyName
				nextPendingStep.Date = dateTask.Date
				foundNextPending = true
			}
		}
	}

	// 如果没有找到 running 的步骤，返回下一个 pending 的步骤
	if foundNextPending {
		return nextPendingStep
	}

	return currentStep
}
