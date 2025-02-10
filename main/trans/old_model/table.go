package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type Table struct {
	TableID        uint64 `gorm:"primaryKey;autoIncrement;comment:id"`
	TableNo        string `gorm:"default:'';comment:桌位编号"`
	Sort           uint   `gorm:"default:0;comment:排序"`
	AreaID         uint64 `gorm:"default:0;comment:区域id"`
	TypeID         uint64 `gorm:"default:0;comment:类型id"`
	Status         uint   `gorm:"default:10;comment:桌台状态 10-未开台 30-已开台"`
	SwitchStatus   uint   `gorm:"default:1;comment:桌台开关状态 0-关 1-开"`
	AreaName       string `gorm:"default:'';comment:区域名称"`
	TypeName       string `gorm:"default:'';comment:类型名称"`
	ShopSupplierID uint   `gorm:"default:0;comment:门店id"`
	MinNum         uint   `gorm:"default:0;comment:最小人数"`
	MaxNum         uint   `gorm:"default:0;comment:最大人数"`
	AppID          uint   `gorm:"default:0;comment:应用id"`
	BindInfo       string `gorm:"default:'';comment:绑定信息"`
	QRCodeValue    string `gorm:"default:'';comment:二维码值"`
	IsBind         uint   `gorm:"default:0;comment:平板绑定状态 0-否 1-是"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type TableRepository interface {
	GetTableList() ([]*Table, error)
	ConvertTable() error
}

type TableService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *TableService) GetTableList() ([]*Table, error) {
	var tables []*Table
	err := s.db.Find(&tables).Error
	return tables, err
}

func (s *TableService) ConvertTable() error {
	tables, err := s.GetTableList()
	if err != nil {
		return err
	}
	for _, table := range tables {
		fmt.Println(fmt.Sprintf("table: %+v", table))

		var isDisable uint
		if table.SwitchStatus == 1 {
			isDisable = 0
		} else {
			isDisable = 1
		}
		status := 0
		if table.Status == 30 {
			status = 1
		}
		desk := model.Desk{
			Uuid:       table.TableID,
			DeskNo:     table.TableNo,
			RegionUuid: table.AreaID,
			TypeUuid:   table.TypeID,
			Sort:       uint(table.Sort),
			Status:     uint(status),
			IsDisable:  isDisable,
			QrcodeUrl:  table.QRCodeValue,
			IsBind:     uint(table.IsBind),
			CreateTime: table.CreateTime,
			UpdateTime: table.UpdateTime,
			DeleteTime: 0,
		}

		_, err := repository.NewDeskRepo(s.targetDB).CreateDesk(desk)
		if err != nil {
			return err
		}
	}
	return nil
}
