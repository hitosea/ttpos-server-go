package service

import (
	"ttpos-server-go/app/dto/resp/cashier_resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// IProductSrv 定义收银服务接口
type IOrderSrv interface {
	CreateOrder(dbId uint64) (cashier_resp.CreateOrderResp, error) // 创建订单
}

// orderSrv 收银服务结构体
type orderSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewOrderSrv 创建新的收银产品类别服务
func NewOrderSrv(dbm *database.DBManager) IOrderSrv {
	return NewOrderSrvImpl(dbm)
}

// NewOrderSrvImpl 创建新的收银服务实现
func NewOrderSrvImpl(dbm *database.DBManager) IOrderSrv {
	return &orderSrv{
		dbm: dbm,
	}
}

// CreateOrder 创建订单
func (s *orderSrv) CreateOrder(dbId uint64) (cashier_resp.CreateOrderResp, error) {
	// 创建销售账单
	db := s.dbm.GetDB(dbId)
	_ = db.Transaction(func(tx *gorm.DB) error {
		uuid, err := database.GetID()
		if err != nil {
			return err
		}
		_, err = repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			Uuid:            uuid,
			BillType:        0,
			DiningMethod:    0,
			IsBuffet:        0,
			IsLock:          0,
			MealNum:         0,
			Status:          0,
			Reason:          "",
			Remark:          "",
			OrderAmount:     0,
			ProductAmount:   0,
			PaymentAmount:   0,
			ConsumerUuid:    0,
			CashierUuid:     0,
			BuffetOrderUUID: 0,
			TableUuid:       0,
			BuffetDuration:  0,
			HideBillTime:    0,
			FinishTime:      0,
			CreateTime:      0,
			UpdateTime:      0,
			DeleteTime:      0,
			SaleOrders:      []model.SaleOrder{},
		})
		if err != nil {
			return err
		}
		return nil
	})
	return cashier_resp.CreateOrderResp{}, nil
}
