package erpnext

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/frame/g"
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

func (s *sDocument) List(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, getDocumentUrl(ctx, req.DocType), params)
	err = detectError(rst)
	return
}

func (s *sDocument) Get(ctx context.Context, req *erp.ErpReq, params *erp.RequestParams) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, getDocumentUrlWithName(ctx, req), params)
	err = detectError(rst)
	return
}

func (s *sDocument) Create(ctx context.Context, docType string, data interface{}) (rst *g.Var, err error) {
	rst = GetClient(ctx).ContentJson().PostVar(ctx, getDocumentUrl(ctx, docType), data)
	err = detectError(rst)
	return
}

func (s *sDocument) Update(ctx context.Context, req *erp.ErpReq, data interface{}) (rst *g.Var, err error) {
	rst = GetClient(ctx).ContentJson().PutVar(ctx, getDocumentUrlWithName(ctx, req), data)
	err = detectError(rst)
	return
}

func (s *sDocument) Delete(ctx context.Context, req *erp.ErpReq) (rst *g.Var, err error) {
	rst = GetClient(ctx).DeleteVar(ctx, getDocumentUrlWithName(ctx, req))
	err = detectError(rst)
	return
}

func (s *sDocument) Copy(ctx context.Context, req *erp.ErpReq) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, fmt.Sprintf("%scopy", getDocumentUrlWithName(ctx, req)))
	err = detectError(rst)
	return
}

func (s *sDocument) Execute(ctx context.Context, req *erp.ErpReq, params interface{}) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, fmt.Sprintf("%smethod/%s", getDocumentUrlWithName(ctx, req), req.Method), params)
	err = detectError(rst)
	return
}

func (s *sDocument) ChangeDocStatus(ctx context.Context, doctype string, name string, docstatus string) (rst *g.Var, err error) {
	rst, err = s.Update(ctx, &erp.ErpReq{
		DocType: doctype,
		Name:    name,
	}, map[string]interface{}{
		"docstatus": docstatus,
	})
	return
}

func getDocumentUrl(ctx context.Context, docType string) string {
	return fmt.Sprintf("%s/%s", documentApiUrl, docType)
}

func getDocumentUrlWithName(ctx context.Context, req *erp.ErpReq) string {
	return fmt.Sprintf("%s/%s", getDocumentUrl(ctx, req.DocType), req.Name)
}
