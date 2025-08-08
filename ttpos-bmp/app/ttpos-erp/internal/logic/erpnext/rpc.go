package erpnext

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

const rpcApiUrl = "/api/v2/method"

type sRpc struct {
}

var Rpc = new(sRpc)

func init() {
	service.RegisterRpc(Rpc)
}

func (s *sRpc) Execute(ctx context.Context, req *dto.ErpReq, params interface{}) (rst *g.Var, err error) {
	rst = GetClient(ctx).PostVar(ctx, fmt.Sprintf("%s%s", getRpcUrlWithName(req), req.Method), params)
	err = detectError(rst)
	return
}
