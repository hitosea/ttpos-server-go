package v1

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

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
	db       *gorm.DB
	targetDB *gorm.DB
}

// 实例化库存服务
func NewStockService(db *gorm.DB, targetDB *gorm.DB) *StockService {
	return &StockService{db: db, targetDB: targetDB}
}

// 迁移出库记录
func (s *StockService) ConvertWarehouseOut() error {

	var (
		records  []InventoryRecord
		newForms []*model.WarehouseOutForm
		newItems []*model.WarehouseOutFormItem
	)
	s.db.Preload("Product").Where("inventory_type = ?", 2).Find(&records)
	for _, record := range records {
		formUuid, err := utils.GetID()
		if err != nil {
			return err
		}
		newForms = append(newForms, &model.WarehouseOutForm{
			BaseModel: model.BaseModel{
				Uuid:       formUuid,
				CreateTime: record.OutTime,
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
				productBomUuid = uint64(record.ProductSkuID)
			} else {
				materialUuid = uint64(record.ProductSkuID)
			}
		}
		newItems = append(newItems, &model.WarehouseOutFormItem{
			BaseModel: model.BaseModel{
				Uuid:       itemUuid,
				CreateTime: record.OutTime,
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
	}

	if len(newItems) > 0 {
		if err := s.targetDB.Create(newItems).Error; err != nil {
			return err
		}
	}

	return nil
}

// 迁移损耗记录
func (s *StockService) ConvertDemaged() error {
	var (
		records    []*DamagedProductRecord
		newRecords []*model.LossReportForm
	)
	s.db.Find(&records)
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

	return nil
}
