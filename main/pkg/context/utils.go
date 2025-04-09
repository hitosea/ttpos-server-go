package context

import (
	"strconv"
	"ttpos-server-go/app/constant"
)

// GetH5OrderUuid 获取H5订单ID,从http请求的Header中
func GetH5OrderUuid(ctx Context) uint64 {
	if ctx.GetSource() == constant.SourceCashier {
		h5OrderUuid := ctx.GetGin().GetHeader("h5-order-uuid")
		if h5OrderUuid != "" {
			h5OrderUuidInt, err := strconv.ParseUint(h5OrderUuid, 10, 64)
			if err == nil {
				return h5OrderUuidInt
			}
		}
	}
	return 0
}
