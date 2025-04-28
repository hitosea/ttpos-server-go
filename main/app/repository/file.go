package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IFileRepo interface {
	CreateFileGroup(fileGroup model.FileGroup) error // 创建文件组
	CreateFile(file model.File) error                // 创建文件
}

func NewFileRepo(db *gorm.DB) IFileRepo {
	return NewFileRepoImpl(db)
}

type fileRepo struct {
	db *gorm.DB
}

func NewFileRepoImpl(db *gorm.DB) IFileRepo {
	return &fileRepo{db: db}
}

func (r *fileRepo) CreateFileGroup(obj model.FileGroup) error {
	group := obj
	group.SetNil()

	err := r.db.Create(&group).Error
	if err != nil {
		return errors.WithMessage(err, "CreateFileGroup")
	}
	return nil
}

func (r *fileRepo) CreateFile(file model.File) error {
	file.SetNil()
	if err := r.db.Create(&file).Error; err != nil {
		return errors.WithMessage(err, "CreateFile")
	}
	return nil
}
