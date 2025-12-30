package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// IOrderItemRemarkRepo 单品备注原因仓库接口
type IOrderItemRemarkRepo interface {
	GetOrderItemRemarkList() ([]model.OrderItemRemark, error)                          // 获取单品备注原因列表
	UpdateOrderItemRemark(uuid uint64, orderItemRemark model.OrderItemRemark) error    // 更新单品备注原因
	CreateOrderItemRemark(orderItemRemark model.OrderItemRemark) (uint64, error)       // 创建单品备注原因
	DeleteOrderItemRemark(uuid uint64) error                                           // 删除单品备注原因
	GetOrderItemRemarkByUuid(uuid uint64) (*model.OrderItemRemark, error)              // 根据uuid获取单品备注原因
	GetOrderItemRemarks(opts ...repository.DBOption) ([]*model.OrderItemRemark, error) // 获取单品备注原因列表
	GetOrderItemRemarkListByUuids(uuids []uint64) ([]*model.OrderItemRemark, error)    // 根据uuid数组获取单品备注原因列表
	CountOrderItemRemark() (int64, error)                                              // 获取单品备注原因数量
}

// NewOrderItemRemarkRepo 创建新的单品备注原因仓库
func NewOrderItemRemarkRepo(db *gorm.DB) IOrderItemRemarkRepo {
	return NewOrderItemRemarkRepoImpl(db)
}

// NewOrderItemRemarkRepoImpl 创建新的单品备注原因仓库实现
func NewOrderItemRemarkRepoImpl(db *gorm.DB) *OrderItemRemarkRepoImpl {
	return &OrderItemRemarkRepoImpl{db: db}
}

type OrderItemRemarkRepoImpl struct {
	db *gorm.DB // 数据库连接
}

func (r *OrderItemRemarkRepoImpl) GetOrderItemRemarkListByUuids(uuids []uint64) ([]*model.OrderItemRemark, error) {
	// 获取商户UUID
	companyUuid := repository.GetCompanyUuid(r.db)
	if companyUuid == 0 {
		// 如果无法获取商户UUID，直接查询数据库
		return r.queryOrderItemRemarks(uuids)
	}

	// 检查是否启用对象存储缓存
	if !adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 未启用缓存，直接查询数据库
		return r.queryOrderItemRemarks(uuids)
	}

	// 使用对象存储模块缓存查询
	return r.getOrderItemRemarksWithCache(companyUuid, uuids)
}

// queryOrderItemRemarks 查询订单商品备注列表（包含预加载的关联数据）
// 这是一个私有方法，用于统一查询逻辑，避免代码重复
func (r *OrderItemRemarkRepoImpl) queryOrderItemRemarks(uuids []uint64) ([]*model.OrderItemRemark, error) {
	remarks, err := r.GetOrderItemRemarks(
		repository.CommonRepo.WhereInUuids(uuids),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.SortWithCreateTime("desc"),
		repository.CommonRepo.Preload(repository.WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return remarks, nil
}

// getOrderItemRemarksWithCache 使用对象存储模块缓存查询订单商品备注列表
func (r *OrderItemRemarkRepoImpl) getOrderItemRemarksWithCache(companyUuid uint64, uuids []uint64) ([]*model.OrderItemRemark, error) {
	if len(uuids) == 0 {
		return []*model.OrderItemRemark{}, nil
	}
	if companyUuid == 0 {
		return nil, errors.New("getOrderItemRemarksWithCache companyUuid cannot be 0")
	}

	// 构建批量查询的 keys
	keys := make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			keys = append(keys, persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeOrderItemRemark, uuid))
		}
	}

	if len(keys) == 0 {
		return []*model.OrderItemRemark{}, nil
	}

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.OrderItemRemark](cache.Global, 5*time.Minute)

	// 使用批量查询缓存
	batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.OrderItemRemark, error) {
		// 缓存未命中时，从数据库查询
		remarks, err := r.queryOrderItemRemarks(uuids)
		if err != nil {
			return nil, err
		}
		// 转换为 map[string]*model.OrderItemRemark
		result := make(map[string]*model.OrderItemRemark)
		for _, remark := range remarks {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeOrderItemRemark, remark.Uuid)
			result[key] = remark
		}
		return result, nil
	})

	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryOrderItemRemarks(uuids)
	}

	// 将批量查询结果转换为列表，保持原有顺序
	result := make([]*model.OrderItemRemark, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeOrderItemRemark, uuid)
			if remark, ok := batchResult[key]; ok && remark != nil {
				result = append(result, remark)
			}
		}
	}

	return result, nil
}

func (r *OrderItemRemarkRepoImpl) GetOrderItemRemarks(opts ...repository.DBOption) ([]*model.OrderItemRemark, error) {
	orderItemRemarks := make([]*model.OrderItemRemark, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&orderItemRemarks)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return orderItemRemarks, nil
}

// GetOrderItemRemarkList 获取单品备注原因列表，排除逻辑删除的单品备注原因
func (r *OrderItemRemarkRepoImpl) GetOrderItemRemarkList() ([]model.OrderItemRemark, error) {
	var orderItemRemarks []model.OrderItemRemark
	err := r.db.Model(&model.OrderItemRemark{}).
		Preload("MultiLanguageName").
		Where("delete_time = ?", 0).
		Order("create_time desc").
		Find(&orderItemRemarks).Error
	return orderItemRemarks, errors.WithMessage(err)
}

// UpdateOrderItemRemark 更新单品备注原因
func (r *OrderItemRemarkRepoImpl) UpdateOrderItemRemark(uuid uint64, orderItemRemark model.OrderItemRemark) error {
	if err := r.db.Model(&model.OrderItemRemark{}).Where("uuid = ?", uuid).Updates(orderItemRemark).Error; err != nil {
		return errors.WithMessage(err)
	}

	if err := r.db.Model(&orderItemRemark.MultiLanguageName).Where("uuid = ?", orderItemRemark.MultiLanguageNameUuid).Updates(orderItemRemark.MultiLanguageName).Error; err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// CreateOrderItemRemark 创建单品备注原因
func (r *OrderItemRemarkRepoImpl) CreateOrderItemRemark(orderItemRemark model.OrderItemRemark) (uint64, error) {
	// 创建多语言名称
	if err := r.db.Create(&orderItemRemark.MultiLanguageName).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	// 创建单品备注原因
	if err := r.db.Create(&orderItemRemark).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	return orderItemRemark.Uuid, nil
}

// DeleteOrderItemRemark 软删除单品备注原因
func (r *OrderItemRemarkRepoImpl) DeleteOrderItemRemark(uuid uint64) error {
	remark := model.OrderItemRemark{}
	r.db.Model(&model.OrderItemRemark{}).Where("uuid = ?", uuid).First(&remark)
	if err := r.db.Model(&model.OrderItemRemark{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
		return errors.WithMessage(err)
	}
	if remark.MultiLanguageNameUuid != 0 {
		if err := r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", remark.MultiLanguageNameUuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

func (r *OrderItemRemarkRepoImpl) GetOrderItemRemarkByUuid(uuid uint64) (*model.OrderItemRemark, error) {
	var remark model.OrderItemRemark
	err := r.db.Where("uuid = ?", uuid).Scopes(repository.NotDeleted).Preload("MultiLanguageName").First(&remark).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &remark, nil
}

func (r *OrderItemRemarkRepoImpl) CountOrderItemRemark() (int64, error) {
	var count int64
	err := r.db.Model(&model.OrderItemRemark{}).Where("delete_time = 0").Count(&count).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return count, nil
}
