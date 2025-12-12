package service

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IWarehouseSrv 仓库服务接口
type IWarehouseSrv interface {
	GetWarehouseList(ctx context.Context, req req.WarehouseListReq, isHeadquarters ...bool) (resp.WarehouseListResp, error) // 仓库列表
	GetHeadquarterWarehouseList(ctx context.Context) (resp.WarehouseListResp, error)                                        // 总部仓库列表
	CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error                                               // 创建仓库
	UpdateWarehouse(ctx context.Context, req req.UpdateWarehouseReq) error                                                  // 更新仓库
	DeleteWarehouse(ctx context.Context, req req.DeleteWarehouseReq) error                                                  // 删除仓库
	SetDefaultWarehouse(ctx context.Context, req req.SetDefaultWarehouseReq) error                                          // 设置默认仓库
	GetWarehouse(ctx context.Context, req req.WarehouseReq) (resp.WarehouseResp, error)                                     // 获取仓库
	GetWarehouseInOutList(ctx context.Context, req req.GetWarehouseInOutListReq) (resp.WarehouseInOutListResp, error)       // 出入库明细列表
	CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error)                  // 检查仓库编码是否存在
	GetOtherOrgList(ctx context.Context) (resp.OtherOrgListResp, error)                                                     // 对方机构列表
	GetWarehouseMaterialList(ctx context.Context, req req.WarehouseMaterialListReq) (resp.WarehouseMaterialListResp, error) // 获取仓库物品列表

	SyncWarehouse(ctx context.Context, syncHeadquarterData bool) error // 同步仓库列表
	SyncWarehouseItemStock(ctx context.Context) error                  // 同步仓库物品库存
}

// NewWarehouseSrv 创建仓库服务
func NewWarehouseSrv(dbm *database.DBManager, settingSrv setting.ISrv, materialSrv IMaterialSrv, translateSrv ITranslateSrv) IWarehouseSrv {
	return NewWarehouseSrvImpl(dbm, settingSrv, materialSrv, translateSrv)
}

// warehouseSrv 仓库服务实现
type warehouseSrv struct {
	dbm          *database.DBManager
	settingSrv   setting.ISrv
	materialSrv  IMaterialSrv
	translateSrv ITranslateSrv
}

// NewWarehouseSrvImpl 创建仓库服务实现
func NewWarehouseSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv, materialSrv IMaterialSrv, translateSrv ITranslateSrv) IWarehouseSrv {
	return &warehouseSrv{
		dbm:          dbm,
		settingSrv:   settingSrv,
		materialSrv:  materialSrv,
		translateSrv: translateSrv,
	}
}

// OtherOrgCode 对方机构编码
const (
	OtherOrgCodeSupplierErpCode = "supplier-erp-code"
	OtherOrgCodeCompanyUuid     = "company-uuid"
)

