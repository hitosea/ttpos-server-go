package crm

import "ttpos-bmp/app/ttpos-erp/internal/service"

var (
	Crm = new(sCrm)
)

type sCrm struct {
}

func init() {
	service.RegisterCrm(Crm)
}
