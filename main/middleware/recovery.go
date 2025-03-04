package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"ttpos-server-go/app/constant"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger, mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				internalErr := gin.H{
					"code":    500,
					"message": "Internal Server Error",
				}
				errMsg := string(debug.Stack())
				if mode == constant.ServerModeDebug {
					internalErr["error"] = err
					internalErr["stack"] = errMsg
					fmt.Println(errMsg)
				}
				logger.Error("panic", zap.Any("stack", errMsg))
				c.AbortWithStatusJSON(http.StatusInternalServerError, internalErr)
			}
		}()
		c.Next()
	}
}
