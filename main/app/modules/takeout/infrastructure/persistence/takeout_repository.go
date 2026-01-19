package persistence

import (
	"encoding/json"
	"time"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// ITakeoutRepository 外卖仓储接口
type ITakeoutRepository interface {
	// Create 创建外卖平台状态记录
	Create(ctx context.Context, takeout *model.Takeout) error

	// GetByUuid 根据UUID获取外卖状态
	GetByUuid(ctx context.Context, uuid uint64) (*model.Takeout, error)

	// GetByPlatform 根据平台获取外卖状态
	GetByPlatform(ctx context.Context, platform string) (*model.Takeout, error)

	// GetAll 获取所有平台状态
	GetAll(ctx context.Context) ([]*model.Takeout, error)

	// UpdateStatus 更新平台状态
	UpdateStatus(ctx context.Context, uuid uint64, enabled bool) error

	// UpdateStatusByPlatform 通过平台名称更新平台状态
	UpdateStatusByPlatform(ctx context.Context, platform string, enabled bool) error

	// UpdateBoundStatus 更新绑定状态
	UpdateBoundStatus(ctx context.Context, uuid uint64, isBound bool) error

	// UpdateSkipStatus 更新跳过绑定状态
	UpdateSkipStatus(ctx context.Context, uuid uint64, skip bool) error

	// UpdateSkipStatusByPlatform 通过平台名称更新跳过绑定状态
	UpdateSkipStatusByPlatform(ctx context.Context, platform string, skip bool) error

	// UpdateMenu 更新平台菜单数据
	UpdateMenu(ctx context.Context, uuid uint64, menu interface{}) error

	// UpdateMenuByPlatform 通过平台名称更新平台菜单数据
	UpdateMenuByPlatform(ctx context.Context, platform string, menu interface{}) error

	// UpdateBindingLink 更新平台绑定链接
	UpdateBindingLink(ctx context.Context, uuid uint64, bindingLink string) error

	// UpdateBindingLinkByPlatform 通过平台名称更新平台绑定链接
	UpdateBindingLinkByPlatform(ctx context.Context, platform string, bindingLink string) error

	// UpdateTtposMenuByPlatform 通过平台名称更新 TTPOS 导出的菜单数据
	UpdateTtposMenuByPlatform(ctx context.Context, platform string, ttposMenu interface{}) error

	// Delete 删除平台状态记录
	Delete(ctx context.Context, uuid uint64) error

	// UpdateImportStatusByPlatform 通过平台名称更新导入状态
	UpdateImportStatusByPlatform(ctx context.Context, platform string, status int8, menu interface{}) error
}

// takeoutRepositoryImpl 外卖仓储实现
type takeoutRepositoryImpl struct {
	db *gorm.DB
}

// NewTakeoutRepository 创建外卖仓储
func NewTakeoutRepository(db *gorm.DB) ITakeoutRepository {
	return &takeoutRepositoryImpl{
		db: db,
	}
}

// Create 创建外卖平台状态记录
func (r *takeoutRepositoryImpl) Create(ctx context.Context, takeout *model.Takeout) error {
	db := ctx.GetDB()
	return db.Create(takeout).Error
}

// GetByUuid 根据UUID获取外卖状态
func (r *takeoutRepositoryImpl) GetByUuid(ctx context.Context, uuid uint64) (*model.Takeout, error) {
	db := ctx.GetDB()
	var takeout model.Takeout

	if err := db.Where("uuid = ? AND delete_time = ?", uuid, 0).First(&takeout).Error; err != nil {
		return nil, err
	}

	return &takeout, nil
}

// GetByPlatform 根据平台获取外卖状态
func (r *takeoutRepositoryImpl) GetByPlatform(ctx context.Context, platform string) (*model.Takeout, error) {
	db := ctx.GetDB()
	var takeout model.Takeout

	if err := db.Where("platform = ? AND delete_time = ?", platform, 0).First(&takeout).Error; err != nil {
		return nil, err
	}

	return &takeout, nil
}

// GetAll 获取所有平台状态
func (r *takeoutRepositoryImpl) GetAll(ctx context.Context) ([]*model.Takeout, error) {
	db := ctx.GetDB()
	var takeouts []*model.Takeout

	if err := db.Where("delete_time = ?", 0).Find(&takeouts).Error; err != nil {
		return nil, err
	}

	return takeouts, nil
}

// UpdateStatus 更新平台状态
func (r *takeoutRepositoryImpl) UpdateStatus(ctx context.Context, uuid uint64, enabled bool) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Updates(map[string]interface{}{
			"enabled":     enabled,
			"update_time": time.Now().Unix(),
		}).Error
}

// UpdateBoundStatus 更新绑定状态
func (r *takeoutRepositoryImpl) UpdateBoundStatus(ctx context.Context, uuid uint64, isBound bool) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Updates(map[string]interface{}{
			"is_bound":    isBound,
			"update_time": time.Now().Unix(),
		}).Error
}

