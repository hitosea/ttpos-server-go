package application

import (
	contexts "context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	menuApi "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/helper"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/repository"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/google/uuid"
	grabfood "github.com/grab/grabfood-api-sdk-go"
	"go.uber.org/zap"
)

// 全局锁管理器（确保所有 takeoutAppService 实例共享同一套锁）
var (
	globalModifierUpdateLocks      sync.Map // map[merchantID]*sync.Mutex
	globalModifierUpdateLocksGuard sync.Mutex

	// 菜单同步防抖管理器
	menuSyncDebounceTimers  sync.Map // map[syncKey]*time.Timer
	menuSyncDebounceMutexes sync.Map // map[syncKey]*sync.Mutex
)

// ITakeoutAppService 外卖应用服务接口
type ITakeoutAppService interface {
	// 处理门店集成状态变更
	HandleIntegrationStatus(ctx context.Context, takeoutIntegrationEvent request.TakeoutIntegrationEvent) error

	// GetTakeoutStatus 获取指定平台外卖状态
	GetTakeoutStatus(ctx context.Context, platform string) (*response.TakeoutStatusResponse, error)

	// ToggleTakeoutStatus 切换指定平台外卖状态
	ToggleTakeoutStatus(ctx context.Context, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error)

	// UpdateTakeoutMenu 更新指定平台的菜单数据
	UpdateTakeoutMenu(ctx context.Context, platform string, menu interface{}) error

	// 外卖绑定管理
	// GetBindingLink 获取绑定链接
	GetBindingLink(ctx context.Context, platform string) (*response.BindingLinkResponse, error)

	// CheckBindingStatus 检查绑定状态 - grab
	CheckBindingStatus(ctx context.Context, platform string) (*response.BindingStatusResponse, error)

	// BindPlatform 绑定平台
	BindPlatform(ctx context.Context, uuid uint64) error

	// UnbindPlatform 解绑平台
	UnbindPlatform(ctx context.Context, uuid uint64) error

	// 外卖菜单管理
	// ExportMenu 导出菜单到指定平台格式
	ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error)

	// ConvertMenuData 转换菜单数据
	ConvertMenuData(ctx context.Context, req request.ImportMenuRequest) (*grabfood.GetMenuNewResponse, error)

	// PushMenu 推送菜单
	PushMenu(ctx context.Context, platform string, currencyUnit string) error

	// GetGrabMenu 获取 Grab 的商品菜单
	GetGrabMenu(ctx context.Context) (*response.GrabMenuResponse, error)

	// 外卖导入进度管理
	// GetImportProgress 获取导入进度
	GetImportProgress(ctx context.Context, platform string) (*response.ImportProgressResponse, error)

	// GetImportLogs 获取导入日志列表
	GetImportLogs(ctx context.Context, req request.GetImportLogsRequest) (*response.ImportLogListResponse, error)

	// SyncMenuChanges 同步菜单变更（灰度更新）
	SyncMenuChanges(ctx context.Context, req request.ExportMenuRequest) (*response.MenuSyncResult, error)
}

// takeoutAppService 外卖应用服务实现
type takeoutAppService struct {
	// 状态管理相关
	takeoutService service.TakeoutService

	// RPC 调用相关
	rpcService *rpcAdapter.TakeoutRPCService

	// 菜单管理相关
	dbm        *database.DBManager
	menuRepo   repository.IMenuDataRepository
	converters map[string]service.IPlatformConverter // 平台菜单转换器映射

	// 订单管理相关
	orderService service.ITakeoutOrderSrv // 订单服务
}

// NewTakeoutAppService 创建外卖应用服务
func NewTakeoutAppService(
	dbm *database.DBManager,
) ITakeoutAppService {
	// 初始化 RPC 服务
	rpcService := rpcAdapter.NewTakeoutRPCService()

	// 初始化平台转换器（用于菜单转换）
	converters := make(map[string]service.IPlatformConverter)
	grabConverter := grab.NewGrabConverter(dbm)
	converters["grab"] = grabConverter
	converters["lineman"] = grabConverter

	// 初始化订单服务
	orderService := service.NewTakeoutOrderSrv(dbm)

	return &takeoutAppService{
		// 状态管理相关
		takeoutService: service.NewTakeoutService(nil),

		// RPC 调用相关
		rpcService: rpcService,

		// 菜单管理相关
		dbm:        dbm,
		menuRepo:   persistence.NewMenuDataRepository(dbm),
		converters: converters,

		// 订单管理相关
		orderService: orderService,
	}
}

