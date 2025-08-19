// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// LanPrinterScan is the golang structure for table lan_printer_scan.
type LanPrinterScan struct {
	Id             uint   `json:"id"             orm:"id"               description:""`               //
	Uuid           uint64 `json:"uuid"           orm:"uuid"             description:"uuid"`           // uuid
	Ip             string `json:"ip"             orm:"ip"               description:"ip"`             // ip
	Port           int    `json:"port"           orm:"port"             description:"端口"`             // 端口
	Status         int    `json:"status"         orm:"status"           description:"状态 0: 离线 1: 在线"` // 状态 0: 离线 1: 在线
	Remark         string `json:"remark"         orm:"remark"           description:"备注"`             // 备注
	SourceDeviceSn string `json:"sourceDeviceSn" orm:"source_device_sn" description:"来源设备SN"`         // 来源设备SN
	CreateTime     int    `json:"createTime"     orm:"create_time"      description:"创建时间"`           // 创建时间
	UpdateTime     int    `json:"updateTime"     orm:"update_time"      description:"更新时间"`           // 更新时间
	DeleteTime     int    `json:"deleteTime"     orm:"delete_time"      description:"删除时间"`           // 删除时间
}