// GetWarehouseList 获取仓库列表
func (s *warehouseSrv) GetWarehouseList(ctx context.Context, req req.WarehouseListReq, isHeadquarters ...bool) (resp.WarehouseListResp, error) {
	isHeadquarter := false
	if len(isHeadquarters) > 0 {
		isHeadquarter = isHeadquarters[0]
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	// 构建查询选项
	var opts []repository.DBOption
	// 名称筛选
	if req.Keyword != "" {
		opts = append(opts, warehouseRepo.WhereNameOrCodeLike(req.Keyword))
	}
	// 类型筛选
	if req.Type != "" && slice.Contain([]string{"normal", "transit"}, req.Type) {
		opts = append(opts, warehouseRepo.WhereType(req.Type))
	}
	// 状态筛选
	if req.Status != nil && slice.Contain([]int{0, 1}, *req.Status) {
		opts = append(opts, warehouseRepo.WhereStatus(*req.Status))
	}
	// 是否总部筛选
	if isHeadquarter {
		opts = append(opts, warehouseRepo.WhereIsHeadquarter(isHeadquarter))
	} else {
		opts = append(opts, warehouseRepo.WhereHeadquarterUuid(0))
	}
	// 排序
	opts = append(opts, warehouseRepo.OrderByUpdateTime(true))
	// 分页查询
	warehouses, total, err := warehouseRepo.GetListWithPagination(
		req.PageReq.PageNo,
		req.PageReq.PageSize,
		opts...,
	)
	if err != nil {
		return resp.WarehouseListResp{}, errors.WithMessage(err, "获取仓库列表失败")
	}
	// 构建响应数据
	warehouseList := make([]resp.WarehouseResp, 0, len(warehouses))
	for _, warehouse := range warehouses {
		warehouseList = append(warehouseList, s.buildWarehouseResp(ctx, warehouse, isHeadquarter))
	}

	return resp.WarehouseListResp{
		List: warehouseList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetHeadquarterWarehouseList 获取总部仓库列表
func (s *warehouseSrv) GetHeadquarterWarehouseList(ctx context.Context) (resp.WarehouseListResp, error) {
	return s.GetWarehouseList(ctx, req.WarehouseListReq{
		PageReq: dto.PageReq{
			PageNo:   1,
			PageSize: 999,
		},
		Type:   "normal",
		Status: &[]int{1}[0],
	}, true)
}

// GetWarehouse 获取仓库
func (s *warehouseSrv) GetWarehouse(ctx context.Context, req req.WarehouseReq) (resp.WarehouseResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	opts := []repository.DBOption{
		warehouseRepo.WithItems(),
		warehouseRepo.WithMultiLanguageName(),
	}
	warehouse, err := warehouseRepo.GetByUuid(req.Uuid, opts...)
	if err != nil {
		return resp.WarehouseResp{}, errors.WithMessage(err, "获取仓库失败")
	}
	return s.buildWarehouseResp(ctx, *warehouse, false), nil
}

// CreateWarehouse 创建仓库
func (s *warehouseSrv) CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error {
	// 大写编码
	addReq.Code = strings.ToUpper(addReq.Code)
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	// 检查仓库编码是否已存在
	exists, err := warehouseRepo.IsCodeExists(addReq.Code, 0)
	if err != nil {
		return errors.WithMessage(err, "检查仓库编码失败")
	}
	if exists {
		return errors.New("仓库编码已存在")
	}
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 140) {
			return errors.New("仓库名称长度不能超过140")
		}
	}
	exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceWarehouse,
		Names:  names,
	})
	if exists {
		return errors.New("仓库名称已存在")
	}
	// 生成UUID
	uuid, err := utils.GetID()
	if err != nil {
		return errors.WithMessage(err, "生成UUID失败")
	}
	// 创建仓库模型
	warehouse := model.Warehouse{
		BaseModel: model.BaseModel{
			Uuid:       uuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
	}
	// 复制请求数据到模型
	err = copier.Copy(&warehouse, &addReq)
	if err != nil {
		return errors.WithMessage(err, "数据复制失败")
	}

	companySetting := ctx.GetCompanySetting()
	typeMap := map[string]string{
		"normal":  "Normal",
		"transit": "Transit",
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		warehouseName, err := GetEnName(ctx, s.settingSrv, addReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		// 保存多语言
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err = tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
		warehouse.Name = addReq.LocaleName.ToJson()
		warehouse.MultiLanguageNameUuid = multiLanguageName.Uuid
		var erpCode string
		// 调用erp接口
		if ctx.GetCompany().IsOpenErp() {
			erpCode, err = erp.NewIErpSrv(s.dbm).CreateWarehouse(ctx.GetContext(), req.CreateErpnextWarehouseReq{
				SiteCode:      companySetting.ErpnextSiteCode,
				WarehouseName: warehouseName,
				AliasName:     warehouseName,
				CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
				Branch:        companySetting.ErpnextBranchName,
				Disabled:      addReq.Status == 0,
				WarehouseType: typeMap[addReq.Type], // 修改为大驼峰
			})
			if err != nil {
				return errors.WithMessage(errors.New("创建仓库失败"), err.Error())
			}
		}
		warehouse.ErpCode = erpCode
		// 保存到数据库
		err = tx.Model(&model.Warehouse{}).Create(&warehouse).Error
		if err != nil {
			return errors.WithMessage(err, "创建仓库失败")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("创建仓库失败"), err.Error())
	}

	// 返回创建的仓库信息
	return nil
}

// UpdateWarehouse 更新仓库
func (s *warehouseSrv) UpdateWarehouse(ctx context.Context, updateReq req.UpdateWarehouseReq) error {
	// 大写编码
	updateReq.Code = strings.ToUpper(updateReq.Code)
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	// 检查仓库是否存在
	warehouse, err := warehouseRepo.GetByUuid(updateReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}

	if !isEditable(ctx, warehouse.HeadquarterUuid) {
		return errors.New("仓库不可编辑")
	}

	if (warehouse.IsDefault == 1 || warehouse.Type == constant.WarehouseTypeTransit) && updateReq.Status == 0 {
		return errors.New("默认仓库或在途仓库不可禁用")
	}

	// 检查仓库编码是否已存在（排除当前仓库）
	exists, err := warehouseRepo.IsCodeExists(updateReq.Code, warehouse.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查仓库编码失败")
	}
	if exists {
		return errors.New("仓库编码已存在")
	}
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, updateReq.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 140) {
			return errors.New("仓库名称长度不能超过140")
		}
	}
	exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   warehouse.Uuid,
		Source: constant.CheckNameSourceCategory,
		Names:  names,
	})
	if exists {
		return errors.New("仓库名称已存在")
	}

	updateData := map[string]any{
		"name":    updateReq.LocaleName.ToJson(),
		"type":    updateReq.Type,
		"code":    updateReq.Code,
		"status":  updateReq.Status,
		"contact": updateReq.Contact,
		"phone":   updateReq.Phone,
		"address": updateReq.Address,
	}

	companySetting := ctx.GetCompanySetting()
	typeMap := map[string]string{
		"normal":  "Normal",
		"transit": "Transit",
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", warehouse.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    updateReq.LocaleName.ZH,
			"th_name":    updateReq.LocaleName.TH,
			"en_name":    updateReq.LocaleName.EN,
			"zh_tw_name": updateReq.LocaleName.ZHTW,
			"ja_name":    updateReq.LocaleName.JA,
			"ko_name":    updateReq.LocaleName.KO,
			"my_name":    updateReq.LocaleName.MY,
			"tr_name":    updateReq.LocaleName.TR,
			"sv_name":    updateReq.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
		// 更新仓库
		err = tx.Model(&model.Warehouse{}).Where("uuid = ?", warehouse.Uuid).Updates(updateData).Error
		if err != nil {
			return errors.WithMessage(err, "更新仓库失败")
		}

		if ctx.GetCompany().IsOpenErp() && warehouse.ErpCode != "" {
			err = erp.NewIErpSrv(s.dbm).UpdateWarehouse(ctx.GetContext(), req.UpdateErpnextWarehouseReq{
				CreateErpnextWarehouseReq: req.CreateErpnextWarehouseReq{
					SiteCode:      companySetting.ErpnextSiteCode,
					CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
					Branch:        companySetting.ErpnextBranchName,
					Disabled:      updateReq.Status == 0,
					WarehouseType: typeMap[updateReq.Type],
				},
				Name: warehouse.ErpCode,
			})
			if err != nil {
				return errors.WithMessage(errors.New("更新仓库失败"), err.Error())
			}
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("更新仓库失败"), err.Error())
	}
	// 仓库 - 将多语言名称uuid从待翻译集合中删除
	s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), warehouse.MultiLanguageNameUuid)
	return nil
}

