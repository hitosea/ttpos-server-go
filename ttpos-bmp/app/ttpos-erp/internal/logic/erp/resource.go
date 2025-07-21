package erp

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

var Resource = new(sResource)

const apiUrl = "/api/resource/"

type sResource struct {
}

func init() {
	// 注册服务
	service.RegisterResource(Resource)
}

func (s *sResource) List(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, getResourceUrl(docType), params)
	return
}

func (s *sResource) Get(ctx context.Context, docType string, name string, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().GetVar(ctx, getResourceUrlWithName(docType, name), params)
	return
}

func (s *sResource) Post(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().PostVar(ctx, getResourceUrl(docType), params)
	return
}

func (s *sResource) Put(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().PutVar(ctx, getResourceUrl(docType), params)
	return
}

func (s *sResource) Delete(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error) {
	rst = g.Client().DeleteVar(ctx, getResourceUrl(docType), params)
	return
}

func getResourceUrl(docType string) string {
	return fmt.Sprintf("%s%s%s", GetApiBase(), apiUrl, docType)
}

func getResourceUrlWithName(docType string, name string) string {
	return fmt.Sprintf("%s/%s", getResourceUrl(docType), name)
}

func GetApiBase() string {
	return g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.serviceUrl").String()
}
