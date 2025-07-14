package takeout

import (
	"context"
	"takeout/api"
	"takeout/internal/consts"
	"takeout/internal/service"
)

type Takeout interface {
	EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error)
}

func GetService(name consts.ProviderName) Takeout {
	switch name {
	case consts.ProviderSkootar:
		return service.Skootar()
	default:
		return service.Skootar()
	}
}
