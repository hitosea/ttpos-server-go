package repository

import "gorm.io/gorm"

type Where func(*gorm.DB) *gorm.DB
type With func(*gorm.DB) *gorm.DB
type DBOption func(*gorm.DB) *gorm.DB

// WithPreload 预加载
type WithPreload struct {
	Query string
	Args  []interface{}
}

// ICommonRepo 公共仓库接口
type ICommonRepo interface {
	WhereByID(id uint) DBOption                                     // 根据ID查询
	WhereByStatus(status uint) DBOption                             // 根据状态查询
	WhereByIsShowCashier(isShowCashier uint) DBOption               // 根据是否显示收银机查询
	WhereBySoftDelete() DBOption                                    // 根据软删除查询
	WhereByOrderNo(orderNo string) DBOption                         // 根据订单编号查询
	WhereByBillType(billType uint) DBOption                         // 根据账单类型查询
	WhereByIsHide(isHide bool) DBOption                             // 根据是否隐藏查询
	WhereLikeByName(name string) DBOption                           // 根据名称查询
	WhereBetweenByCreateTime(startTime uint, endTime uint) DBOption // 根据创建时间查询
	SortWithID(order string) DBOption                               // 根据ID排序
	SortWithSort(order string) DBOption                             // 根据Order By排序
	SortWithIsSpecial(order string) DBOption                        // 根据是否特殊排序
	Preload(preloads ...WithPreload) DBOption                       // 预加载
}

// commonRepo 公共仓库实现
type commonRepo struct{}

// NewCommonRepo 创建新的公共仓库
func NewCommonRepo() ICommonRepo {
	return NewCommonRepoImpl()
}

// NewCommonRepoImpl 创建新的公共仓库实现
func NewCommonRepoImpl() ICommonRepo {
	return &commonRepo{}
}

// WhereByID 根据ID查询
func (r *commonRepo) WhereByID(id uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

// WhereByStatus 根据状态查询
func (r *commonRepo) WhereByStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereByIsShowCashier 根据是否显示收银机查询
func (r *commonRepo) WhereByIsShowCashier(isShowCashier uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_cashier = ?", isShowCashier)
	}
}

// WhereBySoftDelete 根据软删除查询
func (r *commonRepo) WhereBySoftDelete() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = 0")
	}
}

// WhereByOrderNo 根据订单编号查询
func (r *commonRepo) WhereByOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_no = ?", orderNo)
	}
}

// WhereByBillType 根据账单类型查询
func (r *commonRepo) WhereByBillType(billType uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bill_type = ?", billType)
	}
}

// WhereByIsHide 根据是否隐藏查询
func (r *commonRepo) WhereByIsHide(isHide bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if isHide {
			return db.Where("hide_bill_time > 0")
		}
		return db.Where("hide_bill_time = 0")
	}
}

// WhereLikeByName 根据名称查询
func (r *commonRepo) WhereLikeByName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

// WhereBetweenByCreateTime 根据创建时间查询
func (r *commonRepo) WhereBetweenByCreateTime(startTime uint, endTime uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// SortWithID 根据ID排序
func (r *commonRepo) SortWithID(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("id " + order)
	}
}

// SortWithSort 根据Order By排序
func (r *commonRepo) SortWithSort(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("sort " + order)
	}
}

// SortWithIsSpecial 根据是否特殊排序
func (r *commonRepo) SortWithIsSpecial(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("is_special " + order)
	}
}

// Preload 预加载
func (r *commonRepo) Preload(preloads ...WithPreload) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(preload.Query, preload.Args...)
		}
		return db
	}
}
