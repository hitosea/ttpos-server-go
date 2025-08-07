package service

import (
	"fmt"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type ICheckNameSrv interface {
	CheckNameExist(ctx context.Context, param req.CheckNameRequest) (resp.CheckNameResp, error) // 检查名称是否存在
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

func (s *checkNameSrv) CheckNameExist(ctx context.Context, checkNameReq req.CheckNameRequest) (resp.CheckNameResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	keyMap := map[string]string{
		"zh":    "zh_name",
		"zh-TW": "zh_tw_name",
		"en":    "en_name",
		"ja":    "ja_name",
		"ko":    "ko_name",
		"my":    "my_name",
		"sv":    "sv_name",
		"th":    "th_name",
		"tr":    "tr_name",
	}

	var result []resp.CheckNameItem

	for _, name := range checkNameReq.Names {
		switch checkNameReq.Source {
		case "unit":
			{
				var count int64
				_ = db.Model(&model.ProductUnit{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_unit.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case "product":
			{
				var count int64
				_ = db.Model(&model.Product{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case "category":
			{
				var count int64
				_ = db.Model(&model.ProductCategory{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_category.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}
		case "sauce":
			{
				var count int64
				_ = db.Model(&model.ProductSauce{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_sauce.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}
		case "attribute":
			{
				var count int64
				_ = db.Model(&model.ProductAttribute{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_attribute.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

				result = append(result, resp.CheckNameItem{
					Lang:      name.Lang,
					TextExist: count > 0,
				})
			}

		case "attribute_group":
			{
				var count int64
				_ = db.Model(&model.ProductAttributeGroup{}).
					Joins("JOIN ttpos_multi_language_name ON ttpos_product_attribute_group.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
					Count(&count).Error

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
