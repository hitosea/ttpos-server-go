package service

import (
	"errors"
	"fmt"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"
)

// TakeoutDomainService 外卖域服务接口
type TakeoutDomainService interface {
	// GetByUuid 根据UUID获取外卖状态
	GetByUuid(ctx context.Context, uuid uint64) (*model.Takeout, error)

	// GetByPlatform 根据平台获取外卖状态
	GetByPlatform(ctx context.Context, platform string) (*model.Takeout, error)

	// CreatePlatformStatus 创建平台状态记录
	CreatePlatformStatus(ctx context.Context, platform string, enabled bool) (*model.Takeout, error)

	// UpdatePlatformStatus 更新平台状态
	UpdatePlatformStatus(ctx context.Context, uuid uint64, enabled bool) error

	// UpdatePlatformStatusByPlatform 通过平台名称更新平台状态
	UpdatePlatformStatusByPlatform(ctx context.Context, platform string, enabled bool) error

	// UpdatePlatformBoundStatus 更新平台绑定状态
	UpdatePlatformBoundStatus(ctx context.Context, uuid uint64, isBound bool) error

	// UpdatePlatformSkipStatus 更新平台跳过绑定状态
	UpdatePlatformSkipStatus(ctx context.Context, uuid uint64, skip bool) error

	// UpdatePlatformSkipStatusByPlatform 通过平台名称更新跳过绑定状态
	UpdatePlatformSkipStatusByPlatform(ctx context.Context, platform string, skip bool) error

	// UpdatePlatformMenu 更新平台菜单数据
	UpdatePlatformMenu(ctx context.Context, uuid uint64, menu interface{}) error

	// UpdatePlatformMenuByPlatform 通过平台名称更新平台菜单数据
	UpdatePlatformMenuByPlatform(ctx context.Context, platform string, menu interface{}) error

	// UpdatePlatformBindingLink 更新平台绑定链接
	UpdatePlatformBindingLink(ctx context.Context, uuid uint64, bindingLink string) error

	// UpdatePlatformBindingLinkByPlatform 通过平台名称更新平台绑定链接
	UpdatePlatformBindingLinkByPlatform(ctx context.Context, platform string, bindingLink string) error

	// GetAllPlatformStatus 获取所有平台状态
	GetAllPlatformStatus(ctx context.Context) ([]*model.Takeout, error)

	// ValidatePlatform 验证平台名称是否支持
	ValidatePlatform(platform string) error

	// IsPlatformEnabled 检查平台是否开启
	IsPlatformEnabled(ctx context.Context, uuid uint64) (bool, error)

	// IsPlatformBound 检查平台是否已绑定
	IsPlatformBound(ctx context.Context, uuid uint64) (bool, error)
}

// TakeoutDomainServiceImpl 外卖域服务实现
type TakeoutDomainServiceImpl struct {
	takeoutRepo persistence.ITakeoutRepository
}

// NewTakeoutDomainService 创建外卖域服务
func NewTakeoutDomainService(takeoutRepo persistence.ITakeoutRepository) TakeoutDomainService {
	return &TakeoutDomainServiceImpl{
		takeoutRepo: takeoutRepo,
	}
}

// GetByUuid 根据UUID获取外卖状态
func (s *TakeoutDomainServiceImpl) GetByUuid(ctx context.Context, uuid uint64) (*model.Takeout, error) {
	takeout, err := s.takeoutRepo.GetByUuid(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("获取UUID%d的外卖状态失败: %w", uuid, err)
	}

	return takeout, nil
}

// GetByPlatform 根据平台获取外卖状态
func (s *TakeoutDomainServiceImpl) GetByPlatform(ctx context.Context, platform string) (*model.Takeout, error) {
	if err := s.ValidatePlatform(platform); err != nil {
		return nil, err
	}

	takeout, err := s.takeoutRepo.GetByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("获取平台%s状态失败: %w", platform, err)
	}

	return takeout, nil
}

// CreatePlatformStatus 创建平台状态记录
func (s *TakeoutDomainServiceImpl) CreatePlatformStatus(ctx context.Context, platform string, enabled bool) (*model.Takeout, error) {
	if err := s.ValidatePlatform(platform); err != nil {
		return nil, err
	}

	// 检查是否已存在
	existing, err := s.takeoutRepo.GetByPlatform(ctx, platform)
	if err == nil && existing != nil && !existing.IsDelete() {
		return existing, nil
	}

	takeout := &model.Takeout{
		Uuid:     utils.MustGetID(),
		Platform: platform,
		Enabled:  enabled,
		IsBound:  false, // 新创建的默认为未绑定
	}

	if err := s.takeoutRepo.Create(ctx, takeout); err != nil {
		return nil, fmt.Errorf("创建平台%s状态失败: %w", platform, err)
	}

	return takeout, nil
}

// UpdatePlatformStatus 更新平台状态
func (s *TakeoutDomainServiceImpl) UpdatePlatformStatus(ctx context.Context, uuid uint64, enabled bool) error {
	if err := s.takeoutRepo.UpdateStatus(ctx, uuid, enabled); err != nil {
		return fmt.Errorf("更新UUID%d的平台状态失败: %w", uuid, err)
	}

	return nil
}

