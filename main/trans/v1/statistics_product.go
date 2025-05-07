package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type StatisticsProductRepository interface {
	ConvertStatisticsProduct() error
}

func NewStatisticsProductService(db *gorm.DB, targetDB *gorm.DB) StatisticsProductRepository {
	return &StatisticsProductService{db: db, targetDB: targetDB}
}

type StatisticsProductService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *StatisticsProductService) ConvertStatisticsProduct() error {
	return s.convertProduct(0, 20000)
}

// 商品数据
func (s *StatisticsProductService) convertProduct(offset int, limit int) error {
	statisticsRepo := repository.NewStatisticsRepo(s.targetDB)
	// 获取数据
	var statisticsPayments []model.StatisticsProduct

	// (rp.total_pay_price / rp.total_num ) +
	// IF(rp.tax_calc_type = 2, rp.product_consumption_tax / rp.total_num , 0) +
	// (rp.product_service_fee / rp.total_num ) +
	// (rp.product_service_consumption_tax / rp.total_num)
	// as product_final_price,

	// IF(rp.tax_calc_type = 1, rp.product_price - rp.consumption_tax , rp.product_price) as product_price,

	// 统计订单数量
	err := s.db.Raw(`
        SELECT
            a.order_id as sale_bill_uuid,   
            a.order_id as sale_order_uuid,
            a.table_id as desk_uuid,
            rp.product_id as product_package_uuid,
            rp.product_sku_id as product_bom_uuid,
            rp.total_num as product_num,
            rp.line_price as flavor_price,
            rp.feed_price as sauce_price,
			IF(rp.tax_calc_type = 1, rp.product_price - rp.consumption_tax , rp.product_price) as product_price,
            rp.product_price as product_sale_price,
			rp.total_pay_price / rp.total_num as product_final_price,
            rp.tax_rate / 100 as tax_rate,
            rp.product_consumption_tax / rp.total_num as tax_fee,
            rp.product_service_fee / rp.total_num  as service_fee,
            rp.product_service_consumption_tax / rp.total_num  as service_tax,
            IF(rp.is_free != 0, rp.total_num, 0)  as give_num,
            IF(a.is_free != 0, 1, 0)  as free_num,  
            rp.refund_num as refund_num,    
            a.pay_time as complete_time,
            a.create_time as create_time
        FROM jjjfood_order a
        LEFT JOIN jjjfood_order_product rp ON a.order_id = rp.order_id
        WHERE  a.pay_status = '20' 
            AND a.order_status = '30' 
            AND a.is_delete = '0' 
            AND rp.product_id > '0' 
            AND rp.is_return = '0'
            AND a.delete_time = '0' 
        Limit ?, ?
    `, offset, limit).Scan(&statisticsPayments).Error
	if err != nil {
		return err
	}

	// 保存数据
	if len(statisticsPayments) > 0 {
		fmt.Println(fmt.Sprintf("statistics-product - num: %d - offset: %d", len(statisticsPayments), offset))

		// 分批保存数据，每批2000条
		batchSize := 2000
		for i := 0; i < len(statisticsPayments); i += batchSize {
			end := i + batchSize
			if end > len(statisticsPayments) {
				end = len(statisticsPayments)
			}

			// 保存当前批次数据
			err := statisticsRepo.SaveProduct(statisticsPayments[i:end])
			if err != nil {
				return fmt.Errorf("保存数据失败 (offset=%d, batch=%d-%d): %w", offset, i, end, err)
			}

			fmt.Printf("已保存批次: offset=%d, batch=%d-%d\n", offset, i, end)
		}

		// 递归
		offset += limit
		return s.convertProduct(offset, limit)
	}
	//
	return nil
}
