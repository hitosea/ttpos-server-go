package repository

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomCardRepo interface {
	IProductBomCardQueryRepo
	CreateProductBomCard(productBomCard model.ProductBomCard) error
	CreateProductBomCardForSync(productBomCard model.ProductBomCard) error
	CreateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
	UpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
	CreateOrUpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
	UpdateProductBomCardErpCode(uuid uint64, erpCode string) error
	UpdateProductBomCardIsUsed(uuid uint64, isUsed int) error
	GetSubShopProductBomCardList(headquarterUuid uint64) ([]*model.ProductBomCard, error) // 获取子店中来自总部的成本卡列表
}

type IProductBomCardQueryRepo interface {
	GetProductBomCard(opts ...DBOption) (*model.ProductBomCard, error)
	GetProductBomCardList(opts ...DBOption) ([]*model.ProductBomCard, error)
	GetProductBomCardDetail(uuid uint64) (*model.ProductBomCard, error)
	GetProductBomCardName(uuid uint64) (dto.LocaleResponse, error)
}

type productBomCardRepoImpl struct {
	db *gorm.DB
}

func NewProductBomCardRepo(db *gorm.DB) IProductBomCardRepo {
	return &productBomCardRepoImpl{db: db}
}

func (r *productBomCardRepoImpl) GetProductBomCardList(opts ...DBOption) ([]*model.ProductBomCard, error) {
	var productBomCards []*model.ProductBomCard
	db := r.db

	db = db.Model(&model.ProductBomCard{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productBomCards)
	if result.Error != nil {
		return nil, result.Error
	}

	return productBomCards, nil
}

func (r *productBomCardRepoImpl) GetProductBomCard(opts ...DBOption) (*model.ProductBomCard, error) {
	var productBomCard model.ProductBomCard
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBomCard)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBomCard, nil
}

func (r *productBomCardRepoImpl) CreateProductBomCard(productBomCard model.ProductBomCard) error {
	productBomCard.SetNil()
	result := r.db.Create(&productBomCard)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

// 创建成本卡(同步erp数据)，同时创建multi_language_name和related_materials
func (r *productBomCardRepoImpl) CreateProductBomCardForSync(productBomCard model.ProductBomCard) error {
	result := r.db.Create(&productBomCard)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}
	return nil
}

func (r *productBomCardRepoImpl) CreateOrUpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	if productBomCardMaterial.Uuid == 0 {
		return r.CreateProductBomCardMaterial(productBomCardMaterial)
	} else {
		return r.UpdateProductBomCardMaterial(productBomCardMaterial)
	}
}

func (r *productBomCardRepoImpl) CreateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	productBomCardMaterial.SetNil()
	result := r.db.Create(&productBomCardMaterial)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

func (r *productBomCardRepoImpl) UpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	result := r.db.Model(&model.RelatedMaterial{}).Where("uuid = ?", productBomCardMaterial.Uuid).Updates(productBomCardMaterial)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}
	return nil
}

func (r *productBomCardRepoImpl) GetProductBomCardDetail(uuid uint64) (*model.ProductBomCard, error) {
	card, err := r.GetProductBomCard(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
		CommonRepo.Preload(
			WithPreload{
				Query: "RelatedMaterials.Material.MultiLanguageName",
			},
		),
		CommonRepo.Preload(
			WithPreload{
				Query: "RelatedMaterials.Material.NotBaseUnitList.Unit.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (r *productBomCardRepoImpl) GetProductBomCardName(uuid uint64) (dto.LocaleResponse, error) {
	card, err := r.GetProductBomCard(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	)
	if err != nil {
		return dto.LocaleResponse{}, err
	}
	return card.MultiLanguageName.GetNames(), nil
}

// 更新成本卡ERP编码
func (r *productBomCardRepoImpl) UpdateProductBomCardErpCode(uuid uint64, erpCode string) error {
	result := r.db.Model(&model.ProductBomCard{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"erp_code": erpCode,
	})
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}
	return nil
}

func (r *productBomCardRepoImpl) UpdateProductBomCardIsUsed(uuid uint64, isUsed int) error {
	if err := r.db.Model(&model.ProductBomCard{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
		"is_used": isUsed,
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 更新关联材料
	if err := r.db.Model(&model.RelatedMaterial{}).Where("related_uuid = ?", uuid).Updates(map[string]interface{}{
		"is_used": isUsed,
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productBomCardRepoImpl) GetProductBomCardByMaterialUuid(materialUuid uint64) (uint64, error) {
	var productBomCard model.ProductBomCard
	if err := r.db.Model(&model.ProductBomCard{}).Where("related_materials.material_uuid = ?", materialUuid).Preload("RelatedMaterials").First(&productBomCard).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return productBomCard.Uuid, nil
}

// 获取子店中来自总部的成本卡列表
func (r *productBomCardRepoImpl) GetSubShopProductBomCardList(headquarterUuid uint64) ([]*model.ProductBomCard, error) {
	var productBomCards []*model.ProductBomCard
	if err := r.db.Model(&model.ProductBomCard{}).
		Where("headquarter_uuid == 0"). // 只查询本店的成本卡
		Where("delete_time = 0").Where("is_used = 1"). // 只查询未失效的成本卡
		Preload("RelatedMaterials").Find(&productBomCards).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBomCards, nil
}
