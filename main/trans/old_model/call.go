package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"

	"gorm.io/gorm"
)

type Call struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	TableID        uint   `gorm:"default:0;comment:'桌位ID'"`
	TableNo        string `gorm:"default:'';comment:'桌位号'"`
	CallType       uint   `gorm:"default:0;comment:'呼叫类型(1服务员,2收款)'"`
	Status         uint   `gorm:"default:0;comment:'状态(0未处理,1已处理)'"`
	AppID          uint   `gorm:"default:0;comment:'应用id'"`
	ShopSupplierID uint   `gorm:"default:0;comment:'门店id'"`
	IsSend         uint   `gorm:"default:0;comment:'消息发送状态 0-否 1-是'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type CallRepository interface {
	GetCallList() ([]*Call, error)
	ConvertCall() error
}

type CallService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *CallService) GetCallList() ([]*Call, error) {
	var calls []*Call
	err := s.db.Find(&calls).Error
	return calls, err
}

func (s *CallService) ConvertCall() error {
	calls, err := s.GetCallList()
	if err != nil {
		return err
	}
	for _, call := range calls {
		fmt.Println(fmt.Sprintf("call: %+v", call))
		customerCall := model.CustomerCall{
			ID:         0,
			Uuid:       call.ID,
			DeskUUID:   call.TableID,
			DeskNo:     call.TableNo,
			Status:     call.Status,
			IsSend:     call.IsSend,
			CreateTime: call.CreateTime,
			UpdateTime: call.UpdateTime,
		}
		_, err := base.NewCustomerCallRepo(s.targetDB).CreateCustomerCall(customerCall)
		if err != nil {
			return err
		}
	}
	return nil
}
