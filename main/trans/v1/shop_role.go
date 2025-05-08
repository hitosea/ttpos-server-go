package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopRole struct {
	RoleID     uint64 `gorm:"primaryKey;autoIncrement;comment:角色id"`
	RoleName   string `gorm:"default:'';comment:角色名称"`
	Sort       uint   `gorm:"default:100;comment:排序(数字越小越靠前)"`
	AppID      uint   `gorm:"default:0;comment:小程序id"`
	CreateTime int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ShopRoleRepository interface {
	GetShopRoleList() ([]*ShopRole, error)
	ConvertShopRole() error
}

func NewShopRoleService(db *gorm.DB, targetDB *gorm.DB) ShopRoleRepository {
	return &ShopRoleService{db: db, targetDB: targetDB}
}

type ShopRoleService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopRoleService) GetShopRoleList() ([]*ShopRole, error) {
	var shopRoles []*ShopRole
	err := s.db.Find(&shopRoles).Error
	return shopRoles, err
}

func (s *ShopRoleService) ConvertShopRole() error {
	shopRoles, err := s.GetShopRoleList()
	if err != nil {
		return errors.WithMessage(err)
	}

	for _, shopRole := range shopRoles {
		fmt.Println(fmt.Sprintf("shopRole: %+v", shopRole))
		model := model.Role{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopRole.RoleID),
				CreateTime: shopRole.CreateTime,
				UpdateTime: shopRole.UpdateTime,
			},
			Name: shopRole.RoleName,
			Sort: int(shopRole.Sort),
		}
		_, err = repository.NewRoleRepo(s.targetDB).CreateRole(model)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
