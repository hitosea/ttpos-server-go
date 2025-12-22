package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"ttpos-server-go/app/modules/takeout/domain/menu/entity"
	"ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/app/modules/takeout/domain/menu/valueobject"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	grabfood "github.com/grab/grabfood-api-sdk-go"
	"go.uber.org/zap"
)

// ITakeoutAppService 外卖应用服务接口
type ITakeoutAppService interface {
	// GetTakeoutStatus 获取指定平台外卖状态
	GetTakeoutStatus(ctx context.Context, platform string) (*response.TakeoutStatusResponse, error)

	// GetAllTakeoutStatus 获取所有平台外卖状态
	GetAllTakeoutStatus(ctx context.Context) (*response.TakeoutStatusListResponse, error)

	// ToggleTakeoutStatus 切换指定平台外卖状态
	ToggleTakeoutStatus(ctx context.Context, platform string, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error)

	// UpdateTakeoutMenu 更新指定平台的菜单数据
	UpdateTakeoutMenu(ctx context.Context, platform string, menu interface{}) error

	// 外卖绑定管理
	// GetBindingLink 获取绑定链接
	GetBindingLink(ctx context.Context, platform string) (*response.BindingLinkResponse, error)

	// CheckBindingStatus 检查绑定状态
	CheckBindingStatus(ctx context.Context, platform string) (*response.BindingStatusResponse, error)

	// UpdateBindingStatus 更新绑定状态
	UpdateBindingStatus(ctx context.Context, req request.UpdateBindingStatusRequest) error

	// BindPlatform 绑定平台
	BindPlatform(ctx context.Context, uuid uint64) error

	// UnbindPlatform 解绑平台
	UnbindPlatform(ctx context.Context, uuid uint64) error

	// 外卖菜单管理
	// ExportMenu 导出菜单到指定平台格式
	ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error)

	// ConvertMenuData 转换菜单数据
	ConvertMenuData(ctx context.Context, req request.ImportMenuRequest) (*entity.TakeoutMenu, error)

	// PushMenu 推送菜单
	PushMenu(ctx context.Context, platform string, currencyUnit string) error

	// GetGrabMenu 获取 Grab 的商品菜单
	GetGrabMenu(ctx context.Context) (*response.GrabMenuResponse, error)

	// 外卖导入进度管理
	// GetImportProgress 获取导入进度
	GetImportProgress(ctx context.Context, platform string) (*response.ImportProgressResponse, error)

	// GetImportLogs 获取导入日志列表
	GetImportLogs(ctx context.Context, req request.GetImportLogsRequest) (*response.ImportLogListResponse, error)

	// UpdateMenuItem 更新菜单项（商品）
	UpdateMenuItem(ctx context.Context, req request.UpdateMenuItemRequest) error

	// UpdateMenuModifier 更新菜单修饰符
	UpdateMenuModifier(ctx context.Context, req request.UpdateMenuModifierRequest) error

	// SyncMenuChanges 同步菜单变更（灰度更新）
	SyncMenuChanges(ctx context.Context, req request.ExportMenuRequest) (*response.MenuSyncResult, error)
}

type ITakeoutMenuAppService = ITakeoutAppService

