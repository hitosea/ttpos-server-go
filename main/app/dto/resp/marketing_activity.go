package resp

// DecryptQrCodeResp 解密活动二维码响应
type DecryptQrCodeResp struct {
	Uuid         uint64 `json:"uuid"`          // 会员uuid
	Nickname     string `json:"nickname"`      // 会员昵称
	Phone        string `json:"phone"`         // 会员手机号
	ActivityUuid uint64 `json:"activity_uuid"` // 活动uuid
}
