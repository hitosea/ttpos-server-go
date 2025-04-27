package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopAccess struct {
	AccessID       uint64 `gorm:"primaryKey;column:access_id;comment:主键id"`
	Name           string `gorm:"default:'';comment:权限名称"`
	Path           string `gorm:"default:'';comment:路由地址"`
	APIPath        string `gorm:"default:'';comment:后端路由地址"`
	ParentID       uint64 `gorm:"default:0;comment:父级id"`
	Sort           uint64 `gorm:"default:100;comment:排序(数字越小越靠前)"`
	Icon           string `gorm:"default:'';comment:菜单图标"`
	RedirectName   string `gorm:"default:'';comment:重定向名称"`
	IsRoute        uint64 `gorm:"default:0;comment:是否是路由 0=不是1=是"`
	IsMenu         uint64 `gorm:"default:0;comment:是否是菜单 0不是 1是"`
	Alias          string `gorm:"default:'';comment:别名(废弃)"`
	IsShow         uint64 `gorm:"default:1;comment:是否显示1=显示0=不显示"`
	PlusCategoryID uint64 `gorm:"default:0;comment:插件分类id"`
	Remark         string `gorm:"default:'';comment:描述"`
	IsSupplier     uint   `gorm:"default:0;comment:是否门店菜单0否1是"`
	AppID          uint   `gorm:"default:0;comment:app_id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ShopAccessRepository interface {
	GetShopAccessList() ([]*ShopAccess, error)
	ConvertShopAccess() error
}

func NewShopAccessService(db *gorm.DB, targetDB *gorm.DB) ShopAccessRepository {
	return &ShopAccessService{db: db, targetDB: targetDB}
}

type ShopAccessService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopAccessService) GetShopAccessList() ([]*ShopAccess, error) {
	var shopAccesses []*ShopAccess
	err := s.db.Find(&shopAccesses).Error
	return shopAccesses, err
}

func (s *ShopAccessService) ConvertShopAccess() error {
	shopAccesses, err := s.GetShopAccessList()
	if err != nil {
		return err
	}

	for _, shopAccess := range shopAccesses {
		fmt.Println(fmt.Sprintf("shopAccess: %+v", shopAccess))
		model := &model.Access{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopAccess.AccessID),
				CreateTime: shopAccess.CreateTime,
				UpdateTime: shopAccess.UpdateTime,
			},
			Name:             shopAccess.Name,
			Path:             shopAccess.Path,
			ApiPath:          shopAccess.APIPath,
			ParentUuid:       shopAccess.ParentID,
			Sort:             int(shopAccess.Sort),
			Icon:             shopAccess.Icon,
			RedirectName:     shopAccess.RedirectName,
			IsRoute:          int(shopAccess.IsRoute),
			IsMenu:           int(shopAccess.IsMenu),
			IsShow:           int(shopAccess.IsShow),
			PlusCategoryUuid: shopAccess.PlusCategoryID,
			Remark:           shopAccess.Remark,
			IsSupplier:       int(shopAccess.IsSupplier),
		}
		err := repository.NewAccessRepo(s.targetDB).CreateAccess(model)
		if err != nil {
			return err
		}
	}

	return nil
}
