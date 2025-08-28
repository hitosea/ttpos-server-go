package v1

import "github.com/gogf/gf/v2/frame/g"

type DocChangeReq struct {
	g.Meta   `path:"/callback/{siteCode}/doctype/change" method:"post" `
	Event    string `json:"event" dc:"事件类型"`
	Data     any    `json:"data,omitempty" dc:"数据"`
	DocType  string `json:"docType" dc:"文档类型"`
	DocName  string `json:"docName,omitempty" dc:"文档名称"`
	SiteCode string `json:"siteCode" dc:"站点编码"`
}

type DocChangeRes struct {
	g.Meta `mime:"application/json" example:"{code:0}"`
	Code   string `json:"code" dc:"状态码：0成功，其他异常"`
	Msg    string `json:"msg,omitempty" dc:"状态信息"`
	Data   any    `json:"data,omitempty" dc:"数据"`
}
