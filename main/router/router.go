package router

import (
	"net/http"
	"ttpos-server-go/app/api/v1/assistant"
	"ttpos-server-go/app/api/v1/cashier"
	"ttpos-server-go/app/api/v1/kitchen"
	"ttpos-server-go/app/api/v1/passport"
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
		}
		// 点餐助手端
		assistantGroup := apiV1.Group("/assistant")
		{
			assistant.RegisterDeskHandlers(assistantGroup, dbm, cache)
			assistant.RegisterBuffetHandlers(assistantGroup, dbm, cache)
			assistant.RegisterAuthHandlers(assistantGroup, dbm, cache)
		}
		// 厨房端
		kitchenGroup := apiV1.Group("/kitchen")
		{
			kitchen.RegisterHandlers(kitchenGroup)
		}
	}
}
