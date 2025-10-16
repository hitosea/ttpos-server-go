package erpnext

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
)

const docTypeApiUrl = "/api/v2/doctype"

var Doctype = new(sDoctype)

type sDoctype struct {
}

func init() {
	service.RegisterDoctype(Doctype)
}

func (s *sDoctype) Meta(ctx context.Context, req *erp.ErpReq) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).GetVar(ctx, fmt.Sprintf("%s/%s/meta", docTypeApiUrl, req.DocType))
	rst, err = detectError(resp)
	return
}

func (s *sDoctype) Count(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (int, error) {
	resp := GetClient(ctx).GetVar(ctx, fmt.Sprintf("%s/%s/count", docTypeApiUrl, req.DocType), params)
	rst, err := detectError(resp)
	if err == nil {
		return gconv.Int(rst.Get("data").Int()), nil
	}
	return 0, gerror.Newf("获取文档数量失败, docType: %v", req)
}
