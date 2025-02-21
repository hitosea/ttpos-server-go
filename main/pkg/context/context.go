package context

import (
	"context"
	"go.uber.org/zap"
	"ttpos-server-go/app/model"

	"github.com/gin-gonic/gin"
)

type Context interface {
	GetLanguage() string                     // 获取语言
	GetCompanyUuid() uint64                  // 获取商家ID
	GetDbId() uint64                         // 获取商家ID
	GetGinContext() *gin.Context             // 获取gin上下文
	GetContext() context.Context             // 获取上下文
	GetSource() string                       // 获取请求来源
	GetCompany() model.Company               // 获取商家信息
	GetCompanySetting() model.CompanySetting // 获取商家设置
	GetStaff() model.Staff                   // 获取员工信息
	GetStaffUuid() uint64                    // 获取员工uuid
	GetDeviceSn() string                     // 获取设备SN
	GetDeskUuid() uint64                     // 获取桌台ID
	GetDeviceId() string                     // 获取设备ID
	Log() *zap.Logger                        // 获取日志实例
}
type ContextImpl struct {
	context.Context
	ginC           *gin.Context         // gin context。记录当前请求的上下文
	language       string               // 语言。记录当前请求的语言
	companyUuid    uint64               // 商家uuid。记录当前请求的商家
	source         string               // 请求来源
	company        model.Company        // 商家信息
	companySetting model.CompanySetting // 商家设置信息
	staff          model.Staff          // 员工信息，如果是点餐助手，应该是收银员
	staffUuid      uint64               // 员工uuid
	deskUuid       uint64               // 桌台ID
	deviceSn       string               // 设备序列号。用于唯一标识一个设备。如识别是哪个收银机，以找到收银机的未挂单点餐账单
	deviceId       string               // 设备ID
	log            *zap.Logger
}

type Option func(*ContextImpl)

func WithLanguage(language string) Option {
	return func(ctx *ContextImpl) {
		ctx.language = language
	}
}

func WithDeskUuid(deskUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.deskUuid = deskUuid
	}
}

func WithDeviceSn(deviceSn string) Option {
	return func(ctx *ContextImpl) {
		ctx.deviceSn = deviceSn
	}
}

func WithSource(source string) Option {
	return func(ctx *ContextImpl) {
		ctx.source = source
	}
}

func WithCompany(company model.Company) Option {
	return func(ctx *ContextImpl) {
		ctx.company = company
	}
}

func WithCompanySetting(companySetting model.CompanySetting) Option {
	return func(ctx *ContextImpl) {
		ctx.companySetting = companySetting
	}
}

func WithStaff(staff model.Staff) Option {
	return func(ctx *ContextImpl) {
		ctx.staff = staff
	}
}

func WithStaffUuid(staffUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.staffUuid = staffUuid
	}
}

func WithDeviceId(deviceId string) Option {
	return func(ctx *ContextImpl) {
		ctx.deviceId = deviceId
	}
}

func WithCompanyUuid(companyUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.companyUuid = companyUuid
	}
}

func WithLogger(log *zap.Logger) Option {
	return func(ctx *ContextImpl) {
		ctx.log = log
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

func (c *ContextImpl) GetDeskUuid() uint64 {
	return c.deskUuid
}

func (c *ContextImpl) GetDeviceId() string {
	return c.deviceId
}

func (c *ContextImpl) GetSource() string {
	return c.source
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

func (c *ContextImpl) GetCompany() model.Company {
	return c.company
}

func (c *ContextImpl) GetCompanySetting() model.CompanySetting {
	return c.companySetting
}

func (c *ContextImpl) GetStaff() model.Staff {
	return c.staff
}

func (c *ContextImpl) GetStaffUuid() uint64 {
	return c.staffUuid
}

func (c *ContextImpl) GetDeviceSn() string {
	return c.deviceSn
}

func (c *ContextImpl) Log() *zap.Logger {
	return c.log
}
