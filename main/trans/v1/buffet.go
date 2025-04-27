package v1

import (
	"fmt"

	"gorm.io/gorm"
)

type Buffet struct {
	ID                       uint    `gorm:"primaryKey;autoIncrement;not null"`
	Name                     string  `gorm:"type:varchar(2000);default:'';comment:'名称'"`
	Price                    float64 `gorm:"type:decimal(12,2);default:0.00;comment:'价格'"`
	BuyLimitStatus           int     `gorm:"type:int;default:0;comment:'是否限购 0-否 1-是'"`
	IsComb                   int     `gorm:"type:int;default:0;comment:'是否组合 0-否 1-是'"`
	SaleNum                  int     `gorm:"type:int;default:0;comment:'销量'"`
	TimeLimit                int     `gorm:"type:int;default:0;comment:'用餐时间（分）'"`
	Status                   int     `gorm:"type:int;default:1;comment:'状态'"`
	Sort                     int     `gorm:"type:int;default:0;comment:'排序'"`
	IsRemainContinue         int     `gorm:"type:int;default:0;comment:'平板是否可继续点餐开关 0-关闭 1-开启'"`
	RemainContinueTime       int     `gorm:"type:int;default:0;comment:'剩余xx分不可继续点餐'"`
	RemainContinueNoticeTime int     `gorm:"type:int;default:0;comment:'剩余xx分提醒不可继续点餐'"`
	AppID                    int     `gorm:"type:int;default:0;comment:'应用id'"`
	ShopSupplierID           int     `gorm:"type:int;default:0;comment:'门店id'"`
	CreateTime               int64   `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime               int64   `gorm:"type:int;not null;default:0;comment:'更新时间'"`
	DeleteTime               int64   `gorm:"type:int unsigned;not null;default:0;comment:'删除时间'"`

	BuffetTaxes     []*BuffetTax      `gorm:"foreignKey:BuffetID;references:ID"`
	BuffetCustomers []*BuffetCustomer `gorm:"foreignKey:BuffetID;references:ID"`
	BuffetProducts  []*BuffetProduct  `gorm:"foreignKey:BuffetID;references:ID"`
}

type BuffetTax struct {
	ID            uint  `gorm:"primaryKey;autoIncrement;not null"`
	BuffetID      int   `gorm:"type:int;default:0;comment:'自助餐id'"`
	BuffetTaxType int   `gorm:"type:int;default:1;comment:'自助餐税类类型，1-堂食税类'"`
	TaxCategoryID int   `gorm:"type:int;default:0;comment:'税类id'"`
	AppID         int   `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime    int64 `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime    int64 `gorm:"type:int;not null;default:0;comment:'更新时间'"`
}

type BuffetProduct struct {
	ID              uint  `gorm:"primaryKey;autoIncrement;not null"`
	BuffetID        int   `gorm:"type:int;default:0;comment:'关联id'"`
	ProductID       int   `gorm:"type:int;default:0;comment:'商品id'"`
	LimitNum        int   `gorm:"type:int;default:0;comment:'限购数量'"`
	IsShowCashier   int   `gorm:"type:int;default:1;comment:'是否显示在收银端 1-显示 2-不显示'"`
	IsShowTablet    int   `gorm:"type:int;default:1;comment:'是否显示在平板端 1-显示 2-不显示'"`
	IsShowKitchen   int   `gorm:"type:int;default:1;comment:'是否显示在厨显端 1-显示 2-不显示'"`
	IsShowAssistant int   `gorm:"type:int;default:1;comment:'是否显示在点餐助手 1-显示 2-不显示'"`
	IsShowH5        int   `gorm:"type:int;default:1;comment:'是否显示在h5 1-显示 2-不显示'"`
	AppID           int   `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime      int64 `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime      int64 `gorm:"type:int;not null;default:0;comment:'更新时间'"`
}

// 是否限时 0-否、1-是
func (b *Buffet) GetIsTimeLimit() uint {
	if b.TimeLimit > 0 {
		return 1
	}
	return 0
}

func (b *Buffet) GetBuffetTaxUuid() uint64 {
	if len(b.BuffetTaxes) > 0 {
		for _, buffetTax := range b.BuffetTaxes {
			if buffetTax.BuffetTaxType == 1 {
				return uint64(buffetTax.TaxCategoryID)
			}
		}
	}
	return 0
}

type BuffetRepository interface {
	GetBuffetList() ([]*Buffet, error)
	ConvertBuffet() error
}

func NewBuffetService(db *gorm.DB, targetDB *gorm.DB) BuffetRepository {
	return &BuffetService{
		db:       db,
		targetDB: targetDB,
	}
}

type BuffetService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *BuffetService) GetBuffetList() ([]*Buffet, error) {
	var buffets []*Buffet
	err := s.db.Find(&buffets).Error
	return buffets, err
}

func (s *BuffetService) ConvertBuffet() error {
	buffets, err := s.GetBuffetList()
	if err != nil {
		return err
	}
	for _, buffet := range buffets {
		fmt.Println(fmt.Sprintf("buffet: %+v", buffet))

	}
	return nil
}
