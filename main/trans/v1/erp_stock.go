package v1

import (
	"fmt"
	"sync"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"
	pkgUtils "ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// 默认规格，在无法连接数据库时使用
var defaultSpec = &Spec{
	SpecID:   1,
	SpecName: "默认规格",
}

// 出库记录场景映射
var WarehouseOutFormSceneMap = map[int]int{
	30: 0, // 销售出库
	40: 1, // 调整出库
	41: 4, // 删除出库
}

// 出库记录状态映射
var WarehouseOutFormStatusMap = map[int]int{
	20: 0, // 已出库
	30: 1, // 已撤销
}

// 出库记录明细场景映射
var WarehouseOutItemSceneMap = map[int]int{
	30: 0, // 销售出库
	40: 1, // 调整出库
	41: 4, // 删除出库
}

// 损耗记录场景映射
var LossReportFormSceneMap = map[int]int{
	1: 1, // 丢失
	2: 0, // 损耗
}

// 损耗记录状态映射
var LossReportFormStatusMap = map[int]int{
	0: 0, // 未审核
	1: 1, // 通过
	2: 2, // 拒绝
}

// 损耗记录
type DamagedProductRecord struct {
	ID             string  `gorm:"column:id;primaryKey;"`
	Number         string  `gorm:"column:number;comment:编号"`
	Type           int     `gorm:"column:type;comment:损坏类型 1-丢失 2-损坏"`
	ProductID      uint    `gorm:"column:product_id;comment:商品id"`
	ProductSkuID   uint    `gorm:"column:product_sku_id;comment:商品规格id"`
	Num            float64 `gorm:"column:num;comment:报损数量"`
	Remark         string  `gorm:"column:remark;comment:备注"`
	ReviewStatus   int     `gorm:"column:review_status;comment:审核状态 0-未审核 1-通过 2-拒绝"`
	OperatorID     uint    `gorm:"column:operator_id;comment:操作人id"`
	Refused        string  `gorm:"column:refused;comment:拒绝原因"`
	ApprovedTime   int64   `gorm:"column:approved_time;comment:审核通过时间"`
	RejectedTime   int64   `gorm:"column:rejected_time;comment:审核拒绝时间"`
	ShopSupplierID uint    `gorm:"column:shop_supplier_id;comment:门店id"`
	AppID          uint    `gorm:"column:app_id;comment:应用id"`
	CreateTime     int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime     int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime     int64   `gorm:"column:delete_time;comment:删除时间"`

	Product *Product `gorm:"foreignKey:ProductID;references:ProductID"` // 商品
}

// 损耗记录表名
func (DamagedProductRecord) TableName() string {
	return "jjjfood_erp_damaged_product_record"
}

// 库存服务
type StockService struct {
	db             *gorm.DB
	targetDB       *gorm.DB
	bomCache       map[uint64]model.ProductBom
	cacheMu        sync.RWMutex
	specsCache     []*Spec
	specsCacheTime int64
}

// 实例化库存服务
func NewStockService(db *gorm.DB, targetDB *gorm.DB) *StockService {
	return &StockService{
		db:             db,
		targetDB:       targetDB,
		bomCache:       make(map[uint64]model.ProductBom),
		specsCacheTime: 0,
	}
}

// 迁移出库记录
func (s *StockService) ConvertWarehouseOut() error {
	isContinue := true
	page := 1
	pageSize := 2000

	for isContinue {
		var (
			records  []InventoryRecord
			newForms []model.WarehouseOutForm
			newItems []model.WarehouseOutFormItem
		)
		s.db.Preload("Product").Where("inventory_type = ?", 2).Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)
		if len(records) < pageSize {
			isContinue = false
		}
		for _, record := range records {
			formUuid, err := utils.GetID()
			if err != nil {
				return err
			}
			newForms = append(newForms, model.WarehouseOutForm{
				BaseModel: model.BaseModel{
					Uuid:       formUuid,
					CreateTime: record.CreateTime,
					UpdateTime: record.OutTime,
					DeleteTime: record.DeleteTime,
				},
				FormNo:       record.Number,
				Scene:        WarehouseOutFormSceneMap[record.Type],
				Remark:       record.Remark,
				Status:       WarehouseOutFormStatusMap[record.Status],
				RevokeTime:   record.RevokeTime,
				OperatorUuid: uint64(record.OperatorID),
			})

			itemUuid, err := utils.GetID()
			if err != nil {
				return nil
			}
			productBomUuid := uint64(0)
			materialUuid := uint64(0)
			if record.Product != nil {
				if record.Product.Type == 10 {
					// 先从缓存中查询
					s.cacheMu.RLock()
					if bom, exists := s.bomCache[uint64(record.ProductSkuID)]; exists {
						productBomUuid = bom.Uuid
						s.cacheMu.RUnlock()
					} else {
						s.cacheMu.RUnlock()
						// 缓存中不存在，从数据库查询
						productBom := s.GetProductBom(uint64(record.ProductSkuID))
						if productBom.ID > 0 {
							productBomUuid = productBom.Uuid
						} else {
							// 如果数据库中也不存在，则创建新的
							var createErr error
							productBomUuid, createErr = s.CreateProductBom(uint64(record.ProductID), uint64(record.ProductSkuID), record.ProductSkuName)
							if createErr != nil {
								return errors.WithMessage(createErr, "创建商品规格失败")
							}
						}
					}
				} else {
					materialUuid = uint64(record.ProductID)
				}
			}
			newItems = append(newItems, model.WarehouseOutFormItem{
				BaseModel: model.BaseModel{
					Uuid:       itemUuid,
					CreateTime: record.CreateTime,
					UpdateTime: record.OutTime,
					DeleteTime: record.DeleteTime,
				},
				Num:                  record.Num,
				Scene:                WarehouseOutItemSceneMap[record.Type],
				Status:               1,
				ReduceStock:          1,
				RevokeTime:           record.RevokeTime,
				WarehouseOutFormUuid: formUuid,
				ProductBomUuid:       productBomUuid,
				MaterialUuid:         materialUuid,
			})
		}

		if len(newForms) > 0 {
			if err := s.targetDB.Create(newForms).Error; err != nil {
				return err
			}

			fmt.Printf("已保存批次: offset=%d, batch=%d-%d\n", (page-1)*pageSize, page, len(newForms))
		}

		if len(newItems) > 0 {
			if err := s.targetDB.Create(newItems).Error; err != nil {
				return err
			}
		}

		if len(records) > 0 {
			fmt.Println(fmt.Sprintf("warehouse-out - page: %d - num: %d", page, len(records)))
		}

		page++
	}

	return nil
}

