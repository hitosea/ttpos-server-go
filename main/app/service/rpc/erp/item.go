package erp

import (
	"context"
	"errors"
	"ttpos-bmp/app/ttpos-erp/api/item"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/logger"

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
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetUomList", zap.String("code", result.GetCode()), zap.String("result_message", result.GetMessage()))
		return getUomListResp, errors.New(result.GetMessage())
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
	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetAttributeList", zap.String("code", result.GetCode()), zap.String("result_message", result.GetMessage()))
		return getAttributeListResp, errors.New(result.GetMessage())
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
		return errors.New(result.GetMessage())
	}
	return nil
}

func (s *erpSrv) DeleteUom(ctx context.Context, deleteUomReq req.DeleteUomReq) error {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := client.DeleteUom(WithSiteCode(context.TODO(), deleteUomReq.SiteCode), &item.DeleteUomReq{
		Uom: deleteUomReq.UomName,
	})
	if err != nil {
		return err
	}
	if result.GetCode() != "0" {
		return errors.New(result.GetMessage())
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
		return errors.New(result.GetMessage())
	}

	return nil
}

// 保存规格
func (s *erpSrv) SaveFlavor(ctx context.Context, params req.SaveErpFlavorReq) error {
	client, conn, err := NewErpItemClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	// 保存规格值
	valueList := make([]*item.AttributeValueInfo, 0)
	for _, value := range params.ValueList {
		valueList = append(valueList, &item.AttributeValueInfo{
			AttributeValue: value.ValueName,
			Abbr:           value.ValueAliasName,
		})
	}

	// 保存规格组
	result, err := client.SaveAttribute(WithSiteCode(ctx, params.SiteCode), &item.AttributeInfo{
		AttributeName:      params.GroupName,
		Branch:             params.Branch,
		CompanyAbbr:        params.CompanyAbbr,
		AliasName:          params.GroupAliasName,
		AttributeValueList: valueList,
	})
	if err != nil {
		return err
	}
	if result.GetCode() != "0" {
		return errors.New(result.GetMessage())
	}

	return nil
}

// GetFlavorList 获取规格列表
func (s *erpSrv) GetFlavorList(ctx context.Context, params req.GetErpFlavorListReq) (resp.GetErpFlavorListResp, error) {
	var listResp resp.GetErpFlavorListResp
	client, conn, err := NewErpItemClient()
	if err != nil {
		return listResp, err
	}
	defer conn.Close()
	req := &item.GetAttributeListReq{
		Branch:        params.Branch,
		CompanyAbbr:   params.CompanyAbbr,
		AttributeName: params.AttributeName,
		AliasName:     params.AliasName,
	}
	result, err := client.GetAttributeList(WithSiteCode(ctx, params.SiteCode), req)
	if err != nil {
		return listResp, err
	}

	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetFlavorList", zap.String("code", result.GetCode()), zap.String("result_message", result.GetMessage()))
		return listResp, errors.New(result.GetMessage())
	}
	response := &item.GetAttributeListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return listResp, err
	}

	for _, attribute := range response.AttributeList {
		valueList := make([]resp.ErpFlavorValueInfo, 0)
		for _, attributeValue := range attribute.AttributeValueList {
			valueList = append(valueList, resp.ErpFlavorValueInfo{
				AttributeValue: attributeValue.AttributeValue,
				Abbr:           attributeValue.Abbr,
			})
		}

		listResp.List = append(listResp.List, resp.ErpFlavorInfo{
			AttributeName:      attribute.AttributeName,
			AliasName:          attribute.AliasName,
			CompanyAbbr:        attribute.CompanyAbbr,
			Branch:             attribute.Branch,
			AttributeValueList: valueList,
		})
	}
	return listResp, nil
}
