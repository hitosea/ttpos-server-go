package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// SkootarStatusReq 定义 Skootar 状态回调的 RESTful 请求结构体
type SkootarStatusReq struct {
	g.Meta         `path:"/callback/skootar/status" tags:"Callback" method:"post" summary:"Skootar status callback"`
	JobId          string    `json:"job_id"`                     // Job id that you tracking
	StatusBefore   int       `json:"status_before"`              // Job status code is 0 - 10
	StatusAfter    int       `json:"status_after"`               // Job Status code is 0 - 10
	StatusDatetime time.Time `json:"status_datetime"`            // Date time of status change (yyyy-MM-dd HH:mm:ss)
	CallbackDesc   string    `json:"callback_desc"`              // Define callback type: status_change, price_change, arrived_point, finish_point, change_driver
	Seq            int       `json:"seq,omitempty"`              // When "callback_desc" is "arrived_point" or "finish_point", the point sequence
	BeforeNetPrice float64   `json:"before_net_price,omitempty"` // When "callback_desc" is "price_change", price before change
	AfterNetPrice  float64   `json:"after_net_price,omitempty"`  // When "callback_desc" is "price_change", price after change
}

// SkootarStatusRes 定义 Skootar 状态回调的 RESTful 响应结构体
type SkootarStatusRes struct {
	g.Meta  `mime:"text/json"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
