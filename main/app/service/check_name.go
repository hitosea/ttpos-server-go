package service

import (
	"fmt"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type ICheckNameSrv interface {
	CheckNameExists(ctx context.Context, param req.CheckNameRequest) (resp.CheckNameResp, error) // 检查名称是否存在
	MakeCheckNameList(ctx context.Context, param dto.LocaleResponse) []req.CheckingName          // 生成检查名称列表
	InnerCheckNameExists(ctx context.Context, param req.CheckNameRequest) bool                   // 内部检查名称是否存在
}

func NewCheckNameSrv(dbm *database.DBManager) ICheckNameSrv {
	return NewCheckNameSrvImpl(dbm)
}

type checkNameSrv struct {
	dbm *database.DBManager
}

func NewCheckNameSrvImpl(dbm *database.DBManager) ICheckNameSrv {
	return &checkNameSrv{
		dbm: dbm,
	}
}

// CheckNameExists 检查名称是否存在
func (s *checkNameSrv) CheckNameExists(ctx context.Context, checkNameReq req.CheckNameRequest) (resp.CheckNameResp, error) {
	keyMap := map[string]string{
		"zh":   "zh_name",
		"zhtw": "zh_tw_name",
		"en":   "en_name",
		"ja":   "ja_name",
		"ko":   "ko_name",
		"my":   "my_name",
		"sv":   "sv_name",
		"th":   "th_name",
		"tr":   "tr_name",
	}

	var result []resp.CheckNameItem

	for _, name := range checkNameReq.Names {
		// 为每个查询创建新的数据库连接实例
		db := s.dbm.GetDB(ctx.GetCompanyUuid())

		switch checkNameReq.Source {
		case constant.CheckNameSourceUnit:
			{
				var count int64
				query := db.Model(&model.ProductUnit{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_unit.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_unit.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceProduct:
			{
				var count int64
				query := db.Model(&model.ProductPackage{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_package.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_package.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceCategory:
			{
				var count int64
				query := db.Model(&model.ProductCategory{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_category.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_category.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceSauce:
			{
				var count int64
				query := db.Model(&model.ProductSauce{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_sauce.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_sauce.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceAttribute:
			{
				var count int64
				query := db.Model(&model.ProductAttribute{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_attribute.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_attribute.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceAttributeGroup:
			{
				var count int64
				query := db.Model(&model.ProductAttributeGroup{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_attribute_group.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_attribute_group.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case constant.CheckNameSourceFlavor:
			{
				var count int64
				query := db.Model(&model.ProductFlavor{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_flavor.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text)
				if checkNameReq.Uuid != 0 {
					query = query.Where("ttpos_product_flavor.uuid != ?", checkNameReq.Uuid)
				}
				query.Count(&count)

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		default:
			return resp.CheckNameResp{}, errors.New("类型不支持")
		}
	}

	return resp.CheckNameResp{
		List: result,
	}, nil
}

// NOTE: 后续如果需要添加其他语言，请在这里添加
func (s *checkNameSrv) MakeCheckNameList(ctx context.Context, param dto.LocaleResponse) []req.CheckingName {
	var names []req.CheckingName

	if param.ZH != "" {
		names = append(names, req.CheckingName{
			Lang: "zh",
			Text: param.ZH,
		})
	}

	if param.ZHTW != "" {
		names = append(names, req.CheckingName{
			Lang: "zh-TW",
			Text: param.ZHTW,
		})
	}

	if param.EN != "" {
		names = append(names, req.CheckingName{
			Lang: "en",
			Text: param.EN,
		})
	}

	if param.JA != "" {
		names = append(names, req.CheckingName{
			Lang: "ja",
			Text: param.JA,
		})
	}

	if param.KO != "" {
		names = append(names, req.CheckingName{
			Lang: "ko",
			Text: param.KO,
		})
	}

	if param.MY != "" {
		names = append(names, req.CheckingName{
			Lang: "my",
			Text: param.MY,
		})
	}

	if param.SV != "" {
		names = append(names, req.CheckingName{
			Lang: "sv",
			Text: param.SV,
		})
	}

	if param.TH != "" {
		names = append(names, req.CheckingName{
			Lang: "th",
			Text: param.TH,
		})
	}

	if param.TR != "" {
		names = append(names, req.CheckingName{
			Lang: "tr",
			Text: param.TR,
		})
	}

	return names
}

// InnerCheckNameExists 内部检查名称是否存在，不返回错误
func (s *checkNameSrv) InnerCheckNameExists(ctx context.Context, param req.CheckNameRequest) bool {
	checkNameResp, err := s.CheckNameExists(ctx, param)
	if err != nil {
		return false
	}
	for _, item := range checkNameResp.List {
		if item.TextExist {
			return true
		}
	}
	return false
}
