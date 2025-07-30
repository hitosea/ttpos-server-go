package middleware

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gbuild"
)

// MiddlewareHandlerVersion 在头信息中添加版本信息
func MiddlewareHandlerVersion(r *ghttp.Request) {
	r.Middleware.Next()
	r.Response.Header().Set("X-Ttpos-Version", gjson.MustEncodeString(gbuild.Info()))
}
