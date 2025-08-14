package dto

type ErpReq struct {
	DocType string `json:"docType"`
	Name    string `json:"name"`
	Method  string `json:"method"`
}

type ReportParams struct {
	ReportName           string `json:"report_name"`                           // 报表名称
	Filters              string `json:"filters"`                               //json 格式的筛选条件
	IgnorePreparedReport bool   `default:"true" json:"ignore_prepared_report"` // 是否忽略已准备好的报表
}
