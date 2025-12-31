// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MenuLog is the golang structure of table takeout_menu_log for DAO operations like Where/Data.
type MenuLog struct {
	g.Meta       `orm:"table:takeout_menu_log, do:true"`
	Id           any // 主键
	Uuid         any // 唯一ID
	MerchantId   any // 商户ID
	ProviderName any // 渠道: grab
	SyncType     any // 同步类型: FULL, PARTIAL, NOTIFY
	RequestId    any // 请求ID (来自Grab)
	Status       any // 同步状态: QUEUED, SUCCESS, FAIL, PROCESSING
	MenuSnapshot any // 菜单快照(JSON)
	ErrorCode    any // 错误代码
	ErrorMsg     any // 错误信息
	CreatedAt    any // 创建时间
	UpdatedAt    any // 更新时间
}
