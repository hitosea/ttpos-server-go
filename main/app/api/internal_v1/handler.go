package internal_v1

import (
	"github.com/gin-gonic/gin"
	"ttpos-server-go/pkg/cache"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/service"
)

type Handler struct {
	captchaService service.ICaptchaSrv
	encryptService service.IEncryptSrv
}

func (h *Handler) Ping(c *gin.Context) {
	helper.Success(c, struct {
		Name string
	}{Name: "pong"})
}

func RegisterHandlers(router gin.IRouter, cache cache.Cache) {
	wrapper := &Handler{
		captchaService: service.NewCaptchaSrv(cache),
		encryptService: service.NewEncryptSrv(cache),
	}
	router.GET("/ping", wrapper.Ping)
}
