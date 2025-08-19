package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-manager/api/manager"
)

type HelloReq struct {
	g.Meta `path:"/hello" tags:"Hello" method:"get" summary:"You first hello api"`
}
type HelloRes struct {
	g.Meta `mime:"text/html" example:"string"`
}

type GetSettingReq struct {
	g.Meta `path:"/setting" tags:"Setting" method:"get" summary:"read setting"`
	Key    string
}

type GetSettingRes struct {
	g.Meta `mime:"application/json" example:"string"`
	manager.Setting
}
