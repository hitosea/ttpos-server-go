package dto

// OrderWithDriver 订单及司机信息聚合 DTO
type OrderWithDriver struct {
	OrderUuid      string // 订单 UUID
	ProviderName   string // 供应商名称
	PartnerOrderId string // 平台订单号
	OrderStatus    string // 订单状态
	// Skootar 司机信息
	SkootarId       string  // 骑手ID
	SkootarName     string  // 骑手名称
	SkootarPhone    string  // 骑手电话
	SkootarImageUrl string  // 骑手头像
	SkootarRating   float64 // 骑手评分
}

// SkootarJob Skootar 订单信息 DTO（兼容旧接口）
type SkootarJob struct {
	Uuid            string  // 外送订单UUID
	ShopRefNo       string  // 餐馆订单参考，如UUID
	TakeoutRefNo    string  // 外送系统订单号
	ProviderName    string  // 外送供应商
	JobStatus       string  // 外送订单状态
	SkootarId       string  // 骑手ID
	SkootarName     string  // 骑手名称
	SkootarPhone    string  // 骑手电话
	SkootarImageUrl string  // 骑手头像
	SkootarRating   float64 // 骑手评分
}

// PrepareOrderReq 准备订单请求 DTO（接受/拒绝订单）
type PrepareOrderReq struct {
	TakeoutOrderUuid string `json:"takeout_order_uuid" v:"required#订单UUID不能为空"`                                   // TTPOS订单UUID
	ToState          string `json:"to_state" v:"required|in:Accepted,Rejected#目标状态不能为空|目标状态必须为Accepted或Rejected"` // 目标状态: Accepted/Rejected
	RequestId        string `json:"request_id"`                                                                   // 请求追踪ID (可选)
}

// PrepareOrderResp 准备订单响应 DTO
type PrepareOrderResp struct {
	OrderUuid string `json:"order_uuid"` // 订单UUID
}
