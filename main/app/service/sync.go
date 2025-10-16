package service

import (
	"fmt"
	"runtime/debug"
	"slices"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
)

// ISyncSrv同步服务接口
type ISyncSrv interface {
	Sync(ctx context.Context, syncReq req.SyncReq) (resp.SyncResp, error)                                  // 同步
	GetTaskList(ctx context.Context, listReq req.SyncTaskListReq) (resp.SyncTaskListPaginationResp, error) // 获取同步任务列表
	GetTaskDetail(ctx context.Context, detailReq req.SyncTaskDetailReq) (resp.SyncTaskDetailResp, error)   // 获取同步任务详情
}

// syncTaskConfig 同步任务配置
type syncTaskConfig struct {
	TaskType string
	TaskName string
	Executor func(context.Context) error
}

// SyncSrv同步服务结构体
type SyncSrv struct {
	dbm          *database.DBManager
	warehouseSrv IWarehouseSrv
	materialSrv  IMaterialSrv
	supplierSrv  ISupplierSrv
	productSrv   IProductSrv
}

// 全局同步任务管理器
var (
	syncTaskManager = &SyncTaskManager{
		runningTasks: sync.Map{},
	}
)

// SyncTaskManager 同步任务管理器
type SyncTaskManager struct {
	runningTasks sync.Map // key: companyUuid, value: bool
}

// tryStartTask 尝试启动任务，如果已有任务在运行则返回false
func (m *SyncTaskManager) tryStartTask(companyUuid uint64) bool {
	_, loaded := m.runningTasks.LoadOrStore(companyUuid, true)
	return !loaded // 如果之前没有值则返回true，表示成功启动任务
}

// finishTask 完成任务
func (m *SyncTaskManager) finishTask(companyUuid uint64) {
	m.runningTasks.Delete(companyUuid)
}

// getRunningCompanyUuids 获取当前正在执行同步任务的所有companyUuid
func (m *SyncTaskManager) getRunningCompanyUuids() []uint64 {
	var companyUuids []uint64
	m.runningTasks.Range(func(key, value any) bool {
		if companyUuid, ok := key.(uint64); ok {
			companyUuids = append(companyUuids, companyUuid)
		}
		return true
	})
	return companyUuids
}

// NewSyncSrv 创建新同步服务
func NewSyncSrv(dbm *database.DBManager, warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv, materialSrv IMaterialSrv) ISyncSrv {
	return NewSyncSrvImpl(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv)
}

// NewSyncSrvImpl 创建新同步服务实现
func NewSyncSrvImpl(dbm *database.DBManager, warehouseSrv IWarehouseSrv, supplierSrv ISupplierSrv, productSrv IProductSrv, materialSrv IMaterialSrv) ISyncSrv {
	return &SyncSrv{
		dbm:          dbm,
		warehouseSrv: warehouseSrv,
		materialSrv:  materialSrv,
		supplierSrv:  supplierSrv,
		productSrv:   productSrv,
	}
}

