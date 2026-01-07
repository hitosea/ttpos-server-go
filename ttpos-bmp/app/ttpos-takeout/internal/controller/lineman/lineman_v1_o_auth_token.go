package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) OAuthToken(ctx context.Context, req *v1.OAuthTokenReq) (res *v1.OAuthTokenRes, err error) {

	token, expiresIn, err := service.LinemanToken().GenerateToken(ctx, req.ClientId, req.ClientSecret)
	if err != nil {
		return nil, gerror.Wrap(err, "获取 LINE MAN Token 失败")
	}

	res = &v1.OAuthTokenRes{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}

	return res, nil
}
