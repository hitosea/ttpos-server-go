// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// JobStatusLog is the golang structure for table job_status_log.
type JobStatusLog struct {
	Id           int64  `json:"id"           orm:"id"            description:"主键"`       // 主键
	Uuid         string `json:"uuid"         orm:"uuid"          description:"全局唯一ID"`   // 全局唯一ID
	JobUuid      string `json:"jobUuid"      orm:"job_uuid"      description:"外送订单uuid"` // 外送订单uuid
	StatusBefore string `json:"statusBefore" orm:"status_before" description:"变更前状态"`    // 变更前状态
	StatusAfter  string `json:"statusAfter"  orm:"status_after"  description:"变更后状态"`    // 变更后状态
	CreatedAt    int    `json:"createdAt"    orm:"created_at"    description:"创建时间"`     // 创建时间
	UpdatedAt    int    `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`     // 更新时间
	DeletedAt    int    `json:"deletedAt"    orm:"deleted_at"    description:"软删除"`      // 软删除
}