// HandleIntegrationStatus 处理门店集成状态变更
func (s *takeoutAppService) HandleIntegrationStatus(ctx context.Context, takeoutIntegrationEvent request.TakeoutIntegrationEvent) error {
	// 更新门店集成状态
	_, err := s.CheckBindingStatus(ctx, "grab")
	if err != nil {
		return fmt.Errorf("检查绑定状态失败: %w", err)
	}
	return nil
}

// GetTakeoutStatus 获取指定平台外卖状态
func (s *takeoutAppService) GetTakeoutStatus(ctx context.Context, platform string) (*response.TakeoutStatusResponse, error) {
	// 从数据库获取
	takeout, err := s.takeoutService.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}

	resp := &response.TakeoutStatusResponse{
		Platform:  takeout.Platform,
		Enabled:   takeout.Enabled,
		IsBound:   takeout.IsBound,
		Skip:      takeout.Skip,
		UpdatedAt: takeout.UpdateTime,
	}

	return resp, nil
}

// ToggleTakeoutStatus 切换指定平台外卖状态
func (s *takeoutAppService) ToggleTakeoutStatus(ctx context.Context, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error) {
	// 如果开启，则推送菜单到LINE MAN
	if req.Enabled && req.Platform == value_object.TakeoutPlatformLineman {
		takeout, err := s.takeoutService.GetByPlatform(ctx, req.Platform)
		if err != nil {
			return nil, fmt.Errorf("获取平台状态失败: %w", err)
		}
		// 如果开启，则推送菜单到LINE MAN
		err = s.rpcService.ActivateShop(ctx.GetContext(), value_object.TakeoutPlatformLineman, ctx.GetCompanyUuid())
		if err != nil {
			return nil, errors.WithMessage(errors.New("推送菜单到LINE MAN失败"), err.Error())
		}
		err = s.takeoutService.UpdatePlatformBoundStatus(ctx, takeout.Uuid, true)
		if err != nil {
			return nil, fmt.Errorf("更新平台状态失败: %w", err)
		}
	}
	// 更新状态
	err := s.takeoutService.UpdatePlatformStatusByPlatform(ctx, req.Platform, req.Enabled)
	if err != nil {
		return nil, fmt.Errorf("更新平台状态失败: %w", err)
	}
	// 返回最新状态
	return s.GetTakeoutStatus(ctx, req.Platform)
}

// UpdateTakeoutMenu 更新指定平台的菜单数据
func (s *takeoutAppService) UpdateTakeoutMenu(ctx context.Context, platform string, menu interface{}) error {
	err := s.takeoutService.UpdatePlatformMenuByPlatform(ctx, platform, menu)
	if err != nil {
		return fmt.Errorf("更新平台菜单失败: %w", err)
	}

	return nil
}

// GetBindingLink 获取绑定链接
func (s *takeoutAppService) GetBindingLink(ctx context.Context, platform string) (*response.BindingLinkResponse, error) {
	// 1. 先从数据库查询缓存的绑定链接
	takeout, err := s.takeoutService.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}
	if !takeout.IsEnabled() {
		return nil, errors.New("平台未开启")
	}

	// 当前无Grab商品，无法推送订单
	menu, err := s.ExportMenu(ctx, request.ExportMenuRequest{
		Platform:     platform,
		CurrencyUnit: "USD",
	})
	if err != nil {
		return nil, errors.New("获取菜单失败")
	}
	if menu == nil || reflect.ValueOf(menu).IsNil() {
		return nil, errors.NewWithCodeAndData(-1001, nil, "当前无Grab商品，无法推送订单")
	}
	// 类型断言为 Grab 菜单格式
	grabMenu, ok := menu.(*grabfood.GetMenuNewResponse)
	if !ok {
		return nil, errors.New("菜单数据格式错误")
	}
	// 判断 menu 的 categories 是否为空
	if len(grabMenu.GetCategories()) == 0 {
		return nil, errors.NewWithCodeAndData(-1001, nil, "当前无Grab商品，无法推送订单")
	}

	// 2. 如果缓存不存在，调用 RPC 获取
	companyUuid := ctx.GetCompanyUuid()
	bindingLink, err := s.rpcService.GetGrabBindingLink(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.New("获取绑定链接失败")
	}

	// 3. 保存到数据库缓存
	if err := s.takeoutService.UpdatePlatformBindingLink(ctx, takeout.Uuid, bindingLink); err != nil {
		if takeout.BindingLink != "" {
			bindingLink = takeout.BindingLink
		} else {
			logger.Logger.Error("认证失败，请核对信息", zap.Error(err))
			return nil, errors.New("认证失败，请核对信息")
		}
	}

	return &response.BindingLinkResponse{
		BindingLink: bindingLink,
	}, nil
}

