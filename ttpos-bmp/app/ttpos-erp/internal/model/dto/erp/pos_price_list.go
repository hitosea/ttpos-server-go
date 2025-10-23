// Package erp 定义 ERP 相关的数据传输对象
package erp

// PosPriceList POS 价格表规则
// 用于管理采购价格表和销售价格表的映射关系
type PosPriceList struct {
	Name             string `json:"name" description:"文档名称（ERPNext自动生成）"`
	RuleCode         string `json:"rule_code" description:"规则代码"`
	Company          string `json:"company" description:"公司"`
	BuyingPriceList  string `json:"buying_price_list" description:"采购价格表"`
	SellingPriceList string `json:"selling_price_list" description:"销售价格表"`
	Disabled         int    `json:"disabled" description:"是否禁用（0-启用，1-禁用）"`
	Owner            string `json:"owner,omitempty" description:"创建人"`
	Creation         string `json:"creation,omitempty" description:"创建时间"`
	Modified         string `json:"modified,omitempty" description:"修改时间"`
	ModifiedBy       string `json:"modified_by,omitempty" description:"修改人"`
	Docstatus        int    `json:"docstatus,omitempty" description:"文档状态（0-草稿，1-已提交，2-已取消）"`
	Idx              int    `json:"idx,omitempty" description:"索引"`
}

// ListPosPriceListsReq 查询 POS 价格表规则列表请求
type ListPosPriceListsReq struct {
	CompanyAbbr string `json:"company_abbr,omitempty" description:"公司缩写编码"`
	Company     string `json:"company,omitempty" description:"公司名称"`
	RuleCode    string `json:"rule_code,omitempty" description:"规则代码"`
	Disabled    int    `json:"disabled,omitempty" description:"是否禁用（-1-全部，0-启用，1-禁用）"`
	PageNo      int    `json:"page_no" v:"required|min:1#页码不能为空|页码最小为1" description:"页码"`
	PageSize    int    `json:"page_size" v:"required|min:1|max:100#每页数量不能为空|每页数量最小为1|每页数量最大为100" description:"每页数量"`
}
