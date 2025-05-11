package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type Spec struct {
	SpecID         uint   `gorm:"primaryKey;autoIncrement;comment:规格组id"`
	SpecName       string `gorm:"type:varchar(2000);not null;default:'';comment:规格名"`
	ShopSupplierID int    `gorm:"not null;default:0;comment:门店id"`
	Sort           int    `gorm:"not null;default:0;comment:排序"`
	AppID          uint   `gorm:"not null;default:0;comment:应用id"`
	CreateTime     int64  `gorm:"not null;comment:创建时间"`
}

type SpecRepository interface {
	GetSpecList() ([]*Spec, error)
	ConvertSpec() error
}

func NewSpecService(db *gorm.DB, targetDB *gorm.DB) SpecRepository {
	return &SpecService{db: db, targetDB: targetDB}
}

type SpecService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *SpecService) GetSpecList() ([]*Spec, error) {
	var specs []*Spec
	err := s.db.Find(&specs).Error
	if err != nil {
		return nil, err
	}
	return specs, nil
}

func (s *SpecService) ConvertSpec() error {
	specs, err := s.GetSpecList()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, spec := range specs {
		fmt.Println(fmt.Sprintf("-------迁移spec: %+v", spec))
		names := Names{}
		err := names.GetNames(spec.SpecName)
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("spec_id: %d, spec_name: %+v", spec.SpecID, names))

		id, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)

		flavor := model.ProductFlavor{
			BaseModel: model.BaseModel{
				Uuid:       uint64(spec.SpecID),
				CreateTime: spec.CreateTime,
				UpdateTime: spec.CreateTime,
			},
			Name:                  spec.SpecName,
			MultiLanguageNameUuid: uint(id),
			MultiLanguageName:     languageName,
		}
		_, err = base.NewProductFlavorRepo(s.targetDB).CreateProductFlavor(flavor)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
