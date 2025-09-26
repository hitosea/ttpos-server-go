package repository

import (
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 物品仓库接口
type IMaterialRepo interface {
	GetMaterialListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Material, int64, error)
	GetMaterialByUuid(uuid uint64, opts ...DBOption) (model.Material, error)
	GetMaterialByUuids(uuids []uint64, opts ...DBOption) ([]*model.Material, error)
	GetMaterialByCategoryUuid(categoryUuid uint64) ([]*model.Material, error)
	GetMaterialDetailByUuids(uuids []uint64) ([]*model.Material, error)
	GetMaterialDetailByUuid(uuid uint64) (*model.Material, error)
	CreateMaterial(material model.Material) (uint64, error)
	UpdateMaterialCode(uuid uint64, code string) error
	UpdateMaterial(material model.Material) error
	UpdateMaterialStatus(uuid uint64, status bool) error
	ClearMaterialBarcodeValue(uuid uint64) error // 清空物品条形码值
	ClearMaterialValuation(uuid uint64) error    // 清空物品估值率
	ClearMaterialInternalCode(uuid uint64) error // 清空物品内部编码
	DeleteMaterial(uuid uint64) error
	GetMaterialCategory(opts ...DBOption) (*model.MaterialCategory, error)
	GetMaterialCategoryByName(name string) (*model.MaterialCategory, error)
	GetMaterialCategoryByUuid(uuid uint64) (*model.MaterialCategory, error)
	GetMaterialCategoryByCode(code string) (*model.MaterialCategory, bool, error)
	GetMaterialCategoryByEnglishName(englishName string) (*model.MaterialCategory, bool, error)
	UpdateMaterialCategory(materialCategory model.MaterialCategory) error
	DeleteMaterialCategory(uuid uint64) error
	CreateMaterialCategory(materialCategory model.MaterialCategory) (uint64, error)
	GetMaterialCategoryList() ([]model.MaterialCategory, error)
	UpdateMaterialStatusBatch(uuids []uint64, status int) error          // 批量修改物品状态
	UpdateMaterialStockNum(materials []*model.Material) error            // 更新物品库存数量
	AddActualSaleNum(materialUuid uint64, saleNum float64) error         // 增加材料销量
	GetMaterialByErpCode(erpCode string) (*model.Material, error)        // 根据erp_code获取物品
	UpdateMaterialWarehouseUuid(uuid uint64, warehouseUuid uint64) error // 更新物品仓库uuid
	UpdateAllMaterialWarehouseUuid(warehouseUuid uint64) error           // 将所有物品的仓库uuid设置为指定仓库uuid

	CheckMultiLanguageNameExist(localeResponse dto.LocaleResponse) dto.LocaleResponse // 检查多语言名称是否存在
	GetCategoryUuidByNameOptimized(name string) (uint64, error)
	CheckBarcodeExist(barcode string, uuid uint64) bool                       // 检查条形码是否存在
	CheckMaterialInternalCodeExist(internalCode string, uuid uint64) bool     // 检查内部编码是否存在
	CheckMaterialCategoryCodeExist(code string, uuid uint64) bool             // 检查物品类别编码是否存在
	GetMaterialUuidsByCategoryUuids(categoryUuids []uint64) ([]uint64, error) // 根据分类UUID列表获取物品UUID列表
	GetMaterialUuidsByKeyword(keyword string) ([]uint64, error)               // 根据关键字获取物品UUID列表

	WithRelatedMaterialList() DBOption
}

// NewMaterialRepo 创建新的物品仓库
func NewMaterialRepo(db *gorm.DB) IMaterialRepo {
	return NewMaterialRepoImpl(db)
}

// NewMaterialRepoImpl 创建新的物品仓库实现
func NewMaterialRepoImpl(db *gorm.DB) *MaterialRepoImpl {
	return &MaterialRepoImpl{db: db}
}

