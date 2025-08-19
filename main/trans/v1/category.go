package v1

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type Category struct {
	CategoryID     uint   `gorm:"column:category_id;type:int(11) unsigned;primary_key;AUTO_INCREMENT;comment:产品分类id" json:"category_id"`
	Name           string `gorm:"column:name;type:varchar(2000);comment:分类名称;NOT NULL" json:"name"`
	ParentID       uint   `gorm:"column:parent_id;type:int(11) unsigned;default:0;comment:上级分类id;NOT NULL" json:"parent_id"`
	ImageID        uint   `gorm:"column:image_id;type:int(11) unsigned;default:0;comment:分类图片id;NOT NULL" json:"image_id"`
	IsSpecial      int    `gorm:"column:is_special;type:tinyint(1);default:0;comment:0普通1特殊;NOT NULL" json:"is_special"`
	Sort           uint   `gorm:"column:sort;type:int(11) unsigned;default:0;comment:排序方式(数字越小越靠前);NOT NULL" json:"sort"`
	Type           int    `gorm:"column:type;type:tinyint(2);default:0;comment:0外卖1店内;NOT NULL" json:"type"`
	ShopSupplierId int    `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:门店id;NOT NULL" json:"shop_supplier_id"`
	Status         int    `gorm:"column:status;type:tinyint(1);default:1;comment:是否显示1显示0隐藏;NOT NULL" json:"status"`
	IsButton       int    `gorm:"column:is_button;type:int(11);default:0;comment:是否按钮 0-否 1-是" json:"is_button"`
	AppId          uint   `gorm:"column:app_id;type:int(11) unsigned;default:0;comment:应用id;NOT NULL" json:"app_id"`
	CreateTime     int64  `gorm:"column:create_time;type:int(11) unsigned;default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     int64  `gorm:"column:update_time;type:int(11) unsigned;default:0;comment:更新时间;NOT NULL" json:"update_time"`
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
		return errors.WithMessage(err)
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
	_, err := repository.NewMultiLanguageNameRepo(targetDB).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		return errors.WithMessage(err)
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

func NewCategoryService(db *gorm.DB, targetDB *gorm.DB) CategoryRepository {
	return &CategoryService{
		db:       db,
		targetDB: targetDB,
	}
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
		return errors.WithMessage(err)
	}
	for _, category := range categoryList {

		fmt.Println(fmt.Sprintf("category: %+v", category))

		var names Names
		err := json.Unmarshal([]byte(category.Name), &names)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return errors.WithMessage(err)
		}
		fmt.Printf("%+v\n", names)

		id, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err)
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
		_, err = repository.NewMultiLanguageNameRepo(s.targetDB).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err)
		}

		uuid := uint64(0)
		if category.CategoryID == 0 {
			uuid = 1971200000000000000
		} else {
			uuid = uint64(category.CategoryID)
		}
		key := ""
		// 如果category是默认的全部分类，则key为all
		if category.CategoryID == 0 || category.IsButton == 1 {
			key = "all"
		}
		productCategory := model.ProductCategory{
			BaseModel: model.BaseModel{
				Uuid:       uuid,
				CreateTime: category.CreateTime,
				UpdateTime: category.UpdateTime,
			},
			Name:                  category.Name,
			MultiLanguageNameUuid: id,
			Status:                category.Status,
			ParentUuid:            uint64(category.ParentID),
			IsSpecial:             category.IsSpecial,
			CategoryKey:           key,
			Sort:                  category.Sort,
		}
		_, err = repository.NewProductCategoryRepo(s.targetDB).CreateProductCategory(productCategory)
		if err != nil {
			return errors.WithMessage(err)
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
