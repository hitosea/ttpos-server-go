package erp

import (
	"errors"
	companyApi "ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	pkgCtx "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpCompanyClient() (companyApi.CompanyServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return companyApi.NewCompanyServiceClient(conn), conn, nil
}

// GetCompanyList 获取公司列表
func (s *erpSrv) GetCompanyList(ctx pkgCtx.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error) {
	companyResp := resp.ErpnextSiteCompanyResp{
		List: make([]resp.ErpnextSiteCompany, 0),
	}
	client, conn, err := NewErpCompanyClient()
	if err != nil {
		return companyResp, err
	}
	defer conn.Close()

	req := &companyApi.GetCompanyListReq{
		CompanyName:   erpnextSiteCompanyReq.CompanyName,
		CompanyAbbr:   erpnextSiteCompanyReq.CompanyAbbr,
		ParentCompany: erpnextSiteCompanyReq.ParentCompany,
	}
	result, err := client.GetCompanyList(WithSiteCode(ctx.GetContext(), erpnextSiteCompanyReq.SiteCode), req)
	if err != nil {
		return companyResp, err
	}
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetCompanyList", zap.String("code", result.GetCode()), zap.String("result_message", result.GetMessage()))
		return companyResp, errors.New("获取公司列表失败")
	}
	// 反序列化响应数据
	response := &companyApi.GetCompanyListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.ErpnextSiteCompanyResp{}, err
	}

	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB))
	erpnextCompanyAbbrUuidMap, err := companySettingRepo.GetErpnextCompanyAbbrUuidMap(companySettingRepo.WhereErpnextCompanyAbbrNotEmpty(), companySettingRepo.WhereSiteCode(erpnextSiteCompanyReq.SiteCode))
	if err != nil {
		return companyResp, errors.New("获取公司列表失败")
	}

	erpCompanyNameAbbrMap := make(map[string]string)
	for _, company := range response.CompanyList {
		erpCompanyNameAbbrMap[company.CompanyName] = company.CompanyAbbr
	}

	var companyList []resp.ErpnextSiteCompany
	for _, company := range response.CompanyList {
		companyList = append(companyList, resp.ErpnextSiteCompany{
			CompanyName:   company.CompanyName,
			CompanyAbbr:   company.CompanyAbbr,
			ParentCompany: company.ParentCompany,
			IsUsed: func() bool {
				_, ok := erpnextCompanyAbbrUuidMap[company.CompanyAbbr]
				return ok
			}(),
			Children: []resp.ErpnextSiteCompany{},

			ParentCompanyUuid: erpnextCompanyAbbrUuidMap[erpCompanyNameAbbrMap[company.ParentCompany]],
			CompanyUuid:       erpnextCompanyAbbrUuidMap[company.CompanyAbbr],
		})
	}

	// 将companyList 转换成树形结构，parent_company 为空的是根节点
	treeList := buildCompanyTree(companyList)

	companyResp.List = treeList
	return companyResp, nil
}

// buildCompanyTree 将公司列表转换为树形结构
// parent_company 为空或者不存在对应的父公司时，作为根节点
func buildCompanyTree(companies []resp.ErpnextSiteCompany) []resp.ErpnextSiteCompany {
	if len(companies) == 0 {
		return []resp.ErpnextSiteCompany{}
	}

	// 创建一个新的公司切片，避免修改原始数据
	companySlice := make([]resp.ErpnextSiteCompany, len(companies))

	for i := range companies {
		// 复制公司数据
		companySlice[i] = companies[i]
		// 初始化子公司列表
		companySlice[i].Children = []resp.ErpnextSiteCompany{}
	}

	// 使用递归方法构建树形结构，确保多级嵌套正确处理
	return buildTreeRecursively(companySlice, "")
}

// buildTreeRecursively 递归构建树形结构
func buildTreeRecursively(companies []resp.ErpnextSiteCompany, parentName string) []resp.ErpnextSiteCompany {
	var result []resp.ErpnextSiteCompany

	for i := range companies {
		if companies[i].ParentCompany == parentName {
			// 复制当前公司
			company := companies[i]
			// 递归查找当前公司的子公司
			children := buildTreeRecursively(companies, company.CompanyName)
			if children == nil {
				company.Children = []resp.ErpnextSiteCompany{}
			} else {
				company.Children = children
			}
			result = append(result, company)
		}
	}

	return result
}

// BuildCompanyUuidMap 递归遍历公司树，收集所有IsUsed=true的节点的UUID和父级路径映射关系
// 返回 map[uuid][]parent_company_uuids，其中父级路径从根节点到直接父节点（例如：C -> B -> A，则A的父级树是[C, B]）
func (s *erpSrv) BuildCompanyUuidMap(companies []resp.ErpnextSiteCompany) map[uint64][]uint64 {
	result := make(map[uint64][]uint64)

	// 递归遍历函数，parentPath 是从根节点到当前节点的父节点的路径
	var traverse func([]resp.ErpnextSiteCompany, []uint64)
	traverse = func(companyList []resp.ErpnextSiteCompany, parentPath []uint64) {
		for _, company := range companyList {
			// 如果该公司已被使用，则添加到映射中
			if company.IsUsed && company.CompanyUuid != 0 {
				// 复制父级路径，避免引用问题
				path := make([]uint64, len(parentPath))
				copy(path, parentPath)
				result[company.CompanyUuid] = path
			}

			// 递归处理子公司，将当前公司UUID添加到路径中
			if len(company.Children) > 0 {
				// 构建新的父级路径
				newPath := make([]uint64, len(parentPath))
				copy(newPath, parentPath)
				// 只有当前节点有UUID时才添加到路径中
				if company.CompanyUuid != 0 {
					newPath = append(newPath, company.CompanyUuid)
				}
				traverse(company.Children, newPath)
			}
		}
	}

	// 开始遍历，初始父级路径为空
	traverse(companies, []uint64{})

	return result
}