type MaterialRepoImpl struct {
	db *gorm.DB // 数据库连接
}

func (r *MaterialRepoImpl) WithRelatedMaterialList() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("RelatedMaterialList")
	}
}

// GetMaterialListWithPagination 获取物品列表（分页）
func (r *MaterialRepoImpl) GetMaterialListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Material, int64, error) {
	var materials []model.Material
	var total int64

	// 构建查询
	query := r.db.Model(&model.Material{}).Where("delete_time = ?", 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "计算物品总数失败")
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&materials).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询物品列表失败")
	}

	return materials, total, nil
}

// GetMaterialByUuid 根据UUID获取物品详情
func (r *MaterialRepoImpl) GetMaterialByUuid(uuid uint64, opts ...DBOption) (model.Material, error) {
	var material model.Material

	query := r.db.Model(&model.Material{}).Where("uuid = ? AND delete_time = ?", uuid, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&material).Error; err != nil {
		return model.Material{}, errors.WithMessage(err, "查询物品详情失败")
	}

	return material, nil
}

// GetMaterialByUuids 根据UUIDs获取物品详情
func (r *MaterialRepoImpl) GetMaterialByUuids(uuids []uint64, opts ...DBOption) ([]*model.Material, error) {
	var materials []*model.Material

	query := r.db.Model(&model.Material{}).Where("uuid IN (?) AND delete_time = ?", uuids, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.Find(&materials).Error; err != nil {
		return nil, errors.WithMessage(err, "查询物品详情失败")
	}

	return materials, nil
}

func (r *MaterialRepoImpl) GetMaterialByCategoryUuid(categoryUuid uint64) ([]*model.Material, error) {
	var materials []*model.Material
	if err := r.db.Model(&model.Material{}).
		Preload("NotBaseUnitList.Unit").
		Preload("Unit.Unit").
		Preload("MultiLanguageName").
		Where("category_uuid = ?", categoryUuid).Where("delete_time = ?", 0).Find(&materials).Error; err != nil {
		return nil, errors.WithMessage(err, "查询物品失败")
	}
	return materials, nil
}

func (r *MaterialRepoImpl) GetMaterialDetailByUuid(uuid uint64) (*model.Material, error) {
	material, err := r.GetMaterialByUuid(uuid,
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "Category.MultiLanguageName",
			},
			WithPreload{
				Query: "Unit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "PurchaseUnit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "CostUnit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "NotBaseUnitList.Unit.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询物品详情失败")
	}
	return &material, nil
}

// GetMaterialDetailByUuids 根据UUIDs获取物品详情
func (r *MaterialRepoImpl) GetMaterialDetailByUuids(uuids []uint64) ([]*model.Material, error) {
	materials, err := r.GetMaterialByUuids(uuids,
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
			WithPreload{
				Query: "Category.MultiLanguageName",
			},
			WithPreload{
				Query: "Unit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "PurchaseUnit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "CostUnit.Unit.MultiLanguageName",
			},
			WithPreload{
				Query: "NotBaseUnitList.Unit.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询物品详情失败")
	}
	return materials, nil
}

// CreateMaterial 创建物品
func (r *MaterialRepoImpl) CreateMaterial(material model.Material) (uint64, error) {
	material.SetNil()
	if err := r.db.Create(&material).Error; err != nil {
		return 0, errors.WithMessage(err, "创建物品失败")
	}
	return material.Uuid, nil
}

func (r *MaterialRepoImpl) UpdateMaterialCode(uuid uint64, code string) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("code", code).Error; err != nil {
		return errors.WithMessage(err, "更新物品编码失败")
	}
	return nil
}

// UpdateMaterial 更新物品
func (r *MaterialRepoImpl) UpdateMaterial(material model.Material) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", material.Uuid).Updates(material).Error; err != nil {
		return errors.WithMessage(err, "更新物品失败")
	}
	return nil
}

