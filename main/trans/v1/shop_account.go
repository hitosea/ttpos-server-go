package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopAccount struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Amount         float64 `gorm:"type:decimal(12,2);default:0.00;comment:'账户余额'"`
	ShopSupplierID int     `gorm:"default:0;comment:'店铺id'"`
	AppID          int     `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64   `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64   `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type ShopAccountRepository interface {
	GetShopAccountList() ([]*ShopAccount, error)
	ConvertShopAccount() error
}

func NewShopAccountService(db *gorm.DB, targetDB *gorm.DB) ShopAccountRepository {
	return &ShopAccountService{
		db:       db,
		targetDB: targetDB,
	}
}

type ShopAccountService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopAccountService) GetShopAccountList() ([]*ShopAccount, error) {
	var shopAccounts []*ShopAccount
	err := s.db.Find(&shopAccounts).Error
	return shopAccounts, err
}

func (s *ShopAccountService) ConvertShopAccount() error {
	shopAccounts, err := s.GetShopAccountList()
	if err != nil {
		return err
	}
	for _, shopAccount := range shopAccounts {
		fmt.Println(fmt.Sprintf("shopAccount: %+v", shopAccount))
		box := model.CashBox{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopAccount.ID),
				CreateTime: shopAccount.CreateTime,
				UpdateTime: shopAccount.UpdateTime,
			},
			Name:            "",
			Balance:         shopAccount.Amount,
			FrozenBalance:   0,
			PreviousBalance: 0,
			CashWithdrawal:  0,
			CashDeposit:     0,
		}
		_, err = repository.NewCashBoxRepo(s.targetDB).Create(box)
		if err != nil {
			return err
		}
	}
	return nil
}
