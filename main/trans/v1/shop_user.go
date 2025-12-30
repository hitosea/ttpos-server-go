package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
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
	CashierLoginTime int64  `gorm:"default:0;comment:收银员当班登录时间"`
	DutyNo           string `gorm:"default:'';comment:当班编号"`
	CreateTime       int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime       int64  `gorm:"autoUpdateTime;comment:更新时间"`
}

type ShopUserRepository interface {
	GetShopUserList() ([]*ShopUser, error)
	ConvertShopUser() error
}

func NewShopUserService(db *gorm.DB, targetDB *gorm.DB, targetSaasDB *gorm.DB, targetCompanyUuid uint64) ShopUserRepository {
	return &ShopUserService{db: db, targetDB: targetDB, targetSaasDB: targetSaasDB, targetCompanyUuid: targetCompanyUuid}
}

type ShopUserService struct {
	db                *gorm.DB
	targetDB          *gorm.DB
	targetSaasDB      *gorm.DB
	targetCompanyUuid uint64
}

func (s *ShopUserService) GetShopUserList() ([]*ShopUser, error) {
	var shopUsers []*ShopUser
	err := s.db.Find(&shopUsers).Error
	return shopUsers, errors.WithMessage(err)
}

func (s *ShopUserService) ConvertShopUser() error {
	shopUsers, err := s.GetShopUserList()
	if err != nil {
		return errors.WithMessage(err)
	}
	targetStaffRepo := repository.NewStaffRepo(s.targetDB)
	targetCompanyStaffRepo := repository.NewCompanyStaffRepo(s.targetSaasDB)
	for _, shopUser := range shopUsers {
		if shopUser.IsDelete == 1 {
			continue
		}
		fmt.Println(fmt.Sprintf("shopUser: %+v", shopUser))

		existsCompanyStaff := targetCompanyStaffRepo.GetCompanyStaff(targetCompanyStaffRepo.WhereUsername(shopUser.UserName))
		if existsCompanyStaff.Uuid != 0 && existsCompanyStaff.CompanyUuid != s.targetCompanyUuid {
			return errors.WithMessage(errors.New("当前用户名已存在，且不是当前商家员工" + shopUser.UserName))
		}

		err := targetStaffRepo.CreateStaff(model.Staff{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shopUser.ShopUserID),
				CreateTime: shopUser.CreateTime,
				UpdateTime: shopUser.UpdateTime,
				DeleteTime: shopUser.IsDelete,
			},
			CompanyUuid: s.targetCompanyUuid,
			Username:    shopUser.UserName,
			// Password:            shopUser.Password,
			Phone:               shopUser.Phone,
			PasswordChangeCount: int(shopUser.PasswordChange),
			RealName:            shopUser.RealName,
			IsSuper:             int(shopUser.IsSuper),
			UserType:            int(shopUser.UserType),
			IsDisable:           int(shopUser.IsStatus),
			//BindKey:             shopUser.BindKey,
			//CashierOnline:       int(shopUser.CashierOnline),
			//CashierLoginTime:    shopUser.CashierLoginTime,
			//DutyNo:              shopUser.DutyNo,
		})
		if err != nil {
			return errors.WithMessage(err)
		}

		// 从商家数据库获取员工的 is_disable 状态
		targetShopDB := s.targetDB
		targetShopStaffRepo := repository.NewStaffRepo(targetShopDB)
		shopStaff, _ := targetShopStaffRepo.GetStaff(targetShopStaffRepo.WhereUuid(uint64(shopUser.ShopUserID)))
		isDisable := 0
		if shopStaff.IsDisable != 0 {
			isDisable = shopStaff.IsDisable
		}

		if existsCompanyStaff.Uuid > 0 { // 更新saas商家员工
			err = targetCompanyStaffRepo.UpdateCompanyStaff(existsCompanyStaff.Uuid, map[string]any{
				"uuid":         shopUser.ShopUserID,
				"company_uuid": s.targetCompanyUuid,
				"username":     shopUser.UserName,
				"is_super":     shopUser.IsSuper,
				"is_disable":   isDisable,
				"create_time":  shopUser.CreateTime,
				"update_time":  shopUser.UpdateTime,
			})
			if err != nil {
				return errors.WithMessage(err)
			}
		} else { // 创建saas商家员工
			repository.NewCompanyStaffRepo(s.targetSaasDB).CreateCompanyStaff(&model.CompanyStaff{
				BaseModel: model.BaseModel{
					Uuid:       uint64(shopUser.ShopUserID),
					CreateTime: shopUser.CreateTime,
					UpdateTime: shopUser.UpdateTime,
					DeleteTime: shopUser.IsDelete,
				},
				CompanyUuid: s.targetCompanyUuid,
				Username:    shopUser.UserName,
				Phone:       shopUser.Phone,
				IsSuper:     int(shopUser.IsSuper),
				IsDisable:   isDisable,
			})
		}
	}
	return nil
}