// takeoutAppService 外卖应用服务实现
type takeoutAppService struct {
	// 状态管理相关
	takeoutDomainService service.TakeoutDomainService

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
	grabConverter := grab.NewGrabConverter(dbm, nil)
	converters["grab"] = grabConverter
	// 后续可添加其他平台：converters["lineman"] = lineman.NewLinemanConverter(dbm)

	// 初始化订单服务
	orderService := service.NewTakeoutOrderSrv(dbm)

	return &takeoutAppService{
		// 状态管理相关
		takeoutDomainService: service.NewTakeoutDomainService(nil),

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

// NewTakeoutMenuAppService 创建外卖菜单应用服务（向后兼容）
func NewTakeoutMenuAppService(
	dbm *database.DBManager,
) ITakeoutMenuAppService {
	return NewTakeoutAppService(dbm)
}

// GetTakeoutStatus 获取指定平台外卖状态
func (s *takeoutAppService) GetTakeoutStatus(ctx context.Context, platform string) (*response.TakeoutStatusResponse, error) {
	// 从数据库获取
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, platform)
	if err != nil {
		// 如果记录不存在，自动创建一条默认记录（不开启，未绑定）
		createdTakeout, createErr := s.takeoutDomainService.CreatePlatformStatus(ctx, platform, false)
		if createErr != nil {
			return nil, fmt.Errorf("获取平台状态失败，且创建默认记录失败: %w", createErr)
		}
		takeout = createdTakeout
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

// GetAllTakeoutStatus 获取所有平台外卖状态
func (s *takeoutAppService) GetAllTakeoutStatus(ctx context.Context) (*response.TakeoutStatusListResponse, error) {
	// 从数据库获取
	takeouts, err := s.takeoutDomainService.GetAllPlatformStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取所有平台状态失败: %w", err)
	}

	list := make([]*response.TakeoutStatusResponse, 0, len(takeouts))
	for _, takeout := range takeouts {
		list = append(list, &response.TakeoutStatusResponse{
			Platform:  takeout.Platform,
			Enabled:   takeout.Enabled,
			IsBound:   takeout.IsBound,
			Skip:      takeout.Skip,
			UpdatedAt: takeout.UpdateTime,
		})
	}

	resp := &response.TakeoutStatusListResponse{
		List: list,
	}

	return resp, nil
}

// ToggleTakeoutStatus 切换指定平台外卖状态
func (s *takeoutAppService) ToggleTakeoutStatus(ctx context.Context, platform string, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error) {
	// 更新状态
	err := s.takeoutDomainService.UpdatePlatformStatusByPlatform(ctx, platform, req.Enabled)
	if err != nil {
		return nil, fmt.Errorf("更新平台状态失败: %w", err)
	}
	// 返回最新状态
	return s.GetTakeoutStatus(ctx, platform)
}

// UpdateTakeoutMenu 更新指定平台的菜单数据
func (s *takeoutAppService) UpdateTakeoutMenu(ctx context.Context, platform string, menu interface{}) error {
	err := s.takeoutDomainService.UpdatePlatformMenuByPlatform(ctx, platform, menu)
	if err != nil {
		return fmt.Errorf("更新平台菜单失败: %w", err)
	}

	return nil
}

// GetBindingLink 获取绑定链接
func (s *takeoutAppService) GetBindingLink(ctx context.Context, platform string) (*response.BindingLinkResponse, error) {
	// 1. 先从数据库查询缓存的绑定链接
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}

	// 2. 如果缓存不存在，调用 RPC 获取
	companyUuid := ctx.GetCompanyUuid()
	bindingLink, err := s.rpcService.GetGrabBindingLink(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, fmt.Errorf("获取绑定链接失败: %w", err)
	}

	// 3. 保存到数据库缓存
	if err := s.takeoutDomainService.UpdatePlatformBindingLink(ctx, takeout.Uuid, bindingLink); err != nil {
		if takeout.BindingLink != "" {
			bindingLink = takeout.BindingLink
		} else {
			logger.Logger.Error("认证失败，请核对信息", zap.Error(err))
			return nil, fmt.Errorf("认证失败，请核对信息")
		}
	}

	return &response.BindingLinkResponse{
		BindingLink: bindingLink,
	}, nil
}

// CheckBindingStatus 检查绑定状态
func (s *takeoutAppService) CheckBindingStatus(ctx context.Context, platform string) (*response.BindingStatusResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}
	if !takeout.Enabled {
		return nil, errors.New("平台未开启")
	}

	// 调用 bmp RPC 接口检查绑定状态
	isBound, err := s.rpcService.CheckBindingStatus(ctx.GetContext(), platform, companyUuid)
	if err != nil {
		return nil, fmt.Errorf("检查绑定状态失败: %w", err)
	}

	if isBound {
		s.takeoutDomainService.UpdatePlatformBoundStatus(ctx, takeout.Uuid, true)
	} else {
		s.takeoutDomainService.UpdatePlatformBoundStatus(ctx, takeout.Uuid, false)
	}

	return &response.BindingStatusResponse{
		IsBound: isBound,
	}, nil
}

// UpdateBindingStatus 更新绑定状态（包括 skip 字段）
func (s *takeoutAppService) UpdateBindingStatus(ctx context.Context, req request.UpdateBindingStatusRequest) error {
	// 获取平台状态
	_, err := s.takeoutDomainService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return fmt.Errorf("获取平台状态失败: %w", err)
	}
	return nil
}

