package setup

import (
	"context"
	"fmt"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Setup          = new(sSetup)
	companyService = service.Company()
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

	company, err := companyService.GetCompanyWithAbbr(ctx, req.CompanyAbbr)
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

func (s *sSetup) CreateUser(ctx context.Context, req *erp.CreateUserInp) (userEmail string, err error) {
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

// CreateWarehouse 创建仓库
// 参数：ctx 上下文，req 包含 shop_name、company_abbr
// 返回：仓库名称，错误信息
func (s *sSetup) CreateWarehouse(ctx context.Context, req *erp.CreateWarehouseInp) (warehouseName string, err error) {
	// 校验参数

	// 获取公司信息
	company, err := companyService.GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrapf(err, "获取公司信息失败")
	}

	// 仓库名称规则：[分支名]-[仓库类型]-[仓库名称]
	//分支名：取简称，若是英文，则取前三个字母大写，若是中文，则取前三个中文汉字的拼音字母
	//warehouseName = gstr.SubStrRune(req.Branch, 0, 3) + "-" + req.WhType + "-" + req.AliasName
	warehouseName = strings.Join([]string{req.Branch, req.WhType, req.AliasName}, "-")

	// 创建仓库 Warehouse
	warehousePayload := g.Map{
		"warehouse_name":   warehouseName,
		"custom_branch":    req.Branch,
		"custom_aliasname": req.AliasName,
		"company":          company.CompanyName,
	}
	if req.WhType == "Transit" {
		warehousePayload["warehouse_type"] = "Transit"
	}
	if _, err := service.Document().Create(ctx, "Warehouse", warehousePayload); err != nil {
		g.Log().Error(ctx, "创建仓库失败", err)
		return "", gerror.Wrapf(err, "创建仓库失败")
	}

	return warehouseName, nil
}

// CreatePosProfile CreatePosFile 创建 默认 pos profile  配置默认 posprofile
func (s *sSetup) CreatePosProfile(ctx context.Context, req *erp.CreatePosProfileInp) (posFileId string, err error) {

	// 获取公司信息
	company, err := companyService.GetCompanyWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		g.Log().Error(ctx, "获取公司信息失败", err)
		return "", gerror.Wrapf(err, "获取公司信息失败")
	}

	posProfilePayload := g.Map{
		"pos_profile_name": req.PosProfileName,
		"company":          company.CompanyName,
		"warehouse":        req.Warehouse,
		"branch":           req.Branch,
		"currency":         req.Currency,
	}
	if _, err := service.Document().Create(ctx, "Pos Profile", posProfilePayload); err != nil {
		return "", gerror.Wrapf(err, "创建pos profile失败")
	}
	return
}

// InitShop 初始化店铺
// 参数：ctx 上下文，req 包含 shop_name、company_abbr、shop_uuid
// 返回：是否成功，错误信息
func (s *sSetup) InitShop(ctx context.Context, req *setup.InitShopReq) (branchName string, err error) {

	//创建分支
	branchName, err = s.CreateBranch(ctx, req)
	if err != nil {
		return "", gerror.Wrapf(err, "创建分店失败")
	}

	//创建默认用户
	_, err = s.CreateUser(ctx, &erp.CreateUserInp{
		UserEmail: fmt.Sprintf("%s@ttpos-user.com", req.AdminUuid),
		FirstName: req.ShopName,
	})
	if err != nil {
		return "", gerror.Wrapf(err, "创建用户失败")
	}
	//创建默认仓库
	_, err = s.CreateWarehouse(ctx, &erp.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Normal",
		AliasName:   "Default",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return "", gerror.New("创建默认仓库失败")
	}
	//创建在途仓
	_, err = s.CreateWarehouse(ctx, &erp.CreateWarehouseInp{
		Branch:      branchName,
		WhType:      "Transit",
		AliasName:   "Transit",
		CompanyAbbr: req.CompanyAbbr,
	})
	if err != nil {
		return "", gerror.Wrapf(err, "创建用户失败")
	}
	//创建默认的 Cash  Balance 账号关联
	err = service.Selling().CreateDefaultModePaymentAccount(ctx, &erp.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: string(consts.ModeOfPaymentCash),
	})

	err = service.Selling().CreateDefaultModePaymentAccount(ctx, &erp.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: string(consts.ModeOfPaymentBalance),
	})

	// 仓库/pos profile 先不建
	return branchName, nil
}
