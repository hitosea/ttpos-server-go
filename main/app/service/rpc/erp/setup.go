package erp

import (
	"context"
	"errors"
	"strconv"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	cc "ttpos-server-go/pkg/context"

	"google.golang.org/grpc"
)

func NewErpSetupClient() (setup.SetupServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return setup.NewSetupServiceClient(conn), conn, nil
}

func (s *erpSrv) InitShop(ctx cc.Context, initShopReq req.InitShopReq) (resp.InitShopResp, error) {
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(0))
	// 判断是否已经有店铺
	companySetting, _ := companySettingRepo.GetOne(companySettingRepo.WhereErpnextCompanyAbbr(initShopReq.CompanyAbbr))
	if companySetting.Uuid != 0 {
		return resp.InitShopResp{}, errors.New("所属erpnext公司已授权其他商家")
	}

	// 判断商家是否存在
	companyRepo := repository.NewCompanyRepo(s.dbm.GetDB(0))
	company, _ := companyRepo.GetCompanyInfoByUuid(initShopReq.CompanyUuid)
	if company.Uuid == 0 {
		return resp.InitShopResp{}, errors.New("商家不存在")
	}

	// 判断商家超管是否存在
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(initShopReq.CompanyUuid))
	staff, _ := staffRepo.GetStaff(staffRepo.WhereIsSuper(1))
	if staff.Uuid == 0 {
		return resp.InitShopResp{}, errors.New("商家超管不存在")
	}

	var initShopResp resp.InitShopResp
	client, conn, err := NewErpSetupClient()
	if err != nil {
		return initShopResp, err
	}
	defer conn.Close()
	req := &setup.InitShopReq{
		ShopName:    company.Name,
		CompanyAbbr: initShopReq.CompanyAbbr,
		ShopUuid:    strconv.FormatUint(initShopReq.CompanyUuid, 10),
		AdminUuid:   strconv.FormatUint(staff.Uuid, 10),
	}
	result, err := client.InitShop(WithSiteCode(context.Background(), initShopReq.SiteCode), req)
	if err != nil {
		return initShopResp, err
	}
	response := &setup.InitShopResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.InitShopResp{}, err
	}

	// 更新saas库
	s.dbm.GetDB(0).Model(&model.CompanySetting{}).Where("company_uuid = ?", company.Uuid).Updates(map[string]any{
		"erpnext_site_code":        initShopReq.SiteCode,
		"erpnext_company_abbr":     initShopReq.CompanyAbbr,
		"erpnext_branch_name":      response.BranchName,
		"erpnext_pos_profile_name": initShopReq.PosProfileName,
	})

	// 更新商家库
	s.dbm.GetDB(company.Uuid).Model(&model.CompanySetting{}).Where("company_uuid = ?", company.Uuid).Updates(map[string]any{
		"erpnext_site_code":        initShopReq.SiteCode,
		"erpnext_company_abbr":     initShopReq.CompanyAbbr,
		"erpnext_branch_name":      response.BranchName,
		"erpnext_pos_profile_name": initShopReq.PosProfileName,
	})

	return initShopResp, nil
}
