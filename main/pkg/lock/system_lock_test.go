package lock

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"ttpos-server-go/pkg/cache"
)

func TestOpenDesk(t *testing.T) {
	// 分布式锁时需要redis信息
	InitRedisLock(cache.Config{
		//Host:     "192.168.100.69",
		//Port:     "6379",
		//Password: "",
		//DB:       1,

		Host:     "156.255.90.128",
		Port:     "7001",
		Password: "sass@123.com!",
		//DB:       1,
	})
	var wg sync.WaitGroup
	systemLock := NewSystemLock()

	for i := 1; i <= 140; i++ {
		wg.Add(1)
		go func(deskNo int) {
			defer wg.Done()
			service := NewService(systemLock)
			err := service.OpenDesk(1, fmt.Sprintf("test%d", deskNo))
			if err != nil {
				t.Error(err)
			}
		}(i)
	}

	wg.Wait()
}

func TestPay(t *testing.T) {
	//err := logger.Init()
	//if err != nil {
	//	fmt.Println(err)
	//}
	// 分布式锁时需要redis信息
	InitRedisLock(cache.Config{
		Host:     "192.168.100.69",
		Port:     "6379",
		Password: "",
		DB:       1,
	})
	var wg sync.WaitGroup
	systemLock := NewSystemLock()

	for i := 1; i <= 140; i++ {
		wg.Add(1)
		go func(orderId int) {
			defer wg.Done()
			service := NewService(systemLock)
			err := service.Pay(1, fmt.Sprintf("test%d", orderId))
			if err != nil {
				t.Error(err)
			}
		}(i)
	}

	wg.Wait()
}

type Service struct {
	systemLock Lock
}

func NewService(systemLock Lock) *Service {
	return &Service{systemLock: systemLock}
}

func (s *Service) OpenDesk(deskUuid uint64, client string) error {
	s.systemLock.LockUuid(deskUuid)
	defer s.systemLock.UnlockUuid(deskUuid)
	// 判断桌台是否为空闲状态
	if desk.deskUuid != nil {
		fmt.Println("client:", client, "申请的桌台非空闲,deskUuid:", deskUuid, "currentNum:", *desk.deskUuid)
		return nil
	}

	// 创建销售账单、销售订单；标记桌台为开桌状态
	desk.deskUuid = &deskUuid
	desk.client = &client
	fmt.Println("client:", client, "申请成功桌台,deskUuid:", deskUuid)
	return nil
}

func (s *Service) Pay(orderUuid uint64, client string) error {
	s.systemLock.LockUuid(orderUuid)
	defer s.systemLock.UnlockUuid(orderUuid)
	// 是否已经有现金付款单
	if order.payCash != nil {
		fmt.Println("client:", client, "订单已结清,orderUuid:", orderUuid)
		return nil
	}

	// 创建一个付款单
	iTrue := true
	order.payCash = &iTrue
	fmt.Println("client:", client, "付款成功,orderUuid:", orderUuid)
	return nil
}

type Desk struct {
	deskUuid *uint64
	client   *string
}

var desk = Desk{deskUuid: nil, client: nil}

type Order struct {
	payCash *bool
}

var order = Order{payCash: nil}

func TestLockUuid(t *testing.T) {

	rs := NewRedSync()
	mutex := rs.NewMutex("test-redsync")
	ctx := context.Background()

	if err := mutex.LockContext(ctx); err != nil {
		panic(err)
	}

	if _, err := mutex.UnlockContext(ctx); err != nil {
		panic(err)
	}
}