// CheckBindingStatus 检查绑定状态
func (s *takeoutAppService) CheckBindingStatus(ctx context.Context, platform string) (*response.BindingStatusResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	takeout, err := s.takeoutService.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}
	if !takeout.Enabled {
		return nil, errors.New("平台未开启")
	}
	if takeout.IsBound {
		return &response.BindingStatusResponse{
			IsBound: true,
		}, nil
	}

	// 调用 bmp RPC 接口检查绑定状态
	isBound, err := s.rpcService.CheckBindingStatus(ctx.GetContext(), platform, companyUuid)
	if err != nil {
		return nil, fmt.Errorf("检查绑定状态失败: %w", err)
	}

	if isBound {
		s.takeoutService.UpdatePlatformBoundStatus(ctx, takeout.Uuid, true)
	}

	return &response.BindingStatusResponse{
		IsBound: isBound,
	}, nil
}

// BindPlatform 绑定平台
func (s *takeoutAppService) BindPlatform(ctx context.Context, uuid uint64) error {
	err := s.takeoutService.UpdatePlatformBoundStatus(ctx, uuid, true)
	if err != nil {
		return fmt.Errorf("绑定平台失败: %w", err)
	}

	return nil
}

// UnbindPlatform 解绑平台
func (s *takeoutAppService) UnbindPlatform(ctx context.Context, uuid uint64) error {
	err := s.takeoutService.UpdatePlatformBoundStatus(ctx, uuid, false)
	if err != nil {
		return fmt.Errorf("解绑平台失败: %w", err)
	}

	return nil
}

// ExportMenu 导出菜单到指定平台格式
func (s *takeoutAppService) ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error) {
	companyUuid := ctx.GetCompanyUuid()
	// 验证参数
	if req.Platform == "" {
		return nil, errors.New("平台名称不能为空")
	}
	if companyUuid == 0 {
		return nil, errors.New("公司 UUID 不能为空")
	}

	// 获取平台转换器
	converter, err := s.getConverter(req.Platform)
	if err != nil {
		return nil, err
	}

	// 从数据库加载菜单数据并转换为平台格式
	var platformData interface{}
	var singleFlavorMapping map[string]*value_object.SingleFlavorInfo
	if grabConverter, ok := converter.(*grab.GrabConverter); ok {
		// Grab 平台使用专用的加载方法（直接返回 Grab 格式）
		if req.Platform == value_object.TakeoutPlatformLineman {
			grabConverter.SetNameMaxLength(5000)
		} else {
			grabConverter.SetNameMaxLength(40)
		}
		platformData, err = grabConverter.LoadMenuFromDatabase(ctx, req.Platform, companyUuid, req.CurrencyUnit, []uint64{})
		if err != nil {
			return nil, errors.New("加载菜单数据失败")
		}
		// 获取单规格映射
		singleFlavorMapping = grabConverter.GetSingleFlavorMapping()
	} else {
		return nil, errors.New("暂不支持该平台")
	}

	// 保存导出的菜单数据到 ttpos_menu 字段
	if err := s.takeoutService.UpdateTtposMenuByPlatform(ctx, req.Platform, platformData); err != nil {
		// 仅记录日志，不影响主流程
		logger.Logger.Error("保存TTPOS导出菜单失败", zap.Error(err), zap.String("platform", req.Platform))
	}

	// 保存单规格映射到 single_flavor_mapping 字段
	if req.IsSaveSingleFlavorMapping {
		// 任务 #39752: Grab/LINE MAN-单规格不推送规格内的选项
		// 注意：无论有没有单规格映射，都更新数据库（覆盖旧数据），避免残留历史数据
		if err := s.takeoutService.UpdateSingleFlavorMappingByPlatform(ctx, req.Platform, singleFlavorMapping); err != nil {
			logger.Logger.Error("保存单规格映射失败", zap.Error(err), zap.String("platform", req.Platform))
		} else if len(singleFlavorMapping) > 0 {
			logger.Logger.Info("保存单规格映射成功",
				zap.String("platform", req.Platform),
				zap.Int("count", len(singleFlavorMapping)))
		}
	}

	return platformData, nil
}

// GetGrabMenu 获取 Grab 的商品菜单
func (s *takeoutAppService) GetGrabMenu(ctx context.Context) (*response.GrabMenuResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	menuData, err := s.rpcService.GetGrabMenu(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, fmt.Errorf("获取 Grab 菜单失败: %w", err)
	}

	var menu *grabfood.GetMenuNewResponse
	if err := json.Unmarshal([]byte(menuData.(string)), &menu); err != nil {
		return nil, fmt.Errorf("解析 Grab 菜单数据失败: %w", err)
	}

	// 更新 ttpos_takeout 表的 menu 字段
	err = s.takeoutService.UpdatePlatformMenuByPlatform(ctx, "grab", menu)
	if err != nil {
		return nil, fmt.Errorf("更新 TTPOS 菜单失败: %w", err)
	}

	return &response.GrabMenuResponse{
		Platform: "grab",
		Menu:     menu,
	}, nil
}

