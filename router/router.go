package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jjjshop-server-go/app/api/v1"
	"jjjshop-server-go/app/api/v1/cashier"
	"jjjshop-server-go/app/repository"
	"jjjshop-server-go/app/service"
	"jjjshop-server-go/middleware"
	"jjjshop-server-go/pkg/cache"
	"jjjshop-server-go/pkg/database"
)

func Setup(r *gin.Engine, dbm *database.DBManager, cache cache.Cache) {
	// 初始化仓库
	userRepo := repository.NewUserRepository(dbm)
	bindRecordRepo := repository.NewBindRecordRepository(dbm)
	userRoleRepo := repository.NewUserRoleRepository(dbm)
	accessRepo := repository.NewAccessRepository(dbm)
	supplierRepo := repository.NewSupplierRepository(dbm)

	// 初始化验证码服务
	captchaService := service.NewCaptchaService(cache)
	pgpService := service.NewPGPService(cache)
	roleAccessService := service.NewRoleAccessService(userRoleRepo, accessRepo, userRepo)
	bindRecordService := service.NewBindRecordService(bindRecordRepo, supplierRepo)
	cashierAuthService := service.NewCashierAuthService(userRepo, captchaService, roleAccessService, bindRecordService)

	// 初始化处理器
	passportHandler := v1.NewPassportHandler(captchaService, pgpService)
	cashierAuthHandler := cashier.NewCashierAuthHandler(cashierAuthService)

	r.GET("api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	// 公开路由
	publicApiV1 := r.Group("api/v1")
	{
		publicApiV1.GET("/passport/captcha", passportHandler.GetCaptcha)                   // 获取验证码
		publicApiV1.GET("/passport/server-public-key", passportHandler.GetServerPublicKey) // 获取服务端 PGP 公钥
		// 收银端
		cashierGroup := publicApiV1.Group("/cashier")
		{
			cashierGroup.POST("/passport/login", middleware.Encrypt(cache), cashierAuthHandler.Login) // 登录
		}
		// 点餐助手
		assistantGroup := publicApiV1.Group("/assistant")
		{
			assistantGroup.GET("/passport/captcha", passportHandler.GetCaptcha)
		}
	}

	// 需要认证的路由
	privateApiV1 := r.Group("api/v1")
	privateApiV1.Use(middleware.Auth())
	{
		// 收银端
		cashierGroup := privateApiV1.Group("/cashier")
		{
			cashierGroup.POST("/passport/logout", cashierAuthHandler.Logout) // 收银端退出登录
			//cashierGroup.GET("/info", nil)                                   // 获取登录信息
		}
	}
}
