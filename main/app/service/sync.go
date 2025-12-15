package service

import (
	"encoding/json"
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
	Executor func(context.Context, bool) error
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
		{constant.SyncTaskTypeProductCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeProductCategory], s.productSrv.SyncProductShopCategory},
		{constant.SyncTaskTypeMaterialCategory, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterialCategory], s.materialSrv.SyncMaterialCategory},
		{constant.SyncTaskTypeTax, constant.SyncTaskTypeNames[constant.SyncTaskTypeTax], s.productSrv.SyncProductTax}, // 无多语言数据
		{constant.SyncTaskTypeUnit, constant.SyncTaskTypeNames[constant.SyncTaskTypeUnit], s.productSrv.SyncUnit},
		{constant.SyncTaskTypeMaterial, constant.SyncTaskTypeNames[constant.SyncTaskTypeMaterial], s.materialSrv.SyncMaterial},
		{constant.SyncTaskTypeWarehouse, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouse], s.warehouseSrv.SyncWarehouse},
		{constant.SyncTaskTypeFlavor, constant.SyncTaskTypeNames[constant.SyncTaskTypeFlavor], s.productSrv.SyncProductFlavor},
		{constant.SyncTaskTypeAttribute, constant.SyncTaskTypeNames[constant.SyncTaskTypeAttribute], s.productSrv.SyncAttributeGroup},
		{constant.SyncTaskTypeSauce, constant.SyncTaskTypeNames[constant.SyncTaskTypeSauce], s.productSrv.SyncSauce},
		{constant.SyncTaskTypeProduct, constant.SyncTaskTypeNames[constant.SyncTaskTypeProduct], s.productSrv.SyncProduct},
		{constant.SyncTaskTypeBomCard, constant.SyncTaskTypeNames[constant.SyncTaskTypeBomCard], s.materialSrv.SyncProductBomCard},
		{constant.SyncTaskTypeSupplier, constant.SyncTaskTypeNames[constant.SyncTaskTypeSupplier], s.supplierSrv.SyncSupplier}, // 无多语言数据
		{constant.SyncTaskTypeWarehouseStock, constant.SyncTaskTypeNames[constant.SyncTaskTypeWarehouseStock], func(ctx context.Context, syncHeadquarterData bool) error {
			return s.warehouseSrv.SyncWarehouseItemStock(ctx)
		}}, // 无多语言数据
		{constant.SyncTaskTypePackageImage, constant.SyncTaskTypeNames[constant.SyncTaskTypePackageImage], s.productSrv.SyncProductPackageImage}, // 无多语言数据
		{constant.SyncTaskTypeMultiLanguage, constant.SyncTaskTypeNames[constant.SyncTaskTypeMultiLanguage], func(ctx context.Context, syncHeadquarterData bool) error {
			return s.SyncMultiLanguage(ctx)
		}},
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
	err := taskCfg.Executor(ctx, true)
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

