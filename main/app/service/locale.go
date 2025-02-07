package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/model"
)

// ILocaleSrv 多语言名称服务接口
type ILocaleSrv interface {
	GetLocaleNames(m model.MultiLanguageName) dto.LocaleResponse // 获取多语言名称
}

// localeSrv 多语言名称服务结构体
type localeSrv struct {
}

// NewLocaleSrv 创建新的多语言名称服务
func NewLocaleSrv() ILocaleSrv {
	return NewLocaleSrvImpl()
}

// NewLocaleSrvImpl 创建新的多语言名称服务实现
func NewLocaleSrvImpl() ILocaleSrv {
	return &localeSrv{}
}

// GetLocaleNames 获取多语言名称
func (s *localeSrv) GetLocaleNames(m model.MultiLanguageName) dto.LocaleResponse {
	return dto.LocaleResponse{
		ZH:   m.ZhName,
		TH:   m.ThName,
		EN:   m.EnName,
		ZHTW: m.ZhTwName,
		JA:   m.JaName,
		KO:   m.KoName,
		MY:   m.MyName,
		TR:   m.TrName,
	}
}
