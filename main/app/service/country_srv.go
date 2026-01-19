package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/material_resp"
	"ttpos-server-go/pkg/context"
)

// ICountrySrv 国家服务接口
//
// 任务: story-shop-material-origin Phase 3.3
// 需求: R3.4
//
// @version v2.11.0
type ICountrySrv interface {
	GetList(ctx context.Context) (material_resp.CountryListResp, error)
}

// countrySrv 国家服务实现
type countrySrv struct {
	// 无需依赖数据库，直接从常量读取
}

// NewCountrySrv 创建国家服务
func NewCountrySrv() ICountrySrv {
	return NewCountrySrvImpl()
}

// NewCountrySrvImpl 创建国家服务实现
func NewCountrySrvImpl() ICountrySrv {
	return &countrySrv{}
}

// GetList 获取国家列表
func (s *countrySrv) GetList(ctx context.Context) (material_resp.CountryListResp, error) {
	countries := constant.GetAllCountries()
	list := make([]material_resp.CountryItem, 0, len(countries))

	for _, country := range countries {
		list = append(list, material_resp.CountryItem{
			Code:       country.Code,
			LocaleName: country.GetLocaleNames(),
		})
	}

	return material_resp.CountryListResp{
		List: list,
	}, nil
}
