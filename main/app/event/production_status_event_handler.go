package event

import "ttpos-server-go/pkg/utils"

type ProductionStatusEvent struct {
	ProductionUuid uint64
	Status         uint
}

func (e *ProductionStatusEvent) ToJsonString() string {
	return utils.ToJson(e)
}
