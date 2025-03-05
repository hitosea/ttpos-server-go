package router

import (
	"net/http"
	"ttpos-server-go/app/api/v1/assistant"
	"ttpos-server-go/app/api/v1/cashier"
	"ttpos-server-go/app/api/v1/h5"
	"ttpos-server-go/app/api/v1/kitchen"
	"ttpos-server-go/app/api/v1/passport"
	"ttpos-server-go/app/api/v1/shop"
	"ttpos-server-go/app/api/v1/tablet"
	_ "ttpos-server-go/app/event" // 注册事件
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, dbm *database.DBManager, cache cache.Cache) {
	// 判活接口
	r.GET("api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
	apiV1 := r.Group("api/v1")
	{
		// 通用接口
		passportGroup := apiV1.Group("/passport")
		{
			passport.RegisterHandlers(passportGroup, cache)
		}
		// 商家端
		shopGroup := apiV1.Group("/shop")
		{
			shop.RegisterOrderHandlers(shopGroup, dbm, cache)
		}
		// 收银端
		cashierGroup := apiV1.Group("/cashier")
		{
			cashier.RegisterAuthHandlers(cashierGroup, dbm, cache)
			cashier.RegisterProductHandlers(cashierGroup, dbm, cache)
			cashier.RegisterDeskHandlers(cashierGroup, dbm, cache)
			cashier.RegisterBuffetHandlers(cashierGroup, dbm, cache)
			cashier.RegisterInstantHandlers(cashierGroup, dbm, cache)
			cashier.RegisterOrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterMemberHandlers(cashierGroup, dbm, cache)
			cashier.RegisterSoldOutHandlers(cashierGroup, dbm, cache)
			cashier.RegisterBaseHandlers(cashierGroup, dbm, cache)
			cashier.RegisterCallHandlers(cashierGroup, dbm, cache)
			cashier.RegisterH5OrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterRechargeOrderHandlers(cashierGroup, dbm, cache)
		}
		// 点餐助手端
		assistantGroup := apiV1.Group("/assistant")
		{
			assistant.RegisterDeskHandlers(assistantGroup, dbm, cache)
			assistant.RegisterBuffetHandlers(assistantGroup, dbm, cache)
			assistant.RegisterAuthHandlers(assistantGroup, dbm, cache)
			assistant.RegisterBaseHandlers(assistantGroup, dbm, cache)
			assistant.RegisterMemberHandlers(assistantGroup, dbm, cache)
		}
		// H5扫码端
		h5Group := apiV1.Group("/h5")
		{
			h5.RegisterH5Handlers(h5Group, dbm, cache)
		}
		// 厨房端
		kitchenGroup := apiV1.Group("/kitchen")
		{
			kitchen.RegisterHandlers(kitchenGroup)
			kitchen.RegisterAuthHandlers(kitchenGroup, dbm, cache)
		}
		// 厨房端
		tabletGroup := apiV1.Group("/tablet")
		{
			tablet.RegisterAuthHandlers(tabletGroup, dbm, cache)
		}
	}
}
