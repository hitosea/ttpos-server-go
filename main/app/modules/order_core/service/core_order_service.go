package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"ttpos-server-go/app/modules/order_core/dto"
	"ttpos-server-go/app/modules/order_core/model"
	"ttpos-server-go/app/modules/order_core/repository"
	pkg_context "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type CoreOrderService struct {
	dbm *database.DBManager
}

func NewCoreOrderService(dbm *database.DBManager) ICoreOrderService {
	return &CoreOrderService{dbm: dbm}
}

func (s *CoreOrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.CreateOrderResp, error) {
	// 获取DB
	var db *gorm.DB
	if c, ok := ctx.(pkg_context.Context); ok {
		db = c.GetDB()
	} else {
		// Fallback or error if context is not compatible
		return nil, errors.New("invalid context")
	}

	// 生成 Bill UUID
	billUuid, _ := utils.GetID()
	now := time.Now().Unix()

	bill := &model.CoreSaleBill{
		BaseModel: model.BaseModel{
			Uuid:       billUuid,
			CreateTime: now,
			UpdateTime: now,
		},
		OrderNo:  req.OrderNo,
		BillType: req.BillType,
		Status:   0, // Pending
		Amount:   req.Amount,
	}

	var orderUuids []uint64

	err := db.Transaction(func(tx *gorm.DB) error {
		txRepo := repository.NewCoreOrderRepo(tx)
		if err := txRepo.CreateBill(bill); err != nil {
			return err
		}

		for _, orderReq := range req.Orders {
			orderUuid, _ := utils.GetID()
			orderUuids = append(orderUuids, orderUuid)
			order := &model.CoreSaleOrder{
				BaseModel: model.BaseModel{
					Uuid:       orderUuid,
					CreateTime: now,
					UpdateTime: now,
				},
				SaleBillUuid: billUuid,
				OrderNo:      orderReq.OrderNo,
				Status:       0, // Pending
				Amount:       orderReq.Amount,
			}
			if err := txRepo.CreateOrder(order); err != nil {
				return err
			}

			for _, prodReq := range orderReq.Products {
				prodUuid, _ := utils.GetID()
				prod := &model.CoreSaleOrderProduct{
					BaseModel: model.BaseModel{
						Uuid:       prodUuid,
						CreateTime: now,
						UpdateTime: now,
					},
					SaleBillUuid:  billUuid,
					SaleOrderUuid: orderUuid,
					Name:          prodReq.Name.ToJson(),       // 支持多语言JSON字符串
					FlavorName:    prodReq.FlavorName.ToJson(), // 支持多语言JSON字符串
					Num:           prodReq.Num,
					Status:        0, // Uncooked
					SalePrice:     prodReq.SalePrice,
					Price:         prodReq.Price,
					TotalPrice:    prodReq.TotalPrice,
				}
				if err := txRepo.CreateOrderProduct(prod); err != nil {
					return err
				}
			}

			// 发布 OrderCreated 事件
			// eventBus.Publish(event.OrderCreatedEvent{BillUuid: billUuid, OrderUuid: orderUuid})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.CreateOrderResp{BillUuid: billUuid, OrderUuids: orderUuids}, nil
}

func (s *CoreOrderService) MarkAsPaid(ctx context.Context, billUuid uint64) error {
	// 获取DB
	var db *gorm.DB
	if c, ok := ctx.(pkg_context.Context); ok {
		db = c.GetDB()
	} else {
		return errors.New("invalid context")
	}

	repo := repository.NewCoreOrderRepo(db)
	bill, err := repo.GetBillByUuid(billUuid)
	if err != nil {
		return err
	}
	if bill == nil {
		return errors.New("bill not found")
	}

	// 状态机检查: 只能从 0 (Pending) -> 1 (Completed)
	if bill.Status != 0 {
		return fmt.Errorf("invalid status transition from %d to 1", bill.Status)
	}

	// 更新状态
	if err := repo.UpdateBillStatus(billUuid, 1); err != nil {
		return err
	}

	// 发布事件
	// eventBus.Publish(event.OrderPaidEvent{BillUuid: billUuid, PayTime: time.Now().Unix()})

	return nil
}

func (s *CoreOrderService) FinishOrder(ctx context.Context, billUuid uint64) error {
	// 这里 FinishOrder 暂时等同于 MarkAsPaid 或者作为后续归档操作
	// 如果是 1 -> 1 (幂等)
	return s.MarkAsPaid(ctx, billUuid)
}

func (s *CoreOrderService) CancelOrder(ctx context.Context, billUuid uint64) error {
	// 获取DB
	var db *gorm.DB
	if c, ok := ctx.(pkg_context.Context); ok {
		db = c.GetDB()
	} else {
		return errors.New("invalid context")
	}

	repo := repository.NewCoreOrderRepo(db)
	bill, err := repo.GetBillByUuid(billUuid)
	if err != nil {
		return err
	}
	if bill == nil {
		return errors.New("bill not found")
	}

	// 状态机检查: 只能从 0 (Pending) -> 2 (Cancelled)
	if bill.Status != 0 {
		return fmt.Errorf("invalid status transition from %d to 2", bill.Status)
	}

	// 更新状态
	if err := repo.UpdateBillStatus(billUuid, 2); err != nil {
		return err
	}

	return nil
}
