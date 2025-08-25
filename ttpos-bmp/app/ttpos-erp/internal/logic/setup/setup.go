package setup

import (
	"context"
	"fmt"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	setup2 "ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Setup = new(sSetup)
)

type sSetup struct{}

func init() {
	service.RegisterSetup(Setup)
}

// CreateBranch 创建分店
// 参数：
//   - ctx: 上下文对象
//   - req: 初始化店铺请求参数
//
// 返回：
//   - branchName: 分店名称
//   - err: 错误信息
func (s *sSetup) CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, err error) {
	// 参数验证
	if err := s.validateCreateBranchReq(req); err != nil {
		return "", err
	}

	// 获取公司信息
	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrap(err, "获取公司信息失败")
	}

	// 创建分支
	branchPayload := g.Map{
		"branch":         req.ShopName,
		"custom_company": company.CompanyName,
	}

	if _, err := service.Document().Create(ctx, "Branch", branchPayload); err != nil {
		g.Log().Error(ctx, "创建分支失败", err)
		return "", gerror.Wrapf(err, "创建分支失败")
	}

	return req.ShopName, nil
}

// validateCreateBranchReq 验证创建分支请求参数
func (s *sSetup) validateCreateBranchReq(req *setup.InitShopReq) error {
	if strings.TrimSpace(req.ShopUuid) == "" {
		return gerror.New("店铺UUID不能为空")
	}
	if strings.TrimSpace(req.ShopName) == "" {
		return gerror.New("店铺名称不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写编码不能为空")
	}
	return nil
}

// CreateUser 创建网站用户，门店收银账户
// 参数：
//   - ctx: 上下文对象
//   - req: 创建用户请求参数
//
// 返回：
//   - err: 错误信息
func (s *sSetup) CreateUser(ctx context.Context, req *setup2.CreateUserInp) error {
	userPayload := g.Map{
		"email":              req.UserEmail,
		"first_name":         req.FirstName,
		"user_type":          "Website User",
		"enabled":            1,
		"send_welcome_email": false,
		"role_profile_name":  "ShopCashier",
		"time_zone":          "Asia/Bangkok",
	}

	if _, err := service.Document().Create(ctx, "User", userPayload); err != nil {
		g.Log().Error(ctx, "创建网站用户失败", err)
		return gerror.Wrapf(err, "创建网站用户失败")
	}

	return nil
}

// CreateDefaultPosProfile 创建默认的POS配置文件
// 参数：
//   - ctx: 上下文对象
//   - req: 创建默认POS配置文件请求参数
//
// 返回：
//   - posFileId: POS配置文件名称
//   - err: 错误信息
func (s *sSetup) CreateDefaultPosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posFileId string, err error) {
	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrap(err, "获取公司信息失败")
	}

	// 获取默认仓库
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, companyName, req.Branch)
	if err != nil {
		return "", gerror.Wrapf(err, "获取默认仓库失败")
	}

	reqInfo := &setup2.CreatePosProfileInp{
		PosProfileName:     req.Name,
		CompanyAbbr:        req.CompanyAbbr,
		Warehouse:          warehouse.Name,
		Branch:             req.Branch,
		Payments:           g.ArrayStr{"Cash", "Balance"},
		Currency:           "THB", // 泰铢
		WriteOffAccount:    "Sales - " + req.CompanyAbbr,
		WriteOffLimit:      1.00,
		WriteOffCostCenter: "Main - " + req.CompanyAbbr,
	}
	if len(req.Cashiers) > 0 {
		reqInfo.ApplicableForUsers = make([]setup2.Cashier, 0)
		for _, user := range req.Cashiers {
			reqInfo.ApplicableForUsers = append(reqInfo.ApplicableForUsers, setup2.Cashier{
				User: user.User,
			})
		}
	}
	// 创建POS配置文件
	posProfile, err := service.Selling().CreatePosProfile(ctx, reqInfo)
	if err != nil {
		return "", gerror.Wrapf(err, "创建POS配置文件失败")
	}

	return posProfile.Name, nil
}

