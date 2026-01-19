package ttpos

import "github.com/grab/grabfood-api-sdk-go"

type GetMenuExportResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Platform string                      `json:"platform"`
		MenuData grabfood.GetMenuNewResponse `json:"menuData"`
	} `json:"data"`
}
