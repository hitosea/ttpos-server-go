package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type UserGrade struct {
	GradeID       uint   `gorm:"primaryKey;autoIncrement;comment:等级ID"`
	Name          string `gorm:"type:varchar(50);not null;default:'';comment:等级名称"`
	OpenMoney     int8   `gorm:"type:tinyint(3);default:0;comment:是否开放0，否1是"`
	UpgradeMoney  int    `gorm:"not null;default:0;comment:升级条件"`
	OpenPoints    int8   `gorm:"type:tinyint(3);default:0;comment:积分是否开放0否1是"`
	UpgradePoints int    `gorm:"default:0;comment:累计积分升级"`
	OpenInvite    int8   `gorm:"type:tinyint(3);default:0;comment:邀请是否开放0否1是"`
	UpgradeInvite int    `gorm:"default:0;comment:邀请人数升级"`
	Equity        int    `gorm:"not null;default:100;comment:等级权益,百分比"`
	IsDefault     int8   `gorm:"type:tinyint(3);default:0;comment:是否默认，1是，0否"`
	Remark        string `gorm:"type:varchar(500);default:'';comment:备注"`
	Weight        int8   `gorm:"type:tinyint(3);default:100;comment:权重"`
	IsDelete      int8   `gorm:"type:tinyint(3);unsigned;not null;default:0;comment:是否删除"`
	AppID         uint   `gorm:"unsigned;not null;default:0;comment:程序id"`
	CreateTime    uint   `gorm:"unsigned;not null;default:0;comment:创建时间"`
	UpdateTime    uint   `gorm:"unsigned;not null;default:0;comment:更新时间"`
}

type UserGradeRepository interface {
	GetUserGradeList() ([]*Delay, error)
	ConvertUserGrade() error
}

type UserGradeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UserGradeService) GetUserGradeList() ([]*UserGrade, error) {
	var userGrades []*UserGrade
	err := s.db.Find(&userGrades).Error
	return userGrades, err
}

func (s *UserGradeService) ConvertUserGrade() error {
	userGrades, err := s.GetUserGradeList()
	if err != nil {
		return err
	}
	for _, userGrade := range userGrades {
		fmt.Println(fmt.Sprintf("userGrade: %+v", userGrade))
		memberLevel := model.MemberLevel{
			Uuid:          uint64(userGrade.GradeID),
			Name:          userGrade.Name,
			OpenMoney:     int(userGrade.OpenMoney),
			UpgradeMoney:  userGrade.UpgradeMoney,
			OpenPoints:    int(userGrade.OpenPoints),
			UpgradePoints: userGrade.UpgradePoints,
			OpenInvite:    int(userGrade.OpenInvite),
			UpgradeInvite: userGrade.UpgradeInvite,
			Discount:      userGrade.Equity,
			Priority:      int(userGrade.Weight),
			IsDefault:     int(userGrade.IsDefault),
			Remark:        userGrade.Remark,
			CreateTime:    int64(userGrade.CreateTime),
			UpdateTime:    int64(userGrade.UpdateTime),
		}
		_, err = base.NewMemberLevelRepo(s.targetDB).CreateMemberLevel(memberLevel)
		if err != nil {
			return err
		}
	}
	return nil
}
