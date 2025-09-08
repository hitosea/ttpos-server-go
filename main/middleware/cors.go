package middleware

import (
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Ttpos-Company-ld, CLIENT-VERSION, TZ, H5-ORDER-UUID, ACCEPT, ACCEPT-ENCODING, ACCEPT-LANGUAGE, APPID, AUTHORIZATION, AUTHORIZATION-CODE, BATCHNO, BUILD, CACHE-CONTROL, CONTENT-LENGTH, CONTENT-TYPE, DEVICE-ID, ENCRYPT, ORIGIN, PLATFORM, REFERER, SHOW-FAIL-TOAST, SHOW-LOADING, SHOW-SUCCESS-TOAST, SID, SIGN, SUUID, TOKEN, USER-AGENT, VERSION-NAME, X-CSRF-TOKEN, X-REQUESTED-WITH, X-SIGN, X-DESK-UUID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
