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

// SimpleCache 简单的内存缓存实现
type SimpleCache struct {
	items map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

// NewSimpleCache 创建一个新的简单缓存
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		items: make(map[string]cacheItem),
	}
}

// Set 设置缓存
func (c *SimpleCache) Set(key string, value interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get 获取缓存
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().UnixNano() > item.expiration {
		delete(c.items, key)
		return nil, false
	}

	return item.value, true
}

// Del 删除缓存
func (c *SimpleCache) Del(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

// 采购单
type PurchaseOrder struct {
	ID             uint    `gorm:"column:id;primaryKey"`
	Number         string  `gorm:"column:number;comment:采购编号"`
	Name           string  `gorm:"column:name;comment:采购名称"`
	Type           int     `gorm:"column:type;comment:采购方式 10-总部采购 20-自行采购"`
	ApplicantID    int     `gorm:"column:applicant_id;comment:申请人id"`
	TotalNum       float64 `gorm:"column:total_num;comment:商品总数"`
	TotalAmount    float64 `gorm:"column:total_amount;comment:采购总额"`
	ArrivalTime    int64   `gorm:"column:arrival_time;comment:到货时间"`
	Status         int     `gorm:"column:status;comment:状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库"`
	Remark         string  `gorm:"column:remark;comment:备注"`
	ShopSupplierID int     `gorm:"column:shop_supplier_id;comment:门店id"`
	AppID          int     `gorm:"column:app_id;comment:应用id"`
	CreateTime     int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime     int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime     int64   `gorm:"column:delete_time;comment:删除时间"`

	PurchaseOrderDetail       []PurchaseOrderDetail       `gorm:"foreignKey:PurchaseOrderID"` // 采购单明细
	PurchaseOrderOperationLog []PurchaseOrderOperationLog `gorm:"foreignKey:PurchaseOrderID"` // 采购单操作日志
}

// 采购单明细
type PurchaseOrderDetail struct {
	ID                    uint    `gorm:"column:id;primaryKey;"`
	PurchaseOrderID       uint    `gorm:"column:purchase_order_id;comment:采购单id"`
	ProductSkuID          uint    `gorm:"column:product_sku_id;comment:商品规格id"`
	ProductID             uint    `gorm:"column:product_id;comment:产品表id"`
	EstimatePurchasePrice float64 `gorm:"column:estimate_purchase_price;comment:预计采购单价"`
	EstimatePurchaseNum   float64 `gorm:"column:estimate_purchase_num;comment:预计采购数量"`
	EstimateTotalAmount   float64 `gorm:"column:estimate_total_amount;comment:预计采购总额"`
	ActualPurchasePrice   float64 `gorm:"column:actual_purchase_price;comment:实际采购单价"`
	ActualPurchaseNum     float64 `gorm:"column:actual_purchase_num;comment:实际采购数量"`
	ActualTotalAmount     float64 `gorm:"column:actual_total_amount;comment:实际采购总额"`
	ShopSupplierID        uint    `gorm:"column:shop_supplier_id;comment:门店id"`
	AppID                 uint    `gorm:"column:app_id;comment:应用id"`
	CreateTime            int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime            int64   `gorm:"column:update_time;comment:更新时间"`
}

// 采购单操作日志
type PurchaseOrderOperationLog struct {
	ID              string `gorm:"column:id;primaryKey;"`
	PurchaseOrderID uint   `gorm:"column:purchase_order_id;comment:采购单id"`
	OperatorID      int    `gorm:"column:operator_id;comment:操作人id"`
	Username        string `gorm:"column:username;comment:操作人员"`
	Status          int    `gorm:"column:status;comment:操作状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库"`
	Operation       string `gorm:"column:operation;comment:操作动作"`
	Remark          string `gorm:"column:remark;comment:操作说明"`
	ShopSupplierID  uint   `gorm:"column:shop_supplier_id;comment:门店id"`
	AppID           uint   `gorm:"column:app_id;comment:应用id"`
	CreateTime      int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64  `gorm:"column:update_time;comment:更新时间"`
}

// 出入库记录
type InventoryRecord struct {
	ID              string  `gorm:"column:id;primaryKey;"`
	Number          string  `gorm:"column:number;comment:记录编号"`
	InventoryType   int     `gorm:"column:inventory_type;comment:类型 1-入库 2-出库"`
	Type            int     `gorm:"column:type;comment:操作类型 10-采购入库 20-调整入库 21-添加入库 30-销售出库 40-调整出库 41-删除出库"`
	PurchaseOrderID uint    `gorm:"column:purchase_order_id;comment:采购单id"`
	ProductSkuID    uint    `gorm:"column:product_sku_id;comment:商品规格id"`
	ProductSkuName  string  `gorm:"column:product_sku_name;comment:产品规格名称"`
	OrderID         uint    `gorm:"column:order_id;comment:订单id"`
	ProductID       uint    `gorm:"column:product_id;comment:商品id"`
	Num             float64 `gorm:"column:num;comment:商品数量"`
	Remark          string  `gorm:"column:remark;comment:备注"`
	Status          int     `gorm:"column:status;comment:状态 10-已入库 20-已出库 30-已撤销"`
	OperatorID      uint    `gorm:"column:operator_id;comment:操作人id"`
	Username        string  `gorm:"column:username;comment:操作人员"`
	InTime          int64   `gorm:"column:in_time;comment:入库时间"`
	OutTime         int64   `gorm:"column:out_time;comment:出库时间"`
	RevokeTime      int64   `gorm:"column:revoke_time;comment:撤销时间"`
	ShopSupplierID  uint    `gorm:"column:shop_supplier_id;comment:门店id"`
	AppID           uint    `gorm:"column:app_id;comment:应用id"`
	CreateTime      int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime      int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime      int64   `gorm:"column:delete_time;comment:删除时间"`

	Product       *Product       `gorm:"foreignKey:ProductID;references:ProductID"` // 商品
	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID"`                // 采购单
}

// 采购单状态映射
var PurchaseFormStatusMap = map[int]int{
	10: 0, // 待审核
	20: 1, // 已驳回
	30: 2, // 采购中
	40: 3, // 已采购
	50: 4, // 已入库
}

// 入库记录场景映射
var WarehouseFormSceneMap = map[int]int{
	10: 0, // 采购入库
	20: 2, // 调整入库
	21: 1, // 添加入库
}

// 入库记录状态映射
var WarehouseFormStatus = map[int]int{
	10: 0, // 已入库
	30: 2, // 已撤销
}

// 采购单表名
func (PurchaseOrder) TableName() string {
	return "jjjfood_erp_purchase_order"
}

// 采购单明细表名
func (PurchaseOrderDetail) TableName() string {
	return "jjjfood_erp_purchase_detail"
}

// 采购单操作日志表名
func (PurchaseOrderOperationLog) TableName() string {
	return "jjjfood_erp_purchase_operation_log"
}

// erp出入库记录表
func (InventoryRecord) TableName() string {
	return "jjjfood_erp_inventory_record"
}

// 采购服务
type PurchaseService struct {
	db       *gorm.DB
	targetDB *gorm.DB
	cache    *SimpleCache
}

// 实例化采购单服务
func NewPurchaseService(db *gorm.DB, targetDB *gorm.DB) *PurchaseService {
	return &PurchaseService{
		db:       db,
		targetDB: targetDB,
		cache:    NewSimpleCache(),
	}
}

// 获取商品包装
func (s *PurchaseService) GetProductPackage(productUuid uint64) model.ProductPackage {
	// 从缓存中获取
	cacheKey := fmt.Sprintf("product_package_%d", productUuid)
	if cacheData, ok := s.cache.Get(cacheKey); ok {
		return cacheData.(model.ProductPackage)
	}

	var product model.ProductPackage
	s.targetDB.Preload("ProductBoms").Where("uuid = ?", productUuid).First(&product)
	if product.ID == 0 {
		return model.ProductPackage{}
	}

	// 设置缓存
	s.cache.Set(cacheKey, product, 30*time.Minute)
	return product
}

func (s *PurchaseService) GetProductBom(productSkuId uint64) model.ProductBom {
	// 从缓存中获取
	cacheKey := fmt.Sprintf("product_bom_%d", productSkuId)
	if cacheData, ok := s.cache.Get(cacheKey); ok {
		return cacheData.(model.ProductBom)
	}

	var productBom model.ProductBom
	s.targetDB.Where("uuid = ?", productSkuId).First(&productBom)
	if productBom.ID == 0 {
		return model.ProductBom{}
	}

	// 设置缓存
	s.cache.Set(cacheKey, productBom, 30*time.Minute)
	return productBom
}

func (s *PurchaseService) GetProductSpecList() ([]*Spec, error) {
	var specs []*Spec
	if err := s.db.Find(&specs).Error; err != nil {
		return nil, err
	}
	return specs, nil
}

func (s *PurchaseService) CreateProductBom(productID uint64, productSkuID uint64, bomName string) (uint64, error) {
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

	return productBom.Uuid, nil
}

// 获取原材料
func (s *PurchaseService) GetMaterial(productUuid uint64) model.Material {
	// 从缓存中获取
	cacheKey := fmt.Sprintf("material_%d", productUuid)
	if cacheData, ok := s.cache.Get(cacheKey); ok {
		return cacheData.(model.Material)
	}

	var material model.Material
	s.targetDB.Where("uuid = ?", productUuid).First(&material)
	if material.ID == 0 {
		return model.Material{}
	}

	// 设置缓存
	s.cache.Set(cacheKey, material, 30*time.Minute)
	return material
}

// 迁移采购单
func (s *PurchaseService) ConvertPurchaseForm() error {
	var (
		oldOrders       []PurchaseOrder
		newOrders       []model.PurchaseForm
		newOrderDetails []model.PurchaseFormItem
		newOrderLogs    []model.PurchaseFormLog
	)

	s.db.Preload("PurchaseOrderDetail").Preload("PurchaseOrderOperationLog").Find(&oldOrders)
	for _, oldOrder := range oldOrders {
		newOrders = append(newOrders, model.PurchaseForm{
			BaseModel: model.BaseModel{
				ID:         uint(oldOrder.ID),
				Uuid:       uint64(oldOrder.ID),
				CreateTime: oldOrder.CreateTime,
				UpdateTime: oldOrder.UpdateTime,
				DeleteTime: oldOrder.DeleteTime,
			},
			FormNo:        oldOrder.Number,
			Name:          oldOrder.Name,
			ApplicantUuid: uint64(oldOrder.ApplicantID),
			Remark:        oldOrder.Remark,
			Num:           oldOrder.TotalNum,
			Amount:        oldOrder.TotalAmount,
			Status:        PurchaseFormStatusMap[oldOrder.Status],
			ArrivalTime:   oldOrder.ArrivalTime,
		})

		var deleteTime int64 = 0 // 删除时间
		if oldOrder.DeleteTime > 0 {
			deleteTime = oldOrder.DeleteTime
		}

		for _, oldDetail := range oldOrder.PurchaseOrderDetail {
			var (
				materialType int    = 0                              // 物料类型
				materialUuid uint64 = uint64(oldDetail.ProductSkuID) // 材料uuid
			)
			material := s.GetMaterial(uint64(oldDetail.ProductID))
			if material.ID > 0 {
				materialType = 1
				materialUuid = material.Uuid
			}
			newOrderDetails = append(newOrderDetails, model.PurchaseFormItem{
				BaseModel: model.BaseModel{
					ID:         uint(oldDetail.ID),
					Uuid:       uint64(oldDetail.ID),
					CreateTime: oldDetail.CreateTime,
					UpdateTime: oldDetail.UpdateTime,
					DeleteTime: deleteTime,
				},
				PurchaseFormUuid: uint64(oldOrder.ID),
				MaterialType:     materialType,
				MaterialUuid:     materialUuid,
				EstimateNum:      int(oldDetail.EstimatePurchaseNum),
				EstimatePrice:    oldDetail.EstimatePurchasePrice,
				EstimateAmount:   oldDetail.EstimateTotalAmount,
				Num:              oldDetail.ActualPurchaseNum,
				Price:            oldDetail.ActualPurchasePrice,
				Amount:           oldDetail.ActualTotalAmount,
			})
		}

		for _, oldLog := range oldOrder.PurchaseOrderOperationLog {
			uuid, _ := utils.GetID()
			newOrderLogs = append(newOrderLogs, model.PurchaseFormLog{
				BaseModel: model.BaseModel{
					Uuid:       uuid,
					CreateTime: oldLog.CreateTime,
					UpdateTime: oldLog.UpdateTime,
					DeleteTime: deleteTime,
				},
				PurchaseFormUuid: uint64(oldOrder.ID),
				OperatorUuid:     uint64(oldLog.OperatorID),
				Username:         oldLog.Username,
				Status:           PurchaseFormStatusMap[oldLog.Status],
				Operation:        oldLog.Operation,
				Remark:           oldLog.Remark,
			})
		}
	}

	// 批量创建采购单
	if len(newOrders) > 0 {
		if err := s.targetDB.Create(newOrders).Error; err != nil {
			return err
		}
	}

	if len(newOrderDetails) > 0 {
		if err := s.targetDB.Create(newOrderDetails).Error; err != nil {
			return err
		}
	}

	if len(newOrderLogs) > 0 {
		if err := s.targetDB.Create(newOrderLogs).Error; err != nil {
			return err
		}
	}

	return nil
}

// 迁移入库记录
func (s *PurchaseService) ConvertWarehouseForm() error {
	var (
		records    []InventoryRecord
		newRecords []model.WarehouseForm
	)
	s.db.Preload("Product").
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.PurchaseOrderDetail").
		Where("inventory_type = ?", 1).
		Find(&records)
	for _, record := range records {
		// 迁移采购入库
		if record.Type == 10 {
			if record.PurchaseOrder != nil {
				for _, purchaseOrderDetail := range record.PurchaseOrder.PurchaseOrderDetail {
					uuid, err := utils.GetID()
					if err != nil {
						return err
					}
					productBomUuid := uint64(0)
					materialUuid := uint64(0)
					if s.GetProductPackage(uint64(purchaseOrderDetail.ProductID)).ID > 0 {
						productBomUuid = uint64(purchaseOrderDetail.ProductSkuID)
					}
					// productPackage := s.GetProductPackage(uint64(purchaseOrderDetail.ProductID))
					// if productPackage.ID > 0 {
					// 	// productBomUuid = uint64(purchaseOrderDetail.ProductSkuID)
					// 	productBomUuid = productPackage.GetFlavorProductBom().Uuid
					// }

					if s.GetMaterial(uint64(purchaseOrderDetail.ProductID)).ID > 0 {
						materialUuid = uint64(purchaseOrderDetail.ProductID)
					}

					newRecords = append(newRecords, model.WarehouseForm{
						BaseModel: model.BaseModel{
							Uuid:       uuid,
							CreateTime: record.InTime,
							UpdateTime: record.InTime,
							DeleteTime: record.DeleteTime,
						},
						FormNo:            record.Number,
						Scene:             WarehouseFormSceneMap[record.Type],
						Num:               int(purchaseOrderDetail.EstimatePurchaseNum),
						Remark:            record.Remark,
						Status:            WarehouseFormStatus[record.Status],
						ProductBomUuid:    productBomUuid,
						MaterialUuid:      materialUuid,
						PurchaseOrderUuid: uint64(record.PurchaseOrderID),
						OperatorUuid:      uint64(record.OperatorID),
						RevokeTime:        int(record.RevokeTime),
					})
				}
			}
		} else {
			uuid, err := utils.GetID()
			if err != nil {
				return err
			}
			productBomUuid := uint64(0)
			materialUuid := uint64(0)
			if record.Product != nil {
				if record.Product.Type == 10 {
					productBom := s.GetProductBom(uint64(record.ProductSkuID))
					if productBom.ID > 0 {
						productBomUuid = productBom.Uuid
					} else {
						productBomUuid, err = s.CreateProductBom(uint64(record.ProductID), uint64(record.ProductSkuID), record.ProductSkuName)
						if err != nil {
							return errors.WithMessage(err, "创建商品规格失败")
						}
					}
					// productBomUuid = uint64(record.ProductSkuID)
				} else {
					materialUuid = uint64(record.ProductID)
				}
			}
			newRecords = append(newRecords, model.WarehouseForm{
				BaseModel: model.BaseModel{
					Uuid:       uuid,
					CreateTime: record.InTime,
					UpdateTime: record.InTime,
					DeleteTime: record.DeleteTime,
				},
				FormNo:            record.Number,
				Scene:             WarehouseFormSceneMap[record.Type],
				Num:               int(record.Num),
				Remark:            record.Remark,
				Status:            WarehouseFormStatus[record.Status],
				ProductBomUuid:    productBomUuid,
				MaterialUuid:      materialUuid,
				PurchaseOrderUuid: uint64(record.PurchaseOrderID),
				OperatorUuid:      uint64(record.OperatorID),
				RevokeTime:        int(record.RevokeTime),
			})
		}
	}
	if len(newRecords) > 0 {
		if err := s.targetDB.Create(newRecords).Error; err != nil {
			return err
		}
	}

	return nil
}
