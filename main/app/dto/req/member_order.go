package req

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"
)

// MemberOrderListReq 会员端订单列表
type MemberOrderListReq struct {
	dto.PageReq        // 分页参数
	Status      string `form:"status"` // 状态: "all" 全部, "unpaid" 待付款, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
}

// GetMemberOrderDetailReq 外送订单详情
type GetMemberOrderDetailReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid"` // 会员端销售订单UUID
}

// GetMemberOrderManageDetailReq 订单详情
type GetMemberOrderManageDetailReq struct {
	MemberSaleOrderUuid uint64 `form:"member_sale_order_uuid"` // 会员端销售订单UUID
}

// MemberOrderManageListReq 外送订单管理列表查询
type MemberOrderManageListReq struct {
	dto.PageReq           // 分页参数
	Status         string `form:"status"`           // 账单状态, ""=全都, “unpaid”=待付款 "unaccept"=待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
	OrderNo        string `form:"order_no"`         // 订单编号
	SerialNo       string `form:"serial_no"`        // 订单序号
	DateRange      int    `form:"date_range"`       // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	TimeType       int    `form:"time_type"`        // 时间类型  1=下单时间、 2=支付时间
	QueryStartTime int64  `form:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time"`   // 查询结束时间戳
}

type TimeFilterParams struct {
	TimeType       int   // 时间类型  1=下单时间、 2=支付时间
	QueryStartTime int64 // 查询开始时间戳
	QueryEndTime   int64 // 查询结束时间戳
}

func (req *MemberOrderManageListReq) GetTimeFilterParams(timezone string) *TimeFilterParams {
	if req.QueryEndTime != 0 && req.QueryStartTime != 0 {
		// 通过自定义时间筛选
		if req.TimeType == 0 {
			req.TimeType = 1 // 默认选择下单时间
		}
		return &TimeFilterParams{
			TimeType:       req.TimeType,
			QueryStartTime: req.QueryStartTime,
			QueryEndTime:   req.QueryEndTime,
		}
	}
	// 日期类型 -1-全都 0-今天 1-昨天 2-本周
	if req.DateRange >= 0 && req.DateRange <= 2 {
		var startTime, endTime int64
		switch req.DateRange {
		case constant.OrderDateTypeToday: // 今天
			startTime, endTime, _ = utils.SetTimezone(timezone).GetTimeRange(utils.DayTypeToday)
		case constant.OrderDateTypeYesterday: // 昨天
			startTime, endTime, _ = utils.SetTimezone(timezone).GetTimeRange(utils.DayTypeYesterday)
		case constant.OrderDateTypeWeek: // 本周
			startTime, endTime, _ = utils.SetTimezone(timezone).GetTimeRange(utils.DayTypeThisWeek)
		}
		// 通过自定义时间筛选
		if req.TimeType == 0 {
			req.TimeType = 1 // 默认选择下单时间
		}
		return &TimeFilterParams{
			TimeType:       req.TimeType,
			QueryStartTime: startTime,
			QueryEndTime:   endTime,
		}
	}
	// 没有时间筛选条件
	return nil
}

// RejectOrderReq 拒单
type RejectOrderReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
}

// AcceptOrderReq 接单
type AcceptOrderReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
}
