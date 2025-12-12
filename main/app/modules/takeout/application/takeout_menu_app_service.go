package application

import (
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/menu/entity"
	"ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ITakeoutMenuAppService 外卖菜单应用服务接口
type ITakeoutMenuAppService interface {
	// ExportMenu 导出菜单到指定平台格式
	ExportMenu(ctx context.Context, req ExportMenuRequest) (interface{}, error)

	// GetBindingLink 获取绑定链接
	GetBindingLink(ctx context.Context) (*BindingLinkResponse, error)

	// CheckBindingStatus 验证是否已经绑定（定时查询）
	CheckBindingStatus(ctx context.Context) (*BindingStatusResponse, error)

	// GetGrabMenu 获取 Grab 的商品菜单
	GetGrabMenu(ctx context.Context) (*GrabMenuResponse, error)

	// GetImportMenu 导入 Grab 菜单（全新创建商品，不做绑定关系）
	GetImportMenu(ctx context.Context, req ImportMenuRequest) (*entity.TakeoutMenu, error)
}

// BindingLinkResponse 绑定链接响应
type BindingLinkResponse struct {
	BindingLink string `json:"bindingLink"` // 绑定链接 URL
	ExpiresAt   int64  `json:"expiresAt"`   // 过期时间（Unix 时间戳）
}

// BindingStatusResponse 绑定状态响应
type BindingStatusResponse struct {
	IsBound      bool   `json:"isBound"`      // 是否已绑定
	BoundAt      int64  `json:"boundAt"`      // 绑定时间（Unix 时间戳）
	MerchantID   string `json:"merchantId"`   // Grab 商户 ID
	MerchantName string `json:"merchantName"` // Grab 商户名称
}

// GrabMenuResponse Grab 菜单响应
type GrabMenuResponse struct {
	Menu interface{} `json:"menu"` // Grab 菜单数据
}

// ImportMenuRequest 导入 Grab 菜单请求
type ImportMenuRequest struct {
	Platform string      `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	MenuData interface{} `json:"menuData" binding:"required"` // 平台菜单 JSON 数据
}

// GetImportMenu 导入 Grab 菜单（全新创建商品/分类，不做绑定关系）
func (s *takeoutMenuAppService) GetImportMenu(ctx context.Context, req ImportMenuRequest) (*entity.TakeoutMenu, error) {
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
	converters["grab"] = grab.NewGrabConverter(dbm, cache)
	// 后续可添加其他平台：converters["lineman"] = lineman.NewLinemanConverter(dbm)

	return &takeoutMenuAppService{
		dbm:            dbm,
		menuRepo:       persistence.NewMenuDataRepository(dbm),
		converters:     converters,
		cache:          cache,
		productMapRepo: persistence.NewProductMapRepository(),
	}
}

// ExportMenuRequest 导出菜单请求
type ExportMenuRequest struct {
	Platform       string   // 平台名称：grab, lineman 等
	CompanyUuid    uint64   // 公司 UUID
	CurrencyUnit   string   // 货币单位
	CategoryIDs    []uint64 // 分类 ID 列表（可选）
	SellingTimeIDs []uint64 // 售卖时段 ID 列表（可选）
}

// GrabProductImportResult Grab 商品导入结果
type GrabProductImportResult struct {
	SuccessCount int
	FailureCount int
	CreatedItems int
	UpdatedItems int
	Failures     []resp.GrabProductImportFailure
}

// ExportMenu 导出菜单到指定平台格式
func (s *takeoutMenuAppService) ExportMenu(ctx context.Context, req ExportMenuRequest) (interface{}, error) {
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
func (s *takeoutMenuAppService) GetBindingLink(ctx context.Context) (*BindingLinkResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口获取绑定链接
	// 等待 bmp 实现 GetGrabBindingLink RPC 接口
	bindingLink, expiresAt, err := takeout.GetGrabBindingLink(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取绑定链接失败")
	}

	return &BindingLinkResponse{
		BindingLink: bindingLink,
		ExpiresAt:   expiresAt,
	}, nil
}

// CheckBindingStatus 验证是否已经绑定（定时查询）
func (s *takeoutMenuAppService) CheckBindingStatus(ctx context.Context) (*BindingStatusResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口检查绑定状态
	// 等待 bmp 实现 CheckGrabBindingStatus RPC 接口
	isBound, boundAt, merchantID, merchantName, err := takeout.CheckGrabBindingStatus(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "检查绑定状态失败")
	}

	return &BindingStatusResponse{
		IsBound:      isBound,
		BoundAt:      boundAt,
		MerchantID:   merchantID,
		MerchantName: merchantName,
	}, nil
}

// GetGrabMenu 获取 Grab 的商品菜单
func (s *takeoutMenuAppService) GetGrabMenu(ctx context.Context) (*GrabMenuResponse, error) {
	companyUuid := ctx.GetCompanyUuid()
	// TODO: 调用 bmp RPC 接口获取 Grab 菜单
	// 等待 bmp 实现 GetGrabMenu RPC 接口
	menuData, err := takeout.GetGrabMenu(ctx.GetContext(), companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取 Grab 菜单失败")
	}

	return &GrabMenuResponse{
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
