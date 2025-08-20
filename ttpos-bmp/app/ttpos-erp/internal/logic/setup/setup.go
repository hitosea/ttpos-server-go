package setup

import (
	"context"
	"fmt"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	setup2 "ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Setup = new(sSetup)
)

type sSetup struct {
}

func init() {
	service.RegisterSetup(Setup)
}

// CreateBranch 创建分店
// 参数：店铺名称和公司缩写编码
// 返回：ERP用户名和创建结果
func (s *sSetup) CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, err error) {
	// 参数验证
	if strings.TrimSpace(req.ShopUuid) == "" {
		return "", gerror.New("店铺UUID不能为空")
	}

	if strings.TrimSpace(req.ShopName) == "" {
		return "", gerror.New("店铺名称不能为空")
	}

	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return "", gerror.New("公司缩写编码不能为空")
	}

	company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrap(err, "获取公司信息失败")
	}
	//创建分支 Branch
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

// CreateUser 创建网站用户
func (s *sSetup) CreateUser(ctx context.Context, req *setup2.CreateUserInp) (userEmail string, err error) {
	userPayload := g.Map{
		"email":              req.UserEmail,
		"first_name":         req.FirstName,
		"user_type":          "Website User",
		"enabled":            1,
		"send_welcome_email": false,
	}
	if _, err := service.Document().Create(ctx, "User", userPayload); err != nil {
		g.Log().Error(ctx, "创建网站用户失败", err)
		return userEmail, gerror.Wrapf(err, "创建网站用户失败")
	}
	//限制用户数据权限 如果用户需要登录erp系统就约束

	return
}

// CreatePosProfile CreatePosFile 创建 默认 pos profile  配置默认
func (s *sSetup) CreatePosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posFileId string, err error) {
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrap(err, "获取公司信息失败")
	}

	//获取默认仓库
	warehouse, err := service.Warehouse().GetDefaultWarehouse(ctx, companyName, req.Branch)
	if err != nil {
		return "", gerror.Wrapf(err, "获取默认仓库失败")
	}

	posProfile, err := service.Selling().CreatePosProfile(ctx, &setup2.CreatePosProfileInp{
		PosProfileName:     req.Name,
		CompanyAbbr:        req.CompanyAbbr,
		Warehouse:          warehouse.Name,
		Branch:             req.Branch,
		Payments:           g.ArrayStr{"Cash", "Balance"},
		Currency:           "THB", //泰铢
		WriteOffAccount:    "Sales - " + req.CompanyAbbr,
		WriteOffLimit:      1.00,
		WriteOffCostCenter: "Main - " + req.CompanyAbbr,
	})
	if err != nil {
		return "", gerror.Wrapf(err, "创建pos profile失败")
	}
	return posProfile.Name, nil
}

// InitShop 初始化店铺
// 参数：ctx 上下文，req 包含 shop_name、company_abbr、shop_uuid
// 返回：是否成功，错误信息
func (s *sSetup) InitShop(ctx context.Context, req *setup.InitShopReq) (resp *setup.InitShopResp, err error) {
	var (
		branchName string
		adminEmail string
	)
	//创建分支
	branchName, err = s.CreateBranch(ctx, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建分店失败")
	}

	//创建默认用户
	adminEmail, err = s.CreateUser(ctx, &setup2.CreateUserInp{
		UserEmail: fmt.Sprintf("%s@ttpos-user.com", req.AdminUuid),
		FirstName: req.ShopName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建用户失败")
	}
	//创建默认仓库
	_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Normal",
		AliasName:   "Default",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return nil, gerror.New("创建默认仓库失败")
	}
	//创建在途仓
	_, err = service.Warehouse().CreateWarehouse(ctx, &setup2.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Transit",
		AliasName:   "Transit",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建用户失败")
	}
	//创建默认的 Cash  Balance 账号关联
	err = service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: string(consts.ModeOfPaymentCash),
	})

	err = service.Selling().CreateModePaymentAccount(ctx, &setup2.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: string(consts.ModeOfPaymentBalance),
	})

	// 仓库/pos profile 先不建
	return &setup.InitShopResp{
		BranchName: branchName,
		AdminEmail: adminEmail,
	}, nil
}
