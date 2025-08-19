// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Migrations is the golang structure of table ttpos_migrations for DAO operations like Where/Data.
type Migrations struct {
	g.Meta        `orm:"table:ttpos_migrations, do:true"`
	Version       interface{} //
	MigrationName interface{} //
	StartTime     *gtime.Time //
	EndTime       *gtime.Time //
	Breakpoint    interface{} //
}
