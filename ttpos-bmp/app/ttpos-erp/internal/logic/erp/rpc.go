package erp

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

const rpcApiUrl = "/api/v2/method"

type sRpc struct {
}

var Rpc = new(sRpc)

func init() {
	service.RegisterRpc(Rpc)
}

func (s *sRpc) Execute(ctx context.Context, req *dto.ErpReq, params interface{}) (rst *g.Var, err error) {
	rst = g.Client().PostVar(ctx, fmt.Sprintf("%s%s", getRpcUrlWithName(req), req.Method), params)
	return
}

func getRpcUrlWithName(req *dto.ErpReq) string {
	return fmt.Sprintf("%s%s/%s", GetApiBase(), rpcApiUrl, req.DocType)
}
