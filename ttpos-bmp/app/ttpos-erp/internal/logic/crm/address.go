package crm

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-erp/api/crm"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

// ==================== Address CRUD 方法 ====================

// GetAddressList 获取地址列表
// 根据查询条件过滤并返回地址信息列表
func (s *sCrm) GetAddressList(ctx context.Context, req *crm.GetAddressListReq) (res *crm.GetAddressListResp, err error) {
	// 构建查询过滤器
	filters := s.buildAddressListFilters(ctx, req)

	// 查询地址列表
	addressList, err := s.queryAddressList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询地址列表失败")
	}

	return &crm.GetAddressListResp{
		AddressList: addressList,
	}, nil
}

// GetAddress 获取地址详情
// 根据地址名称获取详细信息
func (s *sCrm) GetAddress(ctx context.Context, req *crm.GetAddressReq) (res *crm.GetAddressResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("地址名称不能为空")
	}

	// 查询地址详情
	address, err := s.queryAddressDetail(ctx, req.Name)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询地址详情失败")
	}

	return &crm.GetAddressResp{
		Address: address,
	}, nil
}

// CreateAddress 创建地址
// 创建新的地址记录
func (s *sCrm) CreateAddress(ctx context.Context, req *crm.CreateAddressReq) (res *crm.CreateAddressResp, err error) {
	// 验证必填字段
	if err := s.validateCreateAddressReq(req); err != nil {
		return nil, err
	}

	// 构建创建数据
	addressData := s.buildAddressCreateData(req)

	// 调用ERP API创建地址
	resp, err := service.Document().Create(ctx, erp.DocTypeAddress, addressData)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建地址失败")
	}

	// 解析响应获取创建的地址名称
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建地址响应失败")
	}

	addressName := j.Get("data.name").String()
	if addressName == "" {
		return nil, gerror.New("创建地址失败，未获取到地址名称")
	}

	return &crm.CreateAddressResp{
		Name: addressName,
	}, nil
}

// UpdateAddress 更新地址
// 更新现有地址的信息
func (s *sCrm) UpdateAddress(ctx context.Context, req *crm.UpdateAddressReq) (res *crm.UpdateAddressResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("地址名称不能为空")
	}

	// 构建更新数据
	updateData := s.buildAddressUpdateData(req)

	// 调用ERP API更新地址
	_, err = service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeAddress,
		Name:    req.Name,
	}, updateData)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新地址失败")
	}

	return &crm.UpdateAddressResp{
		Success: true,
		Message: "地址更新成功",
	}, nil
}

// DeleteAddress 删除地址
// 删除指定的地址记录
func (s *sCrm) DeleteAddress(ctx context.Context, req *crm.DeleteAddressReq) (res *crm.DeleteAddressResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("地址名称不能为空")
	}

	// 调用ERP API删除地址
	_, err = service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeAddress,
		Name:    req.Name,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "删除地址失败")
	}

	return &crm.DeleteAddressResp{
		Success: true,
		Message: "地址删除成功",
	}, nil
}

// ==================== Address 辅助方法 ====================

// buildAddressListFilters 构建地址列表查询过滤器
func (s *sCrm) buildAddressListFilters(ctx context.Context, req *crm.GetAddressListReq) [][]string {
	filters := make([][]string, 0)

	// 按地址类型过滤
	if req.AddressType != "" {
		filters = append(filters, g.ArrayStr{"address_type", "=", req.AddressType})
	}

	// 按关联文档类型过滤
	if req.LinkDoctype != "" && req.LinkName != "" {
		// 这里需要通过子表查询，暂时简化处理
		filters = append(filters, g.ArrayStr{"name", "like", "%"})
	}

	return filters
}

// queryAddressList 执行地址列表查询
func (s *sCrm) queryAddressList(ctx context.Context, filters [][]string, req *crm.GetAddressListReq) ([]*crm.AddressInfo, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = consts.Limit9999
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeAddress,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "address_title", "address_type", "address_line1", "city", "country", "phone", "email_id", "creation", "modified"},
		Filters: filters,
		Limit:   limit,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析地址列表响应失败")
	}

	// 转换为地址信息列表
	dataArray := j.GetJsons("data")
	addressList := make([]*crm.AddressInfo, 0, len(dataArray))

	for _, data := range dataArray {
		addressInfo := &crm.AddressInfo{
			Name:         data.Get("name").String(),
			AddressTitle: data.Get("address_title").String(),
			AddressType:  data.Get("address_type").String(),
			AddressLine1: data.Get("address_line1").String(),
			City:         data.Get("city").String(),
			Country:      data.Get("country").String(),
			Phone:        data.Get("phone").String(),
			EmailId:      data.Get("email_id").String(),
		}
		addressList = append(addressList, addressInfo)
	}

	return addressList, nil
}

