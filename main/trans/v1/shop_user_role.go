package v1

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopUserRole struct {
	ID         uint  `gorm:"primary_key;AUTO_INCREMENT;comment:主键id"`
	ShopUserID uint  `gorm:"not null;default:0;comment:超管用户id"`
	RoleID     uint  `gorm:"not null;default:0;comment:角色id"`
	AppID      uint  `gorm:"not null;default:0;comment:小程序id"`
	CreateTime int64 `gorm:"not null;default:0;comment:创建时间"`
	UpdateTime int64 `gorm:"default:0;comment:更新时间"`
}
type ShopUserRoleRepository interface {
	GetShopUserRoleList() ([]*ShopUserRole, error)
	ConvertShopUserRole() error
}

func NewShopUserRoleService(db *gorm.DB, targetDB *gorm.DB) ShopUserRoleRepository {
	return &ShopUserRoleService{
		db:       db,
		targetDB: targetDB,
	}
}

type ShopUserRoleService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopUserRoleService) GetShopUserRoleList() ([]*ShopUserRole, error) {
	var shopUserRoles []*ShopUserRole
	err := s.db.Find(&shopUserRoles).Error
	return shopUserRoles, err
}

func (s *ShopUserRoleService) ConvertShopUserRole() error {
	shopUserRoles, err := s.GetShopUserRoleList()
	if err != nil {
		return err
	}
	for _, shopUserRole := range shopUserRoles {
		staffRole := model.StaffRole{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopUserRole.ID),
				CreateTime: shopUserRole.CreateTime,
				UpdateTime: shopUserRole.UpdateTime,
			},
			StaffUuid: int64(shopUserRole.ShopUserID),
			RoleUuid:  int64(shopUserRole.RoleID),
		}
		err = repository.NewStaffRoleRepo(s.targetDB).CreateStaffRole(staffRole)
		if err != nil {
			return err
		}
	}
	return nil
}
