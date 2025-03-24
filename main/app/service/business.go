package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IBusinessSrv 定义收银服务接口
type IBusinessSrv interface {
	Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error) // 获取自助餐列表
}

// businessSrv 收银服务结构体
type businessSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewBusinessSrv 创建新的收银产品类别服务
func NewBusinessSrv(dbm *database.DBManager) IBusinessSrv {
	return NewBusinessSrvImpl(dbm)
}

// NewBusinessSrvImpl 创建新的收银服务实现
func NewBusinessSrvImpl(dbm *database.DBManager) IBusinessSrv {
	return &businessSrv{dbm: dbm}
}

// Printer 打印
func (s *businessSrv) Printer(ctx context.Context, req req.BusinessDataPrinterReq) (*resp.PrinterData, error) {
	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:             2131231,
		TotalReceivedPrice:     43124,
		TotalPayPrice:          21230,
		TotalPayFeeMoney:       2110,
		TotalServiceMoney:      120,
		TotalTaxMoney:          10124,
		TotalUserDiscountMoney: 120,
		TotalDiscountMoney:     120,
		TotalFreeOrderPrice:    120,
		TotalRefundMoney:       10,
		TotalOrderNum:          1230,
		TotalPeopleNum:         120,
		TotalProductNum:        320,
		TotalTableNum:          120,
		AvgOrderPrice:          620,
		MinOrderPrice:          120,
		MaxOrderPrice:          1200,
		AllTableOrderNum:       1230,
		AllTablePeopleNum:      120,
		AllTableAvgOrderPrice:  620,
		AllTableMinOrderPrice:  120,
		AllTableMaxOrderPrice:  1200,
		AllTablePeopleAvg:      10,
		PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
			{
				Name:     "现金",
				OrderNum: 1,
				Amount:   123213,
				Code:     40,
			},
			{
				Name:     "支付宝",
				OrderNum: 1,
				Amount:   24121,
				Code:     41,
			},
			{
				Name:     "微信支付",
				OrderNum: 1,
				Amount:   123213,
				Code:     42,
			},
		},
		AbnormalData: business_data_resp.AbnormalData{},
		MemberData: business_data_resp.MemberData{
			RechargeAmount: 120,
			GiftMoney:      120,
			GiftPoints:     120,
		},
		PeakHourList: []business_data_resp.PeakHour{
			{
				TimePeriod: "12",
				OrderNum:   120,
				Amount:     120,
			},
			{
				TimePeriod: "121232",
				OrderNum:   120,
				Amount:     120,
			},
		},
		CategoryList: []business_data_resp.Category{
			{
				Name:     "12",
				SalesNum: 1,
				Prices:   323,
			},
			{
				Name:     "121232",
				SalesNum: 2,
				Prices:   23,
			},
		},
		PercentageList: []business_data_resp.Percentage{
			{
				TaxRate:        120,
				ConsumptionTax: 120,
			},
			{
				TaxRate:        110,
				ConsumptionTax: 2120,
			},
		},
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx).PrintingBusinessData(
		&businessData,
		req.StatisticsType,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}
