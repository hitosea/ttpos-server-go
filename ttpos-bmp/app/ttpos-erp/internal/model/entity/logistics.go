// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Logistics is the golang structure for table logistics.
type Logistics struct {
	Id           int64  `json:"id"           orm:"id"             description:"ID"`                        // ID
	Uuid         string `json:"uuid"         orm:"uuid"           description:"UUID"`                      // UUID
	Vendor       string `json:"vendor"       orm:"vendor"         description:"供应商，如 JT:极兔"`               // 供应商，如 JT:极兔
	VendorUserId string `json:"vendorUserId" orm:"vendor_user_id" description:"供应商用户id,如极兔的货主编码"`          // 供应商用户id,如极兔的货主编码
	InfConf      string `json:"infConf"      orm:"inf_conf"       description:"接口连接信息。如ak/sk 根据不同供应商有所不同"` // 接口连接信息。如ak/sk 根据不同供应商有所不同
	Remarks      string `json:"remarks"      orm:"remarks"        description:"备注信息"`                      // 备注信息
	Reserve1     string `json:"reserve1"     orm:"reserve1"       description:"保留字段1"`                     // 保留字段1
	Reserve2     string `json:"reserve2"     orm:"reserve2"       description:"保留字段2"`                     // 保留字段2
}
