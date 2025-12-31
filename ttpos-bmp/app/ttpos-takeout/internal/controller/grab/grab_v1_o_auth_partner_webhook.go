package grab

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

func (c *ControllerV1) OAuthPartnerWebhook(ctx context.Context, req *v1.OAuthPartnerWebhookReq) (*v1.OAuthPartnerWebhookRes, error) {

	token, expiresIn, err := service.Grab().GetPartnerToken(ctx, req.ClientId, req.ClientSecret, req.Scope)
	if err != nil {
		return nil, gerror.Wrap(err, "获取 Partner Token 失败")
	}

	res := &v1.OAuthPartnerWebhookRes{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}

	return res, nil
}
