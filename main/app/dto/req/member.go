package req

type AddMemberReq struct {
	LevelUuid uint64 `json:"level_uuid" binding:"required"`                    // 会员等级Uuid
	Phone     string `json:"phone" binding:"required,max=20"`                  // 手机号
	Nickname  string `json:"nickname" binding:"omitempty,max=50"`              // 昵称
	Password  string `json:"password" binding:"omitempty,number,min=4,max=16"` // 密码
}

var AddMemberReqMessage = map[string]string{
	"level_uuid.required": "会员等级不存在",
	"phone.max":           "手机号不能超过20个字符",
	"nickname.max":        "昵称不能超过50个字符",
	"password.number":     "密码必须为4-16位纯数字",
	"password.min":        "密码必须为4-16位纯数字",
	"password.max":        "密码必须为4-16位纯数字",
}
