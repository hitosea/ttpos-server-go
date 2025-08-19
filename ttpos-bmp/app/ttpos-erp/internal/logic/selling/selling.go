package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Selling = new(sSelling)
)

// sSelling 结构体定义
type sSelling struct{}

func init() {
	service.RegisterSelling(Selling)
}

// GetPosProfileList 查询Pos Profile列表
// 参数：ctx 上下文，req 查询请求
// 返回：erp.ResponseInfo，错误信息
func (s *sSelling) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error) {
	// 构建过滤条件
	var filters = make([][]string, 0)
	if len(req.Name) > 0 {
		filters = append(filters, []string{"name", "like", req.Name})
	}
	if len(req.Company) > 0 {
		filters = append(filters, []string{"company", "like", req.Company})
	}
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "根据公司缩写查询公司失败", err)
			return nil, gerror.Wrapf(err, "根据公司缩写查询公司失败")
		}
		filters = append(filters, []string{"company", "like", company.CompanyName})
	}

	// 查询Pos Profile列表
	list, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "POS Profile",
	}, &erp.RequestParams{
		Fields:  []string{"name", "company", "warehouse", "branch"},
		Filters: filters,
	})
	if err != nil {
		g.Log().Error(ctx, "查询Pos Profile失败", err)
		return nil, gerror.Wrapf(err, "查询Pos Profile失败")
	}
	if j, err := gjson.DecodeToJson(list.Bytes()); err == nil {
		// 遍历j.Get("data") 返回的数组字段，设置到 DataList 中
		dataList := make([]*selling.PosProfile, 0)
		dataArray := j.GetJsons("data")
		for _, item := range dataArray {
			dataInfo := &selling.PosProfile{
				Name:      item.Get("name").String(),
				Company:   item.Get("company").String(),
				Branch:    item.Get("branch").String(),
				Warehouse: item.Get("warehouse").String(),
			}
			dataList = append(dataList, dataInfo)
		}
		res = &selling.PosProfileListResp{
			ProfileList: dataList,
		}

	}
	return
}

// 过时的
func (s *sSelling) CreateDefaultModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) (err error) {
	_, err = service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: "init_shop_cash_account",
	}, req)
	if err != nil {
		return gerror.Wrapf(err, "创建默认支付账户失败")
	}
	return nil
}

func (s *sSelling) CreateModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) (err error) {
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "Mode of Payment",
		Name:    req.PaymentType,
	}, nil)
	if err != nil {
		return gerror.Wrapf(err, "获取支付方式失败")
	}
	j, err := gjson.DecodeToJson(resp)
	if err != nil {
		return gerror.Wrapf(err, "解析支付方式失败")
	}
	modePayment := &erp.ModeOfPayment{}
	j.Scan(modePayment)

	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}
	payAccounts := make([]erp.ModeOfPaymentAccount, 0)
	payAccounts = append(payAccounts, modePayment.Accounts...)
	payAccounts = append(payAccounts, erp.ModeOfPaymentAccount{
		Company:        companyName,
		DefaultAccount: "Cash - " + req.CompanyAbbr,
	})
	modePayment.Accounts = payAccounts

	if err != nil {
		return gerror.Wrapf(err, "创建默认支付账户失败")
	}
	return nil
}

// CreatePosProfile CreatePosFile 创建 默认 pos profile  配置默认 posprofile
func (s *sSelling) CreatePosProfile(ctx context.Context, req *setup.CreatePosProfileInp) (*erp.POSProfile, error) {

	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	profile := &erp.POSProfile{
		Name:               req.PosProfileName,
		Company:            companyName,
		Warehouse:          req.Warehouse,
		Branch:             req.Branch,
		Currency:           req.Currency,
		WriteOffAccount:    req.WriteOffAccount,
		WriteOffLimit:      req.WriteOffLimit,
		WriteOffCostCenter: req.WriteOffCostCenter,
	}
	//处理支付方式
	payments := make([]erp.POSPaymentMethod, 0)
	for _, payment := range req.Payments {
		paymentInfo := erp.POSPaymentMethod{
			ModeOfPayment:  payment,
			AllowInReturns: 1,
		}
		//默认现金
		if payment == "Cash" {
			paymentInfo.Default = 1
		}
		payments = append(payments, paymentInfo)
	}
	profile.Payments = payments
	// 创建pos profile
	resp, err := service.Document().Create(ctx, "POS Profile", profile)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建pos profile失败")
	}
	posProfile := &erp.POSProfile{}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析pos profile失败")
	}
	j.Get("data").Scan(posProfile)
	return posProfile, nil
}
