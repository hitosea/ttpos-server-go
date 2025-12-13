package application

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/menu/entity"
	"ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/types/request"
	"ttpos-server-go/app/modules/takeout/types/response"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ITakeoutMenuAppService 外卖菜单应用服务接口
type ITakeoutMenuAppService interface {
	// ExportMenu 导出菜单到指定平台格式
	ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error)

	// GetBindingLink 获取绑定链接
	GetBindingLink(ctx context.Context) (*response.BindingLinkResponse, error)

	// CheckBindingStatus 验证是否已经绑定（定时查询）
	CheckBindingStatus(ctx context.Context) (*response.BindingStatusResponse, error)

	// GetGrabMenu 获取 Grab 的商品菜单
	GetGrabMenu(ctx context.Context) (*response.GrabMenuResponse, error)

	// GetImportMenu 导入 Grab 菜单（全新创建商品，不做绑定关系）
	GetImportMenu(ctx context.Context, req request.ImportMenuRequest) (*entity.TakeoutMenu, error)
}

// GetImportMenu 导入 Grab 菜单（全新创建商品/分类，不做绑定关系）
func (s *takeoutMenuAppService) GetImportMenu(ctx context.Context, req request.ImportMenuRequest) (*entity.TakeoutMenu, error) {
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
		return nil, errors.WithMessage(err, "解析 Grab 菜单失败")
	}

	return menuEntity, nil
}

// takeoutMenuAppService 外卖菜单应用服务实现
type takeoutMenuAppService struct {
	dbm            *database.DBManager
	menuRepo       repository.IMenuDataRepository
	converters     map[string]service.IPlatformConverter // 平台转换器映射
	cache          cache.Cache
	productMapRepo repository.IProductMapRepository
}

// NewTakeoutMenuAppService 创建外卖菜单应用服务
func NewTakeoutMenuAppService(
	dbm *database.DBManager,
	cache cache.Cache,
) ITakeoutMenuAppService {
	// 初始化平台转换器
	converters := make(map[string]service.IPlatformConverter)
	converters["Grab"] = grab.NewGrabConverter(dbm, cache)
	// 后续可添加其他平台：converters["lineman"] = lineman.NewLinemanConverter(dbm)

	return &takeoutMenuAppService{
		dbm:            dbm,
		menuRepo:       persistence.NewMenuDataRepository(dbm),
		converters:     converters,
		cache:          cache,
		productMapRepo: persistence.NewProductMapRepository(),
	}
}

// ExportMenu 导出菜单到指定平台格式
func (s *takeoutMenuAppService) ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error) {
	// 验证参数
	if req.Platform == "" {
		return nil, errors.New("平台名称不能为空")
	}
	if req.CompanyUuid == 0 {
		return nil, errors.New("公司 UUID 不能为空")
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
		menu, err = grabConverter.LoadMenuFromDatabase(ctx, req.CompanyUuid, req.CurrencyUnit, req.CategoryIDs, req.SellingTimeIDs)
		if err != nil {
			return nil, errors.WithMessage(err, "加载菜单数据失败")
		}
	} else {
		return nil, errors.New("暂不支持该平台")
	}

	// 转换为平台格式
	platformData, err := converter.ConvertFromTTPOS(ctx, menu)
	if err != nil {
		return nil, errors.WithMessage(err, "转换菜单数据失败")
	}

	return platformData, nil
}

// GetBindingLink 获取绑定链接
func (s *takeoutMenuAppService) GetBindingLink(ctx context.Context) (*response.BindingLinkResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口获取绑定链接
	// 等待 bmp 实现 GetGrabBindingLink RPC 接口
	bindingLink, expiresAt, err := takeout.GetGrabBindingLink(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取绑定链接失败")
	}

	return &response.BindingLinkResponse{
		BindingLink: bindingLink,
		ExpiresAt:   expiresAt,
	}, nil
}

// CheckBindingStatus 验证是否已经绑定（定时查询）
func (s *takeoutMenuAppService) CheckBindingStatus(ctx context.Context) (*response.BindingStatusResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口检查绑定状态
	// 等待 bmp 实现 CheckGrabBindingStatus RPC 接口
	isBound, boundAt, merchantID, merchantName, err := takeout.CheckGrabBindingStatus(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "检查绑定状态失败")
	}

	return &response.BindingStatusResponse{
		IsBound:      isBound,
		BoundAt:      boundAt,
		MerchantID:   merchantID,
		MerchantName: merchantName,
	}, nil
}

// GetGrabMenu 获取 Grab 的商品菜单
func (s *takeoutMenuAppService) GetGrabMenu(ctx context.Context) (*response.GrabMenuResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口获取 Grab 菜单
	// 等待 bmp 实现 GetGrabMenu RPC 接口
	menuData, err := takeout.GetGrabMenu(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取 Grab 菜单失败")
	}

	return &response.GrabMenuResponse{
		Menu: menuData,
	}, nil
}

// getConverter 获取平台转换器
func (s *takeoutMenuAppService) getConverter(platform string) (service.IPlatformConverter, error) {
	converter, ok := s.converters[platform]
	if !ok {
		return nil, errors.New("不支持的平台: " + platform)
	}
	return converter, nil
}
