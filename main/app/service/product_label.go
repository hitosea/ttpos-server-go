package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// IProductLabelSrv 商品标签服务接口
type IProductLabelSrv interface {
	GetProductLabelList(ctx context.Context) (*resp.ProductLabelListResp, error)
	AddProductLabel(ctx context.Context, req req.ProductLabelAddReq) error
	EditProductLabel(ctx context.Context, req req.ProductLabelEditReq) error
	DeleteProductLabel(ctx context.Context, req req.ProductLabelDeleteReq) error
}

// NewProductLabelSrv 创建商品标签服务
func NewProductLabelSrv(dbm *database.DBManager) IProductLabelSrv {
	return NewProductLabelSrvImpl(dbm)
}

// NewProductLabelSrvImpl 创建商品标签服务实现
func NewProductLabelSrvImpl(dbm *database.DBManager) *ProductLabelSrvImpl {
	return &ProductLabelSrvImpl{
		dbm: dbm,
	}
}

type ProductLabelSrvImpl struct {
	dbm *database.DBManager // 数据库管理器
}

// GetProductLabelList 获取商品标签列表
func (s *ProductLabelSrvImpl) GetProductLabelList(ctx context.Context) (*resp.ProductLabelListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productLabelRepo := repository.NewProductLabelRepo(db)

	// 查询所有标签，预加载关联的商品和统计数量
	labels, err := productLabelRepo.GetProductLabelList(
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "ProductPackages",
				Args: []interface{}{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
		),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "ProductPackages.MultiLanguageName",
		}),
		productLabelRepo.OrderByCreateTime("DESC"),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 组装响应数据
	result := make([]resp.ProductLabelDetail, 0)
	for _, label := range labels {
		detail := resp.ProductLabelDetail{
			Uuid:                label.Uuid,
			Name:                label.Name,
			Style:               label.Style,
			IsShowCashier:       label.IsShowCashier,
			IsShowTablet:        label.IsShowTablet,
			IsShowAssistant:     label.IsShowAssistant,
			IsShowH5:            label.IsShowH5,
			IsShowDelivery:      label.IsShowDelivery,
			IsShowMenu:          label.IsShowMenu,
			ProductPackageCount: len(label.ProductPackages),
			CreateTime:          label.CreateTime,
			ProductPackages:     []resp.ProductLabelPackageItem{},
		}

		// 组装关联的商品列表
		for _, pkg := range label.ProductPackages {
			detail.ProductPackages = append(detail.ProductPackages, resp.ProductLabelPackageItem{
				Uuid:       pkg.Uuid,
				LocaleName: pkg.MultiLanguageName.GetNames(),
			})
		}

		result = append(result, detail)
	}

	return &resp.ProductLabelListResp{
		List: result,
	}, nil
}

// AddProductLabel 添加商品标签
func (s *ProductLabelSrvImpl) AddProductLabel(ctx context.Context, req req.ProductLabelAddReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 验证请求
	if err := req.Validate(); err != nil {
		return errors.WithMessage(err)
	}

	// 创建事务
	return db.Transaction(func(tx *gorm.DB) error {
		productLabelRepo := repository.NewProductLabelRepo(tx)

		// 创建标签（Uuid 由 BaseModel.BeforeCreate 自动生成）
		label := model.ProductLabel{
			Name:            req.Name,
			Style:           req.Style,
			IsShowCashier:   req.IsShowCashier,
			IsShowTablet:    req.IsShowTablet,
			IsShowAssistant: req.IsShowAssistant,
			IsShowH5:        req.IsShowH5,
			IsShowDelivery:  req.IsShowDelivery,
			IsShowMenu:      req.IsShowMenu,
		}

		uuid, err := productLabelRepo.CreateProductLabel(label)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 更新关联的商品
		if len(req.ProductPackageUuids) > 0 {
			err := productLabelRepo.UpdateProductPackageLabelRelation(req.ProductPackageUuids, uuid)
			if err != nil {
				return errors.WithMessage(err)
			}
		}

		return nil
	})
}

// EditProductLabel 编辑商品标签
func (s *ProductLabelSrvImpl) EditProductLabel(ctx context.Context, req req.ProductLabelEditReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 验证请求
	if err := req.Validate(); err != nil {
		return errors.WithMessage(err)
	}

	// 创建事务
	return db.Transaction(func(tx *gorm.DB) error {
		productLabelRepo := repository.NewProductLabelRepo(tx)

		// 检查标签是否存在
		existLabel, err := productLabelRepo.GetProductLabelByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if existLabel == nil {
			return errors.New("标签不存在")
		}

		// 更新标签
		label := model.ProductLabel{
			BaseModel: model.BaseModel{
				Uuid: req.Uuid,
			},
			Name:            req.Name,
			Style:           req.Style,
			IsShowCashier:   req.IsShowCashier,
			IsShowTablet:    req.IsShowTablet,
			IsShowAssistant: req.IsShowAssistant,
			IsShowH5:        req.IsShowH5,
			IsShowDelivery:  req.IsShowDelivery,
			IsShowMenu:      req.IsShowMenu,
		}

		err = productLabelRepo.UpdateProductLabel(label)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 清除原有的关联关系
		err = productLabelRepo.ClearProductPackageLabelRelation(req.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 更新新的关联商品
		if len(req.ProductPackageUuids) > 0 {
			err := productLabelRepo.UpdateProductPackageLabelRelation(req.ProductPackageUuids, req.Uuid)
			if err != nil {
				return errors.WithMessage(err)
			}
		}

		return nil
	})
}

// DeleteProductLabel 删除商品标签
func (s *ProductLabelSrvImpl) DeleteProductLabel(ctx context.Context, req req.ProductLabelDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 创建事务
	return db.Transaction(func(tx *gorm.DB) error {
		productLabelRepo := repository.NewProductLabelRepo(tx)

		// 检查标签是否存在
		existLabel, err := productLabelRepo.GetProductLabelByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if existLabel == nil {
			return errors.New("标签不存在")
		}

		// 清除商品与标签的关联关系
		err = productLabelRepo.ClearProductPackageLabelRelation(req.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 删除标签
		err = productLabelRepo.DeleteProductLabel(req.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}

		return nil
	})
}