// 关闭物品状态
func (r *MaterialRepoImpl) UpdateMaterialStatus(uuid uint64, status bool) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("status", status).Error; err != nil {
		return errors.WithMessage(err, "更新物品状态失败")
	}
	return nil
}

// DeleteMaterial 删除物品（软删除）
func (r *MaterialRepoImpl) DeleteMaterial(uuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "删除物品失败")
	}
	return nil
}

func (r *MaterialRepoImpl) GetMaterialCategoryByName(name string) (*model.MaterialCategory, error) {
	var materialCategory model.MaterialCategory

	query := r.db.Model(&model.MaterialCategory{}).Where("name = ? AND delete_time = ?", name, 0)

	if err := query.First(&materialCategory).Error; err != nil {
		return nil, errors.WithMessage(err, "根据名称查询物品类别失败")
	}

	return &materialCategory, nil
}

func (r *MaterialRepoImpl) CreateMaterialCategory(materialCategory model.MaterialCategory) (uint64, error) {
	if err := r.db.Create(&materialCategory).Error; err != nil {
		return 0, errors.WithMessage(err, "创建物品类别失败")
	}
	return materialCategory.Uuid, nil
}

func (r *MaterialRepoImpl) GetMaterialCategoryList() ([]model.MaterialCategory, error) {
	var materialCategories []model.MaterialCategory

	if err := r.db.Model(&model.MaterialCategory{}).Where("delete_time = ?", 0).Preload("MultiLanguageName").Order("sort ASC").Find(&materialCategories).Error; err != nil {
		return nil, errors.WithMessage(err, "获取物品类别列表失败")
	}

	return materialCategories, nil
}

func (r *MaterialRepoImpl) UpdateMaterialStatusBatch(uuids []uint64, status int) error {
	if err := r.db.Model(&model.Material{}).Where("uuid IN (?)", uuids).Update("status", status).Error; err != nil {
		return errors.WithMessage(err, "批量修改物品状态失败")
	}
	return nil
}

func (r *MaterialRepoImpl) ClearMaterialBarcodeValue(uuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("barcode_value", "").Error; err != nil {
		return errors.WithMessage(err, "清空物品条形码值失败")
	}
	return nil
}

func (r *MaterialRepoImpl) ClearMaterialValuation(uuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("valuation", 0).Error; err != nil {
		return errors.WithMessage(err, "清空物品估值率失败")
	}
	return nil
}

func (r *MaterialRepoImpl) UpdateMaterialStockNum(materials []*model.Material) error {
	if len(materials) == 0 {
		return nil
	}
	for _, material := range materials {
		if err := r.db.Model(&model.Material{}).Where("uuid = ?", material.Uuid).Update("stock_num", material.StockNum).Error; err != nil {
			return errors.WithMessage(err, "更新物品库存数量失败")
		}
	}
	return nil
}