// DeleteWarehouse 删除仓库
func (s *warehouseSrv) DeleteWarehouse(ctx context.Context, deleteWarehouseReq req.DeleteWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	opts := []repository.DBOption{
		warehouseRepo.WithItems(),
	}
	// 检查仓库是否存在
	existingWarehouse, err := warehouseRepo.GetByUuid(deleteWarehouseReq.Uuid, opts...)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if existingWarehouse.IsDefault == 1 {
		return errors.New("默认仓库不可删除")
	}
	if !isEditable(ctx, existingWarehouse.HeadquarterUuid) {
		return errors.New("仓库不可删除")
	}
	if existingWarehouse.Type == constant.WarehouseTypeTransit {
		return errors.New("在途仓库不可删除")
	}

	if existingWarehouse.Items != nil && len(existingWarehouse.Items) > 0 {
		return errors.New("仓库存在关联的物品，不可删除")
	}

	if ctx.GetCompany().IsOpenErp() && existingWarehouse.ErpCode != "" {
		companySetting := ctx.GetCompanySetting()
		// 删除标记为erp禁用
		err = erp.NewIErpSrv(s.dbm).DeleteWarehouse(ctx.GetContext(), req.DeleteErpnextWarehouseReq{
			SiteCode: companySetting.ErpnextSiteCode,
			Name:     existingWarehouse.ErpCode,
		})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return errors.WithMessage(errors.New("删除仓库失败"), err.Error())
		}
	}
	// 执行软删除
	err = warehouseRepo.Delete(existingWarehouse.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("删除仓库失败"), err.Error())
	}

	return nil
}

// buildWarehouseResp 构建仓库响应
func (s *warehouseSrv) buildWarehouseResp(ctx context.Context, warehouse model.Warehouse, isHeadquarter bool) resp.WarehouseResp {
	var localName dto.LocaleResponse
	if warehouse.MultiLanguageName != nil {
		localName = warehouse.MultiLanguageName.GetNames()
	}
	return resp.WarehouseResp{
		Uuid:      warehouse.Uuid,
		LocalName: localName,
		Type:      warehouse.Type,
		Code: func() string {
			if isHeadquarter {
				return warehouse.ErpCode
			}
			return warehouse.Code
		}(),
		Status:     warehouse.Status,
		Contact:    warehouse.Contact,
		Phone:      warehouse.Phone,
		Address:    warehouse.Address,
		IsDefault:  warehouse.IsDefault,
		IsEditable: isEditable(ctx, warehouse.HeadquarterUuid),
		HasItem:    warehouse.Items != nil && len(warehouse.Items) > 0,
	}
}

// buildWarehouseInOutResp 构建仓库出入库明细响应
func (s *warehouseSrv) buildWarehouseInOutResp(log model.WarehouseInOutLog) resp.WarehouseInOutResp {
	// 获取物品信息
	var materialName dto.LocaleResponse
	var materialCode, materialBarcode string
	var materialCategoryUuid uint64
	if log.Material != nil {
		if log.Material.MultiLanguageName.Uuid != 0 {
			materialName = log.Material.MultiLanguageName.GetNames()
		}
		materialCode = log.Material.Code
		materialBarcode = log.Material.BarcodeValue
		materialCategoryUuid = log.Material.CategoryUuid
	}

	// 获取供应商信息
	var supplierName dto.LocaleResponse
	if log.Supplier != nil {
		supplierName = dto.LocaleResponse{
			ZH:   log.Supplier.Name,
			EN:   log.Supplier.Name,
			TH:   log.Supplier.Name,
			ZHTW: log.Supplier.Name,
			JA:   log.Supplier.Name,
			KO:   log.Supplier.Name,
			MY:   log.Supplier.Name,
			TR:   log.Supplier.Name,
			SV:   log.Supplier.Name,
		}
	} else {
		supplierName = dto.LocaleResponse{
			ZH:   log.SupplierName,
			EN:   log.SupplierName,
			TH:   log.SupplierName,
			ZHTW: log.SupplierName,
			JA:   log.SupplierName,
			KO:   log.SupplierName,
			MY:   log.SupplierName,
			TR:   log.SupplierName,
			SV:   log.SupplierName,
		}
	}

	// 获取仓库信息
	var warehouseName dto.LocaleResponse
	if log.Warehouse != nil && log.Warehouse.MultiLanguageName != nil {
		warehouseName = log.Warehouse.MultiLanguageName.GetNames()
	}

	// 类型
	typeStrMap := map[int]string{
		constant.WarehouseInOutLogScenePurchase:    "purchase",
		constant.WarehouseInOutLogSceneSale:        "sale",
		constant.WarehouseInOutLogSceneDelivery:    "delivery",
		constant.WarehouseInOutLogSceneProfitIn:    "profit_in",
		constant.WarehouseInOutLogSceneLossOut:     "loss_out",
		constant.WarehouseInOutLogSceneTransferIn:  "transfer_in",
		constant.WarehouseInOutLogSceneTransferOut: "transfer_out",
	}

	// 格式化日期
	date := ""
	if log.CreateTime > 0 {
		date = time.Unix(log.CreateTime, 0).Format("2006-01-02")
	}

	return resp.WarehouseInOutResp{
		Uuid:    log.Uuid,
		OrderNo: log.OrderNo,
		Type:    typeStrMap[log.Scene],
		Date:    date,
		Num:     log.Num,
		Amount:  log.Amount,
		// 物品信息
		MaterialUuid:         log.MaterialUuid,
		MaterialName:         materialName,
		MaterialCode:         materialCode,
		MaterialBarcode:      materialBarcode,
		MaterialCategoryUuid: materialCategoryUuid,
		// 供应商信息
		SupplierUuid:    log.SupplierUuid,
		SupplierErpCode: log.SupplierErpCode,
		SupplierName:    supplierName,
		// 仓库信息
		WarehouseUuid: log.WarehouseUuid,
		WarehouseName: warehouseName,
		// 对方机构
		OtherOrgType: log.OtherOrgType,
		OtherOrgName: log.OtherOrgName,
	}
}

