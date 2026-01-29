package event

import (
	"sync"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	takeoutService "ttpos-server-go/app/service/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_product_sold_out_event_handler sync.Once

// ProductSoldOutEventHandler "商品沽清"事件处理器
func ProductSoldOutEventHandler() {
	once_product_sold_out_event_handler.Do(func() {
		event.NewSystemBus().SubscribeProductSoldOutEvent(func(payload event.ProductSoldOutPayload) {
			// 推送菜单到外卖平台
			dbm := database.GetDBManager(config.Database)
			cache := cache.NewCache(cache.Redis, cache.Config{
				Host:     config.Redis.Host,
				Port:     config.Redis.Port,
				Password: config.Redis.Password,
				DB:       config.Redis.DB,
			})
			settingSrv := setting.NewSrv(dbm, cache)
			translateSrv := service.NewTranslateSrv(dbm, cache)
			takeoutSrv := takeoutService.NewTakeoutSrv(dbm, cache, nil, nil, translateSrv, settingSrv)

			utils.Go(func() {
				// 推送菜单到Grab平台
				_, err := takeoutSrv.SyncMenuChanges(payload.Ctx, value_object.TakeoutPlatformGrab)
				if err != nil {
					logger.Logger.Error("推送菜单到Grab平台失败", zap.Error(err))
				}
				// 推送菜单到Lineman平台
				_, err = takeoutSrv.SyncMenuChanges(payload.Ctx, value_object.TakeoutPlatformLineman)
				if err != nil {
					logger.Logger.Error("推送菜单到Lineman平台失败", zap.Error(err))
				}
			})
		})
	})
}
