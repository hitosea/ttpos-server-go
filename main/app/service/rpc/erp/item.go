package erp

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	cc "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewErpItemClient() (item.ItemServiceClient, *grpc.ClientConn, error) {
	conn, err := cloud.GetRpcConnWithName(cloud.ErpServiceName)
	if err != nil {
		return nil, nil, err
	}
	return item.NewItemServiceClient(conn), conn, nil
}

func (s *erpSrv) GetUomList(ctx context.Context, getUomListReq req.GetUomListReq) (resp.GetUomListResp, error) {
	var getUomListResp resp.GetUomListResp
	client, conn, err := NewErpItemClient()
	if err != nil {
		return getUomListResp, err
	}
	defer conn.Close()
	req := &item.GetUomListReq{
		Branch:      getUomListReq.Branch,
		CompanyAbbr: getUomListReq.CompanyAbbr,
		UomName:     getUomListReq.UomName,
		AliasName:   getUomListReq.AliasName,
	}
	result, err := client.GetUomList(WithSiteCode(ctx, getUomListReq.SiteCode), req)
	if err != nil {
		return getUomListResp, err
	}
	response := &item.GetUomListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return getUomListResp, err
	}
	for _, uom := range response.UomList {
		getUomListResp.List = append(getUomListResp.List, resp.UomInfo{
			UomName:           uom.UomName,
			AliasName:         uom.AliasName,
			MustBeWholeNumber: uom.MustBeWholeNumber,
			CompanyAbbr:       uom.CompanyAbbr,
			Branch:            uom.Branch,
		})
	}
	return getUomListResp, nil
}

func (s *erpSrv) GetAttributeList(ctx context.Context, getAttributeListReq req.GetAttributeListReq) (resp.GetAttributeListResp, error) {
	var getAttributeListResp resp.GetAttributeListResp
	client, conn, err := NewErpItemClient()
	if err != nil {
		return getAttributeListResp, err
	}
	defer conn.Close()
	req := &item.GetAttributeListReq{
		Branch:        getAttributeListReq.Branch,
		CompanyAbbr:   getAttributeListReq.CompanyAbbr,
		AttributeName: getAttributeListReq.AttributeName,
		AliasName:     getAttributeListReq.AliasName,
	}
	result, err := client.GetAttributeList(WithSiteCode(ctx, getAttributeListReq.SiteCode), req)
	if err != nil {
		return getAttributeListResp, err
	}
	response := &item.GetAttributeListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return getAttributeListResp, err
	}
	for _, attribute := range response.AttributeList {
		getAttributeListResp.List = append(getAttributeListResp.List, resp.AttributeInfo{
			AttributeName:      attribute.AttributeName,
			AliasName:          attribute.AliasName,
			CompanyAbbr:        attribute.Company,
			Branch:             attribute.Branch,
			AttributeValueList: make([]resp.AttributeValueInfo, 0),
		})
		for _, attributeValue := range attribute.AttributeValueList {
			getAttributeListResp.List[len(getAttributeListResp.List)-1].AttributeValueList = append(getAttributeListResp.List[len(getAttributeListResp.List)-1].AttributeValueList, resp.AttributeValueInfo{
				AttributeValue: attributeValue.AttributeValue,
				Abbr:           attributeValue.Abbr,
			})
		}
	}
	return getAttributeListResp, nil
}