func (s *warehouseSrv) SetDefaultWarehouse(ctx context.Context, req req.SetDefaultWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	// 检查仓库是否存在
	warehouse, err := warehouseRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if !isEditable(ctx, warehouse.HeadquarterUuid) {
		return errors.New("仓库不可设置为默认仓库")
	}
	if warehouse.Type == constant.WarehouseTypeTransit {
		return errors.New("在途仓库不可设置为默认仓库")
	}
	// 设置默认仓库
	err = warehouseRepo.UpdateIsDefault(warehouse.Uuid)
	if err != nil {
		return errors.WithMessage(err, "设置默认仓库失败")
	}

	return nil
}

// GetWarehouseInOutList 获取仓库出入库明细列表
func (s *warehouseSrv) GetWarehouseInOutList(ctx context.Context, req req.GetWarehouseInOutListReq) (resp.WarehouseInOutListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseInOutLogRepo := repository.NewWarehouseInOutLogRepo(db)
	// 构建查询条件
	opts := []repository.DBOption{}
	// 物品名称
	materialUuidsByKeyword := []uint64{}
	filterByKeyword := false
	if req.Keyword != "" {
		filterByKeyword = true
		// 从material表模糊查询，获取物品uuid列表
		materialRepo := repository.NewMaterialRepo(db)
		materialUuids, err := materialRepo.GetMaterialUuidsByKeyword(req.Keyword)
		if err != nil {
			return resp.WarehouseInOutListResp{}, errors.WithMessage(err, "获取物品uuid列表失败")
		}
		materialUuidsByKeyword = append(materialUuidsByKeyword, materialUuids...)
	}
	// 时间区间
	if req.StartTime != 0 && req.EndTime != 0 {
		// 判断前端传来的是毫秒时间戳，还是秒时间戳
		if req.StartTime > 10000000000 {
			req.StartTime = req.StartTime / 1000
		}
		if req.EndTime > 10000000000 {
			req.EndTime = req.EndTime / 1000
		}
		if req.StartTime > req.EndTime {
			return resp.WarehouseInOutListResp{}, errors.New("开始时间不能大于结束时间")
		}
		opts = append(opts, warehouseInOutLogRepo.WhereCreateTimeBetween(int(req.StartTime), int(req.EndTime)))
	}
	// 类型
	if req.Type != "" {
		logTypes := []int{}
		for _, typ := range strings.Split(req.Type, ",") {
			logTypes = append(logTypes, constant.WarehouseInOutLogTypeToInt(typ))
		}
		opts = append(opts, warehouseInOutLogRepo.WhereSceneIn(logTypes))
	}
	// 物料分类ID列表
	materialUuidsByCategory := []uint64{}
	filterByCategory := false
	if len(req.MaterialCategoryUuids) > 0 {
		filterByCategory = true
		// 查询material表，获取这些分类下的物品uuid列表
		materialRepo := repository.NewMaterialRepo(db)
		materialUuids, err := materialRepo.GetMaterialUuidsByCategoryUuids(req.MaterialCategoryUuids)
		if err != nil {
			return resp.WarehouseInOutListResp{}, errors.WithMessage(err, "获取物品uuid列表失败")
		}
		materialUuidsByCategory = append(materialUuidsByCategory, materialUuids...)
	}

	// 如果两个切片都有值的话，合并两个切片中的重复元素。即将keyword搜索的结果和category搜索的结果合并只取两个集合相交的数据

	if filterByKeyword {
		if len(materialUuidsByKeyword) == 0 { // 如果keyword搜索的结果为空，则不进行搜索
			return resp.WarehouseInOutListResp{
				List: make([]resp.WarehouseInOutResp, 0),
				Meta: dto.PageResponse{
					PageNo:   req.PageNo,
					PageSize: req.PageSize,
					Total:    0,
				},
			}, nil
		}
	}

	if filterByCategory {
		if len(materialUuidsByCategory) == 0 { // 如果category搜索的结果为空，则不进行搜索
			return resp.WarehouseInOutListResp{
				List: make([]resp.WarehouseInOutResp, 0),
				Meta: dto.PageResponse{
					PageNo:   req.PageNo,
					PageSize: req.PageSize,
					Total:    0,
				},
			}, nil
		}
	}
	// 如果两个条件都有值，则合并两个条件
	var filterOpt repository.DBOption
	if filterByKeyword && filterByCategory && len(materialUuidsByKeyword) > 0 && len(materialUuidsByCategory) > 0 {
		materialUuids := utils.GetDuplicateElements(materialUuidsByKeyword, materialUuidsByCategory)
		if len(materialUuids) > 0 {
			filterOpt = warehouseInOutLogRepo.WhereMaterialUuids(materialUuids)
		} else {
			// 没有匹配到任何物品
			return resp.WarehouseInOutListResp{
				List: make([]resp.WarehouseInOutResp, 0),
				Meta: dto.PageResponse{
					PageNo:   req.PageNo,
					PageSize: req.PageSize,
					Total:    0,
				},
			}, nil
		}
	}
	// 如果只有keyword条件，则使用keyword条件
	if filterByKeyword && !filterByCategory && len(materialUuidsByKeyword) > 0 {
		filterOpt = warehouseInOutLogRepo.WhereMaterialUuids(materialUuidsByKeyword)
	}
	// 如果只有category条件，则使用category条件
	if filterByCategory && !filterByKeyword && len(materialUuidsByCategory) > 0 {
		filterOpt = warehouseInOutLogRepo.WhereMaterialUuids(materialUuidsByCategory)
	}
	if filterOpt != nil {
		opts = append(opts, filterOpt)
	}

	// 供应商ID列表
	if len(req.SupplierUuids) > 0 {
		opts = append(opts, warehouseInOutLogRepo.WhereSupplierUuids(req.SupplierUuids))
	}
	// 单据编号
	if req.OrderNo != "" {
		opts = append(opts, warehouseInOutLogRepo.WhereOrderNo(req.OrderNo))
	}
	// 对方机构Code列表
	if len(req.OrgCodes) > 0 {
		supplierErpCodes := []string{}
		companyUuids := []uint64{}
		for _, orgCode := range req.OrgCodes {
			orgCodeArr := strings.Split(orgCode, ":")
			if len(orgCodeArr) < 2 {
				continue
			}
			if orgCodeArr[0] == OtherOrgCodeSupplierErpCode {
				supplierErpCodes = append(supplierErpCodes, orgCodeArr[1])
			} else if orgCodeArr[0] == OtherOrgCodeCompanyUuid {
				if companyUuid, err := strconv.ParseUint(orgCodeArr[1], 10, 64); err == nil {
					companyUuids = append(companyUuids, companyUuid)
				}
			}
		}
		if len(supplierErpCodes) > 0 || len(companyUuids) > 0 {
			opts = append(opts, warehouseInOutLogRepo.WhereSupplierErpCodesOrCompanyUuids(supplierErpCodes, companyUuids))
		}
	}

	// 排除在途仓的记录
	opts = append(opts, warehouseInOutLogRepo.WhereSceneNotIn([]int{20, 21}))

	// 获取列表
	warehouseInOutLogs, total, err := warehouseInOutLogRepo.GetListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.WarehouseInOutListResp{}, errors.WithMessage(err, "获取仓库出入库明细列表失败")
	}

	// 构建响应数据
	list := make([]resp.WarehouseInOutResp, 0, len(warehouseInOutLogs))
	for _, log := range warehouseInOutLogs {
		list = append(list, s.buildWarehouseInOutResp(log))
	}

	return resp.WarehouseInOutListResp{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *warehouseSrv) SyncWarehouse(ctx context.Context, syncHeadquarterData bool) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}

	// 本店获取erp仓库列表
	companySetting := ctx.GetCompanySetting()
	var warehouseErpCodes []string
	warehouseList, err := erp.NewIErpSrv(s.dbm).GetWarehouseList(ctx.GetContext(), req.GetErpnextWarehouseListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return errors.WithMessage(errors.New("同步仓库失败"), err.Error())
	}
	for _, warehouse := range warehouseList {
		warehouseErpCodes = append(warehouseErpCodes, warehouse.Name)
	}

	// 子店获取总部ttpos仓库
	var headquarter model.CompanySetting
	var headquarterWarehouses []model.Warehouse
	if companySetting.IsSubShop() && syncHeadquarterData {
		err := s.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).Where("uuid = ?", companySetting.HeadquarterUuid).Scopes(repository.NotDeleted).First(&headquarter).Error
		if err != nil || headquarter.Uuid == 0 {
			return errors.WithMessage(errors.New("获取总部公司失败"))
		}
		s.dbm.GetDB(headquarter.Uuid).Model(&model.Warehouse{}).Preload("MultiLanguageName").Find(&headquarterWarehouses)
	}

	// 本店ttpos仓库
	db := s.dbm.GetDB(companySetting.CompanyUuid)
	var warehouses []model.Warehouse
	// 是否存在默认仓库
	var existsDefaultWarehouse bool
	// ttpos 仓库编码列表
	var warehouseCodes []string
	// 待删除的仓库uuid列表
	var deletingWarehouseUuids []uint64
	// 本店仓库 erp_code 和 model.Warehouse 的映射
	warehouseMap := make(map[string]model.Warehouse)
	db.Model(&model.Warehouse{}).Scopes(repository.ExcludeHeadquarter).Where("erp_code != ''").Find(&warehouses)
	for _, warehouse := range warehouses {
		if warehouse.IsDefault == 1 {
			existsDefaultWarehouse = true
		}
		if !slices.Contains(warehouseErpCodes, warehouse.ErpCode) {
			deletingWarehouseUuids = append(deletingWarehouseUuids, warehouse.Uuid)
		}
		warehouseMap[warehouse.ErpCode] = warehouse
		warehouseCodes = append(warehouseCodes, warehouse.Code)
	}

	// 待翻译多语言Uuid
	var multiLanguageNameUuids []uint64

	err = db.Transaction(func(tx *gorm.DB) error {
		if len(deletingWarehouseUuids) > 0 {
			err = tx.Model(&model.Warehouse{}).Where("uuid IN (?)", deletingWarehouseUuids).Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("erp已删除仓库，标记删除ttpos仓库失败"), err.Error())
			}
		}
		var insertingWarehouses []model.Warehouse
		var defaultWarehouseErpCode string
		for _, erpWarehouse := range warehouseList {
			warehouse := warehouseMap[erpWarehouse.Name]
			contact := warehouse.Contact
			phone := warehouse.Phone
			address := warehouse.Address
			var status int
			if !erpWarehouse.Disabled {
				status = 1
			}
			var code string
			isDefault := warehouse.IsDefault
			if strings.Contains(erpWarehouse.Name, constant.NormalWarehouseCodeContains) {
				code = constant.NormalWarehouseCode
				if !existsDefaultWarehouse {
					isDefault = 1
				}
			} else if strings.Contains(erpWarehouse.Name, constant.TransitWarehouseCodeContains) {
				code = constant.TransitWarehouseCode
			}
			// 默认仓库erp_code
			if isDefault == 1 {
				defaultWarehouseErpCode = erpWarehouse.Name
			}

			var warehouseType string
			if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal1 || erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal2 {
				warehouseType = constant.WarehouseTypeNormal
			} else if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeTransit {
				warehouseType = constant.WarehouseTypeTransit
			}

			if warehouse.Uuid != 0 { // 如果存在，则更新
				db.Model(&model.Warehouse{}).Where("uuid = ?", warehouse.Uuid).Updates(map[string]any{
					"type":        warehouseType,
					"status":      status,
					"is_default":  isDefault,
					"contact":     contact,
					"phone":       phone,
					"address":     address,
					"delete_time": constant.NotDeleted,
				})
			} else { // 新增
				// 全部语言默认为erp返回的名称，应该是英文名称
				multiLanguageName := model.MultiLanguageName{
					ZhName:   erpWarehouse.AliasName,
					ThName:   erpWarehouse.AliasName,
					EnName:   erpWarehouse.AliasName,
					ZhTwName: erpWarehouse.AliasName,
					JaName:   erpWarehouse.AliasName,
					KoName:   erpWarehouse.AliasName,
					MyName:   erpWarehouse.AliasName,
					TrName:   erpWarehouse.AliasName,
					SvName:   erpWarehouse.AliasName,
				}
				err = tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
				if err != nil {
					return errors.WithMessage(errors.New("创建多语言名称失败"), err.Error())
				}
				// 处理code
				if code == "" {
					// 遍历warehouseCodes，获取以WH开头的最大数字
					maxCode := 2
					for _, warehouseCode := range warehouseCodes {
						if after, ok0 := strings.CutPrefix(warehouseCode, "WH"); ok0 {
							codeInt, _ := strconv.Atoi(after)
							if codeInt > maxCode {
								maxCode = codeInt
							}
						}
					}
					code = fmt.Sprintf("WH%02d", maxCode+1)
					warehouseCodes = append(warehouseCodes, code)
				}
				localeName := multiLanguageName.GetNames()
				insertingWarehouses = append(insertingWarehouses, model.Warehouse{
					Name:                  localeName.ToJson(),
					MultiLanguageNameUuid: multiLanguageName.Uuid,
					Type:                  warehouseType,
					Code:                  code,
					Status:                status,
					IsDefault:             isDefault,
					ErpCode:               erpWarehouse.Name,
					Contact:               contact,
					Phone:                 phone,
					Address:               address,
				})
				multiLanguageNameUuids = append(multiLanguageNameUuids, multiLanguageName.Uuid)
			}
		}

		// 同步ttpos总店数据（多语言由 SyncMultiLanguage 任务处理）
		if len(headquarterWarehouses) > 0 && syncHeadquarterData {
			// 删除仓库
			tx.Where("headquarter_uuid > 0").Delete(&model.Warehouse{})

			for _, headquarterWarehouse := range headquarterWarehouses {
				multiLanguageName := headquarterWarehouse.MultiLanguageName.GetNames()
				insertingWarehouses = append(insertingWarehouses, model.Warehouse{
					BaseModel: model.BaseModel{
						Uuid:       headquarterWarehouse.Uuid,
						CreateTime: headquarterWarehouse.CreateTime,
						UpdateTime: headquarterWarehouse.UpdateTime,
						DeleteTime: headquarterWarehouse.DeleteTime,
					},
					Name:                  multiLanguageName.ToJson(),
					MultiLanguageNameUuid: headquarterWarehouse.MultiLanguageNameUuid,
					Type:                  headquarterWarehouse.Type,
					Code:                  headquarterWarehouse.Code,
					Status:                headquarterWarehouse.Status,
					IsDefault:             headquarterWarehouse.IsDefault,
					ErpCode:               headquarterWarehouse.ErpCode,
					Contact:               headquarterWarehouse.Contact,
					Phone:                 headquarterWarehouse.Phone,
					Address:               headquarterWarehouse.Address,
					HeadquarterUuid:       headquarter.Uuid,
				})
			}
		}
		if len(insertingWarehouses) > 0 {
			err = tx.Model(&model.Warehouse{}).Create(&insertingWarehouses).Error
			if err != nil {
				return errors.WithMessage(err, "同步子店erp仓库或总部仓库失败")
			}
		}
		// 更新所有物品的warehouse_uuid为默认仓库Uuid
		var defaultWarehouse model.Warehouse
		tx.Model(&model.Warehouse{}).Where("erp_code = ?", defaultWarehouseErpCode).Scopes(repository.NotDeleted).Find(&defaultWarehouse)
		if err := tx.Model(&model.Material{}).Where("id > 0").Update("warehouse_uuid", defaultWarehouse.Uuid).Error; err != nil {
			return errors.WithMessage(errors.New("更新所有物品的warehouse_uuid为默认仓库Uuid失败"), err.Error())
		}
		return nil
	})

	// 添加多语言uuid到待翻译集合中
	if len(multiLanguageNameUuids) > 0 {
		if err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), multiLanguageNameUuids...); err != nil {
			logger.Logger.Error("仓库同步添加多语言uuid到待翻译集合中失败", zap.Error(err), zap.Any("multiLanguageNameUuids", multiLanguageNameUuids))
		}
	}
	return err
}

