package service

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// ITakeoutSettingsSrv 外卖配置服务接口
type ITakeoutSettingsSrv interface {
	// 获取配置
	GetSettings(ctx context.Context, platform string) (*response.TakeoutSettingsResp, error)

	// 保存配置
	SaveSettings(ctx context.Context, req *request.TakeoutSettingsSaveReq) error
}

// takeoutSettingsSrv 外卖配置服务实现
type takeoutSettingsSrv struct {
	dbm *database.DBManager
}

// NewTakeoutSettingsSrv 创建外卖配置服务
func NewTakeoutSettingsSrv(dbm *database.DBManager) ITakeoutSettingsSrv {
	return &takeoutSettingsSrv{
		dbm: dbm,
	}
}

// GetSettings 获取配置
func (s *takeoutSettingsSrv) GetSettings(ctx context.Context, platform string) (*response.TakeoutSettingsResp, error) {
	db := ctx.GetDB()

	// 创建 Repository
	settingsRepo := persistence.NewTakeoutSettingsRepo(db)

	// 查询配置
	settings, err := settingsRepo.GetByPlatform(platform)
	if err != nil {
		return nil, errors.WithMessage(err, "查询配置失败")
	}

	// 如果不存在，返回默认配置
	if settings == nil {
		return &response.TakeoutSettingsResp{
			Platform:   platform,
			AutoAccept: false,
			MaxAmount:  0,
		}, nil
	}

	return &response.TakeoutSettingsResp{
		Uuid:       settings.Uuid,
		Platform:   settings.Platform,
		AutoAccept: settings.AutoAccept == 1,
		MaxAmount:  settings.MaxAmount,
	}, nil
}

// SaveSettings 保存配置
func (s *takeoutSettingsSrv) SaveSettings(ctx context.Context, req *request.TakeoutSettingsSaveReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 创建 Repository
	settingsRepo := persistence.NewTakeoutSettingsRepo(db)

	// 查询是否已存在配置
	existingSettings, err := settingsRepo.GetByPlatform(req.Platform)
	if err != nil {
		return errors.WithMessage(err, "查询配置失败")
	}

	// 转换布尔值为整型
	autoAccept := 0
	if req.AutoAccept {
		autoAccept = 1
	}

	if existingSettings == nil {
		// 创建新配置
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}

		settings := &model.TakeoutSettings{
			BaseModel: model.BaseModel{
				Uuid:       uuid,
				CreateTime: currentTime,
				UpdateTime: currentTime,
			},
			Platform:   req.Platform,
			AutoAccept: autoAccept,
			MaxAmount:  req.MaxAmount,
		}

		if err := settingsRepo.Create(settings); err != nil {
			return errors.WithMessage(err, "创建配置失败")
		}
	} else {
		// 更新配置
		updateData := map[string]interface{}{
			"auto_accept": autoAccept,
			"max_amount":  req.MaxAmount,
			"update_time": currentTime,
		}

		if err := settingsRepo.UpdateByMap(existingSettings.Uuid, updateData); err != nil {
			return errors.WithMessage(err, "更新配置失败")
		}
	}

	return nil
}
