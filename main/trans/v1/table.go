package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type Table struct {
	TableID        int    `gorm:"column:table_id;type:int(11);primary_key;AUTO_INCREMENT;comment:id" json:"table_id"`
	TableNo        string `gorm:"column:table_no;type:varchar(50);comment:桌位编号;NOT NULL" json:"table_no"`
	Sort           int    `gorm:"column:sort;type:int(10);default:0;comment:排序;NOT NULL" json:"sort"`
	AreaID         int    `gorm:"column:area_id;type:int(11);default:0;comment:区域id;NOT NULL" json:"area_id"`
	TypeID         int    `gorm:"column:type_id;type:int(11);comment:类型id;NOT NULL" json:"type_id"`
	Status         int    `gorm:"column:status;type:int(11);default:10;comment:桌台状态 10-未开台 30-已开台;NOT NULL" json:"status"`
	SwitchStatus   int    `gorm:"column:switch_status;type:int(11);default:1;comment:桌台开关状态 0-关 1-开" json:"switch_status"`
	AreaName       string `gorm:"column:area_name;type:varchar(50);comment:区域名称;NOT NULL" json:"area_name"`
	TypeName       string `gorm:"column:type_name;type:varchar(50);comment:类型名称;NOT NULL" json:"type_name"`
	ShopSupplierId int    `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:门店id;NOT NULL" json:"shop_supplier_id"`
	MinNum         int    `gorm:"column:min_num;type:int(10);default:0;comment:最小人数;NOT NULL" json:"min_num"`
	MaxNum         int    `gorm:"column:max_num;type:int(10);comment:最大人数;NOT NULL" json:"max_num"`
	AppId          uint   `gorm:"column:app_id;type:int(11) unsigned;default:0;comment:应用id;NOT NULL" json:"app_id"`
	BindInfo       string `gorm:"column:bind_info;type:varchar(255);comment:绑定信息;NOT NULL" json:"bind_info"`
	QrcodeValue    string `gorm:"column:qrcode_value;type:varchar(50);comment:二维码值" json:"qrcode_value"`
	IsBind         int    `gorm:"column:is_bind;type:int(11);default:0;comment:平板绑定状态 0-否 1-是;NOT NULL" json:"is_bind"`
	CreateTime     int64  `gorm:"column:create_time;type:int(11) unsigned;default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     int64  `gorm:"column:update_time;type:int(11) unsigned;default:0;comment:更新时间;NOT NULL" json:"update_time"`
}

type TableRepository interface {
	GetTableList() ([]*Table, error)
	ConvertTable() error
}

func NewTableService(db *gorm.DB, targetDB *gorm.DB) TableRepository {
	return &TableService{
		db:       db,
		targetDB: targetDB,
	}
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
			BaseModel: model.BaseModel{
				Uuid:       uint64(table.TableID),
				CreateTime: table.CreateTime,
				UpdateTime: table.UpdateTime,
			},
			DeskNo:      table.TableNo,
			RegionUuid:  uint64(table.AreaID),
			TypeUuid:    uint64(table.TypeID),
			Sort:        uint(table.Sort),
			Status:      uint(status),
			IsDisable:   isDisable,
			QrcodeToken: table.QrcodeValue,
		}
		fmt.Println(fmt.Sprintf("desk: %+v", desk))

		_, err := repository.NewDeskRepo(s.targetDB).CreateDesk(desk)
		if err != nil {
			return err
		}
	}
	return nil
}
