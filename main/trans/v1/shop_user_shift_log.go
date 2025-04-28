package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type ShopUserShiftLog struct {
	ID                uint    `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	ShiftUserId       int     `gorm:"column:shift_user_id;type:int(11);comment:收银员id" json:"shift_user_id"`
	ShiftNo           string  `gorm:"column:shift_no;type:varchar(64);comment:交班编号;NOT NULL" json:"shift_no"`
	Status            int     `gorm:"column:status;type:int(11);default:1;comment:状态： 0未交班，1已交班" json:"status"`
	PreviousShiftCash float64 `gorm:"column:previous_shift_cash;type:decimal(12,2);default:0.00;comment:上一班遗留备用金" json:"previous_shift_cash"`
	CurrentCashTotal  float64 `gorm:"column:current_cash_total;type:decimal(12,2);default:0.00;comment:当前钱箱现金总计" json:"current_cash_total"`
	Incomes           string  `gorm:"column:incomes;type:text;comment:收入详情" json:"incomes"`
	TotalIncome       float64 `gorm:"column:total_income;type:decimal(12,2);default:0.00;comment:总收入" json:"total_income"`
	CashTakenOut      float64 `gorm:"column:cash_taken_out;type:decimal(12,2);default:0.00;comment:本班取出现金" json:"cash_taken_out"`
	CashLeft          float64 `gorm:"column:cash_left;type:decimal(12,2);default:0.00;comment:本班遗留备用金" json:"cash_left"`
	CashIncome        float64 `gorm:"column:cash_income;type:decimal(12,2);default:0.00;comment:本班收入现金" json:"cash_income"`
	TotalBusiness     float64 `gorm:"column:total_business;type:decimal(12,2);default:0.00;comment:本班营业总额（不包含退款）" json:"total_business"`
	IsPrinted         int     `gorm:"column:is_printed;type:tinyint(1);default:0;comment:是否打印 0-未打印 1-已打印" json:"is_printed"`
	Remark            string  `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	WithdrawCash      float64 `gorm:"column:withdraw_cash;type:decimal(12,2);default:0.00;comment:中途取出现金" json:"withdraw_cash"`
	DepositCash       float64 `gorm:"column:deposit_cash;type:decimal(12,2);default:0.00;comment:中途存入现金" json:"deposit_cash"`
	ExceptionRemark   string  `gorm:"column:exception_remark;type:varchar(500);comment:异常报备" json:"exception_remark"`
	Abnormal          string  `gorm:"column:abnormal;type:text;comment:异常信息-json字符串" json:"abnormal"`
	AppId             int     `gorm:"column:app_id;type:int(11);default:0;comment:应用id" json:"app_id"`
	ShopSupplierId    int     `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:门店id" json:"shop_supplier_id"`
	ShiftStartTime    int64   `gorm:"column:shift_start_time;type:int(11);default:0;comment:当班开始时间;NOT NULL" json:"shift_start_time"`
	ShiftEndTime      int64   `gorm:"column:shift_end_time;type:int(11);default:0;comment:当班结束时间;NOT NULL" json:"shift_end_time"`
	CreateTime        int64   `gorm:"column:create_time;type:int(11);default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime        int64   `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
}
type ShopUserShiftLogRepository interface {
	GetShopUserShiftLogList() ([]*ShopUserShiftLog, error)
	ConvertShopUserShiftLog() error
}

func NewShopUserShiftLogService(db *gorm.DB, targetDB *gorm.DB) ShopUserShiftLogRepository {
	return &ShopUserShiftLogService{
		db:       db,
		targetDB: targetDB,
	}
}

type ShopUserShiftLogService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ShopUserShiftLogService) GetShopUserShiftLogList() ([]*ShopUserShiftLog, error) {
	var shopUserShiftLogs []*ShopUserShiftLog
	err := s.db.Find(&shopUserShiftLogs).Error
	return shopUserShiftLogs, err
}

func (s *ShopUserShiftLogService) ConvertShopUserShiftLog() error {
	shopUserShiftLogs, err := s.GetShopUserShiftLogList()
	if err != nil {
		return err
	}
	for _, shiftLog := range shopUserShiftLogs {

		fmt.Println(fmt.Sprintf("shiftLog: %+v", shiftLog))

		_, err := repository.NewShiftLogRepo(s.targetDB).Create(model.StaffShiftLog{
			BaseModel: model.BaseModel{
				Uuid:       uint64(shiftLog.ID),
				CreateTime: shiftLog.CreateTime,
				UpdateTime: shiftLog.UpdateTime,
			},
			StaffUuid:         uint64(shiftLog.ShiftUserId),
			ShiftNo:           shiftLog.ShiftNo,
			Status:            shiftLog.Status,
			PreviousShiftCash: shiftLog.PreviousShiftCash,
			CurrentCashTotal:  shiftLog.CurrentCashTotal,
			Incomes:           shiftLog.Incomes,
			TotalIncome:       shiftLog.TotalIncome,
			CashTakenOut:      shiftLog.CashTakenOut,
			CashLeft:          shiftLog.CashLeft,
			CashIncome:        shiftLog.CashIncome,
			TotalBusiness:     shiftLog.TotalBusiness,
			IsPrinted:         shiftLog.IsPrinted,
			Remark:            shiftLog.Remark,
			WithdrawCash:      shiftLog.WithdrawCash,
			DepositCash:       shiftLog.DepositCash,
			ExceptionRemark:   shiftLog.ExceptionRemark,
			Abnormal:          shiftLog.Abnormal,
			ShiftStartTime:    shiftLog.ShiftStartTime,
			ShiftEndTime:      shiftLog.ShiftEndTime,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
