package middleware

import (
	"bytes"
	"io"
	"net/url"
	"slices"
	"ttpos-server-go/app/constant/jwt"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/service"
)

func OperationLog(operationLogSrv service.IStaffOperationLogSrv) gin.HandlerFunc {
	return func(c *gin.Context) {

		companyId := c.GetUint64(jwt.CompanyUuid)
		staffId := c.GetUint64(jwt.StaffUuid)
		if companyId == 0 || staffId == 0 {
			c.Next()
			return
		}

		urlPath := c.Request.URL.Path
		if slices.Contains(operationLogSrv.GetExcludedApis(), urlPath) {
			c.Next()
			return
		}

		accessName := operationLogSrv.GetAccessName(companyId, urlPath)
		if accessName == "" {
			c.Next()
			return
		}

		var requestData string
		// 请求方式
		requestMethod := c.Request.Method
		// 获取请求参数
		if requestMethod == "POST" { // POST请求
			if c.Request.Body != nil {
				requestBody, _ := io.ReadAll(c.Request.Body)
				// 重新填充请求体，以便后续中间件可以使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				requestData = string(requestBody)
			}
		} else if requestMethod == "GET" { // GET请求
			query := c.Request.URL.RawQuery
			if query != "" {
				requestData, _ = url.QueryUnescape(query)
			}
		}

		// 记录操作日志
		operationLogSrv.SaveLog(service.RequestLog{
			CompanyUuid:   companyId,
			StaffUuid:     staffId,
			Source:        c.GetString(jwt.Source),
			AccessName:    accessName,
			UserAgent:     c.GetHeader("user-agent"),
			UrlPath:       urlPath,
			RequestMethod: requestMethod,
			RequestData:   requestData,
			IP:            c.ClientIP(),
		})
		c.Next()
	}
}
