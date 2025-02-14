package service

import (
	"errors"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/cryptor"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                                    // 获取等级列表
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                // 模糊搜索
	AddMember(companyUuid uint64, addMemberReq req.AddMemberReq) error                    // 添加会员
	GetRechargeMember(companyUuid uint64, memberUuid uint64) (resp.RechargeMember, error) // 获取充值会员信息
}

// memberSrv 会员服务结构体
type memberSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewMemberSrv 创建新的会员服务
func NewMemberSrv(dbm *database.DBManager) IMemberSrv {
	return NewMemberSrvImpl(dbm)
}

// NewMemberSrvImpl 创建新的会员服务实现
func NewMemberSrvImpl(dbm *database.DBManager) IMemberSrv {
	return &memberSrv{dbm: dbm}
}

// GetLevels 获取等级列表
func (s *memberSrv) GetLevels(companyUuid uint64) resp.MemberLevelList {
	memberLevels := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).GetMemberLevels()
	respMemberLevels := make([]resp.MemberLevel, 0)
	for _, memberLevel := range memberLevels {
		var respMemberLevel resp.MemberLevel
		copier.Copy(&respMemberLevel, memberLevel)
		respMemberLevels = append(respMemberLevels, respMemberLevel)
	}
	return resp.MemberLevelList{
		List: respMemberLevels,
	}
}

// SearchMember 模糊搜索会员
func (s *memberSrv) SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList {
	searchMembers := make([]resp.SearchMember, 0)
	if keyword == "" {
		return resp.SearchMemberList{List: searchMembers}
	}
	members := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).SearchMember(keyword)
	for _, member := range members {
		var searchMember resp.SearchMember
		copier.Copy(&searchMember, member)
		searchMembers = append(searchMembers, searchMember)
	}
	return resp.SearchMemberList{
		List: searchMembers,
	}
}

// AddMember 添加会员
func (s *memberSrv) AddMember(companyUuid uint64, addMemberReq req.AddMemberReq) error {
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))

	// 判断等级是否存在
	if !memberRepo.CheckLevelExists(addMemberReq.LevelUuid) {
		return errors.New("会员等级不存在")
	}

	// 判断是否存在
	if memberRepo.CheckMemberExists(addMemberReq.Phone) {
		return errors.New("会员已存在")
	}
	if addMemberReq.Password != "" {
		addMemberReq.Password = cryptor.Md5String(addMemberReq.Password)
	}
	if err := memberRepo.CreateMember(model.Member{
		MemberNo:        utils.RandomNumber(5), // 5位数字
		Nickname:        addMemberReq.Nickname,
		Phone:           addMemberReq.Phone,
		Password:        "",
		MemberLevelUuid: addMemberReq.LevelUuid,
	}); err != nil {
		logger.Logger.Error("添加会员失败", zap.Error(err))
		return errors.New("添加会员失败")
	}
	return nil
}

// GetRechargeMember 获取充值会员信息
func (s *memberSrv) GetRechargeMember(companyUuid uint64, memberUuid uint64) (resp.RechargeMember, error) {
	_ = repository.NewMemberRepo(s.dbm.GetDB(companyUuid))

	// ToDo 获取信息

	return resp.RechargeMember{}, nil
}
