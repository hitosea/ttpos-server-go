package member_resp

import "ttpos-server-go/app/dto"

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

type PointsRecord struct {
	Scene      int     `json:"scene"`       //场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减
	Value      float64 `json:"value"`       //数值,负数:减积分 正数:加积分
	CreateTime int64   `json:"create_time"` //创建时间
}

type PointsRecordListWithPaginationResp struct {
	List       []PointsRecord   `json:"list"`
	Meta       dto.PageResponse `json:"meta"`
	TotalPoint float64          `json:"total_point"` // 总积分, 用于计算总积分
}
