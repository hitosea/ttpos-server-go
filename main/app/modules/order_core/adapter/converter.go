package adapter

import (
	"strconv"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/modules/order_core/dto"
	"ttpos-server-go/pkg/utils"
)

// ToCoreCreateDeskOrderReq converts legacy DeskOrderCreateReq to core CreateOrderReq.
func ToCoreCreateDeskOrderReq(oldReq *req.DeskOrderCreateReq) *dto.CreateOrderReq {
	if oldReq == nil {
		return nil
	}
	id, _ := utils.GetID()
	orderNo := strconv.FormatUint(id, 10)

	return &dto.CreateOrderReq{
		OrderNo:  orderNo,
		BillType: 0, // Desk Order
		Orders: []dto.CreateOrder{
			{
				OrderNo:  orderNo,
				Products: []dto.CreateOrderProduct{},
			},
		},
	}
}

// ToCoreCreateInstantOrderReq converts legacy CreateInstantOrder call to core CreateOrderReq.
func ToCoreCreateInstantOrderReq() *dto.CreateOrderReq {
	id, _ := utils.GetID()
	orderNo := strconv.FormatUint(id, 10)

	return &dto.CreateOrderReq{
		OrderNo:  orderNo,
		BillType: 1, // Instant Order
		Orders: []dto.CreateOrder{
			{
				OrderNo:  orderNo,
				Products: []dto.CreateOrderProduct{},
			},
		},
	}
}

// ToCorePayOrderReq converts legacy InstantOrderPaymentCreateReq to core PayOrderReq.
func ToCorePayOrderReq(oldReq *req.InstantOrderPaymentCreateReq) *dto.PayOrderReq {
	if oldReq == nil {
		return nil
	}
	return &dto.PayOrderReq{
		BillUuid:          oldReq.SaleBillUuid,
		OrderUuid:         oldReq.SaleOrderUuid,
		PaymentMethodUuid: oldReq.PaymentMethodUuid,
		Amount:            oldReq.PaymentAmount,
	}
}

// ToCoreCancelOrderReq converts legacy OrderCancelReq to core CancelOrderReq.
func ToCoreCancelOrderReq(oldReq *req.OrderCancelReq) *dto.CancelOrderReq {
	if oldReq == nil {
		return nil
	}
	return &dto.CancelOrderReq{
		BillUuid: oldReq.SaleBillUuid,
		Reason:   oldReq.CancelReason,
	}
}
