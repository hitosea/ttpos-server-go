package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type StatisticsPaymentRepository interface {
	ConvertStatisticsPayment() error
}

func NewStatisticsPaymentService(db *gorm.DB, targetDB *gorm.DB) StatisticsPaymentRepository {
	return &StatisticsPaymentService{db: db, targetDB: targetDB}
}

type StatisticsPaymentService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *StatisticsPaymentService) ConvertStatisticsPayment() error {
	err := s.convertPayment(0, 2000)
	if err != nil {
		return nil
	}
	return s.convertRechargePayment(0, 2000)
}

// 支付数据
func (s *StatisticsPaymentService) convertPayment(offset int, limit int) error {
	statisticsRepo := repository.NewStatisticsRepo(s.targetDB)
	// 获取数据
	var statisticsPayments []model.StatisticsPayment
	err := s.db.Raw(`
		SELECT 
			a.order_id as sale_bill_uuid,
			a.order_id as sale_order_uuid,
			a.table_id as desk_uuid,
			if(opt.value = -1, 0, pt.id) as payment_method_uuid,
			if(opt.value = 40, opt.price - a.change_due, opt.price) as payment_amount,
			ifnull(ord.refund_money, 0) as refund_amount,
			a.pay_time as complete_time
		FROM jjjfood_order a 
		LEFT JOIN jjjfood_order_pay_type opt ON a.order_id = opt.order_id 
		LEFT JOIN jjjfood_pay_type pt ON opt.value = pt.value 
		LEFT JOIN (
			select ord.order_id, ord.value, ifnull(sum(ord.refund_money), 0) as refund_money
			from jjjfood_order_refund_destination ord
			where ord.status = 1
			group by ord.order_id, ord.value
		) ord ON a.order_id=ord.order_id AND opt.value = ord.value 
		WHERE ( 
		 	a.pay_status = 20  AND 
			a.order_status = 30  AND 
			a.is_delete = 0  AND 
			a.parent_id = 0 AND
			opt.id IS NOT NULL AND 
			opt.value != -1
		) AND a.delete_time = 0 
		order by a.pay_time
		Limit ?, ?
	`, offset, limit).Scan(&statisticsPayments).Error
	if err != nil {
		return err
	}

	// 保存数据
	if len(statisticsPayments) > 0 {
		fmt.Println(fmt.Sprintf("statistics-payments - num: %d", len(statisticsPayments)))
		err := statisticsRepo.SavePayment(statisticsPayments)
		if err != nil {
			return err
		}
		// 递归
		offset += limit
		return s.convertPayment(offset, limit)
	}
	//
	return nil
}

// 充值支付数据
func (s *StatisticsPaymentService) convertRechargePayment(offset int, limit int) error {
	statisticsRepo := repository.NewStatisticsRepo(s.targetDB)
	// 获取数据
	var statisticsPayments []model.StatisticsMemberPayment
	err := s.db.Raw(`
		SELECT 
			a.id as member_recharge_order_uuid,
			duty_no,
			if(opt.value = -1, 0, pt.id) as payment_method_uuid,
			if(opt.value = 40, opt.price - a.change_due, opt.price) as payment_amount,
			ifnull(urord.refund_money, 0) as refund_amount,
			a.pay_time as complete_time
		FROM jjjfood_user_recharge_order a 
		LEFT JOIN jjjfood_user_recharge_order_pay_type opt ON a.id=opt.order_id
		LEFT JOIN jjjfood_pay_type pt ON opt.value = pt.value 
		LEFT JOIN (
			select urord.order_id, urord.value, ifnull(sum(urord.refund_money), 0) as refund_money
			from jjjfood_user_recharge_order_refund_destination urord
			where urord.status = 1
			group by urord.order_id, urord.value
		) urord ON a.id= urord.order_id AND opt.value = urord.value 
		WHERE a.order_status = '1' AND opt.id IS NOT NULL AND opt.value != -1
		order by a.pay_time
		Limit ?, ?
	`, offset, limit).Scan(&statisticsPayments).Error
	if err != nil {
		return err
	}

	// 保存数据
	if len(statisticsPayments) > 0 {
		fmt.Println(fmt.Sprintf("statistics-recharge-payments - num: %d", len(statisticsPayments)))
		err := statisticsRepo.SaveMemberPayment(statisticsPayments)
		if err != nil {
			return err
		}
		// 递归
		offset += limit
		return s.convertRechargePayment(offset, limit)
	}
	//
	return nil
}
