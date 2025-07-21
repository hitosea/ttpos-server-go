package erp

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

const docTypeApiUrl = "/api/v2/doctype"

var Doctype = new(sDoctype)

type sDoctype struct {
}

func init() {
	service.RegisterDoctype(Doctype)
}

func (s *sDoctype) Meta(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, fmt.Sprintf("%s%s/%s/meta", GetApiBase(), docTypeApiUrl, req.DocType))
	return
}

func (s *sDoctype) Count(ctx context.Context, req *dto.ErpReq) (int, error) {
	rst := g.Client().GetVar(ctx, fmt.Sprintf("%s%s/%s/count", GetApiBase(), docTypeApiUrl, req.DocType))
	if rst != nil {
		return gconv.Int(rst.Map()["data"]), nil
	}
	return 0, gerror.Newf("获取文档数量失败, docType: %v", req)
}