// GetHeadquartersDataList 获取总部可同步数据列表（返回三组数据）
func (s *SyncSrv) GetHeadquartersDataList(ctx context.Context, listReq req.GetHeadquartersDataListReq) (resp.HeadquartersDataListResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 只有分店才能查看总部数据
	if !companySetting.IsSubShop() {
		return resp.HeadquartersDataListResp{}, errors.New("非分店账号无法查看总部数据")
	}

	headquarterUuid := companySetting.HeadquarterUuid
	subShopUuid := companySetting.CompanyUuid

	// 获取数据库连接
	subShopDB := s.dbm.GetDB(subShopUuid)

	// 查询最近一次成功的颗粒化同步任务，获取请求参数
	lastSuccessTask, err := repository.NewSyncTaskRepo(subShopDB).GetLastGranularSyncTask()
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Logger.Error("查询最近一次成功的同步任务失败", zap.Error(err))
	}

	var lastRequestParams *req.GranularSyncReq
	if err == nil && lastSuccessTask != nil && lastSuccessTask.RequestParams != "" {
		var reqParams req.GranularSyncReq
		if err := json.Unmarshal([]byte(lastSuccessTask.RequestParams), &reqParams); err == nil {
			lastRequestParams = &reqParams
		} else {
			logger.Logger.Warn("解析上次请求参数失败", zap.Error(err))
		}
	}

	var groups []resp.DataGroupResp

	// 商品数据组
	productDataTypes := []string{
		constant.SyncTaskTypeProductCategory,
		constant.SyncTaskTypeUnit,
		constant.SyncTaskTypeFlavor,
		constant.SyncTaskTypeAttribute,
		constant.SyncTaskTypeSauce,
		constant.SyncTaskTypeProduct,
		constant.SyncTaskTypeMaterialCategory,
		constant.SyncTaskTypeBomCard,
		constant.SyncTaskTypeSupplier,
		constant.SyncTaskTypeTax,
	}
	// 根据上次请求参数判断是否已同步
	var productSynced bool
	if lastRequestParams != nil {
		productSynced = lastRequestParams.ProductDataChecked
	} else {
		// 如果没有上次请求参数，使用原来的逻辑
		productSynced = s.checkGroupSynced(subShopDB, headquarterUuid, productDataTypes)
	}
	groups = append(groups, resp.DataGroupResp{
		Group:        "product_data",
		GroupName:    "商品数据",
		Synced:       productSynced,
		Dependencies: []string{},
	})

	// 活动数据组
	activityDataTypes := []string{
		constant.SyncTaskTypeCoupon,
		constant.SyncTaskTypeFullReduction,
		constant.SyncTaskTypeProductLabel,
		constant.SyncTaskTypeMarketingActivity,
	}
	var activitySynced bool
	if lastRequestParams != nil {
		activitySynced = lastRequestParams.ActivityDataChecked
	} else {
		// 如果没有上次请求参数，使用原来的逻辑
		activitySynced = s.checkGroupSynced(subShopDB, headquarterUuid, activityDataTypes)
	}
	dependencies := []string{}
	if s.hasActivityDataDependsOnProductData(ctx) {
		dependencies = append(dependencies, "product_data")
	}
	groups = append(groups, resp.DataGroupResp{
		Group:        "activity_data",
		GroupName:    "活动数据",
		Synced:       activitySynced,
		Dependencies: dependencies,
	})

	// 支付数据组
	var paymentSynced bool
	if lastRequestParams != nil {
		paymentSynced = lastRequestParams.PaymentDataChecked
	} else {
		// 如果没有上次请求参数，使用原来的逻辑
		paymentSynced = s.checkPaymentGroupSynced(subShopDB, headquarterUuid)
	}
	groups = append(groups, resp.DataGroupResp{
		Group:        "other_data",
		GroupName:    "其他数据",
		Synced:       paymentSynced,
		Dependencies: []string{},
	})

	return resp.HeadquartersDataListResp{
		Groups: groups,
	}, nil
}

// checkGroupSynced 检查分组是否已同步（组级别：只要有一个类型已同步，则认为组已同步）
func (s *SyncSrv) checkGroupSynced(subShopDB *gorm.DB, headquarterUuid uint64, dataTypes []string) bool {
	for _, dataType := range dataTypes {
		tableName := s.getTableNameByDataType(dataType)
		if tableName == "" {
			continue
		}

		var count int64
		err := subShopDB.Table(tableName).
			Where("headquarter_uuid = ?", headquarterUuid).
			Where("delete_time = 0").
			Count(&count).Error
		if err != nil {
			logger.Logger.Error("检查分组同步状态失败", zap.String("dataType", dataType), zap.Error(err))
			continue
		}

		// 只要有一个类型已同步，则认为组已同步
		if count > 0 {
			return true
		}
	}

	return false
}

// checkPaymentGroupSynced 检查支付数据组是否已同步
func (s *SyncSrv) checkPaymentGroupSynced(subShopDB *gorm.DB, headquarterUuid uint64) bool {
	var count int64
	err := subShopDB.Table("ttpos_payment_method").
		Where("headquarter_uuid = ?", headquarterUuid).
		Where("delete_time = 0").
		Count(&count).Error
	if err != nil {
		logger.Logger.Error("检查支付数据组同步状态失败", zap.Error(err))
		return false
	}

	return count > 0
}