// Sync 同步
func (s *SyncSrv) Sync(ctx context.Context, syncReq req.SyncReq) (resp.SyncResp, error) {
	company := ctx.GetCompany()
	companyUuid := company.Uuid

	// 检查是否已有同步任务在运行
	if !syncTaskManager.tryStartTask(companyUuid) {
		return resp.SyncResp{}, errors.New("数据同步中，请稍后再试")
	}

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	var syncTask *model.SyncTask
	var retryMode bool          // 是否为重试模式
	var retryTaskTypes []string // 需要重试的任务类型

	// 定义所有同步任务配置
	allTasks := []syncTaskConfig{
		{constant.SyncTaskTypeProductCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeProductCategory], s.productSrv.SyncProductShopCategory},
		{constant.SyncTaskTypeMaterialCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterialCategory], s.materialSrv.SyncMaterialCategory},
		{constant.SyncTaskTypeTax, constant.SyncTaskTypeNames[constant.SyncTaskTypeTax], s.productSrv.SyncProductTax},
		{constant.SyncTaskTypeUnit, constant.SyncTaskTypeNames[constant.SyncTaskTypeUnit], s.productSrv.SyncUnit},
		{constant.SyncTaskTypeWarehouse, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouse], s.warehouseSrv.SyncWarehouse},
		{constant.SyncTaskTypeMaterial, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterial], s.materialSrv.SyncMaterial},
		{constant.SyncTaskTypeFlavor, constant.SyncTaskTypeNames[constant.SyncTaskTypeFlavor], s.productSrv.SyncProductFlavor},
		{constant.SyncTaskTypeAttribute, constant.SyncTaskTypeNames[constant.SyncTaskTypeAttribute], s.productSrv.SyncAttributeGroup},
		{constant.SyncTaskTypeSauce, constant.SyncTaskTypeNames[constant.SyncTaskTypeSauce], s.productSrv.SyncSauce},
		{constant.SyncTaskTypeProduct, constant.SyncTaskTypeNames[constant.SyncTaskTypeProduct], s.productSrv.SyncProduct},
		{constant.SyncTaskTypeBomCard, constant.SyncTaskTypeNames[constant.SyncTaskTypeBomCard], s.materialSrv.SyncProductBomCard},
		{constant.SyncTaskTypeSupplier, constant.SyncTaskTypeNames[constant.SyncTaskTypeSupplier], s.supplierSrv.SyncSupplier},
		{constant.SyncTaskTypeWarehouseStock, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouseStock], s.warehouseSrv.SyncWarehouseItemStock},
		{constant.SyncTaskTypeProductStock, constant.SyncTaskTypeNames[constant.SyncTaskTypeProductStock], s.productSrv.SyncProductStockByBomCard},
		{constant.SyncTaskTypePackageImage, constant.SyncTaskTypeNames[constant.SyncTaskTypePackageImage], s.productSrv.SyncProductPackageImage},
	}

	// 如果传递了任务UUID，则为重试模式
	if syncReq.TaskUuid > 0 {
		retryMode = true
		// 查询原任务
		existTask, err := syncTaskRepo.GetByUuid(syncReq.TaskUuid, syncTaskRepo.PreloadItems())
		if err != nil {
			syncTaskManager.finishTask(companyUuid)
			return resp.SyncResp{}, errors.WithMessage(err, "查询同步任务失败")
		}

		// 获取失败的任务类型
		for _, item := range existTask.Items {
			if item.Status == constant.SyncTaskItemStatusFailed {
				retryTaskTypes = append(retryTaskTypes, item.TaskType)
			}
		}

		if len(retryTaskTypes) == 0 {
			syncTaskManager.finishTask(companyUuid)
			return resp.SyncResp{}, errors.New("没有需要重试的任务")
		}

		syncTask = existTask
	} else {
		// 创建新的同步任务
		syncTask = &model.SyncTask{
			Status:       constant.SyncTaskStatusRunning,
			TotalCount:   uint32(len(allTasks)),
			SuccessCount: 0,
			FailCount:    0,
			StartTime:    time.Now().Unix(),
		}

		if err := syncTaskRepo.Create(syncTask); err != nil {
			syncTaskManager.finishTask(companyUuid)
			return resp.SyncResp{}, errors.WithMessage(err, "创建同步任务失败")
		}
	}

	// 启动异步同步任务
	if syncReq.IsSyncExecute {
		s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
	} else {
		go s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
	}

	message := "数据同步已启动"
	if retryMode {
		message = "重试同步任务已启动"
	}

	return resp.SyncResp{
		TaskUuid: syncTask.Uuid,
		Message:  message,
	}, nil
}