func (r *MaterialRepoImpl) AddActualSaleNum(materialUuid uint64, saleNum float64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", materialUuid).Update("actual_sale_num", gorm.Expr("actual_sale_num + ?", saleNum)).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *MaterialRepoImpl) CheckMultiLanguageNameExist(localeResponse dto.LocaleResponse) dto.LocaleResponse {
	var result dto.LocaleResponse
	materialTable := r.db.Table("material").Name()
	multiLanguageNameTable := r.db.Table("multi_language_name").Name()

	// 定义语言字段映射
	languageFields := map[string]string{
		"zh":   "zh_name",
		"th":   "th_name",
		"en":   "en_name",
		"zhtw": "zh_tw_name",
		"ja":   "ja_name",
		"ko":   "ko_name",
		"my":   "my_name",
		"tr":   "tr_name",
		"sv":   "sv_name",
	}

	// 构建动态查询条件
	var conditions []string
	var args []interface{}

	for langKey, columnName := range languageFields {
		value := localeResponse.GetLocale(langKey)
		if value != "" {
			conditions = append(conditions, multiLanguageNameTable+"."+columnName+" = ?")
			args = append(args, value)
		}
	}

	// 如果没有任何条件，直接返回空结果
	if len(conditions) == 0 {
		return result
	}

	// 查询匹配的多语言名称记录
	var matchedRecords []model.MultiLanguageName
	err := r.db.Model(&model.Material{}).
		Select(multiLanguageNameTable+".*").
		Joins("JOIN "+multiLanguageNameTable+" ON "+materialTable+".multi_language_name_uuid = "+multiLanguageNameTable+".uuid").
		Where(materialTable+".delete_time = ?", constant.NotDeleted).
		Where("("+strings.Join(conditions, " OR ")+")", args...).
		Find(&matchedRecords).Error

	if err != nil {
		return result
	}

	// 检查每个匹配的记录，设置存在的语言名称
	for _, record := range matchedRecords {
		for langKey := range languageFields {
			inputValue := localeResponse.GetLocale(langKey)
			recordValue := record.GetNameByLang(langKey)

			// 如果输入的名称与数据库中的名称匹配，则标记为存在
			if inputValue != "" && inputValue == recordValue {
				result.SetLocale(langKey, inputValue)
			}
		}
	}

	return result
}

