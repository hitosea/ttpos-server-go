package service

import (
	"context"
	"github.com/gogf/gf/contrib/registry/nacos/v2"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

var RpcClient = new(rpcClient)

type rpcClient struct {
	Ctx context.Context
}

func (r *rpcClient) Init(ctx context.Context) {
	grpcx.Resolver.Register(nacos.New(GetNacosAddress(ctx)))
	r.Ctx = ctx
}
