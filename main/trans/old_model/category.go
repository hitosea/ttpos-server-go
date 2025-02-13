package old_model

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type Category struct {
	CategoryID     uint64 `gorm:"primaryKey;autoIncrement;comment:产品分类id"`
	Name           string `gorm:"type:varchar(2000);not null;default:'';comment:分类名称"`
	ParentID       uint64 `gorm:"not null;default:0;comment:上级分类id"`
	ImageID        uint   `gorm:"not null;default:0;comment:分类图片id"`
	IsSpecial      bool   `gorm:"not null;default:false;comment:0普通1特殊"`
	Sort           uint   `gorm:"not null;default:0;comment:排序方式(数字越小越靠前)"`
	Type           uint8  `gorm:"not null;default:0;comment:0外卖1店内"`
	ShopSupplierID int    `gorm:"not null;default:0;comment:门店id"`
	Status         uint   `gorm:"not null;default:1;comment:是否显示1显示0隐藏"`
	IsButton       int    `gorm:"default:0;comment:是否按钮 0-否 1-是"`
	AppID          uint   `gorm:"not null;default:0;comment:应用id"`
	CreateTime     uint   `gorm:"not null;default:0;comment:创建时间"`
	UpdateTime     uint   `gorm:"not null;default:0;comment:更新时间"`
}

type Names struct {
	Zh   string `json:"zh"`
	En   string `json:"en"`
	Th   string `json:"th"`
	Ja   string `json:"ja"`
	Ko   string `json:"ko"`
	My   string `json:"my"`
	Tr   string `json:"tr"`
	ZhTw string `json:"zhtw"`
}

func (n *Names) GetNames(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &n)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return err
	}
	return nil
}

func (n *Names) CreateMultiLanguageName(nameId uint64, targetDB *gorm.DB) error {
	multiLanguageName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid: nameId,
		},
		EnName:   n.En,
		ZhName:   n.Zh,
		ZhTwName: n.ZhTw,
		ThName:   n.Th,
		MyName:   n.My,
		JaName:   n.Ja,
		KoName:   n.Ko,
		TrName:   n.Tr,
	}
	fmt.Println(fmt.Sprintf("multiLanguageName:%+v", multiLanguageName))
	_, err := repository.NewMultiLanguageNameRepoImpl(targetDB).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		return err
	}
	return nil
}

func (n *Names) GenMultiLanguageName(nameId uint64) model.MultiLanguageName {
	multiLanguageName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid: nameId,
		},
		EnName:   n.En,
		ZhName:   n.Zh,
		ZhTwName: n.ZhTw,
		ThName:   n.Th,
		MyName:   n.My,
		JaName:   n.Ja,
		KoName:   n.Ko,
		TrName:   n.Tr,
	}
	return multiLanguageName
}

type CategoryRepository interface {
	GetCategoryList() ([]*Category, error)
	ConvertCategory() error
}

type CategoryService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *CategoryService) GetCategoryList() ([]*Category, error) {
	var categories []*Category
	if err := s.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// 将原jjjfood_category表中的数据转换为新的category表中的数据
// 1. 将多语言名称的json转为multi_language_name表中的数据
// 2. 将原jjjfood_category表中的数据转换为新的category表中的数据
func (s *CategoryService) ConvertCategory() error {
	categoryList, err := s.GetCategoryList()
	if err != nil {
		return err
	}
	for _, category := range categoryList {
		var names Names
		err := json.Unmarshal([]byte(category.Name), &names)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return err
		}
		fmt.Printf("%+v\n", names)

		id, err := utils.GetID()
		if err != nil {
			return err
		}
		fmt.Printf("uuid:%+v\n", id)
		multiLanguageName := model.MultiLanguageName{
			BaseModel: model.BaseModel{
				Uuid: id,
			},
			EnName:   names.En,
			ZhName:   names.Zh,
			ZhTwName: names.ZhTw,
			ThName:   names.Th,
			MyName:   names.My,
			JaName:   names.Ja,
			KoName:   names.Ko,
			TrName:   names.Tr,
		}
		_, err = repository.NewMultiLanguageNameRepoImpl(s.targetDB).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return err
		}

		//if category.IsSpecial {
		//specialCategory := model.ProductSpecialCategory{
		//	Uuid:                  category.CategoryID,
		//	Status:                category.Status,
		//	Name:                  multiLanguageName.ZhName,
		//	MultiLanguageNameUuid: uint(id),
		//	Sort:               category.Sort,
		//}
		//_, err := repository.NewProductSpecialCategoryRepo(s.targetDB).CreateProductSpecialCategory(specialCategory)
		//if err != nil {
		//	return err
		//}
		//} else {
		productCategory := model.ProductCategory{
			BaseModel: model.BaseModel{
				Uuid: uint64(category.CategoryID),
			},
			Name:                  names.Zh,
			ParentUuid:            uint64(category.ParentID),
			MultiLanguageNameUuid: uint64(id),
			Status:                category.Status,
			Sort:                  category.Sort,
		}
		_, err = repository.NewProductCategoryRepo(s.targetDB).CreateProductCategory(productCategory)
		if err != nil {
			return err
		}
		//}
	}
	return nil
}

func (s *CategoryService) parseBoolToUint(status bool) uint {
	if status {
		return 1
	}
	return 0
}