// getTableNameByDataType 根据数据类型获取表名
func (s *SyncSrv) getTableNameByDataType(dataType string) string {
	switch dataType {
	case constant.SyncTaskTypeProductCategory:
		return "ttpos_product_category"
	case constant.SyncTaskTypeUnit:
		return "ttpos_product_unit"
	case constant.SyncTaskTypeFlavor:
		return "ttpos_product_flavor"
	case constant.SyncTaskTypeAttribute:
		return "ttpos_product_attribute_group"
	case constant.SyncTaskTypeSauce:
		return "ttpos_product_sauce"
	case constant.SyncTaskTypeProduct:
		return "ttpos_product_package"
	case constant.SyncTaskTypeMaterialCategory:
		return "ttpos_material_category"
	case constant.SyncTaskTypeBomCard:
		return "ttpos_product_bom_card"
	case constant.SyncTaskTypeSupplier:
		return "ttpos_supplier"
	case constant.SyncTaskTypeTax:
		return "ttpos_tax"
	case constant.SyncTaskTypeCoupon:
		return "ttpos_marketing_coupon"
	case constant.SyncTaskTypeFullReduction:
		return "ttpos_full_reduction_activity"
	case constant.SyncTaskTypeProductLabel:
		return "ttpos_product_label"
	case constant.SyncTaskTypeMarketingActivity:
		return "ttpos_marketing_activity"
	default:
		return ""
	}
}

// GranularSync 颗粒化同步数据
func (s *SyncSrv) GranularSync(ctx context.Context, syncReq req.GranularSyncReq) (resp.GranularSyncResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 只有分店才能执行颗粒化同步
	if !companySetting.IsSubShop() {
		return resp.GranularSyncResp{}, errors.New("非分店账号无法执行颗粒化同步")
	}

	// 依赖检查：如果勾选了活动数据组但未勾选商品数据组，检查活动数据是否依赖商品数据
	if syncReq.ActivityDataChecked && !syncReq.ProductDataChecked {
		if s.hasActivityDataDependsOnProductData(ctx) {
			return resp.GranularSyncResp{}, errors.New("选择的数据正在使用商品数据，请一并勾选所需内容")
		}
	}

	companyUuid := companySetting.CompanyUuid

	// 检查是否已有同步任务在运行
	if !syncTaskManager.tryStartTask(companyUuid) {
		return resp.GranularSyncResp{}, errors.New("数据同步中，请稍后再试")
	}

	// 实例化repo（同步任务表在公司库）
	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))

	// 序列化请求参数为JSON
	requestParamsJSON, err := json.Marshal(syncReq)
	if err != nil {
		syncTaskManager.finishTask(companyUuid)
		return resp.GranularSyncResp{}, errors.WithMessage(err, "序列化请求参数失败")
	}

	// 创建新的同步任务
	syncTask := &model.SyncTask{
		Status:        constant.SyncTaskStatusRunning,
		TotalCount:    0, // 后续计算
		SuccessCount:  0,
		FailCount:     0,
		StartTime:     time.Now().Unix(),
		RequestParams: string(requestParamsJSON),
	}

	if err := syncTaskRepo.Create(syncTask); err != nil {
		syncTaskManager.finishTask(companyUuid)
		return resp.GranularSyncResp{}, errors.WithMessage(err, "创建同步任务失败")
	}

	// 异步执行颗粒化同步
	utils.Go(func() {
		s.executeGranularSync(ctx, syncTask, syncReq.ProductDataChecked, syncReq.ActivityDataChecked, syncReq.PaymentDataChecked)
	})

	return resp.GranularSyncResp{
		TaskUuid: syncTask.Uuid,
		Message:  "新数据同步完成后将会通知",
	}, nil
}

// hasActivityDataDependsOnProductData 检查活动数据是否依赖商品数据
func (s *SyncSrv) hasActivityDataDependsOnProductData(ctx context.Context) bool {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)

	// 检查product_package表中是否存在使用了product_label的记录
	var count int64
	err := headquarterDB.Table("ttpos_product_package").
		Where("headquarter_uuid = 0 AND delete_time = 0").
		Where("product_label_uuid > 0").
		Count(&count).Error
	if err != nil {
		logger.Logger.Error("查询product_package使用product_label失败", zap.Error(err))
		return false
	}

	// 如果存在使用了product_label的记录，则活动数据依赖商品数据
	return count > 0
}

