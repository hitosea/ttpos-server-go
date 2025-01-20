package resp

type ServerPublicKeyResponse struct {
	Type            string `json:"type"`              // 类型
	ClientId        string `json:"client_id"`         // 客户端随机字符串
	ServerPublicKey string `json:"server_public_key"` // 服务端公钥
}