// ConvertMenuData 转换菜单数据
func (s *takeoutAppService) ConvertMenuData(ctx context.Context, req request.ImportMenuRequest) (*grabfood.GetMenuNewResponse, error) {
	if req.MenuData == nil {
		return nil, errors.New("菜单数据不能为空")
	}

	// 获取 Grab 转换器
	converter, err := s.getConverter(req.Platform)
	if err != nil {
		return nil, err
	}
	grabConverter, ok := converter.(*grab.GrabConverter)
	if !ok {
		return nil, errors.New("转换器类型错误")
	}

	// 解析 Grab 菜单数据（返回 Grab 格式）
	grabMenu, err := grabConverter.ParseGrabMenu(req.MenuData)
	if err != nil {
		return nil, fmt.Errorf("解析 Grab 菜单失败: %w", err)
	}

	// 保存菜单数据到数据库
	err = s.UpdateTakeoutMenu(ctx, "grab", grabMenu)
	if err != nil {
		return nil, fmt.Errorf("保存菜单数据失败: %w", err)
	}

	return grabMenu, nil
}

// SaveMenuSnapshot 保存菜单快照
func (s *takeoutAppService) PushMenu(ctx context.Context, platform string, currencyUnit string) error {
	platform = strings.ToLower(platform)
	companyUuid := ctx.GetCompanyUuid()

	// 判断是否开发模式，如果是开发模式，则使用 6293997752320000
	if config.Server.Mode == "debug" && platform == value_object.TakeoutPlatformLineman {
		if config.Takeout.TakeoutLinemanStoreId != 0 {
			companyUuid = config.Takeout.TakeoutLinemanStoreId
		}
	}

	menu, err := s.ExportMenu(ctx, request.ExportMenuRequest{
		Platform:                  platform,
		CurrencyUnit:              currencyUnit,
		IsSaveSingleFlavorMapping: true,
	})
	if err != nil {
		return fmt.Errorf("导出菜单失败: %w", err)
	}

	err = s.rpcService.SaveMenuSnapshot(ctx.GetContext(), platform, companyUuid, menu)
	if err != nil {
		return fmt.Errorf("推送菜单失败: %w", err)
	}

	return nil
}

// getConverter 获取平台转换器
func (s *takeoutAppService) getConverter(platform string) (service.IPlatformConverter, error) {
	converter, ok := s.converters[platform]
	if !ok {
		return nil, errors.New("不支持的平台: " + platform)
	}
	return converter, nil
}

// GetImportProgress 获取导入进度
func (s *takeoutAppService) GetImportProgress(ctx context.Context, platform string) (*response.ImportProgressResponse, error) {
	// 初始化 repository 和 service
	db := ctx.GetDB()
	importProgressSrv := service.NewImportProgressService(db)

	// 获取最新的导入进度
	progressInfo, err := importProgressSrv.GetLatestProgressByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取导入进度失败: %w", err)
	}

	if progressInfo == nil {
		// 没有导入记录
		return &response.ImportProgressResponse{
			Platform: platform,
			Status:   -1, // 表示无导入记录
		}, nil
	}

	// 转换为响应 DTO
	result := &response.ImportProgressResponse{
		UUID:            progressInfo.UUID,
		Platform:        progressInfo.Platform,
		ImportType:      progressInfo.ImportType,
		ImportDirection: progressInfo.ImportDirection,
		Status:          progressInfo.Status,
		Progress:        progressInfo.Progress,
		SuccessCount:    progressInfo.SuccessCount,
		FailureCount:    progressInfo.FailureCount,
		TotalCount:      progressInfo.TotalCount,
		ErrorMessage:    progressInfo.ErrorMessage,
		StartTime:       progressInfo.StartTime,
		EndTime:         progressInfo.EndTime,
		Duration:        progressInfo.Duration,
		EstimatedTime:   progressInfo.EstimatedTime,
	}

	// 设置当前步骤描述
	if progressInfo.Status == 0 { // ImportStatusInProgress
		if progressInfo.Progress < 50 {
			result.CurrentStep = "正在同步分类..."
		} else {
			result.CurrentStep = "正在同步商品..."
		}
	}

	return result, nil
}

