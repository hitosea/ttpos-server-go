package resp

// VisitorInfoResp 游客信息响应
type VisitorInfoResp struct {
	MemberUuid uint64 `json:"member_uuid"` // 会员UUID
	Nickname   string `json:"nickname"`    // 随机昵称
	DeviceId   string `json:"device_id"`   // 设备ID
	IsVisitor  bool   `json:"is_visitor"`  // 是否游客
}
