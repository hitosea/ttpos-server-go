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

// ==================== Contact CRUD 方法 ====================

// GetContactList 获取联系人列表
// 根据查询条件过滤并返回联系人信息列表
func (s *sCrm) GetContactList(ctx context.Context, req *crm.GetContactListReq) (res *crm.GetContactListResp, err error) {
	// 构建查询过滤器
	filters := s.buildContactListFilters(ctx, req)

	// 查询联系人列表
	contactList, err := s.queryContactList(ctx, filters, req)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询联系人列表失败")
	}

	return &crm.GetContactListResp{
		ContactList: contactList,
	}, nil
}

// GetContact 获取联系人详情
// 根据联系人名称获取详细信息
func (s *sCrm) GetContact(ctx context.Context, req *crm.GetContactReq) (res *crm.GetContactResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("联系人名称不能为空")
	}

	// 查询联系人详情
	contact, err := s.queryContactDetail(ctx, req.Name)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询联系人详情失败")
	}

	return &crm.GetContactResp{
		Contact: contact,
	}, nil
}

// CreateContact 创建联系人
// 创建新的联系人记录
func (s *sCrm) CreateContact(ctx context.Context, req *crm.CreateContactReq) (res *crm.CreateContactResp, err error) {
	// 验证必填字段
	if err := s.validateCreateContactReq(req); err != nil {
		return nil, err
	}

	// 构建创建数据
	contactData := s.buildContactCreateData(req)

	// 调用ERP API创建联系人
	resp, err := service.Document().Create(ctx, erp.DocTypeContact, contactData)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建联系人失败")
	}

	// 解析响应获取创建的联系人名称
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析创建联系人响应失败")
	}

	contactName := j.Get("data.name").String()
	if contactName == "" {
		return nil, gerror.New("创建联系人失败，未获取到联系人名称")
	}

	return &crm.CreateContactResp{
		Name:    contactName,
		Success: true,
		Message: "联系人创建成功",
	}, nil
}

// UpdateContact 更新联系人
// 更新现有联系人的信息
func (s *sCrm) UpdateContact(ctx context.Context, req *crm.UpdateContactReq) (res *crm.UpdateContactResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("联系人名称不能为空")
	}

	// 构建更新数据
	updateData := s.buildContactUpdateData(req)

	// 调用ERP API更新联系人
	_, err = service.Document().Update(ctx, &erp.ErpReq{
		DocType: erp.DocTypeContact,
		Name:    req.Name,
	}, updateData)
	if err != nil {
		return nil, gerror.Wrapf(err, "更新联系人失败")
	}

	return &crm.UpdateContactResp{
		Success: true,
		Message: "联系人更新成功",
	}, nil
}

// DeleteContact 删除联系人
// 删除指定的联系人记录
func (s *sCrm) DeleteContact(ctx context.Context, req *crm.DeleteContactReq) (res *crm.DeleteContactResp, err error) {
	if req.Name == "" {
		return nil, gerror.New("联系人名称不能为空")
	}

	// 调用ERP API删除联系人
	_, err = service.Document().Delete(ctx, &erp.ErpReq{
		DocType: erp.DocTypeContact,
		Name:    req.Name,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "删除联系人失败")
	}

	return &crm.DeleteContactResp{
		Success: true,
		Message: "联系人删除成功",
	}, nil
}

// ==================== Contact 辅助方法 ====================

// buildContactListFilters 构建联系人列表查询过滤器
func (s *sCrm) buildContactListFilters(ctx context.Context, req *crm.GetContactListReq) [][]string {
	filters := make([][]string, 0)

	// 按状态过滤
	if req.Status != "" {
		filters = append(filters, g.ArrayStr{"status", "=", req.Status})
	}

	// 按关联文档类型过滤
	if req.LinkDoctype != "" && req.LinkName != "" {
		// 这里需要通过子表查询，暂时简化处理
		filters = append(filters, g.ArrayStr{"name", "like", "%"})
	}

	return filters
}

// queryContactList 执行联系人列表查询
func (s *sCrm) queryContactList(ctx context.Context, filters [][]string, req *crm.GetContactListReq) ([]*crm.ContactInfo, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = consts.Limit9999
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeContact,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "first_name", "last_name", "full_name", "email_id", "phone", "mobile_no", "status", "company_name", "creation", "modified"},
		Filters: filters,
		Limit:   limit,
	})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析联系人列表响应失败")
	}

	// 转换为联系人信息列表
	dataArray := j.GetJsons("data")
	contactList := make([]*crm.ContactInfo, 0, len(dataArray))

	for _, data := range dataArray {
		contactInfo := &crm.ContactInfo{
			Name:        data.Get("name").String(),
			FirstName:   data.Get("first_name").String(),
			LastName:    data.Get("last_name").String(),
			FullName:    data.Get("full_name").String(),
			EmailId:     data.Get("email_id").String(),
			Phone:       data.Get("phone").String(),
			MobileNo:    data.Get("mobile_no").String(),
			Status:      data.Get("status").String(),
			CompanyName: data.Get("company_name").String(),
		}
		contactList = append(contactList, contactInfo)
	}

	return contactList, nil
}

