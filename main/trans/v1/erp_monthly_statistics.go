package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ERPMonthlyStatistics struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Year           int     `gorm:"default:0;comment:'年'"`
	Month          int     `gorm:"default:0;comment:'月'"`
	RecordType     int     `gorm:"default:0;comment:'记录类型 1-月初 2-月末'"`
	Stock          float64 `gorm:"type:decimal(20,4);default:0.0000;comment:'月初库存'"`
	ShopSupplierID int     `gorm:"default:0;comment:'门店id'"`
	AppID          int     `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64   `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64   `gorm:"autoUpdateTime;comment:'更新时间'"`
}

func (model *ERPMonthlyStatistics) GetScene() int {
	if model.RecordType == 2 {
		return 1
	}
	return 0
}

type ERPMonthlyStatisticsRepository interface {
	GetERPMonthlyStatisticsList() ([]*ERPMonthlyStatistics, error)
	ConvertERPMonthlyStatistics() error
}

func NewErpMonthlyStatisticsService(db *gorm.DB, targetDB *gorm.DB) ERPMonthlyStatisticsRepository {
	return &ERPMonthlyStatisticsService{
		db:       db,
		targetDB: targetDB,
	}
}

type ERPMonthlyStatisticsService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ERPMonthlyStatisticsService) GetERPMonthlyStatisticsList() ([]*ERPMonthlyStatistics, error) {
	var erpMonthlyStatistics []*ERPMonthlyStatistics
	err := s.db.Find(&erpMonthlyStatistics).Error
	return erpMonthlyStatistics, err
}

func (s *ERPMonthlyStatisticsService) ConvertERPMonthlyStatistics() error {
	erpMonthlyStatistics, err := s.GetERPMonthlyStatisticsList()
	if err != nil {
		return err
	}
	for _, erp := range erpMonthlyStatistics {
		fmt.Println(fmt.Sprintf("erpMonthlyStatistics: %+v", erp))
		form := model.WarehouseMonthlyForm{
			BaseModel: model.BaseModel{
				Uuid: uint64(erp.ID),
			},
			Year:  erp.Year,
			Month: erp.Month,
			Scene: erp.GetScene(),
			Stock: erp.Stock,
		}
		err = repository.NewWarehouseMonthlyFormRepo(s.targetDB).CreateWarehouseMonthlyForm(form)
		if err != nil {
			return err
		}
	}
	return nil
}
