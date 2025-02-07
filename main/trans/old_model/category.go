package old_model

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

type Category struct {
	CategoryID     uint   `gorm:"primaryKey;autoIncrement;comment:产品分类id"`
	Name           string `gorm:"type:varchar(2000);not null;default:'';comment:分类名称"`
	ParentID       uint   `gorm:"not null;default:0;comment:上级分类id"`
	ImageID        uint   `gorm:"not null;default:0;comment:分类图片id"`
	IsSpecial      bool   `gorm:"not null;default:false;comment:0普通1特殊"`
	Sort           uint   `gorm:"not null;default:0;comment:排序方式(数字越小越靠前)"`
	Type           uint8  `gorm:"not null;default:0;comment:0外卖1店内"`
	ShopSupplierID int    `gorm:"not null;default:0;comment:门店id"`
	Status         bool   `gorm:"not null;default:true;comment:是否显示1显示0隐藏"`
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

func (n *Names) CreateMultiLanguageName(nameId uint, targetDB *gorm.DB) error {
	multiLanguageName := model.MultiLanguageName{
		Uuid:     nameId,
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
	_, err := repository.NewMultiLanguageNameRepository(targetDB).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		return err
	}
	return nil
}

func (n *Names) GenMultiLanguageName(nameId uint) model.MultiLanguageName {
	multiLanguageName := model.MultiLanguageName{
		Uuid:     nameId,
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

		id, err := database.GetID()
		if err != nil {
			return err
		}
		fmt.Printf("uuid:%+v\n", id)
		multiLanguageName := model.MultiLanguageName{
			Uuid:     uint(id),
			EnName:   names.En,
			ZhName:   names.Zh,
			ZhTwName: names.ZhTw,
			ThName:   names.Th,
			MyName:   names.My,
			JaName:   names.Ja,
			KoName:   names.Ko,
			TrName:   names.Tr,
		}
		_, err = repository.NewMultiLanguageNameRepository(s.targetDB).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return err
		}

		if category.IsSpecial {
			specialCategory := model.ProductSpecialCategory{
				Uuid:                  category.CategoryID,
				Status:                category.Status,
				Name:                  multiLanguageName.ZhName,
				MultiLanguageNameUuid: uint(id),
				OrderBy:               category.Sort,
			}
			_, err := repository.NewProductSpecialCategoryRepository(s.targetDB).CreateProductSpecialCategory(specialCategory)
			if err != nil {
				return err
			}
		} else {
			productCategory := model.ProductCategory{
				Uuid:                  category.CategoryID,
				Name:                  names.Zh,
				ParentUuid:            category.ParentID,
				MultiLanguageNameUuid: uint(id),
				Status:                category.Status,
				OrderBy:               category.Sort,
			}
			_, err = repository.NewProductCategoryRepository(s.targetDB).CreateProductCategory(productCategory)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