// InitShop 初始化店铺
// 参数：
//   - ctx: 上下文对象
//   - req: 初始化店铺请求参数
//
// 返回：
//   - resp: 初始化店铺响应结果
//   - err: 错误信息
func (s *sSetup) InitShop(ctx context.Context, req *setup.InitShopReq) (resp *setup.InitShopResp, err error) {
	var (
		branchName string
		adminEmail = fmt.Sprintf("%s@ttpos-user.com", req.AdminUuid)
	)

	// 创建分支
	branchName, err = s.CreateBranch(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建分店失败")
	}

	// 创建默认用户
	err = s.CreateUser(ctx, &setup2.CreateUserInp{
		UserEmail: adminEmail,
		FirstName: req.ShopName,
		AdminUuid: req.AdminUuid,
		ShopUuid:  req.ShopUuid,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建用户失败")
	}

	// 获取用户的API_KEY/API_SECRET
	apiKey, apiSecret, err := s.GetUserApiKeySecret(ctx, adminEmail)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取用户API_KEY/API_SECRET失败")
	}

	// 记录门店管理员关联关系，记录api_key/api_secret
	_, err = dao.ShopCashier.Ctx(ctx).Insert(&do.ShopCashier{
		ShopUuid:     req.ShopUuid,
		AdminUuid:    req.AdminUuid,
		CashierEmail: adminEmail,
		ApiKey:       apiKey,
		ApiSecret:    apiSecret,
		CompanyAbbr:  req.CompanyAbbr,
		Branch:       branchName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建门店管理员关联关系失败")
	}

	// 创建默认仓库
	_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Normal",
		AliasName:   "Default",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建默认仓库失败")
	}

	// 创建在途仓
	_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Transit",
		AliasName:   "Transit",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建在途仓失败")
	}

	// 创建默认的Cash和Balance账号关联
	if err := s.createPaymentAccounts(ctx, req.CompanyAbbr); err != nil {
		g.Log().Warning(ctx, "创建支付账号关联失败", err)
		// 不返回错误，因为这不是关键步骤
	}

	//创建默认pos profile
	_, err = s.CreateDefaultPosProfile(ctx, &setup.CreateDefaultPosProfileReq{
		Name:        fmt.Sprintf("%s - POS", branchName),
		CompanyAbbr: req.CompanyAbbr,
		Branch:      branchName,
		Cashiers: []*setup.PosProfileCashier{
			{
				User: adminEmail,
			},
		},
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建默认pos profile失败")
	}

	return &setup.InitShopResp{
		BranchName: branchName,
		AdminEmail: adminEmail,
	}, nil
}

// createPaymentAccounts 创建支付方式账号关联
func (s *sSetup) createPaymentAccounts(ctx context.Context, companyAbbr string) error {
	// 创建Cash支付方式账号关联
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: companyAbbr,
		PaymentType: string(consts.ModeOfPaymentCash),
	}); err != nil {
		return gerror.Wrapf(err, "创建Cash支付方式账号关联失败")
	}

	// 创建Balance支付方式账号关联
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: companyAbbr,
		PaymentType: string(consts.ModeOfPaymentBalance),
	}); err != nil {
		return gerror.Wrapf(err, "创建Balance支付方式账号关联失败")
	}

	return nil
}

// GetUserApiKeySecret 获取用户的API密钥和密钥
// 参数：
//   - ctx: 上下文对象
//   - userEmail: 用户邮箱
//
// 返回：
//   - apiKey: API密钥
//   - apiSecret: API密钥
//   - err: 错误信息
func (s *sSetup) GetUserApiKeySecret(ctx context.Context, userEmail string) (apiKey, apiSecret string, err error) {
	// 生成API密钥
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: "frappe.core.doctype.user.user.generate_keys",
	}, g.Map{
		"user": userEmail,
	})
	if err != nil {
		return "", "", gerror.Wrapf(err, "获取用户API密钥失败")
	}

	// 解析响应数据获取API密钥
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return "", "", gerror.Wrapf(err, "解析生成API密钥响应失败")
	}
	apiSecret = j.Get("data.api_secret").String()

	// 获取用户信息
	resp, err = service.Document().Get(ctx, &erp.ErpReq{
		DocType: "User",
		Name:    userEmail,
	}, nil)
	if err != nil {
		return "", "", gerror.Wrapf(err, "获取用户信息失败")
	}

	// 解析响应数据获取API密钥
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return "", "", gerror.Wrapf(err, "解析用户信息响应失败")
	}
	apiKey = j.Get("data.api_key").String()

	return apiKey, apiSecret, nil
}
