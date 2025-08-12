// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"

	"github.com/gogf/gf/v2/frame/g"
)

type (
	IDoctype interface {
		Meta(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error)
		Count(ctx context.Context, req *dto.ErpReq, params *dto.RequestParams) (int, error)
	}
	IDocument interface {
		List(ctx context.Context, req *dto.ErpReq, params *dto.RequestParams) (rst *g.Var, err error)
		Get(ctx context.Context, req *dto.ErpReq, params *dto.RequestParams) (rst *g.Var, err error)
		Create(ctx context.Context, docType string, data interface{}) (rst *g.Var, err error)
		Update(ctx context.Context, req *dto.ErpReq, data interface{}) (rst *g.Var, err error)
		Delete(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error)
		Copy(ctx context.Context, req *dto.ErpReq) (rst *g.Var, err error)
		Execute(ctx context.Context, req *dto.ErpReq, params interface{}) (rst *g.Var, err error)
	}
	IRpc interface {
		Execute(ctx context.Context, req *dto.ErpReq, params interface{}) (rst *g.Var, err error)
	}
	IResource interface {
		List(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error)
		Get(ctx context.Context, docType string, name string, params *dto.RequestParams) (rst *g.Var, err error)
		Post(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error)
		Put(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error)
		Delete(ctx context.Context, docType string, params *dto.RequestParams) (rst *g.Var, err error)
	}
)

var (
	localDoctype  IDoctype
	localDocument IDocument
	localRpc      IRpc
	localResource IResource
)

func Doctype() IDoctype {
	if localDoctype == nil {
		panic("implement not found for interface IDoctype, forgot register?")
	}
	return localDoctype
}

func RegisterDoctype(i IDoctype) {
	localDoctype = i
}

func Document() IDocument {
	if localDocument == nil {
		panic("implement not found for interface IDocument, forgot register?")
	}
	return localDocument
}

func RegisterDocument(i IDocument) {
	localDocument = i
}

func Rpc() IRpc {
	if localRpc == nil {
		panic("implement not found for interface IRpc, forgot register?")
	}
	return localRpc
}

func RegisterRpc(i IRpc) {
	localRpc = i
}

func Resource() IResource {
	if localResource == nil {
		panic("implement not found for interface IResource, forgot register?")
	}
	return localResource
}

func RegisterResource(i IResource) {
	localResource = i
}
