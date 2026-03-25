package v1

import "github.com/gogf/gf/v2/frame/g"

type RetrySalesInvoiceReq struct {
	g.Meta        `path:"/callback/retry/sales_invoice" method:"post"`
	SiteCode      string `json:"siteCode,omitempty" dc:"站点编码（可选过滤）"`
	RecordId      int64  `json:"recordId,omitempty" dc:"记录ID（可选，指定单条重试）"`
	SaleOrderUuid string `json:"saleOrderUuid,omitempty" dc:"订单UUID（可选过滤）"`
}

type RetrySalesInvoiceRes struct {
	g.Meta     `mime:"application/json"`
	Code       string `json:"code" dc:"状态码：0成功"`
	Msg        string `json:"msg,omitempty" dc:"状态信息"`
	RetryCount int    `json:"retryCount" dc:"重新入队的记录数"`
}
