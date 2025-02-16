package h5

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/i18n"
)

func Fail(c *gin.Context, code int, message ...string) {
	msg := "fail"
	if len(message) == 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0])
	} else if len(message) > 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0], message[1:]...)
	}
	c.JSON(http.StatusOK, resp.H5Response{
		Code: code,
		Msg:  msg,
		Data: struct{}{},
	})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, resp.H5Response{
		Code: 1,
		Msg:  "",
		Data: data,
	})
}
