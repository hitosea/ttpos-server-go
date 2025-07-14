package member_resp

type Coupon struct {
	Uuid                 uint64  `json:"uuid"`
	Name                 string  `json:"name"`                   // 优惠券名称
	Amount               float64 `json:"amount"`                 // 优惠券面值
	StartTime            int64   `json:"start_time"`             // 优惠券开始时间
	EndTime              int64   `json:"end_time"`               // 优惠券结束时间
	ApplicableTimePeriod int     `json:"applicable_time_period"` // 适用时间段 全天 0，指定时间 1
	DayStartTime         string  `json:"day_start_time"`         // 每日适用时段开始时间, hh:mm 格式
	DayEndTime           string  `json:"day_end_time"`           // 每日适用时段结束时间, hh:mm 格式
	Status               int     `json:"status"`                 // 优惠券状态 0: 未使用 1: 已使用 2: 已过期
}

type CouponListWithPaginationResp struct {
	List []Coupon `json:"list"`
}
