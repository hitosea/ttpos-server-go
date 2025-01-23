package service

import (
	"context"

	"ttpos-server-go/app/dto/resp"
)

type ICaptchaService interface {
	Generate(ctx context.Context) (*resp.CaptchaResponse, error)
	Verify(ctx context.Context, sign string, answer string) bool
}
