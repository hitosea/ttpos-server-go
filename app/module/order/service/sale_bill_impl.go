package service

type SaleBillImpl struct {
}

func NewSaleBillImpl() *SaleBillImpl {
	return &SaleBillImpl{}
}

func (impl *SaleBillImpl) InterfaceImplName() string {
	return "MysqlSaleBillImpl"
}
