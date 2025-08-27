package erp

import (
	"context"
	"errors"
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

// GetUomList 获取单位列表
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

// GetAttributeList 获取属性组列表
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

// SyncUomAndAttribute 同步单位和属性组
func (s *erpSrv) SyncUomAndAttribute(ctx cc.Context, syncUomAndAttributeReq req.SyncUomAndAttributeReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	translateClient := utils.NewTranslateClient()
	uomList, err := s.GetUomList(context.Background(), req.GetUomListReq{
		SiteCode: syncUomAndAttributeReq.SiteCode,
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
		SiteCode: syncUomAndAttributeReq.SiteCode,
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
			res, err = translateClient.Translate(context.Background(), translateItems)
			if err != nil {
				logger.Logger.Error("SyncUomAndAttribute-Translate-retry", zap.Any("translateItems", translateItems), zap.Any("err", err))
				continue
			}
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
	db.Model(&model.ProductUnit{}).Select("ifnull(MAX(sort), 0)").Scan(&unitSort)
	unitSort++

	for _, uom := range uomList.List {
		var localeName dto.LocaleResponse
		if _, ok := multiLanguageMap[uom.UomName]; !ok {
			localeName = dto.LocaleResponse{
				EN:   uom.UomName,
				ZH:   uom.UomName,
				TH:   uom.UomName,
				MY:   uom.UomName,
				JA:   uom.UomName,
				KO:   uom.UomName,
				TR:   uom.UomName,
				SV:   uom.UomName,
				ZHTW: uom.UomName,
			}
		} else {
			localeName = multiLanguageMap[uom.UomName]
		}
		multiLanguageName := model.MultiLanguageName{
			EnName:   localeName.EN,
			ZhName:   localeName.ZH,
			ThName:   localeName.TH,
			MyName:   localeName.MY,
			JaName:   localeName.JA,
			KoName:   localeName.KO,
			TrName:   localeName.TR,
			SvName:   localeName.SV,
			ZhTwName: localeName.ZHTW,
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
	db.Model(&model.ProductAttributeGroup{}).Select("ifnull(MAX(sort), 0)").Scan(&attributeGroupSort)
	attributeGroupSort++

	for _, erpnextAttributeGroup := range attributeGroupList.List {
		var localeName dto.LocaleResponse
		if _, ok := multiLanguageMap[erpnextAttributeGroup.AttributeName]; !ok {
			localeName = dto.LocaleResponse{
				EN:   erpnextAttributeGroup.AttributeName,
				ZH:   erpnextAttributeGroup.AttributeName,
				TH:   erpnextAttributeGroup.AttributeName,
				MY:   erpnextAttributeGroup.AttributeName,
				JA:   erpnextAttributeGroup.AttributeName,
				KO:   erpnextAttributeGroup.AttributeName,
				TR:   erpnextAttributeGroup.AttributeName,
				SV:   erpnextAttributeGroup.AttributeName,
				ZHTW: erpnextAttributeGroup.AttributeName,
			}
		} else {
			localeName = multiLanguageMap[erpnextAttributeGroup.AttributeName]
		}
		multiLanguageName := model.MultiLanguageName{
			EnName:   localeName.EN,
			ZhName:   localeName.ZH,
			ThName:   localeName.TH,
			MyName:   localeName.MY,
			JaName:   localeName.JA,
			KoName:   localeName.KO,
			TrName:   localeName.TR,
			SvName:   localeName.SV,
			ZhTwName: localeName.ZHTW,
		}
		err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
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

		for i, erpnextAttributeValue := range erpnextAttributeGroup.AttributeValueList {
			var localeName dto.LocaleResponse
			if _, ok := multiLanguageMap[erpnextAttributeValue.AttributeValue]; !ok {
				localeName = dto.LocaleResponse{
					EN:   erpnextAttributeValue.AttributeValue,
					ZH:   erpnextAttributeValue.AttributeValue,
					TH:   erpnextAttributeValue.AttributeValue,
					MY:   erpnextAttributeValue.AttributeValue,
					JA:   erpnextAttributeValue.AttributeValue,
					KO:   erpnextAttributeValue.AttributeValue,
					TR:   erpnextAttributeValue.AttributeValue,
					SV:   erpnextAttributeValue.AttributeValue,
					ZHTW: erpnextAttributeValue.AttributeValue,
				}
			} else {
				localeName = multiLanguageMap[erpnextAttributeValue.AttributeValue]
			}
			multiLanguageName := model.MultiLanguageName{
				EnName:   localeName.EN,
				ZhName:   localeName.ZH,
				ThName:   localeName.TH,
				MyName:   localeName.MY,
				JaName:   localeName.JA,
				KoName:   localeName.KO,
				TrName:   localeName.TR,
				SvName:   localeName.SV,
				ZhTwName: localeName.ZHTW,
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
				Sort:                  i + 1, // 每个属性组的属性值排序，从1开始
				ErpnextAttributeValue: erpnextAttributeValue.AttributeValue,
			}
			err = db.Model(&model.ProductAttribute{}).Create(&productAttribute).Error
			if err != nil {
				logger.Logger.Error("SyncUomAndAttribute-CreateProductAttribute-attribute", zap.Any("productAttribute", productAttribute), zap.Any("err", err))
				continue
			}
		}
	}

	return nil
}

// SaveUom 保存单位
func (s *erpSrv) SaveUom(ctx context.Context, saveUomReq req.SaveUomReq) error {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	req := &item.UomInfo{
		UomName:           saveUomReq.UomName,
		AliasName:         saveUomReq.AliasName,
		MustBeWholeNumber: saveUomReq.MustBeWholeNumber,
		CompanyAbbr:       saveUomReq.CompanyAbbr,
		Branch:            saveUomReq.Branch,
	}
	result, err := client.SaveUom(WithSiteCode(context.TODO(), saveUomReq.SiteCode), req)
	if err != nil {
		return err
	}
	if result.Code != "0" {
		return errors.New("保存失败")
	}
	return nil
}

// SaveAttribute 保存属性组
func (s *erpSrv) SaveAttribute(ctx context.Context, saveAttributeReq req.SaveAttributeReq) error {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	attributeValueList := make([]*item.AttributeValueInfo, 0)
	for _, attributeValue := range saveAttributeReq.AttributeValueList {
		attributeValueList = append(attributeValueList, &item.AttributeValueInfo{
			AttributeValue: attributeValue.AttributeValue,
			Abbr:           attributeValue.Abbr,
		})
	}

	req := &item.AttributeInfo{
		AttributeName:      saveAttributeReq.AttributeName,
		AliasName:          saveAttributeReq.AliasName,
		CompanyAbbr:        saveAttributeReq.CompanyAbbr,
		Branch:             saveAttributeReq.Branch,
		AttributeValueList: attributeValueList,
	}
	result, err := client.SaveAttribute(WithSiteCode(context.TODO(), saveAttributeReq.SiteCode), req)
	if err != nil {
		return err
	}
	if result.Code != "0" {
		return errors.New("保存失败")
	}

	return nil
}
