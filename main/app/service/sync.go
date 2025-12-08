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
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ISyncSrv同步服务接口
type ISyncSrv interface {
	Sync(ctx context.Context, syncReq req.SyncReq) (resp.SyncResp, error)                                  // 同步
	GetTaskList(ctx context.Context, listReq req.SyncTaskListReq) (resp.SyncTaskListPaginationResp, error) // 获取同步任务列表
	GetTaskDetail(ctx context.Context, detailReq req.SyncTaskDetailReq) (resp.SyncTaskDetailResp, error)   // 获取同步任务详情

	// 颗粒化同步接口
	GetHeadquartersDataList(ctx context.Context, listReq req.GetHeadquartersDataListReq) (resp.HeadquartersDataListResp, error) // 获取总部可同步数据列表
	GranularSync(ctx context.Context, syncReq req.GranularSyncReq) (resp.GranularSyncResp, error)                               // 颗粒化同步数据
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

	// 定义所有同步任务配置（保持原样，不添加营销数据类型）
	allTasks := []syncTaskConfig{
		{constant.SyncTaskTypeProductCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeProductCategory], func(ctx context.Context) error { return s.productSrv.SyncProductShopCategory(ctx, false, nil) }},
		{constant.SyncTaskTypeMaterialCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterialCategory], func(ctx context.Context) error { return s.materialSrv.SyncMaterialCategory(ctx, false, nil) }},
		{constant.SyncTaskTypeTax, constant.SyncTaskTypeNames[constant.SyncTaskTypeTax], func(ctx context.Context) error { return s.productSrv.SyncProductTax(ctx, false, nil) }}, // 无多语言数据
		{constant.SyncTaskTypeUnit, constant.SyncTaskTypeNames[constant.SyncTaskTypeUnit], func(ctx context.Context) error { return s.productSrv.SyncUnit(ctx, false, nil) }},
		{constant.SyncTaskTypeMaterial, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterial], func(ctx context.Context) error { return s.materialSrv.SyncMaterial(ctx, false, nil) }},
		{constant.SyncTaskTypeWarehouse, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouse], s.warehouseSrv.SyncWarehouse},
		{constant.SyncTaskTypeFlavor, constant.SyncTaskTypeNames[constant.SyncTaskTypeFlavor], func(ctx context.Context) error { return s.productSrv.SyncProductFlavor(ctx, false, nil) }},
		{constant.SyncTaskTypeAttribute, constant.SyncTaskTypeNames[constant.SyncTaskTypeAttribute], func(ctx context.Context) error { return s.productSrv.SyncAttributeGroup(ctx, false, nil) }},
		{constant.SyncTaskTypeSauce, constant.SyncTaskTypeNames[constant.SyncTaskTypeSauce], func(ctx context.Context) error { return s.productSrv.SyncSauce(ctx, false, nil) }},
		{constant.SyncTaskTypeProduct, constant.SyncTaskTypeNames[constant.SyncTaskTypeProduct], func(ctx context.Context) error { return s.productSrv.SyncProduct(ctx, false, nil) }},
		{constant.SyncTaskTypeBomCard, constant.SyncTaskTypeNames[constant.SyncTaskTypeBomCard], func(ctx context.Context) error { return s.materialSrv.SyncProductBomCard(ctx, false, nil) }},
		{constant.SyncTaskTypeSupplier, constant.SyncTaskTypeNames[constant.SyncTaskTypeSupplier], func(ctx context.Context) error { return s.supplierSrv.SyncSupplier(ctx, false, nil) }}, // 无多语言数据
		{constant.SyncTaskTypeWarehouseStock, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouseStock], s.warehouseSrv.SyncWarehouseItemStock},                                      // 无多语言数据
		{constant.SyncTaskTypeProductStock, constant.SyncTaskTypeNames[constant.SyncTaskTypeProductStock], s.productSrv.SyncProductStockByBomCard},                                         // 无多语言数据
		{constant.SyncTaskTypePackageImage, constant.SyncTaskTypeNames[constant.SyncTaskTypePackageImage], s.productSrv.SyncProductPackageImage},                                           // 无多语言数据
		{constant.SyncTaskTypeMultiLanguage, constant.SyncTaskTypeNames[constant.SyncTaskTypeMultiLanguage], s.SyncMultiLanguage},
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
		utils.Go(func() {
			s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
		})
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
			s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
		}
		// 推送websocket
		utils.Go(func() {
			websocket.PushClient(company.Uuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
				"task_uuid":             syncTask.Uuid,
				"is_exception_occurred": isExceptionOccurred,
				"sync_time":             time.Now().Unix(),
			})
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
		s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", endTime)
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

// SyncMultiLanguage 同步多语言数据
// 从总部同步所有类型（商品分类、物品分类、单位、仓库、物品、规格、属性、加料、商品、成本卡、商品套餐组）的多语言数据到子店
// 采用先删除后创建的策略，确保子店多语言数据与总部保持一致
func (s *SyncSrv) SyncMultiLanguage(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	// 只有子店才需要同步多语言
	if !companySetting.IsSubShop() {
		return nil
	}

	// 获取总部数据库
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)

	// 定义需要同步多语言的表和字段映射
	type tableConfig struct {
		tableName               string   // 表名
		multiLanguageUuidColumn string   // 多语言UUID字段名
		entityUuidColumn        string   // 实体UUID字段名
		preloadRelations        []string // 需要预加载的关联
		filterCondition         string   // 自定义筛选条件（可选，默认使用 headquarter_uuid = 0）
	}

	// 所有需要同步多语言的表配置（按表名字母顺序排列）
	// 注意：product_package_group 表没有 headquarter_uuid 字段，需要通过关联 product_package 表筛选
	tableConfigs := []tableConfig{
		{tableName: config.Database.TablePrefix + "material", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "material_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_attribute", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_attribute_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_bom_card", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_category", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_flavor", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_package", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_package", multiLanguageUuidColumn: "describe_multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_package_group", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid", filterCondition: "product_package_uuid IN (SELECT uuid FROM " + config.Database.TablePrefix + "product_package WHERE headquarter_uuid = 0)"},
		{tableName: config.Database.TablePrefix + "product_sauce", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "product_unit", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "warehouse", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "full_reduction_activity", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "marketing_activity", multiLanguageUuidColumn: "multi_language_name_uuid", entityUuidColumn: "uuid"},
		{tableName: config.Database.TablePrefix + "marketing_activity", multiLanguageUuidColumn: "multi_language_desc_uuid", entityUuidColumn: "uuid"},
	}

	// 收集所有需要同步的多语言UUID
	multiLanguageUuidMap := make(map[uint64]bool) // 用于去重

	// 从总部表中查询所有总部数据的多语言UUID
	for _, cfg := range tableConfigs {
		var records []map[string]any
		query := headquarterDB.Table(cfg.tableName).
			Select(cfg.multiLanguageUuidColumn).
			Where("delete_time = 0").
			Where(cfg.multiLanguageUuidColumn + " > 0")

		// 使用自定义筛选条件或默认条件（headquarter_uuid = 0）
		if cfg.filterCondition != "" {
			query = query.Where(cfg.filterCondition)
		} else {
			query = query.Where("headquarter_uuid = 0")
		}

		err := query.Find(&records).Error
		if err != nil {
			logger.Logger.Error("同步多语言-查询表失败",
				zap.String("table", cfg.tableName),
				zap.Error(err))
			continue
		}

		for _, record := range records {
			// 数据库返回的整数类型可能是 uint64 或 int64，需要分别处理
			var uuid uint64
			switch v := record[cfg.multiLanguageUuidColumn].(type) {
			case uint64:
				uuid = v
			case int64:
				uuid = uint64(v)
			}
			if uuid > 0 {
				multiLanguageUuidMap[uuid] = true
			}
		}
	}

	if len(multiLanguageUuidMap) == 0 {
		logger.Logger.Info("同步多语言-没有需要同步的多语言数据")
		return nil
	}

	// 将map转为切片
	multiLanguageUuids := make([]uint64, 0, len(multiLanguageUuidMap))
	for uuid := range multiLanguageUuidMap {
		multiLanguageUuids = append(multiLanguageUuids, uuid)
	}

	logger.Logger.Info("同步多语言-开始同步", zap.Int("count", len(multiLanguageUuids)))

	// 从总部查询所有多语言数据
	var headquarterMultiLanguages []model.MultiLanguageName
	err := headquarterDB.Model(&model.MultiLanguageName{}).
		Where("delete_time = 0").
		Where("uuid IN (?)", multiLanguageUuids).
		Find(&headquarterMultiLanguages).Error
	if err != nil {
		return errors.WithMessage(err, "获取总部多语言数据失败")
	}

	// 在事务中批量同步多语言数据：先删除，再创建
	err = subShopDB.Transaction(func(tx *gorm.DB) error {
		// 步骤1：删除子店中这些UUID对应的所有多语言记录
		if len(multiLanguageUuids) > 0 {
			err := tx.Where("uuid IN (?)", multiLanguageUuids).Delete(&model.MultiLanguageName{}).Error
			if err != nil {
				logger.Logger.Error("同步多语言-删除子店旧多语言数据失败", zap.Error(err))
				return errors.WithMessage(err, "删除子店旧多语言数据失败")
			}
			logger.Logger.Info("同步多语言-已删除子店旧多语言数据", zap.Int("count", len(multiLanguageUuids)))
		}

		// 步骤2：批量创建总部的多语言记录
		if len(headquarterMultiLanguages) > 0 {
			// 构建要插入的多语言记录列表
			insertMultiLanguages := make([]model.MultiLanguageName, 0, len(headquarterMultiLanguages))
			for _, hqMultiLanguage := range headquarterMultiLanguages {
				insertMultiLanguages = append(insertMultiLanguages, model.MultiLanguageName{
					BaseModel: model.BaseModel{
						Uuid:       hqMultiLanguage.Uuid,
						CreateTime: hqMultiLanguage.CreateTime,
						UpdateTime: hqMultiLanguage.UpdateTime,
						DeleteTime: hqMultiLanguage.DeleteTime,
					},
					EnName:   hqMultiLanguage.EnName,
					ZhName:   hqMultiLanguage.ZhName,
					ZhTwName: hqMultiLanguage.ZhTwName,
					ThName:   hqMultiLanguage.ThName,
					MyName:   hqMultiLanguage.MyName,
					JaName:   hqMultiLanguage.JaName,
					KoName:   hqMultiLanguage.KoName,
					TrName:   hqMultiLanguage.TrName,
					SvName:   hqMultiLanguage.SvName,
				})
			}

			// 批量插入多语言记录
			err := tx.Create(&insertMultiLanguages).Error
			if err != nil {
				logger.Logger.Error("同步多语言-批量创建多语言记录失败", zap.Error(err))
				return errors.WithMessage(err, "批量创建多语言记录失败")
			}
			logger.Logger.Info("同步多语言-已批量创建多语言记录", zap.Int("count", len(insertMultiLanguages)))
		}

		return nil
	})

	if err != nil {
		return errors.WithMessage(err, "同步多语言数据失败")
	}

	logger.Logger.Info("同步多语言-完成", zap.Int("count", len(headquarterMultiLanguages)))
	return nil
}

