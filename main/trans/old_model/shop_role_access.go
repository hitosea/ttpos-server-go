package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopRoleAccess struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;comment:主键id"`
	RoleID     uint64 `gorm:"not null;default:0;comment:角色id"`
	AccessID   uint64 `gorm:"not null;default:0;comment:权限id"`
	AppID      uint   `gorm:"not null;default:0;comment:小程序id"`
	CreateTime int64  `gorm:"not null;default:0;comment:创建时间"`
}

type ShopRoleAccessRepository interface {
	GetShopRoleAccessList() ([]*ShopRoleAccess, error)
	ConvertShopRoleAccess() error
}

type ShopRoleAccessService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopRoleAccessService) GetShopRoleAccessList() ([]*ShopRoleAccess, error) {
	var shopRoleAccesses []*ShopRoleAccess
	if err := s.db.Find(&shopRoleAccesses).Error; err != nil {
		return nil, err
	}
	return shopRoleAccesses, nil
}

func (s *ShopRoleAccessService) ConvertShopRoleAccess() error {
	shopRoleAccesses, err := s.GetShopRoleAccessList()
	if err != nil {
		return err
	}
	for _, shopRoleAccess := range shopRoleAccesses {
		fmt.Println(fmt.Sprintf("shopRoleAccess: %+v", shopRoleAccess))
		roleAccess := model.RoleAccess{
			Uuid:       shopRoleAccess.ID,
			RoleUuid:   shopRoleAccess.RoleID,
			AccessUuid: shopRoleAccess.AccessID,
			CreateTime: shopRoleAccess.CreateTime,
		}
		_, err = repository.NewRoleAccessRepo(s.targetDB).CreateRoleAccess(roleAccess)
		if err != nil {
			return err
		}
	}
	return nil
}
