package service

import (
	stdContext "context"

	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// testContext 测试用的 Context 实现
type testContext struct {
	ctx         stdContext.Context
	companyUuid uint64
	dbId        uint64
}

// NewTestContext 创建测试用 Context
func NewTestContext() context.Context {
	return &testContext{
		ctx:         stdContext.Background(),
		companyUuid: 1,
		dbId:        1,
	}
}

func (t *testContext) GetLanguage() string                     { return "zh" }
func (t *testContext) SetLanguage(language string)             {}
func (t *testContext) GetCompanyUuid() uint64                  { return t.companyUuid }
func (t *testContext) GetDbId() uint64                         { return t.dbId }
func (t *testContext) GetGin() *gin.Context                    { return nil }
func (t *testContext) GetContext() stdContext.Context          { return t.ctx }
func (t *testContext) GetSource() string                       { return "test" }
func (t *testContext) SetSource(source string)                 {}
func (t *testContext) GetScene() string                        { return "" }
func (t *testContext) SetScene(scene string)                   {}
func (t *testContext) GetCompany() *model.Company              { return &model.Company{} }
func (t *testContext) GetCompanySetting() model.CompanySetting { return model.CompanySetting{} }
func (t *testContext) GetStaff() model.Staff                   { return model.Staff{} }
func (t *testContext) GetStaffUuid() uint64                    { return 0 }
func (t *testContext) GetAssistantUuid() uint64                { return 0 }
func (t *testContext) GetDeviceSn() string                     { return "" }
func (t *testContext) GetDeviceUuid() uint64                   { return 0 }
func (t *testContext) GetDeskUuid() uint64                     { return 0 }
func (t *testContext) GetH5OrderUuid() string                  { return "" }
func (t *testContext) GetMember() model.Member                 { return model.Member{} }
func (t *testContext) SetMember(member model.Member)           {}
func (t *testContext) GetMemberUuid() uint64                   { return 0 }
func (t *testContext) SetDB(tx *gorm.DB)                       {}
func (t *testContext) SetCompanyUuid(uuid uint64)              { t.companyUuid = uuid }
func (t *testContext) SetCompanySetting(companySetting model.CompanySetting) {
}
func (t *testContext) SetCompany(company model.Company) {}
func (t *testContext) SetLog(log *zap.Logger)           {}
func (t *testContext) GetDB() *gorm.DB                  { return nil }
func (t *testContext) NoLock() bool                     { return true }
func (t *testContext) AddLock()                         {}
func (t *testContext) GetRequestUuid() string           { return "test-uuid" }
func (t *testContext) GetToken() string                 { return "" }
func (t *testContext) Copy() context.Context            { return t }
func (t *testContext) GetCfIPCountry() string           { return "CN" }
func (t *testContext) IsMobile() bool                   { return false }
func (t *testContext) GetRemoteIp() string              { return "127.0.0.1" }
func (t *testContext) GetCache() cache.Cache            { return nil }
func (t *testContext) Version(op context.Operator, version string) bool {
	return true
}
func (t *testContext) GetVersion() string { return "1.0.0" }
func (t *testContext) Log() *zap.Logger   { return zap.NewNop() }
