package erp

import (
	"context"
	companyApi "ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"

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
func (s *erpSrv) GetCompanyList(ctx context.Context, erpnextSiteCompanyReq req.ErpnextSiteCompanyReq) (resp.ErpnextSiteCompanyResp, error) {
	var companyResp resp.ErpnextSiteCompanyResp
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
	result, err := client.GetCompanyList(WithSiteCode(ctx, erpnextSiteCompanyReq.SiteCode), req)
	if err != nil {
		return companyResp, err
	}
	// 反序列化响应数据
	response := &companyApi.GetCompanyListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return resp.ErpnextSiteCompanyResp{}, err
	}
	var companyList []resp.ErpnextSiteCompany
	for _, company := range response.CompanyList {
		companyList = append(companyList, resp.ErpnextSiteCompany{
			CompanyName:   company.CompanyName,
			CompanyAbbr:   company.CompanyAbbr,
			ParentCompany: company.ParentCompany,
			Children:      []resp.ErpnextSiteCompany{},
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

	// 创建公司名称到公司对象的映射，方便查找
	companyMap := make(map[string]*resp.ErpnextSiteCompany)
	// 创建一个新的公司切片，避免修改原始数据
	companySlice := make([]resp.ErpnextSiteCompany, len(companies))

	for i := range companies {
		// 复制公司数据
		companySlice[i] = companies[i]
		// 初始化子公司列表
		companySlice[i].Children = []resp.ErpnextSiteCompany{}
		// 建立映射关系
		companyMap[companySlice[i].CompanyName] = &companySlice[i]
	}

	// 先构建父子关系
	for i := range companySlice {
		company := &companySlice[i]

		// 如果有父公司，则添加到父公司的子公司列表中
		if company.ParentCompany != "" {
			if parentCompany, exists := companyMap[company.ParentCompany]; exists {
				// 将当前公司添加到父公司的子公司列表中
				parentCompany.Children = append(parentCompany.Children, *company)
			}
		}
	}

	// 再收集根节点（包括已经构建好子树的根节点）
	var rootCompanies []resp.ErpnextSiteCompany
	for i := range companySlice {
		company := &companySlice[i]
		// 如果没有父公司或找不到父公司，则为根节点
		if company.ParentCompany == "" || companyMap[company.ParentCompany] == nil {
			rootCompanies = append(rootCompanies, *company)
		}
	}

	return rootCompanies
}
