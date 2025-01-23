package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/pgp"
	"ttpos-server-go/pkg/utils"
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

	// 1、客户端自己生成PGP密钥对，并传递一个随机字符串client_id，请求获取服务端公钥
	// 2、服务端生成的PGP密钥对，将client_id作为key将服务端PGP密钥对缓存起来
	// 3、服务端返回PGP公钥给客户端，客户端缓存该PGP公钥，后续该服务端PGP公钥加密请求
	// 4、客户端将请求加密后传递给服务端，请求头x-encrypt上带有 [客户端公钥] 和client_id，使用键值对的形式
	// 5、服务端解析请求头，根据client_id [从缓存中] 获取对应的服务端PGP密钥对，使用 [服务端私钥解密]
	// 6、服务端使用 [客户端公钥] 加密响应体，返回给客户端

	return func(c *gin.Context) {
		var (
			// 请求头x-encrypt形如：client_id=xxxx;client_public_key=xxxxxxx;type=pgp
			xEncrypt         = c.GetHeader(config.Pgp.EncryptHeader)
			encrypted        = xEncrypt != ""
			parsedXEncrypt   = utils.PGPParse(xEncrypt, config.Pgp.ClientPublicKey)
			serverPgpKeyPair pgp.KeyPair
		)

		cachedKeyPair, exists := cache.Get(config.Pgp.CachePrefix + parsedXEncrypt[config.Pgp.ClientID])
		if exists {
			_ = json.Unmarshal([]byte(cachedKeyPair.(string)), &serverPgpKeyPair)
			if serverPgpKeyPair.PublicKey == "" || serverPgpKeyPair.PrivateKey == "" || serverPgpKeyPair.Passphrase == "" {
				logger.Logger.Error("pgp秘钥对无效", zap.Any("pgp_key_pair", serverPgpKeyPair))
			}
		}

		// 自定义writer
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
				// 使用客户端公钥加密
				if encryptedResponse, err = pgp.EncryptMessage(string(rawResponse), parsedXEncrypt[config.Pgp.ClientPublicKey]); err != nil {
					logger.Logger.Error("pgp加密失败", zap.Error(err), zap.Any("parsed_encrypt", parsedXEncrypt))
				}
				encryptedResponse = regexp.MustCompile(`\s*-----(BEGIN|END) PGP MESSAGE-----\s*`).ReplaceAllString(encryptedResponse, "")
				b, _ := json.Marshal(map[string]string{"encrypted": encryptedResponse})
				return string(b)
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
			rawRequest, err := pgp.DecryptMessage("-----BEGIN PGP MESSAGE-----\n\n"+reqBody.Encrypted+"\n-----END PGP MESSAGE-----", serverPgpKeyPair.PrivateKey, serverPgpKeyPair.Passphrase)
			if err != nil {
				logger.Logger.Error("pgp解密失败", zap.Error(err), zap.Any("req_body", reqBody), zap.Any("server_pgp_key_pair", serverPgpKeyPair))
			}
			// 修改请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(rawRequest)))
		}

		c.Next()
	}
}
