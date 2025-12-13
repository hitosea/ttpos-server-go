package service

import (
	"log"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha"
)

type ICaptchaSrv interface {
	Generate() (*resp.Captcha, error)
	Verify(sign string, answer string) bool
}

func NewCaptchaSrv(cache cache.Cache) ICaptchaSrv {
	return NewCaptchaSrvImpl(cache)
}

type captchaSrv struct {
	captcha     *captcha.Captcha
	cachePrefix string
}

func NewCaptchaSrvImpl(cache cache.Cache) ICaptchaSrv {
	srv := &captchaSrv{
		cachePrefix: "captcha:",
	}
	captchaTool, err := captcha.New(cache, srv.cachePrefix, 5*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	srv.captcha = captchaTool
	return srv
}

func (s *captchaSrv) Generate() (*resp.Captcha, error) {
	// 生成随机标识
	sign, b64s, err := s.captcha.Generate()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &resp.Captcha{
		Sign:   sign,
		Base64: b64s,
	}, nil
}

func (s *captchaSrv) Verify(sign, answer string) bool {
	ok, err := s.captcha.Verify(sign, answer)
	if err != nil {
		return false
	}
	return ok
}
