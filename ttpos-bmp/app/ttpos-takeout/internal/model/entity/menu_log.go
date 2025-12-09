// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MenuLog is the golang structure for table menu_log.
type MenuLog struct {
	Id           int64       `json:"id"           orm:"id"            description:"主键"`                                      // 主键
	Uuid         string      `json:"uuid"         orm:"uuid"          description:"唯一ID"`                                    // 唯一ID
	MerchantId   string      `json:"merchantId"   orm:"merchant_id"   description:"商户ID"`                                    // 商户ID
	ProviderName string      `json:"providerName" orm:"provider_name" description:"渠道: grab"`                                // 渠道: grab
	SyncType     string      `json:"syncType"     orm:"sync_type"     description:"同步类型: FULL, PARTIAL, NOTIFY"`             // 同步类型: FULL, PARTIAL, NOTIFY
	RequestId    string      `json:"requestId"    orm:"request_id"    description:"请求ID (来自Grab)"`                           // 请求ID (来自Grab)
	Status       string      `json:"status"       orm:"status"        description:"同步状态: QUEUED, SUCCESS, FAIL, PROCESSING"` // 同步状态: QUEUED, SUCCESS, FAIL, PROCESSING
	MenuSnapshot string      `json:"menuSnapshot" orm:"menu_snapshot" description:"菜单快照(JSON)"`                              // 菜单快照(JSON)
	ErrorCode    string      `json:"errorCode"    orm:"error_code"    description:"错误代码"`                                    // 错误代码
	ErrorMsg     string      `json:"errorMsg"     orm:"error_msg"     description:"错误信息"`                                    // 错误信息
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`                                    // 创建时间
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`                                    // 更新时间
}