// UpdateSkipStatus 更新跳过绑定状态
func (r *takeoutRepositoryImpl) UpdateSkipStatus(ctx context.Context, uuid uint64, skip bool) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Updates(map[string]interface{}{
			"skip":        skip,
			"update_time": time.Now().Unix(),
		}).Error
}

// UpdateSkipStatusByPlatform 通过平台名称更新跳过绑定状态
func (r *takeoutRepositoryImpl) UpdateSkipStatusByPlatform(ctx context.Context, platform string, skip bool) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("platform = ? AND delete_time = ?", platform, 0).
		Updates(map[string]interface{}{
			"skip":        skip,
			"update_time": time.Now().Unix(),
		}).Error
}

// UpdateMenu 更新平台菜单数据
func (r *takeoutRepositoryImpl) UpdateMenu(ctx context.Context, uuid uint64, menu interface{}) error {
	db := ctx.GetDB()

	// 将菜单结构体序列化为 JSON
	menuJSON, err := json.Marshal(menu)
	if err != nil {
		return err
	}

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Updates(map[string]interface{}{
			"menu":        string(menuJSON),
			"update_time": time.Now().Unix(),
		}).Error
}

// Delete 删除平台状态记录
func (r *takeoutRepositoryImpl) Delete(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Update("delete_time", time.Now().Unix()).Error
}

// UpdateStatusByPlatform 通过平台名称更新平台状态
func (r *takeoutRepositoryImpl) UpdateStatusByPlatform(ctx context.Context, platform string, enabled bool) error {
	db := ctx.GetDB()
	return db.Model(&model.Takeout{}).Where("platform = ? AND delete_time = ?", platform, 0).Updates(map[string]interface{}{
		"enabled":     enabled,
		"update_time": time.Now().Unix(),
	}).Error
}

// UpdateMenuByPlatform 通过平台名称更新平台菜单数据
func (r *takeoutRepositoryImpl) UpdateMenuByPlatform(ctx context.Context, platform string, menu interface{}) error {
	db := ctx.GetDB()

	// 将菜单结构体序列化为 JSON
	menuJSON, err := json.Marshal(menu)
	if err != nil {
		return err
	}

	return db.Model(&model.Takeout{}).Where("platform = ? AND delete_time = ?", platform, 0).Updates(map[string]interface{}{
		"menu":        string(menuJSON),
		"update_time": time.Now().Unix(),
	}).Error
}

// UpdateBindingLink 更新平台绑定链接
func (r *takeoutRepositoryImpl) UpdateBindingLink(ctx context.Context, uuid uint64, bindingLink string) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Updates(map[string]interface{}{
			"binding_link": bindingLink,
			"update_time":  time.Now().Unix(),
		}).Error
}

// UpdateBindingLinkByPlatform 通过平台名称更新平台绑定链接
func (r *takeoutRepositoryImpl) UpdateBindingLinkByPlatform(ctx context.Context, platform string, bindingLink string) error {
	db := ctx.GetDB()

	return db.Model(&model.Takeout{}).
		Where("platform = ? AND delete_time = ?", platform, 0).
		Updates(map[string]interface{}{
			"binding_link": bindingLink,
			"update_time":  time.Now().Unix(),
		}).Error
}

// UpdateTtposMenuByPlatform 通过平台名称更新 TTPOS 导出的菜单数据
func (r *takeoutRepositoryImpl) UpdateTtposMenuByPlatform(ctx context.Context, platform string, ttposMenu interface{}) error {
	db := ctx.GetDB()

	var ttposMenuJSON []byte
	var err error

	// 判断输入类型，避免重复序列化
	switch v := ttposMenu.(type) {
	case string:
		// 如果已经是 JSON 字符串，直接使用
		ttposMenuJSON = []byte(v)
	case []byte:
		// 如果是字节数组，直接使用
		ttposMenuJSON = v
	default:
		// 如果是对象，需要序列化
		ttposMenuJSON, err = json.Marshal(ttposMenu)
		if err != nil {
			return err
		}
	}

	return db.Model(&model.Takeout{}).
		Where("platform = ? AND delete_time = ?", platform, 0).
		Updates(map[string]interface{}{
			"ttpos_menu":  string(ttposMenuJSON),
			"update_time": time.Now().Unix(),
		}).Error
}

// UpdateImportStatusByPlatform 通过平台名称更新导入状态
func (r *takeoutRepositoryImpl) UpdateImportStatusByPlatform(ctx context.Context, platform string, status int8, menu interface{}) error {
	db := ctx.GetDB()

	updates := map[string]interface{}{
		"import_status": status,
		"update_time":   time.Now().Unix(),
	}

	if menu != nil {
		menuJSON, err := json.Marshal(menu)
		if err != nil {
			return err
		}
		updates["menu"] = string(menuJSON)
	}

	if status == model.ImportStatusSuccess {
		updates["is_bound"] = true
	}

	return db.Model(&model.Takeout{}).
		Where("platform = ? AND delete_time = ?", platform, 0).
		Updates(updates).Error
}