// executeSync 执行同步任务
func (s *SyncSrv) executeSync(ctx context.Context, syncTask *model.SyncTask, allTasks []syncTaskConfig, retryMode bool, retryTaskTypes []string) {
	company := ctx.GetCompany()
	companyUuid := company.Uuid

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	var successCount uint32
	var failCount uint32

	// 确保任务完成时清理状态
	defer func() {
		var isPanicOccurred bool
		if r := recover(); r != nil {
			// 获取堆栈
			stack := string(debug.Stack())
			logger.Logger.Error("同步任务发生panic", zap.Uint64("companyUuid", companyUuid), zap.Any("panic", r), zap.String("stack", stack))
			// 更新任务状态为失败
			syncTaskRepo.Update(syncTask.Uuid, map[string]any{
				"status":   constant.SyncTaskStatusFailed,
				"panic":    fmt.Sprintf("%v: %s", r, stack),
				"end_time": time.Now().Unix(),
			})
			isPanicOccurred = true
		}

		// 发生异常，包括失败和panic
		isExceptionOccurred := failCount > 0 || isPanicOccurred

		syncTaskManager.finishTask(companyUuid)
		logger.Logger.Info("同步任务完成", zap.Uint64("companyUuid", companyUuid),
			zap.Uint32("successCount", successCount),
			zap.Uint32("failCount", failCount))

		lastSyncTime := time.Now().Unix()
		// 未报错才记录上次同步完成时间
		if !isExceptionOccurred {
			s.dbm.GetDB(companyUuid).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
			s.dbm.GetDB(0).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
		}
		// 推送websocket
		go websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
			"task_uuid":             syncTask.Uuid,
			"is_exception_occurred": isExceptionOccurred,
			"sync_time":             time.Now().Unix(),
		})
	}()

	logger.Logger.Info("开始执行同步任务", zap.Uint64("companyUuid", companyUuid), zap.Uint64("taskUuid", syncTask.Uuid))

	// 如果是重试模式，只执行失败的任务
	tasksToExecute := allTasks
	if retryMode {
		tasksToExecute = []syncTaskConfig{}
		for _, task := range allTasks {
			if slices.Contains(retryTaskTypes, task.TaskType) {
				tasksToExecute = append(tasksToExecute, task)
			}
		}
	}

	// 执行每个子任务
	for _, taskCfg := range tasksToExecute {
		s.executeSyncTask(ctx, syncTask.Uuid, taskCfg, &successCount, &failCount, retryMode)
	}

	// 更新主任务状态
	endTime := time.Now().Unix()
	finalStatus := constant.SyncTaskStatusSuccess
	if failCount > 0 {
		finalStatus = constant.SyncTaskStatusFailed
	}

	err := syncTaskRepo.Update(syncTask.Uuid, map[string]any{
		"status":        finalStatus,
		"success_count": successCount,
		"fail_count":    failCount,
		"end_time":      endTime,
	})
	if err != nil {
		logger.Logger.Error("更新同步任务状态失败", zap.Error(err))
	}

	// 更新公司的最后同步时间
	if finalStatus == constant.SyncTaskStatusSuccess {
		s.dbm.GetDB(companyUuid).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", endTime)
		s.dbm.GetDB(0).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", endTime)
	}
}

// executeSyncTask 执行单个同步任务
func (s *SyncSrv) executeSyncTask(ctx context.Context, syncTaskUuid uint64, taskCfg syncTaskConfig, successCount, failCount *uint32, retryMode bool) {
	// 实例化repo（同步任务表在公司库）
	syncTaskItemRepo := repository.NewSyncTaskItemRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	var taskItem *model.SyncTaskItem

	// 如果是重试模式，查找已存在的任务项并更新
	if retryMode {
		items, err := syncTaskItemRepo.GetList(
			syncTaskItemRepo.WhereSyncTaskUuid(syncTaskUuid),
			syncTaskItemRepo.WhereTaskType(taskCfg.TaskType),
		)
		if err == nil && len(items) > 0 {
			taskItem = &items[0]
			// 重置任务项状态
			syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
				"status":        constant.SyncTaskItemStatusRunning,
				"error_message": "",
				"start_time":    time.Now().Unix(),
				"end_time":      0,
			})
		}
	}

	// 如果不是重试或找不到已有任务项，创建新任务项
	if taskItem == nil {
		taskItem = &model.SyncTaskItem{
			SyncTaskUuid: syncTaskUuid,
			TaskType:     taskCfg.TaskType,
			TaskName:     taskCfg.TaskName,
			Status:       constant.SyncTaskItemStatusRunning,
			StartTime:    time.Now().Unix(),
		}

		if err := syncTaskItemRepo.Create(taskItem); err != nil {
			logger.Logger.Error("创建同步任务明细失败", zap.String("taskType", taskCfg.TaskType), zap.Error(err))
			return
		}
	}

	logger.Logger.Info("开始同步", zap.String("taskName", taskCfg.TaskName))

	// 执行同步任务
	err := taskCfg.Executor(ctx)
	endTime := time.Now().Unix()

	if err != nil {
		// 任务失败
		logger.Logger.Error("同步失败", zap.String("taskName", taskCfg.TaskName), zap.Error(err))
		*failCount++

		syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
			"status":        constant.SyncTaskItemStatusFailed,
			"error_message": err.Error(),
			"end_time":      endTime,
		})
	} else {
		// 任务成功
		logger.Logger.Info("同步成功", zap.String("taskName", taskCfg.TaskName))
		*successCount++

		syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
			"status":        constant.SyncTaskItemStatusSuccess,
			"error_message": "",
			"end_time":      endTime,
		})
	}
}

