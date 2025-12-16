package application

import (
	"ttpos-server-go/app/dto/req"
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

	// ImportMenu 导入 Grab 菜单（分类/商品/规格/属性/加料/单位 按需创建或关联）
	ImportMenu(ctx context.Context, req ImportMenuRequest) (*GrabProductImportResult, error)
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
	CompanyUuid uint64
	MenuData    interface{}
	Overwrite   bool
}

// ImportMenu 导入 Grab 菜单（分类/商品/规格/属性/加料/单位）
// 占位实现：解析菜单，逐项调用 ImportGrabProducts 逻辑后续补全
func (s *takeoutMenuAppService) ImportMenu(ctx context.Context, req ImportMenuRequest) (*GrabProductImportResult, error) {
	companyUuid := ctx.GetCompanyUuid()
	if req.CompanyUuid != 0 {
		companyUuid = req.CompanyUuid
	}
	if companyUuid == 0 {
		return nil, errors.New("公司 UUID 不能为空")
	}
	if req.MenuData == nil {
		return nil, errors.New("菜单数据不能为空")
	}

	// TODO: 解析 Grab 菜单（grab_models.go 结构），执行：
	// 1) 分类查找/创建
	// 2) 商品查找/创建（默认下架）、建立映射
	// 3) 规格/变体、属性组/属性值、单位查找/创建
	// 4) 汇总结果

	return &GrabProductImportResult{
		SuccessCount: 0,
		FailureCount: 0,
		CreatedItems: 0,
		UpdatedItems: 0,
		Failures:     make([]resp.GrabProductImportFailure, 0),
	}, nil
}

// takeoutMenuAppService 外卖菜单应用服务实现
type takeoutMenuAppService struct {
	dbm         *database.DBManager
	menuRepo    repository.IMenuDataRepository
	converters  map[string]service.IPlatformConverter // 平台转换器映射
	cache       cache.Cache
	grabMapRepo repository.IGrabProductMapRepository
}

// NewTakeoutMenuAppService 创建外卖菜单应用服务
func NewTakeoutMenuAppService(
	dbm *database.DBManager,
	menuRepo repository.IMenuDataRepository,
	cache cache.Cache,
) ITakeoutMenuAppService {
	// 初始化平台转换器
	converters := make(map[string]service.IPlatformConverter)
	converters["grab"] = grab.NewGrabConverter(dbm, cache)
	// 后续可添加其他平台：converters["lineman"] = lineman.NewLinemanConverter(dbm)

	return &takeoutMenuAppService{
		dbm:         dbm,
		menuRepo:    menuRepo,
		converters:  converters,
		cache:       cache,
		grabMapRepo: persistence.NewGrabProductMapRepository(),
	}
}

// ExportMenuRequest 导出菜单请求
type ExportMenuRequest struct {
	Platform       string   // 平台名称：grab, lineman 等
	CompanyUuid    uint64   // 公司 UUID
	CategoryIDs    []uint64 // 分类 ID 列表（可选）
	SellingTimeIDs []uint64 // 售卖时段 ID 列表（可选）
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
		menu, err = grabConverter.LoadMenuFromDatabase(ctx, req.CompanyUuid, req.CategoryIDs, req.SellingTimeIDs)
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

// GrabProductImportRequest Grab 商品导入请求
type GrabProductImportRequest struct {
	CompanyUuid uint64
	Items       []req.GrabProductImportItem
	Overwrite   bool
}

// GrabProductImportResult Grab 商品导入结果
type GrabProductImportResult struct {
	SuccessCount int
	FailureCount int
	CreatedItems int
	UpdatedItems int
	Failures     []resp.GrabProductImportFailure
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
