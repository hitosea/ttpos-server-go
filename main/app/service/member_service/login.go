package member_service

import (
	"errors"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type ILoginSrv interface {
	GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) // 获取登录信息
	SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                         // 发送验证码
}

type loginSrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewLoginSrv(
	dbm *database.DBManager,
	cache cache.Cache,
) ILoginSrv {
	return NewLoginSrvImpl(dbm, cache)
}

func NewLoginSrvImpl(
	dbm *database.DBManager,
	cache cache.Cache,
) ILoginSrv {
	return &loginSrv{
		dbm:   dbm,
		cache: cache,
	}
}

// GetLoginInfo 获取登录信息
func (s *loginSrv) GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) {
	// 获取国家代码并转换为大写
	countryCode := strings.ToUpper(ctx.GetCFIPCountry())

	// 定义国家代码到区号的映射
	countryToAreaCode := map[string]string{
		"CN": constant.ChinaPrefix,    // 中国
		"TH": constant.ThailandPrefix, // 泰国
	}

	// 默认区号列表
	areaCodes := []string{constant.ChinaPrefix, constant.ThailandPrefix}

	// 如果找到匹配的国家区号，将其移到第一位
	if areaCode, exists := countryToAreaCode[countryCode]; exists {
		// 从原数组中移除该区号
		for i, code := range areaCodes {
			if code == areaCode {
				areaCodes = append(areaCodes[:i], areaCodes[i+1:]...)
				break
			}
		}
		// 将匹配的区号放在第一位
		areaCodes = append([]string{areaCode}, areaCodes...)
	}

	return member_resp.MemberLoginInfoResp{
		CompanyUuid: req.CompanyUuid,
		AreaCode:    areaCodes,
	}, nil
}

// 发送验证码
func (s *loginSrv) SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error {
	// 判断 req.CompanyUuid 是否存在
	if req.CompanyUuid == 0 {
		return errors.New("商家ID不能为空")
	}
	if s.dbm.GetDB(req.CompanyUuid) == nil {
		return errors.New("商家不存在")
	}
	// 判断 req.Phone 是否是手机号
	if req.AreaCode == constant.ChinaPrefix {
		if len(req.Phone) != 11 {
			return errors.New("手机号格式不正确")
		}
		if req.Phone[0] != '1' {
			return errors.New("手机号格式不正确")
		}
	} else if req.AreaCode == constant.ThailandPrefix {
		if len(req.Phone) != 10 {
			return errors.New("手机号格式不正确")
		}
		if req.Phone[0] != '0' {
			return errors.New("手机号格式不正确")
		}
	} else {
		return errors.New("区号格式不正确")
	}

	// 判断 req.Phone 是否是手机号
	// if !utils.IsPhone(req.Phone) {
	// 	return errors.New("手机号格式不正确")
	// }

	return nil
}