// GetTaskList 获取同步任务列表
func (s *SyncSrv) GetTaskList(ctx context.Context, listReq req.SyncTaskListReq) (resp.SyncTaskListPaginationResp, error) {
	companyUuid := ctx.GetCompanyUuid()

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	// 构建查询选项
	opts := []repository.DBOption{
		syncTaskRepo.OrderByCreateTime(true),
	}

	if listReq.Status != nil {
		opts = append(opts, syncTaskRepo.WhereStatus(*listReq.Status))
	}

	// 分页查询
	tasks, total, err := syncTaskRepo.GetListWithPagination(listReq.PageReq.PageNo, listReq.PageReq.PageSize, opts...)
	if err != nil {
		return resp.SyncTaskListPaginationResp{}, errors.WithMessage(err, "查询同步任务列表失败")
	}

	// 转换为响应格式
	list := make([]resp.SyncTaskListResp, 0, len(tasks))
	for _, task := range tasks {
		list = append(list, convertToSyncTaskListResp(task))
	}

	return resp.SyncTaskListPaginationResp{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageReq.PageNo,
			PageSize: listReq.PageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetTaskDetail 获取同步任务详情
func (s *SyncSrv) GetTaskDetail(ctx context.Context, detailReq req.SyncTaskDetailReq) (resp.SyncTaskDetailResp, error) {
	companyUuid := ctx.GetCompanyUuid()

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	// 查询任务
	task, err := syncTaskRepo.GetByUuid(detailReq.TaskUuid, syncTaskRepo.PreloadItems())
	if err != nil {
		return resp.SyncTaskDetailResp{}, errors.WithMessage(err, "查询同步任务失败")
	}

	// 转换为响应格式
	return convertToSyncTaskDetailResp(*task), nil
}

// convertToSyncTaskItemResp 转换为同步任务明细响应
func convertToSyncTaskItemResp(item model.SyncTaskItem) resp.SyncTaskItemResp {
	duration := int64(0)
	if item.EndTime > 0 && item.StartTime > 0 {
		duration = item.EndTime - item.StartTime
	}

	return resp.SyncTaskItemResp{
		Uuid:         item.Uuid,
		TaskType:     item.TaskType,
		TaskName:     item.TaskName,
		Status:       item.Status,
		ErrorMessage: item.ErrorMessage,
		StartTime:    item.StartTime,
		EndTime:      item.EndTime,
		Duration:     duration,
	}
}

// convertToSyncTaskDetailResp 转换为同步任务详情响应
func convertToSyncTaskDetailResp(task model.SyncTask) resp.SyncTaskDetailResp {
	duration := int64(0)
	if task.EndTime > 0 && task.StartTime > 0 {
		duration = task.EndTime - task.StartTime
	}

	items := make([]resp.SyncTaskItemResp, 0, len(task.Items))
	for _, item := range task.Items {
		items = append(items, convertToSyncTaskItemResp(item))
	}

	return resp.SyncTaskDetailResp{
		Uuid:         task.Uuid,
		Status:       task.Status,
		TotalCount:   task.TotalCount,
		SuccessCount: task.SuccessCount,
		FailCount:    task.FailCount,
		StartTime:    task.StartTime,
		EndTime:      task.EndTime,
		Duration:     duration,
		CreateTime:   task.CreateTime,
		Items:        items,
	}
}

// convertToSyncTaskListResp 转换为同步任务列表响应
func convertToSyncTaskListResp(task model.SyncTask) resp.SyncTaskListResp {
	duration := int64(0)
	if task.EndTime > 0 && task.StartTime > 0 {
		duration = task.EndTime - task.StartTime
	}

	return resp.SyncTaskListResp{
		Uuid:         task.Uuid,
		Status:       task.Status,
		TotalCount:   task.TotalCount,
		SuccessCount: task.SuccessCount,
		FailCount:    task.FailCount,
		StartTime:    task.StartTime,
		EndTime:      task.EndTime,
		Duration:     duration,
		CreateTime:   task.CreateTime,
	}
}