// SyncMarketingCouponByUuids 按uuid同步优惠券
func (s *SyncSrv) SyncMarketingCouponByUuids(ctx context.Context, uuids []uint64) error {

	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)

	// 标记删除未勾选的优惠券
	if err := subShopDB.Table("ttpos_marketing_coupon").
		Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).
		Where("uuid NOT IN (?)", uuids).
		Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "标记删除未勾选的优惠券失败")
	}

	// 查询总部优惠券
	var hqCoupons []model.MarketingCoupon
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Find(&hqCoupons).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部优惠券失败")
	}

	// 同步到分店（先删除再创建）
	for _, hqCoupon := range hqCoupons {
		// 删除分店中已有的该优惠券
		subShopDB.Unscoped().Where("uuid = ?", hqCoupon.Uuid).Delete(&model.MarketingCoupon{})

		// 创建新的优惠券（标记来源）
		newCoupon := hqCoupon
		newCoupon.HeadquarterUuid = companySetting.HeadquarterUuid
		newCoupon.ID = 0 // 重置ID，让数据库自动生成

		err = subShopDB.Create(&newCoupon).Error
		if err != nil {
			logger.Logger.Error("同步优惠券失败", zap.Uint64("uuid", hqCoupon.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncFullReductionByUuids 按uuid同步满额减活动
func (s *SyncSrv) SyncFullReductionByUuids(ctx context.Context, uuids []uint64) error {

	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)

	// 标记删除未勾选的满额减活动
	if err := subShopDB.Table("ttpos_full_reduction_activity").
		Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).
		Where("uuid NOT IN (?)", uuids).
		Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "标记删除未勾选的满额减活动失败")
	}

	// 查询总部满额减活动（包含规则）
	var hqActivities []model.FullReductionActivity
	err := headquarterDB.Preload("Rules").
		Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Find(&hqActivities).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部满额减活动失败")
	}

	// 同步到分店（先删除再创建）
	for _, hqActivity := range hqActivities {
		// 删除分店中已有的该活动（包括规则）
		subShopDB.Unscoped().Where("uuid = ?", hqActivity.Uuid).Delete(&model.FullReductionActivity{})
		subShopDB.Unscoped().Where("full_reduction_activity_uuid = ?", hqActivity.Uuid).Delete(&model.FullReductionActivityRule{})

		// 创建新的满额减活动
		newActivity := hqActivity
		newActivity.HeadquarterUuid = companySetting.HeadquarterUuid
		newActivity.ID = 0

		// 同步规则
		for i := range newActivity.Rules {
			newActivity.Rules[i].ID = 0
		}

		err = subShopDB.Create(&newActivity).Error
		if err != nil {
			logger.Logger.Error("同步满额减活动失败", zap.Uint64("uuid", hqActivity.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncProductLabelByUuids 按uuid同步菜品标签
func (s *SyncSrv) SyncProductLabelByUuids(ctx context.Context, uuids []uint64) error {

	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)

	// 标记删除未勾选的菜品标签
	if err := subShopDB.Table("ttpos_product_label").
		Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).
		Where("uuid NOT IN (?)", uuids).
		Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "标记删除未勾选的菜品标签失败")
	}

	// 查询总部菜品标签
	var hqLabels []model.ProductLabel
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Find(&hqLabels).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部菜品标签失败")
	}

	// 同步到分店（先删除再创建）
	for _, hqLabel := range hqLabels {
		subShopDB.Unscoped().Where("uuid = ?", hqLabel.Uuid).Delete(&model.ProductLabel{})

		newLabel := hqLabel
		newLabel.HeadquarterUuid = companySetting.HeadquarterUuid
		newLabel.ID = 0

		err = subShopDB.Create(&newLabel).Error
		if err != nil {
			logger.Logger.Error("同步菜品标签失败", zap.Uint64("uuid", hqLabel.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncMarketingActivityByUuids 按uuid同步营销活动
func (s *SyncSrv) SyncMarketingActivityByUuids(ctx context.Context, uuids []uint64) error {

	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)

	// 标记删除未勾选的营销活动
	if err := subShopDB.Table("ttpos_marketing_activity").
		Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).
		Where("uuid NOT IN (?)", uuids).
		Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "标记删除未勾选的营销活动失败")
	}

	// 查询总部营销活动（包含奖品）
	var hqActivities []model.MarketingActivity
	err := headquarterDB.Preload("Prizes").
		Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Find(&hqActivities).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部营销活动失败")
	}

	// 同步到分店（先删除再创建）
	for _, hqActivity := range hqActivities {
		// 删除分店中已有的该活动（包括奖品）
		subShopDB.Unscoped().Where("uuid = ?", hqActivity.Uuid).Delete(&model.MarketingActivity{})
		subShopDB.Unscoped().Where("activity_uuid = ?", hqActivity.Uuid).Delete(&model.MarketingActivityPrize{})

		// 创建新的营销活动
		newActivity := hqActivity
		newActivity.HeadquarterUuid = companySetting.HeadquarterUuid
		newActivity.ID = 0

		// 同步奖品
		for i := range newActivity.Prizes {
			newActivity.Prizes[i].ID = 0
		}

		err = subShopDB.Create(&newActivity).Error
		if err != nil {
			logger.Logger.Error("同步营销活动失败", zap.Uint64("uuid", hqActivity.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncPaymentMethodByUuids 按uuid同步支付方式（⚠️ 复杂规则）
func (s *SyncSrv) SyncPaymentMethodByUuids(ctx context.Context, uuids []uint64) error {

	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 查询总部支付方式（排除 code=40 和 code=10）
	var hqPayments []model.PaymentMethod
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Where("code NOT IN (?)", []int{model.PaymentMethodCash, model.PaymentMethodBalance}).
		Find(&hqPayments).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部支付方式失败")
	}

	// 特殊code列表（不跳过，只更新headquarter_uuid）
	specialCodes := map[int]bool{
		model.PaymentCodeLianlianWechat:      true, // 90111
		model.PaymentCodeLianlianAli:         true, // 90222
		model.PaymentCodeLianlianQrPromptPay: true, // 90333
	}

	for _, hqPayment := range hqPayments {
		// 检查分店是否已有同名支付方式（payment_name）
		var existPayment model.PaymentMethod
		err := subShopDB.Where("payment_name = ? AND delete_time = 0", hqPayment.PaymentName).
			First(&existPayment).Error

		if err == nil {
			// 分店已有同名支付方式
			if specialCodes[existPayment.Code] {
				// 特殊code：只更新 headquarter_uuid
				err = subShopDB.Model(&model.PaymentMethod{}).
					Where("id = ?", existPayment.ID).
					Update("headquarter_uuid", headquarterUuid).Error
				if err != nil {
					logger.Logger.Error("更新支付方式headquarter_uuid失败",
						zap.String("name", hqPayment.PaymentName),
						zap.Int("code", existPayment.Code),
						zap.Error(err))
				} else {
					logger.Logger.Info("更新支付方式headquarter_uuid",
						zap.String("name", hqPayment.PaymentName),
						zap.Int("code", existPayment.Code))
				}
			} else {
				// 普通code：跳过
				logger.Logger.Info("支付方式已存在，跳过同步",
					zap.String("name", hqPayment.PaymentName),
					zap.Int("code", existPayment.Code))
			}
			continue
		}

		// 分店不存在，创建新支付方式
		newCode := s.generatePaymentCode(subShopDB)

		newPayment := model.PaymentMethod{
			BaseModel: model.BaseModel{
				Uuid:       hqPayment.Uuid, // 保持与总部相同的uuid
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			HeadquarterUuid: headquarterUuid,
			PaymentName:     hqPayment.PaymentName,
			Name:            hqPayment.Name,
			Code:            newCode,                    // 生成新code
			Source:          model.PaymentSourceDefault, // 1-手动添加
			LogoFileUuid:    0,                          // 固定为0
			// 其他字段使用数据库默认值
		}

		err = subShopDB.Create(&newPayment).Error
		if err != nil {
			logger.Logger.Error("创建支付方式失败",
				zap.String("name", hqPayment.PaymentName),
				zap.Error(err))
			continue
		}

		logger.Logger.Info("创建新支付方式",
			zap.String("name", hqPayment.PaymentName),
			zap.Int("code", newCode))
	}

	return nil
}

// generatePaymentCode 生成支付方式code（与手动添加source=1规则一致）
func (s *SyncSrv) generatePaymentCode(db *gorm.DB) int {
	// 根据PHP代码：删除的值也要计算，防止重复，如果数据库找不到，则默认从20000开始
	var maxCode int
	db.Model(&model.PaymentMethod{}).Unscoped(). // 包含已删除的
							Where("source = ? AND code >= 20000", model.PaymentSourceDefault).
							Select("COALESCE(MAX(code), 19900)"). // 如果找不到，返回19900，+100后为20000
							Scan(&maxCode)

	return maxCode + 100 // 每次递增100
}

// GetHeadquartersDataList 获取总部可同步数据列表
func (s *SyncSrv) GetHeadquartersDataList(ctx context.Context, listReq req.GetHeadquartersDataListReq) (resp.HeadquartersDataListResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 只有分店才能查看总部数据
	if !companySetting.IsSubShop() {
		return resp.HeadquartersDataListResp{}, errors.New("非分店账号无法查看总部数据")
	}

	headquarterUuid := companySetting.HeadquarterUuid
	subShopUuid := companySetting.CompanyUuid

	// 获取数据库连接
	headquarterDB := s.dbm.GetDB(headquarterUuid)
	subShopDB := s.dbm.GetDB(subShopUuid)

	// 查询所有类型（不使用传参，固定查询所有）
	dataTypes := []string{
		constant.SyncTaskTypeProductCategory,
		constant.SyncTaskTypeUnit,
		constant.SyncTaskTypeFlavor,
		constant.SyncTaskTypeAttribute,
		constant.SyncTaskTypeSauce,
		constant.SyncTaskTypeProduct,
		constant.SyncTaskTypeMaterialCategory,
		constant.SyncTaskTypeMaterial,
		constant.SyncTaskTypeBomCard,
		constant.SyncTaskTypeSupplier,
		constant.SyncTaskTypeTax,
		constant.SyncTaskTypeCoupon,
		constant.SyncTaskTypeFullReduction,
		constant.SyncTaskTypeProductLabel,
		constant.SyncTaskTypeMarketingActivity,
		constant.SyncTaskTypePaymentMethod,
	}

	// 查询各类型数据
	var dataGroups []resp.DataGroup

	for _, dataType := range dataTypes {
		group, err := s.getDataGroupByType(ctx, dataType, headquarterDB, subShopDB, headquarterUuid)
		if err != nil {
			logger.Logger.Error("查询数据失败", zap.String("dataType", dataType), zap.Error(err))
			continue
		}
		dataGroups = append(dataGroups, group)
	}

	return resp.HeadquartersDataListResp{
		DataGroups: dataGroups,
	}, nil
}

// getDataGroupByType 根据数据类型查询数据分组
func (s *SyncSrv) getDataGroupByType(ctx context.Context, dataType string, headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	switch dataType {
	// 基础数据类型
	case constant.SyncTaskTypeProductCategory:
		return s.getProductCategoryGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeUnit:
		return s.getUnitGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeFlavor:
		return s.getFlavorGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeAttribute:
		return s.getAttributeGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeSauce:
		return s.getSauceGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeProduct:
		return s.getProductGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeMaterialCategory:
		return s.getMaterialCategoryGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeMaterial:
		return s.getMaterialGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeBomCard:
		return s.getBomCardGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeSupplier:
		return s.getSupplierGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeTax:
		return s.getTaxGroup(headquarterDB, subShopDB, headquarterUuid)

	// 营销数据类型
	case constant.SyncTaskTypeCoupon:
		return s.getCouponGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeFullReduction:
		return s.getFullReductionGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeProductLabel:
		return s.getProductLabelGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypeMarketingActivity:
		return s.getMarketingActivityGroup(headquarterDB, subShopDB, headquarterUuid)
	case constant.SyncTaskTypePaymentMethod:
		return s.getPaymentMethodGroup(headquarterDB, subShopDB, headquarterUuid)

	default:
		return resp.DataGroup{}, errors.New("不支持的数据类型")
	}
}

// getCouponGroup 获取优惠券数据分组
func (s *SyncSrv) getCouponGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	// 1. 查询总部优惠券（headquarter_uuid = 0）
	var hqCoupons []model.MarketingCoupon
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqCoupons).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部优惠券失败")
	}

	// 2. 查询分店已同步的优惠券uuid列表
	syncedUuids := make([]uint64, 0)
	err = subShopDB.Model(&model.MarketingCoupon{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步优惠券失败")
	}

	// 3. 组装数据项
	items := make([]resp.DataItem, 0)
	for _, coupon := range hqCoupons {
		items = append(items, resp.DataItem{
			Uuid: coupon.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   coupon.Name,
				TH:   coupon.Name,
				EN:   coupon.Name,
				ZHTW: coupon.Name,
				JA:   coupon.Name,
				KO:   coupon.Name,
				MY:   coupon.Name,
				TR:   coupon.Name,
				SV:   coupon.Name,
			},
			RelatedData: []resp.RelatedData{}, // 优惠券无关联数据
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeCoupon,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeCoupon],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getFullReductionGroup 获取满额减数据分组
func (s *SyncSrv) getFullReductionGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	// 1. 查询总部满额减活动
	var hqActivities []model.FullReductionActivity
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").
		Find(&hqActivities).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部满额减活动失败")
	}

	// 2. 查询分店已同步的满额减uuid列表
	syncedUuids := make([]uint64, 0)
	err = subShopDB.Model(&model.FullReductionActivity{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步满额减失败")
	}

	// 3. 组装数据项
	items := make([]resp.DataItem, 0)
	for _, activity := range hqActivities {
		items = append(items, resp.DataItem{
			Uuid:        activity.Uuid,
			LocaleName:  activity.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeFullReduction,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeFullReduction],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getProductLabelGroup 获取菜品标签数据分组
func (s *SyncSrv) getProductLabelGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	// 1. 查询总部菜品标签（Preload关联商品）
	var hqLabels []model.ProductLabel
	err := headquarterDB.Preload("ProductPackages").
		Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqLabels).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部菜品标签失败")
	}

	// 2. 查询分店已同步的菜品标签uuid列表
	syncedUuids := make([]uint64, 0)
	err = subShopDB.Model(&model.ProductLabel{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步菜品标签失败")
	}

	// 3. 组装数据项
	items := make([]resp.DataItem, 0)
	for _, label := range hqLabels {
		// 提取关联的商品uuid
		var relatedProductUuids []uint64
		for _, pkg := range label.ProductPackages {
			if pkg.ProductLabelUuid == label.Uuid {
				relatedProductUuids = append(relatedProductUuids, pkg.Uuid)
			}
		}

		// 构建关联数据（明确关联类型）
		var relatedData []resp.RelatedData
		if len(relatedProductUuids) > 0 {
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeProduct,
				Uuids: relatedProductUuids,
			})
		}

		items = append(items, resp.DataItem{
			Uuid: label.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   label.Name,
				TH:   label.Name,
				EN:   label.Name,
				ZHTW: label.Name,
				JA:   label.Name,
				KO:   label.Name,
				MY:   label.Name,
				TR:   label.Name,
				SV:   label.Name,
			},
			RelatedData: relatedData,
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeProductLabel,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeProductLabel],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getMarketingActivityGroup 获取营销活动数据分组
func (s *SyncSrv) getMarketingActivityGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	// 1. 查询总部营销活动
	var hqActivities []model.MarketingActivity
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").
		Find(&hqActivities).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部营销活动失败")
	}

	// 2. 查询分店已同步的营销活动uuid列表
	syncedUuids := make([]uint64, 0)
	err = subShopDB.Model(&model.MarketingActivity{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步营销活动失败")
	}

	// 3. 组装数据项
	items := make([]resp.DataItem, 0)
	for _, activity := range hqActivities {
		items = append(items, resp.DataItem{
			Uuid:        activity.Uuid,
			LocaleName:  activity.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeMarketingActivity,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeMarketingActivity],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getPaymentMethodGroup 获取支付方式数据分组（⚠️ 特殊：通过名称匹配判断已同步）
func (s *SyncSrv) getPaymentMethodGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	// 1. 查询总部支付方式（过滤 code=40 和 code=10）
	var hqPayments []model.PaymentMethod
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
		Where("code NOT IN (?)", []int{model.PaymentMethodCash, model.PaymentMethodBalance}).
		Find(&hqPayments).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部支付方式失败")
	}

	// 2. 查询分店中从总部同步的支付方式名称列表
	//    ⚠️ 关键：只查询 headquarter_uuid = 总部uuid 的支付方式
	var syncedPaymentNames []string
	err = subShopDB.Model(&model.PaymentMethod{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("payment_name", &syncedPaymentNames).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询分店已同步支付方式名称失败")
	}

	// 3. 构建已同步名称的map
	syncedNameMap := make(map[string]bool)
	for _, name := range syncedPaymentNames {
		syncedNameMap[name] = true
	}

	// 4. 匹配总部支付方式，找出已同步的总部uuid
	syncedUuids := make([]uint64, 0)
	items := make([]resp.DataItem, 0)

	for _, hqPayment := range hqPayments {
		// 通过名称匹配：分店有同名的总部支付方式，则该总部支付方式已同步
		if syncedNameMap[hqPayment.PaymentName] {
			syncedUuids = append(syncedUuids, hqPayment.Uuid)
		}

		items = append(items, resp.DataItem{
			Uuid: hqPayment.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   hqPayment.PaymentName,
				TH:   hqPayment.PaymentName,
				EN:   hqPayment.PaymentName,
				ZHTW: hqPayment.PaymentName,
				JA:   hqPayment.PaymentName,
				KO:   hqPayment.PaymentName,
				MY:   hqPayment.PaymentName,
				TR:   hqPayment.PaymentName,
				SV:   hqPayment.PaymentName,
			},
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypePaymentMethod,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypePaymentMethod],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getProductCategoryGroup 获取商品分类数据分组
func (s *SyncSrv) getProductCategoryGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqCategories []model.ProductCategory
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqCategories).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部商品分类失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductCategory{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, category := range hqCategories {
		items = append(items, resp.DataItem{
			Uuid:        category.Uuid,
			LocaleName:  category.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeProductCategory,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeProductCategory],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getUnitGroup 获取单位数据分组
func (s *SyncSrv) getUnitGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqUnits []model.ProductUnit
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqUnits).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部单位失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductUnit{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, unit := range hqUnits {
		items = append(items, resp.DataItem{
			Uuid:        unit.Uuid,
			LocaleName:  unit.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeUnit,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeUnit],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getFlavorGroup 获取规格数据分组
func (s *SyncSrv) getFlavorGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqFlavors []model.ProductFlavor
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqFlavors).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部规格失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductFlavor{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, flavor := range hqFlavors {
		items = append(items, resp.DataItem{
			Uuid:        flavor.Uuid,
			LocaleName:  flavor.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeFlavor,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeFlavor],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getAttributeGroup 获取属性数据分组
func (s *SyncSrv) getAttributeGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqAttributes []model.ProductAttributeGroup
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqAttributes).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部属性失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductAttributeGroup{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, attr := range hqAttributes {
		items = append(items, resp.DataItem{
			Uuid:        attr.Uuid,
			LocaleName:  attr.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeAttribute,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeAttribute],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getSauceGroup 获取加料数据分组
func (s *SyncSrv) getSauceGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqSauces []model.ProductSauce
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqSauces).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部加料失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductSauce{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, sauce := range hqSauces {
		items = append(items, resp.DataItem{
			Uuid:        sauce.Uuid,
			LocaleName:  sauce.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeSauce,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeSauce],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getProductGroup 获取商品数据分组
func (s *SyncSrv) getProductGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqProducts []model.ProductPackage
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqProducts).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部商品失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductPackage{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	// 获取商品包的uuid列表
	productPackageUuids := make([]uint64, 0, len(hqProducts))
	for _, p := range hqProducts {
		productPackageUuids = append(productPackageUuids, p.Uuid)
	}

	// 查询商品的 ProductBom 关联（规格、小料、成本卡）
	type ProductBomRelation struct {
		ProductPackageUuid uint64
		ProductFlavorUuid  uint64
		ProductSauceUuid   uint64
		ProductBomCardUuid uint64
	}
	var productBoms []ProductBomRelation
	if len(productPackageUuids) > 0 {
		headquarterDB.Table("ttpos_product_bom").
			Select("product_package_uuid, product_flavor_uuid, product_sauce_uuid, product_bom_card_uuid").
			Where("product_package_uuid IN (?) AND delete_time = 0", productPackageUuids).
			Find(&productBoms)
	}

	// 构建商品 -> ProductBom关联的映射
	productBomMap := make(map[uint64][]ProductBomRelation)
	for _, bom := range productBoms {
		productBomMap[bom.ProductPackageUuid] = append(productBomMap[bom.ProductPackageUuid], bom)
	}

	// 查询商品的属性组关联
	type ProductAttributeGroupRelation struct {
		ProductPackageUuid        uint64
		ProductAttributeGroupUuid uint64
	}
	var productAttrGroups []ProductAttributeGroupRelation
	if len(productPackageUuids) > 0 {
		headquarterDB.Table("ttpos_product_package_attribute_group").
			Select("product_package_uuid, product_attribute_group_uuid").
			Where("product_package_uuid IN (?) AND delete_time = 0", productPackageUuids).
			Find(&productAttrGroups)
	}

	// 构建商品 -> 属性组的映射
	productAttrGroupMap := make(map[uint64][]uint64)
	for _, ag := range productAttrGroups {
		productAttrGroupMap[ag.ProductPackageUuid] = append(productAttrGroupMap[ag.ProductPackageUuid], ag.ProductAttributeGroupUuid)
	}

	items := make([]resp.DataItem, 0)
	for _, product := range hqProducts {
		// 构建关联数据
		var relatedData []resp.RelatedData

		// 关联单位
		if product.UnitUuid > 0 {
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeUnit,
				Uuids: []uint64{product.UnitUuid},
			})
		}

		// 关联分类（普通分类 + 特殊分类，去重）
		categoryUuidMap := make(map[uint64]bool)
		if product.CategoryUuid > 0 {
			categoryUuidMap[product.CategoryUuid] = true
		}
		if product.SpecialCategoryUuid > 0 {
			categoryUuidMap[product.SpecialCategoryUuid] = true
		}
		if len(categoryUuidMap) > 0 {
			var categoryUuids []uint64
			for categoryUuid := range categoryUuidMap {
				categoryUuids = append(categoryUuids, categoryUuid)
			}
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeProductCategory,
				Uuids: categoryUuids,
			})
		}

		// 关联税类（堂食税 + 外卖税，去重）
		taxUuidMap := make(map[uint64]bool)
		if product.DineTaxUuid > 0 {
			taxUuidMap[product.DineTaxUuid] = true
		}
		if product.TakeoutTaxUuid > 0 {
			taxUuidMap[product.TakeoutTaxUuid] = true
		}
		if len(taxUuidMap) > 0 {
			var taxUuids []uint64
			for taxUuid := range taxUuidMap {
				taxUuids = append(taxUuids, taxUuid)
			}
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeTax,
				Uuids: taxUuids,
			})
		}

		// 关联规格、小料、成本卡（通过 ProductBom 中间表）
		if boms, exists := productBomMap[product.Uuid]; exists {
			flavorUuids := make(map[uint64]bool)
			sauceUuids := make(map[uint64]bool)
			bomCardUuids := make(map[uint64]bool)

			for _, bom := range boms {
				if bom.ProductFlavorUuid > 0 {
					flavorUuids[bom.ProductFlavorUuid] = true
				}
				if bom.ProductSauceUuid > 0 {
					sauceUuids[bom.ProductSauceUuid] = true
				}
				if bom.ProductBomCardUuid > 0 {
					bomCardUuids[bom.ProductBomCardUuid] = true
				}
			}

			// 添加规格关联
			if len(flavorUuids) > 0 {
				var uuids []uint64
				for uuid := range flavorUuids {
					uuids = append(uuids, uuid)
				}
				relatedData = append(relatedData, resp.RelatedData{
					Type:  constant.SyncTaskTypeFlavor,
					Uuids: uuids,
				})
			}

			// 添加小料关联
			if len(sauceUuids) > 0 {
				var uuids []uint64
				for uuid := range sauceUuids {
					uuids = append(uuids, uuid)
				}
				relatedData = append(relatedData, resp.RelatedData{
					Type:  constant.SyncTaskTypeSauce,
					Uuids: uuids,
				})
			}

			// 添加成本卡关联
			if len(bomCardUuids) > 0 {
				var uuids []uint64
				for uuid := range bomCardUuids {
					uuids = append(uuids, uuid)
				}
				relatedData = append(relatedData, resp.RelatedData{
					Type:  constant.SyncTaskTypeBomCard,
					Uuids: uuids,
				})
			}
		}

		// 关联属性组（通过 ProductPackageAttributeGroup 中间表）
		if attrGroupUuids, exists := productAttrGroupMap[product.Uuid]; exists && len(attrGroupUuids) > 0 {
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeAttribute,
				Uuids: attrGroupUuids,
			})
		}

		items = append(items, resp.DataItem{
			Uuid:        product.Uuid,
			LocaleName:  product.MultiLanguageName.GetNames(),
			RelatedData: relatedData,
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeProduct,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeProduct],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getMaterialCategoryGroup 获取物品分类数据分组
func (s *SyncSrv) getMaterialCategoryGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqCategories []model.MaterialCategory
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqCategories).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部物品分类失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.MaterialCategory{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, category := range hqCategories {
		items = append(items, resp.DataItem{
			Uuid:        category.Uuid,
			LocaleName:  category.MultiLanguageName.GetNames(),
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeMaterialCategory,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterialCategory],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getMaterialGroup 获取物品数据分组
func (s *SyncSrv) getMaterialGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqMaterials []model.Material
	err := headquarterDB.Preload("NotBaseUnitList").
		Preload("Unit").
		Preload("PurchaseUnit").
		Preload("CostUnit").
		Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqMaterials).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部物品失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.Material{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, material := range hqMaterials {
		var relatedData []resp.RelatedData

		// 提取关联的单位uuid（去重）
		unitUuidMap := make(map[uint64]bool)

		// 直接字段的单位
		if material.Unit != nil && material.Unit.UnitUuid > 0 {
			unitUuidMap[material.Unit.UnitUuid] = true
		}
		if material.PurchaseUnit != nil && material.PurchaseUnit.UnitUuid > 0 {
			unitUuidMap[material.PurchaseUnit.UnitUuid] = true
		}
		if material.CostUnit != nil && material.CostUnit.UnitUuid > 0 {
			unitUuidMap[material.CostUnit.UnitUuid] = true
		}

		// 非基准单位列表的单位（material_unit.unit_uuid → product_unit 表）
		for _, materialUnit := range material.NotBaseUnitList {
			if materialUnit.UnitUuid > 0 {
				unitUuidMap[materialUnit.UnitUuid] = true
			}
		}

		// 转为切片
		if len(unitUuidMap) > 0 {
			var unitUuids []uint64
			for unitUuid := range unitUuidMap {
				unitUuids = append(unitUuids, unitUuid)
			}
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeUnit,
				Uuids: unitUuids,
			})
		}

		// 关联物品分类
		if material.CategoryUuid > 0 {
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeMaterialCategory,
				Uuids: []uint64{material.CategoryUuid},
			})
		}

		items = append(items, resp.DataItem{
			Uuid:        material.Uuid,
			LocaleName:  material.MultiLanguageName.GetNames(),
			RelatedData: relatedData,
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeMaterial,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterial],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getBomCardGroup 获取成本卡数据分组
func (s *SyncSrv) getBomCardGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqBomCards []model.ProductBomCard
	err := headquarterDB.Preload("RelatedMaterials").
		Where("delete_time = 0 AND headquarter_uuid = 0").Preload("MultiLanguageName").Find(&hqBomCards).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部成本卡失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.ProductBomCard{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, bomCard := range hqBomCards {
		var relatedData []resp.RelatedData

		// 提取关联的物品uuid
		var materialUuids []uint64
		for _, relatedMaterial := range bomCard.RelatedMaterials {
			if relatedMaterial.MaterialUuid > 0 {
				materialUuids = append(materialUuids, relatedMaterial.MaterialUuid)
			}
		}

		if len(materialUuids) > 0 {
			relatedData = append(relatedData, resp.RelatedData{
				Type:  constant.SyncTaskTypeMaterial,
				Uuids: materialUuids,
			})
		}

		items = append(items, resp.DataItem{
			Uuid:        bomCard.Uuid,
			LocaleName:  bomCard.MultiLanguageName.GetNames(),
			RelatedData: relatedData,
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeBomCard,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeBomCard],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getSupplierGroup 获取供应商数据分组
func (s *SyncSrv) getSupplierGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqSuppliers []model.Supplier
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Find(&hqSuppliers).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部供应商失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.Supplier{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, supplier := range hqSuppliers {
		items = append(items, resp.DataItem{
			Uuid: supplier.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   supplier.Name,
				TH:   supplier.Name,
				EN:   supplier.Name,
				ZHTW: supplier.Name,
				JA:   supplier.Name,
				KO:   supplier.Name,
				MY:   supplier.Name,
				TR:   supplier.Name,
				SV:   supplier.Name,
			},
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeSupplier,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeSupplier],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// getTaxGroup 获取税类数据分组
func (s *SyncSrv) getTaxGroup(headquarterDB, subShopDB *gorm.DB, headquarterUuid uint64) (resp.DataGroup, error) {
	var hqTaxes []model.Tax
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").Find(&hqTaxes).Error
	if err != nil {
		return resp.DataGroup{}, errors.WithMessage(err, "查询总部税类失败")
	}

	syncedUuids := make([]uint64, 0)
	subShopDB.Model(&model.Tax{}).
		Where("delete_time = 0 AND headquarter_uuid = ?", headquarterUuid).
		Pluck("uuid", &syncedUuids)

	items := make([]resp.DataItem, 0)
	for _, tax := range hqTaxes {
		items = append(items, resp.DataItem{
			Uuid: tax.Uuid,
			LocaleName: dto.LocaleResponse{
				ZH:   tax.Name,
				TH:   tax.Name,
				EN:   tax.Name,
				ZHTW: tax.Name,
				JA:   tax.Name,
				KO:   tax.Name,
				MY:   tax.Name,
				TR:   tax.Name,
				SV:   tax.Name,
			},
			RelatedData: []resp.RelatedData{},
		})
	}

	return resp.DataGroup{
		Type:        constant.SyncTaskTypeTax,
		TypeName:    constant.SyncTaskTypeNames[constant.SyncTaskTypeTax],
		Items:       items,
		SyncedUuids: syncedUuids,
	}, nil
}

// GranularSync 颗粒化同步数据
func (s *SyncSrv) GranularSync(ctx context.Context, syncReq req.GranularSyncReq) (resp.GranularSyncResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 只有分店才能执行颗粒化同步
	if !companySetting.IsSubShop() {
		return resp.GranularSyncResp{}, errors.New("非分店账号无法执行颗粒化同步")
	}

	companyUuid := companySetting.CompanyUuid

	// 检查是否已有同步任务在运行
	if !syncTaskManager.tryStartTask(companyUuid) {
		return resp.GranularSyncResp{}, errors.New("数据同步中，请稍后再试")
	}

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	// 创建新的同步任务
	syncTask := &model.SyncTask{
		Status:       constant.SyncTaskStatusRunning,
		TotalCount:   0, // 后续计算
		SuccessCount: 0,
		FailCount:    0,
		StartTime:    time.Now().Unix(),
	}

	if err := syncTaskRepo.Create(syncTask); err != nil {
		syncTaskManager.finishTask(companyUuid)
		return resp.GranularSyncResp{}, errors.WithMessage(err, "创建同步任务失败")
	}

	// 异步执行颗粒化同步
	utils.Go(func() {
		s.executeGranularSync(ctx, syncTask, syncReq.SyncData)
	})

	return resp.GranularSyncResp{
		TaskUuid: syncTask.Uuid,
		Message:  "数据同步已启动，可在同步历史中查看进度",
	}, nil
}

// executeGranularSync 执行颗粒化同步
func (s *SyncSrv) executeGranularSync(ctx context.Context, syncTask *model.SyncTask, syncData req.GranularSyncData) {
	companySetting := ctx.GetCompanySetting()
	companyUuid := companySetting.CompanyUuid

	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	var successCount uint32
	var failCount uint32

	defer func() {
		var isPanicOccurred bool
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			logger.Logger.Error("颗粒化同步发生panic", zap.Uint64("companyUuid", companyUuid), zap.Any("panic", r), zap.String("stack", stack))
			syncTaskRepo.Update(syncTask.Uuid, map[string]any{
				"status":   constant.SyncTaskStatusFailed,
				"panic":    fmt.Sprintf("%v: %s", r, stack),
				"end_time": time.Now().Unix(),
			})
			isPanicOccurred = true
		}

		isExceptionOccurred := failCount > 0 || isPanicOccurred

		syncTaskManager.finishTask(companyUuid)
		logger.Logger.Info("颗粒化同步完成", zap.Uint64("companyUuid", companyUuid),
			zap.Uint32("successCount", successCount),
			zap.Uint32("failCount", failCount))

		// 推送websocket
		utils.Go(func() {
			websocket.PushClient(companyUuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
				"task_uuid":             syncTask.Uuid,
				"is_exception_occurred": isExceptionOccurred,
				"sync_time":             time.Now().Unix(),
			})
		})
	}()

	logger.Logger.Info("开始执行颗粒化同步", zap.Uint64("companyUuid", companyUuid), zap.Uint64("taskUuid", syncTask.Uuid))

	// 注意：基础数据类型复用 allTasks 中的方法（传 useFilter=true），营销数据类型使用独立方法
	syncTasks := []struct {
		Name     string
		Uuids    []uint64
		Executor func(context.Context, []uint64) error
	}{
		// 基础数据类型（复用 allTasks 中的方法）
		{constant.SyncTaskTypeProductCategory, syncData.ProductCategory, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProductShopCategory(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeMaterialCategory, syncData.MaterialCategory, func(ctx context.Context, uuids []uint64) error {
			return s.materialSrv.SyncMaterialCategory(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeTax, syncData.Tax, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProductTax(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeUnit, syncData.Unit, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncUnit(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeMaterial, syncData.Material, func(ctx context.Context, uuids []uint64) error {
			return s.materialSrv.SyncMaterial(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeWarehouse, nil, func(ctx context.Context, uuids []uint64) error {
			return s.warehouseSrv.SyncWarehouse(ctx) // ok
		}},
		{constant.SyncTaskTypeFlavor, syncData.Flavor, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProductFlavor(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeAttribute, syncData.Attribute, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncAttributeGroup(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeSauce, syncData.Sauce, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncSauce(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeProduct, syncData.Product, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProduct(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeBomCard, syncData.BomCard, func(ctx context.Context, uuids []uint64) error {
			return s.materialSrv.SyncProductBomCard(ctx, true, uuids) // ok
		}},
		{constant.SyncTaskTypeSupplier, syncData.Supplier, func(ctx context.Context, uuids []uint64) error {
			return s.supplierSrv.SyncSupplier(ctx, true, uuids) // ok
		}},

		{constant.SyncTaskTypeWarehouseStock, nil, func(ctx context.Context, uuids []uint64) error {
			return s.warehouseSrv.SyncWarehouseItemStock(ctx) // ok
		}},
		{constant.SyncTaskTypeProductStock, nil, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProductStockByBomCard(ctx) // ok
		}},
		{constant.SyncTaskTypePackageImage, nil, func(ctx context.Context, uuids []uint64) error {
			return s.productSrv.SyncProductPackageImage(ctx) // ok
		}},
		// 营销数据类型（独立方法，新增的数据类型）
		{constant.SyncTaskTypeCoupon, syncData.Coupon, s.SyncMarketingCouponByUuids},
		{constant.SyncTaskTypeFullReduction, syncData.FullReduction, s.SyncFullReductionByUuids},
		{constant.SyncTaskTypeProductLabel, syncData.ProductLabel, s.SyncProductLabelByUuids},
		{constant.SyncTaskTypeMarketingActivity, syncData.MarketingActivity, s.SyncMarketingActivityByUuids},
		{constant.SyncTaskTypePaymentMethod, syncData.PaymentMethod, s.SyncPaymentMethodByUuids},

		{constant.SyncTaskTypeMultiLanguage, nil, func(ctx context.Context, uuids []uint64) error {
			return s.SyncMultiLanguage(ctx) // ok
		}},
	}

	syncTaskItemRepo := repository.NewSyncTaskItemRepo(s.dbm.GetDB(companyUuid))

	for _, task := range syncTasks {
		if len(task.Uuids) == 0 {
			continue
		}

		taskItem := &model.SyncTaskItem{
			SyncTaskUuid: syncTask.Uuid,
			TaskType:     task.Name,
			TaskName:     constant.SyncTaskTypeNames[task.Name],
			Status:       constant.SyncTaskItemStatusRunning,
			StartTime:    time.Now().Unix(),
		}

		syncTaskItemRepo.Create(taskItem)

		logger.Logger.Info("开始同步", zap.String("taskName", taskItem.TaskName))
		err := task.Executor(ctx, task.Uuids)
		endTime := time.Now().Unix()

		if err != nil {
			failCount++
			logger.Logger.Error("同步失败", zap.String("taskName", taskItem.TaskName), zap.Error(err))
			syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
				"status":        constant.SyncTaskItemStatusFailed,
				"error_message": err.Error(),
				"end_time":      endTime,
			})
		} else {
			successCount++
			logger.Logger.Info("同步成功", zap.String("taskName", taskItem.TaskName))
			syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
				"status":   constant.SyncTaskItemStatusSuccess,
				"end_time": endTime,
			})
		}
	}

	// 更新主任务状态
	endTime := time.Now().Unix()
	finalStatus := constant.SyncTaskStatusSuccess
	if failCount > 0 {
		finalStatus = constant.SyncTaskStatusFailed
	}

	syncTaskRepo.Update(syncTask.Uuid, map[string]any{
		"status":        finalStatus,
		"success_count": successCount,
		"fail_count":    failCount,
		"total_count":   successCount + failCount,
		"end_time":      endTime,
	})
}
