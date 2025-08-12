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
	"ttpos-server-go/pkg/utils"

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
		SiteCode: syncUomAndAttributeReq.SiteCode,
		Branch:   syncUomAndAttributeReq.Branch,
	})
	if err != nil {
		return err
	}
	unitSort := 1
	var translateItems []utils.TranslateItem
	for _, uom := range uomList.List {
		translateItems = append(translateItems, utils.TranslateItem{
			Lang:    "en",
			Content: uom.UomName,
		})
	}
	res, err := translateClient.Translate(context.Background(), translateItems)
	if err != nil {
		return err
	}
	multiLanguageMap := make(map[string]dto.LocaleResponse)
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

	for _, uom := range uomList.List {
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
		db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
		db.Model(&model.ProductUnit{}).Create(&model.ProductUnit{
			Name:                  localeName.ToJson(),
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Sort:                  unitSort,
		})
		unitSort++
	}

	// // TODO 属性组和属性是否有区分
	// attributeList, err := s.GetAttributeList(ctx, req.GetAttributeListReq{
	// 	SiteCode: syncUomAndAttributeReq.SiteCode,
	// 	Branch:   syncUomAndAttributeReq.Branch,
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}