// BindPlatform 绑定平台
func (s *takeoutAppService) BindPlatform(ctx context.Context, uuid uint64) error {
	err := s.takeoutDomainService.UpdatePlatformBoundStatus(ctx, uuid, true)
	if err != nil {
		return fmt.Errorf("绑定平台失败: %w", err)
	}

	return nil
}

// UnbindPlatform 解绑平台
func (s *takeoutAppService) UnbindPlatform(ctx context.Context, uuid uint64) error {
	err := s.takeoutDomainService.UpdatePlatformBoundStatus(ctx, uuid, false)
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

	// 判断 ttpos_takeout表 的 menu 是否存在数据，如果存在就判断是否导入成功，不然就返回空
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
	}
	if takeout.Menu != nil && !reflect.ValueOf(takeout.Menu).IsNil() && takeout.ImportStatus != model.ImportStatusSuccess {
		return nil, fmt.Errorf("正在导入数据到TTPOS中，请稍后再试")
	}

	// 获取平台转换器
	converter, err := s.getConverter(req.Platform)
	if err != nil {
		return nil, err
	}

	// 从数据库加载菜单数据
	var menu *entity.TakeoutMenu
	if grabConverter, ok := converter.(*grab.GrabConverter); ok {
		// Grab 平台使用专用的加载方法
		menu, err = grabConverter.LoadMenuFromDatabase(ctx, companyUuid, req.CurrencyUnit, []uint64{})
		if err != nil {
			return nil, fmt.Errorf("加载菜单数据失败: %w", err)
		}
	} else {
		return nil, errors.New("暂不支持该平台")
	}

	// 转换为平台格式
	platformData, err := converter.ConvertFromTTPOS(ctx, menu)
	if err != nil {
		return nil, fmt.Errorf("转换菜单数据失败: %w", err)
	}

	// 保存导出的菜单数据到 ttpos_menu 字段
	if err := s.takeoutDomainService.UpdateTtposMenuByPlatform(ctx, req.Platform, platformData); err != nil {
		// 仅记录日志，不影响主流程
		logger.Logger.Error("保存TTPOS导出菜单失败", zap.Error(err), zap.String("platform", req.Platform))
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

	var menu *entity.TakeoutMenu
	if err := json.Unmarshal([]byte(menuData.(string)), &menu); err != nil {
		return nil, fmt.Errorf("解析 Grab 菜单数据失败: %w", err)
	}

	// 转换为平台格式
	grabConverter, ok := s.converters["grab"].(*grab.GrabConverter)
	if !ok {
		return nil, errors.New("grab 转换器类型错误")
	}
	grabMenu, err := grabConverter.ConvertFromTTPOS(ctx, menu)
	if err != nil {
		return nil, fmt.Errorf("转换菜单数据失败: %w", err)
	}

	return &response.GrabMenuResponse{
		Platform: "grab",
		Menu:     grabMenu,
	}, nil
}

// ConvertMenuData 转换菜单数据
func (s *takeoutAppService) ConvertMenuData(ctx context.Context, req request.ImportMenuRequest) (*entity.TakeoutMenu, error) {
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

	// 解析 Grab 菜单数据
	menuEntity, err := grabConverter.ConvertToTTPOS(ctx, req.MenuData)
	if err != nil {
		return nil, fmt.Errorf("解析 Grab 菜单失败: %w", err)
	}

	// 保存菜单数据到数据库
	err = s.UpdateTakeoutMenu(ctx, "grab", req.MenuData)
	if err != nil {
		return nil, fmt.Errorf("保存菜单数据失败: %w", err)
	}

	return menuEntity, nil
}

