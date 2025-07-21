package erp

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

const (
	documentApiUrl = "/api/v2/document"
)

var Document = new(sDocument)

type sDocument struct {
}

func init() {
	service.RegisterDocument(Document)
}

func (s *sDocument) List(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, getDocumentUrl(req.DocType))
	return
}

func (s *sDocument) Get(ctx context.Context, req *dto.ErpReq, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, getDocumentUrlWithName(req), params)
	return
}

func (s *sDocument) Create(ctx context.Context, docType string, data interface{}) (rst *g.Var, err error) {
	rst = g.Client().PostVar(ctx, getDocumentUrl(docType), data)
	return
}

func (s *sDocument) Update(ctx context.Context, req *dto.ErpReq, data interface{}) (rst *g.Var, err error) {
	rst = g.Client().PutVar(ctx, getDocumentUrlWithName(req), data)
	return
}

func (s *sDocument) Delete(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error) {
	rst = g.Client().DeleteVar(ctx, getDocumentUrlWithName(req))
	return
}

func (s *sDocument) Copy(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, fmt.Sprintf("%scopy", getDocumentUrlWithName(req)))
	return
}

func (s *sDocument) Execute(ctx context.Context, req *dto.ErpReq, params interface{}) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, fmt.Sprintf("%smethod/%s", getDocumentUrlWithName(req), req.Method), params)
	return
}

func getDocumentUrl(docType string) string {
	return fmt.Sprintf("%s%s/%s", GetApiBase(), documentApiUrl, docType)
}

func getDocumentUrlWithName(req *dto.ErpReq) string {
	return fmt.Sprintf("%s/%s/", getDocumentUrl(req.DocType), req.Name)
}
