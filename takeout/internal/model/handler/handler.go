package handler

import (
	"context"
	"database/sql"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/guid"
)

// AutoUUIDInsert 自动插入uuid
var AutoUUIDInsert = gdb.HookHandler{
	Insert: func(ctx context.Context, in *gdb.HookInsertInput) (result sql.Result, err error) {
		for idx, record := range in.Data {
			data := gmap.NewStrAnyMapFrom(record)
			if data.Contains("uuid") {
				data.SetIfNotExist("uuid", guid.S())
			}
			in.Data[idx] = data.Map()
		}
		result, err = in.Next(ctx)
		return
	},
}
