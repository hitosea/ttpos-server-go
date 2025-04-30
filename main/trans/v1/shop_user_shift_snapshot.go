package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopUserShiftSnapshot struct {
	ID             uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	ShiftLogId     int    `gorm:"column:shift_log_id;type:int(11);default:0;comment:交班记录ID" json:"shift_log_id"`
	Content        string `gorm:"column:content;type:text;comment:快照json" json:"content"`
	ShopSupplierId int    `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:店铺id" json:"shop_supplier_id"`
	AppId          int    `gorm:"column:app_id;type:int(11);default:0;comment:应用id" json:"app_id"`
	CreateTime     int64  `gorm:"column:create_time;type:int(11);default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     int64  `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
}
type ShopUserShiftSnapshotRepository interface {
	GetShopUserShiftSnapshotList() ([]*ShopUserShiftSnapshot, error)
	ConvertShopUserShiftSnapshot() error
}

func NewShopUserShiftSnapshotService(db *gorm.DB, targetDB *gorm.DB) ShopUserShiftSnapshotRepository {
	return &ShopUserShiftSnapshotService{
		db:       db,
		targetDB: targetDB,
	}
}

type ShopUserShiftSnapshotService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopUserShiftSnapshotService) GetShopUserShiftSnapshotList() ([]*ShopUserShiftSnapshot, error) {
	var shopUserShiftSnapshots []*ShopUserShiftSnapshot
	err := s.db.Find(&shopUserShiftSnapshots).Error
	return shopUserShiftSnapshots, err
}

func (s *ShopUserShiftSnapshotService) ConvertShopUserShiftSnapshot() error {
	shopUserShiftSnapshots, err := s.GetShopUserShiftSnapshotList()
	if err != nil {
		return err
	}
	for _, shopUserShiftSnapshot := range shopUserShiftSnapshots {
		fmt.Println(fmt.Sprintf("shiftSnapshot: %+v", shopUserShiftSnapshot))
		_, err := repository.NewShiftLogRepo(s.targetDB).CreateSnapshot(model.StaffShiftSnapshot{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopUserShiftSnapshot.ID),
				CreateTime: shopUserShiftSnapshot.CreateTime,
				UpdateTime: shopUserShiftSnapshot.UpdateTime,
			},
			ShiftLogUuid: uint64(shopUserShiftSnapshot.ShiftLogId),
			Content:      shopUserShiftSnapshot.Content,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
