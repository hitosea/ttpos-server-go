package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

type RequestLog struct {
	CompanyUuid   uint64
	StaffUuid     uint64
	Source        string
	AccessName    string
	UserAgent     string
	UrlPath       string
	RequestMethod string
	RequestData   string
	IP            string
}

type IStaffOperationLogSrv interface {
	SaveLog(RequestLog)
	GetExcludedApis() []string
	GetAccessName(companyId uint64, urlPath string) string
}

func NewStaffOperationLogSrv(dbm *database.DBManager, authSrv IAuthSrv) IStaffOperationLogSrv {
	return NewStaffOperationLogSrvImpl(dbm, authSrv)
}

type staffOperationLogSrv struct {
	dbm          *database.DBManager // 数据库管理器
	authSrv      IAuthSrv
	excludedApis []string
}

func NewStaffOperationLogSrvImpl(dbm *database.DBManager, authSrv IAuthSrv) IStaffOperationLogSrv {
	return &staffOperationLogSrv{
		dbm:          dbm,
		authSrv:      authSrv,
		excludedApis: make([]string, 0), // TODO 完善要过滤掉的接口
	}
}

func (s *staffOperationLogSrv) GetExcludedApis() []string {
	return s.excludedApis
}

func (s *staffOperationLogSrv) GetAccessName(companyId uint64, urlPath string) string {
	accessRepo := repository.NewAccessRepo(s.dbm.GetDB(companyId))
	access := accessRepo.GetAccess(accessRepo.WhereApiPath(urlPath)) // TODO 完善权限迁移
	return access.Name
}

func (s *staffOperationLogSrv) SaveLog(requestLog RequestLog) {
	optLogRepo := repository.NewStaffOperationLogRepo(s.dbm.GetDB(requestLog.CompanyUuid))
	err := optLogRepo.SaveStaffOperationLog(model.StaffOperationLog{
		StaffUuid:   requestLog.StaffUuid,
		Title:       requestLog.AccessName,
		Url:         requestLog.UrlPath,
		RequestData: requestLog.RequestData,
		Type:        requestLog.RequestMethod,
		Ip:          requestLog.IP,
		Source:      requestLog.Source,
		Agent:       requestLog.UserAgent,
	})
	if err != nil {
		logger.Logger.Error("记录操作日志失败", zap.Error(err))
	}
}
