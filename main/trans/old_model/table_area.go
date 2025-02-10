package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type TableArea struct {
	AreaID         uint64 `gorm:"primaryKey;autoIncrement;comment:'id'"`
	AreaName       string `gorm:"type:varchar(50);not null;comment:'区域名称'"`
	Sort           int    `gorm:"not null;default:0;comment:'排序'"`
	ShopSupplierID int    `gorm:"not null;default:0;comment:'门店id'"`
	AppID          uint   `gorm:"not null;default:0;comment:'应用id'"`
	CreateTime     uint   `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime     uint   `gorm:"not null;default:0;comment:'更新时间'"`
}

type TableAreaRepository interface {
	GetTableAreaList() ([]*TableArea, error)
	ConvertTableArea() error
}

type TableAreaService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *TableAreaService) GetTableAreaList() ([]*TableArea, error) {
	var tableAreas []*TableArea
	err := s.db.Find(&tableAreas).Error
	return tableAreas, err
}

func (s *TableAreaService) ConvertTableArea() error {
	tableAreas, err := s.GetTableAreaList()
	if err != nil {
		return err
	}
	for _, tableArea := range tableAreas {
		fmt.Println(fmt.Sprintf("tableArea: %+v", tableArea))
		deskRegion := model.DeskRegion{
			Uuid:       tableArea.AreaID,
			Name:       tableArea.AreaName,
			OrderBy:    uint(tableArea.Sort),
			CreateTime: int64(tableArea.CreateTime),
			UpdateTime: int64(tableArea.UpdateTime),
			DeleteTime: 0,
		}

		_, err := repository.NewDeskRegionRepo(s.targetDB).CreateDeskRegion(deskRegion)
		if err != nil {
			return err
		}
	}
	return nil
}
