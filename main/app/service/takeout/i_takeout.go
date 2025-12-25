package takeout

import (
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type ITakeoutSrv interface {
	// ToggleTakeoutStatus 切换指定平台外卖状态
	ToggleTakeoutStatus(ctx context.Context, req request.ToggleTakeoutStatusRequest) (*response.TakeoutStatusResponse, error)

	// 导入菜单到TTPOS
	ImportMenuToTTPOS(ctx context.Context) (*resp.GrabMenuImportResp, error)

	// PushMenuToPlatform 推送菜单到外卖平台
	PushMenuToPlatform(ctx context.Context, platform string) error

	// SyncMenuChanges 同步菜单变更
	SyncMenuChanges(ctx context.Context, platform string) (*response.MenuSyncResult, error)

	// ReimportMenuToTTPOS 重新导入菜单到TTPOS（基于失败日志重试）
	ReimportMenuToTTPOS(ctx context.Context, logUuid uint64) (*resp.GrabMenuImportResp, error)

	// ProcessTakeoutOrderOutboundAndSales 处理外卖订单出库和销量
	ProcessTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64, acceptedBy uint64) error

	// RestoreTakeoutOrderOutboundAndSales 恢复外卖订单出库和销量（取消订单时调用）
	RestoreTakeoutOrderOutboundAndSales(ctx context.Context, orderUuid uint64, companyUuid uint64) error
}

type takeoutSrv struct {
	dbm               *database.DBManager
	cache             cache.Cache
	takeoutAppSrv     application.ITakeoutAppService
	productSrv        service.IProductSrv
	translateSrv      service.ITranslateSrv
	settingSrv        setting.ISrv
	uploadFileSrv     service.IUploadFileSrv
	productTakeoutSrv service.IProductTakeoutSrv
}

func NewTakeoutSrv(
	dbm *database.DBManager,
	cache cache.Cache,
	productSrv service.IProductSrv,
	productTakeoutSrv service.IProductTakeoutSrv,
	translateSrv service.ITranslateSrv,
	settingSrv setting.ISrv,
) ITakeoutSrv {
	return &takeoutSrv{
		dbm:               dbm,
		cache:             cache,
		takeoutAppSrv:     application.NewTakeoutAppService(dbm),
		productSrv:        productSrv,
		translateSrv:      translateSrv,
		settingSrv:        settingSrv,
		uploadFileSrv:     service.NewUploadFileSrv(dbm),
		productTakeoutSrv: productTakeoutSrv,
	}
}