// executeGranularSync 执行颗粒化同步
func (s *SyncSrv) executeGranularSync(ctx context.Context, syncTask *model.SyncTask, productDataChecked, activityDataChecked, paymentDataChecked bool) {
	companySetting := ctx.GetCompanySetting()
	companyUuid := companySetting.CompanyUuid

	syncTaskRepo := repository.NewSyncTaskRepo(s.dbm.GetDB(companyUuid))
	syncTaskItemRepo := repository.NewSyncTaskItemRepo(s.dbm.GetDB(companyUuid))

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

		lastSyncTime := time.Now().Unix()
		// 未报错才记录上次同步完成时间
		if !isExceptionOccurred {
			s.dbm.GetDB(companyUuid).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
			s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Where("uuid = ?", companyUuid).Update("last_sync_time", lastSyncTime)
		}

		// 推送websocket
		utils.Go(func() {
			websocket.PushClient(companyUuid, websocket.SourceShop, websocket.SourceAll, websocket.SYNC_DATA, map[string]any{
				"task_uuid":             syncTask.Uuid,
				"is_exception_occurred": isExceptionOccurred,
				"sync_time":             time.Now().Unix(),
			})
		})
	}()

	logger.Logger.Info("开始执行颗粒化同步", zap.Uint64("companyUuid", companyUuid), zap.Uint64("taskUuid", syncTask.Uuid),
		zap.Bool("productDataChecked", productDataChecked),
		zap.Bool("activityDataChecked", activityDataChecked),
		zap.Bool("paymentDataChecked", paymentDataChecked))

	// 商品数据组：按照指定顺序执行全量同步
	productSyncTasks := []struct {
		Name     string
		Executor func(context.Context, bool) error
	}{
		{constant.SyncTaskTypeProductCategory, s.productSrv.SyncProductShopCategory},
		{constant.SyncTaskTypeMaterialCategory, s.materialSrv.SyncMaterialCategory},
		{constant.SyncTaskTypeTax, s.productSrv.SyncProductTax},
		{constant.SyncTaskTypeUnit, s.productSrv.SyncUnit},
		{constant.SyncTaskTypeMaterial, s.materialSrv.SyncMaterial},
		{constant.SyncTaskTypeWarehouse, s.warehouseSrv.SyncWarehouse},
		{constant.SyncTaskTypeFlavor, s.productSrv.SyncProductFlavor},
		{constant.SyncTaskTypeAttribute, s.productSrv.SyncAttributeGroup},
		{constant.SyncTaskTypeSauce, s.productSrv.SyncSauce},
		{constant.SyncTaskTypeProduct, s.productSrv.SyncProduct},
		{constant.SyncTaskTypeBomCard, s.materialSrv.SyncProductBomCard},
		{constant.SyncTaskTypeSupplier, s.supplierSrv.SyncSupplier},
		{constant.SyncTaskTypeWarehouseStock, func(ctx context.Context, syncHeadquarterData bool) error {
			return s.warehouseSrv.SyncWarehouseItemStock(ctx)
		}},
		{constant.SyncTaskTypePackageImage, s.productSrv.SyncProductPackageImage},
	}

	for _, task := range productSyncTasks {
		taskItem := &model.SyncTaskItem{
			SyncTaskUuid: syncTask.Uuid,
			TaskType:     task.Name,
			TaskName:     constant.SyncTaskTypeNames[task.Name],
			Status:       constant.SyncTaskItemStatusRunning,
			StartTime:    time.Now().Unix(),
		}

		syncTaskItemRepo.Create(taskItem)

		logger.Logger.Info("开始同步", zap.String("taskName", taskItem.TaskName))
		err := task.Executor(ctx, productDataChecked)
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

	// 活动数据组：按照指定顺序执行全量同步
	if activityDataChecked {
		activitySyncTasks := []struct {
			Name     string
			Executor func(context.Context) error
		}{
			{constant.SyncTaskTypeCoupon, s.SyncMarketingCoupon},
			{constant.SyncTaskTypeFullReduction, s.SyncFullReduction},
			{constant.SyncTaskTypeProductLabel, s.SyncProductLabel},
			{constant.SyncTaskTypeMarketingActivity, s.SyncMarketingActivity},
		}

		for _, task := range activitySyncTasks {
			taskItem := &model.SyncTaskItem{
				SyncTaskUuid: syncTask.Uuid,
				TaskType:     task.Name,
				TaskName:     constant.SyncTaskTypeNames[task.Name],
				Status:       constant.SyncTaskItemStatusRunning,
				StartTime:    time.Now().Unix(),
			}

			syncTaskItemRepo.Create(taskItem)

			logger.Logger.Info("开始同步", zap.String("taskName", taskItem.TaskName))
			err := task.Executor(ctx)
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
	}

	// NOTE: v2.11.0 暂不同步支付方式：全量同步
	// if paymentDataChecked {
	// 	taskItem := &model.SyncTaskItem{
	// 		SyncTaskUuid: syncTask.Uuid,
	// 		TaskType:     constant.SyncTaskTypePaymentMethod,
	// 		TaskName:     constant.SyncTaskTypeNames[constant.SyncTaskTypePaymentMethod],
	// 		Status:       constant.SyncTaskItemStatusRunning,
	// 		StartTime:    time.Now().Unix(),
	// 	}

	// 	syncTaskItemRepo.Create(taskItem)

	// 	logger.Logger.Info("开始同步", zap.String("taskName", taskItem.TaskName))
	// 	err := s.SyncPaymentMethod(ctx)
	// 	endTime := time.Now().Unix()

	// 	if err != nil {
	// 		failCount++
	// 		logger.Logger.Error("同步失败", zap.String("taskName", taskItem.TaskName), zap.Error(err))
	// 		syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
	// 			"status":        constant.SyncTaskItemStatusFailed,
	// 			"error_message": err.Error(),
	// 			"end_time":      endTime,
	// 		})
	// 	} else {
	// 		successCount++
	// 		logger.Logger.Info("同步成功", zap.String("taskName", taskItem.TaskName))
	// 		syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
	// 			"status":   constant.SyncTaskItemStatusSuccess,
	// 			"end_time": endTime,
	// 		})
	// 	}
	// }

	// 最后：执行多语言同步
	taskItem := &model.SyncTaskItem{
		SyncTaskUuid: syncTask.Uuid,
		TaskType:     constant.SyncTaskTypeMultiLanguage,
		TaskName:     constant.SyncTaskTypeNames[constant.SyncTaskTypeMultiLanguage],
		Status:       constant.SyncTaskItemStatusRunning,
		StartTime:    time.Now().Unix(),
	}

	syncTaskItemRepo.Create(taskItem)

	logger.Logger.Info("开始同步", zap.String("taskName", taskItem.TaskName))
	err := s.SyncMultiLanguage(ctx)
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

	// 更新主任务状态
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

// SyncMarketingCoupon 全量同步优惠券（先硬删除子店中的总部数据，再全量同步）
func (s *SyncSrv) SyncMarketingCoupon(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 硬删除子店中已有的总部优惠券
	if err := subShopDB.Unscoped().
		Where("headquarter_uuid = ?", headquarterUuid).
		Delete(&model.MarketingCoupon{}).Error; err != nil {
		return errors.WithMessage(err, "硬删除子店中的总部优惠券失败")
	}

	// 查询总部所有优惠券
	var hqCoupons []model.MarketingCoupon
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqCoupons).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部优惠券失败")
	}

	// 同步到分店
	for _, hqCoupon := range hqCoupons {
		newCoupon := hqCoupon
		newCoupon.HeadquarterUuid = headquarterUuid
		newCoupon.ID = 0

		err = subShopDB.Create(&newCoupon).Error
		if err != nil {
			logger.Logger.Error("同步优惠券失败", zap.Uint64("uuid", hqCoupon.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncFullReduction 全量同步满额减活动（先硬删除子店中的总部数据，再全量同步）
func (s *SyncSrv) SyncFullReduction(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 硬删除子店中已有的总部满额减活动（包括规则）
	var existingActivities []model.FullReductionActivity
	subShopDB.Where("headquarter_uuid = ?", headquarterUuid).Find(&existingActivities)
	for _, activity := range existingActivities {
		subShopDB.Unscoped().Where("full_reduction_activity_uuid = ?", activity.Uuid).Delete(&model.FullReductionActivityRule{})
	}
	if err := subShopDB.Unscoped().
		Where("headquarter_uuid = ?", headquarterUuid).
		Delete(&model.FullReductionActivity{}).Error; err != nil {
		return errors.WithMessage(err, "硬删除子店中的总部满额减活动失败")
	}

	// 查询总部所有满额减活动（包含规则）
	var hqActivities []model.FullReductionActivity
	err := headquarterDB.Preload("Rules").
		Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqActivities).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部满额减活动失败")
	}

	// 同步到分店
	for _, hqActivity := range hqActivities {
		newActivity := hqActivity
		newActivity.HeadquarterUuid = headquarterUuid
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

// SyncProductLabel 全量同步菜品标签（先硬删除子店中的总部数据，再全量同步）
func (s *SyncSrv) SyncProductLabel(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 硬删除子店中已有的总部菜品标签
	if err := subShopDB.Unscoped().
		Where("headquarter_uuid = ?", headquarterUuid).
		Delete(&model.ProductLabel{}).Error; err != nil {
		return errors.WithMessage(err, "硬删除子店中的总部菜品标签失败")
	}

	// 查询总部所有菜品标签
	var hqLabels []model.ProductLabel
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqLabels).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部菜品标签失败")
	}

	// 同步到分店
	for _, hqLabel := range hqLabels {
		newLabel := hqLabel
		newLabel.HeadquarterUuid = headquarterUuid
		newLabel.ID = 0

		err = subShopDB.Create(&newLabel).Error
		if err != nil {
			logger.Logger.Error("同步菜品标签失败", zap.Uint64("uuid", hqLabel.Uuid), zap.Error(err))
			continue
		}
	}

	return nil
}

// SyncMarketingActivity 全量同步营销活动（先硬删除子店中的总部数据，再全量同步）
func (s *SyncSrv) SyncMarketingActivity(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 硬删除子店中已有的总部营销活动（包括奖品）
	var existingActivities []model.MarketingActivity
	subShopDB.Where("headquarter_uuid = ?", headquarterUuid).Find(&existingActivities)
	for _, activity := range existingActivities {
		subShopDB.Unscoped().Where("activity_uuid = ?", activity.Uuid).Delete(&model.MarketingActivityPrize{})
	}
	if err := subShopDB.Unscoped().
		Where("headquarter_uuid = ?", headquarterUuid).
		Delete(&model.MarketingActivity{}).Error; err != nil {
		return errors.WithMessage(err, "硬删除子店中的总部营销活动失败")
	}

	// 查询总部所有营销活动（包含奖品）
	var hqActivities []model.MarketingActivity
	err := headquarterDB.Preload("Prizes").
		Where("delete_time = 0 AND headquarter_uuid = 0").
		Find(&hqActivities).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部营销活动失败")
	}

	// 同步到分店
	for _, hqActivity := range hqActivities {
		newActivity := hqActivity
		newActivity.HeadquarterUuid = headquarterUuid
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

// SyncPaymentMethod 全量同步支付方式（遵循原有特殊逻辑，不删除未勾选的支付数据）
func (s *SyncSrv) SyncPaymentMethod(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	// 查询总部支付方式（排除 code=40 和 code=10）
	var hqPayments []model.PaymentMethod
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
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
		err := subShopDB.Where("payment_name = ? and source = ? AND delete_time = 0", hqPayment.PaymentName, model.PaymentSourceDefault).
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
			HeadquarterUuid: headquarterUuid,
			PaymentName:     hqPayment.PaymentName,
			Name:            hqPayment.Name,
			Code:            newCode,                    // 生成新code
			Source:          model.PaymentSourceDefault, // 1-手动添加
			LogoFileUuid:    0,                          // 固定为0
			Sort:            hqPayment.Sort,             // 排序
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
