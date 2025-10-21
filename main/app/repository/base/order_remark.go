package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// IOrderRemarkRepo 整单备注仓库接口
type IOrderRemarkRepo interface {
	GetOrderRemarkList() ([]model.OrderRemark, error)                          // 获取整单备注列表
	UpdateOrderRemark(uuid uint64, orderRemark model.OrderRemark) error        // 更新整单备注
	CreateOrderRemark(orderRemark model.OrderRemark) (uint64, error)           // 创建整单备注
	DeleteOrderRemark(uuid uint64) error                                       // 删除整单备注
	DeleteOrderRemarks(uuids []uint64) error                                   // 批量删除整单备注
	ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error)               // 根据uuid数组验证整单备注是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
	GetOrderRemarkByUuid(uuid uint64) (*model.OrderRemark, error)              // 根据uuid获取整单备注
	GetOrderRemarks(opts ...repository.DBOption) ([]*model.OrderRemark, error) // 获取整单备注列表
	GetOrderRemarkListByUuids(uuids []uint64) ([]*model.OrderRemark, error)    // 根据uuid数组获取整单备注列表
	CountOrderRemark() (int64, error)                                          // 获取整单备注数量
}

// NewOrderRemarkRepo 创建新的整单备注仓库
func NewOrderRemarkRepo(db *gorm.DB) IOrderRemarkRepo {
	return NewOrderRemarkRepoImpl(db)
}

// NewOrderRemarkRepoImpl 创建新的整单备注仓库实现
func NewOrderRemarkRepoImpl(db *gorm.DB) *OrderRemarkRepoImpl {
	return &OrderRemarkRepoImpl{db: db}
}

type OrderRemarkRepoImpl struct {
	db *gorm.DB // 数据库连接
}

func (r *OrderRemarkRepoImpl) GetOrderRemarkListByUuids(uuids []uint64) ([]*model.OrderRemark, error) {
	remarks, err := r.GetOrderRemarks(
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

func (r *OrderRemarkRepoImpl) GetOrderRemarks(opts ...repository.DBOption) ([]*model.OrderRemark, error) {
	orderRemarks := make([]*model.OrderRemark, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&orderRemarks)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return orderRemarks, nil
}

// GetOrderRemarkList 获取整单备注列表，排除逻辑删除的整单备注
func (r *OrderRemarkRepoImpl) GetOrderRemarkList() ([]model.OrderRemark, error) {
	var orderRemarks []model.OrderRemark
	err := r.db.Model(&model.OrderRemark{}).
		Preload("MultiLanguageName").
		Where("delete_time = ?", 0).
		Order("create_time desc").
		Find(&orderRemarks).Error
	return orderRemarks, errors.WithMessage(err)
}

// UpdateOrderRemark 更新整单备注
func (r *OrderRemarkRepoImpl) UpdateOrderRemark(id uint64, orderRemark model.OrderRemark) error {
	if err := r.db.Model(&model.OrderRemark{}).Where("uuid = ?", id).Updates(orderRemark).Error; err != nil {
		return errors.WithMessage(err)
	}

	if err := r.db.Model(&orderRemark.MultiLanguageName).Where("uuid = ?", orderRemark.MultiLanguageNameUuid).Updates(orderRemark.MultiLanguageName).Error; err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// CreateOrderRemark 创建整单备注
func (r *OrderRemarkRepoImpl) CreateOrderRemark(orderRemark model.OrderRemark) (uint64, error) {
	// 创建多语言名称
	if err := r.db.Create(&orderRemark.MultiLanguageName).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	// 创建整单备注
	if err := r.db.Create(&orderRemark).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	return orderRemark.Uuid, nil
}

// DeleteOrderRemark 软删除整单备注
func (r *OrderRemarkRepoImpl) DeleteOrderRemark(uuid uint64) error {
	remark := model.OrderRemark{}
	r.db.Model(&model.OrderRemark{}).Where("uuid = ?", uuid).First(&remark)
	if err := r.db.Model(&model.OrderRemark{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
		return errors.WithMessage(err)
	}
	if remark.MultiLanguageNameUuid != 0 {
		if err := r.db.Model(&model.MultiLanguageName{}).Where("uuid = ?", remark.MultiLanguageNameUuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// ExistsByUuids 根据uuid数组验证整单备注是否存在，返回[uuid, 多语言名称UUID]数组和不存在的UUID列表
func (r *OrderRemarkRepoImpl) ExistsByUuids(uuids []uint64) ([][2]uint64, []uint64, error) {
	if len(uuids) == 0 {
		return [][2]uint64{}, []uint64{}, nil
	}

	var remarks []model.OrderRemark
	err := r.db.Where("uuid IN ? AND delete_time = 0", uuids).Find(&remarks).Error
	if err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	// 创建存在的UUID集合
	existMap := make(map[uint64]model.OrderRemark)
	for _, remark := range remarks {
		existMap[remark.Uuid] = remark
	}

	// 找出不存在的UUID
	notFound := make([]uint64, 0)
	for _, uuid := range uuids {
		if _, ok := existMap[uuid]; !ok {
			notFound = append(notFound, uuid)
		}
	}

	// 创建结果数组，每个元素是[uuid, 多语言名称UUID]
	result := make([][2]uint64, len(remarks))
	for i, remark := range remarks {
		result[i] = [2]uint64{remark.Uuid, remark.MultiLanguageNameUuid}
	}

	return result, notFound, nil
}

func (r *OrderRemarkRepoImpl) DeleteOrderRemarks(uuids []uint64) error {
	return r.db.Model(&model.OrderRemark{}).Where("uuid IN (?)", uuids).Update("delete_time", uint(time.Now().Unix())).Error
}

func (r *OrderRemarkRepoImpl) GetOrderRemarkByUuid(uuid uint64) (*model.OrderRemark, error) {
	var remark model.OrderRemark
	err := r.db.Where("uuid = ?", uuid).Scopes(repository.NotDeleted).Preload("MultiLanguageName").First(&remark).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &remark, nil
}

func (r *OrderRemarkRepoImpl) CountOrderRemark() (int64, error) {
	var count int64
	err := r.db.Model(&model.OrderRemark{}).Where("delete_time = 0").Count(&count).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return count, nil
}
