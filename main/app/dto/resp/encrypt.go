package resp

type ServerKeyResponse struct {
	Type      string `json:"type"` // 类型
	ClientId  string `json:"id"`   // 客户端随机字符串
	ServerKey string `json:"key"`  // 服务端公钥
}
