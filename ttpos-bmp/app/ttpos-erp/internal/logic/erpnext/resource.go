package erpnext

import (
	"context"
	"fmt"

	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
)

var Resource = new(sResource)

const apiUrl = "/api/resource/"

type sResource struct {
}

func init() {
	// 注册服务
	service.RegisterResource(Resource)
}

func (s *sResource) List(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).GetVar(ctx, getResourceUrl(docType), params)
	rst, err = detectError(resp)
	return
}

func (s *sResource) Get(ctx context.Context, docType string, name string, params *erp.RequestParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).GetVar(ctx, getResourceUrlWithName(docType, name), params)
	rst, err = detectError(resp)
	return
}

func (s *sResource) Post(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).PostVar(ctx, getResourceUrl(docType), params)
	rst, err = detectError(resp)
	return
}

func (s *sResource) Put(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).PutVar(ctx, getResourceUrl(docType), params)
	rst, err = detectError(resp)
	return
}

func (s *sResource) Delete(ctx context.Context, docType string, params *erp.RequestParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).DeleteVar(ctx, getResourceUrl(docType), params)
	rst, err = detectError(resp)
	return
}

func getResourceUrl(docType string) string {
	return fmt.Sprintf("%s%s", apiUrl, docType)
}

func getResourceUrlWithName(docType string, name string) string {
	return fmt.Sprintf("%s/%s", getResourceUrl(docType), name)
}
