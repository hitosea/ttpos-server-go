package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// ISupplierSrv 供应商服务接口
type ISupplierSrv interface {
	GetSupplierList(ctx context.Context, req req.SupplierListReq) (resp.SupplierListResp, error)           // 供应商列表
	CreateSupplier(ctx context.Context, req req.SupplierCreateReq) error                                   // 创建供应商
	UpdateSupplier(ctx context.Context, req req.SupplierUpdateReq) error                                   // 更新供应商
	DeleteSupplier(ctx context.Context, req req.SupplierDeleteReq) error                                   // 删除供应商
	GetSupplierSelect(ctx context.Context, req req.SupplierSelectReq) (resp.SupplierSelectResp, error)     // 获取供应商选择器列表
	GetSupplier(ctx context.Context, req req.SupplierReq) (resp.SupplierResp, error)                       // 获取供应商
	CheckNameExists(ctx context.Context, req req.CheckNameExistsReq) (resp.CheckNameCodeExistsResp, error) // 检查名称是否存在
	CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) // 检查编码是否存在

	SyncSupplier(ctx context.Context) error // 同步供应商
}

// NewSupplierSrv 创建供应商服务
func NewSupplierSrv(dbm *database.DBManager) ISupplierSrv {
	return NewSupplierSrvImpl(dbm)
}

// supplierSrv 供应商服务实现
type supplierSrv struct {
	dbm *database.DBManager
}

// NewSupplierSrvImpl 创建供应商服务实现
func NewSupplierSrvImpl(dbm *database.DBManager) ISupplierSrv {
	return &supplierSrv{
		dbm: dbm,
	}
}

