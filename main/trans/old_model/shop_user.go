package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopUser struct {
	ShopUserID       uint64 `gorm:"primaryKey;autoIncrement;comment:主键id"`
	UserName         string `gorm:"default:'';comment:用户名"`
	Password         string `gorm:"default:'';comment:登录密码"`
	Phone            string `gorm:"default:'';comment:手机号"`
	PasswordChange   uint   `gorm:"default:0;comment:修改密码次数"`
	RealName         string `gorm:"default:'';comment:姓名"`
	IsSuper          uint   `gorm:"default:0;comment:是否为超级管理员0不是,1是"`
	ShopSupplierID   uint   `gorm:"default:0;comment:总店id"`
	IsDelete         int64  `gorm:"default:0;comment:0=显示1=伪删除"`
	UserType         uint   `gorm:"default:0;comment:账号类型0总台1门店"`
	IsStatus         uint   `gorm:"default:0;comment:是否禁用1禁用，0未禁用"`
	AppID            uint64 `gorm:"default:0;comment:程序id"`
	BindKey          string `gorm:"default:'';comment:绑定的设备key"`
	CashierOnline    uint   `gorm:"default:0;comment:收银员当班 0-不在线 1-在线"`
	CashierLoginTime uint   `gorm:"default:0;comment:收银员当班登录时间"`
	DutyNo           string `gorm:"default:'';comment:当班编号"`
	CreateTime       int    `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime       int    `gorm:"autoUpdateTime;comment:更新时间"`
}

type ShopUserRepository interface {
	GetShopUserList() ([]*ShopUser, error)
	ConvertShopUser() error
}

type ShopUserService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopUserService) GetShopUserList() ([]*ShopUser, error) {
	var shopUsers []*ShopUser
	err := s.db.Find(&shopUsers).Error
	return shopUsers, err
}

func (s *ShopUserService) ConvertShopUser() error {
	shopUsers, err := s.GetShopUserList()
	if err != nil {
		return err
	}
	for _, shopUser := range shopUsers {
		fmt.Println(fmt.Sprintf("shopUser: %+v", shopUser))
		staff := model.Staff{
			Uuid:             shopUser.ShopUserID,
			CompanyUuid:      shopUser.AppID,
			Username:         shopUser.UserName,
			Password:         shopUser.Password,
			Phone:            shopUser.Phone,
			PasswordChange:   int(shopUser.PasswordChange),
			RealName:         shopUser.RealName,
			IsSuper:          int(shopUser.IsSuper),
			UserType:         int(shopUser.UserType),
			IsDisable:        int(shopUser.IsStatus),
			BindKey:          shopUser.BindKey,
			CashierOnline:    int(shopUser.CashierOnline),
			CashierLoginTime: int(shopUser.CashierLoginTime),
			DutyNo:           shopUser.DutyNo,
			CreateTime:       shopUser.CreateTime,
			UpdateTime:       shopUser.UpdateTime,
			DeleteTime:       int(shopUser.IsDelete),
		}
		err := repository.NewStaffRepo(s.targetDB).CreateStaff(staff)
		if err != nil {
			return err
		}
	}
	return nil
}
