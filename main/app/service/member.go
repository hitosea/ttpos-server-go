package service

import (
	"github.com/jinzhu/copier"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                     // 获取等级列表
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList // 模糊搜索
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