// GetImportLogs 获取导入日志列表
func (s *takeoutAppService) GetImportLogs(ctx context.Context, req request.GetImportLogsRequest) (*response.ImportLogListResponse, error) {
	// 初始化 repository 和 service
	db := ctx.GetDB()
	importProgressSrv := service.NewImportProgressService(db)

	// 设置默认分页参数
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 调用 Domain Service 获取日志列表
	var importType int8 = 0
	if req.ImportType != nil {
		importType = *req.ImportType
	}

	var status int8 = -1
	if req.Status != nil {
		status = *req.Status
	}

	logs, total, err := importProgressSrv.ListImportLogs(
		ctx,
		req.Platform,
		importType,
		status,
		req.PageNo,
		req.PageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("获取导入日志列表失败: %w", err)
	}

	// 转换为响应 DTO
	logList := make([]response.ImportLogResponse, 0, len(logs))
	for i, log := range logs {
		// 判断是否可以重新导入：
		// 1. 导入失败
		// 2. 是第一条记录
		// 3. 是从平台导入到TTPOS (importType=2)
		// 4. 平台是 grab
		canReimport := log.IsFailed() &&
			i == 0 &&
			log.ImportType == model.ImportTypePlatformToTTPOS &&
			log.Platform == "grab"

		logList = append(logList, response.ImportLogResponse{
			UUID:            log.Uuid,
			Platform:        log.GetPlatformName(),
			ImportType:      log.ImportType,
			ImportDirection: log.ImportDirection,
			Status:          log.Status,
			Progress:        log.Progress,
			SuccessCount:    log.SuccessCount,
			FailureCount:    log.FailureCount,
			TotalCount:      log.TotalCount,
			ErrorMessage:    log.ErrorMessage,
			StartTime:       log.StartTime,
			EndTime:         log.EndTime,
			Duration:        log.Duration,
			CreateTime:      log.CreateTime,
			CanReimport:     canReimport,
		})
	}

	return &response.ImportLogListResponse{
		List: logList,
		PageInfo: response.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// SyncMenuChanges 同步菜单变更（灰度更新）
// 实现防抖机制：在2秒内如果有相同同步请求，则取消旧请求的执行
func (s *takeoutAppService) SyncMenuChanges(ctx context.Context, req request.ExportMenuRequest) (*response.MenuSyncResult, error) {
	// 启动异步防抖处理
	utils.Go(func() {
		s.handleDebouncedMenuSync(ctx.Copy(), req)
	})

	// 立即返回初始结果，表示已接受请求
	return &response.MenuSyncResult{
		ItemChanges:     []response.MenuItemChange{},
		ModifierChanges: []response.MenuModifierChange{},
		Errors:          []string{},
	}, nil
}

// handleDebouncedMenuSync 处理防抖菜单同步
// 优化后：使用 Timer 机制，避免阻塞，减少 Redis 操作，降低锁竞争
func (s *takeoutAppService) handleDebouncedMenuSync(ctx context.Context, req request.ExportMenuRequest) {
	// 生成同步键：platform_companyUuid
	syncKey := fmt.Sprintf("menu_sync_%s_%d", req.Platform, req.CompanyUuid)

	// 获取或创建该 syncKey 的专用锁
	muInterface, _ := menuSyncDebounceMutexes.LoadOrStore(syncKey, &sync.Mutex{})
	mu := muInterface.(*sync.Mutex)

	mu.Lock()
	defer mu.Unlock()

	// 如果存在旧的 timer，取消它
	if oldTimer, exists := menuSyncDebounceTimers.Load(syncKey); exists {
		oldTimer.(*time.Timer).Stop()
		logger.Logger.Debug("取消旧的菜单同步任务",
			zap.String("sync_key", syncKey),
			zap.String("platform", req.Platform),
			zap.Uint64("company_uuid", req.CompanyUuid))
	}

	// 创建新的 timer，2秒后执行同步
	// 注意：timer 回调在独立 goroutine 中执行，不能复用原请求的 ctx（其数据库连接已关闭）
	// 需要创建新的 context 并重新获取数据库连接
	companyUuid := req.CompanyUuid
	timer := time.AfterFunc(1*time.Second, func() {
		// 重新获取数据库连接
		newCtx := ctx.Copy()
		newCtx.SetDB(s.dbm.GetDB(companyUuid))

		_, err := s.directSyncMenuChanges(newCtx, req)
		if err != nil {
			logger.Logger.Error("防抖菜单同步失败",
				zap.Error(err),
				zap.String("sync_key", syncKey),
				zap.String("platform", req.Platform),
				zap.Uint64("company_uuid", req.CompanyUuid))
		}
		// 清理 timer
		menuSyncDebounceTimers.Delete(syncKey)
	})

	// 保存新的 timer
	menuSyncDebounceTimers.Store(syncKey, timer)
}

// directSyncMenuChanges 直接执行菜单变更同步（不含防抖逻辑）
func (s *takeoutAppService) directSyncMenuChanges(ctx context.Context, req request.ExportMenuRequest) (*response.MenuSyncResult, error) {
	// 初始化结果
	result := &response.MenuSyncResult{
		ItemChanges:     []response.MenuItemChange{},
		ModifierChanges: []response.MenuModifierChange{},
		Errors:          []string{},
	}

	// 1. 获取旧菜单（从 ttpos_menu 字段）
	takeout, err := s.takeoutService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}
	if takeout.Enabled == false || takeout.IsBound == false {
		logger.Logger.Info("平台未开启或未绑定，跳过同步", zap.String("platform", req.Platform))
		return result, nil
	}

	// 2. 导出最新菜单
	newMenuData, err := s.ExportMenu(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("导出最新菜单失败: %w", err)
	}

	// 如果没有旧菜单数据，直接保存新菜单并返回
	if takeout.TtposMenu == nil || reflect.ValueOf(takeout.TtposMenu).IsNil() {
		logger.Logger.Info("首次同步菜单，无旧数据对比", zap.String("platform", req.Platform))
		return result, nil
	}

	// 获取平台转换器
	converter, err := s.getConverter(req.Platform)
	if err != nil {
		return nil, err
	}

	grabConverter, ok := converter.(*grab.GrabConverter)
	if !ok {
		return nil, errors.New("转换器类型错误")
	}

	// 3. 解析新旧菜单数据
	newMenu, err := grabConverter.ParseGrabMenu(newMenuData)
	if err != nil {
		return nil, fmt.Errorf("解析新菜单数据失败: %w", err)
	}

	oldMenu, err := grabConverter.ParseGrabMenu(takeout.TtposMenu)
	if err != nil {
		return nil, fmt.Errorf("解析旧菜单数据失败: %w", err)
	}

	// 4. 比较并同步变更
	err = s.compareAndSyncMenu(ctx, req.Platform, req.CompanyUuid, oldMenu, newMenu, result)
	if err != nil {
		return nil, fmt.Errorf("同步菜单变更失败: %w", err)
	}

	// 5. 记录同步结果日志
	logger.Logger.Info("菜单同步完成",
		zap.String("platform", req.Platform),
		zap.Int("totalItems", result.TotalItems),
		zap.Int("totalModifiers", result.TotalModifiers),
		zap.Int("changedItems", result.ChangedItems),
		zap.Int("changedModifiers", result.ChangedModifiers),
		zap.Int("successItems", result.SuccessItems),
		zap.Int("successModifiers", result.SuccessModifiers),
		zap.Int("failedItems", result.FailedItems),
		zap.Int("failedModifiers", result.FailedModifiers),
		zap.Strings("errors", result.Errors))

	return result, nil
}

// compareAndSyncMenu 比较并同步菜单变更
func (s *takeoutAppService) compareAndSyncMenu(
	ctx context.Context,
	platform string,
	companyUuid uint64,
	oldMenu *grabfood.GetMenuNewResponse,
	newMenu *grabfood.GetMenuNewResponse,
	result *response.MenuSyncResult,
) error {
	// 创建旧菜单的索引
	oldItemsMap := make(map[string]*grabfood.MenuItem)
	oldModifiersMap := make(map[string]*grabfood.MenuModifier)

	for _, category := range oldMenu.Categories {
		for i := range category.Items {
			item := &category.Items[i]
			oldItemsMap[item.Id] = item

			// 索引修饰符
			for _, modGroup := range item.ModifierGroups {
				for j := range modGroup.Modifiers {
					modifier := &modGroup.Modifiers[j]
					oldModifiersMap[modifier.Id] = modifier
				}
			}
		}
	}

	// 收集需要更新的商品和修饰符
	var changedItems []struct {
		old *grabfood.MenuItem
		new *grabfood.MenuItem
	}
	var changedModifiers []struct {
		old *grabfood.MenuModifier
		new *grabfood.MenuModifier
	}

	// 遍历新菜单，比较并收集变更
	for _, category := range newMenu.Categories {
		for i := range category.Items {
			newItem := &category.Items[i]
			result.TotalItems++

			// 比较商品
			if oldItem, exists := oldItemsMap[newItem.Id]; exists {
				if helper.HasMenuItemChanged(oldItem, newItem) {
					changedItems = append(changedItems, struct {
						old *grabfood.MenuItem
						new *grabfood.MenuItem
					}{old: oldItem, new: newItem})
					result.ChangedItems++
				}
			}

			// 比较修饰符
			for _, modGroup := range newItem.ModifierGroups {
				for j := range modGroup.Modifiers {
					newModifier := &modGroup.Modifiers[j]
					result.TotalModifiers++

					if oldModifier, exists := oldModifiersMap[newModifier.Id]; exists {
						if helper.HasMenuModifierChanged(oldModifier, newModifier) {
							changedModifiers = append(changedModifiers, struct {
								old *grabfood.MenuModifier
								new *grabfood.MenuModifier
							}{old: oldModifier, new: newModifier})
							result.ChangedModifiers++
						}
					}
				}
			}
		}
	}

	// 获取商户 ID
	shopUuidStr := strconv.FormatUint(companyUuid, 10)

	// 判断是否开发模式，如果是开发模式，则使用 6293997752320000
	if config.Server.Mode == "debug" && platform == value_object.TakeoutPlatformLineman {
		if config.Takeout.TakeoutLinemanStoreId != 0 {
			shopUuidStr = strconv.FormatUint(config.Takeout.TakeoutLinemanStoreId, 10)
		}
	}

	// 批量更新商品（每批最多 100 个）
	if len(changedItems) > 0 {
		err := s.batchUpdateItems(ctx, platform, shopUuidStr, changedItems, result)
		if err != nil {
			logger.Logger.Error("批量更新商品失败", zap.Error(err))
		}
	}

	// 逐个更新修饰符（Grab API 暂不支持批量更新修饰符）
	if len(changedModifiers) > 0 {
		s.updateModifiersOneByOne(ctx, platform, shopUuidStr, changedModifiers, result)
	}

	return nil
}

// batchUpdateItems 批量更新商品
func (s *takeoutAppService) batchUpdateItems(
	ctx context.Context,
	platform string,
	shopUuid string,
	items []struct {
		old *grabfood.MenuItem
		new *grabfood.MenuItem
	},
	result *response.MenuSyncResult,
) error {
	const batchSize = 100

	// 创建 RPC 客户端
	client, err := s.rpcService.GetBMPClient()
	if err != nil {
		logger.Logger.Error("创建 RPC 客户端失败", zap.Error(err))
		return fmt.Errorf("创建 RPC 客户端失败: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Logger.Warn("关闭 RPC 客户端失败", zap.Error(closeErr))
		}
	}()

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		// 跳过空批次
		if len(batch) == 0 {
			continue
		}

		// 构建批量更新请求
		entities := make([]*menuApi.MenuEntity, 0, len(batch))

		for _, item := range batch {
			// 计算库存
			maxStock := int64(9999999)
			if item.new.AvailableStatus != string(value_object.AvailableStatusAvailable) {
				maxStock = 0
			}
			entity := &menuApi.MenuEntity{
				Id:              item.new.Id,
				AvailableStatus: &item.new.AvailableStatus,
			}
			if platform != value_object.TakeoutPlatformLineman {
				entity.Price = &item.new.Price
				entity.MaxStock = &maxStock
			}
			entities = append(entities, entity)
		}

		// 再次检查：确保 entities 不为空
		if len(entities) == 0 {
			logger.Logger.Warn("批量更新商品: entities 为空，跳过此批次",
				zap.Int("batchIndex", i/batchSize),
				zap.Int("batchSize", len(batch)))
			continue
		}

		// 调用 RPC 批量更新接口
		req := &menuApi.BatchUpdateMenuReq{
			ShopUuid:     shopUuid,
			Field:        "ITEM",
			MenuEntities: entities,
			RequestId:    uuid.New().String(),
			ProviderName: &platform,
		}

		// 设置超时时间
		newCtx, cancel := contexts.WithTimeout(contexts.Background(), 30*time.Second)
		defer cancel()
		resp, err := client.GetMenuClient().BatchUpdateMenu(newCtx, req)

		// 处理响应并记录结果
		for _, item := range batch {
			var changeErr error
			var changes []string

			if item.old.Price != item.new.Price {
				changes = append(changes, "price")
			}
			if item.old.AvailableStatus != item.new.AvailableStatus {
				changes = append(changes, "status")
			}

			change := response.MenuItemChange{
				ItemID:     item.new.Id,
				ItemName:   item.new.Name,
				ChangeType: strings.Join(changes, ","),
				OldPrice:   &item.old.Price,
				NewPrice:   &item.new.Price,
				OldStatus:  item.old.AvailableStatus,
				NewStatus:  item.new.AvailableStatus,
			}

			if err != nil {
				// 批量调用失败，标记所有为失败
				changeErr = err
			} else if resp.Code != "0" {
				// 业务错误
				changeErr = fmt.Errorf("业务错误: %s", resp.Message)
			}

			if changeErr != nil {
				change.Success = false
				change.ErrorMessage = changeErr.Error()
				result.FailedItems++
				result.Errors = append(result.Errors,
					fmt.Sprintf("商品%s(%s)更新失败: %v", item.new.Name, item.new.Id, changeErr))
			} else {
				change.Success = true
				result.SuccessItems++
			}

			result.ItemChanges = append(result.ItemChanges, change)
		}
	}

	return nil
}

