package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IFileRepo interface {
	CreateFileGroup(fileGroup model.FileGroup) error   // 创建文件组
	CreateFile(file model.File) error                  // 创建文件
	CreateFiles(files []model.File) error              // 批量创建文件
	CreateFileGroups(fileGroups []model.FileGroup) error // 批量创建文件组
	GetFileByUuid(uuid uint64) (*model.File, error)    // 根据UUID获取未删除的文件
	HardDeleteFilesByHeadquarter(headquarterUuid uint64) error      // 硬删除指定总部的文件
	HardDeleteFileGroupsByHeadquarter(headquarterUuid uint64) error // 硬删除指定总部的文件组
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

func (r *fileRepo) GetFileByUuid(uuid uint64) (*model.File, error) {
	var file model.File
	if err := r.db.Model(&model.File{}).Where("uuid = ? AND delete_time = 0", uuid).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepo) CreateFiles(files []model.File) error {
	if len(files) == 0 {
		return nil
	}
	return r.db.Create(&files).Error
}

func (r *fileRepo) CreateFileGroups(fileGroups []model.FileGroup) error {
	if len(fileGroups) == 0 {
		return nil
	}
	return r.db.Create(&fileGroups).Error
}

func (r *fileRepo) HardDeleteFilesByHeadquarter(headquarterUuid uint64) error {
	return r.db.Where("headquarter_uuid = ?", headquarterUuid).Delete(&model.File{}).Error
}

func (r *fileRepo) HardDeleteFileGroupsByHeadquarter(headquarterUuid uint64) error {
	return r.db.Where("headquarter_uuid = ?", headquarterUuid).Delete(&model.FileGroup{}).Error
}
