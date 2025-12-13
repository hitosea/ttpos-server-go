package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/encrypt"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EncryptWriter struct {
	gin.ResponseWriter

	// 这里应该有一个方法处理响应体
	fn func([]byte) string
}

func (w EncryptWriter) Write(b []byte) (int, error) {
	w.Header().Del("Content-Length")
	// 调用w处理响应的方法
	s := w.fn(b)
	return w.ResponseWriter.Write([]byte(s))
}

func (w EncryptWriter) WriteString(s string) (int, error) {
	w.Header().Del("Content-Length")
	s1 := w.fn([]byte(s))
	return w.ResponseWriter.WriteString(s1)
}

// Encrypt 接口加密
func Encrypt(cache cache.Cache) gin.HandlerFunc {

	return func(c *gin.Context) {
		var (
			// 请求头x-encrypt形如：client_id=xxxx;client_key=xxxxxxx;type=jsencrypt
			xEncrypt       = c.GetHeader(config.Encrypt.EncryptHeader)
			encrypted      = xEncrypt != ""
			parsedXEncrypt = utils.ParseEncrypt(xEncrypt, config.Encrypt.ClientKey)
			encryptType    = "jsencrypt"
			keyPair        encrypt.KeyPair
		)

		if typ, ok := parsedXEncrypt["encrypt_type"]; ok {
			encryptType = typ
		}

		cachedKeyPair, exists := cache.Get(config.Encrypt.CachePrefix + parsedXEncrypt[config.Encrypt.ClientID] + "_" + encryptType)
		if exists {
			_ = json.Unmarshal([]byte(cachedKeyPair.(string)), &keyPair)
			if keyPair.PublicKey == "" || keyPair.PrivateKey == "" {
				logger.Logger.Error("rsa秘钥对无效")
			}
		}

		// 自定义加密writer
		writer := &EncryptWriter{
			ResponseWriter: c.Writer,
			fn: func(rawResponse []byte) string {
				if !encrypted {
					return string(rawResponse)
				}
				var (
					encryptedResponse string
					err               error
				)
				if encryptedResponse, err = encrypt.JsEncryptMessage(string(rawResponse), parsedXEncrypt[config.Encrypt.ClientKey]); err != nil {
					logger.Logger.Error("jsencrypt加密失败", zap.Error(err), zap.Any("parsed_encrypt", parsedXEncrypt))
				} else {
					b, _ := json.Marshal(map[string]string{"encrypted": encryptedResponse})
					return string(b)
				}
				return string(rawResponse)
			},
		}
		// 覆盖原writer
		c.Writer = writer

		// 使用服务器私钥解密
		if encrypted {
			var reqBody struct {
				Encrypted string `json:"encrypted"`
			}
			body, _ := io.ReadAll(c.Request.Body)
			_ = json.Unmarshal(body, &reqBody)

			rawRequestBody, err := encrypt.JsDecryptMessage(keyPair.PrivateKey, reqBody.Encrypted)
			if err != nil {
				logger.Logger.Error("jsencrypt解密失败", zap.Error(err))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(rawRequestBody))
			}
		}

		c.Next()
	}
}
