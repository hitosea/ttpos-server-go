// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Migrations is the golang structure for table migrations.
type Migrations struct {
	Version       int64       `json:"version"       orm:"version"        description:""` //
	MigrationName string      `json:"migrationName" orm:"migration_name" description:""` //
	StartTime     *gtime.Time `json:"startTime"     orm:"start_time"     description:""` //
	EndTime       *gtime.Time `json:"endTime"       orm:"end_time"       description:""` //
	Breakpoint    int         `json:"breakpoint"    orm:"breakpoint"     description:""` //
}
