package v1

import (
	"fmt"
	"strconv"
	"strings"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type PayType struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:'支付类型表唯一标识符'"`
	Name           string `gorm:"type:varchar(500);default:'';comment:'支付方式'"`
	Remark         string `gorm:"type:varchar(500);default:'';comment:'名称'"`
	Fee            int    `gorm:"type:int;default:0;comment:'支付手续费0-100'"`
	Source         int    `gorm:"type:int;default:0;comment:'来源 0-系统默认 1-自行添加 2-LianLianPay'"`
	Value          int    `gorm:"type:int;default:null;comment:'value值'"`
	Status         int    `gorm:"type:int;default:1;comment:'状态(0启动,1启用)'"`
	Img            string `gorm:"type:text;default:'';comment:'图像'"`
	Qrcode         string `gorm:"type:text;default:'';comment:'二维码'"`
	Sort           string `gorm:"type:varchar(25);default:'0';comment:'排序'"`
	IsShowRecharge string `gorm:"type:varchar(255);default:'';comment:'充值显示 10-收银机 20-点餐助手'"`
	IsShowCheckout string `gorm:"type:varchar(255);default:'';comment:'结账显示 10-收银机 20-点餐助手'"`
	AppID          int    `gorm:"type:int;default:0;comment:'应用id'"`
	ShopSupplierID int    `gorm:"type:int;default:0;comment:'门店id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	DeleteTime     int64  `gorm:"type:int;not null;default:0;comment:'删除时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

func (p *PayType) IsShowCashier() int {
	if strings.Contains(p.IsShowCheckout, "10") {
		return 1
	}
	return 0
}

func (p *PayType) IsShowAssistant() int {
	if strings.Contains(p.IsShowCheckout, "20") {
		return 1
	}
	return 0
}

func (p *PayType) IsShowMemberRecharge() int {
	if strings.Contains(p.IsShowRecharge, "10") {
		return 1
	}
	return 0
}

func (p *PayType) GetSort() int {
	sort, err := strconv.Atoi(p.Sort)
	if err != nil {
		return 0
	}
	return sort
}

type PayTypeRepository interface {
	GetPayTypeList() ([]*PayType, error)
	ConvertPayType() error
}

func NewPayTypeService(db *gorm.DB, targetDB *gorm.DB) PayTypeRepository {
	return &PayTypeService{db: db, targetDB: targetDB}
}

type PayTypeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *PayTypeService) GetPayTypeList() ([]*PayType, error) {
	var payTypes []*PayType
	err := s.db.Find(&payTypes).Error
	return payTypes, err
}

func (s *PayTypeService) GetUploadFileId(filePath string) uint64 {
	idx := strings.LastIndex(filePath, "/")
	if idx == -1 {
		return 0
	}
	keyword := filePath[idx+1:]
	if keyword == "" {
		return 0
	}
	var uploadFile UploadFile
	s.db.Where("save_name like ?", "%"+keyword+"%").Find(&uploadFile)
	return uploadFile.FileID
}

func (s *PayTypeService) ConvertPayType() error {
	payTypes, err := s.GetPayTypeList()
	if err != nil {
		return err
	}
	for _, payType := range payTypes {
		fmt.Println(fmt.Sprintf("payType: %+v", payType))
		// 值为-1的免单支付方式不转换
		if payType.Value == -1 {
			continue
		}
		remark := payType.Remark
		if remark == "" {
			remark = payType.Name
		}
		payType := model.PaymentMethod{
			BaseModel: model.BaseModel{
				Uuid:       uint64(payType.ID),
				CreateTime: payType.CreateTime,
				UpdateTime: payType.UpdateTime,
				DeleteTime: payType.DeleteTime,
			},
			Name:                 payType.Name,
			Code:                 payType.Value,
			PaymentName:          remark,
			Source:               payType.Source,
			LogoFileUuid:         s.GetUploadFileId(payType.Img),
			QrcodeFileUuid:       s.GetUploadFileId(payType.Qrcode),
			FeePercent:           float64(payType.Fee),
			IsShowCashier:        payType.IsShowCashier(),
			IsShowAssistant:      payType.IsShowAssistant(),
			IsShowMemberRecharge: payType.IsShowMemberRecharge(),
			Status:               payType.Status,
			Sort:                 payType.GetSort(),
			DefaultImg:           payType.Img,
		}
		if err := repository.NewPaymentMethodRepo(s.targetDB).CreatePaymentMethod(payType); err != nil {
			return err
		}
	}
	return nil
}