// UpdatePlatformBoundStatus 更新平台绑定状态
func (s *TakeoutDomainServiceImpl) UpdatePlatformBoundStatus(ctx context.Context, uuid uint64, isBound bool) error {
	if err := s.takeoutRepo.UpdateBoundStatus(ctx, uuid, isBound); err != nil {
		return fmt.Errorf("更新UUID%d的平台绑定状态失败: %w", uuid, err)
	}

	return nil
}

// UpdatePlatformSkipStatus 更新平台跳过绑定状态
func (s *TakeoutDomainServiceImpl) UpdatePlatformSkipStatus(ctx context.Context, uuid uint64, skip bool) error {
	if err := s.takeoutRepo.UpdateSkipStatus(ctx, uuid, skip); err != nil {
		return fmt.Errorf("更新UUID%d的平台跳过绑定状态失败: %w", uuid, err)
	}

	return nil
}

// UpdatePlatformSkipStatusByPlatform 通过平台名称更新跳过绑定状态
func (s *TakeoutDomainServiceImpl) UpdatePlatformSkipStatusByPlatform(ctx context.Context, platform string, skip bool) error {
	if err := s.ValidatePlatform(platform); err != nil {
		return err
	}

	if err := s.takeoutRepo.UpdateSkipStatusByPlatform(ctx, platform, skip); err != nil {
		return fmt.Errorf("更新平台%s的跳过绑定状态失败: %w", platform, err)
	}

	return nil
}

// GetAllPlatformStatus 获取所有平台状态
func (s *TakeoutDomainServiceImpl) GetAllPlatformStatus(ctx context.Context) ([]*model.Takeout, error) {
	takeouts, err := s.takeoutRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取所有平台状态失败: %w", err)
	}

	return takeouts, nil
}

// ValidatePlatform 验证平台名称是否支持
func (s *TakeoutDomainServiceImpl) ValidatePlatform(platform string) error {
	supportedPlatforms := []string{"grab", "lineman", "foodpanda", "shopeefood"}

	for _, p := range supportedPlatforms {
		if p == platform {
			return nil
		}
	}

	return errors.New("不支持的平台: " + platform + "，支持的平台: grab, lineman, foodpanda, shopeefood")
}

// IsPlatformEnabled 检查平台是否开启
func (s *TakeoutDomainServiceImpl) IsPlatformEnabled(ctx context.Context, uuid uint64) (bool, error) {
	takeout, err := s.GetByUuid(ctx, uuid)
	if err != nil {
		return false, err
	}

	return takeout.Enabled, nil
}

// IsPlatformBound 检查平台是否已绑定
func (s *TakeoutDomainServiceImpl) IsPlatformBound(ctx context.Context, uuid uint64) (bool, error) {
	takeout, err := s.GetByUuid(ctx, uuid)
	if err != nil {
		return false, err
	}

	return takeout.IsBound, nil
}

// UpdatePlatformStatusByPlatform 通过平台名称更新平台状态
func (s *TakeoutDomainServiceImpl) UpdatePlatformStatusByPlatform(ctx context.Context, platform string, enabled bool) error {
	if err := s.ValidatePlatform(platform); err != nil {
		return err
	}

	if err := s.takeoutRepo.UpdateStatusByPlatform(ctx, platform, enabled); err != nil {
		return fmt.Errorf("更新平台%s的状态失败: %w", platform, err)
	}
	return nil
}

// UpdatePlatformMenu 通过UUID更新平台菜单数据
func (s *TakeoutDomainServiceImpl) UpdatePlatformMenu(ctx context.Context, uuid uint64, menu interface{}) error {
	if err := s.takeoutRepo.UpdateMenu(ctx, uuid, menu); err != nil {
		return fmt.Errorf("更新平台%d的菜单数据失败: %w", uuid, err)
	}
	return nil
}

// UpdatePlatformMenuByPlatform 通过平台名称更新平台菜单数据
func (s *TakeoutDomainServiceImpl) UpdatePlatformMenuByPlatform(ctx context.Context, platform string, menu interface{}) error {
	if err := s.ValidatePlatform(platform); err != nil {
		return err
	}

	if err := s.takeoutRepo.UpdateMenuByPlatform(ctx, platform, menu); err != nil {
		return fmt.Errorf("更新平台%s的菜单数据失败: %w", platform, err)
	}
	return nil
}

// UpdatePlatformBindingLink 更新平台绑定链接
func (s *TakeoutDomainServiceImpl) UpdatePlatformBindingLink(ctx context.Context, uuid uint64, bindingLink string) error {
	if err := s.takeoutRepo.UpdateBindingLink(ctx, uuid, bindingLink); err != nil {
		return fmt.Errorf("更新UUID%d的绑定链接失败: %w", uuid, err)
	}

	return nil
}

// UpdatePlatformBindingLinkByPlatform 通过平台名称更新平台绑定链接
func (s *TakeoutDomainServiceImpl) UpdatePlatformBindingLinkByPlatform(ctx context.Context, platform string, bindingLink string) error {
	if err := s.ValidatePlatform(platform); err != nil {
		return err
	}

	if err := s.takeoutRepo.UpdateBindingLinkByPlatform(ctx, platform, bindingLink); err != nil {
		return fmt.Errorf("更新平台%s绑定链接失败: %w", platform, err)
	}

	return nil
}