// SyncWarehouseItemStock 同步仓库物品库存
func (s *warehouseSrv) SyncWarehouseItemStock(ctx context.Context) error {
	company := ctx.GetCompany()
	if !company.IsOpenErp() {
		return errors.New("公司未开启erp")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	warehouseRepo := repository.NewWarehouseRepo(db)
	warehouses, err := warehouseRepo.Get(repository.ExcludeHeadquarter)
	if err != nil {
		return errors.WithMessage(err, "查询仓库列表失败")
	}

	// 获取所有material
	var materials []model.Material
	db.Model(&model.Material{}).Where("code != ''").Find(&materials)
	materialMap := make(map[string]model.Material)
	for _, material := range materials {
		materialMap[material.Code] = material
	}

	var insertingWarehouseItems []model.WarehouseItem
	for _, warehouse := range warehouses {
		stockList, err := erp.NewIErpSrv(s.dbm).GetMaterialStockNum(ctx, warehouse.ErpCode)
		if err != nil {
			return errors.WithMessage(err, "获取仓库物品库存数量失败")
		}
		// 仓库内物品
		var warehouseItems []model.WarehouseItem
		db.Model(&model.WarehouseItem{}).Where("warehouse_uuid = ? AND material_code != ''", warehouse.Uuid).Scopes(repository.NotDeleted).Find(&warehouseItems)
		warehouseItemMap := make(map[string]model.WarehouseItem)
		for _, warehouseItem := range warehouseItems {
			warehouseItemMap[warehouseItem.MaterialCode] = warehouseItem
		}

		materialConsumption, err := s.materialSrv.GetWarehouseItemConsumption(ctx, warehouse.Uuid)
		if err != nil {
			return errors.WithMessage(err, "获取仓库物品消耗量失败")
		}
		materialConsumptionMap := make(map[string]float64)
		for _, consumption := range materialConsumption.List {
			materialConsumptionMap[consumption.MaterialCode] = consumption.Consumption
		}

		for _, stock := range stockList {
			material, ok := materialMap[stock.ItemCode]
			if !ok {
				continue
			}
			// 仓库物品库存的计算，可能为负数
			stockNum := stock.ActualQty - materialConsumptionMap[stock.ItemCode]

			// 添加或修改仓库物品库存
			if warehouseItem, ok := warehouseItemMap[stock.ItemCode]; ok {
				if err := db.Model(&model.WarehouseItem{}).Where("uuid = ?", warehouseItem.Uuid).Update("stock", stockNum).Error; err != nil {
					return errors.WithMessage(err, "更新仓库物品库存失败")
				}
			} else {
				insertingWarehouseItems = append(insertingWarehouseItems, model.WarehouseItem{
					WarehouseUuid: warehouse.Uuid,
					MaterialUuid:  material.Uuid,
					MaterialCode:  material.Code,
					Stock:         stockNum,
					Valuation:     1.0, // 默认
				})
			}
		}
	}

	if len(insertingWarehouseItems) > 0 {
		err = db.Model(&model.WarehouseItem{}).Create(&insertingWarehouseItems).Error
		if err != nil {
			return errors.WithMessage(err, "创建仓库物品库存失败")
		}
	}

	return err
}

// CheckCodeExists 检查仓库编码是否存在
func (s *warehouseSrv) CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	exists, err := warehouseRepo.IsCodeExists(req.Code, req.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查仓库编码是否存在失败")
	}
	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}

