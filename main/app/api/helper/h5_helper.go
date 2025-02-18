package helper

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/i18n"
)

func H5Fail(c *gin.Context, code int, message string) {
	msg := i18n.Translate(i18n.GetAcceptLanguage(c), message)
	c.JSON(http.StatusOK, resp.H5Response{
		Code: code,
		Msg:  msg,
		Data: struct{}{},
	})
}

func H5Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, resp.H5Response{
		Code: 1,
		Msg:  "",
		Data: data,
	})
}

func H5SuccessWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, resp.H5Response{
		Code: 1,
		Msg:  i18n.Translate(i18n.GetAcceptLanguage(c), msg),
		Data: struct{}{},
	})
}
