// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import "ttpos-bmp/app/ttpos-takeout/internal/service"

var (
	// Grab Grab服务实例
	Grab = new(sGrab)
)

type sGrab struct{}

func init() {
	service.RegisterGrab(Grab)
}