// GetOtherOrgList 获取对方机构列表
func (s *warehouseSrv) GetOtherOrgList(ctx context.Context) (resp.OtherOrgListResp, error) {
	db := ctx.GetDB()
	supplierRepo := repository.NewSupplierRepo(db)

	// 获取供应商列表
	suppliers, err := supplierRepo.GetList(
		supplierRepo.OrderByCreateTime(true),
		supplierRepo.WhereNotDeleted(),
	)
	if err != nil {
		return resp.OtherOrgListResp{}, errors.WithMessage(err, "获取对方机构列表失败")
	}
	otherOrgs := make([]resp.OtherOrgResp, 0, len(suppliers))
	for _, supplier := range suppliers {
		if !supplier.IsNormalSupplier() && !supplier.IsHeadquartersSupplier() {
			continue
		}
		otherOrgs = append(otherOrgs, resp.OtherOrgResp{
			Name:          supplier.Name,
			Code:          fmt.Sprintf("%s:%s", OtherOrgCodeSupplierErpCode, supplier.ErpCode), // 供应商erp编码前缀
			IsHeadquarter: supplier.HeadquarterUuid > 0,                                        // 供应商是否总部机构
		})
	}

	// 获取公司列表
	companies, err := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).GetAllSubShopsAndHeadquarterListByCompanyUuid(ctx.GetCompanyUuid())
	if err != nil {
		return resp.OtherOrgListResp{}, errors.WithMessage(err, "获取公司列表失败")
	}
	for _, company := range companies {
		otherOrgs = append(otherOrgs, resp.OtherOrgResp{
			Name: company.Name,
			Code: fmt.Sprintf("%s:%s", OtherOrgCodeCompanyUuid, strconv.FormatUint(company.Uuid, 10)), // 公司uuid前缀
		})
	}

	return resp.OtherOrgListResp{List: otherOrgs}, nil
}

