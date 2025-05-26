package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger 打印所有请求参数的中间件
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求基本信息
		method := c.Request.Method
		uri := c.Request.RequestURI
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		fmt.Printf("\n\n\n")
		fmt.Printf("方法: %s\n", method)
		fmt.Printf("URI: %s\n", uri)
		fmt.Printf("客户端IP: %s\n", clientIP)
		fmt.Printf("User-Agent: %s\n", userAgent)
		// 打印请求头
		fmt.Printf("=== 请求头 ===\n")
		for name, values := range c.Request.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		// 打印URL查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			fmt.Printf("=== URL查询参数 ===\n")
			queryParams, _ := url.ParseQuery(c.Request.URL.RawQuery)
			for key, values := range queryParams {
				for _, value := range values {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
		}
		// 打印路径参数
		if len(c.Params) > 0 {
			fmt.Printf("=== 路径参数 ===\n")
			for _, param := range c.Params {
				fmt.Printf("%s: %s\n", param.Key, param.Value)
			}
		}
		// 读取并打印请求体
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// 重新设置请求体，因为已经被读取了
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				contentType := c.GetHeader("Content-Type")
				fmt.Printf("=== 请求体 (Content-Type: %s) ===\n", contentType)

				switch {
				case strings.Contains(contentType, "application/json"):
					// JSON格式化输出
					var jsonData interface{}
					if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
						jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
						fmt.Printf("%s\n", string(jsonBytes))
					} else {
						fmt.Printf("原始JSON: %s\n", string(bodyBytes))
					}

				case strings.Contains(contentType, "application/x-www-form-urlencoded"):
					// 表单数据
					formData, err := url.ParseQuery(string(bodyBytes))
					if err == nil {
						for key, values := range formData {
							for _, value := range values {
								fmt.Printf("%s: %s\n", key, value)
							}
						}
					} else {
						fmt.Printf("原始表单数据: %s\n", string(bodyBytes))
					}

				case strings.Contains(contentType, "multipart/form-data"):
					// 多部分表单数据
					fmt.Printf("多部分表单数据 (文件上传等)\n")
					// 这里可以进一步解析multipart数据，但比较复杂
					fmt.Printf("数据长度: %d bytes\n", len(bodyBytes))

				default:
					// 其他格式，直接输出原始数据
					if len(bodyBytes) < 1000 { // 只打印较小的数据
						fmt.Printf("原始数据: %s\n", string(bodyBytes))
					} else {
						fmt.Printf("数据长度: %d bytes (太大，不显示)\n", len(bodyBytes))
					}
				}
			}
		}

		// 如果是表单请求，也打印表单字段
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil && len(c.Request.Form) > 0 {
				fmt.Printf("\n--- 解析的表单字段 ---\n")
				for key, values := range c.Request.Form {
					for _, value := range values {
						fmt.Printf("%s: %s\n", key, value)
					}
				}
			}

			// 解析多部分表单
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil && c.Request.MultipartForm != nil {
				if len(c.Request.MultipartForm.Value) > 0 {
					fmt.Printf("\n--- 多部分表单字段 ---\n")
					for key, values := range c.Request.MultipartForm.Value {
						for _, value := range values {
							fmt.Printf("%s: %s\n", key, value)
						}
					}
				}

				if len(c.Request.MultipartForm.File) > 0 {
					fmt.Printf("\n--- 上传文件 ---\n")
					for key, files := range c.Request.MultipartForm.File {
						for _, file := range files {
							fmt.Printf("%s: %s (大小: %d bytes)\n", key, file.Filename, file.Size)
						}
					}
				}
			}
		}

		// fmt.Printf("=== 结束 ===\n")

		c.Next()
	}
}

// SimpleRequestLogger 简化版本的请求日志中间件
func SimpleRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[%s] %s - %s\n", c.Request.Method, c.Request.RequestURI, c.ClientIP())

		// 打印查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			fmt.Printf("查询参数: %s\n", c.Request.URL.RawQuery)
		}

		// 打印请求体（仅适用于小数据）
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 && len(bodyBytes) < 500 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				fmt.Printf("请求体: %s\n", string(bodyBytes))
			}
		}

		c.Next()
	}
}
