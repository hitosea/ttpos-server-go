package req

type GetServerPublicKeyRequest struct {
	ClientId string `form:"client_id" binding:"required"`
}

var GetServerPublicKeyRequestMessage = map[string]string{
	"client_id.required": "client_id不能为空",
}
