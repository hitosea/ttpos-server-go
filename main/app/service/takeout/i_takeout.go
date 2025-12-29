package takeout

import (
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

type ITakeoutSrv interface {
	ITakeoutMenuSrv
	ITakeoutOrderSrv
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

func NewTakeoutSrvImpl(
	dbm *database.DBManager,
) ITakeoutSrv {
	return &takeoutSrv{
		dbm:           dbm,
		takeoutAppSrv: application.NewTakeoutAppService(dbm),
		uploadFileSrv: service.NewUploadFileSrv(dbm),
	}
}