// updateModifiersOneByOne 逐个串行更新修饰符
// 注意：Grab API 暂不支持批量更新修饰符，只能逐个调用
// 采用简单串行处理 + 请求间隔 + 重试机制，避免触发频率限制
// 通过商户级互斥锁确保同一商户的修饰符更新排队执行（不同商户之间可并发）
// FIXME: 让6哥弄队列处理，6哥说后面时间空了再说
func (s *takeoutAppService) updateModifiersOneByOne(
	ctx context.Context,
	platform string,
	merchantID string,
	modifiers []struct {
		old *grabfood.MenuModifier
		new *grabfood.MenuModifier
	},
	result *response.MenuSyncResult,
) {
	// 获取或创建当前商户的互斥锁（使用全局锁映射，确保跨实例共享）
	globalModifierUpdateLocksGuard.Lock()
	mutexInterface, loaded := globalModifierUpdateLocks.Load(merchantID)
	if !loaded {
		mutexInterface = &sync.Mutex{}
		globalModifierUpdateLocks.Store(merchantID, mutexInterface)
	}
	globalModifierUpdateLocksGuard.Unlock()
	merchantMutex := mutexInterface.(*sync.Mutex)

	// 获取商户级互斥锁，确保同一商户同一时间只有一个修饰符更新任务在执行
	// 不同商户之间可以并发，避免阻塞其他商户
	merchantMutex.Lock()
	defer func() {
		merchantMutex.Unlock()
	}()

	// 频率控制配置
	const requestInterval = 350 // 每个请求间隔 300ms，约 3 req/s
	const maxRetries = 3        // 遇到 429 时的最大重试次数
	const retryDelay = 600      // 重试基础延迟 300ms

	if len(modifiers) == 0 {
		return
	}

	// 串行处理每个修饰符
	for _, modifier := range modifiers {
		// 计算变更类型
		var changes []string
		if !helper.IsInt64PtrEqual(modifier.old.Price, modifier.new.Price) {
			changes = append(changes, "price")
		}
		if modifier.old.AvailableStatus != modifier.new.AvailableStatus {
			changes = append(changes, "status")
		}

		// 带重试的 RPC 调用
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				// 重试前等待（指数退避）
				retryWait := time.Duration(retryDelay*attempt) * time.Millisecond
				logger.Logger.Warn("修饰符更新失败，准备重试",
					zap.String("modifierId", modifier.new.Id),
					zap.String("modifierName", modifier.new.Name),
					zap.Int("attempt", attempt),
					zap.Duration("retryWait", retryWait))
				time.Sleep(retryWait)
			}

			// 调用 RPC 更新单个修饰符
			err = s.rpcService.UpdateMenuModifier(
				ctx,
				platform,
				merchantID,
				modifier.new.Id,
				modifier.new.Name,
				modifier.new.Price,
				modifier.new.AvailableStatus,
			)

			// 如果成功或者不是 429 错误，跳出重试循环
			if err == nil || !strings.Contains(err.Error(), "429") {
				break
			}
		}

		// 记录变更详情
		change := response.MenuModifierChange{
			ModifierID:   modifier.new.Id,
			ModifierName: modifier.new.Name,
			ChangeType:   strings.Join(changes, ","),
			OldPrice:     modifier.old.Price,
			NewPrice:     modifier.new.Price,
			OldStatus:    modifier.old.AvailableStatus,
			NewStatus:    modifier.new.AvailableStatus,
			Success:      err == nil,
		}

		if err != nil {
			change.ErrorMessage = err.Error()
			result.FailedModifiers++
			result.Errors = append(result.Errors,
				fmt.Sprintf("修饰符%s(%s)更新失败: %v", modifier.new.Name, modifier.new.Id, err))
		} else {
			result.SuccessModifiers++
		}
		result.ModifierChanges = append(result.ModifierChanges, change)

		// 请求间隔控制
		time.Sleep(time.Duration(requestInterval) * time.Millisecond)
	}
}