// GetSupplierList 获取供应商列表
func (s *supplierSrv) GetSupplierList(ctx context.Context, req req.SupplierListReq) (resp.SupplierListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	// 构建查询选项
	var opts []repository.DBOption
	// 名称编码筛选
	if req.Keyword != "" {
		opts = append(opts, supplierRepo.WhereNameOrCodeLike(req.Keyword))
	}
	// 排序
	opts = append(opts, supplierRepo.OrderByCreateTime(true))
	// 分页查询
	suppliers, total, err := supplierRepo.GetListWithPagination(
		req.PageReq.PageNo,
		req.PageReq.PageSize,
		opts...,
	)
	if err != nil {
		return resp.SupplierListResp{}, errors.WithMessage(err, "获取供应商列表失败")
	}
	// 转换响应格式
	supplierList := make([]*resp.SupplierInfo, 0, len(suppliers))
	for _, supplier := range suppliers {
		supplierInfo := &resp.SupplierInfo{}
		err := copier.Copy(supplierInfo, &supplier)
		if err != nil {
			continue
		}
		supplierList = append(supplierList, supplierInfo)
	}
	return resp.SupplierListResp{
		List: supplierList,
		Meta: dto.PageResponse{
			PageNo:   req.PageReq.PageNo,
			PageSize: req.PageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetSupplier 获取供应商
func (s *supplierSrv) GetSupplier(ctx context.Context, req req.SupplierReq) (resp.SupplierResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	supplier, err := supplierRepo.GetByUuid(req.Uuid)
	if err != nil {
		return resp.SupplierResp{}, errors.WithMessage(err, "获取供应商失败")
	}
	return resp.SupplierResp{
		SupplierInfo: &resp.SupplierInfo{
			Uuid:          supplier.Uuid,
			Name:          supplier.Name,
			Code:          supplier.Code,
			IsHeadquarter: supplier.ErpCode == constant.ErpHeadquartersSupplierCode,
		},
		Address:                 supplier.Address,
		ContactName:             supplier.ContactName,
		ContactPhone:            supplier.ContactPhone,
		Status:                  supplier.Status,
		HasRelatedPurchaseOrder: s.hasRelatedPurchaseOrder(ctx, supplier),
	}, nil
}

func (s *supplierSrv) hasRelatedPurchaseOrder(ctx context.Context, supplier *model.Supplier) bool {
	db := s.dbm.GetDB(ctx.GetDbId())
	exists, _ := repository.NewPurchaseOrderRepo(db).IsSupplierExists(supplier.ErpCode)
	return exists
}

// CreateSupplier 创建供应商
func (s *supplierSrv) CreateSupplier(ctx context.Context, createSupplierReq req.SupplierCreateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	// 检查供应商名称是否重复
	exists, err := supplierRepo.IsNameExists(createSupplierReq.Name, 0)
	if err != nil {
		return errors.WithMessage(err, "检查供应商名称失败")
	}
	if exists {
		return errors.New("供应商名称已存在")
	}
	// 检查供应商编码是否重复
	codeExists, err := supplierRepo.IsCodeExists(createSupplierReq.Code, 0)
	if err != nil {
		return errors.WithMessage(err, "检查供应商编码失败")
	}
	if codeExists {
		return errors.New("供应商编码已存在")
	}
	var erpCode string
	companySetting := ctx.GetCompanySetting()
	// 调用erp接口
	if ctx.GetCompany().IsOpenErp() {
		var branch, companyAbbr string
		// 总部调用erp接口创建的供应商，不传递branch、company_abbr
		// 子店、散户调用erp接口创建的供应商，传递branch、company_abbr
		if companySetting.IsHeadquarter() {
			branch = ""
			companyAbbr = ""
		} else {
			branch = companySetting.ErpnextBranchName
			companyAbbr = companySetting.ErpnextCompanyAbbr
		}
		erpCode, err = erp.NewIErpSrv(s.dbm).CreateSupplier(ctx.GetContext(), req.CreateSupplierReq{
			SiteCode:     companySetting.ErpnextSiteCode,
			SupplierName: createSupplierReq.Name,
			CompanyAbbr:  companyAbbr,
			Branch:       branch,
			Disabled:     createSupplierReq.Status == 0,
		})
		if err != nil {
			return errors.WithMessage(errors.New("创建供应商失败"), err.Error())
		}
	}
	// 生成UUID
	uuid, _ := utils.GetID()
	// 创建供应商
	supplier := &model.Supplier{
		BaseModel: model.BaseModel{
			Uuid: uuid,
		},
		Name:         createSupplierReq.Name,
		Code:         createSupplierReq.Code,
		Address:      createSupplierReq.Address,
		ContactName:  createSupplierReq.ContactName,
		ContactPhone: createSupplierReq.ContactPhone,
		Status:       createSupplierReq.Status,
		ErpCode:      erpCode,
		CompanyAbbr:  companySetting.ErpnextCompanyAbbr, // 这个标记是用来后续修改和删除
	}
	err = supplierRepo.Create(supplier)
	if err != nil {
		return errors.WithMessage(errors.New("创建供应商失败"), err.Error())
	}
	return nil
}

// UpdateSupplier 更新供应商
func (s *supplierSrv) UpdateSupplier(ctx context.Context, updateSupplierReq req.SupplierUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	supplier, err := supplierRepo.GetByUuid(updateSupplierReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}

	// 检查供应商名称是否重复（排除自己）
	exists, err := supplierRepo.IsNameExists(updateSupplierReq.Name, updateSupplierReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查供应商名称失败")
	}
	if exists {
		return errors.New("供应商名称已存在")
	}

	// 检查供应商编码是否重复（排除自己）
	codeExists, err := supplierRepo.IsCodeExists(updateSupplierReq.Code, updateSupplierReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查供应商编码失败")
	}
	if codeExists {
		return errors.New("供应商编码已存在")
	}

	companySetting := ctx.GetCompanySetting()
	// 调用erp接口，只能修改自己创建的供应商
	if ctx.GetCompany().IsOpenErp() && supplier.ErpCode != "" &&
		supplier.CompanyAbbr == companySetting.ErpnextCompanyAbbr && supplier.ErpCode != constant.ErpHeadquartersSupplierCode {
		var branch, companyAbbr string
		// 总部调用erp接口创建的供应商，不传递 branch、company_abbr
		// 子店、散户调用erp接口创建的供应商，传递 branch、company_abbr
		if companySetting.IsHeadquarter() {
			branch = ""
			companyAbbr = ""
		} else {
			branch = companySetting.ErpnextBranchName
			companyAbbr = companySetting.ErpnextCompanyAbbr
		}
		err = erp.NewIErpSrv(s.dbm).UpdateSupplier(ctx.GetContext(), req.UpdateSupplierReq{
			CreateSupplierReq: req.CreateSupplierReq{
				SupplierName: updateSupplierReq.Name,
				SiteCode:     companySetting.ErpnextSiteCode,
				CompanyAbbr:  companyAbbr,
				Branch:       branch,
				Disabled:     updateSupplierReq.Status == 0,
			},
			Name: supplier.ErpCode,
		})
		if err != nil {
			return errors.WithMessage(errors.New("更新供应商失败"), err.Error())
		}
	}

	// 所有店铺都能修改名称
	updateData := map[string]any{
		"name": updateSupplierReq.Name,
	}
	// 可以修改自己的供应商、总部可修改“总部-供应商”
	if supplier.CompanyAbbr == supplier.CompanyAbbr || (companySetting.IsHeadquarter() && supplier.ErpCode == constant.ErpHeadquartersSupplierCode) {
		updateData["address"] = updateSupplierReq.Address
		updateData["contact_name"] = updateSupplierReq.ContactName
		updateData["contact_phone"] = updateSupplierReq.ContactPhone
		updateData["status"] = updateSupplierReq.Status
	}
	err = supplierRepo.Update(supplier.Uuid, updateData)
	if err != nil {
		return errors.WithMessage(err, "更新供应商失败")
	}

	return nil
}

// DeleteSupplier 删除供应商
func (s *supplierSrv) DeleteSupplier(ctx context.Context, deleteSupplierReq req.SupplierDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	supplier, err := supplierRepo.GetByUuid(deleteSupplierReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}
	if s.hasRelatedPurchaseOrder(ctx, supplier) {
		return errors.New("该供应商存在关联的采购订单，无法删除")
	}
	companySetting := ctx.GetCompanySetting()
	// 调用erp接口，只能删除自己创建的供应商
	if ctx.GetCompany().IsOpenErp() && supplier.ErpCode != "" &&
		supplier.CompanyAbbr == companySetting.ErpnextCompanyAbbr && supplier.ErpCode != constant.ErpHeadquartersSupplierCode {
		err = erp.NewIErpSrv(s.dbm).DeleteSupplier(ctx.GetContext(), req.DeleteSupplierReq{
			SiteCode: companySetting.ErpnextSiteCode,
			Name:     supplier.ErpCode,
		})
		if err != nil {
			return errors.WithMessage(errors.New("删除供应商失败"), err.Error())
		}
	}
	// 软删除供应商
	err = supplierRepo.Delete(deleteSupplierReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除供应商失败")
	}
	return nil
}

// GetSupplierSelect 获取供应商选择器列表
func (s *supplierSrv) GetSupplierSelect(ctx context.Context, req req.SupplierSelectReq) (resp.SupplierSelectResp, error) {

	if ctx.GetCompany().IsOpenErp() && ctx.Version(context.LT, "2.6.0") {
		erpResp, err := erp.NewIErpSrv(s.dbm).GetSupplierList(ctx)
		if err != nil {
			return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
		}
		// 转换响应格式
		var supplierList []*resp.SupplierSimpleInfo
		for _, supplier := range erpResp.SupplierList {
			supplierList = append(supplierList, &resp.SupplierSimpleInfo{
				Name: supplier.SupplierName,
				Code: supplier.Name,
			})
		}
		return resp.SupplierSelectResp{
			List: supplierList,
		}, nil
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 构建查询选项
	opts := []repository.DBOption{
		supplierRepo.OrderByName(false), // 按名称升序排序
		supplierRepo.WhereNotDeleted(),
	}

	// 如果公司开启了erp，则查询erp供应商
	if ctx.GetCompany().IsOpenErp() {
		opts = append(opts, supplierRepo.WhereErpCodeExists())
	}

	suppliers, err := supplierRepo.GetList(opts...)
	if err != nil {
		return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
	}

	// 转换响应格式
	var supplierList []*resp.SupplierSimpleInfo
	for _, supplier := range suppliers {
		// 外部采购 去掉总部
		if req.PurchaseType == 1 {
			if supplier.ErpCode == constant.ErpHeadquartersSupplierCode {
				continue
			}
		} else {
			// 内部采购 去掉非总部
			if supplier.ErpCode != constant.ErpHeadquartersSupplierCode {
				continue
			}
		}
		supplierList = append(supplierList, &resp.SupplierSimpleInfo{
			Name: supplier.Name,
			Code: supplier.ErpCode,
		})
	}

	return resp.SupplierSelectResp{
		List: supplierList,
	}, nil
}

func (s *supplierSrv) CheckNameExists(ctx context.Context, req req.CheckNameExistsReq) (resp.CheckNameCodeExistsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	exists, err := supplierRepo.IsNameExists(req.Name, req.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查名称是否存在失败")
	}
	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}

func (s *supplierSrv) CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	exists, err := supplierRepo.IsCodeExists(req.Code, req.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查编码是否存在失败")
	}
	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}

func (s *supplierSrv) SyncSupplier(ctx context.Context) error {
	if !ctx.GetCompany().IsOpenErp() {
		return errors.New("公司未授权erp")
	}
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierList, err := erp.NewIErpSrv(s.dbm).ListSuppliers(ctx, req.GetErpnextSupplierListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return errors.WithMessage(errors.New("同步供应商失败"), err.Error())
	}
	for _, erpSupplier := range supplierList {
		var supplier model.Supplier
		db.Model(&model.Supplier{}).Where("erp_code = ?", erpSupplier.Name).Find(&supplier)
		if supplier.Uuid == 0 {
			db.Model(&model.Supplier{}).Create(&model.Supplier{
				Name:        erpSupplier.SupplierName,
				Code:        erpSupplier.Name,
				Status:      1,
				CompanyAbbr: companySetting.ErpnextCompanyAbbr,
			})
		} else {
			db.Model(&model.Supplier{}).Where("uuid = ?", supplier.Uuid).Updates(map[string]any{
				"name": erpSupplier.SupplierName,
			})
		}
	}

	// TODO 还需要同步总部创建允许看到的
	return nil
}
