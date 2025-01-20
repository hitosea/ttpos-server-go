package service

import (
	"context"
	
	"jjjshop-server-go/app/dto/resp"
)

type ICaptchaService interface {
	Generate(ctx context.Context) (*resp.CaptchaResponse, error)
	Verify(ctx context.Context, sign string, answer string) bool
}