// GetWarehouseMaterialList 获取仓库物品列表
func (s *warehouseSrv) GetWarehouseMaterialList(ctx context.Context, req req.WarehouseMaterialListReq) (resp.WarehouseMaterialListResp, error) {
	db := ctx.GetDB()
	warehouseItemRepo := repository.NewWarehouseItemRepo(db)
	materialRepo := repository.NewMaterialRepo(db)

	// 构建查询选项
	var opts []repository.DBOption
	opts = append(opts, warehouseItemRepo.WhereWarehouseUuid(req.WarehouseUuid))

	// 查询仓库物品列表
	warehouseItemMap := make(map[uint64]model.WarehouseItem)
	warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(opts...)
	if err != nil {
		return resp.WarehouseMaterialListResp{}, errors.WithMessage(err, "查询仓库物品列表失败")
	}
	for _, warehouseItem := range warehouseItems {
		warehouseItemMap[warehouseItem.MaterialUuid] = *warehouseItem
	}

	// 构建查询选项
	var materialOpts []repository.DBOption
	materialOpts = append(materialOpts, repository.NewCommonRepo().WhereByStatus(uint(1)))
	materialOpts = append(materialOpts, repository.NotDeleted)

	// 获取物品列表
	paginatedMaterials, total, err := materialRepo.GetMaterialListWithPagination(
		req.PageNo,
		req.PageSize,
		materialOpts...,
	)

	// 获取所有物品UUID
	materialUuids := make([]uint64, 0, len(paginatedMaterials))
	for _, material := range paginatedMaterials {
		materialUuids = append(materialUuids, material.Uuid)
	}

	// 批量查询物品详情（包含多语言名称和单位信息）
	var materials []*model.Material
	if len(materialUuids) > 0 {
		materials, err = materialRepo.GetMaterialDetailByUuids(materialUuids)
		if err != nil {
			return resp.WarehouseMaterialListResp{}, errors.WithMessage(err, "查询物品详情失败")
		}
	}

	// 构建物品UUID到物品详情的映射
	materialMap := make(map[uint64]*model.Material)
	for _, material := range materials {
		materialMap[material.Uuid] = material
	}

	// 转换响应数据
	list := make([]resp.WarehouseMaterialInfo, 0, len(materials))
	for _, material := range materials {
		item, _ := warehouseItemMap[material.Uuid]

		// 构建物品单位信息
		units := make([]resp.MaterialUnitInfo, 0)
		var baseUnit resp.MaterialUnitInfo

		// 获取基准单位
		if material.Unit != nil {
			baseUnit = resp.MaterialUnitInfo{
				MaterialUnitUuid: material.Unit.Uuid,
				UnitUuid:         material.Unit.UnitUuid,
				UnitName: func() dto.LocaleResponse {
					if material.Unit.Unit != nil && material.Unit.Unit.MultiLanguageName.Uuid != 0 {
						return material.Unit.Unit.MultiLanguageName.GetNames()
					}
					return dto.LocaleResponse{}
				}(),
				ConversionRate: material.Unit.ConversionRate,
				IsDefault:      material.Unit.IsDefault,
			}
		}

		// 获取所有非基准单位
		if len(material.NotBaseUnitList) > 0 {
			for _, unit := range material.NotBaseUnitList {
				if unit.Unit != nil {
					unitInfo := resp.MaterialUnitInfo{
						MaterialUnitUuid: unit.Uuid,
						UnitUuid:         unit.UnitUuid,
						UnitName:         unit.Unit.MultiLanguageName.GetNames(),
						ConversionRate:   unit.ConversionRate,
						IsDefault:        unit.IsDefault,
					}
					units = append(units, unitInfo)
				}
			}
		}

		// 构建响应数据
		info := resp.WarehouseMaterialInfo{
			MaterialUuid:    material.Uuid,
			MaterialName:    material.MultiLanguageName.GetNames(),
			MaterialCode:    material.Code,
			MaterialBarcode: material.BarcodeValue,
			InternalCode:    material.InternalCode,
			BookedQuantity:  item.Stock, // 账面库存数量
			BaseUnit:        baseUnit,
			Units:           units,
			CategoryUuid:    material.CategoryUuid,
			Image:           material.GetImage(utils.GetBaseURL(ctx.GetGin().Request)),
			Status:          material.Status,
			DeleteTime:      material.DeleteTime,
		}
		list = append(list, info)
	}

	return resp.WarehouseMaterialListResp{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
