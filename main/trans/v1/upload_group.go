package v1

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type UploadGroup struct {
	GroupID        uint64 `gorm:"primaryKey;autoIncrement;comment:'分类id'"`
	GroupType      string `gorm:"type:varchar(10);not null;default:'';comment:'文件类型'"`
	GroupName      string `gorm:"type:varchar(30);not null;default:'';comment:'分类名称'"`
	Sort           uint64 `gorm:"not null;default:0;comment:'分类排序(数字越小越靠前)'"`
	ShopSupplierID int    `gorm:"default:0;comment:'供应商id'"`
	IsDelete       uint8  `gorm:"not null;default:0;comment:'是否删除'"`
	AppID          uint64 `gorm:"not null;default:0;comment:'应用id'"`
	CreateTime     uint64 `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime     uint64 `gorm:"not null;default:0;comment:'更新时间'"`

	UploadFiles []*UploadFile `gorm:"foreignKey:GroupID;references:GroupID"`
}

type UploadGroupRepository interface {
	GetUploadGroupList() ([]*UploadGroup, error)
	ConvertUploadGroup() error
}

func NewUploadGroupService(db *gorm.DB, targetDB *gorm.DB) UploadGroupRepository {
	return &UploadGroupService{
		db:       db,
		targetDB: targetDB,
	}
}

type UploadGroupService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UploadGroupService) GetUploadGroupList() ([]*UploadGroup, error) {
	var uploadGroups []*UploadGroup
	err := s.db.Model(&UploadGroup{}).Find(&uploadGroups).Error
	return uploadGroups, err
}

func (s *UploadGroupService) ConvertUploadGroup() error {
	uploadGroups, err := s.GetUploadGroupList()
	if err != nil {
		return err
	}
	for _, uploadGroup := range uploadGroups {
		uploadFileGroup := model.FileGroup{
			BaseModel: model.BaseModel{
				Uuid:       uint64(uploadGroup.GroupID),
				CreateTime: int64(uploadGroup.CreateTime),
				UpdateTime: int64(uploadGroup.UpdateTime),
			},
			GroupType: uploadGroup.GroupType,
			GroupName: uploadGroup.GroupName,
			Sort:      uint(uploadGroup.Sort),
		}
		err := repository.NewFileRepo(s.targetDB).CreateFileGroup(uploadFileGroup)
		if err != nil {
			return err
		}
	}
	return nil
}