func (r *MaterialRepoImpl) GetCategoryUuidByNameOptimized(name string) (uint64, error) {
	var categorys []model.MaterialCategory
	err := r.db.Model(&model.MaterialCategory{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&categorys).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	// 在内存中查找匹配的商品规格
	for _, category := range categorys {
		// 然后检查多语言字段
		if category.MultiLanguageName.Uuid != 0 {
			names := category.MultiLanguageName.GetNames()
			if names.ZH == name || names.ZHTW == name || names.EN == name ||
				names.TH == name || names.MY == name || names.JA == name ||
				names.KO == name || names.TR == name || names.SV == name {
				return category.Uuid, nil
			}
		}
	}
	return 0, nil
}

func (r *MaterialRepoImpl) CheckBarcodeExist(barcode string, uuid uint64) bool {
	db := r.db.Model(&model.Material{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("barcode_value = ?", barcode).
		Where("barcode_value <> ?", "")
	if uuid != 0 {
		db = db.Where("uuid <> ?", uuid)
	}
	return db.First(&model.Material{}).Error == nil
}

func (r *MaterialRepoImpl) GetMaterialByErpCode(erpCode string) (*model.Material, error) {
	var material model.Material
	if err := r.db.Model(&model.Material{}).Preload("MultiLanguageName").Where("code = ?", erpCode).First(&material).Error; err != nil {
		return nil, errors.WithMessage(err, "根据erp_code获取物品失败")
	}
	return &material, nil
}

func (r *MaterialRepoImpl) ClearMaterialInternalCode(uuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("internal_code", "").Error; err != nil {
		return errors.WithMessage(err, "清空物品内部编码失败")
	}
	return nil
}

func (r *MaterialRepoImpl) CheckMaterialInternalCodeExist(internalCode string, uuid uint64) bool {
	db := r.db.Model(&model.Material{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("internal_code = ?", internalCode).
		Where("internal_code <> ?", "")
	if uuid != 0 {
		db = db.Where("uuid <> ?", uuid)
	}
	return db.First(&model.Material{}).Error == nil
}

func (r *MaterialRepoImpl) GetMaterialCategoryByUuid(uuid uint64) (*model.MaterialCategory, error) {
	var materialCategory model.MaterialCategory
	if err := r.db.Model(&model.MaterialCategory{}).Preload("MultiLanguageName").Where("uuid = ?", uuid).First(&materialCategory).Error; err != nil {
		return nil, errors.WithMessage(err, "根据UUID获取物品分类失败")
	}
	return &materialCategory, nil
}

func (r *MaterialRepoImpl) CheckMaterialCategoryCodeExist(code string, uuid uint64) bool {
	db := r.db.Model(&model.MaterialCategory{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("code = ?", code).
		Where("code <> ?", "")
	if uuid != 0 {
		db = db.Where("uuid <> ?", uuid)
	}
	return db.First(&model.MaterialCategory{}).Error == nil
}

func (r *MaterialRepoImpl) UpdateMaterialCategory(materialCategory model.MaterialCategory) error {
	if err := r.db.Model(&model.MaterialCategory{}).Where("uuid = ?", materialCategory.Uuid).Updates(map[string]any{
		"name": materialCategory.Name,
		"code": materialCategory.Code,
	}).Error; err != nil {
		return errors.WithMessage(err, "更新物品类别失败")
	}
	return nil
}

func (r *MaterialRepoImpl) DeleteMaterialCategory(uuid uint64) error {
	if err := r.db.Model(&model.MaterialCategory{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "删除物品类别失败")
	}
	return nil
}

func (r *MaterialRepoImpl) GetMaterialCategory(opts ...DBOption) (*model.MaterialCategory, error) {
	var materialCategory model.MaterialCategory
	db := r.db.Model(&model.MaterialCategory{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&materialCategory).Error; err != nil {
		return nil, errors.WithMessage(err, "根据条件获取物品分类失败")
	}
	return &materialCategory, nil
}

// 根据编码获取物品分类
func (r *MaterialRepoImpl) GetMaterialCategoryByCode(code string) (*model.MaterialCategory, bool, error) {
	materialCategory, err := r.GetMaterialCategory(
		CommonRepo.WhereByCode(code),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, false, nil
		}
		return nil, false, errors.WithMessage(err, "根据编码获取物品分类失败")
	}
	return materialCategory, true, nil
}

// 根据英文名称获取物品分类
func (r *MaterialRepoImpl) GetMaterialCategoryByEnglishName(englishName string) (*model.MaterialCategory, bool, error) {
	var materialCategory model.MaterialCategory
	// materialCategory表join multi_language_name表，where multi_language_name.en_name = englishName
	if err := r.db.Model(&model.MaterialCategory{}).Joins("MultiLanguageName").Where("multi_language_name.en_name = ?", englishName).Where("material_category.delete_time = ?", 0).First(&materialCategory).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, false, nil
		}
		return nil, false, errors.WithMessage(err, "根据英文名称获取物品分类失败")
	}
	return &materialCategory, true, nil
}

func (r *MaterialRepoImpl) UpdateMaterialWarehouseUuid(uuid uint64, warehouseUuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("warehouse_uuid", warehouseUuid).Error; err != nil {
		return errors.WithMessage(err, "更新物品仓库uuid失败")
	}
	return nil
}

func (r *MaterialRepoImpl) UpdateAllMaterialWarehouseUuid(warehouseUuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("delete_time = ?", 0).Update("warehouse_uuid", warehouseUuid).Error; err != nil {
		return errors.WithMessage(err, "更新所有物品仓库uuid失败")
	}
	return nil
}

func (r *MaterialRepoImpl) GetMaterialUuidsByCategoryUuids(categoryUuids []uint64) ([]uint64, error) {
	var uuids []uint64
	if err := r.db.Model(&model.Material{}).Where("category_uuid IN (?)", categoryUuids).Where("delete_time = ?", 0).Pluck("uuid", &uuids).Error; err != nil {
		return nil, errors.WithMessage(err, "根据分类UUID列表获取物品UUID列表失败")
	}
	return uuids, nil
}

func (r *MaterialRepoImpl) GetMaterialUuidsByKeyword(keyword string) ([]uint64, error) {
	var uuids []uint64
	if err := r.db.Model(&model.Material{}).Where("name LIKE ? OR code LIKE ? OR barcode_value LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").Where("delete_time = ?", 0).Pluck("uuid", &uuids).Error; err != nil {
		return nil, errors.WithMessage(err, "根据关键字获取物品UUID列表失败")
	}
	return uuids, nil
}