func (s *erpSrv) SyncUomAndAttribute(ctx cc.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	translateClient := utils.NewTranslateClient()
	uomList, err := s.GetUomList(context.Background(), req.GetUomListReq{
		SiteCode:    syncUomAndAttributeReq.SiteCode,
		CompanyAbbr: syncUomAndAttributeReq.CompanyAbbr,
	})
	if err != nil {
		logger.Logger.Error("SyncUomAndAttribute-GetUomList", zap.Any("err", err))
		return err
	}
	var translateItems []utils.TranslateItem
	for _, uom := range uomList.List {
		translateItems = append(translateItems, utils.TranslateItem{
			Lang:    "en",
			Content: uom.UomName,
		})
	}
	attributeGroupList, err := s.GetAttributeList(context.Background(), req.GetAttributeListReq{
		SiteCode:    syncUomAndAttributeReq.SiteCode,
		CompanyAbbr: syncUomAndAttributeReq.CompanyAbbr,
	})
	if err != nil {
		logger.Logger.Error("SyncUomAndAttribute-GetAttributeList", zap.Any("err", err))
		return err
	}
	for _, attributeGroup := range attributeGroupList.List {
		translateItems = append(translateItems, utils.TranslateItem{
			Lang:    "en",
			Content: attributeGroup.AttributeName,
		})
		for _, attributeValue := range attributeGroup.AttributeValueList {
			translateItems = append(translateItems, utils.TranslateItem{
				Lang:    "en",
				Content: attributeValue.AttributeValue,
			})
		}
	}

	// 获取两个数字中较小的那个
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	multiLanguageMap := make(map[string]dto.LocaleResponse)
	// 分组翻译，每次10个，翻译后保存到 multiLanguageMap
	for i := 0; i < len(translateItems); i += 10 {
		translateItems := translateItems[i:min(i+10, len(translateItems))]
		res, err := translateClient.Translate(context.Background(), translateItems)
		if err != nil {
			logger.Logger.Error("SyncUomAndAttribute-Translate", zap.Any("translateItems", translateItems), zap.Any("err", err))
			continue
		}
		for _, item := range res.Data {
			multiLanguageMap[item.Key] = dto.LocaleResponse{
				ZH:   item.Zh,
				TH:   item.Th,
				EN:   item.En,
				ZHTW: item.ZhTw,
				JA:   item.Ja,
				KO:   item.Ko,
				MY:   item.My,
				TR:   item.Tr,
				SV:   item.Sv,
			}
		}
	}

	// 获取单位最大的排序
	var unitSort int
	db.Model(&model.ProductUnit{}).Select("MAX(sort)").Scan(&unitSort)
	unitSort++

	for _, uom := range uomList.List {
		if _, ok := multiLanguageMap[uom.UomName]; !ok {
			logger.Logger.Error("SyncUomAndAttribute-multiLanguageMap-not-found", zap.Any("uomName", uom.UomName))
			continue
		}
		localeName := multiLanguageMap[uom.UomName]
		multiLanguageName := model.MultiLanguageName{
			EnName: localeName.EN,
			ZhName: localeName.ZH,
			ThName: localeName.TH,
			MyName: localeName.MY,
			JaName: localeName.JA,
			KoName: localeName.KO,
			TrName: localeName.TR,
			SvName: localeName.SV,
		}
		err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			logger.Logger.Error("SyncUomAndAttribute-CreateMultiLanguageName-uom", zap.Any("multiLanguageName", multiLanguageName), zap.Any("err", err))
			continue
		}
		productUnit := model.ProductUnit{
			Name:                  localeName.ToJson(),
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Sort:                  unitSort,
			ErpnextUom:            uom.UomName,
		}
		err = db.Model(&model.ProductUnit{}).Create(&productUnit).Error
		if err != nil {
			logger.Logger.Error("SyncUomAndAttribute-CreateProductUnit", zap.Any("productUnit", productUnit), zap.Any("err", err))
			continue
		}
		unitSort++
	}

	// 获取属性组最大的排序
	var attributeGroupSort int
	db.Model(&model.ProductAttributeGroup{}).Select("MAX(sort)").Scan(&attributeGroupSort)
	attributeGroupSort++

	// 获取属性最大的排序
	var attributeSort int
	db.Model(&model.ProductAttribute{}).Select("MAX(sort)").Scan(&attributeSort)
	attributeSort++

	for _, erpnextAttributeGroup := range attributeGroupList.List {
		if _, ok := multiLanguageMap[erpnextAttributeGroup.AttributeName]; !ok {
			logger.Logger.Error("SyncUomAndAttribute-multiLanguageMap-not-found", zap.Any("attributeGroupName", erpnextAttributeGroup.AttributeName))
			continue
		}
		localeName := multiLanguageMap[erpnextAttributeGroup.AttributeName]
		multiLanguageName := model.MultiLanguageName{
			EnName: localeName.EN,
			ZhName: localeName.ZH,
			ThName: localeName.TH,
			MyName: localeName.MY,
			JaName: localeName.JA,
			KoName: localeName.KO,
			TrName: localeName.TR,
			SvName: localeName.SV,
		}
		err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			logger.Logger.Error("SyncUomAndAttribute-CreateMultiLanguageName-attributeGroup", zap.Any("multiLanguageName", multiLanguageName), zap.Any("err", err))
			continue
		}

		attributeGroup := model.ProductAttributeGroup{
			Name:                      localeName.ToJson(),
			MultiLanguageNameUuid:     multiLanguageName.Uuid,
			Sort:                      attributeGroupSort,
			ErpnextAttributeGroupName: erpnextAttributeGroup.AttributeName,
		}
		err = db.Model(&model.ProductAttributeGroup{}).Create(&attributeGroup).Error
		if err != nil {
			logger.Logger.Error("SyncUomAndAttribute-CreateProductAttributeGroup-attributeGroup", zap.Any("attributeGroup", attributeGroup), zap.Any("err", err))
			continue
		}
		attributeGroupSort++

		for _, erpnextAttributeValue := range erpnextAttributeGroup.AttributeValueList {
			if _, ok := multiLanguageMap[erpnextAttributeValue.AttributeValue]; !ok {
				logger.Logger.Error("SyncUomAndAttribute-multiLanguageMap-not-found", zap.Any("attributeValue", erpnextAttributeValue.AttributeValue))
				continue
			}
			localeName := multiLanguageMap[erpnextAttributeValue.AttributeValue]
			multiLanguageName := model.MultiLanguageName{
				EnName: localeName.EN,
				ZhName: localeName.ZH,
				ThName: localeName.TH,
				MyName: localeName.MY,
				JaName: localeName.JA,
				KoName: localeName.KO,
				TrName: localeName.TR,
				SvName: localeName.SV,
			}
			err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
			if err != nil {
				logger.Logger.Error("SyncUomAndAttribute-CreateMultiLanguageName-attribute", zap.Any("multiLanguageName", multiLanguageName), zap.Any("err", err))
				continue
			}
			productAttribute := model.ProductAttribute{
				Name:                  localeName.ToJson(),
				MultiLanguageNameUuid: multiLanguageName.Uuid,
				AttributeGroupUuid:    attributeGroup.Uuid,
				Sort:                  attributeSort,
				ErpnextAttributeValue: erpnextAttributeValue.AttributeValue,
			}
			err = db.Model(&model.ProductAttribute{}).Create(&productAttribute).Error
			if err != nil {
				logger.Logger.Error("SyncUomAndAttribute-CreateProductAttribute-attribute", zap.Any("productAttribute", productAttribute), zap.Any("err", err))
				continue
			}
			attributeSort++
		}
	}

	return nil
}