// 迁移损耗记录
func (s *StockService) ConvertDemaged() error {
	isContinue := true
	page := 1
	pageSize := 2000

	for isContinue {
		var (
			records    []*DamagedProductRecord
			newRecords []*model.LossReportForm
		)
		s.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)
		if len(records) < pageSize {
			isContinue = false
		}
		for _, record := range records {
			formUuid, err := utils.GetID()
			if err != nil {
				return err
			}
			productBomUuid := uint64(0)
			materialUuid := uint64(0)
			if record.Product != nil {
				if record.Product.Type == 10 {
					productBomUuid = uint64(record.ProductSkuID)
				} else {
					materialUuid = uint64(record.ProductSkuID)
				}
			}
			newRecords = append(newRecords, &model.LossReportForm{
				BaseModel: model.BaseModel{
					Uuid:       formUuid,
					CreateTime: record.CreateTime,
					UpdateTime: record.UpdateTime,
					DeleteTime: record.DeleteTime,
				},
				FormNo:         record.Number,
				Scene:          LossReportFormSceneMap[record.Type],
				Num:            record.Num,
				Remark:         record.Remark,
				ProductBomUuid: productBomUuid,
				MaterialUuid:   materialUuid,
				ApplicantUuid:  uint64(record.OperatorID),
				RejectReason:   record.Refused,
				Status:         LossReportFormStatusMap[record.ReviewStatus],
				OperatorUuid:   uint64(record.OperatorID),
				ApprovedTime:   record.ApprovedTime,
				RevokeTime:     record.RejectedTime,
			})
		}

		if len(newRecords) > 0 {
			if err := s.targetDB.Create(newRecords).Error; err != nil {
				return err
			}
		}
		page++
	}

	return nil
}

func (s *StockService) GetProductBom(productSkuId uint64) model.ProductBom {
	// 先从缓存中查询
	s.cacheMu.RLock()
	if bom, exists := s.bomCache[productSkuId]; exists {
		s.cacheMu.RUnlock()
		return bom
	}
	s.cacheMu.RUnlock()

	// 缓存中不存在，从数据库查询
	var productBom model.ProductBom
	s.targetDB.Where("uuid = ?", productSkuId).First(&productBom)
	if productBom.ID == 0 {
		return model.ProductBom{}
	}

	// 更新缓存
	s.cacheMu.Lock()
	s.bomCache[productSkuId] = productBom
	s.cacheMu.Unlock()

	return productBom
}

func (s *StockService) GetProductSpecList() ([]*Spec, error) {
	// 检查缓存是否有效 (5分钟)
	s.cacheMu.RLock()
	if len(s.specsCache) > 0 && time.Now().Unix()-s.specsCacheTime < 300 {
		specs := s.specsCache
		s.cacheMu.RUnlock()
		return specs, nil
	}
	s.cacheMu.RUnlock()

	var specs []*Spec
	if err := s.db.Find(&specs).Error; err != nil {
		return nil, err
	}

	// 更新缓存
	if len(specs) > 0 {
		s.cacheMu.Lock()
		s.specsCache = specs
		s.specsCacheTime = time.Now().Unix()
		s.cacheMu.Unlock()
	}

	return specs, nil
}

func (s *StockService) CreateProductBom(productID uint64, productSkuID uint64, bomName string) (uint64, error) {
	// 如果商品已经删除，则返回一个空的productBom
	// 获取商品规格.随便给个这个空的规格设置一个specId
	// specs, err := s.GetProductSpecList()
	// if err != nil {
	// 	return 0, errors.WithMessage(err, "获取商品规格失败")
	// }
	// if len(specs) == 0 {
	// 	return 0, errors.WithMessage(err, "商品规格为空")
	// }
	// specId := specs[0].SpecID

	productBomUuid, err := pkgUtils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "获取uuid失败")
	}
	productBom := model.ProductBom{
		BaseModel: model.BaseModel{
			Uuid:       productBomUuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
			DeleteTime: 1,
		},
		ProductPackageUuid: productID,
		ProductFlavorUuid:  uint64(productSkuID),
		Name:               bomName,
	}

	if err := s.targetDB.Create(&productBom).Error; err != nil {
		return 0, errors.WithMessage(err, "创建商品规格失败")
	}

	// 更新缓存
	s.cacheMu.Lock()
	s.bomCache[productSkuID] = productBom
	s.cacheMu.Unlock()

	return productBom.Uuid, nil
}