// queryContactDetail 查询联系人详情
func (s *sCrm) queryContactDetail(ctx context.Context, name string) (*crm.ContactInfo, error) {
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeContact,
		Name:    name,
	}, &erp.RequestParams{})

	if err != nil {
		return nil, err
	}

	// 解析响应数据
	j := resp
	if err != nil {
		return nil, gerror.Wrapf(err, "解析联系人详情响应失败")
	}

	data := j.GetJson("data")
	if data.IsNil() {
		return nil, gerror.New("联系人不存在")
	}

	// 转换为联系人信息
	contactInfo := &crm.ContactInfo{
		Name:             data.Get("name").String(),
		FirstName:        data.Get("first_name").String(),
		MiddleName:       data.Get("middle_name").String(),
		LastName:         data.Get("last_name").String(),
		FullName:         data.Get("full_name").String(),
		EmailId:          data.Get("email_id").String(),
		Phone:            data.Get("phone").String(),
		MobileNo:         data.Get("mobile_no").String(),
		Status:           data.Get("status").String(),
		Salutation:       data.Get("salutation").String(),
		Designation:      data.Get("designation").String(),
		Gender:           data.Get("gender").String(),
		CompanyName:      data.Get("company_name").String(),
		Department:       data.Get("department").String(),
		IsPrimaryContact: data.Get("is_primary_contact").Bool(),
		IsBillingContact: data.Get("is_billing_contact").Bool(),
	}

	// 注意：ContactInfo 结构体中没有 Links 字段
	// 如果需要关联链接信息，需要在 protobuf 文件中的 ContactInfo 消息中添加相应定义

	// 解析电话号码
	phoneArray := data.GetJsons("phone_nos")
	for _, phoneData := range phoneArray {
		phone := &crm.ContactPhoneInfo{
			Phone:             phoneData.Get("phone").String(),
			IsPrimaryPhone:    phoneData.Get("is_primary_phone").Bool(),
			IsPrimaryMobileNo: phoneData.Get("is_primary_mobile_no").Bool(),
		}
		contactInfo.PhoneNos = append(contactInfo.PhoneNos, phone)
	}

	return contactInfo, nil
}

// validateCreateContactReq 验证创建联系人请求
func (s *sCrm) validateCreateContactReq(req *crm.CreateContactReq) error {
	if req.FirstName == "" {
		return gerror.New("联系人名字不能为空")
	}
	return nil
}

// buildContactCreateData 构建联系人创建数据
func (s *sCrm) buildContactCreateData(req *crm.CreateContactReq) *erp.Contact {
	contact := &erp.Contact{
		FirstName: req.FirstName,
		Status:    req.Status,
	}

	// 可选字段
	if req.MiddleName != "" {
		contact.MiddleName = &req.MiddleName
	}
	if req.LastName != "" {
		contact.LastName = &req.LastName
	}
	if req.EmailId != "" {
		contact.EmailId = req.EmailId
	}
	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.MobileNo != "" {
		contact.MobileNo = req.MobileNo
	}
	if req.Salutation != "" {
		contact.Salutation = &req.Salutation
	}
	if req.Designation != "" {
		contact.Designation = &req.Designation
	}
	if req.Gender != "" {
		contact.Gender = &req.Gender
	}
	if req.CompanyName != "" {
		contact.CompanyName = &req.CompanyName
	}
	if req.Department != "" {
		contact.Department = &req.Department
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
		contact.Links = links
	}

	// 电话号码
	if len(req.PhoneNos) > 0 {
		phones := make([]erp.ContactPhone, 0, len(req.PhoneNos))
		for _, phone := range req.PhoneNos {
			phoneData := erp.ContactPhone{
				Phone: phone.Phone,
			}
			if phone.IsPrimaryPhone {
				phoneData.IsPrimaryPhone = 1
			}
			if phone.IsPrimaryMobileNo {
				phoneData.IsPrimaryMobileNo = 1
			}
			phones = append(phones, phoneData)
		}
		contact.PhoneNos = phones
	}

	return contact
}

// buildContactUpdateData 构建联系人更新数据
func (s *sCrm) buildContactUpdateData(req *crm.UpdateContactReq) *erp.Contact {
	contact := &erp.Contact{}

	// 更新字段
	if req.FirstName != "" {
		contact.FirstName = req.FirstName
	}
	if req.MiddleName != "" {
		contact.MiddleName = &req.MiddleName
	}
	if req.LastName != "" {
		contact.LastName = &req.LastName
	}
	if req.EmailId != "" {
		contact.EmailId = req.EmailId
	}
	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.MobileNo != "" {
		contact.MobileNo = req.MobileNo
	}
	if req.Status != "" {
		contact.Status = req.Status
	}
	if req.Salutation != "" {
		contact.Salutation = &req.Salutation
	}
	if req.Designation != "" {
		contact.Designation = &req.Designation
	}
	if req.Gender != "" {
		contact.Gender = &req.Gender
	}
	if req.CompanyName != "" {
		contact.CompanyName = &req.CompanyName
	}
	if req.Department != "" {
		contact.Department = &req.Department
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
		contact.Links = links
	}

	// 电话号码
	if len(req.PhoneNos) > 0 {
		phones := make([]erp.ContactPhone, 0, len(req.PhoneNos))
		for _, phone := range req.PhoneNos {
			phoneData := erp.ContactPhone{
				Phone: phone.Phone,
			}
			if phone.IsPrimaryPhone {
				phoneData.IsPrimaryPhone = 1
			}
			if phone.IsPrimaryMobileNo {
				phoneData.IsPrimaryMobileNo = 1
			}
			phones = append(phones, phoneData)
		}
		contact.PhoneNos = phones
	}

	return contact
}
