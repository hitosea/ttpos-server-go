package v1

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
)

type OrderScheme struct {
	ID             uint   `gorm:"column:id;type:int(11) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Name           string `gorm:"column:name;type:varchar(255);comment:方案名称;NOT NULL" json:"name"`
	UseChannel     string `gorm:"column:use_channel;type:varchar(255);comment:使用渠道 10-点餐方式 20-桌台方式;NOT NULL" json:"use_channel"`
	TableAreaIds   string `gorm:"column:table_area_ids;type:varchar(255);comment:桌台区域ids;NOT NULL" json:"table_area_ids"`
	MustType       int    `gorm:"column:must_type;type:int(11);default:1;comment:必点类型 1-每人必点1份 2-每笔订单必点1份" json:"must_type"`
	MustRule       int    `gorm:"column:must_rule;type:int(11);default:1;comment:必点规则 1-固定商品 2-可选商品" json:"must_rule"`
	ProductIds     string `gorm:"column:product_ids;type:text;comment:必点商品ids;NOT NULL" json:"product_ids"`
	Status         int    `gorm:"column:status;type:int(11);default:1;comment:状态，1-开启 0-关闭" json:"status"`
	AutoCart       int    `gorm:"column:auto_cart;type:int(11);default:1;comment:自动加入购物车 1-是 0-否" json:"auto_cart"`
	AutoChange     int    `gorm:"column:auto_change;type:int(11);default:1;comment:顾客可修改必点数量 1-是 0-否" json:"auto_change"`
	AutoCheck      int    `gorm:"column:auto_check;type:int(11);default:1;comment:下单时检查必点商品 1-是 0-否" json:"auto_check"`
	AutoCheckout   int    `gorm:"column:auto_checkout;type:int(11);default:1;comment:结账时检查必点商品 1-是 0-否" json:"auto_checkout"`
	ShopSupplierId int    `gorm:"column:shop_supplier_id;type:int(11);default:0;comment:门店id" json:"shop_supplier_id"`
	AppId          int    `gorm:"column:app_id;type:int(11);default:0;comment:应用id" json:"app_id"`
	CreateTime     int64  `gorm:"column:create_time;type:int(11);default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime     int64  `gorm:"column:update_time;type:int(11);default:0;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime     int64  `gorm:"column:delete_time;type:int(11);default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

type OrderSchemeRepository interface {
	GetOrderSchemeList() ([]*OrderScheme, error)
	ConvertOrderScheme() error
}

func NewOrderSchemeService(db *gorm.DB, targetDB *gorm.DB) OrderSchemeRepository {
	return &OrderSchemeService{db: db, targetDB: targetDB}
}

type OrderSchemeService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *OrderSchemeService) GetOrderSchemeList() ([]*OrderScheme, error) {
	var orderSchemes []*OrderScheme
	err := s.db.Find(&orderSchemes).Error
	return orderSchemes, err
}

func (s *OrderSchemeService) ConvertOrderScheme() error {
	orderSchemes, err := s.GetOrderSchemeList()
	if err != nil {
		return err
	}
	for _, orderScheme := range orderSchemes {
		fmt.Println(fmt.Sprintf("order: %+v", orderScheme))

		mustTypeMap := map[int]int{
			1: 1,
			2: 0,
		}
		mustRuleMap := map[int]int{
			1: 0,
			2: 1,
		}
		mustPlanUuid := uint64(orderScheme.ID)
		mustPlanRepo := repository.NewProductMustPlanRepo(s.targetDB)
		_, err := mustPlanRepo.CreateProductMustPlan(&model.ProductMustPlan{
			BaseModel: model.BaseModel{
				Uuid:       mustPlanUuid,
				CreateTime: orderScheme.CreateTime,
				UpdateTime: orderScheme.UpdateTime,
				DeleteTime: orderScheme.DeleteTime,
			},
			Name:         orderScheme.Name,
			UseChannel:   orderScheme.UseChannel,
			MustType:     uint(mustTypeMap[orderScheme.MustType]),
			MustRule:     uint(mustRuleMap[orderScheme.MustRule]),
			Status:       uint(orderScheme.Status),
			AutoCart:     uint(orderScheme.AutoCart),
			AutoChange:   uint(orderScheme.AutoChange),
			AutoCheck:    uint(orderScheme.AutoCheck),
			AutoCheckout: uint(orderScheme.AutoCheckout),
		})
		if err != nil {
			return err
		}

		regionUuids := strings.Split(orderScheme.TableAreaIds, ",")
		var regions []model.ProductMustPlanRegion
		for _, regionUuidStr := range regionUuids {
			regionUuid, _ := strconv.ParseUint(regionUuidStr, 10, 64)
			regions = append(regions, model.ProductMustPlanRegion{
				BaseModel: model.BaseModel{
					CreateTime: orderScheme.CreateTime,
					UpdateTime: orderScheme.UpdateTime,
					DeleteTime: orderScheme.DeleteTime,
				},
				ProductMustPlanUuid: mustPlanUuid,
				DeskRegionUuid:      regionUuid,
			})
		}
		err = mustPlanRepo.CreateProductMustPlanRegion(regions)
		if err != nil {
			return err
		}

		var productUuidStrSlice []string
		err = json.Unmarshal([]byte(orderScheme.ProductIds), &productUuidStrSlice)
		if err != nil {
			return err
		}
		var items []model.ProductMustPlanItem
		for _, productUuidStr := range productUuidStrSlice {
			productUuid, _ := strconv.ParseUint(productUuidStr, 10, 64)
			items = append(items, model.ProductMustPlanItem{
				BaseModel: model.BaseModel{
					CreateTime: orderScheme.CreateTime,
					UpdateTime: orderScheme.UpdateTime,
					DeleteTime: orderScheme.DeleteTime,
				},
				ProductMustPlanUuid: mustPlanUuid,
				ProductPackageUuid:  productUuid,
			})
		}
		err = mustPlanRepo.CreateProductMustPlanItem(items)
		if err != nil {
			return err
		}
	}
	return nil
}
