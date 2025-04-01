package model

import (
	"ttpos-server-go/app/model"
	v1 "ttpos-server-go/trans/v1"
)

func NewOrder(order *v1.Order) (*model.SaleBill, error) {
	// saleBillUuid := uint64(order.OrderID)
	// orderNo := order.OrderNo
	// buffetUuids := make([]uint64, 0)
	// mealNum := order.MealNum
	// remark := order.TableRemark
	// deskUuid := uint64(order.TableID)
	// deskNo := order.TableNo
	// deskNo := order.DutyNo
	// staffUuid := order.StaffUuid
	// saleBill := model.NewDeskSaleBill(saleBillUuid, orderNo, buffetUuids, mealNum, remark, deskUuid, deskNo, ctx.GetStaff().DutyNo, ctx.GetStaff().Uuid, ctx.GetStaff().Username)

	return nil, nil
}
