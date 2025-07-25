package model

import "github.com/gogf/gf/v2/frame/g"

var (
	ContextKey = "ttpos-takeout"
)

type Context struct {
	ShopId   string
	OrderId  string
	Provider string //外送供应商
	Data     g.Map
}
