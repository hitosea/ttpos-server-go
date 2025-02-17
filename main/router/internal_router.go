package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ttpos-server-go/app/api/internal_v1"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

func SetupInternal(r *gin.Engine, dbm *database.DBManager, cache cache.Cache) {
	// 判活接口
	r.GET("api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
	r.GET("test", func(c *gin.Context) {
		A(10, 0)
	})
	apiV1 := r.Group("/internal/api/v1")
	{
		// 通用接口
		passportGroup := apiV1.Group("/admin")
		{
			internal_v1.RegisterHandlers(passportGroup, cache)
		}
	}
}
