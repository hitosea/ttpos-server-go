package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type StatisticsMemberRepository interface {
	ConvertStatisticsMember() error
}

func NewStatisticsMemberService(db *gorm.DB, targetDB *gorm.DB) StatisticsMemberRepository {
	return &StatisticsMemberService{db: db, targetDB: targetDB}
}

type StatisticsMemberService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *StatisticsMemberService) ConvertStatisticsMember() error {
	err := s.convertMemberBalanceLog(0, 2000)
	if err != nil {
		return err
	}
	return s.convertMember(0, 2000)
}

// 充值支付数据
func (s *StatisticsMemberService) convertMember(offset int, limit int) error {
	statisticsRepo := repository.NewStatisticsRepo(s.targetDB)
	// 获取数据
	var statisticsMembers []model.StatisticsMember
	err := s.db.Raw(`
		SELECT 
			a.id as member_recharge_order_uuid,
			duty_no,
			a.recharge_money as recharge_amount,
			a.gift_money as give_amount,
			a.gift_point as give_point,
			if(opt.value = 40, opt.price - a.change_due, opt.price) as payment_amount,
			a.pay_fee as payment_fee,
			ifnull(urord.refund_money, 0) as refund_amount,
			ifnull(urord.refund_money, 0) as refund_fee,
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
	`, offset, limit).Scan(&statisticsMembers).Error
	if err != nil {
		return err
	}

	// 保存数据
	if len(statisticsMembers) > 0 {
		fmt.Println(fmt.Sprintf("statistics-members - num: %d", len(statisticsMembers)))
		err := statisticsRepo.SaveMembers(statisticsMembers)
		if err != nil {
			return err
		}
		// 递归
		offset += limit
		return s.convertMember(offset, limit)
	}
	//
	return nil
}

// 充值支付数据
func (s *StatisticsMemberService) convertMemberBalanceLog(offset int, limit int) error {
	statisticsRepo := repository.NewStatisticsRepo(s.targetDB)
	// 获取数据
	var statisticsMembers []model.StatisticsMember
	err := s.db.Raw(`
		SELECT
			log.recharge_order_id as member_recharge_order_uuid,
			SUM(log.money - log.gift_money) AS recharge_amount,
			SUM(log.gift_money) AS give_amount,
			0 AS give_point,
			SUM(log.money - log.gift_money) AS payment_amount,
			0 AS payment_fee,
			0 AS refund_amount,
			0 AS refund_fee,
			log.create_time as complete_time
		FROM jjjfood_user_balance_log log
		WHERE log.version = '1.0.7' AND log.scene = '30'
		order by log.create_time
		Limit ?, ?
	`, offset, limit).Scan(&statisticsMembers).Error
	if err != nil {
		return err
	}

	// 保存数据
	if len(statisticsMembers) > 0 {
		fmt.Println(fmt.Sprintf("statistics-members-balance-log - num: %d", len(statisticsMembers)))
		err := statisticsRepo.SaveMembers(statisticsMembers)
		if err != nil {
			return err
		}
		// 递归
		offset += limit
		return s.convertMember(offset, limit)
	}
	//
	return nil
}
