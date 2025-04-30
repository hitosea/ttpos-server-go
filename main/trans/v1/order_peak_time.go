package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type OrderPeakTime struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Date           int     `gorm:"default:0;comment:'日期（天）'"`
	Hour           int     `gorm:"default:0;comment:'小时'"`
	Num            int     `gorm:"default:0;comment:'订单数'"`
	Amount         float64 `gorm:"type:decimal(12,2);default:0.00;comment:'订单金额'"`
	CashierID      int     `gorm:"default:0;comment:'收银员id'"`
	ShopSupplierID int     `gorm:"default:0;comment:'店铺id'"`
	AppID          int     `gorm:"default:0;comment:'应用id'"`
	CreateTime     int     `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime     int     `gorm:"not null;default:0;comment:'更新时间'"`
}

type OrderPeakTimeRepository interface {
	GetOrderPeakTimeList() ([]*OrderPeakTime, error)
	ConvertOrderPeakTime() error
}

func NewOrderPeakTimeService(db *gorm.DB, targetDB *gorm.DB) OrderPeakTimeRepository {
	return &OrderPeakTimeService{db: db, targetDB: targetDB}
}

type OrderPeakTimeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *OrderPeakTimeService) GetOrderPeakTimeList() ([]*OrderPeakTime, error) {
	var orderPeakTimes []*OrderPeakTime
	err := s.db.Find(&orderPeakTimes).Error
	return orderPeakTimes, err
}

func (s *OrderPeakTimeService) ConvertOrderPeakTime() error {
	orderPeakTimes, err := s.GetOrderPeakTimeList()
	if err != nil {
		return err
	}
	for _, orderPeakTime := range orderPeakTimes {
		fmt.Println(fmt.Sprintf("orderPeakTime: %+v", orderPeakTime))
		saleOrderPeakTime := model.SaleOrderPeakTime{
			BaseModel: model.BaseModel{
				Uuid:       orderPeakTime.ID,
				CreateTime: int64(orderPeakTime.CreateTime),
				UpdateTime: int64(orderPeakTime.UpdateTime),
			},
			Date:        int64(orderPeakTime.Date),
			Hour:        int64(orderPeakTime.Hour),
			Num:         orderPeakTime.Num,
			Amount:      orderPeakTime.Amount,
			CashierUuid: uint64(orderPeakTime.CashierID),
		}
		if err := repository.NewSaleOrderPeakTimeRepo(s.targetDB).CreateSaleOrderPeakTime(saleOrderPeakTime); err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
