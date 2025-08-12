package erpnext

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const docTypeApiUrl = "/api/v2/doctype"

var Doctype = new(sDoctype)

type sDoctype struct {
}

func init() {
	service.RegisterDoctype(Doctype)
}

func (s *sDoctype) Meta(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, fmt.Sprintf("%s/%s/meta", docTypeApiUrl, req.DocType))
	err = detectError(rst)
	return
}

func (s *sDoctype) Count(ctx context.Context, req *dto.ErpReq, params *dto.RequestParams) (int, error) {
	rst := GetClient(ctx).GetVar(ctx, fmt.Sprintf("%s/%s/count", docTypeApiUrl, req.DocType), params)
	err := detectError(rst)
	if err == nil {
		return gconv.Int(rst.Map()["data"]), nil
	}
	return 0, gerror.Newf("获取文档数量失败, docType: %v", req)
}
