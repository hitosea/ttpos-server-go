package service

import (
	"time"
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
	// 供应商列表
	GetSupplierList(ctx context.Context, req req.SupplierListReq) (resp.SupplierListResp, error)
	// 创建供应商
	CreateSupplier(ctx context.Context, req req.SupplierCreateReq) (resp.SupplierCreateResp, error)
	// 更新供应商
	UpdateSupplier(ctx context.Context, req req.SupplierUpdateReq) error
	// 删除供应商
	DeleteSupplier(ctx context.Context, req req.SupplierDeleteReq) error
	// 获取供应商详情
	GetSupplierDetail(ctx context.Context, uuid uint64) (resp.SupplierDetailResp, error)
	// 获取供应商选择器列表
	GetSupplierSelect(ctx context.Context) (resp.SupplierSelectResp, error)
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

	// 名称筛选
	if req.Name != "" {
		opts = append(opts, supplierRepo.WhereNameLike(req.Name))
	}

	// 排除软删除
	opts = append(opts, supplierRepo.WhereNotDeleted())

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
	var supplierList []*resp.SupplierInfo
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

// CreateSupplier 创建供应商
func (s *supplierSrv) CreateSupplier(ctx context.Context, req req.SupplierCreateReq) (resp.SupplierCreateResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商名称是否重复
	exists, err := supplierRepo.IsNameExists(req.Name, 0)
	if err != nil {
		return resp.SupplierCreateResp{}, errors.WithMessage(err, "检查供应商名称失败")
	}
	if exists {
		return resp.SupplierCreateResp{}, errors.New("供应商名称已存在")
	}

	// 验证员工是否存在
	staffRepo := repository.NewStaffRepo(db)
	_, err = staffRepo.GetStaff(staffRepo.WhereUuid(req.StaffUuid))
	if err != nil {
		return resp.SupplierCreateResp{}, errors.New("采购负责人不存在")
	}

	// 生成UUID
	supplierUuid, _ := utils.GetID()

	// 创建供应商
	supplier := &model.Supplier{
		BaseModel: model.BaseModel{
			Uuid: supplierUuid,
		},
		Name:         req.Name,
		Address:      req.Address,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Position:     req.Position,
		StaffUuid:    req.StaffUuid,
	}

	err = supplierRepo.Create(supplier)
	if err != nil {
		return resp.SupplierCreateResp{}, errors.WithMessage(err, "创建供应商失败")
	}

	return resp.SupplierCreateResp{
		Uuid: supplierUuid,
	}, nil
}

// UpdateSupplier 更新供应商
func (s *supplierSrv) UpdateSupplier(ctx context.Context, req req.SupplierUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	supplier, err := supplierRepo.GetByUuid(req.Uuid, supplierRepo.WhereNotDeleted())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}

	// 检查供应商名称是否重复（排除自己）
	exists, err := supplierRepo.IsNameExists(req.Name, req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查供应商名称失败")
	}
	if exists {
		return errors.New("供应商名称已存在")
	}

	// 验证员工是否存在
	staffRepo := repository.NewStaffRepo(db)
	_, err = staffRepo.GetStaff(staffRepo.WhereUuid(req.StaffUuid))
	if err != nil {
		return errors.New("采购负责人不存在")
	}

	// 更新供应商信息
	supplier.Name = req.Name
	supplier.Address = req.Address
	supplier.ContactName = req.ContactName
	supplier.ContactPhone = req.ContactPhone
	supplier.Position = req.Position
	supplier.StaffUuid = req.StaffUuid
	supplier.UpdateTime = time.Now().Unix()

	err = supplierRepo.Update(supplier)
	if err != nil {
		return errors.WithMessage(err, "更新供应商失败")
	}

	return nil
}

// DeleteSupplier 删除供应商
func (s *supplierSrv) DeleteSupplier(ctx context.Context, req req.SupplierDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	_, err := supplierRepo.GetByUuid(req.Uuid, supplierRepo.WhereNotDeleted())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}

	// 检查是否有关联的采购订单
	var orderCount int64
	err = db.Model(&model.PurchaseOrder{}).
		Where("supplier_uuid = ? AND delete_time = ?", req.Uuid, 0).
		Count(&orderCount).Error
	if err != nil {
		return errors.WithMessage(err, "检查关联采购订单失败")
	}
	if orderCount > 0 {
		return errors.New("该供应商存在关联的采购订单，无法删除")
	}

	// 软删除供应商
	err = supplierRepo.Delete(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除供应商失败")
	}

	return nil
}

// GetSupplierDetail 获取供应商详情
func (s *supplierSrv) GetSupplierDetail(ctx context.Context, uuid uint64) (resp.SupplierDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 查询供应商
	supplier, err := supplierRepo.GetByUuid(uuid, supplierRepo.WhereNotDeleted())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp.SupplierDetailResp{}, errors.New("供应商不存在")
		}
		return resp.SupplierDetailResp{}, errors.WithMessage(err, "查询供应商失败")
	}

	// 转换响应格式
	supplierInfo := &resp.SupplierInfo{}
	err = copier.Copy(supplierInfo, supplier)
	if err != nil {
		return resp.SupplierDetailResp{}, errors.WithMessage(err, "数据转换失败")
	}

	return resp.SupplierDetailResp{
		SupplierInfo: supplierInfo,
	}, nil
}

// GetSupplierSelect 获取供应商选择器列表
func (s *supplierSrv) GetSupplierSelect(ctx context.Context) (resp.SupplierSelectResp, error) {

	// 调用erp接口
	if ctx.GetCompany().IsOpenErp() {
		erpResp, err := erp.NewIErpSrv(s.dbm).GetSupplierList(ctx)
		if err != nil {
			return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
		}
		// 转换响应格式
		var supplierList []*resp.SupplierSimpleInfo
		for _, supplier := range erpResp.SupplierList {
			supplierList = append(supplierList, &resp.SupplierSimpleInfo{
				Name: supplier.SupplierName,
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
		supplierRepo.WhereNotDeleted(),
		supplierRepo.OrderByName(false), // 按名称升序排序
	}

	suppliers, err := supplierRepo.GetList(opts...)
	if err != nil {
		return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
	}

	// 转换响应格式
	var supplierList []*resp.SupplierSimpleInfo
	for _, supplier := range suppliers {
		supplierList = append(supplierList, &resp.SupplierSimpleInfo{
			Name: supplier.Name,
		})
	}

	return resp.SupplierSelectResp{
		List: supplierList,
	}, nil
}
