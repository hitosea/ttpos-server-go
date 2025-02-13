package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type TableType struct {
	TypeID         uint64 `gorm:"primaryKey;autoIncrement;comment:'id'"`
	TypeName       string `gorm:"type:varchar(50);not null;comment:'桌位类型'"`
	Sort           int    `gorm:"not null;default:0;comment:'排序'"`
	MinNum         int    `gorm:"not null;default:0;comment:'最小人数'"`
	MaxNum         int    `gorm:"not null;comment:'最大人数'"`
	ShopSupplierID int    `gorm:"not null;default:0;comment:'门店id'"`
	AppID          int    `gorm:"not null;default:0;comment:'应用id'"`
	CreateTime     int    `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime     int    `gorm:"not null;default:0;comment:'更新时间'"`
}

type TableTypeRepository interface {
	GetTableTypeList() ([]*TableType, error)
	ConvertTableType() error
}

type TableTypeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *TableTypeService) GetTableTypeList() ([]*TableType, error) {
	var tableTypes []*TableType
	err := s.db.Find(&tableTypes).Error
	return tableTypes, err
}

func (s *TableTypeService) ConvertTableType() error {
	tableTypes, err := s.GetTableTypeList()
	if err != nil {
		return err
	}
	for _, tableType := range tableTypes {
		fmt.Println(fmt.Sprintf("tableType: %+v", tableType))
		deskType := model.DeskType{
			BaseModel: model.BaseModel{
				Uuid:       uint64(tableType.TypeID),
				CreateTime: int64(tableType.CreateTime),
				UpdateTime: int64(tableType.UpdateTime),
			},
			Name:     tableType.TypeName,
			Sort:     uint(tableType.Sort),
			RangeMin: uint(tableType.MinNum),
			RangeMax: uint(tableType.MaxNum),
		}

		_, err := repository.NewDeskTypeRepo(s.targetDB).CreateDeskType(deskType)
		if err != nil {
			return err
		}
	}
	return nil
}
