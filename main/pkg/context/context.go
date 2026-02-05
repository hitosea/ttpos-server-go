package context

import (
	"context"
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// Operator 版本比较操作符类型（向后兼容别名）
// 实际类型定义在 utils 包中
type Operator = utils.VersionOperator

const (
	LT  Operator = "<"
	GT  Operator = ">"
	EQ  Operator = "="
	GTE Operator = ">="
)

type Context interface {
	context.Context
	GetLanguage() string                     // 获取语言
	SetLanguage(language string)             // 设置语言
	GetCompanyUuid() uint64                  // 获取商家ID
	GetDbId() uint64                         // 获取商家ID
	GetGin() *gin.Context                    // 获取gin上下文
	GetContext() context.Context             // 获取上下文
	GetSource() string                       // 获取请求来源
	SetSource(source string)                 // 设置请求来源
	GetScene() string                        // 获取业务场景。用于区分不同的业务场景。如外送场景
	SetScene(scene string)                   // 设置业务场景。用于区分不同的业务场景。如外送场景
	GetCompany() *model.Company              // 获取商家信息
	GetCompanySetting() model.CompanySetting // 获取商家设置

	// 员工相关
	GetStaff() model.Staff    // 获取员工信息
	GetStaffUuid() uint64     // 获取员工uuid
	GetAssistantUuid() uint64 // 获取员工uuid(助手端)

	// 设备相关
	GetDeviceSn() string    // 获取设备SN
	GetDeviceUuid() uint64  // 获取设备uuid
	GetDeskUuid() uint64    // 获取桌台ID
	GetH5OrderUuid() string // 获取H5订单ID

	// 会员相关
	GetMember() model.Member       // 获取会员信息
	SetMember(member model.Member) // 设置会员信息
	GetMemberUuid() uint64         // 获取会员uuid

	// 数据库相关
	SetDB(tx *gorm.DB)                                     // 设置gorm.DB
	SetCompanyUuid(uuid uint64)                            // 设置商家ID
	SetCompanySetting(companySetting model.CompanySetting) // 设置商家设置
	SetCompany(company model.Company)                      // 设置商家
	SetLog(log *zap.Logger)                                // 设置日志
	GetDB() *gorm.DB                                       // 获取gorm.DB

	// 锁相关
	NoLock() bool // 判断上文中是否已经加锁
	AddLock()     // 在上下文中标记是否已经加锁 todo 改名为SetLock

	// 其他
	GetRequestUuid() string                   // 获取请求ID
	GetToken() string                         // 获取请求Token
	Copy() Context                            // 复制一个ctx实例。避免在协程中修改上下文导致主进程的ctx被修改
	GetCfIPCountry() string                   // 获取CF-IPCountry
	IsMobile() bool                           // 判断是否是移动端
	GetRemoteIp() string                      // 获取客户端IP
	GetCache() cache.Cache                    // 获取缓存
	Version(op Operator, version string) bool // 比较版本
	GetVersion() string                       // 获取客户端版本号
	SetVersion(version string)                // 设置版本
	Log() *zap.Logger                         // 获取日志实例

	GetBrand() string // 获取品牌名称

	// 基础信息
	SetBasicInfo(db *gorm.DB, companyInfo *model.Company) error // 设置基础信息

	// 耗时追踪
	GetStartTime() int64 // 获取请求开始时间（毫秒时间戳）
}

type ContextImpl struct {
	context.Context
	cc             *gin.Context         // gin context。记录当前请求的上下文
	language       string               // 语言。记录当前请求的语言
	companyUuid    uint64               // 商家uuid。记录当前请求的商家
	source         string               // 请求来源
	scene          string               // 业务场景。用于区分不同的业务场景。如外送场景
	company        model.Company        // 商家信息
	companySetting model.CompanySetting // 商家设置信息
	staff          model.Staff          // 员工信息，如果是点餐助手，应该是收银员
	staffUuid      uint64               // 员工uuid
	member         model.Member         // 会员信息
	memberUuid     uint64               // 会员uuid
	assistantUuid  uint64               // 员工uuid(助手端)
	deskUuid       uint64               // 桌台ID
	deviceSn       string               // 设备序列号。用于唯一标识一个设备。如识别是哪个收银机，以找到收银机的未挂单点餐账单
	deviceUuid     uint64               // 设备uuid
	requestUuid    string               // 请求uuid
	h5OrderUuid    string               // H5订单ID
	hasLock        bool                 // 是否已经上锁
	version        string               // 客户端版本号（用于队列等非HTTP场景）
	log            *zap.Logger
	db             *gorm.DB
	startTime      int64                // 请求开始时间（毫秒时间戳）
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

func WithDeviceUuid(deviceUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.deviceUuid = deviceUuid
	}
}

func WithRequestUuid(requestUuid string) Option {
	return func(ctx *ContextImpl) {
		ctx.requestUuid = requestUuid
	}
}

func WithH5OrderUuid(h5OrderUuid string) Option {
	return func(ctx *ContextImpl) {
		ctx.h5OrderUuid = h5OrderUuid
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

func WithMember(member model.Member) Option {
	return func(ctx *ContextImpl) {
		ctx.member = member
	}
}

func WithMemberUuid(memberUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.memberUuid = memberUuid
	}
}

func WithAssistantUuid(assistantUuid uint64) Option {
	return func(ctx *ContextImpl) {
		ctx.assistantUuid = assistantUuid
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

func WithGinContext(cc *gin.Context) Option {
	return func(ctx *ContextImpl) {
		ctx.cc = cc
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *ContextImpl) {
		c.Context = ctx
	}
}

func WithStartTime(startTime int64) Option {
	return func(ctx *ContextImpl) {
		ctx.startTime = startTime
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

// 复制一个ctx对象
func (c *ContextImpl) Copy() Context {
	return NewContext(
		WithLanguage(c.language),
		WithCompanyUuid(c.companyUuid),
		WithSource(c.source),
		WithCompany(c.company),
		WithStaff(c.staff),
		WithCompanySetting(c.companySetting),
		WithStaffUuid(c.staffUuid),
		WithDeskUuid(c.deskUuid),
		WithDeviceSn(c.deviceSn),
		WithDeviceUuid(c.deviceUuid),
		WithRequestUuid(c.requestUuid),
		WithLogger(c.log),
		WithGinContext(c.cc),
		WithContext(c.Context),
		WithMember(c.member),
		WithMemberUuid(c.memberUuid),
		WithStartTime(c.startTime),
	)
}

func (c *ContextImpl) GetLanguage() string {
	return c.language
}

func (c *ContextImpl) SetLanguage(language string) {
	c.language = language
}

func (c *ContextImpl) GetDeskUuid() uint64 {
	return c.deskUuid
}

func (c *ContextImpl) GetSource() string {
	return c.source
}

func (c *ContextImpl) SetSource(source string) {
	c.source = source
}

func (c *ContextImpl) GetScene() string {
	return c.scene
}

func (c *ContextImpl) SetScene(scene string) {
	c.scene = scene
}

func (c *ContextImpl) GetCompanyUuid() uint64 {
	return c.companyUuid
}

func (c *ContextImpl) GetDbId() uint64 {
	return c.companyUuid
}

func (c *ContextImpl) GetGin() *gin.Context {
	return c.cc
}

func (c *ContextImpl) GetContext() context.Context {
	return c.Context
}

func (c *ContextImpl) GetCompany() *model.Company {
	return &c.company
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

func (c *ContextImpl) GetMember() model.Member {
	return c.member
}

func (c *ContextImpl) SetMember(member model.Member) {
	c.member = member
}

func (c *ContextImpl) GetMemberUuid() uint64 {
	return c.memberUuid
}

func (c *ContextImpl) GetAssistantUuid() uint64 {
	return c.assistantUuid
}

func (c *ContextImpl) GetDeviceSn() string {
	if c.deviceSn == "" && c.cc != nil {
		c.deviceSn = c.cc.GetHeader("Device-Id")
	}
	return c.deviceSn
}

func (c *ContextImpl) GetDeviceUuid() uint64 {
	return c.deviceUuid
}

func (c *ContextImpl) GetH5OrderUuid() string {
	return c.h5OrderUuid
}

func (c *ContextImpl) NoLock() bool {
	return c.hasLock == false
}

func (c *ContextImpl) AddLock() {
	c.hasLock = true
}

func (c *ContextImpl) GetVersion() string {
	// 队列或后台任务场景，从内部变量读取
	if c.version != "" {
		return c.version
	}
	// HTTP 场景，从 Header 读取
	if c.cc != nil {
		version := c.cc.GetHeader("Client-Version")
		if version != "" {
			return version
		}
		return constant.ClientVersionV2000
	}
	return ""
}

func (c *ContextImpl) SetVersion(version string) {
	// 设置内部变量
	c.version = version
	// 如果有 gin.Context，同步到 Header
	if c.cc != nil {
		c.cc.Request.Header.Set("Client-Version", version)
	}
}
func (c *ContextImpl) Log() *zap.Logger {
	return c.log
}

func (c *ContextImpl) SetDB(tx *gorm.DB) {
	c.db = tx
}

func (c *ContextImpl) SetCompanyUuid(uuid uint64) {
	c.companyUuid = uuid
}

func (c *ContextImpl) SetLog(log *zap.Logger) {
	c.log = log
}

func (c *ContextImpl) SetCompany(company model.Company) {
	c.company = company
}

func (c *ContextImpl) SetCompanySetting(companySetting model.CompanySetting) {
	c.companySetting = companySetting
}

// GetDB 获取gorm实例
func (c *ContextImpl) GetDB() *gorm.DB {
	if c.db == nil {
		if db, exists := c.GetGin().Get(jwt.DB); exists {
			return db.(*gorm.DB)
		}
	}
	return c.db
}

// GetRequestUuid 获取请求uuid
func (c *ContextImpl) GetRequestUuid() string {
	return c.requestUuid
}

// GetToken 获取token
func (c *ContextImpl) GetToken() string {
	return strings.SplitN(c.cc.GetHeader("Authorization"), " ", 2)[1]
}

// GetCfIPCountry 获取CF-IPCountry
func (c *ContextImpl) GetCfIPCountry() string {
	return c.cc.GetHeader("CF-IPCountry")
}

// Version 比较客户端版本号
func (c *ContextImpl) Version(op Operator, version string) bool {
	// 使用 GetVersion() 方法，支持队列和 HTTP 两种场景
	clientVersion := c.GetVersion()
	return utils.CompareVersion(clientVersion, op, version)
}

// generateRandomIP 生成随机IP地址
func generateRandomIP() string {
	// 使用当前时间作为随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成随机IP地址，避免私有IP段
	var firstOctet int
	for {
		firstOctet = rand.Intn(254) + 1 // 1-254
		// 避免私有IP段和保留段
		// 10.0.0.0/8, 127.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
		if firstOctet != 10 && firstOctet != 127 && !(firstOctet >= 172 && firstOctet <= 191) && firstOctet != 192 {
			break
		}
	}

	// 后三段随机生成
	second := rand.Intn(256)
	third := rand.Intn(256)
	fourth := rand.Intn(254) + 1 // 1-254，避免0和255

	return fmt.Sprintf("%d.%d.%d.%d", firstOctet, second, third, fourth)
}

// GetRemoteIp 获取客户端IP
func (c *ContextImpl) GetRemoteIp() string {
	clientIp := c.cc.RemoteIP()
	if clientIp == "127.0.0.1" || clientIp == "::1" {
		return generateRandomIP()
	}
	return clientIp
}

// IsMobile 判断是否是移动端
func (c *ContextImpl) IsMobile() bool {
	userAgent := c.cc.GetHeader("User-Agent")
	// 检查 User-Agent 字符串中的常见移动设备标识
	mobileKeywords := []string{"Mobile", "Android", "iPhone", "iPad", "iPod", "BlackBerry", "Windows Phone", "iPad Desktop Mode", "Desktop with Mobile keyword"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	//
	return false
}

func (c *ContextImpl) GetCache() cache.Cache {
	return cache.Global
}

func (c *ContextImpl) GetBrand() string {
	return c.cc.GetString(jwt.Brand)
}

// GetStartTime 获取请求开始时间（毫秒时间戳）
func (c *ContextImpl) GetStartTime() int64 {
	return c.startTime
}

// 设置基础信息
func (c *ContextImpl) SetBasicInfo(db *gorm.DB, companyInfo *model.Company) error {
	c.SetDB(db)
	c.SetCompanyUuid(companyInfo.Uuid)
	c.SetCompany(*companyInfo)
	c.SetCompanySetting(*companyInfo.CompanySetting)
	c.SetVersion("v2.11.0")
	return nil
}
