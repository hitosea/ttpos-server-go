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
