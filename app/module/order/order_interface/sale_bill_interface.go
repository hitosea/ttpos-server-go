package order_interface

import "jjjshop-server-go/app/module/order/service"

type SaleBillInterface interface {
	InterfaceImplName() string
	// 定义各个表需要使用的增删改查方法
}

func NewSaleBillInterface() SaleBillInterface {
	return service.NewSaleBillImpl()
}
