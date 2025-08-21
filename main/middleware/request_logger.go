package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CustomResponseWriter 自定义ResponseWriter来捕获响应数据
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法来捕获响应体
func (w *CustomResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// RequestLogger 打印所有请求参数和响应结果的中间件
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 获取请求基本信息
		method := c.Request.Method
		uri := c.Request.RequestURI
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		fmt.Printf("\n\n\n")
		fmt.Println("================================= 请求开始 =================================")
		fmt.Println("时间: ", startTime.Format("2006-01-02 15:04:05"))
		fmt.Println("方法: ", method)
		fmt.Println("URI: ", uri)
		fmt.Println("客户端IP: ", clientIP)
		fmt.Println("User-Agent: ", userAgent)
		// 打印请求头
		fmt.Println("------ 请求头 ------")
		for name, values := range c.Request.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		// 打印URL查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			fmt.Println("------ URL查询参数 ------")
			queryParams, _ := url.ParseQuery(c.Request.URL.RawQuery)
			for key, values := range queryParams {
				for _, value := range values {
					fmt.Printf("%s: %s\n", key, value)
				}
			}
		}
		// 打印路径参数
		if len(c.Params) > 0 {
			fmt.Printf("------ 路径参数 ------\n")
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
				fmt.Println("------ 请求体 (Content-Type: ", contentType, ") ------")

				switch {
				case strings.Contains(contentType, "application/json"):
					// JSON格式化输出
					var jsonData interface{}
					if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
						jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
						fmt.Printf("%s\n", string(jsonBytes))
					} else {
						fmt.Println("原始JSON: ", string(bodyBytes))
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
						fmt.Println("原始表单数据: ", string(bodyBytes))
					}

				case strings.Contains(contentType, "multipart/form-data"):
					// 多部分表单数据
					fmt.Printf("多部分表单数据 (文件上传等)\n")
					// 这里可以进一步解析multipart数据，但比较复杂
					fmt.Println("数据长度: ", len(bodyBytes), "bytes")

				default:
					// 其他格式，直接输出原始数据
					if len(bodyBytes) < 1000 { // 只打印较小的数据
						fmt.Println("原始数据: ", string(bodyBytes))
					} else {
						fmt.Println("数据长度: ", len(bodyBytes), "bytes (太大，不显示)")
					}
				}
			}
		}

		// 如果是表单请求，也打印表单字段
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if len(c.Request.Form) > 0 {
				fmt.Println("\n--- 解析的表单字段 ---")
				for key, values := range c.Request.Form {
					for _, value := range values {
						fmt.Printf("%s: %s\n", key, value)
					}
				}
			}

			// 解析多部分表单
			if c.Request.MultipartForm != nil {
				if len(c.Request.MultipartForm.Value) > 0 {
					fmt.Println("\n--- 多部分表单字段 ---")
					for key, values := range c.Request.MultipartForm.Value {
						for _, value := range values {
							fmt.Printf("%s: %s\n", key, value)
						}
					}
				}

				if len(c.Request.MultipartForm.File) > 0 {
					fmt.Println("\n--- 上传文件 ---")
					for key, files := range c.Request.MultipartForm.File {
						for _, file := range files {
							fmt.Println(key, file.Filename, file.Size)
						}
					}
				}
			}
		}

		// 创建自定义的ResponseWriter来捕获响应
		responseBody := &bytes.Buffer{}
		customWriter := &CustomResponseWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = customWriter

		// 处理请求
		c.Next()

		// 请求处理完成后打印响应信息
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		fmt.Println("\n------ 响应结果 ------")
		fmt.Println("响应时间: ", endTime.Format("2006-01-02 15:04:05"))
		fmt.Println("处理耗时: ", duration)
		fmt.Println("状态码: ", c.Writer.Status())

		// 打印响应头
		fmt.Println("------ 响应头 ------")
		for name, values := range c.Writer.Header() {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}

		// 打印响应体
		if responseBody.Len() > 0 {
			fmt.Println("------ 响应体 ------")
			responseData := responseBody.Bytes()

			// 检查Content-Type来决定如何格式化输出
			contentType := c.Writer.Header().Get("Content-Type")

			switch {
			case strings.Contains(contentType, "application/json"):
				// JSON格式化输出
				var jsonData interface{}
				if err := json.Unmarshal(responseData, &jsonData); err == nil {
					jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
					fmt.Printf("%s\n", string(jsonBytes))
				} else {
					fmt.Println("原始JSON响应: ", string(responseData))
				}

			case strings.Contains(contentType, "text/html"), strings.Contains(contentType, "text/plain"):
				// HTML或纯文本
				if len(responseData) < 2000 { // 只打印较小的HTML
					fmt.Println("HTML/文本响应: ", string(responseData))
				} else {
					fmt.Println("HTML/文本响应长度: ", len(responseData), "bytes (太大，不显示)")
				}

			case strings.Contains(contentType, "application/xml"):
				// XML格式
				if len(responseData) < 2000 {
					fmt.Println("XML响应: ", string(responseData))
				} else {
					fmt.Println("XML响应长度: ", len(responseData), "bytes (太大，不显示)")
				}

			default:
				// 其他格式
				if len(responseData) < 1000 {
					fmt.Println("响应数据: ", string(responseData))
				} else {
					fmt.Println("响应数据长度: ", len(responseData), "bytes (太大，不显示)")
				}
			}
		} else {
			fmt.Println("------ 响应体为空 ------")
		}

		fmt.Println("================================= 请求结束 =================================")
	}
}
