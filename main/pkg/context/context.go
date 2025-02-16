package context

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Context interface {
	GetLanguage() string
	GetCompanyUuid() uint64
	GetDbId() uint64
	GetGinContext() *gin.Context
	GetContext() context.Context
}

type ContextImpl struct {
	context.Context
	ginC        *gin.Context // gin context。记录当前请求的上下文
	language    string       // 语言。记录当前请求的语言
	companyUuid uint64       // 商家uuid。记录当前请求的商家
}

type Option func(*ContextImpl)

func WithLanguage(language string) Option {
	return func(ctx *ContextImpl) {
		ctx.language = language
	}
}

func WithCompanyUuid(companyUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.companyUuid = companyUuid
	}
}

func WithGinContext(ginC *gin.Context) Option {
	return func(ctx *ContextImpl) {
		ctx.ginC = ginC
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *ContextImpl) {
		c.Context = ctx
	}
}

func NewDefaultContext() Context {
	return NewContext(WithContext(context.Background()))
}

func NewContextByGin(c *gin.Context) Context {
	return NewContext(WithGinContext(c), WithContext(c.Request.Context()))
}

func NewContext(options ...func(*ContextImpl)) Context {
	ctx := &ContextImpl{Context: context.Background()}
	for _, option := range options {
		option(ctx)
	}
	return ctx
}

func (c *ContextImpl) GetLanguage() string {
	return c.language
}

func (c *ContextImpl) GetCompanyUuid() uint64 {
	return c.companyUuid
}

func (c *ContextImpl) GetDbId() uint64 {
	return c.companyUuid
}

func (c *ContextImpl) GetGinContext() *gin.Context {
	return c.ginC
}

func (c *ContextImpl) GetContext() context.Context {
	return c.Context
}