// queryAddressDetail 查询地址详情
func (s *sCrm) queryAddressDetail(ctx context.Context, name string) (*crm.AddressInfo, error) {
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeAddress,
		Name:    name,
	}, &erp.RequestParams{})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析地址详情响应失败")
	}

	data := j.GetJson("data")
	if data.IsNil() {
		return nil, gerror.New("地址不存在")
	}

	// 转换为地址信息
	addressInfo := &crm.AddressInfo{
		Name:              data.Get("name").String(),
		AddressTitle:      data.Get("address_title").String(),
		AddressType:       data.Get("address_type").String(),
		AddressLine1:      data.Get("address_line1").String(),
		AddressLine2:      data.Get("address_line2").String(),
		City:              data.Get("city").String(),
		County:            data.Get("county").String(),
		State:             data.Get("state").String(),
		Country:           data.Get("country").String(),
		Pincode:           data.Get("pincode").String(),
		EmailId:           data.Get("email_id").String(),
		Phone:             data.Get("phone").String(),
		Fax:               data.Get("fax").String(),
		IsPrimaryAddress:  data.Get("is_primary_address").Bool(),
		IsShippingAddress: data.Get("is_shipping_address").Bool(),
		Disabled:          data.Get("disabled").Bool(),
	}

	// 注意：AddressInfo 结构体中没有 Links 字段
	// 如果需要关联链接信息，需要在 protobuf 文件中的 AddressInfo 消息中添加相应定义

	return addressInfo, nil
}

// validateCreateAddressReq 验证创建地址请求
func (s *sCrm) validateCreateAddressReq(req *crm.CreateAddressReq) error {
	if req.AddressTitle == "" {
		return gerror.New("地址标题不能为空")
	}
	if req.AddressLine1 == "" {
		return gerror.New("地址行1不能为空")
	}
	if req.City == "" {
		return gerror.New("城市不能为空")
	}
	if req.Country == "" {
		return gerror.New("国家不能为空")
	}
	return nil
}

// buildAddressCreateData 构建地址创建数据
func (s *sCrm) buildAddressCreateData(req *crm.CreateAddressReq) *erp.Address {
	address := &erp.Address{
		AddressTitle: req.AddressTitle,
		AddressLine1: req.AddressLine1,
		City:         req.City,
		Country:      req.Country,
		Doctype:      erp.DocTypeAddress,
	}

	// 可选字段
	if req.AddressType != "" {
		address.AddressType = req.AddressType
	}
	if req.AddressLine2 != "" {
		address.AddressLine2 = &req.AddressLine2
	}
	if req.County != "" {
		address.County = &req.County
	}
	if req.State != "" {
		address.State = &req.State
	}
	if req.Pincode != "" {
		address.Pincode = &req.Pincode
	}
	if req.EmailId != "" {
		address.EmailId = &req.EmailId
	}
	if req.Phone != "" {
		address.Phone = &req.Phone
	}
	if req.Fax != "" {
		address.Fax = &req.Fax
	}

	// 关联链接
	if len(req.Links) > 0 {
		links := make([]erp.DynamicLink, 0, len(req.Links))
		for _, link := range req.Links {
			links = append(links, erp.DynamicLink{
				LinkDoctype: link.LinkDoctype,
				LinkName:    link.LinkName,
				LinkTitle:   link.LinkTitle,
			})
		}
		address.Links = links
	}

	return address
}

// buildAddressUpdateData 构建地址更新数据
func (s *sCrm) buildAddressUpdateData(req *crm.UpdateAddressReq) *erp.Address {
	address := &erp.Address{}

	// 更新字段
	if req.AddressTitle != "" {
		address.AddressTitle = req.AddressTitle
	}
	if req.AddressType != "" {
		address.AddressType = req.AddressType
	}
	if req.AddressLine1 != "" {
		address.AddressLine1 = req.AddressLine1
	}
	if req.AddressLine2 != "" {
		address.AddressLine2 = &req.AddressLine2
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.County != "" {
		address.County = &req.County
	}
	if req.State != "" {
		address.State = &req.State
	}
	if req.Country != "" {
		address.Country = req.Country
	}
	if req.Pincode != "" {
		address.Pincode = &req.Pincode
	}
	if req.EmailId != "" {
		address.EmailId = &req.EmailId
	}
	if req.Phone != "" {
		address.Phone = &req.Phone
	}
	if req.Fax != "" {
		address.Fax = &req.Fax
	}

	// 布尔字段处理
	if req.IsPrimaryAddress {
		address.IsPrimaryAddress = 1
	} else {
		address.IsPrimaryAddress = 0
	}
	if req.IsShippingAddress {
		address.IsShippingAddress = 1
	} else {
		address.IsShippingAddress = 0
	}

	// 关联链接
	if len(req.Links) > 0 {
		links := make([]erp.DynamicLink, 0, len(req.Links))
		for _, link := range req.Links {
			links = append(links, erp.DynamicLink{
				LinkDoctype: link.LinkDoctype,
				LinkName:    link.LinkName,
				LinkTitle:   link.LinkTitle,
				Doctype:     "Dynamic Link",
			})
		}
		address.Links = links
	}

	return address
}
