package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"strings"
	"sync"
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

// 全局日志互斥锁，确保请求日志不交叉
var logMutex sync.Mutex

// generateRequestID 生成请求唯一ID
func generateRequestID() string {
	return fmt.Sprintf("%d%04d", time.Now().Unix(), rand.Intn(10000))
}

// RequestLogger 打印所有请求参数和响应结果的中间件
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求唯一ID
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		startTime := time.Now()

		// 使用缓冲区收集所有日志输出
		var logBuffer bytes.Buffer

		// 获取请求基本信息
		method := c.Request.Method
		uri := c.Request.RequestURI
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		logBuffer.WriteString("\n\n\n")
		logBuffer.WriteString(fmt.Sprintf("[%s] ================================= 请求开始 =================================\n", requestID))
		logBuffer.WriteString(fmt.Sprintf("[%s] 时间: %s\n", requestID, startTime.Format("2006-01-02 15:04:05")))
		logBuffer.WriteString(fmt.Sprintf("[%s] 方法: %s\n", requestID, method))
		logBuffer.WriteString(fmt.Sprintf("[%s] URI: %s\n", requestID, uri))
		logBuffer.WriteString(fmt.Sprintf("[%s] 客户端IP: %s\n", requestID, clientIP))
		logBuffer.WriteString(fmt.Sprintf("[%s] User-Agent: %s\n", requestID, userAgent))
		// 打印请求头
		logBuffer.WriteString(fmt.Sprintf("[%s] ------ 请求头 ------\n", requestID))
		for name, values := range c.Request.Header {
			for _, value := range values {
				logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, name, value))
			}
		}
		// 打印URL查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			logBuffer.WriteString(fmt.Sprintf("[%s] ------ URL查询参数 ------\n", requestID))
			queryParams, _ := url.ParseQuery(c.Request.URL.RawQuery)
			for key, values := range queryParams {
				for _, value := range values {
					logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, key, value))
				}
			}
		}
		// 打印路径参数
		if len(c.Params) > 0 {
			logBuffer.WriteString(fmt.Sprintf("[%s] ------ 路径参数 ------\n", requestID))
			for _, param := range c.Params {
				logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, param.Key, param.Value))
			}
		}
		// 读取并打印请求体
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// 重新设置请求体，因为已经被读取了
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				contentType := c.GetHeader("Content-Type")
				logBuffer.WriteString(fmt.Sprintf("[%s] ------ 请求体 (Content-Type: %s) ------\n", requestID, contentType))

				switch {
				case strings.Contains(contentType, "application/json"):
					// JSON格式化输出
					var jsonData interface{}
					if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
						jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
						// 为每行JSON添加请求ID前缀
						lines := strings.Split(string(jsonBytes), "\n")
						for _, line := range lines {
							if strings.TrimSpace(line) != "" {
								logBuffer.WriteString(fmt.Sprintf("[%s] %s\n", requestID, line))
							}
						}
					} else {
						logBuffer.WriteString(fmt.Sprintf("[%s] 原始JSON: %s\n", requestID, string(bodyBytes)))
					}

				case strings.Contains(contentType, "application/x-www-form-urlencoded"):
					// 表单数据
					formData, err := url.ParseQuery(string(bodyBytes))
					if err == nil {
						for key, values := range formData {
							for _, value := range values {
								logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, key, value))
							}
						}
					} else {
						logBuffer.WriteString(fmt.Sprintf("[%s] 原始表单数据: %s\n", requestID, string(bodyBytes)))
					}

				case strings.Contains(contentType, "multipart/form-data"):
					// 多部分表单数据
					logBuffer.WriteString(fmt.Sprintf("[%s] 多部分表单数据 (文件上传等)\n", requestID))
					// 这里可以进一步解析multipart数据，但比较复杂
					logBuffer.WriteString(fmt.Sprintf("[%s] 数据长度: %d bytes\n", requestID, len(bodyBytes)))

				default:
					// 其他格式，直接输出原始数据
					if len(bodyBytes) < 1000 { // 只打印较小的数据
						logBuffer.WriteString(fmt.Sprintf("[%s] 原始数据: %s\n", requestID, string(bodyBytes)))
					} else {
						logBuffer.WriteString(fmt.Sprintf("[%s] 数据长度: %d bytes (太大，不显示)\n", requestID, len(bodyBytes)))
					}
				}
			}
		}

		// 如果是表单请求，也打印表单字段
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if len(c.Request.Form) > 0 {
				logBuffer.WriteString(fmt.Sprintf("[%s] --- 解析的表单字段 ---\n", requestID))
				for key, values := range c.Request.Form {
					for _, value := range values {
						logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, key, value))
					}
				}
			}

			// 解析多部分表单
			if c.Request.MultipartForm != nil {
				if len(c.Request.MultipartForm.Value) > 0 {
					logBuffer.WriteString(fmt.Sprintf("[%s] --- 多部分表单字段 ---\n", requestID))
					for key, values := range c.Request.MultipartForm.Value {
						for _, value := range values {
							logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, key, value))
						}
					}
				}

				if len(c.Request.MultipartForm.File) > 0 {
					logBuffer.WriteString(fmt.Sprintf("[%s] --- 上传文件 ---\n", requestID))
					for key, files := range c.Request.MultipartForm.File {
						for _, file := range files {
							logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s (%d bytes)\n", requestID, key, file.Filename, file.Size))
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

		// // 请求处理完成后收集响应信息
		// endTime := time.Now()
		// duration := endTime.Sub(startTime)

		// logBuffer.WriteString(fmt.Sprintf("\n[%s] ------ 响应结果 ------\n", requestID))
		// logBuffer.WriteString(fmt.Sprintf("[%s] 响应时间: %s\n", requestID, endTime.Format("2006-01-02 15:04:05")))
		// logBuffer.WriteString(fmt.Sprintf("[%s] 处理耗时: %v\n", requestID, duration))
		// logBuffer.WriteString(fmt.Sprintf("[%s] 状态码: %d\n", requestID, c.Writer.Status()))

		// // 收集响应头
		// logBuffer.WriteString(fmt.Sprintf("[%s] ------ 响应头 ------\n", requestID))
		// for name, values := range c.Writer.Header() {
		// 	for _, value := range values {
		// 		logBuffer.WriteString(fmt.Sprintf("[%s] %s: %s\n", requestID, name, value))
		// 	}
		// }

		// // 打印响应体
		// if responseBody.Len() > 0 {
		// 	logBuffer.WriteString(fmt.Sprintf("[%s] ------ 响应体 ------\n", requestID))
		// 	responseData := responseBody.Bytes()

		// 	// 检查Content-Type来决定如何格式化输出
		// 	contentType := c.Writer.Header().Get("Content-Type")

		// 	switch {
		// 	case strings.Contains(contentType, "application/json"):
		// 		// JSON格式化输出
		// 		var jsonData interface{}
		// 		if err := json.Unmarshal(responseData, &jsonData); err == nil {
		// 			jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
		// 			// 为每行JSON添加请求ID前缀
		// 			lines := strings.Split(string(jsonBytes), "\n")
		// 			for _, line := range lines {
		// 				if strings.TrimSpace(line) != "" {
		// 					logBuffer.WriteString(fmt.Sprintf("[%s] %s\n", requestID, line))
		// 				}
		// 			}
		// 		} else {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] 原始JSON响应: %s\n", requestID, string(responseData)))
		// 		}

		// 	case strings.Contains(contentType, "text/html"), strings.Contains(contentType, "text/plain"):
		// 		// HTML或纯文本
		// 		if len(responseData) < 2000 { // 只打印较小的HTML
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] HTML/文本响应: %s\n", requestID, string(responseData)))
		// 		} else {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] HTML/文本响应长度: %d bytes (太大，不显示)\n", requestID, len(responseData)))
		// 		}

		// 	case strings.Contains(contentType, "application/xml"):
		// 		// XML格式
		// 		if len(responseData) < 2000 {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] XML响应: %s\n", requestID, string(responseData)))
		// 		} else {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] XML响应长度: %d bytes (太大，不显示)\n", requestID, len(responseData)))
		// 		}

		// 	default:
		// 		// 其他格式
		// 		if len(responseData) < 1000 {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] 响应数据: %s\n", requestID, string(responseData)))
		// 		} else {
		// 			logBuffer.WriteString(fmt.Sprintf("[%s] 响应数据长度: %d bytes (太大，不显示)\n", requestID, len(responseData)))
		// 		}
		// 	}
		// } else {
		// 	logBuffer.WriteString(fmt.Sprintf("[%s] ------ 响应体为空 ------\n", requestID))
		// }

		logBuffer.WriteString(fmt.Sprintf("[%s] ================================= 请求结束 =================================\n", requestID))

		// 使用互斥锁确保日志输出不交叉
		logMutex.Lock()
		fmt.Print(logBuffer.String())
		logMutex.Unlock()
	}
}
