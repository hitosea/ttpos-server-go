package req

// DBPoolStatsRecord 数据库连接池统计记录
// 用于在内存队列中传输的数据
type DBPoolStatsRecord struct {
	InstanceId        string // 服务实例标识
	DBName            string // 数据库名称
	MaxOpenConns      int    // 最大打开连接数
	OpenConns         int    // 当前打开连接数
	InUse             int    // 正在使用的连接数
	Idle              int    // 空闲连接数
	WaitCount         int64  // 等待连接的累计次数
	WaitDurationMs    int64  // 等待连接的累计时间(毫秒)
	MaxIdleClosed     int64  // 因超过MaxIdleConns被关闭的连接数
	MaxIdleTimeClosed int64  // 因ConnMaxIdleTime被关闭的连接数
	MaxLifetimeClosed int64  // 因ConnMaxLifetime被关闭的连接数
	SampleTime        int64  // 采样时间戳(毫秒)
}
