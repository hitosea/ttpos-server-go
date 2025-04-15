package config

type ServerConf struct {
	Port       string // 端口
	Mode       string // 模式：debug/release/test
	DeployMode string // 部署模式：cloud云上，offline离线
}

type DatabaseConf struct {
	DBType          string
	Host            string
	Port            int
	User            string
	Password        string
	RootPassword    string
	Database        string // 主数据库，如果是sqlite3，则为文件路径
	TablePrefix     string // 表名前缀
	SlowQueryTime   int    // 慢查询阈值，单位秒
	MaxIdleConns    int    // 最大空闲连接数
	MaxOpenConns    int    // 最大打开的连接数
	ConnMaxLifetime int    // 连接的最大可存活时间，单位秒
}

type RedisConf struct {
	Host     string
	Port     string
	Password string
	DB       int // 数据库(0~15)
}

type JWTConf struct {
	Secret string // 密钥
	Expire int    // 有效期，单位秒
}

type CaptchaConf struct {
	CachePrefix string // 缓存前缀
}

type EncryptConf struct {
	CachePrefix   string // 缓存前缀
	EncryptHeader string // 请求头x-encrypt
	ClientID      string // 请求头x-encrypt字段client_id
	ClientKey     string // 请求头x-encrypt字段client_key
}

type LogConf struct {
	Dir           string // 日志路径
	Level         string // 日志级别：debug/info/warn/error
	MaxSize       int    // 每个日志文件保存的最大尺寸 单位：M
	MaxBackup     int    // 日志文件最大数量
	CleanSchedule string // 定时清理日志
}
