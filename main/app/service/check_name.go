package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"unicode/utf8"
)

type ICheckNameSrv interface {
	CheckNameExists(ctx context.Context, param req.CheckNameRequest) (resp.CheckNameResp, error) // 检查名称是否存在
	MakeCheckNameList(ctx context.Context, param dto.LocaleResponse) []req.CheckingName          // 生成检查名称列表
	InnerCheckNameExists(ctx context.Context, param req.CheckNameRequest) bool                   // 内部检查名称是否存在
	CheckNameLength(ctx context.Context, name string, maxLength int) bool                        // 检查名称长度
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
	// source -> 业务表名映射（不含前缀，由 repo 层自动拼接）
	tableMap := map[string]string{
		constant.CheckNameSourceUnit:             "product_unit",
		constant.CheckNameSourceProduct:          "product_package",
		constant.CheckNameSourceCategory:         "product_category",
		constant.CheckNameSourceSauce:            "product_sauce",
		constant.CheckNameSourceAttribute:        "product_attribute",
		constant.CheckNameSourceAttributeGroup:   "product_attribute_group",
		constant.CheckNameSourceFlavor:           "product_flavor",
		constant.CheckNameSourceMaterial:         "material",
		constant.CheckNameSourceMaterialCategory: "material_category",
		constant.CheckNameSourceMaterialUnit:     "material_unit",
		constant.CheckNameSourceWarehouse:        "warehouse",
		constant.CheckNameSourceBatchTag:         "batch_tag",
	}

	// 套餐组不检查重名
	if checkNameReq.Source == constant.CheckNameSourceProductPackageGroup {
		result := make([]resp.CheckNameItem, 0, len(checkNameReq.Names))
		for _, name := range checkNameReq.Names {
			result = append(result, resp.CheckNameItem{
				Lang:      name.Lang,
				TextExist: false,
			})
		}
		return resp.CheckNameResp{List: result}, nil
	}

	tableName, ok := tableMap[checkNameReq.Source]
	if !ok {
		return resp.CheckNameResp{}, errors.New("类型不支持")
	}

	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	var result []resp.CheckNameItem
	for _, name := range checkNameReq.Names {
		count, err := multiLanguageNameRepo.CountNameExistsByTable(tableName, name.Lang, name.Text, checkNameReq.Uuid)
		if err != nil {
			continue
		}
		result = append(result, resp.CheckNameItem{
			Lang:      name.Lang,
			TextExist: count > 0,
		})
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

// CheckNameLength 检查名称长度
func (s *checkNameSrv) CheckNameLength(ctx context.Context, name string, maxLength int) bool {
	if utf8.RuneCountInString(name) > maxLength {
		return false
	}
	return true
}
