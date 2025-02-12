package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type Delay struct {
	ID         uint    `gorm:"primaryKey;autoIncrement;not null"`
	Name       string  `gorm:"type:varchar(2000);default:'';comment:'名称'"`
	DelayTime  uint    `gorm:"type:int;default:0;comment:'加钟时间（分）'"`
	Status     uint    `gorm:"type:int;default:1;comment:'状态'"`
	Price      float64 `gorm:"type:decimal(12,2);default:0.00;comment:'价格'"`
	AppID      int64   `gorm:"type:int;default:0;comment:'应用id'"`
	CreateTime int64   `gorm:"type:int;not null;default:0;comment:'创建时间'"`
	UpdateTime int64   `gorm:"type:int;not null;default:0;comment:'更新时间'"`
}

type BuffetDelayRepository interface {
	GetBuffetDelayList() ([]*Delay, error)
	ConvertBuffetDelay() error
}

type BuffetDelayService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *BuffetDelayService) GetBuffetDelayList() ([]*Delay, error) {
	var buffetDelays []*Delay
	err := s.db.Find(&buffetDelays).Error
	return buffetDelays, err
}

func (s *BuffetDelayService) ConvertBuffetDelay() error {
	buffetDelays, err := s.GetBuffetDelayList()
	if err != nil {
		return err
	}
	for _, buffetDelay := range buffetDelays {
		fmt.Println(fmt.Sprintf("buffetDelay: %+v", buffetDelay))
		buffetDelay := model.BuffetDelay{
			Uuid:       uint64(buffetDelay.ID),
			Name:       buffetDelay.Name,
			DelayTime:  buffetDelay.DelayTime,
			Price:      buffetDelay.Price,
			Status:     buffetDelay.Status,
			CreateTime: buffetDelay.CreateTime,
			UpdateTime: buffetDelay.UpdateTime,
		}
		_, err := base.NewBuffetDelayRepo(s.targetDB).CreateBuffetDelay(buffetDelay)
		if err != nil {
			return err
		}
	}
	return nil
}