// SaveMenuSnapshot 保存菜单快照
func (s *takeoutAppService) PushMenu(ctx context.Context, platform string, currencyUnit string) error {
	platform = strings.ToLower(platform)
	companyUuid := ctx.GetCompanyUuid()

	menu, err := s.ExportMenu(ctx, request.ExportMenuRequest{
		Platform:     platform,
		CurrencyUnit: currencyUnit,
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
			Platform:        log.Platform,
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

// UpdateMenuItem 更新菜单项（商品）
func (s *takeoutAppService) UpdateMenuItem(ctx context.Context, req request.UpdateMenuItemRequest) error {
	// 验证参数
	if req.Platform == "" {
		return errors.New("平台名称不能为空")
	}
	if req.ItemId == "" {
		return errors.New("商品ID不能为空")
	}

	// 获取平台配置获取 MerchantID
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return fmt.Errorf("获取平台状态失败: %w", err)
	}
	if !takeout.IsBound {
		return errors.New("平台未绑定，无法更新菜单项")
	}

	// 获取 MerchantID（从 RPC 获取）
	companyUuid := ctx.GetCompanyUuid()
	_, _, merchantId, _, err := s.rpcService.CheckBindingStatusWithMerchantId(ctx.GetContext(), req.Platform, companyUuid)
	if err != nil {
		return fmt.Errorf("获取商户ID失败: %w", err)
	}
	if merchantId == "" {
		return errors.New("未获取到商户ID，请检查平台绑定状态")
	}

	// 调用 RPC 更新菜单项
	err = s.rpcService.UpdateMenuItem(
		ctx.GetContext(),
		merchantId,
		req.ItemId,
		req.Price,
		req.AvailableStatus,
		req.MaxStock,
	)
	if err != nil {
		return fmt.Errorf("更新菜单项失败: %w", err)
	}

	return nil
}

// UpdateMenuModifier 更新菜单修饰符
func (s *takeoutAppService) UpdateMenuModifier(ctx context.Context, req request.UpdateMenuModifierRequest) error {
	// 验证参数
	if req.Platform == "" {
		return errors.New("平台名称不能为空")
	}
	if req.ModifierId == "" {
		return errors.New("修饰符ID不能为空")
	}
	if req.ModifierName == "" {
		return errors.New("修饰符名称不能为空")
	}

	// 获取平台配置获取 MerchantID
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return fmt.Errorf("获取平台状态失败: %w", err)
	}
	if !takeout.IsBound {
		return errors.New("平台未绑定，无法更新菜单修饰符")
	}

	// 获取 MerchantID（从 RPC 获取）
	companyUuid := ctx.GetCompanyUuid()
	_, _, merchantId, _, err := s.rpcService.CheckBindingStatusWithMerchantId(ctx.GetContext(), req.Platform, companyUuid)
	if err != nil {
		return fmt.Errorf("获取商户ID失败: %w", err)
	}
	if merchantId == "" {
		return errors.New("未获取到商户ID，请检查平台绑定状态")
	}

	// 调用 RPC 更新菜单修饰符
	err = s.rpcService.UpdateMenuModifier(
		ctx.GetContext(),
		merchantId,
		req.ModifierId,
		req.ModifierName,
		req.Price,
		req.AvailableStatus,
	)
	if err != nil {
		return fmt.Errorf("更新菜单修饰符失败: %w", err)
	}

	logger.Logger.Info("更新菜单修饰符成功",
		zap.String("platform", req.Platform),
		zap.String("modifierId", req.ModifierId))

	return nil
}

// SyncMenuChanges 同步菜单变更（灰度更新）
func (s *takeoutAppService) SyncMenuChanges(ctx context.Context, req request.ExportMenuRequest) (*response.MenuSyncResult, error) {
	// 初始化结果
	result := &response.MenuSyncResult{
		ItemChanges:     []response.MenuItemChange{},
		ModifierChanges: []response.MenuModifierChange{},
		Errors:          []string{},
	}

	// 1. 获取旧菜单（从 ttpos_menu 字段）
	takeout, err := s.takeoutDomainService.GetByPlatform(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台状态失败: %w", err)
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
	err = s.compareAndSyncMenu(ctx, req.Platform, oldMenu, newMenu, result)
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
		zap.Int("failedModifiers", result.FailedModifiers))

	return result, nil
}

// compareAndSyncMenu 比较并同步菜单变更
func (s *takeoutAppService) compareAndSyncMenu(
	ctx context.Context,
	platform string,
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

	// 遍历新菜单，比较并更新变更
	for _, category := range newMenu.Categories {
		for i := range category.Items {
			newItem := &category.Items[i]
			result.TotalItems++

			// 比较商品
			if oldItem, exists := oldItemsMap[newItem.Id]; exists {
				s.compareAndSyncItem(ctx, platform, oldItem, newItem, result)
			}

			// 比较修饰符
			for _, modGroup := range newItem.ModifierGroups {
				for j := range modGroup.Modifiers {
					newModifier := &modGroup.Modifiers[j]
					result.TotalModifiers++

					if oldModifier, exists := oldModifiersMap[newModifier.Id]; exists {
						s.compareAndSyncModifier(ctx, platform, oldModifier, newModifier, result)
					}
				}
			}
		}
	}

	return nil
}

// compareAndSyncItem 比较并同步单个商品
func (s *takeoutAppService) compareAndSyncItem(
	ctx context.Context,
	platform string,
	oldItem *grabfood.MenuItem,
	newItem *grabfood.MenuItem,
	result *response.MenuSyncResult,
) {
	var changes []string
	var updatePrice *int64
	var updateStatus string
	var updateStock int64 = 9999999

	// 检查价格变更
	if oldItem.Price != newItem.Price {
		changes = append(changes, "price")
		updatePrice = &newItem.Price
	}

	// 检查状态变更
	if oldItem.AvailableStatus != newItem.AvailableStatus {
		changes = append(changes, "status")
		updateStatus = newItem.AvailableStatus
	}

	// 检查库存变更
	if updateStatus != string(valueobject.AvailableStatusAvailable) {
		updateStock = 0
	}

	// 如果没有变更，直接返回
	if len(changes) == 0 {
		return
	}

	result.ChangedItems++
	changeType := strings.Join(changes, ",")

	// 调用更新API
	updateReq := request.UpdateMenuItemRequest{
		Platform:        platform,
		ItemId:          newItem.Id,
		Price:           updatePrice,
		AvailableStatus: updateStatus,
		MaxStock:        &updateStock,
	}

	err := s.UpdateMenuItem(ctx, updateReq)

	// 记录变更详情
	change := response.MenuItemChange{
		ItemID:     newItem.Id,
		ItemName:   newItem.Name,
		ChangeType: changeType,
		Success:    err == nil,
	}

	if updatePrice != nil {
		change.OldPrice = &oldItem.Price
		change.NewPrice = updatePrice
	}

	if updateStatus != "" {
		change.OldStatus = oldItem.AvailableStatus
		change.NewStatus = updateStatus
	}

	if err != nil {
		change.ErrorMessage = err.Error()
		result.FailedItems++
		result.Errors = append(result.Errors, fmt.Sprintf("商品%s(%s)更新失败: %v", newItem.Name, newItem.Id, err))
	} else {
		result.SuccessItems++
	}

	result.ItemChanges = append(result.ItemChanges, change)
}

// compareAndSyncModifier 比较并同步单个修饰符
func (s *takeoutAppService) compareAndSyncModifier(
	ctx context.Context,
	platform string,
	oldModifier *grabfood.MenuModifier,
	newModifier *grabfood.MenuModifier,
	result *response.MenuSyncResult,
) {
	var changes []string
	var updatePrice *int64
	var updateStatus string

	// 检查价格变更
	if oldModifier.Price != newModifier.Price {
		changes = append(changes, "price")
		updatePrice = newModifier.Price
	}

	// 检查状态变更
	if oldModifier.AvailableStatus != newModifier.AvailableStatus {
		changes = append(changes, "status")
		updateStatus = newModifier.AvailableStatus
	}

	// 如果没有变更，直接返回
	if len(changes) == 0 {
		return
	}

	result.ChangedModifiers++
	changeType := strings.Join(changes, ",")

	// 调用更新API
	updateReq := request.UpdateMenuModifierRequest{
		Platform:        platform,
		ModifierId:      newModifier.Id,
		ModifierName:    newModifier.Name,
		Price:           updatePrice,
		AvailableStatus: updateStatus,
	}

	err := s.UpdateMenuModifier(ctx, updateReq)

	// 记录变更详情
	change := response.MenuModifierChange{
		ModifierID:   newModifier.Id,
		ModifierName: newModifier.Name,
		ChangeType:   changeType,
		Success:      err == nil,
	}

	if updatePrice != nil {
		change.OldPrice = oldModifier.Price
		change.NewPrice = updatePrice
	}

	if updateStatus != "" {
		change.OldStatus = oldModifier.AvailableStatus
		change.NewStatus = updateStatus
	}

	if err != nil {
		change.ErrorMessage = err.Error()
		result.FailedModifiers++
		result.Errors = append(result.Errors, fmt.Sprintf("修饰符%s(%s)更新失败: %v", newModifier.Name, newModifier.Id, err))
	} else {
		result.SuccessModifiers++
	}

	result.ModifierChanges = append(result.ModifierChanges, change)
}
