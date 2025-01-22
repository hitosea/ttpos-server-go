package req

type GetServerPublicKeyRequest struct {
	ClientId string `json:"client_id" form:"client_id" binding:"required"`
	Type     string `json:"type" form:"type" binding:"required,oneof=pgp jsencrypt"`
}

var GetServerPublicKeyRequestMessage = map[string]string{
	"client_id.required": "client_id不能为空",
}
