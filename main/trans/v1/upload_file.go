package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type UploadFile struct {
	FileID         uint64 `gorm:"primaryKey;autoIncrement;comment:'文件id'"`
	Storage        string `gorm:"type:varchar(20);not null;default:'';comment:'存储方式'"`
	GroupID        uint64 `gorm:"not null;default:0;comment:'文件分组id'"`
	FileURL        string `gorm:"type:varchar(255);not null;default:'';comment:'存储域名'"`
	SaveName       string `gorm:"type:varchar(255);default:'';comment:'保存路径'"`
	FileName       string `gorm:"type:varchar(255);not null;default:'';comment:'文件路径'"`
	FileSize       uint64 `gorm:"not null;default:0;comment:'文件大小(字节)'"`
	FileType       string `gorm:"type:varchar(20);not null;default:'';comment:'文件类型'"`
	RealName       string `gorm:"type:varchar(255);default:'';comment:'文件真实名'"`
	URLParam       string `gorm:"type:text;default:null;comment:'签名参数'"`
	IndexFileName  string `gorm:"type:varchar(500);default:'';comment:'文件唯一名'"`
	Extension      string `gorm:"type:varchar(20);not null;default:'';comment:'文件扩展名'"`
	IsUser         uint64 `gorm:"not null;default:0;comment:'是否为c端用户上传'"`
	IsRecycle      uint8  `gorm:"not null;default:0;comment:'是否已回收'"`
	ShopSupplierID int    `gorm:"default:0;comment:'供应商id'"`
	IsDelete       uint8  `gorm:"not null;default:0;comment:'软删除'"`
	AppID          uint64 `gorm:"default:0;comment:'应用id'"`
	CreateTime     uint64 `gorm:"not null;default:0;comment:'创建时间'"`
	UpdateTime     uint64 `gorm:"default:0;comment:'更新时间'"`
}

type UploadFileRepository interface {
	GetUploadFileList() ([]*UploadFile, error)
	ConvertUploadFile() error
}

func NewUploadFileService(db *gorm.DB, targetDB *gorm.DB) UploadFileRepository {
	return &UploadFileService{
		db:       db,
		targetDB: targetDB,
	}
}

type UploadFileService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *UploadFileService) GetUploadFileList() ([]*UploadFile, error) {
	var uploadFiles []*UploadFile
	err := s.db.Model(&UploadFile{}).Find(&uploadFiles).Error
	return uploadFiles, err
}

func (s *UploadFileService) ConvertUploadFile() error {
	uploadFiles, err := s.GetUploadFileList()
	if err != nil {
		return err
	}
	for _, uploadFile := range uploadFiles {
		fmt.Println(fmt.Sprintf("uploadFile: %+v", uploadFile))
		model := model.File{
			BaseModel: model.BaseModel{
				Uuid:       uint64(uploadFile.FileID),
				CreateTime: int64(uploadFile.CreateTime),
				UpdateTime: int64(uploadFile.UpdateTime),
			},
			Storage:       uploadFile.Storage,
			GroupUuid:     uploadFile.GroupID,
			FileUrl:       uploadFile.FileURL,
			SaveName:      uploadFile.SaveName,
			FileName:      uploadFile.FileName,
			FileSize:      int(uploadFile.FileSize),
			FileType:      uploadFile.FileType,
			RealName:      uploadFile.RealName,
			UrlParam:      uploadFile.URLParam,
			IndexFileName: uploadFile.IndexFileName,
			Extension:     uploadFile.Extension,
			IsUser:        int(uploadFile.IsUser),
			IsRecycle:     int(uploadFile.IsRecycle),
		}
		err := repository.NewFileRepo(s.targetDB).CreateFile(model)
		if err != nil {
			return err
		}
	}
	return nil
}
