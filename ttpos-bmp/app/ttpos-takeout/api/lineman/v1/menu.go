package v1

import "github.com/gogf/gf/v2/frame/g"

// ========== 菜单同步通知 API ==========

// MenuSyncNotificationReq 菜单同步通知请求
// LINE MAN 通知 TTPOS 菜单同步任务结果
type MenuSyncNotificationReq struct {
	g.Meta            `path:"/partners/:partnerId/stores/:storeId/menus/notification" method:"post" tags:"LINE MAN Menu" summary:"接收菜单同步通知"`
	PartnerId         string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID（路径参数）"`
	StoreId           string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID（路径参数）"`
	MenuSyncRequestId string `json:"menuSyncRequestId" v:"required#菜单同步请求ID不能为空" dc:"菜单同步请求唯一 ID"`
	UpdatedAt         string `json:"updatedAt" v:"required#更新时间不能为空" dc:"同步任务更新时间，ISO 8601 格式"`
	Status            string `json:"status" v:"required|in:SUCCESS,FAILED#状态不能为空|状态必须为SUCCESS或FAILED" dc:"同步状态：SUCCESS 或 FAILED"`
	ErrorMsg          string `json:"error" dc:"同步失败时的错误信息（SUCCESS 时为空）"`
}

// MenuSyncNotificationRes 菜单同步通知响应
// TTPOS 返回给 LINE MAN 的菜单同步通知结果
type MenuSyncNotificationRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}

// ========== 触发菜单同步 API ==========

// TriggerSyncMenuReq 触发菜单同步请求
// LINE MAN 请求 TTPOS 主动同步菜单数据
type TriggerSyncMenuReq struct {
	g.Meta    `path:"/partners/:partnerId/stores/:storeId/menus/trigger-sync" method:"post" tags:"LINE MAN Menu" summary:"接收菜单同步触发请求"`
	PartnerId string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID（路径参数）"`
	StoreId   string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID（路径参数）"`
}

// TriggerSyncMenuRes 触发菜单同步响应
// TTPOS 返回给 LINE MAN 的菜单同步触发结果
type TriggerSyncMenuRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}
