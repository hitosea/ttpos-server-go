package service

import (
	"fmt"

	"gorm.io/gorm"

	"jjjshop-server-go/app/model"
)

type InitService struct {
	db *gorm.DB
}

func NewInitService(db *gorm.DB) *InitService {
	return &InitService{db: db}
}

func (s *InitService) InitDatabase() error {
	// 执行数据库迁移
	err := s.db.AutoMigrate(
		// 用户相关模型
		&model.App{},            // 应用表
		&model.Supplier{},       // 供应商表
		&model.User{},           // 商家用户表
		&model.BindRecord{},     // 商家设备绑定记录表
		&model.LoginLog{},       // 商家用户绑定记录
		&model.ShopRole{},       //  商家用户角色表
		&model.ShopAccess{},     //  商家权限表
		&model.UserRole{},       //  商家用户角色
		&model.ShopRoleAccess{}, //  商家角色权限
		&model.ShopOptLog{},     //  商家操作日志
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 检查并创建默认角色和权限
	return s.initData()
}

func (s *InitService) initData() error {
	return nil
}
