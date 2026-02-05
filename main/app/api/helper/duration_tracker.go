package helper

import (
	"time"

	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/pkg/context"
)

// DurationRecordPusher 记录推送函数类型
// 由外部设置，避免循环依赖
type DurationRecordPusher func(record *req.OperationDurationRecord)

// durationRecordPusher 全局推送函数
// 在队列初始化时设置
var durationRecordPusher DurationRecordPusher

// SetDurationRecordPusher 设置记录推送函数
// 在队列初始化时调用
func SetDurationRecordPusher(pusher DurationRecordPusher) {
	durationRecordPusher = pusher
}

// DurationTracker 操作耗时跟踪器
// 用于在 Handler 层简洁地记录操作耗时
type DurationTracker struct {
	startTime     int64  // 开始时间（毫秒）
	action        string // 操作类型
	requestPath   string // 请求路径
	saleBillUuid  uint64 // 账单 UUID
	saleOrderUuid uint64 // 订单 UUID
}

// StartTrack 开始跟踪操作耗时
// ctx: 上下文，用于获取请求开始时间
// action: 操作类型，如 "cancel_order", "checkout", "discount"
func StartTrack(ctx context.Context, action string) *DurationTracker {
	startTime := ctx.GetStartTime()
	if startTime == 0 {
		startTime = time.Now().UnixMilli() // 兜底：如果 ctx 没有开始时间则使用当前时间
	}
	return &DurationTracker{
		startTime: startTime,
		action:    action,
	}
}

// WithBill 设置账单信息
func (t *DurationTracker) WithBill(saleBillUuid, saleOrderUuid uint64) *DurationTracker {
	t.saleBillUuid = saleBillUuid
	t.saleOrderUuid = saleOrderUuid
	return t
}

// WithPath 设置请求路径
func (t *DurationTracker) WithPath(path string) *DurationTracker {
	t.requestPath = path
	return t
}

// End 结束跟踪并推送记录到队列
// ctx: 上下文，用于获取 company_uuid、staff_uuid 等信息
// err: 操作结果，nil 表示成功
func (t *DurationTracker) End(ctx context.Context, err error) {
	// 检查推送函数是否已设置
	if durationRecordPusher == nil {
		return
	}

	endTime := time.Now().UnixMilli()
	durationMs := endTime - t.startTime

	// 确定操作状态
	status := 1
	errMsg := ""
	if err != nil {
		status = 0
		errMsg = err.Error()
	}

	// 构建记录
	record := &req.OperationDurationRecord{
		CompanyUuid:   ctx.GetCompanyUuid(),
		SaleBillUuid:  t.saleBillUuid,
		SaleOrderUuid: t.saleOrderUuid,
		Action:        t.action,
		Source:        ctx.GetSource(),
		StaffUuid:     ctx.GetStaffUuid(),
		DeviceSn:      ctx.GetDeviceSn(),
		StartTime:     t.startTime,
		EndTime:       endTime,
		DurationMs:    durationMs,
		RequestPath:   t.requestPath,
		Status:        status,
		ErrorMsg:      errMsg,
	}

	// 推送到队列（非阻塞）
	durationRecordPusher(record)
}
