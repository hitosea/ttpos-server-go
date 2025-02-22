package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type ICallRepo interface {
	WhereStatus(status uint8) DBOption                                                     // 状态条件
	WhereC1Status(status uint8) DBOption                                                   // join表时，c1状态条件
	WhereC2IsNull() DBOption                                                               // join表时，c2 is null 条件
	WhereUuid(uuid uint64) DBOption                                                        // Uuid 条件
	WhereDeskUuid(uuid uint64) DBOption                                                    // 桌台uuid条件
	WhereDeskUuidByCallUuid(uuid uint64) DBOption                                          // 根据uuid找到桌台uuid，并以此为条件
	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.CustomerCall, int64, error) // 分页获取呼叫列表
	GetUnprocessedCallCount(opts ...DBOption) (int64, error)                               // 获取未处理呼叫数量
	Update(vars map[string]any, opts []DBOption) error                                     // 根据条件修改呼
}

func NewCallRepo(db *gorm.DB) ICallRepo {
	return NewCallRepoImpl(db)
}

type callRepo struct {
	db *gorm.DB
}

func NewCallRepoImpl(db *gorm.DB) ICallRepo {
	return &callRepo{db: db}
}

func (r *callRepo) PaginateGet(page, pageSize int, opts ...DBOption) ([]model.CustomerCall, int64, error) {
	var calls []model.CustomerCall
	var total int64
	tableName := model.CustomerCall{}.TableName()
	// 构建基础查询，查询同一个桌台最新的一条数据 // 注意 如果有2个同时间戳的，会返回2条数据
	builder := r.db.Table(tableName + " as c1").
		Select("c1.id, c1.uuid, c1.desk_uuid, c1.desk_no, c1.call_type, c1.status, c1.is_send").
		Joins("LEFT JOIN " + tableName + " as c2 ON c1.desk_uuid = c2.desk_uuid AND c1.create_time < c2.create_time")
	for _, opt := range opts {
		builder = opt(builder)
	}
	// 获取总记录数
	err := builder.Count(&total).Error
	if err != nil {
		return calls, 0, err
	}
	// 执行分页查询
	err = builder.Limit(pageSize).Offset((page - 1) * pageSize).Find(&calls).Error
	if err != nil {
		return calls, 0, err
	}
	return calls, total, nil
}

func (r *callRepo) GetUnprocessedCallCount(opts ...DBOption) (int64, error) {
	var total int64
	tableName := model.CustomerCall{}.TableName()
	// 构建基础查询，查询同一个桌台最新的一条数据 // 注意 如果有2个同时间戳的，会返回2条数据
	builder := r.db.Table(tableName + " as c1").
		Joins("LEFT JOIN " + tableName + " as c2 ON c1.desk_uuid = c2.desk_uuid AND c1.create_time < c2.create_time")
	for _, opt := range opts {
		builder = opt(builder)
	}
	// 获取总记录数
	err := builder.Count(&total).Error

	return total, err
}

func (r *callRepo) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *callRepo) WhereC1Status(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("c1.status = ?", status)
	}
}

func (r *callRepo) WhereC2IsNull() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("c2.desk_uuid IS NULL")
	}
}

func (r *callRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}
func (r *callRepo) WhereDeskUuidByCallUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		model.CustomerCall{}.TableName()
		return db.Where("desk_uuid = (?)",
			r.db.Table(model.CustomerCall{}.TableName()).Select("desk_uuid").Where("uuid = ?", uuid).Limit(1))
	}
}

func (r *callRepo) WhereDeskUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("desk_uuid = ?", uuid)
	}
}

func (r *callRepo) Update(vars map[string]any, opts []DBOption) error {
	db := r.db.Model(&model.CustomerCall{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}
