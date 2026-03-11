package req

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/utils"
)

// MemberOrderListReq 会员端订单列表
type MemberOrderListReq struct {
	dto.PageReq        // 分页参数
	Status      string `form:"status"`  // 状态: "all" 全部, "unpaid" 待付款, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
	Keyword     string `form:"keyword"` // 关键字
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
	Status         string `form:"status"`           // 账单状态, ""=全都, "unpaid"=待付款 "unaccept"=待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
	OrderNo        string `form:"order_no"`         // 订单编号
	SerialNo       string `form:"serial_no"`        // 订单序号
	DateRange      int    `form:"date_range"`       // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周
	TimeType       int    `form:"time_type"`        // 时间类型  1=下单时间、 2=支付时间
	QueryStartTime int64  `form:"query_start_time"` // 查询开始时间戳
	QueryEndTime   int64  `form:"query_end_time"`   // 查询结束时间戳
	QueryStartDate string `form:"query_start_date"` // 查询开始日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
	QueryEndDate   string `form:"query_end_date"`   // 查询结束日期时间（格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）
}

type TimeFilterParams struct {
	TimeType       int   // 时间类型  1=下单时间、 2=支付时间
	QueryStartTime int64 // 查询开始时间戳
	QueryEndTime   int64 // 查询结束时间戳
}

func (req *MemberOrderManageListReq) GetTimeFilterParams(timezone string) *TimeFilterParams {
	// 处理日期时间字符串参数（优先级：DateRange > QueryStartTime/QueryEndTime > QueryStartDate/QueryEndDate）
	queryStartTime := req.QueryStartTime
	queryEndTime := req.QueryEndTime

	// 1. 优先处理日期类型
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
		queryStartTime = startTime
		queryEndTime = endTime
	}

	// 2. 其次处理时间戳参数（如果设置了，会覆盖 DateRange）
	if req.QueryStartTime != 0 && req.QueryEndTime != 0 {
		queryStartTime = req.QueryStartTime
		queryEndTime = req.QueryEndTime
	}

	// 3. 最后处理日期时间字符串参数（如果设置了，会覆盖前面的）
	if req.QueryStartDate != "" && req.QueryEndDate != "" {
		timeUtil := utils.SetTimezone(timezone)
		startTime, err := timeUtil.FormatDateTimeToUnix(req.QueryStartDate)
		if err == nil {
			queryStartTime = startTime
		}
		endTime, err := timeUtil.FormatDateTimeToUnix(req.QueryEndDate)
		if err == nil {
			queryEndTime = endTime
		}
	}

	// 如果有时间范围，返回时间过滤参数
	if queryStartTime != 0 && queryEndTime != 0 {
		if req.TimeType == 0 {
			req.TimeType = 1 // 默认选择下单时间
		}
		return &TimeFilterParams{
			TimeType:       req.TimeType,
			QueryStartTime: queryStartTime,
			QueryEndTime:   queryEndTime,
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

// CookFinishOrderReq 备餐完成
type CookFinishOrderReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
}

type MemberOrderSearchReq struct {
	Keyword string `form:"keyword"` // 关键字
}

type MemberOrderSendAuthCodeReq struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID
}

// MemberDineInOrderListReq 会员端堂食订单列表
type MemberDineInOrderListReq struct {
	dto.PageReq        // 分页参数
	Status      string `form:"status"` // 状态: "all" 全部, "unpaid" 待支付, "inprogress" 进行中(已支付,待接单/备餐中), "completed" 已完成(含部分退款/全部退款), "cancelled" 已取消(含已拒单)
}

// GetMemberDineInOrderDetailReq 堂食订单详情
type GetMemberDineInOrderDetailReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid" binding:"required"` // 销售账单UUID
}

// MockPayDineInOrderCallbackReq 模拟堂食订单支付回调请求（仅用于测试）
type MockPayDineInOrderCallbackReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required"` // 销售账单UUID
}
