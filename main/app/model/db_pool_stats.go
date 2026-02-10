package model

// DBPoolStats 数据库连接池统计记录 `ttpos_db_pool_stats`
// 存储在 SaaS 主库中，用于监控数据库连接池健康状态
type DBPoolStats struct {
	Id                uint64 `gorm:"column:id;type:bigint(20) unsigned;primaryKey;autoIncrement;comment:自增ID"`
	Uuid              uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;uniqueIndex;comment:记录UUID"`
	InstanceId        string `gorm:"column:instance_id;type:varchar(128);default:'';index;comment:服务实例标识"`
	DBName            string `gorm:"column:db_name;type:varchar(100);default:'';index;comment:数据库名称"`
	MaxOpenConns      int    `gorm:"column:max_open_conns;type:int(11);default:0;comment:最大打开连接数"`
	OpenConns         int    `gorm:"column:open_conns;type:int(11);default:0;comment:当前打开连接数"`
	InUse             int    `gorm:"column:in_use;type:int(11);default:0;comment:正在使用的连接数"`
	Idle              int    `gorm:"column:idle;type:int(11);default:0;comment:空闲连接数"`
	WaitCount         int64  `gorm:"column:wait_count;type:bigint(20);default:0;comment:等待连接的累计次数"`
	WaitDurationMs    int64  `gorm:"column:wait_duration_ms;type:bigint(20);default:0;comment:等待连接的累计时间(毫秒)"`
	MaxIdleClosed     int64  `gorm:"column:max_idle_closed;type:bigint(20);default:0;comment:因超过MaxIdleConns被关闭的连接数"`
	MaxIdleTimeClosed int64  `gorm:"column:max_idle_time_closed;type:bigint(20);default:0;comment:因ConnMaxIdleTime被关闭的连接数"`
	MaxLifetimeClosed int64  `gorm:"column:max_lifetime_closed;type:bigint(20);default:0;comment:因ConnMaxLifetime被关闭的连接数"`
	SampleTime        int64  `gorm:"column:sample_time;type:bigint(20);default:0;index;comment:采样时间戳(毫秒)"`
	CreateTime        int64  `gorm:"column:create_time;type:int(10) unsigned;default:0;comment:创建时间(时间戳)"`
	UpdateTime        int64  `gorm:"column:update_time;type:int(10) unsigned;default:0;comment:更新时间(时间戳)"`
	DeleteTime        int64  `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间(时间戳)"`
}

// TableName 指定表名
func (DBPoolStats) TableName() string {
	return "ttpos_db_pool_stats"
}
